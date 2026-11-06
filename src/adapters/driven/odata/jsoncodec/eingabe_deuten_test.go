package jsoncodec_test

import (
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const eTypOrt = "/@odata.type"

// TestEingabeODataTyp hält Zeile 17: `@odata.type` der Wurzel, vor allen Mitgliedern gedeutet. Menge: gültig 4 × 2 Stellen (PATCH) + ungültig
// 11 × 3 Arten.
func TestEingabeODataTyp(t *testing.T) {
	gueltig := []struct{ typ, mitglieder string }{{"Ns.Mitte", `"Id":1`}, {"Ns.Blatt", `"Id":1`}, {"Ns.Blatt2", `"Id":1`}, {"Ns.Blatt", `"Id":1,"Extra":"x"`}}
	n := 0
	for _, g := range gueltig {
		ot := `"@odata.type":"#` + g.typ + `"`
		for _, stelle := range []string{"erstes", "letztes"} {
			body := "{" + ot + "," + g.mitglieder + "}"
			if stelle == "letztes" {
				body = "{" + g.mitglieder + "," + ot + "}"
			}
			props := []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}
			if strings.Contains(g.mitglieder, "Extra") {
				props = append(props, ePW("Extra", model.ArtString, "x"))
			}
			got, err := eA("Mitten", model.EingabePatch, body).los()
			eOK(t, g.typ+" "+stelle, got, err, model.EntityEingabe{Typ: g.typ, Schluessel: "1", Properties: props})
			n++
		}
	}
	if n != 8 {
		t.Errorf("Menge gültig %d, erwartet 8", n)
	}
	ungueltig := []string{`5`, `null`, `"Ns.Mitte"`, `""`, `"#"`, `"#Ns.Gibtsnicht"`, `"#A.Mitte"`, `"#Ns.Mitte "`, `"#Ns.Fremd"`, `"#Ns.Abs"`, `"#Ns.Wurzel"`}
	m := 0
	for _, v := range ungueltig {
		for _, art := range eArten() {
			got, err := eA("Mitten", art, `{"Id":1,"@odata.type":`+v+`}`).los()
			eFehler(t, "@odata.type "+v, got, err, model.EingabeODataTyp, eTypOrt)
			m++
		}
	}
	if m != 33 {
		t.Errorf("Menge ungültig %d, erwartet 33", m)
	}
}

// TestEingabeAbstrakt hält Zeile 18: ein abstrakter Typ legt keine Entity an. Menge 9 = 3 Fälle × 3 Arten.
func TestEingabeAbstrakt(t *testing.T) {
	n := 0
	for _, art := range eArten() {
		got, err := eA("Wurzeln", art, `{"Id":1}`).los()
		eFehler(t, "abstrakt ohne Typ", got, err, model.EingabeODataTyp, eTypOrt)
		got, err = eA("Wurzeln", art, `{"@odata.type":"#Ns.Mitte","Id":1}`).los()
		eOK(t, "abgeleitet", got, err, model.EntityEingabe{Typ: "Ns.Mitte", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}})
		got, err = eA("Wurzeln", art, `{"@odata.type":"#Ns.Wurzel","Id":1}`).los()
		eFehler(t, "abstrakt als Typ", got, err, model.EingabeODataTyp, eTypOrt)
		n += 3
	}
	if n != 9 {
		t.Errorf("Menge %d, erwartet 9", n)
	}
}

// eAnnotation ist eine Form der Zeile 19: Name, JSON-Wert und Wirkung (leer: keine, sonst die Ursache).
type eAnnotation struct {
	name, wert string
	ursache    model.EingabeUrsache
}

