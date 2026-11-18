package urlparser_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestExpandNamenGrenze (Zeile 9, SPEC-262 Nr. 4 a, 6): ein Name hat 1 bis 128 Zeichen (nicht Bytes) und beginnt mit Buchstabe oder
// `_`. Menge 24: Länge {0, 1, 128, 129} × erstes Zeichen {Buchstabe, `_`, Ziffer} × Breite der Füllzeichen {1 Byte, 2 Bytes}.
func TestExpandNamenGrenze(t *testing.T) {
	p := exParser()
	n := 0
	for _, laenge := range []int{0, 1, 128, 129} {
		for _, breit := range []bool{false, true} {
			fuell, buchstabe := "a", "a"
			if breit {
				fuell, buchstabe = exE, exE
			}
			for _, erstes := range []string{buchstabe, "_", "1"} {
				name := ""
				if laenge > 0 {
					name = erstes + strings.Repeat(fuell, laenge-1)
				}
				soll := exErg(name)
				if laenge == 0 || laenge > 128 || erstes == "1" {
					soll = exWert()
				}
				exPruefe(t, p, exQ(name), soll)
				n++
			}
		}
	}
	exZaehlt(t, "Namengrenze", n, 24)
}

// TestExpandStern (Zeile 10, SPEC-262 Nr. 4 b): `*` allein oder neben Namen an jeder Position ist `wert_nicht_unterstuetzt`.
// Menge: 0 bis 3 Namen × alle Positionen (1 + 2 + 3 + 4 = 10); `**` und `*a` sind `wert`.
func TestExpandStern(t *testing.T) {
	p := exParser()
	n := 0
	for zahl := 0; zahl <= 3; zahl++ {
		for pos := 0; pos <= zahl; pos++ {
			namen := exNamen("n", zahl)
			elemente := append(append(append([]string{}, namen[:pos]...), "*"), namen[pos:]...)
			exPruefe(t, p, exQ(strings.Join(elemente, ",")), exNichtUnterstuetzt())
			n++
		}
	}
	exZaehlt(t, "Stern", n, 10)
	for _, w := range []string{"**", "*a", "a*", "*0", "*_"} {
		exPruefe(t, p, exQ(w), exWert())
	}
}

// TestExpandPfad (Zeile 11, SPEC-262 Nr. 4 c): ein Pfad aus 2 bis 9 Namen ist `wert_nicht_unterstuetzt`; ein leeres Segment an
// jeder Stelle ist `wert`. Menge: 8 Pfade und 2 + 3 + … + 9 = 44 Pfade mit leerem Segment. Prüfgrenze: 9 Segmente.
func TestExpandPfad(t *testing.T) {
	p := exParser()
	n := 0
	for segmente := 2; segmente <= 9; segmente++ {
		namen := exNamen("s", segmente)
		exPruefe(t, p, exQ(strings.Join(namen, "/")), exNichtUnterstuetzt())
		n++
		for leer := 0; leer < segmente; leer++ {
			mit := append([]string{}, namen...)
			mit[leer] = ""
			exPruefe(t, p, exQ(strings.Join(mit, "/")), exWert())
			n++
		}
	}
	exZaehlt(t, "Pfad", n, 8+44)
	exPruefe(t, p, exQ("/"), exWert())
	exPruefe(t, p, exQ("//"), exWert())
}

// TestExpandBesondereSegmente (Zeile 12, SPEC-262 Nr. 4 c, d): die Segmente eines Pfads sind Name, qualifizierter Name, `$ref`,
// `$count`, `$value`; jedes andere Segment macht das Element zu `wert`. Menge: 11 Segmente × 3 Stellen (erstes, mittleres,
// letztes eines Pfads aus drei Segmenten) und 11 Segmente allein; das Orakel ist die Tabelle `zulaessig`.
func TestExpandBesondereSegmente(t *testing.T) {
	p := exParser()
	zulaessig := map[string]bool{"Nav": true, "Ns.Typ": true, "a.b.c": true, "$ref": true, "$count": true, "$value": true,
		"*": false, "$levels": false, "$": false, "a.": false, ".a": false}
	n := 0
	for seg, ok := range zulaessig {
		for stelle := 0; stelle < 3; stelle++ {
			teile := []string{"Nav", "Nav", "Nav"}
			teile[stelle] = seg
			soll := exWert()
			if ok {
				soll = exNichtUnterstuetzt()
			}
			exPruefe(t, p, exQ(strings.Join(teile, "/")), soll)
			n++
		}
		allein := exWert()
		switch {
		case seg == "Nav":
			allein = exErg("Nav")
		case seg == "*":
			allein = exNichtUnterstuetzt()
		}
		exPruefe(t, p, exQ(seg), allein)
		n++
	}
	exZaehlt(t, "Segmente", n, 11*4)
}

