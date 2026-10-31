package workspace

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// mitCode bildet den Fehler eines Ports nach SPEC-109 Nr. 5 mit dem Gegenstand g (SPEC-129 Nr. 12):
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
