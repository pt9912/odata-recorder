package outbound

import (
	"context"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// ODataSchreiber erzeugt die Darstellung der Antworten des Mock-OData-Dienstes (SPEC-168): eine Entity, eine Collection,
// die Anzahl, das Fehlerobjekt und die Adresse einer Entity. Je Version gibt es eine Umsetzung; das Lesen gehört zu
// ODataCodec. Jede Operation mit dst schreibt bei Erfolg die Bytes in genau einem Write und liefert den Darstellungskopf.
// Ein Verstoß gegen das Argument oder das Format ist model.ErrUngueltig ohne Code, ein beendeter Kontext liefert ctx.Err()
// unverändert, ein Fehler von dst bleibt unverändert; bei jedem Fehler ist die Darstellung der Nullwert und dst leer.
type ODataSchreiber interface {
	SchreibeEntity(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.EntityAntwort) (model.Darstellung, error)
	SchreibeCollection(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.CollectionAntwort) (model.Darstellung, error)
	SchreibeAnzahl(ctx context.Context, dst io.Writer, n int64) (model.Darstellung, error)
	SchreibeFehler(ctx context.Context, dst io.Writer, f model.Fehlerobjekt) (model.Darstellung, error)
	// Adresse liefert die Adresse der Entity mit dem Schlüssel im Set (`Location`, `@odata.id`); eine reine Funktion.
	Adresse(m model.ServiceModel, wurzel, entitySet, schluessel string) (string, error)
}
