package sqlite

import (
	"context"
	"database/sql"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.WorkspaceTransactor = (*DB)(nil)

// WithWorkspaceTransaction ruft f mit Repositories, die auf einer Schreibtransaktion
// arbeiten (SPEC-108 Nr. 1): Erfolg bestätigt, ein Fehler oder eine Panik von f rollt
// zurück. Die Repositories gelten nur innerhalb von f.
func (d *DB) WithWorkspaceTransaction(ctx context.Context, f func(outbound.WorkspaceStores) error) error {
	return d.inTx(ctx, func(c *sql.Conn) error {
		z := zugriff{db: d, conn: c}
		return f(outbound.WorkspaceStores{
			Workspaces: &WorkspaceRepository{z: z},
			Entities:   &EntityRepository{z: z},
			Cursors:    &CursorRepository{z: z},
			Snapshots:  &SnapshotRepository{z: z},
		})
	})
}
