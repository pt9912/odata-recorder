package admin_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// verboteneVergleiche sind die Aufrufe, die im Paket ohne _test.go nie stehen (SPEC-238 Nr. 4 e); strings.EqualFold
// steht allein mit dem Literal "Authorization" als Argument (der Filter von anfrage).
func verboteneVergleiche() []string {
	return []string{
		"bytes.Equal", "bytes.Compare", "bytes.EqualFold", "strings.Compare", "slices.Equal", "slices.Compare",
		"reflect.DeepEqual", "subtle.ConstantTimeEq",
	}
}

// aufrufName liefert Paket und Name eines Aufrufs der Form paket.name; sonst leere Zeichenketten.
func aufrufName(c *ast.CallExpr) (paket, name string) {
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", ""
	}
	return id.Name, sel.Sel.Name
}

// operandName liefert den Bezeichner oder das Feld, das ein Operand nennt, auch unter string(…), []byte(…), Klammern,
// Index und Ausschnitt; sonst eine leere Zeichenkette.
func operandName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return x.Sel.Name
	case *ast.ParenExpr:
		return operandName(x.X)
	case *ast.IndexExpr:
		return operandName(x.X)
	case *ast.SliceExpr:
		return operandName(x.X)
	case *ast.CallExpr:
		if len(x.Args) != 1 {
			return ""
		}
		if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "string" {
			return operandName(x.Args[0])
		}
		if at, ok := x.Fun.(*ast.ArrayType); ok && at.Len == nil {
			return operandName(x.Args[0])
		}
	}
	return ""
}

// tokenName meldet, ob ein Name das Token oder den Kandidaten nennt, ohne Beachtung der Schreibweise.
func tokenName(n string) bool {
	return strings.EqualFold(n, "token") || strings.EqualFold(n, "kandidat")
}

// quellVerstoesse liefert die Verstöße einer Quelle gegen SPEC-238 Nr. 4 e: ein verbotener Aufruf aus
// verboteneVergleiche, ein strings.EqualFold ohne das Literal "Authorization" als Argument, ein Ausdruck == oder !=
// mit einem Operanden, der das Token oder den Kandidaten nennt.
func quellVerstoesse(t *testing.T, name string, quelle any) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, 0)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var v []string
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			p, fn := aufrufName(x)
			schluessel := p + "." + fn
			if slices.Contains(verboteneVergleiche(), schluessel) {
				v = append(v, name+": "+schluessel)
			}
			if schluessel == "strings.EqualFold" && !hatLiteral(x, `"Authorization"`) {
				v = append(v, name+": strings.EqualFold ohne das Literal Authorization")
			}
		case *ast.BinaryExpr:
			if x.Op == token.EQL || x.Op == token.NEQ {
				if a, b := operandName(x.X), operandName(x.Y); tokenName(a) || tokenName(b) {
					v = append(v, name+": "+x.Op.String()+" mit "+a+" und "+b)
				}
			}
		}
		return true
	})
	return v
}

// hatLiteral meldet, ob eines der Argumente des Aufrufs das Literal ist.
func hatLiteral(c *ast.CallExpr, lit string) bool {
	for _, a := range c.Args {
		if b, ok := a.(*ast.BasicLit); ok && b.Value == lit {
			return true
		}
	}
	return false
}

// zaehleInFunktion zählt in der Funktion fn einer Quelle die Aufrufe von paket.name und liefert zugleich, wie viele
// Deklarationen dieses Namens es gibt.
func zaehleInFunktion(t *testing.T, quelle any, fn, paket, name string) (aufrufe, deklarationen int) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "quelle.go", quelle, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == fn {
			deklarationen++
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					if p, nm := aufrufName(c); p == paket && nm == name {
						aufrufe++
					}
				}
				return true
			})
		}
	}
	return aufrufe, deklarationen
}

