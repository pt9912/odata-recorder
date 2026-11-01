package bootstrap_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestHerunterfahrenOhneAbbruchDesAusloesers fängt S1 am Durchstich: echte Adapter im Modus record,
// ein Exchange hängt am Upstream, ctx wird abgebrochen. Zwischen Anhalten und CloseAll erhält ein Request
// auf einer bestehenden Verbindung 503 ODR-E5010; der Exchange läuft zu Ende, wird angehängt, die Session
// ist complete und Start liefert 0.
func TestHerunterfahrenOhneAbbruchDesAusloesers(t *testing.T) {
	angekommen := make(chan struct{}, 4)
	frei := make(chan struct{})
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/Hold") {
			angekommen <- struct{}{}
			<-frei
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	defer up.Close()
	gegeben := false
	gibFrei := func() {
		if !gegeben {
			gegeben = true
			close(frei)
		}
	}
	defer gibFrei()

	dir := t.TempDir()
	env := []string{
		"MODE=record", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + filepath.Join(dir, "d"), "SERVICE_ID=a",
		"SERVICE_ROOT_PATH=/odata", "UPSTREAM_URL=" + up.URL + "/odata", "SHUTDOWN_TIMEOUT=10s",
	}
	b := bootstrap.StandardBauteile()
	echt := b.Listen
	adresse := make(chan string, 1)
	b.Listen = func(ctx context.Context, network, _ string) (net.Listener, error) {
		l, err := echt(ctx, network, "127.0.0.1:0")
		if err == nil {
			adresse <- l.Addr().String()
		}
		return l, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stderr syncPuffer
	code := make(chan int, 1)
	go func() { code <- bootstrap.StartMit(ctx, env, &stderr, b) }()
	var addr string
	select {
	case addr = <-adresse:
	case <-time.After(10 * time.Second):
		t.Fatalf("kein Listener; stderr:\n%s", stderr.String())
	}
	holen := func(c *http.Client, pfad string) (int, string, error) {
		resp, err := c.Get("http://" + addr + pfad)
		if err != nil {
			return 0, "", err
		}
		defer func() { _ = resp.Body.Close() }()
		b, err := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b), err
	}
	zweiter := &http.Client{Transport: &http.Transport{}}
	if status, _, err := holen(zweiter, "/odata/Ping"); err != nil || status != 200 {
		t.Fatalf("Ping: %d, %v", status, err)
	}
	type antwort struct {
		status int
		body   string
		err    error
	}
	erster := make(chan antwort, 1)
	go func() {
		status, body, err := holen(&http.Client{Transport: &http.Transport{}}, "/odata/Hold")
		erster <- antwort{status, body, err}
	}()
	select {
	case <-angekommen:
	case <-time.After(10 * time.Second):
		t.Fatal("der Exchange erreicht den Upstream nicht")
	}
	cancel()

	// Auf der bestehenden Verbindung antwortet der Server, sobald er angehalten hat, mit 503 ODR-E5010.
	var status int
	var body string
	for ende := time.Now().Add(10 * time.Second); time.Now().Before(ende); time.Sleep(5 * time.Millisecond) {
		s, bd, err := holen(zweiter, "/odata/Ping")
		if err != nil {
			t.Fatalf("zweiter Request: %v", err)
		}
		if s != 200 {
			status, body = s, bd
			break
		}
	}
	if status != 503 || !strings.Contains(body, "ODR-E5010") {
		t.Errorf("zweiter Request: %d %q, erwartet 503 ODR-E5010", status, body)
	}

	gibFrei()
	select {
	case a := <-erster:
		if a.err != nil || a.status != 200 || a.body != `{"value":[]}` {
			t.Errorf("erster Request: %+v", a)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("der erste Request endet nicht")
	}
	select {
	case c := <-code:
		if c != 0 {
			t.Errorf("Rückgabe %d, stderr:\n%s", c, stderr.String())
		}
	case <-time.After(20 * time.Second):
		t.Fatal("start kehrt nicht zurück")
	}

	ctx2 := context.Background()
	db, err := sqlite.Open(ctx2, filepath.Join(dir, "d", dateiDB), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	sessions, err := db.Sessions().List(ctx2, "a")
	if err != nil || len(sessions) != 1 || sessions[0].Status != model.SessionComplete {
		t.Fatalf("Sessions %+v, %v", sessions, err)
	}
	exchanges, err := db.Recording().ListBySession(ctx2, sessions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	gehalten := 0
	for _, e := range exchanges {
		if strings.HasSuffix(e.Path, "/Hold") {
			gehalten++
			if e.Capture != model.CaptureComplete {
				t.Errorf("Exchange Hold: Capture %s", e.Capture)
			}
		}
	}
	if gehalten != 1 {
		t.Errorf("Exchange Hold %d-mal angehängt", gehalten)
	}
}
