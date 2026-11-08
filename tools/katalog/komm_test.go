package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

const katK = "Kommandos"

// TestKatKommLeser: eine Konstante mit einem Zeichenketten-Literal als Wert, das `^[a-z]+$` trifft, ist ein Kommando
// (Nr. 5 a). Menge 16 = 6 gelesene Formen (Art 7 ohne Zeile) + 9 nicht gelesene (kein Befund) + derselbe Text in zwei
// Konstanten (ein Kommando, kein Befund).
func TestKatKommLeser(t *testing.T) {
	gelesen := []string{
		`const a = "neu"`, "const (\n\tb = \"neu\"\n)", `const c string = "neu"`, `const d, e = "neu", "start"`,
		`func f() { const g = "neu" }`, "const h = `neu`",
	}
	nicht := []string{
		`const a = "odata-recorder"`, `const a = "Start"`, `const a = "v2"`, `const a = ""`, `const a = "ne" + "u"`,
		`const a = kommandoName`, `var a = "neu"`, `func f(s string) { switch s { case "neu": } }`, `func f(s string) bool { return s == "neu" }`,
	}
	zaehle(t, "Formen", len(gelesen)+len(nicht)+1, 16)
	for _, s := range gelesen {
		m := grund(codeListe()[:3], nil)
		m[dCli] += "\n" + s + "\n"
		erwarteAus(t, s, m, 1, katZeilen(katK, "neu fehlt im Handbuch"))
	}
	for _, s := range append(nicht, `const a = "start"`) {
		m := grund(codeListe()[:3], nil)
		m[dCli] += "\n" + s + "\n"
		erwarteOk(t, s, m)
	}
}

