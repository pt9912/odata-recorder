package model_test

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Bytes ab 0x80 entstehen aus Bytewerten, nie aus einer Escape-Folge im Quelltext (AGENTS.md §3.18).
func nbspBytes() string { return string([]byte{0xC2, 0xA0}) }
func a0Byte() string    { return string([]byte{0xA0}) }
func langesS() string   { return string([]byte{0xC5, 0xBF}) }

const (
	keine   = model.RueckgabeKeine
	minimal = model.RueckgabeMinimal
	entity  = model.RueckgabeEntity
)

func (z preferFall) String() string { return fmt.Sprintf("%q", z.zeilen) }

type preferFall struct {
	zeilen []string
	soll   model.Rueckgabe
}

// TestLiesRueckgabeBeispiele (Z1, SPEC-285 Nr. 4): jede Zeile der Tabelle der Beispiele, in ihrer Folge.
func TestLiesRueckgabeBeispiele(t *testing.T) {
	nbspBytes, a0Byte, langesS := nbspBytes(), a0Byte(), langesS()
	z := func(soll model.Rueckgabe, zeilen ...string) preferFall { return preferFall{zeilen, soll} }
	faelle := []preferFall{
		z(keine), z(keine, ""), z(keine, "respond-async"), z(keine, "odata.maxpagesize=5"),
		z(minimal, "return=minimal"), z(minimal, "Return=Minimal"), z(minimal, "RETURN=MINIMAL"), z(minimal, "return = minimal"),
		z(minimal, " \t return=minimal \t "), z(minimal, "return=\"minimal\""), z(minimal, "return=minimal; foo=bar"),
		z(minimal, "respond-async, return=minimal"),
		z(entity, "return=representation"), z(entity, "return=representation;"), z(entity, "return=representation; ="),
		z(entity, "return=\"Representation\""),
		z(minimal, "return=minimal, return=representation"), z(minimal, "return=minimal", "return=representation"),
		z(keine, "return=foo, return=minimal"), z(keine, "return, return=minimal"), z(keine, "return=, return=minimal"),
		z(keine, "return=\"\", return=minimal"), z(keine, "return=minimal x, return=representation"),
		z(keine, "return=min imal"), z(keine, "return=minimal=x"), z(keine, "return=\"minimal"), z(keine, "return=minimal\""),
		z(minimal, "return@x, return=minimal"), z(minimal, ", ,return=minimal"), z(entity, "x=1 2, return=representation"),
		z(keine, "x=\"a, return=representation\""),
		z(keine, nbspBytes+"return=minimal"),
		// nicht ASCII: weder `ſ` (U+017F) noch das Byte 0xA0 sind Teil eines Wortes oder Leerraum
		z(keine, "return=repre"+langesS+"entation"), z(keine, "return=\"repre"+langesS+"entation\""), z(keine, a0Byte+"return=minimal"),
		z(keine, "return=minimal"+a0Byte),
	}
	if len(faelle) != 36 {
		t.Fatalf("%d Fälle, erwartet 36", len(faelle))
	}
	for _, f := range faelle {
		if got := model.LiesRueckgabe(f.zeilen); got != f.soll {
			t.Errorf("%v: %d, erwartet %d", f, got, f.soll)
		}
	}
}

// schreibweisen liefert alle 2^n Schreibweisen des Wortes (jeder Buchstabe klein oder groß).
func schreibweisen(wort string) []string {
	var r []string
	for m := 0; m < 1<<len(wort); m++ {
		b := []byte(wort)
		for i := range b {
			if m&(1<<i) != 0 {
				b[i] -= 'a' - 'A'
			}
		}
		r = append(r, string(b))
	}
	return r
}

