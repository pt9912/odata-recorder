package sqlite

import (
	"context"
	"database/sql"
	"errors"

	sqlitedriver "modernc.org/sqlite"
)

// Erweiterte SQLite-Fehlercodes der Constraints.
const (
	codeForeignKey = 787
	// RESTRICT einer Fremdschlüssel-Aktion meldet SQLite als Trigger-Constraint.
	codeRestrict   = 1811
	codePrimaryKey = 1555
	codeUnique     = 2067
)

// zeile ist eine Zeile eines Abfrageergebnisses, einzeln oder aus einer Liste.
type zeile interface {
	Scan(dest ...any) error
}

func codeVon(err error) int {
	var se *sqlitedriver.Error
	if errors.As(err, &se) {
		return se.Code()
	}
	return 0
}

func istEindeutigkeit(err error) bool {
	c := codeVon(err)
	return c == codePrimaryKey || c == codeUnique
}

func istFremdschluessel(err error) bool {
	c := codeVon(err)
	return c == codeForeignKey || c == codeRestrict
}

// inTx führt f in einer Schreibtransaktion auf einer Verbindung aus: Erfolg bestätigt,
// jeder Fehler rollt zurück.
func (d *DB) inTx(ctx context.Context, f func(*sql.Conn) error) (err error) {
	conn, err := d.sql.Conn(ctx)
	if err != nil {
		return klassifiziere(err)
	}
	defer func() {
		if cerr := conn.Close(); err == nil {
			err = cerr
		}
	}()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return klassifiziere(err)
	}
	fertig := false
	defer func() {
		if !fertig {
			_, _ = conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
		}
	}()
	if err = f(conn); err != nil {
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return klassifiziere(err)
	}
	fertig = true
	return nil
}
