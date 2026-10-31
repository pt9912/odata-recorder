package supporttest_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

type ctxSchluessel struct{}

// leererStand prüft SPEC-230 Nr. 10: vier leere, nicht-nil Sammlungen.
func leererStand(t *testing.T, s support.Snapshot) {
	t.Helper()
	if s.Errors == nil || s.Durations == nil || s.Dropped == nil || s.Spans == nil ||
		len(s.Errors)+len(s.Durations)+len(s.Dropped)+len(s.Spans) != 0 {
		t.Errorf("Stand %+v, erwartet vier leere, nicht-nil Sammlungen", s)
	}
}

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

// attrGleich vergleicht Attributlisten Feld für Feld; Gleit zählt über die Bits (NaN gleicht NaN,
// -0 gleicht +0 nicht).
func attrGleich(a, b []support.Attr) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if x.Key != y.Key || x.Art != y.Art || x.Text != y.Text || x.Ganz != y.Ganz ||
			math.Float64bits(x.Gleit) != math.Float64bits(y.Gleit) {
			return false
		}
	}
	return true
}

// P11: jedes Argument jeder Operation mit Argumenten wird ungeprüft und unverändert festgehalten.
// Menge: Operation × Argumentart. Ein Attribut ist ein Grundwert (recorder.mode, String, "record"),
// dessen Felder einzeln wechseln: Key 6 Werte (leer, gültig, fremd), Art 8 (0, -1, die drei gültigen,
// 99, ±MaxInt32), Text 7 (darunter Wertform verletzt: "A B", "Record"), Ganz 5, Gleit 7 (darunter
// NaN, -0, ±Inf), dazu mode und recorder.mode mit "A B" (gültiger Schlüssel, gültige Art, Wertform
// verletzt) = 35 Attribute. Listen: nil, leer, jedes der 35 Attribute einzeln, dazu je eine Liste
// mit 2, 3, 5 und 12 Attributen = 41. ctx: der übergebene und nil. Log 8 Stufen × 6 Texte × 2 ctx
// und 41 Listen × 2 ctx; StartSpan 9 Namen (darunter der leere und vier fremde) × 2 ctx × 41
// Listen; Set 41 Listen; CountError 9 Codes (darunter Codes mit hohen Bits). Die Zählprüfung liest
// die Mengen, das Orakel sind die Werte, die der Test übergibt.
func TestFakeArgumenteUngeprueft(t *testing.T) {
	stufen := []slog.Level{math.MinInt, slog.LevelDebug - 1, slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError, slog.LevelError + 1, math.MaxInt}
	texte := []string{"", " ", "a b", "Zeile\nUmbruch", string([]byte{0x80, 0}), "text"}
	namen := []support.SpanName{"", " ", "unbekannt", "odata.mock.query "}
	for _, s := range support.Spans() {
		namen = append(namen, s.Name)
	}
	grund := support.Str("recorder.mode", "record")
	var attrs []support.Attr
	variiere := func(wechsel func(*support.Attr)) {
		a := grund
		wechsel(&a)
		attrs = append(attrs, a)
	}
	for _, k := range []string{"", "mode", "recorder.mode", "session_id", "ausserhalb_der_menge", "x y"} {
		variiere(func(a *support.Attr) { a.Key = k })
	}
	for _, art := range []support.Art{0, -1, support.ArtString, support.ArtInt, support.ArtFloat, 99, math.MaxInt32, math.MinInt32} {
		variiere(func(a *support.Attr) { a.Art = art })
	}
	for _, text := range []string{"", " ", "A B", "Record", "Zeile\nUmbruch", string([]byte{0x80, 0}), "record"} {
		variiere(func(a *support.Attr) { a.Text = text })
	}
	for _, ganz := range []int64{0, 1, -1, math.MinInt64, math.MaxInt64} {
		variiere(func(a *support.Attr) { a.Ganz = ganz })
	}
	for _, gleit := range []float64{0, math.NaN(), math.Copysign(0, -1), math.Inf(1), math.Inf(-1), 1.5, math.MaxFloat64} {
		variiere(func(a *support.Attr) { a.Gleit = gleit })
	}
	attrs = append(attrs, support.Str("mode", "A B"), support.Str("recorder.mode", "A B"))
	listen := [][]support.Attr{nil, {}}
	for _, a := range attrs {
		listen = append(listen, []support.Attr{a})
	}
	for _, laenge := range []int{2, 3, 5, 12} {
		var l []support.Attr
		for j := range laenge {
			l = append(l, attrs[(j*7+laenge)%len(attrs)])
		}
		listen = append(listen, l)
	}
	codes := []model.Code{0, -1, 1, model.CodeE4001, model.CodeE4001 + math.MaxInt>>1, math.MaxInt, math.MinInt, math.MaxInt - 1, math.MinInt + 1}
	var kein context.Context
	ctx := context.WithValue(context.Background(), ctxSchluessel{}, "wert")
	ctxe := []context.Context{ctx, kein}
	if len(stufen) != 8 || len(texte) != 6 || len(namen) != 9 || len(attrs) != 35 || len(listen) != 41 || len(codes) != 9 || len(ctxe) != 2 {
		t.Fatalf("Mengen %d, %d, %d, %d, %d, %d, %d; erwartet 8, 6, 9, 35, 41, 9, 2", len(stufen), len(texte), len(namen), len(attrs), len(listen), len(codes), len(ctxe))
	}

	type erwartet struct {
		op    supporttest.Operation
		level slog.Level
		msg   string
		name  support.SpanName
		attrs []support.Attr
		code  model.Code
		ctxNr int // Index in ctxe
	}
	f := supporttest.NewFakeTelemetry()
	_, span := f.StartSpan(ctx, "start")
	folge := []erwartet{{op: supporttest.OpStartSpan, name: "start", ctxNr: 0}}
	for _, stufe := range stufen {
		for _, text := range texte {
			for ci, c := range ctxe {
				f.Log(c, stufe, text)
				folge = append(folge, erwartet{op: supporttest.OpLog, level: stufe, msg: text, ctxNr: ci})
			}
		}
	}
	for _, l := range listen {
		for ci, c := range ctxe {
			f.Log(c, slog.LevelInfo, "m", l...)
			folge = append(folge, erwartet{op: supporttest.OpLog, level: slog.LevelInfo, msg: "m", attrs: l, ctxNr: ci})
		}
	}
	for _, name := range namen {
		for ci, c := range ctxe {
			for _, l := range listen {
				f.StartSpan(c, name, l...)
				folge = append(folge, erwartet{op: supporttest.OpStartSpan, name: name, attrs: l, ctxNr: ci})
			}
		}
	}
	for _, l := range listen {
		span.Set(l...)
		folge = append(folge, erwartet{op: supporttest.OpSet, attrs: l})
	}
	for _, c := range codes {
		f.CountError(c)
		folge = append(folge, erwartet{op: supporttest.OpCountError, code: c})
	}

	a := f.Aufrufe()
	gezaehlt := 1 + 8*6*2 + 41*2 + 9*2*41 + 41 + 9
	if len(a) != len(folge) || len(folge) != gezaehlt {
		t.Fatalf("%d Aufrufe, erwartet %d (gezählt %d)", len(a), len(folge), gezaehlt)
	}
	je := map[supporttest.Operation]int{}
	for i, w := range folge {
		je[w.op]++
		g := a[i]
		if g.Op != w.op || g.Level != w.level || g.Msg != w.msg || g.Name != w.name || g.Code != w.code {
			t.Errorf("Aufruf %d: %+v, erwartet %+v", i, g, w)
		}
		if !attrGleich(g.Attrs, w.attrs) {
			t.Errorf("Aufruf %d: Attribute %+v, erwartet %+v", i, g.Attrs, w.attrs)
		}
		if (w.op == supporttest.OpLog || w.op == supporttest.OpStartSpan) && g.Ctx() != ctxe[w.ctxNr] {
			t.Errorf("Aufruf %d: ctx nicht mit seiner Identität festgehalten (erwartet nil: %t)", i, ctxe[w.ctxNr] == nil)
		}
		if w.op == supporttest.OpSet && g.Span != span {
			t.Errorf("Aufruf %d: Set am falschen Span", i)
		}
	}
	for op, n := range je {
		if f.Anzahl(op) != n {
			t.Errorf("Anzahl(%d) = %d, erwartet %d", op, f.Anzahl(op), n)
		}
	}
	if len(je) != 4 {
		t.Errorf("%d Operationen geprüft, erwartet 4", len(je))
	}
}

