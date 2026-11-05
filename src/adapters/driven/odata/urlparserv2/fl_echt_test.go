package urlparserv2_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestFlEchteV2Filter (Zeile 37): die fünf `filter`-Anfragen der Erfassung am Referenzdienst Olingo aus `TestEchteV2URLs` mit
// dem Leser (vier Bäume, `Age bogus 30` ist `wert`), dazu zwölf Ausdrücke in der Schreibweise von V2. Menge 17.
func TestFlEchteV2Filter(t *testing.T) {
	age := func(zahl string) flBaum { return flBin("gt", flEig("Age"), flZahl(zahl)) }
	opt := func(text string, baum flBaum) model.Optionen {
		return model.Optionen{Filter: &model.FilterOption{Text: text, Ausdruck: baum}}
	}
	guid := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	tabelle := []struct {
		pfad, text, zusatz string
		want               erw
	}{
		{"/Employees", "Age gt 30", "&$format=json", ok(mengeR("Employees"), opt("Age gt 30", age("30")))},
		{"/Employees", "Age gt 999", "&$format=json", ok(mengeR("Employees"), opt("Age gt 999", age("999")))},
		{"/Employees", "Age gt 999", "", ok(mengeR("Employees"), opt("Age gt 999", age("999")))},
		{"/Employees", "Age bogus 30", "", ung(model.URLWert, "query/1")},
		{"/Employees/$count", "Age gt 30", "", ok(anzahlR("Employees"), opt("Age gt 30", age("30")))},
		{"/Employees", "substringof('a',Name) eq true", "", kmp(model.KompatFilter, "query/1")},
		{"/Employees", "Date gt datetime'2020-01-01T00:00:00'", "",
			ok(mengeR("Employees"), opt("Date gt datetime'2020-01-01T00:00:00'", flBin("gt", flEig("Date"), flLit(model.LiteralZeitpunkt, "2020-01-01T00:00:00Z"))))},
		{"/Employees", "Id eq guid'" + guid + "'", "",
			ok(mengeR("Employees"), opt("Id eq guid'"+guid+"'", flBin("eq", flEig("Id"), flLit(model.LiteralGUID, guid))))},
		{"/Employees", "Price gt 1.5M", "", ok(mengeR("Employees"), opt("Price gt 1.5M", flBin("gt", flEig("Price"), flZahl("1.5"))))},
		{"/Employees", "Big eq 5L", "", ok(mengeR("Employees"), opt("Big eq 5L", flBin("eq", flEig("Big"), flZahl("5"))))},
		{"/Employees", "not (Age gt 3)", "", ok(mengeR("Employees"), opt("not (Age gt 3)", flNicht(age("3"))))},
		{"/Employees", "Name eq null", "", ok(mengeR("Employees"), opt("Name eq null", flBin("eq", flEig("Name"), flNull())))},
		{"/Employees", "Name eq 'O''Hara'", "", ok(mengeR("Employees"), opt("Name eq 'O''Hara'", flBin("eq", flEig("Name"), flZk("'O''Hara'"))))},
		{"/Employees", "Id eq 1 and Name ne 'x'", "", ok(mengeR("Employees"),
			opt("Id eq 1 and Name ne 'x'", flBin("and", flBin("eq", flEig("Id"), flZahl("1")), flBin("ne", flEig("Name"), flZk("'x'")))))},
		{"/Employees", "startswith(Name,'a') eq true", "", kmp(model.KompatFilter, "query/1")},
		{"/Employees", "Price add 1 gt 5", "", kmp(model.KompatFilter, "query/1")},
		{"/Employees", "Name eq 'x' or Id eq 1L", "", ok(mengeR("Employees"), opt("Name eq 'x' or Id eq 1L",
			flBin("or", flBin("eq", flEig("Name"), flZk("'x'")), flBin("eq", flEig("Id"), flZahl("1")))))},
	}
	p := flMitLeser()
	for _, f := range tabelle {
		pruefe(t, p, f.pfad, flQ(f.text)+f.zusatz, f.want)
	}
	if len(tabelle) != 17 {
		t.Fatalf("%d Anfragen, erwartet 17", len(tabelle))
	}
}
