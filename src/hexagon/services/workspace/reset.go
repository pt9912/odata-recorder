package workspace

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Reset setzt den Workspace id auf seine Baseline zurück (SPEC-129 Nr. 10): Workspace lesen, Baseline
// und Manifest lesen, dann eine Transaktion mit Get, ReplaceAll, ResetAll und Advance. Die Revision wächst
// bei jedem Aufruf um 1; ein Fehler vor oder in der Transaktion lässt den Zustand unverändert.
func (v *Verwaltung) Reset(ctx context.Context, id string) (model.Workspace, error) {
	if ctx == nil {
		return model.Workspace{}, ctxNil()
	}
	w, err := v.k.Workspaces.Get(ctx, id)
	if err != nil {
		return model.Workspace{}, mitCode(err, outbound.GegenstandWorkspace)
	}
	_, eintraege, err := v.baseline(ctx, w.BaselineSnapshotID)
	if err != nil {
		return model.Workspace{}, err
	}
	state := abbilden(id, eintraege)
	var ergebnis model.Workspace
	err = v.transaktion(ctx, func(s outbound.WorkspaceStores) error {
		aktuell, err := s.Workspaces.Get(ctx, id)
		if err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		if err := s.Entities.ReplaceAll(ctx, id, state); err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		if err := s.Cursors.ResetAll(ctx, id); err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		revision, err := s.Workspaces.Advance(ctx, id, aktuell.Revision)
		if err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		aktuell.Revision = revision
		ergebnis = aktuell
		return nil
	})
	if err != nil {
		return model.Workspace{}, err
	}
	return ergebnis, nil
}
