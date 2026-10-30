package main

import (
	"fmt"
	"strings"
	"testing"
)

// Tests zu W-1 bis W-5 der zweiten Verifikation und zu den Mengen des Architect (SPEC-200 Nr. 2 d, 3, 4, 5, 8, 9, 13):
// jede Zusage läuft als Schleife über die Menge, die ihr Wortlaut meint; das Orakel steht im Test, von Hand gesetzt.
// Wo eine Funktion direkt aufrufbar ist, ruft der Test sie direkt, ohne Prozess.

// Z67 (R3, Nr. 3 b, 3 c): ein `\r` gehört nicht zur Zeile, wenn ein `\n` unmittelbar folgt; am Dateiende ohne folgendes
// `\n` ist es ein Zeichen der letzten Zeile. Menge: sieben Formen des Architect (`x`, `x\n`, `x\r\n` ohne Befund; `x\r`,
// `x\r\r\n`, `x\r\r`, `x\ry\r\n` mit je einem Befund U+000D) und drei Formen am Zeilenende vor und nach einer weiteren
// Zeile (`a\n\r`, `\r`, `a\r\nb\r`): 10 Elemente. Orakel: die Zeilen von Hand.
func TestHandbuchCRAmDateiende(t *testing.T) {
	faelle := []struct {
		in      string
		zeilen  []string
		befunde []string
	}{
		{"x", []string{"x"}, nil},
		{"x\n", []string{"x"}, nil},
		{"x\r\n", []string{"x"}, nil},
		{"x\r", []string{"x\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"x\r\r\n", []string{"x\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"x\r\r", []string{"x\r\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"x\ry\r\n", []string{"x\ry"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"a\n\r", []string{"a", "\r"}, []string{"Zeile 2: Zeichen U+000D nicht zulässig"}},
		{"\r", []string{"\r"}, []string{"Zeile 1: Zeichen U+000D nicht zulässig"}},
		{"a\r\nb\r", []string{"a", "b\r"}, []string{"Zeile 2: Zeichen U+000D nicht zulässig"}},
	}
	if len(faelle) != 10 {
		t.Fatalf("%d Elemente, erwartet 10", len(faelle))
	}
	for _, f := range faelle {
		zeilen := zeilenVon(f.in)
		if fmt.Sprintf("%q", zeilen) != fmt.Sprintf("%q", f.zeilen) {
			t.Fatalf("%q: Zeilen %q, erwartet %q", f.in, zeilen, f.zeilen)
		}
		var ist []string
		for _, b := range zeichenBefunde(zeilen) {
			ist = append(ist, b.text)
		}
		if strings.Join(ist, "|") != strings.Join(f.befunde, "|") {
			t.Fatalf("%q: Befunde %q, erwartet %q", f.in, ist, f.befunde)
		}
	}
}
