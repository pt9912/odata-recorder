package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// StorageMaintenance erfüllt outbound.StorageMaintenance auf SQLite (SPEC-108).
type StorageMaintenance struct {
	db *DB
}

var _ outbound.StorageMaintenance = (*StorageMaintenance)(nil)

// Maintenance liefert die Wartungsabfragen.
func (d *DB) Maintenance() *StorageMaintenance { return &StorageMaintenance{db: d} }

// Check liefert die Befunde von foreign_key_check und integrity_check als Texte.
func (m *StorageMaintenance) Check(ctx context.Context) ([]string, error) {
	rows, err := m.db.sql.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return nil, klassifiziere(err)
	}
	befunde, err := sammle(rows, func(z zeile) (string, error) {
		var tabelle, eltern sql.NullString
		var zeilenID, fk sql.NullInt64
		if err := z.Scan(&tabelle, &zeilenID, &eltern, &fk); err != nil {
			return "", err
		}
		return fmt.Sprintf("foreign_key_check: %s Zeile %d verweist auf %s", tabelle.String, zeilenID.Int64, eltern.String), nil
	})
	if err != nil {
		return nil, err
	}
	rows, err = m.db.sql.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return nil, klassifiziere(err)
	}
	meldungen, err := sammle(rows, func(z zeile) (string, error) {
		var s string
		return s, z.Scan(&s)
	})
	if err != nil {
		return nil, err
	}
	for _, s := range meldungen {
		if s != "ok" {
			befunde = append(befunde, "integrity_check: "+s)
		}
	}
	return befunde, nil
}

// ReferencedBlobs liefert die Blob-Referenzen der Exchanges, Snapshots und Service-Metadaten.
func (m *StorageMaintenance) ReferencedBlobs(ctx context.Context) ([]string, error) {
	rows, err := m.db.sql.QueryContext(ctx, `SELECT request_blob FROM exchanges WHERE request_blob IS NOT NULL
UNION SELECT response_blob FROM exchanges WHERE response_blob IS NOT NULL
UNION SELECT state_blob FROM snapshots
UNION SELECT metadata_blob FROM services WHERE metadata_blob IS NOT NULL
ORDER BY 1`)
	if err != nil {
		return nil, klassifiziere(err)
	}
	return sammle(rows, func(z zeile) (string, error) {
		var s string
		return s, z.Scan(&s)
	})
}
