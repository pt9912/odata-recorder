package urlparser_test

import (
	"fmt"
	"strings"
	"testing"
)

// Die Dublette in `$expand` (SPEC-262 Nr. 4 e "bytegleich"): beide Seiten der Gleichheit. Gleiche Elemente aller Formen (`*`,
// Namen von 1 bis 128 Zeichen, Pfade mit 2 bis 40 Segmenten, mit `$ref`, qualifiziert) im Abstand 0, 1, 5 und 20; verschiedene
// Elemente gleicher Länge mit gleichem Präfix, gleichem Ende oder gleicher Mitte. Orakel ist `exOrakel`.

// exDublettenZahl ist die Zahl der Eingaben von TestExpandDublette: 4 * (1 + 7 + 5 * 3) gleiche, 6 * 3 * 2 verschiedene Namen und
// 3 * 3 verschiedene Pfade.
const exDublettenZahl = 137

func exPfadMitSegmenten(s int) string { return strings.TrimSuffix(strings.Repeat("a/", s), "/") }

func TestExpandDublette(t *testing.T) {
	p := exParser()
	n := 0
	formen := []string{"*"}
	for _, l := range []int{1, 2, 16, 17, 64, 65, 128} {
		formen = append(formen, strings.Repeat("a", l))
	}
	for _, s := range []int{2, 3, 4, 10, 40} {
		formen = append(formen, exPfadMitSegmenten(s), exPfadMitSegmenten(s)+"/$ref", "Ns.T/"+exPfadMitSegmenten(s))
	}
	for _, f := range formen {
		for _, abstand := range []int{0, 1, 5, 20} {
			el := []string{f}
			for i := range abstand {
				el = append(el, fmt.Sprintf("q%d", i))
			}
			el = append(el, f)
			e := strings.Join(el, ",")
			exPruefe(t, p, exQ(e), exOrakel(e))
			n++
		}
	}
	for _, l := range []int{2, 3, 17, 40, 65, 128} {
		basis := strings.Repeat("a", l)
		for _, v := range []string{"b" + basis[1:], basis[:l-1] + "b", basis[:l/2] + "b" + basis[l/2+1:]} {
			for _, e := range []string{basis + "," + v, basis + ",x," + v} {
				exPruefe(t, p, exQ(e), exOrakel(e))
				n++
			}
		}
	}
	for _, s := range []int{17, 40, 100} {
		basis := exPfadMitSegmenten(s)
		for _, v := range []string{basis + "/b", "b/" + basis, basis[:len(basis)-1] + "b"} {
			e := basis + "," + v
			exPruefe(t, p, exQ(e), exOrakel(e))
			n++
		}
	}
	exZaehlt(t, "TestExpandDublette", n, exDublettenZahl)
}
