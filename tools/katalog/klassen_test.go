package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Die Klassen hinter den Zusagen von SPEC-201 Nr. 4 a, 4 b, 4 c, 5 a, 7 und 8: jeder Test läuft als Schleife über die Menge
// aus dem Wortlaut der Zusage, mit einem Orakel, das nicht aus dem Programm stammt (`regexp`, `go/parser`, die Rechnung des
// Tests selbst).

// literalPfad parst src mit go/parser und liefert die Zeile des Literals mit `NAME_X` und die Namen der Knotenarten
// über ihm; genau ein solches Literal steht in src.
func literalPfad(t *testing.T, src string) (int, map[string]bool) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("Fixture ist kein Go: %v\n%s", err, src)
	}
	var stack []ast.Node
	arten, zeile, n := map[string]bool{}, 0, 0
	ast.Inspect(f, func(x ast.Node) bool {
		if x == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, x)
		if l, ok := x.(*ast.BasicLit); ok && strings.Contains(l.Value, "NAME_X") {
			n++
			zeile = fset.PositionFor(l.Pos(), false).Line
			for _, s := range stack[:len(stack)-1] {
				arten[reflect.TypeOf(s).Elem().Name()] = true
			}
		}
		return true
	})
	if n != 1 {
		t.Fatalf("%d Literale mit NAME_X statt 1:\n%s", n, src)
	}
	return zeile, arten
}

// knotenFaelle sind die Stellen eines Literals nach Knotenart des Parsers: Rumpf und Kopf jeder Anweisung, Klammer,
// Operator, Ausschnitt, Typ mit Längenausdruck, Struct-Tag, Importpfad.
func knotenFaelle() []string {
	return []string{
		`func f() { if g("NAME_X") {} }`, `func f() { if true { g("NAME_X") } }`, `func f() { if true {} else { g("NAME_X") } }`,
		`func f() { if x := "NAME_X"; x != "" {} }`,
		`func f() { for g("NAME_X") {} }`, `func f() { for i := g("NAME_X"); i < 1; i++ {} }`, `func f() { for { g("NAME_X") } }`,
		`func f() { for i := 0; i < 1; i += g("NAME_X") {} }`,
		`func f() { for range g("NAME_X") {} }`, `func f() { for _, c := range []string{"NAME_X"} { _ = c } }`,
		`func f() { for range 1 { g("NAME_X") } }`,
		`func f() { go g("NAME_X") }`, `func f() { defer func() { g("NAME_X") }() }`,
		`func f(c chan string) { select { case c <- "NAME_X": } }`, `func f(c chan string) { select { case <-c: g("NAME_X") } }`,
		`func f() { select { default: g("NAME_X") } }`, `func f() { select { case v := <-g("NAME_X"): _ = v } }`,
		`func f() { switch y := g("NAME_X"); y {} }`, `func f() { switch g("NAME_X") {} }`,
		`func f(x int) { switch { case x > 0: g("NAME_X") } }`,
		`func f(x any) { switch v := x.(type) { case int: g("NAME_X"); _ = v } }`, `func f() { switch v := g("NAME_X").(type) { default: _ = v } }`,
		`func f(x any) { switch y := g("NAME_X"); x.(type) {} }`,
		`func f() { L: g("NAME_X") }`, `func f() { { g("NAME_X") } }`,
		`func f() { m["NAME_X"]++ }`, `func f() { m["NAME_X"] += 1 }`, `func f() { x.f = "NAME_X" }`, `func f() { a, b := "NAME_X", 1; _, _ = a, b }`,
		`func f() string { return "NAME_X" }`,
		`var x = ("NAME_X")`, `var x = &T{"NAME_X"}`, `var x = !g("NAME_X")`, `var x = t[g("NAME_X"):]`, `var x = "NAME_X"[1:]`,
		`var x = t[1:g("NAME_X")]`, `var x = t[1:2:g("NAME_X")]`, `var x = *g("NAME_X")`, `var x = g("NAME_X").f`, `var x = g("NAME_X").(T)`,
		`var x = f[int, [len("NAME_X")]int]`, `var x = func() { g("NAME_X") }`, `var x = a + "NAME_X"`, `var x = map[string]int{"NAME_X": 1}`,
		`type T struct { F string "NAME_X" }`, "type T struct { F string `NAME_X` }", `type T [len("NAME_X")]int`, `var a [len("NAME_X")]int`,
		`func f(a [len("NAME_X")]int) {}`, `type T interface { M(a [len("NAME_X")]int) }`, `var m map[[len("NAME_X")]int]string`,
		`var c chan [len("NAME_X")]int`, `func f(a ...[len("NAME_X")]int) {}`, `type T struct { F [len("NAME_X")]int }`,
		`func f() { var a [len("NAME_X")]int; _ = a }`,
		`import "NAME_X"`, `import ( "a"; "NAME_X" )`, `import n "NAME_X"`, "import `NAME_X`",
	}
}

