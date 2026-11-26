package bootstrap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/audit"
	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/adapters/driven/clock"
	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
	"github.com/pt9912/odata-recorder/src/hexagon/services/snapshot"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// Namen der Ablage in DATA_DIR (SPEC-139 Nr. 2) und die Modi.
const (
	dateiDB      = "odata-recorder.db"
	verzeichnisB = "blobs"
	modusRecord  = "record"
	modusReplay  = "replay"
	modusMock    = "mock"
	modusHybrid  = "hybrid"
)

// server ist, was der Root vom Listener-Adapter ruft (SPEC-136 Nr. 12 bis 15).
type server interface {
	Serve() error
	Anhalten()
	Schliessen(frist time.Duration) error
}

// bauteile ist die Naht der Tests: jedes Feld ist im Betrieb der Konstruktor selbst; export_test.go
// macht den Typ erreichbar, und ein Test setzt einen Haken, der die übergebene Konfig ablegt oder
// einen Fake liefert. Die Felder tragen große Anfangsbuchstaben, weil nur so das Testpaket sie setzt.
type bauteile struct {
	Oeffne         func(ctx context.Context, pfad string, opt sqlite.Options) (*sqlite.DB, error)
	Schliesser     func(io.Closer) io.Closer
	Dienste        func(outbound.MetadataRepository) outbound.MetadataRepository
	NeuerWaechter  func(outbound.BlobRepository, support.Clock, time.Duration) (*maintenance.BlobWaechter, error)
	NeueTelemetrie func(telemetry.Config) (*telemetry.Adapter, error)
	NeueWartung    func(maintenance.Konfig) (inbound.MaintenanceUseCase, error)
	NeuePipeline   func(capture.Konfig) (*capture.Pipeline, error)
	NeuerUpstream  func(httpupstream.Konfig) (outbound.UpstreamClient, error)
	NeuerRecorder  func(record.Konfig) (inbound.RecordUseCase, error)
	NeueSession    func(session.Konfig) (inbound.SessionUseCase, error)
	NeuerReplay    func(replay.Konfig) (inbound.ReplayUseCase, error)
	NeuerImport    func(baselineimport.Konfig) (inbound.ImportUseCase, error)
	NeueWorkspaces func(workspace.Konfig) (inbound.WorkspaceUseCase, error)
	NeuerSnapshot  func(snapshot.Konfig) (inbound.SnapshotUseCase, error)
	NeuerMock      func(mock.Konfig) (inbound.MockUseCase, error)
	NeuerHybrid    func(hybrid.Konfig) (inbound.HybridUseCase, error)
	NeueInspektion func(inspection.Konfig) (inbound.InspectionUseCase, error)
	Listen         func(ctx context.Context, network, addr string) (net.Listener, error)
	NeuerServer    func(httpodata.Konfig) (server, error)
}

// als macht aus einem Konstruktor einen, der die Schnittstelle I liefert; ein Fehler liefert
// den Nullwert von I, nie einen Zeiger in einer nil-Schnittstelle.
func als[I, K, T any](neu func(K) (T, error)) func(K) (I, error) {
	return func(k K) (I, error) {
		t, err := neu(k)
		var i I
		if err != nil {
			return i, err
		}
		return any(t).(I), nil
	}
}

