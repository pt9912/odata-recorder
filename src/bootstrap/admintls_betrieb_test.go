package bootstrap_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// adminIDs zählt jeden Aufruf des IDGenerator, gleich welche Operation.
type adminIDs struct {
	*supporttest.FestIDs
	n atomic.Int64
}

func (g *adminIDs) SessionID() string      { g.n.Add(1); return g.FestIDs.SessionID() }
func (g *adminIDs) ExchangeID() string     { g.n.Add(1); return g.FestIDs.ExchangeID() }
func (g *adminIDs) KorrelationsID() string { g.n.Add(1); return g.FestIDs.KorrelationsID() }
func (g *adminIDs) WorkspaceID() string    { g.n.Add(1); return g.FestIDs.WorkspaceID() }
func (g *adminIDs) SnapshotID() string     { g.n.Add(1); return g.FestIDs.SnapshotID() }

// adminListener ist der Listener aus `net.Listen`, den der Root umhüllt: er zählt die Annahmen und meldet jedes Close einer angenommenen Verbindung.
type adminListener struct {
	net.Listener
	zu       chan struct{}
	gesehen  int
	angenomm atomic.Int64
}

func (l *adminListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	l.angenomm.Add(1)
	return &adminConn{Conn: c, l: l}, nil
}

// warte wartet, bis insgesamt n Close-Aufrufe angenommener Verbindungen gemeldet sind; bleibt eine Meldung 3 s aus, ist das ein Fehler des Tests.
func (l *adminListener) warte(t *testing.T, n int) {
	t.Helper()
	for l.gesehen < n {
		select {
		case <-l.zu:
			l.gesehen++
		case <-time.After(3 * time.Second):
			t.Errorf("%d von %d erwarteten Close-Aufrufen des Servers, 3 s ohne weitere Meldung", l.gesehen, n)
			return
		}
	}
}

// adminConn meldet sein Close an den Listener, nachdem es die Verbindung geschlossen hat.
type adminConn struct {
	net.Conn
	l *adminListener
}

func (c *adminConn) Close() error {
	err := c.Conn.Close()
	select {
	case c.l.zu <- struct{}{}:
	default:
	}
	return err
}

// adminRoute ist die Route des Test-Servers: sie hält jede Anfrage fest und antwortet 200.
type adminRoute struct {
	mu       sync.Mutex
	anfragen []string
	ids      []string
}

func (r *adminRoute) handle(_ context.Context, a admin.Anfrage) (admin.Antwort, error) {
	body, _ := io.ReadAll(a.Body)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.anfragen = append(r.anfragen, fmt.Sprintf("%s %s ?%s kopf=%v laenge=%d body=%q werte=%v", a.Methode, a.Pfad, a.RawQuery, a.Header, a.Laenge, body, a.Pfadwerte))
	r.ids = append(r.ids, a.KorrelationsID)
	return admin.Antwort{Status: http.StatusOK, Body: []byte(`{"ok":true}`)}, nil
}

func (r *adminRoute) zahl() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.anfragen)
}

// adminServer ist ein Admin-Server im Prozess hinter dem umhüllten Listener; ca ist die Wurzel der Clients (nil: Klartext).
type adminServer struct {
	srv     *admin.Server
	adresse string
	inner   *adminListener
	ids     *adminIDs
	audit   *supporttest.FakeAudit
	tele    *supporttest.FakeTelemetry
	route   *adminRoute
	ca      *tlstest.Aussteller
}

// adminTokenWert ist das Token des Test-Servers (32 Bytes, die Mindestlänge von SPEC-239 Nr. 1 g).
func adminTokenWert() string { return strings.Repeat("t", 32) }

