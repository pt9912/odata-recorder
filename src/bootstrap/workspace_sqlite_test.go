package bootstrap_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// wsDatei ist der Workspace-Dienst über der echten SQLite-Datei, dem echten Blob-Speicher und der echten
// Wartung (SPEC-129 *Ungeprüft*, Verdrahtung): Workspaces, Transaktion und Snapshots über dieselbe Datenbank, Blobs
// über den Wächter derselben Ablage.
type wsDatei struct {
	t       *testing.T
	db      *sqlite.DB
	store   *blobs.Store
	wartung *maintenance.Wartung
	dienst  *workspace.Verwaltung
	uhr     *supporttest.FestUhr
}

func wsT0() time.Time { return time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC) }

func wsNeueDatei(t *testing.T) *wsDatei {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "r.db"), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	store, err := blobs.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	uhr := supporttest.NewFestUhr(wsT0())
	waechter, err := maintenance.NewBlobWaechter(store, uhr, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	wartung, err := maintenance.New(maintenance.Konfig{Sessions: db.Sessions(), Dienste: db.Metadata(), Speicher: db.Maintenance(),
		Blobs: waechter, Uhr: uhr, Log: log})
	if err != nil {
		t.Fatal(err)
	}
	dienst, err := workspace.New(workspace.Konfig{Workspaces: db.Workspaces(), Transaktion: db, Snapshots: db.Snapshots(), Blobs: waechter,
		Dienste: db.Metadata(), Wartung: wartung, Uhr: uhr, IDs: supporttest.NewFestIDs(), Log: log})
	if err != nil {
		t.Fatal(err)
	}
	d := &wsDatei{t: t, db: db, store: store, wartung: wartung, dienst: dienst, uhr: uhr}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(db.Metadata().Create(ctx, model.Service{ID: "svc", ServiceRoot: "/", ODataVersion: "4.0"}))
	must(db.Sessions().Create(ctx, model.Session{ID: "ses-1", ServiceID: "svc", CreatedAt: wsT0(), Status: model.SessionOpen}))
	must(db.Sessions().SetStatus(ctx, "ses-1", model.SessionComplete))
	return d
}

// wsEntity ist ein Eintrag der Baseline; payload ist ein JSON-Text.
type wsEntity struct{ set, key, payload string }

// manifest ist das Manifest in der Form von SPEC-127 Nr. 14 a; der Test schreibt es selbst (der Importer ist nicht
// im Paket).
func wsManifest(id string, es []wsEntity) []byte {
	l := make([]any, 0, len(es))
	for _, e := range es {
		l = append(l, map[string]any{"completeness": "complete", "entity_set": e.set, "payload": json.RawMessage(e.payload),
			"provenance": []any{}, "selected_properties": []string{"Id"}, "typed_key": e.key})
	}
	b, err := json.Marshal(map[string]any{"baseline_snapshot_id": id, "coverage": []any{}, "created_at": "2026-10-20T10:00:00.000000000Z",
		"entities": l, "entity_count": len(es), "import_id": "imp-" + id, "manifest_version": 1, "metadata_hash": "sha256:" + hex.EncodeToString(make([]byte, 32)),
		"provenance": []any{}, "service_id": "svc", "snapshot_kind": "baseline", "source_session_ids": []string{"ses-1"}})
	if err != nil {
		panic(err)
	}
	return b
}

// blob legt b im Blob-Speicher ab, am Wächter vorbei, und liefert die Referenz.
func (d *wsDatei) blob(b []byte) string {
	d.t.Helper()
	w, err := d.store.Begin(context.Background(), int64(len(b))+1)
	if err != nil {
		d.t.Fatal(err)
	}
	if _, err := w.Write(b); err != nil {
		d.t.Fatal(err)
	}
	ref, _, err := w.Commit()
	if err != nil {
		d.t.Fatal(err)
	}
	return ref
}

// baseline veröffentlicht die Baseline id mit den Einträgen es.
func (d *wsDatei) baseline(id string, es []wsEntity) {
	d.t.Helper()
	ref := d.blob(wsManifest(id, es))
	_, neu, err := d.db.Imports().Publish(context.Background(), model.Publication{
		Import:           model.Import{ID: "imp-" + id, RequestHash: "req-" + id, BaselineSnapshotID: id, ServiceID: "svc", CreatedAt: wsT0(), ReportJSON: "{}"},
		Snapshot:         model.BaselineSnapshot{ID: id, ServiceID: "svc", CreatedAt: wsT0(), StateBlob: ref, StateHash: ref, MetadataHash: ref},
		SourceSessionIDs: []string{"ses-1"},
	})
	if err != nil || !neu {
		d.t.Fatalf("Publish: %v, neu=%v", err, neu)
	}
}

func wsStandard() []wsEntity {
	return []wsEntity{{"Products", "Id=1", `{"Id":"1","Name":"Milk"}`}, {"Products", "Id=2", `{"Id":"2","Name":"Bread"}`}, {"Categories", "Id=1", `{"Id":"1"}`}}
}

