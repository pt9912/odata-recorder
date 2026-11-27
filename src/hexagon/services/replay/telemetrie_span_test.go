package replay_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// TestPflichtSchluesselMatch hält, dass die Literale des Tests die Pflicht-Attribute von odata.replay.match sind (SPEC-028).
func TestPflichtSchluesselMatch(t *testing.T) {
	for _, s := range support.Spans() {
		if s.Name == support.SpanReplayMatch {
			if !slices.Equal(s.Pflicht, []string{attrStrategie, attrErgebnis}) {
				t.Errorf("Pflicht-Attribute %v, erwartet [%s %s]", s.Pflicht, attrStrategie, attrErgebnis)
			}
			return
		}
	}
	t.Fatal("support.Spans() nennt odata.replay.match nicht")
}

// bodyMitFehler setzt einen Request-Body, der den Fehler vor dem ersten Byte liefert (S2).
func bodyMitFehler(a *inbound.Anfrage, err error) {
	a.Body, a.Laenge = &fehlerLeser{fehler: err, wann: "vor dem ersten Byte"}, 4
}

// TestKeinSpanOhneZuordnung hält W1 (SPEC-296 Nr. 2): Menge M-OhneSpan-Replay, (a) 4 Argumentfehler, (b) 3 Pfade ohne Dienst, (c) 12 Verstöße
// der Zuordnung (3 Header × fehlt, leere Wertliste, zwei gleiche Werte, dazu 3 Formen der Replay-Session-ID); jeder endet vor dem Beginn:
// kein Aufruf von Telemetrie, kein Port. Zählprüfung 19.
func TestKeinSpanOhneZuordnung(t *testing.T) {
	type fall struct {
		name string
		ctx  func() context.Context
		mod  func(*inbound.Anfrage)
	}
	bg := func() context.Context { return context.Background() }
	faelle := []fall{
		{"a ctx nil", func() context.Context { return nil }, func(*inbound.Anfrage) {}},
		{"a Methode leer", bg, func(a *inbound.Anfrage) { a.Methode = "" }},
		{"a Laenge -2", bg, func(a *inbound.Anfrage) { a.Laenge = -2 }},
		{"a Body nil bei Laenge 1", bg, func(a *inbound.Anfrage) { a.Laenge = 1 }},
		{"b kein Dienst", bg, func(a *inbound.Anfrage) { a.Pfad = "/x" }},
		{"b leerer Pfad", bg, func(a *inbound.Anfrage) { a.Pfad = "" }},
		{"b Pfad ohne Schrägstrich", bg, func(a *inbound.Anfrage) { a.Pfad = "svc/Products" }},
	}
	for _, h := range []string{"X-Recorder-Workspace", "X-Recorder-Recording-Session", "X-Recorder-Replay-Session"} {
		for art, mod := range map[string]func(*inbound.Anfrage){
			"fehlt":        func(a *inbound.Anfrage) { delete(a.Header, h) },
			"leere Liste":  func(a *inbound.Anfrage) { a.Header[h] = []string{} },
			"zwei gleiche": func(a *inbound.Anfrage) { a.Header[h] = append(a.Header[h], a.Header[h][0]) },
		} {
			faelle = append(faelle, fall{"c " + h + " " + art, bg, mod})
		}
	}
	for name, id := range map[string]string{"129 Bytes": strings.Repeat("a", 129), "Leerzeichen": "a b", "Byte ab 0x80": string([]byte{0x61, 0x80})} {
		faelle = append(faelle, fall{"c Form " + name, bg, func(a *inbound.Anfrage) { a.Header["X-Recorder-Replay-Session"] = []string{id} }})
	}
	for _, f := range faelle {
		u := neu(t)
		u.mitJournal()
		a := anf()
		f.mod(&a)
		_, err := u.replay(f.ctx(), a)
		if err == nil {
			t.Errorf("%s: kein Fehler", f.name)
		}
		if n := len(u.tel.Aufrufe()); n != 0 {
			t.Errorf("%s: %d Aufrufe von Telemetrie, erwartet 0", f.name, n)
		}
		if n := u.p.gesamt(); n != 0 {
			t.Errorf("%s: %d Operationen im Protokoll, erwartet 0", f.name, n)
		}
	}
	if len(faelle) != 19 {
		t.Errorf("Menge: %d Elemente, erwartet 19", len(faelle))
	}
}

// erfolgsFolge ist die Folge der Aufrufe eines Treffers mit Cursor: Span, die sieben Ports, Set, End.
func erfolgsFolge(mitPut bool, mitLog bool) []string {
	f := []string{"Telemetrie.StartSpan", "Sessions.Get", "Exchanges.ListByMatchKey", "Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get",
		"tx Cursors.Get", "Blobs.Open"}
	if mitPut {
		f = append(f, "tx Cursors.Put")
	}
	f = append(f, "Span.Set")
	if mitLog {
		f = append(f, "Telemetrie.Log")
	}
	return append(f, "Span.End")
}

