package inspection_test

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// sid ist die Kennung der Session, die ein Test anlegt, wenn er keine eigene braucht.
const sid = "s1"

// nilCtx liefert einen nil-Kontext; ein Literal meldet staticcheck (SA1012).
func nilCtx() context.Context { return nil }

// beispiel ist der i-te Exchange der Session id: lückenlose sequence_no ab 1.
func beispiel(id string, i int) model.Exchange {
	return model.Exchange{
		ID: fmt.Sprintf("ex-%d", i), SessionID: id, SequenceNo: i, Method: "GET", Path: fmt.Sprintf("/p/%d", i),
		StartedAt: time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Second),
	}
}

// fakeRecording bildet RecordingRepository nach (SPEC-105): ListBySession liefert den Fehler eines
// beendeten Kontexts vor dem Lesen und ErrNichtGefunden für eine unbekannte Session. Es hält jedes
// Argument fest (ctx auch als nil, Kennung) und zählt jede Operation, auch Append und Get, die
// der Dienst nie rufen darf; beide verändern den Bestand.
type fakeRecording struct {
	mu        sync.Mutex
	sitzungen map[string][]model.Exchange
	zaehler   map[string]int
	ctxs      []context.Context
	ids       []string
	// fehler ist der Fehler von ListBySession, wenn der Kontext lebt.
	fehler error
	// nachLesen läuft nach dem Lesen, vor der Rückkehr, ohne gehaltene Sperre.
	nachLesen func()
}

func neuesFake() *fakeRecording {
	return &fakeRecording{sitzungen: map[string][]model.Exchange{}, zaehler: map[string]int{}}
}

// legeAn legt die Session id mit n Exchanges an.
func (f *fakeRecording) legeAn(id string, n int) {
	f.mu.Lock()
	f.sitzungen[id] = []model.Exchange{}
	f.mu.Unlock()
	f.haengeAn(id, n)
}

// haengeAn hängt k Exchanges an die Session id an, ohne eine Operation zu zählen.
func (f *fakeRecording) haengeAn(id string, k int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for range k {
		f.sitzungen[id] = append(f.sitzungen[id], beispiel(id, len(f.sitzungen[id])+1))
	}
}

func (f *fakeRecording) Append(_ context.Context, e model.Exchange) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.zaehler["Append"]++
	f.sitzungen[e.SessionID] = append(f.sitzungen[e.SessionID], e)
	return nil
}

func (f *fakeRecording) Get(_ context.Context, id string) (model.Exchange, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.zaehler["Get"]++
	return model.Exchange{}, fmt.Errorf("Exchange %q: %w", id, outbound.ErrNichtGefunden)
}

func (f *fakeRecording) ListBySession(ctx context.Context, id string) ([]model.Exchange, error) {
	f.mu.Lock()
	f.zaehler["ListBySession"]++
	f.ctxs = append(f.ctxs, ctx)
	f.ids = append(f.ids, id)
	var err error
	liste, da := f.sitzungen[id]
	switch {
	case ctx != nil && ctx.Err() != nil:
		err = ctx.Err()
	case f.fehler != nil:
		err = f.fehler
	case !da:
		err = fmt.Errorf("Session %q: %w", id, outbound.ErrNichtGefunden)
	}
	liste = slices.Clone(liste)
	hook := f.nachLesen
	f.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if hook != nil {
		hook()
	}
	return liste, nil
}

// anzahl ist die Zahl der Aufrufe von op; gesamt die aller Operationen.
func (f *fakeRecording) anzahl(op string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.zaehler[op]
}

func (f *fakeRecording) gesamt() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, z := range f.zaehler {
		n += z
	}
	return n
}

// stand ist eine Kopie des Bestands aller Sessions.
func (f *fakeRecording) stand() map[string][]model.Exchange {
	f.mu.Lock()
	defer f.mu.Unlock()
	k := map[string][]model.Exchange{}
	for id, l := range f.sitzungen {
		k[id] = slices.Clone(l)
	}
	return k
}

// kennungen und kontexte sind die Argumente der Aufrufe von ListBySession in der Folge der Aufrufe.
func (f *fakeRecording) kennungen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ids)
}

