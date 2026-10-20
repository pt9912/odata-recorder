package bootstrap_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// basis ist eine gültige Umgebung im Modus record; spätere Einträge überschreiben frühere.
func basis(t *testing.T) []string {
	t.Helper()
	return []string{"MODE=record", "UPSTREAM_URL=https://example.test/odata", "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}
}

func lade(t *testing.T, env ...string) (bootstrap.Konfiguration, []string) {
	t.Helper()
	k, err := bootstrap.LadeKonfiguration(env)
	if err == nil {
		return k, nil
	}
	var kf *bootstrap.KonfigFehler
	if !errors.As(err, &kf) {
		t.Fatalf("Fehlertyp %T", err)
	}
	return k, kf.Zeilen
}

// schluessel liefert die Schlüssel der Zeilen `ODR-E2001 <Schlüssel>: <Grund>`.
func schluessel(t *testing.T, zeilen []string) []string {
	t.Helper()
	var out []string
	for _, z := range zeilen {
		rest, ok := strings.CutPrefix(z, "ODR-E2001 ")
		k, _, ok2 := strings.Cut(rest, ": ")
		if !ok || !ok2 {
			t.Fatalf("Zeile hat nicht die Form: %q", z)
		}
		out = append(out, k)
	}
	return out
}

func erwarteGueltig(t *testing.T, env ...string) bootstrap.Konfiguration {
	t.Helper()
	k, zeilen := lade(t, env...)
	if len(zeilen) != 0 {
		t.Fatalf("Verstöße: %v", zeilen)
	}
	return k
}

func erwarteVerstoss(t *testing.T, key string, env ...string) {
	t.Helper()
	_, zeilen := lade(t, env...)
	if !slices.Contains(schluessel(t, zeilen), key) {
		t.Errorf("kein Verstoß für %s in %v (Umgebung %v)", key, zeilen, env[len(env)-1:])
	}
}

// erwarteGrund verlangt eine Zeile für den Schlüssel, deren Grund den Text enthält.
func erwarteGrund(t *testing.T, key, grund string, env ...string) {
	t.Helper()
	_, zeilen := lade(t, env...)
	for _, z := range zeilen {
		if strings.HasPrefix(z, "ODR-E2001 "+key+": ") && strings.Contains(z, grund) {
			return
		}
	}
	t.Errorf("keine Zeile %s mit Grund %q in %v", key, grund, zeilen)
}

func TestKonfigLeer(t *testing.T) {
	standard := erwarteGueltig(t, basis(t)...)
	for _, key := range []string{
		"LISTEN", "SERVICE_ID", "SERVICE_ROOT_PATH", "REPLAY_STRATEGY", "REPLAY_ASSIGNMENT", "REPLAY_REDACTED_BODIES",
		"MATCH_VARY_HEADERS", "MATCH_INVALID_ENCODING", "MATCH_BODY_MODE", "REDACT_HEADERS", "REDACT_QUERY_PARAMS",
		"REDACT_JSON_PATHS", "MAX_CAPTURE_BODY", "UPSTREAM_TIMEOUT", "UPSTREAM_CA_FILE", "UPSTREAM_TLS_VERIFY",
		"UPSTREAM_TLS_INSECURE_SKIP_VERIFY", "SHUTDOWN_TIMEOUT", "RETENTION_MAX_AGE", "SQLITE_BUSY_TIMEOUT",
		"VOLATILE_HEADERS", "LINK_REWRITE", "SERVICES_FILE", "STRICT_COVERAGE", "ADMIN_TOKEN_FILE",
	} {
		for _, leer := range []string{"", "  ", "\t"} {
			k := erwarteGueltig(t, append(basis(t), key+"="+leer)...)
			k.DataDir = standard.DataDir
			if !reflect.DeepEqual(k, standard) {
				t.Errorf("%s=%q weicht vom Ungesetzten ab", key, leer)
			}
		}
	}
	if k := erwarteGueltig(t, append(basis(t), "MAX_CAPTURE_BODY= 8 MiB ")...); k.MaxCaptureBody != 8<<20 {
		t.Errorf("Leerraum am Rand: %d", k.MaxCaptureBody)
	}
	if k := erwarteGueltig(t, append(basis(t), "LISTEN= 0.0.0.0:99 ")...); k.Listen != "0.0.0.0:99" {
		t.Errorf("Leerraum bei LISTEN: %q", k.Listen)
	}
	erwarteVerstoss(t, "MODE", "MODE=", "DATA_DIR="+t.TempDir()+"/x") // leer ist ungesetzt: Default mock scheitert
}

func TestKonfigDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	k := erwarteGueltig(t, basis(t)...)
	soll := bootstrap.Konfiguration{
		Modus: "record", Listen: "127.0.0.1:8080", DataDir: k.DataDir,
		Dienste:        []bootstrap.Dienst{{ID: "default", Root: "/", UpstreamURL: "https://example.test/odata"}},
		ReplayStrategy: "sequential", ReplayAssignment: "explicit", ReplayRedactedBodies: false,
		MatchVaryHeaders:     []string{"Accept", "Content-Type", "If-Match", "Odata-Version", "Prefer"},
		MatchInvalidEncoding: "literal", MatchBodyMode: "bytes",
		RedactHeaders:     []string{"Authorization", "Cookie", "Proxy-Authorization", "Set-Cookie", "X-Api-Key", "X-Auth-Token", "X-Csrf-Token"},
		RedactQueryParams: []string{"access_token", "api_key", "apikey", "client_secret", "code", "key", "sig", "signature", "token"},
		RedactJSONPaths:   []string{"$..access_token", "$..api_key", "$..client_secret", "$..passwd", "$..password", "$..refresh_token", "$..secret"},
		MaxCaptureBody:    8 << 20, UpstreamTimeout: 30 * time.Second, UpstreamTLSVerify: true,
		ShutdownTimeout: 10 * time.Second, RetentionMaxAge: 0, SQLiteBusyTimeout: 5000 * time.Millisecond,
		VolatileHeaders: []string{"Age", "Date", "Server", "Traceparent", "Via", "X-Correlation-Id", "X-Request-Id"},
		LinkRewrite:     "off",
	}
	if !reflect.DeepEqual(k, soll) {
		t.Errorf("Defaults weichen ab:\n got %+v\nwant %+v", k, soll)
	}
	_, zeilen := lade(t, "MODE=record")
	if len(zeilen) != 1 || !strings.HasPrefix(zeilen[0], "ODR-E2001 UPSTREAM_URL: ") {
		t.Errorf("record ohne UPSTREAM_URL: %v", zeilen)
	}
	r := erwarteGueltig(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "r"))
	if r.Dienste[0].UpstreamURL != "" || r.Modus != "replay" {
		t.Errorf("replay: %+v", r)
	}
	// DATA_DIR ungesetzt: der Default ist `data`, relativ zum Arbeitsverzeichnis, und wird dort angelegt.
	d := erwarteGueltig(t, "MODE=replay")
	if d.DataDir != "data" {
		t.Errorf("DataDir %q", d.DataDir)
	}
	if fi, err := os.Stat("data"); err != nil || !fi.IsDir() {
		t.Errorf("data liegt nicht im Arbeitsverzeichnis: %v", err)
	}
}