// TestKatSchlKnotenarten: ein Literal der Form ist an jeder Stelle von `konfig*.go` ein Name (SPEC-201 Nr. 4 a). Menge
// 59 Stellen, aus den 46 Knotenarten des heutigen go/ast gebildet, in denen ein Ausdruck stehen kann (die Liste `verlangt`);
// der Test prüft mit go/parser, dass jede Knotenart über dem Literal mindestens eines Falls liegt. Je Fall ohne den Namen in
// SPEC-110 genau Art 4 an der Zeile des Literals, mit dem Namen `ok`. Prüfgrenze: künftige Syntax von Go fehlt in der Liste.
func TestKatSchlKnotenarten(t *testing.T) {
	faelle := knotenFaelle()
	zaehle(t, "Fälle", len(faelle), 59)
	gesehen := map[string]bool{}
	for i, s := range faelle {
		src := paket + s + "\n"
		zeile, arten := literalPfad(t, src)
		for a := range arten {
			gesehen[a] = true
		}
		name := fmt.Sprintf("Fall %d: %s", i+1, s)
		rot := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_X"))
		rot[dExtra] = src
		erwarteAus(t, name, rot, 1, katZeilen(katS, art4("NAME_X", zeile)))
		gruen := baumSchl(schluesselBasis(), plus(schluesselBasis(), "NAME_X"), plus(schluesselBasis(), "NAME_X"))
		gruen[dExtra] = src
		erwarteOk(t, name+" grün", gruen)
	}
	verlangt := strings.Fields(`AssignStmt BinaryExpr BlockStmt CallExpr CaseClause CommClause CompositeLit DeclStmt DeferStmt ExprStmt
		Field FieldList ForStmt FuncDecl FuncLit FuncType GenDecl GoStmt IfStmt ImportSpec
		IncDecStmt IndexExpr IndexListExpr KeyValueExpr LabeledStmt ParenExpr RangeStmt ReturnStmt SelectStmt SelectorExpr
		SendStmt SliceExpr StarExpr StructType SwitchStmt TypeAssertExpr TypeSpec TypeSwitchStmt UnaryExpr ValueSpec
		ArrayType ChanType MapType InterfaceType Ellipsis File`)
	zaehle(t, "Knotenarten", len(verlangt), 46)
	for _, a := range verlangt {
		if !gesehen[a] {
			t.Errorf("Knotenart %s: kein Fall führt das Literal darin", a)
		}
	}
}

// asciiWerte sind die 128 Zeichen des ASCII-Zeichensatzes als Zeichenketten der Länge 1.
func asciiWerte() []string {
	w := make([]string, 128)
	for i := range w {
		w[i] = string(rune(i))
	}
	return w
}

// zeichenGruppen sind die Stellen eines Zeichens im Wert: allein (Länge 1; dazu der leere Wert), hinter und vor einem
// Zeichen (Länge 2), vorn, in der Mitte und hinten (Länge 3); a und ab sind gültige Teile der Form.
func zeichenGruppen(a, ab string) [][3]string {
	return [][3]string{{"einzeln", "", ""}, {"vorn, Länge 2", "", a}, {"hinten, Länge 2", a, ""}, {"vorn", "", ab}, {"Mitte", a, ab[1:]}, {"hinten", ab, ""}}
}

