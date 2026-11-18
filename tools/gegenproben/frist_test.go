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

// Randform 6: jeder Aufruf, dessen Funktion als Bezeichner starte oder starteStroeme geschrieben
// ist, in den Quelldateien des Pakets (ohne Tests) nennt fallFrist als erstes Argument. Die Menge
// entsteht aus dem Quelltext; die Weitergabe der Frist im Rumpf von starte selbst zählt nicht.
// Findet die Schleife weniger als sieben Stellen, ist der Test rot. Ungeprüft: ein Aufruf über
// einen Funktionswert, ein im Rumpf überdecktes fallFrist und ein Prozess, der nicht über starte
// oder starteStroeme startet.
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

// aufrufeVon liefert die Aufrufe im Rumpf, deren Funktion als Bezeichner name geschrieben ist.
func aufrufeVon(rumpf ast.Node, name string) []*ast.CallExpr {
	var erg []*ast.CallExpr
	ast.Inspect(rumpf, func(k ast.Node) bool {
		if c, ok := k.(*ast.CallExpr); ok {
			if fn, istName := c.Fun.(*ast.Ident); istName && fn.Name == name {
				erg = append(erg, c)
			}
		}
		return true
	})
	return erg
}

// zuweisungenVon liefert die Zuweisungen im Rumpf, deren einzige rechte Seite ein Aufruf von name ist.
func zuweisungenVon(rumpf ast.Node, name string) []*ast.AssignStmt {
	var erg []*ast.AssignStmt
	ast.Inspect(rumpf, func(k ast.Node) bool {
		if a, ok := k.(*ast.AssignStmt); ok && len(a.Rhs) == 1 {
			if c, istAufruf := a.Rhs[0].(*ast.CallExpr); istAufruf {
				if fn, istName := c.Fun.(*ast.Ident); istName && fn.Name == name {
					erg = append(erg, a)
				}
			}
		}
		return true
	})
	return erg
}

// letzterName gibt den Namen der letzten linken Seite der Zuweisung zurück, "" bei einem Nicht-Bezeichner.
func letzterName(a *ast.AssignStmt) string {
	if id, ok := a.Lhs[len(a.Lhs)-1].(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// Randform 6: in fuehreFall ist das letzte Argument von bewerteBau und von bewerte der Name, an den
// der Aufruf von starte bzw. starteStroeme den Ablauf bindet; der Name steht je Paar nur an diesen
// zwei Stellen, und fuehreFall ruft jede der beiden Funktionen einmal. Ungeprüft: die Wertung
// selbst (TestBewerteBau, TestBewerte) und die Weitergabe in anderen Funktionen als fuehreFall.
func TestFuehreFallReichtDenAblaufWeiter(t *testing.T) {
	paare := []struct{ name, start, wertung string }{
		{"Bau", "starte", "bewerteBau"},
		{"Gate", "starteStroeme", "bewerte"},
	}
	datei, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var rumpf *ast.BlockStmt
	for _, d := range datei.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "fuehreFall" {
			rumpf = fd.Body
		}
	}
	if rumpf == nil {
		t.Fatal("fuehreFall nicht gefunden")
	}
	// Je Name zwei Stellen je Paar, das ihn bindet: die Bindung und das Argument der Wertung.
	erwartet := map[string]int{}
	for _, p := range paare {
		if zs := zuweisungenVon(rumpf, p.start); len(zs) == 1 {
			erwartet[letzterName(zs[0])] += 2
		}
	}
	for _, p := range paare {
		t.Run(p.name, func(t *testing.T) {
			if n := len(aufrufeVon(rumpf, p.start)); n != 1 {
				t.Fatalf("fuehreFall ruft %s %d-mal, erwartet 1", p.start, n)
			}
			zs := zuweisungenVon(rumpf, p.start)
			if len(zs) != 1 {
				t.Fatalf("%d Zuweisungen binden das Ergebnis von %s, erwartet 1", len(zs), p.start)
			}
			name := letzterName(zs[0])
			if name == "" || name == "_" {
				t.Fatalf("der Aufruf von %s bindet den Ablauf an %q", p.start, name)
			}
			ws := aufrufeVon(rumpf, p.wertung)
			if len(ws) != 1 {
				t.Fatalf("fuehreFall ruft %s %d-mal, erwartet 1", p.wertung, len(ws))
			}
			letztes, ok := ws[0].Args[len(ws[0].Args)-1].(*ast.Ident)
			if !ok || letztes.Name != name {
				t.Fatalf("das letzte Argument von %s ist nicht der Bezeichner %s", p.wertung, name)
			}
			vorkommen := 0
			ast.Inspect(rumpf, func(k ast.Node) bool {
				if id, istName := k.(*ast.Ident); istName && id.Name == name {
					vorkommen++
				}
				return true
			})
			if vorkommen != erwartet[name] {
				t.Fatalf("%s steht %d-mal in fuehreFall, erwartet %d (Bindung und Argument je Paar)", name, vorkommen, erwartet[name])
			}
		})
	}
}
