package outboundtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Zeilen G9 bis G14 prüfen den Abschnitt Lesen (SPEC-126 Nr. 2 bis 5, 13 a mit SPEC-172 Nr. 1): er kennt das Modell nicht,
// seine Ursachen hängen von der Hülle des Dokuments nicht ab.

// v2RohAlle bildet aus Bodys, die syntaktisch ungültig sind, die Fälle beider Methoden.
func v2RohAlle(prefix string, bodys []string) []jfall {
	var out []jfall
	for _, b := range bodys {
		out = append(out, v2Roh(fmt.Sprintf("%s %q", prefix, b), b, model.DatenSyntax, "")...)
	}
	return out
}

// Schablonen für die Bytes der Grammatik: die Marken @@1 bis @@6 stehen an den Orten, an denen ein Byte eingesetzt wird.
const (
	v2SchabloneKol = `@@1{@@3"d"@@5:@@6{@@3"results"@@5:@@6[{"ID":1,"Name":"a"}]@@4}@@4}@@2`
	v2SchabloneEin = `@@1{@@3"d"@@5:@@6{@@3"ID"@@5:@@6 1@@4,@@3"Name":"a"@@4}@@4}@@2`
)

// v2FaelleG9Bytes: für jedes Byte 0x00 bis 0xFF an jedem Ort ist der Body genau für Leerzeichen, Tabulator, LF und CR gültig;
// in einem String ist jedes Byte unter 0x20 `syntax`, ab 0x80 `kodierung`, `"` und `\` brechen den String.
func v2FaelleG9Bytes() []jfall {
	var out []jfall
	gut := []model.PropertyWert{pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, "a")}
	for b := range 256 {
		ws := b == 0x20 || b == 0x09 || b == 0x0A || b == 0x0D
		for ort, name := range b2Orte() {
			for _, einzel := range []bool{false, true} {
				schablone := v2SchabloneKol
				if einzel {
					schablone = v2SchabloneEin
				}
				f := jfall{name: fmt.Sprintf("G9 Byte %#02x %s", b, name), m: "R", set: "Products", einzel: einzel,
					body: []byte(einsetzen(schablone, ort+1, string([]byte{byte(b)})))}
				if ws {
					f.props = gut
				} else {
					f.ursache = model.DatenSyntax
				}
				out = append(out, f)
			}
		}
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
			out = append(out, v2BeideGut(fmt.Sprintf("G9 Byte %#02x im String", b), "R", "Products", members,
				pw("ID", model.ArtInt32, "1"), pw("Name", model.ArtString, text))...)
		} else {
			out = append(out, v2Beide(fmt.Sprintf("G9 Byte %#02x im String", b), "R", "Products", members, u, "")...)
		}
	}
	return out
}

func v2FaelleG9() []jfall {
	menge := []string{"", " ", " \t\r\n", "hallo", "{", "[", `"`, `{"d`, `{"d"`, `{"d":`, `{"d":{`, `{"d":{"results":[`, `{"d":1`, `{"d":1,`, "tru", "nul", "-", "1e", "1."}
	out := v2RohAlle("G9", menge)
	out = append(out, v2RohAlle("G9 Präfix E1", praefixe(v2EinR))...)
	out = append(out, v2RohAlle("G9 Präfix C1", praefixe(v2KollR))...)
	out = append(out, v2RohAlle("G9 E", b1EKoerper())...)
	out = append(out, jfB2Verstoesse()...)
	return append(out, v2FaelleG9Bytes()...)
}

// v2G9Dokumente sind die gültigen Dokumente, von denen die Änderungen ausgehen: die Einzel-Entity, die Collection und ein
// verschachteltes Dokument mit jeder Zahlform, den Literalen und jedem Escape.
func v2G9Dokumente() []string { return []string{v2EinR, v2KollR, gDokumente()[2]} }

