package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// adminHoch baut und fährt einen Handle mit Admin-Listener auf echten Adaptern hoch.
func adminHoch(t *testing.T, k bootstrap.Konfiguration) *bootstrap.Handle {
	t.Helper()
	h, _ := adminBau(t, k, nil, nil)
	if err := h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("%s: %v", k.Modus, err)
	}
	return h
}

// TestAdminHealthz (B14): `GET /healthz` ohne Token ist 200, drei andere Formen 401, in 4 Modi am Listener des Handle.
func TestAdminHealthz(t *testing.T) {
	zellen := 0
	for _, modus := range vierModi() {
		url := "http://" + adminHoch(t, listenerKonfig(t.TempDir(), modus)).AdminAdresse()
		for _, f := range []struct {
			methode, pfad string
			status        int
		}{{"GET", "/healthz", 200}, {"GET", "/healthz/", 401}, {"GET", "/HEALTHZ", 401}, {"POST", "/healthz", 401}} {
			zellen++
			status, body, kopf := sende(neuerAdminClient(), f.methode, url+f.pfad, "", "")
			if status != f.status || kopf.Get("X-Correlation-Id") == "" || (status == 200 && body != `{"status":"ok"}`) {
				t.Errorf("%s %s %s: %d %q %q", modus, f.methode, f.pfad, status, body, kopf.Get("X-Correlation-Id"))
			}
		}
	}
	if zellen != 16 {
		t.Errorf("%d Zellen, erwartet 16", zellen)
	}
}

// auditZeile ist eine Zeile von audit.jsonl: die Schlüssel in ihrer Folge und die Werte.
func auditZeile(t *testing.T, zeile string) ([]string, map[string]string) {
	t.Helper()
	d := json.NewDecoder(strings.NewReader(zeile))
	var schluessel []string
	werte := map[string]string{}
	if _, err := d.Token(); err != nil {
		t.Fatalf("Zeile %q: %v", zeile, err)
	}
	for d.More() {
		k, err := d.Token()
		if err != nil {
			t.Fatalf("Zeile %q: %v", zeile, err)
		}
		var v string
		if err := d.Decode(&v); err != nil {
			t.Fatalf("Zeile %q: %v", zeile, err)
		}
		schluessel = append(schluessel, k.(string))
		werte[k.(string)] = v
	}
	return schluessel, werte
}

