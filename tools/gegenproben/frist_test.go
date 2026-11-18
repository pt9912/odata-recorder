package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"
)

// Randform 6: die Frist je Prozess ist 120 s; gleich, nicht größer oder kleiner.
func TestFallFristIst120Sekunden(t *testing.T) {
	if fallFrist != 120*time.Second {
		t.Fatalf("fallFrist = %v, erwartet 120s", fallFrist)
	}
}

// Randform 6: jeder Aufruf von starte und starteStroeme in den Quelldateien des Pakets (ohne
// Tests) nennt fallFrist als erstes Argument. Die Menge entsteht aus dem Quelltext; die
// Weitergabe der Frist im Rumpf von starte selbst ist kein Aufruf einer Stelle und zählt nicht.
// Findet die Schleife weniger als sieben Stellen, ist der Test rot.
func TestJederAufrufNenntDieFallFrist(t *testing.T) {
	const mindestens = 7
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	stellen := 0
	for _, e := range eintraege {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		datei, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range datei.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "starte" {
				continue
			}
			ast.Inspect(d, func(knoten ast.Node) bool {
				c, ok := knoten.(*ast.CallExpr)
				if !ok {
					return true
				}
				fn, ok := c.Fun.(*ast.Ident)
				if !ok || (fn.Name != "starte" && fn.Name != "starteStroeme") {
					return true
				}
				stellen++
				erstes := ""
				if len(c.Args) > 0 {
					if id, istName := c.Args[0].(*ast.Ident); istName {
						erstes = id.Name
					}
				}
				if erstes != "fallFrist" {
					t.Errorf("%s: Aufruf von %s nennt fallFrist nicht als erstes Argument", fset.Position(c.Pos()), fn.Name)
				}
				return true
			})
		}
	}
	if stellen < mindestens {
		t.Fatalf("%d Aufrufstellen gefunden, erwartet mindestens %d", stellen, mindestens)
	}
}
