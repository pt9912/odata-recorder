package httpodata_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// L1: die Voreinstellungen des Servers, fünf Felder mit je einem Wert (SPEC-136 Nr. 11).
func TestServerVoreinstellungen(t *testing.T) {
	u := neuUmgebung(t, "record", nil)
	kopf, leerlauf, groesse, lesen, schreiben := u.srv.GrenzenFuerTest()
	gleich(t, "ReadHeaderTimeout", kopf, 10*time.Second)
	gleich(t, "IdleTimeout", leerlauf, 120*time.Second)
	gleich(t, "MaxHeaderBytes", groesse, 65536)
	gleich(t, "ReadTimeout", lesen, time.Duration(0))
	gleich(t, "WriteTimeout", schreiben, time.Duration(0))
}

// L2: ein Kopf weit unter der Grenze geht durch, einer weit darüber ergibt 431 von net/http.
func TestKopfGrenze(t *testing.T) {
	faelle := []struct {
		name     string
		groesse  int
		status   string
		erreicht int
	}{{"60 KiB", 60 << 10, " 200 ", 1}, {"100 KiB", 100 << 10, " 431 ", 0}}
	for _, f := range faelle {
		u := neuUmgebung(t, "record", nil)
		zeilen, _ := teile(u.roh(t, zeile("GET", "/p", "X-Gross: "+strings.Repeat("a", f.groesse))))
		if !strings.Contains(zeilen[0], f.status) {
			t.Errorf("%s: Statuszeile %q, erwartet %q", f.name, zeilen[0], f.status)
		}
		_, _, _, gesamt := u.d.zaehler()
		gleich(t, f.name+": Aufrufe von Verarbeiten", gesamt, f.erreicht)
		gleich(t, f.name+": Aufrufe von KorrelationsID", int(u.ids.zahlen()[2]), f.erreicht)
		if f.erreicht == 1 {
			if got := u.d.alle()[0].a.Header["X-Gross"]; len(got) != 1 || len(got[0]) != f.groesse {
				t.Errorf("%s: der Dienst sah den Kopf nicht unverändert", f.name)
			}
		}
	}
}

// L3: Anhalten schließt den Listener und kehrt ohne Warten zurück; der laufende Request bleibt offen.
func TestAnhaltenSchliesstListener(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetFrei, []byte("ok"))
	r := u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	if !innerhalb(2*time.Second, u.srv.Anhalten) {
		t.Fatal("Anhalten kehrte in 2 s nicht zurück")
	}
	if !dialScheitert(u.addr) {
		t.Error("nach Anhalten gelingt ein Dial")
	}
	select {
	case x := <-r:
		t.Fatalf("der laufende Request endete mit Anhalten: %+v", x)
	default:
	}
	h.gibFrei()
	if x := erwarte(t, r, "Antwort des laufenden Requests"); x.err != nil || x.status != 200 || string(x.body) != "ok" {
		t.Errorf("laufender Request: %+v", x)
	}
}

// formen sind die Request-Formen von L4, je Element eine eigene Verbindung.
type form struct {
	name, methode string
	kopf          string
	body          []byte
	weiter        bool // der Client sendet den Body in einer Goroutine weiter
	client        bool // der Client schließt selbst, der Body ist nicht fertig gesendet
}

