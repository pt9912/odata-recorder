package record_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestDrainArgumente: ctx nil, leere Kennung und negative Frist sind ErrUngueltig und ändern
// nichts; eine Kennung ohne Exchange ist kein Fehler (B26).
func TestDrainArgumente(t *testing.T) {
	u := neu(t)
	var keinCtx context.Context
	for name, f := range map[string]func() error{
		"ctx nil":      func() error { _, err := u.r.Drain(keinCtx, "s1", 0); return err },
		"Kennung leer": func() error { _, err := u.r.Drain(context.Background(), "", 0); return err },
		"Frist -1":     func() error { _, err := u.r.Drain(context.Background(), "s1", -1); return err },
	} {
		if err := f(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
	u.rundlauf(anfrage("GET", "/a/x"))
	gelesen, listen := u.sessions.anzahlAufrufe(), u.exchanges.listAufrufe
	for _, id := range []string{"s1", "unbekannt"} {
		if b, err := u.r.Drain(context.Background(), id, 0); err != nil || b.Abgebrochen != 0 || b.NichtAngehaengt != 0 {
			t.Errorf("%s: %+v %v", id, b, err)
		}
	}
	if u.sessions.anzahlAufrufe() != gelesen || u.exchanges.listAufrufe != listen || len(u.sessions.unerwartet) > 0 {
		t.Errorf("Drain las die Ablage: %d Session-Aufrufe, %v", u.sessions.anzahlAufrufe()-gelesen, u.sessions.unerwartet)
	}
}

// TestDrainWartet: keine Annahme mehr, dauerhaft; Drain wartet auf den laufenden Exchange; andere
// Sessions bleiben unberührt (B27). synctest.Wait bestimmt, dass Drain wartet.
func TestDrainWartet(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		u := neu(t)
		u.sessions.setze("s2", "svc-a", model.SessionOpen)
		s1 := mitSession(anfrage("GET", "/a/x"), sessionHeader, "s1")
		ant, err := u.record(context.Background(), s1)
		if err != nil {
			t.Fatal(err)
		}
		d := u.drainNebenbei(context.Background(), "s1", time.Hour)
		synctest.Wait()
		_, err = u.record(context.Background(), s1)
		pruefeCode(t, err, model.CodeE5009, 409)
		u.rundlauf(mitSession(anfrage("GET", "/a/y"), sessionHeader, "s2"))
		synctest.Wait()
		select {
		case <-d:
			t.Fatal("Drain kehrte vor Close zurück")
		default:
		}
		liesUndSchliesse(t, ant.Body)
		if b := <-d; b.err != nil || b.b.Abgebrochen != 0 || b.b.NichtAngehaengt != 0 {
			t.Errorf("Drain: %+v", b)
		}
		if l := u.exchanges.angehaengt(); len(l) != 2 || l[0].SessionID != "s2" || l[1].SessionID != "s1" {
			t.Errorf("angehängt: %d", len(l))
		}
		_, err = u.record(context.Background(), s1)
		pruefeCode(t, err, model.CodeE5009, 409)
	})
}

// drainPhasen startet je einen Exchange von s1 in jeder Phase von SPEC-121 Nr. 13 und liefert
// die Ergebnisse der beiden ohne Antwort und die Bodies der drei mit Antwort.
func drainPhasen(t *testing.T, u *umgebung) ([]<-chan ergebnis, []io.ReadCloser) {
	t.Helper()
	json := "application/json"
	u.upstream.plane(
		&plan{lies: -1, warteAufCtx: true},
		&plan{lies: 3, warteAufCtx: true},
		jsonAntwort(koerper(`{"a":`).dann(make(chan struct{}), `1}`, nil)),
		jsonAntwort(koerper(`{"b":2}`)),
		jsonAntwort(koerper(`{"c":`).dann(make(chan struct{}), `3}`, nil)),
	)
	r1 := u.recordNebenbei(context.Background(), anfrage("GET", "/a/p1"))
	<-u.upstream.imSenden
	r2 := u.recordNebenbei(context.Background(), mitBody(anfrage("POST", "/a/p2"), json, `{"a":12345}`))
	<-u.upstream.imSenden
	var bodies []io.ReadCloser
	for i, a := range []struct{ pfad, typ, body string }{{"/a/p3", "", ""}, {"/a/p4", json, `{"a":1}`}, {"/a/p5", "text/plain", "x"}} {
		anf := anfrage("GET", a.pfad)
		if a.body != "" {
			anf = mitBody(anfrage("POST", a.pfad), a.typ, a.body)
		}
		ant, err := u.record(context.Background(), anf)
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			if b, _ := io.ReadAll(ant.Body); string(b) != `{"b":2}` {
				t.Fatalf("p4 gelesen %q", b)
			}
			<-u.blobs.commits
			<-u.blobs.commits
		} else if _, err := ant.Body.Read(make([]byte, 5)); err != nil {
			t.Fatal(err)
		}
		bodies = append(bodies, ant.Body)
	}
	return []<-chan ergebnis{r1, r2}, bodies
}