// TestSpanFolge hält W2 (SPEC-296 Nr. 3 a, 6): Beginn nach der Zuordnung und vor Sessions.Get, dann die Ports, Set nach der Rückkehr der
// Transaktion, Log, End; die Folge je Fall steht als Literal im Test: Treffer mit sequential, Treffer mit KeyRedacted (mit Log), first
// ohne Put, Fehler in f (Set nach der Rückkehr der Transaktion). Menge 4.
func TestSpanFolge(t *testing.T) {
	faelle := []struct {
		name  string
		strat model.ReplayStrategy
		baue  func(u *umg)
		folge []string
	}{
		{"Treffer sequential", model.StrategySequential, func(u *umg) { u.lege(u.austausch(1, u.keyA(), true)) }, erfolgsFolge(true, false)},
		{"Treffer mit KeyRedacted", model.StrategySequential, func(u *umg) {
			e := u.austausch(1, u.keyA(), true)
			e.KeyRedacted = true
			u.lege(e)
		}, erfolgsFolge(true, true)},
		{"first ohne Put", model.StrategyFirst, func(u *umg) { u.lege(u.austausch(1, u.keyA(), true)) }, erfolgsFolge(false, false)},
		{"Fehler in f", model.StrategySequential, func(*umg) {}, []string{"Telemetrie.StartSpan", "Sessions.Get", "Exchanges.ListByMatchKey",
			"Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get", "tx Cursors.Get", "Span.Set", "Span.End"}},
	}
	for _, f := range faelle {
		u := neu(t)
		u.mitJournal()
		u.k.Strategie = f.strat
		f.baue(u)
		resp, err := u.replay(context.Background(), anf())
		if err == nil {
			_ = resp.Body.Close()
		}
		if got := u.p.ops(); !slices.Equal(got, f.folge) {
			t.Errorf("%s: Folge\n%v\nerwartet\n%v", f.name, got, f.folge)
		}
	}
	if len(faelle) != 4 {
		t.Errorf("Menge: %d, erwartet 4", len(faelle))
	}
}

// TestSpanNurInF hält W8 (SPEC-296 Nr. 3 c): zwischen dem Beginn und der Rückkehr von WithWorkspaceTransaction ruft der Dienst Telemetrie nie
// (der Journal-Fake meldet jeden Aufruf in f), und Set folgt der Rückkehr der Transaktion: 4 Strategien × {Erfolg, Fehler in f} gleich 8.
func TestSpanNurInF(t *testing.T) {
	n := 0
	for _, strat := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		for _, erfolg := range []bool{true, false} {
			n++
			u := neu(t)
			u.mitJournal()
			u.k.Strategie = strat
			if erfolg {
				u.lege(u.austausch(1, u.keyA(), true))
			}
			resp, err := u.replay(context.Background(), anf())
			if err == nil {
				_ = resp.Body.Close()
			}
			if (err == nil) != erfolg {
				t.Errorf("%s %v: Fehler %v", strat, erfolg, err)
			}
			ops := u.p.ops()
			letzteTx := -1
			for i, op := range ops {
				if strings.HasPrefix(op, "tx ") {
					letzteTx = i
				}
			}
			if set := slices.Index(ops, "Span.Set"); letzteTx < 0 || set < letzteTx || set < slices.Index(ops, "Transaktion.WithWorkspaceTransaction") {
				t.Errorf("%s %v: Set an Stelle %d vor dem Ende von f (%d): %v", strat, erfolg, set, letzteTx, ops)
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge: %d, erwartet 8", n)
	}
}

// bauePaar legt die Exchanges und den Cursor für ein erreichbares Paar aus Strategie und Ergebnis (SPEC-296 Nr. 4 c).
func bauePaar(u *umg, strat model.ReplayStrategy, ergebnis string) {
	switch ergebnis {
	case ergTreffer:
		u.lege(u.austausch(1, u.keyA(), true))
	case ergNichtReplayf:
		u.lege(u.austausch(1, u.keyA(), false))
	case ergErschoepft:
		u.lege(u.austausch(1, u.keyA(), true))
		u.cursor(strat, 1)
	case ergAndererKey:
		u.lege(u.austausch(1, u.keyB(), true), u.austausch(2, u.keyA(), true))
		u.cursor(strat)
	}
}

// codeZu ist der Code des Ergebnisses der Auswahl, als Literal notiert (SPEC-296 Nr. 4 b).
func codeZu(ergebnis string) model.Code {
	return map[string]model.Code{ergKeinTreffer: model.CodeE5001, ergErschoepft: model.CodeE5002, ergAndererKey: model.CodeE5007,
		ergNichtReplayf: model.CodeE5008}[ergebnis]
}

// TestSpanJedesPaar hält W4 (SPEC-296 Nr. 4 a bis c): Menge M-Paare, 15 erreichbare Paare aus Strategie × Ergebnis (sequential 4, strict 5,
// first 3, last 3); die Endwerte sind je Paar Literale, und das Ergebnis ist das der Auswahl, nicht ihr Gegenteil. Zählprüfung 15 von 20.
func TestSpanJedesPaar(t *testing.T) {
	erreichbar := map[model.ReplayStrategy][]string{
		model.StrategySequential: {ergTreffer, ergKeinTreffer, ergErschoepft, ergNichtReplayf},
		model.StrategyStrict:     {ergTreffer, ergKeinTreffer, ergErschoepft, ergAndererKey, ergNichtReplayf},
		model.StrategyFirst:      {ergTreffer, ergKeinTreffer, ergNichtReplayf},
		model.StrategyLast:       {ergTreffer, ergKeinTreffer, ergNichtReplayf},
	}
	alle := []string{ergTreffer, ergKeinTreffer, ergErschoepft, ergAndererKey, ergNichtReplayf}
	n, paare := 0, 0
	for _, strat := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		paare += len(alle)
		for _, erg := range erreichbar[strat] {
			n++
			erg := erg
			spanFall{name: string(strat) + " " + erg, strat: strat, baue: func(u *umg) inbound.Anfrage { bauePaar(u, strat, erg); return anf() },
				soll: soll{code: codeZu(erg)}, strategie: string(strat), ergebnis: erg, set: true}.lauf(t)
		}
	}
	if n != 15 || paare != 20 {
		t.Errorf("Menge: %d von %d Paaren, erwartet 15 von 20", n, paare)
	}
}

