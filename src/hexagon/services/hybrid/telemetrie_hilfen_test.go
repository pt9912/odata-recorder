package hybrid_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// Schlüssel und Werte des Spans `odata.hybrid.fallback` als Literale des Tests (SPEC-321 Nr. 4), unabhängig vom Code.
const (
	attrArtTest      = "recorder.fallback_type"
	attrErgebnisTest = "recorder.fallback_result"
	frist            = 20 * time.Second
)

// journal ist das gemeinsame Protokoll der Telemetrie und der Stufen eines Dienstes (Journal-Fake, SPEC-321 Nr. 6): jeder
// Aufruf von Mock, Replay, Senden, CountDiscarded, StartSpan, Set, End und das Schließen eines Bodys steht in der Folge;
// dazu führt es Buch je Span (Attribute, Zahl der Set und End) und zählt Log und CountError, die der Dienst nie ruft.
type journal struct {
	mu    sync.Mutex
	folge []string
	spans []*jSpan
	fremd []string
}

func neuesJournal() *journal { return &journal{} }

func (j *journal) add(eintrag string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.folge = append(j.folge, eintrag)
}

// text ist die Folge der Aufrufe, durch Leerzeichen getrennt.
func (j *journal) text() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return strings.Join(j.folge, " ")
}

func (j *journal) spanListe() []*jSpan {
	j.mu.Lock()
	defer j.mu.Unlock()
	return append([]*jSpan(nil), j.spans...)
}

// gezaehlt liefert die Argumente der Aufrufe von CountDiscarded in der Folge der Aufrufe.
func (j *journal) gezaehlt() []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	var out []string
	for _, e := range j.folge {
		if z, ok := strings.CutPrefix(e, "Count("); ok {
			out = append(out, strings.TrimSuffix(z, ")"))
		}
	}
	return out
}

// jTel ist die Telemetrie des Dienstes im Test: sie schreibt in das Journal und liefert ein ctx mit eigenem Wert, so dass
// die Identität des ctx eines Aufrufs unterscheidbar bleibt (SPEC-321 Nr. 3 d).
type jTel struct{ j *journal }

// jSchluessel ist der Schlüssel des Wertes, den StartSpan seinem ctx gibt; anfrageSchluessel trägt die Kennung der Anfrage
// im ctx des Aufrufs, die der Span festhält (0 ohne).
type (
	jSchluessel       struct{}
	anfrageSchluessel struct{}
)

func (t jTel) Log(context.Context, slog.Level, string, ...support.Attr) {
	t.j.mu.Lock()
	defer t.j.mu.Unlock()
	t.j.fremd = append(t.j.fremd, "Log")
}

func (t jTel) CountError(code model.Code) {
	t.j.mu.Lock()
	defer t.j.mu.Unlock()
	t.j.fremd = append(t.j.fremd, "CountError("+code.String()+")")
}

func (t jTel) CountDiscarded(code model.Code) { t.j.add("Count(" + code.String() + ")") }

func (t jTel) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	s := &jSpan{j: t.j, name: name, ctx: ctx, start: append([]support.Attr(nil), attrs...)}
	s.anfrage, _ = ctx.Value(anfrageSchluessel{}).(int)
	t.j.mu.Lock()
	t.j.spans = append(t.j.spans, s)
	s.id = len(t.j.spans)
	t.j.mu.Unlock()
	t.j.add("Start(" + s.art() + ")")
	return context.WithValue(ctx, jSchluessel{}, s.id), s
}

// jSpan ist der Span des Journals mit seinem Buch.
type jSpan struct {
	j       *journal
	id      int
	anfrage int
	name    support.SpanName
	ctx     ctxMethoden
	start   []support.Attr
	mu      sync.Mutex
	sets    [][]support.Attr
	ends    int
}

func (s *jSpan) art() string {
	for _, a := range s.start {
		if a.Key == attrArtTest {
			return a.Text
		}
	}
	return "?"
}

