package baselineimport

import (
	"context"
	"errors"
	"fmt"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Ursachen von ODR-E6004, die der Import selbst feststellt (SPEC-127 Nr. 7); jede macht
// model.ErrMetadatenUngueltig erreichbar.
var (
	ErrKeineMetadaten      = fmt.Errorf("%w: die Session enthält kein $metadata", model.ErrMetadatenUngueltig)
	ErrMetadatenAbweichend = fmt.Errorf("%w: die Session enthält verschiedene $metadata", model.ErrMetadatenUngueltig)
	ErrMetadatenHash       = fmt.Errorf("%w: metadata_hash passt nicht zum $metadata der Session", model.ErrMetadatenUngueltig)
)

// mitCode bildet den Fehler eines Ports nach SPEC-109 Nr. 5 mit dem Gegenstand g (SPEC-127 Nr. 17):
// der Status ist der eine Status des Codes, ein Kontextfehler bleibt unverändert ohne Code.
func mitCode(err error, g outbound.Gegenstand) error {
	c, ok := outbound.Abbilden(err, g)
	if !ok {
		return err
	}
	return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: err}
}

// ohneCode erzeugt den Fehler mit Code, den kein Port liefert.
func ohneCode(c model.Code, ursache error) error {
	return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: ursache}
}

func ungueltig(text string) error {
	return errors.Join(model.ErrUngueltig, errors.New(text))
}

// kontextFehler sagt, ob err ein beendeter Kontext ist.
func kontextFehler(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// sonst bildet den Fehler eines Ports mit dem Gegenstand Sonstiges.
func sonst(err error) error { return mitCode(err, outbound.GegenstandSonstiges) }
