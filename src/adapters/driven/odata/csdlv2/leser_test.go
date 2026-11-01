package csdlv2_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdlv2"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

var _ outbound.MetadataCodec = (*csdlv2.Leser)(nil)

func TestVertrag(t *testing.T) {
	outboundtest.MetadataCodecV2Vertrag(t, func() outbound.MetadataCodec { return csdlv2.New() })
}

// TestLeserOhneZustand hält J1: der Leser hat kein Feld.
func TestLeserOhneZustand(t *testing.T) {
	if n := reflect.TypeFor[csdlv2.Leser]().NumField(); n != 0 {
		t.Errorf("Leser hat %d Felder", n)
	}
}
