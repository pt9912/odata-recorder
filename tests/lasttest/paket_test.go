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

// importe liefert zu jedem Namen, unter dem eine Datei einen Import nennt, dessen Pfad, und die Pfade ihrer Punkt-Importe.
func importe(f *ast.File) (namen map[string]string, punkt []string) {
	namen = map[string]string{}
	for _, s := range f.Imports {
		p, _ := strconv.Unquote(s.Path.Value)
		name := path.Base(p)
		switch {
		case s.Name != nil && s.Name.Name == ".":
			punkt = append(punkt, p)
			continue
		case s.Name != nil:
			name = s.Name.Name
		case p == "math/rand/v2":
			name = "rand"
		}
		namen[name] = p
	}
	return namen, punkt
}

// stellen liefert zu jedem Selektor `<Paket>.<Name>`, dessen Paket eines der Pfade nennt und dessen Name verwirf verwirft, den Namen mit
// Pfad und Zeile, und zu jedem Punkt-Import eines dieser Pfade den Pfad.
func stellen(name, quelle string, pfade []string, verwirf func(pfad, sel string) bool) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, quelle, 0)
	if err != nil {
		return nil, err
	}
	im, punkt := importe(f)
	var out []string
	for _, p := range punkt {
		if slices.Contains(pfade, p) {
			out = append(out, name+": Punkt-Import "+p)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := s.X.(*ast.Ident); ok && slices.Contains(pfade, im[x.Name]) && verwirf(im[x.Name], s.Sel.Name) {
				out = append(out, name+":"+strconv.Itoa(fset.Position(s.Pos()).Line)+": "+im[x.Name]+"."+s.Sel.Name)
			}
		}
		return true
	})
	return out, nil
}

// zeitKlasse ist die Menge der Namen der Standardbibliothek, die ein Paket ohne Zugriff auf die Uhr erreicht und die die Zeit des Prozesses
// lesen oder auf sie warten: Funktionen je Paket, Felder je Typ (in einem Literal des Typs) und Methoden jedes Empfängers.
func zeitKlasse() (funktionen, felder map[string][]string, methoden []string) {
	funktionen = map[string][]string{
		"time":     {"Now", "Since", "Until", "Sleep", "After", "AfterFunc", "NewTimer", "NewTicker", "Tick"},
		"context":  {"WithTimeout", "WithDeadline", "WithTimeoutCause", "WithDeadlineCause"},
		"net/http": {"TimeoutHandler"},
	}
	felder = map[string][]string{
		"net/http.Client":    {"Timeout"},
		"net/http.Server":    {"ReadTimeout", "WriteTimeout", "IdleTimeout", "ReadHeaderTimeout"},
		"net/http.Transport": {"ResponseHeaderTimeout", "TLSHandshakeTimeout", "IdleConnTimeout", "ExpectContinueTimeout"},
		"net.Dialer":         {"Timeout", "Deadline"},
	}
	return funktionen, felder, []string{"SetDeadline", "SetReadDeadline", "SetWriteDeadline"}
}

