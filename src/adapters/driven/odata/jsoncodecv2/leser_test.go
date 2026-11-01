package jsoncodecv2_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodecv2"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

var _ outbound.ODataCodec = (*jsoncodecv2.Leser)(nil)

// TestVertrag fährt den Vertrag von SPEC-172 Nr. 1 bis 5 gegen den Leser.
func TestVertrag(t *testing.T) {
	outboundtest.ODataCodecV2Vertrag(t, func() outbound.ODataCodec { return jsoncodecv2.New() })
}

// TestLeserOhneZustand hält Zeile G19: der Leser hat kein Feld.
func TestLeserOhneZustand(t *testing.T) {
	if n := reflect.TypeFor[jsoncodecv2.Leser]().NumField(); n != 0 {
		t.Errorf("Leser hat %d Felder", n)
	}
}
