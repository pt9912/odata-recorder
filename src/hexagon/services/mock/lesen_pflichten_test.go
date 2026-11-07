package mock_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// einEintrag legt einen Eintrag mit dem Schlüssel 1 und dem Payload in Products, ruft die Collection und liefert die eine
// Entity der Antwort des Schreibers.
func (u *umg) einEintrag(payload string) (model.Entity, error) {
	u.t.Helper()
	u.en.setze(setProducts, eintrag(setProducts, "1", payload, alleNamen()...))
	u.menge(model.Optionen{})
	a, err := u.leseGET()
	if err != nil {
		return model.Entity{}, err
	}
	lies(u.t, a)
	c := u.letzteCollection()
	if len(c.Entities) != 1 {
		u.t.Fatalf("%d Entities, erwartet 1", len(c.Entities))
	}
	return c.Entities[0], nil
}

// wertVon liefert den Wert der Property name und ob sie da ist.
func wertVon(e model.Entity, name string) (model.Wert, bool) {
	for _, p := range e.Properties {
		if p.Name == name {
			return p.Wert, true
		}
	}
	return model.Wert{}, false
}

// TestPayloadWerte hält L04 (M9, SPEC-196 Nr. 3): jede der acht Arten und `null` aus der Kanonform, Orakel eine Tabelle von Hand;
// int64 über 2 hoch 53, decimal mit 30 Stellen, nie über float64. Typ 8 × Wert 4 = 32, dazu der leere String und ein String
// mit NUL; `Properties` stehen in der Folge des Modells, `@odata.type` setzt `Typ`, `Schluessel` ist `EntityKey`.
func TestPayloadWerte(t *testing.T) {
	nul := "a" + string(rune(92)) + "u0000b"
	zahlen30 := "123456789012345678901234567890"
	faelle := []struct {
		edm, json string
		art       model.WertArt
		text      string
	}{
		{"Edm.String", `"abc"`, model.ArtString, "abc"}, {"Edm.String", `""`, model.ArtString, ""}, {"Edm.String", `"` + strings.Repeat("x", 4096) + `"`, model.ArtString, strings.Repeat("x", 4096)},
		{"Edm.String", `null`, model.ArtNull, ""}, {"Edm.String", `"` + nul + `"`, model.ArtString, "a\x00b"},
		{"Edm.Boolean", `true`, model.ArtBoolean, "true"}, {"Edm.Boolean", `false`, model.ArtBoolean, "false"}, {"Edm.Boolean", `true`, model.ArtBoolean, "true"},
		{"Edm.Boolean", `null`, model.ArtNull, ""},
		{"Edm.Int32", `42`, model.ArtInt32, "42"}, {"Edm.Int32", `-2147483648`, model.ArtInt32, "-2147483648"}, {"Edm.Int32", `2147483647`, model.ArtInt32, "2147483647"},
		{"Edm.Int32", `null`, model.ArtNull, ""},
		{"Edm.Int64", `"42"`, model.ArtInt64, "42"}, {"Edm.Int64", `"-9223372036854775808"`, model.ArtInt64, "-9223372036854775808"},
		{"Edm.Int64", `"9223372036854775807"`, model.ArtInt64, "9223372036854775807"}, {"Edm.Int64", `null`, model.ArtNull, ""},
		{"Edm.Decimal", `"1.50"`, model.ArtDecimal, "1.50"}, {"Edm.Decimal", `"-0.000000000000000000000000000001"`, model.ArtDecimal, "-0.000000000000000000000000000001"},
		{"Edm.Decimal", `"` + zahlen30 + `"`, model.ArtDecimal, zahlen30}, {"Edm.Decimal", `null`, model.ArtNull, ""},
		{"Edm.Guid", `"01234567-89ab-cdef-0123-456789abcdef"`, model.ArtGUID, "01234567-89ab-cdef-0123-456789abcdef"},
		{"Edm.Guid", `"00000000-0000-0000-0000-000000000000"`, model.ArtGUID, "00000000-0000-0000-0000-000000000000"},
		{"Edm.Guid", `"ffffffff-ffff-ffff-ffff-ffffffffffff"`, model.ArtGUID, "ffffffff-ffff-ffff-ffff-ffffffffffff"}, {"Edm.Guid", `null`, model.ArtNull, ""},
		{"Edm.Date", `"2026-11-07"`, model.ArtDate, "2026-11-07"}, {"Edm.Date", `"0001-01-01"`, model.ArtDate, "0001-01-01"},
		{"Edm.Date", `"9999-12-31"`, model.ArtDate, "9999-12-31"}, {"Edm.Date", `null`, model.ArtNull, ""},
		{"Edm.DateTimeOffset", `"2026-11-07T12:30:45.123Z"`, model.ArtDateTimeOffset, "2026-11-07T12:30:45.123Z"},
		{"Edm.DateTimeOffset", `"0001-01-01T00:00:00Z"`, model.ArtDateTimeOffset, "0001-01-01T00:00:00Z"},
		{"Edm.DateTimeOffset", `"9999-12-31T23:59:59.9999999Z"`, model.ArtDateTimeOffset, "9999-12-31T23:59:59.9999999Z"},
		{"Edm.DateTimeOffset", `null`, model.ArtNull, ""},
	}
	je := map[string]int{}
	for _, f := range faelle {
		je[f.edm]++
		u := leseUmg(t)
		name := propJeTyp(f.edm)
		e, err := u.einEintrag(`{"Id":1,"` + name + `":` + f.json + `}`)
		if err != nil {
			t.Errorf("%s %s: %v", f.edm, f.json, err)
			continue
		}
		w, da := wertVon(e, name)
		if !da || w != (model.Wert{Art: f.art, Text: f.text}) {
			t.Errorf("%s %.40s: Wert %+v (da=%v), erwartet %v %q", f.edm, f.json, w, da, f.art, f.text)
		}
	}
	if len(je) != 8 || len(faelle) != 8*4+1 {
		t.Errorf("%d Typen und %d Fälle, erwartet 8 und 33", len(je), len(faelle))
	}
	for _, edm := range achtTypen() {
		if je[edm] < 4 {
			t.Errorf("%s: %d Werte, erwartet mindestens 4", edm, je[edm])
		}
	}
	// Folge des Modells, Typ und Schlüssel.
	u := leseUmg(t)
	e, err := u.einEintrag(`{"Name":"n","Count":5,"Id":1}`)
	if err != nil || namenVon(e) != "Id,Name,Count" || e.Typ != typProduct || e.Schluessel != "1" {
		t.Errorf("Folge: %v %+v", err, e)
	}
	u = leseUmg(t)
	e, err = u.einEintrag(`{"@odata.type":"#` + typSpecial + `","Id":1,"Extra":"e"}`)
	if err != nil || namenVon(e) != "Id,Extra" || e.Typ != typSpecial {
		t.Errorf("@odata.type: %v %+v", err, e)
	}
}

