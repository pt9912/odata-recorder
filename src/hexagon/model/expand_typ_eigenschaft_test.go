package model_test

import (
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// exvAlleFaelle sind die Eingaben der Zeilen 1 bis 9 und 12 der Tabelle: jeder Generator über beide Folgen der Listen, die
// Typachsen und die drei Verstöße der Zeile 12 ohne Zusammenhang zu einem Typ des Fixture-Modells.
func exvAlleFaelle() []exvFall {
	m := exvBeide()
	var out []exvFall
	for _, g := range []func([]exvT) []exvFall{exvFaelleGueltig, exvFaelleSchreibweise, exvFaelleProperty, exvFaelleOffen, exvFaelleUnbekannt,
		exvFaelleSubtyp, exvFaelleRang, exvFaelleProgrammfehler, exvFaelleProgrammfehlerVor, exvFaelleTypachsen} {
		out = append(out, g(m)...)
	}
	for _, z := range exvTabelle() {
		typ, o := exvAusFunktion(z.ursache)
		out = append(out, exvFall{"Fehler " + string(z.ursache), typ, o, string(z.ursache)})
	}
	return out
}

func exvKopieTyp(t model.EntityType) model.EntityType {
	t.Keys = slices.Clone(t.Keys)
	t.Properties = slices.Clone(t.Properties)
	t.NavigationProperties = slices.Clone(t.NavigationProperties)
	return t
}

func exvKopieO(o *model.ExpandOption) *model.ExpandOption {
	if o == nil {
		return nil
	}
	return &model.ExpandOption{Elemente: slices.Clone(o.Elemente)}
}

// TestPruefeExpandOhneAenderung (Zeile 14): die Funktion ändert weder den Typ noch den Knoten, gültig und Verstoß, in allen
// Fällen der Zeilen 1 bis 9; jeder Aufruf bekommt frische Kopien, und in der Folge der Spezifikation steht keine Liste des
// Fixture-Modells aufsteigend sortiert (TestPruefeExpandFixture), eine Sortierung ändert dort also jede Liste ab zwei Elementen.
func TestPruefeExpandOhneAenderung(t *testing.T) {
	faelle := exvAlleFaelle()
	for _, f := range faelle {
		typ, o := exvKopieTyp(f.typ), exvKopieO(f.o)
		err := model.PruefeExpand(typ, o)
		if got := exvKlasse(err); got != f.soll {
			t.Errorf("%s: %s, erwartet %s", f.art, got, f.soll)
		}
		if !reflect.DeepEqual(typ, f.typ) || !reflect.DeepEqual(o, f.o) {
			t.Fatalf("%s: Typ oder Knoten verändert", f.art)
		}
	}
}

// TestPruefeExpandWiederholung (Zeile 15): gleiche Eingabe ergibt gleiches Ergebnis, dreimal in Folge, in allen Fällen der
// Zeilen 1 bis 9 und der Listen bis 70 Elemente; die Grenzen der Längen fährt TestPruefeExpandLaenge.
func TestPruefeExpandWiederholung(t *testing.T) {
	for _, f := range slices.Concat(exvAlleFaelle(), exvFaelleLaengen(70)) {
		erste := model.PruefeExpand(f.typ, f.o)
		if got := exvKlasse(erste); got != f.soll {
			t.Errorf("%s: %s, erwartet %s", f.art, got, f.soll)
		}
		for i := 2; i <= 3; i++ {
			if weitere := model.PruefeExpand(f.typ, f.o); !reflect.DeepEqual(erste, weitere) {
				t.Errorf("%s: Aufruf %d: %v, erster Aufruf %v", f.art, i, weitere, erste)
			}
		}
	}
}

// TestPruefeExpandGleichzeitig (Zeile 16): acht Goroutinen fahren je 50 Runden alle Eingaben der Zeilen 1 bis 9 und 12 auf
// gemeinsamem Typ und Knoten, jede ab einer anderen Stelle der Folge, gleichzeitig gestartet; das Ergebnis jedes Aufrufs ist
// das des Einzelaufrufs, verglichen nach dem Ende aller Goroutinen. Der Fänger zählt die Wirkung (ein fremdes Ergebnis) und
// wartet nicht auf einen Datenrennen-Befund.
func TestPruefeExpandGleichzeitig(t *testing.T) {
	const goroutinen, runden = 8, 50
	faelle := exvAlleFaelle()
	n := len(faelle)
	t.Logf("%d Eingaben, %d Aufrufe", n, goroutinen*runden*n)
	einzeln := make([]string, n)
	for i, f := range faelle {
		einzeln[i] = exvKlasse(model.PruefeExpand(f.typ, f.o))
		if einzeln[i] != f.soll {
			t.Fatalf("%s: %s, erwartet %s", f.art, einzeln[i], f.soll)
		}
	}
	ergebnisse := make([][]error, goroutinen)
	start := make(chan struct{})
	var bereit, fertig sync.WaitGroup
	for g := range goroutinen {
		ergebnisse[g] = make([]error, runden*n)
		bereit.Add(1)
		fertig.Add(1)
		go func() {
			defer fertig.Done()
			bereit.Done()
			<-start
			k := 0
			for range runden {
				for i := range n {
					f := faelle[(i+g*n/goroutinen)%n]
					ergebnisse[g][k] = model.PruefeExpand(f.typ, f.o)
					k++
				}
			}
		}()
	}
	bereit.Wait()
	close(start)
	fertig.Wait()
	falsch := 0
	for g, liste := range ergebnisse {
		for k, err := range liste {
			j := (k%n + g*n/goroutinen) % n
			if got := exvKlasse(err); got != einzeln[j] {
				if falsch++; falsch <= 5 {
					t.Errorf("Goroutine %d, Aufruf %d (%s): %s, Einzelaufruf %s", g, k, faelle[j].art, got, einzeln[j])
				}
			}
		}
	}
	if falsch > 5 {
		t.Errorf("insgesamt %d von %d Aufrufen mit fremdem Ergebnis", falsch, goroutinen*runden*n)
	}
}
