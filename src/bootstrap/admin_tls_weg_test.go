package bootstrap_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/tlstest"
)

// transportAn liefert, wenn der Listener an adresse TLS spricht, die Seriennummer des Blatts, das er vorzeigt, und `klartext`, wenn er
// HTTP beantwortet.
func transportAn(adresse string) string {
	c, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", adresse, &tls.Config{InsecureSkipVerify: true})
	if err == nil {
		defer func() { _ = c.Close() }()
		return c.ConnectionState().PeerCertificates[0].SerialNumber.String()
	}
	status, body, _ := sende(neuerAdminClient(), "GET", "http://"+adresse+"/healthz", "", "")
	if status > 0 && !strings.Contains(body, "HTTPS server") {
		return "klartext"
	}
	return err.Error()
}

// TestAdminPaareUnabhaengig (SPEC-299, kein Wert des einen Paars wirkt auf den anderen Listener): gesetzte Paare {keins, ADMIN_TLS_*,
// LISTEN_TLS_*, beide} × ADMIN_LISTEN {Loopback, außerhalb mit Flag, außerhalb ohne Flag} × 4 Modi = 48 Zellen über Start. Jeder
// Listener spricht allein sein eigenes Paar oder HTTP; ohne Admin-Paar außerhalb von Loopback ohne Flag endet der Start, auch mit
// LISTEN_TLS_*; das Schema der abgeleiteten PUBLIC_BASE_URL (mock, hybrid) folgt allein LISTEN_TLS_*.
func TestAdminPaareUnabhaengig(t *testing.T) {
	w := neueWerkstatt(t)
	adminCert, adminKey, adminBlatt := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	listenCert, listenKey, listenBlatt := w.tlsPaar(w.wurzel, tlstest.Optionen{})
	orte := [][]string{nil, klartext(), {"ADMIN_LISTEN=0.0.0.0:8081"}}
	zellen := 0
	for _, modus := range vierModi() {
		for _, paar := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
			for ort, adresse := range orte {
				zellen++
				name := fmt.Sprintf("%s, Admin-Paar %v, LISTEN-Paar %v, Ort %d", modus, paar[0], paar[1], ort)
				env := append([]string{"PUBLIC_BASE_URL="}, adresse...)
				sollAdmin, sollOData, schema := "klartext", "klartext", "http://"
				if paar[0] {
					env = append(env, "ADMIN_TLS_CERT="+adminCert, "ADMIN_TLS_KEY="+adminKey)
					sollAdmin = adminBlatt.Zertifikat.SerialNumber.String()
				}
				if paar[1] {
					env = append(env, "LISTEN_TLS_CERT="+listenCert, "LISTEN_TLS_KEY="+listenKey)
					sollOData, schema = listenBlatt.Zertifikat.SerialNumber.String(), "https://"
				}
				env = adminEnv(modus, filepath.Join(t.TempDir(), "d"), env...)
				l := starte(env)
				l.warte()
				if ort == 2 && !paar[0] {
					if code := l.beende(); code != 1 || l.stderr.String() != zeileZusammenhang+"\n" {
						t.Errorf("%s: Rückgabe %d, %q, erwartet den Zusammenhang", name, code, l.stderr.String())
					}
					continue
				}
				odata, admin := l.listener(t)
				if o, a := transportAn(odata), transportAn(admin); o != sollOData || a != sollAdmin {
					t.Errorf("%s: OData %q, Admin %q; erwartet %q und %q", name, o, a, sollOData, sollAdmin)
				}
				l.beende()
				k, zeilen := lade(t, env...)
				if abgeleitet := modus == modusMock || modus == modusHybrid; zeilen != nil || abgeleitet && !strings.HasPrefix(k.PublicBaseURL, schema) {
					t.Errorf("%s: PUBLIC_BASE_URL %q (%q), erwartet %s", name, k.PublicBaseURL, zeilen, schema)
				}
			}
		}
	}
	if zellen != 48 {
		t.Errorf("%d Zellen, erwartet 48", zellen)
	}
}

