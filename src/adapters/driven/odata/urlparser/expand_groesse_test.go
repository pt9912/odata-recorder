package urlparser_test

import (
	"fmt"
	"strings"
	"testing"
)

// Die Größe eines Elements und die Länge des Werts von `$expand` (SPEC-262 Nr. 3 b, 3 c, 6: die Länge des Werts ist allein durch
// `rawQuery` begrenzt). TestExpandGroesseAnStelle führt ein großes Element (Pfad, Teile, langer Name, Gemisch mit 45, 100 und 200
// Gliedern) an jeder Stelle von Listen mit 2 bis 32 Elementen, mit und ohne Verstoß am Nachbarn; TestExpandLaengeVorKlammerUndZahl
// führt 31 bis 34 Elemente von 1 bis 128 Zeichen mit `(` am Ende und Pfade von 100 bis 4000 Segmenten vor `(`. Orakel ist
// `exOrakel`. Offen bleibt: jede Gliederzahl zwischen den genannten Stufen (siehe Plan, Prüfgrenze).

const (
	exGroesseZahl = 1968
	exLaengeZahl  = 88
	exGroesseMax  = 8000
)

func exQueryPasst(roh string) bool { return len(exQ(roh)) <= exGroesseMax }

// exGrossesElement bildet ein Element der Art art mit s Gliedern: 0 Pfad, 1 Teile mit Pfad, 2 langer Name, 3 Teile mit langem Namen.
func exGrossesElement(art, s int) string {
	teile := strings.TrimSuffix(strings.Repeat("a.", s), ".")
	switch art {
	case 0:
		return strings.TrimSuffix(strings.Repeat("a/", s), "/")
	case 1:
		return teile + "/b"
	case 2:
		return strings.Repeat("a", 128)
	default:
		return teile + "/" + strings.Repeat("b", 128)
	}
}

func TestExpandGroesseAnStelle(t *testing.T) {
	p := exParser()
	n := 0
	for _, k := range []int{2, 4, 8, 16, 20, 32} {
		for pos := range k {
			for art := range 4 {
				for _, s := range []int{45, 100, 200} {
					el := exNamen("n", k)
					el[pos] = exGrossesElement(art, s)
					e := strings.Join(el, ",")
					if !exQueryPasst(e) {
						continue
					}
					exPruefe(t, p, exQ(e), exOrakel(e))
					n++
					el[(pos+1)%k] = "1a"
					e = strings.Join(el, ",")
					exPruefe(t, p, exQ(e), exOrakel(e))
					n++
				}
			}
		}
	}
	exZaehlt(t, "TestExpandGroesseAnStelle", n, exGroesseZahl)
}

func TestExpandLaengeVorKlammerUndZahl(t *testing.T) {
	p := exParser()
	n := 0
	for _, l := range []int{1, 10, 40, 64, 100, 128} {
		for _, zahl := range []int{31, 32, 33, 34} {
			if l == 1 && zahl > 26 {
				continue
			}
			el := make([]string, zahl)
			for i := range el {
				s := fmt.Sprintf("%d", i)
				el[i] = strings.Repeat("a", l-min(l, len(s))) + s
				if l == 1 {
					el[i] = fmt.Sprintf("%c", 'a'+i%26)
				}
			}
			e := strings.Join(el, ",")
			if !exQueryPasst(e) {
				continue
			}
			for _, w := range []string{e, e + "(", e + ",x("} {
				exPruefe(t, p, exQ(w), exOrakel(w))
				n++
			}
		}
	}
	for _, s := range []int{100, 600, 1000, 1500, 2500, 3500, 4000} {
		e := strings.TrimSuffix(strings.Repeat("a/", s), "/")
		for _, ende := range []string{"(", ",x(", "/b(", "(a"} {
			exPruefe(t, p, exQ(e+ende), exOrakel(e+ende))
			n++
		}
	}
	exZaehlt(t, "TestExpandLaengeVorKlammerUndZahl", n, exLaengeZahl)
}
