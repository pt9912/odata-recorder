package bootstrap_test

import (
	"io"
	"log"
	"net"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// Start-Prüfung und Erreichbarkeit des Admin-Listeners (SPEC-041 Nr. 8, LH-QA-04, SPEC-299 Nr. 3, 6, 8 c): der Test erreicht die Funktionen über die Naht
// in export_test.go und importiert admin nicht. Alle Namen dieser Datei tragen das Präfix `sicherheit`.

// sicherheitGrund ist der feste Grund des Verstoßes am Schlüssel ADMIN_LISTEN (SPEC-299 Nr. 6), ohne den Wert von ADMIN_LISTEN.
const sicherheitGrund = "außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY"

// sicherheitStartZelle ist eine Zelle der Menge M-Start: Host, ob ein gültiges Paar gesetzt ist, das Flag ADMIN_ALLOW_PLAINTEXT und das Orakel (Weg,
// ob genau der eine Verstoß am Schlüssel ADMIN_LISTEN gemeldet wird).
type sicherheitStartZelle struct {
	host          string
	paar, flag    bool
	weg           bootstrap.AdminWeg
	verstossListe bool
}

// sicherheitStartMenge sind die 12 Zellen aus 3 Hosts, 2 Paaren und 2 Flags mit ihrem Orakel nach der Tabelle von SPEC-299 Nr. 3: nicht-loopback ohne Paar
// und ohne Flag verweigert den Start, das Flag öffnet allein wegKlartext, das Paar allein wegTLS (vor Loopback), Loopback ohne Paar ist wegLoopback.
func sicherheitStartMenge() []sicherheitStartZelle {
	const (
		kein, tlsW, loop, klar = bootstrap.WegKeiner, bootstrap.WegTLS, bootstrap.WegLoopback, bootstrap.WegKlartext
		v4, v6, lo             = "0.0.0.0:8081", "[::]:8081", "127.0.0.1:8081"
	)
	return []sicherheitStartZelle{
		{v4, false, false, kein, true}, {v4, false, true, klar, false}, {v4, true, false, tlsW, false}, {v4, true, true, tlsW, false},
		{v6, false, false, kein, true}, {v6, false, true, klar, false}, {v6, true, false, tlsW, false}, {v6, true, true, tlsW, false},
		{lo, false, false, loop, false}, {lo, false, true, loop, false}, {lo, true, false, tlsW, false}, {lo, true, true, tlsW, false},
	}
}

// sicherheitPaar legt ein gültiges Paar aus Zertifikat und Schlüssel zur Laufzeit an (kein Schlüssel liegt im Repo) und liefert die zwei Pfade und die
// Wurzel, der die Zertifikate gehören.
func sicherheitPaar(t *testing.T) (cert, key string, wurzel *tlstest.Aussteller) {
	t.Helper()
	wurzel = tlstest.NeueWurzel(t)
	blatt := wurzel.Blatt(t, tlstest.Optionen{})
	dir := t.TempDir()
	cert = tlstest.Datei(t, dir, "cert.pem", blatt.ZertifikatPEM(), 0o600)
	key = tlstest.Datei(t, dir, "key.pem", blatt.SchluesselPEM(t, tlstest.PKCS8), 0o600)
	return cert, key, wurzel
}

// TestSicherheitStartPruefung (TS07, M-Start): ein nicht-loopback ADMIN_LISTEN ohne Paar und ohne Flag liefert genau einen Verstoß am Schlüssel ADMIN_LISTEN
// mit dem festen Grund und den Weg wegKeiner; das Flag öffnet allein wegKlartext ohne Verstoß, ein gültiges Paar allein wegTLS, Loopback ohne Paar
// wegLoopback.
func TestSicherheitStartPruefung(t *testing.T) {
	cert, key, _ := sicherheitPaar(t)
	menge := sicherheitStartMenge()
	if len(menge) != 12 {
		t.Fatalf("M-Start hat %d Zellen, erwartet 12", len(menge))
	}
	for _, z := range menge {
		c, k := "", ""
		if z.paar {
			c, k = cert, key
		}
		e := bootstrap.AdminEingabe{Listen: z.host, CertPfad: c, KeyPfad: k, Klartext: z.flag, FormGueltig: true}
		tr, vs := bootstrap.PruefeAdminTransport(e, time.Now())
		var soll []bootstrap.TLSVerstoss
		if z.verstossListe {
			soll = []bootstrap.TLSVerstoss{{Schluessel: "ADMIN_LISTEN", Grund: sicherheitGrund}}
		}
		if tr.Weg != z.weg || (tr.TLS != nil) != (z.weg == bootstrap.WegTLS) || !slices.Equal(vs, soll) {
			t.Errorf("Host %s, Paar %v, Flag %v: Weg %d, TLS %v, Verstöße %v; erwartet Weg %d, TLS %v, Verstöße %v",
				z.host, z.paar, z.flag, tr.Weg, tr.TLS != nil, vs, z.weg, z.weg == bootstrap.WegTLS, soll)
		}
	}
}

// sicherheitFrage fragt url mit dem Client an und liefert den Status der Antwort, oder 0 und den Fehler des Clients, wenn keine Antwort kommt.
func sicherheitFrage(t *testing.T, url string, client *http.Client) (int, error) {
	t.Helper()
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

// TestSicherheitErreichbar (TS08, M-Erreichbar): ein Listener auf 127.0.0.1:0 geht durch umhuelle, ein kleiner HTTP-Server ohne Bezug zu admin bedient ihn.
// wegLoopback: der Klartext-Client erhält 200, der TLS-Client scheitert im Handshake; wegTLS: der TLS-Client erhält 200, der Klartext-Client keine Antwort 200.
func TestSicherheitErreichbar(t *testing.T) {
	cert, key, wurzel := sicherheitPaar(t)
	zellen := 0
	for _, weg := range []bootstrap.AdminWeg{bootstrap.WegLoopback, bootstrap.WegTLS} {
		e := bootstrap.AdminEingabe{Listen: "127.0.0.1:8081", FormGueltig: true}
		if weg == bootstrap.WegTLS {
			e.CertPfad, e.KeyPfad = cert, key
		}
		tr, vs := bootstrap.PruefeAdminTransport(e, time.Now())
		if tr.Weg != weg || len(vs) != 0 {
			t.Fatalf("Weg %d: Prüfung liefert Weg %d mit Verstößen %v", weg, tr.Weg, vs)
		}
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		if ta, ok := l.Addr().(*net.TCPAddr); !ok || !ta.IP.IsLoopback() {
			t.Fatalf("die Adresse des Listeners ist keine Loopback-Adresse: %v", l.Addr())
		}
		srv := &http.Server{
			Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
			ReadHeaderTimeout: 5 * time.Second, ErrorLog: log.New(io.Discard, "", 0),
		}
		go func() { _ = srv.Serve(bootstrap.AdminUmhuelle(tr, l)) }()
		klartext := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
		mitTLS := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{DisableKeepAlives: true, TLSClientConfig: wurzel.Klient("localhost")}}
		for _, tlsClient := range []bool{false, true} {
			zellen++
			c, schema := klartext, "http://"
			if tlsClient {
				c, schema = mitTLS, "https://"
			}
			status, err := sicherheitFrage(t, schema+l.Addr().String(), c)
			erreicht := status == http.StatusOK && err == nil
			if soll := (weg == bootstrap.WegLoopback) != tlsClient; erreicht != soll {
				t.Errorf("Weg %d, TLS-Client %v: Status %d, Fehler %v; erreicht %v, erwartet %v", weg, tlsClient, status, err, erreicht, soll)
			}
			if tlsClient && weg == bootstrap.WegLoopback && (err == nil || !strings.Contains(err.Error(), "server gave HTTP response to HTTPS client")) {
				t.Errorf("der TLS-Client am Klartext-Listener scheitert nicht im Handshake mit der Klartext-Antwort des Servers: %v", err)
			}
		}
		_ = srv.Close()
		_ = l.Close()
	}
	if zellen != 4 {
		t.Errorf("M-Erreichbar hat %d Zellen, erwartet 4", zellen)
	}
}
