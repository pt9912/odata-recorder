package jsoncodec_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// artFall ist ein Text einer Art; want ist die Ausgabe des Wertes in der Zahlform `Zahl` (nur bei gültigem Text).
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
	stunde := "T00:00:00Z"
	return []artMenge{
		{model.ArtString, "S", []artFall{{"", `""`}, {"a", `"a"`}, {"\xc3\xa9", "\"\xc3\xa9\""}, {"e\xcc\x81", "\"e\xcc\x81\""},
			{"\x00", `"` + esc("0000") + `"`}, {"\xf0\x90\x80\x80", "\"\xf0\x90\x80\x80\""}},
			[]string{"\x80", "\xc0\x80", "\xed\xa0\x80", "a\xc3"}},
		{model.ArtBoolean, "B", zahlen("true", "false"),
			[]string{"True", "TRUE", "1", "0", "", " true", "true\n", "yes"}},
		{model.ArtInt32, "I32", zahlen("0", "-1", "2147483647", "-2147483648"),
			[]string{"2147483648", "-2147483649", "12345678901", "-0", "+1", "01", "00", "-01", "1.0", "1e2", "", " 1", "1 ", "1\n", "\xd9\xa1", "0x1", "--1"}},
		{model.ArtInt64, "I64", zahlen("0", "-1", "9223372036854775807", "-9223372036854775808"),
			[]string{"9223372036854775808", "-9223372036854775809", "99999999999999999999", "-0", "+1", "01", "1.0", "1e2", "", " 1", "1 ", "1\n", "\xd9\xa1", "-"}},
		{model.ArtDecimal, "Dec", zahlen("0", "-1", "0.5", "-0.5", "10", "123456789012345678901234567890.123456789", strings.Repeat("1", 1024)),
			[]string{"-0", "0.0", "-0.0", "1.50", "1.", ".5", "00.5", "01", "+1", "1e2", "1E2", "", " 1", "1 ", "1\n", "1,5", "NaN", "INF", "-", "1.5.5",
				strings.Repeat("1", 1025), "\xd9\xa1"}},
		{model.ArtGUID, "G", zeichenketten("0f8fad5b-d9cb-469f-a165-70867728950e", "00000000-0000-0000-0000-000000000000"),
			[]string{"0F8FAD5B-D9CB-469F-A165-70867728950E", "0f8fad5b-d9cb-469f-a165-70867728950E", "0f8fad5bd9cb469fa16570867728950e",
				"{0f8fad5b-d9cb-469f-a165-70867728950e}", "0f8fad5b-d9cb-469f-a165-70867728950", "0f8fad5b-d9cb-469f-a165-70867728950ee",
				"0f8fad5-bd9cb-469f-a165-70867728950e", "0f8fad5b-d9cb-469f-a165-7086772895ge", " 0f8fad5b-d9cb-469f-a165-70867728950e", "0f8fad5b-d9cb-469f-a165-70867728950e\n"}},
		{model.ArtDate, "D", zeichenketten("0001-01-01", "9999-12-31", "2024-02-29", "2000-02-29"),
			[]string{"0000-01-01", "10000-01-01", "2023-02-29", "1900-02-29", "2024-13-01", "2024-00-01", "2024-01-00", "2024-04-31",
				"2024-1-5", "20240105", "+2024-01-01", "2024-01-01" + stunde, " 2024-01-01", "2024-01-01\n"}},
		{model.ArtDateTimeOffset, "T", zeichenketten("0001-01-01"+stunde, "9999-12-31T23:59:59Z", "2020-01-01T00:00:00.5Z", "2020-01-01T00:00:00.123456789012Z"),
			[]string{"2020-01-01T00:00:00+01:00", "2020-01-01T00:00:00z", "2020-01-01t00:00:00Z", "2020-01-01 00:00:00Z", "2020-01-01T00:00Z",
				"2020-01-01T00:00:00.0Z", "2020-01-01T00:00:00.50Z", "2020-01-01T00:00:00.1234567890123Z", "2020-01-01T00:00:60Z",
				"2020-01-01T24:00:00Z", "2020-01-01T00:60:00Z", "2020-13-01T00:00:00Z", "2023-02-29T00:00:00Z", "0000-01-01" + stunde,
				"2020-01-01T00:00:00", "2020-01-01T00:00:00.Z", " 2020-01-01T00:00:00Z", "2020-01-01T00:00:00Z\n"}},
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
	if gueltig != 33 || ungueltig != 107 || zellen != 560 {
		t.Errorf("Menge gültig %d, ungültig %d, Zellen %d; will 33, 107, 560", gueltig, ungueltig, zellen)
	}
}

