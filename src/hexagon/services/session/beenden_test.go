package session_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// sid ist die Kennung der Session in den Tests des Beendens; sie gleicht nie dem Dienst (dienstA).
const sid = "r1"

// alleOps nennt jede Operation der Fakes und die Bezüge von Zeit und Kennung unter dem Namen, unter
// dem sie gezählt werden.
func alleOps() []string {
	return append(portsDerFakes(), "Record.Record", "Drain", "Append", "Exchanges.Get", "ListBySession", "Now", "SessionID")
}

// e4001 ist der Fehler von Drain, der einen Code trägt.
func e4001() *inbound.Fehler {
	c := model.CodeE4001
	return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: verpackt(outbound.ErrUpstreamNichtErreichbar)}
}

// neuBeenden liefert eine Umgebung mit der offenen Session sid im Dienst dienstA und Exchanges mit
// den genannten Capture-Status.
func neuBeenden(t *testing.T, capture ...model.CaptureStatus) *welt {
	t.Helper()
	u := neu(t)
	u.dienst(dienstA, dienstB)
	u.offene(sid, dienstA, 1)
	for _, c := range capture {
		u.exchange.haenge(sid, c)
	}
	return u
}

// ergebnis ist das Ergebnis eines Aufrufs von Beenden.
type ergebnis struct {
	s   model.Session
	err error
}

// startBeenden führt Beenden in einer Goroutine aus.
func (u *welt) startBeenden(ctx context.Context, id string) <-chan ergebnis {
	ch := make(chan ergebnis, 1)
	go func() {
		s, err := u.beenden(ctx, id)
		ch <- ergebnis{s, err}
	}()
	return ch
}

// abwarten wartet höchstens zeitgrenze auf das Ergebnis.
func abwarten(t *testing.T, ch <-chan ergebnis, was string) ergebnis {
	t.Helper()
	select {
	case r := <-ch:
		return r
	case <-time.After(zeitgrenze):
		t.Fatalf("%s: nach %v kein Ergebnis", was, zeitgrenze)
		return ergebnis{}
	}
}

// TestNewBeenden: New prüft Record, Exchanges und Frist; Frist 0 und 1 ns sind gültig, -1 ns nicht
// (B1, SPEC-124 Nr. 1).
func TestNewBeenden(t *testing.T) {
	u := neu(t)
	faelle := []struct {
		name    string
		aendere func(*session.Konfig)
		gueltig bool
	}{
		{"Record nil", func(k *session.Konfig) { k.Record = nil }, false},
		{"Exchanges nil", func(k *session.Konfig) { k.Exchanges = nil }, false},
		{"Frist -1ns", func(k *session.Konfig) { k.Frist = -time.Nanosecond }, false},
		{"Frist -1h", func(k *session.Konfig) { k.Frist = -time.Hour }, false},
		{"Frist 0", func(k *session.Konfig) { k.Frist = 0 }, true},
		{"Frist 1ns", func(k *session.Konfig) { k.Frist = time.Nanosecond }, true},
		{"Frist 7s", func(k *session.Konfig) { k.Frist = 7 * time.Second }, true},
		{"alle neun Felder", func(*session.Konfig) {}, true},
	}
	for _, f := range faelle {
		k := u.konfig
		f.aendere(&k)
		v, err := session.New(k)
		if f.gueltig {
			if err != nil || v == nil {
				t.Errorf("%s: %v, Dienst %v", f.name, err, v != nil)
			}
			continue
		}
		if !errors.Is(err, model.ErrUngueltig) || v != nil || !ohneCode(err) {
			t.Errorf("%s: %v, Dienst %v", f.name, err, v != nil)
		}
	}
}

// TestCtxNilBeenden: ein nil-ctx ist bei Beenden und CloseAll ErrUngueltig ohne Code, ohne Aufruf
// eines Ports, ohne Now und SessionID, ohne Protokoll (B3, SPEC-124 Nr. 2, 11 a, 14 a). Ein Aufruf
// der Fakes auf dem nil-ctx panict und macht den Fall rot.
func TestCtxNilBeenden(t *testing.T) {
	var keinCtx context.Context
	faelle := map[string]func(*welt) (bool, error){
		"Beenden": func(u *welt) (bool, error) {
			s, err := u.beenden(keinCtx, sid)
			return s == model.Session{}, err
		},
		"CloseAll": func(u *welt) (bool, error) {
			b, err := u.closeAll(keinCtx)
			return b == inbound.CloseBericht{}, err
		},
	}
	for name, f := range faelle {
		u := neuBeenden(t, model.CaptureComplete)
		leer, err := f(u)
		if !errors.Is(err, model.ErrUngueltig) || !ohneCode(err) || !leer {
			t.Errorf("%s: %v, Nullwert %v", name, err, leer)
		}
		for _, op := range alleOps() {
			if n := u.st.anzahl(op); n != 0 {
				t.Errorf("%s: %s %d-mal gerufen", name, op, n)
			}
		}
		if e := u.p.liste(); len(e) != 0 {
			t.Errorf("%s: Ereignisse %v", name, e)
		}
		if n := len(u.log.alle()); n != 0 {
			t.Errorf("%s: %d Einträge im Protokoll", name, n)
		}
	}
}

