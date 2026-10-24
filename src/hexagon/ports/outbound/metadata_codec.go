package outbound

import (
	"context"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// MetadataCodec liest ein EDMX/CSDL-Dokument in das Modell des OData-Dienstes (SPEC-125).
type MetadataCodec interface {
	// ParseMetadata liest src bis io.EOF, höchstens grenze+1 Bytes (SPEC-125 Nr. 5),
	// und liefert Modell und Bericht.
	// Ein Verstoß ist ein *model.MetadatenFehler (ODR-E6004) mit Nullwerten für Modell und Bericht;
	// ein ungültiges Argument ist model.ErrUngueltig, ein beendeter Kontext und ein Fehler von src
	// bleiben ohne MetadatenFehler. Gleichzeitige Aufrufe sind erlaubt.
	ParseMetadata(ctx context.Context, src io.Reader, grenze int64) (model.ServiceModel, model.ValidierungsBericht, error)
}
