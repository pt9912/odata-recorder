package urlparser_test

import (
	"strings"
	"testing"
)

// Die Stelle eines Elements im Wert von `$expand` (SPEC-262 Nr. 3 d "gleich wo er steht", Nr. 6 bis 32 Elemente): jede Art an
// jeder Stelle von Listen mit 2, 3, 7, 12, 25 und 32 Elementen, dazu je Paar von Stellen (i, j) vier Paare von Verstößen.
// Orakel ist `exOrakel`.

// exStellenZahl ist die Zahl der Eingaben von TestExpandStelleImWert: 6 Listenlängen mit je k Stellen und 6 Arten, dazu je
// Listenlänge k * (k - 1) Stellenpaare und 4 Paare von Verstößen.
const exStellenZahl = 7582

func TestExpandStelleImWert(t *testing.T) {
	p := exParser()
	arten := []string{"1a", "*", "a/b", "x/$ref", "Ns.T/a", ""}
	paare := [][2]string{{"*", "1a"}, {"a/b", "1a"}, {"1a", "*"}, {"*", "a/b"}}
	n := 0
	for _, k := range []int{2, 3, 7, 12, 25, 32} {
		for pos := range k {
			for _, a := range arten {
				el := exNamen("n", k)
				el[pos] = a
				e := strings.Join(el, ",")
				exPruefe(t, p, exQ(e), exOrakel(e))
				n++
			}
		}
		for i := range k {
			for j := range k {
				if i == j {
					continue
				}
				for _, ab := range paare {
					el := exNamen("n", k)
					el[i], el[j] = ab[0], ab[1]
					e := strings.Join(el, ",")
					exPruefe(t, p, exQ(e), exOrakel(e))
					n++
				}
			}
		}
	}
	exZaehlt(t, "TestExpandStelleImWert", n, exStellenZahl)
}