// TestKatKommVergleich: jedes Kommando aus C hat eine Zeile (Art 7), das Handbuch nennt nur Kommandos aus C (Art 8); Kopf
// `Kommando`, `Bedeutung`; ein Schlüssel in der Tabelle ist Art 1 (Nr. 5, 6 d, 7). Menge 9.
func TestKatKommVergleich(t *testing.T) {
	zeile := func(m map[string]string, n string) int { return zeileVon(t, m[dHandbuch], "`"+n+"`") }
	n := 0
	art8 := func(name string, w []string) {
		m := baumKomm(kommandoBasis(), w)
		erwarteAus(t, "unbekannt "+name, m, 1, katZeilen(katK, fmt.Sprintf("Handbuch Zeile %d nennt halt, der nicht in cli.go steht", zeile(m, "halt"))))
		n++
	}
	for _, e := range kommandoBasis() {
		erwarteAus(t, "ohne "+e, baumKomm(kommandoBasis(), ohne(kommandoBasis(), e)), 1, katZeilen(katK, e+" fehlt im Handbuch"))
		n++
	}
	art8("erste", []string{"halt", "start", "version"})
	art8("letzte", []string{"start", "version", "halt"})
	form := baumKomm(kommandoBasis(), kommandoBasis())
	nachTabelleK(form, "| `MODE` | x |\n")
	erwarteAus(t, "Schlüssel", form, 1, katZeilen(katK, fmt.Sprintf("Form: Handbuch Zeile %d: \"MODE\" hat nicht die Form", zeileVon(t, form[dHandbuch], "| `MODE` | x |"))))
	kommando := baumKomm(kommandoBasis(), kommandoBasis())
	nachTabelleS(kommando, "| `start` | x |\n")
	erwarteAus(t, "Kommando in Einstellungen", kommando, 1, katZeilen(katS, fmt.Sprintf("Form: Handbuch Zeile %d: \"start\" hat nicht die Form", zeile(kommando, "start"))))
	kopf := baumKomm(kommandoBasis(), kommandoBasis())
	kopf[dHandbuch] = strings.Replace(kopf[dHandbuch], kopfC, kopfS, 1)
	erwarteEingabe(t, "Kopf", kopf, "Handbuch: Kopf in Kommandos weicht ab")
	ohneSpan := baumKomm(kommandoBasis(), kommandoBasis())
	nachTabelleK(ohneSpan, "| kein Name | x |\n")
	erwarteAus(t, "Zelle ohne Span", ohneSpan, 1, katZeilen(katK, fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", zeileVon(t, ohneSpan[dHandbuch], "kein Name"))))
	doppelt := baumKomm(kommandoBasis(), kommandoBasis())
	nachTabelleK(doppelt, "| `start` | x |\n")
	erwarteAus(t, "doppelt", doppelt, 1, katZeilen(katK, fmt.Sprintf("start doppelt in Handbuch, erstes Vorkommen Zeile %d", zeile(doppelt, "start"))))
	zaehle(t, "Fälle", n+5, 9)
}

// TestKatHandbuchAbschnitte: Eingabefehler je Abschnitt `Fehlercodes`, `Einstellungen`, `Kommandos` und der Datei (Nr. 6, 8).
// Menge 27 = 8 Formen × 3 Abschnitte + 3 der Datei (je mit einem fehlenden Abschnitt daneben, den die Datei ausschließt);
// dazu 2 Läufe: alle drei Abschnitte fehlen (Reihenfolge der Kataloge) und `Kommandos` als Unterabschnitt von `Einstellungen`.
func TestKatHandbuchAbschnitte(t *testing.T) {
	codes := codeListe()[:3]
	h := handbuchText(codes)
	abschnitte := []struct{ name, kopf, zeilen string }{
		{"Fehlercodes", kopfH, zeilenH(codes)},
		{"Einstellungen", kopfS, zeilenW(schluesselBasis())},
		{"Kommandos", kopfC, zeilenW(kommandoBasis())},
	}
	n := 0
	for _, a := range abschnitte {
		tab := a.kopf + a.zeilen
		formen := []struct{ name, text, soll string }{
			{"fehlt", strings.Replace(h, "## "+a.name+"\n", "## Anderes\n", 1), "Handbuch: Abschnitt " + a.name + " fehlt"},
			{"doppelt", h + "\n## " + a.name + "\n", "Handbuch: Abschnitt " + a.name + " mehrdeutig"},
			{"Tabelle fehlt", strings.Replace(h, tab, "", 1), "Handbuch: Tabelle in " + a.name + " fehlt"},
			{"Tabelle doppelt", strings.Replace(h, tab, tab+"\nText.\n\n"+a.kopf+"| `x` | y |\n", 1), "Handbuch: Tabelle in " + a.name + " mehrdeutig"},
			{"Kopf weicht ab", strings.Replace(h, tab, strings.Replace(tab, "Bedeutung", "Erkl\xc3\xa4rung", 1), 1), "Handbuch: Kopf in " + a.name + " weicht ab"},
			{"Trennzeile fehlt", strings.Replace(h, tab, strings.Replace(tab, "|---|---|\n", "", 1), 1), "Handbuch: Trennzeile in " + a.name + " fehlt"},
			{"1 Zelle", strings.Replace(h, tab, tab+"| a |\n", 1), "Handbuch Zeile %d: 1 Zellen statt 2"},
			{"3 Zellen", strings.Replace(h, tab, tab+"| a | b | c |\n", 1), "Handbuch Zeile %d: 3 Zellen statt 2"},
		}
		zaehle(t, "Formen", len(formen), 8)
		for _, f := range formen {
			m := grund(codes, nil)
			m[dHandbuch] = f.text
			soll := f.soll
			if strings.Contains(soll, "%d") {
				soll = fmt.Sprintf(soll, zeileVon(t, f.text, "| a |"))
			}
			erwarteEingabe(t, a.name+": "+f.name, m, soll)
			n++
		}
	}
	ohneKomm := strings.Replace(h, "## Kommandos\n", "## Anderes\n", 1)
	datei := []struct{ name, text, soll string }{
		{"kein UTF-8", ohneKomm + "\xff\n", "Handbuch nicht lesbar: kein UTF-8"},
		{"NUL", ohneKomm + "\x00\n", "Handbuch nicht lesbar: NUL"},
		{"Zaun nicht geschlossen", ohneKomm + "\n```\nTeil\n", ""},
	}
	for _, f := range datei {
		m := grund(codes, nil)
		m[dHandbuch] = f.text
		soll := f.soll
		if soll == "" {
			soll = fmt.Sprintf("Handbuch: Zaun nicht geschlossen (Zeile %d)", zeileVon(t, f.text, "```"))
		}
		erwarteEingabe(t, "Datei: "+f.name, m, soll)
		n++
	}
	zaehle(t, "Fälle", n, 27)
	alle := grund(codes, nil)
	alle[dHandbuch] = "# Benutzerhandbuch\n\nText.\n"
	erwarteEingabe(t, "alle drei fehlen", alle, "Handbuch: Abschnitt Fehlercodes fehlt", "Handbuch: Abschnitt Einstellungen fehlt", "Handbuch: Abschnitt Kommandos fehlt")
	unter := grund(codes, nil)
	unter[dHandbuch] = strings.Replace(h, "## Kommandos", "### Kommandos", 1)
	erwarteEingabe(t, "Kommandos unter Einstellungen", unter, "Handbuch: Tabelle in Einstellungen mehrdeutig")
}

// permutiert liefert die Elemente in einer Reihenfolge, die der Zufall mit festem Seed bestimmt.
func permutiert(r *rand.Rand, l []string) []string {
	p := append([]string(nil), l...)
	r.Shuffle(len(p), func(i, j int) { p[i], p[j] = p[j], p[i] })
	return p
}

// TestKatReihenfolgeKataloge: Befunde stehen nach Katalog (Fehlercodes, Schlüssel, Kommandos), Art, Name in Bytefolge,
// Zeile (Nr. 2 d). Menge 20 Permutationen (Seed fest) der Zeilen von fehler.go, konfig.go, cli.go, Ausnahmen und der
// Tabellen des Handbuchs bei je 6 bis 9 Befunden in jedem Katalog, 3 Läufe; in der Bytefolge steht `Kommandos` vor
// `Schlüssel`, die Reihenfolge der Kataloge steht dagegen.
func TestKatReihenfolgeKataloge(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 5))
	codes, ausn := []string{"ODR-E2001", "ODR-E2002", "ODR-E2003"}, []string{"ODR-E9001", "ODR-E9002", "ODR-E9003"}
	neuS, unbS := []string{"NEU_A", "NEU_B", "NEU_C"}, []string{"UNB_A", "UNB_B", "UNB_C"}
	neuK, unbK := []string{"alpha", "beta", "gamma"}, []string{"dos", "tres", "uno"}
	for p := 0; p < 20; p++ {
		m := grund(nil, nil)
		m[dFehler] = fehlerGo(permutiert(r, codes))
		m[dSpez] = spezText([]string{"ODR-E1999"})
		m[dAusnahmen] = ausnahmenText(permutiert(r, ausn))
		m[dKonfig] = konfigGo(permutiert(r, neuS))
		m[dCli] = cliGo(permutiert(r, neuK))
		h := handbuchText(nil)
		h = strings.Replace(h, zeilenW(schluesselBasis()), zeilenW(permutiert(r, unbS)), 1)
		m[dHandbuch] = strings.Replace(h, zeilenW(kommandoBasis()), zeilenW(permutiert(r, unbK)), 1)
		var codesT, schlT, kommT []string
		for _, c := range codes {
			codesT = append(codesT, fmt.Sprintf("%s in fehler.go (Zeile %d), nicht in §4", c, zeileVon(t, m[dFehler], c)))
		}
		for _, c := range ausn {
			codesT = append(codesT, fmt.Sprintf("Ausnahme %s (Zeile %d) steht nicht in fehler.go", c, zeileVon(t, m[dAusnahmen], c)))
		}
		for _, c := range codes {
			codesT = append(codesT, c+" fehlt im Handbuch")
		}
		for _, k := range neuS {
			schlT = append(schlT, art4(k, zeileVon(t, m[dKonfig], "\""+k+"\"")))
		}
		for _, k := range neuS {
			schlT = append(schlT, k+" fehlt im Handbuch")
		}
		for _, k := range unbS {
			schlT = append(schlT, fmt.Sprintf("Handbuch Zeile %d nennt %s, der nicht in konfig*.go steht", zeileVon(t, m[dHandbuch], "`"+k+"`"), k))
		}
		for _, k := range neuK {
			kommT = append(kommT, k+" fehlt im Handbuch")
		}
		for _, k := range unbK {
			kommT = append(kommT, fmt.Sprintf("Handbuch Zeile %d nennt %s, der nicht in cli.go steht", zeileVon(t, m[dHandbuch], "`"+k+"`"), k))
		}
		soll := befunde(codesT...) + katZeilen(katS, schlT...) + katZeilen(katK, kommT...)
		w := schreibe(t, m)
		for l := 0; l < 3; l++ {
			if c, o, e := fahre(w); c != 1 || o != "" || e != soll {
				t.Fatalf("Permutation %d Lauf %d: Exit %d stdout %q\n%s\nerwartet\n%s", p, l, c, o, e, soll)
			}
		}
	}
}

