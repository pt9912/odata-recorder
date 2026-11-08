package admin_test

import (
	"context"
	"errors"
	"net"
	"sync"

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

// routeFake ist der Handler einer Route: er zählt, hält Anfrage und ctx jedes Aufrufs und liefert Antwort und Fehler
// wie eingestellt. aendere läuft im Aufruf mit der Anfrage, die die Route erhält.
type routeFake struct {
	mu       sync.Mutex
	anfragen []admin.Anfrage
	ctxs     []context.Context
	antwort  admin.Antwort
	err      error
	aendere  func(a admin.Anfrage)
}

func neueRoute() *routeFake { return &routeFake{antwort: admin.Antwort{Status: 200}} }

func (f *routeFake) handle(ctx context.Context, a admin.Anfrage) (admin.Antwort, error) {
	f.mu.Lock()
	f.anfragen = append(f.anfragen, a)
	f.ctxs = append(f.ctxs, ctx)
	aendere, antwort, err := f.aendere, f.antwort, f.err
	f.mu.Unlock()
	if aendere != nil {
		aendere(a)
	}
	return antwort, err
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

// listenerFake zählt Accept, Close und Addr; der Fehler von Accept ist einstellbar, kein Aufruf blockiert oder gerät in Panik.
type listenerFake struct {
	mu        sync.Mutex
	accepts   int
	closes    int
	addrs     int
	acceptErr error
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
	defer l.mu.Unlock()
	l.closes++
	return nil
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
