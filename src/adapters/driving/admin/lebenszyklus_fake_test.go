package admin_test

import (
	"bufio"
	"errors"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// liveOpt sind die Haltepunkte eines laufenden Servers; nil heißt: keiner.
type liveOpt struct {
	vorKorr   func()
	umhuelle  func(net.Listener) net.Listener
	vergleich func(erwartet, gegeben []byte) bool
}

// live ist ein laufender Server auf 127.0.0.1:0 mit den Routen GET und POST auf /p und /haengt, gezählten IDs und Audit;
// Serve läuft in einer Goroutine, deren Rückgabe der Kanal serve trägt.
type live struct {
	srv   *admin.Server
	addr  string
	route *routeFake
	ids   *zaehlendeIDs
	audit *supporttest.FakeAudit
	serve chan error
}

func neuLive(t *testing.T, route *routeFake) *live { return neuLiveMit(t, route, liveOpt{}) }

// neuLiveMit baut den Server; die Routen rufen alle route.handle.
func neuLiveMit(t *testing.T, route *routeFake, opt liveOpt) *live {
	t.Helper()
	roh, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = roh.Close() })
	lis := roh
	if opt.umhuelle != nil {
		lis = opt.umhuelle(roh)
	}
	k, u := konfigMit()
	k.Listener, k.Vergleich = lis, opt.vergleich
	for _, m := range []string{"GET", "POST"} {
		for _, p := range []string{"/p", "/haengt"} {
			k.Routen = append(k.Routen, admin.Route{Methode: m, Muster: p, Handler: route.handle})
		}
	}
	u.ids.vor = opt.vorKorr
	srv := neuMitKonfig(t, k, u).srv
	serve := make(chan error, 1)
	go func() { serve <- srv.Serve() }()
	return &live{srv: srv, addr: roh.Addr().String(), route: route, ids: u.ids, audit: u.audit, serve: serve}
}

// neuMitListener baut einen Server mit gültiger Konfiguration auf dem Listener l, ohne Serve zu starten.
func neuMitListener(t *testing.T, l net.Listener) *admin.Server {
	t.Helper()
	k, u := konfigMit()
	k.Listener = l
	return neuMitKonfig(t, k, u).srv
}

// kopfAuth ist die Zeile Authorization mit dem gültigen Token samt Zeilenende.
func kopfAuth() string { return "Authorization: Bearer " + string(tokenGueltig()) + "\r\n" }

// getP ist ein GET auf /p mit Token.
func getP() []byte { return []byte("GET /p HTTP/1.1\r\nHost: h\r\n" + kopfAuth() + "\r\n") }

// ergebnis ist die Antwort eines Requests, den ein Client in einer Goroutine sandte.
type ergebnis struct {
	status int
	body   []byte
	err    error
}

func (u *live) client() *http.Client {
	return &http.Client{
		Transport:     &http.Transport{DisableKeepAlives: true},
		Timeout:       10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// async sendet einen Request mit Token in einer Goroutine; der Kanal trägt genau ein Ergebnis.
func (u *live) async(methode, pfad string) <-chan ergebnis {
	c := make(chan ergebnis, 1)
	go func() {
		req, err := http.NewRequest(methode, "http://"+u.addr+pfad, nil)
		if err != nil {
			c <- ergebnis{err: err}
			return
		}
		req.Header.Set("Authorization", "Bearer "+string(tokenGueltig()))
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

// schleife sendet auf einer eigenen Verbindung GET-Requests nacheinander, bis die Antwort keine 200 ist oder die
// Verbindung endet; so treffen Requests auch nach Anhalten ein.
func (u *live) schleife() ergebnis {
	c, err := net.Dial("tcp", u.addr)
	if err != nil {
		return ergebnis{err: err}
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(10 * time.Second))
	r := bufio.NewReader(c)
	for {
		if _, err := c.Write(getP()); err != nil {
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

// serveEnde wartet auf die Rückgabe von Serve.
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

// schliesse ruft Schliessen in einer Goroutine und wartet höchstens 5 s auf die Rückgabe.
func schliesse(t *testing.T, srv *admin.Server, frist time.Duration) (time.Duration, error) {
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

// warteBis wartet höchstens 5 s darauf, dass ok wahr wird.
func warteBis(t *testing.T, was string, ok func() bool) {
	t.Helper()
	ende := time.Now().Add(5 * time.Second)
	for !ok() {
		if time.Now().After(ende) {
			t.Fatalf("Frist von 5 s abgelaufen: %s", was)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// gleichAuf prüft got gegen soll mit reflect.DeepEqual.
func gleichAuf(t *testing.T, was string, got, soll any) {
	t.Helper()
	if !reflect.DeepEqual(got, soll) {
		t.Errorf("%s = %#v, erwartet %#v", was, got, soll)
	}
}

// muster liefert n Zufallsbytes, deterministisch je Startwert.
func muster(n int, start uint64) []byte {
	r := rand.New(rand.NewPCG(start, 7))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.IntN(256))
	}
	return b
}

// schranke gibt alle frei, sobald n Teilnehmer warten (oder nach 5 s).
type schranke struct {
	mu   sync.Mutex
	n    int
	da   int
	frei chan struct{}
}

func neueSchranke(n int) *schranke { return &schranke{n: n, frei: make(chan struct{})} }

func (s *schranke) warte() {
	s.mu.Lock()
	s.da++
	if s.da == s.n {
		close(s.frei)
	}
	s.mu.Unlock()
	select {
	case <-s.frei:
	case <-time.After(5 * time.Second):
	}
}

// verbindung ist eine rohe Verbindung mit Keep-Alive, auf der der Test Requests und Antworten ordnet.
type verbindung struct {
	c net.Conn
	r *bufio.Reader
}

func (u *live) verbinde(t *testing.T) *verbindung {
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

// erledige sendet einen GET auf /p und erwartet 200; die Verbindung bleibt im Leerlauf.
func (v *verbindung) erledige(t *testing.T) {
	t.Helper()
	v.schreibe(t, getP())
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

// langsamListener wartet vor dem Schließen; ein zweiter Aufruf von Anhalten liegt dann im Fenster zwischen dem Setzen
// des Flags und dem Schließen (W2). Er zählt jeden Aufruf von Close.
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

// haltListener ruft nach einem gescheiterten Accept des inneren Listeners den Haltepunkt, ehe der Fehler zurückkehrt.
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

// kippListener ist ein Listener mit Zustand wie ein echter: das erste Close gelingt und weckt Accept, jedes weitere
// liefert net.ErrClosed (mit closeNil: nil). Accept liefert den Fehler, den ein Test einspeist, oder nach dem Schließen
// beiClose (nil: ein OpError mit net.ErrClosed); danach ruft es nachFehler, ehe der Fehler zurückkehrt.
type kippListener struct {
	zu          chan struct{}
	spontan     chan error
	acceptiert  chan struct{}
	erster      sync.Once
	mu          sync.Mutex
	geschlossen bool
	closeNil    bool
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
		if l.closeNil {
			return nil
		}
		return net.ErrClosed
	}
	l.geschlossen = true
	close(l.zu)
	return nil
}

func (l *kippListener) Addr() net.Addr { return &net.TCPAddr{} }
