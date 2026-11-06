package model_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// --- Zeile 23: Form und Klasse des AbfrageFehler ---

// TestAbfrageFehler (Zeile 23): vier Ursachen × drei Optionen: errors.Is trifft genau die Sentinel der Klasse, Code und Text haben
// das Muster; eine Ursache außerhalb der vier hat die Klasse ungueltig; kein Name der Eingabe steht im Text.
func TestAbfrageFehler(t *testing.T) {
	if len(model.AlleAbfrageUrsachen()) != 4 {
		t.Fatalf("%d Ursachen", len(model.AlleAbfrageUrsachen()))
	}
	sentinels := []error{model.ErrURLUngueltig, model.ErrURLNichtUnterstuetzt, model.ErrFilterNichtBeobachtet}
	klasse := map[model.AbfrageUrsache]int{
		model.AbfrageEigenschaftUnbekannt: 0, model.AbfrageTypNichtUnterstuetzt: 1, model.AbfrageOrdnungNichtUnterstuetzt: 1, model.AbfrageNichtBeobachtet: 2,
	}
	codes := []model.Code{model.CodeE6005, model.CodeE6001, model.CodeE6003}
	praefix := []string{"ODR-E6005: Abfrage ungültig: ", "ODR-E6001: Abfrage nicht unterstützt: ", "ODR-E6003: Abfrage nicht beobachtet: "}
	optionen := []model.AbfrageOption{model.AbfrageFilter, model.AbfrageOrderBy, model.AbfrageSelect}
	if string(optionen[0]) != "$filter" || string(optionen[1]) != "$orderby" || string(optionen[2]) != "$select" {
		t.Errorf("Optionen %v", optionen)
	}
	n := 0
	for _, u := range model.AlleAbfrageUrsachen() {
		k := klasse[u]
		for _, o := range optionen {
			e := &model.AbfrageFehler{Option: o, Ursache: u}
			if e.Code() != codes[k] || e.Error() != praefix[k]+string(u)+" ("+string(o)+")" {
				t.Errorf("%s %s: %v %q", u, o, e.Code(), e.Error())
			}
			for i, s := range sentinels {
				if errors.Is(e, s) != (i == k) || errors.Is(errors.Join(errors.New("x"), e), s) != (i == k) {
					t.Errorf("%s %s: errors.Is(%v)", u, o, s)
				}
				n++
			}
			if errors.Is(e, model.ErrUngueltig) {
				t.Errorf("%s %s: Treffer auf ErrUngueltig", u, o)
			}
		}
	}
	if n != 4*3*3 {
		t.Errorf("%d Fälle", n)
	}
	fremd := &model.AbfrageFehler{Option: model.AbfrageSelect, Ursache: "x"}
	if fremd.Code() != model.CodeE6005 || !errors.Is(fremd, model.ErrURLUngueltig) || errors.Is(fremd, model.ErrURLNichtUnterstuetzt) ||
		errors.Is(fremd, model.ErrFilterNichtBeobachtet) || fremd.Error() != "ODR-E6005: Abfrage ungültig: x ($select)" {
		t.Errorf("Ursache außerhalb der vier: %v %v", fremd.Code(), fremd.Error())
	}
	// aus FuehreAus: genau diese Ursache, kein Teil der Eingabe
	typ := abTyp()
	typ.Properties = append(typ.Properties, model.Property{Name: "GEHEIMBin", EdmType: "Edm.Binary"}, model.Property{Name: "GEHEIMName", EdmType: "Edm.String"})
	beob := func(s string) bool { return s != "GEHEIMName" }
	es := []model.Entity{abE(1)}
	faelle := map[model.AbfrageUrsache]model.Optionen{
		model.AbfrageEigenschaftUnbekannt:     {OrderBy: []model.Sortierung{abOrd("GEHEIMX", false)}},
		model.AbfrageTypNichtUnterstuetzt:     {Select: abSel("GEHEIMBin")},
		model.AbfrageOrdnungNichtUnterstuetzt: {OrderBy: []model.Sortierung{abOrd("Ref", false)}},
		model.AbfrageNichtBeobachtet:          {Select: abSel("GEHEIMName")},
	}
	for u, opt := range faelle {
		_, err := model.FuehreAus(typ, es, opt, beob)
		var ae *model.AbfrageFehler
		if !errors.As(err, &ae) || ae.Ursache != u || strings.Contains(err.Error(), "GEHEIM") {
			t.Errorf("%s: %v", u, err)
		}
	}
}

