package workspace_test

import (
	"context"
	"log/slog"
	"sync"
)

// logWurzel ist die Ablage der Einträge, die alle abgeleiteten Handler (WithAttrs, WithGroup) teilen.
type logWurzel struct {
	mu   sync.Mutex
	list []eintragsKopie
}

// logHandler ist der Fake des Protokolls (F6): ein slog.Handler, der Stufe, Text und Attribute festhält. Er führt
// Gruppen und Attribute eines abgeleiteten Loggers mit und qualifiziert die Namen wie ein Handler des Originals
// („g.name“): ein Dienst, der den Logger mit With oder WithGroup ableitet, verändert so die aufgezeichneten Namen.
type logHandler struct {
	w       *logWurzel
	pre     []slog.Attr // Attribute aus WithAttrs, mit dem Präfix ihrer Zeit qualifiziert
	praefix string      // Namen der geöffneten Gruppen, je mit Punkt
}

func neuerLogHandler() *logHandler { return &logHandler{w: &logWurzel{}} }

func (h *logHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *logHandler) Handle(_ context.Context, r slog.Record) error {
	e := eintragsKopie{stufe: r.Level, text: r.Message, attr: map[string]string{}}
	for _, a := range h.pre {
		e.attr[a.Key] = a.Value.String()
	}
	r.Attrs(func(a slog.Attr) bool { e.attr[h.praefix+a.Key] = a.Value.String(); return true })
	h.w.mu.Lock()
	h.w.list = append(h.w.list, e)
	h.w.mu.Unlock()
	return nil
}

func (h *logHandler) WithAttrs(as []slog.Attr) slog.Handler {
	pre := append([]slog.Attr{}, h.pre...)
	for _, a := range as {
		pre = append(pre, slog.Attr{Key: h.praefix + a.Key, Value: a.Value})
	}
	return &logHandler{w: h.w, pre: pre, praefix: h.praefix}
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &logHandler{w: h.w, pre: h.pre, praefix: h.praefix + name + "."}
}

func (h *logHandler) eintraege() []eintragsKopie {
	h.w.mu.Lock()
	defer h.w.mu.Unlock()
	return append([]eintragsKopie{}, h.w.list...)
}
