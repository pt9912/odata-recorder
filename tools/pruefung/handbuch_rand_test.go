package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
)

// Tests zu den Randformen, die SPEC-200 nach der Entscheidung des Architect trägt.

// Z4: Binary und Handbuch fehlen zugleich, beide stehen in der Ausgabe in der Reihenfolge der Prüfung;
// ein Arbeitsverzeichnis, das sich nicht anlegen lässt, ist ein eigener Eingabefehler.
func TestHandbuchMehrereEingabefehler(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne(handbuchPfad)
	code, out, errText := fuehre(m.u, "handbuch", filepath.Join(m.wurzel, "fehlt"), m.wurzel)
	erwarteN(t, code, errText, exitAufruf, "Eingabe: Binary fehlt: ", "Eingabe: Handbuch fehlt: ")
	z := strings.Split(strings.TrimSuffix(errText, "\n"), "\n")
	if out != "" || len(z) != 2 || !strings.Contains(z[0], "Binary fehlt") || !strings.Contains(z[1], "Handbuch fehlt") {
		t.Fatalf("stdout %q, Ausgabe %q", out, errText)
	}
	m = neuesUmfeld(t, standardAntworten())
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "gibt-es-nicht"))
	code, out, errText = m.lauf()
	erwarte(t, code, errText, exitAufruf, "Eingabe: Arbeitsverzeichnis nicht anlegbar: ")
	nichtEnthalten(t, errText, "Binary fehlt", "Binary nicht startbar")
	if n := len(m.e.starts(t)); out != "" || n != 0 || strings.Count(errText, "\n") != 1 {
		t.Fatalf("stdout %q, %d Starts, Ausgabe %q", out, n, errText)
	}
}

// Z8: die erste nicht leere Zeile des Kopfes ist der Platz der Überschrift.
func TestHandbuchKopfUeberschriftRand(t *testing.T) {
	m := ohneBeispiele(t)
	m.handbuch("Software-Version: x\nGültigkeitsbereich: y\nStand: 2026-10-30\n\n## A\n")
	code, _, errText := m.lauf()
	erwarteN(t, code, errText, exitBefund, "Zeile 1: Kopf: Überschrift fehlt", "Kopf: Software-Version fehlt")
	m.handbuch("\n\n## A\n")
	code, _, errText = m.lauf()
	erwarteN(t, code, errText, exitBefund, "Zeile 1: Kopf: Überschrift fehlt", "Kopf: Stand fehlt")
	m.handbuch("\n\n# H\n\nSoftware-Version: dev\nGültigkeitsbereich: y\nStand: 2026-10-30\n\n## A\n")
	if code, _, errText = m.lauf(); code != exitOK {
		t.Fatalf("Leerzeilen vor der Überschrift: Exit %d %q", code, errText)
	}
}

// Z10: fehlt, leer oder doppelt der Schlüssel Software-Version, steht nur der Befund nach Nr. 4 b, es gibt
// keinen Vergleich, und version läuft einmal.
func TestHandbuchVersionKopfLuecken(t *testing.T) {
	rest := "Gültigkeitsbereich: Binary\nStand: 2026-10-30\n\n## A\n"
	faelle := []struct{ name, kopf, text string }{
		{"fehlt", "# H\n\n" + rest, "Kopf: Software-Version fehlt"},
		{"leer", "# H\n\nSoftware-Version:\n" + rest, "Kopf: Software-Version leer"},
		{"doppelt", "# H\n\nSoftware-Version: 1.0\nSoftware-Version: 2.0\n" + rest, "Kopf: Software-Version doppelt"},
	}
	for _, f := range faelle {
		m := ohneBeispiele(t)
		m.handbuch(f.kopf)
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, f.text)
		nichtEnthalten(t, errText, "weicht von der Ausgabe")
		if s := m.e.starts(t); len(s) != 1 || !gleich(s[0].args, []string{"version"}) {
			t.Fatalf("%s: Starts %+v", f.name, s)
		}
	}
}

// Z13: ein Zaun ohne Schluss nimmt den Rest; für ihn gilt nur der eine Befund.
func TestHandbuchZaunOhneSchlussRand(t *testing.T) {
	m := ohneBeispiele(t)
	m.handbuch(rumpf("```text\nbisher\n"))
	code, _, errText := m.lauf()
	erwarteN(t, code, errText, exitBefund, "Zeile 9: Zaun nicht geschlossen", trefferZeile("B1")+"bisher")
	nichtEnthalten(t, errText, "Info-Zeichenkette", "Bindungszeile")
	m.handbuch(rumpf("```\nbisher\n```text\n"))
	_, _, errText = m.lauf()
	if c := strings.Count(errText, "Zaun nicht geschlossen"); c != 1 {
		t.Fatalf("%d Befunde Zaun nicht geschlossen: %q", c, errText)
	}
	m.handbuch(rumpf("```\nDatei: [beispiele/aufruf.txt](beispiele/aufruf.txt)\n"))
	code, _, errText = m.lauf()
	erwarteN(t, code, errText, exitBefund, "Zeile 10: Bindungszeile ohne Zaun")
}

