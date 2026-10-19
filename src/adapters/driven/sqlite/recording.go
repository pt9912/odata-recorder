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

// RecordingRepository erfüllt outbound.RecordingRepository auf SQLite (SPEC-105).
type RecordingRepository struct {
	db *DB
}

var _ outbound.RecordingRepository = (*RecordingRepository)(nil)

// Recording liefert das Repository für Exchanges.
func (d *DB) Recording() *RecordingRepository { return &RecordingRepository{db: d} }

const exchangeSpalten = `id, session_id, sequence_no, method, path, raw_query, match_key, started_at, finished_at,
key_redacted, request_headers_json, request_body_state, request_blob, response_status, response_headers_json,
response_body_state, response_blob, capture_status, request_redaction_status, response_redaction_status,
body_integrity, replay_equivalence`

func nullText(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Append hängt einen Exchange an eine offene Session an.
func (r *RecordingRepository) Append(ctx context.Context, e model.Exchange) error {
	if err := e.Validate(); err != nil {
		return err
	}
	args, err := exchangeArgumente(e)
	if err != nil {
		return err
	}
	return r.db.inTx(ctx, func(c *sql.Conn) error {
		var status string
		err := c.QueryRowContext(ctx, "SELECT status FROM sessions WHERE id = ?", e.SessionID).Scan(&status)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, e.SessionID)
		}
		if err != nil {
			return klassifiziere(err)
		}
		if model.SessionStatus(status) != model.SessionOpen {
			return fmt.Errorf("%w: Session %s ist %s", outbound.ErrSessionGeschlossen, e.SessionID, status)
		}
		var vorhanden int
		if err := c.QueryRowContext(ctx, "SELECT count(*) FROM exchanges WHERE id = ?", e.ID).Scan(&vorhanden); err != nil {
			return klassifiziere(err)
		}
		if vorhanden > 0 {
			return fmt.Errorf("%w: Exchange %s", outbound.ErrBereitsVorhanden, e.ID)
		}
		var letzte int
		if err := c.QueryRowContext(ctx, "SELECT coalesce(max(sequence_no), 0) FROM exchanges WHERE session_id = ?", e.SessionID).Scan(&letzte); err != nil {
			return klassifiziere(err)
		}
		if e.SequenceNo != letzte+1 {
			return fmt.Errorf("%w: erwartet %d, erhalten %d", outbound.ErrUngueltigeSequenz, letzte+1, e.SequenceNo)
		}
		_, err = c.ExecContext(ctx, "INSERT INTO exchanges("+exchangeSpalten+") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", args...)
		return klassifiziere(err)
	})
}

func exchangeArgumente(e model.Exchange) ([]any, error) {
	if e.RequestHeaders == nil {
		e.RequestHeaders = map[string][]string{}
	}
	reqHeaders, err := json.Marshal(e.RequestHeaders)
	if err != nil {
		return nil, fmt.Errorf("sqlite: Request-Header: %w", err)
	}
	var respHeaders any
	if e.ResponseHeaders != nil {
		b, err := json.Marshal(e.ResponseHeaders)
		if err != nil {
			return nil, fmt.Errorf("sqlite: Response-Header: %w", err)
		}
		respHeaders = string(b)
	}
	var finished, status any
	if !e.FinishedAt.IsZero() {
		finished = zeitText(e.FinishedAt)
	}
	if e.ResponseStatus != 0 {
		status = e.ResponseStatus
	}
	keyRedacted := 0
	if e.KeyRedacted {
		keyRedacted = 1
	}
	return []any{e.ID, e.SessionID, e.SequenceNo, e.Method, e.Path, e.RawQuery, nullText(e.MatchKey),
		zeitText(e.StartedAt), finished, keyRedacted, string(reqHeaders), string(e.RequestBody),
		nullText(e.RequestBlob), status, respHeaders, string(e.ResponseBody), nullText(e.ResponseBlob),
		string(e.Capture), string(e.RequestRedact), string(e.ResponseRedact), string(e.Integrity),
		string(e.Equivalence)}, nil
}

