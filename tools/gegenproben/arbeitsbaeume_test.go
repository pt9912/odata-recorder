package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gdBeispielWt liefert gdBeispiel mit scan.roots und der Ausnahme `.wt/**` in scan.ignore, wie die echte
// .d-check.yml sie führt (ohne roots liest das Modul links keine Datei).
func gdBeispielWt() string {
	return strings.Replace(gdBeispiel, `ignore: [".harness/**"]`, "roots: [\".\"]\n  ignore: [\".harness/**\", \".wt/**\"]", 1)
}

// Die vier Fälle 205 bis 208 zu a-check und die fünf Fälle 209 bis 213 zu docs-check tragen die Namen aus
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
	}
	alle := faelle()
	if len(want) != 9 || len(alle) < 9 {
		t.Fatalf("%d erwartete Namen, %d Fälle", len(want), len(alle))
	}
	ende := alle[len(alle)-9:]
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

// Die Fälle 209 bis 213 führen die Blöcke der echten Konfiguration: scan und links in 209 bis 212 (kein
// Block structure oder file), und Fall 213 legt je Datei der Menge von Fall 167 eine große Datei unter
// .wt/w/ an, AGENTS.md dort mit dem Abschnitt ### 3.1.
func TestArbeitsbaumFaelleBauenDieDateienAusDerEchtenKonfiguration(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(gdBeispielWt()), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := arbeitsbaumFaellePfad(pfad)
	if len(fs) != 5 {
		t.Fatalf("%d Fälle statt 5", len(fs))
	}
	for i, f := range fs {
		dir := t.TempDir()
		if err := f.Vorbereitung(dir); err != nil {
			t.Fatal(err)
		}
		k, err := os.ReadFile(filepath.Join(dir, ".d-check.yml"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(k), `ignore: [".harness/**", ".wt/**"]`) {
			t.Errorf("Fall %d: scan.ignore nennt .wt/** nicht: %q", 209+i, k)
		}
		if i < 4 && (strings.Contains(string(k), "file:") || strings.Contains(string(k), "structure:") || !strings.HasPrefix(string(k), "modules: [links]\nscan:\n")) {
			t.Errorf("Fall %d: Konfiguration %q", 209+i, k)
		}
	}
	dir := t.TempDir()
	if err := fs[4].Vorbereitung(dir); err != nil {
		t.Fatal(err)
	}
	for _, r := range gdMenge() {
		b, err := os.ReadFile(filepath.Join(dir, ".wt/w", r.pfad))
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(b), "\n"); n < gdGross {
			t.Errorf(".wt/w/%s hat %d Zeilen", r.pfad, n)
		}
		if k, _ := os.ReadFile(filepath.Join(dir, r.pfad)); len(k) > 100 {
			t.Errorf("%s im Hauptbaum hat %d Bytes", r.pfad, len(k))
		}
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".wt/w/AGENTS.md")); !strings.HasPrefix(string(b), "### 3.1 ") {
		t.Error(".wt/w/AGENTS.md trägt den Abschnitt ### 3.1 nicht")
	}
}

// gdBloecke nimmt nur die genannten Blöcke und setzt den Kopf davor.
func TestGdBloeckeNimmtNurDieGenanntenBloecke(t *testing.T) {
	k := gdBloecke(gdBeispielWt(), "modules: [links]\n", "scan")
	if k != "modules: [links]\nscan:\n  roots: [\".\"]\n  ignore: [\".harness/**\", \".wt/**\"]\n" {
		t.Errorf("Konfiguration: %q", k)
	}
}

// Die fünf Fälle 209 bis 213 gegen das echte d-check, soweit es im PATH liegt (Stufe gegenproben): sie
// sind grün. Die Konfiguration kommt aus gdBeispielWt().
func TestArbeitsbaumFaelleLaufenGegenDCheck(t *testing.T) {
	if _, err := exec.LookPath("d-check"); err != nil {
		t.Skip("d-check nicht im PATH")
	}
	pfad := filepath.Join(t.TempDir(), "d-check.yml")
	if err := os.WriteFile(pfad, []byte(gdBeispielWt()), 0o644); err != nil {
		t.Fatal(err)
	}
	erg, err := fahre(standard(), arbeitsbaumFaellePfad(pfad), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(erg) != 5 {
		t.Fatalf("%d Ergebnisse", len(erg))
	}
	for _, e := range erg {
		if e.Rot != "" {
			t.Errorf("%s ist rot: %s", e.Name, e.Rot)
		}
	}
}