// P16: Anzahl und Aufrufe() sind für gleichzeitigen Aufruf sicher, auch gegen laufende Schreiber:
// 8 Schreiber zeichnen je 300 Runden mit allen sechs Operationen auf; zwei Leser rufen Anzahl für
// jede der sechs Operationen, zwei rufen Aufrufe(), solange die Schreiber laufen. Die Zähler fallen
// nie, die Aufzeichnung wird nie kürzer, und jeder Eintrag trägt eine der sechs Operationen. Das
// Fehlen der Sperre in Anzahl oder in Aufrufe() meldet der Detektor; die Wirkung zeigt der Test
// nicht in jedem Lauf.
func TestFakeLesenGleichzeitig(t *testing.T) {
	const schreiber, runden = 8, 300
	f := supporttest.NewFakeTelemetry()
	ops := []supporttest.Operation{
		supporttest.OpLog, supporttest.OpStartSpan, supporttest.OpCountError,
		supporttest.OpSet, supporttest.OpEnd, supporttest.OpSnapshot,
	}
	start := make(chan struct{})
	var fertig atomic.Bool
	var fehler, gelesen atomic.Int64
	var schreibend, lesend sync.WaitGroup
	for range schreiber {
		schreibend.Add(1)
		go func() {
			defer schreibend.Done()
			<-start
			for range runden {
				_, sp := f.StartSpan(context.Background(), support.SpanMockQuery)
				f.Log(context.Background(), slog.LevelInfo, "m")
				sp.Set()
				sp.End()
				f.CountError(model.CodeE2001)
				f.Snapshot()
			}
		}()
	}
	for range 2 {
		lesend.Add(1)
		go func() {
			defer lesend.Done()
			<-start
			var zuletzt [6]int
			for {
				for k, op := range ops {
					n := f.Anzahl(op)
					if n < zuletzt[k] {
						fehler.Add(1)
					}
					zuletzt[k] = n
					gelesen.Add(1)
				}
				if fertig.Load() {
					return
				}
			}
		}()
	}
	for range 2 {
		lesend.Add(1)
		go func() {
			defer lesend.Done()
			<-start
			zuletzt := 0
			for {
				a := f.Aufrufe()
				if len(a) < zuletzt {
					fehler.Add(1)
				}
				zuletzt = len(a)
				for _, e := range a {
					if e.Op < supporttest.OpLog || e.Op > supporttest.OpSnapshot {
						fehler.Add(1)
					}
				}
				gelesen.Add(1)
				if fertig.Load() {
					return
				}
			}
		}()
	}
	close(start)
	schreibend.Wait()
	fertig.Store(true)
	lesend.Wait()
	if n := fehler.Load(); n != 0 {
		t.Errorf("%d Lesefehler (ein Zähler fiel, die Aufzeichnung wurde kürzer oder ein Eintrag war unvollständig)", n)
	}
	if gelesen.Load() < 4 {
		t.Errorf("%d Lesungen, erwartet mindestens 4", gelesen.Load())
	}
	for _, op := range ops {
		if f.Anzahl(op) != schreiber*runden {
			t.Errorf("Anzahl(%d) = %d, erwartet %d", op, f.Anzahl(op), schreiber*runden)
		}
	}
	if n := len(f.Aufrufe()); n != 6*schreiber*runden {
		t.Errorf("%d Einträge, erwartet %d", n, 6*schreiber*runden)
	}
}

