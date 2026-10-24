package outboundtest

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// roh bildet aus einem Body, dessen Verstoß im Lesen liegt und von der Methode unabhängig ist, einen Fall je Methode.
func roh(name string, body []byte, u model.DatenUrsache) []jfall {
	return []jfall{
		{name: name + " (Collection)", m: "R", set: "Products", body: body, ursache: u},
		{name: name + " (Entity)", m: "R", set: "Products", einzel: true, body: body, ursache: u},
	}
}

// rohAlle bildet aus Bodys, die syntaktisch ungültig sind, die Fälle beider Methoden.
func rohAlle(prefix string, bodys []string) []jfall {
	var out []jfall
	for _, b := range bodys {
		out = append(out, roh(fmt.Sprintf("%s %q", prefix, b), []byte(b), model.DatenSyntax)...)
	}
	return out
}

// praefixe sind alle echten Präfixe des Bodys (Länge 0 bis L−1).
func praefixe(body string) []string {
	out := make([]string, len(body))
	for i := range body {
		out[i] = body[:i]
	}
	return out
}

func jfB1() []jfall {
	menge := []string{"", " ", " \t\r\n", "hallo", "{", "[", `"`, `{"a`, `{"a"`, `{"a":`, `{"a":1`, `{"a":1,`, "tru", "nul", "-", "1e", "1."}
	// Das Ende des Stroms mitten in einem Escape, mitten in einem Surrogatpaar und mitten in einem mehrbyteigen Zeichen.
	for _, e := range []string{
		"\x5c", esc(""), esc("0"), esc("00"), esc("004"), esc("0041"), esc("d83d"), esc("d83d") + "\x5c", esc("d83d") + esc(""), esc("d83d") + esc("de"),
	} {
		menge = append(menge, `{"a":"`+e)
	}
	out := rohAlle("B1", menge)
	out = append(out, rohAlle("B1 Präfix E1", praefixe(jEinzelR))...)
	return append(out, rohAlle("B1 Präfix C1", praefixe(jKollR))...)
}

func jB1(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfB1()) }

// Schablonen für B2: die Marken @@1 bis @@6 stehen an den Orten, an denen ein Byte eingesetzt wird.
const (
	b2Kol = `@@1{@@3"@odata.context"@@5:@@6"$metadata#Products"@@4,@@3"value":[{"ID":1,"Name":"a"}]@@4}@@2`
	b2Ein = `@@1{@@3"@odata.context"@@5:@@6"$metadata#Products/$entity"@@4,@@3"ID"@@5:@@6 1@@4,@@3"Name":"a"@@4}@@2`
)

func b2Orte() []string {
	return []string{"vor der Wurzel", "nach der Wurzel", "nach dem Komma", "vor dem Komma", "zwischen Name und Doppelpunkt", "nach dem Doppelpunkt"}
}

// einsetzen setzt b an die Marke ort (1 bis 6); die anderen Marken entfallen.
func einsetzen(schablone string, ort int, b string) string {
	for i := 1; i <= 6; i++ {
		ersatz := ""
		if i == ort {
			ersatz = b
		}
		schablone = strings.ReplaceAll(schablone, "@@"+strconv.Itoa(i), ersatz)
	}
	return schablone
}

// jfB2Bytes: für jedes Byte 0x00 bis 0xFF an jedem Ort ist der Body genau für Leerzeichen, Tabulator, LF und CR gültig;
// in einem String ist jedes Byte unter 0x20 `syntax`, ab 0x80 `kodierung`, `"` und `\` brechen den String.
func jfB2Bytes() []jfall {
	var out []jfall
	for b := range 256 {
		ws := b == 0x20 || b == 0x09 || b == 0x0A || b == 0x0D
		for ort, name := range b2Orte() {
			for _, einzel := range []bool{false, true} {
				schablone := b2Kol
				if einzel {
					schablone = b2Ein
				}
				f := jfall{name: fmt.Sprintf("B2 Byte %#02x %s", b, name), m: "R", set: "Products", einzel: einzel,
					body: []byte(einsetzen(schablone, ort+1, string([]byte{byte(b)})))}
				if !ws {
					f.ursache = model.DatenSyntax
				}
				out = append(out, f)
			}
		}
		out = append(out, jfB2String(b)...)
	}
	return out
}