func standardBauteile() bauteile {
	return bauteile{
		Oeffne:         sqlite.Open,
		Schliesser:     func(c io.Closer) io.Closer { return c },
		Dienste:        func(r outbound.MetadataRepository) outbound.MetadataRepository { return r },
		NeuerWaechter:  maintenance.NewBlobWaechter,
		NeueTelemetrie: telemetry.New,
		NeueWartung:    als[inbound.MaintenanceUseCase](maintenance.New),
		NeuePipeline:   capture.New,
		NeuerUpstream:  als[outbound.UpstreamClient](httpupstream.New),
		NeuerRecorder:  als[inbound.RecordUseCase](record.New),
		NeueSession:    als[inbound.SessionUseCase](session.New),
		NeuerReplay:    als[inbound.ReplayUseCase](replay.New),
		NeuerImport:    als[inbound.ImportUseCase](baselineimport.New),
		NeueWorkspaces: als[inbound.WorkspaceUseCase](workspace.New),
		NeuerSnapshot:  als[inbound.SnapshotUseCase](snapshot.New),
		NeuerMock:      als[inbound.MockUseCase](mock.New),
		NeuerHybrid:    als[inbound.HybridUseCase](hybrid.New),
		NeueInspektion: als[inbound.InspectionUseCase](inspection.New),
		Listen: func(ctx context.Context, network, addr string) (net.Listener, error) {
			var lc net.ListenConfig
			return lc.Listen(ctx, network, addr)
		},
		NeuerServer: func(k httpodata.Konfig) (server, error) {
			s, err := httpodata.New(k)
			if err != nil {
				return nil, err
			}
			return s, nil
		},
	}
}

// Handle trägt die Use Cases und den Lebenszyklus eines Prozesses (SPEC-139 Nr. 1). Er trägt keine
// Entscheidung zum Ablauf, die liegen in Hochfahren und Herunterfahren.
type Handle struct {
	k        Konfiguration
	log      *slog.Logger
	senke    io.Writer
	tel      *telemetry.Adapter
	uhr      support.Clock
	ids      support.IDGenerator
	audit    support.AuditLog
	ablage   io.Closer
	b        bauteile
	wahl     wahl
	dienste  outbound.MetadataRepository
	wartung  inbound.MaintenanceUseCase
	record   inbound.RecordUseCase
	sessions inbound.SessionUseCase
	// importer ist nur im Modus record gesetzt; sessions, inspektion, workspaces und snapshots in jedem Modus (SPEC-139 Nr. 1).
	importer   inbound.ImportUseCase
	workspaces inbound.WorkspaceUseCase
	snapshots  inbound.SnapshotUseCase
	inspektion inbound.InspectionUseCase
	// routen sind die Routen des Admin-Adapters in der Folge von SPEC-300 Nr. 5 c.
	routen []admin.Route
	// verarbeiten ist der Dienst des Betriebsmodus am Listener.
	verarbeiten httpodata.Verarbeiter
	fehler      chan error
	// tls ist die Konfiguration der Verbindungen, die Build lud; nil heißt Klartext (SPEC-291 Nr. 8).
	tls *tls.Config

	// mu schützt die folgenden Felder.
	mu       sync.Mutex
	zustand  zustand
	adresse  string
	server   server
	serveEnd chan struct{}
	// sessionBegonnen heißt: mindestens ein OpenInitialSession wurde begonnen.
	sessionBegonnen bool

	einmal   sync.Once
	ergebnis error
}

type zustand int

const (
	gebaut zustand = iota
	hochgefahren
	beendet
)

// Build öffnet die Ablage und verdrahtet Adapter, Dienste und die Routen des Admin-Adapters nach SPEC-139 Nr. 2 bis 5 und
// SPEC-300 Nr. 5 a bis c; es öffnet keinen Listener und ruft weder Wartung noch Services noch Sessions. Die Zeilen des Adapters des Prozesses gehen
// verloren (io.Discard); Start übergibt seinen Fehlerstrom (SPEC-139 Nr. 13).
func Build(ctx context.Context, k Konfiguration, log *slog.Logger) (*Handle, error) {
	return build(ctx, k, log, io.Discard, standardBauteile())
}

func ungueltig(text string) error { return errors.Join(model.ErrUngueltig, errors.New(text)) }

// vorlaeufig sind die Werte, die Build vor dem Öffnen der Ablage prüft und liest; tls ist nil ohne jeden Pfad des TLS des Listeners.
type vorlaeufig struct {
	pfade []model.JSONPfad
	caPEM []byte
	tls   *tls.Config
	wahl  wahl
}

