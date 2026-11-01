package outboundtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen W1 bis W7 und W11 prüfen die Werte der sieben Typen (SPEC-172 Nr. 4 mit SPEC-126 Nr. 10); die Zeilen W8 bis W10 zum
// Zeitpunkt stehen in odata_codec_v2_zeit.go. Die Generatoren der Zahlen (zahlTexte, Orakel math/big) sind die der Suite für
// Version 4, denn die Schreibweise der Zahlen ist in beiden Versionen dieselbe.

// v2WertFaelleFuer bildet aus den Werten die Fälle für die Property prop von W in allen drei Formen des Bodys, oder nur in der
// Collection (nurKol), wenn die Menge groß ist und die Form für den Wert keinen Unterschied macht.
func v2WertFaelleFuer(prefix, prop string, werte []wertFall, nurKol bool) []jfall {
	var out []jfall
	for _, w := range werte {
		name := fmt.Sprintf("%s %.40s", prefix, w.json)
		if w.name != "" {
			name = prefix + " " + w.name
		}
		members := `"Id":1,"` + prop + `":` + w.json
		var fs []jfall
		if w.ursache != "" {
			fs = v2Beide(name, "W", "Ws", members, w.ursache, "/"+prop)
		} else {
			fs = v2BeideGut(name, "W", "Ws", members, pw("Id", model.ArtInt32, "1"), pw(prop, w.art, w.text))
		}
		if nurKol {
			fs = fs[:1]
		}
		out = append(out, fs...)
	}
	return out
}

// v2WertFaelle ist v2WertFaelleFuer für die Property `R<Typ>` (nicht nullable).
func v2WertFaelle(prefix, edm string, werte []wertFall, nurKol bool) []jfall {
	return v2WertFaelleFuer(prefix, "R"+kurz(edm), werte, nurKol)
}

// v2Wertformen sind die sieben JSON-Typen: String, Zahl, true, false, null, Objekt, Array.
func v2Wertformen() []string {
	return []string{"string", "number", "true", "false", "null", "object", "array"}
}

// v2Zelle liefert die Zelle der Tabelle Typ × JSON-Typ aus SPEC-170 Nr. 4 und SPEC-126 Nr. 10 a: den JSON-Wert der Form für den
// Typ und die Art und den Text, wenn die Form erlaubt ist, sonst die Ursache `typ`. null: Nullable steuert die Ursache `null`.
func v2Zelle(edm, form string) wertFall {
	str := map[string]wertFall{
		"Edm.String": gutW(`"a"`, model.ArtString, "a"), "Edm.Int64": gutW(`"5"`, model.ArtInt64, "5"), "Edm.Decimal": gutW(`"1.50"`, model.ArtDecimal, "1.5"),
		"Edm.Guid": gutW(`"`+strings.ToUpper(guidKlein)+`"`, model.ArtGUID, guidKlein), "Edm.DateTimeOffset": gutW(v2Dt("0"), model.ArtDateTimeOffset, "1970-01-01T00:00:00Z"),
	}
	zahl := map[string]wertFall{
		"Edm.Int32": gutW(`1`, model.ArtInt32, "1"), "Edm.Int64": gutW(`5`, model.ArtInt64, "5"), "Edm.Decimal": gutW(`1.50`, model.ArtDecimal, "1.5"),
	}
	typ := func(json string) wertFall { return schlechtW(json, model.DatenTyp) }
	switch form {
	case "string":
		if w, ok := str[edm]; ok {
			return w
		}
		return typ(`"a"`)
	case "number":
		if w, ok := zahl[edm]; ok {
			return w
		}
		return typ(`1`)
	case "true", "false":
		if edm == "Edm.Boolean" {
			return gutW(form, model.ArtBoolean, form)
		}
		return typ(form)
	case "object":
		return typ(`{}`)
	}
	return typ(`[]`)
}

// v2W1: jeder der sieben Typen mal jeder der sieben JSON-Typen: erlaubt, `typ` oder (bei null) das Ergebnis nach Nullable.
func v2FaelleW1() []jfall {
	var out []jfall
	for _, edm := range v2Typen() {
		for _, form := range v2Wertformen() {
			n, r := "N"+kurz(edm), "R"+kurz(edm)
			if form == "null" {
				out = append(out, v2WertFaelleFuer("W1", n, []wertFall{gutW(`null`, model.ArtNull, "")}, false)...)
				out = append(out, v2WertFaelleFuer("W1", r, []wertFall{schlechtW(`null`, model.DatenNull)}, false)...)
				continue
			}
			w := v2Zelle(edm, form)
			w.name = edm + " als " + form
			out = append(out, v2WertFaelleFuer("W1", r, []wertFall{w}, false)...)
			out = append(out, v2WertFaelleFuer("W1", n, []wertFall{w}, false)...)
		}
	}
	return out
}