func TestKonfigListenErgaenzen(t *testing.T) {
	k := erwarteGueltig(t, append(basis(t), "REDACT_HEADERS=x-custom, Authorization", "REDACT_QUERY_PARAMS=mein_param", "REDACT_JSON_PATHS=$..pin")...)
	if !slices.Contains(k.RedactHeaders, "Cookie") || !slices.Contains(k.RedactHeaders, "X-Custom") || len(k.RedactHeaders) != 8 {
		t.Errorf("REDACT_HEADERS: %v", k.RedactHeaders)
	}
	if !slices.Contains(k.RedactQueryParams, "token") || !slices.Contains(k.RedactQueryParams, "mein_param") {
		t.Errorf("REDACT_QUERY_PARAMS: %v", k.RedactQueryParams)
	}
	if !slices.Contains(k.RedactJSONPaths, "$..password") || !slices.Contains(k.RedactJSONPaths, "$..pin") {
		t.Errorf("REDACT_JSON_PATHS: %v", k.RedactJSONPaths)
	}
	// MATCH_VARY_HEADERS ersetzt den Default; Reihenfolge und Schreibweise ohne Wirkung
	a := erwarteGueltig(t, append(basis(t), "MATCH_VARY_HEADERS=Accept,accept, ETAG")...)
	if !slices.Equal(a.MatchVaryHeaders, []string{"Accept", "Etag"}) {
		t.Errorf("MATCH_VARY_HEADERS: %v", a.MatchVaryHeaders)
	}
}

func TestKonfigNone(t *testing.T) {
	k := erwarteGueltig(t, append(basis(t), "MATCH_VARY_HEADERS=none", "VOLATILE_HEADERS=none")...)
	if k.MatchVaryHeaders == nil || len(k.MatchVaryHeaders) != 0 || len(k.VolatileHeaders) != 0 {
		t.Errorf("none: %v %v", k.MatchVaryHeaders, k.VolatileHeaders)
	}
	for _, key := range []string{"MATCH_VARY_HEADERS", "VOLATILE_HEADERS"} {
		for _, w := range []string{"None", "NONE", "nOne", "none,Accept", "Accept,none", "none,none", " Accept , none "} {
			_, zeilen := lade(t, append(basis(t), key+"="+w)...)
			if len(zeilen) != 1 || !strings.HasPrefix(zeilen[0], "ODR-E2001 "+key+": none ist nur") {
				t.Errorf("%s=%q: %v", key, w, zeilen)
			}
		}
	}
	for _, key := range []string{"REDACT_HEADERS", "REDACT_QUERY_PARAMS", "REDACT_JSON_PATHS"} {
		for _, w := range []string{"none", "None", "none,x"} {
			_, zeilen := lade(t, append(basis(t), key+"="+w)...)
			if len(zeilen) != 1 || !strings.HasPrefix(zeilen[0], "ODR-E2001 "+key+": none ist nur") {
				t.Errorf("%s=%q: %v", key, w, zeilen)
			}
		}
	}
}

// TestKonfigHeaderSchreibweise: Header-Mengen kommen in kanonischer Schreibweise (SPEC-110 Nr. 6).
func TestKonfigHeaderSchreibweise(t *testing.T) {
	k := erwarteGueltig(t, append(basis(t), "MATCH_VARY_HEADERS=odata-version,X-API-KEY,etag", "VOLATILE_HEADERS=x-request-id", "REDACT_HEADERS=x-api-key,ETAG")...)
	if want := []string{"Etag", "Odata-Version", "X-Api-Key"}; !slices.Equal(k.MatchVaryHeaders, want) {
		t.Errorf("MatchVaryHeaders %v", k.MatchVaryHeaders)
	}
	if !slices.Equal(k.VolatileHeaders, []string{"X-Request-Id"}) {
		t.Errorf("VolatileHeaders %v", k.VolatileHeaders)
	}
	if !slices.Contains(k.RedactHeaders, "X-Api-Key") || !slices.Contains(k.RedactHeaders, "Etag") || len(k.RedactHeaders) != 8 {
		t.Errorf("RedactHeaders %v", k.RedactHeaders)
	}
	d := erwarteGueltig(t, basis(t)...)
	for _, menge := range [][]string{d.MatchVaryHeaders, d.VolatileHeaders, d.RedactHeaders} {
		for _, h := range menge {
			if h != http.CanonicalHeaderKey(h) {
				t.Errorf("Default %q nicht kanonisch", h)
			}
		}
	}
}

func TestKonfigNichtGeliefert(t *testing.T) {
	for _, key := range []string{"ADMIN_LISTEN", "ADMIN_TOKEN_FILE", "HYBRID_FALLBACK", "HYBRID_REPLAY_FALLBACK", "FALLBACK_ALLOW_HOSTS",
		"FALLBACK_WRITE_ALLOW", "LISTEN_TLS_CERT", "LISTEN_TLS_KEY", "STRICT_COVERAGE", "PUBLIC_BASE_URL"} {
		erwarteVerstoss(t, key, append(basis(t), key+"=irgendwas")...)
		erwarteGueltig(t, append(basis(t), key+"=")...)
	}
}

func TestKonfigFremdeVariable(t *testing.T) {
	erwarteGueltig(t, append(basis(t), "PATH=/usr/bin", "HOME=/root", "ADMINISTRATOR=x", "FOO=bar", "ohne-gleichheitszeichen")...)
}

