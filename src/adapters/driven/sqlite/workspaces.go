package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// WorkspaceRepository erfüllt outbound.WorkspaceRepository auf SQLite (SPEC-108).
type WorkspaceRepository struct {
	z zugriff
}

var _ outbound.WorkspaceRepository = (*WorkspaceRepository)(nil)

// Workspaces liefert das Repository für Workspaces.
func (d *DB) Workspaces() *WorkspaceRepository { return &WorkspaceRepository{z: zugriff{db: d}} }

const workspaceSpalten = "id, service_id, baseline_snapshot_id, revision, created_at"

// Create legt einen Workspace mit Revision 0 an. Die Prüfungen folgen SPEC-108 Nr. 2.
func (r *WorkspaceRepository) Create(ctx context.Context, w model.Workspace) error {
	if err := w.Validate(); err != nil {
		return err
	}
	return r.z.schreibe(ctx, func(q querier) error {
		vorhanden, err := gibtEs(ctx, q, "SELECT count(*) FROM workspaces WHERE id = ?", w.ID)
		if err != nil {
			return err
		}
		if vorhanden {
			return fmt.Errorf("%w: Workspace %s", outbound.ErrBereitsVorhanden, w.ID)
		}
		if err := pruefeDienst(ctx, q, w.ServiceID); err != nil {
			return err
		}
		if err := pruefeBaseline(ctx, q, w); err != nil {
			return err
		}
		_, err = q.ExecContext(ctx, "INSERT INTO workspaces(id, service_id, baseline_snapshot_id, baseline_kind, revision, created_at)"+
			" VALUES (?, ?, ?, 'baseline', 0, ?)", w.ID, w.ServiceID, w.BaselineSnapshotID, zeitText(w.CreatedAt))
		return klassifiziere(err)
	})
}

func pruefeDienst(ctx context.Context, q querier, id string) error {
	bekannt, err := gibtEs(ctx, q, "SELECT count(*) FROM services WHERE id = ?", id)
	if err != nil {
		return err
	}
	if !bekannt {
		return fmt.Errorf("%w: %s", outbound.ErrDienstUnbekannt, id)
	}
	return nil
}

func pruefeBaseline(ctx context.Context, q querier, w model.Workspace) error {
	var dienst, art string
	err := q.QueryRowContext(ctx, "SELECT service_id, snapshot_kind FROM snapshots WHERE id = ?", w.BaselineSnapshotID).Scan(&dienst, &art)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%w: Baseline %s", outbound.ErrNichtGefunden, w.BaselineSnapshotID)
	case err != nil:
		return klassifiziere(err)
	case art != "baseline":
		return fmt.Errorf("%w: %s ist ein Snapshot der Art %s", outbound.ErrFalscheArt, w.BaselineSnapshotID, art)
	case dienst != w.ServiceID:
		return fmt.Errorf("%w: Baseline %s gehört zu Service %s, nicht zu %s", outbound.ErrFremderService, w.BaselineSnapshotID, dienst, w.ServiceID)
	}
	return nil
}

func leseWorkspace(row zeile) (model.Workspace, error) {
	var w model.Workspace
	var created string
	if err := row.Scan(&w.ID, &w.ServiceID, &w.BaselineSnapshotID, &w.Revision, &created); err != nil {
		return model.Workspace{}, err
	}
	t, err := zeitAus(created)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("sqlite: created_at von %s: %w", w.ID, err)
	}
	w.CreatedAt = t
	return w, nil
}

// Get liest einen Workspace.
func (r *WorkspaceRepository) Get(ctx context.Context, id string) (model.Workspace, error) {
	w, err := leseWorkspace(r.z.QueryRowContext(ctx, "SELECT "+workspaceSpalten+" FROM workspaces WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Workspace{}, fmt.Errorf("%w: Workspace %s", outbound.ErrNichtGefunden, id)
	}
	return w, klassifiziere(err)
}

// List liefert die Workspaces eines Service nach Erzeugungszeit und Kennung.
func (r *WorkspaceRepository) List(ctx context.Context, serviceID string) ([]model.Workspace, error) {
	if serviceID == "" {
		return nil, ungueltigerWert("leerer service_id")
	}
	rows, err := r.z.QueryContext(ctx,
		"SELECT "+workspaceSpalten+" FROM workspaces WHERE service_id = ? ORDER BY created_at, id", serviceID)
	if err != nil {
		return nil, klassifiziere(err)
	}
	return sammle(rows, leseWorkspace)
}

// Advance erhöht die Revision von erwartet auf erwartet+1.
func (r *WorkspaceRepository) Advance(ctx context.Context, id string, erwartet int64) (int64, error) {
	if erwartet < 0 {
		return 0, ungueltigerWert("negative Revision %d", erwartet)
	}
	neu := erwartet + 1
	err := r.z.schreibe(ctx, func(q querier) error {
		var aktuell int64
		err := q.QueryRowContext(ctx, "SELECT revision FROM workspaces WHERE id = ?", id).Scan(&aktuell)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: Workspace %s", outbound.ErrNichtGefunden, id)
		}
		if err != nil {
			return klassifiziere(err)
		}
		if aktuell != erwartet {
			return fmt.Errorf("%w: Workspace %s steht auf %d, erwartet war %d", outbound.ErrRevisionVeraltet, id, aktuell, erwartet)
		}
		_, err = q.ExecContext(ctx, "UPDATE workspaces SET revision = ? WHERE id = ?", neu, id)
		return klassifiziere(err)
	})
	if err != nil {
		return 0, err
	}
	return neu, nil
}

// Delete löscht den Workspace mit seinen Workspace-Snapshots; Entity-State und Cursors
// folgen dem CASCADE des Schemas (SPEC-004).
func (r *WorkspaceRepository) Delete(ctx context.Context, id string) error {
	return r.z.schreibe(ctx, func(q querier) error {
		if err := pruefeWorkspace(ctx, q, id); err != nil {
			return err
		}
		if _, err := q.ExecContext(ctx, "DELETE FROM snapshots WHERE snapshot_kind = 'workspace' AND workspace_id = ?", id); err != nil {
			return klassifiziere(err)
		}
		_, err := q.ExecContext(ctx, "DELETE FROM workspaces WHERE id = ?", id)
		return klassifiziere(err)
	})
}
