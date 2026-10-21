package record_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
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
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

var (
	errDisk       = errors.New("platte voll mit S3CR3T-7f")
	errVerbindung = errors.New("verbindung abgerissen")
	errAnders     = errors.New("anderer fehler")
)

// geheim steht an jeder redigierten Stelle der Tests und darf weder Blob noch Exchange noch
// Protokoll erreichen.
const geheim = "S3CR3T-7f"

// zeitgrenze begrenzt jedes Warten eines Tests auf ein Ereignis.
const zeitgrenze = 10 * time.Second

// --- Ereignisprotokoll von Blob- und Recording-Fake ---------------------------------------

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

// --- Blob-Fake nach SPEC-106 Nr. 1 bis 4 --------------------------------------------------

// fakeBlobs hält jedes je angebotene Byte und jeden committeten Blob und meldet jeden Commit
// auf commits und im Ereignisprotokoll. Ist haltCommit gesetzt, wartet jeder Commit vor seiner
// Prüfung des Kontexts, bis haltCommit schließt.
type fakeBlobs struct {
	mu         sync.Mutex
	prot       *ereignisse
	angeboten  []byte
	committed  map[string][]byte
	begins     int
	commits    chan string
	haltCommit chan struct{}
}

type fakeWriter struct {
	b      *fakeBlobs
	cerr   func() error
	grenze int64
	buf    []byte
}

func refVon(b []byte) string {
	h := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(h[:])
}

func (b *fakeBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.begins++
	return &fakeWriter{b: b, cerr: ctx.Err, grenze: grenze}, nil
}

func (w *fakeWriter) Write(p []byte) (int, error) {
	w.b.mu.Lock()
	w.b.angeboten = append(w.b.angeboten, p...)
	w.b.mu.Unlock()
	if err := w.cerr(); err != nil {
		return 0, err
	}
	if int64(len(w.buf)+len(p)) > w.grenze {
		return 0, outbound.ErrZuGross
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *fakeWriter) Commit() (string, int64, error) {
	if w.b.haltCommit != nil {
		<-w.b.haltCommit
	}
	if err := w.cerr(); err != nil {
		return "", 0, err
	}
	ref := refVon(w.buf)
	w.b.mu.Lock()
	w.b.committed[ref] = bytes.Clone(w.buf)
	w.b.mu.Unlock()
	w.b.prot.melde("Commit " + ref)
	w.b.commits <- ref
	return ref, int64(len(w.buf)), nil
}

func (*fakeWriter) Abort() error { return nil }

func (*fakeBlobs) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, outbound.ErrBlobFehlt
}
func (*fakeBlobs) Verify(context.Context, string) error   { return outbound.ErrBlobFehlt }
func (*fakeBlobs) List(context.Context) ([]string, error) { return nil, nil }
func (*fakeBlobs) Remove(context.Context, string) error   { return nil }

func (b *fakeBlobs) inhalt(ref string) (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v, ok := b.committed[ref]
	return string(v), ok
}

func (b *fakeBlobs) anzahlBegin() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.begins
}

func (b *fakeBlobs) alleAngeboten() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.angeboten)
}

// --- Session-Fake nach SPEC-105 Nr. 1 bis 6 -----------------------------------------------

type fakeSessions struct {
	mu         sync.Mutex
	sessions   map[string]model.Session
	getFehler  error
	listFehler error
	aufrufe    int
	// haltList hält List an: es meldet sich auf imList und wartet auf freiList.
	haltList bool
	imList   chan struct{}
	freiList chan struct{}
	// beimLesen läuft in jedem Get und List, etwa um die Uhr vorzurücken.
	beimLesen  func()
	unerwartet []string
}

// lese zählt den Aufruf, führt beimLesen aus und liefert den Fehler des Kontexts (SPEC-105 Nr. 13).
func (s *fakeSessions) lese(ctx context.Context) error {
	s.mu.Lock()
	s.aufrufe++
	beim := s.beimLesen
	s.mu.Unlock()
	if beim != nil {
		beim()
	}
	return ctx.Err()
}

