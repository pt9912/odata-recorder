package replay_test

import (
	"context"
	"log/slog"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Schlüssel und Platzhalter der Pflicht-Attribute von odata.replay.match stehen als Literal im Test (SPEC-028, SPEC-296 Nr. 4);
// TestPflichtSchluesselMatch hält sie gleich der Tabelle support.Spans().
const (
	attrStrategie   = "recorder.match_strategy"
	attrErgebnis    = "recorder.match_result"
	nichtGelesen    = "nicht_gelesen"
	nichtGewaehlt   = "nicht_gewaehlt"
	ergTreffer      = "treffer"
	ergKeinTreffer  = "kein_treffer"
	ergErschoepft   = "erschoepft"
	ergAndererKey   = "anderer_key"
	ergNichtReplayf = "nicht_replayfaehig"
)

// journal ist der Journal-Fake von Telemetry (SPEC-296): FakeTelemetry hält jedes Argument fest und zählt jede Operation; dazu
// schreibt jede Operation in das Protokoll der Port-Fakes, damit Ports und Telemetrie in einer Folge stehen, und meldet jeden Aufruf,
// der in f der Transaktion läuft (SPEC-296 Nr. 3 c). Ohne mitJournal bleibt das Protokoll der übrigen Tests unverändert.
type journal struct {
	*supporttest.FakeTelemetry
	u *umg
}

// mitJournal setzt den Journal-Fake als Telemetrie der Konfig; u.tel bleibt der zugrunde liegende Fake.
func (u *umg) mitJournal() *journal {
	j := &journal{FakeTelemetry: u.tel, u: u}
	u.k.Telemetrie = j
	return j
}

func (j *journal) schreibe(ctx ctxMethoden, op string) {
	if j.u.st.aktiv.Load() {
		j.u.t.Errorf("%s in f der Transaktion gerufen", op)
	}
	_ = j.u.p.beginn(op, ctx)
}

func (j *journal) Log(ctx context.Context, level slog.Level, msg string, attrs ...support.Attr) {
	j.schreibe(ctx, "Telemetrie.Log")
	j.FakeTelemetry.Log(ctx, level, msg, attrs...)
}

func (j *journal) CountError(code model.Code) {
	j.schreibe(nil, "Telemetrie.CountError")
	j.FakeTelemetry.CountError(code)
}

func (j *journal) StartSpan(ctx context.Context, name support.SpanName, attrs ...support.Attr) (context.Context, support.Span) {
	j.schreibe(ctx, "Telemetrie.StartSpan")
	c, s := j.FakeTelemetry.StartSpan(ctx, name, attrs...)
	return c, &spanJournal{Span: s, j: j}
}

// spanJournal ist der Span des Journal-Fakes; Set und End erscheinen im Protokoll.
type spanJournal struct {
	support.Span
	j *journal
}

func (s *spanJournal) Set(attrs ...support.Attr) {
	s.j.schreibe(nil, "Span.Set")
	s.Span.Set(attrs...)
}

func (s *spanJournal) End() {
	s.j.schreibe(nil, "Span.End")
	s.Span.End()
}

// spanBuch ist das Buch über einen Span des Fakes: der Name, das ctx von StartSpan, seine Attribute bei StartSpan, jeder Aufruf von Set
// und die Zahl der End.
type spanBuch struct {
	name  support.SpanName
	ctx   func() context.Context
	start []support.Attr
	setze [][]support.Attr
	enden int
}

// endwerte sind Strategie und Ergebnis am Ende des Spans: die Werte des letzten Set, sonst die von StartSpan.
func (b spanBuch) endwerte() (strategie, ergebnis string) {
	alle := slices.Clone(b.start)
	for _, s := range b.setze {
		alle = append(alle, s...)
	}
	for _, a := range slices.Backward(alle) {
		switch {
		case a.Key == attrStrategie && strategie == "":
			strategie = a.Text
		case a.Key == attrErgebnis && ergebnis == "":
			ergebnis = a.Text
		}
	}
	return strategie, ergebnis
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
			}
		case supporttest.OpEnd:
			if i, da := index[a.Span]; da {
				r[i].enden++
			}
		}
	}
	return r
}