// stoerung ist die Einspeisung eines Fehlers an einer Stelle des Protokolls.
func stoerung(op string, err error) func(*umg) inbound.Anfrage {
	return func(u *umg) inbound.Anfrage {
		a := erfolgsLauf(u)
		u.p.setzeFehler(op, err)
		return a
	}
}

// platzhalterFaelle ist die Menge M-Platzhalter (SPEC-296 Nr. 4 a, b, d): 26 Wege, auf denen die Auswahl nicht lief; Set nur, wo S4 gerufen
// wurde.
func platzhalterFaelle() []spanFall {
	ng, andere := outbound.ErrNichtGefunden, errors.New("andere")
	fall := func(name string, baue func(u *umg) inbound.Anfrage, s soll, set bool) spanFall {
		return spanFall{name: name, baue: baue, soll: s, strategie: nichtGelesen, ergebnis: nichtGewaehlt, set: set}
	}
	var r []spanFall
	// S1: vier Codes.
	r = append(r,
		fall("S1 E7003", stoerung("Sessions.Get", ng), soll{code: model.CodeE7003}, false),
		fall("S1 E3001", stoerung("Sessions.Get", outbound.ErrBeschaedigt), soll{code: model.CodeE3001}, false),
		fall("S1 E5004", func(u *umg) inbound.Anfrage {
			u.sess.setze(model.Session{ID: sesID, ServiceID: "svc-b", CreatedAt: zeit(), Status: model.SessionComplete})
			return erfolgsLauf(u)
		}, soll{code: model.CodeE5004}, false),
		fall("S1 E7009", func(u *umg) inbound.Anfrage {
			u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionOpen})
			return erfolgsLauf(u)
		}, soll{code: model.CodeE7009}, false),
		// Body und Key: drei Codes.
		fall("Body E5001", func(u *umg) inbound.Anfrage {
			a := erfolgsLauf(u)
			a.Header["Content-Encoding"] = []string{"gzip"}
			a.Body, a.Laenge = strings.NewReader("x"), 1
			return a
		}, soll{code: model.CodeE5001}, false),
		fall("Key E5006", func(u *umg) inbound.Anfrage {
			u.k.Regeln.InvalidEncoding = model.KodierungReject
			a := erfolgsLauf(u)
			a.RawQuery = "a=%zz"
			return a
		}, soll{code: model.CodeE5006}, false),
		fall("Key E9001", func(u *umg) inbound.Anfrage { a := erfolgsLauf(u); a.Pfad = root + "/a?b"; return a }, soll{code: model.CodeE9001}, false),
		// S3: zwei Codes.
		fall("S3 E7003", stoerung("Exchanges.ListByMatchKey", ng), soll{code: model.CodeE7003}, false),
		fall("S3 E3001", stoerung("Exchanges.ListByMatchKey", outbound.ErrBeschaedigt), soll{code: model.CodeE3001}, false),
		// S4 vor f.
		fall("S4 vor f", stoerung("Transaktion.WithWorkspaceTransaction", andere), soll{code: model.CodeE9001}, true),
		// S5: drei Codes.
		fall("S5 E7001", stoerung("tx Workspaces.Get", ng), soll{code: model.CodeE7001}, true),
		fall("S5 E5004", func(u *umg) inbound.Anfrage {
			z := u.st.stand()
			w := z.ws[wsID]
			w.ServiceID = "svc-b"
			z.ws[wsID] = w
			u.st.setze(z)
			return erfolgsLauf(u)
		}, soll{code: model.CodeE5004}, true),
		fall("S5 E3001", stoerung("tx Workspaces.Get", outbound.ErrBeschaedigt), soll{code: model.CodeE3001}, true),
		// S6: ein Fehler, der nicht der fehlende Cursor ist.
		fall("S6 E3001", stoerung("tx Cursors.Get", outbound.ErrBeschaedigt), soll{code: model.CodeE3001}, true),
	)
	// Kontextfehler an S1 bis S6 (S2 im Body des Requests), je Art des Kontextfehlers.
	for _, k := range kontextfehler() {
		for _, op := range []string{"Sessions.Get", "", "Exchanges.ListByMatchKey", "Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get", "tx Cursors.Get"} {
			r = append(r, fall("Kontext "+op+" "+k.Error(), kontextAn(op, k), soll{kontext: k}, op != "" && op != "Sessions.Get" && op != "Exchanges.ListByMatchKey"))
		}
	}
	return r
}

// kontextfehler sind die beiden Arten des Kontextfehlers (SPEC-109 Nr. 6): Abbruch und Fristablauf.
func kontextfehler() []error { return []error{context.Canceled, context.DeadlineExceeded} }

// kontextAn speist den Fehler err (ein Kontextfehler, ein Verbund oder ein Fehler mit Code) an der Stelle op ein; die leere Stelle ist der
// Body des Requests (S2).
func kontextAn(op string, err error) func(*umg) inbound.Anfrage {
	if op != "" {
		return stoerung(op, err)
	}
	return func(u *umg) inbound.Anfrage {
		a := erfolgsLauf(u)
		bodyMitFehler(&a, err)
		return a
	}
}

// TestSpanPlatzhalter hält W5 (SPEC-296 Nr. 4 a, b, d): Menge M-Platzhalter, 26 Wege ohne Auswahl; beide Platzhalter bleiben, und Set steht
// nur, wo WithWorkspaceTransaction gerufen wurde (auch wenn f nie lief). Zählprüfung 26 (S1 4, Body und Key 3, S3 2, S4 1, S5 3, S6 1,
// Kontextfehler 6 je Art des Kontextfehlers, die zwei Arten ergeben 12 und zusammen 26).
func TestSpanPlatzhalter(t *testing.T) {
	faelle := platzhalterFaelle()
	for _, f := range faelle {
		f.lauf(t)
	}
	if len(faelle) != 26 || len(kontextfehler()) != 2 {
		t.Errorf("Menge: %d, erwartet 26", len(faelle))
	}
}

