package sqlite_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"testing"
)

// Zählprüfung der Aufrufstellen (SPEC-108 Nr. 1, Kontext in der Transaktion): jedes Statement
// einer Operation in f läuft über die Verbindung der Transaktion, die den Kontext prüft. Der
// Test parst den Quelltext und hält je Datei die Zahl der Aufrufe von ExecContext, QueryContext
// und QueryRowContext fest (Exec, Query, QueryRow); eine Stelle mehr oder weniger ist rot. Ein
// Aufruf an der Verbindung vorbei (conn, db.sql) zählt nicht zu den Aufrufstellen und ist
// außerhalb von Coverage ein Fehler.

type stellen [3]int

func aufrufstellen() map[string]stellen {
	return map[string]stellen{
		"workspaces.go":          {4, 1, 3},
		"entities.go":            {2, 1, 1},
		"cursors.go":             {2, 1, 2},
		"workspace_snapshots.go": {1, 1, 1},
		"baselines.go":           {0, 1, 2},
		"zugriff.go":             {0, 0, 1},
	}
}

// steuerStatements sind die Statements der Steuerung der Transaktion und der Savepoints, je mit
// der Zahl ihrer Vorkommen als Literal in tx.go und zugriff.go.
func steuerStatements() map[string]int {
	return map[string]int{
		"BEGIN IMMEDIATE":       1,
		"ROLLBACK":              1,
		"COMMIT":                1,
		"SAVEPOINT operation":   1,
		"ROLLBACK TO operation": 1,
		"RELEASE operation":     2,
	}
}

func parseDatei(t *testing.T, name string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// vonVerbindung sagt, ob d eine Methode von zugriff oder txVerbindung ist: sie sind der Weg,
// den die Zählung der Aufrufstellen voraussetzt, und zählen nicht mit.
func vonVerbindung(d ast.Decl) bool {
	fd, ok := d.(*ast.FuncDecl)
	if !ok || fd.Recv == nil || len(fd.Recv.List) != 1 {
		return false
	}
	switch types.ExprString(fd.Recv.List[0].Type) {
	case "zugriff", "*txVerbindung":
		return true
	}
	return false
}

func TestAufrufstellenZaehlung(t *testing.T) {
	var summe stellen
	for datei, soll := range aufrufstellen() {
		var got stellen
		sql := 0
		for _, d := range parseDatei(t, datei).Decls {
			if vonVerbindung(d) {
				continue
			}
			ast.Inspect(d, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				empfaenger := types.ExprString(sel.X)
				art := map[string]int{"ExecContext": 0, "QueryContext": 1, "QueryRowContext": 2}
				i, istStelle := art[sel.Sel.Name]
				switch {
				case !istStelle:
				case strings.Contains(empfaenger, ".sql"):
					sql++
				case strings.Contains(empfaenger, "conn"):
					t.Errorf("%s: %s.%s an der Verbindung der Transaktion vorbei", datei, empfaenger, sel.Sel.Name)
				default:
					got[i]++
				}
				return true
			})
		}
		if got != soll {
			t.Errorf("%s: Aufrufstellen (Exec, Query, QueryRow) %v, erwartet %v", datei, got, soll)
		}
		// Coverage liest über die Datenbank, nicht in f.
		if want := map[string]int{"baselines.go": 2}[datei]; sql != want {
			t.Errorf("%s: %d Aufrufe über db.sql, erwartet %d", datei, sql, want)
		}
		for i := range summe {
			summe[i] += got[i]
		}
	}
	if summe != (stellen{9, 5, 10}) {
		t.Errorf("Summe der Aufrufstellen %v, erwartet [9 5 10]", summe)
	}
}

func TestSteuerStatementsZaehlung(t *testing.T) {
	got := map[string]int{}
	soll := steuerStatements()
	for _, datei := range []string{"tx.go", "zugriff.go"} {
		ast.Inspect(parseDatei(t, datei), func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if s, err := strconv.Unquote(lit.Value); err == nil {
					if _, istSteuer := soll[s]; istSteuer {
						got[s]++
					}
				}
			}
			return true
		})
	}
	n := 0
	for s, erwartet := range soll {
		if got[s] != erwartet {
			t.Errorf("Statement %q kommt %d-mal vor, erwartet %d", s, got[s], erwartet)
		}
		n += got[s]
	}
	if n != 7 {
		t.Errorf("%d Statements der Steuerung, erwartet 7", n)
	}
}
