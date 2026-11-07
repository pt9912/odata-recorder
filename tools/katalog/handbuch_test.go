package main

import (
	"fmt"
	"strings"
	"testing"
)

const abschnittKopf = "## Fehlercodes\n\nText.\n"

// TestKatHandbuchLeser: Abschnitt `Fehlercodes` in jeder Ebene, genau eine Tabelle, Köder in Zäunen und hinter dem
// Abschnitt, die Formen der ersten Zelle, doppelt, `\|` und CRLF. Menge 14.
func TestKatHandbuchLeser(t *testing.T) {
	codes := codeListe()[:4]
	h := handbuchText(codes)
	erste := "| `" + codes[0] + "` | Die Bedeutung. |"
	zeile := func(m map[string]string, nadel string) int { return zeileVon(t, m[dHandbuch], nadel) }
	zelleText := func(nadel string) func(m map[string]string) []string {
		return func(m map[string]string) []string {
			return []string{fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", zeile(m, nadel))}
		}
	}
	faelle := []struct {
		name string
		text string
		soll func(m map[string]string) []string // nil: ok
	}{
		{"Ebene 2", h, nil},
		{"Ebene 3", strings.Replace(h, "## Fehlercodes", "### Fehlercodes", 1), nil},
		{"Ebene 4", strings.Replace(h, "## Fehlercodes", "#### Fehlercodes", 1), nil},
		{"Überschrift im Zaun", h + "\n```\n## Fehlercodes\n```\n", nil},
		{"Tabelle im Zaun", strings.Replace(h, abschnittKopf, abschnittKopf+"\n```\n"+kopfH+"| `a` | x |\n```\n", 1), nil},
		{"Zelle ohne Span", strings.Replace(h, kopfH, kopfH+"| kein Name | x |\n", 1), zelleText("kein Name")},
		{"Zelle mit zwei Spans", strings.Replace(h, kopfH, kopfH+"| `a` `b` | x |\n", 1), zelleText("`a` `b`")},
		{"Text hinter dem Span", strings.Replace(h, kopfH, kopfH+"| `a` Text | x |\n", 1), zelleText("`a` Text")},
		{"doppelte Zeile", strings.Replace(h, kopfH, kopfH+erste+"\n", 1), func(m map[string]string) []string {
			return []string{fmt.Sprintf("%s doppelt in Handbuch, erstes Vorkommen Zeile %d", codes[0], zeile(m, codes[0]))}
		}},
		{"Tabelle ohne Datenzeile", handbuchText(nil), func(map[string]string) []string {
			var l []string
			for _, c := range codes {
				l = append(l, c+" fehlt im Handbuch")
			}
			return l
		}},
		{"Ende der Datei ohne Schlussstrich", strings.TrimSuffix("# Benutzerhandbuch\n\n"+abschnittKopf+"\n"+kopfH+zeilenH(codes), " |\n"), nil},
		{"Strich in der Zelle", strings.Replace(h, erste, "| `"+codes[0]+"` | Die \\| Bedeutung. |", 1), nil},
		{"CRLF", strings.ReplaceAll(h, "\n", "\r\n"), nil},
		{"Abschnitt endet an der Überschrift gleicher Ebene", strings.Replace(h, "## Ende\n", "## Ende\n\n"+kopfH+"| `a` | x |\n", 1), nil},
	}
	zaehle(t, "Fälle", len(faelle), 14)
	for _, f := range faelle {
		m := grund(codes, nil)
		m[dHandbuch] = f.text
		if f.soll == nil {
			erwarteOk(t, f.name, m)
		} else {
			erwarteBefunde(t, f.name, m, f.soll(m)...)
		}
	}
}

// TestKatHandbuchEingabe: Eingabefehler des Handbuchs. Menge 13 (dazu zwei Fälle mit mehreren Gründen).
func TestKatHandbuchEingabe(t *testing.T) {
	codes := codeListe()[:3]
	h := handbuchText(codes)
	ohneTabelle := "# Benutzerhandbuch\n\n" + abschnittKopf
	faelle := []struct {
		name string
		bau  func(m map[string]string)
		soll func(m map[string]string) []string
		anf  string
	}{
		{"Datei fehlt", func(m map[string]string) { delete(m, dHandbuch) }, texte("Handbuch fehlt: docs/user/benutzerhandbuch.md"), ""},
		{"Verzeichnis", func(m map[string]string) { delete(m, dHandbuch); m[dHandbuch+"/x"] = "" }, nil, "Handbuch nicht lesbar: "},
		{"kein UTF-8", func(m map[string]string) { m[dHandbuch] = h + "\xff\n" }, texte("Handbuch nicht lesbar: kein UTF-8"), ""},
		{"NUL", func(m map[string]string) { m[dHandbuch] = h + "\x00\n" }, texte("Handbuch nicht lesbar: NUL"), ""},
		{"kein UTF-8 und NUL", func(m map[string]string) { m[dHandbuch] = h + "\xff\x00\n" }, texte("Handbuch nicht lesbar: kein UTF-8"), ""},
		{"Zaun nicht geschlossen", func(m map[string]string) { m[dHandbuch] = h + "\n```\nTeil\n" }, func(m map[string]string) []string {
			return []string{fmt.Sprintf("Handbuch: Zaun nicht geschlossen (Zeile %d)", zeileVon(t, m[dHandbuch], "```"))}
		}, ""},
		{"Abschnitt fehlt", func(m map[string]string) { m[dHandbuch] = strings.Replace(h, "## Fehlercodes", "## Anderes", 1) }, texte("Handbuch: Abschnitt Fehlercodes fehlt"), ""},
		{"Abschnitt doppelt", func(m map[string]string) { m[dHandbuch] = h + "\n" + abschnittKopf }, texte("Handbuch: Abschnitt Fehlercodes mehrdeutig"), ""},
		{"Tabelle fehlt", func(m map[string]string) { m[dHandbuch] = ohneTabelle }, texte("Handbuch: Tabelle in Fehlercodes fehlt"), ""},
		{"Tabelle doppelt", func(m map[string]string) {
			m[dHandbuch] = strings.Replace(h, "\n## Ende", "\nText.\n\n"+kopfH+"| `ODR-E9999` | x |\n\n## Ende", 1)
		}, texte("Handbuch: Tabelle in Fehlercodes mehrdeutig"), ""},
		{"Kopf weicht ab", func(m map[string]string) {
			m[dHandbuch] = strings.Replace(h, "| Code | Bedeutung |", "| Code | Erklärung |", 1)
		}, texte("Handbuch: Kopf in Fehlercodes weicht ab"), ""},
		{"Trennzeile fehlt", func(m map[string]string) { m[dHandbuch] = strings.Replace(h, "|---|---|\n", "", 1) }, texte("Handbuch: Trennzeile in Fehlercodes fehlt"), ""},
		{"Zeile mit 1 Zelle", func(m map[string]string) { m[dHandbuch] = strings.Replace(h, kopfH, kopfH+"| a |\n", 1) }, func(m map[string]string) []string {
			return []string{fmt.Sprintf("Handbuch Zeile %d: 1 Zellen statt 2", zeileVon(t, m[dHandbuch], "| a |"))}
		}, ""},
		{"Zeile mit 3 Zellen", func(m map[string]string) { m[dHandbuch] = strings.Replace(h, kopfH, kopfH+"| a | b | c |\n", 1) }, func(m map[string]string) []string {
			return []string{fmt.Sprintf("Handbuch Zeile %d: 3 Zellen statt 2", zeileVon(t, m[dHandbuch], "| a | b | c |"))}
		}, ""},
		{"zwei Zeilen mit falscher Zellenzahl", func(m map[string]string) {
			m[dHandbuch] = strings.Replace(h, kopfH, kopfH+"| a |\n| b | c | d |\n", 1)
		}, func(m map[string]string) []string {
			return []string{
				fmt.Sprintf("Handbuch Zeile %d: 1 Zellen statt 2", zeileVon(t, m[dHandbuch], "| a |")),
				fmt.Sprintf("Handbuch Zeile %d: 3 Zellen statt 2", zeileVon(t, m[dHandbuch], "| b | c | d |")),
			}
		}, ""},
	}
	zaehle(t, "Fälle", len(faelle), 15)
	for _, f := range faelle {
		m := grund(codes, nil)
		f.bau(m)
		if f.soll != nil {
			erwarteEingabe(t, f.name, m, f.soll(m)...)
			continue
		}
		c, o, e := fahre(schreibe(t, m))
		if c != 2 || o != "" || !strings.HasPrefix(e, "katalog: Eingabe: "+f.anf) || strings.Count(e, "\n") != 1 {
			t.Errorf("%s: Exit %d stdout %q stderr %q", f.name, c, o, e)
		}
	}
}

func texte(t ...string) func(map[string]string) []string {
	return func(map[string]string) []string { return t }
}

// nennorte sind die sechs Orte einer Nennung im Handbuch (SPEC-201 Nr. 3 d) und wie sie `x` in den Text bringen.
func nennorte(x, andererCode string) map[string]func(h string) string {
	return map[string]func(h string) string{
		"Absatz": func(h string) string {
			return strings.Replace(h, abschnittKopf, "## Fehlercodes\n\nText "+x+" Text.\n", 1)
		},
		"Inline-Code": func(h string) string {
			return strings.Replace(h, abschnittKopf, "## Fehlercodes\n\nText `"+x+"` Text.\n", 1)
		},
		"erste Zelle": func(h string) string { return strings.Replace(h, kopfH, kopfH+"| `"+x+"` | Die Bedeutung. |\n", 1) },
		"zweite Zelle": func(h string) string {
			return strings.Replace(h, "| `"+andererCode+"` | Die Bedeutung. |", "| `"+andererCode+"` | siehe "+x+" |", 1)
		},
		"Überschrift": func(h string) string { return h + "\n## " + x + "\n" },
		"Zaun":        func(h string) string { return h + "\n```\n" + x + "\n```\n" },
	}
}

// TestKatNennung: jede Nennung eines Codes zählt, an jedem Ort; sie ist Befund, wenn der Code nicht in fehler.go
// steht (Art 8) oder in den Ausnahmen (Art 9). Menge 12: 6 Orte × 2 Klassen.
func TestKatNennung(t *testing.T) {
	codes := codeListe()[:4]
	n := 0
	klassen := []struct {
		name    string
		x       string
		ausnahm []string
		text    string
	}{
		{"nicht in fehler.go", "ODR-E9999", nil, "Handbuch Zeile %d nennt ODR-E9999, der nicht in fehler.go steht"},
		{"in den Ausnahmen", codes[0], []string{codes[0]}, "Handbuch Zeile %d nennt " + codes[0] + ", der in den Ausnahmen steht"},
	}
	for _, k := range klassen {
		for ort, bau := range nennorte(k.x, codes[1]) {
			m := grund(codes, k.ausnahm)
			m[dHandbuch] = bau(m[dHandbuch])
			erwarteBefunde(t, k.name+" "+ort, m, fmt.Sprintf(k.text, zeileVon(t, m[dHandbuch], k.x)))
			n++
		}
	}
	zaehle(t, "Fälle", n, 12)
	// Mehrere Nennungen desselben Codes sind ein Befund an der Zeile der ersten.
	m := grund(codes, nil)
	m[dHandbuch] = strings.Replace(m[dHandbuch], abschnittKopf, "## Fehlercodes\n\nSiehe ODR-E9999.\nUnd ODR-E9999 und ODR-E9999.\n", 1)
	erwarteBefunde(t, "drei Nennungen", m, fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9999, der nicht in fehler.go steht", zeileVon(t, m[dHandbuch], "Siehe")))
}

// TestKatNennungForm: die Ziffernfolge hinter `ODR-E` gilt ganz; zu kurz oder zu lang ist Befund Form. Menge 4.
func TestKatNennungForm(t *testing.T) {
	codes := codeListe()[:3]
	faelle := []struct{ nenn, form string }{
		{"ODR-E123", "ODR-E123"}, {"ODR-E12345", "ODR-E12345"}, {"ODR-E", ""}, {codes[0] + "x", ""},
	}
	zaehle(t, "Formen", len(faelle), 4)
	for _, f := range faelle {
		m := grund(codes, nil)
		m[dHandbuch] = strings.Replace(m[dHandbuch], abschnittKopf, "## Fehlercodes\n\nSiehe "+f.nenn+" dort.\n", 1)
		if f.form == "" {
			erwarteOk(t, f.nenn, m)
			continue
		}
		erwarteBefunde(t, f.nenn, m, formQuelle("Handbuch", zeileVon(t, m[dHandbuch], "Siehe"), f.form))
	}
	// Mehrere Nennungen in einer Zeile, auch hintereinander ohne Trenner.
	m := grund(codes, nil)
	m[dHandbuch] = strings.Replace(m[dHandbuch], abschnittKopf, "## Fehlercodes\n\nODR-E12ODR-E9999 "+codes[0]+codes[1]+"\n", 1)
	n := zeileVon(t, m[dHandbuch], "ODR-E12ODR")
	erwarteBefunde(t, "mehrere in einer Zeile", m,
		formQuelle("Handbuch", n, "ODR-E12"),
		fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9999, der nicht in fehler.go steht", n))
	// Eine Form in der ersten Zelle der Tabelle ist ein Befund Form, nicht zwei.
	m = grund(codes, nil)
	kuerze(m, dHandbuch, kopfH, kopfH+"| `ODR-E123` | x |\n")
	erwarteBefunde(t, "Form in der Tabelle", m, formQuelle("Handbuch", zeileVon(t, m[dHandbuch], "ODR-E123"), "ODR-E123"))
}

// TestKatPflicht: H ist C ohne die Ausnahmen; jeder Code aus H hat eine Zeile in der Tabelle. Menge 40 Codes × 2:
// je Code fehlt seine Zeile (genau ein Befund), je Code allein als Ausnahme ohne Zeile ist grün.
func TestKatPflicht(t *testing.T) {
	codes := codeListe()
	zaehle(t, "Codes", len(codes), 40)
	for _, e := range codes {
		m := grund(codes, nil)
		kuerze(m, dHandbuch, "| `"+e+"` | Die Bedeutung. |\n", "")
		erwarteBefunde(t, "Zeile fehlt "+e, m, e+" fehlt im Handbuch")
		erwarteOk(t, "Ausnahme "+e, grund(codes, []string{e}))
	}
	// Eine Nennung in Prosa deckt die Pflicht nicht.
	m := grund(codes, nil)
	kuerze(m, dHandbuch, "| `"+codes[0]+"` | Die Bedeutung. |\n", "")
	m[dHandbuch] = strings.Replace(m[dHandbuch], abschnittKopf, "## Fehlercodes\n\nSiehe "+codes[0]+".\n", 1)
	erwarteBefunde(t, "Nennung in Prosa", m, codes[0]+" fehlt im Handbuch")
}

// TestKatCgegenK: C ohne K ist Befund (Art 4), K ohne C ist kein Befund. Menge 40 Codes × 2 Richtungen.
func TestKatCgegenK(t *testing.T) {
	codes := codeListe()
	for _, e := range codes {
		m := grund(codes, nil)
		m[dSpez] = spezText(ohne(codes, e))
		erwarteBefunde(t, "C ohne K "+e, m, fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", e, zeileVon(t, m[dFehler], e)))

		m = grund(ohne(codes, e), nil)
		m[dSpez] = spezText(codes)
		erwarteOk(t, "K ohne C "+e, m)

		// Wird ein Code aus K ohne C im Handbuch genannt, ist das Art 8, kein Sonderfall.
		kuerze(m, dHandbuch, kopfH, kopfH+"| `"+e+"` | Die Bedeutung. |\n")
		erwarteBefunde(t, "K ohne C, genannt "+e, m, fmt.Sprintf("Handbuch Zeile %d nennt %s, der nicht in fehler.go steht", zeileVon(t, m[dHandbuch], e), e))
	}
}

// TestKatZaunOhneZeile: ein Code aus H, der nur im umzäunten Block des Handbuchs steht, hat keine Zeile (Art 7); die Nennung
// im Zaun macht ihn nicht zum Befund der Arten 8 und 9. Ein Code aus den Ausnahmen mit Zeile im Handbuch ist Art 9.
func TestKatZaunOhneZeile(t *testing.T) {
	codes := codeListe()[:3]
	m := grund(codes, nil)
	kuerze(m, dHandbuch, "| `"+codes[0]+"` | Die Bedeutung. |\n", "")
	m[dHandbuch] += "\n```\n" + codes[0] + "\n```\n"
	erwarteBefunde(t, "nur im Zaun", m, codes[0]+" fehlt im Handbuch")

	m = grund(codes, []string{codes[0]})
	kuerze(m, dHandbuch, kopfH, kopfH+"| `"+codes[0]+"` | Die Bedeutung. |\n")
	erwarteBefunde(t, "Ausnahme mit Zeile", m, fmt.Sprintf("Handbuch Zeile %d nennt %s, der in den Ausnahmen steht", zeileVon(t, m[dHandbuch], codes[0]), codes[0]))
}

// TestKatZellen: Zellen einer Tabellenzeile, Trennzeile und Inline-Code.
func TestKatZellen(t *testing.T) {
	zeilen := map[string][]string{
		"| a | b |":       {"a", "b"},
		"| a | b":         {"a", "b"},
		"|a|b|":           {"a", "b"},
		"| a | b |  \t":   {"a", "b"},
		"| a \\| b | c |": {"a \\| b", "c"},
		"| a | |":         {"a", ""},
		"| a |":           {"a"},
		"|":               {""},
		"| a | b \\|":     {"a", "b \\|"},
		"| `a|b` | c |":   {"`a", "b`", "c"},
		"| a |  | c |":    {"a", "", "c"},
	}
	for z, soll := range zeilen {
		if ist := zellen(z); strings.Join(ist, "\x00") != strings.Join(soll, "\x00") || len(ist) != len(soll) {
			t.Errorf("zellen(%q) = %q, erwartet %q", z, ist, soll)
		}
	}
	for z, soll := range map[string]bool{"---": true, ":--": true, "--:": true, ":-:": true, "-": true, "": false, ":": false, "::": false, "-x-": false, "- -": false, "=": false} {
		if trennzeile([]string{z}) != soll {
			t.Errorf("trennzeile(%q) != %v", z, soll)
		}
	}
	for z, soll := range map[string]string{"`a`": "a", "`ODR-E2001`": "ODR-E2001", "``": "", "`": "", "a": "", "`a": "", "`a`b": "", "`a`b`": "", "`a b`": "a b"} {
		if t2, ok := inlineCode(z); (soll != "") != ok || (ok && t2 != soll) {
			t.Errorf("inlineCode(%q) = %q, %v", z, t2, ok)
		}
	}
}
