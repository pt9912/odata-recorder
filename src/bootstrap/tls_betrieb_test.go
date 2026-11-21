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
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// tlsKonfig ist eine gültige Konfiguration im Modus modus mit den Pfaden des TLS; Listen ist `127.0.0.1:0`, die Frist von Schliessen fünf Sekunden.
func tlsKonfig(modus, dir, cert, key string) bootstrap.Konfiguration {
	k := basisKonfig(dir, modus, dienstA(""))
	k.ListenTLSCert, k.ListenTLSKey = cert, key
	k.ShutdownTimeout = 5 * time.Second
	switch modus {
	case modusRecord:
		k.Dienste = []bootstrap.Dienst{dienstA("http://127.0.0.1:1/odata")}
	case modusMock, modusHybrid:
		k.PublicBaseURL, k.StrictCoverage = "https://127.0.0.1:1", true
	}
	return k
}

// tlsPaar legt Zertifikat und Schlüssel eines Blatts von a in w.dir ab und liefert die Pfade.
func (w *werkstatt) tlsPaar(a *tlstest.Aussteller, o tlstest.Optionen) (cert, key string, b *tlstest.Blatt) {
	b = a.Blatt(w.t, o)
	return w.datei(b.ZertifikatPEM()), w.datei(w.schluessel(b)), b
}

// hochgefahren baut und fährt hoch; der Test fährt am Ende herunter.
func hochgefahren(t *testing.T, k bootstrap.Konfiguration) *bootstrap.Handle {
	t.Helper()
	ctx := context.Background()
	if err := os.MkdirAll(k.DataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := h.Hochfahren(ctx); err != nil {
		_ = h.Herunterfahren(ctx)
		t.Fatalf("Hochfahren: %v", err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(ctx) })
	return h
}

// httpsClient ist ein Client, der HTTP/2 anbietet (`ForceAttemptHTTP2`), mit der Konfiguration cfg und ohne Wiederverwendung der Verbindung.
func httpsClient(cfg *tls.Config) *http.Client {
	return &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: cfg, ForceAttemptHTTP2: true, DisableKeepAlives: true}}
}

// antwortVon liest Status, Protokoll, Kopf und Body einer Antwort; ein Fehler der Anfrage beendet den Test.
func antwortVon(t *testing.T, c *http.Client, url string) (status, proto int, kopf http.Header, body string) {
	t.Helper()
	resp, err := c.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Body von %s: %v", url, err)
	}
	return resp.StatusCode, resp.ProtoMajor, resp.Header, string(b)
}

// pruefeUnbekannt verlangt die Antwort auf einen Pfad ohne Service: 404 mit ODR-E5005 (SPEC-136).
func pruefeUnbekannt(t *testing.T, was string, status int, body string, kopf http.Header) {
	t.Helper()
	if status != 404 || !strings.Contains(body, "ODR-E5005") || kopf.Get("X-Correlation-Id") == "" {
		t.Errorf("%s: Status %d, Body %q, X-Correlation-Id %q; erwartet 404, ODR-E5005 und eine ID", was, status, body, kopf.Get("X-Correlation-Id"))
	}
}

