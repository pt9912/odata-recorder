package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Fälle 156 bis 170 (SPEC-102, SPEC-040 Nr. 11): `d-check <Wurzel>` im Wegwerfverzeichnis des Falls.
// Die Fälle 156 bis 165 schreiben die Konfiguration des Falls mit kleinen Grenzen (5 Zeilen, 40 Bytes,
// Hard Rule 6 Zeilen), die Fälle 166 bis 170 verwenden die Blöcke scan, structure und file der echten
// .d-check.yml (Datei gdKonfigPfad im Arbeitsverzeichnis des Programms) und kleine Dateien unter den
// Pfaden der Menge. Der Inhalt jeder Datei entsteht aus Bytewerten; Zeichen jenseits von ASCII führt
// nur Fall 160 (AGENTS.md §3.18).

// gdKonfigPfad ist der Pfad der echten Konfiguration, relativ zum Arbeitsverzeichnis des Programms.
const gdKonfigPfad = ".d-check.yml"

const (
	gdLinien  = "file-lines-exceeded"
	gdBytes   = "file-bytes-exceeded"
	gdAbschn  = "section-lines-exceeded"
	gdZeile8  = "aaaaaaa\n"
	gdZeile16 = "aaaaaaaaaaaaaaa\n"
	gdGross   = 50000
)

// gdRegel ist eine Regel der Menge: ein Pfad der Menge und der Glob, der ihn trifft.
type gdRegel struct{ pfad, glob string }

// gdVier sind die vier Gruppen der Fälle 156 bis 163 mit je einem Pfad.
func gdVier() []gdRegel {
	return []gdRegel{
		{"AGENTS.md", "AGENTS.md"},
		{".claude/commands/a.md", ".claude/commands/*.md"},
		{".claude/agents/a.md", ".claude/agents/*.md"},
		{".harness/skills/a.md", ".harness/skills/*.md"},
	}
}

// gdMenge sind die 13 Dateien der Prüfmenge nach SPEC-040 Nr. 2 mit dem Glob ihrer Regel.
func gdMenge() []gdRegel {
	return []gdRegel{
		{"AGENTS.md", "AGENTS.md"},
		{".claude/commands/close-welle.md", ".claude/commands/*.md"},
		{".claude/commands/implement-slice.md", ".claude/commands/*.md"},
		{".claude/commands/plan-welle.md", ".claude/commands/*.md"},
		{".claude/agents/architect.md", ".claude/agents/*.md"},
		{".claude/agents/implementer.md", ".claude/agents/*.md"},
		{".claude/agents/planner.md", ".claude/agents/*.md"},
		{".claude/agents/reviewer.md", ".claude/agents/*.md"},
		{".claude/agents/validator.md", ".claude/agents/*.md"},
		{".claude/agents/verifier.md", ".claude/agents/*.md"},
		{".harness/skills/closure-note-reviewer.md", ".harness/skills/*.md"},
		{".harness/skills/reviewer.md", ".harness/skills/*.md"},
		{"harness/conventions.md", "harness/conventions.md"},
	}
}

// gdKonfigFile ist die Konfiguration der Fälle mit Modul file: je Glob 5 Zeilen und 40 Bytes.
func gdKonfigFile(globs ...string) string {
	var b strings.Builder
	b.WriteString("modules: [file]\nscan:\n  ignore: [\".harness/**\"]\nfile:\n")
	for _, g := range globs {
		b.WriteString("  - files: \"" + g + "\"\n    max-lines: 5\n    max-bytes: 40\n")
	}
	return b.String()
}

// gdKonfigStructure ist die Konfiguration der Fälle 164 und 165: jeder Abschnitt ### 3.<n> mit
// höchstens 6 Zeilen.
const gdKonfigStructure = "modules: [structure]\nstructure:\n  - files: \"AGENTS.md\"\n    section-pattern: '^### 3\\.\\d+ '\n    sections: each\n    max-lines: 6\n"

