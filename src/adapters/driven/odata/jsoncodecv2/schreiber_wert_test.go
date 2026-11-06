package jsoncodecv2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// dt ist die Ausgabe eines Zeitpunkts mit den Millisekunden ms: der String mit den Schrägstrichen als Escape (SPEC-172 Nr. 6 f).
func dt(ms string) string { return `"` + bs + "/Date(" + ms + ")" + bs + `/"` }

// artFall ist ein Text einer Art; want ist die Ausgabe des Wertes (nur bei gültigem Text).
type artFall struct {
	text, want string
}

// artMenge sind die Elemente der Klasse „Text je Art“ (zweite Tabelle des Plans): gültig und ungültig.
type artMenge struct {
	art       model.WertArt
	prop      string // Name der nicht nullbaren Property in Ns.Alle
	gueltig   []artFall
	ungueltig []string
}

func zahlen(texte ...string) []artFall {
	var out []artFall
	for _, t := range texte {
		out = append(out, artFall{t, t})
	}
	return out
}

func zeichenketten(texte ...string) []artFall {
	var out []artFall
	for _, t := range texte {
		out = append(out, artFall{t, `"` + t + `"`})
	}
	return out
}

func artMengen() []artMenge {
	z := "2020-01-01T00:00:00"
	return []artMenge{
		{model.ArtString, "S", []artFall{{"", `""`}, {"a", `"a"`}, {"\xc3\xa9", "\"\xc3\xa9\""}, {"e\xcc\x81", "\"e\xcc\x81\""},
			{"\x00", `"` + esc("0000") + `"`}, {"\xf0\x90\x80\x80", "\"\xf0\x90\x80\x80\""}},
			[]string{"\x80", "\xc0\x80", "\xed\xa0\x80", "a\xc3"}},
		{model.ArtBoolean, "B", zahlen("true", "false"),
			[]string{"True", "TRUE", "1", "0", "", " true", "true\n", "yes"}},
		{model.ArtInt32, "I32", zahlen("0", "-1", "2147483647", "-2147483648"),
			[]string{"2147483648", "-2147483649", "12345678901", "-0", "+1", "01", "00", "-01", "1.0", "1e2", "", " 1", "1 ", "1\n", "\xd9\xa1", "0x1", "--1"}},
		{model.ArtInt64, "I64", zeichenketten("0", "-1", "9223372036854775807", "-9223372036854775808"),
			[]string{"9223372036854775808", "-9223372036854775809", "99999999999999999999", "-0", "+1", "01", "1.0", "1e2", "", " 1", "1 ", "1\n", "\xd9\xa1", "-"}},
		{model.ArtDecimal, "Dec", zeichenketten("0", "-1", "0.5", "-0.5", "10", "123456789012345678901234567890.123456789", strings.Repeat("1", 1024)),
			[]string{"-0", "0.0", "-0.0", "1.50", "1.", ".5", "00.5", "01", "+1", "1e2", "1E2", "", " 1", "1 ", "1\n", "1,5", "NaN", "INF", "-", "1.5.5",
				strings.Repeat("1", 1025), "\xd9\xa1"}},
		{model.ArtGUID, "G", zeichenketten("0f8fad5b-d9cb-469f-a165-70867728950e", "00000000-0000-0000-0000-000000000000"),
			[]string{"0F8FAD5B-D9CB-469F-A165-70867728950E", "0f8fad5b-d9cb-469f-a165-70867728950E", "0f8fad5bd9cb469fa16570867728950e",
				"{0f8fad5b-d9cb-469f-a165-70867728950e}", "0f8fad5b-d9cb-469f-a165-70867728950", "0f8fad5b-d9cb-469f-a165-70867728950ee",
				"0f8fad5-bd9cb-469f-a165-70867728950e", "0f8fad5b-d9cb-469f-a165-7086772895ge", " 0f8fad5b-d9cb-469f-a165-70867728950e", "0f8fad5b-d9cb-469f-a165-70867728950e\n",
				// die Nachbarn der Hex-Bereiche: das Zeichen vor 0, hinter 9 und vor a
				"0f8fad5b-d9cb-469f-a165-70867728950/", "0f8fad5b-d9cb-469f-a165-70867728950:", "0f8fad5b-d9cb-469f-a165-70867728950`"}},
		{model.ArtDateTimeOffset, "T", []artFall{
			{"0001-01-01T00:00:00Z", dt("-62135596800000")}, {"9999-12-31T23:59:59.999Z", dt("253402300799999")},
			{z + ".5Z", dt("1577836800500")}, {z + ".123Z", dt("1577836800123")},
			{"1969-12-31T23:59:59.999Z", dt("-1")}, {"1970-01-01T00:00:00Z", dt("0")}},
			[]string{z + "+01:00", z + "z", "2020-01-01t00:00:00Z", "2020-01-01 00:00:00Z", "2020-01-01T00:00Z",
				z + ".0Z", z + ".50Z", z + ".1234Z", z + ".123456789012Z", z + ".1234567890123Z", "2020-01-01T00:00:60Z",
				"2020-01-01T24:00:00Z", "2020-01-01T00:60:00Z", "2020-13-01T00:00:00Z", "2023-02-29T00:00:00Z", "0000-01-01T00:00:00Z",
				"10000-01-01T00:00:00Z", z, z + ".Z", z + ",5Z", " " + z + "Z", z + "Z\n", z + ".12 Z", z + ".1aZ"}},
	}
}

