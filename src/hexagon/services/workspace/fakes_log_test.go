package workspace_test

import (
	"context"
	"log/slog"
	"sync"
)

// logWurzel ist die Ablage der Einträge, die alle abgeleiteten Handler (WithAttrs) teilen.
type logWurzel struct {
	mu   sync.Mutex
	list []eintragsKopie
}

// logHandler ist der Fake des Protokolls (F6): ein slog.Handler, der Stufe, Text und Attribute festhält,
// auch die Attribute, die ein abgeleiteter Logger vorab setzt.
type logHandler struct {
	w   *logWurzel
	pre []slog.Attr
}

func neuerLogHandler() *logHandler { return &logHandler{w: &logWurzel{}} }

func (h *logHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *logHandler) Handle(_ context.Context, r slog.Record) error {
	e := eintragsKopie{stufe: r.Level, text: r.Message, attr: map[string]string{}}
	for _, a := range h.pre {
		e.attr[a.Key] = a.Value.String()
	}
	r.Attrs(func(a slog.Attr) bool { e.attr[a.Key] = a.Value.String(); return true })
	h.w.mu.Lock()
	h.w.list = append(h.w.list, e)
	h.w.mu.Unlock()
	return nil
}

func (h *logHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return &logHandler{w: h.w, pre: append(append([]slog.Attr{}, h.pre...), as...)}
}

func (h *logHandler) WithGroup(string) slog.Handler { return h }

func (h *logHandler) eintraege() []eintragsKopie {
	h.w.mu.Lock()
	defer h.w.mu.Unlock()
	return append([]eintragsKopie{}, h.w.list...)
}
