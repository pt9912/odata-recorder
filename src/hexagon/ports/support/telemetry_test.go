package support_test

import (
	"context"
	"log/slog"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// P01: die vier Konstruktoren setzen Schlüssel, Art und Wert.
func TestAttrKonstruktoren(t *testing.T) {
	faelle := []struct {
		name string
		got  support.Attr
		want support.Attr
	}{
		{"Str", support.Str("k", "v"), support.Attr{Key: "k", Art: support.ArtString, Text: "v"}},
		{"Int", support.Int("k", -7), support.Attr{Key: "k", Art: support.ArtInt, Ganz: -7}},
		{"Millis", support.Millis("k", 1500*time.Microsecond), support.Attr{Key: "k", Art: support.ArtFloat, Gleit: 1.5}},
		{"ErrorCode", support.ErrorCode("k", model.CodeE2401), support.Attr{Key: "k", Art: support.ArtString, Text: "ODR-E2401"}},
	}
	if len(faelle) != 4 {
		t.Fatalf("%d Konstruktoren, erwartet 4", len(faelle))
	}
	for _, f := range faelle {
		if f.got != f.want {
			t.Errorf("%s: %+v, erwartet %+v", f.name, f.got, f.want)
		}
	}
	for _, c := range []model.Code{0, -1, model.CodeE9001 + 1} {
		if a := support.ErrorCode("k", c); a != (support.Attr{Key: "k", Art: support.ArtString}) {
			t.Errorf("ErrorCode(%d) = %+v, erwartet die Art String mit leerem Text", int(c), a)
		}
	}
	if a := support.Millis("k", -time.Nanosecond); a.Gleit >= 0 || math.IsNaN(a.Gleit) {
		t.Errorf("Millis(-1ns) = %v, erwartet eine negative Zahl", a.Gleit)
	}
}

// Die Spans und ihre Pflicht-Attribute aus §5 der Spezifikation, aus dem Wortlaut der Tabelle.
func TestSpansDaten(t *testing.T) {
	want := []support.SpanSpec{
		{Name: "odata.gateway.request", Pflicht: []string{"http.request.method", "http.response.status_code", "recorder.mode"}},
		{Name: "odata.record.capture", Pflicht: []string{"recorder.session_id", "recorder.capture_status"}},
		{Name: "odata.replay.match", Pflicht: []string{"recorder.match_strategy", "recorder.match_result"}},
		{Name: "odata.mock.query", Pflicht: []string{"recorder.entity_set", "recorder.coverage_status"}},
		{Name: "odata.workspace.mutation", Pflicht: []string{"recorder.workspace_id", "recorder.mutation_type"}},
	}
	got := support.Spans()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Spans() = %+v, erwartet %+v", got, want)
	}
	elf := 0
	for _, s := range got {
		elf += len(s.Pflicht)
	}
	if len(got) != 5 || elf != 11 {
		t.Errorf("%d Spans mit %d Pflicht-Attributen, erwartet 5 und 11", len(got), elf)
	}
	konstanten := []support.SpanName{
		support.SpanGatewayRequest, support.SpanRecordCapture, support.SpanReplayMatch,
		support.SpanMockQuery, support.SpanWorkspaceMutation,
	}
	for i, k := range konstanten {
		if k != want[i].Name {
			t.Errorf("Konstante %d = %q, erwartet %q", i, k, want[i].Name)
		}
	}
	got[0].Pflicht[0] = "geändert"
	if support.Spans()[0].Pflicht[0] != "http.request.method" {
		t.Error("Spans() teilt Speicher zwischen Aufrufen")
	}
}

func TestEimerGrenzen(t *testing.T) {
	want := []time.Duration{
		1 * time.Millisecond, 2 * time.Millisecond, 5 * time.Millisecond, 10 * time.Millisecond,
		20 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond,
		500 * time.Millisecond, 1000 * time.Millisecond, 5000 * time.Millisecond,
	}
	got := support.Eimer()
	if !slices.Equal(got, want) || len(got) != 11 {
		t.Errorf("Eimer() = %v, erwartet %v (11 Grenzen)", got, want)
	}
	got[0] = 0
	if support.Eimer()[0] != time.Millisecond {
		t.Error("Eimer() teilt Speicher zwischen Aufrufen")
	}
}

func TestVerwuerfe(t *testing.T) {
	want := []support.Verwurf{"senke", "schluessel", "wert", "pflicht", "spanname", "zustand", "fehlercode"}
	if got := support.Verwuerfe(); !slices.Equal(got, want) || len(got) != 7 {
		t.Errorf("Verwuerfe() = %v, erwartet %v", got, want)
	}
}

