package sqlite_test

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestSqliteFehlerAbbildung öffnet die vier Fehlerfälle von Open und bildet sie ab (SPEC-109 Nr. 7).
func TestSqliteFehlerAbbildung(t *testing.T) {
	dir := t.TempDir()
	neuer := filepath.Join(dir, "neuer.db")
	roh(t, neuer, "PRAGMA user_version = 2")
	fremd := filepath.Join(dir, "fremd.db")
	roh(t, fremd, "CREATE TABLE fremd (a INTEGER)")
	text := filepath.Join(dir, "text.db")
	if err := os.WriteFile(text, []byte(strings.Repeat("kein SQLite, nur Text. ", 100)), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		pfad  string
		code  model.Code
		besch bool
	}{
		{"Nicht-Datenbank", text, model.CodeE3001, true},
		{"neuere Version", neuer, model.CodeE3001, true},
		{"fremde Datenbank", fremd, model.CodeE3001, true},
		{"leerer Pfad", "", model.CodeE2001, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sqlite.Open(context.Background(), tc.pfad, sqlite.Options{})
			if err == nil {
				t.Fatal("kein Fehler")
			}
			got, ok := outbound.Abbilden(err, outbound.GegenstandSonstiges)
			if !ok || got != tc.code {
				t.Fatalf("Abbildung %v (%v), erwartet %v: %v", got, ok, tc.code, err)
			}
			if tc.besch != errors.Is(err, outbound.ErrBeschaedigt) {
				t.Errorf("ErrBeschaedigt des Pakets: %v", !tc.besch)
			}
		})
	}
}

// TestSqliteFehlerVollstaendig verlangt, dass jede Variable `Err…` des Adapters einen Fehler des
// Pakets outbound verpackt, den die Abbildung nennt (SPEC-109 Nr. 8).
func TestSqliteFehlerVollstaendig(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	bekannt := outbound.EintraegeAbbildung()
	fset := token.NewFileSet()
	gefunden := 0
	for _, e := range eintraege {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, n := range vs.Names {
					if !strings.HasPrefix(n.Name, "Err") {
						continue
					}
					gefunden++
					if ziel := verpacktesZiel(vs.Values, i); !slices.Contains(bekannt, ziel) {
						t.Errorf("%s verpackt keinen Fehler mit Eintrag in der Abbildung (gefunden: %q)", n.Name, ziel)
					}
				}
			}
		}
	}
	if gefunden < 4 {
		t.Fatalf("nur %d Variablen gefunden", gefunden)
	}
}

// verpacktesZiel liefert den Namen X aus `verpacke("…", outbound.X)`, sonst den leeren Text.
func verpacktesZiel(werte []ast.Expr, i int) string {
	if i >= len(werte) {
		return ""
	}
	call, ok := werte[i].(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		return ""
	}
	if fn, ok := call.Fun.(*ast.Ident); !ok || fn.Name != "verpacke" {
		return ""
	}
	sel, ok := call.Args[1].(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "outbound" {
		return ""
	}
	return sel.Sel.Name
}
