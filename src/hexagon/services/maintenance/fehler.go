package maintenance

import (
	"context"
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// kontextFehler sagt, ob err einen Kontextfehler enthält, auch neben einem anderen Fehler.
func kontextFehler(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// mitCode bildet den Fehler eines Ports nach SPEC-123 Nr. 10: ein Kontextfehler bleibt unverändert
// ohne Code, ein beschädigter oder fehlender Blob und eine beschädigte Datenbank sind ODR-E3001,
// jeder andere Fehler ODR-E9001, immer mit Status 500 und dem Gegenstand Sonstiges.
func mitCode(err error) error {
	c, ok := outbound.Abbilden(err, outbound.GegenstandSonstiges)
	if !ok {
		return err
	}
	if c != model.CodeE3001 {
		c = model.CodeE9001
	}
	return &inbound.Fehler{Code: c, Status: 500, Ursache: err}
}
