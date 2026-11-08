package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// istSchluessel sagt, ob v ganz `^[A-Z][A-Z0-9_]+$` trifft: ASCII, mindestens zwei Zeichen (SPEC-201 Nr. 4 a).
func istSchluessel(v string) bool {
	if len(v) < 2 || v[0] < 'A' || v[0] > 'Z' {
		return false
	}
	for i := 1; i < len(v); i++ {
		c := v[i]
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}

// liesKSchluessel liest die Schlüssel dieser Version aus SPEC-110 Nr. 2: die Spans hinter dem ersten `Geprüft werden`
// der Zeile `2. **Schlüssel dieser Version.**` (SPEC-201 Nr. 4 b).
func liesKSchluessel(text string, b *bericht) menge {
	m := neuesMd(text)
	var abschnitte []int
	for i, l := range m.zeilen {
		if !m.zaun[i] && strings.HasPrefix(l, ueberschriftSchl) {
			abschnitte = append(abschnitte, i)
		}
	}
	if len(abschnitte) != 1 {
		meldeSchl(len(abschnitte), b)
		return nil
	}
	var zeilen []string
	for j := abschnitte[0] + 1; j < len(m.zeilen); j++ {
		if m.zaun[j] {
			continue
		}
		if l := m.zeilen[j]; strings.HasPrefix(l, "### ") || strings.HasPrefix(l, "## ") {
			break
		} else if strings.HasPrefix(l, zeileSchl) {
			zeilen = append(zeilen, l)
		}
	}
	if len(zeilen) != 1 {
		meldeSchl(len(zeilen), b)
		return nil
	}
	k := menge{}
	if _, hinten, da := strings.Cut(zeilen[0], hinterSchl); da {
		for _, s := range spans(hinten) {
			if istSchluessel(s) {
				k[s] = 0
			}
		}
	}
	if len(k) == 0 {
		meldeSchl(0, b)
		return nil
	}
	return k
}

// meldeSchl meldet K der Schlüssel als fehlend (n gleich 0) oder mehrdeutig.
func meldeSchl(n int, b *bericht) {
	if n == 0 {
		b.fehler("SPEC-110: Schlüssel fehlen")
		return
	}
	b.fehler("SPEC-110: Schlüssel mehrdeutig")
}

// spans liefert die Inline-Code-Spans einer Zeile: Backtick, Text ohne Backtick, Backtick.
func spans(zeile string) []string {
	var out []string
	for {
		i := strings.IndexByte(zeile, '`')
		if i < 0 {
			return out
		}
		j := strings.IndexByte(zeile[i+1:], '`')
		if j < 0 {
			return out
		}
		out = append(out, zeile[i+1:i+1+j])
		zeile = zeile[i+j+2:]
	}
}

// konfigDateien nennt die Dateien `konfig*.go` im Verzeichnis, in der Bytefolge ihrer Namen: ein Eintrag, der eine Datei
// ist (ein Verweis gilt als sein Ziel), dessen Name mit `konfig` beginnt, auf `.go` und nicht auf `_test.go` endet
// (SPEC-201 Nr. 4 a).
func konfigDateien(dir string, b *bericht) ([]string, bool) {
	eintraege, err := os.ReadDir(dir)
	if err != nil {
		meldeLesefehler(err, "konfig*.go", pfadKonfig, b)
		return nil, false
	}
	var out []string
	for _, e := range eintraege {
		n := e.Name()
		if !strings.HasPrefix(n, "konfig") || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		// Ein Verweis ohne Ziel bleibt in der Liste: sein Eingabefehler entsteht beim Lesen, in der Bytefolge der Namen.
		if info, err := os.Stat(filepath.Join(dir, n)); err != nil || !info.IsDir() {
			out = append(out, n)
		}
	}
	return out, true
}

// parseGo liest Go-Quelltext ohne Typprüfung; was der Parser ablehnt, ist die Zeile seines ersten Fehlers in der Datei
// (die Zeile der Datei, nicht die einer `//line`-Direktive, SPEC-201 Nr. 7).
func parseGo(name, src string) (*token.FileSet, *ast.File, int) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	if err == nil {
		return fset, f, 0
	}
	zeile := 1
	var el scanner.ErrorList
	if errors.As(err, &el) && len(el) > 0 {
		erster := el[0].Pos.Offset
		for _, e := range el {
			erster = min(erster, e.Pos.Offset)
		}
		fset.Iterate(func(tf *token.File) bool {
			zeile = tf.PositionFor(tf.Pos(min(erster, tf.Size())), false).Line
			return false
		})
	}
	return nil, nil, zeile
}

