package inbound

import "context"

// MockUseCase beantwortet im Modus mock einen OData-Request aus dem Zustand eines Workspace, ohne einen Upstream zu
// kontaktieren: lesend (SPEC-195) und mit POST, PATCH, PUT und DELETE schreibend in einer Workspace-Transaktion (SPEC-192).
// Anfrage und Antwort sind die Werte von SPEC-121.
type MockUseCase interface {
	// Mock ordnet Service und Workspace zu und liefert Status, Header und Body der Antwort; der Adapter ruft Close
	// des Bodys genau einmal. Ein Fehler mit Code ist ein *Fehler, das Fehlerobjekt schreibt der Adapter.
	Mock(ctx context.Context, anfrage Anfrage) (Antwort, error)
}
