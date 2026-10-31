package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	sepStd = "--- Standardausgabe"
	sepErr = "--- Fehlerausgabe"
	sepEx  = "--- Ausgangswert"
	praef  = "$ docker run --rm odata-recorder:build /out/odata-recorder"
)

func ohneBeispiele(t *testing.T) umfeld {
	t.Helper()
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele")
	m.handbuch(rumpf("Text.\n"))
	return m
}

func rumpf(h string) string {
	return kopfSauber + "\n## A\n\n" + h
}

// Z12: nur drei Backticks ohne Einrückung sind ein Zaun; jede andere Form ist Prosa.
func TestHandbuchZaunformen(t *testing.T) {
	m := ohneBeispiele(t)
	faelle := [][2]string{{"~~~", "~~~"}, {"````", "````"}, {" ```", " ```"}, {"  ```", "  ```"}, {"   ```", "   ```"}, {"```a`b", "text"}}
	for _, f := range faelle {
		m.handbuch(rumpf(f[0] + "\ngeplant\n" + f[1] + "\n"))
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, trefferZeile("A1")+"geplant")
		nichtEnthalten(t, errText, "Zaun", "Bindungszeile", "Info-Zeichenkette")
	}
}

// Z13: ein Zaun ohne Schluss ist Befund, der Rest ist Prosa.
func TestHandbuchZaunOhneSchluss(t *testing.T) {
	m := ohneBeispiele(t)
	m.handbuch(rumpf("```\nbisher\n"))
	code, _, errText := m.lauf()
	erwarteN(t, code, errText, exitBefund, "Zeile 9: Zaun nicht geschlossen", trefferZeile("B1")+"bisher")
}

// Z14: die Info-Zeichenkette ist leer.
func TestHandbuchInfo(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	for _, info := range []string{"text", "console", "sh", ""} {
		m.handbuch(rumpf(binde("beispiele/aufruf.txt") + "```" + info + "\n" + aufrufDatei + "```\n"))
		code, _, errText := m.lauf()
		if info == "" {
			if code != exitOK {
				t.Fatalf("Exit %d %q", code, errText)
			}
			continue
		}
		erwarte(t, code, errText, exitBefund, "Info-Zeichenkette "+info+" nicht zulässig")
	}
}

// Z15: Form der Bindungszeile; Generator über Verstöße, Orakel: die Form als Liste.
func TestHandbuchBindungszeile(t *testing.T) {
	zaun := zaunVon(aufrufDatei)
	faelle := []struct{ vor, text string }{
		{binde("beispiele/aufruf.txt"), ""},
		{"Eine Zeile.\n", "Beispiel ohne Bindungszeile"},
		{binde("beispiele/aufruf.txt") + "\n", "Beispiel ohne Bindungszeile"},
		{"Datei: [beispiele/aufruf.txt](beispiele/anderes.txt)\n", "Bindung ungültig: Datei: [beispiele/aufruf.txt](beispiele/anderes.txt)"},
		{binde("beispiele/../x.txt"), "Bindung ungültig: "},
		{binde("/x.txt"), "Bindung ungültig: "},
		{binde("beispiele/X.txt"), "Bindung ungültig: "},
		{binde("beispiele/a_b.txt"), "Bindung ungültig: "},
		{binde("beispiele/a--b.txt"), "Bindung ungültig: "},
		{binde("beispiele/.txt"), "Bindung ungültig: "},
		{binde("beispiele/a.md"), "Bindung ungültig: "},
		{binde("beispiele/u/a.txt"), "Bindung ungültig: "},
	}
	if len(faelle) != 12 {
		t.Fatalf("%d Elemente, erwartet 12", len(faelle))
	}
	m := neuesUmfeld(t, standardAntworten())
	for _, f := range faelle {
		m.handbuch(rumpf(f.vor + zaun))
		code, _, errText := m.lauf()
		if f.text == "" {
			if code != exitOK {
				t.Fatalf("Gutfall: Exit %d %q", code, errText)
			}
			continue
		}
		erwarte(t, code, errText, exitBefund, f.text)
	}
	m.handbuch(rumpf(binde("beispiele/aufruf.txt") + "\n" + zaun))
	_, _, errText := m.lauf()
	enthaelt2(t, errText, "Bindungszeile ohne Zaun")
}

