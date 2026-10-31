package main

import (
	"bytes"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// Tests zu den Befunden des zweiten Reviews (R-1 bis R-5): Mengen aus dem Wortlaut des Vertrags.

// leerraum19 ist die Menge des Leerraums nach Nr. 3 b und 13 a ohne das Leerzeichen: die 16 Zeichen der Kategorie Zs
// außer U+0020 (U+00A0, U+1680, U+2000 bis U+200A, U+202F, U+205F, U+3000), U+2028, U+2029 und der Tabulator; die Zeichen
// entstehen aus Codepunkten (§3.18).
func leerraum19() []string {
	cps := []rune{0xA0, 0x1680}
	for c := 0x2000; c <= 0x200A; c++ {
		cps = append(cps, rune(c))
	}
	cps = append(cps, 0x202F, 0x205F, 0x3000, 0x2028, 0x2029, '\t')
	var out []string
	for _, c := range cps {
		out = append(out, string(c))
	}
	return out
}

func TestLeerraum19(t *testing.T) {
	l := leerraum19()
	if len(l) != 19 {
		t.Fatalf("%d Zeichen, erwartet 19", len(l))
	}
	for _, s := range l {
		r := []rune(s)[0]
		if !unicode.IsSpace(r) || r == ' ' {
			t.Fatalf("U+%04X ist kein Leerraum außer U+0020", r)
		}
	}
}

// ausgabeVon schreibt die Befunde mit meldeBefunde und liefert die Zeilen von stderr.
func ausgabeVon(bf []befund) []string {
	var out, errb bytes.Buffer
	meldeBefunde(bf, &out, &errb)
	return strings.Split(strings.TrimSuffix(errb.String(), "\n"), "\n")
}

// Z38 (a): ein Befund ohne Zeile hat die Zeile 0 und steht vor jedem Befund derselben Nummer mit Zeile, in beiden
// Reihenfolgen der Erzeugung, für die Nummern mit Befunden beider Arten (4 und 5; 4 Elemente).
func TestHandbuchReihenfolgeZeileNull(t *testing.T) {
	n := 0
	for _, f := range []struct {
		ohne, mit befund
	}{
		{befund{nr: nrKopf, text: "Kopf: Stand fehlt"}, befund{nr: nrKopf, zeile: 5, text: "Zeile 5: Kopf: Zeile ohne Schlüssel"}},
		{befund{nr: nrZaun, name: "x.txt", text: "beispiele: x.txt ohne Bindung"}, befund{nr: nrZaun, zeile: 7, text: "Zeile 7: Beispiel ohne Bindungszeile"}},
	} {
		for _, umgekehrt := range []bool{false, true} {
			bf := []befund{f.ohne, f.mit}
			if umgekehrt {
				bf = []befund{f.mit, f.ohne}
			}
			z := ausgabeVon(bf)
			if z[0] != ausgabeKopf+f.ohne.text || z[1] != ausgabeKopf+f.mit.text {
				t.Fatalf("Nr. %d, umgekehrt=%v: %q", f.ohne.nr, umgekehrt, z)
			}
			n++
		}
	}
	if n != 4 {
		t.Fatalf("%d Elemente, erwartet 4", n)
	}
}

// Z38 (b): die Namen stehen in Bytefolge der Namen mit Endung (`version` ohne), alle geordneten Paare (30).
func TestHandbuchReihenfolgeNamen(t *testing.T) {
	namen := []string{"a.txt", "a-b.txt", "a0.txt", "ab.txt", "version.txt", "version"}
	n := 0
	for _, x := range namen {
		for _, y := range namen {
			if x == y {
				continue
			}
			z := ausgabeVon([]befund{{nr: nrLauf, name: x, text: x}, {nr: nrLauf, name: y, text: y}})
			soll := []string{x, y}
			sort.Strings(soll)
			if z[0] != ausgabeKopf+soll[0] || z[1] != ausgabeKopf+soll[1] {
				t.Fatalf("Paar %s, %s: %q", x, y, z)
			}
			n++
		}
	}
	if n != 30 {
		t.Fatalf("%d Paare, erwartet 30", n)
	}
	// Die Zeile geht dem Namen vor: Zeile 3 mit Name `b.txt` steht vor Zeile 4 mit Name `a.txt`.
	z := ausgabeVon([]befund{{nr: nrLauf, zeile: 4, name: "a.txt", text: "vier"}, {nr: nrLauf, zeile: 3, name: "b.txt", text: "drei"}})
	if z[0] != ausgabeKopf+"drei" {
		t.Fatalf("Zeile vor Name: %q", z)
	}
}

// Z38 (a, vom Programm erzeugt): `Kopf: <Schlüssel> fehlt` und `beispiele: <Name> …` haben keine Zeile und stehen vor
// den Befunden derselben Nummer mit Zeile.
func TestHandbuchReihenfolgeZeileNullImLauf(t *testing.T) {
	m := ohneBeispiele(t)
	m.handbuch("# H\n\nSoftware-Version: dev\nGültigkeitsbereich: x\nUnsinn\n\n## A\n")
	_, _, errText := m.lauf()
	vorher(t, errText, "Kopf: Stand fehlt", "Zeile 5: Kopf: Zeile ohne Schlüssel")
	m = neuesUmfeld(t, standardAntworten())
	m.schreibe("docs/user/beispiele/x.txt", "x\n")
	m.schreibe("docs/user/beispiele/Y.txt", "x\n")
	m.handbuch(rumpf("```\ntext\n```\n"))
	_, _, errText = m.lauf()
	vorher(t, errText, "beispiele: x.txt ohne Bindung", "Zeile 9: Beispiel ohne Bindungszeile")
	vorher(t, errText, "beispiele: Y.txt keine Beispiel-Datei", "Zeile 9: Beispiel ohne Bindungszeile")
}

func vorher(t *testing.T, ausgabe, a, b string) {
	t.Helper()
	i, j := strings.Index(ausgabe, a), strings.Index(ausgabe, b)
	if i < 0 || j < 0 || i > j {
		t.Fatalf("%q steht nicht vor %q: %q", a, b, ausgabe)
	}
}

// Z38 (b, Namen der Dateien im Lauf): `a-b.txt` steht vor `a.txt`, in beiden Reihenfolgen der Bindung; der Name
// trägt die Endung.
func TestHandbuchReihenfolgeNamenImLauf(t *testing.T) {
	dateiA := bspDatei(praef, []string{"x"}, nil, "0\n")
	dateiAB := bspDatei(praef+" foo", []string{"y"}, nil, "0\n")
	for _, abErst := range []bool{false, true} {
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("docs/user/beispiele/aufruf.txt")
		m.schreibe("docs/user/beispiele/a.txt", dateiA)
		m.schreibe("docs/user/beispiele/a-b.txt", dateiAB)
		ba := binde("beispiele/a.txt") + zaunVon(dateiA)
		bab := binde("beispiele/a-b.txt") + zaunVon(dateiAB)
		if abErst {
			m.handbuch(rumpf(bab + "\n" + ba))
		} else {
			m.handbuch(rumpf(ba + "\n" + bab))
		}
		_, _, errText := m.lauf()
		vorher(t, errText, "a-b.txt: Standardausgabe weicht ab", "a.txt: Standardausgabe weicht ab")
	}
}

// Z38 (c): bei gleicher Nummer, Zeile und gleichem Namen entscheidet die Reihenfolge der Prüfung; sieben Gruppen,
// je in beiden Reihenfolgen des Textes, wo der Text sie bestimmt.
func TestHandbuchReihenfolgeGleichstaende(t *testing.T) {
	zwsp, soh := string(rune(0x200B)), string(rune(0x01))
	m := ohneBeispiele(t)
	// Nr. 3: Zeichen in der Reihenfolge ihrer ersten Stelle in der Zeile.
	for _, prosa := range []string{"Wort" + zwsp + "mit" + soh + "x.", "Wort" + soh + "mit" + zwsp + "x."} {
		m.handbuch(rumpf(prosa + "\n"))
		_, _, errText := m.lauf()
		a, b := "U+200B", "U+0001"
		if strings.Index(prosa, soh) < strings.Index(prosa, zwsp) {
			a, b = b, a
		}
		vorher(t, errText, a, b)
	}
	// Nr. 4: Software-Version, Gültigkeitsbereich, Stand.
	m.handbuch("# H\n\n## A\n")
	_, _, errText := m.lauf()
	vorher(t, errText, "Kopf: Software-Version fehlt", "Kopf: Gültigkeitsbereich fehlt")
	vorher(t, errText, "Kopf: Gültigkeitsbereich fehlt", "Kopf: Stand fehlt")
	// Nr. 5: Info-Zeichenkette vor dem Befund der Bindung, ohne und mit ungültiger Bindungszeile.
	for _, f := range []struct{ vor, bindung string }{{"Text.\n", "Beispiel ohne Bindungszeile"}, {"Datei: [x](y)\n", "Bindung ungültig"}} {
		m.handbuch(rumpf(f.vor + "```text\n" + aufrufDatei + "```\n"))
		_, _, errText = m.lauf()
		vorher(t, errText, "Info-Zeichenkette text nicht zulässig", f.bindung)
	}
	// Nr. 7: Standardausgabe, Fehlerausgabe, Ausgangswert.
	a := map[string]antwort{"version": {out: "odata-recorder dev\n"}, "": {out: "x\n", err: "y\n", code: 3}}
	m7 := bspUmfeld(t, "k.txt", bspDatei(praef, nil, nil, "0\n"), a)
	_, _, errText = m7.lauf()
	vorher(t, errText, "k.txt: Standardausgabe weicht ab", "k.txt: Fehlerausgabe weicht ab")
	vorher(t, errText, "k.txt: Fehlerausgabe weicht ab", "k.txt: Ausgangswert weicht ab")
	// Nr. 8: Image vor Pfad.
	m8 := neuesUmfeld(t, standardAntworten())
	m8.schreibe("Makefile", "build: ## x\n")
	m8.schreibe("Dockerfile", "RUN x\n")
	_, _, errText = m8.lauf()
	vorher(t, errText, "Aufrufpräfix nicht gedeckt: Image", "Aufrufpräfix nicht gedeckt: Pfad")
	// Nr. 9: Spans in der Reihenfolge im Absatz, in beiden Reihenfolgen des Textes.
	for _, prosa := range []string{"`make nix` und `make nox`", "`make nox` und `make nix`"} {
		m.handbuch(rumpf(prosa + "\n"))
		_, _, errText = m.lauf()
		a, b := "make-Ziel nix", "make-Ziel nox"
		if strings.Index(prosa, "nox") < strings.Index(prosa, "nix") {
			a, b = b, a
		}
		vorher(t, errText, a, b)
	}
	// Nr. 10 bis 12: Muster in der Reihenfolge der Tabelle, unabhängig von der Reihenfolge im Text.
	for _, prosa := range []string{"Das ist geplant und künftig so.", "Das ist künftig und geplant so."} {
		m.handbuch(rumpf(prosa + "\n"))
		_, _, errText = m.lauf()
		vorher(t, errText, "Planung A1: geplant", "Planung A4: künftig")
	}
}

// R-5: der Name einer Beispiel-Datei und die Bindungszeile haben die Form `beispiele/`, ASCII a-z und 0-9, einzelne
// Striche, `.txt` mit wörtlichem Punkt und danach nichts. Menge: jedes Zeichen vor `txt`, Rest nach `.txt`, Verzeichnis,
// Name mit Umlaut, Großbuchstaben, führendem, endendem und doppeltem Strich.
func TestHandbuchNameUndBindungFormSchleife(t *testing.T) {
	var ungueltig []string
	for c := 0x21; c <= 0x7e; c++ {
		if c != '.' && c != '/' {
			ungueltig = append(ungueltig, "a"+string(rune(c))+"txt")
		}
	}
	ungueltig = append(ungueltig, "atxt", "a.txtx", "a.txt.txt", "a.txt ", "\xc3\xa4.txt", "\xc3\x84.txt", "A.txt", "-a.txt", "a--b.txt", "a-.txt", ".txt")
	gueltig := []string{"a.txt", "a-b.txt", "0.txt", "a0-9.txt"}

	m := ohneBeispiele(t)
	for _, name := range append(append([]string{}, ungueltig...), gueltig...) {
		m.schreibe("docs/user/beispiele/"+name, "x\n")
	}
	_, _, errText := m.lauf()
	for _, name := range ungueltig {
		enthaelt2(t, errText, "beispiele: "+name+" keine Beispiel-Datei\n")
		nichtEnthalten(t, errText, "beispiele: "+name+" ohne Bindung")
	}
	for _, name := range gueltig {
		enthaelt2(t, errText, "beispiele: "+name+" ohne Bindung\n")
	}
	n := len(ungueltig) + len(gueltig)

	zaun := zaunVon(aufrufDatei)
	bm := neuesUmfeld(t, standardAntworten())
	pfade := []string{"beispiel/a.txt", "Beispiele/a.txt", "beispiele/x/a.txt", "./beispiele/a.txt", "/beispiele/a.txt", "beispiele//a.txt", "x/beispiele/a.txt"}
	for c := 0x21; c <= 0x7e; c++ {
		if c != '.' {
			pfade = append(pfade, "beispiele/a"+string(rune(c))+"txt")
		}
	}
	for _, name := range ungueltig {
		pfade = append(pfade, "beispiele/"+name)
	}
	for _, p := range pfade {
		z := "Datei: [" + p + "](" + p + ")"
		bm.handbuch(rumpf(z + "\n" + zaun))
		code, _, errText := bm.lauf()
		erwarte(t, code, errText, exitBefund, "Bindung ungültig: "+z)
		n++
	}
	t.Logf("%d Elemente (%d Namen im Verzeichnis, %d Pfade der Bindung)", n, len(ungueltig)+len(gueltig), len(pfade))
}
