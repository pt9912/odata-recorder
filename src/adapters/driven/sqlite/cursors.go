package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// CursorRepository erfüllt outbound.CursorRepository auf SQLite (SPEC-108).
type CursorRepository struct {
	z zugriff
}

var _ outbound.CursorRepository = (*CursorRepository)(nil)

// Cursors liefert das Repository für Replay-Cursors.
func (d *DB) Cursors() *CursorRepository { return &CursorRepository{z: zugriff{db: d}} }

const cursorSpalten = "workspace_id, replay_session_id, recording_session_id, service_id, strategy, position, consumed_json"

// Put schreibt oder ersetzt einen Cursor. Die Prüfungen folgen SPEC-108 Nr. 11.
func (r *CursorRepository) Put(ctx context.Context, c model.ReplayCursor) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return r.z.schreibe(ctx, func(q querier) error {
		wsDienst, err := dienstDesWorkspace(ctx, q, c.WorkspaceID)
		if err != nil {
			return err
		}
		var sessionDienst string
		err = q.QueryRowContext(ctx, "SELECT service_id FROM sessions WHERE id = ?", c.RecordingSessionID).Scan(&sessionDienst)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, c.RecordingSessionID)
		}
		if err != nil {
			return klassifiziere(err)
		}
		if c.ServiceID != wsDienst || c.ServiceID != sessionDienst {
			return fmt.Errorf("%w: Cursor trägt Service %s, Workspace %s, Session %s", outbound.ErrFremderService, c.ServiceID, wsDienst, sessionDienst)
		}
		_, err = q.ExecContext(ctx, "INSERT INTO replay_cursors("+cursorSpalten+") VALUES (?, ?, ?, ?, ?, ?, ?)"+
			" ON CONFLICT(workspace_id, replay_session_id, recording_session_id) DO UPDATE SET service_id = excluded.service_id,"+
			" strategy = excluded.strategy, position = excluded.position, consumed_json = excluded.consumed_json",
			c.WorkspaceID, c.ReplaySessionID, c.RecordingSessionID, c.ServiceID, string(c.Strategy), c.Position, c.ConsumedJSON())
		return klassifiziere(err)
	})
}

func leseCursor(row zeile) (model.ReplayCursor, error) {
	var c model.ReplayCursor
	var strategie, verbraucht string
	if err := row.Scan(&c.WorkspaceID, &c.ReplaySessionID, &c.RecordingSessionID, &c.ServiceID, &strategie, &c.Position, &verbraucht); err != nil {
		return model.ReplayCursor{}, err
	}
	c.Strategy = model.ReplayStrategy(strategie)
	if err := json.Unmarshal([]byte(verbraucht), &c.Consumed); err != nil {
		return model.ReplayCursor{}, fmt.Errorf("sqlite: consumed von %s/%s: %w", c.ReplaySessionID, c.RecordingSessionID, err)
	}
	return c, nil
}

// Get liest einen Cursor.
func (r *CursorRepository) Get(ctx context.Context, workspaceID, replaySessionID, recordingSessionID string) (model.ReplayCursor, error) {
	c, err := leseCursor(r.z.QueryRowContext(ctx, "SELECT "+cursorSpalten+
		" FROM replay_cursors WHERE workspace_id = ? AND replay_session_id = ? AND recording_session_id = ?",
		workspaceID, replaySessionID, recordingSessionID))
	if errors.Is(err, sql.ErrNoRows) {
		return model.ReplayCursor{}, fmt.Errorf("%w: Cursor %s/%s in Workspace %s", outbound.ErrNichtGefunden, replaySessionID, recordingSessionID, workspaceID)
	}
	return c, klassifiziere(err)
}

// List liefert die Cursors eines Workspace nach Replay- und Recording-Session.
func (r *CursorRepository) List(ctx context.Context, workspaceID string) ([]model.ReplayCursor, error) {
	q := r.z
	rows, err := q.QueryContext(ctx, "SELECT "+cursorSpalten+
		" FROM replay_cursors WHERE workspace_id = ? ORDER BY replay_session_id, recording_session_id", workspaceID)
	if err != nil {
		return nil, klassifiziere(err)
	}
	out, err := sammle(rows, leseCursor)
	if err != nil {
		return nil, err
	}
	return leereListe(ctx, q, out, workspaceID)
}

// ResetAll setzt alle Cursors des Workspace auf Position 0 und leeres consumed.
func (r *CursorRepository) ResetAll(ctx context.Context, workspaceID string) error {
	return r.z.schreibe(ctx, func(q querier) error {
		if err := pruefeWorkspace(ctx, q, workspaceID); err != nil {
			return err
		}
		_, err := q.ExecContext(ctx, "UPDATE replay_cursors SET position = 0, consumed_json = '[]' WHERE workspace_id = ?", workspaceID)
		return klassifiziere(err)
	})
}
