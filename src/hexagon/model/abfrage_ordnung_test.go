package model_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// abPermutationen sind drei Reihenfolgen der Indizes 0..n-1: die Folge, ihre Umkehr und eine Drehung.
func abPermutationen(n int) [][]int {
	ident := make([]int, n)
	for i := range ident {
		ident[i] = i
	}
	rev := slices.Clone(ident)
	slices.Reverse(rev)
	dreh := append(slices.Clone(ident[n/2:]), ident[:n/2]...)
	return [][]int{ident, rev, dreh}
}

// --- Zeile 15: Ordnung je Typ ---

// TestOrdnungTypen (Zeile 15): sieben Typen mit Ordnung, je aufsteigende Werte, null und ein doppelter letzter Wert, aufsteigend
// und absteigend in drei Reihenfolgen der Eingabe gegen die ausgeschriebene Liste; absteigend steht null zuletzt, der Gleichstand
// bleibt nach dem Schlüssel aufsteigend.
func TestOrdnungTypen(t *testing.T) {
	typ := abTyp()
	typen := []struct {
		prop  string
		art   model.WertArt
		werte []string
	}{
		{"Name", model.ArtString, []string{"A", "B", "a", "ab", "b"}},
		{"Active", model.ArtBoolean, []string{"false", "true"}},
		{"N", model.ArtInt32, []string{"-5", "-1", "0", "7", "10"}},
		{"Big", model.ArtInt64, []string{"-9007199254740993", "-9007199254740992", "0", "9007199254740992", "9007199254740993"}},
		{"Price", model.ArtDecimal, []string{"-1.5", "-0.5", "0", "0.1", "0.10000000000000001", "2"}},
		{"Day", model.ArtDate, []string{"2019-12-31", "2020-01-31", "2020-02-01", "2021-01-01", "2021-03-04"}},
		{"At", model.ArtDateTimeOffset, []string{"2020-01-01T00:00:00Z", "2020-01-01T00:00:00.5Z", "2020-01-01T00:00:01Z", "2020-01-02T00:00:00Z", "2021-01-01T00:00:00Z"}},
	}
	n := 0
	for _, ty := range typen {
		k := len(ty.werte)
		// Platz i trägt den Schlüssel i+1: 0 ist null, 1..k die Werte, k+1 der doppelte letzte Wert
		wert := func(i int) model.PropertyWert {
			switch {
			case i == 0:
				return abNu(ty.prop)
			case i <= k:
				return abPW(ty.prop, ty.art, ty.werte[i-1])
			}
			return abPW(ty.prop, ty.art, ty.werte[k-1])
		}
		var auf, ab []string
		for i := 0; i < k+2; i++ {
			auf = append(auf, strconv.Itoa(i+1))
		}
		ab = append(ab, strconv.Itoa(k+1), strconv.Itoa(k+2))
		for i := k - 1; i >= 1; i-- {
			ab = append(ab, strconv.Itoa(i+1))
		}
		ab = append(ab, "1")
		for _, p := range abPermutationen(k + 2) {
			es := make([]model.Entity, 0, k+2)
			for _, i := range p {
				es = append(es, abE(i+1, wert(i)))
			}
			for desc, soll := range map[bool][]string{false: auf, true: ab} {
				r := abOk(t, ty.prop, typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd(ty.prop, desc)}}, nil)
				abIdsGleich(t, fmt.Sprintf("%s absteigend %v Reihenfolge %v", ty.prop, desc, p), r.Entities, soll...)
				n++
			}
		}
	}
	if n != 7*2*3 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeile 16: gemischte Schlüssel ---

