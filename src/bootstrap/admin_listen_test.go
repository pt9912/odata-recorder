package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
)

// Die Zeilen der Konfiguration zum Admin-Listener als Literale (SPEC-110 Nr. 7, SPEC-299 Nr. 6, SPEC-300 Nr. 2).
const (
	zeileForm          = "ODR-E2001 ADMIN_LISTEN: Adresse muss host:port sein, Port 1 bis 65535"
	zeileZusammenhang  = "ODR-E2001 ADMIN_LISTEN: außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY"
	adminTokenStandard = "admin.token"
)

// adminEnv ist eine gültige Umgebung im Modus modus; spätere Einträge ersetzen frühere.
func adminEnv(modus, dir string, extra ...string) []string {
	return append([]string{"MODE=" + modus, "DATA_DIR=" + dir, "UPSTREAM_URL=http://127.0.0.1:1/o", "PUBLIC_BASE_URL=http://h:1"}, extra...)
}

// listenerKonfig ist ein Literal mit Admin-Listener auf einem freien Port und der Token-Datei in dir.
func listenerKonfig(dir, modus string) bootstrap.Konfiguration {
	k := routenKonfig(dir, modus, "http://127.0.0.1:1/o")
	k.AdminListen, k.AdminTokenFile = "127.0.0.1:0", filepath.Join(dir, adminTokenStandard)
	return k
}

// TestAdminListenLesen (B01): Schlüssel {LISTEN, ADMIN_LISTEN} × die Werte des Tests von LISTEN × 4 Modi; dazu ungesetzt, leer
// und Leerraum.
func TestAdminListenLesen(t *testing.T) {
	loopback := []string{"127.0.0.1:1", "localhost:8080", "[::1]:8080"}
	werte := slices.Concat(listenGueltig(), listenUngueltig())
	zellen := 0
	for _, modus := range vierModi() {
		for _, schluessel := range []string{"LISTEN", "ADMIN_LISTEN"} {
			for _, w := range werte {
				zellen++
				k, zeilen := lade(t, adminEnv(modus, t.TempDir(), schluessel+"="+w)...)
				var soll []string
				got := k.Listen
				if schluessel == "ADMIN_LISTEN" {
					got = k.AdminListen
				}
				switch {
				case slices.Contains(listenUngueltig(), w):
					soll = []string{"ODR-E2001 " + schluessel + ": Adresse muss host:port sein, Port 1 bis 65535"}
				case schluessel == "ADMIN_LISTEN" && !slices.Contains(loopback, w):
					soll = []string{zeileZusammenhang}
				}
				if !slices.Equal(zeilen, soll) || (soll == nil && got != w) {
					t.Errorf("%s %s=%q: Wert %q, Zeilen %q, erwartet %q", modus, schluessel, w, got, zeilen, soll)
				}
			}
		}
		for name, env := range map[string][]string{"ungesetzt": nil, "leer": {"ADMIN_LISTEN="}, "Leerraum": {"ADMIN_LISTEN= \t"}} {
			if k, zeilen := lade(t, adminEnv(modus, t.TempDir(), env...)...); zeilen != nil || k.AdminListen != "127.0.0.1:8081" {
				t.Errorf("%s %s: %q, %q", modus, name, k.AdminListen, zeilen)
			}
		}
		_, zeilen := lade(t, adminEnv(modus, t.TempDir(), "ADMIN_LISTEN=geheim-4711")...)
		if !slices.Equal(zeilen, []string{zeileForm}) {
			t.Errorf("%s: %q, erwartet allein die Form ohne den Wert", modus, zeilen)
		}
	}
	if zellen != 200 || len(werte) != 25 {
		t.Errorf("%d Zellen aus %d Werten, erwartet 200 aus 25", zellen, len(werte))
	}
}

