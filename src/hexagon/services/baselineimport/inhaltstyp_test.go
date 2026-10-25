package baselineimport_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

const (
	typJSON    = "application/json"
	typXML     = "application/xml"
	typTextXML = "text/xml"
)

// randFolgen sind die sieben Folgen aus Leerzeichen und Tabulator der Länge 0 bis 2.
func randFolgen() []string {
	out := []string{""}
	for _, a := range []string{" ", "\t"} {
		out = append(out, a)
		for _, b := range []string{" ", "\t"} {
			out = append(out, a+b)
		}
	}
	return out
}

// schreibweisenContentType sind die 2048 Schreibweisen von `content-type`: jeder der 11 Buchstaben klein oder groß, der
// Bindestrich hat keine.
func schreibweisenContentType() []string {
	const basis = "content-type"
	var out []string
	for maske := 0; maske < 1<<12; maske++ {
		if maske>>strings.IndexByte(basis, '-')&1 == 1 {
			continue
		}
		b := []byte(basis)
		for i := range b {
			if maske>>i&1 == 1 {
				b[i] -= 'a' - 'A'
			}
		}
		out = append(out, string(b))
	}
	return out
}

// inhaltstypPruefer führt die Elemente von E3 gegen inhaltstypOK und zählt sie je Klasse.
type inhaltstypPruefer struct {
	t     *testing.T
	zaehl map[string]int
}

func kopfCT(v ...string) map[string][]string { return map[string][]string{"Content-Type": v} }

// pruefe wertet einen Header gegen den einen zulässigen Medientyp aus: ok heißt gelesen, ohne heißt ohne Kontext.
func (p *inhaltstypPruefer) pruefe(klasse string, h map[string][]string, medium string, ok, ohne bool) {
	p.t.Helper()
	p.zaehl[klasse]++
	if gotOK, gotOhne := baselineimport.Inhaltstyp(h, medium); gotOK != ok || gotOhne != ohne {
		p.t.Errorf("(%s) %q gegen %s: ok %v, ohne_kontext %v, erwartet %v und %v", klasse, h, medium, gotOK, gotOhne, ok, ohne)
	}
}

// nichtBefreit sind die 36 Zeichen, die an keinem Rand gekürzt werden: 0x00 bis 0x20 außer Leerzeichen und Tabulator (31),
// 0x7F, U+0085, U+00A0 (34; Zeichen, die ein Trim oder TrimSpace anders behandelt als Trim(" \t")) und die Anführungszeichen
// `"` und `'` (2). Nicht geführt sind die 17 weiteren Zeichen, die strings.TrimSpace kürzt (U+1680, U+2000 bis U+200A, U+2028,
// U+2029, U+202F, U+205F, U+3000); eine Änderung, die nur sie kürzt, bleibt grün (akzeptiertes Negativ im Plan, E3).
func nichtBefreit() []string {
	var out []string
	for c := 0; c <= 0x20; c++ {
		if c != ' ' && c != '\t' {
			out = append(out, string([]byte{byte(c)}))
		}
	}
	return append(out, "\x7f", string(rune(0x85)), string(rune(0xa0)), `"`, `'`)
}

// klasseRand: (a) der Rand des Medientyps, sieben Folgen links mal sieben rechts mal drei Medientypen, alle gelesen (147);
// (b) Zeichen am Rand, die nicht befreit werden (36), links und rechts, drei Medientypen, keines gelesen (216).
func (p *inhaltstypPruefer) klasseRand() {
	raender := randFolgen()
	if len(raender) != 7 {
		p.t.Fatalf("Folgen %d, erwartet 7", len(raender))
	}
	medien := []string{typJSON, typXML, typTextXML}
	for _, l := range raender {
		for _, r := range raender {
			for _, m := range medien {
				p.pruefe("a", kopfCT(l+m+r), m, true, false)
			}
		}
	}
	fremd := nichtBefreit()
	if len(fremd) != 36 {
		p.t.Fatalf("Zeichen %d, erwartet 36", len(fremd))
	}
	for _, c := range fremd {
		for _, m := range medien {
			p.pruefe("b", kopfCT(c+m), m, false, false)
			p.pruefe("b", kopfCT(m+c), m, false, false)
		}
	}
}

