package outboundtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// typDatum beschreibt, wie ein EDM-Typ JSON-Werte liest (SPEC-126 Nr. 10, Spalte 3): welche JSON-Typen zulässig sind,
// mit einem gültigen und einem ungültigen Inhalt, und ob der Marker `redigiert` ergibt.
type typDatum struct {
	edm      string
	art      model.WertArt
	zahl     string // gültige JSON-Zahl; leer, wenn der Typ keine zulässt
	zahlText string
	str      string // gültiger JSON-String; leer, wenn der Typ keinen zulässt
	strText  string
	schlecht string // JSON-String mit ungültigem Inhalt (`wert`); leer, wenn jeder Inhalt gültig ist
	boolOK   bool
	markbar  bool // der Marker ergibt `redigiert`
}

func typDaten() []typDatum {
	return []typDatum{
		{"Edm.String", model.ArtString, "", "", `"a"`, "a", "", false, true},
		{"Edm.Boolean", model.ArtBoolean, "", "", "", "", "", true, false},
		{"Edm.Int32", model.ArtInt32, `7`, "7", "", "", "", false, false},
		{"Edm.Int64", model.ArtInt64, `7`, "7", `"7"`, "7", `"x"`, false, true},
		{"Edm.Decimal", model.ArtDecimal, `1.50`, "1.5", `"1.50"`, "1.5", `"x"`, false, true},
	}
}

// f10Fall ist ein JSON-Wert gegen einen Typ mit dem Ausgang nach der Tabelle von Nr. 10.
func f10Faelle(d typDatum) []wertFall {
	art := func(json string, ok bool, text string) wertFall {
		if ok {
			return gutW(json, d.art, text)
		}
		return schlechtW(json, model.DatenTyp)
	}
	out := []wertFall{
		art(`true`, d.boolOK, "true"), art(`false`, d.boolOK, "false"),
		schlechtW(`{"a":1}`, model.DatenTyp), schlechtW(`{}`, model.DatenTyp), schlechtW(`[]`, model.DatenTyp), schlechtW(`[1]`, model.DatenTyp),
	}
	if d.zahl != "" {
		out = append(out, gutW(d.zahl, d.art, d.zahlText))
	} else {
		out = append(out, schlechtW(`7`, model.DatenTyp), schlechtW(`1.5`, model.DatenTyp))
	}
	if d.str != "" {
		out = append(out, gutW(d.str, d.art, d.strText))
	} else {
		out = append(out, schlechtW(`"7"`, model.DatenTyp), schlechtW(`"x"`, model.DatenTyp))
	}
	if d.schlecht != "" {
		out = append(out, schlechtW(d.schlecht, model.DatenWert))
	}
	return out
}

func jfF10() []jfall {
	var out []jfall
	for _, d := range typDaten() {
		out = append(out, wertFaelle("F10 "+kurz(d.edm), d.edm, f10Faelle(d), false)...)
		n := "N" + kurz(d.edm)
		out = append(out, beideGut("F10 "+kurz(d.edm)+" null", "W", "Ws", `"Id":1,"`+n+`":null`, pw("Id", model.ArtInt32, "1"), pw(n, model.ArtNull, ""))...)
	}
	return out
}

func jF10(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF10()) }

// f11Fassungen sind die Schreibweisen des Markers: der Rohtext, die Fassung mit Escape, Kleinbuchstaben, Leerraum dahinter
// und davor, der leere String.
func f11Fassungen() []string {
	return []string{`"[REDACTED]"`, `"` + esc("005b") + `REDACTED]"`, `"[redacted]"`, `"[REDACTED] "`, `" [REDACTED]"`, `""`}
}

// f11Fall ist der Ausgang einer Fassung des Markers (Index in f11Fassungen) für den Typ.
func f11Fall(d typDatum, i int) wertFall {
	json := f11Fassungen()[i]
	switch {
	case d.edm == "Edm.String":
		texte := []string{"", "[REDACTED]", "[redacted]", "[REDACTED] ", " [REDACTED]", ""}
		if i == 0 {
			return wertFall{json: json, art: model.ArtRedigiert}
		}
		return gutW(json, model.ArtString, texte[i])
	case !d.markbar:
		return schlechtW(json, model.DatenTyp)
	case i == 0:
		return wertFall{json: json, art: model.ArtRedigiert}
	}
	return schlechtW(json, model.DatenWert)
}

func jfF11() []jfall {
	var out []jfall
	for _, d := range typDaten() {
		var ws []wertFall
		for i := range f11Fassungen() {
			ws = append(ws, f11Fall(d, i))
		}
		out = append(out, wertFaelle("F11 "+kurz(d.edm), d.edm, ws, false)...)
	}
	return out
}

