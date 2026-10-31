package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Z1: genau zwei nicht leere Argumente, sonst Aufruf: und nichts gestartet.
func TestHandbuchAufruf(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	b, w := m.e.bin, m.wurzel
	faelle := [][]string{{}, {b}, {b, w, "x"}, {"", w}, {b, ""}, {"", ""}, {b, w, ""}}
	for i, f := range faelle {
		code, out, errText := fuehre(m.u, append([]string{"handbuch"}, f...)...)
		erwarteN(t, code, errText, exitAufruf, "Aufruf:", "pruefung handbuch <Binary> <Wurzel>")
		if out != "" || len(m.e.starts(t)) != 0 {
			t.Fatalf("Fall %d: stdout %q, Starts %d", i, out, len(m.e.starts(t)))
		}
	}
}

// Z2: Binary fehlt (kein Eintrag, Verzeichnis) und nicht startbar (nicht ausführbar, fremdes Format).
func TestHandbuchBinary(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "verz"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "nichtx"), []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "fremd"), []byte("kein Programm\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	faelle := []struct{ pfad, text string }{
		{filepath.Join(d, "gibt-es-nicht"), "Eingabe: Binary fehlt: "},
		{filepath.Join(d, "verz"), "Eingabe: Binary fehlt: "},
		{filepath.Join(d, "nichtx"), "Eingabe: Binary nicht startbar: "},
		{filepath.Join(d, "fremd"), "Eingabe: Binary nicht startbar: "},
	}
	m := neuesUmfeld(t, standardAntworten())
	for _, f := range faelle {
		code, out, errText := fuehre(m.u, "handbuch", f.pfad, m.wurzel)
		erwarte(t, code, errText, exitAufruf, f.text)
		if out != "" || strings.Contains(strings.ReplaceAll(errText, "Eingabe: ", ""), "pruefung handbuch: Zeile") {
			t.Fatalf("%s: Befunde trotz Eingabefehler: %q %q", f.pfad, out, errText)
		}
	}
}

// Z3: Handbuch fehlt oder nicht lesbar: Exit 2, das Binary läuft nicht.
func TestHandbuchNichtLesbar(t *testing.T) {
	faelle := []struct {
		name  string
		bytes []byte
		text  string
	}{
		{"fehlt", nil, "Eingabe: Handbuch fehlt: "},
		{"kein UTF-8", []byte{'#', ' ', 0xff, '\n'}, "Handbuch nicht lesbar: kein UTF-8"},
		{"NUL", []byte{'#', ' ', 0, '\n'}, "Handbuch nicht lesbar: NUL"},
		{"BOM", append([]byte{0xEF, 0xBB, 0xBF}, []byte(sauber())...), "Handbuch nicht lesbar: BOM"},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			m := neuesUmfeld(t, standardAntworten())
			if f.bytes == nil {
				m.entferne(handbuchPfad)
			} else {
				m.schreibe(handbuchPfad, string(f.bytes))
			}
			code, _, errText := m.lauf()
			erwarte(t, code, errText, exitAufruf, f.text)
			if n := len(m.e.starts(t)); n != 0 {
				t.Fatalf("%d Starts vor dem Eingabefehler", n)
			}
		})
	}
}

// Z4: Exit 0 ohne, 1 mit Befund, 2 bei Eingabefehler; ein Eingabefehler schließt Befunde aus.
func TestHandbuchExit(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	code, out, errText := m.lauf()
	if code != exitOK || out != "pruefung handbuch: ok\n" || errText != "" {
		t.Fatalf("sauber: Exit %d, stdout %q, stderr %q", code, out, errText)
	}
	m.handbuch(sauber() + "\nDas ist geplant.\n")
	code, out, errText = m.lauf()
	erwarte(t, code, errText, exitBefund, fmt.Sprintf("pruefung handbuch: Zeile %d: Planung A1: geplant", naechste(sauber())))
	if out != "" {
		t.Fatalf("stdout bei Befund: %q", out)
	}
	code, _, errText = fuehre(m.u, "handbuch", filepath.Join(m.wurzel, "fehlt"), m.wurzel)
	erwarte(t, code, errText, exitAufruf, "Eingabe: Binary fehlt: ")
	nichtEnthalten(t, errText, "Planung")
}

