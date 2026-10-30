package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// Tests zu V-1 und V-2 der Verifikation (SPEC-200 Nr. 2 d, 3 c, 4 a, 4 c, 4 d, 6, 8, 9 b, 14 d): jede Zusage läuft als
// Schleife über die Menge, die ihr Wortlaut meint; das Orakel steht im Test, von Hand gesetzt, nicht aus dem Code.

// befundZeilen liefert die Zeilen der Ausgabe ohne den Kopf `pruefung handbuch: `.
func befundZeilen(t *testing.T, errText string) []string {
	t.Helper()
	if errText == "" {
		return nil
	}
	var out []string
	for _, z := range strings.Split(strings.TrimSuffix(errText, "\n"), "\n") {
		s, ok := strings.CutPrefix(z, ausgabeKopf)
		if !ok {
			t.Fatalf("Zeile ohne Kopf der Ausgabe: %q", z)
		}
		out = append(out, s)
	}
	return out
}

// zeileText ist ein Befund mit der Zeile, die ihm der Vertrag zuweist (Nr. 2 d), und seinem Text.
type zeileText struct {
	zeile int
	text  string
}

// erwarteReihenfolge prüft, dass die Ausgabe genau die Befunde nennt, Element für Element: nach Zeile aufsteigend, bei
// gleicher Zeile in der Reihenfolge von soll (die Reihenfolge der Prüfung, Nr. 2 d), die Reihenfolge von soll bleibt bei
// Gleichstand erhalten.
func erwarteReihenfolge(t *testing.T, name, errText string, soll []zeileText) {
	t.Helper()
	soll = append([]zeileText{}, soll...)
	sort.SliceStable(soll, func(i, j int) bool { return soll[i].zeile < soll[j].zeile })
	ist := befundZeilen(t, errText)
	if len(ist) != len(soll) {
		t.Fatalf("%s: %d Befunde, erwartet %d: %q", name, len(ist), len(soll), ist)
	}
	for i := range soll {
		if ist[i] != soll[i].text {
			t.Fatalf("%s: Befund %d ist %q, erwartet %q (Ausgabe %q)", name, i, ist[i], soll[i].text, ist)
		}
	}
}

// kopfMit baut ein Handbuch, dessen Kopf die Zeilen 1 bis 8 hat: Zeile 1 ist die Überschrift, nicht belegte Zeilen sind leer.
func kopfMit(belegt map[int]string) string {
	z := make([]string, 8)
	z[0] = "# H"
	for n, s := range belegt {
		z[n-1] = s
	}
	return strings.Join(z, "\n") + "\n\n## A\n"
}

func schluesselZeilen() map[string]string {
	return map[string]string{"Software-Version": "Software-Version: dev", "Gültigkeitsbereich": "Gültigkeitsbereich: Binary", "Stand": "Stand: 2026-10-30"}
}

// andere belegt die Zeilen 7 und 8 mit den Zeilen der beiden Schlüssel, die nicht s sind.
func andere(belegt map[int]string, s string) map[int]string {
	n := 7
	for _, k := range kopfSchluessel() {
		if k != s {
			belegt[n] = schluesselZeilen()[k]
			n++
		}
	}
	return belegt
}

