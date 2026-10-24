package outboundtest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Generatoren der Suite fahren Mengen, die für eine Liste von Hand zu groß sind (SPEC-126 Nr. 3, 4, 14; Zeilen B2, B3
// und K1 der Tabelle). gFahre verteilt Pakete von Körpern auf bis zu acht Goroutinen, jeder Aufruf nimmt einen frischen
// Leser. Das Orakel einer Menge steht im Test und nie im Leser.

// gMeldungen ist die Zahl der Abweichungen einer Menge, die der Test einzeln meldet; die Zahl aller steht danach.
const gMeldungen = 10

// gPaket ist die Zahl der Körper, die eine Goroutine auf einmal holt.
const gPaket = 2048

// gArbeit ist ein Paket von Körpern mit der laufenden Nummer des ersten.
type gArbeit struct {
	erste   int64
	koerper [][]byte
}

// gPruefer prüft einen Körper (nr zählt von 0 in der Reihenfolge der Menge) und liefert die Klasse, in die er fällt, und bei
// einer Abweichung ihren Text.
type gPruefer func(b []byte, nr int64) (klasse, abweichung string)

// gLauf ist der Lauf einer Menge: die Zähler, die die Goroutinen gemeinsam führen.
type gLauf struct {
	t            *testing.T
	name         string
	pruefe       gPruefer
	mu           sync.Mutex
	klassen      map[string]int64
	abweichungen atomic.Int64
}

// arbeiter prüft die Pakete des Kanals und trägt seine Zählung am Ende ein.
func (l *gLauf) arbeiter(arbeit <-chan gArbeit) {
	lokal := map[string]int64{}
	for p := range arbeit {
		for i, b := range p.koerper {
			k, ab := l.pruefe(b, p.erste+int64(i))
			lokal[k]++
			if ab != "" && l.abweichungen.Add(1) <= gMeldungen {
				l.t.Errorf("%s: %s", l.name, ab)
			}
		}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range lokal {
		l.klassen[k] += v
	}
}

// gVerteile schneidet die Folge in Pakete und gibt sie in den Kanal.
func gVerteile(folge iter.Seq[[]byte], arbeit chan<- gArbeit) {
	var nr int64
	p := gArbeit{}
	for b := range folge {
		if len(p.koerper) == 0 {
			p.erste = nr
		}
		p.koerper = append(p.koerper, bytes.Clone(b))
		nr++
		if len(p.koerper) == gPaket {
			arbeit <- p
			p = gArbeit{}
		}
	}
	if len(p.koerper) > 0 {
		arbeit <- p
	}
}

// gFahre fährt die Körper der Folge gegen pruefe und liefert die Zahl der Körper je Klasse.
func gFahre(t *testing.T, name string, folge iter.Seq[[]byte], pruefe gPruefer) map[string]int64 {
	t.Helper()
	l := &gLauf{t: t, name: name, pruefe: pruefe, klassen: map[string]int64{}}
	arbeit := make(chan gArbeit, 16)
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), 8) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.arbeiter(arbeit)
		}()
	}
	gVerteile(folge, arbeit)
	close(arbeit)
	wg.Wait()
	if n := l.abweichungen.Load(); n > gMeldungen {
		t.Errorf("%s: %d Abweichungen, die ersten %d stehen oben", name, n, gMeldungen)
	}
	return l.klassen
}

// gZaehle prüft die Summe der Klassen gegen die erwartete Zahl der Körper und nennt beide im Log.
func gZaehle(t *testing.T, name string, klassen map[string]int64, erwartet int64) {
	t.Helper()
	var summe int64
	for _, v := range klassen {
		summe += v
	}
	t.Logf("%s: %d Körper %v", name, summe, klassen)
	if summe != erwartet {
		t.Errorf("%s: %d Körper, erwartet %d", name, summe, erwartet)
	}
}

// gFolge ist die Zahl der Folgen der Längen von bis bis über einem Alphabet von k Zeichen: die Summe von k^von bis k^bis.
func gFolge(k, von, bis int) int64 {
	var summe int64
	for n := von; n <= bis; n++ {
		potenz := int64(1)
		for range n {
			potenz *= int64(k)
		}
		summe += potenz
	}
	return summe
}