// TestBeendenPruefung: unbekannt ODR-E7003, dann nicht offen ODR-E7011, beide vor Drain; die Kennung
// kommt unverändert an Get (B4, SPEC-124 Nr. 11 b, c).
func TestBeendenPruefung(t *testing.T) {
	bestand := func(u *welt) {
		u.dienst(dienstA)
		u.sessions.fuege("Ab c", dienstA, erzeugt(1), model.SessionComplete)
		u.sessions.fuege(" r1 ", dienstA, erzeugt(2), model.SessionComplete)
		u.sessions.fuege("r1", dienstA, erzeugt(3), model.SessionOpen)
		u.sessions.fuege("c1", dienstA, erzeugt(4), model.SessionComplete)
		u.sessions.fuege("i1", dienstA, erzeugt(5), model.SessionIncomplete)
		u.sessions.fuege("x1", dienstA, erzeugt(6), model.SessionStatus("x"))
	}
	faelle := []struct{ id, code string }{
		{"", "ODR-E7003"}, {"unbekannt", "ODR-E7003"},
		{"Ab c", "ODR-E7011"}, {" r1 ", "ODR-E7011"},
		{"c1", "ODR-E7011"}, {"i1", "ODR-E7011"}, {"x1", "ODR-E7011"},
	}
	for _, f := range faelle {
		u := neu(t)
		bestand(u)
		vorher, _ := u.sessions.hat(f.id)
		s, err := u.beenden(context.Background(), f.id)
		code, status := codeVon(t, err)
		if code != f.code || s != (model.Session{}) {
			t.Errorf("id %q: %s, Session %v", f.id, code, s)
		}
		if f.code == "ODR-E7011" && (status != 409 || errors.Unwrap(err) != nil) {
			t.Errorf("id %q: Status %d, Ursache %v", f.id, status, errors.Unwrap(err))
		}
		if e := u.p.liste(); !slices.Equal(e, []string{"Sessions.Get(" + f.id + ")"}) {
			t.Errorf("id %q: Ereignisse %v", f.id, e)
		}
		for _, op := range []string{"Drain", "ListBySession", "SetStatus"} {
			if n := u.st.anzahl(op); n != 0 {
				t.Errorf("id %q: %s %d-mal", f.id, op, n)
			}
		}
		if nachher, _ := u.sessions.hat(f.id); nachher != vorher {
			t.Errorf("id %q: Session geändert: %v zu %v", f.id, vorher, nachher)
		}
	}
	// die offene Session geht weiter.
	u := neu(t)
	bestand(u)
	if _, err := u.beenden(context.Background(), "r1"); err != nil {
		t.Errorf("offene Session: %v", err)
	}
	if n := u.st.anzahl("Drain"); n != 1 {
		t.Errorf("offene Session: Drain %d-mal", n)
	}
}

// TestBeenden: Get, Drain, ListBySession und SetStatus je genau einmal in dieser Reihenfolge, mit
// der Kennung der Session; das Ergebnis ist die gelesene Session mit dem neuen Status (B5).
func TestBeenden(t *testing.T) {
	faelle := []struct {
		name      string
		exchanges []model.CaptureStatus
		soll      model.SessionStatus
	}{
		{"complete", []model.CaptureStatus{model.CaptureComplete, model.CaptureComplete}, model.SessionComplete},
		{"incomplete", []model.CaptureStatus{model.CaptureComplete, model.CaptureIncomplete}, model.SessionIncomplete},
	}
	for _, f := range faelle {
		u := neuBeenden(t, f.exchanges...)
		vorher, _ := u.sessions.hat(sid)
		got, err := u.beenden(context.Background(), sid)
		if err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		if e := u.p.liste(); !slices.Equal(e, []string{"Sessions.Get(r1)", "Drain(r1)", "ListBySession(r1)", "SetStatus(r1)"}) {
			t.Errorf("%s: Ereignisse %v", f.name, e)
		}
		want := vorher
		want.Status = f.soll
		if got != want || got.CreatedAt.Nanosecond() != 123456789 || got.ServiceID != dienstA || got.ID != sid {
			t.Errorf("%s: Ergebnis %+v, erwartet %+v", f.name, got, want)
		}
		if s := u.sessions.statusVon(sid); s != f.soll {
			t.Errorf("%s: Status in der Ablage %q", f.name, s)
		}
		if n, ok, st := u.sessions.versuche(sid); n != 1 || ok != 1 || !slices.Equal(st, []model.SessionStatus{f.soll}) {
			t.Errorf("%s: SetStatus %d-mal, %d gelungen, %v", f.name, n, ok, st)
		}
		if a := u.record.aufrufe(); len(a) != 1 || a[0].id != sid {
			t.Errorf("%s: Drain %+v", f.name, a)
		}
		if n := u.st.anzahl("Now") + u.st.anzahl("SessionID"); n != 0 {
			t.Errorf("%s: Now oder SessionID %d-mal", f.name, n)
		}
		if n := len(u.log.alle()); n != 0 {
			t.Errorf("%s: %d Einträge im Protokoll", f.name, n)
		}
	}
}

