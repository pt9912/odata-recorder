package replay_test

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// aufrufMitNullwerten ruft die Methode m von impl mit ctx und Nullwerten für die übrigen Parameter und liefert
// die Zahl der Parameter ohne ctx und den Fehler (den letzten Rückgabewert).
func aufrufMitNullwerten(ctx context.Context, t *testing.T, impl any, m string) (int, error) {
	t.Helper()
	mv := reflect.ValueOf(impl).MethodByName(m)
	typ := mv.Type()
	args := []reflect.Value{reflect.ValueOf(ctx)}
	for i := 1; i < typ.NumIn(); i++ {
		args = append(args, reflect.Zero(typ.In(i)))
	}
	out := mv.Call(args)
	var err error
	if e, ok := out[len(out)-1].Interface().(error); ok {
		err = e
	}
	return typ.NumIn() - 1, err
}

// ports nennt die Ports der Konfig und der Transaktion mit dem Präfix, unter dem das Protokoll ihre Operationen führt.
func portsDerFakes(u *umg, st outbound.WorkspaceStores, tx bool) []struct {
	praefix string
	impl    any
	typ     reflect.Type
} {
	if !tx {
		return []struct {
			praefix string
			impl    any
			typ     reflect.Type
		}{
			{"Sessions", u.sess, reflect.TypeOf((*outbound.SessionRepository)(nil)).Elem()},
			{"Exchanges", u.ex, reflect.TypeOf((*outbound.RecordingRepository)(nil)).Elem()},
			{"Blobs", u.bl, reflect.TypeOf((*outbound.BlobRepository)(nil)).Elem()},
		}
	}
	return []struct {
		praefix string
		impl    any
		typ     reflect.Type
	}{
		{"tx Workspaces", st.Workspaces, reflect.TypeOf((*outbound.WorkspaceRepository)(nil)).Elem()},
		{"tx Entities", st.Entities, reflect.TypeOf((*outbound.EntityRepository)(nil)).Elem()},
		{"tx Cursors", st.Cursors, reflect.TypeOf((*outbound.CursorRepository)(nil)).Elem()},
		{"tx Snapshots", st.Snapshots, reflect.TypeOf((*outbound.SnapshotRepository)(nil)).Elem()},
	}
}

// TestFakesZaehlenJedeOperation hält Z55: über die Methodenmengen der Ports (Reflexion) zählt jede der 34
// Operationen, hält ihre Argumente fest und speist einen Fehler ein; mit den fünf der Telemetrie sind es 39.
func TestFakesZaehlenJedeOperation(t *testing.T) {
	ctx := context.Background()
	gesamt := 0
	pruefe := func(u *umg, st outbound.WorkspaceStores, tx bool) {
		for _, pt := range portsDerFakes(u, st, tx) {
			for i := 0; i < pt.typ.NumMethod(); i++ {
				name := pt.praefix + "." + pt.typ.Method(i).Name
				gesamt++
				u.p.leeren()
				u.p.setzeFehler(name, nil)
				n, _ := aufrufMitNullwerten(ctx, t, pt.impl, pt.typ.Method(i).Name)
				evs := u.p.von(name)
				if len(evs) != 1 || len(evs[0].args) != n {
					t.Errorf("%s: %d Ereignisse, Argumente %v (erwartet %d)", name, len(evs), evs, n)
				}
				u.p.leeren()
				fehler := errors.New("eingespeist " + name)
				u.p.setzeFehler(name, fehler)
				if _, err := aufrufMitNullwerten(ctx, t, pt.impl, pt.typ.Method(i).Name); !errors.Is(err, fehler) {
					t.Errorf("%s: eingespeister Fehler nicht geliefert: %v", name, err)
				}
				u.p.setzeFehler(name, nil)
			}
		}
	}
	u := neu(t)
	pruefe(u, outbound.WorkspaceStores{}, false)
	if err := u.st.WithWorkspaceTransaction(ctx, func(st outbound.WorkspaceStores) error {
		if u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 1 {
			t.Errorf("Transaktion nicht gezählt")
		}
		pruefe(u, st, true)
		return nil
	}); err != nil {
		t.Fatalf("Transaktion: %v", err)
	}
	if gesamt+1 != len(alleOps()) || gesamt != 33 {
		t.Errorf("%d Operationen der Ports, erwartet 33 plus die Transaktion", gesamt)
	}
}

