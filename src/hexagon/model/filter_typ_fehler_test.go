package model_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestFilterTypFehler (Zeile 35): jede der sieben Ursachen trägt Klasse, Code und Text; errors.Is trifft genau die Sentinel der
// Klasse; kein Teil der Eingabe (Name GEHEIM, Literal GEHEIM) steht im Text, geprüft an den Fehlern von PruefeFilter.
func TestFilterTypFehler(t *testing.T) {
	alle := model.AlleFilterTypUrsachen()
	if len(alle) != 7 {
		t.Fatalf("%d Ursachen", len(alle))
	}
	sentinels := []error{model.ErrURLUngueltig, model.ErrURLNichtUnterstuetzt, model.ErrFilterNichtBeobachtet}
	typ := typMit(model.Property{Name: "GEHEIM", EdmType: "Edm.Binary"}, model.Property{Name: "Verborgen", EdmType: "Edm.String"})
	typ.Properties = append(typ.Properties, model.Property{Name: "GEHEIMZahl", EdmType: "Edm.Int32"}, model.Property{Name: "GEHEIMGuid", EdmType: "Edm.Guid"})
	beob := func(s string) bool { return s != "Verborgen" }
	secretLit := lit(model.LiteralZeichenkette, "'GEHEIM'")
	bauen := map[model.FilterTypUrsache]model.FilterAusdruck{
		model.FilterEigenschaftUnbekannt:      prop("GEHEIMX"),
		model.FilterNichtBoolesch:             prop("GEHEIMZahl"),
		model.FilterTypUnvertraeglich:         vgl("eq", prop("GEHEIMZahl"), secretLit),
		model.FilterLiteralBereich:            vgl("eq", prop("GEHEIMZahl"), lit(model.LiteralZahl, "GEHEIM")),
		model.FilterTypNichtUnterstuetzt:      prop("GEHEIM"),
		model.FilterOperatorNichtUnterstuetzt: vgl("gt", prop("GEHEIMGuid"), lit(model.LiteralGUID, "GEHEIM")),
		model.FilterNichtBeobachtet:           prop("Verborgen"),
	}
	klasse := map[model.FilterTypUrsache]int{
		model.FilterEigenschaftUnbekannt: 0, model.FilterNichtBoolesch: 0, model.FilterTypUnvertraeglich: 0, model.FilterLiteralBereich: 0,
		model.FilterTypNichtUnterstuetzt: 1, model.FilterOperatorNichtUnterstuetzt: 1, model.FilterNichtBeobachtet: 2,
	}
	codes := []model.Code{model.CodeE6005, model.CodeE6001, model.CodeE6003}
	praefix := []string{"ODR-E6005: Filter ungültig: ", "ODR-E6001: Filter nicht unterstützt: ", "ODR-E6003: Filter nicht beobachtet: "}
	n := 0
	for _, u := range alle {
		k := klasse[u]
		e := &model.FilterTypFehler{Ursache: u}
		if e.Code() != codes[k] || e.Error() != praefix[k]+string(u) {
			t.Errorf("%s: %v %q", u, e.Code(), e.Error())
		}
		for i, s := range sentinels {
			if errors.Is(e, s) != (i == k) {
				t.Errorf("%s: errors.Is(%v) = %v", u, s, i != k)
			}
			// verpackt und verbunden
			if errors.Is(errors.Join(errors.New("x"), e), s) != (i == k) {
				t.Errorf("%s: verbunden %v", u, s)
			}
			n++
		}
		if errors.Is(e, model.ErrUngueltig) {
			t.Errorf("%s: Treffer auf ErrUngueltig", u)
		}
		// aus PruefeFilter: genau diese Ursache, genau dieser Text, kein Teil der Eingabe
		err := model.PruefeFilter(typ, bauen[u], beob)
		var ft *model.FilterTypFehler
		if !errors.As(err, &ft) || ft.Ursache != u || err.Error() != praefix[k]+string(u) || ft.Code() != codes[k] {
			t.Errorf("%s: %v", u, err)
		}
		if err != nil && (strings.Contains(err.Error(), "GEHEIM") || strings.Contains(err.Error(), "Verborgen")) {
			t.Errorf("%s: Eingabe im Text: %v", u, err)
		}
	}
	if n != 21 {
		t.Errorf("%d Paare", n)
	}
	// jeder Verstoß ist ein FilterTypFehler, nie ErrUngueltig
	for _, s := range stellen() {
		if err := model.PruefeFilter(typ, s.baue(prop("GEHEIMX")), beob); errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", s.name, err)
		}
	}
}
