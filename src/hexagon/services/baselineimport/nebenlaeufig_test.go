package baselineimport_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// schranke hält n Goroutinen an, bis die letzte sie erreicht, höchstens zehn Sekunden (Grenze).
type schranke struct {
	mu   sync.Mutex
	n    int
	ziel int
	tor  chan struct{}
}

func neueSchranke(ziel int) *schranke { return &schranke{ziel: ziel, tor: make(chan struct{})} }

func (s *schranke) warte() {
	s.mu.Lock()
	s.n++
	if s.n == s.ziel {
		close(s.tor)
	}
	s.mu.Unlock()
	select {
	case <-s.tor:
	case <-time.After(10 * time.Second):
	}
}

type lauf struct {
	r   inbound.ImportErgebnis
	err error
}

// parallel startet die Aufrufe hinter einer Startschranke und sammelt die Ergebnisse in Reihenfolge der Anfragen.
func parallel(w *welt, as []inbound.ImportAnfrage) []lauf {
	w.start()
	start := neueSchranke(len(as))
	out := make([]lauf, len(as))
	var wg sync.WaitGroup
	for i, a := range as {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.warte()
			out[i].r, out[i].err = w.imp.Import(context.Background(), a)
		}()
	}
	wg.Wait()
	return out
}

// TestZeileB5GleicheAnfrage: 16 gleichzeitige Aufrufe mit gleicher Anfrage: genau einer Neu, alle
// denselben ReportJSON, alle setzen die Metadaten.
func TestZeileB5GleicheAnfrage(t *testing.T) {
	const n = 16
	w := neu(t)
	w.sessionR()
	tor := neueSchranke(n)
	w.p.setzeHaken("Imports.Get", func(int) { tor.warte() })
	as := make([]inbound.ImportAnfrage, n)
	for i := range as {
		as[i] = anfrage()
	}
	out := parallel(w, as)
	neuZahl := 0
	for i, l := range out {
		if l.err != nil {
			t.Fatalf("Aufruf %d: %v", i, l.err)
		}
		if l.r.Neu {
			neuZahl++
		}
		if l.r.Import.ReportJSON != out[0].r.Import.ReportJSON {
			t.Errorf("Aufruf %d: anderer ReportJSON", i)
		}
	}
	if neuZahl != 1 {
		t.Errorf("Neu in %d von %d Aufrufen", neuZahl, n)
	}
	if got := w.p.anzahl("Metadaten.SetMetadata"); got != n {
		t.Errorf("SetMetadata %d-mal, erwartet %d", got, n)
	}
}

// TestZeileB6GleicheIDAndereAnfrage: dieselbe import_id, zwei Anfragen (2 mal 8 Aufrufe): genau eine
// Gruppe gewinnt, die andere erhält ODR-E6102.
func TestZeileB6GleicheIDAndereAnfrage(t *testing.T) {
	const n = 8
	w := neu(t)
	b := w.sessionR()
	ref := w.exchanges.m[b.ses][0].ResponseBlob
	tor := neueSchranke(2 * n)
	w.p.setzeHaken("Imports.Get", func(int) { tor.warte() })
	var as []inbound.ImportAnfrage
	for i := 0; i < 2*n; i++ {
		a := anfrage()
		if i%2 == 1 {
			a.MetadataHash = ref
		}
		as = append(as, a)
	}
	out := parallel(w, as)
	var fehlerGruppe [2]int
	var neuZahl int
	for i, l := range out {
		if l.err != nil {
			erwarteFehler(t, fmt.Sprint("Aufruf ", i), l.err, model.CodeE6102, 409)
			fehlerGruppe[i%2]++
		} else if l.r.Neu {
			neuZahl++
		}
	}
	if neuZahl != 1 || fehlerGruppe != [2]int{n, 0} && fehlerGruppe != [2]int{0, n} {
		t.Errorf("Neu %d, Fehler je Gruppe %v", neuZahl, fehlerGruppe)
	}
	gewinner := 0
	if fehlerGruppe[0] == n {
		gewinner = 1
	}
	for i, l := range out {
		if i%2 == gewinner && (l.err != nil || l.r.Import.ReportJSON != out[gewinner].r.Import.ReportJSON) {
			t.Errorf("Aufruf %d der Gewinnergruppe: %v", i, l.err)
		}
	}
}

func normiereBytes(s, importID, snapID string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, importID, "IMPORT"), snapID, "SNAPSHOT")
}

// TestZeileB7VerschiedeneIDs: acht import_ids je zweimal, dieselbe Session: alle gelingen, jedes
// Ergebnis gleicht dem eines Einzelaufrufs; der Dienst hält keinen Zustand außer der Konfig.
func TestZeileB7VerschiedeneIDs(t *testing.T) {
	// Einzelaufruf als Maß.
	e := neu(t)
	e.sessionR()
	soll := mustImport(t, e, anfrage())
	sollReport := normiereBytes(soll.Import.ReportJSON, "imp-1", soll.Import.BaselineSnapshotID)
	sollManifest := normiereBytes(string(e.blobs.writers[0].bytes()), "imp-1", soll.Import.BaselineSnapshotID)

	w := neu(t)
	w.sessionR()
	var as []inbound.ImportAnfrage
	for i := 0; i < 16; i++ {
		a := anfrage()
		a.ImportID = fmt.Sprintf("imp-%d", i%8+1)
		as = append(as, a)
	}
	out := parallel(w, as)
	neuZahl := 0
	for i, l := range out {
		if l.err != nil {
			t.Errorf("Aufruf %d: %v", i, l.err)
			continue
		}
		if l.r.Neu {
			neuZahl++
		}
		if got := normiereBytes(l.r.Import.ReportJSON, l.r.Import.ID, l.r.Import.BaselineSnapshotID); got != sollReport {
			t.Errorf("Aufruf %d: ReportJSON weicht vom Einzelaufruf ab", i)
		}
	}
	if neuZahl != 8 {
		t.Errorf("Neu in %d Aufrufen, erwartet 8", neuZahl)
	}
	for _, pub := range w.imports.veroeffentlichungen() {
		if len(pub.SourceSessionIDs) != 1 {
			t.Errorf("Quellen %v", pub.SourceSessionIDs)
		}
	}
	for i, wr := range w.blobs.writers {
		var imp model.Import
		for _, im := range w.imports.m {
			if bytes.Contains(wr.bytes(), []byte(`"baseline_snapshot_id":"`+im.BaselineSnapshotID+`"`)) {
				imp = im
			}
		}
		if imp.ID == "" {
			// Der Verlierer eines Wettlaufs hat ein Manifest mit eigener Kennung geschrieben (SPEC-127 Nr. 18 a).
			continue
		}
		if got := normiereBytes(string(wr.bytes()), imp.ID, imp.BaselineSnapshotID); got != sollManifest {
			t.Errorf("Manifest %d weicht vom Einzelaufruf ab", i)
		}
	}
	if reflect.TypeOf(w.imp).Elem().NumField() != 1 {
		t.Error("der Dienst trägt mehr als die Konfig")
	}
}
