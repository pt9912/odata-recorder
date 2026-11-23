package record_test

import (
	"context"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

// Schlüssel und Platzhalter der Pflicht-Attribute von odata.record.capture stehen als Literal im Test (SPEC-027, SPEC-295 Nr. 4);
// TestPflichtSchluesselCapture hält die Schlüssel gleich der Tabelle support.Spans().
const (
	attrSession = "recorder.session_id"
	attrStatus  = "recorder.capture_status"
	unbestimmt  = "unbestimmt"
)

// ctxAufruf und spanMarke sind die Werte, an denen ein Test das ctx von Record und das ctx von StartSpan erkennt: jeder Port sieht
// jenen und keinen dieser (SPEC-295 Nr. 3 d).
type (
	ctxAufruf struct{}
	spanMarke struct{}
)

// journal ist der Journal-Fake von Telemetry (SPEC-295): FakeTelemetry hält jedes Argument fest und zählt jede Operation; dazu schreibt
// jede Operation und jeder Aufruf der umhüllten Ports in das gemeinsame Protokoll prot. vorEnd läuft vor dem Buch eines End und
// verzögert es (eine Senke, die langsam schreibt, SPEC-295 Nr. 9).
type journal struct {
	*supporttest.FakeTelemetry
	prot   *ereignisse
	vorEnd func()
}

func (j *journal) Log(ctx context.Context, level slog.Level, msg string, attrs ...support.Attr) {
	j.prot.melde("Telemetrie.Log")
	j.FakeTelemetry.Log(ctx, level, msg, attrs...)
}

func (j *journal) CountError(code model.Code) {
	j.prot.melde("Telemetrie.CountError")
	j.FakeTelemetry.CountError(code)
}

// StartSpan liefert ein ctx mit der spanMarke; ein Port, der es erhält, ist an der Marke erkennbar.
func (j *journal) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	j.prot.melde("StartSpan")
	c, s := j.FakeTelemetry.StartSpan(ctx, name, attrs...)
	return context.WithValue(c, spanMarke{}, true), &spanJournal{Span: s, j: j}
}

type spanJournal struct {
	support.Span
	j *journal
}

func (s *spanJournal) Set(attrs ...support.Attr) {
	s.j.prot.melde("Set")
	s.Span.Set(attrs...)
}

func (s *spanJournal) End() {
	if s.j.vorEnd != nil {
		s.j.vorEnd()
	}
	s.j.prot.melde("End")
	s.Span.End()
}

// Die Hüllen der Ports schreiben jeden Aufruf in das Protokoll und reichen ihn unverändert weiter.
type (
	jSessions struct {
		outbound.SessionRepository
		p *ereignisse
	}
	jIDs struct {
		support.IDGenerator
		p *ereignisse
	}
	jUhr struct {
		support.Clock
		p *ereignisse
	}
	jUpstream struct {
		outbound.UpstreamClient
		p *ereignisse
	}
	jExchanges struct {
		outbound.RecordingRepository
		p *ereignisse
	}
	jLog struct {
		slog.Handler
		p *ereignisse
	}
)

func (s jSessions) Get(ctx context.Context, id string) (model.Session, error) {
	s.p.melde("Sessions.Get")
	return s.SessionRepository.Get(ctx, id)
}

func (s jSessions) List(ctx context.Context, dienst string) ([]model.Session, error) {
	s.p.melde("Sessions.List")
	return s.SessionRepository.List(ctx, dienst)
}

func (g jIDs) ExchangeID() string {
	g.p.melde("ExchangeID")
	return g.IDGenerator.ExchangeID()
}

func (u jUhr) Now() time.Time {
	u.p.melde("Now")
	return u.Clock.Now()
}

func (u jUpstream) Senden(ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	u.p.melde("Senden")
	return u.UpstreamClient.Senden(ctx, a)
}

func (e jExchanges) Append(ctx context.Context, x model.Exchange) error {
	e.p.melde("Append")
	return e.RecordingRepository.Append(ctx, x)
}

func (e jExchanges) ListBySession(ctx context.Context, id string) ([]model.Exchange, error) {
	e.p.melde("ListBySession")
	return e.RecordingRepository.ListBySession(ctx, id)
}

func (l jLog) Handle(ctx context.Context, r slog.Record) error {
	l.p.melde("Log " + r.Message)
	return l.Handler.Handle(ctx, r)
}