// gWeiter schaltet den Zähler idx (Ziffern zur Basis k) um eins weiter; falsch heißt: übergelaufen, alle Folgen sind durch.
func gWeiter(idx []int, k int) bool {
	for i := len(idx) - 1; i >= 0; i-- {
		if idx[i]++; idx[i] < k {
			return true
		}
		idx[i] = 0
	}
	return false
}

// gFolgen liefert alle Folgen der Länge von bis bis über dem Alphabet, jede mit vor am Anfang und nach am Ende.
func gFolgen(alphabet []string, von, bis int, vor, nach string) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		var b []byte
		for n := von; n <= bis; n++ {
			idx := make([]int, n)
			for ok := true; ok; ok = gWeiter(idx, len(alphabet)) {
				b = append(b[:0], vor...)
				for _, i := range idx {
					b = append(b, alphabet[i]...)
				}
				if !yield(append(b, nach...)) {
					return
				}
			}
		}
	}
}

// gTokens liefert die Spannen der Token eines Dokuments ohne Leerraum: ein Zeichen aus `{}[],:`, ein String samt Escapes oder
// eine Folge anderer Zeichen (Zahl, Literal).
func gTokens(doc string) [][2]int {
	var out [][2]int
	for i := 0; i < len(doc); {
		j := i + 1
		switch c := doc[i]; {
		case c == '"':
			for doc[j] != '"' {
				if doc[j] == '\\' {
					j++
				}
				j++
			}
			j++
		case !strings.ContainsRune("{}[],:", rune(c)):
			for j < len(doc) && !strings.ContainsRune("{}[],:\"", rune(doc[j])) {
				j++
			}
		}
		out = append(out, [2]int{i, j})
		i = j
	}
	return out
}

// gAenderung ersetzt doc[vor:ab] durch mitte.
type gAenderung struct {
	vor, ab int
	mitte   string
}

// gTokenAenderungen sind die Änderungen auf Token-Ebene: jedes Token gelöscht, durch jedes Token des Alphabets ersetzt, und jedes
// Token des Alphabets vor jedem Token und am Ende eingefügt.
func gTokenAenderungen(doc string, alphabet []string) []gAenderung {
	tokens := gTokens(doc)
	var out []gAenderung
	for _, t := range tokens {
		out = append(out, gAenderung{t[0], t[1], ""})
		for _, a := range alphabet {
			out = append(out, gAenderung{t[0], t[1], a})
		}
	}
	stellen := []int{len(doc)}
	for _, t := range tokens {
		stellen = append(stellen, t[0])
	}
	for _, p := range stellen {
		for _, a := range alphabet {
			out = append(out, gAenderung{p, p, a})
		}
	}
	return out
}

// gByteAenderungen sind die Änderungen auf Byte-Ebene: jedes Byte 0x00 bis 0xFF an jeder Stelle eingefügt und jede Stelle durch
// es ersetzt, dazu jedes Byte gelöscht.
func gByteAenderungen(doc string) []gAenderung {
	var out []gAenderung
	for p := 0; p <= len(doc); p++ {
		for c := range 256 {
			z := string([]byte{byte(c)})
			out = append(out, gAenderung{p, p, z})
			if p < len(doc) {
				out = append(out, gAenderung{p, p + 1, z})
			}
		}
		if p < len(doc) {
			out = append(out, gAenderung{p, p + 1, ""})
		}
	}
	return out
}

// gAenderungen liefert die Einzeländerungen des gültigen Dokuments (Menge G3), auf Token- und auf Byte-Ebene.
func gAenderungen(doc string, alphabet []string) iter.Seq[[]byte] {
	liste := append(gTokenAenderungen(doc, alphabet), gByteAenderungen(doc)...)
	return func(yield func([]byte) bool) {
		var b []byte
		for _, a := range liste {
			b = append(append(append(b[:0], doc[:a.vor]...), a.mitte...), doc[a.ab:]...)
			if !yield(b) {
				return
			}
		}
	}
}

