package snapshot_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"slices"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// ctxMethoden ist die Methodenmenge von context.Context: die Fakes halten den ctx eines Aufrufs als Wert fest und
// vergleichen seine Identität.
type ctxMethoden interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// ereignis ist ein Aufruf eines Fakes: Operation, ctx des Aufrufs (dieselbe Identität) und jedes Argument.
type ereignis struct {
	op   string
	ctx  ctxMethoden
	args []any
}

// prot ist das gemeinsame Protokoll aller Fakes: jede Operation jedes Ports, auch die der Dienst nie rufen darf, steht
// hier mit Namen. fehler speist einen Fehler ein (op oder op#n für den n-ten Aufruf), panik eine Panik, nach ist der
// Haltepunkt nach jeder Operation; verboten zählt die Aufrufe von Operationen, die der Dienst nie rufen darf.
type prot struct {
	t        *testing.T
	mu       sync.Mutex
	ev       []ereignis
	fehler   map[string]error
	panik    map[string]any
	nach     func(op string)
	verboten atomic.Int32
}

func neuesProt(t *testing.T) *prot {
	return &prot{t: t, fehler: map[string]error{}, panik: map[string]any{}}
}

// beginn schreibt den Aufruf ins Protokoll und liefert den Fehler der Operation: zuerst den des Kontexts, dann den
// eingespeisten.
func (p *prot) beginn(op string, ctx ctxMethoden, args ...any) error {
	p.mu.Lock()
	p.ev = append(p.ev, ereignis{op: op, ctx: ctx, args: args})
	n := 0
	for _, e := range p.ev {
		if e.op == op {
			n++
		}
	}
	err, pa := p.fehler[op], p.panik[op]
	if e, da := p.fehler[op+"#"+strconv.Itoa(n)]; da {
		err = e
	}
	p.mu.Unlock()
	if ctx != nil {
		if e := ctx.Err(); e != nil {
			return e
		}
	}
	if pa != nil {
		panic(pa)
	}
	return err
}

func (p *prot) haltepunkt(op string) {
	if p.nach != nil {
		p.nach(op)
	}
}

func (p *prot) setzeFehler(op string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil {
		delete(p.fehler, op)
		return
	}
	p.fehler[op] = err
}

func (p *prot) ops() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	r := make([]string, 0, len(p.ev))
	for _, e := range p.ev {
		r = append(r, e.op)
	}
	return r
}

func (p *prot) anzahl(op string) int {
	n := 0
	for _, o := range p.ops() {
		if o == op {
			n++
		}
	}
	return n
}

func (p *prot) gesamt() int { return len(p.ops()) }

func (p *prot) von(op string) []ereignis {
	p.mu.Lock()
	defer p.mu.Unlock()
	var r []ereignis
	for _, e := range p.ev {
		if e.op == op {
			r = append(r, e)
		}
	}
	return r
}

func (p *prot) leeren() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ev = nil
}

// verbot zählt den Aufruf einer Operation, die der Dienst nie rufen darf, und schreibt ihn ins Protokoll.
func (p *prot) verbot(op string, ctx ctxMethoden, args ...any) error {
	p.verboten.Add(1)
	return p.beginn(op, ctx, args...)
}

// wsFake ist der Fake von WorkspaceRepository: allein Get ist erlaubt.
type wsFake struct {
	p  *prot
	mu sync.Mutex
	ws map[string]model.Workspace
}

