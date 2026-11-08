package workspace_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// restoreWelt ist eine Welt, in der Restore("ws-1", "snp-1") lauffähig ist: der Workspace trägt Revision 5 und einen
// veränderten State samt Cursors, der Snapshot einen dritten State (snapshotState).
func restoreWelt(t *testing.T) *welt {
	t.Helper()
	w := neu(t)
	w.vorbelegen("ws-1", 5)
	w.veraendern("ws-1")
	w.snapshotAblegen("ws-1", "snp-1", snapshotState("ws-1"))
	w.p.leeren()
	return w
}

func restoreLauf(w *welt) (model.Workspace, error) {
	return w.v.Restore(context.Background(), "ws-1", "snp-1")
}

// pruefeLiteral prüft Code, Status und Text eines Fehlers gegen Literale aus dem Vertrag (SPEC-025, SPEC-065,
// SPEC-094, SPEC-096), nicht gegen Konstanten des Pakets.
func pruefeLiteral(t *testing.T, err error, code string, status int, text string) {
	t.Helper()
	f, ok := codeVon(err)
	if !ok {
		t.Errorf("kein *Fehler: %v", err)
		return
	}
	if f.Code.String() != code || f.Status != status || f.Error() != code+": "+text {
		t.Errorf("%q Status %d Text %q, erwartet %s %d %q", f.Code.String(), f.Status, f.Error(), code, status, code+": "+text)
	}
}

// restoreStellen sind die neun Stellen mit Fehlerrückgabe von Restore (SPEC-129 Nr. 15: S4, S5, S7, S9, S10, S11, S13,
// S19, S20).
func restoreStellen() []stelle {
	var r []stelle
	for _, s := range stellen() {
		if s.op == "Restore" {
			r = append(r, s)
		}
	}
	return r
}

// TestRestoreArgumente (T29): ctx nil ist ErrUngueltig ohne Code und ohne Port; Workspace- und Snapshot-Kennung gehen
// unverändert an die Ports, auch leer, ohne Form, mit Großbuchstaben, 200 Bytes, mit NUL.
func TestRestoreArgumente(t *testing.T) {
	var kein context.Context
	w := restoreWelt(t)
	_, err := w.v.Restore(kein, "ws-1", "snp-1")
	pruefeOhneCode(t, err, model.ErrUngueltig)
	if w.p.gesamt() != 0 {
		t.Errorf("ctx nil ruft Ports: %v", w.p.ops())
	}
	werte := []string{"", "A", "x y", strings.Repeat("k", 200), "a" + string([]byte{0}) + "b", "wsp-x"}
	if len(werte) != 6 {
		t.Fatal("Werte")
	}
	for _, v := range werte {
		w := restoreWelt(t)
		_, err := w.v.Restore(context.Background(), v, "snp-1")
		pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
		if ev := w.p.von("Workspaces.Get"); len(ev) != 1 || ev[0].args[0] != v {
			t.Errorf("Workspace-Kennung %q: %v", v, ev)
		}
		if w.p.anzahl("Snapshots.GetWorkspace") != 0 {
			t.Errorf("Workspace-Kennung %q: GetWorkspace gerufen", v)
		}
		w = restoreWelt(t)
		_, err = w.v.Restore(context.Background(), "ws-1", v)
		pruefeCode(t, err, model.CodeE7002, outbound.ErrNichtGefunden)
		ev := w.p.von("Snapshots.GetWorkspace")
		if len(ev) != 1 || ev[0].args[0] != "ws-1" || ev[0].args[1] != v {
			t.Errorf("Snapshot-Kennung %q: %v", v, ev)
		}
	}
}

// alleOps nennt jede Operation jedes Fakes außer Strom und Uhr, in und außerhalb der Transaktion (M10).
func alleOps() []string {
	var r []string
	ports := []struct {
		name string
		ops  []string
		tx   bool
	}{
		{"Workspaces", []string{"Create", "Get", "List", "Advance", "Delete"}, true},
		{"Entities", []string{"Put", "Get", "List", "ListSet", "ReplaceAll"}, true},
		{"Cursors", []string{"Put", "Get", "List", "ResetAll"}, true},
		{"Snapshots", []string{"Get", "List", "CreateWorkspace", "GetWorkspace", "ListWorkspace"}, false},
		{"Blobs", []string{"Begin", "Open", "Verify", "List", "Remove"}, false},
		{"Dienste", []string{"List", "Create", "Get", "SetMetadata"}, false},
		{"Wartung", []string{"Startup", "CleanupBlobs"}, false},
	}
	for _, p := range ports {
		for _, o := range p.ops {
			r = append(r, p.name+"."+o)
			if p.tx {
				r = append(r, "tx "+p.name+"."+o)
			}
		}
	}
	return append(r, lblTx, "Uhr.Now", "IDs.WorkspaceID", "IDs.SnapshotID")
}

