package e2e_test

import (
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestE2EProzess ist die Stufe von LH-RB-03: kein Außennetz, ungültige Konfiguration, SIGTERM
// (SPEC-037 Nr. 7).
func TestE2EProzess(t *testing.T) {
	binaryPfad(t)
	t.Run("außen", func(t *testing.T) {
		c, err := net.DialTimeout("tcp", "192.0.2.1:80", 5*time.Second)
		if err == nil {
			_ = c.Close()
			t.Fatal("192.0.2.1:80 ist erreichbar: der Container hat Außennetz")
		}
		if !errors.Is(err, syscall.ENETUNREACH) {
			t.Fatalf("Dial nach 192.0.2.1:80: %v, erwartet ENETUNREACH", err)
		}
	})
	t.Run("ungültige Konfiguration", ungueltigeKonfiguration)
	t.Run("SIGTERM", func(t *testing.T) {
		ref := neueReferenz(t)
		p := starte(t, konfig{modus: "record", dir: t.TempDir(), upstream: ref.url()})
		p.bereit()
		if code := p.beende(); code != 0 {
			t.Errorf("SIGTERM: Exit %d, erwartet 0\nstderr: %s", code, p.fehl)
		}
		p.verweigert()
	})
}

// ungueltigeKonfiguration fährt die vier Umgebungen von SPEC-037 Nr. 5: Exit 1, `ODR-E2001` je Verstoß
// auf stderr (die Schlüssel in der Reihenfolge der Meldung), nichts auf stdout, kein Listener.
func ungueltigeKonfiguration(t *testing.T) {
	gueltig := func(dir string) []string {
		return []string{"LISTEN=127.0.0.1:%PORT%", "DATA_DIR=" + dir, "SERVICE_ID=ref", "SERVICE_ROOT_PATH=/odata", "UPSTREAM_URL=http://127.0.0.1:9/odata"}
	}
	faelle := []struct {
		name string
		env  func(dir string) []string
		soll []string
	}{
		{"leer", func(string) []string { return []string{} }, []string{"MODE"}},
		{"MODE=Record", func(d string) []string { return append(gueltig(d), "MODE=Record") }, []string{"MODE"}},
		{"beide Verstöße", func(d string) []string {
			return append(gueltig(d)[:2:2], "MODE=Record", "SERVICE_ID=ref", "SERVICE_ROOT_PATH=/odata/")
		}, []string{"MODE", "SERVICE_ROOT_PATH"}},
		{"record ohne UPSTREAM_URL", func(d string) []string {
			return []string{"LISTEN=127.0.0.1:%PORT%", "DATA_DIR=" + d, "SERVICE_ID=ref", "SERVICE_ROOT_PATH=/odata", "MODE=record"}
		}, []string{"UPSTREAM_URL"}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			adresse, env := freierPort(t), f.env(t.TempDir())
			if len(env) == 0 {
				adresse = "127.0.0.1:8080"
			}
			for i, e := range env {
				env[i] = strings.ReplaceAll(e, "%PORT%", adresse[strings.LastIndex(adresse, ":")+1:])
			}
			p := startRoh(t, adresse, env)
			verbunden := false
			for stop := time.Now().Add(bereitFrist); !p.beendet() && time.Now().Before(stop); {
				if c, err := net.DialTimeout("tcp", adresse, 20*time.Millisecond); err == nil {
					verbunden = true
					_ = c.Close()
				}
			}
			p.warteEnde()
			if p.exit != 1 {
				t.Errorf("Exit %d, erwartet 1\nstderr: %s", p.exit, p.fehl)
			}
			var schluessel []string
			for _, z := range strings.Split(strings.TrimSpace(p.fehl.String()), "\n") {
				if s, ok := strings.CutPrefix(z, "ODR-E2001 "); ok {
					schluessel = append(schluessel, s[:strings.Index(s, ":")])
				}
			}
			if strings.Join(schluessel, ",") != strings.Join(f.soll, ",") {
				t.Errorf("Verstöße %v, erwartet %v\nstderr: %s", schluessel, f.soll, p.fehl)
			}
			if p.aus.String() != "" || verbunden {
				t.Errorf("stdout %q, Listener geöffnet: %v", p.aus, verbunden)
			}
		})
	}
}
