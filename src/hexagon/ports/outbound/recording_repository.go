package outbound

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// RecordingRepository speichert die Exchanges einer Session append-only (SPEC-105): es
// gibt keine Operation, die einen Exchange ändert oder einzeln löscht.
type RecordingRepository interface {
	// Append hängt einen Exchange an eine offene Session an.
	Append(ctx context.Context, e model.Exchange) error
	// Get liest einen Exchange.
	Get(ctx context.Context, id string) (model.Exchange, error)
	// ListBySession liefert die Exchanges einer Session nach sequence_no.
	ListBySession(ctx context.Context, sessionID string) ([]model.Exchange, error)
	// ListByMatchKey liefert die Exchanges der Session, deren match_key byte-gleich matchKey ist, nach sequence_no, und die
	// größte sequence_no der Session (0 ohne Exchanges), beide aus demselben Stand (SPEC-105 Nr. 14). Ein leerer matchKey
	// trifft keinen Exchange; eine unbekannte Session, auch die leere Kennung, ist ErrNichtGefunden.
	ListByMatchKey(ctx context.Context, sessionID, matchKey string) ([]model.Exchange, int, error)
}
