package bootstrap_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// basisKonfig ist eine gültige Konfiguration im Modus modus ohne die Umgebung.
func basisKonfig(dir, modus string, dienste ...bootstrap.Dienst) bootstrap.Konfiguration {
	return bootstrap.Konfiguration{
		Modus: modus, Listen: "127.0.0.1:0", DataDir: dir, Dienste: dienste,
		ReplayStrategy: "sequential", ReplayAssignment: "explicit", MatchInvalidEncoding: "literal", MatchBodyMode: "bytes",
		MaxCaptureBody: 1 << 20, UpstreamTimeout: 5 * time.Second, UpstreamTLSVerify: true,
		ShutdownTimeout: 2 * time.Second, SQLiteBusyTimeout: 5 * time.Second,
	}
}

func dienstA(upstream string) bootstrap.Dienst {
	return bootstrap.Dienst{ID: "a", Root: "/a", UpstreamURL: upstream}
}

func oeffneAblage(ctx context.Context, t *testing.T, dir string) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(ctx, filepath.Join(dir, dateiDB), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// lauf baut, fährt hoch und lässt am Ende herunterfahren.
func lauf(ctx context.Context, t *testing.T, k bootstrap.Konfiguration, log *slog.Logger, b bootstrap.Bauteile) (*bootstrap.Handle, error) {
	t.Helper()
	h, err := bootstrap.BuildMit(ctx, k, log, b)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = h.Herunterfahren(ctx) })
	return h, h.Hochfahren(ctx)
}

