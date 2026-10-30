package outbound

import "github.com/pt9912/odata-recorder/src/hexagon/model"

// URLCodec liest Pfad und Query einer OData-V4-URL in den Query-AST des Modells (SPEC-160); die Grammatik des Pfads und der
// Schlüsselliterale steht in SPEC-161, die der Systemoptionen in SPEC-162.
type URLCodec interface {
	// ParseURL liest pfad (relativ zum Service-Root, mit führendem `/`, der Root ist `/`) und rawQuery (ohne führendes `?`).
	// Ein Verstoß gegen den Inhalt ist ein *model.URLFehler (ODR-E6005 oder ODR-E6001) mit dem Nullwert als Ergebnis;
	// ein Verstoß gegen die Form der Argumente ist model.ErrUngueltig ohne URLFehler. Es gibt keinen Kontext: die Funktion
	// liest nur ihre beiden Texte. Gleichzeitige Aufrufe sind erlaubt.
	ParseURL(pfad, rawQuery string) (model.ODataURL, error)
}
