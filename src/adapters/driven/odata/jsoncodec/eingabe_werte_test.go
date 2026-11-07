package jsoncodec_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eWertFall ist ein Element der Menge der Wertformen (Zeile 24): der JSON-Wert und Art und Text oder die Ursache.
type eWertFall struct {
	json, text string
	ursache    model.EingabeUrsache
}

func ok(json, text string) eWertFall { return eWertFall{json: json, text: text} }

func bad(json string, u model.EingabeUrsache) eWertFall { return eWertFall{json: json, ursache: u} }

// eWertTabelle ist die zweite Tabelle der Zeile 24: je Property von W gültige und ungültige Formen, Orakel ist die Tabelle selbst.
func eWertTabelle() map[string][]eWertFall {
	nul, u1, u2 := esc("0000"), esc("D83D"), esc("DE00")
	const guid = "0f8fad5b-d9cb-469f-a165-70867728950e"
	return map[string][]eWertFall{
		"S": {ok(`""`, ""), ok(`"abc"`, "abc"), ok("\"\xc3\xa9\"", "\xc3\xa9"), ok("\"e\xcc\x81\"", "e\xcc\x81"), ok(`"`+nul+`"`, "\x00"),
			ok(`"`+u1+u2+`"`, "\xf0\x9f\x98\x80"), ok("\"\xf0\x9f\x98\x80\"", "\xf0\x9f\x98\x80"),
			bad("5", model.EingabeTyp), bad("true", model.EingabeTyp), bad("false", model.EingabeTyp), bad("{}", model.EingabeTyp), bad("[]", model.EingabeTyp)},
		"B": {ok("true", "true"), ok("false", "false"),
			bad(`"true"`, model.EingabeTyp), bad(`""`, model.EingabeTyp), bad("1", model.EingabeTyp), bad("{}", model.EingabeTyp), bad("[]", model.EingabeTyp)},
		"I32": {ok("0", "0"), ok("-0", "0"), ok("1.0", "1"), ok("1e2", "100"), ok("100E-0", "100"), ok("2147483647", "2147483647"), ok("-2147483648", "-2147483648"),
			bad("2147483648", model.EingabeWert), bad("-2147483649", model.EingabeWert), bad("1.5", model.EingabeWert), bad("1e-1", model.EingabeWert),
			bad(`"5"`, model.EingabeTyp), bad("true", model.EingabeTyp), bad("1e999999999", model.EingabeWert), bad("{}", model.EingabeTyp)},
		"I64": {ok("0", "0"), ok("-0", "0"), ok("5", "5"), ok("1.0", "1"), ok("1e2", "100"), ok("9223372036854775807", "9223372036854775807"),
			ok("-9223372036854775808", "-9223372036854775808"), ok(`"5"`, "5"), ok(`"9223372036854775807"`, "9223372036854775807"),
			bad("9223372036854775808", model.EingabeWert), bad("-9223372036854775809", model.EingabeWert), bad("1.5", model.EingabeWert),
			bad("1e-1", model.EingabeWert), bad(`"1.5"`, model.EingabeWert), bad(`"abc"`, model.EingabeWert), bad(`""`, model.EingabeWert),
			bad("true", model.EingabeTyp), bad("1e999999999", model.EingabeWert), bad(`"1e999999999"`, model.EingabeWert), bad("{}", model.EingabeTyp),
			bad(`" 5"`, model.EingabeWert)},
		"Dec": {ok("0", "0"), ok("-0", "0"), ok("1", "1"), ok("1.0", "1"), ok("1e0", "1"), ok("10e-1", "1"), ok(`"1"`, "1"), ok("1.50", "1.5"),
			ok("-1.25", "-1.25"), ok("1e2", "100"), ok("0.001", "0.001"), ok(`"123456789012345678901234567890.5"`, "123456789012345678901234567890.5"),
			bad(`"abc"`, model.EingabeWert), bad(`""`, model.EingabeWert), bad(`"1,5"`, model.EingabeWert), bad(`" 5"`, model.EingabeWert),
			bad(`"0x10"`, model.EingabeWert), bad(`"NaN"`, model.EingabeWert), bad(`"INF"`, model.EingabeWert), bad("true", model.EingabeTyp),
			bad("{}", model.EingabeTyp), bad("[]", model.EingabeTyp), bad(`"+1"`, model.EingabeWert), bad("1e999999999", model.EingabeZahlZuLang)},
		"G": {ok(`"`+guid+`"`, guid), ok(`"`+strings.ToUpper(guid)+`"`, guid), ok(`"0F8fad5B-d9CB-469f-A165-70867728950e"`, guid),
			bad(`"{`+guid+`}"`, model.EingabeWert), bad(`"0f8fad5bd9cb469fa16570867728950e"`, model.EingabeWert), bad(`" `+guid+` "`, model.EingabeWert),
			bad(`"`+guid[1:]+`"`, model.EingabeWert), bad(`"`+guid+`0"`, model.EingabeWert), bad(`"`+guid[:35]+`g"`, model.EingabeWert),
			bad(`"urn:uuid:`+guid+`"`, model.EingabeWert)},
		"D": {ok(`"2020-01-01"`, "2020-01-01"), ok(`"0001-01-01"`, "0001-01-01"), ok(`"9999-12-31"`, "9999-12-31"), ok(`"2024-02-29"`, "2024-02-29"),
			ok(`"2000-02-29"`, "2000-02-29"),
			bad(`"2023-02-29"`, model.EingabeWert), bad(`"0000-01-01"`, model.EingabeWert), bad(`"2024-1-5"`, model.EingabeWert), bad(`"20240105"`, model.EingabeWert),
			bad(`"2020-01-01T00:00:00Z"`, model.EingabeWert), bad(`"+2024-01-01"`, model.EingabeWert), bad(`"2024-13-01"`, model.EingabeWert),
			bad(`"2024-00-01"`, model.EingabeWert), bad(`"2024-01-00"`, model.EingabeWert), bad(`"2024-04-31"`, model.EingabeWert),
			bad(`"2020-01-01 "`, model.EingabeWert), bad(`"1900-02-29"`, model.EingabeWert), bad("5", model.EingabeTyp)},
		"T": eZeitpunkte(),
	}
}

