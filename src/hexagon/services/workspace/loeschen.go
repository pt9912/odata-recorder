package workspace

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ inbound.WorkspaceUseCase = (*Verwaltung)(nil)

// Delete löscht den Workspace id und ruft danach genau einmal die Bereinigung der Blobs mit dem ctx des
// Aufrufs; deren Bericht ist das Ergebnis, ihr Fehler steht nur im Protokoll (SPEC-129 Nr. 9). Der Dienst liest
// den Workspace davor nicht; die id geht unverändert an den Port.
func (v *Verwaltung) Delete(ctx context.Context, id string) (inbound.Loeschbericht, error) {
	if ctx == nil {
		return inbound.Loeschbericht{}, ctxNil()
	}
	if err := v.k.Workspaces.Delete(ctx, id); err != nil {
		return inbound.Loeschbericht{}, mitCode(err, outbound.GegenstandWorkspace)
	}
	bericht, err := v.k.Wartung.CleanupBlobs(ctx)
	v.protokolliere(ctx, id, bericht, err)
	return bericht, nil
}