// klasseName: (c) der Name des Headers ohne Beachtung der ASCII-Schreibweise: jede der 2048 Schreibweisen allein ist gelesen, neben
// einer zweiten, abweichenden Schreibweise sind es zwei Werte; dazu zehn Namen, die nur ähneln, der Eintrag ohne Wert und
// drei Werte (4108).
func (p *inhaltstypPruefer) klasseName() {
	namen := schreibweisenContentType()
	verschieden := map[string]bool{}
	for _, n := range namen {
		verschieden[n] = true
	}
	if len(namen) != 2048 || len(verschieden) != 2048 {
		p.t.Fatalf("Schreibweisen %d, verschiedene %d, erwartet 2048", len(namen), len(verschieden))
	}
	for _, n := range namen {
		p.pruefe("c", map[string][]string{n: {typJSON}}, typJSON, true, false)
		andere := "Content-Type"
		if n == andere {
			andere = "content-type"
		}
		p.pruefe("c", map[string][]string{n: {typJSON}, andere: {typJSON}}, typJSON, false, false)
	}
	nbsp := string(rune(0xa0))
	aehnlich := []string{
		" content-type", "content-type ", "\tcontent-type", "content-type\t", nbsp + "content-type", "content-type" + nbsp,
		"content_type", "content-types", "xcontent-type", "content-typ",
	}
	if len(aehnlich) != 10 {
		p.t.Fatalf("ähnliche Namen %d, erwartet 10", len(aehnlich))
	}
	for _, n := range aehnlich {
		p.pruefe("c", map[string][]string{n: {typJSON}}, typJSON, false, false)
	}
	p.pruefe("c", map[string][]string{"Content-Type": nil}, typJSON, false, false)
	p.pruefe("c", kopfCT(typJSON, typJSON, typJSON), typJSON, false, false)
}

// klasseASCII: (d) ASCII-klein an den sieben Verwendungen einzeln in Großbuchstaben, dazu asciiKlein über alle 256 Bytes und die
// drei Zeichen, die unicode.ToLower oder EqualFold auf einen ASCII-Buchstaben abbilden (266).
func (p *inhaltstypPruefer) klasseASCII() {
	p.pruefe("d", kopfCT("APPLICATION/JSON"), typJSON, true, false)
	p.pruefe("d", kopfCT("APPLICATION/XML"), typXML, true, false)
	p.pruefe("d", kopfCT("TEXT/XML"), typTextXML, true, false)
	p.pruefe("d", kopfCT(typJSON+";CHARSET=utf-8"), typJSON, true, false)
	p.pruefe("d", kopfCT(typJSON+";charset=UTF-8"), typJSON, true, false)
	p.pruefe("d", kopfCT(typJSON+";ODATA.METADATA=none"), typJSON, true, true)
	p.pruefe("d", kopfCT(typJSON+";odata.metadata=NONE"), typJSON, true, true)
	for c := 0; c < 256; c++ {
		p.zaehl["d"]++
		want := string([]byte{byte(c)})
		if c >= 'A' && c <= 'Z' {
			want = string([]byte{byte(c + 'a' - 'A')})
		}
		if got := baselineimport.ASCIIKlein(string([]byte{byte(c)})); got != want {
			p.t.Errorf("ASCIIKlein(%#x) = %q, erwartet %q", c, got, want)
		}
	}
	for _, r := range []string{"\xe2\x84\xaa", "\xc5\xbf", "\xc4\xb0"} {
		p.zaehl["d"]++
		if got := baselineimport.ASCIIKlein(r); got != r {
			p.t.Errorf("ASCIIKlein(%q) = %q, erwartet unverändert", r, got)
		}
	}
}

// klasseLeerraum: (e) Leerraum um Name und Wert eines Parameters: die vier Orte mal die sieben Folgen (entdoppelt 25) mal vier
// Parameter (100).
func (p *inhaltstypPruefer) klasseLeerraum() {
	for _, par := range []struct {
		name, wert string
		ok, ohne   bool
	}{{"charset", "utf-8", true, false}, {"charset", "iso-8859-1", false, false}, {"odata.metadata", "none", true, true}, {"odata.metadata", "minimal", true, false}} {
		gesehen := map[string]bool{}
		for _, f := range randFolgen() {
			for _, k := range []string{
				";" + f + par.name + "=" + par.wert, ";" + par.name + f + "=" + par.wert, ";" + par.name + "=" + f + par.wert, ";" + par.name + "=" + par.wert + f,
			} {
				if !gesehen[k] {
					gesehen[k] = true
					p.pruefe("e", kopfCT(typJSON+k), typJSON, par.ok, par.ohne)
				}
			}
		}
		if len(gesehen) != 25 {
			p.t.Errorf("Formen von %s=%s: %d, erwartet 25", par.name, par.wert, len(gesehen))
		}
	}
}

