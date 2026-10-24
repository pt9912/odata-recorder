package outboundtest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// produkt ist die Entity von R, die der Body `{"ID":<id>}` ergibt.
func produkt(id int) model.Entity {
	return model.Entity{Typ: "Demo.Product", Schluessel: strconv.Itoa(id), Properties: []model.PropertyWert{pw("ID", model.ArtInt32, strconv.Itoa(id))},
		Uebersprungen: []model.Uebersprungen{}}
}

// entitiesGleich prüft die gelesenen Entities gegen die erwarteten (bei ParseEntity die eine).
func entitiesGleich(t *testing.T, name string, r jergebnis, einzel bool, soll ...model.Entity) {
	t.Helper()
	got := r.seite.Entities
	if einzel {
		got = []model.Entity{r.entity}
	}
	if !reflect.DeepEqual(got, soll) {
		t.Errorf("%s: %+v, erwartet %+v", name, got, soll)
	}
}

// mitUeber setzt die erwarteten übersprungenen Mitglieder der Fälle.
func mitUeber(fs []jfall, ueber ...model.Uebersprungen) []jfall {
	for i := range fs {
		fs[i].ueber = ueber
	}
	return fs
}

// folgeFall ist ein Wert von `@odata.nextLink` mit dem erwarteten Ausgang.
type folgeFall struct {
	json    string
	soll    model.FolgeLink
	ursache model.DatenUrsache
}

func e1Faelle() []folgeFall {
	lang := strings.Repeat("a", 10000)
	l := func(w string) model.FolgeLink { return model.FolgeLink{Vorhanden: true, Wert: w} }
	return []folgeFall{
		{`"http://h/svc/Products?$skiptoken=3"`, l("http://h/svc/Products?$skiptoken=3"), ""}, {`"Products?$skiptoken=3"`, l("Products?$skiptoken=3"), ""},
		{`""`, model.FolgeLink{}, model.DatenWert}, {`"  a b  "`, l("  a b  "), ""}, {`"a?b=1` + esc("0026") + `c=2"`, l("a?b=1&c=2"), ""},
		{`"` + lang + `"`, l(lang), ""}, {`"[REDACTED]"`, model.FolgeLink{Vorhanden: true, Redigiert: true}, ""},
		{`"` + esc("005b") + `REDACTED]"`, l("[REDACTED]"), ""}, {`"[redacted]"`, l("[redacted]"), ""}, {`"[REDACTED] "`, l("[REDACTED] "), ""},
		{`" [REDACTED]"`, l(" [REDACTED]"), ""}, {`5`, model.FolgeLink{}, model.DatenTyp}, {`true`, model.FolgeLink{}, model.DatenTyp},
		{`null`, model.FolgeLink{}, model.DatenTyp}, {`{}`, model.FolgeLink{}, model.DatenTyp}, {`[]`, model.FolgeLink{}, model.DatenTyp},
	}
}

// rootMit setzt das Mitglied vor oder hinter `value` in die Collection.
func rootMit(mitglied string, hinter bool, elems ...string) []byte {
	if hinter {
		return kolMit("", ","+mitglied, elems...)
	}
	return kolMit(mitglied+",", "", elems...)
}

func jfE1() []jfall {
	var out []jfall
	for _, hinter := range []bool{false, true} {
		for _, f := range e1Faelle() {
			fall := jfall{name: fmt.Sprintf("E1 nextLink %.30s hinten=%v", f.json, hinter), m: "R", set: "Products", body: rootMit(`"@odata.nextLink":`+f.json, hinter, `{"ID":1}`)}
			if f.ursache != "" {
				fall.ursache, fall.ort = f.ursache, "/@odata.nextLink"
			} else {
				fall.nach = func(t *testing.T, r jergebnis) {
					if r.seite.FolgeLink != f.soll {
						t.Errorf("%s: %+v, erwartet %+v", fall.name, r.seite.FolgeLink, f.soll)
					}
				}
			}
			out = append(out, fall)
		}
	}
	fehlt := jfall{name: "E1 nextLink fehlt", m: "R", set: "Products", body: jKol("Products", `{"ID":1}`)}
	fehlt.nach = func(t *testing.T, r jergebnis) {
		if r.seite.FolgeLink != (model.FolgeLink{}) {
			t.Errorf("nextLink fehlt: %+v", r.seite.FolgeLink)
		}
	}
	// Nachweise vor einem fehlerhaften Element gehören nicht zu einem Teilergebnis: der Fehler liefert Nullwerte.
	teil := jfall{name: "E1 Nachweise vor einem fehlerhaften Element", m: "R", set: "Products", ursache: model.DatenTyp, ort: "/value/1/ID",
		body: kolMit(`"@odata.count":5,"@odata.nextLink":"x",`, ``, `{"ID":1}`, `{"ID":"x"}`)}
	return append(out, fehlt, teil)
}

