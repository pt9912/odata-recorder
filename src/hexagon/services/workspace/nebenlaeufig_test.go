package workspace_test

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// gleichzeitig startet n Funktionen über einen gemeinsamen Kanal zugleich und wartet auf alle.
func gleichzeitig(n int, f func(i int)) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			f(i)
		}()
	}
	close(start)
	wg.Wait()
}

// TestGleichzeitigesCreateGleicheID (Z46): zwei bis acht gleichzeitige Create mit derselben id und zwei
// Baselines verschiedenen Inhalts: genau eines gelingt, die übrigen sind ODR-E7007, ReplaceAll läuft einmal, jedes
// Create bleibt bei einer Transaktion, der State ist der der Baseline des Gewinners.
func TestGleichzeitigesCreateGleicheID(t *testing.T) {
	for n := 2; n <= 8; n++ {
		w := neu(t)
		a, b := genEnts(3, 1), genEnts(5, 2)
		w.baseline("bsl-a", "svc", a)
		w.baseline("bsl-b", "svc", b)
		w.p.leeren()
		fehler := make([]error, n)
		gleichzeitig(n, func(i int) {
			_, fehler[i] = w.v.Create(context.Background(), []string{"bsl-a", "bsl-b"}[i%2], "ws-neu")
		})
		gelungen := 0
		for _, err := range fehler {
			if err == nil {
				gelungen++
			} else {
				pruefeCode(t, err, model.CodeE7007, outbound.ErrBereitsVorhanden)
			}
		}
		z := w.st.stand()
		var soll []model.EntityState
		switch z.ws["ws-neu"].BaselineSnapshotID {
		case "bsl-a":
			soll = erwarteterState("ws-neu", a)
		case "bsl-b":
			soll = erwarteterState("ws-neu", b)
		default:
			t.Fatalf("n=%d: kein Workspace", n)
		}
		if gelungen != 1 || w.p.anzahl("tx Entities.ReplaceAll") != 1 || w.p.anzahl(lblTx) != n || !equalState(z.ent["ws-neu"], soll) {
			t.Errorf("n=%d: gelungen %d, ReplaceAll %d, Transaktionen %d, State des Gewinners %v",
				n, gelungen, w.p.anzahl("tx Entities.ReplaceAll"), w.p.anzahl(lblTx), equalState(z.ent["ws-neu"], soll))
		}
	}
}

// TestGleichzeitigesCreateGetrennterSpeicher (Z47): zwei gleichzeitige Create mit verschiedener id aus derselben
// Baseline gelingen; die Argumente von ReplaceAll teilen keinen Speicher.
func TestGleichzeitigesCreateGetrennterSpeicher(t *testing.T) {
	w := neu(t)
	gleichzeitig(2, func(i int) {
		if _, err := w.v.Create(context.Background(), "bsl-1", fmt.Sprint("ws-", i)); err != nil {
			t.Error(err)
		}
	})
	ev := w.p.von("tx Entities.ReplaceAll")
	if len(ev) != 2 {
		t.Fatalf("ReplaceAll %d mal", len(ev))
	}
	a, b := ev[0].args[1].([]model.EntityState), ev[1].args[1].([]model.EntityState)
	if len(a) != 3 || len(b) != 3 {
		t.Fatalf("Einträge %d und %d", len(a), len(b))
	}
	vorher := slices.Clone(b[0].KnownProperties)
	for i := range a {
		for j := range a[i].KnownProperties {
			a[i].KnownProperties[j] = "geaendert"
		}
	}
	if !slices.Equal(b[0].KnownProperties, vorher) {
		t.Error("die Einträge teilen Speicher")
	}
	if &a[0] == &b[0] {
		t.Error("dieselbe Liste")
	}
}

