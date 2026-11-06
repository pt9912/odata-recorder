package model_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const abMax = 2147483647

// --- Bausteine ---

// abTyp ist typMit() mit der Key-Property Id und der Zusatz-Property N (Edm.Int32).
func abTyp() model.EntityType {
	t := typMit(model.Property{Name: "N", EdmType: "Edm.Int32"})
	t.Keys = []string{"Id"}
	return t
}

func abPW(name string, art model.WertArt, text string) model.PropertyWert {
	return model.PropertyWert{Name: name, Wert: model.Wert{Art: art, Text: text}}
}
func abI(name string, n int) model.PropertyWert { return abPW(name, model.ArtInt32, strconv.Itoa(n)) }
func abS(name, s string) model.PropertyWert     { return abPW(name, model.ArtString, s) }
func abNu(name string) model.PropertyWert       { return abPW(name, model.ArtNull, "") }

// abE ist eine Entity mit der Key-Property Id und dem Schlüssel `id`.
func abE(id int, props ...model.PropertyWert) model.Entity {
	return model.Entity{Typ: "Ns.T", Schluessel: strconv.Itoa(id),
		Properties: slices.Concat([]model.PropertyWert{abI("Id", id)}, props), Uebersprungen: []model.Uebersprungen{}}
}

func abIds(es []model.Entity) []string {
	r := make([]string, 0, len(es))
	for _, e := range es {
		r = append(r, e.Schluessel)
	}
	return r
}

func abF(a model.FilterAusdruck) *model.FilterOption { return &model.FilterOption{Text: "x", Ausdruck: a} }
func abZ(n int) model.ZahlOption                     { return model.ZahlOption{Gesetzt: true, Wert: n} }
func abSel(namen ...string) *model.SelectOption      { return &model.SelectOption{Namen: namen} }
func abOrd(name string, desc bool) model.Sortierung  { return model.Sortierung{Name: name, Absteigend: desc} }

// abOk führt aus und verlangt keinen Fehler.
func abOk(t *testing.T, was string, typ model.EntityType, es []model.Entity, opt model.Optionen, beob func(string) bool) model.Ergebnis {
	t.Helper()
	r, err := model.FuehreAus(typ, es, opt, beob)
	if err != nil {
		t.Fatalf("%s: %v", was, err)
	}
	return r
}

// abIdsGleich vergleicht die Schlüssel der Entities mit der Erwartung.
func abIdsGleich(t *testing.T, was string, got []model.Entity, want ...string) {
	t.Helper()
	if g := abIds(got); !slices.Equal(g, want) || got == nil {
		t.Errorf("%s: %v, erwartet %v", was, g, want)
	}
}

// abKennung benennt einen Fehler: Option und Ursache, die Ursache der Typprüfung, ein Programmfehler oder ok.
func abKennung(err error) string {
	var ae *model.AbfrageFehler
	var fe *model.FilterTypFehler
	switch {
	case err == nil:
		return "ok"
	case errors.As(err, &ae):
		return string(ae.Option) + ":" + string(ae.Ursache)
	case errors.As(err, &fe):
		return "pruefe:" + string(fe.Ursache)
	case errors.Is(err, model.ErrUngueltig):
		return "programm"
	}
	return "?" + err.Error()
}

func abNamen(es []model.Entity, name string) []string {
	var r []string
	for _, e := range es {
		for _, p := range e.Properties {
			if p.Name == name {
				r = append(r, p.Wert.Text)
			}
		}
	}
	return r
}

// abDaten sind acht Entities mit Name und Price.
func abDaten() []model.Entity {
	namen := []string{"b", "a", "b", "a", "c", "c", "a", "b"}
	preise := []string{"10", "20", "30", "40", "50", "60", "70", "80"}
	es := make([]model.Entity, 0, 8)
	for i := range namen {
		es = append(es, abE(i+1, abS("Name", namen[i]), abPW("Price", model.ArtDecimal, preise[i])))
	}
	return es
}

