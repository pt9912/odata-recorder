package workspace_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// snapshotMetadatenHash ist der MetadataHash der Workspace-Snapshots der Fixtures; er steht im Kopf des Manifests und
// ist von allen anderen Werten verschieden.
func snapshotMetadatenHash() string { return "sha256:" + strings.Repeat("d", 64) }

// snapshotState ist der Entity-State, den die Snapshots der Fixtures tragen; er gleicht weder dem der Baseline noch dem
// des veränderten Workspace (vier Werte: vorher, im Snapshot, nach der Mutation, danach).
func snapshotState(wid string) []model.EntityState {
	return []model.EntityState{
		{WorkspaceID: wid, EntitySet: "Products", EntityKey: "Id=1", Payload: `{"Id":"1","Name":"Snap"}`,
			KnownProperties: []string{"Id", "Name"}, Completeness: model.EntityComplete, Version: 4},
		{WorkspaceID: wid, EntitySet: "Products", EntityKey: "Id=2", Payload: `{"Id":"2"}`,
			KnownProperties: []string{"Id"}, Completeness: model.EntityPartial, Version: 2, Deleted: true},
		{WorkspaceID: wid, EntitySet: "Orders", EntityKey: "Id=7", Payload: `{"Id":"7","Total":3}`,
			KnownProperties: []string{"Id", "Total"}, Completeness: model.EntityComplete, Version: 1},
	}
}

// snapshotEintragBaum ist ein Eintrag des Manifests eines Workspace-Snapshots (SPEC-129 Nr. 20), unabhängig vom Dienst
// geschrieben.
func snapshotEintragBaum(e model.EntityState) obj {
	props := make([]any, 0, len(e.KnownProperties))
	for _, p := range e.KnownProperties {
		props = append(props, p)
	}
	return obj{"completeness": string(e.Completeness), "deleted": e.Deleted, "entity_set": e.EntitySet, "payload": roh(e.Payload),
		"selected_properties": props, "typed_key": e.EntityKey, "version": int(e.Version)}
}

// snapshotManifestBaum ist das Manifest eines Workspace-Snapshots des Service svc; die Einträge stehen in der
// Folge von es (der Leser verlangt keine Sortierung).
func snapshotManifestBaum(svc string, es []model.EntityState) obj {
	l := make([]any, 0, len(es))
	for _, e := range es {
		l = append(l, snapshotEintragBaum(e))
	}
	return obj{"entities": l, "entity_count": len(es), "manifest_version": 1, "metadata_hash": snapshotMetadatenHash(),
		"service_id": svc, "snapshot_kind": "workspace"}
}

// snapshotBytesAblegen legt die Bytes m als Blob und als Workspace-Snapshot id des Workspace wid ab; die Referenz
// gilt als StateBlob und StateHash, der MetadataHash steht im Kopf des Manifests.
func (w *welt) snapshotBytesAblegen(wid, id string, m []byte) string {
	sum := sha256.Sum256(m)
	ref := "sha256:" + hex.EncodeToString(sum[:])
	w.blobs.inhalt[ref] = m
	w.snaps.wsSnaps[[2]string{wid, id}] = model.WorkspaceSnapshot{ID: id, ServiceID: "svc", WorkspaceID: wid, Name: "n-" + id,
		CreatedAt: t0().Add(1), StateBlob: ref, StateHash: ref, MetadataHash: snapshotMetadatenHash()}
	return ref
}

// snapshotAblegen legt den Snapshot id des Workspace wid mit dem State es ab.
func (w *welt) snapshotAblegen(wid, id string, es []model.EntityState) string {
	return w.snapshotBytesAblegen(wid, id, []byte(schreibe(snapshotManifestBaum("svc", es))))
}

// cursorOps sind alle Operationen von CursorRepository, in und außerhalb der Transaktion.
func cursorOps() []string {
	var r []string
	for _, o := range []string{"Put", "Get", "List", "ResetAll"} {
		r = append(r, "Cursors."+o, "tx Cursors."+o)
	}
	return r
}

// cursorAufrufe zählt die Aufrufe jeder Operation von CursorRepository im Protokoll.
func (w *welt) cursorAufrufe() int {
	n := 0
	for _, o := range cursorOps() {
		n += w.p.anzahl(o)
	}
	return n
}