func eZeitpunkte() []eWertFall {
	z := func(in, soll string) eWertFall { return ok(`"`+in+`"`, soll) }
	w := func(in string) eWertFall { return bad(`"`+in+`"`, model.EingabeWert) }
	return []eWertFall{
		z("2020-01-01T00:00:00Z", "2020-01-01T00:00:00Z"), z("2020-01-01T00:00Z", "2020-01-01T00:00:00Z"),
		z("2020-01-01T01:00:00+01:00", "2020-01-01T00:00:00Z"), z("2020-01-01T00:00:00-01:00", "2020-01-01T01:00:00Z"),
		z("2020-01-01T00:00:00.5Z", "2020-01-01T00:00:00.5Z"), z("2020-01-01T00:00:00.500Z", "2020-01-01T00:00:00.5Z"),
		z("2020-01-01T00:00:00.000Z", "2020-01-01T00:00:00Z"), z("2020-01-01T00:00:00.123456789012Z", "2020-01-01T00:00:00.123456789012Z"),
		z("2020-01-01T23:59:59Z", "2020-01-01T23:59:59Z"), z("2020-01-01T00:00:00+00:00", "2020-01-01T00:00:00Z"),
		z("2020-01-01T00:00:00-00:00", "2020-01-01T00:00:00Z"), z("2020-01-01T00:00+05:30", "2019-12-31T18:30:00Z"),
		z("2020-01-01T00:00:00+23:59", "2019-12-31T00:01:00Z"), z("0001-01-01T01:00:00+01:00", "0001-01-01T00:00:00Z"),
		z("9999-12-31T23:59:59Z", "9999-12-31T23:59:59Z"), z("9999-12-31T23:59:59+01:00", "9999-12-31T22:59:59Z"),
		z("2024-02-29T12:00:00Z", "2024-02-29T12:00:00Z"), z("2020-03-01T00:00:00+01:00", "2020-02-29T23:00:00Z"),
		z("2021-01-01T00:30:00+01:00", "2020-12-31T23:30:00Z"), z("2020-01-01T00:00:00.1+01:00", "2019-12-31T23:00:00.1Z"),
		z("2020-01-01T00:00:00.10Z", "2020-01-01T00:00:00.1Z"),
		w("2020-01-01t00:00:00Z"), w("2020-01-01T00:00:00z"), w("2020-01-01 00:00:00Z"), w("2020-01-01T00:00:60Z"), w("2020-01-01T24:00:00Z"),
		w("2020-01-01T00:60:00Z"), w("2020-01-01T00:00:00"), w("2020-01-01T00:00:00.1234567890123Z"), w("2020-01-01T00:00:00+24:00"),
		w("2020-01-01T00:00:00+01:60"), w("0001-01-01T00:00:00+01:00"), w("9999-12-31T23:59:59-01:00"), bad("5", model.EingabeTyp),
	}
}

