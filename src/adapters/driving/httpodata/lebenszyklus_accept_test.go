package httpodata_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// kippListener ist ein Listener mit Zustand wie ein echter: das erste Close gelingt und weckt
// Accept, jedes weitere liefert net.ErrClosed. Accept liefert den Fehler, den ein Test einspeist,
// oder nach dem Schließen beiClose (nil: ein OpError mit net.ErrClosed); danach ruft es
// nachFehler, ehe der Fehler zurückkehrt.
type kippListener struct {
	zu          chan struct{}
	spontan     chan error
	acceptiert  chan struct{}
	erster      sync.Once
	mu          sync.Mutex
	geschlossen bool
	beiClose    error
	nachFehler  func()
}

func neuKippListener() *kippListener {
	return &kippListener{zu: make(chan struct{}), spontan: make(chan error, 1), acceptiert: make(chan struct{})}
}

func (l *kippListener) Accept() (net.Conn, error) {
	l.erster.Do(func() { close(l.acceptiert) })
	var err error
	select {
	case err = <-l.spontan:
	case <-l.zu:
		err = l.beiClose
		if err == nil {
			err = &net.OpError{Op: "accept", Err: net.ErrClosed}
		}
	}
	if l.nachFehler != nil {
		l.nachFehler()
	}
	return nil, err
}

func (l *kippListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.geschlossen {
		return net.ErrClosed
	}
	l.geschlossen = true
	close(l.zu)
	return nil
}

func (l *kippListener) Addr() net.Addr { return &net.TCPAddr{} }

