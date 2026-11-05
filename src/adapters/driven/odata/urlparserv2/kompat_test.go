package urlparserv2_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestAllaussageKompat (Zeile 47): jede V2-Eigenschaft außerhalb des Umfangs, die der Parser kennt, ist ein KompatFehler mit der
// erwarteten Ursache, nie ein URLFehler, nie ein Erfolg (SPEC-170 Nr. 3). Menge 63 Elemente aus den Listen des Tests (12 Optionen,
// 15 Pfadformen, 10 Formate, 25 Literalformen, 1 `$filter` ohne Leser) × 3 Einbettungen = 189.
func TestAllaussageKompat(t *testing.T) {
	p := urlparserv2.New()
	n := 0
	ursachen := map[model.KompatUrsache]int{}
	// query prüft ein Element einer Option in den drei Einbettungen: allein, hinter einem gültigen Pfad und einer gültigen Option,
	// vor einer zweiten gültigen Option. pfad ist der Pfad des Elements, ort der Ort seiner Ursache.
	query := func(pfad, element, ort string, k model.KompatUrsache) {
		for i, q := range []string{element, "$top=1&" + element, element + "&$skip=1"} {
			pf, o := pfad, ort
			if i == 1 {
				pf = "/S(1)"
				if ort == "query/1" {
					o = "query/2"
				}
			}
			pruefe(t, p, pf, q, kmp(k, o))
			ursachen[k]++
			n++
		}
	}
	for _, o := range ausserhalbOptionen() {
		query("/S", o+"=x", "query/1", model.KompatOption)
	}
	for _, f := range ausserhalbFormate() {
		query("/S", "$format="+f, "query/1", model.KompatFormat)
	}
	query("/S", "$filter=x", "query/1", model.KompatFilter)
	pfadElement := func(pfad, ort string, k model.KompatUrsache) {
		for _, q := range []string{"", "$top=1", "$top=1&$skip=1"} {
			pruefe(t, p, pfad, q, kmp(k, ort))
			ursachen[k]++
			n++
		}
	}
	for _, e := range ausserhalbPfade() {
		pfadElement(e.pfad, e.ort, model.KompatPfad)
	}
	for _, l := range ausserhalbLiterale() {
		pfadElement("/S("+l+")", "pfad/1", model.KompatLiteralTyp)
	}
	if n != 189 || len(ursachen) != 5 {
		t.Errorf("%d Fälle und %d Ursachen, erwartet 189 und 5: %v", n, len(ursachen), ursachen)
	}
	for k, z := range map[model.KompatUrsache]int{model.KompatOption: 36, model.KompatFormat: 30, model.KompatFilter: 3, model.KompatPfad: 45,
		model.KompatLiteralTyp: 75} {
		if ursachen[k] != z {
			t.Errorf("Ursache %s: %d Fälle, erwartet %d", k, ursachen[k], z)
		}
	}
}

// TestEchteV2URLs (Zeile 49): die 29 verschiedenen Anfragen der Erfassung am Referenzdienst Olingo (Referenzszenario, 45
// Anfragen, der Header `Accept` ändert die URL nicht) ergeben 18 Ergebnisse, 5 Kompatibilität `pfad`, 1 `option` und 5 `filter`.
func TestEchteV2URLs(t *testing.T) {
	zk := func(s string) model.SchluesselTeil { return teil("", model.LiteralZeichenkette, s) }
	top2 := model.ZahlOption{Gesetzt: true, Wert: 2}
	aeltere := model.Optionen{Top: top2}
	tabelle := []struct {
		pfad, q string
		want    erw
	}{
		{"/", "", kmp(model.KompatPfad, "pfad")},
		{"/$metadata", "", okR(model.Ressource{Art: model.RessourceMetadaten})},
		{"/Employees", "", okR(mengeR("Employees"))},
		{"/Employees", "$format=json", okR(mengeR("Employees"))},
		{"/Employees('1')", "", okR(entityR("Employees", zk("'1'")))},
		{"/Employees/$count", "", okR(anzahlR("Employees"))},
		{"/Employees('1')/EmployeeName/$value", "", kmp(model.KompatPfad, "pfad/2")},
		{"/Employees('1')/Age/$value", "", kmp(model.KompatPfad, "pfad/2")},
		{"/Employees('1')/$value", "", kmp(model.KompatPfad, "pfad/2")},
		{"/Employees", "$top=2&$format=json", ok(mengeR("Employees"), aeltere)},
		{"/Employees", "$skip=2&$top=2&$format=json", ok(mengeR("Employees"), model.Optionen{Top: top2, Skip: top2})},
		{"/Employees", "$filter=Age%20gt%2030&$format=json", kmp(model.KompatFilter, "query/1")},
		{"/Employees", "$orderby=EmployeeName%20desc&$format=json",
			ok(mengeR("Employees"), model.Optionen{OrderBy: []model.Sortierung{{Name: "EmployeeName", Absteigend: true}}})},
		{"/Employees", "$select=EmployeeName,Age&$format=json",
			ok(mengeR("Employees"), model.Optionen{Select: &model.SelectOption{Namen: []string{"EmployeeName", "Age"}}})},
		{"/Employees('1')", "$expand=ne_Team&$format=json", kmp(model.KompatOption, "query/1")},
		{"/Employees", "$inlinecount=allpages&$top=2&$format=json",
			ok(mengeR("Employees"), model.Optionen{Top: top2, Count: model.WahrOption{Gesetzt: true, Wert: true}})},
		{"/Employees", "$inlinecount=allpages&$top=2",
			ok(mengeR("Employees"), model.Optionen{Top: top2, Count: model.WahrOption{Gesetzt: true, Wert: true}})},
		{"/Employees", "$filter=Age%20gt%20999&$format=json", kmp(model.KompatFilter, "query/1")},
		{"/Employees", "$filter=Age%20gt%20999", kmp(model.KompatFilter, "query/1")},
		{"/Employees('999')", "", okR(entityR("Employees", zk("'999'")))},
		{"/Employees", "$filter=Age%20bogus%2030", kmp(model.KompatFilter, "query/1")},
		{"/Nope", "", okR(mengeR("Nope"))},
		{"/Teams", "$format=json", okR(mengeR("Teams"))},
		{"/Rooms", "$format=json&$top=1", ok(mengeR("Rooms"), model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1}})},
		{"/Employees('1')/ne_Team", "$format=json", kmp(model.KompatPfad, "pfad/2")},
		{"/Employees/$count", "$filter=Age%20gt%2030", kmp(model.KompatFilter, "query/1")},
		{"/Rooms('1')", "", okR(entityR("Rooms", zk("'1'")))},
		{"/Rooms", "", okR(mengeR("Rooms"))},
		{"/Buildings(1)", "", okR(entityR("Buildings", zahlT("1")))},
	}
	p := urlparserv2.New()
	arten := map[string]int{}
	for _, f := range tabelle {
		pruefe(t, p, f.pfad, f.q, f.want)
		switch {
		case f.want.url != nil:
			arten["erfolg"]++
		default:
			arten[string(f.want.k)]++
		}
	}
	if len(tabelle) != 29 || arten["erfolg"] != 18 || arten["pfad"] != 5 || arten["option"] != 1 || arten["filter"] != 5 {
		t.Errorf("%d Anfragen, Arten %v; erwartet 29 mit 18, 5, 1, 5", len(tabelle), arten)
	}
}
