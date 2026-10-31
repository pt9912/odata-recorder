package urlparser_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const laeufe = 2000

// TestGeneratorPfad (Zeile 21): Umkehrung für den Pfad. 4 Arten der Ressource × 2000 Läufe, Saat 1 bis 2000: der AST zuerst,
// die Eingabe in zufälliger Schreibweise, das Orakel ist net/url.PathUnescape je Segment.
func TestGeneratorPfad(t *testing.T) {
	n := 0
	for _, art := range alleRessourcen() {
		for saat := uint64(1); saat <= laeufe; saat++ {
			r := rng(saat)
			pfad, want, segmente := genPfad(r, art)
			if len(pfad) > 2048 {
				t.Fatalf("Art %d Saat %d: der erzeugte Pfad ist %d Bytes lang", art, saat, len(pfad))
			}
			pruefeAbstieg(t, saat, pfad, segmente, "", nil)
			got, err := parse(pfad, "")
			if err != nil || !gleich(got, model.ODataURL{Ressource: want}) {
				t.Fatalf("Art %d Saat %d: %q\n  ist  %+v (%v)\n  soll %+v", art, saat, pfad, got, err, want)
			}
			n++
		}
	}
	if n != 4*laeufe {
		t.Fatalf("%d Läufe, erwartet %d", n, 4*laeufe)
	}
}

// TestGleicherWert (Zeile 39, zugleich der Generator der Klasse Literale, Zeilen 32 bis 40): je Art 2000 Läufe, Saat 1 bis
// 2000. Der Wert zuerst, Orakel sind math/big, time und strings; drei Schreibweisen ergeben denselben Text, ein anderer Wert
// einen anderen.
func TestGleicherWert(t *testing.T) {
	n := 0
	defer func() { zaehlpruefung(t, "TestGleicherWert", n, 6*laeufe) }()
	for _, art := range alleArten() {
		for saat := uint64(1); saat <= laeufe; saat++ {
			r := rng(saat)
			w := genLiteral(r, art)
			andere := genLiteral(rng(saat+100000), art)
			var erster string
			for i := range 3 {
				name := pick(r, "", "Id")
				k := w.roh(r)
				if name != "" {
					k = name + "=" + k
				}
				got, err := parse("/S("+kodiere(r, k, pfadZwingend)+")", "")
				if err != nil {
					t.Fatalf("Art %d Saat %d: %q: %v", art, saat, k, err)
				}
				l := got.Ressource.Schluessel[0]
				if l.Name != name || l.Literal.Art != art || l.Literal.Text != w.text {
					t.Fatalf("Art %d Saat %d: %q: %+v, erwartet %q", art, saat, k, l, w.text)
				}
				if i == 0 {
					erster = l.Literal.Text
				}
			}
			if (andere.text == w.text) != (erster == andere.text) {
				t.Fatalf("Art %d Saat %d: %q gegen %q", art, saat, w.text, andere.text)
			}
			n++
		}
	}
}

// vollKodiert schreibt jedes Byte von s als `%XX`.
func vollKodiert(r *rand.Rand, s string) string {
	return kodiere(r, s, func(byte) bool { return true })
}