func jE1(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfE1()) }

// zahlFall ist ein Wert von `@odata.count` mit dem erwarteten Ausgang.
type zahlFall struct {
	json    string
	soll    int64
	ursache model.DatenUrsache
}

func e2Faelle() []zahlFall {
	return []zahlFall{
		{`0`, 0, ""}, {`1`, 1, ""}, {`9223372036854775807`, 9223372036854775807, ""}, {`9223372036854775808`, 0, model.DatenWert},
		{`-1`, 0, model.DatenWert}, {`-0`, 0, ""}, {`1.0`, 1, ""}, {`1e2`, 100, ""}, {`1.5`, 0, model.DatenWert}, {`"5"`, 5, ""},
		{`"x"`, 0, model.DatenWert}, {`""`, 0, model.DatenWert}, {`"[REDACTED]"`, 0, model.DatenWert}, {`true`, 0, model.DatenTyp},
		{`null`, 0, model.DatenTyp}, {`{}`, 0, model.DatenTyp}, {`[]`, 0, model.DatenTyp}, {`1e999999999`, 0, model.DatenWert},
		{`-9223372036854775809`, 0, model.DatenWert}, {`"9223372036854775808"`, 0, model.DatenWert}, {`1e-1`, 0, model.DatenWert},
	}
}

func jfE2() []jfall {
	var out []jfall
	for _, hinter := range []bool{false, true} {
		for _, f := range e2Faelle() {
			fall := jfall{name: fmt.Sprintf("E2 count %s hinten=%v", f.json, hinter), m: "R", set: "Products", body: rootMit(`"@odata.count":`+f.json, hinter, `{"ID":1}`, `{"ID":2}`)}
			if f.ursache != "" {
				fall.ursache, fall.ort = f.ursache, "/@odata.count"
			} else {
				fall.nach = func(t *testing.T, r jergebnis) {
					if r.seite.Anzahl != (model.Anzahl{Vorhanden: true, Wert: f.soll}) || len(r.seite.Entities) != 2 {
						t.Errorf("%s: %+v, erwartet Anzahl %d bei zwei Elementen", fall.name, r.seite, f.soll)
					}
				}
			}
			out = append(out, fall)
		}
	}
	return out
}

func jE2(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfE2()) }

// e3Element ist der Body eines Elements von Animals mit der Angabe von `@odata.type` (leer: ohne Angabe).
func e3Element(id int, angabe string) string {
	if angabe == "" {
		return fmt.Sprintf(`{"Id":%d}`, id)
	}
	return fmt.Sprintf(`{"@odata.type":%s,"Id":%d}`, angabe, id)
}

type e3Fall struct {
	angabe string
	typ    string
	fehler bool
}

func e3Faelle() []e3Fall {
	return []e3Fall{
		{``, "Zoo.Animal", false}, {`"#Zoo.Mammal"`, "Zoo.Mammal", false}, {`"#Zoo.Dog"`, "Zoo.Dog", false}, {`"#Zoo.Puppy"`, "Zoo.Puppy", false},
		{`"#Zoo.Animal"`, "", true}, {`"#Zoo.Keeper"`, "", true}, {`"Zoo.Mammal"`, "", true}, {`"#Zoo.Nope"`, "", true},
		{`"#Zoo.Mammal "`, "", true}, {`" #Zoo.Mammal"`, "", true}, {`""`, "", true}, {`"#"`, "", true}, {`5`, "", true}, {`null`, "", true},
		{`{}`, "", true}, {`"[REDACTED]"`, "", true}, {`"#zoo.mammal"`, "", true}, {`["#Zoo.Mammal"]`, "", true},
	}
}