// cursorStrategien und cursorErgebnisse sind die Strategien und die Ergebnisse der Auswahl (SPEC-296 Nr. 4 a, b).
func cursorStrategien() []model.ReplayStrategy {
	return []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast}
}

func cursorErgebnisse() []string {
	return []string{ergTreffer, ergKeinTreffer, ergErschoepft, ergAndererKey, ergNichtReplayf}
}

// cursorZustaende ist die Zahl der Zustände des Cursors: k gleich 0, 1 und 2 Verbräuche, das ist consumed {1..k} mit position k (SPEC-108 Nr. 10).
const cursorZustaende = 3

// cursorErreichbar ist das Literal der 43 Tripel aus Strategie, Zustand k und Ergebnis, die waehle erreicht (SPEC-296 Nr. 7 d): je Strategie
// die Ergebnisse für k gleich 0, 1 und 2. Nicht erreichbar sind 17 der 60 Tripel: anderer_key bei sequential, first und last (9), erschoepft
// bei first und last (6), erschoepft im Zustand neu bei sequential und strict (2).
func cursorErreichbar() map[model.ReplayStrategy][cursorZustaende][]string {
	drei := []string{ergTreffer, ergKeinTreffer, ergNichtReplayf}
	mitErschoepft := []string{ergTreffer, ergKeinTreffer, ergErschoepft, ergNichtReplayf}
	strictNeu := []string{ergTreffer, ergKeinTreffer, ergAndererKey, ergNichtReplayf}
	strictVerbraucht := []string{ergTreffer, ergKeinTreffer, ergErschoepft, ergAndererKey, ergNichtReplayf}
	return map[model.ReplayStrategy][cursorZustaende][]string{
		model.StrategySequential: {drei, mitErschoepft, mitErschoepft},
		model.StrategyStrict:     {strictNeu, strictVerbraucht, strictVerbraucht},
		model.StrategyFirst:      {drei, drei, drei},
		model.StrategyLast:       {drei, drei, drei},
	}
}

// cursorLage legt die Exchanges für das Ergebnis erg im Zustand k des Cursors und liefert den Request: treffer, kein_treffer und
// nicht_replayfaehig tragen drei Exchanges (SequenceNo 1 bis 3, Key A; replayfähig bei treffer und kein_treffer, bei nicht_replayfaehig
// keiner; kein_treffer fragt Key C); erschoepft trägt k replayfähige Exchanges (1 bis k, Key A); anderer_key trägt drei, der mit SequenceNo k+1
// den Key B, die anderen den Key A.
func cursorLage(u *umg, erg string, k int) inbound.Anfrage {
	var es []model.Exchange
	switch erg {
	case ergTreffer, ergKeinTreffer, ergNichtReplayf:
		for seq := 1; seq <= 3; seq++ {
			es = append(es, u.austausch(seq, u.keyA(), erg != ergNichtReplayf))
		}
	case ergErschoepft:
		for seq := 1; seq <= k; seq++ {
			es = append(es, u.austausch(seq, u.keyA(), true))
		}
	case ergAndererKey:
		for seq := 1; seq <= 3; seq++ {
			key := u.keyA()
			if seq == k+1 {
				key = u.keyB()
			}
			es = append(es, u.austausch(seq, key, true))
		}
	}
	u.lege(es...)
	if erg == ergKeinTreffer {
		return anfC()
	}
	return anf()
}

// verbraucht ist consumed des Cursors nach k Verbräuchen: {1..k}, leer für k gleich 0.
func verbraucht(k int) []int64 {
	var c []int64
	for seq := 1; seq <= k; seq++ {
		c = append(c, int64(seq))
	}
	return c
}

