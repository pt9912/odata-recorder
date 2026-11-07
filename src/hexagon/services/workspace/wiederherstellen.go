package workspace

import (
	"context"
	"errors"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Restore ersetzt den Entity-State des Workspace workspaceID durch den des Workspace-Snapshots snapshotID
// (SPEC-129 Nr. 16): Workspace lesen, Snapshot lesen und Art prüfen, Hash gegen Blob, Blob und Manifest
// prüfen, dann eine Transaktion mit Get, ReplaceAll und Advance. Die Revision wächst bei jedem Aufruf um 1;
// der Dienst ruft keine Operation von Cursors, und ein Fehler vor oder in der Transaktion lässt den Zustand
// unverändert.
func (v *Verwaltung) Restore(ctx context.Context, workspaceID, snapshotID string) (model.Workspace, error) {
	if ctx == nil {
		return model.Workspace{}, ctxNil()
	}
	w, err := v.k.Workspaces.Get(ctx, workspaceID)
	if err != nil {
		return model.Workspace{}, mitCode(err, outbound.GegenstandWorkspace)
	}
	snap, err := v.k.Snapshots.GetWorkspace(ctx, workspaceID, snapshotID)
	if err != nil {
		if !errors.Is(err, outbound.ErrFalscheArt) || kontextFehler(err) {
			return model.Workspace{}, mitCode(err, outbound.GegenstandSnapshot)
		}
		return model.Workspace{}, v.baselineStattSnapshot(ctx, w, snapshotID, err)
	}
	eintraege, err := v.snapshotManifest(ctx, snap, workspaceID)
	if err != nil {
		return model.Workspace{}, err
	}
	state := abbildenSnapshot(workspaceID, eintraege)
	var ergebnis model.Workspace
	err = v.transaktion(ctx, func(s outbound.WorkspaceStores) error {
		aktuell, err := s.Workspaces.Get(ctx, workspaceID)
		if err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		if err := s.Entities.ReplaceAll(ctx, workspaceID, state); err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		revision, err := s.Workspaces.Advance(ctx, workspaceID, aktuell.Revision)
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

// baselineStattSnapshot entscheidet, wenn die Kennung eine Baseline nennt (SPEC-129 Nr. 16 d): die Baseline
// eines anderen Service ist ODR-E5004 ohne Ursache, jede andere Baseline ODR-E7008 aus ErrFalscheArt.
func (v *Verwaltung) baselineStattSnapshot(ctx context.Context, w model.Workspace, snapshotID string, falscheArt error) error {
	b, err := v.k.Snapshots.Get(ctx, snapshotID)
	if err != nil {
		return mitCode(err, outbound.GegenstandSnapshot)
	}
	if b.ServiceID != w.ServiceID {
		return &inbound.Fehler{Code: model.CodeE5004, Status: model.CodeE5004.Stati()[0]}
	}
	return mitCode(falscheArt, outbound.GegenstandSnapshot)
}

// snapshotManifest liest das Manifest des Workspace-Snapshots snap (SPEC-129 Nr. 16 e bis g): Vergleich von
// StateHash und StateBlob, Open, Lesen bis zum Ende, Close genau einmal, Prüfung des Manifests.
func (v *Verwaltung) snapshotManifest(ctx context.Context, snap model.WorkspaceSnapshot, workspaceID string) ([]eintrag, error) {
	if snap.StateHash != snap.StateBlob {
		return nil, mitCode(verstoss("state_hash ungleich state_blob"), outbound.GegenstandSnapshot)
	}
	rc, err := v.k.Blobs.Open(ctx, snap.StateBlob)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	defer func() { _ = rc.Close() }()
	daten, err := io.ReadAll(rc)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	es, err := leseSnapshotManifest(daten, snap, workspaceID)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	return es, nil
}
