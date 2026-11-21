package httpodata_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// tlsUm ist eine Umgebung, deren Listener der Composition Root so umhüllt, wie SPEC-291 Nr. 8 es sagt: tls.NewListener mit Mindestversion TLS 1.2 und
// ALPN allein `http/1.1` (SPEC-291 Nr. 5, 6); der Aussteller und das Blatt (`localhost`) stammen aus tlstest.
type tlsUm struct {
	*umgebung
	wurzel *tlstest.Aussteller
}

func neuTLSUmgebung(t *testing.T, tu func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error), innen func(net.Listener) net.Listener) *tlsUm {
	t.Helper()
	wurzel := tlstest.NeueWurzel(t)
	blatt := wurzel.Blatt(t, tlstest.Optionen{})
	cfg := &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{blatt.Zertifikat.Raw}, PrivateKey: blatt.Schluessel}},
		MinVersion:   tls.VersionTLS12, NextProtos: []string{"http/1.1"},
	}
	u := neuUmgebungMit(t, "record", tu, nil, func(l net.Listener) net.Listener {
		if innen != nil {
			l = innen(l)
		}
		return tls.NewListener(l, cfg)
	})
	return &tlsUm{umgebung: u, wurzel: wurzel}
}

func (u *tlsUm) klient() *tls.Config { return u.wurzel.Klient("localhost") }