// TestAdminAbweisung (B15): 4 Anlässe × 4 Modi: je Abweisung genau eine Zeile `admin_abgewiesen` mit vier Schlüsseln, die
// Korrelations-ID der Antwort und der Zähler ODR-E2101 der Diagnose um eins.
func TestAdminAbweisung(t *testing.T) {
	zellen := 0
	for _, modus := range vierModi() {
		k := listenerKonfig(t.TempDir(), modus)
		h := adminHoch(t, k)
		url := "http://" + h.AdminAdresse() + "/__admin/sessions"
		auth := bearer(t, k.AdminTokenFile)
		anlaesse := [][]string{nil, {"Bearer " + strings.Repeat("0", 64)}, {"Basic " + strings.TrimPrefix(auth, "Bearer ")}, {auth, auth}}
		for i, kopf := range anlaesse {
			zellen++
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header["Authorization"] = kopf
			resp, err := neuerAdminClient().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			inhalt, err := os.ReadFile(filepath.Join(k.DataDir, "audit.jsonl"))
			zeilen := strings.Split(strings.TrimSuffix(string(inhalt), "\n"), "\n")
			if err != nil || resp.StatusCode != 401 || len(zeilen) != i+1 {
				t.Fatalf("%s %d: Status %d, %d Zeilen (%v)", modus, i, resp.StatusCode, len(zeilen), err)
			}
			schluessel, werte := auditZeile(t, zeilen[i])
			if !slices.Equal(schluessel, []string{"timestamp", "correlation_id", "art", "error_code"}) || werte["art"] != "admin_abgewiesen" ||
				werte["error_code"] != "ODR-E2101" || werte["correlation_id"] != resp.Header.Get("X-Correlation-Id") || werte["timestamp"] == "" {
				t.Errorf("%s %d: %s, Korrelations-ID %q", modus, i, zeilen[i], resp.Header.Get("X-Correlation-Id"))
			}
			d, err := h.Inspektion().Diagnose(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			for _, z := range d.Fehler {
				if z.Code == model.CodeE2101 && z.Anzahl != uint64(i+1) {
					t.Errorf("%s %d: Zähler %d, erwartet %d", modus, i, z.Anzahl, i+1)
				}
			}
		}
	}
	if zellen != 16 {
		t.Errorf("%d Zellen, erwartet 16", zellen)
	}
}

// artefakte sind die Wege, auf denen ein Geheimnis den Prozess verlassen könnte (SPEC-300 Nr. 12 a).
type artefakte struct {
	daten          string
	stderr, stdout *syncPuffer
	antworten      *bytes.Buffer
}

// lies liefert je Weg den Inhalt: das Audit-Log, stderr, stdout, die übrigen Dateien außer der Token-Datei und die Antworten.
func (a artefakte) lies(t *testing.T) map[string]string {
	t.Helper()
	audit, _ := os.ReadFile(filepath.Join(a.daten, "audit.jsonl"))
	var andere strings.Builder
	err := filepath.WalkDir(a.daten, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == "audit.jsonl" || d.Name() == adminTokenStandard {
			return err
		}
		b, err := os.ReadFile(p)
		andere.Write(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{
		"audit.jsonl": string(audit), "stderr": a.stderr.String(), "stdout": a.stdout.String(),
		"andere Dateien": andere.String(), "Antworten": a.antworten.String(),
	}
}

func traegt(inhalt, token string) bool {
	return strings.Contains(inhalt, token[:len(token)/2]) || strings.Contains(inhalt, token[len(token)/2:])
}

// TestAdminListenerKeinGeheimnis (B16): 5 Wege × 6 Anlässe × 4 Modi = 120 Zellen über Start, dazu je Weg die Kontrolle, dass die
// Suche einen eingespielten Wert findet.
func TestAdminListenerKeinGeheimnis(t *testing.T) {
	zellen := 0
	for m, modus := range vierModi() {
		dir := filepath.Join(t.TempDir(), "d")
		a := artefakte{daten: dir, stderr: &syncPuffer{}, stdout: &syncPuffer{}, antworten: &bytes.Buffer{}}
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		alt := os.Stdout
		os.Stdout = w
		kopiert := make(chan struct{})
		go func() { _, _ = io.Copy(a.stdout, r); close(kopiert) }()
		adresse := make(chan string, 2)
		b := bootstrap.StandardBauteile()
		echt := b.Listen
		b.Listen = func(ctx context.Context, network, _ string) (net.Listener, error) {
			l, err := echt(ctx, network, "127.0.0.1:0")
			if err == nil {
				adresse <- l.Addr().String()
			}
			return l, err
		}
		ctx, cancel := context.WithCancel(context.Background())
		ergebnis := make(chan int, 1)
		go func() { ergebnis <- bootstrap.StartMit(ctx, adminEnv(modus, dir, "PUBLIC_BASE_URL="), a.stderr, b) }()
		<-adresse
		url := "http://" + <-adresse
		wartetBis(func() bool { return strings.Contains(a.stderr.String(), "start: bereit") })
		token := strings.TrimPrefix(bearer(t, filepath.Join(dir, adminTokenStandard)), "Bearer ")
		letztes := token[:len(token)-1] + "x"
		if token[len(token)-1] == 'x' {
			letztes = token[:len(token)-1] + "y"
		}
		for _, anlass := range []struct{ methode, auth string }{
			{"GET", ""}, {"GET", "Bearer " + strings.Repeat("f", 64)}, {"GET", "Bearer " + letztes}, {"GET", "Bearer " + token[:32]},
			{"GET", "Bearer " + token + "0"}, {"POST", "Bearer " + token},
		} {
			status, body, kopf := sende(neuerAdminClient(), anlass.methode, url+"/__admin/sessions", anlass.auth, `{"service_id":"default"}`)
			if soll := map[bool]int{false: 401, true: 201}[anlass.methode == "POST"]; status != soll {
				t.Errorf("%s %q: %d %s", modus, anlass.methode, status, body)
			}
			fmt.Fprintf(a.antworten, "%s %v\n", body, kopf)
			for weg, inhalt := range a.lies(t) {
				zellen++
				if traegt(inhalt, token) {
					t.Errorf("%s, %s %s: das Token steht in %s", modus, anlass.methode, anlass.auth[:min(len(anlass.auth), 12)], weg)
				}
			}
		}
		cancel()
		select {
		case code := <-ergebnis:
			if code != 0 {
				t.Errorf("%s: Start %d: %s", modus, code, a.stderr.String())
			}
		case <-time.After(20 * time.Second):
			t.Fatal("Start kehrt nicht zurück")
		}
		const kontrolle = "KONTROLLE-4f1d9a7c-KONTROLLE-4f1d9a7c"
		if m == 0 {
			if err := os.WriteFile(filepath.Join(dir, "audit.jsonl"), []byte(kontrolle), 0o600); err != nil {
				t.Fatal(err)
			}
			schreibeDatei(t, filepath.Join(dir, "andere"), kontrolle, 0o600)
			_, _ = a.stderr.Write([]byte(kontrolle))
			_, _ = fmt.Fprint(os.Stdout, kontrolle)
			a.antworten.WriteString(kontrolle)
		}
		os.Stdout = alt
		_ = w.Close()
		<-kopiert
		for weg, inhalt := range a.lies(t) {
			if m == 0 && !traegt(inhalt, kontrolle) {
				t.Errorf("Kontrolle %s: die Suche findet den eingespielten Wert nicht", weg)
			}
			if traegt(inhalt, token) {
				t.Errorf("%s nach dem Herunterfahren: das Token steht in %s", modus, weg)
			}
		}
		_ = r.Close()
	}
	if zellen != 120 {
		t.Errorf("%d Zellen, erwartet 120", zellen)
	}
}