func abIdGt(n string) model.FilterAusdruck { return vgl("gt", prop("Id"), lit(model.LiteralZahl, n)) }

// --- Zeile 1: Folge der Stufen ---

// TestStufen (Zeile 1): Filter, Zählung, Ordnung, $skip, $top, Projektion; dazu je eine Abfrage, in der die Folge zweier
// Nachbarstufen die Menge ändert.
func TestStufen(t *testing.T) {
	typ, es := abTyp(), abDaten()
	r := abOk(t, "alle Stufen", typ, es, model.Optionen{
		Filter: abF(abIdGt("1")), OrderBy: []model.Sortierung{abOrd("Name", false), abOrd("Price", true)},
		Skip: abZ(1), Top: abZ(3), Select: abSel("Name"),
	}, nil)
	abIdsGleich(t, "alle Stufen", r.Entities, "4", "2", "8")
	if r.Anzahl != 7 || !slices.Equal(abNamen(r.Entities, "Name"), []string{"a", "a", "b"}) {
		t.Errorf("Anzahl %d, Namen %v", r.Anzahl, abNamen(r.Entities, "Name"))
	}
	for _, e := range r.Entities {
		if len(e.Properties) != 1 || e.Properties[0].Name != "Name" {
			t.Errorf("Projektion: %v", e.Properties)
		}
	}
	// Filter gegen Paging: die Auswahl geht dem $skip voraus
	r = abOk(t, "Filter vor Paging", typ, es, model.Optionen{Filter: abF(vgl("ne", prop("Name"), lit(model.LiteralZeichenkette, "'a'"))), Skip: abZ(1), Top: abZ(2)}, nil)
	abIdsGleich(t, "Filter vor Paging", r.Entities, "3", "5")
	// Ordnung gegen Paging: die Ordnung geht dem $top voraus
	r = abOk(t, "Ordnung vor Paging", typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd("Price", true)}, Top: abZ(2)}, nil)
	abIdsGleich(t, "Ordnung vor Paging", r.Entities, "8", "7")
	// Ordnung gegen Projektion: $select Id verliert Name nicht vor der Ordnung
	r = abOk(t, "Ordnung vor Projektion", typ, es, model.Optionen{OrderBy: []model.Sortierung{abOrd("Name", false)}, Select: abSel("Id")}, nil)
	abIdsGleich(t, "Ordnung vor Projektion", r.Entities, "2", "4", "7", "1", "3", "8", "5", "6")
	for _, e := range r.Entities {
		if len(e.Properties) != 1 || e.Properties[0].Name != "Id" {
			t.Errorf("Projektion: %v", e.Properties)
		}
	}
	// $skip vor $top, bei $top über dem Rest
	r = abOk(t, "skip vor top", typ, es, model.Optionen{Skip: abZ(6), Top: abZ(5)}, nil)
	abIdsGleich(t, "skip vor top", r.Entities, "7", "8")
}

// --- Zeile 2: Anzahl ---

// TestAnzahl (Zeile 2): Anzahl ist die Zahl nach dem Filter und vor dem Paging, für Count nicht gesetzt, falsch, wahr:
// 3 × 5 × 4 = 60 Fälle; dazu ohne Filter und mit nur dem Filter (der Fall /$count).
func TestAnzahl(t *testing.T) {
	typ, es := abTyp(), abDaten()
	n := 0
	for _, c := range []model.WahrOption{{}, {Gesetzt: true, Wert: false}, {Gesetzt: true, Wert: true}} {
		for _, top := range []model.ZahlOption{{}, abZ(0), abZ(1), abZ(8), abZ(9)} {
			for _, skip := range []model.ZahlOption{{}, abZ(0), abZ(1), abZ(9)} {
				r := abOk(t, "Anzahl", typ, es, model.Optionen{Filter: abF(abIdGt("4")), Count: c, Top: top, Skip: skip}, nil)
				if r.Anzahl != 4 {
					t.Errorf("Count %v Top %v Skip %v: Anzahl %d", c, top, skip, r.Anzahl)
				}
				n++
			}
		}
	}
	if n != 60 {
		t.Errorf("%d Fälle", n)
	}
	if r := abOk(t, "ohne Filter", typ, es, model.Optionen{Top: abZ(2), Skip: abZ(1)}, nil); r.Anzahl != 8 {
		t.Errorf("ohne Filter: %d", r.Anzahl)
	}
	if r := abOk(t, "nur Filter", typ, es, model.Optionen{Filter: abF(abIdGt("4"))}, nil); r.Anzahl != 4 || len(r.Entities) != 4 {
		t.Errorf("nur Filter: %d", r.Anzahl)
	}
}

