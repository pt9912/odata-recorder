package session_test

import (
	"context"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestGet: Get liest unverändert in jedem Status und reicht die Kennung unverändert an den Port;
// kein Bezug von Zeit oder Kennung, keine Wartung (V9).
func TestGet(t *testing.T) {
	u := neu(t)
	soll := map[string]model.Session{}
	for i, st := range []model.SessionStatus{model.SessionOpen, model.SessionComplete, model.SessionIncomplete} {
		id := "s-" + string(st)
		soll[id] = model.Session{ID: id, ServiceID: "svc-" + string(st), CreatedAt: t0().Add(time.Duration(i+1) * time.Hour), Status: st}
		u.sessions.fuege(id, soll[id].ServiceID, soll[id].CreatedAt, st)
	}
	u.sessions.fuege("Ab c", dienstB, t0().Add(7*time.Hour), model.SessionComplete)
	soll["Ab c"] = model.Session{ID: "Ab c", ServiceID: dienstB, CreatedAt: t0().Add(7 * time.Hour), Status: model.SessionComplete}
	var erwartet []string
	for id, s := range soll {
		got, err := u.get(context.Background(), id)
		if err != nil || got != s {
			t.Errorf("%q: %+v, %v; erwartet %+v", id, got, err, s)
		}
		erwartet = append(erwartet, "Sessions.Get("+id+")")
	}
	faelle := []struct {
		name, id string
		fehler   error
		code     string
		status   int
	}{
		{"unbekannt", "nix", nil, "ODR-E7003", 404},
		{"leer", "", nil, "ODR-E7003", 404},
		{"beschädigt", "s-complete", verpackt(outbound.ErrBeschaedigt), "ODR-E3001", 500},
		{"Platte", "s-complete", errDisk, "ODR-E9001", 500},
	}
	for _, f := range faelle {
		if f.fehler != nil {
			u.st.scheitere("Sessions.Get", 0, f.fehler)
		}
		got, err := u.get(context.Background(), f.id)
		if c, st := codeVon(t, err); c != f.code || st != f.status || got != (model.Session{}) {
			t.Errorf("%s: %s %d, %+v", f.name, c, st, got)
		}
		erwartet = append(erwartet, "Sessions.Get("+f.id+")")
		u.st.scheitere("Sessions.Get", 0, nil)
	}
	ev := u.p.liste()
	slices.Sort(ev)
	slices.Sort(erwartet)
	if !slices.Equal(ev, erwartet) {
		t.Errorf("Ereignisse %v, erwartet %v", ev, erwartet)
	}
	for _, op := range []string{"Now", "SessionID", "CleanupBlobs"} {
		if n := u.st.anzahl(op); n != 0 {
			t.Errorf("%s %d-mal gerufen", op, n)
		}
	}
}

// TestList: jede Session jedes Services in der Reihenfolge der Services, dann der Ablage; leer ist
// []; der erste Fehler beendet ohne Teilliste (V10).
func TestList(t *testing.T) {
	u := neu(t)
	u.dienst(dienstA, dienstB, dienstC)
	u.sessions.fuege("a1", dienstA, t0().Add(2*time.Hour), model.SessionOpen)
	u.sessions.fuege("a2", dienstA, t0().Add(time.Hour), model.SessionComplete)
	u.sessions.fuege("c1", dienstC, t0(), model.SessionIncomplete)
	got, err := u.list(context.Background())
	soll := []model.Session{
		{ID: "a2", ServiceID: dienstA, CreatedAt: t0().Add(time.Hour), Status: model.SessionComplete},
		{ID: "a1", ServiceID: dienstA, CreatedAt: t0().Add(2 * time.Hour), Status: model.SessionOpen},
		{ID: "c1", ServiceID: dienstC, CreatedAt: t0(), Status: model.SessionIncomplete},
	}
	if err != nil || !reflect.DeepEqual(got, soll) {
		t.Errorf("%+v, %v; erwartet %+v", got, err, soll)
	}
	ev := []string{"Dienste.List", "Sessions.List(svc-a)", "Sessions.List(svc-b)", "Sessions.List(svc-c)"}
	if !slices.Equal(u.p.liste(), ev) {
		t.Errorf("Ereignisse %v, erwartet %v", u.p.liste(), ev)
	}

	// Kein Service und Services ohne Session: eine leere Liste, nie nil.
	for name, dienste := range map[string][]string{"kein Service": nil, "Services ohne Session": {dienstA, dienstB}} {
		v := neu(t)
		v.dienst(dienste...)
		got, err := v.list(context.Background())
		if err != nil || !reflect.DeepEqual(got, []model.Session{}) {
			t.Errorf("%s: %#v, %v", name, got, err)
		}
	}

	// Fehler: keine Teilliste, kein weiterer Aufruf.
	faelle := []struct {
		name    string
		op      string
		n       int
		fehler  error
		code    string
		ereigns []string
	}{
		{"Dienste.List", "Dienste.List", 1, errDisk, "ODR-E9001", []string{"Dienste.List"}},
		{"Dienste.List beschädigt", "Dienste.List", 1, verpackt(outbound.ErrBeschaedigt), "ODR-E3001", []string{"Dienste.List"}},
		{"Sessions.List des zweiten", "Sessions.List", 2, errDisk, "ODR-E9001", []string{"Dienste.List", "Sessions.List(svc-a)", "Sessions.List(svc-b)"}},
		{"ErrNichtGefunden", "Sessions.List", 1, verpackt(outbound.ErrNichtGefunden), "ODR-E9001", []string{"Dienste.List", "Sessions.List(svc-a)"}},
	}
	for _, f := range faelle {
		v := neu(t)
		v.dienst(dienstA, dienstB, dienstC)
		v.sessions.fuege("a1", dienstA, t0(), model.SessionComplete)
		v.st.scheitere(f.op, f.n, f.fehler)
		got, err := v.list(context.Background())
		if c, st := codeVon(t, err); c != f.code || st != 500 || got != nil {
			t.Errorf("%s: %s %d, %#v", f.name, c, st, got)
		}
		if !slices.Equal(v.p.liste(), f.ereigns) {
			t.Errorf("%s: Ereignisse %v, erwartet %v", f.name, v.p.liste(), f.ereigns)
		}
		if n := v.st.anzahl("CleanupBlobs"); n != 0 {
			t.Errorf("%s: Wartung gerufen", f.name)
		}
	}
}