func TestFakeSnapshotStand(t *testing.T) {
	f := supporttest.NewFakeTelemetry()
	leererStand(t, f.Snapshot())
	f.SetzeStand(support.Snapshot{})
	if got := f.Snapshot(); got.Errors != nil || got.Durations != nil || got.Dropped != nil || got.Spans != nil {
		t.Errorf("gesetzter nil-Stand %+v, erwartet nil wie gesetzt", got)
	}
	f.SetzeStand(support.Snapshot{Errors: map[model.Code]uint64{}, Spans: []support.SpanRecord{}})
	if got := f.Snapshot(); got.Errors == nil || got.Spans == nil || got.Dropped != nil {
		t.Errorf("gesetzter leerer Stand %+v, erwartet leer wie gesetzt", got)
	}
	f = supporttest.NewFakeTelemetry()
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
	t.Run("Kette der Tiefe 1 bis 16", func(t *testing.T) {
		f := supporttest.NewFakeTelemetry()
		var kette func(ctx context.Context, vor *supporttest.FakeSpan, tiefe int)
		kette = func(ctx context.Context, vor *supporttest.FakeSpan, tiefe int) {
			if tiefe == 16 {
				return
			}
			c, s := f.StartSpan(ctx, support.SpanMockQuery)
			fs := s.(*supporttest.FakeSpan)
			if a := f.Aufrufe()[tiefe]; fs.Eltern != vor || a.Eltern != vor || a.Span != fs {
				t.Errorf("Tiefe %d: Eltern nicht festgehalten", tiefe+1)
			}
			kette(c, fs, tiefe+1)
		}
		kette(context.Background(), nil, 0)
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

type ctxZweiter struct{}

// P12: der ctx von StartSpan leitet den übergebenen ab, Werte, Abbruch und Frist bleiben (SPEC-230
// Nr. 10). Menge: 6 Eigenschaften (zwei Werte; Abbruch nach dem Aufruf, mit Ursache; Abbruch vor dem
// Aufruf; Frist als Zeitpunkt; Frist läuft ab; alle drei zugleich) × 2 Aufrufstellen (ohne Span im
// ctx, mit dem Span des Elternaufrufs im ctx) = 12 Fälle; Orakel sind die Eigenschaften des
// Eingabe-ctx.
func TestFakeCtx(t *testing.T) {
	frist, ursache := time.Now().Add(time.Hour), errors.New("ursache")
	warte := func(c context.Context) string {
		select {
		case <-c.Done():
			return ""
		case <-time.After(2 * time.Second):
			return "Done nicht geschlossen"
		}
	}
	werte := func() context.Context {
		return context.WithValue(context.WithValue(context.Background(), ctxSchluessel{}, "wert"), ctxZweiter{}, "zwei")
	}
	eigenschaften := []struct {
		name    string
		pruefen func(start func(context.Context) context.Context) string
	}{
		{"Werte", func(start func(context.Context) context.Context) string {
			if k := start(werte()); k.Value(ctxSchluessel{}) != "wert" || k.Value(ctxZweiter{}) != "zwei" {
				return "ein Wert des Eingabe-ctx fehlt"
			}
			return ""
		}},
		{"Abbruch danach", func(start func(context.Context) context.Context) string {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			k := start(ctx)
			if k.Err() != nil {
				return "vor dem Abbruch schon abgebrochen"
			}
			cancel(ursache)
			if m := warte(k); m != "" || !errors.Is(k.Err(), context.Canceled) || context.Cause(k) != ursache {
				return "der Abbruch (Fehler, Ursache) erreicht den Rückgabe-ctx nicht: " + m
			}
			return ""
		}},
		{"Abbruch davor", func(start func(context.Context) context.Context) string {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if start(ctx).Err() == nil {
				return "der Abbruch vor dem Aufruf geht verloren"
			}
			return ""
		}},
		{"Frist als Zeitpunkt", func(start func(context.Context) context.Context) string {
			ctx, cancel := context.WithDeadline(context.Background(), frist)
			defer cancel()
			if d, ok := start(ctx).Deadline(); !ok || !d.Equal(frist) {
				return "die Frist des Eingabe-ctx fehlt oder weicht ab"
			}
			return ""
		}},
		{"Frist läuft ab", func(start func(context.Context) context.Context) string {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			k := start(ctx)
			if m := warte(k); m != "" || !errors.Is(k.Err(), context.DeadlineExceeded) {
				return "der Ablauf der Frist erreicht den Rückgabe-ctx nicht: " + m
			}
			return ""
		}},
		{"alle drei", func(start func(context.Context) context.Context) string {
			ctx, cancel := context.WithDeadline(werte(), frist)
			defer cancel()
			k := start(ctx)
			if d, ok := k.Deadline(); !ok || !d.Equal(frist) || k.Value(ctxZweiter{}) != "zwei" || k.Err() != nil {
				return "Frist, Wert oder Zustand weichen ab"
			}
			cancel()
			if k.Err() == nil {
				return "der Abbruch geht verloren"
			}
			return ""
		}},
	}
	stellen := []struct {
		name  string
		start func(f *supporttest.FakeTelemetry, ctx context.Context) context.Context
	}{
		{"ohne Span im ctx", func(f *supporttest.FakeTelemetry, ctx context.Context) context.Context {
			k, _ := f.StartSpan(ctx, support.SpanMockQuery)
			return k
		}},
		{"mit Span im ctx", func(f *supporttest.FakeTelemetry, ctx context.Context) context.Context {
			e, _ := f.StartSpan(ctx, support.SpanGatewayRequest)
			k, _ := f.StartSpan(e, support.SpanReplayMatch)
			return k
		}},
	}
	faelle := 0
	for _, e := range eigenschaften {
		for _, s := range stellen {
			faelle++
			f := supporttest.NewFakeTelemetry()
			if m := e.pruefen(func(ctx context.Context) context.Context { return s.start(f, ctx) }); m != "" {
				t.Errorf("%s, %s: %s", e.name, s.name, m)
			}
		}
	}
	if len(eigenschaften) != 6 || len(stellen) != 2 || faelle != 12 {
		t.Errorf("%d Eigenschaften, %d Stellen, %d Fälle; erwartet 6, 2, 12", len(eigenschaften), len(stellen), faelle)
	}
}

// vollerStand baut bei jedem Aufruf denselben Stand, in dem alle sechs Sammlungen gefüllt sind.
func vollerStand() support.Snapshot {
	return support.Snapshot{
		Errors: map[model.Code]uint64{model.CodeE2001: 1, model.CodeE2101: 2},
		Durations: map[support.SpanName]support.Verteilung{
			support.SpanMockQuery:   {Count: 3, Sum: 5, Counts: []uint64{1, 2, 0}},
			support.SpanReplayMatch: {Count: 1, Sum: 2, Counts: []uint64{0, 1, 0}},
		},
		Dropped: map[support.Verwurf]uint64{support.VerwurfWert: 1, support.VerwurfSenke: 2},
		Spans: []support.SpanRecord{
			{ID: "a", Attrs: []support.Attr{support.Str("k", "v"), support.Int("n", 1)}},
			{ID: "b", Attrs: []support.Attr{support.Str("k", "w")}},
		},
	}
}

// aendereSammlung ändert jedes Element der Sammlung Nr. i (Errors, Durations, Dropped, Spans,
// Counts, Attrs) und keine andere.
func aendereSammlung(s *support.Snapshot, i int) {
	switch i {
	case 0:
		for k := range s.Errors {
			s.Errors[k]++
		}
	case 1:
		for k, v := range s.Durations {
			v.Count++
			v.Sum++
			s.Durations[k] = v
		}
	case 2:
		for k := range s.Dropped {
			s.Dropped[k]++
		}
	case 3:
		for j := range s.Spans {
			s.Spans[j].ID += "x"
		}
	case 4:
		for _, v := range s.Durations {
			for j := range v.Counts {
				v.Counts[j]++
			}
		}
	case 5:
		for j := range s.Spans {
			for l := range s.Spans[j].Attrs {
				s.Spans[j].Attrs[l].Ganz++
				s.Spans[j].Attrs[l].Text += "x"
			}
		}
	}
}

// P11: SetzeStand und Snapshot kopieren (SPEC-230 Nr. 10). Menge: 6 Sammlungen (Errors, Durations,
// Dropped, Spans, Counts, Attrs) × 2 Richtungen (die Änderung des Eingabestands nach SetzeStand, die
// Änderung des gelesenen Stands) = 12 Fälle, je Fall wird nur eine Sammlung geändert; Orakel ist ein
// frisch gebauter Stand.
func TestFakeStandKopie(t *testing.T) {
	faelle := 0
	for i := range 6 {
		for _, richtung := range []string{"nach SetzeStand", "nach Snapshot"} {
			faelle++
			f := supporttest.NewFakeTelemetry()
			eingabe := vollerStand()
			f.SetzeStand(eingabe)
			if richtung == "nach SetzeStand" {
				aendereSammlung(&eingabe, i)
			} else {
				gelesen := f.Snapshot()
				aendereSammlung(&gelesen, i)
			}
			if got := f.Snapshot(); !reflect.DeepEqual(got, vollerStand()) {
				t.Errorf("Sammlung %d, Änderung %s: Stand %+v weicht ab", i, richtung, got)
			}
		}
	}
	if faelle != 12 {
		t.Errorf("%d Fälle, erwartet 12", faelle)
	}
}

// P11: Attributlisten jeder Länge von 0 bis 64 werden bei Log, StartSpan und Set ungeprüft und
// vollständig festgehalten (3 × 65 Aufrufe). Über 64 Attribute ist die Menge unbeschränkt; der Fake
// hat keine Grenze im Code (Negativ im Plan).
func TestFakeLaengen(t *testing.T) {
	const bis = 64
	f := supporttest.NewFakeTelemetry()
	ctx := context.Background()
	_, span := f.StartSpan(ctx, "start")
	var listen [][]support.Attr
	for n := 0; n <= bis; n++ {
		l := make([]support.Attr, n)
		for i := range l {
			l[i] = support.Int(fmt.Sprint("k", i), int64(i))
		}
		listen = append(listen, l)
		f.Log(ctx, slog.LevelInfo, "m", l...)
		f.StartSpan(ctx, support.SpanMockQuery, l...)
		span.Set(l...)
	}
	a := f.Aufrufe()
	if len(a) != 1+3*(bis+1) {
		t.Fatalf("%d Aufrufe, erwartet %d", len(a), 1+3*(bis+1))
	}
	for n, l := range listen {
		for k, op := range []supporttest.Operation{supporttest.OpLog, supporttest.OpStartSpan, supporttest.OpSet} {
			if g := a[1+3*n+k]; g.Op != op || !attrGleich(g.Attrs, l) {
				t.Errorf("Länge %d, Operation %d: %d Attribute festgehalten", n, op, len(g.Attrs))
			}
		}
	}
}

// P11: Anzahl liefert für jede Operation außerhalb der sechs 0, ohne Panik: die Werte unter 0, 0, die
// Werte über der letzten Operation und die Grenzen von int.
func TestFakeAnzahl(t *testing.T) {
	f := supporttest.NewFakeTelemetry()
	f.CountError(model.CodeE2001)
	f.Snapshot()
	aussen := []supporttest.Operation{math.MinInt, -100, -2, -1, 0, supporttest.OpSnapshot + 1, supporttest.OpSnapshot + 2, 100, math.MaxInt}
	for _, op := range aussen {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Anzahl(%d) gerät in Panik: %v", op, r)
				}
			}()
			if n := f.Anzahl(op); n != 0 {
				t.Errorf("Anzahl(%d) = %d, erwartet 0", op, n)
			}
		}()
	}
	if f.Anzahl(supporttest.OpCountError) != 1 || f.Anzahl(supporttest.OpSnapshot) != 1 {
		t.Error("Anzahl der Operationen innerhalb falsch")
	}
}

