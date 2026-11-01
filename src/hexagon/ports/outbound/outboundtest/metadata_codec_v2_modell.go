package outboundtest

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func faelleV2F(t *testing.T) []fall {
	t.Helper()
	return []fall{
		gut("Referenzdienst", festesV2(t, "v2_referenz.xml")), gut("Vererbung", festesV2(t, "v2_vererbung.xml")), gut("Composite-Schlüssel", festesV2(t, "v2_composite.xml")),
		gut("Aliase", f4DocV2([6]bool{true, true, true, true, true, true}, true)), gut("Randformen des Modells", f6DocV2()),
		gut("ohne EntityContainer", eins2(typ("T", ""))),
	}
}

// zeileV2F1: das Modell des Referenzdienstes (SPEC-171 Nr. 8).
func zeileV2F1(t *testing.T, neu CodecNeu) {
	m, b := pruefeFall(t, neu, gut("R", festesV2(t, "v2_referenz.xml")))
	if m.Version() != "2.0" {
		t.Errorf("Version %q", m.Version())
	}
	gleich(t, "Namespaces", m.Namespaces(), []string{"Demo"})
	product := model.EntityType{Name: "Demo.Product", Keys: []string{"ID"},
		Properties: []model.Property{{Name: "ID", EdmType: "Edm.Int32"}, {Name: "Name", EdmType: "Edm.String", Nullable: true},
			{Name: "Price", EdmType: "Edm.Decimal"}, {Name: "Released", EdmType: "Edm.DateTimeOffset"}, {Name: "Weight", EdmType: "Edm.Double", Nullable: true}},
		NavigationProperties: []model.NavigationProperty{{Name: "Category", Target: "Demo.Category"}}}
	category := model.EntityType{Name: "Demo.Category", Keys: []string{"ID"},
		Properties:           []model.Property{{Name: "ID", EdmType: "Edm.Guid"}, {Name: "Title", EdmType: "Edm.String", Nullable: true}},
		NavigationProperties: []model.NavigationProperty{{Name: "Products", Target: "Demo.Product", Collection: true}}}
	gleich(t, "EntityTypes", m.EntityTypes(), []model.EntityType{product, category})
	for _, e := range []model.EntityType{product, category} {
		got, ok := m.EntityType(e.Name)
		gleich(t, "EntityType "+e.Name, []any{got, ok}, []any{e, true})
	}
	gleich(t, "EntitySets", m.EntitySets(), []model.EntitySet{{Name: "Products", EntityType: "Demo.Product"}, {Name: "Categories", EntityType: "Demo.Category"}})
	gleich(t, "NavigationBindings", m.NavigationBindings(), []model.NavigationBinding{
		{Set: "Products", Path: "Category", Target: "Demo.Container/Categories"}, {Set: "Categories", Path: "Products", Target: "Demo.Container/Products"}})
	gleich(t, "Funde", b.Funde, []model.Fund{{Art: model.FundPropertyTyp, Ort: "Demo.Product/Weight", Name: "Edm.Double"}})
	if _, ok := m.EntityType("Self.Product"); ok {
		t.Error("EntityType löst einen Alias auf")
	}
	if s, ok := m.EntitySet("Products"); !ok || s.EntityType != "Demo.Product" {
		t.Errorf("EntitySet: %v %v", s, ok)
	}
}