func TestKonfigEnum(t *testing.T) {
	for key, werte := range map[string][]string{
		"REPLAY_STRATEGY": {"sequential", "strict", "first", "last"}, "REPLAY_ASSIGNMENT": {"explicit", "ordered"},
		"MATCH_INVALID_ENCODING": {"literal", "reject"}, "MATCH_BODY_MODE": {"bytes", "json-canonical"}, "LINK_REWRITE": {"off"},
	} {
		for _, w := range werte {
			erwarteGueltig(t, append(basis(t), key+"="+w)...)
			erwarteVerstoss(t, key, append(basis(t), key+"="+strings.ToUpper(w))...)
			erwarteVerstoss(t, key, append(basis(t), key+"="+strings.ToUpper(w[:1])+w[1:])...)
		}
		erwarteVerstoss(t, key, append(basis(t), key+"=gibt-es-nicht")...)
	}
	erwarteVerstoss(t, "LINK_REWRITE", append(basis(t), "LINK_REWRITE=absolute")...)
	erwarteVerstoss(t, "MODE", append(basis(t), "MODE=Record")...)
	erwarteVerstoss(t, "MODE", append(basis(t), "MODE=sonstwas")...)
}

func TestKonfigBool(t *testing.T) {
	for _, key := range []string{"REPLAY_REDACTED_BODIES", "UPSTREAM_TLS_VERIFY", "UPSTREAM_TLS_INSECURE_SKIP_VERIFY"} {
		for _, w := range []string{"Record", "1", "0", "yes", "no", "TRUE", "False", "on"} {
			erwarteVerstoss(t, key, append(basis(t), key+"="+w)...)
		}
	}
	if !erwarteGueltig(t, append(basis(t), "REPLAY_REDACTED_BODIES=true")...).ReplayRedactedBodies {
		t.Error("true nicht gelesen")
	}
}

func TestKonfigDauer(t *testing.T) {
	gueltig := map[string]time.Duration{"30s": 30 * time.Second, "5000ms": 5 * time.Second, "30d": 30 * 24 * time.Hour, "720h": 720 * time.Hour, "1ms": time.Millisecond, "2m": 2 * time.Minute}
	for w, d := range gueltig {
		if k := erwarteGueltig(t, append(basis(t), "UPSTREAM_TIMEOUT="+w)...); k.UpstreamTimeout != d {
			t.Errorf("UPSTREAM_TIMEOUT=%s: %v", w, k.UpstreamTimeout)
		}
	}
	for _, w := range []string{"30", "-1s", "+1s", "1.5h", "1h30m", "1w", "0s", "0", "0ms", "s", "1 s", "1S", "٣s", "106752d", "9999999999999999999999s"} {
		for _, key := range []string{"UPSTREAM_TIMEOUT", "SHUTDOWN_TIMEOUT", "SQLITE_BUSY_TIMEOUT"} {
			erwarteVerstoss(t, key, append(basis(t), key+"="+w)...)
		}
	}
	// SQLITE_BUSY_TIMEOUT: höchstens 2147483647 ms
	if k := erwarteGueltig(t, append(basis(t), "SQLITE_BUSY_TIMEOUT=2147483647ms")...); k.SQLiteBusyTimeout != 2147483647*time.Millisecond {
		t.Errorf("Obergrenze: %v", k.SQLiteBusyTimeout)
	}
	erwarteVerstoss(t, "SQLITE_BUSY_TIMEOUT", append(basis(t), "SQLITE_BUSY_TIMEOUT=2147483648ms")...)
	// RETENTION_MAX_AGE lässt 0 zu, mit und ohne Einheit
	for _, w := range []string{"0", "0s", "0d"} {
		if k := erwarteGueltig(t, append(basis(t), "RETENTION_MAX_AGE="+w)...); k.RetentionMaxAge != 0 {
			t.Errorf("RETENTION_MAX_AGE=%s: %v", w, k.RetentionMaxAge)
		}
	}
	if k := erwarteGueltig(t, append(basis(t), "RETENTION_MAX_AGE=30d")...); k.RetentionMaxAge != 30*24*time.Hour {
		t.Errorf("30d: %v", k.RetentionMaxAge)
	}
	for _, w := range []string{"30", "-1d", "1.5d", "1d1h", "106752d"} {
		erwarteVerstoss(t, "RETENTION_MAX_AGE", append(basis(t), "RETENTION_MAX_AGE="+w)...)
	}
}

func TestKonfigGroesse(t *testing.T) {
	gueltig := map[string]int64{"8 MiB": 8 << 20, "8MiB": 8 << 20, "8388608": 8 << 20, "1GiB": 1 << 30, "1 GiB": 1 << 30, "1": 1, "1B": 1, "2 KiB": 2048, "1073741824": 1 << 30, "1024MiB": 1 << 30}
	for w, n := range gueltig {
		if k := erwarteGueltig(t, append(basis(t), "MAX_CAPTURE_BODY="+w)...); k.MaxCaptureBody != n {
			t.Errorf("MAX_CAPTURE_BODY=%q: %d", w, k.MaxCaptureBody)
		}
	}
	for _, w := range []string{"0", "0 MiB", "8 MB", "8  MiB", "1025MiB", "1073741825", "-1", "+1", "8.5MiB", "MiB", "8 mib", "8\tMiB", "99999999999999999999", "1048577KiB"} {
		erwarteVerstoss(t, "MAX_CAPTURE_BODY", append(basis(t), "MAX_CAPTURE_BODY="+w)...)
	}
}

func TestKonfigListe(t *testing.T) {
	for _, w := range []string{"a,,b", "a,", ",a", ",", "Bad Header", "a b", "Ä"} {
		erwarteVerstoss(t, "MATCH_VARY_HEADERS", append(basis(t), "MATCH_VARY_HEADERS="+w)...)
		erwarteVerstoss(t, "REDACT_HEADERS", append(basis(t), "REDACT_HEADERS="+w)...)
		erwarteVerstoss(t, "VOLATILE_HEADERS", append(basis(t), "VOLATILE_HEADERS="+w)...)
	}
	for _, w := range []string{"a,,b", "a,", "a b", "a=b", "a&b", "a#b", "a;b,"} {
		erwarteVerstoss(t, "REDACT_QUERY_PARAMS", append(basis(t), "REDACT_QUERY_PARAMS="+w)...)
	}
	for _, w := range []string{"$.a", "$..a.b", "a", "$..", "$..a b", "$..a,,$..b", "$..[0]", "$..a*"} {
		erwarteVerstoss(t, "REDACT_JSON_PATHS", append(basis(t), "REDACT_JSON_PATHS="+w)...)
	}
	k := erwarteGueltig(t, append(basis(t), "REDACT_JSON_PATHS=$..a-b_1 , $..a-b_1,$..Z")...)
	if !slices.Contains(k.RedactJSONPaths, "$..a-b_1") || !slices.Contains(k.RedactJSONPaths, "$..Z") || len(k.RedactJSONPaths) != 9 {
		t.Errorf("Pfade: %v", k.RedactJSONPaths)
	}
}