// --- die elf Verstöße von SPEC-190 Nr. 5 ---

// abVerstoss ist einer der elf Verstöße; Teile von Filter, $orderby und $select lassen sich zu Paaren zusammensetzen.
type abVerstoss struct {
	name   string
	stufe  float64 // Rang in der Stufenfolge; Stufe 5 ist $filter vor $orderby
	kenn   string
	filter model.FilterAusdruck
	order  []model.Sortierung
	sel    []string
	top    bool
}

func abVerstoesse() []abVerstoss {
	o := func(n string) []model.Sortierung { return []model.Sortierung{abOrd(n, false)} }
	return []abVerstoss{
		{name: "Programmfehler", stufe: 1, kenn: "programm", top: true},
		{name: "Fehler des Filters", stufe: 2, kenn: "pruefe:eigenschaft_unbekannt", filter: vgl("eq", prop("Foo"), litZahl())},
		{name: "orderby unbekannt", stufe: 3, kenn: "$orderby:eigenschaft_unbekannt", order: o("Foo")},
		{name: "orderby nicht unterstützt", stufe: 3, kenn: "$orderby:typ_nicht_unterstuetzt", order: o("Nav")},
		{name: "orderby Guid", stufe: 3, kenn: "$orderby:ordnung_nicht_unterstuetzt", order: o("Ref")},
		{name: "orderby nicht beobachtet", stufe: 3, kenn: "$orderby:nicht_beobachtet", order: o("Name")},
		{name: "select unbekannt", stufe: 4, kenn: "$select:eigenschaft_unbekannt", sel: []string{"Foo"}},
		{name: "select nicht unterstützt", stufe: 4, kenn: "$select:typ_nicht_unterstuetzt", sel: []string{"Nav"}},
		{name: "select nicht beobachtet", stufe: 4, kenn: "$select:nicht_beobachtet", sel: []string{"Big"}},
		{name: "Entity ohne Filter-Property", stufe: 5, kenn: "$filter:nicht_beobachtet", filter: vgl("gt", prop("Price"), litZahl())},
		{name: "Entity ohne orderby-Property", stufe: 5.5, kenn: "$orderby:nicht_beobachtet", order: o("Day")},
	}
}

// abPaar setzt die Teile der Verstöße in der Folge der Argumente zusammen.
func abPaar(vs ...abVerstoss) model.Optionen {
	var opt model.Optionen
	var filter model.FilterAusdruck
	var namen []string
	for _, v := range vs {
		switch {
		case v.filter == nil:
		case filter == nil:
			filter = v.filter
		default:
			filter = verk("and", filter, v.filter)
		}
		opt.OrderBy = append(opt.OrderBy, v.order...)
		namen = append(namen, v.sel...)
		if v.top {
			opt.Top = abZ(-1)
		}
	}
	if filter != nil {
		opt.Filter = abF(filter)
	}
	if namen != nil {
		opt.Select = abSel(namen...)
	}
	return opt
}

// abLuecken sind fünf Entities; die dritte kennt Price nicht, die vierte Day nicht.
func abLuecken() []model.Entity {
	es := make([]model.Entity, 0, 5)
	for i := 1; i <= 5; i++ {
		props := []model.PropertyWert{abS("Name", "n"), abPW("Big", model.ArtInt64, "1")}
		if i != 3 {
			props = append(props, abPW("Price", model.ArtDecimal, "5"))
		}
		if i != 4 {
			props = append(props, abPW("Day", model.ArtDate, "2020-01-01"))
		}
		es = append(es, abE(i, props...))
	}
	return es
}

func abMaske(n string) bool { return n != "Name" && n != "Big" }