// TestKatFormZeichenklassen: die Form eines Schlüssels `^[A-Z][A-Z0-9_]+$` und die eines Kommandos `^[a-z]+$` gilt für jedes
// ASCII-Zeichen an jeder Stelle (SPEC-201 Nr. 4 a, 5 a). Menge 1550 = 2 Formen × 6 Stellen × 128 Zeichen + 2 leere Werte + je Form 4 Werte mit Zeichen ab 0x80 und 2 lange Werte (der
// Vertrag nennt keine obere Länge);
// das Orakel ist `regexp`. Je Form und Stelle ein Lauf: gültige Werte sind Art 4 (Schlüssel, nicht in SPEC-110) bzw. Art 7
// (Kommando, ohne Zeile), ungültige geben nichts.
func TestKatFormZeichenklassen(t *testing.T) {
	reS, reK := regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`), regexp.MustCompile(`^[a-z]+$`)
	n := 0
	for _, g := range zeichenGruppen("A", "AB") {
		werte := make([]string, 0, 129)
		if g[0] == "einzeln" {
			werte = append(werte, "")
		}
		for _, c := range asciiWerte() {
			werte = append(werte, g[1]+c+g[2])
		}
		if g[0] == "einzeln" { // Zeichen ab 0x80 (zwei Bytes, Vollbreite, ein einzelnes Byte) und lange Werte: keine obere Länge
			werte = append(werte, "A\xc3\x84", "\xc3\x84B", "AB\x80", "\xef\xbc\xa1B", strings.Repeat("A", 40), strings.Repeat("A", 64))
		}
		var src strings.Builder
		src.WriteString(paket)
		var gueltig []string
		zeilen := map[string]int{}
		for i, v := range werte {
			fmt.Fprintf(&src, "var v%d = %s\n", i, strconv.Quote(v))
			if reS.MatchString(v) {
				gueltig = append(gueltig, v)
				zeilen[v] = 3 + i
			}
		}
		n += len(werte)
		m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), gueltig...))
		m[dExtra] = src.String()
		sort.Strings(gueltig)
		var soll []string
		for _, v := range gueltig {
			soll = append(soll, art4(v, zeilen[v]))
		}
		if soll == nil {
			erwarteOk(t, "Schlüssel "+g[0], m)
			continue
		}
		erwarteAus(t, "Schlüssel "+g[0], m, 1, katZeilen(katS, soll...))
	}
	for _, g := range zeichenGruppen("a", "ab") {
		werte := make([]string, 0, 129)
		if g[0] == "einzeln" {
			werte = append(werte, "")
		}
		for _, c := range asciiWerte() {
			werte = append(werte, g[1]+c+g[2])
		}
		if g[0] == "einzeln" {
			werte = append(werte, "\xc3\xa4", "a\xc3\xa4", "ab\x80", "\xef\xbd\x81b", strings.Repeat("a", 10), strings.Repeat("a", 40))
		}
		m := baumKomm(kommandoBasis(), kommandoBasis())
		var gueltig []string
		for i, v := range werte {
			m[dCli] += fmt.Sprintf("const z%d = %s\n", i, strconv.Quote(v))
			if reK.MatchString(v) {
				gueltig = append(gueltig, v)
			}
		}
		n += len(werte)
		sort.Strings(gueltig)
		var soll []string
		for _, v := range gueltig {
			soll = append(soll, v+" fehlt im Handbuch")
		}
		erwarteAus(t, "Kommando "+g[0], m, 1, katZeilen(katK, soll...))
	}
	zaehle(t, "Werte", n, 1550)
}

// TestKatSchlDateinamen: gelesen wird ein Eintrag, der eine Datei ist, dessen Name mit `konfig` beginnt, auf `.go` und nicht
// auf `_test.go` endet; die Überschrift des Abschnitts beginnt mit `### SPEC-110 — `, und `### ` und `## ` beenden ihn
// am Zeilenanfang, außerhalb von Zäunen (SPEC-201 Nr. 4 a, 4 b). Menge 43 = 8 gelesene und 15 nicht gelesene Dateinamen, 5
// Verweise (Datei, Verzeichnis, ins Leere, ins Leere mit anderem Namen, ins Leere auf `_test.go`), 6 Köder der Überschrift
// (vorn ein Zeichen, eingerückt, vier `#`, mitten in der Zeile, Tabulator, im Zaun) und 9 Köder des Endes (mitten, vor `##`,
// mitten mit Wort, vier `#`, fünf `#`, eingerückt mit `##` und `###`, im Zaun mit `##` und `###`).
func TestKatSchlDateinamen(t *testing.T) {
	text := paket + "var x = \"NAME_F\"\n"
	gelesen := []string{"konfigtest.go", "konfig_Test.go", "konfig_test.go.go", "konfig_.go", "konfig-x.go", "konfigX.go", "konfig.go.go", "konfig_ergo.go"}
	nicht := []string{"konfi.go", "konfig", "onfig.go", "konfig.go.bak", "konfig_ergo", "konfig.gox", "konfig.GO", "konfig.go~", "Konfig.go", "KONFIG.go",
		"xkonfig.go", "konfig_test.go", "konfig_x_test.go", "konfigtest_test.go", "_test.go"}
	n := len(gelesen) + len(nicht)
	for _, d := range gelesen {
		m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_F"))
		m["src/bootstrap/"+d] = text
		erwarteAus(t, d, m, 1, katZeilen(katS, art4("NAME_F", 3)))
	}
	for _, d := range nicht {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		m["src/bootstrap/"+d] = text
		erwarteOk(t, d, m)
	}
	verweise := []struct {
		name, link, ziel string // ziel: "" = ins Leere, "datei", "verz"
		code             int
		out, err         string
	}{
		{"Verweis auf Datei", "konfig_l.go", "datei", 1, "", katZeilen(katS, art4("NAME_F", 3))},
		{"Verweis auf Verzeichnis", "konfig_v.go", "verz", 0, "katalog: ok\n", ""},
		{"Verweis ins Leere", "konfig_l.go", "", 2, "", eingaben("konfig*.go nicht lesbar: konfig_l.go: no such file or directory")},
		{"Verweis ins Leere, anderer Name", "x.go", "", 0, "katalog: ok\n", ""},
		{"Verweis ins Leere, _test.go", "konfig_l_test.go", "", 0, "katalog: ok\n", ""},
	}
	for _, v := range verweise {
		w := plus(schluesselBasis())
		if v.ziel == "datei" {
			w = plus(w, "NAME_F")
		}
		m := baumSchl(schluesselBasis(), schluesselBasis(), w)
		m["src/ziel.txt"] = text
		m["src/verz/x.go"] = text
		wurzel := schreibe(t, m)
		ziel := map[string]string{"": "nirgends", "datei": filepath.Join(wurzel, "src", "ziel.txt"), "verz": filepath.Join(wurzel, "src", "verz")}[v.ziel]
		if err := os.Symlink(ziel, filepath.Join(wurzel, "src", "bootstrap", v.link)); err != nil {
			t.Fatal(err)
		}
		c, o, e := fahre(wurzel)
		gleichFahrt(t, v.name, c, o, e, v.code, v.out, v.err)
		n++
	}
	dash := "\xe2\x80\x94"
	anfang := []string{"x### SPEC-110 " + dash + " x", "  ### SPEC-110 " + dash + " x", "#### SPEC-110 " + dash + " x",
		"Siehe ### SPEC-110 " + dash + " dort", "\t### SPEC-110 " + dash + " x", "```\n### SPEC-110 " + dash + " x\n```"}
	for _, k := range anfang {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		m[dSpez] += "\n" + k + "\n"
		erwarteOk(t, fmt.Sprintf("Überschrift %q", k), m)
		n++
	}
	ende := []string{"Siehe ### Anderes", "x## Anderes", "text ## mitten", "#### Tiefer", "##### Tiefer", "  ## Anderes", "  ### Anderes",
		"```\n## Im Zaun\n```", "```\n### Im Zaun\n```"}
	for _, k := range ende {
		m := baumSchl(schluesselBasis(), schluesselBasis(), schluesselBasis())
		kuerzeErste(m, dSpez, "2. **"+katS+" dieser Version.**", k+"\n2. **"+katS+" dieser Version.**")
		erwarteOk(t, fmt.Sprintf("Ende %q", k), m)
		n++
	}
	zaehle(t, "Fälle", n, 43)
}

// TestKatKommWertarten: nur ein Zeichenketten-Literal als Wert einer Konstante gibt ein Kommando; jede andere Wertart und
// jede Konstante ohne eigenen Wert gibt keines und keinen Absturz (SPEC-201 Nr. 5 a). Menge 58 = 8 Wertformen
// (Zeichenkette, Raw-String, Rune, Ganzzahl in Hexschreibweise, Fließkomma, Imaginärzahl, Umwandlung, Klammer) × 6 Stellen
// (Paketebene einzeln, Gruppe, Funktionskörper, Funktionsliteral, `if`-Block, `case` im Körper) + 10 Konstanten mit weniger Werten als Namen, an jeder Stelle der Gruppe.
func TestKatKommWertarten(t *testing.T) {
	werte := []struct {
		v  string
		da bool
	}{{`"neu"`, true}, {"`neu`", true}, {`'a'`, false}, {`0x1F`, false}, {`1.5`, false}, {`1i`, false}, {`string("neu")`, false}, {`("neu")`, false}}
	stellen := []func(v string) string{
		func(v string) string { return "const a = " + v },
		func(v string) string { return "const (\n\ta = " + v + "\n)" },
		func(v string) string { return "func f() { const a = " + v + " }" },
		func(v string) string { return "var f = func() { const a = " + v + " }" },
		func(v string) string { return "func f() { if true { const a = " + v + " } }" },
		func(v string) string { return "func f(s string) { switch s { case \"x\": const a = " + v + " } }" },
	}
	faelle := []struct {
		src  string
		soll []string
	}{
		{"const (\n\ta = \"neu\"\n\tb\n)", []string{"neu"}}, {"const (\n\ta = \"neu\"\n\tb\n\tc\n)", []string{"neu"}},
		{"const (\n\ta\n\tb = \"neu\"\n)", []string{"neu"}}, {"const (\n\ta\n\tb\n\tc = \"neu\"\n)", []string{"neu"}},
		{`const a, b = "neu"`, []string{"neu"}}, {`const a, b, c = "neu", 1`, []string{"neu"}}, {`const a, b = 1`, nil},
		{`const a, b = 1, "neu"`, []string{"neu"}}, {"const (\n\ta, b = \"neu\", \"eins\"\n\tc, d\n)", []string{"eins", "neu"}},
		{`func f() { const a, b = "neu" }`, []string{"neu"}},
	}
	n := 0
	pruefe := func(src string, soll []string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s: Absturz: %v", src, r)
			}
		}()
		m := grund(codeListe()[:3], nil)
		m[dCli] += "\n" + src + "\n"
		var texte []string
		for _, s := range soll {
			texte = append(texte, s+" fehlt im Handbuch")
		}
		if texte == nil {
			erwarteOk(t, src, m)
			return
		}
		erwarteAus(t, src, m, 1, katZeilen(katK, texte...))
	}
	for _, w := range werte {
		for _, st := range stellen {
			var soll []string
			if w.da {
				soll = []string{"neu"}
			}
			pruefe(st(w.v), soll)
			n++
		}
	}
	for _, f := range faelle {
		pruefe(f.src, f.soll)
		n++
	}
	zaehle(t, "Fälle", n, 58)
}

