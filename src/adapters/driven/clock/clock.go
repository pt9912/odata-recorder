// Package clock ist der Driven Adapter der Querschnitts-Ports: Systemuhr liest die Systemzeit,
// ZufallsIDs zieht Kennungen aus crypto/rand (SPEC-114).
package clock

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var (
	_ support.Clock       = (*Systemuhr)(nil)
	_ support.IDGenerator = (*ZufallsIDs)(nil)
)

// Systemuhr liest die Zeit aus einer Quelle und liefert sie in UTC, nicht fallend (SPEC-114 Nr. 1 bis 3).
type Systemuhr struct {
	mu      sync.Mutex
	quelle  func() time.Time
	hoechst time.Time
	gesehen bool
}

// NewSystemuhr liefert eine Uhr über die Systemzeit.
func NewSystemuhr() *Systemuhr { return neuMit(time.Now) }

func neuMit(quelle func() time.Time) *Systemuhr { return &Systemuhr{quelle: quelle} }

// Now liefert die Quellzeit in UTC; UTC() entfernt die monotone Ablesung, gerundet wird nicht.
// Fällt die Quelle unter den bisherigen Höchstwert, bleibt Now darauf stehen. Bereich und
// Nullzeit der Quelle werden nicht geprüft.
func (u *Systemuhr) Now() time.Time {
	u.mu.Lock()
	defer u.mu.Unlock()
	t := u.quelle().UTC()
	if u.gesehen && t.Before(u.hoechst) {
		return u.hoechst
	}
	u.hoechst, u.gesehen = t, true
	return t
}

// ZufallsIDs vergibt Kennungen aus 128 Bit crypto/rand (SPEC-114 Nr. 6 bis 8).
type ZufallsIDs struct{}

// NewZufallsIDs liefert den Generator; er hält keinen Zustand.
func NewZufallsIDs() *ZufallsIDs { return &ZufallsIDs{} }

// SessionID liefert eine Kennung mit dem Präfix ses.
func (*ZufallsIDs) SessionID() string { return zufall("ses") }

// ExchangeID liefert eine Kennung mit dem Präfix exc.
func (*ZufallsIDs) ExchangeID() string { return zufall("exc") }

// KorrelationsID liefert eine Kennung mit dem Präfix cor.
func (*ZufallsIDs) KorrelationsID() string { return zufall("cor") }

// WorkspaceID liefert eine Kennung mit dem Präfix wsp.
func (*ZufallsIDs) WorkspaceID() string { return zufall("wsp") }

// SnapshotID liefert eine Kennung mit dem Präfix snp.
func (*ZufallsIDs) SnapshotID() string { return zufall("snp") }

// zufall hängt 16 Zufallsbyte als Hex an; rand.Read gibt keinen Fehler zurück, der Ausfall der Quelle ist in SPEC-114 als ungeprüft benannt.
func zufall(praefix string) string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return praefix + "-" + hex.EncodeToString(b[:])
}
