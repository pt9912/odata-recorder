package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
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
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// snDatei ist der Snapshot-Dienst über derselben Datei, derselben Ablage und demselben Wächter wie der Workspace-Dienst
// (SPEC-129 Nr. 21: der Composition Root übergibt den Wächter der Wartung als Blobs). Die Workspaces ws-a und ws-b
// stammen aus der Baseline bsl-1.
type snDatei struct {
	*wsDatei
	snap *snapshot.Dienst
	ids  *supporttest.FestIDs
}

func snDienst(t *testing.T, d *wsDatei, blobs outbound.BlobRepository) (*snapshot.Dienst, *supporttest.FestIDs) {
	t.Helper()
	ids := supporttest.NewFestIDs()
	s, err := snapshot.New(snapshot.Konfig{Workspaces: d.db.Workspaces(), Entities: d.db.Entities(), Snapshots: d.db.Snapshots(), Blobs: blobs,
		Uhr: d.uhr, IDs: ids})
	if err != nil {
		t.Fatal(err)
	}
	return s, ids
}

func snNeueDatei(t *testing.T) *snDatei {
	t.Helper()
	d := wsNeueDatei(t)
	d.baseline("bsl-1", wsStandard())
	for _, id := range []string{"ws-a", "ws-b"} {
		if _, err := d.dienst.Create(context.Background(), "bsl-1", id); err != nil {
			t.Fatal(err)
		}
	}
	s, ids := snDienst(t, d, d.waechter)
	return &snDatei{wsDatei: d, snap: s, ids: ids}
}

// legeAn erzeugt den Workspace id aus bsl-1 und verändert ihn (Mutation, Tombstone, zusätzliche Entity, Cursor an Position 5).
func (d *snDatei) legeAn(id string) {
	d.t.Helper()
	if _, err := d.dienst.Create(context.Background(), "bsl-1", id); err != nil {
		d.t.Fatal(err)
	}
	d.veraendere(id)
	d.setzeCursor(id, 5)
}

// setzeCursor schreibt den Cursor rep-1/ses-1 des Workspace id auf Position pos mit consumed 1..pos.
func (d *snDatei) setzeCursor(id string, pos int64) {
	d.t.Helper()
	ctx := context.Background()
	var consumed []int64
	for i := int64(1); i <= pos; i++ {
		consumed = append(consumed, i)
	}
	err := d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		return s.Cursors.Put(ctx, model.ReplayCursor{WorkspaceID: id, ReplaySessionID: "rep-1", RecordingSessionID: "ses-1", ServiceID: "svc",
			Strategy: model.StrategySequential, Position: pos, Consumed: consumed})
	})
	if err != nil {
		d.t.Fatal(err)
	}
}

// mutiere ist der Schreiber der Mutation (SPEC-129 Nr. 11 b): in einer Transaktion eine Entity ändern (Version plus 1) und
// die Revision fortschreiben, die er in derselben Transaktion liest.
func (d *snDatei) mutiere(id, key string) error {
	ctx := context.Background()
	return d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		w, err := s.Workspaces.Get(ctx, id)
		if err != nil {
			return err
		}
		e, err := s.Entities.Get(ctx, id, "Products", key)
		if err != nil {
			return err
		}
		e.Version++
		e.Payload = fmt.Sprintf(`{"Id":"1","Name":"M%d"}`, e.Version)
		if err := s.Entities.Put(ctx, e); err != nil {
			return err
		}
		_, err = s.Workspaces.Advance(ctx, id, w.Revision)
		return err
	})
}

// fortschreiben ist der Schreiber des Cursors (SPEC-129 Nr. 11 a): Lesen und Schreiben in derselben Transaktion.
func (d *snDatei) fortschreiben(id string) error {
	ctx := context.Background()
	return d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		c, err := s.Cursors.Get(ctx, id, "rep-1", "ses-1")
		if err != nil {
			return err
		}
		c.Position++
		c.Consumed = append(c.Consumed, c.Position)
		return s.Cursors.Put(ctx, c)
	})
}

func (d *snDatei) revision(id string) int64 {
	d.t.Helper()
	w, err := d.db.Workspaces().Get(context.Background(), id)
	if err != nil {
		d.t.Fatal(err)
	}
	return w.Revision
}

// norm liefert den State mit aufsteigenden Properties, damit der Vergleich nicht von der Ablage der Maske abhängt.
func norm(es []model.EntityState) []model.EntityState {
	r := make([]model.EntityState, len(es))
	for i, e := range es {
		e.KnownProperties = slices.Clone(e.KnownProperties)
		sort.Strings(e.KnownProperties)
		r[i] = e
	}
	return r
}

func (d *snDatei) erzeuge(id, name string) model.WorkspaceSnapshot {
	d.t.Helper()
	s, err := d.snap.Create(context.Background(), id, name)
	if err != nil {
		d.t.Fatal(err)
	}
	return s
}

