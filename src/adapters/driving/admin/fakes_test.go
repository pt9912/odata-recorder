package admin_test

import (
	"context"
	"errors"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// zaehlendeIDs ist die Hülle um FestIDs, die jede der fünf Operationen zählt und jede vergebene Korrelations-ID hält.
// vor läuft am Anfang von KorrelationsID und erhält die Zahl der bisherigen Aufrufe.
type zaehlendeIDs struct {
	mu       sync.Mutex
	fest     *supporttest.FestIDs
	session  int
	exch     int
	korr     int
	wsp      int
	snap     int
	vergeben []string
	vor      func()
}

func neueIDs() *zaehlendeIDs { return &zaehlendeIDs{fest: supporttest.NewFestIDs()} }

func (z *zaehlendeIDs) SessionID() string {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.session++
	return z.fest.SessionID()
}

func (z *zaehlendeIDs) ExchangeID() string {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.exch++
	return z.fest.ExchangeID()
}

func (z *zaehlendeIDs) KorrelationsID() string {
	z.mu.Lock()
	vor := z.vor
	z.mu.Unlock()
	if vor != nil {
		vor()
	}
	z.mu.Lock()
	defer z.mu.Unlock()
	z.korr++
	id := z.fest.KorrelationsID()
	z.vergeben = append(z.vergeben, id)
	return id
}

func (z *zaehlendeIDs) WorkspaceID() string {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.wsp++
	return z.fest.WorkspaceID()
}

func (z *zaehlendeIDs) SnapshotID() string {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.snap++
	return z.fest.SnapshotID()
}

// zahlen liefert die Zahl der Aufrufe je Operation: Korrelation und die vier übrigen zusammen.
func (z *zaehlendeIDs) zahlen() (korrelation, uebrige int) {
	z.mu.Lock()
	defer z.mu.Unlock()
	return z.korr, z.session + z.exch + z.wsp + z.snap
}

// alle liefert die Zahl aller fünf Operationen.
func (z *zaehlendeIDs) alle() int {
	k, u := z.zahlen()
	return k + u
}

// letzte liefert die zuletzt vergebene Korrelations-ID.
func (z *zaehlendeIDs) letzte() string {
	z.mu.Lock()
	defer z.mu.Unlock()
	if len(z.vergeben) == 0 {
		return ""
	}
	return z.vergeben[len(z.vergeben)-1]
}

// haengModus sagt, wie die Route eines routeFake auf einen Aufruf wartet.
type haengModus int

const (
	sofort     haengModus = iota // kehrt nach endeVerz zurück
	wartetFrei                   // wartet auf gibFrei
	wartetCtx                    // wartet auf das Ende von ctx oder auf gibFrei
)

// routeFake ist der Handler einer Route: er zählt, hält Anfrage und ctx jedes Aufrufs und liefert Antwort und Fehler
// wie eingestellt. aendere läuft im Aufruf mit der Anfrage, die die Route erhält. Mit begonnen ungleich nil blockiert
// der Aufruf nach modus (für den Pfad nurPfad, wenn gesetzt, sonst für jeden), zählt die laufenden Aufrufe und hält den
// Fehler von ctx nach dem Warten.
type routeFake struct {
	mu       sync.Mutex
	anfragen []admin.Anfrage
	ctxs     []context.Context
	antwort  admin.Antwort
	err      error
	aendere  func(a admin.Anfrage)

	modus     haengModus
	nurPfad   string
	endeVerz  time.Duration
	frei      chan struct{}
	begonnen  chan struct{}
	einmal    sync.Once
	laufend   atomic.Int64
	ctxFehler []error
	ctxEnden  []time.Time
}

func neueRoute() *routeFake { return &routeFake{antwort: admin.Antwort{Status: 200}} }

// neueHaengRoute liefert eine Route, die nach modus blockiert; der Test gibt sie am Ende frei.
func neueHaengRoute(t *testing.T, modus haengModus) *routeFake {
	t.Helper()
	f := neueRoute()
	f.modus, f.frei, f.begonnen = modus, make(chan struct{}), make(chan struct{}, 4096)
	f.antwort = admin.Antwort{Status: 200, Body: []byte("ok")}
	t.Cleanup(f.gibFrei)
	return f
}

func (f *routeFake) gibFrei() { f.einmal.Do(func() { close(f.frei) }) }

func (f *routeFake) handle(ctx context.Context, a admin.Anfrage) (admin.Antwort, error) {
	f.mu.Lock()
	f.anfragen = append(f.anfragen, a)
	f.ctxs = append(f.ctxs, ctx)
	aendere, antwort, err := f.aendere, f.antwort, f.err
	f.mu.Unlock()
	if aendere != nil {
		aendere(a)
	}
	if f.begonnen != nil {
		f.halte(ctx, a.Pfad)
	}
	return antwort, err
}

// halte meldet den Beginn des Aufrufs, wartet nach modus und hält den Fehler von ctx; der Aufruf zählt als laufend, bis
// er zurückkehrt.
func (f *routeFake) halte(ctx context.Context, pfad string) {
	f.laufend.Add(1)
	defer f.laufend.Add(-1)
	f.begonnen <- struct{}{}
	modus := f.modus
	if f.nurPfad != "" && pfad != f.nurPfad {
		modus = sofort
	}
	switch modus {
	case wartetFrei:
		<-f.frei
	case wartetCtx:
		select {
		case <-ctx.Done():
		case <-f.frei:
		}
	case sofort:
	}
	f.mu.Lock()
	f.ctxFehler = append(f.ctxFehler, ctx.Err())
	f.ctxEnden = append(f.ctxEnden, time.Now())
	f.mu.Unlock()
	time.Sleep(f.endeVerz)
}

// ctxEndenListe liefert die Zeitpunkte, zu denen die Aufrufe das Warten beendeten.
func (f *routeFake) ctxEndenListe() []time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxEnden)
}

