package model_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Fänger der Klassen, die die Zeilen 4, 5, 8 und 15 über die Auswahl des Entwurfs hinaus tragen (AGENTS.md §3.17): das
// Produkt der Typmerkmale, Namen, die nur fast ein Programmfehler sind, und Listen weit über die Grenze des Parsers.

// exvFaelleTypachsen (Zeilen 4 und 5, SPEC-263 Nr. 2 a bis d und Nr. 4): zu jeder der acht Kombinationen aus offen, abstrakt
// und abgeleitet jeder Typ mit je einem Namen der vier Arten Navigation, Property, fremder Name und `*`. Das Orakel steht in
// der Schleife über die Achse offen, nicht im Typ: nur ein offener Typ meldet 501 für den fremden Namen, jeder andere 400,
// gleich ob abstrakt oder abgeleitet; `*` ist in jeder Kombination ein Programmfehler.
func exvFaelleTypachsen(m []exvT) []exvFall {
	var out []exvFall
	for _, offen := range []bool{false, true} {
		fremd := model.ExpandNavigationUnbekannt
		if offen {
			fremd = model.ExpandNavigationNichtUnterstuetzt
		}
		for _, abstrakt := range []bool{false, true} {
			for _, abgeleitet := range []bool{false, true} {
				for _, t := range m {
					if t.offen != offen || t.abstrakt != abstrakt || (t.basis != "") != abgeleitet {
						continue
					}
					kombination := fmt.Sprintf("%s offen=%v abstrakt=%v abgeleitet=%v", t.name, offen, abstrakt, abgeleitet)
					out = append(out, exvFall{kombination + " fremd", t.typ, exvO(exvUnbekannt(3)), string(fremd)},
						exvFall{kombination + " Stern", t.typ, exvO("*"), "programmfehler"})
					if len(t.navs) > 0 {
						out = append(out, exvFall{kombination + " Navigation", t.typ, exvO(t.navs[0]), "gueltig"})
					}
					if len(t.props) > 0 {
						out = append(out, exvFall{kombination + " Property", t.typ, exvO(t.props[0]), string(model.ExpandKeineNavigation)})
					}
				}
			}
		}
	}
	return out
}

// TestPruefeExpandTypachsen (Zeilen 4 und 5): alle acht Kombinationen aus offen, abstrakt und abgeleitet sind im Fixture-Modell
// geführt, und jeder Typ liefert für seine vier Namen das Ergebnis der Spezifikation.
func TestPruefeExpandTypachsen(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleTypachsen(m)
	soll := 0
	kombinationen := map[[3]bool]bool{}
	for _, tt := range m {
		z := exvZaehlen()[tt.name]
		soll += 2 + exvEins(z.navs > 0) + exvEins(z.props > 0)
		kombinationen[[3]bool{tt.offen, tt.abstrakt, tt.basis != ""}] = true
	}
	exvZaehlt(t, "Typachsen", len(faelle), soll)
	if len(kombinationen) != 2*2*2 {
		t.Errorf("%d von 8 Kombinationen im Fixture-Modell", len(kombinationen))
	}
	exvLaufe(t, faelle)
}

// exvFaelleFastProgrammfehler (Zeile 8, SPEC-263 Nr. 4 nennt allein den leeren Namen und `*`): Namen mit `*`, nur aus Leerraum
// oder mit angehängtem Leerraum sind keine Programmfehler, weder als fremder Name noch als Member eines Typs, und ein Paar
// aus einem Namen und seiner Form mit Leerraum ist kein doppelter Name.
func exvFaelleFastProgrammfehler(m []exvT) []exvFall {
	sonder := []string{"*x", "a*", "**", "*Name", " ", "  ", "\t", nbsp(), "\n"}
	var out []exvFall
	for _, t := range m {
		for _, n := range sonder {
			out = append(out, exvFall{fmt.Sprintf("%s %q", t.name, n), t.typ, exvO(n), exvFehlt(t)})
		}
		// Klasse je Name: nur diese Namen sind Member oder fremd, jede Form mit Leerraum ist fremd
		klassen := map[string]string{exvUnbekannt(4): exvFehlt(t)}
		if len(t.navs) > 0 {
			klassen[t.navs[0]] = "gueltig"
		}
		if len(t.props) > 0 {
			klassen[t.props[0]] = string(model.ExpandKeineNavigation)
		}
		klasse := func(x string) string {
			if k, ok := klassen[x]; ok {
				return k
			}
			return exvFehlt(t)
		}
		var namen []string
		for k := range klassen {
			namen = append(namen, k)
		}
		slices.Sort(namen)
		for _, n := range namen {
			for _, paar := range [][2]string{{n, n + " "}, {n + " ", n}, {" " + n, n}, {n, "\t" + n}} {
				soll := klasse(paar[0])
				if soll == "gueltig" {
					soll = klasse(paar[1])
				}
				out = append(out, exvFall{fmt.Sprintf("%s %q und %q", t.name, paar[0], paar[1]), t.typ, exvO(paar[0], paar[1]), soll})
			}
		}
	}
	return append(out, exvFaelleSondertyp()...)
}