// TestLiesRueckgabeSchreibweise (Z2, SPEC-285 Nr. 3): Name und Wert gelten ohne Beachtung der ASCII-Schreibweise, jede Schreibweise von
// `return` (64), von `minimal` (128) und von `representation` (16384).
func TestLiesRueckgabeSchreibweise(t *testing.T) {
	n := 0
	for _, name := range schreibweisen("return") {
		for wert, soll := range map[string]model.Rueckgabe{"minimal": minimal, "representation": entity} {
			n++
			if got := model.LiesRueckgabe([]string{name + "=" + wert}); got != soll {
				t.Errorf("%s=%s: %d", name, wert, got)
			}
		}
	}
	for wort, soll := range map[string]model.Rueckgabe{"minimal": minimal, "representation": entity} {
		for _, w := range schreibweisen(wort) {
			n++
			if got := model.LiesRueckgabe([]string{"return=" + w}); got != soll {
				t.Errorf("return=%s: %d", w, got)
			}
			if got := model.LiesRueckgabe([]string{`return="` + w + `"`}); got != soll {
				t.Errorf("return=%q: %d", w, got)
			}
		}
	}
	if n != 64*2+128+16384 {
		t.Errorf("Menge %d, erwartet %d", n, 64*2+128+16384)
	}
}

func sollLeerraum(b byte) bool { return b == ' ' || b == '\t' }

// preferStelle ist eine Stelle, an der ein Byte in einem Element steht (Z3), mit der Menge der Bytes, bei denen die Zeile `Minimal` ergibt.
type preferStelle struct {
	name  string
	bau   func(b string) []string
	minim func(b byte) bool
}

// TestLiesRueckgabeLeerraum (Z3, SPEC-285 Nr. 2): Leerraum ist genau Leerzeichen und Tabulator, an jeder Stelle des Elements; jedes Byte
// 0x00 bis 0xFF an sieben Stellen. Orakel: je Stelle die Menge der Bytes, die `return=minimal` erhalten.
func TestLiesRueckgabeLeerraum(t *testing.T) {
	stellen := []preferStelle{
		{"vor dem Namen", func(b string) []string { return []string{b + "return=minimal"} },
			func(b byte) bool { return sollLeerraum(b) || b == ',' }},
		{"vor =", func(b string) []string { return []string{"return" + b + "=minimal"} }, sollLeerraum},
		{"nach =", func(b string) []string { return []string{"return=" + b + "minimal"} }, sollLeerraum},
		{"nach dem Wort", func(b string) []string { return []string{"return=minimal" + b} },
			func(b byte) bool { return sollLeerraum(b) || b == ';' || b == ',' }},
		{"vor ,", func(b string) []string { return []string{"respond-async" + b + ",return=minimal"} },
			func(b byte) bool { return b != '"' }},
		{"nach ,", func(b string) []string { return []string{"respond-async," + b + "return=minimal"} },
			func(b byte) bool { return sollLeerraum(b) || b == ',' }},
		{"vor ;", func(b string) []string { return []string{"return=minimal" + b + ";x=y"} },
			func(b byte) bool { return sollLeerraum(b) || b == ';' || b == ',' }},
	}
	n := 0
	for _, s := range stellen {
		for c := 0; c < 256; c++ {
			n++
			soll := keine
			if s.minim(byte(c)) {
				soll = minimal
			}
			if got := model.LiesRueckgabe(s.bau(string([]byte{byte(c)}))); got != soll {
				t.Errorf("%s, Byte 0x%02X: %d, erwartet %d", s.name, c, got, soll)
			}
		}
	}
	if n != 7*256 {
		t.Errorf("Menge %d, erwartet %d", n, 7*256)
	}
}

// TestLiesRueckgabeUnicodeLeerraum (Z3): kein Zeichen von unicode.IsSpace außerhalb von ASCII ist Leerraum, an den vier Stellen vor dem
// Namen, vor und nach `=` und nach dem Wort; jedes dieser Zeichen steht als UTF-8-Folge, aus dem Wert der Rune gebildet.
func TestLiesRueckgabeUnicodeLeerraum(t *testing.T) {
	var raeume []string
	for r := rune(0x80); r <= 0x3000; r++ {
		if unicode.IsSpace(r) {
			raeume = append(raeume, string(r))
		}
	}
	if len(raeume) != 19 {
		t.Fatalf("%d Zeichen von unicode.IsSpace jenseits von ASCII, erwartet 19", len(raeume))
	}
	for _, r := range raeume {
		for _, zeile := range []string{r + "return=minimal", "return" + r + "=minimal", "return=" + r + "minimal", "return=minimal" + r} {
			if got := model.LiesRueckgabe([]string{zeile}); got != keine {
				t.Errorf("%q: %d, erwartet Keine", zeile, got)
			}
		}
	}
}