func alleAntwort(props ...model.PropertyWert) model.EntityAntwort {
	return model.EntityAntwort{Wurzel: wurzel, Set: "Alles", Entity: model.Entity{Typ: "Ns.Alle", Schluessel: "k", Properties: props}}
}

// fueller sind gültige nullbare Properties, die vor und hinter dem geprüften Wert stehen.
func fueller() (model.PropertyWert, model.PropertyWert) {
	return pw("NS", model.ArtString, "f"), pw("NB", model.ArtBoolean, "true")
}

// TestTextKanonisch hält Zeile 11: Schleife Art × Element × Stelle (4 Stellen: erste und letzte Property der Entity, erstes und letztes Element).
func TestTextKanonisch(t *testing.T) {
	gueltig, ungueltig, zellen := 0, 0, 0
	for _, m := range artMengen() {
		gueltig += len(m.gueltig)
		ungueltig += len(m.ungueltig)
		pruefe := func(text string, want *string) {
			ziel := pw(m.prop, m.art, text)
			f1, f2 := fueller()
			e1 := alleAntwort(ziel, f1, f2)
			e2 := alleAntwort(f1, f2, ziel)
			c1, c2 := model.CollectionAntwort{Wurzel: wurzel, Set: "Alles"}, model.CollectionAntwort{Wurzel: wurzel, Set: "Alles"}
			gut := model.Entity{Typ: "Ns.Alle", Schluessel: "k", Properties: []model.PropertyWert{f1}}
			c1.Entities = []model.Entity{e1.Entity, gut, gut}
			c2.Entities = []model.Entity{gut, gut, e2.Entity}
			name := fmt.Sprintf("%s %q", m.art, text)
			zellen += 4
			if want == nil {
				entityScheitert(t, name+" erste Property", e1)
				entityScheitert(t, name+" letzte Property", e2)
				collectionScheitert(t, name+" erstes Element", c1)
				collectionScheitert(t, name+" letztes Element", c2)
				return
			}
			frag := `"` + m.prop + `":` + *want
			for was, body := range map[string]string{
				"erste Property": entityBody(t, name, e1), "letzte Property": entityBody(t, name, e2),
				"erstes Element": collectionBody(t, name, c1), "letztes Element": collectionBody(t, name, c2),
			} {
				if !strings.Contains(body, frag) {
					t.Errorf("%s %s: %s fehlt in %s", name, was, frag, body)
				}
			}
		}
		for _, g := range m.gueltig {
			pruefe(g.text, &g.want)
		}
		for _, u := range m.ungueltig {
			pruefe(u, nil)
		}
	}
	if gueltig != 31 || ungueltig != 102 || zellen != 532 {
		t.Errorf("Menge gültig %d, ungültig %d, Zellen %d; will 31, 102, 532", gueltig, ungueltig, zellen)
	}
}

