package urlparser_test

import (
	"strings"
	"testing"
)

// textZuschlag ist der Zuschlag der Schranke der Textsumme (SPEC-160 Nr. 6 b): 3 Bytes je Schlüsselteil, höchstens 16 Teile.
const textZuschlag = 48

// TestWaechterTextsumme (Zeile 16, SPEC-160 Nr. 6 b): die Summe der Längen aller Texte des Ergebnisses ist höchstens
// `len(pfad) + len(rawQuery) + 48`; nur der Zeitpunkt ohne Sekunden wird kanonisch länger, um genau 3 Bytes. Menge 53:
// (a) 33 Schlüssel: 1 bis 16 benannte Teile `A=2024-01-05T10:20Z,B=…` mit Namen von 1 und von 100 Zeichen (32) und ein unbenannter
// Teil (1), mit der Textsumme `1 + k · (m + 20)` (bzw. gleich der Länge des Pfads) von Hand; (b) 20 Literalformen (Zeichenkette 5,
// Wahrheitswert 2, Zahl 5, Guid 1, Datum 1, Zeitpunkt 6), deren kanonischer Text um genau den Wert der Spalte vom Text der
// Schreibweise abweicht.
func TestWaechterTextsumme(t *testing.T) {
	n := 0
	for k := 1; k <= 16; k++ {
		for _, m := range []int{1, 100} {
			var teile []string
			for i := range k {
				teile = append(teile, wiederhole(string(rune('A'+i)), m)+"=2024-01-05T10:20Z")
			}
			pfad := "/S(" + strings.Join(teile, ",") + ")"
			u := gelesen(t, pfad, "")
			if got, soll := textsumme(u), 1+k*(m+20); got != soll {
				t.Errorf("k=%d m=%d: Textsumme %d, erwartet %d", k, m, got, soll)
			}
			if got := textsumme(u); got > len(pfad)+textZuschlag {
				t.Errorf("k=%d m=%d: Textsumme %d über %d + %d", k, m, got, len(pfad), textZuschlag)
			}
			n++
		}
	}
	u := gelesen(t, "/S(2024-01-05T10:20Z)", "")
	if got := textsumme(u); got != len("/S(2024-01-05T10:20Z)") {
		t.Errorf("unbenannter Teil: Textsumme %d, erwartet 21", got)
	}
	n++
	if n != 33 {
		t.Fatalf("%d Schlüssel, erwartet 33", n)
	}

	formen := []struct {
		form string
		diff int // len(Text) − len(Form), von Hand
	}{
		{"''", 0}, {"'a'", 0}, {"'O''Hara'", 0}, {"'%41'", -2}, {"'%27%27'", -4},
		{"TRUE", 0}, {"fAlSe", 0},
		{"007", -2}, {"1.50", -1}, {"-0", -1}, {"0.0", -2}, {"-0.50", -1},
		{"0F8FAD5B-D9CB-469F-A165-70867728950E", 0},
		{"2024-01-05", 0},
		{"2024-01-05T10:20Z", 3}, {"2024-01-05T10:20:00Z", 0}, {"2024-01-05T10:20:00.0Z", -2}, {"2024-01-05T10:20:30.500Z", -2},
		{"2024-01-05T10:20+01:00", -2}, {"2024-01-05T10:20:00-23:59", -5},
	}
	if len(formen) != 20 {
		t.Fatalf("%d Literalformen, erwartet 20", len(formen))
	}
	for _, f := range formen {
		u := gelesen(t, "/S(Id="+f.form+")", "")
		if got := len(u.Ressource.Schluessel[0].Literal.Text) - len(f.form); got != f.diff {
			t.Errorf("%q: Text %q, Differenz %d, erwartet %d", f.form, u.Ressource.Schluessel[0].Literal.Text, got, f.diff)
		}
	}
}
