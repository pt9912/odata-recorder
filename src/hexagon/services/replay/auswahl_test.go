package replay_test

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// layouts liefert alle Layouts der Länge 0 bis 3 über {A, a, B, b}: A und B sind replayfähig mit Key A bzw. B,
// a und b haben capture_status failed (85 Layouts).
func layouts() []string {
	alle := []string{""}
	var stufe = []string{""}
	for n := 1; n <= 3; n++ {
		var next []string
		for _, l := range stufe {
			for _, c := range "AaBb" {
				next = append(next, l+string(c))
			}
		}
		alle = append(alle, next...)
		stufe = next
	}
	return alle
}

// legeLayout legt die Exchanges des Layouts an: SequenceNo i+1.
func legeLayout(u *umg, l string) {
	var es []model.Exchange
	for i, c := range l {
		key := u.keyA()
		if c == 'B' || c == 'b' {
			key = u.keyB()
		}
		es = append(es, u.austausch(i+1, key, c == 'A' || c == 'B'))
	}
	u.lege(es...)
}

// anfrageFuer liefert den Request für den Key k aus {A, B, C}.
func anfrageFuer(k byte) inbound.Anfrage {
	switch k {
	case 'A':
		return anf()
	case 'B':
		return anfB()
	}
	return anfC()
}

// ergebnis ist das Orakel einer Auswahl: seq > 0 der gewählte Exchange, sonst der Code.
type ergebnis struct {
	seq  int
	code model.Code
}

func (e ergebnis) String() string {
	if e.seq > 0 {
		return fmt.Sprintf("SequenceNo %d", e.seq)
	}
	return e.code.String()
}

// schluesselVon sagt, welchen Key das Zeichen des Layouts trägt.
func schluesselVon(c byte) byte { return map[byte]byte{'A': 'A', 'a': 'A', 'B': 'B', 'b': 'B'}[c] }
func gutVon(c byte) bool        { return c == 'A' || c == 'B' }

// erwartetSequential ist das Orakel von SPEC-135 Nr. 9 für sequential, aus den Sätzen der Spezifikation
// gebildet: M die Kandidaten, U die unverbrauchten, gewählt die kleinste SequenceNo in U ∩ R.
func erwartetSequential(l string, verbraucht map[int]bool, key byte) ergebnis {
	var m, u, ur []int
	for i := 0; i < len(l); i++ {
		if schluesselVon(l[i]) != key {
			continue
		}
		m = append(m, i+1)
		if verbraucht[i+1] {
			continue
		}
		u = append(u, i+1)
		if gutVon(l[i]) {
			ur = append(ur, i+1)
		}
	}
	switch {
	case len(m) == 0:
		return ergebnis{code: model.CodeE5001}
	case len(ur) > 0:
		return ergebnis{seq: slices.Min(ur)}
	case len(u) > 0:
		return ergebnis{code: model.CodeE5008}
	}
	return ergebnis{code: model.CodeE5002}
}

// erwartetStrict: M leer ist E5001; sonst entscheidet der Exchange mit SequenceNo p+1 der Reihe nach.
func erwartetStrict(l string, p int, key byte) ergebnis {
	m := 0
	for i := 0; i < len(l); i++ {
		if schluesselVon(l[i]) == key {
			m++
		}
	}
	switch {
	case m == 0:
		return ergebnis{code: model.CodeE5001}
	case p >= len(l):
		return ergebnis{code: model.CodeE5002}
	case schluesselVon(l[p]) != key:
		return ergebnis{code: model.CodeE5007}
	case !gutVon(l[p]):
		return ergebnis{code: model.CodeE5008}
	}
	return ergebnis{seq: p + 1}
}