func TestKonfigLISTEN(t *testing.T) {
	for _, w := range []string{"127.0.0.1:1", "0.0.0.0:65535", "localhost:8080", "[::1]:8080", "host.example:80", "a_b.c-d:80"} {
		if k := erwarteGueltig(t, append(basis(t), "LISTEN="+w)...); k.Listen != w {
			t.Errorf("LISTEN=%s: %q", w, k.Listen)
		}
	}
	for _, w := range []string{"127.0.0.1:0", "127.0.0.1:65536", ":8080", "[]:80", "127.0.0.1", "127.0.0.1:", "127.0.0.1:+80", "127.0.0.1:-1", "::1:80", "[abc]:80", "a b:80", "a/b:80", "a_b/c:80", "[::1%eth0]:80", "[a b]:80", "ä:80", "127.0.0.1:80x", "http://h:80", "host:99999999999999999999"} {
		erwarteVerstoss(t, "LISTEN", append(basis(t), "LISTEN="+w)...)
	}
}

func TestKonfigServiceID(t *testing.T) {
	for _, w := range []string{"a", "A.b_c-1", strings.Repeat("a", 64)} {
		if k := erwarteGueltig(t, append(basis(t), "SERVICE_ID="+w)...); k.Dienste[0].ID != w {
			t.Errorf("SERVICE_ID=%s: %v", w, k.Dienste)
		}
	}
	for _, w := range []string{strings.Repeat("a", 65), "a b", "a/b", "ä", "a:b"} {
		erwarteVerstoss(t, "SERVICE_ID", append(basis(t), "SERVICE_ID="+w)...)
	}
}

func TestKonfigRoot(t *testing.T) {
	for _, w := range []string{"/", "/a", "/a/b", "/a.b/c_d~e-f", "/..a", "/a..b"} {
		if k := erwarteGueltig(t, append(basis(t), "SERVICE_ROOT_PATH="+w)...); k.Dienste[0].Root != w {
			t.Errorf("SERVICE_ROOT_PATH=%s: %v", w, k.Dienste)
		}
	}
	for _, w := range []string{"a", "/a/", "//", "/a//b", "/.", "/..", "/a/./b", "/a/../b", "/a b", "/a?x", "/ä", "/a%20b"} {
		erwarteVerstoss(t, "SERVICE_ROOT_PATH", append(basis(t), "SERVICE_ROOT_PATH="+w)...)
	}
}

func TestKonfigURL(t *testing.T) {
	for w, soll := range map[string]string{
		"http://h": "http://h", "https://h:8443/odata": "https://h:8443/odata", "https://h/odata/": "https://h/odata", "https://h/": "https://h", "http://[::1]:80/x": "http://[::1]:80/x",
		"http://h:1": "http://h:1", "https://h:65535/x": "https://h:65535/x",
	} {
		if k := erwarteGueltig(t, append(basis(t), "UPSTREAM_URL="+w)...); k.Dienste[0].UpstreamURL != soll {
			t.Errorf("UPSTREAM_URL=%s: %q", w, k.Dienste[0].UpstreamURL)
		}
	}
	for _, w := range []string{"ftp://h/x", "h/x", "/x", "http://", "http:///x", "http://:80/x", "http://u:p@h/x", "http://u@h/x", "https://h/x?a=1", "https://h/x?", "https://h/x#f", "https://h/x#", "mailto:a@b", "http:h", "https://h x/", "HTTPS-nicht://h",
		"http://h:0", "http://h:65536", "http://h:99999", "http://h:", "http://h:/x", "http://[::1]:", "http://h:99999999999999999999/x"} {
		erwarteVerstoss(t, "UPSTREAM_URL", append(basis(t), "UPSTREAM_URL="+w)...)
	}
}

func TestKonfigModus(t *testing.T) {
	d := "DATA_DIR=" + filepath.Join(t.TempDir(), "m")
	for _, w := range []string{"mock", "hybrid", ""} {
		env := []string{"MODE=" + w, d, "UPSTREAM_URL=https://h"}
		_, zeilen := lade(t, env...)
		hat := false
		for _, z := range zeilen {
			if strings.HasPrefix(z, "ODR-E2001 MODE: ") && strings.Contains(z, "record") && strings.Contains(z, "replay") {
				hat = true
			}
		}
		if !hat {
			t.Errorf("MODE=%q: %v", w, zeilen)
		}
	}
	for _, w := range []string{"record", "replay"} {
		erwarteGueltig(t, "MODE="+w, d, "UPSTREAM_URL=https://h")
	}
	// Die Syntax der übrigen Schlüssel wird trotzdem geprüft und mitgemeldet.
	_, zeilen := lade(t, "MODE=mock", d, "LISTEN=kaputt")
	if s := schluessel(t, zeilen); !slices.Equal(s, []string{"LISTEN", "MODE"}) {
		t.Errorf("Schlüssel %v", s)
	}
	// bei mock und hybrid wird kein DATA_DIR angelegt
	frisch := filepath.Join(t.TempDir(), "frisch")
	lade(t, "MODE=hybrid", "DATA_DIR="+frisch)
	if _, err := os.Stat(frisch); err == nil {
		t.Error("DATA_DIR bei Start-Verstoß angelegt")
	}
}

