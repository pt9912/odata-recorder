package main

import (
	"os/exec"
	"strings"
	"testing"
)

// Die zehn Fälle H1 bis H10 (SPEC-102, Fälle 121 bis 130) tragen die Namen aus dem Vertrag, jeder genau einmal.
func TestFaelleDesHandbuchsHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{
		"h1-sauber", "h2-planung", "h3-chronik", "h4-kennung", "h5-version-weicht-ab", "h6-binary-weicht-ab",
		"h7-block-weicht-ab", "h8-make-ziel", "h9-binary-fehlt", "h10-zitat",
	}
	have := map[string]int{}
	fs := handbuchFaelle()
	for _, f := range fs {
		have[f.Name]++
	}
	if len(fs) != 10 {
		t.Errorf("%d Fälle, erwartet 10", len(fs))
	}
	imGanzen := map[string]bool{}
	for _, f := range faelle() {
		imGanzen[f.Name] = true
	}
	for _, n := range want {
		if have["handbuch/"+n] != 1 || !imGanzen["handbuch/"+n] {
			t.Errorf("Fall handbuch/%s: %d-mal im Satz, in faelle(): %v", n, have["handbuch/"+n], imGanzen["handbuch/"+n])
		}
	}
}

// Jeder Fall legt ein Wegwerfmodul an, ruft pruefung handbuch mit getrennten Strömen und schreibt
// keine Escape-Folge als Text (AGENTS.md §3.18).
func TestFaelleDesHandbuchsSindWegwerfmodulMitGetrenntenStroemen(t *testing.T) {
	for _, f := range handbuchFaelle() {
		if !f.Wegwerfmodul || !f.Getrennt || f.Gate != gatePruefung || len(f.Argumente) != 3 || f.Argumente[0] != "handbuch" {
			t.Errorf("%s: Wegwerfmodul=%v Getrennt=%v Argumente=%v", f.Name, f.Wegwerfmodul, f.Getrennt, f.Argumente)
		}
		if len(f.WantText) > 0 {
			t.Errorf("%s nennt einen Text ohne Strom", f.Name)
		}
		for datei, inhalt := range f.Neu {
			if strings.Contains(inhalt, "\\u") {
				t.Errorf("%s: %s trägt eine Escape-Folge als Text", f.Name, datei)
			}
		}
	}
}

// Die zehn Fälle gegen das echte pruefung, soweit es im PATH liegt (Stufe gegenproben): sie sind grün.
func TestFaelleDesHandbuchsLaufenGegenPruefung(t *testing.T) {
	if _, err := exec.LookPath("pruefung"); err != nil {
		t.Skip("pruefung nicht im PATH")
	}
	erg, err := fahre(standard(), handbuchFaelle(), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(erg) != 10 {
		t.Fatalf("%d Ergebnisse", len(erg))
	}
	for _, e := range erg {
		if e.Rot != "" {
			t.Errorf("%s ist rot: %s", e.Name, e.Rot)
		}
	}
}
