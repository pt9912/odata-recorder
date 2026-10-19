// Package outboundtest enthält die Vertragstests der Outbound Ports (SPEC-105): Sie laufen
// gegen jeden Adapter, der den Port erfüllt.
package outboundtest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// Verbindung ist eine frische, leere Persistenz mit den Hilfen, die der Vertrag braucht.
type Verbindung struct {
	Sessions  outbound.SessionRepository
	Recording outbound.RecordingRepository
	// FuegeDienstHinzu legt einen Service an, den Sessions referenzieren dürfen.
	FuegeDienstHinzu func(t *testing.T, id string)
	// MacheZurBaselineQuelle trägt die Session als Quelle einer Baseline ein.
	MacheZurBaselineQuelle func(t *testing.T, sessionID string)
}

// Neu liefert je Aufruf eine frische Verbindung.
type Neu func(t *testing.T) Verbindung

func t0() time.Time { return time.Date(2026, 10, 19, 10, 0, 0, 123456789, time.UTC) }

func session(id, dienst string, minuten int) model.Session {
	return model.Session{ID: id, ServiceID: dienst, CreatedAt: t0().Add(time.Duration(minuten) * time.Minute), Status: model.SessionOpen}
}

func exchange(sessionID, id string, seq int) model.Exchange {
	return model.Exchange{
		ID: id, SessionID: sessionID, SequenceNo: seq, Method: "GET", Path: "/Products", RawQuery: "$top=10",
		MatchKey: "sha256:abc", StartedAt: t0(), FinishedAt: t0().Add(time.Second),
		RequestHeaders: map[string][]string{"Accept": {"application/json"}}, RequestBody: model.BodyAbsent,
		ResponseStatus: 200, ResponseHeaders: map[string][]string{"Content-Type": {"application/json"}},
		ResponseBody: model.BodyPresent, ResponseBlob: "sha256:def", Capture: model.CaptureComplete,
		RequestRedact: model.RedactionNone, ResponseRedact: model.RedactionNone,
		Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte,
	}
}

func erwarte(t *testing.T, err, soll error) {
	t.Helper()
	if !errors.Is(err, soll) {
		t.Fatalf("Fehler %v, erwartet %v", err, soll)
	}
}

func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Fehler %v", err)
	}
}

// RunSessionVertrag prüft SPEC-105 Nr. 1 bis 6.
func RunSessionVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Create und Get", sessionFall1},
		{"Create lehnt Ungültiges, Doppeltes und unbekannten Service ab", sessionFall2},
		{"Get einer unbekannten Kennung", sessionFall3},
		{"SetStatus erlaubt nur open nach complete oder incomplete", sessionFall4},
		{"List und FindOpen ordnen nach Zeit und Kennung", sessionFall5},
		{"Delete nimmt Exchanges mit und achtet Baseline-Quellen", sessionFall6},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func sessionFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	v.FuegeDienstHinzu(t, "d")
	s := session("s1", "d", 0)
	ok(t, v.Sessions.Create(ctx, s))
	got, err := v.Sessions.Get(ctx, "s1")
	ok(t, err)
	if got.ID != s.ID || got.ServiceID != "d" || got.Status != model.SessionOpen || !got.CreatedAt.Equal(s.CreatedAt) {
		t.Fatalf("gelesen %+v, geschrieben %+v", got, s)
	}
}

func sessionFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	v.FuegeDienstHinzu(t, "d")
	for name, s := range map[string]model.Session{
		"nicht open":   {ID: "x", ServiceID: "d", CreatedAt: t0(), Status: model.SessionComplete},
		"ohne Kennung": {ServiceID: "d", CreatedAt: t0(), Status: model.SessionOpen},
		"ohne Service": {ID: "x", CreatedAt: t0(), Status: model.SessionOpen},
	} {
		erwarte(t, v.Sessions.Create(ctx, s), model.ErrUngueltig)
		_ = name
	}
	ok(t, v.Sessions.Create(ctx, session("s1", "d", 0)))
	erwarte(t, v.Sessions.Create(ctx, session("s1", "d", 1)), outbound.ErrBereitsVorhanden)
	erwarte(t, v.Sessions.Create(ctx, session("s2", "fremd", 1)), outbound.ErrDienstUnbekannt)
	if _, err := v.Sessions.Get(ctx, "s2"); !errors.Is(err, outbound.ErrNichtGefunden) {
		t.Fatalf("Session mit unbekanntem Service wurde angelegt: %v", err)
	}
}

func sessionFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	_, err := neu(t).Sessions.Get(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
}

func sessionFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	v.FuegeDienstHinzu(t, "d")
	for i, ziel := range []model.SessionStatus{model.SessionComplete, model.SessionIncomplete} {
		id := fmt.Sprintf("s%d", i)
		ok(t, v.Sessions.Create(ctx, session(id, "d", i)))
		ok(t, v.Sessions.SetStatus(ctx, id, ziel))
		got, err := v.Sessions.Get(ctx, id)
		ok(t, err)
		if got.Status != ziel {
			t.Fatalf("Zustand %s, erwartet %s", got.Status, ziel)
		}
		for _, weiter := range []model.SessionStatus{model.SessionOpen, model.SessionComplete, model.SessionIncomplete} {
			erwarte(t, v.Sessions.SetStatus(ctx, id, weiter), outbound.ErrUngueltigerUebergang)
		}
		if got, _ := v.Sessions.Get(ctx, id); got.Status != ziel {
			t.Fatalf("abgelehnter Wechsel änderte den Zustand: %s", got.Status)
		}
	}
	ok(t, v.Sessions.Create(ctx, session("o", "d", 5)))
	erwarte(t, v.Sessions.SetStatus(ctx, "o", model.SessionOpen), outbound.ErrUngueltigerUebergang)
	erwarte(t, v.Sessions.SetStatus(ctx, "o", "kaputt"), outbound.ErrUngueltigerUebergang)
	erwarte(t, v.Sessions.SetStatus(ctx, "gibt-es-nicht", model.SessionComplete), outbound.ErrNichtGefunden)
}

func sessionFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	v.FuegeDienstHinzu(t, "a")
	v.FuegeDienstHinzu(t, "b")
	ok(t, v.Sessions.Create(ctx, session("z", "a", 0)))
	ok(t, v.Sessions.Create(ctx, session("m", "a", 0)))
	ok(t, v.Sessions.Create(ctx, session("k", "b", 1)))
	ok(t, v.Sessions.Create(ctx, session("a1", "a", 2)))
	ok(t, v.Sessions.SetStatus(ctx, "m", model.SessionComplete))
	ids := func(l []model.Session) []string {
		out := []string{}
		for _, s := range l {
			out = append(out, s.ID)
		}
		return out
	}
	l, err := v.Sessions.List(ctx, "a")
	ok(t, err)
	if got := ids(l); !reflect.DeepEqual(got, []string{"m", "z", "a1"}) {
		t.Fatalf("List a: %v", got)
	}
	o, err := v.Sessions.FindOpen(ctx)
	ok(t, err)
	if got := ids(o); !reflect.DeepEqual(got, []string{"z", "k", "a1"}) {
		t.Fatalf("FindOpen: %v", got)
	}
	l, err = v.Sessions.List(ctx, "unbekannt")
	ok(t, err)
	if l == nil || len(l) != 0 {
		t.Fatalf("unbekannter Service: %v", l)
	}
	_, err = v.Sessions.List(ctx, "")
	erwarte(t, err, model.ErrUngueltig)
}

func sessionFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := neu(t)
	v.FuegeDienstHinzu(t, "d")
	ok(t, v.Sessions.Create(ctx, session("s", "d", 0)))
	ok(t, v.Recording.Append(ctx, exchange("s", "e1", 1)))
	ok(t, v.Sessions.Create(ctx, session("q", "d", 1)))
	ok(t, v.Recording.Append(ctx, exchange("q", "e2", 1)))
	v.MacheZurBaselineQuelle(t, "q")
	erwarte(t, v.Sessions.Delete(ctx, "q"), outbound.ErrReferenziert)
	if _, err := v.Recording.Get(ctx, "e2"); err != nil {
		t.Fatalf("Exchange einer referenzierten Session ging verloren: %v", err)
	}
	ok(t, v.Sessions.Delete(ctx, "s"))
	_, err := v.Recording.Get(ctx, "e1")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Sessions.Get(ctx, "s")
	erwarte(t, err, outbound.ErrNichtGefunden)
	erwarte(t, v.Sessions.Delete(ctx, "s"), outbound.ErrNichtGefunden)
}

