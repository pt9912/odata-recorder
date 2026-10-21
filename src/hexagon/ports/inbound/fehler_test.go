package inbound_test

import (
	"errors"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

var errUrsache = errors.New("ursache mit S3CR3T-7f")

// TestFehler hält Error und Unwrap fest (SPEC-121 Nr. 14): der Text ist Code und message, ohne
// einen Wert der Ursache; errors.Is erreicht die Ursache, auch ohne Ursache gibt es keinen Panic.
func TestFehler(t *testing.T) {
	f := &inbound.Fehler{Code: model.CodeE4001, Status: 504, Ursache: errUrsache}
	if got, soll := f.Error(), "ODR-E4001: Upstream unavailable"; got != soll {
		t.Errorf("Error() = %q, erwartet %q", got, soll)
	}
	if !errors.Is(f, errUrsache) {
		t.Error("errors.Is erreicht die Ursache nicht")
	}
	ohne := &inbound.Fehler{Code: model.CodeE5009, Status: 409}
	if got, soll := ohne.Error(), "ODR-E5009: No recording session accepts the request"; got != soll {
		t.Errorf("Error() = %q, erwartet %q", got, soll)
	}
	if ohne.Unwrap() != nil {
		t.Errorf("Unwrap ohne Ursache = %v", ohne.Unwrap())
	}
}