// TestBuildTLSVerstoss (B10): Build lädt das Paar vor dem Öffnen der Ablage, in jedem Modus; ein Verstoß ist ein *KonfigFehler mit einer Zeile je
// Schlüssel, und es entsteht keine Datenbank und kein Blob-Verzeichnis (SPEC-291 Nr. 8, SPEC-139 Nr. 2). Menge: 4 Modi × 2 Konfigurationen
// (Zertifikat nicht lesbar; Zertifikat außerhalb der Gültigkeit und Schlüssel passt nicht).
func TestBuildTLSVerstoss(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	faelle := []tlsFall{w.paar(n, zNichtLesbar, zGueltig), w.paar(n, zBesonders, zBesonders)}
	zahl := 0
	for _, modus := range vierModi() {
		for _, f := range faelle {
			zahl++
			dir := t.TempDir()
			h, err := bootstrap.Build(context.Background(), tlsKonfig(modus, dir, f.cert, f.key), nil)
			var soll []string
			for _, v := range f.erwartet {
				soll = append(soll, zeile(v.Schluessel, v.Grund))
			}
			var kf *bootstrap.KonfigFehler
			if h != nil || !errors.As(err, &kf) || !slices.Equal(kf.Zeilen, soll) || bootstrap.Meldung(err) != strings.Join(soll, "\n") {
				t.Errorf("%s, %s: Handle %v, Fehler %v; erwartet ein *KonfigFehler mit %q", modus, f.name, h != nil, err, soll)
			}
			for _, name := range []string{"odata-recorder.db", "blobs"} {
				if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
					t.Errorf("%s, %s: %s entstand", modus, f.name, name)
				}
			}
		}
	}
	if zahl != 8 {
		t.Errorf("Menge hat %d Elemente, erwartet 8", zahl)
	}
}

// TestBuildTLSNurEinPfad: Build mit genau einem Pfad (nur als Literal der Konfiguration möglich) meldet den fehlenden Schlüssel wie das
// Laden (SPEC-291 Nr. 2 Rang 1) und öffnet nichts (SPEC-291 Nr. 8).
func TestBuildTLSNurEinPfad(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	for _, f := range []struct{ name, cert, key, soll string }{
		{"nur das Zertifikat", cert, "", zeile(listenKey, gErforderlich(listenCert))},
		{"nur der Schlüssel", "", key, zeile(listenCert, gErforderlich(listenKey))},
	} {
		h, err := bootstrap.Build(context.Background(), tlsKonfig(modusReplay, t.TempDir(), f.cert, f.key), nil)
		var kf *bootstrap.KonfigFehler
		if h != nil || !errors.As(err, &kf) || !slices.Equal(kf.Zeilen, []string{f.soll}) {
			t.Errorf("%s: Handle %v, Fehler %v; erwartet %q", f.name, h != nil, err, f.soll)
		}
	}
}

// TestHTTPSEndeZuEnde (B11): Hochfahren umhüllt den Listener genau dann, wenn beide Pfade gesetzt sind; die Antwort kommt über HTTPS wie sonst,
// HTTP/1.1 auch bei einem Client, der HTTP/2 anbietet; ein zweiter Build auf denselben Dateien ist unabhängig (SPEC-291 Nr. 6, 8, 12). Menge:
// 4 Modi × {TLS, Neustart, ohne Schlüssel}.
func TestHTTPSEndeZuEnde(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	client := w.wurzel.Klient("localhost")
	zahl := 0
	for _, modus := range vierModi() {
		dir := filepath.Join(t.TempDir(), "d")
		for _, lauf := range []string{"TLS", "Neustart"} {
			zahl++
			h := hochgefahren(t, tlsKonfig(modus, dir, cert, key))
			status, proto, kopf, body := antwortVon(t, httpsClient(client), "https://"+h.Adresse()+"/unbekannt")
			pruefeUnbekannt(t, modus+" "+lauf, status, body, kopf)
			if proto != 1 {
				t.Errorf("%s %s: ProtoMajor %d, erwartet 1 (der Client bot HTTP/2 an)", modus, lauf, proto)
			}
			if err := h.Herunterfahren(context.Background()); err != nil {
				t.Errorf("%s %s: Herunterfahren: %v", modus, lauf, err)
			}
		}
		zahl++
		h := hochgefahren(t, tlsKonfig(modus, dir, "", ""))
		status, _, kopf, body := antwortVon(t, httpsClient(nil), "http://"+h.Adresse()+"/unbekannt")
		pruefeUnbekannt(t, modus+" ohne Schlüssel", status, body, kopf)
		if _, err := httpsClient(client).Get("https://" + h.Adresse() + "/unbekannt"); err == nil {
			t.Errorf("%s ohne Schlüssel: ein TLS-Client erhielt eine Antwort", modus)
		}
	}
	if zahl != 12 {
		t.Errorf("Menge hat %d Elemente, erwartet 12", zahl)
	}
}

