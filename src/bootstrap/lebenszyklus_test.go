package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

type nichtsSchliessen struct{ fehler error }

func (n nichtsSchliessen) Close() error { return n.fehler }

// literal baut den Handle eines hochgefahrenen Prozesses aus Fakes; der Server läuft, bis der
// Handle ihn anhält, und kehrt erst zurück, wenn der Test ihn freigibt.
func literal(modus string, frist time.Duration, uhr support.Clock, p *protokoll, s *fakeSessions, srv *fakeServer, ablage error) *bootstrap.Handle {
	srv.p, s.p = p, p
	return bootstrap.LiteralHandle(modus, frist, uhr, s, srv, &protSchliesser{h: &haken{p: p}, c: nichtsSchliessen{ablage}})
}

// gibFrei lässt Serve zurückkehren, sobald die Ablage geschlossen ist, spätestens nach 100 ms: ein
// Root, der nicht auf Serve wartet, schließt die Ablage vor dem Ende von Serve.
func gibFrei(p *protokoll, srv *fakeServer) {
	srv.freigabe = make(chan struct{})
	go func() {
		for i := 0; i < 50 && p.zaehle(pAblageZu) == 0; i++ {
			time.Sleep(2 * time.Millisecond)
		}
		close(srv.freigabe)
	}()
}

func ohneServe(ev []string) []string {
	return slices.DeleteFunc(slices.Clone(ev), func(e string) bool { return e == pServe })
}

// TestHerunterfahrenReihenfolge fängt B5: Modus {record, replay} × Aufruf {einmal, zweimal nacheinander,
// zweimal gleichzeitig}; dazu jede Quelle eines Fehlers {CloseAll, Fehlgeschlagen, Schliessen, Ablage}.
func TestHerunterfahrenReihenfolge(t *testing.T) {
	erwartet := map[string][]string{
		modusRecord: {"anhalten", pCloseAll, pSchliessen, "serve-ende", pAblageZu},
		modusReplay: {"anhalten", pSchliessen, "serve-ende", pAblageZu},
	}
	for _, modus := range modi() {
		for _, aufruf := range []string{"einmal", "nacheinander", "gleichzeitig"} {
			t.Run(modus+"/"+aufruf, func(t *testing.T) {
				p := &protokoll{}
				srv := &fakeServer{stop: make(chan struct{})}
				s := &fakeSessions{}
				h := literal(modus, time.Second, supporttest.NewFestUhr(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)), p, s, srv, nil)
				gibFrei(p, srv)
				// ctx ist abgebrochen: das Herunterfahren läuft unabhängig davon (SPEC-138 Nr. 8 b).
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				ergebnisse := make([]error, 2)
				switch aufruf {
				case "einmal":
					ergebnisse = ergebnisse[:1]
					ergebnisse[0] = h.Herunterfahren(ctx)
				case "nacheinander":
					ergebnisse[0], ergebnisse[1] = h.Herunterfahren(ctx), h.Herunterfahren(ctx)
				case "gleichzeitig":
					var wg sync.WaitGroup
					for i := range ergebnisse {
						wg.Add(1)
						go func() {
							defer wg.Done()
							ergebnisse[i] = h.Herunterfahren(ctx)
						}()
					}
					wg.Wait()
				}
				for _, err := range ergebnisse {
					if err != nil {
						t.Errorf("Fehler %v", err)
					}
				}
				if got := ohneServe(p.liste()); !slices.Equal(got, erwartet[modus]) {
					t.Errorf("Ereignisse %v, erwartet %v", got, erwartet[modus])
				}
				if modus == modusRecord && s.ctxBeiCloseAll() != nil {
					t.Errorf("CloseAll erhielt den abgebrochenen Kontext: %v", s.ctxBeiCloseAll())
				}
			})
		}
	}

	quellen := []struct {
		name   string
		setze  func(s *fakeSessions, srv *fakeServer)
		ablage error
		ist    func(error) bool
	}{
		{"CloseAll", func(s *fakeSessions, _ *fakeServer) { s.fehler = errPhase }, nil,
			func(e error) bool { return errors.Is(e, errPhase) }},
		{"Fehlgeschlagen", func(s *fakeSessions, _ *fakeServer) { s.bericht = inbound.CloseBericht{Fehlgeschlagen: 1} }, nil,
			func(e error) bool { return e != nil }},
		{"Schliessen", func(_ *fakeSessions, srv *fakeServer) { srv.fehler = httpodata.ErrFristAbgelaufen }, nil,
			func(e error) bool { return errors.Is(e, httpodata.ErrFristAbgelaufen) }},
		{"Ablage", func(*fakeSessions, *fakeServer) {}, io.ErrClosedPipe,
			func(e error) bool { return errors.Is(e, io.ErrClosedPipe) }},
	}
	for _, q := range quellen {
		t.Run("Fehler/"+q.name, func(t *testing.T) {
			p := &protokoll{}
			srv := &fakeServer{stop: make(chan struct{})}
			s := &fakeSessions{}
			q.setze(s, srv)
			h := literal(modusRecord, time.Second, supporttest.NewFestUhr(time.Now()), p, s, srv, q.ablage)
			gibFrei(p, srv)
			err := h.Herunterfahren(context.Background())
			if !q.ist(err) {
				t.Errorf("Fehler %v", err)
			}
			if got := ohneServe(p.liste()); !slices.Equal(got, erwartet[modusRecord]) {
				t.Errorf("Ereignisse %v, erwartet %v", got, erwartet[modusRecord])
			}
		})
	}
}

// TestFristRest fängt B6: Dauer von CloseAll {0, kleiner als die Frist, gleich der Frist, größer} im
// Modus record, dazu replay ohne CloseAll.
func TestFristRest(t *testing.T) {
	const frist = 10 * time.Second
	faelle := []struct {
		modus  string
		dauer  time.Duration
		erwart time.Duration
	}{
		{modusRecord, 0, frist}, {modusRecord, 3 * time.Second, 7 * time.Second},
		{modusRecord, frist, 0}, {modusRecord, frist + 2*time.Second, 0},
		{modusReplay, 0, frist},
	}
	for _, f := range faelle {
		t.Run(fmt.Sprintf("%s/%s", f.modus, f.dauer), func(t *testing.T) {
			uhr := supporttest.NewFestUhr(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
			srv := &fakeServer{stop: make(chan struct{})}
			s := &fakeSessions{beiCloseAll: func() { uhr.Advance(f.dauer) }}
			h := literal(f.modus, frist, uhr, &protokoll{}, s, srv, nil)
			if err := h.Herunterfahren(context.Background()); err != nil {
				t.Fatal(err)
			}
			if got := srv.frist(); !slices.Equal(got, []time.Duration{f.erwart}) {
				t.Errorf("Schliessen mit %v, erwartet %v", got, f.erwart)
			}
		})
	}
}
