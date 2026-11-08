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
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
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
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
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
		logs     int
		sessions int
		ids      []string
	}{
		{"unbekannt", nil, "", 0, 1, []string{"a"}},
		{"bekannt gleich", []model.Service{{ID: "a", ServiceRoot: "/a", ODataVersion: "4.0"}}, "", 0, 1, []string{"a"}},
		{"anderer Root", []model.Service{{ID: "a", ServiceRoot: "/x", ODataVersion: "4.0"}}, wurzel, 1, 0, []string{"a"}},
		{"andere Version", []model.Service{{ID: "a", ServiceRoot: "/a", ODataVersion: "3.0"}}, version, 1, 0, []string{"a"}},
		{"beide Felder", []model.Service{{ID: "a", ServiceRoot: "/x", ODataVersion: "3.0"}}, wurzel + "\n" + version, 2, 0, []string{"a"}},
		{"nicht genannt", []model.Service{{ID: "zz", ServiceRoot: "/zz", ODataVersion: "4.0"}}, "", 0, 1, []string{"a", "zz"}},
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
			}
			// Je abweichendem Feld genau ein Logeintrag mit allen Attributen, ganz und mit den festen Werten
			// (SPEC-139 Nr. 8 c); jeder Eintrag der Stufe ERROR oder mit einem Ereignis zählt.
			soll := map[string]string{
				"level": "ERROR", "msg": "start: gespeicherter Service weicht von der Konfiguration ab",
				"ereignis": "dienst_abweichend", "error_code": "ODR-E2001", "service_id": "a",
			}
			var eintraege []map[string]string
			for _, z := range logZeilen(t, logPuffer.String()) {
				if z["level"] == "ERROR" || z["ereignis"] != "" {
					eintraege = append(eintraege, z)
				}
			}
			if len(eintraege) != f.logs {
				t.Errorf("%d Logeinträge, erwartet %d:\n%s", len(eintraege), f.logs, logPuffer.String())
			}
			for _, e := range eintraege {
				if !maps.Equal(e, soll) {
					t.Errorf("Logeintrag %v, erwartet %v", e, soll)
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
		VolatileHeaders: []string{"x-vol"}, LinkRewrite: "off", StrictCoverage: true, PublicBaseURL: "http://pub.test:1/api",
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	lauf2 := func(modus string) *haken {
		kk := k
		kk.Modus = modus
		kk.DataDir = filepath.Join(dir, modus)
		if err := os.MkdirAll(kk.DataDir, 0o700); err != nil {
			t.Fatal(err)
		}
		h := neuHaken()
		if _, err := lauf(ctx, t, kk, log, h.bauteile()); err != nil {
			t.Fatal(err)
		}
		return h
	}
	r, p := lauf2(modusRecord), lauf2(modusReplay)
	mh, km := neuM3Haken(), k
	km.Modus, km.DataDir = modusMock, filepath.Join(dir, modusMock)
	if err := os.MkdirAll(km.DataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := lauf(ctx, t, km, log, mh.bauteile()); err != nil {
		t.Fatal(err)
	}
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
		"StrictCoverage":       len(mh.mock) == 1 && mh.mock[0].StrictCoverage,
		"PublicBaseURL":        len(mh.mock) == 1 && mh.mock[0].PublicBaseURL == k.PublicBaseURL,
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
	// Menge der Stellen mit einer Uhr (SPEC-139 Nr. 3): record 4, replay 3; replay.Konfig hat keine Uhr.
	uhren := map[string][]stelleUhr{
		modusRecord: {{"Wartung", r.wartung.Uhr}, {"Wächter", r.waechter}, {"Session", r.session.Uhr}, {"Recorder", r.recorder.Uhr}},
		modusReplay: {{"Wartung", p.wartung.Uhr}, {"Wächter", p.waechter}, {"Telemetrie", p.telemetrie.Clock}},
	}
	if len(uhren[modusRecord]) != 4 || len(uhren[modusReplay]) != 3 {
		t.Error("die Menge der Stellen mit einer Uhr hat sich geändert")
	}
	for modus, stellen := range uhren {
		for _, s := range stellen {
			if s.uhr == nil || s.uhr != stellen[0].uhr {
				t.Errorf("Modus %s: %s teilt nicht die Uhr von %s", modus, s.name, stellen[0].name)
			}
		}
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

// stelleUhr ist ein Verbraucher der einen Uhr mit der Uhr, die er erhielt.
type stelleUhr struct {
	name string
	uhr  support.Clock
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
		{"mock ohne PublicBaseURL", func(k *bootstrap.Konfiguration) { k.Modus = "mock" }, istUngueltig, false},
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

// TestCADatei fängt B8 für UPSTREAM_CA_FILE (SPEC-139 Nr. 2): Modus {record, replay} × Pfad {nicht
// vorhanden, Verzeichnis statt Datei} × Zugang {Build mit Literal, Start mit Umgebung} = 8. Build liest die
// Datei nur in record; Start meldet sie in jedem Modus, und nichts bleibt geöffnet.
func TestCADatei(t *testing.T) {
	const zeile = "ODR-E2001 UPSTREAM_CA_FILE: Datei nicht lesbar"
	n := 0
	for _, modus := range modi() {
		for _, art := range []string{"nicht vorhanden", "Verzeichnis"} {
			for _, zugang := range []string{"Build", "Start"} {
				n++
				t.Run(modus+"/"+art+"/"+zugang, func(t *testing.T) {
					pfad := t.TempDir()
					if art == "nicht vorhanden" {
						pfad = filepath.Join(pfad, "keine-ca.pem")
					}
					h := neuHaken()
					if zugang == "Start" {
						// Die Frist ist die Testbremse: meldet der Loader nichts, läuft Start bis zu ihr.
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						var stderr syncPuffer
						code := bootstrap.StartMit(ctx, umgebungFuer(t, modus, 1, "UPSTREAM_CA_FILE="+pfad), &stderr, h.bauteile())
						if zeilen := meldungsZeilen(stderr.String()); code != 1 || !slices.Equal(zeilen, []string{zeile}) {
							t.Errorf("Rückgabe %d, Meldungen %q", code, zeilen)
						}
						if h.p.zaehle(pAblage) != 0 {
							t.Error("die Ablage wurde geöffnet")
						}
						return
					}
					k := basisKonfig(t.TempDir(), modus, dienstA("http://127.0.0.1:9/a"))
					k.UpstreamCAFile = pfad
					hd, err := bootstrap.BuildMit(context.Background(), k, nil, h.bauteile())
					if modus == modusReplay {
						if err != nil {
							t.Fatalf("Build im Modus replay: %v", err)
						}
						if err := hd.Herunterfahren(context.Background()); err != nil {
							t.Error(err)
						}
						return
					}
					if err == nil || bootstrap.Meldung(err) != zeile {
						t.Errorf("Fehler %v", err)
					}
					if h.p.zaehle(pAblage) != 0 {
						t.Error("die Ablage wurde geöffnet")
					}
				})
			}
		}
	}
	if n != 8 {
		t.Errorf("Menge hat %d Elemente, erwartet 8", n)
	}
}

// TestBlobVerzeichnisRechte fängt B9 (SPEC-139 Nr. 2): Modus × Zustand von blobs {fehlt → 0700; vorhanden
// mit 0755, 0750, 0777 → unverändert}. chmod setzt die Rechte der Fixture unabhängig von der umask.
func TestBlobVerzeichnisRechte(t *testing.T) {
	zustaende := []struct {
		name      string
		vorhanden bool
		vorher    os.FileMode
		soll      os.FileMode
	}{
		{"fehlt", false, 0, 0o700},
		{"vorhanden 0755", true, 0o755, 0o755},
		{"vorhanden 0750", true, 0o750, 0o750},
		{"vorhanden 0777", true, 0o777, 0o777},
	}
	for _, modus := range modi() {
		for _, z := range zustaende {
			t.Run(modus+"/"+z.name, func(t *testing.T) {
				dir := t.TempDir()
				blobs := filepath.Join(dir, verzeichnisB)
				if z.vorhanden {
					if err := os.Mkdir(blobs, 0o700); err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(blobs, z.vorher); err != nil {
						t.Fatal(err)
					}
				}
				hd, err := bootstrap.BuildMit(context.Background(), basisKonfig(dir, modus, dienstA("http://127.0.0.1:9/a")), nil, neuHaken().bauteile())
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = hd.Herunterfahren(context.Background()) })
				info, err := os.Stat(blobs)
				if err != nil {
					t.Fatal(err)
				}
				if got := info.Mode().Perm(); got != z.soll {
					t.Errorf("Rechte %04o, erwartet %04o", got, z.soll)
				}
			})
		}
	}
}

// ctxMarke ist der Schlüssel, mit dem ein Test den ctx eines Aufrufs wiedererkennt.
type ctxMarke struct{}

// TestAufrufstellen fängt B11 (SPEC-139 Nr. 6): Modus × Aufrufstelle {Open der Ablage (Build), Startup,
// Create, Get, OpenInitialSession (nur record), Listen}: der Fake hält den ctx je Aufruf fest; er trägt den
// Wert der Quelle und folgt ihrem Abbruch (Background und WithoutCancel erfüllen nur eines von beiden).
// Listen erhält das Netz tcp und die Adresse aus LISTEN. Get wird erreicht, weil der Service vorab liegt.
func TestAufrufstellen(t *testing.T) {
	for _, modus := range modi() {
		t.Run(modus, func(t *testing.T) {
			dir := t.TempDir()
			db := oeffneAblage(context.Background(), t, dir)
			if err := db.Metadata().Create(context.Background(), model.Service{ID: "a", ServiceRoot: "/a", ODataVersion: "4.0"}); err != nil {
				t.Fatal(err)
			}
			_ = db.Close()
			neu := func(marke string) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxMarke{}, marke))
				return ctx, cancel
			}
			ctxBuild, cancelBuild := neu("build")
			defer cancelBuild()
			ctxHoch, cancelHoch := neu("hochfahren")
			defer cancelHoch()
			h := neuHaken()
			k := basisKonfig(dir, modus, dienstA("http://127.0.0.1:9/a"))
			hd, err := bootstrap.BuildMit(ctxBuild, k, nil, h.bauteile())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = hd.Herunterfahren(context.Background()) })
			if err := hd.Hochfahren(ctxHoch); err != nil {
				t.Fatal(err)
			}
			cancelBuild()
			cancelHoch()
			stellen := map[string]string{pAblage: "build", pWartung: "hochfahren", pDienst: "hochfahren", pDienstGet: "hochfahren", pListen: "hochfahren"}
			if modus == modusRecord {
				stellen[pSession] = "hochfahren"
			}
			for stelle, marke := range stellen {
				a, gerufen := h.aufrufVon(stelle)
				switch {
				case !gerufen:
					t.Errorf("Aufrufstelle %s wurde nicht gerufen", stelle)
				case a.marke != marke:
					t.Errorf("Aufrufstelle %s erhielt nicht den ctx des Aufrufs (Wert %v)", stelle, a.marke)
				case !a.abgebrochen():
					t.Errorf("Aufrufstelle %s folgt dem Abbruch des ctx nicht", stelle)
				}
			}
			if h.listenNetz != "tcp" || h.listenAdr != k.Listen {
				t.Errorf("Listen mit Netz %q und Adresse %q, erwartet tcp und %q", h.listenNetz, h.listenAdr, k.Listen)
			}
		})
	}
}

