package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// baueProgramm übersetzt das Programm für die Tests, die es als Prozess starten.
func baueProgramm(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "mutationslauf")
	c := exec.Command("go", "build", "-o", bin, ".")
	c.Env = append(os.Environ(), "CGO_ENABLED=0", "GOFLAGS=-mod=readonly")
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func lesePid(t *testing.T, p string) int {
	t.Helper()
	roh, err := os.ReadFile(p)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(roh)))
	return n
}

// TestAbbruchRaeumtAuf (SPEC-210 Nr. 10 a, b): auf SIGTERM, SIGINT und bei der Gesamtgrenze beendet das Programm den hängenden Test
// samt Kindprozess, entfernt die Ablage, gibt jede nicht fertige Mutation als abbruch aus und endet mit Exit 3.
func TestAbbruchRaeumtAuf(t *testing.T) {
	if testing.Short() {
		t.Skip("Gesamtgrenze braucht 30 Sekunden")
	}
	bin := baueProgramm(t)
	for _, art := range []string{"SIGTERM", "SIGINT", "Gesamtgrenze"} {
		t.Run(art, func(t *testing.T) {
			baum, ablage, marker := fixtureBaum(t), t.TempDir(), t.TempDir()
			liste := filepath.Join(t.TempDir(), "liste.json")
			if err := os.WriteFile(liste, listeJSON(t,
				mut("haengt", "src/w/w.go", "return false", "return true", "TestW", klasseAndere),
				mut("zweite", "src/p/a.go", "return a }", "return a + 0 }", "TestId", klasseAndere),
				mut("dritte", "src/p/a.go", "return a + b", "return b + a", "TestAdd", klasseAndere)), 0o644); err != nil {
				t.Fatal(err)
			}
			var aus bytes.Buffer
			c := exec.Command(bin, liste, baum, ablage, "")
			c.Env = append(os.Environ(), "FX_MARKER="+marker, "GOFLAGS=-mod=readonly", "GOPROXY=off")
			if art == "Gesamtgrenze" {
				c.Env = append(c.Env, "MUTATIONSLAUF_GRENZE=30")
			}
			c.Stdout = &aus
			if err := c.Start(); err != nil {
				t.Fatal(err)
			}
			wache := time.AfterFunc(150*time.Second, func() { _ = c.Process.Kill() })
			defer wache.Stop()
			testPid, kindPid := 0, 0
			for stop := time.Now().Add(120 * time.Second); time.Now().Before(stop) && (testPid == 0 || kindPid == 0); time.Sleep(50 * time.Millisecond) {
				testPid, kindPid = lesePid(t, filepath.Join(marker, "test")), lesePid(t, filepath.Join(marker, "kind"))
			}
			if testPid == 0 || kindPid == 0 {
				_ = c.Process.Kill()
				t.Fatal("der hängende Test ist nicht gestartet")
			}
			if sig, ok := map[string]syscall.Signal{"SIGTERM": syscall.SIGTERM, "SIGINT": syscall.SIGINT}[art]; ok {
				_ = c.Process.Signal(sig)
			}
			err := c.Wait()
			var ee *exec.ExitError
			if !errors.As(err, &ee) || ee.ExitCode() != 3 {
				t.Errorf("Exit %v statt 3", err)
			}
			r := ergebnisse(aus.String())
			if r["haengt"] != ergAbbruch || r["zweite"] != ergAbbruch || r["dritte"] != ergAbbruch ||
				!strings.Contains(aus.String(), "abbruch=3 ") {
				t.Errorf("Ausgabe:\n%s", aus.String())
			}
			if vorhanden(filepath.Join(ablage, "lauf")) {
				t.Error("die Ablage lauf bleibt zurück")
			}
			if lebt(testPid) || lebt(kindPid) {
				t.Errorf("Prozesse leben: Test %v, Kind %v", lebt(testPid), lebt(kindPid))
			}
		})
	}
}

// TestHaengerNachGnade (SPEC-210 Nr. 9 c): ein Test-Prozess, der -test.timeout nicht beachtet, wird nach TestZeit und Gnade
// beendet; die Mutation ist abbruch, der Lauf geht weiter (Exit 1) und die nächste Mutation ist unbeeinflusst.
func TestHaengerNachGnade(t *testing.T) {
	baum := fixtureBaum(t)
	w, aus, _ := echtWerk(t, baum)
	w.TestZeit, w.Gnade = time.Second, time.Second
	t0 := time.Now()
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	code := w.Lauf(ctx, muts(t, baum,
		mut("haengt", "src/h/h.go", "return false", "return true", "TestH", klasseAndere),
		mut("danach", "src/p/a.go", "return a }", "return a + 0 }", "TestId", klasseAndere)))
	r := ergebnisse(aus.String())
	if code != 1 || r["haengt"] != ergAbbruch || r["danach"] != ergGruen {
		t.Errorf("Exit %d, %v", code, r)
	}
	if d := time.Since(t0); d < 2*time.Second {
		t.Errorf("der Hänger endete nach %v, vor TestZeit und Gnade", d)
	}
	if vorhanden(filepath.Join(w.Ablage, "lauf")) {
		t.Error("die Ablage lauf bleibt zurück")
	}
}
