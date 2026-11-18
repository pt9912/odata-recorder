package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// gdBeispielWt liefert gdBeispiel mit scan.roots und der Ausnahme `.wt/**` in scan.ignore, wie die echte
// .d-check.yml sie führt (ohne roots liest das Modul links keine Datei).
func gdBeispielWt() string {
	return strings.Replace(gdBeispiel, `ignore: [".harness/**"]`, "roots: [\".\"]\n  ignore: [\".harness/**\", \".wt/**\"]", 1)
}

// abBeispiel ist gdBeispielWt() mit den zehn Modulen der echten Konfiguration und den Blöcken codepaths und
// matrix in kleiner Form (die Blöcke scan, ids, structure und file trägt gdBeispiel).
func abBeispiel() string {
	k := strings.Replace(gdBeispielWt(), "modules: [links, structure, targets, file]",
		"modules: [links, anchors, ids, matrix, codepaths, spans, planning, structure, targets, file]", 1)
	return k + "codepaths:\n  roots: [spec, docs, harness]\nmatrix:\n  classes:\n" +
		"    - {name: adr, paths: [\"docs/plan/adr/[0-9]*.md\"]}\n    - {name: aussen, paths: [\"**\"]}\n" +
		"  status: {forbidden: [superseded, deprecated]}\n"
}

// Die vier Fälle 205 bis 208 zu a-check und die acht Fälle 209 bis 216 zu docs-check tragen die Namen aus
// SPEC-102, jeder genau einmal, am Ende von faelle() und in dieser Folge; die Fälle zu a-check lesen die
// Ströme gemeinsam, die zu docs-check getrennt.
func TestArbeitsbaumFaelleHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []struct {
		name string
		gate Gate
		getr bool
	}{
		{"a-check/arbeitsbaum-ungemeldet", gateACheck, false},
		{"a-check/arbeitsbaum-verletzung-im-hauptbaum", gateACheck, false},
		{"a-check/wt-im-quellbaum-bleibt-geprueft", gateACheck, false},
		{"a-check/arbeitsbaum-im-arbeitsbaum", gateACheck, false},
		{"docs-check/arbeitsbaum-ungemeldet", gateDCheck, true},
		{"docs-check/arbeitsbaum-verweis-im-hauptbaum", gateDCheck, true},
		{"docs-check/wt-im-quellbaum-bleibt-geprueft", gateDCheck, true},
		{"docs-check/arbeitsbaum-im-arbeitsbaum", gateDCheck, true},
		{"docs-check/arbeitsbaum-file-und-structure", gateDCheck, true},
		{"docs-check/arbeitsbaum-module-ungemeldet", gateDCheck, true},
		{"docs-check/module-im-hauptbaum-gemeldet", gateDCheck, true},
		{"docs-check/zellgrenze-im-hauptbaum", gateDCheck, true},
	}
	alle := faelle()
	if len(want) != 12 || len(alle) < 12 {
		t.Fatalf("%d erwartete Namen, %d Fälle", len(want), len(alle))
	}
	ende := alle[len(alle)-12:]
	imGanzen := map[string]int{}
	for _, f := range alle {
		imGanzen[f.Name]++
	}
	for i, w := range want {
		f := ende[i]
		if f.Name != w.name {
			t.Errorf("Fall %d heißt %s statt %s", 205+i, f.Name, w.name)
		}
		if imGanzen[f.Name] != 1 {
			t.Errorf("Fall %d (%s): %d-mal in faelle()", 205+i, f.Name, imGanzen[f.Name])
		}
		if f.Gate != w.gate || f.Getrennt != w.getr {
			t.Errorf("Fall %d: Gate %d, Getrennt %v", 205+i, f.Gate, f.Getrennt)
		}
	}
}

