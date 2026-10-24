package outboundtest

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// wertFall ist ein JSON-Wert für eine Property von W mit dem erwarteten Ausgang: die Art und der Text oder die Ursache.
type wertFall struct {
	json    string
	art     model.WertArt
	text    string
	ursache model.DatenUrsache
}

func gutW(json string, art model.WertArt, text string) wertFall {
	return wertFall{json: json, art: art, text: text}
}

func schlechtW(json string, u model.DatenUrsache) wertFall { return wertFall{json: json, ursache: u} }

// wertFaelle bildet aus den Werten die Fälle für die Property `R<Typ>` von W: beide Methoden, oder nur die Collection
// (nurKol), wenn die Menge groß ist und die Methode für den Wert keinen Unterschied macht.
func wertFaelle(prefix, edm string, werte []wertFall, nurKol bool) []jfall {
	prop := "R" + kurz(edm)
	var out []jfall
	for _, w := range werte {
		name := fmt.Sprintf("%s %.40s", prefix, w.json)
		members := `"Id":1,"` + prop + `":` + w.json
		var fs []jfall
		if w.ursache != "" {
			fs = beide(name, "W", "Ws", members, w.ursache, "/"+prop)
		} else {
			fs = beideGut(name, "W", "Ws", members, pw("Id", model.ArtInt32, "1"), pw(prop, w.art, w.text))
		}
		if nurKol {
			fs = fs[:1]
		}
		out = append(out, fs...)
	}
	return out
}

func jfF1() []jfall {
	lang := strings.Repeat("ä", 100000)
	var out []jfall
	for _, w := range []struct{ json, text string }{
		{`""`, ""}, {`"a"`, "a"}, {"\"é\"", "é"}, {"\"é\"", "é"}, {`"\n \" \\ \/ \b \f \r \t"`, "\n \" \\ / \b \f \r \t"},
		{`"` + esc("0000") + `"`, "\x00"}, {`"` + esc("d83d") + esc("de00") + `"`, "😀"}, {`"` + lang + `"`, lang}, {`"` + esc("005b") + `REDACTED]"`, "[REDACTED]"},
		{`"[redacted]"`, "[redacted]"}, {`"[REDACTED] "`, "[REDACTED] "}, {`" "`, " "},
	} {
		members := `"Id":1,"RString":` + w.json
		out = append(out, beideGut(fmt.Sprintf("F1 %.30s", w.json), "W", "Ws", members, pw("Id", model.ArtInt32, "1"), pw("RString", model.ArtString, w.text))...)
	}
	return out
}

func jF1(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, jfF1())
	// U+00E9 zusammengesetzt und zerlegt sind verschiedene Werte; gleicher Text bei gleichen Zeichen.
	a := s.lauf(jfall{m: "W", set: "Ws", body: jKol("Ws", `{"Id":1,"RString":"é"}`, `{"Id":2,"RString":"é"}`, `{"Id":3,"RString":"e`+"́"+`"}`)})
	if a.err != nil || len(a.seite.Entities) != 3 {
		t.Fatalf("F1: %+v", a)
	}
	e := func(i int) string { return a.seite.Entities[i].Properties[1].Wert.Text }
	if e(0) != e(1) || e(0) == e(2) {
		t.Errorf("F1: gleiche Zeichen gleicher Text, zusammengesetzt gegen zerlegt verschieden: %q %q %q", e(0), e(1), e(2))
	}
}

func jfF2() []jfall {
	return wertFaelle("F2", "Edm.Boolean", []wertFall{
		gutW(`true`, model.ArtBoolean, "true"), gutW(`false`, model.ArtBoolean, "false"),
		schlechtW(`"true"`, model.DatenTyp), schlechtW(`1`, model.DatenTyp), schlechtW(`0`, model.DatenTyp), schlechtW(`[]`, model.DatenTyp),
		schlechtW(`{}`, model.DatenTyp), schlechtW(`"false"`, model.DatenTyp), schlechtW(`null`, model.DatenNull),
	}, false)
}

func jF2(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF2()) }

// Zahlengenerator für F3 bis F5: Vorzeichen × Mantisse × Dezimalpunkt × Exponent; das Orakel ist math/big.
func zahlTexte() []string {
	var out []string
	for _, vz := range []string{"", "-"} {
		for _, ganz := range []string{"0", "1", "7", "10", "123456", "2147483648", "9007199254740993", "9223372036854775808", "99999999999999999999"} {
			for _, bruch := range []string{"", ".0", ".5", ".50", ".000", ".123456789012345678901234567890"} {
				for _, exp := range []string{"", "e0", "E0", "e2", "e+2", "E-2", "e-1", "e1", "e10", "e-10", "e19"} {
					out = append(out, vz+ganz+bruch+exp)
				}
			}
		}
	}
	return out
}

