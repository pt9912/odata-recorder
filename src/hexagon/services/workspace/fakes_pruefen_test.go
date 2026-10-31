package workspace_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

type fakePort struct {
	name  string
	port  any
	iface reflect.Type
}

func typ[T any]() reflect.Type { return reflect.TypeOf((*T)(nil)).Elem() }

func (w *welt) fakePorts() []fakePort {
	return []fakePort{
		{"Workspaces", w.st.workspaces(), typ[outbound.WorkspaceRepository]()},
		{"Entities", entRepo{w.st, false}, typ[outbound.EntityRepository]()},
		{"Cursors", curRepo{w.st, false}, typ[outbound.CursorRepository]()},
		{"Snapshots", w.snaps, typ[outbound.SnapshotRepository]()},
		{"Blobs", w.blobs, typ[outbound.BlobRepository]()},
		{"Dienste", w.dienste, typ[outbound.MetadataRepository]()},
		{"Wartung", w.wartung, typ[inbound.MaintenanceUseCase]()},
		{"Transaktion", w.st, typ[outbound.WorkspaceTransactor]()},
		{"Uhr", w.uhr, typ[support.Clock]()},
		{"IDs", w.ids, typ[support.IDGenerator]()},
	}
}

// opName ist der Name, unter dem das Protokoll die Methode m des Ports p führt.
func opName(p fakePort, m string) string {
	switch p.name {
	case "Transaktion":
		return "Transaktion.WithWorkspaceTransaction"
	}
	return p.name + "." + m
}

// aufrufen ruft die Methode m des Ports mit ctx, null-Werten und f als Funktion; es liefert, ob die Methode
// einen Fehler zurückgibt, und diesen.
func aufrufen(ctx context.Context, p fakePort, m reflect.Method) (bool, error) {
	args := make([]reflect.Value, 0, m.Type.NumIn())
	for j := 0; j < m.Type.NumIn(); j++ {
		in := m.Type.In(j)
		switch {
		case in == typ[context.Context]():
			args = append(args, reflect.ValueOf(&ctx).Elem())
		case in.Kind() == reflect.Func:
			args = append(args, reflect.ValueOf(func(outbound.WorkspaceStores) error { return nil }))
		default:
			args = append(args, reflect.Zero(in))
		}
	}
	out := reflect.ValueOf(p.port).MethodByName(m.Name).Call(args)
	if n := len(out); n > 0 && out[n-1].Type() == typ[error]() {
		if out[n-1].IsNil() {
			return true, nil
		}
		return true, out[n-1].Interface().(error)
	}
	return false, nil
}

// TestFakesZaehlenJedeOperation (Z63): jede Operation jedes Fakes, auch die der Dienst nie rufen darf, zählt
// genau einmal je Aufruf, hält Kontext und jedes Argument fest, speist einen Fehler ein und ruft den Haltepunkt;
// die Menge folgt den Methoden der Schnittstellen (Reflexion), nicht einer Liste.
func TestFakesZaehlenJedeOperation(t *testing.T) {
	w := neu(t)
	methoden := 0
	gerufen := map[string]int{}
	w.p.nach = func(op string) { gerufen[op]++ }
	for _, p := range w.fakePorts() {
		for i := 0; i < p.iface.NumMethod(); i++ {
			m := p.iface.Method(i)
			op := opName(p, m.Name)
			ctxWert := context.WithValue(context.Background(), ctxSchluessel{}, op)
			vorher := w.p.gesamt()
			hatFehler, _ := aufrufen(ctxWert, p, m)
			if w.p.gesamt() != vorher+1 || w.p.anzahl(op) != 1 {
				t.Errorf("%s: Zähler %d -> %d, Operation %d mal", op, vorher, w.p.gesamt(), w.p.anzahl(op))
			}
			last := w.p.von(op)[0]
			wantArgs := m.Type.NumIn() - 1
			if p.name == "Uhr" || p.name == "IDs" {
				wantArgs = 0
			}
			if len(last.args) != wantArgs {
				t.Errorf("%s: %d Argumente festgehalten, erwartet %d", op, len(last.args), wantArgs)
			}
			if p.name != "Uhr" && p.name != "IDs" && last.ctx != ctxWert {
				t.Errorf("%s: Kontext nicht festgehalten", op)
			}
			// Der Haltepunkt gehört dem Store und Snapshots.Get, den Ports, an denen ein Test einen Schreiber einreiht.
			mitHaltepunkt := p.name == "Workspaces" || p.name == "Entities" || p.name == "Cursors" || op == "Snapshots.Get"
			if (gerufen[op] == 1) != mitHaltepunkt {
				t.Errorf("%s: Haltepunkt %d mal gerufen", op, gerufen[op])
			}
			if hatFehler {
				w.p.setzeFehler(op, errPlatte)
				if _, err := aufrufen(ctxWert, p, m); !errors.Is(err, errPlatte) {
					t.Errorf("%s: eingespeister Fehler %v", op, err)
				}
				abgebr, cancel := context.WithCancel(context.Background())
				cancel()
				if _, err := aufrufen(abgebr, p, m); !errors.Is(err, context.Canceled) {
					t.Errorf("%s: abgebrochener Kontext vor dem eingespeisten Fehler: %v", op, err)
				}
				w.p.setzeFehler(op, nil)
			}
			methoden++
		}
	}
	if methoden != 5+5+4+5+5+4+2+1+1+5 {
		t.Errorf("Methoden %d", methoden)
	}
	if w.p.verboten.Load() != 12 {
		t.Errorf("verbotene Aufrufe %d, erwartet 12 (vier Operationen außer Get, je dreimal gerufen)", w.p.verboten.Load())
	}
	w.p.verboten.Store(0)
}

