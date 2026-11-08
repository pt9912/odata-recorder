package hybrid

import (
	"context"
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// ungueltig ist ein Fehler des Aufrufs, den kein HTTP-Adapter liefert: ohne Code und ohne einen Wert aus dem
// Request (SPEC-284 Nr. 2).
func ungueltig(text string) error {
	return errors.Join(model.ErrUngueltig, errors.New(text))
}

// codeVon liefert den Code eines Fehlers der Stufe: den eines *inbound.Fehler, der nicht der Nullwert ist und neben
// dem kein Kontextfehler steht (SPEC-284 Nr. 3 b).
func codeVon(err error) (model.Code, bool) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return 0, false
	}
	var f *inbound.Fehler
	if !errors.As(err, &f) || f.Code == 0 {
		return 0, false
	}
	return f.Code, true
}
