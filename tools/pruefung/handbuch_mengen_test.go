package main

import (
	"fmt"
	"strings"
	"testing"
)

// Zählprüfung der Mengen A, B, C: kein Muster ohne Form, keine Form ohne Muster, Muster gleich dem Orakel.
func TestHandbuchMengenZaehlung(t *testing.T) {
	muster, formen := map[byte]int{}, map[byte]int{}
	ms := mengen()
	if len(ms) != len(orakel()) {
		t.Fatalf("%d Muster im Programm, %d im Orakel", len(ms), len(orakel()))
	}
	for i, o := range orakel() {
		if ms[i].id != o.id || ms[i].ausdruck != o.ausdruck {
			t.Fatalf("Muster %d: Programm %q %q, Orakel %q %q", i, ms[i].id, ms[i].ausdruck, o.id, o.ausdruck)
		}
		if len(o.formen) == 0 {
			t.Fatalf("Muster %s ohne Form", o.id)
		}
		muster[o.id[0]]++
		formen[o.id[0]] += len(o.formen)
	}
	want := map[byte][2]int{'A': {16, 36}, 'B': {14, 44}, 'C': {14, 22}}
	for k, w := range want {
		if muster[k] != w[0] || formen[k] != w[1] {
			t.Fatalf("Menge %c: %d Muster und %d Formen, erwartet %d und %d", k, muster[k], formen[k], w[0], w[1])
		}
	}
	t.Logf("Elemente: A %d Formen, B %d, C %d; Stellungen je Form %d", formen['A'], formen['B'], formen['C'], len(stellungen("x")))
}

// Z31 bis Z33: jede Form jedes Musters trifft an sieben Stellungen; Generator = Formen × Stellungen,
// Orakel: die Liste der Formen im Test.
func TestHandbuchMengenFormen(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	zaehler := map[byte]int{}
	for _, p := range orakel() {
		for _, form := range p.formen {
			for name, h := range stellungen(form) {
				m.handbuch(h)
				code, _, errText := m.lauf()
				if code != exitBefund || !strings.Contains(errText, trefferZeile(p.id)) {
					t.Fatalf("%s %q an der Stelle %s: Exit %d, Ausgabe %q", p.id, form, name, code, errText)
				}
				zaehler[p.id[0]]++
			}
		}
	}
	if zaehler['A'] != 252 || zaehler['B'] != 308 || zaehler['C'] != 154 {
		t.Fatalf("Elemente A %d, B %d, C %d, erwartet 252, 308, 154", zaehler['A'], zaehler['B'], zaehler['C'])
	}
}

// bspZitat legt ein gebundenes Beispiel an, dessen Standardausgabe wort ist.
func bspZitat(t *testing.T, wort string) (umfeld, string) {
	t.Helper()
	inhalt := bspDatei(praef, []string{wort}, nil, "0\n")
	a := okAntworten()
	a[""] = antwort{out: wort + "\n"}
	m := bspUmfeld(t, "z.txt", inhalt, a)
	return m, inhalt
}

// Z33: Menge C gilt auch im Zitat, jede Form.
func TestHandbuchMengeCImZitat(t *testing.T) {
	n := 0
	for _, p := range orakel() {
		if p.id[0] != 'C' {
			continue
		}
		for _, form := range p.formen {
			m, _ := bspZitat(t, form)
			code, _, errText := m.lauf()
			erwarte(t, code, errText, exitBefund, fmt.Sprintf("Zeile 10: Kennung %s: ", p.id))
			n++
		}
	}
	if n != 22 {
		t.Fatalf("%d Elemente, erwartet 22", n)
	}
}