// zeileV2F2: Vererbung im Modell: wirksame Keys, Properties und Navigation-Properties, Abstract nur am Typ, Open nie (Nr. 8).
func zeileV2F2(t *testing.T, neu CodecNeu) {
	m, _ := pruefeFall(t, neu, gut("V", festesV2(t, "v2_vererbung.xml")))
	for _, w := range []struct {
		name, base, properties, nav string
		abstract                    bool
	}{
		{"Zoo.Animal", "", "Id Name", "", true}, {"Zoo.Mammal", "Zoo.Animal", "Id Name Fur", "Keeper", false},
		{"Zoo.Dog", "Zoo.Mammal", "Id Name Fur Breed", "Keeper", false}, {"Zoo.Puppy", "Zoo.Dog", "Id Name Fur Breed", "Keeper", false},
		{"Zoo.Keeper", "", "Id", "", false},
	} {
		e, ok := m.EntityType(w.name)
		props, nav := namenV2(e)
		if !ok || e.BaseType != w.base || e.Abstract != w.abstract || e.Open || props != w.properties || nav != w.nav || strings.Join(e.Keys, " ") != "Id" {
			t.Errorf("%s: %+v", w.name, e)
		}
		for _, n := range e.NavigationProperties {
			if n.Target != "Zoo.Keeper" || n.Collection {
				t.Errorf("%s: Navigation %+v", w.name, n)
			}
		}
	}
	e, _ := m.EntityType("Zoo.Dog")
	if edmTypVon(e.Properties, 1) != edmTypVon(e.Properties, 3) {
		t.Errorf("Name und Breed sind beide Edm.String: %+v", e.Properties)
	}
	gleich(t, "Bindings", m.NavigationBindings(), []model.NavigationBinding{{Set: "Animals", Path: "Keeper", Target: "Zoo.Park/Keepers"}})
	leer, _ := pruefeFall(t, neu, gut("Ableitung mit leerem Key", eins2(typ("B", "")+ableitung("D", "Ns.B", "<Key/>"))))
	for _, n := range []string{"Ns.B", "Ns.D"} {
		if got, _ := leer.EntityType(n); !slices.Equal(got.Keys, []string{"Id"}) {
			t.Errorf("%s mit leerem Key in der Ableitung: Keys %v, erwartet [Id]", n, got.Keys)
		}
	}
	zeileV2F2Kette(t, neu)
}

// zeileV2F2Kette: eine Kette von drei Typen erbt Properties und Navigation-Properties jeder Stufe.
func zeileV2F2Kette(t *testing.T, neu CodecNeu) {
	m, _ := pruefeFall(t, neu, gut("Kette", eins2(typ("A", prop("PA", "Edm.String")+nav2("NA"))+ableitung("B", "Ns.A", prop("PB", "Edm.Int32")+nav2("NB"))+
		ableitung("C", "Ns.B", prop("PC", "Edm.Guid"))+
		`<Association Name="R"><End Type="Ns.A" Role="F" Multiplicity="*"/><End Type="Ns.A" Role="G" Multiplicity="*"/></Association>`)))
	c, _ := m.EntityType("Ns.C")
	props, navs := namenV2(c)
	if props != "Id PA PB PC" || navs != "NA NB" || !slices.Equal(c.Keys, []string{"Id"}) {
		t.Errorf("Kette: %v %v %v", props, navs, c.Keys)
	}
}

// namenV2 sind die Namen der wirksamen Properties und Navigation-Properties eines Typs, je durch Leerzeichen getrennt.
func namenV2(e model.EntityType) (props, navs string) {
	var p, n []string
	for _, x := range e.Properties {
		p = append(p, x.Name)
	}
	for _, x := range e.NavigationProperties {
		n = append(n, x.Name)
	}
	return strings.Join(p, " "), strings.Join(n, " ")
}

func zeileV2F3(t *testing.T, neu CodecNeu) {
	m, _ := pruefeFall(t, neu, gut("K", festesV2(t, "v2_composite.xml")))
	e, _ := m.EntityType("Shop.OrderLine")
	g, _ := m.EntityType("Shop.Gift")
	if strings.Join(e.Keys, " ") != "OrderID LineNo" || strings.Join(g.Keys, " ") != "OrderID LineNo" || nameVon(e.Properties, 0) != "LineNo" {
		t.Errorf("Keys %v, geerbt %v, Properties %v", e.Keys, g.Keys, e.Properties)
	}
	ordnungen := []string{"A B", "B A"}
	for _, schluessel := range ordnungen {
		for _, properties := range ordnungen {
			doc := f3DokV2(strings.Fields(schluessel), strings.Fields(properties))
			m, _ := pruefeFall(t, neu, gut("Key "+schluessel+", Properties "+properties, doc))
			if got, _ := m.EntityType("Ns.T"); strings.Join(got.Keys, " ") != schluessel {
				t.Errorf("Key %s bei Properties %s: %v", schluessel, properties, got.Keys)
			}
		}
	}
}

