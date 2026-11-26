package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestAdminWorkspacesVerdrahtet fängt A05: je Modus die acht Handler gegen die Ablage und audit.jsonl (8 x 4 Zellen).
func TestAdminWorkspacesVerdrahtet(t *testing.T) {
	ctx := context.Background()
	const basis, je = "/__admin/workspaces", "/__admin/workspaces/{id}"
	zellen := 0
	for _, modus := range adminModi() {
		dir := t.TempDir()
		saeen(t, dir, func(d *wsDatei) { saeeWorkspaces(d) })
		h, db := adminBau(t, routenKonfig(dir, modus, "http://u1.test/o"), nil, nil)
		revision := func(soll int64) {
			t.Helper()
			if w, err := db.Workspaces().Get(ctx, "ws-a"); err != nil || w.Revision != soll || w.BaselineSnapshotID != "bsl-1" {
				t.Errorf("%s: Workspace %+v, %v; erwartet Revision %d", modus, w, err, soll)
			}
		}
		enthaelt := func(body, teil string) {
			t.Helper()
			if !strings.Contains(body, teil) {
				t.Errorf("%s: %s ohne %s", modus, body, teil)
			}
		}
		rufe(t, h, 201, "POST", basis, "", "", `{"baseline_snapshot_id":"bsl-1","id":"ws-a"}`)
		w, err := db.Workspaces().Get(ctx, "ws-a")
		if err != nil {
			t.Fatalf("%s: %v", modus, err)
		}
		enthaelt(rufe(t, h, 200, "GET", basis, "", "", ""), `"ws-a"`)
		enthaelt(rufe(t, h, 200, "GET", je, "ws-a", "", ""), `"ws-a"`)
		var snapshot struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(rufe(t, h, 201, "POST", je+"/snapshots", "ws-a", "", `{"name":"s1"}`)), &snapshot); err != nil {
			t.Fatal(err)
		}
		if s, err := db.Snapshots().ListWorkspace(ctx, "ws-a"); err != nil || len(s) != 1 || s[0].ID != snapshot.ID {
			t.Errorf("%s: Snapshots %+v, %v", modus, s, err)
		}
		enthaelt(rufe(t, h, 200, "GET", je+"/snapshots", "ws-a", "", ""), `"s1"`)
		rufe(t, h, 200, "POST", je+"/reset", "ws-a", "", "")
		revision(w.Revision + 1)
		rufe(t, h, 200, "POST", je+"/restore", "ws-a", "", `{"snapshot_id":"`+snapshot.ID+`"}`)
		revision(w.Revision + 2)
		rufe(t, h, 200, "DELETE", je, "ws-a", "", "")
		if _, err := db.Workspaces().Get(ctx, "ws-a"); !errors.Is(err, outbound.ErrNichtGefunden) {
			t.Errorf("%s: Workspace nach DELETE: %v", modus, err)
		}
		var arten []string
		for _, e := range auditEintraege(t, dir) {
			arten = append(arten, e["art"])
			if e["workspace_id"] != "ws-a" {
				t.Errorf("%s: Eintrag %v", modus, e)
			}
		}
		soll := []string{"workspace_angelegt", "snapshot_erzeugt", "workspace_zurueckgesetzt", "workspace_wiederhergestellt", "workspace_geloescht"}
		if !slices.Equal(arten, soll) {
			t.Errorf("%s: Arten %v", modus, arten)
		}
		zellen += 8
	}
	if zellen != 32 {
		t.Errorf("%d Zellen", zellen)
	}
}
