package main

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// TestKatNennungDavor: jedes Vorkommen von `ODR-E` zählt, gleich welches Zeichen davor steht (Nr. 3 d). Menge: 11
// Zeichen davor (Buchstabe, drei Buchstaben, Ziffer, Bindestrich, Unterstrich, Leerzeichen, Tabulator, Punkt,
// Klammer, Anführungszeichen) und der Zeilenanfang × 5 Orte (Absatz, Inline-Code, zweite Zelle, Überschrift, Zaun)
// × 2 Klassen (Code nicht in C, Code in den Ausnahmen) = 120.
func TestKatNennungDavor(t *testing.T) {
	codes := codeListe()[:4]
	davor := []string{"", "x", "foo", "7", "-", "_", " ", "\t", ".", "(", "\""}
	zaehle(t, "Zeichen davor", len(davor), 11)
	n := 0
	klassen := []struct {
		code    string
		ausnahm []string
		text    string
	}{
		{"ODR-E9999", nil, "Handbuch Zeile %d nennt ODR-E9999, der nicht in fehler.go steht"},
		{codes[0], []string{codes[0]}, "Handbuch Zeile %d nennt " + codes[0] + ", der in den Ausnahmen steht"},
	}
	for _, k := range klassen {
		for _, d := range davor {
			x := d + k.code
			orte := nennorte(x, codes[1])
			for _, ort := range []string{"Absatz", "Inline-Code", "zweite Zelle", "Überschrift", "Zaun"} {
				m := grund(codes, k.ausnahm)
				m[dHandbuch] = orte[ort](m[dHandbuch])
				// Der Anfang der Zeile ist der Fixture-Text selbst, nicht die Zeile mit dem Vorkommen: gesucht wird x.
				erwarteBefunde(t, fmt.Sprintf("%q davor, %s, %s", d, ort, k.code), m, fmt.Sprintf(k.text, zeileVon(t, m[dHandbuch], x)))
				n++
			}
		}
	}
	zaehle(t, "Fälle", n, 110)
}

// TestKatZellenRand: leere Zellen, Zellen aus Leerraum und aus zwei Backticks an jeder Stelle einer Zeile, mit und ohne
// Schlussstrich (Nr. 6 c). Menge: Zellenzahl 2 und 4 (6 Stellen) × 3 Arten × 2 Schlüsse, ohne die 4 Fälle einer leeren letzten Zelle ohne Schlussstrich = 32, dazu eine leere erste
// Zelle ohne Leerzeichen in beiden Tabellen.
func TestKatZellenRand(t *testing.T) {
	n := 0
	for _, zahl := range []int{2, 4} {
		for pos := 0; pos < zahl; pos++ {
			for _, art := range []struct{ roh, soll string }{{"", ""}, {" ", ""}, {"``", "``"}} {
				for _, schluss := range []string{"|", ""} {
					soll := make([]string, zahl)
					roh := make([]string, zahl)
					for i := range soll {
						soll[i], roh[i] = "a", "a"
					}
					soll[pos], roh[pos] = art.soll, art.roh
					zeile := "|" + strings.Join(roh, "|") + schluss
					if schluss == "" && pos == zahl-1 && art.soll == "" {
						continue // ohne Schlussstrich ist eine leere letzte Zelle ein Schlussstrich
					}
					if ist := zellen(zeile); !slices.Equal(ist, soll) {
						t.Errorf("zellen(%q) = %q, erwartet %q", zeile, ist, soll)
					}
					n++
				}
			}
		}
	}
	if n != 32 {
		t.Errorf("%d Elemente", n)
	}
	codes := codeListe()[:3]
	m := grund(codes, nil)
	kuerze(m, dHandbuch, kopfH, kopfH+"||x|\n")
	erwarteBefunde(t, "Handbuch ||x|", m, fmt.Sprintf("Zelle: Handbuch Zeile %d: erste Zelle ist nicht genau ein Name in Inline-Code", zeileVon(t, m[dHandbuch], "||x|")))
	m = grund(codes, nil)
	kuerze(m, dSpez, "\nText danach.\n", "||`ODR-E9001`|b|a|\n\nText danach.\n")
	erwarteOk(t, "§4 ||`…`|b|a|", m)
}