// f3DokV2 ist ein Typ `T` mit den PropertyRef schluessel und den Properties props in der genannten Reihenfolge.
func f3DokV2(schluessel, props []string) []byte {
	refs, ps := "", ""
	for _, k := range schluessel {
		refs += `<PropertyRef Name="` + k + `"/>`
	}
	for _, p := range props {
		ps += prop(p, "Edm.Int32")
	}
	return eins2(`<EntityType Name="T"><Key>` + refs + `</Key>` + ps + `</EntityType>`)
}

// f4DocV2 hat zwei Namensräume mit je einem Alias (NsA/A, bei zwei Schemas NsB/B), sechs Verweise (BaseType, Typ einer
// Property, Relationship, End-Typ, EntityType des Sets, Association der AssociationSet) und, bei zwei Schemas, die Typen
// `U` und die Association `R` im zweiten; alias[i] sagt, ob der Verweis i über den Alias oder den Namensraum geht.
func f4DocV2(alias [6]bool, zwei bool) []byte {
	pick := func(i int, a, ns string) string {
		if alias[i] {
			return a
		}
		return ns
	}
	ns, al := "NsA", "A"
	if zwei {
		ns, al = "NsB", "B"
	}
	u := typ("U", "") + `<Association Name="R"><End Type="` + pick(3, "A.T", "NsA.T") + `" Role="F" Multiplicity="*"/><End Type="` + pick(3, al+".U", ns+".U") +
		`" Role="G" Multiplicity="0..1"/></Association>`
	holder := `<EntityType Name="Holder">` + schluesselID + prop("P2", pick(1, "A.T", "NsA.T")) +
		`<NavigationProperty Name="N3" Relationship="` + pick(2, al+".R", ns+".R") + `" FromRole="F" ToRole="G"/></EntityType>`
	sets := `<EntityContainer Name="Container"><EntitySet Name="S1" EntityType="` + pick(4, "A.T", "NsA.T") + `"/><EntitySet Name="SU" EntityType="` + pick(4, al+".U", ns+".U") +
		`"/><AssociationSet Name="AS" Association="` + pick(5, al+".R", ns+".R") + `"><End Role="F" EntitySet="S1"/><End Role="G" EntitySet="SU"/></AssociationSet></EntityContainer>`
	a := typ("T", prop("X", "Edm.String")) + ableitung("D1", pick(0, "A.T", "NsA.T"), "") + holder
	if zwei {
		return []byte(edmx2(schemaAlias2("NsA", "A", a+sets) + schemaAlias2("NsB", "B", u) + schemaAlias2("NsA", "A2", typ("V", ""))))
	}
	return []byte(edmx2(schemaAlias2("NsA", "A", a+u+sets)))
}

// f4Soll ist die Erwartung zu f4DocV2: dieselben Namen, gleich über welchen Weg die Verweise gehen.
func f4Soll(zwei bool) (typen []model.EntityType, namespaces []string, sets []model.EntitySet, bindings []model.NavigationBinding) {
	u := "NsA"
	namespaces = []string{"NsA"}
	if zwei {
		u = "NsB"
		namespaces = append(namespaces, "NsB")
	}
	id := model.Property{Name: "Id", EdmType: "Edm.Int32", Nullable: true}
	x := model.Property{Name: "X", EdmType: "Edm.String", Nullable: true}
	typen = []model.EntityType{
		{Name: "NsA.T", Keys: []string{"Id"}, Properties: []model.Property{id, x}},
		{Name: "NsA.D1", BaseType: "NsA.T", Keys: []string{"Id"}, Properties: []model.Property{id, x}},
		{Name: "NsA.Holder", Keys: []string{"Id"}, Properties: []model.Property{id, {Name: "P2", EdmType: "NsA.T", Nullable: true}},
			NavigationProperties: []model.NavigationProperty{{Name: "N3", Target: u + ".U"}}},
	}
	if !zwei {
		typen = append(typen, model.EntityType{Name: "NsA.U", Keys: []string{"Id"}, Properties: []model.Property{id}})
	}
	if zwei {
		typen = append(typen, model.EntityType{Name: "NsB.U", Keys: []string{"Id"}, Properties: []model.Property{id}},
			model.EntityType{Name: "NsA.V", Keys: []string{"Id"}, Properties: []model.Property{id}})
	}
	sets = []model.EntitySet{{Name: "S1", EntityType: "NsA.T"}, {Name: "SU", EntityType: u + ".U"}}
	bindings = []model.NavigationBinding{{Set: "S1", Path: "N3", Target: "NsA.Container/SU"}}
	return typen, namespaces, sets, bindings
}

