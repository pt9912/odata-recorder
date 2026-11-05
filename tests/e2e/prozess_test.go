package e2e_test

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Fristen des Prozess-Helfers (SPEC-037 Nr. 5): Bereitschaft höchstens 10 s, Beenden binnen
// SHUTDOWN_TIMEOUT und 5 s Reserve.
const (
	bereitFrist   = 10 * time.Second
	shutdownFrist = 2 * time.Second
	reserve       = 5 * time.Second
)

// binaryPfad ist das Binary der Stufe build; ohne E2E_BINARY (außerhalb des Containers von
// `make test-integration`) endet der Test mit SKIP, und der Treiber wertet SKIP als rot.
func binaryPfad(t *testing.T) string {
	t.Helper()
	p := os.Getenv("E2E_BINARY")
	if p == "" {
		t.Skip("E2E_BINARY nicht gesetzt: die Stufe läuft nur im Container von make test-integration")
	}
	return p
}

// puffer sammelt die Bytes eines Stroms des Binary, auch während der Prozess läuft.
type puffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (p *puffer) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.b.Write(b)
}

func (p *puffer) bytes() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return bytes.Clone(p.b.Bytes())
}

func (p *puffer) String() string { return string(p.bytes()) }

// konfig sind die Werte der Umgebung, die der Test allein setzt (SPEC-037 Nr. 5).
type konfig struct {
	modus, dir, upstream string
	// dienst ist SERVICE_ID, leer heißt "ref".
	dienst string
	extra  []string
}

func (k konfig) env(listen string) []string {
	dienst := k.dienst
	if dienst == "" {
		dienst = "ref"
	}
	env := []string{"MODE=" + k.modus, "LISTEN=" + listen, "DATA_DIR=" + k.dir, "SERVICE_ID=" + dienst,
		"SERVICE_ROOT_PATH=/odata", "SHUTDOWN_TIMEOUT=" + shutdownFrist.String()}
	if k.upstream != "" {
		env = append(env, "UPSTREAM_URL="+k.upstream+"/odata")
	}
	return append(env, k.extra...)
}

// prozess ist das Binary als Kindprozess des Tests.
type prozess struct {
	t         *testing.T
	cmd       *exec.Cmd
	adresse   string
	modus     string
	aus, fehl *puffer
	ende      chan struct{}
	exit      int
	// gewartet wird nur über ende; exit gilt nach dem Schließen.
}

func freierPort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().String()
}

// startRoh startet `binary start` mit genau der Umgebung env und liest beide Ströme mit.
func startRoh(t *testing.T, adresse string, env []string) *prozess {
	t.Helper()
	p := &prozess{t: t, adresse: adresse, aus: &puffer{}, fehl: &puffer{}, ende: make(chan struct{})}
	p.cmd = exec.Command(binaryPfad(t), "start")
	p.cmd.Env, p.cmd.Stdout, p.cmd.Stderr = env, p.aus, p.fehl
	if err := p.cmd.Start(); err != nil {
		t.Fatalf("Start des Binary: %v", err)
	}
	go func() {
		_ = p.cmd.Wait()
		p.exit = p.cmd.ProcessState.ExitCode()
		close(p.ende)
	}()
	t.Cleanup(func() {
		select {
		case <-p.ende:
		default:
			_ = p.cmd.Process.Kill()
			<-p.ende
		}
	})
	return p
}

// starte startet das Binary mit k auf einem freien Port.
func starte(t *testing.T, k konfig) *prozess {
	t.Helper()
	adresse := freierPort(t)
	p := startRoh(t, adresse, k.env(adresse))
	p.modus = k.modus
	return p
}

func (p *prozess) beendet() bool {
	select {
	case <-p.ende:
		return true
	default:
		return false
	}
}

// bereit wartet, bis das Binary bedient: im Modus record gelingt `GET /odata/$metadata` (der Request wird
// aufgezeichnet, seine Antwort ist die Rückgabe), im Modus replay gelingt die Verbindung.
func (p *prozess) bereit() antwort {
	p.t.Helper()
	if p.modus != "record" {
		p.bereitOhneRequest()
		return antwort{}
	}
	var a antwort
	p.warte(func() bool {
		var err error
		a, err = sende(p.adresse, anfrage{pfad: "/odata/$metadata"})
		return err == nil
	})
	return a
}

// bereitOhneRequest wartet auf die Verbindung, ohne einen Request zu stellen.
func (p *prozess) bereitOhneRequest() {
	p.t.Helper()
	p.warte(func() bool {
		c, err := net.DialTimeout("tcp", p.adresse, time.Second)
		if err == nil {
			_ = c.Close()
		}
		return err == nil
	})
}

