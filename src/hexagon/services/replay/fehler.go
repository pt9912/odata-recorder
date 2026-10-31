package replay

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// ungueltig ist ein Fehler des Aufrufs, den kein HTTP-Adapter liefert: ohne Code und ohne einen Wert
// aus dem Request (SPEC-135 Nr. 2, 15).
func ungueltig(text string) error {
	return errors.Join(model.ErrUngueltig, errors.New(text))
}

// mitCode bildet den Fehlerwert eines Codes mit seinem einen Status (SPEC-135 Nr. 15).
func mitCode(c model.Code, ursache error) *inbound.Fehler {
	return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: ursache}
}

// abbilden bildet den Fehler eines Ports nach SPEC-109 Nr. 5 mit dem Gegenstand g (SPEC-135 Nr. 15, 18);
// ein Kontextfehler kehrt unverändert und ohne Code zurück (SPEC-135 Nr. 2).
func abbilden(err error, g outbound.Gegenstand) error {
	c, ok := outbound.Abbilden(err, g)
	if !ok {
		return err
	}
	return mitCode(c, err)
}