// eAnnotationen sind die 65 Formen der Zeile 19 (zweite Tabelle): 42 + 9 + 1 ohne Wirkung, 9 `bind`, 4 `struktur`.
func eAnnotationen() []eAnnotation {
	var out []eAnnotation
	werte := []string{`"s"`, `5`, `true`, `false`, `null`, `{}`, `[]`}
	for _, n := range []string{"@odata.context", "@odata.id", "@odata.etag", "@odata.editLink", "@odata.readLink", "@fremd.term"} {
		for _, w := range werte {
			out = append(out, eAnnotation{n, w, ""})
		}
	}
	for _, n := range []string{"Name", "Unbekannt", "Einzel"} {
		for _, term := range []string{"odata.type", "odata.navigationLink", "fremd.x"} {
			out = append(out, eAnnotation{n + "@" + term, `"x"`, ""})
		}
	}
	out = append(out, eAnnotation{"@odata.bind", `"x"`, ""})
	for _, n := range []string{"Name", "Einzel", "Unbekannt"} {
		for _, w := range []string{`"x"`, `[]`, `null`} {
			out = append(out, eAnnotation{n + "@odata.bind", w, model.EingabeBind})
		}
	}
	for _, n := range []string{"@", "Name@", "a@b@c", "@@x"} {
		out = append(out, eAnnotation{n, `1`, model.EingabeStruktur})
	}
	return out
}

// TestEingabeAnnotationen hält Zeile 19: nur `bind` und eine ungültige Grammatik wirken. Menge 585 = 65 Formen × 3 Stellen × 3 Arten.
func TestEingabeAnnotationen(t *testing.T) {
	n := 0
	for _, f := range eAnnotationen() {
		ann := eQuote(f.name) + ":" + f.wert
		stellen := []string{ann + `,"Id":1,"Name":"x"`, `"Id":1,` + ann + `,"Name":"x"`, `"Id":1,"Name":"x",` + ann}
		for s, mitglieder := range stellen {
			for _, art := range eArten() {
				got, err := eA("Ms", art, "{"+mitglieder+"}").los()
				was := f.name + " " + f.wert + " Stelle " + eP(s) + " " + string(art)
				if f.ursache == "" {
					eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.M", Schluessel: "1",
						Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("Name", model.ArtString, "x")}})
				} else {
					eFehler(t, was, got, err, f.ursache, "/"+f.name)
				}
				n++
			}
		}
	}
	if n != 585 {
		t.Errorf("Menge %d, erwartet 585", n)
	}
}

// TestEingabeAnnotationenNachbarn hält Zeile 19 für die Klasse „Namen, die einen Sonderfall nur berühren“: Präfix, Suffix, Abschnitt, Schreibweise und Leerzeichen
// von `@odata.type` (8 Namen an der Wurzel) und von `odata.bind` (8 Terme × 3 Namen) sind Annotationen ohne Wirkung, nie `bind` und nie ein Typ.
// Menge (8 + 24) × 3 Stellen × 3 Arten = 288.
func TestEingabeAnnotationenNachbarn(t *testing.T) {
	namen := []string{"@odata.typeX", "@odata.typ", "@Xodata.type", "@ODATA.TYPE", "@odata.Type", "@odata.type ", "@ odata.type", "@odata.types"}
	for _, name := range []string{"Name", "Einzel", "Unbekannt"} {
		for _, term := range []string{"odata.bindX", "odata.bin", "Xodata.bind", "x.odata.bind", "ODATA.BIND", "odata.Bind", "odata.bind ", " odata.bind"} {
			namen = append(namen, name+"@"+term)
		}
	}
	n := 0
	for _, name := range namen {
		ann := eQuote(name) + `:"x"`
		stellen := []string{ann + `,"Id":1,"Name":"x"`, `"Id":1,` + ann + `,"Name":"x"`, `"Id":1,"Name":"x",` + ann}
		for s, mitglieder := range stellen {
			for _, art := range eArten() {
				got, err := eA("Ms", art, "{"+mitglieder+"}").los()
				eOK(t, name+" Stelle "+eP(s)+" "+string(art), got, err, model.EntityEingabe{Typ: "Ns.M", Schluessel: "1",
					Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("Name", model.ArtString, "x")}})
				n++
			}
		}
	}
	if n != 32*3*3 {
		t.Errorf("Menge %d, erwartet 288", n)
	}
}

