package csdl_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

var _ outbound.MetadataCodec = (*csdl.Leser)(nil)

func TestVertrag(t *testing.T) {
	outboundtest.MetadataCodecVertrag(t, func() outbound.MetadataCodec { return csdl.New() })
}

// TestLeserOhneZustand hält J2: der Leser hat kein Feld.
func TestLeserOhneZustand(t *testing.T) {
	if n := reflect.TypeFor[csdl.Leser]().NumField(); n != 0 {
		t.Errorf("Leser hat %d Felder", n)
	}
}