func v2W1(t *testing.T, s *jsuite) {
	if len(v2Typen())*len(v2Wertformen()) != 49 {
		t.Errorf("W1: %d Zellen, erwartet 49", len(v2Typen())*len(v2Wertformen()))
	}
	s.pruefeAlle(t, v2FaelleW1())
}

func v2FaelleW2() []jfall {
	lang := strings.Repeat("ä", 100000)
	var out []jfall
	for _, w := range []struct{ json, text string }{
		{`""`, ""}, {`"a"`, "a"}, {"\"é\"", "é"}, {"\"e" + string(rune(0x301)) + "\"", "e" + string(rune(0x301))},
		{`"\n \" \\ \/ \b \f \r \t"`, "\n \" \\ / \b \f \r \t"}, {`"\/"`, "/"}, {`"` + esc("0000") + `"`, "\x00"},
		{`"` + esc("d83d") + esc("de00") + `"`, "😀"}, {`"` + lang + `"`, lang}, {`"` + esc("005b") + `REDACTED]"`, "[REDACTED]"},
		{`"[redacted]"`, "[redacted]"}, {`"[REDACTED] "`, "[REDACTED] "}, {`" "`, " "}, {`"true"`, "true"}, {`"/Date(0)/"`, "/Date(0)/"},
	} {
		out = append(out, v2BeideGut(fmt.Sprintf("W2 String %.30s", w.json), "W", "Ws", `"Id":1,"RString":`+w.json,
			pw("Id", model.ArtInt32, "1"), pw("RString", model.ArtString, w.text))...)
	}
	return append(out, v2WertFaelle("W2 Boolean", "Edm.Boolean", []wertFall{
		gutW(`true`, model.ArtBoolean, "true"), gutW(`false`, model.ArtBoolean, "false"),
		schlechtW(`"true"`, model.DatenTyp), schlechtW(`1`, model.DatenTyp), schlechtW(`0`, model.DatenTyp), schlechtW(`[]`, model.DatenTyp),
		schlechtW(`{}`, model.DatenTyp), schlechtW(`"false"`, model.DatenTyp), schlechtW(`null`, model.DatenNull),
	}, false)...)
}

func v2W2(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleW2())
	// U+00E9 zusammengesetzt und zerlegt sind verschiedene Werte; gleiche Zeichen gleicher Text.
	a := s.lauf(jfall{m: "W", set: "Ws", body: v2Kol(`{"Id":1,"RString":"é"}`, `{"Id":2,"RString":"é"}`, `{"Id":3,"RString":"e`+string(rune(0x301))+`"}`)})
	if a.err != nil || len(a.seite.Entities) != 3 {
		t.Fatalf("W2: %+v", a)
	}
	e := func(i int) string { return a.seite.Entities[i].Properties[1].Wert.Text }
	if e(0) != e(1) || e(0) == e(2) {
		t.Errorf("W2: gleiche Zeichen gleicher Text, zusammengesetzt gegen zerlegt verschieden: %q %q %q", e(0), e(1), e(2))
	}
}

func v2FaelleW3() []jfall {
	w := []wertFall{
		gutW(`0`, model.ArtInt32, "0"), gutW(`-0`, model.ArtInt32, "0"), gutW(`1`, model.ArtInt32, "1"), gutW(`-1`, model.ArtInt32, "-1"),
		gutW(`2147483647`, model.ArtInt32, "2147483647"), gutW(`-2147483648`, model.ArtInt32, "-2147483648"),
		schlechtW(`2147483648`, model.DatenWert), schlechtW(`-2147483649`, model.DatenWert), gutW(`1.0`, model.ArtInt32, "1"),
		gutW(`1e2`, model.ArtInt32, "100"), gutW(`100E-0`, model.ArtInt32, "100"), gutW(`1E+2`, model.ArtInt32, "100"),
		gutW(`10e-1`, model.ArtInt32, "1"), gutW(`0.5e1`, model.ArtInt32, "5"), schlechtW(`1.5`, model.DatenWert), schlechtW(`1e-1`, model.DatenWert),
		gutW(`2147483647.0`, model.ArtInt32, "2147483647"), schlechtW(`2147483648.0`, model.DatenWert), schlechtW(`1e999999999`, model.DatenWert),
		schlechtW(`1e-999999999`, model.DatenWert), schlechtW(`"5"`, model.DatenTyp), schlechtW(`""`, model.DatenTyp), gutW(`0e999999999`, model.ArtInt32, "0"),
	}
	return verbinde(v2WertFaelle("W3", "Edm.Int32", w, false), v2WertFaelle("W3 Generator", "Edm.Int32", generierteZahlen("Edm.Int32"), true))
}

