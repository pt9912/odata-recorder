package model_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestFilterKnoten (Zeile 33): sechs Wertetypen erfüllen FilterAusdruck, die acht Operatoren tragen ihren Text in
// Kleinbuchstaben. Menge 14.
func TestFilterKnoten(t *testing.T) {
	knoten := []model.FilterAusdruck{model.FilterVergleich{}, model.FilterVerknuepfung{}, model.FilterNicht{}, model.FilterEigenschaft{},
		model.FilterLiteral{}, model.FilterNull{}}
	for _, k := range knoten {
		if k == nil {
			t.Error("Knoten nil")
		}
	}
	vergleich := []struct {
		op   model.VergleichOp
		text string
	}{{model.VergleichEq, "eq"}, {model.VergleichNe, "ne"}, {model.VergleichGt, "gt"}, {model.VergleichGe, "ge"},
		{model.VergleichLt, "lt"}, {model.VergleichLe, "le"}}
	verknuepfung := []struct {
		op   model.VerknuepfungOp
		text string
	}{{model.VerknuepfungUnd, "and"}, {model.VerknuepfungOder, "or"}}
	for _, v := range vergleich {
		if string(v.op) != v.text {
			t.Errorf("Vergleich %q statt %q", v.op, v.text)
		}
	}
	for _, v := range verknuepfung {
		if string(v.op) != v.text {
			t.Errorf("Verknüpfung %q statt %q", v.op, v.text)
		}
	}
	if len(knoten)+len(vergleich)+len(verknuepfung) != 14 {
		t.Fatal("Menge")
	}
}
