package outboundtest

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func gleich(t *testing.T, name string, got, want any) {
	t.Helper()
	if g, w := fmt.Sprintf("%+v", got), fmt.Sprintf("%+v", want); g != w {
		t.Errorf("%s:\n  ist    %s\n  soll   %s", name, g, w)
	}
}

func gross(f fall) fall { f.gross = true; return f }

// nameVon und edmTypVon lesen die i-te Property; fehlt sie, melden sie den leeren Text statt abzubrechen.
func nameVon(ps []model.Property, i int) string {
	if i < len(ps) {
		return ps[i].Name
	}
	return ""
}

func edmTypVon(ps []model.Property, i int) string {
	if i < len(ps) {
		return ps[i].EdmType
	}
	return ""
}

func navs(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, `<NavigationProperty Name="n%d" Type="Ns.T"/>`, i)
	}
	return b.String()
}

func ableitungen(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString(ableitung(fmt.Sprintf("D%d", i), "Ns.B", ""))
	}
	return b.String()
}

func vieleTypen(n, je int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString(typ(fmt.Sprintf("T%d", i), props(je)))
	}
	return b.String()
}

func faelleE3() []fall {
	z := model.UrsacheZuVieleMembers
	return []fall{
		gross(gut("65 536 Properties", eins(typ("T", props(65535))))),
		gross(schlecht("65 537 Properties", eins(typ("T", props(65536))), z, "Ns.T")),
		gross(gut("32 768 Properties und 32 768 Navigation-Properties", eins(typ("T", props(32767)+navs(32768))))),
		gross(schlecht("eine Navigation-Property mehr", eins(typ("T", props(32767)+navs(32769))), z, "Ns.T")),
		gross(gut("Basis mit 256 Members und 255 Ableitungen", eins(typ("B", props(255))+ableitungen(255)))),
		gross(schlecht("256 Ableitungen", eins(typ("B", props(255))+ableitungen(256)), z, "Ns.D256")),
		gross(schlecht("66 Typen mit je 1000 Members", eins(vieleTypen(66, 999)), z, "Ns.T66")),
	}
}

func zeileE3(t *testing.T, neu CodecNeu) { pruefeFaelle(t, neu, faelleE3()) }

func faelleF(t *testing.T) []fall {
	t.Helper()
	return []fall{
		gut("Referenzdienst", festes(t, "referenz.xml")), gut("Vererbung", festes(t, "vererbung.xml")), gut("Composite-Schlüssel", festes(t, "composite.xml")),
		gut("Aliase", f4Doc([7]bool{true, true, true, true, true, true, true})), gut("Randformen des Modells", f6Doc()), gut("ohne EntityContainer", eins(typ("T", ""))),
	}
}

// zeileF1: das Modell des Referenzdienstes (SPEC-125 Nr. 11).
func zeileF1(t *testing.T, neu CodecNeu) {
	m, b := pruefeFall(t, neu, gut("R", festes(t, "referenz.xml")))
	if m.Version() != "4.0" {
		t.Errorf("Version %q", m.Version())
	}
	gleich(t, "Namespaces", m.Namespaces(), []string{"Demo"})
	product := model.EntityType{Name: "Demo.Product", Keys: []string{"ID"},
		Properties: []model.Property{{Name: "ID", EdmType: "Edm.Int32"}, {Name: "Name", EdmType: "Edm.String", Nullable: true},
			{Name: "Price", EdmType: "Edm.Decimal"}, {Name: "Weight", EdmType: "Edm.Double", Nullable: true}},
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
		{Set: "Products", Path: "Category", Target: "Categories"}, {Set: "Categories", Path: "Products", Target: "Demo.Container/Products"}})
	gleich(t, "Funde", b.Funde, []model.Fund{{Art: model.FundPropertyTyp, Ort: "Demo.Product/Weight", Name: "Edm.Double"}})
	if _, ok := m.EntityType("Self.Product"); ok {
		t.Error("EntityType löst einen Alias auf")
	}
	if s, ok := m.EntitySet("Products"); !ok || s.EntityType != "Demo.Product" {
		t.Errorf("EntitySet: %v %v", s, ok)
	}
}

