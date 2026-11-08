package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// TestAngehaltenJedeRequestart hält A2 (SPEC-238 Nr. 3 b, 5): ein angehaltener Server antwortet jeder der 26 Requestarten
// aus R mit 503 ODR-E5010 samt Connection: close, mit der einen vergebenen ID und ohne WWW-Authenticate und Allow, vor
// Authentisierung, Zuordnung, Route und Audit; die 27. Art (503) steht in den Schleifen A1, E2 und E4.
func TestAngehaltenJedeRequestart(t *testing.T) {
	n := 0
	for _, a := range requestArten() {
		if a.angehalten {
			continue
		}
		n++
		a.angehalten = true
		l := baue(t, a)
		rec := l.u.rufe(ohneAuth(a.methode, a.ziel, a.kopf))
		pruefeFehlerAntwort(t, a.name, rec, 503, objekt503, "")
		soll := map[string][]string{
			"Connection": {"close"}, "WWW-Authenticate": nil, "Allow": nil, "X-Correlation-Id": {l.u.ids.letzte()},
		}
		for name, v := range soll {
			if got := kopfWerte(rec, name); !slices.Equal(got, v) {
				t.Errorf("%s: %s %q, erwartet %q", a.name, name, got, v)
			}
		}
		if k, uebrige := l.u.ids.zahlen(); k != 1 || uebrige != 0 {
			t.Errorf("%s: KorrelationsID %d-mal, andere Operationen %d-mal, erwartet 1 und 0", a.name, k, uebrige)
		}
		if l.vf.zahl() != 0 || l.ok.aufrufe() != 0 || l.fehler.aufrufe() != 0 || l.uc.summe() != 0 || l.u.audit.Aufrufe() != 0 {
			t.Errorf("%s: Vergleich %d, Routen %d und %d, Use-Case %d, Audit %d Aufrufe, erwartet 0",
				a.name, l.vf.zahl(), l.ok.aufrufe(), l.fehler.aufrufe(), l.uc.summe(), l.u.audit.Aufrufe())
		}
	}
	gleichAuf(t, "Zahl der Requestarten ohne 503", n, 26)
}

// TestZweiServerTeilenNichts hält K5 (SPEC-238 Nr. 1, 11): nach Anhalten von A liefert A 503 und B 200; nach Schliessen
// von A bleibt B bedient und lässt sich selbst schließen.
func TestZweiServerTeilenNichts(t *testing.T) {
	a, b := neuLive(t, neueRoute()), neuLive(t, neueRoute())
	va, vb := a.verbinde(t), b.verbinde(t)
	va.erledige(t)
	vb.erledige(t)
	a.srv.Anhalten()
	va.schreibe(t, getP())
	if resp, _ := va.antwort(t, "GET"); resp.StatusCode != 503 {
		t.Errorf("A nach Anhalten: Status %d, erwartet 503", resp.StatusCode)
	}
	vb.erledige(t)
	if _, err := schliesse(t, a.srv, time.Second); err != nil {
		t.Errorf("Schliessen von A = %v, erwartet nil", err)
	}
	if !dialScheitert(a.addr) || dialScheitert(b.addr) {
		t.Errorf("nach Schliessen von A: Dial A scheitert %v, Dial B scheitert %v; erwartet wahr und falsch", dialScheitert(a.addr), dialScheitert(b.addr))
	}
	if x := erwarte(t, b.async("GET", "/p"), "Antwort von B"); x.err != nil || x.status != 200 {
		t.Errorf("B nach Schliessen von A: %+v, erwartet 200", x)
	}
	gleichAuf(t, "Aufrufe der Route von B", b.route.aufrufe(), 3)
	if _, err := schliesse(t, b.srv, time.Second); err != nil {
		t.Errorf("Schliessen von B = %v, erwartet nil", err)
	}
}

