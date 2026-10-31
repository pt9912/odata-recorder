package workspace

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

// protokolliere schreibt für das Ergebnis der Bereinigung nach dem Löschen des Workspace id höchstens einen
// Eintrag (SPEC-129 Nr. 13): bereinigung_fehlgeschlagen mit dem Code des Fehlers, bereinigung_zurueckgestellt
// bei nil und Zurueckgestellt, sonst keinen; ein Kontextfehler ergibt keinen. Kein Eintrag trägt einen Fehlertext.
func (v *Verwaltung) protokolliere(ctx context.Context, id string, b inbound.Loeschbericht, err error) {
	switch {
	case err != nil && !kontextFehler(err):
		v.k.Log.LogAttrs(ctx, slog.LevelError, "workspace: Bereinigung nach dem Löschen fehlgeschlagen",
			slog.String("ereignis", "bereinigung_fehlgeschlagen"), slog.String("workspace_id", id), slog.String("error_code", fehlerCode(err).String()))
	case err == nil && b.Zurueckgestellt:
		v.k.Log.LogAttrs(ctx, slog.LevelInfo, "workspace: Bereinigung nach dem Löschen zurückgestellt",
			slog.String("ereignis", "bereinigung_zurueckgestellt"), slog.String("workspace_id", id))
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