// pruefeGrundwerte prüft Modus, Dienste, Datenverzeichnis und PublicBaseURL vor allem Öffnen.
func pruefeGrundwerte(k Konfiguration) error {
	switch {
	case k.Modus != modusRecord && k.Modus != modusReplay && k.Modus != modusMock && k.Modus != modusHybrid:
		return ungueltig("Betriebsmodus nicht lieferbar")
	case len(k.Dienste) == 0:
		return ungueltig("keine Dienste")
	case k.DataDir == "":
		return fmt.Errorf("%w: DATA_DIR leer", outbound.ErrPfad)
	case (k.Modus == modusMock || k.Modus == modusHybrid) && k.PublicBaseURL == "":
		return ungueltig("PublicBaseURL fehlt in den Modi mock und hybrid")
	}
	return nil
}

// vorbereiten lädt als ersten Schritt das Paar des TLS des Listeners zur Zeit der Uhr des Prozesses (SPEC-291 Nr. 8), in jedem
// Modus; danach prüft es die Grundwerte, die Pfade der Schwärzung und die Datei der CA.
func vorbereiten(k Konfiguration, uhr support.Clock) (vorlaeufig, error) {
	var v vorlaeufig
	var err error
	if v.tls, err = ladeListenTLS(k, uhr.Now()); err != nil {
		return v, err
	}
	if err := pruefeGrundwerte(k); err != nil {
		return v, err
	}
	if v.wahl, err = waehle(k); err != nil {
		return v, err
	}
	for _, text := range k.RedactJSONPaths {
		p, err := model.ParseJSONPfad(text)
		if err != nil {
			return v, err
		}
		v.pfade = append(v.pfade, p)
	}
	if (k.Modus == modusRecord || k.Modus == modusHybrid) && k.UpstreamCAFile != "" {
		pem, err := os.ReadFile(k.UpstreamCAFile)
		if err != nil {
			return v, &KonfigFehler{Zeilen: []string{fmt.Sprintf("%s UPSTREAM_CA_FILE: Datei nicht lesbar", model.CodeE2001)}}
		}
		v.caPEM = pem
	}
	return v, nil
}

func build(ctx context.Context, k Konfiguration, log *slog.Logger, senke io.Writer, b bauteile) (*Handle, error) {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	uhr := clock.NewSystemuhr()
	k.Dienste = wirksameDienste(k.Dienste)
	v, err := vorbereiten(k, uhr)
	if err != nil {
		return nil, err
	}
	db, err := b.Oeffne(ctx, filepath.Join(k.DataDir, dateiDB), sqlite.Options{BusyTimeout: k.SQLiteBusyTimeout})
	if err != nil {
		return nil, err
	}
	h := &Handle{k: k, log: log, senke: senke, uhr: uhr, ids: clock.NewZufallsIDs(), ablage: b.Schliesser(db), b: b, tls: v.tls, wahl: v.wahl}
	h.dienste = b.Dienste(db.Metadata())
	if err := h.verdrahte(db, v); err != nil {
		_ = h.ablage.Close()
		return nil, err
	}
	h.fehler = make(chan error, 1)
	return h, nil
}

// regeln bildet die Regelmenge des Match-Keys, an Record und Replay dieselbe (SPEC-139 Nr. 4).
func (h *Handle) regeln() model.SchluesselRegeln {
	k := h.k
	return model.SchluesselRegeln{
		VaryHeaders: k.MatchVaryHeaders, VolatileHeaders: k.VolatileHeaders,
		RedactHeaders: k.RedactHeaders, RedactQueryParams: k.RedactQueryParams,
		InvalidEncoding: model.KodierungsModus(k.MatchInvalidEncoding), BodyMode: model.BodyModus(k.MatchBodyMode),
	}
}