// orakelGanz liefert den Wert als ganze Zahl, wenn der Text einen im Bereich von bits Bit nennt.
func orakelGanz(text string, bits uint) (string, bool) {
	r, ok := new(big.Rat).SetString(text)
	if !ok || !r.IsInt() {
		return "", false
	}
	grenze := new(big.Int).Lsh(big.NewInt(1), bits-1)
	n := r.Num()
	if n.Cmp(new(big.Int).Neg(grenze)) < 0 || n.Cmp(grenze) >= 0 {
		return "", false
	}
	return n.String(), true
}

// orakelDezimal liefert die kanonische Dezimalschreibweise: der kleinste Maßstab, in dem der Wert ganz ist.
func orakelDezimal(text string) string {
	r, _ := new(big.Rat).SetString(text)
	if r.IsInt() {
		return r.Num().String()
	}
	zehn := big.NewRat(10, 1)
	skaliert := new(big.Rat).Set(r)
	for k := 1; ; k++ {
		skaliert.Mul(skaliert, zehn)
		if !skaliert.IsInt() {
			continue
		}
		z := skaliert.Num().String()
		vz := ""
		if strings.HasPrefix(z, "-") {
			vz, z = "-", z[1:]
		}
		z = strings.Repeat("0", max(0, k+1-len(z))) + z
		return vz + z[:len(z)-k] + "." + z[len(z)-k:]
	}
}

// generierteZahlen sind die Fälle des Generators für den Typ: in jeder Schreibweise der Zahl und, wo der Typ es zulässt,
// als String.
func generierteZahlen(edm string) []wertFall {
	var out []wertFall
	for _, text := range zahlTexte() {
		var w wertFall
		switch edm {
		case "Edm.Int32":
			w = ganzFall(text, 32, model.ArtInt32)
		case "Edm.Int64":
			w = ganzFall(text, 64, model.ArtInt64)
		default:
			w = gutW(text, model.ArtDecimal, orakelDezimal(text))
		}
		out = append(out, w)
		if edm != "Edm.Int32" && (!strings.ContainsAny(text, "eE") || strings.HasSuffix(text, "e2") || strings.HasSuffix(text, "E-2")) {
			s := w
			s.json = `"` + text + `"`
			out = append(out, s)
		}
	}
	return out
}

func ganzFall(text string, bits uint, art model.WertArt) wertFall {
	if g, ok := orakelGanz(text, bits); ok {
		return gutW(text, art, g)
	}
	return schlechtW(text, model.DatenWert)
}

func jfF3() []jfall {
	w := []wertFall{
		gutW(`0`, model.ArtInt32, "0"), gutW(`-0`, model.ArtInt32, "0"), gutW(`1`, model.ArtInt32, "1"), gutW(`-1`, model.ArtInt32, "-1"),
		gutW(`2147483647`, model.ArtInt32, "2147483647"), gutW(`-2147483648`, model.ArtInt32, "-2147483648"),
		schlechtW(`2147483648`, model.DatenWert), schlechtW(`-2147483649`, model.DatenWert), gutW(`1.0`, model.ArtInt32, "1"),
		gutW(`1e2`, model.ArtInt32, "100"), gutW(`100E-0`, model.ArtInt32, "100"), gutW(`1E+2`, model.ArtInt32, "100"),
		gutW(`10e-1`, model.ArtInt32, "1"), gutW(`0.5e1`, model.ArtInt32, "5"), schlechtW(`1.5`, model.DatenWert), schlechtW(`1e-1`, model.DatenWert),
		gutW(`2147483647.0`, model.ArtInt32, "2147483647"), schlechtW(`2147483648.0`, model.DatenWert), schlechtW(`1e999999999`, model.DatenWert),
		schlechtW(`1e-999999999`, model.DatenWert), schlechtW(`"5"`, model.DatenTyp), schlechtW(`""`, model.DatenTyp), gutW(`0e999999999`, model.ArtInt32, "0"),
	}
	return verbinde(wertFaelle("F3", "Edm.Int32", w, false), wertFaelle("F3 Generator", "Edm.Int32", generierteZahlen("Edm.Int32"), true))
}

func jF3(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF3()) }

