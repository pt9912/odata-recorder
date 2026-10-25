package baselineimport_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

const (
	dienstID  = "catalog"
	upstream  = "https://up.example/odata/v4"
	sitzungID = "ses-1"
	marker    = "M-7f"
)

var errPlatte = errors.New("platte voll mit " + marker)

func t0() time.Time { return time.Date(2026, time.October, 25, 12, 0, 0, 0, time.UTC) }

// --- Protokoll: Zählung, Argumente, Fehlervorgaben und Haltepunkte aller Doppel -------------

type protokoll struct {
	mu       sync.Mutex
	ereignis []string
	zaehler  map[string]int
	ctxs     map[string][]context.Context
	args     map[string][][]any
	fehler   map[string]map[int]error
	haken    map[string]func(n int)
}

func neuesProtokoll() *protokoll {
	return &protokoll{
		zaehler: map[string]int{}, ctxs: map[string][]context.Context{}, args: map[string][][]any{},
		fehler: map[string]map[int]error{}, haken: map[string]func(int){},
	}
}

// betrete zählt den Aufruf, hält Kontext und Argumente fest und liefert die vorgegebene Antwort.
func (p *protokoll) betrete(op string, kontext any, args ...any) (int, error) {
	ctx, _ := kontext.(context.Context)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.zaehler[op]++
	n := p.zaehler[op]
	p.ereignis = append(p.ereignis, op)
	p.ctxs[op] = append(p.ctxs[op], ctx)
	p.args[op] = append(p.args[op], args)
	if err := p.fehler[op][n]; err != nil {
		return n, err
	}
	return n, p.fehler[op][0]
}

// nach ruft den Haltepunkt der Operation, nachdem sie ihr Ergebnis bestimmt hat.
func (p *protokoll) nach(op string, n int) {
	p.mu.Lock()
	h := p.haken[op]
	p.mu.Unlock()
	if h != nil {
		h(n)
	}
}

// scheitere lässt den n-ten Aufruf der Operation scheitern (n gleich 0: jeden).
func (p *protokoll) scheitere(op string, n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fehler[op] == nil {
		p.fehler[op] = map[int]error{}
	}
	p.fehler[op][n] = err
}

func (p *protokoll) anzahl(op string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.zaehler[op]
}

func (p *protokoll) folge() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.ereignis)
}

func (p *protokoll) ctxVon(op string) []context.Context {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.ctxs[op])
}

func (p *protokoll) argsVon(op string) [][]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.args[op])
}

func (p *protokoll) setzeHaken(op string, h func(n int)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.haken[op] = h
}

// verbotene sind die Operationen der Ports, die der Dienst nie rufen darf.
func verbotene() []string {
	return []string{
		"Sessions.Create", "Sessions.SetStatus", "Sessions.List", "Sessions.FindOpen", "Sessions.Delete",
		"Exchanges.Append", "Exchanges.Get", "Blobs.Verify", "Blobs.List", "Blobs.Remove",
		"Imports.BaselinesOf", "Imports.Sources",
		"Snapshots.List", "Snapshots.CreateWorkspace", "Snapshots.GetWorkspace", "Snapshots.ListWorkspace",
		"Metadaten.Create", "Metadaten.Get", "Metadaten.List",
		"IDs.SessionID", "IDs.ExchangeID", "IDs.KorrelationsID", "IDs.WorkspaceID",
	}
}

func (p *protokoll) verbotenGerufen() int {
	n := 0
	for _, op := range verbotene() {
		n += p.anzahl(op)
	}
	return n
}

// --- SessionRepository -------------------------------------------------------------------

type fakeSessions struct {
	p  *protokoll
	mu sync.Mutex
	m  map[string]model.Session
}

