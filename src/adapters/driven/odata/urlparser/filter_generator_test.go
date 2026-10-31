package urlparser_test

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const flLaeufe = 2000

// flZufall liefert aus dem Zufall eines Laufs den Baum (alle Knotenarten, alle Literalarten, Tiefe bis 6) und zählt, was er
// erzeugt hat.
type flZufall struct {
	r     *rand.Rand
	arten map[string]int
}

func (z *flZufall) wahl(xs ...string) string { return xs[z.r.IntN(len(xs))] }

func (z *flZufall) atom() interface{ IstFilterAusdruck() } {
	lit := func(a model.LiteralArt, text string) interface{ IstFilterAusdruck() } {
		return model.FilterLiteral{Literal: model.Literal{Art: a, Text: text}}
	}
	switch k := z.r.IntN(8); k {
	case 0:
		z.arten["eigenschaft"]++
		return eig(z.wahl("Id", "Name", "x_1", "and", "has", "eq", "in", "Nullx", flRune(0xE4)+"b"))
	case 1:
		z.arten["null"]++
		return model.FilterNull{}
	case 2:
		z.arten["zahl"]++
		return lit(model.LiteralZahl, z.wahl("42", "-1.5", "0", "7.25"))
	case 3:
		z.arten["zeichenkette"]++
		return lit(model.LiteralZeichenkette, z.wahl("'a'", "'O''Hara'", "''", "'a b(c),d'", "'x)y'"))
	case 4:
		z.arten["wahrheitswert"]++
		return lit(model.LiteralWahrheitswert, z.wahl("true", "false"))
	case 5:
		z.arten["guid"]++
		return lit(model.LiteralGUID, flGUID)
	case 6:
		z.arten["datum"]++
		return lit(model.LiteralDatum, "2020-01-02")
	}
	z.arten["zeitpunkt"]++
	return lit(model.LiteralZeitpunkt, "2020-01-02T03:04:05Z")
}

func (z *flZufall) baum(tiefe int) interface{ IstFilterAusdruck() } {
	k := z.r.IntN(5)
	if tiefe == 0 || k == 0 {
		return z.atom()
	}
	switch k {
	case 1:
		z.arten["nicht"]++
		return nicht(z.baum(tiefe - 1))
	case 2:
		z.arten["verknuepfung"]++
		return bin(z.wahl("and", "or"), z.baum(tiefe-1), z.baum(tiefe-1))
	}
	z.arten["vergleich"]++
	return bin(z.wahl("eq", "ne", "gt", "ge", "lt", "le"), z.baum(tiefe-1), z.baum(tiefe-1))
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

// render schreibt den Baum mit minimaler Klammerung nach der Rangfolge und liefert die Stufe des Knotens: 0 bis 3 für
// `or`, `and`, `eq`/`ne`, `gt`/`ge`/`lt`/`le`; 5 für `not` und Operanden.
func (z *flZufall) render(a interface{ IstFilterAusdruck() }) (string, int) {
	klammer := func(s string) string { return "(" + z.leer(0, 3) + s + z.leer(0, 3) + ")" }
	switch n := a.(type) {
	case model.FilterNicht:
		s, st := z.render(n.Operand)
		if st < 5 {
			s = klammer(s)
		}
		return z.schreibweise("not") + z.leer(1, 3) + s, 5
	case model.FilterVergleich:
		return z.binaer(string(n.Op), n.Links, n.Rechts)
	case model.FilterVerknuepfung:
		return z.binaer(string(n.Op), n.Links, n.Rechts)
	case model.FilterNull:
		return z.schreibweise("null"), 5
	case model.FilterEigenschaft:
		return n.Name, 5
	case model.FilterLiteral:
		if n.Literal.Art == model.LiteralWahrheitswert {
			return z.schreibweise(n.Literal.Text), 5
		}
		return n.Literal.Text, 5
	}
	panic("Knoten ohne Darstellung")
}

func (z *flZufall) binaer(op string, l, r interface{ IstFilterAusdruck() }) (string, int) {
	stufe := flStufe(op)
	ls, lst := z.render(l)
	rs, rst := z.render(r)
	if lst < stufe {
		ls = "(" + z.leer(0, 3) + ls + z.leer(0, 3) + ")"
	}
	if rst <= stufe {
		rs = "(" + z.leer(0, 3) + rs + z.leer(0, 3) + ")"
	}
	return ls + z.leer(1, 3) + z.schreibweise(op) + z.leer(1, 3) + rs, stufe
}

// TestGeneratorFilter (Zeile 34): ein zufälliger Baum, vom Test mit minimaler Klammerung nach der Rangfolge zu Text gerendert
// (Schreibweise der Schlüsselwörter und 1 bis 3 Leerzeichen zufällig), vom Leser gelesen, ergibt denselben Baum. Orakel ist der
// Renderer des Tests, nie der Leser. Menge: 2000 Läufe, Saat 1 bis 2000.
func TestGeneratorFilter(t *testing.T) {
	z := &flZufall{arten: map[string]int{}}
	for saat := uint64(1); saat <= flLaeufe; saat++ {
		z.r = rng(saat)
		baum := z.baum(6)
		text, _ := z.render(baum)
		a, err := flLies(text)
		if err != nil || !gleich(a, baum) {
			t.Fatalf("Saat %d: %q\n  ist  %#v %v\n  soll %#v", saat, text, a, err, baum)
		}
	}
	for _, art := range []string{"eigenschaft", "null", "zahl", "zeichenkette", "wahrheitswert", "guid", "datum", "zeitpunkt", "nicht",
		"verknuepfung", "vergleich"} {
		if z.arten[art] < flLaeufe/10 {
			t.Errorf("Art %s nur %d-mal erzeugt", art, z.arten[art])
		}
	}
}
