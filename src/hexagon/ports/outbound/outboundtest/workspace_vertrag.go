package outboundtest

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// RunWorkspaceVertrag prüft für WorkspaceRepository SPEC-108 Nr. 2 bis 5 und Nr. 17.
func RunWorkspaceVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Create und Get tragen alle Felder", workspaceFall1},
		{"Create lehnt Ungültiges ab und schreibt nichts", workspaceFall2},
		{"Prüfreihenfolge von Create", workspaceFall3},
		{"Get und List", workspaceFall4},
		{"List ordnet nach Zeit, nicht nach Text", workspaceFall5},
		{"Advance zählt und lehnt Veraltetes ab", workspaceFall6},
		{"gleichzeitige Advance mit derselben Revision", workspaceNebenlaeufig},
		{"Delete nimmt Entities, Cursors und Snapshots mit", workspaceFall7},
		{"Delete ist atomar", workspaceDeleteAtomar},
		{"leere Kennung", workspaceFall8},
		{"abgebrochener Kontext beendet jede Operation ohne Änderung", workspaceKontextFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

// wsBasis legt zu publikationBasis die Baselines snap-1 (Service d) und snap-e (Service e) an.
func wsBasis(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	v := publikationBasis(t, neu)
	veroeffentliche(t, v, publikation("d", "1", "s1"))
	veroeffentliche(t, v, publikation("e", "e", "x1"))
	return v
}

func arbeitsraum(id string, minuten int) model.Workspace {
	return model.Workspace{ID: id, ServiceID: "d", BaselineSnapshotID: "snap-1", CreatedAt: t0().Add(time.Duration(minuten) * time.Minute)}
}

// legeArbeitsraeumeAn legt Workspaces in Service d an, jeder eine Minute nach dem vorigen.
func legeArbeitsraeumeAn(t *testing.T, v Verbindung, ids ...string) {
	t.Helper()
	for i, id := range ids {
		ok(t, v.Workspaces.Create(context.Background(), arbeitsraum(id, i)))
	}
}

func utcZeit(t *testing.T, got, soll time.Time) {
	t.Helper()
	if !got.Equal(soll) || got.Location() != time.UTC {
		t.Fatalf("Zeit %v (%v), erwartet %v in UTC", got, got.Location(), soll)
	}
}

func workspaceFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	w := arbeitsraum("w1", 0)
	w.CreatedAt = w.CreatedAt.In(time.FixedZone("p5", 5*3600))
	ok(t, v.Workspaces.Create(ctx, w))
	got, err := v.Workspaces.Get(ctx, "w1")
	ok(t, err)
	utcZeit(t, got.CreatedAt, w.CreatedAt)
	got.CreatedAt = w.CreatedAt
	if got != w || got.Revision != 0 {
		t.Fatalf("gelesen %+v, geschrieben %+v", got, w)
	}
	st, err := v.Entities.List(ctx, "w1")
	ok(t, err)
	if st == nil || len(st) != 0 {
		t.Fatalf("der Entity-State entstand mit: %v", st)
	}
}

func workspaceFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	fehlerhaft := map[string]func(*model.Workspace){
		"ohne Kennung":      func(w *model.Workspace) { w.ID = "" },
		"ohne Service":      func(w *model.Workspace) { w.ServiceID = "" },
		"ohne Baseline":     func(w *model.Workspace) { w.BaselineSnapshotID = "" },
		"ohne Zeit":         func(w *model.Workspace) { w.CreatedAt = time.Time{} },
		"Jahr 10000":        func(w *model.Workspace) { w.CreatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		"Jahr 0":            func(w *model.Workspace) { w.CreatedAt = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC) },
		"Revision 1":        func(w *model.Workspace) { w.Revision = 1 },
		"negative Revision": func(w *model.Workspace) { w.Revision = -1 },
	}
	for name, mut := range fehlerhaft {
		w := arbeitsraum("w", 0)
		mut(&w)
		if err := v.Workspaces.Create(ctx, w); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if l, err := v.Workspaces.List(ctx, "d"); err != nil || len(l) != 0 {
		t.Fatalf("Ungültiges wurde gespeichert: %v %v", l, err)
	}
}

