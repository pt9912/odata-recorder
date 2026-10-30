package urlparser_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// stellenBasis ist eine Basisform eines Literals samt den Stellen, die das Orakel von Hand benennt: die Stellen einer
// Ziffer, die Stellen eines Trennzeichens (Strich, `T`, `:`, `.`, `Z`, Versatzzeichen) und die Stellen, an denen ein
// Zeichen gestrichen wird. Die Basis selbst ist ein Gutfall (`gut`) oder `literal_typ` (`basisU`).
type stellenBasis struct {
	name, form     string
	ziffern, trenn []int
	streichen      []int // nil: jede Stelle
	einfuegen      bool  // `0` an jeder der len+1 Stellen
	ohneErsatz     map[int][]string
	gutArt         model.LiteralArt
	gutText        string
	basisU         model.URLUrsache
	erwartet       int // Elemente ohne die Basis, aus der Rechnung von Hand
}

func zahlen(von, bis int) []int {
	var r []int
	for i := von; i <= bis; i++ {
		r = append(r, i)
	}
	return r
}

func vereint(teile ...[]int) []int {
	var r []int
	for _, t := range teile {
		r = append(r, t...)
	}
	slices.Sort(r)
	return r
}

// ersatzNichtZiffer sind die Ersatzzeichen einer Ziffernstelle (sechs); ersatzTrenn die einer Trennstelle (acht, zwei davon
// Ziffern, die ein Trennzeichen ersetzen und keine Ziffernstelle füllen).
var (
	ersatzNichtZiffer = []string{"x", "-", "+", ":", ".", "_"}
	ersatzTrenn       = []string{"x", "-", "+", ":", ".", "_", "0", "9"}
)

// stellenBasen sind die Basisformen von TestLiteralStellen: Datum, drei Zeitpunkte (mit Versatz, mit Sekunden, ohne Sekunden),
// Uhrzeit, ein Typpräfix, eine Guid und ein Zeitpunkt mit Bruchteil. Zählung je Basis aus der Rechnung von Hand (Ziffernstellen ×
// 6, Trennstellen × 7 oder 8 oder 6, Streichen, Einfügen) und nicht aus dem Lauf.
func stellenBasen() []stellenBasis {
	return []stellenBasis{
		{name: "B4 Datum", form: "2024-01-05", ziffern: vereint(zahlen(0, 3), zahlen(5, 6), zahlen(8, 9)), trenn: []int{4, 7}, einfuegen: true,
			gutArt: model.LiteralDatum, gutText: "2024-01-05", erwartet: 8*6 + 2*7 + 10 + 11},
		{name: "B1 Zeitpunkt mit Versatz", form: "2024-01-05T10:20:30+01:00",
			ziffern: vereint(zahlen(0, 3), zahlen(5, 6), zahlen(8, 9), zahlen(11, 12), zahlen(14, 15), zahlen(17, 18), zahlen(20, 21), zahlen(23, 24)),
			trenn:   []int{4, 7, 10, 13, 16, 19, 22}, einfuegen: true,
			ohneErsatz: map[int][]string{19: {"-"}}, // `-` an der Stelle des Vorzeichens ist ein gültiger Versatz
			gutArt:     model.LiteralZeitpunkt, gutText: "2024-01-05T09:20:30Z", erwartet: 18*6 + (7 + 7 + 8 + 7 + 7 + 6 + 7) + 25 + 26},
		{name: "B2 Zeitpunkt mit Sekunden", form: "2024-01-05T10:20:30Z",
			ziffern: vereint(zahlen(0, 3), zahlen(5, 6), zahlen(8, 9), zahlen(11, 12), zahlen(14, 15), zahlen(17, 18)),
			trenn:   []int{4, 7, 10, 13, 16, 19}, einfuegen: true,
			gutArt: model.LiteralZeitpunkt, gutText: "2024-01-05T10:20:30Z", erwartet: 14*6 + (7 + 7 + 8 + 7 + 7 + 8) + 20 + 21},
		{name: "B3 Zeitpunkt ohne Sekunden", form: "2024-01-05T10:20Z",
			ziffern: vereint(zahlen(0, 3), zahlen(5, 6), zahlen(8, 9), zahlen(11, 12), zahlen(14, 15)),
			trenn:   []int{4, 7, 10, 13, 16}, einfuegen: true,
			gutArt: model.LiteralZeitpunkt, gutText: "2024-01-05T10:20:00Z", erwartet: 12*6 + (7 + 7 + 8 + 7 + 8) + 17 + 18},
		{name: "B5 Uhrzeit", form: "12:30:45", ziffern: []int{0, 1, 3, 4, 6, 7}, trenn: []int{2, 5}, einfuegen: true,
			basisU: model.URLLiteralTyp, erwartet: 6*6 + 2*7 + 8 + 9},
		{name: "B7 Guid", form: "0f8fad5b-d9cb-469f-a165-70867728950e",
			ziffern: vereint(zahlen(0, 7), zahlen(9, 12), zahlen(14, 17), zahlen(19, 22), zahlen(24, 35)), trenn: []int{8, 13, 18, 23}, einfuegen: true,
			gutArt: model.LiteralGUID, gutText: "0f8fad5b-d9cb-469f-a165-70867728950e", erwartet: 32*6 + 4*7 + 36 + 37},
		// B8: nur die Stellen des Bruchteils: `.` (Stelle 19), die Ziffer (20, allein ihr Ersatz durch eine Nicht-Ziffer) und `Z` (21)
		{name: "B8 Zeitpunkt mit Bruchteil", form: "2024-01-05T10:20:30.5Z", ziffern: []int{20}, trenn: []int{19, 21}, streichen: []int{19, 20, 21},
			gutArt: model.LiteralZeitpunkt, gutText: "2024-01-05T10:20:30.5Z", erwartet: 7 + 8 + 6 + 3},
	}
}