// TestExpandKlammer (Zeile 13, SPEC-262 Nr. 3 b): ein `(` irgendwo im Wert macht den ganzen Wert zu `wert_nicht_unterstuetzt`, der
// Rest wird nicht gelesen. Menge: 9 Systemoptionen × 3 Stellen im Element × 7 Vorläufe (kein, leeres Element, ungültiger Name,
// doppeltes Element, 33 Elemente, 1 und 2 gültige Elemente) × 2 Nachläufe (kein, `,b`) = 378, dazu 7 feste Formen.
func TestExpandKlammer(t *testing.T) {
	p := exParser()
	optionen := []string{"$filter=x", "$select=Id,Name", "$orderby=Id", "$top=1", "$skip=1", "$count=true", "$levels=2", "$expand=b", "$search=x"}
	stellen := func(opt string) [3]string {
		return [3]string{"(" + opt + ")", "a(" + opt + ")", "a" + opt + "("}
	}
	vorlaeufe := []string{"", ",", "1x,", "a,a,", strings.Join(exNamen("N", 33), ",") + ",", "a,", "a,b,"}
	n := 0
	for _, opt := range optionen {
		for _, el := range stellen(opt) {
			for _, vor := range vorlaeufe {
				for _, nach := range []string{"", ",b"} {
					exPruefe(t, p, exQ(vor+el+nach), exNichtUnterstuetzt())
					n++
				}
			}
		}
	}
	exZaehlt(t, "Klammer", n, 9*3*7*2)
	fest := []struct {
		wert string
		soll exSoll
	}{
		{"a(", exNichtUnterstuetzt()}, {"(", exNichtUnterstuetzt()}, {"a()", exNichtUnterstuetzt()}, {"a)", exWert()},
		{")a(", exNichtUnterstuetzt()}, {"a($select=Id,Name),b", exNichtUnterstuetzt()}, {",(", exNichtUnterstuetzt()},
	}
	for _, f := range fest {
		exPruefe(t, p, exQ(f.wert), f.soll)
	}
	exZaehlt(t, "Klammer fest", len(fest), 7)
}

// TestExpandZahl (Zeile 14, SPEC-262 Nr. 3 c, 6): bis 32 Elemente sind lesbar, 33 sind `zu_viele`, vor der Prüfung der Elemente;
// der Ort ist `query/<n>`. Menge: Zahl {31, 32, 33, 34, 65, 200} × Verstoß {keiner, erstes, letztes Element ungültig} × Stelle der Option {1, 2, 3},
// dazu 31, 32 und 33 leere Elemente.
func TestExpandZahl(t *testing.T) {
	p := exParser()
	n := 0
	for _, zahl := range []int{31, 32, 33, 34, 65, 200} {
		for verstoss := 0; verstoss < 3; verstoss++ {
			namen := exNamen("N", zahl)
			switch verstoss {
			case 1:
				namen[0] = "1a"
			case 2:
				namen[zahl-1] = "1a"
			}
			for stelle := 1; stelle <= 3; stelle++ {
				soll := exErg(namen...)
				ort := "query/" + strconv.Itoa(stelle)
				switch {
				case zahl > 32:
					soll = exFehl(model.URLZuViele, ort)
				case verstoss > 0:
					soll = exFehl(model.URLWert, ort)
				}
				query := optionenDavor(stelle-1) + exQ(strings.Join(namen, ","))
				if u, err := p.ParseURL("/S", query); soll.u == "" && (err != nil || !gleich(u.Optionen.Expand, exKnoten(namen...).Optionen.Expand)) {
					t.Errorf("%.40q: %+v %v", query, u.Optionen.Expand, err)
				} else if soll.u != "" {
					exPruefe(t, p, query, soll)
				}
				n++
			}
		}
	}
	exZaehlt(t, "Zahl", n, 54)
	for kommas, soll := range map[int]exSoll{30: exWert(), 31: exWert(), 32: exZuViele()} {
		exPruefe(t, p, exQ(strings.Repeat(",", kommas)), soll)
	}
}

// optionenDavor sind n Optionen vor `$expand`, je mit eigenem Namen.
func optionenDavor(n int) string {
	var s string
	for i := range n {
		s += fmt.Sprintf("x%d=1&", i)
	}
	return s
}