// L7, Klasse der Fehler des Accept (SPEC-136 Nr. 12): nur ein Fehler, den das Schließen durch
// Anhalten oder Schliessen selbst auslöste, ist nil; jeder andere Fehler des Accept ist der
// Rückgabewert, wie spät Anhalten oder Schliessen auch folgt. Drei Teilmengen, 35 Elemente:
//
//	A: Fehler ohne Bezug zum Schließen (3 Arten) × Zeitpunkt von Anhalten (5) = 15, Fehler
//	B: Fehler durch äußeres Schließen (2 Arten von net.ErrClosed) × Zeitpunkt (5) = 10, Fehler
//	C: Anhalten oder Schliessen zuerst (2) × Fehler des Accept danach (2 Arten net.ErrClosed → nil,
//	   3 andere Arten → Fehler) = 10
func TestServeFehlerKlasse(t *testing.T) {
	neu := func(t *testing.T) (*httpodata.Server, *kippListener) {
		l := neuKippListener()
		srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: l})
		if err != nil {
			t.Fatal(err)
		}
		return srv, l
	}
	start := func(srv *httpodata.Server) chan error {
		c := make(chan error, 1)
		go func() { c <- srv.Serve() }()
		return c
	}
	anAcceptErfolgt := func(l *kippListener) { <-l.acceptiert }

	andere := map[string]error{
		"einfacher Fehler":     errors.New("accept kaputt"),
		"umhüllter Fehler":     fmt.Errorf("accept: %w", io.ErrUnexpectedEOF),
		"OpError ohne Schluss": &net.OpError{Op: "accept", Err: errors.New("kaputt")},
	}
	geschlossen := map[string]error{
		"net.ErrClosed":         net.ErrClosed,
		"OpError mit ErrClosed": &net.OpError{Op: "accept", Err: net.ErrClosed},
	}
	// Die Zeitpunkte von Anhalten oder Schliessen relativ zum Fehler; ausloeser erzeugt den Fehler.
	zeitpunkte := []struct {
		name string
		los  func(t *testing.T, srv *httpodata.Server, l *kippListener, ausloeser func()) error
	}{
		{"kein Anhalten", func(t *testing.T, srv *httpodata.Server, l *kippListener, ausloeser func()) error {
			c := start(srv)
			anAcceptErfolgt(l)
			ausloeser()
			return serveEnde(t, c, "kein Anhalten")
		}},
		{"Serve kehrt zurück, dann Anhalten", func(t *testing.T, srv *httpodata.Server, l *kippListener, ausloeser func()) error {
			c := start(srv)
			anAcceptErfolgt(l)
			ausloeser()
			err := serveEnde(t, c, "Serve vor Anhalten")
			srv.Anhalten()
			return err
		}},
		{"Anhalten im Accept vor der Rückkehr", func(t *testing.T, srv *httpodata.Server, l *kippListener, ausloeser func()) error {
			l.nachFehler = srv.Anhalten
			c := start(srv)
			anAcceptErfolgt(l)
			ausloeser()
			return serveEnde(t, c, "Anhalten im Accept")
		}},
		{"Schliessen gleichzeitig am Haltepunkt", func(t *testing.T, srv *httpodata.Server, l *kippListener, ausloeser func()) error {
			// Shutdown wartet, bis Serve zurückkehrt; Schliessen läuft deshalb in einer
			// eigenen Goroutine und hat den Server angehalten, ehe der Haltepunkt freigegeben wird.
			erreicht, weiter := make(chan struct{}), make(chan struct{})
			l.nachFehler = func() { close(erreicht); <-weiter }
			c := start(srv)
			anAcceptErfolgt(l)
			ausloeser()
			<-erreicht
			ende := make(chan struct{})
			go func() { _ = srv.Schliessen(0); close(ende) }()
			time.Sleep(100 * time.Millisecond)
			close(weiter)
			<-ende
			return serveEnde(t, c, "Schliessen am Haltepunkt")
		}},
		{"Anhalten gleichzeitig am Haltepunkt", func(t *testing.T, srv *httpodata.Server, l *kippListener, ausloeser func()) error {
			erreicht, weiter := make(chan struct{}), make(chan struct{})
			l.nachFehler = func() { close(erreicht); <-weiter }
			c := start(srv)
			anAcceptErfolgt(l)
			ausloeser()
			<-erreicht
			srv.Anhalten()
			close(weiter)
			return serveEnde(t, c, "Anhalten am Haltepunkt")
		}},
	}
	gleich(t, "Zahl der Zeitpunkte", len(zeitpunkte), 5)
	istFehler := func(err, art error) bool {
		return err != nil && !errors.Is(err, http.ErrServerClosed) && errors.Is(err, art)
	}

	n := 0
	for artName, art := range andere {
		for _, z := range zeitpunkte {
			n++
			srv, l := neu(t)
			err := z.los(t, srv, l, func() { l.spontan <- art })
			if !istFehler(err, art) {
				t.Errorf("A %s, %s: Serve lieferte %v, erwartet den Fehler des Accept", artName, z.name, err)
			}
		}
	}
	for artName, art := range geschlossen {
		for _, z := range zeitpunkte {
			n++
			srv, l := neu(t)
			l.beiClose = art
			err := z.los(t, srv, l, func() { _ = l.Close() })
			if !istFehler(err, art) {
				t.Errorf("B %s, %s: Serve lieferte %v, erwartet den Fehler des Accept", artName, z.name, err)
			}
		}
	}
	schliesser := map[string]func(*httpodata.Server){
		"Anhalten":     func(s *httpodata.Server) { s.Anhalten() },
		"Schliessen 0": func(s *httpodata.Server) { _ = s.Schliessen(0) },
	}
	for schName, schliesse := range schliesser {
		for artName, art := range geschlossen {
			n++
			srv, l := neu(t)
			l.beiClose = art
			c := start(srv)
			anAcceptErfolgt(l)
			schliesse(srv)
			if err := serveEnde(t, c, schName); err != nil {
				t.Errorf("C %s, %s: Serve lieferte %v, erwartet nil", schName, artName, err)
			}
		}
		for artName, art := range andere {
			n++
			srv, l := neu(t)
			l.beiClose = art
			c := start(srv)
			anAcceptErfolgt(l)
			schliesse(srv)
			if err := serveEnde(t, c, schName); !istFehler(err, art) {
				t.Errorf("C %s, %s: Serve lieferte %v, erwartet den Fehler des Accept", schName, artName, err)
			}
		}
	}
	gleich(t, "Zahl der Elemente der Klasse", n, 35)
}