// --- Zeile 4: Paging ---

// TestPaging (Zeile 4): Top und Skip je aus {nicht gesetzt, 0, 1, n−1, n, n+1, 2147483647} (n−1 ab 0) × n ∈ {0, 1, 5}: 147 Fälle
// gegen den Ausschnitt der geordneten Liste.
func TestPaging(t *testing.T) {
	typ := abTyp()
	n := 0
	for _, zahl := range []int{0, 1, 5} {
		var es []model.Entity
		var alle []string
		for i := 1; i <= zahl; i++ {
			es = append(es, abE(i))
			alle = append(alle, strconv.Itoa(i))
		}
		werte := []model.ZahlOption{{}, abZ(0), abZ(1), abZ(max(zahl-1, 0)), abZ(zahl), abZ(zahl + 1), abZ(abMax)}
		for _, top := range werte {
			for _, skip := range werte {
				von := 0
				if skip.Gesetzt {
					von = min(skip.Wert, zahl)
				}
				bis := zahl
				if top.Gesetzt {
					bis = int(min(int64(zahl), int64(von)+int64(top.Wert)))
				}
				r := abOk(t, "Paging", typ, es, model.Optionen{Top: top, Skip: skip}, nil)
				abIdsGleich(t, fmt.Sprintf("n %d Top %v Skip %v", zahl, top, skip), r.Entities, alle[von:bis]...)
				if r.Anzahl != zahl {
					t.Errorf("Anzahl %d", r.Anzahl)
				}
				n++
			}
		}
	}
	if n != 147 {
		t.Errorf("%d Fälle", n)
	}
}

// --- Zeile 5: Gesetzt falsch ---

// TestGesetztFalsch (Zeile 5): mit Gesetzt falsch wird Wert nicht gelesen, auch nicht ein negativer.
func TestGesetztFalsch(t *testing.T) {
	typ := abTyp()
	es := []model.Entity{abE(1), abE(2), abE(3)}
	for _, z := range []model.ZahlOption{{Wert: 5}, {Wert: -3}, {Wert: abMax}, {Wert: 0}, {Wert: -1}, {Wert: -abMax - 1}} {
		for _, opt := range []model.Optionen{{Top: z}, {Skip: z}, {Top: z, Skip: z}} {
			r := abOk(t, "Gesetzt falsch", typ, es, opt, nil)
			abIdsGleich(t, fmt.Sprintf("%+v", opt), r.Entities, "1", "2", "3")
		}
	}
}

// --- Zeile 19: Reihenfolge der Eingabe ---

func permutationen(n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	var aus [][]int
	for _, p := range permutationen(n - 1) {
		for i := 0; i <= len(p); i++ {
			q := slices.Insert(slices.Clone(p), i, n-1)
			aus = append(aus, q)
		}
	}
	return aus
}

