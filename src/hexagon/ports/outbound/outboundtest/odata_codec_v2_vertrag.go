package outboundtest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ODataCodecV2Vertrag prüft die Zusagen von SPEC-172 Nr. 1 bis 5 (Zeilen 1 bis 3, G1 bis G21, H1 bis H10, N1 bis N6, W1 bis
// W11, K1 bis K3, R1 und R2 der Tabelle Zusage · Fänger) gegen jeden Adapter von outbound.ODataCodec für die Version 2.0. Die
// Hilfen der Suite für Version 4 (jsuite, Ströme, Generatoren) sind dieselben; die Modelle und die Bodys sind die von V2.
func ODataCodecV2Vertrag(t *testing.T, neu ODataCodecNeu) {
	zeile, stop := speicherwaechter()
	defer stop()
	defer debug.SetMaxStack(debug.SetMaxStack(256 << 20))
	s := &jsuite{neu: neu, modelle: v2Modelle(), faelle: v2AlleFaelle()}
	for _, z := range []struct {
		name string
		lauf func(*testing.T, *jsuite)
	}{
		{"G1 Argumente", v2G1}, {"G2 Version des Modells", v2G2}, {"G3 Prüffolge der Argumente", v2G3}, {"G4 entitySet", v2G4},
		{"G5 Kontext", v2G5}, {"G6 Fehler von src", v2G6}, {"G7 Zerlegung in Reads", v2G7}, {"G8 Strom gehört dem Aufrufer", jA5},
		{"G9 Grammatik", v2G9}, {"G10 Kodierung", v2G10}, {"G11 Tiefe", v2G11}, {"G12 Größe", v2G12}, {"G13 doppelte Namen", v2G13},
		{"G14 Rangfolge im Lesen", v2G14}, {"G15 Fehlerform und Ort", v2G15}, {"G16 Ursachen", v2G16}, {"G17 Wiederholung", v2G17},
		{"G18 Ergebnis gehört dem Aufrufer", v2G18}, {"G19 Port, kein Typ der Bibliothek", v2G19},
		{"G20 gleichzeitig, geordnet", v2G20}, {"G21 gleichzeitig, unabhängig", v2G21},
		{"H1 Wurzel", v2H1}, {"H2 d fehlt, weitere Mitglieder", v2H2}, {"H3 d als Array", v2H3}, {"H4 d als Objekt", v2H4},
		{"H5 leeres Ergebnis", v2H5}, {"H6 __count", v2H6}, {"H7 __next", v2H7}, {"H8 ParseEntity", v2H8}, {"H9 die Methode entscheidet", v2H9},
		{"H10 Rangfolge im Deuten", v2H10},
		{"N1 __metadata", v2N1}, {"N2 andere Mitglieder von __metadata", v2N2}, {"N3 Properties", v2N3},
		{"N4 nicht unterstützter Typ", v2N4}, {"N5 Navigation", v2N5}, {"N6 andere Namen", v2N6},
		{"W1 Typ gegen JSON-Typ", v2W1}, {"W2 String und Boolean", v2W2}, {"W3 Int32", v2W3}, {"W4 Int64", v2W4}, {"W5 Decimal", v2W5},
		{"W6 zahl_zu_lang", v2W6}, {"W7 Guid", v2W7}, {"W8 DateTime", v2W8}, {"W9 Form des Datums", v2W9},
		{"W10 Schreibweisen des Schrägstrichs", v2W10}, {"W11 null und Marker", v2W11},
		{"K1 Schlüsselliteral", v2K1}, {"K2 Verletzungen des Schlüssels", v2K2}, {"K3 Gleichheit der Schlüssel", v2K3},
		{"R1 Echtfälle", v2R1}, {"R2 Echtfälle mit Verstoß", v2R2},
	} {
		zeile(z.name)
		t.Run(z.name, func(t *testing.T) {
			vor := s.optional.Load()
			z.lauf(t, s)
			if n := s.optional.Load() - vor; n != 0 {
				t.Errorf("%s: %d Aufrufe, in denen der Leser eine optionale Schnittstelle von src ruft (Menge O)", z.name, n)
			}
		})
	}
}

// v2AlleFaelle sind die Fälle aller Zeilen mit Bodys: gültige und fehlerhafte; G7, G8, G16, G20 und G21 laufen über sie. Die
// Generatoren von G9, G10 und W8 stehen nicht darin (ihre Mengen fährt gFahre).
func v2AlleFaelle() []jfall {
	return verbinde(v2Grund(), v2FaelleG9(), v2FaelleG10(), v2FaelleG11(), v2FaelleG12(), v2FaelleG13(), v2FaelleG14(), v2FaelleH(),
		v2FaelleN(), v2FaelleW(), v2FaelleW8(), v2FaelleK(), v2FaelleEcht(), v2FaelleG15())
}

// Der Body der Einzel-Entity und der Collection von R (Produkte), ohne Zeilenumbruch am Ende.
const (
	v2EinR  = `{"d":{"__metadata":{"type":"Demo.Product"},"ID":1,"Name":"Milk"}}`
	v2KollR = `{"d":{"__count":"3","results":[{"ID":1,"Name":"Milk"},{"ID":2,"Name":null},{"ID":3,"Name":"Bread"}],"__next":"Products?$skip=3"}}`
)

// v2Basis ist der gültige Body von R für die Methode.
func v2Basis(einzel bool) jfall {
	if einzel {
		return jfall{name: "E1", m: "R", set: "Products", einzel: true, body: []byte(v2EinR)}
	}
	return jfall{name: "C1", m: "R", set: "Products", body: []byte(v2KollR)}
}

// v2Grund sind die Bodys von R und die Fälle eines unbekannten Sets.
func v2Grund() []jfall {
	out := []jfall{v2Basis(true), v2Basis(false)}
	for _, einzel := range []bool{true, false} {
		for _, set := range []string{"Nope", "products", "Demo.Products"} {
			b := v2Basis(einzel)
			b.name, b.set, b.ursache, b.ort = "G4 Set "+set, set, model.DatenEntitySet, set
			out = append(out, b)
		}
	}
	return out
}

// v2Kol ist der Body einer Collection in der Objektform, v2KolA in der Array-Form, v2Ein der einer Entity.
func v2Kol(elems ...string) []byte {
	return []byte(`{"d":{"results":[` + strings.Join(elems, ",") + `]}}`)
}

func v2KolA(elems ...string) []byte { return []byte(`{"d":[` + strings.Join(elems, ",") + `]}`) }

func v2Ein(members string) []byte { return []byte(`{"d":{` + members + `}}`) }

// v2Beide bildet aus den Mitgliedern einer Entity einen Fall je Form des Bodys: als einziges Element von `results`, als einziges
// Element des Arrays `d` und als Entity `d`. ort ist der Pointer in der Entity.
func v2Beide(name, m, set, members string, u model.DatenUrsache, ort string) []jfall {
	return []jfall{
		{name: name + " (Collection)", m: m, set: set, body: v2Kol("{" + members + "}"), ursache: u, ort: ortVor(u, "/d/results/0"+ort)},
		{name: name + " (Array)", m: m, set: set, body: v2KolA("{" + members + "}"), ursache: u, ort: ortVor(u, "/d/0"+ort)},
		{name: name + " (Entity)", m: m, set: set, einzel: true, body: v2Ein(members), ursache: u, ort: ortVor(u, "/d"+ort)},
	}
}

// v2BeideGut ist v2Beide für einen gültigen Fall; props sind die erwarteten Properties der Entity.
func v2BeideGut(name, m, set, members string, props ...model.PropertyWert) []jfall {
	fs := v2Beide(name, m, set, members, "", "")
	for i := range fs {
		fs[i].props = props
	}
	return fs
}

// v2Roh bildet aus einem Body, dessen Verstoß von der Methode unabhängig ist, einen Fall je Methode (Modell R, Set Products).
func v2Roh(name, body string, u model.DatenUrsache, ort string) []jfall {
	return []jfall{
		{name: name + " (Collection)", m: "R", set: "Products", body: []byte(body), ursache: u, ort: ort},
		{name: name + " (Entity)", m: "R", set: "Products", einzel: true, body: []byte(body), ursache: u, ort: ort},
	}
}

// v2Dt ist der JSON-String eines Zeitpunkts in der Schreibweise von V2: `"\/Date(<ms>)\/"` mit Backslash und Schrägstrich.
func v2Dt(ms string) string { return "\"\x5c/Date(" + ms + ")\x5c/\"" }

// v2Version ist das Modell mit der Version v und sonst gleichem Inhalt.
func v2Version(m model.ServiceModel, v string) model.ServiceModel {
	return model.NeuesServiceModel(v, m.Namespaces(), m.EntityTypes(), m.EntitySets(), m.NavigationBindings())
}

// v2Typen sind die sieben EDM-Typen des Modells, die der Leser liest (SPEC-170 Nr. 4).
func v2Typen() []string {
	return []string{"Edm.String", "Edm.Boolean", "Edm.Int32", "Edm.Int64", "Edm.Decimal", "Edm.Guid", "Edm.DateTimeOffset"}
}

// v2Uebersprungen sind die acht Typen, die ein Modell von V2 nennen kann und die der Leser überspringt: kein Typ der sieben
// (SPEC-170 Nr. 4) und `Edm.Date`, der trotz Unterstuetzt() kein Typ von V2 ist (SPEC-172 Nr. 3).
func v2Uebersprungen() []string {
	return []string{"Edm.Int16", "Edm.Double", "Edm.Binary", "Edm.Time", "Edm.Single", "RefScenario.c_Location", "Edm.DateTimeOffset(V2)", "Edm.Date"}
}

func v2Modelle() map[string]model.ServiceModel {
	return map[string]model.ServiceModel{"R": v2ModellR(), "W": v2ModellW(), "T": v2ModellT(), "Z": v2ModellZ(), "M": v2ModellM(), "Y": v2ModellY()}
}

// v2ModellY ist ein Modell, wie `model.NeuesServiceModel` es annimmt und kein Leser von CSDL es liefert: die Ketten von
// `BaseType` schließen sich. Der Typ des Sets `Ss` ist `Y.S` ohne Basistyp; `Y.Self` ist sein eigener Basistyp, `Y.R1` und `Y.R2`
// sind ein Ring über zwei, `Y.Q1` bis `Y.Q3` ein Ring über drei Typen, `Y.T` hängt mit seinem Basistyp an einem Ring.
func v2ModellY() model.ServiceModel {
	id := p("Id", "Edm.Int32", false)
	typ := func(name, basis string) model.EntityType {
		return model.EntityType{Name: name, BaseType: basis, Keys: []string{"Id"}, Properties: []model.Property{id}}
	}
	typen := []model.EntityType{typ("Y.S", ""), typ("Y.Self", "Y.Self"), typ("Y.R1", "Y.R2"), typ("Y.R2", "Y.R1"),
		typ("Y.Q1", "Y.Q2"), typ("Y.Q2", "Y.Q3"), typ("Y.Q3", "Y.Q1"), typ("Y.T", "Y.R1")}
	return model.NeuesServiceModel("2.0", []string{"Y"}, typen, []model.EntitySet{{Name: "Ss", EntityType: "Y.S"}}, nil)
}

// v2ModellR ist das Modell der Echtfälle (Olingo-Referenzszenario nach 03-metadata) und der Produkte der Strom-Zeilen.
func v2ModellR() model.ServiceModel {
	emp := model.EntityType{Name: "RefScenario.Employee", Keys: []string{"EmployeeId"}, Properties: []model.Property{
		p("EmployeeId", "Edm.String", false), p("EmployeeName", "Edm.String", true), p("ManagerId", "Edm.String", true),
		p("RoomId", "Edm.String", true), p("TeamId", "Edm.String", true), p("Location", "RefScenario.c_Location", true),
		p("Age", "Edm.Int16", true), p("EntryDate", "Edm.DateTimeOffset", true), p("ImageUrl", "Edm.String", true)},
		NavigationProperties: []model.NavigationProperty{{Name: "ne_Manager", Target: "RefScenario.Employee"},
			{Name: "ne_Team", Target: "RefScenario.Team"}, {Name: "ne_Room", Target: "RefScenario.Room"}}}
	room := model.EntityType{Name: "RefScenario.Room", Keys: []string{"Id"}, Properties: []model.Property{
		p("Id", "Edm.String", false), p("Name", "Edm.String", true), p("Seats", "Edm.Int16", true), p("Version", "Edm.Int16", true)},
		NavigationProperties: []model.NavigationProperty{{Name: "nr_Employees", Target: "RefScenario.Employee", Collection: true},
			{Name: "nr_Building", Target: "RefScenario.Building"}}}
	team := model.EntityType{Name: "RefScenario.Team", Keys: []string{"Id"}, Properties: []model.Property{
		p("Id", "Edm.String", false), p("Name", "Edm.String", true), p("isScrumTeam", "Edm.Boolean", true)},
		NavigationProperties: []model.NavigationProperty{{Name: "nt_Employees", Target: "RefScenario.Employee", Collection: true}}}
	prod := model.EntityType{Name: "Demo.Product", Keys: []string{"ID"}, Properties: []model.Property{
		p("ID", "Edm.Int32", false), p("Name", "Edm.String", true)}}
	return model.NeuesServiceModel("2.0", []string{"RefScenario", "Demo"}, []model.EntityType{emp, room, team, prod},
		[]model.EntitySet{{Name: "Employees", EntityType: "RefScenario.Employee"}, {Name: "Rooms", EntityType: "RefScenario.Room"},
			{Name: "Teams", EntityType: "RefScenario.Team"}, {Name: "Products", EntityType: "Demo.Product"}}, nil)
}

// v2ModellW hat den Typ `T.W` mit dem Schlüssel `Id` und je Typ der sieben zwei Properties, `N<Typ>` (nullable) und `R<Typ>`
// (nicht nullable), dazu `X<i>` je Typ von v2Uebersprungen und die Navigationen `Nav` und `Navs` (Set `Ws`); `T.WO` ist offen
// (Set `Wo`); `T.WN` trägt die Properties `results`, `__count` und `__next` (Set `Wn`).
func v2ModellW() model.ServiceModel {
	props := []model.Property{p("Id", "Edm.Int32", false)}
	for _, edm := range v2Typen() {
		props = append(props, p("N"+kurz(edm), edm, true), p("R"+kurz(edm), edm, false))
	}
	for i, edm := range v2Uebersprungen() {
		props = append(props, p("X"+strconv.Itoa(i), edm, true))
	}
	navs := []model.NavigationProperty{{Name: "Nav", Target: "T.W"}, {Name: "Navs", Target: "T.W", Collection: true}}
	typen := []model.EntityType{
		{Name: "T.W", Keys: []string{"Id"}, Properties: props, NavigationProperties: navs},
		{Name: "T.WO", Keys: []string{"Id"}, Open: true, Properties: props, NavigationProperties: navs},
		{Name: "T.WN", Keys: []string{"Id"}, Properties: []model.Property{p("Id", "Edm.Int32", false), p("results", "Edm.String", true),
			p("__count", "Edm.Int32", true), p("__next", "Edm.String", true)}},
	}
	sets := []model.EntitySet{{Name: "Ws", EntityType: "T.W"}, {Name: "Wo", EntityType: "T.WO"}, {Name: "Wn", EntityType: "T.WN"}}
	return model.NeuesServiceModel("2.0", []string{"T"}, typen, sets, nil)
}

// v2ModellT hat je Typ der sieben einen Entity-Typ mit einem Schlüssel `Id` dieses Typs (Set `K<Typ>`), dazu die Typen mit zwei
// und drei Schlüsseln (Sets `K2`, `K3`, `KS`) und einen Typ mit einem nullbaren Schlüssel (`KN`).
func v2ModellT() model.ServiceModel {
	var typen []model.EntityType
	var sets []model.EntitySet
	for _, edm := range v2Typen() {
		n := "T.K" + kurz(edm)
		typen = append(typen, model.EntityType{Name: n, Keys: []string{"Id"}, Properties: []model.Property{p("Id", edm, false)}})
		sets = append(sets, model.EntitySet{Name: "K" + kurz(edm), EntityType: n})
	}
	typen = append(typen,
		model.EntityType{Name: "T.K2", Keys: []string{"A", "B"}, Properties: []model.Property{
			p("A", "Edm.Int32", false), p("B", "Edm.String", false), p("Note", "Edm.String", true)}},
		model.EntityType{Name: "T.K3", Keys: []string{"A", "B", "C"}, Properties: []model.Property{
			p("A", "Edm.Int32", false), p("B", "Edm.String", false), p("C", "Edm.Boolean", false), p("Note", "Edm.String", true)}},
		model.EntityType{Name: "T.KN", Keys: []string{"Id"}, Properties: []model.Property{p("Id", "Edm.Int32", true)}},
		model.EntityType{Name: "T.KS", Keys: []string{"A", "B"}, Properties: []model.Property{p("A", "Edm.String", false), p("B", "Edm.String", false)}},
	)
	sets = append(sets, model.EntitySet{Name: "K2", EntityType: "T.K2"}, model.EntitySet{Name: "K3", EntityType: "T.K3"},
		model.EntitySet{Name: "KS", EntityType: "T.KS"}, model.EntitySet{Name: "KN", EntityType: "T.KN"})
	return model.NeuesServiceModel("2.0", []string{"T"}, typen, sets, nil)
}

// v2ModellZ hat die Ableitung `Z.A` ← `Z.A1` ← `Z.A2`, den abstrakten `Z.AbsD` unter `Z.A` und den fremden Typ `Z.Other` (Set
// `As`, Typ `Z.A`), dazu den abstrakten Basistyp ohne Schlüssel `Z.B` mit der Ableitung `Z.B1` (Set `Bs`, Typ `Z.B`).
func v2ModellZ() model.ServiceModel {
	id, name := p("Id", "Edm.Int32", false), p("Name", "Edm.String", true)
	typen := []model.EntityType{
		{Name: "Z.A", Keys: []string{"Id"}, Properties: []model.Property{id, name}},
		{Name: "Z.A1", BaseType: "Z.A", Keys: []string{"Id"}, Properties: []model.Property{id, name, p("E1", "Edm.Int32", true)}},
		{Name: "Z.A2", BaseType: "Z.A1", Keys: []string{"Id"}, Properties: []model.Property{id, name, p("E1", "Edm.Int32", true), p("E2", "Edm.String", true)}},
		{Name: "Z.AbsD", BaseType: "Z.A", Abstract: true, Keys: []string{"Id"}, Properties: []model.Property{id, name}},
		{Name: "Z.Other", Keys: []string{"Id"}, Properties: []model.Property{id, name}},
		{Name: "Z.B", Abstract: true, Properties: []model.Property{name}},
		{Name: "Z.B1", BaseType: "Z.B", Keys: []string{"Id"}, Properties: []model.Property{name, id}},
	}
	sets := []model.EntitySet{{Name: "As", EntityType: "Z.A"}, {Name: "Bs", EntityType: "Z.B"}}
	return model.NeuesServiceModel("2.0", []string{"Z"}, typen, sets, nil)
}

// v2Namen40 ist die Zahl der Properties und der übersprungenen Namen des Typs von G17.
const v2Namen40 = 40

// v2ModellM hat den Typ `M.T` mit dem Schlüssel `Id`, 40 Properties `P00` bis `P39` (Edm.Int32) und 40 Properties `U00` bis
// `U39` nicht unterstützten Typs (Set `Ms`): mehr Namen, als eine lineare Suche bevorzugt.
func v2ModellM() model.ServiceModel {
	props := []model.Property{p("Id", "Edm.Int32", false)}
	for i := range v2Namen40 {
		props = append(props, p(fmt.Sprintf("P%02d", i), "Edm.Int32", true), p(fmt.Sprintf("U%02d", i), "Edm.Double", true))
	}
	return model.NeuesServiceModel("2.0", []string{"M"}, []model.EntityType{{Name: "M.T", Keys: []string{"Id"}, Properties: props}},
		[]model.EntitySet{{Name: "Ms", EntityType: "M.T"}}, nil)
}

// v2G15 prüft die Form des Fehlers (SPEC-126 Nr. 14 mit SPEC-172 Nr. 2): Typ, Code, errors.Is, Ort als JSON-Pointer in den drei
// Formen des Bodys, Error() mit dem Ort in der Schreibweise von strconv.Quote und ohne den Wert des Dokuments.
func v2G15(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleG15())
	for _, f := range v2FaelleG15() {
		r := s.lauf(f)
		var df *model.DatenFehler
		if !errors.As(r.err, &df) {
			continue
		}
		soll := "ODR-E6101: Baseline-Import ungültig oder inkonsistent: " + string(df.Ursache)
		if df.Ort != "" {
			soll += " " + strconv.Quote(df.Ort)
		}
		if r.err.Error() != soll {
			t.Errorf("%s: Error() %q, erwartet %q", f.name, r.err.Error(), soll)
		}
		if strings.Contains(r.err.Error(), "GEHEIMWERT") {
			t.Errorf("%s: Error() trägt den Wert des Dokuments: %q", f.name, r.err.Error())
		}
	}
}

