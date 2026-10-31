package support

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Telemetry gibt Logs, Spans und Zähler aus (SPEC-230); keine Operation liefert einen Fehler oder
// gerät in Panik. Lesen gehört nicht zu diesem Port, sondern zu TelemetryReader.
type Telemetry interface {
	Log(ctx context.Context, level slog.Level, msg string, attrs ...Attr)
	StartSpan(ctx context.Context, name SpanName, attrs ...Attr) (context.Context, Span)
	CountError(code model.Code)
}

// Span ist ein gestarteter Span (SPEC-230 Nr. 5 bis 7).
type Span interface {
	Set(attrs ...Attr)
	End()
}

// TelemetryReader liest den Stand der Zähler, Verteilungen und Spans (SPEC-230 Nr. 9).
type TelemetryReader interface {
	Snapshot() Snapshot
}

// Art ist die Art des Werts eines Attributs (SPEC-230 Nr. 1).
type Art int

// Die drei Arten; jeder andere Wert von Art ist ein Verstoß gegen die Form.
const (
	ArtString Art = iota + 1
	ArtInt
	ArtFloat
)

// Attr trägt einen Schlüssel und einen Wert genau einer Art: Text bei ArtString, Ganz bei
// ArtInt, Gleit bei ArtFloat. Die Felder der anderen Arten sind ohne Bedeutung.
type Attr struct {
	Key   string
	Art   Art
	Text  string
	Ganz  int64
	Gleit float64
}

// Str bildet ein Attribut der Art String.
func Str(key, v string) Attr { return Attr{Key: key, Art: ArtString, Text: v} }

// Int bildet ein Attribut der Art Int.
func Int(key string, v int64) Attr { return Attr{Key: key, Art: ArtInt, Ganz: v} }

// Millis bildet ein Attribut der Art Float in Millisekunden.
func Millis(key string, d time.Duration) Attr {
	return Attr{Key: key, Art: ArtFloat, Gleit: float64(d) / float64(time.Millisecond)}
}

// ErrorCode bildet ein Attribut der Art String mit dem Text von c; ein Code außerhalb der Tabelle
// ergibt den leeren Text.
func ErrorCode(key string, c model.Code) Attr {
	return Attr{Key: key, Art: ArtString, Text: c.String()}
}

// SpanName ist der Name eines Spans; die Menge der Namen ist geschlossen (SPEC-230 Nr. 5).
type SpanName string

// Die fünf Spans aus §5 der Spezifikation.
const (
	SpanGatewayRequest    SpanName = "odata.gateway.request"
	SpanRecordCapture     SpanName = "odata.record.capture"
	SpanReplayMatch       SpanName = "odata.replay.match"
	SpanMockQuery         SpanName = "odata.mock.query"
	SpanWorkspaceMutation SpanName = "odata.workspace.mutation"
)

// SpanSpec ist ein Span mit seinen Pflicht-Attributen in der Reihenfolge der Tabelle in §5.
type SpanSpec struct {
	Name    SpanName
	Pflicht []string
}

// Spans liefert die fünf Spans mit zusammen elf Pflicht-Attributen; jeder Aufruf liefert frische
// Slices.
func Spans() []SpanSpec {
	return []SpanSpec{
		{SpanGatewayRequest, []string{"http.request.method", "http.response.status_code", "recorder.mode"}},
		{SpanRecordCapture, []string{"recorder.session_id", "recorder.capture_status"}},
		{SpanReplayMatch, []string{"recorder.match_strategy", "recorder.match_result"}},
		{SpanMockQuery, []string{"recorder.entity_set", "recorder.coverage_status"}},
		{SpanWorkspaceMutation, []string{"recorder.workspace_id", "recorder.mutation_type"}},
	}
}

// Eimer liefert die elf oberen Grenzen der Verteilung (1, 2, 5, 10, 20, 50, 100, 200, 500, 1000 und
// 5000 ms); der zwölfte Eimer nimmt jede längere Dauer (SPEC-230 Nr. 8). Jeder Aufruf liefert
// einen frischen Slice.
func Eimer() []time.Duration {
	ms := []time.Duration{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 5000}
	for i := range ms {
		ms[i] *= time.Millisecond
	}
	return ms
}

// Verwurf ist der Grund, aus dem Telemetrie ein Attribut oder eine Meldung verwirft (SPEC-230 Nr. 4).
type Verwurf string

// Die sieben Gründe.
const (
	VerwurfSenke      Verwurf = "senke"
	VerwurfSchluessel Verwurf = "schluessel"
	VerwurfWert       Verwurf = "wert"
	VerwurfPflicht    Verwurf = "pflicht"
	VerwurfSpanname   Verwurf = "spanname"
	VerwurfZustand    Verwurf = "zustand"
	VerwurfFehlercode Verwurf = "fehlercode"
)

// Verwuerfe liefert die sieben Gründe in der Reihenfolge von SPEC-230 Nr. 4.
func Verwuerfe() []Verwurf {
	return []Verwurf{
		VerwurfSenke, VerwurfSchluessel, VerwurfWert, VerwurfPflicht,
		VerwurfSpanname, VerwurfZustand, VerwurfFehlercode,
	}
}

// Verteilung ist die Dauerverteilung eines Span-Namens: Count Enden, Sum ihre Summe, Counts je
// Eimer (SPEC-230 Nr. 8).
type Verteilung struct {
	Count  uint64
	Sum    time.Duration
	Counts []uint64
}

// SpanRecord ist ein beendeter Span im Ring (SPEC-230 Nr. 9).
type SpanRecord struct {
	ID, ParentID string
	Name         SpanName
	Start        time.Time
	Duration     time.Duration
	Attrs        []Attr
}

// Snapshot ist ein Stand zu einem Zeitpunkt (SPEC-230 Nr. 9).
type Snapshot struct {
	Errors    map[model.Code]uint64
	Durations map[SpanName]Verteilung
	Dropped   map[Verwurf]uint64
	Spans     []SpanRecord
}

// Kopie liefert einen Stand, der mit s keinen Speicher teilt; eine nil-Map oder ein nil-Slice
// bleibt nil.
func (s Snapshot) Kopie() Snapshot {
	k := Snapshot{
		Errors:  maps.Clone(s.Errors),
		Dropped: maps.Clone(s.Dropped),
		Spans:   slices.Clone(s.Spans),
	}
	if s.Durations != nil {
		k.Durations = make(map[SpanName]Verteilung, len(s.Durations))
		for n, v := range s.Durations {
			v.Counts = slices.Clone(v.Counts)
			k.Durations[n] = v
		}
	}
	for i := range k.Spans {
		k.Spans[i].Attrs = slices.Clone(k.Spans[i].Attrs)
	}
	return k
}
