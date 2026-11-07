package outbound

import (
	"context"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ODataEingabe liest den JSON-Body eines POST, PATCH oder PUT an den Mock-OData-Dienst gegen das Modell des Dienstes (SPEC-180);
// das Lesen der Antworten gehört zu ODataCodec, das Erzeugen zu ODataSchreiber. Je Version gibt es eine Umsetzung.
// Ein Programmfehler des Aufrufers (Argument, Set, Methode) ist model.ErrUngueltig ohne Code, ein beendeter Kontext liefert
// ctx.Err() unverändert, ein Fehler von src bleibt unverändert, jeder Verstoß gegen contentType oder Inhalt ist ein
// *model.EingabeFehler; bei jedem Fehler ist das Ergebnis der Nullwert. contentType sind die Werte der Kopfzeilen Content-Type,
// ein Element je Zeile; src wird nie geschlossen.
type ODataEingabe interface {
	LeseEntity(ctx context.Context, src io.Reader, contentType []string, m model.ServiceModel, entitySet string, art model.EingabeArt, grenze int64) (model.EntityEingabe, error)
}