// TestDrainAbbruch: nach der Frist oder mit beendetem ctx bricht Drain jeden laufenden Exchange
// ab und wartet auf sein Ende; der Status folgt den Ergebnissen je Phase (B28).
func TestDrainAbbruch(t *testing.T) {
	for _, mitCtx := range []bool{false, true} {
		synctest.Test(t, func(t *testing.T) {
			beginn := time.Now()
			u := neu(t)
			ohne, bodies := drainPhasen(t, u)
			ctx, abbrechen := context.WithCancel(context.Background())
			defer abbrechen()
			frist := 20 * time.Millisecond
			if mitCtx {
				frist = time.Hour
			}
			d := u.drainNebenbei(ctx, "s1", frist)
			synctest.Wait()
			if mitCtx {
				abbrechen()
			}
			for _, r := range ohne {
				pruefeCode(t, (<-r).err, model.CodeE4001, 504)
			}
			synctest.Wait()
			select {
			case <-d:
				t.Fatal("Drain kehrte vor dem Ende der Exchanges zurück")
			default:
			}
			for _, b := range bodies {
				if err := b.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if b := <-d; b.err != nil || b.b.Abgebrochen != 5 {
				t.Errorf("Bericht %+v", b)
			}
			// Die Uhr der Blase läuft nur, wenn alle Goroutinen warten: die Frist von 1 h ist nie
			// abgelaufen, wenn der Abbruch von ctx wirkt.
			if vergangen := time.Since(beginn); vergangen >= time.Hour || !mitCtx && vergangen < frist {
				t.Errorf("ctx %v: Abbruch nach %v", mitCtx, vergangen)
			}
			soll := map[string]model.CaptureStatus{"/p1": model.CaptureIncomplete, "/p2": model.CaptureFailed,
				"/p3": model.CaptureIncomplete, "/p4": model.CaptureComplete, "/p5": model.CaptureFailed}
			l := u.exchanges.angehaengt()
			for _, e := range l {
				if e.Capture != soll[e.Path] {
					t.Errorf("ctx %v, %s: %s, erwartet %s", mitCtx, e.Path, e.Capture, soll[e.Path])
				}
			}
			if len(l) != 5 {
				t.Errorf("%d angehängt", len(l))
			}
		})
	}
}

// TestDrainBericht: Abgebrochen zählt nur einen Exchange, dessen Kontext ein Drain beendete, nicht
// einen, den der Client schon beendet hatte; ein abgebrochener, nicht angehängter Exchange zählt in
// beiden Feldern (B28, B20; SPEC-121 Nr. 10, 13).
func TestDrainBericht(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		u := neu(t)
		u.upstream.plane(jsonAntwort(koerper(`{"a":`).dann(make(chan struct{}), `1}`, nil)))
		ctx, abbrechen := context.WithCancel(context.Background())
		defer abbrechen()
		ant, err := u.record(ctx, anfrage("GET", "/a/x"))
		if err != nil {
			t.Fatal(err)
		}
		abbrechen()
		d := u.drainNebenbei(context.Background(), "s1", 0)
		synctest.Wait()
		if err := ant.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if b := <-d; b.err != nil || b.b.Abgebrochen != 0 || b.b.NichtAngehaengt != 0 {
			t.Errorf("Client-Abbruch vor Drain: %+v", b)
		}
		if e := u.letzter(); e.Capture != model.CaptureIncomplete {
			t.Errorf("Client-Abbruch vor Drain: %s", e.Capture)
		}
	})

	n := neu(t)
	n.exchanges.appendFehler[0] = errDisk
	n.upstream.plane(jsonAntwort(koerper(`{"a":`).dann(make(chan struct{}), `1}`, nil)))
	ant, err := n.record(context.Background(), anfrage("GET", "/a/x"))
	if err != nil {
		t.Fatal(err)
	}
	dr := n.drainNebenbei(context.Background(), "s1", 0)
	empfange(t, "Ende des Kontexts", n.upstream.gesehen()[0].done)
	schliesse(t, ant.Body)
	if b := empfange(t, "Drain", dr); b.err != nil || b.b.Abgebrochen != 1 || b.b.NichtAngehaengt != 1 {
		t.Errorf("Drain und Fehler beim Anhängen: %+v", b)
	}

	// Je Session: Abbruch und Verlust in s2 stehen nur im Bericht von s2, nicht in dem von s1 und
	// nicht in dem einer Kennung ohne Exchange (SPEC-121 Nr. 13).
	j := neu(t)
	j.sessions.setze("s2", "svc-a", model.SessionOpen)
	j.rundlauf(mitSession(anfrage("GET", "/a/x"), sessionHeader, "s1"))
	j.exchanges.appendFehler[1] = errDisk
	j.upstream.plane(jsonAntwort(koerper(`{"a":`).dann(make(chan struct{}), `1}`, nil)))
	ant, err = j.record(context.Background(), mitSession(anfrage("GET", "/a/y"), sessionHeader, "s2"))
	if err != nil {
		t.Fatal(err)
	}
	dr = j.drainNebenbei(context.Background(), "s2", 0)
	empfange(t, "Ende des Kontexts", j.upstream.gesehen()[1].done)
	schliesse(t, ant.Body)
	if b := empfange(t, "Drain s2", dr); b.err != nil || b.b.Abgebrochen != 1 || b.b.NichtAngehaengt != 1 {
		t.Errorf("Drain s2: %+v", b)
	}
	for _, id := range []string{"s1", "unbekannt"} {
		if b := j.bericht(id); b.Abgebrochen != 0 || b.NichtAngehaengt != 0 {
			t.Errorf("Bericht %s trägt %+v aus s2", id, b)
		}
	}
}

