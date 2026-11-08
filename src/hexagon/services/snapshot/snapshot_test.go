package snapshot_test

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
)

var _ inbound.SnapshotUseCase = (*snapshot.Dienst)(nil)

// alleOps nennt jede Operation jedes Fakes (M10): Workspaces 5, Entities 5, Snapshots 5, Blobs 5, Writer 3, Uhr 1, IDs 1.
func alleOps() []string {
	var r []string
	for p, ops := range map[string][]string{
		"Workspaces": {"Get", "Create", "List", "Advance", "Delete"}, "Entities": {"List", "Put", "Get", "ListSet", "ReplaceAll"},
		"Snapshots": {"Get", "List", "CreateWorkspace", "GetWorkspace", "ListWorkspace"}, "Blobs": {"Begin", "Open", "Verify", "List", "Remove"},
		"Writer": {"Write", "Commit", "Abort"},
	} {
		for _, o := range ops {
			r = append(r, p+"."+o)
		}
	}
	return append(r, "Uhr.Now", "IDs.SnapshotID")
}

// portAufrufe zählt in dienst.go die Aufrufe der Formen Port.Operation (d.k.Blobs.Begin) und Writer (wr.Write).
func portAufrufe(t *testing.T) map[string]int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "dienst.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	r := map[string]int{}
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch x := sel.X.(type) {
		case *ast.SelectorExpr:
			if y, ok := x.X.(*ast.SelectorExpr); ok && y.Sel.Name == "k" {
				r[x.Sel.Name+"."+sel.Sel.Name]++
			}
		case *ast.Ident:
			if x.Name == "wr" {
				r["Writer."+sel.Sel.Name]++
			}
		}
		return true
	})
	return r
}

// TestKonfigSnapshot (T01): New liefert ErrUngueltig und keinen Dienst bei nil in jedem der sechs Felder und ruft keinen
// Port; jedes Feld dient nur dem Genannten (Quelltext: die Operation je Port gegen die Liste).
func TestKonfigSnapshot(t *testing.T) {
	w := neu(t)
	typ := reflect.TypeOf(snapshot.Konfig{})
	if typ.NumField() != 6 {
		t.Fatalf("Felder %d", typ.NumField())
	}
	for i := 0; i < typ.NumField(); i++ {
		k := w.konfig
		reflect.ValueOf(&k).Elem().Field(i).Set(reflect.Zero(typ.Field(i).Type))
		d, err := snapshot.New(k)
		if d != nil {
			t.Errorf("Feld %s: Dienst trotz nil", typ.Field(i).Name)
		}
		pruefeOhneCode(t, err, model.ErrUngueltig)
	}
	if _, err := snapshot.New(w.konfig); err != nil || w.p.gesamt() != 0 {
		t.Errorf("New: %v, Aufrufe %v", err, w.p.ops())
	}
	soll := map[string]int{"Workspaces.Get": 1, "Entities.List": 1, "Snapshots.Get": 1, "Snapshots.CreateWorkspace": 1, "Snapshots.ListWorkspace": 1,
		"Blobs.Begin": 1, "Uhr.Now": 1, "IDs.SnapshotID": 1, "Writer.Write": 1, "Writer.Commit": 1, "Writer.Abort": 2}
	got := portAufrufe(t)
	if !reflect.DeepEqual(got, soll) {
		t.Errorf("Aufrufstellen %v, erwartet %v", got, soll)
	}
}

// TestCtxNilSnapshot (T02): ctx nil ist bei Create und List ErrUngueltig ohne Fehler-Wert und ohne Code; kein Port, weder
// Uhr noch Generator.
func TestCtxNilSnapshot(t *testing.T) {
	var kein context.Context
	n := 0
	for _, op := range []func(w *welt) error{
		func(w *welt) error { _, err := w.d.Create(kein, "ws-1", "n"); return err },
		func(w *welt) error { _, err := w.d.List(kein, "ws-1"); return err },
	} {
		w := neu(t)
		pruefeOhneCode(t, op(w), model.ErrUngueltig)
		if w.p.gesamt() != 0 {
			t.Errorf("Aufrufe %v", w.p.ops())
		}
		n++
	}
	if n != 2 {
		t.Error("Operationen")
	}
}