// TestStatusErstNachDrain: im Haltepunkt von Drain ist die Session open und weder ListBySession noch
// SetStatus gerufen; der Status folgt auf jedem Rückkehrweg von Drain nur bei nil (B6, SPEC-124
// Nr. 11 d bis f, SPEC-123 Nr. 8a). Die Fälle ordnen mit Kanälen und sind deterministisch.
func TestStatusErstNachDrain(t *testing.T) {
	imHalt := func(t *testing.T, u *welt, h *halt) {
		t.Helper()
		warte(t, h.erreicht, "Drain betreten")
		if s := u.sessions.statusVon(sid); s != model.SessionOpen {
			t.Errorf("im Halt von Drain: Status %q", s)
		}
		for _, op := range []string{"ListBySession", "SetStatus"} {
			if n := u.st.anzahl(op); n != 0 {
				t.Errorf("im Halt von Drain: %s %d-mal", op, n)
			}
		}
		if n, _, _ := u.sessions.versuche(sid); n != 0 {
			t.Errorf("im Halt von Drain: SetStatus %d-mal versucht", n)
		}
	}

	t.Run("Bericht mit nil", func(t *testing.T) {
		u := neuBeenden(t, model.CaptureComplete)
		h := u.st.halte("Drain", 1)
		defer h.gib()
		ch := u.startBeenden(context.Background(), sid)
		imHalt(t, u, h)
		h.gib()
		r := abwarten(t, ch, "Rückkehr")
		if r.err != nil || r.s.Status != model.SessionComplete {
			t.Errorf("Ergebnis %+v, %v", r.s, r.err)
		}
		if u.st.anzahl("ListBySession") != 1 || u.st.anzahl("SetStatus") != 1 || u.sessions.statusVon(sid) != model.SessionComplete {
			t.Errorf("nach Drain: Status %q", u.sessions.statusVon(sid))
		}
	})

	arten := map[string]error{
		"Fehler mit Code":     e4001(),
		"ErrUngueltig":        model.ErrUngueltig,
		"Platte":              errDisk,
		"Canceled":            context.Canceled,
		"DeadlineExceeded":    context.DeadlineExceeded,
		"Canceled und Platte": errors.Join(context.Canceled, errDisk),
	}
	for name, art := range arten {
		t.Run("Fehler "+name, func(t *testing.T) {
			u := neuBeenden(t, model.CaptureComplete)
			u.st.scheitere("Drain", 1, art)
			h := u.st.halte("Drain", 1)
			defer h.gib()
			ch := u.startBeenden(context.Background(), sid)
			imHalt(t, u, h)
			h.gib()
			r := abwarten(t, ch, "Rückkehr")
			if r.err != art || r.s != (model.Session{}) {
				t.Errorf("Ergebnis %+v, %v, erwartet den Fehler %v unverändert", r.s, r.err, art)
			}
			if s := u.sessions.statusVon(sid); s != model.SessionOpen {
				t.Errorf("nach dem Fehler: Status %q", s)
			}
			for _, op := range []string{"ListBySession", "SetStatus"} {
				if n := u.st.anzahl(op); n != 0 {
					t.Errorf("nach dem Fehler: %s %d-mal", op, n)
				}
			}
			if n, _, _ := u.sessions.versuche(sid); n != 0 {
				t.Errorf("nach dem Fehler: SetStatus %d-mal versucht", n)
			}
		})
	}

	t.Run("ctx im Halt beendet", func(t *testing.T) {
		u := neuBeenden(t, model.CaptureComplete)
		ctx, abbruch := context.WithCancel(context.Background())
		defer abbruch()
		h := u.st.halte("Drain", 1)
		defer h.gib()
		ch := u.startBeenden(ctx, sid)
		imHalt(t, u, h)
		abbruch()
		h.gib()
		r := abwarten(t, ch, "Rückkehr")
		if r.err != nil || r.s.Status != model.SessionComplete {
			t.Errorf("Ergebnis %+v, %v", r.s, r.err)
		}
		if s := u.sessions.statusVon(sid); s != model.SessionComplete {
			t.Errorf("Status %q", s)
		}
	})
}

// TestLesenNachDrain: ein Exchange, den Drain abbricht und anhängt, geht in die Statusregel ein
// (B7, SPEC-124 Nr. 11 e, 12).
func TestLesenNachDrain(t *testing.T) {
	for _, f := range []struct {
		capture model.CaptureStatus
		soll    model.SessionStatus
	}{{model.CaptureIncomplete, model.SessionIncomplete}, {model.CaptureComplete, model.SessionComplete}} {
		u := neuBeenden(t, model.CaptureComplete, model.CaptureComplete, model.CaptureComplete)
		u.record.bericht = inbound.DrainBericht{Abgebrochen: 1}
		u.st.bei("Drain", 1, func() { u.exchange.haenge(sid, f.capture) })
		s, err := u.beenden(context.Background(), sid)
		if err != nil || s.Status != f.soll || u.sessions.statusVon(sid) != f.soll {
			t.Errorf("angehängt %s: Ergebnis %+v, %v, Ablage %q", f.capture, s, err, u.sessions.statusVon(sid))
		}
	}
}

// TestStatusregel: ein Fall je Element der Menge und je Position bei Mengen bis zu drei Exchanges (B8,
// SPEC-124 Nr. 12); größere Mengen führt TestStatusregelMenge. Der Fake von SetStatus bricht den Test
// bei einem Status außer complete und incomplete ab.
func TestStatusregel(t *testing.T) {
	c, i, f := model.CaptureComplete, model.CaptureIncomplete, model.CaptureFailed
	faelle := []struct {
		name      string
		exchanges []model.CaptureStatus
		bericht   inbound.DrainBericht
		soll      model.SessionStatus
	}{
		{"keine Exchanges", nil, inbound.DrainBericht{}, model.SessionComplete},
		{"ein complete", []model.CaptureStatus{c}, inbound.DrainBericht{}, model.SessionComplete},
		{"drei complete", []model.CaptureStatus{c, c, c}, inbound.DrainBericht{}, model.SessionComplete},
		{"incomplete vorn", []model.CaptureStatus{i, c, c}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"incomplete mitten", []model.CaptureStatus{c, i, c}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"incomplete hinten", []model.CaptureStatus{c, c, i}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"failed vorn", []model.CaptureStatus{f, c, c}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"failed mitten", []model.CaptureStatus{c, f, c}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"failed hinten", []model.CaptureStatus{c, c, f}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"leerer Wert", []model.CaptureStatus{c, ""}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"unbekannter Wert", []model.CaptureStatus{c, "x"}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"einziger unbekannter Wert", []model.CaptureStatus{"x"}, inbound.DrainBericht{}, model.SessionIncomplete},
		{"NichtAngehaengt 1 mit complete", []model.CaptureStatus{c, c}, inbound.DrainBericht{NichtAngehaengt: 1}, model.SessionIncomplete},
		{"NichtAngehaengt 2 mit complete", []model.CaptureStatus{c, c}, inbound.DrainBericht{NichtAngehaengt: 2}, model.SessionIncomplete},
		{"NichtAngehaengt 1 ohne Exchange", nil, inbound.DrainBericht{NichtAngehaengt: 1}, model.SessionIncomplete},
		{"NichtAngehaengt 2 ohne Exchange", nil, inbound.DrainBericht{NichtAngehaengt: 2}, model.SessionIncomplete},
		{"Abgebrochen 3 mit complete", []model.CaptureStatus{c, c, c}, inbound.DrainBericht{Abgebrochen: 3}, model.SessionComplete},
		{"Abgebrochen 3 ohne Exchange", nil, inbound.DrainBericht{Abgebrochen: 3}, model.SessionComplete},
		{"Abgebrochen 3 mit failed", []model.CaptureStatus{c, f, c}, inbound.DrainBericht{Abgebrochen: 3}, model.SessionIncomplete},
	}
	for _, fall := range faelle {
		pruefeStatus(t, fall.name, fall.exchanges, fall.bericht, fall.soll)
	}
}

