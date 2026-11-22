package admin_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// importsRoutenDateien liefert die Dateien der Routen: jede Datei `.go` ohne `_test.go` im Verzeichnis außer denen des
// Rahmens (die Dateien eines weiteren Teils des Schnitts gehören dazu, ohne dass dieser Test sie nennt). Die Zählprüfung
// hält die Menge gegen das Verzeichnis: Rahmen plus Routen sind alle Dateien ohne _test.go, jede Datei des Rahmens besteht,
// und die Dateien der Routen sind mindestens json.go, sessions.go, imports.go und imports_body.go.
func importsRoutenDateien(t *testing.T) []string {
	t.Helper()
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var alle, routen []string
	for _, e := range eintraege {
		if n := e.Name(); strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			alle = append(alle, n)
			if !slices.Contains(rahmenDateien(), n) {
				routen = append(routen, n)
			}
		}
	}
	for _, r := range rahmenDateien() {
		if !slices.Contains(alle, r) {
			t.Errorf("die Datei %s des Rahmens fehlt im Verzeichnis", r)
		}
	}
	for _, r := range []string{"json.go", "sessions.go", "imports.go", "imports_body.go"} {
		if !slices.Contains(routen, r) {
			t.Errorf("die Datei %s der Routen fehlt", r)
		}
	}
	gleichAuf(t, "Rahmen plus Routen gleich alle Dateien", len(rahmenDateien())+len(routen), len(alle))
	return routen
}

// zaehleImporte zählt in einer Quelle die Importe des Pfades pfad.
func zaehleImporte(t *testing.T, name string, quelle any, pfad string) int {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	n := 0
	for _, im := range f.Imports {
		if p, _ := strconv.Unquote(im.Path.Value); p == pfad {
			n++
		}
	}
	return n
}

// TestIM15QuelleDerRouten hält IM15 (SPEC-297 Nr. 2, 9 a; SPEC-238 Nr. 15 b, c) am Quelltext: in den Dateien der Routen ruft
// und nennt keine Stelle `NeuesFehlerobjekt` oder `CountError` (2 Namen) oder eine der Arten `ArtAdminAbgewiesen`,
// `ArtImportFehlgeschlagen`, `ArtSessionGeloescht`, `ArtSessionAufbewahrung`, `ArtWorkspaceGeloescht` (5 Namen), keine startet
// eine Goroutine, keine leitet den Kontext ab oder löst ihn ab (`WithCancel`, `WithTimeout`, `WithDeadline`, `WithoutCancel`,
// `WithCancelCause`, `WithTimeoutCause`, `WithDeadlineCause`, `AfterFunc`: 8 Namen); in den Dateien `imports*.go` liest keine
// Stelle `Header`, `RawQuery` oder `Laenge` (3 Namen), keine ruft `Err`, `Done` oder `Deadline` auf einem Kontext oder
// `context.Cause` (4 Namen) und keine importiert `sync` oder `sync/atomic` (2 Pfade). Der Quelltext sagt dazu nur, was diese
// Dateien nicht schreiben; was ein Kontext nach dem Erfolg oder ein Zustand zwischen Requests bewirkt, fangen Tests mit Wirkung.
// Die Proben halten jeden Zähler gegen eine Quelle mit je einer Stelle.
func TestIM15QuelleDerRouten(t *testing.T) {
	zaehler := []string{"NeuesFehlerobjekt", "CountError"}
	arten := []string{"ArtAdminAbgewiesen", "ArtImportFehlgeschlagen", "ArtSessionGeloescht", "ArtSessionAufbewahrung", "ArtWorkspaceGeloescht"}
	ableitungen := []string{"WithCancel", "WithTimeout", "WithDeadline", "WithoutCancel", "WithCancelCause", "WithTimeoutCause", "WithDeadlineCause", "AfterFunc"}
	felder := []string{"Header", "RawQuery", "Laenge", "Err", "Done", "Deadline"}
	gleichAuf(t, "Zahl der Namen", []int{len(zaehler), len(arten), len(ableitungen), len(felder)}, []int{2, 5, 8, 6})

	probe := "package x\nimport \"context\"\nfunc f() { model.NeuesFehlerobjekt(1); s.CountError(1); support.ArtAdminAbgewiesen; support.ArtImportFehlgeschlagen; " +
		"support.ArtSessionGeloescht; support.ArtSessionAufbewahrung; support.ArtWorkspaceGeloescht; go g(); go func() {}(); " +
		"context.WithCancel(); context.WithTimeout(); context.WithDeadline(); context.WithoutCancel(); context.WithCancelCause(); " +
		"context.WithTimeoutCause(); context.WithDeadlineCause(); context.AfterFunc(); a.Header; a.RawQuery; a.Laenge; " +
		"ctx.Err(); ctx.Done(); ctx.Deadline(); context.Cause(ctx) }"
	for _, n := range slices.Concat(zaehler, arten, felder) {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "", n), 1)
	}
	for _, n := range append(ableitungen, "Cause") {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "context", n), 1)
	}
	gleichAuf(t, "Probe: Anweisungen go", zaehleGoAnweisungen(t, "probe.go", probe), 2)
	zustand := []string{"sync", "sync/atomic"}
	for _, p := range zustand {
		gleichAuf(t, "Probe: Import "+p, zaehleImporte(t, "probe.go", "package x\nimport (\n\"sync\"\n\"sync/atomic\"\n\"syncx\"\n)\n", p), 1)
	}

	geprueft := 0
	for _, d := range importsRoutenDateien(t) {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		geprueft++
		for _, n := range slices.Concat(zaehler, arten) {
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
		if !strings.HasPrefix(d, "imports") {
			continue
		}
		for _, n := range felder {
			if z := zaehleSelektoren(t, d, src, "", n); z != 0 {
				t.Errorf("%s: %d Zugriffe auf %s", d, z, n)
			}
		}
		if z := zaehleSelektoren(t, d, src, "context", "Cause"); z != 0 {
			t.Errorf("%s: %d Aufrufe von context.Cause", d, z)
		}
		for _, p := range zustand {
			if z := zaehleImporte(t, d, src, p); z != 0 {
				t.Errorf("%s: %d Importe von %s", d, z, p)
			}
		}
	}
	gleichAuf(t, "Zahl der geprüften Dateien gleich der der Dateien der Routen", geprueft, len(importsRoutenDateien(t)))
}

// TestIM16BinaryUnveraendert hält IM16 (SPEC-297 Nr. 13) am Quelltext: keine Datei des Pakets bootstrap (auch keine _test.go)
// importiert admin oder nennt ImportRouten; der Katalog und das Handbuch sind kein Gegenstand dieses Tests (`make
// katalog-check`), und `git diff --stat` über `src/bootstrap` ist in diesem Teil leer. Die Zahl der geprüften Dateien ist die
// der Dateien des Pakets.
func TestIM16BinaryUnveraendert(t *testing.T) {
	dateien, err := filepath.Glob(filepath.Join("..", "..", "..", "bootstrap", "*.go"))
	if err != nil || len(dateien) < 10 {
		t.Fatalf("%d Dateien von bootstrap, Fehler %v", len(dateien), err)
	}
	probe := "package x\nimport \"p/driving/admin\"\nfunc f() { admin.ImportRouten(a, b) }"
	gleichAuf(t, "Probe: Nennungen von ImportRouten", zaehleSelektoren(t, "probe.go", probe, "", "ImportRouten"), 1)
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
		if z := zaehleSelektoren(t, d, src, "", "ImportRouten"); z != 0 {
			t.Errorf("%s: %d Nennungen von ImportRouten", d, z)
		}
	}
}