func schreibe(t *testing.T, inhalt string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "services.json")
	if err := os.WriteFile(p, []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestKonfigServicesFile(t *testing.T) {
	gut := `[{"id":"a","service_root":"/a","upstream_url":"https://a.test/x/"},{"id":"b","service_root":"/b"}] ` + "\n"
	env := []string{"MODE=replay", "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}
	k := erwarteGueltig(t, append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, gut))...)
	if !reflect.DeepEqual(k.Dienste, []bootstrap.Dienst{{ID: "a", Root: "/a", UpstreamURL: "https://a.test/x"}, {ID: "b", Root: "/b"}}) {
		t.Errorf("Dienste %+v", k.Dienste)
	}
	fehlerfaelle := map[string]string{
		"leeres Array":            `[]`,
		"Objekt statt Array":      `{"id":"a","service_root":"/a"}`,
		"Nicht-JSON":              `kein json`,
		"leere Datei":             ``,
		"Anhang":                  `[{"id":"a","service_root":"/a"}] x`,
		"zweites Array":           `[{"id":"a","service_root":"/a"}][]`,
		"Eintrag kein Objekt":     `["a"]`,
		"Pflicht id fehlt":        `[{"service_root":"/a"}]`,
		"Pflicht root fehlt":      `[{"id":"a"}]`,
		"unbekanntes Feld":        `[{"id":"a","service_root":"/a","extra":1}]`,
		"Feld falsch geschrieben": `[{"ID":"a","service_root":"/a"}]`,
		"falscher Typ":            `[{"id":1,"service_root":"/a"}]`,
		"falscher Typ root":       `[{"id":"a","service_root":["/a"]}]`,
		"ungültige id":            `[{"id":"a b","service_root":"/a"}]`,
		"ungültiger Root":         `[{"id":"a","service_root":"a"}]`,
		"ungültige URL":           `[{"id":"a","service_root":"/a","upstream_url":"ftp://h"}]`,
		"doppelte id":             `[{"id":"a","service_root":"/a"},{"id":"a","service_root":"/b"}]`,
		"gleicher Root":           `[{"id":"a","service_root":"/a"},{"id":"b","service_root":"/a"}]`,
		"Überlappung":             `[{"id":"a","service_root":"/"},{"id":"b","service_root":"/a"}]`,
	}
	for name, inhalt := range fehlerfaelle {
		t.Run(name, func(t *testing.T) {
			_, zeilen := lade(t, append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, inhalt))...)
			if len(zeilen) == 0 {
				t.Fatal("kein Verstoß")
			}
			for _, s := range schluessel(t, zeilen) {
				if !strings.HasPrefix(s, "SERVICES_FILE") {
					t.Errorf("fremder Schlüssel %s", s)
				}
			}
		})
	}
	t.Run("Datei fehlt", func(t *testing.T) {
		erwarteVerstoss(t, "SERVICES_FILE", append(slices.Clone(env), "SERVICES_FILE="+filepath.Join(t.TempDir(), "gibt-es-nicht"))...)
	})
	t.Run("Verzeichnis", func(t *testing.T) {
		erwarteVerstoss(t, "SERVICES_FILE", append(slices.Clone(env), "SERVICES_FILE="+t.TempDir())...)
	})
	t.Run("über 1 MiB", func(t *testing.T) {
		gross := `[{"id":"a","service_root":"/a"}` + strings.Repeat(" ", 1<<20) + `]`
		erwarteVerstoss(t, "SERVICES_FILE", append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, gross))...)
	})
	t.Run("genau 1 MiB", func(t *testing.T) {
		kopf := `[{"id":"a","service_root":"/a"}`
		genau := kopf + strings.Repeat(" ", (1<<20)-len(kopf)-1) + `]`
		erwarteGueltig(t, append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, genau))...)
	})
	t.Run("kein UTF-8", func(t *testing.T) {
		erwarteVerstoss(t, "SERVICES_FILE", append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, "[{\"id\":\"a\xff\",\"service_root\":\"/a\"}]"))...)
	})
	t.Run("Schlüssel der Einträge", func(t *testing.T) {
		_, zeilen := lade(t, append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, `[{"id":"a","service_root":"/a"},{"id":"a b","service_root":"x"}]`))...)
		if s := schluessel(t, zeilen); !slices.Equal(s, []string{"SERVICES_FILE[1].id", "SERVICES_FILE[1].service_root"}) {
			t.Errorf("Schlüssel %v", s)
		}
	})
	t.Run("Doppelkonfiguration", func(t *testing.T) {
		datei := "SERVICES_FILE=" + schreibe(t, gut)
		for _, einzel := range []string{"SERVICE_ID=a", "SERVICE_ROOT_PATH=/a", "UPSTREAM_URL=https://h"} {
			key, _, _ := strings.Cut(einzel, "=")
			erwarteVerstoss(t, key, append(slices.Clone(env), datei, einzel)...)
		}
	})
	dienstenv := func(inhalt string) []string {
		return append(slices.Clone(env), "SERVICES_FILE="+schreibe(t, inhalt))
	}
	t.Run("Feld doppelt: der letzte Wert gilt", func(t *testing.T) {
		k := erwarteGueltig(t, dienstenv(`[{"id":"a","id":"b","service_root":"/x","service_root":"/a"}]`)...)
		if !reflect.DeepEqual(k.Dienste, []bootstrap.Dienst{{ID: "b", Root: "/a"}}) {
			t.Errorf("Dienste %+v", k.Dienste)
		}
		erwarteGrund(t, "SERVICES_FILE[0].id", "Pflichtfeld fehlt", dienstenv(`[{"id":"a","id":null,"service_root":"/a"}]`)...)
		erwarteGueltig(t, dienstenv(`[{"id":null,"id":"a","service_root":"/a"}]`)...)
	})
	t.Run("null, leer und Leerraum sind fehlend", func(t *testing.T) {
		for _, w := range []string{`null`, `""`, `" "`, `"\t \n"`} {
			erwarteGrund(t, "SERVICES_FILE[0].id", "Pflichtfeld fehlt", dienstenv(`[{"id":`+w+`,"service_root":"/a"}]`)...)
			erwarteGrund(t, "SERVICES_FILE[0].service_root", "Pflichtfeld fehlt", dienstenv(`[{"id":"a","service_root":`+w+`}]`)...)
		}
	})
	t.Run("upstream_url null, leer und Leerraum", func(t *testing.T) {
		for _, w := range []string{`null`, `""`, `" "`} {
			inhalt := `[{"id":"a","service_root":"/a","upstream_url":` + w + `}]`
			erwarteGrund(t, "SERVICES_FILE[0].upstream_url", "im Modus record erforderlich",
				"MODE=record", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, inhalt))
			erwarteGueltig(t, dienstenv(inhalt)...) // replay: kein Verstoß
		}
	})
	t.Run("Leerraum am Rand wird entfernt", func(t *testing.T) {
		k := erwarteGueltig(t, dienstenv(`[{"id":" x ","service_root":" /a ","upstream_url":" https://h/ "}]`)...)
		if !reflect.DeepEqual(k.Dienste, []bootstrap.Dienst{{ID: "x", Root: "/a", UpstreamURL: "https://h"}}) {
			t.Errorf("Dienste %+v", k.Dienste)
		}
	})
	t.Run("Element null ist kein Objekt", func(t *testing.T) {
		erwarteGrund(t, "SERVICES_FILE[0]", "kein Objekt", dienstenv(`[null]`)...)
		erwarteGrund(t, "SERVICES_FILE[1]", "kein Objekt", dienstenv(`[{"id":"a","service_root":"/a"},null]`)...)
	})
	t.Run("Gründe", func(t *testing.T) {
		erwarteGrund(t, "SERVICES_FILE[0].id", "Pflichtfeld fehlt", dienstenv(`[{"service_root":"/a"}]`)...)
		erwarteGrund(t, "SERVICES_FILE[0]", "unbekanntes Feld", dienstenv(`[{"id":"a","service_root":"/a","extra":1}]`)...)
		erwarteGrund(t, "SERVICES_FILE[0].id", "muss eine Zeichenkette sein", dienstenv(`[{"id":1,"service_root":"/a"}]`)...)
		erwarteGrund(t, "SERVICES_FILE[1].id", "doppelte id", dienstenv(`[{"id":"a","service_root":"/a"},{"id":"a","service_root":"/b"}]`)...)
		erwarteGrund(t, "SERVICES_FILE", "größer als 1 MiB", dienstenv(`[{"id":"a","service_root":"/a"}`+strings.Repeat(" ", 1<<20)+`]`)...)
	})
	t.Run("upstream_url Pflicht in record", func(t *testing.T) {
		erwarteVerstoss(t, "SERVICES_FILE[1].upstream_url", "MODE=record", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, gut))
	})
}

