package mock_test

import (
	"bytes"
	"context"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// keinKontext ist der Kontext der Operationen ohne ctx im Protokoll.
func keinKontext() context.Context { return nil }

// kanarienwert steht in jedem Wert, der in keinem Fehlertext des Dienstes vorkommen darf (SPEC-195 Nr. 10).
const kanarienwert = "S3CR3T-7f"

// ereignis ist ein Aufruf eines Fakes: Name der Operation, der ctx des Aufrufs (dieselbe Identität, bei
// ParseURL nil) und jedes Argument, wie der Dienst es übergab.
type ereignis struct {
	op   string
	ctx  ctxMethoden
	args []any
}

// ctxMethoden ist die Methodenmenge von context.Context: die Fakes halten den ctx eines Aufrufs als Wert dieser
// Schnittstelle fest und vergleichen seine Identität.
type ctxMethoden interface {
	Deadline() (time.Time, bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

// prot ist das gemeinsame Protokoll der Fakes eines Aufbaus: jede Operation jedes Ports, auch die der Dienst nie rufen
// darf, erscheint mit Namen; fehler speist an einer Operation einen Fehler ein (für jeden Aufruf).
type prot struct {
	mu     sync.Mutex
	ev     []ereignis
	fehler map[string]error
	// beiNr speist einen Fehler allein beim n-ten Aufruf der Operation ein (die Stelle S19 ist der zweite Aufruf von
	// Workspaces.Get); panik lässt den Aufruf der Operation mit einer Panik enden.
	beiNr map[string]nummerFehler
	panik map[string]bool
}

type nummerFehler struct {
	n   int
	err error
}

func neuesProt() *prot {
	return &prot{fehler: map[string]error{}, beiNr: map[string]nummerFehler{}, panik: map[string]bool{}}
}

func (p *prot) beginn(ctx context.Context, op string, args ...any) error {
	p.mu.Lock()
	p.ev = append(p.ev, ereignis{op: op, ctx: ctx, args: args})
	err := p.fehler[op]
	if nf, da := p.beiNr[op]; da {
		k := 0
		for _, e := range p.ev {
			if e.op == op {
				k++
			}
		}
		if k == nf.n {
			err = nf.err
		}
	}
	panik := p.panik[op]
	p.mu.Unlock()
	if panik {
		panic("Panik des Fakes in " + op)
	}
	return err
}

// setzeFehlerBeiNr speist err allein beim n-ten Aufruf von op ein.
func (p *prot) setzeFehlerBeiNr(op string, n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.beiNr[op] = nummerFehler{n: n, err: err}
}

// setzePanik lässt jeden Aufruf von op mit einer Panik enden.
func (p *prot) setzePanik(op string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.panik[op] = true
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

func (p *prot) anzahl(op string) int { return len(p.von(op)) }

func (p *prot) leeren() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ev = nil
}

// workspaces ist der Fake von WorkspaceRepository: Get liest die Tabelle (oder ersatz), jede andere Operation zählt.
type workspaces struct {
	p      *prot
	mu     sync.Mutex
	m      map[string]model.Workspace
	ersatz func(id string) (model.Workspace, error)
	// nach ist der Haltepunkt: er läuft nach dem n-ten Aufruf von Get (n zählt alle bisherigen), bevor Get zurückkehrt.
	nach func(n int)
}

func (w *workspaces) setze(x model.Workspace) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.m[x.ID] = x
}

func (w *workspaces) entferne(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.m, id)
}

func (w *workspaces) Get(ctx context.Context, id string) (model.Workspace, error) {
	if err := w.p.beginn(ctx, "Workspaces.Get", id); err != nil {
		return model.Workspace{}, err
	}
	x, err := w.lies(id)
	if w.nach != nil {
		w.nach(w.p.anzahl("Workspaces.Get"))
	}
	return x, err
}

// lies liest den Workspace unter der Sperre der Tabelle.
func (w *workspaces) lies(id string) (model.Workspace, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.ersatz != nil {
		return w.ersatz(id)
	}
	x, da := w.m[id]
	if !da {
		return model.Workspace{}, outbound.ErrNichtGefunden
	}
	return x, nil
}

func (w *workspaces) Create(ctx context.Context, x model.Workspace) error {
	return w.p.beginn(ctx, "Workspaces.Create", x)
}

func (w *workspaces) List(ctx context.Context, serviceID string) ([]model.Workspace, error) {
	return nil, w.p.beginn(ctx, "Workspaces.List", serviceID)
}

// Advance erhöht die Revision von erwartet auf erwartet+1; eine andere gespeicherte Revision ist ErrRevisionVeraltet, ein
// unbekannter Workspace ErrNichtGefunden.
func (w *workspaces) Advance(ctx context.Context, id string, erwartet int64) (int64, error) {
	if err := w.p.beginn(ctx, "Workspaces.Advance", id, erwartet); err != nil {
		return 0, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	x, da := w.m[id]
	switch {
	case !da:
		return 0, outbound.ErrNichtGefunden
	case x.Revision != erwartet:
		return 0, outbound.ErrRevisionVeraltet
	}
	x.Revision++
	w.m[id] = x
	return x.Revision, nil
}

// revision liest die gespeicherte Revision des Workspace.
func (w *workspaces) revision(id string) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.m[id].Revision
}

func (w *workspaces) Delete(ctx context.Context, id string) error {
	return w.p.beginn(ctx, "Workspaces.Delete", id)
}

// metadaten ist der Fake von MetadataRepository: Get liest die Tabelle, jede andere Operation zählt.
type metadaten struct {
	p  *prot
	mu sync.Mutex
	m  map[string]model.Service
}

func (m *metadaten) setze(s model.Service) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[s.ID] = s
}