// TestSpanCursorStrategie hält W6 (SPEC-296 Nr. 4 a, 7 d): die Strategie des bestehenden Cursors gilt im Span, in jedem Zustand des Cursors
// und für jedes Ergebnis, die der Konfig nur ohne Cursor. Menge M-Cursor mit 185 Elementen: die 43 erreichbaren Tripel aus Strategie am Cursor,
// Zustand (k gleich 0, 1, 2 Verbräuche) und Ergebnis (Literal cursorErreichbar) × Strategie der Konfig 4 (die gleiche und drei abweichende)
// gleich 172, dazu ohne Cursor 4 Strategien der Konfig × die Ergebnisse, die die Strategie im Zustand neu erreicht (3, 4, 3, 3), gleich 13.
// Zählprüfung als Literal: 4, 4, 3, 5, 60 Tripel, 43 (11, 14, 9, 9), 17 nicht erreichbar (9, 6, 2), 172, 13 und 185; dass die 43 die Tripel sind,
// die waehle erreicht, hält TestCursorTripelErreichbar. Nicht geführt: ein Zustand des Cursors ab drei Verbräuchen.
func TestSpanCursorStrategie(t *testing.T) {
	erreichbar, strategien, ergebnisse := cursorErreichbar(), cursorStrategien(), cursorErgebnisse()
	n, mitCursor, ohneCursor, tripel := 0, 0, 0, 0
	je := map[model.ReplayStrategy]int{}
	for _, amCursor := range strategien {
		for k := range cursorZustaende {
			for _, erg := range erreichbar[amCursor][k] {
				tripel++
				je[amCursor]++
				for _, inKonfig := range strategien {
					n++
					mitCursor++
					cursorFall(inKonfig, amCursor, k, erg, true).lauf(t)
				}
			}
		}
	}
	jeOhne := map[model.ReplayStrategy]int{}
	for _, inKonfig := range strategien {
		for _, erg := range erreichbar[inKonfig][0] {
			n++
			ohneCursor++
			jeOhne[inKonfig]++
			cursorFall(inKonfig, inKonfig, 0, erg, false).lauf(t)
		}
	}
	if len(strategien) != 4 || cursorZustaende != 3 || len(ergebnisse) != 5 || len(strategien)*cursorZustaende*len(ergebnisse) != 60 ||
		tripel != 43 || je[model.StrategySequential] != 11 || je[model.StrategyStrict] != 14 || je[model.StrategyFirst] != 9 || je[model.StrategyLast] != 9 ||
		mitCursor != 172 || ohneCursor != 13 || jeOhne[model.StrategySequential] != 3 || jeOhne[model.StrategyStrict] != 4 || jeOhne[model.StrategyFirst] != 3 ||
		jeOhne[model.StrategyLast] != 3 || n != 185 {
		t.Errorf("Menge: %d Tripel (je Strategie %v), %d mit Cursor, %d ohne Cursor (je Strategie %v), %d Elemente; erwartet 43 (11, 14, 9, 9), 172, 13 (3, 4, 3, 3) und 185",
			tripel, je, mitCursor, ohneCursor, jeOhne, n)
	}
	nicht := map[string]int{}
	for _, s := range strategien {
		for k := range cursorZustaende {
			for _, erg := range ergebnisse {
				if slices.Contains(erreichbar[s][k], erg) {
					continue
				}
				switch {
				case erg == ergAndererKey && s != model.StrategyStrict:
					nicht["anderer_key"]++
				case erg == ergErschoepft && (s == model.StrategyFirst || s == model.StrategyLast):
					nicht["erschoepft bei first und last"]++
				case erg == ergErschoepft && k == 0:
					nicht["erschoepft im Zustand neu"]++
				default:
					nicht["sonst"]++
				}
			}
		}
	}
	if len(nicht) != 3 || nicht["anderer_key"] != 9 || nicht["erschoepft bei first und last"] != 6 || nicht["erschoepft im Zustand neu"] != 2 {
		t.Errorf("nicht erreichbar: %v, erwartet 9 anderer_key, 6 erschoepft bei first und last, 2 erschoepft im Zustand neu und sonst keins", nicht)
	}
}

// cursorFall ist ein Element von M-Cursor: Strategie der Konfig inKonfig, mit Cursor der Strategie amCursor im Zustand k (sonst ohne Cursor),
// das Ergebnis erg; erwartet ist die Strategie des Cursors, ohne Cursor die der Konfig, das Ergebnis und ein Set.
func cursorFall(inKonfig, amCursor model.ReplayStrategy, k int, erg string, mitCursor bool) spanFall {
	name := fmt.Sprintf("Cursor %s, Konfig %s, k %d, %s", amCursor, inKonfig, k, erg)
	if !mitCursor {
		name = fmt.Sprintf("ohne Cursor, Konfig %s, %s", inKonfig, erg)
	}
	s := soll{}
	if erg != ergTreffer {
		s.code = codeZu(erg)
	}
	return spanFall{name: name, strat: inKonfig, strategie: string(amCursor), ergebnis: erg, set: true, soll: s, baue: func(u *umg) inbound.Anfrage {
		a := cursorLage(u, erg, k)
		if mitCursor {
			u.cursor(amCursor, verbraucht(k)...)
		}
		return a
	}}
}

// TestCursorTripelErreichbar hält die Zählprüfung von M-Cursor gegen die Quelle: waehle über alle Bestände mit 1 bis 3 Exchanges (SequenceNo 1 bis
// n lückenlos, je Exchange Key A oder Key B und replayfähig oder nicht: 4 + 16 + 64 gleich 84), alle 4 Strategien, alle k von 0 bis n
// (consumed {1..k}, position k) und den Key A des Requests erreicht in den Zuständen k gleich 0, 1 und 2 genau die Tripel von
// cursorErreichbar, 43 an der Zahl; der Zustand k gleich 3, den die Menge nicht führt (Prüfgrenze), erreicht 10 weitere. Das Ergebnis ist
// das des Codes des Fehlers, nicht ergebnisVon. Bestände mit mehr als drei Exchanges sind nicht geführt.
func TestCursorTripelErreichbar(t *testing.T) {
	type tripel struct {
		strat model.ReplayStrategy
		k     int
		erg   string
	}
	u := neu(t)
	keyA := u.keyA()
	erreicht, jenseits := map[tripel]bool{}, map[tripel]bool{}
	bestaende := 0
	for n := 1; n <= 3; n++ {
		for kodiert := range 1 << (2 * n) {
			var liste []model.Exchange
			for i := range n {
				v := kodiert >> (2 * i) & 3
				key := keyA
				if v&1 == 1 {
					key = u.keyB()
				}
				liste = append(liste, u.austausch(i+1, key, v&2 == 0))
			}
			bestaende++
			for _, s := range cursorStrategien() {
				for k := 0; k <= n; k++ {
					_, err := replay.Waehle(s, keyA, liste, n, false, verbraucht(k), int64(k))
					erg := ergTreffer
					if err != nil {
						var f *inbound.Fehler
						if !errors.As(err, &f) {
							t.Fatalf("Fehler der Auswahl ohne Code: %v", err)
						}
						erg = ""
						for _, e := range cursorErgebnisse()[1:] {
							if codeZu(e) == f.Code {
								erg = e
							}
						}
						if erg == "" {
							t.Fatalf("Code %s der Auswahl ist keinem Ergebnis zugeordnet", f.Code)
						}
					}
					if k >= cursorZustaende {
						jenseits[tripel{s, k, erg}] = true
						continue
					}
					erreicht[tripel{s, k, erg}] = true
				}
			}
		}
	}
	literal := map[tripel]bool{}
	imZustandNeu, erreichtImZustandNeu := 0, 0
	for s, je := range cursorErreichbar() {
		for k, ergebnisse := range je {
			for _, e := range ergebnisse {
				literal[tripel{s, k, e}] = true
				if k == 0 {
					imZustandNeu++
				}
			}
		}
	}
	for tr := range erreicht {
		if !literal[tr] {
			t.Errorf("Tripel %+v erreicht, steht nicht im Literal", tr)
		}
		if tr.k == 0 {
			erreichtImZustandNeu++
		}
	}
	for tr := range literal {
		if !erreicht[tr] {
			t.Errorf("Tripel %+v steht im Literal, wird nicht erreicht", tr)
		}
	}
	if bestaende != 84 || len(erreicht) != 43 || len(literal) != 43 || imZustandNeu != 13 || erreichtImZustandNeu != 13 || len(jenseits) != 10 {
		t.Errorf("Menge: %d Bestände, %d erreichte und %d im Literal, %d im Zustand neu (erreicht %d), %d bei drei Verbräuchen; erwartet 84, 43, 43, 13, 13 und 10",
			bestaende, len(erreicht), len(literal), imZustandNeu, erreichtImZustandNeu, len(jenseits))
	}
}

