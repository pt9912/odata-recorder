package httpodata_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// haengModus sagt, wie der Dienst des Fakes auf einen Aufruf wartet.
type haengModus int

const (
	sofort     haengModus = iota // kehrt nach dauer zurück
	wartetFrei                   // wartet auf gibFrei
	wartetCtx                    // wartet auf das Ende von ctx oder auf gibFrei
)

// haenger ist der Dienst-Fake der Lebenszyklus-Tests: er zählt die Aufrufe, die laufenden Handler
// (ein Handler läuft bis zum Ende von Close des Bodys), die Close-Aufrufe und hält den Fehler von
// ctx nach dem Warten fest. Der ctx liegt nie im Fake, containedctx verbietet es.
type haenger struct {
	modus     haengModus
	body      []byte
	dauer     time.Duration
	closeVerz time.Duration
	nurPfad   string // wenn gesetzt, wartet nur ein Request auf diesen Pfad, jeder andere kehrt sofort zurück
	frei      chan struct{}
	begonnen  chan struct{}
	einmal    sync.Once
	laufend   atomic.Int64
	aufrufe   atomic.Int64
	closes    atomic.Int64
	mu        sync.Mutex
	ctxFehler []error
}

func neuHaenger(t *testing.T, modus haengModus, body []byte) *haenger {
	t.Helper()
	h := &haenger{modus: modus, body: body, frei: make(chan struct{}), begonnen: make(chan struct{}, 4096)}
	t.Cleanup(h.gibFrei)
	return h
}

func (h *haenger) gibFrei() { h.einmal.Do(func() { close(h.frei) }) }

func (h *haenger) tu(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	h.aufrufe.Add(1)
	h.laufend.Add(1)
	h.begonnen <- struct{}{}
	modus := h.modus
	if h.nurPfad != "" && a.Pfad != h.nurPfad {
		modus = sofort
	}
	switch modus {
	case wartetFrei:
		<-h.frei
	case wartetCtx:
		select {
		case <-ctx.Done():
		case <-h.frei:
		}
	case sofort:
		time.Sleep(h.dauer)
	}
	h.mu.Lock()
	h.ctxFehler = append(h.ctxFehler, ctx.Err())
	h.mu.Unlock()
	k := ausBytes(h.body)
	k.beiClose = func() {
		time.Sleep(h.closeVerz)
		h.closes.Add(1)
		h.laufend.Add(-1)
	}
	return inbound.Antwort{Status: 200, Body: k}, nil
}

func (h *haenger) ctxFehlerListe() []error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]error(nil), h.ctxFehler...)
}

// warteBegonnen wartet, bis n Aufrufe im Fake begonnen haben.
func (h *haenger) warteBegonnen(t *testing.T, n int) {
	t.Helper()
	for range n {
		select {
		case <-h.begonnen:
		case <-time.After(5 * time.Second):
			t.Fatal("Frist von 5 s abgelaufen: der Request erreichte den Dienst nicht")
		}
	}
}

func neuHaengUmgebung(t *testing.T, modus haengModus, body []byte) (*umgebung, *haenger) {
	t.Helper()
	h := neuHaenger(t, modus, body)
	return neuUmgebung(t, "record", h.tu), h
}

// ergebnis ist die Antwort eines Requests, den ein Client in einer Goroutine sandte.
type ergebnis struct {
	status int
	body   []byte
	err    error
}

