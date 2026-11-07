package inbound

import "context"

// MockUseCase beantwortet im Modus mock einen lesenden OData-Request aus dem Zustand eines Workspace, ohne einen
// Upstream zu kontaktieren (SPEC-195). Anfrage und Antwort sind die Werte von SPEC-121.
type MockUseCase interface {
	// Mock ordnet Service und Workspace zu und liefert Status, Header und Body der Antwort; der Adapter ruft Close
	// des Bodys genau einmal. Ein Fehler mit Code ist ein *Fehler, das Fehlerobjekt schreibt der Adapter.
	Mock(ctx context.Context, anfrage Anfrage) (Antwort, error)
}