// TestAblageFehler fängt B8 für den Fehler der Ablage (SPEC-139 Nr. 2, SPEC-109 Nr. 5): Build gibt ihn
// unverändert zurück (errors.Is findet die Quelle), und die Meldung ist der Code seiner Abbildung, nie
// ODR-E9001. Menge 1: echte Ablage {DATA_DIR fehlt, DATA_DIR ist eine Datei, odata-recorder.db ist ein
// Verzeichnis, Datei beschädigt} im Modus record. Menge 2: Fehler der Ablage, den der Haken liefert, {roh,
// umhüllt} × Quelle {ErrBeschaedigt (ODR-E3001), ErrPfad (ODR-E2001)} × Modus × Zugang {Build, Start}; die
// Quellen haben nie ODR-E9001 als Abbildung, daher unterscheidet jeder Fall einen verflachten Fehler.
func TestAblageFehler(t *testing.T) {
	ctx := context.Background()
	zeile := func(c model.Code) string { return c.String() + ": " + c.Message() }
	echte := []struct {
		name   string
		richte func(t *testing.T, dir string) string
		quelle error
		code   model.Code
	}{
		{"DATA_DIR fehlt", func(_ *testing.T, dir string) string { return filepath.Join(dir, "fehlt") }, outbound.ErrPfad, model.CodeE2001},
		{"DATA_DIR ist eine Datei", func(t *testing.T, dir string) string {
			p := filepath.Join(dir, "datei")
			if err := os.WriteFile(p, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			return p
		}, outbound.ErrPfad, model.CodeE2001},
		{"db ist ein Verzeichnis", func(t *testing.T, dir string) string {
			if err := os.Mkdir(filepath.Join(dir, dateiDB), 0o700); err != nil {
				t.Fatal(err)
			}
			return dir
		}, outbound.ErrPfad, model.CodeE2001},
		{"Datei beschädigt", func(t *testing.T, dir string) string {
			if err := os.WriteFile(filepath.Join(dir, dateiDB), []byte("keine SQLite-Datenbank, aber lang genug fuer einen Kopf"), 0o600); err != nil {
				t.Fatal(err)
			}
			return dir
		}, outbound.ErrBeschaedigt, model.CodeE3001},
	}
	for _, f := range echte {
		t.Run("echt/"+f.name, func(t *testing.T) {
			dir := f.richte(t, t.TempDir())
			_, err := bootstrap.BuildMit(ctx, basisKonfig(dir, modusRecord, dienstA("http://127.0.0.1:9/a")), nil, bootstrap.StandardBauteile())
			if !errors.Is(err, f.quelle) || bootstrap.Meldung(err) != zeile(f.code) {
				t.Errorf("Fehler %v (Meldung %q), erwartet %v und %q", err, bootstrap.Meldung(err), f.quelle, zeile(f.code))
			}
		})
	}
	quellen := []struct {
		name   string
		fehler error
		code   model.Code
	}{{"ErrBeschaedigt", outbound.ErrBeschaedigt, model.CodeE3001}, {"ErrPfad", outbound.ErrPfad, model.CodeE2001}}
	formen := map[string]func(error) error{
		"roh":      func(e error) error { return e },
		"umhuellt": func(e error) error { return fmt.Errorf("ablage: %w", e) },
	}
	n := 0
	for _, modus := range modi() {
		for _, q := range quellen {
			for form, bilde := range formen {
				n += 2
				t.Run("Build/"+modus+"/"+q.name+"/"+form, func(t *testing.T) {
					h := neuHaken()
					h.fehlerIn = map[string]error{pAblage: bilde(q.fehler)}
					_, err := bootstrap.BuildMit(ctx, basisKonfig(t.TempDir(), modus, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
					if !errors.Is(err, q.fehler) || bootstrap.Meldung(err) != zeile(q.code) {
						t.Errorf("Fehler %v (Meldung %q), erwartet %v und %q", err, bootstrap.Meldung(err), q.fehler, zeile(q.code))
					}
				})
				t.Run("Start/"+modus+"/"+q.name+"/"+form, func(t *testing.T) {
					h := neuHaken()
					h.fehlerIn = map[string]error{pAblage: bilde(q.fehler)}
					var stderr syncPuffer
					code := bootstrap.StartMit(ctx, umgebungFuer(t, modus, 1), &stderr, h.bauteile())
					if zeilen := meldungsZeilen(stderr.String()); code != 1 || !slices.Equal(zeilen, []string{zeile(q.code)}) {
						t.Errorf("Rückgabe %d, Meldungen %q, erwartet %q", code, zeilen, zeile(q.code))
					}
				})
			}
		}
	}
	if n != 16 {
		t.Errorf("Menge 2 hat %d Elemente, erwartet 16", n)
	}
}

// TestLogNilVerwirft fängt SPEC-139 Nr. 1 („log nil verwirft die Einträge“): Build ohne Logger schreibt
// nirgends hin, auch nicht in den Standard-Logger von slog; derselbe Ablauf mit dem Standard-Logger als
// übergebenem log schreibt (der Ablauf löst den ERROR-Eintrag des abweichenden Service aus).
func TestLogNilVerwirft(t *testing.T) {
	ctx := context.Background()
	var standard syncPuffer
	vorher := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&standard, nil)))
	t.Cleanup(func() { slog.SetDefault(vorher) })
	abweichend := func(t *testing.T) string {
		dir := t.TempDir()
		db := oeffneAblage(ctx, t, dir)
		if err := db.Metadata().Create(ctx, model.Service{ID: "a", ServiceRoot: "/x", ODataVersion: "4.0"}); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		return dir
	}
	for name, log := range map[string]*slog.Logger{"übergeben": slog.Default(), "nil": nil} {
		standard.reset()
		_, err := lauf(ctx, t, basisKonfig(abweichend(t), modusRecord, dienstA("http://127.0.0.1:9/a")), log, bootstrap.StandardBauteile())
		if err == nil {
			t.Fatalf("%s: Hochfahren gelingt trotz abweichendem Service", name)
		}
		if got := standard.String() != ""; got != (name == "übergeben") {
			t.Errorf("%s: Standard-Logger hat geschrieben: %v (%q)", name, got, standard.String())
		}
	}
}

// TestReplayTelemetrieSchreibtNirgendhin fängt SPEC-139 Nr. 3: im Modus replay erhält die Telemetrie des
// Replay-Dienstes io.Discard, in record wird keine gebaut. Pinnt den festen Wert; ein anderes Ziel
// (SPEC-232) ändert Zusage und Test zugleich.
func TestReplayTelemetrieSchreibtNirgendhin(t *testing.T) {
	for _, modus := range modi() {
		h := neuHaken()
		hd, err := bootstrap.BuildMit(context.Background(), basisKonfig(t.TempDir(), modus, dienstA("http://127.0.0.1:9/a")), nil, h.bauteile())
		if err != nil {
			t.Fatal(err)
		}
		_ = hd.Herunterfahren(context.Background())
		if gebaut := h.telemetrie.Clock != nil; gebaut != (modus == modusReplay) {
			t.Errorf("%s: Telemetrie gebaut: %v", modus, gebaut)
		}
		if modus == modusReplay && h.telemetrie.Writer != io.Discard {
			t.Errorf("%s: Writer der Telemetrie %v, erwartet io.Discard", modus, h.telemetrie.Writer)
		}
	}
}