func gdGlobs(rs []gdRegel) []string {
	var g []string
	for _, r := range rs {
		g = append(g, r.glob)
	}
	return g
}

// gdBefund ist die Befund-Zeile von SPEC-040 Nr. 5 (stdout): Fundstelle, Glob, Code, Meldung.
func gdBefund(pfad, glob, code, meldung string) string {
	return pfad + ":1\t" + glob + "\t" + code + "\t" + meldung + "\n"
}

// gdFall baut einen Fall der Gruppe: d-check mit dem Wegwerfverzeichnis als Wurzel und der
// Konfiguration des Falls; ein Text mit `!` davor darf nicht in der Ausgabe stehen, einer mit `1>`
// steht auf stdout, sonst auf stderr.
func gdFall(name, konfig string, neu map[string]string, exit int, want ...string) Fall {
	n := map[string]string{".d-check.yml": konfig}
	for k, v := range neu {
		n[k] = v
	}
	f := gmFall(name, gateDCheck, n, []string{"{dir}"}, exit, want...)
	f.Name = "guide/" + name
	return f
}

// gdAlle legt in jede der vier Gruppen eine Datei mit dem Inhalt text.
func gdAlle(text string) map[string]string {
	m := map[string]string{}
	for _, r := range gdVier() {
		m[r.pfad] = text
	}
	return m
}

// gdSumme prüft die Zahl der Zeilen mit dem Code code auf stdout.
func gdSumme(code string, n int) func(Lauf) string {
	return func(l Lauf) string {
		if k := strings.Count(l.Out, "\t"+code+"\t"); k != n {
			return fmt.Sprintf("%d Befunde %s statt %d", k, code, n)
		}
		return ""
	}
}

// gdZeilen zählt die Zeilen von zeilen mit der Fundstelle <Pfad>:1, dem Glob der Regel, dem Code und
// dem Text `erlaubt sind`.
func gdZeilen(zeilen []string, r gdRegel, code string) int {
	n := 0
	for _, z := range zeilen {
		f := strings.Split(z, "\t")
		if len(f) == 4 && f[0] == r.pfad+":1" && f[1] == r.glob && f[2] == code && strings.Contains(f[3], "erlaubt sind") {
			n++
		}
	}
	return n
}

// gdBefunde prüft die Zeilen auf stdout: je Element von erwartet und je Code genau eine Zeile nach
// gdZeilen, und gesamt nichtleere Zeilen insgesamt.
func gdBefunde(erwartet []gdRegel, codes []string, gesamt int) func(Lauf) string {
	return func(l Lauf) string {
		zeilen := strings.Split(strings.TrimRight(l.Out, "\n"), "\n")
		for _, r := range erwartet {
			for _, code := range codes {
				n := gdZeilen(zeilen, r, code)
				if n != 1 {
					return fmt.Sprintf("%s: %d Zeilen %s statt 1", r.pfad, n, code)
				}
			}
		}
		if len(zeilen) != gesamt {
			return fmt.Sprintf("%d Zeilen auf stdout statt %d", len(zeilen), gesamt)
		}
		return ""
	}
}

// gdKette führt mehrere Prüfungen nach dem Aufruf nacheinander aus.
func gdKette(ps ...func(Lauf) string) func(Lauf) string {
	return func(l Lauf) string {
		for _, p := range ps {
			if r := p(l); r != "" {
				return r
			}
		}
		return ""
	}
}

// guideFaelle sind die Fälle 156 bis 170 in der Folge des Vertrags.
func guideFaelle() []Fall { return guideFaellePfad(gdKonfigPfad) }

// guideFaellePfad sind die Fälle 156 bis 170 mit der echten Konfiguration unter pfad.
func guideFaellePfad(pfad string) []Fall {
	return []Fall{
		gdAnDerGrenze(), gdBytePlusEins(), gdZeilePlusEins(), gdBeideUndMehrere(), gdEinheitBytes(),
		gdZeilenzaehlung(), gdDateiFehlt(), gdNeueDatei(), gdHardRuleGrenze(), gdHardRuleFehlt(),
		gdEchtUnveraendert(pfad), gdEchtJedeDatei(pfad), gdEchtNeueDatei(pfad), gdEchtOhneHardRule(pfad), gdEchtModulesZeile(pfad),
	}
}

