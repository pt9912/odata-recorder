package inspection_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// ctxBeendet liefert einen Kontext, der mit context.Canceled beendet ist.
func ctxBeendet() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// kontextfehler sind die fünf Formen eines Kontextfehlers der Menge M-Fehler.
func kontextfehler() []struct {
	name string
	err  error
} {
	return []struct {
		name string
		err  error
	}{
		{"Canceled", context.Canceled},
		{"DeadlineExceeded", context.DeadlineExceeded},
		{"verpackt", fmt.Errorf("lesen: %w", context.Canceled)},
		{"Join vorn", errors.Join(outbound.ErrNichtGefunden, context.Canceled)},
		{"Join hinten", errors.Join(context.DeadlineExceeded, outbound.ErrNichtGefunden)},
	}
}

// tabelleSnapshot ist SPEC-109 Nr. 5 für den Gegenstand Snapshot: die Tabelle für Session, abgeschrieben
// in port_test.go, mit den drei Zeilen, die der Gegenstand ändert (ErrNichtGefunden ist ODR-E7002;
// ErrBereitsVorhanden und ErrUngueltig nennen keine Zeile für Snapshot und fallen auf ODR-E9001).
func tabelleSnapshot() []struct {
	name string
	err  error
	code model.Code
} {
	t := tabelleSession()
	for i := range t {
		switch t[i].name {
		case "ErrNichtGefunden":
			t[i].code = model.CodeE7002
		case "ErrBereitsVorhanden", "ErrUngueltig":
			t[i].code = model.CodeE9001
		}
	}
	return t
}

// erwarteterStatus ist der Status des Codes nach SPEC-236 Nr. 8 für die drei erreichbaren Codes.
func erwarteterStatus(c model.Code) (int, bool) {
	switch c {
	case model.CodeE7002:
		return 404, true
	case model.CodeE3001, model.CodeE9001:
		return 500, true
	}
	return 0, false
}

// TestNewPortNil (K1, K2): ein nil in jedem Feld von Konfig, dessen Typ eine Schnittstelle ist, weist
// New ab; Coverage und Zaehler gehören dazu.
func TestNewPortNil(t *testing.T) {
	var gepruft []string
	for _, f := range reflect.VisibleFields(reflect.TypeFor[inspection.Konfig]()) {
		if f.Type.Kind() != reflect.Interface {
			continue
		}
		gepruft = append(gepruft, f.Name)
		u := neueUmgebung()
		reflect.ValueOf(&u.k).Elem().FieldByName(f.Name).SetZero()
		d, err := inspection.New(u.k)
		if d != nil {
			t.Errorf("%s nil: New liefert einen Dienst", f.Name)
		}
		ohneCode(t, f.Name+" nil", err)
	}
	for _, name := range []string{"Exchanges", "Coverage", "Zaehler"} {
		if !slices.Contains(gepruft, name) {
			t.Errorf("das Feld %s ist kein Port von Konfig", name)
		}
	}
}

// TestNewModus (K3): genau die vier Modi des Vertrags sind gültig; jede Abwandlung ist ErrUngueltig.
func TestNewModus(t *testing.T) {
	gueltig := []string{"record", "replay", "mock", "hybrid"}
	var ungueltig []string
	for _, m := range gueltig {
		ungueltig = append(ungueltig, strings.ToUpper(m), strings.ToUpper(m[:1])+m[1:], m+" ", " "+m, m+"x", "x"+m, m+"\n")
	}
	ungueltig = append(ungueltig, "", "recorder", "live", "record,replay")
	for _, m := range gueltig {
		u := neueUmgebung()
		u.k.Modus = m
		if d, err := inspection.New(u.k); d == nil || err != nil {
			t.Errorf("Modus %q: %v, %v", m, d, err)
		}
	}
	for _, m := range ungueltig {
		u := neueUmgebung()
		u.k.Modus = m
		d, err := inspection.New(u.k)
		if d != nil {
			t.Errorf("Modus %q: New liefert einen Dienst", m)
		}
		ohneCode(t, fmt.Sprintf("Modus %q", m), err)
		if strings.Contains(err.Error(), m) && m != "" {
			t.Errorf("Modus %q: der Text des Fehlers trägt den Wert: %v", m, err)
		}
		if slices.Contains(gueltig, m) {
			t.Errorf("Modus %q ist gültig und steht in der Liste der ungültigen", m)
		}
	}
}