// namensFaelle sind die 40 Namen von M1 mit der Soll-Tabelle von Hand; die Bytes entstehen aus Bytewerten.
func namensFaelle() (namen []string, gueltig []bool) {
	add := func(n string, g bool) { namen, gueltig = append(namen, n), append(gueltig, g) }
	for _, l := range []int{0, 1, 2, 127, 128, 129, 130, 300} {
		add(strings.Repeat("a", l), l >= 1 && l <= 128)
	}
	for _, b := range []byte{0x00, 0x01, 0x09, 0x0A, 0x1F, 0x7F} {
		add(string([]byte{b, 'b', 'c'}), false)
		add(string([]byte{'a', b, 'c'}), false)
		add(string([]byte{'a', 'b', b}), false)
	}
	for _, b := range [][]byte{{0xC3, 0xA9}, {0xE2, 0x82, 0xAC}, {0xF0, 0x9F, 0x98, 0x80}} {
		add("x"+string(b), true)
	}
	for _, b := range [][]byte{{0x80}, {0xC0, 0xAF}, {0xED, 0xA0, 0x80}, {0xFF}, {0xE2, 0x82}} {
		add("x"+string(b), false)
	}
	add(" x", true)
	add("x ", true)
	add("   ", true)
	add(strings.Repeat(string([]byte{0xC3, 0xA9}), 64), true)
	add(strings.Repeat(string([]byte{0xC3, 0xA9}), 65), false)
	add(strings.Repeat("a", 127)+string([]byte{0xC3, 0xA9}), false)
	return namen, gueltig
}

// TestNameForm (T03): M1, 40 Namen: 1 bis 128 Bytes, gültiges UTF-8, kein Byte unter 0x20, nicht 0x7F; Leerraum zählt; ein
// Verstoß ist ErrUngueltig ohne Code und ohne den Wert im Text, ruft keinen Port, keine Kennung, keine Zeit.
func TestNameForm(t *testing.T) {
	namen, gueltig := namensFaelle()
	if len(namen) != 40 {
		t.Fatalf("M1: %d Namen", len(namen))
	}
	for i, n := range namen {
		w := neu(t)
		_, err := w.d.Create(context.Background(), "ws-1", n)
		if gueltig[i] {
			if err != nil {
				t.Errorf("Name %d (%d Bytes): %v", i, len(n), err)
			}
			continue
		}
		pruefeOhneCode(t, err, model.ErrUngueltig)
		if w.p.gesamt() != 0 {
			t.Errorf("Name %d: Aufrufe %v", i, w.p.ops())
		}
	}
	w := neu(t)
	_, err := w.d.Create(context.Background(), "ws-1", "CANARY"+string([]byte{0}))
	if err == nil || strings.Contains(err.Error(), "CANARY") {
		t.Errorf("der Wert steht im Text: %v", err)
	}
}