func (s *jSpan) Set(attrs ...support.Attr) {
	s.mu.Lock()
	s.sets = append(s.sets, append([]support.Attr(nil), attrs...))
	s.mu.Unlock()
	ergebnis := "?"
	if len(attrs) > 0 {
		ergebnis = attrs[0].Text
	}
	s.j.add("Set(" + s.art() + "=" + ergebnis + ")")
}

func (s *jSpan) End() {
	s.mu.Lock()
	s.ends++
	s.mu.Unlock()
	s.j.add("End(" + s.art() + ")")
}

// buch liefert Zahl der Set, Zahl der End und das Ergebnis des ersten Set (leer ohne Set).
func (s *jSpan) buch() (sets, ends int, ergebnis string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sets) > 0 && len(s.sets[0]) == 1 {
		ergebnis = s.sets[0][0].Text
	}
	return len(s.sets), s.ends, ergebnis
}

// pruefeSpan hält das Buch eines Spans gegen SPEC-321 Nr. 4 und 6: der Name, genau ein Attribut mit der Art in
// StartSpan, genau ein Set mit genau einem Attribut mit einem der drei Ergebnisse und genau ein End.
func pruefeSpan(t *testing.T, was string, s *jSpan) {
	t.Helper()
	sets, ends, ergebnis := s.buch()
	if s.name != support.SpanHybridFallback {
		t.Errorf("%s: Span %d heißt %q", was, s.id, s.name)
	}
	if len(s.start) != 1 || s.start[0].Key != attrArtTest || (s.art() != "replay" && s.art() != "upstream") {
		t.Errorf("%s: Span %d trägt beim Beginn %+v; erwartet genau recorder.fallback_type mit replay oder upstream", was, s.id, s.start)
	}
	if sets != 1 || ergebnis == "" || s.sets[0][0].Key != attrErgebnisTest {
		t.Errorf("%s: Span %d hat %d Set (%+v); erwartet genau eines mit recorder.fallback_result", was, s.id, sets, s.sets)
	}
	if ergebnis != "antwort" && ergebnis != "fehler" && ergebnis != "abgebrochen" {
		t.Errorf("%s: Span %d trägt das Ergebnis %q", was, s.id, ergebnis)
	}
	if ends != 1 {
		t.Errorf("%s: Span %d hat %d End; erwartet genau 1", was, s.id, ends)
	}
}

// pruefeBuch hält das Buch des Journals: jeder Span nach pruefeSpan, kein Log und kein CountError (SPEC-321 Nr. 1).
func pruefeBuch(t *testing.T, was string, j *journal) {
	t.Helper()
	for _, s := range j.spanListe() {
		pruefeSpan(t, was, s)
	}
	j.mu.Lock()
	fremd := append([]string(nil), j.fremd...)
	j.mu.Unlock()
	if len(fremd) != 0 {
		t.Errorf("%s: der Dienst rief %v; erwartet kein Log und kein CountError", was, fremd)
	}
}

// jMock, jReplay und jUp sind die Stufen des Dienstes, die jeden Aufruf in das Journal schreiben.
type jMock struct {
	*stufe
	j *journal
}

func (f jMock) Mock(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	f.j.add("Mock")
	return f.rufe(ctx, a)
}

type jReplay struct {
	*stufe
	j *journal
}

func (f jReplay) Replay(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	f.j.add("Replay")
	return f.rufe(ctx, a)
}

type jUp struct {
	*stufeU
	j *journal
}

func (f jUp) Senden(ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	f.j.add("Senden")
	return f.stufeU.Senden(ctx, a)
}

// jBody ist ein Body mit dem Inhalt, dessen Close im Journal steht.
type jBody struct {
	j *journal
	io.Reader
}

func (b jBody) Close() error {
	b.j.add("Close")
	return nil
}