// formen ist die Klasse „jede Form des Requests“ (SPEC-136 Nr. 13b): 10 Methoden (darunter `get`
// und ein Token ohne Bedeutung) mal 4 Body-Formen (keiner, Länge 0, 1 KiB, chunked) in der
// Origin-Form, dazu je ein Element für die Formen, die der Server anders liest: Body größer als der
// Rest, den net/http verwirft (1 MiB, weitergesendet), angekündigter und nicht gesendeter Body,
// Expect, `OPTIONS *`, `CONNECT`, absolute-form (GET, POST) und eine Länge, die nie ankommt.
func formen() []form {
	m := muster(1<<20, 3)
	var out []form
	methoden := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "TRACE", "PROPFIND", "get"}
	bodies := []struct {
		name, kopf string
		body       []byte
	}{
		{"ohne Body", "", nil},
		{"Content-Length 0", "Content-Length: 0\r\n", nil},
		{"Content-Length 1 KiB", "Content-Length: 1024\r\n", m[:1024]},
		{"chunked", "Transfer-Encoding: chunked\r\n", []byte("5\r\nhello\r\n0\r\n\r\n")},
	}
	for _, me := range methoden {
		for _, b := range bodies {
			out = append(out, form{name: me + " " + b.name, methode: me, kopf: me + " /p HTTP/1.1\r\nHost: h\r\n" + b.kopf + "\r\n", body: b.body})
		}
	}
	return append(out,
		form{name: "POST 1 MiB weitergesendet", methode: "POST", kopf: "POST /p HTTP/1.1\r\nHost: h\r\nContent-Length: 1048576\r\n\r\n", body: m, weiter: true},
		form{name: "POST angekündigt 1 MiB, 10 gesendet", methode: "POST", kopf: "POST /p HTTP/1.1\r\nHost: h\r\nContent-Length: 1048576\r\n\r\n", body: m[:10], client: true},
		form{name: "POST Expect", methode: "POST", kopf: "POST /p HTTP/1.1\r\nHost: h\r\nExpect: 100-continue\r\nContent-Length: 1024\r\n\r\n", client: true},
		form{name: "OPTIONS *", methode: "OPTIONS", kopf: "OPTIONS * HTTP/1.1\r\nHost: h\r\n\r\n"},
		form{name: "CONNECT", methode: "CONNECT", kopf: "CONNECT h:443 HTTP/1.1\r\nHost: h:443\r\n\r\n"},
		form{name: "GET absolute-form", methode: "GET", kopf: "GET http://h/p?q=1 HTTP/1.1\r\nHost: h\r\n\r\n"},
		form{name: "POST absolute-form", methode: "POST", kopf: "POST http://h/p HTTP/1.1\r\nHost: h\r\nContent-Length: 0\r\n\r\n"},
	)
}

// L4: nach Anhalten ergibt jeder Request auf einer bestehenden Verbindung 503 ODR-E5010 mit
// Connection: close, ohne Verarbeiten und ohne den Body zu lesen.
func TestAngehaltenDreiNullDrei(t *testing.T) {
	u := neuUmgebung(t, "record", nil)
	fs := formen()
	gleich(t, "Zahl der Formen", len(fs), 47)
	vs := make([]*verbindung, len(fs))
	for i := range fs {
		vs[i] = u.verbinde(t)
		vs[i].erledige(t)
	}
	u.srv.Anhalten()
	soll := fehlerobjekt(t, model.CodeE5010)
	for i, f := range fs {
		v := vs[i]
		v.schreibe(t, []byte(f.kopf))
		if f.weiter {
			go func() { _, _ = v.c.Write(f.body) }()
		} else {
			v.schreibe(t, f.body)
		}
		resp, body := v.antwort(t, f.methode)
		if resp.StatusCode != 503 || !resp.Close || resp.Header.Get("X-Correlation-Id") == "" {
			t.Errorf("%s: Status %d, Close %v, ID %q; erwartet 503, Close und eine ID", f.name, resp.StatusCode, resp.Close, resp.Header.Get("X-Correlation-Id"))
		}
		if f.methode == "HEAD" {
			if len(body) != 0 || resp.Header.Get("Content-Length") != fmt.Sprint(len(soll)) {
				t.Errorf("%s: Body %q und Länge %s; erwartet leer und %d", f.name, body, resp.Header.Get("Content-Length"), len(soll))
			}
		} else if !bytes.Equal(body, soll) {
			t.Errorf("%s: Body %q, erwartet %q", f.name, body, soll)
		}
		if _, _, _, gesamt := u.d.zaehler(); gesamt != len(fs) {
			t.Errorf("%s: Verarbeiten wurde gerufen (%d Aufrufe, erwartet %d)", f.name, gesamt, len(fs))
		}
		gleich(t, f.name+": Aufrufe von KorrelationsID", int(u.ids.zahlen()[2]), len(fs)+i+1)
	}
	for i, f := range fs {
		if f.client {
			continue
		}
		if n, err := vs[i].ende(3 * time.Second); n != 0 || err == nil || istZeitgrenze(err) {
			t.Errorf("%s: nach der Antwort %d Bytes und Fehler %v; erwartet kein Byte und eine geschlossene Verbindung", f.name, n, err)
		}
	}
}

