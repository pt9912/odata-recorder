package baselineimport_test

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestFakesZaehlenJedeOperation: jede Operation der acht Fakes, der Uhr und des Generators zählt
// genau einmal je Aufruf, auch die, die der Dienst nie rufen darf; die Menge folgt den Methoden der
// Schnittstellen (Reflexion), nicht einer Liste.
func TestFakesZaehlenJedeOperation(t *testing.T) {
	w := neu(t)
	ports := []any{w.sessions, w.exchanges, w.blobs, w.metadaten, w.snapshots, w.imports, w.meta, w.daten}
	ifaces := []reflect.Type{
		reflect.TypeOf((*outbound.SessionRepository)(nil)).Elem(), reflect.TypeOf((*outbound.RecordingRepository)(nil)).Elem(),
		reflect.TypeOf((*outbound.BlobRepository)(nil)).Elem(), reflect.TypeOf((*outbound.MetadataRepository)(nil)).Elem(),
		reflect.TypeOf((*outbound.SnapshotRepository)(nil)).Elem(), reflect.TypeOf((*outbound.ImportRepository)(nil)).Elem(),
		reflect.TypeOf((*outbound.MetadataCodec)(nil)).Elem(), reflect.TypeOf((*outbound.ODataCodec)(nil)).Elem(),
	}
	methoden := 0
	rufe := func(v reflect.Value, typ reflect.Type) {
		for i := 0; i < typ.NumMethod(); i++ {
			m := typ.Method(i)
			args := make([]reflect.Value, m.Type.NumIn())
			for j := range args {
				args[j] = reflect.Zero(m.Type.In(j))
			}
			if m.Type.IsVariadic() {
				t.Fatal("variadische Methode")
			}
			vorher := w.gesamt()
			v.MethodByName(m.Name).Call(args)
			if w.gesamt() != vorher+1 {
				t.Errorf("%s.%s: Zähler %d -> %d", typ.Name(), m.Name, vorher, w.gesamt())
			}
			methoden++
		}
	}
	for i, p := range ports {
		rufe(reflect.ValueOf(p), ifaces[i])
	}
	if methoden != 6+3+5+4+5+4+1+2 {
		t.Errorf("Methoden %d", methoden)
	}
	// Writer, Strom, Uhr und Generator.
	wr, err := w.blobs.Begin(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	for name, f := range map[string]func(){
		"Write":  func() { _, _ = wr.Write([]byte("x")) },
		"Commit": func() { _, _, _ = wr.Commit() },
		"Abort":  func() { _ = wr.Abort() },
	} {
		vorher := w.p.anzahl(name)
		f()
		if w.p.anzahl(name) != vorher+1 {
			t.Errorf("Writer.%s zählt nicht", name)
		}
	}
	w.blobs.inhalt["sha256:x"] = []byte("abc")
	rc, err := w.blobs.Open(context.Background(), "sha256:x")
	if err != nil {
		t.Fatal(err)
	}
	_ = rc.Close()
	if w.p.anzahl("Close") != 1 {
		t.Error("Close zählt nicht")
	}
	uhr, ids := w.konfig.Uhr, w.konfig.IDs
	vorher := w.gesamt()
	uhr.Now()
	ids.SnapshotID()
	ids.SessionID()
	ids.ExchangeID()
	ids.KorrelationsID()
	ids.WorkspaceID()
	if w.gesamt() != vorher+6 {
		t.Errorf("Uhr und Generator zählen %d von 6", w.gesamt()-vorher)
	}
	// Die Zähler der verbotenen Operationen stehen unter den gezählten.
	if w.p.verbotenGerufen() == 0 {
		t.Error("die verbotenen Operationen zählen nicht")
	}
	w.meta.fremd, w.daten.fremd = 0, 0
}

// TestFakesHaltenJedesArgument: Kontext und Argumente jeder Operation stehen im Protokoll.
func TestFakesHaltenJedesArgument(t *testing.T) {
	w := neu(t)
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, 1)
	_, _ = w.sessions.Get(ctx, "s")
	_, _ = w.exchanges.ListBySession(ctx, "s")
	_, _ = w.blobs.Begin(ctx, 7)
	_ = w.metadaten.SetMetadata(ctx, "d", "b", "h")
	_, _ = w.snapshots.Get(ctx, "x")
	_, _ = w.imports.Get(ctx, "i")
	_, _, _ = w.imports.Publish(ctx, model.Publication{})
	_, _, _ = w.meta.ParseMetadata(ctx, nil, 5)
	_, _ = w.daten.ParseCollection(ctx, nil, model.ServiceModel{}, "Set", 6)
	_, _ = w.daten.ParseEntity(ctx, nil, model.ServiceModel{}, "Set", 6)
	soll := map[string][]any{
		"Sessions.Get": {"s"}, "Exchanges.ListBySession": {"s"}, "Blobs.Begin": {int64(7)}, "Metadaten.SetMetadata": {"d", "b", "h"},
		"Snapshots.Get": {"x"}, "Imports.Get": {"i"}, "ParseMetadata": {nil, int64(5)}, "ParseCollection": {nil, model.ServiceModel{}, "Set", int64(6)},
	}
	for op, args := range soll {
		got := w.p.argsVon(op)
		if len(got) != 1 || len(got[0]) != len(args) {
			t.Errorf("%s: %v", op, got)
			continue
		}
		for i, a := range args {
			if !reflect.DeepEqual(got[0][i], a) {
				t.Errorf("%s Argument %d: %v, erwartet %v", op, i, got[0][i], a)
			}
		}
		if c := w.p.ctxVon(op); len(c) != 1 || c[0] != ctx {
			t.Errorf("%s: Kontext", op)
		}
	}
	if len(w.p.argsVon("Imports.Publish")) != 1 || len(w.daten.aufrufe["ParseEntity"]) != 1 {
		t.Error("Publish oder ParseEntity nicht festgehalten")
	}
	w.meta.fremd, w.daten.fremd = 0, 0
}

