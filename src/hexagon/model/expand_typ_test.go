package model_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Fälle der Zeilen 1 bis 9 der Tabelle Zusage · Fänger von PruefeExpand (SPEC-263). Jeder Generator liefert die Menge
// seiner Zeile aus dem Wortlaut der Spezifikation und dem Fixture-Modell; die Erwartung steht durch die Bildung der Eingabe
// fest, nicht durch einen Aufruf der Funktion.

func exvEAkutNamen(namen []string) int { return strings.Count(strings.Join(namen, "|"), exvEAkut()) }

// exvFaelleGueltig (Zeile 1): jede wirksame Navigation je Typ einzeln, dazu alle Navigationen als Liste in aufsteigender und
// in absteigender Folge.
func exvFaelleGueltig(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		for _, n := range t.navs {
			out = append(out, exvFall{"einzeln " + t.name + " " + n, t.typ, exvO(n), "gueltig"})
		}
		if len(t.navs) >= 2 {
			auf := slices.Sorted(slices.Values(t.navs))
			ab := slices.Clone(auf)
			slices.Reverse(ab)
			out = append(out, exvFall{"aufsteigend " + t.name, t.typ, exvO(auf...), "gueltig"},
				exvFall{"absteigend " + t.name, t.typ, exvO(ab...), "gueltig"})
		}
	}
	return out
}

// exvFaelleSchreibweise (Zeile 2): jede Navigation und jede Property je Typ in den Abwandlungen von exvVarianten.
func exvFaelleSchreibweise(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		for _, n := range slices.Concat(t.navs, t.props) {
			for _, v := range exvVarianten(n) {
				out = append(out, exvFall{fmt.Sprintf("%s %q als %q", t.name, n, v), t.typ, exvO(v), exvFehlt(t)})
			}
		}
	}
	return out
}

// exvFaelleProperty (Zeile 3): jede wirksame Property je Typ, offen und nicht offen.
func exvFaelleProperty(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		for _, p := range t.props {
			out = append(out, exvFall{t.name + " " + p, t.typ, exvO(p), string(model.ExpandKeineNavigation)})
		}
	}
	return out
}

// exvFaelleOffen (Zeile 4): in jedem offenen Typ jeder Name, den der Typ in keiner Liste führt, und jede Abwandlung einer
// seiner Navigationen.
func exvFaelleOffen(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		if !t.offen {
			continue
		}
		namen := exvFremd(m, t)
		for _, n := range t.navs {
			namen = append(namen, exvVarianten(n)...)
		}
		for _, n := range namen {
			out = append(out, exvFall{t.name + " " + n, t.typ, exvO(n), string(model.ExpandNavigationNichtUnterstuetzt)})
		}
	}
	return out
}

// exvFaelleUnbekannt (Zeile 5): in jedem nicht offenen Typ, auch im abstrakten, jeder Name, den der Typ in keiner Liste führt.
func exvFaelleUnbekannt(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		if t.offen {
			continue
		}
		for _, n := range exvFremd(m, t) {
			out = append(out, exvFall{fmt.Sprintf("%s abstrakt=%v %s", t.name, t.abstrakt, n), t.typ, exvO(n), string(model.ExpandNavigationUnbekannt)})
		}
	}
	return out
}

// exvSubtyp sagt, ob u über die Basis-Kette von t abstammt.
func exvSubtyp(m []exvT, u exvT, t string) bool {
	for b := u.basis; b != ""; {
		if b == t {
			return true
		}
		i := slices.IndexFunc(m, func(x exvT) bool { return x.name == b })
		b = m[i].basis
	}
	return false
}

// exvFaelleSubtyp (Zeile 6): für jedes Paar (Typ, abgeleiteter Typ) jede Navigation und Property, die nur der abgeleitete Typ
// führt: am Typ fehlt sie, am abgeleiteten Typ steht sie.
func exvFaelleSubtyp(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		for _, u := range m {
			if !exvSubtyp(m, u, t.name) {
				continue
			}
			for _, n := range slices.Concat(u.navs, u.props) {
				if slices.Contains(t.navs, n) || slices.Contains(t.props, n) {
					continue
				}
				soll := "gueltig"
				if slices.Contains(u.props, n) {
					soll = string(model.ExpandKeineNavigation)
				}
				out = append(out, exvFall{fmt.Sprintf("%s: %s von %s", t.name, n, u.name), t.typ, exvO(n), exvFehlt(t)},
					exvFall{fmt.Sprintf("%s: %s", u.name, n), u.typ, exvO(n), soll})
			}
		}
	}
	return out
}