// pruefeStatus beendet die Session sid mit den Exchanges und dem Bericht von Drain und prüft Ergebnis
// und Ablage gegen den erwarteten Status.
func pruefeStatus(t *testing.T, name string, exchanges []model.CaptureStatus, bericht inbound.DrainBericht, soll model.SessionStatus) {
	t.Helper()
	u := neuBeenden(t, exchanges...)
	u.record.bericht = bericht
	s, err := u.beenden(context.Background(), sid)
	if err != nil || s.Status != soll || u.sessions.statusVon(sid) != soll {
		t.Errorf("%s: Ergebnis %+v, %v, Ablage %q, erwartet %q", name, s, err, u.sessions.statusVon(sid), soll)
	}
}

// mengenGroessen sind die Größen der Mengen von TestStatusregelMenge; 10000 ist die größte.
func mengenGroessen() []int {
	return []int{1, 2, 3, 4, 5, 8, 9, 16, 63, 64, 65, 100, 127, 128, 129, 150, 1000, 10000}
}

// positionen nennt die Stellen einer Menge der Größe n, an die TestStatusregelMenge den abweichenden
// Exchange setzt: bis 16 Elemente jede Stelle, darüber die ersten beiden, die Mitte mit ihren Nachbarn
// und die letzten beiden.
func positionen(n int) []int {
	if n <= 16 {
		alle := make([]int, n)
		for i := range alle {
			alle[i] = i
		}
		return alle
	}
	return []int{0, 1, n/2 - 1, n / 2, n/2 + 1, n - 2, n - 1}
}

// TestStatusregelMenge: die Regel gilt für jeden gelieferten Exchange, gleich wie groß die Menge ist
// und an welcher Stelle der abweichende steht (B8, SPEC-124 Nr. 12). Schleife über Größe × Stelle × Art
// des abweichenden Exchanges (incomplete, failed, leerer und unbekannter Wert); Größen 1 bis 10000
// (mengenGroessen), Stellen nach positionen. Eine Obergrenze der geprüften Exchanges bis unter 10000
// fängt der Test; eine ab 10000 nicht (akzeptiertes Negativ, Plan §3 Zeile B8). Dazu je Größe: lauter
// complete ergibt complete, NichtAngehaengt 1 incomplete, Abgebrochen 3 complete.
func TestStatusregelMenge(t *testing.T) {
	c := model.CaptureComplete
	abweichend := []model.CaptureStatus{model.CaptureIncomplete, model.CaptureFailed, "", "x"}
	for _, n := range mengenGroessen() {
		lauter := make([]model.CaptureStatus, n)
		for i := range lauter {
			lauter[i] = c
		}
		name := fmt.Sprintf("%d Exchanges", n)
		pruefeStatus(t, name+", lauter complete", lauter, inbound.DrainBericht{}, model.SessionComplete)
		pruefeStatus(t, name+", NichtAngehaengt 1", lauter, inbound.DrainBericht{NichtAngehaengt: 1}, model.SessionIncomplete)
		pruefeStatus(t, name+", Abgebrochen 3", lauter, inbound.DrainBericht{Abgebrochen: 3}, model.SessionComplete)
		for _, pos := range positionen(n) {
			for _, abw := range abweichend {
				cs := slices.Clone(lauter)
				cs[pos] = abw
				pruefeStatus(t, fmt.Sprintf("%s, %q an Stelle %d", name, abw, pos), cs, inbound.DrainBericht{}, model.SessionIncomplete)
			}
		}
	}
}

// aufrufStelle nennt eine Aufrufstelle der Abfolge von Beenden: den Namen, den Zähler der Fakes und den
// Fehler der Tabelle, den der Port dort liefern kann.
type aufrufStelle struct {
	name, op string
	tabelle  error
}

func stellenBeenden() []aufrufStelle {
	return []aufrufStelle{
		{"Get", "Sessions.Get", verpackt(outbound.ErrNichtGefunden)},
		{"Drain", "Drain", e4001()},
		{"ListBySession", "ListBySession", verpackt(outbound.ErrNichtGefunden)},
		{"SetStatus", "SetStatus", verpackt(outbound.ErrUngueltigerUebergang)},
	}
}

// tabellenzeile ist ein Fehler, den die Abbildung von SPEC-109 Nr. 5 benennt, oder einer ohne Zeile.
type tabellenzeile struct {
	name string
	err  error
}

