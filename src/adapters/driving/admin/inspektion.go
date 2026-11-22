package admin

import (
	"context"
	"errors"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Muster der drei Routen der Inspektion (SPEC-293 Nr. 1) und der Name des Pfadwerts von `exchanges`.
const (
	inspExchangesMuster = "/__admin/sessions/{id}/exchanges"
	inspCoverageMuster  = "/__admin/coverage"
	inspDiagnoseMuster  = "/__admin/diagnostics"
	inspPfadwertID      = "id"
)

// inspektionHandler hält allein den Port der Routen (SPEC-293 Nr. 1, 10).
type inspektionHandler struct {
	inspektion inbound.InspectionUseCase
}

// InspektionRouten liefert die drei Routen der Inspektion in der Folge von SPEC-293 Nr. 1: Exchanges einer Session, Coverage einer
// Baseline und Diagnose, alle mit `GET`. Die Funktion ruft keinen Port und startet keine Goroutine; sie liefert
// model.ErrUngueltig und keine Routen bei einem nil-Wert, und der Text nennt keinen Wert.
func InspektionRouten(inspektion inbound.InspectionUseCase) ([]Route, error) {
	if inspektion == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("admin: InspektionRouten: inspektion fehlt"))
	}
	h := inspektionHandler{inspektion: inspektion}
	return []Route{
		{Methode: http.MethodGet, Muster: inspExchangesMuster, Handler: h.exchanges},
		{Methode: http.MethodGet, Muster: inspCoverageMuster, Handler: h.coverage},
		{Methode: http.MethodGet, Muster: inspDiagnoseMuster, Handler: h.diagnose},
	}, nil
}

// exchanges liest limit und after (SPEC-293 Nr. 3, 4), ruft Exchanges mit dem ctx des Handlers und dem Pfadwert `{id}`
// byte-gleich und schreibt die Seite (Nr. 6). Ein Verstoß der Query ruft den Dienst nie; jeder Fehler des Dienstes geht
// unverändert und ohne Antwort zurück (Nr. 9).
func (h inspektionHandler) exchanges(ctx context.Context, a Anfrage) (Antwort, error) {
	seite, err := leseSeite(a.RawQuery)
	if err != nil {
		return Antwort{}, err
	}
	s, err := h.inspektion.Exchanges(ctx, a.Pfadwerte[inspPfadwertID], seite)
	if err != nil {
		return Antwort{}, err
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(exchangeSeiteAus(s))}, nil
}

// coverage liest baseline_id (SPEC-293 Nr. 3, 5), ruft Coverage mit dem ctx des Handlers und schreibt die Einträge (Nr. 7);
// eine Evidenz, die kein gültiges JSON in gültigem UTF-8 ist, ist ein Fehler ohne Code und keine Teilantwort.
func (h inspektionHandler) coverage(ctx context.Context, a Anfrage) (Antwort, error) {
	baselineID, err := leseBaselineID(a.RawQuery)
	if err != nil {
		return Antwort{}, err
	}
	es, err := h.inspektion.Coverage(ctx, baselineID)
	if err != nil {
		return Antwort{}, err
	}
	body, err := schreibeCoverage(es)
	if err != nil {
		return Antwort{}, err
	}
	return Antwort{Status: http.StatusOK, Body: body}, nil
}

// diagnose ruft Diagnose mit dem ctx des Handlers; die Query bleibt ungelesen (SPEC-293 Nr. 2, 8).
func (h inspektionHandler) diagnose(ctx context.Context, _ Anfrage) (Antwort, error) {
	d, err := h.inspektion.Diagnose(ctx)
	if err != nil {
		return Antwort{}, err
	}
	return Antwort{Status: http.StatusOK, Body: schreibeJSON(diagnoseAus(d))}, nil
}