// TestFakeStromHuelle: der Strom von Open trägt jede Schnittstelle des Lesers, den er umhüllt, und
// zählt Close; Open liefert für jeden Aufruf einen eigenen Strom, den der Fake festhält.
func TestFakeStromHuelle(t *testing.T) {
	st := reflect.TypeOf(&strom{})
	for i := 0; i < reflect.TypeOf(&bytes.Reader{}).NumMethod(); i++ {
		m := reflect.TypeOf(&bytes.Reader{}).Method(i).Name
		if _, ok := st.MethodByName(m); !ok {
			t.Errorf("der Strom verdeckt die Methode %s des Lesers", m)
		}
	}
	for _, schnittstelle := range []reflect.Type{
		reflect.TypeOf((*io.ReaderAt)(nil)).Elem(), reflect.TypeOf((*io.ByteScanner)(nil)).Elem(), reflect.TypeOf((*io.RuneScanner)(nil)).Elem(),
		reflect.TypeOf((*io.WriterTo)(nil)).Elem(), reflect.TypeOf((*io.Seeker)(nil)).Elem(), reflect.TypeOf((*io.Closer)(nil)).Elem(),
	} {
		if !st.Implements(schnittstelle) {
			t.Errorf("der Strom trägt %v nicht", schnittstelle)
		}
	}
	w := neu(t)
	w.blobs.inhalt["sha256:a"] = []byte("abc")
	a, _ := w.blobs.Open(context.Background(), "sha256:a")
	b, _ := w.blobs.Open(context.Background(), "sha256:a")
	if a == b || len(w.blobs.strome()) != 2 || io.Reader(w.blobs.strome()[0]) != io.Reader(a.(*strom)) {
		t.Error("Open liefert nicht für jeden Aufruf einen eigenen, festgehaltenen Strom")
	}
	_ = a.Close()
	_ = a.Close()
	if w.blobs.strome()[0].geschlossen() != 2 || w.blobs.strome()[1].geschlossen() != 0 {
		t.Error("Close zählt nicht je Strom")
	}
}