func (m *metadaten) entferne(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, id)
}

func (m *metadaten) Get(ctx context.Context, id string) (model.Service, error) {
	if err := m.p.beginn(ctx, "Metadaten.Get", id); err != nil {
		return model.Service{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, da := m.m[id]
	if !da {
		return model.Service{}, outbound.ErrNichtGefunden
	}
	return s, nil
}

func (m *metadaten) Create(ctx context.Context, s model.Service) error {
	return m.p.beginn(ctx, "Metadaten.Create", s)
}

func (m *metadaten) List(ctx context.Context) ([]model.Service, error) {
	return nil, m.p.beginn(ctx, "Metadaten.List")
}

func (m *metadaten) SetMetadata(ctx context.Context, id, blob, hash string) error {
	return m.p.beginn(ctx, "Metadaten.SetMetadata", id, blob, hash)
}

// strom ist der Leser eines Blobs: er zählt Close.
type strom struct {
	*bytes.Reader
	mu     sync.Mutex
	closes int
}

func (s *strom) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closes++
	return nil
}

func (s *strom) geschlossen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closes
}

// blobs ist der Fake von BlobRepository: Open liefert einen eigenen, festgehaltenen Strom, jede andere Operation zählt.
type blobs struct {
	p       *prot
	mu      sync.Mutex
	inhalt  map[string][]byte
	stroeme []*strom
}

func (b *blobs) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	if err := b.p.beginn(ctx, "Blobs.Open", ref); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	d, da := b.inhalt[ref]
	if !da {
		return nil, outbound.ErrBlobFehlt
	}
	s := &strom{Reader: bytes.NewReader(d)}
	b.stroeme = append(b.stroeme, s)
	return s, nil
}

func (b *blobs) setzeInhalt(ref string, inhalt []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inhalt[ref] = inhalt
}

func (b *blobs) strome() []*strom {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.stroeme)
}

func (b *blobs) Begin(ctx context.Context, grenze int64) (outbound.CaptureWriter, error) {
	return nil, b.p.beginn(ctx, "Blobs.Begin", grenze)
}

func (b *blobs) Verify(ctx context.Context, ref string) error {
	return b.p.beginn(ctx, "Blobs.Verify", ref)
}

func (b *blobs) List(ctx context.Context) ([]string, error) {
	return nil, b.p.beginn(ctx, "Blobs.List")
}

func (b *blobs) Remove(ctx context.Context, ref string) error {
	return b.p.beginn(ctx, "Blobs.Remove", ref)
}