func v2W3(t *testing.T, s *jsuite) {
	if n := len(generierteZahlen("Edm.Int32")); n != len(zahlTexte()) || n != 1188 {
		t.Errorf("W3: %d Zahlen im Generator, erwartet 1188", n)
	}
	s.pruefeAlle(t, v2FaelleW3())
}

func v2FaelleW4() []jfall {
	w := []wertFall{
		gutW(`0`, model.ArtInt64, "0"), gutW(`9223372036854775807`, model.ArtInt64, "9223372036854775807"), schlechtW(`9223372036854775808`, model.DatenWert),
		gutW(`-9223372036854775808`, model.ArtInt64, "-9223372036854775808"), schlechtW(`-9223372036854775809`, model.DatenWert),
		gutW(`9007199254740992`, model.ArtInt64, "9007199254740992"), gutW(`9007199254740993`, model.ArtInt64, "9007199254740993"),
		gutW(`"9007199254740993"`, model.ArtInt64, "9007199254740993"), gutW(`"0"`, model.ArtInt64, "0"),
		gutW(`"9223372036854775807"`, model.ArtInt64, "9223372036854775807"), schlechtW(`"9223372036854775808"`, model.DatenWert),
		gutW(`"1e2"`, model.ArtInt64, "100"), gutW(`"1.0"`, model.ArtInt64, "1"), gutW(`"-0"`, model.ArtInt64, "0"), schlechtW(`" 5"`, model.DatenWert),
		schlechtW(`"5 "`, model.DatenWert), schlechtW(`"+5"`, model.DatenWert), schlechtW(`""`, model.DatenWert), schlechtW(`"abc"`, model.DatenWert),
		schlechtW(`"0x10"`, model.DatenWert), schlechtW(`"NaN"`, model.DatenWert), schlechtW(`"1,5"`, model.DatenWert), schlechtW(`"01"`, model.DatenWert),
		schlechtW(`1e999999999`, model.DatenWert), schlechtW(`"1e999999999"`, model.DatenWert), schlechtW(`"1."`, model.DatenWert),
		schlechtW(`".5"`, model.DatenWert), schlechtW(`"1e"`, model.DatenWert), schlechtW(`"-"`, model.DatenWert), schlechtW(`"1.5.5"`, model.DatenWert),
		schlechtW(`1.5`, model.DatenWert), gutW(`1e2`, model.ArtInt64, "100"), gutW(`5`, model.ArtInt64, "5"), gutW(`"5"`, model.ArtInt64, "5"),
	}
	return verbinde(v2WertFaelle("W4", "Edm.Int64", w, false), v2WertFaelle("W4 Generator", "Edm.Int64", generierteZahlen("Edm.Int64"), true))
}

func v2W4(t *testing.T, s *jsuite) {
	if n := len(generierteZahlen("Edm.Int64")); n <= len(zahlTexte()) {
		t.Errorf("W4: %d Zahlen im Generator, erwartet mehr als %d (mit den Stringformen)", n, len(zahlTexte()))
	}
	s.pruefeAlle(t, v2FaelleW4())
	// 5 und "5" ergeben denselben Wert, 2^53+1 behält seine Ziffern.
	r := s.lauf(jfall{m: "W", set: "Ws", body: v2Kol(`{"Id":1,"RInt64":5}`, `{"Id":2,"RInt64":"5"}`, `{"Id":3,"RInt64":9007199254740993}`)})
	text := func(i int) string { return r.seite.Entities[i].Properties[1].Wert.Text }
	if r.err != nil || text(0) != text(1) || text(2) != "9007199254740993" {
		t.Errorf("W4: %+v %v", r.seite, r.err)
	}
}