func (d *snDatei) restore(id, snap string) (model.Workspace, error) {
	return d.dienst.Restore(context.Background(), id, snap)
}

// TestSnapshotListeSqlite (T26): List liefert die Folge der Datei: bei gleicher Zeit nach Kennung, auch wenn die
// Snapshots in anderer Folge angelegt wurden; ein Workspace ohne Snapshot eine leere Liste, ein unbekannter ODR-E7001.
func TestSnapshotListeSqlite(t *testing.T) {
	d := snNeueDatei(t)
	ref := d.blob([]byte("Inhalt"))
	for _, id := range []string{"snp-3", "snp-1", "snp-5", "snp-2", "snp-4"} {
		if err := d.db.Snapshots().CreateWorkspace(context.Background(), wsWorkspaceSnapshot(id, "ws-a", ref)); err != nil {
			t.Fatal(err)
		}
	}
	l, err := d.snap.List(context.Background(), "ws-a")
	if err != nil || len(l) != 5 {
		t.Fatalf("%v (%d)", err, len(l))
	}
	for i, s := range l {
		if s.ID != fmt.Sprintf("snp-%d", i+1) || !s.CreatedAt.Equal(wsT0()) {
			t.Errorf("Position %d: %s", i, s.ID)
		}
	}
	if l, err := d.snap.List(context.Background(), "ws-b"); err != nil || l == nil || len(l) != 0 {
		t.Errorf("leer: %v %v", l, err)
	}
	_, err = d.snap.List(context.Background(), "ws-gibtsnicht")
	wsCode(t, err, model.CodeE7001)
}

// TestSnapshotRundlauf (T46, T55): was Create für jeden gültigen State schreibt, liest Restore und stellt denselben State
// her (M2: 48 Zustände, M3: 14 Payload-Arten, echter SQLite-Adapter, echte Dienste); der Happy Path nach LH-FA-10:
// verändern, Snapshot, weiter verändern, Restore stellt den State zur Erzeugung her, die Cursors bleiben auf dem Stand
// nach dem Fortschreiben.
func TestSnapshotRundlauf(t *testing.T) {
	ctx := context.Background()
	d := snNeueDatei(t)
	d.veraendere("ws-a")
	s1 := norm(d.state("ws-a"))
	snap := d.erzeuge("ws-a", "s1")
	for _, k := range []string{"Id=1", "Id=1"} {
		if err := d.mutiere("ws-a", k); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		if err := s.Entities.Put(ctx, model.EntityState{WorkspaceID: "ws-a", EntitySet: "Neu", EntityKey: "K", Payload: `{}`, Completeness: model.EntityComplete, Version: 1}); err != nil {
			return err
		}
		return s.Entities.Put(ctx, model.EntityState{WorkspaceID: "ws-a", EntitySet: "Products", EntityKey: "Id=2", Payload: `{"Id":"2","wieder":true}`,
			KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 5})
	}); err != nil {
		t.Fatal(err)
	}
	if err := d.fortschreiben("ws-a"); err != nil {
		t.Fatal(err)
	}
	cursorsVorher := d.cursors("ws-a")
	revVorher := d.revision("ws-a")
	if reflect.DeepEqual(norm(d.state("ws-a")), s1) || len(cursorsVorher) != 1 || cursorsVorher[0].Position != 3 {
		t.Fatalf("Fixture: Cursors %+v", cursorsVorher)
	}
	got, err := d.restore("ws-a", snap.ID)
	if err != nil || got.Revision != revVorher+1 {
		t.Fatalf("Restore: %+v, %v", got, err)
	}
	if !reflect.DeepEqual(norm(d.state("ws-a")), s1) {
		t.Errorf("State %+v, erwartet %+v", d.state("ws-a"), s1)
	}
	if !reflect.DeepEqual(d.cursors("ws-a"), cursorsVorher) {
		t.Errorf("Cursors %+v, vorher %+v", d.cursors("ws-a"), cursorsVorher)
	}

	// M2 und M3 im echten Speicher.
	var viele []model.EntityState
	i := 0
	for _, c := range []model.Completeness{model.EntityComplete, model.EntityPartial} {
		for _, del := range []bool{false, true} {
			for _, v := range []int64{0, 1, 7, 1 << 40} {
				for _, m := range [][]string{{}, {"Id"}, {"Z", "M", "A"}} {
					i++
					viele = append(viele, model.EntityState{WorkspaceID: "ws-b", EntitySet: "Zustand", EntityKey: fmt.Sprintf("K%02d", i), Payload: `{"Id":"1"}`,
						KnownProperties: m, Completeness: c, Version: v, Deleted: del})
				}
			}
		}
	}
	for _, p := range []string{`{"a":1}`, `{ "a" :  1 }`, `[1, 2]`, `"x"`, `1.50`, `null`, `{"s":"a<>&` + string([]byte{0xC3, 0xA9}) + `"}`} {
		for _, del := range []bool{false, true} {
			i++
			viele = append(viele, model.EntityState{WorkspaceID: "ws-b", EntitySet: "Payload", EntityKey: fmt.Sprintf("K%02d", i), Payload: p,
				KnownProperties: []string{"Id"}, Completeness: model.EntityComplete, Version: 1, Deleted: del})
		}
	}
	if len(viele) != 48+14 {
		t.Fatalf("M2 und M3: %d Zustände", len(viele))
	}
	if err := d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error { return s.Entities.ReplaceAll(ctx, "ws-b", viele) }); err != nil {
		t.Fatal(err)
	}
	vorher := norm(d.state("ws-b"))
	sb := d.erzeuge("ws-b", "viele")
	if err := d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error { return s.Entities.ReplaceAll(ctx, "ws-b", nil) }); err != nil {
		t.Fatal(err)
	}
	if _, err := d.restore("ws-b", sb.ID); err != nil {
		t.Fatal(err)
	}
	if got := norm(d.state("ws-b")); !reflect.DeepEqual(got, vorher) {
		t.Errorf("Rundlauf: %d Einträge gegen %d", len(got), len(vorher))
	}
}