// TestKatTexteKataloge: der Wortlaut der Befunde (Nr. 7) mit dem Namen des Katalogs und der Quelle: Art 1, 2, 3, 4, 7, 8, 10
// für Schlüssel (`konfig*.go`, `SPEC-110`) und Art 1, 2, 3, 7, 8 für Kommandos (`cli.go`). Menge 12 Texte.
func TestKatTexteKataloge(t *testing.T) {
	zeileIn := func(m map[string]string, n string) int { return zeileVon(t, m[dHandbuch], n) }
	s := func(rows string) map[string]string {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		nachTabelleS(m, rows)
		return m
	}
	k := func(rows string) map[string]string {
		m := baumKomm(kommandoBasis(), kommandoBasis())
		nachTabelleK(m, rows)
		return m
	}
	falsch, zelle := "Form: Handbuch Zeile %d: \"%s\" hat nicht die Form", "Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code"
	faelle := []struct {
		name, katalog string
		m             map[string]string
		text          func(m map[string]string) string
	}{
		{"S1", katS, s("| `mode` | x |\n"), func(m map[string]string) string { return fmt.Sprintf(falsch, zeileIn(m, "`mode`"), "mode") }},
		{"S2", katS, s("| kein Name | x |\n"), func(m map[string]string) string { return fmt.Sprintf(zelle, zeileIn(m, "kein Name")) }},
		{"S3", katS, s("| `MODE` | x |\n"), func(m map[string]string) string {
			return fmt.Sprintf("MODE doppelt in Handbuch, erstes Vorkommen Zeile %d", zeileIn(m, "`MODE`"))
		}},
		{"S4", katS, baumSchl(plus(schluesselBasis(), "NEU_X"), schluesselBasis(), plus(schluesselBasis(), "NEU_X")), func(m map[string]string) string {
			return fmt.Sprintf("NEU_X in konfig*.go (Zeile %d), nicht in SPEC-110", zeileVon(t, m[dKonfig], "NEU_X"))
		}},
		{"S7", katS, baumSchl(schluesselBasis(), schluesselBasis(), ohne(schluesselBasis(), "MODE")), func(map[string]string) string { return "MODE fehlt im Handbuch" }},
		{"S8", katS, s("| `UNB_X` | x |\n"), func(m map[string]string) string {
			return fmt.Sprintf("Handbuch Zeile %d nennt UNB_X, der nicht in konfig*.go steht", zeileIn(m, "`UNB_X`"))
		}},
		{"S10", katS, s("| `ADMIN_X` | x |\n"), func(m map[string]string) string {
			return fmt.Sprintf("Handbuch Zeile %d nennt ADMIN_X, den das Binary ablehnt", zeileIn(m, "`ADMIN_X`"))
		}},
		{"K1", katK, k("| `MODE` | x |\n"), func(m map[string]string) string { return fmt.Sprintf(falsch, zeileIn(m, "| `MODE` | x |"), "MODE") }},
		{"K2", katK, k("| kein Name | x |\n"), func(m map[string]string) string { return fmt.Sprintf(zelle, zeileIn(m, "kein Name")) }},
		{"K3", katK, k("| `start` | x |\n"), func(m map[string]string) string {
			return fmt.Sprintf("start doppelt in Handbuch, erstes Vorkommen Zeile %d", zeileIn(m, "`start`"))
		}},
		{"K7", katK, baumKomm(kommandoBasis(), []string{"version"}), func(map[string]string) string { return "start fehlt im Handbuch" }},
		{"K8", katK, k("| `halt` | x |\n"), func(m map[string]string) string {
			return fmt.Sprintf("Handbuch Zeile %d nennt halt, der nicht in cli.go steht", zeileIn(m, "`halt`"))
		}},
	}
	zaehle(t, "Texte", len(faelle), 12)
	for _, f := range faelle {
		erwarteAus(t, f.name, f.m, 1, katZeilen(f.katalog, f.text(f.m)))
	}
}

