package supporttest_test

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// eintragVoll trägt in jedem der sieben Felder einen eigenen Wert, auch solche, die der Adapter abwiese.
func eintragVoll() support.Eintrag {
	return support.Eintrag{
		Art: "keine-art", ServiceID: "dienst \xc3\xbc", SessionID: "s", WorkspaceID: "w!", ImportID: "i",
		ErrorCode: model.Code(-3), KorrelationsID: "k",
	}
}

// TestFakeAuditHaeltJedesFeld prüft SPEC-234 Nr. 9: jeder Eintrag steht mit allen sieben Feldern
// unverändert da, je Feld ein Fall mit nur diesem Feld und einer mit allen.
func TestFakeAuditHaeltJedesFeld(t *testing.T) {
	voll := eintragVoll()
	typ := reflect.TypeOf(voll)
	if typ.NumField() != 7 {
		t.Fatalf("Eintrag hat %d Felder, erwartet 7", typ.NumField())
	}
	faelle := []support.Eintrag{voll}
	for i := range typ.NumField() {
		nur := reflect.New(typ).Elem()
		nur.Field(i).Set(reflect.ValueOf(voll).Field(i))
		faelle = append(faelle, nur.Interface().(support.Eintrag))
	}
	if len(faelle) != 8 {
		t.Fatalf("%d Fälle, erwartet 8", len(faelle))
	}
	f := supporttest.NewFakeAudit()
	for i, e := range faelle {
		if err := f.Append(e); err != nil {
			t.Fatalf("Fall %d: Fehler %v", i, err)
		}
	}
	got := f.Eintraege()
	if len(got) != len(faelle) {
		t.Fatalf("%d Einträge, erwartet %d", len(got), len(faelle))
	}
	for i := range faelle {
		if got[i] != faelle[i] {
			t.Errorf("Fall %d: Eintrag %+v, erwartet %+v", i, got[i], faelle[i])
		}
	}
}

// TestFakeAuditZaehltJedeOperation prüft SPEC-234 Nr. 9: Aufrufe ist die Summe der Aufrufe von
// Append und Eintraege, die Länge von Eintraege die Zahl der Aufrufe von Append, und Eintraege
// liefert eine Kopie.
func TestFakeAuditZaehltJedeOperation(t *testing.T) {
	f := supporttest.NewFakeAudit()
	if f.Aufrufe() != 0 || len(f.Eintraege()) != 0 {
		t.Fatalf("Nullstand: Aufrufe %d", f.Aufrufe())
	}
	f = &supporttest.FakeAudit{}
	for range 3 {
		_ = f.Append(eintragVoll())
	}
	_ = f.Eintraege()
	kopie := f.Eintraege()
	if len(kopie) != 3 {
		t.Fatalf("Länge %d, erwartet 3", len(kopie))
	}
	if got := f.Aufrufe(); got != 5 {
		t.Errorf("Aufrufe %d, erwartet 5 (3 Append, 2 Eintraege)", got)
	}
	kopie[0].ServiceID = "ge\xc3\xa4ndert"
	if f.Eintraege()[0].ServiceID != eintragVoll().ServiceID {
		t.Error("eine Änderung am gelieferten Slice änderte den Stand des Fakes")
	}
	if got := f.Aufrufe(); got != 6 {
		t.Errorf("Aufrufe %d, erwartet 6", got)
	}
}

