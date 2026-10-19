package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// ImportRepository erfüllt outbound.ImportRepository auf SQLite (SPEC-107).
type ImportRepository struct {
	db *DB
}

var _ outbound.ImportRepository = (*ImportRepository)(nil)

// Imports liefert das Repository für Imports und ihre Provenienz.
func (d *DB) Imports() *ImportRepository { return &ImportRepository{db: d} }

const importSpalten = "id, request_hash, baseline_snapshot_id, service_id, created_at, report_json"

func leseImport(row zeile) (model.Import, error) {
	var i model.Import
	var created string
	if err := row.Scan(&i.ID, &i.RequestHash, &i.BaselineSnapshotID, &i.ServiceID, &created, &i.ReportJSON); err != nil {
		return model.Import{}, err
	}
	t, err := zeitAus(created)
	if err != nil {
		return model.Import{}, fmt.Errorf("sqlite: created_at von %s: %w", i.ID, err)
	}
	i.CreatedAt = t
	return i, nil
}

// Publish schreibt Snapshot, Import, Baseline-Quellen und Coverage in einer Transaktion.
func (r *ImportRepository) Publish(ctx context.Context, p model.Publication) (model.Import, bool, error) {
	if err := p.Validate(); err != nil {
		return model.Import{}, false, err
	}
	var ergebnis model.Import
	neu := false
	err := r.db.inTx(ctx, func(c *sql.Conn) error {
		bekannt, err := leseImport(c.QueryRowContext(ctx, "SELECT "+importSpalten+
			" FROM imports WHERE service_id = ? AND request_hash = ?", p.Import.ServiceID, p.Import.RequestHash))
		if err == nil {
			ergebnis = bekannt
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return klassifiziere(err)
		}
		if err := pruefeKennungen(ctx, c, p); err != nil {
			return err
		}
		if err := schreibePublikation(ctx, c, p); err != nil {
			return err
		}
		ergebnis, neu = p.Import, true
		return nil
	})
	if err != nil {
		return model.Import{}, false, err
	}
	return ergebnis, neu, nil
}

// pruefeKennungen prüft in der Reihenfolge von SPEC-107 Nr. 8: Service, dann die
// Kennungen von Snapshot und Import.
func pruefeKennungen(ctx context.Context, c *sql.Conn, p model.Publication) error {
	dienst, err := gibtEs(ctx, c, "SELECT count(*) FROM services WHERE id = ?", p.Import.ServiceID)
	if err != nil {
		return err
	}
	if !dienst {
		return fmt.Errorf("%w: %s", outbound.ErrDienstUnbekannt, p.Import.ServiceID)
	}
	for _, k := range []struct{ abfrage, was, id string }{
		{"SELECT count(*) FROM snapshots WHERE id = ?", "Snapshot", p.Snapshot.ID},
		{"SELECT count(*) FROM imports WHERE id = ?", "Import", p.Import.ID},
	} {
		vorhanden, err := gibtEs(ctx, c, k.abfrage, k.id)
		if err != nil {
			return err
		}
		if vorhanden {
			return fmt.Errorf("%w: %s %s", outbound.ErrBereitsVorhanden, k.was, k.id)
		}
	}
	return nil
}

// schreibePublikation schreibt Snapshot und Import vor den Quellen und der Coverage; die
// Prüfung der Quell-Sessions geschieht beim Schreiben, die Transaktion nimmt das Bisherige zurück.
func schreibePublikation(ctx context.Context, c *sql.Conn, p model.Publication) error {
	s, i := p.Snapshot, p.Import
	if _, err := c.ExecContext(ctx, "INSERT INTO snapshots("+snapshotSpalten+", snapshot_kind) VALUES (?, ?, ?, ?, ?, ?, 'baseline')",
		s.ID, s.ServiceID, zeitText(s.CreatedAt), s.StateBlob, s.StateHash, s.MetadataHash); err != nil {
		return klassifiziere(err)
	}
	if _, err := c.ExecContext(ctx, "INSERT INTO imports("+importSpalten+") VALUES (?, ?, ?, ?, ?, ?)",
		i.ID, i.RequestHash, i.BaselineSnapshotID, i.ServiceID, zeitText(i.CreatedAt), i.ReportJSON); err != nil {
		return klassifiziere(err)
	}
	for _, sessionID := range p.SourceSessionIDs {
		if err := schreibeQuelle(ctx, c, s, sessionID); err != nil {
			return err
		}
	}
	for _, cov := range p.Coverage {
		if _, err := c.ExecContext(ctx,
			"INSERT INTO coverage(baseline_snapshot_id, baseline_kind, entity_set, status, evidence_json) VALUES (?, 'baseline', ?, ?, ?)",
			s.ID, cov.EntitySet, string(cov.Status), cov.EvidenceJSON); err != nil {
			return klassifiziere(err)
		}
	}
	return nil
}

func schreibeQuelle(ctx context.Context, c *sql.Conn, s model.BaselineSnapshot, sessionID string) error {
	var dienst string
	err := c.QueryRowContext(ctx, "SELECT service_id FROM sessions WHERE id = ?", sessionID).Scan(&dienst)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, sessionID)
	case err != nil:
		return klassifiziere(err)
	case dienst != s.ServiceID:
		return errors.Join(model.ErrUngueltig, fmt.Errorf("Session %s gehört zu Service %s, nicht zu %s", sessionID, dienst, s.ServiceID))
	}
	_, err = c.ExecContext(ctx,
		"INSERT INTO baseline_sources(baseline_snapshot_id, baseline_kind, session_id) VALUES (?, 'baseline', ?)", s.ID, sessionID)
	return klassifiziere(err)
}

// Get liest einen Import-Eintrag.
func (r *ImportRepository) Get(ctx context.Context, id string) (model.Import, error) {
	i, err := leseImport(r.db.sql.QueryRowContext(ctx, "SELECT "+importSpalten+" FROM imports WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Import{}, fmt.Errorf("%w: Import %s", outbound.ErrNichtGefunden, id)
	}
	return i, klassifiziere(err)
}

// BaselinesOf liefert die Kennungen der Baselines, die aus der Session hervorgingen.
func (r *ImportRepository) BaselinesOf(ctx context.Context, sessionID string) ([]string, error) {
	bekannt, err := gibtEs(ctx, r.db.sql, "SELECT count(*) FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return nil, err
	}
	if !bekannt {
		return nil, fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, sessionID)
	}
	return r.kennungen(ctx, "SELECT baseline_snapshot_id FROM baseline_sources WHERE session_id = ? ORDER BY 1", sessionID)
}

// Sources liefert die Quell-Sessions einer Baseline.
func (r *ImportRepository) Sources(ctx context.Context, baselineID string) ([]string, error) {
	if err := baselineBekannt(ctx, r.db.sql, baselineID); err != nil {
		return nil, err
	}
	return r.kennungen(ctx, "SELECT session_id FROM baseline_sources WHERE baseline_snapshot_id = ? ORDER BY 1", baselineID)
}

func (r *ImportRepository) kennungen(ctx context.Context, query, arg string) ([]string, error) {
	rows, err := r.db.sql.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, klassifiziere(err)
	}
	defer func() { _ = rows.Close() }()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, klassifiziere(err)
		}
		out = append(out, id)
	}
	return out, klassifiziere(rows.Err())
}
