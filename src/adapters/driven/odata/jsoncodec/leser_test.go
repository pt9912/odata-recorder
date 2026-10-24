package jsoncodec_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.ODataCodec = (*jsoncodec.Leser)(nil)

// TestLeserOhneZustand hält Zeile K4: der Leser hat kein Feld.
func TestLeserOhneZustand(t *testing.T) {
	if n := reflect.TypeFor[jsoncodec.Leser]().NumField(); n != 0 {
		t.Errorf("Leser hat %d Felder", n)
	}
}
