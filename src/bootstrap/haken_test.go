package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// Die Haken setzen die Felder von bootstrap.Bauteile. Wartung, Sessions und Server sind Fakes, die
// Ablage, die Ablage der Dienste, Pipeline, Upstream-Client, Recorder und Replay-Dienst die echten,
// deren Konfig der Haken ablegt.

// Verträge, die der Test festhält (SPEC-139 Nr. 2, 5): Namen der Ablage und die Modi.
const (
	dateiDB      = "odata-recorder.db"
	verzeichnisB = "blobs"
	modusRecord  = "record"
	modusReplay  = "replay"
)

func modi() []string { return []string{modusRecord, modusReplay} }

// protokoll sammelt die Ereignisse der Fakes in der Reihenfolge ihres Eintretens.
type protokoll struct {
	mu sync.Mutex
	ev []string
}

func (p *protokoll) add(format string, a ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ev = append(p.ev, fmt.Sprintf(format, a...))
}

func (p *protokoll) liste() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.ev)
}

func (p *protokoll) zaehle(name string) int {
	n := 0
	for _, e := range p.liste() {
		if e == name {
			n++
		}
	}
	return n
}

// Die Phasen, in denen ein Fake wirkt; die Namen sind zugleich die Ereignisse im Protokoll.
const (
	pAblage     = "ablage"
	pWartung    = "wartung"
	pDienst     = "dienst"
	pDienstGet  = "dienst-get"
	pSession    = "session"
	pListen     = "listen"
	pServe      = "serve"
	pCloseAll   = "closeall"
	pSchliessen = "schliessen"
	pAblageZu   = "ablage-zu"
)

// wirkung ist, was der Fake der Phase tut.
type wirkung int

const (
	keine wirkung = iota
	// blockiere bricht ctx ab und wartet auf ctx.Done des Aufrufs: liefert dessen Fehler.
	blockiere
	// scheitere liefert errPhase.
	scheitere
	// brecheAb bricht ctx ab und gelingt.
	brecheAb
	// scheitereNachAbbruch bricht ctx ab und liefert errPhase.
	scheitereNachAbbruch
	// beendeOhneFehler lässt Serve sofort mit nil zurückkehren, ohne dass der Root angehalten hat.
	beendeOhneFehler
	// wartetAufAbbruch lässt Serve auf den Abbruch von ctxServe warten und dann mit nil zurückkehren.
	wartetAufAbbruch
)

// errPhase ist der Fehler eines Fakes; sein Text steht nie in einer Meldung.
var errPhase = errors.New("Fehler-der-Phase-GEHEIMER-WERT")

type haken struct {
	p         *protokoll
	phase     string
	auch      string
	wirkung   wirkung
	abbrechen context.CancelFunc
	// ctxServe ist der Kontext, auf dessen Abbruch Serve bei wartetAufAbbruch wartet.
	ctxServe context.Context

	// sessionFehler liefert OpenInitialSession, wenn keine Phase wirkt.
	sessionFehler error
	// fehlerIn liefert den Fehler je Phase, unabhängig von phase und wirkung; gesetzt vor dem Start, danach nur gelesen.
	fehlerIn map[string]error
	// bericht liefert CloseAll der Sessions.
	bericht inbound.CloseBericht
	// vorListen läuft im Haken von Listen, bevor er bindet.
	vorListen func()

	mu         sync.Mutex
	sqliteOpt  sqlite.Options
	wartung    maintenance.Konfig
	waechter   support.Clock
	telemetrie telemetry.Config
	pipeline   capture.Konfig
	upstream   *httpupstream.Konfig
	recorder   *record.Konfig
	session    *session.Konfig
	replay     *replay.Konfig
	server     httpodata.Konfig
	listener   []net.Listener
	listenAdr  string
	listenNetz string
	ctxs       map[string]context.Context
}

// merke hält den ctx fest, den eine Aufrufstelle erhielt.
func (h *haken) merke(ctx context.Context, stelle string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ctxs == nil {
		h.ctxs = map[string]context.Context{}
	}
	h.ctxs[stelle] = ctx
}

// ctxVon liefert den ctx des letzten Aufrufs der Stelle, nil ohne Aufruf.
func (h *haken) ctxVon(stelle string) context.Context {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.ctxs[stelle]
}

func neuHaken() *haken { return &haken{p: &protokoll{}} }

func (h *haken) wirke(ctx context.Context, phase string) error {
	if e := h.fehlerIn[phase]; e != nil {
		return e
	}
	if h.phase != phase && h.auch != phase {
		return nil
	}
	switch h.wirkung {
	case blockiere:
		h.abbrechen()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return errors.New("ctx des Schritts nicht abgebrochen")
		}
	case scheitere:
		return errPhase
	case brecheAb:
		h.abbrechen()
	case scheitereNachAbbruch:
		h.abbrechen()
		return errPhase
	case keine:
	}
	return nil
}