// eWArt ist die Art der Property name des Typs W.
func eWArt(name string) model.WertArt {
	return map[string]model.WertArt{"S": model.ArtString, "B": model.ArtBoolean, "I32": model.ArtInt32, "I64": model.ArtInt64,
		"Dec": model.ArtDecimal, "G": model.ArtGUID, "D": model.ArtDate, "T": model.ArtDateTimeOffset}[name]
}

// eWertPruefen liest den Wert f der Property name des Typs W bei der Art und prüft Ergebnis oder Fehler.
func eWertPruefen(t *testing.T, name string, f eWertFall, art model.EingabeArt) {
	t.Helper()
	got, err := eA("Ws", art, eObj("Id", "1", name, f.json)).los()
	was := name + " " + f.json + " " + string(art)
	if f.ursache != "" {
		eFehler(t, was, got, err, f.ursache, "/"+name)
		return
	}
	eOK(t, was, got, err, model.EntityEingabe{Typ: "Ns.W", Schluessel: "1", Properties: []model.PropertyWert{ePW("Id", model.ArtInt32, "1"), ePW(name, eWArt(name), f.text)}})
}

// TestEingabeWerte hält Zeile 24: Art und kanonischer Text, jeder andere Wert `typ`, `wert` oder `zahl_zu_lang`. Menge 141 × 3 = 423 (gültig 66,
// ungültig 75).
func TestEingabeWerte(t *testing.T) {
	gueltig, ungueltig := 0, 0
	for name, faelle := range eWertTabelle() {
		for _, f := range faelle {
			if f.ursache == "" {
				gueltig++
			} else {
				ungueltig++
			}
			for _, art := range eArten() {
				eWertPruefen(t, name, f, art)
			}
		}
	}
	if gueltig != 66 || ungueltig != 75 {
		t.Errorf("gültig %d, ungültig %d, erwartet 66 und 75", gueltig, ungueltig)
	}
}

// TestEingabeZahlform hält Zeile 25: `Edm.Int64` und `Edm.Decimal` als Zahl und als Zeichenkette ergeben dasselbe; `Edm.Int32` als
// Zeichenkette ist `typ`. Menge 48 = 8 Wertpaare × 2 Typen × 3 Arten, dazu 3 Gegenfälle.
func TestEingabeZahlform(t *testing.T) {
	paare := []struct{ zahl, int64Text, decText string }{
		{"5", "5", "5"}, {"0", "0", "0"}, {"-0", "0", "0"}, {"1e2", "100", "100"},
		{"9223372036854775807", "9223372036854775807", "9223372036854775807"}, {"-9223372036854775808", "-9223372036854775808", "-9223372036854775808"},
		{"1.50", "", "1.5"}, {"123456789012345678901234567890.5", "", "123456789012345678901234567890.5"},
	}
	n := 0
	for _, p := range paare {
		for _, typ := range []string{"I64", "Dec"} {
			text := p.int64Text
			if typ == "Dec" {
				text = p.decText
			}
			f := ok(p.zahl, text)
			if text == "" {
				f = bad(p.zahl, model.EingabeWert)
			}
			g := f
			g.json = `"` + p.zahl + `"`
			for _, art := range eArten() {
				eWertPruefen(t, typ, f, art)
				eWertPruefen(t, typ, g, art)
				n++
			}
		}
	}
	if n != 48 {
		t.Errorf("Menge %d, erwartet 48", n)
	}
	for _, s := range []string{`"5"`, `"0"`, `"-1"`} {
		for _, art := range eArten() {
			eWertPruefen(t, "I32", bad(s, model.EingabeTyp), art)
		}
	}
}

