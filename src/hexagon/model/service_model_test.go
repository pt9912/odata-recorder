package model_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eingaben sind die vier Eingaben des Konstruktors; ein Test verändert sie nach der Übergabe.
type eingaben struct {
	namespaces []string
	typen      []model.EntityType
	sets       []model.EntitySet
	bindings   []model.NavigationBinding
}

func leseSicht() (model.ServiceModel, eingaben) {
	e := eingaben{
		namespaces: []string{"N", "M"},
		typen: []model.EntityType{
			{Name: "N.A", Keys: []string{"Id", "Part"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Part", EdmType: "Edm.String"}},
				NavigationProperties: []model.NavigationProperty{{Name: "Next", Target: "N.B"}}},
			{Name: "N.B", BaseType: "N.A", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}},
		},
		sets:     []model.EntitySet{{Name: "As", EntityType: "N.A"}},
		bindings: []model.NavigationBinding{{Set: "As", Path: "Next", Target: "Bs"}},
	}
	return model.NeuesServiceModel("4.0", e.namespaces, e.typen, e.sets, e.bindings), e
}

func ansicht(m model.ServiceModel) string {
	a, _ := m.EntityType("N.A")
	s, _ := m.EntitySet("As")
	return fmt.Sprintf("%q %q %+v %+v %+v %+v %+v %v", m.Version(), m.Namespaces(), m.EntityTypes(), a, m.EntitySets(), s, m.NavigationBindings(), model.UnterstuetzteEDMTypen())
}

// TestServiceModelLeseSicht (Zeile F5): jeder Zugriff liefert eine Kopie, die der Aufrufer ändern darf.
func TestServiceModelLeseSicht(t *testing.T) {
	aendere := []struct {
		name string
		f    func(m model.ServiceModel, e eingaben)
	}{
		{"Namespaces", func(m model.ServiceModel, _ eingaben) { m.Namespaces()[0] = "X" }},
		{"EntityTypes", func(m model.ServiceModel, _ eingaben) {
			e := m.EntityTypes()
			e[0].Name, e[0].Keys[0], e[0].Properties[0].Name = "X", "X", "X"
		}},
		{"EntityTypes Nav", func(m model.ServiceModel, _ eingaben) {
			m.EntityTypes()[0].NavigationProperties[0].Name = "X"
		}},
		{"EntityType", func(m model.ServiceModel, _ eingaben) {
			e, _ := m.EntityType("N.A")
			e.Keys[0], e.Properties[0].Name, e.NavigationProperties[0].Name = "X", "X", "X"
		}},
		{"EntitySets", func(m model.ServiceModel, _ eingaben) { m.EntitySets()[0].Name = "X" }},
		{"NavigationBindings", func(m model.ServiceModel, _ eingaben) { m.NavigationBindings()[0].Path = "X" }},
		{"UnterstuetzteEDMTypen", func(model.ServiceModel, eingaben) { model.UnterstuetzteEDMTypen()[0] = "X" }},
		{"Eingabe Namespaces", func(_ model.ServiceModel, e eingaben) { e.namespaces[0] = "X" }},
		{"Eingabe EntityTypes", func(_ model.ServiceModel, e eingaben) {
			e.typen[0].Keys[0], e.typen[0].Properties[0].Name, e.typen[0].NavigationProperties[0].Name, e.typen[1].Name = "X", "X", "X", "X"
		}},
		{"Eingabe EntitySets", func(_ model.ServiceModel, e eingaben) { e.sets[0].Name, e.sets[0].EntityType = "X", "X" }},
		{"Eingabe NavigationBindings", func(_ model.ServiceModel, e eingaben) { e.bindings[0].Path, e.bindings[0].Target = "X", "X" }},
	}
	for _, a := range aendere {
		m, eingabe := leseSicht()
		soll := ansicht(m)
		a.f(m, eingabe)
		if got := ansicht(m); got != soll {
			t.Errorf("%s verändert das Modell:\n%s\n%s", a.name, got, soll)
		}
	}
}

func TestServiceModelNullwerteUndNamen(t *testing.T) {
	m, _ := leseSicht()
	for _, name := range []string{"", "Self.A", "N.Z", "A", "n.a"} {
		if e, ok := m.EntityType(name); ok || !slices.Equal(e.Keys, nil) || e.Name != "" {
			t.Errorf("EntityType(%q): %+v %v", name, e, ok)
		}
		if s, ok := m.EntitySet(name); ok || s.Name != "" {
			t.Errorf("EntitySet(%q): %+v %v", name, s, ok)
		}
	}
	var leer model.ServiceModel
	if leer.Version() != "" || len(leer.Namespaces())+len(leer.EntityTypes())+len(leer.EntitySets())+len(leer.NavigationBindings()) != 0 {
		t.Errorf("Nullwert ist nicht leer: %s", ansicht(leer))
	}
	if _, ok := leer.EntityType(""); ok {
		t.Error("EntityType(\"\") des Nullwerts")
	}
}

func TestUnterstuetzteEDMTypen(t *testing.T) {
	acht := []string{"Edm.String", "Edm.Boolean", "Edm.Int32", "Edm.Int64", "Edm.Decimal", "Edm.Guid", "Edm.Date", "Edm.DateTimeOffset"}
	if got := model.UnterstuetzteEDMTypen(); !slices.Equal(got, acht) {
		t.Errorf("%v", got)
	}
	for _, ty := range append(acht, "Edm.Double", "Edm.Untyped", "Collection(Edm.String)", "", "edm.string", "N.A") {
		if got := (model.Property{EdmType: ty}).Unterstuetzt(); got != slices.Contains(acht, ty) {
			t.Errorf("Unterstuetzt(%q) = %v", ty, got)
		}
	}
}

// TestMetadatenFehlerForm (Zeile H1): Text, Code und Unwrap.
func TestMetadatenFehlerForm(t *testing.T) {
	const kopf = "ODR-E6004: Metadaten ungültig oder nicht unterstützt: "
	for _, w := range []struct {
		e    *model.MetadatenFehler
		text string
	}{
		{&model.MetadatenFehler{Ursache: model.UrsacheSyntax}, kopf + "syntax"},
		{&model.MetadatenFehler{Ursache: model.UrsacheStruktur, Ort: "Ns.T"}, kopf + `struktur "Ns.T"`},
		{&model.MetadatenFehler{Ursache: model.UrsacheStruktur, Ort: "a\"b\n"}, kopf + `struktur "a\"b\n"`},
	} {
		var err error = w.e
		if w.e.Error() != w.text {
			t.Errorf("Text %q, erwartet %q", w.e.Error(), w.text)
		}
		if w.e.Code() != model.CodeE6004 {
			t.Errorf("Code %v, erwartet ODR-E6004", w.e.Code())
		}
		if !errors.Is(err, model.ErrMetadatenUngueltig) || errors.Unwrap(err) != model.ErrMetadatenUngueltig {
			t.Errorf("%q: ErrMetadatenUngueltig über errors.Is oder Unwrap nicht erreichbar", w.e.Error())
		}
		var mf *model.MetadatenFehler
		if !errors.As(fmt.Errorf("x: %w", err), &mf) || mf != w.e {
			t.Errorf("errors.As über Verpackung")
		}
	}
}