// wsMitSnapshots legt die Workspaces w1 (Service d) und we (Service e) mit den Snapshots
// ws-d und ws-e an.
func wsMitSnapshots(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	ctx := context.Background()
	v := wsBasis(t, neu)
	ok(t, v.Workspaces.Create(ctx, arbeitsraum("w1", 0)))
	we := arbeitsraum("we", 1)
	we.ServiceID, we.BaselineSnapshotID = "e", "snap-e"
	ok(t, v.Workspaces.Create(ctx, we))
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("ws-d", "d", "w1", "n", 0)))
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("ws-e", "e", "we", "n", 0)))
	return v
}

func arbeitsraumSnapshot(id, dienst, workspace, name string, minuten int) model.WorkspaceSnapshot {
	return model.WorkspaceSnapshot{ID: id, ServiceID: dienst, WorkspaceID: workspace, Name: name,
		CreatedAt: t0().Add(time.Duration(minuten) * time.Minute), StateBlob: ref("c"), StateHash: "sh-" + id, MetadataHash: "mh"}
}

func workspaceFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsMitSnapshots(t, neu)
	for _, f := range []struct {
		name                 string
		id, dienst, baseline string
		revision             int64
		soll                 error
	}{
		{"ungültig vor bereits vorhanden", "w1", "d", "snap-1", 1, model.ErrUngueltig},
		{"ungültig vor unbekanntem Service", "w", "x", "snap-1", 1, model.ErrUngueltig},
		{"vorhanden vor unbekanntem Service", "w1", "x", "snap-1", 0, outbound.ErrBereitsVorhanden},
		{"vorhanden vor unbekannter Baseline", "w1", "d", "x", 0, outbound.ErrBereitsVorhanden},
		{"vorhanden vor falscher Art", "w1", "d", "ws-d", 0, outbound.ErrBereitsVorhanden},
		{"vorhanden vor fremdem Service", "w1", "d", "snap-e", 0, outbound.ErrBereitsVorhanden},
		{"Service vor unbekannter Baseline", "w", "x", "y", 0, outbound.ErrDienstUnbekannt},
		{"Service vor falscher Art", "w", "x", "ws-d", 0, outbound.ErrDienstUnbekannt},
		{"Service vor fremdem Service", "w", "x", "snap-1", 0, outbound.ErrDienstUnbekannt},
		{"unbekannte Baseline", "w", "d", "x", 0, outbound.ErrNichtGefunden},
		{"falsche Art", "w", "d", "ws-d", 0, outbound.ErrFalscheArt},
		{"falsche Art vor fremdem Service", "w", "d", "ws-e", 0, outbound.ErrFalscheArt},
		{"fremder Service", "w", "d", "snap-e", 0, outbound.ErrFremderService},
	} {
		w := model.Workspace{ID: f.id, ServiceID: f.dienst, BaselineSnapshotID: f.baseline, Revision: f.revision, CreatedAt: t0()}
		if err := v.Workspaces.Create(ctx, w); !errors.Is(err, f.soll) {
			t.Fatalf("%s: Fehler %v, erwartet %v", f.name, err, f.soll)
		}
	}
	if _, err := v.Workspaces.Get(ctx, "w"); !errors.Is(err, outbound.ErrNichtGefunden) {
		t.Fatalf("abgelehnter Workspace wurde angelegt: %v", err)
	}
	if got, _ := v.Workspaces.Get(ctx, "w1"); got.BaselineSnapshotID != "snap-1" {
		t.Fatalf("doppeltes Create änderte den Workspace: %+v", got)
	}
}

func workspaceFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	_, err := v.Workspaces.Get(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	l, err := v.Workspaces.List(ctx, "d")
	ok(t, err)
	if l == nil || len(l) != 0 {
		t.Fatalf("List ohne Workspaces: %v", l)
	}
	l, err = v.Workspaces.List(ctx, "unbekannt")
	if err != nil || l == nil || len(l) != 0 {
		t.Fatalf("unbekannter Service: %v %v", l, err)
	}
	_, err = v.Workspaces.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
	legeArbeitsraeumeAn(t, v, "w1")
	we := arbeitsraum("we", 0)
	we.ServiceID, we.BaselineSnapshotID = "e", "snap-e"
	ok(t, v.Workspaces.Create(ctx, we))
	if l, err := v.Workspaces.List(ctx, "d"); err != nil || len(l) != 1 || l[0].ID != "w1" {
		t.Fatalf("List d: %v %v", l, err)
	}
}

func workspaceFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
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
		w := arbeitsraum(z.id, 0)
		w.CreatedAt = z.t
		ok(t, v.Workspaces.Create(ctx, w))
	}
	l, err := v.Workspaces.List(ctx, "d")
	ok(t, err)
	var ids []string
	for _, w := range l {
		ids = append(ids, w.ID)
	}
	if want := []string{"a", "gleich", "c", "b", "zone"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("Reihenfolge %v, erwartet %v", ids, want)
	}
}

func workspaceFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w", "anderer")
	for i := range int64(3) {
		neue, err := v.Workspaces.Advance(ctx, "w", i)
		ok(t, err)
		if neue != i+1 {
			t.Fatalf("Advance von %d lieferte %d", i, neue)
		}
	}
	for _, erwartet := range []int64{0, 2, 4, 100} {
		_, err := v.Workspaces.Advance(ctx, "w", erwartet)
		erwarte(t, err, outbound.ErrRevisionVeraltet)
	}
	if got, _ := v.Workspaces.Get(ctx, "w"); got.Revision != 3 {
		t.Fatalf("Revision %d nach abgelehnten Advance", got.Revision)
	}
	if got, _ := v.Workspaces.Get(ctx, "anderer"); got.Revision != 0 {
		t.Fatalf("Advance änderte einen anderen Workspace: %d", got.Revision)
	}
	_, err := v.Workspaces.Advance(ctx, "gibt-es-nicht", 0)
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Workspaces.Advance(ctx, "gibt-es-nicht", 7)
	erwarte(t, err, outbound.ErrNichtGefunden)
	for _, id := range []string{"w", "gibt-es-nicht", ""} {
		_, err = v.Workspaces.Advance(ctx, id, -1)
		erwarte(t, err, model.ErrUngueltig)
		_, err = v.Workspaces.Advance(abgebrochen(), id, -1)
		erwarte(t, err, model.ErrUngueltig)
	}
}

func workspaceNebenlaeufig(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	const n = 8
	fehler := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, fehler[i] = v.Workspaces.Advance(ctx, "w", 0)
		}()
	}
	wg.Wait()
	erfolge := 0
	for _, err := range fehler {
		switch {
		case err == nil:
			erfolge++
		case !errors.Is(err, outbound.ErrRevisionVeraltet):
			t.Fatalf("unerwarteter Fehler %v", err)
		}
	}
	if got, _ := v.Workspaces.Get(ctx, "w"); erfolge != 1 || got.Revision != 1 {
		t.Fatalf("%d Advance gelangen, Revision %d", erfolge, got.Revision)
	}
}

