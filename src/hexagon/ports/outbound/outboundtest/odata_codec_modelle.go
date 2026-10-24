package outboundtest

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// edmTypenJSON sind die EDM-Typen, die der Leser liest (SPEC-126 Nr. 10) und die die Literalmodelle T und W führen.
func edmTypenJSON() []string {
	return []string{"Edm.String", "Edm.Boolean", "Edm.Int32", "Edm.Int64", "Edm.Decimal"}
}

// kurz ist der Name des EDM-Typs ohne `Edm.`.
func kurz(edm string) string { return strings.TrimPrefix(edm, "Edm.") }

func p(name, edm string, nullable bool) model.Property {
	return model.Property{Name: name, EdmType: edm, Nullable: nullable}
}

// modellT hat je EDM-Typ einen Entity-Typ mit einem Schlüssel `Id` dieses Typs (Set `K<Typ>`), dazu die
// Typen mit zwei und drei Schlüsseln (Sets `K2`, `K3`) und einen abstrakten Typ ohne Schlüssel mit einer
// Ableitung, die einen hat (Sets `Abs`).
func modellT() model.ServiceModel {
	var typen []model.EntityType
	var sets []model.EntitySet
	for _, edm := range edmTypenJSON() {
		n := "T.K" + kurz(edm)
		typen = append(typen, model.EntityType{Name: n, Keys: []string{"Id"}, Properties: []model.Property{p("Id", edm, false)}})
		sets = append(sets, model.EntitySet{Name: "K" + kurz(edm), EntityType: n})
	}
	typen = append(typen,
		model.EntityType{Name: "T.K2", Keys: []string{"A", "B"}, Properties: []model.Property{
			p("A", "Edm.Int32", false), p("B", "Edm.String", false), p("Note", "Edm.String", true)}},
		model.EntityType{Name: "T.K3", Keys: []string{"A", "B", "C"}, Properties: []model.Property{
			p("A", "Edm.Int32", false), p("B", "Edm.String", false), p("C", "Edm.Boolean", false), p("Note", "Edm.String", true)}},
		model.EntityType{Name: "T.KN", Keys: []string{"Id"}, Properties: []model.Property{p("Id", "Edm.Int32", true)}},
		model.EntityType{Name: "T.KS", Keys: []string{"A", "B"}, Properties: []model.Property{p("A", "Edm.String", false), p("B", "Edm.String", false)}},
		model.EntityType{Name: "T.Base", Abstract: true, Properties: []model.Property{p("Name", "Edm.String", true)}},
		model.EntityType{Name: "T.Der", BaseType: "T.Base", Keys: []string{"Id"}, Properties: []model.Property{
			p("Name", "Edm.String", true), p("Id", "Edm.Int32", false)}},
	)
	sets = append(sets,
		model.EntitySet{Name: "K2", EntityType: "T.K2"}, model.EntitySet{Name: "K3", EntityType: "T.K3"}, model.EntitySet{Name: "KS", EntityType: "T.KS"}, model.EntitySet{Name: "KN", EntityType: "T.KN"},
		model.EntitySet{Name: "Abs", EntityType: "T.Base"}, model.EntitySet{Name: "Ders", EntityType: "T.Der"})
	return model.NeuesServiceModel("4.0", []string{"T"}, typen, sets, nil)
}

// modellW hat einen Typ `T.W` mit dem Schlüssel `Id` und je EDM-Typ zwei Properties, `N<Typ>` (nullable) und
// `R<Typ>` (nicht nullable), dazu Properties nicht unterstützter Typen und Navigationen (Set `Ws`); `T.WO` ist offen (Set `Wo`).
func modellW() model.ServiceModel {
	props := []model.Property{p("Id", "Edm.Int32", false)}
	for _, edm := range edmTypenJSON() {
		props = append(props, p("N"+kurz(edm), edm, true), p("R"+kurz(edm), edm, false))
	}
	props = append(props, p("value", "Edm.String", true), p("Dbl", "Edm.Double", true), p("Tags", "Collection(Edm.String)", true),
		p("Cx", "T.Complex", true), p("En", "T.Enum", true))
	navs := []model.NavigationProperty{{Name: "Nav", Target: "T.W"}, {Name: "Navs", Target: "T.W", Collection: true}}
	typen := []model.EntityType{
		{Name: "T.W", Keys: []string{"Id"}, Properties: props, NavigationProperties: navs},
		{Name: "T.WO", Keys: []string{"Id"}, Open: true, Properties: props, NavigationProperties: navs},
	}
	sets := []model.EntitySet{{Name: "Ws", EntityType: "T.W"}, {Name: "Wo", EntityType: "T.WO"}}
	return model.NeuesServiceModel("4.0", []string{"T"}, typen, sets, nil)
}

// paarTypen sind die EDM-Typen der Properties des Modells P, die je nach Stand der Typen wechseln.
func paarTypen() (part, shipped string) { return "Edm.Int64", "Edm.Int64" }

// modellP hat den Typ `T.P` mit dem Schlüssel `Part`, `Seq`, einem String, einem Datum, einer nicht nullbaren
// Ganzzahl und einer Dezimalzahl (Set `Ps`), die Properties der Stellen S1 bis S15 der Paare.
func modellP() model.ServiceModel {
	part, shipped := paarTypen()
	t := model.EntityType{Name: "T.P", Keys: []string{"Part", "Seq"}, Properties: []model.Property{
		p("Part", part, false), p("Seq", "Edm.Int32", false), p("Note", "Edm.String", true), p("Shipped", shipped, true),
		p("Req", "Edm.Int32", false), p("Amount", "Edm.Decimal", true),
	}}
	return model.NeuesServiceModel("4.0", []string{"T"}, []model.EntityType{t}, []model.EntitySet{{Name: "Ps", EntityType: "T.P"}}, nil)
}

// literalModelle sind die Modelle, die die Suite ohne den MetadataCodec baut.
func literalModelle() map[string]model.ServiceModel {
	return map[string]model.ServiceModel{"T": modellT(), "W": modellW(), "P": modellP()}
}