// TestPermutation (Zeile 19): sechs Entities mit Gleichständen in jeder der 720 Reihenfolgen × drei Abfragen ergeben dieselbe
// Folge; hundert Wiederholungen derselben Eingabe sind tief gleich.
func TestPermutation(t *testing.T) {
	typ := abTyp()
	ns := []int{2, 1, 2, 1, 3, 2}
	basis := make([]model.Entity, 0, 6)
	for i, n := range ns {
		basis = append(basis, abE(i+1, abI("N", n)))
	}
	abfragen := []struct {
		opt  model.Optionen
		soll []string
	}{
		{model.Optionen{OrderBy: []model.Sortierung{abOrd("N", false)}}, []string{"2", "4", "1", "3", "6", "5"}},
		{model.Optionen{Filter: abF(vgl("gt", prop("N"), lit(model.LiteralZahl, "1"))), Top: abZ(3)}, []string{"1", "3", "5"}},
		{model.Optionen{}, []string{"1", "2", "3", "4", "5", "6"}},
	}
	perms := permutationen(6)
	if len(perms) != 720 {
		t.Fatalf("%d Permutationen", len(perms))
	}
	for _, p := range perms {
		es := make([]model.Entity, 0, 6)
		for _, i := range p {
			es = append(es, basis[i])
		}
		for i, q := range abfragen {
			r := abOk(t, "Permutation", typ, es, q.opt, nil)
			if g := abIds(r.Entities); !slices.Equal(g, q.soll) {
				t.Fatalf("Reihenfolge %v, Abfrage %d: %v, erwartet %v", p, i, g, q.soll)
			}
		}
	}
	erste := abOk(t, "Wiederholung", typ, basis, abfragen[0].opt, nil)
	for i := 0; i < 100; i++ {
		if r := abOk(t, "Wiederholung", typ, basis, abfragen[0].opt, nil); !reflect.DeepEqual(r, erste) {
			t.Fatalf("Wiederholung %d weicht ab", i)
		}
	}
}

// --- Zeilen 27 bis 30: Projektion ---

// refProjektion ist die Erwartung der Projektion: ohne Select oder mit Alle alle Properties, sonst die genannten in der Folge
// der Entity.
func refProjektion(props []model.PropertyWert, sel *model.SelectOption) []string {
	var r []string
	for _, p := range props {
		if sel == nil || sel.Alle || slices.Contains(sel.Namen, p.Name) {
			r = append(r, p.Name)
		}
	}
	return r
}

func propNamen(e model.Entity) []string {
	r := []string{}
	for _, p := range e.Properties {
		r = append(r, p.Name)
	}
	return r
}

// TestProjektion (Zeile 27): sieben Select-Formen × Zustand (bekannt, null, fehlend) × vier Properties, und die Folge der Eingabe
// statt der URL.
func TestProjektion(t *testing.T) {
	typ := abTyp()
	vier := []string{"Name", "Price", "Day", "At"}
	selects := []*model.SelectOption{nil, {Alle: true}, abSel("Name"), abSel(vier...), abSel("Day", "Name"), {Alle: true, Namen: []string{"Name"}},
		abSel("At", "Day", "Price", "Name")}
	n := 0
	for _, sel := range selects {
		for _, p := range vier {
			for zustand := 0; zustand < 3; zustand++ {
				props := []model.PropertyWert{abS("Name", "x"), abPW("Price", model.ArtDecimal, "1"), abPW("Day", model.ArtDate, "2020-01-01"),
					abPW("At", model.ArtDateTimeOffset, "2020-01-01T00:00:00Z")}
				i := slices.Index(vier, p)
				switch zustand {
				case 1:
					props[i] = abNu(p)
				case 2:
					props = slices.Delete(props, i, i+1)
				}
				e := abE(1, props...)
				r := abOk(t, "Projektion", typ, []model.Entity{e}, model.Optionen{Select: sel}, nil)
				if g, soll := propNamen(r.Entities[0]), refProjektion(e.Properties, sel); !slices.Equal(g, soll) {
					t.Errorf("Select %+v, %s Zustand %d: %v, erwartet %v", sel, p, zustand, g, soll)
				}
				n++
			}
		}
	}
	if n != 7*4*3 {
		t.Errorf("%d Fälle", n)
	}
	e := abE(1, abS("Name", "x"), abPW("Price", model.ArtDecimal, "1"))
	for _, sel := range []*model.SelectOption{abSel("Price", "Name"), abSel("Name", "Price")} {
		r := abOk(t, "Folge", typ, []model.Entity{e}, model.Optionen{Select: sel}, nil)
		if g := propNamen(r.Entities[0]); !slices.Equal(g, []string{"Name", "Price"}) {
			t.Errorf("Folge der Eingabe: %v", g)
		}
	}
	// die Key-Property wird nicht ergänzt, `*` neben Namen liefert alle
	r := abOk(t, "Key", typ, []model.Entity{e}, model.Optionen{Select: abSel("Name")}, nil)
	if g := propNamen(r.Entities[0]); !slices.Equal(g, []string{"Name"}) {
		t.Errorf("Key ergänzt: %v", g)
	}
	r = abOk(t, "Stern", typ, []model.Entity{e}, model.Optionen{Select: &model.SelectOption{Alle: true, Namen: []string{"Name"}}}, nil)
	if g := propNamen(r.Entities[0]); !slices.Equal(g, []string{"Id", "Name", "Price"}) {
		t.Errorf("Stern mit Namen: %v", g)
	}
}

