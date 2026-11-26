package bootstrap_test

import (
	"context"
	"encoding/json"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/inspection"
)

// einstellungenDefault sind die Einträge zur geladenen Konfiguration ohne weiteren Schlüssel.
func einstellungenDefault() map[string]string {
	return map[string]string{
		"REPLAY_STRATEGY": "sequential", "REPLAY_ASSIGNMENT": "explicit", "REPLAY_REDACTED_BODIES": "false",
		"MATCH_VARY_HEADERS": "Accept,Content-Type,If-Match,Odata-Version,Prefer", "MATCH_INVALID_ENCODING": "literal",
		"MATCH_BODY_MODE": "bytes", "REDACT_HEADERS": "Authorization,Cookie,Proxy-Authorization,Set-Cookie,X-Api-Key,X-Auth-Token,X-Csrf-Token",
		"REDACT_QUERY_PARAMS": "access_token,api_key,apikey,client_secret,code,key,sig,signature,token",
		"REDACT_JSON_PATHS":   "$..access_token,$..api_key,$..client_secret,$..passwd,$..password,$..refresh_token,$..secret",
		"MAX_CAPTURE_BODY":    "8388608", "UPSTREAM_TIMEOUT": "30000ms", "UPSTREAM_TLS_VERIFY": "true",
		"UPSTREAM_TLS_INSECURE_SKIP_VERIFY": "false", "SHUTDOWN_TIMEOUT": "10000ms", "RETENTION_MAX_AGE": "0ms",
		"SQLITE_BUSY_TIMEOUT": "5000ms", "VOLATILE_HEADERS": "Age,Date,Server,Traceparent,Via,X-Correlation-Id,X-Request-Id",
		"LINK_REWRITE": "off",
	}
}

// einstellungenAbweichend sind die Einträge zu abweichendeKonfig: 30 Tage sind 2592000000 ms, die Frist der Ablage steht an der
// Grenze des Ladens (2147483647 ms); der Rest unter 1 ms entfällt.
func einstellungenAbweichend() map[string]string {
	return map[string]string{
		"REPLAY_STRATEGY": "strict", "REPLAY_ASSIGNMENT": "ordered", "REPLAY_REDACTED_BODIES": "true",
		"MATCH_VARY_HEADERS": "X-A,X-B", "MATCH_INVALID_ENCODING": "reject", "MATCH_BODY_MODE": "json-canonical",
		"REDACT_HEADERS": "X-E,X-F", "REDACT_QUERY_PARAMS": "q1,q2", "REDACT_JSON_PATHS": "$..g1,$..g2",
		"MAX_CAPTURE_BODY": "1073741824", "UPSTREAM_TIMEOUT": "2592000000ms", "UPSTREAM_TLS_VERIFY": "false",
		"UPSTREAM_TLS_INSECURE_SKIP_VERIFY": "true", "SHUTDOWN_TIMEOUT": "2592000000ms", "RETENTION_MAX_AGE": "2592000000ms",
		"SQLITE_BUSY_TIMEOUT": "2147483647ms", "VOLATILE_HEADERS": "X-C,X-D", "LINK_REWRITE": "anders",
	}
}

// Die Werte der Konfiguration, die in keiner Antwort stehen dürfen: Datenverzeichnis, Adresse und Upstream.
const (
	canaryDir      = "canary-datenordner"
	canaryListen   = "127.0.0.1:39999"
	canaryUpstream = "http://canary-upstream.test/o"
)

func geladeneKonfig(t *testing.T, modus string) bootstrap.Konfiguration {
	t.Helper()
	k, err := bootstrap.LadeKonfiguration([]string{
		"MODE=" + modus, "DATA_DIR=" + filepath.Join(t.TempDir(), canaryDir), "LISTEN=" + canaryListen,
		"UPSTREAM_URL=" + canaryUpstream, "SERVICE_ID=x1", "SERVICE_ROOT_PATH=/r1",
	})
	if err != nil {
		t.Fatalf("%s: %v", modus, err)
	}
	return k
}

