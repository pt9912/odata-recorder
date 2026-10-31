package urlparser_test

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func queryEntkodiert(s string) string {
	t, err := url.QueryUnescape(strings.ReplaceAll(s, "+", "%2B"))
	if err != nil {
		return s
	}
	return t
}

// TestParserZerlegtWieModell (Zeile 44): der Parser hat keine eigene Zerlegung. Differentialtest über 2000 erzeugte Queries: die
// Hälfte Zeichen des Alphabets `a`, `=`, `&`, `;`, `%`, `?`, `$`, `+`, die Hälfte Wörter daraus. Wird die Query gelesen, sind die Namen
// und Werte der Optionen die der Funktion des Modells, dekodiert; ein Fehler nennt eine Option, die es dort gibt.
func TestParserZerlegtWieModell(t *testing.T) {
	zeichen := []string{"a", "=", "&", ";", "%", "?", "$", "+"}
	woerter := []string{"x", "y", "a", "=", "1", "&", "&&", ";", "?", "%41", "+", "@p"}
	gelesenAnzahl := 0
	for saat := uint64(1); saat <= 2000; saat++ {
		r := rng(saat)
		quelle := zeichen
		if saat%2 == 0 {
			quelle = woerter
		}
		var b strings.Builder
		for range r.IntN(10) {
			b.WriteString(quelle[r.IntN(len(quelle))])
		}
		raw := b.String()
		segs := model.QuerySegmente(raw)
		u, err := parse("/S", raw)
		if err != nil {
			var uf *model.URLFehler
			if !errors.As(err, &uf) || (uf.Ort != "query" && !strings.HasPrefix(uf.Ort, "query/")) {
				t.Fatalf("Saat %d: %q: %v", saat, raw, err)
			}
			if nr, ok := strings.CutPrefix(uf.Ort, "query/"); ok {
				if n, e := strconv.Atoi(nr); e != nil || n < 1 || n > len(segs) {
					t.Fatalf("Saat %d: %q: Ort %s, die Funktion des Modells zerlegt in %d Segmente", saat, raw, uf.Ort, len(segs))
				}
			}
			continue
		}
		gelesenAnzahl++
		if len(u.Optionen.Sonstige) != len(segs) {
			t.Fatalf("Saat %d: %q: %d Optionen, die Funktion des Modells zerlegt in %d", saat, raw, len(u.Optionen.Sonstige), len(segs))
		}
		for i, s := range u.Optionen.Sonstige {
			if s.Name != queryEntkodiert(segs[i].Name) || s.Wert != queryEntkodiert(segs[i].Wert) {
				t.Fatalf("Saat %d: %q: %+v gegen %+v", saat, raw, s, segs[i])
			}
		}
	}
	if gelesenAnzahl < 400 {
		t.Fatalf("nur %d von 2000 Queries gelesen", gelesenAnzahl)
	}
}
