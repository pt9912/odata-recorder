package main

import (
	"fmt"
	"strings"
	"testing"
)

// Die Namen des grünen Baums der Schlüssel und der Kommandos.
func schluesselBasis() []string { return []string{"MODE", "LISTEN", "DATA_DIR"} }

func kommandoBasis() []string { return []string{"start", "version"} }

const (
	dKonfig = "src/bootstrap/konfig.go"
	dCli    = "src/adapters/driving/cli/cli.go"
	kopfS   = "| Schlüssel | Bedeutung |\n|---|---|\n"
	kopfC   = "| Kommando | Bedeutung |\n|---|---|\n"
	// ueberschrift110 trägt den Gedankenstrich U+2014 aus Bytewerten.
	ueberschrift110 = "### SPEC-110 \xe2\x80\x94 Startkonfiguration\n\n"
)

// spec110 ist der Abschnitt SPEC-110 mit der Zeile der Schlüssel dieser Version.
func spec110(namen []string) string {
	s := make([]string, len(namen))
	for i, n := range namen {
		s[i] = "`" + n + "`"
	}
	return ueberschrift110 + "1. **Leer.** Text.\n2. **Schlüssel dieser Version.** Geprüft werden " + strings.Join(s, ", ") +
		", und zwar unabhängig vom Modus.\n   Gesetzt (nichtleer) und nicht geliefert sind `ADMIN_*` und `STRICT_COVERAGE`: Text.\n3. **Weiter.** Text.\n"
}

// konfigLade liest je Name ein Literal als Argument.
func konfigLade(namen []string) string {
	var b strings.Builder
	b.WriteString("package bootstrap\n\nfunc lade(p *pruefer) {\n")
	for _, n := range namen {
		fmt.Fprintf(&b, "\tp.wert(%q)\n", n)
	}
	b.WriteString("}\n")
	return b.String()
}

// konfigN nennt in `nichtGelieferte` das Präfix `ADMIN_` und den Namen `STRICT_COVERAGE`.
const konfigN = "\nfunc nichtGelieferte() (praefixe, namen []string) {\n\treturn []string{\"ADMIN_\"}, []string{\"STRICT_COVERAGE\"}\n}\n"

func konfigGo(namen []string) string { return konfigLade(namen) + konfigN }

// katS ist der Name des Katalogs der Schlüssel aus Bytewerten (ü = C3 BC).
const katS = "Schl\xc3\xbcssel"

// katZeilen ist die Ausgabe auf stderr: je Text eine Zeile des Katalogs.
func katZeilen(katalog string, texte ...string) string {
	var b strings.Builder
	for _, x := range texte {
		b.WriteString("katalog: " + katalog + ": " + x + "\n")
	}
	return b.String()
}

// baumSchl ist der grüne Baum mit den Schlüsseln c in konfig.go, k in SPEC-110 und w in der Tabelle Einstellungen.
func baumSchl(c, k, w []string) map[string]string {
	m := grund(codeListe()[:3], nil)
	m[dKonfig] = konfigGo(c)
	m[dSpez] = strings.Replace(m[dSpez], spec110(schluesselBasis()), spec110(k), 1)
	m[dHandbuch] = strings.Replace(m[dHandbuch], zeilenW(schluesselBasis()), zeilenW(w), 1)
	return m
}

// baumKomm ist der grüne Baum mit den Kommandos c in cli.go und w in der Tabelle Kommandos.
func baumKomm(c, w []string) map[string]string {
	m := grund(codeListe()[:3], nil)
	m[dCli] = cliGo(c)
	m[dHandbuch] = strings.Replace(m[dHandbuch], zeilenW(kommandoBasis()), zeilenW(w), 1)
	return m
}

// nachTabelleS und nachTabelleK hängen Text hinter die Tabelle Einstellungen bzw. Kommandos des grünen Handbuchs.
func nachTabelleS(m map[string]string, text string) {
	m[dHandbuch] = strings.Replace(m[dHandbuch], zeilenW(schluesselBasis()), zeilenW(schluesselBasis())+text, 1)
}

func nachTabelleK(m map[string]string, text string) {
	m[dHandbuch] = strings.Replace(m[dHandbuch], zeilenW(kommandoBasis()), zeilenW(kommandoBasis())+text, 1)
}

// erwarteAus fährt das Programm und vergleicht Ausgangswert und stderr (stdout bleibt leer).
func erwarteAus(t *testing.T, name string, m map[string]string, code int, err string) {
	t.Helper()
	erwarte(t, name, m, code, "", err)
}

// cliGo trägt je Kommando eine Konstante und den Programmnamen.
func cliGo(kommandos []string) string {
	var b strings.Builder
	b.WriteString("package cli\n\nconst (\n\tkommandoName = \"odata-recorder\"\n")
	for i, k := range kommandos {
		fmt.Fprintf(&b, "\tk%d = %q\n", i, k)
	}
	b.WriteString(")\n")
	return b.String()
}

func zeilenW(namen []string) string {
	var b strings.Builder
	for _, n := range namen {
		fmt.Fprintf(&b, "| `%s` | Die Bedeutung. |\n", n)
	}
	return b.String()
}

// sektionen sind die Abschnitte Einstellungen und Kommandos des grünen Baums.
func sektionen() string {
	return "## Einstellungen\n\nText.\n\n" + kopfS + zeilenW(schluesselBasis()) + "\n## Kommandos\n\nText.\n\n" + kopfC + zeilenW(kommandoBasis())
}

// handbuchVorn ist das grüne Handbuch mit den Abschnitten Einstellungen und Kommandos vor dem Abschnitt Fehlercodes,
// von ihm durch einen Abschnitt getrennt: die Überschrift von Fehlercodes darf jede Ebene tragen, ohne die Tabelle
// eines anderen Abschnitts mitzutragen oder in ihn zu fallen.
func handbuchVorn(codes []string) string {
	return "# Benutzerhandbuch\n\nSoftware-Version: dev\n\n" + sektionen() + "\n## Zwischen\n\nText.\n\n## Fehlercodes\n\nText.\n\n" +
		kopfH + zeilenH(codes) + "\n## Ende\n\nText.\n"
}

// kuerzeErste ersetzt das erste Vorkommen, das mindestens einmal vorkommt.
func kuerzeErste(m map[string]string, pfad, alt, neu string) {
	if !strings.Contains(m[pfad], alt) {
		panic("Fixture: " + alt + " nicht in " + pfad)
	}
	m[pfad] = strings.Replace(m[pfad], alt, neu, 1)
}
