package hybrid_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestMutierenderReplayFallback hält Z29: POST, PUT, PATCH, DELETE und MERGE mit ODR-E6002 und der Kategorie
// unsupported (ohne und mit Body): M genau einmal und danach nie, R einmal, die aufgezeichnete Erfolgsantwort geht mit
// Kennzeichnung aus, und der Zustands-Fake von M verzeichnet keinen Schreibzugriff.
func TestMutierenderReplayFallback(t *testing.T) {
	methoden := []string{"POST", "PUT", "PATCH", "DELETE", "MERGE"}
	for _, m := range methoden {
		for _, mitBodyBytes := range []bool{false, true} {
			g := neuesRig(t, mitKategorien(t, katUnsupported), standardMax)
			g.m.fn = func(_ int, _ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
				if a.Body != nil {
					_, _ = io.ReadAll(a.Body)
				}
				return inbound.Antwort{}, code(codeNamed(t, "ODR-E6002"))
			}
			want := antwortVon(201, []byte("aufgezeichnet"))
			g.replayAntwort(want)
			a := anfrage(m)
			soll := muster(5, 4)
			if mitBodyBytes {
				a.Body, a.Laenge = bytes.NewReader(bytes.Clone(soll)), -1
				g.r.lies = liesAlles
			}
			got, err := g.v.Hybrid(context.Background(), a)
			if err != nil || got.Status != 201 || got.Body != want.Body || len(kennzeichen(got)) != 1 {
				t.Errorf("%s (Body %v): Antwort %+v, Fehler %v", m, mitBodyBytes, got, err)
			}
			g.zaehlen(t, m, 1, 1)
			if mitBodyBytes && !bytes.Equal(g.r.erster(t).gelesen, soll) {
				t.Errorf("%s: R erhielt %v, erwartet %v", m, g.r.erster(t).gelesen, soll)
			}
		}
	}
}

// pfad ist ein Pfad der Wiederholung: die Antworten der Stufen hängen vom Zähler n ihrer Aufrufe ab, zeit sind die
// Aufrufe von M und R je Durchlauf.
type pfad struct {
	name string
	m    func(n int) (inbound.Antwort, error)
	r    func(n int) (inbound.Antwort, error)
	zeit [2]int
}

// TestWiederholung hält Z30: derselbe Aufruf zweimal ruft in jedem Pfad alle Stufen erneut und merkt kein Ergebnis:
// Erfolg des Mock, Replay-Treffer, Replay-Miss, Fehler ohne Fallback; Aufruf 2 liefert die Antwort des zweiten
// Aufrufs der Stufe.
func TestWiederholung(t *testing.T) {
	e6003 := code(codeNamed(t, "ODR-E6003"))
	e6005 := code(codeNamed(t, "ODR-E6005"))
	miss := code(codeNamed(t, "ODR-E5001"))
	zaehl := func(n int) inbound.Antwort {
		return inbound.Antwort{Status: 200 + n, Body: neuerStrom([]byte{byte('a' + n)})}
	}
	pfade := []pfad{
		{name: "Erfolg", zeit: [2]int{1, 0}, m: func(n int) (inbound.Antwort, error) { return zaehl(n), nil }},
		{name: "Replay-Treffer", zeit: [2]int{1, 1},
			m: func(int) (inbound.Antwort, error) { return inbound.Antwort{}, e6003 },
			r: func(n int) (inbound.Antwort, error) { return zaehl(n), nil }},
		{name: "Replay-Miss", zeit: [2]int{1, 1},
			m: func(int) (inbound.Antwort, error) { return inbound.Antwort{}, e6003 },
			r: func(int) (inbound.Antwort, error) { return inbound.Antwort{}, miss }},
		{name: "Fehler ohne Fallback", zeit: [2]int{1, 0},
			m: func(int) (inbound.Antwort, error) { return inbound.Antwort{}, e6005 }},
	}
	for _, p := range pfade {
		for _, methode := range []string{"GET", "POST"} {
			g := neuesRig(t, offenste(t), standardMax)
			g.m.fn = func(n int, _ context.Context, _ inbound.Anfrage) (inbound.Antwort, error) { return p.m(n) }
			if p.r != nil {
				g.r.fn = func(n int, _ context.Context, _ inbound.Anfrage) (inbound.Antwort, error) { return p.r(n) }
			}
			for durchlauf := range 2 {
				got, err := g.v.Hybrid(context.Background(), anfrage(methode))
				switch p.name {
				case "Erfolg", "Replay-Treffer":
					if err != nil || got.Status != 200+durchlauf {
						t.Errorf("%s %s Aufruf %d: Status %d, Fehler %v", p.name, methode, durchlauf+1, got.Status, err)
					}
				case "Replay-Miss":
					if err != error(miss) {
						t.Errorf("%s %s Aufruf %d: Fehler %v", p.name, methode, durchlauf+1, err)
					}
				default:
					if err != error(e6005) {
						t.Errorf("%s %s Aufruf %d: Fehler %v", p.name, methode, durchlauf+1, err)
					}
				}
				g.zaehlen(t, p.name+" "+methode, p.zeit[0]*(durchlauf+1), p.zeit[1]*(durchlauf+1))
			}
		}
	}
}
