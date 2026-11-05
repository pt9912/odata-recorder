package urlparserv2_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestFlVergleiche (Zeile 1): jeder Vergleichsoperator mit jedem Paar von Operandenarten. Menge 54 = 6 × 9.
func TestFlVergleiche(t *testing.T) {
	arten := []flFall{{"Id", flEig("Id")}, {"5", flZahl("5")}, {"null", flNull()}}
	n := 0
	for _, op := range []string{"eq", "ne", "gt", "ge", "lt", "le"} {
		for _, l := range arten {
			for _, r := range arten {
				flGut(t, l.text+" "+op+" "+r.text, flBin(op, l.want, r.want))
				n++
			}
		}
	}
	if n != 54 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlVerknuepfungen (Zeile 2): `and` und `or` über zwei Eigenschaften und zwei Vergleiche, `not` über Eigenschaft und
// geklammerten Vergleich. Menge 6.
func TestFlVerknuepfungen(t *testing.T) {
	a, b := flEig("a"), flEig("b")
	v1, v2 := flBin("eq", a, flZahl("1")), flBin("gt", b, flZahl("2"))
	faelle := []flFall{
		{"a and b", flBin("and", a, b)}, {"a or b", flBin("or", a, b)},
		{"a eq 1 and b gt 2", flBin("and", v1, v2)}, {"a eq 1 or b gt 2", flBin("or", v1, v2)},
		{"not a", flNicht(a)}, {"not (a eq 1)", flNicht(v1)},
	}
	for _, f := range faelle {
		flGut(t, f.text, f.want)
	}
	if len(faelle) != 6 {
		t.Fatal(len(faelle))
	}
}

// TestFlRangfolge (Zeile 3): jedes Paar der fünf Stufen (or, and, eq/ne, gt/ge/lt/le, not) mit jedem Operator der Stufen, in
// beiden Schreibfolgen; der Baum ist gleich dem mit Klammer um die stärkere Stufe. Menge 58 = 2 × (21 + 8).
func TestFlRangfolge(t *testing.T) {
	stufen := flStufenOps()
	a, b, c := flEig("a"), flEig("b"), flEig("c")
	n := 0
	for i := range stufen {
		for _, w := range stufen[i] {
			for j := i + 1; j < len(stufen); j++ {
				for _, s := range stufen[j] {
					flGut(t, "a "+w+" b "+s+" c", flBin(w, a, flBin(s, b, c)))
					flGut(t, "a "+s+" b "+w+" c", flBin(w, flBin(s, a, b), c))
					n += 2
				}
			}
			flGut(t, "a "+w+" not b", flBin(w, a, flNicht(b)))
			flGut(t, "not a "+w+" b", flBin(w, flNicht(a), b))
			n += 2
		}
	}
	flGut(t, "not Price gt 5", flBin("gt", flNicht(flEig("Price")), flZahl("5")))
	if n != 58 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestFlAssoziativitaet (Zeile 4): jeder der acht binären Operatoren in Ketten von 3 und 4 Operanden, links gebaut; dazu zwei
// gemischte Ketten derselben Stufe. Menge 18.
func TestFlAssoziativitaet(t *testing.T) {
	a, b, c, d := flEig("a"), flEig("b"), flEig("c"), flEig("d")
	n := 0
	for _, op := range flOps() {
		flGut(t, "a "+op+" b "+op+" c", flBin(op, flBin(op, a, b), c))
		flGut(t, "a "+op+" b "+op+" c "+op+" d", flBin(op, flBin(op, flBin(op, a, b), c), d))
		n += 2
	}
	flGut(t, "a eq 1 ne 2", flBin("ne", flBin("eq", a, flZahl("1")), flZahl("2")))
	flGut(t, "a gt 1 le 2", flBin("le", flBin("gt", a, flZahl("1")), flZahl("2")))
	if n+2 != 18 {
		t.Fatalf("%d Fälle", n+2)
	}
}

// TestFlKlammern (Zeile 5): Klammern überschreiben die Rangfolge und sind kein Knoten. Menge 67 = 50 Überschreibungen (jedes Paar
// von Operatoren zweier Stufen mit Klammer um die schwächere Seite, 8 mit `not`) + 15 (Operand, Wurzel, Operand von `not` ×
// Tiefe 1 bis 5) + 2 Assoziativ-Paare.
func TestFlKlammern(t *testing.T) {
	stufen := flStufenOps()
	a, b, c := flEig("a"), flEig("b"), flEig("c")
	n := 0
	for i := range stufen {
		for _, w := range stufen[i] {
			for j := i + 1; j < len(stufen); j++ {
				for _, s := range stufen[j] {
					flGut(t, "(a "+w+" b) "+s+" c", flBin(s, flBin(w, a, b), c))
					flGut(t, "a "+s+" (b "+w+" c)", flBin(s, a, flBin(w, b, c)))
					n += 2
				}
			}
			flGut(t, "not (a "+w+" b)", flNicht(flBin(w, a, b)))
			n++
		}
	}
	for tiefe := 1; tiefe <= 5; tiefe++ {
		auf, zu := flWiederholt("(", tiefe), flWiederholt(")", tiefe)
		flGut(t, auf+"a eq 1"+zu, flBin("eq", a, flZahl("1")))
		flGut(t, auf+"a"+zu+" eq 1", flBin("eq", a, flZahl("1")))
		flGut(t, "not "+auf+"a"+zu, flNicht(a))
		n += 3
	}
	flGut(t, "(a and b) and c", flBin("and", flBin("and", a, b), c))
	flGut(t, "a and (b and c)", flBin("and", a, flBin("and", b, c)))
	if n+2 != 67 {
		t.Fatalf("%d Fälle", n+2)
	}
}

// TestFlNull (Zeile 20): `null` in jeder der 16 Schreibweisen rechts und links, `null eq null`; `nullx`, `xnull`, `null1`,
// `_null` sind Namen. Menge 37.
func TestFlNull(t *testing.T) {
	n := 0
	for _, w := range flSchreibweisen("null") {
		flGut(t, "a eq "+w, flBin("eq", flEig("a"), flNull()))
		flGut(t, w+" eq a", flBin("eq", flNull(), flEig("a")))
		n += 2
	}
	flGut(t, "null eq null", flBin("eq", flNull(), flNull()))
	for _, name := range []string{"nullx", "xnull", "null1", "_null"} {
		flGut(t, name+" eq 1", flBin("eq", flEig(name), flZahl("1")))
		n++
	}
	if n+1 != 37 {
		t.Fatalf("%d Fälle", n+1)
	}
}

// TestFlWurzelOhneTyp (Zeile 21): jeder Operand ist als Wurzel lesbar, der Leser kennt das Modell nicht. Menge 9.
func TestFlWurzelOhneTyp(t *testing.T) {
	faelle := []flFall{
		{"Id", flEig("Id")}, {"5", flZahl("5")}, {"'x'", flZk("'x'")}, {"null", flNull()}, {"true", flWahr("true")},
		{"Id eq 'x'", flBin("eq", flEig("Id"), flZk("'x'"))}, {"1 eq 'a'", flBin("eq", flZahl("1"), flZk("'a'"))},
		{"'a' eq 1", flBin("eq", flZk("'a'"), flZahl("1"))}, {"not 5", flNicht(flZahl("5"))},
	}
	for _, f := range faelle {
		flGut(t, f.text, f.want)
	}
	if len(faelle) != 9 {
		t.Fatal(len(faelle))
	}
}

// TestFlStruktur (Zeile 30): ein fehlender oder überzähliger Operand, Operator oder eine Klammer und das `,` sind `wert`.
// Menge 24.
func TestFlStruktur(t *testing.T) {
	k := flFehlerKorpus()
	for _, s := range k {
		flFehler(t, s, flU(model.URLWert))
	}
	if len(k) != 24 {
		t.Fatal(len(k))
	}
}

// TestFlKorpus (Zeile 31): der Korpus der 40 Gutfälle ergibt den erwarteten Baum aus den sechs Wertetypen mit Kindern ungleich
// nil.
func TestFlKorpus(t *testing.T) {
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