// zeileF2: Vererbung im Modell: wirksame Keys und Properties, Open geerbt, Abstract nur am Typ.
func zeileF2(t *testing.T, neu CodecNeu) {
	m, _ := pruefeFall(t, neu, gut("V", festes(t, "vererbung.xml")))
	for _, w := range []struct {
		name, base      string
		abstract, open  bool
		properties, nav string
	}{
		{"Zoo.Animal", "", true, false, "Id Name", ""}, {"Zoo.Mammal", "Zoo.Animal", false, true, "Id Name Fur", "Keeper"},
		{"Zoo.Dog", "Zoo.Mammal", false, true, "Id Name Fur Breed", "Keeper"}, {"Zoo.Puppy", "Zoo.Dog", false, true, "Id Name Fur Breed", "Keeper"},
		{"Zoo.Keeper", "", false, false, "Id", ""},
	} {
		e, ok := m.EntityType(w.name)
		var props, nav []string
		for _, p := range e.Properties {
			props = append(props, p.Name)
		}
		for _, n := range e.NavigationProperties {
			nav = append(nav, n.Name)
		}
		if !ok || e.BaseType != w.base || e.Abstract != w.abstract || e.Open != w.open || strings.Join(props, " ") != w.properties ||
			strings.Join(nav, " ") != w.nav || strings.Join(e.Keys, " ") != "Id" {
			t.Errorf("%s: %+v", w.name, e)
		}
	}
	e, _ := m.EntityType("Zoo.Dog")
	if edmTypVon(e.Properties, 1) != edmTypVon(e.Properties, 3) {
		t.Errorf("Name und Breed sind beide Edm.String: %+v", e.Properties)
	}
}

func zeileF3(t *testing.T, neu CodecNeu) {
	m, _ := pruefeFall(t, neu, gut("K", festes(t, "composite.xml")))
	e, _ := m.EntityType("Shop.OrderLine")
	g, _ := m.EntityType("Shop.Gift")
	if strings.Join(e.Keys, " ") != "OrderID LineNo" || strings.Join(g.Keys, " ") != "OrderID LineNo" || nameVon(e.Properties, 0) != "LineNo" {
		t.Errorf("Keys %v, geerbt %v, Properties %v", e.Keys, g.Keys, e.Properties)
	}
	for _, w := range []struct{ key, props, soll string }{{"A B", "B A", "A B"}, {"B A", "A B", "B A"}, {"B A", "B A", "B A"}} {
		var refs, ps string
		for _, k := range strings.Fields(w.key) {
			refs += `<PropertyRef Name="` + k + `"/>`
		}
		for _, p := range strings.Fields(w.props) {
			ps += prop(p, "Edm.Int32")
		}
		m, _ := pruefeFall(t, neu, gut("Key "+w.key, eins(`<EntityType Name="T"><Key>`+refs+`</Key>`+ps+`</EntityType>`)))
		if got, _ := m.EntityType("Ns.T"); strings.Join(got.Keys, " ") != w.soll {
			t.Errorf("Key %s bei Properties %s: %v", w.key, w.props, got.Keys)
		}
	}
}

// f4Doc hat zwei Namensräume mit je einem Alias und demselben Typnamen `T`, einen Namensraum in zwei Schemas
// und sechs Verweise; alias[i] sagt, ob der Verweis i über den Alias oder den Namensraum geht.
func f4Doc(alias [7]bool) []byte {
	pick := func(i int, a, ns string) string {
		if alias[i] {
			return a
		}
		return ns
	}
	holder := `<EntityType Name="Holder">` + schluesselID + prop("P2", pick(2, "A.T", "NsA.T")) + prop("P6", "Collection("+pick(6, "B.T", "NsB.T")+")") +
		`<NavigationProperty Name="N3" Type="` + pick(3, "B.T", "NsB.T") + `"/></EntityType>`
	sets := `<EntityContainer Name="Container"><EntitySet Name="S1" EntityType="` + pick(4, "A.T", "NsA.T") + `"><NavigationPropertyBinding Path="P" Target="` +
		pick(5, "A.Container", "NsA.Container") + `/S1"/></EntitySet></EntityContainer>`
	a := schemaAlias("NsA", "A", typ("T", prop("X", "Edm.String"))+ableitung("D1", pick(1, "A.T", "NsA.T"), "")+holder+sets)
	b := schemaAlias("NsB", "B", typ("T", prop("Y", "Edm.Int32")))
	return []byte(edmx(a + b + schemaAlias("NsA", "A2", typ("V", ""))))
}