// L5: ein vor Anhalten zugelassener Request läuft zu Ende, vollständig, mit Close einmal.
func TestLaufenderRequestEndet(t *testing.T) {
	body := muster(1<<20, 5)
	u, h := neuHaengUmgebung(t, wartetFrei, body)
	r := u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	u.srv.Anhalten()
	h.gibFrei()
	x := erwarte(t, r, "Antwort des laufenden Requests")
	if x.err != nil || x.status != 200 || !bytes.Equal(x.body, body) {
		t.Errorf("laufender Request: Status %d, %d Bytes, Fehler %v", x.status, len(x.body), x.err)
	}
	warte(t, "Close", func() bool { return h.closes.Load() == 1 })
	gleich(t, "Fehler von ctx: Anhalten beendet ihn nicht", h.ctxFehlerListe(), []error{nil})
}

// L6: ein weiteres und ein gleichzeitiges Anhalten sind wirkungslos; nach der Rückkehr jedes
// Aufrufs ist der Listener zu.
func TestAnhaltenWiederholt(t *testing.T) {
	h := neuHaenger(t, wartetFrei, []byte("ok"))
	u := neuUmgebungMit(t, "record", h.tu, nil, func(l net.Listener) net.Listener {
		return &langsamListener{Listener: l, verz: 30 * time.Millisecond}
	})
	r := u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	const aufrufer = 8
	s := neueSchranke(aufrufer)
	var wg sync.WaitGroup
	offen := make([]bool, aufrufer)
	for i := range aufrufer {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.warte()
			u.srv.Anhalten()
			offen[i] = !dialScheitert(u.addr)
		}()
	}
	wg.Wait()
	for i, o := range offen {
		if o {
			t.Errorf("Aufrufer %d: nach der Rückkehr von Anhalten gelingt ein Dial", i)
		}
	}
	u.srv.Anhalten()
	if !dialScheitert(u.addr) {
		t.Error("nach dem weiteren Anhalten gelingt ein Dial")
	}
	// Ohne Serve ruft nur Anhalten Close des Listeners (Serve schließt ihn beim Beenden selbst):
	// 8 gleichzeitige und ein weiterer Aufruf ergeben genau einen Close.
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	zl := &langsamListener{Listener: raw, verz: 10 * time.Millisecond}
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: zl})
	if err != nil {
		t.Fatal(err)
	}
	var wg2 sync.WaitGroup
	for range aufrufer {
		wg2.Add(1)
		go func() { defer wg2.Done(); srv.Anhalten() }()
	}
	wg2.Wait()
	srv.Anhalten()
	gleich(t, "Aufrufe von Close des Listeners nach 9 Aufrufen von Anhalten", zl.closes.Load(), int64(1))
	h.gibFrei()
	if x := erwarte(t, r, "Antwort des laufenden Requests"); x.err != nil || x.status != 200 {
		t.Errorf("laufender Request: %+v", x)
	}
}

// serveEnde wartet auf die Rückkehr von Serve.
func serveEnde(t *testing.T, c <-chan error, was string) error {
	t.Helper()
	select {
	case err := <-c:
		return err
	case <-time.After(5 * time.Second):
		t.Fatalf("Frist von 5 s abgelaufen: Serve kehrte nicht zurück (%s)", was)
		return nil
	}
}