// Z14: eine Info-Zeichenkette nur aus Leerraum ist leer (ein Leerzeichen, drei, Tabulator, NBSP aus Bytewerten).
func TestHandbuchInfoLeerraum(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	for _, info := range []string{" ", "   ", "\t", "\xc2\xa0"} {
		m.handbuch(rumpf(binde("beispiele/aufruf.txt") + "```" + info + "\n" + aufrufDatei + "```\n"))
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("Info %q: Exit %d %q", info, code, errText)
		}
	}
	m.handbuch(rumpf(binde("beispiele/aufruf.txt") + "``` text \n" + aufrufDatei + "```\n"))
	_, _, errText := m.lauf()
	enthaelt2(t, errText, "Info-Zeichenkette text nicht zulässig")
}

// Z20: eine Zeile zwischen Kopfzeile und erster Trennzeile, auch eine leere, ist ein Verstoß; mit den neun
// Fehlerfällen und den drei Gutfällen von TestHandbuchBeispielAbschnitte sind das 14 Elemente.
func TestHandbuchBeispielZeileVorTrennzeile(t *testing.T) {
	for _, zw := range []string{"vorweg", ""} {
		erwarteUngueltig(t, "k.txt", strings.Join([]string{praef, zw, sepStd, sepErr, sepEx, "0"}, "\n")+"\n")
	}
}

// Z35: die Leerraum-Menge von Nr. 13 a als Schleife an der Form `noch nicht`: die 16 Zeichen der Kategorie Zs
// außer dem Leerzeichen, U+2028, U+2029 und der Tabulator; Orakel: die Codepunkte, abgeglichen mit unicode.Zs.
func TestHandbuchLeerraumMenge(t *testing.T) {
	cps := []rune{0xA0, 0x1680}
	for c := 0x2000; c <= 0x200A; c++ {
		cps = append(cps, rune(c))
	}
	cps = append(cps, 0x202F, 0x205F, 0x3000, 0x2028, 0x2029, '\t')
	if len(cps) != 19 {
		t.Fatalf("%d Zeichen, erwartet 19", len(cps))
	}
	zs := 0
	for _, r := range unicode.Zs.R16 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			if c != ' ' {
				zs++
			}
		}
	}
	for _, r := range unicode.Zs.R32 {
		for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
			zs++
		}
	}
	if zs != 16 {
		t.Fatalf("unicode.Zs trägt %d Zeichen außer dem Leerzeichen, erwartet 16", zs)
	}
	m := ohneBeispiele(t)
	for _, c := range cps {
		m.handbuch(rumpf("Das ist noch" + string(c) + "nicht hier.\n"))
		code, _, errText := m.lauf()
		if code != exitBefund || !strings.Contains(errText, trefferZeile("A11")) {
			t.Fatalf("U+%04X: Exit %d %q", c, code, errText)
		}
	}
}

// Nr. 13 a und 14 c: das Entfernen von `*`, `_` und Backtick gilt wörtlich: hervorgehobene Wörter treffen,
// ein Bezeichner, der mit `_` getrennt ist, nicht.
func TestHandbuchMarkupWoertlich(t *testing.T) {
	m := ohneBeispiele(t)
	for _, prosa := range []string{"Das ist _geplant_ hier.", "Das ist **geplant** hier.", "Das ist `geplant` hier.", "Das ist ge_plant hier."} {
		m.handbuch(rumpf(prosa + "\n"))
		if code, _, errText := m.lauf(); code != exitBefund || !strings.Contains(errText, trefferZeile("A1")) {
			t.Fatalf("%q: Exit %d %q", prosa, code, errText)
		}
	}
	m.handbuch(rumpf("Das ist noch_nicht hier.\n"))
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("noch_nicht: Exit %d %q", code, errText)
	}
}