func zeileF4(t *testing.T, neu CodecNeu) {
	m0, b0 := pruefeFall(t, neu, gut("alle über Namensraum", f4Doc([7]bool{})))
	ergebnis0 := ergebnis(m0, b0, nil)
	gleich(t, "Namespaces", m0.Namespaces(), []string{"NsA", "NsB"})
	a, _ := m0.EntityType("NsA.T")
	b, _ := m0.EntityType("NsB.T")
	d, _ := m0.EntityType("NsA.D1")
	h, _ := m0.EntityType("NsA.Holder")
	gleich(t, "NsA.T", nameVon(a.Properties, 1), "X")
	gleich(t, "NsB.T", nameVon(b.Properties, 1), "Y")
	gleich(t, "BaseType", d.BaseType, "NsA.T")
	gleich(t, "Properties", h.Properties, []model.Property{{Name: "Id", EdmType: "Edm.Int32", Nullable: true}, {Name: "P2", EdmType: "NsA.T", Nullable: true}, {Name: "P6", EdmType: "Collection(NsB.T)", Nullable: true}})
	gleich(t, "Navigation", h.NavigationProperties, []model.NavigationProperty{{Name: "N3", Target: "NsB.T"}})
	gleich(t, "Set", m0.EntitySets(), []model.EntitySet{{Name: "S1", EntityType: "NsA.T"}})
	gleich(t, "Binding", m0.NavigationBindings(), []model.NavigationBinding{{Set: "S1", Path: "P", Target: "NsA.Container/S1"}})
	gleich(t, "Funde", b0.Funde, []model.Fund{{Art: model.FundPropertyTyp, Ort: "NsA.Holder/P2", Name: "NsA.T"}, {Art: model.FundPropertyTyp, Ort: "NsA.Holder/P6", Name: "Collection(NsB.T)"}})
	varianten := [][7]bool{{true, true, true, true, true, true, true}}
	for i := 1; i <= 6; i++ {
		var v [7]bool
		v[i] = true
		varianten = append(varianten, v)
	}
	for i, v := range varianten {
		m, b := pruefeFall(t, neu, gut(fmt.Sprintf("Variante %d", i), f4Doc(v)))
		if ergebnis(m, b, nil) != ergebnis0 {
			t.Errorf("Variante %d (%v) ergibt ein anderes Modell als die über den Namensraum", i, v)
		}
	}
}

func f6Doc() []byte {
	t := typ("T", `<Property Name="B" Type="Edm.String" Nullable="false"/><Property Name="C" Type="Edm.String" Nullable="true"/>`+
		`<NavigationProperty Name="N1" Type="Ns.T"/><NavigationProperty Name="N2" Type="Collection(Ns.T)"/>`)
	return eins(t + `<EntityContainer Name="C"><Singleton Name="Me" Type="Ns.T"/><EntitySet Name="S" EntityType="Ns.T">` +
		`<NavigationPropertyBinding Path="Nope" Target="S"/><NavigationPropertyBinding Path="N1" Target="Me"/>` +
		`<NavigationPropertyBinding Path="N2" Target="Ns.C/Me"/></EntitySet></EntityContainer>`)
}

func zeileF6(t *testing.T, neu CodecNeu) {
	m, _ := pruefeFall(t, neu, gut("F6", f6Doc()))
	e, _ := m.EntityType("Ns.T")
	gleich(t, "Properties", e.Properties, []model.Property{{Name: "Id", EdmType: "Edm.Int32", Nullable: true}, {Name: "B", EdmType: "Edm.String"}, {Name: "C", EdmType: "Edm.String", Nullable: true}})
	gleich(t, "Navigation", e.NavigationProperties, []model.NavigationProperty{{Name: "N1", Target: "Ns.T"}, {Name: "N2", Target: "Ns.T", Collection: true}})
	gleich(t, "Bindings", m.NavigationBindings(), []model.NavigationBinding{{Set: "S", Path: "Nope", Target: "S"}, {Set: "S", Path: "N1", Target: "Me"}, {Set: "S", Path: "N2", Target: "Ns.C/Me"}})
	gleich(t, "Sets", m.EntitySets(), []model.EntitySet{{Name: "S", EntityType: "Ns.T"}})
	m, _ = pruefeFall(t, neu, gut("ohne EntityContainer", eins(typ("T", ""))))
	if len(m.EntitySets()) != 0 || len(m.NavigationBindings()) != 0 || len(m.EntityTypes()) != 1 {
		t.Errorf("ohne EntityContainer: %v %v", m.EntitySets(), m.NavigationBindings())
	}
}

// zeileH1: die Ursache steht im Text, der Ort in der Schreibweise von strconv.Quote (Nr. 13).
func zeileH1(t *testing.T, neu CodecNeu) {
	doc := eins(`<EntityType Name="T"><Key><PropertyRef Name="Id&#10;x"/></Key>` + prop("Id", "Edm.Int32") + `</EntityType>`)
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
}