// Z51: die Zeile eines Befundes des Kopfes ist die Zeile des Schlüssels, den er nennt, bei `doppelt` die des ersten
// Vorkommens, bei `fehlt` die Zeile 0 (Nr. 2 d). Menge: `doppelt` und `leer` je Schlüssel (6) und `Stand ungültig` (1),
// jeweils mit der Zeile ohne Schlüssel in Zeile 2 (vor dem ersten Vorkommen), 4 (zwischen den Vorkommen) und 6 (hinter
// allen): 21; dazu `fehlt` je Schlüssel (3). Der Schlüssel steht in Zeile 3, bei `doppelt` zusätzlich in Zeile 5.
func TestHandbuchKopfBefundZeile(t *testing.T) {
	const fremd = "Zeile ohne Schlüssel"
	type element struct {
		name   string
		s      string
		zeilen []string
		befund string
	}
	var elemente []element
	for _, s := range kopfSchluessel() {
		z := schluesselZeilen()[s]
		elemente = append(elemente,
			element{s + " doppelt", s, []string{z, z}, "Kopf: " + s + " doppelt"},
			element{s + " leer", s, []string{s + ":"}, "Kopf: " + s + " leer"})
	}
	elemente = append(elemente, element{"Stand ungültig", "Stand", []string{"Stand: 2026-02-30"}, "Kopf: Stand ungültig: 2026-02-30"})
	m := ohneBeispiele(t)
	n := 0
	for _, e := range elemente {
		for _, f := range []int{2, 4, 6} {
			belegt := map[int]string{3: e.zeilen[0], f: fremd}
			if len(e.zeilen) == 2 {
				belegt[5] = e.zeilen[1]
			}
			m.handbuch(kopfMit(andere(belegt, e.s)))
			code, _, errText := m.lauf()
			if code != exitBefund {
				t.Fatalf("%s, Fremdzeile in %d: Exit %d %q", e.name, f, code, errText)
			}
			erwarteReihenfolge(t, fmt.Sprintf("%s, Fremdzeile in %d", e.name, f), errText,
				[]zeileText{{3, e.befund}, {f, fmt.Sprintf("Zeile %d: Kopf: %s", f, fremd)}})
			n++
		}
	}
	for _, s := range kopfSchluessel() {
		m.handbuch(kopfMit(andere(map[int]string{3: fremd}, s)))
		code, _, errText := m.lauf()
		if code != exitBefund {
			t.Fatalf("%s fehlt: Exit %d %q", s, code, errText)
		}
		erwarteReihenfolge(t, s+" fehlt", errText, []zeileText{{0, "Kopf: " + s + " fehlt"}, {3, "Zeile 3: Kopf: " + fremd}})
		n++
	}
	if n != 24 {
		t.Fatalf("%d Elemente, erwartet 24", n)
	}
}

// Z52: `Software-Version … weicht … ab` hat die Zeile des Schlüssels; `version: Ausgabe nicht lesbar` hat sie, wenn der
// Schlüssel genau einmal und nicht leer steht, sonst die Zeile 0 (Nr. 2 d). Menge: fünf Befunde (einmal nicht leer:
// `weicht` und `nicht lesbar`; leer, doppelt, fehlt: `nicht lesbar`), je mit der Zeile ohne Schlüssel in Zeile 2 (vor
// dem Schlüssel) und hinter ihm (Zeile 5): 10 Läufe. Bei leerem, doppeltem und fehlendem Schlüssel entfällt `weicht`.
func TestHandbuchVersionBefundZeile(t *testing.T) {
	const fremd = "Zeile ohne Schlüssel"
	weicht := "Software-Version 1.0 weicht von der Ausgabe von version ab (2.0)"
	nl := "version: Ausgabe nicht lesbar: Exit 1"
	eins := "Software-Version: 1.0"
	type element struct {
		name    string
		zeilen  []string
		ausgabe antwort
		soll    []zeileText
	}
	elemente := []element{
		{"einmal, weicht ab", []string{eins}, antwort{out: "odata-recorder 2.0\n"}, []zeileText{{3, weicht}}},
		{"einmal, nicht lesbar", []string{eins}, antwort{out: "odata-recorder 1.0\n", code: 1}, []zeileText{{3, nl}}},
		{"leer, nicht lesbar", []string{"Software-Version:"}, antwort{out: "odata-recorder 1.0\n", code: 1}, []zeileText{{0, nl}, {3, "Kopf: Software-Version leer"}}},
		{"doppelt, nicht lesbar", []string{eins, eins}, antwort{out: "odata-recorder 1.0\n", code: 1}, []zeileText{{0, nl}, {3, "Kopf: Software-Version doppelt"}}},
		{"fehlt, nicht lesbar", nil, antwort{out: "odata-recorder 1.0\n", code: 1}, []zeileText{{0, "Kopf: Software-Version fehlt"}, {0, nl}}},
	}
	n := 0
	for _, e := range elemente {
		for _, f := range []int{2, 5} {
			m := neuesUmfeld(t, map[string]antwort{"version": e.ausgabe})
			m.entferne("docs/user/beispiele")
			belegt := map[int]string{f: fremd}
			for i, z := range e.zeilen {
				belegt[3+i] = z
			}
			m.handbuch(kopfMit(andere(belegt, "Software-Version")))
			code, _, errText := m.lauf()
			if code != exitBefund {
				t.Fatalf("%s, Fremdzeile in %d: Exit %d %q", e.name, f, code, errText)
			}
			soll := append(append([]zeileText{}, e.soll...), zeileText{f, fmt.Sprintf("Zeile %d: Kopf: %s", f, fremd)})
			erwarteReihenfolge(t, fmt.Sprintf("%s, Fremdzeile in %d", e.name, f), errText, soll)
			n++
		}
	}
	if n != 10 {
		t.Fatalf("%d Läufe, erwartet 10", n)
	}
}

