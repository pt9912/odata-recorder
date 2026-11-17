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

// importErlaubt sagt, ob das Paket hybrid den Pfad importieren darf: nur die Standardpakete der Zulassungsliste, das
// Modell, die Inbound Ports und das Paket policy; kein anderes Standardpaket, kein Adapter, kein anderer Dienst
// (SPEC-284 Kopf). Grenze: cgo, unsafe, Reflexion und Aufrufe über bereits importierte Pakete erkennt die Liste nicht.
func importErlaubt(pfad string) bool {
	if !strings.HasPrefix(pfad, modulPfad) {
		return slices.Contains([]string{"bytes", "context", "errors", "io", "math", "slices", "strings"}, pfad)
	}
	rest := strings.TrimPrefix(pfad, modulPfad)
	return rest == "hexagon/model" || rest == "hexagon/ports/inbound" || rest == "hexagon/ports/outbound" ||
		rest == "hexagon/services/hybrid/policy"
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
			if !importErlaubt(pfad) {
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
					if len(d.Recv.List[0].Names) == 1 && d.Body != nil {
						for _, pos := range schreibtEmpfaenger(d.Body, d.Recv.List[0].Names[0].Name) {
							t.Errorf("%s: Methode %s schreibt in den Empfänger (%s)", name, d.Name.Name, fset.Position(pos))
						}
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

// schreibtEmpfaenger liefert die Stellen einer Zuweisung oder eines Inkrements, deren linke Seite über Feld, Index,
// Zeiger oder Klammer auf den Empfänger zurückgeht (`v.f = …`, `v.dienste[0].UpstreamURL = …`, `v.n++`): der Dienst
// ändert nach New nichts (SPEC-284 Nr. 13).
// Grenze: ein Schreibzugriff über einen Aufruf (`copy(v.dienste, …)`, `append` auf einem Alias, ein Zeiger, der aus
// dem Empfänger gebildet und weitergereicht wird) und über einen Alias (`d := v.dienste; d[0] = …`) bleibt ungesehen;
// ihn fängt der Wirkungs-Fänger TestGleichzeitigUpstream nur, wenn er das Ziel eines Aufrufs ändert.
func schreibtEmpfaenger(body *ast.BlockStmt, recv string) []token.Pos {
	var stellen []token.Pos
	wurzel := func(x ast.Expr) string {
		for {
			switch e := x.(type) {
			case *ast.SelectorExpr:
				x = e.X
			case *ast.IndexExpr:
				x = e.X
			case *ast.StarExpr:
				x = e.X
			case *ast.ParenExpr:
				x = e.X
			case *ast.Ident:
				return e.Name
			default:
				return ""
			}
		}
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			for _, l := range s.Lhs {
				if wurzel(l) == recv {
					stellen = append(stellen, s.Pos())
				}
			}
		case *ast.IncDecStmt:
			if wurzel(s.X) == recv {
				stellen = append(stellen, s.Pos())
			}
		}
		return true
	})
	return stellen
}

// TestSchreibtEmpfaenger fängt die Prüfung selbst: jede Form der linken Seite wird gefunden, Lesen und Schreiben auf
// lokale Namen nicht.
func TestSchreibtEmpfaenger(t *testing.T) {
	quelle := func(zeile string) *ast.BlockStmt {
		f, err := parser.ParseFile(token.NewFileSet(), "x.go", "package p\nfunc (v *T) m() {\n"+zeile+"\n}\n", parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		return f.Decls[0].(*ast.FuncDecl).Body
	}
	schlecht := []string{
		"v.f = 1", "v.dienste[0].UpstreamURL = z", "v.n++", "v.a.b.c = 1", "(*v).f = 1", "v.m[k] = 1", "v.f += 1", "a, v.f = 1, 2",
		"v.l[0]--",
	}
	gut := []string{"x := v.f", "x.f = v.f", "d := v.dienste[0]; d.URL = z", "w.f = 1", "x[v.i] = 1"}
	for _, z := range schlecht {
		if len(schreibtEmpfaenger(quelle(z), "v")) == 0 {
			t.Errorf("%q nicht gefunden", z)
		}
	}
	for _, z := range gut {
		if got := schreibtEmpfaenger(quelle(z), "v"); len(got) != 0 {
			t.Errorf("%q gemeldet", z)
		}
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
		"errors", "io", "slices", "math", "bytes", "context", "strings",
		modulPfad + "hexagon/model", modulPfad + "hexagon/ports/inbound", modulPfad + "hexagon/ports/outbound",
		modulPfad + "hexagon/services/hybrid/policy",
	}
	schlecht := []string{
		modulPfad + "hexagon/services/mock", modulPfad + "hexagon/services/replay", modulPfad + "hexagon/ports/outbound/outboundtest",
		modulPfad + "adapters/driven/sqlite", modulPfad + "adapters/driving/httpodata", "modernc.org/sqlite",
		"net", "net/http", "net/url", "os", "os/exec", "syscall", "sync", "sync/atomic", "time", "unsafe", "fmt", "regexp",
		"log", "reflect", "runtime", "plugin",
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
