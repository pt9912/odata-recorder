package supporttest_test

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

type ctxSchluessel struct{}

// P11: der Fake hält jedes Argument ungeprüft fest und zählt jede der sechs Operationen.
func TestFakeAufzeichnung(t *testing.T) {
	f := supporttest.NewFakeTelemetry()
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, "wert")
	fremd := []support.Attr{support.Str("ausserhalb_der_menge", "mit Leerraum"), {Key: "", Art: 99}}
	f.Log(ctx, slog.LevelWarn, "text", fremd...)
	eingabe := []support.Attr{support.Str("recorder.mode", "record")}
	ctxKind, span := f.StartSpan(ctx, support.SpanGatewayRequest, eingabe...)
	eingabe[0] = support.Str("geaendert", "x")
	setzen := []support.Attr{support.Int("http.response.status_code", 200)}
	span.Set(setzen...)
	setzen[0] = support.Str("geaendert", "x")
	span.End()
	span.End()
	span.Set(support.Str("nach_dem_ende", "x"))
	f.CountError(model.CodeE4001)
	f.CountError(0)
	f.Snapshot()
	fremd[0] = support.Str("geaendert", "x")

	ops := []supporttest.Operation{
		supporttest.OpLog, supporttest.OpStartSpan, supporttest.OpCountError,
		supporttest.OpSet, supporttest.OpEnd, supporttest.OpSnapshot,
	}
	want := map[supporttest.Operation]int{
		supporttest.OpLog: 1, supporttest.OpStartSpan: 1, supporttest.OpCountError: 2,
		supporttest.OpSet: 2, supporttest.OpEnd: 2, supporttest.OpSnapshot: 1,
	}
	if len(ops) != 6 {
		t.Fatalf("%d Operationen, erwartet 6", len(ops))
	}
	for _, op := range ops {
		if got := f.Anzahl(op); got != want[op] {
			t.Errorf("Anzahl(%d) = %d, erwartet %d", op, got, want[op])
		}
	}

	a := f.Aufrufe()
	folge := []supporttest.Operation{
		supporttest.OpLog, supporttest.OpStartSpan, supporttest.OpSet, supporttest.OpEnd, supporttest.OpEnd,
		supporttest.OpSet, supporttest.OpCountError, supporttest.OpCountError, supporttest.OpSnapshot,
	}
	if len(a) != len(folge) {
		t.Fatalf("%d Aufrufe, erwartet %d", len(a), len(folge))
	}
	for i, op := range folge {
		if a[i].Op != op {
			t.Errorf("Aufruf %d: Operation %d, erwartet %d", i, a[i].Op, op)
		}
	}
	if a[0].Ctx() != ctx || a[0].Level != slog.LevelWarn || a[0].Msg != "text" {
		t.Errorf("Log: ctx, Stufe oder Text nicht festgehalten: %+v", a[0])
	}
	if want := []support.Attr{support.Str("ausserhalb_der_menge", "mit Leerraum"), {Key: "", Art: 99}}; !reflect.DeepEqual(a[0].Attrs, want) {
		t.Errorf("Log: Attribute %+v, erwartet %+v (ungeprüft, unverändert nach der Änderung des Aufrufers)", a[0].Attrs, want)
	}
	if a[1].Ctx() != ctx || a[1].Name != support.SpanGatewayRequest || a[1].Span != span || a[1].Eltern != nil {
		t.Errorf("StartSpan: nicht festgehalten: %+v", a[1])
	}
	if want := []support.Attr{support.Str("recorder.mode", "record")}; !reflect.DeepEqual(a[1].Attrs, want) {
		t.Errorf("StartSpan: Attribute %+v, erwartet %+v", a[1].Attrs, want)
	}
	if ctxKind == ctx || ctxKind.Value(ctxSchluessel{}) != "wert" {
		t.Error("StartSpan gibt keinen abgeleiteten ctx zurück")
	}
	if want := []support.Attr{support.Int("http.response.status_code", 200)}; !reflect.DeepEqual(a[2].Attrs, want) || a[2].Span != span {
		t.Errorf("Set: %+v, erwartet %+v am Span", a[2], want)
	}
	if a[5].Span != span || a[5].Attrs[0].Key != "nach_dem_ende" {
		t.Errorf("Set nach End nicht aufgezeichnet: %+v", a[5])
	}
	if a[6].Code != model.CodeE4001 || a[7].Code != 0 {
		t.Errorf("CountError: Codes %d und %d, erwartet %d und 0", a[6].Code, a[7].Code, model.CodeE4001)
	}
	a[0].Attrs[0] = support.Str("geaendert", "x")
	if f.Aufrufe()[0].Attrs[0].Key != "ausserhalb_der_menge" {
		t.Error("Aufrufe() teilt Speicher mit der Aufzeichnung")
	}
}