// nurZwingend schreibt s mit den Escapes, die die Stelle verlangt, sonst roh.
func nurZwingend(s string, zwingend func(byte) bool) string {
	var b strings.Builder
	for i := range len(s) {
		if zwingend(s[i]) {
			fmt.Fprintf(&b, "%%%02X", s[i])
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// TestGeneratorKodierung (Zeilen 22 bis 26): ein beliebiges Zeichen an jeder der sechs Stellen, roh und mit jedem Byte als
// `%XX` (Hex-Ziffern zufällig): derselbe AST. Orakel ist net/url. 6 Stellen × 2000 Läufe, Saat 1 bis 2000.
func TestGeneratorKodierung(t *testing.T) {
	asciiName := []rune("abcXYZ019_-.")
	buchstabe := func(r *rand.Rand) rune { return buchstaben()[r.IntN(len(buchstaben()))] }
	stellen := []struct {
		name     string
		baue     func(e string) (pfad, query string)
		zwingend func(byte) bool
		zeichen  func(r *rand.Rand) rune
		pruefe   func(u model.ODataURL, text string) bool
	}{
		{"Name des Sets", func(e string) (string, string) { return "/S" + e, "" }, pfadZwingend, buchstabe,
			func(u model.ODataURL, c string) bool { return u.Ressource.EntitySet == "S"+c }},
		{"Name eines Schlüsselteils", func(e string) (string, string) { return "/S(K" + e + "=1)", "" }, pfadZwingend, buchstabe,
			func(u model.ODataURL, c string) bool { return u.Ressource.Schluessel[0].Name == "K"+c }},
		{"Schlüsselliteral", func(e string) (string, string) { return "/S('" + e + "')", "" }, pfadZwingend, zufallsZeichen,
			func(u model.ODataURL, c string) bool { return u.Ressource.Schluessel[0].Literal.Text == "'"+c+"'" }},
		{"Rest-Segment", func(e string) (string, string) { return "/S/R" + e, "" }, pfadZwingend, buchstabe, nil},
		{"Name einer Option", func(e string) (string, string) { return "/S", "x" + e + "=1" }, nameZwingend,
			func(r *rand.Rand) rune { return pick(r, asciiName...) },
			func(u model.ODataURL, c string) bool { return u.Optionen.Sonstige[0].Name == "x"+c }},
		{"Wert einer Option", func(e string) (string, string) { return "/S", "x=" + e }, wertZwingend, zufallsZeichen,
			func(u model.ODataURL, c string) bool { return u.Optionen.Sonstige[0].Wert == c }},
	}
	n := 0
	defer func() { zaehlpruefung(t, "TestGeneratorKodierung", n, len(stellen)*laeufe) }()
	for i, st := range stellen {
		for saat := uint64(1); saat <= laeufe; saat++ {
			r := rng(saat)
			c := string(st.zeichen(r))
			if i == 2 && c == "'" {
				c = "''" // im Literal steht ein Apostroph doppelt
			}
			ap, aq := st.baue(nurZwingend(c, st.zwingend))
			bp, bq := st.baue(vollKodiert(r, c))
			a, ea := parse(ap, aq)
			b, eb := parse(bp, bq)
			if st.pruefe == nil { // Rest-Segment: nicht unterstützt, beide gleich
				if ea == nil || eb == nil || ea.Error() != eb.Error() || !strings.Contains(ea.Error(), "pfad_nicht_unterstuetzt") {
					t.Fatalf("%s Saat %d Zeichen %q: roh %v, kodiert %v", st.name, saat, c, ea, eb)
				}
			} else if ea != nil || eb != nil || !st.pruefe(a, c) || !gleich(a, b) {
				t.Fatalf("%s Saat %d Zeichen %q: %q %q gegen %q %q: %+v (%v), %+v (%v)", st.name, saat, c, ap, aq, bp, bq, a, ea, b, eb)
			}
			if got, err := url.PathUnescape(vollKodiert(r, c)); err != nil || got != c {
				t.Fatalf("Saat %d: PathUnescape = %q (%v)", saat, got, err)
			}
			n++
		}
	}
}

// TestGeneratorQuery (Zeile 56): Umkehrung für die Query. 6 Klassen × 2000 Läufe, Saat 1 bis 2000, zufällige Reihenfolge,
// Escapes, führende Nullen, leere Segmente; Orakel sind QueryUnescape (nach Ersetzen von `+` durch `%2B`) und strings.Split.
// Eine siebte Klasse „alle“ mischt die sechs in einer Query.
func TestGeneratorQuery(t *testing.T) {
	n, klassen := 0, append(optionKlassen(), "alle")
	defer func() { zaehlpruefung(t, "TestGeneratorQuery", n, len(klassen)*laeufe) }()
	for _, klasse := range klassen {
		for saat := uint64(1); saat <= laeufe; saat++ {
			r := rng(saat)
			raw, want, plain := genQuery(r, genOptionen(r, klasse))
			pruefeAbstieg(t, saat, "", nil, raw, plain)
			got, err := parse("/S", raw)
			if err != nil || !gleich(got, model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMenge, EntitySet: "S"}, Optionen: want}) {
				t.Fatalf("Klasse %s Saat %d: %q\n  ist  %+v (%v)\n  soll %+v", klasse, saat, raw, got.Optionen, err, want)
			}
			n++
		}
	}
}

// rbasis ist die Klartext-Form einer gültigen URL, an der ein Mutator genau einen Verstoß herstellt.
type rbasis struct {
	name   string
	teile  []string
	paare  []optionsPaar
	suffix []string // Klartext-Segmente hinter dem Segment der Klammer
	wurzel bool     // der Pfad ist `/`
	roh    string   // Zugabe an die Query ohne Schreibweise (kaputte Kodierung)
	// Die Mutatoren von `kodierung` und `key` wählen nach der Saat die Stelle: ungerade im Pfad bzw. mit offener Klammer.
	ungerade   bool
	rohSegment string // Zugabe an den Pfad ohne Schreibweise (kaputte Kodierung)
	offen      bool   // die Klammer des Schlüssels ist nicht geschlossen
}

func (b *rbasis) gesetzt(name, wert string) {
	b.paare = append(slices.DeleteFunc(slices.Clone(b.paare), func(p optionsPaar) bool { return p.name == name }), einfach(name, wert))
}

func einfach(name, wert string) optionsPaar {
	return optionsPaar{name: name, wert: wert, setze: func(*model.Optionen) {}}
}

// anhaengen hängt ein Literal an den Körper: benannt als neuer Teil, bei einem unbenannten Körper ersetzt es ihn.
func anhaengen(teile []string, literal string) []string {
	if strings.Contains(teile[0], "=") && !strings.HasPrefix(teile[0], "'") {
		return append(teile, "Zq9="+literal)
	}
	return []string{literal}
}

// mutatoren stellt je Ursache genau einen Verstoß her.
func mutatoren() map[model.URLUrsache]func(*rbasis) {
	return map[model.URLUrsache]func(*rbasis){
		model.URLZuLang: func(b *rbasis) { b.paare = append(b.paare, einfach("x", wiederhole("a", 8200))) },
		model.URLZuViele: func(b *rbasis) {
			for i := range 65 {
				b.paare = append(b.paare, einfach(fmt.Sprintf("v%d", i), "1"))
			}
		},
		model.URLKodierung: func(b *rbasis) {
			if b.ungerade {
				b.rohSegment = "%G1"
			} else {
				b.roh = "k=%G1"
			}
		},
		model.URLSegment:  func(b *rbasis) { b.name += "-" },
		model.URLPfadForm: func(b *rbasis) { b.suffix = []string{"$count"} },
		model.URLKey: func(b *rbasis) {
			if b.ungerade {
				b.offen = true
			} else {
				b.teile = append(b.teile, "1")
			}
		},
		model.URLLiteral:    func(b *rbasis) { b.teile = anhaengen(b.teile, "1x") },
		model.URLOptionName: func(b *rbasis) { b.gesetzt("$zzz", "1") },
		model.URLOptionDoppelt: func(b *rbasis) {
			b.gesetzt("$top", "1")
			b.paare = append(b.paare, einfach("$top", "2"))
		},
		model.URLWert:                  func(b *rbasis) { b.gesetzt("$skip", "-1") },
		model.URLServiceDokument:       func(b *rbasis) { b.wurzel = true },
		model.URLPfadNichtUnterstuetzt: func(b *rbasis) { b.suffix = []string{"Rest"} },
		model.URLLiteralTyp:            func(b *rbasis) { b.teile = anhaengen(b.teile, "X'AA'") },
		model.URLOptionNichtUnterstuetzt: func(b *rbasis) {
			b.gesetzt("$expand", "1")
		},
		model.URLFilter:                func(b *rbasis) { b.gesetzt("$filter", "a") },
		model.URLWertNichtUnterstuetzt: func(b *rbasis) { b.gesetzt("$select", "a/b") },
	}
}

func basisAus(r *rand.Rand) *rbasis {
	b := &rbasis{name: genName(r, 12)}
	b.teile, _ = genKey(r, 4, alleArten())
	for _, k := range optionKlassen() {
		if r.IntN(2) == 0 {
			b.paare = append(b.paare, genOptionen(r, k)...)
		}
	}
	return b
}

func (b *rbasis) pfad(r *rand.Rand) string {
	if b.wurzel {
		return "/"
	}
	zu := ")"
	if b.offen {
		zu = ""
	}
	segs := append([]string{b.name + "(" + strings.Join(b.teile, ",") + zu}, b.suffix...)
	for i, s := range segs {
		segs[i] = kodiere(r, s, pfadZwingend)
	}
	if b.rohSegment != "" {
		segs = append(segs, b.rohSegment)
	}
	return "/" + strings.Join(segs, "/")
}

// ortGueltig prüft die Form des Orts: `pfad`, `pfad/<n>`, `query` oder `query/<n>`, <n> dezimal ab 1 ohne führende Null.
func ortGueltig(ort string) bool {
	for _, wurzel := range []string{"pfad", "query"} {
		rest, ok := strings.CutPrefix(ort, wurzel)
		if !ok {
			continue
		}
		if rest == "" {
			return true
		}
		n, hatSchraeg := strings.CutPrefix(rest, "/")
		v, err := strconv.Atoi(n)
		return hatSchraeg && err == nil && v >= 1 && n == strconv.Itoa(v)
	}
	return false
}

// pruefeFehlerform hält Ort und Text des Fehlers gegen die feste Form (SPEC-160 Nr. 4): der Ort ist `pfad`, `pfad/<n>`, `query`
// oder `query/<n>`, der Text `ODR-<Code>: OData-URL <ungültig|nicht unterstützt>: <Ursache> "<Ort>"`; ein Teil der Eingabe
// (Name, Wert, Literal, Rest eines Segments) steht darin nie.
func pruefeFehlerform(t *testing.T, err error) {
	t.Helper()
	var uf *model.URLFehler
	if !errors.As(err, &uf) {
		t.Fatalf("kein URLFehler: %v", err)
	}
	kopf := "ODR-E6005: OData-URL ungültig: "
	if klasse501(uf.Ursache) {
		kopf = "ODR-E6001: OData-URL nicht unterstützt: "
	}
	if !ortGueltig(uf.Ort) || err.Error() != kopf+string(uf.Ursache)+` "`+uf.Ort+`"` {
		t.Fatalf("Form des Fehlers: Ort %q, Text %q", uf.Ort, err.Error())
	}
}

func (b *rbasis) query(r *rand.Rand) string {
	var teile []string
	for _, p := range b.paare {
		teile = append(teile, kodiere(r, p.name, nameZwingend)+"="+kodiere(r, p.wert, wertZwingend))
	}
	if b.roh != "" {
		teile = append(teile, b.roh)
	}
	return strings.Join(teile, "&")
}

// TestGeneratorVerstoss (Zeile 57): Rückweg. Eine gültige URL mit genau einem Verstoß ergibt genau dessen Ursache und Klasse.
// 16 Ursachen × 500 Läufe, Saat 1 bis 500.
func TestGeneratorVerstoss(t *testing.T) {
	muts := mutatoren()
	if len(muts) != 16 {
		t.Fatalf("%d Mutatoren", len(muts))
	}
	n := 0
	defer func() { zaehlpruefung(t, "TestGeneratorVerstoss", n, 16*500) }()
	for _, u := range model.AlleURLUrsachen() {
		for saat := uint64(1); saat <= 500; saat++ {
			r := rng(saat)
			b := basisAus(r)
			if _, err := parse(b.pfad(rng(saat)), b.query(rng(saat))); err != nil {
				t.Fatalf("Saat %d: Basis ungültig: %v", saat, err)
			}
			b.ungerade = saat%2 == 1
			muts[u](b)
			p, q := b.pfad(rng(saat)), b.query(rng(saat))
			_, err := parse(p, q)
			var uf *model.URLFehler
			if !errors.As(err, &uf) || uf.Ursache != u {
				t.Fatalf("Ursache %s Saat %d: %.80q %.80q: %v", u, saat, p, q, err)
			}
			pruefeFehlerform(t, err)
			if errors.Is(err, model.ErrURLNichtUnterstuetzt) != klasse501(u) {
				t.Fatalf("Ursache %s: Klasse", u)
			}
			n++
		}
	}
}

// zaehlpruefung hält die Zahl der Läufe, die ein Generator gefahren hat, gegen die der Zusage; nach einem Fatalf ist der Test
// ohnehin rot.
func zaehlpruefung(t *testing.T, name string, ist, soll int) {
	t.Helper()
	if !t.Failed() && ist != soll {
		t.Errorf("%s: %d Läufe, erwartet %d", name, ist, soll)
	}
}
