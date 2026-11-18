package model_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestOptionenFelder (Zeile 27, SPEC-160 Typen, SPEC-262): `Optionen` trägt die acht Felder in dieser Reihenfolge, `Expand` als
// letztes; der Knoten trägt `Elemente`, das Element `Name`. Menge: 8 + 1 + 1 Felder.
func TestOptionenFelder(t *testing.T) {
	soll := map[reflect.Type][][2]string{
		reflect.TypeFor[model.Optionen](): {
			{"Select", "*model.SelectOption"}, {"OrderBy", "[]model.Sortierung"}, {"Top", "model.ZahlOption"}, {"Skip", "model.ZahlOption"},
			{"Count", "model.WahrOption"}, {"Filter", "*model.FilterOption"}, {"Sonstige", "[]model.SonstigeOption"}, {"Expand", "*model.ExpandOption"},
		},
		reflect.TypeFor[model.ExpandOption]():  {{"Elemente", "[]model.ExpandElement"}},
		reflect.TypeFor[model.ExpandElement](): {{"Name", "string"}},
	}
	for typ, felder := range soll {
		if typ.NumField() != len(felder) {
			t.Errorf("%s hat %d Felder, erwartet %d", typ, typ.NumField(), len(felder))
			continue
		}
		for i, f := range felder {
			if got := typ.Field(i); got.Name != f[0] || got.Type.String() != f[1] {
				t.Errorf("%s Feld %d: %s %s, erwartet %s %s", typ, i, got.Name, got.Type, f[0], f[1])
			}
		}
	}
}
