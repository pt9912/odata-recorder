package maintenance_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
)

// errDisk trägt den Marker, der in keinem Eintrag des Protokolls und in keinem Fehlertext eines
// *inbound.Fehler stehen darf.
var errDisk = errors.New("platte voll mit " + marker)

const (
	marker     = "M-7f"
	dienstA    = "svc-a"
	dienstB    = "svc-b"
	zeitgrenze = 10 * time.Second
	tag        = 24 * time.Hour
)

func t0() time.Time { return time.Date(2026, time.October, 23, 12, 0, 0, 0, time.UTC) }

// r bildet eine gültige Referenz aus einem Namen; die Folge der Namen ist die Folge der Referenzen
// (gleich lange Namen vorausgesetzt).
func r(name string) string {
	h := hex.EncodeToString([]byte(name))
	return "sha256:" + strings.Repeat("0", 64-len(h)) + h
}

// refVon ist die Referenz der Bytes, wie der Blob-Fake sie bei Commit bildet.
func refVon(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}

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

// mit liefert die Ereignisse, die mit praefix beginnen.
func (p *ereignisse) mit(praefix string) []string {
	var out []string
	for _, e := range p.liste() {
		if strings.HasPrefix(e, praefix) {
			out = append(out, e)
		}
	}
	return out
}

type aufruf struct {
	op string
	n  int
}

// steuerung zählt die Aufrufe je Operation und lässt den n-ten Aufruf scheitern oder eine Aktion
// ausführen; n 0 gilt für jeden Aufruf.
type steuerung struct {
	mu      sync.Mutex
	zaehler map[string]int
	fehler  map[aufruf]error
	aktion  map[aufruf]func()
}

func neueSteuerung() *steuerung {
	return &steuerung{zaehler: map[string]int{}, fehler: map[aufruf]error{}, aktion: map[aufruf]func(){}}
}

// vor zählt den Aufruf, führt die Aktion aus und liefert dann den Fehler des Kontexts, sonst den
// eingestellten Fehler.
func (s *steuerung) vor(op string, cerr func() error) (int, error) {
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
	s.mu.Unlock()
	if a != nil {
		a()
	}
	if cerr != nil {
		if err := cerr(); err != nil {
			return n, err
		}
	}
	return n, f
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

// zaehle zählt einen Aufruf, der vor jeder Aktion endet.
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

// warte wartet höchstens zeitgrenze auf das Schließen von ch.
func warte(t *testing.T, ch <-chan struct{}, was string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(zeitgrenze):
		t.Fatalf("%s: nach %v kein Ereignis", was, zeitgrenze)
	}
}

// --- Uhr ------------------------------------------------------------------------------------

// zaehlUhr ist FestUhr mit Zählung von Now in der Steuerung.
type zaehlUhr struct {
	*supporttest.FestUhr
	st *steuerung
}

func (u *zaehlUhr) Now() time.Time {
	v := u.FestUhr.Now()
	_, _ = u.st.vor("Now", nil)
	return v
}

// --- Sessions und Dienste (SPEC-105 Nr. 3 bis 6, SPEC-107 Nr. 2) --------------------------

type fakeSessions struct {
	st       *steuerung
	p        *ereignisse
	mu       sync.Mutex
	alle     map[string]model.Session
	refs     map[string][]string
	ziele    []model.SessionStatus
	gelistet []string
}

func (s *fakeSessions) fuege(id, dienst string, erzeugt time.Time, status model.SessionStatus, refs ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alle[id] = model.Session{ID: id, ServiceID: dienst, CreatedAt: erzeugt, Status: status}
	s.refs[id] = refs
}

func (s *fakeSessions) status(id string) (model.SessionStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	x, ok := s.alle[id]
	return x.Status, ok
}

