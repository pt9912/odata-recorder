package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

// hashWriter ist ein ResponseWriter, der Status, Header und Body in einen Hash schreibt und den Body auf Wunsch behält.
type hashWriter struct {
	h       hash.Hash
	header  http.Header
	body    bytes.Buffer
	behalte bool
}

func (w *hashWriter) Header() http.Header { return w.header }

func (w *hashWriter) WriteHeader(status int) { fmt.Fprintf(w.h, "Status %d\n", status) }

func (w *hashWriter) Write(b []byte) (int, error) {
	if w.behalte {
		w.body.Write(b)
	}
	return w.h.Write(b)
}

// gesamthash bildet einen Hash über alle Antworten des Referenzdienstes (`$metadata`, jede Seite über ihren Link, jede Entity einzeln) und
// über beide Listen der Anfragen (replay und mock) des Profils und des Startwerts.
func gesamthash(t *testing.T, p Profil, startwert uint64) string {
	t.Helper()
	r := referenz{p: p, startwert: startwert, basis: testBasis}
	h := sha256.New()
	hole := func(ziel string, behalte bool) []byte {
		w := &hashWriter{h: h, header: http.Header{}, behalte: behalte}
		fmt.Fprintf(h, "GET %s\n", ziel)
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, testBasis+ziel, nil))
		fmt.Fprintf(h, "\n%v\n", w.header)
		return w.body.Bytes()
	}
	hole("/odata/$metadata", false)
	for set := 1; set <= p.Sets; set++ {
		for ziel := fmt.Sprintf("/odata/Set%d", set); ziel != ""; {
			var s struct {
				Weiter string `json:"@odata.nextLink"`
			}
			if err := json.Unmarshal(hole(ziel, true), &s); err != nil {
				t.Fatal(err)
			}
			ziel = ""
			if s.Weiter != "" {
				ziel = s.Weiter[len(testBasis):]
			}
		}
		for id := 1; id <= p.Entities; id++ {
			hole(fmt.Sprintf("/odata/Set%d(%d)", set, id), false)
		}
	}
	for _, l := range [][]Anfrage{listeReplay(p, startwert, 1, "ws", "rec"), listeMock(p, startwert, "ws")} {
		for _, a := range l {
			fmt.Fprintf(h, "%s %v\n", a.Ziel(), a.Header)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// goldenAbnahme ist der Hash von gesamthash am Profil abnahme und Startwert 1 (200 Seiten, 10000 Entities, beide Listen), einmal gebildet und
// hier als Literal festgehalten: jede Änderung an Antworten oder Listen ändert ihn, eine absichtliche Änderung setzt ihn neu.
const goldenAbnahme = "bd7ba70845248a136782f0b2a0a425f9222f0e110d137bfea27040ec92d53bd5"

// TestDeterminismus: gleicher Startwert ergibt byte-gleiche Antworten des Referenzdienstes und gleiche Listen (SPEC-216 Nr. 6 c): am Profil
// abnahme ist der Hash für die Startwerte 0, 1 und 10 hoch 17 zweimal gebildet gleich, der für den Startwert 1 gleicht dem Literal
// goldenAbnahme, das ein anderer Prozess bildete, und die Startwerte 0, 1, 2 und 10 hoch 17 ergeben paarweise verschiedene Hashes. Kein Wert
// hängt von Zeit, Prozess oder Reihenfolge einer Map ab: jeder Hash entsteht in einem neuen Aufruf.
func TestDeterminismus(t *testing.T) {
	wache(t, 400*time.Second)
	hashes := map[uint64]string{2: gesamthash(t, abnahme(), 2)}
	for _, startwert := range []uint64{0, 1, 100000000000000000} {
		a, b := gesamthash(t, abnahme(), startwert), gesamthash(t, abnahme(), startwert)
		if a != b {
			t.Errorf("Startwert %d: zwei Hashes %s und %s", startwert, a, b)
		}
		hashes[startwert] = a
	}
	if hashes[1] != goldenAbnahme {
		t.Errorf("Startwert 1: Hash %s, erwartet %s", hashes[1], goldenAbnahme)
	}
	werte := make([]string, 0, len(hashes))
	for _, h := range hashes {
		werte = append(werte, h)
	}
	sort.Strings(werte)
	for i := 1; i < len(werte); i++ {
		if werte[i] == werte[i-1] {
			t.Errorf("zwei der Startwerte 0, 1, 2 und 10 hoch 17 ergeben denselben Hash %s", werte[i])
		}
	}
	if len(werte) != 4 {
		t.Errorf("%d Hashes, erwartet 4", len(werte))
	}
}

// startwerte sind die Startwerte der Komponenten-Prüfungen: der Rand 0, zwei kleine und ein achtzehnstelliger.
func startwerte() []uint64 { return []uint64{0, 1, 2, 100000000000000000} }

// wiederholungen ist die Zahl der Ausgaben je Komponente und Startwert, die einander gleichen müssen: die Reihenfolge einer Map in einer
// Komponente weicht bei einem Aufruf nur mit Wahrscheinlichkeit ab, bei vielen Aufrufen fast sicher.
const wiederholungen = 10

// TestStartwertKomponenten: jede Komponente der Fixture, die der Startwert bestimmt (SPEC-216 Nr. 6 c und 7: die Zeichen der Nutzlast, die
// Permutation der Lookups, die Permutation der Liste replay, die Parameter der Liste mock), liefert für gleichen Startwert dieselbe Ausgabe
// und für die vier Startwerte paarweise verschiedene. Die Permutation von replay ist die Folge der Anfragen als Index in der Menge des Aufbaus
// (ein Profil, dessen Lookups alle Paare nennen: die Menge ist dann bis auf die Folge unabhängig vom Startwert).
func TestStartwertKomponenten(t *testing.T) {
	wache(t, 60*time.Second)
	alle := Profil{Sets: 2, Entities: 20, SeitenGroesse: 10, Lookups: 40, Exchanges: 45, Rate: 20, Warmup: 3 * time.Second, Messphase: 10 * time.Second}
	komponenten := []struct {
		name string
		aus  func(sw uint64) string
	}{
		{"Nutzlast", func(sw uint64) string {
			_, _, body := bediene(referenz{p: abnahme(), startwert: sw, basis: testBasis}, http.MethodGet, "/odata/Set3(9)")
			return string(body)
		}},
		{"Permutation der Lookups", func(sw uint64) string { return fmt.Sprint(paare(abnahme(), sw)[:100]) }},
		{"Permutation von replay", func(sw uint64) string {
			menge, index := aufbauAnfragen(alle, sw), map[string]int{}
			for i, a := range menge {
				index[a.Ziel()] = i
			}
			var folge []int
			for _, a := range listeReplay(alle, sw, 1, "ws", "rec")[:len(menge)] {
				folge = append(folge, index[a.Ziel()])
			}
			return fmt.Sprint(folge)
		}},
		{"Parameter von mock", func(sw uint64) string {
			var ziele []string
			for _, a := range listeMock(rauch(), sw, "ws") {
				ziele = append(ziele, a.Ziel())
			}
			return strings.Join(ziele, " ")
		}},
	}
	for _, k := range komponenten {
		von := map[string]uint64{}
		for _, sw := range startwerte() {
			a := k.aus(sw)
			for i := 1; i < wiederholungen; i++ {
				if b := k.aus(sw); a != b {
					t.Errorf("%s, Startwert %d: Ausgabe %d weicht von der ersten ab", k.name, sw, i+1)
				}
			}
			if frueher, ok := von[a]; ok {
				t.Errorf("%s: die Startwerte %d und %d ergeben dieselbe Ausgabe", k.name, frueher, sw)
			}
			von[a] = sw
		}
	}
	if len(komponenten) != 4 || len(startwerte()) != 4 {
		t.Errorf("%d Komponenten, %d Startwerte, erwartet 4 und 4", len(komponenten), len(startwerte()))
	}
}

// TestPaare: die Permutation der Lookups trägt jedes Paar aus Set und Id genau einmal (je Profil), und die ersten fünf Paare für den Startwert 1
// sind festgelegt (Literal); der Startwert 2 ergibt andere.
func TestPaare(t *testing.T) {
	wache(t, 30*time.Second)
	for _, p := range []Profil{abnahme(), rauch(), kleinesProfil()} {
		gesehen := map[[2]int]int{}
		for _, s := range paare(p, 1) {
			if s[0] < 1 || s[0] > p.Sets || s[1] < 1 || s[1] > p.Entities {
				t.Fatalf("Profil %s: Paar %v außerhalb", p.Name, s)
			}
			gesehen[s]++
		}
		if len(gesehen) != p.Sets*p.Entities || len(paare(p, 1)) != p.Sets*p.Entities {
			t.Errorf("Profil %s: %d verschiedene Paare, erwartet %d", p.Name, len(gesehen), p.Sets*p.Entities)
		}
		for s, n := range gesehen {
			if n != 1 {
				t.Errorf("Profil %s: Paar %v %d-mal", p.Name, s, n)
			}
		}
	}
	if got := fmt.Sprint(paare(abnahme(), 1)[:5]); got != "[[5 314] [3 1327] [2 1987] [2 1243] [5 1611]]" {
		t.Errorf("erste Paare für Startwert 1: %s", got)
	}
	if got := fmt.Sprint(paare(abnahme(), 2)[:5]); got == fmt.Sprint(paare(abnahme(), 1)[:5]) {
		t.Errorf("Startwert 2 ergibt dieselben ersten Paare %s", got)
	}
}
