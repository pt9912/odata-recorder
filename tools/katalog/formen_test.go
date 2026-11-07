package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestKatZaunFormen: ein Zaun öffnet eine Zeile aus genau drei Backticks mit Info-Zeichenkette ohne Backtick, ohne
// Einrückung, und schließt die nächste Zeile aus genau drei Backticks (rechts getrimmt). Die Tabelle im Zaun zählt
// nicht, eine Tabelle außerhalb macht den Abschnitt mehrdeutig. Menge: 7 Öffnungen (4 öffnen, 3 nicht) und 6 Schlüsse
// (3 schließen, 3 nicht).
func TestKatZaunFormen(t *testing.T) {
	codes := codeListe()[:3]
	h := handbuchText(codes)
	zweite := kopfH + "| `a` | x |\n"
	oeffner := []struct {
		zeile string
		zaun  bool
	}{
		{"```", true}, {"```go", true}, {"``` x", true}, {"```\t", true},
		{"````", false}, {"``` `x", false}, {"  ```", false}, {"~~~", false},
	}
	for _, o := range oeffner {
		m := grund(codes, nil)
		if o.zaun {
			m[dHandbuch] = strings.Replace(h, abschnittKopf, abschnittKopf+"\n"+o.zeile+"\n"+zweite+"```\n", 1)
			erwarteOk(t, "Öffnung "+strconv.Quote(o.zeile), m)
			continue
		}
		m[dHandbuch] = strings.Replace(h, abschnittKopf, abschnittKopf+"\n"+o.zeile+"\n"+zweite, 1)
		erwarteEingabe(t, "Öffnung "+strconv.Quote(o.zeile), m, "Handbuch: Tabelle in Fehlercodes mehrdeutig")
	}
	schluesse := []struct {
		zeile string
		zu    bool
	}{
		{"```", true}, {"```   ", true}, {"```\t", true},
		{"  ```", false}, {"````", false}, {"``` x", false},
	}
	for _, s := range schluesse {
		m := grund(codes, nil)
		m[dHandbuch] = strings.Replace(h, abschnittKopf, abschnittKopf+"\n```\n"+zweite+s.zeile+"\n", 1)
		if s.zu {
			erwarteOk(t, "Schluss "+strconv.Quote(s.zeile), m)
			continue
		}
		erwarteEingabe(t, "Schluss "+strconv.Quote(s.zeile), m, fmt.Sprintf("Handbuch: Zaun nicht geschlossen (Zeile %d)", zeileVon(t, m[dHandbuch], "```")))
	}
}

// TestKatUeberschriftFormen: eine Überschrift ist eine Zeile außerhalb von Zäunen aus ein bis sechs `#`, einem
// Leerzeichen und dem Text, rechts getrimmt. Menge 20 Schreibungen: 13 der Zeile `Fehlercodes` im Handbuch (6 gelten,
// 7 nicht) und 7 der Zeile von §4 in der Spezifikation (3 gelten, 4 nicht).
func TestKatUeberschriftFormen(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	handbuch := []struct {
		zeile string
		gilt  bool
	}{
		{"## Fehlercodes", true}, {"## Fehlercodes \t", true}, {"# Fehlercodes", true}, {"###### Fehlercodes", true},
		{"#### Fehlercodes   ", true}, {"### Fehlercodes", true},
		{"##Fehlercodes", false}, {"##xFehlercodes", false}, {"####### Fehlercodes", false}, {"##  Fehlercodes", false}, {" ## Fehlercodes", false},
		{"## fehlercodes", false}, {"## Fehlercodes:", false},
	}
	for _, f := range handbuch {
		m := grund(codes, nil)
		kuerze(m, dHandbuch, "## Fehlercodes\n", f.zeile+"\n")
		if f.gilt {
			erwarteOk(t, strconv.Quote(f.zeile), m)
		} else {
			erwarteEingabe(t, strconv.Quote(f.zeile), m, "Handbuch: Abschnitt Fehlercodes fehlt")
		}
		n++
	}
	spez := []struct {
		zeile string
		gilt  bool
	}{
		{"## 4. Fehler-Codes und Logging-Felder", true}, {"## 4. Fehler-Codes und Logging-Felder  ", true}, {"## 4. Fehler-Codes und Logging-Felder\t", true},
		{"# 4. Fehler-Codes und Logging-Felder", false}, {"### 4. Fehler-Codes und Logging-Felder", false}, {"##4. Fehler-Codes und Logging-Felder", false},
		{"##x4. Fehler-Codes und Logging-Felder", false},
	}
	for _, f := range spez {
		m := grund(codes, nil)
		kuerze(m, dSpez, "## 4. Fehler-Codes und Logging-Felder\n", f.zeile+"\n")
		if f.gilt {
			erwarteOk(t, strconv.Quote(f.zeile), m)
		} else {
			erwarteEingabe(t, strconv.Quote(f.zeile), m, "§4: Überschrift fehlt")
		}
		n++
	}
	zaehle(t, "Schreibungen", n, 20)
}

