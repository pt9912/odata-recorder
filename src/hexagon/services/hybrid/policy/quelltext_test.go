package policy_test

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

const modellPfad = "github.com/pt9912/odata-recorder/src/hexagon/model"

// importErlaubt sagt, ob das Paket policy den Pfad importieren darf: die Standardbibliothek ohne
// Umgebung, Zeit, Kontext, Sperren und Netz (außer net/url und net/netip), und von allem anderen
// allein model (SPEC-261 Nr. 9 c).
func importErlaubt(pfad string) bool {
	if pfad == modellPfad {
		return true
	}
	erstes, _, _ := strings.Cut(pfad, "/")
	if strings.Contains(erstes, ".") {
		return false
	}
	if slices.Contains([]string{"os", "sync", "time", "context", "net", "io/ioutil", "syscall", "unsafe", "plugin", "log"}, pfad) {
		return false
	}
	if strings.HasPrefix(pfad, "os/") || strings.HasPrefix(pfad, "sync/") || strings.HasPrefix(pfad, "net/") && pfad != "net/url" && pfad != "net/netip" {
		return false
	}
	return true
}

// TestQuelltext liest die Dateien des Pakets (ohne Tests) und prüft Importe, go-Statements,
// Variablen auf Paketebene und Methoden.
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
			if !importErlaubt(pfad) {
				t.Errorf("%s importiert %q", name, pfad)
			}
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.GenDecl:
				if d.Tok == token.VAR {
					t.Errorf("%s: Variable auf Paketebene", name)
				}
			case *ast.FuncDecl:
				funktionen++
				if d.Recv != nil {
					if typ := empfaenger(d.Recv.List[0].Type); typ != "Policy" && typ != "KonfigFehler" {
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
	if dateien < 3 || importe < 4 || funktionen < 8 {
		t.Fatalf("zu wenig gelesen: %d Dateien, %d Importe, %d Funktionen", dateien, importe, funktionen)
	}
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

// TestImportErlaubt fängt die Liste der Importe selbst: jede verbotene Form ist rot, die
// zulässigen sind grün.
func TestImportErlaubt(t *testing.T) {
	gut := []string{modellPfad, "errors", "fmt", "net/netip", "net/url", "strings", "strconv", "slices", "regexp"}
	schlecht := []string{
		"os", "os/exec", "sync", "sync/atomic", "time", "context", "net", "net/http", "net/netip/x",
		"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound",
		"github.com/pt9912/odata-recorder/src/hexagon/services/mock",
		"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite",
		"modernc.org/sqlite", "log", "unsafe", "syscall",
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
