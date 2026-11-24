package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// importe liefert zu jedem Namen, unter dem eine Datei einen Import nennt, dessen Pfad; ein Punkt-Import steht unter ".".
func importe(f *ast.File) map[string]string {
	m := map[string]string{}
	for _, s := range f.Imports {
		p, _ := strconv.Unquote(s.Path.Value)
		name := path.Base(p)
		if s.Name != nil {
			name = s.Name.Name
		}
		if p == "math/rand/v2" && s.Name == nil {
			name = "rand"
		}
		m[name] = p
	}
	return m
}

// stellen liefert zu jedem Selektor `<Paket>.<Name>`, dessen Paket eines der Pfade nennt und dessen Name pruefe verwirft, den Namen mit
// Pfad und Zeile, zu jedem Punkt-Import eines dieser Pfade den Pfad.
func stellen(name, quelle string, pfade []string, verwirf func(pfad, sel string) bool) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, quelle, 0)
	if err != nil {
		return nil, err
	}
	im := importe(f)
	var out []string
	for n, p := range im {
		if n == "." && slices.Contains(pfade, p) {
			out = append(out, name+": Punkt-Import "+p)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := s.X.(*ast.Ident); ok && im[x.Name] != "" && slices.Contains(pfade, im[x.Name]) && verwirf(im[x.Name], s.Sel.Name) {
				out = append(out, name+":"+strconv.Itoa(fset.Position(s.Pos()).Line)+": "+im[x.Name]+"."+s.Sel.Name)
			}
		}
		return true
	})
	return out, nil
}

// zeitStellen nennt die Stellen, an denen eine Datei eine der sieben Funktionen der Zeit ruft oder nennt (SPEC-216 Nr. 14 a).
func zeitStellen(name, quelle string) ([]string, error) {
	return stellen(name, quelle, []string{"time", "context"}, func(pfad, sel string) bool {
		return slices.Contains(map[string][]string{
			"time":    {"Now", "Sleep", "After", "NewTimer", "NewTicker", "AfterFunc"},
			"context": {"WithTimeout"},
		}[pfad], sel)
	})
}

// zufallStellen nennt die Stellen, an denen eine Datei `math/rand` importiert oder den globalen Generator von `math/rand/v2` ruft; erlaubt
// sind allein die Konstruktoren und die Typen eines eigenen Generators.
func zufallStellen(name, quelle string) ([]string, error) {
	st, err := stellen(name, quelle, []string{"math/rand/v2"}, func(_, sel string) bool {
		return !slices.Contains([]string{"New", "NewPCG", "Rand", "PCG"}, sel)
	})
	if err != nil {
		return nil, err
	}
	f, err := parser.ParseFile(token.NewFileSet(), name, quelle, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	for _, p := range importe(f) {
		if p == "math/rand" {
			st = append(st, name+": Import math/rand")
		}
	}
	return st, nil
}

// quellen liest alle Dateien des Pakets, die Tests eingeschlossen.
func quellen(t *testing.T) map[string]string {
	t.Helper()
	dateien, err := filepath.Glob("*.go")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("Dateien des Pakets: %v, %v", dateien, err)
	}
	m := map[string]string{}
	for _, d := range dateien {
		b, err := os.ReadFile(d)
		if err != nil {
			t.Fatal(err)
		}
		m[d] = string(b)
	}
	return m
}

