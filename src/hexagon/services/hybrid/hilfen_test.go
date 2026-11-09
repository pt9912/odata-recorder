package hybrid_test

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// Lesemodus einer Stufe: wie viel des Bodys der Anfrage sie vor ihrer Antwort liest.
const (
	liesNichts = iota
	liesEinByte
	liesAlles
)

// ctxMethoden ist die Methodenmenge von context.Context: die Fakes halten den ctx eines Aufrufs als Wert fest und
// vergleichen seine Identität.
type ctxMethoden interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// abbruchKontext liefert einen ctx mit Abbruchfunktion.
func abbruchKontext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// aufruf hält jedes Argument eines Aufrufs einer Stufe und das, was die Stufe aus dem Body las.
type aufruf struct {
	ctx     ctxMethoden
	a       inbound.Anfrage
	gelesen []byte
	leseErr error
}

// stufe ist der Fake von MockUseCase oder ReplayUseCase: sie zählt jeden Aufruf, hält jedes Argument fest und
// antwortet mit fn (n zählt ab 0).
type stufe struct {
	lies    int
	fn      func(n int, ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error)
	mu      sync.Mutex
	aufrufe []aufruf
}

func (s *stufe) rufe(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	var gel []byte
	var lerr error
	if a.Body != nil {
		switch s.lies {
		case liesEinByte:
			b := make([]byte, 1)
			var n int
			n, lerr = io.ReadFull(a.Body, b)
			gel = b[:n]
		case liesAlles:
			gel, lerr = io.ReadAll(a.Body)
		}
	}
	s.mu.Lock()
	n := len(s.aufrufe)
	s.aufrufe = append(s.aufrufe, aufruf{ctx: ctx, a: a, gelesen: gel, leseErr: lerr})
	s.mu.Unlock()
	if s.fn == nil {
		return inbound.Antwort{}, nil
	}
	return s.fn(n, ctx, a)
}

func (s *stufe) zahl() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.aufrufe)
}

func (s *stufe) erster(t *testing.T) aufruf {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.aufrufe) == 0 {
		t.Fatal("die Stufe wurde nie gerufen")
	}
	return s.aufrufe[0]
}

type fakeMock struct{ *stufe }

func (f fakeMock) Mock(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	return f.rufe(ctx, a)
}

type fakeReplay struct{ *stufe }

func (f fakeReplay) Replay(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	return f.rufe(ctx, a)
}

// strom ist ein Leser mit Close, der Read und Close zählt; als Body einer Antwort ein io.ReadCloser, als Body einer
// Anfrage ein io.Reader, der ein Close trotzdem zählt.
type strom struct {
	r      io.Reader
	reads  atomic.Int32
	closes atomic.Int32
}

func neuerStrom(inhalt []byte) *strom { return &strom{r: bytes.NewReader(inhalt)} }

func (s *strom) Read(p []byte) (int, error) {
	s.reads.Add(1)
	return s.r.Read(p)
}

func (s *strom) Close() error {
	s.closes.Add(1)
	return nil
}

// fehlerNach liefert zuerst die Bytes und danach bei jedem Read den Fehler.
type fehlerNach struct {
	rest []byte
	err  error
}

func (f *fehlerNach) Read(p []byte) (int, error) {
	if len(f.rest) == 0 {
		return 0, f.err
	}
	n := copy(p, f.rest)
	f.rest = f.rest[n:]
	return n, nil
}

// muster liefert n Bytes, deren Werte sich über die ersten 251 Stellen nicht wiederholen und die ab dem Wert seed
// beginnen: ein Orakel für die Reihenfolge und die Vollständigkeit eines Bodys.
func muster(n int, seed int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i*7 + seed) % 251)
	}
	return b
}

// code liefert den Fehler eines Codes mit dem ersten Status des Codes.
func code(c model.Code) *inbound.Fehler {
	return &inbound.Fehler{Code: c, Status: c.Stati()[0]}
}

func codeNamen() map[string]model.Code {
	m := map[string]model.Code{}
	for _, c := range model.AlleCodes() {
		m[c.String()] = c
	}
	return m
}

func codeNamed(t *testing.T, name string) model.Code {
	t.Helper()
	c, ok := codeNamen()[name]
	if !ok {
		t.Fatalf("Code %s fehlt in model.AlleCodes()", name)
	}
	return c
}