// TestWerteDesServers hält G1 (SPEC-238 Nr. 10): die sieben Felder des http.Server als Literale; ErrorLog schreibt auf
// io.Discard.
func TestWerteDesServers(t *testing.T) {
	g := neuMit(t).srv.GrenzenFuerTest()
	felder := []struct {
		name      string
		got, soll any
	}{
		{"ReadHeaderTimeout", g.ReadHeaderTimeout, 10 * time.Second},
		{"ReadTimeout", g.ReadTimeout, 30 * time.Second},
		{"IdleTimeout", g.IdleTimeout, 120 * time.Second},
		{"MaxHeaderBytes", g.MaxHeaderBytes, 65536},
		{"WriteTimeout", g.WriteTimeout, time.Duration(0)},
		{"ErrorLog verwirft", g.ErrorLog != nil && g.ErrorLog.Writer() == io.Discard, true},
		{"DisableGeneralOptionsHandler", g.DisableGeneralOptionsHandler, true},
	}
	gleichAuf(t, "Zahl der Felder", len(felder), reflect.TypeOf(g).NumField())
	gleichAuf(t, "Zahl der Felder", len(felder), 7)
	for _, f := range felder {
		gleichAuf(t, f.name, f.got, f.soll)
	}
}

const kindUmgebung = "ODR_G1_KIND"

// TestKeinLogDerBibliothek hält G1 (SPEC-238 Nr. 10, 13 h): der Adapter schreibt kein Log von net/http; die Panik der
// Route erscheint weder in log.Default() noch auf stderr oder stdout. Der Test fährt das Szenario in einem Kindprozess
// (dieselbe Testbinärdatei) und liest dessen stderr und stdout, den Standardlogger unverändert (Prüfgrenze N-6).
func TestKeinLogDerBibliothek(t *testing.T) {
	if dir := os.Getenv(kindUmgebung); dir != "" {
		kindG1(t, dir)
		return
	}
	dir := t.TempDir()
	ctx, abbruch := context.WithTimeout(context.Background(), 60*time.Second)
	defer abbruch()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestKeinLogDerBibliothek$", "-test.count=1")
	cmd.Env = append(os.Environ(), kindUmgebung+"="+dir)
	var aus, fehl bytes.Buffer
	cmd.Stdout, cmd.Stderr = &aus, &fehl
	if err := cmd.Run(); err != nil {
		t.Fatalf("Kindprozess: %v\nstdout: %s\nstderr: %s", err, aus.String(), fehl.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "lief")); err != nil {
		t.Fatalf("der Kindprozess hat das Szenario nicht ausgeführt: %v", err)
	}
	for senke, text := range map[string]string{"stderr": fehl.String(), "stdout": aus.String()} {
		if strings.Contains(text, "geheim-wert-G1") {
			t.Errorf("%s trägt die Panik: %q", senke, text)
		}
	}
}