// TestHTTPSClientFehler (B12): Zertifikatsfehler sind am Client sichtbar und erreichen den Dienst nicht (LH-RB-02.a; SPEC-291 Nr. 9 b). Menge H4,
// 4 Elemente: Client ohne die CA im Pool, Blatt ohne den Namen des Hosts, Klartext-Client am TLS-Port, TLS-Client am Klartext-Port.
func TestHTTPSClientFehler(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	fremdCert, fremdKey, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{Name: "anderer-host.test"})
	tlsServer := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), cert, key))
	nameServer := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), fremdCert, fremdKey))
	klartext := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), "", ""))
	zahl := 0

	zahl++
	_, err := httpsClient(&tls.Config{ServerName: "localhost", RootCAs: x509.NewCertPool()}).Get("https://" + tlsServer.Adresse() + "/unbekannt")
	var unbekannt x509.UnknownAuthorityError
	if !errors.As(err, &unbekannt) {
		t.Errorf("Client ohne die CA: Fehler %v, erwartet x509.UnknownAuthorityError", err)
	}

	zahl++
	_, err = httpsClient(w.wurzel.Klient("localhost")).Get("https://" + nameServer.Adresse() + "/unbekannt")
	var name x509.HostnameError
	if !errors.As(err, &name) {
		t.Errorf("Blatt ohne den Namen des Hosts: Fehler %v, erwartet x509.HostnameError", err)
	}

	zahl++
	resp, err := http.Get("http://" + tlsServer.Adresse() + "/unbekannt")
	if err == nil {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if strings.Contains(string(b), "ODR-") || resp.Header.Get("X-Correlation-Id") != "" {
			t.Errorf("Klartext-Client am TLS-Port: Antwort des Dienstes (Status %d, Body %q, ID %q)", resp.StatusCode, b, resp.Header.Get("X-Correlation-Id"))
		}
	}

	zahl++
	_, err = httpsClient(w.wurzel.Klient("localhost")).Get("https://" + klartext.Adresse() + "/unbekannt")
	if err == nil || !strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
		t.Errorf("TLS-Client am Klartext-Port: Fehler %v, erwartet der Fehler des Handshakes `server gave HTTP response to HTTPS client`", err)
	}
	if zahl != 4 {
		t.Errorf("Menge H4 hat %d Elemente, erwartet 4", zahl)
	}
}

// TestHaengenderHandshake (B15): ein Handshake, der hängt, hält Accept und andere Clients nicht auf (SPEC-291 Nr. 9 a). Eine TCP-Verbindung ohne
// Byte bleibt offen, ein zweiter Client mit gültigem Handshake erhält binnen 5 s die Antwort. Wirkungs-Fänger über 20 Wiederholungen, jede muss grün
// sein, je mit eigenem Root, eigenem Listener und eigener hängender Verbindung (parallel, damit eine Mutation nicht 20 mal 5 s kostet).
func TestHaengenderHandshake(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	for i := range 20 {
		t.Run(fmt.Sprintf("Wiederholung-%02d", i), func(t *testing.T) {
			t.Parallel()
			h := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), cert, key))
			haengt, err := net.Dial("tcp", h.Adresse())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = haengt.Close() }()
			c := httpsClient(w.wurzel.Klient("localhost"))
			c.Timeout = 5 * time.Second
			status, _, kopf, body := antwortVon(t, c, "https://"+h.Adresse()+"/unbekannt")
			pruefeUnbekannt(t, "zweiter Client neben der hängenden Verbindung", status, body, kopf)
		})
	}
}