// exvFolgen liefert alle Folgen der Länge l über die drei Klassen g (Navigation), u (kein Name des Typs) und k (Property).
func exvFolgen(l int) []string {
	folgen := []string{""}
	for range l {
		var neu []string
		for _, f := range folgen {
			for _, k := range "guk" {
				neu = append(neu, f+string(k))
			}
		}
		folgen = neu
	}
	return folgen
}

// exvFaelleRang (Zeile 7): in jedem Typ mit mindestens vier Navigationen und vier Properties jede Folge der Länge 1 bis 4
// aus den drei Klassen, die Namen je Stelle verschieden; der erste Verstoß von links entscheidet.
func exvFaelleRang(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		if len(t.navs) < 4 || len(t.props) < 4 {
			continue
		}
		for l := 1; l <= 4; l++ {
			for _, f := range exvFolgen(l) {
				namen, soll := make([]string, l), "gueltig"
				for i, k := range f {
					switch k {
					case 'g':
						namen[i] = t.navs[i]
					case 'u':
						namen[i] = exvUnbekannt(i)
					case 'k':
						namen[i] = t.props[i]
					}
					if soll == "gueltig" && k == 'u' {
						soll = exvFehlt(t)
					}
					if soll == "gueltig" && k == 'k' {
						soll = string(model.ExpandKeineNavigation)
					}
				}
				out = append(out, exvFall{t.name + " " + f, t.typ, exvO(namen...), soll})
			}
		}
	}
	return out
}

// exvFaelleRangLang (Zeile 7): in den Typen mit 32 Navigationen ein Verstoß (fremd oder Property) an jeder Stelle der Liste von
// 32 Elementen, dazu jedes Paar von Stellen (i < j) mit jedem Paar der zwei Klassen; der erste Verstoß von links entscheidet.
func exvFaelleRangLang(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		if len(t.navs) != 32 || len(t.props) < 2 {
			continue
		}
		verstoss := func(k byte, n int) (string, string) {
			if k == 'u' {
				return exvUnbekannt(n), exvFehlt(t)
			}
			return t.props[n], string(model.ExpandKeineNavigation)
		}
		for i := range 32 {
			for _, k := range []byte("uk") {
				namen := slices.Clone(t.navs)
				var soll string
				namen[i], soll = verstoss(k, 0)
				out = append(out, exvFall{fmt.Sprintf("%s: %c an Stelle %d", t.name, k, i), t.typ, exvO(namen...), soll})
				for j := i + 1; j < 32; j++ {
					for _, kb := range []byte("uk") {
						namen[j], _ = verstoss(kb, 1)
						out = append(out, exvFall{fmt.Sprintf("%s: %c an Stelle %d, %c an Stelle %d", t.name, k, i, kb, j), t.typ, exvO(slices.Clone(namen)...), soll})
						namen[j] = t.navs[j]
					}
				}
			}
		}
	}
	return out
}

// exvFaelleDoppeltAbstand (Zeile 8): in Listen der Länge 32 jedes Paar von Stellen (i < j) mit demselben Namen, als Navigation,
// als Property und als fremder Name.
func exvFaelleDoppeltAbstand(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		namen := []string{"Doppelt"}
		if len(t.navs) > 0 {
			namen = append(namen, t.navs[0])
		}
		if len(t.props) > 0 {
			namen = append(namen, t.props[0])
		}
		for _, d := range namen {
			for i := range 32 {
				for j := i + 1; j < 32; j++ {
					liste := make([]string, 32)
					for k := range liste {
						liste[k] = exvFuell(k)
					}
					liste[i], liste[j] = d, d
					out = append(out, exvFall{fmt.Sprintf("%s: %q an Stelle %d und %d", t.name, d, i, j), t.typ, exvO(liste...), "programmfehler"})
				}
			}
		}
	}
	return out
}

func exvFuell(i int) string { return fmt.Sprintf("Fuell%d", i) }