func (s *fakeSessions) Get(ctx context.Context, id string) (model.Session, error) {
	if err := s.lese(ctx); err != nil {
		return model.Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.getFehler != nil {
		return model.Session{}, s.getFehler
	}
	ses, ok := s.sessions[id]
	if !ok {
		return model.Session{}, outbound.ErrNichtGefunden
	}
	return ses, nil
}

func (s *fakeSessions) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	s.mu.Lock()
	halt := s.haltList
	s.mu.Unlock()
	if halt {
		s.imList <- struct{}{}
		<-s.freiList
	}
	if err := s.lese(ctx); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listFehler != nil {
		return nil, s.listFehler
	}
	var out []model.Session
	for _, id := range slices.Sorted(maps.Keys(s.sessions)) {
		if s.sessions[id].ServiceID == serviceID {
			out = append(out, s.sessions[id])
		}
	}
	return out, nil
}

func (s *fakeSessions) melde(e string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unerwartet = append(s.unerwartet, e)
	return errors.New("unerwartet")
}

func (s *fakeSessions) Create(context.Context, model.Session) error { return s.melde("Create") }
func (s *fakeSessions) SetStatus(context.Context, string, model.SessionStatus) error {
	return s.melde("SetStatus")
}
func (s *fakeSessions) FindOpen(context.Context) ([]model.Session, error) {
	return nil, s.melde("FindOpen")
}
func (s *fakeSessions) Delete(context.Context, string) error { return s.melde("Delete") }