// stellenElement ist eine Eingabe der Menge samt der Erwartung von Hand: `u` leer ist gelesen.
type stellenElement struct {
	beschreibung, form string
	u                  model.URLUrsache
	art                model.LiteralArt
	text               string
}

// erzeugeStellen baut die Elemente einer Basis: die Basis, jede Stelle × jedes Ersatzzeichen, Streichen, Einfügen von `0`.
// Jedes Element außer der Basis ist `literal`.
func (b stellenBasis) erzeugeStellen() []stellenElement {
	els := []stellenElement{{beschreibung: b.name + " Basis", form: b.form, u: b.basisU, art: b.gutArt, text: b.gutText}}
	ersetze := func(i int, z string) string { return b.form[:i] + z + b.form[i+1:] }
	for _, i := range b.ziffern {
		for _, z := range ersatzNichtZiffer {
			els = append(els, stellenElement{fmt.Sprintf("%s Stelle %d (Ziffer) -> %q", b.name, i, z), ersetze(i, z), model.URLLiteral, 0, ""})
		}
	}
	for _, i := range b.trenn {
		for _, z := range ersatzTrenn {
			if string(b.form[i]) == z || slices.Contains(b.ohneErsatz[i], z) {
				continue
			}
			els = append(els, stellenElement{fmt.Sprintf("%s Stelle %d (Trenner) -> %q", b.name, i, z), ersetze(i, z), model.URLLiteral, 0, ""})
		}
	}
	streichen := b.streichen
	if streichen == nil {
		streichen = zahlen(0, len(b.form)-1)
	}
	for _, i := range streichen {
		els = append(els, stellenElement{fmt.Sprintf("%s Stelle %d gestrichen", b.name, i), b.form[:i] + b.form[i+1:], model.URLLiteral, 0, ""})
	}
	if b.einfuegen {
		for i := 0; i <= len(b.form); i++ {
			els = append(els, stellenElement{fmt.Sprintf("%s vor Stelle %d eine 0 eingefügt", b.name, i), b.form[:i] + "0" + b.form[i:], model.URLLiteral, 0, ""})
		}
	}
	return els
}