// TestEntitiesNieNil (Zeile 3): bei leerer Eingabe, $top=0, $skip über das Ende und einem Filter ohne Treffer ist Entities nicht
// nil und leer; bei jedem der elf Verstöße ist das Ergebnis der Nullwert.
func TestEntitiesNieNil(t *testing.T) {
	typ := abTyp()
	filterNichts := abF(vgl("eq", prop("Id"), lit(model.LiteralZahl, "99")))
	for name, c := range map[string]struct {
		es  []model.Entity
		opt model.Optionen
	}{
		"nil":          {nil, model.Optionen{}},
		"leer":         {[]model.Entity{}, model.Optionen{}},
		"Top 0":        {abLuecken(), model.Optionen{Top: abZ(0)}},
		"Skip":         {abLuecken(), model.Optionen{Skip: abZ(5)}},
		"Skip max":     {abLuecken(), model.Optionen{Skip: abZ(abMax)}},
		"Filter":       {abLuecken(), model.Optionen{Filter: filterNichts}},
		"Filter, leer": {nil, model.Optionen{Filter: filterNichts, Select: abSel("Name")}},
	} {
		r := abOk(t, name, typ, c.es, c.opt, nil)
		if r.Entities == nil || len(r.Entities) != 0 {
			t.Errorf("%s: Entities %#v", name, r.Entities)
		}
	}
	for _, v := range abVerstoesse() {
		r, err := model.FuehreAus(typ, abLuecken(), abPaar(v), abMaske)
		if err == nil || !reflect.DeepEqual(r, model.Ergebnis{}) || r.Entities != nil || r.Anzahl != 0 {
			t.Errorf("%s: %#v %v", v.name, r, err)
		}
	}
}

// TestRangfolge (Zeile 25): alle 110 geordneten Paare aus den elf Verstößen; der erste der Stufenfolge entscheidet, gleich welcher
// Klasse, innerhalb von $orderby und $select das erste Element.
func TestRangfolge(t *testing.T) {
	typ := abTyp()
	vs := abVerstoesse()
	n := 0
	for i, a := range vs {
		if _, err := model.FuehreAus(typ, abLuecken(), abPaar(a), abMaske); abKennung(err) != a.kenn {
			t.Errorf("%s allein: %v", a.name, err)
		}
		for j, b := range vs {
			if i == j {
				continue
			}
			soll := a
			if b.stufe < a.stufe {
				soll = b
			}
			_, err := model.FuehreAus(typ, abLuecken(), abPaar(a, b), abMaske)
			if abKennung(err) != soll.kenn {
				t.Errorf("(%s, %s): %v, erwartet %s", a.name, b.name, err, soll.kenn)
			}
			n++
		}
	}
	if n != 110 {
		t.Errorf("%d Paare", n)
	}
	// derselbe Name als Guid und mit falscher Maske: 501, nicht 422
	_, err := model.FuehreAus(typ, nil, model.Optionen{OrderBy: []model.Sortierung{abOrd("Ref", false)}}, func(string) bool { return false })
	if abKennung(err) != "$orderby:ordnung_nicht_unterstuetzt" {
		t.Errorf("Guid mit Maske falsch: %v", err)
	}
	// ein Programmfehler der Stufe 1 geht dem Fehler des Filters voraus, auch wenn der Filter vor dem Fehler steht
	opt := abPaar(vs[1], vs[0])
	if _, err := model.FuehreAus(typ, nil, opt, nil); abKennung(err) != "programm" {
		t.Errorf("Programmfehler nach dem Filter: %v", err)
	}
	// innerhalb von $orderby und $select das erste Element, auch wenn das zweite die frühere Ursache hätte
	_, err = model.FuehreAus(typ, nil, model.Optionen{OrderBy: []model.Sortierung{abOrd("Ref", false), abOrd("Foo", false)}}, nil)
	if abKennung(err) != "$orderby:ordnung_nicht_unterstuetzt" {
		t.Errorf("zwei Namen: %v", err)
	}
}

