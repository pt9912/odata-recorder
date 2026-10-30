package workspace_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// erlaubteImporte sind die Pakete, die der Dienst importieren darf: Pakete der Standardbibliothek ohne Zustand und die
// Domain und Ports des Repos (SPEC-129 *Ungeprüft*). Jeder andere Pfad ist ein Befund; ein neuer Import ist eine
// Zeile hier und im Diff sichtbar.
func erlaubteImporte() map[string]bool {
	pfad := "github.com/pt9912/odata-recorder/src/hexagon/"
	m := map[string]bool{}
	for _, p := range []string{"context", "encoding/json", "errors", "fmt", "io", "log/slog", "strconv", "unicode/utf8",
		pfad + "model", pfad + "ports/inbound", pfad + "ports/outbound", pfad + "ports/support"} {
		m[p] = true
	}
	return m
}

// erlaubteTypen sind die Typen des Pakets (SPEC-129 Nr. 14).
func erlaubteTypen() map[string]bool {
	return map[string]bool{"Verwaltung": true, "Konfig": true, "eintrag": true}
}

// zustandsTyp sagt, ob der Typausdruck Verwaltung, Konfig oder ein Zeiger darauf ist.
func zustandsTyp(e ast.Expr) bool {
	if s, ok := e.(*ast.StarExpr); ok {
		e = s.X
	}
	id, ok := e.(*ast.Ident)
	return ok && (id.Name == "Verwaltung" || id.Name == "Konfig")
}

// wurzel schält Selektor, Index, Stern, Klammer und Ausschnitt ab und liefert die Wurzel und ob etwas abgeschält wurde.
func wurzel(e ast.Expr) (*ast.Ident, bool) {
	abgeschaelt := false
	for {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		case *ast.ParenExpr:
			e = x.X
		case *ast.SliceExpr:
			e = x.X
		case *ast.Ident:
			return x, abgeschaelt
		default:
			return nil, false
		}
		abgeschaelt = true
	}
}

// zustandsNamen sind die Namen von Empfänger und Parametern der Funktion, deren Typ Verwaltung oder Konfig ist.
func zustandsNamen(fd *ast.FuncDecl) map[string]bool {
	m := map[string]bool{}
	listen := []*ast.FieldList{fd.Type.Params}
	if fd.Recv != nil {
		listen = append(listen, fd.Recv)
	}
	for _, l := range listen {
		if l == nil {
			continue
		}
		for _, f := range l.List {
			if zustandsTyp(f.Type) {
				for _, n := range f.Names {
					m[n.Name] = true
				}
			}
		}
	}
	return m
}

// quelltextBefunde liest die Dateien des Pakets ohne _test.go (das Arbeitsverzeichnis eines Tests ist das
// Paketverzeichnis) und nennt jeden Verstoß gegen die Zusage „kein Zustand, keine Sperre, keine Goroutine“ (SPEC-129
// Nr. 14); es liefert auch die Zahl der Dateien.
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

// dateiBefunde wendet die Regeln auf eine Datei an: Import-Liste, init, go, AfterFunc, Paketvariable, Typen und
// Methoden, Feld von Verwaltung, Zuweisung in New, Schreibzugriff auf Verwaltung und Konfig, Kopie von v und v.k.
func dateiBefunde(fset *token.FileSet, f *ast.File) []string {
	var b []string
	ort := func(p token.Pos) string { return fset.Position(p).String() }
	for _, imp := range f.Imports {
		if p, err := strconv.Unquote(imp.Path.Value); err != nil || !erlaubteImporte()[p] {
			b = append(b, ort(imp.Pos())+": Import "+imp.Path.Value+" ist nicht in der Liste")
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GoStmt:
			b = append(b, ort(x.Pos())+": go-Anweisung")
		case *ast.SelectorExpr:
			if x.Sel.Name == "AfterFunc" {
				b = append(b, ort(x.Pos())+": AfterFunc startet eine Goroutine")
			}
		case *ast.TypeSpec:
			b = append(b, typBefunde(fset, x)...)
		}
		return true
	})
	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.GenDecl:
			b = append(b, varBefunde(fset, x)...)
		case *ast.FuncDecl:
			b = append(b, funcBefunde(fset, x)...)
		}
	}
	return b
}