// praefixElemente baut die Menge der Typpräfixe (B6): jede der sechs Stellen von `aA0_.b` vor `'x'`, ersetzt durch jedes Zeichen von K
// (bleibt `literal_typ`; Buchstabe, Ziffer, `_`, `.`, auch jenseits von ASCII) und von I (wird `literal`), ohne das Zeichen, das dort
// schon steht. Die Zeichen jenseits von ASCII stehen als Bytewerte.
func praefixElemente() []stellenElement {
	const basis = "aA0_.b"
	k := []string{"a", "A", "b", "z", "\xc3\xa4", "\xd9\xa3", "0", "9", "_", "."} // ä, arabisch-indische Ziffer drei
	ii := []string{"-", "+", ":", "$", "@", "*", "~", "!"}
	els := []stellenElement{{beschreibung: "B6 Basis", form: basis + "'x'", u: model.URLLiteralTyp}}
	for i := range len(basis) {
		for _, z := range k {
			if string(basis[i]) == z {
				continue
			}
			els = append(els, stellenElement{fmt.Sprintf("B6 Stelle %d -> K %q", i, z), basis[:i] + z + basis[i+1:] + "'x'", model.URLLiteralTyp, 0, ""})
		}
		for _, z := range ii {
			els = append(els, stellenElement{fmt.Sprintf("B6 Stelle %d -> I %q", i, z), basis[:i] + z + basis[i+1:] + "'x'", model.URLLiteral, 0, ""})
		}
	}
	return els
}

// pruefeStellenElement liest `/S(Id=<form>)` und hält das Ergebnis gegen die Erwartung von Hand: Ursache und Ort bei einem Verstoß,
// Art und Text bei einem Gutfall. Eine Panik des Lesers (ein Zugriff außerhalb der Länge) ist ein Befund dieses Elements.
func pruefeStellenElement(t *testing.T, e stellenElement) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s (%q): Panik %v", e.beschreibung, e.form, r)
		}
	}()
	u, err := parse("/S(Id="+e.form+")", "")
	if e.u == "" {
		if err != nil || len(u.Ressource.Schluessel) != 1 || u.Ressource.Schluessel[0].Literal != (model.Literal{Art: e.art, Text: e.text}) {
			t.Errorf("%s (%q): gelesen %+v %v, erwartet Art %v Text %q", e.beschreibung, e.form, u.Ressource.Schluessel, err, e.art, e.text)
		}
		return
	}
	var uf *model.URLFehler
	if !errors.As(err, &uf) || uf.Ursache != e.u || uf.Ort != "pfad/1" {
		t.Errorf("%s (%q): %v, erwartet %s am Ort pfad/1", e.beschreibung, e.form, err, e.u)
	}
}

// TestLiteralStellen (Zeilen 35 bis 38, SPEC-161 Nr. 6 d bis g, Klasse „Trenn-, Längen- und Namenszeichen der Literalformen“):
// Schleife über Basisform × Stelle × Ersatzzeichen, dazu Streichen und Einfügen einer Ziffer an jeder Stelle. Jedes Element
// außer den acht Basisformen ist `literal`, im Typpräfix bleibt ein Buchstabe, eine Ziffer, `_` oder `.` ein `literal_typ`.
// Menge 1098: 83 + 208 + 169 + 144 + 67 + 102 + 293 + 24 Elemente der Stellen und die 8 Basisformen als Gutfälle; die Zählprobe
// steht je Basis (aus der Rechnung von Hand) und für die Summe.
func TestLiteralStellen(t *testing.T) {
	ohneBasis, basen := 0, 0
	for _, b := range stellenBasen() {
		els := b.erzeugeStellen() // das erste Element ist die Basis
		if got := len(els) - 1; got != b.erwartet {
			t.Errorf("%s: %d Elemente, erwartet %d", b.name, got, b.erwartet)
		}
		for _, e := range els {
			pruefeStellenElement(t, e)
		}
		ohneBasis += len(els) - 1
		basen++
	}
	praefix := praefixElemente()
	for _, e := range praefix {
		pruefeStellenElement(t, e)
	}
	if got := len(praefix) - 1; got != 6*(9+8) {
		t.Errorf("B6: %d Elemente, erwartet 102", got)
	}
	ohneBasis += len(praefix) - 1
	basen++
	if ohneBasis != 83+208+169+144+67+102+293+24 || basen != 8 || ohneBasis+basen != 1098 {
		t.Fatalf("Zählprobe: %d Elemente und %d Basisformen, erwartet 1090 und 8 (1098)", ohneBasis, basen)
	}
}