// TestExpandDoppeltesElement (Zeile 15, SPEC-262 Nr. 4 e): ein Element, das einem früheren bytegleich ist, ist `wert`, auch `*`
// und ein Pfad; die Schreibweise zählt. Menge: Abstand 0 bis 3 × Form {Name, `*`, Pfad} × {gleich, Schreibweise geändert}; für `*`
// gibt es keine geänderte Schreibweise (3 × (2 + 1 + 2) = 20 Fälle).
func TestExpandDoppeltesElement(t *testing.T) {
	p := exParser()
	n := 0
	formen := []struct {
		a, b   string
		gleich exSoll
		anders func(a, b string, fuell []string) exSoll
	}{
		{"Nx", "nx", exWert(), func(a, b string, f []string) exSoll { return exErg(append(append([]string{a}, f...), b)...) }},
		{"*", "*", exWert(), nil},
		{"p/q", "P/q", exWert(), func(string, string, []string) exSoll { return exNichtUnterstuetzt() }},
	}
	for abstand := 0; abstand <= 3; abstand++ {
		fuell := exNamen("f", abstand)
		for _, f := range formen {
			exPruefe(t, p, exQ(strings.Join(append(append([]string{f.a}, fuell...), f.a), ",")), f.gleich)
			n++
			if f.anders != nil {
				exPruefe(t, p, exQ(strings.Join(append(append([]string{f.a}, fuell...), f.b), ",")), f.anders(f.a, f.b, fuell))
				n++
			}
		}
	}
	exZaehlt(t, "doppeltes Element", n, 4*5)
}

// exVerstoss ist ein Element, das einen Verstoß trägt, mit seiner Klasse: `klammer` (ein `(`), `klasse501` (Klasse
// `nicht_unterstuetzt`), sonst `ungueltig`. `doppelt` ist bytegleich dem Element `n1`, das jede Liste an erster Stelle führt.
type exVerstoss struct {
	name, element string
	klammer       bool
	klasse501     bool
}

func exVerstoesse() []exVerstoss {
	return []exVerstoss{
		{"leer", "", false, false}, {"ungueltig", "1a", false, false}, {"doppelt", "n1", false, false},
		{"stern", "*", false, true}, {"pfad", "p/q", false, true}, {"klammer", "k(x)", true, true},
	}
}

// TestExpandRangImWert (Zeile 16, SPEC-262 Nr. 3, 8): ein `(` geht allem vor, sonst geht ein Verstoß der Klasse `ungueltig` jedem
// der anderen Klasse vor, gleich an welcher Stelle. Menge: 6 × 6 Paare × 12 Stellenpaare (i ≠ j aus 1 bis 4) = 432; zweimal `*`
// ist ein doppeltes Element und damit `wert`.
func TestExpandRangImWert(t *testing.T) {
	p := exParser()
	art := exVerstoesse()
	n := 0
	for _, a := range art {
		for _, b := range art {
			for i := 1; i <= 4; i++ {
				for j := 1; j <= 4; j++ {
					if i == j {
						continue
					}
					elemente := []string{"n1", "f2", "f3", "f4", "f5"}
					elemente[i], elemente[j] = a.element, b.element
					if a.element == "p/q" && b.element == "p/q" {
						elemente[j] = "p/r"
					}
					soll := exWert()
					switch {
					case a.klammer || b.klammer:
						soll = exNichtUnterstuetzt()
					case a.klasse501 && b.klasse501 && !(a.element == "*" && b.element == "*"):
						soll = exNichtUnterstuetzt()
					}
					exPruefe(t, p, exQ(strings.Join(elemente, ",")), soll)
					n++
				}
			}
		}
	}
	exZaehlt(t, "Rang im Wert", n, 432)
}

// TestExpandRangQuery (Zeile 17, SPEC-160 Nr. 5): ein Verstoß von `$expand` und einer einer anderen Option folgen der Rangfolge,
// in beiden Reihenfolgen. Menge: 4 Verstöße von `$expand` × 4 anderer Optionen × 2 Seiten = 32; das Orakel ist `rang`.
func TestExpandRangQuery(t *testing.T) {
	type option struct {
		query string
		u     model.URLUrsache
	}
	ex := []option{
		{"$expand=1a", model.URLWert}, {"$expand=" + strings.Join(exNamen("N", 33), ","), model.URLZuViele},
		{"$expand=a%G1", model.URLKodierung}, {"$expand=*", model.URLWertNichtUnterstuetzt},
	}
	andere := []option{
		{"$top=-1", model.URLWert}, {"$foo=1", model.URLOptionName}, {"$select=a,,b", model.URLWert}, {"$filter=a", model.URLFilter},
	}
	p := exParser()
	n := 0
	for _, e := range ex {
		for _, a := range andere {
			for seite := 0; seite < 2; seite++ {
				l, r := e, a
				if seite == 1 {
					l, r = a, e
				}
				soll := rang(nil, []verstoss{{l.u, "query/1"}, {r.u, "query/2"}})
				exPruefe(t, p, l.query+"&"+r.query, exFehl(soll.u, soll.ort))
				n++
			}
		}
	}
	exZaehlt(t, "Rang der Query", n, 32)
}