// TestArtGegenTyp hält Zeile 10: Wert.Art gehört zum EdmType; null nur bei nullbarer Property und mit leerem Text.
func TestArtGegenTyp(t *testing.T) {
	typen := model.UnterstuetzteEDMTypen()
	namen := []string{"S", "B", "I32", "I64", "Dec", "G", "D", "T"}
	proben := map[model.WertArt]string{
		model.ArtString: "a", model.ArtBoolean: "true", model.ArtInt32: "1", model.ArtInt64: "1", model.ArtDecimal: "1.5",
		model.ArtGUID: "0f8fad5b-d9cb-469f-a165-70867728950e", model.ArtDate: "2024-01-05", model.ArtDateTimeOffset: "2020-01-01T00:00:00Z",
	}
	arten := []model.WertArt{model.ArtString, model.ArtBoolean, model.ArtInt32, model.ArtInt64, model.ArtDecimal, model.ArtGUID, model.ArtDate, model.ArtDateTimeOffset}
	gueltig, ungueltig := 0, 0
	for i := range typen {
		for j, art := range arten {
			a := alleAntwort(pw(namen[i], art, proben[art]))
			if i == j {
				gueltig++
				entityBody(t, typen[i]+" "+string(art), a)
			} else {
				ungueltig++
				entityScheitert(t, typen[i]+" mit Art "+string(art), a)
			}
		}
	}
	for i := range namen {
		ungueltig++
		entityScheitert(t, "null bei nicht nullbarer "+namen[i], alleAntwort(pw(namen[i], model.ArtNull, "")))
		gueltig++
		if body := entityBody(t, "null bei nullbarer", alleAntwort(pw("N"+namen[i], model.ArtNull, ""))); !strings.HasSuffix(body, `"N`+namen[i]+`":null}`) {
			t.Errorf("null bei nullbarer %s: %s", namen[i], body)
		}
		ungueltig++
		entityScheitert(t, "null mit Text", alleAntwort(pw("N"+namen[i], model.ArtNull, "x")))
		ungueltig++
		entityScheitert(t, "redigiert", alleAntwort(pw("N"+namen[i], model.ArtRedigiert, "")))
		for _, art := range []model.WertArt{"", "unbekannt"} {
			ungueltig++
			entityScheitert(t, "Art "+string(art), alleAntwort(pw("N"+namen[i], art, proben[arten[i]])))
		}
	}
	if gueltig != 16 || ungueltig != 96 {
		t.Errorf("Menge gültig %d, ungültig %d; will 16, 96", gueltig, ungueltig)
	}
}

// TestZahlform hält Zeile 23: int32 immer Zahl; int64, decimal und @odata.count nach der Zahlform.
func TestZahlform(t *testing.T) {
	werte := []struct {
		prop string
		art  model.WertArt
		text string
		wert bool // in der Zeichenkette-Form ein String
	}{
		{"I32", model.ArtInt32, "2147483647", false},
		{"I64", model.ArtInt64, "9223372036854775807", true},
		{"Dec", model.ArtDecimal, "123456789012345678901234567890.123456789", true},
		{"I32", model.ArtInt32, "-2147483648", false},
		{"I64", model.ArtInt64, "-9223372036854775808", true},
		{"Dec", model.ArtDecimal, "-0.5", true},
	}
	n := 0
	for _, w := range werte {
		for _, accept := range []string{"", "application/json;IEEE754Compatible=true"} {
			e := alleAntwort(pw(w.prop, w.art, w.text))
			c := model.CollectionAntwort{Wurzel: wurzel, Set: "Alles", Entities: []model.Entity{e.Entity}}
			if accept != "" {
				e.Accept, c.Accept = []string{accept}, []string{accept}
			}
			want := `"` + w.prop + `":` + w.text
			if accept != "" && w.wert {
				want = `"` + w.prop + `":"` + w.text + `"`
			}
			if body := entityBody(t, w.prop, e); !strings.HasSuffix(body, want+"}") {
				t.Errorf("Entity %s %q: %s", w.prop, accept, body)
			}
			if body := collectionBody(t, w.prop, c); !strings.HasSuffix(body, want+"}]}") {
				t.Errorf("Collection %s %q: %s", w.prop, accept, body)
			}
			n += 2
		}
	}
	for accept, want := range map[string]string{"": `"@odata.count":9223372036854775807,"value"`, "application/json;IEEE754Compatible=true": `"@odata.count":"9223372036854775807","value"`} {
		c := gutCollection()
		c.Anzahl = model.Anzahl{Vorhanden: true, Wert: 9223372036854775807}
		if accept != "" {
			c.Accept = []string{accept}
		}
		if body := collectionBody(t, "Anzahl", c); !strings.Contains(body, want) {
			t.Errorf("Anzahl %q: %s", accept, body)
		}
		n++
	}
	if n != 26 {
		t.Errorf("Menge %d, will 26 (6 Werte × 2 Formen × 2 Operationen + Anzahl × 2 Formen)", n)
	}
}

// sonderModell hat Namen mit dem Zeichen z an den vier Stellen: Property, Set, Auswahl (Navigation-Property) und Typ.
func sonderModell(z string) model.ServiceModel { return modellStellen(z, z, z, z) }

