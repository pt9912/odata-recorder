package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// listeBaum ist der Fixture-Baum mit einem verschachtelten Modul für `paket außerhalb des Moduls`.
func listeBaum(t *testing.T) string {
	t.Helper()
	d := fixtureBaum(t)
	schreibe(t, d, "tools/x/go.mod", "module x\n\ngo 1.27\n")
	schreibe(t, d, "tools/x/x.go", "package x\n")
	return d
}

func basis(klasse string) ml {
	return mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasse)
}

func befundeVon(t *testing.T, baum string, ms ...ml) []string {
	t.Helper()
	_, b := leseListe(listeJSON(t, ms...), baum, true)
	return b
}

// TestListeBefunde (SPEC-210 Nr. 2 d): je Feld und Form (fehlt, leer, unbekannt, fremd) meldet die Liste genau den erwarteten
// Befund oder keinen; die Orakel-Tabelle nennt jede Zeile einzeln.
func TestListeBefunde(t *testing.T) {
	baum := listeBaum(t)
	lang := strings.Repeat("n", 65)
	fehlt := struct{}{}
	zeilen := []struct {
		feld, form string
		wert       any
		klasse     string
		befund     string
	}{
		{"name", "fehlt", fehlt, klasseAndere, "Feld name fehlt"},
		{"name", "leer", "", klasseAndere, "name ungültig"},
		{"name", "unbekannt", "a b", klasseAndere, "name ungültig"},
		{"name", "fremd", lang, klasseAndere, "name ungültig"},
		{"name", "Rand 64", strings.Repeat("n", 64), klasseAndere, ""},
		{"datei", "fehlt", fehlt, klasseAndere, "Feld datei fehlt"},
		{"datei", "leer", "", klasseAndere, "datei ungültig"},
		{"datei", "unbekannt", "src/p/x.go", klasseAndere, "Datei fehlt"},
		{"datei", "fremd ..", "../x.go", klasseAndere, "datei ungültig"},
		{"datei", "fremd absolut", "/etc/x.go", klasseAndere, "datei ungültig"},
		{"datei", "fremd Endung", "src/p/a_test.go.txt", klasseAndere, "datei ungültig"},
		{"datei", "fremd Segment .", "src/./p/a.go", klasseAndere, "datei ungültig"},
		{"alt", "fehlt", fehlt, klasseAndere, "Feld alt fehlt"},
		{"alt", "leer", "", klasseAndere, "alt leer"},
		{"alt", "unbekannt", "zzz", klasseAndere, "alt trifft nicht"},
		{"alt", "fremd mehrfach", "return", klasseAndere, "alt trifft mehrfach"},
		{"alt", "fremd genau zweimal", "mu.", klasseAndere, "alt trifft mehrfach"},
		{"neu", "fehlt", fehlt, klasseAndere, "Feld neu fehlt"},
		{"neu", "leer", "", klasseAndere, ""},
		{"neu", "unbekannt gleich alt", "return a + b", klasseAndere, "neu ändert nichts"},
		{"test", "fehlt", fehlt, klasseAndere, "Feld test fehlt"},
		{"test", "leer", "", klasseAndere, "test ungültig"},
		{"test", "unbekannt", "KeinTest", klasseAndere, "test ungültig"},
		{"test", "fremd", "Test-X", klasseAndere, "test ungültig"},
		{"test", "Rand 128", "Test" + strings.Repeat("t", 124), klasseAndere, ""},
		{"test", "Rand 129", "Test" + strings.Repeat("t", 125), klasseAndere, "test ungültig"},
		{"klasse", "fehlt", fehlt, klasseAndere, "Feld klasse fehlt"},
		{"klasse", "leer", "", klasseAndere, "klasse unbekannt"},
		{"klasse", "unbekannt", "x", klasseAndere, "klasse unbekannt"},
		{"klasse", "fremd Groß", "Andere", klasseAndere, "klasse unbekannt"},
		{"klasse", "fremd ae", "nebenlaeufig", klasseAndere, "klasse unbekannt"},
		{"paket", "fehlt", fehlt, klasseAndere, ""},
		{"paket", "leer", "", klasseAndere, "paket ungültig"},
		{"paket", "unbekannt", "src/nichts", klasseAndere, "Paket fehlt"},
		{"paket", "fremd Modul", "tools/x", klasseAndere, "paket außerhalb des Moduls"},
		{"paket", "gültig", "src/p", klasseAndere, ""},
		{"n_detektor", "fehlt", fehlt, klasseNebenlaeufig, ""},
		{"n_detektor", "leer 0", 0, klasseNebenlaeufig, ""},
		{"n_detektor", "unbekannt 201", 201, klasseNebenlaeufig, "n_detektor ungültig"},
		{"n_detektor", "fremd -1", -1, klasseNebenlaeufig, "n_detektor ungültig"},
		{"n_detektor", "Rand 200", 200, klasseNebenlaeufig, ""},
		{"n_detektor", "bei andere", 5, klasseAndere, "Rate bei Klasse andere"},
		{"n_detektor", "bei andere 0", 0, klasseAndere, "Rate bei Klasse andere"},
		{"n_ohne", "fehlt", fehlt, klasseNebenlaeufig, ""},
		{"n_ohne", "leer 0", 0, klasseNebenlaeufig, ""},
		{"n_ohne", "unbekannt 201", 201, klasseNebenlaeufig, "n_ohne ungültig"},
		{"n_ohne", "fremd -1", -1, klasseNebenlaeufig, "n_ohne ungültig"},
		{"n_ohne", "Rand 200", 200, klasseNebenlaeufig, ""},
		{"n_ohne", "bei andere", 5, klasseAndere, "Rate bei Klasse andere"},
		{"x", "Feld unbekannt", 1, klasseAndere, "Feld x unbekannt"},
	}
	for _, z := range zeilen {
		t.Run(z.feld+"/"+z.form, func(t *testing.T) {
			m := basis(z.klasse)
			if z.wert == fehlt {
				delete(m, z.feld)
			} else {
				m[z.feld] = z.wert
			}
			name := "M"
			if z.feld == "name" && z.wert != fehlt {
				name = z.wert.(string)
			} else if z.feld == "name" {
				name = ""
			}
			var erwartet []string
			if z.befund != "" {
				erwartet = []string{fmt.Sprintf("Mutation 1 (%s): %s", name, z.befund)}
			}
			if got := befundeVon(t, baum, m); !reflect.DeepEqual(got, erwartet) {
				t.Errorf("Befunde %q statt %q", got, erwartet)
			}
		})
	}
}

