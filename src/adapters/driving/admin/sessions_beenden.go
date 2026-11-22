package admin

import (
	"context"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// beende ruft Beenden mit dem ctx des Handlers und dem Pfadwert unverändert (SPEC-292 Nr. 2, 4, 9 a) und wartet auf die
// Rückkehr; ein Fehler des Dienstes geht unverändert und ohne Antwort zurück (Nr. 8). Nach dem Erfolg schreibt die Route
// genau einen Eintrag `session_beendet` aus dem Ergebnis des Dienstes und der Korrelations-ID der Antwort, auch bei einem
// beendeten ctx (Nr. 9 c, 11); das Ergebnis von Append ändert nichts.
func (h sessionHandler) beende(ctx context.Context, a Anfrage) (Antwort, error) {
	s, err := h.sessions.Beenden(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{
		Art: support.ArtSessionBeendet, ServiceID: s.ServiceID, SessionID: s.ID, KorrelationsID: a.KorrelationsID,
	})
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(sessionAus(s))}, nil
}