// sortiert liefert die passenden Sessions nach Erzeugungszeit, dann Kennung; nie nil.
func (s *fakeSessions) sortiert(passt func(model.Session) bool) []model.Session {
	out := []model.Session{}
	for _, x := range s.alle {
		if passt(x) {
			out = append(out, x)
		}
	}
	slices.SortFunc(out, func(a, b model.Session) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out
}

func (s *fakeSessions) Create(_ context.Context, x model.Session) error {
	s.fuege(x.ID, x.ServiceID, x.CreatedAt, x.Status)
	return nil
}

func (s *fakeSessions) Get(_ context.Context, id string) (model.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	x, ok := s.alle[id]
	if !ok {
		return model.Session{}, outbound.ErrNichtGefunden
	}
	return x, nil
}

func (s *fakeSessions) SetStatus(ctx context.Context, id string, status model.SessionStatus) error {
	s.p.melde("SetStatus(" + id + ")")
	if _, err := s.st.vor("SetStatus", ctx.Err); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ziele = append(s.ziele, status)
	x, ok := s.alle[id]
	switch {
	case !ok:
		return outbound.ErrNichtGefunden
	case !x.Status.KannWechselnNach(status):
		return outbound.ErrUngueltigerUebergang
	}
	x.Status = status
	s.alle[id] = x
	return nil
}

func (s *fakeSessions) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	s.p.melde("Sessions.List")
	if _, err := s.st.vor("Sessions.List", ctx.Err); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gelistet = append(s.gelistet, serviceID)
	return s.sortiert(func(x model.Session) bool { return x.ServiceID == serviceID }), nil
}

func (s *fakeSessions) FindOpen(ctx context.Context) ([]model.Session, error) {
	s.p.melde("FindOpen")
	if _, err := s.st.vor("FindOpen", ctx.Err); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sortiert(func(x model.Session) bool { return x.Status == model.SessionOpen }), nil
}

func (s *fakeSessions) Delete(ctx context.Context, id string) error {
	s.p.melde("Delete(" + id + ")")
	if _, err := s.st.vor("Delete", ctx.Err); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.alle[id]; !ok {
		return outbound.ErrNichtGefunden
	}
	delete(s.alle, id)
	delete(s.refs, id)
	return nil
}

type fakeDienste struct {
	st      *steuerung
	p       *ereignisse
	dienste []model.Service
}

func (d *fakeDienste) Create(_ context.Context, s model.Service) error {
	d.dienste = append(d.dienste, s)
	return nil
}

func (d *fakeDienste) Get(_ context.Context, id string) (model.Service, error) {
	for _, s := range d.dienste {
		if s.ID == id {
			return s, nil
		}
	}
	return model.Service{}, outbound.ErrNichtGefunden
}

func (d *fakeDienste) List(ctx context.Context) ([]model.Service, error) {
	d.p.melde("Dienste.List")
	if _, err := d.st.vor("Dienste.List", ctx.Err); err != nil {
		return nil, err
	}
	return append([]model.Service{}, d.dienste...), nil
}

func (d *fakeDienste) SetMetadata(context.Context, string, string, string) error { return nil }

// --- StorageMaintenance (SPEC-108 Nr. 15 bis 17) --------------------------------------------

// fakeSpeicher liefert als Referenzmenge die Referenzen der vorhandenen Sessions, die festen
// Referenzen (Snapshots, Metadaten) und die zusätzlichen der n-ten Lesung.
type fakeSpeicher struct {
	st       *steuerung
	p        *ereignisse
	sessions *fakeSessions
	befunde  []string
	fest     []string
	zusatz   map[int][]string
}

func (s *fakeSpeicher) Check(ctx context.Context) ([]string, error) {
	s.p.melde("Check")
	if err := ctx.Err(); err != nil {
		s.st.zaehle("Check")
		return nil, err
	}
	if _, err := s.st.vor("Check", nil); err != nil {
		return nil, err
	}
	return s.befunde, nil
}

func (s *fakeSpeicher) ReferencedBlobs(ctx context.Context) ([]string, error) {
	s.p.melde("ReferencedBlobs")
	n, err := s.st.vor("ReferencedBlobs", ctx.Err)
	if err != nil {
		return nil, err
	}
	menge := map[string]bool{}
	for _, ref := range append(slices.Clone(s.fest), s.zusatz[n]...) {
		menge[ref] = true
	}
	s.sessions.mu.Lock()
	for _, refs := range s.sessions.refs {
		for _, ref := range refs {
			menge[ref] = true
		}
	}
	s.sessions.mu.Unlock()
	out := []string{}
	for ref := range menge {
		out = append(out, ref)
	}
	slices.Sort(out)
	return out, nil
}

// --- BlobRepository (SPEC-106 Nr. 3, 5 bis 8) -------------------------------------------------

