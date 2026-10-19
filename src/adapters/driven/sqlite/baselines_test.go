package sqlite_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func TestPublishSchreibtArtZeitformatUndHaeltDieSchluessel(t *testing.T) {
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
	_, neu, err := d.Imports().Publish(ctx, model.Publication{
		Import:           model.Import{ID: "i", RequestHash: "r", BaselineSnapshotID: "b", ServiceID: "d", CreatedAt: versetzt, ReportJSON: "{}"},
		Snapshot:         model.BaselineSnapshot{ID: "b", ServiceID: "d", CreatedAt: versetzt, StateBlob: blob, StateHash: "h", MetadataHash: "m"},
		SourceSessionIDs: []string{"s"},
		Coverage:         []model.Coverage{{EntitySet: "P", Status: model.CoverageComplete, EvidenceJSON: "{}"}},
	})
	if err != nil || !neu {
		t.Fatalf("Publish: %v, neu=%v", err, neu)
	}
	const text = "2026-10-19T10:00:00.000000005Z"
	for q, soll := range map[string]string{
		"SELECT snapshot_kind FROM snapshots":        "baseline",
		"SELECT created_at FROM snapshots":           text,
		"SELECT created_at FROM imports":             text,
		"SELECT baseline_kind FROM baseline_sources": "baseline",
		"SELECT baseline_kind FROM coverage":         "baseline",
		"PRAGMA integrity_check":                     "ok",
	} {
		if got := skalar(t, d.SQL(), q); got != soll {
			t.Errorf("%s: %q, erwartet %q", q, got, soll)
		}
	}
	rows, err := d.SQL().Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		t.Error("foreign_key_check meldet eine Verletzung")
	}
}