// Die Fälle 205, 206 und 208 tragen dieselbe Verletzung unter drei Pfaden; 205 und 208 melden nichts
// (Exit 0, kein wrong-direction), 206 meldet sie, 207 meldet einen Dienst im Modell.
func TestACheckArbeitsbaumFaelleTragenDieselbeVerletzung(t *testing.T) {
	fs := aCheckArbeitsbaeumeFaelle()
	if len(fs) != 4 {
		t.Fatalf("%d Fälle statt 4", len(fs))
	}
	var inhalte []string
	for _, i := range []int{0, 1, 3} {
		var pfade []string
		for p, inhalt := range fs[i].Neu {
			if strings.HasSuffix(p, "/x.go") {
				pfade = append(pfade, p)
				inhalte = append(inhalte, inhalt)
			}
		}
		if len(pfade) != 1 {
			t.Errorf("%s: %d Dateien x.go", fs[i].Name, len(pfade))
		}
	}
	for _, in := range inhalte {
		if in != inhalte[0] || !strings.Contains(in, "src/hexagon/model/kern") {
			t.Errorf("Verletzung weicht ab: %q", in)
		}
	}
	for i, exit := range []int{0, 1, 1, 0} {
		if fs[i].WantExit != exit {
			t.Errorf("Fall %d: Exit %d statt %d", 205+i, fs[i].WantExit, exit)
		}
	}
	if len(fs[0].Verboten) != 1 || len(fs[3].Verboten) != 1 || len(fs[1].WantText) != 1 || len(fs[2].WantText) != 1 {
		t.Error("Erwartung an die Ausgabe fehlt")
	}
}

// abBauen führt die Vorbereitung des Falls aus und liefert das Wegwerfverzeichnis.
func abBauen(t *testing.T, f Fall) string {
	t.Helper()
	dir := t.TempDir()
	if err := f.Vorbereitung(dir); err != nil {
		t.Fatalf("%s: %v", f.Name, err)
	}
	return dir
}

func lies(t *testing.T, dir, p string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, p))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Die Fälle 209 bis 216 führen Blöcke der echten Konfiguration: scan und links in 209 bis 212, scan, ids,
// codepaths und matrix mit den sechs lesenden Modulen in 214 und 215, scan, structure und file in 213 und 216.
func TestArbeitsbaumFaelleBauenDieDateienAusDerEchtenKonfiguration(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(abBeispiel()), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := arbeitsbaumFaellePfad(pfad)
	if len(fs) != 8 {
		t.Fatalf("%d Fälle statt 8", len(fs))
	}
	for i, f := range fs {
		k := lies(t, abBauen(t, f), ".d-check.yml")
		if !strings.Contains(k, `ignore: [".harness/**", ".wt/**"]`) {
			t.Errorf("Fall %d: scan.ignore nennt .wt/** nicht: %q", 209+i, k)
		}
		switch {
		case i < 4:
			if !strings.HasPrefix(k, "modules: [links]\nscan:\n") || strings.Contains(k, "file:") || strings.Contains(k, "structure:") {
				t.Errorf("Fall %d: Konfiguration %q", 209+i, k)
			}
		case i == 5 || i == 6:
			if !strings.HasPrefix(k, "modules: [links, anchors, ids, matrix, codepaths, spans]\nscan:\n") ||
				!strings.Contains(k, "\nids:\n") || !strings.Contains(k, "\ncodepaths:\n") || !strings.Contains(k, "\nmatrix:\n") ||
				strings.Contains(k, "file:") || strings.Contains(k, "structure:") {
				t.Errorf("Fall %d: Konfiguration %q", 209+i, k)
			}
		default:
			if !strings.HasPrefix(k, "modules: [structure, file]\nscan:\n") || !strings.Contains(k, "\nstructure:\n") || !strings.Contains(k, "\nfile:\n") {
				t.Errorf("Fall %d: Konfiguration %q", 209+i, k)
			}
		}
	}
}

// Fall 213 legt je Paar Regel/Datei der echten Konfiguration (15 Paare, 14 Dateien) die verletzende Eingabe
// unter .wt/w/<Pfad> an, die Dateien des Hauptbaums sind klein.
func TestFall213LegtJedesPaarUnterWtAn(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(abBeispiel()), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := abBauen(t, arbeitsbaumFaellePfad(pfad)[4])
	paare, err := abPaare(abBeispiel())
	if err != nil {
		t.Fatal(err)
	}
	dateien := map[string]bool{}
	for _, p := range paare {
		dateien[p.pfad] = true
		e, err := p.regel.eingabe()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(lies(t, dir, ".wt/w/"+p.pfad), e) {
			t.Errorf(".wt/w/%s trägt die Eingabe der Regel %s %q nicht", p.pfad, p.regel.block, p.regel.glob)
		}
		if len(lies(t, dir, p.pfad)) > 200 {
			t.Errorf("%s im Hauptbaum ist groß", p.pfad)
		}
	}
	if len(paare) != 15 || len(dateien) != 14 {
		t.Errorf("%d Paare und %d Dateien statt 15 und 14", len(paare), len(dateien))
	}
	if b := lies(t, dir, ".wt/w/AGENTS.md"); !strings.Contains(b, "### 3.1 ") {
		t.Error(".wt/w/AGENTS.md trägt den Abschnitt ### 3.1 nicht")
	}
	if b := lies(t, dir, ".wt/w/harness/README.md"); !strings.Contains(b, "| "+strings.Repeat("a", 201)+" | "+strings.Repeat("a", 201)+" |") {
		t.Error(".wt/w/harness/README.md trägt die Tabelle mit Zellen von 201 Zeichen nicht")
	}
}