// Z16: eine Bindungszeile ohne öffnenden Zaun unmittelbar danach ist Befund.
func TestHandbuchBindungOhneZaun(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	b := binde("beispiele/aufruf.txt")
	for _, h := range []string{b, b + "Prosa\n", b + "\n" + zaunVon(aufrufDatei)} {
		m.handbuch(kopfSauber + "\n## A\n\n" + h)
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "Bindungszeile ohne Zaun")
	}
}

// Z17: der Block ist bytegleich der Datei.
func TestHandbuchBlockGleichDatei(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	b := binde("beispiele/aufruf.txt")
	abw := map[string]string{
		"Zeile anders":        strings.Replace(aufrufDatei, "odata-recorder dev", "odata-recorder dex", 1),
		"Zeile mehr":          aufrufDatei + "x\n",
		"Zeile weniger":       strings.TrimSuffix(aufrufDatei, "0\n"),
		"Leerzeile am Ende":   aufrufDatei + "\n",
		"Leerzeichen am Ende": strings.TrimSuffix(aufrufDatei, "\n") + " \n",
		"Kleinschreibung":     strings.Replace(aufrufDatei, "--- Standardausgabe", "--- standardausgabe", 1),
	}
	for name, block := range abw {
		m.handbuch(rumpf(b + zaunVon(block)))
		code, _, errText := m.lauf()
		if code != exitBefund || !strings.Contains(errText, "Zeile 10: Beispiel weicht von beispiele/aufruf.txt ab") {
			t.Fatalf("%s: Exit %d %q", name, code, errText)
		}
	}
	m.handbuch(rumpf(binde("beispiele/gibtsnicht.txt") + zaunVon(aufrufDatei)))
	_, _, errText := m.lauf()
	enthaelt2(t, errText, "Zeile 10: beispiele/gibtsnicht.txt fehlt")
	gleich := []string{
		rumpf(b + zaunVon(aufrufDatei)),
		strings.ReplaceAll(rumpf(b+zaunVon(aufrufDatei)), "\n", "\r\n"),
		rumpf(b + zaunVon(aufrufDatei) + "\n" + b + zaunVon(aufrufDatei)),
	}
	for _, h := range gleich {
		m.handbuch(h)
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("gleicher Block: Exit %d %q", code, errText)
		}
	}
}

// Z18: verwaiste Einträge in docs/user/beispiele.
func TestHandbuchVerwaiste(t *testing.T) {
	faelle := []struct{ eintrag, text string }{
		{"extra.txt", "beispiele: extra.txt ohne Bindung"},
		{"sub/", "beispiele: sub keine Beispiel-Datei"},
		{"ordner.txt/", "beispiele: ordner.txt keine Beispiel-Datei"},
		{"notiz.md", "beispiele: notiz.md keine Beispiel-Datei"},
		{"A_b.txt", "beispiele: A_b.txt keine Beispiel-Datei"},
	}
	for _, f := range faelle {
		m := neuesUmfeld(t, standardAntworten())
		if strings.HasSuffix(f.eintrag, "/") {
			if err := os.Mkdir(filepath.Join(m.wurzel, "docs/user/beispiele", strings.TrimSuffix(f.eintrag, "/")), 0o755); err != nil {
				t.Fatal(err)
			}
		} else {
			m.schreibe("docs/user/beispiele/"+f.eintrag, "x\n")
		}
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, f.text)
	}
	m := neuesUmfeld(t, standardAntworten())
	b := binde("beispiele/aufruf.txt")
	m.handbuch(rumpf(b + zaunVon(aufrufDatei) + "\n" + b + zaunVon(aufrufDatei)))
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("zwei Blöcke einer Datei: %d %q", code, errText)
	}
	m = ohneBeispiele(t)
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("ohne Verzeichnis: %d %q", code, errText)
	}
}

