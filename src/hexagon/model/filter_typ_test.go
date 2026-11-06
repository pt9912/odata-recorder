package model_test

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// --- Bausteine ---

func prop(name string) model.FilterAusdruck { return model.FilterEigenschaft{Name: name} }
func lit(art model.LiteralArt, text string) model.FilterAusdruck {
	return model.FilterLiteral{Literal: model.Literal{Art: art, Text: text}}
}
func vgl(op model.VergleichOp, l, r model.FilterAusdruck) model.FilterAusdruck {
	return model.FilterVergleich{Op: op, Links: l, Rechts: r}
}
func verk(op model.VerknuepfungOp, l, r model.FilterAusdruck) model.FilterAusdruck {
	return model.FilterVerknuepfung{Op: op, Links: l, Rechts: r}
}
func nicht(a model.FilterAusdruck) model.FilterAusdruck { return model.FilterNicht{Operand: a} }

var (
	nul      model.FilterAusdruck = model.FilterNull{}
	wahrLit                       = lit(model.LiteralWahrheitswert, "true")
	falschLi                      = lit(model.LiteralWahrheitswert, "false")
	litText                       = lit(model.LiteralZeichenkette, "'x'")
	litZahl                       = lit(model.LiteralZahl, "1")
	litGUID                       = lit(model.LiteralGUID, "00000000-0000-0000-0000-000000000000")
	litDatum                      = lit(model.LiteralDatum, "2020-01-01")
	litZeit                       = lit(model.LiteralZeitpunkt, "2020-01-01T00:00:00Z")
)

// typen hat die acht unterstützten Properties, eine Navigation und, je Aufruf, die Properties mit Typen außerhalb der acht.
func typMit(extra ...model.Property) model.EntityType {
	return model.EntityType{
		Name: "Ns.T",
		Properties: append([]model.Property{
			{Name: "Name", EdmType: "Edm.String"}, {Name: "Active", EdmType: "Edm.Boolean"}, {Name: "Id", EdmType: "Edm.Int32"},
			{Name: "Big", EdmType: "Edm.Int64"}, {Name: "Price", EdmType: "Edm.Decimal"}, {Name: "Ref", EdmType: "Edm.Guid"},
			{Name: "Day", EdmType: "Edm.Date"}, {Name: "At", EdmType: "Edm.DateTimeOffset"},
		}, extra...),
		NavigationProperties: []model.NavigationProperty{{Name: "Nav", Target: "Ns.T"}},
	}
}

// die sieben Eigenschaften, die nicht Edm.Boolean sind, und die fünf Literalarten, die keine Wahrheit sind
func nichtBoolesch() []model.FilterAusdruck {
	return []model.FilterAusdruck{prop("Name"), prop("Id"), prop("Big"), prop("Price"), prop("Ref"), prop("Day"), prop("At"),
		litText, litZahl, litGUID, litDatum, litZeit}
}

func ursacheVon(t *testing.T, err error) model.FilterTypUrsache {
	t.Helper()
	var ft *model.FilterTypFehler
	if !errors.As(err, &ft) {
		t.Errorf("kein FilterTypFehler: %v", err)
		return ""
	}
	return ft.Ursache
}

// erwarte prüft das Ergebnis von PruefeFilter: u == "" heißt nil, sonst genau diese Ursache.
func erwarte(t *testing.T, beschr string, typ model.EntityType, a model.FilterAusdruck, beob func(string) bool, u model.FilterTypUrsache) {
	t.Helper()
	err := model.PruefeFilter(typ, a, beob)
	if u == "" {
		if err != nil {
			t.Errorf("%s: %v, erwartet nil", beschr, err)
		}
		return
	}
	if err == nil {
		t.Errorf("%s: nil, erwartet %s", beschr, u)
		return
	}
	if got := ursacheVon(t, err); got != u {
		t.Errorf("%s: %s, erwartet %s", beschr, got, u)
	}
}

// erwarteProgrammfehler: ErrUngueltig und kein FilterTypFehler.
func erwarteProgrammfehler(t *testing.T, beschr string, typ model.EntityType, a model.FilterAusdruck, beob func(string) bool) {
	t.Helper()
	if err := model.PruefeFilter(typ, a, beob); !istUngueltig(err) {
		t.Errorf("%s: %v, erwartet ErrUngueltig", beschr, err)
	}
}

type stelle struct {
	name string
	baue func(x model.FilterAusdruck) model.FilterAusdruck
}

