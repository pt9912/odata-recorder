package outbound

import (
	"context"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ODataBatch liest den Body eines POST auf `$batch` an den Mock-OData-Dienst eines Service der Version 4.0 (SPEC-182). Er liest
// allein: die Struktur prüft model.PruefeBatch. Ein Programmfehler des Aufrufers (ctx oder src nil, grenze kleiner 1) ist
// model.ErrUngueltig ohne Code, ein beendeter Kontext liefert ctx.Err() unverändert, ein Fehler von src bleibt erreichbar, jeder
// Verstoß gegen contentType oder Inhalt ist ein *model.BatchFehler; bei jedem Fehler ist das Ergebnis der Nullwert. contentType
// sind die Werte der Kopfzeilen Content-Type, ein Element je Zeile; src wird nie geschlossen.
type ODataBatch interface {
	LeseBatch(ctx context.Context, src io.Reader, contentType []string, grenze int64) (model.Batch, error)
}