// L7: die Enden von Serve, sechs Elemente: nach Anhalten oder Schliessen, auch davor, nil; ein
// von außen geschlossener oder geschlossen übergebener Listener ist ein Fehler.
func TestServeEndetMitNil(t *testing.T) {
	neu := func(t *testing.T) (*httpodata.Server, net.Listener, *meldeListener) {
		raw, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		lis := neuMeldeListener(raw)
		srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: lis})
		if err != nil {
			t.Fatal(err)
		}
		return srv, raw, lis
	}
	start := func(srv *httpodata.Server) chan error {
		c := make(chan error, 1)
		go func() { c <- srv.Serve() }()
		return c
	}
	faelle := []struct {
		name   string
		los    func(srv *httpodata.Server, raw net.Listener, lis *meldeListener) chan error
		istNil bool
	}{
		{"Serve läuft, dann Anhalten", func(srv *httpodata.Server, _ net.Listener, lis *meldeListener) chan error {
			c := start(srv)
			<-lis.acceptiert
			srv.Anhalten()
			return c
		}, true},
		{"Serve läuft, dann Schliessen(0)", func(srv *httpodata.Server, _ net.Listener, lis *meldeListener) chan error {
			c := start(srv)
			<-lis.acceptiert
			_, _ = schliesse(t, srv, 0)
			return c
		}, true},
		{"Anhalten vor Serve", func(srv *httpodata.Server, _ net.Listener, _ *meldeListener) chan error {
			srv.Anhalten()
			return start(srv)
		}, true},
		{"Schliessen(0) vor Serve", func(srv *httpodata.Server, _ net.Listener, _ *meldeListener) chan error {
			_, _ = schliesse(t, srv, 0)
			return start(srv)
		}, true},
		{"Listener von außen geschlossen", func(srv *httpodata.Server, raw net.Listener, lis *meldeListener) chan error {
			c := start(srv)
			<-lis.acceptiert
			_ = raw.Close()
			return c
		}, false},
		{"Listener geschlossen übergeben", func(srv *httpodata.Server, raw net.Listener, _ *meldeListener) chan error {
			_ = raw.Close()
			return start(srv)
		}, false},
	}
	gleich(t, "Zahl der Enden", len(faelle), 6)
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

// L8: eine negative Frist ergibt ErrUngueltig und ändert nichts; 0 ist gültig.
func TestSchliessenNegativ(t *testing.T) {
	u := neuUmgebung(t, "record", nil)
	fristen := []time.Duration{-1, time.Duration(math.MinInt64)}
	for _, f := range fristen {
		if err := u.srv.Schliessen(f); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Schliessen(%d) = %v, erwartet ErrUngueltig", int64(f), err)
		}
	}
	if resp, _ := hole(t, u, "GET", "/p"); resp.StatusCode != 200 {
		t.Errorf("nach den negativen Fristen: Status %d, erwartet 200", resp.StatusCode)
	}
	select {
	case err := <-u.serve:
		t.Errorf("Serve endete nach einer negativen Frist: %v", err)
	default:
	}
	if err := u.srv.Schliessen(0); err != nil {
		t.Errorf("Schliessen(0) = %v, erwartet nil", err)
	}
}

// L9: Schliessen kehrt früher zurück, wenn alle Requests fertig sind.
func TestSchliessenFrueh(t *testing.T) {
	body := muster(1<<20, 9)
	u, h := neuHaengUmgebung(t, wartetFrei, body)
	r := u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	time.AfterFunc(100*time.Millisecond, h.gibFrei)
	start := time.Now()
	err := u.srv.Schliessen(30 * time.Second)
	if dauer := time.Since(start); err != nil || dauer < 90*time.Millisecond || dauer >= 5*time.Second {
		t.Errorf("Schliessen(30 s) = %v nach %v, erwartet nil in [90 ms, 5 s)", err, dauer)
	}
	gleich(t, "Close bei der Rückkehr: Schliessen wartete auf den Handler", h.closes.Load(), int64(1))
	if x := erwarte(t, r, "Antwort"); x.err != nil || x.status != 200 || !bytes.Equal(x.body, body) {
		t.Errorf("Antwort: Status %d, %d Bytes, Fehler %v", x.status, len(x.body), x.err)
	}
}

// schliesse ruft Schliessen in einer Goroutine und wartet höchstens 5 s auf die Rückgabe.
func schliesse(t *testing.T, srv *httpodata.Server, frist time.Duration) (time.Duration, error) {
	t.Helper()
	c := make(chan error, 1)
	start := time.Now()
	go func() { c <- srv.Schliessen(frist) }()
	select {
	case err := <-c:
		return time.Since(start), err
	case <-time.After(5 * time.Second):
		t.Fatalf("Frist von 5 s abgelaufen: Schliessen(%v) kehrte nicht zurück", frist)
		return 0, nil
	}
}