// soll ist der Ausgang eines Aufrufs: ein Code, ein Kontextfehler ohne Code oder (Nullwert) die Antwort.
type soll struct {
	code    model.Code
	kontext error
}

func pruefeAusgang(t *testing.T, was string, err error, s soll) {
	t.Helper()
	switch {
	case s.kontext != nil:
		erwarteOhneCode(t, was, err, s.kontext)
	case s.code != 0:
		erwarteCode(t, was, err, s.code)
	case err != nil:
		t.Errorf("%s: Fehler %v, erwartet die Antwort", was, err)
	}
}

// pruefeSpan prüft nach einem Aufruf den Stand des Fakes (SPEC-296 Nr. 2 bis 4, 6): ein Span odata.replay.match mit den Platzhaltern bei
// StartSpan, genau einem End, einem Set mit genau den zwei Attributen in der Folge Strategie, Ergebnis genau dann, wenn set wahr ist, und
// den Endwerten strategie und ergebnis; kein CountError und kein Snapshot.
func pruefeSpan(t *testing.T, was string, f *supporttest.FakeTelemetry, strategie, ergebnis string, set bool) {
	t.Helper()
	b := buecher(f)
	if len(b) != 1 {
		t.Errorf("%s: Bücher %+v, erwartet ein Span %s mit einem End", was, b, support.SpanReplayMatch)
		return
	}
	pruefeBuch(t, was, b[0], strategie, ergebnis, set)
	for _, op := range []supporttest.Operation{supporttest.OpCountError, supporttest.OpSnapshot} {
		if n := f.Anzahl(op); n != 0 {
			t.Errorf("%s: Operation %d %d-mal gerufen", was, op, n)
		}
	}
}

// pruefeBuch prüft das Buch eines Spans: Name, genau ein End, die Platzhalter bei StartSpan, genau ein Set mit genau den zwei Attributen in der
// Folge Strategie, Ergebnis genau dann, wenn set wahr ist, und die Endwerte strategie und ergebnis.
func pruefeBuch(t *testing.T, was string, b spanBuch, strategie, ergebnis string, set bool) {
	t.Helper()
	if b.name != support.SpanReplayMatch || b.enden != 1 {
		t.Errorf("%s: Buch %+v, erwartet ein Span %s mit einem End", was, b, support.SpanReplayMatch)
		return
	}
	if want := []support.Attr{support.Str(attrStrategie, nichtGelesen), support.Str(attrErgebnis, nichtGewaehlt)}; !reflect.DeepEqual(b.start, want) {
		t.Errorf("%s: StartSpan mit %v, erwartet %v", was, b.start, want)
	}
	sets := 0
	if set {
		sets = 1
		want := []support.Attr{support.Str(attrStrategie, strategie), support.Str(attrErgebnis, ergebnis)}
		if len(b.setze) != 1 || !reflect.DeepEqual(b.setze[0], want) {
			t.Errorf("%s: Set %v, erwartet genau ein Set mit %v", was, b.setze, want)
		}
	}
	if len(b.setze) != sets {
		t.Errorf("%s: %d Set, erwartet %d", was, len(b.setze), sets)
	}
	if s, e := b.endwerte(); s != strategie || e != ergebnis {
		t.Errorf("%s: Endwerte %s und %s, erwartet %s und %s", was, s, e, strategie, ergebnis)
	}
}

// spanFall ist ein Lauf von Replay mit seinem Soll für den Span.
type spanFall struct {
	name      string
	strat     model.ReplayStrategy
	baue      func(u *umg) inbound.Anfrage
	soll      soll
	strategie string
	ergebnis  string
	set       bool
}

// lauf fährt den Fall mit dem Journal-Fake und prüft Ausgang und Span; die Antwort wird geschlossen.
func (f spanFall) lauf(t *testing.T) *umg {
	t.Helper()
	u := neu(t)
	u.mitJournal()
	if f.strat != "" {
		u.k.Strategie = f.strat
	}
	a := f.baue(u)
	resp, err := u.replay(context.Background(), a)
	if err == nil {
		_ = resp.Body.Close()
	}
	pruefeAusgang(t, f.name, err, f.soll)
	pruefeSpan(t, f.name, u.tel, f.strategie, f.ergebnis, f.set)
	return u
}
