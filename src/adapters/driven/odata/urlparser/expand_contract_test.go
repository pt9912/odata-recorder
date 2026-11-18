package urlparser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestExpandContract (Zeile 26, LH-FA-16.a, SPEC-262 Nr. 9): der Parser-Contract-Test. Die Tabelle der Beispiele der Spezifikation
// ist die Fixture, ein Fall je Zeile (27 Zeilen, die zwei Zeilen mit 32 und 33 Namen erzeugt der Test); keine Eingabe stammt aus
// einem Generator. Je Ressourcenart (`/$metadata`, `/Set`, `/Set(Schlüssel)`, `/Set/$count`) liest der Parser dieselbe Option.
func TestExpandContract(t *testing.T) {
	p := exParser()
	zeilen := []struct {
		query string
		soll  exSoll
	}{
		{"$expand=Category", exErg("Category")},
		{"$expand=Category,Supplier", exErg("Category", "Supplier")},
		{"$expand=Supplier,Category", exErg("Supplier", "Category")},
		{"$expand=Category%2CSupplier", exErg("Category", "Supplier")},
		{"%24expand=Category", exErg("Category")},
		{"$expand=Category,category", exErg("Category", "category")},
		{"$expand=", exWert()},
		{"$expand=Category,,Supplier", exWert()},
		{"$expand=Category,Category", exWert()},
		{"$expand=Category, Supplier", exWert()},
		{"$expand=1a", exWert()},
		{"$expand=Category%252CSupplier", exWert()},
		{"$expand=Category)", exWert()},
		{"$expand=Category/", exWert()},
		{"$expand=*,,", exWert()},
		{"$expand=*,*", exWert()},
		{"$expand=*", exNichtUnterstuetzt()},
		{"$expand=Category/Products", exNichtUnterstuetzt()},
		{"$expand=Ns.Sub/Category", exNichtUnterstuetzt()},
		{"$expand=Category/$ref", exNichtUnterstuetzt()},
		{"$expand=Category($select=Id,Name)", exNichtUnterstuetzt()},
		{"$expand=,(", exNichtUnterstuetzt()},
		{exQ(strings.Join(exNamen("N", 32), ",")), exErg(exNamen("N", 32)...)},
		{exQ(strings.Join(exNamen("N", 33), ",")), exZuViele()},
		{"$expand=Category&$expand=Supplier", exFehl(model.URLOptionDoppelt, "query/2")},
		{"$Expand=Category", exFehl(model.URLOptionName, "query/1")},
	}
	for _, z := range zeilen {
		exPruefe(t, p, z.query, z.soll)
	}
	_, err := urlparser.New().ParseURL("/S", "$expand=Category")
	var uf *model.URLFehler
	if !errors.As(err, &uf) || uf.Ursache != model.URLOptionNichtUnterstuetzt || uf.Ort != "query/1" {
		t.Errorf("ohne WithExpand: %v", err)
	}
	exZaehlt(t, "Beispiele", len(zeilen)+1, 27)
	for _, pfad := range []string{"/$metadata", "/S", "/S(1)", "/S/$count"} {
		u, err := p.ParseURL(pfad, "$expand=Category")
		if err != nil || u.Optionen.Expand == nil || !gleich(u.Optionen.Expand.Elemente, []model.ExpandElement{{Name: "Category"}}) {
			t.Errorf("%s: %+v %v", pfad, u.Optionen.Expand, err)
		}
	}
}