// TestRangCreate (T04): die Rangfolge von Create (Stufe 9); ein Fehler einer Stufe lässt die späteren ungerufen; dazu die
// Paare Z1 bis Z4.
func TestRangCreate(t *testing.T) {
	stufen := []struct{ op, code, spaeter string }{
		{"Workspaces.Get", "ODR-E9001", "Snapshots.Get"}, {"Snapshots.Get", "ODR-E9001", "Entities.List"},
		{"Entities.List", "ODR-E9001", "Blobs.Begin"}, {"Blobs.Begin", "ODR-E9001", "Writer.Write"},
		{"Writer.Write", "ODR-E9001", "Writer.Commit"}, {"Writer.Commit", "ODR-E9001", "IDs.SnapshotID"},
		{"Snapshots.CreateWorkspace", "ODR-E9001", ""},
	}
	for _, s := range stufen {
		w := neu(t)
		w.p.setzeFehler(s.op, errPlatte)
		_, err := w.d.Create(context.Background(), "ws-1", "n")
		pruefeCode(t, err, s.code, errPlatte)
		if s.spaeter != "" && w.p.anzahl(s.spaeter) != 0 {
			t.Errorf("%s: %s gerufen", s.op, s.spaeter)
		}
	}
	// Stufe 3: Workspace unbekannt, Stufe 2: Name ungültig.
	w := neu(t)
	_, err := w.d.Create(context.Background(), "nope", "n")
	pruefeCode(t, err, "ODR-E7001", outbound.ErrNichtGefunden)
	if w.p.gesamt() != 1 {
		t.Errorf("Workspace unbekannt: %v", w.p.ops())
	}
	// Z1 Name ungültig und Workspace unbekannt: der Name geht dem Port voraus.
	w = neu(t)
	_, err = w.d.Create(context.Background(), "nope", "")
	pruefeOhneCode(t, err, model.ErrUngueltig)
	if w.p.gesamt() != 0 {
		t.Error("Z1: Port gerufen")
	}
	// Z2 Name vergeben und Workspace unbekannt: ODR-E7001.
	w = neu(t)
	if _, _, err := w.erzeuge("n"); err != nil {
		t.Fatal(err)
	}
	_, err = w.d.Create(context.Background(), "nope", "n")
	pruefeCode(t, err, "ODR-E7001", nil)
	// Z3 Name vergeben und Fehler des Blobs: der Fehler des Blobs, der Name wird nie geprüft.
	for _, op := range []string{"Blobs.Begin", "Writer.Write", "Writer.Commit"} {
		w := neu(t)
		if _, _, err := w.erzeuge("n"); err != nil {
			t.Fatal(err)
		}
		w.p.leeren()
		w.p.setzeFehler(op, errPlatte)
		_, err := w.d.Create(context.Background(), "ws-1", "n")
		pruefeCode(t, err, "ODR-E9001", errPlatte)
		if w.p.anzahl("Snapshots.CreateWorkspace") != 0 {
			t.Errorf("Z3 %s: CreateWorkspace gerufen", op)
		}
	}
	// Z4 Name vergeben und Workspace inzwischen gelöscht: ODR-E7001 geht dem Namen voraus.
	w = neu(t)
	if _, _, err := w.erzeuge("n"); err != nil {
		t.Fatal(err)
	}
	w.p.nach = func(op string) {
		if op == "Workspaces.Get" {
			w.p.nach = nil
			w.ws.loeschen("ws-1")
		}
	}
	_, err = w.d.Create(context.Background(), "ws-1", "n")
	pruefeCode(t, err, "ODR-E7001", nil)
}

// TestArgumenteSnapshot (T05): workspaceID geht unverändert an jeden Port, auch leer, ohne Form, mit Großbuchstaben
// (Wert 6 × Operation 2); die Antwort ist ODR-E7001.
func TestArgumenteSnapshot(t *testing.T) {
	werte := []string{"", "A", "x y", strings.Repeat("k", 200), "a" + string([]byte{0}) + "b", "wsp-x"}
	n := 0
	for _, v := range werte {
		w := neu(t)
		_, err := w.d.Create(context.Background(), v, "n")
		pruefeCode(t, err, "ODR-E7001", outbound.ErrNichtGefunden)
		if ev := w.p.von("Workspaces.Get"); len(ev) != 1 || ev[0].args[0] != v {
			t.Errorf("Create %q: %v", v, ev)
		}
		w = neu(t)
		_, err = w.d.List(context.Background(), v)
		pruefeCode(t, err, "ODR-E7001", outbound.ErrNichtGefunden)
		if ev := w.p.von("Snapshots.ListWorkspace"); len(ev) != 1 || ev[0].args[0] != v {
			t.Errorf("List %q: %v", v, ev)
		}
		n += 2
	}
	if n != 12 {
		t.Errorf("%d Elemente", n)
	}
}