// exvFaelleProgrammfehler (Zeile 8): o nil, Elemente nil und leer, ein leerer Name und `*` an Stelle 0 bis 2, ein doppelter
// Name mit Abstand 0 bis 2 (sechs Stellenpaare in Listen der Länge 4) als Navigation, Property und unbekannter Name.
func exvFaelleProgrammfehler(m []exvT) []exvFall {
	soll := "programmfehler"
	var out []exvFall
	for _, t := range m {
		out = append(out, exvFall{t.name + " o nil", t.typ, nil, soll}, exvFall{t.name + " Elemente nil", t.typ, &model.ExpandOption{}, soll},
			exvFall{t.name + " Elemente leer", t.typ, exvO(), soll})
		for p := range 3 {
			for _, w := range []string{"", "*"} {
				namen := []string{exvFuell(0), exvFuell(1), exvFuell(2)}
				namen[p] = w
				out = append(out, exvFall{fmt.Sprintf("%s %q an Stelle %d", t.name, w, p), t.typ, exvO(namen...), soll})
			}
		}
		doppelt := []string{"Doppelt"}
		if len(t.navs) > 0 {
			doppelt = append(doppelt, t.navs[0])
		}
		if len(t.props) > 0 {
			doppelt = append(doppelt, t.props[0])
		}
		for _, d := range doppelt {
			for _, paar := range [][2]int{{0, 1}, {1, 2}, {2, 3}, {0, 2}, {1, 3}, {0, 3}} {
				namen := []string{exvFuell(0), exvFuell(1), exvFuell(2), exvFuell(3)}
				namen[paar[0]], namen[paar[1]] = d, d
				out = append(out, exvFall{fmt.Sprintf("%s %q an Stelle %v", t.name, d, paar), t.typ, exvO(namen...), soll})
			}
		}
	}
	return out
}

// exvFaelleProgrammfehlerVor (Zeile 9): ein Verstoß (unbekannt, im offenen Typ nicht unterstützt, keine Navigation) vor und
// nach einem leeren Namen, `*` und einem doppelten Namen; der Programmfehler geht vor.
func exvFaelleProgrammfehlerVor(m []exvT) []exvFall {
	var out []exvFall
	for _, t := range m {
		verstoesse := []string{exvUnbekannt(7)}
		if len(t.props) > 0 {
			verstoesse = append(verstoesse, t.props[0])
		}
		arten := []struct {
			name   string
			fehler []string
		}{{"leer", []string{""}}, {"stern", []string{"*"}}, {"doppelt", []string{"Doppelt", "Doppelt"}}}
		for _, art := range arten {
			for _, v := range verstoesse {
				out = append(out, exvFall{fmt.Sprintf("%s %s mit %s davor", t.name, art.name, v), t.typ, exvO(slices.Concat([]string{v}, art.fehler)...), "programmfehler"},
					exvFall{fmt.Sprintf("%s %s mit %s danach", t.name, art.name, v), t.typ, exvO(slices.Concat(art.fehler, []string{v})...), "programmfehler"})
			}
		}
	}
	return out
}

func exvTyp(m []exvT, name string) exvT {
	return m[slices.IndexFunc(m, func(t exvT) bool { return t.name == name })]
}

// TestPruefeExpandGueltig (Zeile 1): jede wirksame Navigation jedes Typs in beiden Folgen der Listen, einzeln und als Liste
// aufsteigend und absteigend, ist gültig; Zählprüfung gegen die Zahl der Member, die die Kette der Basistypen führt.
func TestPruefeExpandGueltig(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleGueltig(m)
	soll := 0
	for _, tt := range m {
		n := exvZaehlen()[tt.name].navs
		soll += n + 2*exvEins(n >= 2)
	}
	exvZaehlt(t, "gültig", len(faelle), soll)
	exvLaufe(t, faelle)
}

// TestPruefeExpandSchreibweise (Zeile 2): jeder Name wird bytegleich gesucht; kleiner, größer, anders geschrieben, mit
// Leerzeichen und in der Form NFD ist er fremd (Fixture aus Bytewerten, AGENTS.md §3.18).
func TestPruefeExpandSchreibweise(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleSchreibweise(m)
	soll := 0
	for _, tt := range m {
		z := exvZaehlen()[tt.name]
		soll += 4*(z.navs+z.props) + exvEAkutNamen(slices.Concat(tt.navs, tt.props))
	}
	exvZaehlt(t, "Schreibweise", len(faelle), soll)
	exvLaufe(t, faelle)
}

// TestPruefeExpandProperty (Zeile 3): jede wirksame Property jedes Typs, auch eine geerbte und eine mit nicht unterstütztem
// Typ, im offenen wie im nicht offenen Typ, ist `keine_navigation`.
func TestPruefeExpandProperty(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleProperty(m)
	soll := 0
	for _, tt := range m {
		soll += exvZaehlen()[tt.name].props
	}
	exvZaehlt(t, "Property", len(faelle), soll)
	exvLaufe(t, faelle)
}

