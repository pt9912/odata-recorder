package main

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

// Die Menge M6 (Marken und Messungen auf und neben jeder Grenze, SPEC-128 Nr. 5 a bis i) und die
// Menge M10 (Ströme des Programms, Nr. 6 a und 7 a). Das Orakel zu M6 ist aus dem Wortlaut von
// Nr. 5 geschrieben, rechnet die Marke mit big.Rat (orakelMarke) und nutzt keinen Code des Programms.

type m6Eintrag struct {
	n     int64
	grund string
}

type m6Fall struct {
	tabelleS       int64 // S aus der Tabelle: eine Zeile, Produktcode S − 80
	marke, schaetz int64
	schnittGrund   string // Text hinter `schnitt-grund:`; "fehlt" lässt die Zeile weg
	messung        int64
	eintraege      []m6Eintrag
}

func (f m6Fall) plan() string {
	aend := []string{
		fmt.Sprintf("produktcode: %d", f.tabelleS-80), fmt.Sprintf("schaetzung: %d", f.schaetz), fmt.Sprintf("marke: %d", f.marke),
	}
	if f.schnittGrund != "fehlt" {
		aend = append(aend, "schnitt-grund: "+f.schnittGrund)
	}
	for _, e := range f.eintraege {
		aend = append(aend, fmt.Sprintf("ueberschreitung: %d |%s", e.n, e.grund))
	}
	return planMit(tabelleText(0, 1), blockB(aend...))
}

// m6Soll ist, was Nr. 5 für den Fall vorschreibt: die Befunde (Anfang der Meldung und ein Text,
// den sie tragen muss) in der Reihenfolge g, b, c, d, e und die Hinweise.
type m6Soll struct {
	befunde  [][2]string
	hinweise []string
}

func (f m6Fall) soll() m6Soll {
	var s m6Soll
	gruppe := func(text string) { s.befunde = append(s.befunde, [2]string{text, ""}) }
	// g) Schnittfrage: 135·S über 300000 und kein Grund (leer oder Leerraum gilt als nicht gesetzt).
	if 135*f.schaetz > 300000 && (f.schnittGrund == "fehlt" || strings.TrimSpace(f.schnittGrund) == "") {
		gruppe("Schnittfrage ohne Grund")
	}
	// b) Marke gleich Formel.
	if erwartet := orakelMarke(f.schaetz); f.marke != erwartet {
		gruppe(fmt.Sprintf("Marke %d weicht von der Formel ab (Schätzung %d, erwartet %d)", f.marke, f.schaetz, erwartet))
	}
	// c) Schätzung gleich Tabelle: vor dem Code (Messung 0) ein Befund, im Code ein Hinweis.
	if f.schaetz != f.tabelleS {
		if f.messung == 0 {
			gruppe(fmt.Sprintf("Schätzung %d weicht von der Tabelle ab (Tabelle ergibt %d)", f.schaetz, f.tabelleS))
		} else {
			s.hinweise = append(s.hinweise, fmt.Sprintf("Tabelle ergibt %d, Schätzung %d", f.tabelleS, f.schaetz))
		}
	}
	// d) Überschreitung: Messung über der Marke; gedeckt, wenn ein Eintrag mit N ≥ Messung steht.
	if f.messung > f.marke {
		zeile := fmt.Sprintf("ueberschreitung: %d | <Grund>", orakelAufHundert(f.messung))
		var groesstes int64 = -1
		for _, e := range f.eintraege {
			groesstes = max(groesstes, e.n)
		}
		switch {
		case len(f.eintraege) == 0:
			s.befunde = append(s.befunde, [2]string{fmt.Sprintf("Messung %d über Marke %d ohne Eintrag", f.messung, f.marke), zeile})
		case groesstes < f.messung:
			s.befunde = append(s.befunde, [2]string{fmt.Sprintf("Messung %d über Marke %d: Eintrag reicht nicht (größtes N %d)", f.messung, f.marke, groesstes), zeile})
		}
	}
	// e) Grund: ein leerer Grund ist ein Befund, unabhängig von der Messung.
	for _, e := range f.eintraege {
		if strings.TrimSpace(e.grund) == "" {
			gruppe(fmt.Sprintf("ueberschreitung bis %d ohne Grund", e.n))
		}
	}
	// f) Eintrag ohne Überschreitung ist ein Hinweis.
	if f.messung <= f.marke && len(f.eintraege) > 0 {
		s.hinweise = append(s.hinweise, "Eintrag ohne Überschreitung")
	}
	return s
}

// orakelAufHundert ist 100·⌈m/100⌉ mit big.Rat: die Messung aufgerundet auf volle Hundert.
func orakelAufHundert(m int64) int64 {
	r := big.NewRat(m, 100)
	q := new(big.Int).Div(r.Num(), r.Denom())
	if !r.IsInt() {
		q.Add(q, big.NewInt(1))
	}
	return q.Int64() * 100
}

