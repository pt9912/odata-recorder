package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// abfrager ist eine Datenbank oder eine Verbindung, die eine Zeile liest.
type abfrager interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// gibtEs führt eine Abfrage der Form SELECT count(*) aus und sagt, ob sie etwas zählt.
func gibtEs(ctx context.Context, a abfrager, query string, args ...any) (bool, error) {
	var n int
	if err := a.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return false, klassifiziere(err)
	}
	return n > 0, nil
}

// SnapshotRepository erfüllt outbound.SnapshotRepository auf SQLite (SPEC-107, SPEC-108);
// Get und List sehen nur Snapshots der Art baseline, die Workspace-Operationen nur die der
// Art workspace.
type SnapshotRepository struct {
	z zugriff
}

var _ outbound.SnapshotRepository = (*SnapshotRepository)(nil)

// Snapshots liefert das Repository für Baselines.
func (d *DB) Snapshots() *SnapshotRepository { return &SnapshotRepository{z: zugriff{db: d}} }

const snapshotSpalten = "id, service_id, created_at, state_blob, state_hash, metadata_hash"

func leseSnapshot(row zeile) (model.BaselineSnapshot, error) {
	var b model.BaselineSnapshot
	var created string
	if err := row.Scan(&b.ID, &b.ServiceID, &created, &b.StateBlob, &b.StateHash, &b.MetadataHash); err != nil {
		return model.BaselineSnapshot{}, err
	}
	t, err := zeitAus(created)
	if err != nil {
		return model.BaselineSnapshot{}, fmt.Errorf("sqlite: created_at von %s: %w", b.ID, err)
	}
	b.CreatedAt = t
	return b, nil
}

// Get liest eine Baseline.
func (r *SnapshotRepository) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	row := r.z.QueryRowContext(ctx,
		"SELECT "+snapshotSpalten+" FROM snapshots WHERE id = ? AND snapshot_kind = 'baseline'", id)
	b, err := leseSnapshot(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.BaselineSnapshot{}, fmt.Errorf("%w: Baseline %s", outbound.ErrNichtGefunden, id)
	}
	return b, klassifiziere(err)
}

// List liefert die Baselines eines Service nach Erzeugungszeit und Kennung.
func (r *SnapshotRepository) List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error) {
	if serviceID == "" {
		return nil, errors.Join(model.ErrUngueltig, errors.New("leerer service_id"))
	}
	rows, err := r.z.QueryContext(ctx, "SELECT "+snapshotSpalten+
		" FROM snapshots WHERE service_id = ? AND snapshot_kind = 'baseline' ORDER BY created_at, id", serviceID)
	if err != nil {
		return nil, klassifiziere(err)
	}
	defer func() { _ = rows.Close() }()
	out := []model.BaselineSnapshot{}
	for rows.Next() {
		b, err := leseSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, klassifiziere(rows.Err())
}

// CoverageRepository erfüllt outbound.CoverageRepository auf SQLite (SPEC-107).
type CoverageRepository struct {
	db *DB
}

var _ outbound.CoverageRepository = (*CoverageRepository)(nil)

// Coverage liefert das Repository für die Coverage der Baselines.
func (d *DB) Coverage() *CoverageRepository { return &CoverageRepository{db: d} }

func leseCoverage(row zeile) (model.Coverage, error) {
	var c model.Coverage
	var status string
	err := row.Scan(&c.EntitySet, &status, &c.EvidenceJSON)
	c.Status = model.CoverageStatus(status)
	return c, err
}

// Get liest den Eintrag eines Entity-Set.
func (r *CoverageRepository) Get(ctx context.Context, baselineID, entitySet string) (model.Coverage, error) {
	if baselineID == "" || entitySet == "" {
		return model.Coverage{}, errors.Join(model.ErrUngueltig, errors.New("leere Baseline oder leeres entity_set"))
	}
	c, err := leseCoverage(r.db.sql.QueryRowContext(ctx,
		"SELECT entity_set, status, evidence_json FROM coverage WHERE baseline_snapshot_id = ? AND entity_set = ?",
		baselineID, entitySet))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Coverage{}, fmt.Errorf("%w: Coverage %s/%s", outbound.ErrNichtGefunden, baselineID, entitySet)
	}
	return c, klassifiziere(err)
}

// List liefert alle Einträge einer Baseline nach entity_set.
func (r *CoverageRepository) List(ctx context.Context, baselineID string) ([]model.Coverage, error) {
	if baselineID == "" {
		return nil, errors.Join(model.ErrUngueltig, errors.New("leere Baseline"))
	}
	if err := baselineBekannt(ctx, r.db.sql, baselineID); err != nil {
		return nil, err
	}
	rows, err := r.db.sql.QueryContext(ctx,
		"SELECT entity_set, status, evidence_json FROM coverage WHERE baseline_snapshot_id = ? ORDER BY entity_set", baselineID)
	if err != nil {
		return nil, klassifiziere(err)
	}
	defer func() { _ = rows.Close() }()
	out := []model.Coverage{}
	for rows.Next() {
		c, err := leseCoverage(rows)
		if err != nil {
			return nil, klassifiziere(err)
		}
		out = append(out, c)
	}
	return out, klassifiziere(rows.Err())
}

// baselineBekannt meldet ErrNichtGefunden, wenn id keine Baseline ist.
func baselineBekannt(ctx context.Context, a abfrager, id string) error {
	ok, err := gibtEs(ctx, a, "SELECT count(*) FROM snapshots WHERE id = ? AND snapshot_kind = 'baseline'", id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: Baseline %s", outbound.ErrNichtGefunden, id)
	}
	return nil
}