func abweichendeKonfig(t *testing.T, modus string) bootstrap.Konfiguration {
	t.Helper()
	dir := filepath.Join(t.TempDir(), canaryDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	k := routenKonfig(dir, modus, canaryUpstream)
	const dauer = 30*24*time.Hour + 999999*time.Nanosecond
	k.Listen, k.LinkRewrite, k.MaxCaptureBody = canaryListen, "anders", 1<<30
	k.ReplayStrategy, k.ReplayAssignment, k.ReplayRedactedBodies = "strict", "ordered", true
	k.MatchInvalidEncoding, k.MatchBodyMode, k.UpstreamTLSVerify = "reject", "json-canonical", false
	k.UpstreamTimeout, k.ShutdownTimeout, k.RetentionMaxAge = dauer, dauer, dauer
	k.SQLiteBusyTimeout = 2147483647*time.Millisecond + 999999*time.Nanosecond
	k.MatchVaryHeaders, k.VolatileHeaders = []string{"X-A", "X-B"}, []string{"X-C", "X-D"}
	k.RedactHeaders, k.RedactQueryParams, k.RedactJSONPaths = []string{"X-E", "X-F"}, []string{"q1", "q2"}, []string{"$..g1", "$..g2"}
	return k
}

// diagnose ist die Antwort von GET /__admin/diagnostics.
type diagnose struct {
	Mode     string            `json:"mode"`
	Settings map[string]string `json:"settings"`
	Errors   []struct {
		Code  string `json:"code"`
		Count int    `json:"count"`
	} `json:"errors"`
}

// holeDiagnose liest die Diagnose, prüft den Canary und liefert zusätzlich den Zähler von ODR-E5001.
func holeDiagnose(t *testing.T, h *bootstrap.Handle) (diagnose, int) {
	t.Helper()
	body := rufe(t, h, 200, "GET", "/__admin/diagnostics", "", "", "")
	for _, wert := range []string{canaryDir, "39999", "canary-upstream"} {
		if strings.Contains(body, wert) {
			t.Errorf("Diagnose trägt %q: %s", wert, body)
		}
	}
	var d diagnose
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		t.Fatal(err)
	}
	anzahl := -1
	for _, e := range d.Errors {
		if e.Code == "ODR-E5001" {
			anzahl = e.Count
		}
	}
	return d, anzahl
}

// inspektionGebaut baut und liefert die inspection.Konfig, die der Haken NeueInspektion sah.
func inspektionGebaut(t *testing.T, k bootstrap.Konfiguration) (*bootstrap.Handle, inspection.Konfig) {
	t.Helper()
	var got inspection.Konfig
	h, _ := adminBau(t, k, nil, func(b *bootstrap.Bauteile) {
		echt := b.NeueInspektion
		b.NeueInspektion = func(c inspection.Konfig) (inbound.InspectionUseCase, error) { got = c; return echt(c) }
	})
	return h, got
}

