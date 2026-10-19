package sqlite_test

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func TestWorkspaceZustandSchreibtArtZeitformatUndHaeltDieSchluessel(t *testing.T) {
	ctx := context.Background()
	d := open(t, filepath.Join(t.TempDir(), "r.db"))
	if err := d.Metadata().Create(ctx, model.Service{ID: "d", ServiceRoot: "/", ODataVersion: "4.0"}); err != nil {
		t.Fatal(err)
	}
	zeit := time.Date(2026, 10, 19, 10, 0, 0, 5, time.UTC)
	versetzt := zeit.In(time.FixedZone("p5", 5*3600))
	if err := d.Sessions().Create(ctx, model.Session{ID: "s", ServiceID: "d", CreatedAt: zeit, Status: model.SessionOpen}); err != nil {
		t.Fatal(err)
	}
	blob := "sha256:" + strings.Repeat("a", 64)
	p := model.Publication{
		Import:           model.Import{ID: "i", RequestHash: "r", BaselineSnapshotID: "b", ServiceID: "d", CreatedAt: zeit, ReportJSON: "{}"},
		Snapshot:         model.BaselineSnapshot{ID: "b", ServiceID: "d", CreatedAt: zeit, StateBlob: blob, StateHash: "h", MetadataHash: "m"},
		SourceSessionIDs: []string{"s"},
	}
	if _, _, err := d.Imports().Publish(ctx, p); err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(d.Workspaces().Create(ctx, model.Workspace{ID: "w", ServiceID: "d", BaselineSnapshotID: "b", CreatedAt: versetzt}))
	must(d.Entities().Put(ctx, model.EntityState{WorkspaceID: "w", EntitySet: "P", EntityKey: "1", Payload: "{ }",
		KnownProperties: []string{"b", "a"}, Completeness: model.EntityPartial, Deleted: true}))
	must(d.Cursors().Put(ctx, model.ReplayCursor{WorkspaceID: "w", ReplaySessionID: "r", RecordingSessionID: "s", ServiceID: "d",
		Strategy: model.StrategyFirst, Position: 2, Consumed: []int64{9, 2}}))
	must(d.Snapshots().CreateWorkspace(ctx, model.WorkspaceSnapshot{ID: "ws", ServiceID: "d", WorkspaceID: "w", Name: "n",
		CreatedAt: versetzt, StateBlob: blob, StateHash: "h", MetadataHash: "m"}))
	if _, err := d.Workspaces().Advance(ctx, "w", 0); err != nil {
		t.Fatal(err)
	}
	must(d.Cursors().ResetAll(ctx, "w"))
	for q, soll := range map[string]string{
		"SELECT created_at FROM workspaces":                   "2026-10-19T10:00:00.000000005Z",
		"SELECT created_at FROM snapshots WHERE id = 'ws'":    "2026-10-19T10:00:00.000000005Z",
		"SELECT baseline_kind FROM workspaces":                "baseline",
		"SELECT revision FROM workspaces":                     "1",
		"SELECT snapshot_kind FROM snapshots WHERE id = 'ws'": "workspace",
		"SELECT known_properties_json FROM entities":          `["a","b"]`,
		"SELECT deleted FROM entities":                        "1",
		"SELECT position FROM replay_cursors":                 "0",
		"SELECT consumed_json FROM replay_cursors":            "[]",
		"PRAGMA integrity_check":                              "ok",
	} {
		if got := skalar(t, d.SQL(), q); got != soll {
			t.Errorf("%s: %q, erwartet %q", q, got, soll)
		}
	}
	if b, err := d.Maintenance().Check(ctx); err != nil || len(b) != 0 {
		t.Errorf("Check: %v %v", b, err)
	}
}

