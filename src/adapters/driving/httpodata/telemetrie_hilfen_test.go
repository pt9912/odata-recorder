package httpodata_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// journal ist das gemeinsame Protokoll der Fakes eines Tests: IDs, Telemetrie, Dienst und ResponseWriter schreiben je einen Eintrag
// je Aufruf in der Folge der Aufrufe.
type journal struct {
	mu sync.Mutex
	z  []string
}

func (j *journal) add(s string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.z = append(j.z, s)
}

func (j *journal) folge() []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return append([]string(nil), j.z...)
}

// jTel ist der Journal-Fake der Telemetrie: ein FakeTelemetry, dessen Operationen zusätzlich ins Protokoll schreiben; der Haltepunkt
// (setzeBei) läuft je Operation vor ihr und vor der Aufzeichnung des Fakes, der Haltepunkt danach (setzeNach) nach der Aufzeichnung, außerhalb
// jeder Sperre des Fakes, mit dem Namen der Operation ohne Code (`StartSpan`, `CountError`, `Set`, `End`).
type jTel struct {
	*supporttest.FakeTelemetry
	j    *journal
	mu   sync.Mutex
	hook func(op string)
	nach func(op string)
}

func neuJTel(j *journal) *jTel { return &jTel{FakeTelemetry: supporttest.NewFakeTelemetry(), j: j} }

func (t *jTel) setzeBei(f func(op string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hook = f
}

func (t *jTel) setzeNach(f func(op string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nach = f
}

func (t *jTel) halte(op string) {
	t.j.add(op)
	t.mu.Lock()
	f := t.hook
	t.mu.Unlock()
	if f != nil {
		f(op)
	}
}

func (t *jTel) danach(op string) {
	t.mu.Lock()
	f := t.nach
	t.mu.Unlock()
	if f != nil {
		f(op)
	}
}

func (t *jTel) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	t.halte("StartSpan")
	ctx, sp := t.FakeTelemetry.StartSpan(ctx, name, attrs...)
	t.danach("StartSpan")
	return ctx, &jSpan{Span: sp, t: t}
}

func (t *jTel) CountError(code model.Code) {
	t.halte("CountError:" + code.String())
	t.FakeTelemetry.CountError(code)
	t.danach("CountError")
}

// jSpan ist der Span des Journal-Fakes.
type jSpan struct {
	support.Span
	t *jTel
}

func (s *jSpan) Set(attrs ...support.Attr) {
	s.t.halte("Set")
	s.Span.Set(attrs...)
	s.t.danach("Set")
}

func (s *jSpan) End() {
	s.t.halte("End")
	s.Span.End()
	s.t.danach("End")
}

// jIDs schreibt jeden Aufruf von KorrelationsID ins Protokoll.
type jIDs struct {
	*supporttest.FestIDs
	j *journal
}

func (i *jIDs) KorrelationsID() string {
	i.j.add("KorrelationsID")
	return i.FestIDs.KorrelationsID()
}

// jWriter ist der aufzeichnende ResponseWriter: WriteHeader, Write und Flush stehen im Protokoll; der Status ist der des ersten
// WriteHeader, schreibFehler der Fehler jedes Write.
type jWriter struct {
	j             *journal
	h             http.Header
	status        int
	koepfe        int
	body          bytes.Buffer
	schreibFehler error
}

func neuJWriter(j *journal) *jWriter { return &jWriter{j: j, h: http.Header{}} }

func (w *jWriter) Header() http.Header { return w.h }

func (w *jWriter) WriteHeader(status int) {
	w.j.add(fmt.Sprintf("WriteHeader:%d", status))
	w.koepfe++
	if w.koepfe == 1 {
		w.status = status
	}
}

func (w *jWriter) Write(b []byte) (int, error) {
	w.j.add("Write")
	if w.schreibFehler != nil {
		return 0, w.schreibFehler
	}
	return w.body.Write(b)
}

func (w *jWriter) Flush() { w.j.add("Flush") }

// jRequest ist ein Request, wie ihn der Server dem Handler gibt: mit RequestURI und Body NoBody.
func jRequest(ctx context.Context, methode, ziel string) *http.Request {
	r, err := http.NewRequestWithContext(ctx, methode, "http://h"+ziel, http.NoBody)
	if err != nil {
		panic(err)
	}
	r.RequestURI = ziel
	return r
}

// probe ist ein Server ohne Verbindung: der Test ruft seinen Handler mit jWriter und jRequest.
type probe struct {
	j   *journal
	tel *jTel
	srv *httpodata.Server
	h   http.Handler
}

// neuProbe baut den Server mit dem Journal-Fake der Telemetrie, dem Modus und dem Dienst tu; tu darf nil sein (200 mit leerem Body).
func neuProbe(t *testing.T, modus string, tu httpodata.Verarbeiter) *probe {
	t.Helper()
	j := &journal{}
	return neuProbeMit(t, modus, j, neuJTel(j), tu)
}

func neuProbeMit(t *testing.T, modus string, j *journal, tel *jTel, tu httpodata.Verarbeiter) *probe {
	t.Helper()
	verarbeiten := func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		j.add("Verarbeiten")
		if tu == nil {
			return inbound.Antwort{Status: 200, Body: leer()}, nil
		}
		return tu(ctx, a)
	}
	srv, err := httpodata.New(httpodata.Konfig{
		Verarbeiten: verarbeiten, IDs: &jIDs{FestIDs: supporttest.NewFestIDs(), j: j}, Listener: &zaehlListener{},
		Telemetrie: tel, Modus: modus,
	})
	if err != nil {
		t.Fatal(err)
	}
	return &probe{j: j, tel: tel, srv: srv, h: srv.HandlerFuerTest()}
}