// TestRestoreAblauf (T30): Workspaces.Get, Snapshots.GetWorkspace, Blobs.Open (mit StateBlob), Lesen, Close und die
// Transaktion mit Get, ReplaceAll und Advance je genau einmal, sonst keine Operation eines Fakes.
func TestRestoreAblauf(t *testing.T) {
	w := restoreWelt(t)
	if _, err := restoreLauf(w); err != nil {
		t.Fatal(err)
	}
	soll := map[string]int{"Workspaces.Get": 1, "Snapshots.GetWorkspace": 1, "Blobs.Open": 1, lblTx: 1, "tx Workspaces.Get": 1,
		"tx Entities.ReplaceAll": 1, "tx Workspaces.Advance": 1}
	ops := alleOps()
	if len(ops) != 48 {
		t.Fatalf("Zählprüfung: %d Operationen", len(ops))
	}
	for _, op := range ops {
		if got := w.p.anzahl(op); got != soll[op] {
			t.Errorf("%s: %d mal, erwartet %d", op, got, soll[op])
		}
	}
	folge := slices.DeleteFunc(w.p.ops(), func(op string) bool { return op == "Strom.Read" })
	want := []string{"Workspaces.Get", "Snapshots.GetWorkspace", "Blobs.Open", "Strom.Close", lblTx, "tx Workspaces.Get",
		"tx Entities.ReplaceAll", "tx Workspaces.Advance"}
	if !slices.Equal(folge, want) {
		t.Errorf("Folge %v, erwartet %v", folge, want)
	}
	s := w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}]
	if s.StateBlob == s.MetadataHash {
		t.Fatal("Fixture: StateBlob gleicht MetadataHash")
	}
	if got := w.p.von("Blobs.Open")[0].args[0]; got != s.StateBlob {
		t.Errorf("Open mit %v, erwartet StateBlob %v", got, s.StateBlob)
	}
}

// mutierterState legt an Position pos (0 erster, 1 mittlerer, 2 letzter von fünf) eine Änderung (art 0), einen neuen
// Eintrag (1) oder einen Tombstone (2) in eine Kopie von es.
func mutierterState(es []model.EntityState, pos, art int) []model.EntityState {
	r := kopiereEntities(es)
	i := []int{0, len(r) / 2, len(r) - 1}[pos]
	switch art {
	case 0:
		r[i].Payload = `{"Id":"x","Geaendert":true}`
		r[i].Version += 10
		r[i].Completeness = model.EntityComplete
	case 1:
		extra := model.EntityState{WorkspaceID: r[i].WorkspaceID, EntitySet: "Extra", EntityKey: fmt.Sprintf("Id=%d", i), Payload: `{"Id":"e"}`,
			Completeness: model.EntityComplete, Version: 1}
		r = slices.Insert(r, i, extra)
	default:
		r[i].Deleted = !r[i].Deleted
		r[i].Version += 3
	}
	return r
}

// fuenfEintraege ist der State der Position-Tests: fünf Einträge in zwei Mengen mit verschiedenen Versionen.
func fuenfEintraege(wid string) []model.EntityState {
	var r []model.EntityState
	for i := 0; i < 5; i++ {
		c := model.EntityComplete
		if i%2 == 1 {
			c = model.EntityPartial
		}
		r = append(r, model.EntityState{WorkspaceID: wid, EntitySet: []string{"Orders", "Products"}[i%2], EntityKey: fmt.Sprintf("Id=%d", i),
			Payload: fmt.Sprintf(`{"Id":"%d","Wert":%d}`, i, i*7), KnownProperties: []string{"Id", "Wert"}[:1+i%2], Completeness: c,
			Version: int64(2*i + 1), Deleted: i == 3})
	}
	return r
}

// TestRestoreWirkung (T31): nach Restore ist der State genau der des Manifests (M14: Position 3 × Art 3 = 9); State
// vorher, im Snapshot, nach der Mutation und danach sind verschieden, die Folge an ReplaceAll ist die des Manifests.
func TestRestoreWirkung(t *testing.T) {
	n := 0
	for pos := 0; pos < 3; pos++ {
		for art := 0; art < 3; art++ {
			w := neu(t)
			w.vorbelegen("ws-1", 5)
			soll := fuenfEintraege("ws-1")
			w.snapshotAblegen("ws-1", "snp-1", soll)
			z := w.st.stand()
			z.ent["ws-1"] = mutierterState(soll, pos, art)
			w.st.setze(z)
			vorher := w.st.stand().ent["ws-1"]
			if equalState(vorher, soll) {
				t.Fatalf("Pos %d Art %d: Fixture gleicht dem Snapshot", pos, art)
			}
			w.p.leeren()
			if _, err := restoreLauf(w); err != nil {
				t.Fatalf("Pos %d Art %d: %v", pos, art, err)
			}
			if got := w.st.stand().ent["ws-1"]; !equalState(got, soll) {
				t.Errorf("Pos %d Art %d: State %v, erwartet %v", pos, art, got, soll)
			}
			if got := w.p.von("tx Entities.ReplaceAll")[0].args[1].([]model.EntityState); !equalState(got, soll) {
				t.Errorf("Pos %d Art %d: Folge an ReplaceAll %v", pos, art, got)
			}
			n++
		}
	}
	if n != 9 {
		t.Errorf("M14: %d Elemente", n)
	}
}

