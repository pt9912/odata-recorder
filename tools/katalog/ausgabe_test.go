package main

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"testing"
)

// TestKatAufruf: genau ein nicht leeres Argument, sonst nur die Aufrufzeile und Exit 2; nichts wird gelesen.
// Menge 4: kein Argument, zwei, drei, ein leeres. Die Mehrfachfälle tragen als erstes Argument eine grüne
// Wurzel, die bei einem Lauf `ok` ergäbe.
func TestKatAufruf(t *testing.T) {
	w := schreibe(t, grund(codeListe(), nil))
	faelle := [][]string{{}, {w, w}, {w, w, w}, {""}}
	zaehle(t, "Aufrufe", len(faelle), 4)
	for _, a := range faelle {
		c, o, e := fahre(a...)
		if c != 2 || o != "" || e != "Aufruf: katalog <Wurzel>\n" {
			t.Errorf("%d Argumente: Exit %d stdout %q stderr %q", len(a), c, o, e)
		}
	}
}

// artFall ist eine Befundart der Ausgabe: wie sie einen grünen Baum verändert und welchen Text sie erzeugt.
type artFall struct {
	art  int
	baue func(m map[string]string)
	text func(m map[string]string, t *testing.T) string
}

func kuerze(m map[string]string, pfad, alt, neu string) {
	if strings.Count(m[pfad], alt) != 1 {
		panic("Fixture: " + alt + " nicht genau einmal in " + pfad)
	}
	m[pfad] = strings.Replace(m[pfad], alt, neu, 1)
}

// artBasis sind die vier Codes des grünen Baums der Arten; E2101 ist Ausnahme.
func artBasis() []string { return []string{"ODR-E2001", "ODR-E2101", "ODR-E2201", "ODR-E2301"} }

func artBaum() map[string]string { return grund(artBasis(), []string{"ODR-E2101"}) }

func artFaelle() []artFall {
	handbuchPunkt := "## Fehlercodes\n\nText.\n"
	return []artFall{
		{artForm, func(m map[string]string) { kuerze(m, dFehler, "}\n", "\t\"ODR-E123\",\n}\n") },
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("Form: fehler.go Zeile %d: \"ODR-E123\" hat nicht die Form", zeileVon(t, m[dFehler], "ODR-E123"))
			}},
		{artZelle, func(m map[string]string) { kuerze(m, dHandbuch, kopfH, kopfH+"| kein Name | x |\n") },
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", zeileVon(t, m[dHandbuch], "kein Name"))
			}},
		{artDoppelt, func(m map[string]string) { kuerze(m, dFehler, "}\n", "\t\"ODR-E2001\",\n}\n") },
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("ODR-E2001 doppelt in fehler.go, erstes Vorkommen Zeile %d", zeileVon(t, m[dFehler], "ODR-E2001"))
			}},
		{artOhneSpezifikation, func(m map[string]string) {
			kuerze(m, dFehler, "}\n", "\t\"ODR-E9901\",\n}\n")
			kuerze(m, dHandbuch, kopfH, kopfH+"| `ODR-E9901` | Die Bedeutung. |\n")
		},
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("ODR-E9901 in fehler.go (Zeile %d), nicht in §4", zeileVon(t, m[dFehler], "ODR-E9901"))
			}},
		{artAusnahmeOhneGrund, func(m map[string]string) {
			m[dAusnahmen] += "ODR-E2301 |\n"
			kuerze(m, dHandbuch, "| `ODR-E2301` | Die Bedeutung. |\n", "")
		},
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("Ausnahme ODR-E2301 (Zeile %d) ohne Grund", zeileVon(t, m[dAusnahmen], "ODR-E2301"))
			}},
		{artAusnahmeOhneCode, func(m map[string]string) { m[dAusnahmen] += "ODR-E9902 | Grund\n" },
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("Ausnahme ODR-E9902 (Zeile %d) steht nicht in fehler.go", zeileVon(t, m[dAusnahmen], "ODR-E9902"))
			}},
		{artFehlt, func(m map[string]string) { kuerze(m, dHandbuch, "| `ODR-E2201` | Die Bedeutung. |\n", "") },
			func(_ map[string]string, _ *testing.T) string { return "ODR-E2201 fehlt im Handbuch" }},
		{artUnbekannt, func(m map[string]string) { kuerze(m, dHandbuch, handbuchPunkt, handbuchPunkt+"Siehe ODR-E9903.\n") },
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9903, der nicht in fehler.go steht", zeileVon(t, m[dHandbuch], "ODR-E9903"))
			}},
		{artAusnahmeGenannt, func(m map[string]string) { kuerze(m, dHandbuch, handbuchPunkt, handbuchPunkt+"Siehe ODR-E2101.\n") },
			func(m map[string]string, t *testing.T) string {
				return fmt.Sprintf("Handbuch Zeile %d nennt ODR-E2101, der in den Ausnahmen steht", zeileVon(t, m[dHandbuch], "ODR-E2101"))
			}},
	}
}