// TestKatEntschieden: die entschiedenen Randformen (Nr. 4 a, 4 b, 5 a, 9 e, 9 h). Menge 5 = `//go:build ignore` wird gelesen
// (rot); das Literal `GET` ist ein Name (rot); `case "stop"` ist kein Kommando (grün); die Konstante `ja` ist ein Kommando
// (rot); ein Name zweimal in SPEC-110 (grün).
func TestKatEntschieden(t *testing.T) {
	baut := func(x string) map[string]string {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		m[dExtra] = x
		return m
	}
	m := baut("//go:build ignore\n\n" + paket + "var x = \"BAU_X\"\n")
	nachTabelleS(m, "| `BAU_X` | x |\n")
	erwarteAus(t, "Build-Bedingung", m, 1, katZeilen(katS, art4("BAU_X", 5)))
	m = baut(paket + "var x = \"GET\"\n")
	erwarteAus(t, "GET", m, 1, katZeilen(katS, art4("GET", 3), "GET fehlt im Handbuch"))
	m = grund(codeListe()[:3], nil)
	m[dCli] += "\nfunc f(s string) { switch s { case \"stop\": } }\n"
	erwarteOk(t, "case stop", m)
	m = grund(codeListe()[:3], nil)
	m[dCli] += "\nconst ja = \"ja\"\n"
	erwarteAus(t, "Konstante ja", m, 1, katZeilen(katK, "ja fehlt im Handbuch"))
	m = grund(codeListe()[:3], nil)
	m[dSpez] = strings.Replace(m[dSpez], "`MODE`", "`MODE`, `MODE`", 1)
	erwarteOk(t, "Name doppelt in SPEC-110", m)
}

