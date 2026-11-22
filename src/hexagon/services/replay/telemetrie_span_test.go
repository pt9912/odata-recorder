package replay_test

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
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
	f := []string{"Telemetrie.StartSpan", "Sessions.Get", "Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get",
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
		{"Fehler in f", model.StrategySequential, func(*umg) {}, []string{"Telemetrie.StartSpan", "Sessions.Get", "Exchanges.ListBySession",
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
		fall("S3 E7003", stoerung("Exchanges.ListBySession", ng), soll{code: model.CodeE7003}, false),
		fall("S3 E3001", stoerung("Exchanges.ListBySession", outbound.ErrBeschaedigt), soll{code: model.CodeE3001}, false),
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
		for _, op := range []string{"Sessions.Get", "", "Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get", "tx Cursors.Get"} {
			r = append(r, fall("Kontext "+op+" "+k.Error(), kontextAn(op, k), soll{kontext: k}, op != "" && op != "Sessions.Get" && op != "Exchanges.ListBySession"))
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

// TestSpanCursorStrategie hält W6 (SPEC-296 Nr. 4 a, 7 d): die Strategie des bestehenden Cursors gilt im Span, sonst die der Konfig; Menge
// M-Cursor, 4 Strategien am Cursor × eine abweichende der Konfig plus ein fehlender Cursor × 4 Strategien der Konfig gleich 8.
func TestSpanCursorStrategie(t *testing.T) {
	abweichend := map[model.ReplayStrategy]model.ReplayStrategy{model.StrategySequential: model.StrategyLast, model.StrategyStrict: model.StrategyFirst,
		model.StrategyFirst: model.StrategyStrict, model.StrategyLast: model.StrategySequential}
	n := 0
	for _, strat := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		n += 2
		spanFall{name: "Cursor " + string(strat), strat: abweichend[strat], strategie: string(strat), ergebnis: ergTreffer, set: true,
			baue: func(u *umg) inbound.Anfrage { u.lege(u.austausch(1, u.keyA(), true)); u.cursor(strat); return anf() }}.lauf(t)
		spanFall{name: "ohne Cursor " + string(strat), strat: strat, strategie: string(strat), ergebnis: ergTreffer, set: true,
			baue: func(u *umg) inbound.Anfrage { u.lege(u.austausch(1, u.keyA(), true)); return anf() }}.lauf(t)
	}
	if n != 8 {
		t.Errorf("Menge: %d, erwartet 8", n)
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
	stellen8 := []string{"Sessions.Get", "", "Exchanges.ListBySession", "Transaktion.WithWorkspaceTransaction", "tx Workspaces.Get", "tx Cursors.Get",
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
