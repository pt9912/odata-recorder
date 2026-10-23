package session_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// errDisk trägt den Marker, der in keinem Eintrag des Protokolls und in keinem Fehlertext eines
// *inbound.Fehler stehen darf.
var errDisk = errors.New("platte voll mit " + marker)

const (
	marker     = "M-7f"
	dienstA    = "svc-a"
	dienstB    = "svc-b"
	dienstC    = "svc-c"
	zeitgrenze = 5 * time.Second
)

func t0() time.Time { return time.Date(2026, time.October, 23, 12, 0, 0, 0, time.UTC) }

// sesID ist die n-te Kennung, die FestIDs für eine Session vergibt.
func sesID(n int) string { return fmt.Sprintf("ses-%032d", n) }

// --- Ereignisprotokoll und Steuerung der Fakes --------------------------------------------

type ereignisse struct {
	mu   sync.Mutex
	alle []string
}

func (p *ereignisse) melde(e string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.alle = append(p.alle, e)
}

func (p *ereignisse) liste() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.alle)
}

type aufruf struct {
	op string
	n  int
}

// steuerung zählt die Aufrufe je Operation und lässt den n-ten Aufruf scheitern oder eine Aktion
// ausführen; n 0 gilt für jeden Aufruf.
type steuerung struct {
	mu       sync.Mutex
	zaehler  map[string]int
	fehler   map[aufruf]error
	aktion   map[aufruf]func()
	ctxs     map[string][]context.Context
	fehlerID map[[2]string]error
	aktionID map[[2]string]func()
}

func neueSteuerung() *steuerung {
	return &steuerung{
		zaehler: map[string]int{}, fehler: map[aufruf]error{}, aktion: map[aufruf]func(){}, ctxs: map[string][]context.Context{},
		fehlerID: map[[2]string]error{}, aktionID: map[[2]string]func(){},
	}
}

// scheitereFuer lässt jeden Aufruf von op mit der Kennung id scheitern, unabhängig von der Reihenfolge der Aufrufe.
func (s *steuerung) scheitereFuer(op, id string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fehlerID[[2]string{op, id}] = err
}

// beiFuer führt f in jedem Aufruf von op mit der Kennung id aus.
func (s *steuerung) beiFuer(op, id string, f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aktionID[[2]string{op, id}] = f
}

// halteFuer hält jeden Aufruf von op mit der Kennung id an.
func (s *steuerung) halteFuer(op, id string) *halt {
	h := &halt{erreicht: make(chan struct{}), frei: make(chan struct{})}
	s.beiFuer(op, id, func() {
		close(h.erreicht)
		<-h.frei
	})
	return h
}

// sehe hält den ctx fest, den der Aufruf von op erhielt; ein nil-ctx wird festgehalten, nicht
// ausgewertet.
func (s *steuerung) sehe(ctx context.Context, op string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctxs[op] = append(s.ctxs[op], ctx)
}

// gesehen liefert die erhaltenen ctx von op in der Folge der Aufrufe.
func (s *steuerung) gesehen(op string) []context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.ctxs[op])
}

// vor zählt den Aufruf, führt die Aktion aus und liefert dann den Fehler des Kontexts, sonst den
// eingestellten Fehler. Ein nil-ctx panict beim Auswerten von cerr, vor dem Zählen.
func (s *steuerung) vor(op string, cerr func() error) error {
	return s.vorMit(op, "", cerr)
}

// vorMit ist vor für einen Aufruf mit der Kennung id: nach der Aktion des n-ten Aufrufs läuft die
// Aktion der Kennung, nach dem Fehler des n-ten Aufrufs gilt der Fehler der Kennung.
func (s *steuerung) vorMit(op, id string, cerr func() error) error {
	s.mu.Lock()
	s.zaehler[op]++
	n := s.zaehler[op]
	a, ok := s.aktion[aufruf{op, n}]
	if !ok {
		a = s.aktion[aufruf{op, 0}]
	}
	f, ok := s.fehler[aufruf{op, n}]
	if !ok {
		f = s.fehler[aufruf{op, 0}]
	}
	var aID func()
	var fID error
	if id != "" {
		aID, fID = s.aktionID[[2]string{op, id}], s.fehlerID[[2]string{op, id}]
	}
	s.mu.Unlock()
	if a != nil {
		a()
	}
	if aID != nil {
		aID()
	}
	if cerr != nil {
		if err := cerr(); err != nil {
			return err
		}
	}
	if f != nil {
		return f
	}
	return fID
}