// TestEingabeMarker hält Zeile 27: `[REDACTED]` ist ein gewöhnlicher Text, `redigiert` kommt nie vor. Menge 96 = 8 Typen × 4 Schreibweisen × 3 Arten.
func TestEingabeMarker(t *testing.T) {
	schreibweisen := []string{`"[REDACTED]"`, `"[redacted]"`, `"[REDACTED] "`, `"` + esc("005B") + `REDACTED]"`}
	texte := []string{"[REDACTED]", "[redacted]", "[REDACTED] ", "[REDACTED]"}
	ursachen := map[string]model.EingabeUrsache{"S": "", "B": model.EingabeTyp, "I32": model.EingabeTyp, "I64": model.EingabeWert,
		"Dec": model.EingabeWert, "G": model.EingabeWert, "D": model.EingabeWert, "T": model.EingabeWert}
	n := 0
	for _, name := range eWNamen() {
		for i, s := range schreibweisen {
			f := eWertFall{json: s, text: texte[i], ursache: ursachen[name]}
			for _, art := range eArten() {
				eWertPruefen(t, name, f, art)
				n++
			}
		}
	}
	if n != 96 {
		t.Errorf("Menge %d, erwartet 96", n)
	}
}

// TestEingabeDezimalLaenge hält Zeile 28: 1024 Zeichen kanonisch sind erlaubt, 1025 `zahl_zu_lang`; ein Exponent wird nicht ausgeschrieben.
// Menge 6 Werte × 3 Arten = 18, dazu `1e999999999` für Decimal, Int32 und Int64.
func TestEingabeDezimalLaenge(t *testing.T) {
	eins := func(nullen int) string { return "1" + strings.Repeat("0", nullen) }
	n := 0
	for _, l := range []int{1024, 1025} {
		text := eins(l - 1)
		for _, json := range []string{text, `"` + text + `"`, "1e" + eP(l-1)} {
			f := ok(json, text)
			if l == 1025 {
				f = bad(json, model.EingabeZahlZuLang)
			}
			for _, art := range eArten() {
				eWertPruefen(t, "Dec", f, art)
				n++
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
	// Ein Exponent, dessen Ausschreiben 10 Millionen Zeichen ergäbe, wird nicht ausgeschrieben: Zuteilung höchstens 256 · n + 1 MiB (SPEC-181 Nr. 2 c iii).
	m := 0
	for _, f := range []struct {
		typ     string
		ursache model.EingabeUrsache
	}{{"Dec", model.EingabeZahlZuLang}, {"I32", model.EingabeWert}, {"I64", model.EingabeWert}} {
		for _, art := range eArten() {
			body := eObj("Id", "1", f.typ, "1e9999999")
			z := &eZaehler{r: strings.NewReader(body)}
			a := eA("Ws", art, "")
			a.src = z
			var vor, nach runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&vor)
			got, err := a.los()
			runtime.ReadMemStats(&nach)
			eFehler(t, f.typ+" 1e9999999 "+string(art), got, err, f.ursache, "/"+f.typ)
			if zugeteilt, grenze := int64(nach.TotalAlloc-vor.TotalAlloc), 256*z.n+1<<20; zugeteilt > grenze {
				t.Fatalf("%s %s: %d Bytes zugeteilt, höchstens %d erlaubt (n = %d)", f.typ, art, zugeteilt, grenze, z.n)
			}
			m++
		}
	}
	if m != 9 {
		t.Errorf("Menge %d, erwartet 9", m)
	}
}

// TestEingabeDezimalExponentGross hält Zeile 28 für `1e999999999`: der Exponent wird nie ausgeschrieben (ausgeschrieben mehr als 1 GiB). Eigene
// Funktion, damit die Mutation „Exponent ausgeschrieben“ sie mit `-run TestEingabeDezimalLaenge` auslässt und nur gegen ungemutierten Code läuft.
func TestEingabeDezimalExponentGross(t *testing.T) {
	for _, art := range eArten() {
		eWertPruefen(t, "Dec", bad("1e999999999", model.EingabeZahlZuLang), art)
		eWertPruefen(t, "I32", bad("1e999999999", model.EingabeWert), art)
		eWertPruefen(t, "I64", bad("1e999999999", model.EingabeWert), art)
	}
}
