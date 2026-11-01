package httpodata_test

import (
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// L7, Reihenfolgen von Fehler des Accept und Anhalten (SPEC-136 Nr. 12): ein Fehler vor Anhalten
// bleibt der Fehler, auch wenn Anhalten vor der Rückkehr von Accept läuft; schloss Anhalten den
// Listener zuerst, ist die Rückgabe nil. 4 Elemente.
func TestServeFehlerVorAnhalten(t *testing.T) {
	neu := func(t *testing.T) (*httpodata.Server, net.Listener, *meldeListener, *haltListener) {
		raw, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		ml := neuMeldeListener(raw)
		hl := &haltListener{Listener: ml}
		srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: hl})
		if err != nil {
			t.Fatal(err)
		}
		return srv, raw, ml, hl
	}
	start := func(srv *httpodata.Server) chan error {
		c := make(chan error, 1)
		go func() { c <- srv.Serve() }()
		return c
	}
	faelle := []struct {
		name   string
		los    func(t *testing.T, srv *httpodata.Server, raw net.Listener, ml *meldeListener, hl *haltListener) error
		istNil bool
	}{
		{"Fehler, Serve kehrt zurück, dann Anhalten", func(t *testing.T, srv *httpodata.Server, raw net.Listener, ml *meldeListener, _ *haltListener) error {
			c := start(srv)
			<-ml.acceptiert
			_ = raw.Close()
			err := serveEnde(t, c, "Fehler")
			srv.Anhalten()
			return err
		}, false},
		{"Fehler, Anhalten im Accept vor der Rückkehr", func(t *testing.T, srv *httpodata.Server, raw net.Listener, ml *meldeListener, hl *haltListener) error {
			hl.nachFehler = srv.Anhalten
			c := start(srv)
			<-ml.acceptiert
			_ = raw.Close()
			return serveEnde(t, c, "Anhalten im Accept")
		}, false},
		{"Fehler, Anhalten gleichzeitig am Haltepunkt", func(t *testing.T, srv *httpodata.Server, raw net.Listener, ml *meldeListener, hl *haltListener) error {
			erreicht, weiter := make(chan struct{}), make(chan struct{})
			hl.nachFehler = func() { close(erreicht); <-weiter }
			c := start(srv)
			<-ml.acceptiert
			_ = raw.Close()
			<-erreicht
			srv.Anhalten()
			close(weiter)
			return serveEnde(t, c, "Anhalten am Haltepunkt")
		}, false},
		{"Anhalten schließt zuerst, dann scheitert Accept", func(t *testing.T, srv *httpodata.Server, _ net.Listener, ml *meldeListener, _ *haltListener) error {
			c := start(srv)
			<-ml.acceptiert
			srv.Anhalten()
			return serveEnde(t, c, "Anhalten")
		}, true},
	}
	gleich(t, "Zahl der Reihenfolgen", len(faelle), 4)
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

// L13, Verbindungszustände (SPEC-136 Nr. 14): Schliessen schließt jede Verbindung ohne laufenden
// Request sofort (im Leerlauf, neu ohne ein Byte, mit halbem Kopf) und lässt die mit einem
// laufenden Request offen (halber Body, hängender Handler), bis dieser antwortet. 5 Elemente.
func TestVerbindungszustaende(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetFrei, []byte("ok"))
	h.nurPfad = "/haengt"
	leerlauf := u.verbinde(t)
	leerlauf.erledige(t)
	h.warteBegonnen(t, 1)
	neu := u.verbinde(t)
	kopf := u.verbinde(t)
	kopf.schreibe(t, []byte("GET /p HTTP/1.1\r\nHo"))
	halbBody := u.verbinde(t)
	halbBody.schreibe(t, []byte("POST /haengt HTTP/1.1\r\nHost: h\r\nContent-Length: 100\r\n\r\nabc"))
	h.warteBegonnen(t, 1)
	laeuft := u.verbinde(t)
	laeuft.schreibe(t, []byte("GET /haengt HTTP/1.1\r\nHost: h\r\n\r\n"))
	h.warteBegonnen(t, 1)
	// Der Server hat die Verbindungen angenommen und vermerkt, ehe Schliessen sie ansieht.
	time.Sleep(100 * time.Millisecond)

	schluss := make(chan error, 1)
	go func() { schluss <- u.srv.Schliessen(30 * time.Second) }()
	for name, v := range map[string]*verbindung{"Leerlauf": leerlauf, "neu ohne Byte": neu, "halber Kopf": kopf} {
		if n, err := v.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
			t.Errorf("%s: %d Bytes, Fehler %v; erwartet EOF", name, n, err)
		}
	}
	for name, v := range map[string]*verbindung{"halber Body": halbBody, "laufender Request": laeuft} {
		if n, err := v.ende(300 * time.Millisecond); n != 0 || !istZeitgrenze(err) {
			t.Errorf("%s: %d Bytes, Fehler %v; erwartet offen (Zeitgrenze)", name, n, err)
		}
	}
	h.gibFrei()
	if resp, _ := laeuft.antwort(t, "GET"); resp.StatusCode != 200 {
		t.Errorf("laufender Request: Status %d, erwartet 200", resp.StatusCode)
	}
	select {
	case err := <-schluss:
		if err != nil {
			t.Errorf("Schliessen(30 s) = %v, erwartet nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Schliessen kehrte nach der Antwort nicht zurück")
	}
}