// TestGleichzeitigeReset (Z48): zwei gleichzeitige Reset gelingen, die Revisionen tragen n plus 1 und n plus 2,
// auch wenn beide den Workspace lesen, bevor eine Transaktion läuft; State und Cursors sind die der Baseline.
func TestGleichzeitigeReset(t *testing.T) {
	w := resetWelt(t, 4, true)
	var sperre sync.WaitGroup
	sperre.Add(2)
	w.p.nach = func(op string) {
		if op == "Workspaces.Get" {
			sperre.Done()
			sperre.Wait()
		}
	}
	revisionen := make([]int64, 2)
	gleichzeitig(2, func(i int) {
		ws, err := w.v.Reset(context.Background(), "ws-1")
		if err != nil {
			t.Errorf("Reset %d: %v", i, err)
		}
		revisionen[i] = ws.Revision
	})
	sort.Slice(revisionen, func(i, j int) bool { return revisionen[i] < revisionen[j] })
	if revisionen[0] != 5 || revisionen[1] != 6 {
		t.Errorf("Revisionen %v", revisionen)
	}
	z := w.st.stand()
	if !equalState(z.ent["ws-1"], erwarteterState("ws-1", standardEnts())) || z.ws["ws-1"].Revision != 6 {
		t.Errorf("State oder Revision: %+v", z.ws["ws-1"])
	}
	for _, c := range z.cur["ws-1"] {
		if c.Position != 0 || len(c.Consumed) != 0 {
			t.Errorf("Cursor %+v", c)
		}
	}
}

// loeschenAlsEindringling reiht ein Löschen des Workspace ws-1 in die Schlange der Transaktionen ein.
func loeschenAlsEindringling(w *welt) <-chan error {
	return w.st.eindringling(context.Background(), func(s outbound.WorkspaceStores) error { return s.Workspaces.Delete(context.Background(), "ws-1") })
}

// TestResetGegenLoeschen (Z49): Reset gegen ein Löschen in beiden Reihenfolgen, bestimmt durch Haltepunkt und
// FIFO: liegt das Löschen vor der Transaktion des Reset, ist das Ergebnis ODR-E7001, liegt es danach, gelingt
// der Reset; nie bleibt ein State oder Cursor ohne Workspace.
func TestResetGegenLoeschen(t *testing.T) {
	for _, vorher := range []bool{true, false} {
		w := resetWelt(t, 2, true)
		label := "tx Workspaces.Get"
		if vorher {
			label = "Workspaces.Get"
		}
		var ch <-chan error
		w.p.nach = func(op string) {
			if op == label && ch == nil {
				ch = loeschenAlsEindringling(w)
			}
		}
		_, err := w.v.Reset(context.Background(), "ws-1")
		if ch == nil {
			t.Fatalf("vorher=%v: der Haltepunkt %s wurde nicht erreicht", vorher, label)
		}
		if e := <-ch; e != nil {
			t.Errorf("vorher=%v: Löschen: %v", vorher, e)
		}
		if vorher {
			pruefeCode(t, err, model.CodeE7001, outbound.ErrNichtGefunden)
		} else if err != nil {
			t.Errorf("Reset vor dem Löschen: %v", err)
		}
		z := w.st.stand()
		if _, da := z.ws["ws-1"]; da || len(z.ent["ws-1"]) != 0 || len(z.cur["ws-1"]) != 0 || len(z.snaps["ws-1"]) != 0 {
			t.Errorf("vorher=%v: Rest ohne Workspace: %+v", vorher, z)
		}
		if _, da := z.ws["ws-2"]; !da || len(z.ent["ws-2"]) == 0 {
			t.Errorf("vorher=%v: ws-2 berührt", vorher)
		}
	}
}

// fortschreiben ist das Fortschreiben eines Cursors nach der Regel von SPEC-129 Nr. 11 a: Lesen, Wählen und
// Schreiben in einer Transaktion.
func fortschreiben(w *welt) <-chan error {
	return w.st.eindringling(context.Background(), func(s outbound.WorkspaceStores) error {
		c, err := s.Cursors.Get(context.Background(), "ws-1", "rep-1", "ses-1")
		if err != nil {
			return err
		}
		c.Consumed = append(slices.Clone(c.Consumed), int64(len(c.Consumed)+1))
		c.Position = int64(len(c.Consumed))
		return s.Cursors.Put(context.Background(), c)
	})
}

