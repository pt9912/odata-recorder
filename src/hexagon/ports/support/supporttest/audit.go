package supporttest

import (
	"slices"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var _ support.AuditLog = (*FakeAudit)(nil)

// FakeAudit erfüllt support.AuditLog (SPEC-234 Nr. 9). Append hält den Eintrag mit allen sieben
// Feldern unverändert und ungeprüft fest und zählt den Aufruf, Eintraege zählt ebenfalls; Aufrufe
// ist die Summe beider Zähler. Fehler ist das Ergebnis von Append (der Eintrag steht dann trotzdem).
// Vor ist ein Haltepunkt: Append ruft ihn nach dem Zählen und vor dem Festhalten ohne die Sperre des
// Fakes, kehrt er nicht zurück, kehrt Append nicht zurück. Fehler und Vor setzt der Test vor dem
// ersten Aufruf. Der Nullwert ist ein nutzbarer Fake; er ist für gleichzeitigen Aufruf sicher.
type FakeAudit struct {
	Fehler error
	Vor    func(support.Eintrag)

	mu        sync.Mutex
	eintraege []support.Eintrag
	appends   int
	lesungen  int
}

// NewFakeAudit liefert einen Fake ohne Eintrag, ohne Fehler und ohne Haltepunkt.
func NewFakeAudit() *FakeAudit {
	return &FakeAudit{}
}

// Append zählt den Aufruf, ruft Vor, hält den Eintrag fest und liefert Fehler.
func (f *FakeAudit) Append(e support.Eintrag) error {
	f.mu.Lock()
	f.appends++
	f.mu.Unlock()
	if f.Vor != nil {
		f.Vor(e)
	}
	f.mu.Lock()
	f.eintraege = append(f.eintraege, e)
	f.mu.Unlock()
	return f.Fehler
}

// Eintraege zählt den Aufruf und liefert die festgehaltenen Einträge in der Reihenfolge der Aufrufe
// als Kopie.
func (f *FakeAudit) Eintraege() []support.Eintrag {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lesungen++
	return slices.Clone(f.eintraege)
}

// Aufrufe liefert die Zahl der Aufrufe von Append und von Eintraege zusammen.
func (f *FakeAudit) Aufrufe() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.appends + f.lesungen
}