// TestKatKAbschnittEnde: der Abschnitt §4 endet an der nächsten Zeile außerhalb von Zäunen, die mit `## ` beginnt, und
// an keiner anderen. Menge 5 Zeilen hinter der Tabelle (2 enden den Abschnitt, 3 nicht).
func TestKatKAbschnittEnde(t *testing.T) {
	codes := codeListe()[:3]
	zeilen := []struct {
		zeile string
		ende  bool
	}{
		{"## 5. Metriken", true}, {"## x", true},
		{"### 4.1 Unterabschnitt", false}, {"# Titel", false}, {"##x", false},
	}
	for _, z := range zeilen {
		m := grund(codes, nil)
		kuerze(m, dSpez, "\nText danach.\n", "\nText danach.\n\n"+z.zeile+"\n\n"+koederTabelle)
		if z.ende {
			erwarteOk(t, z.zeile, m)
		} else {
			erwarteEingabe(t, z.zeile, m, "§4: Tabelle mehrdeutig")
		}
	}
}

// TestKatPfadKomponenteDatei: eine Komponente des Pfads, die eine Datei ist, gilt wie ein fehlender Eintrag
// (`<Datei> fehlt`). Menge 4 Quellen.
func TestKatPfadKomponenteDatei(t *testing.T) {
	quellen := []struct{ pfad, wurzel, text string }{
		{dFehler, "src", "fehler.go fehlt: src/hexagon/model/fehler.go"},
		{dSpez, "spec", "spezifikation.md fehlt: spec/spezifikation.md"},
		{dAusnahmen, "harness", "Ausnahmen fehlt: harness/katalog-ausnahmen.txt"},
		{dHandbuch, "docs", "Handbuch fehlt: docs/user/benutzerhandbuch.md"},
	}
	zaehle(t, "Quellen", len(quellen), 4)
	for _, q := range quellen {
		m := grund(codeListe()[:3], nil)
		delete(m, q.pfad)
		m[q.wurzel] = "eine Datei statt eines Verzeichnisses"
		erwarteEingabe(t, q.pfad, m, q.text)
	}
}

// TestKatFormGekuerzt: der Wert in einem Befund Form steht in Go-Schreibweise und ist auf 200 Zeichen (nicht Bytes)
// gekürzt. Menge: Länge der Schreibung 199, 200, 201 und 300 Zeichen, dazu 100 Zeichen mit je zwei Bytes.
func TestKatFormGekuerzt(t *testing.T) {
	codes := codeListe()[:3]
	werte := []string{
		"ODR-" + strings.Repeat("x", 193), "ODR-" + strings.Repeat("x", 194), "ODR-" + strings.Repeat("x", 195),
		"ODR-" + strings.Repeat("x", 296), "ODR-" + strings.Repeat("ä", 100), "ODR-" + strings.Repeat("ä", 300),
	}
	soll := []int{199, 200, 200, 200, 106, 200}
	for i, v := range werte {
		m := grund(codes, nil)
		m[dFehler] = mitLiteralen(codes, "\t"+strconv.QuoteToASCII(v)+",")
		n := 7 // nach den drei Zeilen der Codes ab Zeile 4
		q := []rune(strconv.Quote(v))
		if len(q) > 200 {
			q = q[:200]
		}
		if len(q) != soll[i] {
			t.Fatalf("Fixture %d: %d Zeichen statt %d", i, len(q), soll[i])
		}
		erwarteBefunde(t, fmt.Sprintf("%d Zeichen", len(q)), m, fmt.Sprintf("Form: fehler.go Zeile %d: %s hat nicht die Form", n, string(q)))
	}
}