// TestRestoreRevision (T32): die Revision wächst bei jedem Restore um 1, auch beim selben Snapshot dreimal
// hintereinander; Advance erhält die in der Transaktion gelesene; ID, Service, Baseline und Zeit bleiben.
func TestRestoreRevision(t *testing.T) {
	n := 0
	for _, rev := range []int64{0, 5, 1 << 40} {
		w := neu(t)
		w.vorbelegen("ws-1", rev)
		w.snapshotAblegen("ws-1", "snp-1", snapshotState("ws-1"))
		vorher := w.st.stand().ws["ws-1"]
		for i := int64(1); i <= 3; i++ {
			w.p.leeren()
			got, err := restoreLauf(w)
			if err != nil {
				t.Fatal(err)
			}
			if got.Revision != rev+i || got.ID != vorher.ID || got.ServiceID != vorher.ServiceID ||
				got.BaselineSnapshotID != vorher.BaselineSnapshotID || !got.CreatedAt.Equal(vorher.CreatedAt) {
				t.Errorf("Revision %d Lauf %d: %+v", rev, i, got)
			}
			if a := w.p.von("tx Workspaces.Advance")[0].args[1]; a != rev+i-1 {
				t.Errorf("Revision %d Lauf %d: Advance mit %v", rev, i, a)
			}
			if w.st.stand().ws["ws-1"].Revision != rev+i {
				t.Errorf("Revision %d Lauf %d: gespeichert %d", rev, i, w.st.stand().ws["ws-1"].Revision)
			}
			n++
		}
	}
	if n != 9 {
		t.Errorf("%d Läufe", n)
	}
}

// restoreFehlerLaeufe ruft fn für jeden Fehlerlauf von Restore (jede Stelle × jeder Fehler der Tabelle samt dem ohne
// Zeile, ohne S19 mit ErrFalscheArt); es liefert die Zahl der Läufe.
func restoreFehlerLaeufe(t *testing.T, fn func(w *welt, s stelle, fe tabellenFehler, err error)) int {
	n := 0
	for _, s := range restoreStellen() {
		for _, fe := range tabellenFehlerListe() {
			if s.nr == "S19" && fe.err == outbound.ErrFalscheArt {
				continue
			}
			w, o := s.vorbereitet(t)
			w.p.setzeFehler(s.label, fe.err)
			w.p.leeren()
			fn(w, s, fe, o.lauf(w, context.Background()))
			n++
		}
	}
	return n
}

// TestRestoreCursorNull (T33): Restore ruft keine Operation von Cursors, weder lesend noch schreibend, auch nicht bei
// einem Fehler; der Zähler jeder der acht Operationen (mit und ohne Transaktion) bleibt in allen Läufen 0.
func TestRestoreCursorNull(t *testing.T) {
	if len(cursorOps()) != 8 {
		t.Fatal("Cursor-Operationen")
	}
	w := restoreWelt(t)
	vorher := w.st.stand().cur["ws-1"]
	if _, err := restoreLauf(w); err != nil || w.cursorAufrufe() != 0 {
		t.Errorf("Erfolg: %v, Cursor-Aufrufe %d", err, w.cursorAufrufe())
	}
	if nachher := w.st.stand().cur["ws-1"]; len(vorher) != 2 || !reflect.DeepEqual(vorher, nachher) {
		t.Errorf("Cursors geändert: %v -> %v", vorher, nachher)
	}
	n := restoreFehlerLaeufe(t, func(w *welt, s stelle, fe tabellenFehler, err error) {
		if err == nil || w.cursorAufrufe() != 0 {
			t.Errorf("%v %s: Fehler %v, Cursor-Aufrufe %d", s, fe.name, err, w.cursorAufrufe())
		}
	})
	if n != 9*len(tabellenFehlerListe())-1 {
		t.Errorf("%d Fehlerläufe", n)
	}
}

// TestRestoreUnbekannt (T34): eine unbekannte, die leere und die Kennung eines Snapshots eines anderen Workspace
// desselben Service ergeben ODR-E7002 (404) ohne Teilzustand.
func TestRestoreUnbekannt(t *testing.T) {
	for _, id := range []string{"snp-gibtsnicht", "", "snp-fremd"} {
		w := restoreWelt(t)
		w.vorbelegen("ws-2", 0)
		w.snapshotAblegen("ws-2", "snp-fremd", snapshotState("ws-2"))
		vorher := w.st.stand()
		w.p.leeren()
		_, err := w.v.Restore(context.Background(), "ws-1", id)
		pruefeLiteral(t, err, "ODR-E7002", 404, "Snapshot unknown")
		if w.p.anzahl("tx Entities.ReplaceAll") != 0 || w.p.anzahl("tx Workspaces.Advance") != 0 || !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%q: Teilzustand", id)
		}
		if w.p.anzahl("Snapshots.Get") != 0 || w.p.anzahl("Blobs.Open") != 0 {
			t.Errorf("%q: Snapshots.Get (S20 liegt allein hinter ErrFalscheArt) oder Open gerufen", id)
		}
	}
}

