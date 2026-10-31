package replay_test

import (
	"bytes"
	"context"
	"io"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// ereignis ist ein Aufruf eines Fakes: Name der Operation, der ctx des Aufrufs (dieselbe Identität) und
// jedes Argument, wie der Dienst es übergab.
type ereignis struct {
	op   string
	ctx  ctxMethoden
	args []any
}

// ctxMethoden ist die Methodenmenge von context.Context: die Fakes halten den ctx eines Aufrufs als Wert fest
// und vergleichen seine Identität.
type ctxMethoden interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// kontext ist eine Hülle um einen ctx: sie trägt alle vier Methoden von context.Context und zählt jeden Aufruf.
type kontext struct {
	innen                      ctxMethoden
	deadline, done, err, value atomic.Int32
}

func neuerKontext(c context.Context) *kontext { return &kontext{innen: c} }

func (k *kontext) Deadline() (time.Time, bool) { k.deadline.Add(1); return k.innen.Deadline() }
func (k *kontext) Done() <-chan struct{}       { k.done.Add(1); return k.innen.Done() }
func (k *kontext) Err() error                  { k.err.Add(1); return k.innen.Err() }
func (k *kontext) Value(key any) any           { k.value.Add(1); return k.innen.Value(key) }

// prot ist das gemeinsame Protokoll aller Fakes: jede Operation jedes Ports, auch die der Dienst nie rufen
// darf, erscheint mit Namen; Operationen innerhalb einer Transaktion tragen das Präfix "tx ". fehler speist
// an einer Operation einen Fehler ein ("Name" für jeden Aufruf, "Name#n" für den n-ten), panik eine Panik,
// nach ist der Haltepunkt, der nach jeder Operation gerufen wird.
type prot struct {
	t      *testing.T
	mu     sync.Mutex
	ev     []ereignis
	fehler map[string]error
	panik  map[string]any
	nach   func(op string)
}

func neuesProt(t *testing.T) *prot {
	return &prot{t: t, fehler: map[string]error{}, panik: map[string]any{}}
}

// beginn schreibt den Aufruf ins Protokoll und liefert den Fehler der Operation: zuerst den des Kontexts,
// dann den eingespeisten.
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
	p.mu.Lock()
	f := p.nach
	p.mu.Unlock()
	if f != nil {
		f(op)
	}
}

