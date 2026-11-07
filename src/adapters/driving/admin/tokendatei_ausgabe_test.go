//go:build linux

package admin_test

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// erfasse ruft f und liefert, was f auf Dateideskriptor 1, auf Dateideskriptor 2 und in das Standard-Log schrieb.
func erfasse(t *testing.T, f func()) (aus, fehl, logs string) {
	t.Helper()
	r1, w1, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	r2, w2, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	alt1, err := syscall.Dup(1)
	if err != nil {
		t.Fatal(err)
	}
	alt2, err := syscall.Dup(2)
	if err != nil {
		t.Fatal(err)
	}
	var logPuffer bytes.Buffer
	altLog := log.Writer()
	if err := syscall.Dup3(int(w1.Fd()), 1, 0); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Dup3(int(w2.Fd()), 2, 0); err != nil {
		t.Fatal(err)
	}
	log.SetOutput(&logPuffer)
	func() {
		defer func() {
			log.SetOutput(altLog)
			_ = syscall.Dup3(alt1, 1, 0)
			_ = syscall.Dup3(alt2, 2, 0)
		}()
		f()
	}()
	_ = syscall.Close(alt1)
	_ = syscall.Close(alt2)
	_ = w1.Close()
	_ = w2.Close()
	b1, _ := io.ReadAll(r1)
	b2, _ := io.ReadAll(r2)
	_ = r1.Close()
	_ = r2.Close()
	return string(b1), string(b2), logPuffer.String()
}

// T25: LadeToken schreibt weder Log noch stdout noch stderr, beim Erzeugen, bei einer gültigen Datei und bei jedem Verstoß.
func TestKeineAusgabe(t *testing.T) {
	vs := verstoesse(t)
	neu := filepath.Join(t.TempDir(), "admin.token")
	gueltig := filepath.Join(t.TempDir(), "admin.token")
	schreibe(t, gueltig, hex64(), 0o600)
	aufrufe := 0
	aus, fehl, logs := erfasse(t, func() {
		for _, p := range []string{neu, gueltig, neu} {
			_, _ = admin.LadeToken(p)
			aufrufe++
		}
		for _, v := range vs {
			_, _ = mitFrist(t, v.name, func() ([]byte, error) { return admin.LadeToken(v.pfad) })
			aufrufe++
		}
	})
	if aufrufe != 3+29 {
		t.Fatalf("%d Aufrufe, erwartet 32", aufrufe)
	}
	if aus != "" || fehl != "" || logs != "" {
		t.Errorf("Ausgabe: stdout %q, stderr %q, log %q; erwartet leer", aus, fehl, logs)
	}
}

// Die Erfassung selbst sieht Ausgabe: ein Schreiben auf Deskriptor 1, 2 und ins Log wird gezählt.
func TestErfassungSiehtAusgabe(t *testing.T) {
	aus, fehl, logs := erfasse(t, func() {
		_, _ = syscall.Write(1, []byte("a"))
		_, _ = syscall.Write(2, []byte("b"))
		log.Print("c")
	})
	if aus != "a" || fehl != "b" || logs == "" {
		t.Errorf("stdout %q, stderr %q, log %q", aus, fehl, logs)
	}
}