func jfB2String(b int) []jfall {
	u := model.DatenUrsache("")
	switch {
	case b < 0x20 || b == '"' || b == '\\':
		u = model.DatenSyntax
	case b >= 0x80:
		u = model.DatenKodierung
	}
	text := "a" + string([]byte{byte(b)}) + "c"
	members := `"ID":1,"Name":"` + text + `"`
	if u == "" {
		return beideGut(fmt.Sprintf("B2 Byte %#02x im String", b), "R", "Products", members,
			pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, text))
	}
	return beide(fmt.Sprintf("B2 Byte %#02x im String", b), "R", "Products", members, u, "")
}

func jfB2Verstoesse() []jfall {
	kol := func(frag string) string {
		return `{"@odata.context":"$metadata#Products","value":[` + frag + `]}`
	}
	var bodys []string
	for _, frag := range []string{
		"True", "FALSE", "Null", "NaN", "Infinity", "-Infinity", "01", "00", "-01", "1.", ".5", "+1", "-", "1e", "1e+", "0x10", "1e+-1",
		"1.,2", "-,2", "1e,2", "1e+,2", "1.]", "-]", "1e]", "0.e1", "1.e1", "[1.,2]", "tru,1", "nul]",
		`'a'`, `{a:1}`, `{'a':1}`, `/* x */ {}`, `// x` + "\n" + `{}`, `{"a" 1}`, `{"a":}`, `{"a"}`, `{"a":1 "b":2}`, `[1 2]`, `[1,,2]`, `{"a":1,}`,
		`[1,]`, `["\a"]`, `["\x41"]`, `["` + esc("12") + `"]`, `["` + esc("12G4") + `"]`, `["\U0041"]`, `["\`, `"` + esc("00e") + `"`, `{,}`, `[,]`,
		`{"a":1}}`, `]`, `{"a"::1}`,
	} {
		bodys = append(bodys, kol(frag))
	}
	bodys = append(bodys, kol("")+" {}", kol("")+"x", kol("")+",", kol("")+kol(""), "{}"+kol(""))
	return rohAlle("B2 Verstoß", bodys)
}

func jfB2Gueltig() []jfall {
	var out []jfall
	for _, e := range []struct{ roh, text string }{
		{`\"`, `"`}, {`\\`, `\`}, {`\/`, `/`}, {`\b`, "\b"}, {`\f`, "\f"}, {`\n`, "\n"}, {`\r`, "\r"}, {`\t`, "\t"},
		{"é", "é"}, {esc("00e9"), "é"}, {esc("00E9"), "é"}, {"\x7f", "\x7f"}, {" ", " "},
	} {
		out = append(out, beideGut("B2 gültig Escape "+e.roh, "R", "Products", `"ID":1,"Name":"`+e.roh+`"`,
			pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, e.text))...)
	}
	for _, z := range []struct{ roh, text string }{
		{"-0", "0"}, {"0", "0"}, {"-0.0", "0"}, {"1E5", "100000"}, {"1e+5", "100000"}, {"0.5e-3", "0.0005"}, {"12.50", "12.5"},
	} {
		out = append(out, beideGut("B2 gültig Zahl "+z.roh, "R", "Products", `"ID":1,"Price":`+z.roh,
			pw("ID", model.ArtInt32, "1"), pw("Price", model.ArtDecimal, z.text))...)
	}
	return out
}

func jfB2() []jfall {
	return verbinde(jfB2Bytes(), jfB2Verstoesse(), jfB2Gueltig())
}

func jB2(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfB2()) }

// b3Orte sind die Orte, an denen ein Byte oder eine Folge in Strings des Bodys steht: der Wert einer Property, der Name
// eines Mitglieds, der Wert einer Annotation, der Inhalt einer Navigation.
func b3Orte(seq string) map[string]string {
	return map[string]string{
		"Wert":       `"ID":1,"Name":"` + seq + `"`,
		"Name":       `"ID":1,"` + seq + `":1`,
		"Annotation": `"ID":1,"@x.y":"` + seq + `"`,
		"Navigation": `"ID":1,"Category":{"a":"` + seq + `"}`,
	}
}

func b3Ungueltig() []string {
	hex := func(b ...byte) string { return string(b) }
	return []string{
		hex(0xC0, 0x80), hex(0xC1, 0xBF), hex(0xE0, 0x80, 0x80), hex(0xED, 0xA0, 0x80), hex(0xED, 0xBF, 0xBF),
		hex(0xF0, 0x80, 0x80, 0x80), hex(0xF4, 0x90, 0x80, 0x80), hex(0xF5, 0x80, 0x80, 0x80), hex(0x80), hex(0xBF), hex(0xFF),
		hex(0xC3), hex(0xE4, 0xB8), hex(0xF0, 0x9F, 0x98), hex(0xC2, 0x41), hex(0xE0, 0xA0, 0x41), hex(0xF1, 0x80, 0x80, 0x41),
	}
}

// b3Gueltig sind Zeichen roh im String: die Grenzen der UTF-8-Folgen und der Bereich der Surrogate.
func b3Gueltig() []string {
	var out []string
	for _, r := range []rune{0x7f, 0x80, 0x7ff, 0x800, 0xd7ff, 0xe000, 0xfffd, 0xffff, 0x10000, 0x10ffff, 0xfeff} {
		out = append(out, string(r))
	}
	return out
}

func jfB3() []jfall {
	var out []jfall
	for _, bom := range []string{"\xEF\xBB\xBF", "\xFF\xFE", "\xFE\xFF"} {
		out = append(out, roh(fmt.Sprintf("B3 BOM %x vor E1", bom), []byte(bom+jEinzelR), model.DatenKodierung)...)
		out = append(out, roh(fmt.Sprintf("B3 BOM %x vor C1", bom), []byte(bom+jKollR), model.DatenKodierung)...)
		out = append(out, roh(fmt.Sprintf("B3 BOM %x allein", bom), []byte(bom), model.DatenKodierung)...)
	}
	for _, seq := range b3Ungueltig() {
		for ort, members := range b3Orte(seq) {
			out = append(out, beide(fmt.Sprintf("B3 %x im %s", seq, ort), "R", "Products", members, model.DatenKodierung, "")...)
		}
		out = append(out, roh(fmt.Sprintf("B3 %x im Kontext", seq), []byte(`{"@odata.context":"$metadata#Products`+seq+`","value":[]}`), model.DatenKodierung)...)
	}
	for _, seq := range []string{"\xE4\xB8", "\xC3", "\xF0\x9F\x98"} {
		out = append(out, roh(fmt.Sprintf("B3 %x am Ende des Stroms", seq), []byte(`{"@odata.context":"$metadata#Products","value":[{"Name":"`+seq), model.DatenSyntax)...)
	}
	for _, seq := range b3Gueltig() {
		out = append(out, beideGut(fmt.Sprintf("B3 gültig %U", []rune(seq)[0]), "R", "Products", b3Orte(seq)["Wert"],
			pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, seq))...)
	}
	return append(out, jfB3Escapes()...)
}

