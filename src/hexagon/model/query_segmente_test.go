package model_test

import (
	"math/rand/v2"
	"runtime"
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

// TestQuerySegmenteLinear (Zeile 16, SPEC-160 Nr. 6 c, SPEC-162 Nr. 1): die Zuteilung der Zerlegung ist linear in der Länge, auch für
// Texte mit dichten Segmenten. Schleife über sechs Formen (Einheit, die bis zur Länge n wiederholt wird) × n in {2048, 8192}, Menge 12.
// Orakel von Hand: die Zahl der Einheiten ist die Zahl der Segmente (die Form `&` hat keines und ergibt nil), Name und Wert des
// Segments stehen in der Tabelle. Schranke 32 · n + 4096 Bytes (`runtime.MemStats.TotalAlloc`, Differenz vor und nach dem Aufruf).
func TestQuerySegmenteLinear(t *testing.T) {
	formen := []struct {
		einheit    string
		name, wert string
		segmente   int // je Einheit
	}{
		{"a&", "a", "", 1}, {"a=&", "a", "", 1}, {"=a&", "", "a", 1}, {"a=b&", "a", "b", 1}, {"&&a", "a", "", 1}, {"&", "", "", 0},
	}
	zahl := 0
	for _, f := range formen {
		for _, n := range []int{2048, 8192} {
			einheiten := n / len(f.einheit)
			raw := strings.Repeat(f.einheit, einheiten)
			var a, b runtime.MemStats
			runtime.ReadMemStats(&a)
			got := model.QuerySegmente(raw)
			runtime.ReadMemStats(&b)
			runtime.KeepAlive(got)
			if z, grenze := b.TotalAlloc-a.TotalAlloc, uint64(32*len(raw)+4096); z > grenze {
				t.Errorf("%q × %d: Zuteilung %d über %d", f.einheit, einheiten, z, grenze)
			}
			if f.segmente == 0 {
				if got != nil {
					t.Errorf("%q × %d: %d Segmente, erwartet nil", f.einheit, einheiten, len(got))
				}
			} else if len(got) != einheiten {
				t.Errorf("%q × %d: %d Segmente, erwartet %d", f.einheit, einheiten, len(got), einheiten)
			}
			for i, s := range got {
				if s.Name != f.name || s.Wert != f.wert {
					t.Errorf("%q × %d: Segment %d ist %+v", f.einheit, einheiten, i, s)
					break
				}
			}
			zahl++
		}
	}
	if zahl != 12 {
		t.Fatalf("%d Elemente, erwartet 12", zahl)
	}
}
