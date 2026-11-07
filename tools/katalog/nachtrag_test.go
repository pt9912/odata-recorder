package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// rune als UTF-8-Bytes (AGENTS.md §3.18: kein Escape-Text, nur Bytewerte).
func utf8Text(r rune) string { return string(utf8.AppendRune(nil, r)) }

// leerraum sind die Mitglieder des Unicode-Leerraums nach SPEC-201 Nr. 2 e ohne U+000A (ein Zeilenende ist kein
// Zeichen einer Zeile): 24 Zeichen; nichtLeerraum sind drei Zeichen daneben.
func leerraum() []rune {
	l := []rune{0x09, 0x0B, 0x0C, 0x0D, 0x20, 0x85, 0xA0, 0x1680}
	for r := rune(0x2000); r <= 0x200A; r++ {
		l = append(l, r)
	}
	return append(l, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000)
}

func nichtLeerraum() []rune { return []rune{0x200B, 0x180E, 0xFEFF} }

// liesZurueck prüft auf Rohbytes, dass die geschriebene Datei den erwarteten Text trägt (AGENTS.md §3.18).
func liesZurueck(t *testing.T, w, pfad, want string) {
	t.Helper()
	roh, err := os.ReadFile(filepath.Join(w, filepath.FromSlash(pfad)))
	if err != nil || !bytes.Contains(roh, []byte(want)) {
		t.Fatalf("%s trägt die Bytes %x nicht", pfad, want)
	}
}

// TestKatDateiText: der Text der Eingabefehler `<Datei> fehlt: <Pfad>` und `<Datei> nicht lesbar: <Grund>`, bytegleich;
// der Grund ist der Text des Systemfehlers ohne Pfad (Nr. 8). Menge 8 = 4 Quellen × 2 Formen.
func TestKatDateiText(t *testing.T) {
	quellen := []struct{ pfad, datei string }{
		{dFehler, "fehler.go"}, {dSpez, "spezifikation.md"}, {dAusnahmen, "Ausnahmen"}, {dHandbuch, "Handbuch"},
	}
	n := 0
	for _, q := range quellen {
		m := grund(codeListe()[:3], nil)
		delete(m, q.pfad)
		erwarteEingabe(t, q.datei+" fehlt", m, q.datei+" fehlt: "+q.pfad)
		m = grund(codeListe()[:3], nil)
		delete(m, q.pfad)
		m[q.pfad+"/x"] = ""
		erwarteEingabe(t, q.datei+" Verzeichnis", m, q.datei+" nicht lesbar: is a directory")
		n += 2
	}
	zaehle(t, "Fälle", n, 8)
}

// TestKatKEndeZaun: eine Zeile `## ` in einem Zaun beendet §4 nicht; ein Zaun ohne Schluss reicht bis zum Dateiende und
// ist kein Eingabefehler; eine Zeile `## ` hinter dem Schluss beendet (Nr. 3 b). Menge 4.
func TestKatKEndeZaun(t *testing.T) {
	codes := codeListe()[:3]
	mit := func(einsatz string) map[string]string {
		m := grund(codes, nil)
		kuerze(m, dSpez, "\nText danach.\n", einsatz)
		return m
	}
	zaun := "\n```\n## x\n```\n\n" + koederTabelle + "\nText danach.\n"
	erwarteEingabe(t, "Zaun leer", mit(zaun), "§4: Tabelle mehrdeutig")
	erwarteEingabe(t, "Zaun mit Info", mit(strings.Replace(zaun, "```\n## x", "```go\n## x", 1)), "§4: Tabelle mehrdeutig")
	erwarteOk(t, "Zaun ohne Schluss", mit("\n```\nText danach.\n"))
	erwarteOk(t, "Ende hinter dem Schluss", mit("\n```\nx\n```\n## y\n\n"+koederTabelle+"\nText danach.\n"))
}