// TestResetNachRestore (T54): Restore ruft ResetAll nie, Reset weiter: nach Restore und Reset ist der State die Baseline,
// jeder Cursor 0 und die Revision +2; nach Reset und Restore ist der State der des Snapshots (zwei Folgen).
func TestResetNachRestore(t *testing.T) {
	d := snNeueDatei(t)
	d.veraendere("ws-a")
	d.setzeCursor("ws-a", 5)
	snap := d.erzeuge("ws-a", "s1")
	zustand := norm(d.state("ws-a"))
	rev := d.revision("ws-a")
	if _, err := d.restore("ws-a", snap.ID); err != nil {
		t.Fatal(err)
	}
	if c := d.cursors("ws-a"); len(c) != 1 || c[0].Position != 5 {
		t.Errorf("Restore berührt Cursors: %+v", c)
	}
	if _, err := d.dienst.Reset(context.Background(), "ws-a"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d.state("ws-a"), wsErwartet("ws-a", wsStandard())) || d.cursors("ws-a")[0].Position != 0 || d.revision("ws-a") != rev+2 {
		t.Errorf("Restore, Reset: %+v, Cursors %+v, Revision %d", d.state("ws-a"), d.cursors("ws-a"), d.revision("ws-a"))
	}
	d.setzeCursor("ws-a", 4)
	if _, err := d.restore("ws-a", snap.ID); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(norm(d.state("ws-a")), zustand) || d.cursors("ws-a")[0].Position != 4 || d.revision("ws-a") != rev+3 {
		t.Errorf("Reset, Restore: %+v", d.state("ws-a"))
	}
}

