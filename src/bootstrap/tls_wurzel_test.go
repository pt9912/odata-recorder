package bootstrap_test

import (
	"context"
	"crypto/tls"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// TestBuildTLSErsterSchritt (B10): `Build` lädt das Paar als ersten Schritt (SPEC-291 Nr. 8): ein Verstoß des TLS zusammen mit dem Verstoß jedes anderen Schritts, der
// einen Fehler liefern kann, meldet allein die Zeile des TLS. Menge: 8 Quellen (die vier Grundwerte, ein Pfad der Schwärzung, die Datei der CA, `Oeffne`, ein Schritt der
// Verdrahtung nach `Oeffne`) × 1 Verstoß des TLS; die Zählprüfung zählt die Rückgaben eines Fehlers in `pruefeGrundwerte`, `vorbereiten` und `build` (11).
func TestBuildTLSErsterSchritt(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	soll := []string{zeile(listenCert, gNichtLesbar)}
	faelle := []struct {
		name    string
		modus   string
		aendere func(k *bootstrap.Konfiguration, dir string)
	}{
		{"Modus nicht lieferbar", modusRecord, func(k *bootstrap.Konfiguration, _ string) { k.Modus = "unbekannt" }},
		{"keine Dienste", modusRecord, func(k *bootstrap.Konfiguration, _ string) { k.Dienste = nil }},
		{"DATA_DIR leer", modusRecord, func(k *bootstrap.Konfiguration, _ string) { k.DataDir = "" }},
		{"PublicBaseURL fehlt", modusMock, func(k *bootstrap.Konfiguration, _ string) { k.PublicBaseURL = "" }},
		{"Pfad der Schwärzung ungültig", modusRecord, func(k *bootstrap.Konfiguration, _ string) { k.RedactJSONPaths = []string{"kein Pfad"} }},
		{"Datei der CA nicht lesbar", modusRecord, func(k *bootstrap.Konfiguration, dir string) { k.UpstreamCAFile = filepath.Join(dir, "fehlt-ca.pem") }},
		{"Ablage nicht zu öffnen", modusRecord, func(k *bootstrap.Konfiguration, dir string) { k.DataDir = filepath.Join(dir, "gibt-es-nicht", "d") }},
		{"Verdrahtung nach der Ablage", modusRecord, func(_ *bootstrap.Konfiguration, dir string) {
			if err := os.WriteFile(filepath.Join(dir, "blobs"), []byte("Datei statt Verzeichnis"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}
	if len(faelle) != 8 {
		t.Fatalf("Menge hat %d Quellen, erwartet 8", len(faelle))
	}
	for _, f := range faelle {
		dir := t.TempDir()
		k := tlsKonfig(f.modus, dir, w.fehlt(), key)
		_ = cert
		f.aendere(&k, dir)
		h, err := bootstrap.Build(context.Background(), k, nil)
		var kf *bootstrap.KonfigFehler
		if h != nil || !errors.As(err, &kf) || !slices.Equal(kf.Zeilen, soll) {
			t.Errorf("%s mit Verstoß des TLS: Handle %v, Fehler %v; erwartet allein %q", f.name, h != nil, err, soll)
		}
	}
	// Ohne den Verstoß des TLS melden dieselben Quellen ihren eigenen Fehler (der Fänger führt wirklich einen zweiten Verstoß).
	for _, f := range faelle {
		dir := t.TempDir()
		k := tlsKonfig(f.modus, dir, cert, key)
		f.aendere(&k, dir)
		h, err := bootstrap.Build(context.Background(), k, nil)
		var kf *bootstrap.KonfigFehler
		if h != nil || err == nil || (errors.As(err, &kf) && slices.Equal(kf.Zeilen, soll)) {
			t.Errorf("%s ohne Verstoß des TLS: Handle %v, Fehler %v; erwartet den eigenen Fehler", f.name, h != nil, err)
		}
	}
	if n := rueckgabenMitFehler(t, "wurzel.go", "pruefeGrundwerte", "vorbereiten", "build"); n != 11 {
		t.Errorf("pruefeGrundwerte, vorbereiten und build geben %d Fehler zurück, erwartet 11 (4 Grundwerte, TLS, Grundwerte, Pfad der Schwärzung, CA, `Oeffne`, `vorbereiten`, `verdrahte`)", n)
	}
}

// rueckgabenMitFehler zählt die Rückgaben der Funktionen, deren letzter Wert nicht das Literal nil ist (go/parser).
func rueckgabenMitFehler(t *testing.T, datei string, funktionen ...string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || !slices.Contains(funktionen, fn.Name.Name) {
			continue
		}
		ast.Inspect(fn.Body, func(x ast.Node) bool {
			if r, ok := x.(*ast.ReturnStmt); ok && len(r.Results) > 0 {
				if id, ok := r.Results[len(r.Results)-1].(*ast.Ident); !ok || id.Name != "nil" {
					n++
				}
			}
			return true
		})
	}
	return n
}

// TestListenerKonfigAmRoot (B11): der Listener, den der Root öffnet, trägt das Config von `ladeTLS` (SPEC-291 Nr. 5): Mindestversion TLS 1.2, Höchstversion der Standard,
// ALPN allein `http/1.1`, die Kette der Datei, kein Client-Zertifikat, die Kurven und Cipher-Suiten des Standards, Sitzungstickets des Standards. Menge: jedes
// exportierte Feld von tls.Config ist einer Sonde zugeordnet oder steht mit Grund in der Negativliste (Zählprüfung gegen reflect); die Sonden sind Handshakes eines Clients.
func TestListenerKonfigAmRoot(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	zw := w.wurzel.Zwischen(t, tlstest.Optionen{})
	kette := zw.Blatt(t, tlstest.Optionen{})
	kCert, kKey := w.datei(kette.KettePEM()), w.datei(w.schluessel(kette))
	h := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), cert, key))
	hKette := hochgefahren(t, tlsKonfig(modusReplay, filepath.Join(t.TempDir(), "d"), kCert, kKey))
	klient := func(mod func(*tls.Config)) *tls.Config {
		c := w.wurzel.Klient("localhost")
		mod(c)
		return c
	}
	dial := func(a string, c *tls.Config) (tls.ConnectionState, error) {
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", a, c)
		if err != nil {
			return tls.ConnectionState{}, err
		}
		defer func() { _ = conn.Close() }()
		return conn.ConnectionState(), nil
	}
	sonden := map[string]bool{}
	ok := func(sonde string, gut bool, was string, args ...any) {
		t.Helper()
		sonden[sonde] = true
		if !gut {
			t.Errorf("Sonde "+sonde+": "+was, args...)
		}
	}
	for _, v := range []struct {
		name string
		wert uint16
		gut  bool
	}{{"TLS 1.0", tls.VersionTLS10, false}, {"TLS 1.1", tls.VersionTLS11, false}, {"TLS 1.2", tls.VersionTLS12, true}, {"TLS 1.3", tls.VersionTLS13, true}} {
		st, err := dial(h.Adresse(), klient(func(c *tls.Config) { c.MinVersion, c.MaxVersion = v.wert, v.wert }))
		ok("MinVersion", (err == nil) == v.gut && (!v.gut || st.Version == v.wert), "%s: Version %x, Fehler %v; erwartet Erfolg %v", v.name, st.Version, err, v.gut)
	}
	st13, err13 := dial(h.Adresse(), klient(func(*tls.Config) {}))
	ok("MaxVersion", err13 == nil && st13.Version == tls.VersionTLS13, "Client mit Standard: Version %x, Fehler %v; erwartet TLS 1.3", st13.Version, err13)
	for _, p := range []struct {
		angebot []string
		soll    string
		gut     bool
	}{{[]string{"h2"}, "", false}, {[]string{"http/1.1"}, "http/1.1", true}, {[]string{"h2", "http/1.1"}, "http/1.1", true}, {nil, "", true}} {
		st, err := dial(h.Adresse(), klient(func(c *tls.Config) { c.NextProtos = p.angebot }))
		ok("NextProtos", (err == nil) == p.gut && st.NegotiatedProtocol == p.soll, "Angebot %q: Protokoll %q, Fehler %v; erwartet %q, Erfolg %v", p.angebot, st.NegotiatedProtocol, err, p.soll, p.gut)
	}
	st, err := dial(hKette.Adresse(), klient(func(c *tls.Config) { c.RootCAs = zw.Klient("localhost").RootCAs }))
	ok("Certificates", err == nil && len(st.PeerCertificates) == 2, "Kette: %d Zertifikate, Fehler %v; erwartet 2 (Blatt und Zwischenzertifikat)", len(st.PeerCertificates), err)
	_, err = dial(h.Adresse(), klient(func(*tls.Config) {}))
	ok("ClientAuth", err == nil, "Client ohne Zertifikat: %v", err)
	for _, c := range []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384, tls.CurveP521} {
		_, err := dial(h.Adresse(), klient(func(k *tls.Config) { k.CurvePreferences = []tls.CurveID{c} }))
		ok("CurvePreferences", err == nil, "Client mit allein der Kurve %v: %v", c, err)
	}
	n := 0
	for _, s := range tls.CipherSuites() {
		if !strings.Contains(s.Name, "ECDHE_ECDSA") || !slices.Contains(s.SupportedVersions, uint16(tls.VersionTLS12)) {
			continue
		}
		n++
		st, err := dial(h.Adresse(), klient(func(k *tls.Config) { k.MaxVersion, k.CipherSuites = tls.VersionTLS12, []uint16{s.ID} }))
		ok("CipherSuites", err == nil && st.CipherSuite == s.ID, "Client mit allein %s: Suite %x, Fehler %v", s.Name, st.CipherSuite, err)
	}
	if n != 5 {
		t.Errorf("Zahl der Suiten ECDHE-ECDSA unter TLS 1.2: %d, erwartet 5", n)
	}
	cache := tls.NewLRUClientSessionCache(4)
	c := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: klient(func(k *tls.Config) { k.ClientSessionCache = cache }), DisableKeepAlives: true}}
	wiederaufgenommen := false
	for range 2 {
		resp, err := c.Get("https://" + h.Adresse() + "/unbekannt")
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		wiederaufgenommen = resp.TLS != nil && resp.TLS.DidResume
	}
	ok("SessionTicketsDisabled", wiederaufgenommen, "die zweite Verbindung desselben Clients nahm die Sitzung nicht wieder auf")

	// Zählprüfung gegen die Quelle: jedes exportierte Feld von tls.Config ist geführt (sonden) oder steht mit Grund in der Negativliste.
	ohneSonde := map[string]string{
		"Rand": "Zufallsquelle, keine Wirkung auf den Vertrag", "Time": "Uhr des Servers, nur für Zertifikate des Clients", "NameToCertificate": "veraltet, ohne Wirkung",
		"GetCertificate": "TestKeinNeuladen und TestAbgelaufenWirdBedient (Zertifikat unverändert)", "GetClientCertificate": "Feld des Clients",
		"GetConfigForClient": "TestAbgelaufenWirdBedient (Zertifikat je Handshake unverändert)", "VerifyPeerCertificate": "wirkt nur mit Client-Zertifikat, ClientAuth ist geführt",
		"VerifyConnection": "wirkt nur mit Client-Zertifikat, ClientAuth ist geführt", "RootCAs": "Feld des Clients", "ServerName": "Feld des Clients", "ClientCAs": "wirkt nur mit ClientAuth",
		"InsecureSkipVerify": "Feld des Clients", "PreferServerCipherSuites": "veraltet, ohne Wirkung", "SessionTicketKey": "Schlüssel der Tickets, SessionTicketsDisabled ist geführt",
		"ClientSessionCache": "Feld des Clients", "UnwrapSession": "ohne Wirkung ohne WrapSession", "WrapSession": "ohne Wirkung auf das Protokoll der Verbindung",
		"DynamicRecordSizingDisabled": "ohne Wirkung auf Inhalt und Protokoll", "Renegotiation": "Feld des Clients", "KeyLogWriter": "schreibt Geheimnisse in eine Datei, die ein Test nicht liest",
		"EncryptedClientHelloConfigList": "Feld des Clients", "EncryptedClientHelloRejectionVerify": "Feld des Clients", "EncryptedClientHelloKeys": "ECH, im Image des Starts nicht angeboten", "GetEncryptedClientHelloKeys": "ECH, im Image des Starts nicht angeboten",
	}
	typ := reflect.TypeFor[tls.Config]()
	gefuehrt, ohne := 0, 0
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		_, sonde := sonden[f.Name]
		_, grund := ohneSonde[f.Name]
		switch {
		case sonde == grund:
			t.Errorf("Feld %s von tls.Config: Sonde %v, Negativliste %v; erwartet genau eines", f.Name, sonde, grund)
		case sonde:
			gefuehrt++
		default:
			ohne++
		}
	}
	for name := range sonden {
		if _, da := typ.FieldByName(name); !da {
			t.Errorf("Sonde für das Feld %s, das tls.Config nicht trägt", name)
		}
	}
	t.Logf("Felder von tls.Config: %d geführt, %d in der Negativliste", gefuehrt, ohne)
}