// Z41: Kürzung auf die ersten 200 Zeichen der gequoteten Zeichenkette ohne Auslassungszeichen, in erwartet
// und erhalten, in beiden Strömen und für version; Orakel: die Regel, nicht der Code (17 Elemente).
func TestHandbuchKuerzung200(t *testing.T) {
	gekuerzt := func(s string) string {
		r := []rune(strconv.Quote(s))
		if len(r) > 200 {
			r = r[:200]
		}
		return string(r)
	}
	n := 0
	for _, l := range []int{199, 200, 201} {
		lang := strings.Repeat("a", l-4) + "\n"
		if got := len([]rune(strconv.Quote(lang))); got != l {
			t.Fatalf("Fixture: Länge %d statt %d", got, l)
		}
		for _, strom := range []string{"Standardausgabe", "Fehlerausgabe"} {
			for _, seite := range []string{"erwartet", "erhalten"} {
				erw, erh := "x\n", "x\n"
				if seite == "erwartet" {
					erw = lang
				} else {
					erh = lang
				}
				var out, errA []string
				var a antwort
				if strom == "Standardausgabe" {
					out, a = []string{strings.TrimSuffix(erw, "\n")}, antwort{out: erh}
				} else {
					errA, a = []string{strings.TrimSuffix(erw, "\n")}, antwort{err: erh}
				}
				m := bspMit(t, nil, out, errA, "0\n", a)
				_, _, errText := m.lauf()
				want := fmt.Sprintf("k.txt: %s weicht ab (erwartet %s, erhalten %s)", strom, gekuerzt(erw), gekuerzt(erh))
				enthaelt2(t, errText, want)
				n++
			}
		}
		m := neuesUmfeld(t, map[string]antwort{"version": {out: lang}, "": {out: "odata-recorder dev\n"}})
		_, _, errText := m.lauf()
		enthaelt2(t, errText, "version: Ausgabe nicht lesbar: Standardausgabe "+gekuerzt(lang)+"\n")
		n++
	}
	ue := strings.Repeat("\xc3\xbc", 200) + "\n"
	for _, seite := range []string{"erwartet", "erhalten"} {
		file, ist := "x", ue
		if seite == "erwartet" {
			file, ist = strings.TrimSuffix(ue, "\n"), "x\n"
		}
		m := bspMit(t, nil, []string{file}, nil, "0\n", antwort{out: ist})
		_, _, errText := m.lauf()
		erw, erh := "x\n", ue
		if seite == "erwartet" {
			erw, erh = ue, "x\n"
		}
		enthaelt2(t, errText, fmt.Sprintf("(erwartet %s, erhalten %s)", gekuerzt(erw), gekuerzt(erh)))
		n++
	}
	if n != 17 {
		t.Fatalf("%d Elemente, erwartet 17", n)
	}
}

// Z42: Befunde der Nr. 6 d, 7 c, 7 d beginnen mit dem Dateinamen samt .txt, die der Aufrufstelle version mit
// version (5 Elemente).
func TestHandbuchBefundName(t *testing.T) {
	const name = "aufruf-version.txt"
	m := bspUmfeld(t, name, bspDatei(praef, nil, nil, "256\n"), okAntworten())
	_, _, errText := m.lauf()
	enthaelt2(t, errText, "pruefung handbuch: "+name+": Beispiel-Datei ungültig: ")
	for _, stelle := range []string{"version", "Beispiel"} {
		a := okAntworten()
		text := name + ": Zeitgrenze überschritten"
		if stelle == "version" {
			a["version"] = antwort{schlaf: "5"}
			text = "version: Zeitgrenze überschritten"
		} else {
			a[""] = antwort{schlaf: "5"}
		}
		m := bspUmfeld(t, name, bspDatei(praef, nil, nil, "0\n"), a)
		m.u.Zeitgrenze = 200 * time.Millisecond
		_, _, errText := m.lauf()
		enthaelt2(t, errText, "pruefung handbuch: "+text)
	}
	a := okAntworten()
	a[""] = antwort{out: "x\n"}
	m = bspUmfeld(t, name, bspDatei(praef, nil, nil, "0\n"), a)
	_, _, errText = m.lauf()
	enthaelt2(t, errText, "pruefung handbuch: "+name+": Standardausgabe weicht ab")
	m = neuesUmfeld(t, map[string]antwort{"version": {out: "x\n"}, "": {out: "odata-recorder dev\n"}})
	_, _, errText = m.lauf()
	enthaelt2(t, errText, "pruefung handbuch: version: Ausgabe nicht lesbar: ")
}

// Z43: eine Beispiel-Datei mit zwei Verstößen gibt einen Befund; die Datei wird nicht gefahren (3 Paare).
func TestHandbuchEinBefundJeDatei(t *testing.T) {
	schlecht := praef + ` a"b`
	paare := []string{
		bspDatei(schlecht, nil, nil, "256\n"),
		strings.Join([]string{praef, sepStd, sepErr, "256"}, "\n") + "\n",
		bspDatei(schlecht, nil, []string{"e\r"}, "0\n"),
	}
	for _, inhalt := range paare {
		m := bspUmfeld(t, "k.txt", inhalt, okAntworten())
		code, _, errText := m.lauf()
		if c := strings.Count(errText, "k.txt: Beispiel-Datei ungültig: "); code != exitBefund || c != 1 {
			t.Fatalf("%q: Exit %d, %d Befunde: %q", inhalt, code, c, errText)
		}
		if n := len(m.e.starts(t)); n != 1 {
			t.Fatalf("%q: %d Starts", inhalt, n)
		}
	}
}