func (f *fakeSessions) Get(ctx context.Context, id string) (model.Session, error) {
	n, err := f.p.betrete("Sessions.Get", ctx, id)
	defer f.p.nach("Sessions.Get", n)
	if err != nil {
		return model.Session{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.m[id]
	if !ok {
		return model.Session{}, outbound.ErrNichtGefunden
	}
	return s, nil
}

func (f *fakeSessions) hat(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.m[id]
	return ok
}

func (f *fakeSessions) loesche(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, id)
}

func (f *fakeSessions) Create(ctx context.Context, s model.Session) error {
	_, err := f.p.betrete("Sessions.Create", ctx, s)
	return err
}

func (f *fakeSessions) SetStatus(ctx context.Context, id string, st model.SessionStatus) error {
	_, err := f.p.betrete("Sessions.SetStatus", ctx, id, st)
	return err
}

func (f *fakeSessions) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	_, err := f.p.betrete("Sessions.List", ctx, serviceID)
	return nil, err
}

func (f *fakeSessions) FindOpen(ctx context.Context) ([]model.Session, error) {
	_, err := f.p.betrete("Sessions.FindOpen", ctx)
	return nil, err
}

func (f *fakeSessions) Delete(ctx context.Context, id string) error {
	_, err := f.p.betrete("Sessions.Delete", ctx, id)
	return err
}

// --- RecordingRepository -----------------------------------------------------------------

type fakeExchanges struct {
	p  *protokoll
	mu sync.Mutex
	m  map[string][]model.Exchange
}

func (f *fakeExchanges) ListBySession(ctx context.Context, id string) ([]model.Exchange, error) {
	n, err := f.p.betrete("Exchanges.ListBySession", ctx, id)
	defer f.p.nach("Exchanges.ListBySession", n)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.m[id]), nil
}

func (f *fakeExchanges) Append(ctx context.Context, e model.Exchange) error {
	_, err := f.p.betrete("Exchanges.Append", ctx, e)
	return err
}

func (f *fakeExchanges) Get(ctx context.Context, id string) (model.Exchange, error) {
	_, err := f.p.betrete("Exchanges.Get", ctx, id)
	return model.Exchange{}, err
}

// --- BlobRepository mit CaptureWriter ------------------------------------------------------

// strom ist der Leser, den Open liefert. Er trägt jede Schnittstelle des Lesers der Bytes
// (`*bytes.Reader` ist eingebettet) und zählt Close.
type strom struct {
	*bytes.Reader
	ref string
	p   *protokoll
	mu  sync.Mutex
	zu  int
}

func (s *strom) Close() error {
	s.mu.Lock()
	s.zu++
	s.mu.Unlock()
	_, err := s.p.betrete("Close", nil, s.ref)
	return err
}

func (s *strom) geschlossen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.zu
}

type fakeBlobs struct {
	p       *protokoll
	mu      sync.Mutex
	inhalt  map[string][]byte
	geoeff  []*strom
	writers []*fakeWriter
	// stueck begrenzt die Bytes, die ein Write annimmt (0: alle); laenge lässt Write ohne Fehler eine Länge
	// melden, die der Vertrag von io.Writer nicht zulässt (0: die Länge der Bytes).
	stueck int
	laenge func(n int) int
}

func refVon(b []byte) string {
	s := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(s[:])
}

func (f *fakeBlobs) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	n, err := f.p.betrete("Blobs.Open", ctx, ref)
	defer f.p.nach("Blobs.Open", n)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.inhalt[ref]
	if !ok {
		return nil, outbound.ErrBlobFehlt
	}
	s := &strom{Reader: bytes.NewReader(b), ref: ref, p: f.p}
	f.geoeff = append(f.geoeff, s)
	return s, nil
}

func (f *fakeBlobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	n, err := f.p.betrete("Blobs.Begin", ctx, grenze)
	defer f.p.nach("Blobs.Begin", n)
	if err != nil {
		return nil, err
	}
	w := &fakeWriter{f: f}
	f.mu.Lock()
	f.writers = append(f.writers, w)
	f.mu.Unlock()
	return w, nil
}

func (f *fakeBlobs) Verify(ctx context.Context, ref string) error {
	_, err := f.p.betrete("Blobs.Verify", ctx, ref)
	return err
}