// Z34: ein Buchstabe davor oder dahinter ist kein Treffer, außer das Muster nimmt die Beugung auf;
// Orakel: die Ausdrücke im Test.
func TestHandbuchTrefferGrenze(t *testing.T) {
	m := ohneBeispiele(t)
	n := 0
	lauf := func(prosa string) string {
		m.handbuch(rumpf(prosa + "\n"))
		_, _, errText := m.lauf()
		return errText
	}
	for _, p := range orakel() {
		vornOffen := strings.HasPrefix(p.ausdruck, `\p{L}*`)
		hintenOffen := strings.HasSuffix(p.ausdruck, `\p{L}*`)
		for _, form := range p.formen {
			k := kern(form)
			if got := strings.Contains(lauf("Das ist x"+k+" hier."), trefferZeile(p.id)); got != vornOffen {
				t.Fatalf("%s: Buchstabe davor %q: Treffer=%v", p.id, k, got)
			}
			for _, nicht := range []string{"1", "-", "/"} {
				if !strings.Contains(lauf("Das ist "+nicht+k+" hier."), trefferZeile(p.id)) {
					t.Fatalf("%s: %q vor %q ist kein Buchstabe, aber kein Treffer", p.id, nicht, k)
				}
			}
			if !strings.Contains(lauf("Das ist x"+k+" und "+k+" hier."), trefferZeile(p.id)) {
				t.Fatalf("%s: der zweite, gültige Treffer neben dem ersten mit Buchstaben davor fehlt (%q)", p.id, k)
			}
			if !hintenOffen && strings.Contains(lauf("Das ist "+k+"x hier."), trefferZeile(p.id)) {
				t.Fatalf("%s: Buchstabe dahinter %q: Treffer", p.id, k)
			}
			n++
		}
	}
	if n != 102 {
		t.Fatalf("%d Formen, erwartet 102", n)
	}
}

// Z35: Normalisierung; Generator = Formen mit Leerzeichen im Treffer × Varianten des Leerraums, des Markups
// und der Schreibung.
func TestHandbuchNormalisierung(t *testing.T) {
	m := ohneBeispiele(t)
	ersetze := func(k, mit string) string { return strings.Replace(k, " ", mit, 1) }
	varianten := map[string]func(string) string{
		"Großschreibung": strings.ToUpper,
		"Zeilenumbruch":  func(k string) string { return ersetze(k, "\n") },
		"mehrere":        func(k string) string { return ersetze(k, "   ") },
		"Tabulator":      func(k string) string { return ersetze(k, "\t") },
		"NBSP":           func(k string) string { return ersetze(k, string(rune(0xA0))) },
		"U+2003":         func(k string) string { return ersetze(k, string(rune(0x2003))) },
		"U+2028":         func(k string) string { return ersetze(k, string(rune(0x2028))) },
		"U+2029":         func(k string) string { return ersetze(k, string(rune(0x2029))) },
		"Stern":          func(k string) string { return ersetze(k, "* ") },
		"Unterstrich":    func(k string) string { return ersetze(k, "_ ") },
		"Backtick":       func(k string) string { return ersetze(k, "` ") },
	}
	n := 0
	for _, p := range orakel() {
		for _, form := range p.formen {
			k := kern(form)
			if !strings.Contains(k, " ") {
				continue
			}
			for name, v := range varianten {
				m.handbuch(rumpf("Das ist " + v(k) + " hier.\n"))
				code, _, errText := m.lauf()
				if code != exitBefund || !strings.Contains(errText, trefferZeile(p.id)) {
					t.Fatalf("%s %q Variante %s: Exit %d %q", p.id, k, name, code, errText)
				}
				n++
			}
		}
	}
	t.Logf("%d Elemente (%d Varianten)", n, len(varianten))
	if n == 0 {
		t.Fatal("keine Form mit Leerzeichen")
	}
}

// Z36: Zitat; ein gebundener Block fällt nicht unter A und B, wohl unter C; ein ungebundener ist Prosa.
func TestHandbuchZitat(t *testing.T) {
	menge := []struct{ wort, id string }{{"geplant", "A1"}, {"bisher", "B1"}, {"spec-110", "C6"}}
	n := 0
	for _, w := range menge {
		for _, art := range []string{"gebunden", "ohne Bindungszeile", "Block ungleich Datei", "Datei fehlt", "falsche Bindungszeile"} {
			m, inhalt := bspZitat(t, w.wort)
			switch art {
			case "ohne Bindungszeile":
				m.handbuch(rumpf(zaunVon(inhalt)))
			case "Block ungleich Datei":
				m.handbuch(rumpf(binde("beispiele/z.txt") + zaunVon(inhalt+"x\n")))
			case "Datei fehlt":
				m.handbuch(rumpf(binde("beispiele/fehlt.txt") + zaunVon(inhalt)))
			case "falsche Bindungszeile":
				m.handbuch(rumpf("Datei: [beispiele/z.txt](beispiele/z2.txt)\n" + zaunVon(inhalt)))
			}
			code, _, errText := m.lauf()
			trifft := strings.Contains(errText, trefferZeile(w.id))
			wollen := art != "gebunden" || w.id[0] == 'C'
			if trifft != wollen {
				t.Fatalf("%s %s: Treffer=%v, erwartet %v, Exit %d, Ausgabe %q", w.id, art, trifft, wollen, code, errText)
			}
			if art == "gebunden" && w.id[0] != 'C' && code != exitOK {
				t.Fatalf("%s gebunden: Exit %d %q", w.id, code, errText)
			}
			n++
		}
	}
	if n != 15 {
		t.Fatalf("%d Elemente, erwartet 15", n)
	}
}

