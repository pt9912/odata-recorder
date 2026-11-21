package mock_test

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// telemetrieFake ist der Journal-Fake von Telemetry (SPEC-290): FakeTelemetry hält jedes Argument fest und zählt jede
// Operation; mitProt schreibt dazu jede Operation in das Protokoll der Port-Fakes, damit die Folge von Ports und
// Telemetrie in einer Liste steht. Ohne mitProt bleibt das Protokoll der übrigen Tests unverändert.
type telemetrieFake struct {
	*supporttest.FakeTelemetry
	p       *prot
	mu      sync.Mutex
	mitProt bool
}

func neueTelemetrie(p *prot) *telemetrieFake {
	return &telemetrieFake{FakeTelemetry: supporttest.NewFakeTelemetry(), p: p}
}

// journal schreibt op in das Protokoll, wenn mitProt gesetzt ist.
func (t *telemetrieFake) journal(ctx context.Context, op string, args ...any) {
	t.mu.Lock()
	an := t.mitProt
	t.mu.Unlock()
	if an {
		_ = t.p.beginn(ctx, op, args...)
	}
}

func (t *telemetrieFake) setzeMitProt() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mitProt = true
}

func (t *telemetrieFake) Log(ctx context.Context, level slog.Level, msg string, attrs ...support.Attr) {
	t.journal(ctx, "Telemetrie.Log", level, msg)
	t.FakeTelemetry.Log(ctx, level, msg, attrs...)
}

func (t *telemetrieFake) CountError(code model.Code) {
	t.journal(keinKontext(), "Telemetrie.CountError", code)
	t.FakeTelemetry.CountError(code)
}

func (t *telemetrieFake) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	t.journal(ctx, "Telemetrie.StartSpan", name)
	c, s := t.FakeTelemetry.StartSpan(ctx, name, attrs...)
	return c, &spanJournal{Span: s, t: t}
}

// spanJournal ist der Span des Fakes; Set und End erscheinen im Protokoll, Set mit dem Schlüssel des Attributs im Namen.
type spanJournal struct {
	support.Span
	t *telemetrieFake
}

func (s *spanJournal) Set(attrs ...support.Attr) {
	keys := make([]string, 0, len(attrs))
	for _, a := range attrs {
		keys = append(keys, a.Key)
	}
	s.t.journal(keinKontext(), "Span.Set "+strings.Join(keys, ","))
	s.Span.Set(attrs...)
}

func (s *spanJournal) End() {
	s.t.journal(keinKontext(), "Span.End")
	s.Span.End()
}

// spanBuch ist das Buch über einen Span des Fakes: seine Attribute in der Folge der Aufrufe (StartSpan, dann jedes Set), die
// Zahl der Set und der End.
type spanBuch struct {
	name  support.SpanName
	attrs []support.Attr
	sets  int
	enden int
}

// wert ist der letzte Wert des Schlüssels im Buch, der Wert am Ende des Spans; ok ist falsch, wenn der Schlüssel fehlt.
func (b spanBuch) wert(key string) (string, bool) {
	for _, a := range slices.Backward(b.attrs) {
		if a.Key == key {
			return a.Text, true
		}
	}
	return "", false
}

// buecher bildet je Span des Fakes ein Buch, in der Folge der StartSpan.
func buecher(f *supporttest.FakeTelemetry) []spanBuch {
	var r []spanBuch
	index := map[*supporttest.FakeSpan]int{}
	for _, a := range f.Aufrufe() {
		switch a.Op {
		case supporttest.OpStartSpan:
			index[a.Span] = len(r)
			r = append(r, spanBuch{name: a.Name, attrs: slices.Clone(a.Attrs)})
		case supporttest.OpSet:
			if i, da := index[a.Span]; da {
				r[i].attrs = append(r[i].attrs, a.Attrs...)
				r[i].sets++
			}
		case supporttest.OpEnd:
			if i, da := index[a.Span]; da {
				r[i].enden++
			}
		}
	}
	return r
}

// alleAttribute liefert jedes Attribut jedes Aufrufs des Fakes (StartSpan, Set, Log), geordnet nach Aufruf.
func alleAttribute(f *supporttest.FakeTelemetry) []support.Attr {
	var r []support.Attr
	for _, a := range f.Aufrufe() {
		r = append(r, a.Attrs...)
	}
	return r
}