func typBefunde(fset *token.FileSet, x *ast.TypeSpec) []string {
	pos := fset.Position(x.Pos()).String()
	if !erlaubteTypen()[x.Name.Name] {
		return []string{pos + ": Typ " + x.Name.Name + " ist nicht erlaubt"}
	}
	if x.Name.Name != "Verwaltung" {
		return nil
	}
	st, ok := x.Type.(*ast.StructType)
	if !ok || len(st.Fields.List) != 1 || len(st.Fields.List[0].Names) != 1 || st.Fields.List[0].Names[0].Name != "k" {
		return []string{pos + ": Verwaltung hat nicht genau das Feld k"}
	}
	if id, ok := st.Fields.List[0].Type.(*ast.Ident); !ok || id.Name != "Konfig" {
		return []string{pos + ": das Feld von Verwaltung ist keine Konfig"}
	}
	return nil
}

// varBefunde: eine Paketvariable außer `_`, und `_` nur in der Form `var _ T = (*U)(nil)`.
func varBefunde(fset *token.FileSet, g *ast.GenDecl) []string {
	var b []string
	if g.Tok != token.VAR {
		return nil
	}
	for _, s := range g.Specs {
		v := s.(*ast.ValueSpec)
		for _, n := range v.Names {
			if n.Name != "_" {
				b = append(b, fset.Position(n.Pos()).String()+": Paketvariable "+n.Name)
			}
		}
		if len(v.Names) != 1 || v.Names[0].Name != "_" || len(v.Values) != 1 || !zuweisungNil(v.Values[0]) {
			b = append(b, fset.Position(v.Pos()).String()+": Paketvariable `_` ist nicht die Zuweisung (*T)(nil)")
		}
	}
	return b
}

func zuweisungNil(e ast.Expr) bool {
	c, ok := e.(*ast.CallExpr)
	if !ok || len(c.Args) != 1 {
		return false
	}
	p, ok := c.Fun.(*ast.ParenExpr)
	if !ok {
		return false
	}
	if _, ok := p.X.(*ast.StarExpr); !ok {
		return false
	}
	a, ok := c.Args[0].(*ast.Ident)
	return ok && a.Name == "nil"
}

func funcBefunde(fset *token.FileSet, fd *ast.FuncDecl) []string {
	var b []string
	ort := func(p token.Pos) string { return fset.Position(p).String() }
	if fd.Recv == nil && fd.Name.Name == "init" {
		b = append(b, ort(fd.Pos())+": Funktion init")
	}
	var empfaenger string
	if fd.Recv != nil {
		typ := fd.Recv.List[0].Type
		if s, ok := typ.(*ast.StarExpr); ok {
			typ = s.X
		}
		if id, ok := typ.(*ast.Ident); !ok || id.Name != "Verwaltung" {
			b = append(b, ort(fd.Pos())+": Methode an einem anderen Typ als Verwaltung")
		} else if len(fd.Recv.List[0].Names) == 1 {
			empfaenger = fd.Recv.List[0].Names[0].Name
		}
	}
	if fd.Body == nil {
		return b
	}
	b = append(b, schreibBefunde(fset, fd, zustandsNamen(fd))...)
	if empfaenger != "" {
		b = append(b, empfaengerBefunde(fset, fd, empfaenger)...)
	}
	if fd.Recv == nil && fd.Name.Name == "New" {
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt, *ast.IncDecStmt:
				b = append(b, ort(n.Pos())+": Zuweisung in New")
			case *ast.RangeStmt:
				if x.Tok == token.ASSIGN {
					b = append(b, ort(n.Pos())+": Zuweisung durch range in New")
				}
			}
			return true
		})
	}
	return b
}

// schreibBefunde nennt jeden Schreibzugriff (Zuweisung, ++, --, &, range-Ziel), dessen Ausdruck über Selektor, Index,
// Stern, Klammer oder Ausschnitt auf einen Empfänger oder Parameter vom Typ Verwaltung oder Konfig zurückgeht; der
// bloße Name bindet nur die lokale Kopie neu und zählt nicht.
func schreibBefunde(fset *token.FileSet, fd *ast.FuncDecl, namen map[string]bool) []string {
	var b []string
	pruefe := func(e ast.Expr) {
		if e == nil {
			return
		}
		if w, abgeschaelt := wurzel(e); w != nil && abgeschaelt && namen[w.Name] {
			b = append(b, fset.Position(e.Pos()).String()+": Schreibzugriff auf "+w.Name)
		}
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			for _, l := range x.Lhs {
				pruefe(l)
			}
		case *ast.IncDecStmt:
			pruefe(x.X)
		case *ast.UnaryExpr:
			if x.Op == token.AND {
				pruefe(x.X)
			}
		case *ast.RangeStmt:
			pruefe(x.Key)
			pruefe(x.Value)
		}
		return true
	})
	return b
}