// neuesRigJ baut den Dienst mit den Stufen und der Telemetrie des Journals; ein Dienst unter /odata/svc mit der
// UpstreamURL ziel.
func neuesRigJ(t *testing.T, pol *policy.Policy, ziel string, j *journal) *rig {
	t.Helper()
	g := &rig{m: &stufe{}, r: &stufe{}, u: &stufeU{}}
	v, err := hybrid.New(hybrid.Konfig{
		Dienste: []hybrid.Dienst{{Root: "/odata/svc", UpstreamURL: ziel}}, Mock: jMock{g.m, j}, Replay: jReplay{g.r, j},
		Upstream: jUp{g.u, j}, Policy: pol, MaxBytes: standardMax, Telemetrie: jTel{j},
	})
	if err != nil {
		t.Fatalf("hybrid.New: %v", err)
	}
	g.v = v
	return g
}

// schaltCtx ist ein ctx, dessen Fehler der Test setzt: Err liefert nil, bis beende gerufen wurde, danach den übergebenen
// Fehler selbst (Canceled oder DeadlineExceeded, ohne die Zeit zu bemühen). Der Dienst liest allein Err.
type schaltCtx struct {
	mu  sync.Mutex
	err error
}

func (c *schaltCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *schaltCtx) Done() <-chan struct{}       { return nil }
func (c *schaltCtx) Value(any) any               { return nil }

func (c *schaltCtx) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *schaltCtx) beende(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
}

// verhalten ist, was eine Stufe tut: sie beendet den ctx des Aufrufs, wenn beende gesetzt ist, und liefert dann den Fehler
// (mit einem Body, den der Dienst schließen muss, wenn body gesetzt ist) oder die Antwort mit dem Status; header und
// inhalt füllen Header und Body der Antwort.
type verhalten struct {
	err    error
	status int
	body   bool
	beende error
	header map[string][]string
	inhalt string
}

func (v verhalten) anwenden(j *journal, c *schaltCtx) (inbound.Antwort, error) {
	if v.beende != nil {
		c.beende(v.beende)
	}
	var a inbound.Antwort
	if v.body || v.err == nil {
		a = inbound.Antwort{Status: v.status, Header: v.header, Body: jBody{j, strings.NewReader(v.inhalt)}}
	}
	return a, v.err
}

func (v verhalten) anwendenU(j *journal, c *schaltCtx) (outbound.UpstreamAntwort, error) {
	a, err := v.anwenden(j, c)
	return outbound.UpstreamAntwort{Status: a.Status, Header: a.Header, Body: a.Body}, err
}

// umgebung ist ein Dienst mit Journal und dem ctx, den der Test steuert.
type umgebung struct {
	g *rig
	j *journal
	c *schaltCtx
}

// neueUmgebung baut den Dienst unter pol; m, r und u sind das Verhalten der Stufen, vorher beendet den ctx vor dem Aufruf.
func neueUmgebung(t *testing.T, pol *policy.Policy, ziel string, m, r, u verhalten) umgebung {
	t.Helper()
	j, c := neuesJournal(), &schaltCtx{}
	g := neuesRigJ(t, pol, ziel, j)
	g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return m.anwenden(j, c) }
	g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return r.anwenden(j, c) }
	g.u.fn = func(int, context.Context, outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
		return u.anwendenU(j, c)
	}
	return umgebung{g, j, c}
}

func (u umgebung) rufe(methode string) (inbound.Antwort, error) {
	return u.g.v.Hybrid(u.c, anfrage(methode))
}

// Die Policies der Wege: nur Replay (beide Kategorien), nur Upstream, Replay und danach Upstream (ein Weg je Kategorie).
func polR(t *testing.T) *policy.Policy { return mitKategorien(t, katUnsupported, katCoverage) }

func polU(t *testing.T) *policy.Policy { return uPolicy(t, []string{"h"}, nil) }

func polK(t *testing.T) *policy.Policy { return uPolicy(t, []string{"h"}, nil, katCoverage) }