// erwartetFirstLast: kleinste bzw. größte replayfähige SequenceNo unter den Kandidaten.
func erwartetFirstLast(l string, key byte, letzter bool) ergebnis {
	var m, r []int
	for i := 0; i < len(l); i++ {
		if schluesselVon(l[i]) != key {
			continue
		}
		m = append(m, i+1)
		if gutVon(l[i]) {
			r = append(r, i+1)
		}
	}
	switch {
	case len(m) == 0:
		return ergebnis{code: model.CodeE5001}
	case len(r) == 0:
		return ergebnis{code: model.CodeE5008}
	case letzter:
		return ergebnis{seq: slices.Max(r)}
	}
	return ergebnis{seq: slices.Min(r)}
}

// pruefeErgebnis vergleicht Antwort oder Fehler mit dem Orakel; bei einem Fehler zählen weder Put noch Open.
func pruefeErgebnis(t *testing.T, was string, u *umg, resp inbound.Antwort, err error, want ergebnis) bool {
	t.Helper()
	if want.seq > 0 {
		if err != nil {
			t.Errorf("%s: Fehler %v, erwartet %v", was, err, want)
			return false
		}
		if got := seqVon(t, resp); got != want.seq {
			t.Errorf("%s: SequenceNo %d, erwartet %v", was, got, want)
			return false
		}
		_ = resp.Body.Close()
		return true
	}
	erwarteCode(t, was, err, want.code)
	if u.p.anzahl("tx Cursors.Put") != 0 || u.p.anzahl("Blobs.Open") != 0 {
		t.Errorf("%s: Put oder Open bei Fehler", was)
	}
	return false
}

// teilmengen liefert alle Teilmengen der SequenceNo 1..n als aufsteigende Listen.
func teilmengen(n int) [][]int64 {
	var r [][]int64
	for mask := 0; mask < 1<<n; mask++ {
		var s []int64
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				s = append(s, int64(i+1))
			}
		}
		r = append(r, s)
	}
	return r
}

func menge(s []int64) map[int]bool {
	m := map[int]bool{}
	for _, x := range s {
		m[int(x)] = true
	}
	return m
}

// TestAuswahlSequential hält Z26: Menge 1755 (85 Layouts × 585/85 verbrauchte Teilmengen × 3 Keys).
func TestAuswahlSequential(t *testing.T) {
	n := 0
	for _, l := range layouts() {
		for _, s := range teilmengen(len(l)) {
			for _, key := range []byte("ABC") {
				n++
				u := neu(t)
				legeLayout(u, l)
				if len(s) > 0 {
					u.cursor(model.StrategySequential, s...)
				}
				was := fmt.Sprintf("Layout %q verbraucht %v Key %c", l, s, key)
				resp, err := u.replay(context.Background(), anfrageFuer(key))
				want := erwartetSequential(l, menge(s), key)
				if !pruefeErgebnis(t, was, u, resp, err, want) {
					continue
				}
				c, da := u.cursorStand()
				neu := append(slices.Clone(s), int64(want.seq))
				if !da || c.Position != int64(len(neu)) || !reflect.DeepEqual(menge(c.Consumed), menge(neu)) || c.Strategy != model.StrategySequential {
					t.Errorf("%s: Cursor %+v, erwartet verbraucht %v", was, c, neu)
				}
			}
		}
	}
	if n != 1755 {
		t.Errorf("Menge: %d Elemente, erwartet 1755", n)
	}
}

// TestAuswahlStrict hält Z27: Menge 939 (85 Layouts × position 0..n × 3 Keys).
func TestAuswahlStrict(t *testing.T) {
	n := 0
	for _, l := range layouts() {
		for p := 0; p <= len(l); p++ {
			for _, key := range []byte("ABC") {
				n++
				u := neu(t)
				u.k.Strategie = model.StrategyStrict
				legeLayout(u, l)
				var verbraucht []int64
				for i := 1; i <= p; i++ {
					verbraucht = append(verbraucht, int64(i))
				}
				if p > 0 {
					u.cursor(model.StrategyStrict, verbraucht...)
				}
				was := fmt.Sprintf("Layout %q position %d Key %c", l, p, key)
				resp, err := u.replay(context.Background(), anfrageFuer(key))
				want := erwartetStrict(l, p, key)
				if !pruefeErgebnis(t, was, u, resp, err, want) {
					continue
				}
				if c, da := u.cursorStand(); !da || c.Position != int64(p+1) || !slices.Contains(c.Consumed, int64(want.seq)) {
					t.Errorf("%s: Cursor %+v", was, c)
				}
			}
		}
	}
	if n != 939 {
		t.Errorf("Menge: %d Elemente, erwartet 939", n)
	}
}

