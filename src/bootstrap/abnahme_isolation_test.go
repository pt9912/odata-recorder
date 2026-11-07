package bootstrap_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// Isolation (SPEC-049 Nr. 6, LH-QA-03): 20 Workspaces in einer Datei, jeder mit der Folge S(k); jeder Schritt gleich dem Einzellauf von k.

const abIsoN = 20

// abBeob ist, was ein Schritt von S(k) beobachten lässt: die normalisierte Antwort oder das Ergebnis des Aufrufs, danach Revision und Cursor-Liste.
type abBeob struct {
	Schritt  int
	Text     string
	Revision int64
	Cursors  string
}

func abWsName(k int) string { return fmt.Sprintf("ws-%02d", k) }

// abIsoOrakelUmgebung ist die Umgebung des Einzellaufs von k: eine frische Datei mit dem Workspace k allein.
func abIsoOrakelUmgebung(t *testing.T, k int) *abUmgebung {
	t.Helper()
	u := abOeffne(t, t.TempDir()+"/r.db", t.TempDir())
	u.ausgangsdaten()
	u.neuerWorkspace(abWsName(k))
	u.cursor(abWsName(k), int64(k+1))
	return u
}

// abIsoUmgebung ist die Umgebung des Paralleltests: eine Datei mit den 20 Workspaces, je mit dem Cursor auf Position k+1.
func abIsoUmgebung(t *testing.T) *abUmgebung {
	t.Helper()
	u := abOeffne(t, t.TempDir()+"/r.db", t.TempDir())
	u.ausgangsdaten()
	for k := 0; k < abIsoN; k++ {
		u.neuerWorkspace(abWsName(k))
		u.cursor(abWsName(k), int64(k+1))
	}
	return u
}

// abFolgeS führt die Folge S(k) für den Workspace k aus (SPEC-049 Nr. 6 b) und liefert die Beobachtungen; cursorLesen zählt die gelesenen
// Cursor-Listen. Der Aufruf ruft kein Fatal (er läuft auf Goroutinen); ein Fehler steht als Text der Beobachtung.
func abFolgeS(ctx context.Context, u *abUmgebung, k int, cursorLesen *atomic.Int32) []abBeob {
	ws := abWsName(k)
	var beob []abBeob
	beo := func(text string) {
		b := abBeob{Schritt: len(beob) + 1, Text: text}
		if w, err := u.db.Workspaces().Get(ctx, ws); err != nil {
			b.Text += " | Revision nicht lesbar: " + err.Error()
		} else {
			b.Revision = w.Revision
		}
		cs, err := u.db.Cursors().List(ctx, ws)
		cursorLesen.Add(1)
		if err != nil {
			b.Text += " | Cursors nicht lesbar: " + err.Error()
		}
		b.Cursors = fmt.Sprintf("%+v", cs)
		beob = append(beob, b)
	}
	http := func(a abAnfrage) { beo(abNormDet(abSende(ctx, u.mock, ws, a), nil)) }
	ks := strconv.Itoa(k)
	http(abAnfrage{Methode: "GET", Pfad: "/Products"})
	http(abAnfrage{Methode: "POST", Pfad: "/Products", Body: fmt.Sprintf(`{"Id":%d,"Name":"W%s","Active":%t}`, 100+k, ks, k%2 == 0)})
	http(abAnfrage{Methode: "PATCH", Pfad: "/Products(1)", Body: `{"Name":"N` + ks + `"}`})
	http(abAnfrage{Methode: "DELETE", Pfad: fmt.Sprintf("/Products(%d)", 2+k%3)})
	http(abAnfrage{Methode: "GET", Pfad: "/Products", Query: "$select=Id,Name&$orderby=Id"})
	sn, err := u.snap.Create(ctx, ws, "s1")
	beo(fmt.Sprintf("Snapshot %s|%s|%s|%s|%v", sn.Name, sn.WorkspaceID, sn.ServiceID, sn.StateHash, err))
	http(abAnfrage{Methode: "PUT", Pfad: "/Products(5)", Body: `{"Id":5,"Name":"P` + ks + `"}`})
	http(abAnfrage{Methode: "GET", Pfad: "/Products/$count"})
	w, err := u.ws.Restore(ctx, ws, sn.ID)
	beo(fmt.Sprintf("Restore Revision %d|%v", w.Revision, err))
	http(abAnfrage{Methode: "GET", Pfad: "/Products(5)"})
	if k%5 == 0 {
		w, err = u.ws.Reset(ctx, ws)
		beo(fmt.Sprintf("Reset Revision %d|%v", w.Revision, err))
		http(abAnfrage{Methode: "GET", Pfad: "/Products"})
	}
	l, err := u.snap.List(ctx, ws)
	text := fmt.Sprintf("Liste %d|%v", len(l), err)
	for _, s := range l {
		text += fmt.Sprintf("|%s %s %v", s.Name, s.WorkspaceID, s.ID == sn.ID)
	}
	beo(text)
	return beob
}

var (
	abIsoOnce   sync.Once
	abIsoOrakel [][]abBeob
)

// abIsoOrakelFolgen liefert die Einzelläufe der 20 Workspaces: das Orakel von k (SPEC-049 Nr. 6 c).
func abIsoOrakelFolgen(t *testing.T) [][]abBeob {
	t.Helper()
	abIsoOnce.Do(func() {
		var z atomic.Int32
		for k := 0; k < abIsoN; k++ {
			abIsoOrakel = append(abIsoOrakel, abFolgeS(context.Background(), abIsoOrakelUmgebung(t, k), k, &z))
		}
	})
	return abIsoOrakel
}