// TestNewEinstellungenKopie (K4): nil ist gültig; New kopiert die Abbildung, eine spätere Änderung
// (Wert, neuer Schlüssel, gelöschter Schlüssel) wirkt nicht auf die Diagnose.
func TestNewEinstellungenKopie(t *testing.T) {
	u := neueUmgebung()
	u.k.Einstellungen = nil
	if d, err := inspection.New(u.k); d == nil || err != nil {
		t.Errorf("nil: %v, %v", d, err)
	}

	m := map[string]string{"LINK_REWRITE": "off", "MATCH_BODY_MODE": "raw"}
	u.k.Einstellungen = m
	d := u.dienst(t)
	m["LINK_REWRITE"] = "geaendert"
	m["REPLAY_STRATEGY"] = "neu"
	delete(m, "MATCH_BODY_MODE")
	diag, err := d.Diagnose(context.Background())
	want := map[string]string{"LINK_REWRITE": "off", "MATCH_BODY_MODE": "raw"}
	if err != nil || !reflect.DeepEqual(diag.Einstellungen, want) {
		t.Errorf("nach der Änderung: %v, Fehler %v, erwartet %v", diag.Einstellungen, err, want)
	}
}

// TestKontextNilCoverageDiagnose (A2): ctx nil ist bei Coverage (auch mit leerer Kennung) und Diagnose
// ErrUngueltig ohne Code und ohne Port-Aufruf.
func TestKontextNilCoverageDiagnose(t *testing.T) {
	faelle := []struct {
		name string
		lauf func(d *inspection.Inspektion) (any, error)
		leer any
	}{
		{"Coverage", func(d *inspection.Inspektion) (any, error) { return d.Coverage(nilCtx(), "b1") }, []inbound.CoverageEintrag(nil)},
		{"Coverage, Kennung leer", func(d *inspection.Inspektion) (any, error) { return d.Coverage(nilCtx(), "") }, []inbound.CoverageEintrag(nil)},
		{"Diagnose", func(d *inspection.Inspektion) (any, error) { return d.Diagnose(nilCtx()) }, inbound.Diagnose{}},
	}
	for _, c := range faelle {
		u := neueUmgebung()
		u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
		erg, err := c.lauf(u.dienst(t))
		ohneCode(t, c.name, err)
		if !reflect.DeepEqual(erg, c.leer) {
			t.Errorf("%s: Ergebnis %+v statt Nullwert", c.name, erg)
		}
		if u.cov.gesamt() != 0 || u.rec.gesamt() != 0 || len(u.tel.Aufrufe()) != 0 {
			t.Errorf("%s: ein Port wurde gerufen (Coverage %d, Recording %d, Telemetrie %d)", c.name, u.cov.gesamt(), u.rec.gesamt(), len(u.tel.Aufrufe()))
		}
	}
}

// TestCoverageKennung (C1, Menge M-Coverage Achse Kennung × ctx-Zustand): die leere Kennung ist
// ErrUngueltig ohne Code, auch bei beendetem ctx, und ruft keinen Port; jede andere geht unverändert
// und genau einmal an List.
func TestCoverageKennung(t *testing.T) {
	kennungen := []string{
		"", "b1", "B1", " ", " b1", "b1 ", "w1", strings.Repeat("a", 64), "bl-" + string([]byte{0xc3, 0xa4}),
	}
	zustaende := []string{"lebend", "beendet"}
	faelle := 0
	for _, id := range kennungen {
		for _, zs := range zustaende {
			faelle++
			u := neueUmgebung()
			u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete", EvidenceJSON: "{}"}})
			ctx := context.Background()
			if zs == "beendet" {
				ctx = ctxBeendet()
			}
			erg, err := u.dienst(t).Coverage(ctx, id)
			was := fmt.Sprintf("Kennung %q, ctx %s", id, zs)
			if id == "" {
				ohneCode(t, was, err)
				if errors.Is(err, context.Canceled) {
					t.Errorf("%s: der Kontextfehler geht der Argumentprüfung vor", was)
				}
				if erg != nil || u.cov.gesamt() != 0 {
					t.Errorf("%s: Ergebnis %v, Aufrufe %d", was, erg, u.cov.gesamt())
				}
				continue
			}
			if !slices.Equal(u.cov.kennungen(), []string{id}) || u.cov.anzahl("List") != 1 || u.cov.gesamt() != 1 {
				t.Errorf("%s: Kennungen %q, Aufrufe %d", was, u.cov.kennungen(), u.cov.gesamt())
			}
			if zs == "lebend" && (err != nil || len(erg) != 1) {
				t.Errorf("%s: Fehler %v, %d Einträge", was, err, len(erg))
			}
			if zs == "beendet" && err != context.Canceled {
				t.Errorf("%s: Fehler %v", was, err)
			}
		}
	}
	if want := len(kennungen) * len(zustaende); faelle != want {
		t.Errorf("%d Fälle gelaufen, die Achsen verlangen %d", faelle, want)
	}
}