// TestFakesInDerTransaktion (Z63): die Sichten innerhalb der Transaktion zählen mit dem Präfix "tx ".
func TestFakesInDerTransaktion(t *testing.T) {
	w := neu(t)
	zahl := 0
	err := w.st.WithWorkspaceTransaction(context.Background(), func(s outbound.WorkspaceStores) error {
		for _, sicht := range []fakePort{
			{"Workspaces", s.Workspaces, typ[outbound.WorkspaceRepository]()}, {"Entities", s.Entities, typ[outbound.EntityRepository]()},
			{"Cursors", s.Cursors, typ[outbound.CursorRepository]()}, {"Snapshots", s.Snapshots, typ[outbound.SnapshotRepository]()},
		} {
			for i := 0; i < sicht.iface.NumMethod(); i++ {
				m := sicht.iface.Method(i)
				vorher := w.p.anzahl("tx " + sicht.name + "." + m.Name)
				_, _ = aufrufen(context.Background(), sicht, m)
				if w.p.anzahl("tx "+sicht.name+"."+m.Name) != vorher+1 {
					t.Errorf("tx %s.%s: nicht gezählt", sicht.name, m.Name)
				}
				zahl++
			}
		}
		return nil
	})
	w.p.verboten.Store(0)
	if err != nil || zahl != 5+5+4+5 {
		t.Errorf("%v, Methoden %d", err, zahl)
	}
}

// TestFakeTransaktion (Z63): Rollback bei Fehler, Panik und Kontextabbruch, Weitergabe der Panik, FIFO.
func TestFakeTransaktion(t *testing.T) {
	ctx := context.Background()
	schreiben := func(s outbound.WorkspaceStores) error {
		return s.Workspaces.Create(ctx, wsNeu("ws-x"))
	}
	t.Run("Fehler", func(t *testing.T) {
		w := neu(t)
		err := w.st.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
			_ = schreiben(s)
			return errPlatte
		})
		if err != errPlatte || len(w.st.stand().ws) != 0 {
			t.Errorf("%v, Workspaces %d", err, len(w.st.stand().ws))
		}
	})
	t.Run("Panik", func(t *testing.T) {
		w := neu(t)
		var r any
		func() {
			defer func() { r = recover() }()
			_ = w.st.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
				_ = schreiben(s)
				panic("Wert")
			})
		}()
		if r != "Wert" || len(w.st.stand().ws) != 0 || w.st.aktiv.Load() {
			t.Errorf("recover %v, Workspaces %d, aktiv %v", r, len(w.st.stand().ws), w.st.aktiv.Load())
		}
		if err := w.st.WithWorkspaceTransaction(ctx, schreiben); err != nil {
			t.Errorf("die Schlange ist nach der Panik blockiert: %v", err)
		}
	})
	t.Run("Kontextabbruch", func(t *testing.T) {
		w := neu(t)
		c, cancel := context.WithCancel(ctx)
		err := w.st.WithWorkspaceTransaction(c, func(s outbound.WorkspaceStores) error {
			_ = schreiben(s)
			cancel()
			return nil
		})
		if !errors.Is(err, context.Canceled) || len(w.st.stand().ws) != 0 {
			t.Errorf("%v, Workspaces %d", err, len(w.st.stand().ws))
		}
	})
	t.Run("Erfolg", func(t *testing.T) {
		w := neu(t)
		if err := w.st.WithWorkspaceTransaction(ctx, schreiben); err != nil || len(w.st.stand().ws) != 1 {
			t.Errorf("%v", err)
		}
	})
	t.Run("FIFO", func(t *testing.T) {
		w := neu(t)
		var folge []int
		var ch []<-chan error
		err := w.st.WithWorkspaceTransaction(ctx, func(s outbound.WorkspaceStores) error {
			for i := 0; i < 5; i++ {
				ch = append(ch, w.st.eindringling(ctx, func(outbound.WorkspaceStores) error { folge = append(folge, i); return nil }))
			}
			folge = append(folge, -1)
			return nil
		})
		for _, c := range ch {
			<-c
		}
		if err != nil || !reflect.DeepEqual(folge, []int{-1, 0, 1, 2, 3, 4}) {
			t.Errorf("Folge %v, %v", folge, err)
		}
	})
	t.Run("außerhalb der Transaktion", func(t *testing.T) {
		w := neu(t)
		if err := w.st.workspaces().Delete(ctx, "gibt-es-nicht"); !errors.Is(err, outbound.ErrNichtGefunden) {
			t.Errorf("Delete: %v", err)
		}
		if w.p.anzahl("Workspaces.Delete") != 1 || len(w.st.q) != 0 {
			t.Errorf("Schlange nach einer Operation außerhalb: %d", len(w.st.q))
		}
	})
}