// TestLesenSnapshot (T06): Workspaces.Get 1×, Snapshots.Get 1× mit der Baseline des Workspace, Entities.List 1× (ein
// atomarer Aufruf, nie ListSet je Set), sonst keine Operation; die Fixture hat drei Mengen.
func TestLesenSnapshot(t *testing.T) {
	w := neu(t)
	if _, _, err := w.erzeuge("n"); err != nil {
		t.Fatal(err)
	}
	soll := map[string]int{"Workspaces.Get": 1, "Snapshots.Get": 1, "Entities.List": 1, "Blobs.Begin": 1, "Writer.Write": 1, "Writer.Commit": 1,
		"Snapshots.CreateWorkspace": 1, "IDs.SnapshotID": 1, "Uhr.Now": 1}
	ops := alleOps()
	if len(ops) != 25 {
		t.Fatalf("M10: %d Operationen", len(ops))
	}
	for _, op := range ops {
		if got := w.p.anzahl(op); got != soll[op] {
			t.Errorf("%s: %d mal, erwartet %d", op, got, soll[op])
		}
	}
	if got := w.p.von("Snapshots.Get")[0].args[0]; got != "bsl-1" {
		t.Errorf("Baseline %v", got)
	}
	if got := w.p.von("Workspaces.Get")[0].args[0]; got != "ws-1" {
		t.Errorf("Workspace %v", got)
	}
	sets := map[string]bool{}
	for _, e := range standard("ws-1") {
		sets[e.EntitySet] = true
	}
	if len(sets) != 2 {
		t.Errorf("Fixture: %d Mengen", len(sets))
	}
}

// TestBlobAblauf (T15): Begin(ctx, MaxInt64) 1×, alle Bytes über Write, Commit 1×; Abort genau 1× nach gescheitertem Write
// oder Commit, nie nach Begin-Fehler und nie nach gelungenem Commit; ein Fehler von Abort ist ohne Wirkung (Fehlerstelle 4 ×
// Fehler von Abort 2 = 8).
func TestBlobAblauf(t *testing.T) {
	n := 0
	for _, stelle := range []string{"", "Blobs.Begin", "Writer.Write", "Writer.Commit"} {
		for _, af := range []error{nil, errPlatte} {
			w := neu(t)
			w.blobs.abortFehler = af
			if stelle != "" {
				w.p.setzeFehler(stelle, errPlatte)
			}
			_, err := w.d.Create(context.Background(), "ws-1", "n")
			begin, write, commit, abort := 1, 1, 1, 1
			switch stelle {
			case "":
				abort = 0
			case "Blobs.Begin":
				write, commit, abort = 0, 0, 0
			case "Writer.Write":
				commit = 0
			}
			for op, soll := range map[string]int{"Blobs.Begin": begin, "Writer.Write": write, "Writer.Commit": commit, "Writer.Abort": abort} {
				if got := w.p.anzahl(op); got != soll {
					t.Errorf("Stelle %q Abort-Fehler %v: %s %d mal, erwartet %d", stelle, af, op, got, soll)
				}
			}
			if stelle == "" {
				if err != nil {
					t.Errorf("Erfolg: %v", err)
				}
				var teile []byte
				for _, e := range w.p.von("Writer.Write") {
					teile = append(teile, e.args[0].([]byte)...)
				}
				if len(teile) == 0 || !bytes.HasPrefix(teile, []byte(`{"entities":`)) {
					t.Errorf("Bytes des Writers: %q", teile)
				}
			} else {
				pruefeCode(t, err, "ODR-E9001", errPlatte)
			}
			if g := w.p.von("Blobs.Begin")[0].args[0]; g != int64(9223372036854775807) {
				t.Errorf("Grenze %v", g)
			}
			n++
		}
	}
	if n != 8 {
		t.Errorf("%d Elemente", n)
	}
}

// TestReferenzGleichHash (T16): StateBlob und StateHash sind die Referenz aus Commit, nicht ein eigener Hash der Bytes;
// MetadataHash ist der der Baseline; Validate() gilt.
func TestReferenzGleichHash(t *testing.T) {
	w := neu(t)
	fremd := "sha256:" + strings.Repeat("f", 64)
	w.blobs.fremd = fremd
	s, _, err := w.erzeuge("n")
	if err != nil {
		t.Fatal(err)
	}
	if s.StateBlob != fremd || s.StateHash != fremd || s.MetadataHash != metadatenHash() || s.Validate() != nil {
		t.Errorf("%+v", s)
	}
}