// TestAdminInspektionVerdrahtet fängt A03: je Modus und Konfiguration die Konfig am Haken, Inspektion() am Handle und die Diagnose
// (18 Schlüssel x 2 Konfigurationen), den Zähler des einen Adapters, jede Liste ohne und mit 1 bis 3 Einträgen und im Modus record
// Exchanges und Coverage.
func TestAdminInspektionVerdrahtet(t *testing.T) {
	konfigurationen := map[string]struct {
		konfig func(*testing.T, string) bootstrap.Konfiguration
		soll   map[string]string
	}{"Defaults": {geladeneKonfig, einstellungenDefault()}, "abweichend": {abweichendeKonfig, einstellungenAbweichend()}}
	zellen := 0
	for name, f := range konfigurationen {
		for _, modus := range adminModi() {
			h, got := inspektionGebaut(t, f.konfig(t, modus))
			if !maps.Equal(got.Einstellungen, f.soll) || got.Modus != modus {
				t.Errorf("%s/%s: Konfig %q %v", name, modus, got.Modus, got.Einstellungen)
			}
			_, istExchanges := got.Exchanges.(*sqlite.RecordingRepository)
			_, istCoverage := got.Coverage.(*sqlite.CoverageRepository)
			if !istExchanges || !istCoverage || got.Zaehler != any(bootstrap.TelemetrieVon(h)) {
				t.Errorf("%s/%s: Ports %T, %T, %T", name, modus, got.Exchanges, got.Coverage, got.Zaehler)
			}
			if h.Inspektion() == nil {
				t.Fatalf("%s/%s: Inspektion() ist nil", name, modus)
			}
			if am, err := h.Inspektion().Diagnose(context.Background()); err != nil || am.Modus != modus || !maps.Equal(am.Einstellungen, f.soll) {
				t.Errorf("%s/%s: Diagnose am Handle %q %v, %v", name, modus, am.Modus, am.Einstellungen, err)
			}
			d, vorher := holeDiagnose(t, h)
			if d.Mode != modus || len(d.Errors) != len(model.AlleCodes()) || vorher != 0 {
				t.Errorf("%s/%s: Modus %q, %d Fehlerzähler, %d", name, modus, d.Mode, len(d.Errors), vorher)
			}
			for _, schluessel := range inspection.Erlaubt() {
				if d.Settings[schluessel] != f.soll[schluessel] {
					t.Errorf("%s/%s: %s ist %q, erwartet %q", name, modus, schluessel, d.Settings[schluessel], f.soll[schluessel])
				}
			}
			if len(d.Settings) != 18 {
				t.Errorf("%s/%s: %d Einstellungen", name, modus, len(d.Settings))
			}
			bootstrap.TelemetrieVon(h).CountError(model.CodeE5001)
			if _, nachher := holeDiagnose(t, h); nachher != 1 {
				t.Errorf("%s/%s: Zähler %d nach CountError", name, modus, nachher)
			}
		}
		zellen += len(f.soll)
	}
	if zellen != 36 || len(inspection.Erlaubt()) != 18 {
		t.Errorf("%d Zellen, %d erlaubte Schlüssel", zellen, len(inspection.Erlaubt()))
	}
	// Jede Liste je Länge, mit eigenem Anfang je Liste: nil (-1), leer, ein Eintrag, zwei, drei (5 x 5 Zellen).
	listen := [5][2]string{
		{"MATCH_VARY_HEADERS", "X-V"}, {"VOLATILE_HEADERS", "X-O"}, {"REDACT_HEADERS", "X-R"},
		{"REDACT_QUERY_PARAMS", "q"}, {"REDACT_JSON_PATHS", "$..j"},
	}
	laengen := 0
	for _, n := range []int{-1, 0, 1, 2, 3} {
		var werte [5][]string
		for i, l := range listen {
			if n >= 0 {
				werte[i] = []string{l[1] + "1", l[1] + "2", l[1] + "3"}[:n]
			}
		}
		k := routenKonfig(t.TempDir(), modusReplay, "")
		k.MatchVaryHeaders, k.VolatileHeaders, k.RedactHeaders, k.RedactQueryParams, k.RedactJSONPaths = werte[0], werte[1], werte[2], werte[3], werte[4]
		_, got := inspektionGebaut(t, k)
		for _, l := range listen {
			schluessel, a := l[0], l[1]
			soll := map[int]string{-1: "none", 0: "none", 1: a + "1", 2: a + "1," + a + "2", 3: a + "1," + a + "2," + a + "3"}[n]
			if got.Einstellungen[schluessel] != soll {
				t.Errorf("Länge %d: %s ist %q, erwartet %q", n, schluessel, got.Einstellungen[schluessel], soll)
			}
			laengen++
		}
	}
	if laengen != 25 {
		t.Errorf("%d Zellen der Listen", laengen)
	}

	h, db, _, quelle := adminAufnahme(t)
	rufe(t, h, 201, "POST", "/__admin/imports", "", "", `{"import_id":"imp-a","service_id":"x1","source_session_ids":["`+quelle+`"]}`)
	imp, err := db.Imports().Get(context.Background(), "imp-a")
	if err != nil {
		t.Fatal(err)
	}
	exchanges := rufe(t, h, 200, "GET", "/__admin/sessions/{id}/exchanges", quelle, "", "")
	if !strings.Contains(exchanges, `"total":2`) || !strings.Contains(exchanges, `"path":"/Products"`) || strings.Contains(exchanges, "canary") {
		t.Errorf("Exchanges %s", exchanges)
	}
	coverage := rufe(t, h, 200, "GET", "/__admin/coverage", "", "baseline_id="+url.QueryEscape(imp.BaselineSnapshotID), "")
	if !strings.Contains(coverage, `"Products"`) {
		t.Errorf("Coverage %s", coverage)
	}
}