// neuerAdminServer bedient mit admin.New den Listener, den tr aus einem Listener auf 127.0.0.1 macht; er routet `GET /probe` und die Methoden und Muster von
// weitere, alle auf dieselbe Route. Die Uhr steht fest, das Fenster der Einträge `admin_abgewiesen` (SPEC-301) läuft im Test nicht ab. Der Test schließt den
// Server am Ende.
func neuerAdminServer(t *testing.T, tr bootstrap.AdminTransport, ca *tlstest.Aussteller, weitere ...admin.Route) *adminServer {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &adminServer{
		inner: &adminListener{Listener: l, zu: make(chan struct{}, 64)}, ids: &adminIDs{FestIDs: supporttest.NewFestIDs()},
		audit: supporttest.NewFakeAudit(), tele: supporttest.NewFakeTelemetry(), route: &adminRoute{}, ca: ca, adresse: l.Addr().String(),
	}
	routen := append([]admin.Route{{Methode: http.MethodGet, Muster: "/probe"}}, weitere...)
	for i := range routen {
		routen[i].Handler = s.route.handle
	}
	srv, err := admin.New(admin.Konfig{
		Listener: bootstrap.AdminUmhuelle(tr, s.inner), Token: []byte(adminTokenWert()), IDs: s.ids, Audit: s.audit, Telemetrie: s.tele,
		Routen: routen,
		Uhr:    supporttest.NewFestUhr(time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("admin.New: %v", err)
	}
	s.srv = srv
	ende := make(chan error, 1)
	go func() { ende <- srv.Serve() }()
	t.Cleanup(func() {
		_ = srv.Schliessen(5 * time.Second)
		select {
		case <-ende:
		case <-time.After(10 * time.Second):
			t.Error("Serve kehrt nach Schliessen nicht zurück")
		}
	})
	return s
}

// adminTLS legt ein Paar von a ab und bedient es.
func (w *werkstatt) adminTLS(t *testing.T, a *tlstest.Aussteller) *adminServer {
	t.Helper()
	cert, key, _ := w.tlsPaar(a, tlstest.Optionen{})
	return neuerAdminServer(t, w.ladeAdmin(cert, key, w.jetzt), a)
}

// dial verbindet mit TLS; Frist 5 s für Verbindung und Handshake.
func (s *adminServer) dial(cfg *tls.Config) (*tls.Conn, error) {
	return tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", s.adresse, cfg)
}

// adminAntwort ist die Antwort einer Anfrage; kopf sind die Header ohne Datum und ohne `X-Correlation-Id`, id und allow stehen einzeln.
type adminAntwort struct {
	status int
	kopf   string
	body   string
	id     string
	allow  string
}

// adminAntwortVon bildet aus Status, Headern und Body die adminAntwort.
func adminAntwortVon(status int, header http.Header, body string) adminAntwort {
	kopf := header.Clone()
	kopf.Del("Date")
	kopf.Del("X-Correlation-Id")
	return adminAntwort{status: status, kopf: fmt.Sprint(kopf), body: body, id: header.Get("X-Correlation-Id"), allow: header.Get("Allow")}
}

// hole sendet `GET pfad` (mit Token, wenn token nicht leer ist; inhalt ist der Body der Anfrage) per HTTP oder, mit ca, per HTTPS; ein Fehler beendet den
// Test.
func (s *adminServer) hole(t *testing.T, pfad, token string, inhalt ...string) adminAntwort {
	t.Helper()
	tr := &http.Transport{DisableKeepAlives: true}
	schema := "http"
	if s.ca != nil {
		tr.TLSClientConfig, schema = s.ca.Klient("localhost"), "https"
	}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, schema+"://"+s.adresse+pfad, strings.NewReader(strings.Join(inhalt, "")))
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second, Transport: tr}).Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Body von %s: %v", pfad, err)
	}
	return adminAntwortVon(resp.StatusCode, resp.Header, string(body))
}

