package admin_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// sessionsDateien sind die Dateien der Routen der Sessions ohne _test.go (SPEC-292): die Teile des Schnitts ergänzen die
// Liste um ihre Dateien.
func sessionsDateien() []string { return []string{"json.go", "sessions.go", "sessions_body.go"} }

// zaehleGoAnweisungen zählt die Anweisungen `go` in einer Quelle.
func zaehleGoAnweisungen(t *testing.T, name string, quelle any) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, 0)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	n := 0
	ast.Inspect(f, func(x ast.Node) bool {
		if _, ok := x.(*ast.GoStmt); ok {
			n++
		}
		return true
	})
	return n
}

// TestS1S16QuelleDerRouten hält S1, S2 und S16 am Quelltext (SPEC-292 Nr. 1, 2; SPEC-238 Nr. 15 b): in den Dateien der Routen
// ruft und nennt keine Stelle `NeuesFehlerobjekt`, `CountError` oder `ArtAdminAbgewiesen`, keine startet eine Goroutine, und
// keine leitet den Kontext ab oder ersetzt ihn (`WithTimeout`, `WithDeadline`, `WithCancel`, `WithoutCancel`, `WithValue`).
// Die Proben halten die Zähler gegen eine Quelle mit je einer Stelle.
func TestS1S16QuelleDerRouten(t *testing.T) {
	namen := []string{"NeuesFehlerobjekt", "CountError", "ArtAdminAbgewiesen"}
	ableitungen := []string{"WithTimeout", "WithDeadline", "WithCancel", "WithoutCancel", "WithValue"}
	gleichAuf(t, "Zahl der Namen", []int{len(namen), len(ableitungen)}, []int{3, 5})
	probe := "package x\nimport \"context\"\nfunc f() { model.NeuesFehlerobjekt(1); s.CountError(1); support.ArtAdminAbgewiesen; go g(); go func() {}(); " +
		"context.WithTimeout(); context.WithDeadline(); context.WithCancel(); context.WithoutCancel(); context.WithValue() }"
	for _, n := range namen {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "", n), 1)
	}
	for _, n := range ableitungen {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "context", n), 1)
	}
	gleichAuf(t, "Probe: Anweisungen go", zaehleGoAnweisungen(t, "probe.go", probe), 2)

	geprueft := 0
	for _, d := range sessionsDateien() {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		geprueft++
		for _, n := range namen {
			if z := zaehleSelektoren(t, d, src, "", n); z != 0 {
				t.Errorf("%s: %d Nennungen von %s", d, z, n)
			}
		}
		for _, n := range ableitungen {
			if z := zaehleSelektoren(t, d, src, "context", n); z != 0 {
				t.Errorf("%s: %d Aufrufe von context.%s", d, z, n)
			}
		}
		if z := zaehleGoAnweisungen(t, d, src); z != 0 {
			t.Errorf("%s: %d Anweisungen go", d, z)
		}
	}
	gleichAuf(t, "Zahl der geprüften Dateien", geprueft, 3)
}

// TestS17BinaryUnveraendert hält S17 (SPEC-292 Nr. 14) am Quelltext: keine Datei des Pakets bootstrap (auch keine _test.go)
// importiert admin oder nennt SessionRouten. Die Zahl der geprüften Dateien ist die der Dateien des Pakets.
func TestS17BinaryUnveraendert(t *testing.T) {
	dateien, err := filepath.Glob(filepath.Join("..", "..", "..", "bootstrap", "*.go"))
	if err != nil || len(dateien) < 10 {
		t.Fatalf("%d Dateien von bootstrap, Fehler %v", len(dateien), err)
	}
	for _, d := range dateien {
		f, err := parser.ParseFile(token.NewFileSet(), d, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			if p, _ := strconv.Unquote(im.Path.Value); strings.HasSuffix(p, "/driving/admin") {
				t.Errorf("%s importiert %s", d, p)
			}
		}
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		if z := zaehleSelektoren(t, d, src, "", "SessionRouten"); z != 0 {
			t.Errorf("%s: %d Nennungen von SessionRouten", d, z)
		}
	}
}