// entities ist der Fake von EntityRepository: Get und ListSet lesen die Einträge je Set (ListSet auch über ersatz), jede
// andere Operation zählt; jede Operation hält ihre Argumente im Protokoll fest.
type entities struct {
	p      *prot
	mu     sync.Mutex
	m      map[string][]model.EntityState
	ersatz func(set string) ([]model.EntityState, error)
	// beobachter sieht jedes Ergebnis von ListSet (der Zustand, den die Antwort zeigt).
	beobachter func([]model.EntityState)
	// jeWs trennt den Zustand nach der Kennung des Workspace (W73); sonst gilt allein das Set.
	jeWs bool
}

// schl ist der Schlüssel der Tabelle zu Workspace und Set.
func (e *entities) schl(workspaceID, set string) string {
	if e.jeWs {
		return workspaceID + "\x00" + set
	}
	return set
}

// ausWs liest die Einträge eines Sets in einem Workspace.
func (e *entities) ausWs(workspaceID, set string) []model.EntityState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.m[e.schl(workspaceID, set)])
}

func (e *entities) setze(set string, eintraege ...model.EntityState) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.m == nil {
		e.m = map[string][]model.EntityState{}
	}
	e.m[set] = slices.Clone(eintraege)
}

// Put ersetzt den Eintrag mit demselben Schlüssel oder fügt ihn in der Folge der Bytes des Schlüssels ein; bei einem
// eingespeisten Fehler schreibt der Fake nichts.
func (e *entities) Put(ctx context.Context, x model.EntityState) error {
	if err := e.p.beginn(ctx, "Entities.Put", x); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.m == nil {
		e.m = map[string][]model.EntityState{}
	}
	x.KnownProperties = slices.Clone(x.KnownProperties)
	k := e.schl(x.WorkspaceID, x.EntitySet)
	liste := e.m[k]
	if i := slices.IndexFunc(liste, func(y model.EntityState) bool { return y.EntityKey == x.EntityKey }); i >= 0 {
		liste[i] = x
		return nil
	}
	e.m[k] = append(liste, x)
	slices.SortFunc(e.m[k], func(a, b model.EntityState) int { return strings.Compare(a.EntityKey, b.EntityKey) })
	return nil
}

// inhalt liefert eine tiefe Kopie des gespeicherten Zustands.
func (e *entities) inhalt() map[string][]model.EntityState {
	e.mu.Lock()
	defer e.mu.Unlock()
	r := make(map[string][]model.EntityState, len(e.m))
	for set, liste := range e.m {
		kopie := make([]model.EntityState, len(liste))
		for i, x := range liste {
			x.KnownProperties = slices.Clone(x.KnownProperties)
			kopie[i] = x
		}
		r[set] = kopie
	}
	return r
}

// stelleHer ersetzt den gespeicherten Zustand durch eine Sicherung von inhalt.
func (e *entities) stelleHer(s map[string][]model.EntityState) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.m = s
}

func (e *entities) Get(ctx context.Context, workspaceID, entitySet, entityKey string) (model.EntityState, error) {
	if err := e.p.beginn(ctx, "Entities.Get", workspaceID, entitySet, entityKey); err != nil {
		return model.EntityState{}, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, x := range e.m[e.schl(workspaceID, entitySet)] {
		if x.EntityKey == entityKey {
			return x, nil
		}
	}
	return model.EntityState{}, outbound.ErrNichtGefunden
}

func (e *entities) List(ctx context.Context, workspaceID string) ([]model.EntityState, error) {
	return nil, e.p.beginn(ctx, "Entities.List", workspaceID)
}

func (e *entities) ListSet(ctx context.Context, workspaceID, entitySet string) ([]model.EntityState, error) {
	if err := e.p.beginn(ctx, "Entities.ListSet", workspaceID, entitySet); err != nil {
		return nil, err
	}
	if e.ersatz != nil {
		return e.ersatz(entitySet)
	}
	e.mu.Lock()
	r := slices.Clone(e.m[e.schl(workspaceID, entitySet)])
	e.mu.Unlock()
	if e.beobachter != nil {
		e.beobachter(r)
	}
	return r, nil
}

func (e *entities) ReplaceAll(ctx context.Context, workspaceID string, eintraege []model.EntityState) error {
	return e.p.beginn(ctx, "Entities.ReplaceAll", workspaceID, eintraege)
}

// coverage ist der Fake von CoverageRepository: Get liest die Einträge je (Baseline, Set), ohne Eintrag ErrNichtGefunden;
// List zählt.
type coverage struct {
	p  *prot
	mu sync.Mutex
	m  map[string]model.Coverage
}

func (c *coverage) setze(baseline, set string, status model.CoverageStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = map[string]model.Coverage{}
	}
	c.m[baseline+"|"+set] = model.Coverage{EntitySet: set, Status: status, EvidenceJSON: "{}"}
}

