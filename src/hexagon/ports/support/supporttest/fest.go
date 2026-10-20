// Package supporttest enthält die Test-Implementierungen FestUhr und FestIDs und die
// Vertragstests der Querschnitts-Ports (SPEC-114): Sie laufen gegen echte und feste
// Implementierung.
package supporttest

import (
	"fmt"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var (
	_ support.Clock       = (*FestUhr)(nil)
	_ support.IDGenerator = (*FestIDs)(nil)
)

// FestUhr ist eine Uhr, die nur Advance vorrückt (SPEC-114 Nr. 5).
type FestUhr struct {
	mu    sync.Mutex
	stand time.Time
}

// NewFestUhr rechnet start nach UTC um; die Nullzeit und ein Jahr außerhalb 1 bis 9999
// nach der Umrechnung sind eine Panik.
func NewFestUhr(start time.Time) *FestUhr {
	if start.IsZero() {
		panic("supporttest: NewFestUhr mit der Nullzeit")
	}
	u := start.UTC()
	if !imBereich(u) {
		panic(fmt.Sprintf("supporttest: NewFestUhr mit Jahr %d außerhalb 1 bis 9999", u.Year()))
	}
	return &FestUhr{stand: u}
}

func imBereich(t time.Time) bool { return t.Year() >= 1 && t.Year() <= 9999 }

// Now liefert den Stand und ändert ihn nicht.
func (u *FestUhr) Now() time.Time {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.stand
}

// Advance rückt den Stand um d vor; d < 0 oder ein Ergebnis nach dem Jahr 9999 ist eine Panik
// und ändert den Stand nicht.
func (u *FestUhr) Advance(d time.Duration) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if d < 0 {
		panic("supporttest: Advance mit negativer Dauer")
	}
	n := u.stand.Add(d)
	if !imBereich(n) {
		panic(fmt.Sprintf("supporttest: Advance führt in das Jahr %d", n.Year()))
	}
	u.stand = n
}

// FestIDs vergibt Kennungen mit einem Zähler je Operation, der bei 1 beginnt (SPEC-114 Nr. 10).
type FestIDs struct {
	mu      sync.Mutex
	zaehler map[string]uint64
}

// NewFestIDs liefert einen Generator mit frischen Zählern.
func NewFestIDs() *FestIDs { return &FestIDs{zaehler: map[string]uint64{}} }

func (g *FestIDs) naechste(praefix string) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.zaehler[praefix]++
	return fmt.Sprintf("%s-%032d", praefix, g.zaehler[praefix])
}

// SessionID liefert die nächste Kennung mit dem Präfix ses.
func (g *FestIDs) SessionID() string { return g.naechste("ses") }

// ExchangeID liefert die nächste Kennung mit dem Präfix exc.
func (g *FestIDs) ExchangeID() string { return g.naechste("exc") }

// KorrelationsID liefert die nächste Kennung mit dem Präfix cor.
func (g *FestIDs) KorrelationsID() string { return g.naechste("cor") }

// WorkspaceID liefert die nächste Kennung mit dem Präfix wsp.
func (g *FestIDs) WorkspaceID() string { return g.naechste("wsp") }

// SnapshotID liefert die nächste Kennung mit dem Präfix snp.
func (g *FestIDs) SnapshotID() string { return g.naechste("snp") }
