package baselineimport_test

import (
	"math/rand/v2"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// nachsichtig ist das Orakel der Dekodierung des Imports: net/url nach Ersetzen von `+` durch `%2B`; ein Name mit ungültiger
// Kodierung bleibt unverändert (SPEC-162 Nr. 1).
func nachsichtig(s string) string {
	t, err := url.QueryUnescape(strings.ReplaceAll(s, "+", "%2B"))
	if err != nil {
		return s
	}
	return t
}

// TestImportZerlegtWieModell (Zeile 45): der Import zerlegt mit model.QuerySegmente und bleibt nachsichtig: 2000 Texte über dem
// Alphabet `a`, `=`, `&`, `;`, `%`, `?`, `$`, `+`; die Folge der Optionen ist die der Funktion, der Name dekodiert, der Wert roh.
func TestImportZerlegtWieModell(t *testing.T) {
	alphabet := []byte("a=&;%?$+")
	for saat := uint64(1); saat <= 2000; saat++ {
		r := rand.New(rand.NewPCG(saat, 0x6f64))
		b := make([]byte, r.IntN(16))
		for i := range b {
			b[i] = alphabet[r.IntN(len(alphabet))]
		}
		raw := string(b)
		var soll [][2]string
		for _, s := range model.QuerySegmente(raw) {
			soll = append(soll, [2]string{nachsichtig(s.Name), s.Wert})
		}
		if got := baselineimport.OptionenPaare(raw); !slices.Equal(got, soll) {
			t.Fatalf("Saat %d: %q: %v, erwartet %v", saat, raw, got, soll)
		}
	}
	// der Import scheitert an keiner Query: ein Name mit ungültiger Kodierung bleibt unverändert
	if got := baselineimport.OptionenPaare("%G1=1&a%41=2&&;=3"); !slices.Equal(got, [][2]string{{"%G1", "1"}, {"aA", "2"}, {";", "3"}}) {
		t.Errorf("%v", got)
	}
}