// klasseParameter: (f) Leerraum und Anführungszeichen (4), (g) `;` in Anführungszeichen trennt (3), (h) Randformen der Parameter:
// leere Teile zählen nicht, ohne `=` ist der Wert leer, der Wert ist der Rest hinter dem ersten `=` (9), (i) Werte von charset neben
// utf-8, alle nicht gelesen (6).
func (p *inhaltstypPruefer) klasseParameter() {
	p.pruefe("f", kopfCT(typJSON+`;charset=" utf-8"`), typJSON, false, false)
	p.pruefe("f", kopfCT(typJSON+`;charset="utf-8 "`), typJSON, false, false)
	p.pruefe("f", kopfCT(typJSON+`;charset = "utf-8"`), typJSON, true, false)
	p.pruefe("f", kopfCT(typJSON+`;charset="utf-8" `), typJSON, true, false)

	p.pruefe("g", kopfCT(typJSON+`;x=";charset=iso-8859-1"`), typJSON, false, false)
	p.pruefe("g", kopfCT(typJSON+`;x=";charset=utf-8"`), typJSON, false, false)
	p.pruefe("g", kopfCT(typJSON+`;x=";odata.metadata=none;y="`), typJSON, true, true)

	p.pruefe("h", kopfCT(typJSON+";;"), typJSON, true, false)
	p.pruefe("h", kopfCT(typJSON+"; "), typJSON, true, false)
	p.pruefe("h", kopfCT(typJSON+";"), typJSON, true, false)
	p.pruefe("h", kopfCT(typJSON+";charset"), typJSON, false, false)
	p.pruefe("h", kopfCT(typJSON+";charset="), typJSON, false, false)
	p.pruefe("h", kopfCT(typJSON+";charset=utf-8=x"), typJSON, false, false)
	p.pruefe("h", kopfCT(typJSON+";odata.metadata"), typJSON, true, false)
	p.pruefe("h", kopfCT(typJSON+";odata.metadata="), typJSON, true, false)
	p.pruefe("h", kopfCT(typJSON+";=utf-8"), typJSON, true, false)

	for _, c := range []string{"utf8", "utf_8", "utf-16", "utf-8x", "xutf-8", "us-ascii"} {
		p.pruefe("i", kopfCT(typJSON+";charset="+c), typJSON, false, false)
	}
}

// klasseParameterRand: (k) die 36 nicht befreiten Zeichen links und rechts von Name und Wert eines Parameters, mal die zwei
// Parameter (288): am Namen ist der Name ein anderer und der Parameter ohne Wirkung, die Fixture trägt den Wert, bei dem er die
// Antwort abwiese oder ohne Kontext ließe (`charset` mit iso-8859-1, `odata.metadata` mit none), erwartet ist gelesen; am Wert
// ist der Wert ein anderer, die Fixture trägt den Wert, der bei Befreiung gelesen würde (`charset=utf-8`: nicht gelesen,
// `odata.metadata=none`: gelesen ohne ohne_kontext); (l) Namen, die nur ähneln: ein `x` davor, ein `x` dahinter, ein Zeichen
// weniger am Ende, für die zwei Parameter mit ihren wirksamen Werten, alle gelesen (6); (m) Anführungszeichen ohne Paar mit
// Anhängsel: vier Formen je Parameter (8).
func (p *inhaltstypPruefer) klasseParameterRand() {
	fremd := nichtBefreit()
	if len(fremd) != 36 {
		p.t.Fatalf("Zeichen %d, erwartet 36", len(fremd))
	}
	for _, par := range []struct {
		name, amName, amWert string
		okAmWert             bool
	}{{"charset", "iso-8859-1", "utf-8", false}, {"odata.metadata", "none", "none", true}} {
		for _, c := range fremd {
			p.pruefe("k", kopfCT(typJSON+";"+c+par.name+"="+par.amName), typJSON, true, false)
			p.pruefe("k", kopfCT(typJSON+";"+par.name+c+"="+par.amName), typJSON, true, false)
			p.pruefe("k", kopfCT(typJSON+";"+par.name+"="+c+par.amWert), typJSON, par.okAmWert, false)
			p.pruefe("k", kopfCT(typJSON+";"+par.name+"="+par.amWert+c), typJSON, par.okAmWert, false)
		}
	}
	for _, par := range []struct{ name, wert string }{{"charset", "iso-8859-1"}, {"odata.metadata", "none"}} {
		for _, n := range []string{"x" + par.name, par.name + "x", par.name[:len(par.name)-1]} {
			p.pruefe("l", kopfCT(typJSON+";"+n+"="+par.wert), typJSON, true, false)
		}
	}
	for _, w := range []string{`"utf-8x`, `xutf-8"`, `"utf-8x"`, `x"utf-8"`} {
		p.pruefe("m", kopfCT(typJSON+";charset="+w), typJSON, false, false)
	}
	for _, w := range []string{`"nonex`, `xnone"`, `"nonex"`, `x"none"`} {
		p.pruefe("m", kopfCT(typJSON+";odata.metadata="+w), typJSON, true, false)
	}
}

