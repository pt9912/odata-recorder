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

// Fehler des BlobRepository (SPEC-106). Die Schicht darüber meldet ErrBlobFehlt und
// ErrBlobBeschaedigt als ODR-E3001.
var (
	ErrBlobFehlt       = errors.New("Blob fehlt")
	ErrBlobBeschaedigt = errors.New("Blob beschädigt")
	ErrZuGross         = errors.New("Body überschreitet die Grenze")
	ErrAbgeschlossen   = errors.New("Writer ist abgeschlossen")
)

// Fehler der Workspace-Repositories (SPEC-108).
var (
	ErrRevisionVeraltet = errors.New("Revision ist veraltet")
	ErrFremderService   = errors.New("gehört zu einem anderen Service")
	ErrFalscheArt       = errors.New("Snapshot hat die falsche Art")
)
