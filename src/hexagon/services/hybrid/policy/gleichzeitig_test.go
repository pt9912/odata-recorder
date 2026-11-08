package policy_test

import (
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// TestGleichzeitig (SPEC-261 Nr. 9 c): vier Goroutinen rufen Entscheide auf derselben Policy mit
// 112 gemischten Anfragen (7 Methoden, 4 Codes, 2 Ziele, 2 Pfade), je 50 Runden, und bekommen die
// Entscheidungen eines einzelnen Aufrufers. Den Beleg, dass Entscheide nichts schreibt, liefert
// der Detektor von make test-race; ohne ihn bliebe ein Schreibzugriff ohne Wirkung auf das
// Ergebnis unbemerkt.
func TestGleichzeitig(t *testing.T) {
	p := neu(t, voll(kategorien0))
	var anfragen []policy.Anfrage
	for _, m := range bekannte() {
		for _, c := range []model.Code{e6001, e6002, e6003, e5001} {
			for _, z := range []string{zielPassend, "https://x"} {
				for _, pfad := range []string{pfadRoute, "/svc/Other"} {
					anfragen = append(anfragen, policy.Anfrage{Methode: m, Pfad: pfad, Ziel: z, Code: c})
				}
			}
		}
	}
	soll := make([]policy.Entscheidung, len(anfragen))
	frei := 0
	for i, a := range anfragen {
		soll[i] = p.Entscheide(a)
		frei += b2i(soll[i] == policy.Upstream)
	}
	if len(anfragen) != 112 || frei == 0 || frei == len(anfragen) {
		t.Fatalf("%d Anfragen, %d frei", len(anfragen), frei)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for r := 0; r < 50; r++ {
				for i, a := range anfragen {
					if got := p.Entscheide(a); got != soll[i] {
						t.Errorf("Anfrage %d: %s, erwartet %s", i, name(got), name(soll[i]))
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}