func (c *coverage) Get(ctx context.Context, baselineID, entitySet string) (model.Coverage, error) {
	if err := c.p.beginn(ctx, "Coverage.Get", baselineID, entitySet); err != nil {
		return model.Coverage{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	x, da := c.m[baselineID+"|"+entitySet]
	if !da {
		return model.Coverage{}, outbound.ErrNichtGefunden
	}
	return x, nil
}

func (c *coverage) List(ctx context.Context, baselineID string) ([]model.Coverage, error) {
	return nil, c.p.beginn(ctx, "Coverage.List", baselineID)
}

// metacodec ist der Fake von MetadataCodec: ParseMetadata liefert das Modell (der Nullwert, wenn keines gesetzt ist) und
// hält Strom und Grenze im Protokoll fest.
type metacodec struct {
	p      *prot
	modell model.ServiceModel
}

func (m *metacodec) ParseMetadata(ctx context.Context, src io.Reader, grenze int64) (model.ServiceModel, model.ValidierungsBericht, error) {
	if err := m.p.beginn(ctx, "MetadatenCodec.ParseMetadata", src, grenze); err != nil {
		return model.ServiceModel{}, model.ValidierungsBericht{}, err
	}
	return m.modell, model.ValidierungsBericht{}, nil
}

// urlcodec ist der Fake von URLCodec: f bildet (Pfad, Query) auf das Ergebnis ab; ParseURL hat keinen ctx.
type urlcodec struct {
	p *prot
	f func(pfad, rawQuery string) (model.ODataURL, error)
}

func (u *urlcodec) ParseURL(pfad, rawQuery string) (model.ODataURL, error) {
	if err := u.p.beginn(keinKontext(), "URL.ParseURL", pfad, rawQuery); err != nil {
		return model.ODataURL{}, err
	}
	if u.f != nil {
		return u.f(pfad, rawQuery)
	}
	return model.ODataURL{Ressource: model.Ressource{Art: model.RessourceMetadaten}}, nil
}

// schreiber ist der Fake von ODataSchreiber: SchreibeAnzahl schreibt die Bytes der Probe nach dst und liefert den
// festgelegten Kopf; jede andere Operation zählt.
type schreiber struct {
	p     *prot
	kopf  model.Darstellung
	probe []byte
	kopfe map[string]model.Darstellung
	// zeigeIDs hängt die Schlüssel der Entities einer Collection an den geschriebenen Body.
	zeigeIDs bool
	// ort ist das Ergebnis von Adresse.
	ort string
}

func (s *schreiber) SchreibeAnzahl(ctx context.Context, dst io.Writer, n int64) (model.Darstellung, error) {
	if err := s.p.beginn(ctx, "Schreiber.SchreibeAnzahl", dst, n); err != nil {
		return model.Darstellung{}, err
	}
	if dst != nil {
		_, _ = dst.Write(s.probe)
	}
	if d, da := s.kopfe["Schreiber.SchreibeAnzahl"]; da {
		return d, nil
	}
	return s.kopf, nil
}

// ausgabe schreibt die Bytes der Probe und liefert die Darstellung der Operation: kopfe[op], sonst der feste Kopf.
func (s *schreiber) ausgabe(op string, dst io.Writer) model.Darstellung {
	if dst != nil {
		_, _ = dst.Write(append([]byte(op+":"), s.probe...))
	}
	if d, da := s.kopfe[op]; da {
		return d
	}
	return s.kopf
}

func (s *schreiber) SchreibeEntity(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.EntityAntwort) (model.Darstellung, error) {
	if err := s.p.beginn(ctx, "Schreiber.SchreibeEntity", dst, m, a); err != nil {
		return model.Darstellung{}, err
	}
	return s.ausgabe("Schreiber.SchreibeEntity", dst), nil
}

func (s *schreiber) SchreibeCollection(ctx context.Context, dst io.Writer, m model.ServiceModel, a model.CollectionAntwort) (model.Darstellung, error) {
	if err := s.p.beginn(ctx, "Schreiber.SchreibeCollection", dst, m, a); err != nil {
		return model.Darstellung{}, err
	}
	d := s.ausgabe("Schreiber.SchreibeCollection", dst)
	if s.zeigeIDs && dst != nil {
		_, _ = dst.Write([]byte("|" + ids(a.Entities)))
	}
	return d, nil
}

func (s *schreiber) SchreibeFehler(ctx context.Context, dst io.Writer, f model.Fehlerobjekt) (model.Darstellung, error) {
	return model.Darstellung{}, s.p.beginn(ctx, "Schreiber.SchreibeFehler", dst, f)
}

func (s *schreiber) Adresse(m model.ServiceModel, wurzel, entitySet, schluessel string) (string, error) {
	if err := s.p.beginn(keinKontext(), "Schreiber.Adresse", m, wurzel, entitySet, schluessel); err != nil {
		return "", err
	}
	return s.ort, nil
}

// codecs sind die drei Codec-Fakes eines Satzes, auf einem Protokoll.
type codecs struct {
	p       *prot
	url     *urlcodec
	sch     *schreiber
	mc      *metacodec
	eingabe outbound.ODataEingabe
}

func neueCodecs(p *prot, kennzeichen model.Kopfzeile) *codecs {
	return &codecs{
		p:   p,
		url: &urlcodec{p: p},
		sch: &schreiber{p: p, kopf: model.Darstellung{ContentType: "text/plain", Kennzeichen: kennzeichen}, probe: []byte("PROBE-" + kanarienwert), ort: "http://adresse.test/fake"},
		mc:  &metacodec{p: p},
	}
}

func (c *codecs) satz(nurLesend bool, version string) *outbound.CodecSatz {
	return &outbound.CodecSatz{Version: version, Metadaten: c.mc, URL: c.url, Schreiber: c.sch, NurLesend: nurLesend, Eingabe: c.eingabe}
}

const (
	dienstID   = "svc"
	dienstRoot = "/svc"
	wsID       = "ws1"
	metaRef    = "sha256:0000000000000000000000000000000000000000000000000000000000000001"
)

// umg ist ein Aufbau: die acht Fakes auf einem Protokoll und die Konfig, die sie verdrahtet.
type umg struct {
	t   *testing.T
	p   *prot
	ws  *workspaces
	md  *metadaten
	bl  *blobs
	en  *entities
	co  *coverage
	c   *codecs
	tx  *transaktion
	ei  *eingabeFake
	k   mock.Konfig
	ctx ctxMethoden
	// fehlerobjektErlaubt nimmt den Test von der Prüfung aus, die den Aufruf von SchreibeFehler verbietet (der Test der
	// Fakes ruft jede Operation selbst).
	fehlerobjektErlaubt bool
}

type schluesselKtx struct{}

// metaInhalt ist der Inhalt des Blobs der Metadaten im Standardaufbau.
func metaInhalt() []byte { return []byte("<edmx:Edmx>" + kanarienwert + "</edmx:Edmx>") }

func neueUmg(t *testing.T) *umg {
	t.Helper()
	p := neuesProt()
	u := &umg{
		t:   t,
		p:   p,
		ws:  &workspaces{p: p, m: map[string]model.Workspace{}},
		md:  &metadaten{p: p, m: map[string]model.Service{}},
		bl:  &blobs{p: p, inhalt: map[string][]byte{}},
		en:  &entities{p: p},
		co:  &coverage{p: p},
		c:   neueCodecs(p, model.Kopfzeile{Name: "X-Fake-Kennzeichen", Wert: "kennzeichen-1"}),
		ctx: context.WithValue(context.Background(), schluesselKtx{}, "ctx-des-aufrufs"),
	}
	u.tx = &transaktion{p: p, ws: u.ws, en: u.en, cu: &cursors{p: p}, sn: &snapshots{p: p}}
	u.ei = &eingabeFake{p: p}
	u.c.eingabe = u.ei
	// Der Dienst ruft SchreibeFehler nie: der Fake prüft den Zähler für jeden Test des Pakets (SPEC-195 Nr. 10, L45).
	t.Cleanup(func() {
		if n := p.anzahl("Schreiber.SchreibeFehler"); n != 0 && !u.fehlerobjektErlaubt {
			t.Errorf("SchreibeFehler %d Mal gerufen, der Dienst ruft es nie", n)
		}
	})
	u.ws.setze(model.Workspace{ID: wsID, ServiceID: dienstID, BaselineSnapshotID: "b1"})
	u.md.setze(model.Service{ID: dienstID, ServiceRoot: "http://up.test", ODataVersion: "4.0", MetadataBlob: metaRef})
	u.bl.inhalt[metaRef] = metaInhalt()
	u.k = mock.Konfig{
		Dienste:        []mock.Dienst{{ID: dienstID, Root: dienstRoot}},
		Metadaten:      u.md,
		Workspaces:     u.ws,
		Entities:       u.en,
		Coverage:       u.co,
		Blobs:          u.bl,
		MetadatenCodec: u.c.mc,
		URLCodec:       u.c.url,
		Schreiber:      u.c.sch,
		PublicBaseURL:  "http://recorder.test",
		MaxBytes:       1 << 20,
		Transaktion:    u.tx,
		Eingabe:        u.ei,
	}
	return u
}

func (u *umg) mocker() *mock.Mocker {
	u.t.Helper()
	m, err := mock.New(u.k)
	if err != nil {
		u.t.Fatalf("New: %v", err)
	}
	return m
}

// rufe baut den Dienst aus der Konfig und ruft ihn mit dem ctx des Aufbaus.
func (u *umg) rufe(a inbound.Anfrage) (inbound.Antwort, error) {
	u.t.Helper()
	return u.mocker().Mock(u.ctx, a)
}

// anfrage ist ein Request an den Standarddienst mit dem Header der Zuordnung.
func anfrage(methode, pfad string) inbound.Anfrage {
	return inbound.Anfrage{Methode: methode, Pfad: pfad, Header: map[string][]string{"X-Recorder-Workspace": {wsID}}, Laenge: -1}
}

// metadatenAnfrage ist der Standardrequest: GET auf $metadata des Standarddienstes.
func metadatenAnfrage() inbound.Anfrage { return anfrage("GET", dienstRoot+"/$metadata") }

// lies liest den Body einer Antwort bis zum Ende und schließt ihn.
func lies(t *testing.T, a inbound.Antwort) []byte {
	t.Helper()
	if a.Body == nil {
		t.Fatal("Body ist nil")
	}
	b, err := io.ReadAll(a.Body)
	if err != nil {
		t.Fatalf("Body lesen: %v", err)
	}
	if err := a.Body.Close(); err != nil {
		t.Fatalf("Body schließen: %v", err)
	}
	return b
}

// fehlerMitCode zerlegt err in den *inbound.Fehler; ist es keiner, bricht der Test ab.
func fehlerMitCode(t *testing.T, err error) *inbound.Fehler {
	t.Helper()
	f, ok := err.(*inbound.Fehler)
	if !ok {
		t.Fatalf("Fehler = %T %v, erwartet *inbound.Fehler", err, err)
	}
	return f
}

// erwarteFehler prüft Code, Status, Ursache (nil oder über errors.Is) und den Text ohne Kanarienwert.
func erwarteFehler(t *testing.T, kontext string, err error, code model.Code) *inbound.Fehler {
	t.Helper()
	if err == nil {
		t.Errorf("%s: kein Fehler, erwartet %s", kontext, code)
		return nil
	}
	f, ok := err.(*inbound.Fehler)
	if !ok {
		t.Errorf("%s: Fehler = %T %v, erwartet %s", kontext, err, err, code)
		return nil
	}
	if f.Code != code || f.Status != code.Stati()[0] {
		t.Errorf("%s: Code %s Status %d, erwartet %s Status %d", kontext, f.Code, f.Status, code, code.Stati()[0])
	}
	if want := code.String() + ": " + code.Message(); f.Error() != want {
		t.Errorf("%s: Error() = %q, erwartet %q", kontext, f.Error(), want)
	}
	return f
}