// Die Literale des Vertrags (SPEC-236 Nr. 6): das Orakel liest die Status nie aus dem Modell.
const (
	m101  = 101
	m1001 = 1001
)

func statusL() []string { return []string{"complete", "partial", "unknown", "fremder Status"} }

// evidenzen sind drei Formen der Evidenz; die dritte trägt Leerraum und ein Zeichen jenseits von
// ASCII aus Bytewerten.
func evidenzen() []string {
	return []string{`{"zeilen":3,"grund":"x"}`, "", " \t" + string([]byte{0xc3, 0xa4}) + " Ende\n"}
}

// TestCoverageLiterale (C2): die Literale des Orakels sind die Konstanten des Modells; der fremde
// Status bleibt für das Modell fremd.
func TestCoverageLiterale(t *testing.T) {
	konstanten := []model.CoverageStatus{model.CoverageComplete, model.CoveragePartial, model.CoverageUnknown}
	for i, k := range konstanten {
		if string(k) != statusL()[i] {
			t.Errorf("Konstante %q, Literal %q", k, statusL()[i])
		}
		if err := (model.Coverage{EntitySet: "A", Status: model.CoverageStatus(statusL()[i]), EvidenceJSON: "{}"}).Validate(); err != nil {
			t.Errorf("Status %q: %v", statusL()[i], err)
		}
	}
	fremd := model.Coverage{EntitySet: "A", Status: model.CoverageStatus(statusL()[len(konstanten)]), EvidenceJSON: "{}"}
	if len(statusL()) != len(konstanten)+1 || fremd.Validate() == nil {
		t.Errorf("der fremde Status %q ist dem Modell nicht fremd", fremd.Status)
	}
	if m101 <= inspection.StandardLimit || m1001 <= inspection.MaxLimit {
		t.Errorf("m %d und %d liegen nicht über StandardLimit %d und MaxLimit %d", m101, m1001, inspection.StandardLimit, inspection.MaxLimit)
	}
}

// eintrag ist der i-te Eintrag der Fixture mit m Einträgen: absteigende Namen (nie aufsteigend), Status
// und Evidenz drehen sich je Eintrag.
func eintrag(m, i, s, e int) inbound.CoverageEintrag {
	ev := evidenzen()
	return inbound.CoverageEintrag{
		EntitySet: fmt.Sprintf("Set%04d", m-1-i),
		Status:    model.CoverageStatus(statusL()[(s+i)%len(statusL())]),
		Evidenz:   ev[(e+i)%len(ev)],
	}
}

// TestCoverageEintraege (C2, Menge M-Coverage: Status × Evidenz × m): das Ergebnis hat einen Eintrag je
// Eintrag des Ports in dessen Reihenfolge, Feld für Feld gleich; eine Baseline ohne Einträge liefert
// eine leere Liste, nie nil; keine Seitenbildung.
func TestCoverageEintraege(t *testing.T) {
	ms := []int{0, 1, 2, m101, m1001}
	faelle := 0
	for s := range statusL() {
		for e := range evidenzen() {
			for _, m := range ms {
				faelle++
				soll := make([]inbound.CoverageEintrag, m)
				fixture := make([]model.Coverage, m)
				for i := range m {
					soll[i] = eintrag(m, i, s, e)
					fixture[i] = model.Coverage{EntitySet: soll[i].EntitySet, Status: soll[i].Status, EvidenceJSON: soll[i].Evidenz}
				}
				if m >= 2 && slices.IsSorted(sets(soll)) {
					t.Fatalf("m %d: die Fixture liegt aufsteigend", m)
				}
				varianten := [][]model.Coverage{fixture}
				if m == 0 {
					varianten = [][]model.Coverage{nil, {}}
				}
				for _, fix := range varianten {
					u := neueUmgebung()
					u.cov.setzeListe(fix)
					erg, err := u.dienst(t).Coverage(context.Background(), "b1")
					was := fmt.Sprintf("Status %d Evidenz %d m %d", s, e, m)
					if err != nil || erg == nil || !slices.Equal(erg, soll) {
						t.Errorf("%s: Fehler %v, Ergebnis nil: %v, %d Einträge, erwartet %d", was, err, erg == nil, len(erg), m)
					}
					if u.cov.anzahl("List") != 1 || u.cov.gesamt() != 1 {
						t.Errorf("%s: Aufrufe %d", was, u.cov.gesamt())
					}
				}
			}
		}
	}
	if want := len(statusL()) * len(evidenzen()) * len(ms); faelle != want {
		t.Errorf("%d Fälle gelaufen, die Achsen verlangen %d", faelle, want)
	}
}

