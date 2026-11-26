package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// kopf liest die Header einer Anfrage als Abbildung und meldet doppelte Namen.
func kopf(t *testing.T, a Anfrage) map[string]string {
	t.Helper()
	m := map[string]string{}
	for _, h := range a.Header {
		if _, doppelt := m[h[0]]; doppelt {
			t.Fatalf("Header %s doppelt in %s", h[0], a.Ziel())
		}
		m[h[0]] = h[1]
	}
	return m
}

// erwarteteMenge bildet die Menge der Anfragen des Aufbaus aus dem Profil (Orakel: `$metadata`, je Set die Seiten, die ersten Lookups der
// Permutation) und meldet die Zahl der Elemente je Ziel.
func erwarteteMenge(p Profil, startwert uint64) map[string]int {
	m := map[string]int{"/odata/$metadata": 1}
	for set := 1; set <= p.Sets; set++ {
		m[fmt.Sprintf("/odata/Set%d", set)]++
		for k := p.SeitenGroesse; k < p.Entities; k += p.SeitenGroesse {
			m[fmt.Sprintf("/odata/Set%d?$skiptoken=%d", set, k)]++
		}
	}
	for _, s := range paare(p, startwert)[:p.Lookups] {
		m[fmt.Sprintf("/odata/Set%d(%d)", s[0], s[1])]++
	}
	return m
}

// TestListeReplay: jede Anfrage des Aufbaus kommt je Durchgang einmal vor, im ersten Durchgang genau einmal und im zweiten die ersten 3200
// der Folge noch einmal, mit den vier Headern (Accept wie in der Aufnahme, Workspace, Recording-Session, Replay-Session); die Replay-Session
// wechselt alle 100 Anfragen, in einer Session steht keine Anfrage zweimal, und die Sessions zweier Läufe haben keine gemeinsame Kennung.
func TestListeReplay(t *testing.T) {
	wache(t, 60*time.Second)
	p := abnahme()
	menge := erwarteteMenge(p, 1)
	if len(menge) != 10000 {
		t.Fatalf("Menge hat %d Elemente, erwartet 10000", len(menge))
	}
	liste := listeReplay(p, 1, 1, "ws-1", "rec-1")
	if len(liste) != 13200 {
		t.Fatalf("Liste hat %d Anfragen, erwartet 13200", len(liste))
	}
	zaehler := map[string]int{}
	sessions := map[string]map[string]bool{}
	for i, a := range liste {
		h := kopf(t, a)
		session := fmt.Sprintf("lasttest-replay-1-%d", i/100)
		if len(h) != 4 || h["Accept"] != "application/json" || h["X-Recorder-Workspace"] != "ws-1" ||
			h["X-Recorder-Recording-Session"] != "rec-1" || h["X-Recorder-Replay-Session"] != session {
			t.Fatalf("Anfrage %d: Header %v, erwartet vier mit Session %s", i, h, session)
		}
		z := a.Ziel()
		if _, ok := menge[z]; !ok {
			t.Fatalf("Anfrage %d: %s gehört nicht zum Aufbau", i, z)
		}
		zaehler[z]++
		if i == 9999 {
			for k := range menge {
				if zaehler[k] != 1 {
					t.Fatalf("im ersten Durchgang kommt %s %d-mal vor", k, zaehler[k])
				}
			}
		}
		if sessions[session] == nil {
			sessions[session] = map[string]bool{}
		}
		if sessions[session][z] {
			t.Fatalf("Anfrage %d: %s steht zweimal in %s", i, z, session)
		}
		sessions[session][z] = true
		if i >= 10000 && z != liste[i-10000].Ziel() {
			t.Fatalf("Anfrage %d: %s, erwartet %s des ersten Durchgangs", i, z, liste[i-10000].Ziel())
		}
	}
	zweimal := 0
	for _, n := range zaehler {
		if n == 2 {
			zweimal++
		}
	}
	if zweimal != 3200 || len(zaehler) != 10000 || len(sessions) != 132 {
		t.Errorf("zweimal %d, verschiedene Anfragen %d, Sessions %d, erwartet 3200, 10000 und 132", zweimal, len(zaehler), len(sessions))
	}
	for i, erwartet := range map[int]string{99: "0", 100: "1", 101: "1", 9999: "99", 10000: "100", 13199: "131"} {
		if got := kopf(t, liste[i])["X-Recorder-Replay-Session"]; got != "lasttest-replay-1-"+erwartet {
			t.Errorf("Anfrage %d: Session %s, erwartet lasttest-replay-1-%s", i, got, erwartet)
		}
	}
	for _, a := range listeReplay(p, 1, 2, "ws-1", "rec-1") {
		if _, ok := sessions[kopf(t, a)["X-Recorder-Replay-Session"]]; ok {
			t.Fatalf("die Replay-Session %s trägt Lauf 1 und Lauf 2", kopf(t, a)["X-Recorder-Replay-Session"])
		}
	}
	r := listeReplay(rauch(), 1, 1, "ws", "rec")
	if len(r) != 260 || kopf(t, r[259])["X-Recorder-Replay-Session"] != "lasttest-replay-1-2" {
		t.Errorf("rauch: %d Anfragen, letzte Session %s, erwartet 260 und lasttest-replay-1-2", len(r), kopf(t, r[259])["X-Recorder-Replay-Session"])
	}
}