// TestKennungZeit (T17): IDs.SnapshotID und Uhr.Now je genau 1× und nach dem Blob; ein Aufruf, der vor diesem Schritt endet,
// verbraucht keine (Stufe 8); zwei Aufrufe liefern die Kennungen 1 und 2.
func TestKennungZeit(t *testing.T) {
	for _, op := range []string{"Workspaces.Get", "Snapshots.Get", "Entities.List", "Blobs.Begin", "Writer.Write", "Writer.Commit"} {
		w := neu(t)
		w.p.setzeFehler(op, errPlatte)
		_, _ = w.d.Create(context.Background(), "ws-1", "n")
		if w.p.anzahl("IDs.SnapshotID") != 0 || w.p.anzahl("Uhr.Now") != 0 {
			t.Errorf("%s: Kennung oder Zeit verbraucht", op)
		}
	}
	w := neu(t)
	_, _ = w.d.Create(context.Background(), "ws-1", "")
	_, _ = w.d.Create(context.Background(), "nope", "n")
	if w.p.anzahl("IDs.SnapshotID") != 0 || w.p.anzahl("Uhr.Now") != 0 {
		t.Error("Name oder Workspace: Kennung oder Zeit verbraucht")
	}
	s1, _, err := w.erzeuge("a")
	if err != nil {
		t.Fatal(err)
	}
	s2, _, _ := w.erzeuge("b")
	null31 := strings.Repeat("0", 31)
	if s1.ID != "snp-"+null31+"1" || s2.ID != "snp-"+null31+"2" || !s1.CreatedAt.Equal(t0()) {
		t.Errorf("Kennungen %s %s Zeit %v", s1.ID, s2.ID, s1.CreatedAt)
	}
	if w.p.anzahl("IDs.SnapshotID") != 2 || w.p.anzahl("Uhr.Now") != 2 {
		t.Errorf("Zähler %d %d", w.p.anzahl("IDs.SnapshotID"), w.p.anzahl("Uhr.Now"))
	}
	ops := w.p.ops()
	c, id, now := slices.Index(ops, "Writer.Commit"), slices.Index(ops, "IDs.SnapshotID"), slices.Index(ops, "Uhr.Now")
	if c < 0 || c >= id || id >= now {
		t.Errorf("Reihenfolge Commit %d, Kennung %d, Zeit %d", c, id, now)
	}
}

// TestCreateWorkspaceArgumente (T18): CreateWorkspace 1× mit {ID, Service des Workspace, Workspace, Name wie geliefert,
// Zeit, StateBlob, StateHash, MetadataHash}; das Ergebnis ist der übergebene Wert; die Fixture trägt für jedes Feld einen
// anderen Wert (Service des Workspace und der Baseline verschieden).
func TestCreateWorkspaceArgumente(t *testing.T) {
	w := neu(t)
	w.ws.mu.Lock()
	x := w.ws.ws["ws-1"]
	x.ServiceID = "svc-w"
	w.ws.ws["ws-1"] = x
	w.ws.mu.Unlock()
	w.snaps.mu.Lock()
	b := w.snaps.baselines["bsl-1"]
	b.ServiceID = "svc-b"
	w.snaps.baselines["bsl-1"] = b
	w.snaps.mu.Unlock()
	w.festUhr.Advance(time.Hour)
	s, bs, err := w.erzeuge(" Name mit Rand ")
	if err != nil {
		t.Fatal(err)
	}
	ev := w.p.von("Snapshots.CreateWorkspace")
	if len(ev) != 1 || !reflect.DeepEqual(ev[0].args[0], s) {
		t.Fatalf("CreateWorkspace: %v", ev)
	}
	soll := model.WorkspaceSnapshot{ID: "snp-" + strings.Repeat("0", 31) + "1", ServiceID: "svc-w", WorkspaceID: "ws-1", Name: " Name mit Rand ",
		CreatedAt: t0().Add(time.Hour), StateBlob: s.StateBlob, StateHash: s.StateHash, MetadataHash: metadatenHash()}
	if s != soll || !bytes.Contains(bs, []byte(`"service_id":"svc-w"`)) {
		t.Errorf("%+v, erwartet %+v", s, soll)
	}
	werte := []string{s.ID, s.ServiceID, s.WorkspaceID, s.Name, s.StateBlob, s.MetadataHash, "svc-b"}
	for i := range werte {
		for j := i + 1; j < len(werte); j++ {
			if werte[i] == werte[j] {
				t.Errorf("Fixture: Felder %d und %d tragen denselben Wert", i, j)
			}
		}
	}
}