func workspaceFall7(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w", "anderer")
	for _, id := range []string{"w", "anderer"} {
		ok(t, v.Entities.Put(ctx, entity(id, "Products", "1")))
		ok(t, v.Entities.Put(ctx, entity(id, "Products", "2")))
		ok(t, v.Cursors.Put(ctx, cursor(id, "r1", "s1")))
		ok(t, v.Cursors.Put(ctx, cursor(id, "r2", "s2")))
		ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn-"+id, "d", id, "n", 0)))
	}
	zeilen := func(tabelle string) int { return v.Zeilen(t, tabelle) }
	if zeilen("entities") != 4 || zeilen("replay_cursors") != 4 {
		t.Fatalf("Aufbau: %d Entities, %d Cursors", zeilen("entities"), zeilen("replay_cursors"))
	}
	sessions, baselines := zeilen("sessions"), zeilen("snapshots")-2
	ok(t, v.Workspaces.Delete(ctx, "w"))
	for tabelle, soll := range map[string]int{"workspaces": 1, "entities": 2, "replay_cursors": 2, "snapshots": baselines + 1, "sessions": sessions} {
		if got := zeilen(tabelle); got != soll {
			t.Errorf("%s: %d Zeilen, erwartet %d", tabelle, got, soll)
		}
	}
	_, err := v.Workspaces.Get(ctx, "w")
	erwarte(t, err, outbound.ErrNichtGefunden)
	if l, err := v.Entities.List(ctx, "anderer"); err != nil || len(l) != 2 {
		t.Fatalf("Entities des anderen Workspace: %v %v", l, err)
	}
	if l, err := v.Cursors.List(ctx, "anderer"); err != nil || len(l) != 2 {
		t.Fatalf("Cursors des anderen Workspace: %v %v", l, err)
	}
	if _, err := v.Snapshots.GetWorkspace(ctx, "anderer", "sn-anderer"); err != nil {
		t.Fatalf("Snapshot des anderen Workspace: %v", err)
	}
	if _, err := v.Snapshots.Get(ctx, "snap-1"); err != nil {
		t.Fatalf("Baseline nach Delete: %v", err)
	}
	erwarte(t, v.Workspaces.Delete(ctx, "w"), outbound.ErrNichtGefunden)
	ok(t, v.Workspaces.Create(ctx, arbeitsraum("w", 5)))
	if l, err := v.Entities.List(ctx, "w"); err != nil || len(l) != 0 {
		t.Fatalf("State des gelöschten Workspace kehrte zurück: %v %v", l, err)
	}
}

func workspaceDeleteAtomar(t *testing.T, neu Neu) {
	bg := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	ok(t, v.Entities.Put(bg, entity("w", "Products", "1")))
	ok(t, v.Cursors.Put(bg, cursor("w", "r", "s1")))
	ok(t, v.Snapshots.CreateWorkspace(bg, arbeitsraumSnapshot("sn", "d", "w", "n", 0)))
	ctx, stop := context.WithCancel(bg)
	v.BrichNachLoeschenAb(t, "snapshots", stop)
	erwarte(t, v.Workspaces.Delete(ctx, "w"), context.Canceled)
	if _, err := v.Workspaces.Get(bg, "w"); err != nil {
		t.Fatalf("Workspace nach abgebrochenem Delete: %v", err)
	}
	if _, err := v.Snapshots.GetWorkspace(bg, "w", "sn"); err != nil {
		t.Fatalf("Snapshot nach abgebrochenem Delete: %v", err)
	}
	if v.Zeilen(t, "entities") != 1 || v.Zeilen(t, "replay_cursors") != 1 {
		t.Fatalf("Entities und Cursors nach abgebrochenem Delete: %d, %d", v.Zeilen(t, "entities"), v.Zeilen(t, "replay_cursors"))
	}
}

func workspaceFall8(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	_, err := v.Workspaces.Get(ctx, "")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Workspaces.Advance(ctx, "", 0)
	erwarte(t, err, outbound.ErrNichtGefunden)
	erwarte(t, v.Workspaces.Delete(ctx, ""), outbound.ErrNichtGefunden)
}

func workspaceKontextFall(t *testing.T, neu Neu) {
	bg, ctx := context.Background(), abgebrochen()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	erwarte(t, v.Workspaces.Create(ctx, arbeitsraum("neu", 5)), context.Canceled)
	_, err := v.Workspaces.Get(ctx, "w")
	erwarte(t, err, context.Canceled)
	_, err = v.Workspaces.List(ctx, "d")
	erwarte(t, err, context.Canceled)
	_, err = v.Workspaces.Advance(ctx, "w", 0)
	erwarte(t, err, context.Canceled)
	erwarte(t, v.Workspaces.Delete(ctx, "w"), context.Canceled)
	// Das Argument geht dem Kontext voraus.
	_, err = v.Workspaces.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
	erwarte(t, v.Workspaces.Create(ctx, model.Workspace{}), model.ErrUngueltig)
	if _, err := v.Workspaces.Get(bg, "neu"); !errors.Is(err, outbound.ErrNichtGefunden) {
		t.Fatalf("Create trotz Abbruch: %v", err)
	}
	if got, err := v.Workspaces.Get(bg, "w"); err != nil || got.Revision != 0 {
		t.Fatalf("Advance oder Delete trotz Abbruch: %+v %v", got, err)
	}
}