// kindG1 ist das Szenario im Kindprozess: die Route gerät in Panik, der Client sieht den Abbruch.
func kindG1(t *testing.T, dir string) {
	t.Helper()
	route := neueRoute()
	route.aendere = func(admin.Anfrage) { panic("geheim-wert-G1") }
	u := neuLive(t, route)
	if x := erwarte(t, u.async("GET", "/p"), "Antwort"); x.err == nil {
		t.Fatalf("der Client sah keinen Abbruch, Status %d", x.status)
	}
	if err := os.WriteFile(filepath.Join(dir, "lief"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestOptionsStern hält G2 (SPEC-238 Nr. 10, 3 d, 2): `OPTIONS *` geht durch die Authentisierung: ohne Token 401, mit
// Token 404 (`*` passt auf keine Route); die Route läuft nie.
func TestOptionsStern(t *testing.T) {
	route := neueRoute()
	u := neuLive(t, route)
	faelle := []struct {
		name, kopf string
		status     int
		body       string
	}{
		{"ohne Token", "", 401, objekt401},
		{"mit Token", kopfAuth(), 404, objekt404},
	}
	for _, f := range faelle {
		v := u.verbinde(t)
		v.schreibe(t, []byte("OPTIONS * HTTP/1.1\r\nHost: h\r\n"+f.kopf+"\r\n"))
		resp, body := v.antwort(t, "OPTIONS")
		if resp.StatusCode != f.status || string(body) != f.body {
			t.Errorf("%s: Antwort %d %q, erwartet %d %q", f.name, resp.StatusCode, body, f.status, f.body)
		}
	}
	gleichAuf(t, "Aufrufe der Route", route.aufrufe(), 0)
}

// gelesen ist, was die Route vom Body las.
type gelesen struct {
	n   int64
	err error
}

// TestBodyGrenze hält G3 (SPEC-238 Nr. 10): 1048576 Bytes lassen sich lesen, das 1048577. ergibt *http.MaxBytesError;
// 4 Größen × 2 Formen (Content-Length, chunked).
func TestBodyGrenze(t *testing.T) {
	groessen := []int{0, 1048575, 1048576, 1048577}
	formen := []string{"Content-Length", "chunked"}
	n := 0
	for _, g := range groessen {
		for _, f := range formen {
			n++
			route := neueRoute()
			res := make(chan gelesen, 1)
			route.aendere = func(a admin.Anfrage) {
				b, err := io.Copy(io.Discard, a.Body)
				res <- gelesen{b, err}
			}
			u := neuLive(t, route)
			v := u.verbinde(t)
			kopf := "POST /p HTTP/1.1\r\nHost: h\r\n" + kopfAuth()
			body := muster(g, 9)
			if f == "chunked" {
				kopf += "Transfer-Encoding: chunked\r\n\r\n"
				if g > 0 {
					body = append([]byte(fmt.Sprintf("%x\r\n", g)), append(body, "\r\n"...)...)
				}
				body = append(body, "0\r\n\r\n"...)
			} else {
				kopf += fmt.Sprintf("Content-Length: %d\r\n\r\n", g)
			}
			v.schreibe(t, append([]byte(kopf), body...))
			var got gelesen
			select {
			case got = <-res:
			case <-time.After(5 * time.Second):
				t.Fatalf("%d Bytes, %s: die Route las nicht zu Ende", g, f)
			}
			var grenze *http.MaxBytesError
			soll := int64(min(g, 1048576))
			if got.n != soll || (g > 1048576) != errors.As(got.err, &grenze) || (g <= 1048576 && got.err != nil) {
				t.Errorf("%d Bytes, %s: gelesen %d, Fehler %v; erwartet %d und %v", g, f, got.n, got.err, soll, map[bool]string{true: "*http.MaxBytesError", false: "kein Fehler"}[g > 1048576])
			}
		}
	}
	gleichAuf(t, "Zahl der Elemente", n, 8)
}

// TestKontextEndetBeimTrennen hält X2 Fall 1 (SPEC-238 Nr. 9): der ctx der Route endet, wenn der Client trennt, bei
// einem Request ohne Body und bei einem mit gelesenem Body. Fall 2 und 3 (Ablauf der Frist) hält
// TestSchliessenKontextNachFrist.
func TestKontextEndetBeimTrennen(t *testing.T) {
	faelle := []struct {
		name, request string
		liest         bool
	}{
		{"GET ohne Body", "GET /p HTTP/1.1\r\nHost: h\r\n" + kopfAuth() + "\r\n", false},
		{"POST mit gelesenem Body", "POST /p HTTP/1.1\r\nHost: h\r\n" + kopfAuth() + "Content-Length: 5\r\n\r\nhello", true},
	}
	for _, f := range faelle {
		route := neueHaengRoute(t, wartetCtx)
		if f.liest {
			route.aendere = func(a admin.Anfrage) { _, _ = io.Copy(io.Discard, a.Body) }
		}
		u := neuLive(t, route)
		v := u.verbinde(t)
		v.schreibe(t, []byte(f.request))
		route.warteBegonnen(t, 1)
		_ = v.c.Close()
		warteBis(t, f.name+": die Route sah das Ende von ctx", func() bool { return len(route.ctxFehlerListe()) == 1 })
		gleichAuf(t, f.name+": Fehler von ctx", route.ctxFehlerListe(), []error{context.Canceled})
	}
}
