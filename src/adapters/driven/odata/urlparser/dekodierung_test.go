package urlparser_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// dekodStelle ist eine Stelle, an der eine ganze Einheit der URL dekodiert wird: bau setzt die Einheit ein, ort ist der erwartete Ort
// eines Verstoßes.
type dekodStelle struct {
	name, ort string
	bau       func(einheit string) (pfad, query string)
}

// dekodStellen sind die vier Stellen: das erste und ein weiteres Segment des Pfads, der Name und der Wert einer Option (drei
// Aufrufe der Dekodierung im Code). Die Einheit ist das ganze Segment, der ganze Name, der ganze Wert (SPEC-161 Nr. 2, SPEC-162 Nr. 2).
func dekodStellen() []dekodStelle {
	return []dekodStelle{
		{"Segment 1", "pfad/1", func(e string) (string, string) { return "/" + e, "" }},
		{"Segment 2", "pfad/2", func(e string) (string, string) { return "/S/" + e, "" }},
		{"Name einer Option", "query/1", func(e string) (string, string) { return "/S", e + "=1" }},
		{"Wert einer Option", "query/1", func(e string) (string, string) { return "/S", "x=" + e }},
	}
}

// pruefeKodierungsverstoss erwartet für jede Einheit an jeder Stelle `kodierung` und liefert die Zahl der Elemente.
func pruefeKodierungsverstoss(t *testing.T, einheiten []string) int {
	t.Helper()
	n := 0
	for _, s := range dekodStellen() {
		for _, e := range einheiten {
			pfad, query := s.bau(e)
			pruefeFall(t, parserNeu(), fall{pfad: pfad, query: query, u: model.URLKodierung, ort: s.ort})
			n++
		}
	}
	return n
}

// TestNulEinheit (Zeile 26, SPEC-161 Nr. 2, SPEC-162 Nr. 2): U+0000 ist `kodierung`, auch als erstes Zeichen einer dekodierten Einheit.
// Menge 16: 4 Stellen (Segment 1, Segment 2, Name und Wert einer Option) × 4 Lagen von `%00` in der Einheit (allein, am Anfang,
// in der Mitte, am Ende).
func TestNulEinheit(t *testing.T) {
	if n := pruefeKodierungsverstoss(t, []string{"%00", "%00x", "a%00b", "a%00"}); n != 16 {
		t.Fatalf("%d Elemente, erwartet 16", n)
	}
}

// hexFormen bildet die Menge der ungültigen Prozent-Folgen aus dem Wortlaut „ein `%` ohne zwei Hex-Ziffern“ (SPEC-161 Nr. 2):
// (a) Nicht-Hex an der ersten Stelle, eine Hex-Ziffer an der zweiten (10 × 7), (b) eine Hex-Ziffer an der ersten, Nicht-Hex an der
// zweiten (4 × 10), (c) Nicht-Hex an beiden (10), je allein und gefolgt von gültigen Fortsetzungsbytes (`%90%80%80`, `%80%80%80`):
// der Wert des ersten Escapes wäre sonst ein Vorspannbyte, das mit den Fortsetzungsbytes gültiges UTF-8 ergibt; dazu (d) vier Formen
// mit `%` am Ende (`%`, `%1`, `x%`, `x%1`). 364 Einheiten.
func hexFormen() []string {
	nichtHex := []string{"G", "g", "Z", "z", ":", "@", "`", "-", "+", "%"}
	var formen []string
	for _, rest := range []string{"", "%90%80%80", "%80%80%80"} {
		for _, c1 := range nichtHex {
			for _, c2 := range []string{"0", "1", "2", "3", "4", "A", "a"} {
				formen = append(formen, "%"+c1+c2+rest)
			}
			formen = append(formen, "%"+c1+c1+rest)
		}
		for _, c1 := range []string{"0", "1", "A", "a"} {
			for _, c2 := range nichtHex {
				formen = append(formen, "%"+c1+c2+rest)
			}
		}
	}
	return append(formen, "%", "%1", "x%", "x%1")
}

// TestHexUngueltig (Zeile 24, SPEC-161 Nr. 2): jede Prozent-Folge mit einer Nicht-Hex-Ziffer ist `kodierung`, auch wenn der Wert
// des Escapes mit den folgenden Bytes gültiges UTF-8 ergäbe. Menge 1456: 4 Stellen × 364 Einheiten.
func TestHexUngueltig(t *testing.T) {
	if len(hexFormen()) != 364 {
		t.Fatalf("%d Einheiten, erwartet 364", len(hexFormen()))
	}
	if n := pruefeKodierungsverstoss(t, hexFormen()); n != 1456 {
		t.Fatalf("%d Elemente, erwartet 1456", n)
	}
}
