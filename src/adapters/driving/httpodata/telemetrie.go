package httpodata

import (
	"context"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Die Schlüssel der drei Pflicht-Attribute des Gesamtspans (SPEC-294 Nr. 4).
const (
	attrMethode = "http.request.method"
	attrStatus  = "http.response.status_code"
	attrModus   = "recorder.mode"
)

// anfrageSpan hält Telemetrie und Gesamtspan der einen Anfrage (SPEC-294). Er lebt in einer Variable
// des Handlers und gehört genau einem Aufruf; zwischen Anfragen bleibt nichts (SPEC-136 Nr. 10).
type anfrageSpan struct {
	telemetrie support.Telemetry
	span       support.Span
	beendet    bool
}

// beginneSpan öffnet den Gesamtspan mit Methode und Modus (SPEC-294 Nr. 4 a, b) und liefert das ctx
// des Spans, das den Kontext des Requests ableitet (SPEC-294 Nr. 3 d).
func (s *Server) beginneSpan(ctx context.Context, r *http.Request) (context.Context, *anfrageSpan) {
	ctx, span := s.telemetrie.StartSpan(ctx, support.SpanGatewayRequest,
		support.Str(attrMethode, r.Method), support.Str(attrModus, s.modus))
	return ctx, &anfrageSpan{telemetrie: s.telemetrie, span: span}
}

// zaehle zählt das Fehlerobjekt, das der Adapter schreibt, mit seinem Code (SPEC-294 Nr. 5).
func (a *anfrageSpan) zaehle(code model.Code) {
	a.telemetrie.CountError(code)
}

// vorKopf setzt den Status und beendet den Span; es läuft unmittelbar vor WriteHeader
// (SPEC-294 Nr. 3 b, Nr. 4 c).
func (a *anfrageSpan) vorKopf(status int) {
	a.span.Set(support.Int(attrStatus, int64(status)))
	a.ende()
}

// ende beendet den Span genau einmal; ohne Antwort ruft es der Handler beim Verlassen (SPEC-294 Nr. 3 b).
func (a *anfrageSpan) ende() {
	if a.beendet {
		return
	}
	a.beendet = true
	a.span.End()
}
