package model_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// abEingaben sind die vier Formen der Eingabe: nil, leer, eine, fünf Entities.
func abEingaben() [][]model.Entity {
	return [][]model.Entity{nil, {}, {abE(1, abS("Name", "x"))}, {abE(1), abE(2), abE(3), abE(4), abE(5)}}
}

// --- Zeile 8: PruefeFilter vor der Auswertung ---

// TestPruefeFilterAufruf (Zeile 8): jede der sieben Ursachen der Typprüfung × Eingabe (nil, leer, 1, 5 Entities) kommt unverändert
// als *FilterTypFehler; ohne Filter wird die Maske nicht gerufen.
func TestPruefeFilterAufruf(t *testing.T) {
	typ := abTyp()
	faelle := []struct {
		filter model.FilterAusdruck
		ursach model.FilterTypUrsache
		code   model.Code
	}{
		{vgl("eq", prop("Foo"), litZahl()), model.FilterEigenschaftUnbekannt, model.CodeE6005},
		{nicht(prop("Price")), model.FilterNichtBoolesch, model.CodeE6005},
		{vgl("eq", prop("Id"), litText()), model.FilterTypUnvertraeglich, model.CodeE6005},
		{vgl("eq", prop("Id"), lit(model.LiteralZahl, "1.5")), model.FilterLiteralBereich, model.CodeE6005},
		{vgl("eq", prop("Nav"), nul()), model.FilterTypNichtUnterstuetzt, model.CodeE6001},
		{vgl("gt", prop("Ref"), litGUID()), model.FilterOperatorNichtUnterstuetzt, model.CodeE6001},
		{vgl("eq", prop("Name"), litText()), model.FilterNichtBeobachtet, model.CodeE6003},
	}
	maske := func(n string) bool { return n != "Name" }
	n := 0
	for _, f := range faelle {
		for _, es := range abEingaben() {
			_, err := model.FuehreAus(typ, es, model.Optionen{Filter: abF(f.filter)}, maske)
			fe, ok := err.(*model.FilterTypFehler) // unverändert, nicht verpackt
			if !ok || fe.Ursache != f.ursach || fe.Code() != f.code {
				t.Errorf("%s mit %d Entities: %v", f.ursach, len(es), err)
			}
			n++
		}
	}
	if n != 28 {
		t.Errorf("%d Fälle", n)
	}
	ruf := 0
	zaehlt := func(string) bool { ruf++; return false }
	for _, es := range abEingaben() {
		abOk(t, "ohne Filter", typ, es, model.Optionen{Top: abZ(1), Skip: abZ(0)}, zaehlt)
	}
	if ruf != 0 {
		t.Errorf("Maske ohne Filter %d-mal gerufen", ruf)
	}
	// nil heißt: alle beobachtet
	abOk(t, "Maske nil", typ, []model.Entity{abE(1, abS("Name", "x"))}, model.Optionen{Filter: abF(vgl("eq", prop("Name"), litText()))}, nil)
}

// --- Zeile 9: die Maske wird je Vorkommen einmal gerufen ---