// TestHerunterfahrenMitTLS (B18): Herunterfahren des Roots gilt mit TLS: der Listener ist danach zu, jede Verbindung geschlossen, binnen
// SHUTDOWN_TIMEOUT (SPEC-291 Nr. 9 c, SPEC-139 Nr. 7). Modus mock: eine HTTPS-Verbindung im Leerlauf und ein Handshake in der Schwebe.
func TestHerunterfahrenMitTLS(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	k := tlsKonfig(modusMock, t.TempDir(), cert, key)
	ctx := context.Background()
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatalf("Hochfahren: %v", err)
	}
	leerlauf, err := tls.Dial("tcp", h.Adresse(), w.wurzel.Klient("localhost"))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer func() { _ = leerlauf.Close() }()
	if _, err := io.WriteString(leerlauf, "GET /unbekannt HTTP/1.1\r\nHost: h\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(leerlauf)
	resp, err := http.ReadResponse(r, &http.Request{Method: "GET"})
	if err != nil {
		t.Fatalf("Antwort: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	schwebe, err := net.Dial("tcp", h.Adresse())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = schwebe.Close() }()

	vorher := time.Now()
	fertig := make(chan error, 1)
	go func() { fertig <- h.Herunterfahren(context.Background()) }()
	select {
	case err := <-fertig:
		if err != nil {
			t.Errorf("Herunterfahren: %v", err)
		}
	case <-time.After(k.ShutdownTimeout + 5*time.Second):
		t.Fatal("Herunterfahren kehrt nicht zurück")
	}
	if dauer := time.Since(vorher); dauer >= k.ShutdownTimeout {
		t.Errorf("Herunterfahren dauerte %v, erwartet unter %v", dauer, k.ShutdownTimeout)
	}
	if c, err := net.DialTimeout("tcp", h.Adresse(), time.Second); err == nil {
		_ = c.Close()
		t.Error("nach Herunterfahren gelingt ein Dial")
	}
	for name, c := range map[string]net.Conn{"Leerlauf": leerlauf, "Schwebe": schwebe} {
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, err := c.Read(make([]byte, 16)); err == nil || errors.Is(err, os.ErrDeadlineExceeded) {
			t.Errorf("Verbindung %s nach Herunterfahren: Fehler %v, erwartet eine geschlossene Verbindung", name, err)
		}
	}
}

// TestKeinNeuladen (B20): nach Build überschriebene Dateien ändern das gelieferte Zertifikat nicht (SPEC-291 Nr. 10, 12), weder im Fenster zwischen
// Build und Hochfahren noch danach: ein neuer Handshake zeigt die Seriennummer des ersten Blatts, der Client mit der Wurzel des zweiten und des dritten
// scheitert; ein Neustart (zweiter Build) liefert das dritte.
func TestKeinNeuladen(t *testing.T) {
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
	dir := filepath.Join(t.TempDir(), "d")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	h, err := bootstrap.Build(ctx, tlsKonfig(modusReplay, dir, cert, key), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(ctx) })
	ueberschreibe(zweites)
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatalf("Hochfahren: %v", err)
	}
	ueberschreibe(drittes)
	gezeigt := func(a string, cfg *tls.Config) (*big.Int, error) {
		c, err := tls.Dial("tcp", a, cfg)
		if err != nil {
			return nil, err
		}
		defer func() { _ = c.Close() }()
		return c.ConnectionState().PeerCertificates[0].SerialNumber, nil
	}
	if got, err := gezeigt(h.Adresse(), w.wurzel.Klient("localhost")); err != nil || got.Cmp(erstes.Zertifikat.SerialNumber) != 0 {
		t.Errorf("neuer Handshake: Seriennummer %v, Fehler %v; erwartet die des ersten Blatts %v", got, err, erstes.Zertifikat.SerialNumber)
	}
	for name, a := range map[string]*tlstest.Aussteller{"zweiten": zweiteWurzel, "dritten": dritteWurzel} {
		if _, err := gezeigt(h.Adresse(), a.Klient("localhost")); err == nil {
			t.Errorf("der Client mit der Wurzel des %s Blatts nahm das Zertifikat an", name)
		}
	}
	if err := h.Herunterfahren(context.Background()); err != nil {
		t.Fatal(err)
	}
	h2 := hochgefahren(t, tlsKonfig(modusReplay, dir, cert, key))
	if got, err := gezeigt(h2.Adresse(), dritteWurzel.Klient("localhost")); err != nil || got.Cmp(drittes.Zertifikat.SerialNumber) != 0 {
		t.Errorf("nach dem Neustart: Seriennummer %v, Fehler %v; erwartet die des dritten Blatts", got, err)
	}
}

