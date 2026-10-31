package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gdBeispiel ist eine Konfiguration mit den Blöcken der echten .d-check.yml in kleiner Form: ein
// Kommentar zwischen den Blöcken, ein Block, den die Fälle nicht übernehmen (ids), und eine Zeile
// modules:, die file und structure nennt.
const gdBeispiel = `# Kopf
scan:
  # Kommentar im Block
  ignore: [".harness/**"]
modules: [links, structure, targets, file]
ids:
  patterns:
    - {regex: 'ADR-\d{4}', target: docs/plan/adr/, link-policy: always}
structure:
  - files: "harness/README.md"
    section: "## Sensors (Feedback-Gates)"
    table:
      column:
        - name: "Vertrag"
          cell-max-chars: 200
        - name: "Tut was"
          cell-max-chars: 200
  - files: "AGENTS.md"
    section-pattern: '^### 3\.\d+ '
    sections: each
    max-lines: 29
# file
file:
  - files: "AGENTS.md"
    max-lines: 337
    max-bytes: 24648
  - files: ".claude/commands/*.md"
    max-lines: 221
    max-bytes: 16992
  - files: ".claude/agents/*.md"
    max-lines: 53
    max-bytes: 3759
  - files: ".harness/skills/*.md"
    max-lines: 137
    max-bytes: 7397
  - files: "harness/conventions.md"
    max-lines: 228
    max-bytes: 16065
`

// Die 15 Fälle 156 bis 170 tragen die Namen aus SPEC-102, jeder genau einmal und in dieser Folge, lesen
// die Ströme getrennt und stehen in faelle().
func TestGuideFaelleHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{"an-der-grenze", "byte-plus-eins", "zeile-plus-eins", "beide-und-mehrere", "einheit-bytes",
		"zeilenzaehlung", "datei-fehlt", "neue-datei-in-gruppe", "hard-rule-grenze", "hard-rule-fehlt",
		"echt-unveraendert", "echt-jede-datei", "echt-neue-datei", "echt-ohne-hard-rule", "echt-modules-zeile"}
	fs := guideFaelle()
	if len(want) != 15 || len(fs) != 15 {
		t.Fatalf("%d erwartete Namen und %d Fälle, erwartet 15", len(want), len(fs))
	}
	imGanzen := map[string]int{}
	for _, f := range faelle() {
		imGanzen[f.Name]++
	}
	for i, n := range want {
		f := fs[i]
		if f.Name != "guide/"+n {
			t.Errorf("Fall %d heißt %s statt guide/%s", 156+i, f.Name, n)
		}
		if imGanzen[f.Name] != 1 {
			t.Errorf("Fall %d (%s): %d-mal in faelle()", 156+i, f.Name, imGanzen[f.Name])
		}
		if f.Gate != gateDCheck || !f.Getrennt || !f.Wegwerfmodul || len(f.WantText) > 0 {
			t.Errorf("Fall %d: Gate %d, Getrennt %v, Wegwerfmodul %v, WantText %v", 156+i, f.Gate, f.Getrennt, f.Wegwerfmodul, f.WantText)
		}
		for datei, inhalt := range f.Neu {
			if strings.Contains(inhalt, "\\u") || strings.Contains(datei, "\\u") {
				t.Errorf("%s: %s trägt eine Escape-Folge als Text", f.Name, datei)
			}
		}
	}
}

// Die Eingaben nennen die echte Konfiguration, und die Fälle rufen d-check im Verzeichnis des Falls.
func TestStandardNenntDCheck(t *testing.T) {
	u := standard()
	have := map[string]bool{}
	for _, p := range u.Eingaben {
		have[p] = true
	}
	if !have[".d-check.yml"] {
		t.Error(".d-check.yml fehlt in standard().Eingaben")
	}
	name, args, wd := u.Befehl("/k", Fall{Gate: gateDCheck, Argumente: []string{"{dir}"}})
	if name != "d-check" || strings.Join(args, "|") != "/k" || wd != "/k" {
		t.Errorf("Programm: %q %v %q", name, args, wd)
	}
}

