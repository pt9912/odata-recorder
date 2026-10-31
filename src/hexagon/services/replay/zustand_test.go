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
// und Methode außer auf Replayer, Schreibzugriff auf die Konfig des Empfängers, Kopie des Empfängers und jedes Feld
// des Typs Replayer außer der Konfig k.
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
			if s.Name.Name == "Replayer" && !nurKonfig(s.Type) {
				add("Replayer hält mehr als die Konfig")
			}
		}
	}
}

// nurKonfig sagt, ob t ein struct mit genau einem benannten Feld k vom Typ Konfig ist: der Dienst hält nach New nur
// seine Konfig (SPEC-135 Nr. 17).
func nurKonfig(t ast.Expr) bool {
	st, ok := t.(*ast.StructType)
	if !ok || len(st.Fields.List) != 1 {
		return false
	}
	f := st.Fields.List[0]
	typ, ok := f.Type.(*ast.Ident)
	return ok && typ.Name == "Konfig" && len(f.Names) == 1 && f.Names[0].Name == "k"
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

// TestReplayerGestalt hält Z54 für den Zustand in der Instanz (SPEC-135 Nr. 17: der Dienst hält nach New nur seine
// Konfig): Menge 10 = neun Gestalten des Typs Replayer, die mehr oder anderes halten als das eine Feld k vom Typ Konfig
// (ein zweites Feld, eine Map, kein Feld, anderer Name, anderer Typ als Zeiger und als anderer Name, eingebettet,
// zwei Namen in einem Feld, kein struct), und der Gutfall.
func TestReplayerGestalt(t *testing.T) {
	const kopf = "package p\ntype Konfig struct{}\n"
	verstoesse := map[string]string{
		"zweites Feld":    "type Replayer struct {\n\tk Konfig\n\tn int\n}\n",
		"Map":             "type Replayer struct {\n\tk     Konfig\n\tcache map[string]int\n}\n",
		"kein Feld":       "type Replayer struct{}\n",
		"anderer Name":    "type Replayer struct{ c Konfig }\n",
		"Zeiger":          "type Replayer struct{ k *Konfig }\n",
		"anderer Typname": "type Replayer struct{ k int }\n",
		"eingebettet":     "type Replayer struct{ Konfig }\n",
		"zwei Namen":      "type Replayer struct{ k, k2 Konfig }\n",
		"kein struct":     "type Replayer int\n",
	}
	for name, q := range verstoesse {
		gemeldet := false
		for _, b := range befunde(t, name+".go", kopf+q) {
			gemeldet = gemeldet || strings.Contains(b, "Replayer hält mehr als die Konfig")
		}
		if !gemeldet {
			t.Errorf("Gestalt %q: Verstoß nicht gemeldet", name)
		}
	}
	if b := befunde(t, "gut.go", kopf+"type Replayer struct{ k Konfig }\n"); len(b) != 0 {
		t.Errorf("Gutfall gemeldet: %v", b)
	}
	if len(verstoesse) != 9 {
		t.Errorf("Menge: %d Verstöße, erwartet 9", len(verstoesse))
	}
}

// TestPaketOhneZustand hält Z54: der Quelltext des Pakets (ohne _test.go) bleibt ohne Befund; die sechs Regeln
// melden je einen Verstoß und lassen den Gutfall durch (Menge 12).
func TestPaketOhneZustand(t *testing.T) {
	const kopf = "package p\ntype Konfig struct{ S string }\n"
	verstoesse := map[string]string{
		"Paketvariable":  "package p\nvar zaehler int\n",
		"go-Anweisung":   "package p\nfunc f() { go func() {}() }\n",
		"sync":           "package p\nimport \"sync\"\nvar _ sync.Mutex\n",
		"Typ":            "package p\ntype Hilfe struct{}\nfunc (h Hilfe) M() {}\n",
		"Schreibzugriff": kopf + "type Replayer struct{ k Konfig }\nfunc (r *Replayer) F() { r.k.S = \"x\" }\n",
		"Feld":           kopf + "type Replayer struct {\n\tk Konfig\n\tn int\n}\n",
	}
	gut := map[string]string{
		"Paketvariable":  "package p\nvar _ int\nconst c = 1\n",
		"go-Anweisung":   "package p\nfunc f() { func() {}() }\n",
		"sync":           "package p\nimport \"strings\"\nvar _ = strings.ToLower\n",
		"Typ":            kopf + "type Replayer struct{ k Konfig }\nfunc (r *Replayer) M() {}\n",
		"Schreibzugriff": kopf + "type Replayer struct{ k Konfig }\nfunc (r *Replayer) F() string { s := r.k.S; s = s + \"x\"; return s }\n",
		"Feld":           kopf + "type Replayer struct{ k Konfig }\n",
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
