package workspace_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// quelltextBefunde liest die Dateien des Pakets ohne _test.go (das Arbeitsverzeichnis eines Tests ist das
// Paketverzeichnis) und nennt jeden Verstoß gegen die Zusage „kein Zustand, keine Sperre, keine Goroutine“
// (SPEC-129 Nr. 14): eine Paketvariable außer `_`, eine Funktion init, eine go-Anweisung, den Import von sync
// oder sync/atomic, und eine Verwaltung, die mehr als das Feld Konfig hat. Es liefert auch die Zahl der Dateien.
func quelltextBefunde(t *testing.T, verzeichnis string) ([]string, int) {
	t.Helper()
	dateien, err := filepath.Glob(filepath.Join(verzeichnis, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var befunde []string
	gelesen := 0
	fset := token.NewFileSet()
	for _, d := range dateien {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(fset, d, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		gelesen++
		befunde = append(befunde, dateiBefunde(fset, f)...)
	}
	return befunde, gelesen
}

func dateiBefunde(fset *token.FileSet, f *ast.File) []string {
	var b []string
	ort := func(p token.Pos) string { return fset.Position(p).String() }
	for _, imp := range f.Imports {
		if p := strings.Trim(imp.Path.Value, `"`); p == "sync" || p == "sync/atomic" {
			b = append(b, ort(imp.Pos())+": Import von "+p)
		}
	}
	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.GenDecl:
			b = append(b, genBefunde(fset, x)...)
		case *ast.FuncDecl:
			if x.Recv == nil && x.Name.Name == "init" {
				b = append(b, ort(x.Pos())+": Funktion init")
			}
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if g, ok := n.(*ast.GoStmt); ok {
			b = append(b, ort(g.Pos())+": go-Anweisung")
		}
		return true
	})
	return b
}

func genBefunde(fset *token.FileSet, g *ast.GenDecl) []string {
	var b []string
	for _, s := range g.Specs {
		switch x := s.(type) {
		case *ast.ValueSpec:
			if g.Tok != token.VAR {
				continue
			}
			for _, n := range x.Names {
				if n.Name != "_" {
					b = append(b, fset.Position(n.Pos()).String()+": Paketvariable "+n.Name)
				}
			}
		case *ast.TypeSpec:
			st, ok := x.Type.(*ast.StructType)
			if x.Name.Name != "Verwaltung" || !ok {
				continue
			}
			if len(st.Fields.List) != 1 || len(st.Fields.List[0].Names) != 1 || st.Fields.List[0].Names[0].Name != "k" {
				b = append(b, fset.Position(x.Pos()).String()+": Verwaltung hat nicht genau das Feld k")
			} else if id, ok := st.Fields.List[0].Type.(*ast.Ident); !ok || id.Name != "Konfig" {
				b = append(b, fset.Position(x.Pos()).String()+": das Feld von Verwaltung ist keine Konfig")
			}
		}
	}
	return b
}

// TestPaketOhneZustand (Z57, Z47, Z04): der Quelltext des Pakets hat keine Paketvariable außer `_`, keine Funktion
// init, keine go-Anweisung, keinen Import von sync und sync/atomic, und Verwaltung hat genau das Feld Konfig.
// Der Test ist deterministisch (keine Goroutine, keine Zeit) und fängt die ganze Klasse, auch Orte, die kein Lauf
// aus zwei Goroutinen berührt.
func TestPaketOhneZustand(t *testing.T) {
	befunde, dateien := quelltextBefunde(t, ".")
	if dateien != 7 {
		t.Errorf("%d Dateien ohne _test.go gelesen, erwartet 7", dateien)
	}
	for _, b := range befunde {
		t.Error(b)
	}
}