// TestPruefeExpandOffen (Zeile 4): im offenen Typ ist jeder Name außerhalb beider Listen `navigation_nicht_unterstuetzt`,
// auch der Name einer Navigation eines anderen Typs oder des Subtyps und die Abwandlung einer vorhandenen Navigation.
func TestPruefeExpandOffen(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleOffen(m)
	soll := 0
	for _, tt := range m {
		if tt.offen {
			z := exvZaehlen()[tt.name]
			soll += len(exvNamen(m)) - z.navs - z.props + 4*z.navs + exvEAkutNamen(tt.navs)
		}
	}
	exvZaehlt(t, "offen", len(faelle), soll)
	exvLaufe(t, faelle)
}

// TestPruefeExpandUnbekannt (Zeile 5): im nicht offenen Typ, auch im abstrakten, ist jeder Name außerhalb beider Listen
// `navigation_unbekannt`, darunter jeder Name jedes anderen Typs.
func TestPruefeExpandUnbekannt(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleUnbekannt(m)
	soll := 0
	for _, tt := range m {
		if !tt.offen {
			z := exvZaehlen()[tt.name]
			soll += len(exvNamen(m)) - z.navs - z.props
		}
	}
	exvZaehlt(t, "unbekannt", len(faelle), soll)
	exvLaufe(t, faelle)
	if !slices.ContainsFunc(m, func(x exvT) bool { return x.abstrakt && !x.offen }) {
		t.Error("kein abstrakter Typ im Fixture-Modell")
	}
}

// TestPruefeExpandSubtyp (Zeile 6): was nur ein abgeleiteter Typ führt, ist am Typ des Sets fremd und am abgeleiteten Typ
// gültig; Paare bis zur dritten Stufe, Zählprüfung über die Tiefe der Kette.
func TestPruefeExpandSubtyp(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleSubtyp(m)
	zahlen, nach := exvZaehlen(), map[string]exvSpez{}
	for _, s := range exvSpezen() {
		nach[s.name] = s
	}
	soll := 0
	for _, s := range exvSpezen() {
		for a := nach[s.basis]; a.name != ""; a = nach[a.basis] {
			soll += 4 * 2 * (zahlen[s.name].navs - zahlen[a.name].navs + zahlen[s.name].props - zahlen[a.name].props)
		}
	}
	exvZaehlt(t, "Subtyp", len(faelle), soll)
	exvLaufe(t, faelle)
	if !slices.ContainsFunc(exvModell(), func(x exvT) bool { return x.basis != "" && exvTyp(exvModell(), x.basis).basis != "" }) {
		t.Error("keine Kette von drei Stufen")
	}
}

// TestPruefeExpandRang (Zeile 7): der erste Verstoß von links entscheidet, gleich welcher Klasse: jede Folge der Länge 1
// bis 4 aus gültig, fremd und Property (im offenen Typ: nicht unterstützt), dazu die vier ersten Paare der Zusatztabelle
// als benannte Fälle.
func TestPruefeExpandRang(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleRang(m)
	typen := 0
	for _, tt := range m {
		z := exvZaehlen()[tt.name]
		typen += exvEins(z.navs >= 4 && z.props >= 4)
	}
	exvZaehlt(t, "Rang", len(faelle), typen*(3+9+27+81))
	exvLaufe(t, faelle)
	lang, soll := exvFaelleRangLang(m), 0
	for _, tt := range m {
		if n := exvZaehlen()[tt.name]; n.navs == 32 && n.props >= 2 {
			soll += 2*32 + 4*(32*31/2)
		}
	}
	exvZaehlt(t, "Rang in Listen von 32 Elementen", len(lang), soll)
	exvLaufe(t, lang)
	nicht, offen := exvTyp(m, "Ns.Stufe3"), exvTyp(m, "Ns.OffenKind")
	benannt := []exvFall{
		{"unbekannt, Property", nicht.typ, exvO(exvUnbekannt(0), nicht.props[0]), string(model.ExpandNavigationUnbekannt)},
		{"Property, unbekannt", nicht.typ, exvO(nicht.props[0], exvUnbekannt(0)), string(model.ExpandKeineNavigation)},
		{"Name im offenen Typ, Property", offen.typ, exvO(exvUnbekannt(0), offen.props[0]), string(model.ExpandNavigationNichtUnterstuetzt)},
		{"Property, Name im offenen Typ", offen.typ, exvO(offen.props[0], exvUnbekannt(0)), string(model.ExpandKeineNavigation)},
	}
	exvLaufe(t, benannt)
}

