package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// SessionRepository erfüllt outbound.SessionRepository auf SQLite (SPEC-105).
type SessionRepository struct {
	db *DB
}

var _ outbound.SessionRepository = (*SessionRepository)(nil)

// Sessions liefert das Repository für Sessions.
func (d *DB) Sessions() *SessionRepository { return &SessionRepository{db: d} }

func zeitText(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func zeitAus(s string) (time.Time, error) { return time.Parse(time.RFC3339Nano, s) }

// Create legt eine Session im Zustand open an.
func (r *SessionRepository) Create(ctx context.Context, s model.Session) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return r.db.inTx(ctx, func(c *sql.Conn) error {
		_, err := c.ExecContext(ctx,
			"INSERT INTO sessions(id, service_id, created_at, status) VALUES (?, ?, ?, ?)",
			s.ID, s.ServiceID, zeitText(s.CreatedAt), string(s.Status))
		switch {
		case err == nil:
			return nil
		case istEindeutigkeit(err):
			return fmt.Errorf("%w: Session %s", outbound.ErrBereitsVorhanden, s.ID)
		case istFremdschluessel(err):
			return fmt.Errorf("%w: %s", outbound.ErrDienstUnbekannt, s.ServiceID)
		}
		return klassifiziere(err)
	})
}

// Get liest eine Session.
func (r *SessionRepository) Get(ctx context.Context, id string) (model.Session, error) {
	row := r.db.sql.QueryRowContext(ctx,
		"SELECT id, service_id, created_at, status FROM sessions WHERE id = ?", id)
	s, err := leseSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Session{}, fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, id)
	}
	return s, err
}

func leseSession(row zeile) (model.Session, error) {
	var s model.Session
	var created, status string
	if err := row.Scan(&s.ID, &s.ServiceID, &created, &status); err != nil {
		return model.Session{}, err
	}
	t, err := zeitAus(created)
	if err != nil {
		return model.Session{}, fmt.Errorf("sqlite: created_at von %s: %w", s.ID, err)
	}
	s.CreatedAt, s.Status = t, model.SessionStatus(status)
	return s, nil
}

// SetStatus wechselt von open nach complete oder incomplete.
func (r *SessionRepository) SetStatus(ctx context.Context, id string, status model.SessionStatus) error {
	return r.db.inTx(ctx, func(c *sql.Conn) error {
		var aktuell string
		err := c.QueryRowContext(ctx, "SELECT status FROM sessions WHERE id = ?", id).Scan(&aktuell)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, id)
		}
		if err != nil {
			return klassifiziere(err)
		}
		if !model.SessionStatus(aktuell).KannWechselnNach(status) {
			return fmt.Errorf("%w: %s → %s", outbound.ErrUngueltigerUebergang, aktuell, status)
		}
		_, err = c.ExecContext(ctx, "UPDATE sessions SET status = ? WHERE id = ?", string(status), id)
		return klassifiziere(err)
	})
}

// List liefert die Sessions eines Service nach Erzeugungszeit und Kennung.
func (r *SessionRepository) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	if serviceID == "" {
		return nil, errors.Join(model.ErrUngueltig, errors.New("leerer service_id"))
	}
	return r.liste(ctx, "WHERE service_id = ?", serviceID)
}

// FindOpen liefert alle Sessions im Zustand open.
func (r *SessionRepository) FindOpen(ctx context.Context) ([]model.Session, error) {
	return r.liste(ctx, "WHERE status = ?", string(model.SessionOpen))
}

func (r *SessionRepository) liste(ctx context.Context, where string, arg any) ([]model.Session, error) {
	rows, err := r.db.sql.QueryContext(ctx,
		"SELECT id, service_id, created_at, status FROM sessions "+where+" ORDER BY created_at, id", arg)
	if err != nil {
		return nil, klassifiziere(err)
	}
	defer func() { _ = rows.Close() }()
	out := []model.Session{}
	for rows.Next() {
		s, err := leseSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, klassifiziere(rows.Err())
}

// Delete löscht eine Session mit ihren Exchanges und Cursors.
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	return r.db.inTx(ctx, func(c *sql.Conn) error {
		var n int
		if err := c.QueryRowContext(ctx, "SELECT count(*) FROM sessions WHERE id = ?", id).Scan(&n); err != nil {
			return klassifiziere(err)
		}
		if n == 0 {
			return fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, id)
		}
		if _, err := c.ExecContext(ctx, "DELETE FROM exchanges WHERE session_id = ?", id); err != nil {
			return klassifiziere(err)
		}
		_, err := c.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
		if istFremdschluessel(err) {
			return fmt.Errorf("%w: Session %s ist Quelle einer Baseline", outbound.ErrReferenziert, id)
		}
		return klassifiziere(err)
	})
}