// TestPaketOhneZeit: keine Datei des Pakets außer uhr.go nennt eine der sieben Funktionen der Zeit (time.Now, time.Sleep, time.After,
// time.NewTimer, time.NewTicker, time.AfterFunc, context.WithTimeout), auch nicht unter einem anderen Namen des Imports und nicht als
// Punkt-Import; die Datei der Uhr trägt sie (Kontrolle), und der Fänger erkennt jede der sieben in jeder der drei Formen.
func TestPaketOhneZeit(t *testing.T) {
	wache(t, 30*time.Second)
	namen := []struct{ paket, name string }{
		{"time", "Now"}, {"time", "Sleep"}, {"time", "After"}, {"time", "NewTimer"}, {"time", "NewTicker"}, {"time", "AfterFunc"},
		{"context", "WithTimeout"},
	}
	formen := []struct{ name, importZeile, aufruf string }{
		{"direkt", `"%s"`, "%s.%s"}, {"mit Namen", `x "%s"`, "x.%s"}, {"Punkt-Import", `. "%s"`, "%s"},
	}
	n := 0
	for _, z := range namen {
		for _, f := range formen {
			n++
			imp := strings.Replace(f.importZeile, "%s", z.paket, 1)
			var ruf string
			switch f.name {
			case "direkt":
				ruf = z.paket + "." + z.name
			case "mit Namen":
				ruf = "x." + z.name
			default:
				ruf = z.name
			}
			quelle := "package p\nimport " + imp + "\nvar _ = " + ruf + "\n"
			st, err := zeitStellen("p.go", quelle)
			if err != nil || len(st) == 0 {
				t.Errorf("%s.%s (%s): Stellen %v, %v, erwartet mindestens eine", z.paket, z.name, f.name, st, err)
			}
		}
	}
	if n != 21 {
		t.Errorf("%d Fälle, erwartet 7 Namen mal 3 Formen", n)
	}
	for _, ok := range []string{"var _ time.Duration = time.Second", "var _ = time.Until(time.Time{})", "var _ = context.Background"} {
		if st, err := zeitStellen("p.go", "package p\nimport (\"time\"; \"context\")\n"+ok+"\n"); err != nil || len(st) != 0 {
			t.Errorf("%s: Stellen %v, %v, erwartet keine", ok, st, err)
		}
	}
	files := quellen(t)
	if _, ok := files["uhr.go"]; !ok {
		t.Fatal("die Datei der Uhr uhr.go fehlt")
	}
	for name, quelle := range files {
		st, err := zeitStellen(name, quelle)
		if err != nil {
			t.Fatal(err)
		}
		if name == "uhr.go" {
			if len(st) < 3 {
				t.Errorf("die Datei der Uhr trägt %v, erwartet time.Now, time.NewTimer und time.NewTicker", st)
			}
		} else if len(st) > 0 {
			t.Errorf("außerhalb der Uhr: %v", st)
		}
	}
	if _, ok := files["paket_test.go"]; !ok || len(files) < 10 {
		t.Errorf("die Prüfung liest %d Dateien, darunter Tests: %v", len(files), ok)
	}
}

// TestPaketOhneGlobalenGenerator: keine Datei des Pakets importiert `math/rand` oder ruft den globalen Generator von `math/rand/v2` (SPEC-216
// Nr. 6 c); der Fänger erkennt jede der 19 Funktionen des globalen Generators, auch unter einem anderen Namen des Imports, und lässt die
// Konstruktoren und Typen eines eigenen Generators zu.
func TestPaketOhneGlobalenGenerator(t *testing.T) {
	wache(t, 30*time.Second)
	global := []string{"Float32", "Float64", "ExpFloat64", "Int", "Int32", "Int32N", "Int64", "Int64N", "IntN", "N", "NormFloat64", "Perm",
		"Shuffle", "Uint", "Uint32", "Uint32N", "Uint64", "Uint64N", "UintN"}
	if len(global) != 19 {
		t.Fatalf("%d Namen, erwartet 19", len(global))
	}
	for _, g := range global {
		for _, kopf := range []string{`import "math/rand/v2"`, `import r "math/rand/v2"`} {
			ruf := "rand." + g
			if strings.Contains(kopf, " r ") {
				ruf = "r." + g
			}
			if st, err := zufallStellen("p.go", "package p\n"+kopf+"\nvar _ = "+ruf+"\n"); err != nil || len(st) != 1 {
				t.Errorf("%s %s: Stellen %v, %v, erwartet eine", kopf, ruf, st, err)
			}
		}
	}
	for _, ok := range []string{"var _ = rand.New(rand.NewPCG(1, 2))", "var _ *rand.Rand", "var _ *rand.PCG"} {
		if st, err := zufallStellen("p.go", "package p\nimport \"math/rand/v2\"\n"+ok+"\n"); err != nil || len(st) != 0 {
			t.Errorf("%s: Stellen %v, %v, erwartet keine", ok, st, err)
		}
	}
	if st, err := zufallStellen("p.go", "package p\nimport \"math/rand\"\n"); err != nil || len(st) != 1 {
		t.Errorf("Import math/rand: Stellen %v, %v, erwartet eine", st, err)
	}
	for name, quelle := range quellen(t) {
		if st, err := zufallStellen(name, quelle); err != nil || len(st) != 0 {
			t.Errorf("%s: %v, %v", name, st, err)
		}
	}
}