// gAenderungenZahl ist die Zahl der Körper von gAenderungen, aus den Maßen des Dokuments gerechnet: T Löschungen, T × k
// Ersetzungen, (T + 1) × k Einschübe, (L + 1) × 256 eingefügte, L × 256 ersetzte und L gelöschte Bytes für T Token, ein Alphabet
// von k Token und ein Dokument von L Bytes.
func gAenderungenZahl(doc string, k int) int64 {
	t, l := int64(len(gTokens(doc))), int64(len(doc))
	return t + t*int64(k) + (t+1)*int64(k) + (l+1)*256 + l*256 + l
}

// gEscapes liefert die Menge G4: der Rest nach `\` in einem String, jedes Byte 0x00 bis 0x7F roh und jedes Zeichen U+0080 bis
// U+00FF (das einzelne Byte ab 0x80 ist ungültiges UTF-8 und gehört zu B3), und jedes Literal mit einer Änderung der Schreibweise
// an jeder Stelle (`tRue`).
func gEscapes() iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for c := range 256 {
			rest := string(rune(c))
			if c < 0x80 {
				rest = string([]byte{byte(c)})
			}
			if !yield([]byte(`["\` + rest + `"]`)) {
				return
			}
		}
		for _, lit := range []string{"true", "false", "null"} {
			for i := range lit {
				if !yield([]byte("[" + lit[:i] + strings.ToUpper(lit[i:i+1]) + lit[i+1:] + "]")) {
					return
				}
			}
		}
	}
}

// gEscapesZahl ist die Zahl der Körper von gEscapes: 256 Escape-Bytes und die 13 Stellen der drei Literale.
const gEscapesZahl = 256 + 4 + 5 + 4

// gSoll ist, was das Orakel von einem Körper der Grammatik erwartet.
type gSoll string

const (
	gSyntax      gSoll = "syntax"
	gNichtSyntax gSoll = "nicht syntax"
	gDoppelt     gSoll = "doppelter_name"
	gOhneUTF8    gSoll = "ausgenommen: ungültiges UTF-8"
)

// gOrakel ist das Orakel der Grammatik: encoding/json.Valid. Ein Körper mit ungültigem UTF-8 gehört zu B3 (die Bibliothek nimmt
// es an) und ist ausgenommen. Die Dokumente von G3 tragen keinen Escape eines Surrogats, eine Einzeländerung bildet keinen
// (die Bibliothek nähme einen einzelnen Surrogat-Escape an, der Leser lehnt ihn ab: B3). Ein Name, der vor der ersten
// Verletzung doppelt steht, ist `doppelter_name`.
func gOrakel(b []byte) gSoll {
	switch {
	case !utf8.Valid(b):
		return gOhneUTF8
	case json.Valid(b):
		return gNichtSyntax
	case gDubletteVorVerletzung(b):
		return gDoppelt
	}
	return gSyntax
}

// gRahmen ist ein offenes Objekt oder Array des Orakels der doppelten Namen; n zählt die Werte und Namen darin.
type gRahmen struct {
	objekt bool
	n      int
	namen  map[string]bool
}

// nimm trägt ein Token in das offene Objekt oder Array ein und meldet, ob es ein Name ist, der schon im Objekt steht.
func (f *gRahmen) nimm(tok json.Token) (dublette bool) {
	name, istName := tok.(string)
	if istName && f.objekt && f.n%2 == 0 {
		dublette = f.namen[name]
		f.namen[name] = true
	}
	f.n++
	return dublette
}

