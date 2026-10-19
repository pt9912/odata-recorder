package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// CreateWorkspace legt einen Snapshot der Art workspace an. Die Prüfungen folgen SPEC-108 Nr. 13.
func (r *SnapshotRepository) CreateWorkspace(ctx context.Context, s model.WorkspaceSnapshot) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return r.z.schreibe(ctx, func(q querier) error {
		dienst, err := dienstDesWorkspace(ctx, q, s.WorkspaceID)
		if err != nil {
			return err
		}
		if dienst != s.ServiceID {
			return fmt.Errorf("%w: Snapshot trägt Service %s, Workspace %s", outbound.ErrFremderService, s.ServiceID, dienst)
		}
		for _, k := range []struct {
			abfrage, was string
			args         []any
		}{
			{"SELECT count(*) FROM snapshots WHERE id = ?", "Snapshot " + s.ID, []any{s.ID}},
			{"SELECT count(*) FROM snapshots WHERE workspace_id = ? AND name = ?", "Name " + s.Name, []any{s.WorkspaceID, s.Name}},
		} {
			vorhanden, err := gibtEs(ctx, q, k.abfrage, k.args...)
			if err != nil {
				return err
			}
			if vorhanden {
				return fmt.Errorf("%w: %s", outbound.ErrBereitsVorhanden, k.was)
			}
		}
		_, err = q.ExecContext(ctx, "INSERT INTO snapshots(id, service_id, workspace_id, name, created_at, state_blob, state_hash, metadata_hash, snapshot_kind)"+
			" VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'workspace')",
			s.ID, s.ServiceID, s.WorkspaceID, s.Name, zeitText(s.CreatedAt), s.StateBlob, s.StateHash, s.MetadataHash)
		return klassifiziere(err)
	})
}

// workspaceSnapshotLesespalten liest workspace_id und name auch bei einer Baseline, die beide nicht trägt.
const workspaceSnapshotLesespalten = "id, service_id, coalesce(workspace_id, ''), coalesce(name, ''), created_at, state_blob, state_hash, metadata_hash"

// leseWorkspaceSnapshot liest die Spalten von workspaceSnapshotLesespalten, hinter den Zielen
// vorab, die eine Abfrage davor stellt.
func leseWorkspaceSnapshot(row zeile, vorab ...any) (model.WorkspaceSnapshot, error) {
	var s model.WorkspaceSnapshot
	var created string
	if err := row.Scan(append(vorab, &s.ID, &s.ServiceID, &s.WorkspaceID, &s.Name, &created, &s.StateBlob, &s.StateHash, &s.MetadataHash)...); err != nil {
		return model.WorkspaceSnapshot{}, err
	}
	t, err := zeitAus(created)
	if err != nil {
		return model.WorkspaceSnapshot{}, fmt.Errorf("sqlite: created_at von %s: %w", s.ID, err)
	}
	s.CreatedAt = t
	return s, nil
}

// GetWorkspace liest einen Workspace-Snapshot des Workspace.
func (r *SnapshotRepository) GetWorkspace(ctx context.Context, workspaceID, id string) (model.WorkspaceSnapshot, error) {
	var art string
	s, err := leseWorkspaceSnapshot(r.z.QueryRowContext(ctx,
		"SELECT snapshot_kind, "+workspaceSnapshotLesespalten+" FROM snapshots WHERE id = ?", id), &art)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return model.WorkspaceSnapshot{}, fmt.Errorf("%w: Snapshot %s", outbound.ErrNichtGefunden, id)
	case err != nil:
		return model.WorkspaceSnapshot{}, klassifiziere(err)
	case art == "baseline":
		return model.WorkspaceSnapshot{}, fmt.Errorf("%w: %s ist eine Baseline", outbound.ErrFalscheArt, id)
	case s.WorkspaceID != workspaceID:
		return model.WorkspaceSnapshot{}, fmt.Errorf("%w: Snapshot %s in Workspace %s", outbound.ErrNichtGefunden, id, workspaceID)
	}
	return s, nil
}

// ListWorkspace liefert die Snapshots eines Workspace nach Erzeugungszeit und Kennung.
func (r *SnapshotRepository) ListWorkspace(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error) {
	q := r.z
	rows, err := q.QueryContext(ctx, "SELECT "+workspaceSnapshotLesespalten+
		" FROM snapshots WHERE snapshot_kind = 'workspace' AND workspace_id = ? ORDER BY created_at, id", workspaceID)
	if err != nil {
		return nil, klassifiziere(err)
	}
	out, err := sammle(rows, func(z zeile) (model.WorkspaceSnapshot, error) { return leseWorkspaceSnapshot(z) })
	if err != nil {
		return nil, err
	}
	return leereListe(ctx, q, out, workspaceID)
}