// TestNameVergeben (T19): M17, Paar 4: derselbe Name im selben Workspace ist ODR-E7006 (409, workspace, Snapshot name
// already taken) und lässt die Liste unverändert; derselbe Name in einem anderen Workspace, in anderer Schreibweise und
// der Name eines Snapshots eines Workspace eines anderen Service gelten.
func TestNameVergeben(t *testing.T) {
	w := neu(t)
	w.ws.mu.Lock()
	w.ws.ws["ws-3"] = model.Workspace{ID: "ws-3", ServiceID: "svc-andere", BaselineSnapshotID: "bsl-1", CreatedAt: t0()}
	w.ws.mu.Unlock()
	w.ent.setze("ws-3", []model.EntityState{})
	if _, _, err := w.erzeuge("Name"); err != nil {
		t.Fatal(err)
	}
	vorher := w.snaps.alle()
	_, _, err := w.erzeuge("Name")
	pruefeCode(t, err, "ODR-E7006", outbound.ErrBereitsVorhanden)
	if f, _ := codeVon(err); f.Status != 409 || f.Code.Zielbereich() != model.ZielWorkspace || f.Error() != "ODR-E7006: Snapshot name already taken" {
		t.Errorf("Literale: %v", err)
	}
	if !reflect.DeepEqual(vorher, w.snaps.alle()) {
		t.Error("Liste geändert")
	}
	for _, c := range []struct{ ws, name string }{{"ws-2", "Name"}, {"ws-1", "name"}, {"ws-3", "Name"}} {
		if _, err := w.d.Create(context.Background(), c.ws, c.name); err != nil {
			t.Errorf("%s %q: %v", c.ws, c.name, err)
		}
	}
}

// TestAbbruchNachCommit (T23): der Abbruch nach Commit vor CreateWorkspace lässt keinen Snapshot, Abort 0 und den Blob;
// die Antwort ist der Kontextfehler ohne Code.
func TestAbbruchNachCommit(t *testing.T) {
	w := neu(t)
	ctx, abbruch := context.WithCancel(context.Background())
	w.p.nach = func(op string) {
		if op == "Writer.Commit" {
			abbruch()
		}
	}
	_, err := w.d.Create(ctx, "ws-1", "n")
	pruefeOhneCode(t, err, context.Canceled)
	if len(w.snaps.alle()) != 0 || w.p.anzahl("Writer.Abort") != 0 || w.p.anzahl("Snapshots.CreateWorkspace") != 1 || len(w.blobs.inhalt) != 1 {
		t.Errorf("Snapshots %d, Abort %d, CreateWorkspace %d, Blobs %d", len(w.snaps.alle()), w.p.anzahl("Writer.Abort"),
			w.p.anzahl("Snapshots.CreateWorkspace"), len(w.blobs.inhalt))
	}
}

