package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Fälle 209 bis 216 (SPEC-102, SPEC-212 Nr. 5 Zeilen 5 bis 12): die Ausnahme `.wt/**` in scan.ignore der
// echten .d-check.yml (Datei gdKonfigPfad im Arbeitsverzeichnis des Programms). Die Fälle 209 bis 212
// verwenden den Block scan der echten Konfiguration mit `modules: [links]` und eine Datei mit einem Verweis
// auf eine fehlende Datei. Fall 213 baut aus den Blöcken structure und file der echten Konfiguration je
// Regel und je Datei ihres Globs eine Verletzung unter `.wt/w/`; die Fälle 214 und 215 führen je lesendes
// Modul von d-check einen Defekt unter `.wt/w/` und unter `x/`, Fall 216 die Zellgrenze im Hauptbaum. Die
// Zusage über Mengen (Regeln, Module) gilt am Stand der Konfiguration: eine Regel ohne Pfad der Menge, in
// unbekannter Form oder ein Modul außerhalb beider Listen ist ein Fehler des Aufbaus. Ohne Fall bleiben
// Symlinks, ein fehlendes `.wt/` und ein getrackter Pfad unter `.wt/` (SPEC-212 Nr. 6 a, d) und die Randformen
// des Musters (tiefer verschachtelt, `.wt` im Quellbaum) für die fünf Module außer links.

// abVerweis ist der Inhalt jeder Datei der Fälle 209 bis 212: ein Verweis auf eine fehlende Datei.
const abVerweis = "[l](fehlt.md)\n"

// abBefund ist die Befund-Zeile des Verweises in pfad (stdout).
func abBefund(pfad string) string {
	return "1>" + pfad + ":1\tfehlt.md\ttarget-missing\tLinkziel existiert nicht\n"
}

// abLesend sind die Module von d-check, die scan.ignore lesen; abNichtLesend die, die es nicht lesen
// (SPEC-212 Nr. 3 b).
var (
	abLesend      = []string{"links", "anchors", "ids", "matrix", "codepaths", "spans"}
	abNichtLesend = []string{"file", "structure", "targets", "planning"}
)

// abModule liefert die Wörter der Zeile `modules:` der Konfiguration quelle.
func abModule(quelle string) ([]string, error) {
	for _, z := range strings.Split(quelle, "\n") {
		if rest, ok := strings.CutPrefix(z, "modules:"); ok {
			rest = strings.TrimSpace(rest)
			rest, ok1 := strings.CutPrefix(rest, "[")
			rest, ok2 := strings.CutSuffix(rest, "]")
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("modules: keine Liste in eckigen Klammern: %q", z)
			}
			var m []string
			for _, w := range strings.Split(rest, ",") {
				if w = strings.TrimSpace(w); w != "" {
					m = append(m, w)
				}
			}
			return m, nil
		}
	}
	return nil, fmt.Errorf("modules: Zeile fehlt")
}

