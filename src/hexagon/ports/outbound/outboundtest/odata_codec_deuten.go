package outboundtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// kolMit ist der Body einer Collection von R, dessen Wurzel Mitglieder vor und hinter `value` trägt.
func kolMit(vor, nach string, elems ...string) []byte {
	return []byte(`{` + kontextAn("Products", false) + `,` + vor + `"value":[` + strings.Join(elems, ",") + `]` + nach + `}`)
}

func jfC1() []jfall {
	var out []jfall
	for _, w := range []string{`[]`, `[{}]`, `1`, `"a"`, `true`, `false`, `null`} {
		out = append(out, roh("C1 Wurzel "+w, []byte(w), model.DatenStruktur)...)
	}
	for _, f := range roh("C1 {}", []byte(`{}`), model.DatenKontext) {
		f.ort = "/@odata.context"
		out = append(out, f)
	}
	return out
}

func jC1(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfC1()) }

func jfC2() []jfall {
	var out []jfall
	kolOhne := func(w string) []byte { return []byte(`{` + kontextAn("Products", false) + w + `}`) }
	out = append(out, jfall{name: "C2 value fehlt", m: "R", set: "Products", body: kolOhne(``), ursache: model.DatenStruktur, ort: "/value"})
	for _, v := range []string{`null`, `{}`, `{"ID":1}`, `"a"`, `1`, `true`, `false`} {
		out = append(out, jfall{name: "C2 value " + v, m: "R", set: "Products", body: kolOhne(`,"value":` + v), ursache: model.DatenStruktur, ort: "/value"})
	}
	gut := func(i int) string { return fmt.Sprintf(`{"ID":%d}`, i) }
	for _, el := range []string{`null`, `5`, `"x"`, `[]`, `true`, `[{"ID":9}]`} {
		for i := range 3 {
			elems := []string{gut(1), gut(2), gut(3)}
			elems[i] = el
			out = append(out, jfall{name: fmt.Sprintf("C2 Element %s an Stelle %d", el, i), m: "R", set: "Products", body: jKol("Products", elems...),
				ursache: model.DatenStruktur, ort: fmt.Sprintf("/value/%d", i)})
		}
	}
	for _, mitglied := range []struct{ name, json string }{
		{"other", `"other":1`}, {"Name@odata.type", `"Name@odata.type":"x"`}, {"value@odata.count", `"value@odata.count":1`}, {"ID", `"ID":1`},
	} {
		out = append(out,
			jfall{name: "C2 Wurzelmitglied " + mitglied.name + " vor value", m: "R", set: "Products", body: kolMit(mitglied.json+`,`, ``),
				ursache: model.DatenStruktur, ort: "/" + mitglied.name},
			jfall{name: "C2 Wurzelmitglied " + mitglied.name + " hinter value", m: "R", set: "Products", body: kolMit(``, `,`+mitglied.json),
				ursache: model.DatenStruktur, ort: "/" + mitglied.name})
	}
	leer := jfall{name: "C2 leeres value", m: "R", set: "Products", body: jKol("Products")}
	leer.nach = func(t *testing.T, r jergebnis) {
		if r.seite.Entities == nil || len(r.seite.Entities) != 0 || r.seite.FolgeLink != (model.FolgeLink{}) || r.seite.Anzahl != (model.Anzahl{}) {
			t.Errorf("leeres value: %+v, erwartet leere, nicht nil Entities und Nullwerte der Nachweise", r.seite)
		}
	}
	return append(out, leer)
}

func jC2(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfC2()) }

func jfC3() []jfall {
	kontext := func(name string, f jfall) jfall {
		f.name, f.ursache, f.ort = name, model.DatenKontext, "/@odata.context"
		return f
	}
	out := []jfall{
		kontext("C3 Collection-Body an ParseEntity", jfall{m: "R", set: "Products", einzel: true, body: []byte(jKollR)}),
		kontext("C3 Entity-Body an ParseCollection", jfall{m: "R", set: "Products", body: []byte(jEinzelR)}),
		{name: "C3 Property value bei ParseEntity", m: "W", set: "Ws", einzel: true, body: jEin("Ws", `"Id":1,"value":"x"`),
			props: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("value", model.ArtString, "x")}},
		kontext("C3 Property value, Entity-Body an ParseCollection", jfall{m: "W", set: "Ws", body: jEin("Ws", `"Id":1,"value":"x"`)}),
	}
	for _, n := range []string{"@odata.count", "@odata.nextLink"} {
		out = append(out, jfall{name: "C3 " + n + " an der Entity", m: "R", set: "Products", einzel: true, body: jEin("Products", `"ID":1,"`+n+`":1`),
			ursache: model.DatenStruktur, ort: "/" + n})
	}
	return out
}

func jC3(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfC3()) }

