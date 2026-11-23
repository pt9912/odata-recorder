package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// importsPfad ist das Muster der Routen für Imports (SPEC-297 Nr. 1).
const importsPfad = "/__admin/imports"

// importHandler hält allein die beiden Ports der Routen (SPEC-297 Nr. 1, 11); der Handler von POST ruft von audit nur Append.
type importHandler struct {
	imports inbound.ImportUseCase
	audit   support.AuditLog
}

// importPortFehlt bildet den Fehler von ImportRouten bei einem nil-Port: model.ErrUngueltig, der Text nennt keinen Wert.
func importPortFehlt(was string) error {
	return errors.Join(model.ErrUngueltig, errors.New("admin: ImportRouten: "+was+" fehlt"))
}

// ImportRouten liefert die zwei Routen der Imports in der Folge von SPEC-297 Nr. 1: `POST /__admin/imports` und
// `GET /__admin/imports/{id}`. Die Funktion ruft keinen Port und startet keine Goroutine; sie liefert model.ErrUngueltig
// und keine Routen bei einem nil-Wert.
func ImportRouten(imports inbound.ImportUseCase, audit support.AuditLog) ([]Route, error) {
	switch {
	case imports == nil:
		return nil, importPortFehlt("imports")
	case audit == nil:
		return nil, importPortFehlt("audit")
	}
	h := importHandler{imports: imports, audit: audit}
	return []Route{
		{Methode: http.MethodPost, Muster: importsPfad, Handler: h.importiere},
		{Methode: http.MethodGet, Muster: importsPfad + "/{" + pfadID + "}", Handler: h.lies},
	}, nil
}

// importiere liest den Body (SPEC-297 Nr. 5) und ruft Import mit dem ctx des Handlers; 201, wenn der Aufruf die Baseline
// veröffentlicht hat, sonst 200 (Nr. 6). Ein Fehler des Dienstes geht unverändert und ohne Antwort zurück (Nr. 7). Steht die
// Antwort fest (der Report ist gültig), schreibt die Route genau einen Eintrag `import_veroeffentlicht` aus dem Ergebnis des
// Dienstes und der Korrelations-ID der Antwort, auch bei einem beendeten ctx (Nr. 9 d, 10); das Ergebnis von Append ändert nichts.
func (h importHandler) importiere(ctx context.Context, a Anfrage) (Antwort, error) {
	anfrage, err := leseImport(a.Body)
	if err != nil {
		return Antwort{}, err
	}
	e, err := h.imports.Import(ctx, anfrage)
	if err != nil {
		return Antwort{}, err
	}
	status := http.StatusOK
	if e.Neu {
		status = http.StatusCreated
	}
	antwort, err := reportAntwort(status, e.Import.ReportJSON)
	if err != nil {
		return Antwort{}, err
	}
	_ = h.audit.Append(support.Eintrag{
		Art: support.ArtImportVeroeffentlicht, ServiceID: e.Import.ServiceID, ImportID: e.Import.ID, KorrelationsID: a.KorrelationsID,
	})
	return antwort, nil
}

// lies ruft Get mit dem ctx des Handlers und dem Pfadwert unverändert (SPEC-297 Nr. 2, 4); Query, Header und Body
// bleiben ungelesen (Nr. 3). Ein Fehler des Dienstes geht unverändert und ohne Antwort zurück (Nr. 7).
func (h importHandler) lies(ctx context.Context, a Anfrage) (Antwort, error) {
	i, err := h.imports.Get(ctx, a.Pfadwerte[pfadID])
	if err != nil {
		return Antwort{}, err
	}
	return reportAntwort(http.StatusOK, i.ReportJSON)
}

// reportAntwort liefert den Report als Body, Byte für Byte (SPEC-297 Nr. 6). Ein Report, der kein gültiges JSON in UTF-8 ist,
// ist ein Fehler ohne Code (ODR-E9001) und keine Teilantwort; der Text nennt keinen Wert.
func reportAntwort(status int, report string) (Antwort, error) {
	if !utf8.ValidString(report) || !json.Valid([]byte(report)) {
		return Antwort{}, errors.New("admin: Report des Imports ist kein gültiges JSON")
	}
	return Antwort{Status: status, Body: []byte(report)}, nil
}
