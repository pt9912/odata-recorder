package outboundtest

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// RunTransaktionVertrag prüft WithWorkspaceTransaction, SPEC-108 Nr. 1, und die Abläufe, die
// Reset und Restore aus den Operationen zusammensetzen.
func RunTransaktionVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"nil bestätigt alles", transaktionFall1},
		{"ein Fehler rollt zurück und wird unverändert geliefert", transaktionFall2},
		{"eine Panik rollt zurück und wird weitergereicht", transaktionFall3},
		{"ein abgebrochener Kontext rollt zurück", transaktionFall4},
		{"Transaktionen laufen nacheinander", transaktionNebenlaeufig},
		{"Reset: ReplaceAll, ResetAll und Advance in einer Transaktion", transaktionResetFall},
		{"Restore: Entity-State ersetzen, Cursors bleiben", transaktionRestoreFall},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func transaktionFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	ok(t, v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		ok(t, s.Workspaces.Create(ctx, arbeitsraum("w", 0)))
		ok(t, s.Entities.Put(ctx, entity("w", "Products", "1")))
		ok(t, s.Cursors.Put(ctx, cursor("w", "r", "s1")))
		ok(t, s.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn", "d", "w", "n", 0)))
		// Die Repositories sehen die eigenen, noch nicht bestätigten Schreibvorgänge.
		if _, err := s.Entities.Get(ctx, "w", "Products", "1"); err != nil {
			t.Fatalf("Get in der Transaktion: %v", err)
		}
		return nil
	}))
	if _, err := v.Workspaces.Get(ctx, "w"); err != nil {
		t.Fatalf("Workspace: %v", err)
	}
	if l, err := v.Entities.List(ctx, "w"); err != nil || len(l) != 1 {
		t.Fatalf("Entities: %v %v", l, err)
	}
	if l, err := v.Cursors.List(ctx, "w"); err != nil || len(l) != 1 {
		t.Fatalf("Cursors: %v %v", l, err)
	}
	if _, err := v.Snapshots.GetWorkspace(ctx, "w", "sn"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
}

// schreibeAlles schreibt in jedes der vier Repositories.
func schreibeAlles(ctx context.Context, t *testing.T, s outbound.WorkspaceStores) {
	t.Helper()
	ok(t, s.Workspaces.Create(ctx, arbeitsraum("w", 0)))
	ok(t, s.Entities.Put(ctx, entity("w", "Products", "1")))
	ok(t, s.Cursors.Put(ctx, cursor("w", "r", "s1")))
	ok(t, s.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn", "d", "w", "n", 0)))
}

func nichtsVonSchreibeAlles(t *testing.T, v Verbindung) {
	t.Helper()
	for _, tabelle := range []string{"workspaces", "entities", "replay_cursors"} {
		if n := v.Zeilen(t, tabelle); n != 0 {
			t.Fatalf("%s: %d Zeilen nach dem Zurückrollen", tabelle, n)
		}
	}
	if n := v.Zeilen(t, "snapshots"); n != 2 {
		t.Fatalf("snapshots: %d Zeilen, erwartet die 2 Baselines", n)
	}
}

func transaktionFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	eigener := fmt.Errorf("eigener Fehler: %w", outbound.ErrReferenziert)
	err := v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		schreibeAlles(ctx, t, s)
		return eigener
	})
	if err != eigener {
		t.Fatalf("Fehler %v, erwartet unverändert %v", err, eigener)
	}
	nichtsVonSchreibeAlles(t, v)
	// Ein Fehler einer Operation in f, den f weiterreicht, rollt ebenfalls zurück.
	err = v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		ok(t, s.Workspaces.Create(ctx, arbeitsraum("w", 0)))
		return s.Entities.Put(ctx, entity("gibt-es-nicht", "P", "1"))
	})
	erwarte(t, err, outbound.ErrNichtGefunden)
	nichtsVonSchreibeAlles(t, v)
	// Nach dem Zurückrollen ist die nächste Transaktion nutzbar.
	ok(t, v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		schreibeAlles(ctx, t, s)
		return nil
	}))
}

func transaktionFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := wsBasis(t, neu)
	gepanikt := func() (wert any) {
		defer func() { wert = recover() }()
		_ = v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
			schreibeAlles(ctx, t, s)
			panic("kaputt")
		})
		return nil
	}()
	if gepanikt != "kaputt" {
		t.Fatalf("Panik %v, erwartet \"kaputt\" weitergereicht", gepanikt)
	}
	nichtsVonSchreibeAlles(t, v)
	ok(t, v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		schreibeAlles(ctx, t, s)
		return nil
	}))
}

func transaktionFall4(t *testing.T, neu Neu) {
	v := wsBasis(t, neu)
	aufgerufen := false
	err := v.Transaktion.WithWorkspaceTransaction(abgebrochen(), func(outbound.WorkspaceStores) error {
		aufgerufen = true
		return nil
	})
	erwarte(t, err, context.Canceled)
	if aufgerufen {
		t.Fatal("f lief trotz abgebrochenem Kontext")
	}
	// Bricht der Kontext in f ab und f liefert nil, ist das Ergebnis der Fehler des Kontexts.
	ctx, abbruch := context.WithCancel(context.Background())
	err = v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		ok(t, s.Workspaces.Create(context.Background(), arbeitsraum("w", 0)))
		abbruch()
		return nil
	})
	erwarte(t, err, context.Canceled)
	if n := v.Zeilen(t, "workspaces"); n != 0 {
		t.Fatalf("%d Workspaces nach abgebrochener Transaktion", n)
	}
}

