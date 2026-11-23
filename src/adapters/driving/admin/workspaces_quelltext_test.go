package admin_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// wsDateien liefert die Dateien der Routen dieses Endpunkts (`workspaces*.go` ohne `_test.go`) und prüft die Zählung gegen
// das Verzeichnis: jede Datei `.go` ohne `_test.go` ist eine des Rahmens (rahmenDateien), eine der übrigen Endpunkte (`json.go`
// und die Präfixe `sessions`, `inspektion`, `imports`, die der Test nennt) oder eine dieser Routen; eine Datei, die keiner der
// drei Mengen angehört, ist ein Fehler (der Test nennt sie nicht).
func wsDateien(t *testing.T) []string {
	t.Helper()
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var alle, rahmen, uebrige, routen []string
	for _, e := range eintraege {
		n := e.Name()
		if !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		alle = append(alle, n)
		switch {
		case slices.Contains(rahmenDateien(), n):
			rahmen = append(rahmen, n)
		case strings.HasPrefix(n, "workspaces"):
			routen = append(routen, n)
		case n == "json.go" || strings.HasPrefix(n, "sessions") || strings.HasPrefix(n, "inspektion") || strings.HasPrefix(n, "imports"):
			uebrige = append(uebrige, n)
		default:
			t.Errorf("die Datei %s gehört keiner Menge des Tests an (Rahmen, übrige Endpunkte, Routen der Workspaces)", n)
		}
	}
	for _, r := range []string{"workspaces.go", "workspaces_body.go", "workspaces_json.go"} {
		if !slices.Contains(routen, r) {
			t.Errorf("die Datei %s der Routen fehlt", r)
		}
	}
	gleichAuf(t, "Rahmen plus Endpunkte plus Routen gleich alle Dateien", len(rahmen)+len(uebrige)+len(routen), len(alle))
	return routen
}

// wsImporte liefert die Importpfade einer Quelle.
func wsImporte(t *testing.T, name string, quelle any) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var out []string
	for _, im := range f.Imports {
		p, _ := strconv.Unquote(im.Path.Value)
		out = append(out, p)
	}
	return out
}

// wsAufrufstellen zählt je Operation die Aufrufe der Form `<x>.workspaces.<Name>(…)` und `<x>.snapshots.<Name>(…)`, Schlüssel
// `workspaces.List` und so weiter.
func wsAufrufstellen(t *testing.T, name string, quelle any) map[string]int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, 0)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	out := map[string]int{}
	ast.Inspect(f, func(x ast.Node) bool {
		c, ok := x.(*ast.CallExpr)
		if !ok {
			return true
		}
		if op, ok := c.Fun.(*ast.SelectorExpr); ok {
			if feld, ok := op.X.(*ast.SelectorExpr); ok && (feld.Sel.Name == "workspaces" || feld.Sel.Name == "snapshots") {
				out[feld.Sel.Name+"."+op.Sel.Name]++
			}
		}
		return true
	})
	return out
}