// tabellenzeilen nennt jeden Fehler, den die Tabelle der Abbildung benennt, und die Platte als einen
// ohne Zeile; TestTabellenzeilenVollstaendig hält die Liste gegen die Tabelle.
func tabellenzeilen() []tabellenzeile {
	return []tabellenzeile{
		{"ErrBlobFehlt", outbound.ErrBlobFehlt}, {"ErrBlobBeschaedigt", outbound.ErrBlobBeschaedigt},
		{"ErrBeschaedigt", outbound.ErrBeschaedigt}, {"ErrPfad", outbound.ErrPfad}, {"ErrZuGross", outbound.ErrZuGross},
		{"ErrUpstreamNichtErreichbar", outbound.ErrUpstreamNichtErreichbar}, {"ErrUpstreamZeit", outbound.ErrUpstreamZeit},
		{"ErrFremderService", outbound.ErrFremderService}, {"ErrDienstUnbekannt", outbound.ErrDienstUnbekannt},
		{"ErrFalscheArt", outbound.ErrFalscheArt}, {"ErrReferenziert", outbound.ErrReferenziert},
		{"ErrSessionGeschlossen", outbound.ErrSessionGeschlossen}, {"ErrUngueltigerUebergang", outbound.ErrUngueltigerUebergang},
		{"ErrNichtGefunden", outbound.ErrNichtGefunden}, {"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden},
		{"ErrUngueltig", model.ErrUngueltig}, {"ErrMetadatenUngueltig", model.ErrMetadatenUngueltig},
		{"ErrDatenUngueltig", model.ErrDatenUngueltig}, {"ErrURLUngueltig", model.ErrURLUngueltig},
		{"ErrURLNichtUnterstuetzt", model.ErrURLNichtUnterstuetzt}, {"ErrUngueltigeSequenz", outbound.ErrUngueltigeSequenz},
		{"ErrAbgeschlossen", outbound.ErrAbgeschlossen}, {"ErrRevisionVeraltet", outbound.ErrRevisionVeraltet},
		{"Platte (ohne Zeile)", errDisk},
	}
}

// TestTabellenzeilenVollstaendig: die Liste der Tests nennt jede Zeile der Tabelle, sonst deckt die
// Matrix Kontext neben Tabellenfehler nicht jede.
func TestTabellenzeilenVollstaendig(t *testing.T) {
	zeilen := tabellenzeilen()
	namen := outbound.EintraegeAbbildung()
	if len(zeilen) != len(namen)+1 {
		t.Fatalf("%d Fehler in der Liste, %d in der Tabelle (+ einer ohne Zeile)", len(zeilen), len(namen))
	}
	hat := map[string]bool{}
	for _, z := range zeilen {
		hat[z.name] = true
	}
	for _, n := range namen {
		if !hat[n] {
			t.Errorf("%s fehlt in der Liste", n)
		}
	}
}

// kontextNeben nennt die Formen, in denen der Kontextfehler c neben dem Fehler tab ankommt: allein,
// verpackt, in beiden Reihenfolgen verbunden (Join und %w mit %w).
func kontextNeben(c, tab error) []tabellenzeile {
	return []tabellenzeile{
		{"allein", c},
		{"verpackt", fmt.Errorf("%w: %s", c, marker)},
		{"Join(c, tab)", errors.Join(c, tab)},
		{"Join(tab, c)", errors.Join(tab, c)},
		{"%w c, %w tab", fmt.Errorf("%w: %w", c, tab)},
		{"%w tab, %w c", fmt.Errorf("%w: %w", tab, c)},
	}
}

func istKontext(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// kontextKombinationen ist die Menge der Kontextfehler neben jeder Zeile der Tabelle: beide Kontextfehler
// in allen Formen von kontextNeben und beide zugleich neben der Zeile.
func kontextKombinationen(z tabellenzeile) []tabellenzeile {
	var out []tabellenzeile
	for _, c := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, f := range kontextNeben(c, z.err) {
			out = append(out, tabellenzeile{fmt.Sprintf("%v %s", c, f.name), f.err})
		}
	}
	return append(out, tabellenzeile{"beide Kontextfehler", errors.Join(context.Canceled, context.DeadlineExceeded, z.err)})
}

// TestKontextBeenden: ein Kontextfehler kehrt unverändert und ohne Code zurück und beendet Beenden an
// jeder Stelle, neben jeder Zeile der Tabelle von SPEC-109 Nr. 5 und in den Formen von kontextNeben
// und kontextKombinationen; die Matrix ist die Schleife über Stelle × Zeile × Form, und ein vorab
// beendeter Kontext beendet Beenden an Get (B9, SPEC-124 Nr. 2, 11).
func TestKontextBeenden(t *testing.T) {
	stellen := stellenBeenden()
	for i, st := range stellen {
		for _, z := range tabellenzeilen() {
			for _, komb := range kontextKombinationen(z) {
				art := komb.err
				name := fmt.Sprintf("%s, %s, %s", st.name, z.name, komb.name)
				u := neuBeenden(t, model.CaptureComplete)
				u.st.scheitere(st.op, 1, art)
				s, err := u.beenden(context.Background(), sid)
				if err != art || s != (model.Session{}) || !istKontext(err) {
					t.Errorf("%s: %v, Session %v, erwartet den Fehler unverändert", name, err, s)
				}
				if st.name != "Drain" && !ohneCode(err) {
					t.Errorf("%s: Fehler mit Code: %v", name, err)
				}
				for k, spaeter := range stellen {
					want := 0
					if k <= i {
						want = 1
					}
					if n := u.st.anzahl(spaeter.op); n != want {
						t.Errorf("%s: %s %d-mal, erwartet %d", name, spaeter.op, n, want)
					}
				}
				if n := len(u.log.alle()); n != 0 {
					t.Errorf("%s: %d Einträge im Protokoll", name, n)
				}
				if s := u.sessions.statusVon(sid); s != model.SessionOpen {
					t.Errorf("%s: Status %q", name, s)
				}
			}
		}
	}
	// ein vorab beendeter ctx beendet Beenden an Get.
	ende, abbruch := context.WithCancel(context.Background())
	abbruch()
	vorbei, abbruch2 := context.WithDeadline(context.Background(), time.Unix(1, 0))
	defer abbruch2()
	for name, ctx := range map[string]context.Context{"Canceled": ende, "DeadlineExceeded": vorbei} {
		u := neuBeenden(t, model.CaptureComplete)
		s, err := u.beenden(ctx, sid)
		if err != ctx.Err() || s != (model.Session{}) {
			t.Errorf("%s vorab: %v, Session %v", name, err, s)
		}
		for k, spaeter := range stellen {
			want := 0
			if k == 0 {
				want = 1
			}
			if n := u.st.anzahl(spaeter.op); n != want {
				t.Errorf("%s vorab: %s %d-mal, erwartet %d", name, spaeter.op, n, want)
			}
		}
	}
}

// TestBeendenFehler: ein Fehler eines Ports ist ein Fehler mit dem Code der Tabelle von SPEC-109 Nr. 5
// mit dem Gegenstand Session, der Fehler von Drain kehrt unverändert zurück; danach folgt kein Port,
// die Session bleibt open (B10, SPEC-124 Nr. 8, 11).
func TestBeendenFehler(t *testing.T) {
	arten := []struct {
		art  error
		code string
	}{
		{verpackt(outbound.ErrNichtGefunden), "ODR-E7003"},
		{verpackt(outbound.ErrBeschaedigt), "ODR-E3001"},
		{errDisk, "ODR-E9001"},
	}
	stellen := stellenBeenden()
	for i, st := range stellen {
		if st.name == "Drain" {
			continue
		}
		faelle := slices.Clone(arten)
		if st.name == "SetStatus" {
			faelle = append(faelle, struct {
				art  error
				code string
			}{verpackt(outbound.ErrUngueltigerUebergang), "ODR-E7011"})
		}
		for _, f := range faelle {
			name := st.name + " " + f.code
			u := neuBeenden(t, model.CaptureComplete)
			u.st.scheitere(st.op, 1, f.art)
			s, err := u.beenden(context.Background(), sid)
			code, status := codeVon(t, err)
			if code != f.code || s != (model.Session{}) || !errors.Is(err, f.art) {
				t.Errorf("%s: %s, Status %d, Session %v", name, code, status, s)
			}
			if f.code == "ODR-E7011" && status != 409 {
				t.Errorf("%s: Status %d", name, status)
			}
			for k, spaeter := range stellen {
				want := 0
				if k <= i {
					want = 1
				}
				if n := u.st.anzahl(spaeter.op); n != want {
					t.Errorf("%s: %s %d-mal, erwartet %d", name, spaeter.op, n, want)
				}
			}
			if s := u.sessions.statusVon(sid); s != model.SessionOpen {
				t.Errorf("%s: Status %q", name, s)
			}
		}
	}
	// Drain: der Fehler kehrt unverändert zurück, mit und ohne Code.
	for name, art := range map[string]error{"Fehler mit Code": e4001(), "ErrUngueltig": model.ErrUngueltig, "Platte": errDisk} {
		u := neuBeenden(t, model.CaptureComplete)
		u.st.scheitere("Drain", 1, art)
		_, err := u.beenden(context.Background(), sid)
		if err != art {
			t.Errorf("Drain %s: %v, erwartet den Fehler unverändert", name, err)
		}
		if _, mitCode := art.(*inbound.Fehler); !mitCode && !ohneCode(err) {
			t.Errorf("Drain %s: Code an einem Fehler ohne Code", name)
		}
		for _, op := range []string{"ListBySession", "SetStatus"} {
			if n := u.st.anzahl(op); n != 0 {
				t.Errorf("Drain %s: %s %d-mal", name, op, n)
			}
		}
		if s := u.sessions.statusVon(sid); s != model.SessionOpen {
			t.Errorf("Drain %s: Status %q", name, s)
		}
	}
}

// TestBeendenWiederholung: nach einem Fehler bleibt die Session open, und ein zweites Beenden gelingt;
// der Dienst merkt sich nichts (B11, SPEC-124 Nr. 11, Zustand).
func TestBeendenWiederholung(t *testing.T) {
	for _, st := range []aufrufStelle{{"ListBySession", "ListBySession", errDisk}, {"SetStatus", "SetStatus", errDisk}, {"Drain", "Drain", errDisk}} {
		u := neuBeenden(t, model.CaptureComplete)
		u.st.scheitere(st.op, 1, st.tabelle)
		if _, err := u.beenden(context.Background(), sid); err == nil {
			t.Fatalf("%s: erster Aufruf ohne Fehler", st.name)
		}
		if s := u.sessions.statusVon(sid); s != model.SessionOpen || u.st.anzahl("Drain") != 1 {
			t.Errorf("%s: Status %q, Drain %d-mal", st.name, s, u.st.anzahl("Drain"))
		}
		s, err := u.beenden(context.Background(), sid)
		if err != nil || s.Status != model.SessionComplete {
			t.Errorf("%s: zweiter Aufruf %+v, %v", st.name, s, err)
		}
		for _, op := range []string{"Drain", "ListBySession"} {
			want := 2
			if st.name == "Drain" && op == "ListBySession" {
				want = 1
			}
			if n := u.st.anzahl(op); n != want {
				t.Errorf("%s: %s %d-mal, erwartet %d", st.name, op, n, want)
			}
		}
		if u.sessions.statusVon(sid) != model.SessionComplete {
			t.Errorf("%s: Status %q", st.name, u.sessions.statusVon(sid))
		}
	}
}

// TestCtxBeenden: Get und Drain erhalten den ctx des Aufrufs unverändert; ListBySession und SetStatus
// einen Kontext mit den Werten des Aufrufs, den sein Abbruch und seine Frist nicht erreichen (B12,
// SPEC-124 Nr. 11 b, d, e, f).
func TestCtxBeenden(t *testing.T) {
	nachDrain := func(t *testing.T, u *welt) {
		t.Helper()
		for _, op := range []string{"ListBySession", "SetStatus"} {
			g := u.st.gesehen(op)
			if len(g) != 1 {
				t.Fatalf("%s: %d Aufrufe", op, len(g))
			}
			c := g[0]
			if _, hat := c.Deadline(); c.Value(ctxSchluessel{}) != "w1" || c.Err() != nil || c.Done() != nil || hat {
				t.Errorf("%s: Wert %v, Err %v, Done %v, Frist %v", op, c.Value(ctxSchluessel{}), c.Err(), c.Done(), hat)
			}
		}
	}

	t.Run("lebender ctx mit Wert und Frist", func(t *testing.T) {
		ctx, abbruch := context.WithTimeout(context.WithValue(context.Background(), ctxSchluessel{}, "w1"), time.Hour)
		defer abbruch()
		u := neuBeenden(t, model.CaptureComplete)
		if _, err := u.beenden(ctx, sid); err != nil {
			t.Fatal(err)
		}
		for _, op := range []string{"Sessions.Get", "Drain"} {
			if g := u.st.gesehen(op); len(g) != 1 || g[0] != ctx {
				t.Errorf("%s erhielt nicht den ctx des Aufrufs: %v", op, g)
			}
		}
		nachDrain(t, u)
	})

	t.Run("Canceled in Drain", func(t *testing.T) {
		ctx, abbruch := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, "w1"))
		defer abbruch()
		u := neuBeenden(t, model.CaptureComplete)
		u.st.bei("Drain", 1, abbruch)
		s, err := u.beenden(ctx, sid)
		if err != nil || s.Status != model.SessionComplete || u.sessions.statusVon(sid) != model.SessionComplete {
			t.Errorf("Ergebnis %+v, %v, Ablage %q", s, err, u.sessions.statusVon(sid))
		}
		if g := u.st.gesehen("Drain"); len(g) != 1 || g[0] != ctx || ctx.Err() == nil {
			t.Errorf("Drain: %v, ctx.Err %v", g, ctx.Err())
		}
		nachDrain(t, u)
	})

	t.Run("DeadlineExceeded in Drain", func(t *testing.T) {
		ctx := &verfallCtx{wert: "w1"}
		u := neuBeenden(t, model.CaptureComplete)
		u.st.bei("Drain", 1, ctx.verfalle)
		s, err := u.beenden(ctx, sid)
		if err != nil || s.Status != model.SessionComplete || u.sessions.statusVon(sid) != model.SessionComplete {
			t.Errorf("Ergebnis %+v, %v, Ablage %q", s, err, u.sessions.statusVon(sid))
		}
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("ctx.Err %v", ctx.Err())
		}
		nachDrain(t, u)
	})
}