// warte ruft versuch, bis er gelingt; nach bereitFrist oder beim Ende des Prozesses ist der Test rot.
func (p *prozess) warte(versuch func() bool) {
	p.t.Helper()
	ende := time.Now().Add(bereitFrist)
	for time.Now().Before(ende) && !p.beendet() {
		if versuch() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	p.t.Fatalf("Binary nicht bereit binnen %v (beendet: %v)\nstdout: %s\nstderr: %s", bereitFrist, p.beendet(), p.aus, p.fehl)
}

// beende sendet SIGTERM einmal und liefert den Exit-Code; hängt der Prozess über SHUTDOWN_TIMEOUT und die
// Reserve hinaus, kommt SIGKILL und der Test ist rot.
func (p *prozess) beende() int {
	p.t.Helper()
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		p.t.Fatalf("SIGTERM: %v", err)
	}
	select {
	case <-p.ende:
	case <-time.After(shutdownFrist + reserve):
		_ = p.cmd.Process.Kill()
		<-p.ende
		p.t.Errorf("Binary endet nicht binnen %v nach SIGTERM", shutdownFrist+reserve)
	}
	return p.exit
}

// verweigert prüft, dass die Schleife keine Verbindung mehr annimmt (`connection refused`).
func (p *prozess) verweigert() {
	p.t.Helper()
	c, err := net.DialTimeout("tcp", p.adresse, time.Second)
	if err == nil {
		_ = c.Close()
		p.t.Errorf("%s nimmt nach dem Ende noch Verbindungen an", p.adresse)
	} else if !errors.Is(err, syscall.ECONNREFUSED) {
		p.t.Errorf("Verbindung zu %s: %v, erwartet connection refused", p.adresse, err)
	}
}

// warteEnde wartet höchstens bereitFrist auf das Ende des Prozesses, der von selbst enden muss; sonst ist der
// Test rot.
func (p *prozess) warteEnde() {
	p.t.Helper()
	select {
	case <-p.ende:
	case <-time.After(bereitFrist):
		p.t.Fatalf("Binary endet nicht binnen %v\nstdout: %s\nstderr: %s", bereitFrist, p.aus, p.fehl)
	}
}

// warteAufZeile wartet bis zu drei Sekunden auf eine Zeile von stderr, die alle Teile trägt.
func (p *prozess) warteAufZeile(teile ...string) string {
	p.t.Helper()
	if z := p.warteAufZeilen(1, teile...); len(z) > 0 {
		return z[0]
	}
	return ""
}

// warteAufZeilen wartet bis zu drei Sekunden auf n Zeilen von stderr, die alle Teile tragen; ohne sie ist
// der Test rot.
func (p *prozess) warteAufZeilen(n int, teile ...string) []string {
	p.t.Helper()
	ende := time.Now().Add(3 * time.Second)
	for {
		var treffer []string
		for _, z := range strings.Split(p.fehl.String(), "\n") {
			if alle(z, teile) {
				treffer = append(treffer, z)
			}
		}
		if len(treffer) >= n {
			return treffer
		}
		if time.Now().After(ende) {
			p.t.Errorf("%d von %d Zeilen auf stderr mit %q:\n%s", len(treffer), n, teile, p.fehl)
			return treffer
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func alle(s string, teile []string) bool {
	for _, t := range teile {
		if !strings.Contains(s, t) {
			return false
		}
	}
	return true
}

// ---- HTTP-Client des Tests

// anfrage ist ein Request des Clients.
type anfrage struct {
	methode, pfad string
	header        map[string][]string
	body          string
}

// antwort hält, was der Client sah.
type antwort struct {
	status int
	header http.Header
	body   []byte
}

// sende stellt die Anfrage an adresse mit den Headern, die der Referenzdienst für die drei Geheimnisse
// braucht: Authorization trägt GEHEIM-A, Cookie GEHEIM-B.
func sende(adresse string, a anfrage) (antwort, error) {
	m := a.methode
	if m == "" {
		m = http.MethodGet
	}
	req, err := http.NewRequest(m, "http://"+adresse+a.pfad, strings.NewReader(a.body))
	if err != nil {
		return antwort{}, err
	}
	req.Header.Set("Authorization", "Bearer "+geheimA)
	req.Header.Set("Cookie", "sid="+geheimB)
	req.Header.Set("Accept", "*/*")
	for k, vs := range a.header {
		req.Header.Del(k)
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	tr := &http.Transport{DisableKeepAlives: true}
	defer tr.CloseIdleConnections()
	resp, err := (&http.Client{Transport: tr, Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return antwort{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	b := new(bytes.Buffer)
	if _, err := b.ReadFrom(resp.Body); err != nil {
		return antwort{}, fmt.Errorf("Body: %w", err)
	}
	return antwort{status: resp.StatusCode, header: resp.Header, body: b.Bytes()}, nil
}

func holen(t *testing.T, adresse string, a anfrage) antwort {
	t.Helper()
	r, err := sende(adresse, a)
	if err != nil {
		t.Fatalf("%s %s: %v", a.methode, a.pfad, err)
	}
	return r
}
