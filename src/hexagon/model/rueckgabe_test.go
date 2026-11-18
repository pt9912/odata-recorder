package model_test

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
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

// backslashWorte sind die Wörter, deren Auflösen Z6a prüft: zwei Wörter der Wahl und ein Wort in anderer Schreibweise.
func backslashWorte() []string { return []string{"minimal", "representation", "MiNiMaL"} }

// backslashFall ist der Inhalt eines Wortes in Anführungszeichen mit Backslash, der Wert nach dem Auflösen und die Aussage, ob das Paar schließt.
type backslashFall struct {
	inhalt, aufgeloest string
	schliesst          bool
}

// loeseAuf ist das Orakel des Auflösens (SPEC-285 Nr. 2): ein Backslash ersetzt sich durch das folgende Zeichen; ok ist falsch, wenn er am Ende steht.
func loeseAuf(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 == len(s) {
			return "", false
		}
		i++
		b.WriteByte(s[i])
	}
	return b.String(), true
}

// backslashFaelle erzeugt je Wort v die Klasse „Backslash im Wort“: jede Stelle von v mit einem Backslash vor dem eigenen Zeichen (L), alle Zeichen so
// geschützt (1), ein geschütztes Zeichen aus sieben vor jeder Stelle eingefügt (7·(L+1)), jede Stelle durch eines von drei geschützten Zeichen ersetzt (3·L),
// der Backslash vor dem Schlusszeichen und am Ende der Zeile (2).
func backslashFaelle(v string) []backslashFall {
	var r []backslashFall
	add := func(inhalt, aufgeloest string, schliesst bool) {
		r = append(r, backslashFall{inhalt, aufgeloest, schliesst})
	}
	for p := 0; p < len(v); p++ {
		add(v[:p]+"\\"+v[p:p+1]+v[p+1:], v, true)
	}
	var alle strings.Builder
	for p := 0; p < len(v); p++ {
		alle.WriteString("\\" + v[p:p+1])
	}
	add(alle.String(), v, true)
	for p := 0; p <= len(v); p++ {
		for _, c := range []string{`"`, `\`, ",", ";", "=", " ", "x"} {
			add(v[:p]+"\\"+c+v[p:], v[:p]+c+v[p:], true)
		}
	}
	for p := 0; p < len(v); p++ {
		for _, c := range []string{"x", `"`, `\`} {
			add(v[:p]+"\\"+c+v[p+1:], v[:p]+c+v[p+1:], true)
		}
	}
	add(v+"\\", "", true)  // `\"` schützt das Schlusszeichen: das Paar bleibt offen
	add(v+"\\", "", false) // der Backslash steht am Ende der Zeile
	return r
}

// TestLiesRueckgabeBackslashWort (Z6a, SPEC-285 Nr. 2, 3): ein Backslash im Wort in Anführungszeichen ersetzt sich durch das folgende Zeichen, und der Wert gilt
// nach dem Auflösen; die Klasse „Backslash im Wort“ von drei Wörtern in drei Zusammenhängen (allein, mit Rest, mit folgendem Element `return`), 3 · 338 = 1014 Fälle;
// das Orakel ist das Auflösen selbst, und der Orakel-Leser aus Z9 stimmt zu.
func TestLiesRueckgabeBackslashWort(t *testing.T) {
	zusammenhaenge := []struct {
		name string
		bau  func(wort string) string
	}{
		{"allein", func(w string) string { return `return=` + w }},
		{"mit Rest", func(w string) string { return ` return = ` + w + ` ; x=1` }},
		{"mit folgendem Element", func(w string) string { return `x=1, return=` + w + `, return=minimal` }},
	}
	n := 0
	for _, v := range backslashWorte() {
		faelle := backslashFaelle(v)
		if len(faelle) != 11*len(v)+10 {
			t.Fatalf("%s: %d Fälle, erwartet %d", v, len(faelle), 11*len(v)+10)
		}
		for _, f := range faelle {
			if a, ok := loeseAuf(f.inhalt); f.aufgeloest != "" && (!ok || a != f.aufgeloest) {
				t.Fatalf("Erzeuger: %q löst zu %q auf, erwartet %q", f.inhalt, a, f.aufgeloest)
			}
			soll := keine
			switch {
			case !f.schliesst, f.aufgeloest == "":
			case strings.EqualFold(f.aufgeloest, "minimal"):
				soll = minimal
			case strings.EqualFold(f.aufgeloest, "representation"):
				soll = entity
			}
			wort := `"` + f.inhalt
			if f.schliesst {
				wort += `"`
			}
			for _, z := range zusammenhaenge {
				n++
				zeile := z.bau(wort)
				got := model.LiesRueckgabe([]string{zeile})
				if got != soll {
					t.Errorf("%s %q: %d, erwartet %d", z.name, zeile, got, soll)
				}
				if orakel := preferOrakel(zeile); got != orakel {
					t.Errorf("%s %q: %d, Orakel %d", z.name, zeile, got, orakel)
				}
			}
		}
	}
	if n != 3*(87+164+87) {
		t.Errorf("Menge %d, erwartet %d", n, 3*(87+164+87))
	}
}