// TestPayloadVerstoesse hält L05 (M8, SPEC-196 Nr. 3 d): ein Payload gegen die Form ist ODR-E9001 ohne Wert und ohne
// Ursache, der Kanarienwert im Payload steht nicht im Fehlertext. Verstoß 9 + falscher JSON-Typ × Art 8 = 17 Läufe, dazu
// drei Randformen.
func TestPayloadVerstoesse(t *testing.T) {
	k := kanarienwert
	faelle := []struct{ name, payload string }{
		{"kein Objekt", `["` + k + `"]`}, {"kein JSON", `{"Id":1,"Name":"` + k + `"`}, {"@odata.type kein String", `{"@odata.type":5,"Id":1,"Name":"` + k + `"}`},
		{"@odata.type unbekannt", `{"@odata.type":"#N.` + k + `","Id":1}`}, {"@odata.type abstrakt", `{"@odata.type":"#N.Abstrakt","Id":1}`},
		{"@odata.type fremd", `{"@odata.type":"#N.Fremd","Id":1}`}, {"unbekannte Property", `{"Id":1,"` + k + `":1}`},
		{"null bei nicht nullbarer Property", `{"Id":null,"Name":"` + k + `"}`}, {"Mitglied doppelt", `{"Id":1,"Name":"a","Name":"` + k + `"}`},
		{"String als Zahl", `{"Id":1,"Name":5}`}, {"Boolean als String", `{"Id":1,"Flag":"` + k + `"}`}, {"Int32 als String", `{"Id":1,"Count":"` + k + `"}`},
		{"Int64 als Zahl", `{"Id":1,"Big":5}`}, {"Decimal als Zahl", `{"Id":1,"Price":5}`}, {"Guid als Zahl", `{"Id":1,"Ref":5}`},
		{"Date als Zahl", `{"Id":1,"Day":5}`}, {"DateTimeOffset als Zahl", `{"Id":1,"Ts":5}`},
		{"@odata.type ohne #", `{"@odata.type":"` + typSpecial + `","Id":1}`}, {"Rest hinter dem Objekt", `{"Id":1} x`},
		{"Property des abgeleiteten Typs ohne @odata.type", `{"Id":1,"Extra":"e"}`},
	}
	for _, f := range faelle {
		u := leseUmg(t)
		_, err := u.einEintrag(f.payload)
		fe := erwarteFehler(t, f.name, err, model.CodeE9001)
		if fe != nil && fe.Ursache != nil {
			t.Errorf("%s: Ursache %v", f.name, fe.Ursache)
		}
		if u.p.anzahl("Schreiber.SchreibeCollection") != 0 {
			t.Errorf("%s: der Schreiber wurde gerufen", f.name)
		}
	}
	if len(faelle) != 9+8+3 {
		t.Errorf("%d Fälle, erwartet 20", len(faelle))
	}
}