func jF11(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfF11()) }

// schluesselWert ist ein Wert eines Schlüssels mit seiner Literalform und den Schreibweisen im Body.
type schluesselWert struct {
	literal string
	json    []string
}

func schluesselWerte(edm string) []schluesselWert {
	switch edm {
	case "Edm.String":
		return []schluesselWert{
			{"'a'", []string{`"a"`, `"` + esc("0061") + `"`}}, {"'O''Hara'", []string{`"O'Hara"`}}, {"''", []string{`""`}}, {"''''", []string{`"'"`}},
			{"'a,b'", []string{`"a,b"`}}, {"'A=1'", []string{`"A=1"`}}, {"'(x)'", []string{`"(x)"`}}, {"'a b'", []string{`"a b"`}},
			{"'a\nb'", []string{`"a\nb"`}}, {"'" + strings.Repeat("z", 1000) + "'", []string{`"` + strings.Repeat("z", 1000) + `"`}},
			{"'ä'", []string{`"ä"`, `"` + esc("00e4") + `"`}}, {"'A'", []string{`"A"`}},
		}
	case "Edm.Boolean":
		return []schluesselWert{{"true", []string{`true`}}, {"false", []string{`false`}}}
	case "Edm.Int32":
		return []schluesselWert{{"42", []string{`42`, `42.0`, `4.2e1`, `420e-1`}}, {"7", []string{`7`}}, {"-5", []string{`-5`, `-5.0`}}, {"0", []string{`0`, `-0`, `0e5`}}}
	case "Edm.Int64":
		return []schluesselWert{
			{"9007199254740993", []string{`9007199254740993`, `"9007199254740993"`, `9007199254740993.0`}},
			{"9007199254740992", []string{`9007199254740992`, `"9007199254740992"`}}, {"-1", []string{`-1`, `"-1"`}},
		}
	}
	return []schluesselWert{{"1.5", []string{`1.5`, `15e-1`, `"1.50"`}}, {"1", []string{`1`, `1.0`, `"1"`, `10e-1`}}, {"0.1", []string{`0.1`, `1e-1`}}, {"-2.5", []string{`-2.5`, `-25e-1`}}}
}

func jfG1() []jfall {
	var out []jfall
	for _, d := range typDaten() {
		for _, w := range schluesselWerte(d.edm) {
			for _, j := range w.json {
				out = append(out, beideGut(fmt.Sprintf("G1 %s %.30s", kurz(d.edm), j), "T", "K"+kurz(d.edm), `"Id":`+j,
					pw("Id", d.art, textDesLiterals(d.edm, w.literal)))...)
				out[len(out)-1].schluessel, out[len(out)-2].schluessel = w.literal, w.literal
			}
		}
	}
	// Ein Schlüssel darf nullable deklariert sein (SPEC-125); gelesen wird er wie jeder andere.
	nullable := beideGut("G1 Schlüssel mit Nullable wahr", "T", "KN", `"Id":5`, pw("Id", model.ArtInt32, "5"))
	nullable[0].schluessel, nullable[1].schluessel = "5", "5"
	return append(out, nullable...)
}

// textDesLiterals ist der kanonische Text zu der Literalform des Schlüssels.
func textDesLiterals(edm, literal string) string {
	if edm == "Edm.String" {
		return strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(literal, "'"), "'"), "''", "'")
	}
	return literal
}

func jG1(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, jfG1())
	// Allaussage über alle Paare der Menge eines Typs: der Schlüssel ist genau bei gleichem Wert gleich.
	for _, d := range typDaten() {
		type eintrag struct{ gruppe, schluessel, json string }
		var es []eintrag
		for g, w := range schluesselWerte(d.edm) {
			for _, j := range w.json {
				r := s.lauf(jfall{m: "T", set: "K" + kurz(d.edm), einzel: true, body: jEin("K"+kurz(d.edm), `"Id":`+j)})
				if r.err != nil {
					t.Fatalf("G1 %s %s: %v", d.edm, j, r.err)
				}
				es = append(es, eintrag{fmt.Sprint(g), r.entity.Schluessel, j})
			}
		}
		for _, a := range es {
			for _, b := range es {
				if (a.gruppe == b.gruppe) != (a.schluessel == b.schluessel) {
					t.Errorf("G1 %s: %s (%q) gegen %s (%q): Schlüssel gleich genau bei gleichem Wert", d.edm, a.json, a.schluessel, b.json, b.schluessel)
				}
			}
		}
	}
}

