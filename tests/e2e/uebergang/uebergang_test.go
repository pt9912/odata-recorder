package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/tests/e2e/fixture"
)

// ablage ist ein DATA_DIR mit einer beendeten Session und dem Workspace aus ihrem Import.
type ablage struct {
	dir, session, workspace string
}

// aufbau nimmt `$metadata` eines Dienstes im Prozess des Tests im Modus record auf, beendet die Session und
// fährt herunter; danach läuft `import` des Übergangs.
func aufbau(t *testing.T) ablage {
	t.Helper()
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(fixture.CSDL))
	}))
	t.Cleanup(srv.Close)
	a := ablage{dir: t.TempDir()}
	k, err := bootstrap.LadeKonfiguration([]string{
		"MODE=record", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + a.dir, "SERVICE_ID=ref",
		"SERVICE_ROOT_PATH=/odata", "UPSTREAM_URL=" + srv.URL + "/odata",
	})
	if err != nil {
		t.Fatal(err)
	}
	k.Listen = "127.0.0.1:0"
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get("http://" + h.Adresse() + "/odata/$metadata")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("$metadata über den Recorder: %v, %v", resp, err)
	}
	_ = resp.Body.Close()
	ss, err := h.Sessions().List(ctx)
	if err != nil || len(ss) != 1 {
		t.Fatalf("Sessions: %v, %v", ss, err)
	}
	a.session = ss[0].ID
	if _, err := h.Sessions().Beenden(ctx, a.session); err != nil {
		t.Fatal(err)
	}
	if err := h.Herunterfahren(ctx); err != nil {
		t.Fatal(err)
	}
	out, errOut, code := fuehre("import", a.dir, a.session)
	if code != 0 {
		t.Fatalf("import: Exit %d, stderr %q", code, errOut)
	}
	for _, z := range strings.Split(out, "\n") {
		if w, ok := strings.CutPrefix(z, "workspace="); ok {
			a.workspace = w
		}
	}
	if a.workspace == "" {
		t.Fatalf("import: stdout ohne workspace=: %q", out)
	}
	return a
}

// fuehre ruft lauf und liefert stdout, stderr und den Exit-Code.
func fuehre(args ...string) (stdout, stderr string, code int) {
	var o, e bytes.Buffer
	code = lauf(context.Background(), args, &o, &e)
	return o.String(), e.String(), code
}

// TestAufrufformen: für jeden der drei Befehle ist jedes der beiden Argumente fehlend, leer oder ein
// zusätzliches Argument Exit 2 mit `Aufruf:` und ohne Ausgabe; ein unbekannter Befehl ebenso (SPEC-037 Nr. 6).
func TestAufrufformen(t *testing.T) {
	befehle := []string{"import", "zeige", "reset"}
	formen := []struct {
		name string
		args func(befehl string) []string
	}{
		{"beide Argumente fehlen", func(b string) []string { return []string{b} }},
		{"zweites Argument fehlt", func(b string) []string { return []string{b, "/tmp"} }},
		{"DATA_DIR leer", func(b string) []string { return []string{b, "", "x"} }},
		{"zweites Argument leer", func(b string) []string { return []string{b, "/tmp", ""} }},
		{"beide leer", func(b string) []string { return []string{b, "", ""} }},
		{"ein Argument zu viel", func(b string) []string { return []string{b, "/tmp", "x", "y"} }},
	}
	n := 0
	for _, b := range befehle {
		for _, f := range formen {
			n++
			out, errOut, code := fuehre(f.args(b)...)
			if code != 2 || !strings.Contains(errOut, "Aufruf:") || out != "" {
				t.Errorf("%s, %s: Exit %d, stdout %q, stderr %q, erwartet 2 mit Aufruf: und ohne stdout", b, f.name, code, out, errOut)
			}
		}
	}
	if n != 18 {
		t.Errorf("Menge hat %d Elemente, erwartet 18 (3 Befehle, 6 Formen)", n)
	}
	for _, args := range [][]string{{"loesche", "/tmp", "x"}, {}} {
		if _, errOut, code := fuehre(args...); code != 2 || !strings.Contains(errOut, "Aufruf:") {
			t.Errorf("Aufruf %q: Exit %d, stderr %q", args, code, errOut)
		}
	}
}

// TestReset: der gültige Aufruf liefert `revision=<Zahl>` und erhöht sie mit jedem Aufruf; ein unbekannter
// Workspace und die Kennung einer Session sind Exit 1 mit ODR-E7001; ein DATA_DIR, das nicht besteht, und
// eine Datei an seiner Stelle sind Exit 1, und das Verzeichnis wird nicht angelegt (SPEC-037 Nr. 6).
func TestReset(t *testing.T) {
	a := aufbau(t)
	revision := func() int {
		t.Helper()
		out, errOut, code := fuehre("reset", a.dir, a.workspace)
		var r int
		if n, err := fmt.Sscanf(out, "revision=%d\n", &r); code != 0 || n != 1 || err != nil || out != fmt.Sprintf("revision=%d\n", r) {
			t.Fatalf("reset: Exit %d, stdout %q, stderr %q, erwartet revision=<Zahl>", code, out, errOut)
		}
		return r
	}
	erste := revision()
	if zweite := revision(); zweite != erste+1 || erste < 1 {
		t.Errorf("Revisionen %d, dann %d, erwartet jeder Reset um 1 höher und ab 1", erste, zweite)
	}
	fehler := []struct {
		name, dir, id, code string
	}{
		{"unbekannter Workspace", a.dir, "ws-gibt-es-nicht", "ODR-E7001"},
		{"Kennung einer Session", a.dir, a.session, "ODR-E7001"},
	}
	for _, f := range fehler {
		out, errOut, code := fuehre("reset", f.dir, f.id)
		if code != 1 || !strings.Contains(errOut, f.code) || out != "" {
			t.Errorf("%s: Exit %d, stdout %q, stderr %q, erwartet 1 mit %s", f.name, code, out, errOut, f.code)
		}
	}
	fehlt := filepath.Join(t.TempDir(), "gibt-es-nicht")
	datei := filepath.Join(t.TempDir(), "datei")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, dir := range map[string]string{"DATA_DIR besteht nicht": fehlt, "DATA_DIR ist eine Datei": datei} {
		out, errOut, code := fuehre("reset", dir, a.workspace)
		if code != 1 || errOut == "" || out != "" {
			t.Errorf("%s: Exit %d, stdout %q, stderr %q, erwartet 1 mit Meldung", name, code, out, errOut)
		}
	}
	if _, err := os.Stat(fehlt); !os.IsNotExist(err) {
		t.Errorf("reset hat das DATA_DIR angelegt: %v", err)
	}
}