// TestRundlauf schreibt alle Repositories in eine Datei, schließt sie, öffnet sie wieder und
// liest zurück: Service, Session, Exchange (Header, Response-Blob), Baseline, Import, Coverage,
// Workspace, Entities, Cursor und Workspace-Snapshot samt Hashes; dazu Check und ReferencedBlobs.
func TestRundlauf(t *testing.T) {
	ctx := context.Background()
	pfad := filepath.Join(t.TempDir(), "r.db")
	d := open(t, pfad)
	ref := func(z string) string { return "sha256:" + strings.Repeat(z, 64) }
	t0 := time.Date(2026, 10, 19, 10, 0, 0, 123456789, time.UTC)
	dienst := model.Service{ID: "d", ServiceRoot: "https://example.test/odata", ODataVersion: "4.0", MetadataBlob: ref("1"), MetadataHash: "mh"}
	session := model.Session{ID: "s", ServiceID: "d", CreatedAt: t0, Status: model.SessionOpen}
	exchange := model.Exchange{ID: "e", SessionID: "s", SequenceNo: 1, Method: "GET", Path: "/P", RawQuery: "$top=1", MatchKey: "sha256:k",
		StartedAt: t0, FinishedAt: t0.Add(time.Second), RequestHeaders: map[string][]string{"Accept": {"a"}}, RequestBody: model.BodyPresent,
		RequestBlob: ref("2"), ResponseStatus: 200, ResponseHeaders: map[string][]string{"X": {"y"}}, ResponseBody: model.BodyPresent,
		ResponseBlob: ref("3"), Capture: model.CaptureComplete, RequestRedact: model.RedactionNone, ResponseRedact: model.RedactionNone,
		Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte}
	publikation := model.Publication{
		Import:           model.Import{ID: "i", RequestHash: "r", BaselineSnapshotID: "b", ServiceID: "d", CreatedAt: t0, ReportJSON: `{"ok":true}`},
		Snapshot:         model.BaselineSnapshot{ID: "b", ServiceID: "d", CreatedAt: t0, StateBlob: ref("4"), StateHash: "sh", MetadataHash: "mh"},
		SourceSessionIDs: []string{"s"},
		Coverage:         []model.Coverage{{EntitySet: "P", Status: model.CoveragePartial, EvidenceJSON: `{"seiten": 1}`}},
	}
	workspace := model.Workspace{ID: "w", ServiceID: "d", BaselineSnapshotID: "b", CreatedAt: t0}
	entitaeten := []model.EntityState{
		{WorkspaceID: "w", EntitySet: "P", EntityKey: "1", Payload: "{ \"a\" : 1 }", KnownProperties: []string{"a", "b"}, Completeness: model.EntityComplete, Version: 3},
		{WorkspaceID: "w", EntitySet: "P", EntityKey: "2", Payload: "null", KnownProperties: []string{}, Completeness: model.EntityPartial, Deleted: true},
	}
	cursor := model.ReplayCursor{WorkspaceID: "w", ReplaySessionID: "r", RecordingSessionID: "s", ServiceID: "d",
		Strategy: model.StrategyLast, Position: 2, Consumed: []int64{1, 4}}
	wsSnapshot := model.WorkspaceSnapshot{ID: "ws", ServiceID: "d", WorkspaceID: "w", Name: "vorher", CreatedAt: t0, StateBlob: ref("5"), StateHash: "sh", MetadataHash: "mh"}

	fehler := []error{
		d.Metadata().Create(ctx, dienst),
		d.Sessions().Create(ctx, session),
		d.Recording().Append(ctx, exchange),
		d.Sessions().SetStatus(ctx, "s", model.SessionComplete),
	}
	_, _, err := d.Imports().Publish(ctx, publikation)
	fehler = append(fehler, err, d.Workspaces().Create(ctx, workspace), d.Entities().ReplaceAll(ctx, "w", entitaeten),
		d.Cursors().Put(ctx, cursor), d.Snapshots().CreateWorkspace(ctx, wsSnapshot))
	_, err = d.Workspaces().Advance(ctx, "w", 0)
	fehler = append(fehler, err)
	for i, err := range fehler {
		if err != nil {
			t.Fatalf("Schritt %d: %v", i, err)
		}
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}

	d = open(t, pfad)
	if b, err := d.Maintenance().Check(ctx); err != nil || b == nil || len(b) != 0 {
		t.Fatalf("Check nach dem Wiederöffnen: %v %v", b, err)
	}
	if got, err := d.Metadata().Get(ctx, "d"); err != nil || got != dienst {
		t.Errorf("Service %+v %v", got, err)
	}
	if got, err := d.Sessions().Get(ctx, "s"); err != nil || got.Status != model.SessionComplete || !got.CreatedAt.Equal(t0) {
		t.Errorf("Session %+v %v", got, err)
	}
	if got, err := d.Recording().Get(ctx, "e"); err != nil || !reflect.DeepEqual(got.RequestHeaders, exchange.RequestHeaders) || got.ResponseBlob != ref("3") {
		t.Errorf("Exchange %+v %v", got, err)
	}
	if got, err := d.Snapshots().Get(ctx, "b"); err != nil || got != publikation.Snapshot {
		t.Errorf("Baseline %+v %v", got, err)
	}
	if got, err := d.Imports().Get(ctx, "i"); err != nil || got != publikation.Import {
		t.Errorf("Import %+v %v", got, err)
	}
	if got, err := d.Coverage().List(ctx, "b"); err != nil || !reflect.DeepEqual(got, publikation.Coverage) {
		t.Errorf("Coverage %+v %v", got, err)
	}
	pruefeWorkspaceRundlauf(t, d, workspace, entitaeten, cursor, wsSnapshot)
	if got, err := d.Maintenance().ReferencedBlobs(ctx); err != nil || !reflect.DeepEqual(got, []string{ref("1"), ref("2"), ref("3"), ref("4"), ref("5")}) {
		t.Errorf("ReferencedBlobs %v %v", got, err)
	}
}

func pruefeWorkspaceRundlauf(t *testing.T, d *sqlite.DB, w model.Workspace, e []model.EntityState, c model.ReplayCursor, sn model.WorkspaceSnapshot) {
	t.Helper()
	ctx := context.Background()
	gotW, err := d.Workspaces().Get(ctx, "w")
	if err != nil || gotW.Revision != 1 || !gotW.CreatedAt.Equal(w.CreatedAt) || gotW.BaselineSnapshotID != "b" {
		t.Errorf("Workspace %+v %v", gotW, err)
	}
	e[0].KnownProperties = []string{"a", "b"}
	if got, err := d.Entities().List(ctx, "w"); err != nil || !reflect.DeepEqual(got, e) {
		t.Errorf("Entities %+v %v", got, err)
	}
	if got, err := d.Entities().Get(ctx, "w", "P", "2"); err != nil || !got.Deleted || got.Payload != "null" {
		t.Errorf("Tombstone %+v %v", got, err)
	}
	c.Consumed = []int64{1, 4}
	if got, err := d.Cursors().Get(ctx, "w", "r", "s"); err != nil || !reflect.DeepEqual(got, c) {
		t.Errorf("Cursor %+v %v", got, err)
	}
	got, err := d.Snapshots().GetWorkspace(ctx, "w", "ws")
	if err != nil || !got.CreatedAt.Equal(sn.CreatedAt) || got.Name != "vorher" || got.StateBlob != sn.StateBlob || got.StateHash != sn.StateHash || got.MetadataHash != sn.MetadataHash {
		t.Errorf("Workspace-Snapshot %+v %v", got, err)
	}
}