func wsAndere() []wsEntity {
	return []wsEntity{{"Products", "Id=9", `{"Id":"9","Name":"Andere"}`}, {"Orders", "Id=1", `{"Id":"1"}`}}
}

// state liefert den Entity-State des Workspace in der Folge der Datei (nach Menge und Schlüssel), ohne Version
// und Tombstones herauszurechnen.
func (d *wsDatei) state(id string) []model.EntityState {
	d.t.Helper()
	es, err := d.db.Entities().List(context.Background(), id)
	if err != nil {
		d.t.Fatal(err)
	}
	return es
}

func wsErwartet(id string, es []wsEntity) []model.EntityState {
	var r []model.EntityState
	for _, e := range es {
		r = append(r, model.EntityState{WorkspaceID: id, EntitySet: e.set, EntityKey: e.key, Payload: e.payload,
			KnownProperties: []string{"Id"}, Completeness: model.EntityComplete})
	}
	// Die Datei liefert nach entity_set und entity_key.
	for i := range r {
		for j := i + 1; j < len(r); j++ {
			if r[j].EntitySet < r[i].EntitySet || (r[j].EntitySet == r[i].EntitySet && r[j].EntityKey < r[i].EntityKey) {
				r[i], r[j] = r[j], r[i]
			}
		}
	}
	return r
}

func wsCode(t *testing.T, err error, c model.Code) {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) || f.Code != c {
		t.Errorf("Fehler %v, erwartet %v", err, c)
	}
}

