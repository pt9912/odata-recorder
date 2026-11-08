package admin_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// startServe startet Serve in einer Goroutine; der Kanal trägt die Rückgabe.
func startServe(srv *admin.Server) chan error {
	c := make(chan error, 1)
	go func() { c <- srv.Serve() }()
	return c
}

// TestServeEndetMitNil hält S1 und die Randfälle von S2: nach Anhalten oder Schliessen, auch davor, ist die Rückgabe
// nil, nie http.ErrServerClosed; ein von außen geschlossener oder geschlossen übergebener Listener ist ein Fehler.
// Sechs Elemente.
func TestServeEndetMitNil(t *testing.T) {
	neu := func(t *testing.T) (*admin.Server, net.Listener, *meldeListener) {
		raw, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		lis := neuMeldeListener(raw)
		return neuMitListener(t, lis), raw, lis
	}
	faelle := []struct {
		name   string
		los    func(srv *admin.Server, raw net.Listener, lis *meldeListener) chan error
		istNil bool
	}{
		{"Serve läuft, dann Anhalten", func(srv *admin.Server, _ net.Listener, lis *meldeListener) chan error {
			c := startServe(srv)
			<-lis.acceptiert
			srv.Anhalten()
			return c
		}, true},
		{"Serve läuft, dann Schliessen(0)", func(srv *admin.Server, _ net.Listener, lis *meldeListener) chan error {
			c := startServe(srv)
			<-lis.acceptiert
			_, _ = schliesse(t, srv, 0)
			return c
		}, true},
		{"Anhalten vor Serve", func(srv *admin.Server, _ net.Listener, _ *meldeListener) chan error {
			srv.Anhalten()
			return startServe(srv)
		}, true},
		{"Schliessen(0) vor Serve", func(srv *admin.Server, _ net.Listener, _ *meldeListener) chan error {
			_, _ = schliesse(t, srv, 0)
			return startServe(srv)
		}, true},
		{"Listener von außen geschlossen", func(srv *admin.Server, raw net.Listener, lis *meldeListener) chan error {
			c := startServe(srv)
			<-lis.acceptiert
			_ = raw.Close()
			return c
		}, false},
		{"Listener geschlossen übergeben", func(srv *admin.Server, raw net.Listener, _ *meldeListener) chan error {
			_ = raw.Close()
			return startServe(srv)
		}, false},
	}
	gleichAuf(t, "Zahl der Enden", len(faelle), 6)
	for _, f := range faelle {
		srv, raw, lis := neu(t)
		err := serveEnde(t, f.los(srv, raw, lis), f.name)
		switch {
		case f.istNil && err != nil:
			t.Errorf("%s: Serve lieferte %v, erwartet nil", f.name, err)
		case !f.istNil && (err == nil || errors.Is(err, http.ErrServerClosed)):
			t.Errorf("%s: Serve lieferte %v, erwartet einen Fehler des Accept", f.name, err)
		}
	}
}

// TestServeFehlerKlasse hält S2 (SPEC-136 Nr. 12): nur ein Fehler, den das Schließen durch Anhalten oder Schliessen
// selbst auslöste, ist nil; jeder andere Fehler des Accept ist der Rückgabewert, wie spät Anhalten oder Schliessen auch
// folgt. Drei Teilmengen, 35 Elemente:
//
//	A: Fehler ohne Bezug zum Schließen (3 Arten) × Zeitpunkt von Anhalten (5) = 15, Fehler
//	B: Fehler durch äußeres Schließen (2 Arten von net.ErrClosed) × Zeitpunkt (5) = 10, Fehler
//	C: Anhalten oder Schliessen zuerst (2) × Fehler des Accept danach (2 Arten net.ErrClosed → nil,
//	   3 andere Arten → Fehler) = 10
func TestServeFehlerKlasse(t *testing.T) {
	neu := func(t *testing.T) (*admin.Server, *kippListener) {
		l := neuKippListener()
		return neuMitListener(t, l), l
	}
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
		los  func(t *testing.T, srv *admin.Server, l *kippListener, ausloeser func()) error
	}{
		{"kein Anhalten", func(t *testing.T, srv *admin.Server, l *kippListener, ausloeser func()) error {
			c := startServe(srv)
			<-l.acceptiert
			ausloeser()
			return serveEnde(t, c, "kein Anhalten")
		}},
		{"Serve kehrt zurück, dann Anhalten", func(t *testing.T, srv *admin.Server, l *kippListener, ausloeser func()) error {
			c := startServe(srv)
			<-l.acceptiert
			ausloeser()
			err := serveEnde(t, c, "Serve vor Anhalten")
			srv.Anhalten()
			return err
		}},
		{"Anhalten im Accept vor der Rückkehr", func(t *testing.T, srv *admin.Server, l *kippListener, ausloeser func()) error {
			l.nachFehler = srv.Anhalten
			c := startServe(srv)
			<-l.acceptiert
			ausloeser()
			return serveEnde(t, c, "Anhalten im Accept")
		}},
		{"Schliessen gleichzeitig am Haltepunkt", func(t *testing.T, srv *admin.Server, l *kippListener, ausloeser func()) error {
			// Schliessen läuft in einer eigenen Goroutine und hat den Server angehalten, ehe der Haltepunkt
			// freigegeben wird.
			erreicht, weiter := make(chan struct{}), make(chan struct{})
			l.nachFehler = func() { close(erreicht); <-weiter }
			c := startServe(srv)
			<-l.acceptiert
			ausloeser()
			<-erreicht
			ende := make(chan struct{})
			go func() { _ = srv.Schliessen(0); close(ende) }()
			time.Sleep(100 * time.Millisecond)
			close(weiter)
			<-ende
			return serveEnde(t, c, "Schliessen am Haltepunkt")
		}},
		{"Anhalten gleichzeitig am Haltepunkt", func(t *testing.T, srv *admin.Server, l *kippListener, ausloeser func()) error {
			erreicht, weiter := make(chan struct{}), make(chan struct{})
			l.nachFehler = func() { close(erreicht); <-weiter }
			c := startServe(srv)
			<-l.acceptiert
			ausloeser()
			<-erreicht
			srv.Anhalten()
			close(weiter)
			return serveEnde(t, c, "Anhalten am Haltepunkt")
		}},
	}
	gleichAuf(t, "Zahl der Zeitpunkte", len(zeitpunkte), 5)
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
	schliesser := map[string]func(*admin.Server){
		"Anhalten":     func(s *admin.Server) { s.Anhalten() },
		"Schliessen 0": func(s *admin.Server) { _ = s.Schliessen(0) },
	}
	for schName, schliessen := range schliesser {
		for artName, art := range geschlossen {
			n++
			srv, l := neu(t)
			l.beiClose = art
			c := startServe(srv)
			<-l.acceptiert
			schliessen(srv)
			if err := serveEnde(t, c, schName); err != nil {
				t.Errorf("C %s, %s: Serve lieferte %v, erwartet nil", schName, artName, err)
			}
		}
		for artName, art := range andere {
			n++
			srv, l := neu(t)
			l.beiClose = art
			c := startServe(srv)
			<-l.acceptiert
			schliessen(srv)
			if err := serveEnde(t, c, schName); !istFehler(err, art) {
				t.Errorf("C %s, %s: Serve lieferte %v, erwartet den Fehler des Accept", schName, artName, err)
			}
		}
	}
	gleichAuf(t, "Zahl der Elemente der Klasse", n, 35)
}