func sets(es []inbound.CoverageEintrag) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.EntitySet
	}
	return out
}

// TestCoverageOhneZwischenspeicher (C2): zwei Aufrufe nacheinander mit geänderter Fixture dazwischen:
// das zweite Ergebnis ist neu, List wurde zweimal gerufen.
func TestCoverageOhneZwischenspeicher(t *testing.T) {
	u := neueUmgebung()
	d := u.dienst(t)
	u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete", EvidenceJSON: "1"}})
	erst, err := d.Coverage(context.Background(), "b1")
	u.cov.setzeListe([]model.Coverage{{EntitySet: "B", Status: "partial", EvidenceJSON: "2"}, {EntitySet: "A", Status: "unknown", EvidenceJSON: "3"}})
	zweit, err2 := d.Coverage(context.Background(), "b1")
	if err != nil || err2 != nil || len(erst) != 1 || erst[0].EntitySet != "A" || len(zweit) != 2 || zweit[0].EntitySet != "B" || zweit[1].Evidenz != "3" || u.cov.anzahl("List") != 2 {
		t.Errorf("erst %v (%v), zweit %v (%v), List %d", erst, err, zweit, err2, u.cov.anzahl("List"))
	}
}

// TestCoverageFehlerMitCode (C3, Menge M-Fehler Aufrufstelle List): jeder Fehler der Tabelle wird mit
// dem Gegenstand Snapshot abgebildet: Code, Status, Ursache erreichbar. Die Tabelle des Tests hat genau
// die Namen von outbound.EintraegeAbbildung() (und den anderen Fehler).
func TestCoverageFehlerMitCode(t *testing.T) {
	var namen []string
	for _, z := range tabelleSnapshot() {
		namen = append(namen, z.name)
	}
	quelle := append(outbound.EintraegeAbbildung(), "anderer Fehler")
	slices.Sort(namen)
	slices.Sort(quelle)
	if !slices.Equal(namen, quelle) {
		t.Fatalf("Namen der Tabelle des Tests %v, der Quelle %v", namen, quelle)
	}
	for _, z := range tabelleSnapshot() {
		u := neueUmgebung()
		u.cov.fehler = fmt.Errorf("lesen: %w", z.err)
		erg, err := u.dienst(t).Coverage(context.Background(), "b1")
		var fe *inbound.Fehler
		if !errors.As(err, &fe) {
			t.Errorf("%s: kein *Fehler: %v", z.name, err)
			continue
		}
		if fe.Code != z.code || !errors.Is(err, z.err) || !errors.Is(fe.Ursache, z.err) || erg != nil {
			t.Errorf("%s: Code %v Ursache %v Ergebnis %v, erwartet %v", z.name, fe.Code, fe.Ursache, erg, z.code)
		}
		status, bekannt := erwarteterStatus(z.code)
		if bekannt && fe.Status != status || !z.code.StatusErlaubt(fe.Status) {
			t.Errorf("%s: Status %d, erwartet %d (bekannt: %v)", z.name, fe.Status, status, bekannt)
		}
	}
}

// TestCoverageKontext (C4, Menge M-Fehler Achse ctx-Zustand und Form): der ctx geht unverändert an
// den Port; ein Kontextfehler des Ports bleibt unverändert (gleicher Wert) und ohne Code, auch neben
// einem Fehler der Tabelle; ein Abbruch nach der Rückkehr ändert das Ergebnis nicht.
func TestCoverageKontext(t *testing.T) {
	u := neueUmgebung()
	u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, "mein ctx")
	if _, err := u.dienst(t).Coverage(ctx, "b1"); err != nil || len(u.cov.kontexte()) != 1 || u.cov.kontexte()[0] != ctx {
		t.Errorf("lebender ctx: Fehler %v, der Port erhielt %v", err, u.cov.kontexte())
	}

	v := neueUmgebung()
	v.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
	var fe *inbound.Fehler
	erg, err := v.dienst(t).Coverage(ctxBeendet(), "b1")
	if err != context.Canceled || errors.As(err, &fe) || v.cov.anzahl("List") != 1 || v.cov.anzahl("Get") != 0 || erg != nil {
		t.Errorf("vor dem Aufruf abgebrochen: %v, List %d, Get %d, Ergebnis %v", err, v.cov.anzahl("List"), v.cov.anzahl("Get"), erg)
	}

	w := neueUmgebung()
	w.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}, {EntitySet: "B", Status: "partial"}})
	spaeter, cancel := context.WithCancel(context.Background())
	w.cov.nachLesen = cancel
	erg, err = w.dienst(t).Coverage(spaeter, "b1")
	if err != nil || len(erg) != 2 || spaeter.Err() == nil {
		t.Errorf("nach der Rückkehr beendet: Fehler %v, %d Einträge, beendet: %v", err, len(erg), spaeter.Err() != nil)
	}

	for _, k := range kontextfehler() {
		x := neueUmgebung()
		x.cov.fehler = k.err
		_, err := x.dienst(t).Coverage(context.Background(), "b1")
		if err != k.err || errors.As(err, &fe) || x.cov.anzahl("List") != 1 {
			t.Errorf("%s: Fehler %v (gleich: %v)", k.name, err, err == k.err)
		}
	}
}

