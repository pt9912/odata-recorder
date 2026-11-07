// Package snapshot ist der Snapshot-Dienst: er erzeugt benannte Workspace-Snapshots, die allein den Entity-State
// festhalten, und listet sie; er setzt SnapshotUseCase um (SPEC-129 Nr. 21 bis 26). Er importiert Domain und Ports,
// keinen anderen Use Case und keinen Adapter.
package snapshot

import (
	"context"
	"errors"
	"math"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// MaxNameLaenge ist die größte Länge eines Snapshotnamens in Bytes (SPEC-129 Nr. 21).
const MaxNameLaenge = 128

// Konfig ist die Konfiguration des Dienstes (SPEC-129 Nr. 21). Workspaces dient allein Get, Entities allein List,
// Snapshots Get (die Baseline), CreateWorkspace und ListWorkspace, Blobs allein Begin, Uhr und IDs allein Create.
type Konfig struct {
	Workspaces outbound.WorkspaceRepository
	Entities   outbound.EntityRepository
	Snapshots  outbound.SnapshotRepository
	Blobs      outbound.BlobRepository
	Uhr        support.Clock
	IDs        support.IDGenerator
}

// Dienst setzt die Operationen von SnapshotUseCase um. Er hält keinen Zustand außer seiner Konfiguration und
// keine Sperre (SPEC-129 Nr. 25).
type Dienst struct {
	k Konfig
}

var _ inbound.SnapshotUseCase = (*Dienst)(nil)

// New prüft die Konfiguration und ruft keinen Port (SPEC-129 Nr. 21).
func New(k Konfig) (*Dienst, error) {
	if k.Workspaces == nil || k.Entities == nil || k.Snapshots == nil || k.Blobs == nil || k.Uhr == nil || k.IDs == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Port, Uhr oder Generator fehlt"))
	}
	return &Dienst{k: k}, nil
}

// nameGueltig sagt, ob name der Form aus SPEC-129 Nr. 21 genügt: 1 bis MaxNameLaenge Bytes, gültiges UTF-8, kein
// Byte unter 0x20 und nicht 0x7F.
func nameGueltig(name string) bool {
	if len(name) == 0 || len(name) > MaxNameLaenge || !utf8.ValidString(name) {
		return false
	}
	for i := 0; i < len(name); i++ {
		if name[i] < 0x20 || name[i] == 0x7F {
			return false
		}
	}
	return true
}

// Create erfasst den Entity-State des Workspace workspaceID unter dem Namen name (SPEC-129 Nr. 22): Name,
// Workspace, Baseline, State in einem atomaren Aufruf, Manifest, Blob, erst dann Kennung und Zeit und die eine
// atomare Anlage. Die Eindeutigkeit des Namens entscheidet allein CreateWorkspace.
func (d *Dienst) Create(ctx context.Context, workspaceID, name string) (model.WorkspaceSnapshot, error) {
	if ctx == nil {
		return model.WorkspaceSnapshot{}, ctxNil()
	}
	if !nameGueltig(name) {
		return model.WorkspaceSnapshot{}, nameUngueltig()
	}
	w, err := d.k.Workspaces.Get(ctx, workspaceID)
	if err != nil {
		return model.WorkspaceSnapshot{}, mitCode(err, outbound.GegenstandWorkspace)
	}
	baseline, err := d.k.Snapshots.Get(ctx, w.BaselineSnapshotID)
	if err != nil {
		return model.WorkspaceSnapshot{}, mitCode(err, outbound.GegenstandSnapshot)
	}
	state, err := d.k.Entities.List(ctx, workspaceID)
	if err != nil {
		return model.WorkspaceSnapshot{}, mitCode(err, outbound.GegenstandWorkspace)
	}
	ref, err := d.ablegen(ctx, manifest(w.ServiceID, baseline.MetadataHash, state))
	if err != nil {
		return model.WorkspaceSnapshot{}, err
	}
	id := d.k.IDs.SnapshotID()
	jetzt := d.k.Uhr.Now()
	s := model.WorkspaceSnapshot{
		ID: id, ServiceID: w.ServiceID, WorkspaceID: workspaceID, Name: name, CreatedAt: jetzt,
		StateBlob: ref, StateHash: ref, MetadataHash: baseline.MetadataHash,
	}
	if err := d.k.Snapshots.CreateWorkspace(ctx, s); err != nil {
		g := outbound.GegenstandWorkspace
		if errors.Is(err, outbound.ErrBereitsVorhanden) {
			g = outbound.GegenstandSnapshotname
		}
		return model.WorkspaceSnapshot{}, mitCode(err, g)
	}
	return s, nil
}

// ablegen schreibt die Bytes b als Blob und liefert dessen Referenz (SPEC-129 Nr. 22 g): Begin, Write, Commit;
// Abort genau einmal nach gescheitertem Write oder Commit, ein Fehler von Abort ändert nichts.
func (d *Dienst) ablegen(ctx context.Context, b []byte) (string, error) {
	wr, err := d.k.Blobs.Begin(ctx, math.MaxInt64)
	if err != nil {
		return "", mitCode(err, outbound.GegenstandSonstiges)
	}
	if _, err := wr.Write(b); err != nil {
		_ = wr.Abort()
		return "", mitCode(err, outbound.GegenstandSonstiges)
	}
	ref, _, err := wr.Commit()
	if err != nil {
		_ = wr.Abort()
		return "", mitCode(err, outbound.GegenstandSonstiges)
	}
	return ref, nil
}

// List liefert die Workspace-Snapshots des Workspace workspaceID in der Folge des Ports (SPEC-129 Nr. 23);
// ohne Snapshot eine leere Liste.
func (d *Dienst) List(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error) {
	if ctx == nil {
		return nil, ctxNil()
	}
	l, err := d.k.Snapshots.ListWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandWorkspace)
	}
	if l == nil {
		l = []model.WorkspaceSnapshot{}
	}
	return l, nil
}