// zeileV2F4: jede Referenzart mit Alias oder Namensraum ergibt denselben qualifizierten Namen (SPEC-171 Nr. 7): Schleife
// über die sechs Arten mal Alias und Namensraum mal ein und zwei Schemas.
func zeileV2F4(t *testing.T, neu CodecNeu) {
	for _, zwei := range []bool{false, true} {
		typen, namespaces, sets, bindings := f4Soll(zwei)
		var varianten [][6]bool
		varianten = append(varianten, [6]bool{}, [6]bool{true, true, true, true, true, true})
		for i := range 6 {
			var v [6]bool
			v[i] = true
			varianten = append(varianten, v)
		}
		for _, v := range varianten {
			m, b := pruefeFall(t, neu, gut(fmt.Sprintf("Aliase %v, zwei Schemas %v", v, zwei), f4DocV2(v, zwei)))
			name := fmt.Sprintf("Aliase %v, zwei Schemas %v", v, zwei)
			gleich(t, name+" Namespaces", m.Namespaces(), namespaces)
			gleich(t, name+" Typen", m.EntityTypes(), typen)
			gleich(t, name+" Sets", m.EntitySets(), sets)
			gleich(t, name+" Bindings", m.NavigationBindings(), bindings)
			gleich(t, name+" Funde", b.Funde, []model.Fund{{Art: model.FundPropertyTyp, Ort: "NsA.Holder/P2", Name: "NsA.T"}})
		}
	}
	pruefeFall(t, neu, gut("Alias gleich Namespace", []byte(edmx2(schemaAlias2("NsA", "NsA", typ("T", ""))))))
}

// zeileV2F5: Edm.DateTime ergibt Edm.DateTimeOffset, die sechs anderen unterstützten Typen bleiben, die acht nicht unterstützten
// bleiben im Modell mit Unterstuetzt() falsch, Edm.DateTimeOffset des Dokuments trägt einen Namen, der keiner der acht Typen ist.
func zeileV2F5(t *testing.T, neu CodecNeu) {
	typen := strings.Fields(siebenV2 + " " + achtV2)
	var inhalt string
	for i, ty := range typen {
		inhalt += prop(fmt.Sprintf("P%d", i), ty)
	}
	m, b := pruefeFall(t, neu, gut("alle Typen", eins2(typ("T", inhalt))))
	e, _ := m.EntityType("Ns.T")
	if len(e.Properties) != 16 {
		t.Fatalf("%d Properties, erwartet 16", len(e.Properties))
	}
	var funde []model.Fund
	for i, ty := range typen {
		p := e.Properties[i+1]
		unterstuetzt := i < 7
		if p.EdmType != modellTypV2(ty) || p.Unterstuetzt() != unterstuetzt {
			t.Errorf("%s: Property %+v, erwartet %s und Unterstützung %v", ty, p, modellTypV2(ty), unterstuetzt)
		}
		if !unterstuetzt {
			funde = append(funde, model.Fund{Art: model.FundPropertyTyp, Ort: fmt.Sprintf("Ns.T/P%d", i), Name: ty})
		}
	}
	gleich(t, "Funde", b.Funde, funde)
	gleich(t, "UnterstuetzteEDMTypen", model.UnterstuetzteEDMTypen(), strings.Fields("Edm.String Edm.Boolean Edm.Int32 Edm.Int64 Edm.Decimal Edm.Guid Edm.Date Edm.DateTimeOffset"))
}