// Fall 156: vier Dateien mit genau 5 Zeilen und 40 Bytes sind grün.
func gdAnDerGrenze() Fall {
	return gdFall("an-der-grenze", gdKonfigFile(gdGlobs(gdVier())...), gdAlle(strings.Repeat(gdZeile8, 5)), 0,
		"0 Befund(e)", "!file-")
}

// Fall 157: ein Byte darüber (41 Bytes bei 5 Zeilen) färbt jede der vier Dateien, nur im Maß Bytes.
func gdBytePlusEins() Fall {
	text := "aaaaaaaa\n" + strings.Repeat(gdZeile8, 4)
	want := []string{"4 Befund(e)", "!" + gdLinien}
	for _, r := range gdVier() {
		want = append(want, "1>"+gdBefund(r.pfad, r.glob, gdBytes, "Datei hat 41 Bytes, erlaubt sind 40"))
	}
	f := gdFall("byte-plus-eins", gdKonfigFile(gdGlobs(gdVier())...), gdAlle(text), 1, want...)
	f.Nachher = gdSumme(gdBytes, 4)
	return f
}

// Fall 158: eine Zeile darüber (6 Zeilen bei 36 Bytes) färbt jede der vier Dateien, nur im Maß Zeilen.
func gdZeilePlusEins() Fall {
	want := []string{"4 Befund(e)", "!" + gdBytes}
	for _, r := range gdVier() {
		want = append(want, "1>"+gdBefund(r.pfad, r.glob, gdLinien, "Datei hat 6 Zeilen, erlaubt sind 5"))
	}
	f := gdFall("zeile-plus-eins", gdKonfigFile(gdGlobs(gdVier())...), gdAlle(strings.Repeat("aaaaa\n", 6)), 1, want...)
	f.Nachher = gdSumme(gdLinien, 4)
	return f
}

// Fall 159: zwei Dateien einer Regel, beide über beiden Maßen: vier Befunde.
func gdBeideUndMehrere() Fall {
	text := strings.Repeat("aaaaaaaaa\n", 6)
	neu := map[string]string{".claude/commands/a.md": text, ".claude/commands/b.md": text}
	g := ".claude/commands/*.md"
	f := gdFall("beide-und-mehrere", gdKonfigFile(g), neu, 1, "4 Befund(e)",
		"1>"+gdBefund(".claude/commands/a.md", g, gdLinien, "Datei hat 6 Zeilen, erlaubt sind 5"),
		"1>"+gdBefund(".claude/commands/a.md", g, gdBytes, "Datei hat 60 Bytes, erlaubt sind 40"),
		"1>"+gdBefund(".claude/commands/b.md", g, gdLinien, "Datei hat 6 Zeilen, erlaubt sind 5"),
		"1>"+gdBefund(".claude/commands/b.md", g, gdBytes, "Datei hat 60 Bytes, erlaubt sind 40"))
	f.Nachher = gdKette(gdSumme(gdLinien, 2), gdSumme(gdBytes, 2))
	return f
}

// Fall 160: 21 Zeichen U+00E4 und ein Zeilenvorschub sind 43 Bytes; Einheit ist das Byte. Der Wert
// entsteht aus Bytewerten.
func gdEinheitBytes() Fall {
	text := strings.Repeat("\xc3\xa4", 21) + "\n"
	return gdFall("einheit-bytes", gdKonfigFile("AGENTS.md"), map[string]string{"AGENTS.md": text}, 1, "1 Befund(e)",
		"1>"+gdBefund("AGENTS.md", "AGENTS.md", gdBytes, "Datei hat 43 Bytes, erlaubt sind 40"), "!"+gdLinien)
}