// TestAuswahlFirstLast hält Z28: Menge 3510 (2 Strategien × 585 × 3 Keys); consumed ohne Wirkung, Put 0.
func TestAuswahlFirstLast(t *testing.T) {
	n := 0
	for _, strat := range []model.ReplayStrategy{model.StrategyFirst, model.StrategyLast} {
		for _, l := range layouts() {
			for _, s := range teilmengen(len(l)) {
				for _, key := range []byte("ABC") {
					n++
					u := neu(t)
					u.k.Strategie = strat
					legeLayout(u, l)
					if len(s) > 0 {
						u.cursor(strat, s...)
					}
					vorher := u.st.stand()
					was := fmt.Sprintf("%s Layout %q verbraucht %v Key %c", strat, l, s, key)
					resp, err := u.replay(context.Background(), anfrageFuer(key))
					pruefeErgebnis(t, was, u, resp, err, erwartetFirstLast(l, key, strat == model.StrategyLast))
					if u.p.anzahl("tx Cursors.Put") != 0 || !reflect.DeepEqual(u.st.stand(), vorher) {
						t.Errorf("%s: Cursor geschrieben", was)
					}
				}
			}
		}
	}
	if n != 3510 {
		t.Errorf("Menge: %d Elemente, erwartet 3510", n)
	}
}

// TestLeererKey hält Z24: Menge 8; ein leerer MatchKey passt nie. Im Modus strict steht der Exchange ohne
// Key an Position 1 und ist deshalb ein anderer Key (ODR-E5007, SPEC-135 Nr. 9).
func TestLeererKey(t *testing.T) {
	strategien := []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast}
	n := 0
	for _, s := range strategien {
		for _, vorA := range []bool{false, true} {
			n++
			u := neu(t)
			u.k.Strategie = s
			ohne := u.austausch(1, "", false)
			if vorA {
				u.lege(ohne, u.austausch(2, u.keyA(), true))
			} else {
				u.lege(ohne)
			}
			resp, err := u.replay(context.Background(), anf())
			want := ergebnis{code: model.CodeE5001}
			if vorA {
				want = ergebnis{seq: 2}
				if s == model.StrategyStrict {
					want = ergebnis{code: model.CodeE5007}
				}
			}
			pruefeErgebnis(t, fmt.Sprintf("%s vorA=%v", s, vorA), u, resp, err, want)
		}
	}
	if n != 8 {
		t.Errorf("Menge: %d, erwartet 8", n)
	}
}