func (p *prot) setzeNach(f func(op string)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nach = f
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

// sessions ist der Fake von SessionRepository (F1): Get liest den Zustand, jede andere Operation zählt;
// Delete löscht auch die Exchanges, damit der Eindringling die Session samt Inhalt entfernt.
type sessions struct {
	p  *prot
	mu sync.Mutex
	m  map[string]model.Session
}

func (s *sessions) setze(x model.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[x.ID] = x
}

func (s *sessions) da(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, da := s.m[id]
	return da
}

func (s *sessions) Get(ctx context.Context, id string) (model.Session, error) {
	err := s.p.beginn("Sessions.Get", ctx, id)
	defer s.p.haltepunkt("Sessions.Get")
	if err != nil {
		return model.Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	x, da := s.m[id]
	if !da {
		return model.Session{}, outbound.ErrNichtGefunden
	}
	return x, nil
}

func (s *sessions) Create(ctx context.Context, x model.Session) error {
	return s.p.beginn("Sessions.Create", ctx, x)
}

func (s *sessions) SetStatus(ctx context.Context, id string, st model.SessionStatus) error {
	return s.p.beginn("Sessions.SetStatus", ctx, id, st)
}

func (s *sessions) List(ctx context.Context, serviceID string) ([]model.Session, error) {
	return nil, s.p.beginn("Sessions.List", ctx, serviceID)
}

func (s *sessions) FindOpen(ctx context.Context) ([]model.Session, error) {
	return nil, s.p.beginn("Sessions.FindOpen", ctx)
}

func (s *sessions) Delete(ctx context.Context, id string) error {
	if err := s.p.beginn("Sessions.Delete", ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
	return nil
}

// exchanges ist der Fake von RecordingRepository (F2): ListBySession liefert die Liste in der programmierten
// Ordnung (eine Kopie), Append und Get zählen.
type exchanges struct {
	p     *prot
	mu    sync.Mutex
	liste map[string][]model.Exchange
}

func (x *exchanges) setze(session string, l []model.Exchange) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.liste[session] = slices.Clone(l)
}

func (x *exchanges) ListBySession(ctx context.Context, id string) ([]model.Exchange, error) {
	err := x.p.beginn("Exchanges.ListBySession", ctx, id)
	defer x.p.haltepunkt("Exchanges.ListBySession")
	if err != nil {
		return nil, err
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	return slices.Clone(x.liste[id]), nil
}

func (x *exchanges) Append(ctx context.Context, e model.Exchange) error {
	return x.p.beginn("Exchanges.Append", ctx, e)
}

func (x *exchanges) Get(ctx context.Context, id string) (model.Exchange, error) {
	return model.Exchange{}, x.p.beginn("Exchanges.Get", ctx, id)
}

// strom ist die Hülle um den Leser eines Blobs (F3): sie trägt jede Methode des umhüllten *bytes.Reader (die
// Einbettung) und überschreibt Read, WriteTo und Close, damit jeder Aufruf zählt.
type strom struct {
	*bytes.Reader
	p                       *prot
	mu                      sync.Mutex
	reads, writeTos, closes int
	gelesen                 int
}

func (s *strom) Read(b []byte) (int, error) {
	s.mu.Lock()
	s.reads++
	s.mu.Unlock()
	n, err := s.Reader.Read(b)
	s.mu.Lock()
	s.gelesen += n
	s.mu.Unlock()
	return n, err
}

func (s *strom) WriteTo(w io.Writer) (int64, error) {
	s.mu.Lock()
	s.writeTos++
	s.mu.Unlock()
	return s.Reader.WriteTo(w)
}

func (s *strom) Close() error {
	s.mu.Lock()
	s.closes++
	s.mu.Unlock()
	return nil
}

func (s *strom) geschlossen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closes
}

// blobs ist der Fake von BlobRepository (F3); jedes Open liefert einen eigenen, festgehaltenen Strom.
type blobs struct {
	p       *prot
	mu      sync.Mutex
	inhalt  map[string][]byte
	stroeme []*strom
}

func (b *blobs) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	err := b.p.beginn("Blobs.Open", ctx, ref)
	defer b.p.haltepunkt("Blobs.Open")
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	d, da := b.inhalt[ref]
	if !da {
		return nil, outbound.ErrBlobFehlt
	}
	s := &strom{Reader: bytes.NewReader(d), p: b.p}
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

// zustand ist der Inhalt des Stores.
type zustand struct {
	ws  map[string]model.Workspace
	cur map[string][]model.ReplayCursor
}

func leererZustand() zustand {
	return zustand{ws: map[string]model.Workspace{}, cur: map[string][]model.ReplayCursor{}}
}

func (z zustand) kopie() zustand {
	k := leererZustand()
	for id, w := range z.ws {
		k.ws[id] = w
	}
	for id, cs := range z.cur {
		l := make([]model.ReplayCursor, len(cs))
		for i, c := range cs {
			c.Consumed = slices.Clone(c.Consumed)
			l[i] = c
		}
		k.cur[id] = l
	}
	return k
}

// karte ist ein Platz in der FIFO-Schlange des Stores: los wird geschlossen, wenn sie an der Reihe ist.
type karte struct{ los chan struct{} }

// store ist der Fake von WorkspaceRepository, EntityRepository, CursorRepository, SnapshotRepository und
// WorkspaceTransactor auf demselben Zustand (F4). Die Transaktion nimmt eine Kopie des Zustands und gibt sie
// bei Fehler, Panik und Kontextabbruch zurück; wartende Transaktionen ordnet eine FIFO-Schlange nach dem
// Eintritt. Cursors.Put prüft, dass die Recording-Session besteht (ErrNichtGefunden sonst).
type store struct {
	p      *prot
	sess   *sessions
	mu     sync.Mutex
	z      zustand
	qmu    sync.Mutex
	q      []*karte
	aktiv  atomic.Bool
	commit error
}

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

func (s *store) ausf(name string, ctx ctxMethoden, args []any, wirkung func() error) error {
	op := "tx " + name
	if !s.aktiv.Load() {
		s.p.t.Errorf("%s außerhalb der Transaktion", op)
	}
	err := s.p.beginn(op, ctx, args...)
	defer s.p.haltepunkt(op)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return wirkung()
}

// WithWorkspaceTransaction führt f in einer Transaktion aus (SPEC-108 Nr. 1).
func (s *store) WithWorkspaceTransaction(ctx context.Context, f func(outbound.WorkspaceStores) error) error {
	if err := s.p.beginn("Transaktion.WithWorkspaceTransaction", ctx, f); err != nil {
		return err
	}
	return s.laufe(ctx, s.einreihen(), f)
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
	err = f(outbound.WorkspaceStores{Workspaces: wsRepo{s}, Entities: entRepo{s}, Cursors: curRepo{s}, Snapshots: snapRepo{s}})
	if err == nil && s.commit != nil {
		err = s.commit
	}
	if err == nil && ctx != nil {
		err = ctx.Err()
	}
	if err != nil {
		return err
	}
	erfolg = true
	return nil
}

// eindringling reiht eine Transaktion ein: sie steht in der Schlange, wenn der Aufruf zurückkehrt, und läuft
// nach allen vor ihr; das Ergebnis kommt auf dem Kanal.
func (s *store) eindringling(ctx context.Context, f func(outbound.WorkspaceStores) error) <-chan error {
	k := s.einreihen()
	ch := make(chan error, 1)
	go func() { ch <- s.laufe(ctx, k, f) }()
	return ch
}

type wsRepo struct{ s *store }

func (r wsRepo) Create(ctx context.Context, w model.Workspace) error {
	return r.s.ausf("Workspaces.Create", ctx, []any{w}, func() error { return nil })
}

func (r wsRepo) Get(ctx context.Context, id string) (w model.Workspace, err error) {
	err = r.s.ausf("Workspaces.Get", ctx, []any{id}, func() error {
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

func (r wsRepo) List(ctx context.Context, serviceID string) ([]model.Workspace, error) {
	return nil, r.s.ausf("Workspaces.List", ctx, []any{serviceID}, func() error { return nil })
}

func (r wsRepo) Advance(ctx context.Context, id string, erwartet int64) (int64, error) {
	return 0, r.s.ausf("Workspaces.Advance", ctx, []any{id, erwartet}, func() error { return nil })
}

func (r wsRepo) Delete(ctx context.Context, id string) error {
	return r.s.ausf("Workspaces.Delete", ctx, []any{id}, func() error {
		if _, da := r.s.z.ws[id]; !da {
			return outbound.ErrNichtGefunden
		}
		delete(r.s.z.ws, id)
		delete(r.s.z.cur, id)
		return nil
	})
}

type entRepo struct{ s *store }

func (r entRepo) Put(ctx context.Context, e model.EntityState) error {
	return r.s.ausf("Entities.Put", ctx, []any{e}, func() error { return nil })
}

func (r entRepo) Get(ctx context.Context, wid, set, key string) (model.EntityState, error) {
	return model.EntityState{}, r.s.ausf("Entities.Get", ctx, []any{wid, set, key}, func() error { return nil })
}

func (r entRepo) List(ctx context.Context, wid string) ([]model.EntityState, error) {
	return nil, r.s.ausf("Entities.List", ctx, []any{wid}, func() error { return nil })
}

func (r entRepo) ListSet(ctx context.Context, wid, set string) ([]model.EntityState, error) {
	return nil, r.s.ausf("Entities.ListSet", ctx, []any{wid, set}, func() error { return nil })
}

func (r entRepo) ReplaceAll(ctx context.Context, wid string, es []model.EntityState) error {
	return r.s.ausf("Entities.ReplaceAll", ctx, []any{wid, es}, func() error { return nil })
}

type curRepo struct{ s *store }

func (r curRepo) Put(ctx context.Context, c model.ReplayCursor) error {
	return r.s.ausf("Cursors.Put", ctx, []any{c}, func() error {
		if !r.s.sess.da(c.RecordingSessionID) {
			return outbound.ErrNichtGefunden
		}
		c.Consumed = slices.Clone(c.Consumed)
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
	err = r.s.ausf("Cursors.Get", ctx, []any{wid, rep, rec}, func() error {
		for _, x := range r.s.z.cur[wid] {
			if x.ReplaySessionID == rep && x.RecordingSessionID == rec {
				c = x
				c.Consumed = slices.Clone(x.Consumed)
				return nil
			}
		}
		return outbound.ErrNichtGefunden
	})
	return c, err
}

func (r curRepo) List(ctx context.Context, wid string) ([]model.ReplayCursor, error) {
	return nil, r.s.ausf("Cursors.List", ctx, []any{wid}, func() error { return nil })
}

func (r curRepo) ResetAll(ctx context.Context, wid string) error {
	return r.s.ausf("Cursors.ResetAll", ctx, []any{wid}, func() error {
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

type snapRepo struct{ s *store }

func (r snapRepo) Get(ctx context.Context, id string) (model.BaselineSnapshot, error) {
	return model.BaselineSnapshot{}, r.s.ausf("Snapshots.Get", ctx, []any{id}, func() error { return nil })
}

func (r snapRepo) List(ctx context.Context, serviceID string) ([]model.BaselineSnapshot, error) {
	return nil, r.s.ausf("Snapshots.List", ctx, []any{serviceID}, func() error { return nil })
}

func (r snapRepo) CreateWorkspace(ctx context.Context, w model.WorkspaceSnapshot) error {
	return r.s.ausf("Snapshots.CreateWorkspace", ctx, []any{w}, func() error { return nil })
}

func (r snapRepo) GetWorkspace(ctx context.Context, wid, id string) (model.WorkspaceSnapshot, error) {
	return model.WorkspaceSnapshot{}, r.s.ausf("Snapshots.GetWorkspace", ctx, []any{wid, id}, func() error { return nil })
}

func (r snapRepo) ListWorkspace(ctx context.Context, wid string) ([]model.WorkspaceSnapshot, error) {
	return nil, r.s.ausf("Snapshots.ListWorkspace", ctx, []any{wid}, func() error { return nil })
}

// weg entfernt die Session ohne Protokolleintrag (Zustand der Fixture).
func (s *sessions) weg(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
}