func (f *fakeRecording) kontexte() []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.ctxs)
}

func neuerDienst(t *testing.T, f *fakeRecording) *inspection.Inspektion {
	t.Helper()
	d, err := inspection.New(inspection.Konfig{Exchanges: f})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return d
}

// lies liest eine Seite mit lebendem Kontext und verlangt Erfolg.
func lies(t *testing.T, d *inspection.Inspektion, id string, s inbound.Seite) inbound.ExchangeSeite {
	t.Helper()
	z, err := d.Exchanges(context.Background(), id, s)
	if err != nil {
		t.Fatalf("Exchanges(%q, %+v): %v", id, s, err)
	}
	return z
}

// folge liest die Folge der Seiten ab nach bis Weiter 0 und liefert jede Seite; mehr Seiten als
// Exchanges plus 2 sind ein Fehler (Weiter käme nicht voran).
func folge(t *testing.T, d *inspection.Inspektion, id string, limit, nach int) []inbound.ExchangeSeite {
	t.Helper()
	var seiten []inbound.ExchangeSeite
	for {
		z := lies(t, d, id, inbound.Seite{Limit: limit, Nach: nach})
		seiten = append(seiten, z)
		if z.Weiter == 0 {
			return seiten
		}
		if z.Weiter <= nach || len(seiten) > z.Gesamt+2 {
			t.Fatalf("Weiter %d kommt von Nach %d nicht voran", z.Weiter, nach)
		}
		nach = z.Weiter
	}
}

// nummern sind die sequence_no der Einträge einer Seite.
func nummern(zs ...inbound.ExchangeSeite) []int {
	nums := []int{}
	for _, z := range zs {
		for _, e := range z.Eintraege {
			nums = append(nums, e.SequenceNo)
		}
	}
	return nums
}

// bereich ist von bis bis einschließlich, leer für bis kleiner von.
func bereich(von, bis int) []int {
	r := []int{}
	for i := von; i <= bis; i++ {
		r = append(r, i)
	}
	return r
}

// wirksam ist das wirksame Limit nach SPEC-236 Nr. 3, gerechnet aus den Konstanten des Pakets.
func wirksam(limit int) int {
	if limit == 0 {
		return inspection.StandardLimit
	}
	return min(limit, inspection.MaxLimit)
}

// modellSeite ist das Orakel der Seite: die Exchanges 1 bis n, geschnitten in einer eigenen Schleife
// (sequence_no über nach, die ersten limit), und Weiter als Nummer des letzten, wenn sie unter n liegt.
func modellSeite(n, nach, limit int) ([]int, int) {
	nums := []int{}
	for q := 1; q <= n; q++ {
		if q > nach && len(nums) < limit {
			nums = append(nums, q)
		}
	}
	weiter := 0
	if len(nums) > 0 && nums[len(nums)-1] < n {
		weiter = nums[len(nums)-1]
	}
	return nums, weiter
}

// gleichSeite vergleicht eine Seite mit dem Modell der Session aus n Exchanges.
func gleichSeite(t *testing.T, was string, z inbound.ExchangeSeite, n, nach, limit int) {
	t.Helper()
	nums, weiter := modellSeite(n, nach, wirksam(limit))
	if z.Eintraege == nil {
		t.Errorf("%s: Eintraege ist nil", was)
	}
	if got := nummern(z); !slices.Equal(got, nums) {
		t.Errorf("%s: Nummern %v, erwartet %v", was, got, nums)
	}
	for _, e := range z.Eintraege {
		if e.ID != fmt.Sprintf("ex-%d", e.SequenceNo) {
			t.Errorf("%s: Eintrag %d trägt ID %q", was, e.SequenceNo, e.ID)
		}
	}
	if z.Weiter != weiter || z.Gesamt != n || z.Limit != wirksam(limit) {
		t.Errorf("%s: Weiter %d Gesamt %d Limit %d, erwartet %d %d %d", was, z.Weiter, z.Gesamt, z.Limit, weiter, n, wirksam(limit))
	}
}