// Fall 161: die letzte Zeile ohne Zeilenvorschub zählt mit; die leere Datei hat 0 Zeilen, "\n" eine.
func gdZeilenzaehlung() Fall {
	neu := map[string]string{
		"AGENTS.md":             strings.Repeat(gdZeile8, 4) + "aaaaaaa",
		".claude/commands/a.md": strings.Repeat("aaaaa\n", 5) + "aaaaa",
		".claude/agents/a.md":   "",
		".harness/skills/a.md":  "\n",
	}
	f := gdFall("zeilenzaehlung", gdKonfigFile(gdGlobs(gdVier())...), neu, 1, "1 Befund(e)",
		"1>"+gdBefund(".claude/commands/a.md", ".claude/commands/*.md", gdLinien, "Datei hat 6 Zeilen, erlaubt sind 5"))
	f.Nachher = gdKette(gdSumme(gdLinien, 1), gdSumme(gdBytes, 0), gdSumme("file-no-match", 0))
	return f
}

// Fall 162: keine Datei in einer Regel ist ein Befund file-no-match; .claude/agents/ ist leer.
func gdDateiFehlt() Fall {
	neu := map[string]string{".claude/commands/a.md": "a\n", ".harness/skills/a.md": "a\n"}
	f := gdFall("datei-fehlt", gdKonfigFile(gdGlobs(gdVier())...), neu, 1, "2 Befund(e)",
		"1>AGENTS.md:1\tAGENTS.md\tfile-no-match\t",
		"1>.claude/agents/*.md:1\t.claude/agents/*.md\tfile-no-match\t")
	f.Vorbereitung = func(dir string) error { return os.MkdirAll(filepath.Join(dir, ".claude/agents"), 0o755) }
	f.Nachher = gdSumme("file-no-match", 2)
	return f
}

// Fall 163: eine neue Datei in einem Glob erbt die Grenze der Regel, in drei Gruppen.
func gdNeueDatei() Fall {
	neu := map[string]string{}
	want := []string{"3 Befund(e)", "!neu-klein"}
	for _, r := range gdVier()[1:] {
		d := filepath.Dir(r.pfad)
		neu[d+"/neu-gross.md"] = "aaaaaaaa\n" + strings.Repeat(gdZeile8, 4)
		neu[d+"/neu-klein.md"] = strings.Repeat(gdZeile8, 5)
		want = append(want, "1>"+gdBefund(d+"/neu-gross.md", r.glob, gdBytes, "Datei hat 41 Bytes, erlaubt sind 40"))
	}
	f := gdFall("neue-datei-in-gruppe", gdKonfigFile(gdGlobs(gdVier()[1:])...), neu, 1, want...)
	f.Nachher = gdKette(gdSumme(gdBytes, 3), gdSumme(gdLinien, 0))
	return f
}

// gdAgents ist eine AGENTS.md mit den Abschnitten ### 3.1 (6 Zeilen), ### 3.2 (7), ### 3.3 (3) und
// einem Abschnitt ## 4. mit 500 Zeilen; Überschrift, Leerzeile und Text zählen, die Leerzeile vor der
// nächsten Überschrift nicht. Die Zeile von ### 3.2 ist 10.
func gdAgents() string {
	var b strings.Builder
	b.WriteString("# Briefing\n## 3. Harte Regeln\n")
	b.WriteString("### 3.1 Eins\n\n" + strings.Repeat("a\n", 4) + "\n")
	b.WriteString("### 3.2 Zwei\n\n" + strings.Repeat("a\n", 5) + "\n")
	b.WriteString("### 3.3 Drei\n\na\n\n")
	b.WriteString("## 4. Weiter\n" + strings.Repeat("a\n", 500))
	return b.String()
}

