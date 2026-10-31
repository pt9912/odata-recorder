package workspace_test

import (
	"context"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// resetWelt legt den Workspace ws-1 (Revision rev) an; ist veraendert wahr, trägt er Mutationen, Tombstones,
// zwei Cursors und einen Workspace-Snapshot. Der Workspace ws-2 aus derselben Baseline ändert sich nie.
func resetWelt(t *testing.T, rev int64, veraendert bool) *welt {
	w := neu(t)
	w.vorbelegen("ws-1", rev)
	w.vorbelegen("ws-2", 2)
	z := w.st.stand()
	z.cur["ws-2"] = []model.ReplayCursor{{WorkspaceID: "ws-2", ReplaySessionID: "rep-9", RecordingSessionID: "ses-1", ServiceID: "svc",
		Strategy: model.StrategyFirst, Position: 1, Consumed: []int64{3}}}
	z.snaps["ws-2"] = []string{"snp-ws-9"}
	w.st.setze(z)
	if veraendert {
		w.veraendern("ws-1")
	}
	w.p.leeren()
	return w
}

// TestResetLiestZuerstDenWorkspace (Z39): Reset liest zuerst den Workspace (id unverändert, auch leer und
// außerhalb der Form), dann Snapshots.Get mit der Baseline des Workspace; ein unbekannter Workspace ist ODR-E7001
// ohne weiteren Port.
func TestResetLiestZuerstDenWorkspace(t *testing.T) {
	for _, id := range idFormen() {
		w := neu(t)
		w.baseline("bsl-2", "svc", genEnts(3, 2))
		z := w.st.stand()
		z.ws[id] = model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-2", CreatedAt: t0()}
		w.st.setze(z)
		w.p.leeren()
		if _, err := w.v.Reset(context.Background(), id); err != nil {
			t.Errorf("id %q: %v", id, err)
			continue
		}
		ops := w.p.ops()
		if ops[0] != "Workspaces.Get" || slices.Index(ops, "Snapshots.Get") < 1 {
			t.Errorf("id %q: Folge %v", id, ops)
		}
		if a := w.p.von("Workspaces.Get")[0].args[0]; a != id {
			t.Errorf("id %q: Get mit %q", id, a)
		}
		if a := w.p.von("Snapshots.Get")[0].args[0]; a != "bsl-2" {
			t.Errorf("id %q: Snapshots.Get mit %q, erwartet die Baseline des Workspace", id, a)
		}
		if got := w.st.stand().ent[id]; !equalState(got, erwarteterState(id, genEnts(3, 2))) {
			t.Errorf("id %q: State nicht der der Baseline des Workspace", id)
		}
		w.p.leeren()
		_, err := w.v.Reset(context.Background(), id+"x")
		pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
		if !slices.Equal(w.p.ops(), []string{"Workspaces.Get"}) {
			t.Errorf("id %q: unbekannt, Folge %v", id, w.p.ops())
		}
	}
}

func equalState(a, b []model.EntityState) bool {
	return len(a) == len(b) && func() bool {
		for i := range a {
			x, y := a[i], b[i]
			if x.WorkspaceID != y.WorkspaceID || x.EntitySet != y.EntitySet || x.EntityKey != y.EntityKey || x.Payload != y.Payload ||
				!slices.Equal(x.KnownProperties, y.KnownProperties) || x.Completeness != y.Completeness || x.Version != y.Version || x.Deleted != y.Deleted {
				return false
			}
		}
		return true
	}()
}

// TestTransaktionDesReset (Z40): genau eine Transaktion; darin Get, ReplaceAll, ResetAll, Advance(revision aus
// Get) je genau einmal in dieser Folge, sonst keine Operation; die Revision ist die der Transaktion, auch wenn
// ein Schreiber zwischen dem Lesen des Workspace und der Transaktion die Revision ändert.
func TestTransaktionDesReset(t *testing.T) {
	for _, rev := range []int64{0, 1, 7} {
		w := resetWelt(t, rev, true)
		if _, err := w.v.Reset(context.Background(), "ws-1"); err != nil {
			t.Fatal(err)
		}
		var folge []string
		for _, op := range w.p.ops() {
			if len(op) > 3 && op[:3] == "tx " {
				folge = append(folge, op)
			}
		}
		want := []string{"tx Workspaces.Get", "tx Entities.ReplaceAll", "tx Cursors.ResetAll", "tx Workspaces.Advance"}
		if !slices.Equal(folge, want) || w.p.anzahl(lblTx) != 1 {
			t.Errorf("Revision %d: Folge %v, Transaktionen %d", rev, folge, w.p.anzahl(lblTx))
		}
		if a := w.p.von("tx Workspaces.Advance")[0].args; a[0] != "ws-1" || a[1] != rev {
			t.Errorf("Revision %d: Advance mit %v", rev, a)
		}
		for _, op := range []string{"Entities.Put", "tx Entities.Put", "tx Workspaces.Create", "tx Workspaces.Delete", "Workspaces.Delete", "Workspaces.Advance",
			"tx Snapshots.Get", "tx Cursors.Put", "Cursors.Put"} {
			if w.p.anzahl(op) != 0 {
				t.Errorf("Revision %d: %s gerufen", rev, op)
			}
		}
	}
	// Ein Schreiber ändert die Revision zwischen dem Lesen des Workspace und der Transaktion.
	w := resetWelt(t, 2, false)
	w.p.nach = func(op string) {
		if op == "Workspaces.Get" {
			w.p.nach = nil
			w.st.eindringling(context.Background(), func(s outbound.WorkspaceStores) error {
				_, err := s.Workspaces.Advance(context.Background(), "ws-1", 2)
				return err
			})
		}
	}
	got, err := w.v.Reset(context.Background(), "ws-1")
	if err != nil || got.Revision != 4 {
		t.Errorf("Reset nach fremdem Advance: %+v, %v (erwartet Revision 4)", got, err)
	}
}

// TestFehlerInDerTransaktionDesReset (Z41): ein Fehler an jedem der vier Schritte, in der Menge der Tabelle
// und als Kontextfehler, und ein Abbruch nach jedem Schritt lassen State, Cursors und Revision wie vorher.
func TestFehlerInDerTransaktionDesReset(t *testing.T) {
	schritte := []string{"tx Workspaces.Get", "tx Entities.ReplaceAll", "tx Cursors.ResetAll", "tx Workspaces.Advance"}
	faelle := 0
	for _, label := range schritte {
		var fehler []error
		for _, fe := range tabellenFehlerListe() {
			fehler = append(fehler, fe.err)
		}
		for _, fe := range kontextFehler() {
			fehler = append(fehler, fe)
		}
		for _, fe := range fehler {
			w := resetWelt(t, 3, true)
			vorher := w.st.stand()
			w.p.setzeFehler(label, fe)
			_, err := w.v.Reset(context.Background(), "ws-1")
			if err == nil {
				t.Errorf("%s/%v: kein Fehler", label, fe)
			}
			if !zustaendeGleich(vorher, w.st.stand()) {
				t.Errorf("%s/%v: Zustand geändert", label, fe)
			}
			faelle++
		}
		w := resetWelt(t, 3, true)
		vorher := w.st.stand()
		ctx, cancel := context.WithCancel(context.Background())
		w.p.nach = func(op string) {
			if op == label {
				cancel()
			}
		}
		_, err := w.v.Reset(ctx, "ws-1")
		pruefeOhneCode(t, err, context.Canceled)
		if !zustaendeGleich(vorher, w.st.stand()) {
			t.Errorf("Abbruch nach %s: Zustand geändert", label)
		}
		faelle++
	}
	if faelle != 4*(24+4+1) {
		t.Errorf("Fälle %d", faelle)
	}
}

// TestWirkungDesReset (Z42): State gleich dem der Baseline (byte-gleich, Version 0, nicht gelöscht), Mutationen und
// Tombstones weg, jeder Cursor auf 0 mit leerem consumed und unveränderter Strategie, Workspace-Snapshots bleiben,
// ein anderer Workspace aus derselben Baseline bleibt, die Revision wächst um 1.
func TestWirkungDesReset(t *testing.T) {
	for _, rev := range []int64{0, 1, 7} {
		for _, veraendert := range []bool{false, true} {
			w := resetWelt(t, rev, veraendert)
			vorher := w.st.stand()
			got, err := w.v.Reset(context.Background(), "ws-1")
			if err != nil || got.Revision != rev+1 {
				t.Fatalf("Revision %d: %+v, %v", rev, got, err)
			}
			nachher := w.st.stand()
			if !equalState(nachher.ent["ws-1"], erwarteterState("ws-1", standardEnts())) {
				t.Errorf("Revision %d veraendert=%v: State %+v", rev, veraendert, nachher.ent["ws-1"])
			}
			if len(nachher.cur["ws-1"]) != len(vorher.cur["ws-1"]) {
				t.Errorf("Cursors %d statt %d", len(nachher.cur["ws-1"]), len(vorher.cur["ws-1"]))
			}
			for i, c := range nachher.cur["ws-1"] {
				soll := vorher.cur["ws-1"][i]
				soll.Position, soll.Consumed = 0, nil
				if c.Position != 0 || len(c.Consumed) != 0 || c.Strategy != soll.Strategy || c.ReplaySessionID != soll.ReplaySessionID ||
					c.RecordingSessionID != soll.RecordingSessionID || c.ServiceID != soll.ServiceID {
					t.Errorf("Cursor %d: %+v", i, c)
				}
			}
			if !slices.Equal(nachher.snaps["ws-1"], vorher.snaps["ws-1"]) {
				t.Error("Workspace-Snapshots geändert")
			}
			// ws-2 und alles, was nicht zu ws-1 gehört, bleibt.
			if !zustaendeGleich(nurWorkspace(vorher, "ws-2"), nurWorkspace(nachher, "ws-2")) {
				t.Error("ws-2 geändert")
			}
		}
	}
}

// nurWorkspace ist der Teil des Zustands, der zum Workspace id gehört.
func nurWorkspace(z zustand, id string) zustand {
	r := leererZustand()
	if w, da := z.ws[id]; da {
		r.ws[id] = w
	}
	r.ent[id], r.cur[id], r.snaps[id] = z.ent[id], z.cur[id], z.snaps[id]
	return r.kopie()
}

// TestZweiResetHintereinander (Z43): State und Cursors bleiben gleich, die Revision wächst je um 1, auch ohne
// Änderung.
func TestZweiResetHintereinander(t *testing.T) {
	for _, veraendert := range []bool{false, true} {
		w := resetWelt(t, 5, veraendert)
		a, err := w.v.Reset(context.Background(), "ws-1")
		if err != nil {
			t.Fatal(err)
		}
		z1 := w.st.stand()
		b, err := w.v.Reset(context.Background(), "ws-1")
		if err != nil {
			t.Fatal(err)
		}
		z2 := w.st.stand()
		if a.Revision != 6 || b.Revision != 7 || z2.ws["ws-1"].Revision != 7 {
			t.Errorf("Revisionen %d, %d, %d", a.Revision, b.Revision, z2.ws["ws-1"].Revision)
		}
		z1.ws["ws-1"], z2.ws["ws-1"] = model.Workspace{}, model.Workspace{}
		if !zustaendeGleich(z1, z2) {
			t.Error("der zweite Reset änderte State oder Cursors")
		}
	}
}

// TestErgebnisVonReset (Z44): das Ergebnis ist der Workspace aus Get der Transaktion mit der neuen Revision; ID,
// Service, Baseline und CreatedAt bleiben.
func TestErgebnisVonReset(t *testing.T) {
	w := resetWelt(t, 4, true)
	z := w.st.stand()
	ws := z.ws["ws-1"]
	ws.ServiceID, ws.BaselineSnapshotID, ws.CreatedAt = "svc", "bsl-1", t0().Add(99)
	z.ws["ws-1"] = ws
	w.st.setze(z)
	got, err := w.v.Reset(context.Background(), "ws-1")
	if err != nil {
		t.Fatal(err)
	}
	want := model.Workspace{ID: "ws-1", ServiceID: "svc", BaselineSnapshotID: "bsl-1", Revision: 5, CreatedAt: t0().Add(99)}
	if got != want {
		t.Errorf("Ergebnis %+v, erwartet %+v", got, want)
	}
	if st := w.st.stand().ws["ws-1"]; st != want {
		t.Errorf("gespeichert %+v", st)
	}
}