// ersetzeStelle ersetzt das Byte an Stelle i von s durch c.
func ersetzeStelle(s string, i int, c byte) string {
	return s[:i] + string([]byte{c}) + s[i+1:]
}

// Z11 (Form): jede der zehn Stellen von `2026-10-30` durch ein Zeichen außerhalb der Form ersetzt. Ziffernstellen
// (0 bis 3, 5, 6, 8, 9; 8) × {`x`, `/` unter `0`, `:` über `9`} = 24; Strichstellen (4 und 7) × {`0`, `x`, `.`, `/`} = 8;
// Länge 9, 11 und 0 = 3. Dazu der Gutfall.
func TestHandbuchStandStellen(t *testing.T) {
	const gut = "2026-10-30"
	var faelle []standFall
	for _, i := range []int{0, 1, 2, 3, 5, 6, 8, 9} {
		for _, c := range []byte{'x', '/', ':'} {
			faelle = append(faelle, standFall{ersetzeStelle(gut, i, c), false})
		}
	}
	for _, i := range []int{4, 7} {
		for _, c := range []byte{'0', 'x', '.', '/'} {
			faelle = append(faelle, standFall{ersetzeStelle(gut, i, c), false})
		}
	}
	faelle = append(faelle, standFall{"2026-10-3", false}, standFall{"2026-10-300", false}, standFall{"", false})
	if len(faelle) != 35 {
		t.Fatalf("%d Elemente, erwartet 35", len(faelle))
	}
	pruefeStaende(t, append(faelle, standFall{gut, true}))
}

// grundFall ist eine Ausgabe von `version` mit den Gründen, die zugleich zutreffen, und dem ersten Grund der Reihenfolge
// `Signal`, `Exit <n>`, `Fehlerausgabe nicht leer`, `Standardausgabe` (Nr. 4 c).
type grundFall struct {
	signal, exit, err, std bool
	soll                   string
}

// Z10 (Reihenfolge): alle nichtleeren Kombinationen der gleichzeitig möglichen Gründe, 11 Elemente: ohne Signal Exit ≠ 0 ×
// Teilmenge von {Fehlerausgabe, Standardausgabe falsch} (4), mit Signal × Teilmenge (4), ohne Exit und Signal die drei
// nichtleeren Teilmengen.
func TestHandbuchVersionGruende(t *testing.T) {
	const exit3, sig, fehl, std = "Exit 3", "Signal", "Fehlerausgabe nicht leer", "Standardausgabe "
	faelle := []grundFall{
		{exit: true, soll: exit3}, {exit: true, err: true, soll: exit3}, {exit: true, std: true, soll: exit3}, {exit: true, err: true, std: true, soll: exit3},
		{signal: true, soll: sig}, {signal: true, err: true, soll: sig}, {signal: true, std: true, soll: sig}, {signal: true, err: true, std: true, soll: sig},
		{err: true, soll: fehl}, {std: true, soll: std}, {err: true, std: true, soll: fehl},
	}
	if len(faelle) != 11 {
		t.Fatalf("%d Elemente, erwartet 11", len(faelle))
	}
	for _, f := range faelle {
		a := antwort{out: "odata-recorder dev\n", signal: f.signal}
		if f.exit {
			a.code = 3
		}
		if f.err {
			a.err = "x\n"
		}
		if f.std {
			a.out = "falsch\n"
		}
		m := neuesUmfeld(t, map[string]antwort{"version": a})
		m.entferne("docs/user/beispiele")
		m.handbuch(rumpf("Text.\n"))
		code, _, errText := m.lauf()
		if code != exitBefund {
			t.Fatalf("%+v: Exit %d %q", f, code, errText)
		}
		zeilen := befundZeilen(t, errText)
		soll := "version: Ausgabe nicht lesbar: " + f.soll
		if len(zeilen) != 1 {
			t.Fatalf("%+v: %q, erwartet %q", f, zeilen, soll)
		}
		gleich := zeilen[0] == soll
		if f.soll == std {
			gleich = strings.HasPrefix(zeilen[0], soll)
		}
		if !gleich {
			t.Fatalf("%+v: %q, erwartet %q", f, zeilen, soll)
		}
	}
}