// TestSnapshotLeerMitCursor (T56): ein leerer Workspace und ein Cursor (Anzahl 0, 1, 3 über zwei Replay- und zwei
// Recording-Sessions × Stand: Position 0, fortgeschritten, ganz verbraucht = 9): Snapshot und Restore lassen den State leer
// und die Cursors unverändert.
func TestSnapshotLeerMitCursor(t *testing.T) {
	ctx := context.Background()
	d := snNeueDatei(t)
	if err := d.db.Sessions().Create(ctx, model.Session{ID: "ses-2", ServiceID: "svc", CreatedAt: wsT0(), Status: model.SessionOpen}); err != nil {
		t.Fatal(err)
	}
	if err := d.db.Sessions().SetStatus(ctx, "ses-2", model.SessionComplete); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, anzahl := range []int{0, 1, 3} {
		for _, pos := range []int64{0, 5, 9} {
			n++
			id := fmt.Sprintf("ws-c%d", n)
			if _, err := d.dienst.Create(ctx, "bsl-1", id); err != nil {
				t.Fatal(err)
			}
			err := d.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
				if err := s.Entities.ReplaceAll(ctx, id, nil); err != nil {
					return err
				}
				paare := [][2]string{{"rep-1", "ses-1"}, {"rep-1", "ses-2"}, {"rep-2", "ses-1"}, {"rep-2", "ses-2"}}[:anzahl]
				for _, p := range paare {
					var consumed []int64
					for i := int64(1); i <= pos; i++ {
						consumed = append(consumed, i)
					}
					c := model.ReplayCursor{WorkspaceID: id, ReplaySessionID: p[0], RecordingSessionID: p[1], ServiceID: "svc", Strategy: model.StrategyStrict, Position: pos, Consumed: consumed}
					if err := s.Cursors.Put(ctx, c); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			vorher := d.cursors(id)
			snap := d.erzeuge(id, "leer")
			if _, err := d.restore(id, snap.ID); err != nil {
				t.Fatal(err)
			}
			if got := d.state(id); len(got) != 0 || !reflect.DeepEqual(d.cursors(id), vorher) || len(vorher) != anzahl {
				t.Errorf("Anzahl %d Stand %d: State %+v, Cursors %+v gegen %+v", anzahl, pos, got, d.cursors(id), vorher)
			}
		}
	}
	if n != 9 {
		t.Errorf("M15: %d Elemente", n)
	}
}

// TestSnapshotUngueltigeID (T57): eine ungültige Kennung (unbekannt, leer, Baseline, Snapshot eines anderen Workspace)
// lässt State, Revision und Cursors im echten Speicher unverändert.
func TestSnapshotUngueltigeID(t *testing.T) {
	d := snNeueDatei(t)
	d.veraendere("ws-a")
	fremd := d.erzeuge("ws-b", "fremd")
	vorherState, vorherCursors, rev := d.state("ws-a"), d.cursors("ws-a"), d.revision("ws-a")
	for id, code := range map[string]model.Code{"snp-gibtsnicht": model.CodeE7002, "": model.CodeE7002, "bsl-1": model.CodeE7008, fremd.ID: model.CodeE7002} {
		_, err := d.restore("ws-a", id)
		wsCode(t, err, code)
		if !reflect.DeepEqual(d.state("ws-a"), vorherState) || !reflect.DeepEqual(d.cursors("ws-a"), vorherCursors) || d.revision("ws-a") != rev {
			t.Errorf("Kennung %q: Zustand geändert", id)
		}
	}
}

// snOeffne baut die Dienste über dem Verzeichnis dir: Datei r.db und Ablage blobs.
func snOeffne(t *testing.T, dir string) (*snDatei, func()) {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(dir, "r.db"), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "blobs"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := blobs.New(filepath.Join(dir, "blobs"))
	if err != nil {
		t.Fatal(err)
	}
	uhr := supporttest.NewFestUhr(wsT0())
	waechter, err := maintenance.NewBlobWaechter(store, uhr, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	audit := supporttest.NewFakeAudit()
	wartung, err := maintenance.New(maintenance.Konfig{Sessions: db.Sessions(), Dienste: db.Metadata(), Speicher: db.Maintenance(), Blobs: waechter, Uhr: uhr, Log: log, Audit: audit})
	if err != nil {
		t.Fatal(err)
	}
	dienst, err := workspace.New(workspace.Konfig{Workspaces: db.Workspaces(), Transaktion: db, Snapshots: db.Snapshots(), Blobs: waechter,
		Dienste: db.Metadata(), Wartung: wartung, Uhr: uhr, IDs: supporttest.NewFestIDs(), Log: log, Audit: audit})
	if err != nil {
		t.Fatal(err)
	}
	d := &wsDatei{t: t, db: db, store: store, waechter: waechter, wartung: wartung, dienst: dienst, uhr: uhr}
	s, ids := snDienst(t, d, waechter)
	return &snDatei{wsDatei: d, snap: s, ids: ids}, func() { _ = db.Close() }
}

// TestSnapshotNeustart (T58): nach regulärem Schließen und Öffnen der Datei sind die Snapshots vorhanden und
// wiederherstellbar; Datenbank und Blob-Speicher liegen im selben Verzeichnis.
func TestSnapshotNeustart(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	d, schliessen := snOeffne(t, dir)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(d.db.Metadata().Create(ctx, model.Service{ID: "svc", ServiceRoot: "/", ODataVersion: "4.0"}))
	must(d.db.Sessions().Create(ctx, model.Session{ID: "ses-1", ServiceID: "svc", CreatedAt: wsT0(), Status: model.SessionOpen}))
	must(d.db.Sessions().SetStatus(ctx, "ses-1", model.SessionComplete))
	d.baseline("bsl-1", wsStandard())
	_, err := d.dienst.Create(ctx, "bsl-1", "ws-a")
	must(err)
	d.veraendere("ws-a")
	zustand := norm(d.state("ws-a"))
	snap := d.erzeuge("ws-a", "vor dem Neustart")
	schliessen()
	d2, schliessen2 := snOeffne(t, dir)
	defer schliessen2()
	l, err := d2.snap.List(ctx, "ws-a")
	if err != nil || len(l) != 1 || l[0] != snap {
		t.Fatalf("List: %+v, %v", l, err)
	}
	if err := d2.db.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error { return s.Entities.ReplaceAll(ctx, "ws-a", nil) }); err != nil {
		t.Fatal(err)
	}
	if _, err := d2.restore("ws-a", snap.ID); err != nil || !reflect.DeepEqual(norm(d2.state("ws-a")), zustand) {
		t.Errorf("Restore nach Neustart: %v", err)
	}
}

// hakenSchreiber ruft nach Commit den Haken; hakenBlobs reicht den Rest an den Wächter durch.
type hakenBlobs struct {
	outbound.BlobRepository
	nachCommit func()
}

type hakenSchreiber struct {
	outbound.CaptureWriter
	nachCommit func()
}

func (h hakenBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	w, err := h.BlobRepository.Begin(ctx, grenze)
	return hakenSchreiber{w, h.nachCommit}, err
}

func (h hakenSchreiber) Commit() (string, int64, error) {
	ref, n, err := h.CaptureWriter.Commit()
	if h.nachCommit != nil {
		h.nachCommit()
	}
	return ref, n, err
}

// TestSnapshotBlobGeschuetzt (T59): der Manifest-Blob eines lebenden Snapshots ist referenziert und überlebt die
// Bereinigung, auch wenn sie zwischen Commit und Referenz läuft (Wächter, Karenz-Uhr als FestUhr); nach dem Löschen des
// Workspace und dem Ende der Karenz wird er entfernt.
func TestSnapshotBlobGeschuetzt(t *testing.T) {
	ctx := context.Background()
	d := snNeueDatei(t)
	d.veraendere("ws-a")
	hk := hakenBlobs{BlobRepository: d.waechter, nachCommit: func() { _, _ = d.wartung.CleanupBlobs(ctx) }}
	s, _ := snDienst(t, d.wsDatei, hk)
	snap, err := s.Create(ctx, "ws-a", "geschuetzt")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(d.blobListe(), snap.StateBlob) {
		t.Fatal("der Blob fehlt nach Create")
	}
	if _, err := d.wartung.CleanupBlobs(ctx); err != nil || !slices.Contains(d.blobListe(), snap.StateBlob) {
		t.Errorf("Bereinigung entfernte den Blob eines lebenden Snapshots: %v", err)
	}
	refs, err := d.db.Maintenance().ReferencedBlobs(ctx)
	if err != nil || !slices.Contains(refs, snap.StateBlob) {
		t.Errorf("ReferencedBlobs %v: %v", refs, err)
	}
	if _, err := d.restore("ws-a", snap.ID); err != nil {
		t.Errorf("Restore: %v", err)
	}
	d.uhr.Advance(time.Second)
	if _, err := d.dienst.Delete(ctx, "ws-a"); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(d.blobListe(), snap.StateBlob) {
		t.Error("der Blob blieb nach dem Löschen des Workspace")
	}
}

// gestartet läuft n Funktionen gleichzeitig; es liefert die Fehler in der Reihenfolge der Nummern.
func gestartet(n int, f func(i int) error) []error {
	r := make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			r[i] = f(i)
		}()
	}
	close(start)
	wg.Wait()
	return r
}