// neuMitJournal baut die Umgebung mit dem Journal-Fake als Telemetrie und mit Hüllen um Sessions, IDs, Uhr, Upstream, Exchanges und Log,
// die alle in dasselbe Protokoll schreiben.
func neuMitJournal(t *testing.T, aendere ...func(*record.Konfig)) (*umgebung, *journal) {
	t.Helper()
	j := &journal{FakeTelemetry: supporttest.NewFakeTelemetry(), prot: &ereignisse{}}
	u := neu(t, append([]func(*record.Konfig){func(k *record.Konfig) {
		k.Telemetrie, k.IDs, k.Uhr = j, jIDs{k.IDs, j.prot}, jUhr{k.Uhr, j.prot}
		k.Sessions, k.Upstream, k.Exchanges = jSessions{k.Sessions, j.prot}, jUpstream{k.Upstream, j.prot}, jExchanges{k.Exchanges, j.prot}
		k.Log = slog.New(jLog{k.Log.Handler(), j.prot})
	}}, aendere...)...)
	u.tel = j.FakeTelemetry
	return u, j
}

// folge liefert die Einträge des Protokolls in ihrer Folge, gefiltert auf die Namen in nur.
func (j *journal) folge(nur ...string) []string {
	var r []string
	for _, e := range j.prot.liste() {
		if slices.Contains(nur, e) {
			r = append(r, e)
		}
	}
	return r
}

// spanBuch ist das Buch über einen Span des Fakes: das ctx von StartSpan, die Attribute dort, jeder Aufruf von Set, die Zahl der End und die
// Folge der Aufrufe an genau diesem Span (`Set`, `End`).
type spanBuch struct {
	name  support.SpanName
	ctx   func() context.Context
	start []support.Attr
	setze [][]support.Attr
	enden int
	folge []string
}

// buecher bildet je Span des Fakes ein Buch, in der Folge der StartSpan.
func buecher(f *supporttest.FakeTelemetry) []spanBuch {
	var r []spanBuch
	index := map[*supporttest.FakeSpan]int{}
	for _, a := range f.Aufrufe() {
		switch a.Op {
		case supporttest.OpStartSpan:
			index[a.Span] = len(r)
			r = append(r, spanBuch{name: a.Name, ctx: a.Ctx, start: slices.Clone(a.Attrs)})
		case supporttest.OpSet:
			if i, da := index[a.Span]; da {
				r[i].setze = append(r[i].setze, slices.Clone(a.Attrs))
				r[i].folge = append(r[i].folge, "Set")
			}
		case supporttest.OpEnd:
			if i, da := index[a.Span]; da {
				r[i].enden++
				r[i].folge = append(r[i].folge, "End")
			}
		}
	}
	return r
}

// pruefeBuch prüft das Buch eines Spans: Name, die Platzhalter bei StartSpan und am eigenen Span genau ein Set mit genau dem Status, gefolgt von genau
// einem End (SPEC-295 Nr. 4 b, 3 b; V-1: kein Set hinter dem End, kein End an einem fremden Span). Jeder Test, der einen Exchange führt, ruft es je Span.
func pruefeBuch(t *testing.T, was string, b spanBuch, sessionID, status string) {
	t.Helper()
	wantStart := []support.Attr{support.Str(attrSession, sessionID), support.Str(attrStatus, unbestimmt)}
	wantSet := [][]support.Attr{{support.Str(attrStatus, status)}}
	wantFolge := []string{"Set", "End"}
	if b.name != support.SpanRecordCapture || b.enden != 1 || !slices.Equal(b.start, wantStart) || len(b.setze) != 1 || !slices.Equal(b.setze[0], wantSet[0]) ||
		!slices.Equal(b.folge, wantFolge) {
		t.Errorf("%s: Buch %+v, erwartet %s mit Start %v, Set %v und einem End in der Folge %v", was, b, support.SpanRecordCapture, wantStart, wantSet, wantFolge)
	}
}

// TestPflichtSchluesselCapture hält die Schlüssel des Tests gleich der Tabelle der Pflicht-Attribute (SPEC-230 Nr. 5).
func TestPflichtSchluesselCapture(t *testing.T) {
	for _, s := range support.Spans() {
		if s.Name == support.SpanRecordCapture {
			if !slices.Equal(s.Pflicht, []string{attrSession, attrStatus}) {
				t.Errorf("Pflicht-Attribute %v, erwartet %s und %s", s.Pflicht, attrSession, attrStatus)
			}
			return
		}
	}
	t.Errorf("kein Span %s in der Tabelle", support.SpanRecordCapture)
}