// TestAdminBuildLiestErneut (SPEC-299 Nr. 8 b, SPEC-291 Nr. 8): die Dateien ändern sich nach dem Laden; Listener {OData, Admin} ×
// Änderung {neues Paar, Schlüssel passt nicht, Zertifikat gelöscht} × 4 Modi = 24 Zellen. Es gilt, was Build liest: der Listener zeigt
// das neue Blatt, nach einem weiteren Wechsel auch ein zweites Build, sonst meldet Build die Zeile des Schlüssels.
func TestAdminBuildLiestErneut(t *testing.T) {
	w := neueWerkstatt(t)
	zellen := 0
	for _, modus := range vierModi() {
		for _, praefix := range []string{"LISTEN_TLS_", "ADMIN_TLS_"} {
			for _, aenderung := range []string{"neu", "passt nicht", "gelöscht"} {
				zellen++
				cert, key, _ := w.tlsPaar(w.wurzel, tlstest.Optionen{})
				k, zeilen := lade(t, adminEnv(modus, filepath.Join(t.TempDir(), "d"), praefix+"CERT="+cert, praefix+"KEY="+key)...)
				if zeilen != nil {
					t.Fatalf("%s %s: %q", modus, praefix, zeilen)
				}
				switch aenderung {
				case "neu":
					for range 2 {
						blatt := w.wurzel.Blatt(t, tlstest.Optionen{})
						schreibeDatei(t, cert, string(blatt.ZertifikatPEM()), 0o600)
						schreibeDatei(t, key, string(w.schluessel(blatt)), 0o600)
						if got := bedientVon(t, k, praefix == "ADMIN_TLS_"); got != blatt.Zertifikat.SerialNumber.String() {
							t.Errorf("%s %s: der Listener zeigt %q, erwartet das Blatt der Dateien vor Build", modus, praefix, got)
						}
					}
					continue
				case "passt nicht":
					schreibeDatei(t, key, string(w.schluessel(w.fremd)), 0o600)
					zeilen = []string{"ODR-E2001 " + praefix + "KEY: Schlüssel ungültig oder passt nicht zum Zertifikat"}
				default:
					if err := os.Remove(cert); err != nil {
						t.Fatal(err)
					}
					zeilen = []string{"ODR-E2001 " + praefix + "CERT: Datei nicht lesbar"}
				}
				if _, got := buildMit(t, &pruefProtokoll{echt: true}, k); !slices.Equal(got, zeilen) {
					t.Errorf("%s %s %s: Build %q, erwartet %q", modus, praefix, aenderung, got, zeilen)
				}
			}
		}
	}
	if zellen != 24 {
		t.Errorf("%d Zellen, erwartet 24", zellen)
	}
}

// bedientVon baut k, fährt hoch und liefert transportAn des Admin-Listeners (admin) oder des OData-Listeners; danach fährt es herunter.
func bedientVon(t *testing.T, k bootstrap.Konfiguration, admin bool) string {
	t.Helper()
	b := bootstrap.StandardBauteile()
	b.Listen = loopbackBind(b.Listen, make(chan string))
	h, err := bootstrap.BuildMit(context.Background(), k, nil, b)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer func() { _ = h.Herunterfahren(context.Background()) }()
	if err := h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("Hochfahren: %v", err)
	}
	if admin {
		return transportAn(h.AdminAdresse())
	}
	return transportAn(h.Adresse())
}

// TestAdminUmhuellung (C05): 3 Wege {TLS mit gültigem Paar, Loopback, Klartext} × 4 Modi = 12 Zellen an Build und Hochfahren. Nur der
// Weg TLS umhüllt; die Dateien des Paars sind nach Build gelöscht, kein späterer Aufruf lädt sie. Jeder Bind geht auf 127.0.0.1.
func TestAdminUmhuellung(t *testing.T) {
	w := neueWerkstatt(t)
	zellen := 0
	for _, modus := range vierModi() {
		for _, weg := range []bootstrap.AdminWeg{bootstrap.WegTLS, bootstrap.WegLoopback, bootstrap.WegKlartext} {
			zellen++
			k := listenerKonfig(t.TempDir(), modus)
			switch weg {
			case bootstrap.WegTLS:
				k.AdminTLSCert, k.AdminTLSKey, _ = w.tlsPaar(w.wurzel, tlstest.Optionen{})
			case bootstrap.WegKlartext:
				k.AdminListen, k.AdminAllowPlaintext = "0.0.0.0:0", true
			}
			b := bootstrap.StandardBauteile()
			b.Listen = loopbackBind(b.Listen, make(chan string))
			h, err := bootstrap.BuildMit(context.Background(), k, nil, b)
			if err != nil {
				t.Fatalf("%s Weg %d: %v", modus, weg, err)
			}
			t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
			_, _ = os.Remove(k.AdminTLSCert), os.Remove(k.AdminTLSKey)
			if err := h.Hochfahren(context.Background()); err != nil {
				t.Fatalf("%s Weg %d: %v", modus, weg, err)
			}
			pruefeTransportAm(t, fmt.Sprintf("%s Weg %d", modus, weg), h.AdminAdresse(), weg == bootstrap.WegTLS, w.wurzel)
		}
	}
	if zellen != 12 {
		t.Errorf("%d Zellen, erwartet 12", zellen)
	}
}

