package httpupstream_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// geheim steht in jedem Wert einer Anfrage und jeder Antwort, den ein Fehlertext nicht nennen darf
// (SPEC-120 Nr. 8, A14).
const geheim = "S3CR3T"

// zeitgrenze begrenzt jedes Warten eines Tests auf ein Ereignis; sie ist die Grenze des
// Fehlschlags, kein Zeitfenster.
const zeitgrenze = 10 * time.Second

var (
	errBody       = errors.New("body kaputt " + geheim + "-e")
	errVerbindung = errors.New("verbindung abgerissen " + geheim + "-v")
)

// --- Roher Listener auf 127.0.0.1 --------------------------------------------------------------

// roherServer nimmt Verbindungen an und gibt jede an die Funktion des Tests; angenommen meldet die
// Gegenadresse jeder angenommenen Verbindung in ihrer Reihenfolge.
type roherServer struct {
	ln         net.Listener
	angenommen chan string
	wg         sync.WaitGroup
	mu         sync.Mutex
	zu         bool
	conns      []net.Conn
}

func starteServer(t *testing.T, h func(c net.Conn, r *bufio.Reader)) *roherServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &roherServer{ln: ln, angenommen: make(chan string, 256)}
	s.wg.Add(1)
	go s.annehmen(h)
	t.Cleanup(func() {
		_ = ln.Close()
		s.mu.Lock()
		s.zu = true
		for _, c := range s.conns {
			_ = c.Close()
		}
		s.mu.Unlock()
		s.wg.Wait()
	})
	return s
}

func (s *roherServer) annehmen(h func(c net.Conn, r *bufio.Reader)) {
	defer s.wg.Done()
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		if s.zu {
			s.mu.Unlock()
			_ = c.Close()
			return
		}
		s.conns = append(s.conns, c)
		s.wg.Add(1)
		s.mu.Unlock()
		select {
		case s.angenommen <- c.RemoteAddr().String():
		default:
		}
		go func() {
			defer s.wg.Done()
			defer func() { _ = c.Close() }()
			h(c, bufio.NewReader(c))
		}()
	}
}

func (s *roherServer) ziel() string { return "http://" + s.ln.Addr().String() }

func (s *roherServer) port() int { return s.ln.Addr().(*net.TCPAddr).Port }