// permutationen liefert alle Anordnungen von 0 bis n−1.
func permutationen(n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	var out [][]int
	for _, p := range permutationen(n - 1) {
		for i := 0; i <= len(p); i++ {
			q := append(append(append([]int{}, p[:i]...), n-1), p[i:]...)
			out = append(out, q)
		}
	}
	return out
}

func jfG2() []jfall {
	var out []jfall
	for _, st := range []struct {
		name, m, set, typ string
		teile             []string // Members in der Reihenfolge des Modells
		soll              string
	}{
		{"K2", "T", "K2", "", []string{`"A":1`, `"B":"x"`, `"Note":"n"`}, `A=1,B='x'`},
		{"K3", "T", "K3", "", []string{`"A":1`, `"B":"x"`, `"C":true`, `"Note":"n"`}, `A=1,B='x',C=true`},
		{"Lines", "K", "Lines", "", []string{`"OrderID":5`, `"LineNo":2`, `"Note":"n"`}, `OrderID=5,LineNo=2`},
		{"Gift", "K", "Lines", "Shop.Gift", []string{`"OrderID":5`, `"LineNo":2`, `"Note":"n"`}, `OrderID=5,LineNo=2`},
		{"KS", "T", "KS", "", []string{`"A":"x"`, `"B":"y"`}, `A='x',B='y'`},
	} {
		for _, p := range permutationen(len(st.teile)) {
			var ms []string
			if st.typ != "" {
				ms = append(ms, `"@odata.type":"#`+st.typ+`"`)
			}
			for _, i := range p {
				ms = append(ms, st.teile[i])
			}
			fs := beideGut(fmt.Sprintf("G2 %s Reihenfolge %v", st.name, p), st.m, st.set, strings.Join(ms, ","))
			for i := range fs {
				fs[i].schluessel = st.soll
			}
			out = append(out, fs...)
		}
	}
	return append(out, jfG2Paare()...)
}

// jfG2Paare: zwei Schlüssel, deren Literale ohne Verdopplung der Anführungszeichen zusammenfielen.
func jfG2Paare() []jfall {
	return []jfall{
		{name: "G2 Paar 1", m: "T", set: "KS", einzel: true, body: jEin("KS", `"A":"x',B='y","B":"z"`), schluessel: `A='x'',B=''y',B='z'`},
		{name: "G2 Paar 2", m: "T", set: "KS", einzel: true, body: jEin("KS", `"A":"x","B":"y',B='z"`), schluessel: `A='x',B='y'',B=''z'`},
	}
}

func jG2(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, jfG2())
	a := s.lauf(jfG2Paare()[0])
	b := s.lauf(jfG2Paare()[1])
	if a.err != nil || b.err != nil || a.entity.Schluessel == b.entity.Schluessel {
		t.Errorf("G2 Paare mit Trennzeichen im Text: %q %q (%v %v), erwartet verschieden", a.entity.Schluessel, b.entity.Schluessel, a.err, b.err)
	}
}

// schluesselProp ist ein Schlüssel eines Setups mit der Art seines Werts.
type schluesselProp struct{ name, art string }

type schluesselSetup struct {
	name, m, set, typ string
	keys              []schluesselProp
}

func g3Setups() []schluesselSetup {
	return []schluesselSetup{
		{"einzeln Int32", "T", "KInt32", "", []schluesselProp{{"Id", "int32"}}},
		{"einzeln String", "T", "KString", "", []schluesselProp{{"Id", "string"}}},
		{"einzeln Int64", "T", "KInt64", "", []schluesselProp{{"Id", "int64"}}},
		{"zwei", "T", "K2", "", []schluesselProp{{"A", "int32"}, {"B", "string"}}},
		{"drei", "T", "K3", "", []schluesselProp{{"A", "int32"}, {"B", "string"}, {"C", "boolean"}}},
		{"geerbt", "K", "Lines", "Shop.Gift", []schluesselProp{{"OrderID", "int64"}, {"LineNo", "int32"}}},
		{"im Typ von odata.type", "T", "Abs", "T.Der", []schluesselProp{{"Id", "int32"}}},
	}
}

func gutWert(art string) string {
	return map[string]string{"int32": `1`, "string": `"x"`, "boolean": `true`, "int64": `1`}[art]
}

// g3Verletzung ist eine Verletzung eines Schlüssels: der JSON-Wert (leer: fehlt) und der Ausgang.
type g3Verletzung struct {
	name    string
	json    string
	ursache model.DatenUrsache
	ort     bool // der Ort ist die Property (sonst leer)
}