// ersteFehlerzeile ist die Zeile des ersten Fehlers von go/parser in src (ohne Direktive); src hat mindestens zwei Fehler
// in verschiedenen Zeilen.
func ersteFehlerzeile(t *testing.T, src string) int {
	t.Helper()
	_, err := parser.ParseFile(token.NewFileSet(), "x.go", src, parser.SkipObjectResolution)
	el, ok := err.(scanner.ErrorList)
	if !ok || len(el) < 2 || el[0].Pos.Line == el[len(el)-1].Pos.Line {
		t.Fatalf("Fixture braucht zwei Fehler in zwei Zeilen: %v\n%s", err, src)
	}
	erste := el[0].Pos.Line
	for _, e := range el {
		erste = min(erste, e.Pos.Line)
	}
	return erste
}

// TestKatErsteNennung: jede Zusage mit „erste“ nennt die erste Stelle, nie die letzte (SPEC-201 Nr. 4 a, 4 b, 7, 8). Menge
// 20 = (i) 3 Spans der Zeile der Schlüssel mit zwei `Geprüft werden` (vor dem ersten, zwischen beiden, hinter dem zweiten);
// (ii) 3 Quelltexte mit zwei Fehlern in zwei Zeilen × 2 Quellen (`konfig*.go`, `cli.go`); (iii) 11 Namen mit mehr als einem
// Vorkommen: Schlüssel Art 4 (zwei Zeilen einer Datei), Art 3 mit Art 8, mit Art 10 am Namen und mit Art 10 am Präfix,
// Kommandos Art 3 mit Art 8, Fehlercodes Art 8, Art 9 und Art 3 in fehler.go, §4, den Ausnahmen und der Tabelle des Handbuchs.
func TestKatErsteNennung(t *testing.T) {
	n := 0
	base := schluesselBasis()
	zeile2 := func(vor, zwischen, hinten string) string {
		return "2. **" + katS + " dieser Version.** `" + vor + "` Gepr\xc3\xbcft werden `MODE`, `LISTEN`, `DATA_DIR`, `" + zwischen +
			"` und noch einmal Gepr\xc3\xbcft werden `" + hinten + "`, Rest.\n   Gesetzt (nichtleer) und nicht geliefert sind `ADMIN_*` und `STRICT_COVERAGE`: Text.\n"
	}
	spec := ueberschrift110 + "1. **Leer.** Text.\n" + zeile2("VOR_X", "ZWI_X", "HIN_X") + "3. **Weiter.** Text.\n"
	for i, name := range []string{"VOR_X", "ZWI_X", "HIN_X"} {
		m := baumSchl(plus(base, name), base, plus(base, name))
		m[dSpez] = strings.Replace(m[dSpez], spec110(base), spec, 1)
		if i > 0 { // hinter dem ersten `Geprüft werden` gehört zu K
			erwarteOk(t, name, m)
		} else {
			erwarteAus(t, name, m, 1, katZeilen(katS, art4(name, zeileVon(t, m[dKonfig], name))))
		}
		n++
	}
	zweiFehler := []string{
		"package bootstrap\n\nvar a = )\n\nvar b = )\n",
		"package bootstrap\n\nvar a = \"ab\nvar b = )\n",
		"package bootstrap\n\nfunc f() {\n\tx := )\n}\n\nfunc g() {\n\ty := )\n}\n",
	}
	for _, src := range zweiFehler {
		z := ersteFehlerzeile(t, src)
		m := grund(codeListe()[:3], nil)
		m[dKonfig] = src
		erwarteEingabe(t, "konfig "+src, m, fmt.Sprintf("konfig*.go nicht lesbar: konfig.go: kein Go (Zeile %d)", z))
		m = grund(codeListe()[:3], nil)
		m[dCli] = strings.Replace(src, "package bootstrap", "package cli", 1)
		erwarteEingabe(t, "cli "+src, m, fmt.Sprintf("cli.go nicht lesbar: kein Go (Zeile %d)", z))
		n += 2
	}
	// Schlüssel: derselbe Name in zwei Zeilen einer Datei.
	m := baumSchl(base, base, plus(base, "NAME_X"))
	m[dExtra] = paket + "var a = \"NAME_X\"\n\nvar b = \"NAME_X\"\n"
	erwarteAus(t, "Art 4, zwei Zeilen", m, 1, katZeilen(katS, art4("NAME_X", 3)))
	n++
	// Schlüssel und Kommandos: ein Name in drei Zeilen der Tabelle, getrennt durch andere Zeilen.
	for _, f := range []struct{ name, art string }{{"UNB_X", "der nicht in konfig*.go steht"}, {"STRICT_COVERAGE", "den das Binary ablehnt"}, {"ADMIN_X", "den das Binary ablehnt"}} {
		mm := baumSchl(base, base, []string{f.name, "MODE", "LISTEN", f.name, "DATA_DIR", f.name})
		z := zeileVon(t, mm[dHandbuch], "`"+f.name+"`")
		erwarteAus(t, "Art 3 und "+f.art, mm, 1, katZeilen(katS, fmt.Sprintf("%s doppelt in Handbuch, erstes Vorkommen Zeile %d", f.name, z),
			fmt.Sprintf("Handbuch Zeile %d nennt %s, %s", z, f.name, f.art)))
		n++
	}
	komm := baumKomm(kommandoBasis(), []string{"halt", "start", "version", "halt", "halt"})
	z := zeileVon(t, komm[dHandbuch], "`halt`")
	erwarteAus(t, "Kommando Art 3 und 8", komm, 1, katZeilen(katK, fmt.Sprintf("halt doppelt in Handbuch, erstes Vorkommen Zeile %d", z),
		fmt.Sprintf("Handbuch Zeile %d nennt halt, der nicht in cli.go steht", z)))
	n++
	// Fehlercodes: zwei Nennungen im Text, ein Code ohne Fundstelle und einer aus den Ausnahmen.
	codes := codeListe()[:3]
	nennungen := "## Ende\n\nText ODR-E9999.\n\nNoch ein Satz mit ODR-E9999.\n"
	cm := grund(codes, nil)
	kuerzeErste(cm, dHandbuch, "## Ende\n\nText.\n", nennungen)
	erwarteBefunde(t, "Art 8", cm, fmt.Sprintf("Handbuch Zeile %d nennt ODR-E9999, der nicht in fehler.go steht", zeileVon(t, cm[dHandbuch], "ODR-E9999")))
	nennungen = "## Ende\n\nText " + codes[0] + ".\n\nNoch ein Satz mit " + codes[0] + ".\n"
	cm = grund(codes, codes[:1])
	kuerzeErste(cm, dHandbuch, "## Ende\n\nText.\n", nennungen)
	erwarteBefunde(t, "Art 9", cm, fmt.Sprintf("Handbuch Zeile %d nennt %s, der in den Ausnahmen steht", zeileVon(t, cm[dHandbuch], "Text "+codes[0]), codes[0]))
	n += 2
	cm = grund(codes, nil)
	cm[dFehler] += "\nvar zwei = \"" + codes[0] + "\"\n\nvar drei = \"" + codes[0] + "\"\n"
	erwarteBefunde(t, "Art 3 fehler.go", cm, fmt.Sprintf("%s doppelt in fehler.go, erstes Vorkommen Zeile %d", codes[0], zeileVon(t, cm[dFehler], codes[0])))
	cm = grund(codes, nil)
	kuerzeErste(cm, dSpez, "\nText danach.\n", "| `SPEC-901` | `"+codes[0]+"` | B | A |\n| `SPEC-902` | `"+codes[0]+"` | B | A |\n\nText danach.\n")
	erwarteBefunde(t, "Art 3 §4", cm, fmt.Sprintf("%s doppelt in §4, erstes Vorkommen Zeile %d", codes[0], zeileVon(t, cm[dSpez], codes[0])))
	cm = grund(codes, codes[:2])
	cm[dAusnahmen] = ausnahmenText([]string{codes[0], codes[1], codes[0], codes[1], codes[0]})
	erwarteBefunde(t, "Art 3 Ausnahmen", cm, codes[0]+" doppelt in Ausnahmen, erstes Vorkommen Zeile 1", codes[1]+" doppelt in Ausnahmen, erstes Vorkommen Zeile 2")
	cm = grund(codes, nil)
	kuerzeErste(cm, dHandbuch, zeilenH(codes), zeilenH([]string{codes[0], codes[1], codes[0], codes[2], codes[0]}))
	erwarteBefunde(t, "Art 3 Handbuch", cm, fmt.Sprintf("%s doppelt in Handbuch, erstes Vorkommen Zeile %d", codes[0], zeileVon(t, cm[dHandbuch], codes[0])))
	n += 4
	zaehle(t, "Fälle", n, 20)
}

