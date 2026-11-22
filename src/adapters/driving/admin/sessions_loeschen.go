package admin

import (
	"context"
	"net/http"
)

// loesche ruft Delete mit dem ctx des Handlers und dem Pfadwert unverändert (SPEC-292 Nr. 2, 4) und antwortet mit dem
// Bericht der Bereinigung (Nr. 10); ein Fehler des Dienstes geht unverändert und ohne Antwort zurück (Nr. 8). Die Route
// schreibt nie einen Eintrag: `session_geloescht` schreibt der Dienst (Nr. 11 c).
func (h sessionHandler) loesche(ctx context.Context, a Anfrage) (Antwort, error) {
	b, err := h.sessions.Delete(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	return Antwort{
		Status: http.StatusOK,
		Body:   schreibeJSON(loeschberichtJSON{BlobsRemoved: b.Entfernt, CleanupDeferred: b.Zurueckgestellt}),
	}, nil
}
