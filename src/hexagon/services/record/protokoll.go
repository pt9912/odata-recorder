package record

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// protokolliere schreibt einen Eintrag der Stufe ERROR mit festem Text und den Attributen aus
// SPEC-121 Nr. 12; kein Attribut trägt einen Fehlertext, Pfad, Query, Header oder Body.
func (x *austausch) protokolliere(ctx context.Context, text, ereignis string, c model.Code, extra ...slog.Attr) {
	attrs := append([]slog.Attr{
		slog.String("mode", "record"),
		slog.String("ereignis", ereignis),
		slog.String("error_code", c.String()),
		slog.String("service_id", x.dienst.ID),
		slog.String("session_id", x.e.SessionID),
		slog.String("exchange_id", x.e.ID),
		slog.String("correlation_id", x.anfrage.KorrelationsID),
	}, extra...)
	x.r.k.Log.LogAttrs(ctx, slog.LevelError, text, attrs...)
}

// protokolliereCapture schreibt je Ergebnis.Fehler einen Eintrag capture (SPEC-121 Nr. 12).
func (x *austausch) protokolliereCapture(ctx context.Context, ergebnisse ...capture.Ergebnis) {
	for _, e := range ergebnisse {
		if e.Fehler == nil {
			continue
		}
		richtung := "request"
		if e.Richtung == capture.Response {
			richtung = "response"
		}
		x.protokolliere(ctx, "record: Aufzeichnung eines Bodys verworfen", "capture", e.Fehler.Code(),
			slog.String("richtung", richtung), slog.String("grund", string(e.Fehler.Grund)))
	}
}

// protokolliereAblage schreibt den Eintrag ablage eines nicht angehängten Exchange: der Code nach
// SPEC-109 Nr. 5 mit dem Gegenstand Sonstiges, ohne Code dort ODR-E9001 (SPEC-121 Nr. 12).
func (x *austausch) protokolliereAblage(ctx context.Context, err error) {
	c, ok := outbound.Abbilden(err, outbound.GegenstandSonstiges)
	if !ok {
		c = model.CodeE9001
	}
	x.protokolliere(ctx, "record: Exchange nicht angehängt", "ablage", c)
}

// fehlerOhneAntwort bildet den Fehler eines Exchange ohne Antwort in der Rangfolge von SPEC-121
// Nr. 11 und protokolliert ihn, wenn er einen Code trägt.
func (x *austausch) fehlerOhneAntwort(ctx context.Context, err error) error {
	if cerr := x.client(); cerr != nil {
		return cerr
	}
	c, status, grund := model.CodeE4001, 0, ""
	switch {
	case errors.Is(x.l.ursache(), errDrain):
		status, grund = 504, "drain"
	case errors.Is(err, outbound.ErrUpstreamZeit):
		status, grund = 504, "zeit"
	case errors.Is(err, outbound.ErrUpstreamNichtErreichbar):
		status, grund = 502, "nicht_erreichbar"
	default:
		c, status, grund = model.CodeE9001, 500, "intern"
	}
	x.protokolliere(ctx, "record: Upstream ohne Antwort", "upstream", c, slog.String("grund", grund))
	return &inbound.Fehler{Code: c, Status: status, Ursache: err}
}