// TestAbgelaufenWirdBedient (B20): läuft das Zertifikat im Betrieb ab, bedient der Listener es weiter und meldet nichts; der Client lehnt es
// ab (SPEC-291 Nr. 10).
func TestAbgelaufenWirdBedient(t *testing.T) {
	w := neueWerkstatt(t)
	bis := time.Now().Add(2 * time.Second).Truncate(time.Second).Add(time.Second)
	cert, key, blatt := w.tlsPaar(w.wurzel, tlstest.Optionen{Von: time.Now().Add(-time.Hour), Bis: bis})
	h := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), cert, key))
	time.Sleep(time.Until(blatt.Zertifikat.NotAfter) + 1500*time.Millisecond)
	if time.Now().Before(blatt.Zertifikat.NotAfter) {
		t.Fatal("das Zertifikat ist noch nicht abgelaufen")
	}
	ablauf := blatt.Zertifikat.NotAfter
	gutgeschrieben := w.wurzel.Klient("localhost")
	gutgeschrieben.Time = func() time.Time { return ablauf.Add(-time.Second) }
	status, _, kopf, body := antwortVon(t, httpsClient(gutgeschrieben), "https://"+h.Adresse()+"/unbekannt")
	pruefeUnbekannt(t, "abgelaufenes Zertifikat, Uhr des Clients davor", status, body, kopf)
	_, err := httpsClient(w.wurzel.Klient("localhost")).Get("https://" + h.Adresse() + "/unbekannt")
	var ungueltig x509.CertificateInvalidError
	if !errors.As(err, &ungueltig) || ungueltig.Reason != x509.Expired {
		t.Errorf("Client mit der Uhr des Hosts: Fehler %v, erwartet ein abgelaufenes Zertifikat", err)
	}
}

// tlsBedient meldet, ob der Listener TLS spricht und einem Client mit der Wurzel a das Blatt mit der Seriennummer serie zeigt: ein Handshake am
// Listener, den der Root dem Adapter übergab (der Fake des Servers nimmt nicht an, der Test tut es).
func tlsBedient(t *testing.T, l net.Listener, a *tlstest.Aussteller, serie *big.Int) bool {
	t.Helper()
	gesehen := make(chan *big.Int, 1)
	go func() {
		c, err := tls.Dial("tcp", l.Addr().String(), a.Klient("localhost"))
		if err != nil {
			gesehen <- nil
			return
		}
		defer func() { _ = c.Close() }()
		gesehen <- c.ConnectionState().PeerCertificates[0].SerialNumber
	}()
	sc, err := l.Accept()
	if err != nil {
		t.Errorf("Accept: %v", err)
		return false
	}
	defer func() { _ = sc.Close() }()
	tc, ok := sc.(*tls.Conn)
	if !ok {
		return false
	}
	_ = tc.SetDeadline(time.Now().Add(5 * time.Second))
	_ = tc.Handshake()
	select {
	case s := <-gesehen:
		return s != nil && s.Cmp(serie) == 0
	case <-time.After(5 * time.Second):
		return false
	}
}

