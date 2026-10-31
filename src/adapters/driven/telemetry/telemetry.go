// Package telemetry ist der Driven Adapter der Ports Telemetry und TelemetryReader (SPEC-230 bis
// SPEC-232, ARC-020): Regeln der Attribute, Zähler je Fehlercode und der Stand zum Lesen. Er nutzt
// nur die Standardbibliothek. Dieser Kern schreibt keine Zeile und beendet keinen Span:
// Writer, Clock und MinLevel werden gehalten, nicht gelesen, und StartSpan ist ohne Wirkung.
package telemetry

import (
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var (
	_ support.Telemetry       = (*Adapter)(nil)
	_ support.TelemetryReader = (*Adapter)(nil)
)

const (
	maxSpansStandard = 1024
	maxSpansGrenze   = 65536
)

// Config ist die Eingabe von New (SPEC-232 Nr. 1).
type Config struct {
	Writer   io.Writer
	Clock    support.Clock
	MinLevel slog.Level
	MaxSpans int
}

// Adapter erfüllt Telemetry und TelemetryReader; er ist für gleichzeitigen Aufruf sicher
// (SPEC-232 Nr. 5): jeder Zugriff auf den Zustand liegt unter mu.
type Adapter struct {
	writer   io.Writer
	clock    support.Clock
	minLevel slog.Level
	maxSpans int

	mu        sync.Mutex
	errors    map[model.Code]uint64
	durations map[support.SpanName]support.Verteilung
	dropped   map[support.Verwurf]uint64
}

// New weist einen nil-Writer, eine nil-Clock und ein MaxSpans unter 0 oder über 65536 mit
// model.ErrUngueltig ab, ohne den Wert im Fehler; MaxSpans 0 gilt als 1024, jeder MinLevel ist
// gültig (SPEC-232 Nr. 1).
func New(cfg Config) (*Adapter, error) {
	if cfg.Writer == nil || cfg.Clock == nil || cfg.MaxSpans < 0 || cfg.MaxSpans > maxSpansGrenze {
		return nil, model.ErrUngueltig
	}
	if cfg.MaxSpans == 0 {
		cfg.MaxSpans = maxSpansStandard
	}
	a := &Adapter{
		writer:    cfg.Writer,
		clock:     cfg.Clock,
		minLevel:  cfg.MinLevel,
		maxSpans:  cfg.MaxSpans,
		errors:    map[model.Code]uint64{},
		durations: map[support.SpanName]support.Verteilung{},
		dropped:   map[support.Verwurf]uint64{},
	}
	for _, c := range model.AlleCodes() {
		a.errors[c] = 0
	}
	for _, s := range support.Spans() {
		a.durations[s.Name] = support.Verteilung{Counts: make([]uint64, len(support.Eimer())+1)}
	}
	for _, v := range support.Verwuerfe() {
		a.dropped[v] = 0
	}
	return a, nil
}

// Log wendet die Regeln der Attribute an (SPEC-230 Nr. 2 bis 4) und zählt die Verwürfe; es
// schreibt keine Zeile und wendet MinLevel nicht an.
func (a *Adapter) Log(_ context.Context, _ slog.Level, _ string, attrs ...support.Attr) {
	_, verworfen := filtern(opLog, attrs)
	a.verwerfen(verworfen...)
}

// StartSpan ist ohne Wirkung: es liefert einen nicht-nil ctx und einen Span, dessen Set und End
// nichts tun.
func (*Adapter) StartSpan(ctx context.Context, _ support.SpanName, _ ...support.Attr) (context.Context, support.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, keinSpan{}
}

// CountError erhöht den Zähler von code um 1; ein Code außerhalb der Tabelle erhöht keinen
// Zähler und zählt fehlercode (SPEC-230 Nr. 8).
func (a *Adapter) CountError(code model.Code) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.errors[code]; ok {
		a.errors[code]++
		return
	}
	a.dropped[support.VerwurfFehlercode]++
}

// Snapshot liefert eine Kopie des Stands; jeder Schlüssel steht darin, auch mit 0 (SPEC-230 Nr. 9).
func (a *Adapter) Snapshot() support.Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	return support.Snapshot{
		Errors:    a.errors,
		Durations: a.durations,
		Dropped:   a.dropped,
		Spans:     []support.SpanRecord{},
	}.Kopie()
}

func (a *Adapter) verwerfen(gruende ...support.Verwurf) {
	if len(gruende) == 0 {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, g := range gruende {
		a.dropped[g]++
	}
}

// keinSpan ist der Span ohne Wirkung.
type keinSpan struct{}

func (keinSpan) Set(...support.Attr) {}

func (keinSpan) End() {}