// TestListBySession hält Z25: Menge 12; genau einmal, mit der Kennung unverändert, vor der Transaktion; keine
// Ordnung der Liste angenommen.
func TestListBySession(t *testing.T) {
	ordnungen := map[string][]int{"aufsteigend": {1, 2, 3}, "absteigend": {3, 2, 1}, "gemischt": {2, 3, 1}}
	erwartet := map[model.ReplayStrategy]int{model.StrategySequential: 1, model.StrategyStrict: 1, model.StrategyFirst: 1, model.StrategyLast: 3}
	n := 0
	for name, ord := range ordnungen {
		for s, seq := range erwartet {
			n++
			u := neu(t)
			u.k.Strategie = s
			var es []model.Exchange
			for _, i := range ord {
				es = append(es, u.austausch(i, u.keyA(), true))
			}
			u.lege(es...)
			vorTx := -1
			u.p.setzeNach(func(op string) {
				if op == "Exchanges.ListBySession" {
					vorTx = u.p.anzahl("Transaktion.WithWorkspaceTransaction")
					for _, o := range u.p.ops() {
						if strings.HasPrefix(o, "tx ") {
							vorTx = 99
						}
					}
				}
			})
			resp, err := u.replay(context.Background(), anf())
			pruefeErgebnis(t, name+" "+string(s), u, resp, err, ergebnis{seq: seq})
			evs := u.p.von("Exchanges.ListBySession")
			if len(evs) != 1 || evs[0].args[0] != sesID {
				t.Errorf("%s %s: ListBySession %v", name, s, evs)
			}
			if vorTx != 0 {
				t.Errorf("%s %s: ListBySession lief nicht vor der Transaktion (%d)", name, s, vorTx)
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge: %d, erwartet 12", n)
	}
}

// folge ruft Replay mit a n-mal und liefert je Aufruf SequenceNo oder Code.
func folge(t *testing.T, u *umg, as ...inbound.Anfrage) []string {
	t.Helper()
	var r []string
	for _, a := range as {
		resp, err := u.replay(context.Background(), a)
		if err != nil {
			c, _, ok := fehlerCode(err)
			if !ok {
				t.Fatalf("Fehler ohne Code: %v", err)
			}
			r = append(r, c.String())
			continue
		}
		r = append(r, fmt.Sprint(seqVon(t, resp)))
		_ = resp.Body.Close()
	}
	return r
}

// TestBeispieleLH hält Z29: sechs benannte Läufe der Beispiele von LH-FA-04.
func TestBeispieleLH(t *testing.T) {
	aba := func(u *umg) {
		u.lege(u.austausch(1, u.keyA(), true), u.austausch(2, u.keyB(), true), u.austausch(3, u.keyA(), true))
	}
	lauf := func(name string, strat model.ReplayStrategy, aufbau func(u *umg), as []inbound.Anfrage, want ...string) {
		u := neu(t)
		u.k.Strategie = strat
		aufbau(u)
		if got := folge(t, u, as...); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, erwartet %v", name, got, want)
		}
	}
	lauf("sequential B, A, A, A", model.StrategySequential, aba, []inbound.Anfrage{anfB(), anf(), anf(), anf()}, "2", "1", "3", "ODR-E5002")
	lauf("sequential zwei gleiche", model.StrategySequential, func(u *umg) {
		u.lege(u.austausch(1, u.keyA(), true), u.austausch(2, u.keyA(), true))
	}, []inbound.Anfrage{anf(), anf()}, "1", "2")
	lauf("strict A, B, A", model.StrategyStrict, aba, []inbound.Anfrage{anf(), anfB(), anf()}, "1", "2", "3")
	lauf("strict B zuerst, dann A", model.StrategyStrict, aba, []inbound.Anfrage{anfB(), anf()}, "ODR-E5007", "1")
	lauf("first", model.StrategyFirst, aba, []inbound.Anfrage{anf(), anf(), anf()}, "1", "1", "1")
	lauf("last", model.StrategyLast, aba, []inbound.Anfrage{anf(), anf(), anf()}, "3", "3", "3")
	u := neu(t)
	u.lege(u.austausch(1, u.keyA(), false))
	if got := folge(t, u, anf(), anf()); !reflect.DeepEqual(got, []string{"ODR-E5008", "ODR-E5008"}) {
		t.Errorf("nur a(1): %v", got)
	}
	if _, da := u.cursorStand(); da {
		t.Error("nur a(1): Fortschritt trotz Fehler")
	}
	// Verbrauch nach SequenceNo des Exchange: B, A, A verbraucht 2, 1, 3 in dieser Reihenfolge.
	u = neu(t)
	aba(u)
	folge(t, u, anfB(), anf(), anf())
	if c, _ := u.cursorStand(); !reflect.DeepEqual(c.Consumed, []int64{2, 1, 3}) || c.Position != 3 {
		t.Errorf("Cursor %+v, erwartet consumed [2 1 3]", c)
	}
}

// TestStrategieDesCursors hält Z30: Menge 20 (Konfig × Cursor); die Strategie eines bestehenden Cursors gilt,
// ein neuer Cursor bekommt die der Konfig, first und last legen keinen an.
func TestStrategieDesCursors(t *testing.T) {
	strategien := []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast}
	// B(1), A(2), A(3), Key A: sequential 2 dann 3; strict ODR-E5007 zweimal; first 2 zweimal; last 3 zweimal.
	erwartet := map[model.ReplayStrategy][]string{
		model.StrategySequential: {"2", "3"}, model.StrategyStrict: {"ODR-E5007", "ODR-E5007"},
		model.StrategyFirst: {"2", "2"}, model.StrategyLast: {"3", "3"},
	}
	n := 0
	for _, konfig := range strategien {
		for _, vorhanden := range append([]model.ReplayStrategy{""}, strategien...) {
			n++
			u := neu(t)
			u.k.Strategie = konfig
			u.lege(u.austausch(1, u.keyB(), true), u.austausch(2, u.keyA(), true), u.austausch(3, u.keyA(), true))
			wirksam := konfig
			if vorhanden != "" {
				u.cursor(vorhanden)
				wirksam = vorhanden
			}
			was := fmt.Sprintf("Konfig %s Cursor %q", konfig, vorhanden)
			if got := folge(t, u, anf(), anf()); !reflect.DeepEqual(got, erwartet[wirksam]) {
				t.Errorf("%s: %v, erwartet %v", was, got, erwartet[wirksam])
			}
			c, da := u.cursorStand()
			switch {
			case wirksam == model.StrategySequential && (!da || c.Strategy != wirksam || c.Position != 2):
				t.Errorf("%s: Cursor %+v", was, c)
			case (wirksam == model.StrategyFirst || wirksam == model.StrategyLast) && vorhanden == "" && da:
				t.Errorf("%s: Cursor angelegt", was)
			case da && vorhanden != "" && c.Strategy != vorhanden:
				t.Errorf("%s: Strategie des Cursors geändert: %s", was, c.Strategy)
			}
		}
	}
	if n != 20 {
		t.Errorf("Menge: %d, erwartet 20", n)
	}
}

