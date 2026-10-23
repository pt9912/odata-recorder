package inbound

import "context"

// MaintenanceUseCase hält den Speicher in Ordnung (SPEC-123): Startup ist die Wartung beim Start
// vor der Annahme von Requests, CleanupBlobs die Bereinigung der Blobs, die ein löschender Use
// Case nach dem Commit seines Löschens ruft.
type MaintenanceUseCase interface {
	// Startup prüft die Datenbank, beendet offene Sessions, wendet die Aufbewahrung an und gleicht
	// die Blobs gegen die Referenzmenge ab; der Bericht trägt auch bei einem Fehler, was geschah.
	Startup(ctx context.Context) (Wartungsbericht, error)
	// CleanupBlobs entfernt die Blobs ohne Referenz, stellt sich aber bei einer offenen Session
	// zurück und schont, was ein Schreiber vor kurzem committete.
	CleanupBlobs(ctx context.Context) (Loeschbericht, error)
}

// Wartungsbericht zählt je Aufruf von Startup (SPEC-123); seine Listen sind nie nil.
type Wartungsbericht struct {
	OffeneBeendet         int
	SessionsGeloescht     int
	SessionsUebersprungen int
	BlobsEntfernt         int
	Befunde               []string
	BlobsFehlend          []string
	BlobsBeschaedigt      []string
}

// Loeschbericht zählt je Aufruf von CleanupBlobs (SPEC-123).
type Loeschbericht struct {
	Entfernt        int
	Zurueckgestellt bool
}