// TestRestoreFremderWorkspace (T35): der Snapshot eines anderen Workspace desselben Service und der eines Workspace
// eines anderen Service sind ODR-E7002 (404), nie ODR-E5004.
func TestRestoreFremderWorkspace(t *testing.T) {
	for _, svc := range []string{"svc", "svc2"} {
		w := restoreWelt(t)
		w.vorbelegen("ws-2", 0)
		z := w.st.stand()
		wz := z.ws["ws-2"]
		wz.ServiceID = svc
		z.ws["ws-2"] = wz
		w.st.setze(z)
		w.snapshotAblegen("ws-2", "snp-fremd", snapshotState("ws-2"))
		s := w.snaps.wsSnaps[[2]string{"ws-2", "snp-fremd"}]
		s.ServiceID = svc
		w.snaps.wsSnaps[[2]string{"ws-2", "snp-fremd"}] = s
		_, err := w.v.Restore(context.Background(), "ws-1", "snp-fremd")
		pruefeLiteral(t, err, "ODR-E7002", 404, "Snapshot unknown")
		if w.p.anzahl("Snapshots.Get") != 0 {
			t.Errorf("Dienst %s: Snapshots.Get gerufen", svc)
		}
	}
}

// TestRestoreBaseline (T36): die Kennung einer Baseline desselben Service, auch die des eigenen Workspace, ist
// ODR-E7008 (422); Snapshots.Get läuft einmal, nichts ändert sich.
func TestRestoreBaseline(t *testing.T) {
	for _, id := range []string{"bsl-1", "bsl-2"} {
		w := restoreWelt(t)
		w.baseline("bsl-2", "svc", standardEnts())
		vorher := w.st.stand()
		w.p.leeren()
		_, err := w.v.Restore(context.Background(), "ws-1", id)
		pruefeLiteral(t, err, "ODR-E7008", 422, "Snapshot kind does not fit the operation")
		if w.p.anzahl("Snapshots.Get") != 1 || w.p.anzahl("tx Entities.ReplaceAll") != 0 || !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%s: Get %d, Zustand", id, w.p.anzahl("Snapshots.Get"))
		}
	}
}

// TestRestoreFremderService (T37): die Baseline eines anderen Service ist ODR-E5004 (409) ohne Ursache; nichts ändert
// sich, Cursors werden nicht berührt.
func TestRestoreFremderService(t *testing.T) {
	w := restoreWelt(t)
	w.baseline("bsl-x", "svc2", standardEnts())
	vorher := w.st.stand()
	w.p.leeren()
	_, err := w.v.Restore(context.Background(), "ws-1", "bsl-x")
	pruefeLiteral(t, err, "ODR-E5004", 409, "Assignment belongs to another service")
	f, _ := codeVon(err)
	if f.Ursache != nil || errors.Is(err, outbound.ErrFalscheArt) || f.Code.Zielbereich() != model.ZielReplay {
		t.Errorf("Ursache %v, Ziel %v", f.Ursache, f.Code.Zielbereich())
	}
	if w.cursorAufrufe() != 0 || w.p.anzahl(lblTx) != 0 || !zustaendeGleich(vorher, w.st.stand()) {
		t.Error("Zustand oder Cursors berührt")
	}
}

