package main

import (
	"go/ast"
	"go/token"
	"strconv"
)

// istKommando sagt, ob v ganz `^[a-z]+$` trifft (SPEC-201 Nr. 5 a).
func istKommando(v string) bool {
	if v == "" {
		return false
	}
	for i := 0; i < len(v); i++ {
		if v[i] < 'a' || v[i] > 'z' {
			return false
		}
	}
	return true
}

// liesCli liest die Konstanten des Adapters: jede Konstante mit einem Zeichenketten-Literal als Wert, dessen Text
// `^[a-z]+$` trifft, ist ein Kommando (SPEC-201 Nr. 5 a).
func liesCli(wurzel string, b *bericht) menge {
	text, ok := lies(wurzel, pfadCli, "cli.go", b)
	if !ok {
		return nil
	}
	fset, f, zeile := parseGo("cli.go", text)
	if f == nil {
		b.fehler("cli.go nicht lesbar: kein Go (Zeile " + strconv.Itoa(zeile) + ")")
		return nil
	}
	c := menge{}
	ast.Inspect(f, func(n ast.Node) bool {
		if d, ok := n.(*ast.GenDecl); ok && d.Tok == token.CONST {
			for _, l := range kommandoLiterale(d) {
				t, _ := strconv.Unquote(l.Value)
				if _, da := c[t]; !da {
					c[t] = fset.PositionFor(l.Pos(), false).Line
				}
			}
		}
		return true
	})
	if len(c) == 0 {
		b.fehler("cli.go: kein Kommando gefunden")
	}
	return c
}

// kommandoLiterale liefert die Literale der Konstanten einer Deklaration, deren Text ein Kommando ist.
func kommandoLiterale(d *ast.GenDecl) []*ast.BasicLit {
	var out []*ast.BasicLit
	for _, s := range d.Specs {
		v := s.(*ast.ValueSpec)
		for i := range v.Names {
			if i >= len(v.Values) {
				break
			}
			l, ok := v.Values[i].(*ast.BasicLit)
			if !ok || l.Kind != token.STRING {
				continue
			}
			if t, err := strconv.Unquote(l.Value); err == nil && istKommando(t) {
				out = append(out, l)
			}
		}
	}
	return out
}

// vergleicheKommandos prüft die Beziehungen der Kommandos (SPEC-201 Nr. 5, 7): jeder Name aus C hat eine Zeile, jeder
// Name der Tabelle ist ein Name aus C.
func vergleicheKommandos(c, w menge, b *bericht) {
	for name := range c {
		if _, da := w[name]; !da {
			b.befundIn(katKommandos, artFehlt, name, 0, name+" fehlt im Handbuch")
		}
	}
	for name, zeile := range w {
		if _, da := c[name]; !da {
			b.befundIn(katKommandos, artUnbekannt, name, zeile, "Handbuch Zeile "+strconv.Itoa(zeile)+" nennt "+name+", der nicht in cli.go steht")
		}
	}
}