// TestNullUndFehlend (Zeile 28): null bleibt mit der Art null, eine fehlende Property fehlt.
func TestNullUndFehlend(t *testing.T) {
	typ := abTyp()
	es := []model.Entity{abE(1, abNu("N")), abE(2), abE(3, abI("N", 5))}
	for _, sel := range []*model.SelectOption{nil, abSel("N")} {
		r := abOk(t, "null und fehlend", typ, es, model.Optionen{Select: sel}, nil)
		a, b, c := r.Entities[0], r.Entities[1], r.Entities[2]
		wa, okA := wertVonEntity(a, "N")
		_, okB := wertVonEntity(b, "N")
		wc, okC := wertVonEntity(c, "N")
		if !okA || wa.Art != model.ArtNull || okB || !okC || wc.Text != "5" {
			t.Errorf("Select %v: A %v %v, B %v, C %v %v", sel, wa, okA, okB, wc, okC)
		}
		wantLen := []int{2, 1, 2}
		if sel != nil {
			wantLen = []int{1, 0, 1}
		}
		if len(a.Properties) != wantLen[0] || len(b.Properties) != wantLen[1] || len(c.Properties) != wantLen[2] {
			t.Errorf("Select %v: Längen %d %d %d", sel, len(a.Properties), len(b.Properties), len(c.Properties))
		}
	}
}

func wertVonEntity(e model.Entity, name string) (model.Wert, bool) {
	for _, p := range e.Properties {
		if p.Name == name {
			return p.Wert, true
		}
	}
	return model.Wert{}, false
}

// TestSelectMaskeUndTeilprojektion (Zeile 29): eine Property, die die Maske nicht kennt, ist 422; eine, die nur eine Entity nicht
// kennt, fehlt in genau dieser Entity.
func TestSelectMaskeUndTeilprojektion(t *testing.T) {
	typ := abTyp()
	voll := make([]model.Entity, 0, 5)
	for i := 1; i <= 5; i++ {
		voll = append(voll, abE(i, abS("Name", "x")))
	}
	_, err := model.FuehreAus(typ, voll, model.Optionen{Select: abSel("Name")}, func(n string) bool { return n != "Name" })
	if k := abKennung(err); k != "$select:nicht_beobachtet" || !errors.Is(err, model.ErrFilterNichtBeobachtet) {
		t.Errorf("Maske falsch: %v", err)
	}
	teil := slices.Clone(voll)
	for _, i := range []int{0, 2, 4} {
		teil[i] = abE(i + 1)
	}
	for _, beob := range []func(string) bool{nil, func(string) bool { return true }} {
		r := abOk(t, "Teilprojektion", typ, teil, model.Optionen{Select: abSel("Name")}, beob)
		for i, e := range r.Entities {
			if _, ok := wertVonEntity(e, "Name"); ok != (i == 1 || i == 3) {
				t.Errorf("Entity %d: Name vorhanden %v", i, ok)
			}
		}
	}
}