// frageRoh schreibt `methode ziel` unverändert als Request-Zeile auf c, mit Token, wenn token nicht leer ist, und liest die Antwort; Frist 5 s.
func frageRoh(c net.Conn, br *bufio.Reader, methode, ziel, token string) (*http.Response, string, error) {
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	kopf := methode + " " + ziel + " HTTP/1.1\r\nHost: admin\r\n"
	if token != "" {
		kopf += "Authorization: Bearer " + token + "\r\n"
	}
	if _, err := io.WriteString(c, kopf+"\r\n"); err != nil {
		return nil, "", fmt.Errorf("Anfrage: %w", err)
	}
	resp, err := http.ReadResponse(br, &http.Request{Method: methode})
	if err != nil {
		return nil, "", fmt.Errorf("Antwort: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("Body: %w", err)
	}
	return resp, string(body), nil
}

// holeRoh schreibt `GET pfad` mit Token auf c und liest die Antwort; ein Fehler beendet den Test.
func holeRoh(t *testing.T, c net.Conn, br *bufio.Reader, pfad string) (*http.Response, string) {
	t.Helper()
	resp, body, err := frageRoh(c, br, http.MethodGet, pfad, adminTokenWert())
	if err != nil {
		t.Fatalf("GET %s: %v", pfad, err)
	}
	return resp, body
}

// bedient fragt `GET /healthz` auf c und meldet jede andere Antwort als HTTP/1.1 200 mit dem Body von /healthz. Unter TLS 1.3 erfährt der Client eine
// Ablehnung seines Handshakes erst beim Lesen, darum belegt allein die Antwort, dass der Server die Verbindung angenommen hat.
func bedient(c net.Conn) error {
	resp, body, err := frageRoh(c, bufio.NewReader(c), http.MethodGet, "/healthz", "")
	switch {
	case err != nil:
		return err
	case resp.StatusCode != 200 || resp.ProtoMajor != 1 || resp.ProtoMinor != 1 || body != `{"status":"ok"}`:
		return fmt.Errorf("%s %s, Body %q; erwartet HTTP/1.1 200 mit dem Body von /healthz", resp.Proto, resp.Status, body)
	}
	return nil
}

// TestAdminVerbindung (AT11, M-Verbindung): die Verbindung folgt SPEC-291 Nr. 5 und 6: TLS 1.2 und 1.3 gelingen, 1.1 scheitert; ALPN: allein `h2` scheitert,
// `h2` und `http/1.1`, kein ALPN und `http/1.1` sprechen HTTP/1.1; ein Client ohne Zertifikat gelingt; die Kette der Datei wird gezeigt (SPEC-299 Nr. 7).
// Menge: Version (3) + ALPN (4) + Client ohne Zertifikat + Kette = 9 Zellen.
func TestAdminVerbindung(t *testing.T) {
	w := neueWerkstatt(t)
	s := w.adminTLS(t, w.wurzel)
	kb := w.wurzel.Zwischen(t, tlstest.Optionen{}).Blatt(t, tlstest.Optionen{})
	kette := neuerAdminServer(t, w.ladeAdmin(w.datei(kb.KettePEM()), w.datei(w.schluessel(kb)), w.jetzt), w.wurzel)
	h1 := []string{"http/1.1"}
	faelle := []struct {
		name      string
		min, max  uint16
		protos    []string
		scheitert bool
		version   uint16
		proto     string
	}{
		{"TLS 1.1", tls.VersionTLS10, tls.VersionTLS11, h1, true, 0, ""},
		{"TLS 1.2", 0, tls.VersionTLS12, h1, false, tls.VersionTLS12, "http/1.1"},
		{"TLS 1.3", 0, 0, h1, false, tls.VersionTLS13, "http/1.1"},
		{"ALPN allein h2", 0, 0, []string{"h2"}, true, 0, ""},
		{"ALPN h2 und http/1.1", 0, 0, []string{"h2", "http/1.1"}, false, tls.VersionTLS13, "http/1.1"},
		{"ALPN keines", 0, 0, nil, false, tls.VersionTLS13, ""},
		{"ALPN http/1.1", 0, 0, h1, false, tls.VersionTLS13, "http/1.1"},
	}
	zahl := 0
	for _, f := range faelle {
		zahl++
		cfg := w.wurzel.Klient("localhost")
		cfg.MinVersion, cfg.MaxVersion, cfg.NextProtos = f.min, f.max, f.protos
		c, err := s.dial(cfg)
		if (err != nil) != f.scheitert {
			t.Errorf("%s: Fehler %v, erwartet scheitert=%v", f.name, err, f.scheitert)
		}
		if err != nil {
			continue
		}
		defer func() { _ = c.Close() }()
		if st := c.ConnectionState(); st.Version != f.version || st.NegotiatedProtocol != f.proto {
			t.Errorf("%s: Version %x und Protokoll %q, erwartet %x und %q", f.name, st.Version, st.NegotiatedProtocol, f.version, f.proto)
		}
		if err := bedient(c); err != nil {
			t.Errorf("%s: %v", f.name, err)
		}
	}
	zahl++
	ohne := w.wurzel.Klient("localhost")
	if len(ohne.Certificates) != 0 || ohne.GetClientCertificate != nil {
		t.Errorf("Client ohne Zertifikat: die Konfiguration trägt eines")
	}
	if c, err := s.dial(ohne); err != nil {
		t.Errorf("Client ohne Zertifikat: %v", err)
	} else {
		if err := bedient(c); err != nil {
			t.Errorf("Client ohne Zertifikat: %v", err)
		}
		_ = c.Close()
	}
	zahl++
	if c, err := kette.dial(w.wurzel.Klient("localhost")); err != nil || len(c.ConnectionState().PeerCertificates) != 2 {
		t.Errorf("Kette: Fehler %v; erwartet das Blatt und das Zwischenzertifikat (2 Zertifikate)", err)
	} else {
		_ = c.Close()
	}
	if zahl != 9 {
		t.Fatalf("Menge M-Verbindung hat %d Zellen, erwartet 9", zahl)
	}
}

// TestAdminRahmenUeberTLS (AT12, M-Anfrage): der Rahmen gilt über TLS wie im Klartext: dieselbe Antwort, dieselbe `Anfrage` an die Route, dieselben Einträge und
// Zähler (SPEC-299 Nr. 10 d). Menge: Anfrage (`GET /healthz` ohne Token, Route mit Token, Route ohne Token, Route mit falschem Token) × Transport (Klartext, TLS) = 8.
// Die Anfrage an die Route trägt eine Query und einen Body, damit keines von beiden leer verglichen wird.
func TestAdminRahmenUeberTLS(t *testing.T) {
	w := neueWerkstatt(t)
	server := []*adminServer{neuerAdminServer(t, bootstrap.AdminTransport{Weg: bootstrap.WegLoopback}, nil), w.adminTLS(t, w.wurzel)}
	anfragen := []struct {
		name, pfad, token, inhalt string
		status                    int
		body                      string
	}{
		{"GET /healthz ohne Token", "/healthz", "", "", 200, `{"status":"ok"}`},
		{"Route mit Token", "/probe?a=1&b=%20", adminTokenWert(), "nutzlast", 200, `{"ok":true}`},
		{"Route ohne Token", "/probe", "", "", 401, `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`},
		{"Route mit falschem Token", "/probe", strings.Repeat("x", 32), "", 401, `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`},
	}
	zahl := 0
	ids := [2][]string{}
	for _, a := range anfragen {
		var aus [2]adminAntwort
		for i, s := range server {
			zahl++
			aus[i] = s.hole(t, a.pfad, a.token, a.inhalt)
			ids[i] = append(ids[i], aus[i].id)
			if aus[i].status != a.status || aus[i].body != a.body || aus[i].id == "" {
				t.Errorf("%s, Transport %d: Status %d, Body %q, ID %q; erwartet %d und %q", a.name, i, aus[i].status, aus[i].body, aus[i].id, a.status, a.body)
			}
		}
		if aus[0] != aus[1] {
			t.Errorf("%s: die Antwort über TLS weicht ab:\nKlartext %+v\nTLS      %+v", a.name, aus[0], aus[1])
		}
	}
	// Die Zählprüfung steht vor den Zugriffen auf ids nach Index: eine fehlende Zelle ist ihr Befund.
	if zahl != 8 {
		t.Fatalf("Menge M-Anfrage hat %d Zellen, erwartet 8", zahl)
	}
	for i, s := range server {
		soll := []support.Eintrag{
			{Art: support.ArtAdminAbgewiesen, ErrorCode: model.CodeE2101, KorrelationsID: ids[i][2]},
			{Art: support.ArtAdminAbgewiesen, ErrorCode: model.CodeE2101, KorrelationsID: ids[i][3]},
		}
		got := s.audit.Eintraege()
		if fmt.Sprint(got) != fmt.Sprint(soll) || s.ids.n.Load() != 4 || s.tele.Anzahl(supporttest.OpCountError) != 2 || len(s.tele.Aufrufe()) != 2 {
			t.Errorf("Transport %d: Einträge %v (erwartet %v), IDs %d (erwartet 4), CountError %d und %d Aufrufe der Telemetrie (erwartet 2 und 2)",
				i, got, soll, s.ids.n.Load(), s.tele.Anzahl(supporttest.OpCountError), len(s.tele.Aufrufe()))
		}
		for _, a := range s.tele.Aufrufe() {
			if a.Code != model.CodeE2101 {
				t.Errorf("Transport %d: CountError mit %s, erwartet ODR-E2101", i, a.Code)
			}
		}
		if len(s.route.anfragen) != 1 || s.route.ids[0] != ids[i][1] {
			t.Errorf("Transport %d: Anfragen an die Route %q und IDs %q; erwartet eine mit der ID %s", i, s.route.anfragen, s.route.ids, ids[i][1])
		} else if got := s.route.anfragen[0]; !strings.Contains(got, "GET /probe ?a=1&b=%20 kopf=") || !strings.Contains(got, `laenge=8 body="nutzlast"`) {
			t.Errorf("Transport %d: die Anfrage an die Route ist %s; erwartet Query, Länge und Body der Anfrage", i, got)
		}
	}
	if a, b := server[0].route.anfragen, server[1].route.anfragen; len(a) == 1 && len(b) == 1 && a[0] != b[0] {
		t.Errorf("die Anfrage an die Route weicht über TLS ab:\nKlartext %s\nTLS      %s", a[0], b[0])
	}
}

// frage sendet `methode ziel` unverändert über eine eigene Verbindung (Klartext oder, mit ca, TLS), mit Token, wenn token nicht leer ist; ein Fehler beendet
// den Test.
func (s *adminServer) frage(t *testing.T, methode, ziel, token string) adminAntwort {
	t.Helper()
	var c net.Conn
	var err error
	if s.ca == nil {
		c, err = net.DialTimeout("tcp", s.adresse, 5*time.Second)
	} else {
		c, err = s.dial(s.ca.Klient("localhost"))
	}
	if err != nil {
		t.Fatalf("%s %s: %v", methode, ziel, err)
	}
	defer func() { _ = c.Close() }()
	resp, body, err := frageRoh(c, bufio.NewReader(c), methode, ziel, token)
	if err != nil {
		t.Fatalf("%s %s: %v", methode, ziel, err)
	}
	return adminAntwortVon(resp.StatusCode, resp.Header, body)
}

// adminZuordnung ist eine Anfragezeile mit der Antwort, die sie mit gültigem Token erhält; ohne Token ist jede dieser Antworten 401.
type adminZuordnung struct {
	methode, ziel string
	status        int
	allow         string
}

// adminZuordnungen sind die Zeilen von M-Zuordnung, je Antwortklasse der Zuordnung (SPEC-238 Nr. 2): Treffer (Literal, Parameter), 404 (kein Pfad-Treffer) und
// 405 (Pfad-Treffer ohne die Methode, mit `Allow`).
func adminZuordnungen() []adminZuordnung {
	return []adminZuordnung{
		{"GET", "/probe", 200, ""},
		{"GET", "/probe/a%2Fb", 200, ""},
		{"GET", "/unbekannt", 404, ""},
		{"GET", "/probe/", 404, ""},
		{"GET", "//probe", 404, ""},
		{"GET", "/Probe", 404, ""},
		{"GET", "/probe/x/y", 404, ""},
		{"OPTIONS", "*", 404, ""},
		{"GET", "/healthz/", 404, ""},
		{"DELETE", "/probe", 405, "GET, PUT"},
		{"get", "/probe", 405, "GET, PUT"},
		{"HEAD", "/probe", 405, "GET, PUT"},
		{"POST", "/probe/x", 405, "GET"},
		{"POST", "/healthz", 405, "GET"},
	}
}

// sollZuordnung ist Body und Code der Antwort mit dem Status als Literal; die Antwort auf `HEAD` trägt keinen Body, eine 200 keinen Code.
func sollZuordnung(methode string, status int) (body string, code model.Code) {
	switch status {
	case 200:
		return `{"ok":true}`, code
	case 401:
		body, code = `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`, model.CodeE2101
	default:
		body, code = `{"error":{"code":"ODR-E2102","message":"Admin request invalid","target":"admin"}}`, model.CodeE2102
	}
	if methode == http.MethodHead {
		body = ""
	}
	return body, code
}

// TestAdminZuordnungUeberTLS (AT12, M-Zuordnung): die Zuordnung ist über TLS dieselbe wie im Klartext, samt der Abweisung vor ihr und den Einträgen und Zählern
// ihrer Antworten (SPEC-299 Nr. 10 d, SPEC-238 Nr. 2, 3 d und e). Der Server routet `GET /probe`, `PUT /probe` und `GET /probe/{id}`. Menge: Anfragezeile (14,
// adminZuordnungen) × Authorization (gültiges Token, keine) × Transport (Klartext, TLS) = 56 Zellen.
func TestAdminZuordnungUeberTLS(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	weitere := []admin.Route{{Methode: http.MethodPut, Muster: "/probe"}, {Methode: http.MethodGet, Muster: "/probe/{id}"}}
	server := []*adminServer{
		neuerAdminServer(t, bootstrap.AdminTransport{Weg: bootstrap.WegLoopback}, nil, weitere...),
		neuerAdminServer(t, w.ladeAdmin(cert, key, w.jetzt), w.wurzel, weitere...),
	}
	zellen := 0
	var eintraege [2][]support.Eintrag
	var codes []model.Code
	for _, f := range adminZuordnungen() {
		for _, token := range []string{adminTokenWert(), ""} {
			status, allow := f.status, f.allow
			if token == "" {
				status, allow = 401, ""
			}
			body, code := sollZuordnung(f.methode, status)
			was := fmt.Sprintf("%s %s, Token %v", f.methode, f.ziel, token != "")
			var aus [2]adminAntwort
			for i, s := range server {
				zellen++
				aus[i] = s.frage(t, f.methode, f.ziel, token)
				if aus[i].status != status || aus[i].body != body || aus[i].allow != allow || aus[i].id == "" {
					t.Errorf("%s, Transport %d: Status %d, Body %q, Allow %q, ID %q; erwartet %d, %q und %q", was, i, aus[i].status, aus[i].body, aus[i].allow, aus[i].id, status, body, allow)
				}
				if status == 401 {
					eintraege[i] = append(eintraege[i], support.Eintrag{Art: support.ArtAdminAbgewiesen, ErrorCode: model.CodeE2101, KorrelationsID: aus[i].id})
				}
			}
			if status != 200 {
				codes = append(codes, code)
			}
			if aus[0] != aus[1] {
				t.Errorf("%s: die Antwort über TLS weicht ab:\nKlartext %+v\nTLS      %+v", was, aus[0], aus[1])
			}
		}
	}
	if zellen != 56 {
		t.Fatalf("Menge M-Zuordnung hat %d Zellen, erwartet 56", zellen)
	}
	for i, s := range server {
		var gezaehlt []model.Code
		for _, a := range s.tele.Aufrufe() {
			gezaehlt = append(gezaehlt, a.Code)
		}
		if got := s.audit.Eintraege(); len(got) != 14 || fmt.Sprint(got) != fmt.Sprint(eintraege[i]) || s.ids.n.Load() != 28 {
			t.Errorf("Transport %d: Einträge %v (erwartet 14: %v), IDs %d (erwartet 28)", i, got, eintraege[i], s.ids.n.Load())
		}
		if s.tele.Anzahl(supporttest.OpCountError) != 26 || !slices.Equal(gezaehlt, codes) {
			t.Errorf("Transport %d: Codes der Telemetrie %v bei %d CountError; erwartet 26: %v", i, gezaehlt, s.tele.Anzahl(supporttest.OpCountError), codes)
		}
		if a := s.route.anfragen; len(a) != 2 || !strings.Contains(a[0], "GET /probe ?") || !strings.Contains(a[1], "GET /probe/a/b ?") || !strings.Contains(a[1], "werte=map[id:a/b]") {
			t.Errorf("Transport %d: Anfragen an die Route %q; erwartet `GET /probe` und `GET /probe/a/b` mit dem Pfadwert id", i, a)
		}
	}
	if !slices.Equal(server[0].route.anfragen, server[1].route.anfragen) {
		t.Errorf("die Anfragen an die Route weichen über TLS ab:\nKlartext %q\nTLS      %q", server[0].route.anfragen, server[1].route.anfragen)
	}
}

// TestAdminHandshakeScheitert (AT13, M-Ausgang): ein Handshake, der scheitert, erreicht den Adapter nicht: kein Handler, keine ID, kein Eintrag, keine Zählung, keine
// Zeile in einer Senke, und ein Klartext-Request an den TLS-Port erhält nie die Antwort des Rahmens (SPEC-299 Nr. 10 b). Senke sind `stdout`, `stderr` und
// `log.Default()` des Prozesses. Menge: unbekannte CA, Klartext-Request an den TLS-Port, Version 1.1, allein `h2` = 4 Zellen.
// Das Benutzerhandbuch sagt im Abschnitt *TLS der Admin-Schnittstelle*, zu einem gescheiterten Handshake schreibe der Dienst keine Zeile; dieser Test
// prüft das an Adapter und Umhüllung, ohne den Composition Root.
func TestAdminHandshakeScheitert(t *testing.T) {
	w := neueWerkstatt(t)
	zahl := 0
	senke := ausgabeErfasst(t, func() {
		s := w.adminTLS(t, w.wurzel)
		unbekannt, version11, nurH2 := tlstest.NeueWurzel(t).Klient("localhost"), w.wurzel.Klient("localhost"), w.wurzel.Klient("localhost")
		version11.MinVersion, version11.MaxVersion, nurH2.NextProtos = tls.VersionTLS10, tls.VersionTLS11, []string{"h2"}
		faelle := []struct {
			name    string
			schluss int // die Zahl der Close-Aufrufe des Servers an der Verbindung: ein Klartext-Request wird von net/http beantwortet und zweimal geschlossen
			lauf    func() error
		}{
			{"unbekannte CA", 1, func() error {
				_, err := s.dial(unbekannt)
				if !errors.As(err, new(x509.UnknownAuthorityError)) {
					return fmt.Errorf("Fehler %v, erwartet x509.UnknownAuthorityError", err)
				}
				return nil
			}},
			{"Klartext-Request an den TLS-Port", 2, func() error {
				c, err := net.DialTimeout("tcp", s.adresse, 5*time.Second)
				if err != nil {
					return err
				}
				defer func() { _ = c.Close() }()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				_, _ = io.WriteString(c, "GET /healthz HTTP/1.1\r\nHost: admin\r\n\r\n")
				antwort, _ := io.ReadAll(c)
				if strings.Contains(strings.ToLower(string(antwort)), "x-correlation-id") || strings.Contains(string(antwort), `"ok"`) || strings.Contains(string(antwort), "ODR-") {
					return fmt.Errorf("die Antwort ist die des Rahmens: %q", antwort)
				}
				return nil
			}},
			{"Version 1.1", 1, func() error { return erwarteFehler(s.dial(version11)) }},
			{"allein h2", 1, func() error { return erwarteFehler(s.dial(nurH2)) }},
		}
		erwartet := 0
		for _, f := range faelle {
			zahl++
			if err := f.lauf(); err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
			erwartet += f.schluss
			s.inner.warte(t, erwartet)
			if got := s.inner.angenomm.Load(); got != int64(zahl) {
				t.Errorf("%s: %d angenommene Verbindungen, erwartet %d: die Zelle erreicht den Server nicht genau einmal", f.name, got, zahl)
			}
			if n, a, e, r := s.ids.n.Load(), s.audit.Aufrufe(), len(s.tele.Aufrufe()), s.route.zahl(); n+int64(a+e+r) != 0 {
				t.Errorf("%s: IDs %d, Audit %d, Telemetrie %d, Route %d; erwartet 0", f.name, n, a, e, r)
			}
		}
	})
	if senke != "" {
		t.Errorf("eine Senke trägt eine Zeile: %q", senke)
	}
	if zahl != 4 {
		t.Fatalf("Menge M-Ausgang hat %d Zellen, erwartet 4", zahl)
	}
}

// erwarteFehler meldet, wenn ein Handshake gelang.
func erwarteFehler(c *tls.Conn, err error) error {
	if err == nil {
		_ = c.Close()
		return errors.New("der Handshake gelang")
	}
	return nil
}

// TestAdminHaengenderHandshake (AT14, M-Haengend): ein Client, der verbindet und nichts sendet, hält weder Accept noch einen anderen Client auf (SPEC-299 Nr. 10 a).
// Menge: 1 und 8 stumme Verbindungen, angenommen vor dem zweiten Client; der Client erhält `GET /healthz` 200 binnen 5 s (Frist des Clients, kein Schlaf).
func TestAdminHaengenderHandshake(t *testing.T) {
	w := neueWerkstatt(t)
	zellen := 0
	for _, zahl := range []int{1, 8} {
		zellen++
		t.Run(fmt.Sprintf("%d stumme Verbindungen", zahl), func(t *testing.T) {
			s := w.adminTLS(t, w.wurzel)
			for range zahl {
				c, err := net.DialTimeout("tcp", s.adresse, 5*time.Second)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = c.Close() })
			}
			if a := s.hole(t, "/healthz", ""); a.status != 200 || a.body != `{"status":"ok"}` {
				t.Errorf("Status %d, Body %q; erwartet 200", a.status, a.body)
			}
		})
	}
	if zellen != 2 {
		t.Fatalf("Menge M-Haengend hat %d Zellen, erwartet 2", zellen)
	}
}