// TestDiensteAnlegen fängt B2: Dienst {unbekannt, bekannt gleich, anderer Root, andere Version,
// nicht genannter Eintrag} gegen echte Ablage, dazu Neustart im selben DATA_DIR.
func TestDiensteAnlegen(t *testing.T) {
	ctx := context.Background()
	wurzel := "ODR-E2001 service_root: gespeicherter Service gleicher id hat einen anderen service_root"
	version := "ODR-E2001 odata_version: gespeicherter Service gleicher id hat eine andere odata_version"
	faelle := []struct {
		name     string
		vorab    []model.Service
		zeile    string
		sessions int
		ids      []string
	}{
		{"unbekannt", nil, "", 1, []string{"a"}},
		{"bekannt gleich", []model.Service{{ID: "a", ServiceRoot: "/a", ODataVersion: "4.0"}}, "", 1, []string{"a"}},
		{"anderer Root", []model.Service{{ID: "a", ServiceRoot: "/x", ODataVersion: "4.0"}}, wurzel, 0, []string{"a"}},
		{"andere Version", []model.Service{{ID: "a", ServiceRoot: "/a", ODataVersion: "3.0"}}, version, 0, []string{"a"}},
		{"nicht genannt", []model.Service{{ID: "zz", ServiceRoot: "/zz", ODataVersion: "4.0"}}, "", 1, []string{"a", "zz"}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			dir := t.TempDir()
			if f.vorab != nil {
				db := oeffneAblage(ctx, t, dir)
				for _, s := range f.vorab {
					if err := db.Metadata().Create(ctx, s); err != nil {
						t.Fatal(err)
					}
				}
				_ = db.Close()
			}
			var logPuffer syncPuffer
			h, err := lauf(ctx, t, basisKonfig(dir, modusRecord, dienstA("http://127.0.0.1:9/a")),
				slog.New(slog.NewTextHandler(&logPuffer, nil)), bootstrap.StandardBauteile())
			if f.zeile == "" && err != nil {
				t.Fatalf("Hochfahren: %v", err)
			}
			if f.zeile != "" {
				if err == nil || bootstrap.Meldung(err) != f.zeile {
					t.Fatalf("Hochfahren: %v, erwartet %q", err, f.zeile)
				}
				for _, s := range []string{"ereignis=dienst_abweichend", "error_code=ODR-E2001", "service_id=a", "level=ERROR"} {
					if !strings.Contains(logPuffer.String(), s) {
						t.Errorf("Log ohne %q:\n%s", s, logPuffer.String())
					}
				}
			}
			if err := h.Herunterfahren(ctx); err != nil {
				t.Fatal(err)
			}
			db := oeffneAblage(ctx, t, dir)
			dienste, err := db.Metadata().List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			var ids []string
			for _, d := range dienste {
				ids = append(ids, d.ID)
				// Ein vorhandener Service bleibt, wie er war; ein neuer trägt Root, 4.0 und leere Metadaten.
				soll := model.Service{ID: "a", ServiceRoot: "/a", ODataVersion: "4.0"}
				if len(f.vorab) > 0 && d.ID == f.vorab[0].ID {
					soll = f.vorab[0]
				}
				if d.ID != "zz" && d != soll {
					t.Errorf("gespeicherter Service %+v, erwartet %+v", d, soll)
				}
			}
			sessions, _ := db.Sessions().List(ctx, "a")
			if !slices.Equal(ids, f.ids) || len(sessions) != f.sessions {
				t.Errorf("Services %v (erwartet %v), Sessions %d (erwartet %d)", ids, f.ids, len(sessions), f.sessions)
			}
		})
	}

	t.Run("Neustart", func(t *testing.T) {
		dir := t.TempDir()
		k := basisKonfig(dir, modusRecord, dienstA("http://127.0.0.1:9/a"))
		h, err := lauf(ctx, t, k, nil, bootstrap.StandardBauteile())
		if err != nil {
			t.Fatal(err)
		}
		if err := h.Herunterfahren(ctx); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{dateiDB, verzeichnisB} {
			if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
				t.Errorf("%s fehlt: %v", name, err)
			}
		}
		h, err = lauf(ctx, t, k, nil, bootstrap.StandardBauteile())
		if err != nil {
			t.Fatal(err)
		}
		sessions, err := h.Sessions().List(ctx)
		if err != nil || len(sessions) != 2 || sessions[0].Status != model.SessionComplete || sessions[1].Status != model.SessionOpen {
			t.Errorf("Sessions %+v, %v", sessions, err)
		}
		if err := h.Herunterfahren(ctx); err != nil {
			t.Fatal(err)
		}
		dienste, _ := oeffneAblage(ctx, t, dir).Metadata().List(ctx)
		if len(dienste) != 1 {
			t.Errorf("Services %+v", dienste)
		}
	})
}

