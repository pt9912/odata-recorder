package mock_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// schluesselModell hat je einen Typ mit einem Schlüssel `K` der acht EDM-Typen (Sets K0 bis K7) und den Typ T2 mit den
// Schlüsseln I (Edm.Int32) und S (Edm.String).
func schluesselModell() model.ServiceModel {
	var typen []model.EntityType
	var sets []model.EntitySet
	for i, edm := range achtTypen() {
		n := fmt.Sprintf("N.K%d", i)
		typen = append(typen, model.EntityType{Name: n, Keys: []string{"K"}, Properties: []model.Property{{Name: "K", EdmType: edm}}})
		sets = append(sets, model.EntitySet{Name: fmt.Sprintf("K%d", i), EntityType: n})
	}
	typen = append(typen, model.EntityType{Name: "N.T2", Keys: []string{"I", "S"}, Properties: []model.Property{{Name: "I", EdmType: "Edm.Int32"}, {Name: "S", EdmType: "Edm.String"}}})
	sets = append(sets, model.EntitySet{Name: "T2", EntityType: "N.T2"})
	return model.NeuesServiceModel("4.0", []string{"N"}, typen, sets, nil)
}

// literalJeTyp ist das gültige Literal des Schlüssels je EDM-Typ.
func literalJeTyp(edm string) model.Literal {
	return map[string]model.Literal{
		"Edm.String": {Art: model.LiteralZeichenkette, Text: "'a'"}, "Edm.Boolean": {Art: model.LiteralWahrheitswert, Text: "true"},
		"Edm.Int32": {Art: model.LiteralZahl, Text: "1"}, "Edm.Int64": {Art: model.LiteralZahl, Text: "1"}, "Edm.Decimal": {Art: model.LiteralZahl, Text: "1.5"},
		"Edm.Guid": {Art: model.LiteralGUID, Text: "00000000-0000-0000-0000-000000000000"}, "Edm.Date": {Art: model.LiteralDatum, Text: "2026-11-07"},
		"Edm.DateTimeOffset": {Art: model.LiteralZeitpunkt, Text: "2026-11-07T00:00:00Z"},
	}[edm]
}

// falschesLiteral ist ein Literal einer Art, die nicht zum Typ passt.
func falschesLiteral(edm string) model.Literal {
	if edm == "Edm.String" {
		return model.Literal{Art: model.LiteralZahl, Text: "1"}
	}
	return model.Literal{Art: model.LiteralZeichenkette, Text: "'x'"}
}

