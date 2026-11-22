package model_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestPruefeExpandContract (Zeile 18, LH-FA-16.a): der Validator-Contract-Test. Die Tabelle nimmt die Fälle aus dem Wortlaut
// von SPEC-263 Nr. 2 a bis d, 3 und 4 (die Spezifikation führt keine eigene Beispieltabelle); sie hat keine Eingabe aus einem
// Generator und kein Fixture-Modell, nur eigene Typen mit den Namen eines Beispiels. Die Signatur ist die der Spezifikation.
func TestPruefeExpandContract(t *testing.T) {
	produkt := model.EntityType{Name: "Ns.Product", Keys: []string{"ProductID"},
		Properties: []model.Property{{Name: "ProductID", EdmType: "Edm.Int32"}, {Name: "ProductName", EdmType: "Edm.String"}, {Name: "Photo", EdmType: "Edm.Binary"}},
		NavigationProperties: []model.NavigationProperty{
			{Name: "Category", Target: "Ns.Category"}, {Name: "Supplier", Target: "Ns.Supplier"}, {Name: "OrderDetails", Target: "Ns.OrderDetail", Collection: true}}}
	abgeleitet := model.EntityType{Name: "Ns.DiscontinuedProduct", BaseType: "Ns.Product", Keys: produkt.Keys,
		Properties:           append([]model.Property{{Name: "Reason", EdmType: "Edm.String"}}, produkt.Properties...),
		NavigationProperties: append([]model.NavigationProperty{{Name: "Replacement", Target: "Ns.Product"}}, produkt.NavigationProperties...)}
	offen := model.EntityType{Name: "Ns.OpenProduct", Open: true, Properties: produkt.Properties, NavigationProperties: produkt.NavigationProperties}
	abstrakt := model.EntityType{Name: "Ns.Shape", Abstract: true, Properties: []model.Property{{Name: "ShapeID", EdmType: "Edm.Int32"}},
		NavigationProperties: []model.NavigationProperty{{Name: "Owner", Target: "Ns.Product"}}}
	unbekannt, keine, nichtUnterstuetzt, programm := string(model.ExpandNavigationUnbekannt), string(model.ExpandKeineNavigation),
		string(model.ExpandNavigationNichtUnterstuetzt), "programmfehler"

	faelle := []struct {
		art   string
		typ   model.EntityType
		namen []string
		soll  string
	}{
		// Nr. 2 a: eine Navigation des Typs, auf eine Entity oder eine Collection, auch eine geerbte
		{"2a Entity", produkt, []string{"Category"}, "gueltig"},
		{"2a Collection", produkt, []string{"OrderDetails"}, "gueltig"},
		{"2a mehrere", produkt, []string{"Supplier", "OrderDetails", "Category"}, "gueltig"},
		{"2a geerbt", abgeleitet, []string{"Category"}, "gueltig"},
		{"2a eigene und geerbte", abgeleitet, []string{"OrderDetails", "Replacement"}, "gueltig"},
		{"2a im offenen Typ", offen, []string{"Category"}, "gueltig"},
		{"2a im abstrakten Typ", abstrakt, []string{"Owner"}, "gueltig"},
		// Nr. 2 b: eine Property des Typs, auch mit nicht unterstütztem Typ und geerbt
		{"2b Property", produkt, []string{"ProductName"}, keine},
		{"2b nicht unterstützter Typ", produkt, []string{"Photo"}, keine},
		{"2b geerbte Property", abgeleitet, []string{"ProductID"}, keine},
		{"2b im offenen Typ", offen, []string{"ProductID"}, keine},
		{"2b nach einer Navigation", produkt, []string{"Category", "ProductName"}, keine},
		// Nr. 2 c: ein Name im offenen Typ, den keine Liste nennt
		{"2c offener Typ", offen, []string{"Beliebig"}, nichtUnterstuetzt},
		{"2c nach einer Navigation", offen, []string{"Category", "Beliebig"}, nichtUnterstuetzt},
		// Nr. 2 d: ein Name im nicht offenen Typ, den keine Liste nennt
		{"2d fremder Name", produkt, []string{"Foo"}, unbekannt},
		{"2d Schreibweise", produkt, []string{"category"}, unbekannt},
		{"2d Navigation des Subtyps", produkt, []string{"Replacement"}, unbekannt},
		{"2d Navigation eines anderen Typs", produkt, []string{"Owner"}, unbekannt},
		{"2d abstrakter Typ", abstrakt, []string{"Foo"}, unbekannt},
		{"2d abgeleiteter Typ", abgeleitet, []string{"Foo"}, unbekannt},
		// Nr. 2: der erste Verstoß von links entscheidet, gleich welcher Klasse
		{"erster Verstoß: fremd vor Property", produkt, []string{"Foo", "ProductID"}, unbekannt},
		{"erster Verstoß: Property vor fremd", produkt, []string{"ProductID", "Foo"}, keine},
		{"erster Verstoß: offen vor Property", offen, []string{"Beliebig", "ProductID"}, nichtUnterstuetzt},
		{"erster Verstoß: Property vor offen", offen, []string{"ProductID", "Beliebig"}, keine},
		// Nr. 4: Programmfehler, vor jedem Verstoß
		{"4 leere Liste", produkt, nil, programm},
		{"4 leerer Name", produkt, []string{""}, programm},
		{"4 Stern", produkt, []string{"*"}, programm},
		{"4 doppelter Name", produkt, []string{"Category", "Category"}, programm},
		{"4 Programmfehler nach Verstoß", produkt, []string{"Foo", ""}, programm},
		{"4 zwei Namen in anderer Schreibweise", produkt, []string{"Category", "category"}, unbekannt},
	}
	for _, f := range faelle {
		if got := exvKlasse(model.PruefeExpand(f.typ, exvO(f.namen...))); got != f.soll {
			t.Errorf("%s: %s, erwartet %s", f.art, got, f.soll)
		}
	}
	if got := exvKlasse(model.PruefeExpand(produkt, nil)); got != programm {
		t.Errorf("4 o nil: %s", got)
	}

	// Nr. 3: Klasse, Code und Text je Ursache
	for _, z := range exvTabelle() {
		typ, o := produkt, exvO("Foo")
		switch z.ursache {
		case model.ExpandKeineNavigation:
			o = exvO("ProductID")
		case model.ExpandNavigationNichtUnterstuetzt:
			typ = offen
		}
		exvPruefeFehler(t, "3 "+string(z.ursache), model.PruefeExpand(typ, o), z)
	}

	// Signatur: PruefeExpand(EntityType, *ExpandOption) error, ohne ctx und ohne Interface als Eingabe
	sig := reflect.TypeOf(model.PruefeExpand)
	ctx := reflect.TypeFor[context.Context]()
	if sig.NumIn() != 2 || sig.IsVariadic() || sig.In(0) != reflect.TypeFor[model.EntityType]() || sig.In(1) != reflect.TypeFor[*model.ExpandOption]() ||
		sig.NumOut() != 1 || sig.Out(0) != reflect.TypeFor[error]() {
		t.Errorf("Signatur %s", sig)
	}
	for i := range sig.NumIn() {
		if sig.In(i).Kind() == reflect.Interface || sig.In(i).Implements(ctx) {
			t.Errorf("Parameter %d ist ein Interface oder ein Kontext: %s", i, sig.In(i))
		}
	}
}