// P11: jede der sechs Operationen hält den Span fest, den sie betrifft, und nur sie: zwei Spans,
// deren Enden in umgekehrter Reihenfolge stehen.
func TestFakeSpanJeOperation(t *testing.T) {
	f := supporttest.NewFakeTelemetry()
	_, s1 := f.StartSpan(context.Background(), support.SpanMockQuery)
	_, s2 := f.StartSpan(context.Background(), support.SpanReplayMatch)
	f.Log(context.Background(), slog.LevelInfo, "m")
	s2.End()
	s1.Set(support.Str("a", "b"))
	s1.End()
	s2.Set(support.Str("a", "b"))
	f.CountError(model.CodeE2001)
	f.Snapshot()
	want := []struct {
		op   supporttest.Operation
		span support.Span
	}{
		{supporttest.OpStartSpan, s1}, {supporttest.OpStartSpan, s2}, {supporttest.OpLog, nil},
		{supporttest.OpEnd, s2}, {supporttest.OpSet, s1}, {supporttest.OpEnd, s1}, {supporttest.OpSet, s2},
		{supporttest.OpCountError, nil}, {supporttest.OpSnapshot, nil},
	}
	a := f.Aufrufe()
	if len(a) != len(want) {
		t.Fatalf("%d Aufrufe, erwartet %d", len(a), len(want))
	}
	ops := map[supporttest.Operation]bool{}
	for i, w := range want {
		ops[w.op] = true
		got := a[i]
		if got.Op != w.op {
			t.Errorf("Aufruf %d: Operation %d, erwartet %d", i, got.Op, w.op)
		}
		if w.span == nil && got.Span != nil {
			t.Errorf("Aufruf %d (Operation %d): Span %v, erwartet keiner", i, got.Op, got.Span)
		}
		if w.span != nil && (got.Span == nil || support.Span(got.Span) != w.span) {
			t.Errorf("Aufruf %d (Operation %d): Span %v, erwartet %v", i, got.Op, got.Span, w.span)
		}
	}
	if len(ops) != 6 {
		t.Errorf("%d Operationen geprüft, erwartet 6", len(ops))
	}
}

