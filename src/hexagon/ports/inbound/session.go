package inbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// SessionUseCase verwaltet die Recording-Sessions (SPEC-124): anlegen, lesen, auflisten und
// löschen. Jeder Fehler mit Code ist ein *Fehler; model.ErrUngueltig und die Kontextfehler
// kehren ohne Code zurück.
type SessionUseCase interface {
	// Create legt eine Session im Zustand open an; eine leere id vergibt der Dienst, eine
	// gelieferte muss der Form aus SPEC-124 Nr. 4 genügen.
	Create(ctx context.Context, serviceID, id string) (model.Session, error)
	// OpenInitialSession legt die erste Session eines Service an: Create mit leerer id.
	OpenInitialSession(ctx context.Context, serviceID string) (model.Session, error)
	// Get liest eine Session in jedem Zustand.
	Get(ctx context.Context, id string) (model.Session, error)
	// List liefert die Sessions aller Services; ohne Session eine leere Liste.
	List(ctx context.Context) ([]model.Session, error)
	// Delete löscht eine nicht offene Session und ruft danach die Bereinigung der Blobs; deren
	// Bericht ist das Ergebnis, ihr Fehler ändert es nicht.
	Delete(ctx context.Context, id string) (Loeschbericht, error)
}