// TestServeFehlerVorAnhalten hält S2 und das Fenster W6: ein Fehler vor Anhalten bleibt der Fehler, auch wenn Anhalten
// vor der Rückkehr von Accept läuft; schloss Anhalten den Listener zuerst, ist die Rückgabe nil. 4 Elemente.
func TestServeFehlerVorAnhalten(t *testing.T) {
	neu := func(t *testing.T) (*admin.Server, net.Listener, *meldeListener, *haltListener) {
		raw, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		ml := neuMeldeListener(raw)
		hl := &haltListener{Listener: ml}
		return neuMitListener(t, hl), raw, ml, hl
	}
	faelle := []struct {
		name   string
		los    func(t *testing.T, srv *admin.Server, raw net.Listener, ml *meldeListener, hl *haltListener) error
		istNil bool
	}{
		{"Fehler, Serve kehrt zurück, dann Anhalten", func(t *testing.T, srv *admin.Server, raw net.Listener, ml *meldeListener, _ *haltListener) error {
			c := startServe(srv)
			<-ml.acceptiert
			_ = raw.Close()
			err := serveEnde(t, c, "Fehler")
			srv.Anhalten()
			return err
		}, false},
		{"Fehler, Anhalten im Accept vor der Rückkehr", func(t *testing.T, srv *admin.Server, raw net.Listener, ml *meldeListener, hl *haltListener) error {
			hl.nachFehler = srv.Anhalten
			c := startServe(srv)
			<-ml.acceptiert
			_ = raw.Close()
			return serveEnde(t, c, "Anhalten im Accept")
		}, false},
		{"Fehler, Anhalten gleichzeitig am Haltepunkt", func(t *testing.T, srv *admin.Server, raw net.Listener, ml *meldeListener, hl *haltListener) error {
			erreicht, weiter := make(chan struct{}), make(chan struct{})
			hl.nachFehler = func() { close(erreicht); <-weiter }
			c := startServe(srv)
			<-ml.acceptiert
			_ = raw.Close()
			<-erreicht
			srv.Anhalten()
			close(weiter)
			return serveEnde(t, c, "Anhalten am Haltepunkt")
		}, false},
		{"Anhalten schließt zuerst, dann scheitert Accept", func(t *testing.T, srv *admin.Server, _ net.Listener, ml *meldeListener, _ *haltListener) error {
			c := startServe(srv)
			<-ml.acceptiert
			srv.Anhalten()
			return serveEnde(t, c, "Anhalten")
		}, true},
	}
	gleichAuf(t, "Zahl der Reihenfolgen", len(faelle), 4)
	for _, f := range faelle {
		srv, raw, ml, hl := neu(t)
		err := f.los(t, srv, raw, ml, hl)
		switch {
		case f.istNil && err != nil:
			t.Errorf("%s: Serve lieferte %v, erwartet nil", f.name, err)
		case !f.istNil && (err == nil || errors.Is(err, http.ErrServerClosed)):
			t.Errorf("%s: Serve lieferte %v, erwartet den Fehler des Accept", f.name, err)
		}
	}
}

// TestServeCloseNilNachAeusserem hält S2: liefert Close nach äußerem Schließen nil statt net.ErrClosed, lässt sich das
// äußere Schließen vom eigenen nicht unterscheiden, und die Rückgabe ist nil (Anhalten läuft im Accept, vor dessen Rückkehr).
func TestServeCloseNilNachAeusserem(t *testing.T) {
	l := neuKippListener()
	l.closeNil = true
	srv := neuMitListener(t, l)
	erreicht, weiter := make(chan struct{}), make(chan struct{})
	l.nachFehler = func() { close(erreicht); <-weiter }
	c := startServe(srv)
	<-l.acceptiert
	_ = l.Close()
	<-erreicht
	srv.Anhalten()
	close(weiter)
	if err := serveEnde(t, c, "Close mit nil"); err != nil {
		t.Errorf("Serve lieferte %v, erwartet nil", err)
	}
}