// ctxFehlerListe liefert die Fehler von ctx, die die Aufrufe nach dem Warten sahen.
func (f *routeFake) ctxFehlerListe() []error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxFehler)
}

// warteBegonnen wartet, bis n Aufrufe der Route begonnen haben.
func (f *routeFake) warteBegonnen(t *testing.T, n int) {
	t.Helper()
	for range n {
		select {
		case <-f.begonnen:
		case <-time.After(5 * time.Second):
			t.Fatal("Frist von 5 s abgelaufen: der Request erreichte die Route nicht")
		}
	}
}

func (f *routeFake) aufrufe() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.anfragen)
}

// einzige liefert die Anfrage des einen Aufrufs; ok ist falsch, wenn die Route nicht genau einmal lief.
func (f *routeFake) einzige() (admin.Anfrage, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.anfragen) != 1 {
		return admin.Anfrage{}, false
	}
	return f.anfragen[0], true
}

// listenerFake zählt Accept, Close und Addr; der Fehler von Accept ist einstellbar, kein Aufruf gerät in Panik. Close
// wartet closeVerz, nachdem es gezählt hat, und liefert closeErr.
type listenerFake struct {
	mu        sync.Mutex
	accepts   int
	closes    int
	addrs     int
	acceptErr error
	closeVerz time.Duration
	closeErr  error
}

func neuerListener() *listenerFake {
	return &listenerFake{acceptErr: errors.New("listenerFake: kein Socket")}
}

func (l *listenerFake) Accept() (net.Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.accepts++
	return nil, l.acceptErr
}

func (l *listenerFake) Close() error {
	l.mu.Lock()
	l.closes++
	verz, err := l.closeVerz, l.closeErr
	l.mu.Unlock()
	time.Sleep(verz)
	return err
}

func (l *listenerFake) Addr() net.Addr {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.addrs++
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}
}

// zahlen liefert Accept, Close und Addr.
func (l *listenerFake) zahlen() (accept, closes, addr int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.accepts, l.closes, l.addrs
}