// Die Menge der Fälle 166 bis 168 hat 13 Pfade in 5 Gruppen; die Zählung hängt nicht an der
// Konfiguration.
func TestGuideMengeHat13PfadeIn5Gruppen(t *testing.T) {
	gruppen := map[string]int{}
	for _, r := range gdMenge() {
		gruppen[r.glob]++
	}
	if len(gdMenge()) != 13 || len(gruppen) != 5 {
		t.Errorf("%d Pfade in %d Gruppen statt 13 in 5", len(gdMenge()), len(gruppen))
	}
	for glob, n := range map[string]int{"AGENTS.md": 1, ".claude/commands/*.md": 3, ".claude/agents/*.md": 6, ".harness/skills/*.md": 2, "harness/conventions.md": 1} {
		if gruppen[glob] != n {
			t.Errorf("%s: %d Pfade statt %d", glob, gruppen[glob], n)
		}
	}
}

// Die Echt-Konfiguration übernimmt scan, structure und file ohne Kommentare, setzt die Zeile modules:
// davor und lässt jeden anderen Block weg; ein Block, der in der Datei fehlt, fehlt auch in ihr.
func TestEchtKonfigNimmtDreiBloecke(t *testing.T) {
	k := gdEchtKonfig(gdBeispiel)
	if !strings.HasPrefix(k, "modules: [structure, file]\nscan:\n  ignore: [\".harness/**\"]\nstructure:\n") {
		t.Errorf("Anfang: %q", k[:min(len(k), 120)])
	}
	for _, z := range strings.Split(k, "\n") {
		if strings.HasPrefix(strings.TrimSpace(z), "#") {
			t.Errorf("Kommentarzeile %q", z)
		}
	}
	for _, nein := range []string{"ids:", "ADR-", "links", "targets", "Kopf"} {
		if strings.Contains(k, nein) {
			t.Errorf("Konfiguration nennt %q", nein)
		}
	}
	for _, ja := range []string{"max-bytes: 16065", "section-pattern:", "cell-max-chars: 200"} {
		if !strings.Contains(k, ja) {
			t.Errorf("Konfiguration nennt %q nicht", ja)
		}
	}
	for _, block := range []string{"scan:", "structure:", "file:"} {
		if strings.Contains(gdEchtKonfig(strings.Replace(gdBeispiel, "\n"+block, "\nx"+block, 1)), "\n"+block) {
			t.Errorf("Block %s fehlt in der Datei, steht aber in der Konfiguration", block)
		}
	}
}

// Die Prüfung der Menge ist eine Schleife über Pfad mal Code: fehlt eine Zeile, ein Code oder der
// Glob, oder steht eine zu viel da, meldet sie es.
func TestGdBefundeFaengtJedesElement(t *testing.T) {
	codes := []string{gdLinien, gdBytes}
	var zeilen []string
	for _, r := range gdMenge() {
		for _, c := range codes {
			zeilen = append(zeilen, gdBefund(r.pfad, r.glob, c, "Datei hat 9 Bytes, erlaubt sind 1"))
		}
	}
	pruefe := gdBefunde(gdMenge(), codes, 26)
	if r := pruefe(Lauf{Out: strings.Join(zeilen, "")}); r != "" {
		t.Fatalf("vollständige Ausgabe ist rot: %s", r)
	}
	for i := range zeilen {
		ohne := append(append([]string{}, zeilen[:i]...), zeilen[i+1:]...)
		if pruefe(Lauf{Out: strings.Join(ohne, "")}) == "" {
			t.Errorf("Ausgabe ohne Zeile %d ist grün", i)
		}
		falsch := append([]string{}, zeilen...)
		falsch[i] = strings.Replace(falsch[i], "\t", "\tx", 2)
		if pruefe(Lauf{Out: strings.Join(falsch, "")}) == "" {
			t.Errorf("Ausgabe mit falschem Glob oder Code in Zeile %d ist grün", i)
		}
	}
	if pruefe(Lauf{Out: strings.Join(zeilen, "") + "x\n"}) == "" {
		t.Error("Ausgabe mit einer Zeile zu viel ist grün")
	}
}

// Die 15 Fälle gegen das echte d-check, soweit es im PATH liegt (Stufe gegenproben): sie sind grün.
// Die Fälle 166 bis 170 lesen die Konfiguration aus gdBeispiel.
func TestGuideFaelleLaufenGegenDCheck(t *testing.T) {
	if _, err := exec.LookPath("d-check"); err != nil {
		t.Skip("d-check nicht im PATH")
	}
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(gdBeispiel), 0o644); err != nil {
		t.Fatal(err)
	}
	erg, err := fahre(standard(), guideFaellePfad(pfad), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(erg) != 15 {
		t.Fatalf("%d Ergebnisse", len(erg))
	}
	for _, e := range erg {
		if e.Rot != "" {
			t.Errorf("%s ist rot: %s", e.Name, e.Rot)
		}
	}
}