// exvFaelleSondertyp führt Member, deren Namen `*` enthalten oder nur aus Leerraum bestehen: sie sind Navigationen und
// Properties wie jede andere.
func exvFaelleSondertyp() []exvFall {
	navs := []string{"*x", "a*", "**", "*Name"}
	props := []string{" ", "  ", "\t", nbsp(), "\n"}
	var out []exvFall
	for _, offen := range []bool{false, true} {
		typ := model.EntityType{Name: "Ns.Sonder", Open: offen}
		for _, n := range navs {
			typ.NavigationProperties = append(typ.NavigationProperties, exvN(n, "Ns.Sonder", false))
			out = append(out, exvFall{fmt.Sprintf("Navigation %q offen=%v", n, offen), typ, exvO(n), "gueltig"})
		}
		for _, p := range props {
			typ.Properties = append(typ.Properties, exvP(p, "Edm.String"))
			out = append(out, exvFall{fmt.Sprintf("Property %q offen=%v", p, offen), typ, exvO(p), string(model.ExpandKeineNavigation)})
		}
		out = append(out, exvFall{"alle Navigationen", typ, exvO(navs...), "gueltig"},
			exvFall{"alle Properties", typ, exvO(props...), string(model.ExpandKeineNavigation)},
			exvFall{"Navigation, Property", typ, exvO(navs[0], props[0]), string(model.ExpandKeineNavigation)},
			exvFall{"Property, Navigation", typ, exvO(props[0], navs[0]), string(model.ExpandKeineNavigation)})
	}
	return out
}

// exvLaengenListe sind die Längen der Listen von Zeile 15: jede Länge von 1 bis 70 und Längen über den Schwellen 100, 255,
// 256, 257, 1000 und 4096.
func exvLaengenListe() []int {
	var l []int
	for n := 1; n <= 70; n++ {
		l = append(l, n)
	}
	return append(l, 100, 255, 256, 257, 1000, 4096)
}

// exvSchwellen sind die Stellen um die plausiblen Grenzen einer Liste: 2^k - 1, 2^k und 2^k + 1.
func exvSchwellen() []int {
	return []int{31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256, 257, 1023, 1024, 1025, 4095}
}