// TestPutFelder hält Z31: Menge 8 (4 Strategien × Cursor neu oder bestehend mit consumed {7, 9}); Put trägt
// genau die Felder, first und last rufen Put nie.
func TestPutFelder(t *testing.T) {
	n := 0
	for _, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast} {
		for _, bestehend := range []bool{false, true} {
			n++
			u := neu(t)
			u.k.Strategie = s
			u.lege(u.austausch(1, u.keyB(), true), u.austausch(2, u.keyA(), true), u.austausch(3, u.keyA(), true))
			wantSeq, wantConsumed := 2, []int64{2}
			if s == model.StrategyStrict {
				u.lege(u.austausch(1, u.keyA(), true), u.austausch(2, u.keyA(), true), u.austausch(3, u.keyA(), true))
				wantSeq, wantConsumed = 1, []int64{1}
			}
			if bestehend {
				u.cursor(s, 7, 9)
				wantConsumed = []int64{7, 9, int64(wantSeq)}
				if s == model.StrategyStrict {
					wantSeq, wantConsumed = 3, []int64{7, 9, 3}
				}
			}
			resp, err := u.replay(context.Background(), anf())
			if err != nil {
				t.Errorf("%s bestehend=%v: %v", s, bestehend, err)
				continue
			}
			_ = resp.Body.Close()
			puts := u.p.von("tx Cursors.Put")
			if s == model.StrategyFirst || s == model.StrategyLast {
				if len(puts) != 0 {
					t.Errorf("%s: Put gerufen", s)
				}
				continue
			}
			if len(puts) != 1 {
				t.Errorf("%s bestehend=%v: %d Put", s, bestehend, len(puts))
				continue
			}
			got := puts[0].args[0].(model.ReplayCursor)
			want := model.ReplayCursor{WorkspaceID: wsID, ReplaySessionID: repID, RecordingSessionID: sesID, ServiceID: svcID,
				Strategy: s, Position: int64(len(wantConsumed)), Consumed: wantConsumed}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s bestehend=%v: Put %+v, erwartet %+v (SequenceNo %d)", s, bestehend, got, want, wantSeq)
			}
			if err := got.Validate(); err != nil {
				t.Errorf("%s bestehend=%v: Cursor ungültig: %v", s, bestehend, err)
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge: %d, erwartet 8", n)
	}
}

