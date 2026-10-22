package httpupstream_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestMain setzt vor jedem Test ein erzeugtes Zertifikat S für 127.0.0.1 als einzigen Anker des
// Systems: SSL_CERT_FILE zeigt auf S, SSL_CERT_DIR auf ein leeres Verzeichnis. Die Bibliothek liest
// die Anker je Prozess einmal, beim ersten Gebrauch. Der Schlüssel von S liegt neben S
// (systemAnker).
func TestMain(m *testing.M) {
	os.Exit(mitSystemAnker(m))
}

func mitSystemAnker(m *testing.M) int {
	dir, err := os.MkdirTemp("", "anker")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	s, pemS := zertifikat(time.Now().Add(time.Hour), "127.0.0.1")
	der, err := x509.MarshalECPrivateKey(s.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		panic(err)
	}
	datei := filepath.Join(dir, "s.pem")
	leer := filepath.Join(dir, "leer")
	if err := os.WriteFile(datei, pemS, 0o600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "s.key"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), 0o600); err != nil {
		panic(err)
	}
	if err := os.Mkdir(leer, 0o700); err != nil {
		panic(err)
	}
	if err := os.Setenv("SSL_CERT_FILE", datei); err != nil {
		panic(err)
	}
	if err := os.Setenv("SSL_CERT_DIR", leer); err != nil {
		panic(err)
	}
	return m.Run()
}

// systemAnker lädt S, das Zertifikat, das TestMain zum Anker des Systems macht, samt Schlüssel.
func systemAnker(t *testing.T) *tls.Certificate {
	t.Helper()
	dir := filepath.Dir(os.Getenv("SSL_CERT_FILE"))
	s, err := tls.LoadX509KeyPair(filepath.Join(dir, "s.pem"), filepath.Join(dir, "s.key"))
	if err != nil {
		t.Fatal(err)
	}
	return &s
}

// zertifikat erzeugt ein selbst signiertes Zertifikat einer Stelle für die IP-Adressen, gültig bis
// nachAblauf; es liefert Schlüsselpaar und PEM.
func zertifikat(nachAblauf time.Time, ips ...string) (tls.Certificate, []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: "upstream-test"},
		NotBefore: nachAblauf.Add(-48 * time.Hour), NotAfter: nachAblauf,
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, ip := range ips {
		tmpl.IPAddresses = append(tmpl.IPAddresses, net.ParseIP(ip))
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func pemVon(c *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
}

// tlsServer startet einen TLS-Server mit cert; ist cert leer, den von httptest (für 127.0.0.1,
// ::1 und example.com, nicht für localhost). Der Handler antwortet 200 mit dem Protokoll.
func tlsServer(t *testing.T, cert *tls.Certificate, http2 bool) *httptest.Server {
	t.Helper()
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, r.Proto)
	}))
	s.EnableHTTP2 = http2
	if cert != nil {
		s.TLS = &tls.Config{Certificates: []tls.Certificate{*cert}}
	}
	s.StartTLS()
	t.Cleanup(s.Close)
	return s
}

// tlsAbruf ruft den Server über host mit Pruefen und CAPEM ab.
func tlsAbruf(t *testing.T, s *httptest.Server, host string, pruefen bool, capem []byte) (string, error) {
	t.Helper()
	c, err := httpupstream.New(httpupstream.Konfig{Frist: zeitgrenze, Pruefen: pruefen, CAPEM: capem})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(httpupstream.Transport(c).CloseIdleConnections)
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(s.URL, "https://"))
	ant, err := sende(t.Context(), t, c, mitGeheim(anfrage("https://"+net.JoinHostPort(host, port), "GET", "")))
	if err != nil {
		return "", err
	}
	b, err := liesAlles(t, ant.Body)
	if err != nil || ant.Status != 200 {
		t.Fatalf("Status %d, %v", ant.Status, err)
	}
	return string(b), nil
}

// tlsFehlerFaelle: jeder Fehler der Prüfung ist ErrUpstreamNichtErreichbar (A10, SPEC-120 Nr. 7).
func tlsFehlerFaelle() []fehlerFall {
	nicht := outbound.ErrUpstreamNichtErreichbar
	return []fehlerFall{
		{name: "TLS unbekannte Stelle", art: nicht, lauf: func(t *testing.T) error {
			_, err := tlsAbruf(t, tlsServer(t, nil, false), "127.0.0.1", true, nil)
			return err
		}},
		{name: "TLS Hostname", art: nicht, lauf: func(t *testing.T) error {
			a := tlsServer(t, nil, false)
			_, err := tlsAbruf(t, a, "localhost", true, pemVon(a.Certificate()))
			return err
		}},
		{name: "TLS abgelaufen", art: nicht, lauf: func(t *testing.T) error {
			e, pemE := zertifikat(time.Now().Add(-time.Hour), "127.0.0.1")
			_, err := tlsAbruf(t, tlsServer(t, &e, false), "127.0.0.1", true, pemE)
			return err
		}},
	}
}