// TestMarker hält L06 (SPEC-196 Nr. 3 b): der Marker `[REDACTED]` ist bei den sechs Typen, die als String stehen, `redigiert`
// (der Wert fehlt in der Ausgabe), bei Edm.Int32 und Edm.Boolean ein Verstoß; `[redacted]` und `[REDACTED] ` sind gewöhnliche
// Strings. Schleife über 8 Typen × 3 Texte.
func TestMarker(t *testing.T) {
	n := 0
	for _, edm := range achtTypen() {
		for _, text := range []string{model.Marker, "[redacted]", model.Marker + " "} {
			n++
			u := leseUmg(t)
			name := propJeTyp(edm)
			e, err := u.einEintrag(`{"Id":1,"` + name + `":"` + text + `"}`)
			istString := edm != "Edm.Int32" && edm != "Edm.Boolean"
			w, da := wertVon(e, name)
			switch {
			case !istString:
				_ = erwarteFehler(t, edm+" "+text, err, model.CodeE9001)
			case err != nil:
				t.Errorf("%s %q: %v", edm, text, err)
			case text == model.Marker && da:
				t.Errorf("%s: der Marker wird ausgeliefert: %+v", edm, w)
			case text != model.Marker && (!da || w.Text != text):
				t.Errorf("%s %q: Wert %+v (da=%v)", edm, text, w, da)
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Läufe, erwartet 24", n)
	}
}

// TestMaske hält L07 (M13, SPEC-190 Nr. 1 d, SPEC-196 Nr. 4 a): die Maske ist die Vereinigung der KnownProperties über alle
// Einträge des Sets einschließlich der Tombstones; nil bei leerem Set und `complete`, leer bei leerem Set und `partial` oder
// `unknown`. Fall 6 × Option 3 (`$filter`, `$orderby`, `$select`) = 18; eine nicht beobachtete Property ist ODR-E6003.
func TestMaske(t *testing.T) {
	mit := func(id int, known ...string) model.EntityState { return produkt(id, "n", known...) }
	tot := func(id int, known ...string) model.EntityState { e := mit(id, known...); e.Deleted = true; return e }
	faelle := []struct {
		name      string
		eintraege []model.EntityState
		status    model.CoverageStatus
		ok        bool
	}{
		{"Property in einem Eintrag", []model.EntityState{mit(1, "Id"), mit(2, "Id", "Name"), mit(3, "Id")}, model.CoverageComplete, true},
		{"Property nur im Tombstone", []model.EntityState{mit(1, "Id"), tot(2, "Id", "Name")}, model.CoverageComplete, true},
		{"Property in keinem", []model.EntityState{mit(1, "Id"), tot(2, "Id")}, model.CoverageComplete, false},
		{"leeres Set complete", nil, model.CoverageComplete, true},
		{"leeres Set partial", nil, model.CoveragePartial, false},
		{"leeres Set unknown", nil, model.CoverageUnknown, false},
	}
	optionen := []struct {
		name string
		o    model.Optionen
	}{
		{"filter", model.Optionen{Filter: filterOpt(nameGleich("n"))}},
		{"orderby", model.Optionen{OrderBy: []model.Sortierung{{Name: "Name"}}}},
		{"select", model.Optionen{Select: &model.SelectOption{Namen: []string{"Name"}}}},
	}
	n := 0
	for _, f := range faelle {
		for _, o := range optionen {
			n++
			u := leseUmg(t)
			u.co.setze(baselineID, setProducts, f.status)
			u.en.setze(setProducts, f.eintraege...)
			u.menge(o.o)
			_, err := u.leseGET()
			if f.ok && err != nil || !f.ok && fehlercode(err) != model.CodeE6003 {
				t.Errorf("%s, %s: %v, erwartet ok=%v", f.name, o.name, err, f.ok)
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Läufe, erwartet 18", n)
	}
	// Bei der Entity ist die Maske nil: die Vollständigkeit prüft der Lookup selbst (L26).
}

// TestAnzahlNurFilter hält L08 (SPEC-190 Nr. 2, SPEC-196 Nr. 4 b): `/$count` übergibt allein den Filter; `$orderby`, `$select`
// (auch mit unbekannter Property), `$top`, `$skip` und `$count` ändern weder Zahl noch Status, und die Zahl ist die von
// `@odata.count`. Schleife über 5 Optionen × Filter (ohne, mit).
func TestAnzahlNurFilter(t *testing.T) {
	optionen := []model.Optionen{
		{OrderBy: []model.Sortierung{{Name: "Unbekannt"}}}, {Select: &model.SelectOption{Namen: []string{"Unbekannt"}}},
		{Top: model.ZahlOption{Gesetzt: true, Wert: 1}}, {Skip: model.ZahlOption{Gesetzt: true, Wert: 5}},
		{Count: model.WahrOption{Gesetzt: true, Wert: true}},
	}
	n := 0
	for i, o := range optionen {
		for _, mitFilter := range []bool{false, true} {
			n++
			if mitFilter {
				o.Filter = idGroesser("1")
			} else {
				o.Filter = nil
			}
			u := leseUmg(t)
			u.anzahl(o)
			if _, err := u.leseGET(); err != nil {
				t.Errorf("Option %d, Filter=%v: %v", i, mitFilter, err)
				continue
			}
			zahl := u.letzteAnzahl()
			v := leseUmg(t)
			v.menge(model.Optionen{Filter: o.Filter, Count: model.WahrOption{Gesetzt: true, Wert: true}})
			if _, err := v.leseGET(); err != nil {
				t.Fatal(err)
			}
			if soll := v.letzteCollection().Anzahl.Wert; zahl != soll || zahl != map[bool]int64{false: 3, true: 2}[mitFilter] {
				t.Errorf("Option %d, Filter=%v: Zahl %d, Collection %d", i, mitFilter, zahl, soll)
			}
		}
	}
	if n != 10 {
		t.Errorf("%d Läufe, erwartet 10", n)
	}
}

// pflichtMenge legt Einträge in ein Set mit Coverage complete und ruft die Collection.
func pflichtMenge(t *testing.T, set string, art model.RessourcenArt, eintraege ...model.EntityState) (*umg, error) {
	t.Helper()
	u := leseUmg(t)
	u.co.setze(baselineID, set, model.CoverageComplete)
	u.en.setze(set, eintraege...)
	u.ressource(art, set, model.Optionen{})
	_, err := u.leseGET()
	return u, err
}

// TestSchluesselProperties hält L09 (SPEC-190 Nr. 3 f, SPEC-196 Nr. 4 c): jede Entity trägt alle Keys in Properties; fehlt eine,
// ist sie `null` oder `redigiert`, endet der Aufruf mit ODR-E9001. Typ mit 1 und mit 2 Schlüsseln × Schlüssel × Verstoß 3
// (fehlt, null, Marker) = 9, dazu der Sollfall je Typ mit 200.
func TestSchluesselProperties(t *testing.T) {
	typen := []struct {
		set  string
		keys []string
	}{{"Sks", []string{"K"}}, {"Pairs", []string{"A", "B"}}}
	payload := func(keys []string, schlecht, wie string) string {
		var teile []string
		for _, k := range keys {
			switch {
			case k != schlecht:
				teile = append(teile, `"`+k+`":"x"`)
			case wie == "null":
				teile = append(teile, `"`+k+`":null`)
			case wie == "Marker":
				teile = append(teile, `"`+k+`":"`+model.Marker+`"`)
			}
		}
		return "{" + strings.Join(teile, ",") + "}"
	}
	n := 0
	for _, ty := range typen {
		_, err := pflichtMenge(t, ty.set, model.RessourceMenge, eintrag(ty.set, "'x'", payload(ty.keys, "", ""), "K", "A", "B"))
		if err != nil {
			t.Errorf("%s Sollfall: %v", ty.set, err)
		}
		for _, k := range ty.keys {
			for _, wie := range []string{"fehlt", "null", "Marker"} {
				n++
				u, err := pflichtMenge(t, ty.set, model.RessourceMenge, eintrag(ty.set, "'x'", payload(ty.keys, k, wie), "K", "A", "B"))
				_ = erwarteFehler(t, fmt.Sprintf("%s %s %s", ty.set, k, wie), err, model.CodeE9001)
				if u.p.anzahl("Schreiber.SchreibeCollection") != 0 {
					t.Errorf("%s %s %s: der Schreiber wurde gerufen", ty.set, k, wie)
				}
			}
		}
	}
	if n != 9 {
		t.Errorf("%d Läufe, erwartet 9", n)
	}
}

// TestSchluesselEindeutig hält L10 (SPEC-190 Nr. 3 e, SPEC-196 Nr. 4 d): `Schluessel` ist nichtleer und je Entity eindeutig; ein
// Port, der einen leeren oder doppelten Schlüssel liefert, ergibt ODR-E9001, kein stilles Zusammenführen. Verstoß 2 × Art 2.
func TestSchluesselEindeutig(t *testing.T) {
	for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceAnzahl} {
		leer := produkt(2, "b")
		leer.EntityKey = ""
		doppelt := produkt(1, "b")
		for name, es := range map[string]model.EntityState{"leer": leer, "doppelt": doppelt} {
			_, err := pflichtMenge(t, setProducts, art, produkt(1, "a"), es)
			_ = erwarteFehler(t, fmt.Sprintf("Art %d, %s", art, name), err, model.CodeE9001)
		}
	}
}

// redigiertStellen sind die neun Stellen von M15: `$filter` sechs Stellen des Baums, `$orderby` drei Elemente; Name steht
// jeweils an der genannten Stelle.
func redigiertStellen() []struct {
	name string
	o    model.Optionen
} {
	id := func(op model.VergleichOp, n string) model.FilterAusdruck {
		return vergleich(op, "Id", model.Literal{Art: model.LiteralZahl, Text: n})
	}
	name := nameGleich("a")
	// Jede Stelle trägt zusätzlich eine unbekannte Property: fehlt die Prüfung des Dienstes an der Stelle, meldet der Interpreter
	// ODR-E6005 statt ODR-E9001 (sonst endet der Aufruf an einem Wert `redigiert` in beiden Fällen als ODR-E9001).
	f := func(a model.FilterAusdruck) model.Optionen {
		return model.Optionen{Filter: filterOpt(model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: a, Rechts: vergleich(model.VergleichEq, "Unbekannt", model.Literal{Art: model.LiteralZahl, Text: "1"})})}
	}
	ord := func(namen ...string) model.Optionen {
		var o model.Optionen
		for _, n := range append(namen, "Unbekannt") {
			o.OrderBy = append(o.OrderBy, model.Sortierung{Name: n})
		}
		return o
	}
	return []struct {
		name string
		o    model.Optionen
	}{
		{"Vergleich links", f(name)},
		{"Vergleich rechts", f(model.FilterVergleich{Op: model.VergleichEq, Links: model.FilterLiteral{Literal: model.Literal{Art: model.LiteralZeichenkette, Text: "'a'"}}, Rechts: model.FilterEigenschaft{Name: "Name"}})},
		{"Verknüpfung links", f(model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: name, Rechts: id(model.VergleichGt, "0")})},
		{"Verknüpfung rechts", f(model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: id(model.VergleichGt, "0"), Rechts: name})},
		{"unter not", f(model.FilterNicht{Operand: name})},
		{"Tiefe 3", f(model.FilterNicht{Operand: model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: id(model.VergleichGt, "0"),
			Rechts: model.FilterVerknuepfung{Op: model.VerknuepfungOder, Links: id(model.VergleichLt, "9"), Rechts: name}}})},
		{"orderby erstes", ord("Name", "Id", "Count")}, {"orderby mittleres", ord("Id", "Name", "Count")}, {"orderby letztes", ord("Id", "Count", "Name")},
	}
}

