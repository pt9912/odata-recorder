package outboundtest

import (
	"context"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// InTransaktion liefert die Verbindung, deren Workspace-, Entity-, Cursor- und
// Snapshot-Repositories jede Operation in einer eigenen WithWorkspaceTransaction auf den
// Repositories der Transaktion ausführen. Dieselben Verträge laufen so gegen beide Wege. Die
// Transaktion öffnet ohne den Abbruch des Kontexts, die Operation läuft mit ihm.
func InTransaktion(v Verbindung) Verbindung {
	t := txLauf{tx: v.Transaktion}
	v.Workspaces = txWorkspaces{t}
	v.Entities = txEntities{t}
	v.Cursors = txCursors{t}
	v.Snapshots = txSnapshots{t}
	return v
}

type txLauf struct{ tx outbound.WorkspaceTransactor }

func lauf[T any](ctx context.Context, l txLauf, f func(outbound.WorkspaceStores) (T, error)) (T, error) {
	var out T
	err := l.tx.WithWorkspaceTransaction(context.WithoutCancel(ctx), func(s outbound.WorkspaceStores) error {
		var err error
		out, err = f(s)
		return err
	})
	return out, err
}

func laufOhne(ctx context.Context, l txLauf, f func(outbound.WorkspaceStores) error) error {
	_, err := lauf(ctx, l, func(s outbound.WorkspaceStores) (struct{}, error) { return struct{}{}, f(s) })
	return err
}

type txWorkspaces struct{ txLauf }

func (r txWorkspaces) Create(ctx context.Context, w model.Workspace) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Workspaces.Create(ctx, w) })
}

func (r txWorkspaces) Get(ctx context.Context, id string) (model.Workspace, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) (model.Workspace, error) { return s.Workspaces.Get(ctx, id) })
}

func (r txWorkspaces) List(ctx context.Context, serviceID string) ([]model.Workspace, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) ([]model.Workspace, error) { return s.Workspaces.List(ctx, serviceID) })
}

func (r txWorkspaces) Advance(ctx context.Context, id string, erwartet int64) (int64, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) (int64, error) { return s.Workspaces.Advance(ctx, id, erwartet) })
}

func (r txWorkspaces) Delete(ctx context.Context, id string) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Workspaces.Delete(ctx, id) })
}

type txEntities struct{ txLauf }

func (r txEntities) Put(ctx context.Context, e model.EntityState) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Entities.Put(ctx, e) })
}

func (r txEntities) Get(ctx context.Context, workspaceID, entitySet, entityKey string) (model.EntityState, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) (model.EntityState, error) {
		return s.Entities.Get(ctx, workspaceID, entitySet, entityKey)
	})
}

func (r txEntities) List(ctx context.Context, workspaceID string) ([]model.EntityState, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) ([]model.EntityState, error) {
		return s.Entities.List(ctx, workspaceID)
	})
}

func (r txEntities) ListSet(ctx context.Context, workspaceID, entitySet string) ([]model.EntityState, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) ([]model.EntityState, error) {
		return s.Entities.ListSet(ctx, workspaceID, entitySet)
	})
}

func (r txEntities) ReplaceAll(ctx context.Context, workspaceID string, eintraege []model.EntityState) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Entities.ReplaceAll(ctx, workspaceID, eintraege) })
}

type txCursors struct{ txLauf }

func (r txCursors) Put(ctx context.Context, c model.ReplayCursor) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Cursors.Put(ctx, c) })
}

func (r txCursors) Get(ctx context.Context, workspaceID, replaySessionID, recordingSessionID string) (model.ReplayCursor, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) (model.ReplayCursor, error) {
		return s.Cursors.Get(ctx, workspaceID, replaySessionID, recordingSessionID)
	})
}

func (r txCursors) List(ctx context.Context, workspaceID string) ([]model.ReplayCursor, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) ([]model.ReplayCursor, error) {
		return s.Cursors.List(ctx, workspaceID)
	})
}

func (r txCursors) ResetAll(ctx context.Context, workspaceID string) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Cursors.ResetAll(ctx, workspaceID) })
}

type txSnapshots struct{ txLauf }

func (r txSnapshots) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) (model.BaselineSnapshot, error) { return s.Snapshots.Get(ctx, id) })
}

func (r txSnapshots) List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) ([]model.BaselineSnapshot, error) {
		return s.Snapshots.List(ctx, serviceID)
	})
}

func (r txSnapshots) CreateWorkspace(ctx context.Context, sn model.WorkspaceSnapshot) error {
	return laufOhne(ctx, r.txLauf, func(s outbound.WorkspaceStores) error { return s.Snapshots.CreateWorkspace(ctx, sn) })
}

func (r txSnapshots) GetWorkspace(ctx context.Context, workspaceID, id string) (model.WorkspaceSnapshot, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) (model.WorkspaceSnapshot, error) {
		return s.Snapshots.GetWorkspace(ctx, workspaceID, id)
	})
}

func (r txSnapshots) ListWorkspace(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error) {
	return lauf(ctx, r.txLauf, func(s outbound.WorkspaceStores) ([]model.WorkspaceSnapshot, error) {
		return s.Snapshots.ListWorkspace(ctx, workspaceID)
	})
}
