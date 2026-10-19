package sqlite

import (
	"context"
	"database/sql"
)

// SQL gibt die Datenbank für die Tests zurück.
func (d *DB) SQL() *sql.DB { return d.sql }

// MigrateWith wendet ein anderes DDL an, damit Tests den Abbruch mitten in der Migration prüfen.
func (d *DB) MigrateWith(ctx context.Context, ddl string) error { return d.migrate(ctx, ddl) }