func jfD1() []jfall {
	var out []jfall
	pfade := []string{"$metadata", "/odata/$metadata", "../$metadata", "../../$metadata", "http://h/svc/$metadata", "https://h:8443/a/b/$metadata", "/$metadata"}
	fragmente := []string{"Products", "Products(ID)", "Products(ID,Name)", "Products(*)", "Products(ID,*)", "Products(ID,Category)", "Products(ID,ID)"}
	for _, pf := range pfade {
		for _, fr := range fragmente {
			kol := []byte(`{"@odata.context":"` + pf + `#` + fr + `","value":[{"ID":1}]}`)
			ein := []byte(`{"@odata.context":"` + pf + `#` + fr + `/$entity","ID":1}`)
			out = append(out,
				jfall{name: "D1 " + pf + "#" + fr + " (Collection)", m: "R", set: "Products", body: kol, schluessel: "1"},
				jfall{name: "D1 " + pf + "#" + fr + " (Entity)", m: "R", set: "Products", einzel: true, body: ein, schluessel: "1"})
		}
	}
	return append(out,
		jfall{name: "D1 K Lines(OrderID,Note)", m: "K", set: "Lines", body: []byte(`{"@odata.context":"$metadata#Lines(OrderID,Note)","value":[{"OrderID":1,"LineNo":2}]}`), schluessel: "OrderID=1,LineNo=2"},
		jfall{name: "D1 geerbte Property in der Auswahl", m: "T", set: "Ders", body: []byte(`{"@odata.context":"$metadata#Ders(Id,Name)","value":[{"Id":1}]}`), schluessel: "1"},
		jfall{name: "D1 Z Animals(Id,Name)", m: "Z", set: "Animals", body: []byte(`{"@odata.context":"$metadata#Animals(Id,Name)","value":[]}`)})
}

func jD1(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfD1()) }

// d2Kontexte sind die Kontexte, die je eine Minimalabweichung vom gültigen sind; sie gelten als Text der Collection, die Entity trägt `/$entity` dahinter.
func d2Kontexte() []string {
	return []string{
		`Products`, `x$metadata#Products`, `$metadata/x#Products`, `$metadata#`, `$metadata#Nope`, `$metadata#products`, `$metadata#Demo.Products`,
		`$metadata#Products2`, `$metadata#Products()`, `$metadata#Products(ID,)`, `$metadata#Products(,ID)`, `$metadata#Products(Nope)`, `$metadata#Products(ID, Name)`,
		`$metadata# Products`, `$metadata#Products(1)`, `$metadata#Products/Demo.Product`, `$metadata#Collection(Edm.String)`, `$metadata#Edm.String`,
		`$metadata#$ref`, `$metadata#Products#x`, `$metadata#Products(ID)x`, `$metadata#Products(ID`, `$metadata#Products)`, `$metadata#Products(ID)(Name)`,
		`#Products`, ``, `$metadata`, `/$metadata`, `a#/$metadata#Products`, `$metadata #Products`, `$metadata#Container/Products`, `$metadata#Products(id)`, `$metadata#Products(Category/ID)`, `$metadata#Products ()`,
	}
}

func jfD2() []jfall {
	var out []jfall
	add := func(name string, f jfall) {
		f.name, f.ursache, f.ort = name, model.DatenKontext, "/@odata.context"
		out = append(out, f)
	}
	for _, c := range d2Kontexte() {
		add("D2 Collection "+c, jfall{m: "R", set: "Products", body: []byte(`{"@odata.context":"` + c + `","value":[]}`)})
		add("D2 Entity "+c, jfall{m: "R", set: "Products", einzel: true, body: []byte(`{"@odata.context":"` + c + `/$entity","ID":1}`)})
	}
	for _, w := range []string{`5`, `null`, `true`, `{}`, `[]`, `""`, `["$metadata#Products"]`} {
		add("D2 Collection Kontext "+w, jfall{m: "R", set: "Products", body: []byte(`{"@odata.context":` + w + `,"value":[]}`)})
		add("D2 Entity Kontext "+w, jfall{m: "R", set: "Products", einzel: true, body: []byte(`{"@odata.context":` + w + `,"ID":1}`)})
	}
	add("D2 Collection ohne Kontext", jfall{m: "R", set: "Products", body: []byte(`{"value":[]}`)})
	add("D2 Entity ohne Kontext", jfall{m: "R", set: "Products", einzel: true, body: []byte(`{"ID":1}`)})
	add("D2 /$entity bei ParseCollection", jfall{m: "R", set: "Products", body: []byte(`{"@odata.context":"$metadata#Products/$entity","value":[]}`)})
	add("D2 kein /$entity bei ParseEntity", jfall{m: "R", set: "Products", einzel: true, body: []byte(`{"@odata.context":"$metadata#Products","ID":1}`)})
	add("D2 Rest hinter /$entity", jfall{m: "R", set: "Products", einzel: true, body: []byte(`{"@odata.context":"$metadata#Products/$entity/x","ID":1}`)})
	add("D2 /$entity mit Leerraum", jfall{m: "R", set: "Products", einzel: true, body: []byte(`{"@odata.context":"$metadata#Products/$entity ","ID":1}`)})
	add("D2 Auswahl mit Property eines abgeleiteten Typs", jfall{m: "Z", set: "Animals", body: []byte(`{"@odata.context":"$metadata#Animals(Id,Breed)","value":[]}`)})
	add("D2 anderes Set des Modells", jfall{m: "R", set: "Products", body: []byte(`{"@odata.context":"$metadata#Categories","value":[]}`)})
	return out
}

func jD2(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfD2()) }
