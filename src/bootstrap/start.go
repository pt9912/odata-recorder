package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// exitFehler ist die Rückgabe von Start bei einem Fehler, der nicht der Abbruch von ctx ist.
const exitFehler = 1

// Start ist der Start-Befehl des Adapters cli (SPEC-139 Nr. 8, SPEC-138 Nr. 8): Konfiguration laden,
// bauen, hochfahren, auf Abbruch oder Ausfall des Listeners warten, herunterfahren. Meldungen und
// Log gehen nach stderr, nach stdout nichts; die Rückgabe ist 0 oder 1.
func Start(ctx context.Context, _, stderr io.Writer) int {
	return start(ctx, os.Environ(), stderr, standardBauteile())
}

func start(ctx context.Context, environ []string, stderr io.Writer, b bauteile) int {
	k, code := LadeOderMelde(environ, stderr)
	if code != exitOK {
		return code
	}
	if ctx.Err() != nil {
		return exitOK
	}
	log := slog.New(slog.NewTextHandler(stderr, nil))
	if k.Modus == modusRecord && !k.UpstreamTLSVerify {
		log.Warn("start: Zertifikatsprüfung zum Upstream abgeschaltet", "ereignis", "tls_pruefung_aus")
	}
	h, err := build(ctx, k, log, b)
	if err != nil {
		return abschluss(ctx, stderr, []error{err})
	}
	fehler, ausfall := betreibe(ctx, h, log)
	if err := h.Herunterfahren(ctx); err != nil {
		fehler = append(fehler, zerlege(err)...)
	}
	if ausfall == nil {
		ausfall = ausfallDanach(h)
	}
	if ausfall != nil {
		fehler = append(fehler, &inbound.Fehler{Code: model.CodeE9001, Status: 500, Ursache: ausfall})
	}
	return abschluss(ctx, stderr, fehler)
}

// betreibe fährt h hoch und wartet auf den Abbruch von ctx oder den Ausfall des Listeners; es liefert den
// Fehler des Hochfahrens und den Fehler von Serve, wenn er vor dem Abbruch eintrat.
func betreibe(ctx context.Context, h *Handle, log *slog.Logger) (fehler []error, ausfall error) {
	if err := h.Hochfahren(ctx); err != nil {
		return []error{err}, nil
	}
	log.Info("start: bereit", "mode", h.k.Modus, "listen", h.Adresse())
	select {
	case <-ctx.Done():
	case e, offen := <-h.Fehler():
		if offen {
			return nil, e
		}
		<-ctx.Done()
	}
	return nil, nil
}

// ausfallDanach liefert den Fehler von Serve, der erst mit oder nach dem Abbruch von ctx eintrat.
func ausfallDanach(h *Handle) error {
	select {
	case e, offen := <-h.Fehler():
		if offen {
			return e
		}
	default:
	}
	return nil
}

// zerlege liefert die Fehler, die errors.Join gesammelt hat, einzeln.
func zerlege(err error) []error {
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		return j.Unwrap()
	}
	return []error{err}
}

// abschluss schreibt je Fehler genau eine Zeile und liefert 1; ein Fehler, der der Abbruch von ctx
// ist, bleibt ohne Meldung und ohne Wirkung auf die Rückgabe (SPEC-138 Nr. 8 c, f).
func abschluss(ctx context.Context, stderr io.Writer, fehler []error) int {
	ergebnis := exitOK
	for _, e := range fehler {
		if ctx.Err() != nil && errors.Is(e, ctx.Err()) {
			continue
		}
		_, _ = fmt.Fprintln(stderr, meldung(e))
		ergebnis = exitFehler
	}
	return ergebnis
}

// meldung ist die Zeile zu einem Fehler: nie ein Wert der Konfiguration und nie der Text der Ursache.
func meldung(err error) string {
	var f *inbound.Fehler
	if errors.As(err, &f) {
		return f.Error()
	}
	var kf *KonfigFehler
	if errors.As(err, &kf) {
		return kf.Error()
	}
	code, ok := outbound.Abbilden(err, outbound.GegenstandSonstiges)
	if !ok {
		code = model.CodeE9001
	}
	return code.String() + ": " + code.Message()
}