// empfaengerBefunde: der Empfänger kommt nur als `v.k.<Feld>` oder `v.<Methode>` vor, keine Kopie von v oder v.k.
func empfaengerBefunde(fset *token.FileSet, fd *ast.FuncDecl, v string) []string {
	operandIdent := map[*ast.Ident]bool{}
	operandSel := map[*ast.SelectorExpr]bool{}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			switch x := s.X.(type) {
			case *ast.Ident:
				operandIdent[x] = true
			case *ast.SelectorExpr:
				operandSel[x] = true
			}
		}
		return true
	})
	var b []string
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Name == v && !operandIdent[x] {
				b = append(b, fset.Position(x.Pos()).String()+": der Empfänger steht außerhalb von v.k.<Feld> und v.<Methode>")
			}
		case *ast.SelectorExpr:
			if id, ok := x.X.(*ast.Ident); ok && id.Name == v && x.Sel.Name == "k" && !operandSel[x] {
				b = append(b, fset.Position(x.Pos()).String()+": v.k wird kopiert oder weitergegeben")
			}
		}
		return true
	})
	return b
}

// TestPaketOhneZustand (Z57, Z47, Z04): der Quelltext des Pakets erfüllt die Regeln aus dateiBefunde. Der Test ist
// deterministisch (keine Goroutine, keine Zeit); die Zahl der Dateien (7) schützt gegen leeres Lesen.
func TestPaketOhneZustand(t *testing.T) {
	befunde, dateien := quelltextBefunde(t, ".")
	if dateien != 7 {
		t.Errorf("%d Dateien ohne _test.go gelesen, erwartet 7", dateien)
	}
	if n := len(erlaubteImporte()); n != 12 {
		t.Errorf("%d erlaubte Importe, erwartet 12", n)
	}
	for _, b := range befunde {
		t.Error(b)
	}
}

// grundlage ist ein Paket, das alle Regeln erfüllt; die Formen setzen eine Abweichung ein.
const grundlage = `package p

type Konfig struct{ Uhr, Log, Workspaces int }

type Verwaltung struct{ k Konfig }

%s

func New(k Konfig) *Verwaltung {
	%s
	return &Verwaltung{k: k}
}

func (v *Verwaltung) Get() int {
	%s
	return v.k.Workspaces
}
`

func quelle(kopf, neu, get string) string { return fmt.Sprintf(grundlage, kopf, neu, get) }

func befundeVon(t *testing.T, name, src string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatalf("%s: %v\n%s", name, err, src)
	}
	return dateiBefunde(fset, f)
}