// Z8 (Form): die erste nicht leere Zeile des Kopfes beginnt mit `# `. Menge: `# T` (gültig), `#T`, `#`, `##T`, `#` und
// Tabulator und `T`, `#` und NBSP und `T`, ` # T`, `* T`, `T` (9), dazu `## T` (beendet den Kopf: Kopf leer, Zeile 1).
func TestHandbuchUeberschriftPraefix(t *testing.T) {
	keys := "Software-Version: dev\nGültigkeitsbereich: Binary\nStand: 2026-10-30\n"
	m := ohneBeispiele(t)
	formen := []string{"# T", "#T", "#", "##T", "#\tT", "#" + nbsp + "T", " # T", "* T", "T"}
	for i, f := range formen {
		m.handbuch("\n" + f + "\n" + keys + "\n## A\n")
		code, _, errText := m.lauf()
		if i == 0 {
			if code != exitOK {
				t.Fatalf("%q: Exit %d %q", f, code, errText)
			}
			continue
		}
		erwarteReihenfolge(t, fmt.Sprintf("%q", f), errText, []zeileText{{2, "Zeile 2: Kopf: Überschrift fehlt"}})
	}
	m.handbuch("\n## T\n" + keys + "\n## A\n")
	_, _, errText := m.lauf()
	erwarteReihenfolge(t, "## T", errText, []zeileText{
		{0, "Kopf: Software-Version fehlt"}, {0, "Kopf: Gültigkeitsbereich fehlt"}, {0, "Kopf: Stand fehlt"},
		{1, "Zeile 1: Kopf: Überschrift fehlt"}})
	if len(formen)+1 != 10 {
		t.Fatalf("%d Elemente, erwartet 10", len(formen)+1)
	}
}

// Z9 (Form): eine Zeile des Kopfes beendet ihn nur, wenn sie mit `## ` beginnt. Menge: `##`, `###`, `##x`, `##` und
// Tabulator, `#`, ` ## T`, `##` und NBSP und `T` (7) beenden nicht, der Schlüssel dahinter wird gelesen; `## T` und
// `## ` (2) beenden, der Schlüssel dahinter zählt nicht.
func TestHandbuchKopfEndeZeilen(t *testing.T) {
	keys := "Software-Version: dev\nGültigkeitsbereich: Binary\nStand: 2026-10-30\n"
	m := ohneBeispiele(t)
	nein := []string{"##", "###", "##x", "##\t", "#", " ## T", "##" + nbsp + "T"}
	for _, z := range nein {
		m.handbuch("# H\n" + z + "\n" + keys + "\n## A\n")
		_, _, errText := m.lauf()
		erwarteReihenfolge(t, fmt.Sprintf("%q", z), errText, []zeileText{{2, "Zeile 2: Kopf: Zeile ohne Schlüssel"}})
	}
	ja := []string{"## T", "## "}
	for _, z := range ja {
		m.handbuch("# H\n" + z + "\n" + keys + "\n## A\n")
		_, _, errText := m.lauf()
		erwarteReihenfolge(t, fmt.Sprintf("%q", z), errText, []zeileText{
			{0, "Kopf: Software-Version fehlt"}, {0, "Kopf: Gültigkeitsbereich fehlt"}, {0, "Kopf: Stand fehlt"}})
	}
	if len(nein)+len(ja) != 9 {
		t.Fatalf("%d Elemente, erwartet 9", len(nein)+len(ja))
	}
}