func (s *steuerung) scheitere(op string, n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fehler[aufruf{op, n}] = err
}

func (s *steuerung) bei(op string, n int, f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aktion[aufruf{op, n}] = f
}

func (s *steuerung) zaehle(op string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.zaehler[op]++
}

func (s *steuerung) anzahl(op string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.zaehler[op]
}

func (s *steuerung) gesamt() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, z := range s.zaehler {
		n += z
	}
	return n
}

// halt hält den n-ten Aufruf einer Operation an: erreicht schließt, wenn er ankommt, und er
// läuft weiter, wenn frei schließt.
type halt struct {
	erreicht chan struct{}
	frei     chan struct{}
	einmal   sync.Once
}

func (h *halt) gib() { h.einmal.Do(func() { close(h.frei) }) }

func (s *steuerung) halte(op string, n int) *halt {
	h := &halt{erreicht: make(chan struct{}), frei: make(chan struct{})}
	s.bei(op, n, func() {
		close(h.erreicht)
		<-h.frei
	})
	return h
}

// barriere lässt n Aufrufe warten, bis alle angekommen sind; wer länger als zeitgrenze wartet,
// läuft weiter und vermerkt es in abgelaufen.
type barriere struct {
	mu         sync.Mutex
	da, n      int
	frei       chan struct{}
	abgelaufen bool
}

func neueBarriere(n int) *barriere { return &barriere{n: n, frei: make(chan struct{})} }

func (b *barriere) an() {
	b.mu.Lock()
	b.da++
	if b.da == b.n {
		close(b.frei)
	}
	b.mu.Unlock()
	select {
	case <-b.frei:
	case <-time.After(zeitgrenze):
		b.mu.Lock()
		b.abgelaufen = true
		b.mu.Unlock()
	}
}

func (b *barriere) lief() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.abgelaufen
}

// warte wartet höchstens zeitgrenze auf das Schließen von ch.
func warte(t *testing.T, ch <-chan struct{}, was string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(zeitgrenze):
		t.Fatalf("%s: nach %v kein Ereignis", was, zeitgrenze)
	}
}

// --- Uhr und Kennungen ----------------------------------------------------------------------

// zaehlUhr ist FestUhr, die bei jedem Now um 1 ns vorrückt und es in Ereignissen und Steuerung zählt.
type zaehlUhr struct {
	*supporttest.FestUhr
	st *steuerung
	p  *ereignisse
	mu sync.Mutex
}

func (u *zaehlUhr) Now() time.Time {
	u.mu.Lock()
	defer u.mu.Unlock()
	v := u.FestUhr.Now()
	u.Advance(time.Nanosecond)
	u.p.melde("Now")
	u.st.zaehle("Now")
	return v
}

// festeZeit ist eine Uhr mit einem festen Wert, auch außerhalb der Jahre 1 bis 9999.
type festeZeit time.Time

func (z festeZeit) Now() time.Time { return time.Time(z) }

// zaehlIDs ist FestIDs, die SessionID in Ereignissen und Steuerung zählt.
type zaehlIDs struct {
	*supporttest.FestIDs
	st *steuerung
	p  *ereignisse
}

func (g *zaehlIDs) SessionID() string {
	g.p.melde("SessionID")
	g.st.zaehle("SessionID")
	return g.FestIDs.SessionID()
}

// --- Sessions und Dienste (SPEC-105 Nr. 1 bis 6, SPEC-107 Nr. 2) --------------------------

// fakeSessions bildet SessionRepository nach: Create meldet die Doppelung, Delete löscht jede
// Session unabhängig vom Status, Get und Delete einer unbekannten Kennung ErrNichtGefunden.
type fakeSessions struct {
	st         *steuerung
	p          *ereignisse
	mu         sync.Mutex
	alle       map[string]model.Session
	nachDelete func()
	// beiSetStatus läuft vor dem Statuswechsel mit dem erhaltenen ctx und der Kennung.
	beiSetStatus func(ctx context.Context, id string)
	// gesetzt hält den Status jedes Aufrufs von SetStatus je Kennung fest, erfolge die gelungenen.
	gesetzt map[string][]model.SessionStatus
	erfolge map[string]int
	// feste Liste von FindOpen; nil heißt: die offenen Sessions des Bestands.
	feste *[]model.Session
}

