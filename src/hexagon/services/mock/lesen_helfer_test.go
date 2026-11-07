package mock_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Die Sets des Testmodells: Products (Typ Product, mit abgeleitetem Typ Special), Kunden (Typ Product), Sks (ein Schlüssel,
// nullbar), Pairs (zwei Schlüssel, nullbar), Bunt (ein Schlüssel Edm.Int32 und je ein Schlüssel der übrigen Typen).
const (
	setProducts = "Products"
	baselineID  = "b1"
	typProduct  = "N.Product"
	typSpecial  = "N.Special"
)

// Die Namen der Properties je EDM-Typ in Product.
func propJeTyp(edm string) string {
	return map[string]string{
		"Edm.String": "Name", "Edm.Boolean": "Flag", "Edm.Int32": "Count", "Edm.Int64": "Big",
		"Edm.Decimal": "Price", "Edm.Guid": "Ref", "Edm.Date": "Day", "Edm.DateTimeOffset": "Ts",
	}[edm]
}

// achtTypen sind die acht unterstützten EDM-Typen in der Folge von model.UnterstuetzteEDMTypen().
func achtTypen() []string { return model.UnterstuetzteEDMTypen() }

// produktProperties sind die wirksamen Properties von Product: der Schlüssel Id und eine Property je EDM-Typ.
func produktProperties() []model.Property {
	ps := []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}
	for _, t := range achtTypen() {
		ps = append(ps, model.Property{Name: propJeTyp(t), EdmType: t, Nullable: true})
	}
	return ps
}

// nichtUnterstuetzteTypen sind die EDM-Typen (und ein Typ außerhalb von Edm), für die Property.Unterstuetzt() falsch ist; die
// Zählprüfung steht in TestNichtUnterstuetzteMenge.
func nichtUnterstuetzteTypen() []string {
	return []string{"Edm.Binary", "Edm.Byte", "Edm.SByte", "Edm.Int16", "Edm.Double", "Edm.Single", "Edm.Duration", "Edm.TimeOfDay",
		"Edm.Stream", "Edm.GeographyPoint", "Edm.GeometryPoint", "N.Komplex"}
}

// vielProp ist der Name der Property von Viele zu einem nicht unterstützten Typ.
func vielProp(edm string) string { return "P" + strings.ReplaceAll(edm, ".", "_") }

// testModell ist das Modell der Lesetests (SPEC-125 Nr. 11), mit den Typen Product, Special (abgeleitet), Abstrakt, Fremd,
// Sk, Pair und den Sets von oben.
func testModell() model.ServiceModel {
	prod := produktProperties()
	special := append(append([]model.Property{}, prod...), model.Property{Name: "Extra", EdmType: "Edm.String", Nullable: true})
	viele := []model.Property{{Name: "Id", EdmType: "Edm.Int32"}}
	for _, t := range nichtUnterstuetzteTypen() {
		viele = append(viele, model.Property{Name: vielProp(t), EdmType: t, Nullable: true})
	}
	nullbar := func(n string) model.Property { return model.Property{Name: n, EdmType: "Edm.String", Nullable: true} }
	typen := []model.EntityType{
		{Name: typProduct, Keys: []string{"Id"}, Properties: prod},
		{Name: typSpecial, BaseType: typProduct, Keys: []string{"Id"}, Properties: special},
		{Name: "N.Abstrakt", BaseType: typProduct, Abstract: true, Keys: []string{"Id"}, Properties: prod},
		{Name: "N.SpezAbstrakt", BaseType: typSpecial, Abstract: true, Keys: []string{"Id"}, Properties: special},
		{Name: "N.Tief", BaseType: typSpecial, Keys: []string{"Id"}, Properties: special},
		{Name: "N.Viele", Keys: []string{"Id"}, Properties: viele},
		{Name: "N.Fremd", Keys: []string{"Id"}, Properties: prod},
		{Name: "N.Sk", Keys: []string{"K"}, Properties: []model.Property{nullbar("K"), nullbar("Name")}},
		{Name: "N.Pair", Keys: []string{"A", "B"}, Properties: []model.Property{nullbar("A"), nullbar("B"), nullbar("Name")}},
		{Name: "N.Seltsam", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Blob", EdmType: "Edm.Binary", Nullable: true}},
			NavigationProperties: []model.NavigationProperty{{Name: "Nav", Target: typProduct}}},
		{Name: "N.Bunt", Keys: []string{"I", "S"}, Properties: []model.Property{
			{Name: "I", EdmType: "Edm.Int32"}, {Name: "S", EdmType: "Edm.String"}, {Name: "Name", EdmType: "Edm.String", Nullable: true}}},
		{Name: "N.Drei", Keys: []string{"Id"}, Properties: []model.Property{{Name: "Id", EdmType: "Edm.Int32"}, nullbar("S1"), nullbar("S2"), nullbar("S3")}},
	}
	sets := []model.EntitySet{
		{Name: setProducts, EntityType: typProduct}, {Name: "Kunden", EntityType: typProduct}, {Name: "Sks", EntityType: "N.Sk"},
		{Name: "Pairs", EntityType: "N.Pair"}, {Name: "Bunt", EntityType: "N.Bunt"}, {Name: "Seltsam", EntityType: "N.Seltsam"},
		{Name: "Viele", EntityType: "N.Viele"}, {Name: "Waise", EntityType: "N.GibtEsNicht"}, {Name: "Drei", EntityType: "N.Drei"},
	}
	return model.NeuesServiceModel("4.0", []string{"N"}, typen, sets, nil)
}

// alleNamen sind die Namen aller Properties von Product, in der Folge des Modells.
func alleNamen() []string {
	var r []string
	for _, p := range produktProperties() {
		r = append(r, p.Name)
	}
	return r
}