// TestKatSchlPaare: die Prüfungen am selben Namen sind unabhängig (Nr. 7: kein Befund unterdrückt einen anderen). Menge 8 Paare.
func TestKatSchlPaare(t *testing.T) {
	base := schluesselBasis()
	nName := func(n string) string {
		return konfigLade(plus(base, n)) + "\nfunc nichtGelieferte() (praefixe, namen []string) {\n\treturn nil, []string{\"" + n + "\"}\n}\n"
	}
	z := func(m map[string]string, n string) int { return zeileVon(t, m[dHandbuch], n) }
	m := baumSchl(plus(base, "PAAR_X"), base, base)
	erwarteAus(t, "1 Art 4 und 7", m, 1, katZeilen(katS, art4("PAAR_X", zeileVon(t, m[dKonfig], "PAAR_X")), "PAAR_X fehlt im Handbuch"))
	m = baumSchl(base, base, plus(base, "PAAR_N"))
	m[dKonfig] = nName("PAAR_N")
	erwarteAus(t, "2 Art 4 und 10", m, 1, katZeilen(katS, art4("PAAR_N", zeileVon(t, m[dKonfig], "PAAR_N")),
		fmt.Sprintf("Handbuch Zeile %d nennt PAAR_N, den das Binary ablehnt", z(m, "`PAAR_N`"))))
	m = baumSchl(base, base, base)
	nachTabelleS(m, "| `UNB_X` | x |\n| `UNB_X` | y |\n")
	erwarteAus(t, "3 Art 3 und 8", m, 1, katZeilen(katS, fmt.Sprintf("UNB_X doppelt in Handbuch, erstes Vorkommen Zeile %d", z(m, "`UNB_X`")),
		fmt.Sprintf("Handbuch Zeile %d nennt UNB_X, der nicht in konfig*.go steht", z(m, "`UNB_X`"))))
	m = baumSchl(base, base, base)
	nachTabelleS(m, "| `ADMIN_X` | x |\n| `ADMIN_X` | y |\n")
	erwarteAus(t, "4 Art 3 und 10", m, 1, katZeilen(katS, fmt.Sprintf("ADMIN_X doppelt in Handbuch, erstes Vorkommen Zeile %d", z(m, "`ADMIN_X`")),
		fmt.Sprintf("Handbuch Zeile %d nennt ADMIN_X, den das Binary ablehnt", z(m, "`ADMIN_X`"))))
	m = baumSchl(base, base, base)
	nachTabelleS(m, "| `bad` | x |\n| `bad` | y |\n")
	erwarteAus(t, "5 zwei Art 1", m, 1, katZeilen(katS, fmt.Sprintf("Form: Handbuch Zeile %d: \"bad\" hat nicht die Form", z(m, "`bad`")),
		fmt.Sprintf("Form: Handbuch Zeile %d: \"bad\" hat nicht die Form", z(m, "`bad`")+1)))
	m = baumSchl(plus(base, "PAAR_X"), plus(base, "PAAR_X"), base)
	nachTabelleS(m, "| PAAR_X | x |\n")
	erwarteAus(t, "6 Art 2 und 7", m, 1, katZeilen(katS, fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", z(m, "| PAAR_X |")), "PAAR_X fehlt im Handbuch"))
	m = baumSchl(base, plus(base, "SERVICE_NEU"), plus(base, "SERVICE_NEU"))
	erwarteAus(t, "7 Art 8", m, 1, katZeilen(katS, fmt.Sprintf("Handbuch Zeile %d nennt SERVICE_NEU, der nicht in konfig*.go steht", z(m, "`SERVICE_NEU`"))))
	// Ein Name am Präfix `ADMIN_` aus N, in C und in K: ohne Zeile kein Befund (Art 7 entfällt), mit Zeile Art 10, nie Art 7 und 10 zugleich.
	m = baumSchl(plus(base, "ADMIN_TOKEN"), plus(base, "ADMIN_TOKEN"), base)
	erwarteOk(t, "8 Präfix ohne Zeile", m)
	m = baumSchl(plus(base, "ADMIN_TOKEN"), plus(base, "ADMIN_TOKEN"), plus(base, "ADMIN_TOKEN"))
	erwarteAus(t, "8 Präfix mit Zeile", m, 1, katZeilen(katS, fmt.Sprintf("Handbuch Zeile %d nennt ADMIN_TOKEN, den das Binary ablehnt", z(m, "`ADMIN_TOKEN`"))))
}
