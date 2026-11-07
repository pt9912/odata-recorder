package inbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// SnapshotUseCase erzeugt benannte Workspace-Snapshots, die allein den Entity-State festhalten, und listet
// sie (SPEC-129 Nr. 21 bis 26). Jeder Fehler mit Code ist ein *Fehler; model.ErrUngueltig und die
// Kontextfehler kehren ohne Code zurück.
type SnapshotUseCase interface {
	// Create erfasst den Entity-State des Workspace unter dem Namen name (1 bis 128 Bytes); der Name ist im
	// Workspace eindeutig. Die Kennung und die Zeit vergibt der Dienst.
	Create(ctx context.Context, workspaceID, name string) (model.WorkspaceSnapshot, error)
	// List liefert die Workspace-Snapshots des Workspace in der Folge des Ports; ohne Snapshot eine leere Liste.
	List(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error)
}