// TestCreateOhneSchreiben (T24): Create ändert weder State noch Cursors noch Revision und öffnet keine Transaktion:
// die verbotenen Operationen (Entities.Put, ReplaceAll, Get, ListSet, Workspaces.Advance, Create, Delete) zählen 0 bei
// Erfolg und bei jedem Fehler von M7; Konfig hat kein Feld für Cursors oder eine Transaktion, der Quelltext nennt weder
// Cursor noch Transaktion noch Advance.
func TestCreateOhneSchreiben(t *testing.T) {
	verboten := []string{"Entities.Put", "Entities.ReplaceAll", "Entities.Get", "Entities.ListSet", "Workspaces.Advance", "Workspaces.Create", "Workspaces.Delete"}
	laeufe := 0
	pruefe := func(w *welt, was string) {
		for _, op := range verboten {
			if w.p.anzahl(op) != 0 {
				t.Errorf("%s: %s gerufen", was, op)
			}
		}
		laeufe++
	}
	w := neu(t)
	_, _, _ = w.erzeuge("n")
	pruefe(w, "Erfolg")
	for _, s := range stellenSnapshot() {
		if !s.fehler || s.op != "Create" || s.label == "Writer.Abort" {
			continue
		}
		for _, fe := range m7Fehler() {
			w := neu(t)
			w.p.setzeFehler(s.label, fe.err)
			_, _ = w.d.Create(context.Background(), "ws-1", "n")
			pruefe(w, s.nr+" "+fe.name)
		}
	}
	if laeufe != 1+7*8 {
		t.Errorf("%d Läufe", laeufe)
	}
	typ := reflect.TypeOf(snapshot.Konfig{})
	for i := 0; i < typ.NumField(); i++ {
		if n := typ.Field(i).Name; n == "Cursors" || n == "Transaktion" {
			t.Errorf("Feld %s", n)
		}
	}
	for _, id := range identifikatoren(t, "dienst.go", "manifest.go", "fehler.go") {
		if strings.Contains(id, "Cursor") || strings.Contains(id, "Transaktion") || id == "Advance" {
			t.Errorf("Quelltext nennt %s", id)
		}
	}
}

// identifikatoren liefert die Namen aller Bezeichner der Dateien.
func identifikatoren(t *testing.T, dateien ...string) []string {
	t.Helper()
	var r []string
	for _, d := range dateien {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				r = append(r, id.Name)
			}
			return true
		})
	}
	return r
}

// TestUnveraenderlich (T25): ein Snapshot ändert sich nie: der Dienst ruft von Snapshots nur Get, CreateWorkspace und
// ListWorkspace, je an genau einer Stelle; GetWorkspace und List (verbotene Operationen) zählen 0.
func TestUnveraenderlich(t *testing.T) {
	got := portAufrufe(t)
	n := 0
	for k, c := range got {
		if strings.HasPrefix(k, "Snapshots.") {
			if !slices.Contains([]string{"Snapshots.Get", "Snapshots.CreateWorkspace", "Snapshots.ListWorkspace"}, k) || c != 1 {
				t.Errorf("Stelle %s: %d", k, c)
			}
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d Operationen von Snapshots", n)
	}
	w := neu(t)
	_, _, _ = w.erzeuge("a")
	_, _ = w.d.List(context.Background(), "ws-1")
	for _, op := range []string{"Snapshots.GetWorkspace", "Snapshots.List"} {
		if w.p.anzahl(op) != 0 {
			t.Errorf("%s gerufen", op)
		}
	}
}

// TestListe (T26): List ruft ListWorkspace 1× mit der Kennung unverändert und liefert die Folge des Ports unverändert
// (aufsteigend, absteigend, gemischt, auch bei gleicher Zeit), eine leere Liste statt nil, ODR-E7001 bei unbekanntem
// Workspace (Position erster, mittlerer, letzter bei 5 Snapshots × Folge 3).
func TestListe(t *testing.T) {
	var fuenf []model.WorkspaceSnapshot
	for i := 0; i < 5; i++ {
		fuenf = append(fuenf, model.WorkspaceSnapshot{ID: fmt.Sprintf("snp-%d", i), ServiceID: "svc", WorkspaceID: "ws-1", Name: fmt.Sprintf("n%d", i),
			CreatedAt: t0().Add(time.Duration(i/2) * time.Second)})
	}
	folgen := map[string][]int{"aufsteigend": {0, 1, 2, 3, 4}, "absteigend": {4, 3, 2, 1, 0}, "gemischt": {2, 4, 0, 3, 1}}
	n := 0
	for name, f := range folgen {
		w := neu(t)
		w.snaps.snaps = fuenf
		w.snaps.liste = func(_ string, _ []model.WorkspaceSnapshot) []model.WorkspaceSnapshot {
			var r []model.WorkspaceSnapshot
			for _, i := range f {
				r = append(r, fuenf[i])
			}
			return r
		}
		got, err := w.d.List(context.Background(), "ws-1")
		if err != nil || len(got) != 5 {
			t.Fatalf("%s: %v (%d)", name, err, len(got))
		}
		for _, pos := range []int{0, 2, 4} {
			if got[pos].ID != fuenf[f[pos]].ID {
				t.Errorf("%s Position %d: %s", name, pos, got[pos].ID)
			}
			n++
		}
		if ev := w.p.von("Snapshots.ListWorkspace"); len(ev) != 1 || ev[0].args[0] != "ws-1" || w.p.gesamt() != 1 {
			t.Errorf("%s: Aufrufe %v", name, w.p.ops())
		}
	}
	w := neu(t)
	w.snaps.nilLeer = true
	got, err := w.d.List(context.Background(), "ws-2")
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("leer: %v %v", got, err)
	}
	_, err = w.d.List(context.Background(), "nope")
	pruefeCode(t, err, "ODR-E7001", outbound.ErrNichtGefunden)
	if n != 9 {
		t.Errorf("%d Elemente", n)
	}
}