// gDubletteVorVerletzung meldet, ob ein Name im selben Objekt zweimal vollständig gelesen ist, bevor der Körper den ersten
// Verstoß enthält: der Dekodierer der Bibliothek liefert die Token bis zum ersten Fehler. Er hört mit dem ersten Wert auf.
func gDubletteVorVerletzung(b []byte) bool {
	if bytes.IndexByte(b, '{') < 0 {
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	var stapel []gRahmen
	for {
		tok, err := dec.Token()
		if err != nil {
			return false
		}
		d, istDelim := tok.(json.Delim)
		if istDelim && (d == '}' || d == ']') {
			if stapel = stapel[:len(stapel)-1]; len(stapel) == 0 {
				return false
			}
			continue
		}
		if len(stapel) > 0 && stapel[len(stapel)-1].nimm(tok) {
			return true
		}
		switch {
		case istDelim:
			stapel = append(stapel, gRahmen{objekt: d == '{', namen: map[string]bool{}})
		case len(stapel) == 0:
			return false
		}
	}
}

// gGrammatik prüft einen Körper der Menge G gegen das Orakel; die Methode wechselt mit der Nummer (das Lesen gilt für beide
// gleich, SPEC-126 Kopf).
func (s *jsuite) gGrammatik(b []byte, nr int64) (klasse, abweichung string) {
	soll := gOrakel(b)
	if soll == gOhneUTF8 {
		return string(soll), ""
	}
	r := s.aufruf(context.Background(), nr%2 == 0, bytes.NewReader(b), s.modelle["R"], "Products", 1<<30)
	var df *model.DatenFehler
	ist := errors.As(r.err, &df)
	syntax := ist && df.Ursache == model.DatenSyntax && df.Ort == ""
	doppelt := ist && df.Ursache == model.DatenDoppelterName
	switch {
	case r.err != nil && !ist:
		return string(soll), fmt.Sprintf("Körper %.300q: Ergebnis %v ist kein DatenFehler", b, r.err)
	case soll == gSyntax && !syntax, soll == gNichtSyntax && syntax:
		return string(soll), fmt.Sprintf("Körper %.300q: Ergebnis %v, erwartet %s", b, r.err, soll)
	case soll == gDoppelt && !doppelt:
		return string(soll), fmt.Sprintf("Körper %.300q: Ergebnis %v, erwartet doppelter_name vor der ersten Verletzung", b, r.err)
	}
	return string(soll), ""
}

// gDokumente sind die gültigen Dokumente, von denen G3 ausgeht: die Einzel-Entity E1, die Collection C1 und ein verschachteltes
// Dokument (Objekt in Array in Objekt) mit jeder Zahlform, den Literalen und jedem Escape.
func gDokumente() []string {
	verschachtelt := `{"a":[{"b":-1.5e+3,"c":[0,-0,12,1E5,0.5e-3],"d":true,"e":false,"f":null,` +
		`"g":"\"\\\/\b\f\n\r\t` + esc("00e9") + `","h":{}}],"i":[]}`
	return []string{jEinzelR, jKollR, verschachtelt}
}

// jB2Generatoren fährt die Mengen G1 bis G4 der Allaussage „Jeder Verstoß der Grammatik ist `syntax`“ (SPEC-126 Nr. 4): sonst
// ist ein Körper genau dann nicht `syntax`, wenn encoding/json.Valid ihn annimmt. Unter dem Detektor läuft die Probe (G1 bis
// Länge 4, G2 bis Länge 5, G3 nur an E1); `make test` fährt die ganze Menge. Die Länge der Folgen und die Zahl der Dokumente
// begrenzen die Menge (akzeptiertes Negativ im Plan, Zeile B2).
func jB2Generatoren(t *testing.T, s *jsuite) {
	g1, g2, docs := 6, 7, gDokumente()
	if unterDetektor {
		g1, g2, docs = 4, 5, docs[:1]
	}
	token, zahl := gAlphabetToken(), gAlphabetZahl()
	gZaehle(t, "B2 G1 Token-Folgen", gFahre(t, "B2 G1", gFolgen(token, 1, g1, "", ""), s.gGrammatik), gFolge(len(token), 1, g1))
	gZaehle(t, "B2 G2 Zahlfolgen", gFahre(t, "B2 G2", gFolgen(zahl, 1, g2, "[", "]"), s.gGrammatik), gFolge(len(zahl), 1, g2))
	for i, doc := range docs {
		name := fmt.Sprintf("B2 G3 Dokument %d (%d Bytes)", i+1, len(doc))
		gZaehle(t, name, gFahre(t, name, gAenderungen(doc, token), s.gGrammatik), gAenderungenZahl(doc, len(token)))
	}
	gZaehle(t, "B2 G4 Escapes und Literale", gFahre(t, "B2 G4", gEscapes(), s.gGrammatik), gEscapesZahl)
}

// gAlphabetToken sind die neun Token von G1 und der Ersatz und Einschub von G3.
func gAlphabetToken() []string {
	return []string{"{", "}", "[", "]", ",", ":", `"a"`, "1", "true"}
}

// gAlphabetZahl sind die sieben Zeichen von G2.
func gAlphabetZahl() []string { return []string{"0", "1", "-", "+", ".", "e", "E"} }