// ruf ruft den Handler; abbruch ist wahr, wenn er den Abbruch der Verbindung als Panik meldet (SPEC-136 Nr. 9).
func (p *probe) ruf(w http.ResponseWriter, r *http.Request) (abbruch bool) {
	defer func() {
		if v := recover(); v != nil {
			if v != http.ErrAbortHandler {
				panic(v)
			}
			abbruch = true
		}
	}()
	p.h.ServeHTTP(w, r)
	return false
}

// spanBuch ist das Buch je Span des Fakes: die Attribute von StartSpan, jeder Aufruf von Set und die Zahl der End.
type spanBuch struct {
	name   support.SpanName
	start  []support.Attr
	sets   [][]support.Attr
	ends   int
	eltern *supporttest.FakeSpan
	span   *supporttest.FakeSpan
}

// buecher führt das Buch je Span in der Folge von StartSpan.
func buecher(f *supporttest.FakeTelemetry) []*spanBuch {
	var out []*spanBuch
	nach := map[*supporttest.FakeSpan]*spanBuch{}
	for _, a := range f.Aufrufe() {
		switch a.Op {
		case supporttest.OpStartSpan:
			b := &spanBuch{name: a.Name, start: a.Attrs, eltern: a.Eltern, span: a.Span}
			nach[a.Span] = b
			out = append(out, b)
		case supporttest.OpSet:
			nach[a.Span].sets = append(nach[a.Span].sets, a.Attrs)
		case supporttest.OpEnd:
			nach[a.Span].ends++
		}
	}
	return out
}

// attrText liefert den Text des Attributs key aus attrs und ob es da ist.
func attrText(attrs []support.Attr, key string) (string, bool) {
	for _, a := range attrs {
		if a.Key == key {
			return a.Text, true
		}
	}
	return "", false
}

// zaehlerFolge liefert die Codes der Aufrufe von CountError in der Folge der Aufrufe.
func zaehlerFolge(f *supporttest.FakeTelemetry) []model.Code {
	var out []model.Code
	for _, a := range f.Aufrufe() {
		if a.Op == supporttest.OpCountError {
			out = append(out, a.Code)
		}
	}
	return out
}

// weg ist einer der vierzehn Wege des Handlers nach dem Beginn (SPEC-294 Nr. 7 a): was der Dienst liefert, was der Client sieht und was
// die Telemetrie sehen muss. status 0 heißt: kein WriteHeader; zaehler 0: kein Fehlerobjekt.
type weg struct {
	name          string
	methode       string
	tu            httpodata.Verarbeiter
	anhalten      bool
	schreibFehler error
	status        int
	zaehler       model.Code
	abbruch       bool
}