// TestAdminTokenDateiLesen (B02): 5 Wertarten × 4 Modi; das Laden liest und legt die Datei nicht an und trägt ihren Inhalt nie.
func TestAdminTokenDateiLesen(t *testing.T) {
	const canary = "CANARY-token-inhalt-0123456789abcdef"
	zellen := 0
	for _, modus := range vierModi() {
		for _, art := range []string{"ungesetzt", "leer", "Leerraum", "absolut", "relativ"} {
			zellen++
			dir := t.TempDir()
			t.Chdir(dir)
			daten := filepath.Join(dir, "d")
			wert, soll := "", filepath.Join(daten, adminTokenStandard)
			var env []string
			switch art {
			case "leer":
				env = []string{"ADMIN_TOKEN_FILE="}
			case "Leerraum":
				env = []string{"ADMIN_TOKEN_FILE=\t "}
			case "absolut":
				wert, soll = filepath.Join(dir, "tok"), filepath.Join(dir, "tok")
			case "relativ":
				wert, soll = " rel.token ", "rel.token"
			}
			if wert != "" {
				env = []string{"ADMIN_TOKEN_FILE=" + wert}
				if err := os.WriteFile(soll, []byte(canary), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			k, zeilen := lade(t, adminEnv(modus, daten, env...)...)
			if zeilen != nil || k.AdminTokenFile != soll || strings.Contains(fmt.Sprintf("%+v", k), canary) {
				t.Errorf("%s %s: AdminTokenFile %q, Zeilen %q, erwartet %q ohne Inhalt", modus, art, k.AdminTokenFile, zeilen, soll)
			}
			if inhalt, err := os.ReadFile(soll); wert == "" && !errors.Is(err, os.ErrNotExist) || wert != "" && string(inhalt) != canary {
				t.Errorf("%s %s: das Laden hat die Datei angelegt oder geändert (%v)", modus, art, err)
			}
		}
	}
	if zellen != 20 {
		t.Errorf("%d Zellen, erwartet 20", zellen)
	}
}

// TestAdminPraefix (B03): jeder andere Schlüssel mit dem Präfix ist abgelehnt, die fünf gelesenen nie (ein Pfad des Paars ohne den
// anderen meldet die Gründe von SPEC-299 Nr. 4), andere Schreibweisen werden ignoriert.
func TestAdminPraefix(t *testing.T) {
	zellen := 0
	pruefe := func(name, wert string, soll ...string) {
		zellen++
		t.Run(name+"="+wert, func(t *testing.T) {
			_, zeilen := lade(t, adminEnv("record", t.TempDir(), name+"="+wert)...)
			if !slices.Equal(zeilen, soll) {
				t.Errorf("%q, erwartet %q", zeilen, soll)
			}
		})
	}
	for _, name := range []string{"ADMIN_TOKEN", "ADMIN_", "ADMIN_LISTEN2", "ADMIN_TLS", "ADMIN_TLS_CA", "ADMIN_PASSWORD", "ADMIN_TOKEN_FILE_X", "ADMIN_X"} {
		pruefe(name, "x", zeile(name, grundAbgelehnt))
		pruefe(name, "")
	}
	pruefe("ADMIN_LISTEN", "127.0.0.1:1")
	pruefe("ADMIN_TOKEN_FILE", "x")
	pruefe("ADMIN_TLS_CERT", "x", zeile("ADMIN_TLS_CERT", "Datei nicht lesbar"),
		zeile("ADMIN_TLS_KEY", "erforderlich, wenn ADMIN_TLS_CERT gesetzt ist"))
	pruefe("ADMIN_TLS_KEY", "x", zeile("ADMIN_TLS_CERT", "erforderlich, wenn ADMIN_TLS_KEY gesetzt ist"),
		zeile("ADMIN_TLS_KEY", "Datei nicht lesbar"))
	pruefe("ADMIN_ALLOW_PLAINTEXT", "true")
	pruefe("admin_listen", "x")
	pruefe("XADMIN_LISTEN", "x")
	pruefe("ADMIN", "x")
	if zellen != 24 {
		t.Errorf("%d Zellen, erwartet 24", zellen)
	}
}

// zaehlendeBauteile zählt Listen und das Öffnen der Ablage.
func zaehlendeBauteile(binds, oeffnen *int) bootstrap.Bauteile {
	b := bootstrap.StandardBauteile()
	listen, oeffne := b.Listen, b.Oeffne
	b.Listen = func(ctx context.Context, network, addr string) (net.Listener, error) {
		*binds++
		return listen(ctx, network, addr)
	}
	b.Oeffne = func(ctx context.Context, pfad string, opt sqlite.Options) (*sqlite.DB, error) {
		*oeffnen++
		return oeffne(ctx, pfad, opt)
	}
	return b
}

// TestAdminLoopbackOhnePaar (B19): ohne Paar und ohne Flag, 5 Hosts außerhalb von Loopback × {Laden, Build} × 4 Modi; dazu in Build
// beide Zeilen neben einem Verstoß von LISTEN_TLS_* × 4 Modi.
func TestAdminLoopbackOhnePaar(t *testing.T) {
	zellen, binds, oeffnen := 0, 0, 0
	for _, modus := range vierModi() {
		for _, host := range []string{"0.0.0.0:8081", "[::]:8081", ":8081", "recorder.internal:8081", "10.0.0.5:8081"} {
			zellen += 2
			t.Run(modus+"/Laden/"+host, func(t *testing.T) {
				frisch := filepath.Join(t.TempDir(), "frisch")
				soll := zeileZusammenhang
				if host == ":8081" {
					soll = zeileForm
				}
				_, zeilen := lade(t, adminEnv(modus, frisch, "ADMIN_LISTEN="+host)...)
				if _, err := os.Stat(frisch); !slices.Equal(zeilen, []string{soll}) || err == nil {
					t.Errorf("%q (DATA_DIR: %v), erwartet %q ohne DATA_DIR", zeilen, err, soll)
				}
			})
			t.Run(modus+"/Build/"+host, func(t *testing.T) {
				k := listenerKonfig(t.TempDir(), modus)
				k.AdminListen = host
				pruefeBuild(t, k, &binds, &oeffnen, zeileZusammenhang)
			})
		}
		zellen++
		t.Run(modus+"/Build neben LISTEN_TLS", func(t *testing.T) {
			k := listenerKonfig(t.TempDir(), modus)
			k.AdminListen, k.ListenTLSCert = "0.0.0.0:8081", filepath.Join(t.TempDir(), "fehlt")
			pruefeBuild(t, k, &binds, &oeffnen, zeileZusammenhang, zeile("LISTEN_TLS_CERT", "Datei nicht lesbar"),
				zeile("LISTEN_TLS_KEY", "erforderlich, wenn LISTEN_TLS_CERT gesetzt ist"))
		})
	}
	if zellen != 44 || binds != 0 || oeffnen != 0 {
		t.Errorf("%d Zellen (erwartet 44), %d Bind, %d-mal die Ablage geöffnet (erwartet 0)", zellen, binds, oeffnen)
	}
}

// pruefeBuild erwartet von Build allein die Zeilen soll und keinen Handle.
func pruefeBuild(t *testing.T, k bootstrap.Konfiguration, binds, oeffnen *int, soll ...string) {
	t.Helper()
	h, err := bootstrap.BuildMit(context.Background(), k, nil, zaehlendeBauteile(binds, oeffnen))
	var kf *bootstrap.KonfigFehler
	if h != nil || !errors.As(err, &kf) || !slices.Equal(kf.Zeilen, soll) {
		t.Errorf("Build %s %s: %v, erwartet %q", k.Modus, k.AdminListen, err, soll)
	}
}