// adminHalt hält jeden Write der Verbindung nach dem ersten zurück, bis freigabe geschlossen ist; gehalten meldet den ersten gehaltenen Write.
type adminHalt struct {
	net.Conn
	writes   atomic.Int64
	gehalten chan struct{}
	freigabe chan struct{}
}

func (c *adminHalt) Write(b []byte) (int, error) {
	if c.writes.Add(1) > 1 {
		select {
		case c.gehalten <- struct{}{}:
		default:
		}
		select {
		case <-c.freigabe:
		case <-time.After(10 * time.Second):
			return 0, errors.New("Write nicht freigegeben")
		}
	}
	return c.Conn.Write(b)
}

// pruefe503 verlangt die Antwort eines angehaltenen Servers: 503 `ODR-E5010` mit `Connection: close` (`resp.Close`, der Leser nimmt den Header heraus) und
// `X-Correlation-Id`, und `soll` Aufrufe der Route.
func (s *adminServer) pruefe503(t *testing.T, was string, resp *http.Response, body string, soll int) {
	t.Helper()
	if resp.StatusCode != 503 || !strings.Contains(body, "ODR-E5010") || !resp.Close || resp.Header.Get("X-Correlation-Id") == "" || s.route.zahl() != soll {
		t.Errorf("%s: Status %d, Body %q, Close %v, ID %q, %d Aufrufe der Route; erwartet 503, ODR-E5010, close, eine ID und %d Aufrufe",
			was, resp.StatusCode, body, resp.Close, resp.Header.Get("X-Correlation-Id"), s.route.zahl(), soll)
	}
}