// elemente der Folgen von Z4 mit dem, was das Element für die Wahl bedeutet: entscheidet es, und mit welchem Ergebnis.
func preferElemente() []struct {
	text        string
	entscheidet bool
	soll        model.Rueckgabe
} {
	return []struct {
		text        string
		entscheidet bool
		soll        model.Rueckgabe
	}{
		{"return=minimal", true, minimal}, {"return=representation", true, entity}, {"return=foo", true, keine}, {"return", true, keine},
		{"return=", true, keine}, {"respond-async", false, keine}, {"odata.x=1", false, keine}, {"return@x", false, keine}, {"", false, keine},
	}
}

// jedeFolge ruft f mit jeder Folge der Länge 0 bis n über den Zahlen 0 bis k-1; f behält die Folge nicht.
func jedeFolge(k, n int, f func([]int)) {
	var rek func(cur []int)
	rek = func(cur []int) {
		f(cur)
		if len(cur) == n {
			return
		}
		for i := 0; i < k; i++ {
			rek(append(cur, i))
		}
	}
	rek(nil)
}

// TestLiesRueckgabeErsteInstanz (Z4, SPEC-285 Nr. 3): die erste lesbare Instanz `return` entscheidet; jede Folge der Länge 0 bis 3 aus den neun
// Elementen, jede Verteilung auf Kopfzeilen (je Lücke dieselbe Zeile mit `,` oder die nächste Zeile): 1 + 9 + 81·2 + 729·4 = 3088 Fälle.
func TestLiesRueckgabeErsteInstanz(t *testing.T) {
	preferElemente := preferElemente()
	n := 0
	jedeFolge(len(preferElemente), 3, func(f []int) {
		soll := keine
		for _, i := range f {
			if preferElemente[i].entscheidet {
				soll = preferElemente[i].soll
				break
			}
		}
		for v := 0; v < 1<<max(len(f)-1, 0); v++ {
			var zeilen []string
			for k, i := range f {
				if k > 0 && v&(1<<(k-1)) == 0 {
					zeilen[len(zeilen)-1] += "," + preferElemente[i].text
				} else {
					zeilen = append(zeilen, preferElemente[i].text)
				}
			}
			n++
			if got := model.LiesRueckgabe(zeilen); got != soll {
				t.Errorf("%q: %d, erwartet %d", zeilen, got, soll)
			}
		}
	})
	if n != 3088 {
		t.Errorf("Menge %d, erwartet 3088", n)
	}
}

