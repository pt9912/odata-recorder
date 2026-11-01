package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
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

// gibFrei lässt Serve zurückkehren, sobald die Ablage geschlossen ist, spätestens nach 100 ms.
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
// zweimal gleichzeitig} × Wert {nil, Fehler} = 12, jeder Aufruf liefert denselben Wert wie der erste
// (bei Fehler dieselben Quellen {CloseAll nur in record, Schliessen, Ablage}); dazu jede Quelle eines
// Fehlers {CloseAll, Fehlgeschlagen, Schliessen, Ablage}.
func TestHerunterfahrenReihenfolge(t *testing.T) {
	erwartet := map[string][]string{
		modusRecord: {"anhalten", pCloseAll, pSchliessen, "serve-ende", pAblageZu},
		modusReplay: {"anhalten", pSchliessen, "serve-ende", pAblageZu},
	}
	n := 0
	for _, modus := range modi() {
		for _, aufruf := range []string{"einmal", "nacheinander", "gleichzeitig"} {
			for _, wert := range []string{"nil", "Fehler"} {
				n++
				t.Run(modus+"/"+aufruf+"/"+wert, func(t *testing.T) {
					herunterfahrenAufruf(t, modus, aufruf, wert == "Fehler", erwartet[modus])
				})
			}
		}
	}
	if n != 12 {
		t.Errorf("Menge hat %d Elemente, erwartet 12", n)
	}
	herunterfahrenQuellen(t, erwartet[modusRecord])
}

// herunterfahrenAufruf ruft Herunterfahren so oft, wie aufruf sagt, und hält die Ergebnisse gegen den
// Wert: nil, oder ein Fehler, der jede Quelle nennt und bei jedem Aufruf derselbe ist.
func herunterfahrenAufruf(t *testing.T, modus, aufruf string, mitFehler bool, soll []string) {
	p := &protokoll{}
	srv := &fakeServer{stop: make(chan struct{})}
	s := &fakeSessions{}
	var ablage error
	quellen := []error(nil)
	if mitFehler {
		ablage, srv.fehler = io.ErrClosedPipe, httpodata.ErrFristAbgelaufen
		quellen = []error{io.ErrClosedPipe, httpodata.ErrFristAbgelaufen}
		if modus == modusRecord {
			s.fehler = errPhase
			quellen = append(quellen, errPhase)
		}
	}
	h := literal(modus, time.Second, supporttest.NewFestUhr(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)), p, s, srv, ablage)
	gibFrei(p, srv)
	// ctx ist abgebrochen und trägt einen Wert: das Herunterfahren läuft unabhängig vom Abbruch (SPEC-138
	// Nr. 8 b) und behält die Werte des Aufrufers.
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxMarke{}, "herunterfahren"))
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
	for i, err := range ergebnisse {
		if !mitFehler && err != nil {
			t.Errorf("Aufruf %d: Fehler %v", i, err)
		}
		for _, q := range quellen {
			if !errors.Is(err, q) {
				t.Errorf("Aufruf %d: Ergebnis %v nennt %v nicht", i, err, q)
			}
		}
		if err != ergebnisse[0] {
			t.Errorf("Aufruf %d liefert %v, der erste %v", i, err, ergebnisse[0])
		}
	}
	if got := ohneServe(p.liste()); !slices.Equal(got, soll) {
		t.Errorf("Ereignisse %v, erwartet %v", got, soll)
	}
	if modus == modusRecord && s.ctxBeiCloseAll() != nil {
		t.Errorf("CloseAll erhielt den abgebrochenen Kontext: %v", s.ctxBeiCloseAll())
	}
	if modus == modusRecord && s.wertBeiCloseAll() != "herunterfahren" {
		t.Errorf("CloseAll erhielt nicht die Werte des Kontexts (Wert %v)", s.wertBeiCloseAll())
	}
}

func herunterfahrenQuellen(t *testing.T, soll []string) {
	t.Helper()
	erwartet := map[string][]string{modusRecord: soll}

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

// TestZugaengeGleichzeitig fängt B12 (SPEC-139 Nr. 10): Adresse, Fehler, Wartung, Record und Sessions werden
// von vier Goroutinen gelesen, während Hochfahren läuft; der Haken von Listen hält Hochfahren an, bis
// die Leser 400 Runden hinter sich haben, die Zugriffe liegen also vor dem Schreiben der Adresse. Fänger
// ist der Detektor von make test-race (der Wirkungs-Fänger fehlt: die Wirkung eines fehlenden Schutzes
// ist das Datenrennen selbst); ohne -race prüft der Test nur den Endwert.
func TestZugaengeGleichzeitig(t *testing.T) {
	ctx := context.Background()
	h := neuHaken()
	hd, err := bootstrap.BuildMit(ctx, basisKonfig(t.TempDir(), modusRecord, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = hd.Herunterfahren(ctx) })
	const leser, runden = 4, 100
	var gelesen atomic.Int64
	var ende atomic.Bool
	var fertig sync.WaitGroup
	for range leser {
		fertig.Go(func() {
			for !ende.Load() {
				_ = hd.Adresse()
				_ = hd.Fehler()
				_ = hd.Wartung()
				_ = hd.Record()
				_ = hd.Sessions()
				gelesen.Add(1)
				runtime.Gosched()
			}
		})
	}
	h.vorListen = func() {
		for gelesen.Load() < leser*runden {
			runtime.Gosched()
		}
	}
	err = hd.Hochfahren(ctx)
	ende.Store(true)
	fertig.Wait()
	if err != nil {
		t.Fatal(err)
	}
	if hd.Adresse() == "" || hd.Adresse() != h.listener[0].Addr().String() {
		t.Errorf("Adresse %q", hd.Adresse())
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