func jfB3Escapes() []jfall {
	var out []jfall
	for _, e := range []string{
		esc("D800"), esc("DBFF"), esc("DC00"), esc("DFFF"), esc("D800") + " ", esc("D800") + "A", esc("D800") + esc("D800"),
		esc("DC00") + esc("D800"), esc("D800") + `\n`, esc("D800") + esc("0041"), esc("DBFF") + esc("E000"), esc("D800") + esc("DBFF"),
		esc("DC00") + esc("DC00"), esc("D800") + `\`, esc("d800") + esc("d800"), esc("dbff") + esc("0041"), esc("D800") + esc("DC0"),
	} {
		u := model.DatenKodierung
		if strings.HasSuffix(e, esc("DC0")) {
			u = model.DatenSyntax
		}
		out = append(out, beide("B3 Escape "+e, "R", "Products", b3Orte(e)["Wert"], u, "")...)
	}
	for _, g := range []struct{ esc, text string }{
		{esc("d7ff"), string(rune(0xd7ff))}, {esc("e000"), string(rune(0xe000))}, {esc("d800") + esc("dc00"), string(rune(0x10000))},
		{esc("dbff") + esc("dfff"), string(rune(0x10ffff))}, {esc("d83d") + esc("de00"), string(rune(0x1f600))}, {esc("0000"), "\x00"},
		{esc("D83D") + esc("DE00"), string(rune(0x1f600))}, {"a" + esc("0000") + "b", "a\x00b"}, {esc("fffd"), string(rune(0xfffd))},
		{esc("FFFF"), string(rune(0xffff))},
	} {
		out = append(out, beideGut("B3 gültig Escape "+g.esc, "R", "Products", b3Orte(g.esc)["Wert"],
			pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, g.text))...)
	}
	return out
}

func jB3(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfB3()) }

// b4Wrapper sind die Orte, an denen B4 geschachtelte Klammern einsetzt; basis ist die Zahl der Klammern über dem Ort
// (Wurzel, `value`, Element).
type b4Wrapper struct {
	name   string
	m, set string
	vor    string
	nach   string
	basis  int
	soll   model.DatenUrsache // Ausgang bei der Tiefe 512
	ort    string
}

func b4Wrappers() []b4Wrapper {
	kol := func(vor, nach string) (string, string) {
		return `{"@odata.context":"$metadata#Products","value":[{"ID":1,` + vor, nach + `}]}`
	}
	nav1, nav2 := kol(`"Category":`, ``)
	weight1, weight2 := kol(`"Weight":`, ``)
	ann1, ann2 := kol(`"@x.y":`, ``)
	dyn1 := `{"@odata.context":"$metadata#Animals","value":[{"@odata.type":"#Zoo.Mammal","Id":1,"dyn":`
	return []b4Wrapper{
		{"Navigation", "R", "Products", nav1, nav2, 3, "", ""},
		{"nicht unterstützte Property", "R", "Products", weight1, weight2, 3, "", ""},
		{"Annotation", "R", "Products", ann1, ann2, 3, "", ""},
		{"dynamischer Member", "Z", "Animals", dyn1, `}]}`, 3, "", ""},
		{"unter einem Element von value", "R", "Products", `{"@odata.context":"$metadata#Products","value":[`, `]}`, 2, model.DatenStruktur, "/value/0"},
	}
}

// klammern liefert n ineinander liegende Klammern der Art (o Objekte, a Arrays, w abwechselnd); n ist die Tiefe der innersten.
func klammern(art byte, n int) string {
	var b strings.Builder
	var schluss []byte
	for i := 1; i <= n; i++ {
		obj := art == 'o' || (art == 'w' && i%2 == 1)
		switch {
		case obj && i < n:
			b.WriteString(`{"a":`)
			schluss = append(schluss, '}')
		case obj:
			b.WriteString(`{`)
			schluss = append(schluss, '}')
		default:
			b.WriteByte('[')
			schluss = append(schluss, ']')
		}
	}
	for i := len(schluss) - 1; i >= 0; i-- {
		b.WriteByte(schluss[i])
	}
	return b.String()
}

func jfB4() []jfall {
	var out []jfall
	for _, w := range b4Wrappers() {
		for _, art := range []byte{'o', 'a', 'w'} {
			for _, tiefe := range []int{512, 513} {
				f := jfall{name: fmt.Sprintf("B4 %s Klammern %c Tiefe %d", w.name, art, tiefe), m: w.m, set: w.set,
					body: []byte(w.vor + klammern(art, tiefe-w.basis) + w.nach)}
				switch {
				case tiefe == 513:
					f.ursache = model.DatenTiefe
				case w.soll != "" && art != 'a':
					f.ursache, f.ort = model.DatenUnbekannteProperty, "/value/0/a"
				default:
					f.ursache, f.ort = w.soll, w.ort
				}
				out = append(out, f)
			}
		}
	}
	flach := strings.Repeat("[],", 699) + "[]"
	out = append(out, jfall{name: "B4 700 flache Geschwister", m: "R", set: "Products",
		body: []byte(`{"@odata.context":"$metadata#Products","value":[{"ID":1,"Category":[` + flach + `]}]}`)})
	out = append(out, jfall{name: "B4 Wurzel allein Tiefe 513", m: "R", set: "Products", body: []byte(klammern('a', 513)), ursache: model.DatenTiefe})
	return out
}

func jB4(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, jfB4())
	// Wächter: 100 000 offene Klammern enden mit `tiefe`, ohne dass der Speicher wächst.
	tief := []byte(strings.Repeat("[", 100000))
	for _, einzel := range []bool{true, false} {
		f := jfall{name: "B4 100000 Ebenen", m: "R", set: "Products", einzel: einzel, body: tief, ursache: model.DatenTiefe}
		s.pruefe(t, f)
	}
}

// jfB5 sind die Bodys von B5: der Body mit Leerraum aufgefüllt, bei verschiedenen Werten von grenze.
func jfB5() []jfall {
	var out []jfall
	for _, einzel := range []bool{true, false} {
		basis := basisR(einzel)
		body := string(basis.body) + strings.Repeat(" ", 50)
		l := int64(len(body))
		for _, g := range []struct {
			name    string
			grenze  int64
			ursache model.DatenUrsache
		}{
			{"L-1", l - 1, model.DatenZuGross}, {"L", l, ""}, {"L+1", l + 1, ""}, {"MaxInt64", math.MaxInt64, ""},
		} {
			out = append(out, jfall{name: "B5 grenze " + g.name + " " + basis.name, m: "R", set: "Products", einzel: einzel,
				body: []byte(body), grenze: g.grenze, ursache: g.ursache})
		}
		out = append(out,
			jfall{name: "B5 grenze 1 und 2 Bytes " + basis.name, m: "R", set: "Products", einzel: einzel, body: []byte("{}"), grenze: 1, ursache: model.DatenZuGross},
			jfall{name: "B5 Verstoß vor dem Byte grenze+1 " + basis.name, m: "R", set: "Products", einzel: einzel,
				body: []byte(`{"a" 1` + strings.Repeat(" ", 10)), grenze: 6, ursache: model.DatenSyntax},
			jfall{name: "B5 gültiges Präfix, dahinter Müll " + basis.name, m: "R", set: "Products", einzel: einzel,
				body: []byte(string(basis.body) + "x"), grenze: int64(len(basis.body)), ursache: model.DatenZuGross},
			jfall{name: "B5 Verstoß hinter dem Byte grenze+1 " + basis.name, m: "R", set: "Products", einzel: einzel,
				body: []byte(string(basis.body) + "   x"), grenze: int64(len(basis.body)), ursache: model.DatenZuGross},
			jfall{name: "B5 gültiger Body, grenze gleich L " + basis.name, m: "R", set: "Products", einzel: einzel,
				body: basis.body, grenze: int64(len(basis.body))},
		)
	}
	return out
}

func jB5(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfB5()) }

// b6Paare sind die Schreibweisen zweier Mitglieder desselben Namens `a`: gleich, mit Escape, mit verschiedenen Escapes,
// mit gleichem und verschiedenem Wert, mit `null`, dreimal.
func b6Paare() []string {
	a := `"` + esc("0061") + `"`
	return []string{
		`"a":1,"a":1`, `"a":1,"a":2`, `"a":1,` + a + `:1`, a + `:1,"a":2`, a + `:1,` + a + `:1`, a + `:1,` + a + `:2`,
		`"a":null,"a":1`, `"a":1,"a":null`, `"a":{},"a":{}`, `"a":1,"b":2,"a":3`, `"a":1,"a":2,"a":3`, `"":1,"":2`,
		`"` + esc("00e9") + `":1,"é":2`, `"é":1,"é":2`,
	}
}

func jfB6() []jfall {
	var out []jfall
	orte := []struct {
		name, m, set string
		inhalt       bool
		kol          func(paar string) string
	}{
		{"Wurzel", "R", "Products", false, func(p string) string { return `{"@odata.context":"$metadata#Products","value":[],` + p + `}` }},
		{"erstes Element", "R", "Products", false, func(p string) string { return jKolS("Products", `{"ID":1,`+p+`}`, `{"ID":2}`) }},
		{"letztes Element", "R", "Products", false, func(p string) string { return jKolS("Products", `{"ID":2}`, `{"ID":1,`+p+`}`) }},
		{"Objekt einer Navigation", "R", "Products", true, func(p string) string { return jKolS("Products", `{"ID":1,"Category":{`+p+`}}`) }},
		{"Objekt in einer Annotation", "R", "Products", true, func(p string) string { return jKolS("Products", `{"ID":1,"@x.y":{`+p+`}}`) }},
		{"Objekt in einem dynamischen Member", "Z", "Animals", true, func(p string) string {
			return jKolS("Animals", `{"@odata.type":"#Zoo.Mammal","Id":1,"dyn":{`+p+`}}`)
		}},
		{"Objekt in einer nicht unterstützten Property", "R", "Products", true, func(p string) string { return jKolS("Products", `{"ID":1,"Weight":{`+p+`}}`) }},
		{"Objekt in einem Array in einem Objekt", "R", "Products", true, func(p string) string { return jKolS("Products", `{"ID":1,"Weight":[{`+p+`}]}`) }},
	}
	for _, o := range orte {
		for _, paar := range b6Paare() {
			out = append(out, jfall{name: "B6 " + o.name + " " + paar, m: o.m, set: o.set, body: []byte(o.kol(paar)), ursache: model.DatenDoppelterName})
		}
		if o.inhalt {
			out = append(out, jfall{name: "B6 gültig a und A " + o.name, m: o.m, set: o.set, body: []byte(o.kol(`"a":1,"A":2`))})
		}
	}
	out = append(out, jfB6Gross()...)
	for _, d := range []string{
		`{"@odata.context":"$metadata#Products","value":[],"value":[]}`,
		`{"@odata.context":"$metadata#Products","@odata.context":"$metadata#Products","value":[]}`,
		`{"@odata.context":"$metadata#Products","@odata.nextLink":"a","@odata.nextLink":"a","value":[]}`,
		`{"@odata.context":"$metadata#Products","value":[{"ID":1,"ID":1}]}`,
	} {
		out = append(out, roh("B6 "+d, []byte(d), model.DatenDoppelterName)...)
	}
	return out
}

// jfB6Gross sind Objekte mit vielen Namen, in denen ein Name doppelt vorkommt: an der Grenze, bis zu der der Leser die
// Namen der Reihe nach vergleicht, und darüber; dazu ein Objekt mit vielen verschiedenen Namen, das gültig ist.
func jfB6Gross() []jfall {
	var out []jfall
	namen := func(n int, dazu string) string {
		var ms []string
		for i := range n {
			ms = append(ms, fmt.Sprintf(`"n%d":%d`, i, i))
		}
		return strings.Join(append(ms, dazu), ",")
	}
	kol := func(inhalt string) []byte {
		return []byte(`{"@odata.context":"$metadata#Products","value":[{"ID":1,"Weight":{` + inhalt + `}}]}`)
	}
	for _, n := range []int{15, 16, 17, 18, 40, 1000} {
		for _, dup := range []string{`"n0":0`, fmt.Sprintf(`"n%d":0`, n-1), fmt.Sprintf(`"n%d":0`, n/2)} {
			out = append(out, jfall{name: fmt.Sprintf("B6 %d Namen, doppelt %s", n, dup), m: "R", set: "Products", body: kol(namen(n, dup)), ursache: model.DatenDoppelterName})
		}
		out = append(out, jfall{name: fmt.Sprintf("B6 %d verschiedene Namen", n), m: "R", set: "Products", body: kol(namen(n, `"n`+strconv.Itoa(n)+`":0`))})
	}
	return out
}

// jKolS ist jKol als Text.
func jKolS(set string, elems ...string) string { return string(jKol(set, elems...)) }

func jB6(t *testing.T, s *jsuite) { s.pruefeAlle(t, jfB6()) }