// setzeOffene legt fest, was FindOpen liefert, auch nil.
func (s *fakeSessions) setzeOffene(liste []model.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feste = &liste
}

// statusVon liefert den Status der Session im Bestand.
func (s *fakeSessions) statusVon(id string) model.SessionStatus {
	x, _ := s.hat(id)
	return x.Status
}

// versuche liefert die Zahl der Aufrufe von SetStatus für id, die Zahl der gelungenen und die Status.
func (s *fakeSessions) versuche(id string) (int, int, []model.SessionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.gesetzt[id]), s.erfolge[id], slices.Clone(s.gesetzt[id])
}

func (s *fakeSessions) fuege(id, dienst string, erzeugt time.Time, status model.SessionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alle[id] = model.Session{ID: id, ServiceID: dienst, CreatedAt: erzeugt, Status: status}
}

func (s *fakeSessions) hat(id string) (model.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	x, ok := s.alle[id]
	return x, ok
}

func (s *fakeSessions) anzahl() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.alle)
}

func (s *fakeSessions) Create(ctx context.Context, x model.Session) error {
	s.p.melde("Sessions.Create(" + x.ID + ")")
	s.st.sehe(ctx, "Sessions.Create")
	if err := s.st.vor("Sessions.Create", ctx.Err); err != nil {
		return err
	}
	if err := x.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, da := s.alle[x.ID]; da {
		return fmt.Errorf("%s: %w", marker, outbound.ErrBereitsVorhanden)
	}
	s.alle[x.ID] = x
	return nil
}

func (s *fakeSessions) Get(ctx context.Context, id string) (model.Session, error) {
	s.p.melde("Sessions.Get(" + id + ")")
	s.st.sehe(ctx, "Sessions.Get")
	if err := s.st.vorMit("Sessions.Get", id, ctx.Err); err != nil {
		return model.Session{}, err
	}
	x, ok := s.hat(id)
	if !ok {
		return model.Session{}, fmt.Errorf("%s: %w", marker, outbound.ErrNichtGefunden)
	}
	return x, nil
}

