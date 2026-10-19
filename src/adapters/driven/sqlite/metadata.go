package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// MetadataRepository erfüllt outbound.MetadataRepository auf SQLite (SPEC-107).
type MetadataRepository struct {
	db *DB
}

var _ outbound.MetadataRepository = (*MetadataRepository)(nil)

// Metadata liefert das Repository für Services.
func (d *DB) Metadata() *MetadataRepository { return &MetadataRepository{db: d} }

const serviceSpalten = "id, service_root, odata_version, metadata_blob, metadata_hash"

// Create legt einen Service an.
func (r *MetadataRepository) Create(ctx context.Context, s model.Service) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return r.db.inTx(ctx, func(c *sql.Conn) error {
		_, err := c.ExecContext(ctx, "INSERT INTO services("+serviceSpalten+") VALUES (?, ?, ?, ?, ?)",
			s.ID, s.ServiceRoot, s.ODataVersion, nullText(s.MetadataBlob), nullText(s.MetadataHash))
		if istEindeutigkeit(err) {
			return fmt.Errorf("%w: Service %s", outbound.ErrBereitsVorhanden, s.ID)
		}
		return klassifiziere(err)
	})
}

func leseService(row zeile) (model.Service, error) {
	var s model.Service
	var blob, hash sql.NullString
	if err := row.Scan(&s.ID, &s.ServiceRoot, &s.ODataVersion, &blob, &hash); err != nil {
		return model.Service{}, err
	}
	s.MetadataBlob, s.MetadataHash = blob.String, hash.String
	return s, nil
}

// Get liest einen Service.
func (r *MetadataRepository) Get(ctx context.Context, id string) (model.Service, error) {
	s, err := leseService(r.db.sql.QueryRowContext(ctx, "SELECT "+serviceSpalten+" FROM services WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Service{}, fmt.Errorf("%w: Service %s", outbound.ErrNichtGefunden, id)
	}
	return s, klassifiziere(err)
}

// List liefert alle Services nach Kennung.
func (r *MetadataRepository) List(ctx context.Context) ([]model.Service, error) {
	rows, err := r.db.sql.QueryContext(ctx, "SELECT "+serviceSpalten+" FROM services ORDER BY id")
	if err != nil {
		return nil, klassifiziere(err)
	}
	defer func() { _ = rows.Close() }()
	out := []model.Service{}
	for rows.Next() {
		s, err := leseService(rows)
		if err != nil {
			return nil, klassifiziere(err)
		}
		out = append(out, s)
	}
	return out, klassifiziere(rows.Err())
}

// SetMetadata ersetzt Blob-Referenz und Hash; beide leer löscht sie.
func (r *MetadataRepository) SetMetadata(ctx context.Context, id, blob, hash string) error {
	if err := model.MetadatenGueltig(blob, hash); err != nil {
		return err
	}
	return r.db.inTx(ctx, func(c *sql.Conn) error {
		res, err := c.ExecContext(ctx, "UPDATE services SET metadata_blob = ?, metadata_hash = ? WHERE id = ?",
			nullText(blob), nullText(hash), id)
		if err != nil {
			return klassifiziere(err)
		}
		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return fmt.Errorf("%w: Service %s", outbound.ErrNichtGefunden, id)
		}
		return nil
	})
}
