package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// querier ist eine Datenbank oder eine Verbindung: beide führen Anweisungen und Abfragen aus.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var (
	_ querier = (*sql.DB)(nil)
	_ querier = (*sql.Conn)(nil)
	_ querier = (*txVerbindung)(nil)
	_ querier = zugriff{}
)

// txVerbindung ist die Verbindung der offenen Transaktion von WithWorkspaceTransaction
// (SPEC-108 Nr. 1, Kontext in der Transaktion). Sie prüft den Kontext, bevor ein Statement
// beginnt, und führt das Statement ohne den Abbruch des Aufrufs aus: der Treiber ruft bei
// einem Abbruch sqlite3_interrupt, und SQLite rollt dann die ganze Transaktion zurück. Ein
// Fehler beim Zurücknehmen oder beim Abschluss eines Savepoints macht die Transaktion
// unbrauchbar (unbrauchbar ist der Fehler).
type txVerbindung struct {
	// conn ist die *sql.Conn der Transaktion; ein Test tauscht sie gegen einen Wrapper, der ein
	// Steuerstatement scheitern lässt (export_test.go).
	conn        querier
	unbrauchbar error
}

func (v *txVerbindung) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return v.conn.ExecContext(context.WithoutCancel(ctx), query, args...)
}

func (v *txVerbindung) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return v.conn.QueryContext(context.WithoutCancel(ctx), query, args...)
}

// QueryRowContext reicht einen beendeten Kontext an die Verbindung weiter: database/sql liefert
// seinen Fehler aus Row.Scan, ohne den Treiber zu rufen.
func (v *txVerbindung) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if ctx.Err() != nil {
		return v.conn.QueryRowContext(ctx, query, args...)
	}
	return v.conn.QueryRowContext(context.WithoutCancel(ctx), query, args...)
}

// steuere führt ein Statement des Savepoints einer Operation ohne den Kontext des Aufrufs aus;
// scheitert es, ist die Transaktion unbrauchbar.
func (v *txVerbindung) steuere(ctx context.Context, query string) error {
	_, err := v.conn.ExecContext(context.WithoutCancel(ctx), query)
	if err != nil {
		err = klassifiziere(err)
		if v.unbrauchbar == nil {
			v.unbrauchbar = err
		}
	}
	return err
}

// zugriff ist der Weg eines Repositories zur Datenbank: allein auf der Datenbank (conn ist
// nil, jede schreibende Operation ist eine eigene Transaktion) oder auf der Verbindung einer
// laufenden Transaktion (WithWorkspaceTransaction).
type zugriff struct {
	db   *DB
	conn *txVerbindung
}

// zugriff selbst erfüllt querier: Abfragen laufen auf der Verbindung der Transaktion oder auf
// der Datenbank.
func (z zugriff) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if z.conn != nil {
		return z.conn.ExecContext(ctx, query, args...)
	}
	return z.db.sql.ExecContext(ctx, query, args...)
}

func (z zugriff) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if z.conn != nil {
		return z.conn.QueryContext(ctx, query, args...)
	}
	return z.db.sql.QueryContext(ctx, query, args...)
}

func (z zugriff) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if z.conn != nil {
		return z.conn.QueryRowContext(ctx, query, args...)
	}
	return z.db.sql.QueryRowContext(ctx, query, args...)
}

// schreibe führt f atomar aus: in einer eigenen Transaktion oder, in der laufenden, in einem
// Savepoint, den ein Fehler von f zurücknimmt; die Transaktion bestätigt danach nur, was
// gelang. Scheitert das Zurücknehmen oder der Abschluss des Savepoints, liefert die Operation
// diesen Fehler und jede weitere schreibende Operation der Transaktion ebenso (SPEC-108 Nr. 1 f).
func (z zugriff) schreibe(ctx context.Context, f func(querier) error) error {
	v := z.conn
	if v == nil {
		return z.db.inTx(ctx, func(c *sql.Conn) error { return f(c) })
	}
	if v.unbrauchbar != nil {
		return v.unbrauchbar
	}
	if _, err := v.ExecContext(ctx, "SAVEPOINT operation"); err != nil {
		return klassifiziere(err)
	}
	if err := f(v); err != nil {
		scheitert := v.steuere(ctx, "ROLLBACK TO operation")
		if frei := v.steuere(ctx, "RELEASE operation"); scheitert == nil {
			scheitert = frei
		}
		if scheitert != nil {
			return scheitert
		}
		return err
	}
	return v.steuere(ctx, "RELEASE operation")
}

// sammle liest alle Zeilen; das Ergebnis ist nie nil.
func sammle[T any](rows *sql.Rows, lese func(zeile) (T, error)) ([]T, error) {
	defer func() { _ = rows.Close() }()
	out := []T{}
	for rows.Next() {
		v, err := lese(rows)
		if err != nil {
			return nil, klassifiziere(err)
		}
		out = append(out, v)
	}
	return out, klassifiziere(rows.Err())
}

func ungueltigerWert(format string, a ...any) error {
	return errors.Join(model.ErrUngueltig, fmt.Errorf(format, a...))
}

// dienstDesWorkspace liefert den Service eines Workspace; ein unbekannter ist ErrNichtGefunden.
func dienstDesWorkspace(ctx context.Context, q querier, id string) (string, error) {
	var dienst string
	err := q.QueryRowContext(ctx, "SELECT service_id FROM workspaces WHERE id = ?", id).Scan(&dienst)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: Workspace %s", outbound.ErrNichtGefunden, id)
	}
	return dienst, klassifiziere(err)
}

// pruefeWorkspace meldet ErrNichtGefunden, wenn id kein Workspace ist.
func pruefeWorkspace(ctx context.Context, q querier, id string) error {
	_, err := dienstDesWorkspace(ctx, q, id)
	return err
}

// leereListe prüft bei einer leeren Ergebnisliste, ob der Workspace existiert.
func leereListe[T any](ctx context.Context, q querier, liste []T, workspaceID string) ([]T, error) {
	if len(liste) > 0 {
		return liste, nil
	}
	if err := pruefeWorkspace(ctx, q, workspaceID); err != nil {
		return nil, err
	}
	return liste, nil
}
