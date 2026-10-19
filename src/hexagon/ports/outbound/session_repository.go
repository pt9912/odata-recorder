package outbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// SessionRepository speichert Sessions (SPEC-105). Jede Operation ist atomar.
type SessionRepository interface {
	// Create legt eine Session im Zustand open an.
	Create(ctx context.Context, s model.Session) error
	// Get liest eine Session.
	Get(ctx context.Context, id string) (model.Session, error)
	// SetStatus wechselt von open nach complete oder incomplete.
	SetStatus(ctx context.Context, id string, status model.SessionStatus) error
	// List liefert die Sessions eines Service nach Erzeugungszeit und Kennung.
	List(ctx context.Context, serviceID string) ([]model.Session, error)
	// FindOpen liefert alle Sessions im Zustand open.
	FindOpen(ctx context.Context) ([]model.Session, error)
	// Delete löscht eine Session mit ihren Exchanges und Cursors.
	Delete(ctx context.Context, id string) error
}
