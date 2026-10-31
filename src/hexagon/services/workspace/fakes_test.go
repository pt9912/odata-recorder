package workspace_test

import (
	"bytes"
	"context"
	"io"
	"slices"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// ereignis ist ein Aufruf eines Fakes: Name der Operation, der ctx des Aufrufs (dieselbe Identität)
// und jedes Argument, wie der Dienst es übergab.
type ereignis struct {
	op   string
	ctx  ctxMethoden
	args []any
}

// prot ist das gemeinsame Protokoll aller Fakes: jede Operation jedes Ports, auch die der Dienst nie
// rufen darf, erscheint hier mit Namen; Operationen innerhalb einer Transaktion tragen das Präfix "tx ".
// fehler speist an einer Operation einen Fehler ein, panik eine Panik, nach ist der Haltepunkt, der nach
// jeder Operation gerufen wird.
type prot struct {
	t      *testing.T
	mu     sync.Mutex
	ev     []ereignis
	fehler map[string]error
	panik  map[string]any
	nach   func(op string)
	// verboten zählt die Aufrufe von Operationen, die der Dienst nie rufen darf; ein Test, der sie absichtlich
	// ruft, setzt den Zähler zurück, sonst schlägt der Aufräumschritt von neu fehl.
	verboten atomic.Int32
}

func neuesProt(t *testing.T) *prot {
	return &prot{t: t, fehler: map[string]error{}, panik: map[string]any{}}
}

// beginn schreibt den Aufruf ins Protokoll und liefert den Fehler der Operation: zuerst den des
// Kontexts (SPEC-129 Nr. 2: jede Operation der Ports endet mit ihm), dann den eingespeisten.
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

func (p *prot) setzePanik(op string, v any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.panik[op] = v
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

// zustand ist der Inhalt des Stores.
type zustand struct {
	ws    map[string]model.Workspace
	ent   map[string][]model.EntityState
	cur   map[string][]model.ReplayCursor
	snaps map[string][]string
}

func leererZustand() zustand {
	return zustand{ws: map[string]model.Workspace{}, ent: map[string][]model.EntityState{},
		cur: map[string][]model.ReplayCursor{}, snaps: map[string][]string{}}
}

func kopiereEntities(es []model.EntityState) []model.EntityState {
	r := make([]model.EntityState, len(es))
	for i, e := range es {
		e.KnownProperties = slices.Clone(e.KnownProperties)
		r[i] = e
	}
	return r
}

func (z zustand) kopie() zustand {
	k := leererZustand()
	for id, w := range z.ws {
		k.ws[id] = w
	}
	for id, es := range z.ent {
		k.ent[id] = kopiereEntities(es)
	}
	for id, cs := range z.cur {
		l := make([]model.ReplayCursor, len(cs))
		for i, c := range cs {
			c.Consumed = slices.Clone(c.Consumed)
			l[i] = c
		}
		k.cur[id] = l
	}
	for id, s := range z.snaps {
		k.snaps[id] = slices.Clone(s)
	}
	return k
}

// karte ist ein Platz in der FIFO-Schlange des Stores: los wird geschlossen, wenn sie an der Reihe ist.
type karte struct{ los chan struct{} }

// store ist der Fake von WorkspaceRepository, EntityRepository, CursorRepository und
// WorkspaceTransactor auf demselben Zustand (F1). Die Transaktion nimmt eine Kopie des Zustands und gibt
// sie bei Fehler, Panik und Kontextabbruch zurück; wartende Transaktionen und schreibende Operationen
// außerhalb einer Transaktion ordnet eine FIFO-Schlange nach dem Eintritt.
type store struct {
	p     *prot
	mu    sync.Mutex
	z     zustand
	qmu   sync.Mutex
	q     []*karte
	aktiv atomic.Bool
}

func neuerStore(p *prot) *store { return &store{p: p, z: leererZustand()} }

func (s *store) einreihen() *karte {
	s.qmu.Lock()
	defer s.qmu.Unlock()
	k := &karte{los: make(chan struct{})}
	s.q = append(s.q, k)
	if len(s.q) == 1 {
		close(k.los)
	}
	return k
}

func (s *store) fertig() {
	s.qmu.Lock()
	defer s.qmu.Unlock()
	s.q = s.q[1:]
	if len(s.q) > 0 {
		close(s.q[0].los)
	}
}

// stand liefert eine Kopie des Zustands.
func (s *store) stand() zustand {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.z.kopie()
}

func (s *store) setze(z zustand) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.z = z.kopie()
}

func (s *store) ausf(tx bool, name string, ctx ctxMethoden, args []any, schreibt bool, wirkung func() error) error {
	op := name
	if tx {
		op = "tx " + name
		if !s.aktiv.Load() {
			s.p.t.Errorf("%s außerhalb der Transaktion", op)
		}
	}
	if err := s.p.beginn(op, ctx, args...); err != nil {
		s.p.haltepunkt(op)
		return err
	}
	var err error
	if !tx && schreibt {
		k := s.einreihen()
		<-k.los
		s.mu.Lock()
		err = wirkung()
		s.mu.Unlock()
		s.fertig()
	} else {
		s.mu.Lock()
		err = wirkung()
		s.mu.Unlock()
	}
	s.p.haltepunkt(op)
	return err
}

// WithWorkspaceTransaction führt f in einer Transaktion aus (SPEC-108 Nr. 1).
func (s *store) WithWorkspaceTransaction(ctx context.Context, f func(outbound.WorkspaceStores) error) error {
	if err := s.p.beginn("Transaktion.WithWorkspaceTransaction", ctx, f); err != nil {
		return err
	}
	k := s.einreihen()
	return s.laufe(ctx, k, f)
}

func (s *store) laufe(ctx context.Context, k *karte, f func(outbound.WorkspaceStores) error) (err error) {
	<-k.los
	defer s.fertig()
	s.mu.Lock()
	sicherung := s.z.kopie()
	s.mu.Unlock()
	erfolg := false
	s.aktiv.Store(true)
	defer func() {
		s.aktiv.Store(false)
		if !erfolg {
			s.mu.Lock()
			s.z = sicherung
			s.mu.Unlock()
		}
	}()
	err = f(outbound.WorkspaceStores{Workspaces: wsRepo{s, true}, Entities: entRepo{s, true}, Cursors: curRepo{s, true}, Snapshots: &snapshots{p: s.p, tx: true}})
	if err == nil && ctx != nil {
		err = ctx.Err()
	}
	if err != nil {
		return err
	}
	erfolg = true
	return nil
}

// eindringling reiht eine Transaktion ein: sie steht in der Schlange, wenn der Aufruf zurückkehrt, und
// läuft nach allen vor ihr; das Ergebnis kommt auf dem Kanal.
func (s *store) eindringling(ctx context.Context, f func(outbound.WorkspaceStores) error) <-chan error {
	k := s.einreihen()
	ch := make(chan error, 1)
	go func() { ch <- s.laufe(ctx, k, f) }()
	return ch
}

func (s *store) workspaces() outbound.WorkspaceRepository { return wsRepo{s, false} }

type wsRepo struct {
	s  *store
	tx bool
}

func (r wsRepo) Create(ctx context.Context, w model.Workspace) error {
	return r.s.ausf(r.tx, "Workspaces.Create", ctx, []any{w}, true, func() error {
		if err := w.Validate(); err != nil {
			return err
		}
		if _, da := r.s.z.ws[w.ID]; da {
			return outbound.ErrBereitsVorhanden
		}
		r.s.z.ws[w.ID] = w
		return nil
	})
}

func (r wsRepo) Get(ctx context.Context, id string) (w model.Workspace, err error) {
	err = r.s.ausf(r.tx, "Workspaces.Get", ctx, []any{id}, false, func() error {
		var da bool
		if w, da = r.s.z.ws[id]; !da {
			return outbound.ErrNichtGefunden
		}
		return nil
	})
	if err != nil {
		return model.Workspace{}, err
	}
	return w, nil
}

func (r wsRepo) List(ctx context.Context, serviceID string) (l []model.Workspace, err error) {
	err = r.s.ausf(r.tx, "Workspaces.List", ctx, []any{serviceID}, false, func() error {
		l = []model.Workspace{}
		for _, w := range r.s.z.ws {
			if w.ServiceID == serviceID {
				l = append(l, w)
			}
		}
		sort.Slice(l, func(i, j int) bool {
			if !l[i].CreatedAt.Equal(l[j].CreatedAt) {
				return l[i].CreatedAt.Before(l[j].CreatedAt)
			}
			return l[i].ID < l[j].ID
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r wsRepo) Advance(ctx context.Context, id string, erwartet int64) (n int64, err error) {
	err = r.s.ausf(r.tx, "Workspaces.Advance", ctx, []any{id, erwartet}, true, func() error {
		w, da := r.s.z.ws[id]
		switch {
		case !da:
			return outbound.ErrNichtGefunden
		case w.Revision != erwartet:
			return outbound.ErrRevisionVeraltet
		}
		w.Revision++
		r.s.z.ws[id] = w
		n = w.Revision
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r wsRepo) Delete(ctx context.Context, id string) error {
	return r.s.ausf(r.tx, "Workspaces.Delete", ctx, []any{id}, true, func() error {
		if _, da := r.s.z.ws[id]; !da {
			return outbound.ErrNichtGefunden
		}
		delete(r.s.z.ws, id)
		delete(r.s.z.ent, id)
		delete(r.s.z.cur, id)
		delete(r.s.z.snaps, id)
		return nil
	})
}

type entRepo struct {
	s  *store
	tx bool
}

func (r entRepo) Put(ctx context.Context, e model.EntityState) error {
	return r.s.ausf(r.tx, "Entities.Put", ctx, []any{e}, true, func() error {
		es := r.s.z.ent[e.WorkspaceID]
		for i := range es {
			if es[i].EntitySet == e.EntitySet && es[i].EntityKey == e.EntityKey {
				es[i] = e
				return nil
			}
		}
		r.s.z.ent[e.WorkspaceID] = append(es, e)
		return nil
	})
}

func (r entRepo) Get(ctx context.Context, wid, set, key string) (e model.EntityState, err error) {
	err = r.s.ausf(r.tx, "Entities.Get", ctx, []any{wid, set, key}, false, func() error {
		for _, x := range r.s.z.ent[wid] {
			if x.EntitySet == set && x.EntityKey == key {
				e = x
				return nil
			}
		}
		return outbound.ErrNichtGefunden
	})
	return e, err
}

func (r entRepo) List(ctx context.Context, wid string) (l []model.EntityState, err error) {
	err = r.s.ausf(r.tx, "Entities.List", ctx, []any{wid}, false, func() error {
		l = kopiereEntities(r.s.z.ent[wid])
		return nil
	})
	return l, err
}

func (r entRepo) ListSet(ctx context.Context, wid, set string) (l []model.EntityState, err error) {
	err = r.s.ausf(r.tx, "Entities.ListSet", ctx, []any{wid, set}, false, func() error {
		for _, x := range r.s.z.ent[wid] {
			if x.EntitySet == set {
				l = append(l, x)
			}
		}
		return nil
	})
	return l, err
}

func (r entRepo) ReplaceAll(ctx context.Context, wid string, es []model.EntityState) error {
	return r.s.ausf(r.tx, "Entities.ReplaceAll", ctx, []any{wid, es}, true, func() error {
		if _, da := r.s.z.ws[wid]; !da {
			return outbound.ErrNichtGefunden
		}
		r.s.z.ent[wid] = kopiereEntities(es)
		return nil
	})
}

type curRepo struct {
	s  *store
	tx bool
}

func (r curRepo) Put(ctx context.Context, c model.ReplayCursor) error {
	return r.s.ausf(r.tx, "Cursors.Put", ctx, []any{c}, true, func() error {
		cs := r.s.z.cur[c.WorkspaceID]
		for i := range cs {
			if cs[i].ReplaySessionID == c.ReplaySessionID && cs[i].RecordingSessionID == c.RecordingSessionID {
				cs[i] = c
				return nil
			}
		}
		r.s.z.cur[c.WorkspaceID] = append(cs, c)
		return nil
	})
}

func (r curRepo) Get(ctx context.Context, wid, rep, rec string) (c model.ReplayCursor, err error) {
	err = r.s.ausf(r.tx, "Cursors.Get", ctx, []any{wid, rep, rec}, false, func() error {
		for _, x := range r.s.z.cur[wid] {
			if x.ReplaySessionID == rep && x.RecordingSessionID == rec {
				c = x
				return nil
			}
		}
		return outbound.ErrNichtGefunden
	})
	return c, err
}

func (r curRepo) List(ctx context.Context, wid string) (l []model.ReplayCursor, err error) {
	err = r.s.ausf(r.tx, "Cursors.List", ctx, []any{wid}, false, func() error {
		l = slices.Clone(r.s.z.cur[wid])
		return nil
	})
	return l, err
}

func (r curRepo) ResetAll(ctx context.Context, wid string) error {
	return r.s.ausf(r.tx, "Cursors.ResetAll", ctx, []any{wid}, true, func() error {
		if _, da := r.s.z.ws[wid]; !da {
			return outbound.ErrNichtGefunden
		}
		for i := range r.s.z.cur[wid] {
			r.s.z.cur[wid][i].Position = 0
			r.s.z.cur[wid][i].Consumed = nil
		}
		return nil
	})
}

// snapshots ist der Fake von SnapshotRepository (F2): er hält die Baselines; außer Get zählt jede
// Operation und schlägt den Test fehl, denn der Dienst ruft nur Get.
type snapshots struct {
	p         *prot
	tx        bool
	mu        sync.Mutex
	baselines map[string]model.BaselineSnapshot
}

func (s *snapshots) name(op string) string {
	if s.tx {
		return "tx Snapshots." + op
	}
	return "Snapshots." + op
}

func (s *snapshots) verboten(op string, ctx ctxMethoden, args ...any) error {
	s.p.verboten.Add(1)
	return s.p.beginn(s.name(op), ctx, args...)
}

func (s *snapshots) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	if err := s.p.beginn(s.name("Get"), ctx, id); err != nil {
		s.p.haltepunkt(s.name("Get"))
		return model.BaselineSnapshot{}, err
	}
	s.mu.Lock()
	b, da := s.baselines[id]
	s.mu.Unlock()
	s.p.haltepunkt(s.name("Get"))
	if !da {
		return model.BaselineSnapshot{}, outbound.ErrNichtGefunden
	}
	return b, nil
}

func (s *snapshots) List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error) {
	return nil, s.verboten("List", ctx, serviceID)
}

func (s *snapshots) CreateWorkspace(ctx context.Context, w model.WorkspaceSnapshot) error {
	return s.verboten("CreateWorkspace", ctx, w)
}

func (s *snapshots) GetWorkspace(ctx context.Context, wid, id string) (model.WorkspaceSnapshot, error) {
	return model.WorkspaceSnapshot{}, s.verboten("GetWorkspace", ctx, wid, id)
}

func (s *snapshots) ListWorkspace(ctx context.Context, wid string) ([]model.WorkspaceSnapshot, error) {
	return nil, s.verboten("ListWorkspace", ctx, wid)
}

// strom ist die Hülle um den Leser eines Blobs (F3): sie trägt jede Methode des umhüllten
// *bytes.Reader (die Einbettung) und überschreibt Read, WriteTo und Close, damit jeder Aufruf zählt und
// ein Lesefehler an einem Offset möglich ist.
type strom struct {
	*bytes.Reader
	p                       *prot
	gesamt                  int
	fehlerBei               int
	lesefehler              error
	closeFehler             error
	reads, writeTos, closes int
	restBeiClose            int
	mu                      sync.Mutex
}

func (s *strom) pos() int { return s.gesamt - s.Len() }

func (s *strom) Read(b []byte) (int, error) {
	s.mu.Lock()
	s.reads++
	s.mu.Unlock()
	if err := s.p.beginn("Strom.Read", nil); err != nil {
		return 0, err
	}
	if s.fehlerBei >= 0 {
		rest := s.fehlerBei - s.pos()
		if rest <= 0 {
			return 0, s.lesefehler
		}
		if len(b) > rest {
			b = b[:rest]
		}
	}
	return s.Reader.Read(b)
}

func (s *strom) WriteTo(w io.Writer) (int64, error) {
	s.mu.Lock()
	s.writeTos++
	s.mu.Unlock()
	_ = s.p.beginn("Strom.WriteTo", nil)
	var n int64
	buf := make([]byte, 64)
	for {
		k, err := s.Read(buf)
		if k > 0 {
			m, werr := w.Write(buf[:k])
			n += int64(m)
			if werr != nil {
				return n, werr
			}
		}
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
	}
}

func (s *strom) Close() error {
	s.mu.Lock()
	s.closes++
	s.restBeiClose = s.Len()
	s.mu.Unlock()
	_ = s.p.beginn("Strom.Close", nil)
	return s.closeFehler
}

func (s *strom) geschlossen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closes
}

// blobs ist der Fake von BlobRepository (F3); jedes Open liefert einen eigenen, festgehaltenen Strom.
type blobs struct {
	p           *prot
	mu          sync.Mutex
	inhalt      map[string][]byte
	stroeme     []*strom
	fehlerBei   int
	lesefehler  error
	closeFehler error
}

func (b *blobs) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	if err := b.p.beginn("Blobs.Open", ctx, ref); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	d, da := b.inhalt[ref]
	if !da {
		return nil, outbound.ErrBlobFehlt
	}
	s := &strom{Reader: bytes.NewReader(d), p: b.p, gesamt: len(d), fehlerBei: b.fehlerBei, lesefehler: b.lesefehler, closeFehler: b.closeFehler}
	b.stroeme = append(b.stroeme, s)
	return s, nil
}

func (b *blobs) strome() []*strom {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.stroeme)
}

func (b *blobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	return nil, b.p.beginn("Blobs.Begin", ctx, grenze)
}

func (b *blobs) Verify(ctx context.Context, ref string) error {
	return b.p.beginn("Blobs.Verify", ctx, ref)
}

func (b *blobs) List(ctx context.Context) ([]string, error) {
	return nil, b.p.beginn("Blobs.List", ctx)
}

func (b *blobs) Remove(ctx context.Context, ref string) error {
	return b.p.beginn("Blobs.Remove", ctx, ref)
}

// dienste ist der Fake von MetadataRepository (F5): List liefert die Services, jede andere Operation
// zählt.
type dienste struct {
	p    *prot
	mu   sync.Mutex
	alle []model.Service
}

func (d *dienste) List(ctx context.Context) ([]model.Service, error) {
	if err := d.p.beginn("Dienste.List", ctx); err != nil {
		return nil, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.alle), nil
}

func (d *dienste) Create(ctx context.Context, s model.Service) error {
	return d.p.beginn("Dienste.Create", ctx, s)
}

func (d *dienste) Get(ctx context.Context, id string) (model.Service, error) {
	return model.Service{}, d.p.beginn("Dienste.Get", ctx, id)
}

func (d *dienste) SetMetadata(ctx context.Context, id, blob, hash string) error {
	return d.p.beginn("Dienste.SetMetadata", ctx, id, blob, hash)
}

// wartung ist der Fake von MaintenanceUseCase (F4): CleanupBlobs hält den ctx fest, liefert den programmierten
// Bericht und Fehler und zählt; Startup zählt.
type wartung struct {
	p       *prot
	mu      sync.Mutex
	bericht inbound.Loeschbericht
	fehler  error
}

func (w *wartung) setze(b inbound.Loeschbericht, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bericht, w.fehler = b, err
}

func (w *wartung) Startup(ctx context.Context) (inbound.Wartungsbericht, error) {
	return inbound.Wartungsbericht{}, w.p.beginn("Wartung.Startup", ctx)
}

func (w *wartung) CleanupBlobs(ctx context.Context) (inbound.Loeschbericht, error) {
	w.mu.Lock()
	b, err := w.bericht, w.fehler
	w.mu.Unlock()
	if e := w.p.beginn("Wartung.CleanupBlobs", ctx); e != nil {
		return b, e
	}
	return b, err
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
func (g zaehlIDs) SnapshotID() string { _ = g.p.beginn("IDs.SnapshotID", nil); return g.g.SnapshotID() }

var (
	_ support.Clock       = zaehlUhr{}
	_ support.IDGenerator = zaehlIDs{}
)

// ctxMethoden ist die Methodenmenge von context.Context: die Fakes halten den ctx eines Aufrufs als Wert fest
// und vergleichen seine Identität.
type ctxMethoden interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// kontext ist eine Hülle um einen ctx (SPEC-129 Nr. 2): sie trägt alle vier Methoden von
// context.Context und zählt jeden Aufruf.
type kontext struct {
	innen                      ctxMethoden
	deadline, done, err, value atomic.Int32
}

func neuerKontext(c context.Context) *kontext { return &kontext{innen: c} }

func (k *kontext) Deadline() (time.Time, bool) { k.deadline.Add(1); return k.innen.Deadline() }
func (k *kontext) Done() <-chan struct{}       { k.done.Add(1); return k.innen.Done() }
func (k *kontext) Err() error                  { k.err.Add(1); return k.innen.Err() }
func (k *kontext) Value(key any) any           { k.value.Add(1); return k.innen.Value(key) }
