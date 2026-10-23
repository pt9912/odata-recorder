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
		c := model.CodeE9001
		var f *inbound.Fehler
		if errors.As(err, &f) {
			c = f.Code
		}
		attrs = append(attrs, slog.String("ereignis", "bereinigung_fehlgeschlagen"), slog.String("error_code", c.String()))
		v.k.Log.LogAttrs(ctx, slog.LevelError, "session: Bereinigung nach dem Löschen fehlgeschlagen", attrs...)
	case err == nil && b.Zurueckgestellt:
		attrs = append(attrs, slog.String("ereignis", "bereinigung_zurueckgestellt"))
		v.k.Log.LogAttrs(ctx, slog.LevelInfo, "session: Bereinigung nach dem Löschen zurückgestellt", attrs...)
	}
}