// TestFakeStoreRollbackFifoPanik hält Z55: Rollback bei Fehler und Panik, FIFO der Eindringlinge, die Panik
// wird weitergereicht und gibt die Schlange frei.
func TestFakeStoreRollbackFifoPanik(t *testing.T) {
	ctx := context.Background()
	cursor := model.ReplayCursor{WorkspaceID: wsID, ReplaySessionID: repID, RecordingSessionID: sesID, ServiceID: svcID, Strategy: model.StrategyStrict}
	schreibe := func(st outbound.WorkspaceStores) error { return st.Cursors.Put(ctx, cursor) }
	u := neu(t)
	// Eine eingespeiste Panik einer Operation erreicht den Aufrufer, auch aus der Transaktion, und gibt die Schlange frei.
	for _, op := range []string{"Sessions.Get", "tx Workspaces.Get"} {
		up := neu(t)
		erfolg(up)
		up.p.setzePanik(op, "paff-"+op)
		func() {
			defer func() {
				if r := recover(); r != "paff-"+op {
					t.Errorf("%s: Panik %v nicht weitergereicht", op, r)
				}
			}()
			_, _ = up.replay(ctx, anf())
		}()
		up.p.setzePanik(op, nil)
		if err := up.st.WithWorkspaceTransaction(ctx, func(outbound.WorkspaceStores) error { return nil }); err != nil {
			t.Errorf("%s: Schlange nach der Panik blockiert: %v", op, err)
		}
	}
	// Fehler: zurückgerollt.
	boom := errors.New("boom")
	if err := u.st.WithWorkspaceTransaction(ctx, func(st outbound.WorkspaceStores) error { _ = schreibe(st); return boom }); !errors.Is(err, boom) {
		t.Fatalf("Fehler nicht weitergereicht: %v", err)
	}
	if _, da := u.cursorStand(); da {
		t.Error("Rollback bei Fehler fehlt")
	}
	// Bestätigt.
	if err := u.st.WithWorkspaceTransaction(ctx, schreibe); err != nil {
		t.Fatal(err)
	}
	if _, da := u.cursorStand(); !da {
		t.Error("Bestätigung schreibt nicht")
	}
	// Panik: weitergereicht, zurückgerollt, Schlange frei.
	u2 := neu(t)
	func() {
		defer func() {
			if r := recover(); r != "paff" {
				t.Errorf("Panik nicht weitergereicht: %v", r)
			}
		}()
		_ = u2.st.WithWorkspaceTransaction(ctx, func(st outbound.WorkspaceStores) error { _ = schreibe(st); panic("paff") })
	}()
	if _, da := u2.cursorStand(); da {
		t.Error("Rollback bei Panik fehlt")
	}
	if err := u2.st.WithWorkspaceTransaction(ctx, func(outbound.WorkspaceStores) error { return nil }); err != nil {
		t.Errorf("Schlange nach der Panik blockiert: %v", err)
	}
	// FIFO: die Eindringlinge laufen in der Reihenfolge ihres Eintritts, nach der laufenden Transaktion.
	var reihe []int
	merke := func(n int) func(outbound.WorkspaceStores) error {
		return func(outbound.WorkspaceStores) error { reihe = append(reihe, n); return nil }
	}
	var kanaele []<-chan error
	err := u2.st.WithWorkspaceTransaction(ctx, func(outbound.WorkspaceStores) error {
		for n := 1; n <= 3; n++ {
			kanaele = append(kanaele, u2.st.eindringling(ctx, merke(n)))
		}
		reihe = append(reihe, 0)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range kanaele {
		<-c
	}
	if !reflect.DeepEqual(reihe, []int{0, 1, 2, 3}) {
		t.Errorf("Reihenfolge %v, erwartet [0 1 2 3]", reihe)
	}
}

// TestHuellen hält Z56: die Hülle des Stroms trägt jede Methode des Originals und zählt Read, WriteTo und
// Close; die Hülle des ctx trägt die vier Methoden und zählt jede.
func TestHuellen(t *testing.T) {
	orig := reflect.TypeOf(&bytes.Reader{})
	huelle := reflect.TypeOf(&strom{})
	for i := 0; i < orig.NumMethod(); i++ {
		if _, da := huelle.MethodByName(orig.Method(i).Name); !da {
			t.Errorf("Hülle des Stroms trägt %s nicht", orig.Method(i).Name)
		}
	}
	u := neu(t)
	s := &strom{Reader: bytes.NewReader([]byte("abc")), p: u.p}
	var w bytes.Buffer
	_, _ = s.WriteTo(&w)
	_, _ = s.Read(make([]byte, 1))
	_ = s.Close()
	if s.writeTos != 1 || s.reads != 1 || s.geschlossen() != 1 || w.String() != "abc" {
		t.Errorf("Zähler der Hülle: WriteTo %d, Read %d, Close %d", s.writeTos, s.reads, s.geschlossen())
	}
	k := neuerKontext(context.Background())
	k.Deadline()
	k.Done()
	_ = k.Err()
	k.Value(1)
	if k.deadline.Load() != 1 || k.done.Load() != 1 || k.err.Load() != 1 || k.value.Load() != 1 {
		t.Error("Zähler der Hülle des ctx")
	}
	if reflect.TypeOf(k).NumMethod() < reflect.TypeOf((*context.Context)(nil)).Elem().NumMethod() {
		t.Error("Hülle des ctx trägt nicht alle Methoden")
	}
}
