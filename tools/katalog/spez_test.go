package main

import (
	"fmt"
	"strings"
	"testing"
)

const koederTabelle = "| ID | Code | Bedingung | Aktion |\n|---|---|---|---|\n| `SPEC-900` | `ODR-E9999` | x | y |\n"

// TestKatKLeser: §4 der Spezifikation, die Tabelle mit dem Kopf ID, Code, Bedingung, Aktion, der Code aus
// Zelle 2. Menge 40 Zeilen: je Code fehlt er in K, und genau die Art 4 für ihn steht in der Ausgabe. Dazu drei Köder
// (Tabelle gleicher Form in einem Abschnitt davor, einem danach und in einem Zaun im Abschnitt), die nichts ändern.
func TestKatKLeser(t *testing.T) {
	codes := codeListe()
	zaehle(t, "Codes", len(codes), 40)
	for _, e := range codes {
		m := grund(codes, nil)
		m[dSpez] = spezText(ohne(codes, e))
		erwarteBefunde(t, e, m, fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", e, zeileVon(t, m[dFehler], e)))
	}
	koeder := map[string]func(m map[string]string){
		"davor":  func(m map[string]string) { kuerze(m, dSpez, "## 3. Davor\n\nText.\n", "## 3. Davor\n\n"+koederTabelle) },
		"danach": func(m map[string]string) { m[dSpez] += "\n" + koederTabelle },
		"im Zaun": func(m map[string]string) {
			kuerze(m, dSpez, "\nText danach.\n", "\n```\n"+koederTabelle+"```\n\nText danach.\n")
		},
	}
	zaehle(t, "Köder", len(koeder), 3)
	for name, bau := range koeder {
		m := grund(codes, nil)
		bau(m)
		erwarteOk(t, name, m)
	}
}

// TestKatKForm: Form und doppelt in K. Menge 11: die acht Schreibungen der Form in einer Zelle als Inline-Code
// (die neunte, mit Zeilenumbruch, passt in keine Zeile), eine Zelle ohne Inline-Code, ein Code in zwei Zeilen und
// in dreien (ein Befund).
func TestKatKForm(t *testing.T) {
	codes := codeListe()[:3]
	n := 0
	for _, x := range formen()[:8] {
		m := grund(codes, nil)
		kuerze(m, dSpez, "\nText danach.\n", "| `SPEC-900` | `"+x+"` | b | a |\n\nText danach.\n")
		erwarteBefunde(t, strconvQuote(x), m, formQuelle("§4", zeileVon(t, m[dSpez], "SPEC-900"), x))
		n++
	}
	m := grund(codes, nil)
	kuerze(m, dSpez, "\nText danach.\n", "| `SPEC-900` | ODR-E9001 | b | a |\n\nText danach.\n")
	erwarteBefunde(t, "ohne Inline-Code", m, formQuelle("§4", zeileVon(t, m[dSpez], "SPEC-900"), "ODR-E9001"))
	n++
	for _, mal := range []int{2, 3} {
		m := grund(codes, nil)
		kuerze(m, dSpez, "\nText danach.\n", strings.Repeat("| `SPEC-901` | `"+codes[1]+"` | b | a |\n", mal-1)+"\nText danach.\n")
		erwarteBefunde(t, fmt.Sprintf("doppelt ×%d", mal), m,
			fmt.Sprintf("%s doppelt in §4, erstes Vorkommen Zeile %d", codes[1], zeileVon(t, m[dSpez], codes[1])))
		n++
	}
	zaehle(t, "Fälle", n, 11)
}

func strconvQuote(x string) string { return fmt.Sprintf("%q", x) }

// TestKatKEingabe: Eingabefehler von §4. Menge 10: Datei fehlt, Überschrift fehlt, doppelt, Tabelle fehlt, doppelt,
// Kopf weicht ab, Trennzeile fehlt, keine Datenzeile, Zeile mit 3 und mit 5 Zellen.
func TestKatKEingabe(t *testing.T) {
	codes := codeListe()[:3]
	ueberschrift := "## 4. Fehler-Codes und Logging-Felder\n"
	faelle := []struct {
		name string
		bau  func(m map[string]string)
		text func(m map[string]string) string
	}{
		{"Datei fehlt", func(m map[string]string) { delete(m, dSpez) }, func(map[string]string) string { return "spezifikation.md fehlt: spec/spezifikation.md" }},
		{"Überschrift fehlt", func(m map[string]string) { kuerze(m, dSpez, ueberschrift, "## 4. Anderes\n") }, func(map[string]string) string { return "§4: Überschrift fehlt" }},
		{"Überschrift doppelt", func(m map[string]string) { m[dSpez] += "\n" + ueberschrift }, func(map[string]string) string { return "§4: Überschrift mehrdeutig" }},
		{"Tabelle fehlt", func(m map[string]string) {
			m[dSpez] = spec110(schluesselBasis()) + "\n## 4. Fehler-Codes und Logging-Felder\n\nText.\n"
		}, func(map[string]string) string { return "§4: Tabelle fehlt" }},
		{"Tabelle doppelt", func(m map[string]string) { kuerze(m, dSpez, "\nText danach.\n", "\n"+koederTabelle+"\nText danach.\n") }, func(map[string]string) string { return "§4: Tabelle mehrdeutig" }},
		{"Kopf weicht ab", func(m map[string]string) {
			kuerze(m, dSpez, "| ID | Code | Bedingung | Aktion |", "| ID | Code | Bedingung |")
		}, func(map[string]string) string { return "§4: Kopf weicht ab" }},
		{"Trennzeile fehlt", func(m map[string]string) { kuerze(m, dSpez, "|---|---|---|---|\n", "") }, func(map[string]string) string { return "§4: Trennzeile fehlt" }},
		{"keine Datenzeile", func(m map[string]string) { m[dSpez] = spezText(nil) }, func(map[string]string) string { return "§4: keine Datenzeile" }},
		{"3 Zellen", func(m map[string]string) { kuerze(m, dSpez, "\nText danach.\n", "| a | b | c |\n\nText danach.\n") },
			func(m map[string]string) string {
				return fmt.Sprintf("§4 Zeile %d: 3 Zellen statt 4", zeileVon(t, m[dSpez], "| a | b | c |"))
			}},
		{"5 Zellen", func(m map[string]string) {
			kuerze(m, dSpez, "\nText danach.\n", "| a | b | c | d | e |\n\nText danach.\n")
		},
			func(m map[string]string) string {
				return fmt.Sprintf("§4 Zeile %d: 5 Zellen statt 4", zeileVon(t, m[dSpez], "| a | b | c | d | e |"))
			}},
	}
	zaehle(t, "Formen", len(faelle), 10)
	for _, f := range faelle {
		m := grund(codes, nil)
		f.bau(m)
		erwarteEingabe(t, f.name, m, f.text(m))
	}
}