// Z5: CRLF gilt wie LF; ein einzelnes \r ist ein Zeichen der Zeile.
func TestHandbuchCRLF(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	h := sauber() + "\nDas ist geplant.\n"
	m.handbuch(h)
	c1, o1, e1 := m.lauf()
	m.handbuch(strings.ReplaceAll(h, "\n", "\r\n"))
	c2, o2, e2 := m.lauf()
	if c1 != c2 || o1 != o2 || e1 != e2 || c1 != exitBefund {
		t.Fatalf("LF: %d %q; CRLF: %d %q", c1, e1, c2, e2)
	}
	m.handbuch(strings.ReplaceAll(sauber(), "\n", "\r\n"))
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("sauberes CRLF-Handbuch: %d %q", code, errText)
	}
	m.handbuch(sauber() + "\nDas ist ge\rplant.\n")
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitBefund, fmt.Sprintf("Zeile %d: Zeichen U+000D nicht zulässig", naechste(sauber())))
}

// Z6: Zeichen der Kategorien Cc (außer Tabulator), Cf, Co, Mn, Me sind Befund, Tabulator und Zs nicht.
func TestHandbuchZeichen(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	schlecht := []rune{0x01, 0x7F, 0x200B, 0xAD, 0xFEFF, 0xE000, 0x301, 0x20DD, 0x202E}
	for _, r := range schlecht {
		m.handbuch(sauber() + "\nWort" + string(r) + "Wort.\n")
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, fmt.Sprintf("Zeile %d: Zeichen U+%04X nicht zulässig", naechste(sauber()), r))
	}
	for _, r := range []rune{'\t', 0xA0, 0x2003} {
		m.handbuch(sauber() + "\nWort" + string(r) + "Wort.\n")
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("U+%04X: Exit %d %q", r, code, errText)
		}
	}
}

func kopfOhneUndMit(schluessel, ersatzZeilen string) string {
	k := kopfSauber
	for _, z := range strings.Split(strings.TrimSuffix(kopfSauber, "\n"), "\n") {
		if strings.HasPrefix(z, schluessel+":") {
			k = strings.Replace(k, z+"\n", ersatzZeilen, 1)
		}
	}
	return k + "\n## A\n"
}

// Z7: je Schlüssel fehlt, doppelt, leer; ein Befund, doppelt vor leer.
func TestHandbuchKopfSchluessel(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele")
	werte := map[string]string{"Software-Version": "dev", "Gültigkeitsbereich": "Binary", "Stand": "2026-10-30"}
	n := 0
	for _, s := range []string{"Software-Version", "Gültigkeitsbereich", "Stand"} {
		faelle := map[string]string{
			"fehlt":            "",
			"doppelt":          s + ": " + werte[s] + "\n" + s + ": " + werte[s] + "\n",
			"leer":             s + ":\n",
			"doppelt und leer": s + ":\n" + s + ":\n",
		}
		for art, zeilen := range faelle {
			n++
			m.handbuch(kopfOhneUndMit(s, zeilen))
			code, _, errText := m.lauf()
			erwarte(t, code, errText, exitBefund, "Kopf: "+s+" "+strings.Fields(art)[0])
			if c := strings.Count(errText, "Kopf: "+s); c != 1 || strings.Count(errText, "\n") != 1 {
				t.Fatalf("%s %s: %q", s, art, errText)
			}
		}
	}
	if n != 12 {
		t.Fatalf("%d Elemente, erwartet 12", n)
	}
}

// Z8: fremder Schlüssel, Zeile ohne Schlüssel, Überschrift fehlt.
func TestHandbuchKopfFremd(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele")
	faelle := []struct{ zeile, text string }{
		{"Version: dev", "Kopf: Schlüssel Version unbekannt"},
		{"Handbuch-Version: 1", "Kopf: Schlüssel Handbuch-Version unbekannt"},
		{"Autor: x", "Kopf: Schlüssel Autor unbekannt"},
		{"Änderungsdatum: 2026-10-30", "Kopf: Schlüssel Änderungsdatum unbekannt"},
		{"Das ist Prosa", "Kopf: Zeile ohne Schlüssel"},
	}
	for _, f := range faelle {
		m.handbuch(kopfSauber + f.zeile + "\n\n## A\n")
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "Zeile 6: "+f.text)
	}
	m.handbuch(strings.Replace(kopfSauber, "# Handbuch\n", "", 1) + "\n## A\n")
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 2: Kopf: Überschrift fehlt")
}