// TestKatTrennzeile: eine Tabelle nur aus der Kopfzeile oder mit einer zweiten Zeile, die keine Trennzeile ist, gibt
// `Trennzeile fehlt`; die Zahl der Zellen der Trennzeile wird nicht geprüft (Nr. 6 c, 8). Menge 8 = 2 Quellen × 4 Formen.
func TestKatTrennzeile(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	for _, q := range []struct {
		pfad, kopf, sep, rest, fehlt string
	}{
		{dSpez, "| ID | Code | Bedingung | Aktion |\n", "|---|---|---|---|\n", zeilenK(codes), "§4: Trennzeile fehlt"},
		{dHandbuch, "| Code | Bedeutung |\n", "|---|---|\n", zeilenH(codes), "Handbuch: Trennzeile in Fehlercodes fehlt"},
	} {
		bau := func(text string) map[string]string {
			m := grund(codes, nil)
			kuerze(m, q.pfad, q.kopf+q.sep+q.rest, text)
			return m
		}
		erwarteEingabe(t, q.pfad+" nur Kopf", bau(q.kopf), q.fehlt)
		erwarteEingabe(t, q.pfad+" zweite Zeile Daten", bau(q.kopf+q.rest), q.fehlt)
		erwarteOk(t, q.pfad+" eine Zelle", bau(q.kopf+"|---|\n"+q.rest))
		erwarteOk(t, q.pfad+" sechs Zellen", bau(q.kopf+"|---|---|---|---|---|---|\n"+q.rest))
		n += 4
	}
	zaehle(t, "Fälle", n, 8)
}

