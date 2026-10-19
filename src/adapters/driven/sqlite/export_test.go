package sqlite

import (
	"context"
	"database/sql"
	"time"
)

// SQL gibt die Datenbank für die Tests zurück.
func (d *DB) SQL() *sql.DB { return d.sql }

// Klassifiziere macht klassifiziere für die Tests erreichbar.
func Klassifiziere(err error) error { return klassifiziere(err) }

// StelleWalSicherMit fährt die Wiederholung mit einem eigenen Versuch.
func (d *DB) StelleWalSicherMit(ctx context.Context, frist time.Duration, versuch func(context.Context) error) error {
	return stelleWalSicher(ctx, frist, versuch)
}

// MigrateWith wendet ein anderes DDL an, damit Tests den Abbruch mitten in der Migration prüfen.
func (d *DB) MigrateWith(ctx context.Context, ddl string) error { return d.migrate(ctx, ddl) }
