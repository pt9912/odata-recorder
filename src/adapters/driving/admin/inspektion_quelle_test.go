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

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// inspDateien liefert die Dateien der Inspektion im Verzeichnis: jede Datei `.go` ohne `_test.go` mit dem Präfix `inspektion`.
func inspDateien(t *testing.T) []string {
	t.Helper()
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range eintraege {
		if n := e.Name(); strings.HasPrefix(n, "inspektion") && strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			out = append(out, n)
		}
	}
	return out
}

// TestI17QuelleDerInspektion hält I17 und I1 am Quelltext (SPEC-293 Nr. 1, 2; SPEC-238 Nr. 15 b): in den Dateien der Inspektion ruft und
// nennt keine Stelle `NeuesFehlerobjekt`, `CountError`, `ArtAdminAbgewiesen` oder `Append`, keine startet eine Goroutine, keine leitet
// den Kontext ab oder ersetzt ihn (`WithTimeout`, `WithDeadline`, `WithCancel`, `WithoutCancel`, `WithValue`), und keine importiert das
// Paket der Querschnitts-Ports (`AuditLog` und `Telemetry` liegen dort). Die Dateien sind die drei des Plans und gehören zur
// Dateimenge des Tests von Teil 1 (`sessionsDateien`). Die Proben halten die Zähler gegen eine Quelle mit je einer Stelle.
func TestI17QuelleDerInspektion(t *testing.T) {
	namen := []string{"NeuesFehlerobjekt", "CountError", "ArtAdminAbgewiesen", "Append"}
	ableitungen := []string{"WithTimeout", "WithDeadline", "WithCancel", "WithoutCancel", "WithValue"}
	gleichAuf(t, "Zahl der Namen", []int{len(namen), len(ableitungen)}, []int{4, 5})
	probe := "package x\nimport \"context\"\nfunc f() { model.NeuesFehlerobjekt(1); s.CountError(1); support.ArtAdminAbgewiesen; a.Append(1); go g(); " +
		"context.WithTimeout(); context.WithDeadline(); context.WithCancel(); context.WithoutCancel(); context.WithValue() }"
	for _, n := range namen {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "", n), 1)
	}
	for _, n := range ableitungen {
		gleichAuf(t, "Probe "+n, zaehleSelektoren(t, "probe.go", probe, "context", n), 1)
	}
	gleichAuf(t, "Probe: Anweisungen go", zaehleGoAnweisungen(t, "probe.go", probe), 1)

	dateien := inspDateien(t)
	gleichAuf(t, "Dateien der Inspektion", dateien, []string{"inspektion.go", "inspektion_json.go", "inspektion_query.go"})
	for _, d := range dateien {
		if !slices.Contains(sessionsDateien(t), d) {
			t.Errorf("%s fehlt in der Dateimenge der Routen von Teil 1", d)
		}
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
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
		f, err := parser.ParseFile(token.NewFileSet(), d, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			if p, _ := strconv.Unquote(im.Path.Value); strings.HasSuffix(p, "/ports/support") {
				t.Errorf("%s importiert %s", d, p)
			}
		}
	}
}

// TestI17GesamtzaehlungAllerTeile hält die Zählprüfung von M-Audit über alle Teile gegen die Quelle: die 33 Zellen der Inspektion
// und die Zellen der Routen der Sessions (die Zahl der Zellen von `fehlerZellen` mit je einem Erfolg), mit Eintrag allein die Zellen
// einer Route `POST` von SessionRouten (jede schreibt bei Erfolg einen Eintrag), alle anderen ohne. Stand mit den Routen von Teil 1
// und 3: 24 + 33 = 57 Zellen, genau 1 mit Eintrag.
func TestI17GesamtzaehlungAllerTeile(t *testing.T) {
	eigene, andere := 0, 0
	for _, fs := range inspFehlerZellen() {
		eigene += len(fs) + 1
	}
	for _, fs := range fehlerZellen() {
		andere += len(fs) + 1
	}
	_, u := konfigMit()
	routen, err := admin.SessionRouten(&sessionsFake{}, u.audit)
	if err != nil {
		t.Fatal(err)
	}
	mitEintrag := 0
	for _, r := range routen {
		if r.Methode == "POST" {
			mitEintrag++
		}
	}
	gleichAuf(t, "Zellen der Inspektion, der Sessions, gesamt", []int{eigene, andere, eigene + andere}, []int{33, 24, 57})
	gleichAuf(t, "Zellen mit Eintrag und ohne", []int{mitEintrag, eigene + andere - mitEintrag}, []int{1, 56})
}

// TestI18BinaryUnveraendert hält I18 (SPEC-293 Nr. 12) am Quelltext: keine Datei des Pakets bootstrap (auch keine `_test.go`)
// importiert das Paket admin oder nennt InspektionRouten; der Teil führt weder Fehlercode noch Schlüssel noch Kommando ein (die
// benannte Ausgabe dafür ist `make katalog-check`). Die Zahl der geprüften Dateien ist die der Dateien des Pakets.
func TestI18BinaryUnveraendert(t *testing.T) {
	dateien, err := filepath.Glob(filepath.Join("..", "..", "..", "bootstrap", "*.go"))
	if err != nil || len(dateien) < 10 {
		t.Fatalf("%d Dateien von bootstrap, Fehler %v", len(dateien), err)
	}
	for _, d := range dateien {
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		if text := string(src); strings.Contains(text, "/driving/admin\"") || strings.Contains(text, "InspektionRouten") {
			t.Errorf("%s importiert admin oder nennt InspektionRouten", d)
		}
	}
}