var (
	lookupRe = regexp.MustCompile(`^/odata/Set([1-5])\((\d+)\)$`)
	mengeRe  = regexp.MustCompile(`^/odata/Set([1-5])$`)
	filterRe = regexp.MustCompile(`^\$filter=Gruppe%20eq%20(\d+)&\$top=10$`)
	topRe    = regexp.MustCompile(`^\$top=(\d+)$`)
)

// TestListeMock: die Art der Anfrage i ist i modulo 10 (70 Prozent Lookup, 20 Prozent Filter, 10 Prozent Top), je Art alle fünf Sets, Id von 1 bis
// 2000, Gruppe von 0 bis 19, Größe 10 oder 50, Query prozent-kodiert, Header Workspace und Accept, kein `$metadata`; jede Zahl der
// Arten stimmt, im Profil abnahme erreichen die Anfragen jede Grenze der Wertebereiche und beide Größen (Orakel: Literale für den Startwert 1),
// im Profil rauch (52 Filter) bleiben sie in den Bereichen.
func TestListeMock(t *testing.T) {
	wache(t, 60*time.Second)
	for _, f := range []struct {
		p                                     Profil
		anfragen, lookups, filter, top, maxID int
		grenzen                               bool
	}{{abnahme(), 13200, 9240, 2640, 1320, 2000, true}, {rauch(), 260, 182, 52, 26, 60, false}} {
		liste := listeMock(f.p, 1, "ws-1")
		if len(liste) != f.anfragen {
			t.Fatalf("Profil %s: %d Anfragen, erwartet %d", f.p.Name, len(liste), f.anfragen)
		}
		var arten [3]int
		var zellen [3][5]int
		minID, maxID, minG, maxG := 1<<30, 0, 1<<30, -1
		groessen := map[int]int{}
		for i, a := range liste {
			h := kopf(t, a)
			if len(h) != 2 || h["X-Recorder-Workspace"] != "ws-1" || h["Accept"] != "application/json" || strings.Contains(a.Ziel(), "$metadata") {
				t.Fatalf("Anfrage %d: Header %v, Ziel %s", i, h, a.Ziel())
			}
			art, set := 0, 0
			switch m := lookupRe.FindStringSubmatch(a.Pfad); {
			case m != nil && a.Query == "":
				id, _ := strconv.Atoi(m[2])
				set, _ = strconv.Atoi(m[1])
				minID, maxID = min(minID, id), max(maxID, id)
			case mengeRe.MatchString(a.Pfad) && filterRe.MatchString(a.Query):
				art = 1
				g, _ := strconv.Atoi(filterRe.FindStringSubmatch(a.Query)[1])
				minG, maxG = min(minG, g), max(maxG, g)
			case mengeRe.MatchString(a.Pfad) && topRe.MatchString(a.Query):
				art = 2
				n, _ := strconv.Atoi(topRe.FindStringSubmatch(a.Query)[1])
				groessen[n]++
			default:
				t.Fatalf("Anfrage %d: %s gehört zu keiner Art", i, a.Ziel())
			}
			if art > 0 {
				set, _ = strconv.Atoi(mengeRe.FindStringSubmatch(a.Pfad)[1])
			}
			erwartet := 0
			if i%10 == 9 {
				erwartet = 2
			} else if i%10 >= 7 {
				erwartet = 1
			}
			if art != erwartet {
				t.Fatalf("Anfrage %d (modulo 10 gleich %d): Art %d, erwartet %d", i, i%10, art, erwartet)
			}
			arten[art]++
			zellen[art][set-1]++
		}
		innerhalb := minID >= 1 && maxID <= f.maxID && minG >= 0 && maxG <= 19 && groessen[10]+groessen[50] == f.top
		erreicht := minID == 1 && maxID == f.maxID && minG == 0 && maxG == 19 && groessen[10] > 0 && groessen[50] > 0
		if arten != [3]int{f.lookups, f.filter, f.top} || !innerhalb || (f.grenzen && !erreicht) {
			t.Errorf("Profil %s: Arten %v, Id %d bis %d, Gruppe %d bis %d, Größen %v, erwartet %d, %d, %d, Id 1 bis %d, Gruppe 0 bis 19, 10 und 50 (Grenzen erreicht: %v)",
				f.p.Name, arten, minID, maxID, minG, maxG, groessen, f.lookups, f.filter, f.top, f.maxID, f.grenzen)
		}
		for art, je := range zellen {
			for set, n := range je {
				if n < 1 {
					t.Errorf("Profil %s: Art %d, Set%d ohne Anfrage", f.p.Name, art, set+1)
				}
			}
		}
	}
	a, b := listeMock(abnahme(), 1, "ws"), listeMock(abnahme(), 2, "ws")
	gleich := 0
	for i := range a {
		if a[i].Ziel() == b[i].Ziel() {
			gleich++
		}
	}
	if gleich == len(a) {
		t.Error("die Startwerte 1 und 2 ergeben dieselbe Liste")
	}
}