func (f m6Fall) pruefe(t *testing.T, was string) bool {
	t.Helper()
	r := pr(f.plan(), messungB(int(f.messung)))
	s := f.soll()
	code := 0
	if len(s.befunde) > 0 {
		code = 1
	}
	fehler := ""
	if r.code != code {
		fehler += fmt.Sprintf(" Exit %d statt %d;", r.code, code)
	}
	zeilen := nichtLeer(strings.Split(r.err, "\n"))
	if len(zeilen) != len(s.befunde) {
		fehler += fmt.Sprintf(" %d Befunde statt %d;", len(zeilen), len(s.befunde))
	} else {
		for i, b := range s.befunde {
			if !strings.HasPrefix(zeilen[i], "groessenmarke p.md: "+b[0]) || !strings.Contains(zeilen[i], b[1]) {
				fehler += fmt.Sprintf(" Befund %d ist %q statt %q mit %q;", i, zeilen[i], b[0], b[1])
			}
		}
	}
	var hinweise []string
	for _, z := range nichtLeer(strings.Split(r.out, "\n")) {
		if h, ok := strings.CutPrefix(z, "groessenmarke p.md: Hinweis: "); ok {
			hinweise = append(hinweise, h)
		}
	}
	if strings.Join(hinweise, "|") != strings.Join(s.hinweise, "|") {
		fehler += fmt.Sprintf(" Hinweise %q statt %q;", hinweise, s.hinweise)
	}
	if ok := strings.Contains(r.out, "ok, Messung"); ok != (len(s.befunde) == 0) {
		fehler += fmt.Sprintf(" Zeile ok: %v;", ok)
	}
	if fehler != "" {
		t.Errorf("%s %+v:%s", was, f, fehler)
		return false
	}
	return true
}

func nichtLeer(zeilen []string) []string {
	var out []string
	for _, z := range zeilen {
		if z != "" {
			out = append(out, z)
		}
	}
	return out
}

func TestM6MarkenUndMessungen(t *testing.T) {
	grund := " Grund"
	fehlend := 0
	n := 0
	melde := func(f m6Fall) {
		n++
		if !f.pruefe(t, "M6") {
			fehlend++
			if fehlend >= 10 {
				t.Fatal("zu viele Abweichungen")
			}
		}
	}

	// (a) 5 Schätzungen × 5 Marken × 3 Schätzungen im Block × 3 Schnitt-Gründe × 41 Messungen mit
	// Einträgen.
	a := 0
	for _, s := range []int64{2000, 2001, 2222, 2223, 4670} {
		m := orakelMarke(s)
		for _, marke := range []int64{m - 100, m - 1, m, m + 1, m + 100} {
			for _, ds := range []int64{-1, 0, 1} {
				for _, sg := range []string{"fehlt", "   ", "Text"} {
					fall := m6Fall{tabelleS: s, marke: marke, schaetz: s + ds, schnittGrund: sg}
					// Messung 0: kein Eintrag; N = Messung und N = Messung + 1, je mit Grund und mit leerem Grund.
					fall.messung = 0
					fall.eintraege = nil
					melde(fall)
					a++
					for _, n0 := range []int64{0, 1} {
						for _, g := range []string{grund, ""} {
							fall.eintraege = []m6Eintrag{{n0, g}}
							melde(fall)
							a++
						}
					}
					// Messung 1, Marke − 1, Marke, Marke + 1.
					for _, x := range []int64{1, marke - 1, marke, marke + 1} {
						fall.messung = x
						fall.eintraege = nil
						melde(fall)
						a++
						for _, nn := range []int64{x - 1, x, x + 1} {
							for _, g := range []string{grund, ""} {
								fall.eintraege = []m6Eintrag{{nn, g}}
								melde(fall)
								a++
							}
						}
						fall.eintraege = []m6Eintrag{{x - 1, grund}, {x + 1, grund}}
						melde(fall)
						a++
						fall.eintraege = []m6Eintrag{{x + 1, grund}, {x - 1, grund}}
						melde(fall)
						a++
					}
				}
			}
		}
	}
	if a != 9225 {
		t.Errorf("M6 (a): %d Elemente statt 9225", a)
	}

	// (b) Die Zeile zum Eintragen: die Messung aufgerundet auf volle Hundert. Marke 0 lässt jede
	// Messung ab 1 über der Marke liegen.
	b := 0
	for _, x := range []int64{1, 99, 100, 101, 1200, 1201, 6400, 6401} {
		b++
		f := m6Fall{tabelleS: 800, marke: 0, schaetz: 800, schnittGrund: "fehlt", messung: x}
		melde(f)
		want := fmt.Sprintf("ueberschreitung: %d | <Grund>", (x+99)/100*100)
		if r := pr(f.plan(), messungB(int(x))); !strings.Contains(r.err, want) {
			t.Errorf("Messung %d: Zeile %q fehlt:\n%s", x, want, r.err)
		}
	}
	if b != 8 {
		t.Errorf("M6 (b): %d Elemente statt 8", b)
	}

	// (c) Leerer Grund bei einer Überschreitung: N ab der Messung ist `ohne Grund` und nicht `ohne
	// Eintrag` oder `reicht nicht`; darunter beides.
	c := 0
	for _, g := range []string{"", "   ", "\t"} {
		for _, nn := range []int64{1100, 1101, 1102} {
			c++
			f := m6Fall{tabelleS: 800, marke: 1100, schaetz: 800, schnittGrund: "fehlt", messung: 1101, eintraege: []m6Eintrag{{nn, g}}}
			melde(f)
			r := pr(f.plan(), messungB(1101))
			if strings.Contains(r.err, "ohne Eintrag") || !strings.Contains(r.err, fmt.Sprintf("ueberschreitung bis %d ohne Grund", nn)) ||
				strings.Contains(r.err, "reicht nicht") != (nn < 1101) {
				t.Errorf("Grund %q N %d:\n%s", g, nn, r.err)
			}
		}
	}
	if c != 9 {
		t.Errorf("M6 (c): %d Elemente statt 9", c)
	}
	if n != 9225+8+9 {
		t.Errorf("M6: %d Läufe statt %d", n, 9225+8+9)
	}
}