// TestRedigiertInOptionen hält L11 (SPEC-190 Nr. 3 g, SPEC-196 Nr. 4 e): ein Wert `redigiert` in einer Property, die `$filter`
// (an jeder Stelle des Baums) oder `$orderby` (jedes Element) nennt, ist ODR-E9001 vor der Auswertung, auch bei nur einer
// Entity unter vielen. M15: 9 Stellen.
func TestRedigiertInOptionen(t *testing.T) {
	stellen := redigiertStellen()
	for _, s := range stellen {
		u := leseUmg(t)
		u.en.setze(setProducts,
			eintrag(setProducts, "1", `{"Id":1,"Name":"a","Count":1}`, alleNamen()...),
			eintrag(setProducts, "2", `{"Id":2,"Name":"b","Count":2}`, alleNamen()...),
			eintrag(setProducts, "3", `{"Id":3,"Name":"`+model.Marker+`","Count":3}`, alleNamen()...))
		u.menge(s.o)
		_, err := u.leseGET()
		_ = erwarteFehler(t, s.name, err, model.CodeE9001)
		if u.p.anzahl("Schreiber.SchreibeCollection") != 0 {
			t.Errorf("%s: der Schreiber wurde gerufen", s.name)
		}
	}
	if len(stellen) != 9 {
		t.Errorf("%d Stellen, erwartet 9", len(stellen))
	}
	// Dieselbe Entity mit dem Wert in einer Property, die keine Option nennt: 200 (L12).
	u := leseUmg(t)
	u.en.setze(setProducts, eintrag(setProducts, "1", `{"Id":1,"Name":"`+model.Marker+`","Count":1}`, alleNamen()...))
	u.menge(model.Optionen{OrderBy: []model.Sortierung{{Name: "Count"}}})
	if _, err := u.leseGET(); err != nil {
		t.Errorf("redigiert außerhalb der Optionen: %v", err)
	}
}

