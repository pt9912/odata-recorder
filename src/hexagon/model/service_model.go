package model

import "slices"

// Property ist eine Property eines Entity-Typs (SPEC-125 Nr. 11); EdmType ist nach Auflösung der
// Aliase qualifiziert.
type Property struct {
	Name     string
	EdmType  string
	Nullable bool
}

// UnterstuetzteEDMTypen liefert die acht unterstützten EDM-Typen in der Reihenfolge von SPEC-049
// (SPEC-125 Nr. 10).
func UnterstuetzteEDMTypen() []string {
	return []string{
		"Edm.String", "Edm.Boolean", "Edm.Int32", "Edm.Int64",
		"Edm.Decimal", "Edm.Guid", "Edm.Date", "Edm.DateTimeOffset",
	}
}

// Unterstuetzt ist wahr genau für die acht Typen von UnterstuetzteEDMTypen.
func (p Property) Unterstuetzt() bool {
	return slices.Contains(UnterstuetzteEDMTypen(), p.EdmType)
}

// NavigationProperty verweist auf einen Entity-Typ; Target ist sein qualifizierter Name ohne
// `Collection(…)`.
type NavigationProperty struct {
	Name       string
	Target     string
	Collection bool
}

// EntityType ist ein Entity-Typ mit seinen wirksamen Schlüsseln und Members, die geerbten
// eingeschlossen (SPEC-125 Nr. 11). Name und BaseType sind qualifiziert.
type EntityType struct {
	Name                 string
	BaseType             string
	Abstract             bool
	Open                 bool
	Keys                 []string
	Properties           []Property
	NavigationProperties []NavigationProperty
}

// EntitySet ist ein Entity-Set des Containers; EntityType ist der qualifizierte Name.
type EntitySet struct {
	Name       string
	EntityType string
}

// NavigationBinding bindet einen Pfad eines Entity-Sets an ein Ziel.
type NavigationBinding struct {
	Set    string
	Path   string
	Target string
}

// ServiceModel ist die Lese-Sicht auf das Modell eines OData-Dienstes (SPEC-125 Nr. 11): ein Wert
// ohne exportierte Felder, der nach der Rückgabe unverändert bleibt. Jeder Zugriff liefert eine
// Kopie; der Nullwert ist ein leeres Modell.
type ServiceModel struct {
	version    string
	namespaces []string
	typen      []EntityType
	typIndex   map[string]int
	sets       []EntitySet
	setIndex   map[string]int
	bindings   []NavigationBinding
}

func kopie[T any](s []T) []T { return append([]T{}, s...) }

func kopieTyp(t EntityType) EntityType {
	t.Keys = kopie(t.Keys)
	t.Properties = kopie(t.Properties)
	t.NavigationProperties = kopie(t.NavigationProperties)
	return t
}

// NeuesServiceModel bildet die Lese-Sicht aus Werten in der Reihenfolge des Dokuments; die
// Eingaben werden kopiert.
func NeuesServiceModel(version string, namespaces []string, typen []EntityType, sets []EntitySet, bindings []NavigationBinding) ServiceModel {
	m := ServiceModel{
		version: version, namespaces: kopie(namespaces), sets: kopie(sets), bindings: kopie(bindings),
		typIndex: map[string]int{}, setIndex: map[string]int{},
	}
	for i, t := range typen {
		m.typen = append(m.typen, kopieTyp(t))
		m.typIndex[t.Name] = i
	}
	for i, s := range sets {
		m.setIndex[s.Name] = i
	}
	return m
}

// Version liefert `4.0`; das leere Modell liefert den leeren Text.
func (m ServiceModel) Version() string { return m.version }

// Namespaces liefert die verschiedenen Namensräume in der Reihenfolge des Dokuments.
func (m ServiceModel) Namespaces() []string { return kopie(m.namespaces) }

// EntityTypes liefert die Entity-Typen in der Reihenfolge des Dokuments.
func (m ServiceModel) EntityTypes() []EntityType {
	out := make([]EntityType, len(m.typen))
	for i, t := range m.typen {
		out[i] = kopieTyp(t)
	}
	return out
}

// EntityType liefert den Typ zum qualifizierten Namen (nie zu einem Alias); ein unbekannter oder
// leerer Name liefert den Nullwert und false.
func (m ServiceModel) EntityType(qualifiedName string) (EntityType, bool) {
	i, ok := m.typIndex[qualifiedName]
	if !ok {
		return EntityType{}, false
	}
	return kopieTyp(m.typen[i]), true
}

// EntitySets liefert die Entity-Sets in der Reihenfolge des Dokuments.
func (m ServiceModel) EntitySets() []EntitySet { return kopie(m.sets) }

// EntitySet liefert das Set zum Namen; ein unbekannter oder leerer Name liefert den Nullwert und false.
func (m ServiceModel) EntitySet(name string) (EntitySet, bool) {
	i, ok := m.setIndex[name]
	if !ok {
		return EntitySet{}, false
	}
	return m.sets[i], true
}

// NavigationBindings liefert die Bindings in der Reihenfolge des Dokuments.
func (m ServiceModel) NavigationBindings() []NavigationBinding { return kopie(m.bindings) }