// schluessel sind die Namen aus `konfig*.go`: C außerhalb von `nichtGelieferte`, N darin.
type schluessel struct {
	c menge
	n map[string]bool
}

// liesKonfig liest die Literale der Dateien `konfig*.go` (SPEC-201 Nr. 4 a).
func liesKonfig(wurzel string, b *bericht) schluessel {
	s := schluessel{c: menge{}, n: map[string]bool{}}
	dir := filepath.Join(wurzel, filepath.FromSlash(pfadKonfig))
	dateien, ok := konfigDateien(dir, b)
	if !ok {
		return s
	}
	vorher := len(b.eingabe)
	for _, n := range dateien {
		roh, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			b.fehler("konfig*.go nicht lesbar: " + n + ": " + grundVon(err))
			continue
		}
		fset, f, zeile := parseGo(n, string(roh))
		if f == nil {
			b.fehler("konfig*.go nicht lesbar: " + n + ": kein Go (Zeile " + strconv.Itoa(zeile) + ")")
			continue
		}
		aussen, innen := literale(fset, f)
		for name, z := range aussen {
			if _, da := s.c[name]; !da {
				s.c[name] = z
			}
		}
		for name := range innen {
			s.n[name] = true
		}
	}
	if len(b.eingabe) == vorher && len(s.c) == 0 {
		b.fehler("konfig*.go: kein Schlüssel gefunden")
	}
	return s
}

// literale liefert die Namen der Zeichenketten-Literale einer Datei mit der Zeile ihres ersten Vorkommens, getrennt
// nach außerhalb und innerhalb des Körpers jeder Deklaration `func nichtGelieferte`.
func literale(fset *token.FileSet, f *ast.File) (aussen, innen menge) {
	type bereich struct{ von, bis token.Pos }
	var koerper []bereich
	var lits []*ast.BasicLit
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.Name == "nichtGelieferte" && x.Body != nil {
				koerper = append(koerper, bereich{x.Body.Lbrace, x.Body.Rbrace})
			}
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				lits = append(lits, x)
			}
		}
		return true
	})
	aussen, innen = menge{}, menge{}
	for _, l := range lits {
		v, err := strconv.Unquote(l.Value)
		if err != nil || !istSchluessel(v) {
			continue
		}
		ziel := aussen
		if slices.ContainsFunc(koerper, func(k bereich) bool { return l.Pos() > k.von && l.Pos() < k.bis }) {
			ziel = innen
		}
		z := fset.PositionFor(l.Pos(), false).Line
		if alt, da := ziel[v]; !da || z < alt {
			ziel[v] = z
		}
	}
	return aussen, innen
}

// abgelehnt sagt, ob ein Name ein Name aus N ist (gleich) oder an einem Präfix aus N beginnt (SPEC-201 Nr. 4 c); ein
// Element von N, das auf `_` endet, ist ein Präfix. Es gilt für H (keine Pflicht) und für W (Art 10) gleich.
func (s schluessel) abgelehnt(name string) bool {
	for e := range s.n {
		if strings.HasSuffix(e, "_") && strings.HasPrefix(name, e) || !strings.HasSuffix(e, "_") && name == e {
			return true
		}
	}
	return false
}

// vergleicheSchluessel prüft die Beziehungen der Schlüssel (SPEC-201 Nr. 1, 4, 7): jeder Name aus C steht in K, jeder
// Name aus H = C ohne die abgelehnten Namen (Name aus N, Name an einem Präfix aus N) hat eine Zeile, und jeder Name der Tabelle ist ein Name aus C oder (Art 10) aus N.
func vergleicheSchluessel(s schluessel, k, w menge, b *bericht) {
	for name, zeile := range s.c {
		if _, da := k[name]; !da {
			b.befundIn(katSchluessel, artOhneSpezifikation, name, zeile, name+" in konfig*.go (Zeile "+strconv.Itoa(zeile)+"), nicht in SPEC-110")
		}
		if _, tabelle := w[name]; !tabelle && !s.abgelehnt(name) {
			b.befundIn(katSchluessel, artFehlt, name, 0, name+" fehlt im Handbuch")
		}
	}
	for name, zeile := range w {
		z := strconv.Itoa(zeile)
		if s.abgelehnt(name) {
			b.befundIn(katSchluessel, artAbgelehnt, name, zeile, "Handbuch Zeile "+z+" nennt "+name+", den das Binary ablehnt")
		} else if _, da := s.c[name]; !da {
			b.befundIn(katSchluessel, artUnbekannt, name, zeile, "Handbuch Zeile "+z+" nennt "+name+", der nicht in konfig*.go steht")
		}
	}
}