func (h *haken) bauteile() bootstrap.Bauteile {
	b := bootstrap.StandardBauteile()
	echt := bootstrap.StandardBauteile()
	b.Oeffne = func(ctx context.Context, pfad string, opt sqlite.Options) (*sqlite.DB, error) {
		h.mu.Lock()
		h.sqliteOpt = opt
		h.mu.Unlock()
		h.merke(ctx, pAblage)
		h.p.add(pAblage)
		// brecheAb wirkt nach dem Öffnen: ein abgebrochener Kontext öffnet nichts.
		if h.wirkung != brecheAb {
			if err := h.wirke(ctx, pAblage); err != nil {
				return nil, err
			}
		}
		db, err := echt.Oeffne(ctx, pfad, opt)
		if err == nil && h.wirkung == brecheAb {
			_ = h.wirke(ctx, pAblage)
		}
		return db, err
	}
	b.Schliesser = func(c io.Closer) io.Closer { return &protSchliesser{h: h, c: c} }
	b.Dienste = func(r outbound.MetadataRepository) outbound.MetadataRepository {
		return &protDienste{MetadataRepository: r, h: h}
	}
	b.NeuerWaechter = func(blobs outbound.BlobRepository, uhr support.Clock, karenz time.Duration) (*maintenance.BlobWaechter, error) {
		h.mu.Lock()
		h.waechter = uhr
		h.mu.Unlock()
		return echt.NeuerWaechter(blobs, uhr, karenz)
	}
	b.NeueTelemetrie = func(k telemetry.Config) (*telemetry.Adapter, error) {
		h.mu.Lock()
		h.telemetrie = k
		h.mu.Unlock()
		return echt.NeueTelemetrie(k)
	}
	b.NeueWartung = func(k maintenance.Konfig) (inbound.MaintenanceUseCase, error) {
		h.mu.Lock()
		h.wartung = k
		h.mu.Unlock()
		return &fakeWartung{h: h}, nil
	}
	b.NeuePipeline = func(k capture.Konfig) (*capture.Pipeline, error) {
		h.mu.Lock()
		h.pipeline = k
		h.mu.Unlock()
		return echt.NeuePipeline(k)
	}
	b.NeuerUpstream = func(k httpupstream.Konfig) (outbound.UpstreamClient, error) {
		h.mu.Lock()
		h.upstream = &k
		h.mu.Unlock()
		return echt.NeuerUpstream(k)
	}
	b.NeuerRecorder = func(k record.Konfig) (inbound.RecordUseCase, error) {
		h.mu.Lock()
		h.recorder = &k
		h.mu.Unlock()
		return echt.NeuerRecorder(k)
	}
	b.NeueSession = func(k session.Konfig) (inbound.SessionUseCase, error) {
		h.mu.Lock()
		h.session = &k
		h.mu.Unlock()
		return &fakeSessions{h: h, bericht: h.bericht}, nil
	}
	b.NeuerReplay = func(k replay.Konfig) (inbound.ReplayUseCase, error) {
		h.mu.Lock()
		h.replay = &k
		h.mu.Unlock()
		return echt.NeuerReplay(k)
	}
	b.Listen = func(ctx context.Context, network, addr string) (net.Listener, error) {
		h.merke(ctx, pListen)
		h.p.add(pListen)
		h.mu.Lock()
		h.listenAdr, h.listenNetz = addr, network
		h.mu.Unlock()
		if h.vorListen != nil {
			h.vorListen()
		}
		if err := h.wirke(ctx, pListen); err != nil {
			return nil, err
		}
		// Port 0 lässt die Konfiguration nicht zu; der Haken bindet einen freien Port.
		l, err := echt.Listen(ctx, network, "127.0.0.1:0")
		if err == nil {
			h.mu.Lock()
			h.listener = append(h.listener, l)
			h.mu.Unlock()
		}
		return l, err
	}
	b.NeuerServer = func(k httpodata.Konfig) (bootstrap.Server, error) {
		h.mu.Lock()
		h.server = k
		h.mu.Unlock()
		return &fakeServer{h: h, l: k.Listener, stop: make(chan struct{})}, nil
	}
	return b
}

// offeneListener zählt die Listener, die ein Dial noch erreicht.
func (h *haken) offeneListener() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := 0
	for _, l := range h.listener {
		if c, err := net.DialTimeout("tcp", l.Addr().String(), time.Second); err == nil {
			_ = c.Close()
			n++
		}
	}
	return n
}

type protSchliesser struct {
	h *haken
	c io.Closer
}

func (s *protSchliesser) Close() error {
	s.h.p.add(pAblageZu)
	err := s.c.Close()
	if werr := s.h.wirke(context.Background(), pAblageZu); werr != nil {
		return werr
	}
	return err
}

type protDienste struct {
	outbound.MetadataRepository
	h *haken
}