// TestGemischteSchluessel (Zeile 16): drei Properties verschiedener Typen mit null und Gleichständen in jeder, zwölf Entities; alle
// 6 Folgen der drei Schlüssel × 8 Richtungen gegen eine lexikographische Referenz.
func TestGemischteSchluessel(t *testing.T) {
	typ := abTyp()
	eig := []struct {
		name  string
		art   model.WertArt
		werte [2]string
		cmp   func(a, b string) int
	}{
		{"Name", model.ArtString, [2]string{"a", "b"}, orakelText},
		{"Price", model.ArtDecimal, [2]string{"1", "2.5"}, orakelZahl(t)},
		{"At", model.ArtDateTimeOffset, [2]string{"2020-01-01T00:00:00Z", "2020-01-01T00:00:00.5Z"}, orakelZeit(t)},
	}
	// zustand[i][j]: 0 und 1 sind die Werte, 2 ist null
	zustand := make([][3]int, 12)
	es := make([]model.Entity, 12)
	for i := range es {
		c := (i * 7) % 27
		var props []model.PropertyWert
		for j, e := range eig {
			z := (c / []int{1, 3, 9}[j]) % 3
			zustand[i][j] = z
			if z == 2 {
				props = append(props, abNu(e.name))
			} else {
				props = append(props, abPW(e.name, e.art, e.werte[z]))
			}
		}
		es[i] = abE(i+1, props...)
	}
	n := 0
	for _, folge := range [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		for maske := 0; maske < 8; maske++ {
			var ordnung []model.Sortierung
			for pos, j := range folge {
				ordnung = append(ordnung, abOrd(eig[j].name, maske&(1<<pos) != 0))
			}
			ref := make([]int, 12)
			for i := range ref {
				ref[i] = i
			}
			sort.SliceStable(ref, func(x, y int) bool {
				a, b := ref[x], ref[y]
				for pos, j := range folge {
					za, zb := zustand[a][j], zustand[b][j]
					c := 0
					switch {
					case za == 2 && zb == 2:
					case za == 2:
						c = -1
					case zb == 2:
						c = 1
					default:
						c = eig[j].cmp(eig[j].werte[za], eig[j].werte[zb])
					}
					if maske&(1<<pos) != 0 {
						c = -c
					}
					if c != 0 {
						return c < 0
					}
				}
				return a < b
			})
			var soll []string
			for _, i := range ref {
				soll = append(soll, strconv.Itoa(i+1))
			}
			for _, p := range abPermutationen(12)[:2] {
				ein := make([]model.Entity, 0, 12)
				for _, i := range p {
					ein = append(ein, es[i])
				}
				r := abOk(t, "gemischt", typ, ein, model.Optionen{OrderBy: ordnung}, nil)
				abIdsGleich(t, fmt.Sprintf("Folge %v Richtungen %03b Eingabe %v", folge, maske, p[:3]), r.Entities, soll...)
			}
			n++
		}
	}
	if n != 48 {
		t.Errorf("%d Abfragen", n)
	}
}

// --- Zeile 17: Länge von $orderby ---

// TestOrderByLaenge (Zeile 17): ein Typ mit 33 Properties; $orderby mit 1, 2, 3, 32 und 33 Elementen, wobei erst das letzte
// Element die Gleichstände auflöst (gegen die Folge der Schlüssel).
func TestOrderByLaenge(t *testing.T) {
	typ := model.EntityType{Name: "Ns.T33", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}}
	for j := 1; j <= 33; j++ {
		typ.Properties = append(typ.Properties, model.Property{Name: fmt.Sprintf("P%02d", j), EdmType: "Edm.Int32"})
	}
	for _, k := range []int{1, 2, 3, 32, 33} {
		var es []model.Entity
		for i := 1; i <= 4; i++ {
			var props []model.PropertyWert
			for j := 1; j <= k; j++ {
				w := 0
				if j == k {
					w = 5 - i
				}
				props = append(props, abI(fmt.Sprintf("P%02d", j), w))
			}
			es = append(es, abE(i, props...))
		}
		var ordnung []model.Sortierung
		for j := 1; j <= k; j++ {
			ordnung = append(ordnung, abOrd(fmt.Sprintf("P%02d", j), false))
		}
		r := abOk(t, "OrderBy", typ, es, model.Optionen{OrderBy: ordnung}, nil)
		abIdsGleich(t, fmt.Sprintf("%d Elemente", k), r.Entities, "4", "3", "2", "1")
	}
}

// --- Zeile 18: Gleichstand nach dem Schlüssel ---

// abEntOhneId ist eine Entity mit dem Schlüssel s, ohne die Property Id.
func abEntOhneId(s string, props ...model.PropertyWert) model.Entity {
	return model.Entity{Typ: "Ns.K", Schluessel: s, Properties: props, Uebersprungen: []model.Uebersprungen{}}
}