func jfF4() []jfall {
	w := []wertFall{
		gutW(`0`, model.ArtInt64, "0"), gutW(`9223372036854775807`, model.ArtInt64, "9223372036854775807"), schlechtW(`9223372036854775808`, model.DatenWert),
		gutW(`-9223372036854775808`, model.ArtInt64, "-9223372036854775808"), schlechtW(`-9223372036854775809`, model.DatenWert),
		gutW(`9007199254740992`, model.ArtInt64, "9007199254740992"), gutW(`9007199254740993`, model.ArtInt64, "9007199254740993"),
		gutW(`"9007199254740993"`, model.ArtInt64, "9007199254740993"), gutW(`"0"`, model.ArtInt64, "0"),
		gutW(`"9223372036854775807"`, model.ArtInt64, "9223372036854775807"), schlechtW(`"9223372036854775808"`, model.DatenWert),
		gutW(`"1e2"`, model.ArtInt64, "100"), gutW(`"1.0"`, model.ArtInt64, "1"), gutW(`"-0"`, model.ArtInt64, "0"), schlechtW(`" 5"`, model.DatenWert),
		schlechtW(`"5 "`, model.DatenWert), schlechtW(`"+5"`, model.DatenWert), schlechtW(`""`, model.DatenWert), schlechtW(`"abc"`, model.DatenWert),
		schlechtW(`"0x1"`, model.DatenWert), schlechtW(`"01"`, model.DatenWert), schlechtW(`1e999999999`, model.DatenWert), schlechtW(`"1e999999999"`, model.DatenWert),
		schlechtW(`"1."`, model.DatenWert), schlechtW(`".5"`, model.DatenWert), schlechtW(`"1e"`, model.DatenWert), schlechtW(`"1e+"`, model.DatenWert),
		schlechtW(`"-"`, model.DatenWert), schlechtW(`"--1"`, model.DatenWert), schlechtW(`"1.5.5"`, model.DatenWert), schlechtW(`"1 "`, model.DatenWert),
		schlechtW(`1.5`, model.DatenWert), gutW(`1e2`, model.ArtInt64, "100"),
	}
	return verbinde(wertFaelle("F4", "Edm.Int64", w, false), wertFaelle("F4 Generator", "Edm.Int64", generierteZahlen("Edm.Int64"), true))
}

func jF4(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF4()) }

// dezimalLaenge liefert die Fälle an der Grenze von 1024 Zeichen in vier Formen: ganzer Teil, Bruchteil, Exponent, Vorzeichen.
func dezimalLaenge() []wertFall {
	eins := func(n int) string { return "1" + strings.Repeat("0", n) }
	return []wertFall{
		gutW(eins(1023), model.ArtDecimal, eins(1023)), schlechtW(eins(1024), model.DatenZahlZuLang),
		gutW(`1e-1022`, model.ArtDecimal, "0."+strings.Repeat("0", 1021)+"1"), schlechtW(`1e-1023`, model.DatenZahlZuLang),
		gutW(`1e1023`, model.ArtDecimal, eins(1023)), schlechtW(`1e1024`, model.DatenZahlZuLang),
		gutW(`-1e1022`, model.ArtDecimal, "-"+eins(1022)), schlechtW(`-1e1023`, model.DatenZahlZuLang),
		gutW(`"`+eins(1023)+`"`, model.ArtDecimal, eins(1023)), schlechtW(`"`+eins(1024)+`"`, model.DatenZahlZuLang),
		gutW(`-`+eins(1022), model.ArtDecimal, "-"+eins(1022)), schlechtW(`-`+eins(1023), model.DatenZahlZuLang),
		gutW(`0.`+strings.Repeat("0", 1021)+`1`, model.ArtDecimal, "0."+strings.Repeat("0", 1021)+"1"),
		schlechtW(`0.`+strings.Repeat("0", 1022)+`1`, model.DatenZahlZuLang),
		gutW(`-1e-1021`, model.ArtDecimal, "-0."+strings.Repeat("0", 1020)+"1"), schlechtW(`-1e-1022`, model.DatenZahlZuLang),
		gutW(`12345e-1022`, model.ArtDecimal, "0."+strings.Repeat("0", 1017)+"12345"), schlechtW(`12345e-1023`, model.DatenZahlZuLang),
	}
}

