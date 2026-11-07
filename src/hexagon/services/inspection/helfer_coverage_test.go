package inspection_test

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

var _ outbound.CoverageRepository = (*fakeCoverage)(nil)

// fakeCoverage bildet CoverageRepository nach (SPEC-107). Er hält jedes Argument fest (ctx auch als
// nil, Kennung) und zählt jede Operation, auch Get, die der Dienst nie rufen darf. List liefert die
// Fixture und den gesetzten Fehler zugleich (ein Port liefert bei einem Fehler nil; der Fake ist
// strenger, damit ein Dienst, der Daten neben dem Fehler weitergibt, auffällt) und bei einem
// beendeten Kontext dessen Fehler. Er gerät nicht in Panik.
type fakeCoverage struct {
	mu      sync.Mutex
	liste   []model.Coverage
	fehler  error
	zaehler map[string]int
	ctxs    []context.Context
	ids     []string
	// nachLesen läuft nach dem Lesen, vor der Rückkehr, ohne gehaltene Sperre, nur ohne Fehler.
	nachLesen func()
	// nummeriere setzt je Aufruf eine fortlaufende Nummer ab 1 in die Evidenz des ersten Eintrags.
	nummeriere bool
	nummer     int
}

func neuesFakeCoverage() *fakeCoverage { return &fakeCoverage{zaehler: map[string]int{}} }

// setzeListe legt die Fixture fest, die List liefert; nil bleibt nil.
func (f *fakeCoverage) setzeListe(l []model.Coverage) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.liste = slices.Clone(l)
}

func (f *fakeCoverage) List(ctx context.Context, id string) ([]model.Coverage, error) {
	f.mu.Lock()
	f.zaehler["List"]++
	f.ctxs = append(f.ctxs, ctx)
	f.ids = append(f.ids, id)
	var err error
	switch {
	case ctx != nil && ctx.Err() != nil:
		err = ctx.Err()
	case f.fehler != nil:
		err = f.fehler
	}
	liste := slices.Clone(f.liste)
	if f.nummeriere && len(liste) > 0 {
		f.nummer++
		liste[0].EvidenceJSON = strconv.Itoa(f.nummer)
	}
	hook := f.nachLesen
	f.mu.Unlock()
	if err == nil && hook != nil {
		hook()
	}
	return liste, err
}

func (f *fakeCoverage) Get(_ context.Context, id, entitySet string) (model.Coverage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.zaehler["Get"]++
	return model.Coverage{}, fmt.Errorf("Coverage %q/%q: %w", id, entitySet, outbound.ErrNichtGefunden)
}

// anzahl ist die Zahl der Aufrufe von op; gesamt die aller Operationen.
func (f *fakeCoverage) anzahl(op string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.zaehler[op]
}

func (f *fakeCoverage) gesamt() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, z := range f.zaehler {
		n += z
	}
	return n
}

// kennungen und kontexte sind die Argumente der Aufrufe von List in der Folge der Aufrufe.
func (f *fakeCoverage) kennungen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ids)
}

func (f *fakeCoverage) kontexte() []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxs)
}

var _ support.TelemetryReader = (*hakenTelemetrie)(nil)

// hakenTelemetrie bettet den Fake des Ports ein und überschreibt allein Snapshot: der eingebettete
// Snapshot zählt und zeichnet auf, danach läuft der Haken auf dem gelieferten Stand. Die
// Operationen von Telemetry bleiben über den eingebetteten Fake erreichbar und zählbar.
type hakenTelemetrie struct {
	*supporttest.FakeTelemetry
	haken func(*support.Snapshot)
}

func (h *hakenTelemetrie) Snapshot() support.Snapshot {
	s := h.FakeTelemetry.Snapshot()
	if h.haken != nil {
		h.haken(&s)
	}
	return s
}

// umgebung bündelt die Fakes einer Konfiguration.
type umgebung struct {
	rec *fakeRecording
	cov *fakeCoverage
	tel *supporttest.FakeTelemetry
	k   inspection.Konfig
}

// neueUmgebung liefert eine gültige Konfiguration mit Modus record und frischen Fakes.
func neueUmgebung() *umgebung {
	u := &umgebung{rec: neuesFake(), cov: neuesFakeCoverage(), tel: supporttest.NewFakeTelemetry()}
	u.k = inspection.Konfig{Exchanges: u.rec, Coverage: u.cov, Zaehler: u.tel, Modus: "record"}
	return u
}

// konfigMit ist die gültige Konfiguration mit dem Fake f für die Exchanges.
func konfigMit(f *fakeRecording) inspection.Konfig {
	u := neueUmgebung()
	u.k.Exchanges = f
	return u.k
}

// dienst baut den Dienst aus u.k und verlangt Erfolg.
func (u *umgebung) dienst(t *testing.T) *inspection.Inspektion {
	t.Helper()
	d, err := inspection.New(u.k)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return d
}