// P13: Telemetry hat 3 Methoden, Span 2, TelemetryReader 1; kein Lesen in Telemetry; der Fake
// erfüllt beide Ports (der Adapter in seinem Paket).
func TestTelemetrySignaturen(t *testing.T) {
	ctxT := reflect.TypeFor[context.Context]()
	attrs := reflect.TypeFor[[]support.Attr]()
	formen := []struct {
		typ     reflect.Type
		methode string
		form    reflect.Type
	}{
		{reflect.TypeFor[support.Telemetry](), "Log", reflect.FuncOf([]reflect.Type{ctxT, reflect.TypeFor[slog.Level](), reflect.TypeFor[string](), attrs}, nil, true)},
		{reflect.TypeFor[support.Telemetry](), "StartSpan", reflect.FuncOf([]reflect.Type{ctxT, reflect.TypeFor[support.SpanName](), attrs}, []reflect.Type{ctxT, reflect.TypeFor[support.Span]()}, true)},
		{reflect.TypeFor[support.Telemetry](), "CountError", reflect.TypeFor[func(model.Code)]()},
		{reflect.TypeFor[support.Span](), "Set", reflect.FuncOf([]reflect.Type{attrs}, nil, true)},
		{reflect.TypeFor[support.Span](), "End", reflect.TypeFor[func()]()},
		{reflect.TypeFor[support.TelemetryReader](), "Snapshot", reflect.TypeFor[func() support.Snapshot]()},
	}
	zahl := map[reflect.Type]int{
		reflect.TypeFor[support.Telemetry]():       3,
		reflect.TypeFor[support.Span]():            2,
		reflect.TypeFor[support.TelemetryReader](): 1,
	}
	if len(zahl) != 3 {
		t.Fatalf("%d Schnittstellen geprüft, erwartet 3", len(zahl))
	}
	for typ, n := range zahl {
		if typ.NumMethod() != n {
			t.Errorf("%v hat %d Methoden, erwartet %d", typ, typ.NumMethod(), n)
		}
	}
	for _, f := range formen {
		m, ok := f.typ.MethodByName(f.methode)
		if !ok || m.Type != f.form {
			t.Errorf("%v.%s hat die Form %v, erwartet %v", f.typ, f.methode, m.Type, f.form)
		}
	}
	if _, ok := reflect.TypeFor[support.Telemetry]().MethodByName("Snapshot"); ok {
		t.Error("Telemetry hat eine Leseoperation")
	}
	fake := supporttest.NewFakeTelemetry()
	var (
		_ support.Telemetry       = fake
		_ support.TelemetryReader = fake
	)
}

func TestSnapshotKopie(t *testing.T) {
	s := support.Snapshot{
		Errors:    map[model.Code]uint64{model.CodeE2001: 1},
		Durations: map[support.SpanName]support.Verteilung{support.SpanMockQuery: {Count: 1, Counts: []uint64{1, 0}}},
		Dropped:   map[support.Verwurf]uint64{support.VerwurfWert: 1},
		Spans:     []support.SpanRecord{{ID: "spn-1", Attrs: []support.Attr{support.Str("k", "v")}}},
	}
	k := s.Kopie()
	if !reflect.DeepEqual(s, k) {
		t.Fatalf("Kopie %+v weicht von %+v ab", k, s)
	}
	k.Errors[model.CodeE2001] = 9
	k.Durations[support.SpanMockQuery].Counts[0] = 9
	k.Dropped[support.VerwurfWert] = 9
	k.Spans[0].Attrs[0].Text = "x"
	k.Spans[0].ID = "x"
	if s.Errors[model.CodeE2001] != 1 || s.Durations[support.SpanMockQuery].Counts[0] != 1 ||
		s.Dropped[support.VerwurfWert] != 1 || s.Spans[0].Attrs[0].Text != "v" || s.Spans[0].ID != "spn-1" {
		t.Errorf("Änderung an der Kopie wirkt auf das Original: %+v", s)
	}
	leer := support.Snapshot{
		Errors:    map[model.Code]uint64{},
		Durations: map[support.SpanName]support.Verteilung{},
		Dropped:   map[support.Verwurf]uint64{},
		Spans:     []support.SpanRecord{},
	}
	if z := leer.Kopie(); z.Errors == nil || z.Durations == nil || z.Dropped == nil || z.Spans == nil {
		t.Errorf("Kopie des leeren Stands hat eine nil-Sammlung: %+v", z)
	}
	if z := (support.Snapshot{}).Kopie(); z.Errors != nil || z.Durations != nil || z.Dropped != nil || z.Spans != nil {
		t.Errorf("Kopie des leeren Stands ist nicht leer: %+v", z)
	}
}
