// Package model enthält die fachlichen Werte des Recorders: Sessions, Exchanges und ihre
// Zustände. Es nutzt nur die Standardbibliothek und führt keine I/O aus.
package model

import (
	"errors"
	"time"
)

// SessionStatus ist der Zustand einer Session (SPEC-105).
type SessionStatus string

// Zustände einer Session: open geht in complete oder incomplete über und bleibt dann.
const (
	SessionOpen       SessionStatus = "open"
	SessionComplete   SessionStatus = "complete"
	SessionIncomplete SessionStatus = "incomplete"
)

// Gueltig sagt, ob s ein bekannter Zustand ist.
func (s SessionStatus) Gueltig() bool {
	return s == SessionOpen || s == SessionComplete || s == SessionIncomplete
}

// KannWechselnNach sagt, ob ein Wechsel von s nach ziel erlaubt ist: nur open → complete
// und open → incomplete.
func (s SessionStatus) KannWechselnNach(ziel SessionStatus) bool {
	return s == SessionOpen && (ziel == SessionComplete || ziel == SessionIncomplete)
}

// zeitGueltig sagt, ob t gesetzt ist und in den Jahren 1 bis 9999 liegt, die das
// Speicherformat darstellt (SPEC-105).
func zeitGueltig(t time.Time) bool {
	j := t.UTC().Year()
	return !t.IsZero() && j >= 1 && j <= 9999
}

// Session ist eine Aufzeichnungssitzung eines Service.
type Session struct {
	ID        string
	ServiceID string
	CreatedAt time.Time
	Status    SessionStatus
}

// ErrUngueltig meldet einen Wert, der die Invarianten des Modells verletzt.
var ErrUngueltig = errors.New("model: ungültiger Wert")

// Validate prüft die Invarianten einer neuen Session: Kennungen gesetzt, Zeit gesetzt,
// Zustand open.
func (s Session) Validate() error {
	switch {
	case s.ID == "":
		return errors.Join(ErrUngueltig, errors.New("Session ohne Kennung"))
	case s.ServiceID == "":
		return errors.Join(ErrUngueltig, errors.New("Session ohne Service"))
	case !zeitGueltig(s.CreatedAt):
		return errors.Join(ErrUngueltig, errors.New("Session ohne gültige Erzeugungszeit"))
	case s.Status != SessionOpen:
		return errors.Join(ErrUngueltig, errors.New("neue Session muss open sein"))
	}
	return nil
}