// L10: läuft die Frist ab, endet ctx, Schliessen wartet auf Close und liefert ErrFristAbgelaufen.
func TestSchliessenFristAbgelaufen(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetCtx, []byte("ok"))
	h.closeVerz = 150 * time.Millisecond
	_ = u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	_, err := schliesse(t, u.srv, 200*time.Millisecond)
	if !errors.Is(err, httpodata.ErrFristAbgelaufen) || errors.Is(err, model.ErrUngueltig) {
		t.Errorf("Schliessen(200 ms) = %v, erwartet ErrFristAbgelaufen und nicht ErrUngueltig", err)
	}
	gleich(t, "Close bei der Rückkehr", h.closes.Load(), int64(1))
	gleich(t, "Fehler von ctx", h.ctxFehlerListe(), []error{context.Canceled})
}

// L11: Schliessen(0) wartet nicht; jede Frist ab 0 ohne laufenden Request kehrt sofort mit nil zurück.
func TestSchliessenNull(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetCtx, []byte("ok"))
	_ = u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	_, err := schliesse(t, u.srv, 0)
	if !errors.Is(err, httpodata.ErrFristAbgelaufen) {
		t.Errorf("Schliessen(0) mit laufendem Request = %v, erwartet ErrFristAbgelaufen", err)
	}
	gleich(t, "Close nach Schliessen(0)", h.closes.Load(), int64(1))

	fristen := []time.Duration{0, 1, 30 * time.Second, math.MaxInt64}
	gleich(t, "Zahl der Fristen", len(fristen), 4)
	for _, f := range fristen {
		v := neuUmgebung(t, "record", nil)
		if dauer, err := schliesse(t, v.srv, f); err != nil || dauer >= 5*time.Second {
			t.Errorf("Schliessen(%d) ohne Request = %v nach %v, erwartet nil in unter 5 s", int64(f), err, dauer)
		}
	}

	// Frist 0 und 1 ns ohne Request, 64 frische Server ohne Serve: sind der leere Zähler und der
	// abgelaufene Timer zugleich bereit, darf die Wahl nicht dem Zufall gehören.
	for i := range 64 {
		raw, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = raw.Close() })
		srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: raw})
		if err != nil {
			t.Fatal(err)
		}
		if err := srv.Schliessen(time.Duration(i % 2)); err != nil {
			t.Fatalf("Schliessen(%d ns) ohne Request im Lauf %d = %v, erwartet nil", i%2, i, err)
		}
	}

	// Die größte Frist mit laufendem Request: sie läuft nicht ab, der Request endet von selbst.
	w, wh := neuHaengUmgebung(t, wartetFrei, []byte("ok"))
	_ = w.async("GET", "/p")
	wh.warteBegonnen(t, 1)
	time.AfterFunc(100*time.Millisecond, wh.gibFrei)
	if _, err := schliesse(t, w.srv, math.MaxInt64); err != nil {
		t.Errorf("Schliessen(MaxInt64) mit laufendem Request = %v, erwartet nil", err)
	}
	gleich(t, "Fehler von ctx", wh.ctxFehlerListe(), []error{nil})
}