// tlsVerbraucher legt ein Paar in dir ab und liefert die Pfade und die Prüfung, ob jeder der Listener dieses Blatt zeigt (tlsBedient).
func tlsVerbraucher(t *testing.T, dir string) (cert, key string, bedient func(...net.Listener) bool) {
	t.Helper()
	wurzel := tlstest.NeueWurzel(t)
	blatt := wurzel.Blatt(t, tlstest.Optionen{})
	cert = tlstest.Datei(t, dir, "tls-cert.pem", blatt.ZertifikatPEM(), 0o600)
	key = tlstest.Datei(t, dir, "tls-key.pem", blatt.SchluesselPEM(t, tlstest.PKCS8), 0o600)
	return cert, key, func(ls ...net.Listener) bool {
		return slices.IndexFunc(ls, func(l net.Listener) bool { return !tlsBedient(t, l, wurzel, blatt.Zertifikat.SerialNumber) }) < 0
	}
}

// wartetBis wartet höchstens 10 s, bis ok wahr ist.
func wartetBis(ok func() bool) {
	for ende := time.Now().Add(10 * time.Second); !ok() && time.Now().Before(ende); {
		time.Sleep(5 * time.Millisecond)
	}
}

// TestStartTLSKeinGeheimnis (B19): der Start nennt in keiner Zeile von stderr und keinem Logeintrag einen Pfad, Bytes der Dateien oder das Subject
// des Zertifikats (SPEC-291 Nr. 11). Menge: ein gültiges Paar (Modus replay, bis `start: bereit`) und die 8 Verstöße von B02.
func TestStartTLSKeinGeheimnis(t *testing.T) {
	w := neueWerkstatt(t)
	n := tlsNamen{listenCert, listenKey}
	gut := w.paar(n, zGueltig, zGueltig)
	verstoesse := []tlsFall{
		w.paar(n, zLeer, zGueltig), w.paar(n, zGueltig, zLeer), w.paar(n, zNichtLesbar, zGueltig), w.paar(n, zGueltig, zNichtLesbar),
		w.paar(n, zKeinPEM, zGueltig), w.paar(n, zGueltig, zKeinPEM), w.paar(n, zBesonders, zGueltig), w.paar(n, zGueltig, zBesonders),
	}
	if len(verstoesse) != 8 {
		t.Fatalf("Menge hat %d Verstöße, erwartet 8", len(verstoesse))
	}
	verboten := []string{"geheim-pfad-4711", "geheim-cn-4711", w.dir, pem2(w.schluessel(w.gut)), pem2(w.gut.ZertifikatPEM()), w.gut.Zertifikat.SerialNumber.String()}
	for _, f := range append([]tlsFall{gut}, verstoesse...) {
		env := []string{"MODE=replay", "DATA_DIR=" + filepath.Join(t.TempDir(), "d"), "SERVICE_ID=a"}
		if f.cert != "" {
			env = append(env, listenCert+"="+f.cert)
		}
		if f.key != "" {
			env = append(env, listenKey+"="+f.key)
		}
		b := bootstrap.StandardBauteile()
		echt := b.Listen
		b.Listen = func(ctx context.Context, network, _ string) (net.Listener, error) {
			return echt(ctx, network, "127.0.0.1:0")
		}
		ctx, cancel := context.WithCancel(context.Background())
		var stderr syncPuffer
		code := make(chan int, 1)
		go func() { code <- bootstrap.StartMit(ctx, env, &stderr, b) }()
		gueltig := len(f.erwartet) == 0
		if gueltig {
			wartetBis(func() bool { return strings.Contains(stderr.String(), "start: bereit") })
			cancel()
		}
		var got int
		select {
		case got = <-code:
		case <-time.After(20 * time.Second):
			t.Fatalf("%s: Start kehrt nicht zurück", f.name)
		}
		cancel()
		want := 1
		if gueltig {
			want = 0
		}
		if bereit := strings.Contains(stderr.String(), "start: bereit"); got != want || bereit != gueltig {
			t.Errorf("%s: Rückgabe %d (erwartet %d), bereit %v; stderr:\n%s", f.name, got, want, bereit, stderr.String())
		}
		for _, v := range verboten {
			if strings.Contains(stderr.String(), v) {
				t.Errorf("%s: stderr nennt %q:\n%s", f.name, v, stderr.String())
			}
		}
	}
}