// Fall 164: ein Abschnitt mit 6 Zeilen ist grün, einer mit 7 rot auf seiner Überschriftszeile.
func gdHardRuleGrenze() Fall {
	f := gdFall("hard-rule-grenze", gdKonfigStructure, map[string]string{"AGENTS.md": gdAgents()}, 1, "1 Befund(e)",
		"1>AGENTS.md:10\tAGENTS.md :: ^### 3\\.\\d+ \t"+gdAbschn+"\tAbschnitt tr\xc3\xa4gt 7 Zeilen", "1>erlaubt sind 6\n")
	f.Nachher = gdSumme(gdAbschn, 1)
	return f
}

// Fall 165: trifft das Muster keinen Abschnitt, ist das Ziel rot.
func gdHardRuleFehlt() Fall {
	return gdFall("hard-rule-fehlt", gdKonfigStructure, map[string]string{"AGENTS.md": "# Briefing\n## 3. Harte Regeln\nText\n"}, 1,
		"1 Befund(e)", "1>\tsection-missing\t")
}

// gdReadme ist harness/README.md mit der Tabelle, die die echte Konfiguration bindet.
const gdReadme = "# Harness\n\n## Sensors (Feedback-Gates)\n\n| Target | Vertrag | Tut was | Bindung |\n|---|---|---|---|\n| `make x` | x | x | x |\n"

// gdEcht baut einen Fall mit der echten Konfiguration: 13 Stubs je "x" (AGENTS.md mit den Abschnitten
// ### 3.1 und ### 3.2, ohne Abschnitt, wenn ohneAbschnitt gilt), harness/README.md mit der Tabelle und, je Eintrag von gross,
// eine Datei mit 50 000 Zeilen zu 16 Bytes unter dem Pfad; steht ein Pfad der Menge darin, ersetzt
// sie den Stub. Gilt grossAbschnitt, hat AGENTS.md 50 000 Zeilen und 799 974 Bytes, also nicht mehr als 50 000 Zeilen und 800 000 Bytes.
func gdEcht(pfad, name string, exit int, gross []string, grossAbschnitt, ohneAbschnitt bool, nachher func(Lauf) string, want ...string) Fall {
	f := gdFall(name, "", nil, exit, want...)
	f.Neu = nil
	f.Nachher = nachher
	f.Vorbereitung = func(dir string) error {
		quelle, err := os.ReadFile(pfad)
		if err != nil {
			return err
		}
		dateien := map[string]string{".d-check.yml": gdEchtKonfig(string(quelle)), "harness/README.md": gdReadme}
		for _, r := range gdMenge() {
			dateien[r.pfad] = "x\n"
		}
		switch {
		case grossAbschnitt:
			dateien["AGENTS.md"] = "### 3.1 x\n" + strings.Repeat(gdZeile16, gdGross-3) + "### 3.2 x\nx\n"
		case ohneAbschnitt:
			dateien["AGENTS.md"] = "x\n"
		default:
			dateien["AGENTS.md"] = "### 3.1 x\nx\n### 3.2 x\nx\n"
		}
		for _, p := range gross {
			if p != "AGENTS.md" {
				dateien[p] = strings.Repeat(gdZeile16, gdGross)
			}
		}
		return gdSchreibe(dir, dateien)
	}
	return f
}