// TestKatSchlPraefix: ein Name, der einem Namen aus N gleicht oder mit einem Präfix aus N beginnt, steht nicht in H und ist,
// genannt, Art 10 (SPEC-201 Nr. 4 c). Menge 20 = 10 Namen (am Präfix `ADMIN_`: `ADMIN_TOKEN`, `ADMIN_`, `ADMIN__X`; gleich dem
// Namen `STRICT`; nicht abgelehnt: `XADMIN_TOKEN`, `ADMIN`, `ADMINX`, `STRICT_COVERAGE`, `STRICT_`, `MIN_TOKEN`) × 2 Fälle
// (ohne und mit Zeile im Handbuch); die Namen stehen in C und in K. Die Rechnung des Tests nennt Präfix und Namen selbst.
func TestKatSchlPraefix(t *testing.T) {
	konfig := func(extra []string) string {
		return konfigLade(extra) + "\nfunc nichtGelieferte() (praefixe, namen []string) {\n\treturn []string{\"ADMIN_\"}, []string{\"STRICT\"}\n}\n"
	}
	namen := []string{"ADMIN_TOKEN", "ADMIN_", "ADMIN__X", "STRICT", "XADMIN_TOKEN", "ADMIN", "ADMINX", "STRICT_COVERAGE", "STRICT_", "MIN_TOKEN"}
	n := 0
	for _, name := range namen {
		abgelehnt := strings.HasPrefix(name, "ADMIN_") || name == "STRICT"
		c := plus(schluesselBasis(), name)
		ohneZeile := baumSchl(c, c, schluesselBasis())
		ohneZeile[dKonfig] = konfig(c)
		if abgelehnt {
			erwarteOk(t, name+" ohne Zeile", ohneZeile)
		} else {
			erwarteAus(t, name+" ohne Zeile", ohneZeile, 1, katZeilen(katS, name+" fehlt im Handbuch"))
		}
		mitZeile := baumSchl(c, c, c)
		mitZeile[dKonfig] = konfig(c)
		if abgelehnt {
			erwarteAus(t, name+" mit Zeile", mitZeile, 1, katZeilen(katS, fmt.Sprintf("Handbuch Zeile %d nennt %s, den das Binary ablehnt", zeileVon(t, mitZeile[dHandbuch], "`"+name+"`"), name)))
		} else {
			erwarteOk(t, name+" mit Zeile", mitZeile)
		}
		n += 2
	}
	zaehle(t, "Fälle", n, 20)
}

