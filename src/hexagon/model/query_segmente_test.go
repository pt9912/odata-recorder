package model_test

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestQuerySegmente (Zeile 43): die Zerlegung der Query ohne Dekodierung, 12 Texte.
func TestQuerySegmente(t *testing.T) {
	s := func(n, w string) model.QuerySegment { return model.QuerySegment{Name: n, Wert: w} }
	faelle := []struct {
		raw  string
		soll []model.QuerySegment
	}{
		{"a=1&b=2", []model.QuerySegment{s("a", "1"), s("b", "2")}},
		{"a=1&&b=2", []model.QuerySegment{s("a", "1"), s("b", "2")}},
		{"&a=1", []model.QuerySegment{s("a", "1")}},
		{"a=1&", []model.QuerySegment{s("a", "1")}},
		{"", nil},
		{"&", nil},
		{"a=b=c", []model.QuerySegment{s("a", "b=c")}},
		{"=v", []model.QuerySegment{s("", "v")}},
		{"a", []model.QuerySegment{s("a", "")}},
		{"a=", []model.QuerySegment{s("a", "")}},
		{"a;b=1", []model.QuerySegment{s("a;b", "1")}},
		{"?a=1", []model.QuerySegment{s("?a", "1")}},
	}
	if len(faelle) != 12 {
		t.Fatalf("%d Texte", len(faelle))
	}
	for _, f := range faelle {
		if got := model.QuerySegmente(f.raw); !slices.Equal(got, f.soll) || (f.soll == nil) != (got == nil) {
			t.Errorf("%q: %+v, erwartet %+v", f.raw, got, f.soll)
		}
	}
}

// TestQuerySegmenteGenerator (Zeilen 43 bis 45, Klasse Zerlegung): 2000 Texte über dem Alphabet `a`, `=`, `&`, `;`, `%`, `?`, `$`, `+`;
// das Orakel sind strings.FieldsFunc und strings.Cut.
func TestQuerySegmenteGenerator(t *testing.T) {
	alphabet := []byte("a=&;%?$+")
	for saat := uint64(1); saat <= 2000; saat++ {
		r := rand.New(rand.NewPCG(saat, 0x6f64))
		b := make([]byte, r.IntN(16))
		for i := range b {
			b[i] = alphabet[r.IntN(len(alphabet))]
		}
		raw := string(b)
		var soll []model.QuerySegment
		for _, seg := range strings.FieldsFunc(raw, func(c rune) bool { return c == '&' }) {
			n, w, _ := strings.Cut(seg, "=")
			soll = append(soll, model.QuerySegment{Name: n, Wert: w})
		}
		if got := model.QuerySegmente(raw); !slices.Equal(got, soll) || (soll == nil) != (got == nil) {
			t.Fatalf("Saat %d: %q: %+v, erwartet %+v", saat, raw, got, soll)
		}
	}
}
