package main

import (
	"strings"
	"testing"
)

func TestBewerte(t *testing.T) {
	verstoss := Fall{Name: "v", WantExit: 1, WantText: "regel"}
	erlaubt := Fall{Name: "e"}
	tests := []struct {
		name       string
		f          Fall
		code       int
		out        string
		abgelaufen bool
		rot        string // Teiltext des Grundes; leer: grün
	}{
		{"Verstoß gemeldet", verstoss, 1, "regel: x", false, ""},
		{"Verstoß nicht gemeldet", verstoss, 0, "", false, "Verstoß nicht gemeldet"},
		{"anderer Code", verstoss, 2, "regel", false, "erwartet Exit 1, war 2"},
		{"falsches Wort", verstoss, 1, "anderes", false, `nennt "regel" nicht`},
		{"Erlaubt grün", erlaubt, 0, "", false, ""},
		{"Erlaubt abgelehnt", erlaubt, 1, "boese", false, "erlaubter Fall abgelehnt"},
		{"Frist", verstoss, -1, "", true, "Zeit überschritten"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bewerte(tc.f, tc.code, tc.out, tc.abgelaufen)
			if tc.rot == "" && got != "" || !strings.Contains(got, tc.rot) {
				t.Fatalf("bewerte = %q, erwartet %q", got, tc.rot)
			}
		})
	}
}

func TestFaelleHabenEindeutigeNamen(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range faelle() {
		if seen[f.Name] {
			t.Fatalf("Name %s doppelt", f.Name)
		}
		seen[f.Name] = true
	}
}