// TestHuellen (Z64): die Hülle um den Strom trägt jede Methode des umhüllten Lesers und Close, die Hülle um das
// ctx jede Methode von context.Context; beide zählen je Aufruf, und der Lesefehler liegt am Offset k.
func TestHuellen(t *testing.T) {
	st := reflect.TypeOf(&strom{})
	for _, orig := range []reflect.Type{reflect.TypeOf(&bytes.Reader{}), typ[io.Closer]()} {
		for i := 0; i < orig.NumMethod(); i++ {
			m := orig.Method(i)
			if _, ok := st.MethodByName(m.Name); !ok {
				t.Errorf("der Strom verdeckt die Methode %s", m.Name)
			}
		}
	}
	for _, s := range []reflect.Type{typ[io.ReaderAt](), typ[io.ByteScanner](), typ[io.RuneScanner](), typ[io.WriterTo](), typ[io.Seeker](), typ[io.ReadCloser]()} {
		if !st.Implements(s) {
			t.Errorf("der Strom trägt %v nicht", s)
		}
	}
	kt := reflect.TypeOf(&kontext{})
	for i := 0; i < typ[context.Context]().NumMethod(); i++ {
		if !kt.Implements(typ[context.Context]()) {
			t.Fatal("die Hülle trägt context.Context nicht")
		}
	}
	k := neuerKontext(context.Background())
	_, _ = k.Deadline()
	_ = k.Done()
	_ = k.Err()
	_ = k.Err()
	_ = k.Value(ctxSchluessel{})
	if k.deadline.Load() != 1 || k.done.Load() != 1 || k.err.Load() != 2 || k.value.Load() != 1 {
		t.Errorf("Zähler der Kontext-Hülle: %d %d %d %d", k.deadline.Load(), k.done.Load(), k.err.Load(), k.value.Load())
	}
	w := neu(t)
	w.blobs.inhalt["sha256:a"] = []byte("abcdefghij")
	w.blobs.fehlerBei, w.blobs.lesefehler = 4, errPlatte
	rc, err := w.blobs.Open(context.Background(), "sha256:a")
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(rc)
	if string(got) != "abcd" || !errors.Is(err, errPlatte) {
		t.Errorf("Lesen bis Offset 4: %q, %v", got, err)
	}
	if w.p.anzahl("Strom.Read") < 2 {
		t.Errorf("Read zählt %d", w.p.anzahl("Strom.Read"))
	}
	var sb strings.Builder
	rc2, _ := w.blobs.Open(context.Background(), "sha256:a")
	if _, err := rc2.(io.WriterTo).WriteTo(&sb); !errors.Is(err, errPlatte) || sb.String() != "abcd" {
		t.Errorf("WriteTo: %q, %v", sb.String(), err)
	}
	if w.p.anzahl("Strom.WriteTo") != 1 {
		t.Errorf("WriteTo zählt %d", w.p.anzahl("Strom.WriteTo"))
	}
	_ = rc.Close()
	_ = rc.Close()
	if w.blobs.strome()[0].geschlossen() != 2 || w.p.anzahl("Strom.Close") != 2 {
		t.Error("Close zählt nicht je Aufruf")
	}
	if len(w.blobs.strome()) != 2 || w.blobs.strome()[0] == w.blobs.strome()[1] {
		t.Error("Open liefert nicht für jeden Aufruf einen eigenen Strom")
	}
}

func wsNeu(id string) model.Workspace {
	return model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", CreatedAt: t0()}
}