// TestCtxEndetZwischen: der Dienst prüft den Kontext nicht selbst; endet der Kontext zwischen zwei
// Aufrufen (Err und Done melden ihn nach dem n-ten Aufruf als beendet, n = 1 bis 4: zwischen Get und
// Drain, in Drain, danach), ist Beenden gelungen und trägt den Status (B12, SPEC-124 Nr. 2, 11 d).
// Schleife über die Art des Kontextfehlers × n. Den Kontext wertet allein der Fake aus: Err und Done
// werden zusammen genau zweimal gerufen, von Get und von Drain; jeder weitere Aufruf von Err oder
// Done im Dienst, wo auch immer, ist rot.
func TestCtxEndetZwischen(t *testing.T) {
	for _, art := range []error{context.Canceled, context.DeadlineExceeded} {
		for n := 1; n <= 4; n++ {
			name := fmt.Sprintf("%v nach %d Aufrufen von Err", art, n)
			ctx := &endCtx{wert: "w1", nOK: n, err: art}
			u := neuBeenden(t, model.CaptureComplete)
			s, err := u.beenden(ctx, sid)
			if err != nil || s.Status != model.SessionComplete || u.sessions.statusVon(sid) != model.SessionComplete {
				t.Errorf("%s: Ergebnis %+v, %v, Ablage %q", name, s, err, u.sessions.statusVon(sid))
			}
			for _, st := range stellenBeenden() {
				if z := u.st.anzahl(st.op); z != 1 {
					t.Errorf("%s: %s %d-mal", name, st.name, z)
				}
			}
			if z := ctx.aufrufe(); z != 2 {
				t.Errorf("%s: Err und Done %d-mal gerufen, erwartet 2 (Get und Drain)", name, z)
			}
		}
	}
}