// TestAbfrageMaskeAufrufe (Zeile 9): an $filter, $orderby und $select × sieben Namensarten × drei Stellen unter drei Namen: gerufen wird
// genau für Namen unterstützter Properties, je Vorkommen einmal und nach dem ersten Verstoß nicht mehr.
func TestAbfrageMaskeAufrufe(t *testing.T) {
	arten := []struct {
		name, prop string
		offen      bool
		ruft       func(stelle string) bool
		verstoss   func(stelle string) bool
	}{
		{"unterstützt", "Day", false, func(string) bool { return true }, func(string) bool { return false }},
		{"unbekannt", "Foo", false, func(string) bool { return false }, func(string) bool { return true }},
		{"Navigation", "Nav", false, func(string) bool { return false }, func(string) bool { return true }},
		{"offener Typ", "Zed", true, func(string) bool { return false }, func(string) bool { return true }},
		{"Typ außerhalb", "Bin", false, func(string) bool { return false }, func(string) bool { return true }},
		{"Guid", "Ref", false, func(s string) bool { return s != "$orderby" }, func(s string) bool { return s == "$orderby" }},
		{"Maske falsch", "Price", false, func(string) bool { return true }, func(string) bool { return true }},
	}
	n := 0
	for _, a := range arten {
		typ := abTyp()
		typ.Properties = append(typ.Properties, model.Property{Name: "Bin", EdmType: "Edm.Binary"})
		typ.Open = a.offen
		for _, stelle := range []string{"$filter", "$orderby", "$select", "$select mit *"} {
			for pos := 0; pos < 3; pos++ {
				namen := []string{"Id", "Name", "At"}
				namen[pos] = a.prop
				var rufe []string
				beob := func(s string) bool { rufe = append(rufe, s); return s != "Price" }
				var opt model.Optionen
				switch stelle {
				case "$filter":
					c := func(i int) model.FilterAusdruck { return vgl("eq", prop(namen[i]), nul()) }
					opt.Filter = abF(verk("and", verk("and", c(0), c(1)), c(2)))
				case "$orderby":
					for _, nm := range namen {
						opt.OrderBy = append(opt.OrderBy, abOrd(nm, false))
					}
				default:
					opt.Select = abSel(namen...)
					opt.Select.Alle = stelle == "$select mit *"
				}
				_, _ = model.FuehreAus(typ, nil, opt, beob) // Gegenstand sind die Aufrufe der Maske, nicht der Fehler
				kurz := stelle
				if stelle == "$select mit *" {
					kurz = "$select"
				}
				var soll []string
				for i, nm := range namen {
					if i != pos || a.ruft(kurz) {
						soll = append(soll, nm)
					}
					if i == pos && a.verstoss(kurz) {
						break
					}
				}
				if !slices.Equal(rufe, soll) {
					t.Errorf("%s %s Stelle %d: gerufen %v, erwartet %v", a.name, stelle, pos, rufe, soll)
				}
				n++
			}
		}
	}
	if n != 7*4*3 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeile 10: nur wahr lässt durch ---

// TestFilterDurchlass (Zeile 10): Wurzeln aus Vergleich, Edm.Boolean-Property (true, false, null) und Literalen gegen die
// ausgeschriebene Erwartung.
func TestFilterDurchlass(t *testing.T) {
	typ := abTyp()
	es := []model.Entity{abE(1, abPW("Active", model.ArtBoolean, "true")), abE(2, abPW("Active", model.ArtBoolean, "false")), abE(3, abNu("Active"))}
	a := prop("Active")
	faelle := []struct {
		name  string
		baum  model.FilterAusdruck
		soll  []string
		nlaut string
	}{
		{"Active", a, []string{"1"}, ""},
		{"Active eq true", vgl("eq", a, wahrLit()), []string{"1"}, ""},
		{"Active eq false", vgl("eq", a, falschLi()), []string{"2"}, ""},
		{"Active ne true", vgl("ne", a, wahrLit()), []string{"2", "3"}, ""},
		{"Active ne false", vgl("ne", a, falschLi()), []string{"1", "3"}, ""},
		{"Active eq null", vgl("eq", a, nul()), []string{"3"}, ""},
		{"Active ne null", vgl("ne", a, nul()), []string{"1", "2"}, ""},
		{"not Active", nicht(a), []string{"2"}, ""},
		{"true", wahrLit(), []string{"1", "2", "3"}, ""},
		{"false", falschLi(), nil, ""},
		{"null", nul(), nil, ""},
		{"not null", nicht(nul()), nil, ""},
		{"not false", nicht(falschLi()), []string{"1", "2", "3"}, ""},
		{"null and true", verk("and", nul(), wahrLit()), nil, ""},
		{"null and false", verk("and", nul(), falschLi()), nil, ""},
		{"null or true", verk("or", nul(), wahrLit()), []string{"1", "2", "3"}, ""},
		{"true or null", verk("or", wahrLit(), nul()), []string{"1", "2", "3"}, ""},
		{"null or false", verk("or", nul(), falschLi()), nil, ""},
		{"Active or null", verk("or", a, nul()), []string{"1"}, ""},
		{"Active and null", verk("and", a, nul()), nil, ""},
		{"not (Active and null)", nicht(verk("and", a, nul())), []string{"2"}, ""},
		{"Id gt 1", vgl("gt", prop("Id"), litZahl()), []string{"2", "3"}, ""},
		{"Id le 1", vgl("le", prop("Id"), litZahl()), []string{"1"}, ""},
	}
	for _, f := range faelle {
		r := abOk(t, f.name, typ, es, model.Optionen{Filter: abF(f.baum)}, nil)
		abIdsGleich(t, f.name, r.Entities, f.soll...)
		if r.Anzahl != len(f.soll) {
			t.Errorf("%s: Anzahl %d", f.name, r.Anzahl)
		}
	}
}

// --- Zeile 11: Null-Semantik ---

// TestFilterNull (Zeile 11): je Property (Edm.String, Edm.Int64, Edm.Boolean) drei Entities mit null, mit dem leeren Text, 0 bzw.
// false und mit einem Wert; sechs Operatoren × beide Seiten von null gegen die Tabelle von LH-FA-08.b: bei zwei null sind eq, ge,
// le wahr, bei einem null nur ne.
func TestFilterNull(t *testing.T) {
	typ := abTyp()
	props := []struct {
		name         string
		art          model.WertArt
		leer, anders string
	}{{"Name", model.ArtString, "", "x"}, {"Big", model.ArtInt64, "0", "5"}, {"Active", model.ArtBoolean, "false", "true"}}
	soll := map[model.VergleichOp][]string{"eq": {"1"}, "ne": {"2", "3"}, "gt": nil, "ge": {"1"}, "lt": nil, "le": {"1"}}
	n := 0
	for _, p := range props {
		es := []model.Entity{abE(1, abNu(p.name)), abE(2, abPW(p.name, p.art, p.leer)), abE(3, abPW(p.name, p.art, p.anders))}
		for _, op := range alleOps() {
			for _, links := range []bool{true, false} {
				f := vgl(op, prop(p.name), nul())
				if !links {
					f = vgl(op, nul(), prop(p.name))
				}
				r := abOk(t, p.name, typ, es, model.Optionen{Filter: abF(f)}, nil)
				abIdsGleich(t, fmt.Sprintf("%s %s Property links %v", p.name, op, links), r.Entities, soll[op]...)
				n++
			}
		}
	}
	if n != 36 {
		t.Errorf("%d Fälle", n)
	}
	// P eq null und null eq P auf null selbst
	r := abOk(t, "null eq null", typ, []model.Entity{abE(1)}, model.Optionen{Filter: abF(vgl("ne", nul(), nul()))}, nil)
	abIdsGleich(t, "null ne null", r.Entities)
}

// --- Zeile 12: die sechs Operatoren auf jedem Typ ---

func abHolds(op model.VergleichOp, c int) bool {
	switch op {
	case "eq":
		return c == 0
	case "ne":
		return c != 0
	case "gt":
		return c > 0
	case "ge":
		return c >= 0
	case "lt":
		return c < 0
	}
	return c <= 0
}

// TestFilterOperatoren (Zeile 12): sieben Typen × sechs Operatoren und Edm.Guid mit eq und ne × drei Lagen (Wert kleiner, gleich,
// größer als das Literal) × zwei Seiten (Property links, Literal links): 88 Abfragen über 264 Entities.
func TestFilterOperatoren(t *testing.T) {
	typ := abTyp()
	type lage struct {
		prop string
		lit  model.FilterAusdruck
		c    int
	}
	zeit := lit(model.LiteralZeitpunkt, "2020-01-01T00:00:00.5Z")
	g := func(n int) string { return fmt.Sprintf("00000000-0000-0000-0000-%012d", n) }
	typen := []struct {
		name  string
		art   model.WertArt
		lagen [3]lage
		ops   []model.VergleichOp
	}{
		{"Name", model.ArtString, [3]lage{{"B", lit(model.LiteralZeichenkette, "'a'"), -1}, {"a", lit(model.LiteralZeichenkette, "'a'"), 0}, {"b", lit(model.LiteralZeichenkette, "'a'"), 1}}, alleOps()},
		{"Active", model.ArtBoolean, [3]lage{{"false", wahrLit(), -1}, {"true", wahrLit(), 0}, {"true", falschLi(), 1}}, alleOps()},
		{"N", model.ArtInt32, [3]lage{{"-5", lit(model.LiteralZahl, "-2"), -1}, {"2", lit(model.LiteralZahl, "2"), 0}, {"10", lit(model.LiteralZahl, "7"), 1}}, alleOps()},
		{"Big", model.ArtInt64, [3]lage{{"9007199254740992", lit(model.LiteralZahl, "9007199254740993"), -1}, {"9007199254740993", lit(model.LiteralZahl, "9007199254740993"), 0},
			{"9007199254740993", lit(model.LiteralZahl, "9007199254740992"), 1}}, alleOps()},
		{"Price", model.ArtDecimal, [3]lage{{"0.1", lit(model.LiteralZahl, "0.10000000000000001"), -1}, {"0.5", lit(model.LiteralZahl, "0.5"), 0},
			{"0.10000000000000002", lit(model.LiteralZahl, "0.10000000000000001"), 1}}, alleOps()},
		{"Day", model.ArtDate, [3]lage{{"2019-12-31", lit(model.LiteralDatum, "2020-06-15"), -1}, {"2020-06-15", lit(model.LiteralDatum, "2020-06-15"), 0},
			{"2021-01-01", lit(model.LiteralDatum, "2020-06-15"), 1}}, alleOps()},
		{"At", model.ArtDateTimeOffset, [3]lage{{"2020-01-01T00:00:00Z", zeit, -1}, {"2020-01-01T00:00:00.5Z", zeit, 0}, {"2020-01-01T00:00:01Z", zeit, 1}}, alleOps()},
		{"Ref", model.ArtGUID, [3]lage{{g(1), lit(model.LiteralGUID, g(2)), -1}, {g(2), lit(model.LiteralGUID, g(2)), 0}, {g(3), lit(model.LiteralGUID, g(2)), 1}},
			[]model.VergleichOp{"eq", "ne"}},
	}
	gruppen, laeufe := 0, 0
	for _, ty := range typen {
		for _, op := range ty.ops {
			for _, propLinks := range []bool{true, false} {
				for i, l := range ty.lagen {
					var f model.FilterAusdruck = vgl(op, prop(ty.name), l.lit)
					c := l.c
					if !propLinks {
						f, c = vgl(op, l.lit, prop(ty.name)), -l.c
					}
					es := []model.Entity{abE(i+1, abPW(ty.name, ty.art, l.prop))}
					r := abOk(t, ty.name, typ, es, model.Optionen{Filter: abF(f)}, nil)
					if got := len(r.Entities) == 1; got != abHolds(op, c) {
						t.Errorf("%s %s Lage %d Property links %v: durchgelassen %v", ty.name, op, l.c, propLinks, got)
					}
					laeufe++
				}
				gruppen++
			}
		}
	}
	if gruppen != 88 || laeufe != 264 {
		t.Errorf("%d Gruppen, %d Läufe", gruppen, laeufe)
	}
}

// TestFilterZahlKlassen (Zeile 12): int32, int64 und decimal vergleichen sich untereinander nach dem Wert: sechs geordnete Paare
// von Properties × sechs Operatoren × drei Lagen.
func TestFilterZahlKlassen(t *testing.T) {
	typ := abTyp()
	arten := map[string]model.WertArt{"N": model.ArtInt32, "Big": model.ArtInt64, "Price": model.ArtDecimal}
	lagen := [3][2]string{{"5", "6"}, {"5", "5"}, {"6", "5"}}
	n := 0
	for a, artA := range arten {
		for b, artB := range arten {
			if a == b {
				continue
			}
			for _, op := range alleOps() {
				for i, l := range lagen {
					es := []model.Entity{abE(1, abPW(a, artA, l[0]), abPW(b, artB, l[1]))}
					r := abOk(t, a+" "+string(op)+" "+b, typ, es, model.Optionen{Filter: abF(vgl(op, prop(a), prop(b)))}, nil)
					if got := len(r.Entities) == 1; got != abHolds(op, i-1) {
						t.Errorf("%s %s %s Lage %d: %v", a, op, b, i-1, got)
					}
					n++
				}
			}
		}
	}
	if n != 6*6*3 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeile 13: Wahrheitswerte als Operanden ---

// TestWahrheitAlsOperand (Zeile 13): ein Vergleich als Operand ist boolean, unbekannt ist null.
func TestWahrheitAlsOperand(t *testing.T) {
	typ := abTyp()
	es := []model.Entity{abE(1, abI("N", 1)), abE(2, abI("N", 2)), abE(3, abNu("N"))}
	n1 := vgl("eq", prop("N"), litZahl())
	faelle := []struct {
		name string
		baum model.FilterAusdruck
		soll []string
	}{
		{"(N eq 1) eq true", vgl("eq", n1, wahrLit()), []string{"1"}},
		{"(N eq 1) eq false", vgl("eq", n1, falschLi()), []string{"2", "3"}},
		{"(not null) eq null", vgl("eq", nicht(nul()), nul()), []string{"1", "2", "3"}},
		{"(not null) eq true", vgl("eq", nicht(nul()), wahrLit()), nil},
		{"(not null) eq false", vgl("eq", nicht(nul()), falschLi()), nil},
		{"(not null) ne null", vgl("ne", nicht(nul()), nul()), nil},
		{"(not null) ne true", vgl("ne", nicht(nul()), wahrLit()), []string{"1", "2", "3"}},
		{"(null and true) eq false", vgl("eq", verk("and", nul(), wahrLit()), falschLi()), nil},
		{"(null and true) eq null", vgl("eq", verk("and", nul(), wahrLit()), nul()), []string{"1", "2", "3"}},
		{"(null or true) eq true", vgl("eq", verk("or", nul(), wahrLit()), wahrLit()), []string{"1", "2", "3"}},
		{"true eq (N eq 1)", vgl("eq", wahrLit(), n1), []string{"1"}},
	}
	for _, f := range faelle {
		r := abOk(t, f.name, typ, es, model.Optionen{Filter: abF(f.baum)}, nil)
		abIdsGleich(t, f.name, r.Entities, f.soll...)
	}
}

// --- Zeile 14: jede Property des Filters steht in jeder Entity ---

// TestFilterEigenschaftFehlt (Zeile 14): die Property von $filter an acht Stellen des Baums × Entity ohne sie an jeder der
// fünf Stellen × fünf Verknüpfungen mit einem für jede Entity wahren Gegenüber: immer 422 mit der Option $filter;
// derselbe Filter auf lauter Entities mit der Property (auch mit Wert null) gilt.
func TestFilterEigenschaftFehlt(t *testing.T) {
	typ := abTyp()
	immer := abIdGt("0")
	huellen := []func(x model.FilterAusdruck) model.FilterAusdruck{
		func(x model.FilterAusdruck) model.FilterAusdruck { return x },
		func(x model.FilterAusdruck) model.FilterAusdruck { return verk("or", immer, x) },
		func(x model.FilterAusdruck) model.FilterAusdruck { return verk("or", x, immer) },
		func(x model.FilterAusdruck) model.FilterAusdruck { return verk("and", immer, x) },
		func(x model.FilterAusdruck) model.FilterAusdruck { return verk("and", x, immer) },
	}
	voll := func(luecke int) []model.Entity {
		es := make([]model.Entity, 0, 5)
		for i := 1; i <= 5; i++ {
			a := abPW("Active", model.ArtBoolean, "true")
			if i == 3 {
				a = abNu("Active")
			}
			if i-1 == luecke {
				es = append(es, abE(i))
			} else {
				es = append(es, abE(i, a))
			}
		}
		return es
	}
	n := 0
	for _, s := range stellen() {
		for _, h := range huellen {
			f := h(s.baue(prop("Active")))
			abOk(t, "alle Entities tragen die Property: "+s.name, typ, voll(-1), model.Optionen{Filter: abF(f)}, nil)
			for _, luecke := range []int{0, 1, 2, 3, 4} {
				r, err := model.FuehreAus(typ, voll(luecke), model.Optionen{Filter: abF(f)}, nil)
				if abKennung(err) != "$filter:nicht_beobachtet" || !errors.Is(err, model.ErrFilterNichtBeobachtet) || !reflect.DeepEqual(r, model.Ergebnis{}) {
					t.Errorf("%s, Entity %d ohne Property: %v", s.name, luecke+1, err)
				}
				n++
			}
		}
	}
	if n != 8*5*5 {
		t.Errorf("%d Fälle", n)
	}
}