// TestDrainWiederholt: ein zweiter Drain kehrt mit demselben Bericht zurück; zwei gleichzeitige
// warten beide und zählen den Exchange einmal (B29).
func TestDrainWiederholt(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		beginn := time.Now()
		u := neu(t)
		u.upstream.plane(jsonAntwort(koerper(`{"a":`).dann(make(chan struct{}), `1}`, nil)))
		ant, err := u.record(context.Background(), anfrage("GET", "/a/x"))
		if err != nil {
			t.Fatal(err)
		}
		d1 := u.drainNebenbei(context.Background(), "s1", 0)
		d2 := u.drainNebenbei(context.Background(), "s1", 0)
		synctest.Wait()
		select {
		case <-d1:
			t.Fatal("erster Drain vor dem Ende")
		case <-d2:
			t.Fatal("zweiter Drain vor dem Ende")
		default:
		}
		if err := ant.Body.Close(); err != nil {
			t.Fatal(err)
		}
		for _, d := range []<-chan drainErgebnis{d1, d2} {
			if b := <-d; b.err != nil || b.b.Abgebrochen != 1 {
				t.Errorf("gleichzeitig: %+v", b)
			}
		}
		if b, err := u.r.Drain(context.Background(), "s1", time.Hour); err != nil || b.Abgebrochen != 1 {
			t.Errorf("wiederholt: %+v %v", b, err)
		}
		if vergangen := time.Since(beginn); vergangen != 0 {
			t.Errorf("wiederholter Drain wartete %v", vergangen)
		}
	})
}
