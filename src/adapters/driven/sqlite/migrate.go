package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// migrate bringt die Datei in einer Transaktion auf Schemaversion 1. BEGIN IMMEDIATE
// ordnet gleichzeitige Öffnungen: die zweite sieht die fertige Version.
func (d *DB) migrate(ctx context.Context, ddl string) (err error) {
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
	done := false
	defer func() {
		if !done {
			_, _ = conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
		}
	}()
	var version int
	if err = conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return klassifiziere(err)
	}
	switch {
	case version == Schemaversion:
		return nil
	case version > Schemaversion:
		return fmt.Errorf("%w: Datei %d, bekannt %d", ErrSchemaNeuer, version, Schemaversion)
	}
	n, err := tabellenzahl(ctx, conn)
	if err != nil {
		return klassifiziere(err)
	}
	if n > 0 {
		return fmt.Errorf("%w: %d Tabellen ohne Schemaversion", ErrFremdeDatenbank, n)
	}
	if _, err = conn.ExecContext(ctx, ddl); err != nil {
		return klassifiziere(err)
	}
	if _, err = conn.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", Schemaversion)); err != nil {
		return klassifiziere(err)
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return klassifiziere(err)
	}
	done = true
	// Der Modus bleibt in der Datei; nur der Öffner, der migriert hat, schaltet ihn um.
	var modus string
	if err = conn.QueryRowContext(ctx, "PRAGMA journal_mode = wal").Scan(&modus); err != nil {
		return klassifiziere(err)
	}
	return nil
}

func tabellenzahl(ctx context.Context, conn *sql.Conn) (int, error) {
	var n int
	err := conn.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'").Scan(&n)
	return n, err
}
