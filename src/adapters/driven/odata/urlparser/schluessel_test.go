package urlparser_test

import (
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	keyU = model.URLKey
	litU = model.URLLiteral
	typU = model.URLLiteralTyp
)

// TestKeyFormen (Zeile 29): sechs gültige Formen des Schlüssels in der Reihenfolge der URL.
func TestKeyFormen(t *testing.T) {
	eu := teil("Region", textLit("'EU'"))
	faelle := []fall{
		{pfad: "/P(42)", want: entity("P", teil("", zahlLit("42")))},
		{pfad: "/P('abc')", want: entity("P", teil("", textLit("'abc'")))},
		{pfad: "/P(Id=42)", want: entity("P", teil("Id", zahlLit("42")))},
		{pfad: "/P(Id=42,Region='EU')", want: entity("P", teil("Id", zahlLit("42")), eu)},
		{pfad: "/P(Region='EU',Id=42)", want: entity("P", eu, teil("Id", zahlLit("42")))},
		{pfad: "/P(C=3,A=1,B=2)", want: entity("P", teil("C", zahlLit("3")), teil("A", zahlLit("1")), teil("B", zahlLit("2")))},
	}
	if len(faelle) != 6 {
		t.Fatalf("%d Formen", len(faelle))
	}
	pruefeFaelle(t, faelle)
}

// TestKeyStruktur (Zeile 30): 19 Körper und Segmente, die gegen die Struktur des Schlüssels verstoßen, sind `key`.
func TestKeyStruktur(t *testing.T) {
	koerper := []string{"(1,2)", "(Id=1,2)", "(1,Id=2)", "(Id=1,Id=2)", "(Id=1,Id=1)", "(=1)", "(Id=)", "(Id)", "(Id=1,)", "(,Id=1)",
		"(1,,2)", "(1x=1)", "(a-b=1)", "(Id=1,Region=)", "(Id=1,=2)", "(Id=1,Id=2,Id=3)", "(1)x", "(", "(1"}
	var faelle []fall
	for _, k := range koerper {
		faelle = append(faelle, fall{pfad: "/Set" + k, u: keyU, ort: "pfad/1"})
	}
	if len(faelle) != 19 {
		t.Fatalf("%d Körper", len(faelle))
	}
	pruefeFaelle(t, faelle)
}