func TestKonfigRootUeberlappung(t *testing.T) {
	tests := []struct {
		a, b string
		soll bool
	}{
		{"/", "/a", true}, {"/a", "/", true}, {"/a", "/ab", false}, {"/ab", "/a", false}, {"/a", "/a/b", true},
		{"/a/b", "/a", true}, {"/a", "/a", true}, {"/", "/", true}, {"/a/b", "/a/c", false}, {"/a", "/b", false},
	}
	for _, tc := range tests {
		inhalt := `[{"id":"x","service_root":"` + tc.a + `"},{"id":"y","service_root":"` + tc.b + `"}]`
		_, zeilen := lade(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, inhalt))
		if (len(zeilen) > 0) != tc.soll {
			t.Errorf("%s neben %s: Verstöße %v, erwartet Überlappung %v", tc.a, tc.b, zeilen, tc.soll)
		}
	}
}

func TestKonfigRecordOhneUpstream(t *testing.T) {
	erwarteVerstoss(t, "UPSTREAM_URL", "MODE=record", "DATA_DIR="+filepath.Join(t.TempDir(), "d"))
	erwarteVerstoss(t, "UPSTREAM_URL", "MODE=record", "UPSTREAM_URL=  ", "DATA_DIR="+filepath.Join(t.TempDir(), "d"))
	erwarteGueltig(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "d"))
	erwarteGueltig(t, "MODE=replay", "UPSTREAM_URL=https://h", "DATA_DIR="+filepath.Join(t.TempDir(), "d"))
}

func TestKonfigTLSWiderspruch(t *testing.T) {
	tests := []struct {
		verify, skip string
		fehler       bool
		pruefen      bool
	}{
		{"false", "", true, false}, {"false", "false", true, false}, {"true", "true", true, false},
		{"false", "true", false, false}, {"", "true", false, false}, {"", "", false, true}, {"true", "", false, true}, {"true", "false", false, true},
	}
	for _, tc := range tests {
		env := basis(t)
		if tc.verify != "" {
			env = append(env, "UPSTREAM_TLS_VERIFY="+tc.verify)
		}
		if tc.skip != "" {
			env = append(env, "UPSTREAM_TLS_INSECURE_SKIP_VERIFY="+tc.skip)
		}
		k, zeilen := lade(t, env...)
		if (len(zeilen) > 0) != tc.fehler {
			t.Errorf("verify=%q skip=%q: %v", tc.verify, tc.skip, zeilen)
			continue
		}
		if !tc.fehler && k.UpstreamTLSVerify != tc.pruefen {
			t.Errorf("verify=%q skip=%q: UpstreamTLSVerify=%v", tc.verify, tc.skip, k.UpstreamTLSVerify)
		}
	}
}

