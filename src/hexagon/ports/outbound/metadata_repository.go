package outbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// MetadataRepository speichert Services und ihre Metadaten-Referenz (SPEC-107).
type MetadataRepository interface {
	// Create legt einen Service an.
	Create(ctx context.Context, s model.Service) error
	// Get liest einen Service.
	Get(ctx context.Context, id string) (model.Service, error)
	// List liefert alle Services nach Kennung.
	List(ctx context.Context) ([]model.Service, error)
	// SetMetadata ersetzt Blob-Referenz und Hash; beide leer löscht sie.
	SetMetadata(ctx context.Context, id, blob, hash string) error
}