// RunRecordingVertrag prüft SPEC-105 Nr. 7 bis 13.
func RunRecordingVertrag(t *testing.T, neu Neu) {
	for _, f := range []struct {
		name string
		lauf func(*testing.T, Neu)
	}{
		{"Append und Get tragen alle Felder", recordingFall1},
		{"Append prüft Session und Zustand", recordingFall2},
		{"sequence_no ist lückenlos", recordingFall3},
		{"gleichzeitige Anhänge ordnen sich", recordingFall4},
		{"doppelte Kennung ändert nichts", recordingFall5},
		{"ungültige Exchanges werden abgelehnt", recordingFall6},
		{"Get und ListBySession bei Unbekanntem und Leerem", recordingFall7},
		{"Header-Randformen", recordingFall8},
	} {
		t.Run(f.name, func(t *testing.T) { f.lauf(t, neu) })
	}
}

func offeneSession(t *testing.T, neu Neu) Verbindung {
	t.Helper()
	v := neu(t)
	v.FuegeDienstHinzu(t, "d")
	ok(t, v.Sessions.Create(context.Background(), session("s", "d", 0)))
	return v
}

func recordingFall1(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	e := exchange("s", "e1", 1)
	ok(t, v.Recording.Append(ctx, e))
	got, err := v.Recording.Get(ctx, "e1")
	ok(t, err)
	if !got.StartedAt.Equal(e.StartedAt) || !got.FinishedAt.Equal(e.FinishedAt) {
		t.Fatalf("Zeiten: %v %v", got.StartedAt, got.FinishedAt)
	}
	got.StartedAt, got.FinishedAt = e.StartedAt, e.FinishedAt
	if !reflect.DeepEqual(got, e) {
		t.Fatalf("gelesen %+v\ngeschrieben %+v", got, e)
	}
}

func recordingFall2(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	erwarte(t, v.Recording.Append(ctx, exchange("fremd", "e1", 1)), outbound.ErrNichtGefunden)
	ok(t, v.Sessions.SetStatus(ctx, "s", model.SessionComplete))
	erwarte(t, v.Recording.Append(ctx, exchange("s", "e1", 1)), outbound.ErrSessionGeschlossen)
	if _, err := v.Recording.Get(ctx, "e1"); !errors.Is(err, outbound.ErrNichtGefunden) {
		t.Fatalf("Exchange an geschlossene Session angehängt: %v", err)
	}
}

func recordingFall3(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	erwarte(t, v.Recording.Append(ctx, exchange("s", "e0", 0)), model.ErrUngueltig)
	erwarte(t, v.Recording.Append(ctx, exchange("s", "e2", 2)), outbound.ErrUngueltigeSequenz)
	ok(t, v.Recording.Append(ctx, exchange("s", "e1", 1)))
	erwarte(t, v.Recording.Append(ctx, exchange("s", "e1b", 1)), outbound.ErrUngueltigeSequenz)
	erwarte(t, v.Recording.Append(ctx, exchange("s", "e3", 3)), outbound.ErrUngueltigeSequenz)
	ok(t, v.Recording.Append(ctx, exchange("s", "e2", 2)))
	l, err := v.Recording.ListBySession(ctx, "s")
	ok(t, err)
	if len(l) != 2 || l[0].ID != "e1" || l[1].ID != "e2" {
		t.Fatalf("Reihenfolge: %v", l)
	}
}

func recordingFall4(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	const n = 8
	fehler := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fehler[i] = v.Recording.Append(ctx, exchange("s", fmt.Sprintf("e%d", i), 1))
		}()
	}
	wg.Wait()
	erfolge := 0
	for _, err := range fehler {
		switch {
		case err == nil:
			erfolge++
		case !errors.Is(err, outbound.ErrUngueltigeSequenz):
			t.Fatalf("unerwarteter Fehler %v", err)
		}
	}
	if erfolge != 1 {
		t.Fatalf("%d Anhänge mit sequence_no 1 gelangen", erfolge)
	}
}