// TestIsolationSnapshot (T27): zwei Workspaces mit gleichen Namen der Snapshots; List zeigt nie die des anderen, und jeder
// Aufruf der Ports nennt die Kennung des Workspace des Aufrufs.
func TestIsolationSnapshot(t *testing.T) {
	w := neu(t)
	w.ent.setze("ws-2", standard("ws-2"))
	for _, id := range []string{"ws-1", "ws-2"} {
		for _, name := range []string{"eins", "zwei"} {
			if _, err := w.d.Create(context.Background(), id, name); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, id := range []string{"ws-1", "ws-2"} {
		l, err := w.d.List(context.Background(), id)
		if err != nil || len(l) != 2 {
			t.Fatalf("%s: %v (%d)", id, err, len(l))
		}
		for _, s := range l {
			if s.WorkspaceID != id {
				t.Errorf("%s zeigt %s", id, s.WorkspaceID)
			}
		}
	}
	for _, op := range []string{"Workspaces.Get", "Entities.List", "Snapshots.ListWorkspace"} {
		for i, e := range w.p.von(op) {
			if want := []string{"ws-1", "ws-1", "ws-2", "ws-2"}[i%4]; op != "Snapshots.ListWorkspace" && e.args[0] != want {
				t.Errorf("%s Aufruf %d: %v", op, i, e.args[0])
			}
		}
	}
	for i, e := range w.p.von("Snapshots.CreateWorkspace") {
		if e.args[0].(model.WorkspaceSnapshot).WorkspaceID != []string{"ws-1", "ws-1", "ws-2", "ws-2"}[i] {
			t.Errorf("CreateWorkspace %d: %v", i, e.args[0])
		}
	}
}

// TestV2Snapshot (T28): jeder Modus und jede Version: derselbe Ablauf, der Dienst liest weder Dienste noch Version noch
// Modell; das Ergebnis eines V2-Service gleicht dem des V4-Service bis auf service_id.
func TestV2Snapshot(t *testing.T) {
	typ := reflect.TypeOf(snapshot.Konfig{})
	var namen []string
	for i := 0; i < typ.NumField(); i++ {
		namen = append(namen, typ.Field(i).Name)
	}
	if !slices.Equal(namen, []string{"Workspaces", "Entities", "Snapshots", "Blobs", "Uhr", "IDs"}) {
		t.Errorf("Felder %v", namen)
	}
	_, v4, _ := erzeugeMitState(t, standard("ws-1"))
	w := neu(t)
	w.ws.mu.Lock()
	x := w.ws.ws["ws-1"]
	x.ServiceID = "v2-svc"
	w.ws.ws["ws-1"] = x
	w.ws.mu.Unlock()
	_, v2, err := w.erzeuge("n")
	if err != nil || !bytes.Equal(bytes.ReplaceAll(v2, []byte(`"service_id":"v2-svc"`), []byte(`"service_id":"svc"`)), v4) {
		t.Errorf("V2: %v", err)
	}
}
