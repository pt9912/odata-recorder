package workspace

import (
	"context"
	"errors"
	"io"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Create legt den Workspace id aus der Baseline baselineID an (SPEC-129 Nr. 4): Argumente, Baseline
// lesen und Manifest prüfen, erst dann Kennung und Zeit, Abbildung und die eine Transaktion mit
// Workspaces.Create und Entities.ReplaceAll. Die Doppelung der Kennung entscheidet allein Create.
func (v *Verwaltung) Create(ctx context.Context, baselineID, id string) (model.Workspace, error) {
	if ctx == nil {
		return model.Workspace{}, ctxNil()
	}
	if baselineID == "" || (id != "" && !idGueltig(id)) {
		return model.Workspace{}, errors.Join(model.ErrUngueltig, errors.New("baselineID fehlt oder id außerhalb der Form"))
	}
	snap, eintraege, err := v.baseline(ctx, baselineID)
	if err != nil {
		return model.Workspace{}, err
	}
	if id == "" {
		id = v.k.IDs.WorkspaceID()
	}
	w := model.Workspace{ID: id, ServiceID: snap.ServiceID, BaselineSnapshotID: baselineID, Revision: 0, CreatedAt: v.k.Uhr.Now()}
	state := abbilden(id, eintraege)
	err = v.transaktion(ctx, func(s outbound.WorkspaceStores) error {
		if err := s.Workspaces.Create(ctx, w); err != nil {
			g := outbound.GegenstandWorkspace
			if errors.Is(err, outbound.ErrNichtGefunden) {
				g = outbound.GegenstandSnapshot
			}
			return mitCode(err, g)
		}
		if err := s.Entities.ReplaceAll(ctx, id, state); err != nil {
			return mitCode(err, outbound.GegenstandWorkspace)
		}
		return nil
	})
	if err != nil {
		return model.Workspace{}, err
	}
	return w, nil
}

// baseline liest die Baseline baselineID und ihr Manifest (SPEC-129 Nr. 5): Snapshot, Vergleich von
// StateHash und StateBlob, Open, Lesen bis zum Ende, Close genau einmal, Prüfung des Manifests.
func (v *Verwaltung) baseline(ctx context.Context, baselineID string) (model.BaselineSnapshot, []eintrag, error) {
	snap, err := v.k.Snapshots.Get(ctx, baselineID)
	if err != nil {
		return model.BaselineSnapshot{}, nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	if snap.StateHash != snap.StateBlob {
		return model.BaselineSnapshot{}, nil, mitCode(verstoss("state_hash ungleich state_blob"), outbound.GegenstandSnapshot)
	}
	rc, err := v.k.Blobs.Open(ctx, snap.StateBlob)
	if err != nil {
		return model.BaselineSnapshot{}, nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	defer func() { _ = rc.Close() }()
	daten, err := io.ReadAll(rc)
	if err != nil {
		return model.BaselineSnapshot{}, nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	es, err := leseManifest(daten, snap)
	if err != nil {
		return model.BaselineSnapshot{}, nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	return snap, es, nil
}