// keineVerbindung prüft, dass der Server vor einer Marken-Verbindung des Tests keine Verbindung
// annahm; Verbindungen werden in der Reihenfolge des Aufbaus angenommen.
func (s *roherServer) keineVerbindung(t *testing.T) {
	t.Helper()
	d, err := net.Dial("tcp", s.ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	marke := d.LocalAddr().String()
	grenze := time.NewTimer(zeitgrenze)
	defer grenze.Stop()
	for {
		select {
		case a := <-s.angenommen:
			if a == marke {
				return
			}
			t.Error("der Server nahm vor der Marke eine Verbindung an")
		case <-grenze.C:
			t.Fatal("Marken-Verbindung nicht angenommen")
		}
	}
}

// --- Request roh lesen ---------------------------------------------------------------------------

type kopfzeile struct{ name, wert string }

// empfangen ist ein Request, wie er auf der Leitung stand: Zeile, Kopfzeilen in Reihenfolge, Body
// nach seiner Rahmung und bei chunked die Größen der Stücke samt dem letzten.
type empfangen struct {
	zeile   string
	kopf    []kopfzeile
	body    []byte
	stuecke []int
}

// werte liefert die Werte der Kopfzeilen mit genau diesem Namen in ihrer Reihenfolge.
func (e empfangen) werte(name string) []string {
	var w []string
	for _, k := range e.kopf {
		if k.name == name {
			w = append(w, k.wert)
		}
	}
	return w
}

// zeilen liefert die Kopfzeilen als Text `Name: Wert` in ihrer Reihenfolge.
func (e empfangen) zeilen() []string {
	z := make([]string, 0, len(e.kopf))
	for _, k := range e.kopf {
		z = append(z, k.name+": "+k.wert)
	}
	return z
}

func leseZeile(r *bufio.Reader) (string, error) {
	z, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	z, ok := strings.CutSuffix(z, "\r\n")
	if !ok {
		return "", errors.New("Zeile ohne CRLF")
	}
	return z, nil
}

// leseKopf liest Request-Zeile und Kopfzeilen bis zur Leerzeile; ein Wert verliert genau ein
// Leerzeichen nach dem Doppelpunkt.
func leseKopf(r *bufio.Reader) (empfangen, error) {
	var e empfangen
	z, err := leseZeile(r)
	if err != nil {
		return e, err
	}
	e.zeile = z
	for {
		z, err := leseZeile(r)
		if err != nil {
			return e, err
		}
		if z == "" {
			return e, nil
		}
		name, wert, _ := strings.Cut(z, ":")
		e.kopf = append(e.kopf, kopfzeile{name, strings.TrimPrefix(wert, " ")})
	}
}

// leseBody liest den Body nach Transfer-Encoding chunked oder Content-Length.
func leseBody(r *bufio.Reader, e *empfangen) error {
	if te := e.werte("Transfer-Encoding"); len(te) == 1 && te[0] == "chunked" {
		return leseStuecke(r, e)
	}
	if cl := e.werte("Content-Length"); len(cl) == 1 {
		n, err := strconv.Atoi(cl[0])
		if err != nil {
			return err
		}
		e.body = make([]byte, n)
		_, err = io.ReadFull(r, e.body)
		return err
	}
	return nil
}

func leseStuecke(r *bufio.Reader, e *empfangen) error {
	for {
		z, err := leseZeile(r)
		if err != nil {
			return err
		}
		n, err := strconv.ParseInt(z, 16, 64)
		if err != nil {
			return err
		}
		e.stuecke = append(e.stuecke, int(n))
		b := make([]byte, n+2)
		if _, err := io.ReadFull(r, b); err != nil {
			return err
		}
		if string(b[n:]) != "\r\n" {
			return errors.New("Stück ohne CRLF")
		}
		if n == 0 {
			return nil
		}
		e.body = append(e.body, b[:n]...)
	}
}

func leseAnfrage(r *bufio.Reader) (empfangen, error) {
	e, err := leseKopf(r)
	if err != nil {
		return e, err
	}
	return e, leseBody(r, &e)
}

// beantworte liest Requests auf einer Verbindung, meldet jeden auf an, solange dort Platz ist, und
// schreibt je Request die nächste Antwort aus antworten (die letzte gilt für alle weiteren).
func beantworte(an chan<- empfangen, antworten ...string) func(net.Conn, *bufio.Reader) {
	return func(c net.Conn, r *bufio.Reader) {
		for i := 0; ; i++ {
			e, err := leseAnfrage(r)
			if err != nil {
				return
			}
			select {
			case an <- e:
			default:
			}
			if _, err := io.WriteString(c, antworten[min(i, len(antworten)-1)]); err != nil {
				return
			}
		}
	}
}

const antwortOK = "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"

// --- Client und Anfrage ------------------------------------------------------------------------

func neuerClient(t *testing.T, frist time.Duration) *httpupstream.Client {
	t.Helper()
	c, err := httpupstream.New(httpupstream.Konfig{Frist: frist, Pruefen: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(httpupstream.Transport(c).CloseIdleConnections)
	return c
}

func anfrage(ziel, methode, pfad string) outbound.UpstreamAnfrage {
	return outbound.UpstreamAnfrage{Ziel: ziel, Methode: methode, Pfad: pfad}
}

// mitGeheim setzt Pfad, Query und einen Header, die jeder Fehlertext verschweigen muss (A14).
func mitGeheim(a outbound.UpstreamAnfrage) outbound.UpstreamAnfrage {
	a.Pfad = "/" + geheim + "-pfad"
	a.RawQuery = "token=" + geheim + "-q"
	a.Header = map[string][]string{"X-" + geheim + "-n": {geheim + "-h"}}
	return a
}

type ergebnis struct {
	ant outbound.UpstreamAntwort
	err error
}

// sendeNebenbei ruft Senden in einer eigenen Goroutine; das Ergebnis kommt auf dem Kanal.
func sendeNebenbei(ctx context.Context, c *httpupstream.Client, a outbound.UpstreamAnfrage) <-chan ergebnis {
	ch := make(chan ergebnis, 1)
	go func() {
		ant, err := c.Senden(ctx, a)
		ch <- ergebnis{ant, err}
	}()
	return ch
}

// warte liefert den nächsten Wert von ch; nach der Zeitgrenze schlägt der Test fehl.
func warte[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	grenze := time.NewTimer(zeitgrenze)
	defer grenze.Stop()
	select {
	case v := <-ch:
		return v
	case <-grenze.C:
		t.Fatal("Zeitgrenze erreicht")
		var leer T
		return leer
	}
}

// sende ruft Senden mit der Zeitgrenze.
func sende(ctx context.Context, t *testing.T, c *httpupstream.Client, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	t.Helper()
	e := warte(t, sendeNebenbei(ctx, c, a))
	return e.ant, e.err
}

// liesAlles liest den Body bis zum Ende mit der Zeitgrenze und schließt ihn.
func liesAlles(t *testing.T, b io.ReadCloser) ([]byte, error) {
	t.Helper()
	ch := make(chan ergebnisLesen, 1)
	go func() {
		d, err := io.ReadAll(b)
		ch <- ergebnisLesen{d, err}
	}()
	e := warte(t, ch)
	_ = b.Close()
	return e.daten, e.err
}

type ergebnisLesen struct {
	daten []byte
	err   error
}

// --- Bodies der Anfrage ------------------------------------------------------------------------

// zaehlBody zählt die Aufrufe von Read.
type zaehlBody struct {
	r io.Reader
	n atomic.Int32
}

func (z *zaehlBody) Read(p []byte) (int, error) {
	z.n.Add(1)
	return z.r.Read(p)
}

// stueckBody liefert je Read das nächste Stück; ein Stück mit freigabe wartet vorher auf sie. Nach
// dem letzten Stück liefert er ende. aufrufe zählt die Aufrufe von Read.
type stueckBody struct {
	stuecke []stueckWert
	ende    error
	aufrufe atomic.Int32
}

type stueckWert struct {
	daten    string
	err      error
	freigabe <-chan struct{}
}

func (s *stueckBody) Read(p []byte) (int, error) {
	i := int(s.aufrufe.Add(1)) - 1
	if i >= len(s.stuecke) {
		return 0, s.ende
	}
	w := s.stuecke[i]
	if w.freigabe != nil {
		<-w.freigabe
	}
	return copy(p, w.daten), w.err
}

// --- Fake des Transports und Timer der Frist ---------------------------------------------------

// fakeTransport führt die Folge aus, die der Test bestimmt.
type fakeTransport func(req *http.Request) (*http.Response, error)

func (f fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// merkBody ist der Body einer Antwort des Fakes; er merkt sich Close.
type merkBody struct {
	io.Reader
	zu atomic.Bool
}

func (m *merkBody) Close() error {
	m.zu.Store(true)
	return nil
}

func fakeKopf(b *merkBody) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: b}
}

// testUhr ersetzt time.AfterFunc für die Frist: ausloesen ruft die Funktion des letzten Timers
// synchron. Wie bei time.AfterFunc bleibt Auslösen nach Stop ohne Wirkung, und Stop nach dem
// Auslösen meldet falsch.
type testUhr struct {
	mu         sync.Mutex
	f          func()
	d          time.Duration
	gestoppt   bool
	ausgeloest bool
	gestartet  chan struct{}
}

func neueTestUhr(c *httpupstream.Client) *testUhr {
	u := &testUhr{gestartet: make(chan struct{}, 1)}
	httpupstream.SetzeFrist(c, u.starte)
	return u
}

func (u *testUhr) starte(d time.Duration, f func()) func() bool {
	u.mu.Lock()
	u.f, u.d, u.gestoppt, u.ausgeloest = f, d, false, false
	u.mu.Unlock()
	select {
	case u.gestartet <- struct{}{}:
	default:
	}
	return u.stop
}

func (u *testUhr) stop() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.gestoppt || u.ausgeloest {
		return false
	}
	u.gestoppt = true
	return true
}

func (u *testUhr) ausloesen() {
	u.mu.Lock()
	if u.gestoppt || u.ausgeloest || u.f == nil {
		u.mu.Unlock()
		return
	}
	u.ausgeloest = true
	f := u.f
	u.mu.Unlock()
	f()
}

// --- Blase von testing/synctest mit net.Pipe ---------------------------------------------------

// blase führt f in einer Blase aus: der Client entsteht in ihr, sein Transport verbindet genau
// einmal über net.Pipe mit server, der in einer eigenen Goroutine der Blase läuft. Am Ende, auch
// nach t.Fatal in f, schließt sie die ruhenden Verbindungen und beide Enden der Pipe und wartet auf
// server.
func blase(t *testing.T, frist time.Duration, server func(c net.Conn), f func(t *testing.T, c *httpupstream.Client)) {
	t.Helper()
	synctest.Test(t, func(t *testing.T) {
		c, err := httpupstream.New(httpupstream.Konfig{Frist: frist, Pruefen: true})
		if err != nil {
			t.Fatal(err)
		}
		tr := httpupstream.Transport(c)
		cli, srv := net.Pipe()
		verbindungen := make(chan net.Conn, 1)
		verbindungen <- cli
		tr.DialContext = func(context.Context, string, string) (net.Conn, error) {
			select {
			case v := <-verbindungen:
				return v, nil
			default:
				return nil, errors.New("zweite Verbindung")
			}
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { _ = srv.Close() }()
			server(srv)
		}()
		defer func() {
			tr.CloseIdleConnections()
			_ = cli.Close()
			_ = srv.Close()
			wg.Wait()
		}()
		f(t, c)
	})
}

// zielBlase ist das Ziel der Anfragen in der Blase; verbunden wird über die Pipe.
const zielBlase = "http://upstream.test"

// pruefeArt prüft, dass err die erwartete Art trägt und keinen anderen Fehler des Ports.
func pruefeArt(t *testing.T, err, art error) {
	t.Helper()
	if !errors.Is(err, art) {
		t.Fatalf("Fehler %v, erwartet %v", err, art)
	}
	for _, anderer := range []error{outbound.ErrUpstreamNichtErreichbar, outbound.ErrUpstreamZeit} {
		if anderer != art && errors.Is(err, anderer) {
			t.Errorf("Fehler %v trägt zusätzlich %v", err, anderer)
		}
	}
}

func text(b []byte) string { return fmt.Sprintf("%q", b) }
