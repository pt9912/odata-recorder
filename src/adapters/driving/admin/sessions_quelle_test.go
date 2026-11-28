package admin_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"
)

// rahmenDateien sind die Dateien des Rahmens (SPEC-238), des Fensters des Abweisens (SPEC-301) und der Token-Datei
// (SPEC-239) ohne _test.go; jede andere Datei ohne _test.go im Paket gehört den Routen.
func rahmenDateien() []string {
	return []string{"abweisen_fenster.go", "admin.go", "fehler.go", "lebenszyklus.go", "route.go", "tokendatei.go"}
}

// sessionsDateien liefert die Dateien der Routen: jede Datei `.go` ohne `_test.go` im Verzeichnis außer denen des Rahmens
// (eine neue Datei eines Teils des Schnitts gehört dazu, ohne dass dieser Test sie nennt). Die Zählprüfung hält die Menge
// gegen das Verzeichnis: jede Datei des Rahmens besteht, die Dateien der Routen sind mindestens json.go, sessions.go,
// sessions_body.go, sessions_beenden.go und sessions_loeschen.go, und Rahmen plus Routen sind alle Dateien ohne _test.go.
func sessionsDateien(t *testing.T) []string {
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
	for _, r := range []string{"json.go", "sessions.go", "sessions_body.go", "sessions_beenden.go", "sessions_loeschen.go"} {
		if !slices.Contains(routen, r) {
			t.Errorf("die Datei %s der Routen fehlt", r)
		}
	}
	gleichAuf(t, "Rahmen plus Routen gleich alle Dateien", len(rahmenDateien())+len(routen), len(alle))
	return routen
}

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
// keine leitet den Kontext ab oder ersetzt ihn (`WithTimeout`, `WithDeadline`, `WithCancel`, `WithoutCancel`, `WithValue`, die
// drei `…Cause` und `AfterFunc`; das Verhalten der Ableitung fängt außerdem die Identität des `ctx` in `TestB2AufrufeUndKontext`).
// Die Proben halten die Zähler gegen eine Quelle mit je einer Stelle.
func TestS1S16QuelleDerRouten(t *testing.T) {
	namen := []string{"NeuesFehlerobjekt", "CountError", "ArtAdminAbgewiesen"}
	ableitungen := []string{"WithTimeout", "WithDeadline", "WithCancel", "WithoutCancel", "WithValue",
		"WithCancelCause", "WithTimeoutCause", "WithDeadlineCause", "AfterFunc"}
	gleichAuf(t, "Zahl der Namen", []int{len(namen), len(ableitungen)}, []int{3, 9})
	probe := "package x\nimport \"context\"\nfunc f() { model.NeuesFehlerobjekt(1); s.CountError(1); support.ArtAdminAbgewiesen; go g(); go func() {}(); " +
		"context.WithTimeout(); context.WithDeadline(); context.WithCancel(); context.WithoutCancel(); context.WithValue(); " +
		"context.WithCancelCause(); context.WithTimeoutCause(); context.WithDeadlineCause(); context.AfterFunc() }"
	for _, n := range namen {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "", n), 1)
	}
	for _, n := range ableitungen {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "context", n), 1)
	}
	gleichAuf(t, "Probe: Anweisungen go", zaehleGoAnweisungen(t, "probe.go", probe), 2)

	geprueft := 0
	for _, d := range sessionsDateien(t) {
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
	gleichAuf(t, "Zahl der geprüften Dateien gleich der der Dateien der Routen", geprueft, len(sessionsDateien(t)))
}