// TestFristBeenden: die frist jedes Drain ist Frist der Konfig unverändert, die id die der Session,
// genau ein Aufruf (B13, SPEC-124 Nr. 1, 11 d).
func TestFristBeenden(t *testing.T) {
	for _, frist := range []time.Duration{fristTest, 0, time.Nanosecond} {
		u := neu(t, func(k *session.Konfig) { k.Frist = frist })
		u.dienst(dienstA)
		u.offene(sid, dienstA, 1)
		if _, err := u.beenden(context.Background(), sid); err != nil {
			t.Fatal(err)
		}
		a := u.record.aufrufe()
		if len(a) != 1 || a[0].frist != frist || a[0].id != sid {
			t.Errorf("Frist %v: Drain %+v", frist, a)
		}
	}
}

// pfad ist ein Weg durch Beenden: ein Aufbau, die Kennung und der ctx.
type pfad struct {
	name  string
	setup func(*welt)
	id    string
	ctx   func() context.Context
}

// beendenPfade nennt die Wege von B4, B5, B6, B9 und B10 und einen nil-ctx.
func beendenPfade() []pfad {
	bg := func() context.Context { return context.Background() }
	scheitere := func(op string, err error) func(*welt) {
		return func(u *welt) { u.st.scheitere(op, 1, err) }
	}
	ps := []pfad{
		{"Gutfall complete", nil, sid, bg},
		{"Gutfall incomplete", func(u *welt) { u.exchange.haenge(sid, model.CaptureIncomplete) }, sid, bg},
		{"unbekannt", nil, "unbekannt", bg},
		{"nicht offen complete", nil, "c1", bg},
		{"nicht offen incomplete", nil, "i1", bg},
		{"Get Fehler", scheitere("Sessions.Get", errDisk), sid, bg},
		{"Drain Fehler mit Code", scheitere("Drain", e4001()), sid, bg},
		{"Drain Fehler ohne Code", scheitere("Drain", errDisk), sid, bg},
		{"ListBySession Fehler", scheitere("ListBySession", errDisk), sid, bg},
		{"SetStatus Fehler", scheitere("SetStatus", errDisk), sid, bg},
		{"SetStatus Übergang", scheitere("SetStatus", verpackt(outbound.ErrUngueltigerUebergang)), sid, bg},
		{"ctx vorab beendet", nil, sid, func() context.Context {
			c, abbruch := context.WithCancel(context.Background())
			abbruch()
			return c
		}},
		{"nil-ctx", nil, sid, func() context.Context { return nil }},
	}
	for _, st := range stellenBeenden() {
		ps = append(ps, pfad{"Kontextfehler in " + st.name, scheitere(st.op, context.Canceled), sid, bg})
	}
	return ps
}

// laufPfad führt einen Weg in einer frischen Umgebung aus.
func laufPfad(t *testing.T, p pfad) *welt {
	t.Helper()
	u := neuBeenden(t, model.CaptureComplete)
	u.sessions.fuege("c1", dienstA, erzeugt(2), model.SessionComplete)
	u.sessions.fuege("i1", dienstA, erzeugt(3), model.SessionIncomplete)
	if p.setup != nil {
		p.setup(u)
	}
	_, _ = u.beenden(p.ctx(), p.id)
	return u
}