func f6DocV2() []byte {
	t := typ("T", `<Property Name="B" Type="Edm.String" Nullable="false"/><Property Name="C" Type="Edm.String" Nullable="true"/>`+nav2("N1")+nav2("N2"))
	return eins2(`<EntityType xmlns:m="` + v2MetaNs + `" Name="O" OpenType="true" m:HasStream="false">` + schluesselID + `</EntityType>` + t + assoz2)
}

// zeileV2F6: Open ist immer falsch, Nullable gilt als wahr und wird gelesen, Zugriffe liefern Kopien (SPEC-171 Nr. 8).
func zeileV2F6(t *testing.T, neu CodecNeu) {
	m, b := pruefeFall(t, neu, gut("F6", f6DocV2()))
	e, _ := m.EntityType("Ns.T")
	gleich(t, "Properties", e.Properties, []model.Property{{Name: "Id", EdmType: "Edm.Int32", Nullable: true}, {Name: "B", EdmType: "Edm.String"}, {Name: "C", EdmType: "Edm.String", Nullable: true}})
	gleich(t, "Navigation", e.NavigationProperties, []model.NavigationProperty{{Name: "N1", Target: "Ns.T"}, {Name: "N2", Target: "Ns.T"}})
	for _, x := range m.EntityTypes() {
		if x.Open {
			t.Errorf("%s ist offen", x.Name)
		}
	}
	gleich(t, "Funde", b.Funde, []model.Fund{{Art: model.FundErweiterung, Ort: "Ns.O", Name: "{" + v2MetaNs + "}HasStream"}})
	soll := ergebnis(m, b, nil)
	for _, x := range m.EntityTypes() {
		x.Keys[0], x.Name = "X", "X"
		x.Properties[0].Name = "X"
		if len(x.NavigationProperties) > 0 {
			x.NavigationProperties[0].Name = "X"
		}
	}
	got, _ := m.EntityType("Ns.T")
	got.Properties[0].Name, got.Keys[0] = "Y", "Y"
	m.Namespaces()[0] = "Y"
	if len(m.EntitySets()) > 0 {
		m.EntitySets()[0].Name = "Y"
	}
	if ergebnis(m, b, nil) != soll {
		t.Errorf("eine geänderte Kopie ändert das Modell")
	}
}

// zeileV2H1: die Ursache steht im Text, der Ort in der Schreibweise von strconv.Quote (SPEC-125 Nr. 13): ein Fall je Ursache.
func zeileV2H1(t *testing.T, neu CodecNeu) {
	doc := eins2(`<EntityType Name="T"><Key><PropertyRef Name="Id&#10;x"/></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
	_, _, err := lies(neu, fall{doc: doc})
	erwarteFehler(t, err, model.UrsacheNichtAufloesbar, "Ns.T/Id\nx@Name")
	for i := range len(err.Error()) {
		if err.Error()[i] < 0x20 {
			t.Fatalf("Byte %#x im Text %q", err.Error()[i], err.Error())
		}
	}
	want := `ODR-E6004: Metadaten ungültig oder nicht unterstützt: nicht_aufloesbar "Ns.T/Id\nx@Name"`
	var mf *model.MetadatenFehler
	if err.Error() != want || !errors.As(err, &mf) || strings.Contains(model.CodeE6004.Message(), "Ns.T") {
		t.Errorf("Text %q, erwartet %q", err.Error(), want)
	}
	gesehen := map[model.MetadatenUrsache]bool{}
	for _, f := range alleFaelleV2(t) {
		if f.ursache == "" || gesehen[f.ursache] {
			continue
		}
		gesehen[f.ursache] = true
		_, _, err := lies(neu, f)
		soll := "ODR-E6004: Metadaten ungültig oder nicht unterstützt: " + string(f.ursache)
		if f.ort != "" {
			soll += fmt.Sprintf(" %q", f.ort)
		}
		if err == nil || err.Error() != soll || !errors.Is(err, model.ErrMetadatenUngueltig) {
			t.Errorf("%s: Text %v, erwartet %q", f.name, err, soll)
		}
	}
	if len(gesehen) != 13 {
		t.Errorf("%d Ursachen mit Text geprüft, erwartet 13", len(gesehen))
	}
}
