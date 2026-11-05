package urlparserv2_test

import (
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const flLaeufe = 2000

// flGUIDZufall ist die Guid der erzeugten Bäume (kleingeschrieben, der Text des Literals).
const flGUIDZufall = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

// flZufall erzeugt aus dem Zufall eines Laufs den Baum (alle Knotenarten, alle Literalarten aus Zeile 18, Tiefe bis 6), rendert
// ihn zu Text und hält die Verstöße fest, die der Renderer in den Text sprengt, in der Reihenfolge des Textes.
type flZufall struct {
	r            *rand.Rand
	arten        map[string]int
	einsprengsel bool
	verstoesse   []flVerstoss
}

// flVerstoss ist ein Einsprengsel mit seiner Klasse und dem Urteil, das der Leser an seiner Stelle fällen muss.
type flVerstoss struct {
	klasse string
	v      flV
}

func (z *flZufall) wahl(xs ...string) string { return xs[z.r.IntN(len(xs))] }

func (z *flZufall) atom() interface{ IstFilterAusdruck() } {
	switch k := z.r.IntN(8); k {
	case 0:
		z.arten["eigenschaft"]++
		return flEig(z.wahl("Id", "Name", "x_1", "and", "has", "eq", "in", "Nullx", flRune(0xE4)+"b"))
	case 1:
		z.arten["null"]++
		return flNull()
	case 2:
		z.arten["zahl"]++
		return flZahl(z.wahl("42", "-1.5", "0", "7.25"))
	case 3:
		z.arten["zeichenkette"]++
		return flZk(z.wahl("'a'", "'O''Hara'", "''", "'a b(c),d'", "'x)y'"))
	case 4:
		z.arten["wahrheitswert"]++
		return flWahr(z.wahl("true", "false"))
	case 5:
		z.arten["guid"]++
		return flLit(model.LiteralGUID, flGUIDZufall)
	}
	z.arten["zeitpunkt"]++
	return flLit(model.LiteralZeitpunkt, "2020-01-02T03:04:05Z")
}

func (z *flZufall) baum(tiefe int) interface{ IstFilterAusdruck() } {
	k := z.r.IntN(5)
	if tiefe == 0 || k == 0 {
		return z.atom()
	}
	switch k {
	case 1:
		z.arten["nicht"]++
		return flNicht(z.baum(tiefe - 1))
	case 2:
		z.arten["verknuepfung"]++
		return flBin(z.wahl("and", "or"), z.baum(tiefe-1), z.baum(tiefe-1))
	}
	z.arten["vergleich"]++
	return flBin(z.wahl("eq", "ne", "gt", "ge", "lt", "le"), z.baum(tiefe-1), z.baum(tiefe-1))
}

func (z *flZufall) leer(von, bis int) string { return strings.Repeat(" ", von+z.r.IntN(bis-von+1)) }

// schreibweise setzt jeden Buchstaben des Schlüsselworts zufällig groß oder klein.
func (z *flZufall) schreibweise(w string) string {
	b := []byte(w)
	for i := range b {
		if z.r.IntN(2) == 1 {
			b[i] = b[i] &^ 0x20
		}
	}
	return string(b)
}

// flEinsprengsel sind die Klassen außerhalb des Umfangs mit je einer Auswahl an Formen und dem Urteil des Lesers. Der Operator
// steht am Platz eines Operators, alle anderen am Platz eines Operanden.
func flEinsprengsel() map[string]struct {
	formen []string
	v      flV
} {
	f := flK(model.KompatFilter)
	return map[string]struct {
		formen []string
		v      flV
	}{
		"operator": {[]string{"add", "sub", "mul", "div", "mod", "has", "in", "divby"}, f},
		"funktion": {[]string{"foo(1)", "substringof(Name)", "contains(Id,1)", "Items/any(x)"}, f},
		"pfad":     {[]string{"a/b", "a/b/c", "Ns.Typ"}, f},
		"minus":    {[]string{"-Id", "-(1)", "$it"}, f},
		"typ":      {[]string{"X'a'", "1e3", "1.5d", "binary'AA'"}, flK(model.KompatLiteralTyp)},
		"v4":       {[]string{"2020-01-01", flGUIDZufall, "12:30"}, flU(model.URLWert)},
	}
}

// sprenge entscheidet, ob an dieser Stelle ein Einsprengsel der Klassen steht, und hält es fest. Der Operator wird nur am Platz
// eines Operators, alle anderen nur am Platz eines Operanden gewählt.
func (z *flZufall) sprenge(amOperator bool) (string, bool) {
	if !z.einsprengsel || z.r.IntN(12) != 0 {
		return "", false
	}
	klassen := []string{"funktion", "pfad", "minus", "typ", "v4"}
	if amOperator {
		klassen = []string{"operator"}
	}
	k := klassen[z.r.IntN(len(klassen))]
	e := flEinsprengsel()[k]
	z.verstoesse = append(z.verstoesse, flVerstoss{klasse: k, v: e.v})
	return e.formen[z.r.IntN(len(e.formen))], true
}

// literalText wählt eine Schreibweise des Literals, deren kanonischer Text der des Baums ist.
func (z *flZufall) literalText(l model.Literal) string {
	switch l.Art {
	case model.LiteralWahrheitswert:
		return z.schreibweise(l.Text)
	case model.LiteralGUID:
		return "guid'" + z.wahl(l.Text, strings.ToUpper(l.Text)) + "'"
	case model.LiteralZeitpunkt:
		return "datetime'" + strings.TrimSuffix(l.Text, "Z") + "'"
	case model.LiteralZahl:
		return z.wahl(map[string][]string{"42": {"42", "42L", "42M", "042"}, "-1.5": {"-1.5", "-1.50", "-1.5M"}, "0": {"0", "-0", "0.0"},
			"7.25": {"7.25", "7.250", "7.25M"}}[l.Text]...)
	}
	return l.Text
}

// render schreibt den Baum mit minimaler Klammerung nach der Rangfolge und liefert die Stufe des Knotens: 0 bis 3 für `or`,
// `and`, `eq`/`ne`, `gt`/`ge`/`lt`/`le`; 5 für `not` und Operanden. Die Verstöße entstehen in der Reihenfolge des Textes.
func (z *flZufall) render(a flBaum) (string, int) {
	klammer := func(s string) string { return "(" + z.leer(0, 3) + s + z.leer(0, 3) + ")" }
	switch n := a.(type) {
	case model.FilterNicht:
		sw := z.schreibweise("not") + z.leer(1, 3)
		s, st := z.render(n.Operand)
		if st < 5 {
			s = klammer(s)
		}
		return sw + s, 5
	case model.FilterVergleich:
		return z.binaer(string(n.Op), n.Links, n.Rechts)
	case model.FilterVerknuepfung:
		return z.binaer(string(n.Op), n.Links, n.Rechts)
	}
	if s, ok := z.sprenge(false); ok {
		return s, 5
	}
	switch n := a.(type) {
	case model.FilterNull:
		return z.schreibweise("null"), 5
	case model.FilterEigenschaft:
		return n.Name, 5
	case model.FilterLiteral:
		return z.literalText(n.Literal), 5
	}
	panic("Knoten ohne Darstellung")
}

func (z *flZufall) binaer(op string, l, r flBaum) (string, int) {
	stufe := flStufe(op)
	ls, lst := z.render(l)
	if lst < stufe {
		ls = "(" + z.leer(0, 3) + ls + z.leer(0, 3) + ")"
	}
	opText, ersetzt := z.sprenge(true)
	if !ersetzt {
		opText = z.schreibweise(op)
	}
	rs, rst := z.render(r)
	if rst <= stufe {
		rs = "(" + z.leer(0, 3) + rs + z.leer(0, 3) + ")"
	}
	return ls + z.leer(1, 3) + opText + z.leer(1, 3) + rs, stufe
}

// TestFlGenerator (Zeile 36): ein zufälliger Baum, vom Test mit minimaler Klammerung nach der Rangfolge zu Text gerendert
// (Schreibweise der Schlüsselwörter, der Literale und 1 bis 3 Leerzeichen zufällig), vom Leser gelesen, ergibt denselben Baum;
// mit einem Einsprengsel aus den Klassen außerhalb (Operator, Funktion, Pfad, Minus, Literal fremden Typs, V4-Form) an
// zufälliger Stelle ist die Ursache die des ersten Verstoßes im Text. Orakel ist der Renderer mit seiner Verdikt-Regel, nie der
// Leser. Menge: 2000 Läufe, Saat 1 bis 2000, in jeder zweiten Saat mit Einsprengseln. Prüfgrenze: die Klasse der Ausdrücke ist
// unbeschränkt, die Saaten sind eine Auswahl.
func TestFlGenerator(t *testing.T) {
	z := &flZufall{arten: map[string]int{}}
	ersteKlassen := map[string]int{}
	for saat := uint64(1); saat <= flLaeufe; saat++ {
		z.r, z.einsprengsel, z.verstoesse = flRng(saat), saat%2 == 0, nil
		baum := z.baum(6)
		text, _ := z.render(baum)
		if len(z.verstoesse) == 0 {
			a, v := flUrteil(t, text)
			if v != (flV{}) || !reflect.DeepEqual(a, baum) {
				t.Fatalf("Saat %d: %q\n  ist  %#v %v\n  soll %#v", saat, text, a, v, baum)
			}
			continue
		}
		ersteKlassen[z.verstoesse[0].klasse]++
		if _, v := flUrteil(t, text); v != z.verstoesse[0].v {
			t.Fatalf("Saat %d: %q: %v, erwartet %v (%s)", saat, text, v, z.verstoesse[0].v, z.verstoesse[0].klasse)
		}
	}
	for _, art := range []string{"eigenschaft", "null", "zahl", "zeichenkette", "wahrheitswert", "guid", "zeitpunkt", "nicht", "verknuepfung", "vergleich"} {
		if z.arten[art] < flLaeufe/10 {
			t.Errorf("Art %s nur %d-mal erzeugt", art, z.arten[art])
		}
	}
	for k := range flEinsprengsel() {
		if ersteKlassen[k] < 40 {
			t.Errorf("Klasse %s nur %d-mal als erster Verstoß getroffen", k, ersteKlassen[k])
		}
	}
	t.Logf("erste Verstöße je Klasse: %v", ersteKlassen)
}