type fakeBlobs struct {
	st      *steuerung
	p       *ereignisse
	mu      sync.Mutex
	dateien map[string][]byte
	kaputt  map[string]bool
}

type fakeWriter struct {
	b      *fakeBlobs
	cerr   func() error
	grenze int64
	buf    []byte
	zu     bool
}

func (b *fakeBlobs) lege(refs ...string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ref := range refs {
		b.dateien[ref] = nil
	}
}

func (b *fakeBlobs) beschaedige(ref string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.kaputt[ref] = true
}

func (b *fakeBlobs) hat(ref string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.dateien[ref]
	return ok
}

func (b *fakeBlobs) refs() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := []string{}
	for ref := range b.dateien {
		out = append(out, ref)
	}
	slices.Sort(out)
	return out
}

func (b *fakeBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	b.p.melde(fmt.Sprintf("Begin(%d)", grenze))
	if _, err := b.st.vor("Begin", ctx.Err); err != nil {
		return nil, err
	}
	if grenze < 1 {
		return nil, model.ErrUngueltig
	}
	return &fakeWriter{b: b, cerr: ctx.Err, grenze: grenze}, nil
}

func (w *fakeWriter) Write(p []byte) (int, error) {
	if w.zu {
		return 0, outbound.ErrAbgeschlossen
	}
	if err := w.cerr(); err != nil {
		return 0, err
	}
	if int64(len(w.buf)+len(p)) > w.grenze {
		w.zu = true
		return 0, outbound.ErrZuGross
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *fakeWriter) Commit() (string, int64, error) {
	if w.zu {
		return "", 0, outbound.ErrAbgeschlossen
	}
	if _, err := w.b.st.vor("Commit", w.cerr); err != nil {
		w.zu = true
		return "", 0, err
	}
	w.zu = true
	ref := refVon(w.buf)
	w.b.mu.Lock()
	defer w.b.mu.Unlock()
	if _, da := w.b.dateien[ref]; !da || w.b.kaputt[ref] {
		w.b.dateien[ref] = bytes.Clone(w.buf)
		delete(w.b.kaputt, ref)
	}
	return ref, int64(len(w.buf)), nil
}

func (w *fakeWriter) Abort() error {
	w.zu = true
	return nil
}

// pruefe liefert den Fehler von Open und Verify zu ref.
func (b *fakeBlobs) pruefe(ref string) error {
	if !model.BlobReferenzGueltig(ref) {
		return model.ErrUngueltig
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.dateien[ref]; !ok {
		return outbound.ErrBlobFehlt
	}
	if b.kaputt[ref] {
		return outbound.ErrBlobBeschaedigt
	}
	return nil
}

func (b *fakeBlobs) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	b.p.melde("Open(" + ref + ")")
	if _, err := b.st.vor("Open", ctx.Err); err != nil {
		return nil, err
	}
	if err := b.pruefe(ref); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return io.NopCloser(bytes.NewReader(b.dateien[ref])), nil
}

func (b *fakeBlobs) Verify(ctx context.Context, ref string) error {
	b.p.melde("Verify(" + ref + ")")
	if _, err := b.st.vor("Verify", ctx.Err); err != nil {
		return err
	}
	return b.pruefe(ref)
}

func (b *fakeBlobs) List(ctx context.Context) ([]string, error) {
	b.p.melde("Blobs.List")
	if _, err := b.st.vor("Blobs.List", ctx.Err); err != nil {
		return nil, err
	}
	return b.refs(), nil
}

func (b *fakeBlobs) Remove(ctx context.Context, ref string) error {
	b.p.melde("Remove(" + ref + ")")
	if _, err := b.st.vor("Remove", ctx.Err); err != nil {
		return err
	}
	if !model.BlobReferenzGueltig(ref) {
		return model.ErrUngueltig
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.dateien, ref)
	delete(b.kaputt, ref)
	return nil
}

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
	sessions *fakeSessions
	dienste  *fakeDienste
	speicher *fakeSpeicher
	blobs    *fakeBlobs
	waechter *maintenance.BlobWaechter
	log      *fang
	konfig   maintenance.Konfig
	w        *maintenance.Wartung
}

// neu liefert eine Umgebung mit leeren Fakes, der Uhr bei t0, dem Wächter mit StandardKarenz und
// MaxAge 0; aendere ändert die Konfiguration vor New.
func neu(t *testing.T, aendere ...func(*maintenance.Konfig)) *welt {
	t.Helper()
	return neuAb(t, t0(), maintenance.StandardKarenz, aendere...)
}

