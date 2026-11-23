package record_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

// schranke hält n Aufrufe an, bis der letzte eintrifft; eine Frist beendet das Warten, damit ein Fehler den Test nicht hängen lässt.
type schranke struct {
	n    int32
	da   atomic.Int32
	frei chan struct{}
	t    *testing.T
}

func neueSchranke(t *testing.T, n int) *schranke {
	return &schranke{n: int32(n), frei: make(chan struct{}), t: t}
}

func (s *schranke) warte() {
	if s.da.Add(1) == s.n {
		close(s.frei)
	}
	select {
	case <-s.frei:
	case <-time.After(zeitgrenze):
		s.t.Errorf("Schranke: %d von %d Aufrufen eingetroffen", s.da.Load(), s.n)
	}
}

// ctxNummer trägt die Nummer der Goroutine im ctx ihres Aufrufs; der Fake hält das ctx jedes StartSpan fest.
type ctxNummer struct{}

// pfadUpstream hält jeden Aufruf an der Schranke an und antwortet nach dem Pfad: /c mit einer vollständigen Antwort, /i ohne Antwort,
// /f mit einer vollständigen Antwort auf einen Request, den die Aufnahme verwirft.
type pfadUpstream struct{ sch *schranke }

func (p pfadUpstream) Senden(_ context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	p.sch.warte()
	if a.Body != nil {
		_, _ = io.Copy(io.Discard, a.Body)
	}
	if a.Pfad == "/i" {
		return outbound.UpstreamAntwort{}, outbound.ErrUpstreamNichtErreichbar
	}
	return outbound.UpstreamAntwort{Status: 200, Header: map[string][]string{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader("{}"))}, nil
}

// gleichzeitigerLauf führt n Exchanges gleichzeitig (SPEC-295 Nr. 9 a): Goroutine g wählt mit dem Header die Session s1 oder s2 (g mod 2) und
// endet mit complete, incomplete oder failed (g mod 3). Die Schranke liegt in Senden: alle n haben StartSpan gerufen, bevor einer Set oder End
// ruft; ein Span, den Exchanges teilen, trägt dann die Werte des letzten. Danach prüft das Buch je Span Kennung und Status seines Aufrufs.
func gleichzeitigerLauf(t *testing.T, n int) {
	t.Helper()
	sessions, status := []string{"s1", "s2"}, []string{"complete", "incomplete", "failed"}
	sch := neueSchranke(t, n)
	u := neu(t, func(k *record.Konfig) { k.Upstream = pfadUpstream{sch} })
	u.sessions.setze("s2", "svc-a", model.SessionOpen)
	var wg sync.WaitGroup
	for g := 0; g < n; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a := anfrage("GET", "/a/"+[]string{"c", "i", "f"}[g%3])
			if g%3 == 2 {
				a = mitBody(anfrage("POST", "/a/f"), "text/plain", "x")
			}
			a = mitSession(a, sessionHeader, sessions[g%2])
			ant, err := u.r.Record(context.WithValue(context.Background(), ctxNummer{}, g), a)
			if err == nil {
				_, _ = io.ReadAll(ant.Body)
				_ = ant.Body.Close()
			}
		}()
	}
	mitZeitgrenze(t, "Record", wg.Wait)
	b := buecher(u.tel)
	if len(b) != n || u.tel.Anzahl(supporttest.OpStartSpan) != n || u.tel.Anzahl(supporttest.OpEnd) != n {
		t.Fatalf("%d Bücher, %d StartSpan, %d End; erwartet je %d", len(b), u.tel.Anzahl(supporttest.OpStartSpan), u.tel.Anzahl(supporttest.OpEnd), n)
	}
	gesehen := map[int]bool{}
	for _, sb := range b {
		g, _ := sb.ctx().Value(ctxNummer{}).(int)
		pruefeBuch(t, fmt.Sprintf("Aufruf %d", g), sb, sessions[g%2], status[g%3])
		gesehen[g] = true
	}
	if len(gesehen) != n {
		t.Errorf("%d verschiedene Aufrufe im Buch, erwartet %d", len(gesehen), n)
	}
}

// TestSpanGleichzeitig hält C10 (SPEC-295 Nr. 9 a): n gleichzeitige Exchanges ergeben n StartSpan und n End, jeder Span trägt Kennung und Status seines
// Exchange, keiner endet zweimal; Menge M-Gleichzeitig, 16 Goroutinen × 2 Sessions × 3 Status in 20 Läufen. Ein Span oder End, den Exchanges teilen,
// fängt der Test in jedem Lauf. Eine Variable ohne Ordnung, die Status, Kennung oder Zähler eines Spans trägt, fängt er ohne Detektor nicht verlässlich
// (den Status in keinem Prozess, die Kennung mit weitem Fenster nur bei mehreren Prozessoren); den ungeordneten Zugriff meldet allein der Detektor
// von `make test-race`, das Maß steht im Negativ (8) des Plans.
func TestSpanGleichzeitig(t *testing.T) {
	for i := 0; i < 20; i++ {
		gleichzeitigerLauf(t, 16)
	}
}

// haltUpstream hält jeden Aufruf bis zum Ende seines ctx an und meldet jeden Eintritt.
type haltUpstream struct{ imSenden chan struct{} }

func (h haltUpstream) Senden(ctx context.Context, _ outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	h.imSenden <- struct{}{}
	<-ctx.Done()
	return outbound.UpstreamAntwort{}, fmt.Errorf("upstream-fake: %w", ctx.Err())
}

// drainLauf lässt 16 Exchanges in Senden laufen und ruft Drain(0) gleichzeitig mit 16 weiteren Record; jedes End braucht Zeit. Bei der Rückkehr von
// Drain ist End so oft gerufen wie StartSpan.
func drainLauf(t *testing.T) {
	t.Helper()
	h := haltUpstream{imSenden: make(chan struct{}, 64)}
	u, j := neuMitJournal(t, func(k *record.Konfig) { k.Upstream = h })
	j.vorEnd = func() { time.Sleep(3 * time.Millisecond) }
	var wg sync.WaitGroup
	ruf := func() {
		defer wg.Done()
		_, _ = u.r.Record(context.Background(), anfrage("GET", "/a/x"))
	}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go ruf()
	}
	for i := 0; i < 16; i++ {
		empfange(t, "Senden", h.imSenden)
	}
	start := make(chan struct{})
	var starts, enden int
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		if _, err := u.r.Drain(context.Background(), "s1", 0); err != nil {
			t.Errorf("Drain: %v", err)
		}
		starts, enden = j.Anzahl(supporttest.OpStartSpan), j.Anzahl(supporttest.OpEnd)
	}()
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			<-start
			ruf()
		}()
	}
	close(start)
	mitZeitgrenze(t, "Record und Drain", wg.Wait)
	if starts < 16 || enden != starts {
		t.Errorf("bei der Rückkehr von Drain %d StartSpan und %d End; erwartet gleich viele, mindestens 16", starts, enden)
	}
}

// TestSpanNachDrain hält C11 (SPEC-295 Nr. 3 b, 9 b): nach der Rückkehr von Drain ist jeder Span der Session beendet, auch gegen gleichzeitig beginnende
// Record; 20 Läufe. Der Test hält die Rückkehr gegen das End, das Zeit braucht; die Rate je Prozess steht im Plan.
func TestSpanNachDrain(t *testing.T) {
	for i := 0; i < 20; i++ {
		drainLauf(t)
	}
}