// Die ID eines Spans zählt je Fake ab 1.
func TestFakeSpanID(t *testing.T) {
	for range 2 {
		f := supporttest.NewFakeTelemetry()
		for want := 1; want <= 3; want++ {
			_, s := f.StartSpan(context.Background(), support.SpanMockQuery)
			if got := s.(*supporttest.FakeSpan).ID; got != want {
				t.Errorf("ID %d, erwartet %d", got, want)
			}
		}
	}
}

// Der Nullwert ist nutzbar: jede der sechs Operationen, ohne Panik, und gezählt.
func TestFakeNullwert(t *testing.T) {
	f := &supporttest.FakeTelemetry{}
	ctx, s := f.StartSpan(context.Background(), support.SpanMockQuery)
	f.Log(ctx, slog.LevelInfo, "m")
	f.CountError(model.CodeE2001)
	s.Set()
	s.End()
	leererStand(t, f.Snapshot())
	for op := supporttest.OpLog; op <= supporttest.OpSnapshot; op++ {
		if f.Anzahl(op) != 1 {
			t.Errorf("Anzahl(%d) = %d, erwartet 1", op, f.Anzahl(op))
		}
	}
	if f.Anzahl(0) != 0 || f.Anzahl(supporttest.OpSnapshot+1) != 0 {
		t.Error("Operation außerhalb der sechs gezählt")
	}
}

