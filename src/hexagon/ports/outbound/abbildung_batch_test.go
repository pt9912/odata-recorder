package outbound_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestAbbildungBatch hält Zeile 29: `ErrBatchUngueltig` ergibt ODR-E6007, `ErrBatchNichtUnterstuetzt` ODR-E6001, bei jedem der acht
// Gegenstände, als Sentinel, als BatchFehler jeder der sechzehn Ursachen, verpackt, verbunden und neben einem Fehler der Gruppe
// ODR-E9001; neben einem Kontextfehler gibt es keinen Code. Menge 2 × 8 Sentinel, 16 Ursachen × 4 Formen × 8.
func TestAbbildungBatch(t *testing.T) {
	n := 0
	for _, s := range []struct {
		err  error
		soll model.Code
	}{{model.ErrBatchUngueltig, model.CodeE6007}, {model.ErrBatchNichtUnterstuetzt, model.CodeE6001}} {
		for _, g := range alleGegenstaende() {
			if got, ok := outbound.Abbilden(s.err, g); !ok || got != s.soll {
				t.Errorf("%v / Gegenstand %d: %v %v, erwartet %v", s.err, g, got, ok, s.soll)
			}
			n++
		}
	}
	nicht := []model.BatchUrsache{model.BatchFormat, model.BatchTeilTyp, model.BatchVerschachtelt, model.BatchTransferKodierung, model.BatchMethodeNichtUnterstuetzt}
	for _, u := range model.AlleBatchUrsachen() {
		soll := model.CodeE6007
		if slices.Contains(nicht, u) {
			soll = model.CodeE6001
		}
		e := &model.BatchFehler{Ursache: u, Ort: "1"}
		formen := map[string]error{
			"Fehler": e, "Errorf": fmt.Errorf("lesen: %w", e), "Join": errors.Join(errors.New("x"), e),
			"neben E9001": errors.Join(outbound.ErrAbgeschlossen, e),
		}
		for name, err := range formen {
			for _, g := range alleGegenstaende() {
				if got, ok := outbound.Abbilden(err, g); !ok || got != soll {
					t.Errorf("%s %s / Gegenstand %d: %v %v, erwartet %v", u, name, g, got, ok, soll)
				}
				n++
			}
		}
		for _, kontext := range []error{context.Canceled, context.DeadlineExceeded} {
			if _, ok := outbound.Abbilden(errors.Join(e, kontext), gSonst); ok {
				t.Errorf("%s: ein Kontextfehler neben dem Fehler bekommt einen Code", u)
			}
		}
	}
	if n != 2*8+16*4*8 {
		t.Errorf("%d Fälle", n)
	}
	for _, name := range []string{"ErrBatchUngueltig", "ErrBatchNichtUnterstuetzt"} {
		if !slices.Contains(outbound.EintraegeAbbildung(), name) {
			t.Errorf("die Tabelle nennt %s nicht", name)
		}
	}
}