// TestWS15Quelltext hält WS15 (SPEC-298 Nr. 2, 9 a, 12; SPEC-238 Nr. 15 b, c) am Quelltext der Routen (`workspaces*.go`): 0
// Aufrufe von `NeuesFehlerobjekt` und `CountError`, 0 Nennungen von fünf Arten des Audit (`ArtAdminAbgewiesen`,
// `ArtImportFehlgeschlagen`, `ArtSessionGeloescht`, `ArtSessionAufbewahrung`, `ArtWorkspaceGeloescht`) und von drei fremden
// (`ArtSessionAngelegt`, `ArtSessionBeendet`, `ArtImportVeroeffentlicht`), 0 Anweisungen `go`, 0 Aufrufe von `time.AfterFunc`
// (startet eine Goroutine ohne `go`), 0 Aufrufe von acht Ableitungen eines Kontextes (darunter `context.AfterFunc`), 0 Aufrufe von `Err`, `Done`, `Deadline` und `context.Cause`, 0 Importe von `sync` und `sync/atomic`, 0
// Zugriffe auf `Header`, `RawQuery` und `Laenge`, und jede der acht Operationen hat genau eine Aufrufstelle. Die Proben halten
// die Zähler gegen eine Quelle mit je einer Stelle.
func TestWS15Quelltext(t *testing.T) {
	wsFrist(t)
	namen := []string{"NeuesFehlerobjekt", "CountError"}
	arten := []string{
		"ArtAdminAbgewiesen", "ArtImportFehlgeschlagen", "ArtSessionGeloescht", "ArtSessionAufbewahrung", "ArtWorkspaceGeloescht",
		"ArtSessionAngelegt", "ArtSessionBeendet", "ArtImportVeroeffentlicht",
	}
	ableitungen := []string{"WithCancel", "WithTimeout", "WithDeadline", "WithoutCancel", "WithCancelCause", "WithTimeoutCause", "WithDeadlineCause", "AfterFunc"}
	zustand := []string{"Err", "Done", "Deadline"}
	anfrage := []string{"Header", "RawQuery", "Laenge"}
	importe := []string{"sync", "sync/atomic"}
	gleichAuf(t, "Zahl der Namen", []int{len(namen), len(arten), len(ableitungen), len(zustand) + 1, len(anfrage), len(importe)}, []int{2, 8, 8, 4, 3, 2})

	probe := "package x\nimport (\n\"context\"\n\"sync\"\n\"sync/atomic\"\n\"time\"\n)\nfunc f() { model.NeuesFehlerobjekt(1); s.CountError(1); go g(); go func() {}(); " +
		"support.ArtAdminAbgewiesen; support.ArtImportFehlgeschlagen; support.ArtSessionGeloescht; support.ArtSessionAufbewahrung; " +
		"support.ArtWorkspaceGeloescht; support.ArtSessionAngelegt; support.ArtSessionBeendet; support.ArtImportVeroeffentlicht; " +
		"context.WithCancel(); context.WithTimeout(); context.WithDeadline(); context.WithoutCancel(); context.WithCancelCause(); " +
		"context.WithTimeoutCause(); context.WithDeadlineCause(); context.AfterFunc(); ctx.Err(); ctx.Done(); ctx.Deadline(); " +
		"context.Cause(ctx); time.AfterFunc(0, f); a.Header; a.RawQuery; a.Laenge; h.workspaces.List(ctx); h.snapshots.List(ctx) }"
	for _, n := range slices.Concat(namen, arten, zustand, anfrage) {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "", n), 1)
	}
	for _, n := range append(slices.Clone(ableitungen), "Cause") {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "context", n), 1)
	}
	gleichAuf(t, "Probe time.AfterFunc", zaehleSelektoren(t, "probe.go", probe, "time", "AfterFunc"), 1)
	gleichAuf(t, "Probe: Anweisungen go", zaehleGoAnweisungen(t, "probe.go", probe), 2)
	gleichAuf(t, "Probe: Importe", wsImporte(t, "probe.go", probe), []string{"context", "sync", "sync/atomic", "time"})
	gleichAuf(t, "Probe: Aufrufstellen", wsAufrufstellen(t, "probe.go", probe), map[string]int{"workspaces.List": 1, "snapshots.List": 1})

	stellen := map[string]int{}
	dateien := wsDateien(t)
	for _, d := range dateien {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range slices.Concat(namen, arten, zustand, anfrage) {
			if z := zaehleSelektoren(t, d, src, "", n); z != 0 {
				t.Errorf("%s: %d Nennungen von %s", d, z, n)
			}
		}
		for _, n := range append(slices.Clone(ableitungen), "Cause") {
			if z := zaehleSelektoren(t, d, src, "context", n); z != 0 {
				t.Errorf("%s: %d Aufrufe von context.%s", d, z, n)
			}
		}
		if z := zaehleSelektoren(t, d, src, "time", "AfterFunc"); z != 0 {
			t.Errorf("%s: %d Aufrufe von time.AfterFunc (startet eine Goroutine ohne Anweisung go)", d, z)
		}
		if z := zaehleGoAnweisungen(t, d, src); z != 0 {
			t.Errorf("%s: %d Anweisungen go", d, z)
		}
		for _, p := range wsImporte(t, d, src) {
			if slices.Contains(importe, p) {
				t.Errorf("%s importiert %s", d, p)
			}
		}
		for k, v := range wsAufrufstellen(t, d, src) {
			stellen[k] += v
		}
	}
	gleichAuf(t, "Aufrufstellen je Operation", stellen, map[string]int{
		"workspaces.List": 1, "workspaces.Create": 1, "workspaces.Get": 1, "workspaces.Delete": 1, "workspaces.Reset": 1, "workspaces.Restore": 1,
		"snapshots.List": 1, "snapshots.Create": 1,
	})
	gleichAuf(t, "Zahl der Aufrufstellen", len(stellen), 8)
}

// TestWS15BinaryUnveraendert hält SPEC-298 Nr. 14 am Quelltext: keine Datei des Pakets bootstrap (auch keine _test.go)
// importiert admin oder nennt WorkspaceRouten; die Prüfung des Verzeichnisses `src/bootstrap` gegen die Basis (leer in diesem
// Teil) ist Handarbeit des Handoffs (`git diff --stat`), kein Test.
func TestWS15BinaryUnveraendert(t *testing.T) {
	wsFrist(t)
	dateien, err := filepath.Glob(filepath.Join("..", "..", "..", "bootstrap", "*.go"))
	if err != nil || len(dateien) < 10 {
		t.Fatalf("%d Dateien von bootstrap, Fehler %v", len(dateien), err)
	}
	for _, d := range dateien {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range wsImporte(t, d, src) {
			if strings.HasSuffix(p, "/driving/admin") {
				t.Errorf("%s importiert %s", d, p)
			}
		}
		if z := zaehleSelektoren(t, d, src, "", "WorkspaceRouten"); z != 0 {
			t.Errorf("%s: %d Nennungen von WorkspaceRouten", d, z)
		}
	}
}