// TestStatischOhneEntities (Zeile 26): die neun Verstöße der Stufen 1 bis 4 × Eingabe (nil, leer, 1, 5 Entities): derselbe Fehler.
func TestStatischOhneEntities(t *testing.T) {
	typ := abTyp()
	n := 0
	for _, v := range abVerstoesse()[:9] {
		for _, es := range [][]model.Entity{nil, {}, abLuecken()[:1], abLuecken()} {
			_, err := model.FuehreAus(typ, es, abPaar(v), abMaske)
			if abKennung(err) != v.kenn {
				t.Errorf("%s mit %d Entities: %v", v.name, len(es), err)
			}
			n++
		}
	}
	if n != 36 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeile 6: Programmfehler der Stufe 1 ---

// TestAbfrageProgrammfehler (Zeile 6): jeder Programmfehler der Stufe 1 ist ErrUngueltig ohne Fehlerwert, mit leerer und mit nichtleerer
// Eingabe (die Fälle der Entities nur mit nichtleerer), die Fehler der Entities an erster, mittlerer und letzter Stelle.
func TestAbfrageProgrammfehler(t *testing.T) {
	typ := abTyp()
	basis := func() []model.Entity {
		return []model.Entity{abE(1, abS("Name", "a")), abE(2, abS("Name", "b")), abE(3, abS("Name", "c"))}
	}
	type fall struct {
		name string
		opt  model.Optionen
	}
	optionen := []fall{
		{"Top -1", model.Optionen{Top: abZ(-1)}},
		{"Top MinInt32", model.Optionen{Top: abZ(-abMax - 1)}},
		{"Skip -1", model.Optionen{Skip: abZ(-1)}},
		{"Skip MinInt32", model.Optionen{Skip: abZ(-abMax - 1)}},
		{"Ausdruck nil", model.Optionen{Filter: &model.FilterOption{Text: "x"}}},
		{"Select leer", model.Optionen{Select: &model.SelectOption{}}},
		{"Select doppelt 1 2", model.Optionen{Select: abSel("Id", "Id", "Name")}},
		{"Select doppelt 2 3", model.Optionen{Select: abSel("Id", "Name", "Name")}},
		{"Select doppelt 1 3", model.Optionen{Select: abSel("Id", "Name", "Id")}},
		{"Select doppelt mit Alle", model.Optionen{Select: &model.SelectOption{Alle: true, Namen: []string{"Id", "Id"}}}},
		{"OrderBy doppelt", model.Optionen{OrderBy: []model.Sortierung{abOrd("Id", false), abOrd("Name", false), abOrd("Id", true)}}},
	}
	n := 0
	pruefe := func(name string, es []model.Entity, opt model.Optionen) {
		t.Helper()
		r, err := model.FuehreAus(typ, es, opt, nil)
		var ae *model.AbfrageFehler
		var fe *model.FilterTypFehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &ae) || errors.As(err, &fe) || !reflect.DeepEqual(r, model.Ergebnis{}) {
			t.Errorf("%s: %v %#v", name, err, r)
		}
		n++
	}
	for _, f := range optionen {
		pruefe(f.name+" leer", nil, f.opt)
		pruefe(f.name+" leere Liste", []model.Entity{}, f.opt)
		pruefe(f.name, basis(), f.opt)
	}
	for stelle := 0; stelle < 3; stelle++ {
		mit := func(f func(*model.Entity)) []model.Entity {
			es := basis()
			f(&es[stelle])
			return es
		}
		pruefe("Schlüssel leer", mit(func(e *model.Entity) { e.Schluessel = "" }), model.Optionen{})
		pruefe("Key fehlt", mit(func(e *model.Entity) { e.Properties = e.Properties[1:] }), model.Optionen{})
		pruefe("Key null", mit(func(e *model.Entity) { e.Properties[0] = abNu("Id") }), model.Optionen{})
		pruefe("Key redigiert", mit(func(e *model.Entity) { e.Properties[0] = abPW("Id", model.ArtRedigiert, "") }), model.Optionen{})
	}
	for _, paar := range [][2]int{{0, 2}, {0, 1}, {1, 2}} {
		es := basis()
		es[paar[1]].Schluessel = es[paar[0]].Schluessel
		pruefe("Schlüssel doppelt", es, model.Optionen{})
	}
	if n != 11*3+12+3 {
		t.Errorf("%d Fälle", n)
	}
	// zwei Key-Properties: jede muss stehen
	zwei := abTyp()
	zwei.Keys = []string{"Id", "Name"}
	if _, err := model.FuehreAus(zwei, []model.Entity{abE(1)}, model.Optionen{}, nil); !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("zweite Key-Property fehlt: %v", err)
	}
	if _, err := model.FuehreAus(zwei, []model.Entity{abE(1, abS("Name", "a"))}, model.Optionen{}, nil); err != nil {
		t.Errorf("beide Key-Properties: %v", err)
	}
}