// TestKeyLeerraum (Zeile 31): Leerraum gehört zu keiner Grammatik; 6 Stellen × 2 Schreibweisen = 12 Fälle.
func TestKeyLeerraum(t *testing.T) {
	stellen := []struct {
		pfad string
		u    model.URLUrsache
	}{
		{"/S(%sId=1,Region='EU')", keyU}, {"/S(Id=1,Region='EU'%s)", litU}, {"/S(Id=1%s,Region='EU')", litU},
		{"/S(Id=1,%sRegion='EU')", keyU}, {"/S(Id%s=1,Region='EU')", keyU}, {"/S(Id=%s1,Region='EU')", litU},
	}
	n := 0
	for _, st := range stellen {
		for _, w := range []string{"%20", "%09"} {
			pruefeFall(t, parserNeu(), fall{pfad: eingesetzt(st.pfad, w), u: st.u, ort: "pfad/1"})
			n++
		}
	}
	if n != 12 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestRangImSegment (Zeile 10): die Folge innerhalb eines Segments, 4 Paare, dazu die unverschlossene Klammer hinter einer
// ungültigen Kodierung.
func TestRangImSegment(t *testing.T) {
	pruefeFaelle(t, []fall{
		{pfad: "/Set(1%G1", u: kodi, ort: "pfad/1"},              // Kodierung vor der unverschlossenen Klammer
		{pfad: "/S(Id=1e3,Region=null)", u: litU, ort: "pfad/1"}, // 400 gewinnt gegen ein 501 früher im selben Körper
		{pfad: "/Set%G1x(", u: kodi, ort: "pfad/1"},              // Kodierung vor Name und Klammer
		{pfad: "/Set-(1,2)", u: segm, ort: "pfad/1"},             // Name und Klammer vor der Struktur
		{pfad: "/Set(1,2,1x)", u: keyU, ort: "pfad/1"},           // Struktur vor den Literalen
		{pfad: "/Set(" + wiederhole("1,", 17) + "1x=)", u: model.URLZuViele, ort: "pfad/1"},
		{pfad: "/Set(A=1x,B=" + wiederhole("a", 513) + ")", u: litU, ort: "pfad/1"}, // Literale von links nach rechts
	})
}

// literalFall ist ein Schlüsselliteral als `/S(Id=<l>)` mit dem erwarteten Text oder der Ursache.
type literalFall struct {
	eingabe string
	art     model.LiteralArt
	text    string
	u       model.URLUrsache
}

// pruefeLiterale liest jedes Literal benannt (`Id=<l>`) und unbenannt (`<l>`); ein nacktes Wort ist unbenannt ein
// Name ohne Literal (`key`), deshalb entfällt dort die unbenannte Form.
func pruefeLiterale(t *testing.T, faelle []literalFall) {
	t.Helper()
	for _, f := range faelle {
		formen := []string{"Id=" + f.eingabe}
		if f.u == "" || f.u == typU || !nacktesWort(f.eingabe) {
			formen = append(formen, f.eingabe)
		}
		for _, k := range formen {
			name := ""
			if k != f.eingabe {
				name = "Id"
			}
			fa := fall{pfad: "/S(" + k + ")", u: f.u, ort: "pfad/1"}
			if f.u == "" {
				fa.want = entity("S", teil(name, model.Literal{Art: f.art, Text: f.text}))
			}
			pruefeFall(t, parserNeu(), fa)
		}
	}
}

func nacktesWort(s string) bool {
	for _, r := range s {
		if r != '_' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// TestLiteralString (Zeile 32): 20 Zeichenketten, 14 gültige und 6 ungültige. Die Eingaben entstehen aus Bytewerten.
func TestLiteralString(t *testing.T) {
	rueck := string([]byte{0x5c})
	z, a := model.LiteralZeichenkette, "\xc3\xa4"
	faelle := []literalFall{
		{"''", z, "''", ""}, {"'a'", z, "'a'", ""}, {"'O''Hara'", z, "'O''Hara'", ""}, {"''''", z, "''''", ""},
		{"'a,b'", z, "'a,b'", ""}, {"'a)b'", z, "'a)b'", ""}, {"'a(b'", z, "'a(b'", ""}, {"'a=b'", z, "'a=b'", ""},
		{"'a%2Fb'", z, "'a/b'", ""}, {"'a%5Cb'", z, "'a" + rueck + "b'", ""}, {"'" + a + "'", z, "'" + a + "'", ""},
		{"'%C3%A4'", z, "'" + a + "'", ""}, {"'%27%27'", z, "''''", ""}, {"'x%0Ay'", z, "'x\ny'", ""},
		{"'a", 0, "", litU}, {"'a'b", 0, "", litU}, {"'", 0, "", litU}, {"a'", 0, "", litU}, {"'a''", 0, "", litU}, {"'a'x'", 0, "", litU},
	}
	if len(faelle) != 20 {
		t.Fatalf("%d Fälle", len(faelle))
	}
	pruefeLiterale(t, faelle)
	// Das Komma und die Klammer im Literal trennen keinen Teil; der Schlüssel bleibt einer.
	if u := gelesen(t, "/S(Id='a,b)c',N='x')", ""); len(u.Ressource.Schluessel) != 2 || u.Ressource.Schluessel[0].Literal.Text != "'a,b)c'" {
		t.Errorf("Teile: %+v", u.Ressource.Schluessel)
	}
}

// TestLiteralBool (Zeile 33): `true` und `false` in jeder ASCII-Schreibweise (16 + 32), 9 ungültige Formen als Wert benannter
// Teile und 12 nackte Wörter ohne `=`.
func TestLiteralBool(t *testing.T) {
	n := 0
	for _, wort := range []string{"true", "false"} {
		for maske := range 1 << len(wort) {
			b := []byte(wort)
			for i := range b {
				if maske>>i&1 == 1 {
					b[i] -= 'a' - 'A'
				}
			}
			pruefeLiterale(t, []literalFall{{string(b), model.LiteralWahrheitswert, wort, ""}})
			n++
		}
	}
	for _, s := range []string{"t", "tru", "truee", "yes", "1", "0", "null", "NULL", "nUll"} {
		f := fall{pfad: "/S(Id=" + s + ")", u: litU, ort: "pfad/1"}
		if s == "1" || s == "0" { // keine Wahrheitswerte, aber Zahlen
			f = fall{pfad: f.pfad, want: entity("S", teil("Id", zahlLit(s)))}
		}
		pruefeFall(t, parserNeu(), f)
		n++
	}
	// 12 nackte Wörter ohne `=` (SPEC-161 Nr. 5 c): sechs Ausnahmen gehören dem Literal, sechs Namen sind `key`.
	nackt := []struct {
		wort string
		u    model.URLUrsache
		want *model.ODataURL
	}{
		{"true", "", entity("S", teil("", model.Literal{Art: model.LiteralWahrheitswert, Text: "true"}))},
		{"FALSE", "", entity("S", teil("", model.Literal{Art: model.LiteralWahrheitswert, Text: "false"}))},
		{"null", litU, nil}, {"nUll", litU, nil}, {"INF", typU, nil}, {"NaN", typU, nil},
		{"t", keyU, nil}, {"tru", keyU, nil}, {"truee", keyU, nil}, {"yes", keyU, nil}, {"inf", keyU, nil}, {"nan", keyU, nil},
	}
	for _, w := range nackt {
		pruefeFall(t, parserNeu(), fall{pfad: "/S(" + w.wort + ")", u: w.u, ort: "pfad/1", want: w.want})
		n++
	}
	if n != 69 {
		t.Fatalf("%d Fälle, erwartet 69", n)
	}
}

// TestLiteralZahl (Zeile 34): kanonischer Text; 12 gültige, 13 ungültige, 6 nicht unterstützte Formen.
func TestLiteralZahl(t *testing.T) {
	z := model.LiteralZahl
	lang := wiederhole("9", 38)
	gueltig := []literalFall{
		{"0", z, "0", ""}, {"-0", z, "0", ""}, {"00", z, "0", ""}, {"007", z, "7", ""}, {"42", z, "42", ""}, {"-42", z, "-42", ""},
		{"1.5", z, "1.5", ""}, {"1.50", z, "1.5", ""}, {"0.0", z, "0", ""}, {"-0.0", z, "0", ""}, {"0.10", z, "0.1", ""}, {lang, z, lang, ""},
	}
	ungueltig := []string{"+1", ".5", "5.", "-", "1.2.3", "42L", "1.5M", "1d", "0x10", "\xef\xbc\x91", "\xd9\xa1", "--1", "(1)"}
	nicht := []string{"1e3", "1E3", "1.5e-2", "INF", "-INF", "NaN"}
	var faelle []literalFall
	faelle = append(faelle, gueltig...)
	for _, s := range ungueltig {
		faelle = append(faelle, literalFall{eingabe: s, u: litU})
	}
	for _, s := range nicht {
		faelle = append(faelle, literalFall{eingabe: s, u: typU})
	}
	if len(faelle) != 31 {
		t.Fatalf("%d Fälle", len(faelle))
	}
	pruefeLiterale(t, faelle)
}

// TestLiteralGuid (Zeile 35): kleingeschrieben; 3 gültige und 8 ungültige Formen.
func TestLiteralGuid(t *testing.T) {
	g := "0f8fad5b-d9cb-469f-a165-70867728950e"
	faelle := []literalFall{
		{g, model.LiteralGUID, g, ""}, {"0F8FAD5B-D9CB-469F-A165-70867728950E", model.LiteralGUID, g, ""},
		{"0f8FAD5b-D9cb-469F-a165-70867728950E", model.LiteralGUID, g, ""},
	}
	for _, s := range []string{"{" + g + "}", "0f8fad5bd9cb469fa16570867728950e", "0f8fad5b-d9cb-469f-a165-70867728950e0",
		"0f8fad5b-d9cb-469f-a165-7086772895", "0f8fad5b-d9cb-469f-a165-70867728950g", "0f8fad5b0-d9c-469f-a165-70867728950e",
		"0f8fad5b-d9cb-469f-a165-708677289500e", "urn:uuid:" + g} {
		faelle = append(faelle, literalFall{eingabe: s, u: litU})
	}
	if len(faelle) != 11 {
		t.Fatalf("%d Fälle", len(faelle))
	}
	pruefeLiterale(t, faelle)
}

func tageMonat(jahr, monat int) int {
	switch monat {
	case 4, 6, 9, 11:
		return 30
	case 2:
		if jahr%4 == 0 && (jahr%100 != 0 || jahr%400 == 0) {
			return 29
		}
		return 28
	}
	return 31
}

// TestLiteralDatum (Zeile 36): Form und Bereich, beide Seiten jeder Grenze; das Orakel rechnet den Kalender selbst.
func TestLiteralDatum(t *testing.T) {
	n := 0
	for _, jahr := range []int{0, 1, 1900, 2000, 2023, 2024, 9999} {
		for monat := 0; monat <= 13; monat++ {
			for tag := 0; tag <= 32; tag++ {
				s := fmt.Sprintf("%04d-%02d-%02d", jahr, monat, tag)
				gueltig := jahr >= 1 && monat >= 1 && monat <= 12 && tag >= 1 && tag <= tageMonat(jahr, monat)
				f := literalFall{eingabe: s, u: litU}
				if gueltig {
					f = literalFall{s, model.LiteralDatum, s, ""}
				}
				pruefeLiterale(t, []literalFall{f})
				n++
			}
		}
	}
	for _, s := range []string{"10000-01-01", "2024-1-5", "+2024-01-05", "2024-01-05%20", "2024.01.05"} {
		pruefeLiterale(t, []literalFall{{eingabe: s, u: litU}})
	}
	// `20240105` ist kein Datum, aber eine Zahl.
	pruefeLiterale(t, []literalFall{{"20240105", model.LiteralZahl, "20240105", ""}})
	if n != 7*14*33 {
		t.Fatalf("%d Fälle", n)
	}
}

// TestLiteralZeitpunkt (Zeile 37): UTC, kanonisch; die Tabelle nennt den Text, den das Orakel (time) liefert.
func TestLiteralZeitpunkt(t *testing.T) {
	z := model.LiteralZeitpunkt
	faelle := []literalFall{
		{"2020-01-01T23:59:59Z", z, "2020-01-01T23:59:59Z", ""}, {"2020-01-01T24:00:00Z", 0, "", litU},
		{"2020-01-01T00:59:00Z", z, "2020-01-01T00:59:00Z", ""}, {"2020-01-01T00:60:00Z", 0, "", litU},
		{"2020-01-01T00:00:59Z", z, "2020-01-01T00:00:59Z", ""}, {"2020-01-01T00:00:60Z", 0, "", litU},
		{"2020-01-01T10:00:00%2B23:59", z, "2019-12-31T10:01:00Z", ""}, {"2020-01-01T10:00:00%2B24:00", 0, "", litU},
		{"2020-01-01T10:00:00%2B00:60", 0, "", litU},
		{"2020-01-01T00:00:00.5Z", z, "2020-01-01T00:00:00.5Z", ""}, {"2020-01-01T00:00:00.123456789012Z", z, "2020-01-01T00:00:00.123456789012Z", ""},
		{"2020-01-01T00:00:00.1234567890123Z", 0, "", litU}, {"2020-01-01T00:00:00.500Z", z, "2020-01-01T00:00:00.5Z", ""},
		{"2020-01-01T00:00:00.000Z", z, "2020-01-01T00:00:00Z", ""}, {"2020-01-01T00:00Z", z, "2020-01-01T00:00:00Z", ""},
		{"2020-01-01T01:00:00Z", z, "2020-01-01T01:00:00Z", ""}, {"2020-01-01T01:00:00%2B00:00", z, "2020-01-01T01:00:00Z", ""},
		{"2020-01-01T02:00:00%2B01:00", z, "2020-01-01T01:00:00Z", ""}, {"2020-01-01T00:00:00-01:00", z, "2020-01-01T01:00:00Z", ""},
		{"2020-01-01t00:00:00Z", 0, "", litU}, {"2020-01-01T00:00:00z", 0, "", litU}, {"2020-01-01%2000:00:00Z", 0, "", litU},
		{"0001-01-01T00:00:00%2B01:00", 0, "", litU}, {"9999-12-31T23:59:59-01:00", 0, "", litU},
		{"0001-01-01T01:00:00%2B01:00", z, "0001-01-01T00:00:00Z", ""}, {"9999-12-31T22:59:59-01:00", z, "9999-12-31T23:59:59Z", ""},
		{"2020-01-01T00:00:00+01:00", z, "2019-12-31T23:00:00Z", ""}, {"2020-01-01T00:00:00Z.5", 0, "", litU},
	}
	pruefeLiterale(t, faelle)
}

// TestLiteralTyp (Zeile 38): Literale eines Typs, den das Modell nicht trägt, sind `literal_typ`; `binary'AA` ist `literal`.
func TestLiteralTyp(t *testing.T) {
	var faelle []literalFall
	for _, s := range []string{"binary'AA=='", "X'AA'", "duration'P1D'", "geography'POINT(1 2)'", "geometry'POINT(1 2)'", "Ns.Farbe'Rot'",
		"datetime'2020-01-01T00:00:00'", "12:30", "12:30:00", "12:30:00.5"} {
		faelle = append(faelle, literalFall{eingabe: s, u: typU})
	}
	if len(faelle) != 10 {
		t.Fatalf("%d Fälle", len(faelle))
	}
	faelle = append(faelle, literalFall{eingabe: "binary'AA", u: litU}, literalFall{eingabe: "25:30", u: litU})
	pruefeLiterale(t, faelle)
}

// TestBeispieleSpec126 (Zeile 40): die Beispiele von SPEC-126 Nr. 11 sind die Texte der Literale; 9 Werte.
func TestBeispieleSpec126(t *testing.T) {
	g, z := "0f8fad5b-d9cb-469f-a165-70867728950e", model.LiteralZeichenkette
	faelle := []literalFall{
		{"42", model.LiteralZahl, "42", ""}, {"true", model.LiteralWahrheitswert, "true", ""}, {"1.5", model.LiteralZahl, "1.5", ""},
		{g, model.LiteralGUID, g, ""}, {"2020-01-01", model.LiteralDatum, "2020-01-01", ""},
		{"2020-01-01T00:00:00Z", model.LiteralZeitpunkt, "2020-01-01T00:00:00Z", ""}, {"'O''Hara'", z, "'O''Hara'", ""}, {"''", z, "''", ""},
	}
	pruefeLiterale(t, faelle)
	u := gelesen(t, "/P(Id=1,Region='EU')", "")
	if len(u.Ressource.Schluessel) != 2 || u.Ressource.Schluessel[0].Literal.Text != "1" || u.Ressource.Schluessel[1].Literal.Text != "'EU'" {
		t.Errorf("Schlüssel: %+v", u.Ressource.Schluessel)
	}
}