// zeitStellen nennt die Stellen, an denen eine Datei ein Element von zeitKlasse nennt: eine Funktion (auch als Wert, unter einem anderen Namen
// des Imports und als Punkt-Import), ein Feld im Literal seines Typs oder eine Methode SetDeadline, SetReadDeadline, SetWriteDeadline.
func zeitStellen(name, quelle string) ([]string, error) {
	funktionen, felder, methoden := zeitKlasse()
	pfade := []string{"net"}
	for p := range funktionen {
		pfade = append(pfade, p)
	}
	out, err := stellen(name, quelle, pfade, func(pfad, sel string) bool { return slices.Contains(funktionen[pfad], sel) })
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, quelle, 0)
	if err != nil {
		return nil, err
	}
	im, _ := importe(f)
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			typ, ok := x.Type.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			paket, ok := typ.X.(*ast.Ident)
			if !ok {
				return true
			}
			for _, e := range x.Elts {
				if kv, ok := e.(*ast.KeyValueExpr); ok {
					if k, ok := kv.Key.(*ast.Ident); ok && slices.Contains(felder[im[paket.Name]+"."+typ.Sel.Name], k.Name) {
						out = append(out, name+":"+strconv.Itoa(fset.Position(k.Pos()).Line)+": "+typ.Sel.Name+"{"+k.Name+"}")
					}
				}
			}
		case *ast.SelectorExpr:
			if slices.Contains(methoden, x.Sel.Name) {
				out = append(out, name+":"+strconv.Itoa(fset.Position(x.Pos()).Line)+": "+x.Sel.Name)
			}
		}
		return true
	})
	return out, nil
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
	im, alle := importe(f)
	for _, p := range im {
		alle = append(alle, p)
	}
	if slices.Contains(alle, "math/rand") {
		st = append(st, name+": Import math/rand")
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

// zeitBefunde nennt von den Stellen einer Datei (zeitStellen) die, die SPEC-216 Nr. 14 a nicht zulässt: in uhr.go ist jede zulässig, in
// uhr_test.go allein time.NewTimer (nach dem aufgelösten Paket und Namen, nicht nach der Schreibweise des Imports; ein Punkt-Import ist kein
// time.NewTimer), in jeder anderen Datei keine.
func zeitBefunde(name string, stellen []string) []string {
	var out []string
	for _, s := range stellen {
		switch {
		case name == "uhr.go":
		case name == "uhr_test.go" && strings.HasSuffix(s, ": time.NewTimer"):
		default:
			out = append(out, s)
		}
	}
	return out
}

// TestPaketOhneZeit: keine Datei des Pakets außer uhr.go nennt ein Element der Klasse der Namen, die die Zeit des Prozesses lesen oder auf sie
// warten, und uhr_test.go allein time.NewTimer: 14 Funktionen (9 von time, 4 von context, TimeoutHandler von net/http), auch als Wert, unter
// einem anderen Namen des Imports und als Punkt-Import, 11 Felder von Literalen (Client, Server, Transport, Dialer) und 3 Methoden SetDeadline;
// die Datei der Uhr trägt Funktionen der Klasse (Kontrolle), und der Fänger erkennt jedes Element der Klasse in jeder Form (Orakel: die Literale
// dieses Tests) und wertet es je Datei: uhr.go lässt alle zu, uhr_test.go allein time.NewTimer, jede andere Datei keines.
func TestPaketOhneZeit(t *testing.T) {
	wache(t, 30*time.Second)
	funktionen := []struct{ paket, name string }{
		{"time", "Now"}, {"time", "Since"}, {"time", "Until"}, {"time", "Sleep"}, {"time", "After"}, {"time", "AfterFunc"},
		{"time", "NewTimer"}, {"time", "NewTicker"}, {"time", "Tick"},
		{"context", "WithTimeout"}, {"context", "WithDeadline"}, {"context", "WithTimeoutCause"}, {"context", "WithDeadlineCause"},
		{"net/http", "TimeoutHandler"},
	}
	felder := []struct{ paket, typ, feld string }{
		{"net/http", "Client", "Timeout"},
		{"net/http", "Server", "ReadTimeout"}, {"net/http", "Server", "WriteTimeout"}, {"net/http", "Server", "IdleTimeout"},
		{"net/http", "Server", "ReadHeaderTimeout"},
		{"net/http", "Transport", "ResponseHeaderTimeout"}, {"net/http", "Transport", "TLSHandshakeTimeout"},
		{"net/http", "Transport", "IdleConnTimeout"}, {"net/http", "Transport", "ExpectContinueTimeout"},
		{"net", "Dialer", "Timeout"}, {"net", "Dialer", "Deadline"},
	}
	gesehen := map[string]string{}
	n := 0
	// pruefe führt ein Element in einer Form des Quelltexts: der Fänger nennt mindestens eine Stelle; uhr.go lässt jede zu, jede andere Datei
	// keine, uhr_test.go keine außer einem time.NewTimer in der Schreibweise eines Selektors (erlaubtInTest).
	pruefe := func(was, kopf, aufruf string, erlaubtInTest bool) {
		n++
		st, err := zeitStellen("p.go", "package p\n"+kopf+"\nvar _ = "+aufruf+"\n")
		if err != nil || len(st) == 0 {
			t.Errorf("%s: Stellen %v, %v, erwartet mindestens eine", was, st, err)
		}
		erwartet := map[string]int{"uhr.go": 0, "uhr_test.go": len(st), "profil.go": len(st), "aufbau_test.go": len(st)}
		if erlaubtInTest {
			erwartet["uhr_test.go"] = 0
		}
		for datei, nBefunde := range erwartet {
			if b := zeitBefunde(datei, st); len(b) != nBefunde {
				t.Errorf("%s in %s: Befunde %v, erwartet %d", was, datei, b, nBefunde)
			}
		}
		gesehen[was] = aufruf
	}
	for _, f := range funktionen {
		neuerTimer := f.paket == "time" && f.name == "NewTimer"
		pruefe(f.paket+"."+f.name+" direkt", `import "`+f.paket+`"`, path.Base(f.paket)+"."+f.name, neuerTimer)
		pruefe(f.paket+"."+f.name+" mit Namen", `import x "`+f.paket+`"`, "x."+f.name, neuerTimer)
		pruefe(f.paket+"."+f.name+" Punkt-Import", `import . "`+f.paket+`"`, f.name, false)
	}
	for _, f := range felder {
		pruefe(f.typ+"{"+f.feld+"} direkt", `import "`+f.paket+`"`, path.Base(f.paket)+"."+f.typ+"{"+f.feld+": 1}", false)
		pruefe(f.typ+"{"+f.feld+"} mit Namen", `import x "`+f.paket+`"`, "x."+f.typ+"{"+f.feld+": 1}", false)
	}
	for _, m := range []string{"SetDeadline", "SetReadDeadline", "SetWriteDeadline"} {
		pruefe(m, `import "net"`, "c."+m, false)
	}
	if n != 14*3+11*2+3 || len(gesehen) != n {
		t.Errorf("%d Fälle, erwartet 14 mal 3, 11 mal 2 und 3", n)
	}
	for _, ok := range []string{
		"var _ time.Duration = time.Second", "var _ = time.Time{}", "var _ = context.Background", "var _ = http.Client{Transport: nil}",
		"var _ = http.Server{Handler: nil}", "var _ = net.Dialer{KeepAlive: 0}", "var _ = c.SetDeadlineX",
	} {
		quelle := "package p\nimport (\"time\"; \"context\"; \"net\"; \"net/http\")\n" + ok + "\n"
		if st, err := zeitStellen("p.go", quelle); err != nil || len(st) != 0 {
			t.Errorf("%s: Stellen %v, %v, erwartet keine", ok, st, err)
		}
	}
	for _, kopf := range []string{`import (. "errors"; . "time")`, `import (. "time"; . "errors")`} {
		if st, err := zeitStellen("p.go", "package p\n"+kopf+"\nvar _ = Now\n"); err != nil || len(st) != 1 {
			t.Errorf("%s: Stellen %v, %v, erwartet eine", kopf, st, err)
		}
	}
	gemischt := "package p\nimport \"time\"\nvar _ = time.NewTimer\nvar _ = time.Sleep\nvar _ = time.NewTimer\n"
	if st, err := zeitStellen("uhr_test.go", gemischt); err != nil || len(zeitBefunde("uhr_test.go", st)) != 1 || len(zeitBefunde("uhr.go", st)) != 0 || len(zeitBefunde("profil.go", st)) != 3 {
		t.Errorf("zwei time.NewTimer und ein time.Sleep: Stellen %v, %v, erwartet je Datei 1, 0 und 3 Befunde", st, err)
	}
	files := quellen(t)
	for _, d := range []string{"uhr.go", "uhr_test.go", "paket_test.go"} {
		if _, ok := files[d]; !ok {
			t.Fatalf("die Datei %s fehlt", d)
		}
	}
	for name, quelle := range files {
		st, err := zeitStellen(name, quelle)
		if err != nil {
			t.Fatal(err)
		}
		if name == "uhr.go" && len(st) < 4 {
			t.Errorf("die Datei der Uhr trägt %v, erwartet time.Now, time.Until, time.NewTimer und time.NewTicker", st)
		}
		if b := zeitBefunde(name, st); len(b) > 0 {
			t.Errorf("außerhalb der Uhr: %v", b)
		}
	}
	if len(files) < 10 {
		t.Errorf("die Prüfung liest %d Dateien", len(files))
	}
}

// TestPaketOhneGlobalenGenerator: keine Datei des Pakets importiert `math/rand` oder ruft den globalen Generator von `math/rand/v2` (SPEC-216
// Nr. 6 c); der Fänger erkennt jede der 19 Funktionen des globalen Generators direkt, unter einem anderen Namen des Imports und als
// Punkt-Import (auch hinter einem anderen Punkt-Import) und lässt die Konstruktoren und Typen eines eigenen Generators zu.
func TestPaketOhneGlobalenGenerator(t *testing.T) {
	wache(t, 30*time.Second)
	global := []string{"Float32", "Float64", "ExpFloat64", "Int", "Int32", "Int32N", "Int64", "Int64N", "IntN", "N", "NormFloat64", "Perm",
		"Shuffle", "Uint", "Uint32", "Uint32N", "Uint64", "Uint64N", "UintN"}
	if len(global) != 19 {
		t.Fatalf("%d Namen, erwartet 19", len(global))
	}
	for _, g := range global {
		for kopf, ruf := range map[string]string{
			`import "math/rand/v2"`: "rand." + g, `import r "math/rand/v2"`: "r." + g, `import . "math/rand/v2"`: g,
			`import (. "errors"; . "math/rand/v2")`: g,
		} {
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
	for _, kopf := range []string{`import "math/rand"`, `import . "math/rand"`} {
		if st, err := zufallStellen("p.go", "package p\n"+kopf+"\n"); err != nil || len(st) != 1 {
			t.Errorf("%s: Stellen %v, %v, erwartet eine", kopf, st, err)
		}
	}
	for name, quelle := range quellen(t) {
		if st, err := zufallStellen(name, quelle); err != nil || len(st) != 0 {
			t.Errorf("%s: %v, %v", name, st, err)
		}
	}
}