func (d *protDienste) Create(ctx context.Context, s model.Service) error {
	d.h.merke(ctx, pDienst)
	d.h.p.add("%s:%s", pDienst, s.ID)
	// brecheAb wirkt nach dem Anlegen: ein abgebrochener Kontext legt nichts an.
	if d.h.wirkung != brecheAb {
		if err := d.h.wirke(ctx, pDienst); err != nil {
			return err
		}
	}
	err := d.MetadataRepository.Create(ctx, s)
	if err == nil && d.h.wirkung == brecheAb {
		_ = d.h.wirke(ctx, pDienst)
	}
	return err
}

func (d *protDienste) Get(ctx context.Context, id string) (model.Service, error) {
	d.h.merke(ctx, pDienstGet)
	if d.h.phase == pDienstGet {
		return model.Service{}, errPhase
	}
	return d.MetadataRepository.Get(ctx, id)
}

type fakeWartung struct {
	inbound.MaintenanceUseCase
	h *haken
}

func (w *fakeWartung) Startup(ctx context.Context) (inbound.Wartungsbericht, error) {
	w.h.merke(ctx, pWartung)
	w.h.p.add(pWartung)
	return inbound.Wartungsbericht{}, w.h.wirke(ctx, pWartung)
}

// fakeSessions zählt jede Operation; die nicht erwarteten rufen die nil-Schnittstelle und scheitern laut.
type fakeSessions struct {
	inbound.SessionUseCase
	h       *haken
	p       *protokoll
	bericht inbound.CloseBericht
	fehler  error
	// beiCloseAll läuft in CloseAll, vor der Rückgabe.
	beiCloseAll func()

	mu sync.Mutex
	// ctxFehler ist ctx.Err() beim Aufruf von CloseAll.
	ctxFehler error
}

func (s *fakeSessions) protokoll() *protokoll {
	if s.p != nil {
		return s.p
	}
	return s.h.p
}

func (s *fakeSessions) OpenInitialSession(ctx context.Context, id string) (model.Session, error) {
	s.h.merke(ctx, pSession)
	s.protokoll().add("%s:%s", pSession, id)
	if err := s.h.wirke(ctx, pSession); err != nil {
		return model.Session{}, err
	}
	return model.Session{}, s.h.sessionFehler
}

func (s *fakeSessions) CloseAll(ctx context.Context) (inbound.CloseBericht, error) {
	s.protokoll().add(pCloseAll)
	s.mu.Lock()
	s.ctxFehler = ctx.Err()
	s.mu.Unlock()
	if s.beiCloseAll != nil {
		s.beiCloseAll()
	}
	if s.h != nil {
		if err := s.h.wirke(ctx, pCloseAll); err != nil {
			return s.bericht, err
		}
	}
	return s.bericht, s.fehler
}

func (s *fakeSessions) ctxBeiCloseAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctxFehler
}

// fakeServer ist der Listener-Adapter: Serve kehrt nach Anhalten oder Schliessen zurück, bei
// freigabe erst, wenn der Test sie schließt.
type fakeServer struct {
	h        *haken
	p        *protokoll
	l        net.Listener
	stop     chan struct{}
	once     sync.Once
	freigabe chan struct{}
	fehler   error

	mu      sync.Mutex
	fristen []time.Duration
}

func (s *fakeServer) protokoll() *protokoll {
	if s.p != nil {
		return s.p
	}
	return s.h.p
}

func (s *fakeServer) Serve() error {
	s.protokoll().add(pServe)
	if s.h != nil {
		if e := s.h.fehlerIn[pServe]; e != nil {
			return e
		}
		if s.h.phase == pServe {
			switch s.h.wirkung {
			case beendeOhneFehler:
				return nil
			case wartetAufAbbruch:
				<-s.h.ctxServe.Done()
				return nil
			case keine, blockiere, scheitere, brecheAb, scheitereNachAbbruch:
			}
		}
		if s.h.wirke(context.Background(), pServe) != nil {
			return errPhase
		}
	}
	<-s.stop
	if s.freigabe != nil {
		<-s.freigabe
	}
	s.protokoll().add("serve-ende")
	return nil
}

func (s *fakeServer) halt() {
	if s.l != nil {
		_ = s.l.Close()
	}
	s.once.Do(func() { close(s.stop) })
}

func (s *fakeServer) Anhalten() {
	s.protokoll().add("anhalten")
	s.halt()
}

func (s *fakeServer) Schliessen(frist time.Duration) error {
	s.protokoll().add(pSchliessen)
	s.mu.Lock()
	s.fristen = append(s.fristen, frist)
	s.mu.Unlock()
	s.halt()
	if s.h != nil {
		if err := s.h.wirke(context.Background(), pSchliessen); err != nil {
			return err
		}
	}
	return s.fehler
}

func (s *fakeServer) frist() []time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.fristen)
}

// syncPuffer ist ein Writer, den der Test liest, während Start schreibt.
type syncPuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (p *syncPuffer) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.Write(b)
}

func (p *syncPuffer) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.String()
}