// httpClient bietet HTTP/2 an (`ForceAttemptHTTP2`) und hält keine Verbindung offen.
func (u *tlsUm) httpClient() *http.Client {
	return &http.Client{
		Timeout:       10 * time.Second,
		Transport:     &http.Transport{TLSClientConfig: u.klient(), ForceAttemptHTTP2: true, DisableKeepAlives: true},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// tlsVerbindung ist eine Verbindung mit beendetem Handshake und Keep-Alive.
type tlsVerbindung struct {
	c *tls.Conn
	r *bufio.Reader
}

func (u *tlsUm) tlsVerbinde(t *testing.T) *tlsVerbindung {
	t.Helper()
	c, err := tls.Dial("tcp", u.addr, u.klient())
	if err != nil {
		t.Fatalf("Handshake: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return &tlsVerbindung{c: c, r: bufio.NewReader(c)}
}

// get sendet einen GET auf /p ohne Connection: close und liest die Antwort.
func (v *tlsVerbindung) get(t *testing.T) (*http.Response, []byte) {
	t.Helper()
	_ = v.c.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := io.WriteString(v.c, "GET /p HTTP/1.1\r\nHost: h\r\n\r\n"); err != nil {
		t.Fatalf("Schreiben: %v", err)
	}
	resp, err := http.ReadResponse(v.r, &http.Request{Method: "GET"})
	if err != nil {
		t.Fatalf("Antwort: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Body: %v", err)
	}
	return resp, b
}

// gesehen hält, was der Dienst von einer Anfrage sah; der Body ist gelesen.
type gesehen struct {
	Methode, Pfad, RawQuery, KID string
	Header                       map[string][]string
	Body                         string
	Laenge                       int64
}

// strom liefert die Stücke eines Bodys; das nächste Stück erst, wenn der Client das vorige gelesen hat (Flush je Stück).
type strom struct {
	stuecke []string
	i       int
	gelesen chan struct{}
}

func (s *strom) Read(p []byte) (int, error) {
	if s.i >= len(s.stuecke) {
		return 0, io.EOF
	}
	if s.i > 0 {
		select {
		case <-s.gelesen:
		case <-time.After(5 * time.Second):
			return 0, errors.New("der Client las das vorige Stück nicht")
		}
	}
	s.i++
	return copy(p, s.stuecke[s.i-1]), nil
}

func (s *strom) Close() error { return nil }

// beobachtet ist, was der Client von einer Antwort sah.
type beobachtet struct {
	Status      int
	Header      map[string][]string
	Body        string
	Uebertragen []string
	Stuecke     []string
}

func beobachte(t *testing.T, resp *http.Response, gelesen chan<- struct{}) beobachtet {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	h := resp.Header.Clone()
	h.Del("Date")
	o := beobachtet{Status: resp.StatusCode, Header: h, Uebertragen: resp.TransferEncoding}
	puffer := make([]byte, 64)
	for {
		n, err := resp.Body.Read(puffer)
		if n > 0 {
			o.Body += string(puffer[:n])
			if gelesen != nil {
				o.Stuecke = append(o.Stuecke, string(puffer[:n]))
				gelesen <- struct{}{}
			}
		}
		if err != nil {
			if err != io.EOF {
				t.Fatalf("Body: %v", err)
			}
			return o
		}
	}
}

// TestAnfrageMitTLS (B13): der Adapter reicht kein Merkmal von TLS an Verarbeiten: dieselbe Anfrage und Antwort ohne und mit TLS (SPEC-291 Nr. 9 d).
// Menge T, 6 Elemente: 2 Transporte × 3 Fälle (GET mit Query und Header; POST mit Länge und POST chunked; Antwort chunked mit Flush je Stück);
// verglichen werden alle Felder der Anfrage und der Antwort.
func TestAnfrageMitTLS(t *testing.T) {
	var mu sync.Mutex
	var gelesen chan struct{}
	lauf := func(transport string) ([]gesehen, []beobachtet) {
		var seen []gesehen
		tu := func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			g := gesehen{Methode: a.Methode, Pfad: a.Pfad, RawQuery: a.RawQuery, KID: a.KorrelationsID, Header: kopiereMap(a.Header), Laenge: a.Laenge}
			if a.Body != nil {
				b, _ := io.ReadAll(a.Body)
				g.Body = string(b)
			}
			mu.Lock()
			seen = append(seen, g)
			ch := gelesen
			mu.Unlock()
			if a.Pfad == "/strom" {
				return inbound.Antwort{Status: 200, Header: map[string][]string{"X-Art": {"strom"}}, Body: &strom{stuecke: []string{"eins", "zwei", "drei"}, gelesen: ch}}, nil
			}
			return inbound.Antwort{Status: 201, Header: map[string][]string{"X-Art": {"normal"}, "Content-Type": {"text/plain"}}, Body: ausBytes([]byte("ok"))}, nil
		}
		var url string
		var client *http.Client
		if transport == "TLS" {
			u := neuTLSUmgebung(t, tu, nil)
			url, client = "https://"+u.addr, u.httpClient()
		} else {
			u := neuUmgebung(t, "record", tu)
			url, client = "http://"+u.addr, u.client()
		}
		var aus []beobachtet
		anfragen := []struct {
			methode, pfad string
			header        map[string]string
			body          io.Reader
		}{
			{"GET", "/get/Items?$filter=a%20eq%201&x=%C3%A4", map[string]string{"X-Test": "a", "Accept": "application/json", "If-Match": `W/"1"`}, nil},
			{"POST", "/post", map[string]string{"Content-Type": "application/json"}, strings.NewReader(`{"a":1}`)},
			{"POST", "/chunked", map[string]string{"Content-Type": "application/json"}, io.NopCloser(strings.NewReader(`{"b":2}`))},
			{"GET", "/strom", nil, nil},
		}
		for _, a := range anfragen {
			mu.Lock()
			gelesen = nil
			if a.pfad == "/strom" {
				gelesen = make(chan struct{}, 4)
			}
			ch := gelesen
			mu.Unlock()
			req, err := http.NewRequest(a.methode, url+a.pfad, a.body)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range a.header {
				req.Header.Set(k, v)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s %s %s: %v", transport, a.methode, a.pfad, err)
			}
			aus = append(aus, beobachte(t, resp, ch))
		}
		return seen, aus
	}
	klartextAnfragen, klartextAntworten := lauf("Klartext")
	tlsAnfragen, tlsAntworten := lauf("TLS")
	if len(klartextAnfragen) != 4 || len(tlsAnfragen) != 4 || len(klartextAntworten) != 4 {
		t.Fatalf("Mengen: %d, %d Anfragen, %d Antworten; erwartet 4 je Transport (3 Fälle, der zweite mit zwei Requests)", len(klartextAnfragen), len(tlsAnfragen), len(klartextAntworten))
	}
	if !reflect.DeepEqual(klartextAnfragen, tlsAnfragen) {
		t.Errorf("Anfragen unterscheiden sich:\nKlartext %+v\nTLS      %+v", klartextAnfragen, tlsAnfragen)
	}
	if !reflect.DeepEqual(klartextAntworten, tlsAntworten) {
		t.Errorf("Antworten unterscheiden sich:\nKlartext %+v\nTLS      %+v", klartextAntworten, tlsAntworten)
	}
	if got := klartextAnfragen[2]; got.Laenge != -1 || got.Body != `{"b":2}` {
		t.Errorf("chunked: Länge %d, Body %q; erwartet -1 und den Body", got.Laenge, got.Body)
	}
	if got := tlsAntworten[3].Stuecke; !reflect.DeepEqual(got, []string{"eins", "zwei", "drei"}) {
		t.Errorf("Flush je Stück über TLS: Stücke %q", got)
	}
}

// endeListener meldet, sobald der Server eine angenommene Verbindung schließt: nach dem Handshake-Fehler und der Antwort der Bibliothek.
type endeListener struct {
	net.Listener
	zu chan struct{}
}

func (l endeListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &endeConn{Conn: c, zu: l.zu}, nil
}

type endeConn struct {
	net.Conn
	zu     chan struct{}
	einmal sync.Once
}

func (c *endeConn) Close() error {
	c.einmal.Do(func() { c.zu <- struct{}{} })
	return c.Conn.Close()
}

// puffer ist ein Writer, den der Test liest, während der Server schreibt.
type puffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (p *puffer) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.Write(b)
}

func (p *puffer) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.String()
}

// TestHandshakeFehler (B14): ein gescheiterter Handshake erreicht den Adapter nicht: kein Verarbeiten, keine KorrelationsID, keine Zeile in einer Senke;
// danach kehrt Schliessen(0) sofort mit nil zurück (SPEC-291 Nr. 9 b). Menge H', 4 Elemente: unbekannte CA, Client allein mit h2, Client höchstens
// TLS 1.1, Klartext-Bytes.
func TestHandshakeFehler(t *testing.T) {
	var senke puffer
	altesZiel, alteFlags := log.Writer(), log.Flags()
	log.SetOutput(&senke)
	t.Cleanup(func() { log.SetOutput(altesZiel); log.SetFlags(alteFlags) })
	zu := make(chan struct{}, 8)
	u := neuTLSUmgebung(t, nil, func(l net.Listener) net.Listener { return endeListener{Listener: l, zu: zu} })
	pool := x509.NewCertPool()
	pool.AddCert(u.wurzel.Zertifikat)
	faelle := []struct {
		name string
		los  func() error
		fehl string
	}{
		{"unbekannte CA", func() error {
			_, err := tls.Dial("tcp", u.addr, &tls.Config{ServerName: "localhost", RootCAs: x509.NewCertPool()})
			return err
		}, "certificate signed by unknown authority"},
		{"Client allein mit h2", func() error {
			_, err := tls.Dial("tcp", u.addr, &tls.Config{ServerName: "localhost", RootCAs: pool, NextProtos: []string{"h2"}})
			return err
		}, "remote error: tls: no application protocol"},
		{"Client höchstens TLS 1.1", func() error {
			_, err := tls.Dial("tcp", u.addr, &tls.Config{ServerName: "localhost", RootCAs: pool, MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS11})
			return err
		}, "remote error: tls: protocol version not supported"},
		{"Klartext-Bytes", func() error {
			c, err := net.Dial("tcp", u.addr)
			if err != nil {
				return err
			}
			defer func() { _ = c.Close() }()
			_ = c.SetDeadline(time.Now().Add(5 * time.Second))
			if _, err := io.WriteString(c, "GET /p HTTP/1.1\r\nHost: h\r\n\r\n"); err != nil {
				return err
			}
			_, err = io.ReadAll(c)
			return err
		}, ""},
	}
	if len(faelle) != 4 {
		t.Fatalf("Menge H' hat %d Elemente, erwartet 4", len(faelle))
	}
	for _, f := range faelle {
		err := f.los()
		if f.fehl == "" && err != nil || f.fehl != "" && (err == nil || !strings.Contains(err.Error(), f.fehl)) {
			t.Errorf("%s: Fehler %v, erwartet %q", f.name, err, f.fehl)
		}
		select {
		case <-zu:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: der Server schloss die Verbindung nicht", f.name)
		}
	}
	if _, _, _, gesamt := u.d.zaehler(); gesamt != 0 {
		t.Errorf("Verarbeiten wurde %d-mal gerufen, erwartet 0", gesamt)
	}
	if n := u.ids.zahlen()[2]; n != 0 {
		t.Errorf("KorrelationsID wurde %d-mal gerufen, erwartet 0", n)
	}
	if got := senke.String(); got != "" {
		t.Errorf("log.Default() trägt %q, erwartet leer", got)
	}
	if dauer, err := schliesse(t, u.srv, 0); err != nil || dauer >= time.Second {
		t.Errorf("Schliessen(0) = %v nach %v, erwartet nil in unter 1 s", err, dauer)
	}
}

// zweiterAccept meldet den zweiten Aufruf von Accept: dann hat der Server die erste Verbindung angenommen und an ihre Goroutine übergeben.
type zweiterAccept struct {
	net.Listener
	n    atomic.Int64
	fest chan struct{}
}

func (l *zweiterAccept) Accept() (net.Conn, error) {
	if l.n.Add(1) == 2 {
		close(l.fest)
	}
	return l.Listener.Accept()
}

// pruefeAngehalten verlangt die Antwort 503 ODR-E5010 mit Connection: close (SPEC-136 Nr. 13 b).
func pruefeAngehalten(t *testing.T, was string, resp *http.Response, body []byte) {
	t.Helper()
	if resp.StatusCode != 503 || !resp.Close || !bytes.Equal(body, fehlerobjekt(t, model.CodeE5010)) {
		t.Errorf("%s: Status %d, Close %v, Body %q; erwartet 503, Close und ODR-E5010", was, resp.StatusCode, resp.Close, body)
	}
}

// TestAnhaltenMitTLS (B16): Anhalten gilt mit TLS: ein Request, dessen Zulassung danach liegt, erhält 503 ODR-E5010 mit Connection: close, auch wenn
// sein Handshake erst nach Anhalten endet; Verarbeiten wird nicht gerufen (SPEC-291 Nr. 9 c). Menge, 2 Elemente: Verbindung mit beendetem Handshake
// und Keep-Alive; Verbindung, die der Server angenommen hat und deren Handshake nach Anhalten endet.
func TestAnhaltenMitTLS(t *testing.T) {
	t.Run("beendeter Handshake, Keep-Alive", func(t *testing.T) {
		u := neuTLSUmgebung(t, nil, nil)
		v := u.tlsVerbinde(t)
		if resp, _ := v.get(t); resp.StatusCode != 200 || resp.Close {
			t.Fatalf("erster Request: Status %d, Close %v; erwartet 200 und Keep-Alive", resp.StatusCode, resp.Close)
		}
		u.srv.Anhalten()
		resp, body := v.get(t)
		pruefeAngehalten(t, "zweiter Request", resp, body)
		if _, _, _, gesamt := u.d.zaehler(); gesamt != 1 {
			t.Errorf("Verarbeiten wurde %d-mal gerufen, erwartet 1 (der erste Request)", gesamt)
		}
	})
	t.Run("Handshake endet nach Anhalten", func(t *testing.T) {
		var l *zweiterAccept
		fest := make(chan struct{})
		u := neuTLSUmgebung(t, nil, func(roh net.Listener) net.Listener {
			l = &zweiterAccept{Listener: roh, fest: fest}
			return l
		})
		c, err := net.Dial("tcp", u.addr)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = c.Close() }()
		select {
		case <-fest:
		case <-time.After(5 * time.Second):
			t.Fatal("der Server nahm die Verbindung nicht an")
		}
		u.srv.Anhalten()
		tc := tls.Client(c, u.klient())
		_ = tc.SetDeadline(time.Now().Add(5 * time.Second))
		if err := tc.Handshake(); err != nil {
			t.Fatalf("Handshake nach Anhalten: %v", err)
		}
		v := &tlsVerbindung{c: tc, r: bufio.NewReader(tc)}
		resp, body := v.get(t)
		pruefeAngehalten(t, "Request nach dem Handshake", resp, body)
		if _, _, _, gesamt := u.d.zaehler(); gesamt != 0 {
			t.Errorf("Verarbeiten wurde %d-mal gerufen, erwartet 0", gesamt)
		}
	})
}