// v2G9: genau ein JSON-Wert nach RFC 8259 streng; jeder Verstoß ist `syntax`, sonst ist ein Körper genau dann nicht `syntax`,
// wenn encoding/json.Valid ihn annimmt (SPEC-126 Nr. 4). Die Generatoren G1 (Token-Folgen), G2 (Zahl-Folgen), G3 (Einzeländerungen
// der Dokumente) und G4 (Escapes und Literale) stehen eingebettet in `{"d":[ … ]}`; unter dem Detektor läuft die Probe, `make test`
// fährt die ganze Menge. Die Länge der Folgen und die Zahl der Dokumente begrenzen die Menge (akzeptiertes Negativ im Plan, G9).
func v2G9(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleG9())
	g1, g2, docs := 6, 7, v2G9Dokumente()
	if unterDetektor {
		g1, g2, docs = 4, 5, docs[:1]
	}
	token, zahl := gAlphabetToken(), gAlphabetZahl()
	gZaehle(t, "G9 G1 Token-Folgen", gFahre(t, "G9 G1", gFolgen(token, 1, g1, `{"d":[`, `]}`), s.gGrammatik), gFolge(len(token), 1, g1))
	gZaehle(t, "G9 G2 Zahlfolgen", gFahre(t, "G9 G2", gFolgen(zahl, 1, g2, `{"d":[`, `]}`), s.gGrammatik), gFolge(len(zahl), 1, g2))
	for i, doc := range docs {
		name := fmt.Sprintf("G9 G3 Dokument %d (%d Bytes)", i+1, len(doc))
		gZaehle(t, name, gFahre(t, name, gAenderungen(doc, token), s.gGrammatik), gAenderungenZahl(doc, len(token)))
	}
	gZaehle(t, "G9 G4 Escapes und Literale", gFahre(t, "G9 G4", gEscapes(), s.gGrammatik), gEscapesZahl)
}

// v2OrteKodierung sind die Orte, an denen eine Folge in Strings des Dokuments steht: der Wert einer Property, der Name eines
// Mitglieds, der Inhalt einer Navigation, der Wert einer nicht unterstützten Property und der Inhalt von `__metadata`.
func v2OrteKodierung(seq string) map[string]string {
	return map[string]string{
		"Wert":         `"Id":1,"RString":"` + seq + `"`,
		"Name":         `"Id":1,"` + seq + `":1`,
		"Navigation":   `"Id":1,"Nav":{"a":"` + seq + `"}`,
		"übersprungen": `"Id":1,"X1":"` + seq + `"`,
		"__metadata":   `"Id":1,"__metadata":{"uri":"` + seq + `"}`,
	}
}