// Z28 (Zeichenfolgen): jede Stelle der Zeichenfolgen `-t odata-recorder:build` (23 Zeichen) und `-o /out/odata-recorder`
// (22 Zeichen) durch `X` ersetzt; die Zeichenfolge des Images in den beiden Quellen der Menge (Makefile und eine mk-Datei
// unter harness/mk), die des Pfades im Dockerfile; die jeweils andere Seite bleibt vollständig. Dazu drei Gutfälle, in
// denen die Zeichenfolge Teil einer längeren Zeile ist: 23 × 2 + 22 + 3 = 71.
func TestHandbuchAufrufpraefixStellen(t *testing.T) {
	const image, pfad = "-t odata-recorder:build", "-o /out/odata-recorder"
	if len(image) != 23 || len(pfad) != 22 {
		t.Fatalf("Längen %d und %d, erwartet 23 und 22", len(image), len(pfad))
	}
	m := neuesUmfeld(t, standardAntworten())
	n := 0
	setze := func(makefile, mk, dockerfile string) {
		m.entferne("harness/mk")
		m.schreibe("Makefile", makefile)
		if mk != "" {
			m.schreibe("harness/mk/x.mk", mk)
		}
		m.schreibe("Dockerfile", dockerfile)
	}
	docker := "RUN go build " + pfad + " ./x\n"
	for i := 0; i < len(image); i++ {
		falsch := ersetzeStelle(image, i, 'X')
		for _, quelle := range []string{"Makefile", "harness/mk/x.mk"} {
			if quelle == "Makefile" {
				setze("build: ## x\n\tdocker build "+falsch+" .\n", "", docker)
			} else {
				setze("build: ## x\n", "\tdocker build "+falsch+" .\n", docker)
			}
			code, _, errText := m.lauf()
			erwarte(t, code, errText, exitBefund, "Aufrufpräfix nicht gedeckt: Image")
			nichtEnthalten(t, errText, "gedeckt: Pfad")
			n++
		}
	}
	for i := 0; i < len(pfad); i++ {
		setze("build: ## x\n\tdocker build "+image+" .\n", "", "RUN go build "+ersetzeStelle(pfad, i, 'X')+" ./x\n")
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "Aufrufpräfix nicht gedeckt: Pfad")
		nichtEnthalten(t, errText, "gedeckt: Image")
		n++
	}
	gut := [][3]string{
		{"build: ## x\n\tdocker build --target build " + image + " .\n", "", docker},
		{"build: ## x\n", "\tdocker build --target build " + image + " .\n", docker},
		{"build: ## x\n\tdocker build " + image + " .\n", "", "RUN go build -v " + pfad + " -trimpath ./x\n"},
	}
	for _, g := range gut {
		setze(g[0], g[1], g[2])
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("Gutfall %q: Exit %d %q", g, code, errText)
		}
		n++
	}
	if n != 71 {
		t.Fatalf("%d Elemente, erwartet 71", n)
	}
}

// Z21 (UTF-8): die Beispiel-Datei ist UTF-8. Menge: fünf Bytefolgen (0xFF, überlang 0xC0 0x80, Surrogat 0xED 0xA0 0x80,
// über U+10FFFF 0xF4 0x90 0x80 0x80, abgeschnitten 0xC3 am Zeilenende) × Abschnitt {Standardausgabe, Fehlerausgabe} = 10,
// aus Bytewerten gebildet. Die Datei ist ungültig und wird nicht gefahren (ein Start: `version`).
func TestHandbuchBeispielUTF8(t *testing.T) {
	folgen := [][]byte{{0xFF}, {0xC0, 0x80}, {0xED, 0xA0, 0x80}, {0xF4, 0x90, 0x80, 0x80}, {0xC3}}
	n := 0
	for _, f := range folgen {
		if utf8.Valid(f) {
			t.Fatalf("% X ist gültiges UTF-8", f)
		}
		for abschnitt := 0; abschnitt < 2; abschnitt++ {
			var zeilen [2][]string
			zeilen[abschnitt] = []string{"a" + string(f)}
			datei := bspDatei(praef, zeilen[0], zeilen[1], "0\n")
			if utf8.ValidString(datei) {
				t.Fatalf("% X: die Datei ist gültiges UTF-8", f)
			}
			m := neuesUmfeld(t, okAntworten())
			m.entferne("docs/user/beispiele/aufruf.txt")
			m.schreibe("docs/user/beispiele/k.txt", datei)
			m.handbuch(rumpf(binde("beispiele/k.txt") + zaunVon(bspDatei(praef, nil, nil, "0\n"))))
			code, _, errText := m.lauf()
			erwarte(t, code, errText, exitBefund, "k.txt: Beispiel-Datei ungültig: ")
			if s := len(m.e.starts(t)); s != 1 {
				t.Fatalf("% X, Abschnitt %d: %d Starts, erwartet 1 (version)", f, abschnitt, s)
			}
			n++
		}
	}
	if n != 10 {
		t.Fatalf("%d Elemente, erwartet 10", n)
	}
}