func (s *fakeSessions) setze(id, dienst string, status model.SessionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = model.Session{ID: id, ServiceID: dienst, Status: status, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func (s *fakeSessions) anzahlAufrufe() int { s.mu.Lock(); defer s.mu.Unlock(); return s.aufrufe }

// --- Recording-Fake nach SPEC-105 Nr. 7 bis 11 --------------------------------------------

type fakeExchanges struct {
	mu       sync.Mutex
	prot     *ereignisse
	sessions *fakeSessions
	liste    []model.Exchange
	// vorhanden sind Exchanges, die vor dem Test in der Ablage stehen.
	vorhanden map[string][]model.Exchange
	// appendFehler liefert der n-te Append (ab 0), bevor er etwas prüft.
	appendFehler map[int]error
	listFehler   error
	listAufrufe  int
	appends      int
	ctxFehler    []error
	pause        bool
	aktiv, hoch  int
	unerwartet   []string
	// haltAppend hält jedes Append an, mit haltNur nur die der Session haltNur: es meldet sich auf
	// imAppend und wartet, bis haltAppend schließt.
	haltAppend chan struct{}
	imAppend   chan struct{}
	haltNur    string
}

func (f *fakeExchanges) Append(ctx context.Context, e model.Exchange) error {
	f.mu.Lock()
	nr := f.appends
	f.appends++
	fehler := f.appendFehler[nr]
	f.ctxFehler = append(f.ctxFehler, ctx.Err())
	f.aktiv++
	f.hoch = max(f.hoch, f.aktiv)
	pause, halt := f.pause, f.haltAppend
	if f.haltNur != "" && e.SessionID != f.haltNur {
		halt = nil
	}
	f.mu.Unlock()
	f.prot.melde("Append " + e.ID)
	defer func() { f.mu.Lock(); f.aktiv--; f.mu.Unlock() }()
	if halt != nil {
		f.imAppend <- struct{}{}
		<-halt
	}
	if pause {
		time.Sleep(time.Millisecond)
	}
	if fehler != nil {
		return fehler
	}
	return f.fuegeEin(ctx, e)
}

// fuegeEin prüft in der Reihenfolge von SPEC-105 Nr. 8: Gültigkeit, Session, Kennung, Sequenz.
func (f *fakeExchanges) fuegeEin(ctx context.Context, e model.Exchange) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	f.sessions.mu.Lock()
	s, ok := f.sessions.sessions[e.SessionID]
	f.sessions.mu.Unlock()
	switch {
	case !ok:
		return outbound.ErrNichtGefunden
	case s.Status != model.SessionOpen:
		return outbound.ErrSessionGeschlossen
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if slices.ContainsFunc(f.liste, func(x model.Exchange) bool { return x.ID == e.ID }) {
		f.unerwartet = append(f.unerwartet, "Aufruf nach Append für "+e.ID)
		return outbound.ErrBereitsVorhanden
	}
	if e.SequenceNo != len(f.vonSession(e.SessionID))+1 {
		return outbound.ErrUngueltigeSequenz
	}
	f.liste = append(f.liste, e)
	return nil
}

func (f *fakeExchanges) vonSession(id string) []model.Exchange {
	out := slices.Clone(f.vorhanden[id])
	for _, e := range f.liste {
		if e.SessionID == id {
			out = append(out, e)
		}
	}
	return out
}

func (f *fakeExchanges) ListBySession(ctx context.Context, id string) ([]model.Exchange, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listAufrufe++
	if f.listFehler != nil {
		return nil, f.listFehler
	}
	return f.vonSession(id), ctx.Err()
}

func (f *fakeExchanges) Get(_ context.Context, id string) (model.Exchange, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unerwartet = append(f.unerwartet, "Get "+id)
	return model.Exchange{}, outbound.ErrNichtGefunden
}

func (f *fakeExchanges) angehaengt() []model.Exchange {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.liste)
}

// --- skriptgesteuerter UpstreamClient nach SPEC-120 Nr. 4, 5, 8 ----------------------------

// schritt ist ein Read des Upstream-Bodys: daten und err kommen zurück; warte hält ihn vor der
// Lieferung an, bis warte schließt oder der Kontext endet.
type schritt struct {
	daten string
	err   error
	warte <-chan struct{}
}

// fakeBody ist der Body einer Antwort: Read ist an den Kontext von Senden gebunden (SPEC-120 Nr. 5).
type fakeBody struct {
	mu       sync.Mutex
	schritte []schritt
	done     <-chan struct{}
	cerr     func() error
	closes   atomic.Int32
}

func koerper(teile ...string) *fakeBody {
	b := &fakeBody{}
	for _, t := range teile {
		b.schritte = append(b.schritte, schritt{daten: t})
	}
	return b
}

// dann hängt einen Schritt an, der bis warte wartet und daten mit err liefert.
func (b *fakeBody) dann(warte <-chan struct{}, daten string, err error) *fakeBody {
	b.schritte = append(b.schritte, schritt{daten: daten, err: err, warte: warte})
	return b
}

func (b *fakeBody) ctxFehler() error { return fmt.Errorf("upstream-fake: Lesen beendet: %w", b.cerr()) }

func (b *fakeBody) Read(p []byte) (int, error) {
	if b.cerr() != nil {
		return 0, b.ctxFehler()
	}
	b.mu.Lock()
	if len(b.schritte) == 0 {
		b.mu.Unlock()
		return 0, io.EOF
	}
	s := b.schritte[0]
	b.mu.Unlock()
	if s.warte != nil {
		select {
		case <-s.warte:
		case <-b.done:
			return 0, b.ctxFehler()
		}
	}
	n := copy(p, s.daten)
	b.mu.Lock()
	defer b.mu.Unlock()
	if n < len(s.daten) {
		b.schritte[0].daten, b.schritte[0].warte = s.daten[n:], nil
		return n, nil
	}
	b.schritte = b.schritte[1:]
	return n, s.err
}

func (b *fakeBody) Close() error {
	b.closes.Add(1)
	return nil
}

// plan steuert einen Aufruf von Senden.
type plan struct {
	// lies ist die Zahl der Bytes, die der Fake vom Request-Body liest; -1 heißt bis io.EOF.
	lies int
	// warteAufCtx wartet nach dem Lesen auf das Ende von ctx; nachCtx hält danach bis zu seinem Schließen.
	warteAufCtx bool
	nachCtx     chan struct{}
	// halt hält den Kopf bis zu seinem Schließen oder bis ctx endet.
	halt   chan struct{}
	fehler error
	// vorAntwort läuft, bevor der Fake antwortet oder mit fehler endet.
	vorAntwort func()
	status     int
	header     map[string][]string
	body       *fakeBody
}

func antwort(status int, header map[string][]string, b *fakeBody) *plan {
	return &plan{lies: -1, status: status, header: header, body: b}
}

// gesehen ist, was der Fake von einem Aufruf sah.
type gesehen struct {
	a    outbound.UpstreamAnfrage
	body string
	done <-chan struct{}
}

type fakeUpstream struct {
	mu      sync.Mutex
	plaene  []*plan
	aufrufe []gesehen
	// imSenden meldet jeden Aufruf; ctxEnde meldet, dass ein wartender Aufruf das Ende von ctx sah.
	imSenden chan struct{}
	ctxEnde  chan struct{}
}

func (u *fakeUpstream) Senden(ctx context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	u.mu.Lock()
	p := antwort(200, map[string][]string{"Content-Type": {"application/json"}}, koerper(`{"a":1}`))
	if len(u.plaene) > 0 {
		p, u.plaene = u.plaene[0], u.plaene[1:]
	}
	i := len(u.aufrufe)
	a.Header = maps.Clone(a.Header)
	u.aufrufe = append(u.aufrufe, gesehen{a: a, done: ctx.Done()})
	u.mu.Unlock()
	u.imSenden <- struct{}{}
	gelesen, err := liesBody(a, p.lies)
	u.mu.Lock()
	u.aufrufe[i].body = gelesen
	u.mu.Unlock()
	if err != nil {
		if ctx.Err() != nil {
			return outbound.UpstreamAntwort{}, fmt.Errorf("upstream-fake: %w", ctx.Err())
		}
		return outbound.UpstreamAntwort{}, fmt.Errorf("upstream-fake: Body: %w", err)
	}
	return u.antworte(ctx, p)
}

func liesBody(a outbound.UpstreamAnfrage, n int) (string, error) {
	if a.Body == nil || a.Laenge == 0 {
		return "", nil
	}
	if n < 0 {
		b, err := io.ReadAll(a.Body)
		return string(b), err
	}
	b := make([]byte, n)
	k, err := io.ReadFull(a.Body, b)
	return string(b[:k]), err
}

func (u *fakeUpstream) antworte(ctx context.Context, p *plan) (outbound.UpstreamAntwort, error) {
	if p.warteAufCtx {
		<-ctx.Done()
		u.ctxEnde <- struct{}{}
		if p.nachCtx != nil {
			<-p.nachCtx
		}
	}
	if p.halt != nil {
		select {
		case <-p.halt:
		case <-ctx.Done():
		}
	}
	if ctx.Err() != nil {
		return outbound.UpstreamAntwort{}, fmt.Errorf("upstream-fake: %w", ctx.Err())
	}
	if p.vorAntwort != nil {
		p.vorAntwort()
	}
	if p.fehler != nil {
		return outbound.UpstreamAntwort{}, p.fehler
	}
	p.body.done, p.body.cerr = ctx.Done(), ctx.Err
	return outbound.UpstreamAntwort{Status: p.status, Header: p.header, Body: p.body}, nil
}

func (u *fakeUpstream) plane(p ...*plan) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.plaene = append(u.plaene, p...)
}

