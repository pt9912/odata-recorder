package urlparser_test

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
)

// TestParserOhneZustand (Zeile 14, SPEC-160 Nr. 7): der Parser hat genau ein Feld, den Leser `filter` vom Kind Interface. Jedes
// weitere Feld wäre Zustand der Instanz, den ein Aufruf nach der Rückkehr des FilterLeser schreiben und am Ende lesen könnte;
// dieses Fenster hat keinen Haltepunkt, den ein Test erzwingen kann, die Gestalt des Typs ist deshalb der Fänger. Zustand in einer
// Variable des Pakets fängt `make lint` (`gochecknoglobals`); Zustand in einem Objekt hinter dem Feld fängt
// TestGleichzeitigMitFilter, mit der Grenze „ein Kern“; der Rest ist das akzeptierte Negativ (3) des Plans.
func TestParserOhneZustand(t *testing.T) {
	typ := reflect.TypeFor[urlparser.Parser]()
	if typ.NumField() != 1 {
		names := make([]string, typ.NumField())
		for i := range names {
			names[i] = typ.Field(i).Name
		}
		t.Fatalf("Parser hat %d Felder %v, erwartet genau eines (`filter`); SPEC-160 Nr. 7: kein veränderlicher Zustand", typ.NumField(), names)
	}
	if f := typ.Field(0); f.Name != "filter" || f.Type.Kind() != reflect.Interface {
		t.Errorf("Feld %s vom Kind %v, erwartet `filter` vom Kind Interface; SPEC-160 Nr. 7: kein veränderlicher Zustand", f.Name, f.Type.Kind())
	}
}

// sofortLeser kehrt ohne Wartestelle zurück und trägt den Text, den er bekam, als Ausdruck.
type sofortLeser struct{}

func (sofortLeser) LeseFilter(text string) (interface{ IstFilterAusdruck() }, error) {
	return testAusdruck{text}, nil
}

// TestGleichzeitigMitFilter (Zeile 14, SPEC-160 Nr. 7, Wirkungs-Fänger): 16 Goroutinen rufen je 300-mal eine Instanz mit einem
// FilterLeser auf, der sofort zurückkehrt; jeder Aufruf trägt einen eigenen Text, und das Ergebnis muss genau diesen Text als
// `Filter.Text` und als Ausdruck tragen. Das Fenster nach der Rückkehr des Lesers hat keinen Haltepunkt, die Wirkung (ein fremder
// Text im Ergebnis) wird deshalb gezählt; sie tritt nur bei mehreren gleichzeitig laufenden Threads ein, auf einem Kern bleibt der
// Fänger leer (Grenze, Plan §3 Negativ (3)). Menge 4800 Aufrufe.
func TestGleichzeitigMitFilter(t *testing.T) {
	p := urlparser.New(urlparser.WithFilter(sofortLeser{}))
	var wg sync.WaitGroup
	start := make(chan struct{})
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for n := range 300 {
				text := fmt.Sprintf("g%d_%d", g, n)
				u, err := p.ParseURL("/S", "$filter="+text)
				if err != nil || u.Optionen.Filter == nil || u.Optionen.Filter.Text != text || !gleich(u.Optionen.Filter.Ausdruck, testAusdruck{text}) {
					t.Errorf("Goroutine %d, Aufruf %d: %+v %v, erwartet den Text %q", g, n, u.Optionen.Filter, err, text)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}