// TestSchluesselBilden hält L23 (M10, SPEC-196 Nr. 7 a): Namen genau die Keys, Literal-Art zum EDM-Typ, Text, Folge des Modells;
// ein Verstoß ist ODR-E6005 vor jedem Lesen. Verstoß 8 × Typ 8, nur die Paare, die der Vertrag erreichbar macht (Zählprüfung:
// 42 erreichbar, 22 gelten nicht); dazu der Rang Set unbekannt, Schlüssel, `$filter` an der Entity.
func TestSchluesselBilden(t *testing.T) {
	verstoesse := []string{"Name fremd", "Name doppelt", "Art passt nicht", "Teil fehlt", "unbenannt bei zwei Teilen", "Edm.Int32 außerhalb",
		"Edm.Int64 nicht ganzzahlig", "Name fehlt"}
	erreichbar, nichtErreichbar := 0, 0
	for i, edm := range achtTypen() {
		set := fmt.Sprintf("K%d", i)
		gut := model.SchluesselTeil{Name: "K", Literal: literalJeTyp(edm)}
		for _, v := range verstoesse {
			var teile []model.SchluesselTeil
			switch v {
			case "Name fremd":
				teile = []model.SchluesselTeil{{Name: "Z", Literal: gut.Literal}}
			case "Name doppelt":
				teile = []model.SchluesselTeil{gut, gut}
			case "Art passt nicht":
				teile = []model.SchluesselTeil{{Name: "K", Literal: falschesLiteral(edm)}}
			case "Teil fehlt":
				teile = nil
			case "unbenannt bei zwei Teilen":
				teile = []model.SchluesselTeil{{Literal: gut.Literal}, {Literal: gut.Literal}}
			case "Edm.Int32 außerhalb":
				teile = []model.SchluesselTeil{zahl("K", "2147483648")}
			case "Edm.Int64 nicht ganzzahlig":
				teile = []model.SchluesselTeil{zahl("K", "1.5")}
			}
			if v == "Name fehlt" || v == "Edm.Int32 außerhalb" && edm != "Edm.Int32" || v == "Edm.Int64 nicht ganzzahlig" && edm != "Edm.Int64" {
				nichtErreichbar++
				continue
			}
			erreichbar++
			faelle := [][]model.SchluesselTeil{teile}
			if v == "Edm.Int64 nicht ganzzahlig" {
				faelle = append(faelle, []model.SchluesselTeil{zahl("K", "9223372036854775808")})
			}
			for _, tl := range faelle {
				u := leseUmg(t)
				u.c.mc.modell = schluesselModell()
				u.ressource(model.RessourceEntity, set, model.Optionen{}, tl...)
				_, err := u.leseGET()
				f := erwarteFehler(t, edm+" "+v, err, model.CodeE6005)
				if f != nil && f.Ursache != nil {
					t.Errorf("%s %s: Ursache %v", edm, v, f.Ursache)
				}
				for _, op := range []string{"Entities.Get", "Entities.ListSet", "Coverage.Get"} {
					if u.p.anzahl(op) != 0 {
						t.Errorf("%s %s: %s gerufen", edm, v, op)
					}
				}
			}
		}
		// Kontrolle: der gültige Schlüssel erreicht Get mit seinem Text.
		u := leseUmg(t)
		u.c.mc.modell = schluesselModell()
		u.ressource(model.RessourceEntity, set, model.Optionen{}, gut)
		_, _ = u.leseGET()
		if ev := u.p.von("Entities.Get"); len(ev) != 1 || !reflect.DeepEqual(ev[0].args, []any{wsID, set, gut.Literal.Text}) {
			t.Errorf("%s: gültiger Schlüssel: Get %v", edm, ev)
		}
	}
	if erreichbar != 42 || nichtErreichbar != 22 || erreichbar+nichtErreichbar != 8*len(achtTypen()) {
		t.Errorf("%d erreichbar, %d nicht, erwartet 42 und 22", erreichbar, nichtErreichbar)
	}
	// Zusammengesetzt: die Folge des Modells, nicht der URL; ein Teil unbenannt bei zwei Teilen ist ein Verstoß.
	u := leseUmg(t)
	u.c.mc.modell = schluesselModell()
	s := model.SchluesselTeil{Name: "S", Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'x'"}}
	u.ressource(model.RessourceEntity, "T2", model.Optionen{}, s, zahl("I", "2"))
	_, _ = u.leseGET()
	if ev := u.p.von("Entities.Get"); len(ev) != 1 || ev[0].args[2] != "I=2,S='x'" {
		t.Errorf("zusammengesetzt: Get %v", ev)
	}
	// Rang: Set unbekannt vor Schlüssel vor `$filter` an der Entity.
	rang := []struct {
		name string
		set  string
		teil []model.SchluesselTeil
		opt  model.Optionen
		code model.Code
	}{
		{"Set und Schlüssel", "Unbekannt", nil, model.Optionen{}, model.CodeE6008},
		{"Schlüssel und Filter", setProducts, nil, model.Optionen{Filter: idGroesser("1")}, model.CodeE6005},
		{"Set und Filter", "Unbekannt", []model.SchluesselTeil{zahl("", "1")}, model.Optionen{Filter: idGroesser("1")}, model.CodeE6008},
		{"alle drei", "Unbekannt", nil, model.Optionen{Filter: idGroesser("1")}, model.CodeE6008},
	}
	for _, r := range rang {
		u := leseUmg(t)
		u.ressource(model.RessourceEntity, r.set, r.opt, r.teil...)
		_, err := u.leseGET()
		_ = erwarteFehler(t, r.name, err, r.code)
	}
}

// TestLookupTreffer hält L24 (SPEC-196 Nr. 7 c): Entities.Get genau einmal (nicht ListSet) mit (ctx, Workspace, Set,
// Literalform); ein vollständiger Treffer und ein `partial` mit bekanntem Namen sind 200, ein Tombstone ist ODR-E6008, ohne
// dass die Coverage gelesen wird.
func TestLookupTreffer(t *testing.T) {
	voll := produkt(1, "a")
	teil := eintrag(setProducts, "1", `{"Id":1,"Name":"a"}`, "Id", "Name")
	teil.Completeness = model.EntityPartial
	tot := produkt(1, "a")
	tot.Deleted = true
	for name, es := range map[string]model.EntityState{"vollständig": voll, "partial": teil, "Tombstone": tot} {
		u := leseUmg(t)
		u.en.setze(setProducts, es)
		u.entity(model.Optionen{Select: &model.SelectOption{Namen: []string{"Name"}}}, "1")
		_, err := u.leseGET()
		if name == "Tombstone" {
			_ = erwarteFehler(t, name, err, model.CodeE6008)
		} else if err != nil {
			t.Errorf("%s: %v", name, err)
		}
		ev := u.p.von("Entities.Get")
		if len(ev) != 1 || ev[0].ctx != u.ctx || !reflect.DeepEqual(ev[0].args, []any{wsID, setProducts, "1"}) {
			t.Errorf("%s: Get %v", name, ev)
		}
		if u.p.anzahl("Entities.ListSet") != 0 || u.p.anzahl("Coverage.Get") != 0 {
			t.Errorf("%s: ListSet %d, Coverage.Get %d", name, u.p.anzahl("Entities.ListSet"), u.p.anzahl("Coverage.Get"))
		}
	}
}

// TestLookupVollstaendigkeit hält L26 (M12, SPEC-196 Nr. 7 e): ein angeforderter Name außerhalb von KnownProperties ist
// ODR-E6003, `partial` ohne `$select` ist ODR-E6003, `$select` allein mit bekannten Properties die Projektion; ein unbekannter
// Name in `$select` ist ODR-E6005 und geht der Prüfung voraus. Eintrag 3 × `$select` 5 = 15.
func TestLookupVollstaendigkeit(t *testing.T) {
	eintraege := []struct {
		name  string
		known []string
		soll  [5]model.Code
	}{
		{"vollständig", alleNamen(), [5]model.Code{0, 0, 0, model.CodeE6005, 0}},
		{"partial", []string{"Id"}, [5]model.Code{model.CodeE6003, model.CodeE6003, model.CodeE6003, model.CodeE6005, 0}},
		{"partial mit bekanntem Teil", []string{"Id", "Name"}, [5]model.Code{model.CodeE6003, model.CodeE6003, 0, model.CodeE6005, 0}},
	}
	selects := []*model.SelectOption{nil, {Alle: true}, {Namen: []string{"Name"}}, {Namen: []string{"Unbekannt"}}, {Namen: []string{"Id"}}}
	n := 0
	for _, e := range eintraege {
		for i, s := range selects {
			n++
			u := leseUmg(t)
			u.en.setze(setProducts, eintrag(setProducts, "1", `{"Id":1,"Name":"a"}`, e.known...))
			u.entity(model.Optionen{Select: s}, "1")
			_, err := u.leseGET()
			if fehlercode(err) != e.soll[i] {
				t.Errorf("%s, Select %d: %v, erwartet %s", e.name, i, err, e.soll[i])
			}
		}
	}
	if n != 15 {
		t.Errorf("%d Läufe, erwartet 15", n)
	}
}

// TestLookupOptionen hält L27 (SPEC-196 Nr. 7 d): FuehreAus mit Optionen{Select}; ein unbekannter Name ist ODR-E6005, eine
// Navigation und eine Property eines nicht unterstützten Typs ODR-E6001 (vor der Prüfung der Vollständigkeit), `$filter` ist
// ODR-E6001, die übrigen Optionen sind ohne Wirkung (auch `$top=0`).
func TestLookupOptionen(t *testing.T) {
	ohne := []model.Optionen{
		{OrderBy: []model.Sortierung{{Name: "Unbekannt"}}}, {Top: model.ZahlOption{Gesetzt: true, Wert: 0}}, {Skip: model.ZahlOption{Gesetzt: true, Wert: 5}},
		{Count: model.WahrOption{Gesetzt: true, Wert: true}}, {Sonstige: []model.SonstigeOption{{Name: "a", Wert: "b"}}},
	}
	for i, o := range ohne {
		u := leseUmg(t)
		u.entity(o, "1")
		if _, err := u.leseGET(); err != nil || namenVon(u.letzteEntity().Entity) != "Id,Name" {
			t.Errorf("Option %d: %v", i, err)
		}
	}
	u := leseUmg(t)
	u.entity(model.Optionen{Filter: idGroesser("0")}, "1")
	_, err := u.leseGET()
	_ = erwarteFehler(t, "$filter", err, model.CodeE6001)
	if u.p.anzahl("Entities.Get") != 0 {
		t.Error("$filter: Get gerufen")
	}
	selects := []struct {
		name, set, sel string
		code           model.Code
	}{
		{"unbekannter Name", setProducts, "Unbekannt", model.CodeE6005}, {"Navigation", "Seltsam", "Nav", model.CodeE6001},
		{"nicht unterstützter Typ", "Seltsam", "Blob", model.CodeE6001},
	}
	for _, s := range selects {
		u := leseUmg(t)
		u.co.setze(baselineID, s.set, model.CoverageComplete)
		u.en.setze(s.set, eintrag(s.set, "1", `{"Id":1}`, "Id"))
		u.ressource(model.RessourceEntity, s.set, model.Optionen{Select: &model.SelectOption{Namen: []string{s.sel}}}, zahl("", "1"))
		_, err := u.leseGET()
		_ = erwarteFehler(t, s.name, err, s.code)
	}
}

// TestAuswahl hält L28 (SPEC-198 Nr. 3): Auswahl an den Schreiber ist nil ohne `$select`, `*` zuerst, die Namen in der Folge der
// URL; für Collection und Entity. Schleife über 5 Formen × Art 2.
func TestAuswahl(t *testing.T) {
	formen := []struct {
		sel  *model.SelectOption
		soll []string
	}{
		{nil, nil}, {&model.SelectOption{Alle: true}, []string{"*"}}, {&model.SelectOption{Namen: []string{"Id", "Name"}}, []string{"Id", "Name"}},
		{&model.SelectOption{Namen: []string{"Name", "Id"}}, []string{"Name", "Id"}}, {&model.SelectOption{Alle: true, Namen: []string{"Name"}}, []string{"*", "Name"}},
	}
	for i, f := range formen {
		for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity} {
			u := leseUmg(t)
			u.ressource(art, setProducts, model.Optionen{Select: f.sel}, zahl("", "1"))
			if _, err := u.leseGET(); err != nil {
				t.Fatal(err)
			}
			var got []string
			if art == model.RessourceMenge {
				got = u.letzteCollection().Auswahl
			} else {
				got = u.letzteEntity().Auswahl
			}
			if !reflect.DeepEqual(got, f.soll) {
				t.Errorf("Form %d, Art %d: Auswahl %#v, erwartet %#v", i, art, got, f.soll)
			}
		}
	}
}

