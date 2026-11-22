package replay_test

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// schranke hält n Aufrufe an, bis der letzte eintrifft; eine Frist beendet das Warten, damit ein Fehler den Test nicht hängen lässt.
type schranke struct {
	n     int32
	da    atomic.Int32
	frei  chan struct{}
	t     *testing.T
	frist time.Duration
}

func neueSchranke(t *testing.T, n int) *schranke {
	return &schranke{n: int32(n), frei: make(chan struct{}), t: t, frist: 20 * time.Second}
}

func (s *schranke) warte() {
	if s.da.Add(1) == s.n {
		close(s.frei)
	}
	select {
	case <-s.frei:
	case <-time.After(s.frist):
		s.t.Errorf("Schranke: %d von %d Aufrufen eingetroffen", s.da.Load(), s.n)
	}
}

// ctxNummer trägt die Nummer der Goroutine im ctx ihres Aufrufs; der Fake hält das ctx jedes StartSpan fest.
type ctxNummer struct{}

// gleichzeitigerLauf führt n Aufrufe gleichzeitig (SPEC-296 Nr. 9 a): Goroutine g hat den Cursor rep-g der Strategie g mod 4 und fragt
// einen Treffer oder, in jedem zweiten Viertel, einen unbekannten Key (kein_treffer). Der Treffpunkt liegt am Ende jeder Transaktion
// (nachTransaktion): alle n Transaktionen sind gelaufen, bevor ein Aufruf Set ruft; ein Ergebnis in geteiltem Zustand ist dann das des
// letzten Schreibers. Danach prüft das Buch je Span Strategie und Ergebnis seines Aufrufs.
func gleichzeitigerLauf(t *testing.T, n int) {
	t.Helper()
	strategien := []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast}
	u := neu(t)
	u.lege(u.austausch(1, u.keyA(), true))
	z := u.st.stand()
	for g := 0; g < n; g++ {
		z.cur[wsID] = append(z.cur[wsID], model.ReplayCursor{WorkspaceID: wsID, ReplaySessionID: "rep-" + strconv.Itoa(g), RecordingSessionID: sesID,
			ServiceID: svcID, Strategy: strategien[g%4]})
	}
	u.st.setze(z)
	sch := neueSchranke(t, n)
	u.st.nachTransaktion = sch.warte
	r := u.replayer()
	var wg sync.WaitGroup
	for g := 0; g < n; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a := anf()
			if (g/4)%2 == 1 {
				a = anfC()
			}
			a.Header["X-Recorder-Replay-Session"] = []string{"rep-" + strconv.Itoa(g)}
			resp, err := r.Replay(context.WithValue(context.Background(), ctxNummer{}, g), a)
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
	}
	wg.Wait()
	b := buecher(u.tel)
	if len(b) != n || u.tel.Anzahl(supporttest.OpStartSpan) != n || u.tel.Anzahl(supporttest.OpEnd) != n {
		t.Fatalf("%d Bücher, %d StartSpan, %d End; erwartet je %d", len(b), u.tel.Anzahl(supporttest.OpStartSpan), u.tel.Anzahl(supporttest.OpEnd), n)
	}
	gesehen := map[int]bool{}
	for _, sb := range b {
		g, _ := sb.ctx().Value(ctxNummer{}).(int)
		erg := ergTreffer
		if (g/4)%2 == 1 {
			erg = ergKeinTreffer
		}
		want := []support.Attr{support.Str(attrStrategie, string(strategien[g%4])), support.Str(attrErgebnis, erg)}
		if sb.enden != 1 || len(sb.setze) != 1 || len(sb.setze[0]) != 2 || sb.setze[0][0] != want[0] || sb.setze[0][1] != want[1] {
			t.Errorf("Aufruf %d: Span mit %d End und Set %v, erwartet ein End und %v", g, sb.enden, sb.setze, want)
		}
		gesehen[g] = true
	}
	if len(gesehen) != n {
		t.Errorf("%d verschiedene Aufrufe im Buch, erwartet %d", len(gesehen), n)
	}
}

// TestSpanGleichzeitig hält W13 (SPEC-296 Nr. 9 a): n gleichzeitige Replay ergeben n StartSpan und n End, und jeder Span trägt Strategie und
// Ergebnis seines Aufrufs; Menge M-Gleichzeitig, 16 Goroutinen × 4 Strategien × 2 Ergebnisse (Treffer, kein_treffer) in 20 Läufen.
func TestSpanGleichzeitig(t *testing.T) {
	for i := 0; i < 20; i++ {
		gleichzeitigerLauf(t, 16)
	}
}