// TestRestoreRang (T38): die Rangfolge der Schritte (Stufe 9); ein Fehler einer Stufe lässt die späteren ungerufen. Die
// Paare Z5 bis Z11 stehen darunter.
func TestRestoreRang(t *testing.T) {
	type fall struct {
		name  string
		pre   func(w *welt)
		ws    string
		id    string
		code  string
		letzt string // letzter Aufruf
	}
	hashFalsch := func(w *welt) {
		s := w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}]
		s.StateHash = "sha256:" + strings.Repeat("e", 64)
		w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}] = s
	}
	kaputt := func(w *welt) { w.snapshotBytesAblegen("ws-1", "snp-1", []byte("kein json")) }
	faelle := []fall{
		{"Stufe 2 Workspace unbekannt", nil, "nope", "snp-1", "ODR-E7001", "Workspaces.Get"},
		{"Stufe 3 Snapshot unbekannt", nil, "ws-1", "nope", "ODR-E7002", "Snapshots.GetWorkspace"},
		{"Stufe 4 Baseline eines anderen Service", func(w *welt) { w.baseline("bsl-x", "svc2", standardEnts()) }, "ws-1", "bsl-x", "ODR-E5004", "Snapshots.Get"},
		{"Stufe 4 Baseline desselben Service", nil, "ws-1", "bsl-1", "ODR-E7008", "Snapshots.Get"},
		{"Stufe 5 Hash ungleich Blob", hashFalsch, "ws-1", "snp-1", "ODR-E3001", "Snapshots.GetWorkspace"},
		{"Stufe 6 Blob fehlt", func(w *welt) { clear(w.blobs.inhalt) }, "ws-1", "snp-1", "ODR-E3001", "Blobs.Open"},
		{"Stufe 7 Manifest", kaputt, "ws-1", "snp-1", "ODR-E3001", "Strom.Close"},
		{"Z7 Workspace unbekannt und Baseline", nil, "nope", "bsl-1", "ODR-E7001", "Workspaces.Get"},
		{"Z9 Hash ungleich Blob und Manifest-Verstoß", func(w *welt) { kaputt(w); hashFalsch(w) }, "ws-1", "snp-1", "ODR-E3001", "Snapshots.GetWorkspace"},
	}
	for _, f := range faelle {
		w := restoreWelt(t)
		if f.pre != nil {
			f.pre(w)
		}
		w.p.leeren()
		_, err := w.v.Restore(context.Background(), f.ws, f.id)
		if c, ok := codeVon(err); !ok || c.Code.String() != f.code {
			t.Errorf("%s: %v, erwartet %s", f.name, err, f.code)
		}
		ops := w.p.ops()
		if ops[len(ops)-1] != f.letzt || w.p.anzahl(lblTx) != 0 {
			t.Errorf("%s: Folge %v, erwartet Ende bei %s", f.name, ops, f.letzt)
		}
	}
	// Z10: Manifest-Verstoß und Workspace inzwischen gelöscht: der Verstoß gewinnt.
	w := restoreWelt(t)
	w.snapshotBytesAblegen("ws-1", "snp-1", []byte("kein json"))
	w.p.nach = func(op string) {
		if op == "Blobs.Open" {
			w.p.nach = nil
			_ = wsRepo{w.st, false}.Delete(context.Background(), "ws-1")
		}
	}
	_, err := restoreLauf(w)
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
	// Z11: Workspace in der Transaktion gelöscht und Fehler von ReplaceAll: ODR-E7001, ReplaceAll 0.
	w = restoreWelt(t)
	w.p.setzeFehler("tx Entities.ReplaceAll", errPlatte)
	w.p.nach = func(op string) {
		if op == "Blobs.Open" {
			w.p.nach = nil
			_ = wsRepo{w.st, false}.Delete(context.Background(), "ws-1")
		}
	}
	_, err = restoreLauf(w)
	pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
	if got := len(w.p.von("tx Entities.ReplaceAll")); got != 0 {
		t.Errorf("Z11: ReplaceAll %d mal", got)
	}
	// Stufe 8 und 9: Fehler in der Transaktion lassen die späteren Schritte ungerufen.
	for _, st := range []struct{ op, spaeter string }{{"tx Workspaces.Get", "tx Entities.ReplaceAll"}, {"tx Entities.ReplaceAll", "tx Workspaces.Advance"}} {
		w := restoreWelt(t)
		w.p.setzeFehler(st.op, errPlatte)
		_, err := restoreLauf(w)
		pruefeCode(t, err, model.CodeE9001, errPlatte)
		if w.p.anzahl(st.spaeter) != 0 {
			t.Errorf("%s: %s gerufen", st.op, st.spaeter)
		}
	}
}

// TestRestoreHashBlob (T39): StateHash ungleich StateBlob ist ODR-E3001 mit ErrBeschaedigt ohne Open; ErrBlobFehlt,
// ErrBlobBeschaedigt und ein Lesefehler sind ODR-E3001.
func TestRestoreHashBlob(t *testing.T) {
	w := restoreWelt(t)
	s := w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}]
	s.StateHash = "sha256:" + strings.Repeat("e", 64)
	w.snaps.wsSnaps[[2]string{"ws-1", "snp-1"}] = s
	_, err := restoreLauf(w)
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBeschaedigt)
	if w.p.anzahl("Blobs.Open") != 0 {
		t.Error("Open trotz Hash-Verstoß")
	}
	w = restoreWelt(t)
	clear(w.blobs.inhalt)
	_, err = restoreLauf(w)
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobFehlt)
	w = restoreWelt(t)
	w.p.setzeFehler("Blobs.Open", outbound.ErrBlobBeschaedigt)
	_, err = restoreLauf(w)
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobBeschaedigt)
	w = restoreWelt(t)
	w.blobs.fehlerBei, w.blobs.lesefehler = 10, outbound.ErrBlobBeschaedigt
	_, err = restoreLauf(w)
	pruefeCode(t, err, model.CodeE3001, outbound.ErrBlobBeschaedigt)
}