// L12: die Dauer ist eine Frist, nicht die Summe: Anhalten plus Schliessen(300 ms).
func TestSummeDerFristen(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetCtx, []byte("ok"))
	_ = u.async("GET", "/p")
	h.warteBegonnen(t, 1)
	start := time.Now()
	u.srv.Anhalten()
	_, err := schliesse(t, u.srv, 300*time.Millisecond)
	dauer := time.Since(start)
	if !errors.Is(err, httpodata.ErrFristAbgelaufen) || dauer < 300*time.Millisecond || dauer >= 550*time.Millisecond {
		t.Errorf("Anhalten und Schliessen(300 ms) = %v nach %v, erwartet ErrFristAbgelaufen in [300 ms, 550 ms)", err, dauer)
	}

	// Die Frist läuft ab dem Aufruf von Schliessen, nicht ab Anhalten: 300 ms nach Anhalten dauert
	// Schliessen(300 ms) noch 300 ms.
	v, vh := neuHaengUmgebung(t, wartetCtx, []byte("ok"))
	_ = v.async("GET", "/p")
	vh.warteBegonnen(t, 1)
	v.srv.Anhalten()
	time.Sleep(300 * time.Millisecond)
	spaet, err := schliesse(t, v.srv, 300*time.Millisecond)
	if !errors.Is(err, httpodata.ErrFristAbgelaufen) || spaet < 300*time.Millisecond || spaet >= 550*time.Millisecond {
		t.Errorf("Schliessen(300 ms) 300 ms nach Anhalten = %v nach %v, erwartet ErrFristAbgelaufen in [300 ms, 550 ms)", err, spaet)
	}
}

// L13: Schliessen schließt Verbindungen ohne laufenden Request sofort, Anhalten lässt sie offen;
// ein abgewiesener Request zählt nicht als laufend.
func TestIdleVerbindung(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetFrei, []byte("ok"))
	h.nurPfad = "/haengt"
	v1, v2 := u.verbinde(t), u.verbinde(t)
	v1.erledige(t)
	v2.erledige(t)
	h.warteBegonnen(t, 2)
	laeuft := u.async("GET", "/haengt")
	h.warteBegonnen(t, 1)
	u.srv.Anhalten()
	if n, err := v1.ende(200 * time.Millisecond); n != 0 || !istZeitgrenze(err) {
		t.Errorf("V1 nach Anhalten: %d Bytes, Fehler %v; erwartet offen (Zeitgrenze)", n, err)
	}
	v2.schreibe(t, []byte("GET /p HTTP/1.1\r\nHost: h\r\n\r\n"))
	if resp, _ := v2.antwort(t, "GET"); resp.StatusCode != 503 {
		t.Errorf("V2 nach Anhalten: Status %d, erwartet 503", resp.StatusCode)
	}
	// Der laufende Request endet nach dem abgewiesenen: dessen 503 darf Schliessen nicht aufhalten.
	h.gibFrei()
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= 5*time.Second {
		t.Errorf("Schliessen(30 s) = %v nach %v, erwartet nil in unter 5 s", err, dauer)
	}
	if x := erwarte(t, laeuft, "Antwort des laufenden Requests"); x.err != nil || x.status != 200 {
		t.Errorf("laufender Request: %+v", x)
	}
	if n, err := v1.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
		t.Errorf("V1 nach Schliessen: %d Bytes, Fehler %v; erwartet EOF", n, err)
	}
}

// L14: Schliessen ohne vorheriges Anhalten hält an.
func TestSchliessenOhneAnhalten(t *testing.T) {
	u := neuUmgebung(t, "record", nil)
	if _, err := schliesse(t, u.srv, time.Second); err != nil {
		t.Errorf("Schliessen(1 s) = %v, erwartet nil", err)
	}
	if !dialScheitert(u.addr) {
		t.Error("nach Schliessen gelingt ein Dial")
	}
	if err := serveEnde(t, u.serve, "nach Schliessen"); err != nil {
		t.Errorf("Serve lieferte %v, erwartet nil", err)
	}

	// Ohne Serve schließt allein Schliessen den Listener: kein Shutdown, der ihn verfolgte.
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: (&dienst{}).Record, IDs: supporttest.NewFestIDs(), Listener: raw})
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Schliessen(time.Second); err != nil {
		t.Errorf("Schliessen(1 s) ohne Serve = %v, erwartet nil", err)
	}
	if !dialScheitert(raw.Addr().String()) {
		t.Error("ohne Serve gelingt nach Schliessen ein Dial")
	}
}