// TestRestoreGegenMutation (T62): n gleichzeitige Mutationen und ein Restore: die Revision am Ende ist Start + n + 1, und
// die Version der mutierten Entity ist ihre Version im Snapshot plus die Zahl der Mutationen nach dem Restore (Revision am
// Ende minus Revision des Ergebnisses); 16 Goroutinen × 4 Runden, echter SQLite-Adapter, Wirkungs-Fänger.
func TestRestoreGegenMutation(t *testing.T) {
	d := snNeueDatei(t)
	for runde := 0; runde < 4; runde++ {
		id := fmt.Sprintf("ws-m%d", runde)
		d.legeAn(id)
		snap := d.erzeuge(id, "s")
		versionImSnapshot := int64(2) // veraendere setzt Products/Id=1 auf Version 2
		start := d.revision(id)
		var ergebnis model.Workspace
		fehler := gestartet(16, func(i int) error {
			if i == 7 {
				var err error
				ergebnis, err = d.restore(id, snap.ID)
				return err
			}
			return d.mutiere(id, "Id=1")
		})
		for i, err := range fehler {
			if err != nil {
				t.Fatalf("Runde %d Lauf %d: %v", runde, i, err)
			}
		}
		ende := d.revision(id)
		var e model.EntityState
		for _, x := range d.state(id) {
			if x.EntitySet == "Products" && x.EntityKey == "Id=1" {
				e = x
			}
		}
		if ende != start+16 || e.Version != versionImSnapshot+(ende-ergebnis.Revision) {
			t.Errorf("Runde %d: Revision %d (Start %d), Version %d, Ergebnis des Restore %d", runde, ende, start, e.Version, ergebnis.Revision)
		}
	}
}

// pruefeEntweder: der State des Workspace id ist genau a oder genau b (nie eine Mischung); es liefert, welcher.
func pruefeEntweder(t *testing.T, d *snDatei, id string, a, b []model.EntityState) string {
	t.Helper()
	switch got := norm(d.state(id)); {
	case reflect.DeepEqual(got, a):
		return "a"
	case reflect.DeepEqual(got, b):
		return "b"
	}
	t.Errorf("%s: State ist weder der erwartete noch der andere (Mischung): %+v", id, d.state(id))
	return ""
}

