package model_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flGestalt meldet jede Slice, Map, jeden Zeiger, Kanal und jede Funktion im Typ, auch in verschachtelten Strukturen; Schnittstellen
// sind die Kinder des Baums und führen nicht weiter (SPEC-165 Nr. 2).
func flGestalt(t *testing.T, typ reflect.Type, pfad string) {
	t.Helper()
	switch typ.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Array:
		t.Errorf("%s: %s im Typ", pfad, typ.Kind())
	case reflect.Struct:
		for i := range typ.NumField() {
			flGestalt(t, typ.Field(i).Type, pfad+"."+typ.Field(i).Name)
		}
	}
}

// TestFilterNurWerte (SPEC-165 Nr. 2, Zeile 33): die sechs Knotentypen enthalten keine Slice, keine Map, keinen Zeiger, keinen
// Kanal, keine Funktion und kein Array, auch nicht verschachtelt. Menge 6 Typen.
func TestFilterNurWerte(t *testing.T) {
	typen := []reflect.Type{reflect.TypeOf(model.FilterVergleich{}), reflect.TypeOf(model.FilterVerknuepfung{}),
		reflect.TypeOf(model.FilterNicht{}), reflect.TypeOf(model.FilterEigenschaft{}), reflect.TypeOf(model.FilterLiteral{}),
		reflect.TypeOf(model.FilterNull{})}
	for _, typ := range typen {
		flGestalt(t, typ, typ.Name())
	}
	if len(typen) != 6 {
		t.Fatal(len(typen))
	}
}

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
