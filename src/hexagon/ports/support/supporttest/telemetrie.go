package supporttest

import (
	"context"
	"log/slog"
	"slices"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var (
	_ support.Telemetry       = (*FakeTelemetry)(nil)
	_ support.TelemetryReader = (*FakeTelemetry)(nil)
	_ support.Span            = (*FakeSpan)(nil)
)

// Operation benennt eine der sechs Operationen der Ports Telemetry, Span und TelemetryReader.
type Operation int

// Die sechs Operationen.
const (
	OpLog Operation = iota + 1
	OpStartSpan
	OpCountError
	OpSet
	OpEnd
	OpSnapshot
)

// Aufruf ist ein aufgezeichneter Aufruf mit jedem Argument unverändert und ungeprüft. Ctx()
// liefert den übergebenen Wert mit seiner Identität (auch nil). Span ist bei StartSpan der
// zurückgegebene, bei Set und End der gerufene Span; Eltern ist bei StartSpan der Span aus dem
// ctx (sonst nil).
type Aufruf struct {
	Op     Operation
	ctx    func() context.Context
	Level  slog.Level
	Msg    string
	Name   support.SpanName
	Attrs  []support.Attr
	Code   model.Code
	Span   *FakeSpan
	Eltern *FakeSpan
}

// Ctx liefert den ctx des Aufrufs; ohne ctx im Aufruf (CountError, Set, End, Snapshot) nil.
func (a Aufruf) Ctx() context.Context {
	if a.ctx == nil {
		return nil
	}
	return a.ctx()
}

// halten bindet ctx, ohne ihn in einem Feld vom Typ context.Context abzulegen.
func halten(ctx context.Context) func() context.Context {
	return func() context.Context { return ctx }
}

// FakeTelemetry zeichnet jeden Aufruf auf und zählt jede Operation (SPEC-230 Nr. 10). Er prüft
// nichts, liefert keinen Fehler und gerät nicht in Panik; er ist für gleichzeitigen Aufruf sicher.
type FakeTelemetry struct {
	mu       sync.Mutex
	aufrufe  []Aufruf
	zaehler  map[Operation]int
	stand    support.Snapshot
	naechste int
}

// NewFakeTelemetry liefert einen Fake ohne Aufzeichnung und mit leerem Stand.
func NewFakeTelemetry() *FakeTelemetry {
	return &FakeTelemetry{zaehler: map[Operation]int{}}
}

type spanSchluessel struct{}

func (f *FakeTelemetry) aufzeichnen(a Aufruf) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.zaehler[a.Op]++
	f.aufrufe = append(f.aufrufe, a)
}

// Log zeichnet den Aufruf auf.
func (f *FakeTelemetry) Log(ctx context.Context, level slog.Level, msg string, attrs ...support.Attr) {
	f.aufzeichnen(Aufruf{Op: OpLog, ctx: halten(ctx), Level: level, Msg: msg, Attrs: slices.Clone(attrs)})
}

// StartSpan zeichnet den Aufruf samt Eltern-Span aus dem ctx auf und gibt einen ctx zurück, der
// den übergebenen ableitet (Abbruch, Frist und Werte bleiben) und den neuen Span trägt.
func (f *FakeTelemetry) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	basis := ctx
	if basis == nil {
		basis = context.Background()
	}
	eltern, _ := basis.Value(spanSchluessel{}).(*FakeSpan)
	f.mu.Lock()
	f.naechste++
	span := &FakeSpan{fake: f, ID: f.naechste, Name: name, Eltern: eltern}
	f.mu.Unlock()
	f.aufzeichnen(Aufruf{Op: OpStartSpan, ctx: halten(ctx), Name: name, Attrs: slices.Clone(attrs), Span: span, Eltern: eltern})
	return context.WithValue(basis, spanSchluessel{}, span), span
}

// CountError zeichnet den Aufruf auf.
func (f *FakeTelemetry) CountError(code model.Code) {
	f.aufzeichnen(Aufruf{Op: OpCountError, Code: code})
}

// Snapshot zählt den Aufruf und liefert eine Kopie des gesetzten Stands, sonst den leeren.
func (f *FakeTelemetry) Snapshot() support.Snapshot {
	f.aufzeichnen(Aufruf{Op: OpSnapshot})
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stand.Kopie()
}

// SetzeStand legt den Stand fest, den Snapshot liefert; s wird kopiert.
func (f *FakeTelemetry) SetzeStand(s support.Snapshot) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stand = s.Kopie()
}

// Aufrufe liefert die Aufzeichnung in der Reihenfolge der Aufrufe als Kopie.
func (f *FakeTelemetry) Aufrufe() []Aufruf {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := slices.Clone(f.aufrufe)
	for i := range out {
		out[i].Attrs = slices.Clone(out[i].Attrs)
	}
	return out
}

// Anzahl liefert, wie oft op gerufen wurde.
func (f *FakeTelemetry) Anzahl(op Operation) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.zaehler[op]
}

// FakeSpan ist der Span des Fakes; ID zählt je Fake ab 1, Eltern ist der Span aus dem ctx von
// StartSpan oder nil.
type FakeSpan struct {
	fake   *FakeTelemetry
	ID     int
	Name   support.SpanName
	Eltern *FakeSpan
}

// Set zeichnet den Aufruf auf, jeden, auch nach End.
func (s *FakeSpan) Set(attrs ...support.Attr) {
	s.fake.aufzeichnen(Aufruf{Op: OpSet, Attrs: slices.Clone(attrs), Span: s})
}

// End zeichnet den Aufruf auf, jeden, auch das zweite.
func (s *FakeSpan) End() {
	s.fake.aufzeichnen(Aufruf{Op: OpEnd, Span: s})
}
