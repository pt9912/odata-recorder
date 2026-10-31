package urlparser_test

import (
	"strings"
	"testing"
)

// flStufenOps sind die vier binären Stufen von schwach nach stark; die fünfte Stufe ist `not`.
func flStufenOps() [][]string {
	return [][]string{{"or"}, {"and"}, {"eq", "ne"}, {"gt", "ge", "lt", "le"}}
}

// TestVergleichOperanden (Zeile 2): jeder Vergleichsoperator mit jedem Paar von Operandenarten. Menge 54 = 6 × 9.
func TestVergleichOperanden(t *testing.T) {
	arten := []struct {
		text string
		node interface{ IstFilterAusdruck() }
	}{{"Id", eig("Id")}, {"5", zahl("5")}, {"null", flNull()}}
	n := 0
	for _, op := range []string{"eq", "ne", "gt", "ge", "lt", "le"} {
		for _, l := range arten {
			for _, r := range arten {
				flGut(t, l.text+" "+op+" "+r.text, bin(op, l.node, r.node))
				n++
			}
		}
	}
	if n != 54 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestVerknuepfungen (Zeile 3): `and`, `or` über zwei Eigenschaften und zwei Vergleiche, `not` über Eigenschaft und
// geklammerten Vergleich. Menge 6.
func TestVerknuepfungen(t *testing.T) {
	a, b := eig("a"), eig("b")
	v1, v2 := bin("eq", a, zahl("1")), bin("gt", b, zahl("2"))
	flGut(t, "a and b", bin("and", a, b))
	flGut(t, "a or b", bin("or", a, b))
	flGut(t, "a eq 1 and b gt 2", bin("and", v1, v2))
	flGut(t, "a eq 1 or b gt 2", bin("or", v1, v2))
	flGut(t, "not a", nicht(a))
	flGut(t, "not (a eq 1)", nicht(v1))
}

// TestRangfolge (Zeile 4): jedes Paar der fünf Stufen (or, and, eq/ne, gt/ge/lt/le, not) mit jedem Operator der Stufen, in
// beiden Schreibfolgen; der Baum gleich dem mit Klammer um die stärkere Stufe. Menge 58 = 2 × (21 + 8).
func TestRangfolge(t *testing.T) {
	stufen := flStufenOps()
	a, b, c := eig("a"), eig("b"), eig("c")
	n := 0
	for i := range stufen {
		for _, w := range stufen[i] {
			for j := i + 1; j < len(stufen); j++ {
				for _, s := range stufen[j] {
					flGut(t, "a "+w+" b "+s+" c", bin(w, a, bin(s, b, c)))
					flGut(t, "a "+s+" b "+w+" c", bin(w, bin(s, a, b), c))
					n += 2
				}
			}
			flGut(t, "a "+w+" not b", bin(w, a, nicht(b)))
			flGut(t, "not a "+w+" b", bin(w, nicht(a), b))
			n += 2
		}
	}
	flGut(t, "not Price gt 5", bin("gt", nicht(eig("Price")), zahl("5")))
	if n != 58 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestAssoziativitaet (Zeile 5): jeder der acht binären Operatoren, Ketten von 3 und 4 Operanden, links gebaut; dazu zwei
// gemischte Ketten derselben Stufe. Menge 18.
func TestAssoziativitaet(t *testing.T) {
	n := 0
	for _, op := range flOps() {
		flGut(t, "a "+op+" b "+op+" c", bin(op, bin(op, eig("a"), eig("b")), eig("c")))
		flGut(t, "a "+op+" b "+op+" c "+op+" d", bin(op, bin(op, bin(op, eig("a"), eig("b")), eig("c")), eig("d")))
		n += 2
	}
	flGut(t, "a eq 1 ne 2", bin("ne", bin("eq", eig("a"), zahl("1")), zahl("2")))
	flGut(t, "a gt 1 le 2", bin("le", bin("gt", eig("a"), zahl("1")), zahl("2")))
	if n+2 != 18 {
		t.Fatalf("%d Fälle", n+2)
	}
}

// TestKlammern (Zeile 6): Klammern überschreiben die Rangfolge und sind kein Knoten. Menge 67 = 50 Überschreibungen (jedes Paar
// von Operatoren zweier Stufen mit Klammer um die schwächere Seite, bei `not` eine Form) + 15 (Operand, Wurzel, Operand von
// `not` × Tiefe 1 bis 5) + 2 Assoziativ-Paare.
func TestKlammern(t *testing.T) {
	stufen := flStufenOps()
	a, b, c := eig("a"), eig("b"), eig("c")
	n := 0
	for i := range stufen {
		for _, w := range stufen[i] {
			for j := i + 1; j < len(stufen); j++ {
				for _, s := range stufen[j] {
					flGut(t, "(a "+w+" b) "+s+" c", bin(s, bin(w, a, b), c))
					flGut(t, "a "+s+" (b "+w+" c)", bin(s, a, bin(w, b, c)))
					n += 2
				}
			}
			flGut(t, "not (a "+w+" b)", nicht(bin(w, a, b)))
			n++
		}
	}
	for tiefe := 1; tiefe <= 5; tiefe++ {
		auf, zu := strings.Repeat("(", tiefe), strings.Repeat(")", tiefe)
		flGut(t, auf+"a eq 1"+zu, bin("eq", a, zahl("1")))
		flGut(t, auf+"a"+zu+" eq 1", bin("eq", a, zahl("1")))
		flGut(t, "not "+auf+"a"+zu, nicht(a))
		n += 3
	}
	flGut(t, "(a and b) and c", bin("and", bin("and", a, b), c))
	flGut(t, "a and (b and c)", bin("and", a, bin("and", b, c)))
	n += 2
	if n != 67 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestNull (Zeile 17): `null` in jeder der 16 Schreibweisen rechts und links, `null eq null`; `nullx`, `xnull`, `null1`,
// `_null` sind Namen. Menge 37.
func TestNull(t *testing.T) {
	n := 0
	for _, w := range flSchreibweisen("null") {
		flGut(t, "a eq "+w, bin("eq", eig("a"), flNull()))
		flGut(t, w+" eq a", bin("eq", flNull(), eig("a")))
		n += 2
	}
	flGut(t, "null eq null", bin("eq", flNull(), flNull()))
	for _, name := range []string{"nullx", "xnull", "null1", "_null"} {
		flGut(t, name+" eq 1", bin("eq", eig(name), zahl("1")))
		n++
	}
	if n+1 != 37 {
		t.Fatalf("%d Fälle", n+1)
	}
}

// TestWurzelOhneTyp (Zeile 18): jeder Operand ist als Wurzel lesbar, der Leser kennt das Modell nicht. Menge 9.
func TestWurzelOhneTyp(t *testing.T) {
	flGut(t, "Id", eig("Id"))
	flGut(t, "5", zahl("5"))
	flGut(t, "'x'", zk("'x'"))
	flGut(t, "null", flNull())
	flGut(t, "true", wahr("true"))
	flGut(t, "Id eq 'x'", bin("eq", eig("Id"), zk("'x'")))
	flGut(t, "1 eq 'a'", bin("eq", zahl("1"), zk("'a'")))
	flGut(t, "'a' eq 1", bin("eq", zk("'a'"), zahl("1")))
	flGut(t, "not 5", nicht(zahl("5")))
}

// TestFilterKnotenKorpus (Zeile 33): der Korpus der 40 Gutfälle ergibt den erwarteten Baum, und jeder Knoten ist einer der sechs
// Wertetypen mit Kindern ungleich nil.
func TestFilterKnotenKorpus(t *testing.T) {
	k := flKorpus()
	if len(k) != 40 {
		t.Fatalf("%d Gutfälle", len(k))
	}
	for _, f := range k {
		flGut(t, f.text, f.want)
		a, _ := flLies(f.text)
		if flZaehleKnoten(t, a) < 1 {
			t.Errorf("%q: kein Knoten", f.text)
		}
	}
}
