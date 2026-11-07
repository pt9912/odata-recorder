package inspection_test

import (
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// bestandLimit ist das Limit der Seiten in TestAnhaengenZwischenSeiten (der Wert L der Menge M-Bestand).
const bestandLimit = 3

// TestAnhaengenZwischenSeiten (X6, Menge M-Bestand: k × Zeitpunkt × n): ein Exchange, der zwischen
// zwei Seiten angehängt wird, verschiebt nichts. Die Folge der Seiten und ein anschließendes Lesen ab
// dem letzten gelesenen Eintrag liefern jede sequence_no von 1 bis n + k genau einmal in Reihenfolge,
// Gesamt wächst mit dem Anhängen, und die Folge selbst erreicht die angehängten genau dann, wenn nach
// der ersten Seite angehängt wird und die Folge dann noch läuft (n über dem Limit).
func TestAnhaengenZwischenSeiten(t *testing.T) {
	l := bestandLimit
	ks, ns := []int{0, 1, l, l + 1}, []int{0, 1, l, l + 1, 2*l + 1}
	const nachErster, nachLetzter = 0, 1
	faelle := 0
	for _, k := range ks {
		for _, zeit := range []int{nachErster, nachLetzter} {
			for _, n := range ns {
				faelle++
				f := neuesFake()
				f.legeAn(sid, n)
				d := neuerDienst(t, f)
				gelesen, nach, bestand := []int{}, 0, n
				for seite := 0; ; seite++ {
					z := lies(t, d, sid, inbound.Seite{Limit: l, Nach: nach})
					gelesen = append(gelesen, nummern(z)...)
					if z.Gesamt != bestand {
						t.Errorf("k %d Zeit %d n %d: Gesamt %d, erwartet %d", k, zeit, n, z.Gesamt, bestand)
					}
					if seite == 0 && zeit == nachErster {
						f.haengeAn(sid, k)
						bestand = n + k
					}
					if z.Weiter == 0 {
						break
					}
					nach = z.Weiter
				}
				imLauf := len(gelesen)
				if zeit == nachLetzter {
					f.haengeAn(sid, k)
					bestand = n + k
				}
				letzte := 0
				if len(gelesen) > 0 {
					letzte = gelesen[len(gelesen)-1]
				}
				for _, z := range folge(t, d, sid, l, letzte) {
					gelesen = append(gelesen, nummern(z)...)
					if z.Gesamt != bestand {
						t.Errorf("k %d Zeit %d n %d: Gesamt %d bei der Nachlese, erwartet %d", k, zeit, n, z.Gesamt, bestand)
					}
				}
				erreicht := zeit == nachErster && n > l
				if nachgelesen := len(gelesen) - imLauf; (erreicht && nachgelesen != 0) || (!erreicht && nachgelesen != k) {
					t.Errorf("k %d Zeit %d n %d: %d Einträge nachgelesen, die Folge erreichte die angehängten: %v", k, zeit, n, nachgelesen, erreicht)
				}
				if want := bereich(1, n+k); !slices.Equal(gelesen, want) {
					t.Errorf("k %d Zeit %d n %d: gelesen %v, erwartet %v", k, zeit, n, gelesen, want)
				}
			}
		}
	}
	if want := len(ks) * 2 * len(ns); faelle != want {
		t.Errorf("%d Fälle gelaufen, die Achsen verlangen %d", faelle, want)
	}
}