// TestBeendenNurGenannte: Beenden ruft von den Fakes nur Get, Drain, ListBySession und SetStatus, je
// höchstens einmal, nichts anderes, weder Now noch SessionID (B14, SPEC-124 Nr. 11).
func TestBeendenNurGenannte(t *testing.T) {
	erlaubt := map[string]int{"Sessions.Get": 1, "Drain": 1, "ListBySession": 1, "SetStatus": 1}
	for _, p := range beendenPfade() {
		u := laufPfad(t, p)
		for _, op := range alleOps() {
			if n := u.st.anzahl(op); n > erlaubt[op] {
				t.Errorf("%s: %s %d-mal, erlaubt %d", p.name, op, n, erlaubt[op])
			}
		}
		if p.name == "Gutfall complete" {
			for op, soll := range erlaubt {
				if n := u.st.anzahl(op); n != soll {
					t.Errorf("%s: %s %d-mal, erwartet %d", p.name, op, n, soll)
				}
			}
		}
	}
}

// TestBeendenProtokollLeer: Beenden schreibt auf keinem Pfad ins Protokoll (B15, SPEC-124 Nr. 9).
func TestBeendenProtokollLeer(t *testing.T) {
	for _, p := range beendenPfade() {
		u := laufPfad(t, p)
		if e := u.log.alle(); len(e) != 0 {
			t.Errorf("%s: %d Einträge im Protokoll: %+v", p.name, len(e), e)
		}
	}
}

// kennungsformen nennt jede Form einer Kennung, die SPEC-124 Nr. 4 als außerhalb der Form bezeichnet
// (Großbuchstaben, _ . / : \, Leerzeichen, Steuerzeichen, NUL, DEL, jedes Byte ab 0x80 als gültiges und
// als ungültiges UTF-8, mehr als 64 Bytes). Die Zeichen _ . / : Tabulator, Zeilenvorschub, Wagenrücklauf,
// NUL und DEL stehen in der Mitte, am linken und am rechten Rand; . / - stehen auch doppelt. Dazu die
// Längen 63, 64, 65, 70, 71, 80, 100, 128, 255, 256 und 1000 Bytes und je Form der Nachbar, der durch
// Schreibweise, Entfernen oder Ersetzen des Zeichens, Kürzen, Trimmen oder Abschneiden entsteht. Der
// Bestand führt jede offen; die leere Kennung führt TestBeendenPruefung.
func kennungsformen() []string {
	return []string{
		// Schreibweise, Leerzeichen am Rand, Präfix
		"r1", "R1", " r1 ", " r1", "r1 ", "R1 ", "Ab c", "ab c", "r1x", "r",
		// Zeichen außerhalb von 0-9a-z- und ihre Nachbarn
		"a", "ab", "a-b", "a_b", "a.b", "a/b", "a:b", "A", "%41", "a%2fb",
		"a\tb", "a\nb", "a\x00b", "a\x7f", "a\x00",
		// Bytes ab 0x80: gültiges UTF-8 (auch zwei Schreibweisen desselben Zeichens), ungültiges
		"1", "ä1", "Ä1", "é", "é", "a\xffb", "\xff1",
		// - am Rand
		"-a", "a-", "-a-", "--",
		"/a", "a/", ".a", "a.", "_a", "a_", ":a", "a:", "\ta", "a\t", "\na", "a\n", "\ra", "a\r", "a\r\n",
		"\x00a", "\x7fa", "a\rb", "a\x7fb", "1ä", "a\xff", "0a", "a0", "a..b", "a//b", "a---b", "a\\b",
		// Länge: 63, 64 (die Grenze der Form), 65 und 70 Bytes, darüber bis 1000
		strings.Repeat("a", 63), strings.Repeat("a", 64), strings.Repeat("a", 65), strings.Repeat("a", 70),
		strings.Repeat("a", 71), strings.Repeat("a", 80), strings.Repeat("a", 100), strings.Repeat("a", 128),
		strings.Repeat("a", 255), strings.Repeat("a", 256), strings.Repeat("a", 1000),
	}
}

// bestandKennungen legt jede Kennung als offene Session des Dienstes mit einem vollständigen
// Exchange an.
func (u *welt) bestandKennungen(ids []string) {
	u.dienst(dienstA)
	for i, id := range ids {
		u.offene(id, dienstA, i+1)
		u.exchange.haenge(id, model.CaptureComplete)
	}
}

// TestKennungenBeenden: jede Aufrufstelle von Beenden (Get, Drain, ListBySession, SetStatus) erhält die
// Kennung unverändert, ob sie in der Form liegt oder nicht, und beendet allein diese Session; der
// Bestand trägt die Nachbarn jeder Kennung offen (B4, SPEC-124 Nr. 4, 11 b). Schleife über Kennungsform
// (kennungsformen: jede Form von Nr. 4 und die Grenzen der Länge) × Aufrufstelle.
func TestKennungenBeenden(t *testing.T) {
	ids := kennungsformen()
	for _, k := range ids {
		u := neu(t)
		u.bestandKennungen(ids)
		s, err := u.beenden(context.Background(), k)
		if err != nil || s.ID != k || s.Status != model.SessionComplete {
			t.Errorf("Kennung %q: Ergebnis %+v, %v", k, s, err)
		}
		for _, st := range stellenBeenden() {
			if z := u.zahl(st.op, k); z != 1 {
				t.Errorf("Kennung %q: %s mit dieser Kennung %d-mal", k, st.name, z)
			}
			if n := u.st.anzahl(st.op); n != 1 {
				t.Errorf("Kennung %q: %s %d-mal", k, st.name, n)
			}
		}
		if a := u.record.aufrufe(); len(a) != 1 || a[0].id != k {
			t.Errorf("Kennung %q: Drain %+v", k, a)
		}
		for _, andere := range ids {
			want := model.SessionOpen
			if andere == k {
				want = model.SessionComplete
			}
			if got := u.sessions.statusVon(andere); got != want {
				t.Errorf("Kennung %q: Session %q hat den Status %q, erwartet %q", k, andere, got, want)
			}
		}
	}
}
