package replay_test

import (
	"context"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// TestGeteilterReplayer hält Z54 für den Zustand in der Instanz (SPEC-135 Nr. 17: Replay ist für gleichzeitigen Aufruf
// sicher): EIN Replayer, 16 Goroutinen mit je 20 Aufrufen ohne gegenseitige Ordnung (eine Startsperre gibt alle frei),
// 6 verschiedene Requests (jeder Pfad hat einen Exchange mit eigener SequenceNo und eigenem Body), Strategie first
// (kein Fortschreiben, kein Wettlauf um den Cursor). Jeder Aufruf wird einzeln geprüft: Status, SequenceNo und Body
// gehören zum Request des Aufrufs (Wirkung), und unter make test-race meldet der Detektor jeden ungeordneten Zugriff
// auf Zustand, den die Aufrufe teilen (die Instanz, die Konfig und deren Slices und Maps). Die Fänger der Fenster
// bauen je Aufruf einen eigenen Dienst und sehen diesen Zustand nicht. Dazu gehören zwei Dienste in der Konfig, damit
// ein Schreibzugriff auf die Liste der Dienste wirkt.
func TestGeteilterReplayer(t *testing.T) {
	const (
		goroutinen = 16
		aufrufe    = 20
		pfade      = 6
	)
	u := neu(t)
	u.k.Strategie = model.StrategyFirst
	u.k.Dienste = append(u.k.Dienste, replay.Dienst{ID: "svc-z", Root: "/other"})
	var es []model.Exchange
	anfragen := make([]inbound.Anfrage, pfade)
	for i := range anfragen {
		a := anf()
		a.Pfad = root + "/P" + strconv.Itoa(i)
		anfragen[i] = a
		es = append(es, u.austausch(i+1, u.keyFuer(a, svcID, "/P"+strconv.Itoa(i)), true))
	}
	u.lege(es...)
	r := u.replayer()
	var fertig atomic.Int64
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < goroutinen; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < aufrufe; j++ {
				i := (g + j) % pfade
				resp, err := r.Replay(context.Background(), anfragen[i])
				if err != nil {
					t.Errorf("Pfad %d: %v", i, err)
					continue
				}
				body, lesefehler := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				fertig.Add(1)
				if seq := resp.Header["X-Seq"]; resp.Status != 200 || len(seq) != 1 || seq[0] != strconv.Itoa(i+1) ||
					lesefehler != nil || string(body) != string(bodyVon(i+1)) {
					t.Errorf("Pfad %d: Status %d, X-Seq %v, Body %q (%v); erwartet %d", i, resp.Status, seq, body, lesefehler, i+1)
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := fertig.Load(); got != goroutinen*aufrufe {
		t.Errorf("Menge: %d Aufrufe, erwartet %d", got, goroutinen*aufrufe)
	}
}
