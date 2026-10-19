package outboundtest

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// RunWorkspaceSnapshotVertrag prüft die Workspace-Operationen des SnapshotRepository, SPEC-108
// Nr. 13, 14 und 17, und die Kopfzusagen: Argument vor Kontext, leere Kennung, Zeiten in UTC,
// Listen nie nil.
func RunWorkspaceSnapshotVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"CreateWorkspace und GetWorkspace tragen alle Felder", wsSnapshotFall1},
		{"CreateWorkspace lehnt Ungültiges ab", wsSnapshotFall2},
		{"Prüfreihenfolge von CreateWorkspace", wsSnapshotFall3},
		{"der Name ist je Workspace eindeutig", wsSnapshotFall4},
		{"GetWorkspace bei Unbekanntem, Fremdem und Baselines", wsSnapshotFall5},
		{"ListWorkspace ordnet nach Zeit und Kennung", wsSnapshotFall6},
		{"Baselines und Workspace-Snapshots bleiben getrennt", wsSnapshotFall7},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", wsSnapshotKontextFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func wsSnapshotFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	s := arbeitsraumSnapshot("sn", "d", "w", "Vorher", 0)
	s.CreatedAt = s.CreatedAt.In(time.FixedZone("p5", 5*3600))
	ok(t, v.Snapshots.CreateWorkspace(ctx, s))
	got, err := v.Snapshots.GetWorkspace(ctx, "w", "sn")
	ok(t, err)
	utcZeit(t, got.CreatedAt, s.CreatedAt)
	got.CreatedAt = s.CreatedAt
	if got != s {
		t.Fatalf("gelesen %+v, geschrieben %+v", got, s)
	}
	l, err := v.Snapshots.ListWorkspace(ctx, "w")
	ok(t, err)
	utcZeit(t, l[0].CreatedAt, s.CreatedAt)
}

func wsSnapshotFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	fehlerhaft := map[string]func(*model.WorkspaceSnapshot){
		"ohne Kennung":              func(s *model.WorkspaceSnapshot) { s.ID = "" },
		"ohne Service":              func(s *model.WorkspaceSnapshot) { s.ServiceID = "" },
		"ohne Workspace":            func(s *model.WorkspaceSnapshot) { s.WorkspaceID = "" },
		"ohne Name":                 func(s *model.WorkspaceSnapshot) { s.Name = "" },
		"ohne state_hash":           func(s *model.WorkspaceSnapshot) { s.StateHash = "" },
		"ohne metadata_hash":        func(s *model.WorkspaceSnapshot) { s.MetadataHash = "" },
		"state_blob leer":           func(s *model.WorkspaceSnapshot) { s.StateBlob = "" },
		"state_blob ungültig":       func(s *model.WorkspaceSnapshot) { s.StateBlob = "sha256:abc" },
		"state_blob Großbuchstaben": func(s *model.WorkspaceSnapshot) { s.StateBlob = ref("F") },
		"state_blob ohne Präfix":    func(s *model.WorkspaceSnapshot) { s.StateBlob = ref("a")[len("sha256:"):] },
		"ohne Zeit":                 func(s *model.WorkspaceSnapshot) { s.CreatedAt = time.Time{} },
		"Jahr 10000":                func(s *model.WorkspaceSnapshot) { s.CreatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		"Jahr 0":                    func(s *model.WorkspaceSnapshot) { s.CreatedAt = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
	}
	for name, mut := range fehlerhaft {
		s := arbeitsraumSnapshot("sn", "d", "w", "n", 0)
		mut(&s)
		if err := v.Snapshots.CreateWorkspace(ctx, s); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if l, err := v.Snapshots.ListWorkspace(ctx, "w"); err != nil || len(l) != 0 {
		t.Fatalf("Ungültiges wurde gespeichert: %v %v", l, err)
	}
}

func wsSnapshotFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsMitSnapshots(t, neu)
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn", "d", "w1", "Name", 1)))
	for _, f := range []struct {
		name                      string
		id, dienst, workspace, nm string
		hash                      string
		soll                      error
	}{
		{"ungültig vor unbekanntem Workspace", "x", "d", "gibt-es-nicht", "n", "", model.ErrUngueltig},
		{"unbekannter Workspace vor fremdem Service", "x", "e", "gibt-es-nicht", "n", "h", outbound.ErrNichtGefunden},
		{"unbekannter Workspace vor vorhandener Kennung", "sn", "d", "gibt-es-nicht", "n", "h", outbound.ErrNichtGefunden},
		{"fremder Service vor vorhandener Kennung", "sn", "e", "w1", "n", "h", outbound.ErrFremderService},
		{"fremder Service vor vorhandenem Namen", "x", "e", "w1", "Name", "h", outbound.ErrFremderService},
		{"vorhandene Kennung", "sn", "d", "w1", "anderer", "h", outbound.ErrBereitsVorhanden},
		{"vorhandene Kennung einer Baseline", "snap-1", "d", "w1", "anderer", "h", outbound.ErrBereitsVorhanden},
		{"vorhandener Name", "x", "d", "w1", "Name", "h", outbound.ErrBereitsVorhanden},
	} {
		s := arbeitsraumSnapshot(f.id, f.dienst, f.workspace, f.nm, 2)
		s.StateHash = f.hash
		if err := v.Snapshots.CreateWorkspace(ctx, s); !errors.Is(err, f.soll) {
			t.Fatalf("%s: Fehler %v, erwartet %v", f.name, err, f.soll)
		}
	}
	got, err := v.Snapshots.GetWorkspace(ctx, "w1", "sn")
	ok(t, err)
	if got.Name != "Name" || got.StateHash != "sh-sn" {
		t.Fatalf("abgelehntes CreateWorkspace änderte den Snapshot: %+v", got)
	}
	if l, _ := v.Snapshots.ListWorkspace(ctx, "w1"); len(l) != 2 {
		t.Fatalf("%d Snapshots in w1, erwartet 2", len(l))
	}
}

func wsSnapshotFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w", "w2")
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("a", "d", "w", "Name", 0)))
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("b", "d", "w2", "Name", 0)))
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("c", "d", "w", "name", 0)))
	erwarte(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("d", "d", "w2", "Name", 0)), outbound.ErrBereitsVorhanden)
}

func wsSnapshotFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsMitSnapshots(t, neu)
	for _, k := range [][2]string{{"w1", "gibt-es-nicht"}, {"we", "ws-d"}, {"gibt-es-nicht", "ws-d"}, {"", "ws-d"}, {"w1", ""}, {"", ""}} {
		_, err := v.Snapshots.GetWorkspace(ctx, k[0], k[1])
		erwarte(t, err, outbound.ErrNichtGefunden)
	}
	for _, w := range []string{"w1", "we", "gibt-es-nicht", ""} {
		_, err := v.Snapshots.GetWorkspace(ctx, w, "snap-1")
		erwarte(t, err, outbound.ErrFalscheArt)
	}
	if got, err := v.Snapshots.GetWorkspace(ctx, "w1", "ws-d"); err != nil || got.ID != "ws-d" {
		t.Fatalf("GetWorkspace: %+v %v", got, err)
	}
}

func wsSnapshotFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsMitSnapshots(t, neu)
	zeiten := []struct {
		id string
		t  time.Time
	}{
		{"gleich", t0()},
		{"b", t0().Add(500 * time.Millisecond)},
		{"a", t0()},
		{"c", t0().Add(100 * time.Millisecond)},
		// 09:00 in UTC-5 sind 14:00 UTC: die Uhrzeit des Texts läge vor 10:00, die Zeit danach.
		{"zone", time.Date(2026, 10, 19, 9, 0, 0, 0, time.FixedZone("m5", -5*3600))},
	}
	for _, z := range zeiten {
		s := arbeitsraumSnapshot(z.id, "d", "w1", "n-"+z.id, 0)
		s.CreatedAt = z.t
		ok(t, v.Snapshots.CreateWorkspace(ctx, s))
	}
	l, err := v.Snapshots.ListWorkspace(ctx, "w1")
	ok(t, err)
	var ids []string
	for _, s := range l {
		ids = append(ids, s.ID)
	}
	// ws-d trägt t0() und steht deshalb zwischen a und gleich.
	if want := []string{"a", "gleich", "ws-d", "c", "b", "zone"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("Reihenfolge %v, erwartet %v", ids, want)
	}
	for _, w := range []string{"gibt-es-nicht", ""} {
		_, err = v.Snapshots.ListWorkspace(ctx, w)
		erwarte(t, err, outbound.ErrNichtGefunden)
	}
	legeArbeitsraeumeAn(t, v, "leer")
	l, err = v.Snapshots.ListWorkspace(ctx, "leer")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("Workspace ohne Snapshots: %v %v", l, err)
	}
}

func wsSnapshotFall7(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsMitSnapshots(t, neu)
	_, err := v.Snapshots.Get(ctx, "ws-d")
	erwarte(t, err, outbound.ErrNichtGefunden)
	l, err := v.Snapshots.List(ctx, "d")
	ok(t, err)
	if len(l) != 1 || l[0].ID != "snap-1" {
		t.Fatalf("Baselines von d: %+v", l)
	}
	if l, err := v.Snapshots.ListWorkspace(ctx, "w1"); err != nil || len(l) != 1 || l[0].ID != "ws-d" {
		t.Fatalf("Workspace-Snapshots von w1: %+v %v", l, err)
	}
}

func wsSnapshotKontextFall(t *testing.T, neu Neu) {
	bg, ctx := context.Background(), abgebrochen()
	v := wsMitSnapshots(t, neu)
	erwarte(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("neu", "d", "w1", "neu", 3)), context.Canceled)
	_, err := v.Snapshots.GetWorkspace(ctx, "w1", "ws-d")
	erwarte(t, err, context.Canceled)
	_, err = v.Snapshots.ListWorkspace(ctx, "w1")
	erwarte(t, err, context.Canceled)
	erwarte(t, v.Snapshots.CreateWorkspace(ctx, model.WorkspaceSnapshot{}), model.ErrUngueltig)
	if l, err := v.Snapshots.ListWorkspace(bg, "w1"); err != nil || len(l) != 1 {
		t.Fatalf("CreateWorkspace trotz Abbruch: %v %v", l, err)
	}
}