// TestZustandFormen (Z57): die Regeln fangen die Klasse, nicht die genannten Beispiele. Zwei Schleifen: Schreibform ×
// Wurzel × Ort (8 × 6 × 10 = 480 Quelltexte, jeder mit mindestens einem Befund) und die übrigen Klassen (Typ, Variable,
// Import, Aufruf, Kopie des Empfängers); dazu die Gutfälle ohne Befund. Nicht erfasst (akzeptiertes Negativ): eine Sperre
// über einen Kanal, eine Goroutine aus einer erlaubten Funktion der Standardbibliothek und Zustand in einem Port.
// Grund: ein Quelltext-Test sieht Schreibweisen im Paket, keine Wirkung. Folge: ein solcher Weg bleibt dem Review.
func TestZustandFormen(t *testing.T) {
	formen := []string{"%s = nil", "%s += 1", "%s++", "%s--", "_ = &%s", "%s, _ = a, b", "for %s = range s {\n}", "%s[\"m\"] = 1"}
	wurzeln := []string{"v.k.Uhr", "v.k.Log", "v.k.Workspaces", "v.k", "*v", "v"}
	orte := []string{"Create", "Reset", "Delete", "Get", "List", "baseline", "transaktion", "protokolliere", "FuncLit", "New"}
	n := 0
	for _, form := range formen {
		for _, w := range wurzeln {
			for _, ort := range orte {
				wz := w
				if ort == "New" {
					wz = strings.NewReplacer("v.k", "k", "*v", "*k", "v", "k").Replace(w)
				}
				anweisung := fmt.Sprintf(form, wz)
				var src string
				switch ort {
				case "New":
					src = quelle("", anweisung, "")
				case "FuncLit":
					src = quelle("", "", "func() {\n"+anweisung+"\n}()")
				case "Get":
					src = quelle("", "", anweisung)
				default:
					src = quelle(fmt.Sprintf("func (v *Verwaltung) %s() {\n%s\n}", ort, anweisung), "", "")
				}
				n++
				if len(befundeVon(t, form+"/"+w+"/"+ort, src)) == 0 {
					t.Errorf("kein Befund: %q an %s in %s", form, w, ort)
				}
			}
		}
	}
	if n != 8*6*10 {
		t.Errorf("%d Elemente, erwartet 480", n)
	}
	uebrige := map[string]string{
		"Typ mit Feld":                  quelle("type z struct{ n int }", "", ""),
		"Funktionstyp":                  quelle("type f func()", "", ""),
		"lokaler Typ":                   quelle("", "", "type l struct{ n int }"),
		"Typ mit Methode und Zuweisung": quelle("type z struct{}\n\nfunc (z) Now() int { return 0 }", "k.Uhr = 0", ""),
		"var _ mit Aufruf":              quelle("func f() int { return 1 }\n\nvar _ = f()", "", ""),
		"var _ mit Literal":             quelle("var _ = struct{}{}", "", ""),
		"var x":                         quelle("var x int", "", ""),
		"var in Gruppe":                 quelle("var (\n\ta, b int\n)", "", ""),
		"Import sync":                   "package p\n\nimport \"sync\"\n\nvar _ sync.Mutex\n",
		"Import sync Raw-String":        "package p\n\nimport `sync`\n",
		"Import sync/atomic":            "package p\n\nimport \"sync/atomic\"\n",
		"Import sync Alias":             "package p\n\nimport s \"sync\"\n",
		"Import sync Punkt":             "package p\n\nimport . \"sync\"\n",
		"Import net/http":               "package p\n\nimport \"net/http\"\n",
		"Import os":                     "package p\n\nimport \"os\"\n",
		"Import time":                   "package p\n\nimport \"time\"\n",
		"Import runtime":                "package p\n\nimport \"runtime\"\n",
		"Import sync Unterstrich":       "package p\n\nimport _ \"sync\"\n",
		"time.AfterFunc":                quelle("", "", "time.AfterFunc(0, nil)"),
		"context.AfterFunc":             quelle("", "", "context.AfterFunc(nil, nil)"),
		"go":                            quelle("", "", "go f()"),
		"go im Funktionsliteral":        quelle("", "", "func() { go func() {}() }()"),
		"go mit Methodenwert":           quelle("", "", "go v.Get()"),
		"init":                          quelle("func init() {}", "", ""),
		"Kopie w := v":                  quelle("", "", "w := v\n\t_ = w"),
		"Aufruf f(v)":                   quelle("", "", "f(v)"),
		"Kopie q := v.k":                quelle("", "", "q := v.k\n\t_ = q"),
		"return v.k":                    quelle("func (v *Verwaltung) K() Konfig {\n\treturn v.k\n}", "", ""),
		"return v":                      quelle("func (v *Verwaltung) Selbst() *Verwaltung {\n\treturn v\n}", "", ""),
		"zweites Feld":                  strings.Replace(quelle("", "", ""), "struct{ k Konfig }", "struct {\n\tk Konfig\n\tn int\n}", 1),
		"erstes Feld":                   strings.Replace(quelle("", "", ""), "struct{ k Konfig }", "struct {\n\tn int\n\tk Konfig\n}", 1),
		"Methode an anderem Typ":        quelle("type eintrag struct{}\n\nfunc (e eintrag) m() {}", "", ""),
	}
	for name, src := range uebrige {
		if len(befundeVon(t, name, src)) == 0 {
			t.Errorf("kein Befund: %s", name)
		}
	}
	if len(uebrige) != 32 {
		t.Errorf("%d übrige Formen, erwartet 32", len(uebrige))
	}
	gut := map[string]string{
		"Grundlage":                  quelle("", "", ""),
		"Zuweisung an Schnittstelle": quelle("var _ interface{} = (*Verwaltung)(nil)", "", ""),
		"lokale Struktur":            quelle("", "", "var a struct{ r int }\n\ta.r = 1"),
		"lokale Map":                 quelle("", "", "g := map[string]bool{}\n\tg[\"k\"] = true"),
		"Lesen mit Aufruf":           quelle("", "", "_ = v.k.Uhr.Now()"),
		"Methode des Dienstes":       quelle("", "", "_ = v.baseline(1)"),
		"Typ eintrag":                quelle("type eintrag struct{ n int }", "", ""),
		"Import aus der Liste":       "package p\n\nimport (\n\t\"context\"\n\t\"errors\"\n)\n",
	}
	for name, src := range gut {
		if b := befundeVon(t, name, src); len(b) != 0 {
			t.Errorf("Gutfall %s: %v", name, b)
		}
	}
	if len(gut) != 8 {
		t.Errorf("%d Gutfälle, erwartet 8", len(gut))
	}
}
