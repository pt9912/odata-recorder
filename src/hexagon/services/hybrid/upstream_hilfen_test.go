package hybrid_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// sendung hält jedes Argument eines Aufrufs von Senden und die Bytes, die der Fake aus Body las.
type sendung struct {
	ctx     ctxMethoden
	a       outbound.UpstreamAnfrage
	gelesen []byte
}

// stufeU ist der Fake von UpstreamClient: er zählt jeden Aufruf, hält jedes Argument fest, liest den Body der Anfrage
// vollständig und antwortet mit fn (n zählt ab 0); ohne fn antwortet er mit Status 200 und einem Zählstrom.
type stufeU struct {
	fn      func(n int, ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error)
	mu      sync.Mutex
	aufrufe []sendung
}

func (s *stufeU) Senden(ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	var gel []byte
	if a.Body != nil {
		gel, _ = io.ReadAll(a.Body)
	}
	s.mu.Lock()
	n := len(s.aufrufe)
	s.aufrufe = append(s.aufrufe, sendung{ctx: ctx, a: a, gelesen: gel})
	s.mu.Unlock()
	if s.fn == nil {
		return outbound.UpstreamAntwort{Status: 200, Body: neuerStrom(nil)}, nil
	}
	return s.fn(n, ctx, a)
}

func (s *stufeU) zahl() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.aufrufe)
}

func (s *stufeU) erster(t *testing.T) sendung {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.aufrufe) == 0 {
		t.Fatal("Senden wurde nie gerufen")
	}
	return s.aufrufe[0]
}

// liefert lässt Senden mit Antwort und Fehler antworten.
func (s *stufeU) liefert(a outbound.UpstreamAntwort, err error) {
	s.fn = func(int, context.Context, outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) { return a, err }
}

// uPolicy ist eine Policy mit Hauptschalter an, den Hosts, den Routen und den Replay-Kategorien.
func uPolicy(t *testing.T, hosts, routen []string, kategorien ...string) *policy.Policy {
	t.Helper()
	return neuePolicy(t, policy.Werte{Upstream: true, Hosts: hosts, Schreibfreigaben: routen, ReplayKategorien: kategorien})
}

const (
	zielH     = "http://h"
	mockCode  = "ODR-E6003"
	missCode  = "ODR-E5001"
	standardP = "/odata/svc/Products"
)

// ausMockZuU lässt M mit ODR-E6003 scheitern; unter einer Policy ohne Replay-Kategorie folgt der Upstream.
func (g *rig) mockE6003(t *testing.T) *inbound.Fehler {
	t.Helper()
	f := code(codeNamed(t, mockCode))
	g.mockFehler(f)
	return f
}

// referenz ist ein Referenzdienst: er zählt Verbindungen und hält jeden Request fest.
type referenz struct {
	srv          *httptest.Server
	verbindungen atomic.Int32
	mu           sync.Mutex
	requests     []empfangen
}

type empfangen struct {
	Methode, Host, RawQuery, Pfad string
	Header                        http.Header
	Body                          []byte
}

func neuerReferenz(t *testing.T) *referenz {
	t.Helper()
	r := &referenz{}
	r.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		b, _ := io.ReadAll(q.Body)
		r.mu.Lock()
		r.requests = append(r.requests, empfangen{
			Methode: q.Method, Host: q.Host, RawQuery: q.URL.RawQuery, Pfad: q.URL.Path, Header: q.Header.Clone(), Body: b,
		})
		r.mu.Unlock()
		w.Header().Set("X-Upstream", "ja")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("antwort"))
	}))
	r.srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			r.verbindungen.Add(1)
		}
	}
	r.srv.Start()
	t.Cleanup(r.srv.Close)
	return r
}

// adresse ist host:port des Referenzdienstes.
func (r *referenz) adresse() string { return r.srv.Listener.Addr().String() }

func (r *referenz) alle() []empfangen {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]empfangen(nil), r.requests...)
}

// pruefClient ist der Prüf-Client: er liest Ziel mit net/http, sendet die Header unter ihrem Namen und folgt keiner
// Weiterleitung. Er ist kein Adapter.
type pruefClient struct{ c *http.Client }

func neuerPruefClient() pruefClient {
	return pruefClient{c: &http.Client{
		Transport:     &http.Transport{Proxy: nil, DisableKeepAlives: true},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

func (p pruefClient) Senden(ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	u := a.Ziel + a.Pfad
	if a.RawQuery != "" {
		u += "?" + a.RawQuery
	}
	var body io.Reader
	if a.Body != nil {
		body = a.Body
	}
	req, err := http.NewRequestWithContext(ctx, a.Methode, u, body)
	if err != nil {
		return outbound.UpstreamAntwort{}, errors.Join(outbound.ErrUpstreamNichtErreichbar, err)
	}
	if a.Body != nil {
		req.ContentLength = a.Laenge
	}
	for n, w := range a.Header {
		req.Header[n] = w
	}
	resp, err := p.c.Do(req)
	if err != nil {
		return outbound.UpstreamAntwort{}, errors.Join(outbound.ErrUpstreamNichtErreichbar, err)
	}
	return outbound.UpstreamAntwort{Status: resp.StatusCode, Header: resp.Header, Body: resp.Body}, nil
}

// rigMitClient baut den Dienst mit dem Prüf-Client statt des Fakes U.
func rigMitClient(t *testing.T, pol *policy.Policy, dienste []hybrid.Dienst) *rig {
	t.Helper()
	return neuesRigUp(t, pol, standardMax, dienste, neuerPruefClient())
}
