package urlparser_test

import "testing"

// Die Folge der Segmentarten (SPEC-262 Nr. 4 c: jedes Segment einzeln ein Name, ein qualifizierter Name, `$ref`, `$count` oder
// `$value`): alle Folgen der Länge 2 und 3 aus den Arten von `exSegmentArten`. Orakel ist `exOrakel`.

// exFolgeZahl ist die Zahl der Eingaben von TestExpandSegmentFolge: 11 Arten, je 11 * 11 Paare und 11 * 11 * 11 Tripel.
const exFolgeZahl = 11*11 + 11*11*11

func TestExpandSegmentFolge(t *testing.T) {
	p := exParser()
	arten := exSegmentArten()
	exZaehlt(t, "exSegmentArten", len(arten), 11)
	n := 0
	for _, a := range arten {
		for _, b := range arten {
			e := a + "/" + b
			exPruefe(t, p, exQ(e), exOrakel(e))
			n++
			for _, c := range arten {
				e := a + "/" + b + "/" + c
				exPruefe(t, p, exQ(e), exOrakel(e))
				n++
			}
		}
	}
	exZaehlt(t, "TestExpandSegmentFolge", n, exFolgeZahl)
}