// TestErgebnisUnabhaengig (Zeile 30): Typ und Schluessel bleiben, Uebersprungen ist nil, und Ergebnis und Eingabe teilen nichts.
func TestErgebnisUnabhaengig(t *testing.T) {
	typ := abTyp()
	mk := func() []model.Entity {
		e := abE(1, abS("Name", "x"), abI("N", 7))
		e.Uebersprungen = []model.Uebersprungen{{Name: "a", Grund: model.UebersprungTyp}, {Name: "b", Grund: model.UebersprungOffen}}
		return []model.Entity{e}
	}
	for _, sel := range []*model.SelectOption{nil, {Alle: true}, abSel("Name", "Id")} {
		in, vorher := mk(), mk()
		r := abOk(t, "unabhängig", typ, in, model.Optionen{Select: sel}, nil)
		e := r.Entities[0]
		if e.Typ != "Ns.T" || e.Schluessel != "1" || e.Uebersprungen != nil {
			t.Errorf("Select %v: %q %q %v", sel, e.Typ, e.Schluessel, e.Uebersprungen)
		}
		for i := range e.Properties {
			e.Properties[i] = abS("X", "geändert")
		}
		if !reflect.DeepEqual(in, vorher) {
			t.Errorf("Select %v: die Änderung des Ergebnisses erreicht die Eingabe", sel)
		}
		r2 := abOk(t, "unabhängig", typ, in, model.Optionen{Select: sel}, nil)
		for i := range in[0].Properties {
			in[0].Properties[i] = abS("Y", "geändert")
		}
		if g := r2.Entities[0].Properties; g[0].Name == "Y" || g[len(g)-1].Name == "Y" {
			t.Errorf("Select %v: die Änderung der Eingabe erreicht das Ergebnis", sel)
		}
	}
}

// --- Zeile 31: die Eingabe bleibt unverändert ---

type abAbfrage struct {
	name string
	opt  func() model.Optionen
	beob func(string) bool
}

// abShared sind zehn Entities mit allen Properties; die Folge ist durchmischt.
func abShared() []model.Entity {
	es := make([]model.Entity, 0, 10)
	for _, i := range []int{7, 3, 9, 1, 5, 10, 2, 8, 4, 6} {
		e := abE(i, abS("Name", string(rune('a'+i%4))), abI("N", i%3), abPW("Big", model.ArtInt64, strconv.Itoa(i*1000)),
			abPW("Price", model.ArtDecimal, strconv.Itoa(i%5)+".5"), abPW("Active", model.ArtBoolean, strconv.FormatBool(i%2 == 0)),
			abPW("Ref", model.ArtGUID, fmt.Sprintf("00000000-0000-0000-0000-%012d", 100-i)), abPW("Day", model.ArtDate, fmt.Sprintf("2020-01-%02d", i)),
			abPW("At", model.ArtDateTimeOffset, fmt.Sprintf("2020-01-01T00:00:%02dZ", i)))
		if i == 3 || i == 6 {
			e.Properties = slices.DeleteFunc(e.Properties, func(p model.PropertyWert) bool { return p.Name == "Day" })
		}
		es = append(es, e)
	}
	return es
}