// TestRestoreStrom (T40): der Strom wird auf jedem Weg nach gelungenem Open genau einmal geschlossen (Erfolg,
// Lesefehler, Manifest-Verstoß, Kontextfehler, Panik in f), auch wenn Close selbst scheitert; der Fehler von Close ändert
// das Ergebnis nicht.
func TestRestoreStrom(t *testing.T) {
	wege := map[string]func(w *welt){
		"Erfolg":     func(w *welt) {},
		"Lesefehler": func(w *welt) { w.blobs.fehlerBei, w.blobs.lesefehler = 10, outbound.ErrBlobBeschaedigt },
		"Manifest":   func(w *welt) { w.snapshotBytesAblegen("ws-1", "snp-1", []byte("kein json")) },
		"Kontext":    func(w *welt) { w.p.setzeFehler("Strom.Read", context.Canceled) },
		"Panik":      func(w *welt) { w.p.setzePanik("tx Entities.ReplaceAll", "peng") },
	}
	n := 0
	for name, weg := range wege {
		var ohne error
		for _, cf := range []error{nil, errPlatte} {
			w := restoreWelt(t)
			weg(w)
			w.blobs.closeFehler = cf
			var err error
			func() {
				defer func() { _ = recover() }()
				_, err = restoreLauf(w)
			}()
			if st := w.blobs.strome(); len(st) != 1 || st[0].geschlossen() != 1 {
				t.Errorf("%s Close-Fehler %v: Ströme %d", name, cf, len(st))
			}
			if cf == nil {
				ohne = err
			} else if (err == nil) != (ohne == nil) || (err != nil && err.Error() != ohne.Error()) {
				t.Errorf("%s: Fehler von Close ändert das Ergebnis: %v gegen %v", name, err, ohne)
			}
			n++
		}
	}
	if n != 10 {
		t.Errorf("%d Elemente", n)
	}
}

