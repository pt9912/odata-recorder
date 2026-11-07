package snapshot

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// mitCode bildet den Fehler eines Ports nach SPEC-109 Nr. 5 mit dem Gegenstand g (SPEC-129 Nr. 24, 26):
// der Status ist der eine Status des Codes, ein Kontextfehler bleibt unverändert ohne Code.
func mitCode(err error, g outbound.Gegenstand) error {
	c, ok := outbound.Abbilden(err, g)
	if !ok {
		return err
	}
	return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: err}
}

func ctxNil() error {
	return errors.Join(model.ErrUngueltig, errors.New("Kontext fehlt"))
}

// nameUngueltig ist der Fehler eines ungültigen Namens (SPEC-129 Nr. 21); der Wert steht nicht im Text.
func nameUngueltig() error {
	return errors.Join(model.ErrUngueltig, errors.New("Snapshotname ungültig"))
}
