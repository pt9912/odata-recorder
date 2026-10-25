package baselineimport_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

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

// nichtLeer baut einen Wert des Typs typ, der nirgends dem Nullwert gleicht (jedes Feld, das ein Test setzen
// kann, trägt den Text tag), damit ein Fake, der ein Argument verwirft oder nullt, auffällt.
func nichtLeer(t *testing.T, w *welt, typ reflect.Type, tag string) reflect.Value {
	t.Helper()
	switch typ {
	case reflect.TypeOf((*io.Reader)(nil)).Elem():
		return reflect.ValueOf(&strom{Reader: bytes.NewReader([]byte("abc")), ref: "r-" + tag, p: w.p})
	case reflect.TypeOf(model.ServiceModel{}):
		return reflect.ValueOf(modellR())
	case reflect.TypeOf(time.Time{}):
		return reflect.ValueOf(t0())
	}
	v := reflect.New(typ).Elem()
	switch typ.Kind() {
	case reflect.String:
		v.SetString(tag)
	case reflect.Int64, reflect.Int:
		v.SetInt(1000 + int64(len(tag)))
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.String {
			v.Set(reflect.ValueOf([]string{tag}).Convert(typ))
		}
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).IsExported() {
				v.Field(i).Set(nichtLeer(t, w, typ.Field(i).Type, tag+"."+typ.Field(i).Name))
			}
		}
	}
	return v
}

// TestFakesHaltenJedesArgument: Kontext und jedes Argument jeder Methode der acht Schnittstellen stehen im Protokoll
// (30 Methoden, die Zahl der Argumente zählt der Test per Reflexion); die Werte sind nirgends der Nullwert.
func TestFakesHaltenJedesArgument(t *testing.T) {
	w := neu(t)
	ports := []struct {
		praefix string
		port    any
		iface   reflect.Type
	}{
		{"Sessions", w.sessions, reflect.TypeOf((*outbound.SessionRepository)(nil)).Elem()},
		{"Exchanges", w.exchanges, reflect.TypeOf((*outbound.RecordingRepository)(nil)).Elem()},
		{"Blobs", w.blobs, reflect.TypeOf((*outbound.BlobRepository)(nil)).Elem()},
		{"Metadaten", w.metadaten, reflect.TypeOf((*outbound.MetadataRepository)(nil)).Elem()},
		{"Snapshots", w.snapshots, reflect.TypeOf((*outbound.SnapshotRepository)(nil)).Elem()},
		{"Imports", w.imports, reflect.TypeOf((*outbound.ImportRepository)(nil)).Elem()},
		{"", w.meta, reflect.TypeOf((*outbound.MetadataCodec)(nil)).Elem()},
		{"", w.daten, reflect.TypeOf((*outbound.ODataCodec)(nil)).Elem()},
	}
	methoden, argumente := 0, 0
	for _, p := range ports {
		for i := 0; i < p.iface.NumMethod(); i++ {
			m := p.iface.Method(i)
			op := m.Name
			if p.praefix != "" {
				op = p.praefix + "." + m.Name
			}
			ctx := context.WithValue(context.Background(), ctxSchluessel{}, op)
			args := []reflect.Value{reflect.ValueOf(ctx)}
			var soll []any
			for j := 1; j < m.Type.NumIn(); j++ {
				v := nichtLeer(t, w, m.Type.In(j), fmt.Sprint(op, "-", j))
				if v.IsZero() {
					t.Fatalf("%s Argument %d: kein Wert ungleich dem Nullwert für %v", op, j, m.Type.In(j))
				}
				args = append(args, v)
				soll = append(soll, v.Interface())
				argumente++
			}
			reflect.ValueOf(p.port).MethodByName(m.Name).Call(args)
			methoden++
			got := w.p.argsVon(op)
			if len(got) != 1 || len(got[0]) != len(soll) {
				t.Errorf("%s: %d Aufrufe mit %v, erwartet ein Aufruf mit %d Argumenten", op, len(got), got, len(soll))
				continue
			}
			for j, a := range soll {
				if !reflect.DeepEqual(got[0][j], a) {
					t.Errorf("%s Argument %d: %v, erwartet %v", op, j+1, got[0][j], a)
				}
			}
			if c := w.p.ctxVon(op); len(c) != 1 || c[0] != ctx {
				t.Errorf("%s: Kontext", op)
			}
		}
	}
	if methoden != 30 || argumente != 38 {
		t.Errorf("Methoden %d, Argumente %d, erwartet 30 und 38", methoden, argumente)
	}
	// Die Schreibseite des Blobs: Write hält die Bytes fest, Commit und Abort zählen.
	wr, err := w.blobs.Begin(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = wr.Write([]byte("bytes"))
	if a := w.p.argsVon("Write"); len(a) != 1 || !bytes.Equal(a[0][0].([]byte), []byte("bytes")) {
		t.Errorf("Write: %v", a)
	}
	w.meta.fremd, w.daten.fremd, w.daten.fehlt = 0, 0, 0
}

// TestFakeCodecsPruefenStrom: die Fakes der Codecs erkennen einen Strom, der nicht am Anfang steht (die Hülle ist
// geprüft, bevor ein Test auf sie baut): an allen drei Stellen, ParseMetadata, ParseCollection und ParseEntity,
// meldet der Fake einen gelesenen Strom, einen fremden Strom und zählt beide; ein unberührter Strom geht durch.
func TestFakeCodecsPruefenStrom(t *testing.T) {
	ctx := context.Background()
	rufe := map[string]func(w *welt, src io.Reader) error{
		"ParseMetadata": func(w *welt, src io.Reader) error { _, _, err := w.meta.ParseMetadata(ctx, src, 1); return err },
		"ParseCollection": func(w *welt, src io.Reader) error {
			_, err := w.daten.ParseCollection(ctx, src, model.ServiceModel{}, "Products", 1)
			return err
		},
		"ParseEntity": func(w *welt, src io.Reader) error {
			_, err := w.daten.ParseEntity(ctx, src, model.ServiceModel{}, "Products", 1)
			return err
		},
	}
	if len(rufe) != 3 {
		t.Fatalf("Stellen %d", len(rufe))
	}
	for name, f := range rufe {
		w := neu(t)
		w.blobs.inhalt["sha256:a"] = []byte("abc")
		w.daten.nachRef["sha256:a"] = datenVorgabe{}
		unberuehrt, _ := w.blobs.Open(ctx, "sha256:a")
		if err := f(w, unberuehrt); err != nil {
			t.Errorf("%s mit unberührtem Strom: %v", name, err)
		}
		gelesen, _ := w.blobs.Open(ctx, "sha256:a")
		_, _ = gelesen.Read(make([]byte, 1))
		if err := f(w, gelesen); !errors.Is(err, errStromVerbraucht) {
			t.Errorf("%s mit gelesenem Strom: %v", name, err)
		}
		if err := f(w, bytes.NewReader([]byte("abc"))); !errors.Is(err, errFremderStrom) {
			t.Errorf("%s mit fremdem Strom: %v", name, err)
		}
		if w.meta.verbraucht+w.daten.verbraucht != 1 || w.meta.fremd+w.daten.fremd != 1 {
			t.Errorf("%s: verbraucht %d, fremd %d", name, w.meta.verbraucht+w.daten.verbraucht, w.meta.fremd+w.daten.fremd)
		}
		w.meta.verbraucht, w.daten.verbraucht, w.meta.fremd, w.daten.fremd = 0, 0, 0, 0
	}
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
