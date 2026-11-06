package main

import (
	"strings"
	"testing"
)

// Die 15 Fälle 174 bis 188 tragen die Namen aus SPEC-102, jeder genau einmal und in dieser Folge.
func TestKatalogFaelleHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{"g1-sauber", "g2-code-ohne-spezifikation", "g3-code-ohne-zeile", "g4-handbuch-nennt-unbekannt",
		"g5-handbuch-nennt-ausnahme", "g6-ausnahme-ohne-grund", "g7-ausnahme-ohne-code", "g8-code-doppelt", "g9-form",
		"g10-abschnitt-fehlt", "g11-fehler-go-fehlt", "g12-ohne-argument", "g13-k-ohne-c", "g14-code-nur-im-zaun",
		"g15-ausnahme-mit-zeile"}
	exits := []int{0, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 0, 1, 1}
	fs := katalogFaelle()
	if len(want) != 15 || len(fs) != 15 || len(exits) != 15 {
		t.Fatalf("%d erwartete Namen und %d Fälle, erwartet 15", len(want), len(fs))
	}
	imGanzen := map[string]int{}
	for _, f := range faelle() {
		imGanzen[f.Name]++
	}
	for i, n := range want {
		f := fs[i]
		if f.Name != "katalog/"+n {
			t.Errorf("Fall %d heißt %s statt katalog/%s", 174+i, f.Name, n)
		}
		if imGanzen[f.Name] != 1 {
			t.Errorf("Fall %d (%s): %d-mal in faelle()", 174+i, f.Name, imGanzen[f.Name])
		}
		if f.Gate != gateKatalog || !f.Getrennt || !f.Wegwerfmodul || len(f.WantText) > 0 || f.WantExit != exits[i] {
			t.Errorf("Fall %d: Gate %d, Getrennt %v, Wegwerfmodul %v, WantText %v, Exit %d", 174+i, f.Gate, f.Getrennt, f.Wegwerfmodul, f.WantText, f.WantExit)
		}
		if len(f.WantOut)+len(f.WantErr) == 0 {
			t.Errorf("Fall %d (%s) erwartet keinen Text", 174+i, f.Name)
		}
		for datei, inhalt := range f.Neu {
			if strings.Contains(inhalt, "\\u") || strings.Contains(datei, "\\u") {
				t.Errorf("%s: %s trägt eine Escape-Folge als Text", f.Name, datei)
			}
		}
	}
}

// G12 ruft das Programm ohne Argument, jeder andere Fall mit dem Wegwerfverzeichnis als einzigem Argument; G11 hat
// keine fehler.go, die übrigen Fälle tragen die vier Dateien.
func TestKatalogFaelleArgumenteUndDateien(t *testing.T) {
	for i, f := range katalogFaelle() {
		wantArgs := "{dir}"
		if i == 11 {
			wantArgs = ""
		}
		if strings.Join(f.Argumente, "|") != wantArgs {
			t.Errorf("%s: Argumente %q", f.Name, f.Argumente)
		}
		n := 4
		if i == 10 {
			n = 3
			if _, da := f.Neu[kaFehler]; da {
				t.Errorf("%s: fehler.go liegt im Baum", f.Name)
			}
		}
		if len(f.Neu) != n {
			t.Errorf("%s: %d Dateien statt %d", f.Name, len(f.Neu), n)
		}
	}
}

// Die Programmliste nennt katalog, und die Fälle rufen das Programm im Verzeichnis des Falls.
func TestStandardNenntKatalog(t *testing.T) {
	u := standard()
	gefunden := false
	for _, p := range u.Programme {
		gefunden = gefunden || p == "katalog"
	}
	if !gefunden {
		t.Error("katalog fehlt in standard().Programme")
	}
	name, args, wd := u.Befehl("/k", Fall{Gate: gateKatalog, Argumente: []string{"{dir}"}})
	if name != "katalog" || strings.Join(args, "|") != "/k" || wd != "/k" {
		t.Errorf("Programm: %q %v %q", name, args, wd)
	}
}
