package integration_test

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// abfrageDaten sind fünf Produkte; Price von Delta ist null, Active von Gamma ist null.
func abfrageDaten() (model.EntityType, []model.Entity) {
	typ := model.EntityType{Name: "Ns.Product", Keys: []string{"Id"}, Properties: []model.Property{
		{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Name", EdmType: "Edm.String"}, {Name: "Price", EdmType: "Edm.Decimal"}, {Name: "Active", EdmType: "Edm.Boolean"},
	}}
	zeile := func(id int, name string, preis, aktiv model.Wert) model.Entity {
		return model.Entity{Typ: "Ns.Product", Schluessel: strconv.Itoa(id), Uebersprungen: []model.Uebersprungen{}, Properties: []model.PropertyWert{
			{Name: "Id", Wert: model.Wert{Art: model.ArtInt32, Text: strconv.Itoa(id)}},
			{Name: "Name", Wert: model.Wert{Art: model.ArtString, Text: name}},
			{Name: "Price", Wert: preis}, {Name: "Active", Wert: aktiv},
		}}
	}
	dez := func(s string) model.Wert { return model.Wert{Art: model.ArtDecimal, Text: s} }
	wahrheit := func(s string) model.Wert { return model.Wert{Art: model.ArtBoolean, Text: s} }
	null := model.Wert{Art: model.ArtNull}
	return typ, []model.Entity{
		zeile(1, "Alpha", dez("10"), wahrheit("true")), zeile(2, "Beta", dez("20.5"), wahrheit("false")), zeile(3, "Gamma", dez("5"), null),
		zeile(4, "Delta", null, wahrheit("true")), zeile(5, "Epsilon", dez("30"), wahrheit("true")),
	}
}

// TestAbfrageAusURL (Zeile 34): der Query-Text des Clients führt über den Parser und den Interpreter zum erwarteten Teilbestand:
// zwölf Query-Texte mit Ergebnis und zwei, die im Parser enden.
func TestAbfrageAusURL(t *testing.T) {
	typ, es := abfrageDaten()
	parser := urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser()))
	alle := []string{"Id", "Name", "Price", "Active"}
	faelle := []struct {
		query string
		ids   []string
		anz   int
		props []string
	}{
		{"$filter=Price gt 5&$orderby=Name desc&$skip=1&$top=2&$select=Name", []string{"2", "1"}, 3, []string{"Name"}},
		{"$top=0", []string{}, 5, alle},
		{"$skip=10", []string{}, 5, alle},
		{"$count=true", []string{"1", "2", "3", "4", "5"}, 5, alle},
		{"$orderby=Name desc", []string{"3", "5", "4", "2", "1"}, 5, alle},
		{"$select=*,Name", []string{"1", "2", "3", "4", "5"}, 5, alle},
		{"$filter=Price eq null", []string{"4"}, 1, alle},
		{"$filter=Active eq true and Price lt 25", []string{"1"}, 1, alle},
		{"$orderby=Price", []string{"4", "3", "1", "2", "5"}, 5, alle},
		{"$filter=not Active", []string{"2"}, 1, alle},
		{"$skip=1&$top=1&$orderby=Price desc", []string{"2"}, 5, alle},
		{"$select=Name,Price", []string{"1", "2", "3", "4", "5"}, 5, []string{"Name", "Price"}},
	}
	for _, f := range faelle {
		u, err := parser.ParseURL("/Products", strings.ReplaceAll(f.query, " ", "%20"))
		if err != nil {
			t.Fatalf("%s: %v", f.query, err)
		}
		r, err := model.FuehreAus(typ, es, u.Optionen, nil)
		if err != nil {
			t.Fatalf("%s: %v", f.query, err)
		}
		var ids []string
		for _, e := range r.Entities {
			ids = append(ids, e.Schluessel)
			var props []string
			for _, p := range e.Properties {
				props = append(props, p.Name)
			}
			if !slices.Equal(props, f.props) {
				t.Errorf("%s: Properties %v, erwartet %v", f.query, props, f.props)
			}
		}
		if !slices.Equal(ids, f.ids) && (len(ids) != 0 || len(f.ids) != 0) || r.Anzahl != f.anz || r.Entities == nil {
			t.Errorf("%s: %v Anzahl %d, erwartet %v Anzahl %d", f.query, ids, r.Anzahl, f.ids, f.anz)
		}
	}
	for _, q := range []string{"$top=-1", "$top=1.5"} {
		_, err := parser.ParseURL("/Products", q)
		var uf *model.URLFehler
		if !errors.As(err, &uf) || uf.Ursache != model.URLWert || !errors.Is(err, model.ErrURLUngueltig) {
			t.Errorf("%s: %v, erwartet die Ursache wert", q, err)
		}
	}
}
