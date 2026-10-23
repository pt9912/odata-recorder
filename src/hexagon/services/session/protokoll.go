package session

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// kontextFehler sagt, ob err einen Kontextfehler enthält, auch neben einem anderen Fehler.
func kontextFehler(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// protokolliere schreibt für das Ergebnis der Bereinigung nach dem Löschen von s höchstens einen
// Eintrag (SPEC-124 Nr. 9): bereinigung_fehlgeschlagen mit dem Code des Fehlers, bereinigung_zurueckgestellt
// bei nil und Zurueckgestellt, sonst keinen; ein Kontextfehler ergibt keinen. Kein Eintrag trägt
// einen Fehlertext.
func (v *Verwaltung) protokolliere(ctx context.Context, s model.Session, b inbound.Loeschbericht, err error) {
	attrs := []slog.Attr{slog.String("service_id", s.ServiceID), slog.String("session_id", s.ID)}
	switch {
	case err != nil && !kontextFehler(err):
		attrs = append(attrs, slog.String("ereignis", "bereinigung_fehlgeschlagen"), slog.String("error_code", fehlerCode(err).String()))
		v.k.Log.LogAttrs(ctx, slog.LevelError, "session: Bereinigung nach dem Löschen fehlgeschlagen", attrs...)
	case err == nil && b.Zurueckgestellt:
		attrs = append(attrs, slog.String("ereignis", "bereinigung_zurueckgestellt"))
		v.k.Log.LogAttrs(ctx, slog.LevelInfo, "session: Bereinigung nach dem Löschen zurückgestellt", attrs...)
	}
}

// fehlerCode ist der Code des Fehlers für das Protokoll: der eines *inbound.Fehler, sonst ODR-E9001.
func fehlerCode(err error) model.Code {
	var f *inbound.Fehler
	if errors.As(err, &f) {
		return f.Code
	}
	return model.CodeE9001
}

// protokolliereBeenden schreibt für die Session s, die CloseAll nicht beenden konnte, einen Eintrag
// beenden_fehlgeschlagen mit dem Code des Fehlers (SPEC-124 Nr. 9); ein Kontextfehler ergibt keinen.
// Kein Eintrag trägt einen Fehlertext.
func (v *Verwaltung) protokolliereBeenden(ctx context.Context, s model.Session, err error) {
	if kontextFehler(err) {
		return
	}
	v.k.Log.LogAttrs(ctx, slog.LevelError, "session: Beenden beim Herunterfahren fehlgeschlagen",
		slog.String("service_id", s.ServiceID), slog.String("session_id", s.ID),
		slog.String("ereignis", "beenden_fehlgeschlagen"), slog.String("error_code", fehlerCode(err).String()))
}