// TestSpanNachDerAuswahl hält W7 (SPEC-296 Nr. 4 b): das Ergebnis bleibt treffer, wenn nach der Auswahl der Aufruf scheitert; Menge
// M-NachAuswahl, 2 Zustände des Bodys, 2 Fehler von Open, 3 Fehler von Put, 1 Fehler der Bestätigung, 2 Stellen × 2 Arten des Kontextfehlers
// gleich 12.
func TestSpanNachDerAuswahl(t *testing.T) {
	treffer := func(u *umg) { u.lege(u.austausch(1, u.keyA(), true)) }
	mit := func(setze func(u *umg), c model.Code, k error) spanFall {
		return spanFall{baue: func(u *umg) inbound.Anfrage { treffer(u); setze(u); return anf() }, soll: soll{code: c, kontext: k},
			strategie: "sequential", ergebnis: ergTreffer, set: true}
	}
	zustand := func(z model.BodyState) func(u *umg) {
		return func(u *umg) { e := u.austausch(1, u.keyA(), true); e.ResponseBody = z; u.lege(e) }
	}
	faelle := map[string]spanFall{
		"Body discarded":     mit(zustand(model.BodyDiscarded), model.CodeE9001, nil),
		"Body unbekannt":     mit(zustand("x"), model.CodeE9001, nil),
		"Open ErrBlobFehlt":  mit(func(u *umg) { u.p.setzeFehler("Blobs.Open", outbound.ErrBlobFehlt) }, model.CodeE3001, nil),
		"Open Beschaedigt":   mit(func(u *umg) { u.p.setzeFehler("Blobs.Open", outbound.ErrBlobBeschaedigt) }, model.CodeE3001, nil),
		"Put NichtGefunden":  mit(func(u *umg) { u.p.setzeFehler("tx Cursors.Put", outbound.ErrNichtGefunden) }, model.CodeE7003, nil),
		"Put FremderService": mit(func(u *umg) { u.p.setzeFehler("tx Cursors.Put", outbound.ErrFremderService) }, model.CodeE5004, nil),
		"Put anderer":        mit(func(u *umg) { u.p.setzeFehler("tx Cursors.Put", errors.New("andere")) }, model.CodeE9001, nil),
		"Bestätigung":        mit(func(u *umg) { u.st.commit = errors.New("commit") }, model.CodeE9001, nil),
	}
	for _, k := range kontextfehler() {
		faelle["Kontext Open "+k.Error()] = mit(func(u *umg) { u.p.setzeFehler("Blobs.Open", k) }, 0, k)
		faelle["Kontext Put "+k.Error()] = mit(func(u *umg) { u.p.setzeFehler("tx Cursors.Put", k) }, 0, k)
	}
	for name, f := range faelle {
		f.name = name
		f.lauf(t)
	}
	if len(faelle) != 12 {
		t.Errorf("Menge: %d, erwartet 12", len(faelle))
	}
}