// geschlossenInnerhalb meldet, ob die Verbindung binnen frist endet (Lesen liefert einen Fehler, kein Timeout).
func geschlossenInnerhalb(c net.Conn, frist time.Duration) bool {
	_ = c.SetReadDeadline(time.Now().Add(frist))
	_, err := c.Read(make([]byte, 16))
	return err != nil && !istZeitgrenze(err)
}

// TestSchliessenMitTLS (B17): Schliessen gilt mit TLS: ein Handshake in der Schwebe wird sofort geschlossen, ein laufender Request wird bis zur Frist
// abgewartet, danach ErrFristAbgelaufen (SPEC-291 Nr. 9 c). Menge, 4 Elemente: Verbindung ohne Byte; Verbindung mit halbem ClientHello; laufender
// Request, der innerhalb der Frist endet; laufender Request über die Frist.
func TestSchliessenMitTLS(t *testing.T) {
	// Der Anfang eines TLS-Records (Handshake, Version 1.0, Länge 240), dem der Rest fehlt.
	halbesClientHello := []byte{0x16, 0x03, 0x01, 0x00, 0xf0, 0x01, 0x00, 0x00, 0xec, 0x03, 0x03}
	for _, f := range []struct {
		name  string
		sende []byte
	}{{"Verbindung ohne Byte", nil}, {"Verbindung mit halbem ClientHello", halbesClientHello}} {
		t.Run(f.name, func(t *testing.T) {
			fest := make(chan struct{})
			u := neuTLSUmgebung(t, nil, func(roh net.Listener) net.Listener { return &zweiterAccept{Listener: roh, fest: fest} })
			c, err := net.Dial("tcp", u.addr)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = c.Close() }()
			if f.sende != nil {
				if _, err := c.Write(f.sende); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case <-fest:
			case <-time.After(5 * time.Second):
				t.Fatal("der Server nahm die Verbindung nicht an")
			}
			if dauer, err := schliesse(t, u.srv, 10*time.Second); err != nil || dauer >= time.Second {
				t.Errorf("Schliessen(10 s) = %v nach %v, erwartet nil in unter 1 s", err, dauer)
			}
			if !geschlossenInnerhalb(c, time.Second) {
				t.Error("die Verbindung ist nach Schliessen offen")
			}
		})
	}
	laufend := func(t *testing.T, frist, freigabe time.Duration) (ergebnis, error) {
		h := neuHaenger(t, wartetCtx, []byte("ok"))
		u := neuTLSUmgebung(t, h.tu, nil)
		r := make(chan ergebnis, 1)
		go func() {
			resp, err := u.httpClient().Get("https://" + u.addr + "/p")
			if err != nil {
				r <- ergebnis{err: err}
				return
			}
			defer func() { _ = resp.Body.Close() }()
			b, err := io.ReadAll(resp.Body)
			r <- ergebnis{status: resp.StatusCode, body: b, err: err}
		}()
		h.warteBegonnen(t, 1)
		if freigabe > 0 {
			time.AfterFunc(freigabe, h.gibFrei)
		}
		_, err := schliesse(t, u.srv, frist)
		return erwarte(t, r, "Antwort des laufenden Requests"), err
	}
	t.Run("laufender Request endet innerhalb der Frist", func(t *testing.T) {
		x, err := laufend(t, 10*time.Second, 100*time.Millisecond)
		if err != nil || x.err != nil || x.status != 200 || string(x.body) != "ok" {
			t.Errorf("Schliessen = %v, Antwort %+v; erwartet nil und 200 ok", err, x)
		}
	})
	t.Run("laufender Request über die Frist", func(t *testing.T) {
		_, err := laufend(t, 200*time.Millisecond, 0)
		if !errors.Is(err, httpodata.ErrFristAbgelaufen) {
			t.Errorf("Schliessen = %v, erwartet ErrFristAbgelaufen", err)
		}
	})
}