// Die Mengen von Fall 213: die Regeln beider Blöcke werden gelesen, und jede Regel ohne Pfad der Menge,
// in unbekannter Form oder mit einer Grenze über der Eingabe ist ein Fehler des Aufbaus.
func TestAbPaareFaengtDieMengenUndFormen(t *testing.T) {
	regeln, err := abRegeln(abBeispiel())
	if err != nil || len(regeln) != 7 {
		t.Fatalf("%d Regeln, Fehler %v", len(regeln), err)
	}
	for _, r := range regeln {
		if _, err := r.eingabe(); err != nil {
			t.Errorf("Regel %q: %v", r.glob, err)
		}
	}
	mutiert := map[string]string{
		"achte Regel in file ohne Pfad":    strings.Replace(abBeispiel(), "codepaths:\n", "  - files: \".claude/hooks/*.md\"\n    max-lines: 5\n    max-bytes: 50\ncodepaths:\n", 1),
		"Regel in structure ohne Pfad":     strings.Replace(abBeispiel(), "\n# file\nfile:", "\n  - files: \"docs/x.md\"\n    section-pattern: '^### 3\\.\\d+ '\n    max-lines: 29\n# file\nfile:", 1),
		"unbekannter Schlüssel":            strings.Replace(abBeispiel(), "    sections: each\n", "    sections: each\n    hint: x\n", 1),
		"unbekanntes section-pattern":      strings.Replace(abBeispiel(), `'^### 3\.\d+ '`, `'^## 9 '`, 1),
		"Grenze über der Eingabe (Zeilen)": strings.Replace(abBeispiel(), "max-lines: 53", "max-lines: 50000", 1),
		"Grenze über der Eingabe (Bytes)":  strings.Replace(abBeispiel(), "max-bytes: 3759", "max-bytes: 900000", 1),
		"Regel ohne Grenze":                strings.Replace(abBeispiel(), "    max-lines: 53\n    max-bytes: 3759\n", "", 1),
		"Tabelle ohne cell-max-chars":      strings.Replace(abBeispiel(), "          cell-max-chars: 200\n", "", 1),
		"Block structure fehlt":            strings.Replace(abBeispiel(), "\nstructure:", "\nxstructure:", 1),
		"Block file fehlt":                 strings.Replace(abBeispiel(), "\nfile:", "\nxfile:", 1),
		"Wert ohne Schlüssel":              strings.Replace(abBeispiel(), "    sections: each\n", "    sections each\n", 1),
		"sections nicht each":              strings.Replace(abBeispiel(), "sections: each", "sections: one", 1),
		"keine Zahl":                       strings.Replace(abBeispiel(), "max-lines: 29", "max-lines: viele", 1),
		"Schlüssel vor der ersten Regel":   strings.Replace(abBeispiel(), "structure:\n  - files: \"harness/README.md\"", "structure:\n  max-lines: 3\n  - files: \"harness/README.md\"", 1),
		"Glob ohne Treffer in der Menge":   strings.Replace(abBeispiel(), "\"harness/conventions.md\"", "\"harness/nichts.md\"", 1),
		"cell-max-chars vor der Spalte":    strings.Replace(abBeispiel(), "      column:\n        - name: \"Vertrag\"\n", "      column:\n          cell-max-chars: 3\n        - name: \"Vertrag\"\n", 1),
		"table mit Wert":                   strings.Replace(abBeispiel(), "    table:\n", "    table: x\n", 1),
	}
	for name, k := range mutiert {
		if k == abBeispiel() {
			t.Errorf("%s: Mutation ändert nichts", name)
			continue
		}
		fehler := false
		if paare, err := abPaare(k); err != nil {
			fehler = true
		} else {
			for _, p := range paare {
				if _, err := p.regel.eingabe(); err != nil {
					fehler = true
				}
			}
		}
		if !fehler {
			t.Errorf("%s: kein Fehler des Aufbaus", name)
		}
	}
}