// zeitFormen sind die Formen langer Zeilen, die TestLiesRueckgabeZeit prüft; n zählt die Wiederholungen des Bausteins.
func zeitFormen() []struct {
	name string
	bau  func(n int) []string
} {
	return []struct {
		name string
		bau  func(n int) []string
	}{
		{"Elemente", func(n int) []string { return []string{strings.Repeat("x,", n)} }},
		{"Zeilen", func(n int) []string {
			z := make([]string, n)
			for i := range z {
				z[i] = "x"
			}
			return z
		}},
		{"Name", func(n int) []string { return []string{strings.Repeat("a", 2*n)} }},
		{"Leerraum", func(n int) []string { return []string{strings.Repeat(" ", 2*n) + "return=minimal"} }},
		{"Paar mit Backslash", func(n int) []string { return []string{`return="` + strings.Repeat(`\a`, n) + `"`} }},
		{"offenes Paar", func(n int) []string { return []string{`x="` + strings.Repeat("a,", n)} }},
		{"Kommas", func(n int) []string { return []string{strings.Repeat(", ", n)} }},
	}
}

// besteZeit ist die kleinste Dauer von k Läufen von f.
func besteZeit(k int, f func()) time.Duration {
	best := time.Duration(math.MaxInt64)
	for i := 0; i < k; i++ {
		t0 := time.Now()
		f()
		best = min(best, time.Since(t0))
	}
	return best
}

// TestLiesRueckgabeZeit (Z9a, SPEC-285 Nr. 1): die Zeit wächst nicht schneller als linear mit der Länge der Zeilen; sieben Formen langer Zeilen (Elemente, Zeilen, Name, Leerraum,
// Paar mit Backslash, offenes Paar, Kommas) mit 2000 und mit 64000 Wiederholungen. Bei linearem Wachstum wächst die Zeit um das 32-fache, bei quadratischem um das 1024-fache;
// die Grenze liegt beim 128-fachen der besten Zeit der kleinen Länge, und der Lauf mit der großen Länge zählt als bestanden, sobald einer von fünf Versuchen darunter liegt.
// Der Fänger misst Zeit: unter Last des Hosts kann er ohne Fehler im Code rot werden, und er fängt nur die sieben Formen; die Raten stehen im Plan (§3).
func TestLiesRueckgabeZeit(t *testing.T) {
	const klein, gross = 2000, 64000
	n := 0
	for _, f := range zeitFormen() {
		n++
		zk, zg := f.bau(klein), f.bau(gross)
		var wahl model.Rueckgabe // hält das Ergebnis, damit der Aufruf nicht entfällt
		t1 := max(besteZeit(20, func() { wahl = model.LiesRueckgabe(zk) }), time.Microsecond)
		grenze := 128 * t1
		ok, letzte := false, time.Duration(0)
		for versuch := 0; versuch < 5 && !ok; versuch++ {
			letzte = besteZeit(1, func() { wahl = model.LiesRueckgabe(zg) })
			ok = letzte <= grenze
		}
		if wahl < keine || wahl > entity {
			t.Errorf("%s: Wahl %d außerhalb der drei Werte", f.name, wahl)
		}
		if !ok {
			t.Errorf("%s: %v für %d Wiederholungen, %v für %d; Grenze das 128-fache", f.name, letzte, gross, t1, klein)
		}
	}
	if n != 7 {
		t.Errorf("Menge %d, erwartet 7", n)
	}
}