// TestArtGegenTyp hält Zeile 10: Wert.Art gehört zum EdmType; null nur bei nullbarer Property und mit leerem Text; `redigiert` und `date` nie.
func TestArtGegenTyp(t *testing.T) {
	typen, namen := siebenTypen()
	proben := map[model.WertArt]string{
		model.ArtString: "a", model.ArtBoolean: "true", model.ArtInt32: "1", model.ArtInt64: "1", model.ArtDecimal: "1.5",
		model.ArtGUID: "0f8fad5b-d9cb-469f-a165-70867728950e", model.ArtDate: "2024-01-05", model.ArtDateTimeOffset: "2020-01-01T00:00:00Z",
	}
	arten := []model.WertArt{model.ArtString, model.ArtBoolean, model.ArtInt32, model.ArtInt64, model.ArtDecimal, model.ArtGUID, model.ArtDate, model.ArtDateTimeOffset}
	eigene := []model.WertArt{model.ArtString, model.ArtBoolean, model.ArtInt32, model.ArtInt64, model.ArtDecimal, model.ArtGUID, model.ArtDateTimeOffset}
	gueltig, ungueltig := 0, 0
	for i := range typen {
		for _, art := range arten {
			a := alleAntwort(pw(namen[i], art, proben[art]))
			if art == eigene[i] {
				gueltig++
				entityBody(t, typen[i]+" "+string(art), a)
			} else {
				ungueltig++
				entityScheitert(t, typen[i]+" mit Art "+string(art), a)
			}
		}
		ungueltig++
		entityScheitert(t, "null bei nicht nullbarer "+namen[i], alleAntwort(pw(namen[i], model.ArtNull, "")))
		gueltig++
		if body := entityBody(t, "null bei nullbarer", alleAntwort(pw("N"+namen[i], model.ArtNull, ""))); !strings.HasSuffix(body, `"N`+namen[i]+`":null}}`) {
			t.Errorf("null bei nullbarer %s: %s", namen[i], body)
		}
		ungueltig++
		entityScheitert(t, "null mit Text", alleAntwort(pw("N"+namen[i], model.ArtNull, "x")))
		ungueltig++
		entityScheitert(t, "redigiert", alleAntwort(pw("N"+namen[i], model.ArtRedigiert, "")))
		for _, art := range []model.WertArt{"", "unbekannt"} {
			ungueltig++
			entityScheitert(t, "Art "+string(art), alleAntwort(pw("N"+namen[i], art, proben[eigene[i]])))
		}
	}
	for _, art := range arten {
		ungueltig++
		entityScheitert(t, "Edm.Date mit Art "+string(art), alleAntwort(pw("D", art, proben[art])))
	}
	ungueltig++
	entityScheitert(t, "null bei Edm.Date", alleAntwort(pw("ND", model.ArtNull, "")))
	if gueltig != 14 || ungueltig != 93 {
		t.Errorf("Menge gültig %d, ungültig %d; will 14, 93", gueltig, ungueltig)
	}
}

// modellStellen hat an jeder der zwei Stellen den Zusatz des Namens: Property `P<z>x` und Typ `Ns.D<z>`.
func modellStellen(zProp, zTyp string) model.ServiceModel {
	props := []model.Property{prop("Id", "Edm.Int32", false), prop("P"+zProp+"x", "Edm.String", true)}
	typen := []model.EntityType{
		{Name: "Ns.B", Keys: []string{"Id"}, Properties: props},
		{Name: "Ns.D" + zTyp, BaseType: "Ns.B", Keys: []string{"Id"}, Properties: props},
	}
	return model.NeuesServiceModel("2.0", []string{"Ns"}, typen, []model.EntitySet{{Name: "Set", EntityType: "Ns.B"}}, nil)
}

// stringStelle ist eine der fünf Stellen, an denen ein Text in die Ausgabe geht (Zeile 30): ein Wert am Anfang, in der Mitte, am Ende,
// der Name einer Property und der Typ; bau liefert die Antwort zum Text und das Stück der Ausgabe, in dem der Text erscheint.
type stringStelle struct {
	name string
	bau  func(text, esc string) (m model.ServiceModel, a model.EntityAntwort, frag string)
}

func stringStellen() []stringStelle {
	wert := func(vor, nach string) func(text, esc string) (model.ServiceModel, model.EntityAntwort, string) {
		return func(text, esc string) (model.ServiceModel, model.EntityAntwort, string) {
			a := gutEntity()
			a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, vor+text+nach)}
			return testModell(), a, `"Name":"` + vor + esc + nach + `"}}`
		}
	}
	return []stringStelle{
		{"Wert am Anfang", wert("", "b")}, {"Wert in der Mitte", wert("a", "b")}, {"Wert am Ende", wert("a", "")},
		{"Name einer Property", func(text, esc string) (model.ServiceModel, model.EntityAntwort, string) {
			a := model.EntityAntwort{Wurzel: wurzel, Set: "Set", Entity: model.Entity{Typ: "Ns.B", Schluessel: "1",
				Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("P"+text+"x", model.ArtString, "v")}}}
			return modellStellen(text, ""), a, `,"P` + esc + `x":"v"}}`
		}},
		{"Typ in type", func(text, esc string) (model.ServiceModel, model.EntityAntwort, string) {
			a := model.EntityAntwort{Wurzel: wurzel, Set: "Set", Entity: model.Entity{Typ: "Ns.D" + text, Schluessel: "1",
				Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1")}}}
			return modellStellen("", text), a, `"type":"Ns.D` + esc + `"},"Id":1}}`
		}},
	}
}