// Eine Regel, deren Glob mit `**/` oder `*/*/` beginnt, bleibt im Aufbau dieselbe Regel: die Mutation
// einer Regel wird als Befund des Falls sichtbar, nicht als Fehler des Aufbaus.
func TestAbPaareLoestDenGlobVonDerWurzel(t *testing.T) {
	grund, err := abPaare(abBeispiel())
	if err != nil {
		t.Fatal(err)
	}
	regeln, _ := abRegeln(abBeispiel())
	for _, r := range regeln {
		for _, vor := range []string{"**/", "*/*/"} {
			k := strings.Replace(abBeispiel(), "files: \""+r.glob+"\"", "files: \""+vor+r.glob+"\"", -1)
			paare, err := abPaare(k)
			if err != nil || len(paare) != len(grund) {
				t.Errorf("%s%s: %d Paare statt %d, Fehler %v", vor, r.glob, len(paare), len(grund), err)
			}
		}
	}
}

// Fall 214 und 215 legen je Modul von abLesend seinen Defekt unter .wt/w/ bzw. x/ an, dazu die ADR-Datei
// an der Wurzel; die Menge der Module ist die Klasse der Zeile modules:.
func TestModuleFaelleLegenJedenDefektAn(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(abBeispiel()), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := arbeitsbaumFaellePfad(pfad)
	for i, prefix := range map[int]string{5: ".wt/w/", 6: "x/"} {
		dir := abBauen(t, fs[i])
		if !strings.Contains(lies(t, dir, abAdr), "**Status:** Superseded") {
			t.Errorf("Fall %d: ADR-Datei ohne Status Superseded", 209+i)
		}
		for _, m := range abLesend {
			x, ok := abDefektVon(m)
			if !ok {
				t.Fatalf("kein Defekt für %s", m)
			}
			if lies(t, dir, prefix+x.datei) != x.inhalt {
				t.Errorf("Fall %d: %s%s trägt den Defekt von %s nicht", 209+i, prefix, x.datei, m)
			}
			for p, inhalt := range x.mehr {
				if lies(t, dir, prefix+p) != inhalt {
					t.Errorf("Fall %d: %s%s fehlt", 209+i, prefix, p)
				}
			}
		}
	}
	if len(abDefekte()) != len(abLesend) {
		t.Errorf("%d Defekte für %d Module", len(abDefekte()), len(abLesend))
	}
	if fs[5].WantExit != 0 || fs[6].WantExit != 1 || fs[5].Nachher != nil || fs[6].Nachher == nil {
		t.Error("Erwartung der Fälle 214 und 215 weicht ab")
	}
}

// Die Klassifikation: jedes Modul der Zeile modules: steht in genau einer der Listen, und jedes Element
// der Listen steht in der Zeile; ein Modul außerhalb beider Listen, in einer Liste ohne Zeile oder eine
// Zeile ohne Liste ist ein Fehler.
func TestKlassifikationDerModule(t *testing.T) {
	if len(abLesend) != 6 || len(abNichtLesend) != 4 {
		t.Fatalf("%d lesende und %d nicht lesende Module statt 6 und 4", len(abLesend), len(abNichtLesend))
	}
	if err := abPruefeModule(abBeispiel()); err != nil {
		t.Fatalf("Beispiel: %v", err)
	}
	zeile := "modules: [links, anchors, ids, matrix, codepaths, spans, planning, structure, targets, file]"
	mutiert := map[string]string{
		"elftes Modul":      strings.Replace(abBeispiel(), zeile, strings.TrimSuffix(zeile, "]")+", diagrams]", 1),
		"Modul fehlt":       strings.Replace(abBeispiel(), "codepaths, ", "", 1),
		"Zeile fehlt":       strings.Replace(abBeispiel(), "modules:", "xmodules:", 1),
		"keine Liste":       strings.Replace(abBeispiel(), zeile, "modules: links", 1),
		"Liste führt mehr":  strings.Replace(abBeispiel(), "file]", "]", 1),
		"Wort fremd statt":  strings.Replace(abBeispiel(), "spans", "span", 1),
		"Modul ohne Liste":  strings.Replace(abBeispiel(), "planning", "reviews", 1),
		"nur lesende Liste": strings.Replace(abBeispiel(), ", planning, structure, targets, file]", "]", 1),
	}
	for name, k := range mutiert {
		if k == abBeispiel() {
			t.Errorf("%s: Mutation ändert nichts", name)
			continue
		}
		if err := abPruefeModule(k); err == nil {
			t.Errorf("%s: kein Fehler", name)
		}
	}
	for _, m := range abLesend {
		if slices.Contains(abNichtLesend, m) {
			t.Errorf("%s steht in beiden Listen", m)
		}
	}
}