func v2FaelleW5() []jfall {
	w := []wertFall{
		gutW(`1`, model.ArtDecimal, "1"), gutW(`1.0`, model.ArtDecimal, "1"), gutW(`1.00`, model.ArtDecimal, "1"), gutW(`1e0`, model.ArtDecimal, "1"),
		gutW(`10e-1`, model.ArtDecimal, "1"), gutW(`0.1e1`, model.ArtDecimal, "1"), gutW(`"1"`, model.ArtDecimal, "1"), gutW(`"1.0"`, model.ArtDecimal, "1"),
		gutW(`0`, model.ArtDecimal, "0"), gutW(`-0`, model.ArtDecimal, "0"), gutW(`0.0`, model.ArtDecimal, "0"), gutW(`0e5`, model.ArtDecimal, "0"),
		gutW(`-0.0e-9`, model.ArtDecimal, "0"), gutW(`"0"`, model.ArtDecimal, "0"), gutW(`-1.5`, model.ArtDecimal, "-1.5"), gutW(`-15e-1`, model.ArtDecimal, "-1.5"),
		gutW(`"-1.50"`, model.ArtDecimal, "-1.5"), gutW(`0.5`, model.ArtDecimal, "0.5"), gutW(`5e-1`, model.ArtDecimal, "0.5"),
		gutW(`100`, model.ArtDecimal, "100"), gutW(`1e2`, model.ArtDecimal, "100"), gutW(`100.00`, model.ArtDecimal, "100"),
		gutW(`12345678901234567890.123456789012345`, model.ArtDecimal, "12345678901234567890.123456789012345"),
		gutW(`0e99999999999`, model.ArtDecimal, "0"), schlechtW(`""`, model.DatenWert), schlechtW(`"1,5"`, model.DatenWert),
		schlechtW(`"NaN"`, model.DatenWert), schlechtW(`"INF"`, model.DatenWert), schlechtW(`"1."`, model.DatenWert), schlechtW(`".5"`, model.DatenWert),
		schlechtW(`"1e"`, model.DatenWert), schlechtW(`"-"`, model.DatenWert), schlechtW(`"01"`, model.DatenWert), schlechtW(`" 1"`, model.DatenWert),
		schlechtW(`true`, model.DatenTyp), gutW(`0.000001`, model.ArtDecimal, "0.000001"), gutW(`123e-2`, model.ArtDecimal, "1.23"),
	}
	return verbinde(v2WertFaelle("W5", "Edm.Decimal", w, false), v2WertFaelle("W5 Generator", "Edm.Decimal", generierteZahlen("Edm.Decimal"), true))
}

func v2W5(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleW5())
	// Äquivalenzklassen: gleiche Werte gleicher Text, verschiedene Werte verschiedene Texte.
	r := s.lauf(jfall{m: "W", set: "Ws", body: v2Kol(`{"Id":1,"RDecimal":1}`, `{"Id":2,"RDecimal":"1.0"}`, `{"Id":3,"RDecimal":10e-1}`, `{"Id":4,"RDecimal":1.01}`)})
	text := func(i int) string { return r.seite.Entities[i].Properties[1].Wert.Text }
	if r.err != nil || text(0) != text(1) || text(0) != text(2) || text(0) == text(3) {
		t.Errorf("W5 Äquivalenzklassen: %+v %v", r.seite, r.err)
	}
}

// v2FaelleW6: die Grenze von 1024 Zeichen der kanonischen Schreibweise einer Dezimalzahl (`zahl_zu_lang`) und dieselbe Zusage für
// Edm.Int32, Edm.Int64 und `__count` ohne eigene Ursache (SPEC-126 Nr. 5): Schleife über die Grenzfälle mal vier Stellen.
func v2FaelleW6() []jfall {
	if n := len(dezimalStellen()); n != 5*2*3*3 {
		panic(fmt.Sprintf("W6: %d Fälle der Länge je Stelle des Punkts, erwartet 90", n))
	}
	out := verbinde(v2WertFaelle("W6 Länge", "Edm.Decimal", dezimalLaenge(), false), v2WertFaelle("W6 Stelle des Punkts", "Edm.Decimal", dezimalStellen(), false))
	for _, j := range []string{`1e999999999`, `-1e999999999`, `1` + strings.Repeat("0", 1024), `1e1024`, `1e-1024`} {
		out = append(out, v2WertFaelle("W6 Int32", "Edm.Int32", []wertFall{schlechtW(j, model.DatenWert)}, false)...)
	}
	for _, j := range []string{`1e999999999`, `"1e999999999"`, `1` + strings.Repeat("0", 1024), `"1` + strings.Repeat("0", 1024) + `"`} {
		out = append(out, v2WertFaelle("W6 Int64", "Edm.Int64", []wertFall{schlechtW(j, model.DatenWert)}, false)...)
		out = append(out, jfall{name: "W6 __count " + fmt.Sprintf("%.20s", j), m: "R", set: "Products", body: []byte(`{"d":{"__count":` + j + `,"results":[]}}`),
			ursache: model.DatenWert, ort: "/d/__count"})
	}
	return out
}