func transaktionNebenlaeufig(t *testing.T, neu Neu) {
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
			// Lesen und Erhöhen in einer Transaktion: nur wenn sie nacheinander laufen, gelingen alle.
			fehler[i] = v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
				w, err := s.Workspaces.Get(ctx, "w")
				if err != nil {
					return err
				}
				_, err = s.Workspaces.Advance(ctx, "w", w.Revision)
				return err
			})
		}()
	}
	wg.Wait()
	for i, err := range fehler {
		if err != nil {
			t.Fatalf("Transaktion %d: %v", i, err)
		}
	}
	if got, _ := v.Workspaces.Get(ctx, "w"); got.Revision != n {
		t.Fatalf("Revision %d, erwartet %d", got.Revision, n)
	}
}

// zustand fasst Entity-State, Cursors und Revision eines Workspace zusammen.
type zustand struct {
	Entities []model.EntityState
	Cursors  []model.ReplayCursor
	Revision int64
}

func liesZustand(t *testing.T, v Verbindung) zustand {
	t.Helper()
	ctx := context.Background()
	e, err := v.Entities.List(ctx, "w")
	ok(t, err)
	c, err := v.Cursors.List(ctx, "w")
	ok(t, err)
	w, err := v.Workspaces.Get(ctx, "w")
	ok(t, err)
	return zustand{e, c, w.Revision}
}

// transaktionBasis liefert die Verbindung mit einem Workspace w auf Revision 1, einem
// Entity-State und zwei Cursors.
func transaktionBasis(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	ctx := context.Background()
	v := wsBasis(t, neu)
	legeArbeitsraeumeAn(t, v, "w")
	ok(t, v.Entities.Put(ctx, entity("w", "Products", "alt")))
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s1")))
	ok(t, v.Cursors.Put(ctx, cursor("w", "r", "s2")))
	_, err := v.Workspaces.Advance(ctx, "w", 0)
	ok(t, err)
	return v
}

func transaktionResetFall(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := transaktionBasis(t, neu)
	alt := liesZustand(t, v)
	reset := func(s outbound.WorkspaceStores, erwartet int64) error {
		if err := s.Entities.ReplaceAll(ctx, "w", []model.EntityState{entity("w", "Products", "basis")}); err != nil {
			return err
		}
		if err := s.Cursors.ResetAll(ctx, "w"); err != nil {
			return err
		}
		_, err := s.Workspaces.Advance(ctx, "w", erwartet)
		return err
	}
	err := v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error { return reset(s, 7) })
	erwarte(t, err, outbound.ErrRevisionVeraltet)
	if got := liesZustand(t, v); !reflect.DeepEqual(got, alt) {
		t.Fatalf("ein Fehler am Ende ließ den Zustand nicht unverändert:\n%+v\n%+v", got, alt)
	}
	ok(t, v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error { return reset(s, 1) }))
	got := liesZustand(t, v)
	if keys := entityKeys(got.Entities); !reflect.DeepEqual(keys, []string{"Products/basis"}) || got.Revision != 2 {
		t.Fatalf("State %v, Revision %d nach Reset", keys, got.Revision)
	}
	for _, c := range got.Cursors {
		if c.Position != 0 || len(c.Consumed) != 0 {
			t.Fatalf("Cursor nach Reset: %+v", c)
		}
	}
	if len(got.Cursors) != 2 {
		t.Fatalf("%d Cursors nach Reset", len(got.Cursors))
	}
}

func transaktionRestoreFall(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := transaktionBasis(t, neu)
	ok(t, v.Snapshots.CreateWorkspace(ctx, arbeitsraumSnapshot("sn", "d", "w", "n", 0)))
	cursorsVorher := liesZustand(t, v).Cursors
	restore := func(s outbound.WorkspaceStores, eintraege []model.EntityState) error {
		if _, err := s.Snapshots.GetWorkspace(ctx, "w", "sn"); err != nil {
			return err
		}
		if err := s.Entities.ReplaceAll(ctx, "w", eintraege); err != nil {
			return err
		}
		_, err := s.Workspaces.Advance(ctx, "w", 1)
		return err
	}
	doppelt := []model.EntityState{entity("w", "P", "1"), entity("w", "P", "1")}
	err := v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error { return restore(s, doppelt) })
	erwarte(t, err, model.ErrUngueltig)
	if got := liesZustand(t, v); got.Revision != 1 || !reflect.DeepEqual(entityKeys(got.Entities), []string{"Products/alt"}) {
		t.Fatalf("fehlgeschlagener Restore änderte den Zustand: %+v", got)
	}
	ok(t, v.Transaktion.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
		return restore(s, []model.EntityState{entity("w", "P", "1")})
	}))
	got := liesZustand(t, v)
	if !reflect.DeepEqual(entityKeys(got.Entities), []string{"P/1"}) || got.Revision != 2 {
		t.Fatalf("State %v, Revision %d nach Restore", entityKeys(got.Entities), got.Revision)
	}
	if !reflect.DeepEqual(got.Cursors, cursorsVorher) {
		t.Fatalf("Restore änderte die Cursors:\n%+v\n%+v", got.Cursors, cursorsVorher)
	}
}