func (h *Handle) verdrahte(db *sqlite.DB, v vorlaeufig) error {
	k, b := h.k, h.b
	blobDir := filepath.Join(k.DataDir, verzeichnisB)
	if err := os.MkdirAll(blobDir, 0o700); err != nil {
		return fmt.Errorf("%w: %v", outbound.ErrPfad, err)
	}
	store, err := blobs.New(blobDir)
	if err != nil {
		return err
	}
	waechter, err := b.NeuerWaechter(store, h.uhr, maintenance.StandardKarenz)
	if err != nil {
		return err
	}
	// Der Adapter des Prozesses (SPEC-139 Nr. 13): Senke, Uhr und Stufe INFO; Mock, Replay- und Record-Dienst erhalten ihn als Telemetrie,
	// die Inspektion als Zaehler (SPEC-300 Nr. 5 e), Workspace-, Record-, Session-, Wartungs-Dienst und Audit-Adapter seinen Handler als Log.
	h.tel, err = b.NeueTelemetrie(telemetry.Config{Writer: h.senke, Clock: h.uhr, MinLevel: slog.LevelInfo})
	if err != nil {
		return err
	}
	h.audit, err = audit.New(audit.Config{Pfad: filepath.Join(k.DataDir, audit.DateiName), Uhr: h.uhr, IDs: h.ids, Log: slog.New(h.tel.Handler())})
	if err != nil {
		return err
	}
	h.wartung, err = b.NeueWartung(maintenance.Konfig{
		Sessions: db.Sessions(), Dienste: h.dienste, Speicher: db.Maintenance(), Blobs: waechter,
		Uhr: h.uhr, MaxAge: k.RetentionMaxAge, Log: slog.New(h.tel.Handler()), Audit: h.audit,
	})
	if err != nil {
		return err
	}
	r := ablageRepos{workspaces: db.Workspaces(), entities: db.Entities(), snapshots: db.Snapshots()}
	h.workspaces, err = b.NeueWorkspaces(workspace.Konfig{
		Workspaces: r.workspaces, Transaktion: db, Snapshots: r.snapshots, Blobs: waechter, Dienste: h.dienste,
		Wartung: h.wartung, Uhr: h.uhr, IDs: h.ids, Log: slog.New(h.tel.Handler()), Audit: h.audit,
	})
	if err != nil {
		return err
	}
	h.snapshots, err = b.NeuerSnapshot(snapshot.Konfig{
		Workspaces: r.workspaces, Entities: r.entities, Snapshots: r.snapshots, Blobs: waechter, Uhr: h.uhr, IDs: h.ids,
	})
	if err != nil {
		return err
	}
	if err := h.verdrahteModus(db, r, waechter, v); err != nil {
		return err
	}
	return h.verdrahteAdmin(db)
}

// verdrahteModus baut die Dienste des Betriebsmodus (SPEC-139 Nr. 5, 11, 12).
func (h *Handle) verdrahteModus(db *sqlite.DB, r ablageRepos, waechter *maintenance.BlobWaechter, v vorlaeufig) error {
	switch h.k.Modus {
	case modusReplay:
		return h.verdrahteReplay(db, waechter, v)
	case modusMock:
		return h.verdrahteMock(db, r, waechter)
	case modusHybrid:
		return h.verdrahteHybrid(db, r, waechter, v)
	}
	return h.verdrahteRecord(db, waechter, v)
}

// ablageRepos sind die Repositories der einen Ablage, die mehrere Dienste teilen (SPEC-139 Nr. 11 c).
type ablageRepos struct {
	workspaces *sqlite.WorkspaceRepository
	entities   *sqlite.EntityRepository
	snapshots  *sqlite.SnapshotRepository
}

func (h *Handle) verdrahteMock(db *sqlite.DB, r ablageRepos, waechter *maintenance.BlobWaechter) error {
	m, err := h.baueMock(db, r, waechter)
	if err != nil {
		return err
	}
	h.verarbeiten = m.Mock
	return nil
}

