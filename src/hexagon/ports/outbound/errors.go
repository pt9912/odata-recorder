// Package outbound enthält die Outbound Ports: die Schnittstellen, über die der Kern
// Persistenz und Upstream erreicht, und ihre Fehler.
package outbound

import "errors"

// Fehler der Repositories (SPEC-105). Ein Adapter meldet sie, gegebenenfalls mit
// Zusatztext verpackt; Aufrufer prüfen mit errors.Is.
var (
	ErrNichtGefunden        = errors.New("nicht gefunden")
	ErrBereitsVorhanden     = errors.New("bereits vorhanden")
	ErrDienstUnbekannt      = errors.New("Service unbekannt")
	ErrUngueltigerUebergang = errors.New("ungültiger Zustandsübergang")
	ErrSessionGeschlossen   = errors.New("Session ist geschlossen")
	ErrUngueltigeSequenz    = errors.New("sequence_no passt nicht zur Folge")
	ErrReferenziert         = errors.New("wird noch referenziert")
)