func TestKonfigDatenverzeichnis(t *testing.T) {
	wurzel := t.TempDir()
	dir := filepath.Join(wurzel, "a", "b")
	erwarteGueltig(t, "MODE=replay", "DATA_DIR="+dir)
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() || fi.Mode().Perm()&0o077 != 0 {
		t.Fatalf("Verzeichnis %v %v", fi, err)
	}
	erwarteGueltig(t, "MODE=replay", "DATA_DIR="+dir) // idempotent
	// Verstoß anderswo: nichts wird angelegt
	neu := filepath.Join(wurzel, "nicht", "anlegen")
	erwarteVerstoss(t, "LISTEN", "MODE=replay", "LISTEN=kaputt", "DATA_DIR="+neu)
	if _, err := os.Stat(filepath.Join(wurzel, "nicht")); err == nil {
		t.Error("DATA_DIR trotz Verstoß angelegt")
	}
	// Nicht-Ordner
	datei := filepath.Join(wurzel, "datei")
	if err := os.WriteFile(datei, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	erwarteVerstoss(t, "DATA_DIR", "MODE=replay", "DATA_DIR="+datei)
	erwarteVerstoss(t, "DATA_DIR", "MODE=replay", "DATA_DIR="+filepath.Join(datei, "unter"))
	erwarteVerstoss(t, "DATA_DIR", "MODE=replay", "DATA_DIR=a\x00b")
}

func zertifikat(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestKonfigCADatei(t *testing.T) {
	dir := t.TempDir()
	gut := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(gut, zertifikat(t), 0o600); err != nil {
		t.Fatal(err)
	}
	ohne := filepath.Join(dir, "leer.pem")
	if err := os.WriteFile(ohne, []byte("kein Zertifikat"), 0o600); err != nil {
		t.Fatal(err)
	}
	if k := erwarteGueltig(t, append(basis(t), "UPSTREAM_CA_FILE="+gut)...); k.UpstreamCAFile != gut {
		t.Errorf("UpstreamCAFile %q", k.UpstreamCAFile)
	}
	for name, pfad := range map[string]string{"ohne Zertifikat": ohne, "fehlt": filepath.Join(dir, "fehlt.pem"), "Verzeichnis": dir} {
		t.Run(name, func(t *testing.T) {
			erwarteVerstoss(t, "UPSTREAM_CA_FILE", append(basis(t), "UPSTREAM_CA_FILE="+pfad)...)
		})
	}
	// Verstoß an der Datei verhindert auch hier das Anlegen von DATA_DIR
	d := filepath.Join(dir, "daten")
	erwarteVerstoss(t, "UPSTREAM_CA_FILE", "MODE=replay", "UPSTREAM_CA_FILE="+ohne, "DATA_DIR="+d)
	if _, err := os.Stat(d); err == nil {
		t.Error("DATA_DIR trotz Verstoß an der Datei angelegt")
	}
}

func TestKonfigMeldung(t *testing.T) {
	const geheim = "geheim-123"
	env := []string{
		"MODE=" + geheim, "LISTEN=" + geheim, "UPSTREAM_TIMEOUT=" + geheim, "MAX_CAPTURE_BODY=" + geheim, "SERVICE_ID=" + geheim + "/x",
		"REDACT_JSON_PATHS=" + geheim, "UPSTREAM_URL=ftp://" + geheim, "ADMIN_TOKEN_FILE=" + geheim, "REPLAY_STRATEGY=" + geheim,
		"DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "SERVICES_FILE=" + filepath.Join(t.TempDir(), geheim),
	}
	var stderr bytes.Buffer
	_, code := bootstrap.LadeOderMelde(env, &stderr)
	if code != 1 {
		t.Errorf("Exit-Code %d", code)
	}
	if strings.Contains(stderr.String(), geheim) {
		t.Errorf("die Meldung nennt den Wert:\n%s", stderr.String())
	}
	zeilen := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
	if len(zeilen) < 8 {
		t.Fatalf("zu wenige Zeilen (alle Verstöße, nicht nur der erste): %v", zeilen)
	}
	if s := schluessel(t, zeilen); !slices.IsSorted(s) {
		t.Errorf("Zeilen nicht nach Schlüssel sortiert: %v", zeilen)
	}
	for _, z := range zeilen {
		if !strings.HasPrefix(z, "ODR-E2001 ") {
			t.Errorf("Zeile ohne Code: %q", z)
		}
	}
	if _, code := bootstrap.LadeOderMelde(basis(t), &stderr); code != 0 {
		t.Errorf("gültig: Exit-Code %d", code)
	}
	if _, code := bootstrap.LadeOderMelde([]string{"MODE=kaputt"}, nil); code != 1 {
		t.Errorf("nil-Writer: Exit-Code %d", code)
	}
	var kf *bootstrap.KonfigFehler
	_, err := bootstrap.LadeKonfiguration(env)
	if !errors.As(err, &kf) || kf.Code() != model.CodeE2001 {
		t.Errorf("Fehler %v", err)
	}
}

func TestKonfigWiederholung(t *testing.T) {
	env := append(basis(t), "REDACT_HEADERS=b,a", "MATCH_VARY_HEADERS=z,y")
	a, b := erwarteGueltig(t, env...), erwarteGueltig(t, env...)
	if !reflect.DeepEqual(a, b) {
		t.Error("zweimal geladen, Werte verschieden")
	}
	bad := []string{"MODE=mock", "LISTEN=x", "UPSTREAM_TIMEOUT=y", "DATA_DIR=" + filepath.Join(t.TempDir(), "d")}
	_, z1 := lade(t, bad...)
	_, z2 := lade(t, bad...)
	if !reflect.DeepEqual(z1, z2) || len(z1) == 0 {
		t.Errorf("Meldungen verschieden: %v / %v", z1, z2)
	}
}

// TestKonfigModusUngueltig: ein ungültiger Wert wird einmal gemeldet, nicht zusätzlich als „Modus nicht unterstützt“.
func TestKonfigModusUngueltig(t *testing.T) {
	_, zeilen := lade(t, "MODE=Record", "DATA_DIR="+filepath.Join(t.TempDir(), "d"))
	if s := schluessel(t, zeilen); !slices.Equal(s, []string{"MODE"}) {
		t.Errorf("Zeilen %v", zeilen)
	}
}

// TestKonfigUpstreamModus: die Pflicht von UPSTREAM_URL hängt am gültigen Modus, die Form wird in jedem Modus geprüft (SPEC-110 Nr. 9).
func TestKonfigUpstreamModus(t *testing.T) {
	d := func() string { return "DATA_DIR=" + filepath.Join(t.TempDir(), "d") }
	erwarteGueltig(t, "MODE=replay", d())
	erwarteGrund(t, "UPSTREAM_URL", "keine absolute http- oder https-URL", "MODE=replay", "UPSTREAM_URL=ftp://x", d())
	erwarteGrund(t, "UPSTREAM_URL", "im Modus record erforderlich", "MODE=record", d())
	_, zeilen := lade(t, "MODE=Record", d())
	if len(zeilen) != 1 || !strings.HasPrefix(zeilen[0], "ODR-E2001 MODE: ") {
		t.Errorf("MODE=Record ohne URL: %v", zeilen)
	}
	_, zeilen = lade(t, "MODE=Record", "UPSTREAM_URL=ftp://x", d())
	if s := schluessel(t, zeilen); !slices.Equal(s, []string{"MODE", "UPSTREAM_URL"}) {
		t.Errorf("MODE=Record mit ftp: %v", zeilen)
	}
}

// TestKonfigMeldungSortierung: die Meldung folgt der Bytefolge des Schlüssels, dann dem Grund; gleiche Zeilen stehen einmal (SPEC-110 Nr. 12).
func TestKonfigMeldungSortierung(t *testing.T) {
	// Die Folge richtet sich nach dem Schlüssel allein: SERVICES_FILE[0] geht SERVICES_FILE[0].id voraus.
	datei := schreibe(t, `[{"extra":1,"service_root":"/a"}]`)
	_, zeilen := lade(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+datei)
	if s := schluessel(t, zeilen); !slices.Equal(s, []string{"SERVICES_FILE[0]", "SERVICES_FILE[0].id"}) {
		t.Errorf("Reihenfolge %v", zeilen)
	}
	// Index als Text: [10] vor [2]
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < 12; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"service_root":"/r` + strconv.Itoa(i) + `"}`)
	}
	b.WriteString("]")
	_, zeilen = lade(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, b.String()))
	soll := schluessel(t, zeilen)
	if len(soll) != 12 || slices.Index(soll, "SERVICES_FILE[10].id") > slices.Index(soll, "SERVICES_FILE[2].id") || !slices.IsSorted(soll) {
		t.Errorf("Index als Text: %v", soll)
	}
	// bei gleichem Schlüssel nach Grund, gleiche Zeilen einmal: ein Eintrag mit zwei unbekannten Feldern
	_, zeilen = lade(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, `[{"id":"a","service_root":"/a","x":1,"y":2}]`))
	if len(zeilen) != 1 {
		t.Errorf("gleiche Zeilen nicht zusammengefasst: %v", zeilen)
	}
	// dritte gleiche id: „doppelte id“ für [2] genau einmal, obwohl [2] mit [0] und [1] kollidiert
	_, zeilen = lade(t, "MODE=replay", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, `[{"id":"a","service_root":"/a"},{"id":"a","service_root":"/b"},{"id":"a","service_root":"/c"}]`))
	soll = []string{"ODR-E2001 SERVICES_FILE[1].id: doppelte id", "ODR-E2001 SERVICES_FILE[2].id: doppelte id"}
	if !slices.Equal(zeilen, soll) {
		t.Errorf("dritte gleiche id %v, erwartet %v", zeilen, soll)
	}
}

// TestKonfigDienstefeldTyp: ein Feld falschen Typs meldet nur den Typ, weder „Pflichtfeld fehlt“ noch die Form noch „im Modus record erforderlich“ (SPEC-110 Nr. 9, Nr. 12).
// Geprüft sind id, service_root und upstream_url in record sowie upstream_url in replay und mock; hybrid ist nicht geprüft.
func TestKonfigDienstefeldTyp(t *testing.T) {
	for name, tc := range map[string]struct{ inhalt, zeile string }{
		"id":           {`[{"id":1,"service_root":"/a","upstream_url":"https://h"}]`, "ODR-E2001 SERVICES_FILE[0].id: muss eine Zeichenkette sein"},
		"service_root": {`[{"id":"a","service_root":1,"upstream_url":"https://h"}]`, "ODR-E2001 SERVICES_FILE[0].service_root: muss eine Zeichenkette sein"},
		"upstream_url": {`[{"id":"a","service_root":"/a","upstream_url":1}]`, "ODR-E2001 SERVICES_FILE[0].upstream_url: muss eine Zeichenkette sein"},
	} {
		t.Run(name, func(t *testing.T) {
			_, zeilen := lade(t, "MODE=record", "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, tc.inhalt))
			if !slices.Equal(zeilen, []string{tc.zeile}) {
				t.Errorf("Zeilen %v, erwartet [%s]", zeilen, tc.zeile)
			}
		})
	}
	const eintrag = `[{"id":"a","service_root":"/a","upstream_url":1}]`
	const typ = "ODR-E2001 SERVICES_FILE[0].upstream_url: muss eine Zeichenkette sein"
	for modus, soll := range map[string][]string{
		"replay": {typ},
		"mock":   {"ODR-E2001 MODE: Modus nicht unterstützt; lieferbar sind record und replay", typ},
	} {
		t.Run("upstream_url_"+modus, func(t *testing.T) {
			_, zeilen := lade(t, "MODE="+modus, "DATA_DIR="+filepath.Join(t.TempDir(), "d"), "SERVICES_FILE="+schreibe(t, eintrag))
			if !slices.Equal(zeilen, soll) {
				t.Errorf("Zeilen %v, erwartet %v", zeilen, soll)
			}
		})
	}
}

// TestKonfigTLSUngueltig: ein formal ungültiger Wahrheitswert meldet nur seine Form; der Widerspruch wird nur bei zwei gültigen Werten geprüft (SPEC-110 Nr. 10, Nr. 12).
func TestKonfigTLSUngueltig(t *testing.T) {
	const v = "ODR-E2001 UPSTREAM_TLS_VERIFY: Wahrheitswert muss true oder false sein"
	const s = "ODR-E2001 UPSTREAM_TLS_INSECURE_SKIP_VERIFY: Wahrheitswert muss true oder false sein"
	tests := []struct {
		verify, skip string
		soll         []string
	}{
		{"yes", "", []string{v}}, {"yes", "true", []string{v}}, {"1", "", []string{v}}, {"1", "true", []string{v}},
		{"false", "yes", []string{s}}, {"true", "yes", []string{s}}, {"", "TRUE", []string{s}},
		{"yes", "yes", []string{s, v}},
	}
	for _, tc := range tests {
		env := basis(t)
		if tc.verify != "" {
			env = append(env, "UPSTREAM_TLS_VERIFY="+tc.verify)
		}
		if tc.skip != "" {
			env = append(env, "UPSTREAM_TLS_INSECURE_SKIP_VERIFY="+tc.skip)
		}
		_, zeilen := lade(t, env...)
		if !slices.Equal(zeilen, tc.soll) {
			t.Errorf("verify=%q skip=%q: Zeilen %v, erwartet %v", tc.verify, tc.skip, zeilen, tc.soll)
		}
		for _, z := range zeilen {
			if strings.Contains(z, "braucht") || strings.Contains(z, "widerspricht") {
				t.Errorf("verify=%q skip=%q: Widerspruch gemeldet: %s", tc.verify, tc.skip, z)
			}
		}
	}
}

// TestKonfigJSONPfadWieModell: der Start prüft REDACT_JSON_PATHS mit dem Parser des Modells;
// jede Form, die der Parser ablehnt, lehnt auch LadeKonfiguration ab, und umgekehrt.
func TestKonfigJSONPfadWieModell(t *testing.T) {
	gueltig := []string{"$..password", "$..api_key", "$..A-b_1"}
	ungueltig := []string{"$", "$..", "$.a", "$...a", "..a", "$..a.b", "$..a b", "$..*", "$..[0]", "$..a[0]", "$..ä", "$..a;b"}
	for _, w := range gueltig {
		if _, err := model.ParseJSONPfad(w); err != nil {
			t.Fatalf("Modell lehnt %q ab: %v", w, err)
		}
		erwarteGueltig(t, append(basis(t), "REDACT_JSON_PATHS="+w)...)
	}
	for _, w := range ungueltig {
		if _, err := model.ParseJSONPfad(w); err == nil {
			t.Fatalf("Modell nimmt %q an", w)
		}
		erwarteVerstoss(t, "REDACT_JSON_PATHS", append(basis(t), "REDACT_JSON_PATHS="+w)...)
	}
}
