//go:build unix

package admin_test

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// T27 (SPEC-239 Einleitung: „das Paket liest die Umgebung nie“). Menge: sieben Funktionen, die die Umgebung lesen
// (os.Getenv, LookupEnv, Environ, ExpandEnv, Expand, syscall.Getenv, Environ), über alle Quelldateien ohne _test.go,
// gefangen im Quelltext (go/parser) und in der Wirkung (gesetzte Umgebungsvariable ADMIN_TOKEN_FILE).
func TestKeineUmgebung(t *testing.T) {
	verboten := map[string][]string{
		"os":      {"Getenv", "LookupEnv", "Environ", "ExpandEnv", "Expand"},
		"syscall": {"Getenv", "Environ"},
	}
	n := 0
	for _, fs := range verboten {
		n += len(fs)
	}
	if n != 7 {
		t.Fatalf("Menge = %d Funktionen, erwartet 7", n)
	}
	es, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	dateien := 0
	for _, e := range es {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		dateien++
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			p, _ := strconv.Unquote(im.Path.Value)
			if _, ok := verboten[p]; ok && im.Name != nil {
				t.Errorf("%s: Import %q unter dem Namen %s", e.Name(), p, im.Name.Name)
			}
		}
		ast.Inspect(f, func(nd ast.Node) bool {
			sel, ok := nd.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			for _, name := range verboten[id.Name] {
				if sel.Sel.Name == name {
					t.Errorf("%s: %s.%s liest die Umgebung", e.Name(), id.Name, name)
				}
			}
			return true
		})
	}
	if dateien == 0 {
		t.Fatal("keine Quelldatei gefunden")
	}

	// Wirkung: eine gesetzte Variable ändert weder den leeren Pfad noch den Pfad der Datei.
	dir := t.TempDir()
	umgebung := filepath.Join(dir, "aus-der-umgebung")
	t.Setenv("ADMIN_TOKEN_FILE", umgebung)
	if _, err := admin.LadeToken(""); !errors.Is(err, model.ErrUngueltig) {
		t.Errorf("leerer Pfad mit gesetzter Umgebung: %v, erwartet model.ErrUngueltig", err)
	}
	if _, err := admin.LadeToken(filepath.Join(dir, "admin.token")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(umgebung); err == nil {
		t.Error("die Datei aus der Umgebung entstand")
	}
}

// T28 (Nr. 4: „LadeToken hält keinen Zustand“). Menge: fünf Änderungen der Datei zwischen zwei Aufrufen am selben
// Pfad (gelöscht, anderer gültiger Inhalt, Rechte, zu kurz, durch ein Verzeichnis ersetzt); jeder Aufruf liest neu.
func TestKeinZustand(t *testing.T) {
	andere := wdh("f", 40)
	faelle := []struct {
		name    string
		aendere func(pfad string)
		grund   string
		token   func(erst []byte) string
	}{
		{"gelöscht", func(p string) { _ = os.Remove(p) }, "", func(erst []byte) string { return "<neu>" + string(erst) }},
		{"anderer Inhalt", func(p string) { schreibe(t, p, andere, 0o600) }, "", func([]byte) string { return andere }},
		{"Rechte", func(p string) { _ = os.Chmod(p, 0o644) }, gRechte, nil},
		{"zu kurz", func(p string) { schreibe(t, p, "kurz", 0o600) }, gZuKurz, nil},
		{"Verzeichnis", func(p string) { _ = os.Remove(p); _ = os.Mkdir(p, 0o700) }, gVerzeichnis, nil},
	}
	if len(faelle) != 5 {
		t.Fatal("Menge")
	}
	for _, f := range faelle {
		pfad := filepath.Join(t.TempDir(), "admin.token")
		erst, err := admin.LadeToken(pfad)
		if err != nil {
			t.Fatal(err)
		}
		if again, err := admin.LadeToken(pfad); err != nil || string(again) != string(erst) {
			t.Fatalf("%s: zweiter Aufruf ohne Änderung: %q, %v", f.name, again, err)
		}
		f.aendere(pfad)
		zweit, err := admin.LadeToken(pfad)
		if f.grund != "" {
			erwarteGrund(t, f.name, err, f.grund)
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", f.name, err)
			continue
		}
		if string(zweit) == string(erst) {
			t.Errorf("%s: das alte Token %q gilt weiter", f.name, erst)
		}
		if want := f.token(erst); !strings.HasPrefix(want, "<neu>") && string(zweit) != want {
			t.Errorf("%s: Token = %q, erwartet %q", f.name, zweit, want)
		}
	}
}