// TestAnzahl hält L30 (SPEC-196 Nr. 6, SPEC-198 Nr. 3): `/$count` ist 200 mit der Darstellung von SchreibeAnzahl(Ergebnis.Anzahl),
// die Zahl der nicht gelöschten Einträge nach dem Filter, mit dem Header der Coverage. Filter 4 × Tombstones 2.
func TestAnzahl(t *testing.T) {
	filter := []struct {
		name string
		f    *model.FilterOption
		soll int64
	}{{"ohne", nil, 4}, {"trifft keinen", idGroesser("100"), 0}, {"trifft alle", idGroesser("0"), 4}, {"trifft die Hälfte", idGroesser("2"), 2}}
	for _, f := range filter {
		for _, mitTot := range []bool{false, true} {
			u := leseUmg(t)
			es := []model.EntityState{produkt(1, "a"), produkt(2, "b"), produkt(3, "c"), produkt(4, "d")}
			if mitTot {
				t5, t6 := produkt(5, "e"), produkt(6, "f")
				t5.Deleted, t6.Deleted = true, true
				es = append(es, t5, t6)
			}
			u.en.setze(setProducts, es...)
			u.anzahl(model.Optionen{Filter: f.f})
			a, err := u.leseGET()
			if err != nil {
				t.Fatal(err)
			}
			body := string(lies(t, a))
			h, _ := coverageHeader(a)
			if u.letzteAnzahl() != f.soll || a.Status != 200 || h != "complete" || body != "PROBE-"+kanarienwert || a.Header["Content-Type"][0] != "text/plain" {
				t.Errorf("%s, Tombstones=%v: Zahl %d, Status %d, Header %q, Body %q", f.name, mitTot, u.letzteAnzahl(), a.Status, h, body)
			}
		}
	}
}