// TestKatLineDirektive: die Zeile eines Befunds und die eines Go-Fehlers ist die der Datei, nicht die einer `//line`-Direktive
// (SPEC-201 Nr. 7). Menge 12 = 4 Formen der Direktive (`//line x.go:100`, `//line :100`, `//line x.go:100:7`,
// `/*line x.go:100*/`) × 3 Stellen (Art 4 in `konfig*.go`, Go-Fehler in `konfig*.go`, Go-Fehler in `cli.go`); der Test prüft
// mit go/parser, dass die Direktive die Zeile ändert.
func TestKatLineDirektive(t *testing.T) {
	formen := []struct{ vor, auf string }{{"//line x.go:100\n", ""}, {"//line :100\n", ""}, {"//line x.go:100:7\n", ""}, {"", "/*line x.go:100*/"}}
	n := 0
	for _, f := range formen {
		art4Src := paket + f.vor + f.auf + "var x = \"NAME_X\"\n"
		zeile, _ := literalPfad(t, art4Src)
		fset := token.NewFileSet()
		pf, err := parser.ParseFile(fset, "x.go", art4Src, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		if fset.Position(pf.Decls[0].Pos()).Line == zeile {
			t.Fatalf("die Direktive %q ändert die Zeile nicht", f.vor+f.auf)
		}
		m := baumSchl(schluesselBasis(), schluesselBasis(), plus(schluesselBasis(), "NAME_X"))
		m[dExtra] = art4Src
		erwarteAus(t, "Art 4 "+f.vor+f.auf, m, 1, katZeilen(katS, art4("NAME_X", zeile)))
		fehler := paket + f.vor + f.auf + "var x = )\n"
		dateiZeile := zeileVon(t, fehler, "var x = )")
		_, perr := parser.ParseFile(token.NewFileSet(), "x.go", fehler, parser.SkipObjectResolution)
		el, ok := perr.(scanner.ErrorList)
		if !ok || len(el) == 0 || el[0].Pos.Line == dateiZeile {
			t.Fatalf("die Direktive %q ändert die Zeile des Fehlers nicht: %v", f.vor+f.auf, perr)
		}
		m = grund(codeListe()[:3], nil)
		m[dKonfig] = fehler
		erwarteEingabe(t, "konfig-Fehler "+f.vor+f.auf, m, fmt.Sprintf("konfig*.go nicht lesbar: konfig.go: kein Go (Zeile %d)", dateiZeile))
		m = grund(codeListe()[:3], nil)
		m[dCli] = strings.Replace(fehler, "package bootstrap", "package cli", 1)
		erwarteEingabe(t, "cli-Fehler "+f.vor+f.auf, m, fmt.Sprintf("cli.go nicht lesbar: kein Go (Zeile %d)", dateiZeile))
		n += 3
	}
	zaehle(t, "Fälle", n, 12)
}

// TestKatVerweisEingabe: ein Verweis ohne Ziel ist ein Eintrag, der nicht zu lesen ist; sein Eingabefehler steht in der
// Bytefolge der Namen unter den Eingabefehlern der Dateien, und `kein Schlüssel gefunden` steht nur, wenn keine Datei einen
// Eingabefehler trägt (SPEC-201 Nr. 4 a, 8). Menge 6: der Baum hat nur die genannten Dateien `konfig*.go`; der Verweis
// `konfig_l.go` ist die einzige Datei, steht hinter und vor einer Datei ohne Go (`konfig_a.go`, `konfig_z.go`), neben einer
// gültigen Datei, neben einer Datei ohne Namen und neben einer Datei mit Namen nur aus N.
func TestKatVerweisEingabe(t *testing.T) {
	ohneGo, gueltig, nurN := "package bootstrap\n\nfunc f() {\n", konfigLade(schluesselBasis()), paket+konfigN
	leer := "konfig*.go nicht lesbar: konfig_l.go: no such file or directory"
	kg := func(d string) string { return "konfig*.go nicht lesbar: " + d + ": kein Go (Zeile 3)" }
	faelle := []struct {
		name   string
		dazu   map[string]string
		fehler []string
	}{
		{"einzige Datei", nil, []string{leer}},
		{"hinter einer Datei ohne Go", map[string]string{"konfig_a.go": ohneGo}, []string{kg("konfig_a.go"), leer}},
		{"vor einer Datei ohne Go", map[string]string{"konfig_z.go": ohneGo}, []string{leer, kg("konfig_z.go")}},
		{"neben einer gültigen Datei", map[string]string{"konfig_a.go": gueltig}, []string{leer}},
		{"neben einer Datei ohne Namen", map[string]string{"konfig_a.go": paket + "var x = 1\n"}, []string{leer}},
		{"neben einer Datei mit Namen nur aus N", map[string]string{"konfig_a.go": nurN}, []string{leer}},
	}
	zaehle(t, "Fälle", len(faelle), 6)
	for _, f := range faelle {
		m := grund(codeListe()[:3], nil)
		delete(m, dKonfig)
		m["src/bootstrap/.leer"] = ""
		for d, inhalt := range f.dazu {
			m["src/bootstrap/"+d] = inhalt
		}
		w := schreibe(t, m)
		if err := os.Symlink("nirgends", filepath.Join(w, "src", "bootstrap", "konfig_l.go")); err != nil {
			t.Fatal(err)
		}
		c, o, e := fahre(w)
		gleichFahrt(t, f.name, c, o, e, 2, "", eingaben(f.fehler...))
	}
}