// TestRedigiertNichtAusgeliefert hält L12 (SPEC-196 Nr. 4 e): ein Wert `redigiert` in einer Property außerhalb der Optionen ist
// 200, der Wert steht nicht in der Ausgabe (Collection, Entity). Art 2 × Typ der Property 6.
func TestRedigiertNichtAusgeliefert(t *testing.T) {
	n := 0
	for _, edm := range []string{"Edm.String", "Edm.Int64", "Edm.Decimal", "Edm.Guid", "Edm.Date", "Edm.DateTimeOffset"} {
		for _, art := range []model.RessourcenArt{model.RessourceMenge, model.RessourceEntity} {
			n++
			u := leseUmg(t)
			name := propJeTyp(edm)
			u.en.setze(setProducts, eintrag(setProducts, "1", `{"Id":1,"`+name+`":"`+model.Marker+`","Count":7}`, alleNamen()...))
			u.standard(art, "1")
			if _, err := u.leseGET(); err != nil {
				t.Errorf("%s Art %d: %v", edm, art, err)
				continue
			}
			var e model.Entity
			if art == model.RessourceMenge {
				e = u.letzteCollection().Entities[0]
			} else {
				e = u.letzteEntity().Entity
			}
			if _, da := wertVon(e, name); da || !slices.Contains(strings.Split(namenVon(e), ","), "Count") {
				t.Errorf("%s Art %d: Properties %s", edm, art, namenVon(e))
			}
		}
	}
	if n != 12 {
		t.Errorf("%d Läufe, erwartet 12", n)
	}
}
