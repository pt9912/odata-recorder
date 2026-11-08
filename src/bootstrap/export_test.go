package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// EtagFuerSatz ist die Entscheidung des Composition Root, ob ein Dienst mit diesem Satz ETags liefert (SPEC-280 Nr. 2).
func EtagFuerSatz(satz outbound.CodecSatz, an bool) bool { return etagFuerSatz(satz, an) }

// SatzVersion4 ist der Satz der Codecs der Version 4.0 des Mock.
func SatzVersion4() outbound.CodecSatz { return satzVersion4() }

// Die Naht der Tests: Das Testpaket bootstrap_test setzt die Haken von Bauteile und ruft Start und
// Build mit ihnen; der Betrieb nutzt standardBauteile.

// Bauteile ist die Naht der Verdrahtung (SPEC-139).
type Bauteile = bauteile

// Server ist, was der Root vom Listener-Adapter ruft.
type Server = server

// StandardBauteile liefert die Konstruktoren des Betriebs.
func StandardBauteile() Bauteile { return standardBauteile() }

// StartMit ist Start mit gegebener Umgebung und gegebenen Bauteilen.
func StartMit(ctx context.Context, environ []string, stderr io.Writer, b Bauteile) int {
	return start(ctx, environ, stderr, b)
}

// BuildMit ist Build mit gegebenen Bauteilen.
func BuildMit(ctx context.Context, k Konfiguration, log *slog.Logger, b Bauteile) (*Handle, error) {
	return build(ctx, k, log, b)
}

// IDsVon liefert den IDGenerator des Handle.
func IDsVon(h *Handle) support.IDGenerator { return h.ids }

// Warte ist das Warten von Start auf den Abbruch von ctx oder das Ende von Serve.
func Warte(ctx context.Context, fehler <-chan error) error { return warte(ctx, fehler) }

// ErrServeBeendet ist der Ausfall, wenn Serve ohne Anhalten und ohne Abbruch von ctx zurückkehrt.
var ErrServeBeendet = errServeBeendet

// Meldung ist die Zeile, die Start zu einem Fehler schreibt.
func Meldung(err error) string { return meldung(err) }

// LiteralHandle liefert den Handle eines hochgefahrenen Prozesses aus Fakes; srv.Serve läuft in einer
// eigenen Goroutine, und Herunterfahren wartet auf ihre Rückkehr.
func LiteralHandle(modus string, frist time.Duration, uhr support.Clock, sessions inbound.SessionUseCase, srv Server, ablage io.Closer) *Handle {
	h := &Handle{
		k: Konfiguration{Modus: modus, ShutdownTimeout: frist}, uhr: uhr, sessions: sessions, server: srv, ablage: ablage,
		serveEnd: make(chan struct{}), fehler: make(chan error, 1), sessionBegonnen: true, zustand: hochgefahren,
	}
	go func() {
		_ = srv.Serve()
		close(h.serveEnd)
	}()
	return h
}