// TestWorkspaceZweiAusDerselbenBaseline (Z58): zwei Workspaces aus derselben Baseline auf der Datei; eine Änderung
// (Entities.Put und Advance in einer Transaktion) im einen lässt den anderen unverändert.
func TestWorkspaceZweiAusDerselbenBaseline(t *testing.T) {
	ctx := context.Background()
	d := wsNeueDatei(t)
	d.baseline("bsl-1", wsStandard())
	a, err := d.dienst.Create(ctx, "bsl-1", "ws-a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := d.dienst.Create(ctx, "bsl-1", "ws-b")
	if err != nil {
		t.Fatal(err)
	}
	if a.ServiceID != "svc" || a.Revision != 0 || b.ID != "ws-b" {
		t.Errorf("Workspaces %+v %+v", a, b)
	}
	for _, id := range []string{"ws-a", "ws-b"} {
		if got := d.state(id); !reflect.DeepEqual(got, wsErwartet(id, wsStandard())) {
			t.Fatalf("%s: State %+v", id, got)
		}
	}
	err = d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		if err := s.Entities.Put(ctx, model.EntityState{WorkspaceID: "ws-a", EntitySet: "Products", EntityKey: "Id=1",
			Payload: `{"Id":"1","Name":"Geaendert"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 1}); err != nil {
			return err
		}
		_, err := s.Workspaces.Advance(ctx, "ws-a", 0)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := d.state("ws-b"); !reflect.DeepEqual(got, wsErwartet("ws-b", wsStandard())) {
		t.Errorf("ws-b änderte sich: %+v", got)
	}
	if got, _ := d.dienst.Get(ctx, "ws-b"); got.Revision != 0 {
		t.Errorf("Revision von ws-b: %d", got.Revision)
	}
	if got := d.state("ws-a"); reflect.DeepEqual(got, wsErwartet("ws-a", wsStandard())) {
		t.Error("ws-a änderte sich nicht")
	}
	liste, err := d.dienst.List(ctx)
	if err != nil || len(liste) != 2 {
		t.Errorf("List: %+v, %v", liste, err)
	}
}

// veraendere legt in ws-a Mutation, Tombstone, zusätzliche Entity und einen fortgeschrittenen Cursor an.
func (d *wsDatei) veraendere(id string) {
	d.t.Helper()
	ctx := context.Background()
	err := d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		for _, e := range []model.EntityState{
			{WorkspaceID: id, EntitySet: "Products", EntityKey: "Id=1", Payload: `{"Id":"1","Name":"Geaendert"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 2},
			{WorkspaceID: id, EntitySet: "Products", EntityKey: "Id=2", Payload: `{"Id":"2"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 1, Deleted: true},
			{WorkspaceID: id, EntitySet: "Products", EntityKey: "Id=7", Payload: `{"Id":"7"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 1},
		} {
			if err := s.Entities.Put(ctx, e); err != nil {
				return err
			}
		}
		return s.Cursors.Put(ctx, model.ReplayCursor{WorkspaceID: id, ReplaySessionID: "rep-1", RecordingSessionID: "ses-1", ServiceID: "svc",
			Strategy: model.StrategySequential, Position: 2, Consumed: []int64{1, 2}})
	})
	if err != nil {
		d.t.Fatal(err)
	}
}

func (d *wsDatei) cursors(id string) []model.ReplayCursor {
	d.t.Helper()
	cs, err := d.db.Cursors().List(context.Background(), id)
	if err != nil {
		d.t.Fatal(err)
	}
	return cs
}

// TestWorkspaceResetAufDerDatei (Z59): Reset ohne Änderung lässt State und Cursors gleich der Baseline und erhöht die
// Revision; mit Mutation, Tombstone und fortgeschrittenem Cursor stellt er die Baseline und Position 0 her.
func TestWorkspaceResetAufDerDatei(t *testing.T) {
	ctx := context.Background()
	d := wsNeueDatei(t)
	d.baseline("bsl-1", wsStandard())
	if _, err := d.dienst.Create(ctx, "bsl-1", "ws-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.dienst.Create(ctx, "bsl-1", "ws-b"); err != nil {
		t.Fatal(err)
	}
	vorher := d.state("ws-a")
	ws, err := d.dienst.Reset(ctx, "ws-a")
	if err != nil || ws.Revision != 1 || !reflect.DeepEqual(d.state("ws-a"), vorher) || len(d.cursors("ws-a")) != 0 {
		t.Errorf("Reset ohne Änderung: %+v, %v", ws, err)
	}
	d.veraendere("ws-a")
	if reflect.DeepEqual(d.state("ws-a"), wsErwartet("ws-a", wsStandard())) {
		t.Fatal("die Änderung fehlt")
	}
	ws, err = d.dienst.Reset(ctx, "ws-a")
	if err != nil || ws.Revision != 2 {
		t.Fatalf("Reset: %+v, %v", ws, err)
	}
	if got := d.state("ws-a"); !reflect.DeepEqual(got, wsErwartet("ws-a", wsStandard())) {
		t.Errorf("State nach Reset: %+v", got)
	}
	cs := d.cursors("ws-a")
	if len(cs) != 1 || cs[0].Position != 0 || len(cs[0].Consumed) != 0 || cs[0].Strategy != model.StrategySequential || cs[0].ReplaySessionID != "rep-1" {
		t.Errorf("Cursors nach Reset: %+v", cs)
	}
	if got, _ := d.dienst.Get(ctx, "ws-a"); got.Revision != 2 || got.BaselineSnapshotID != "bsl-1" {
		t.Errorf("Get: %+v", got)
	}
}

// TestWorkspaceResetUnbekannt (Z60): Reset einer unbekannten id ist ODR-E7001, ein anderer Workspace bleibt
// unverändert.
func TestWorkspaceResetUnbekannt(t *testing.T) {
	ctx := context.Background()
	d := wsNeueDatei(t)
	d.baseline("bsl-1", wsStandard())
	if _, err := d.dienst.Create(ctx, "bsl-1", "ws-a"); err != nil {
		t.Fatal(err)
	}
	d.veraendere("ws-a")
	vorher, cvorher := d.state("ws-a"), d.cursors("ws-a")
	_, err := d.dienst.Reset(ctx, "unbekannt")
	wsCode(t, err, model.CodeE7001)
	if !reflect.DeepEqual(d.state("ws-a"), vorher) || !reflect.DeepEqual(d.cursors("ws-a"), cvorher) {
		t.Error("ws-a änderte sich")
	}
	if got, _ := d.dienst.Get(ctx, "ws-a"); got.Revision != 0 {
		t.Errorf("Revision %d", got.Revision)
	}
}

// TestWorkspaceGleichzeitigesCreate (Z61): zwei gleichzeitige Create mit derselben id auf der Datei (SQLite
// serialisiert): genau eines gelingt, das andere ist ODR-E7007, der State ist der der Baseline des Gewinners.
func TestWorkspaceGleichzeitigesCreate(t *testing.T) {
	ctx := context.Background()
	d := wsNeueDatei(t)
	d.baseline("bsl-a", wsStandard())
	d.baseline("bsl-b", wsAndere())
	gewinner := map[string]int{}
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("ws-%d", i)
		fehler := make([]error, 2)
		var wg sync.WaitGroup
		start := make(chan struct{})
		for k, b := range []string{"bsl-a", "bsl-b"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, fehler[k] = d.dienst.Create(ctx, b, id)
			}()
		}
		close(start)
		wg.Wait()
		gelungen := 0
		for _, err := range fehler {
			if err == nil {
				gelungen++
			} else {
				wsCode(t, err, model.CodeE7007)
			}
		}
		ws, err := d.dienst.Get(ctx, id)
		if err != nil || gelungen != 1 {
			t.Fatalf("Durchlauf %d: gelungen %d, Get: %v", i, gelungen, err)
		}
		soll := wsStandard()
		if ws.BaselineSnapshotID == "bsl-b" {
			soll = wsAndere()
		}
		gewinner[ws.BaselineSnapshotID]++
		if got := d.state(id); !reflect.DeepEqual(got, wsErwartet(id, soll)) {
			t.Errorf("Durchlauf %d: State %+v statt der Baseline %s", i, got, ws.BaselineSnapshotID)
		}
	}
	t.Logf("Gewinner: %v", gewinner)
}