// anfrage liefert eine Anfrage mit frischer Header-Map, die alle drei Zuordnungs-Header trägt.
func anfrage(methode string) inbound.Anfrage {
	return inbound.Anfrage{
		Methode:        methode,
		Pfad:           "/odata/svc/Products",
		RawQuery:       "$top=1",
		KorrelationsID: "k-1",
		Header: map[string][]string{
			"X-Recorder-Workspace":         {"ws"},
			"X-Recorder-Recording-Session": {"rec"},
			"X-Recorder-Replay-Session":    {"rep"},
			"Accept":                       {"a", "b"},
		},
	}
}

func neuePolicy(t *testing.T, w policy.Werte) *policy.Policy {
	t.Helper()
	p, err := policy.New(w)
	if err != nil {
		t.Fatalf("policy.New: %v", err)
	}
	return p
}

const (
	katUnsupported = "unsupported"
	katCoverage    = "coverage_insufficient"
)

// offenste ist die offenste Policy: Hauptschalter an, ein Host frei, jede mutierende Methode auf jeder Route frei, beide
// Replay-Kategorien.
func offenste(t *testing.T) *policy.Policy {
	t.Helper()
	return neuePolicy(t, policy.Werte{
		Upstream:         true,
		Hosts:            []string{"h"},
		Schreibfreigaben: []string{"POST /**", "PUT /**", "PATCH /**", "DELETE /**", "MERGE /**"},
		ReplayKategorien: []string{katUnsupported, katCoverage},
	})
}

func mitKategorien(t *testing.T, ks ...string) *policy.Policy {
	t.Helper()
	return neuePolicy(t, policy.Werte{ReplayKategorien: ks})
}

// rig bündelt den Dienst und seine Fakes.
type rig struct {
	m, r *stufe
	v    *hybrid.Verteiler
}

const standardMax = 8

func neuesRig(t *testing.T, pol *policy.Policy, grenze int64) *rig {
	t.Helper()
	g := &rig{m: &stufe{}, r: &stufe{}}
	v, err := hybrid.New(hybrid.Konfig{
		Dienste:  []hybrid.Dienst{{Root: "/odata/svc"}},
		Mock:     fakeMock{g.m},
		Replay:   fakeReplay{g.r},
		Policy:   pol,
		MaxBytes: grenze,
	})
	if err != nil {
		t.Fatalf("hybrid.New: %v", err)
	}
	g.v = v
	return g
}

// mockFehler lässt den Mock mit err scheitern.
func (g *rig) mockFehler(err error) {
	g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return inbound.Antwort{}, err }
}

// replayAntwort lässt den Replay mit der Antwort a antworten.
func (g *rig) replayAntwort(a inbound.Antwort) {
	g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return a, nil }
}

func (g *rig) replayFehler(err error) {
	g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return inbound.Antwort{}, err }
}

func zeiger(m map[string][]string) uintptr { return reflect.ValueOf(m).Pointer() }

// tiefeKopie kopiert eine Header-Map samt Wertlisten; nil bleibt nil.
func tiefeKopie(m map[string][]string) map[string][]string {
	if m == nil {
		return nil
	}
	k := make(map[string][]string, len(m))
	for n, w := range m {
		if w == nil {
			k[n] = nil
			continue
		}
		k[n] = append([]string{}, w...)
	}
	return k
}

// zaehlen prüft die Zahl der Aufrufe beider Stufen.
func (g *rig) zaehlen(t *testing.T, was string, m, r int) {
	t.Helper()
	if got := g.m.zahl(); got != m {
		t.Errorf("%s: Mock %d-mal gerufen, erwartet %d", was, got, m)
	}
	if got := g.r.zahl(); got != r {
		t.Errorf("%s: Replay %d-mal gerufen, erwartet %d", was, got, r)
	}
}

func kennzeichen(a inbound.Antwort) []string {
	var w []string
	for n, v := range a.Header {
		if len(n) == len("x-recorder-fallback") && asciiGleich(n, "x-recorder-fallback") {
			w = append(w, v...)
		}
	}
	return w
}

// asciiGleich ist ein Orakel für den Vergleich ohne Beachtung der ASCII-Schreibweise, unabhängig vom Code.
func asciiGleich(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if 'A' <= x && x <= 'Z' {
			x += 32
		}
		if 'A' <= y && y <= 'Z' {
			y += 32
		}
		if x != y {
			return false
		}
	}
	return true
}
