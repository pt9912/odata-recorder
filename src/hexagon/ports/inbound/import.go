package inbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ImportAnfrage ist die Anfrage von POST /__admin/imports (SPEC-047, SPEC-127 Nr. 3): leere
// ConflictPolicy und Target gelten als reject und new_baseline.
type ImportAnfrage struct {
	ImportID         string
	ServiceID        string
	SourceSessionIDs []string
	MetadataHash     string
	ConflictPolicy   string
	Target           string
}

// ImportErgebnis ist der veröffentlichte Import; Neu sagt, ob dieser Aufruf die Baseline
// veröffentlicht hat. Import.ReportJSON ist der Body der Antwort (SPEC-127 Nr. 14).
type ImportErgebnis struct {
	Import model.Import
	Neu    bool
}

// ImportUseCase überführt die Exchanges einer beendeten Session in eine Baseline (SPEC-127).
// Jeder Fehler mit Code ist ein *Fehler; model.ErrUngueltig und die Kontextfehler kehren ohne
// Code zurück.
type ImportUseCase interface {
	// Import liest, führt zusammen, belegt die Coverage und veröffentlicht atomar; eine
	// Wiederholung mit gleicher import_id und gleichem Inhalt liefert den gespeicherten Import.
	Import(ctx context.Context, a ImportAnfrage) (ImportErgebnis, error)
	// Get liest einen Import-Eintrag.
	Get(ctx context.Context, importID string) (model.Import, error)
}
