package bootstrap_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound/outboundtest"
)

// TestODataCodecVertrag verdrahtet den Leser der JSON-Antworten mit der Vertragssuite des Ports; der Composition Root
// ist die Stelle, die beide Adapter kennen darf: der Leser des CSDL baut die Modelle der Fixtures.
func TestODataCodecVertrag(t *testing.T) {
	outboundtest.ODataCodecVertrag(t, func() outbound.ODataCodec { return jsoncodec.New() }, csdl.New())
}