func TestFakeSnapshotStand(t *testing.T) {
	f := supporttest.NewFakeTelemetry()
	if got := f.Snapshot(); !reflect.DeepEqual(got, support.Snapshot{}) {
		t.Errorf("ohne Stand %+v, erwartet den leeren", got)
	}
	stand := support.Snapshot{Errors: map[model.Code]uint64{model.CodeE2001: 3}}
	f.SetzeStand(stand)
	stand.Errors[model.CodeE2001] = 9
	got := f.Snapshot()
	if got.Errors[model.CodeE2001] != 3 {
		t.Errorf("Stand %+v, erwartet 3 (Kopie beim Setzen)", got)
	}
	got.Errors[model.CodeE2001] = 9
	if f.Snapshot().Errors[model.CodeE2001] != 3 {
		t.Error("Snapshot liefert keine Kopie")
	}
}

// P12: der Fake hält den Eltern-Span fest und gibt einen abgeleiteten ctx zurück.
func TestFakeEltern(t *testing.T) {
	t.Run("Kind im ctx des Elternspans", func(t *testing.T) {
		f := supporttest.NewFakeTelemetry()
		ctxE, eltern := f.StartSpan(context.Background(), support.SpanGatewayRequest)
		_, kind := f.StartSpan(ctxE, support.SpanReplayMatch)
		a := f.Aufrufe()
		if a[1].Eltern == nil || a[1].Eltern != eltern || a[1].Span != kind || a[0].Eltern != nil {
			t.Errorf("Eltern nicht festgehalten: %+v", a)
		}
		if kind.(*supporttest.FakeSpan).Eltern != eltern {
			t.Error("FakeSpan.Eltern ist nicht der Elternspan")
		}
	})
	t.Run("ohne Eltern", func(t *testing.T) {
		f := supporttest.NewFakeTelemetry()
		_, s := f.StartSpan(context.Background(), support.SpanMockQuery)
		if f.Aufrufe()[0].Eltern != nil || s.(*supporttest.FakeSpan).Eltern != nil {
			t.Error("Eltern ohne Eltern-Span im ctx")
		}
	})
	t.Run("Wert und Abbruch bleiben", func(t *testing.T) {
		f := supporttest.NewFakeTelemetry()
		ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, "wert"))
		ctxK, _ := f.StartSpan(ctx, support.SpanMockQuery)
		if ctxK.Value(ctxSchluessel{}) != "wert" {
			t.Error("der Wert des Eingabe-ctx fehlt")
		}
		if ctxK.Err() != nil {
			t.Fatal("ctx vor dem Abbruch schon abgebrochen")
		}
		cancel()
		if ctxK.Err() == nil {
			t.Error("der Abbruch des Eingabe-ctx erreicht den Rückgabe-ctx nicht")
		}
	})
	t.Run("nil-ctx", func(t *testing.T) {
		f := supporttest.NewFakeTelemetry()
		var kein context.Context
		ctxK, s := f.StartSpan(kein, support.SpanMockQuery)
		if ctxK == nil || s == nil {
			t.Error("StartSpan liefert nil")
		}
		if f.Aufrufe()[0].Ctx() != nil {
			t.Error("nil-ctx nicht mit seiner Identität festgehalten")
		}
	})
}

// P16: gleichzeitiger Aufruf; 64 Goroutinen mit je 500 Aufrufen, reihum über die sechs Operationen.
func TestFakeGleichzeitig(t *testing.T) {
	const goroutinen, aufrufe = 64, 500
	f := supporttest.NewFakeTelemetry()
	var wg sync.WaitGroup
	for range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var sp support.Span
			for i := range aufrufe {
				switch i % 6 {
				case 0:
					f.Log(context.Background(), slog.LevelInfo, "m")
				case 1:
					_, sp = f.StartSpan(context.Background(), support.SpanMockQuery)
				case 2:
					sp.Set()
				case 3:
					sp.End()
				case 4:
					f.CountError(model.CodeE2001)
				default:
					f.Snapshot()
				}
			}
		}()
	}
	wg.Wait()
	if n := len(f.Aufrufe()); n != goroutinen*aufrufe {
		t.Errorf("%d Einträge, erwartet %d", n, goroutinen*aufrufe)
	}
	je := [6]int{}
	for i := range aufrufe {
		je[i%6]++
	}
	ops := []supporttest.Operation{
		supporttest.OpLog, supporttest.OpStartSpan, supporttest.OpSet,
		supporttest.OpEnd, supporttest.OpCountError, supporttest.OpSnapshot,
	}
	for k, op := range ops {
		if got, want := f.Anzahl(op), je[k]*goroutinen; got != want {
			t.Errorf("Anzahl(%d) = %d, erwartet %d", op, got, want)
		}
	}
}