func antwortMit(status int, body *koerper) httpodata.Verarbeiter {
	return func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
		return inbound.Antwort{Status: status, Body: body}, nil
	}
}

func fehlerMit(err error) httpodata.Verarbeiter {
	return func(context.Context, inbound.Anfrage) (inbound.Antwort, error) { return inbound.Antwort{}, err }
}

// bruchOhneBody ist die Antwort des Dienstes mit Body nil und Status 200 (ein Bruch des Vertrags, SPEC-136 Nr. 5a).
func bruchOhneBody() httpodata.Verarbeiter {
	return func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
		return inbound.Antwort{Status: 200}, nil
	}
}

// wege liefert die vierzehn Wege von SPEC-294 Nr. 7 a, mit dem erwarteten Status und Zähler als Literal.
func wege() []weg {
	lesefehler := &koerper{lies: func([]byte) (int, error) { return 0, errors.New("Lesen") }}
	return []weg{
		{name: "Antwort mit Body", methode: "GET", tu: antwortMit(200, ausBytes([]byte("abc"))), status: 200},
		{name: "Antwort ohne Body", methode: "HEAD", tu: antwortMit(200, leer()), status: 200},
		{name: "Fehler mit Code", methode: "GET", tu: fehlerMit(&inbound.Fehler{Code: model.CodeE6008, Status: 404}), status: 404, zaehler: model.CodeE6008},
		{name: "Fehler ohne Code", methode: "GET", tu: fehlerMit(errors.New("roh")), status: 500, zaehler: model.CodeE9001},
		{name: "Bruch Body nil", methode: "GET", tu: bruchOhneBody(), status: 500, zaehler: model.CodeE9001},
		{name: "Bruch Status 100", methode: "GET", tu: antwortMit(100, leer()), status: 500, zaehler: model.CodeE9001},
		{name: "Bruch Status 199", methode: "GET", tu: antwortMit(199, leer()), status: 500, zaehler: model.CodeE9001},
		{name: "Bruch Status 600", methode: "GET", tu: antwortMit(600, leer()), status: 500, zaehler: model.CodeE9001},
		{name: "Antwort 503", methode: "GET", anhalten: true, status: 503, zaehler: model.CodeE5010},
		{name: "Canceled", methode: "GET", tu: fehlerMit(context.Canceled), abbruch: true},
		{name: "DeadlineExceeded", methode: "GET", tu: fehlerMit(context.DeadlineExceeded), abbruch: true},
		{name: "Canceled mit Fehler", methode: "GET", tu: fehlerMit(errors.Join(context.Canceled, &inbound.Fehler{Code: model.CodeE6008, Status: 404})), abbruch: true},
		{name: "Lesefehler nach dem Kopf", methode: "GET", tu: antwortMit(200, lesefehler), status: 200, abbruch: true},
		{name: "Schreibfehler beim Client", methode: "GET", tu: antwortMit(200, ausBytes([]byte("abc"))), schreibFehler: errors.New("Schreiben"), status: 200},
	}
}

// wegMitMethoden liefert den Weg i von wege() einmal je Methode von M-Methoden (14 frische Fälle).
func wegMitMethoden(i int) []weg {
	var out []weg
	for _, m := range methodenSatz() {
		w := wege()[i]
		w.methode = m
		out = append(out, w)
	}
	return out
}

// lauf führt den Weg w durch einen frischen Server: Aufruf des Handlers mit der Methode des Weges, vorher Anhalten, wenn der Weg es
// verlangt. Das Ergebnis trägt die Probe, den Writer und ob der Handler abbrach.
func lauf(t *testing.T, modus string, w weg) (*probe, *jWriter, bool) {
	t.Helper()
	p := neuProbe(t, modus, w.tu)
	if w.anhalten {
		p.srv.Anhalten()
	}
	jw := neuJWriter(p.j)
	jw.schreibFehler = w.schreibFehler
	abbruch := p.ruf(jw, jRequest(context.Background(), w.methode, "/p"))
	return p, jw, abbruch
}