// bspUmfeld legt die Beispiel-Datei name an und bindet sie mit gleichem Block.
func bspUmfeld(t *testing.T, name, inhalt string, a map[string]antwort) umfeld {
	t.Helper()
	m := neuesUmfeld(t, a)
	m.entferne("docs/user/beispiele/aufruf.txt")
	m.schreibe("docs/user/beispiele/"+name, inhalt)
	if !strings.HasSuffix(inhalt, "\n") {
		inhalt += "\n"
	}
	m.handbuch(rumpf(binde("beispiele/"+name) + zaunVon(inhalt)))
	return m
}

func bspDatei(kopf string, out, err []string, ex string) string {
	return strings.Join(append(append(append([]string{kopf, sepStd}, out...), append([]string{sepErr}, err...)...), sepEx), "\n") + "\n" + ex
}

func okAntworten() map[string]antwort {
	return map[string]antwort{"version": {out: "odata-recorder dev\n"}, "": {}}
}

func erwarteUngueltig(t *testing.T, name, inhalt string) {
	t.Helper()
	m := bspUmfeld(t, name, inhalt, okAntworten())
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitBefund, name+": Beispiel-Datei ungültig: ")
	if n := len(m.e.starts(t)); n != 1 {
		t.Fatalf("%q: %d Starts, erwartet 1 (version)", inhalt, n)
	}
}

// Z19: Kopfzeile der Beispiel-Datei; Generator über Verstöße am Präfix und am Argument,
// Orakel: der Zeichensatz als Liste.
func TestHandbuchBeispielKopfzeile(t *testing.T) {
	koepfe := []string{
		"$ docker run odata-recorder:build /out/odata-recorder",
		"$ docker run --rm odata-recorder:test /out/odata-recorder",
		"$ docker run --rm odata-recorder:build /out/andere",
		"docker run --rm odata-recorder:build /out/odata-recorder",
		praef + ` "x"`, praef + " $x", praef + " a  b", praef + " a ", praef + " ",
	}
	for _, c := range []string{"/", ":", ",", ";", "*", "?", "~", "+", "'", "\\", "(", "%"} {
		koepfe = append(koepfe, praef+" a"+c+"b")
	}
	if len(koepfe) != 21 {
		t.Fatalf("%d Elemente, erwartet 21", len(koepfe))
	}
	for _, k := range koepfe {
		erwarteUngueltig(t, "k.txt", bspDatei(k, nil, nil, "0\n"))
	}
	for _, rest := range []string{"X", "Xa", "-x"} {
		erwarteUngueltig(t, "k.txt", bspDatei(praef+rest, nil, nil, "0\n"))
	}
	gut := []struct {
		kopf string
		out  []string
	}{{praef, nil}, {praef + " version", []string{"odata-recorder dev"}}, {praef + " a.b_c=D-1", nil}}
	for _, g := range gut {
		m := bspUmfeld(t, "k.txt", bspDatei(g.kopf, g.out, nil, "0\n"), map[string]antwort{"version": okAntworten()["version"], "": {}, "a.b_c=D-1": {}})
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("Gutfall %q: Exit %d %q", g.kopf, code, errText)
		}
	}
}

// Z19 (Rand): jedes ASCII-Zeichen im Argument gegen den Zeichensatz A-Z a-z 0-9 . _ = -.
func TestHandbuchBeispielZeichensatz(t *testing.T) {
	erlaubt := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._=-"
	for c := 0x21; c < 0x7f; c++ {
		inhalt := bspDatei(praef+" a"+string(rune(c))+"b", nil, nil, "0\n")
		m := bspUmfeld(t, "k.txt", inhalt, map[string]antwort{"version": okAntworten()["version"]})
		_, _, errText := m.lauf()
		ungueltig := strings.Contains(errText, "Beispiel-Datei ungültig")
		if ungueltig == strings.ContainsRune(erlaubt, rune(c)) {
			t.Fatalf("Zeichen %q: ungültig=%v", rune(c), ungueltig)
		}
	}
	for _, c := range []string{"ä", "\t"} {
		erwarteUngueltig(t, "k.txt", bspDatei(praef+" a"+c+"b", nil, nil, "0\n"))
	}
}