// exvStellen sind die Stellen eines Verstoßes in einer Liste der Länge n: bis 70 jede Stelle, darüber die erste, die zweite,
// die mittlere, die zwei letzten und jede Schwelle unter n.
func exvStellen(n int) []int {
	if n <= 70 {
		out := make([]int, n)
		for i := range out {
			out[i] = i
		}
		return out
	}
	out := []int{0, 1, n / 2, n - 2, n - 1}
	for _, s := range exvSchwellen() {
		if s < n {
			out = append(out, s)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// exvDoppelPaare sind die Stellenpaare (i < j) eines doppelten Namens in einer Liste der Länge n.
func exvDoppelPaare(n int) [][2]int {
	var out [][2]int
	for _, p := range [][2]int{{0, n - 1}, {0, 1}, {n - 2, n - 1}, {n / 2, n - 1}} {
		if p[0] >= 0 && p[0] < p[1] && p[1] < n && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

// exvLeerStellen sind die Stellen eines leeren Namens und von `*` in einer Liste der Länge n.
func exvLeerStellen(n int) []int {
	out := []int{0, n / 2, n - 1}
	slices.Sort(out)
	return slices.Compact(out)
}

// exvFaelleLaengen (Zeile 15, SPEC-263 Nr. 6: keine eigene Grenze): in einem Typ mit 4096 Navigationen, offen und nicht offen,
// für jede Länge aus exvLaengenListe bis maxN die gültige Liste, ein Verstoß (fremd, Property) an jeder Stelle aus exvStellen,
// ein doppelter Name an den Paaren aus exvDoppelPaare, ein leerer Name und `*` an den Stellen aus exvLeerStellen.
func exvFaelleLaengen(maxN int) []exvFall {
	var out []exvFall
	for _, offen := range []bool{false, true} {
		typ := model.EntityType{Name: "Ns.Gross", Open: offen, Properties: []model.Property{exvP("GProp", "Edm.String")}}
		navs := make([]string, 4096)
		for i := range navs {
			navs[i] = fmt.Sprintf("G%05d", i)
			typ.NavigationProperties = append(typ.NavigationProperties, exvN(navs[i], "Ns.Kategorie", i%2 == 0))
		}
		fremd := string(model.ExpandNavigationUnbekannt)
		if offen {
			fremd = string(model.ExpandNavigationNichtUnterstuetzt)
		}
		mit := func(n, stelle int, name string) *model.ExpandOption {
			namen := slices.Clone(navs[:n])
			namen[stelle] = name
			return exvO(namen...)
		}
		for _, n := range exvLaengenListe() {
			if n > maxN {
				continue
			}
			art := fmt.Sprintf("offen=%v Länge %d", offen, n)
			out = append(out, exvFall{art + " gültig", typ, exvO(navs[:n]...), "gueltig"})
			for _, p := range exvStellen(n) {
				out = append(out, exvFall{fmt.Sprintf("%s fremd an Stelle %d", art, p), typ, mit(n, p, exvUnbekannt(1)), fremd},
					exvFall{fmt.Sprintf("%s Property an Stelle %d", art, p), typ, mit(n, p, "GProp"), string(model.ExpandKeineNavigation)})
			}
			for _, p := range exvDoppelPaare(n) {
				out = append(out, exvFall{fmt.Sprintf("%s doppelt an Stelle %v", art, p), typ, mit(n, p[1], navs[p[0]]), "programmfehler"})
			}
			for _, p := range exvLeerStellen(n) {
				for _, w := range []string{"", "*"} {
					out = append(out, exvFall{fmt.Sprintf("%s %q an Stelle %d", art, w, p), typ, mit(n, p, w), "programmfehler"})
				}
			}
		}
	}
	return out
}

// TestPruefeExpandLaenge (Zeile 15, SPEC-263 Nr. 6): die Funktion hat keine eigene Grenze; Listen der Länge 1 bis 70 mit einem
// Verstoß an jeder Stelle und Listen bis 4096 Elemente mit Verstößen an den Stellen um jede Schwelle sind gültig oder
// melden den ersten Verstoß, ein doppelter Name, ein leerer Name und `*` sind Programmfehler an jeder geführten Stelle.
func TestPruefeExpandLaenge(t *testing.T) {
	for _, n := range exvLaengenListe() {
		st := exvStellen(n)
		if !slices.IsSorted(st) || st[0] != 0 || st[len(st)-1] != n-1 || (n <= 70 && len(st) != n) {
			t.Errorf("Stellen für Länge %d: %v", n, st)
		}
		if n > 70 {
			for _, p := range []int{1, n / 2, n - 2} {
				if !slices.Contains(st, p) {
					t.Errorf("Länge %d: Stelle %d fehlt", n, p)
				}
			}
			for _, s := range exvSchwellen() {
				if s < n && !slices.Contains(st, s) {
					t.Errorf("Länge %d: Schwelle %d fehlt", n, s)
				}
			}
		}
	}
	soll := 0
	for _, n := range exvLaengenListe() {
		soll += 1 + 2*len(exvStellen(n)) + len(exvDoppelPaare(n)) + 2*len(exvLeerStellen(n))
	}
	faelle := exvFaelleLaengen(1 << 30)
	exvZaehlt(t, "Längen", len(faelle), 2*soll)
	exvLaufe(t, faelle)
}