// TestTLS: Kette und Hostname gegen die Anker des Systems und CAPEM, CAPEM ergänzt; Pruefen=false
// prüft nichts; jeder Fehler der Prüfung ist ErrUpstreamNichtErreichbar (A10, SPEC-120 Nr. 1, 7).
func TestTLS(t *testing.T) {
	fahre(t, tlsFehlerFaelle())
	a := tlsServer(t, nil, false)
	s := tlsServer(t, systemAnker(t), false)
	e, _ := zertifikat(time.Now().Add(-time.Hour), "127.0.0.1")
	abgelaufen := tlsServer(t, &e, false)
	pemA := pemVon(a.Certificate())
	for _, f := range []struct {
		name    string
		server  *httptest.Server
		host    string
		pruefen bool
		capem   []byte
	}{
		{"A als CAPEM gegen A", a, "127.0.0.1", true, pemA},
		{"ohne CAPEM gegen S", s, "127.0.0.1", true, nil},
		{"A als CAPEM gegen S", s, "127.0.0.1", true, pemA},
		{"ohne Prüfung gegen A", a, "127.0.0.1", false, nil},
		{"ohne Prüfung über localhost", a, "localhost", false, nil},
		{"ohne Prüfung gegen E", abgelaufen, "127.0.0.1", false, nil},
		{"ohne Prüfung, A als CAPEM, gegen E", abgelaufen, "127.0.0.1", false, pemA},
	} {
		t.Run(f.name, func(t *testing.T) {
			if _, err := tlsAbruf(t, f.server, f.host, f.pruefen, f.capem); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestTransportTLS: der Client spricht HTTP/1.1 auch mit einem Server, der HTTP/2 anbietet (A11,
// SPEC-120 Nr. 7).
func TestTransportTLS(t *testing.T) {
	s := tlsServer(t, nil, true)
	proto, err := tlsAbruf(t, s, "127.0.0.1", true, pemVon(s.Certificate()))
	if err != nil || proto != "HTTP/1.1" {
		t.Errorf("Protokoll %q, %v", proto, err)
	}
}

// TestNewCAPEM: CAPEM braucht mindestens einen Block CERTIFICATE mit gültigem Inhalt, auch bei
// Pruefen=false; jedes lesbare Zertifikat ist Anker, auch eines vor einem Block ohne Wirkung; Text
// und andere Blöcke daneben bleiben ohne Wirkung, ein Block mit PEM-Headern auch bei gültigem
// Inhalt (A1, SPEC-120 Nr. 1).
func TestNewCAPEM(t *testing.T) {
	a := tlsServer(t, nil, false)
	pemA := pemVon(a.Certificate())
	schluessel := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{1, 2, 3}})
	kaputt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1, 2, 3}})
	mitHeader := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Headers: map[string]string{"Comment": "x"}, Bytes: a.Certificate().Raw})
	fremd := pem.EncodeToMemory(&pem.Block{Type: "X509 CERTIFICATE", Bytes: a.Certificate().Raw})
	for _, capem := range [][]byte{[]byte("kein PEM"), schluessel, kaputt, mitHeader, fremd} {
		for _, pruefen := range []bool{true, false} {
			c, err := httpupstream.New(httpupstream.Konfig{Frist: time.Second, Pruefen: pruefen, CAPEM: capem})
			if !errors.Is(err, model.ErrUngueltig) || c != nil {
				t.Errorf("CAPEM %q, Pruefen %v: %v, Client %v", capem, pruefen, err, c != nil)
			}
		}
	}
	for _, f := range []struct {
		name  string
		capem []byte
	}{
		{"allein", pemA},
		{"zwischen Text", append(append([]byte("vorher\n"), pemA...), "nachher\n"...)},
		{"Schlüssel davor", append(append([]byte{}, schluessel...), pemA...)},
		{"ungültiger Block davor", append(append([]byte{}, kaputt...), pemA...)},
		{"ungültiger Block danach", append(append([]byte{}, pemA...), kaputt...)},
		{"Schlüssel danach", append(append([]byte{}, pemA...), schluessel...)},
	} {
		t.Run(f.name, func(t *testing.T) {
			if _, err := tlsAbruf(t, a, "127.0.0.1", true, f.capem); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("drei Zertifikate", func(t *testing.T) {
		zertB, pemB := zertifikat(time.Now().Add(time.Hour), "127.0.0.1")
		zertC, pemC := zertifikat(time.Now().Add(time.Hour), "127.0.0.1")
		capem := slices.Concat(pemA, pemB, pemC)
		for _, f := range []struct {
			name   string
			server *httptest.Server
		}{
			{"A, B, C gegen A", a},
			{"A, B, C gegen B", tlsServer(t, &zertB, false)},
			{"A, B, C gegen C", tlsServer(t, &zertC, false)},
		} {
			t.Run(f.name, func(t *testing.T) {
				if _, err := tlsAbruf(t, f.server, "127.0.0.1", true, capem); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("Block mit Headern (A) vor E", func(t *testing.T) {
		_, pemE := zertifikat(time.Now().Add(-time.Hour), "127.0.0.1")
		_, err := tlsAbruf(t, a, "127.0.0.1", true, append(append([]byte{}, mitHeader...), pemE...))
		pruefeArt(t, err, outbound.ErrUpstreamNichtErreichbar)
	})
	for _, capem := range [][]byte{nil, {}} {
		t.Run(fmt.Sprintf("CAPEM %#v", capem), func(t *testing.T) {
			if _, err := tlsAbruf(t, tlsServer(t, systemAnker(t), false), "127.0.0.1", true, capem); err != nil {
				t.Fatal(err)
			}
		})
	}
}