// abAbfragen sind Abfragen, die jeden Zweig des Interpreters durchlaufen; opt baut die Optionen frisch (und die Bäume mit ihnen).
func abAbfragen() []abAbfrage {
	str := func(s string) model.FilterAusdruck { return lit(model.LiteralZeichenkette, s) }
	alle := func(string) bool { return true }
	ohne := func(n string) bool { return n != "Price" && n != "Big" }
	return []abAbfrage{
		{"alles", func() model.Optionen { return model.Optionen{} }, nil},
		{"Filter und Paging", func() model.Optionen {
			return model.Optionen{Filter: abF(verk("and", abIdGt("2"), nicht(vgl("eq", prop("Name"), str("'b'"))))), Skip: abZ(1), Top: abZ(4)}
		}, alle},
		{"Filter or null", func() model.Optionen {
			return model.Optionen{Filter: abF(verk("or", vgl("eq", prop("Active"), wahrLit()), vgl("eq", prop("N"), nul())))}
		}, nil},
		{"Ordnung gemischt", func() model.Optionen {
			return model.Optionen{OrderBy: []model.Sortierung{abOrd("Name", true), abOrd("Price", false), abOrd("At", true), abOrd("Big", false)}}
		}, nil},
		{"Ordnung Day", func() model.Optionen {
			return model.Optionen{OrderBy: []model.Sortierung{abOrd("N", false)}, Select: abSel("N", "Name")}
		}, nil},
		{"Select Alle", func() model.Optionen { return model.Optionen{Select: &model.SelectOption{Alle: true}, Top: abZ(3)} }, nil},
		{"Select Namen", func() model.Optionen { return model.Optionen{Select: abSel("Name", "Id", "N"), Skip: abZ(2)} }, nil},
		{"Top 0", func() model.Optionen { return model.Optionen{Top: abZ(0)} }, nil},
		{"Skip über das Ende", func() model.Optionen { return model.Optionen{Skip: abZ(abMax)} }, nil},
		{"eigenschaft_unbekannt", func() model.Optionen { return model.Optionen{OrderBy: []model.Sortierung{abOrd("Foo", false)}} }, nil},
		{"typ_nicht_unterstuetzt", func() model.Optionen { return model.Optionen{Select: abSel("Nav")} }, nil},
		{"ordnung_nicht_unterstuetzt", func() model.Optionen { return model.Optionen{OrderBy: []model.Sortierung{abOrd("Ref", false)}} }, nil},
		{"nicht_beobachtet", func() model.Optionen { return model.Optionen{Select: abSel("Price")} }, ohne},
		{"Entity ohne Property", func() model.Optionen { return model.Optionen{OrderBy: []model.Sortierung{abOrd("Day", false)}} }, nil},
		{"Filter ohne Property", func() model.Optionen { return model.Optionen{Filter: abF(vgl("eq", prop("Day"), nul()))} }, nil},
		{"Fehler der Typprüfung", func() model.Optionen { return model.Optionen{Filter: abF(vgl("eq", prop("Foo"), nul()))} }, nil},
		{"Programmfehler", func() model.Optionen { return model.Optionen{Top: abZ(-1)} }, nil},
		{"Programmfehler Select", func() model.Optionen { return model.Optionen{Select: abSel("Id", "Id")} }, nil},
		{"Programmfehler Filter", func() model.Optionen { return model.Optionen{Filter: &model.FilterOption{}} }, nil},
		{"Programmfehler Auswertung", func() model.Optionen {
			return model.Optionen{Filter: abF(vgl("eq", prop("N"), lit(model.LiteralWahrheitswert, "true")))}
		}, nil},
		{"Programmfehler OrderBy", func() model.Optionen {
			return model.Optionen{OrderBy: []model.Sortierung{abOrd("N", false), abOrd("N", true)}}
		}, nil},
	}
}

// TestEingabeUnveraendert (Zeile 31): Reihenfolge der Liste, Entities, Optionen und Baum sind nach dem Aufruf gleich einer zweiten,
// unabhängig gebauten Eingabe, auch bei einem Fehler jeder Stufe. Rot in jedem Lauf, nicht erst unter dem Detektor.
func TestEingabeUnveraendert(t *testing.T) {
	typ := abTyp()
	for _, q := range abAbfragen() {
		es, opt := abShared(), q.opt()
		_, _ = model.FuehreAus(typ, es, opt, q.beob)
		if !reflect.DeepEqual(es, abShared()) || !reflect.DeepEqual(opt, q.opt()) {
			t.Errorf("%s: die Eingabe ist verändert", q.name)
		}
	}
	// Abfragen der Zeilen 16 und 19: Ordnung nach mehreren Schlüsseln
	for _, o := range [][]model.Sortierung{{abOrd("Name", true), abOrd("N", false)}, {abOrd("Big", true)}} {
		es, opt := abShared(), model.Optionen{OrderBy: o, Select: abSel("Name", "Id"), Filter: abF(abIdGt("1"))}
		abOk(t, "Ordnung", typ, es, opt, nil)
		if !reflect.DeepEqual(es, abShared()) || !reflect.DeepEqual(opt, model.Optionen{OrderBy: o, Select: abSel("Name", "Id"), Filter: abF(abIdGt("1"))}) {
			t.Errorf("Ordnung %v: die Eingabe ist verändert", o)
		}
	}
}