func jfE3() []jfall {
	var out []jfall
	for _, f := range e3Faelle() {
		for stelle := range 3 {
			elems := []string{e3Element(1, ""), e3Element(2, ""), e3Element(3, "")}
			elems[stelle] = e3Element(stelle+1, f.angabe)
			c := jfall{name: fmt.Sprintf("E3 %s an Element %d", f.angabe, stelle), m: "Z", set: "Animals", body: jKol("Animals", elems...)}
			if f.fehler {
				c.ursache, c.ort = model.DatenODataTyp, fmt.Sprintf("/value/%d/@odata.type", stelle)
			} else {
				c.nach = typDesElements(stelle, f.typ)
			}
			out = append(out, c)
		}
		members := strings.TrimPrefix(strings.TrimSuffix(e3Element(1, f.angabe), "}"), "{")
		e := jfall{name: "E3 " + f.angabe + " an der Entity", m: "Z", set: "Animals", einzel: true, body: jEin("Animals", members), typ: f.typ}
		if f.fehler {
			e.ursache, e.ort = model.DatenODataTyp, "/@odata.type"
		}
		out = append(out, e)
	}
	return append(out, jfE3Weitere()...)
}

// typDesElements prüft, dass nur das Element an der Stelle den Typ trägt und die anderen den Typ des Sets.
func typDesElements(stelle int, typ string) func(*testing.T, jergebnis) {
	return func(t *testing.T, r jergebnis) {
		for i, e := range r.seite.Entities {
			soll := "Zoo.Animal"
			if i == stelle {
				soll = typ
			}
			if e.Typ != soll {
				t.Errorf("Element %d hat den Typ %q, erwartet %q", i, e.Typ, soll)
			}
		}
	}
}

func jfE3Weitere() []jfall {
	dog := func(members string) []byte { return jEin("Animals", `"@odata.type":"#Zoo.Dog",`+members) }
	return []jfall{
		{name: "E3 Breed ohne Angabe", m: "Z", set: "Animals", einzel: true, body: jEin("Animals", `"Id":1,"Breed":"x"`), ursache: model.DatenUnbekannteProperty, ort: "/Breed"},
		{name: "E3 Breed mit Dog", m: "Z", set: "Animals", einzel: true, body: dog(`"Id":1,"Breed":"x"`), typ: "Zoo.Dog",
			props: []model.PropertyWert{pw("Id", model.ArtInt64, "1"), pw("Breed", model.ArtString, "x")}},
		{name: "E3 Fur von Mammal mit Dog", m: "Z", set: "Animals", einzel: true, body: dog(`"Id":1,"Fur":true`), typ: "Zoo.Dog",
			props: []model.PropertyWert{pw("Id", model.ArtInt64, "1"), pw("Fur", model.ArtBoolean, "true")}},
		{name: "E3 Angabe nach den Properties", m: "Z", set: "Animals", einzel: true, body: jEin("Animals", `"Id":1,"Breed":"x","@odata.type":"#Zoo.Dog"`), typ: "Zoo.Dog",
			props: []model.PropertyWert{pw("Id", model.ArtInt64, "1"), pw("Breed", model.ArtString, "x")}},
		{name: "E3 Typ des Sets", m: "R", set: "Products", einzel: true, body: jEin("Products", `"@odata.type":"#Demo.Product","ID":1`), typ: "Demo.Product"},
		{name: "E3 fremder Typ bei R", m: "R", set: "Products", einzel: true, body: jEin("Products", `"@odata.type":"#Demo.Category","ID":1`), ursache: model.DatenODataTyp, ort: "/@odata.type"},
		{name: "E3 Alias", m: "K", set: "Lines", einzel: true, body: jEin("Lines", `"@odata.type":"#S.Gift","OrderID":1,"LineNo":1`), ursache: model.DatenODataTyp, ort: "/@odata.type"},
		{name: "E3 qualifizierter Name", m: "K", set: "Lines", einzel: true, body: jEin("Lines", `"@odata.type":"#Shop.Gift","OrderID":1,"LineNo":1`), typ: "Shop.Gift", schluessel: "OrderID=1,LineNo=1"},
		{name: "E3 Typ des Sets bei K", m: "K", set: "Lines", einzel: true, body: jEin("Lines", `"@odata.type":"#Shop.OrderLine","OrderID":1,"LineNo":1`), typ: "Shop.OrderLine"},
	}
}

