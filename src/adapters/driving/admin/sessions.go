package admin

import (
	"context"
	"errors"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// sessionsPfad ist das Muster der Routen für Sessions (SPEC-292 Nr. 1).
const sessionsPfad = "/__admin/sessions"

// sessionHandler hält allein die beiden Ports der Routen (SPEC-292 Nr. 1, 12).
type sessionHandler struct {
	sessions inbound.SessionUseCase
	audit    support.AuditLog
}

// portFehlt bildet den Fehler von SessionRouten bei einem nil-Port: model.ErrUngueltig, der Text nennt keinen Wert.
func portFehlt(was string) error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: SessionRouten: "+was+" fehlt"))
}

// SessionRouten liefert die Routen der Sessions in der Folge von SPEC-292 Nr. 1: `GET` und `POST /__admin/sessions`.
// Die Funktion ruft keinen Port und startet keine Goroutine; sie liefert model.ErrUngueltig und keine Routen bei einem
// nil-Wert.
func SessionRouten(sessions inbound.SessionUseCase, audit support.AuditLog) ([]Route, error) {
	switch {
	case sessions == nil:
		return nil, portFehlt("sessions")
	case audit == nil:
		return nil, portFehlt("audit")
	}
	h := sessionHandler{sessions: sessions, audit: audit}
	return []Route{
		{Methode: http.MethodGet, Muster: sessionsPfad, Handler: h.liste},
		{Methode: http.MethodPost, Muster: sessionsPfad, Handler: h.anlege},
	}, nil
}

// liste ruft List mit dem ctx des Handlers; Query und Body bleiben ungelesen (SPEC-292 Nr. 2, 3, 6, 7 d). Jeder Fehler des
// Dienstes geht unverändert und ohne Antwort zurück (Nr. 8).
func (h sessionHandler) liste(ctx context.Context, _ Anfrage) (Antwort, error) {
	ss, err := h.sessions.List(ctx)
	if err != nil {
		return Antwort{}, err
	}
	out := make([]sessionJSON, len(ss))
	for i, s := range ss {
		out[i] = sessionAus(s)
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(sessionListeJSON{Sessions: out})}, nil
}

// anlege liest den Body (SPEC-292 Nr. 5), ruft Create mit dem ctx des Handlers und schreibt nach dem Erfolg genau einen
// Eintrag `session_angelegt` aus dem Ergebnis des Dienstes und der Korrelations-ID der Antwort (Nr. 11); das Ergebnis von
// Append ändert nichts.
func (h sessionHandler) anlege(ctx context.Context, a Anfrage) (Antwort, error) {
	serviceID, id, err := leseAnlegen(a.Body)
	if err != nil {
		return Antwort{}, err
	}
	s, err := h.sessions.Create(ctx, serviceID, id)
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{
		Art: support.ArtSessionAngelegt, ServiceID: s.ServiceID, SessionID: s.ID, KorrelationsID: a.KorrelationsID,
	})
	return Antwort{Status: http.StatusCreated, Body: schreibeJSON(sessionAus(s))}, nil
}
