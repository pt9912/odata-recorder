package main

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// md ist ein Text in Zeilen mit der Lage der Zäune (SPEC-201 Nr. 6 a).
type md struct {
	zeilen []string
	zaun   []bool // die Zeile gehört zu einem Zaun (öffnende, innere und schließende)
	offen  int    // Zeile (ab 1) des Zauns ohne Schluss, sonst 0
}

// zerlege trennt an `\n`; ein `\r` unmittelbar davor gehört nicht zur Zeile, ein `\r` am Ende ohne `\n` schon.
func zerlege(text string) []string {
	z := strings.Split(text, "\n")
	for i := 0; i < len(z)-1; i++ {
		z[i] = strings.TrimSuffix(z[i], "\r")
	}
	if n := len(z); z[n-1] == "" {
		z = z[:n-1]
	}
	return z
}

func rechtsGetrimmt(s string) string { return strings.TrimRightFunc(s, unicode.IsSpace) }

func neuesMd(text string) md {
	z := zerlege(text)
	m := md{zeilen: z, zaun: make([]bool, len(z))}
	for i, l := range z {
		switch {
		case m.offen > 0:
			m.zaun[i] = true
			if rechtsGetrimmt(l) == "```" {
				m.offen = 0
			}
		case strings.HasPrefix(l, "```") && !strings.Contains(l[3:], "`"):
			m.zaun[i] = true
			m.offen = i + 1
		}
	}
	return m
}

// ueberschrift liest eine Überschrift außerhalb von Zäunen: Ebene (ein bis sechs `#`) und Text.
func (m md) ueberschrift(i int) (int, string, bool) {
	l := m.zeilen[i]
	n := 0
	for n < len(l) && l[n] == '#' {
		n++
	}
	if m.zaun[i] || n < 1 || n > 6 || n >= len(l) || l[n] != ' ' {
		return 0, "", false
	}
	return n, rechtsGetrimmt(l[n+1:]), true
}

// spanne ist ein Bereich von Zeilen: von (einschließlich) bis bis (ausschließlich), als Indizes.
type spanne struct{ von, bis int }

// abschnitte liefert den Abschnitt jeder Überschrift mit genau diesem Text: bis zur nächsten Überschrift
// derselben oder einer höheren Ebene (SPEC-201 Nr. 6 b).
func (m md) abschnitte(text string) []spanne {
	var out []spanne
	for i := range m.zeilen {
		e, t, ok := m.ueberschrift(i)
		if !ok || t != text {
			continue
		}
		bis := len(m.zeilen)
		for j := i + 1; j < len(m.zeilen); j++ {
			if e2, _, ok := m.ueberschrift(j); ok && e2 <= e {
				bis = j
				break
			}
		}
		out = append(out, spanne{i + 1, bis})
	}
	return out
}

// tabellen liefert die Tabellen einer Spanne: Folgen von Zeilen außerhalb von Zäunen, die mit `|` beginnen.
func (m md) tabellen(s spanne) [][]int {
	var out [][]int
	var lauf []int
	for i := s.von; i < s.bis; i++ {
		if !m.zaun[i] && strings.HasPrefix(m.zeilen[i], "|") {
			lauf = append(lauf, i)
			continue
		}
		if lauf != nil {
			out = append(out, lauf)
			lauf = nil
		}
	}
	if lauf != nil {
		out = append(out, lauf)
	}
	return out
}

// zellen trennt eine Tabellenzeile an `|` außer `\|`; der Anfangsstrich steht, der Schlussstrich darf
// fehlen, jede Zelle ist getrimmt (SPEC-201 Nr. 6 c).
func zellen(zeile string) []string {
	z := strings.TrimPrefix(rechtsGetrimmt(zeile), "|")
	if strings.HasSuffix(z, "|") && !strings.HasSuffix(z, `\|`) {
		z = z[:len(z)-1]
	}
	var out []string
	var zelle strings.Builder
	for i := 0; i < len(z); i++ {
		switch {
		case z[i] == '\\' && i+1 < len(z) && z[i+1] == '|':
			zelle.WriteString(`\|`)
			i++
		case z[i] == '|':
			out = append(out, strings.TrimSpace(zelle.String()))
			zelle.Reset()
		default:
			zelle.WriteByte(z[i])
		}
	}
	return append(out, strings.TrimSpace(zelle.String()))
}

// trennzeile sagt, ob jede Zelle aus `-` besteht, mit einem `:` an einem oder beiden Enden.
func trennzeile(zellen []string) bool {
	for _, c := range zellen {
		c = strings.TrimSuffix(strings.TrimPrefix(c, ":"), ":")
		if c == "" || strings.Trim(c, "-") != "" {
			return false
		}
	}
	return true
}

// inlineCode liest eine Zelle, die genau ein Inline-Code-Span ist: Backtick, Text ohne Backtick, Backtick.
func inlineCode(zelle string) (string, bool) {
	if len(zelle) < 3 || zelle[0] != '`' || zelle[len(zelle)-1] != '`' {
		return "", false
	}
	t := zelle[1 : len(zelle)-1]
	return t, !strings.Contains(t, "`")
}

// tabellenForm nennt, wie eine Tabelle heißt und wie sie aussehen muss; die Texte der Eingabefehler setzen
// sich aus quelle und ort zusammen (SPEC-201 Nr. 6 c, 8).
type tabellenForm struct {
	quelle string
	ort    string
	kopf   []string
	daten  bool // eine Tabelle ohne Datenzeile ist Eingabefehler
}

type datenzeile struct {
	n      int // Zeile ab 1
	zellen []string
}

// tabelle liest die eine Tabelle der Spanne; jeder Verstoß gegen die Form ist ein Eingabefehler.
func (m md) tabelle(s spanne, f tabellenForm, b *bericht) []datenzeile {
	ts := m.tabellen(s)
	if len(ts) != 1 {
		art := "fehlt"
		if len(ts) > 1 {
			art = "mehrdeutig"
		}
		b.fehler(fmt.Sprintf("%s: Tabelle%s %s", f.quelle, f.ort, art))
		return nil
	}
	t := ts[0]
	if !slices.Equal(zellen(m.zeilen[t[0]]), f.kopf) {
		b.fehler(fmt.Sprintf("%s: Kopf%s weicht ab", f.quelle, f.ort))
		return nil
	}
	if len(t) < 2 || !trennzeile(zellen(m.zeilen[t[1]])) {
		b.fehler(fmt.Sprintf("%s: Trennzeile%s fehlt", f.quelle, f.ort))
		return nil
	}
	if f.daten && len(t) < 3 {
		b.fehler(f.quelle + ": keine Datenzeile")
	}
	var out []datenzeile
	for _, i := range t[2:] {
		c := zellen(m.zeilen[i])
		if len(c) != len(f.kopf) {
			b.fehler(fmt.Sprintf("%s Zeile %d: %d Zellen statt %d", f.quelle, i+1, len(c), len(f.kopf)))
			continue
		}
		out = append(out, datenzeile{i + 1, c})
	}
	return out
}