func (s *fakeSessions) SetStatus(ctx context.Context, id string, status model.SessionStatus) error {
	s.p.melde("SetStatus(" + id + ")")
	s.st.sehe(ctx, "SetStatus")
	if status != model.SessionComplete && status != model.SessionIncomplete {
		panic(fmt.Sprintf("SetStatus mit dem Status %q: der Dienst setzt nur complete oder incomplete", status))
	}
	s.mu.Lock()
	hak := s.beiSetStatus
	s.mu.Unlock()
	if hak != nil {
		hak(ctx, id)
	}
	if err := s.st.vorMit("SetStatus", id, ctx.Err); err != nil {
		s.vermerke(id, status, false)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	x, ok := s.alle[id]
	switch {
	case !ok:
		s.vermerkeGesperrt(id, status, false)
		return outbound.ErrNichtGefunden
	case !x.Status.KannWechselnNach(status):
		s.vermerkeGesperrt(id, status, false)
		return outbound.ErrUngueltigerUebergang
	}
	x.Status = status
	s.alle[id] = x
	s.vermerkeGesperrt(id, status, true)
	return nil
}

func (s *fakeSessions) vermerke(id string, status model.SessionStatus, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vermerkeGesperrt(id, status, ok)
}

func (s *fakeSessions) vermerkeGesperrt(id string, status model.SessionStatus, ok bool) {
	if s.gesetzt == nil {
		s.gesetzt, s.erfolge = map[string][]model.SessionStatus{}, map[string]int{}
	}
	s.gesetzt[id] = append(s.gesetzt[id], status)
	if ok {
		s.erfolge[id]++
	}
}

// List liefert die Sessions eines Service nach Erzeugungszeit, dann Kennung; nie nil.
func (s *fakeSessions) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	s.p.melde("Sessions.List(" + serviceID + ")")
	s.st.sehe(ctx, "Sessions.List")
	if err := s.st.vor("Sessions.List", ctx.Err); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []model.Session{}
	for _, x := range s.alle {
		if x.ServiceID == serviceID {
			out = append(out, x)
		}
	}
	slices.SortFunc(out, func(a, b model.Session) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out, nil
}

func (s *fakeSessions) FindOpen(ctx context.Context) ([]model.Session, error) {
	s.p.melde("FindOpen")
	s.st.sehe(ctx, "FindOpen")
	if err := s.st.vor("FindOpen", ctx.Err); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.feste != nil {
		return slices.Clone(*s.feste), nil
	}
	out := []model.Session{}
	for _, x := range s.alle {
		if x.Status == model.SessionOpen {
			out = append(out, x)
		}
	}
	slices.SortFunc(out, func(a, b model.Session) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out, nil
}

func (s *fakeSessions) Delete(ctx context.Context, id string) error {
	s.p.melde("Sessions.Delete(" + id + ")")
	s.st.sehe(ctx, "Sessions.Delete")
	if err := s.st.vor("Sessions.Delete", ctx.Err); err != nil {
		return err
	}
	s.mu.Lock()
	if _, ok := s.alle[id]; !ok {
		s.mu.Unlock()
		return fmt.Errorf("%s: %w", marker, outbound.ErrNichtGefunden)
	}
	delete(s.alle, id)
	s.mu.Unlock()
	if s.nachDelete != nil {
		s.nachDelete()
	}
	return nil
}

// fakeDienste bildet MetadataRepository nach; gebraucht wird nur List. Create, Get und SetMetadata
// zählen und melden ein Ereignis wie jede Operation der Fakes und liefern den Nullwert; dass der
// Dienst sie nie ruft, hält TestDienstDientAllein.
type fakeDienste struct {
	st      *steuerung
	p       *ereignisse
	dienste []model.Service
}

func (d *fakeDienste) Create(context.Context, model.Service) error {
	d.p.melde("Dienste.Create")
	d.st.zaehle("Dienste.Create")
	return nil
}

func (d *fakeDienste) Get(_ context.Context, id string) (model.Service, error) {
	d.p.melde("Dienste.Get(" + id + ")")
	d.st.zaehle("Dienste.Get")
	return model.Service{}, nil
}

func (d *fakeDienste) List(ctx context.Context) ([]model.Service, error) {
	d.p.melde("Dienste.List")
	d.st.sehe(ctx, "Dienste.List")
	if err := d.st.vor("Dienste.List", ctx.Err); err != nil {
		return nil, err
	}
	return slices.Clone(d.dienste), nil
}

func (d *fakeDienste) SetMetadata(context.Context, string, string, string) error {
	d.p.melde("Dienste.SetMetadata")
	d.st.zaehle("Dienste.SetMetadata")
	return nil
}

// --- MaintenanceUseCase -----------------------------------------------------------------------

// ctxSchluessel ist der Schlüssel des Werts, den ein Test in den ctx legt; ein
// context.Background() trägt ihn nicht.
type ctxSchluessel struct{}

// sicht ist, was CleanupBlobs vom erhaltenen ctx sah.
type sicht struct {
	wert     any
	fehler   error
	frist    time.Time
	hatFrist bool
}

// fakeWartung bildet MaintenanceUseCase nach: CleanupBlobs zählt, hält auf Wunsch, meldet den
// erhaltenen ctx und liefert den eingestellten Bericht und Fehler; Startup bricht den Test ab.
type fakeWartung struct {
	st      *steuerung
	p       *ereignisse
	mu      sync.Mutex
	bericht inbound.Loeschbericht
	fehler  error
	sichten []sicht
}

func (w *fakeWartung) Startup(context.Context) (inbound.Wartungsbericht, error) {
	w.st.zaehle("Startup")
	panic("Startup gerufen: der Sessiondienst ruft es nie")
}

func (w *fakeWartung) CleanupBlobs(ctx context.Context) (inbound.Loeschbericht, error) {
	w.p.melde("CleanupBlobs")
	w.st.sehe(ctx, "CleanupBlobs")
	fr, ok := ctx.Deadline()
	w.mu.Lock()
	w.sichten = append(w.sichten, sicht{wert: ctx.Value(ctxSchluessel{}), fehler: ctx.Err(), frist: fr, hatFrist: ok})
	b, f := w.bericht, w.fehler
	w.mu.Unlock()
	_ = w.st.vor("CleanupBlobs", nil)
	return b, f
}

func (w *fakeWartung) gesehen() []sicht {
	w.mu.Lock()
	defer w.mu.Unlock()
	return slices.Clone(w.sichten)
}

// --- RecordUseCase und RecordingRepository (SPEC-121 Nr. 13, SPEC-105 Nr. 11) ----------------

// drainArg hält Kennung und Frist eines Aufrufs von Drain fest; den ctx hält steuerung.gesehen("Drain").
type drainArg struct {
	id    string
	frist time.Duration
}

// fakeRecord bildet RecordUseCase nach: Drain zählt, hält jedes Argument fest, hält und scheitert
// auf Wunsch (steuerung) und liefert den eingestellten Bericht, auch bei beendetem ctx; Record
// bricht den Test ab.
type fakeRecord struct {
	st       *steuerung
	p        *ereignisse
	mu       sync.Mutex
	bericht  inbound.DrainBericht
	berichte map[string]inbound.DrainBericht
	args     []drainArg
}

func (r *fakeRecord) Record(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
	r.st.zaehle("Record.Record")
	panic("Record gerufen: der Sessiondienst ruft nur Drain")
}

func (r *fakeRecord) Drain(ctx context.Context, id string, frist time.Duration) (inbound.DrainBericht, error) {
	_ = ctx.Err() // ein nil-ctx panict hier, vor dem Zählen
	r.p.melde("Drain(" + id + ")")
	r.st.sehe(ctx, "Drain")
	r.mu.Lock()
	r.args = append(r.args, drainArg{id, frist})
	b, ok := r.berichte[id]
	if !ok {
		b = r.bericht
	}
	r.mu.Unlock()
	if err := r.st.vorMit("Drain", id, nil); err != nil {
		return inbound.DrainBericht{}, err
	}
	return b, nil
}

func (r *fakeRecord) aufrufe() []drainArg {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.args)
}

// fakeExchanges bildet RecordingRepository nach: ListBySession liefert die Exchanges einer
// bekannten Session nach sequence_no, eine leere Liste ohne Exchange und ErrNichtGefunden für eine
// unbekannte Session; Append und Get brechen den Test ab.
type fakeExchanges struct {
	st       *steuerung
	p        *ereignisse
	sessions *fakeSessions
	mu       sync.Mutex
	alle     map[string][]model.Exchange
}

// haenge fügt der Session einen Exchange mit dem Capture-Status c an.
func (x *fakeExchanges) haenge(sessionID string, c model.CaptureStatus) {
	x.mu.Lock()
	defer x.mu.Unlock()
	n := len(x.alle[sessionID]) + 1
	x.alle[sessionID] = append(x.alle[sessionID], model.Exchange{
		ID: fmt.Sprintf("%s-x%d", sessionID, n), SessionID: sessionID, SequenceNo: n, Capture: c,
	})
}

func (x *fakeExchanges) Append(context.Context, model.Exchange) error {
	x.st.zaehle("Append")
	panic("Append gerufen: der Sessiondienst liest nur")
}

func (x *fakeExchanges) Get(context.Context, string) (model.Exchange, error) {
	x.st.zaehle("Exchanges.Get")
	panic("Exchanges.Get gerufen: der Sessiondienst ruft nur ListBySession")
}

func (x *fakeExchanges) ListBySession(ctx context.Context, id string) ([]model.Exchange, error) {
	x.p.melde("ListBySession(" + id + ")")
	x.st.sehe(ctx, "ListBySession")
	if err := x.st.vorMit("ListBySession", id, ctx.Err); err != nil {
		return nil, err
	}
	if _, ok := x.sessions.hat(id); !ok {
		return nil, verpackt(outbound.ErrNichtGefunden)
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	return append([]model.Exchange{}, x.alle[id]...), nil
}

// verfallCtx ist ein Kontext, dessen Err nach verfalle context.DeadlineExceeded liefert, ohne
// dass eine Zeit vergeht; er trägt unter ctxSchluessel den Wert wert, WithoutCancel davon
// behält den Wert und liefert Err nil.
type verfallCtx struct {
	wert      any
	verfallen atomic.Bool
}

func (*verfallCtx) Deadline() (time.Time, bool) { return time.Time{}, false }

func (*verfallCtx) Done() <-chan struct{} { return nil }

func (c *verfallCtx) Value(key any) any {
	if key == (ctxSchluessel{}) {
		return c.wert
	}
	return nil
}

func (c *verfallCtx) Err() error {
	if c.verfallen.Load() {
		return context.DeadlineExceeded
	}
	return nil
}

func (c *verfallCtx) verfalle() { c.verfallen.Store(true) }

// --- Fang-Handler für slog ------------------------------------------------------------------

type eintrag struct {
	stufe slog.Level
	text  string
	attrs map[string]string
}

type fang struct {
	mu        sync.Mutex
	eintraege []eintrag
}

func (*fang) Enabled(context.Context, slog.Level) bool { return true }

func (f *fang) Handle(_ context.Context, rec slog.Record) error {
	e := eintrag{stufe: rec.Level, text: rec.Message, attrs: map[string]string{}}
	rec.Attrs(func(a slog.Attr) bool {
		if _, doppelt := e.attrs[a.Key]; doppelt {
			e.attrs["!doppelt"] = a.Key
		}
		e.attrs[a.Key] = a.Value.String()
		return true
	})
	f.mu.Lock()
	defer f.mu.Unlock()
	f.eintraege = append(f.eintraege, e)
	return nil
}

func (f *fang) WithAttrs([]slog.Attr) slog.Handler { return f }
func (f *fang) WithGroup(string) slog.Handler      { return f }

func (f *fang) alle() []eintrag {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.eintraege)
}

// --- Umgebung -------------------------------------------------------------------------------

type welt struct {
	t        *testing.T
	p        *ereignisse
	st       *steuerung
	uhr      *zaehlUhr
	ids      *zaehlIDs
	sessions *fakeSessions
	dienste  *fakeDienste
	wartung  *fakeWartung
	record   *fakeRecord
	exchange *fakeExchanges
	log      *fang
	konfig   session.Konfig
	v        *session.Verwaltung
}

// fristTest ist die Frist der Konfig in den Tests: weder 0 noch 10 s.
const fristTest = 7 * time.Second

// neu liefert eine Umgebung mit leeren Fakes und der Uhr bei t0; aendere ändert die Konfiguration
// vor New. Am Ende des Tests steht fest, dass Startup nie gerufen wurde.
func neu(t *testing.T, aendere ...func(*session.Konfig)) *welt {
	t.Helper()
	p, st := &ereignisse{}, neueSteuerung()
	u := &welt{t: t, p: p, st: st, log: &fang{}}
	u.uhr = &zaehlUhr{FestUhr: supporttest.NewFestUhr(t0()), st: st, p: p}
	u.ids = &zaehlIDs{FestIDs: supporttest.NewFestIDs(), st: st, p: p}
	u.sessions = &fakeSessions{st: st, p: p, alle: map[string]model.Session{}}
	u.dienste = &fakeDienste{st: st, p: p}
	u.wartung = &fakeWartung{st: st, p: p, bericht: inbound.Loeschbericht{Entfernt: 2}}
	u.record = &fakeRecord{st: st, p: p, berichte: map[string]inbound.DrainBericht{}}
	u.exchange = &fakeExchanges{st: st, p: p, sessions: u.sessions, alle: map[string][]model.Exchange{}}
	u.konfig = session.Konfig{
		Sessions: u.sessions, Dienste: u.dienste, Wartung: u.wartung, Uhr: u.uhr, IDs: u.ids, Log: slog.New(u.log),
		Record: u.record, Exchanges: u.exchange, Frist: fristTest,
	}
	for _, f := range aendere {
		f(&u.konfig)
	}
	var err error
	if u.v, err = session.New(u.konfig); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if n := st.anzahl("Startup"); n != 0 {
			t.Errorf("Startup %d-mal gerufen", n)
		}
	})
	return u
}