// Z20: Abschnitte der Beispiel-Datei; Generator über fehlende, doppelte und vertauschte Trennzeilen,
// Orakel: die Grammatik als Tabelle.
func TestHandbuchBeispielAbschnitte(t *testing.T) {
	seps := []string{sepStd, sepErr, sepEx}
	basis := func() []string { return []string{praef, sepStd, sepErr, sepEx, "0"} }
	type fall struct {
		zeilen []string
		text   string
	}
	var faelle []fall
	add := func(z []string, text string) {
		faelle = append(faelle, fall{z, text})
	}
	for _, s := range seps {
		var z []string
		for _, l := range basis() {
			if l != s {
				z = append(z, l)
			}
		}
		add(z, "Trennzeile fehlt: "+s)
		add(append(basis(), s), "Trennzeile doppelt: "+s)
	}
	for _, p := range [][2]int{{1, 2}, {1, 3}, {2, 3}} {
		z := basis()
		z[p[0]], z[p[1]] = z[p[1]], z[p[0]]
		add(z, "Trennzeilen nicht in der Reihenfolge")
	}
	if len(faelle) != 9 {
		t.Fatalf("%d Fehlerfälle, erwartet 9", len(faelle))
	}
	for _, f := range faelle {
		m := bspUmfeld(t, "k.txt", strings.Join(f.zeilen, "\n")+"\n", okAntworten())
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "k.txt: Beispiel-Datei ungültig: "+f.text)
	}
	erwarteUngueltig(t, "k.txt", strings.Join([]string{praef, "vorweg", sepStd, sepErr, sepEx, "0"}, "\n")+"\n")
	erwarteUngueltig(t, "k.txt", bspDatei(praef, []string{"a\rb"}, nil, "0\n"))
	gut := []struct {
		out, err []string
		a        antwort
	}{{nil, nil, antwort{}}, {[]string{"a"}, nil, antwort{out: "a\n"}}, {nil, []string{"b"}, antwort{err: "b\n"}}}
	for _, g := range gut {
		m := bspUmfeld(t, "k.txt", bspDatei(praef, g.out, g.err, "0\n"), map[string]antwort{"version": okAntworten()["version"], "": g.a})
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("Gutfall %v %v: Exit %d %q", g.out, g.err, code, errText)
		}
	}
}

// Z21: Ausgangswert 0 bis 255 ohne führende Null, kein \r, Schluss-LF; die Datei wird nicht gefahren.
func TestHandbuchBeispielAusgangswert(t *testing.T) {
	for _, n := range []int{0, 1, 255} {
		m := bspUmfeld(t, "k.txt", bspDatei(praef, nil, nil, fmt.Sprint(n, "\n")),
			map[string]antwort{"version": okAntworten()["version"], "": {code: n}})
		code, _, errText := m.lauf()
		if code != exitOK || len(m.e.starts(t)) != 2 {
			t.Fatalf("Ausgangswert %d: Exit %d %q, %d Starts", n, code, errText, len(m.e.starts(t)))
		}
	}
	ungueltig := []string{"256\n", "-1\n", "01\n", "", "2a\n", "2 \n", "2\r\n", "2", "2 x\n", "2\nx\n"}
	if len(ungueltig) != 10 {
		t.Fatalf("%d Elemente, erwartet 10", len(ungueltig))
	}
	for _, u := range ungueltig {
		erwarteUngueltig(t, "k.txt", bspDatei(praef, nil, nil, u))
	}
}