// TestResetGegenFortschreibenDesCursors (Z50): liegt das Fortschreiben vor dem Reset, steht der Cursor danach auf
// 0, liegt es danach, auf 1.
func TestResetGegenFortschreibenDesCursors(t *testing.T) {
	for _, vorher := range []bool{true, false} {
		w := resetWelt(t, 2, true)
		label := "tx Workspaces.Get"
		if vorher {
			label = "Workspaces.Get"
		}
		var ch <-chan error
		w.p.nach = func(op string) {
			if op == label && ch == nil {
				ch = fortschreiben(w)
			}
		}
		if _, err := w.v.Reset(context.Background(), "ws-1"); err != nil {
			t.Fatal(err)
		}
		if ch == nil {
			t.Fatalf("vorher=%v: Haltepunkt nicht erreicht", vorher)
		}
		if e := <-ch; e != nil {
			t.Fatalf("Fortschreiben: %v", e)
		}
		var c model.ReplayCursor
		for _, x := range w.st.stand().cur["ws-1"] {
			if x.ReplaySessionID == "rep-1" {
				c = x
			}
		}
		want := int64(1)
		if vorher {
			want = 0
		}
		if c.Position != want || int64(len(c.Consumed)) != want {
			t.Errorf("vorher=%v: Position %d, consumed %v, erwartet %d", vorher, c.Position, c.Consumed, want)
		}
	}
}

// beobachtung ist, was ein Eindringling sieht, wenn er an der Reihe ist.
type beobachtung struct {
	ent       []model.EntityState
	cursors   []model.ReplayCursor
	revision  int64
	workspace bool
}

// beobachter reiht eine Transaktion ein, die den Zustand von ws-1 festhält.
func beobachter(w *welt, aus *beobachtung) <-chan error {
	return w.st.eindringling(context.Background(), func(s outbound.WorkspaceStores) error {
		ctx := context.Background()
		ws, err := s.Workspaces.Get(ctx, "ws-1")
		aus.workspace = err == nil
		aus.revision = ws.Revision
		if aus.ent, err = s.Entities.List(ctx, "ws-1"); err != nil {
			return err
		}
		aus.cursors, err = s.Cursors.List(ctx, "ws-1")
		return err
	})
}

// TestFensterDerTransaktionen (Z51 bis Z54): zwischen den Schritten der Anlage (Create, ReplaceAll) und des Reset
// (Get, ReplaceAll, ResetAll, Advance) läuft kein anderer Schreiber: ein Eindringling, der der Schlange nach dem
// Schritt k beitritt, sieht den Zustand nach der ganzen Transaktion.
func TestFensterDerTransaktionen(t *testing.T) {
	t.Run("Z51 Anlage nach Create", func(t *testing.T) {
		w := neu(t)
		var obs beobachtung
		var ch <-chan error
		w.p.nach = func(op string) {
			if op == "tx Workspaces.Create" && ch == nil {
				ch = w.st.eindringling(context.Background(), func(s outbound.WorkspaceStores) error {
					ws, err := s.Workspaces.Get(context.Background(), "ws-neu")
					obs.workspace, obs.revision = err == nil, ws.Revision
					obs.ent, _ = s.Entities.List(context.Background(), "ws-neu")
					return nil
				})
			}
		}
		if _, err := w.v.Create(context.Background(), "bsl-1", "ws-neu"); err != nil {
			t.Fatal(err)
		}
		if ch == nil {
			t.Fatal("Haltepunkt nicht erreicht")
		}
		<-ch
		if !obs.workspace || !equalState(obs.ent, erwarteterState("ws-neu", standardEnts())) {
			t.Errorf("der Eindringling sah Workspace %v und %d Einträge", obs.workspace, len(obs.ent))
		}
	})
	for _, k := range []struct{ nr, label string }{
		{"Z52 Reset nach Get", "tx Workspaces.Get"}, {"Z53 Reset nach ReplaceAll", "tx Entities.ReplaceAll"},
		{"Z54 Reset nach ResetAll", "tx Cursors.ResetAll"},
	} {
		t.Run(k.nr, func(t *testing.T) {
			w := resetWelt(t, 3, true)
			var obs beobachtung
			var ch <-chan error
			w.p.nach = func(op string) {
				if op == k.label && ch == nil {
					ch = beobachter(w, &obs)
				}
			}
			if _, err := w.v.Reset(context.Background(), "ws-1"); err != nil {
				t.Fatal(err)
			}
			if ch == nil {
				t.Fatal("Haltepunkt nicht erreicht")
			}
			if err := <-ch; err != nil {
				t.Fatal(err)
			}
			if !equalState(obs.ent, erwarteterState("ws-1", standardEnts())) || obs.revision != 4 {
				t.Errorf("State gleich der Baseline: %v, Revision %d", equalState(obs.ent, erwarteterState("ws-1", standardEnts())), obs.revision)
			}
			for _, c := range obs.cursors {
				if c.Position != 0 || len(c.Consumed) != 0 {
					t.Errorf("Cursor %+v", c)
				}
			}
		})
	}
}
