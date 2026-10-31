package workspace_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

type ctxSchluessel struct{}

// paareMitCtx sind die Paare aus Operation und Aufruf, die ein ctx erhalten (SPEC-129 Nr. 15).
func paareMitCtx() map[string][]string {
	return map[string][]string{
		"Create": {"Snapshots.Get", "Blobs.Open", lblTx, "tx Workspaces.Create", "tx Entities.ReplaceAll"},
		"Get":    {"Workspaces.Get"},
		"List":   {"Dienste.List", "Workspaces.List"},
		"Delete": {"Workspaces.Delete", "Wartung.CleanupBlobs"},
		"Reset": {"Workspaces.Get", "Snapshots.Get", "Blobs.Open", lblTx, "tx Workspaces.Get", "tx Entities.ReplaceAll",
			"tx Cursors.ResetAll", "tx Workspaces.Advance"},
	}
}

// TestKontextUnveraendertAnJedemPort (Z06): jeder Aufruf eines Ports erhält dasselbe ctx des Aufrufs,
// das einen Wert trägt; die Menge sind die Paare aus Operation und Aufrufstelle mit ctx.
func TestKontextUnveraendertAnJedemPort(t *testing.T) {
	paare := 0
	for _, o := range operationen() {
		w := neu(t)
		o.vor(w)
		w.p.leeren()
		ctx := neuerKontext(context.WithValue(context.Background(), ctxSchluessel{}, "wert"))
		if err := o.lauf(w, ctx); err != nil {
			t.Fatalf("%s: %v", o.name, err)
		}
		var got []string
		for _, e := range w.p.ev {
			if e.ctx == nil {
				continue
			}
			got = append(got, e.op)
			if e.ctx != context.Context(ctx) {
				t.Errorf("%s: %s erhielt ein anderes ctx", o.name, e.op)
			}
		}
		want := slices.Clone(paareMitCtx()[o.name])
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("%s: Aufrufe mit ctx %v, erwartet %v", o.name, got, want)
		}
		paare += len(got)
	}
	if paare != 18 {
		t.Errorf("Paare mit ctx %d, erwartet 18", paare)
	}
}

// TestKontextfehlerDesPorts (Z07): ein Kontextfehler eines Ports, in vier Formen, beendet die Operation an
// der Stelle ohne Code und unverändert; danach wird kein Port mehr gerufen (nur Close des Stroms).
func TestKontextfehlerDesPorts(t *testing.T) {
	faelle := 0
	for _, s := range stellen() {
		for name, inj := range kontextFehler() {
			w, o := s.vorbereitet(t)
			vorher := w.st.stand()
			w.p.setzeFehler(s.label, inj)
			w.p.leeren()
			err := o.lauf(w, context.Background())
			if err != inj {
				t.Errorf("%v %s: %v, erwartet den Fehler des Ports unverändert", s, name, err)
			}
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("%v %s: kein Kontextfehler", s, name)
			}
			if _, ok := codeVon(err); ok {
				t.Errorf("%v %s: Code", s, name)
			}
			ops := slices.DeleteFunc(w.p.ops(), func(op string) bool { return op == "Strom.Close" })
			if last := ops[len(ops)-1]; last != s.label {
				t.Errorf("%v %s: letzter Aufruf %s, erwartet %s", s, name, last, s.label)
			}
			if !zustaendeGleich(vorher, w.st.stand()) {
				t.Errorf("%v %s: Zustand geändert", s, name)
			}
			faelle++
		}
	}
	if faelle != 19*4 {
		t.Errorf("Fälle %d", faelle)
	}
}

// TestFehlerAbbildungJedeStelle (Z09): jeder Fehler der Tabelle (und einer ohne Zeile) an jeder Stelle
// ist ein *Fehler mit dem Code, den outbound.Abbilden für den Gegenstand der Stelle liefert, dem einen
// Status des Codes und der Ursache.
func TestFehlerAbbildungJedeStelle(t *testing.T) {
	faelle := 0
	for _, s := range stellen() {
		for _, fe := range tabellenFehlerListe() {
			w, o := s.vorbereitet(t)
			w.p.setzeFehler(s.label, fe.err)
			err := o.lauf(w, context.Background())
			c, ok := outbound.Abbilden(fe.err, s.gegenstandVon(fe.err))
			if !ok {
				t.Fatalf("%s: Abbilden liefert keinen Code", fe.name)
			}
			pruefeCode(t, err, c, fe.err)
			faelle++
		}
	}
	if faelle != 19*22 {
		t.Errorf("Fälle %d", faelle)
	}
}

// TestAnlageAbbildungAnCreate (Z10): Workspaces.Create bildet die fünf Fehler auf ihre Codes ab; ReplaceAll
// läuft dann nicht.
func TestAnlageAbbildungAnCreate(t *testing.T) {
	for fe, c := range map[error]model.Code{
		outbound.ErrBereitsVorhanden: model.CodeE7007, outbound.ErrNichtGefunden: model.CodeE7002,
		outbound.ErrDienstUnbekannt: model.CodeE5005, outbound.ErrFremderService: model.CodeE5004,
		outbound.ErrFalscheArt: model.CodeE7008,
	} {
		w := neu(t)
		w.p.setzeFehler("tx Workspaces.Create", fe)
		_, err := w.v.Create(context.Background(), "bsl-1", "")
		pruefeCode(t, err, c, fe)
		if w.p.anzahl("tx Entities.ReplaceAll") != 0 {
			t.Errorf("%v: ReplaceAll gerufen", fe)
		}
	}
}

// TestRevisionVeraltetIstProgrammfehler (Z11): ErrRevisionVeraltet aus Advance ist ODR-E9001, und die
// Transaktion rollt zurück.
func TestRevisionVeraltetIstProgrammfehler(t *testing.T) {
	w, o := stelle{op: "Reset"}.vorbereitet(t)
	vorher := w.st.stand()
	w.p.setzeFehler("tx Workspaces.Advance", outbound.ErrRevisionVeraltet)
	pruefeCode(t, o.lauf(w, context.Background()), model.CodeE9001, outbound.ErrRevisionVeraltet)
	if !zustaendeGleich(vorher, w.st.stand()) {
		t.Error("Zustand geändert")
	}
}

// TestPanikWirdWeitergereicht (Z12): eine Panik an jeder Stelle (in einem Port und in f) kommt beim
// Aufrufer an; der Zustand ist zurückgerollt.
func TestPanikWirdWeitergereicht(t *testing.T) {
	for _, s := range stellen() {
		w, o := s.vorbereitet(t)
		vorher := w.st.stand()
		wert := "Panik an " + s.label
		w.p.setzePanik(s.label, wert)
		var r any
		func() {
			defer func() { r = recover() }()
			_ = o.lauf(w, context.Background())
		}()
		if r != wert {
			t.Errorf("%v: recover %v, erwartet %q", s, r, wert)
		}
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%v: Zustand geändert", s)
		}
	}
}
