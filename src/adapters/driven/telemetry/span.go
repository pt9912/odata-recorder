package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var _ support.Span = (*span)(nil)

// spanSchluessel ist der Schlüssel, unter dem ein ctx den Span dieses Pakets trägt (SPEC-230 Nr. 6).
type spanSchluessel struct{}

// span ist ein gestarteter Span. Der Adapter hält ihn nicht; ihn trägt allein der ctx des Aufrufers
// (SPEC-230 Nr. 7). mu schützt beendet und werte und liegt nie über Clock.Now() oder einem Write
// (SPEC-232 Nr. 5 f).
type span struct {
	a       *Adapter
	name    support.SpanName
	pflicht []string
	id      string
	parent  string
	start   time.Time

	mu      sync.Mutex
	beendet bool
	werte   map[string]support.Attr
}

// StartSpan legt für einen Namen aus support.Spans() einen Span an, dessen ctx den übergebenen
// ableitet; ein anderer Name zählt spanname und liefert den übergebenen ctx und einen Span ohne
// Wirkung (SPEC-230 Nr. 5, 6). Die Uhr wird einmal gelesen, vor jeder Sperre (SPEC-230 Nr. 7).
func (a *Adapter) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	pflicht, ok := a.pflicht[name]
	if !ok {
		a.verwerfen(support.VerwurfSpanname)
		return ctx, keinSpan{}
	}
	angenommen, verworfen := filtern(operation(name), attrs)
	a.verwerfen(verworfen...)
	s := &span{a: a, name: name, pflicht: pflicht, werte: map[string]support.Attr{}}
	s.id = a.naechsteID()
	if p, ok := ctx.Value(spanSchluessel{}).(*span); ok && p.a == a {
		s.parent = p.id
	}
	for _, x := range angenommen {
		s.werte[x.Key] = x
	}
	s.start = a.clock.Now()
	return context.WithValue(ctx, spanSchluessel{}, s), s
}

// naechsteID vergibt spn- und 16 Dezimalziffern aus dem Zähler des Adapters, der bei 1 beginnt
// (SPEC-230 Nr. 6, SPEC-232 Nr. 5 e).
func (a *Adapter) naechsteID() string {
	a.idMu.Lock()
	defer a.idMu.Unlock()
	a.spanZaehler++
	return fmt.Sprintf("spn-%016d", a.spanZaehler)
}

// Set ersetzt gültige Werte; nach dem Ende zählt es zustand einmal je Aufruf und prüft nichts
// (SPEC-230 Nr. 5, 7).
func (s *span) Set(attrs ...support.Attr) {
	s.mu.Lock()
	if s.beendet {
		s.mu.Unlock()
		s.a.verwerfen(support.VerwurfZustand)
		return
	}
	angenommen, verworfen := filtern(operation(s.name), attrs)
	for _, x := range angenommen {
		s.werte[x.Key] = x
	}
	s.mu.Unlock()
	s.a.verwerfen(verworfen...)
}

// End legt fest, welches Ende das erste ist, bevor es die Uhr liest; danach zeichnet es auf und
// schreibt die Zeile (SPEC-230 Nr. 7, SPEC-231 Nr. 6). Ein zweites End zählt zustand.
func (s *span) End() {
	s.mu.Lock()
	if s.beendet {
		s.mu.Unlock()
		s.a.verwerfen(support.VerwurfZustand)
		return
	}
	s.beendet = true
	attrs := make([]support.Attr, 0, len(s.pflicht))
	for _, k := range s.pflicht {
		if x, ok := s.werte[k]; ok {
			attrs = append(attrs, x)
		}
	}
	s.mu.Unlock()

	ende := s.a.clock.Now()
	dauer := max(ende.Sub(s.start), 0)
	s.a.aufzeichnen(support.SpanRecord{
		ID: s.id, ParentID: s.parent, Name: s.name, Start: s.start, Duration: dauer, Attrs: attrs,
	}, len(s.pflicht)-len(attrs))

	if slog.LevelInfo < s.a.minLevel {
		return
	}
	felder := make([]support.Attr, 0, 4+len(attrs))
	felder = append(felder, support.Str("ereignis", "span"), support.Str("span_id", s.id))
	if s.parent != "" {
		felder = append(felder, support.Str("parent_id", s.parent))
	}
	felder = append(felder, support.Millis("duration_ms", dauer))
	felder = append(felder, attrs...)
	s.a.zeileSchreiben(ende, slog.LevelInfo, string(s.name), felder)
}

// aufzeichnen legt r in Verteilung und Ring und zählt die fehlenden Pflicht-Attribute, alles in
// einem Abschnitt unter der Sperre des Stands (SPEC-232 Nr. 5 b).
func (a *Adapter) aufzeichnen(r support.SpanRecord, fehlend int) {
	eimer := len(a.grenzen)
	for i, g := range a.grenzen {
		if r.Duration <= g {
			eimer = i
			break
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dropped[support.VerwurfPflicht] += uint64(fehlend)
	v := a.durations[r.Name]
	v.Count++
	v.Sum += r.Duration
	v.Counts[eimer]++
	a.durations[r.Name] = v
	if len(a.ring) < a.maxSpans {
		a.ring = append(a.ring, r)
		return
	}
	a.ring[a.ringNaechster] = r
	a.ringNaechster = (a.ringNaechster + 1) % a.maxSpans
}
