package bootstrap

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// EtagFuerSatz ist die Entscheidung des Composition Root, ob ein Dienst mit diesem Satz ETags liefert (SPEC-280 Nr. 2).
func EtagFuerSatz(satz outbound.CodecSatz, an bool) bool { return etagFuerSatz(satz, an) }

// Leerraum sind die Zeichen, die die Konfiguration am Rand entfernt (SPEC-110 Nr. 1).
const Leerraum = leerraum

// Kappe entfernt Leerraum an beiden Rändern, wie die Konfiguration es an jeder Stelle tut (SPEC-110 Nr. 1).
func Kappe(s string) string { return kappe(s) }

// IstQueryName ist die Prüfung eines Eintrags von REDACT_QUERY_PARAMS (SPEC-110 Nr. 6).
func IstQueryName(s string) bool { return istQueryName(s) }

// TLSVerstoss ist ein Verstoß von ladeTLS: der Schlüssel, unter dem der Aufrufer meldet, und der Grund (SPEC-291 Nr. 2).
type TLSVerstoss struct{ Schluessel, Grund string }

// LadeTLS ist ladeTLS mit den Namen und Pfaden des Aufrufers (SPEC-291 Nr. 2); es kennt keinen Namen von sich aus.
func LadeTLS(certSchluessel, keySchluessel, certPfad, keyPfad string, jetzt time.Time) (*tls.Config, []TLSVerstoss) {
	cfg, vs := ladeTLS(tlsDateien{certSchluessel, keySchluessel, certPfad, keyPfad}, jetzt)
	var aus []TLSVerstoss
	for _, v := range vs {
		aus = append(aus, TLSVerstoss{v.schluessel, v.grund})
	}
	return cfg, aus
}

// LadeKonfigurationZu ist LadeKonfiguration zu einer festen Zeit des Starts (SPEC-291 Nr. 3).
func LadeKonfigurationZu(environ []string, jetzt time.Time) (Konfiguration, error) {
	return ladeKonfigurationZu(environ, jetzt)
}

// ListenTLSPraefix ist der Anfang der Schlüssel `LISTEN_TLS_*`, den die Konfiguration aus den zwei Namen bildet.
func ListenTLSPraefix() string { return listenTLSPraefix() }

// SatzVersion4 ist der Satz der Codecs der Version 4.0 des Mock.
func SatzVersion4() outbound.CodecSatz { return satzVersion4() }

// SatzTabelle ist die Tabelle Version → Satz des Root (SPEC-139 Nr. 11 b).
func SatzTabelle() map[string]outbound.CodecSatz { return satzTabelle() }

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
	return build(ctx, k, log, io.Discard, b)
}

// BuildMitSenke ist Build mit gegebenen Bauteilen und der Senke des Adapters des Prozesses (SPEC-139 Nr. 13).
func BuildMitSenke(ctx context.Context, k Konfiguration, log *slog.Logger, senke io.Writer, b Bauteile) (*Handle, error) {
	return build(ctx, k, log, senke, b)
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