// TestRestoreMaske (T47): eine Entity mit Property-Maske (beobachtet mit Wert, beobachtet mit null, nicht beobachtet bei
// partial, Tombstone) bleibt an jeder Position byte-gleich; nichts wird null (M16: 4 × 3 = 12).
func TestRestoreMaske(t *testing.T) {
	zustaende := []model.EntityState{
		{Payload: `{"Id":"1","Name":"Milk"}`, KnownProperties: []string{"Id", "Name"}, Completeness: model.EntityComplete, Version: 3},
		{Payload: `{"Id":"1","Name":null}`, KnownProperties: []string{"Id", "Name"}, Completeness: model.EntityComplete, Version: 4},
		{Payload: `{"Id":"1"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityPartial, Version: 5},
		{Payload: `{"Id":"1"}`, KnownProperties: []string{"Id"}, Completeness: model.EntityPartial, Version: 6, Deleted: true},
	}
	n := 0
	for zi, z := range zustaende {
		for pos := 0; pos < 3; pos++ {
			soll := fuenfEintraege("ws-1")[:3]
			z.WorkspaceID, z.EntitySet, z.EntityKey = "ws-1", "Masken", "Id=1"
			soll = slices.Insert(soll, pos, z)
			w := neu(t)
			w.vorbelegen("ws-1", 0)
			w.snapshotAblegen("ws-1", "snp-1", soll)
			anders := kopiereEntities(soll)
			anders[pos].Payload, anders[pos].KnownProperties, anders[pos].Completeness, anders[pos].Deleted = `{"Id":"1","Name":"X"}`, []string{"Id", "Name"}, model.EntityComplete, !z.Deleted
			st := w.st.stand()
			st.ent["ws-1"] = anders
			w.st.setze(st)
			if _, err := restoreLauf(w); err != nil {
				t.Fatal(err)
			}
			got := w.st.stand().ent["ws-1"][pos]
			if got.Payload != z.Payload || !slices.Equal(got.KnownProperties, z.KnownProperties) || got.Completeness != z.Completeness || got.Deleted != z.Deleted || got.Version != z.Version {
				t.Errorf("Zustand %d Position %d: %+v", zi, pos, got)
			}
			n++
		}
	}
	if n != 12 {
		t.Errorf("%d Elemente", n)
	}
}

// TestRestoreAtomar (T48): scheitert ein Schritt von f (Fehler, Kontext, Panik), rollt die Transaktion zurück: State,
// Revision und Cursors wie vor dem Aufruf (3 Fehlerstellen × 3 Fehler = 9).
func TestRestoreAtomar(t *testing.T) {
	n := 0
	for _, op := range []string{"tx Workspaces.Get", "tx Entities.ReplaceAll", "tx Workspaces.Advance"} {
		for _, art := range []string{"roh", "kontext", "panik"} {
			w := restoreWelt(t)
			vorher := w.st.stand()
			switch art {
			case "roh":
				w.p.setzeFehler(op, errPlatte)
			case "kontext":
				w.p.setzeFehler(op, context.Canceled)
			default:
				w.p.setzePanik(op, "peng")
			}
			var err error
			func() {
				defer func() { _ = recover() }()
				_, err = restoreLauf(w)
			}()
			if art != "panik" && err == nil {
				t.Errorf("%s %s: kein Fehler", op, art)
			}
			if !zustaendeGleich(vorher, w.st.stand()) {
				t.Errorf("%s %s: Zustand geändert", op, art)
			}
			if w.p.anzahl(lblTx) != 1 {
				t.Errorf("%s %s: %d Transaktionen", op, art, w.p.anzahl(lblTx))
			}
			n++
		}
	}
	if n != 9 {
		t.Errorf("%d Elemente", n)
	}
}

// m7Oracle ist die Tabelle von SPEC-109 Nr. 5 von Hand: der Code für den Fehler fe am Gegenstand g.
func m7Oracle(fe string, g outbound.Gegenstand) string {
	switch fe {
	case "ErrNichtGefunden":
		return map[outbound.Gegenstand]string{outbound.GegenstandWorkspace: "ODR-E7001", outbound.GegenstandSnapshot: "ODR-E7002"}[g]
	case "ErrBereitsVorhanden":
		return map[outbound.Gegenstand]string{outbound.GegenstandWorkspace: "ODR-E7007"}[g]
	case "ErrFremderService":
		return "ODR-E5004"
	case "ErrFalscheArt":
		return "ODR-E7008"
	case "ErrBlobFehlt", "ErrBlobBeschaedigt", "ErrBeschaedigt":
		return "ODR-E3001"
	}
	return ""
}

// TestRestorePortFehler (T49): M7, Stelle 9 × Fehler 8 × Form 2 = 144, Orakel die Tabelle von Hand mit dem Gegenstand
// je Stelle; ErrFalscheArt an S19 führt zu S20 (vier eigene Fälle am Ende).
func TestRestorePortFehler(t *testing.T) {
	fehler := []tabellenFehler{{"ErrNichtGefunden", outbound.ErrNichtGefunden}, {"ErrBereitsVorhanden", outbound.ErrBereitsVorhanden},
		{"ErrFremderService", outbound.ErrFremderService}, {"ErrFalscheArt", outbound.ErrFalscheArt}, {"ErrBlobFehlt", outbound.ErrBlobFehlt},
		{"ErrBeschaedigt", outbound.ErrBeschaedigt}, {"ErrUngueltig", model.ErrUngueltig}, {"roh", errPlatte}}
	n := 0
	for _, s := range restoreStellen() {
		for _, fe := range fehler {
			for form := 0; form < 2; form++ {
				n++
				if s.nr == "S19" && fe.name == "ErrFalscheArt" {
					continue
				}
				w, o := s.vorbereitet(t)
				err := error(fe.err)
				if form == 1 {
					err = fmt.Errorf("verpackt: %w", fe.err)
				}
				w.p.setzeFehler(s.label, err)
				got := o.lauf(w, context.Background())
				soll := m7Oracle(fe.name, s.g)
				if soll == "" {
					soll = "ODR-E9001"
				}
				if c, ok := codeVon(got); !ok || c.Code.String() != soll || c.Status != c.Code.Stati()[0] || !errors.Is(got, fe.err) {
					t.Errorf("%v %s Form %d: %v, erwartet %s", s, fe.name, form, got, soll)
				}
			}
		}
	}
	if n != 144 {
		t.Errorf("M7: %d Elemente", n)
	}
	// ErrFalscheArt an S19 führt zu S20: eigene Baseline E7008, fremde E5004, S20 unbekannt E7002, S20 roh E9001.
	faelle := []struct {
		pre  func(w *welt)
		s20  error
		id   string
		code string
	}{
		{nil, nil, "bsl-1", "ODR-E7008"},
		{func(w *welt) { w.baseline("bsl-x", "svc2", standardEnts()) }, nil, "bsl-x", "ODR-E5004"},
		{nil, outbound.ErrNichtGefunden, "bsl-1", "ODR-E7002"},
		{nil, errPlatte, "bsl-1", "ODR-E9001"},
	}
	for i, f := range faelle {
		w := restoreWelt(t)
		if f.pre != nil {
			f.pre(w)
		}
		if f.s20 != nil {
			w.p.setzeFehler("Snapshots.Get", f.s20)
		}
		_, err := w.v.Restore(context.Background(), "ws-1", f.id)
		if c, ok := codeVon(err); !ok || c.Code.String() != f.code {
			t.Errorf("S20 Fall %d: %v, erwartet %s", i, err, f.code)
		}
	}
}

// TestRestoreKontext (T50): ein Kontextfehler in vier Formen an jeder der neun Stellen kehrt unverändert und ohne Code
// zurück, auch neben ErrNichtGefunden (der Code gewinnt nicht), und danach wird kein Port mehr gerufen; das ctx des
// Aufrufs erreicht jeden Port per Identität, auch S20 (M8: 4 × 9 = 36).
func TestRestoreKontext(t *testing.T) {
	formen := map[string]func(error) error{
		"nackt":       func(e error) error { return e },
		"%w":          func(e error) error { return fmt.Errorf("x: %w", e) },
		"Join vorn":   func(e error) error { return errors.Join(e, outbound.ErrNichtGefunden) },
		"Join hinten": func(e error) error { return errors.Join(outbound.ErrNichtGefunden, e) },
	}
	n := 0
	for _, s := range restoreStellen() {
		for name, f := range formen {
			w, o := s.vorbereitet(t)
			inj := f(context.DeadlineExceeded)
			w.p.setzeFehler(s.label, inj)
			w.p.leeren()
			err := o.lauf(w, context.Background())
			if err != inj {
				t.Errorf("%v %s: %v", s, name, err)
			}
			if _, ok := codeVon(err); ok {
				t.Errorf("%v %s: Code", s, name)
			}
			ops := slices.DeleteFunc(w.p.ops(), func(op string) bool { return op == "Strom.Close" || op == "Strom.Read" && s.nr != "S5" })
			if ops[len(ops)-1] != s.label {
				t.Errorf("%v %s: letzter Aufruf %s", s, name, ops[len(ops)-1])
			}
			n++
		}
	}
	if n != 36 {
		t.Errorf("M8: %d Elemente", n)
	}
	w := restoreWelt(t)
	ctx := neuerKontext(context.Background())
	if _, err := w.v.Restore(ctx, "ws-1", "bsl-1"); err == nil {
		t.Fatal("Baseline als Erfolg")
	}
	for _, op := range []string{"Workspaces.Get", "Snapshots.GetWorkspace", "Snapshots.Get"} {
		ev := w.p.von(op)
		if len(ev) != 1 || ev[0].ctx != context.Context(ctx) {
			t.Errorf("%s: ctx nicht per Identität weitergereicht", op)
		}
	}
}

// TestRestorePanik (T51): eine Panik in einem Port oder in f kommt beim Aufrufer an; der Dienst fängt sie nicht, und der
// Zustand bleibt unverändert.
func TestRestorePanik(t *testing.T) {
	ops := []string{"Workspaces.Get", "Snapshots.GetWorkspace", "Blobs.Open", "tx Entities.ReplaceAll", "tx Workspaces.Advance"}
	for _, op := range ops {
		w := restoreWelt(t)
		vorher := w.st.stand()
		w.p.setzePanik(op, "Panik "+op)
		var r any
		func() {
			defer func() { r = recover() }()
			_, _ = restoreLauf(w)
		}()
		if r != "Panik "+op || !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("%s: recover %v", op, r)
		}
	}
}

// TestRestoreKeinAudit (T52): Restore schreibt kein Audit und kein Protokoll, bei Erfolg und bei den neun Fehlern der
// Stellen.
func TestRestoreKeinAudit(t *testing.T) {
	var puffer bytes.Buffer
	n := 0
	lauf := func(w *welt, f func(*workspace.Verwaltung)) {
		k := w.konfig
		k.Log = slog.New(slog.NewTextHandler(&puffer, nil))
		v, err := workspace.New(k)
		if err != nil {
			t.Fatal(err)
		}
		f(v)
		n++
	}
	w := restoreWelt(t)
	lauf(w, func(v *workspace.Verwaltung) {
		if _, err := v.Restore(context.Background(), "ws-1", "snp-1"); err != nil {
			t.Fatal(err)
		}
	})
	if w.audit.Aufrufe() != 0 {
		t.Error("Audit bei Erfolg")
	}
	for _, s := range restoreStellen() {
		w, _ := s.vorbereitet(t)
		w.p.setzeFehler(s.label, errPlatte)
		id := "snp-1"
		if s.nr == "S20" {
			id = "bsl-1"
		}
		lauf(w, func(v *workspace.Verwaltung) { _, _ = v.Restore(context.Background(), "ws-1", id) })
		if w.audit.Aufrufe() != 0 {
			t.Errorf("%v: Audit", s)
		}
	}
	if puffer.Len() != 0 || n != 10 {
		t.Errorf("Protokoll %q, Läufe %d", puffer.String(), n)
	}
}

// TestRestoreIsolation (T53): ein Restore des einen Workspace berührt den anderen nie; jedes Argument jedes Ports nennt
// den Workspace des Aufrufs (zwei Workspaces, Schleife über alle Einträge des Protokolls).
func TestRestoreIsolation(t *testing.T) {
	for _, ziel := range []string{"ws-1", "ws-2"} {
		w := restoreWelt(t)
		w.vorbelegen("ws-2", 3)
		w.veraendern("ws-2")
		w.snapshotAblegen("ws-2", "snp-2", snapshotState("ws-2"))
		andere := "ws-1"
		if ziel == "ws-1" {
			andere = "ws-2"
		}
		vorher := w.st.stand()
		w.p.leeren()
		id := map[string]string{"ws-1": "snp-1", "ws-2": "snp-2"}[ziel]
		if _, err := w.v.Restore(context.Background(), ziel, id); err != nil {
			t.Fatal(err)
		}
		geprueft := 0
		for _, e := range w.p.ev {
			for _, pre := range []string{"Workspaces.", "Entities.", "Snapshots.GetWorkspace"} {
				if strings.Contains(e.op, pre) {
					if e.args[0] != ziel {
						t.Errorf("%s: Argument %v, erwartet %s", e.op, e.args[0], ziel)
					}
					geprueft++
				}
			}
		}
		if geprueft != 5 {
			t.Errorf("%s: %d geprüfte Aufrufe", ziel, geprueft)
		}
		nachher := w.st.stand()
		if !equalState(nachher.ent[andere], vorher.ent[andere]) || nachher.ws[andere] != vorher.ws[andere] {
			t.Errorf("%s berührt %s", ziel, andere)
		}
	}
}
