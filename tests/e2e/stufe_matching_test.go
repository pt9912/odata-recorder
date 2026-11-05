package e2e_test

import (
	"testing"
)

// TestE2EMatching ist die Stufe von LH-FA-05 (SPEC-037 Nr. 7, SPEC-116).
func TestE2EMatching(t *testing.T) {
	binaryPfad(t)
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	aufgenommen := map[string]antwort{}
	for _, pfad := range []string{"/odata/Products?$top=1&$skip=0", "/odata/Products?$search=%c3%a4%c3%b6",
		"/odata/Products?$filter=Id%20eq%201", "/odata/Products?$filter=Id%20eq%202", "/odata/Products?q=%zz"} {
		aufgenommen[pfad] = l.frage(anfrage{pfad: pfad})
	}
	l.beendeAufnahme()
	ref.stoppe()
	l.importiere()
	rp := l.spieleAb()
	t.Run("Abweichungen", func(t *testing.T) {
		for _, f := range []struct{ name, aufnahme, replay string }{
			{"Reihenfolge der Query-Parameter", "/odata/Products?$top=1&$skip=0", "/odata/Products?$skip=0&$top=1"},
			{"Schreibweise der Hexziffern", "/odata/Products?$search=%c3%a4%c3%b6", "/odata/Products?$search=%C3%A4%C3%B6"},
		} {
			pruefeGleich(t, f.name, aufgenommen[f.aufnahme], l.replay(rp, anfrage{pfad: f.replay}, "m1"))
		}
	})
	t.Run("Filter", func(t *testing.T) {
		zwei, eins := "/odata/Products?$filter=Id%20eq%202", "/odata/Products?$filter=Id%20eq%201"
		w2, w1 := l.replay(rp, anfrage{pfad: zwei}, "m2"), l.replay(rp, anfrage{pfad: eins}, "m2")
		pruefeGleich(t, "Id eq 2", aufgenommen[zwei], w2)
		pruefeGleich(t, "Id eq 1", aufgenommen[eins], w1)
		if string(w1.body) == string(w2.body) {
			t.Errorf("zwei Filterwerte, eine Antwort: %q", w1.body)
		}
	})
	rp.beende()
	t.Run("ungültige Kodierung", func(t *testing.T) {
		for _, modus := range []string{"reject", "literal"} {
			t.Run(modus, func(t *testing.T) {
				p := l.spieleAb("MATCH_INVALID_ENCODING=" + modus)
				defer p.beende()
				wort := l.replay(p, anfrage{pfad: "/odata/Products?q=%zz"}, "m3")
				anders := l.replay(p, anfrage{pfad: "/odata/Products?q=%7a%7a"}, "m3")
				if modus == "reject" {
					pruefeFehler(t, "reject, %zz", wort, 400, "ODR-E5006")
					return
				}
				pruefeGleich(t, "literal, %zz", aufgenommen["/odata/Products?q=%zz"], wort)
				pruefeFehler(t, "literal, %7a%7a", anders, 404, "ODR-E5001")
			})
		}
	})
}