func g3Verletzungen(art string) []g3Verletzung {
	out := []g3Verletzung{{"fehlt", "", model.DatenSchluessel, true}, {"null", `null`, model.DatenSchluessel, true}}
	switch art {
	case "string":
		out = append(out, g3Verletzung{"Marker", `"[REDACTED]"`, model.DatenSchluessel, true}, g3Verletzung{"U+0000", `"a\u0000b"`, model.DatenSchluessel, true},
			g3Verletzung{"falscher Typ", `5`, model.DatenTyp, true})
	case "int64":
		out = append(out, g3Verletzung{"Marker", `"[REDACTED]"`, model.DatenSchluessel, true}, g3Verletzung{"falscher Typ", `true`, model.DatenTyp, true},
			g3Verletzung{"ungültig", `1.5`, model.DatenWert, true})
	case "int32":
		out = append(out, g3Verletzung{"falscher Typ", `"x"`, model.DatenTyp, true}, g3Verletzung{"ungültig", `1.5`, model.DatenWert, true},
			g3Verletzung{"Marker (String)", `"[REDACTED]"`, model.DatenTyp, true})
	default:
		out = append(out, g3Verletzung{"falscher Typ", `"x"`, model.DatenTyp, true})
	}
	return out
}

// g3Element ist das Objekt des Setups; verletze ist der Index des Schlüssels mit der Verletzung (-1: keiner), doppelt
// nennt den Schlüssel des Index zweimal.
func g3Element(st schluesselSetup, verletze int, v g3Verletzung, doppelt bool) string {
	var ms []string
	if st.typ != "" {
		ms = append(ms, `"@odata.type":"#`+st.typ+`"`)
	}
	for i, k := range st.keys {
		w := gutWert(k.art)
		if i == verletze {
			w = v.json
		}
		if i == verletze && w == "" {
			continue
		}
		ms = append(ms, `"`+k.name+`":`+w)
		if i == verletze && doppelt {
			ms = append(ms, `"`+k.name+`":`+w)
		}
	}
	return "{" + strings.Join(ms, ",") + "}"
}

// g3Stellen setzt das Objekt an die Stelle der Entity: als Wurzel der Entity oder als erstes, mittleres oder letztes von drei Elementen.
func g3Stellen(name string, st schluesselSetup, gut, schlecht string, u model.DatenUrsache, ort string) []jfall {
	members := strings.TrimSuffix(strings.TrimPrefix(schlecht, "{"), "}")
	out := []jfall{{name: name + " Entity", m: st.m, set: st.set, einzel: true, body: jEin(st.set, members), ursache: u, ort: ort}}
	for i := range 3 {
		elems := []string{gut, gut, gut}
		elems[i] = schlecht
		o := ort
		if u != model.DatenDoppelterName {
			o = fmt.Sprintf("/value/%d%s", i, ort)
		}
		out = append(out, jfall{name: fmt.Sprintf("%s Element %d", name, i), m: st.m, set: st.set, body: jKol(st.set, elems...), ursache: u, ort: o})
	}
	return out
}

func jfG3() []jfall {
	var out []jfall
	for _, st := range g3Setups() {
		gut := g3Element(st, -1, g3Verletzung{}, false)
		for k, key := range st.keys {
			for _, v := range g3Verletzungen(key.art) {
				name := fmt.Sprintf("G3 %s %s %s", st.name, key.name, v.name)
				out = append(out, g3Stellen(name, st, gut, g3Element(st, k, v, false), v.ursache, "/"+key.name)...)
			}
			doppelt := g3Verletzung{"doppelt", gutWert(key.art), model.DatenDoppelterName, false}
			out = append(out, g3Stellen("doppelt "+st.name+" "+key.name, st, gut, g3Element(st, k, doppelt, true), model.DatenDoppelterName, "")...)
		}
		// Projektion ohne Schlüssel: der erste fehlende Teil in der Reihenfolge des Modells.
		ohne := g3Element(schluesselSetup{typ: st.typ}, -1, g3Verletzung{}, false)
		out = append(out, g3Stellen("G3 "+st.name+" alle fehlen", st, gut, ohne, model.DatenSchluessel, "/"+st.keys[0].name)...)
	}
	// Typ ohne Schlüssel: der abstrakte Typ des Sets, ohne @odata.type.
	abs := schluesselSetup{m: "T", set: "Abs"}
	out = append(out, g3Stellen("G3 Typ ohne Schlüssel", abs, `{"@odata.type":"#T.Der","Id":1}`, `{"Name":"x"}`, model.DatenSchluessel, "")...)
	return out
}

func jG3(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfG3()) }