func v2W6(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleW6()) }

// v2GuidFaelle: die Form `8-4-4-4-12` mit Hex-Ziffern in jeder Schreibweise, sonst `wert`: ein Fall je Position (Nicht-Hex an
// den 32 Ziffernstellen, ein anderes Zeichen an den 4 Strichstellen, ein fehlender Strich, ein zusätzlicher Strich an jeder Stelle).
func v2GuidFaelle() []wertFall {
	base := guidKlein
	gut := func(s string) wertFall { return gutW(`"`+s+`"`, model.ArtGUID, guidKlein) }
	schlecht := func(s string) wertFall { return schlechtW(`"`+s+`"`, model.DatenWert) }
	out := []wertFall{gut(base), gut(strings.ToUpper(base)), gut("0F8fAD5b-D9cb-469F-a165-70867728950E"), schlecht("{" + base + "}"), schlecht(strings.ReplaceAll(base, "-", "")),
		schlecht(" " + base), schlecht(base + " "), schlecht(base[:35]), schlecht(base + "0"), schlecht(""), schlecht("urn:uuid:" + base),
		schlecht(strings.ToUpper(base[:8]) + "_" + base[9:]), schlechtW(`1`, model.DatenTyp), schlechtW(`true`, model.DatenTyp)}
	for i := range len(base) {
		ersatz := "g"
		if base[i] == '-' {
			ersatz = "x"
		}
		out = append(out, schlecht(base[:i]+ersatz+base[i+1:]))
		out = append(out, schlecht(base[:i]+base[i+1:]))
	}
	for i := 0; i <= len(base); i++ {
		out = append(out, schlecht(base[:i]+"-"+base[i:]))
	}
	for _, i := range []int{8, 13, 18, 23} {
		out = append(out, schlecht(base[:i]+"0"+base[i+1:]))
	}
	return out
}

func v2FaelleW7() []jfall { return v2WertFaelle("W7", "Edm.Guid", v2GuidFaelle(), false) }

func v2W7(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleW7()) }

// v2FaelleW11: `null` und der Marker `[REDACTED]` je Typ (SPEC-126 Nr. 10 e, f mit SPEC-172 Nr. 4): der Marker (Rohtext ohne Escape)
// ist redigiert für Edm.String, Edm.Guid, Edm.Int64, Edm.Decimal und Edm.DateTime, für Edm.Int32 und Edm.Boolean `typ`; mit
// Escape ein gewöhnlicher String (für die vier Formtypen `wert`). Schleife 7 Typen mal {Marker, Marker mit Escape, null} mal
// {Nullable, nicht}.
func v2FaelleW11() []jfall {
	var out []jfall
	for _, edm := range v2Typen() {
		alsString := edm == "Edm.String" || edm == "Edm.Guid" || edm == "Edm.Int64" || edm == "Edm.Decimal" || edm == "Edm.DateTimeOffset"
		for _, nullable := range []bool{true, false} {
			prop := "R" + kurz(edm)
			if nullable {
				prop = "N" + kurz(edm)
			}
			marker := schlechtW(`"[REDACTED]"`, model.DatenTyp)
			mitEscape := schlechtW(`"`+esc("005B")+`REDACTED]"`, model.DatenTyp)
			switch {
			case edm == "Edm.String":
				marker, mitEscape = gutW(`"[REDACTED]"`, model.ArtRedigiert, ""), gutW(`"`+esc("005B")+`REDACTED]"`, model.ArtString, "[REDACTED]")
			case alsString:
				marker, mitEscape = gutW(`"[REDACTED]"`, model.ArtRedigiert, ""), schlechtW(`"`+esc("005B")+`REDACTED]"`, model.DatenWert)
			}
			null := schlechtW(`null`, model.DatenNull)
			if nullable {
				null = gutW(`null`, model.ArtNull, "")
			}
			marker.name, mitEscape.name, null.name = "Marker", "Marker mit Escape", "null"
			out = append(out, v2WertFaelleFuer("W11 "+prop, prop, []wertFall{marker, mitEscape, null}, false)...)
		}
	}
	return out
}

func v2W11(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleW11()) }

func v2FaelleW() []jfall {
	return verbinde(v2FaelleW1(), v2FaelleW2(), v2FaelleW3(), v2FaelleW4(), v2FaelleW5(), v2FaelleW6(), v2FaelleW7(), v2FaelleW11())
}
