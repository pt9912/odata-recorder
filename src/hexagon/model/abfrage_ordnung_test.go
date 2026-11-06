package model_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
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
				abIDsGleich(t, fmt.Sprintf("%s absteigend %v Reihenfolge %v", ty.prop, desc, p), r.Entities, soll...)
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
				abIDsGleich(t, fmt.Sprintf("Folge %v Richtungen %03b Eingabe %v", folge, maske, p[:3]), r.Entities, soll...)
			}
			n++
		}
	}
	if n != 48 {
		t.Errorf("%d Abfragen", n)
	}
}

// --- Zeile 17: Länge von $orderby ---

// TestOrderByLaenge (Zeile 17): ein Typ mit 33 Properties; $orderby mit 1, 2, 3, 32 und 33 Elementen, wobei jedes Element einmal
// allein die Gleichstände auflöst (1 + 2 + 3 + 32 + 33 = 71 Läufe, gegen die Folge der Schlüssel).
func TestOrderByLaenge(t *testing.T) {
	typ := model.EntityType{Name: "Ns.T33", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}}
	for j := 1; j <= 33; j++ {
		typ.Properties = append(typ.Properties, model.Property{Name: fmt.Sprintf("P%02d", j), EdmType: "Edm.Int32"})
	}
	n := 0
	for _, kd := range abLaengeUndStelle() {
		k, dec := kd[0], kd[1]
		var es []model.Entity
		for i := 1; i <= 4; i++ {
			var props []model.PropertyWert
			for j := 1; j <= k; j++ {
				w := 0
				if j == dec {
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
		abIDsGleich(t, fmt.Sprintf("%d Elemente, entscheidend Element %d", k, dec), r.Entities, "4", "3", "2", "1")
		n++
	}
	if n != 1+2+3+32+33 {
		t.Errorf("%d Läufe", n)
	}
}

// abLaengeUndStelle sind die Paare (Länge von $orderby, Stelle des einen Elements, das die Gleichstände auflöst) für die Längen
// 1, 2, 3, 32 und 33 und jede Stelle von 1 bis zur Länge.
func abLaengeUndStelle() [][2]int {
	var r [][2]int
	for _, k := range []int{1, 2, 3, 32, 33} {
		for dec := 1; dec <= k; dec++ {
			r = append(r, [2]int{k, dec})
		}
	}
	return r
}

// --- Zeile 18: Gleichstand nach dem Schlüssel ---

// abEntOhneID ist eine Entity mit dem Schlüssel s, ohne die Property Id.
func abEntOhneID(s string, props ...model.PropertyWert) model.Entity {
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
			abIDsGleich(t, "Werte "+name+fmt.Sprint(p), r.Entities, "1", "2", "10")
		}
	}
	es := []model.Entity{gleichN(1), abE(4, abI("N", 2)), gleichN(2), gleichN(3)}
	r := abOk(t, "desc", typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd("N", true)}}, nil)
	abIDsGleich(t, "absteigend, Gleichstand aufsteigend", r.Entities, "4", "1", "2", "3")

	// zusammengesetzter Schlüssel: Keys in der Folge Region, Id; Properties in der Folge Id, Region
	komp := model.EntityType{Name: "Ns.K", Keys: []string{"Region", "Id"},
		Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Region", EdmType: "Edm.String"}}}
	paar := func(region string, id int) model.Entity {
		return abEntOhneID("z"+region+strconv.Itoa(10-id), abI("Id", id), abS("Region", region))
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
	r = abOk(t, "Text", text, []model.Entity{abEntOhneID("1", abS("Region", "b")), abEntOhneID("2", abS("Region", "a")), abEntOhneID("3", abS("Region", "B"))}, model.Optionen{}, nil)
	abIDsGleich(t, "Text B a b", r.Entities, "3", "2", "1")
	// Guid: nach dem Text der Guid, nicht nach dem Schlüssel
	guid := model.EntityType{Name: "Ns.K", Keys: []string{"Ref"}, Properties: []model.Property{{Name: "Ref", EdmType: "Edm.Guid"}}}
	g := func(n int) model.PropertyWert {
		return abPW("Ref", model.ArtGUID, fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
	}
	for _, p := range abPermutationen(3) {
		basis := []model.Entity{abEntOhneID("z", g(1)), abEntOhneID("y", g(2)), abEntOhneID("x", g(3))}
		ein := []model.Entity{basis[p[0]], basis[p[1]], basis[p[2]]}
		r = abOk(t, "Guid", guid, ein, model.Optionen{}, nil)
		abIDsGleich(t, "Guid "+fmt.Sprint(p), r.Entities, "z", "y", "x")
	}
}

// TestSchluesselRest (Zeile 35): der Rest der Ordnung ist der Text von Schluessel, bytegleich aufsteigend (SPEC-191 Nr. 3 e), gleich
// in welcher Richtung das letzte Element von $orderby ordnet. Menge: 19 Texte (10 vor 2, Länge, Groß und Klein, führende Null, Texte über 8 und über 100 Bytes, die sich erst am Ende unterscheiden,
// Leerraum, ein Zeichen jenseits von ASCII) × 4 Formen (Typ ohne Keys; gleiche Key-Werte; gleiche Key-Werte mit $orderby N
// aufsteigend und absteigend) × 3 Reihenfolgen der Eingabe gegen die Reihenfolge von sort.Strings.
func TestSchluesselRest(t *testing.T) {
	texte := []string{"1", "2", "10", "01", "a", "B", "b", "ab", "A1", "a1", "10 ", " 10", string([]byte{0xc3, 0xa9}),
		"abcdefgh1", "abcdefgh2", "abcdefgh10", strings.Repeat("x", 100) + "1", strings.Repeat("x", 100) + "2", strings.Repeat("x", 100) + "10"}
	soll := slices.Clone(texte)
	sort.Strings(soll)
	ohneKeys := abTyp()
	ohneKeys.Keys = nil
	formen := []struct {
		name string
		typ  model.EntityType
		opt  model.Optionen
	}{
		{"ohne Keys", ohneKeys, model.Optionen{}},
		{"gleiche Werte", abTyp(), model.Optionen{}},
		{"gleiche Werte, N aufsteigend", abTyp(), model.Optionen{OrderBy: []model.Sortierung{abOrd("N", false)}}},
		{"gleiche Werte, N absteigend", abTyp(), model.Optionen{OrderBy: []model.Sortierung{abOrd("N", true)}}},
	}
	n := 0
	for _, f := range formen {
		basis := make([]model.Entity, 0, len(texte))
		for _, s := range texte {
			e := abE(1, abI("N", 1))
			e.Schluessel = s
			basis = append(basis, e)
		}
		for _, p := range abPermutationen(len(texte)) {
			ein := make([]model.Entity, 0, len(texte))
			for _, i := range p {
				ein = append(ein, basis[i])
			}
			r := abOk(t, f.name, f.typ, ein, f.opt, nil)
			abIDsGleich(t, f.name+fmt.Sprint(p[:3]), r.Entities, soll...)
			n++
		}
	}
	if n != 4*3 {
		t.Errorf("%d Läufe", n)
	}
}

// TestSchluesselTypen (Zeile 36): der Schlüssel ordnet nach den Werten seines Typs (SPEC-191 Nr. 3 d), nicht nach dem Text der
// Werte und nicht nach Schluessel. Menge: die 8 Typen mit Schlüsselordnung (Edm.Int32 mit negativen Werten und mehreren Stellen,
// Edm.Int64 über 2^53, Edm.Decimal, Edm.DateTimeOffset mit Bruchteil, Edm.String mit B vor a, Edm.Boolean, Edm.Date, Edm.Guid mit
// Hex-Buchstaben und einer letzten Gruppe gegen die Ordnung) × 3 Reihenfolgen der Eingabe; Schluessel trägt die umgekehrte Folge der Werte.
func TestSchluesselTypen(t *testing.T) {
	typen := []struct {
		edm   string
		art   model.WertArt
		werte []string
	}{
		{"Edm.Int32", model.ArtInt32, []string{"-10", "-2", "0", "2", "10"}},
		{"Edm.Int64", model.ArtInt64, []string{"-9007199254740993", "-9007199254740992", "-1", "0", "9", "10", "9007199254740993"}},
		{"Edm.Decimal", model.ArtDecimal, []string{"-10.5", "-2", "0", "0.5", "2", "10"}},
		{"Edm.DateTimeOffset", model.ArtDateTimeOffset, []string{"2020-01-01T00:00:00Z", "2020-01-01T00:00:00.5Z", "2020-01-01T00:00:01Z"}},
		{"Edm.String", model.ArtString, []string{"B", "a", "ab", "b"}},
		{"Edm.Boolean", model.ArtBoolean, []string{"false", "true"}},
		{"Edm.Date", model.ArtDate, []string{"2019-12-31", "2020-01-31", "2021-01-01"}},
		{"Edm.Guid", model.ArtGUID, []string{
			"00000001-0000-0000-0000-000000000009", "00000002-0000-0000-0000-000000000008", "0000000a-0000-0000-0000-000000000007",
			"0000000b-0000-0000-0000-000000000006", "00000100-0000-0000-0000-000000000005",
		}},
	}
	n := 0
	for _, ty := range typen {
		typ := model.EntityType{Name: "Ns.S", Keys: []string{"K"}, Properties: []model.Property{{Name: "K", EdmType: ty.edm}}}
		k := len(ty.werte)
		var soll []string
		basis := make([]model.Entity, 0, k)
		for i, w := range ty.werte {
			s := fmt.Sprintf("s%02d", 50-i)
			soll = append(soll, s)
			basis = append(basis, abEntOhneID(s, abPW("K", ty.art, w)))
		}
		for _, p := range abPermutationen(k) {
			ein := make([]model.Entity, 0, k)
			for _, i := range p {
				ein = append(ein, basis[i])
			}
			r := abOk(t, ty.edm, typ, ein, model.Optionen{}, nil)
			abIDsGleich(t, ty.edm+fmt.Sprint(p), r.Entities, soll...)
			n++
		}
	}
	if n != 8*3 {
		t.Errorf("%d Läufe", n)
	}
}

// TestSchluesselLaenge (Zeile 36): der Schlüssel aus 1, 2, 3 und 4 Key-Properties, jede Key-Property einmal allein entscheidend
// (1 + 2 + 3 + 4 = 10 Fälle) × 3 Reihenfolgen der Eingabe; Schluessel trägt die umgekehrte Folge.
func TestSchluesselLaenge(t *testing.T) {
	n := 0
	for kn := 1; kn <= 4; kn++ {
		for dec := 1; dec <= kn; dec++ {
			typ := model.EntityType{Name: "Ns.S"}
			var props [4]string
			for j := 1; j <= kn; j++ {
				props[j-1] = fmt.Sprintf("K%d", j)
				typ.Keys = append(typ.Keys, props[j-1])
				typ.Properties = append(typ.Properties, model.Property{Name: props[j-1], EdmType: "Edm.Int32"})
			}
			basis := make([]model.Entity, 0, 4)
			var soll []string
			for i := 1; i <= 4; i++ {
				var ps []model.PropertyWert
				for j := 1; j <= kn; j++ {
					w := 0
					if j == dec {
						w = i
					}
					ps = append(ps, abI(props[j-1], w))
				}
				s := fmt.Sprintf("s%d", 9-i)
				soll = append(soll, s)
				basis = append(basis, abEntOhneID(s, ps...))
			}
			for _, p := range abPermutationen(4) {
				ein := []model.Entity{basis[p[0]], basis[p[1]], basis[p[2]], basis[p[3]]}
				r := abOk(t, "Schlüssel", typ, ein, model.Optionen{}, nil)
				abIDsGleich(t, fmt.Sprintf("%d Keys, entscheidend %d, %v", kn, dec, p), r.Entities, soll...)
				n++
			}
		}
	}
	if n != 10*3 {
		t.Errorf("%d Läufe", n)
	}
}

// abSetze ersetzt die Property gleichen Namens in der Entity.
func abSetze(e *model.Entity, w model.PropertyWert) {
	for i, p := range e.Properties {
		if p.Name == w.Name {
			e.Properties[i] = w
		}
	}
}

// TestNamenlisten (Zeilen 6, 21, 22): die Listen von Namen sind nicht begrenzt (SPEC-190 Nr. 7); jede Stelle jeder Länge von 1, 2, 3, 4, 7,
// 17, 21, 32 und 33 trägt den Verstoß, an vier Orten: unbekannter Name in $select und in $orderby, Entity ohne die Property in Stufe 5
// (zweite Entity), doppelter Name in $select und in $orderby (zweites Vorkommen an jeder Stelle ab der zweiten, gleich dem ersten und
// gleich dem Vorgänger). Menge: Σ Länge = 120 Stellen × 3 + 2 Orte × 2 Formen × Σ (Länge − 1) = 111 → 804 Fälle.
func TestNamenlisten(t *testing.T) {
	typ := model.EntityType{Name: "Ns.L", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}}
	var alle []string
	for j := 1; j <= 33; j++ {
		nm := fmt.Sprintf("N%02d", j)
		alle = append(alle, nm)
		typ.Properties = append(typ.Properties, model.Property{Name: nm, EdmType: "Edm.Int32"})
	}
	opt := func(stelle model.AbfrageOption, namen []string) model.Optionen {
		if stelle == model.AbfrageSelect {
			return model.Optionen{Select: abSel(namen...)}
		}
		var o model.Optionen
		for _, nm := range namen {
			o.OrderBy = append(o.OrderBy, abOrd(nm, false))
		}
		return o
	}
	stellen := []model.AbfrageOption{model.AbfrageSelect, model.AbfrageOrderBy}
	n := 0
	for _, laenge := range []int{1, 2, 3, 4, 7, 17, 21, 32, 33} {
		for p := 1; p <= laenge; p++ {
			for _, st := range stellen {
				namen := slices.Clone(alle[:laenge])
				namen[p-1] = "Foo"
				_, err := model.FuehreAus(typ, nil, opt(st, namen), nil)
				if abKennung(err) != string(st)+":eigenschaft_unbekannt" {
					t.Errorf("%s, Länge %d, Stelle %d: %v", st, laenge, p, err)
				}
				n++
			}
			// Stufe 5: die zweite Entity kennt die Property an Stelle p nicht
			es := []model.Entity{abE(1), abE(2)}
			for _, nm := range alle[:laenge] {
				es[0].Properties = append(es[0].Properties, abI(nm, 1))
				if nm != alle[p-1] {
					es[1].Properties = append(es[1].Properties, abI(nm, 2))
				}
			}
			r, err := model.FuehreAus(typ, es, opt(model.AbfrageOrderBy, alle[:laenge]), nil)
			if abKennung(err) != "$orderby:nicht_beobachtet" || !reflect.DeepEqual(r, model.Ergebnis{}) {
				t.Errorf("Stufe 5, Länge %d, Stelle %d: %v", laenge, p, err)
			}
			n++
			if p < 2 {
				continue
			}
			for _, st := range stellen {
				for _, vorlage := range []int{0, p - 2} {
					namen := slices.Clone(alle[:laenge])
					namen[p-1] = namen[vorlage]
					_, err := model.FuehreAus(typ, nil, opt(st, namen), nil)
					if !errors.Is(err, model.ErrUngueltig) || abKennung(err) != "programm" {
						t.Errorf("doppelt %s, Länge %d, Stelle %d, gleich Stelle %d: %v", st, laenge, p, vorlage+1, err)
					}
					n++
				}
			}
		}
	}
	if n != 120*3+2*2*111 {
		t.Errorf("%d Fälle", n)
	}
	// ohne Verstoß gelten alle 33 Namen in beiden Listen
	for _, st := range stellen {
		abOk(t, "33 Namen", typ, []model.Entity{abE(1, abI("N01", 1))}, opt(st, alle[:1]), nil)
		abOk(t, "33 Namen", typ, nil, opt(st, alle), nil)
	}
}

// TestOrdnungProgrammfehler (Zeile 37): ein Wert, den kein Leser liefert, in einer Spalte der Ordnung ist ErrUngueltig und kein
// Ergebnis (SPEC-190 Nr. 3 g, Nr. 5 Stufe 6). Menge: 5 Formen des Werts (redigiert, unbekannte Art, Form 01, andere Klasse,
// Guid unter Zahlen) × 4 Spalten ($orderby P1, P2, P3; zweite Key-Property) × 5 Positionen der Entity = 100, dazu P2 mit Guid in
// allen Entities (Ordnung auf Guid).
func TestOrdnungProgrammfehler(t *testing.T) {
	typ := model.EntityType{Name: "Ns.P", Keys: []string{"Id", "K2"}}
	for _, p := range []string{"Id", "K2", "P1", "P2", "P3"} {
		typ.Properties = append(typ.Properties, model.Property{Name: p, EdmType: "Edm.Int32"})
	}
	basisN := func(n int) []model.Entity {
		es := make([]model.Entity, 0, n)
		for i := 1; i <= n; i++ {
			es = append(es, abE(i, abI("K2", i), abI("P1", i), abI("P2", i), abI("P3", i)))
		}
		return es
	}
	basis := func() []model.Entity { return basisN(5) }
	opt := model.Optionen{OrderBy: []model.Sortierung{abOrd("P1", false), abOrd("P2", false), abOrd("P3", false)}}
	guid := "00000000-0000-0000-0000-000000000001"
	formen := []func(name string) model.PropertyWert{
		func(name string) model.PropertyWert { return abPW(name, model.ArtRedigiert, "") },
		func(name string) model.PropertyWert { return abPW(name, "foo", "1") },
		func(name string) model.PropertyWert { return abPW(name, model.ArtInt32, "01") },
		func(name string) model.PropertyWert { return abS(name, "x") },
		func(name string) model.PropertyWert { return abPW(name, model.ArtGUID, guid) },
	}
	pruefe := func(was string, es []model.Entity) {
		t.Helper()
		r, err := model.FuehreAus(typ, es, opt, nil)
		var ae *model.AbfrageFehler
		var fe *model.FilterTypFehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &ae) || errors.As(err, &fe) || !reflect.DeepEqual(r, model.Ergebnis{}) {
			t.Errorf("%s: %v %#v", was, err, r)
		}
	}
	n := 0
	for fi, form := range formen {
		for _, spalte := range []string{"P1", "P2", "P3", "K2"} {
			for pos := 0; pos < 5; pos++ {
				es := basis()
				abSetze(&es[pos], form(spalte))
				pruefe(fmt.Sprintf("Form %d, Spalte %s, Entity %d", fi, spalte, pos+1), es)
				n++
			}
		}
	}
	if n != 5*4*5 {
		t.Errorf("%d Fälle", n)
	}
	// Zahl der Entities: genau eine, zwei, drei und vier (bei einer Entity gibt es keinen Vergleich, der den Fehler fände)
	m := 0
	for zahl := 1; zahl <= 4; zahl++ {
		for fi, form := range formen {
			for _, spalte := range []string{"P1", "P2", "P3", "K2"} {
				// ein einzelner Text ist eine Spalte einer Klasse, eine einzelne Guid ist als Schlüssel zulässig: kein Fehler
				if zahl == 1 && (fi == 3 || (fi == 4 && spalte == "K2")) {
					continue
				}
				for pos := 0; pos < zahl; pos++ {
					es := basisN(zahl)
					abSetze(&es[pos], form(spalte))
					pruefe(fmt.Sprintf("%d Entities, Form %d, Spalte %s, Entity %d", zahl, fi, spalte, pos+1), es)
					m++
				}
			}
		}
	}
	if m != (2+3+4)*5*4+(5*4-5) {
		t.Errorf("%d Fälle nach Zahl der Entities", m)
	}
	es := basis()
	for i := range es {
		abSetze(&es[i], abPW("P2", model.ArtGUID, fmt.Sprintf("00000000-0000-0000-0000-%012d", i)))
	}
	pruefe("P2 nur Guid", es)
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
	abIDsGleich(t, "Ref eq", abOk(t, "eq", typ, es, model.Optionen{Filter: abF(vgl("eq", prop("Ref"), guid))}, nil).Entities, "3")
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
	navs := []string{"Nav2", "Nav3", "Nav4", "Nav5"}
	for _, nav := range navs {
		geschlossen.NavigationProperties = append(geschlossen.NavigationProperties, model.NavigationProperty{Name: nav, Target: "Ns.T"})
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
	for _, nav := range navs {
		arten = append(arten, art{"Navigation " + nav, nav, geschlossen, model.AbfrageTypNichtUnterstuetzt})
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
	if len(arten) != 43 || n != 2*43*3 {
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

// TestOrderByEigenschaftFehlt (Zeile 22): Entity ohne die Property an jeder der fünf Stellen × Element
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
		for _, luecke := range []int{1, 2, 3, 4, 5} {
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
	if n != 3*5*2 {
		t.Errorf("%d Fälle", n)
	}
	// mit Filter (Modus der Zusatztabelle): Stufe 5 prüft $orderby unabhängig davon, ob ein Filter gesetzt ist und wen er durchlässt
	filter := []model.FilterAusdruck{abIDGt("0"), abIDGt("100"), nicht(abIDGt("100"))}
	m := 0
	for _, f := range filter {
		for _, element := range namen {
			for _, luecke := range []int{1, 2, 3, 4, 5} {
				var es []model.Entity
				for i := 1; i <= 5; i++ {
					if i == luecke {
						es = append(es, voll(i, element))
					} else {
						es = append(es, voll(i, ""))
					}
				}
				o := []model.Sortierung{abOrd("Name", false), abOrd("Price", false), abOrd("Day", true)}
				r, err := model.FuehreAus(typ, es, model.Optionen{Filter: abF(f), OrderBy: o}, nil)
				if abKennung(err) != "$orderby:nicht_beobachtet" || !reflect.DeepEqual(r, model.Ergebnis{}) {
					t.Errorf("Filter %v, %s fehlt in Entity %d: %v", f, element, luecke, err)
				}
				m++
			}
		}
	}
	if m != 3*3*5 {
		t.Errorf("%d Fälle mit Filter", m)
	}
	es := []model.Entity{abE(1, abNu("Name"), abNu("Price"), abNu("Day")), voll(2, "")}
	abOk(t, "null gilt", typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd("Name", false), abOrd("Price", false), abOrd("Day", false)}}, nil)
}