func jfF5() []jfall {
	w := []wertFall{
		gutW(`1`, model.ArtDecimal, "1"), gutW(`1.0`, model.ArtDecimal, "1"), gutW(`1.00`, model.ArtDecimal, "1"), gutW(`1e0`, model.ArtDecimal, "1"),
		gutW(`10e-1`, model.ArtDecimal, "1"), gutW(`0.1e1`, model.ArtDecimal, "1"), gutW(`"1"`, model.ArtDecimal, "1"), gutW(`"1.0"`, model.ArtDecimal, "1"),
		gutW(`0`, model.ArtDecimal, "0"), gutW(`-0`, model.ArtDecimal, "0"), gutW(`0.0`, model.ArtDecimal, "0"), gutW(`0e5`, model.ArtDecimal, "0"),
		gutW(`-0.0e-9`, model.ArtDecimal, "0"), gutW(`"0"`, model.ArtDecimal, "0"), gutW(`-1.5`, model.ArtDecimal, "-1.5"), gutW(`-15e-1`, model.ArtDecimal, "-1.5"),
		gutW(`"-1.50"`, model.ArtDecimal, "-1.5"), gutW(`0.5`, model.ArtDecimal, "0.5"), gutW(`5e-1`, model.ArtDecimal, "0.5"),
		gutW(`100`, model.ArtDecimal, "100"), gutW(`1e2`, model.ArtDecimal, "100"), gutW(`100.00`, model.ArtDecimal, "100"),
		gutW(`12345678901234567890.123456789012345`, model.ArtDecimal, "12345678901234567890.123456789012345"),
		gutW(`0e99999999999`, model.ArtDecimal, "0"), schlechtW(`1e999999999`, model.DatenZahlZuLang), schlechtW(`-1e999999999`, model.DatenZahlZuLang),
		schlechtW(`"1e999999999"`, model.DatenZahlZuLang), schlechtW(`1e-999999999`, model.DatenZahlZuLang),
		schlechtW(`""`, model.DatenWert), schlechtW(`"1,5"`, model.DatenWert), schlechtW(`"NaN"`, model.DatenWert), schlechtW(`"INF"`, model.DatenWert),
		schlechtW(`"1."`, model.DatenWert), schlechtW(`".5"`, model.DatenWert), schlechtW(`"1e"`, model.DatenWert), schlechtW(`"1e+"`, model.DatenWert),
		schlechtW(`"-"`, model.DatenWert), schlechtW(`"01"`, model.DatenWert), schlechtW(`"1.5.5"`, model.DatenWert), schlechtW(`" 1"`, model.DatenWert),
		schlechtW(`true`, model.DatenTyp), gutW(`0.000001`, model.ArtDecimal, "0.000001"), gutW(`123e-2`, model.ArtDecimal, "1.23"),
	}
	return verbinde(wertFaelle("F5", "Edm.Decimal", w, false), wertFaelle("F5 Länge", "Edm.Decimal", dezimalLaenge(), false),
		wertFaelle("F5 Generator", "Edm.Decimal", generierteZahlen("Edm.Decimal"), true))
}

func jF5(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, jfF5())
	// Äquivalenzklassen: gleiche Werte gleicher Text, verschiedene Werte verschiedene Texte.
	r := s.lauf(jfall{m: "W", set: "Ws", body: jKol("Ws", `{"Id":1,"RDecimal":1}`, `{"Id":2,"RDecimal":"1.0"}`, `{"Id":3,"RDecimal":10e-1}`, `{"Id":4,"RDecimal":1.01}`)})
	text := func(i int) string { return r.seite.Entities[i].Properties[1].Wert.Text }
	if r.err != nil || text(0) != text(1) || text(0) != text(2) || text(0) == text(3) {
		t.Errorf("F5 Äquivalenzklassen: %+v %v", r.seite, r.err)
	}
}

// jfF9: `null` gegen nicht selektiert (SPEC-126 Nr. 9, 10 e) für beide Werte von Nullable und jeden Typ.
func jfF9() []jfall {
	var out []jfall
	for _, edm := range edmTypenJSON() {
		n, r := "N"+kurz(edm), "R"+kurz(edm)
		out = append(out, beideGut("F9 "+n+" null", "W", "Ws", `"Id":1,"`+n+`":null`, pw("Id", model.ArtInt32, "1"), pw(n, model.ArtNull, ""))...)
		out = append(out, beide("F9 "+r+" null", "W", "Ws", `"Id":1,"`+r+`":null`, model.DatenNull, "/"+r)...)
		out = append(out, beideGut("F9 "+n+" fehlt", "W", "Ws", `"Id":1`, pw("Id", model.ArtInt32, "1"))...)
		out = append(out, beideGut("F9 "+r+" fehlt", "W", "Ws", `"Id":1`, pw("Id", model.ArtInt32, "1"))...)
		out = append(out, beideGut("F9 "+n+" null an der Stelle des Bodys", "W", "Ws", `"`+n+`":null,"Id":1`, pw(n, model.ArtNull, ""), pw("Id", model.ArtInt32, "1"))...)
	}
	return out
}

func jF9(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF9()) }
