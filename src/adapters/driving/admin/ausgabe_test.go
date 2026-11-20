package admin_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// kindLauf fährt dieselbe Testbinärdatei mit der Umgebungsvariablen umg (Wert: ein Verzeichnis) und nur dem Test test in
// einem Kindprozess und liefert dessen stdout und stderr; der Kindprozess schreibt die Datei `lief`, wenn er sein
// Szenario ausführte.
func kindLauf(t *testing.T, test, umg string) (aus, fehl string) {
	t.Helper()
	dir := t.TempDir()
	ctx, abbruch := context.WithTimeout(context.Background(), 60*time.Second)
	defer abbruch()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+test+"$", "-test.count=1")
	cmd.Env = append(os.Environ(), umg+"="+dir)
	var a, f bytes.Buffer
	cmd.Stdout, cmd.Stderr = &a, &f
	if err := cmd.Run(); err != nil {
		t.Fatalf("Kindprozess: %v\nstdout: %s\nstderr: %s", err, a.String(), f.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "lief")); err != nil {
		t.Fatalf("der Kindprozess hat das Szenario nicht ausgeführt: %v", err)
	}
	return a.String(), f.String()
}

// kindFertig meldet dem Elternprozess, dass das Szenario lief.
func kindFertig(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "lief"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
}

const kindUmgebungH2 = "ODR_H2_KIND"

// ohneTestAusgabe liefert die Zeilen, die die Testbinärdatei selbst schreibt (`PASS`, `coverage: …`), nicht mehr.
func ohneTestAusgabe(text string) string {
	var rest []string
	for _, z := range strings.Split(text, "\n") {
		if z != "" && z != "PASS" && !strings.HasPrefix(z, "coverage: ") {
			rest = append(rest, z)
		}
	}
	return strings.Join(rest, "\n")
}

// TestAusgabeLeer hält H2 (SPEC-238 Nr. 13): der Adapter schreibt kein Log und nichts auf stdout oder stderr, nach jeder
// der 30 Requestarten R, jedem der neun Schreibstellen W (auch mit Ursache und Panikwert), jeder Panik der Route und jedem
// Kontextfehler. Der Test fährt das Szenario in einem Kindprozess und liest dessen stdout und stderr; der Standardlogger
// schreibt dort auf stderr. Außer den Zeilen der Testbinärdatei (`PASS`, `coverage: …`) steht dort nichts.
func TestAusgabeLeer(t *testing.T) {
	if dir := os.Getenv(kindUmgebungH2); dir != "" {
		kindH2(t, dir)
		return
	}
	aus, fehl := kindLauf(t, "TestAusgabeLeer", kindUmgebungH2)
	if rest := ohneTestAusgabe(aus); rest != "" {
		t.Errorf("stdout trägt außer den Zeilen der Testbinärdatei: %q", rest)
	}
	if fehl != "" {
		t.Errorf("stderr trägt %q, erwartet nichts", fehl)
	}
}

const kindUmgebungProbe = "ODR_H2_PROBE"

// TestAusgabeFaengerProbe hält den Fänger von H2 gegen eine Probe: ein Kindprozess, der je eine Zeile auf stdout und auf
// stderr schreibt, ist über kindLauf und ohneTestAusgabe auf beiden Strömen sichtbar.
func TestAusgabeFaengerProbe(t *testing.T) {
	if dir := os.Getenv(kindUmgebungProbe); dir != "" {
		_, _ = syscall.Write(1, []byte("probe-stdout\n"))
		_, _ = syscall.Write(2, []byte("probe-stderr\n"))
		kindFertig(t, dir)
		return
	}
	aus, fehl := kindLauf(t, "TestAusgabeFaengerProbe", kindUmgebungProbe)
	if got := ohneTestAusgabe(aus); got != "probe-stdout" || fehl != "probe-stderr\n" {
		t.Errorf("stdout ohne Zeilen der Testbinärdatei %q, stderr %q; erwartet %q und %q", got, fehl, "probe-stdout", "probe-stderr\n")
	}
}

// kindH2 ist das Szenario im Kindprozess.
func kindH2(t *testing.T, dir string) {
	t.Helper()
	for _, a := range requestArten() {
		baue(t, a).u.rufeSicher(t, a.name, ohneAuth(a.methode, a.ziel, a.kopf))
	}
	for _, s := range schreibstellen() {
		s.lauf(t, nil)
	}
	for _, e := range []error{context.Canceled, errors.Join(model.ErrUngueltig, context.DeadlineExceeded)} {
		route := neueRoute()
		route.err = e
		u := neuLive(t, route)
		if x := erwarte(t, u.async("GET", "/p"), "Kontextfehler"); x.err == nil {
			t.Errorf("der Client sah einen Status %d, erwartet einen Transportfehler", x.status)
		}
	}
	route := neueRoute()
	route.aendere = func(admin.Anfrage) { panic(&inbound.Fehler{Code: model.CodeE7003, Status: 404}) }
	u := neuLive(t, route)
	if x := erwarte(t, u.async("GET", "/p"), "Panik"); x.err != nil || x.status != 500 {
		t.Errorf("Panik der Route: Antwort %+v, erwartet 500", x)
	}
	kindFertig(t, dir)
}