// L15: ein weiteres Schliessen liefert nil und wartet nicht, auch während das erste wartet.
func TestSchliessenZweimal(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetFrei, []byte("ok"))
	h.nurPfad = "/haengt"
	v1 := u.verbinde(t)
	v1.erledige(t)
	h.warteBegonnen(t, 1)
	_ = u.async("GET", "/haengt")
	h.warteBegonnen(t, 1)
	erstes := make(chan error, 1)
	go func() { erstes <- u.srv.Schliessen(30 * time.Second) }()
	// V1 im Leerlauf schließt Shutdown als Erstes: dann hat das erste Schliessen die Zulassung passiert.
	if n, err := v1.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("V1: %d Bytes, Fehler %v; erwartet EOF", n, err)
	}
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= time.Second {
		t.Errorf("zweites Schliessen = %v nach %v, erwartet nil in unter 1 s", err, dauer)
	}
	h.gibFrei()
	select {
	case err := <-erstes:
		if err != nil {
			t.Errorf("erstes Schliessen = %v, erwartet nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("erstes Schliessen kehrte nach der Freigabe nicht zurück")
	}
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= time.Second {
		t.Errorf("drittes Schliessen = %v nach %v, erwartet nil in unter 1 s", err, dauer)
	}
}

// L16 Fall 0: der Haltepunkt liegt im Fenster zwischen Eintreffen und Zulassung (KorrelationsID);
// Schliessen kehrt zurück, bevor der Request zugelassen ist, danach läuft kein Handler.
func TestZulassungAtomar(t *testing.T) {
	t.Run("Fall 0: Haltepunkt an KorrelationsID", func(t *testing.T) {
		erreicht, weiter := make(chan struct{}), make(chan struct{})
		var einmal sync.Once
		u := neuUmgebungMit(t, "record", nil, func() {
			einmal.Do(func() { close(erreicht); <-weiter })
		}, nil)
		r := u.async("GET", "/p")
		select {
		case <-erreicht:
		case <-time.After(5 * time.Second):
			t.Fatal("der Request erreichte KorrelationsID nicht")
		}
		u.srv.Anhalten()
		if _, err := schliesse(t, u.srv, time.Second); err != nil {
			t.Errorf("Schliessen(1 s) = %v, erwartet nil", err)
		}
		close(weiter)
		x := erwarte(t, r, "Antwort")
		if x.status != 503 {
			t.Errorf("Status %d, erwartet 503", x.status)
		}
		if _, _, _, gesamt := u.d.zaehler(); gesamt != 0 {
			t.Errorf("Verarbeiten wurde nach Schliessen %d-mal gerufen, erwartet 0", gesamt)
		}
	})
	t.Run("Fall 1: Wirkung unter 200 gleichzeitigen Requests", func(t *testing.T) {
		const n = 200
		h := neuHaenger(t, sofort, []byte("ok"))
		h.dauer = 2 * time.Millisecond
		// Alle n Requests warten vor der Zulassung am Haltepunkt, das Tor öffnet sich unmittelbar
		// vor Anhalten: n Zulassungen laufen gleichzeitig mit dem Setzen des Flags.
		tor := make(chan struct{})
		var da atomic.Int64
		var auf sync.Once
		oeffne := func() { auf.Do(func() { close(tor) }) }
		t.Cleanup(oeffne)
		u := neuUmgebungMit(t, "record", h.tu, func() { da.Add(1); <-tor }, nil)
		res := make(chan ergebnis, n)
		for range n {
			go func() { res <- u.schleife() }()
		}
		warte(t, "alle Requests am Haltepunkt", func() bool { return da.Load() == n })
		oeffne()
		u.srv.Anhalten()
		_, err := schliesse(t, u.srv, 2*time.Second)
		aufrufe, laufend := h.aufrufe.Load(), h.laufend.Load()
		time.Sleep(200 * time.Millisecond)
		if err != nil || laufend != 0 || h.aufrufe.Load() != aufrufe {
			t.Errorf("Schliessen = %v; laufende Handler bei der Rückkehr %d; Aufrufe %d bei der Rückkehr, %d danach; erwartet nil, 0, gleich",
				err, laufend, aufrufe, h.aufrufe.Load())
		}
		for range n {
			if x := <-res; x.err == nil && x.status != 200 && x.status != 503 {
				t.Errorf("Status %d, erwartet 200, 503 oder einen Verbindungsfehler", x.status)
			}
		}
	})
	// 50 zugelassene Handler enden gleichzeitig: Zähler und Zulassung werden von vielen
	// Goroutinen geändert, ohne dass Anhalten die meisten abweist (L18).
	t.Run("Fall 2: 50 zugelassene Handler enden gleichzeitig", func(t *testing.T) {
		const n = 50
		u, h := neuHaengUmgebung(t, wartetFrei, []byte("ok"))
		res := make([]<-chan ergebnis, n)
		for i := range res {
			res[i] = u.async("GET", "/p")
		}
		h.warteBegonnen(t, n)
		h.gibFrei()
		for i, c := range res {
			if x := erwarte(t, c, "Antwort"); x.err != nil || x.status != 200 {
				t.Errorf("Request %d: %+v", i, x)
			}
		}
		u.srv.Anhalten()
		if _, err := schliesse(t, u.srv, time.Second); err != nil {
			t.Errorf("Schliessen(1 s) = %v, erwartet nil", err)
		}
		gleich(t, "laufende Handler nach Schliessen", h.laufend.Load(), int64(0))
	})
}

// L19: der Adapter schreibt kein Log von net/http; die Panik des Dienstes erscheint weder in
// log.Default() noch auf stderr. Der Test fährt das Szenario in einem Kindprozess (dieselbe
// Testbinärdatei) und liest dessen stderr und stdout, den Standardlogger unverändert.
func TestKeinLogDerBibliothek(t *testing.T) {
	if dir := os.Getenv(kindUmgebung); dir != "" {
		kindL19(t, dir)
		return
	}
	dir := t.TempDir()
	ctx, abbruch := context.WithTimeout(context.Background(), 60*time.Second)
	defer abbruch()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestKeinLogDerBibliothek$", "-test.count=1")
	cmd.Env = append(os.Environ(), kindUmgebung+"="+dir)
	var aus, fehl bytes.Buffer
	cmd.Stdout, cmd.Stderr = &aus, &fehl
	if err := cmd.Run(); err != nil {
		t.Fatalf("Kindprozess: %v\nstdout: %s\nstderr: %s", err, aus.String(), fehl.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "lief")); err != nil {
		t.Fatalf("der Kindprozess hat das Szenario nicht ausgeführt: %v", err)
	}
	for senke, text := range map[string]string{"stderr": fehl.String(), "stdout": aus.String()} {
		if strings.Contains(text, "geheim-wert-L19") {
			t.Errorf("%s trägt die Panik: %q", senke, text)
		}
	}
}

const kindUmgebung = "ODR_L19_KIND"

// kindL19 ist das Szenario im Kindprozess: der Dienst panikt, der Client sieht den Abbruch.
func kindL19(t *testing.T, dir string) {
	t.Helper()
	u := neuUmgebung(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
		panic("geheim-wert-L19")
	})
	req, err := http.NewRequest("GET", "http://"+u.addr+"/p", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp, err := u.client().Do(req); err == nil {
		_ = resp.Body.Close()
		t.Fatalf("der Client sah keinen Abbruch, Status %d", resp.StatusCode)
	}
	if err := os.WriteFile(filepath.Join(dir, "lief"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// L20: der ctx endet beim Ablauf der Frist auch bei ungelesenem Body.
func TestSchliessenKontextUngelesenerBody(t *testing.T) {
	u, h := neuHaengUmgebung(t, wartetCtx, []byte("ok"))
	v := u.verbinde(t)
	v.schreibe(t, []byte("POST /p HTTP/1.1\r\nHost: h\r\nContent-Length: 100\r\n\r\nabc"))
	h.warteBegonnen(t, 1)
	_, err := schliesse(t, u.srv, 200*time.Millisecond)
	if !errors.Is(err, httpodata.ErrFristAbgelaufen) {
		t.Errorf("Schliessen(200 ms) = %v, erwartet ErrFristAbgelaufen", err)
	}
	gleich(t, "Close", h.closes.Load(), int64(1))
	gleich(t, "Fehler von ctx", h.ctxFehlerListe(), []error{context.Canceled})
}