// TestAbnahmeIsolation (T54 bis T60): 20 Goroutinen auf ein Startsignal, jeder Schritt von S(k) gleich dem Orakel von k, Revision und Cursor-Liste je
// Schritt, Snapshot s1 in allen 20 Workspaces, Restore und Reset berühren nur den eigenen Workspace, am Ende trägt kein Eintrag einen fremden Wert.
func TestAbnahmeIsolation(t *testing.T) {
	orakel := abIsoOrakelFolgen(t)
	u := abIsoUmgebung(t)
	ergebnis := make([][]abBeob, abIsoN)
	var cursorLesen, frueh atomic.Int32
	p := &abParallel{N: abIsoN}
	p.lauf(func(k int) {
		if p.Bereit.Load() != abIsoN || !p.Gefallen.Load() {
			frueh.Add(1)
		}
		ergebnis[k] = abFolgeS(context.Background(), u, k, &cursorLesen)
	})
	if p.Gestartet.Load() != abIsoN || p.Beendet.Load() != abIsoN || frueh.Load() != 0 {
		t.Errorf("Goroutinen gestartet %d, beendet %d, vor dem Startsignal begonnen %d; erwartet 20, 20, 0", p.Gestartet.Load(), p.Beendet.Load(), frueh.Load())
	}
	verglichen, erwartet := 0, 20*11+4*2
	for k := 0; k < abIsoN; k++ {
		if len(ergebnis[k]) != len(orakel[k]) {
			t.Errorf("ws-%02d: %d Beobachtungen, Orakel %d", k, len(ergebnis[k]), len(orakel[k]))
			continue
		}
		for i := range ergebnis[k] {
			verglichen++
			if a, o := ergebnis[k][i], orakel[k][i]; a != o {
				t.Errorf("ws-%02d Schritt %d weicht vom Orakel ab:\n ist  %+v\n soll %+v", k, a.Schritt, a, o)
			}
		}
	}
	if verglichen != erwartet {
		t.Errorf("%d Schritte verglichen, erwartet %d", verglichen, erwartet)
	}
	if int(cursorLesen.Load()) != erwartet {
		t.Errorf("%d Cursor-Listen gelesen, erwartet %d", cursorLesen.Load(), erwartet)
	}
	// Fremder Zustand am Ende: kein Eintrag von ws-k trägt einen Wert aus der Folge von ws-j (20·19 Paare).
	paare := 0
	for k := 0; k < abIsoN; k++ {
		es, err := u.db.Entities().List(context.Background(), abWsName(k))
		if err != nil {
			t.Fatal(err)
		}
		for j := 0; j < abIsoN; j++ {
			if j == k {
				continue
			}
			paare++
			js := strconv.Itoa(j)
			for _, e := range es {
				for _, fremd := range []string{`"W` + js + `"`, `"N` + js + `"`, `"P` + js + `"`} {
					if strings.Contains(e.Payload, fremd) {
						t.Errorf("ws-%02d trägt den Wert %s von ws-%02d: %s/%s %s", k, fremd, j, e.EntitySet, e.EntityKey, e.Payload)
					}
				}
				if e.EntitySet == "Products" && e.EntityKey == strconv.Itoa(100+j) {
					t.Errorf("ws-%02d trägt den Schlüssel %d von ws-%02d", k, 100+j, j)
				}
			}
		}
	}
	if paare != abIsoN*(abIsoN-1) {
		t.Errorf("%d Paare geprüft, erwartet %d", paare, abIsoN*(abIsoN-1))
	}
}

// TestAbnahmeIsolationOrakelUnterscheidet (T61): die Orakel der 20 Workspaces unterscheiden sich paarweise; die Antwort von Schritt 5 (eigene
// Entity, eigener Name, eigener Tombstone) und die des Schrittes 2 tragen k.
func TestAbnahmeIsolationOrakelUnterscheidet(t *testing.T) {
	orakel := abIsoOrakelFolgen(t)
	if len(orakel) != abIsoN {
		t.Fatalf("%d Orakel, erwartet %d", len(orakel), abIsoN)
	}
	for _, schritt := range []int{2, 5} {
		paare := 0
		for a := 0; a < abIsoN; a++ {
			for b := a + 1; b < abIsoN; b++ {
				paare++
				if orakel[a][schritt-1].Text == orakel[b][schritt-1].Text {
					t.Errorf("Schritt %d: das Orakel von ws-%02d gleicht dem von ws-%02d", schritt, a, b)
				}
			}
		}
		if paare != abIsoN*(abIsoN-1)/2 {
			t.Errorf("Schritt %d: %d Paare, erwartet %d", schritt, paare, abIsoN*(abIsoN-1)/2)
		}
	}
	for k := 0; k < abIsoN; k++ {
		if n := len(orakel[k]); n != 11 && n != 13 || (k%5 == 0) != (len(orakel[k]) == 13) {
			t.Errorf("ws-%02d: %d Beobachtungen", k, n)
		}
	}
}

// TestAbnahmeIsolationStart (T62): alle 20 Goroutinen sind am Startsignal, bevor eine den ersten Schritt beginnt.
func TestAbnahmeIsolationStart(t *testing.T) {
	var frueh atomic.Int32
	p := &abParallel{N: abIsoN}
	p.lauf(func(int) {
		if p.Bereit.Load() != abIsoN || !p.Gefallen.Load() {
			frueh.Add(1)
		}
	})
	if frueh.Load() != 0 || p.Gestartet.Load() != abIsoN || p.Beendet.Load() != abIsoN {
		t.Errorf("vor dem Signal begonnen %d, gestartet %d, beendet %d", frueh.Load(), p.Gestartet.Load(), p.Beendet.Load())
	}
}