func (f *fakeBlobs) List(ctx context.Context) ([]string, error) {
	_, err := f.p.betrete("Blobs.List", ctx)
	return nil, err
}

func (f *fakeBlobs) Remove(ctx context.Context, ref string) error {
	_, err := f.p.betrete("Blobs.Remove", ctx, ref)
	return err
}

func (f *fakeBlobs) strome() []*strom {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.geoeff)
}

type fakeWriter struct {
	f      *fakeBlobs
	mu     sync.Mutex
	buf    bytes.Buffer
	aufruf int
	fertig bool
}

func (w *fakeWriter) Write(b []byte) (int, error) {
	if _, err := w.f.p.betrete("Write", nil, slices.Clone(b)); err != nil {
		return 0, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.aufruf++
	if w.f.laenge != nil {
		if w.aufruf > 1000 {
			return 0, errors.New("1000 Aufrufe ohne Fortschritt")
		}
		return w.f.laenge(len(b)), nil
	}
	if s := w.f.stueck; s > 0 && len(b) > s {
		b = b[:s]
	}
	return w.buf.Write(b)
}

func (w *fakeWriter) Commit() (string, int64, error) {
	if _, err := w.f.p.betrete("Commit", nil); err != nil {
		return "", 0, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.fertig = true
	ref := refVon(w.buf.Bytes())
	w.f.mu.Lock()
	w.f.inhalt[ref] = bytes.Clone(w.buf.Bytes())
	w.f.mu.Unlock()
	return ref, int64(w.buf.Len()), nil
}

func (w *fakeWriter) Abort() error {
	_, err := w.f.p.betrete("Abort", nil)
	return err
}

func (w *fakeWriter) bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return bytes.Clone(w.buf.Bytes())
}

// --- MetadataRepository ------------------------------------------------------------------

type setzung struct{ id, blob, hash string }

type fakeMetadaten struct {
	p       *protokoll
	mu      sync.Mutex
	gesetzt []setzung
}

func (f *fakeMetadaten) SetMetadata(ctx context.Context, id, blob, hash string) error {
	n, err := f.p.betrete("Metadaten.SetMetadata", ctx, id, blob, hash)
	defer f.p.nach("Metadaten.SetMetadata", n)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gesetzt = append(f.gesetzt, setzung{id, blob, hash})
	return nil
}

func (f *fakeMetadaten) Create(ctx context.Context, s model.Service) error {
	_, err := f.p.betrete("Metadaten.Create", ctx, s)
	return err
}

func (f *fakeMetadaten) Get(ctx context.Context, id string) (model.Service, error) {
	_, err := f.p.betrete("Metadaten.Get", ctx, id)
	return model.Service{}, err
}

func (f *fakeMetadaten) List(ctx context.Context) ([]model.Service, error) {
	_, err := f.p.betrete("Metadaten.List", ctx)
	return nil, err
}

// --- SnapshotRepository und ImportRepository ---------------------------------------------------

type fakeSnapshots struct {
	p  *protokoll
	mu sync.Mutex
	m  map[string]model.BaselineSnapshot
}

func (f *fakeSnapshots) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	n, err := f.p.betrete("Snapshots.Get", ctx, id)
	defer f.p.nach("Snapshots.Get", n)
	if err != nil {
		return model.BaselineSnapshot{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.m[id]
	if !ok {
		return model.BaselineSnapshot{}, outbound.ErrNichtGefunden
	}
	return s, nil
}

func (f *fakeSnapshots) List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error) {
	_, err := f.p.betrete("Snapshots.List", ctx, serviceID)
	return nil, err
}

func (f *fakeSnapshots) CreateWorkspace(ctx context.Context, s model.WorkspaceSnapshot) error {
	_, err := f.p.betrete("Snapshots.CreateWorkspace", ctx, s)
	return err
}

func (f *fakeSnapshots) GetWorkspace(ctx context.Context, workspaceID, id string) (model.WorkspaceSnapshot, error) {
	_, err := f.p.betrete("Snapshots.GetWorkspace", ctx, workspaceID, id)
	return model.WorkspaceSnapshot{}, err
}

func (f *fakeSnapshots) ListWorkspace(ctx context.Context, workspaceID string) ([]model.WorkspaceSnapshot, error) {
	_, err := f.p.betrete("Snapshots.ListWorkspace", ctx, workspaceID)
	return nil, err
}

type fakeImports struct {
	p         *protokoll
	sessions  *fakeSessions
	snapshots *fakeSnapshots
	mu        sync.Mutex
	m         map[string]model.Import
	pubs      []model.Publication
	// uebersteuere liefert statt der Ordnung von SPEC-107 Nr. 7 die Antwort des Tests.
	uebersteuere func(model.Publication) (model.Import, bool, error)
}

func (f *fakeImports) Get(ctx context.Context, id string) (model.Import, error) {
	n, err := f.p.betrete("Imports.Get", ctx, id)
	defer f.p.nach("Imports.Get", n)
	if err != nil {
		return model.Import{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	i, ok := f.m[id]
	if !ok {
		return model.Import{}, outbound.ErrNichtGefunden
	}
	return i, nil
}

// Publish ordnet wie SPEC-107 Nr. 7: genau ein gleichzeitiger Aufruf mit gleichem Hash schreibt,
// die anderen liefern dessen Import; eine bekannte Kennung mit anderem Hash ist ErrBereitsVorhanden.
func (f *fakeImports) Publish(ctx context.Context, pub model.Publication) (model.Import, bool, error) {
	n, err := f.p.betrete("Imports.Publish", ctx, pub)
	defer f.p.nach("Imports.Publish", n)
	if err != nil {
		return model.Import{}, false, err
	}
	if err := pub.Validate(); err != nil {
		return model.Import{}, false, err
	}
	for _, id := range pub.SourceSessionIDs {
		if !f.sessions.hat(id) {
			return model.Import{}, false, outbound.ErrNichtGefunden
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pubs = append(f.pubs, pub)
	if f.uebersteuere != nil {
		return f.uebersteuere(pub)
	}
	if alt, ok := f.m[pub.Import.ID]; ok {
		if alt.RequestHash != pub.Import.RequestHash {
			return model.Import{}, false, outbound.ErrBereitsVorhanden
		}
		return alt, false, nil
	}
	for _, alt := range f.m {
		if alt.ServiceID == pub.Import.ServiceID && alt.RequestHash == pub.Import.RequestHash {
			return alt, false, nil
		}
	}
	f.m[pub.Import.ID] = pub.Import
	f.snapshots.mu.Lock()
	f.snapshots.m[pub.Snapshot.ID] = pub.Snapshot
	f.snapshots.mu.Unlock()
	return pub.Import, true, nil
}

func (f *fakeImports) BaselinesOf(ctx context.Context, sessionID string) ([]string, error) {
	_, err := f.p.betrete("Imports.BaselinesOf", ctx, sessionID)
	return nil, err
}

func (f *fakeImports) Sources(ctx context.Context, baselineID string) ([]string, error) {
	_, err := f.p.betrete("Imports.Sources", ctx, baselineID)
	return nil, err
}

func (f *fakeImports) veroeffentlichungen() []model.Publication {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.pubs)
}

// --- Codecs ------------------------------------------------------------------------------

var errFremderStrom = errors.New("src ist nicht der Strom von Open")

type metaVorgabe struct {
	modell  model.ServiceModel
	bericht model.ValidierungsBericht
	err     error
}

type metaAufruf struct {
	src    io.Reader
	grenze int64
}

type fakeMetaCodec struct {
	p       *protokoll
	mu      sync.Mutex
	vorgabe metaVorgabe
	nachRef map[string]metaVorgabe
	fremd   int
	aufrufe []metaAufruf
}

func (c *fakeMetaCodec) ParseMetadata(ctx context.Context, src io.Reader, grenze int64) (model.ServiceModel, model.ValidierungsBericht, error) {
	n, err := c.p.betrete("ParseMetadata", ctx, src, grenze)
	defer c.p.nach("ParseMetadata", n)
	c.mu.Lock()
	c.aufrufe = append(c.aufrufe, metaAufruf{src, grenze})
	v := c.vorgabe
	s, ok := src.(*strom)
	if !ok {
		c.fremd++
		c.mu.Unlock()
		return model.ServiceModel{}, model.ValidierungsBericht{}, errFremderStrom
	}
	if x, ok := c.nachRef[s.ref]; ok {
		v = x
	}
	c.mu.Unlock()
	if err != nil {
		return model.ServiceModel{}, model.ValidierungsBericht{}, err
	}
	if v.err != nil {
		return model.ServiceModel{}, model.ValidierungsBericht{}, v.err
	}
	return v.modell, v.bericht, nil
}

type datenVorgabe struct {
	seite  model.EntitySeite
	entity model.Entity
	err    error
}

type datenAufruf struct {
	src       io.Reader
	m         model.ServiceModel
	entitySet string
	grenze    int64
}

type fakeDatenCodec struct {
	p       *protokoll
	mu      sync.Mutex
	nachRef map[string]datenVorgabe
	fremd   int
	fehlt   int
	aufrufe map[string][]datenAufruf
}

func (c *fakeDatenCodec) hole(op string, src io.Reader, m model.ServiceModel, set string, grenze int64) (datenVorgabe, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aufrufe[op] = append(c.aufrufe[op], datenAufruf{src, m, set, grenze})
	s, ok := src.(*strom)
	if !ok {
		c.fremd++
		return datenVorgabe{}, errFremderStrom
	}
	v, ok := c.nachRef[s.ref]
	if !ok {
		c.fehlt++
		return datenVorgabe{}, errors.New("keine Vorgabe für " + s.ref)
	}
	return v, nil
}

func (c *fakeDatenCodec) ParseCollection(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (model.EntitySeite, error) {
	n, err := c.p.betrete("ParseCollection", ctx, src, m, entitySet, grenze)
	defer c.p.nach("ParseCollection", n)
	v, herr := c.hole("ParseCollection", src, m, entitySet, grenze)
	switch {
	case err != nil:
		return model.EntitySeite{}, err
	case herr != nil:
		return model.EntitySeite{}, herr
	case v.err != nil:
		return model.EntitySeite{}, v.err
	}
	return v.seite, nil
}

func (c *fakeDatenCodec) ParseEntity(ctx context.Context, src io.Reader, m model.ServiceModel, entitySet string, grenze int64) (model.Entity, error) {
	n, err := c.p.betrete("ParseEntity", ctx, src, m, entitySet, grenze)
	defer c.p.nach("ParseEntity", n)
	v, herr := c.hole("ParseEntity", src, m, entitySet, grenze)
	switch {
	case err != nil:
		return model.Entity{}, err
	case herr != nil:
		return model.Entity{}, herr
	case v.err != nil:
		return model.Entity{}, v.err
	}
	return v.entity, nil
}

// --- Uhr und Kennungen mit Zählung -----------------------------------------------------------

type zaehlUhr struct {
	p *protokoll
	u *supporttest.FestUhr
}

func (z zaehlUhr) Now() time.Time {
	_, _ = z.p.betrete("Uhr.Now", nil)
	return z.u.Now()
}

type zaehlIDs struct {
	p *protokoll
	g *supporttest.FestIDs
}

func (z zaehlIDs) SnapshotID() string {
	_, _ = z.p.betrete("IDs.SnapshotID", nil)
	return z.g.SnapshotID()
}

func (z zaehlIDs) SessionID() string {
	_, _ = z.p.betrete("IDs.SessionID", nil)
	return z.g.SessionID()
}

func (z zaehlIDs) ExchangeID() string {
	_, _ = z.p.betrete("IDs.ExchangeID", nil)
	return z.g.ExchangeID()
}

func (z zaehlIDs) KorrelationsID() string {
	_, _ = z.p.betrete("IDs.KorrelationsID", nil)
	return z.g.KorrelationsID()
}

func (z zaehlIDs) WorkspaceID() string {
	_, _ = z.p.betrete("IDs.WorkspaceID", nil)
	return z.g.WorkspaceID()
}

// --- Die Welt eines Tests --------------------------------------------------------------------

type welt struct {
	t         *testing.T
	p         *protokoll
	sessions  *fakeSessions
	exchanges *fakeExchanges
	blobs     *fakeBlobs
	metadaten *fakeMetadaten
	snapshots *fakeSnapshots
	imports   *fakeImports
	meta      *fakeMetaCodec
	daten     *fakeDatenCodec
	konfig    baselineimport.Konfig
	imp       *baselineimport.Importer
}

func regeln() model.SchluesselRegeln {
	return model.SchluesselRegeln{
		VaryHeaders: []string{"Accept"}, InvalidEncoding: model.KodierungLiteral, BodyMode: model.BodyBytes,
		RedactHeaders: []string{"Authorization"},
	}
}

// neu baut die Welt mit dem Referenzmodell R und einer vollständigen Session ohne Exchanges.
func neu(t *testing.T) *welt {
	t.Helper()
	p := neuesProtokoll()
	w := &welt{t: t, p: p}
	w.sessions = &fakeSessions{p: p, m: map[string]model.Session{}}
	w.exchanges = &fakeExchanges{p: p, m: map[string][]model.Exchange{}}
	w.blobs = &fakeBlobs{p: p, inhalt: map[string][]byte{}}
	w.metadaten = &fakeMetadaten{p: p}
	w.snapshots = &fakeSnapshots{p: p, m: map[string]model.BaselineSnapshot{}}
	w.imports = &fakeImports{p: p, sessions: w.sessions, snapshots: w.snapshots, m: map[string]model.Import{}}
	w.meta = &fakeMetaCodec{p: p, vorgabe: metaVorgabe{modell: modellR(), bericht: model.ValidierungsBericht{Funde: []model.Fund{}}}, nachRef: map[string]metaVorgabe{}}
	w.daten = &fakeDatenCodec{p: p, nachRef: map[string]datenVorgabe{}, aufrufe: map[string][]datenAufruf{}}
	w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: dienstID, CreatedAt: t0(), Status: model.SessionComplete}
	w.konfig = baselineimport.Konfig{
		Dienste:  []baselineimport.Dienst{{ID: dienstID, UpstreamURL: upstream}},
		Sessions: w.sessions, Exchanges: w.exchanges, Blobs: w.blobs, Metadaten: w.metadaten,
		Snapshots: w.snapshots, Imports: w.imports, MetadatenCodec: w.meta, DatenCodec: w.daten,
		Regeln: regeln(), MaxCaptureBody: 8 << 20,
		Uhr: zaehlUhr{p, supporttest.NewFestUhr(t0())}, IDs: zaehlIDs{p, supporttest.NewFestIDs()},
	}
	t.Cleanup(func() {
		if w.meta.fremd+w.daten.fremd > 0 {
			t.Errorf("ein Codec erhielt einen Strom, der nicht der von Open ist (Hülle): %d", w.meta.fremd+w.daten.fremd)
		}
		if w.daten.fehlt > 0 {
			t.Errorf("der Fake des Daten-Codecs hatte für %d Aufrufe keine Vorgabe", w.daten.fehlt)
		}
	})
	return w
}

// start baut den Dienst aus der Konfig der Welt.
func (w *welt) start() *welt {
	w.t.Helper()
	imp, err := baselineimport.New(w.konfig)
	if err != nil {
		w.t.Fatalf("New: %v", err)
	}
	w.imp = imp
	return w
}

func anfrage() inbound.ImportAnfrage {
	return inbound.ImportAnfrage{ImportID: "imp-1", ServiceID: dienstID, SourceSessionIDs: []string{sitzungID}}
}

func (w *welt) importiere(a inbound.ImportAnfrage) (inbound.ImportErgebnis, error) {
	w.t.Helper()
	if w.imp == nil {
		w.start()
	}
	return w.imp.Import(context.Background(), a)
}

// gesamt zählt alle Aufrufe der Ports, der Codecs, der Uhr und des Generators.
func (w *welt) gesamt() int {
	w.p.mu.Lock()
	defer w.p.mu.Unlock()
	n := 0
	for _, c := range w.p.zaehler {
		n += c
	}
	return n
}

// --- Das Referenzmodell R ------------------------------------------------------------------------

const (
	typProdukt  = "Shop.Product"
	typGeschenk = "Shop.Gift"
	typBestell  = "Shop.Order"
	typEtikett  = "Shop.Tag"
)

func produktProps() []model.Property {
	return []model.Property{
		{Name: "Id", EdmType: "Edm.Int32"}, {Name: "Name", EdmType: "Edm.String", Nullable: true},
		{Name: "Price", EdmType: "Edm.Decimal", Nullable: true}, {Name: "Rating", EdmType: "Edm.Double", Nullable: true},
	}
}

func modellR() model.ServiceModel {
	nav := []model.NavigationProperty{{Name: "Category", Target: "Shop.Category"}}
	typen := []model.EntityType{
		{Name: typProdukt, Keys: []string{"Id"}, Properties: produktProps(), NavigationProperties: nav},
		{Name: typGeschenk, BaseType: typProdukt, Keys: []string{"Id"}, Properties: produktProps(), NavigationProperties: nav},
		{Name: typBestell, Keys: []string{"OrderID", "LineNo"}, Properties: []model.Property{
			{Name: "OrderID", EdmType: "Edm.Int32"}, {Name: "LineNo", EdmType: "Edm.Int32"}, {Name: "Qty", EdmType: "Edm.Int32", Nullable: true},
		}},
		{Name: typEtikett, Keys: []string{"Code"}, Properties: []model.Property{
			{Name: "Code", EdmType: "Edm.String"}, {Name: "Label", EdmType: "Edm.String", Nullable: true},
		}},
	}
	sets := []model.EntitySet{{Name: "Products", EntityType: typProdukt}, {Name: "Orders", EntityType: typBestell}, {Name: "Tags", EntityType: typEtikett}}
	return model.NeuesServiceModel("4.0", []string{"Shop"}, typen, sets, nil)
}

// --- Entities und Antworten ---------------------------------------------------------------------

func wert(art model.WertArt, text string) model.Wert { return model.Wert{Art: art, Text: text} }

func pw(name string, art model.WertArt, text string) model.PropertyWert {
	return model.PropertyWert{Name: name, Wert: wert(art, text)}
}

func ent(typ, schluessel string, props ...model.PropertyWert) model.Entity {
	return model.Entity{Typ: typ, Schluessel: schluessel, Properties: props, Uebersprungen: []model.Uebersprungen{}}
}

// produkt ist die Entity Product mit Id, Name und Price.
func produkt(id int, name, preis string) model.Entity {
	e := ent(typProdukt, fmt.Sprint(id), pw("Id", model.ArtInt32, fmt.Sprint(id)), pw("Name", model.ArtString, name), pw("Price", model.ArtDecimal, preis))
	e.Uebersprungen = []model.Uebersprungen{{Name: "Rating", Grund: model.UebersprungTyp}, {Name: "Category", Grund: model.UebersprungNavigation}}
	return e
}

func seite(es ...model.Entity) model.EntitySeite {
	if es == nil {
		es = []model.Entity{}
	}
	return model.EntitySeite{Entities: es}
}

func mitLink(s model.EntitySeite, link string) model.EntitySeite {
	s.FolgeLink = model.FolgeLink{Vorhanden: true, Wert: link}
	return s
}

// --- Der Baukasten für Exchanges ------------------------------------------------------------------

type bauer struct {
	w   *welt
	ses string
	seq int
}

func (w *welt) sitzung() *bauer { return &bauer{w: w, ses: sitzungID} }

type aenderung func(*model.Exchange)

func jsonKopf() map[string][]string {
	return map[string][]string{"Content-Type": {"application/json;odata.metadata=minimal"}}
}

// roh legt einen Exchange GET <pfad>?<query> mit Antwort 200, complete und Body present an; der
// Body hat einen eigenen Inhalt, damit jeder Exchange eine eigene Blob-Referenz trägt.
func (b *bauer) roh(pfad, query string, as ...aenderung) model.Exchange {
	b.seq++
	e := model.Exchange{
		ID: fmt.Sprintf("exc-%03d", b.seq), SessionID: b.ses, SequenceNo: b.seq, Method: "GET", Path: pfad, RawQuery: query,
		StartedAt: t0(), FinishedAt: t0(), RequestBody: model.BodyAbsent, ResponseStatus: 200, ResponseHeaders: jsonKopf(),
		ResponseBody: model.BodyPresent, Capture: model.CaptureComplete, RequestRedact: model.RedactionNone,
		ResponseRedact: model.RedactionNone, Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte,
	}
	for _, a := range as {
		if a != nil {
			a(&e)
		}
	}
	if e.MatchKey == "" {
		k, err := model.BildeSchluessel(model.SchluesselEingabe{
			Method: e.Method, ServiceID: dienstID, Path: e.Path, RawQuery: e.RawQuery, Header: e.RequestHeaders, RequestRedaction: model.RedactionNone,
		}, regeln())
		if err == nil {
			e.MatchKey, e.KeyRedacted = k.MatchKey, e.KeyRedacted || k.KeyRedacted
		} else {
			e.MatchKey = "sha256:" + strings.Repeat("0", 64)
		}
	}
	if e.ResponseBody == model.BodyPresent && e.ResponseBlob == "" {
		inhalt := []byte(fmt.Sprintf("Antwort %s %d", e.ID, e.SequenceNo))
		e.ResponseBlob = refVon(inhalt)
		b.w.blobs.inhalt[e.ResponseBlob] = inhalt
	}
	b.w.exchanges.mu.Lock()
	b.w.exchanges.m[b.ses] = append(b.w.exchanges.m[b.ses], e)
	b.w.exchanges.mu.Unlock()
	return e
}

// xml ist die Kopfzeile eines $metadata.
func xmlKopf() aenderung {
	return func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{"Content-Type": {"application/xml"}} }
}

// metadaten legt ein $metadata an; der Inhalt bestimmt die Referenz (gleicher Inhalt, gleiche Referenz).
func (b *bauer) metadaten(inhalt string, as ...aenderung) model.Exchange {
	ref := refVon([]byte(inhalt))
	b.w.blobs.inhalt[ref] = []byte(inhalt)
	all := append([]aenderung{xmlKopf(), func(e *model.Exchange) { e.ResponseBlob = ref }}, as...)
	return b.roh("/$metadata", "", all...)
}

func (b *bauer) vorgabe(e model.Exchange, v datenVorgabe) model.Exchange {
	b.w.daten.mu.Lock()
	b.w.daten.nachRef[e.ResponseBlob] = v
	b.w.daten.mu.Unlock()
	return e
}

// collection legt GET /<set>?<query> an, die der Leser als s liest.
func (b *bauer) collection(set, query string, s model.EntitySeite, as ...aenderung) model.Exchange {
	return b.vorgabe(b.roh("/"+set, query, as...), datenVorgabe{seite: s})
}

// entity legt GET /<set>(<praedikat>) an, die der Leser als e liest.
func (b *bauer) entity(set, praedikat string, e model.Entity, as ...aenderung) model.Exchange {
	return b.vorgabe(b.roh("/"+set+"("+praedikat+")", "", as...), datenVorgabe{entity: e})
}

func status(s int) aenderung { return func(e *model.Exchange) { e.ResponseStatus = s } }

func (w *welt) standard() *bauer {
	b := w.sitzung()
	b.metadaten("<edmx 1/>")
	return b
}
