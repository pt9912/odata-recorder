package model_test

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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

// bs ist eine Folge von k Backslashes.
func bs(k int) string { return strings.Repeat(`\`, k) }

// TestLiesRueckgabeBackslashFremd (Z6b, SPEC-285 Nr. 2): ein Backslash im Paar eines fremden Elements schützt das folgende Zeichen; eine Folge von k Backslashes vor
// dem Schlusszeichen schützt es bei ungeradem k und lässt es bei geradem k schließen: Stelle des Paares (Wert, Rest, Rest eines anderen Elements, Name) 4 × Inhalt vor
// der Folge 3 × k 0 bis 7 × Zeichen hinter der Folge (`"`, `,`, `a`, Ende der Zeile) 4 = 384 Fälle; das Orakel ist die Parität von k und der Orakel-Leser aus Z9.
// Der Backslash außerhalb eines Paares und an den übrigen Stellen der Grammatik steht in TestLiesRueckgabeBackslashStellen.
func TestLiesRueckgabeBackslashFremd(t *testing.T) {
	n := 0
	stellen := []string{`x=`, `x=1;p=`, `respond-async; p=`, `x`}
	vorher := []string{``, `a`, `, return=representation`}
	for _, st := range stellen {
		for _, pre := range vorher {
			for k := 0; k <= 7; k++ {
				for _, c := range []string{`"`, `,`, `a`, ``} {
					n++
					kopf := st + `"` + pre + bs(k) + c
					soll := minimal
					if c == `"` && k%2 == 1 {
						soll = keine // der Backslash schützt das Schlusszeichen, das Paar reicht bis zum Ende der Zeile
					}
					zeile := kopf + `", return=minimal`
					if c == `` {
						// das Paar ist am Ende der Zeile offen und reicht nicht in die nächste
						if got := model.LiesRueckgabe([]string{kopf}); got != keine {
							t.Errorf("%q: %d, erwartet Keine", kopf, got)
						}
						if got := model.LiesRueckgabe([]string{kopf, "return=minimal"}); got != minimal {
							t.Errorf("%q, dann return=minimal: %d, erwartet Minimal", kopf, got)
						}
						continue
					}
					if c == `"` {
						zeile = kopf + `, return=minimal`
					}
					if got := model.LiesRueckgabe([]string{zeile}); got != soll {
						t.Errorf("%q: %d, erwartet %d (k=%d)", zeile, got, soll, k)
					}
					if orakel := preferOrakel(zeile); orakel != soll {
						t.Errorf("%q: Orakel %d, erwartet %d", zeile, orakel, soll)
					}
				}
			}
		}
	}
	if n != 4*3*8*4 {
		t.Errorf("Menge im Paar %d, erwartet %d", n, 4*3*8*4)
	}
}

// backslashStellen sind die Stellen einer Kopfzeile, an denen ein Backslash stehen kann (SPEC-285 Nr. 1 bis 3), je Positionsklasse der Grammatik; `@` ist die
// Stelle, an die der Test die Folge von Backslashes samt dem Zeichen dahinter setzt. Die Klassen: Zeilenanfang, Elementanfang (hinter `,`, hinter Leerraum),
// im Namen, hinter dem Namen, vor und hinter `=`, im Wert als Token, im Wort in Anführungszeichen (Anfang, Mitte, Ende, unmittelbar hinter dem Schlusszeichen,
// offen am Zeilenende), im Paar im Namen, hinter `;` und im Rest (auch im Paar des Rests), hinter `,`, Zeilenende; dazu dieselben Stellen im Element `return` selbst.
func backslashStellen() []string {
	return []string{
		// Zeilenanfang und Elementanfang
		`@`, `@x=1`, `x=1,@y=2`, `x=1, @y=2`, "x=1,\t@y=2", `x=1,@`, `x=1,@,y`,
		// Elementanfang unmittelbar vor `return`: hinter `,`, hinter Leerraum, hinter dem Paar, hinter dem Rest
		`x=1,@return=minimal`, `x=1, @return=minimal`, "x=1,\t@return=minimal", `x="a",@return=minimal`, `x=1;p=a,@return=minimal`, `x=1;p="a",@return=minimal`,
		// Name, hinter dem Namen, vor und hinter `=`
		`x@y=1`, `x@=1`, `x@`, `x @=1`, `x=@1`, `x= @1`, `x=@`,
		// Wert als Token
		`x=1@2`, `x=1@`, `x=1@;p=a`, `x=1@,y`, `x=1 @`, `x=1 @;p=a`, `x=1@,@y`,
		// Wort in Anführungszeichen: Anfang, Mitte, Ende, unmittelbar hinter dem Schlusszeichen, offen
		`x="@"`, `x="@a"`, `x="a@b"`, `x="a@"`, `x="a"@`, `x="a"@;p=1`, `x="a" @`, `x=""@`, `x="a"@"b"`, `x="@`, `x="a@`,
		// Paar im Namen
		`x"a@b"=1`, `x"a"@=1`, `x"@`,
		// hinter `;` und im Rest, auch im Paar des Rests
		`x=1;@`, `x=1;@p=a`, `x=1;p@a`, `x=1;p=@a`, `x=1;p=@`, `x=1; @`, `x=1;p=a,@y`, `respond-async;@`, `respond-async; p=@`,
		`x=1;p="a@b"`, `x=1;p="a"@`, `x=1;p="@`,
		// das Element `return` selbst
		`@return=minimal`, `re@turn=minimal`, `return@=minimal`, `return@`, `return=@`, `return =@minimal`, `return=@minimal`, `return=@"minimal"`,
		`return="@minimal"`, `return="min@imal"`, `return="minimal@"`, `return="minimal"@`, `return=minimal@`, `return=minimal @`, `return=minimal@;x`,
		`return=minimal;@`, `return=minimal;p="a@b"`, `return=minimal,@`,
	}
}

// backslashLagen setzen eine Stelle in die Zeilen einer Anfrage: das Element mit der Stelle vor dem Element `return`, nach ihm, in der Mitte, hinter einer
// Zeilengrenze, vor einer Zeile mit `return=representation` und allein.
func backslashLagen() []func(f string) []string {
	return []func(f string) []string{
		func(f string) []string { return []string{f + ", return=minimal"} },
		func(f string) []string { return []string{"return=minimal, " + f} },
		func(f string) []string { return []string{"y, " + f + ", return=minimal"} },
		func(f string) []string { return []string{f, "return=minimal"} },
		func(f string) []string { return []string{"return=representation", f} },
		func(f string) []string { return []string{f} },
	}
}

// TestLiesRueckgabeBackslashStellen (Z6c, SPEC-285 Nr. 1 bis 3): ein Backslash an den Stellen aus backslashStellen (die Positionsklassen dort, von Hand gewählt, kein
// Anspruch auf jede Stelle): 71 Stellen × 6 Lagen (Element vor, nach, in der Mitte, über eine Zeilengrenze, vor `return=representation`, allein) × Folgenlänge 0 bis 7 ×
// Zeichen hinter der Folge (`"`, `,`, `;`, `=`, `a`, Leerzeichen, Ende der Stelle) 7 = 23856 Zeilen; je Zeile gleich dem Orakel aus Z9, das keinen Code mit dem Leser
// teilt. Die Menge ist der Gegenstand der Lage und der Folge; die Zeichen vor der Folge und der Kontext eines Paares führt TestLiesRueckgabeErschoepfend. Ungeprüft
// bleibt hier: ein Zeichen hinter der Folge außer den sieben, mehr als ein Element zwischen Stelle und `return`.
func TestLiesRueckgabeBackslashStellen(t *testing.T) {
	stellen := backslashStellen()
	if len(stellen) != 71 {
		t.Fatalf("Stellen %d, erwartet 71", len(stellen))
	}
	lagen := backslashLagen()
	n := 0
	ergebnisse := map[model.Rueckgabe]int{}
	for _, st := range stellen {
		for _, lage := range lagen {
			for k := 0; k <= 7; k++ {
				for _, c := range []string{`"`, `,`, `;`, `=`, `a`, ` `, ``} {
					zeilen := lage(strings.ReplaceAll(st, "@", bs(k)+c))
					vorher := slices.Clone(zeilen)
					soll := orakelZeilen(zeilen)
					got := model.LiesRueckgabe(zeilen)
					n++
					ergebnisse[got]++
					if got != soll {
						t.Errorf("%q: %d, Orakel %d (Stelle %q, k=%d, Zeichen dahinter %q)", zeilen, got, soll, st, k, c)
					}
					if !slices.Equal(zeilen, vorher) {
						t.Fatalf("Eingabe verändert: %q", zeilen)
					}
				}
			}
		}
	}
	if n != 71*6*8*7 {
		t.Errorf("Menge %d, erwartet %d", n, 71*6*8*7)
	}
	for _, r := range []model.Rueckgabe{keine, minimal, entity} {
		if ergebnisse[r] == 0 {
			t.Errorf("kein Fall mit dem Ergebnis %d: die Menge trennt die Ergebnisse nicht", r)
		}
	}
}

// folgenParallel ruft f mit jeder Folge der Länge 0 bis n über den Symbolen (die Folge als Text, ihre Symbole aneinandergehängt), je erstes Symbol in einer
// Goroutine, und liefert die Zahl der Folgen; f ist nebenläufig aufrufbar.
func folgenParallel(symbole []string, n int, f func(s string)) int64 {
	var gezaehlt atomic.Int64
	f("")
	gezaehlt.Add(1)
	var wg sync.WaitGroup
	for _, erstes := range symbole {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var c int64
			var rek func(buf []byte, tiefe int)
			rek = func(buf []byte, tiefe int) {
				f(string(buf))
				c++
				if tiefe == n {
					return
				}
				for _, s := range symbole {
					rek(append(buf, s...), tiefe+1)
				}
			}
			rek([]byte(erstes), 1)
			gezaehlt.Add(c)
		}()
	}
	wg.Wait()
	return gezaehlt.Load()
}

// summeHoch ist 1 + k + k² + … + kⁿ, die Zahl der Folgen der Länge 0 bis n über k Symbolen.
func summeHoch(k, n int) int64 {
	var s, p int64 = 0, 1
	for i := 0; i <= n; i++ {
		s += p
		p *= int64(k)
	}
	return s
}

// erschoepfend vergleicht die Funktion mit dem Orakel für jede Folge von rahmen(s), s aus folgenParallel; meldet höchstens zehn Abweichungen und zählt die Ergebnisse.
func erschoepfend(t *testing.T, symbole []string, n int, rahmen func(s string) []string) (int64, [3]int64) {
	t.Helper()
	var fehler atomic.Int64
	var ergebnisse [3]atomic.Int64
	gezaehlt := folgenParallel(symbole, n, func(s string) {
		zeilen := rahmen(s)
		got := model.LiesRueckgabe(zeilen)
		soll := orakelZeilen(zeilen)
		ergebnisse[soll].Add(1)
		if got != soll && fehler.Add(1) <= 10 {
			t.Errorf("%q: %d, Orakel %d", zeilen, got, soll)
		}
	})
	return gezaehlt, [3]int64{ergebnisse[0].Load(), ergebnisse[1].Load(), ergebnisse[2].Load()}
}

// TestLiesRueckgabeErschoepfend (Z6d, SPEC-285 Nr. 1 bis 3): drei erschöpfende Mengen, je Zeile gleich dem Orakel aus Z9, das keinen Code mit dem Leser teilt.
// Die Symbolmenge: jede Folge von 0 bis 6 Symbolen über {Backslash, `"`, `,`, `;`, `=`, Leerzeichen, Tabulator, `a`, `return`, `minimal`, `representation`}
// als eine Zeile, 1.948.717 Zeilen. Die Zeichenmenge: jede Folge von 0 bis 7 Zeichen über {Backslash, `"`, `,`, `;`, `=`, Leerzeichen, Tabulator, `a`} vor
// `,return=minimal`, 2.396.745 Zeilen. Die Strukturmenge: jede Folge von 0 bis 7 Symbolen über {Backslash, `"`, `,`, `;`, `=`, Leerzeichen, `return`, `minimal`}
// als eine Zeile, 2.396.745 Zeilen; sie lässt Tabulator, `a` und `representation` weg, weil der Leser sie wie Leerzeichen, ein Tokenzeichen und `minimal`
// behandelt (die Symbolmenge führt sie). Die Mengen führen jede Lage des Backslash zu den Zeichen davor und dahinter und zu den Paaren (am Elementanfang, hinter `;`,
// hinter einem Paar, späteres Paar) bis zu ihrer Länge; die Zählung steht im Test.
// Ungeprüft bleibt, was länger ist: mehr als sechs, sieben und sieben Symbole; dazu mehr als eine Zeile, ein anderer Buchstabe als `a` und in der Zeichenmenge ein
// Element `return` vor dem festen.
func TestLiesRueckgabeErschoepfend(t *testing.T) {
	symbole := []string{`\`, `"`, ",", ";", "=", " ", "\t", "a", "return", "minimal", "representation"}
	n, erg := erschoepfend(t, symbole, 6, func(s string) []string { return []string{s} })
	if want := summeHoch(len(symbole), 6); n != want || want != 1948717 {
		t.Errorf("Menge der Symbolfolgen %d, erwartet %d", n, want)
	}
	for r, c := range erg {
		if c == 0 {
			t.Errorf("Symbolmenge: kein Fall mit dem Ergebnis %d", r)
		}
	}
	zeichen := []string{`\`, `"`, ",", ";", "=", " ", "\t", "a"}
	n, erg = erschoepfend(t, zeichen, 7, func(s string) []string { return []string{s + ",return=minimal"} })
	if want := summeHoch(len(zeichen), 7); n != want || want != 2396745 {
		t.Errorf("Menge der Zeichenfolgen %d, erwartet %d", n, want)
	}
	if erg[keine] == 0 || erg[minimal] == 0 {
		t.Errorf("Zeichenmenge: Ergebnisse %v trennen Keine und Minimal nicht", erg)
	}
	struktur := []string{`\`, `"`, ",", ";", "=", " ", "return", "minimal"}
	n, erg = erschoepfend(t, struktur, 7, func(s string) []string { return []string{s} })
	if want := summeHoch(len(struktur), 7); n != want || want != 2396745 {
		t.Errorf("Menge der Strukturfolgen %d, erwartet %d", n, want)
	}
	for r, c := range erg[:2] {
		if c == 0 {
			t.Errorf("Strukturmenge: kein Fall mit dem Ergebnis %d", r)
		}
	}
}

// orakelZeilen ist das Orakel über mehrere Kopfzeilen: die erste Zeile mit einem lesbaren Element `return` entscheidet; eine Zeile reicht nie in die nächste.
func orakelZeilen(zeilen []string) model.Rueckgabe {
	for _, z := range zeilen {
		if r, entschieden := preferOrakelZeile(z); entschieden {
			return r
		}
	}
	return keine
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
	r, _ := preferOrakelZeile(zeile)
	return r
}

// preferOrakelZeile ist die Wahl einer Zeile und die Aussage, ob ein lesbares Element `return` darin stand.
func preferOrakelZeile(zeile string) (model.Rueckgabe, bool) {
	for _, m := range preferElement.FindAllStringSubmatch(alsRunen(zeile), -1) {
		e := strings.TrimLeft(m[1], " \t")
		if !preferReturn.MatchString(e) {
			continue
		}
		w := preferWert.FindStringSubmatch(e)
		if w == nil {
			return keine, true
		}
		wort := w[1]
		if w[2] != "" {
			wort = preferBackslash.ReplaceAllString(w[2], "$1")
		}
		switch {
		case strings.EqualFold(wort, "minimal"):
			return minimal, true
		case strings.EqualFold(wort, "representation"):
			return entity, true
		}
		return keine, true
	}
	return keine, false
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

// zeitForm ist eine Stelle der Grammatik von SPEC-285 Nr. 2, an der ein Lauf beliebig lang sein kann; n zählt die Wiederholungen des Bausteins (zwei Zeichen je Baustein,
// Zeilen und Elemente je einen).
type zeitForm struct {
	name string
	bau  func(n int) []string
}

// zeitFormen sind die Stellen des langen Laufs, die TestLiesRueckgabeZeit prüft: die Stellen der Grammatik, an denen ein Lauf von Zeichen steht, je Zeile (Zeilen),
// je Element (Elemente, Kommas), im Leerraum (vor dem Namen, hinter dem Namen, hinter `=`, hinter dem Wort), im Namen, im Wort als Token, im Wort in Anführungszeichen
// (ohne Backslash, mit Backslash, offen), im Rest hinter dem Wert, in einem fremden Element (Wert, Rest, Paar ohne Backslash, Paar mit Backslash, offenes Paar) und als
// Backslash außerhalb eines Paares.
func zeitFormen() []zeitForm {
	wdh := func(s string) func(n int) string { return func(n int) string { return strings.Repeat(s, n) } }
	a2, bsa, bs2, leer, komma := wdh("aa"), wdh(`\a`), wdh(`\\`), wdh("  "), wdh("a,")
	eine := func(f func(n int) string) func(n int) []string {
		return func(n int) []string { return []string{f(n)} }
	}
	rund := func(vor string, f func(n int) string, nach string) func(n int) []string {
		return eine(func(n int) string { return vor + f(n) + nach })
	}
	return []zeitForm{
		{"Zeilen", func(n int) []string {
			z := make([]string, n)
			for i := range z {
				z[i] = "x"
			}
			return z
		}},
		{"Elemente", rund("", wdh("x,"), "")},
		{"Kommas", rund("", wdh(", "), "")},
		{"Leerraum vor dem Namen", rund("", leer, "return=minimal")},
		{"Leerraum hinter dem Namen", rund("return", leer, "=minimal")},
		{"Leerraum hinter =", rund("return=", leer, "minimal")},
		{"Leerraum hinter dem Wort", rund("return=minimal", leer, "")},
		{"Name", rund("", a2, "")},
		{"Wert als Token", rund("return=", a2, "")},
		{"Wort ohne Backslash", rund(`return="`, a2, `"`)},
		{"Wort mit Backslash", rund(`return="`, bsa, `"`)},
		{"offenes Wort", rund(`return="`, a2, "")},
		{"Rest hinter dem Wert", rund("return=minimal;", a2, "")},
		{"fremder Wert", rund("x=", a2, "")},
		{"fremder Rest", rund("x;", a2, "")},
		{"fremdes Paar ohne Backslash", rund(`x="`, a2, `"`)},
		{"fremdes Paar mit Backslash", rund(`x="`, bsa, `"`)},
		{"fremdes Paar mit Folge von Backslashes", rund(`x="`, bs2, `"`)},
		{"offenes fremdes Paar", rund(`x="`, komma, "")},
		{"Backslash außerhalb eines Paares", rund("x", bs2, "")},
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

// TestLiesRueckgabeZeit (Z9a, SPEC-285 Nr. 1): die Zeit wächst nicht schneller als linear mit der Länge der Zeilen; die 20 Stellen des langen Laufs aus zeitFormen mit 2000
// und mit 64000 Wiederholungen. Bei linearem Wachstum wächst die Zeit um das 32-fache, bei quadratischem um das 1024-fache;
// die Grenze liegt beim 128-fachen der besten Zeit der kleinen Länge, und der Lauf mit der großen Länge zählt als bestanden, sobald einer von fünf Versuchen darunter liegt.
// Der Fänger misst Zeit: unter Last des Hosts kann er ohne Fehler im Code rot werden, und er fängt nur die Stellen der Liste mit je einem Baustein; die Raten stehen im Plan (§3).
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
	if n != 20 {
		t.Errorf("Menge %d, erwartet 20", n)
	}
}