// stringTexte sind die 269 Texte je Stelle (Zeile 30): jedes der 256 Bytes allein, acht Mehrbyte-Zeichen, fünf ungültige Folgen.
// Das Orakel steht im Test: erwartet liefert die Ausgabe des Textes oder false bei ungültigem UTF-8.
func stringTexte() []string {
	var out []string
	for c := range 256 {
		out = append(out, string([]byte{byte(c)}))
	}
	for _, r := range []rune{0xE9, 0x800, 0x10000, 0x2028, 0x2029, 0xFEFF, 0xFFFF, 0x10FFFF} {
		out = append(out, string(r))
	}
	return append(out, "\xc0\x80", "\xed\xa0\x80", "a\xc3", "\xf4\x90\x80\x80", "a\xa9")
}

// stringErwartet ist das Orakel des Escapes: ok ist falsch, wenn der Text kein gültiges UTF-8 ist (Tabelle des Tests: Byte unter 0x20 mit
// Hex-Escape, `"` und Backslash mit Backslash, ab 0x80 ein einzelnes Byte ungültig).
func stringErwartet(s string) (esc2 string, ok bool) {
	if len(s) == 1 {
		switch c := s[0]; {
		case c >= 0x80:
			return "", false
		case c == '"':
			return bs + `"`, true
		case c == '\\':
			return bs + bs, true
		case c < 0x20:
			return esc(fmt.Sprintf("%04x", c)), true
		}
		return s, true
	}
	for _, ungueltig := range []string{"\xc0\x80", "\xed\xa0\x80", "a\xc3", "\xf4\x90\x80\x80", "a\xa9"} {
		if s == ungueltig {
			return "", false
		}
	}
	return s, true
}

// TestStringKlasse hält Zeile 30: 269 Texte × 5 Stellen, jede Ausgabe gegen die Tabelle des Orakels.
func TestStringKlasse(t *testing.T) {
	texte := stringTexte()
	if len(texte) != 269 {
		t.Fatalf("Texte %d, will 269", len(texte))
	}
	n := 0
	for _, st := range stringStellen() {
		for _, text := range texte {
			esc2, ok := stringErwartet(text)
			m, a, frag := st.bau(text, esc2)
			z := neuesZiel()
			_, err := schreiber().SchreibeEntity(context.Background(), z, m, a)
			n++
			if !ok {
				istUngueltig(t, err, fmt.Sprintf("%s %q", st.name, text))
				z.leer(t, st.name)
				continue
			}
			if err != nil || !strings.HasSuffix(z.body(), frag) {
				t.Errorf("%s %q: %v; %s fehlt am Ende von %s", st.name, text, err, frag, z.body())
			}
		}
	}
	if n != 1345 {
		t.Errorf("Zellen %d, will 1345", n)
	}
}

// TestMarkerText hält Zeile 31: der Text `[REDACTED]` eines Edm.String wird mit dem Escape für `[` geschrieben, jeder andere Text unverändert.
func TestMarkerText(t *testing.T) {
	faelle := []struct{ text, want string }{
		{"[REDACTED]", `"` + esc("005b") + `REDACTED]"`}, {"[REDACTED] ", `"[REDACTED] "`}, {" [REDACTED]", `" [REDACTED]"`},
		{"[redacted]", `"[redacted]"`}, {"[REDACTED", `"[REDACTED"`}, {"REDACTED]", `"REDACTED]"`}, {"[REDACTED]x", `"[REDACTED]x"`},
		{"[[REDACTED]", `"[[REDACTED]"`},
	}
	n := 0
	for _, f := range faelle {
		e := gutEntity()
		e.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, f.text)}
		c := gutCollection()
		c.Entities = []model.Entity{e.Entity, produkt("2", "B"), e.Entity}
		frag := `"Name":` + f.want
		if body := entityBody(t, f.text, e); !strings.HasSuffix(body, frag+"}}") {
			t.Errorf("Entity %q: %s", f.text, body)
		}
		body := collectionBody(t, f.text, c)
		if !strings.HasPrefix(body, `{"d":{"results":[`) || strings.Count(body, frag+"}") != 2 {
			t.Errorf("Collection %q: %s", f.text, body)
		}
		n += 3
	}
	if n != 24 {
		t.Errorf("Zellen %d, will 24", n)
	}
}