// TestTombstones hält L31 (SPEC-196 Nr. 5): ein Tombstone erscheint nie in Collection oder Anzahl. Gemischte Menge (3
// lebende, 2 Tombstones) × Art 2.
func TestTombstones(t *testing.T) {
	gemischt := func() []model.EntityState {
		t4, t5 := produkt(4, "d"), produkt(5, "e")
		t4.Deleted, t5.Deleted = true, true
		return []model.EntityState{produkt(1, "a"), t4, produkt(2, "b"), t5, produkt(3, "c")}
	}
	u := leseUmg(t)
	u.en.setze(setProducts, gemischt()...)
	u.menge(model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}})
	if _, err := u.leseGET(); err != nil {
		t.Fatal(err)
	}
	if c := u.letzteCollection(); ids(c.Entities) != "1,2,3" || c.Anzahl.Wert != 3 {
		t.Errorf("Collection: %s, Anzahl %d", ids(c.Entities), c.Anzahl.Wert)
	}
	u = leseUmg(t)
	u.en.setze(setProducts, gemischt()...)
	u.anzahl(model.Optionen{})
	if _, err := u.leseGET(); err != nil || u.letzteAnzahl() != 3 {
		t.Errorf("Anzahl: %v %d", err, u.letzteAnzahl())
	}
}

