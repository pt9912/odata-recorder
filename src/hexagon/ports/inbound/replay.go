package inbound

import "context"

// ReplayUseCase beantwortet im Modus replay einen Request aus einer ausdrücklich zugeordneten
// Recording-Session und einem Workspace, ohne einen Upstream zu kontaktieren (SPEC-135). Anfrage und
// Antwort sind die Werte von SPEC-121.
type ReplayUseCase interface {
	// Replay prüft die Zuordnung, wählt je Strategie einen replayfähigen Exchange, schreibt den Cursor
	// fort und liefert Status, Header und Body der Aufzeichnung; der Adapter ruft Close des Bodys genau einmal.
	Replay(ctx context.Context, anfrage Anfrage) (Antwort, error)
}