// TestKatAusgabe: ohne Befund `katalog: ok` auf stdout und Exit 0; mit Befund je Zeile `katalog: Fehlercodes: <Text>`
// auf stderr und Exit 1, alle Befunde, keiner bricht ab. Menge 10: die Arten 1 bis 9 je einmal, dazu ein
// Lauf mit je einem Befund aller neun in der Reihenfolge der Arten.
func TestKatAusgabe(t *testing.T) {
	erwarteOk(t, "grüner Baum", artBaum())
	faelle := artFaelle()
	zaehle(t, "Arten", len(faelle), 9)
	alle := artBaum()
	var texte []string
	for i, f := range faelle {
		if f.art != i+1 {
			t.Fatalf("Art %d an Stelle %d", f.art, i+1)
		}
		m := artBaum()
		f.baue(m)
		erwarteBefunde(t, fmt.Sprintf("Art %d", f.art), m, f.text(m, t))
		f.baue(alle)
	}
	for _, f := range faelle {
		texte = append(texte, f.text(alle, t))
	}
	erwarteBefunde(t, "alle neun", alle, texte...)
}

// TestKatEingabeAlle: ein Eingabefehler schließt jeden Befund aus und steht allein; alle Eingabefehler stehen in der
// Reihenfolge der Quellen. Menge 7 = 6 Quellen fehlen je allein, dazu der Lauf mit allen sechs; die übrigen
// Quellen im Baum lieferten ohne den Eingabefehler Befunde in allen drei Katalogen (Codes, Schlüssel und Kommandos
// ohne Gegenstück).
func TestKatEingabeAlle(t *testing.T) {
	quellen := []struct{ pfad, text string }{
		{dFehler, "fehler.go fehlt: src/hexagon/model/fehler.go"},
		{dSpez, "spezifikation.md fehlt: spec/spezifikation.md"},
		{dAusnahmen, "Ausnahmen fehlt: harness/katalog-ausnahmen.txt"},
		{dKonfig, "konfig*.go fehlt: src/bootstrap"},
		{dCli, "cli.go fehlt: src/adapters/driving/cli/cli.go"},
		{dHandbuch, "Handbuch fehlt: docs/user/benutzerhandbuch.md"},
	}
	zaehle(t, "Quellen", len(quellen), 6)
	basis := func() map[string]string {
		m := grund(codeListe()[:4], nil)
		m[dFehler] = fehlerGo(codeListe()[:5])
		m[dSpez] = spezText(codeListe()[:3])
		m[dAusnahmen] = "ODR-E2001 |\nODR-E9999 | Grund\n"
		m[dKonfig] = konfigGo(append([]string{"EXTRA_X"}, schluesselBasis()...))
		m[dCli] = cliGo(append([]string{"stop"}, kommandoBasis()...))
		return m
	}
	c, _, e := fahre(schreibe(t, basis()))
	for _, k := range []string{"Fehlercodes", katS, "Kommandos"} {
		if c != 1 || !strings.Contains(e, "katalog: "+k+": ") {
			t.Errorf("Basis ohne Eingabefehler: Exit %d, kein Befund im Katalog %s: %q", c, k, e)
		}
	}
	var alle []string
	for _, q := range quellen {
		m := basis()
		delete(m, q.pfad)
		erwarteEingabe(t, q.pfad, m, q.text)
		alle = append(alle, q.text)
	}
	erwarteEingabe(t, "alle sechs", map[string]string{}, alle...)
}