// TestHandleRandformen fängt die Zellen des Rasters zu Handle: Zugänge je Modus, Zustand vor
// Hochfahren, zweites Hochfahren, Herunterfahren ohne Hochfahren, ein Fehler von Create.
func TestHandleRandformen(t *testing.T) {
	ctx := context.Background()
	for _, modus := range modi() {
		t.Run(modus, func(t *testing.T) {
			h := neuHaken()
			hd, err := bootstrap.BuildMit(ctx, basisKonfig(t.TempDir(), modus, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
			if err != nil {
				t.Fatal(err)
			}
			if (hd.Record() != nil) != (modus == modusRecord) || (hd.Sessions() != nil) != (modus == modusRecord) || hd.Wartung() == nil {
				t.Errorf("Zugänge im Modus %s: Record %v, Sessions %v", modus, hd.Record(), hd.Sessions())
			}
			select {
			case e, offen := <-hd.Fehler():
				t.Fatalf("Fehler() vor Hochfahren: %v, %v", e, offen)
			default:
			}
			if hd.Adresse() != "" {
				t.Errorf("Adresse vor Hochfahren: %q", hd.Adresse())
			}
			if err := hd.Hochfahren(ctx); err != nil {
				t.Fatal(err)
			}
			if hd.Adresse() == "" || hd.Adresse() != h.listener[0].Addr().String() {
				t.Errorf("Adresse %q", hd.Adresse())
			}
			if err := hd.Hochfahren(ctx); !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("zweites Hochfahren: %v", err)
			}
			if err := hd.Herunterfahren(ctx); err != nil {
				t.Fatal(err)
			}
			if err := hd.Hochfahren(ctx); !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("Hochfahren nach Herunterfahren: %v", err)
			}
			select {
			case _, offen := <-hd.Fehler():
				if offen {
					t.Error("Fehler() trägt einen Wert")
				}
			case <-time.After(2 * time.Second):
				t.Error("Fehler() ist nach dem Ende von Serve nicht geschlossen")
			}
		})
	}
	t.Run("Herunterfahren ohne Hochfahren", func(t *testing.T) {
		h := neuHaken()
		hd, err := bootstrap.BuildMit(ctx, basisKonfig(t.TempDir(), modusRecord, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
		if err != nil {
			t.Fatal(err)
		}
		if err := hd.Herunterfahren(ctx); err != nil {
			t.Fatal(err)
		}
		if got := h.p.liste(); !slices.Equal(got, []string{pAblage, pAblageZu}) {
			t.Errorf("Ereignisse %v", got)
		}
		if err := hd.Hochfahren(ctx); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("Hochfahren nach Herunterfahren ohne Hochfahren: %v", err)
		}
	})
	for _, phase := range []string{pDienst, pDienstGet} {
		t.Run(phase+" scheitert", func(t *testing.T) {
			dir := t.TempDir()
			db := oeffneAblage(ctx, t, dir)
			if err := db.Metadata().Create(ctx, model.Service{ID: "a", ServiceRoot: "/a", ODataVersion: "4.0"}); err != nil {
				t.Fatal(err)
			}
			_ = db.Close()
			h := neuHaken()
			h.phase, h.wirkung = phase, scheitere
			_, err := lauf(ctx, t, basisKonfig(dir, modusRecord, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
			if !errors.Is(err, errPhase) {
				t.Errorf("Fehler %v", err)
			}
		})
	}
}

// TestEinWaechter fängt B3: Pipeline (record), Replay-Dienst (replay) und Wartung tragen denselben
// Wächter, und er umhüllt den Blob-Adapter.
func TestEinWaechter(t *testing.T) {
	ctx := context.Background()
	for _, modus := range modi() {
		t.Run(modus, func(t *testing.T) {
			h := neuHaken()
			dir := t.TempDir()
			hd, err := bootstrap.BuildMit(ctx, basisKonfig(dir, modus, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = hd.Herunterfahren(ctx) })
			verbraucher := h.pipeline.Blobs
			if modus == modusReplay {
				verbraucher = h.replay.Blobs
			}
			w, ok := verbraucher.(*maintenance.BlobWaechter)
			if !ok || w != h.wartung.Blobs {
				t.Fatalf("Verbraucher %T, Wartung %p", verbraucher, h.wartung.Blobs)
			}
			cw, err := w.Begin(ctx, 16)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = cw.Write([]byte("abc"))
			ref, _, err := cw.Commit()
			if err != nil {
				t.Fatal(err)
			}
			gefunden := false
			_ = filepath.WalkDir(filepath.Join(dir, verzeichnisB), func(_ string, d fs.DirEntry, _ error) error {
				gefunden = gefunden || (d != nil && !d.IsDir() && d.Name() == strings.TrimPrefix(ref, "sha256:"))
				return nil
			})
			if liste, _ := w.List(ctx); !slices.Contains(liste, ref) || !gefunden {
				t.Errorf("Blob %s nicht im Adapter: %v, Datei %v", ref, liste, gefunden)
			}
		})
	}
}

func selbstsigniertesPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	vorlage := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "ca.test"}, IsCA: true, BasicConstraintsValid: true,
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, vorlage, vorlage, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func mustPfad(t *testing.T, text string) model.JSONPfad {
	t.Helper()
	p, err := model.ParseJSONPfad(text)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// TestWerteErreichenVerbraucher fängt B7: Schleife über die Felder von Konfiguration (jedes trägt einen
// eigenen Wert, die Menge hält reflect gegen die Tabelle) und ihre Verbraucher.
func TestWerteErreichenVerbraucher(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	caDatei := filepath.Join(dir, "ca.pem")
	caPEM := selbstsigniertesPEM(t)
	if err := os.WriteFile(caDatei, caPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	k := bootstrap.Konfiguration{
		Modus: modusRecord, Listen: "127.0.0.1:7777", DataDir: filepath.Join(dir, "d"),
		Dienste:        []bootstrap.Dienst{{ID: "x1", Root: "/r1", UpstreamURL: "http://u1.test/o"}, {ID: "x2", Root: "/r2", UpstreamURL: "http://u2.test/o"}},
		ReplayStrategy: "last", ReplayAssignment: "ordered", ReplayRedactedBodies: true,
		MatchVaryHeaders: []string{"x-vary"}, MatchInvalidEncoding: "reject", MatchBodyMode: "json-canonical",
		RedactHeaders: []string{"x-red"}, RedactQueryParams: []string{"qred"}, RedactJSONPaths: []string{"$..geheim"},
		MaxCaptureBody: 12345, UpstreamTimeout: 7 * time.Second, UpstreamCAFile: caDatei, UpstreamTLSVerify: true,
		ShutdownTimeout: 11 * time.Second, RetentionMaxAge: 13 * time.Hour, SQLiteBusyTimeout: 17 * time.Second,
		VolatileHeaders: []string{"x-vol"}, LinkRewrite: "off",
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	lauf2 := func(modus string) *haken {
		kk := k
		kk.Modus = modus
		kk.DataDir = filepath.Join(dir, modus)
		_ = os.MkdirAll(kk.DataDir, 0o700)
		h := neuHaken()
		if _, err := lauf(ctx, t, kk, log, h.bauteile()); err != nil {
			t.Fatal(err)
		}
		return h
	}
	r, p := lauf2(modusRecord), lauf2(modusReplay)
	pfade := []model.JSONPfad{mustPfad(t, "$..geheim")}
	dr := []record.Dienst{{ID: "x1", Root: "/r1", UpstreamURL: "http://u1.test/o"}, {ID: "x2", Root: "/r2", UpstreamURL: "http://u2.test/o"}}
	dp := []replay.Dienst{{ID: "x1", Root: "/r1"}, {ID: "x2", Root: "/r2"}}
	gl := reflect.DeepEqual
	pruefungen := map[string]bool{
		"Modus":                r.recorder != nil && r.replay == nil && p.replay != nil && p.recorder == nil && p.session == nil && p.upstream == nil,
		"Listen":               r.listenAdr == k.Listen && p.listenAdr == k.Listen,
		"DataDir":              fileExists(filepath.Join(dir, modusRecord, dateiDB)) && fileExists(filepath.Join(dir, modusReplay, verzeichnisB)),
		"Dienste":              gl(r.recorder.Dienste, dr) && gl(p.replay.Dienste, dp),
		"ReplayStrategy":       p.replay.Strategie == model.StrategyLast,
		"ReplayRedactedBodies": p.replay.RedigierteBodies,
		"MatchVaryHeaders":     gl(r.recorder.Regeln.VaryHeaders, k.MatchVaryHeaders) && gl(p.replay.Regeln.VaryHeaders, k.MatchVaryHeaders),
		"MatchInvalidEncoding": r.recorder.Regeln.InvalidEncoding == model.KodierungReject && p.replay.Regeln.InvalidEncoding == model.KodierungReject,
		"MatchBodyMode":        r.recorder.Regeln.BodyMode == model.BodyJSONCanonical && p.replay.Regeln.BodyMode == model.BodyJSONCanonical,
		"RedactHeaders":        gl(r.recorder.Regeln.RedactHeaders, k.RedactHeaders) && gl(p.replay.Regeln.RedactHeaders, k.RedactHeaders),
		"RedactQueryParams":    gl(r.recorder.Regeln.RedactQueryParams, k.RedactQueryParams) && gl(p.replay.Regeln.RedactQueryParams, k.RedactQueryParams),
		"VolatileHeaders":      gl(r.recorder.Regeln.VolatileHeaders, k.VolatileHeaders) && gl(p.replay.Regeln.VolatileHeaders, k.VolatileHeaders),
		"RedactJSONPaths":      gl(r.pipeline.Pfade, pfade) && gl(p.replay.Pfade, pfade),
		"MaxCaptureBody":       r.pipeline.MaxBytes == k.MaxCaptureBody && p.replay.MaxBytes == k.MaxCaptureBody,
		"UpstreamTimeout":      r.upstream != nil && r.upstream.Frist == k.UpstreamTimeout,
		"UpstreamCAFile":       r.upstream != nil && slices.Equal(r.upstream.CAPEM, caPEM),
		"UpstreamTLSVerify":    r.upstream != nil && r.upstream.Pruefen,
		"ShutdownTimeout":      r.session != nil && r.session.Frist == k.ShutdownTimeout,
		"RetentionMaxAge":      r.wartung.MaxAge == k.RetentionMaxAge && p.wartung.MaxAge == k.RetentionMaxAge,
		"SQLiteBusyTimeout":    r.sqliteOpt.BusyTimeout == k.SQLiteBusyTimeout && p.sqliteOpt.BusyTimeout == k.SQLiteBusyTimeout,
	}
	ohneVerbraucher := []string{"ReplayAssignment", "LinkRewrite"}
	typ, wert := reflect.TypeOf(k), reflect.ValueOf(k)
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		_, geprueft := pruefungen[name]
		if geprueft == slices.Contains(ohneVerbraucher, name) {
			t.Errorf("Feld %s steht weder in der Tabelle der Verbraucher noch bei den Feldern ohne", name)
		}
		if wert.Field(i).IsZero() {
			t.Errorf("Feld %s trägt in diesem Test keinen eigenen Wert", name)
		}
		if ok, da := pruefungen[name]; da && !ok {
			t.Errorf("der Wert von %s erreicht seinen Verbraucher nicht", name)
		}
	}
	if len(pruefungen) != typ.NumField()-len(ohneVerbraucher) {
		t.Errorf("Tabelle hat %d Zeilen, Konfiguration %d Felder", len(pruefungen), typ.NumField())
	}
	// Eine Uhr, eine ID-Quelle, ein Logger in allen Verbrauchern.
	if r.wartung.Uhr != r.recorder.Uhr || r.wartung.Uhr != r.session.Uhr {
		t.Error("die Verbraucher teilen nicht eine Uhr")
	}
	if r.recorder.IDs != r.session.IDs || r.recorder.IDs != r.server.IDs {
		t.Error("die Verbraucher teilen nicht eine ID-Quelle")
	}
	for name, l := range map[string]*slog.Logger{
		"Wartung": r.wartung.Log, "Recorder": r.recorder.Log, "Session": r.session.Log, "Wartung replay": p.wartung.Log,
	} {
		if l != log {
			t.Errorf("%s trägt nicht den übergebenen Logger", name)
		}
	}
}

func istUngueltig(e error) bool { return errors.Is(e, model.ErrUngueltig) }

// TestModus fängt B8: der Modus wählt den Dienst am Listener (echte Adapter), und Build weist alles
// außer record und replay ab, ohne die Ablage offen zu lassen.
func TestModus(t *testing.T) {
	ctx := context.Background()
	var anfragen atomic.Int32
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		anfragen.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	defer up.Close()
	hole := func(h *bootstrap.Handle) (int, string) {
		resp, err := http.Get("http://" + h.Adresse() + "/a/Products")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	t.Run("record", func(t *testing.T) {
		anfragen.Store(0)
		dir := t.TempDir()
		h, err := lauf(ctx, t, basisKonfig(dir, modusRecord, dienstA(up.URL)), nil, bootstrap.StandardBauteile())
		if err != nil {
			t.Fatal(err)
		}
		if status, body := hole(h); status != 200 || body != `{"value":[]}` || anfragen.Load() != 1 {
			t.Errorf("Antwort %d %q, Upstream %d-mal", status, body, anfragen.Load())
		}
		if err := h.Herunterfahren(ctx); err != nil {
			t.Fatal(err)
		}
		db := oeffneAblage(ctx, t, dir)
		sessions, _ := db.Sessions().List(ctx, "a")
		if len(sessions) != 1 {
			t.Fatalf("Sessions %+v", sessions)
		}
		if ex, err := db.Recording().ListBySession(ctx, sessions[0].ID); err != nil || len(ex) != 1 {
			t.Errorf("Exchanges %d, %v", len(ex), err)
		}
	})
	t.Run("replay", func(t *testing.T) {
		anfragen.Store(0)
		h, err := lauf(ctx, t, basisKonfig(t.TempDir(), modusReplay, dienstA(up.URL)), nil, bootstrap.StandardBauteile())
		if err != nil {
			t.Fatal(err)
		}
		if status, body := hole(h); status != 400 || !strings.Contains(body, "ODR-E5003") || anfragen.Load() != 0 {
			t.Errorf("Antwort %d %q, Upstream %d-mal", status, body, anfragen.Load())
		}
	})

	datei := filepath.Join(t.TempDir(), "keine-ca.pem")
	faelle := []struct {
		name    string
		aendere func(k *bootstrap.Konfiguration)
		pruefe  func(error) bool
		offen   bool
	}{
		{"mock", func(k *bootstrap.Konfiguration) { k.Modus = "mock" }, istUngueltig, false},
		{"hybrid", func(k *bootstrap.Konfiguration) { k.Modus = "hybrid" }, istUngueltig, false},
		{"leer", func(k *bootstrap.Konfiguration) { k.Modus = "" }, istUngueltig, false},
		{"ohne Dienste", func(k *bootstrap.Konfiguration) { k.Dienste = nil }, istUngueltig, false},
		{"DATA_DIR leer", func(k *bootstrap.Konfiguration) { k.DataDir = "" }, func(e error) bool { return errors.Is(e, outbound.ErrPfad) }, false},
		{"Pfadmuster", func(k *bootstrap.Konfiguration) { k.RedactJSONPaths = []string{"$.x"} }, istUngueltig, false},
		{"CA-Datei", func(k *bootstrap.Konfiguration) { k.UpstreamCAFile = datei }, func(e error) bool {
			return bootstrap.Meldung(e) == "ODR-E2001 UPSTREAM_CA_FILE: Datei nicht lesbar"
		}, false},
		{"blobs ist eine Datei", func(k *bootstrap.Konfiguration) { _ = os.WriteFile(filepath.Join(k.DataDir, verzeichnisB), nil, 0o600) },
			func(e error) bool { return errors.Is(e, outbound.ErrPfad) }, true},
	}
	for _, f := range faelle {
		t.Run("Build/"+f.name, func(t *testing.T) {
			h := neuHaken()
			k := basisKonfig(t.TempDir(), modusRecord, dienstA(up.URL))
			f.aendere(&k)
			_, err := bootstrap.BuildMit(ctx, k, nil, h.bauteile())
			if err == nil || !f.pruefe(err) {
				t.Errorf("Fehler %v", err)
			}
			if got, soll := h.p.zaehle(pAblage), map[bool]int{false: 0, true: 1}[f.offen]; got != soll || h.p.zaehle(pAblageZu) != soll {
				t.Errorf("Ablage %d-mal geöffnet, %d-mal geschlossen, erwartet je %d", got, h.p.zaehle(pAblageZu), soll)
			}
		})
	}
}