func (f *wsFake) Get(ctx context.Context, id string) (model.Workspace, error) {
	defer f.p.haltepunkt("Workspaces.Get")
	if err := f.p.beginn("Workspaces.Get", ctx, id); err != nil {
		return model.Workspace{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	w, da := f.ws[id]
	if !da {
		return model.Workspace{}, outbound.ErrNichtGefunden
	}
	return w, nil
}

func (f *wsFake) existiert(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, da := f.ws[id]
	return da
}

func (f *wsFake) loeschen(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.ws, id)
}

func (f *wsFake) Create(ctx context.Context, w model.Workspace) error {
	return f.p.verbot("Workspaces.Create", ctx, w)
}

func (f *wsFake) List(ctx context.Context, serviceID string) ([]model.Workspace, error) {
	return nil, f.p.verbot("Workspaces.List", ctx, serviceID)
}

func (f *wsFake) Advance(ctx context.Context, id string, erwartet int64) (int64, error) {
	return 0, f.p.verbot("Workspaces.Advance", ctx, id, erwartet)
}

func (f *wsFake) Delete(ctx context.Context, id string) error {
	return f.p.verbot("Workspaces.Delete", ctx, id)
}

// entFake ist der Fake von EntityRepository: allein List ist erlaubt; es liefert eine Kopie des State in der Folge,
// die der Test gesetzt hat.
type entFake struct {
	p       *prot
	mu      sync.Mutex
	ent     map[string][]model.EntityState
	nilLeer bool
}

func (f *entFake) List(ctx context.Context, wid string) ([]model.EntityState, error) {
	defer f.p.haltepunkt("Entities.List")
	if err := f.p.beginn("Entities.List", ctx, wid); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	es, da := f.ent[wid]
	if !da {
		return nil, outbound.ErrNichtGefunden
	}
	if len(es) == 0 && f.nilLeer {
		return nil, nil
	}
	r := make([]model.EntityState, len(es))
	for i, e := range es {
		e.KnownProperties = slices.Clone(e.KnownProperties)
		r[i] = e
	}
	return r, nil
}

func (f *entFake) setze(wid string, es []model.EntityState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ent[wid] = es
}

func (f *entFake) Put(ctx context.Context, e model.EntityState) error {
	return f.p.verbot("Entities.Put", ctx, e)
}

func (f *entFake) Get(ctx context.Context, wid, set, key string) (model.EntityState, error) {
	return model.EntityState{}, f.p.verbot("Entities.Get", ctx, wid, set, key)
}

func (f *entFake) ListSet(ctx context.Context, wid, set string) ([]model.EntityState, error) {
	return nil, f.p.verbot("Entities.ListSet", ctx, wid, set)
}

func (f *entFake) ReplaceAll(ctx context.Context, wid string, es []model.EntityState) error {
	return f.p.verbot("Entities.ReplaceAll", ctx, wid, es)
}

// snapFake ist der Fake von SnapshotRepository: Get, CreateWorkspace und ListWorkspace sind erlaubt. CreateWorkspace
// prüft Validate, den Workspace, die Eindeutigkeit der Kennung und des Namens im Workspace (bytegleich).
type snapFake struct {
	p         *prot
	ws        *wsFake
	mu        sync.Mutex
	baselines map[string]model.BaselineSnapshot
	snaps     []model.WorkspaceSnapshot
	// liste setzt die Folge von ListWorkspace; ohne Wert gilt Zeit, dann Kennung.
	liste   func(wid string, alle []model.WorkspaceSnapshot) []model.WorkspaceSnapshot
	nilLeer bool
}

func (f *snapFake) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	defer f.p.haltepunkt("Snapshots.Get")
	if err := f.p.beginn("Snapshots.Get", ctx, id); err != nil {
		return model.BaselineSnapshot{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	b, da := f.baselines[id]
	if !da {
		return model.BaselineSnapshot{}, outbound.ErrNichtGefunden
	}
	return b, nil
}

func (f *snapFake) CreateWorkspace(ctx context.Context, s model.WorkspaceSnapshot) error {
	defer f.p.haltepunkt("Snapshots.CreateWorkspace")
	if err := f.p.beginn("Snapshots.CreateWorkspace", ctx, s); err != nil {
		return err
	}
	if err := s.Validate(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.ws.existiert(s.WorkspaceID) {
		return outbound.ErrNichtGefunden
	}
	for _, x := range f.snaps {
		if x.ID == s.ID || (x.WorkspaceID == s.WorkspaceID && x.Name == s.Name) {
			return outbound.ErrBereitsVorhanden
		}
	}
	f.snaps = append(f.snaps, s)
	return nil
}

func (f *snapFake) ListWorkspace(ctx context.Context, wid string) ([]model.WorkspaceSnapshot, error) {
	defer f.p.haltepunkt("Snapshots.ListWorkspace")
	if err := f.p.beginn("Snapshots.ListWorkspace", ctx, wid); err != nil {
		return nil, err
	}
	if !f.ws.existiert(wid) {
		return nil, outbound.ErrNichtGefunden
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	var eigene []model.WorkspaceSnapshot
	for _, x := range f.snaps {
		if x.WorkspaceID == wid {
			eigene = append(eigene, x)
		}
	}
	if f.liste != nil {
		return f.liste(wid, eigene), nil
	}
	sort.Slice(eigene, func(i, j int) bool {
		if !eigene[i].CreatedAt.Equal(eigene[j].CreatedAt) {
			return eigene[i].CreatedAt.Before(eigene[j].CreatedAt)
		}
		return eigene[i].ID < eigene[j].ID
	})
	if len(eigene) == 0 && f.nilLeer {
		return nil, nil
	}
	return eigene, nil
}

func (f *snapFake) alle() []model.WorkspaceSnapshot {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.snaps)
}

func (f *snapFake) List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error) {
	return nil, f.p.verbot("Snapshots.List", ctx, serviceID)
}

func (f *snapFake) GetWorkspace(ctx context.Context, wid, id string) (model.WorkspaceSnapshot, error) {
	return model.WorkspaceSnapshot{}, f.p.verbot("Snapshots.GetWorkspace", ctx, wid, id)
}

// blobFake ist der Fake von BlobRepository: allein Begin ist erlaubt. Der Writer hält jeden Aufruf fest; Commit
// liefert die Referenz des Inhalts oder fest (fremd), wenn der Test sie setzt.
type blobFake struct {
	p      *prot
	mu     sync.Mutex
	inhalt map[string][]byte
	fremd  string
	// abortFehler ist der Fehler von Abort.
	abortFehler error
	grenzen     []int64
}

func (b *blobFake) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	defer b.p.haltepunkt("Blobs.Begin")
	if err := b.p.beginn("Blobs.Begin", ctx, grenze); err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.grenzen = append(b.grenzen, grenze)
	b.mu.Unlock()
	return &schreiber{b: b}, nil
}

func (b *blobFake) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	return nil, b.p.verbot("Blobs.Open", ctx, ref)
}

