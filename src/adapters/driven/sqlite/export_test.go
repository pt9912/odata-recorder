package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
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

// DiagTransaktionAktiv sagt, ob die Verbindung der Stores eine offene Transaktion trägt.
// Der Treiber bietet kein sqlite3_get_autocommit; die Probe führt BEGIN aus: scheitert es,
// läuft eine Transaktion; gelingt es, war keine offen, und ein ROLLBACK stellt den Zustand her.
func DiagTransaktionAktiv(ctx context.Context, s outbound.WorkspaceStores) bool {
	c := s.Workspaces.(*WorkspaceRepository).z.conn
	if _, err := c.ExecContext(ctx, "BEGIN"); err != nil {
		return true
	}
	_, _ = c.ExecContext(ctx, "ROLLBACK")
	return false
}