func (u *fakeUpstream) gesehen() []gesehen {
	u.mu.Lock()
	defer u.mu.Unlock()
	return slices.Clone(u.aufrufe)
}

// --- Fang-Handler für slog ----------------------------------------------------------------

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

func (f *fang) Handle(_ context.Context, r slog.Record) error {
	e := eintrag{stufe: r.Level, text: r.Message, attrs: map[string]string{}}
	r.Attrs(func(a slog.Attr) bool {
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

func (f *fang) alle() []eintrag { f.mu.Lock(); defer f.mu.Unlock(); return slices.Clone(f.eintraege) }

// --- Umgebung -----------------------------------------------------------------------------

type umgebung struct {
	t         *testing.T
	uhr       *supporttest.FestUhr
	ids       *supporttest.FestIDs
	prot      *ereignisse
	blobs     *fakeBlobs
	sessions  *fakeSessions
	exchanges *fakeExchanges
	upstream  *fakeUpstream
	log       *fang
	konfig    record.Konfig
	r         *record.Recorder
}

func t0() time.Time { return time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC) }

func regeln() model.SchluesselRegeln {
	return model.SchluesselRegeln{
		VaryHeaders:       []string{"Accept", "Accept-Encoding"},
		VolatileHeaders:   []string{"Date"},
		RedactHeaders:     []string{"Authorization", "Cookie", "Set-Cookie"},
		RedactQueryParams: []string{"token"},
		InvalidEncoding:   model.KodierungLiteral,
		BodyMode:          model.BodyBytes,
	}
}

// neu baut die Umgebung: Dienst svc-a unter /a und svc-b unter /b, die Session s1 von svc-a
// offen, die Pipeline mit MaxBytes 256 und einer Frist von 1 h; aendere setzt die Konfiguration
// vor New.
func neu(t *testing.T, aendere ...func(*record.Konfig)) *umgebung {
	t.Helper()
	prot := &ereignisse{}
	u := &umgebung{
		t: t, uhr: supporttest.NewFestUhr(t0()), ids: supporttest.NewFestIDs(), prot: prot,
		blobs:    &fakeBlobs{prot: prot, committed: map[string][]byte{}, commits: make(chan string, 64)},
		sessions: &fakeSessions{sessions: map[string]model.Session{}, imList: make(chan struct{}, 8), freiList: make(chan struct{})},
		upstream: &fakeUpstream{imSenden: make(chan struct{}, 64), ctxEnde: make(chan struct{}, 64)},
		log:      &fang{},
	}
	u.exchanges = &fakeExchanges{prot: prot, sessions: u.sessions, vorhanden: map[string][]model.Exchange{},
		appendFehler: map[int]error{}}
	u.sessions.setze("s1", "svc-a", model.SessionOpen)
	pfad, err := model.ParseJSONPfad("$..password")
	if err != nil {
		t.Fatal(err)
	}
	p, err := capture.New(capture.Konfig{Blobs: u.blobs, Pfade: []model.JSONPfad{pfad}, MaxBytes: 256,
		QueueBytes: 1 << 20, Frist: time.Hour, MaxVerhaeltnis: 20})
	if err != nil {
		t.Fatal(err)
	}
	u.konfig = record.Konfig{
		Dienste:  []record.Dienst{{ID: "svc-a", Root: "/a", UpstreamURL: "http://a.example"}, {ID: "svc-b", Root: "/b", UpstreamURL: "http://b.example"}},
		Sessions: u.sessions, Exchanges: u.exchanges, Upstream: u.upstream, Pipeline: p, Regeln: regeln(),
		Uhr: u.uhr, IDs: u.ids, Log: slog.New(u.log),
	}
	for _, f := range aendere {
		f(&u.konfig)
	}
	if u.r, err = record.New(u.konfig); err != nil {
		t.Fatal(err)
	}
	return u
}

// anfrage ist ein Request ohne Body und ohne Header an svc-a.
func anfrage(methode, pfad string) inbound.Anfrage {
	return inbound.Anfrage{Methode: methode, Pfad: pfad, KorrelationsID: "cor-1", Header: map[string][]string{}}
}

// mitBody setzt Body, Laenge und Content-Type.
func mitBody(a inbound.Anfrage, typ, body string) inbound.Anfrage {
	a.Header = maps.Clone(a.Header)
	a.Header["Content-Type"] = []string{typ}
	a.Body, a.Laenge = strings.NewReader(body), int64(len(body))
	return a
}

// record ruft Record; jeden *Fehler hält es gegen SPEC-121 Nr. 14 (B25): erlaubter Status, Text aus
// Code und message.
func (u *umgebung) record(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	ant, err := u.r.Record(ctx, a)
	pruefeFehlerWert(u.t, err)
	return ant, err
}

func pruefeFehlerWert(t *testing.T, err error) {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		return
	}
	if !f.Code.StatusErlaubt(f.Status) {
		t.Errorf("%v trägt den Status %d", f.Code, f.Status)
	}
	if soll := f.Code.String() + ": " + f.Code.Message(); f.Error() != soll {
		t.Errorf("Error() = %q, erwartet %q", f.Error(), soll)
	}
}

// rundlauf führt einen Request aus, liest die Antwort ganz und schließt sie.
func (u *umgebung) rundlauf(a inbound.Anfrage) (inbound.Antwort, string) {
	u.t.Helper()
	ant, err := u.record(context.Background(), a)
	if err != nil {
		u.t.Fatalf("Record: %v", err)
	}
	return ant, liesUndSchliesse(u.t, ant.Body)
}

func liesUndSchliesse(t *testing.T, b io.ReadCloser) string {
	t.Helper()
	var daten []byte
	mitZeitgrenze(t, "Lesen und Close", func() {
		daten, _ = io.ReadAll(b)
		if err := b.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return string(daten)
}

func schliesse(t *testing.T, b io.Closer) {
	t.Helper()
	mitZeitgrenze(t, "Close", func() {
		if err := b.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
}

// mitZeitgrenze führt f aus und bricht den Test ab, wenn f nicht innerhalb der Zeitgrenze endet.
func mitZeitgrenze(t *testing.T, was string, f func()) {
	t.Helper()
	fertig := make(chan struct{})
	go func() {
		defer close(fertig)
		f()
	}()
	select {
	case <-fertig:
	case <-time.After(zeitgrenze):
		t.Fatalf("%s endet nicht innerhalb von %v", was, zeitgrenze)
	}
}

// empfange wartet auf ein Ereignis auf c, höchstens die Zeitgrenze.
func empfange[T any](t *testing.T, was string, c <-chan T) T {
	t.Helper()
	select {
	case v := <-c:
		return v
	case <-time.After(zeitgrenze):
		t.Fatalf("%s kam nicht innerhalb von %v", was, zeitgrenze)
	}
	var null T
	return null
}

// nebenbei führt f in einer Goroutine aus und liefert seinen Ausgang auf dem Kanal.
func nebenbei[T any](f func() T) <-chan T {
	c := make(chan T, 1)
	go func() { c <- f() }()
	return c
}

type ergebnis struct {
	ant inbound.Antwort
	err error
}

func (u *umgebung) recordNebenbei(ctx context.Context, a inbound.Anfrage) <-chan ergebnis {
	return nebenbei(func() ergebnis {
		ant, err := u.record(ctx, a)
		return ergebnis{ant, err}
	})
}

type drainErgebnis struct {
	b   inbound.DrainBericht
	err error
}

func (u *umgebung) drainNebenbei(ctx context.Context, id string, frist time.Duration) <-chan drainErgebnis {
	return nebenbei(func() drainErgebnis {
		b, err := u.r.Drain(ctx, id, frist)
		return drainErgebnis{b, err}
	})
}

// letzter liefert den einzigen angehängten Exchange.
func (u *umgebung) letzter() model.Exchange {
	u.t.Helper()
	l := u.exchanges.angehaengt()
	if len(l) != 1 {
		u.t.Fatalf("%d Exchanges angehängt, erwartet 1", len(l))
	}
	return l[0]
}

// alsFehler liefert den *Fehler aus err oder bricht ab.
func alsFehler(t *testing.T, err error) *inbound.Fehler {
	t.Helper()
	var f *inbound.Fehler
	if !errors.As(err, &f) {
		t.Fatalf("kein *Fehler: %v", err)
	}
	return f
}

func pruefeCode(t *testing.T, err error, c model.Code, status int) {
	t.Helper()
	f := alsFehler(t, err)
	if f.Code != c || f.Status != status {
		t.Errorf("Fehler %v mit %d, erwartet %v mit %d", f.Code, f.Status, c, status)
	}
}

func kennung(praefix string, n int) string { return fmt.Sprintf("%s-%032d", praefix, n) }