// TestAdminAnhaltenUeberTLS (AT15, M-Anhalten): nach Anhalten erhält ein Request über TLS 503 `ODR-E5010` mit `Connection: close` und `X-Correlation-Id`, auch wenn
// sein Handshake erst nach Anhalten endet (SPEC-299 Nr. 10 c). Menge: (i) eine fertige TLS-Verbindung sendet nach Anhalten einen Request, (ii) eine Verbindung,
// deren Handshake vor Anhalten begann und danach endet (Haltepunkt: eine Hülle um das `net.Conn` des Clients hält ihren letzten Flug bis nach Anhalten).
func TestAdminAnhaltenUeberTLS(t *testing.T) {
	w := neueWerkstatt(t)
	zellen := 0
	zelle := func(name string, f func(*testing.T)) {
		zellen++
		t.Run(name, f)
	}
	zelle("(i) fertige Verbindung", func(t *testing.T) {
		s := w.adminTLS(t, w.wurzel)
		c, err := s.dial(w.wurzel.Klient("localhost"))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = c.Close() }()
		br := bufio.NewReader(c)
		if resp, body := holeRoh(t, c, br, "/probe"); resp.StatusCode != 200 || body != `{"ok":true}` {
			t.Fatalf("vor Anhalten: Status %d, Body %q", resp.StatusCode, body)
		}
		s.srv.Anhalten()
		resp, body := holeRoh(t, c, br, "/probe")
		s.pruefe503(t, "nach Anhalten", resp, body, 1)
	})
	zelle("(ii) Handshake endet nach Anhalten", func(t *testing.T) {
		s := w.adminTLS(t, w.wurzel)
		raw, err := net.DialTimeout("tcp", s.adresse, 5*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		halt := &adminHalt{Conn: raw, gehalten: make(chan struct{}, 1), freigabe: make(chan struct{})}
		tc := tls.Client(halt, w.wurzel.Klient("localhost"))
		defer func() { _ = tc.Close() }()
		fertig := make(chan error, 1)
		go func() { fertig <- tc.HandshakeContext(t.Context()) }()
		select {
		case <-halt.gehalten:
		case <-time.After(5 * time.Second):
			t.Fatal("der Handshake erreicht den Haltepunkt nicht")
		}
		s.hole(t, "/healthz", "") // der zweite Client wird nach der gehaltenen Verbindung angenommen: sie ist dann angenommen
		s.srv.Anhalten()
		close(halt.freigabe)
		select {
		case err := <-fertig:
			if err != nil {
				t.Fatalf("Handshake nach Anhalten: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("der Handshake endet nicht")
		}
		resp, body := holeRoh(t, tc, bufio.NewReader(tc), "/probe")
		s.pruefe503(t, "nach Anhalten", resp, body, 0)
	})
	if zellen != 2 {
		t.Fatalf("Menge M-Anhalten hat %d Zellen, erwartet 2", zellen)
	}
}

// TestAdminSchliessenOffenerHandshake (AT16, M-Offen): Schliessen schließt eine Verbindung, deren Handshake nicht beendet ist, sofort und wartet nicht auf sie
// (SPEC-299 Nr. 10 c). Menge: eine stumme Verbindung, eine mit einem halben ClientHello (Kopf des Records und des Handshakes, 11 von 245 Bytes).
func TestAdminSchliessenOffenerHandshake(t *testing.T) {
	w := neueWerkstatt(t)
	halbes := []byte{0x16, 0x03, 0x01, 0x00, 0xf0, 0x01, 0x00, 0x00, 0xec, 0x03, 0x03}
	const frist = 4 * time.Second
	zellen := 0
	for _, f := range []struct {
		name   string
		senden []byte
	}{{"stumm", nil}, {"halbes ClientHello", halbes}} {
		zellen++
		t.Run(f.name, func(t *testing.T) {
			s := w.adminTLS(t, w.wurzel)
			c, err := net.DialTimeout("tcp", s.adresse, 5*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = c.Close() }()
			if f.senden != nil {
				if _, err := c.Write(f.senden); err != nil {
					t.Fatal(err)
				}
			}
			s.hole(t, "/healthz", "") // die Verbindung ist angenommen, wenn der zweite Client bedient ist
			ergebnis := make(chan error, 1)
			vorher := time.Now()
			go func() { ergebnis <- s.srv.Schliessen(frist) }()
			select {
			case err := <-ergebnis:
				if dauer := time.Since(vorher); err != nil || dauer > frist/2 {
					t.Errorf("Schliessen lieferte %v nach %v; erwartet nil binnen %v", err, dauer, frist/2)
				}
			case <-time.After(frist + 5*time.Second):
				t.Fatal("Schliessen kehrt nicht zurück")
			}
			_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, err := c.Read(make([]byte, 16)); err == nil || errors.Is(err, os.ErrDeadlineExceeded) {
				t.Errorf("Lesen nach Schliessen lieferte %v; erwartet eine geschlossene Verbindung", err)
			}
		})
	}
	if zellen != 2 {
		t.Fatalf("Menge M-Offen hat %d Zellen, erwartet 2", zellen)
	}
}

// TestAdminKeinNeuladen (AT17, M-Neuladen): der Start liest die Dateien einmal: ein Wechsel nach dem Laden ändert das gezeigte Zertifikat nicht, ein Neustart zeigt das
// neue; läuft das Zertifikat im Betrieb ab, bedient der Listener es weiter und schreibt dazu in keine Senke (SPEC-299 Nr. 11). Menge: Dateien überschrieben vor
// dem ersten Handshake, danach, und ein Zertifikat, das im Betrieb abläuft (geladen zu einer Zeit in seiner Gültigkeit, der Client mit einer Uhr davor) = 3 Zellen.
func TestAdminKeinNeuladen(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, erstes := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	zweiteWurzel, dritteWurzel := tlstest.NeueWurzel(t), tlstest.NeueWurzel(t)
	zweites, drittes := zweiteWurzel.Blatt(t, tlstest.Optionen{}), dritteWurzel.Blatt(t, tlstest.Optionen{})
	ueberschreibe := func(b *tlstest.Blatt) {
		t.Helper()
		if err := os.WriteFile(cert, b.ZertifikatPEM(), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(key, w.schluessel(b), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	gezeigt := func(s *adminServer, cfg *tls.Config) (*big.Int, error) {
		c, err := s.dial(cfg)
		if err != nil {
			return nil, err
		}
		defer func() { _ = c.Close() }()
		return c.ConnectionState().PeerCertificates[0].SerialNumber, nil
	}
	zellen := 0
	pruefeSerie := func(was string, s *adminServer, a *tlstest.Aussteller, soll *tlstest.Blatt) {
		t.Helper()
		if got, err := gezeigt(s, a.Klient("localhost")); err != nil || got.Cmp(soll.Zertifikat.SerialNumber) != 0 {
			t.Errorf("%s: Seriennummer %v, Fehler %v; erwartet %v", was, got, err, soll.Zertifikat.SerialNumber)
		}
	}
	tr := w.ladeAdmin(cert, key, w.jetzt)
	ueberschreibe(zweites)
	s := neuerAdminServer(t, tr, w.wurzel)
	zellen++
	pruefeSerie("Dateien vor dem ersten Handshake überschrieben", s, w.wurzel, erstes)
	ueberschreibe(drittes)
	zellen++
	pruefeSerie("Dateien nach dem ersten Handshake überschrieben", s, w.wurzel, erstes)
	for name, a := range map[string]*tlstest.Aussteller{"zweiten": zweiteWurzel, "dritten": dritteWurzel} {
		if _, err := gezeigt(s, a.Klient("localhost")); err == nil {
			t.Errorf("der Client mit der Wurzel des %s Blatts nahm das Zertifikat an", name)
		}
	}
	pruefeSerie("Neustart", neuerAdminServer(t, w.ladeAdmin(cert, key, w.jetzt), dritteWurzel), dritteWurzel, drittes)

	// Das Blatt ist seit fünf Minuten abgelaufen, die Wurzel gilt noch; der Start lud es vor 30 Minuten.
	alt := w.wurzel.Blatt(t, tlstest.Optionen{Von: w.jetzt.Add(-50 * time.Minute), Bis: w.jetzt.Add(-5 * time.Minute)})
	frueher := w.jetzt.Add(-30 * time.Minute)
	certAlt, keyAlt := w.datei(alt.ZertifikatPEM()), w.datei(w.schluessel(alt))
	uhrDavor := w.wurzel.Klient("localhost")
	uhrDavor.Time = func() time.Time { return frueher }
	zellen++
	// Erfasst sind der Start (Laden, Umhüllen, admin.New: ein dort an `stdout` oder `stderr` gebundener Melder trifft die Senke) und der Betrieb mit einem
	// angenommenen und einem vom Client abgelehnten Handshake, je bis zum Close des Servers. Außerhalb der Senke bleiben ein vor der Erfassung gebundener
	// Melder, ein Schreiben am Paket `os` vorbei und eine Meldung nach dem Ende der Erfassung.
	senke := ausgabeErfasst(t, func() {
		s := neuerAdminServer(t, w.ladeAdmin(certAlt, keyAlt, frueher), w.wurzel)
		c, err := s.dial(uhrDavor)
		if err != nil {
			t.Errorf("abgelaufenes Zertifikat, Uhr des Clients davor: %v", err)
			return
		}
		if got := c.ConnectionState().PeerCertificates[0].SerialNumber; got.Cmp(alt.Zertifikat.SerialNumber) != 0 {
			t.Errorf("abgelaufenes Zertifikat, Uhr des Clients davor: Seriennummer %v, erwartet %v", got, alt.Zertifikat.SerialNumber)
		}
		// Die Antwort belegt, dass der Server den Handshake beendet hat und die Verbindung bedient.
		if err := bedient(c); err != nil {
			t.Errorf("abgelaufenes Zertifikat: %v", err)
		}
		_ = c.Close()
		s.inner.warte(t, 1)
		var ungueltig x509.CertificateInvalidError
		if _, err := gezeigt(s, w.wurzel.Klient("localhost")); !errors.As(err, &ungueltig) || ungueltig.Reason != x509.Expired {
			t.Errorf("Client mit der Uhr des Hosts: Fehler %v, erwartet ein abgelaufenes Zertifikat", err)
		}
		s.inner.warte(t, 2)
		if a, e := s.audit.Aufrufe(), len(s.tele.Aufrufe()); a+e != 0 {
			t.Errorf("abgelaufenes Zertifikat: Audit %d, Telemetrie %d; erwartet 0", a, e)
		}
	})
	if senke != "" {
		t.Errorf("abgelaufenes Zertifikat: eine Senke trägt eine Zeile: %q", senke)
	}
	if zellen != 3 {
		t.Fatalf("Menge M-Neuladen hat %d Zellen, erwartet 3", zellen)
	}
}