// Z9: Der Kopf endet an der ersten Zeile `## ` außerhalb von Zäunen.
func TestHandbuchKopfEnde(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele")
	m.handbuch("# H\n\n## A\n\nSoftware-Version: dev\nGültigkeitsbereich: x\nStand: 2026-10-30\n")
	_, _, errText := m.lauf()
	enthaelt2(t, errText, "Kopf: Software-Version fehlt", "Kopf: Gültigkeitsbereich fehlt", "Kopf: Stand fehlt")
	m.handbuch("# H\n\n```\n## im Zaun\n```\nSoftware-Version: dev\nGültigkeitsbereich: x\nStand: 2026-10-30\n")
	_, _, errText = m.lauf()
	nichtEnthalten(t, errText, "Kopf: Software-Version fehlt", "Kopf: Gültigkeitsbereich fehlt", "Kopf: Stand fehlt")
}

// Z10: Software-Version gleich der Ausgabe von version, Aufrufstelle version.
func TestHandbuchVersion(t *testing.T) {
	faelle := []struct {
		name string
		a    antwort
		text string
	}{
		{"gleich", antwort{out: "odata-recorder 1.0\n"}, ""},
		{"anderer Wert", antwort{out: "odata-recorder 2.0\n"}, "Software-Version 1.0 weicht von der Ausgabe von version ab (2.0)"},
		{"ohne Präfix", antwort{out: "1.0\n"}, "version: Ausgabe nicht lesbar: "},
		{"Zusatztext", antwort{out: "odata-recorder 1.0 extra\n"}, "version: Ausgabe nicht lesbar: "},
		{"Exit 1", antwort{out: "odata-recorder 1.0\n", code: 1}, "version: Ausgabe nicht lesbar: Exit 1"},
		{"Fehlerausgabe", antwort{out: "odata-recorder 1.0\n", err: "x\n"}, "version: Ausgabe nicht lesbar: Fehlerausgabe nicht leer"},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			m := neuesUmfeld(t, map[string]antwort{"version": f.a})
			m.entferne("docs/user/beispiele")
			m.handbuch(strings.Replace(kopfSauber, "dev", "1.0", 1) + "\n## A\n")
			code, _, errText := m.lauf()
			if f.text == "" {
				if code != exitOK {
					t.Fatalf("Exit %d %q", code, errText)
				}
				return
			}
			erwarte(t, code, errText, exitBefund, f.text)
			if strings.Contains(f.text, "nicht lesbar") {
				nichtEnthalten(t, errText, "weicht von der Ausgabe")
			}
		})
	}
}

func monatsletzte() []int {
	return []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
}

// Z11: Kalender des Stands; Generator über den letzten Tag jedes Monats 2026 und den Folgetag,
// Orakel: die Tabelle der Monatslängen.
func TestHandbuchStand(t *testing.T) {
	var faelle []standFall
	for mon, letzter := range monatsletzte() {
		faelle = append(faelle,
			standFall{fmt.Sprintf("2026-%02d-%02d", mon+1, letzter), true},
			standFall{fmt.Sprintf("2026-%02d-%02d", mon+1, letzter+1), false})
	}
	faelle = append(faelle, standFall{"2028-02-29", true}, standFall{"30.10.2026", false}, standFall{"2026-1-1", false}, standFall{"", false})
	if len(faelle) != 28 {
		t.Fatalf("%d Elemente, erwartet 28", len(faelle))
	}
	pruefeStaende(t, faelle)
}

type standFall struct {
	wert    string
	gueltig bool
}

func pruefeStaende(t *testing.T, faelle []standFall) {
	t.Helper()
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele")
	for _, f := range faelle {
		m.handbuch(strings.Replace(kopfSauber, "2026-10-30", f.wert, 1) + "\n## A\n")
		code, _, errText := m.lauf()
		if f.gueltig && code != exitOK {
			t.Fatalf("Stand %q: Exit %d %q", f.wert, code, errText)
		}
		text := "Kopf: Stand ungültig: " + f.wert
		if f.wert == "" {
			text = "Kopf: Stand leer"
		}
		if !f.gueltig {
			erwarte(t, code, errText, exitBefund, text)
		}
	}
}

// Z11 (Rand): die Regeln des Jahrhunderts und die Grenzen von Monat und Tag.
func TestHandbuchStandRegeln(t *testing.T) {
	pruefeStaende(t, []standFall{
		{"2000-02-29", true}, {"1900-02-29", false}, {"2100-02-29", false}, {"2024-02-29", true},
		{"2025-02-29", false}, {"2026-13-01", false}, {"2026-00-10", false}, {"2026-05-00", false},
		{"2026-0x-10", false}, {"2026/05/10", false}, {"2026-05-100", false}, {"2026-04-31", false},
	})
}
