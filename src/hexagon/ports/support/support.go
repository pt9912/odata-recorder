// Package support enthält die Querschnitts-Ports Clock und IDGenerator (SPEC-114): Zeit und
// serverseitig vergebene Kennungen kommen nur über sie in den Kern; dazu Telemetry und
// TelemetryReader (SPEC-230).
package support

import "time"

// Clock liefert die aktuelle Zeit in UTC, nicht fallend und ohne Rundung (SPEC-114 Nr. 1 bis 4).
type Clock interface {
	Now() time.Time
}

// IDGenerator liefert Kennungen der Form <Präfix>-<32 Zeichen aus 0-9a-f> (SPEC-114 Nr. 6).
type IDGenerator interface {
	SessionID() string
	ExchangeID() string
	KorrelationsID() string
	WorkspaceID() string
	SnapshotID() string
}
