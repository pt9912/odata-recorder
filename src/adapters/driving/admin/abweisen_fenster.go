package admin

import (
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Die Grenze der Einträge des Abweisens (SPEC-301 Nr. 1, 2): feste Werte, kein Schlüssel der Konfiguration.
const (
	// AbweisungenJeFenster ist die Zahl der Einträge admin_abgewiesen, die ein Fenster höchstens zulässt.
	AbweisungenJeFenster = 20
	// FensterDauer ist die Dauer eines Fensters ab der Abweisung, die es eröffnet.
	FensterDauer = 60 * time.Second
)

// abweisenFenster ist das Fenster der Einträge admin_abgewiesen eines Server (SPEC-301 Nr. 2, 5 f). Der Nullwert hat
// noch kein Fenster eröffnet: das sagt offen, nie der Nullwert von beginn, denn eine verstellte Uhr kann die Nullzeit
// liefern (SPEC-301 Nr. 5 c). mu schützt alle Felder und gehört allein dem Fenster.
type abweisenFenster struct {
	mu     sync.Mutex
	offen  bool
	beginn time.Time
	zahl   int
}

// zulasse entscheidet, ob eine Abweisung ihren Eintrag schreibt (SPEC-301 Nr. 2): Lesen der Uhr, Eröffnen des Fensters
// und Zählen sind ein Schritt unter einer Sperre, die vor der Rückkehr frei ist. Ein neues Fenster beginnt bei der ersten
// Abweisung und bei jeder, deren Zeit mindestens FensterDauer nach dem Beginn liegt; die Differenz sättigt
// (time.Time.Sub), ein Wert vor dem Beginn lässt das Fenster offen. Die Zahl steigt nie über AbweisungenJeFenster.
func (f *abweisenFenster) zulasse(uhr support.Clock) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	jetzt := uhr.Now()
	if !f.offen || jetzt.Sub(f.beginn) >= FensterDauer {
		f.offen, f.beginn, f.zahl = true, jetzt, 0
	}
	if f.zahl >= AbweisungenJeFenster {
		return false
	}
	f.zahl++
	return true
}
