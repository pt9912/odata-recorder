package workspace_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestRestoreRevisionVeraltet (T82): ErrRevisionVeraltet aus Advance ist bei Restore ODR-E9001 mit der Ursache, und die
// Transaktion rollt zurück (State, Revision, Cursors wie vorher).
func TestRestoreRevisionVeraltet(t *testing.T) {
	w := restoreWelt(t)
	vorher := w.st.stand()
	w.p.setzeFehler("tx Workspaces.Advance", outbound.ErrRevisionVeraltet)
	_, err := restoreLauf(w)
	pruefeCode(t, err, model.CodeE9001, outbound.ErrRevisionVeraltet)
	if !zustaendeGleich(vorher, w.st.stand()) {
		t.Error("Zustand geändert")
	}
}

// TestRestoreSnapshotsBleiben (T31, SPEC-129 Nr. 17): die Workspace-Snapshots bleiben, auch der wiederhergestellte; Restore
// löscht weder Workspace noch Snapshot (Workspaces.Delete 0) und legt keinen an (CreateWorkspace ist verboten und zählt im
// Aufräumschritt). Menge: die Snapshots beider Workspaces, 2 + 1 Einträge.
func TestRestoreSnapshotsBleiben(t *testing.T) {
	w := restoreWelt(t)
	w.snapshotAblegen("ws-1", "snp-2", snapshotState("ws-1"))
	w.vorbelegen("ws-2", 0)
	w.snapshotAblegen("ws-2", "snp-3", snapshotState("ws-2"))
	vorher := map[[2]string]model.WorkspaceSnapshot{}
	for k, s := range w.snaps.wsSnaps {
		vorher[k] = s
	}
	if len(vorher) != 3 {
		t.Fatalf("Fixture: %d Snapshots", len(vorher))
	}
	w.p.leeren()
	if _, err := restoreLauf(w); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(vorher, w.snaps.wsSnaps) || w.p.anzahl("Workspaces.Delete")+w.p.anzahl("tx Workspaces.Delete") != 0 {
		t.Errorf("Snapshots oder Workspace verändert: %v", w.snaps.wsSnaps)
	}
	if _, da := w.st.stand().ws["ws-1"]; !da {
		t.Error("der Workspace fehlt")
	}
}