// --- Zeile 32: gleichzeitige Aufrufe ---

// TestAbfrageGleichzeitig (Zeile 32): acht Goroutinen mit je 1000 Aufrufen über die Abfragen von abAbfragen auf geteilter Liste, geteilten
// Optionen und geteiltem Baum; jedes Ergebnis ist gleich dem sequentiellen Lauf, und die Zahl der Aufrufe je Abfrage wird gezählt.
func TestAbfrageGleichzeitig(t *testing.T) {
	typ, es := abTyp(), abShared()
	qs := abAbfragen()
	opts := make([]model.Optionen, len(qs))
	soll := make([]model.Ergebnis, len(qs))
	sollErr := make([]string, len(qs))
	for i, q := range qs {
		opts[i] = q.opt()
		r, err := model.FuehreAus(typ, es, opts[i], q.beob)
		soll[i], sollErr[i] = r, abKennung(err)
	}
	const goroutinen, aufrufe = 8, 1000
	zaehler := make([]int, len(qs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	fehler := make(chan string, goroutinen)
	for g := 0; g < goroutinen; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lokal := make([]int, len(qs))
			for k := 0; k < aufrufe; k++ {
				i := (k + g) % len(qs)
				r, err := model.FuehreAus(typ, es, opts[i], qs[i].beob)
				if !reflect.DeepEqual(r, soll[i]) || abKennung(err) != sollErr[i] {
					select {
					case fehler <- qs[i].name:
					default:
					}
					return
				}
				lokal[i]++
			}
			mu.Lock()
			for i, n := range lokal {
				zaehler[i] += n
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	close(fehler)
	for name := range fehler {
		t.Errorf("Abfrage %q: das Ergebnis weicht vom sequentiellen Lauf ab", name)
	}
	gesamt := 0
	for i, n := range zaehler {
		if n == 0 {
			t.Errorf("Abfrage %q wurde nie ausgeführt", qs[i].name)
		}
		gesamt += n
	}
	if !t.Failed() && gesamt != goroutinen*aufrufe {
		t.Errorf("%d Aufrufe gezählt", gesamt)
	}
	if !reflect.DeepEqual(es, abShared()) {
		t.Error("die geteilte Liste ist verändert")
	}
}

// --- Zeile 33: Prüfgrenze ---

// TestGrosseMenge (Zeile 33): 10.000 und 10.001 Entities, Gleichstände auf 1.000 Werten, gegen eine Referenz.
func TestGrosseMenge(t *testing.T) {
	typ := abTyp()
	for _, zahl := range []int{10000, 10001} {
		es := make([]model.Entity, 0, zahl)
		type paar struct{ id, n int }
		var ref []paar
		for i := 0; i < zahl; i++ {
			id := (i*7919)%zahl + 1 // Permutation der Schlüssel 1..zahl (7919 teilt weder 10000 noch 10001)
			n := id % 1000
			es = append(es, abE(id, abI("N", n)))
			if id > 100 {
				ref = append(ref, paar{id, n})
			}
		}
		slices.SortFunc(ref, func(a, b paar) int {
			if a.n != b.n {
				return b.n - a.n
			}
			return a.id - b.id
		})
		var soll []string
		for _, p := range ref[5:55] {
			soll = append(soll, strconv.Itoa(p.id))
		}
		r := abOk(t, "große Menge", typ, es, model.Optionen{Filter: abF(abIdGt("100")), OrderBy: []model.Sortierung{abOrd("N", true)}, Skip: abZ(5), Top: abZ(50)}, nil)
		abIdsGleich(t, fmt.Sprintf("%d Entities", zahl), r.Entities, soll...)
		if r.Anzahl != zahl-100 {
			t.Errorf("%d Entities: Anzahl %d", zahl, r.Anzahl)
		}
	}
}
