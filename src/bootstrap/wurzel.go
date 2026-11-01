package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/adapters/driven/clock"
	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/csdl"
	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/jsoncodec"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
	"github.com/pt9912/odata-recorder/src/hexagon/services/workspace"
)

// Namen der Ablage in DATA_DIR (SPEC-139 Nr. 2) und die odata_version neu angelegter Services (SPEC-107 Nr. 14).
const (
	dateiDB      = "odata-recorder.db"
	verzeichnisB = "blobs"
	odataVersion = "4.0"
	modusRecord  = "record"
	modusReplay  = "replay"
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
	uhr      support.Clock
	ids      support.IDGenerator
	ablage   io.Closer
	b        bauteile
	dienste  outbound.MetadataRepository
	wartung  inbound.MaintenanceUseCase
	record   inbound.RecordUseCase
	sessions inbound.SessionUseCase
	// importer ist nur im Modus record gesetzt, workspaces in beiden Modi (SPEC-139 Nr. 1).
	importer   inbound.ImportUseCase
	workspaces inbound.WorkspaceUseCase
	// verarbeiten ist der Dienst des Betriebsmodus am Listener.
	verarbeiten httpodata.Verarbeiter
	fehler      chan error

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

// Build öffnet die Ablage und verdrahtet Adapter und Dienste nach SPEC-139 Nr. 2 bis 5; es öffnet
// keinen Listener und ruft weder Wartung noch Services noch Sessions.
func Build(ctx context.Context, k Konfiguration, log *slog.Logger) (*Handle, error) {
	return build(ctx, k, log, standardBauteile())
}

func ungueltig(text string) error { return errors.Join(model.ErrUngueltig, errors.New(text)) }

// vorlaeufig sind die Werte, die Build vor dem Öffnen der Ablage prüft und liest.
type vorlaeufig struct {
	pfade []model.JSONPfad
	caPEM []byte
}

func vorbereiten(k Konfiguration) (vorlaeufig, error) {
	var v vorlaeufig
	switch {
	case k.Modus != modusRecord && k.Modus != modusReplay:
		return v, ungueltig("Betriebsmodus nicht lieferbar")
	case len(k.Dienste) == 0:
		return v, ungueltig("keine Dienste")
	case k.DataDir == "":
		return v, fmt.Errorf("%w: DATA_DIR leer", outbound.ErrPfad)
	}
	for _, text := range k.RedactJSONPaths {
		p, err := model.ParseJSONPfad(text)
		if err != nil {
			return v, err
		}
		v.pfade = append(v.pfade, p)
	}
	if k.Modus == modusRecord && k.UpstreamCAFile != "" {
		pem, err := os.ReadFile(k.UpstreamCAFile)
		if err != nil {
			return v, &KonfigFehler{Zeilen: []string{fmt.Sprintf("%s UPSTREAM_CA_FILE: Datei nicht lesbar", model.CodeE2001)}}
		}
		v.caPEM = pem
	}
	return v, nil
}

func build(ctx context.Context, k Konfiguration, log *slog.Logger, b bauteile) (*Handle, error) {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	v, err := vorbereiten(k)
	if err != nil {
		return nil, err
	}
	db, err := b.Oeffne(ctx, filepath.Join(k.DataDir, dateiDB), sqlite.Options{BusyTimeout: k.SQLiteBusyTimeout})
	if err != nil {
		return nil, err
	}
	h := &Handle{k: k, log: log, uhr: clock.NewSystemuhr(), ids: clock.NewZufallsIDs(), ablage: b.Schliesser(db), b: b}
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
	h.wartung, err = b.NeueWartung(maintenance.Konfig{
		Sessions: db.Sessions(), Dienste: h.dienste, Speicher: db.Maintenance(), Blobs: waechter,
		Uhr: h.uhr, MaxAge: k.RetentionMaxAge, Log: h.log,
	})
	if err != nil {
		return err
	}
	h.workspaces, err = b.NeueWorkspaces(workspace.Konfig{
		Workspaces: db.Workspaces(), Transaktion: db, Snapshots: db.Snapshots(), Blobs: waechter, Dienste: h.dienste,
		Wartung: h.wartung, Uhr: h.uhr, IDs: h.ids, Log: h.log,
	})
	if err != nil {
		return err
	}
	if k.Modus == modusReplay {
		return h.verdrahteReplay(db, waechter, v)
	}
	return h.verdrahteRecord(db, waechter, v)
}

func (h *Handle) verdrahteRecord(db *sqlite.DB, waechter *maintenance.BlobWaechter, v vorlaeufig) error {
	k, b := h.k, h.b
	pipeline, err := b.NeuePipeline(capture.Standard(waechter, v.pfade, k.MaxCaptureBody))
	if err != nil {
		return err
	}
	upstream, err := b.NeuerUpstream(httpupstream.Konfig{Frist: k.UpstreamTimeout, Pruefen: k.UpstreamTLSVerify, CAPEM: v.caPEM})
	if err != nil {
		return err
	}
	dienste := make([]record.Dienst, len(k.Dienste))
	for i, d := range k.Dienste {
		dienste[i] = record.Dienst{ID: d.ID, Root: d.Root, UpstreamURL: d.UpstreamURL}
	}
	h.record, err = b.NeuerRecorder(record.Konfig{
		Dienste: dienste, Sessions: db.Sessions(), Exchanges: db.Recording(), Upstream: upstream,
		Pipeline: pipeline, Regeln: h.regeln(), Uhr: h.uhr, IDs: h.ids, Log: h.log,
	})
	if err != nil {
		return err
	}
	h.sessions, err = b.NeueSession(session.Konfig{
		Sessions: db.Sessions(), Dienste: h.dienste, Wartung: h.wartung, Uhr: h.uhr, IDs: h.ids, Log: h.log,
		Record: h.record, Exchanges: db.Recording(), Frist: k.ShutdownTimeout,
	})
	if err != nil {
		return err
	}
	h.importer, err = b.NeuerImport(baselineimport.Konfig{
		Dienste: importDienste(k.Dienste), Sessions: db.Sessions(), Exchanges: db.Recording(), Blobs: waechter,
		Metadaten: h.dienste, Snapshots: db.Snapshots(), Imports: db.Imports(), MetadatenCodec: csdl.New(),
		DatenCodec: jsoncodec.New(), Regeln: h.regeln(), MaxCaptureBody: k.MaxCaptureBody, Uhr: h.uhr, IDs: h.ids,
	})
	if err != nil {
		return err
	}
	h.verarbeiten = h.record.Record
	return nil
}

func importDienste(ds []Dienst) []baselineimport.Dienst {
	out := make([]baselineimport.Dienst, len(ds))
	for i, d := range ds {
		out[i] = baselineimport.Dienst{ID: d.ID, UpstreamURL: d.UpstreamURL}
	}
	return out
}

func (h *Handle) verdrahteReplay(db *sqlite.DB, waechter *maintenance.BlobWaechter, v vorlaeufig) error {
	k := h.k
	tel, err := h.b.NeueTelemetrie(telemetry.Config{Writer: io.Discard, Clock: h.uhr})
	if err != nil {
		return err
	}
	dienste := make([]replay.Dienst, len(k.Dienste))
	for i, d := range k.Dienste {
		dienste[i] = replay.Dienst{ID: d.ID, Root: d.Root}
	}
	rep, err := h.b.NeuerReplay(replay.Konfig{
		Dienste: dienste, Sessions: db.Sessions(), Exchanges: db.Recording(), Blobs: waechter, Transaktion: db,
		Regeln: h.regeln(), Pfade: v.pfade, MaxBytes: k.MaxCaptureBody, Strategie: model.ReplayStrategy(k.ReplayStrategy),
		RedigierteBodies: k.ReplayRedactedBodies, Telemetrie: tel,
	})
	if err != nil {
		return err
	}
	h.verarbeiten = rep.Replay
	return nil
}

// Wartung liefert den Use Case der Wartung.
func (h *Handle) Wartung() inbound.MaintenanceUseCase { return h.wartung }

// Record liefert den Use Case der Aufnahme; im Modus replay nil.
func (h *Handle) Record() inbound.RecordUseCase { return h.record }

// Sessions liefert den Use Case der Sessions; im Modus replay nil.
func (h *Handle) Sessions() inbound.SessionUseCase { return h.sessions }

// Import liefert den Use Case des Imports; im Modus replay nil.
func (h *Handle) Import() inbound.ImportUseCase { return h.importer }

// Workspaces liefert den Use Case der Workspaces, in beiden Modi.
func (h *Handle) Workspaces() inbound.WorkspaceUseCase { return h.workspaces }

// Fehler liefert höchstens einen Wert, den Fehler von Serve, und ist geschlossen, wenn Serve zurückgekehrt ist.
func (h *Handle) Fehler() <-chan error { return h.fehler }

// Adresse ist die Adresse des gebundenen Listeners, vor Hochfahren leer.
func (h *Handle) Adresse() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.adresse
}