// TestReihenfolge hält L32 (SPEC-190 Nr. 1 b, LH-QA-02): die Reihenfolge der Antwort von ListSet hat keine Bedeutung, die
// Ausgabe ist gleich. Folge 3 (aufsteigend, absteigend, gemischt) × Option 3 (ohne, `$orderby`, `$top`), je 10 Wiederholungen
// (eine Iteration über eine Map in den Properties zeigte verschiedene Folgen).
func TestReihenfolge(t *testing.T) {
	volle := func(id int) model.EntityState {
		return eintrag(setProducts, fmt.Sprint(id), fmt.Sprintf(`{"Count":%d,"Ts":null,"Day":null,"Ref":null,"Price":null,"Big":null,"Flag":true,"Name":"n%d","Id":%d}`, id, id, id), alleNamen()...)
	}
	folgen := [][]int{{1, 2, 3, 4, 5, 6}, {6, 5, 4, 3, 2, 1}, {4, 1, 6, 2, 5, 3}}
	optionen := []model.Optionen{
		{}, {OrderBy: []model.Sortierung{{Name: "Name", Absteigend: true}}}, {Top: model.ZahlOption{Gesetzt: true, Wert: 4}},
	}
	for oi, o := range optionen {
		var erste string
		for fi, f := range folgen {
			for r := 0; r < 10; r++ {
				u := leseUmg(t)
				var es []model.EntityState
				for _, id := range f {
					es = append(es, volle(id))
				}
				u.en.setze(setProducts, es...)
				u.menge(o)
				if _, err := u.leseGET(); err != nil {
					t.Fatal(err)
				}
				var sb strings.Builder
				for _, e := range u.letzteCollection().Entities {
					fmt.Fprintf(&sb, "%s:%s;", e.Schluessel, namenVon(e))
				}
				if erste == "" {
					erste = sb.String()
				} else if sb.String() != erste {
					t.Errorf("Option %d, Folge %d, Lauf %d: %q, erwartet %q", oi, fi, r, sb.String(), erste)
				}
			}
		}
	}
}
