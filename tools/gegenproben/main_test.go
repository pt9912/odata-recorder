package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBewerte(t *testing.T) {
	verstoss := Fall{Name: "v", WantExit: 1, WantText: []string{"regel"}}
	erlaubt := Fall{Name: "e"}
	tests := []struct {
		name       string
		f          Fall
		code       int
		out        string
		abgelaufen bool
		rot        string // Teiltext des Grundes; leer: grün
	}{
		{"Verstoß gemeldet", verstoss, 1, "regel: x", false, ""},
		{"Verstoß nicht gemeldet", verstoss, 0, "", false, "Verstoß nicht gemeldet"},
		{"anderer Code", verstoss, 2, "regel", false, "erwartet Exit 1, war 2"},
		{"falsches Wort", verstoss, 1, "anderes", false, `nennt "regel" nicht`},
		{"zweiter Text fehlt", Fall{Name: "w", WantExit: 1, WantText: []string{"a", "b"}}, 1, "a", false, `nennt "b" nicht`},
		{"Erlaubt grün", erlaubt, 0, "", false, ""},
		{"Erlaubt abgelehnt", erlaubt, 1, "boese", false, "erlaubter Fall abgelehnt"},
		{"Frist", verstoss, -1, "", true, "Zeit überschritten"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bewerte(tc.f, tc.code, tc.out, tc.abgelaufen)
			if tc.rot == "" && got != "" || !strings.Contains(got, tc.rot) {
				t.Fatalf("bewerte = %q, erwartet %q", got, tc.rot)
			}
		})
	}
}

func TestFaelleHabenEindeutigeNamen(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range faelle() {
		if seen[f.Name] {
			t.Fatalf("Name %s doppelt", f.Name)
		}
		seen[f.Name] = true
	}
}

// wurzelMitEingaben legt ein Verzeichnis an, das alle Eingaben als kleine Dateien trägt.
func wurzelMitEingaben(t *testing.T) string {
	t.Helper()
	w := t.TempDir()
	for _, d := range []string{"src/a.go", "go.mod", ".a-check.yml", ".golangci.yml", "tools/harness/lint.sh"} {
		p := filepath.Join(w, d)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

// mitBefehl ersetzt befehl und programme für die Dauer des Tests.
func mitBefehl(t *testing.T, skript func(dir string, g Gate) string, prog ...string) {
	t.Helper()
	ob, op := befehl, programme
	t.Cleanup(func() { befehl, programme = ob, op })
	programme = prog
	befehl = func(dir string, g Gate) (string, []string, string) {
		return "sh", []string{"-c", skript(dir, g)}, ""
	}
}

func TestRunEingabeFehlt(t *testing.T) {
	mitBefehl(t, func(string, Gate) string { return "exit 0" }, "sh")
	var out, errb bytes.Buffer
	if got := run(&out, &errb, t.TempDir(), []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if out.Len() != 0 {
		t.Fatalf("kein Fall darf laufen, Ausgabe: %q", out.String())
	}
}

func TestRunProgrammFehlt(t *testing.T) {
	mitBefehl(t, func(string, Gate) string { return "exit 0" }, "gibt-es-nicht-gegenproben")
	var out, errb bytes.Buffer
	if got := run(&out, &errb, wurzelMitEingaben(t), []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if !strings.Contains(errb.String(), "nicht aufrufbar") {
		t.Fatalf("Meldung: %q", errb.String())
	}
}

func TestRunAlleFaelleLaufenZuEnde(t *testing.T) {
	mitBefehl(t, func(dir string, g Gate) string { return "exit 0" }, "sh")
	fs := []Fall{
		{Name: "verstoss-nicht-gemeldet", WantExit: 1},
		{Name: "erlaubt-a"},
		{Name: "erlaubt-b"},
	}
	var out, errb bytes.Buffer
	if got := run(&out, &errb, wurzelMitEingaben(t), fs); got != exitRot {
		t.Fatalf("Exit %d, erwartet %d", got, exitRot)
	}
	for _, n := range []string{"ROT  verstoss-nicht-gemeldet", "ok   erlaubt-a", "ok   erlaubt-b", "3 Fälle, 1 rot"} {
		if !strings.Contains(out.String(), n) {
			t.Fatalf("Ausgabe nennt %q nicht:\n%s", n, out.String())
		}
	}
}

func TestRunRaeumtWegwerfverzeichnisseAuf(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	mitBefehl(t, func(string, Gate) string { return "exit 1" }, "sh")
	var out, errb bytes.Buffer
	run(&out, &errb, wurzelMitEingaben(t), []Fall{{Name: "rot"}})
	rest, err := os.ReadDir(tmp)
	if err != nil || len(rest) != 0 {
		t.Fatalf("Wegwerfverzeichnisse bleiben: %v %v", rest, err)
	}
}

func TestRunErkenntVeraenderteEingabe(t *testing.T) {
	w := wurzelMitEingaben(t)
	mitBefehl(t, func(string, Gate) string { return "echo x >> " + filepath.Join(w, "go.mod") }, "sh")
	var out, errb bytes.Buffer
	if got := run(&out, &errb, w, []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if !strings.Contains(errb.String(), "verändert") {
		t.Fatalf("Meldung: %q", errb.String())
	}
}

func TestStarteBeendetDieProzessgruppeBeiFristablauf(t *testing.T) {
	start := time.Now()
	_, _, abgelaufen := starte(200*time.Millisecond, "", nil, "sh", "-c", "sleep 30 & wait")
	if !abgelaufen {
		t.Fatal("Frist nicht erkannt")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("Rückkehr nach %v, der Kindprozess hält die Ausgabe offen", d)
	}
}

func TestJederFallHatEigenesTmpdir(t *testing.T) {
	sammel := filepath.Join(t.TempDir(), "tmpdirs")
	mitBefehl(t, func(string, Gate) string { return `echo "$TMPDIR" >> ` + sammel }, "sh")
	fs := []Fall{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	var out, errb bytes.Buffer
	if got := run(&out, &errb, wurzelMitEingaben(t), fs); got != exitOK {
		t.Fatalf("Exit %d: %s%s", got, out.String(), errb.String())
	}
	b, err := os.ReadFile(sammel)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, z := range strings.Fields(string(b)) {
		if z == os.Getenv("TMPDIR") {
			t.Fatalf("Fall teilt das TMPDIR des Prozesses: %s", z)
		}
		seen[z] = true
	}
	if len(seen) != len(fs) {
		t.Fatalf("%d verschiedene TMPDIR bei %d Fällen: %v", len(seen), len(fs), seen)
	}
}

func TestHashEingabenGoSumIstOptional(t *testing.T) {
	w := wurzelMitEingaben(t)
	ohne, err := hashEingaben(w)
	if err != nil {
		t.Fatalf("ohne go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(w, "go.sum"), []byte("s\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mit, err := hashEingaben(w)
	if err != nil || mit == ohne {
		t.Fatalf("go.sum geht nicht in die Prüfsumme ein: %v", err)
	}
	if err := os.Remove(filepath.Join(w, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if _, err := hashEingaben(w); err == nil {
		t.Fatal("fehlendes go.mod bleibt unbemerkt")
	}
}