// dienst legt Services an.
func (u *welt) dienst(ids ...string) {
	for _, id := range ids {
		u.dienste.dienste = append(u.dienste.dienste, model.Service{ID: id})
	}
}

// folgeID ist die Kennung, die der Generator als nächste vergibt; der Aufruf zählt als Bezug.
func (u *welt) folgeID() string { return u.ids.SessionID() }

// errPanik ist der Fehler, den die Operationen der Umgebung nach einer Panik liefern; die Panik
// selbst ist als Fehler des Tests gemeldet.
var errPanik = errors.New("Panik")

// fange führt f aus und meldet eine Panik (ein Fake, der auf einem nil-ctx oder bei Startup
// panict) als Fehler von t, statt das Testprogramm abzubrechen; so wird der Test rot, der den Aufruf
// macht, und die übrigen laufen weiter. Es liefert, ob f panict hat.
func fange(t *testing.T, name string, f func()) (panikte bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Helper()
			t.Errorf("%s: Panik %v", name, r)
			panikte = true
		}
	}()
	f()
	return false
}

func (u *welt) create(ctx context.Context, dienst, id string) (s model.Session, err error) {
	u.t.Helper()
	if fange(u.t, "Create", func() { s, err = u.v.Create(ctx, dienst, id) }) {
		return model.Session{}, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return s, err
}

func (u *welt) open(ctx context.Context, dienst string) (s model.Session, err error) {
	u.t.Helper()
	if fange(u.t, "OpenInitialSession", func() { s, err = u.v.OpenInitialSession(ctx, dienst) }) {
		return model.Session{}, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return s, err
}

func (u *welt) get(ctx context.Context, id string) (s model.Session, err error) {
	u.t.Helper()
	if fange(u.t, "Get", func() { s, err = u.v.Get(ctx, id) }) {
		return model.Session{}, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return s, err
}

func (u *welt) list(ctx context.Context) (s []model.Session, err error) {
	u.t.Helper()
	if fange(u.t, "List", func() { s, err = u.v.List(ctx) }) {
		return nil, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return s, err
}

func (u *welt) del(ctx context.Context, id string) (b inbound.Loeschbericht, err error) {
	u.t.Helper()
	if fange(u.t, "Delete", func() { b, err = u.v.Delete(ctx, id) }) {
		return inbound.Loeschbericht{}, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return b, err
}

func (u *welt) beenden(ctx context.Context, id string) (s model.Session, err error) {
	u.t.Helper()
	if fange(u.t, "Beenden", func() { s, err = u.v.Beenden(ctx, id) }) {
		return model.Session{}, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return s, err
}

func (u *welt) closeAll(ctx context.Context) (b inbound.CloseBericht, err error) {
	u.t.Helper()
	if fange(u.t, "CloseAll", func() { b, err = u.v.CloseAll(ctx) }) {
		return inbound.CloseBericht{}, errPanik
	}
	pruefeFehlerWert(u.t, err)
	return b, err
}

// erzeugt ist die Erzeugungszeit der n-ten Session der Tests: mit Bruchteil einer Sekunde.
func erzeugt(n int) time.Time {
	return t0().Add(123456789*time.Nanosecond + time.Duration(n)*time.Second)
}

// offene legt eine offene Session des Dienstes an, deren Kennung von der des Dienstes abweicht.
func (u *welt) offene(id, dienst string, n int) {
	u.sessions.fuege(id, dienst, erzeugt(n), model.SessionOpen)
}

// pruefeFehlerWert hält einen *Fehler gegen SPEC-124 Nr. 8: erlaubter Status, Text aus Code und
// message ohne den Text der Ursache, Ursache über Unwrap (ODR-E7009 und ODR-E7011 aus dem Status
// haben keine).
func pruefeFehlerWert(t *testing.T, err error) {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		return
	}
	if !f.Code.StatusErlaubt(f.Status) || f.Status != f.Code.Stati()[0] {
		t.Errorf("Status %d bei %v nicht der eine Status des Codes", f.Status, f.Code)
	}
	if soll := f.Code.String() + ": " + f.Code.Message(); f.Error() != soll {
		t.Errorf("Error() = %q, erwartet %q", f.Error(), soll)
	}
	if f.Ursache != nil && !errors.Is(err, f.Ursache) {
		t.Errorf("Ursache %v bei %v nicht über Unwrap erreichbar", f.Ursache, f.Code)
	}
	if f.Ursache == nil && f.Code != model.CodeE7009 && f.Code != model.CodeE7011 {
		t.Errorf("keine Ursache bei %v", f.Code)
	}
	if f.Ursache == nil && errors.Unwrap(err) != nil {
		t.Errorf("%v trägt die Ursache %v", f.Code, errors.Unwrap(err))
	}
	if strings.Contains(f.Error(), marker) {
		t.Errorf("Text %q trägt den Marker", f.Error())
	}
}

// codeVon liefert Code und Status eines *Fehler.
func codeVon(t *testing.T, err error) (string, int) {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		t.Fatalf("kein *Fehler: %v", err)
	}
	return f.Code.String(), f.Status
}

// ohneCode sagt, ob err kein *Fehler ist.
func ohneCode(err error) bool {
	var f *inbound.Fehler
	return !errors.As(err, &f)
}

// verpackt hängt den Marker an die Ursache.
func verpackt(err error) error { return fmt.Errorf("%s: %w", marker, err) }