// TestSpanJederRueckweg hält W3 (SPEC-296 Nr. 3 b, 6): Menge M-Rückwege-Replay, 52 Fälle: S1 bis S8 mit Fehler mit Code (8), mit Kontextfehler
// (8 Stellen × 2 Arten gleich 16), die elf Entscheidungen des Dienstes ohne Port (11), die Rückkehr mit Antwort (1) und das Paar aus
// Kontextfehler und Fehler der Tabelle (errors.Join) an den 8 Stellen (16); je Fall genau ein StartSpan und ein End und der Ausgang unverändert
// (Code, Kontextfehler ohne Code oder Antwort). Zählprüfung 8, 2, 11 als Literal (SPEC-135 Nr. 18, SPEC-296 Nr. 6, SPEC-109 Nr. 6).
func TestSpanJederRueckweg(t *testing.T) {
	stellen8 := []string{"Sessions.Get", "", "Exchanges.ListByMatchKey", "Transaktion.WithWorkspaceTransaction","tx Workspaces.Get", "tx Cursors.Get",
		"Blobs.Open", "tx Cursors.Put"}
	var faelle []spanFall
	for _, op := range stellen8 {
		cause := error(outbound.ErrBeschaedigt)
		code := model.CodeE3001
		switch op {
		case "":
			cause, code = io.ErrUnexpectedEOF, model.CodeE5001
		case "Blobs.Open":
			cause = outbound.ErrBlobFehlt
		}
		faelle = append(faelle, spanFall{name: "Code " + op, baue: kontextAn(op, cause), soll: soll{code: code}})
		for _, k := range kontextfehler() {
			faelle = append(faelle,
				spanFall{name: "Kontext " + op + " " + k.Error(), baue: kontextAn(op, k), soll: soll{kontext: k}},
				spanFall{name: "Paar " + op + " " + k.Error(), baue: kontextAn(op, errors.Join(k, cause)), soll: soll{kontext: k}})
		}
	}
	entscheidungen := []spanFall{
		{name: "E5004 Session", baue: func(u *umg) inbound.Anfrage {
			u.sess.setze(model.Session{ID: sesID, ServiceID: "svc-b", CreatedAt: zeit(), Status: model.SessionComplete})
			return anf()
		}, soll: soll{code: model.CodeE5004}},
		{name: "E7009 Session", baue: func(u *umg) inbound.Anfrage {
			u.sess.setze(model.Session{ID: sesID, ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionOpen})
			return anf()
		}, soll: soll{code: model.CodeE7009}},
		{name: "E5001 Body", baue: func(*umg) inbound.Anfrage {
			a := anf()
			a.Header["Content-Encoding"] = []string{"gzip"}
			a.Body, a.Laenge = strings.NewReader("x"), 1
			return a
		}, soll: soll{code: model.CodeE5001}},
		{name: "E5006 Key", baue: func(u *umg) inbound.Anfrage {
			u.k.Regeln.InvalidEncoding = model.KodierungReject
			a := anf()
			a.RawQuery = "a=%zz"
			return a
		}, soll: soll{code: model.CodeE5006}},
		{name: "E9001 Key", baue: func(*umg) inbound.Anfrage { a := anf(); a.Pfad = root + "/a?b"; return a }, soll: soll{code: model.CodeE9001}},
		{name: "E5001 Auswahl", baue: func(*umg) inbound.Anfrage { return anf() }, soll: soll{code: model.CodeE5001}},
		{name: "E5002 Auswahl", baue: func(u *umg) inbound.Anfrage { bauePaar(u, model.StrategySequential, ergErschoepft); return anf() }, soll: soll{code: model.CodeE5002}},
		{name: "E5007 Auswahl", baue: func(u *umg) inbound.Anfrage {
			u.k.Strategie = model.StrategyStrict
			bauePaar(u, model.StrategyStrict, ergAndererKey)
			return anf()
		},
			soll: soll{code: model.CodeE5007}},
		{name: "E5008 Auswahl", baue: func(u *umg) inbound.Anfrage { bauePaar(u, model.StrategySequential, ergNichtReplayf); return anf() }, soll: soll{code: model.CodeE5008}},
		{name: "E5004 Workspace", baue: func(u *umg) inbound.Anfrage {
			z := u.st.stand()
			w := z.ws[wsID]
			w.ServiceID = "svc-b"
			z.ws[wsID] = w
			u.st.setze(z)
			return erfolgsLauf(u)
		}, soll: soll{code: model.CodeE5004}},
		{name: "E9001 Zustand", baue: func(u *umg) inbound.Anfrage {
			e := u.austausch(1, u.keyA(), true)
			e.ResponseBody = model.BodyDiscarded
			u.lege(e)
			return anf()
		}, soll: soll{code: model.CodeE9001}},
	}
	antwort := spanFall{name: "Antwort", baue: func(u *umg) inbound.Anfrage { u.lege(u.austausch(1, u.keyA(), true)); return anf() }}
	faelle = append(faelle, entscheidungen...)
	faelle = append(faelle, antwort)
	for _, f := range faelle {
		u := neu(t)
		u.mitJournal()
		a := f.baue(u)
		resp, err := u.replay(context.Background(), a)
		if err == nil {
			_ = resp.Body.Close()
		}
		pruefeAusgang(t, f.name, err, f.soll)
		b := buecher(u.tel)
		if len(b) != 1 || b[0].enden != 1 || u.tel.Anzahl(supporttest.OpStartSpan) != 1 || u.tel.Anzahl(supporttest.OpEnd) != 1 {
			t.Errorf("%s: Bücher %+v, erwartet ein StartSpan und ein End", f.name, b)
		}
	}
	if len(stellen8) != 8 || len(kontextfehler()) != 2 || len(entscheidungen) != 11 || len(faelle) != 52 {
		t.Errorf("Menge: %d Stellen, %d Entscheidungen, %d Fälle; erwartet 8, 11 und 52", len(stellen8), len(entscheidungen), len(faelle))
	}
}

// beendetesCtx liefert ein ctx, das vor dem Aufruf beendet ist: Canceled über WithCancel und cancel, DeadlineExceeded über WithDeadline in der
// Vergangenheit.
func beendetesCtx(t *testing.T, art error) context.Context {
	t.Helper()
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if errors.Is(art, context.DeadlineExceeded) {
		ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	} else {
		ctx, cancel = context.WithCancel(context.Background())
		cancel()
	}
	t.Cleanup(cancel)
	if ctx.Err() != art {
		t.Fatalf("ctx.Err() ist %v, erwartet %v", ctx.Err(), art)
	}
	return ctx
}

// spanStand ist der Stand des Spans nach einem Aufruf: Strategie und Ergebnis am Ende und ob ein Set steht.
type spanStand struct {
	strategie, ergebnis string
	set                 bool
}