// TestListeZusatztabelle (SPEC-210 Nr. 11): mehrere Verstöße am selben Schlüssel, Reihenfolge nach Mutation und Feld, jeder
// Befund einmal, auch wenn ein früherer vorliegt.
func TestListeZusatztabelle(t *testing.T) {
	baum := listeBaum(t)
	mit := func(m ml, k string, v any) ml { m[k] = v; return m }
	faelle := []struct {
		name string
		ms   []ml
		want []string
	}{
		{"datei fehlt und alt leer", []ml{mit(mit(basis(klasseAndere), "datei", "src/p/x.go"), "alt", "")},
			[]string{"Mutation 1 (M): Datei fehlt", "Mutation 1 (M): alt leer"}},
		{"name doppelt und klasse unbekannt", []ml{basis(klasseAndere), mit(basis("x"), "name", "M")},
			[]string{"Mutation 2 (M): Name doppelt", "Mutation 2 (M): klasse unbekannt"}},
		{"alt trifft nicht und neu gleich alt", []ml{mit(mit(basis(klasseAndere), "alt", "zzz"), "neu", "zzz")},
			[]string{"Mutation 1 (M): alt trifft nicht", "Mutation 1 (M): neu ändert nichts"}},
		{"datei fehlt und paket in fremdem Modul", []ml{mit(mit(basis(klasseAndere), "datei", "src/p/x.go"), "paket", "tools/x")},
			[]string{"Mutation 1 (M): Datei fehlt"}},
		{"nebenläufig 0 und 0", []ml{mit(mit(basis(klasseNebenlaeufig), "n_detektor", 0), "n_ohne", 0)},
			[]string{"Mutation 1 (M): Rate ohne Prozess"}},
		{"andere mit Rate und Datei fehlt", []ml{mit(mit(basis(klasseAndere), "n_detektor", 5), "datei", "src/p/x.go")},
			[]string{"Mutation 1 (M): Datei fehlt", "Mutation 1 (M): Rate bei Klasse andere"}},
		{"Reihenfolge nach Mutation vor Feld", []ml{mit(basis(klasseAndere), "test", "x"), mit(basis(klasseAndere), "name", "")},
			[]string{"Mutation 1 (M): test ungültig", "Mutation 2 (): name ungültig"}},
		{"alle Felder einer Mutation", []ml{{"name": "", "datei": "x", "alt": "", "neu": "", "test": "", "klasse": "x", "paket": "", "n_detektor": 300, "n_ohne": -1, "zz": 1}},
			[]string{"Mutation 1 (): name ungültig", "Mutation 1 (): datei ungültig", "Mutation 1 (): alt leer", "Mutation 1 (): test ungültig",
				"Mutation 1 (): klasse unbekannt", "Mutation 1 (): paket ungültig", "Mutation 1 (): n_detektor ungültig", "Mutation 1 (): n_ohne ungültig", "Mutation 1 (): Feld zz unbekannt"}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			if got := befundeVon(t, baum, f.ms...); !reflect.DeepEqual(got, f.want) {
				t.Errorf("Befunde\n got %q\nwant %q", got, f.want)
			}
		})
	}
}