func recordingFall5(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	ok(t, v.Recording.Append(ctx, exchange("s", "e1", 1)))
	anders := exchange("s", "e1", 2)
	anders.Path = "/Anders"
	erwarte(t, v.Recording.Append(ctx, anders), outbound.ErrBereitsVorhanden)
	got, err := v.Recording.Get(ctx, "e1")
	ok(t, err)
	if got.Path != "/Products" || got.SequenceNo != 1 {
		t.Fatalf("gespeicherter Exchange wurde überschrieben: %+v", got)
	}
}

func recordingFall6(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	fehlerhaft := map[string]func(*model.Exchange){
		"ohne Kennung":                func(e *model.Exchange) { e.ID = "" },
		"ohne Methode":                func(e *model.Exchange) { e.Method = "" },
		"unbekannter Capture-Zustand": func(e *model.Exchange) { e.Capture = "kaputt" },
		"present ohne Blob":           func(e *model.Exchange) { e.ResponseBlob = "" },
		"Blob ohne present":           func(e *model.Exchange) { e.RequestBlob = "sha256:x" },
		"match_key leer bei complete": func(e *model.Exchange) { e.MatchKey = "" },
		"discarded bei complete":      func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = model.BodyDiscarded, "" },
		"unsafe Request ohne failed":  func(e *model.Exchange) { e.RequestRedact = model.RedactionUnsafe },
		"unsafe Response ohne none":   func(e *model.Exchange) { e.ResponseRedact = model.RedactionUnsafe },
		"ohne Startzeit":              func(e *model.Exchange) { e.StartedAt = time.Time{} },
	}
	for name, mut := range fehlerhaft {
		e := exchange("s", "x", 1)
		mut(&e)
		if err := v.Recording.Append(ctx, e); !errors.Is(err, model.ErrUngueltig) {
			t.Fatalf("%s: Fehler %v", name, err)
		}
	}
	if l, err := v.Recording.ListBySession(ctx, "s"); err != nil || len(l) != 0 {
		t.Fatalf("ungültige Exchanges wurden gespeichert: %v %v", l, err)
	}
	fehlgeschlagen := exchange("s", "f", 1)
	fehlgeschlagen.MatchKey, fehlgeschlagen.Capture = "", model.CaptureFailed
	fehlgeschlagen.RequestRedact, fehlgeschlagen.RequestBody = model.RedactionUnsafe, model.BodyDiscarded
	ok(t, v.Recording.Append(ctx, fehlgeschlagen))
}

func recordingFall7(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	_, err := v.Recording.Get(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	_, err = v.Recording.ListBySession(ctx, "gibt-es-nicht")
	erwarte(t, err, outbound.ErrNichtGefunden)
	l, err := v.Recording.ListBySession(ctx, "s")
	ok(t, err)
	if l == nil || len(l) != 0 {
		t.Fatalf("leere Session: %v", l)
	}
}

func recordingFall8(t *testing.T, neu Neu) {
	ctx := context.Background()
	v := offeneSession(t, neu)
	a := exchange("s", "e1", 1)
	a.RequestHeaders, a.ResponseHeaders, a.ResponseStatus = nil, nil, 0
	b := exchange("s", "e2", 2)
	b.RequestHeaders, b.ResponseHeaders = map[string][]string{}, map[string][]string{}
	ok(t, v.Recording.Append(ctx, a))
	ok(t, v.Recording.Append(ctx, b))
	ga, err := v.Recording.Get(ctx, "e1")
	ok(t, err)
	gb, err := v.Recording.Get(ctx, "e2")
	ok(t, err)
	if ga.RequestHeaders == nil || len(ga.RequestHeaders) != 0 {
		t.Fatalf("Request-Header ohne Wert: %#v", ga.RequestHeaders)
	}
	if ga.ResponseHeaders != nil || ga.ResponseStatus != 0 {
		t.Fatalf("Response ohne Wert: %#v %d", ga.ResponseHeaders, ga.ResponseStatus)
	}
	if gb.ResponseHeaders == nil || len(gb.ResponseHeaders) != 0 {
		t.Fatalf("leere Response-Header: %#v", gb.ResponseHeaders)
	}
}