func v2FaelleG10() []jfall {
	var out []jfall
	for _, bom := range []string{"\xEF\xBB\xBF", "\xFF\xFE", "\xFE\xFF"} {
		out = append(out, v2Roh(fmt.Sprintf("G10 BOM %x vor E1", bom), bom+v2EinR, model.DatenKodierung, "")...)
		out = append(out, v2Roh(fmt.Sprintf("G10 BOM %x vor C1", bom), bom+v2KollR, model.DatenKodierung, "")...)
		out = append(out, v2Roh(fmt.Sprintf("G10 BOM %x allein", bom), bom, model.DatenKodierung, "")...)
	}
	for _, seq := range b3Ungueltig() {
		for ort, members := range v2OrteKodierung(seq) {
			out = append(out, v2Beide(fmt.Sprintf("G10 %x im %s", seq, ort), "W", "Ws", members, model.DatenKodierung, "")...)
		}
		out = append(out, v2Roh(fmt.Sprintf("G10 %x hinter der Hülle", seq), `{"d":{"results":[]},"`+seq+`":1}`, model.DatenKodierung, "")...)
	}
	for _, seq := range []string{"\xE4\xB8", "\xC3", "\xF0\x9F\x98"} {
		out = append(out, v2Roh(fmt.Sprintf("G10 %x am Ende des Stroms", seq), `{"d":{"results":[{"Name":"`+seq, model.DatenSyntax, "")...)
	}
	for _, seq := range append(b3Gueltig(), b3GZ()...) {
		out = append(out, v2BeideGut(fmt.Sprintf("G10 gültig %x", seq), "W", "Ws", v2OrteKodierung(seq)["Wert"],
			pw("Id", model.ArtInt32, "1"), pw("RString", model.ArtString, seq))...)
	}
	for _, e := range []string{
		esc("D800"), esc("DBFF"), esc("DC00"), esc("DFFF"), esc("D800") + " ", esc("D800") + "A", esc("D800") + esc("D800"),
		esc("DC00") + esc("D800"), esc("D800") + `\n`, esc("D800") + esc("0041"), esc("DBFF") + esc("E000"), esc("D800") + esc("DBFF"),
		esc("DC00") + esc("DC00"), esc("D800") + `\`, esc("d800") + esc("d800"), esc("dbff") + esc("0041"), esc("D800") + esc("DC0"),
	} {
		u := model.DatenKodierung
		if strings.HasSuffix(e, esc("DC0")) {
			u = model.DatenSyntax
		}
		for ort, members := range v2OrteKodierung(e) {
			out = append(out, v2Beide("G10 Escape "+e+" im "+ort, "W", "Ws", members, u, "")...)
		}
	}
	for _, g := range []struct{ esc, text string }{
		{esc("d7ff"), string(rune(0xd7ff))}, {esc("e000"), string(rune(0xe000))}, {esc("d800") + esc("dc00"), string(rune(0x10000))},
		{esc("dbff") + esc("dfff"), string(rune(0x10ffff))}, {esc("d83d") + esc("de00"), string(rune(0x1f600))}, {esc("0000"), "\x00"},
		{esc("D83D") + esc("DE00"), string(rune(0x1f600))}, {"a" + esc("0000") + "b", "a\x00b"}, {esc("fffd"), string(rune(0xfffd))},
		{esc("FFFF"), string(rune(0xffff))}, {esc("feff"), string(rune(0xfeff))}, {"\x7f", "\x7f"},
	} {
		out = append(out, v2BeideGut("G10 gültig Escape "+g.esc, "W", "Ws", v2OrteKodierung(g.esc)["Wert"],
			pw("Id", model.ArtInt32, "1"), pw("RString", model.ArtString, g.text))...)
	}
	return append(out, b3A2()...)
}

// Kopf und Fuß des Bodys der Menge U: die Folge steht roh als Wert der Property Name eines Elements.
const (
	v2UKopf = `{"d":{"results":[{"ID":1,"Name":"`
	v2UFuss = `"}]}}`
)

// v2UPruefer prüft einen Body der Menge U gegen das Orakel utf8: gültig heißt kein Fehler und der Text gleich der Folge, sonst die
// Ursache des Orakels (b3Orakel).
func (s *jsuite) v2UPruefer(b []byte, _ int64) (klasse, abweichung string) {
	seq := b[len(v2UKopf) : len(b)-len(v2UFuss)]
	soll := b3Orakel(seq)
	r := s.aufruf(context.Background(), false, bytes.NewReader(b), s.modelle["R"], "Products", 1<<30)
	klasse = "gültig"
	if soll != "" {
		klasse = string(soll)
	}
	var df *model.DatenFehler
	switch {
	case soll == "" && r.err == nil && len(r.seite.Entities) == 1 && len(r.seite.Entities[0].Properties) == 2 &&
		r.seite.Entities[0].Properties[1].Wert.Text == string(seq):
		return klasse, ""
	case soll != "" && errors.As(r.err, &df) && df.Ursache == soll && df.Ort == "":
		return klasse, ""
	}
	return klasse, fmt.Sprintf("Folge %x: Ergebnis %v %+v, erwartet %s", seq, r.err, r.seite.Entities, klasse)
}

// v2LeadSteuer liefert die Menge U3: jedes Lead-Byte 0xC2 bis 0xF4 mit jedem Steuerzeichen 0x00 bis 0x1F dahinter.
func v2LeadSteuer() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for lead := 0xC2; lead <= 0xF4; lead++ {
			for c := range 0x20 {
				if !yield([]byte(v2UKopf + string([]byte{byte(lead), byte(c)}) + v2UFuss)) {
					return
				}
			}
		}
	}
}

// v2G10: BOM, UTF-8 streng, Surrogat-Escapes paarweise, U+0000 gültig, U+FEFF und DEL im String Zeichen (SPEC-126 Nr. 3). Die
// Mengen: das teilweise BOM (A2), die gültigen Zeichen roh an den Grenzen der Lead-Bytes (GZ) und die Einzelverstöße der
// UTF-8-Folgen (U); Orakel ist utf8. Unter dem Detektor läuft von U die Probe.
func v2G10(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleG10())
	if n := len(b3A2()); n != 2*b3A2Probe {
		t.Errorf("G10 A2: %d Fälle, erwartet %d (%d Körper je Methode)", n, 2*b3A2Probe, b3A2Probe)
	}
	if !unterDetektor {
		klassen := gFahre(t, "G10 A2 hinten", b3A2Hinten(), s.b3A2Pruefer)
		gZaehle(t, "G10 A2 hinten", klassen, 4*255*256)
	}
	if n := len(b3GZ()); n != b3GZZahl {
		t.Errorf("G10 GZ: %d Folgen, erwartet %d", n, b3GZZahl)
	}
	grenzBytes := []string{"\x20", "\x7F", "\x80", "\x8F", "\x90", "\x9F", "\xA0", "\xBF", "\xC0", "\xC1", "\xC2", "\xDF", "\xE0", "\xE1", "\xEC", "\xED",
		"\xEE", "\xEF", "\xF0", "\xF1", "\xF3", "\xF4", "\xF5", "\xFF"}
	drei, vier := grenzBytes, 4
	if unterDetektor {
		drei, vier = []string{"\x7F", "\x80", "\xBF", "\xC2", "\xE0", "\xED", "\xF0", "\xF4"}, 3
	}
	alle := b3Alphabet(0x00, 0xFF)
	zwei := b3Alphabet(0x20, 0xFF, '"', '\\')
	gZaehle(t, "G10 U Ein-Byte-Folgen", gFahre(t, "G10 U1", gFolgen(alle, 1, 1, v2UKopf, v2UFuss), s.v2UPruefer), 256)
	gZaehle(t, "G10 U Zwei-Byte-Folgen", gFahre(t, "G10 U2", gFolgen(zwei, 2, 2, v2UKopf, v2UFuss), s.v2UPruefer), 222*222)
	gZaehle(t, "G10 U Lead-Byte und Steuerzeichen", gFahre(t, "G10 U3", v2LeadSteuer(), s.v2UPruefer), 51*32)
	gZaehle(t, "G10 U Drei- und Vier-Byte-Folgen", gFahre(t, "G10 U4", gFolgen(drei, 3, vier, v2UKopf, v2UFuss), s.v2UPruefer), gFolge(len(drei), 3, vier))
}

// v2Tiefenform ist eine Form des Bodys: die Zahl der Klammern, die vor dem Mitglied einer Entity offen sind (basis), und Kopf
// und Fuß um die Mitglieder.
type v2Tiefenform struct {
	name, kopf, fuss string
	basis            int
	einzel           bool
}

func v2Tiefenformen() []v2Tiefenform {
	return []v2Tiefenform{
		{"Collection", `{"d":{"results":[{"Id":1,`, `}]}}`, 4, false},
		{"Array", `{"d":[{"Id":1,`, `}]}`, 3, false},
		{"Entity", `{"d":{"Id":1,`, `}}`, 2, true},
	}
}

// v2Tiefenort ist ein Ort, an dem Klammern eingesetzt werden: der Inhalt, der nicht gedeutet wird. extra ist die Zahl der
// Klammern, die der Ort selbst öffnet.
type v2Tiefenort struct {
	name, vor, nach string
	extra           int
}

func v2Tiefenorte() []v2Tiefenort {
	return []v2Tiefenort{
		{"komplexe Property", `"X5":`, ``, 0}, {"__deferred", `"Nav":{"__deferred":`, `}`, 1},
		{"__metadata", `"__metadata":{"x":`, `}`, 1}, {"Navigation", `"Nav":`, ``, 0},
	}
}

func v2FaelleG11() []jfall {
	var out []jfall
	for _, form := range v2Tiefenformen() {
		for _, ort := range v2Tiefenorte() {
			for _, art := range []byte{'o', 'a', 'w'} {
				for _, tiefe := range []int{512, 513} {
					f := jfall{name: fmt.Sprintf("G11 %s %s Klammern %c Tiefe %d", form.name, ort.name, art, tiefe), m: "W", set: "Ws", einzel: form.einzel,
						body: []byte(form.kopf + ort.vor + klammern(art, tiefe-form.basis-ort.extra) + ort.nach + form.fuss)}
					if tiefe == 513 {
						f.ursache = model.DatenTiefe
					}
					out = append(out, f)
				}
			}
		}
	}
	flach := strings.Repeat("[],", 699) + "[]"
	out = append(out, jfall{name: "G11 700 flache Geschwister", m: "W", set: "Ws", body: v2Kol(`{"Id":1,"Nav":[` + flach + `]}`)})
	for _, einzel := range []bool{false, true} {
		out = append(out, jfall{name: "G11 Wurzel allein Tiefe 513", m: "R", set: "Products", einzel: einzel, body: []byte(klammern('a', 513)), ursache: model.DatenTiefe})
	}
	return out
}

func v2G11(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleG11())
	// 100 000 offene Klammern enden mit `tiefe`, ohne dass der Speicher wächst.
	tief := []byte(strings.Repeat("[", 100000))
	for _, einzel := range []bool{true, false} {
		s.pruefe(t, jfall{name: "G11 100000 Ebenen", m: "R", set: "Products", einzel: einzel, body: tief, ursache: model.DatenTiefe})
	}
}

// v2Orte13 sind die Objekte, in denen ein Name doppelt stehen kann; inhalt ist wahr, wo ein gültiges Paar `a` und `A` den Body
// gültig lässt.
func v2Orte13() []struct {
	name   string
	inhalt bool
	kol    func(paar string) string
} {
	return []struct {
		name   string
		inhalt bool
		kol    func(paar string) string
	}{
		{"Wurzel", false, func(p string) string { return `{"d":{"results":[]},` + p + `}` }},
		{"d", false, func(p string) string { return `{"d":{"results":[],` + p + `}}` }},
		{"Element von results", false, func(p string) string { return `{"d":{"results":[{"ID":1,` + p + `}]}}` }},
		{"__metadata", false, func(p string) string { return `{"d":{"ID":1,"__metadata":{` + p + `}}}` }},
		{"__deferred", true, func(p string) string { return `{"d":{"results":[{"ID":1,"Nav":{"__deferred":{` + p + `}}}]}}` }},
		{"Navigation", true, func(p string) string { return `{"d":{"results":[{"ID":1,"Nav":{` + p + `}}]}}` }},
		{"komplexe Property", true, func(p string) string { return `{"d":{"results":[{"ID":1,"X5":{` + p + `}}]}}` }},
		{"unbekannter Inhalt", false, func(p string) string { return `{"d":{"results":[{"ID":1,"zzz":{` + p + `}}]}}` }},
		{"Objekt im Array im Objekt", true, func(p string) string { return `{"d":{"results":[{"ID":1,"Nav":[{` + p + `}]}]}}` }},
	}
}

func v2FaelleG13() []jfall {
	var out []jfall
	for _, o := range v2Orte13() {
		for _, paar := range b6Paare() {
			for _, einzel := range []bool{false, true} {
				out = append(out, jfall{name: "G13 " + o.name + " " + paar, m: "W", set: "Ws", einzel: einzel, body: []byte(o.kol(paar)), ursache: model.DatenDoppelterName})
			}
		}
		if o.inhalt {
			out = append(out, jfall{name: "G13 gültig a und A " + o.name, m: "W", set: "Ws", body: []byte(strings.Replace(o.kol(`"a":1,"A":2`), `"ID"`, `"Id"`, 1))})
		}
	}
	namen := func(n int, dazu string) string {
		var ms []string
		for i := range n {
			ms = append(ms, fmt.Sprintf(`"n%d":%d`, i, i))
		}
		return strings.Join(append(ms, dazu), ",")
	}
	for _, n := range []int{15, 16, 17, 18, 40, 1000} {
		for _, dup := range []string{`"n0":0`, fmt.Sprintf(`"n%d":0`, n-1), fmt.Sprintf(`"n%d":0`, n/2)} {
			out = append(out, jfall{name: fmt.Sprintf("G13 %d Namen, doppelt %s", n, dup), m: "W", set: "Ws",
				body: v2Kol(`{"Id":1,"Nav":{` + namen(n, dup) + `}}`), ursache: model.DatenDoppelterName})
		}
		out = append(out, jfall{name: fmt.Sprintf("G13 %d verschiedene Namen", n), m: "W", set: "Ws",
			body: v2Kol(`{"Id":1,"Nav":{` + namen(n, `"n`+strconv.Itoa(n)+`":0`) + `}}`)})
	}
	for _, d := range []string{
		`{"d":{"results":[]},"d":{"results":[]}}`, `{"d":{"results":[],"results":[]}}`, `{"d":{"__count":"1","__count":"1","results":[]}}`,
		`{"d":{"__next":"a","__next":"a","results":[]}}`, `{"d":{"results":[{"ID":1,"ID":1}]}}`, `{"d":[{"ID":1,"ID":1}]}`,
	} {
		out = append(out, v2Roh("G13 "+d, d, model.DatenDoppelterName, "")...)
	}
	return out
}

func v2G13(t *testing.T, s *jsuite) { s.pruefeAlle(t, v2FaelleG13()) }

// v2FaelleG14 sind die Paare der Zusatztabelle Z2 (Rangfolge im Lesen, SPEC-126 Nr. 13 a): der erste Verstoß in der Reihenfolge
// der Bytes entscheidet, das Lesen geht dem Deuten vor.
func v2FaelleG14() []jfall {
	tief := strings.Repeat("[", 520)
	var out []jfall
	paar := func(name, body string, u model.DatenUrsache) {
		out = append(out, v2Roh("G14 "+name, body, u, "")...)
	}
	paar("doppelter Name früher als syntax", `{"d":{"ID":1,"a":1,"a":2,"b":}}`, model.DatenDoppelterName)
	paar("syntax früher als doppelter Name", `{"d":{"ID":1,"b":,"a":1,"a":2}}`, model.DatenSyntax)
	paar("kodierung früher als tiefe", "{\"d\":{\"ID\":1,\"X0\":[\"\xff\","+tief+"}}", model.DatenKodierung)
	paar("tiefe früher als kodierung", `{"d":{"ID":1,"X0":`+tief+"\"\xff\"}}", model.DatenTiefe)
	paar("syntax am Ende gegen struktur am Anfang", `{"x":1,"d":{"results":[]}`, model.DatenSyntax)
	paar("doppelter Name in einer Navigation gegen frühere unbekannte Property", `{"d":{"ID":1,"zzz":1,"Nav":{"a":1,"a":2}}}`, model.DatenDoppelterName)
	paar("Müll hinter der Wurzel gegen ungültiges UTF-8 dort", v2EinR+"\xff", model.DatenSyntax)
	paar("Müll hinter der Wurzel gegen tiefe dort", v2EinR+tief, model.DatenSyntax)
	paar("Müll hinter der Wurzel gegen doppelten Namen dort", v2EinR+`{"a":1,"a":2}`, model.DatenSyntax)
	for _, f := range v2Verstoesse() {
		out = append(out, v2Roh("G14 Verstoß "+f.name+" vor dem Ende der Wurzel", f.body, f.ursache, "")...)
	}
	return out
}

// v2G14: jedes Paar der Zusatztabelle Z2 in beiden Methoden; dazu die Paare, in denen die Größe entscheidet (G12), und der Fehler
// von src im selben Read wie ein Verstoß (G6).
func v2G14(t *testing.T, s *jsuite) {
	s.pruefeAlle(t, v2FaelleG14())
	for _, einzel := range []bool{true, false} {
		f := jfall{m: "R", set: "Products", einzel: einzel}
		// Fehler von src im selben Read gegen Verstoß des Lesens in den Daten: der Verstoß geht vor; bei unvollständigem Dokument
		// ist der Fehler von src der Fehler.
		rd := &teilReader{teile: [][]byte{[]byte(`{"d" 1`)}, ende: errQuelle, mitEnde: true}
		datenFehler(t, methodeName(einzel)+" Verstoß im selben Read wie der Fehler von src", s.ruf(context.Background(), rd, f).err, model.DatenSyntax, "")
		rd = &teilReader{teile: [][]byte{[]byte(`{"d":`)}, ende: errQuelle, mitEnde: true}
		if err := s.ruf(context.Background(), rd, f).err; !errors.Is(err, errQuelle) {
			t.Errorf("%s unvollständiges Dokument und Fehler von src: %v", methodeName(einzel), err)
		}
	}
}