// TestPruefeExpandProgrammfehler (Zeile 8): o nil, leere Liste, leerer Name, `*` und ein doppelter Name sind ErrUngueltig und
// kein ExpandFehler, in jedem Typ; `a` und `A` sind zwei Namen, und das Paar ist so wenig ein Programmfehler wie ein
// Paar fremder Namen in anderer Schreibweise.
func TestPruefeExpandProgrammfehler(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleProgrammfehler(m)
	soll := 0
	for _, tt := range m {
		z := exvZaehlen()[tt.name]
		soll += 3 + 6 + 6*(1+exvEins(z.navs > 0)+exvEins(z.props > 0))
	}
	exvZaehlt(t, "Programmfehler", len(faelle), soll)
	exvLaufe(t, faelle)
	abstand, soll := exvFaelleDoppeltAbstand(m), 0
	for _, tt := range m {
		z := exvZaehlen()[tt.name]
		soll += (32 * 31 / 2) * (1 + exvEins(z.navs > 0) + exvEins(z.props > 0))
	}
	exvZaehlt(t, "doppelter Name in Listen von 32 Elementen", len(abstand), soll)
	exvLaufe(t, abstand)

	klein := model.EntityType{Name: "Ns.Klein", Properties: []model.Property{exvP("b", "Edm.String"), exvP("B", "Edm.String")},
		NavigationProperties: []model.NavigationProperty{exvN("a", "Ns.Klein", false), exvN("A", "Ns.Klein", false)}}
	offen := klein
	offen.Open = true
	gegen := []exvFall{
		{"a, A", klein, exvO("a", "A"), "gueltig"},
		{"A, a", klein, exvO("A", "a"), "gueltig"},
		{"b, B", klein, exvO("b", "B"), string(model.ExpandKeineNavigation)},
		{"B, b", klein, exvO("B", "b"), string(model.ExpandKeineNavigation)},
		{"x, X", klein, exvO("x", "X"), string(model.ExpandNavigationUnbekannt)},
		{"x, X offen", offen, exvO("x", "X"), string(model.ExpandNavigationNichtUnterstuetzt)},
		{"a, a1, A", klein, exvO("a", "a1", "A"), string(model.ExpandNavigationUnbekannt)},
	}
	basis := exvTyp(m, "Ns.Basis")
	for _, n := range basis.navs {
		gegen = append(gegen, exvFall{"vorhanden, abgewandelt: " + n, basis.typ, exvO(n, strings.ToLower(n)), string(model.ExpandNavigationUnbekannt)})
	}
	exvLaufe(t, gegen)
}

// TestPruefeExpandProgrammfehlerVor (Zeile 9): ein Programmfehler (leerer Name, `*`, doppelter Name) geht jedem Verstoß vor,
// auch einem früheren; leere Liste und o nil kennen keine Stelle und stehen in Zeile 8.
func TestPruefeExpandProgrammfehlerVor(t *testing.T) {
	m := exvBeide()
	faelle := exvFaelleProgrammfehlerVor(m)
	soll := 0
	for _, tt := range m {
		soll += 3 * 2 * (1 + exvEins(exvZaehlen()[tt.name].props > 0))
	}
	exvZaehlt(t, "Programmfehler vor Verstoß", len(faelle), soll)
	exvLaufe(t, faelle)
}

// TestPruefeExpandOhneBinding (Zeile 13): eine Navigation mit leerem Ziel, mit einem Ziel, das das Modell nicht führt, mit
// einem Ziel des Modells und als Collection ist gültig, einzeln und alle zusammen, im offenen und im nicht offenen Typ.
func TestPruefeExpandOhneBinding(t *testing.T) {
	var navs []model.NavigationProperty
	var namen []string
	for i, ziel := range []string{"", "Ns.NichtImModell", "Ns.Basis", "Basis"} {
		for _, sammlung := range []bool{false, true} {
			n := fmt.Sprintf("Z%dS%v", i, sammlung)
			navs, namen = append(navs, exvN(n, ziel, sammlung)), append(namen, n)
		}
	}
	faelle := 0
	for _, offen := range []bool{false, true} {
		typ := model.EntityType{Name: "Ns.OhneBinding", Open: offen, NavigationProperties: navs}
		alle := []exvFall{{"alle", typ, exvO(namen...), "gueltig"}}
		for _, n := range namen {
			alle = append(alle, exvFall{n, typ, exvO(n), "gueltig"})
		}
		faelle += len(alle)
		exvLaufe(t, alle)
	}
	exvZaehlt(t, "ohne Binding", faelle, 2*(len(navs)+1))
}