// modellStellen hat an jeder der vier Stellen den Zusatz des Namens: Property `P`, Set `Set`, Typ `Ns.D` und Navigation-Property `Nav`.
func modellStellen(zProp, zSet, zTyp, zNav string) model.ServiceModel {
	props := []model.Property{prop("Id", "Edm.Int32", false), prop("P"+zProp, "Edm.String", true)}
	typen := []model.EntityType{
		{Name: "Ns.B", Keys: []string{"Id"}, Properties: props, NavigationProperties: []model.NavigationProperty{{Name: "Nav" + zNav, Target: "Ns.B"}}},
		{Name: "Ns.D" + zTyp, BaseType: "Ns.B", Keys: []string{"Id"}, Properties: props},
	}
	return model.NeuesServiceModel("4.0", []string{"Ns"}, typen, []model.EntitySet{{Name: "Set" + zSet, EntityType: "Ns.B"}}, nil)
}

// TestUngueltigesUTF8AnVierStellen hält Zeile 27 (Folge): ein Name, der kein gültiges UTF-8 ist, ist ErrUngueltig an jeder Stelle, an der
// ein Name in die Ausgabe geht (Property, Set, Typ, Navigation-Property in der Auswahl), an Entity und Collection.
func TestUngueltigesUTF8AnVierStellen(t *testing.T) {
	n := 0
	for _, z := range []string{"\xff", "\xc3", "\xed\xa0\x80", "\xc0\x80"} {
		for stelle := range 4 {
			var zs [4]string
			zs[stelle] = z
			m := modellStellen(zs[0], zs[1], zs[2], zs[3])
			e := model.EntityAntwort{Wurzel: wurzel, Set: "Set" + zs[1], Auswahl: []string{"Nav" + zs[3]},
				Entity: model.Entity{Typ: "Ns.D" + zs[2], Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("P"+zs[0], model.ArtString, "x")}}}
			c := model.CollectionAntwort{Wurzel: wurzel, Set: e.Set, Auswahl: e.Auswahl, Entities: []model.Entity{e.Entity}}
			scheitert(t, fmt.Sprintf("Entity Stelle %d %q", stelle, z), func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
				return schreiber().SchreibeEntity(ctx, d, m, e)
			})
			scheitert(t, fmt.Sprintf("Collection Stelle %d %q", stelle, z), func(ctx context.Context, d io.Writer) (model.Darstellung, error) {
				return schreiber().SchreibeCollection(ctx, d, m, c)
			})
			n += 2
		}
	}
	if n != 32 {
		t.Errorf("Menge %d, will 32", n)
	}
	m := modellStellen("", "", "", "")
	zl := neuesZiel()
	e := model.EntityAntwort{Wurzel: wurzel, Set: "Set", Auswahl: []string{"Nav"}, Entity: model.Entity{Typ: "Ns.D", Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("P", model.ArtString, "x")}}}
	if _, err := schreiber().SchreibeEntity(context.Background(), zl, m, e); err != nil {
		t.Errorf("Gegenfall ohne Zusatz: %v", err)
	}
}

// TestEscapes hält Zeile 27: der Zeichensatz des Escapes an fünf Stellen.
func TestEscapes(t *testing.T) {
	erwartet := func(z string) string {
		switch {
		case z == `"`:
			return bs + `"`
		case z == bs:
			return bs + bs
		case z[0] < 0x20:
			return esc(fmt.Sprintf("%04x", z[0]))
		}
		return z
	}
	n := 0
	var zeichen []string
	for c := range 0x20 {
		zeichen = append(zeichen, string(rune(c)))
	}
	zeichen = append(zeichen, " ", `"`, bs, "/", "\x7f", "\xe2\x80\xa8", "\xe2\x80\xa9", "\xc3\xa9", "\xe2\x82\xac", "\xf0\x90\x80\x80")
	for _, z := range zeichen {
		a := gutEntity()
		a.Entity.Properties = []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("Name", model.ArtString, "a"+z+"b")}
		want := `"Name":"a` + erwartet(z) + `b"}`
		if body := entityBody(t, fmt.Sprintf("Wert %q", z), a); !strings.HasSuffix(body, want) {
			t.Errorf("Wert %q: %s, will Ende %s", z, body, want)
		}
		n++
	}
	for _, z := range []string{`"`, bs, "\x01", "\x1f"} {
		m := sonderModell(z)
		e := model.EntityAntwort{Wurzel: wurzel, Set: "Set" + z, Auswahl: []string{"P" + z},
			Entity: model.Entity{Typ: "Ns.D" + z, Schluessel: "1", Properties: []model.PropertyWert{pw("Id", model.ArtInt32, "1"), pw("P"+z, model.ArtString, "x")}}}
		pct := fmt.Sprintf("%%%02X", z[0])
		x := erwartet(z)
		want := `{"@odata.context":"` + wurzel + `/$metadata#Set` + x + `(P` + x + `)/$entity","@odata.type":"#Ns.D` + x + `","@odata.id":"` + wurzel + `/Set` + pct + `(1)","Id":1,"P` + x + `":"x"}`
		zl := neuesZiel()
		if _, err := schreiber().SchreibeEntity(context.Background(), zl, m, e); err != nil || zl.body() != want {
			t.Errorf("Stellen mit %q: %v\n got %s\nwant %s", z, err, zl.body(), want)
		}
		n += 4
	}
	if n != 58 {
		t.Errorf("Menge %d, will 58 (41 Zeichen und das Leerzeichen als Grenze, dazu 4 mal 4)", n)
	}
}