// TestSpanAbbruch hält W3 (SPEC-296 Nr. 3 b, f) für den Zustand des ctx des Aufrufs: Menge M-Abbruch-Replay mit 18 Elementen, in der das ctx
// vor dem Aufruf beendet ist (Canceled über WithCancel und cancel, DeadlineExceeded über WithDeadline in der Vergangenheit) und die Fakes der
// Ports und der Body den Fehler dieses ctx liefern: die acht Stellen S1 bis S8 (S2 über den Body) × 2 Arten gleich 16, dazu die Rückkehr mit
// Antwort × 2 Arten gleich 2. Die Fakes prüfen das ctx hier nicht selbst, so dass allein die eingespeiste Stelle scheitert und jede frühere
// und spätere Operation das beendete ctx unverändert erhält. Je Fall genau ein StartSpan und ein End, ein Set genau dann, wenn S4 gerufen
// wurde, und der Ausgang unverändert (Kontextfehler ohne Code oder Antwort). Zählprüfung 8, 2 und 18.
func TestSpanAbbruch(t *testing.T) {
	stellen8 := []string{"Sessions.Get", "", "Exchanges.ListByMatchKey", "Transaktion.WithWorkspaceTransaction","tx Workspaces.Get", "tx Cursors.Get",
		"Blobs.Open", "tx Cursors.Put"}
	vorS4 := spanStand{nichtGelesen, nichtGewaehlt, false}
	inS4 := spanStand{nichtGelesen, nichtGewaehlt, true}
	nachAuswahl := spanStand{"sequential", ergTreffer, true}
	erwartung := map[string]spanStand{"Sessions.Get": vorS4, "": vorS4, "Exchanges.ListByMatchKey": vorS4, "Transaktion.WithWorkspaceTransaction": inS4,
		"tx Workspaces.Get": inS4, "tx Cursors.Get": inS4, "Blobs.Open": nachAuswahl, "tx Cursors.Put": nachAuswahl}
	pruefe := func(u *umg, was string, resp inbound.Antwort, err error, s soll, stand spanStand) {
		t.Helper()
		if err == nil {
			_ = resp.Body.Close()
		}
		pruefeAusgang(t, was, err, s)
		pruefeSpan(t, was, u.tel, stand.strategie, stand.ergebnis, stand.set)
		if u.tel.Anzahl(supporttest.OpStartSpan) != 1 || u.tel.Anzahl(supporttest.OpEnd) != 1 {
			t.Errorf("%s: %d StartSpan und %d End, erwartet je 1", was, u.tel.Anzahl(supporttest.OpStartSpan), u.tel.Anzahl(supporttest.OpEnd))
		}
	}
	n := 0
	for _, art := range kontextfehler() {
		for _, op := range stellen8 {
			n++
			u := neu(t)
			u.mitJournal()
			ctx := beendetesCtx(t, art)
			u.p.ctxNichtPruefen()
			a := kontextAn(op, ctx.Err())(u)
			resp, err := u.replay(ctx, a)
			pruefe(u, "beendetes ctx ("+art.Error()+") an "+op, resp, err, soll{kontext: art}, erwartung[op])
		}
		n++
		u := neu(t)
		u.mitJournal()
		ctx := beendetesCtx(t, art)
		u.p.ctxNichtPruefen()
		resp, err := u.replay(ctx, erfolgsLauf(u))
		pruefe(u, "beendetes ctx ("+art.Error()+") mit Antwort", resp, err, soll{}, nachAuswahl)
	}
	if len(stellen8) != 8 || len(kontextfehler()) != 2 || n != 18 {
		t.Errorf("Menge: %d Stellen, %d Arten, %d Fälle; erwartet 8, 2 und 18", len(stellen8), len(kontextfehler()), n)
	}
}

// TestSpanWiederholung hält W4 (SPEC-296 Nr. 7 c): Menge M-Wiederholung mit 4 Fällen, je Strategie zwei Aufrufe nacheinander am selben Replayer
// und denselben Stores mit demselben Request und einem Exchange; sequential und strict ergeben treffer, dann erschoepft, first und last zweimal
// treffer. Je Aufruf ein eigenes Buch mit einem StartSpan, einem Set und einem End. Zählprüfung 4.
func TestSpanWiederholung(t *testing.T) {
	faelle := []struct {
		strat    model.ReplayStrategy
		zweiter  soll
		ergebnis string
	}{
		{model.StrategySequential, soll{code: model.CodeE5002}, ergErschoepft},
		{model.StrategyStrict, soll{code: model.CodeE5002}, ergErschoepft},
		{model.StrategyFirst, soll{}, ergTreffer},
		{model.StrategyLast, soll{}, ergTreffer},
	}
	for _, f := range faelle {
		u := neu(t)
		u.mitJournal()
		u.k.Strategie = f.strat
		u.lege(u.austausch(1, u.keyA(), true))
		r := u.replayer()
		for i, a := range []struct {
			soll     soll
			ergebnis string
		}{{soll{}, ergTreffer}, {f.zweiter, f.ergebnis}} {
			was := fmt.Sprintf("%s, Aufruf %d", f.strat, i+1)
			resp, err := r.Replay(context.Background(), anf())
			if err == nil {
				_ = resp.Body.Close()
			}
			pruefeAusgang(t, was, err, a.soll)
			b := buecher(u.tel)
			if len(b) != i+1 {
				t.Fatalf("%s: %d Bücher, erwartet %d", was, len(b), i+1)
			}
			pruefeBuch(t, was, b[i], string(f.strat), a.ergebnis, true)
		}
		for op, n := range map[supporttest.Operation]int{supporttest.OpStartSpan: 2, supporttest.OpSet: 2, supporttest.OpEnd: 2} {
			if got := u.tel.Anzahl(op); got != n {
				t.Errorf("%s: Operation %d %d-mal, erwartet %d", f.strat, op, got, n)
			}
		}
	}
	if len(faelle) != 4 {
		t.Errorf("Menge: %d, erwartet 4", len(faelle))
	}
}