// v2NameJSON schreibt den Namen als JSON-String: `"` und `\` mit Backslash, jedes Zeichen unter U+0020 als Escape.
func v2NameJSON(name string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, c := range []byte(name) {
		switch {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c < 0x20:
			b.WriteString(esc(fmt.Sprintf("%04x", c)))
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// v2Zeiger ist der Pointer-Teil eines Namens (RFC 6901): `~` zu `~0`, `/` zu `~1`, sonst nichts.
func v2Zeiger(name string) string { return strings.NewReplacer("~", "~0", "/", "~1").Replace(name) }

// v2FaelleG15: Orte mit Namen, die der Pointer ersetzt oder nicht ersetzt, in den drei Formen des Bodys, dazu die Orte unter
// `__metadata` und der fehlende Schlüssel.
func v2FaelleG15() []jfall {
	var out []jfall
	namen := []string{"a~b", "a/b", "a\x5cb", "a\"b", "a\x01b", "~", "/", "~0", "~1", "a~1b/c~0", "\xc3\xa9", "", "\x7f", "a b", "\x1f"}
	for _, n := range namen {
		out = append(out, v2Beide("G15 Name "+strconv.Quote(n), "R", "Products", `"ID":1,`+v2NameJSON(n)+`:1`, model.DatenUnbekannteProperty, "/"+v2Zeiger(n))...)
	}
	out = append(out,
		v2Beide("G15 Name mit Wert", "R", "Products", `"ID":"GEHEIMWERT"`, model.DatenTyp, "/ID")...)
	out = append(out,
		v2Beide("G15 fehlender Schlüssel", "R", "Products", `"Name":"x"`, model.DatenSchluessel, "/ID")...)
	for _, f := range []struct{ name, kopf, ort string }{
		{"Collection", `{"d":{"results":[{"__metadata":{"type":"Nope"},"ID":1}]}}`, "/d/results/0/__metadata/type"},
		{"Array", `{"d":[{"__metadata":{"type":"Nope"},"ID":1}]}`, "/d/0/__metadata/type"},
		{"Entity", `{"d":{"__metadata":{"type":"Nope"},"ID":1}}`, "/d/__metadata/type"},
	} {
		out = append(out, jfall{name: "G15 __metadata/type " + f.name, m: "R", set: "Products", einzel: f.name == "Entity", body: []byte(f.kopf),
			ursache: model.DatenODataTyp, ort: f.ort})
	}
	return out
}

// v2G16: jede der 14 Ursachen ohne `kontext` kommt in einem Fall der Suite vor, `kontext` nie (SPEC-172 Nr. 1).
func v2G16(t *testing.T, s *jsuite) {
	soll := []model.DatenUrsache{model.DatenSyntax, model.DatenKodierung, model.DatenTiefe, model.DatenZuGross, model.DatenDoppelterName,
		model.DatenStruktur, model.DatenODataTyp, model.DatenUnbekannteProperty, model.DatenTyp, model.DatenWert, model.DatenNull,
		model.DatenZahlZuLang, model.DatenSchluessel, model.DatenEntitySet}
	zaehler := map[model.DatenUrsache]int{}
	for _, f := range s.alleFaelle(t) {
		if f.ursache == "" {
			continue
		}
		var df *model.DatenFehler
		if errors.As(s.lauf(f).err, &df) {
			zaehler[df.Ursache]++
		}
	}
	for _, u := range soll {
		if zaehler[u] == 0 {
			t.Errorf("G16: die Ursache %s kommt in keinem Fall vor", u)
		}
	}
	if zaehler[model.DatenKontext] != 0 {
		t.Errorf("G16: die Ursache kontext kommt %d-mal vor, in V2 nie", zaehler[model.DatenKontext])
	}
	if len(zaehler) != len(soll) {
		t.Errorf("G16: %d verschiedene Ursachen, erwartet %d: %v", len(zaehler), len(soll), zaehler)
	}
	t.Logf("G16: %d Ursachen, %v", len(zaehler), zaehler)
}

// v2G17: gleiche Bytes und Modell ergeben gleiche Werte; die Reihenfolge ist die des Dokuments, nie die einer Map (SPEC-126
// Nr. 15). Das Dokument nennt 40 Properties und 40 übersprungene Namen in einer Reihenfolge, die die des Modells nicht ist.
func v2G17(t *testing.T, s *jsuite) {
	var members []string
	var props []model.PropertyWert
	var ueber []model.Uebersprungen
	for j := range v2Namen40 {
		// 7 und 9 sind teilerfremd zu 40: jeder Name steht genau einmal, in einer Reihenfolge, die die des Modells nicht ist.
		i, u := (j*7)%v2Namen40, (j*9)%v2Namen40
		members = append(members, fmt.Sprintf(`"U%02d":%d.5`, u, u), fmt.Sprintf(`"P%02d":%d`, i, i))
		ueber = append(ueber, model.Uebersprungen{Name: fmt.Sprintf("U%02d", u), Grund: model.UebersprungTyp})
		props = append(props, pw(fmt.Sprintf("P%02d", i), model.ArtInt32, strconv.Itoa(i)))
		if j == v2Namen40/2 {
			members = append(members, `"Id":9`)
			props = append(props, pw("Id", model.ArtInt32, "9"))
		}
	}
	f := jfall{name: "G17", m: "M", set: "Ms", einzel: true, body: v2Ein(strings.Join(members, ",")), props: props, ueber: ueber, schluessel: "9"}
	erste := s.lauf(f)
	for i := range 50 {
		r := s.lauf(f)
		if !reflect.DeepEqual(r, erste) {
			t.Fatalf("G17 Wiederholung %d: Ergebnis weicht ab", i)
		}
		f.pruefeErste(t, []model.Entity{r.entity})
		if t.Failed() {
			return
		}
	}
	var elems []string
	for i := range 40 {
		elems = append(elems, fmt.Sprintf(`{"Id":%d}`, 39-i))
	}
	r := s.lauf(jfall{m: "M", set: "Ms", body: v2Kol(elems...)})
	for i, e := range r.seite.Entities {
		if e.Schluessel != strconv.Itoa(39-i) {
			t.Fatalf("G17: Element %d hat den Schlüssel %s, erwartet %d", i, e.Schluessel, 39-i)
		}
	}
}

// v2G18: das Ergebnis gehört dem Aufrufer; die Mengen sind nie nil; zwei Elemente gleichen Schlüssels stehen beide im Ergebnis
// (SPEC-126 Nr. 12).
func v2G18(t *testing.T, s *jsuite) {
	f := jfall{m: "R", set: "Products", body: v2Kol(`{"ID":1,"Name":"a"}`, `{"ID":1,"Name":"b"}`)}
	a := s.lauf(f)
	if a.err != nil || len(a.seite.Entities) != 2 || a.seite.Entities[0].Schluessel != "1" || a.seite.Entities[1].Schluessel != "1" ||
		a.seite.Entities[0].Properties[1].Wert.Text != "a" || a.seite.Entities[1].Properties[1].Wert.Text != "b" {
		t.Fatalf("G18 Dubletten: %+v %v", a.seite, a.err)
	}
	a.seite.Entities[0].Properties[1].Wert.Text = "verändert"
	a.seite.Entities[0].Properties = append(a.seite.Entities[0].Properties[:1], model.PropertyWert{})
	a.seite.Entities = append(a.seite.Entities[:1], model.Entity{})
	if b := s.lauf(f); !reflect.DeepEqual(b, s.lauf(f)) || b.seite.Entities[0].Properties[1].Wert.Text != "a" || len(b.seite.Entities) != 2 {
		t.Errorf("G18: der zweite Aufruf hängt von der Änderung des ersten ab: %+v", b.seite)
	}
	v2G18Leer(t, s)
	r := s.lauf(jfall{m: "R", set: "Products", body: v2Kol(`{"ID":1}`)})
	if r.seite.Entities[0].Uebersprungen == nil {
		t.Error("G18: Uebersprungen einer Entity ist nil")
	}
}

// v2G18Leer: leere Mengen sind nie nil, weder die Entities einer Sammlung noch die Properties und die übersprungenen Namen einer
// einzelnen Entity.
func v2G18Leer(t *testing.T, s *jsuite) {
	for _, einzel := range []bool{false, true} {
		body := v2Kol()
		if einzel {
			body = v2Ein(`"ID":1`)
		}
		r := s.lauf(jfall{m: "R", set: "Products", einzel: einzel, body: body})
		if r.err != nil {
			t.Fatalf("G18: %v", r.err)
		}
		switch {
		case !einzel && (r.seite.Entities == nil || len(r.seite.Entities) != 0):
			t.Errorf("G18: leeres results liefert Entities %#v, erwartet leer und nicht nil", r.seite.Entities)
		case einzel && (r.entity.Properties == nil || r.entity.Uebersprungen == nil):
			t.Errorf("G18: Entity ohne übersprungene Namen liefert %#v", r.entity)
		}
	}
}

// v2G19: kein Typ der Bibliothek steht in einer Signatur oder einem Typ, den sie erreicht; der Leser hat kein Feld, und
// jeder Verstoß ist ODR-E6101 mit der Abbildung des Ports (SPEC-126 Nr. 14, 15).
func v2G19(t *testing.T, s *jsuite) {
	jK3(t, s)
	jK4(t, s)
	jK2(t, s)
}

// v2Parallel fährt n Goroutinen hinter einer Startbarriere über die Fälle und vergleicht jedes Ergebnis mit dem eines einzelnen
// Aufrufs; mit gemischt enden einzelne Aufrufe durch einen beendeten Kontext und einen Fehler von src.
func v2Parallel(t *testing.T, s *jsuite, faelle []jfall, n int, gemischt bool) {
	soll := make([]jergebnis, len(faelle))
	for i, f := range faelle {
		soll[i] = s.lauf(f)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for k := range faelle {
				i := (k + g*7) % len(faelle)
				f := faelle[i]
				if got := s.lauf(f); !reflect.DeepEqual(got, soll[i]) {
					t.Errorf("%s: %+v, allein gelesen %+v", f.name, got, soll[i])
					return
				}
				if gemischt && k%25 == 0 && len(f.body) > 1 {
					s.ruf(abgebrochen(), strings.NewReader(string(f.body)), f)
					s.ruf(context.Background(), &endlichStrom{inner: &teilReader{teile: [][]byte{f.body[:len(f.body)/2]}, ende: errQuelle}}, f)
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}

// v2G20: gleichzeitige Aufrufe sind geordnet: 16 Goroutinen lesen die Echtdateien und die kleinen Fälle, jedes Ergebnis ist
// gleich dem eines einzelnen Aufrufs. Der Detektor von `make test-race` meldet jeden Zugriff ohne Ordnung.
func v2G20(t *testing.T, s *jsuite) {
	var faelle []jfall
	faelle = append(faelle, v2FaelleEcht()...)
	for i, f := range j1Faelle(t, s) {
		if i%11 == 0 {
			faelle = append(faelle, f)
		}
	}
	v2Parallel(t, s, faelle, 16, false)
}

// v2G21: gleichzeitige Aufrufe sind unabhängig: gültige und ungültige Bodys zugleich, jeder Aufruf mit seiner eigenen Erwartung
// (der Vergleich mit dem einzelnen Aufruf zählt jede Vermischung), dazu beendete Kontexte und Fehler von src.
func v2G21(t *testing.T, s *jsuite) {
	jJ1(t, s)
	faelle := []jfall{v2Basis(true), v2Basis(false),
		{m: "R", set: "Products", body: v2Kol(`{"ID":1,"Name":5}`), ursache: model.DatenTyp, ort: "/d/results/0/Name"},
		{m: "R", set: "Products", einzel: true, body: []byte(`{"d":[`), ursache: model.DatenSyntax},
		{m: "W", set: "Ws", body: v2Kol(`{"Id":1,"RDecimal":"1.50"}`)},
		{m: "R", set: "Products", body: v2Kol(`{"ID":1,"ID":2}`), ursache: model.DatenDoppelterName}}
	v2Parallel(t, s, faelle, 16, true)
}