// TestKatGetrimmt: getrimmt ist der Unicode-Leerraum (Nr. 2 e). Menge 27 Zeichen (24 Mitglieder ohne U+000A, drei
// Nichtmitglieder) × 4 Orte = 108: Zelle der Tabelle, Ausnahmen-Zeile samt Zeile aus dem Zeichen allein, Ende der
// Überschriftzeilen, Ende der Schlusszeile eines Zauns.
func TestKatGetrimmt(t *testing.T) {
	codes := codeListe()[:3]
	x := codes[0]
	zeichen := map[rune]bool{}
	for _, r := range leerraum() {
		zeichen[r] = true
	}
	for _, r := range nichtLeerraum() {
		zeichen[r] = false
	}
	zaehle(t, "Zeichen", len(zeichen), 27)
	n := 0
	for r, mitglied := range zeichen {
		c := utf8Text(r)
		name := fmt.Sprintf("U+%04X", r)

		// Ort 1: Zelle der Tabelle an beiden Rändern.
		m := grund(codes, nil)
		kuerze(m, dHandbuch, "| `"+x+"` |", "| "+c+"`"+x+"`"+c+" |")
		w := schreibe(t, m)
		liesZurueck(t, w, dHandbuch, c+"`"+x+"`"+c)
		zl := zeileVon(t, m[dHandbuch], c+"`")
		c1, o1, e1 := fahre(w)
		if mitglied {
			gleichFahrt(t, name+" Zelle", c1, o1, e1, 0, "katalog: ok\n", "")
		} else {
			gleichFahrt(t, name+" Zelle", c1, o1, e1, 1, "", befunde(
				fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", zl), x+" fehlt im Handbuch"))
		}

		// Ort 2: Ausnahmen-Zeile mit Zeichen an den Rändern und neben dem Strich, dazu eine Zeile aus dem Zeichen allein.
		m = grund(codes, []string{codes[1]})
		m[dAusnahmen] = c + codes[1] + c + "|" + c + "Grund" + c + "\n" + c + "\n"
		kuerze(m, dHandbuch, "|---|---|\n", "|"+c+"---"+c+"|"+c+"---"+c+"|\n") // Trennzeile mit Zeichen an den Rändern der Zellen
		w = schreibe(t, m)
		liesZurueck(t, w, dAusnahmen, c+codes[1]+c+"|"+c+"Grund"+c)
		c2, o2, e2 := fahre(w)
		if mitglied {
			gleichFahrt(t, name+" Ausnahmen", c2, o2, e2, 0, "katalog: ok\n", "")
		} else {
			gleichFahrt(t, name+" Ausnahmen", c2, o2, e2, 2, "", eingaben("Ausnahmen Zeile 2: Zeile ohne Trennstrich", "Handbuch: Trennzeile in Fehlercodes fehlt"))
		}

		// Ort 3: Ende der Überschriftzeilen von §4 und des Abschnitts Fehlercodes.
		m = grund(codes, nil)
		kuerze(m, dSpez, "## 4. Fehler-Codes und Logging-Felder\n", "## 4. Fehler-Codes und Logging-Felder"+c+"\n")
		kuerze(m, dHandbuch, "## Fehlercodes\n", "## Fehlercodes"+c+"\n")
		w = schreibe(t, m)
		liesZurueck(t, w, dHandbuch, "## Fehlercodes"+c+"\n")
		c3, o3, e3 := fahre(w)
		if mitglied {
			gleichFahrt(t, name+" Überschrift", c3, o3, e3, 0, "katalog: ok\n", "")
		} else {
			gleichFahrt(t, name+" Überschrift", c3, o3, e3, 2, "", eingaben("§4: Überschrift fehlt", "Handbuch: Abschnitt Fehlercodes fehlt"))
		}

		// Ort 4: Ende der Schlusszeile eines Zauns mit einer zweiten Tabelle darin.
		m = grund(codes, nil)
		kuerze(m, dHandbuch, abschnittKopf, abschnittKopf+"\n```\n"+kopfH+"| `a` | x |\n```"+c+"\n")
		w = schreibe(t, m)
		liesZurueck(t, w, dHandbuch, "```"+c+"\n")
		c4, o4, e4 := fahre(w)
		if mitglied {
			gleichFahrt(t, name+" Zaun", c4, o4, e4, 0, "katalog: ok\n", "")
		} else {
			gleichFahrt(t, name+" Zaun", c4, o4, e4, 2, "", eingaben(fmt.Sprintf("Handbuch: Zaun nicht geschlossen (Zeile %d)", zeileVon(t, m[dHandbuch], "```"))))
		}
		n += 4
	}
	zaehle(t, "Fälle", n, 108)
}

func gleichFahrt(t *testing.T, name string, c int, o, e string, wc int, wo, we string) {
	t.Helper()
	if c != wc || o != wo || e != we {
		t.Errorf("%s: Exit %d stdout %q stderr %q; erwartet Exit %d stdout %q stderr %q", name, c, o, e, wc, wo, we)
	}
}

// TestKatFormZaehlung: ein Wert falscher Form ist in keiner Quelle doppelt; im Handbuch steht Form einmal je Zeile und
// Wert (Nr. 3 a, 3 c, 3 d). Menge 7: derselbe Wert zweimal in fehler.go, §4, den Ausnahmen, der Tabelle des Handbuchs
// (je zwei Befunde), derselbe Wert in der ersten Zelle und als Nennung einer Zeile (einer), zwei Werte in einer Zeile
// (zwei), derselbe Wert in zwei Nennungszeilen (zwei).
func TestKatFormZaehlung(t *testing.T) {
	codes := codeListe()[:3]
	x := "ODR-E123"
	form := func(q string, n int) string { return formQuelle(q, n, x) }
	n := 0

	m := grund(codes, nil)
	m[dFehler] = mitLiteralen(codes, "\t\""+x+"\",", "\t\""+x+"\",")
	erwarteBefunde(t, "fehler.go", m, form("fehler.go", 7), form("fehler.go", 8))
	n++

	m = grund(codes, nil)
	kuerze(m, dSpez, "\nText danach.\n", "| `SPEC-900` | `"+x+"` | b | a |\n| `SPEC-901` | `"+x+"` | b | a |\n\nText danach.\n")
	erwarteBefunde(t, "§4", m, form("§4", zeileVon(t, m[dSpez], "SPEC-900")), form("§4", zeileVon(t, m[dSpez], "SPEC-901")))
	n++

	m = grund(codes, nil)
	m[dAusnahmen] = x + " | Grund\n" + x + " | Grund\n"
	erwarteBefunde(t, "Ausnahmen", m, form("Ausnahmen", 1), form("Ausnahmen", 2))
	n++

	m = grund(codes, nil)
	kuerze(m, dHandbuch, kopfH, kopfH+"| `"+x+"` | x |\n| `"+x+"` | y |\n")
	erwarteBefunde(t, "Tabelle", m, form("Handbuch", zeileVon(t, m[dHandbuch], "| x |")), form("Handbuch", zeileVon(t, m[dHandbuch], "| y |")))
	n++

	m = grund(codes, nil)
	kuerze(m, dHandbuch, kopfH, kopfH+"| `"+x+"` | siehe "+x+" |\n")
	erwarteBefunde(t, "Zelle und Nennung", m, form("Handbuch", zeileVon(t, m[dHandbuch], "siehe")))
	n++

	m = grund(codes, nil)
	kuerze(m, dHandbuch, abschnittKopf, "## Fehlercodes\n\nODR-E123 und ODR-E12345.\n")
	z := zeileVon(t, m[dHandbuch], "und")
	erwarteBefunde(t, "zwei Werte", m, formQuelle("Handbuch", z, "ODR-E123"), formQuelle("Handbuch", z, "ODR-E12345"))
	n++

	m = grund(codes, nil)
	kuerze(m, dHandbuch, abschnittKopf, "## Fehlercodes\n\nOben ODR-E123.\nUnten ODR-E123.\n")
	erwarteBefunde(t, "zwei Zeilen", m, form("Handbuch", zeileVon(t, m[dHandbuch], "Oben")), form("Handbuch", zeileVon(t, m[dHandbuch], "Unten")))
	n++
	zaehle(t, "Fälle", n, 7)
}

// TestKatBefundPaare: die Prüfungen am selben Schlüssel sind unabhängig (Nr. 3 c, 7). Menge 5.
func TestKatBefundPaare(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	m := grund(codes, nil)
	m[dAusnahmen] = "ODR-E123 |\n"
	erwarteBefunde(t, "Form und leerer Grund", m, formQuelle("Ausnahmen", 1, "ODR-E123"), "Ausnahme ODR-E123 (Zeile 1) ohne Grund")
	n++

	m = grund(codes, nil)
	m[dAusnahmen] = "ODR-E123 | Grund\n"
	erwarteBefunde(t, "Form, nicht in C", m, formQuelle("Ausnahmen", 1, "ODR-E123"))
	n++

	m = grund(codes, nil)
	m[dAusnahmen] = "ODR-E9998 |\n"
	erwarteBefunde(t, "nicht in C, leerer Grund", m, "Ausnahme ODR-E9998 (Zeile 1) ohne Grund", "Ausnahme ODR-E9998 (Zeile 1) steht nicht in fehler.go")
	n++

	m = grund(codes, nil)
	m[dAusnahmen] = "ODR-E9998 | Grund\n"
	kuerze(m, dHandbuch, abschnittKopf, "## Fehlercodes\n\nErst ODR-E9998.\nDann ODR-E9998.\nZuletzt ODR-E9998.\n")
	z := zeileVon(t, m[dHandbuch], "Erst")
	erwarteBefunde(t, "drei Nennungen", m, "Ausnahme ODR-E9998 (Zeile 1) steht nicht in fehler.go",
		fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9998, der nicht in fehler.go steht", z),
		fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9998, der in den Ausnahmen steht", z))
	n++

	m = grund(codes, []string{codes[1]})
	m[dAusnahmen] = codes[1] + " |\n" + codes[1] + " |\n"
	erwarteBefunde(t, "doppelt ohne Grund", m, codes[1]+" doppelt in Ausnahmen, erstes Vorkommen Zeile 1",
		"Ausnahme "+codes[1]+" (Zeile 1) ohne Grund", "Ausnahme "+codes[1]+" (Zeile 2) ohne Grund")
	n++
	zaehle(t, "Fälle", n, 5)
}

// TestKatZelleOhneSpan: eine Zelle aus Backticks ohne Span ist kein Name (Nr. 3 b, 6 d). Menge 6 = 3 Zellen × 2 Orte.
func TestKatZelleOhneSpan(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	for _, zelle := range []string{"``", "```", "`"} {
		m := grund(codes, nil)
		kuerze(m, dSpez, "\nText danach.\n", "| `SPEC-900` | "+zelle+" | b | a |\n\nText danach.\n")
		erwarteBefunde(t, "§4 "+zelle, m, formQuelle("§4", zeileVon(t, m[dSpez], "SPEC-900"), zelle))

		m = grund(codes, nil)
		kuerze(m, dHandbuch, kopfH, kopfH+"| "+zelle+" | x |\n")
		erwarteBefunde(t, "Handbuch "+zelle, m, fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", zeileVon(t, m[dHandbuch], "| x |")))
		n += 2
	}
	zaehle(t, "Fälle", n, 6)
}

// fuelle hängt Zeilen an text, bis die nächste Zeile die Nummer l (ab 1) trägt.
func fuelle(t *testing.T, text string, l int) string {
	t.Helper()
	have := strings.Count(text, "\n")
	if have > l-1 {
		t.Fatalf("Text hat schon %d Zeilen, verlangt %d", have, l-1)
	}
	return text + strings.Repeat("Text.\n", l-1-have)
}

// TestKatReihenfolgeGleichstand: Befunde mit gleicher Art und gleichem Namen stehen nach Zeile als Zahl, am
// Gleichstand nach Text (Nr. 2 d). Menge 7 = 2 Arten × 3 Sprünge der Stellenzahl und ein Gleichstand in zwei Quellen.
func TestKatReihenfolgeGleichstand(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	for _, l := range []int{9, 99, 999} {
		// Art 1: derselbe Wert falscher Form in den Zeilen l und l+1 von fehler.go.
		m := grund(codes, nil)
		f := fuelle(t, "package model\n", l) + "var a = \"ODR-E123\"\nvar b = \"ODR-E123\"\n" + fehlerGo(codes)[len("package model\n"):]
		m[dFehler] = f
		erwarteBefunde(t, fmt.Sprintf("Art 1 ab Zeile %d", l), m, formQuelle("fehler.go", l, "ODR-E123"), formQuelle("fehler.go", l+1, "ODR-E123"))
		n++

		// Art 2: zwei Zeilen der Tabelle des Handbuchs ohne Name.
		m = grund(codes, nil)
		h := "# Benutzerhandbuch\n" + strings.Repeat("Text.\n", l-7) + "\n## Fehlercodes\n\n" + kopfH
		m[dHandbuch] = h + "| kein Name | x |\n| kein Name | y |\n" + zeilenH(codes)
		zelle := "Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code"
		erwarteBefunde(t, fmt.Sprintf("Art 2 ab Zeile %d", l), m, fmt.Sprintf(zelle, l), fmt.Sprintf(zelle, l+1))
		n++
	}
	// Gleichstand: derselbe Wert in Zeile 1 von fehler.go und von den Ausnahmen; der Text mit `Ausnahmen` steht vor `fehler.go`,
	// gelesen wird fehler.go zuerst.
	m := grund(codes, nil)
	m[dFehler] = "var a = \"ODR-E123\"\n" + fehlerGo(codes)
	m[dAusnahmen] = "ODR-E123 | Grund\n"
	erwarteBefunde(t, "Gleichstand", m, formQuelle("Ausnahmen", 1, "ODR-E123"), formQuelle("fehler.go", 1, "ODR-E123"))
	n++
	zaehle(t, "Fälle", n, 7)
}

// TestKatAbschnittEbene: der Abschnitt endet an der nächsten Überschrift derselben oder einer höheren Ebene, nicht an
// einer tieferen (Nr. 6 b). Menge 36 = 6 Ebenen des Abschnitts × 6 Ebenen der folgenden Überschrift; eine zweite
// Tabelle hinter der Überschrift macht den Abschnitt mehrdeutig, wenn er nicht endet.
func TestKatAbschnittEbene(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	for e := 1; e <= 6; e++ {
		for e2 := 1; e2 <= 6; e2++ {
			h := strings.Replace(handbuchText(codes), "\n## Ende\n", "\n"+strings.Repeat("#", e2)+" Danach\n\n"+kopfH+"| `a` | x |\n", 1)
			h = strings.Replace(h, "## Fehlercodes", strings.Repeat("#", e)+" Fehlercodes", 1)
			m := grund(codes, nil)
			m[dHandbuch] = h
			name := fmt.Sprintf("Ebene %d, danach %d", e, e2)
			if e2 <= e {
				erwarteOk(t, name, m)
			} else {
				erwarteEingabe(t, name, m, "Handbuch: Tabelle in Fehlercodes mehrdeutig")
			}
			n++
		}
	}
	zaehle(t, "Fälle", n, 36)
}

// TestKatWortlaut: Wortlaut-Nachbarn der Eingabe (Nr. 3 a, 3 b, 6 a, 6 c): Kopf bytegleich, Überschrift von §4 genau,
// Präfix `ODR-`, Trennzeile mit einem `:` an einem oder beiden Enden, Tabellenzeile beginnt mit `|`, Zaun in der
// ersten Zeile.
func TestKatWortlaut(t *testing.T) {
	codes := codeListe()[:3]
	n := 0

	// Kopf: 6 Zellen × 5 Abwandlungen (ohne die, die den Text nicht ändern).
	koepfe := []struct {
		pfad   string
		zellen []string
		fehlt  string
	}{
		{dSpez, []string{"ID", "Code", "Bedingung", "Aktion"}, "§4: Kopf weicht ab"},
		{dHandbuch, []string{"Code", "Bedeutung"}, "Handbuch: Kopf in Fehlercodes weicht ab"},
	}
	for _, k := range koepfe {
		alt := "| " + strings.Join(k.zellen, " | ") + " |"
		for i, z := range k.zellen {
			for _, v := range []string{strings.ToLower(z), strings.ToUpper(z), z[:1] + " " + z[1:], z + "x", "x" + z} {
				if v == z {
					continue
				}
				zs := append([]string(nil), k.zellen...)
				zs[i] = v
				m := grund(codes, nil)
				kuerze(m, k.pfad, alt, "| "+strings.Join(zs, " | ")+" |")
				erwarteEingabe(t, k.pfad+" Kopf "+v, m, k.fehlt)
				n++
			}
		}
	}
	zaehle(t, "Köpfe", n, 29)

	// Überschrift von §4: genau.
	for _, v := range []string{"## 4. Fehler-Codes und Logging-Felder und mehr", "## 4. Fehler-Codes und Logging-Felder:", "## 4. Fehler-Codes und Logging-Feld", "## 4. fehler-codes und logging-felder"} {
		m := grund(codes, nil)
		kuerze(m, dSpez, "## 4. Fehler-Codes und Logging-Felder\n", v+"\n")
		erwarteEingabe(t, v, m, "§4: Überschrift fehlt")
		n++
	}

	// Präfix ODR-: Literale, die nicht mit `ODR-` beginnen, sind keine Codes.
	for _, v := range []string{"ODR", "ODRE1234", "ODR_E1234", "ODRX-E1234", "ODR E1234", "odr-E1234"} {
		m := grund(codes, nil)
		m[dFehler] = mitLiteralen(codes, "\t\""+v+"\",")
		erwarteOk(t, v, m)
		n++
	}

	// Trennzeile: 4 gültige und 6 ungültige Zellen, je Quelle.
	for _, q := range []struct{ pfad, sep, fehlt string }{
		{dSpez, "|---|---|---|---|\n", "§4: Trennzeile fehlt"},
		{dHandbuch, "|---|---|\n", "Handbuch: Trennzeile in Fehlercodes fehlt"},
	} {
		zellen := strings.Count(q.sep, "|") - 1
		for _, s := range []string{"---", ":---", "---:", ":---:", "::---", "---::", ":::", "--:-", ":", ""} {
			gilt := s != "::---" && s != "---::" && s != ":::" && s != "--:-" && s != ":" && s != ""
			m := grund(codes, nil)
			kuerze(m, q.pfad, q.sep, "|"+strings.Repeat(s+"|", zellen)+"\n")
			if gilt {
				erwarteOk(t, q.pfad+" Trennzeile "+s, m)
			} else {
				erwarteEingabe(t, q.pfad+" Trennzeile "+s, m, q.fehlt)
			}
			n++
		}
	}

	// Tabellenzeile beginnt mit `|`: eingerückte Zeilen hinter der Tabelle gehören nicht dazu.
	einrueckungen := []string{" ", "\t", "  ", utf8Text(0xA0)}
	for _, e := range einrueckungen {
		m := grund(codes, nil)
		kuerze(m, dHandbuch, "\n## Ende", e+"| `a` | x |\n\n## Ende")
		erwarteOk(t, "Handbuch eingerückt", m)
		m = grund(codes, nil)
		kuerze(m, dSpez, "\nText danach.\n", e+"| a | b | c |\n\nText danach.\n")
		erwarteOk(t, "§4 eingerückt", m)
		n += 2
	}

	// Zaun ohne Schluss ab Zeile 1, 2 und 3 des Handbuchs.
	for l := 1; l <= 3; l++ {
		m := grund(codes, nil)
		m[dHandbuch] = strings.Repeat("Text.\n", l-1) + "```\n" + m[dHandbuch]
		erwarteEingabe(t, "Zaun ab Zeile "+strconv.Itoa(l), m, fmt.Sprintf("Handbuch: Zaun nicht geschlossen (Zeile %d)", l))
		n++
	}
	zaehle(t, "Fälle", n, 29+4+6+20+8+3)
}