// TestRestoreGegenReset (T63): Restore und Reset desselben Workspace: die Revision wächst um 2, der State ist genau die
// Baseline (Reset zuletzt) oder genau der des Snapshots, jeder Cursor steht auf 0; 20 Paare × 4 Runden, Cursor vorher 5.
func TestRestoreGegenReset(t *testing.T) {
	d := snNeueDatei(t)
	for runde := 0; runde < 4; runde++ {
		var ids []string
		snaps := map[string]model.WorkspaceSnapshot{}
		for i := 0; i < 20; i++ {
			id := fmt.Sprintf("ws-r%d-%d", runde, i)
			d.legeAn(id)
			snaps[id] = d.erzeuge(id, "s")
			ids = append(ids, id)
		}
		ergebnis := make([]model.Workspace, 20)
		zs := map[string][]model.EntityState{}
		for _, id := range ids {
			zs[id] = norm(d.state(id))
		}
		fehler := gestartet(40, func(i int) error {
			id := ids[i/2]
			var err error
			if i%2 == 0 {
				ergebnis[i/2], err = d.restore(id, snaps[id].ID)
			} else {
				_, err = d.dienst.Reset(context.Background(), id)
			}
			return err
		})
		for i, err := range fehler {
			if err != nil {
				t.Fatalf("Runde %d Lauf %d: %v", runde, i, err)
			}
		}
		for i, id := range ids {
			basis := wsErwartet(id, wsStandard())
			if d.revision(id) != 2 {
				t.Errorf("%s: Revision %d", id, d.revision(id))
			}
			for _, c := range d.cursors(id) {
				if c.Position != 0 || len(c.Consumed) != 0 {
					t.Errorf("%s: Cursor %+v", id, c)
				}
			}
			welcher := pruefeEntweder(t, d, id, norm(basis), zs[id])
			restoreZuletzt := ergebnis[i].Revision == 2
			if (welcher == "a") == restoreZuletzt && welcher != "" {
				t.Errorf("%s: State passt nicht zur Folge (Restore zuletzt: %v, State Baseline: %v)", id, restoreZuletzt, welcher == "a")
			}
		}
	}
}

// TestRestoreGegenRestore (T64): zwei Restore mit verschiedenen Snapshots desselben Workspace: die Revision wächst um 2,
// die Ergebnisse tragen n+1 und n+2, der State ist genau der des später serialisierten; 16 Paare × 4 Runden.
func TestRestoreGegenRestore(t *testing.T) {
	d := snNeueDatei(t)
	for runde := 0; runde < 4; runde++ {
		type paar struct{ a, b model.WorkspaceSnapshot }
		var ids []string
		ps := map[string]paar{}
		zustand := map[string][2][]model.EntityState{}
		for i := 0; i < 16; i++ {
			id := fmt.Sprintf("ws-p%d-%d", runde, i)
			d.legeAn(id)
			a := d.erzeuge(id, "a")
			za := norm(d.state(id))
			if err := d.mutiere(id, "Id=1"); err != nil {
				t.Fatal(err)
			}
			b := d.erzeuge(id, "b")
			zb := norm(d.state(id))
			ps[id], zustand[id] = paar{a, b}, [2][]model.EntityState{za, zb}
			ids = append(ids, id)
		}
		start := map[string]int64{}
		for _, id := range ids {
			start[id] = d.revision(id)
		}
		ergebnis := make([]model.Workspace, 32)
		fehler := gestartet(32, func(i int) error {
			id := ids[i/2]
			snap := ps[id].a
			if i%2 == 1 {
				snap = ps[id].b
			}
			var err error
			ergebnis[i], err = d.restore(id, snap.ID)
			return err
		})
		for i, err := range fehler {
			if err != nil {
				t.Fatalf("Runde %d Lauf %d: %v", runde, i, err)
			}
		}
		for i, id := range ids {
			ra, rb := ergebnis[2*i].Revision, ergebnis[2*i+1].Revision
			menge := []int64{ra, rb}
			slices.Sort(menge)
			if d.revision(id) != start[id]+2 || menge[0] != start[id]+1 || menge[1] != start[id]+2 {
				t.Errorf("%s: Revision %d, Ergebnisse %v (Start %d)", id, d.revision(id), menge, start[id])
			}
			soll := zustand[id][0]
			if rb > ra {
				soll = zustand[id][1]
			}
			if !reflect.DeepEqual(norm(d.state(id)), soll) {
				t.Errorf("%s: State nicht der des später serialisierten Restore", id)
			}
		}
	}
}