// Get liest einen Exchange.
func (r *RecordingRepository) Get(ctx context.Context, id string) (model.Exchange, error) {
	row := r.db.sql.QueryRowContext(ctx, "SELECT "+exchangeSpalten+" FROM exchanges WHERE id = ?", id)
	e, err := leseExchange(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Exchange{}, fmt.Errorf("%w: Exchange %s", outbound.ErrNichtGefunden, id)
	}
	return e, err
}

// ListBySession liefert die Exchanges einer Session nach sequence_no.
func (r *RecordingRepository) ListBySession(ctx context.Context, sessionID string) ([]model.Exchange, error) {
	rows, err := r.db.sql.QueryContext(ctx, "SELECT "+exchangeSpalten+" FROM exchanges WHERE session_id = ? ORDER BY sequence_no", sessionID)
	if err != nil {
		return nil, klassifiziere(err)
	}
	defer func() { _ = rows.Close() }()
	out := []model.Exchange{}
	for rows.Next() {
		e, err := leseExchange(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, klassifiziere(err)
	}
	if len(out) > 0 {
		return out, nil
	}
	var n int
	if err := r.db.sql.QueryRowContext(ctx, "SELECT count(*) FROM sessions WHERE id = ?", sessionID).Scan(&n); err != nil {
		return nil, klassifiziere(err)
	}
	if n == 0 {
		return nil, fmt.Errorf("%w: Session %s", outbound.ErrNichtGefunden, sessionID)
	}
	return out, nil
}

func leseExchange(row zeile) (model.Exchange, error) {
	var e model.Exchange
	var matchKey, finished, reqBlob, respBlob, respHeaders sql.NullString
	var status sql.NullInt64
	var started, reqHeaders, reqBody, respBody, capture, reqRedact, respRedact, integrity, equivalence string
	var keyRedacted int
	err := row.Scan(&e.ID, &e.SessionID, &e.SequenceNo, &e.Method, &e.Path, &e.RawQuery, &matchKey, &started,
		&finished, &keyRedacted, &reqHeaders, &reqBody, &reqBlob, &status, &respHeaders, &respBody, &respBlob,
		&capture, &reqRedact, &respRedact, &integrity, &equivalence)
	if err != nil {
		return model.Exchange{}, err
	}
	e.MatchKey, e.RequestBlob, e.ResponseBlob = matchKey.String, reqBlob.String, respBlob.String
	e.KeyRedacted = keyRedacted == 1
	e.RequestBody, e.ResponseBody = model.BodyState(reqBody), model.BodyState(respBody)
	e.Capture, e.Integrity, e.Equivalence = model.CaptureStatus(capture), model.BodyIntegrity(integrity), model.ReplayEquivalence(equivalence)
	e.RequestRedact, e.ResponseRedact = model.RedactionStatus(reqRedact), model.RedactionStatus(respRedact)
	if status.Valid {
		e.ResponseStatus = int(status.Int64)
	}
	if e.StartedAt, err = zeitAus(started); err != nil {
		return model.Exchange{}, fmt.Errorf("sqlite: started_at von %s: %w", e.ID, err)
	}
	if finished.Valid {
		if e.FinishedAt, err = zeitAus(finished.String); err != nil {
			return model.Exchange{}, fmt.Errorf("sqlite: finished_at von %s: %w", e.ID, err)
		}
	}
	if err = json.Unmarshal([]byte(reqHeaders), &e.RequestHeaders); err != nil {
		return model.Exchange{}, fmt.Errorf("sqlite: Request-Header von %s: %w", e.ID, err)
	}
	if respHeaders.Valid {
		if err = json.Unmarshal([]byte(respHeaders.String), &e.ResponseHeaders); err != nil {
			return model.Exchange{}, fmt.Errorf("sqlite: Response-Header von %s: %w", e.ID, err)
		}
		if e.ResponseHeaders == nil {
			e.ResponseHeaders = map[string][]string{}
		}
	}
	return e, nil
}
