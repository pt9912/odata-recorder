package admin_test

import (
	"net"
	"net/http"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Bindung der Fakes und des Adapters an die Ports zur Übersetzungszeit: eine geänderte Signatur übersetzt das Paket nicht mehr.
var (
	_ support.AuditLog    = (*supporttest.FakeAudit)(nil)
	_ support.Telemetry   = (*supporttest.FakeTelemetry)(nil)
	_ support.IDGenerator = (*zaehlendeIDs)(nil)
	_ net.Listener        = (*listenerFake)(nil)
	_ admin.Handler       = (*routeFake)(nil).handle
	_ http.Handler        = (*admin.Server)(nil)
)