// TestCoverageSchreibtNichts (C5): Coverage ruft von CoverageRepository nur List, höchstens einmal,
// bei Erfolg, Fehler, Parameterfehler, ctx nil und Kontextfehler; Get nie.
func TestCoverageSchreibtNichts(t *testing.T) {
	faelle := []struct {
		name   string
		ctx    func() context.Context
		id     string
		fehler error
		liste  int
	}{
		{"Erfolg", context.Background, "b1", nil, 1},
		{"Fehler des Ports", context.Background, "b1", errors.New("platte voll"), 1},
		{"Parameterfehler", context.Background, "", nil, 0},
		{"ctx nil", nilCtx, "b1", nil, 0},
		{"Kontextfehler", ctxBeendet, "b1", nil, 1},
	}
	for _, c := range faelle {
		u := neueUmgebung()
		u.cov.setzeListe([]model.Coverage{{EntitySet: "A", Status: "complete"}})
		u.cov.fehler = c.fehler
		_, _ = u.dienst(t).Coverage(c.ctx(), c.id)
		if u.cov.anzahl("Get") != 0 || u.cov.anzahl("List") != c.liste || u.cov.gesamt() != c.liste {
			t.Errorf("%s: Get %d List %d gesamt %d", c.name, u.cov.anzahl("Get"), u.cov.anzahl("List"), u.cov.gesamt())
		}
		if u.rec.gesamt() != 0 || len(u.tel.Aufrufe()) != 0 {
			t.Errorf("%s: ein anderer Port wurde gerufen", c.name)
		}
	}
}

// TestNullwertImFehlerfall (C6): im Fehlerfall ist das Ergebnis der Nullwert, bei Coverage nil (nicht
// leer), auch wenn der Port Daten neben dem Fehler liefert; bei Diagnose Diagnose{} (Einstellungen und
// Fehler nil).
func TestNullwertImFehlerfall(t *testing.T) {
	daten := []model.Coverage{{EntitySet: "A", Status: "complete", EvidenceJSON: "{}"}}
	type fall struct {
		name   string
		ctx    func() context.Context
		id     string
		fehler error
	}
	faelle := []fall{{"ctx nil", nilCtx, "b1", nil}, {"Kennung leer", context.Background, "", nil}}
	for _, k := range kontextfehler() {
		faelle = append(faelle, fall{"Kontextfehler " + k.name, context.Background, "b1", k.err})
	}
	for _, z := range tabelleSnapshot() {
		faelle = append(faelle, fall{"Fehler " + z.name, context.Background, "b1", fmt.Errorf("lesen: %w", z.err)})
	}
	for _, c := range faelle {
		u := neueUmgebung()
		u.cov.setzeListe(daten)
		u.cov.fehler = c.fehler
		erg, err := u.dienst(t).Coverage(c.ctx(), c.id)
		if err == nil || erg != nil {
			t.Errorf("%s: Fehler %v, Ergebnis nil: %v", c.name, err, erg == nil)
		}
	}

	u := neueUmgebung()
	u.k.Modus = "mock"
	u.k.Einstellungen = map[string]string{"LINK_REWRITE": "off"}
	d := u.dienst(t)
	abgelaufen, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel()
	for name, ctx := range map[string]context.Context{"ctx nil": nilCtx(), "Canceled": ctxBeendet(), "DeadlineExceeded": abgelaufen} {
		erg, err := d.Diagnose(ctx)
		if err == nil || !reflect.DeepEqual(erg, inbound.Diagnose{}) || erg.Einstellungen != nil || erg.Fehler != nil {
			t.Errorf("Diagnose, %s: Fehler %v, Ergebnis %+v", name, err, erg)
		}
	}
}