func (b *blobFake) Verify(ctx context.Context, ref string) error {
	return b.p.verbot("Blobs.Verify", ctx, ref)
}

func (b *blobFake) List(ctx context.Context) ([]string, error) {
	return nil, b.p.verbot("Blobs.List", ctx)
}

func (b *blobFake) Remove(ctx context.Context, ref string) error {
	return b.p.verbot("Blobs.Remove", ctx, ref)
}

// schreiber ist der Writer des Fakes; er zählt Write, Commit und Abort und hält die Bytes fest.
type schreiber struct {
	b   *blobFake
	buf bytes.Buffer
}

func (s *schreiber) Write(p []byte) (int, error) {
	defer s.b.p.haltepunkt("Writer.Write")
	if err := s.b.p.beginn("Writer.Write", nil, p); err != nil {
		return 0, err
	}
	return s.buf.Write(p)
}

func (s *schreiber) Commit() (string, int64, error) {
	defer s.b.p.haltepunkt("Writer.Commit")
	if err := s.b.p.beginn("Writer.Commit", nil); err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(s.buf.Bytes())
	ref := "sha256:" + hex.EncodeToString(sum[:])
	if s.b.fremd != "" {
		ref = s.b.fremd
	}
	s.b.mu.Lock()
	s.b.inhalt[ref] = bytes.Clone(s.buf.Bytes())
	s.b.mu.Unlock()
	return ref, int64(s.buf.Len()), nil
}

func (s *schreiber) Abort() error {
	_ = s.b.p.beginn("Writer.Abort", nil)
	return s.b.abortFehler
}

// zaehlUhr und zaehlIDs sind FestUhr und FestIDs, deren Aufrufe das Protokoll zählt.
type zaehlUhr struct {
	p *prot
	u *supporttest.FestUhr
}

func (u zaehlUhr) Now() time.Time {
	_ = u.p.beginn("Uhr.Now", nil)
	return u.u.Now()
}

type zaehlIDs struct {
	p *prot
	g *supporttest.FestIDs
}

func (g zaehlIDs) SessionID() string  { _ = g.p.beginn("IDs.SessionID", nil); return g.g.SessionID() }
func (g zaehlIDs) ExchangeID() string { _ = g.p.beginn("IDs.ExchangeID", nil); return g.g.ExchangeID() }
func (g zaehlIDs) KorrelationsID() string {
	_ = g.p.beginn("IDs.KorrelationsID", nil)
	return g.g.KorrelationsID()
}
func (g zaehlIDs) WorkspaceID() string {
	_ = g.p.beginn("IDs.WorkspaceID", nil)
	return g.g.WorkspaceID()
}
func (g zaehlIDs) SnapshotID() string {
	_ = g.p.beginn("IDs.SnapshotID", nil)
	return g.g.SnapshotID()
}

// Bindung der Fakes an die Ports.
var (
	_ outbound.WorkspaceRepository = (*wsFake)(nil)
	_ outbound.EntityRepository    = (*entFake)(nil)
	_ outbound.SnapshotRepository  = (*snapFake)(nil)
	_ outbound.BlobRepository      = (*blobFake)(nil)
	_ outbound.CaptureWriter       = (*schreiber)(nil)
	_ support.Clock                = zaehlUhr{}
	_ support.IDGenerator          = zaehlIDs{}
)