func gdSchreibe(dir string, dateien map[string]string) error {
	for p, inhalt := range dateien {
		ziel := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(ziel), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(ziel, []byte(inhalt), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// gdEchtKonfig bildet aus der echten .d-check.yml die Konfiguration der Fälle 166 bis 169: die
// Top-Level-Blöcke scan, structure und file samt Einrückung, ohne Kommentarzeilen und Leerzeilen,
// mit der Zeile `modules: [structure, file]` davor. Fehlt ein Block in der Datei, fehlt er auch in der
// Konfiguration des Falls; die Fälle 167 bis 169 melden es als fehlenden Befund.
func gdEchtKonfig(quelle string) string {
	var b strings.Builder
	b.WriteString("modules: [structure, file]\n")
	aktuell := ""
	for _, z := range strings.Split(quelle, "\n") {
		if strings.TrimSpace(z) == "" || strings.HasPrefix(strings.TrimSpace(z), "#") {
			continue
		}
		if z[0] != ' ' && z[0] != '\t' && z[0] != '-' {
			aktuell, _, _ = strings.Cut(z, ":")
		}
		switch aktuell {
		case "scan", "structure", "file":
			b.WriteString(z + "\n")
		}
	}
	return b.String()
}

// Fall 166: die echte Konfiguration ist auf kleinen Stubs grün.
func gdEchtUnveraendert(pfad string) Fall {
	return gdEcht(pfad, "echt-unveraendert", 0, nil, false, false, nil, "0 Befund(e)")
}

// Fall 167: jede der 13 Dateien mit 50 000 Zeilen zu 16 Bytes ist rot in beiden Maßen, und der
// Abschnitt ### 3.1 von AGENTS.md dazu; die Erwartung läuft als Schleife über die 13 Pfade mal die
// zwei Codes.
func gdEchtJedeDatei(pfad string) Fall {
	var pfade []string
	for _, r := range gdMenge() {
		pfade = append(pfade, r.pfad)
	}
	pruefe := gdBefunde(gdMenge(), []string{gdLinien, gdBytes}, 27)
	return gdEcht(pfad, "echt-jede-datei", 1, pfade, true, false, gdKette(pruefe, gdSumme(gdAbschn, 1)),
		"27 Befund(e)", "1>AGENTS.md:1\tAGENTS.md :: ^### 3\\.\\d+ \t"+gdAbschn+"\t")
}

// Fall 168: eine neue Datei in drei Gruppen erbt die Grenze der Regel, eine kleine nicht.
func gdEchtNeueDatei(pfad string) Fall {
	neu := []gdRegel{
		{".claude/commands/neu.md", ".claude/commands/*.md"},
		{".claude/agents/neu.md", ".claude/agents/*.md"},
		{".harness/skills/neu.md", ".harness/skills/*.md"},
	}
	var pfade []string
	for _, r := range neu {
		pfade = append(pfade, r.pfad)
	}
	f := gdEcht(pfad, "echt-neue-datei", 1, pfade, false, false, gdBefunde(neu, []string{gdLinien, gdBytes}, 6), "6 Befund(e)", "!neu-klein")
	vorher := f.Vorbereitung
	f.Vorbereitung = func(dir string) error {
		if err := vorher(dir); err != nil {
			return err
		}
		return gdSchreibe(dir, map[string]string{".claude/commands/neu-klein.md": "x\n"})
	}
	return f
}

// Fall 169: AGENTS.md ohne ### 3.<n> ist rot, mit der echten Konfiguration.
func gdEchtOhneHardRule(pfad string) Fall {
	return gdEcht(pfad, "echt-ohne-hard-rule", 1, nil, false, true, gdSumme("section-missing", 1), "1 Befund(e)", "1>\tsection-missing\t")
}

var gdWort = regexp.MustCompile(`[A-Za-z]+`)

// Fall 170: die Zeile modules: der echten .d-check.yml nennt file und structure als Wörter. Der
// Aufruf von d-check (--print-config) liest das Repo nicht; die Prüfung steht in Nachher.
func gdEchtModulesZeile(pfad string) Fall {
	f := gdFall("echt-modules-zeile", "", nil, 0)
	f.Neu = nil
	f.Argumente = []string{"-print-config"}
	f.Nachher = func(Lauf) string {
		quelle, err := os.ReadFile(pfad)
		if err != nil {
			return err.Error()
		}
		for _, z := range strings.Split(string(quelle), "\n") {
			if rest, ok := strings.CutPrefix(z, "modules:"); ok {
				haben := map[string]bool{}
				for _, w := range gdWort.FindAllString(rest, -1) {
					haben[w] = true
				}
				if haben["file"] && haben["structure"] {
					return ""
				}
				return "modules: ohne file/structure"
			}
		}
		return "modules: ohne file/structure (Zeile fehlt)"
	}
	return f
}
