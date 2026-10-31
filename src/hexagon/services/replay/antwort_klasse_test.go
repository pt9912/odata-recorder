package replay_test

import (
	"context"
	"errors"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

func vierStrategien() []model.ReplayStrategy {
	return []model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast}
}

// fristKontext endet auf Abruf mit context.DeadlineExceeded: ein ctx mit Frist lässt sich nicht an einem Zeitpunkt
// der Transaktion enden, ohne die Uhr zu halten.
type fristKontext struct {
	mu   sync.Mutex
	done chan struct{}
	err  error
}

func neuerFristKontext() *fristKontext {
	return &fristKontext{done: make(chan struct{})}
}

func (c *fristKontext) Deadline() (time.Time, bool) { return time.Time{}, false }

func (c *fristKontext) Value(any) any { return nil }

func (c *fristKontext) Done() <-chan struct{} { return c.done }

func (c *fristKontext) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *fristKontext) beende() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err == nil {
		c.err = context.DeadlineExceeded
		close(c.done)
	}
}

// endeGrund ist ein Zustand des beendeten ctx; neu liefert den ctx und die Funktion, die ihn beendet.
type endeGrund struct {
	name string
	neu  func() (context.Context, context.CancelFunc)
}

// endeGruende sind die drei Zustände (SPEC-135 Nr. 7 c nennt beide Kontextfehler): Canceled, DeadlineExceeded und
// Canceled mit eigener Ursache (context.Cause).
func endeGruende() []endeGrund {
	return []endeGrund{
		{"Canceled", func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }},
		{"DeadlineExceeded", func() (context.Context, context.CancelFunc) {
			c := neuerFristKontext()
			return c, c.beende
		}},
		{"Ursache", func() (context.Context, context.CancelFunc) {
			ctx, cancel := context.WithCancelCause(context.Background())
			return ctx, func() { cancel(errors.New("eigene Ursache")) }
		}},
	}
}

// TestKontextNachDerBestaetigung hält Z44 für die Stellen nach der Bestätigung: Menge 36 (4 Strategien × 3 Zeitpunkte
// × 3 Gründe). Der ctx endet, sobald die Transaktion bestätigt ist (vor jeder Stelle des Dienstes bis zur Rückkehr, mit
// und ohne Eintrag), oder im Eintrag (vor jeder Stelle danach), mit Canceled, mit DeadlineExceeded oder mit einer
// eigenen Ursache; der Dienst prüft ihn nicht mehr und liefert die Antwort.
func TestKontextNachDerBestaetigung(t *testing.T) {
	faelle := []struct {
		name      string
		eintrag   bool
		imEintrag bool
	}{
		{"nach der Bestätigung, ohne Eintrag", false, false},
		{"nach der Bestätigung, mit Eintrag", true, false},
		{"im Eintrag", true, true},
	}
	n := 0
	for _, strat := range vierStrategien() {
		for _, f := range faelle {
			for _, grund := range endeGruende() {
				n++
				kontextNachBestaetigung(t, string(strat)+" "+f.name+" "+grund.name, strat, f.eintrag, f.imEintrag, grund)
			}
		}
	}
	if n != 36 {
		t.Errorf("Menge: %d, erwartet 36", n)
	}
}

// kontextNachBestaetigung fährt ein Element von TestKontextNachDerBestaetigung.
func kontextNachBestaetigung(t *testing.T, was string, strat model.ReplayStrategy, eintrag, imEintrag bool, grund endeGrund) {
	t.Helper()
	u := neu(t)
	u.k.Strategie = strat
	e := u.austausch(1, u.keyA(), true)
	e.KeyRedacted = eintrag
	u.lege(e)
	ctx, cancel := grund.neu()
	if imEintrag {
		u.k.Telemetrie = abbruchTel{u.tel, cancel}
	} else {
		u.st.nachCommit = cancel
	}
	resp, err := u.replay(ctx, anf())
	cancel()
	if err != nil {
		t.Errorf("%s: %v", was, err)
		return
	}
	if got := seqVon(t, resp); got != 1 || resp.Status != 200 {
		t.Errorf("%s: SequenceNo %d, Status %d", was, got, resp.Status)
	}
	if got := string(lies(t, resp)); got != string(bodyVon(1)) {
		t.Errorf("%s: Body %q", was, got)
	}
	if st := u.bl.strome(); len(st) != 1 || st[0].geschlossen() != 0 {
		t.Errorf("%s: Strom vom Dienst geschlossen", was)
	}
	if got, want := u.tel.Anzahl(supporttest.OpLog), map[bool]int{false: 0, true: 1}[eintrag]; got != want {
		t.Errorf("%s: %d Einträge, erwartet %d", was, got, want)
	}
	c, da := u.cursorStand()
	if schreibt := strat == model.StrategySequential || strat == model.StrategyStrict; da != schreibt || schreibt && c.Position != 1 {
		t.Errorf("%s: Cursor %+v (da: %v)", was, c, da)
	}
}

// TestStromUnveraendert hält Z56 am Dienst: Menge 44 (4 Strategien × 11). Der Body der Antwort ist der Strom, den
// Blobs.Open lieferte (Identität), trägt jede der neun Schnittstellen des Originals, und io.Copy nutzt sein WriteTo,
// ohne Read.
func TestStromUnveraendert(t *testing.T) {
	schnittstellen := []reflect.Type{
		reflect.TypeOf((*io.Reader)(nil)).Elem(), reflect.TypeOf((*io.WriterTo)(nil)).Elem(),
		reflect.TypeOf((*io.Closer)(nil)).Elem(), reflect.TypeOf((*io.Seeker)(nil)).Elem(),
		reflect.TypeOf((*io.ReaderAt)(nil)).Elem(), reflect.TypeOf((*io.ByteReader)(nil)).Elem(),
		reflect.TypeOf((*io.ByteScanner)(nil)).Elem(), reflect.TypeOf((*io.RuneReader)(nil)).Elem(),
		reflect.TypeOf((*io.RuneScanner)(nil)).Elem(),
	}
	n := 0
	for _, strat := range vierStrategien() {
		u := neu(t)
		u.k.Strategie = strat
		erfolg(u)
		resp, err := u.replay(context.Background(), anf())
		if err != nil {
			t.Errorf("%s: %v", strat, err)
			continue
		}
		st := u.bl.strome()
		if len(st) != 1 {
			t.Fatalf("%s: %d Ströme", strat, len(st))
		}
		for _, s := range schnittstellen {
			n++
			if !reflect.TypeOf(resp.Body).Implements(s) {
				t.Errorf("%s: Body trägt %s nicht", strat, s)
			}
		}
		n++
		if got, ok := resp.Body.(*strom); !ok || got != st[0] {
			t.Errorf("%s: Body ist nicht der Strom von Blobs.Open (%T)", strat, resp.Body)
		}
		n++
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			t.Errorf("%s: io.Copy: %v", strat, err)
		}
		if st[0].writeTos != 1 || st[0].reads != 0 {
			t.Errorf("%s: WriteTo %d, Read %d; erwartet 1 und 0", strat, st[0].writeTos, st[0].reads)
		}
		_ = resp.Body.Close()
	}
	if n != 44 {
		t.Errorf("Menge: %d, erwartet 44", n)
	}
}
