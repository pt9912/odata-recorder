package inbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// CloseBericht zählt je Aufruf von CloseAll (SPEC-124 Nr. 14): Beendet die gelungenen Sessions,
// Fehlgeschlagen die, die ein Fehler außer ODR-E7011 an der Beendigung hinderte.
type CloseBericht struct {
	Beendet        int
	Fehlgeschlagen int
}

// SessionUseCase verwaltet die Recording-Sessions (SPEC-124): anlegen, lesen, auflisten,
// löschen und beenden. Jeder Fehler mit Code ist ein *Fehler; model.ErrUngueltig und die
// Kontextfehler kehren ohne Code zurück.
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
	// Beenden hält die Aufzeichnung der Session an (Drain) und setzt danach ihren Status auf
	// complete oder incomplete (SPEC-124 Nr. 11, 12); das Ergebnis ist die Session mit dem neuen Status.
	Beenden(ctx context.Context, id string) (model.Session, error)
	// CloseAll beendet jede Session, die beim Aufruf offen ist, alle zugleich (SPEC-124 Nr. 14).
	CloseAll(ctx context.Context) (CloseBericht, error)
}