func jE3(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfE3()) }

// e4Gilt meldet, ob die Annotation an der Stelle keine Wirkung hat und dort zu prüfen ist: an der Wurzel der Collection
// tragen `@odata.context`, `@odata.count` und `@odata.nextLink` ihre Bedeutung und gibt es keine Annotation einer Property;
// an der Wurzel der Entity ebenso, dazu `@odata.type`; an einem Element hat nur `@odata.type` eine Bedeutung.
func e4Gilt(name, stelle string) bool {
	eigene := name == "@odata.context" || name == "@odata.count" || name == "@odata.nextLink"
	switch stelle {
	case "Wurzel der Collection":
		return strings.HasPrefix(name, "@") && !eigene
	case "Wurzel der Entity":
		return !eigene && name != "@odata.type"
	}
	return name != "@odata.type"
}

func e4Annotationen() []string {
	return []string{
		"@odata.id", "@odata.etag", "@odata.editLink", "@odata.readLink", "@odata.context", "@odata.nextLink", "@odata.count", "@odata.type",
		"@x.y", "@x.y#q", "ID@odata.type", "Price@odata.type", "Category@odata.navigationLink", "Foo@x.y",
	}
}

func e4Werte() []string { return []string{`"a"`, `1`, `null`, `{"a":[1]}`, `[1]`, `true`} }

func jfE4() []jfall {
	var out []jfall
	for _, name := range e4Annotationen() {
		for _, w := range e4Werte() {
			m := `"` + name + `":` + w
			for _, st := range []struct {
				stelle string
				f      jfall
				soll   []model.Entity
			}{
				{"Wurzel der Collection", jfall{body: rootMit(m, false, `{"ID":1}`, `{"ID":2}`)}, []model.Entity{produkt(1), produkt(2)}},
				{"erstes Element", jfall{body: jKol("Products", `{`+m+`,"ID":1}`, `{"ID":2}`)}, []model.Entity{produkt(1), produkt(2)}},
				{"letztes Element", jfall{body: jKol("Products", `{"ID":1}`, `{"ID":2,`+m+`}`)}, []model.Entity{produkt(1), produkt(2)}},
				{"Wurzel der Entity", jfall{einzel: true, body: jEin("Products", `"ID":1,`+m)}, []model.Entity{produkt(1)}},
			} {
				if !e4Gilt(name, st.stelle) {
					continue
				}
				f := st.f
				f.name, f.m, f.set = fmt.Sprintf("E4 %s %s an %s", name, w, st.stelle), "R", "Products"
				soll := st.soll
				f.nach = func(t *testing.T, r jergebnis) { entitiesGleich(t, f.name, r, f.einzel, soll...) }
				out = append(out, f)
			}
		}
	}
	for _, name := range []string{"@", "Name@", "a@b@c", "@@x"} {
		out = append(out, beide("E4 Verstoß "+name, "R", "Products", `"ID":1,"`+name+`":1`, model.DatenStruktur, "/"+name)...)
		out = append(out, jfall{name: "E4 Verstoß " + name + " an der Wurzel der Collection", m: "R", set: "Products", body: rootMit(`"`+name+`":1`, false),
			ursache: model.DatenStruktur, ort: "/" + name})
	}
	return out
}

func jE4(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfE4()) }

func wertformen() []string { return []string{`1.5`, `"x"`, `null`, `{"a":1}`, `[1]`, `true`} }

// jfE5 sind die Mitgliedsarten (SPEC-126 Nr. 9): nicht unterstützte Typen, Navigationen, unbekannte und dynamische
// Members, Namen mit Sonderzeichen und die Reihenfolge des Bodys.
func jfE5() []jfall {
	return verbinde(jfE5Typ(), jfE5Navigation(), jfE5Offen(), jfE5Namen())
}

func jfE5Typ() []jfall {
	var out []jfall
	id := pw("Id", model.ArtInt32, "1")
	for _, w := range wertformen() {
		out = append(out, mitUeber(beideGut("E5 Weight "+w, "R", "Products", `"ID":1,"Weight":`+w, pw("ID", model.ArtInt32, "1")),
			model.Uebersprungen{Name: "Weight", Grund: model.UebersprungTyp})...)
		for _, name := range []string{"Dbl", "Tags", "Cx", "En"} {
			out = append(out, mitUeber(beideGut("E5 "+name+" "+w, "W", "Ws", `"Id":1,"`+name+`":`+w, id),
				model.Uebersprungen{Name: name, Grund: model.UebersprungTyp})...)
		}
	}
	return out
}