// Z37: Meldung mit Zeile des Absatzanfangs; je Absatz und Muster ein Befund, mehrere Mengen im Absatz.
func TestHandbuchMeldung(t *testing.T) {
	m := ohneBeispiele(t)
	faelle := []struct {
		prosa string
		texte []string
		zahl  int
	}{
		{"Das ist geplant und spec-110.\n", []string{"Zeile 9: Planung A1: geplant", "Zeile 9: Kennung C6: spec-110"}, 2},
		{"Das ist geplant und bisher so.\n", []string{"Zeile 9: Planung A1: geplant", "Zeile 9: Chronik B1: bisher"}, 2},
		{"Das ist geplant und künftig so.\n", []string{"Zeile 9: Planung A1: geplant", "Zeile 9: Planung A4: künftig"}, 2},
		{"Das ist geplant und geplante Sache.\n", []string{"Zeile 9: Planung A1: geplant\n"}, 1},
		{"Eins.\nzwei geplant\ndrei.\n", []string{"Zeile 9: Planung A1: geplant"}, 1},
		{"Eins.\n\nzwei geplant\n", []string{"Zeile 11: Planung A1: geplant"}, 1},
	}
	for _, f := range faelle {
		m.handbuch(rumpf(f.prosa))
		code, _, errText := m.lauf()
		erwarteN(t, code, errText, exitBefund, f.texte...)
		if c := strings.Count(errText, "\n"); c != f.zahl {
			t.Fatalf("%q: %d Befunde, erwartet %d: %q", f.prosa, c, f.zahl, errText)
		}
	}
}

// Z38: Reihenfolge der Befunde nach den Nummern des Vertrags (3 bis 12), gleiche Eingabe gleiche Ausgabe.
func TestHandbuchReihenfolge(t *testing.T) {
	m := neuesUmfeld(t, map[string]antwort{"version": {out: "odata-recorder dev\n"}, "": {out: "x\n"}})
	m.schreibe("Makefile", "build: ## x\n")
	m.schreibe("docs/user/beispiele/kaputt.txt", "x\n")
	h := strings.Replace(kopfSauber, "dev", "1.0", 1) + "\n## A\n\n" +
		"Das ist spec-110.\n\nDas war bisher so.\n\nDas ist geplant.\n\nSiehe `make nix`.\n\n" +
		"Wort" + string(rune(0x200B)) + "Wort.\n\n" +
		binde("beispiele/kaputt.txt") + zaunVon("x\n") + "\n" +
		"```\nohne Bindung\n```\n\n" +
		binde("beispiele/aufruf.txt") + zaunVon(aufrufDatei)
	m.handbuch(h)
	code, _, erste := m.lauf()
	if code != exitBefund {
		t.Fatalf("Exit %d %q", code, erste)
	}
	marken := []string{
		"Zeichen U+200B", "Software-Version 1.0", "Beispiel ohne Bindungszeile", "kaputt.txt: Beispiel-Datei ungültig",
		"aufruf.txt: Standardausgabe weicht ab", "Aufrufpräfix nicht gedeckt", "make-Ziel nix", "Planung A1", "Chronik B1", "Kennung C6",
	}
	letzte := -1
	for _, mk := range marken {
		i := strings.Index(erste, mk)
		if i < 0 || i <= letzte {
			t.Fatalf("Marke %q an Stelle %d (vorige %d): %q", mk, i, letzte, erste)
		}
		letzte = i
	}
	_, _, zweite := m.lauf()
	if zweite != erste {
		t.Fatalf("zweiter Lauf weicht ab:\n%q\n%q", erste, zweite)
	}
}

func TestHandbuchKuerzen(t *testing.T) {
	m := bspMit(t, nil, nil, nil, "0\n", antwort{out: strings.Repeat("a", 300) + "\n"})
	_, _, errText := m.lauf()
	if !strings.Contains(errText, "erhalten \""+strings.Repeat("a", 199)) || strings.Contains(errText, strings.Repeat("a", 200)) {
		t.Fatalf("Ausgabe nicht auf 200 Zeichen gekürzt: %q", errText)
	}
}