// Z6 (je Zeile und Zeichen einmal): 9 Zeichen aus Z6 × {dasselbe Zeichen zweimal in einer Zeile: 1 Befund; dasselbe in zwei
// Zeilen: 2 Befunde; zwei verschiedene in einer Zeile: 2 Befunde, in der Reihenfolge ihrer ersten Stelle} = 27.
func TestHandbuchZeichenJeZeileEinmal(t *testing.T) {
	zeichen := []rune{0x01, 0x7F, 0x200B, 0xAD, 0xFEFF, 0xE000, 0x301, 0x20DD, 0x202E}
	m := neuesUmfeld(t, standardAntworten())
	erste := naechste(sauber())
	meldung := func(zeile int, r rune) string {
		return fmt.Sprintf("Zeile %d: Zeichen U+%04X nicht zulässig", zeile, r)
	}
	n := 0
	for i, c := range zeichen {
		p := zeichen[(i+1)%len(zeichen)]
		faelle := []struct {
			prosa string
			soll  []string
		}{
			{"Wort" + string(c) + "Wort" + string(c) + "Ende.\n", []string{meldung(erste, c)}},
			{"Wort" + string(c) + "Wort.\nWort" + string(c) + "Wort.\n", []string{meldung(erste, c), meldung(erste+1, c)}},
			{"Wort" + string(c) + "Wort" + string(p) + "Ende.\n", []string{meldung(erste, c), meldung(erste, p)}},
		}
		for _, f := range faelle {
			m.handbuch(sauber() + "\n" + f.prosa)
			code, _, errText := m.lauf()
			if code != exitBefund {
				t.Fatalf("U+%04X %q: Exit %d %q", c, f.prosa, code, errText)
			}
			if ist := befundZeilen(t, errText); strings.Join(ist, "\n") != strings.Join(f.soll, "\n") {
				t.Fatalf("U+%04X %q: %q, erwartet %q", c, f.prosa, ist, f.soll)
			}
			n++
		}
	}
	if n != 27 {
		t.Fatalf("%d Elemente, erwartet 27", n)
	}
}

// Z53: der Text eines Spans verliert ein Leerzeichen nur, wenn beide Enden eines haben und der Inhalt nicht nur aus
// Leerzeichen besteht; sonst bleibt es (Nr. 9 b, 14 d). Menge: Anfang {kein, ein Leerzeichen, zwei Leerzeichen,
// Zeilenumbruch} × Ende (dieselben vier) × Inhalt {`make nix`, `make`, ein Leerzeichen, zwei Leerzeichen} = 64, je Element
// ein Absatz mit dem Span allein. Orakel von Hand: ein Aufruf mit unbekanntem Ziel `nix` wird gemeldet, wenn der Text nach
// der Regel mit `make ` beginnt (Zeile der Tabelle: Anfang, Spalte: Ende); jeder andere Inhalt gibt keinen Befund.
func TestHandbuchSpanLeerzeichenAnDenEnden(t *testing.T) {
	enden := []string{"", " ", "  ", "\n"}
	geprueft := [4][4]bool{
		{true, true, true, true},     // Anfang kein: keines zu entfernen, `make nix` beginnt mit `make `
		{false, true, true, true},    // Anfang ein Leerzeichen: nur als Paar entfernt
		{false, false, false, false}, // Anfang zwei: nach dem Paar bleibt ein führendes Leerzeichen
		{false, true, true, true},    // Anfang Zeilenumbruch: zählt als ein Leerzeichen
	}
	m := ohneBeispiele(t)
	zielDateien(m)
	n := 0
	for ia, a := range enden {
		for ie, e := range enden {
			for _, kern := range []string{"make nix", "make", " ", "  "} {
				m.handbuch(rumpf("`" + a + kern + e + "`\n"))
				code, _, errText := m.lauf()
				if kern == "make nix" && geprueft[ia][ie] {
					erwarte(t, code, errText, exitBefund, "make-Ziel nix unbekannt")
				} else if code != exitOK {
					t.Fatalf("Anfang %q, Ende %q, Inhalt %q: Exit %d %q", a, e, kern, code, errText)
				}
				n++
			}
		}
	}
	if n != 64 {
		t.Fatalf("%d Elemente, erwartet 64", n)
	}
}