func jfE5Navigation() []jfall {
	var out []jfall
	nav := model.Uebersprungen{Name: "Category", Grund: model.UebersprungNavigation}
	for _, w := range []string{`{"ID":"x"}`, `[{"a":1},{"a":2}]`, `null`, `{}`, `[]`} {
		out = append(out, mitUeber(beideGut("E5 Category "+w, "R", "Products", `"ID":1,"Category":`+w, pw("ID", model.ArtInt32, "1")), nav)...)
	}
	for _, w := range []string{`"x"`, `5`, `true`, `false`} {
		out = append(out, beide("E5 Category "+w, "R", "Products", `"ID":1,"Category":`+w, model.DatenTyp, "/Category")...)
	}
	for _, name := range []string{"Nav", "Navs"} {
		out = append(out, mitUeber(beideGut("E5 "+name, "W", "Ws", `"Id":1,"`+name+`":[{"Id":2}]`, pw("Id", model.ArtInt32, "1")),
			model.Uebersprungen{Name: name, Grund: model.UebersprungNavigation})...)
	}
	return out
}

func jfE5Offen() []jfall {
	var out []jfall
	for _, typ := range []string{"Zoo.Mammal", "Zoo.Dog", "Zoo.Puppy"} {
		for _, w := range wertformen() {
			members := `"@odata.type":"#` + typ + `","Id":1,"dyn":` + w
			out = append(out, mitUeber(beideGut("E5 dyn bei "+typ+" "+w, "Z", "Animals", members, pw("Id", model.ArtInt64, "1")),
				model.Uebersprungen{Name: "dyn", Grund: model.UebersprungOffen})...)
		}
		out = append(out, mitUeber(beideGut("E5 Keeper bei "+typ, "Z", "Animals", `"@odata.type":"#`+typ+`","Id":1,"Keeper":{"Id":2}`, pw("Id", model.ArtInt64, "1")),
			model.Uebersprungen{Name: "Keeper", Grund: model.UebersprungNavigation})...)
	}
	for _, w := range wertformen() {
		out = append(out, beide("E5 dyn beim geschlossenen Typ "+w, "Z", "Animals", `"Id":1,"dyn":`+w, model.DatenUnbekannteProperty, "/dyn")...)
		out = append(out, beide("E5 Nope "+w, "R", "Products", `"ID":1,"Nope":`+w, model.DatenUnbekannteProperty, "/Nope")...)
	}
	return out
}

func jfE5Namen() []jfall {
	var out []jfall
	for _, n := range []struct{ json, ort string }{
		{``, "/"}, {`a/b`, "/a~1b"}, {`a~b`, "/a~0b"}, {`~1`, "/~01"}, {`a~/b`, "/a~0~1b"}, {`id`, "/id"}, {`é`, "/é"}, {`é`, "/é"},
		{`~`, "/~0"}, {`/`, "/~1"}, {esc("007e"), "/~0"}, {esc("002f"), "/~1"}, {"e" + esc("0301"), "/e" + string(rune(0x301))},
	} {
		ort := n.ort
		out = append(out, beide("E5 Name "+n.json, "R", "Products", `"ID":1,"`+n.json+`":1`, model.DatenUnbekannteProperty, ort)...)
	}
	// Die Reihenfolge des Bodys ist die Reihenfolge von Properties und Uebersprungen, nicht die des Modells.
	out = append(out, mitUeber(beideGut("E5 Reihenfolge", "R", "Products", `"Weight":2,"Price":1,"Name":"x","Category":null,"ID":1`,
		pw("Price", model.ArtDecimal, "1"), pw("Name", model.ArtString, "x"), pw("ID", model.ArtInt32, "1")),
		model.Uebersprungen{Name: "Weight", Grund: model.UebersprungTyp}, model.Uebersprungen{Name: "Category", Grund: model.UebersprungNavigation})...)
	return out
}

func jE5(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfE5()) }