// T29 (Nr. 3 a: 256 Bit). Menge: 1000 erzeugte Token; paarweise verschieden, keine der 32 Byte-Positionen konstant
// und keines der 496 Paare von Positionen in allen 1000 Token gleich (fängt Hälfte, Viertel und Spiegel).
// Grenze: Entropie ist nicht beweisbar; der Test fängt Abhängigkeit zwischen Positionen und Konstanten, nicht eine
// schwache, aber unabhängige Quelle. Falsch rot bei gleichverteilten Bytes: 256^-1000 je Position.
func TestEntropieDerPositionen(t *testing.T) {
	const n = 1000
	dir := t.TempDir()
	gesehen := map[string]bool{}
	var toks [][32]byte
	for i := 0; i < n; i++ {
		g, err := admin.LadeToken(filepath.Join(dir, "t"+strconv.Itoa(i)))
		if err != nil {
			t.Fatal(err)
		}
		if gesehen[string(g)] {
			t.Fatalf("Token doppelt: %q", g)
		}
		gesehen[string(g)] = true
		var b [32]byte
		for k := range b {
			v, err := strconv.ParseUint(string(g[2*k:2*k+2]), 16, 8)
			if err != nil {
				t.Fatal(err)
			}
			b[k] = byte(v)
		}
		toks = append(toks, b)
	}
	paare := 0
	for i := 0; i < 32; i++ {
		konstant := true
		for _, tk := range toks {
			if tk[i] != toks[0][i] {
				konstant = false
				break
			}
		}
		if konstant {
			t.Errorf("Position %d in %d Token konstant", i, n)
		}
		for j := i + 1; j < 32; j++ {
			paare++
			gleich := true
			for _, tk := range toks {
				if tk[i] != tk[j] {
					gleich = false
					break
				}
			}
			if gleich {
				t.Errorf("Positionen %d und %d in %d Token stets gleich", i, j, n)
			}
		}
	}
	if paare != 496 {
		t.Fatalf("%d Paare, erwartet 496", paare)
	}
}

// T30 (Nr. 3 d, 3 e): scheitert Link mit einem anderen Fehler als fs.ErrExist (der Haltepunkt entfernt die Hilfsdatei,
// Link meldet ENOENT), ist es Datei nicht anlegbar, nichts bleibt, und der Fehler beim Entfernen der schon fehlenden
// Hilfsdatei bleibt ohne Wirkung (kein Abbruch).
func TestLinkFehlerOhneErrExist(t *testing.T) {
	dir := t.TempDir()
	pfad := filepath.Join(dir, "admin.token")
	got, err := admin.LadeMitHaltepunkten(pfad, nil, func(h string) {
		if err := os.Remove(h); err != nil {
			t.Error(err)
		}
	})
	erwarteGrund(t, "Link scheitert", err, gNichtAnlegbar)
	if got != nil {
		t.Errorf("Token %q trotz Fehler", got)
	}
	if n := namen(t, dir); len(n) != 0 {
		t.Errorf("Verzeichnis = %v, erwartet leer", n)
	}
}

// T31 (Nr. 1 a): ein Symlink ohne Ziel führt nicht in das Erzeugen; es entsteht keine Hilfsdatei.
func TestSymlinkOhneZielOhneHilfsdatei(t *testing.T) {
	dir := t.TempDir()
	pfad := filepath.Join(dir, "admin.token")
	if err := os.Symlink(filepath.Join(dir, "nirgends"), pfad); err != nil {
		t.Fatal(err)
	}
	_, err := admin.LadeMitHaltepunkten(pfad, func(string) { t.Error("Hilfsdatei angelegt") }, func(string) { t.Error("Link versucht") })
	erwarteGrund(t, "Symlink ohne Ziel", err, gKeineDatei)
	if n := namen(t, dir); !gleich(n, []string{"admin.token"}) {
		t.Errorf("Verzeichnis = %v", n)
	}
}