// async sendet einen Request in einer Goroutine; der Kanal trägt genau ein Ergebnis.
func (u *umgebung) async(methode, pfad string) <-chan ergebnis {
	c := make(chan ergebnis, 1)
	go func() {
		req, err := http.NewRequest(methode, "http://"+u.addr+pfad, nil)
		if err != nil {
			c <- ergebnis{err: err}
			return
		}
		resp, err := u.client().Do(req)
		if err != nil {
			c <- ergebnis{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		c <- ergebnis{status: resp.StatusCode, body: body, err: err}
	}()
	return c
}

// schleife sendet auf einer eigenen Verbindung GET-Requests nacheinander, bis die Antwort keine 200
// ist oder die Verbindung endet; so treffen Requests auch nach Anhalten ein. Das Ergebnis trägt die
// letzte Antwort oder den Fehler.
func (u *umgebung) schleife() ergebnis {
	c, err := net.Dial("tcp", u.addr)
	if err != nil {
		return ergebnis{err: err}
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(10 * time.Second))
	r := bufio.NewReader(c)
	for {
		if _, err := c.Write([]byte("GET /p HTTP/1.1\r\nHost: h\r\n\r\n")); err != nil {
			return ergebnis{err: err}
		}
		resp, err := http.ReadResponse(r, &http.Request{Method: "GET"})
		if err != nil {
			return ergebnis{err: err}
		}
		_, err = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode != 200 {
			return ergebnis{status: resp.StatusCode, err: err}
		}
	}
}

// erwarte wartet höchstens 5 s auf das Ergebnis.
func erwarte(t *testing.T, c <-chan ergebnis, was string) ergebnis {
	t.Helper()
	select {
	case r := <-c:
		return r
	case <-time.After(5 * time.Second):
		t.Fatalf("Frist von 5 s abgelaufen: %s", was)
		return ergebnis{}
	}
}

// dialScheitert meldet, ob ein Verbindungsaufbau scheitert (der Listener ist zu).
func dialScheitert(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return true
	}
	_ = c.Close()
	return false
}

// innerhalb ruft f in einer Goroutine und meldet, ob sie in der Frist zurückkehrt.
func innerhalb(frist time.Duration, f func()) bool {
	fertig := make(chan struct{})
	go func() { f(); close(fertig) }()
	select {
	case <-fertig:
		return true
	case <-time.After(frist):
		return false
	}
}

// verbindung ist eine rohe Verbindung mit Keep-Alive, auf der der Test Requests und Antworten ordnet.
type verbindung struct {
	c net.Conn
	r *bufio.Reader
}

func (u *umgebung) verbinde(t *testing.T) *verbindung {
	t.Helper()
	c, err := net.Dial("tcp", u.addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return &verbindung{c: c, r: bufio.NewReader(c)}
}

func (v *verbindung) schreibe(t *testing.T, b []byte) {
	t.Helper()
	_ = v.c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := v.c.Write(b); err != nil {
		t.Fatalf("Schreiben: %v", err)
	}
}

// antwort liest eine Antwort und ihren Body ganz; die Methode des Requests bestimmt die Rahmung.
func (v *verbindung) antwort(t *testing.T, methode string) (*http.Response, []byte) {
	t.Helper()
	_ = v.c.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp, err := http.ReadResponse(v.r, &http.Request{Method: methode})
	if err != nil {
		t.Fatalf("Antwort auf %s lesen: %v", methode, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Body der Antwort auf %s lesen: %v", methode, err)
	}
	return resp, body
}

// erledige sendet einen GET und erwartet 200; die Verbindung bleibt im Leerlauf.
func (v *verbindung) erledige(t *testing.T) {
	t.Helper()
	v.schreibe(t, []byte("GET /p HTTP/1.1\r\nHost: h\r\n\r\n"))
	if resp, _ := v.antwort(t, "GET"); resp.StatusCode != 200 || resp.Close {
		t.Fatalf("erster Request: Status %d, Close %v; erwartet 200 und Keep-Alive", resp.StatusCode, resp.Close)
	}
}

// ende liest höchstens frist lang und liefert die Zahl der Bytes und den Fehler des ersten Read.
func (v *verbindung) ende(frist time.Duration) (int, error) {
	_ = v.c.SetReadDeadline(time.Now().Add(frist))
	n, err := v.r.Read(make([]byte, 512))
	return n, err
}

func istZeitgrenze(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// langsamListener wartet vor dem Schließen; ein zweiter Aufruf von Anhalten liegt dann im Fenster
// zwischen dem Setzen des Flags und dem Schließen (W2).
// Es zählt jeden Aufruf von Close.
type langsamListener struct {
	net.Listener
	verz   time.Duration
	closes atomic.Int64
}

func (l *langsamListener) Close() error {
	l.closes.Add(1)
	time.Sleep(l.verz)
	return l.Listener.Close()
}

// haltListener ruft nach einem gescheiterten Accept des inneren Listeners den Haltepunkt, ehe der
// Fehler zurückkehrt.
type haltListener struct {
	net.Listener
	nachFehler func()
}

func (l *haltListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil && l.nachFehler != nil {
		l.nachFehler()
	}
	return c, err
}

// meldeListener meldet den ersten Aufruf von Accept.
type meldeListener struct {
	net.Listener
	erster     sync.Once
	acceptiert chan struct{}
}

func neuMeldeListener(l net.Listener) *meldeListener {
	return &meldeListener{Listener: l, acceptiert: make(chan struct{})}
}

func (l *meldeListener) Accept() (net.Conn, error) {
	l.erster.Do(func() { close(l.acceptiert) })
	return l.Listener.Accept()
}