// TestLiesRueckgabeAnfuehrung (Z6, SPEC-285 Nr. 2): ein Komma in einem Paar trennt nicht, ein offenes Paar reicht bis zum Ende der Zeile und nie in
// die nächste; das Paar an den Stellen Name, Wert und Rest eines Elements, mit jedem Inhalt.
func TestLiesRueckgabeAnfuehrung(t *testing.T) {
	inhalte := []string{`a, return=representation`, `a,return=minimal`, `;return=minimal`, `\", return=minimal`, `\\, return=minimal`, `,`}
	n := 0
	for _, in := range inhalte {
		paar := `"` + in + `"`
		for _, zeile := range []string{
			`x=` + paar,                // Wert
			`x=1;p=` + paar,            // Rest
			`x` + paar + `=1`,          // im Namen: das Element ist unlesbar
			`x=` + paar + `;p=` + paar, // zwei Paare
			`respond-async; p=` + paar, // Rest eines anderen Elements
		} {
			n++
			if got := model.LiesRueckgabe([]string{zeile}); got != keine {
				t.Errorf("%q: %d, erwartet Keine (das Paar hält jedes Komma)", zeile, got)
			}
			n++
			if got := model.LiesRueckgabe([]string{zeile + ", return=minimal"}); got != minimal {
				t.Errorf("%q mit folgendem Element: %d, erwartet Minimal", zeile, got)
			}
		}
	}
	// ein Paar, das in der Zeile nicht schließt, reicht nicht in die nächste Zeile
	offen := []string{`x="`, `x="a`, `x="a\`, `x="a\"`, `x=1; p="`}
	for _, o := range offen {
		n++
		if got := model.LiesRueckgabe([]string{o, "return=minimal"}); got != minimal {
			t.Errorf("%q, dann return=minimal: %d, erwartet Minimal", o, got)
		}
		n++
		if got := model.LiesRueckgabe([]string{o + ", return=minimal"}); got != keine {
			t.Errorf("%q in einer Zeile mit return=minimal: %d, erwartet Keine (das Paar reicht bis zum Ende)", o, got)
		}
	}
	// ein Backslash schützt das Schlusszeichen und sich selbst
	for zeile, soll := range map[string]model.Rueckgabe{
		`x="a\\", return=minimal`: minimal, `x="a\", return=minimal"`: keine, `x="a\", return=minimal", return=representation`: entity,
	} {
		n++
		if got := model.LiesRueckgabe([]string{zeile}); got != soll {
			t.Errorf("%q: %d, erwartet %d", zeile, got, soll)
		}
	}
	if n != len(inhalte)*5*2+len(offen)*2+3 {
		t.Errorf("Menge %d", n)
	}
}

// TestLiesRueckgabeRest (Z7, SPEC-285 Nr. 2): `Rest` hinter dem ersten `;` hat keine Wirkung, auch ein Komma darin außerhalb eines Paares beginnt nur
// ein neues Element; sieben Formen des Rests × beide Werte, ohne und mit einem späteren Element `return`.
func TestLiesRueckgabeRest(t *testing.T) {
	reste := []string{";", "; a", "; a=b", ";;", "; =", `; "`, "; ,x", "; return=representation", ";,return=representation"}
	n := 0
	for wert, soll := range map[string]model.Rueckgabe{"return=minimal": minimal, "return=representation": entity} {
		for _, rest := range reste {
			n++
			if got := model.LiesRueckgabe([]string{wert + rest}); got != soll {
				t.Errorf("%q: %d, erwartet %d", wert+rest, got, soll)
			}
			n++
			if got := model.LiesRueckgabe([]string{wert + rest + ", return=" + "foo"}); got != soll {
				t.Errorf("%q mit späterem return: %d, erwartet %d", wert+rest, got, soll)
			}
		}
	}
	if n != 2*len(reste)*2 {
		t.Errorf("Menge %d", n)
	}
}

// byteStelle setzt ein Byte an die Stelle eines Elements: Anfang des Namens, direkt hinter `return`, direkt hinter dem Wort (Z8), mit dem Ergebnis je Byte.
type byteStelle struct {
	name string
	bau  func(b string) []string
	soll func(b byte) model.Rueckgabe
}

func byteStellen() []byteStelle {
	return []byteStelle{
		{"erstes Zeichen des Namens", func(b string) []string { return []string{b + "return=minimal, return=representation"} }, func(b byte) model.Rueckgabe {
			switch {
			case sollLeerraum(b) || b == ',':
				return minimal
			case b == '"':
				return keine
			}
			return entity
		}},
		{"direkt hinter return", func(b string) []string { return []string{"return" + b + "=minimal, return=representation"} }, func(b byte) model.Rueckgabe {
			switch {
			case sollLeerraum(b):
				return minimal
			case b == ',' || b == '=' || b == ';' || b == '"':
				return keine
			}
			return entity
		}},
		{"hinter dem Wort", func(b string) []string { return []string{"return=minimal" + b + ", return=representation"} }, func(b byte) model.Rueckgabe {
			if sollLeerraum(b) || b == ';' || b == ',' {
				return minimal
			}
			return keine
		}},
	}
}

// TestLiesRueckgabeUnlesbar (Z8, SPEC-285 Nr. 2, 3): ein unlesbares Element zählt nicht, ein lesbares mit unbrauchbarem Wert zählt; jedes Byte an drei
// Stellen, mit einem folgenden Element `return=representation`.
func TestLiesRueckgabeUnlesbar(t *testing.T) {
	n := 0
	for _, s := range byteStellen() {
		for c := 0; c < 256; c++ {
			n++
			if got, soll := model.LiesRueckgabe(s.bau(string([]byte{byte(c)}))), s.soll(byte(c)); got != soll {
				t.Errorf("%s, Byte 0x%02X: %d, erwartet %d", s.name, c, got, soll)
			}
		}
	}
	if n != 3*256 {
		t.Errorf("Menge %d", n)
	}
}

// Das Orakel der Grammatik (SPEC-285 Nr. 2, 3) ist ein regulärer Ausdruck je Schritt über der Zeile, in der jedes Byte eine Rune gleichen Werts ist; es
// teilt keinen Code mit dem Leser des Modells.
var (
	preferElement   = regexp.MustCompile(`(?s)((?:[^",]|"(?:\\.|[^"\\])*(?:"|\\?$))*)(?:,|$)`)
	preferReturn    = regexp.MustCompile(`(?s)^(?i:return)(?:[ \t]|=|;|$)`)
	preferWert      = regexp.MustCompile("(?s)^(?i:return)[ \t]*=[ \t]*(?:([-!#$%&'*+.^_`|~0-9A-Za-z]+)|\"((?:\\\\.|[^\"\\\\])*)\")[ \t]*(?:;|$)")
	preferBackslash = regexp.MustCompile(`(?s)\\(.)`)
)

func preferOrakel(zeile string) model.Rueckgabe {
	for _, m := range preferElement.FindAllStringSubmatch(alsRunen(zeile), -1) {
		e := strings.TrimLeft(m[1], " \t")
		if !preferReturn.MatchString(e) {
			continue
		}
		w := preferWert.FindStringSubmatch(e)
		if w == nil {
			return keine
		}
		wort := w[1]
		if w[2] != "" {
			wort = preferBackslash.ReplaceAllString(w[2], "$1")
		}
		switch {
		case strings.EqualFold(wort, "minimal"):
			return minimal
		case strings.EqualFold(wort, "representation"):
			return entity
		}
		return keine
	}
	return keine
}

// TestLiesRueckgabeFolgen (Z9, SPEC-285 Nr. 1): die Funktion ändert die Eingabe nie, löst keine Panik aus und liefert für jede Folge der Länge 0 bis 6 über
// dem Alphabet {return, =, ;, ", Backslash, ',', Leerzeichen, Tabulator, minimal, x, 0xA0} in einer Zeile dasselbe wie in drei Zeilen mit zwei fremden
// ("z" und ""); bis Länge 5 über dem Alphabet mit `representation` dasselbe wie das Orakel.
func TestLiesRueckgabeFolgen(t *testing.T) {
	alphabet := []string{"return", "=", ";", `"`, `\`, ",", " ", "\t", "minimal", "x", a0Byte()}
	baue := func(alpha []string, f []int) string {
		var b strings.Builder
		for _, i := range f {
			b.WriteString(alpha[i])
		}
		return b.String()
	}
	gezaehlt := 0
	jedeFolge(len(alphabet), 6, func(f []int) {
		s := baue(alphabet, f)
		zeilen := []string{"z", s, ""}
		vorher := slices.Clone(zeilen)
		got := model.LiesRueckgabe(zeilen)
		gezaehlt++
		if !slices.Equal(zeilen, vorher) {
			t.Fatalf("Eingabe verändert: %q, vorher %q", zeilen, vorher)
		}
		if one := model.LiesRueckgabe([]string{s}); one != got {
			t.Fatalf("%q: allein %d, zwischen fremden Zeilen %d", s, one, got)
		}
	})
	if want := 1 + 11 + 121 + 1331 + 14641 + 161051 + 1771561; gezaehlt != want {
		t.Errorf("Menge %d, erwartet %d", gezaehlt, want)
	}
	orakelAlphabet := append(slices.Clone(alphabet), "representation")
	verglichen := 0
	jedeFolge(len(orakelAlphabet), 5, func(f []int) {
		s := baue(orakelAlphabet, f)
		verglichen++
		if got, soll := model.LiesRueckgabe([]string{s}), preferOrakel(s); got != soll {
			t.Fatalf("%q: %d, Orakel %d", s, got, soll)
		}
	})
	if want := 1 + 12 + 144 + 1728 + 20736 + 248832; verglichen != want {
		t.Errorf("Menge des Orakels %d, erwartet %d", verglichen, want)
	}
}