// P16: SetzeStand und Snapshot gleichzeitig; 8 Schreiber setzen je 500 Stände (Errors und Dropped
// tragen denselben Zähler n, Spans und Counts sind gefüllt), 8 Leser lesen je 500 Stände und
// prüfen, dass jeder ganz ist (beide Zähler gleich, Counts und Spans in der Länge von n).
func TestFakeSnapshotGleichzeitig(t *testing.T) {
	const schreiber, leser, runden = 8, 8, 500
	f := supporttest.NewFakeTelemetry()
	start := make(chan struct{})
	fehler := make(chan string, leser*runden)
	var wg sync.WaitGroup
	for g := range schreiber {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := range runden {
				n := uint64(g*runden + i)
				f.SetzeStand(support.Snapshot{
					Errors:    map[model.Code]uint64{model.CodeE2001: n},
					Durations: map[support.SpanName]support.Verteilung{support.SpanMockQuery: {Count: n, Counts: []uint64{n, n}}},
					Dropped:   map[support.Verwurf]uint64{support.VerwurfWert: n},
					Spans:     []support.SpanRecord{{ID: "x", Attrs: []support.Attr{support.Int("n", int64(n))}}},
				})
			}
		}()
	}
	for range leser {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range runden {
				s := f.Snapshot()
				if len(s.Errors) == 0 {
					continue
				}
				n := s.Errors[model.CodeE2001]
				d := s.Durations[support.SpanMockQuery]
				if s.Dropped[support.VerwurfWert] != n || d.Count != n || len(d.Counts) != 2 || d.Counts[0] != n ||
					len(s.Spans) != 1 || len(s.Spans[0].Attrs) != 1 || s.Spans[0].Attrs[0].Ganz != int64(n) {
					fehler <- "ein Stand ist nicht ganz"
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(fehler)
	for m := range fehler {
		t.Error(m)
	}
	if f.Anzahl(supporttest.OpSnapshot) != leser*runden {
		t.Errorf("Snapshot %d-mal gezählt, erwartet %d", f.Anzahl(supporttest.OpSnapshot), leser*runden)
	}
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