// neuAb wie neu, mit der Uhr bei start und dem Wächter mit karenz.
func neuAb(t *testing.T, start time.Time, karenz time.Duration, aendere ...func(*maintenance.Konfig)) *welt {
	t.Helper()
	p, st := &ereignisse{}, neueSteuerung()
	u := &welt{t: t, p: p, st: st, log: &fang{}}
	u.uhr = &zaehlUhr{FestUhr: supporttest.NewFestUhr(start), st: st}
	u.sessions = &fakeSessions{st: st, p: p, alle: map[string]model.Session{}, refs: map[string][]string{}}
	u.dienste = &fakeDienste{st: st, p: p}
	u.speicher = &fakeSpeicher{st: st, p: p, sessions: u.sessions, zusatz: map[int][]string{}}
	u.blobs = &fakeBlobs{st: st, p: p, dateien: map[string][]byte{}, kaputt: map[string]bool{}}
	var err error
	if u.waechter, err = maintenance.NewBlobWaechter(u.blobs, u.uhr, karenz); err != nil {
		t.Fatal(err)
	}
	u.konfig = maintenance.Konfig{
		Sessions: u.sessions, Dienste: u.dienste, Speicher: u.speicher, Blobs: u.waechter,
		Uhr: u.uhr, Log: slog.New(u.log),
	}
	for _, f := range aendere {
		f(&u.konfig)
	}
	if u.w, err = maintenance.New(u.konfig); err != nil {
		t.Fatal(err)
	}
	return u
}

func mitMaxAge(d time.Duration) func(*maintenance.Konfig) {
	return func(k *maintenance.Konfig) { k.MaxAge = d }
}

// dienst legt einen Service an.
func (u *welt) dienst(ids ...string) {
	for _, id := range ids {
		u.dienste.dienste = append(u.dienste.dienste, model.Service{ID: id})
	}
}

// startup ruft Startup und hält jeden *Fehler gegen SPEC-123 Nr. 10.
func (u *welt) startup(ctx context.Context) (inbound.Wartungsbericht, error) {
	u.t.Helper()
	b, err := u.w.Startup(ctx)
	pruefeFehlerWert(u.t, err)
	return b, err
}

// cleanup ruft CleanupBlobs und hält jeden *Fehler gegen SPEC-123 Nr. 10.
func (u *welt) cleanup(ctx context.Context) (inbound.Loeschbericht, error) {
	u.t.Helper()
	b, err := u.w.CleanupBlobs(ctx)
	pruefeFehlerWert(u.t, err)
	return b, err
}

// pruefeFehlerWert hält einen *Fehler gegen SPEC-123 Nr. 10: erlaubter Status, Text aus Code und
// message ohne den Text der Ursache, Ursache über Unwrap.
func pruefeFehlerWert(t *testing.T, err error) {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		return
	}
	if !f.Code.StatusErlaubt(f.Status) {
		t.Errorf("Status %d bei %v nicht erlaubt", f.Status, f.Code)
	}
	if soll := f.Code.String() + ": " + f.Code.Message(); f.Error() != soll {
		t.Errorf("Error() = %q, erwartet %q", f.Error(), soll)
	}
	if f.Ursache == nil || !errors.Is(err, f.Ursache) || strings.Contains(f.Error(), marker) {
		t.Errorf("Ursache %v, Text %q", f.Ursache, f.Error())
	}
}

func leerBericht() inbound.Wartungsbericht {
	return inbound.Wartungsbericht{Befunde: []string{}, BlobsFehlend: []string{}, BlobsBeschaedigt: []string{}}
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

// commit legt Bytes über den Wächter ab und liefert die Referenz.
func (u *welt) commit(inhalt string) string {
	u.t.Helper()
	w, err := u.waechter.Begin(context.Background(), 1<<20)
	if err != nil {
		u.t.Fatal(err)
	}
	if _, err := w.Write([]byte(inhalt)); err != nil {
		u.t.Fatal(err)
	}
	ref, _, err := w.Commit()
	if err != nil {
		u.t.Fatal(err)
	}
	return ref
}

func ev(op, ref string) string { return op + "(" + ref + ")" }
