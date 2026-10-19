package outbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// WorkspaceRepository speichert Workspaces (SPEC-108). Jede Operation ist atomar.
type WorkspaceRepository interface {
	// Create legt einen Workspace mit Revision 0 an.
	Create(ctx context.Context, w model.Workspace) error
	// Get liest einen Workspace.
	Get(ctx context.Context, id string) (model.Workspace, error)
	// List liefert die Workspaces eines Service nach Erzeugungszeit und Kennung.
	List(ctx context.Context, serviceID string) ([]model.Workspace, error)
	// Advance erhöht die Revision von erwartet auf erwartet+1 und liefert die neue.
	Advance(ctx context.Context, id string, erwartet int64) (int64, error)
	// Delete löscht den Workspace mit Entity-State, Cursors und Workspace-Snapshots.
	Delete(ctx context.Context, id string) error
}

// EntityRepository speichert den Entity-State eines Workspace (SPEC-108).
type EntityRepository interface {
	// Put schreibt oder ersetzt einen Eintrag.
	Put(ctx context.Context, e model.EntityState) error
	// Get liest einen Eintrag, auch einen Tombstone.
	Get(ctx context.Context, workspaceID, entitySet, entityKey string) (model.EntityState, error)
	// List liefert alle Einträge eines Workspace nach entity_set und entity_key.
	List(ctx context.Context, workspaceID string) ([]model.EntityState, error)
	// ListSet liefert die Einträge einer Menge nach entity_key.
	ListSet(ctx context.Context, workspaceID, entitySet string) ([]model.EntityState, error)
	// ReplaceAll ersetzt den gesamten State durch die Einträge.
	ReplaceAll(ctx context.Context, workspaceID string, eintraege []model.EntityState) error
}

// CursorRepository speichert die Replay-Cursors eines Workspace (SPEC-108).
type CursorRepository interface {
	// Put schreibt oder ersetzt einen Cursor.
	Put(ctx context.Context, c model.ReplayCursor) error
	// Get liest einen Cursor.
	Get(ctx context.Context, workspaceID, replaySessionID, recordingSessionID string) (model.ReplayCursor, error)
	// List liefert die Cursors eines Workspace nach Replay- und Recording-Session.
	List(ctx context.Context, workspaceID string) ([]model.ReplayCursor, error)
	// ResetAll setzt alle Cursors des Workspace auf Position 0 und leeres consumed.
	ResetAll(ctx context.Context, workspaceID string) error
}

// StorageMaintenance liefert die Befunde und Referenzen, auf denen die Wartung aufsetzt
// (SPEC-108).
type StorageMaintenance interface {
	// Check liefert die Befunde von foreign_key_check und integrity_check als Texte.
	Check(ctx context.Context) ([]string, error)
	// ReferencedBlobs liefert die referenzierten Blob-Referenzen ohne Doppelte aufsteigend.
	ReferencedBlobs(ctx context.Context) ([]string, error)
}

// WorkspaceStores sind die Repositories einer Workspace-Transaktion; sie gelten nur
// innerhalb von f.
type WorkspaceStores struct {
	Workspaces WorkspaceRepository
	Entities   EntityRepository
	Cursors    CursorRepository
	Snapshots  SnapshotRepository
}

// WorkspaceTransactor führt Operationen auf mehreren Repositories in einer Transaktion
// aus (SPEC-108 Nr. 1).
type WorkspaceTransactor interface {
	// WithWorkspaceTransaction bestätigt, wenn f nil liefert, und rollt sonst zurück.
	WithWorkspaceTransaction(ctx context.Context, f func(WorkspaceStores) error) error
}