// TestAuswertungProgrammfehler (Zeile 7): ein Wert, den kein Leser liefert, ist ErrUngueltig und kein Ergebnis, an jeder Stelle des
// Baums und in der letzten Entity.
func TestAuswertungProgrammfehler(t *testing.T) {
	typ := abTyp()
	str := lit(model.LiteralZeichenkette, "'x'")
	schlecht := []struct {
		name string
		prop string
		wert model.PropertyWert
		ein  func(p model.FilterAusdruck) model.FilterAusdruck
	}{
		{"redigiert", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", p, str) }},
		{"redigiert rechts", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", str, p) }},
		{"Klasse", "Name", abI("Name", 1), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", p, str) }},
		{"Klasse rechts", "Name", abI("Name", 1), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("ne", str, p) }},
		{"unbekannte Art", "Name", abPW("Name", "foo", "x"), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", p, str) }},
		{"Form 01", "N", abPW("N", model.ArtInt32, "01"), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("eq", p, litZahl()) }},
		{"Form 01 rechts", "N", abPW("N", model.ArtInt32, "01"), func(p model.FilterAusdruck) model.FilterAusdruck { return vgl("gt", litZahl(), p) }},
		{"Active Wurzel", "Active", abS("Active", "true"), func(p model.FilterAusdruck) model.FilterAusdruck { return p }},
		{"Active not", "Active", abS("Active", "true"), func(p model.FilterAusdruck) model.FilterAusdruck { return nicht(p) }},
		{"Active and links", "Active", abS("Active", "true"), func(p model.FilterAusdruck) model.FilterAusdruck { return verk("and", p, wahrLit()) }},
		{"Active or rechts", "Active", abS("Active", "true"), func(p model.FilterAusdruck) model.FilterAusdruck { return verk("or", wahrLit(), p) }},
		{"Active Form", "Active", abPW("Active", model.ArtBoolean, "True"), func(p model.FilterAusdruck) model.FilterAusdruck { return nicht(p) }},
		{"Literal True", "Active", abPW("Active", model.ArtBoolean, "true"), func(model.FilterAusdruck) model.FilterAusdruck {
			return vgl("eq", prop("Active"), lit(model.LiteralWahrheitswert, "True"))
		}},
		{"Literal True links", "Active", abPW("Active", model.ArtBoolean, "true"), func(model.FilterAusdruck) model.FilterAusdruck {
			return vgl("eq", lit(model.LiteralWahrheitswert, "True"), prop("Active"))
		}},
		{"Literal True not", "Active", abPW("Active", model.ArtBoolean, "true"), func(model.FilterAusdruck) model.FilterAusdruck {
			return nicht(lit(model.LiteralWahrheitswert, "True"))
		}},
		{"Literal True and", "Active", abPW("Active", model.ArtBoolean, "true"), func(model.FilterAusdruck) model.FilterAusdruck {
			return verk("and", wahrLit(), lit(model.LiteralWahrheitswert, "True"))
		}},
		// der Fehler steckt in einem verschachtelten Vergleich: er geht durch not, Links und Rechts von Vergleich, and und or
		{"verschachtelt not", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck { return nicht(vgl("eq", p, str)) }},
		{"verschachtelt Links", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck {
			return vgl("eq", vgl("eq", p, str), wahrLit())
		}},
		{"verschachtelt Rechts", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck {
			return vgl("eq", wahrLit(), vgl("eq", p, str))
		}},
		{"verschachtelt and", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck {
			return verk("and", vgl("eq", p, str), wahrLit())
		}},
		{"verschachtelt or", "Name", abPW("Name", model.ArtRedigiert, ""), func(p model.FilterAusdruck) model.FilterAusdruck {
			return verk("or", wahrLit(), vgl("eq", p, str))
		}},
		{"Zeichenkette ohne Anführung", "Name", abS("Name", "x"), func(model.FilterAusdruck) model.FilterAusdruck {
			return vgl("eq", prop("Name"), lit(model.LiteralZeichenkette, "x"))
		}},
		{"Zeichenkette ohne Anführung links", "Name", abS("Name", "x"), func(model.FilterAusdruck) model.FilterAusdruck {
			return vgl("eq", lit(model.LiteralZeichenkette, "x"), prop("Name"))
		}},
	}
	for _, s := range schlecht {
		for _, schlechteStelle := range []int{0, 2} {
			es := make([]model.Entity, 0, 3)
			for i := 1; i <= 3; i++ {
				props := []model.PropertyWert{abS("Name", "x"), abI("N", 1), abPW("Active", model.ArtBoolean, "true")}
				if i == schlechteStelle+1 {
					props = slices.DeleteFunc(props, func(p model.PropertyWert) bool { return p.Name == s.prop })
					props = append(props, s.wert)
				}
				es = append(es, abE(i, props...))
			}
			r, err := model.FuehreAus(typ, es, model.Optionen{Filter: abF(s.ein(prop(s.prop)))}, nil)
			var ae *model.AbfrageFehler
			var fe *model.FilterTypFehler
			if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &ae) || errors.As(err, &fe) || !reflect.DeepEqual(r, model.Ergebnis{}) {
				t.Errorf("%s in Entity %d: %v %#v", s.name, schlechteStelle+1, err, r)
			}
		}
	}
}