// klasseMetadaten: (j) bei Metadaten hat odata.metadata=none keine Wirkung, das Dokument bleibt der Kandidat (1).
func (p *inhaltstypPruefer) klasseMetadaten() {
	p.zaehl["j"]++
	w := neu(p.t)
	w.sitzung().metadaten("<edmx/>", ct("application/xml;odata.metadata=none"))
	if _, err := w.importiere(anfrage()); err != nil {
		p.t.Errorf("(j) Metadaten mit odata.metadata=none: %v", err)
	}
}

// TestZeileE3InhaltstypKlassen: der Wortlaut von SPEC-127 Nr. 8 g in allen seinen Teilen als Schleifen über die reine Funktion
// inhaltstypOK, zusammen 5162 Elemente: (a) Rand des Medientyps 147, (b) Zeichen am Rand, die nicht befreit werden 216, (c) Name
// des Headers 4108, (d) ASCII-klein 266, (e) Leerraum um Name und Wert eines Parameters 100, (f) Leerraum und Anführungszeichen 4,
// (g) `;` in Anführungszeichen 3, (h) Randformen der Parameter 9, (i) Werte von `charset` 6, (j) Metadaten mit
// odata.metadata=none 1, (k) nicht befreite Zeichen am Rand von Name und Wert eines Parameters 288, (l) Namen, die nur ähneln 6,
// (m) Anführungszeichen ohne Paar mit Anhängsel 8. Jedes Element nennt, ob die Antwort gelesen wird (ok) und ob sie keinen Kontext trägt (ohne_kontext); ein
// nicht gelesener Inhalt trägt nie ohne_kontext.
func TestZeileE3InhaltstypKlassen(t *testing.T) {
	p := &inhaltstypPruefer{t: t, zaehl: map[string]int{}}
	p.klasseRand()
	p.klasseName()
	p.klasseASCII()
	p.klasseLeerraum()
	p.klasseParameter()
	p.klasseParameterRand()
	p.klasseMetadaten()
	soll := map[string]int{"a": 147, "b": 216, "c": 4108, "d": 266, "e": 100, "f": 4, "g": 3, "h": 9, "i": 6, "j": 1, "k": 288, "l": 6, "m": 8}
	gesamt := 0
	for _, n := range p.zaehl {
		gesamt += n
	}
	if fmt.Sprint(p.zaehl) != fmt.Sprint(soll) || gesamt != 5162 {
		t.Errorf("Mengen %v, erwartet %v (gesamt 5162)", p.zaehl, soll)
	}
}

// TestZeileD1FormatWerte: der Wert von `$format` im $metadata ist beliebig (SPEC-127 Nr. 7 a): sechs Klassen {leer, ohne `=`, `json`,
// `xml`, `application%2Fxml`, `=` im Wert} mal drei Zahlen der Vorkommen {1, 2 gleiche, 2 verschiedene} (18), jeder ein Kandidat.
func TestZeileD1FormatWerte(t *testing.T) {
	klassen := []string{"$format=", "$format", "$format=json", "$format=xml", "$format=application%2Fxml", "$format=a=b"}
	n := 0
	for i, seg := range klassen {
		for _, vorkommen := range []string{"einmal", "gleich", "verschieden"} {
			n++
			q := map[string]string{"einmal": seg, "gleich": seg + "&" + seg, "verschieden": seg + "&" + klassen[(i+1)%len(klassen)]}[vorkommen]
			w := neu(t)
			w.sitzung().metadaten("<edmx/>", func(e *model.Exchange) { e.RawQuery = q })
			if _, err := w.importiere(anfrage()); err != nil {
				t.Errorf("Query %q: kein Kandidat, %v", q, err)
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge %d, erwartet 18", n)
	}
}