// TestEingabeAbgeleiteterTypPflicht hält Zeile 30 für den abgeleiteten Typ: ein `POST` oder `PUT` mit `@odata.type` eines abgeleiteten Typs verlangt dessen eigene nicht
// nullbare Property (SPEC-181 Nr. 7 a „wirksame Properties des Typs“), nicht die des Typs des Sets.
func TestEingabeAbgeleiteterTypPflicht(t *testing.T) {
	faelle := []struct {
		typ, body, fehlt string // fehlt: leer = gültig
	}{
		{"Ns.Pflicht", `"Id":1`, "/Muss"}, {"Ns.Pflicht", `"Id":1,"Muss":3`, ""},
		{"Ns.Pflicht2", `"Id":1`, "/Muss"}, {"Ns.Pflicht2", `"Id":1,"Muss":3`, "/Muss2"}, {"Ns.Pflicht2", `"Id":1,"Muss2":4`, "/Muss"},
		{"Ns.Pflicht2", `"Id":1,"Muss":3,"Muss2":4`, ""},
	}
	n := 0
	for _, f := range faelle {
		body := `{"@odata.type":"#` + f.typ + `",` + f.body + `}`
		for _, art := range eArten() {
			got, err := eA("Mitten", art, body).los()
			if f.fehlt != "" && art != model.EingabePatch {
				eFehler(t, f.typ+" "+f.body+" "+string(art), got, err, model.EingabeFehlend, f.fehlt)
			} else if err != nil || got.Typ != f.typ || got.Schluessel != "1" {
				t.Errorf("%s %s %s: %+v %v", f.typ, f.body, art, got, err)
			}
			n++
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// TestEingabeMitgliederAbgeleiteterTyp hält Zeile 20 für die Achse „Typ der Entity“: jede Mitglieder-Klasse (Navigation, nicht unterstützte Property, unbekannter
// Name, Schlüssel mit U+0000, gültige Property) gegen den Typ des Sets ohne `BaseType` (`Ns.VB`) und gegen die abgeleiteten Typen mit `BaseType` (`Ns.VD` geschlossen,
// `Ns.VDO` offen), der Typ steht in `@odata.type`. Menge 3 Typen × 5 Klassen × 3 Arten = 45.
func TestEingabeMitgliederAbgeleiteterTyp(t *testing.T) {
	typen := []struct{ name, kopf string }{{"Ns.VB", ""}, {"Ns.VD", `"@odata.type":"#Ns.VD",`}, {"Ns.VDO", `"@odata.type":"#Ns.VDO",`}}
	nul := esc("0000")
	n := 0
	for _, ty := range typen {
		unbekannt := model.EingabeUnbekannteProperty
		if ty.name == "Ns.VDO" {
			unbekannt = model.EingabeOffen
		}
		klassen := []struct {
			name, body string
			ursache    model.EingabeUrsache
			ort        string
		}{
			{"Navigation", `"Id":"k","Nav":{}`, model.EingabeNavigation, "/Nav"},
			{"nicht unterstützt", `"Id":"k","Dbl":1`, model.EingabePropertyTyp, "/Dbl"},
			{"unbekannt", `"Id":"k","zz":1`, unbekannt, "/zz"},
			{"Schlüssel mit U+0000", `"Id":"a` + nul + `b"`, model.EingabeSchluessel, "/Id"},
			{"gültig", `"Id":"k","N1":5`, "", ""},
		}
		for _, k := range klassen {
			for _, art := range eArten() {
				got, err := eA("VB", art, "{"+ty.kopf+k.body+"}").los()
				was := ty.name + " " + k.name + " " + string(art)
				if k.ursache != "" {
					eFehler(t, was, got, err, k.ursache, k.ort)
				} else {
					eOK(t, was, got, err, model.EntityEingabe{Typ: ty.name, Schluessel: "'k'", Properties: []model.PropertyWert{ePW("Id", model.ArtString, "k"), ePW("N1", model.ArtInt32, "5")}})
				}
				n++
			}
		}
	}
	if n != 45 {
		t.Errorf("Menge %d, erwartet 45", n)
	}
}

// TestEingabeMitglieder hält Zeile 20: nicht unterstützter Typ, Navigation, dynamisches Member und unbekannter Name, jeder JSON-Wert.
// Menge 315 = 15 Klassen (10 im geschlossenen Typ `Ms` und `Os`, 5 Namen des Modells im offenen Typ) × 7 Werte × 3 Arten.
func TestEingabeMitglieder(t *testing.T) {
	klassen := []struct {
		set, name string
		ursache   model.EingabeUrsache
	}{
		{"Ms", "Dbl", model.EingabePropertyTyp}, {"Ms", "Bin", model.EingabePropertyTyp}, {"Ms", "Col", model.EingabePropertyTyp},
		{"Ms", "Kom", model.EingabePropertyTyp}, {"Ms", "Auf", model.EingabePropertyTyp},
		{"Ms", "Einzel", model.EingabeNavigation}, {"Ms", "Liste", model.EingabeNavigation},
		{"Os", "dyn", model.EingabeOffen}, {"Ms", "dyn", model.EingabeUnbekannteProperty}, {"Ms", "", model.EingabeUnbekannteProperty},
		// im offenen Typ gelten die Namen des Modells wie im geschlossenen
		{"Os", "Dbl", model.EingabePropertyTyp}, {"Os", "Kom", model.EingabePropertyTyp}, {"Os", "Einzel", model.EingabeNavigation},
		{"Os", "Liste", model.EingabeNavigation}, {"Os", "", model.EingabeOffen},
	}
	n := 0
	for _, k := range klassen {
		for _, w := range []string{`"s"`, `5`, `true`, `false`, `null`, `{}`, `[]`} {
			for _, art := range eArten() {
				got, err := eA(k.set, art, eObj("Id", "1", k.name, w)).los()
				eFehler(t, k.set+" "+k.name+" "+w, got, err, k.ursache, "/"+k.name)
				n++
			}
		}
	}
	if n != 15*7*3 {
		t.Errorf("Menge %d, erwartet 315", n)
	}
}

// TestEingabeNamen hält Zeile 21: Zeichen für Zeichen, keine Normalisierung; der Ort trägt `~0` und `~1`. Menge 30 = 10 Namen × 3 Arten.
func TestEingabeNamen(t *testing.T) {
	namen := []struct{ name, ort string }{
		{"id", "/id"}, {"ID", "/ID"}, {"Id ", "/Id "}, {"\xc3\x89a", "/\xc3\x89a"}, {"E\xcc\x81a", "/E\xcc\x81a"}, {"a/b", "/a~1b"},
		{"a~b", "/a~0b"}, {"a b", "/a b"}, {`a"b`, `/a"b`}, {"a\x01b", "/a\x01b"},
	}
	n := 0
	for _, x := range namen {
		for _, art := range eArten() {
			got, err := eA("Ws", art, eObj("Id", "1", x.name, "1")).los()
			eFehler(t, "Name "+eQuote(x.name), got, err, model.EingabeUnbekannteProperty, x.ort)
			n++
		}
	}
	if n != 30 {
		t.Errorf("Menge %d, erwartet 30", n)
	}
}

// TestEingabeReihenfolge hält Zeile 22: `Properties` steht in der Reihenfolge des Dokuments. Menge 18 = 6 Permutationen × 3 Arten, dazu 20 Läufe.
func TestEingabeReihenfolge(t *testing.T) {
	teile := map[string][2]string{"Id": {"1", "int32"}, "S": {`"x"`, "string"}, "B": {"true", "boolean"}}
	arten := map[string]model.WertArt{"Id": model.ArtInt32, "S": model.ArtString, "B": model.ArtBoolean}
	texte := map[string]string{"Id": "1", "S": "x", "B": "true"}
	perms := [][]string{{"Id", "S", "B"}, {"Id", "B", "S"}, {"S", "Id", "B"}, {"S", "B", "Id"}, {"B", "Id", "S"}, {"B", "S", "Id"}}
	n := 0
	for _, p := range perms {
		var paare []string
		var props []model.PropertyWert
		for _, name := range p {
			paare = append(paare, name, teile[name][0])
			props = append(props, ePW(name, arten[name], texte[name]))
		}
		for _, art := range eArten() {
			got, err := eA("Ws", art, eObj(paare...)).los()
			eOK(t, strings.Join(p, ","), got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: props})
			n++
		}
		for range 20 {
			got, err := eA("Ws", model.EingabePatch, eObj(paare...)).los()
			eOK(t, "Wiederholung", got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: props})
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}

// eZustandWerte sind je Typ von Z ein gültiger Wert (JSON, Art, Text) und die Ursache des leeren Textes `""` (leer: gültig).
func eZustandWerte() []struct {
	json, text string
	art        model.WertArt
	leer       model.EingabeUrsache
} {
	return []struct {
		json, text string
		art        model.WertArt
		leer       model.EingabeUrsache
	}{
		{`"x"`, "x", model.ArtString, ""}, {`true`, "true", model.ArtBoolean, model.EingabeTyp}, {`7`, "7", model.ArtInt32, model.EingabeTyp},
		{`8`, "8", model.ArtInt64, model.EingabeWert}, {`9.5`, "9.5", model.ArtDecimal, model.EingabeWert},
		{`"0f8fad5b-d9cb-469f-a165-70867728950e"`, "0f8fad5b-d9cb-469f-a165-70867728950e", model.ArtGUID, model.EingabeWert},
		{`"2020-01-01"`, "2020-01-01", model.ArtDate, model.EingabeWert},
		{`"2020-01-01T00:00:00Z"`, "2020-01-01T00:00:00Z", model.ArtDateTimeOffset, model.EingabeWert},
	}
}

// TestEingabeZustaende hält Zeile 23: genannt mit Wert, genannt mit `null`, genannt mit leerem Text, nicht genannt. Menge 192 = 8 Typen × 4
// Zustände × 2 Nullbarkeiten × 3 Arten.
func TestEingabeZustaende(t *testing.T) {
	werte := eZustandWerte()
	n := 0
	for i, w := range werte {
		for _, nullbar := range []bool{false, true} {
			x := "P" + eP(i)
			if nullbar {
				x = "Q" + eP(i)
			}
			for _, zustand := range []string{"wert", "null", "leer", "fehlt"} {
				for _, art := range eArten() {
					var paare []string
					var props []model.PropertyWert
					fuege := func(name, json string, wert model.Wert, mit bool) {
						paare = append(paare, name, json)
						if mit {
							props = append(props, model.PropertyWert{Name: name, Wert: wert})
						}
					}
					fuege("K", "1", model.Wert{Art: model.ArtInt32, Text: "1"}, true)
					for j, v := range werte {
						if name := "P" + eP(j); name != x {
							fuege(name, v.json, model.Wert{Art: v.art, Text: v.text}, true)
						}
					}
					var ursache model.EingabeUrsache
					einsetzen := func(json string, wert model.Wert) {
						// P<i> steht an seiner Stelle, Q<i> am Ende
						fuege(x, json, wert, true)
					}
					switch zustand {
					case "wert":
						einsetzen(w.json, model.Wert{Art: w.art, Text: w.text})
					case "null":
						if nullbar {
							einsetzen("null", model.Wert{Art: model.ArtNull})
						} else {
							ursache = model.EingabeNull
						}
					case "leer":
						if w.leer == "" {
							einsetzen(`""`, model.Wert{Art: w.art})
						} else {
							ursache = w.leer
						}
					case "fehlt":
						if !nullbar && art != model.EingabePatch {
							ursache = model.EingabeFehlend
						}
					}
					body := eObj(paare...)
					if ursache != "" && ursache != model.EingabeFehlend {
						// der Verstoß steht an seiner Stelle im Body
						body = eObj(append(paare, x, map[string]string{"null": "null", "leer": `""`}[zustand])...)
					}
					got, err := eA("Zs", art, body).los()
					was := x + " " + zustand + " " + string(art)
					if ursache == "" {
						eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.Z", Schluessel: "1", Properties: props})
					} else {
						eFehler(t, was, got, err, ursache, "/"+x)
					}
					n++
				}
			}
		}
	}
	if n != 192 {
		t.Errorf("Menge %d, erwartet 192", n)
	}
}

// eNamenVon liefert die Namen der erforderlichen Properties von R und ihren Wert.
func eNamenVon(paare [][2]string, ohne ...string) string {
	var kv []string
	for i := len(paare) - 1; i >= 0; i-- { // umgekehrt zur Reihenfolge des Modells
		p := paare[i]
		weg := false
		for _, o := range ohne {
			weg = weg || o == p[0]
		}
		if !weg {
			kv = append(kv, p[0], p[1])
		}
	}
	return eObj(kv...)
}

// eRErgebnis ist das Ergebnis der erforderlichen Properties von R in der Reihenfolge des Bodys (umgekehrt zum Modell).
func eRErgebnis(paare [][2]string, ohne ...string) []model.PropertyWert {
	werte := map[string]model.PropertyWert{
		"Id": ePW("Id", model.ArtInt32, "1"), "R0": ePW("R0", model.ArtString, "x"), "R1": ePW("R1", model.ArtBoolean, "true"),
		"R2": ePW("R2", model.ArtInt32, "7"), "R3": ePW("R3", model.ArtInt64, "8"), "R4": ePW("R4", model.ArtDecimal, "9.5"),
		"R5": ePW("R5", model.ArtGUID, "0f8fad5b-d9cb-469f-a165-70867728950e"), "R6": ePW("R6", model.ArtDate, "2020-01-01"),
		"R7": ePW("R7", model.ArtDateTimeOffset, "2020-01-01T00:00:00Z"),
	}
	out := []model.PropertyWert{}
	for i := len(paare) - 1; i >= 0; i-- {
		weg := false
		for _, o := range ohne {
			weg = weg || o == paare[i][0]
		}
		if !weg {
			out = append(out, werte[paare[i][0]])
		}
	}
	return out
}

// TestEingabePflicht hält Zeile 30: bei POST und PUT ist die erste nicht genannte erforderliche Property in Modellreihenfolge `fehlend`; bei
// PATCH nie. Menge: 9 × 2 + 36 × 2 + 10 + 8 × 3 + 2 × 3.
func TestEingabePflicht(t *testing.T) {
	er := eRErforderlich()
	n := 0
	for _, art := range []model.EingabeArt{model.EingabePost, model.EingabePut} {
		for _, p := range er {
			got, err := eA("Rs", art, eNamenVon(er, p[0])).los()
			eFehler(t, "fehlt "+p[0], got, err, model.EingabeFehlend, "/"+p[0])
			n++
		}
		for a := range er {
			for b := a + 1; b < len(er); b++ {
				got, err := eA("Rs", art, eNamenVon(er, er[a][0], er[b][0])).los()
				eFehler(t, "fehlt "+er[a][0]+" "+er[b][0], got, err, model.EingabeFehlend, "/"+er[a][0])
				n++
			}
		}
	}
	for _, p := range er {
		got, err := eA("Rs", model.EingabePatch, eNamenVon(er, p[0])).los()
		eOK(t, "PATCH ohne "+p[0], got, err, model.EntityEingabe{Typ: "Ns.R", Schluessel: schluesselOhne(p[0]), Properties: eRErgebnis(er, p[0])})
		n++
	}
	got, err := eA("Rs", model.EingabePatch, "{}").los()
	eOK(t, "PATCH {}", got, err, model.EntityEingabe{Typ: "Ns.R"})
	n++
	for i := range 8 {
		body := strings.TrimSuffix(eNamenVon(er), "}")
		for j := range 8 {
			if j != i {
				body += `,"N` + eP(j) + `":null`
			}
		}
		body += "}"
		for _, art := range eArten() {
			got, err := eA("Rs", art, body).los()
			props := eRErgebnis(er)
			for j := range 8 {
				if j != i {
					props = append(props, ePW("N"+eP(j), model.ArtNull, ""))
				}
			}
			eOK(t, "nullbar fehlt N"+eP(i), got, err, model.EntityEingabe{Typ: "Ns.R", Schluessel: "1", Properties: props})
			n++
		}
	}
	for range 2 {
		for _, art := range eArten() {
			got, err := eA("Rs", art, eNamenVon(er)).los()
			eOK(t, "nicht unterstützt fehlt", got, err, model.EntityEingabe{Typ: "Ns.R", Schluessel: "1", Properties: eRErgebnis(er)})
			n++
		}
	}
	if want := 18 + 72 + 10 + 24 + 6; n != want {
		t.Errorf("Menge %d, erwartet %d", n, want)
	}
}

func schluesselOhne(name string) string {
	if name == "Id" {
		return ""
	}
	return "1"
}

// eSchluesselFall ist eine Zelle der Zeile 31.
type eSchluesselFall struct {
	name, set, body, soll string
	nurPatch              bool
	props                 []model.PropertyWert
}

func eSchluesselFaelle() []eSchluesselFall {
	var out []eSchluesselFall
	for i, w := range eZustandWerte() {
		lit := w.text
		if i == 0 {
			lit = "'x'"
		}
		out = append(out, eSchluesselFall{"K" + eP(i), "K" + eP(i), eObj("K", w.json), lit, false, []model.PropertyWert{ePW("K", w.art, w.text)}})
	}
	g := "0f8fad5b-d9cb-469f-a165-70867728950e"
	out = append(out,
		eSchluesselFall{"KZ2 Id Region", "KZ2", eObj("Id", "1", "Region", `"EU"`), "Region='EU',Id=1", false,
			[]model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("Region", model.ArtString, "EU")}},
		eSchluesselFall{"KZ2 Region Id", "KZ2", eObj("Region", `"EU"`, "Id", "1"), "Region='EU',Id=1", false,
			[]model.PropertyWert{ePW("Region", model.ArtString, "EU"), ePW("Id", model.ArtInt32, "1")}},
		eSchluesselFall{"KZ3 A B C", "KZ3", eObj("A", "1", "B", `"x"`, "C", `"`+g+`"`), "A=1,B='x',C=" + g, false,
			[]model.PropertyWert{ePW("A", model.ArtInt32, "1"), ePW("B", model.ArtString, "x"), ePW("C", model.ArtGUID, g)}},
		eSchluesselFall{"KZ3 C B A", "KZ3", eObj("C", `"`+strings.ToUpper(g)+`"`, "B", `"x"`, "A", "1"), "A=1,B='x',C=" + g, false,
			[]model.PropertyWert{ePW("C", model.ArtGUID, g), ePW("B", model.ArtString, "x"), ePW("A", model.ArtInt32, "1")}},
		eSchluesselFall{"O'Hara", "K0", eObj("K", `"O'Hara"`), "'O''Hara'", false, []model.PropertyWert{ePW("K", model.ArtString, "O'Hara")}},
		eSchluesselFall{"leerer String", "K0", eObj("K", `""`), "''", false, []model.PropertyWert{ePW("K", model.ArtString, "")}},
		eSchluesselFall{"PATCH mit allen Teilen", "KZ2", eObj("Id", "1", "Region", `"EU"`), "Region='EU',Id=1", true,
			[]model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("Region", model.ArtString, "EU")}},
		eSchluesselFall{"PATCH mit einem Teil", "KZ2", eObj("Id", "1"), "", true, []model.PropertyWert{ePW("Id", model.ArtInt32, "1")}},
		eSchluesselFall{"PATCH ohne Teil", "KZ2", "{}", "", true, nil},
	)
	return out
}

// TestEingabeSchluessel hält Zeile 31: die Literalform in der Reihenfolge des Modells, wenn der Body alle Schlüssel nennt, sonst leer.
// Menge 17 Zellen: 14 × 3 Arten + 3 nur PATCH = 45 Läufe.
func TestEingabeSchluessel(t *testing.T) {
	n := 0
	for _, f := range eSchluesselFaelle() {
		arten := eArten()
		if f.nurPatch {
			arten = []model.EingabeArt{model.EingabePatch}
		}
		for _, art := range arten {
			got, err := eA(f.set, art, f.body).los()
			eOK(t, f.name+" "+string(art), got, err, model.EntityEingabe{Typ: "Ns." + f.set, Schluessel: f.soll, Properties: f.props})
			n++
		}
	}
	if n != 14*3+3 {
		t.Errorf("Läufe %d, erwartet 45", n)
	}
	for _, art := range eArten() {
		got, err := eA("OhneKey", art, `{"X":"a"}`).los()
		eOK(t, "ohne Schlüssel", got, err, model.EntityEingabe{Typ: "Ns.OhneKey", Properties: []model.PropertyWert{ePW("X", model.ArtString, "a")}})
	}
}

// TestEingabeSchluesselNul hält Zeile 32: U+0000 in einer Schlüssel-Property ist `schluessel`, in einer anderen ein gültiges Zeichen.
// Menge 111 = 4 Formen × 2 Stellen × 3 Arten (24), `null` des Schlüssels `Sts.Id` × 3, U+0000 an jeder der 25 Stellen eines Schlüssels × 3 Arten (75) und `null`
// an jeder nullbaren Schlüssel-Property (`Rs.Id`, `KN.A`, `KN.B`; das Modell führt sie als nullbar, der Schlüssel ist trotzdem `null`) × 3 Arten (9).
func TestEingabeSchluesselNul(t *testing.T) {
	nul := esc("0000")
	formen := []struct{ json, text string }{{`"` + nul + `ab"`, "\x00ab"}, {`"a` + nul + `b"`, "a\x00b"}, {`"ab` + nul + `"`, "ab\x00"}, {`"` + nul + `"`, "\x00"}}
	n := 0
	for _, f := range formen {
		for _, art := range eArten() {
			got, err := eA("Sts", art, eObj("Id", f.json)).los()
			eFehler(t, "Schlüssel "+f.json, got, err, model.EingabeSchluessel, "/Id")
			got, err = eA("Ws", art, eObj("Id", "1", "S", f.json)).los()
			eOK(t, "andere Property "+f.json, got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1",
				Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW("S", model.ArtString, f.text)}})
			n += 2
		}
	}
	for _, art := range eArten() {
		got, err := eA("Sts", art, eObj("Id", "null")).los()
		eFehler(t, "Schlüssel null", got, err, model.EingabeNull, "/Id")
		n++
	}
	// U+0000 an jeder Stelle eines Schlüssels von 25 Zeichen
	lang := strings.Repeat("a", 25)
	for i := range 25 {
		text := lang[:i] + "\x00" + lang[i+1:]
		for _, art := range eArten() {
			got, err := eA("Sts", art, eObj("Id", eQuote(text))).los()
			eFehler(t, "NUL an Stelle "+eP(i), got, err, model.EingabeSchluessel, "/Id")
			n++
		}
	}
	// `null` an einer nullbaren Schlüssel-Property
	for _, f := range []struct{ set, name string }{{"Rs", "Id"}, {"KN", "A"}, {"KN", "B"}} {
		for _, art := range eArten() {
			got, err := eA(f.set, art, eObj(f.name, "null")).los()
			eFehler(t, f.set+"."+f.name+" null", got, err, model.EingabeNull, "/"+f.name)
			n++
		}
	}
	if n != 111 {
		t.Errorf("Menge %d, erwartet 111", n)
	}
}
