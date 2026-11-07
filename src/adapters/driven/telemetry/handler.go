package telemetry

import (
	"context"
	"log/slog"
	"math"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var _ slog.Handler = (*handler)(nil)

// handler schreibt über eintrag dieselbe Zeile wie Log (SPEC-231 Nr. 5). Alle aus einem Handler
// abgeleiteten Handler teilen den Adapter und damit Senke, Sperre und Stand. praefix und vorgelegt
// sind nach dem Ableiten unveränderlich; Handle baut je Aufruf eine neue Liste.
type handler struct {
	a         *Adapter
	praefix   string
	vorgelegt []support.Attr
}

// Handler liefert den slog.Handler des Adapters (SPEC-231 Nr. 5).
func (a *Adapter) Handler() slog.Handler { return &handler{a: a} }

// Enabled folgt MinLevel; es liest ctx nicht, ruft die Uhr nicht und zählt nichts.
func (h *handler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.a.minLevel }

// Handle liefert immer nil, ignoriert Record.Time und Record.PC und wendet MinLevel selbst an.
func (h *handler) Handle(_ context.Context, r slog.Record) error {
	attrs := make([]support.Attr, 0, len(h.vorgelegt)+r.NumAttrs())
	attrs = append(attrs, h.vorgelegt...)
	r.Attrs(func(x slog.Attr) bool {
		attrs = append(attrs, abbilden(h.praefix, x))
		return true
	})
	h.a.eintrag(r.Level, r.Message, attrs)
	return nil
}

// WithAttrs bildet attrs mit dem Präfix des Empfängers ab und legt sie hinter die vorgelegten des
// Empfängers; ein leeres attrs liefert den Empfänger.
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	vorgelegt := slices.Clone(h.vorgelegt)
	for _, x := range attrs {
		vorgelegt = append(vorgelegt, abbilden(h.praefix, x))
	}
	return &handler{a: h.a, praefix: h.praefix, vorgelegt: vorgelegt}
}

// WithGroup setzt name und einen Punkt vor jeden folgenden Schlüssel; ein leerer Name liefert den
// Empfänger.
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &handler{a: h.a, praefix: h.praefix + name + ".", vorgelegt: h.vorgelegt}
}

// abbilden bildet x in ein Attribut mit dem Schlüssel praefix+x.Key: ein LogValuer wird aufgelöst,
// String, Int64, Uint64 bis MaxInt64, Float64 und Duration (in Millisekunden) haben ihre Art, jede
// andere Art (auch ein error) und ein Uint64 darüber haben die Art 0, die filtern als wert verwirft
// (SPEC-231 Nr. 5).
func abbilden(praefix string, x slog.Attr) support.Attr {
	key, v := praefix+x.Key, x.Value.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return support.Str(key, v.String())
	case slog.KindInt64:
		return support.Int(key, v.Int64())
	case slog.KindUint64:
		if u := v.Uint64(); u <= math.MaxInt64 {
			return support.Int(key, int64(u))
		}
	case slog.KindFloat64:
		return support.Attr{Key: key, Art: support.ArtFloat, Gleit: v.Float64()}
	case slog.KindDuration:
		return support.Millis(key, v.Duration())
	}
	return support.Attr{Key: key}
}
