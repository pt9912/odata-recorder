package main

import (
	"os"
	"strings"
)

// Fälle 209 bis 213 (SPEC-102, SPEC-212 Nr. 5 Zeilen 5 bis 9): die Ausnahme `.wt/**` in scan.ignore der
// echten .d-check.yml (Datei gdKonfigPfad im Arbeitsverzeichnis des Programms). Die Fälle 209 bis 212
// verwenden den Block scan der echten Konfiguration mit `modules: [links]` und eine Datei mit einem Verweis
// auf eine fehlende Datei; Fall 213 verwendet die Blöcke scan, structure und file wie die Fälle 166 bis
// 169. Ohne Fall bleiben Symlinks, ein fehlendes `.wt/` und ein getrackter Pfad unter `.wt/` (SPEC-212
// Nr. 6 a, d); die Module anchors, ids, matrix und codepaths lesen scan.ignore wie links, ein Fall fährt
// nur links.

// abVerweis ist der Inhalt jeder Datei der Fälle 209 bis 212: ein Verweis auf eine fehlende Datei.
const abVerweis = "[l](fehlt.md)\n"

// abBefund ist die Befund-Zeile des Verweises in pfad (stdout).
func abBefund(pfad string) string {
	return "1>" + pfad + ":1\tfehlt.md\ttarget-missing\tLinkziel existiert nicht\n"
}

// abLinks baut einen Fall der Gruppe: die Dateien, der Block scan der Konfiguration unter pfad und das
// Modul links.
func abLinks(pfad, name string, dateien map[string]string, exit int, want ...string) Fall {
	f := gdFall(name, "", nil, exit, want...)
	f.Name = "docs-check/" + name
	f.Neu = nil
	f.Vorbereitung = func(dir string) error {
		quelle, err := os.ReadFile(pfad)
		if err != nil {
			return err
		}
		d := map[string]string{".d-check.yml": gdBloecke(string(quelle), "modules: [links]\n", "scan")}
		for p, inhalt := range dateien {
			d[p] = inhalt
		}
		return gdSchreibe(dir, d)
	}
	return f
}

// arbeitsbaumFaelle sind die Fälle 209 bis 213 in der Folge des Vertrags.
func arbeitsbaumFaelle() []Fall { return arbeitsbaumFaellePfad(gdKonfigPfad) }

// arbeitsbaumFaellePfad sind die Fälle 209 bis 213 mit der echten Konfiguration unter pfad.
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
	}
}

// Fall 213: jede der 13 Dateien der Menge von Fall 167 mit 50 000 Zeilen zu 16 Bytes unter .wt/w/ ist
// ungemeldet, die Datei AGENTS.md dort mit dem Abschnitt ### 3.1 über der Grenze von structure; die
// Datei des Hauptbaums ist klein. Die Menge ist die Schleife über gdMenge().
func abFileStructure(pfad string) Fall {
	var gross []string
	for _, r := range gdMenge() {
		gross = append(gross, ".wt/w/"+r.pfad)
	}
	f := gdEcht(pfad, "arbeitsbaum-file-und-structure", 0, gross, false, false, nil,
		"0 Befund(e)", "!file-", "!section-")
	f.Name = "docs-check/arbeitsbaum-file-und-structure"
	vorher := f.Vorbereitung
	f.Vorbereitung = func(dir string) error {
		if err := vorher(dir); err != nil {
			return err
		}
		return gdSchreibe(dir, map[string]string{
			".wt/w/AGENTS.md": "### 3.1 x\n" + strings.Repeat(gdZeile16, gdGross),
		})
	}
	return f
}