// TestQuelleProbenAmZaehler hält den Zähler der Quelle gegen Proben: jeder verbotene Aufruf, jedes strings.EqualFold ohne
// das Literal und jeder Vergleich des Tokens oder des Kandidaten wird gemeldet, die erlaubten Formen nicht.
func TestQuelleProbenAmZaehler(t *testing.T) {
	for _, k := range verboteneVergleiche() {
		if v := quellVerstoesse(t, "probe.go", "package x\nfunc f(a, b []byte) bool { return "+k+"(a, b) }"); len(v) != 1 {
			t.Errorf("%s: %d Meldungen, erwartet 1", k, len(v))
		}
	}
	schlecht := []string{
		`strings.EqualFold(a, b)`, `a == token`, `kandidat != b`, `string(s.token) == b`, `b == string(s.Kandidat)`,
		`s.token[0] == b`, `(token) == b`, `string(x) == string(kandidat[1:])`,
	}
	for _, s := range schlecht {
		if v := quellVerstoesse(t, "probe.go", "package x\nfunc f() bool { return "+s+" }"); len(v) != 1 {
			t.Errorf("%s: %d Meldungen, erwartet 1", s, len(v))
		}
	}
	gut := []string{`strings.EqualFold(k, "Authorization")`, `a == b`, `len(token) == 32`, `tokenMin == 32`, `x.token < 3`}
	for _, s := range gut {
		if v := quellVerstoesse(t, "probe.go", "package x\nfunc f() bool { return "+s+" }"); len(v) != 0 {
			t.Errorf("%s: %v, erwartet keine Meldung", s, v)
		}
	}
	probe := "package x\nfunc vergleiche(a, b []byte) bool { subtle.ConstantTimeCompare(a, b); subtle.ConstantTimeCompare(a, b); sha256.Sum256(a); return true }\nfunc andere() { sha256.Sum256(nil) }"
	if n, d := zaehleInFunktion(t, probe, "vergleiche", "subtle", "ConstantTimeCompare"); n != 2 || d != 1 {
		t.Errorf("Zähler der Probe: %d Aufrufe in %d Deklarationen, erwartet 2 und 1", n, d)
	}
	if n, _ := zaehleInFunktion(t, probe, "vergleiche", "sha256", "Sum256"); n != 1 {
		t.Errorf("Zähler der Probe: %d Aufrufe von Sum256 in vergleiche, erwartet 1 (der in andere zählt nicht)", n)
	}
}

// TestVergleichNurUeberConstantTimeCompare hält U6: vergleiche ruft genau einmal subtle.ConstantTimeCompare und
// zweimal sha256.Sum256, und das Paket ohne _test.go enthält keinen der verbotenen Vergleiche und keinen Ausdruck
// == oder != mit dem Token oder dem Kandidaten. Die Zahl der geprüften Dateien ist die der Dateien ohne _test.go.
func TestVergleichNurUeberConstantTimeCompare(t *testing.T) {
	alle, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	geprueft, ohneTest := 0, 0
	for _, e := range eintraege {
		if n := e.Name(); strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			ohneTest++
		}
	}
	compare, summen, deklarationen := 0, 0, 0
	for _, d := range alle {
		if strings.HasSuffix(d, "_test.go") {
			continue
		}
		src, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		geprueft++
		for _, v := range quellVerstoesse(t, d, src) {
			t.Errorf("Verstoß: %s", v)
		}
		c, n := zaehleInFunktion(t, src, "vergleiche", "subtle", "ConstantTimeCompare")
		s, _ := zaehleInFunktion(t, src, "vergleiche", "sha256", "Sum256")
		compare, summen, deklarationen = compare+c, summen+s, deklarationen+n
	}
	if deklarationen != 1 || compare != 1 || summen != 2 {
		t.Errorf("vergleiche: %d Deklarationen, %d Aufrufe von ConstantTimeCompare, %d von Sum256, erwartet 1, 1 und 2", deklarationen, compare, summen)
	}
	if geprueft != ohneTest || geprueft < 3 {
		t.Errorf("%d Dateien geprüft, %d ohne _test.go, erwartet gleich und mindestens admin.go, route.go, tokendatei.go", geprueft, ohneTest)
	}
}