// verdrahteHybrid baut Mock, Replay, Upstream-Client und Policy und bindet den Hybrid-Dienst mit dem Adapter des Prozesses
// als Telemetrie (SPEC-139 Nr. 12 a, 13 g): kein Recorder, keine Pipeline, kein Import. Der
// Upstream-Client entsteht auch bei disabled.
func (h *Handle) verdrahteHybrid(db *sqlite.DB, r ablageRepos, waechter *maintenance.BlobWaechter, v vorlaeufig) error {
	k := h.k
	m, err := h.baueMock(db, r, waechter)
	if err != nil {
		return err
	}
	rep, err := h.baueReplay(db, waechter, v)
	if err != nil {
		return err
	}
	upstream, err := h.baueUpstream(v)
	if err != nil {
		return err
	}
	pol, err := k.NeuePolicy()
	if err != nil {
		return err
	}
	dienste := make([]hybrid.Dienst, len(k.Dienste))
	for i, d := range k.Dienste {
		dienste[i] = hybrid.Dienst{Root: d.Root, UpstreamURL: d.UpstreamURL}
	}
	hy, err := h.b.NeuerHybrid(hybrid.Konfig{
		Dienste: dienste, Mock: m, Replay: rep, Upstream: upstream, Policy: pol, MaxBytes: k.MaxCaptureBody, Telemetrie: h.tel,
	})
	if err != nil {
		return err
	}
	h.verarbeiten = hy.Hybrid
	return nil
}

// baueUpstream baut den Upstream-Client mit den Werten von UPSTREAM_TIMEOUT, UPSTREAM_TLS_VERIFY und UPSTREAM_CA_FILE
// (SPEC-139 Nr. 4), in den Modi record und hybrid.
func (h *Handle) baueUpstream(v vorlaeufig) (outbound.UpstreamClient, error) {
	k := h.k
	return h.b.NeuerUpstream(httpupstream.Konfig{Frist: k.UpstreamTimeout, Pruefen: k.UpstreamTLSVerify, CAPEM: v.caPEM})
}

func (h *Handle) baueMock(db *sqlite.DB, r ablageRepos, waechter *maintenance.BlobWaechter) (inbound.MockUseCase, error) {
	k := h.k
	dienste := make([]mock.Dienst, len(k.Dienste))
	for i, d := range k.Dienste {
		satz := h.wahl.Dienste[i]
		dienste[i] = mock.Dienst{ID: d.ID, Root: d.Root, Codecs: &satz, Etag: etagFuerSatz(satz, k.EtagEnabled)}
	}
	rueckfall := h.wahl.Rueckfall
	return h.b.NeuerMock(mock.Konfig{
		Dienste: dienste, Metadaten: h.dienste, Workspaces: r.workspaces, Entities: r.entities, Coverage: db.Coverage(),
		Blobs: waechter, MetadatenCodec: rueckfall.Metadaten, URLCodec: rueckfall.URL, Schreiber: rueckfall.Schreiber,
		PublicBaseURL: k.PublicBaseURL, StrictCoverage: k.StrictCoverage, MaxBytes: k.MaxCaptureBody, Transaktion: db, Telemetrie: h.tel,
	})
}

func (h *Handle) verdrahteRecord(db *sqlite.DB, waechter *maintenance.BlobWaechter, v vorlaeufig) error {
	k, b := h.k, h.b
	pipeline, err := b.NeuePipeline(capture.Standard(waechter, v.pfade, k.MaxCaptureBody))
	if err != nil {
		return err
	}
	upstream, err := h.baueUpstream(v)
	if err != nil {
		return err
	}
	dienste := make([]record.Dienst, len(k.Dienste))
	for i, d := range k.Dienste {
		dienste[i] = record.Dienst{ID: d.ID, Root: d.Root, UpstreamURL: d.UpstreamURL}
	}
	h.record, err = b.NeuerRecorder(record.Konfig{
		Dienste: dienste, Sessions: db.Sessions(), Exchanges: db.Recording(), Upstream: upstream,
		Pipeline: pipeline, Regeln: h.regeln(), Uhr: h.uhr, IDs: h.ids, Log: slog.New(h.tel.Handler()), Telemetrie: h.tel,
	})
	if err != nil {
		return err
	}
	if err := h.baueSessions(db, h.record); err != nil {
		return err
	}
	h.importer, err = b.NeuerImport(baselineimport.Konfig{
		Dienste: importDienste(k.Dienste, h.wahl.Dienste), Sessions: db.Sessions(), Exchanges: db.Recording(), Blobs: waechter,
		Metadaten: h.dienste, Snapshots: db.Snapshots(), Imports: db.Imports(), MetadatenCodec: h.wahl.Rueckfall.Metadaten,
		DatenCodec: h.wahl.Rueckfall.Daten, Regeln: h.regeln(), MaxCaptureBody: k.MaxCaptureBody, Uhr: h.uhr, IDs: h.ids, Audit: h.audit,
	})
	if err != nil {
		return err
	}
	h.verarbeiten = h.record.Record
	return nil
}