// TestGleichstandSchluessel (Zeile 18): der Schlüssel ordnet aufsteigend nach seinen Werten, typisiert, in der Folge von Keys;
// ein absteigendes $orderby löst Gleichstände aufsteigend; die Folge der Eingabe ist ohne Bedeutung.
func TestGleichstandSchluessel(t *testing.T) {
	gleichN := func(id int) model.Entity { return abE(id, abI("N", 1)) }
	// Werte statt Text: 1, 2, 10; dazu absteigend mit Gleichständen
	typ := abTyp()
	for _, p := range abPermutationen(3) {
		basis := []model.Entity{gleichN(1), gleichN(2), gleichN(10)}
		es := []model.Entity{basis[p[0]], basis[p[1]], basis[p[2]]}
		for name, opt := range map[string]model.Optionen{"ohne": {}, "N asc": {OrderBy: []model.Sortierung{abOrd("N", false)}}, "N desc": {OrderBy: []model.Sortierung{abOrd("N", true)}}} {
			r := abOk(t, name, typ, es, opt, nil)
			abIdsGleich(t, "Werte "+name+fmt.Sprint(p), r.Entities, "1", "2", "10")
		}
	}
	es := []model.Entity{gleichN(1), abE(4, abI("N", 2)), gleichN(2), gleichN(3)}
	r := abOk(t, "desc", typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd("N", true)}}, nil)
	abIdsGleich(t, "absteigend, Gleichstand aufsteigend", r.Entities, "4", "1", "2", "3")

	// zusammengesetzter Schlüssel: Keys in der Folge Region, Id; Properties in der Folge Id, Region
	komp := model.EntityType{Name: "Ns.K", Keys: []string{"Region", "Id"},
		Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Region", EdmType: "Edm.String"}}}
	paar := func(region string, id int) model.Entity {
		return abEntOhneId("z"+region+strconv.Itoa(10-id), abI("Id", id), abS("Region", region))
	}
	for _, p := range abPermutationen(4) {
		basis := []model.Entity{paar("A", 2), paar("B", 1), paar("A", 1), paar("B", 2)}
		ein := []model.Entity{basis[p[0]], basis[p[1]], basis[p[2]], basis[p[3]]}
		r := abOk(t, "zusammengesetzt", komp, ein, model.Optionen{}, nil)
		var got []string
		for _, e := range r.Entities {
			got = append(got, abNamen([]model.Entity{e}, "Region")[0]+abNamen([]model.Entity{e}, "Id")[0])
		}
		if !slices.Equal(got, []string{"A1", "A2", "B1", "B2"}) {
			t.Errorf("zusammengesetzt %v: %v", p, got)
		}
	}
	// Text: B vor a vor b
	text := model.EntityType{Name: "Ns.K", Keys: []string{"Region"}, Properties: []model.Property{{Name: "Region", EdmType: "Edm.String"}}}
	r = abOk(t, "Text", text, []model.Entity{abEntOhneId("1", abS("Region", "b")), abEntOhneId("2", abS("Region", "a")), abEntOhneId("3", abS("Region", "B"))}, model.Optionen{}, nil)
	abIdsGleich(t, "Text B a b", r.Entities, "3", "2", "1")
	// Guid: nach dem Text der Guid, nicht nach dem Schlüssel
	guid := model.EntityType{Name: "Ns.K", Keys: []string{"Ref"}, Properties: []model.Property{{Name: "Ref", EdmType: "Edm.Guid"}}}
	g := func(n int) model.PropertyWert {
		return abPW("Ref", model.ArtGUID, fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	}
	for _, p := range abPermutationen(3) {
		basis := []model.Entity{abEntOhneId("z", g(1)), abEntOhneId("y", g(2)), abEntOhneId("x", g(3))}
		ein := []model.Entity{basis[p[0]], basis[p[1]], basis[p[2]]}
		r = abOk(t, "Guid", guid, ein, model.Optionen{}, nil)
		abIdsGleich(t, "Guid "+fmt.Sprint(p), r.Entities, "z", "y", "x")
	}
}

// TestSchluesselRest: Entities, deren Werte der Key-Properties gleich sind, aber verschiedene Schluessel tragen (eine Eingabe,
// die der Vertrag nicht prüft), ordnet der Text von Schluessel; die Folge hängt nicht von der Eingabe ab.
func TestSchluesselRest(t *testing.T) {
	typ := abTyp()
	a, b := abE(1), abE(1)
	a.Schluessel, b.Schluessel = "b", "a"
	for _, es := range [][]model.Entity{{a, b}, {b, a}} {
		r := abOk(t, "Rest", typ, es, model.Optionen{}, nil)
		abIdsGleich(t, "Rest", r.Entities, "a", "b")
	}
	ohne := model.EntityType{Name: "Ns.K", Properties: []model.Property{{Name: "N", EdmType: "Edm.Int32"}}}
	for _, es := range [][]model.Entity{{abEntOhneId("2"), abEntOhneId("1")}, {abEntOhneId("1"), abEntOhneId("2")}} {
		r := abOk(t, "ohne Keys", ohne, es, model.Optionen{}, nil)
		abIdsGleich(t, "ohne Keys", r.Entities, "1", "2")
	}
}

// --- Zeile 20: Edm.Guid in $orderby ---

// TestOrderByGuid (Zeile 20): $orderby nach Edm.Guid ist 501 mit ordnung_nicht_unterstuetzt: an erster, mittlerer und letzter Stelle
// von drei × Eingabe (leer, 10) × Top (nicht gesetzt, 0) × Richtung; im Filter gelten eq und ne; die sieben anderen Typen ordnen.
func TestOrderByGuid(t *testing.T) {
	typ := abTyp()
	n := 0
	for pos := 0; pos < 3; pos++ {
		namen := []string{"Id", "Name", "At"}
		namen[pos] = "Ref"
		for _, es := range [][]model.Entity{nil, abSharedMit(false)} {
			for _, top := range []model.ZahlOption{{}, abZ(0)} {
				for _, desc := range []bool{false, true} {
					var o []model.Sortierung
					for _, nm := range namen {
						o = append(o, abOrd(nm, desc))
					}
					r, err := model.FuehreAus(typ, es, model.Optionen{OrderBy: o, Top: top}, nil)
					var ae *model.AbfrageFehler
					if !errors.As(err, &ae) || ae.Option != model.AbfrageOrderBy || ae.Ursache != model.AbfrageOrdnungNichtUnterstuetzt ||
						!errors.Is(err, model.ErrURLNichtUnterstuetzt) || ae.Code() != model.CodeE6001 || !reflect.DeepEqual(r, model.Ergebnis{}) {
						t.Errorf("Stelle %d, %d Entities, Top %v: %v", pos, len(es), top, err)
					}
					n++
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Fälle", n)
	}
	guid := lit(model.LiteralGUID, "00000000-0000-0000-0000-000000000097")
	es := abSharedMit(false)
	abIdsGleich(t, "Ref eq", abOk(t, "eq", typ, es, model.Optionen{Filter: abF(vgl("eq", prop("Ref"), guid))}, nil).Entities, "3")
	if r := abOk(t, "ne", typ, es, model.Optionen{Filter: abF(vgl("ne", prop("Ref"), guid))}, nil); len(r.Entities) != 9 {
		t.Errorf("Ref ne: %d", len(r.Entities))
	}
	for _, p := range []string{"Name", "Active", "N", "Big", "Price", "Day", "At"} {
		if !model.HatOrdnung(propTyp(typ, p)) {
			t.Errorf("%s: HatOrdnung falsch", p)
		}
		abOk(t, p, typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd(p, false)}}, nil)
	}
}

func propTyp(typ model.EntityType, name string) string {
	for _, p := range typ.Properties {
		if p.Name == name {
			return p.EdmType
		}
	}
	return ""
}

// --- Zeile 21: Namen in $orderby und $select ---

// TestNamen (Zeile 21): Stelle ($orderby, $select) × Namensart (sechs unbekannte Schreibweisen, Navigation, offener und geschlossener
// Typ, 29 Typen außerhalb der acht, Maske falsch) × Position (erster, mittlerer, letzter von drei): Ursache, Option und Klasse.
// Die Maske antwortet für jeden Namen außer Id und Name falsch, so dass eine Prüfung der Maske vor dem Typ 422 ergäbe.
func TestNamen(t *testing.T) {
	ausserhalb := typenAusserhalb()
	if len(ausserhalb) != 29 {
		t.Fatalf("%d Typnamen", len(ausserhalb))
	}
	geschlossen := abTyp()
	for i, e := range ausserhalb {
		geschlossen.Properties = append(geschlossen.Properties, model.Property{Name: fmt.Sprintf("X%02d", i), EdmType: e})
	}
	offen := geschlossen
	offen.Open = true
	maske := func(n string) bool { return n == "Id" || n == "Name" }
	type art struct {
		name, prop string
		typ        model.EntityType
		ursache    model.AbfrageUrsache
	}
	arten := []art{
		{"Foo", "Foo", geschlossen, model.AbfrageEigenschaftUnbekannt}, {"klein", "id", geschlossen, model.AbfrageEigenschaftUnbekannt},
		{"groß", "ID", geschlossen, model.AbfrageEigenschaftUnbekannt}, {"Leerzeichen vor", " Id", geschlossen, model.AbfrageEigenschaftUnbekannt},
		{"Leerzeichen nach", "Id ", geschlossen, model.AbfrageEigenschaftUnbekannt}, {"leer", "", geschlossen, model.AbfrageEigenschaftUnbekannt},
		{"Navigation", "Nav", geschlossen, model.AbfrageTypNichtUnterstuetzt}, {"offener Typ", "Zed", offen, model.AbfrageTypNichtUnterstuetzt},
		{"geschlossener Typ", "Zed", geschlossen, model.AbfrageEigenschaftUnbekannt}, {"Maske falsch", "Price", geschlossen, model.AbfrageNichtBeobachtet},
	}
	for i, e := range ausserhalb {
		arten = append(arten, art{e, fmt.Sprintf("X%02d", i), geschlossen, model.AbfrageTypNichtUnterstuetzt})
	}
	klasse := map[model.AbfrageUrsache]error{
		model.AbfrageEigenschaftUnbekannt: model.ErrURLUngueltig, model.AbfrageTypNichtUnterstuetzt: model.ErrURLNichtUnterstuetzt, model.AbfrageNichtBeobachtet: model.ErrFilterNichtBeobachtet,
	}
	n := 0
	for _, stelle := range []model.AbfrageOption{model.AbfrageOrderBy, model.AbfrageSelect} {
		for _, a := range arten {
			for pos := 0; pos < 3; pos++ {
				namen := slices.Insert([]string{"Id", "Name"}, pos, a.prop)
				var opt model.Optionen
				if stelle == model.AbfrageOrderBy {
					for _, nm := range namen {
						opt.OrderBy = append(opt.OrderBy, abOrd(nm, false))
					}
				} else {
					opt.Select = abSel(namen...)
				}
				_, err := model.FuehreAus(a.typ, nil, opt, maske)
				var ae *model.AbfrageFehler
				if !errors.As(err, &ae) || ae.Option != stelle || ae.Ursache != a.ursache || !errors.Is(err, klasse[a.ursache]) {
					t.Errorf("%s %s Stelle %d: %v", stelle, a.name, pos, err)
				}
				n++
			}
		}
	}
	if n != 2*39*3 {
		t.Errorf("%d Fälle", n)
	}
	// `*` mit einem unbekannten Namen daneben, und Id als gültiger Name
	_, err := model.FuehreAus(geschlossen, nil, model.Optionen{Select: &model.SelectOption{Alle: true, Namen: []string{"Foo"}}}, maske)
	if abKennung(err) != "$select:eigenschaft_unbekannt" {
		t.Errorf("* mit unbekanntem Namen: %v", err)
	}
	abOk(t, "Id gültig", geschlossen, nil, model.Optionen{Select: abSel("Id"), OrderBy: []model.Sortierung{abOrd("Id", true)}}, maske)
}

// --- Zeile 22: Properties von $orderby in jeder Entity ---

// TestOrderByEigenschaftFehlt (Zeile 22): Entity ohne die Property an erster, mittlerer und letzter Stelle von fünf × Element
// (erstes, mittleres, letztes von drei) × Richtung: 422 mit der Option $orderby; eine Property mit Wert null gilt.
func TestOrderByEigenschaftFehlt(t *testing.T) {
	typ := abTyp()
	namen := []string{"Name", "Price", "Day"}
	voll := func(i int, ohne string) model.Entity {
		props := []model.PropertyWert{abS("Name", "x"), abPW("Price", model.ArtDecimal, "1"), abPW("Day", model.ArtDate, "2020-01-01")}
		props = slices.DeleteFunc(props, func(p model.PropertyWert) bool { return p.Name == ohne })
		return abE(i, props...)
	}
	n := 0
	for _, element := range namen {
		for _, luecke := range []int{1, 3, 5} {
			for _, desc := range []bool{false, true} {
				var es []model.Entity
				for i := 1; i <= 5; i++ {
					if i == luecke {
						es = append(es, voll(i, element))
					} else {
						es = append(es, voll(i, ""))
					}
				}
				var o []model.Sortierung
				for _, nm := range namen {
					o = append(o, abOrd(nm, desc))
				}
				r, err := model.FuehreAus(typ, es, model.Optionen{OrderBy: o}, nil)
				if abKennung(err) != "$orderby:nicht_beobachtet" || !errors.Is(err, model.ErrFilterNichtBeobachtet) || !reflect.DeepEqual(r, model.Ergebnis{}) {
					t.Errorf("%s fehlt in Entity %d: %v", element, luecke, err)
				}
				n++
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Fälle", n)
	}
	es := []model.Entity{abE(1, abNu("Name"), abNu("Price"), abNu("Day")), voll(2, "")}
	abOk(t, "null gilt", typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd("Name", false), abOrd("Price", false), abOrd("Day", false)}}, nil)
}
