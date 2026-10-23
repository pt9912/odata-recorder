package session

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ inbound.SessionUseCase = (*Verwaltung)(nil)

// offen ist der Fehler für das Löschen einer offenen Session; kein Port liefert ihn, er hat keine
// Ursache (SPEC-124 Nr. 7 c).
func offen() error {
	c := model.CodeE7009
	return &inbound.Fehler{Code: c, Status: c.Stati()[0]}
}

// Delete liest die Session, weist eine offene ab, löscht und ruft dann genau einmal die
// Bereinigung der Blobs mit dem ctx des Aufrufs; deren Bericht ist das Ergebnis, ihr Fehler
// steht nur im Protokoll (SPEC-124 Nr. 7). Rangfolge der Verstöße: unbekannt, offen, referenziert.
func (v *Verwaltung) Delete(ctx context.Context, id string) (inbound.Loeschbericht, error) {
	if ctx == nil {
		return inbound.Loeschbericht{}, ctxNil()
	}
	s, err := v.k.Sessions.Get(ctx, id)
	if err != nil {
		return inbound.Loeschbericht{}, mitCode(err, outbound.GegenstandSession)
	}
	if s.Status == model.SessionOpen {
		return inbound.Loeschbericht{}, offen()
	}
	if err := v.k.Sessions.Delete(ctx, id); err != nil {
		return inbound.Loeschbericht{}, mitCode(err, outbound.GegenstandSession)
	}
	bericht, err := v.k.Wartung.CleanupBlobs(ctx)
	v.protokolliere(ctx, s, bericht, err)
	return bericht, nil
}
