package hybrid_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const modulPfad = "github.com/pt9912/odata-recorder/src/"

// importErlaubt sagt, ob das Paket hybrid den Pfad importieren darf: die Standardbibliothek, das Modell, die Inbound
// Ports und das Paket policy; kein Adapter, kein anderer Dienst (SPEC-284 Kopf).
func importErlaubt(pfad string) bool {
	if !strings.HasPrefix(pfad, modulPfad) {
		erstes, _, _ := strings.Cut(pfad, "/")
		return !strings.Contains(erstes, ".")
	}
	rest := strings.TrimPrefix(pfad, modulPfad)
	return rest == "hexagon/model" || rest == "hexagon/ports/inbound" || rest == "hexagon/services/hybrid/policy"
}

// TestQuelltext liest die Dateien des Pakets (ohne Tests) und prüft Importe, go-Statements, Variablen auf Paketebene
// und Methoden: kein Adapter, kein anderer Dienst, keine Goroutine, kein Zustand, keine Sperre, Methoden nur auf
// Verteiler und den Lesern über den Puffer (SPEC-284 Nr. 13).
//
// Grenze: gelesen wird nur das Wurzelverzeichnis des Pakets; die Zählprüfung hat feste Untergrenzen (4 Dateien, 8 Importe,
// 8 Funktionen) und meldet zu wenig Gelesenes als Fehler.
func TestQuelltext(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	dateien, importe, funktionen := 0, 0, 0
	for _, e := range eintraege {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		dateien++
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			pfad, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			importe++
			if !importErlaubt(pfad) || slices.Contains([]string{"sync", "sync/atomic", "os", "net/http", "time", "unsafe"}, pfad) {
				t.Errorf("%s importiert %q", name, pfad)
			}
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.GenDecl:
				if d.Tok == token.VAR && !pruefeBlankVar(d) {
					t.Errorf("%s: Variable auf Paketebene", name)
				}
			case *ast.FuncDecl:
				funktionen++
				if d.Recv != nil {
					if typ := empfaenger(d.Recv.List[0].Type); typ != "Verteiler" && typ != "fehlerLeser" {
						t.Errorf("%s: Methode %s auf %q", name, d.Name.Name, typ)
					}
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if _, ok := n.(*ast.GoStmt); ok {
				t.Errorf("%s: go-Statement", name)
			}
			return true
		})
	}
	if dateien < 4 || importe < 8 || funktionen < 8 {
		t.Fatalf("zu wenig gelesen: %d Dateien, %d Importe, %d Funktionen", dateien, importe, funktionen)
	}
}

// pruefeBlankVar sagt, ob alle Namen der Deklaration `_` sind (die Zusicherung `var _ Port = ...`).
func pruefeBlankVar(d *ast.GenDecl) bool {
	for _, s := range d.Specs {
		for _, n := range s.(*ast.ValueSpec).Names {
			if n.Name != "_" {
				return false
			}
		}
	}
	return true
}

func empfaenger(x ast.Expr) string {
	switch x := x.(type) {
	case *ast.StarExpr:
		return empfaenger(x.X)
	case *ast.Ident:
		return x.Name
	}
	return "?"
}

// TestImportErlaubt fängt die Liste der Importe selbst.
func TestImportErlaubt(t *testing.T) {
	gut := []string{
		"errors", "io", "slices", "math", "bytes",
		modulPfad + "hexagon/model", modulPfad + "hexagon/ports/inbound", modulPfad + "hexagon/services/hybrid/policy",
	}
	schlecht := []string{
		modulPfad + "hexagon/services/mock", modulPfad + "hexagon/services/replay", modulPfad + "hexagon/ports/outbound",
		modulPfad + "adapters/driven/sqlite", modulPfad + "adapters/driving/httpodata", "modernc.org/sqlite",
	}
	for _, p := range gut {
		if !importErlaubt(p) {
			t.Errorf("%q abgelehnt", p)
		}
	}
	for _, p := range schlecht {
		if importErlaubt(p) {
			t.Errorf("%q erlaubt", p)
		}
	}
}