// TestListeForm (SPEC-210 Nr. 2 c): Grenzen und Form der Liste.
func TestListeForm(t *testing.T) {
	baum := listeBaum(t)
	viele := func(n int) []byte {
		var ms []ml
		for i := range n {
			ms = append(ms, mut(fmt.Sprintf("m%d", i), "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere))
		}
		return listeJSON(t, ms...)
	}
	gueltig := string(viele(1))
	auffuellen := func(n int) []byte { return []byte(gueltig + strings.Repeat(" ", n-len(gueltig))) }
	faelle := []struct {
		name string
		roh  []byte
		want []string
	}{
		{"200 Mutationen", viele(200), nil},
		{"201 Mutationen", viele(201), []string{"Liste zu lang"}},
		{"leer", []byte(`{"mutationen":[]}`), []string{"Liste leer"}},
		{"kein JSON", []byte("{"), []string{"Liste nicht lesbar: kein JSON-Objekt"}},
		{"Wurzel kein Objekt", []byte("[]"), []string{"Liste nicht lesbar: kein JSON-Objekt"}},
		{"mutationen fehlt", []byte(`{}`), []string{"Liste nicht lesbar: Feld mutationen fehlt"}},
		{"null als Wurzel", []byte(`null`), []string{"Liste nicht lesbar: kein JSON-Objekt"}},
		{"null als Liste", []byte(`{"mutationen":null}`), []string{"Liste nicht lesbar: Feld mutationen: falscher Typ"}},
		{"null als Eintrag", []byte(`{"mutationen":[null]}`), []string{"Liste nicht lesbar: Mutation 1: kein Objekt"}},
		{"Feld im Wurzelobjekt", []byte(`{"mutationen":[],"extra":1}`), []string{"Liste nicht lesbar: Feld extra unbekannt"}},
		{"falscher Typ", []byte(`{"mutationen":[{"name":5}]}`), []string{"Liste nicht lesbar: Mutation 1 Feld name: falscher Typ"}},
		{"null", []byte(`{"mutationen":[{"name":null}]}`), []string{"Liste nicht lesbar: Mutation 1 Feld name: falscher Typ"}},
		{"kein Objekt", []byte(`{"mutationen":[3]}`), []string{"Liste nicht lesbar: Mutation 1: kein Objekt"}},
		{"BOM", append([]byte("\xef\xbb\xbf"), gueltig...), []string{"Liste nicht lesbar: BOM"}},
		{"kein UTF-8", []byte("{\"mutationen\":[{\"name\":\"\xff\"}]}"), []string{"Liste nicht lesbar: kein UTF-8"}},
		{"1 MiB", auffuellen(1 << 20), nil},
		{"1 MiB plus 1", auffuellen(1<<20 + 1), []string{"Liste nicht lesbar: größer als 1 MiB"}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			_, got := leseListe(f.roh, baum, true)
			if !reflect.DeepEqual(got, f.want) {
				t.Errorf("Befunde %q statt %q", got, f.want)
			}
		})
	}
	t.Run("ohne Compiler", func(t *testing.T) {
		_, got := leseListe(listeJSON(t, basis(klasseNebenlaeufig)), baum, false)
		if want := []string{"Mutation 1 (M): Detektor nicht verfügbar"}; !reflect.DeepEqual(got, want) {
			t.Errorf("Befunde %q statt %q", got, want)
		}
	})
	t.Run("Klasse als Rohbytes", func(t *testing.T) {
		if klasseNebenlaeufig != "nebenl\xc3\xa4ufig" {
			t.Errorf("Klasse %q ist nicht nebenl\\xc3\\xa4ufig", klasseNebenlaeufig)
		}
	})
}

// TestRunBefundeStartenNichts (SPEC-210 Nr. 2 d): bei einem Befund endet der Lauf mit Exit 2 und der Zeilenform von Nr. 2 d; es
// gibt keine Ergebniszeile und keine Kopie.
func TestRunBefundeStartenNichts(t *testing.T) {
	baum, ablage := listeBaum(t), t.TempDir()
	liste := filepath.Join(t.TempDir(), "liste.json")
	if err := os.WriteFile(liste, listeJSON(t, basis(klasseAndere), func() ml { m := basis(klasseAndere); m["name"] = "N"; m["test"] = "x"; return m }()), 0o644); err != nil {
		t.Fatal(err)
	}
	var aus, fehl bytes.Buffer
	code := run(t.Context(), []string{liste, baum, ablage, ""}, os.Environ(), &aus, &fehl)
	if e, _ := os.ReadDir(ablage); code != 2 || aus.Len() != 0 || len(e) != 0 ||
		fehl.String() != "mutationslauf: Eingabe: Mutation 2 (N): test ungültig\n" {
		t.Errorf("Exit %d, stdout %q, stderr %q, Ablage %v", code, aus.String(), fehl.String(), e)
	}
}