// importDienste bildet die Dienste des Imports; jeder trägt in Codecs den Satz der Version seines Dienstes (SPEC-139 Nr. 4).
func importDienste(ds []Dienst, saetze []outbound.CodecSatz) []baselineimport.Dienst {
	out := make([]baselineimport.Dienst, len(ds))
	for i, d := range ds {
		satz := saetze[i]
		out[i] = baselineimport.Dienst{ID: d.ID, UpstreamURL: d.UpstreamURL, Codecs: &satz}
	}
	return out
}

func (h *Handle) verdrahteReplay(db *sqlite.DB, waechter *maintenance.BlobWaechter, v vorlaeufig) error {
	rep, err := h.baueReplay(db, waechter, v)
	if err != nil {
		return err
	}
	h.verarbeiten = rep.Replay
	return nil
}

func (h *Handle) baueReplay(db *sqlite.DB, waechter *maintenance.BlobWaechter, v vorlaeufig) (inbound.ReplayUseCase, error) {
	k := h.k
	dienste := make([]replay.Dienst, len(k.Dienste))
	for i, d := range k.Dienste {
		dienste[i] = replay.Dienst{ID: d.ID, Root: d.Root}
	}
	return h.b.NeuerReplay(replay.Konfig{
		Dienste: dienste, Sessions: db.Sessions(), Exchanges: db.Recording(), Blobs: waechter, Transaktion: db,
		Regeln: h.regeln(), Pfade: v.pfade, MaxBytes: k.MaxCaptureBody, Strategie: model.ReplayStrategy(k.ReplayStrategy),
		RedigierteBodies: k.ReplayRedactedBodies, Telemetrie: h.tel,
	})
}

// Wartung liefert den Use Case der Wartung.
func (h *Handle) Wartung() inbound.MaintenanceUseCase { return h.wartung }

// Record liefert den Use Case der Aufnahme; außer im Modus record nil.
func (h *Handle) Record() inbound.RecordUseCase { return h.record }

// Sessions liefert den Use Case der Sessions, in jedem Modus.
func (h *Handle) Sessions() inbound.SessionUseCase { return h.sessions }

// Inspektion liefert den Use Case der Inspektion, in jedem Modus.
func (h *Handle) Inspektion() inbound.InspectionUseCase { return h.inspektion }

// Import liefert den Use Case des Imports; außer im Modus record nil.
func (h *Handle) Import() inbound.ImportUseCase { return h.importer }

// Workspaces liefert den Use Case der Workspaces, in jedem Modus.
func (h *Handle) Workspaces() inbound.WorkspaceUseCase { return h.workspaces }

// Snapshots liefert den Use Case der Snapshots, in jedem Modus.
func (h *Handle) Snapshots() inbound.SnapshotUseCase { return h.snapshots }

// Fehler liefert höchstens einen Wert, den Fehler von Serve, und ist geschlossen, wenn Serve zurückgekehrt ist.
func (h *Handle) Fehler() <-chan error { return h.fehler }

// Adresse ist die Adresse des gebundenen Listeners, vor Hochfahren leer.
func (h *Handle) Adresse() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.adresse
}
