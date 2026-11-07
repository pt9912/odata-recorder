//go:build linux

package admin_test

import (
	"bytes"
	"errors"
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
	// Die übrigen Gründe: nicht lesbar (drei Formen von Stat-Fehlern), nicht anlegbar (zwei Formen) und der Weg des
	// Verlierers (fs.ErrExist, dann ladeVorhandene) mit gültiger Datei, Symlink ohne Ziel und Rechteverstoß.
	dir := t.TempDir()
	schreibe(t, filepath.Join(dir, "datei"), hex64(), 0o600)
	if err := os.Symlink(filepath.Join(dir, "b"), filepath.Join(dir, "a")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "a"), filepath.Join(dir, "b")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "ohne-ziel"), filepath.Join(dir, "ziellos")); err != nil {
		t.Fatal(err)
	}
	zusatz := []verstoss{
		{"Pfadelement ist Datei", filepath.Join(dir, "datei", "x"), gNichtLesbar},
		{"Symlinkschleife", filepath.Join(dir, "a"), gNichtLesbar},
		{"NUL im Pfad", filepath.Join(dir, "x") + "\x00y", gNichtLesbar},
		{"Eltern fehlt", filepath.Join(dir, "fehlt", "x", "admin.token"), gNichtAnlegbar},
		{"Pfadelement ist Symlink ohne Ziel", filepath.Join(dir, "ziellos", "admin.token"), gNichtAnlegbar},
	}
	gruende := map[string]bool{}
	aufrufe := 0
	zaehle := func(err error) {
		aufrufe++
		var tf *admin.TokenFehler
		if errors.As(err, &tf) {
			gruende[tf.Grund] = true
		}
	}
	verlierer := func(name string, halt func(pfad string)) {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		_, err := admin.LadeMitHaltepunkten(pfad, nil, func(string) { halt(pfad) })
		zaehle(err)
	}
	aus, fehl, logs := erfasse(t, func() {
		for _, p := range []string{neu, gueltig, neu} {
			_, err := admin.LadeToken(p)
			zaehle(err)
		}
		for _, v := range append(vs, zusatz...) {
			_, err := mitFrist(t, v.name, func() ([]byte, error) { return admin.LadeToken(v.pfad) })
			zaehle(err)
			if v.grund == gNichtAnlegbar || v.grund == gNichtLesbar {
				if _, lerr := os.Lstat(filepath.Join(dir, "fehlt")); lerr == nil {
					t.Errorf("%s: es entstand ein Verzeichnis", v.name)
				}
			}
		}
		verlierer("gültig", func(p string) { _, _ = admin.LadeToken(p) })
		verlierer("Symlink ohne Ziel", func(p string) { _ = os.Symlink(filepath.Join(filepath.Dir(p), "ohne-ziel"), p) })
		verlierer("Rechte", func(p string) {
			_ = os.WriteFile(p, []byte(hex64()), 0o644)
			_ = os.Chmod(p, 0o644)
		})
		// Der vierte Weg zu Datei nicht anlegbar: Link scheitert mit einem anderen Fehler als fs.ErrExist.
		_, err := admin.LadeMitHaltepunkten(filepath.Join(t.TempDir(), "admin.token"), nil, func(h string) { _ = os.Remove(h) })
		zaehle(err)
	})
	if aufrufe != 3+29+5+3+1 {
		t.Fatalf("%d Aufrufe, erwartet 41", aufrufe)
	}
	sieben := []string{gNichtLesbar, gVerzeichnis, gKeineDatei, gRechte, gMehrZeilen, gZuKurz, gNichtAnlegbar}
	for _, g := range sieben {
		if !gruende[g] {
			t.Errorf("Grund %q kam in keinem Aufruf vor", g)
		}
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