// ausgabeArt ist eine der 30 Ausgabearten von M10: der Aufruf, der Text und der Strom, in dem er
// steht (SPEC-128 Nr. 6 a, 7 a).
type ausgabeArt struct {
	name  string
	lauf  func() ergebnis
	text  string
	strom string // "stdout" oder "stderr"
}

// ausgabeArten sind die 30 Arten; plan ist der Pfad eines vorhandenen Plans (für `Messung fehlt`).
func ausgabeArten(plan string) []ausgabeArt {
	pfNeu := func(plan string, messung int) func() ergebnis {
		return func() ergebnis { return pr(plan, messungB(messung)) }
	}
	pfM := func(plan, messung string) func() ergebnis { return func() ergebnis { return pr(plan, messung) } }
	abf := func(plan string) func() ergebnis { return func() ergebnis { return ab(plan) } }
	binaer := kopfText(basisB, spitzeB, "src/") + satz("-", "-", "x.bin")
	var as []ausgabeArt

	// pruefen, stdout (5)
	as = append(as,
		ausgabeArt{"ok", pfNeu(planB(), 1100), "ok, Messung 1100 von Marke 1100", "stdout"},
		ausgabeArt{"Messpunkt-Zeile", pfNeu(planB(), 1100), "Messpunkt p.md: zeilen=10", "stdout"},
		ausgabeArt{"Hinweis Binärdatei", pfM(planB(), binaer), "Hinweis: 1 Bin\xc3\xa4rdatei(en) nicht gez\xc3\xa4hlt", "stdout"},
		ausgabeArt{"Hinweis Tabelle ergibt", pfNeu(planB("schaetzung: 799"), 5), "Hinweis: Tabelle ergibt 800, Sch\xc3\xa4tzung 799", "stdout"},
		ausgabeArt{"Hinweis Eintrag ohne Überschreitung", pfNeu(planB(eintragZeile(1200, "Grund")), 1100), "Hinweis: Eintrag ohne \xc3\x9cberschreitung", "stdout"},
	)
	// pruefen, stderr: 7 Befunde
	as = append(as,
		ausgabeArt{"Befund Marke", pfNeu(planB("marke: 1000"), 0), "Marke 1000 weicht von der Formel ab", "stderr"},
		ausgabeArt{"Befund Schätzung", pfNeu(planB("schaetzung: 799"), 0), "weicht von der Tabelle ab", "stderr"},
		ausgabeArt{"Befund über Marke", pfNeu(planB(), 1101), "Messung 1101 \xc3\xbcber Marke 1100 ohne Eintrag", "stderr"},
		ausgabeArt{"Befund Eintrag reicht nicht", pfNeu(planB(eintragZeile(1100, "Grund")), 1101), "Eintrag reicht nicht", "stderr"},
		ausgabeArt{"Befund ueberschreitung ohne Grund", pfNeu(planB(eintragZeile(1200, "")), 1100), "ueberschreitung bis 1200 ohne Grund", "stderr"},
		ausgabeArt{"Befund ersetzt ohne Grund", pfNeu(planB("ersetzt: 1000 |"), 1100), "ersetzt 1000 ohne Grund", "stderr"},
		ausgabeArt{"Befund Schnittfrage", pfNeu(pruefPlan(2223), 1), "Schnittfrage ohne Grund", "stderr"},
	)
	// pruefen, stderr: 9 Eingabefehler
	as = append(as,
		ausgabeArt{"Eingabe Tabelle fehlt", pfNeu(strings.Replace(planB(), "| Nr. |", "| Nummer |", 1), 0), "Eingabe: Tabelle fehlt", "stderr"},
		ausgabeArt{"Eingabe Zellen", pfNeu(planMit(kopf5+"| Z1 | z | f | m |\n", blockB()), 0), "Eingabe: Zeile 1: 4 Zellen statt 5", "stderr"},
		ausgabeArt{"Eingabe Block mehrdeutig", pfNeu(planB()+"```groessenmarke\n```\n", 0), "Eingabe: Block mehrdeutig", "stderr"},
		ausgabeArt{"Eingabe Zahl", pfNeu(planB("marke: 11,00"), 0), "Eingabe: Zahl ung\xc3\xbcltig: marke", "stderr"},
		ausgabeArt{"Eingabe Pfad", pfNeu(planB("pfad: /x"), 0), "Eingabe: Pfad ung\xc3\xbcltig: /x", "stderr"},
		ausgabeArt{"Eingabe Basis", pfNeu(planB("basis: HEAD"), 0), "Eingabe: Basis ung\xc3\xbcltig: HEAD", "stderr"},
		ausgabeArt{"Eingabe Messung nicht lesbar", pfM(planB(), "unsinn"), "Eingabe: Messung nicht lesbar", "stderr"},
		ausgabeArt{"Eingabe Messung fremd", pfM(planB(), kopfText("abcdef0", spitzeB, "src/")), "Eingabe: Messung geh\xc3\xb6rt nicht zu diesem Plan", "stderr"},
		ausgabeArt{"Eingabe kein vollständiger Block", pfNeu(planB("-basis", "-schaetzung", "-marke"), 0), "Eingabe: kein vollst\xc3\xa4ndiger Block", "stderr"},
	)
	// Aufruf, Plan fehlt, Messung fehlt (stderr)
	as = append(as,
		ausgabeArt{"Aufruf", func() ergebnis { return lauf() }, "Aufruf: groessenmarke abfrage <Plan>", "stderr"},
		ausgabeArt{"Plan fehlt", func() ergebnis { return lauf("abfrage", "/gibt/es/nicht") }, "Plan fehlt: /gibt/es/nicht", "stderr"},
		ausgabeArt{"Messung fehlt", func() ergebnis { return lauf("pruefen", plan, plan+".fehlt") }, "Messung fehlt: " + plan + ".fehlt", "stderr"},
	)
	// abfrage, stdout (2)
	as = append(as,
		ausgabeArt{"abfrage keine-messung", abf("# P\n"), "keine-messung", "stdout"},
		ausgabeArt{"abfrage messen", abf(planB()), "messen\nbasis " + basisB + "\npfad src/\n", "stdout"},
	)
	// abfrage, stderr: 3 Hinweise und ein Eingabefehler (4)
	as = append(as,
		ausgabeArt{"abfrage Hinweis kein Block", abf("# P\n"), "Hinweis: kein Block", "stderr"},
		ausgabeArt{"abfrage Hinweis Prosa-Form", abf("# P\n\n" + prosaZeile + "\n"), "Hinweis: Prosa-Form", "stderr"},
		ausgabeArt{"abfrage Hinweis Marke steht aus", abf(planB("-basis", "-schaetzung", "-marke")), "Hinweis: Marke steht aus", "stderr"},
		ausgabeArt{"abfrage Eingabe", abf(strings.Replace(planB(), "| Nr. |", "| Nummer |", 1)), "Eingabe: Tabelle fehlt", "stderr"},
	)
	return as
}

func TestM10StroemeDesProgramms(t *testing.T) {
	as := ausgabeArten(schreibe(t, t.TempDir(), "slice-x.md", planB()))
	stdout, stderr := 0, 0
	for _, a := range as {
		r := a.lauf()
		in, anderer := r.out, r.err
		if a.strom == "stderr" {
			in, anderer = r.err, r.out
			stderr++
		} else {
			stdout++
		}
		if !strings.Contains(in, a.text) {
			t.Errorf("%s: %q fehlt im Strom %s\nstdout: %s\nstderr: %s", a.name, a.text, a.strom, r.out, r.err)
		}
		if strings.Contains(anderer, strings.Split(a.text, "\n")[0]) {
			t.Errorf("%s: %q steht auch im anderen Strom\nstdout: %s\nstderr: %s", a.name, a.text, r.out, r.err)
		}
	}
	if len(as) != 30 || stdout != 7 || stderr != 23 {
		t.Errorf("M10: %d Ausgabearten (%d stdout, %d stderr) statt 30 (7 und 23)", len(as), stdout, stderr)
	}
}