// produkt ist ein vollständiger Eintrag von Products mit Id und Name; known nennt die bekannten Properties (Vorgabe: alle).
func produkt(id int, name string, known ...string) model.EntityState {
	if len(known) == 0 {
		known = alleNamen()
	}
	return model.EntityState{
		WorkspaceID: wsID, EntitySet: setProducts, EntityKey: strconv.Itoa(id), Completeness: model.EntityComplete,
		Payload: fmt.Sprintf(`{"Id":%d,"Name":%q}`, id, name), KnownProperties: known,
	}
}

// eintrag ist ein Eintrag mit freiem Payload.
func eintrag(set, key, payload string, known ...string) model.EntityState {
	return model.EntityState{WorkspaceID: wsID, EntitySet: set, EntityKey: key, Payload: payload, Completeness: model.EntityComplete, KnownProperties: known}
}

// leseUmg ist ein Aufbau mit dem Testmodell, der Coverage complete für Products und drei Produkten.
func leseUmg(t *testing.T) *umg {
	t.Helper()
	u := neueUmg(t)
	u.c.mc.modell = testModell()
	u.co.setze(baselineID, setProducts, model.CoverageComplete)
	u.en.setze(setProducts, produkt(1, "a"), produkt(2, "b"), produkt(3, "c"))
	return u
}

// ressource setzt den Fake des URL-Lesers auf eine feste Ressource und feste Optionen.
func (u *umg) ressource(art model.RessourcenArt, set string, opt model.Optionen, teile ...model.SchluesselTeil) {
	u.c.url.f = func(string, string) (model.ODataURL, error) {
		return model.ODataURL{Ressource: model.Ressource{Art: art, EntitySet: set, Schluessel: teile}, Optionen: opt}, nil
	}
}

// menge, anzahl und entity setzen die Ressource von Products.
func (u *umg) menge(opt model.Optionen)  { u.ressource(model.RessourceMenge, setProducts, opt) }
func (u *umg) anzahl(opt model.Optionen) { u.ressource(model.RessourceAnzahl, setProducts, opt) }
func (u *umg) entity(opt model.Optionen, id string) {
	u.ressource(model.RessourceEntity, setProducts, opt, zahl("", id))
}

// zahl ist ein Schlüsselteil mit einem Zahlenliteral.
func zahl(name, text string) model.SchluesselTeil {
	return model.SchluesselTeil{Name: name, Literal: model.Literal{Art: model.LiteralZahl, Text: text}}
}

// lies ruft GET (oder HEAD) auf den Pfad des Dienstes.
func (u *umg) leseGET() (inbound.Antwort, error) { return u.rufe(anfrage("GET", dienstRoot+"/x")) }

// letzteCollection ist die CollectionAntwort des einen Aufrufs von SchreibeCollection.
func (u *umg) letzteCollection() model.CollectionAntwort {
	u.t.Helper()
	ev := u.p.von("Schreiber.SchreibeCollection")
	if len(ev) != 1 {
		u.t.Fatalf("%d Aufrufe von SchreibeCollection, erwartet 1", len(ev))
	}
	return ev[0].args[2].(model.CollectionAntwort)
}

// letzteEntity ist die EntityAntwort des einen Aufrufs von SchreibeEntity.
func (u *umg) letzteEntity() model.EntityAntwort {
	u.t.Helper()
	ev := u.p.von("Schreiber.SchreibeEntity")
	if len(ev) != 1 {
		u.t.Fatalf("%d Aufrufe von SchreibeEntity, erwartet 1", len(ev))
	}
	return ev[0].args[2].(model.EntityAntwort)
}

// letzteAnzahl ist die Zahl des Aufrufs von SchreibeAnzahl für die Anzahl (nicht die Probe von $metadata, n = 0 ohne dst).
func (u *umg) letzteAnzahl() int64 {
	u.t.Helper()
	ev := u.p.von("Schreiber.SchreibeAnzahl")
	if len(ev) != 1 {
		u.t.Fatalf("%d Aufrufe von SchreibeAnzahl, erwartet 1", len(ev))
	}
	return ev[0].args[1].(int64)
}

// ids nennt die Schlüssel der Entities, kommagetrennt.
func ids(es []model.Entity) string {
	var r []string
	for _, e := range es {
		r = append(r, e.Schluessel)
	}
	return strings.Join(r, ",")
}

// eigenschaft ist der Filter `<name> <op> <literal>` (Zahl oder Zeichenkette mit Anführungszeichen).
func vergleich(op model.VergleichOp, name string, lit model.Literal) model.FilterVergleich {
	return model.FilterVergleich{Op: op, Links: model.FilterEigenschaft{Name: name}, Rechts: model.FilterLiteral{Literal: lit}}
}

func filterOpt(a model.FilterAusdruck) *model.FilterOption {
	return &model.FilterOption{Text: "x", Ausdruck: a}
}

// idGroesser ist der Filter `Id gt n`.
func idGroesser(n string) *model.FilterOption {
	return filterOpt(vergleich(model.VergleichGt, "Id", model.Literal{Art: model.LiteralZahl, Text: n}))
}

// nameGleich ist der Filter `Name eq '<s>'`.
func nameGleich(s string) model.FilterVergleich {
	return vergleich(model.VergleichEq, "Name", model.Literal{Art: model.LiteralZeichenkette, Text: "'" + s + "'"})
}

// fehlercode zerlegt err in den Code (0 ohne Fehler und für einen Fehler ohne Code).
func fehlercode(err error) model.Code {
	if f, ok := err.(*inbound.Fehler); ok {
		return f.Code
	}
	return 0
}
