package hybrid

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Schlüssel und Ergebnisse des Spans `odata.hybrid.fallback` (SPEC-321 Nr. 4); die Art ist der Wert der Kennzeichnung
// (wertReplay, wertUpstream).
const (
	attrArt         = "recorder.fallback_type"
	attrErgebnis    = "recorder.fallback_result"
	ergebnisAntwort = "antwort"
	ergebnisFehler  = "fehler"
	ergebnisAbbruch = "abgebrochen"
)

// ergebnisVon bildet das Ergebnis einer Stufe aus ihrem Fehler (SPEC-321 Nr. 4 b): kein Fehler ist eine Antwort, ein
// Kontextfehler (auch neben einem Fehler mit Code) ein Abbruch, jeder andere Fehler ein Fehler.
func ergebnisVon(err error) string {
	switch {
	case err == nil:
		return ergebnisAntwort
	case kontextFehler(err):
		return ergebnisAbbruch
	}
	return ergebnisFehler
}

// stufe führt eine Stufe des Fallbacks im Span aus (SPEC-321 Nr. 3 und 5, Aufrufstellen B1 bis B4): der Versuch mit
// dem Code verworfen ist mit der Entscheidung verworfen und wird gezählt, der Span beginnt danach und vor der Prüfung
// von ctx; lauf ruft die Stufe mit dem ctx des Aufrufs. Das ctx von StartSpan verwirft der Dienst (Nr. 3 d). Ein
// beendetes ctx ruft lauf nicht und endet den Span mit dem Ergebnis abgebrochen; der Fehler ist dann ctx.Err()
// (SPEC-284 Nr. 3 e).
func (v *Verteiler) stufe(ctx context.Context, art string, verworfen model.Code, lauf func(context.Context) (inbound.Antwort, error)) (inbound.Antwort, error) {
	v.telemetrie.CountDiscarded(verworfen)
	_, span := v.telemetrie.StartSpan(ctx, support.SpanHybridFallback, support.Str(attrArt, art))
	var antwort inbound.Antwort
	err, ergebnis := ctx.Err(), ergebnisAbbruch
	if err == nil {
		antwort, err = lauf(ctx)
		ergebnis = ergebnisVon(err)
	}
	span.Set(support.Str(attrErgebnis, ergebnis))
	span.End()
	return antwort, err
}