// stellen sind die acht Stellen, an denen ein Knoten stehen kann; der Rest ist gültig (null oder true).
func stellen() []stelle {
	return []stelle{
		{"Wurzel", func(x model.FilterAusdruck) model.FilterAusdruck { return x }},
		{"Operand von not", func(x model.FilterAusdruck) model.FilterAusdruck { return nicht(x) }},
		{"Links von eq", func(x model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", x, nul) }},
		{"Rechts von eq", func(x model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", nul, x) }},
		{"Links von and", func(x model.FilterAusdruck) model.FilterAusdruck { return verk("and", x, wahrLit) }},
		{"Rechts von and", func(x model.FilterAusdruck) model.FilterAusdruck { return verk("and", wahrLit, x) }},
		{"Links von or", func(x model.FilterAusdruck) model.FilterAusdruck { return verk("or", x, wahrLit) }},
		{"Rechts von or", func(x model.FilterAusdruck) model.FilterAusdruck { return verk("or", wahrLit, x) }},
	}
}

func alleBeobachtet(string) bool { return true }

// --- Zeilen 21 bis 23: Wurzel und boolesche Operanden ---

// TestWurzel (Zeile 21): acht gültige Wurzeln, zwölf ungültige (die sieben anderen Property-Typen, fünf andere Literalarten).
func TestWurzel(t *testing.T) {
	gueltig := []model.FilterAusdruck{vgl("eq", prop("Id"), litZahl), verk("and", wahrLit, wahrLit), verk("or", wahrLit, nul), nicht(wahrLit),
		prop("Active"), wahrLit, falschLi, nul}
	for i, a := range gueltig {
		erwarte(t, fmt.Sprintf("gültige Wurzel %d", i), typMit(), a, nil, "")
	}
	ungueltig := nichtBoolesch()
	for i, a := range ungueltig {
		erwarte(t, fmt.Sprintf("Wurzel %d", i), typMit(), a, nil, model.FilterNichtBoolesch)
	}
	if len(gueltig) != 8 || len(ungueltig) != 12 {
		t.Errorf("%d und %d Wurzeln", len(gueltig), len(ungueltig))
	}
}

// TestBooleschOperand (Zeile 22): an allen fünf Stellen von not, and, or ist jeder der zwölf nicht booleschen Operanden nicht_boolesch.
func TestBooleschOperand(t *testing.T) {
	n := 0
	for _, x := range nichtBoolesch() {
		for _, s := range stellen()[1:2] {
			erwarte(t, s.name, typMit(), s.baue(x), nil, model.FilterNichtBoolesch)
			n++
		}
		for _, s := range stellen()[4:] {
			erwarte(t, s.name, typMit(), s.baue(x), nil, model.FilterNichtBoolesch)
			n++
		}
	}
	if n != 12*5 {
		t.Errorf("%d Fälle", n)
	}
	// `not Price gt 5` liest der Parser als `(not Price) gt 5` (SPEC-165 Nr. 3)
	erwarte(t, "not Price gt 5", typMit(), vgl("gt", nicht(prop("Price")), lit(model.LiteralZahl, "5")), nil, model.FilterNichtBoolesch)
}

// TestBooleschGueltig (Zeile 23): dieselben fünf Stellen mit booleschen Operanden und null sind gültig.
func TestBooleschGueltig(t *testing.T) {
	operanden := []model.FilterAusdruck{prop("Active"), wahrLit, falschLi, vgl("eq", prop("Id"), litZahl), verk("and", wahrLit, wahrLit),
		verk("or", wahrLit, wahrLit), nicht(wahrLit), nul}
	n := 0
	for _, x := range operanden {
		for _, s := range append(stellen()[1:2], stellen()[4:]...) {
			erwarte(t, s.name, typMit(), s.baue(x), nil, "")
			n++
		}
	}
	if n != 8*5 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeilen 24 bis 26: Namen und Typen ---

// TestEigenschaftUnbekannt (Zeile 24): Namen, die keine Property sind, an den acht Stellen; bytegleich gesucht, auch alle
// Randformen der Bytes der Namen der Properties.
func TestEigenschaftUnbekannt(t *testing.T) {
	typ := typMit()
	namen := []string{"Foo", "id", "ID", " Id", "Id ", "", "Id\n", "Id" + string([]byte{0}), "Id" + nbsp(), "\xef\xbb\xbfId"}
	n := 0
	for _, p := range typ.Properties {
		for _, m := range randformen(p.Name) {
			if !hatName(typ, m) {
				namen = append(namen, m)
			}
		}
	}
	for _, name := range namen {
		for _, s := range stellen() {
			erwarte(t, fmt.Sprintf("%q an %s", name, s.name), typ, s.baue(prop(name)), nil, model.FilterEigenschaftUnbekannt)
			n++
		}
	}
	if n != 8*(10+randformenNamen) {
		t.Errorf("%d Fälle", n)
	}
	// der Name selbst ist bekannt, an jeder Stelle (mit Gegenüber, das den Typ erfüllt)
	erwarte(t, "Active", typ, prop("Active"), nil, "")
}

func hatName(t model.EntityType, name string) bool {
	for _, p := range t.Properties {
		if p.Name == name {
			return true
		}
	}
	return name == "Nav"
}

// TestNavigationUndOffen (Zeile 25): Navigation an den acht Stellen ist 501, im offenen Typ jeder unbekannte Name,
// im geschlossenen Typ derselbe Name 400.
func TestNavigationUndOffen(t *testing.T) {
	offen := typMit()
	offen.Open = true
	n := 0
	for _, s := range stellen() {
		erwarte(t, "Navigation "+s.name, typMit(), s.baue(prop("Nav")), nil, model.FilterTypNichtUnterstuetzt)
		erwarte(t, "Navigation im offenen Typ "+s.name, offen, s.baue(prop("Nav")), nil, model.FilterTypNichtUnterstuetzt)
		erwarte(t, "offen "+s.name, offen, s.baue(prop("Dynamisch")), nil, model.FilterTypNichtUnterstuetzt)
		erwarte(t, "offen, leerer Name "+s.name, offen, s.baue(prop("")), nil, model.FilterTypNichtUnterstuetzt)
		erwarte(t, "geschlossen "+s.name, typMit(), s.baue(prop("Dynamisch")), nil, model.FilterEigenschaftUnbekannt)
		n += 5
	}
	// ein Name der Properties gilt im offenen Typ wie im geschlossenen
	erwarte(t, "bekannt im offenen Typ", offen, vgl("eq", prop("Id"), litZahl), nil, "")
	erwarte(t, "Tippfehler im offenen Typ", offen, vgl("eq", prop("id"), litZahl), nil, model.FilterTypNichtUnterstuetzt)
	if n != 40 {
		t.Errorf("%d Fälle", n)
	}
}

func typenAusserhalb() []string {
	namen := []string{"Edm.Binary", "Edm.Duration", "Edm.TimeOfDay", "Edm.Stream", "Edm.Double", "Edm.Single", "Edm.Int16", "Edm.Byte",
		"Edm.SByte", "Edm.Untyped"}
	for _, g := range []string{"Geography", "Geometry"} {
		for _, f := range []string{"", "Point", "LineString", "Polygon", "MultiPoint", "MultiLineString", "MultiPolygon", "Collection"} {
			namen = append(namen, "Edm."+g+f)
		}
	}
	return append(namen, "Collection(Edm.String)", "Ns.Complex", "Ns.Color")
}

// TestTypAusserhalb (Zeile 26): jeder Typ außerhalb der acht ist typ_nicht_unterstuetzt an den acht Stellen, auch gegen null und
// bei einer Maske, die die Property nicht kennt; dazu alle Randformen der Bytes der acht Typnamen.
func TestTypAusserhalb(t *testing.T) {
	namen := typenAusserhalb()
	if len(namen) != 29 {
		t.Fatalf("%d Typnamen", len(namen))
	}
	acht := model.UnterstuetzteEDMTypen()
	for _, e := range acht {
		for _, m := range randformen(e) {
			if !slicesContains(acht, m) {
				namen = append(namen, m)
			}
		}
	}
	nichts := func(string) bool { return false }
	n := 0
	for _, typName := range namen {
		typ := typMit(model.Property{Name: "P", EdmType: typName})
		for _, s := range stellen() {
			erwarte(t, typName+" "+s.name, typ, s.baue(prop("P")), nil, model.FilterTypNichtUnterstuetzt)
			erwarte(t, typName+" unbeobachtet "+s.name, typ, s.baue(prop("P")), nichts, model.FilterTypNichtUnterstuetzt)
			n += 2
		}
		erwarte(t, typName+" eq null", typ, vgl("eq", prop("P"), nul), nil, model.FilterTypNichtUnterstuetzt)
	}
	if n != 16*(29+randformenTypen) {
		t.Errorf("%d Fälle", n)
	}
}

func slicesContains(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

// --- Zeilen 27 bis 31: Verträglichkeit, null, Literale, Bereich, Guid ---

type operand struct {
	name   string
	klasse string
	baum   model.FilterAusdruck
}

func operanden() []operand {
	return []operand{
		{"Id", "zahl", prop("Id")}, {"Name", "text", prop("Name")}, {"Active", "wahrheit", prop("Active")}, {"Big", "zahl", prop("Big")},
		{"Price", "zahl", prop("Price")}, {"Ref", "guid", prop("Ref")}, {"Day", "datum", prop("Day")}, {"At", "zeit", prop("At")},
		{"'x'", "text", litText}, {"true", "wahrheit", wahrLit}, {"1", "zahl", litZahl}, {"guid", "guid", litGUID},
		{"date", "datum", litDatum}, {"zeit", "zeit", litZeit}, {"null", "null", nul},
	}
}

// TestVertraeglichkeit (Zeile 27): alle 225 geordneten Paare aus 8 Properties, 6 Literalarten und null mit eq; verträglich
// ist, was derselben Klasse angehört oder null ist (Int32, Int64, Decimal bilden eine Klasse).
func TestVertraeglichkeit(t *testing.T) {
	ops := operanden()
	n, unvertraeglich := 0, 0
	for _, l := range ops {
		for _, r := range ops {
			soll := model.FilterTypUrsache("")
			if l.klasse != r.klasse && l.klasse != "null" && r.klasse != "null" {
				soll = model.FilterTypUnvertraeglich
				unvertraeglich++
			}
			erwarte(t, l.name+" eq "+r.name, typMit(), vgl("eq", l.baum, r.baum), nil, soll)
			n++
		}
	}
	if n != 225 || unvertraeglich != 160 {
		t.Errorf("%d Paare, %d unverträglich", n, unvertraeglich)
	}
	erwarte(t, "Id eq 'x'", typMit(), vgl("eq", prop("Id"), litText), nil, model.FilterTypUnvertraeglich)
	erwarte(t, "Day eq Zeitpunkt", typMit(), vgl("eq", prop("Day"), litZeit), nil, model.FilterTypUnvertraeglich)
	erwarte(t, "At eq Datum", typMit(), vgl("eq", prop("At"), litDatum), nil, model.FilterTypUnvertraeglich)
	erwarte(t, "Name eq Guid", typMit(), vgl("eq", prop("Name"), litGUID), nil, model.FilterTypUnvertraeglich)
	erwarte(t, "Ref eq 'x'", typMit(), vgl("eq", prop("Ref"), litText), nil, model.FilterTypUnvertraeglich)
}

var ordnungsOps = []model.VergleichOp{"gt", "ge", "lt", "le"}

// TestNullGegenTyp (Zeile 28): null gegen die acht Property-Typen, beide Seiten, sechs Operatoren; auch gegen eine Property,
// die nicht Nullable ist; Guid mit Ordnungsoperator ist operator_nicht_unterstuetzt.
func TestNullGegenTyp(t *testing.T) {
	n := 0
	for _, o := range operanden()[:8] {
		for _, op := range alleOps {
			soll := model.FilterTypUrsache("")
			if o.klasse == "guid" && op != "eq" && op != "ne" {
				soll = model.FilterOperatorNichtUnterstuetzt
			}
			erwarte(t, o.name+" "+string(op)+" null", typMit(), vgl(op, o.baum, nul), nil, soll)
			erwarte(t, "null "+string(op)+" "+o.name, typMit(), vgl(op, nul, o.baum), nil, soll)
			n += 2
		}
	}
	for _, op := range alleOps {
		erwarte(t, "null "+string(op)+" null", typMit(), vgl(op, nul, nul), nil, "")
		n++
	}
	typ := typMit()
	for i := range typ.Properties {
		typ.Properties[i].Nullable = i%2 == 0 // gemischt: Nullable gilt nicht
	}
	for _, o := range operanden()[:8] {
		erwarte(t, o.name+" eq null, Nullable gemischt", typ, vgl("eq", o.baum, nul), nil, "")
		n++
	}
	if n != 8*6*2+6+8 {
		t.Errorf("%d Fälle", n)
	}
}

// TestLiteralGegenLiteral (Zeile 29): 6 mal 6 Literalarten mit eq (gültig genau bei gleicher Klasse), ohne Bereichsprüfung.
func TestLiteralGegenLiteral(t *testing.T) {
	arten := []operand{operanden()[8], operanden()[9], operanden()[10], operanden()[11], operanden()[12], operanden()[13]}
	n := 0
	for _, l := range arten {
		for _, r := range arten {
			soll := model.FilterTypUrsache("")
			if l.klasse != r.klasse {
				soll = model.FilterTypUnvertraeglich
			}
			erwarte(t, l.name+" eq "+r.name, typMit(), vgl("eq", l.baum, r.baum), nil, soll)
			n++
		}
	}
	if n != 36 {
		t.Errorf("%d Fälle", n)
	}
	gross := lit(model.LiteralZahl, "99999999999999999999")
	for _, op := range alleOps {
		erwarte(t, "Zahl groß gegen 1", typMit(), vgl(op, gross, litZahl), nil, "")
		erwarte(t, "1 gegen Zahl groß", typMit(), vgl(op, litZahl, gross), nil, "")
		erwarte(t, "Zahl ohne Form", typMit(), vgl(op, lit(model.LiteralZahl, ""), lit(model.LiteralZahl, "1e3")), nil, "")
	}
	erwarte(t, "Zeichenkette lt", typMit(), vgl("lt", litText, litText), nil, "")
	for _, op := range ordnungsOps {
		erwarte(t, "Guid "+string(op), typMit(), vgl(op, litGUID, litGUID), nil, model.FilterOperatorNichtUnterstuetzt)
	}
	erwarte(t, "Guid eq Guid", typMit(), vgl("eq", litGUID, litGUID), nil, "")
}

// TestBereich (Zeile 30): die Literale der Zahl gegen Edm.Int32 und Edm.Int64, auf beiden Seiten des Vergleichs und in sechs
// Operatoren: gültig genau dann, wenn der Text die ganze Zahl der kanonischen Form ist und im Bereich des Typs liegt (Orakel
// math/big); dazu alle Randformen der Bytes der Grenzen. Edm.Decimal hat keine Grenze.
func TestBereich(t *testing.T) {
	grenzen := map[string][2]string{"Id": {"-2147483648", "2147483647"}, "Big": {"-9223372036854775808", "9223372036854775807"}}
	texte := []string{"2147483647", "2147483648", "-2147483648", "-2147483649", "1.5", "1.0", "0", "1e3", "", "1234567890123456789012345",
		"9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809", "-0", "00", "01", "+1", " 1", "1 ", "1.",
		".5", "-", "0x10", "1" + nbsp(), "-1"}
	n := 0
	for _, seed := range []string{"2147483647", "-2147483648", "0", "9223372036854775807", "-9223372036854775808"} {
		texte = append(texte, randformen(seed)...)
	}
	for eigenschaft, g := range grenzen {
		unten, oben := ratVonText(g[0]), ratVonText(g[1])
		for _, text := range texte {
			soll := model.FilterTypUrsache(model.FilterLiteralBereich)
			if gueltigeGanzzahl(text) {
				v, _ := new(big.Int).SetString(text, 10)
				if v.Cmp(unten) >= 0 && v.Cmp(oben) <= 0 {
					soll = ""
				}
			}
			for _, op := range alleOps {
				l := lit(model.LiteralZahl, text)
				erwarte(t, eigenschaft+" "+string(op)+" "+text, typMit(), vgl(op, prop(eigenschaft), l), nil, soll)
				erwarte(t, text+" "+string(op)+" "+eigenschaft, typMit(), vgl(op, l, prop(eigenschaft)), nil, soll)
				n += 2
			}
		}
	}
	if n != 2*6*2*(len(texte)) || len(texte) != randformenBereich {
		t.Errorf("%d Fälle, %d Texte", n, len(texte))
	}
	for _, text := range texte {
		for _, op := range alleOps {
			l := lit(model.LiteralZahl, text)
			erwarte(t, "Decimal", typMit(), vgl(op, prop("Price"), l), nil, "")
			erwarte(t, "Decimal rechts", typMit(), vgl(op, l, prop("Price")), nil, "")
			erwarte(t, "Zahl gegen null", typMit(), vgl(op, l, nul), nil, "")
			erwarte(t, "Zahl gegen Zahl", typMit(), vgl(op, l, litZahl), nil, "")
		}
	}
	// Int32 gegen Int64-Property ohne Literal ist gültig; das Literal prüft nur gegen die Property unmittelbar daneben
	erwarte(t, "Id eq Big", typMit(), vgl("eq", prop("Id"), prop("Big")), nil, "")
	erwarte(t, "Zahl gegen Vergleich", typMit(), vgl("eq", lit(model.LiteralZahl, "99999999999"), vgl("eq", prop("Id"), litZahl)), nil, model.FilterTypUnvertraeglich)
}

func ratVonText(s string) *big.Int { v, _ := new(big.Int).SetString(s, 10); return v }

// gueltigeGanzzahl ist die kanonische Form einer ganzen Zahl: `0` oder `-?[1-9][0-9]*`, mit ASCII-Ziffern.
func gueltigeGanzzahl(s string) bool {
	if s == "0" {
		return true
	}
	s = strings.TrimPrefix(s, "-")
	if s == "" || s[0] < '1' || s[0] > '9' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// TestGuidOperatoren (Zeile 31): die vier Ordnungsoperatoren auf Guid sind operator_nicht_unterstuetzt, mit Property und
// Literal, auf beiden Seiten, gegen Guid und null; eq und ne sind gültig; `Guid gt 5` ist typ_unvertraeglich.
func TestGuidOperatoren(t *testing.T) {
	guids := []model.FilterAusdruck{prop("Ref"), litGUID}
	gegen := []model.FilterAusdruck{prop("Ref"), litGUID, nul}
	n := 0
	for _, g := range guids {
		for _, h := range gegen {
			for _, op := range alleOps {
				soll := model.FilterTypUrsache("")
				if op != "eq" && op != "ne" {
					soll = model.FilterOperatorNichtUnterstuetzt
				}
				erwarte(t, "Guid links "+string(op), typMit(), vgl(op, g, h), nil, soll)
				erwarte(t, "Guid rechts "+string(op), typMit(), vgl(op, h, g), nil, soll)
				n += 2
			}
		}
	}
	if n != 2*3*6*2 {
		t.Errorf("%d Fälle", n)
	}
	for _, g := range guids {
		for _, op := range alleOps {
			erwarte(t, "Guid gegen 5", typMit(), vgl(op, g, lit(model.LiteralZahl, "5")), nil, model.FilterTypUnvertraeglich)
			erwarte(t, "5 gegen Guid", typMit(), vgl(op, lit(model.LiteralZahl, "5"), g), nil, model.FilterTypUnvertraeglich)
		}
	}
}

// --- Zeilen 32 und 33: die Maske ---

type maskenFake struct {
	gefragt []string
	erlaubt func(string) bool
}

func (m *maskenFake) beobachtet(name string) bool {
	m.gefragt = append(m.gefragt, name)
	return m.erlaubt(name)
}

// TestNichtBeobachtet (Zeile 32): an jeder der acht Stellen ist eine nicht beobachtete Property nicht_beobachtet; die Maske nil,
// die Maske nur dieser Property und die Maske, die alles erlaubt, lassen sie durch; Literal und null fragen die Maske nicht.
func TestNichtBeobachtet(t *testing.T) {
	n := 0
	masken := []struct {
		name string
		f    func(string) bool
		soll model.FilterTypUrsache
	}{
		{"nil", nil, ""},
		{"leer", func(string) bool { return false }, model.FilterNichtBeobachtet},
		{"nur diese", func(s string) bool { return s == "Active" }, ""},
		{"nur andere", func(s string) bool { return s != "Active" }, model.FilterNichtBeobachtet},
		{"alle", alleBeobachtet, ""},
	}
	for _, s := range stellen() {
		for _, m := range masken {
			erwarte(t, s.name+" Maske "+m.name, typMit(), s.baue(prop("Active")), m.f, m.soll)
			n++
		}
	}
	if n != 8*5 {
		t.Errorf("%d Fälle", n)
	}
	// `P eq null` mit unbeobachteter P ist nicht_beobachtet, nicht wahr oder falsch
	erwarte(t, "Active eq null", typMit(), vgl("eq", prop("Active"), nul), func(string) bool { return false }, model.FilterNichtBeobachtet)
	// Literale und null fragen die Maske nicht
	fake := &maskenFake{erlaubt: func(string) bool { return false }}
	erwarte(t, "nur Literale", typMit(), verk("and", vgl("eq", litZahl, litZahl), vgl("eq", nul, nul)), fake.beobachtet, "")
	if len(fake.gefragt) != 0 {
		t.Errorf("Maske gefragt: %v", fake.gefragt)
	}
}

// TestMaskeAufrufe (Zeile 33): die Maske sieht nur Namen von Properties mit unterstütztem Typ, einmal je Vorkommen und in der
// Lesefolge, und nach dem ersten Verstoß nicht mehr.
func TestMaskeAufrufe(t *testing.T) {
	offen := typMit(model.Property{Name: "Bin", EdmType: "Edm.Binary"})
	offen.Open = true
	geschlossen := typMit(model.Property{Name: "Bin", EdmType: "Edm.Binary"})
	for _, c := range []struct {
		name   string
		typ    model.EntityType
		baum   model.FilterAusdruck
		erlaubt func(string) bool
		ursache model.FilterTypUrsache
		gefragt []string
	}{
		{"Links vor Rechts", geschlossen, verk("and", prop("Active"), vgl("gt", prop("Id"), litZahl)), alleBeobachtet, "", []string{"Active", "Id"}},
		{"doppeltes Vorkommen", geschlossen, verk("and", prop("Active"), prop("Active")), alleBeobachtet, "", []string{"Active", "Active"}},
		{"Navigation", geschlossen, vgl("eq", prop("Nav"), nul), alleBeobachtet, model.FilterTypNichtUnterstuetzt, nil},
		{"unbekannter Name", geschlossen, vgl("eq", prop("Foo"), nul), alleBeobachtet, model.FilterEigenschaftUnbekannt, nil},
		{"Name im offenen Typ", offen, vgl("eq", prop("Foo"), nul), alleBeobachtet, model.FilterTypNichtUnterstuetzt, nil},
		{"nicht unterstützter Typ", geschlossen, vgl("eq", prop("Bin"), nul), func(string) bool { return false }, model.FilterTypNichtUnterstuetzt, nil},
		{"Verstoß in der Mitte", geschlossen, verk("and", verk("and", prop("Active"), vgl("eq", prop("Foo"), nul)), vgl("eq", prop("Id"), litZahl)),
			alleBeobachtet, model.FilterEigenschaftUnbekannt, []string{"Active"}},
		{"erster Verstoß ist die Maske", geschlossen, verk("and", prop("Active"), vgl("gt", prop("Id"), litZahl)), func(string) bool { return false },
			model.FilterNichtBeobachtet, []string{"Active"}},
		{"Maske nach Verstoß an Links", geschlossen, verk("and", prop("Id"), prop("Active")), alleBeobachtet, model.FilterNichtBoolesch, []string{"Id"}},
		{"Vergleich: Links vor Rechts", geschlossen, vgl("eq", prop("Name"), prop("Name")), alleBeobachtet, "", []string{"Name", "Name"}},
		{"Vergleich: Verstoß Links", geschlossen, vgl("eq", prop("Name"), prop("Id")), func(s string) bool { return s != "Name" },
			model.FilterNichtBeobachtet, []string{"Name"}},
	} {
		fake := &maskenFake{erlaubt: c.erlaubt}
		erwarte(t, c.name, c.typ, c.baum, fake.beobachtet, c.ursache)
		if !reflect.DeepEqual(fake.gefragt, c.gefragt) {
			t.Errorf("%s: Maske gefragt %v, erwartet %v", c.name, fake.gefragt, c.gefragt)
		}
	}
}

// --- Zeile 34: Reihenfolge der Verstöße ---

func verstoesse() []struct {
	u    model.FilterTypUrsache
	baum model.FilterAusdruck
} {
	return []struct {
		u    model.FilterTypUrsache
		baum model.FilterAusdruck
	}{
		{model.FilterEigenschaftUnbekannt, prop("Foo")},
		{model.FilterNichtBoolesch, nicht(prop("Id"))},
		{model.FilterTypUnvertraeglich, vgl("eq", prop("Id"), litText)},
		{model.FilterLiteralBereich, vgl("eq", prop("Id"), lit(model.LiteralZahl, "2147483648"))},
		{model.FilterTypNichtUnterstuetzt, prop("Bin")},
		{model.FilterOperatorNichtUnterstuetzt, vgl("gt", prop("Ref"), litGUID)},
		{model.FilterNichtBeobachtet, prop("Hidden")},
	}
}

func typVerstoesse() (model.EntityType, func(string) bool) {
	typ := typMit(model.Property{Name: "Bin", EdmType: "Edm.Binary"}, model.Property{Name: "Hidden", EdmType: "Edm.Boolean"})
	return typ, func(s string) bool { return s != "Hidden" }
}

// TestReihenfolgeVerstoesse (Zeile 34): alle 49 geordneten Paare der sieben Ursachen als Links und Rechts einer Verknüpfung und
// eines Vergleichs: der erste in der Lesefolge entscheidet, gleich welcher Klasse; die Kinder vor der Regel des Knotens.
func TestReihenfolgeVerstoesse(t *testing.T) {
	typ, beob := typVerstoesse()
	vs := verstoesse()
	for _, v := range vs {
		erwarte(t, "einzeln "+string(v.u), typ, v.baum, beob, v.u)
		erwarte(t, "einzeln in or "+string(v.u), typ, verk("or", v.baum, wahrLit), beob, v.u)
	}
	n := 0
	for _, a := range vs {
		for _, b := range vs {
			erwarte(t, string(a.u)+" vor "+string(b.u)+" in and", typ, verk("and", a.baum, b.baum), beob, a.u)
			erwarte(t, string(a.u)+" vor "+string(b.u)+" in or", typ, verk("or", a.baum, b.baum), beob, a.u)
			erwarte(t, string(a.u)+" vor "+string(b.u)+" in eq", typ, vgl("eq", a.baum, b.baum), beob, a.u)
			erwarte(t, string(a.u)+" vor "+string(b.u)+" in ne", typ, vgl("ne", a.baum, b.baum), beob, a.u)
			n += 4
		}
	}
	if n != 49*4 {
		t.Errorf("%d Fälle", n)
	}
	// ein Verstoß im Kind kommt vor der Regel des Knotens
	erwarte(t, "Kind vor Vergleich", typ, vgl("eq", prop("Id"), vgl("eq", prop("Foo"), nul)), beob, model.FilterEigenschaftUnbekannt)
	erwarte(t, "Kind vor Vergleich, links", typ, vgl("eq", vgl("eq", prop("Foo"), nul), prop("Id")), beob, model.FilterEigenschaftUnbekannt)
	erwarte(t, "Kind vor Literal", typ, vgl("eq", prop("Hidden"), lit(model.LiteralZahl, "1")), beob, model.FilterNichtBeobachtet)
	erwarte(t, "Kind vor Bereich", typ, vgl("eq", prop("Foo"), lit(model.LiteralZahl, "99999999999")), beob, model.FilterEigenschaftUnbekannt)
	// die Anforderung an Links gilt, bevor Rechts geprüft wird
	for _, op := range []model.VerknuepfungOp{"and", "or"} {
		for _, x := range nichtBoolesch() {
			erwarte(t, "Links nicht boolesch vor Rechts", typ, verk(op, x, prop("Foo")), beob, model.FilterNichtBoolesch)
			erwarte(t, "Links nicht boolesch vor Rechts", typ, verk(op, x, prop("Bin")), beob, model.FilterNichtBoolesch)
			erwarte(t, "Rechts nach Links", typ, verk(op, wahrLit, x), beob, model.FilterNichtBoolesch)
		}
	}
	// Ordnungsoperator auf Guid und Typ: typ_unvertraeglich zuerst; Bereich vor Operator
	erwarte(t, "Guid gt 5", typ, vgl("gt", prop("Ref"), litZahl), beob, model.FilterTypUnvertraeglich)
	erwarte(t, "5 gt Guid", typ, vgl("gt", litZahl, litGUID), beob, model.FilterTypUnvertraeglich)
	// die Wurzel zuletzt: ihre Property prüft sich selbst vor der Regel der Wurzel
	erwarte(t, "Wurzel Name", typ, prop("Foo"), beob, model.FilterEigenschaftUnbekannt)
	erwarte(t, "Wurzel Bin", typ, prop("Bin"), beob, model.FilterTypNichtUnterstuetzt)
	erwarte(t, "Wurzel Id unbeobachtet", typ, prop("Id"), func(string) bool { return false }, model.FilterNichtBeobachtet)
	erwarte(t, "Wurzel Id", typ, prop("Id"), beob, model.FilterNichtBoolesch)
	// Verstoß des Knotens gegen Verstoß im Teilbaum von Rechts bei not: Operand zuerst
	erwarte(t, "not Foo", typ, nicht(prop("Foo")), beob, model.FilterEigenschaftUnbekannt)
}

// --- Zeile 37: Programmfehler ---

type fremderKnoten struct{}

func (fremderKnoten) IstFilterAusdruck() {}

// TestProgrammfehler (Zeile 37): ein Aufruf, den der Leser nie liefert, ist ErrUngueltig und kein FilterTypFehler, an jeder der
// acht Stellen mit gültigem Rest: nil, Kind nil, Knoten als Zeiger (auch typed nil), fremder Typ, unbekannte Art des Literals,
// unbekannter Operator.
func TestProgrammfehler(t *testing.T) {
	gueltigerVgl := vgl("eq", prop("Id"), litZahl)
	fehlerhaft := map[string]model.FilterAusdruck{
		"nil":                nil,
		"Zeiger Vergleich":   &model.FilterVergleich{Op: "eq", Links: prop("Id"), Rechts: litZahl},
		"Zeiger Verknüpfung": &model.FilterVerknuepfung{Op: "and", Links: wahrLit, Rechts: wahrLit},
		"Zeiger Nicht":       &model.FilterNicht{Operand: wahrLit},
		"Zeiger Eigenschaft": &model.FilterEigenschaft{Name: "Active"},
		"Zeiger Literal":     &model.FilterLiteral{Literal: model.Literal{Art: model.LiteralWahrheitswert, Text: "true"}},
		"Zeiger Null":        &model.FilterNull{},
		"nil Vergleich":      (*model.FilterVergleich)(nil),
		"nil Verknüpfung":    (*model.FilterVerknuepfung)(nil),
		"nil Nicht":          (*model.FilterNicht)(nil),
		"nil Eigenschaft":    (*model.FilterEigenschaft)(nil),
		"nil Literal":        (*model.FilterLiteral)(nil),
		"nil Null":           (*model.FilterNull)(nil),
		"fremder Typ":        fremderKnoten{},
		"Literal{}":          model.FilterLiteral{},
		"Literal Art 7":      lit(7, "x"),
		"Literal Art -1":     lit(-1, "x"),
		"Nicht{nil}":         model.FilterNicht{},
		"Links nil (and)":    model.FilterVerknuepfung{Op: "and", Rechts: wahrLit},
		"Rechts nil (and)":   model.FilterVerknuepfung{Op: "and", Links: wahrLit},
		"Links nil (or)":     model.FilterVerknuepfung{Op: "or", Rechts: wahrLit},
		"Rechts nil (or)":    model.FilterVerknuepfung{Op: "or", Links: wahrLit},
		"Links nil (vgl)":    model.FilterVergleich{Op: "eq", Rechts: nul},
		"Rechts nil (vgl)":   model.FilterVergleich{Op: "eq", Links: nul},
		"Vergleich{}":        model.FilterVergleich{},
		"Verknüpfung{}":      model.FilterVerknuepfung{},
	}
	for _, op := range []model.VergleichOp{"", "EQ", "like", "and", "Eq", "eq "} {
		fehlerhaft["Operator "+string(op)] = vgl(op, prop("Id"), litZahl)
	}
	for _, op := range []model.VerknuepfungOp{"", "AND", "xor", "eq", "or "} {
		fehlerhaft["Verknüpfung "+string(op)] = verk(op, wahrLit, wahrLit)
	}
	n := 0
	for name, x := range fehlerhaft {
		for _, s := range stellen() {
			erwarteProgrammfehler(t, name+" an "+s.name, typMit(), s.baue(x), nil)
			n++
		}
	}
	if n != 8*(26+6+5) {
		t.Errorf("%d Fälle", n)
	}
	erwarte(t, "gültiger Vergleich", typMit(), gueltigerVgl, nil, "")
	// der erste Fund der Folge entscheidet, auch zwischen Verstoß und Programmfehler
	erwarte(t, "Verstoß vor nil", typMit(), model.FilterVergleich{Op: "eq", Links: prop("Foo")}, nil, model.FilterEigenschaftUnbekannt)
	erwarteProgrammfehler(t, "nil vor Verstoß", typMit(), model.FilterVergleich{Op: "eq", Rechts: prop("Foo")}, nil)
}

// --- Zeile 38: Tiefe und unveränderter Baum ---

func kette(n int, blatt model.FilterAusdruck) model.FilterAusdruck {
	b := blatt
	for i := 0; i < n; i++ {
		b = verk("or", b, prop("Active"))
	}
	return b
}

func gruppen(n int, blatt model.FilterAusdruck) model.FilterAusdruck {
	b := blatt
	for i := 0; i < n; i++ {
		b = verk("and", prop("Active"), b)
	}
	return b
}

func nichts(n int, blatt model.FilterAusdruck) model.FilterAusdruck {
	b := blatt
	for i := 0; i < n; i++ {
		b = nicht(b)
	}
	return b
}

// TestTiefeUndUnveraendert (Zeile 38): eine Kette von 2000 Ebenen, 32 verschachtelte Gruppen und 32 not werden bewältigt, mit
// und ohne Verstoß im Blatt; der Baum ist danach unverändert (DeepEqual gegen einen zweiten, gleich gebauten Baum).
func TestTiefeUndUnveraendert(t *testing.T) {
	n := 0
	for _, b := range []struct {
		name  string
		baue  func(model.FilterAusdruck) model.FilterAusdruck
		tiefe int
	}{
		{"Kette", func(x model.FilterAusdruck) model.FilterAusdruck { return kette(2000, x) }, 2000},
		{"Gruppen", func(x model.FilterAusdruck) model.FilterAusdruck { return gruppen(32, x) }, 32},
		{"not", func(x model.FilterAusdruck) model.FilterAusdruck { return nichts(32, x) }, 32},
		{"Kette 20000", func(x model.FilterAusdruck) model.FilterAusdruck { return kette(20000, x) }, 20000},
		{"Gruppen 1000", func(x model.FilterAusdruck) model.FilterAusdruck { return gruppen(1000, x) }, 1000},
	} {
		for _, blatt := range []struct {
			name string
			x    model.FilterAusdruck
			u    model.FilterTypUrsache
		}{{"gültig", prop("Active"), ""}, {"unbekannt", prop("Foo"), model.FilterEigenschaftUnbekannt}, {"nicht unterstützt", prop("Bin"), model.FilterTypNichtUnterstuetzt}} {
			typ, _ := typVerstoesse()
			baum, kopie := b.baue(blatt.x), b.baue(blatt.x)
			fake := &maskenFake{erlaubt: alleBeobachtet}
			erwarte(t, b.name+" "+blatt.name, typ, baum, fake.beobachtet, blatt.u)
			if !reflect.DeepEqual(baum, kopie) {
				t.Errorf("%s %s: der Baum hat sich geändert", b.name, blatt.name)
			}
			if blatt.u == "" && len(fake.gefragt) != b.tiefe+1 && b.name != "not" && b.name != "Kette 20000" && b.name != "Gruppen 1000" {
				t.Errorf("%s: Maske %d mal gefragt", b.name, len(fake.gefragt))
			}
			n++
		}
	}
	if n != 15 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeile 20: Nebenläufigkeit ---

// TestGleichzeitig (Zeile 20): acht Goroutinen rufen je 1000 mal Vergleiche, OrdneWerte, Und, Oder, Nicht und PruefeFilter auf
// geteilten Werten, Typ und Baum; jedes Ergebnis gleicht dem des sequentiellen Laufs. Das Datenrennen findet `make test-race`.
func TestGleichzeitig(t *testing.T) {
	typ, beob := typVerstoesse()
	gueltig := verk("and", prop("Active"), vgl("gt", prop("Big"), lit(model.LiteralZahl, "9223372036854775807")))
	kaputt := verk("and", prop("Active"), vgl("eq", prop("Id"), lit(model.LiteralZahl, "2147483648")))
	var faelle []func() string
	zahlen := werte(model.ArtDecimal, "9007199254740993", "9007199254740992", "-1.5", "0.1", "0.10000000000000001", lang("0.", 1021, "0", "1"))
	zeiten := zeitListe()[:6]
	texte := textListe()[:6]
	for _, liste := range [][]model.Wert{zahlen, zeiten, texte} {
		for _, a := range liste {
			for _, b := range liste {
				a, b := a, b
				faelle = append(faelle,
					func() string { w, err := model.Vergleiche(model.VergleichLt, a, b); return fmt.Sprint(w, err) },
					func() string { c, err := model.OrdneWerte(a, b); return fmt.Sprint(c, err) })
			}
		}
	}
	for _, a := range tafelEingaben() {
		for _, b := range tafelEingaben() {
			a, b := a, b
			faelle = append(faelle, func() string { return fmt.Sprint(model.Und(a, b), model.Oder(a, b), model.Nicht(a)) })
		}
	}
	faelle = append(faelle,
		func() string { return fmt.Sprint(model.PruefeFilter(typ, gueltig, beob)) },
		func() string { return fmt.Sprint(model.PruefeFilter(typ, kaputt, beob)) })
	soll := make([]string, len(faelle))
	for i, f := range faelle {
		soll[i] = f()
	}
	const goroutinen, durchlaeufe = 8, 1000
	var wg sync.WaitGroup
	start := make(chan struct{})
	var mu sync.Mutex
	abweichungen := 0
	for g := 0; g < goroutinen; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < durchlaeufe; i++ {
				k := (i*7 + g*13) % len(faelle)
				if got := faelle[k](); got != soll[k] {
					mu.Lock()
					abweichungen++
					mu.Unlock()
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if abweichungen != 0 {
		t.Errorf("%d Abweichungen vom sequentiellen Lauf", abweichungen)
	}
	if len(faelle) != 2*(36*3)+25+2 {
		t.Errorf("%d Fälle", len(faelle))
	}
}

// Zählprüfung der Mengen der Generatoren (Zeilen 24 und 26 und 30): die gemessene Größe.
const (
	randformenNamen   = 904
	randformenTypen   = 2368
	randformenBereich = 1689
)
