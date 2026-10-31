package inbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// WorkspaceUseCase verwaltet die Workspaces (SPEC-129): aus einer Baseline anlegen und deren
// Entity-State materialisieren, lesen, auflisten, löschen und auf die Baseline zurücksetzen. Jeder
// Fehler mit Code ist ein *Fehler; model.ErrUngueltig und die Kontextfehler kehren ohne Code zurück.
type WorkspaceUseCase interface {
	// Create legt einen Workspace aus der Baseline baselineID an; eine leere id vergibt der Dienst,
	// eine gelieferte muss der Form aus SPEC-129 Nr. 3 genügen. Der Service ist der der Baseline.
	Create(ctx context.Context, baselineID, id string) (model.Workspace, error)
	// Get liest einen Workspace; die id geht unverändert an den Port.
	Get(ctx context.Context, id string) (model.Workspace, error)
	// List liefert die Workspaces aller Services; ohne Workspace eine leere Liste.
	List(ctx context.Context) ([]model.Workspace, error)
	// Delete löscht den Workspace mit State, Cursors und Snapshots und ruft danach die Bereinigung
	// der Blobs; deren Bericht ist das Ergebnis, ihr Fehler ändert es nicht.
	Delete(ctx context.Context, id string) (Loeschbericht, error)
	// Reset setzt State und Cursors auf die Baseline zurück, erhöht die Revision um 1 und liefert
	// den Workspace mit der neuen Revision.
	Reset(ctx context.Context, id string) (model.Workspace, error)
}