// Fall 216 legt harness/README.md an der Wurzel mit Zellen von cell-max-chars + 1 Zeichen in jeder Spalte
// der ersten Tabellenregel an.
func TestFall216LegtDieTabelleImHauptbauAn(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(abBeispiel()), 0o644); err != nil {
		t.Fatal(err)
	}
	f := arbeitsbaumFaellePfad(pfad)[7]
	b := lies(t, abBauen(t, f), "harness/README.md")
	if !strings.HasPrefix(b, "## Sensors (Feedback-Gates)\n\n| Vertrag | Tut was |\n|---|---|\n| ") ||
		!strings.Contains(b, "| "+strings.Repeat("a", 201)+" | "+strings.Repeat("a", 201)+" |") {
		t.Errorf("harness/README.md: %q", b)
	}
	if f.WantExit != 1 || f.Nachher == nil {
		t.Error("Erwartung von Fall 216 weicht ab")
	}
	ohne := strings.Replace(abBeispiel(), "    section: \"## Sensors (Feedback-Gates)\"\n", "", 1)
	if err := os.WriteFile(pfad, []byte(ohne), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := arbeitsbaumFaellePfad(pfad)[7].Vorbereitung(t.TempDir()); err == nil {
		t.Error("ohne Tabellenregel kein Fehler des Aufbaus")
	}
}

// Die Prüfung von Fall 215 ist eine Schleife über die sechs Module: fehlt eine Zeile, ein Code oder die
// Datei, oder steht eine zu viel da, meldet sie es.
func TestBefundeJeModulFaengtJedesElement(t *testing.T) {
	var zeilen []string
	for _, m := range abLesend {
		x, _ := abDefektVon(m)
		zeilen = append(zeilen, "x/"+x.datei+":1\tziel\t"+x.code+"\tmeldung\n")
	}
	pruefe := abBefundeJeModul("x/")
	if r := pruefe(Lauf{Out: strings.Join(zeilen, "")}); r != "" {
		t.Fatalf("vollständige Ausgabe ist rot: %s", r)
	}
	for i := range zeilen {
		ohne := append(append([]string{}, zeilen[:i]...), zeilen[i+1:]...)
		if pruefe(Lauf{Out: strings.Join(ohne, "")}) == "" {
			t.Errorf("Ausgabe ohne Zeile %d ist grün", i)
		}
		falsch := append([]string{}, zeilen...)
		falsch[i] = strings.Replace(falsch[i], "\tziel\t", "\tziel\tx", 1)
		if pruefe(Lauf{Out: strings.Join(falsch, "")}) == "" {
			t.Errorf("Ausgabe mit falschem Code in Zeile %d ist grün", i)
		}
		falsch[i] = strings.Replace(zeilen[i], "x/", ".wt/w/", 1)
		if pruefe(Lauf{Out: strings.Join(falsch, "")}) == "" {
			t.Errorf("Ausgabe mit falschem Pfad in Zeile %d ist grün", i)
		}
	}
	if pruefe(Lauf{Out: strings.Join(zeilen, "") + "x/b.md:1\tz\tc\tm\n"}) == "" {
		t.Error("Ausgabe mit einer Zeile zu viel ist grün")
	}
}

// gdBloecke nimmt nur die genannten Blöcke und setzt den Kopf davor.
func TestGdBloeckeNimmtNurDieGenanntenBloecke(t *testing.T) {
	k := gdBloecke(gdBeispielWt(), "modules: [links]\n", "scan")
	if k != "modules: [links]\nscan:\n  roots: [\".\"]\n  ignore: [\".harness/**\", \".wt/**\"]\n" {
		t.Errorf("Konfiguration: %q", k)
	}
}

// Die acht Fälle 209 bis 216 gegen das echte d-check, soweit es im PATH liegt (Stufe gegenproben): sie
// sind grün. Die Konfiguration kommt aus abBeispiel().
func TestArbeitsbaumFaelleLaufenGegenDCheck(t *testing.T) {
	if _, err := exec.LookPath("d-check"); err != nil {
		t.Skip("d-check nicht im PATH")
	}
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(abBeispiel()), 0o644); err != nil {
		t.Fatal(err)
	}
	erg, err := fahre(standard(), arbeitsbaumFaellePfad(pfad), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(erg) != 8 {
		t.Fatalf("%d Ergebnisse", len(erg))
	}
	for _, e := range erg {
		if e.Rot != "" {
			t.Errorf("%s ist rot: %s", e.Name, e.Rot)
		}
	}
}