// TestRestoreGegenLoeschen (T65): Restore und Delete desselben Workspace: das Restore gelingt oder ist ODR-E7001 (oder
// ODR-E7002, wenn das Löschen den Snapshot vor dem Lesen nimmt); der Workspace ist danach weg, und nie bleibt ein State
// ohne Workspace (foreign_key_check meldet nichts); 20 Paare.
func TestRestoreGegenLoeschen(t *testing.T) {
	ctx := context.Background()
	d := snNeueDatei(t)
	var ids []string
	snaps := map[string]model.WorkspaceSnapshot{}
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("ws-l%d", i)
		d.legeAn(id)
		snaps[id] = d.erzeuge(id, "s")
		ids = append(ids, id)
	}
	fehler := gestartet(40, func(i int) error {
		id := ids[i/2]
		if i%2 == 0 {
			_, err := d.dienst.Restore(ctx, id, snaps[id].ID)
			return err
		}
		_, err := d.dienst.Delete(ctx, id)
		return err
	})
	for i, err := range fehler {
		var f *inbound.Fehler
		switch {
		case i%2 == 1 && err != nil:
			t.Errorf("Delete %d: %v", i/2, err)
		case i%2 == 0 && err != nil && (!errors.As(err, &f) || (f.Code != model.CodeE7001 && f.Code != model.CodeE7002)):
			t.Errorf("Restore %d: %v", i/2, err)
		}
	}
	for _, id := range ids {
		if _, err := d.db.Workspaces().Get(ctx, id); !errors.Is(err, outbound.ErrNichtGefunden) {
			t.Errorf("%s besteht noch: %v", id, err)
		}
	}
	if befunde, err := d.db.Maintenance().Check(ctx); err != nil || len(befunde) != 0 {
		t.Errorf("Check: %v, %v", befunde, err)
	}
}

// TestRestoreGegenCursor (T66): 20 Fortschreibungen eines Cursors und ein Restore: die Position am Ende ist 5 + 20; das
// Restore beginnt, nachdem das erste Fortschreiben gelungen ist (Signal); 4 Runden.
func TestRestoreGegenCursor(t *testing.T) {
	d := snNeueDatei(t)
	for runde := 0; runde < 4; runde++ {
		id := fmt.Sprintf("ws-c%d", runde)
		d.legeAn(id)
		snap := d.erzeuge(id, "s")
		signal := make(chan struct{})
		var einmal sync.Once
		fehler := gestartet(21, func(i int) error {
			if i == 20 {
				<-signal
				_, err := d.restore(id, snap.ID)
				return err
			}
			err := d.fortschreiben(id)
			if err == nil {
				einmal.Do(func() { close(signal) })
			}
			return err
		})
		for i, err := range fehler {
			if err != nil {
				t.Fatalf("Runde %d Lauf %d: %v", runde, i, err)
			}
		}
		if c := d.cursors(id); len(c) != 1 || c[0].Position != 25 || len(c[0].Consumed) != 25 {
			t.Errorf("Runde %d: Cursor %+v, erwartet Position 25", runde, c)
		}
	}
}

// TestNameWettlauf (T68): 16 gleichzeitige Create mit demselben Namen: genau eines gelingt, die übrigen sind ODR-E7006,
// List zeigt einen Snapshot, höchstens ein Blob kommt hinzu (gleicher State, eine Referenz); 4 Runden.
func TestNameWettlauf(t *testing.T) {
	d := snNeueDatei(t)
	for runde := 0; runde < 4; runde++ {
		id := fmt.Sprintf("ws-n%d", runde)
		d.legeAn(id)
		vorher := len(d.blobListe())
		fehler := gestartet(16, func(i int) error {
			_, err := d.snap.Create(context.Background(), id, "gleich")
			return err
		})
		gelungen := 0
		for _, err := range fehler {
			if err == nil {
				gelungen++
			} else {
				wsCode(t, err, model.CodeE7006)
			}
		}
		l, _ := d.snap.List(context.Background(), id)
		if gelungen != 1 || len(l) != 1 || len(d.blobListe()) > vorher+1 {
			t.Errorf("Runde %d: gelungen %d, Snapshots %d, Blobs %d (vorher %d)", runde, gelungen, len(l), len(d.blobListe()), vorher)
		}
	}
}