// wegBild ist die Form eines Bildes: Eingabe und Weg der Prüfung, Weg und TLS des Handle, die Zeilen und die Warnung.
const wegBild = "Eingaben %+v, geprüft %v, Weg %d, TLS %v, Zeilen %q, Warnung %q"

// bildWie baut k mit der echten Prüfung und liefert sein Bild.
func bildWie(t *testing.T, k bootstrap.Konfiguration) string {
	t.Helper()
	p := &pruefProtokoll{echt: true}
	tr, zeilen := buildMit(t, p, k)
	var warnung strings.Builder
	bootstrap.AdminWarne(tr, slog.New(slog.NewTextHandler(&warnung, nil)))
	_, rest, _ := strings.Cut(strings.TrimSpace(warnung.String()), " ")
	return fmt.Sprintf(wegBild, p.eingaben, p.wege, tr.Weg, tr.TLS != nil, zeilen, rest)
}

// TestAdminWegJeModus (C07): M-Weg, 8 Hosts × Paar × Flag = 32 Zellen von SPEC-299 Nr. 3, × 4 Modi = 128 Zellen an der Naht in Build.
// Orakel je Zelle für jeden Modus dasselbe Bild nach der Tabelle: ohne Weg die Zeile des Zusammenhangs, im Weg Klartext die Warnung.
func TestAdminWegJeModus(t *testing.T) {
	w := neueWerkstatt(t)
	cert, key := w.datei(w.gut.ZertifikatPEM()), w.datei(w.schluessel(w.gut))
	hosts := []struct {
		listen   string
		loopback bool
	}{
		{"127.0.0.1:8081", true}, {"[::1]:8081", true}, {"localhost:8081", true},
		{"0.0.0.0:8081", false}, {"[::]:8081", false}, {":8081", false}, {"recorder.internal:8081", false}, {"10.0.0.1:8081", false},
	}
	zellen := 0
	for _, host := range hosts {
		for _, paar := range []bool{false, true} {
			for _, flag := range []bool{false, true} {
				e := bootstrap.AdminEingabe{Listen: host.listen, Klartext: flag, FormGueltig: true}
				if paar {
					e.CertPfad, e.KeyPfad = cert, key
				}
				weg, zeilen, warnung := bootstrap.WegKeiner, []string{zeileZusammenhang}, ""
				switch {
				case paar:
					weg, zeilen = bootstrap.WegTLS, nil
				case host.loopback:
					weg, zeilen = bootstrap.WegLoopback, nil
				case flag:
					weg, zeilen, warnung = bootstrap.WegKlartext, nil, zeileWarnung
				}
				soll := fmt.Sprintf(wegBild, []bootstrap.AdminEingabe{e}, []bootstrap.AdminWeg{weg}, weg, paar, zeilen, warnung)
				for _, modus := range vierModi() {
					zellen++
					k := listenerKonfig(t.TempDir(), modus)
					k.AdminListen, k.AdminTLSCert, k.AdminTLSKey, k.AdminAllowPlaintext = host.listen, e.CertPfad, e.KeyPfad, flag
					if bild := bildWie(t, k); bild != soll {
						t.Errorf("%s Paar %v Flag %v, %s: %s, erwartet %s", host.listen, paar, flag, modus, bild, soll)
					}
				}
			}
		}
	}
	if zellen != 128 {
		t.Errorf("%d Zellen, erwartet 128", zellen)
	}
}
