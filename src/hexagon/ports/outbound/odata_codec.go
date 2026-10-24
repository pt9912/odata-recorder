package outbound

import (
	"context"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ODataCodec liest OData-V4-JSON-Antworten auf Einzel-Entity- und Collection-Abfragen gegen das Modell des
// Dienstes (SPEC-126); das Schreiben von JSON und die OData-URL gehören nicht zu ihm.
type ODataCodec interface {
	// ParseEntity liest die Antwort auf `GET <entitySet>(<Schlüssel>)` aus src bis io.EOF, höchstens grenze+1 Bytes.
	// Ein Verstoß ist ein *model.DatenFehler (ODR-E6101) mit Nullwerten für das Ergebnis; ein ungültiges Argument
	// ist model.ErrUngueltig, ein beendeter Kontext und ein Fehler von src bleiben ohne DatenFehler.
	ParseEntity(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (model.Entity, error)
	// ParseCollection liest die Antwort auf `GET <entitySet>` mit denselben Zusagen für Fehler und Argumente.
	ParseCollection(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (model.EntitySeite, error)
}