// TestCreateGegenLoeschen (T69, echter Speicher): ein Löschen des Workspace zwischen Workspaces.Get und dem Rest und nach
// Commit ist ODR-E7001; es entsteht kein Snapshot, und der Blob nach Commit bleibt ohne Referenz (ReferencedBlobs kennt ihn
// nicht).
func TestCreateGegenLoeschen(t *testing.T) {
	ctx := context.Background()
	for _, punkt := range []string{"Get", "Commit"} {
		d := snNeueDatei(t)
		d.veraendere("ws-a")
		var s *snapshot.Dienst
		loesche := func() { _ = d.db.Workspaces().Delete(ctx, "ws-a") }
		vorher := d.blobListe()
		if punkt == "Get" {
			var err error
			s, err = snapshot.New(snapshot.Konfig{Workspaces: hakenWorkspaces{d.db.Workspaces(), loesche}, Entities: d.db.Entities(), Snapshots: d.db.Snapshots(),
				Blobs: d.waechter, Uhr: d.uhr, IDs: d.ids})
			if err != nil {
				t.Fatal(err)
			}
		} else {
			s, _ = snDienst(t, d.wsDatei, hakenBlobs{BlobRepository: d.waechter, nachCommit: loesche})
		}
		_, err := s.Create(ctx, "ws-a", "n")
		wsCode(t, err, model.CodeE7001)
		refs, _ := d.db.Maintenance().ReferencedBlobs(ctx)
		var neu []string
		for _, r := range d.blobListe() {
			if !slices.Contains(vorher, r) {
				neu = append(neu, r)
			}
		}
		if punkt == "Get" && len(neu) != 0 || punkt == "Commit" && (len(neu) != 1 || slices.Contains(refs, neu[0])) {
			t.Errorf("%s: neue Blobs %v, referenziert %v", punkt, neu, refs)
		}
		if _, err := d.db.Snapshots().ListWorkspace(ctx, "ws-a"); !errors.Is(err, outbound.ErrNichtGefunden) {
			t.Errorf("%s: Snapshots eines gelöschten Workspace: %v", punkt, err)
		}
	}
}

// hakenWorkspaces ruft nach Get den Haken.
type hakenWorkspaces struct {
	outbound.WorkspaceRepository
	nachGet func()
}

func (h hakenWorkspaces) Get(ctx context.Context, id string) (model.Workspace, error) {
	w, err := h.WorkspaceRepository.Get(ctx, id)
	h.nachGet()
	return w, err
}

// TestIsolationSnapshotSqlite (T70): 20 Workspaces mit je einer Folge aus Create, Mutation, Restore und List gleichzeitig,
// gleiche Namen in allen: jeder endet in seinem Soll, kein Snapshot ist im anderen sichtbar.
func TestIsolationSnapshotSqlite(t *testing.T) {
	d := snNeueDatei(t)
	var ids []string
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("ws-i%d", i)
		d.legeAn(id)
		ids = append(ids, id)
	}
	soll := map[string][]model.EntityState{}
	for _, id := range ids {
		soll[id] = norm(d.state(id))
	}
	fehler := gestartet(20, func(i int) error {
		id := ids[i]
		ctx := context.Background()
		snap, err := d.snap.Create(ctx, id, "gemeinsamer Name")
		if err != nil {
			return err
		}
		if err := d.mutiere(id, "Id=1"); err != nil {
			return err
		}
		if _, err := d.restore(id, snap.ID); err != nil {
			return err
		}
		l, err := d.snap.List(ctx, id)
		if err == nil && (len(l) != 1 || l[0].WorkspaceID != id) {
			err = fmt.Errorf("Liste von %s: %+v", id, l)
		}
		return err
	})
	for i, err := range fehler {
		if err != nil {
			t.Errorf("Workspace %d: %v", i, err)
		}
	}
	for _, id := range ids {
		if !reflect.DeepEqual(norm(d.state(id)), soll[id]) {
			t.Errorf("%s: State nicht im Soll", id)
		}
	}
}

// TestNamenParallel (T71): 16 gleichzeitige Create mit verschiedenen Namen: alle gelingen; bei gleichem State dieselbe
// Referenz und derselbe Hash, verschiedene Kennungen; 4 Runden.
func TestNamenParallel(t *testing.T) {
	d := snNeueDatei(t)
	for runde := 0; runde < 4; runde++ {
		id := fmt.Sprintf("ws-v%d", runde)
		d.legeAn(id)
		ergebnis := make([]model.WorkspaceSnapshot, 16)
		fehler := gestartet(16, func(i int) error {
			var err error
			ergebnis[i], err = d.snap.Create(context.Background(), id, fmt.Sprintf("name-%d", i))
			return err
		})
		kennungen := map[string]bool{}
		for i, err := range fehler {
			if err != nil {
				t.Fatalf("Runde %d Lauf %d: %v", runde, i, err)
			}
			kennungen[ergebnis[i].ID] = true
			if ergebnis[i].StateBlob != ergebnis[0].StateBlob || ergebnis[i].StateHash != ergebnis[0].StateHash {
				t.Errorf("Runde %d Lauf %d: Referenz verschieden", runde, i)
			}
		}
		if len(kennungen) != 16 {
			t.Errorf("Runde %d: %d verschiedene Kennungen", runde, len(kennungen))
		}
	}
}