// abPruefeModule meldet jedes Modul der Zeile `modules:`, das in keiner oder in beiden Listen steht, und
// jeden Eintrag einer Liste, den die Zeile nicht nennt (SPEC-212 Nr. 5 Zeile 13).
func abPruefeModule(quelle string) error {
	m, err := abModule(quelle)
	if err != nil {
		return err
	}
	for _, w := range m {
		if n := btoi(slices.Contains(abLesend, w)) + btoi(slices.Contains(abNichtLesend, w)); n != 1 {
			return fmt.Errorf("modules: %q steht in %d der zwei Listen von SPEC-212 Nr. 3 b statt in einer", w, n)
		}
	}
	for _, w := range slices.Concat(abLesend, abNichtLesend) {
		if !slices.Contains(m, w) {
			return fmt.Errorf("modules: nennt %q nicht, die Listen von SPEC-212 Nr. 3 b führen es", w)
		}
	}
	return nil
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// abSpalte ist eine Spalte einer Tabellenregel: Name und cell-max-chars.
type abSpalte struct {
	name string
	max  int
}

// abRegel ist eine Regel der Blöcke structure oder file der echten Konfiguration.
type abRegel struct {
	block, glob, muster, section string
	zeilen, bytes                int
	spalten                      []abSpalte
}

// abWert nimmt die Anführungszeichen um einen Wert.
func abWert(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

// abRegeln liest die Regeln der Blöcke structure und file aus der Konfiguration quelle. Ein Schlüssel
// außerhalb der Form, die die echte Datei führt, ist ein Fehler.
func abRegeln(quelle string) ([]abRegel, error) {
	var regeln []abRegel
	for _, block := range []string{"structure", "file"} {
		zeilen := strings.Split(strings.TrimRight(gdBloecke(quelle, "", block), "\n"), "\n")
		if zeilen[0] != block+":" {
			return nil, fmt.Errorf("Block %s fehlt", block)
		}
		ri := -1
		for _, z := range zeilen[1:] {
			k, v, ok := strings.Cut(strings.TrimPrefix(strings.TrimSpace(z), "- "), ":")
			if !ok {
				return nil, fmt.Errorf("Block %s: Zeile ohne Schlüssel: %q", block, z)
			}
			v = abWert(v)
			if k == "files" {
				regeln = append(regeln, abRegel{block: block, glob: v})
				ri = len(regeln) - 1
				continue
			}
			if ri < 0 {
				return nil, fmt.Errorf("Block %s: Schlüssel %q vor der ersten Regel", block, k)
			}
			r := &regeln[ri]
			var err error
			switch k {
			case "section":
				r.section = v
			case "section-pattern":
				r.muster = v
			case "sections":
				if v != "each" {
					err = fmt.Errorf("sections: %q", v)
				}
			case "max-lines":
				r.zeilen, err = strconv.Atoi(v)
			case "max-bytes":
				r.bytes, err = strconv.Atoi(v)
			case "table", "column":
				if v != "" {
					err = fmt.Errorf("%s: Wert %q", k, v)
				}
			case "name":
				r.spalten = append(r.spalten, abSpalte{name: v})
			case "cell-max-chars":
				if len(r.spalten) == 0 {
					err = fmt.Errorf("cell-max-chars ohne Spalte")
				} else {
					r.spalten[len(r.spalten)-1].max, err = strconv.Atoi(v)
				}
			default:
				err = fmt.Errorf("unbekannter Schlüssel")
			}
			if err != nil {
				return nil, fmt.Errorf("Block %s, Regel %q, Schlüssel %q: %w", block, r.glob, k, err)
			}
		}
	}
	return regeln, nil
}

// abUeberschriften sind die Überschriften, die ein section-pattern der Konfiguration trifft.
var abUeberschriften = map[string]string{`^### 3\.\d+ `: "### 3.1 x\n"}

// eingabe ist der Inhalt, der diese Regel an ihrem Pfad an der Wurzel verletzte: bei file 50 000 Zeilen zu
// 16 Bytes, bei section-pattern ein Abschnitt dieser Länge, bei section und table eine Tabelle unter der
// Überschrift mit Zellen von cell-max-chars + 1 Zeichen in den Spalten der Regel. Jede andere Form und jede
// Grenze, die die Eingabe nicht überschreitet, ist ein Fehler.
func (r abRegel) eingabe() (string, error) {
	gross := strings.Repeat(gdZeile16, gdGross)
	switch {
	case r.block == "file" && r.muster == "" && r.section == "" && (r.zeilen > 0 || r.bytes > 0):
		if r.zeilen >= gdGross || r.bytes >= len(gross) {
			return "", fmt.Errorf("Regel %q: Grenze %d Zeilen, %d Bytes liegt nicht unter der Eingabe", r.glob, r.zeilen, r.bytes)
		}
		return gross, nil
	case r.block == "structure" && r.muster != "" && r.section == "" && r.zeilen > 0 && len(r.spalten) == 0:
		h, ok := abUeberschriften[r.muster]
		if !ok {
			return "", fmt.Errorf("Regel %q: section-pattern %q ist dem Aufbau unbekannt", r.glob, r.muster)
		}
		if re, err := regexp.Compile(r.muster); err != nil || !re.MatchString(h) {
			return "", fmt.Errorf("Regel %q: die Überschrift %q trifft %q nicht", r.glob, h, r.muster)
		}
		if r.zeilen >= gdGross {
			return "", fmt.Errorf("Regel %q: max-lines %d liegt nicht unter der Eingabe", r.glob, r.zeilen)
		}
		return h + gross, nil
	case r.block == "structure" && r.muster == "" && r.section != "" && len(r.spalten) > 0 && r.zeilen == 0 && r.bytes == 0:
		return r.tabelle()
	}
	return "", fmt.Errorf("Regel %q des Blocks %s hat eine dem Aufbau unbekannte Form", r.glob, r.block)
}

// tabelle ist die Tabelle unter der Überschrift der Regel mit Zellen von cell-max-chars + 1 Zeichen.
func (r abRegel) tabelle() (string, error) {
	var kopf, trenn, zelle []string
	for _, s := range r.spalten {
		if s.max <= 0 {
			return "", fmt.Errorf("Regel %q: Spalte %q ohne cell-max-chars", r.glob, s.name)
		}
		kopf = append(kopf, s.name)
		trenn = append(trenn, "---")
		zelle = append(zelle, strings.Repeat("a", s.max+1))
	}
	return r.section + "\n\n| " + strings.Join(kopf, " | ") + " |\n|" + strings.Join(trenn, "|") + "|\n| " + strings.Join(zelle, " | ") + " |\n", nil
}

// abPaar ist eine Regel und ein Pfad der Menge, den ihr Glob trifft.
type abPaar struct {
	regel abRegel
	pfad  string
}

// abDateien sind die Pfade, die ein Glob der Regeln treffen kann: die 13 Dateien der Prüfmenge von file
// und harness/README.md.
func abDateien() []string {
	var p []string
	for _, r := range gdMenge() {
		p = append(p, r.pfad)
	}
	return append(p, "harness/README.md")
}

// abAnker nimmt die führenden Segmente `**/` und `*/` eines Globs, die ihn von der Wurzel lösen: der Aufbau
// wählt die Pfade der Menge nach dem Rest, und eine Regel, die `.wt/` erreicht, meldet dann die Eingabe
// unter `.wt/w/` als Befund des Falls statt als Fehler des Aufbaus.
func abAnker(glob string) string {
	for {
		if r, ok := strings.CutPrefix(glob, "**/"); ok {
			glob = r
		} else if r, ok := strings.CutPrefix(glob, "*/"); ok {
			glob = r
		} else {
			return glob
		}
	}
}

// abPaare ist die Menge der Paare Regel/Datei der Konfiguration quelle. Keine Regel oder eine Regel ohne
// Pfad in abDateien() ist ein Fehler.
func abPaare(quelle string) ([]abPaar, error) {
	regeln, err := abRegeln(quelle)
	if err != nil {
		return nil, err
	}
	if len(regeln) == 0 {
		return nil, fmt.Errorf("keine Regel in structure und file")
	}
	var paare []abPaar
	for _, r := range regeln {
		n := 0
		for _, p := range abDateien() {
			if ok, err := path.Match(abAnker(r.glob), p); err != nil {
				return nil, fmt.Errorf("Regel %q: %w", r.glob, err)
			} else if ok {
				paare = append(paare, abPaar{r, p})
				n++
			}
		}
		if n == 0 {
			return nil, fmt.Errorf("Regel %q des Blocks %s trifft keinen Pfad der Menge", r.glob, r.block)
		}
	}
	return paare, nil
}

// abLinks baut einen Fall der Gruppe: die Dateien, der Block scan der Konfiguration unter pfad und das
// Modul links.
func abLinks(pfad, name string, dateien map[string]string, exit int, want ...string) Fall {
	return abFall(pfad, name, "modules: [links]\n", []string{"scan"}, dateien, exit, nil, want...)
}

// abFall baut einen Fall der Fälle 209 bis 212, 214 und 215: kopf und die Blöcke der echten Konfiguration
// unter pfad, die Dateien; bauen prüft den Text der Konfiguration und liefert weitere Dateien.
func abFall(pfad, name, kopf string, bloecke []string, dateien map[string]string, exit int, bauen func(quelle string) (map[string]string, error), want ...string) Fall {
	f := gdFall(name, "", nil, exit, want...)
	f.Name = "docs-check/" + name
	f.Neu = nil
	f.Vorbereitung = func(dir string) error {
		quelle, err := os.ReadFile(pfad)
		if err != nil {
			return err
		}
		d := map[string]string{".d-check.yml": gdBloecke(string(quelle), kopf, bloecke...)}
		for p, inhalt := range dateien {
			d[p] = inhalt
		}
		if bauen != nil {
			mehr, err := bauen(string(quelle))
			if err != nil {
				return err
			}
			for p, inhalt := range mehr {
				d[p] = inhalt
			}
		}
		return gdSchreibe(dir, d)
	}
	return f
}

// arbeitsbaumFaelle sind die Fälle 209 bis 216 in der Folge des Vertrags.
func arbeitsbaumFaelle() []Fall { return arbeitsbaumFaellePfad(gdKonfigPfad) }

// arbeitsbaumFaellePfad sind die Fälle 209 bis 216 mit der echten Konfiguration unter pfad.
func arbeitsbaumFaellePfad(pfad string) []Fall {
	return []Fall{
		// Fall 209: der Verweis unter .wt/w/ bleibt ungemeldet.
		abLinks(pfad, "arbeitsbaum-ungemeldet", map[string]string{".wt/w/a.md": abVerweis}, 0,
			"0 Befund(e)", "!target-missing"),
		// Fall 210: derselbe Verweis in x/a.md wird gemeldet.
		abLinks(pfad, "arbeitsbaum-verweis-im-hauptbaum", map[string]string{"x/a.md": abVerweis}, 1,
			"1 Befund(e)", abBefund("x/a.md")),
		// Fall 211: ein Verzeichnis .wt unter x/ bleibt geprüft.
		abLinks(pfad, "wt-im-quellbaum-bleibt-geprueft", map[string]string{"x/.wt/b.md": abVerweis}, 1,
			"1 Befund(e)", abBefund("x/.wt/b.md")),
		// Fall 212: ein Arbeitsbaum in einem Arbeitsbaum bleibt ungemeldet.
		abLinks(pfad, "arbeitsbaum-im-arbeitsbaum", map[string]string{".wt/a/.wt/b/c.md": abVerweis}, 0,
			"0 Befund(e)", "!target-missing"),
		abFileStructure(pfad),
		abModuleFall(pfad, "arbeitsbaum-module-ungemeldet", ".wt/w/", 0, "0 Befund(e)", "!target-missing", "!anchor-missing",
			"!id-unlinked", "!codepath-missing", "!span-unclosed", "!matrix-inactive"),
		abModuleFall(pfad, "module-im-hauptbaum-gemeldet", "x/", 1, "6 Befund(e)"),
		abZellgrenze(pfad),
	}
}

// Fall 213: je Regel der Blöcke structure und file der echten Konfiguration und je Datei ihres Globs liegt
// die verletzende Eingabe unter .wt/w/<Pfad>; die Dateien des Hauptbaums sind klein. Die Menge ist die
// Schleife über abPaare der echten Konfiguration.
func abFileStructure(pfad string) Fall {
	f := gdEcht(pfad, "arbeitsbaum-file-und-structure", 0, nil, false, false, nil,
		"0 Befund(e)", "!file-", "!section-")
	f.Name = "docs-check/arbeitsbaum-file-und-structure"
	vorher := f.Vorbereitung
	f.Vorbereitung = func(dir string) error {
		if err := vorher(dir); err != nil {
			return err
		}
		quelle, err := os.ReadFile(pfad)
		if err != nil {
			return err
		}
		if err := abPruefeModule(string(quelle)); err != nil {
			return err
		}
		paare, err := abPaare(string(quelle))
		if err != nil {
			return err
		}
		inhalt := map[string]string{}
		for _, p := range paare {
			e, err := p.regel.eingabe()
			if err != nil {
				return err
			}
			inhalt[".wt/w/"+p.pfad] += e
		}
		return gdSchreibe(dir, inhalt)
	}
	return f
}

// abDefekt ist der Defekt eines lesenden Moduls: die Datei, die der Befund nennt, ihr Inhalt, der Code
// des Befunds und weitere Dateien, die der Defekt braucht.
type abDefekt struct {
	modul, datei, inhalt, code string
	mehr                       map[string]string
}

// abAdr ist die Datei, auf die matrix.md verweist: eine ADR mit dem Status Superseded.
const abAdr = "docs/plan/adr/0001-x.md"

// abDefekte sind die Defekte je lesendes Modul in der Folge von abLesend, mit den Dateien aus SPEC-212
// Nr. 5 Zeile 10.
func abDefekte() []abDefekt {
	return []abDefekt{
		{"links", "links.md", "[l](fehlt.md)\n", "target-missing", nil},
		{"anchors", "anchors.md", "[l](b.md#nope)\n", "anchor-missing", map[string]string{"b.md": "# T\n"}},
		{"ids", "ids.md", "ADR-0001\n", "id-unlinked", nil},
		{"matrix", "matrix.md", "[l](/" + abAdr + ")\n", "matrix-inactive", nil},
		{"codepaths", "codepaths.md", "`spec/fehlt.md`\n", "codepath-missing", nil},
		{"spans", "spans.md", "`offen\n", "span-unclosed", nil},
	}
}

// abDefektVon ist der Defekt des Moduls m.
func abDefektVon(m string) (abDefekt, bool) {
	i := slices.IndexFunc(abDefekte(), func(x abDefekt) bool { return x.modul == m })
	if i < 0 {
		return abDefekt{}, false
	}
	return abDefekte()[i], true
}

// abModuleDateien sind die Dateien der Defekte aller Module von abLesend unter prefix und die ADR-Datei
// der Wurzel; ein Modul ohne Defekt macht den Aufbau mit der Datei ungültig, die Fälle melden es als
// fehlenden Befund (abBefundeJeModul).
func abModuleDateien(prefix string) map[string]string {
	d := map[string]string{abAdr: "# T\n\n**Status:** Superseded\n"}
	for _, m := range abLesend {
		x, ok := abDefektVon(m)
		if !ok {
			continue
		}
		d[prefix+x.datei] = x.inhalt
		for p, inhalt := range x.mehr {
			d[prefix+p] = inhalt
		}
	}
	return d
}

// abModuleKonfig ist die Zeile modules: der Fälle 214 und 215, die sechs lesenden Module.
func abModuleKonfig() string { return "modules: [" + strings.Join(abLesend, ", ") + "]\n" }

// abModuleFall baut Fall 214 (prefix `.wt/w/`) und Fall 215 (prefix `x/`): die Blöcke scan, ids, codepaths
// und matrix der echten Konfiguration und je lesendes Modul sein Defekt unter prefix. Steht der Defekt
// im Hauptbaum, nennt jedes Modul seinen Code in der Zeile der Datei.
func abModuleFall(pfad, name, prefix string, exit int, want ...string) Fall {
	f := abFall(pfad, name, abModuleKonfig(), []string{"scan", "ids", "codepaths", "matrix"}, abModuleDateien(prefix), exit,
		func(quelle string) (map[string]string, error) { return nil, abPruefeModule(quelle) }, want...)
	if exit != 0 {
		f.Nachher = abBefundeJeModul(prefix)
	}
	return f
}

// abBefundeJeModul prüft die Zeilen auf stdout: je Modul von abLesend genau eine Zeile `<prefix><Datei>:1`
// mit seinem Code, und nicht mehr Zeilen als Module.
func abBefundeJeModul(prefix string) func(Lauf) string {
	return func(l Lauf) string {
		zeilen := strings.Split(strings.TrimRight(l.Out, "\n"), "\n")
		for _, m := range abLesend {
			x, ok := abDefektVon(m)
			if !ok {
				return "kein Defekt für " + m
			}
			n := 0
			for _, z := range zeilen {
				if f := strings.Split(z, "\t"); len(f) >= 3 && f[0] == prefix+x.datei+":1" && f[2] == x.code {
					n++
				}
			}
			if n != 1 {
				return fmt.Sprintf("%s: %d Zeilen %s statt 1", prefix+x.datei, n, x.code)
			}
		}
		if len(zeilen) != len(abLesend) {
			return fmt.Sprintf("%d Zeilen auf stdout statt %d", len(zeilen), len(abLesend))
		}
		return ""
	}
}

// Fall 216: die Gegenseite zu Fall 213 für die Zelle: harness/README.md an der Wurzel mit der Tabelle der
// ersten Tabellenregel von structure und Zellen von cell-max-chars + 1 Zeichen in jeder Spalte ergibt je
// Spalte einen Befund section-cell-oversized.
func abZellgrenze(pfad string) Fall {
	f := gdEcht(pfad, "zellgrenze-im-hauptbaum", 1, nil, false, false, gdSumme("section-cell-oversized", 2),
		"2 Befund(e)", "!file-", "!section-lines")
	f.Name = "docs-check/zellgrenze-im-hauptbaum"
	vorher := f.Vorbereitung
	f.Vorbereitung = func(dir string) error {
		if err := vorher(dir); err != nil {
			return err
		}
		quelle, err := os.ReadFile(pfad)
		if err != nil {
			return err
		}
		regeln, err := abRegeln(string(quelle))
		if err != nil {
			return err
		}
		for _, r := range regeln {
			if r.block == "structure" && r.section != "" && len(r.spalten) > 0 {
				if strings.ContainsAny(abAnker(r.glob), "*?[") {
					return fmt.Errorf("Regel %q: der Glob ist kein Pfad", r.glob)
				}
				t, err := r.tabelle()
				if err != nil {
					return err
				}
				return gdSchreibe(dir, map[string]string{abAnker(r.glob): t})
			}
		}
		return fmt.Errorf("structure: keine Regel mit section und table")
	}
	return f
}
