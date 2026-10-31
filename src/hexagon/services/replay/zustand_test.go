package replay_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// erlaubteTypen sind die Typen des Pakets (SPEC-135 Nr. 17); Methoden trägt allein Replayer.
func erlaubteTypen() map[string]bool {
	return map[string]bool{"Replayer": true, "Konfig": true, "Dienst": true}
}

// befunde liest den Quelltext einer Datei und meldet jede Verletzung von SPEC-135 Nr. 17: Paketvariable (außer
// _) und init, go-Anweisung und AfterFunc, Import von sync und sync/atomic, Typ außer Replayer, Konfig und Dienst
// und Methode außer auf Replayer, Schreibzugriff auf die Konfig des Empfängers und Kopie des Empfängers.
func befunde(t *testing.T, name, quelle string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, quelle, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var b []string
	add := func(regel string) { b = append(b, name+": "+regel) }
	for _, imp := range f.Imports {
		if p, _ := strconv.Unquote(imp.Path.Value); p == "sync" || p == "sync/atomic" {
			add("Import " + p)
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			befundeGenDecl(d, add)
		case *ast.FuncDecl:
			befundeFunc(d, add)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.GoStmt:
			add("go-Anweisung")
		case *ast.SelectorExpr:
			if n.Sel.Name == "AfterFunc" {
				add("AfterFunc")
			}
		case *ast.AssignStmt:
			for _, l := range n.Lhs {
				if schreibtKonfig(l) {
					add("Schreibzugriff auf die Konfig")
				}
			}
			for _, r := range n.Rhs {
				if s, ok := r.(*ast.StarExpr); ok && istEmpfaenger(s.X) {
					add("Kopie des Empfängers")
				}
			}
		case *ast.IncDecStmt:
			if schreibtKonfig(n.X) {
				add("Schreibzugriff auf die Konfig")
			}
		}
		return true
	})
	return b
}

func befundeGenDecl(d *ast.GenDecl, add func(string)) {
	for _, s := range d.Specs {
		switch s := s.(type) {
		case *ast.ValueSpec:
			if d.Tok == token.VAR {
				for _, n := range s.Names {
					if n.Name != "_" {
						add("Paketvariable " + n.Name)
					}
				}
			}
		case *ast.TypeSpec:
			if !erlaubteTypen()[s.Name.Name] {
				add("Typ " + s.Name.Name)
			}
		}
	}
}

func befundeFunc(d *ast.FuncDecl, add func(string)) {
	if d.Recv == nil {
		if d.Name.Name == "init" {
			add("init")
		}
		return
	}
	typ := d.Recv.List[0].Type
	if s, ok := typ.(*ast.StarExpr); ok {
		typ = s.X
	}
	if id, ok := typ.(*ast.Ident); !ok || id.Name != "Replayer" {
		add("Methode " + d.Name.Name + " auf einem anderen Typ")
	} else if _, ptr := d.Recv.List[0].Type.(*ast.StarExpr); !ptr {
		add("Empfänger als Wert: Kopie")
	}
}

func istEmpfaenger(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && (id.Name == "r" || id.Name == "v")
}

// schreibtKonfig meldet eine Zuweisung an ein Feld der Konfig: r.k.X oder tiefer.
func schreibtKonfig(e ast.Expr) bool {
	for {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			if x.Sel.Name == "k" && istEmpfaenger(x.X) {
				return true
			}
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		default:
			return false
		}
	}
}

// TestPaketOhneZustand hält Z54: der Quelltext des Pakets (ohne _test.go) bleibt ohne Befund; die fünf Regeln
// melden je einen Verstoß und lassen den Gutfall durch (Menge 10).
func TestPaketOhneZustand(t *testing.T) {
	verstoesse := map[string]string{
		"Paketvariable":  "package p\nvar zaehler int\n",
		"go-Anweisung":   "package p\nfunc f() { go func() {}() }\n",
		"sync":           "package p\nimport \"sync\"\ntype Replayer struct{ mu sync.Mutex }\n",
		"Typ":            "package p\ntype Hilfe struct{}\nfunc (h Hilfe) M() {}\n",
		"Schreibzugriff": "package p\ntype Replayer struct{ k struct{ S string } }\nfunc (r *Replayer) F() { r.k.S = \"x\" }\n",
	}
	gut := map[string]string{
		"Paketvariable":  "package p\nvar _ int\nconst c = 1\n",
		"go-Anweisung":   "package p\nfunc f() { func() {}() }\n",
		"sync":           "package p\nimport \"strings\"\nvar _ = strings.ToLower\n",
		"Typ":            "package p\ntype Replayer struct{}\nfunc (r *Replayer) M() {}\ntype Konfig struct{}\n",
		"Schreibzugriff": "package p\ntype Replayer struct{ k struct{ S string } }\nfunc (r *Replayer) F() string { s := r.k.S; s = s + \"x\"; return s }\n",
	}
	for name, q := range verstoesse {
		if len(befunde(t, name+".go", q)) == 0 {
			t.Errorf("Regel %s: Verstoß nicht gemeldet", name)
		}
		if b := befunde(t, name+"_gut.go", gut[name]); len(b) != 0 {
			t.Errorf("Regel %s: Gutfall gemeldet: %v", name, b)
		}
	}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	gelesen := 0
	for _, e := range eintraege {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		gelesen++
		src, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range befunde(t, e.Name(), string(src)) {
			t.Errorf("Zustand im Paket: %s", b)
		}
	}
	if gelesen < 6 {
		t.Errorf("nur %d Quelldateien gelesen", gelesen)
	}
}