// TestFakeAuditFehlerUndHaltepunkt prüft SPEC-234 Nr. 9: Fehler ist das Ergebnis von Append und der
// Eintrag steht trotzdem; Vor läuft nach dem Zählen und vor dem Festhalten, mit dem Eintrag und ohne
// die Sperre.
func TestFakeAuditFehlerUndHaltepunkt(t *testing.T) {
	boese := errors.New("Ablage")
	for _, fehler := range []error{nil, boese} {
		f := &supporttest.FakeAudit{Fehler: fehler}
		if err := f.Append(eintragVoll()); err != fehler {
			t.Errorf("Fehler %v, erwartet %v", err, fehler)
		}
		if n := len(f.Eintraege()); n != 1 {
			t.Errorf("Fehler %v: %d Einträge, erwartet 1", fehler, n)
		}
	}
	f := &supporttest.FakeAudit{}
	type lage struct {
		aufrufe  int
		eintrage int
		arg      support.Eintrag
	}
	gesehen := make(chan lage, 1)
	f.Vor = func(e support.Eintrag) {
		a := f.Aufrufe()
		n := len(f.Eintraege())
		gesehen <- lage{a, n, e}
	}
	fertig := make(chan struct{})
	go func() {
		_ = f.Append(eintragVoll())
		close(fertig)
	}()
	select {
	case <-fertig:
	case <-time.After(5 * time.Second):
		t.Fatal("Append kehrte nicht zurück: Vor läuft unter der Sperre des Fakes")
	}
	l := <-gesehen
	if l.aufrufe != 1 {
		t.Errorf("in Vor: Aufrufe %d, erwartet 1 (Append ist gezählt)", l.aufrufe)
	}
	if l.eintrage != 0 {
		t.Errorf("in Vor: %d Einträge, erwartet 0 (der Eintrag fehlt)", l.eintrage)
	}
	if l.arg != eintragVoll() {
		t.Errorf("in Vor: Argument %+v", l.arg)
	}
	if n := len(f.Eintraege()); n != 1 {
		t.Errorf("nach Append: %d Einträge, erwartet 1", n)
	}
}

// TestFakeAuditGleichzeitig prüft SPEC-234 Nr. 9: 64 Goroutinen × 8 Aufrufe verlieren keinen
// Eintrag und keinen Zähler.
func TestFakeAuditGleichzeitig(t *testing.T) {
	const g, n, leser, runden = 64, 8, 8, 50
	// Zweige: ohne und mit Haltepunkt; neben den Aufrufen von Append rufen leser Goroutinen Eintraege
	// und Aufrufe, und Vor ruft beide aus der Goroutine von Append.
	for _, mitVor := range []bool{false, true} {
		f := supporttest.NewFakeAudit()
		var vor atomic.Int64
		if mitVor {
			f.Vor = func(support.Eintrag) {
				vor.Add(1)
				_ = f.Aufrufe()
				_ = f.Eintraege()
			}
		}
		var wg sync.WaitGroup
		for range g {
			wg.Go(func() {
				for range n {
					_ = f.Append(eintragVoll())
				}
			})
		}
		for range leser {
			wg.Go(func() {
				for range runden {
					_ = f.Eintraege()
					_ = f.Aufrufe()
				}
			})
		}
		wg.Wait()
		soll := g*n + leser*runden
		if mitVor {
			soll += g * n
			if got := vor.Load(); got != g*n {
				t.Errorf("mit Vor: %d Aufrufe von Vor, erwartet %d", got, g*n)
			}
		}
		if got := f.Aufrufe(); got != soll {
			t.Errorf("mit Vor %v: Aufrufe %d, erwartet %d", mitVor, got, soll)
		}
		if got := len(f.Eintraege()); got != g*n {
			t.Errorf("mit Vor %v: %d Einträge, erwartet %d", mitVor, got, g*n)
		}
	}
}

// TestFakeAuditFehlerMalZaehlerUndVor prüft das Paar Fehler × Zähler und Fehler × Vor (SPEC-234 Nr. 9):
// mit und ohne Fehler zählt Append, ruft Vor genau einmal und hält den Eintrag fest.
func TestFakeAuditFehlerMalZaehlerUndVor(t *testing.T) {
	n := 0
	for _, fehler := range []error{nil, errors.New("Ablage")} {
		for _, mitVor := range []bool{false, true} {
			n++
			f := &supporttest.FakeAudit{Fehler: fehler}
			gerufen := 0
			if mitVor {
				f.Vor = func(support.Eintrag) { gerufen++ }
			}
			for range 3 {
				_ = f.Append(eintragVoll())
			}
			if got := f.Aufrufe(); got != 3 {
				t.Errorf("Fehler %v, Vor %v: Aufrufe %d, erwartet 3", fehler, mitVor, got)
			}
			if mitVor && gerufen != 3 || !mitVor && gerufen != 0 {
				t.Errorf("Fehler %v, Vor %v: Vor %d-mal gerufen", fehler, mitVor, gerufen)
			}
			if got := len(f.Eintraege()); got != 3 {
				t.Errorf("Fehler %v, Vor %v: %d Einträge, erwartet 3", fehler, mitVor, got)
			}
		}
	}
	if n != 4 {
		t.Errorf("Paare %d, erwartet 4", n)
	}
}
