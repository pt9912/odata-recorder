package outbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// SnapshotRepository liest Baselines, die Snapshots der Art baseline (SPEC-107), und
// verwaltet Workspace-Snapshots (SPEC-108). Baselines entstehen über
// ImportRepository.Publish.
type SnapshotRepository interface {
	// Get liest eine Baseline.
	Get(ctx context.Context, id string) (model.BaselineSnapshot, error)
	// List liefert die Baselines eines Service nach Erzeugungszeit und Kennung.
	List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error)
	// CreateWorkspace legt einen Snapshot der Art workspace an.
	CreateWorkspace(ctx context.Context, s model.WorkspaceSnapshot) error
	// GetWorkspace liest einen Workspace-Snapshot des Workspace.
	GetWorkspace(ctx context.Context, workspaceID, id string) (model.WorkspaceSnapshot, error)
	// ListWorkspace liefert die Snapshots eines Workspace nach Erzeugungszeit und Kennung.
	ListWorkspace(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error)
}

// ImportRepository veröffentlicht Imports und liest ihre Provenienz (SPEC-107).
type ImportRepository interface {
	// Publish schreibt Snapshot, Import, Baseline-Quellen und Coverage in einer Transaktion.
	// Es liefert den Import und ob er neu ist; ein bekannter Anfrage-Hash liefert den
	// gespeicherten Import und schreibt nichts.
	Publish(ctx context.Context, p model.Publication) (model.Import, bool, error)
	// Get liest einen Import-Eintrag.
	Get(ctx context.Context, id string) (model.Import, error)
	// BaselinesOf liefert die Kennungen der Baselines, die aus der Session hervorgingen.
	BaselinesOf(ctx context.Context, sessionID string) ([]string, error)
	// Sources liefert die Quell-Sessions einer Baseline.
	Sources(ctx context.Context, baselineID string) ([]string, error)
}

// CoverageRepository liest die Coverage einer Baseline (SPEC-107); geschrieben wird sie
// nur über ImportRepository.Publish.
type CoverageRepository interface {
	// Get liest den Eintrag eines Entity-Set.
	Get(ctx context.Context, baselineID, entitySet string) (model.Coverage, error)
	// List liefert alle Einträge einer Baseline nach entity_set.
	List(ctx context.Context, baselineID string) ([]model.Coverage, error)
}