// TestKatReihenfolge: Befunde nach Art, Name in Bytefolge, Zeile; gleiche Eingabe gleiche Ausgabe,
// unabhängig von der Reihenfolge der Zeilen der Quellen. Menge: 40 Codes mit je einem Befund der Art 4 (nicht in §4) und
// der Art 7 (nicht im Handbuch), fehler.go in 20 Permutationen (Seed fest), je 5 Läufe.
func TestKatReihenfolge(t *testing.T) {
	codes := codeListe()
	zaehle(t, "Codes", len(codes), 40)
	r := rand.New(rand.NewPCG(7, 11))
	for p := 0; p < 20; p++ {
		perm := append([]string(nil), codes...)
		r.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
		m := map[string]string{
			dFehler:    fehlerGo(perm),
			dSpez:      spezText([]string{"ODR-E1999"}),
			dAusnahmen: "",
			dHandbuch:  handbuchText(nil),
			dKonfig:    konfigGo(schluesselBasis()),
			dCli:       cliGo(kommandoBasis()),
		}
		var texte []string
		for _, c := range codes {
			texte = append(texte, fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", c, zeileVon(t, m[dFehler], c)))
		}
		for _, c := range codes {
			texte = append(texte, c+" fehlt im Handbuch")
		}
		w := schreibe(t, m)
		for l := 0; l < 5; l++ {
			c, o, e := fahre(w)
			if c != 1 || o != "" || e != befunde(texte...) {
				t.Fatalf("Permutation %d Lauf %d: Exit %d stdout %q\n%s", p, l, c, o, e)
			}
		}
	}
}

// TestKatSchreibfrei: das Programm schreibt nichts in die Wurzel. Menge 4 Läufe: ok, Befund, Eingabefehler, Aufrufverstoß.
func TestKatSchreibfrei(t *testing.T) {
	ok := grund(codeListe()[:3], nil)
	befund := grund(codeListe()[:3], nil)
	befund[dHandbuch] = handbuchText(codeListe()[:2])
	eingabe := grund(codeListe()[:3], nil)
	delete(eingabe, dAusnahmen)
	laeufe := []struct {
		name string
		m    map[string]string
		args func(w string) []string
		exit int
	}{
		{"ok", ok, func(w string) []string { return []string{w} }, 0},
		{"Befund", befund, func(w string) []string { return []string{w} }, 1},
		{"Eingabefehler", eingabe, func(w string) []string { return []string{w} }, 2},
		{"Aufrufverstoß", ok, func(w string) []string { return []string{w, w} }, 2},
	}
	zaehle(t, "Läufe", len(laeufe), 4)
	for _, l := range laeufe {
		w := schreibe(t, l.m)
		vorher := verzeichnisBaum(t, w)
		if c, _, _ := fahre(l.args(w)...); c != l.exit {
			t.Errorf("%s: Exit %d statt %d", l.name, c, l.exit)
		}
		if nachher := verzeichnisBaum(t, w); nachher != vorher {
			t.Errorf("%s: die Wurzel hat sich geändert", l.name)
		}
	}
}

// TestKatPaareAmSelbenCode: kein Befund unterdrückt einen anderen (SPEC-201 Nr. 7). Paare der Zusatztabelle.
func TestKatPaareAmSelbenCode(t *testing.T) {
	codes := codeListe()[:3]
	// Code fehlt in §4 und im Handbuch: Art 4 und Art 7.
	m := grund(codes, nil)
	m[dSpez] = spezText(codes[1:])
	m[dHandbuch] = handbuchText(codes[1:])
	erwarteBefunde(t, "Art 4 und 7", m,
		fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", codes[0], zeileVon(t, m[dFehler], codes[0])),
		codes[0]+" fehlt im Handbuch")

	// Ausnahme doppelt und nicht in fehler.go: Art 3 und Art 6.
	m = grund(codes, nil)
	m[dAusnahmen] = "ODR-E9999 | a\nODR-E9999 | b\n"
	erwarteBefunde(t, "Art 3 und 6", m,
		"ODR-E9999 doppelt in Ausnahmen, erstes Vorkommen Zeile 1",
		"Ausnahme ODR-E9999 (Zeile 1) steht nicht in fehler.go")

	// Ausnahme mit Zeile im Handbuch, fehlt zugleich in fehler.go: Art 6, 8 und 9.
	m = grund(codes, nil)
	m[dAusnahmen] = "ODR-E9999 | a\n"
	kuerze(m, dHandbuch, kopfH, kopfH+"| `ODR-E9999` | Die Bedeutung. |\n")
	n := zeileVon(t, m[dHandbuch], "ODR-E9999")
	erwarteBefunde(t, "Art 6, 8 und 9", m,
		"Ausnahme ODR-E9999 (Zeile 1) steht nicht in fehler.go",
		fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9999, der nicht in fehler.go steht", n),
		fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9999, der in den Ausnahmen steht", n))

	// Reihenfolge nach Name innerhalb der Art, dann nach Zeile.
	namen := []string{"ODR-E9903", "ODR-E9901", "ODR-E9902"}
	m = grund(codes, nil)
	for _, c := range namen {
		m[dAusnahmen] += c + " | g\n"
	}
	soll := append([]string(nil), namen...)
	sort.Strings(soll)
	var texte []string
	for _, c := range soll {
		texte = append(texte, fmt.Sprintf("Ausnahme %s (Zeile %d) steht nicht in fehler.go", c, zeileVon(t, m[dAusnahmen], c)))
	}
	erwarteBefunde(t, "Name vor Zeile", m, texte...)
}