// TestCursorIdentitaet hält Z32: Menge 4; je eine Komponente verschieden (die zweite Anfrage beginnt bei 0)
// und alle gleich (sie setzt fort).
func TestCursorIdentitaet(t *testing.T) {
	for _, komp := range []string{"Workspace", "Replay-Session", "Recording-Session", "alle gleich"} {
		u := neu(t)
		z := u.st.stand()
		z.ws["wsp-b"] = model.Workspace{ID: "wsp-b", ServiceID: svcID, BaselineSnapshotID: "snap-a", CreatedAt: zeit()}
		u.st.setze(z)
		u.sess.setze(model.Session{ID: "ses-b", ServiceID: svcID, CreatedAt: zeit(), Status: model.SessionComplete})
		es := []model.Exchange{u.austausch(1, u.keyA(), true), u.austausch(2, u.keyA(), true)}
		u.lege(es...)
		u.ex.setze("ses-b", es)
		erste, zweite := anf(), anf()
		switch komp {
		case "Workspace":
			zweite.Header["X-Recorder-Workspace"] = []string{"wsp-b"}
		case "Replay-Session":
			zweite.Header["X-Recorder-Replay-Session"] = []string{"rep-b"}
		case "Recording-Session":
			zweite.Header["X-Recorder-Recording-Session"] = []string{"ses-b"}
		}
		want := []string{"1", "1"}
		if komp == "alle gleich" {
			want = []string{"1", "2"}
		}
		if got := folge(t, u, erste, zweite); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, erwartet %v", komp, got, want)
		}
	}
}

// TestFehlerLassenCursorStehen hält Z33 und Z42: Menge 26 (13 Codes × 2 Strategien) und 4 Blob-Fehler; bei
// jedem Fehler ist der Store vorher und nachher gleich, Put und Advance werden nicht gerufen.
func TestFehlerLassenCursorStehen(t *testing.T) {
	n := 0
	for _, sz := range szenarien() {
		for _, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict} {
			n++
			u := neu(t)
			u.k.Strategie = s
			a := sz.baue(u, s)
			vorher := u.st.stand()
			_, err := u.replay(context.Background(), a)
			erwarteCode(t, fmt.Sprintf("%s %s", sz.code, s), err, sz.code)
			if !reflect.DeepEqual(u.st.stand(), vorher) {
				t.Errorf("%s %s: Store geändert", sz.code, s)
			}
			if u.p.anzahl("tx Cursors.Put") != 0 || u.p.anzahl("tx Workspaces.Advance") != 0 {
				t.Errorf("%s %s: Put oder Advance gerufen", sz.code, s)
			}
		}
	}
	if n != 26 {
		t.Errorf("Menge: %d, erwartet 26", n)
	}
	m := 0
	for _, fehler := range []error{blobFehlt(), blobBeschaedigt()} {
		for _, s := range []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict} {
			m++
			u := neu(t)
			u.k.Strategie = s
			erfolg(u)
			u.p.setzeFehler("Blobs.Open", fehler)
			vorher := u.st.stand()
			_, err := u.replay(context.Background(), anf())
			erwarteCode(t, "Blob "+string(s), err, model.CodeE3001)
			if !reflect.DeepEqual(u.st.stand(), vorher) || u.p.anzahl("tx Cursors.Put") != 0 {
				t.Errorf("Blob %s: Cursor verbraucht", s)
			}
		}
	}
	if m != 4 {
		t.Errorf("Menge Z42: %d, erwartet 4", m)
	}
}
