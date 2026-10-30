package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// muster ist ein Muster der Mengen A, B und C (SPEC-129 Nr. 10 bis 12); der erste Buchstabe der
// Kennung nennt die Menge.
type muster struct {
	id, ausdruck string
}

// mengen liefert die 44 Muster: A Planung und Ausblick (16), B Chronik (14), C Kennungen und
// Pfade der Planung (14). Die Ausdrücke sind RE2 auf dem normalisierten Text (Nr. 13 a, b).
func mengen() []muster {
	return []muster{
		{"A1", `geplant\p{L}*`},
		{"A2", `planung\p{L}*`},
		{"A3", `vorgesehen\p{L}*`},
		{"A4", `(künftig|kuenftig)\p{L}*`},
		{"A5", `(zukünftig|zukuenftig)\p{L}*`},
		{"A6", `zukunft\p{L}*`},
		{"A7", `(spät|spaet)\p{L}*`},
		{"A8", `(demnächst|demnaechst)`},
		{"A9", `bald\p{L}*`},
		{"A10", `folgt`},
		{"A11", `noch nicht`},
		{"A12", `noch kein\p{L}*`},
		{"A13", `nicht (verfügbar|verfuegbar)`},
		{"A14", `wird kommen`},
		{"A15", `ausblick\p{L}*`},
		{"A16", `irgendwann`},
		{"B1", `bisher\p{L}*`},
		{"B2", `(früher|frueher)\p{L}*`},
		{"B3", `neu (in|ab|seit)`},
		{"B4", `seit (version|release|v\d)`},
		{"B5", `ab (version|release|v\d)`},
		{"B6", `nicht mehr`},
		{"B7", `(änderungs|aenderungs)\p{L}*`},
		{"B8", `changelog`},
		{"B9", `release[ -]?notes?`},
		{"B10", `\p{L}*verlauf\p{L}*`},
		{"B11", `\p{L}*historie\p{L}*`},
		{"B12", `handbuch-version`},
		{"B13", `wurden? (geändert|eingeführt|entfernt|ergänzt|hinzugefügt|ersetzt|umbenannt|behoben|korrigiert|verbessert)`},
		{"B14", `(damals|ehemals|vormals|inzwischen|mittlerweile)`},
		{"C1", `roadmap\p{L}*`},
		{"C2", `welle\p{L}*`},
		{"C3", `slice\p{L}*`},
		{"C4", `adr-([a-z]+-)?\d{4}`},
		{"C5", `lh-[a-z]{2}(-[a-z]+)?-\d+`},
		{"C6", `spec-\d{3}`},
		{"C7", `arc-\d{3}`},
		{"C8", `mr-\d{3}`},
		{"C9", `co-\d{3}`},
		{"C10", `rc-\d{3}`},
		{"C11", `beo-[a-z]\p{L}*`},
		{"C12", `(docs/plan/)\p{L}*`},
		{"C13", `(spec/)\p{L}*`},
		{"C14", `(harness/vertraege)\p{L}*`},
	}
}

// mengeVon liefert die Nummer des Vertrags und den Namen der Menge eines Musters.
func mengeVon(id string) (int, string) {
	switch id[0] {
	case 'A':
		return nrPlanung, "Planung"
	case 'B':
		return nrChronik, "Chronik"
	}
	return nrKennung, "Kennung"
}

// normalisiere verbindet die Zeilen mit einem Leerzeichen, entfernt `*`, `_` und Backtick, macht
// aus Tabulator und den Zeichen der Kategorien Zs, Zl, Zp ein Leerzeichen, faltet Folgen von
// Leerzeichen und schreibt klein (SPEC-129 Nr. 13 a).
func normalisiere(zeilen []string) string {
	s := strings.Join(zeilen, " ")
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '*' || r == '_' || r == '`':
			return -1
		case r == '\t' || unicode.In(r, unicode.Zs, unicode.Zl, unicode.Zp):
			return ' '
		}
		return r
	}, s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.ToLower(s)
}

// trifft liefert den ersten Treffer von re in text, der am Anfang oder nach einem Nicht-Buchstaben
// beginnt und am Ende oder vor einem Nicht-Buchstaben endet (SPEC-129 Nr. 13 b).
func trifft(re *regexp.Regexp, text string) (string, bool) {
	for off := 0; off < len(text); {
		loc := re.FindStringIndex(text[off:])
		if loc == nil {
			return "", false
		}
		s, e := off+loc[0], off+loc[1]
		vor, _ := utf8.DecodeLastRuneInString(text[:s])
		nach, _ := utf8.DecodeRuneInString(text[e:])
		if (s == 0 || !unicode.IsLetter(vor)) && (e == len(text) || !unicode.IsLetter(nach)) {
			return text[s:e], true
		}
		_, w := utf8.DecodeRuneInString(text[s:])
		off = s + w
	}
	return "", false
}

// mengenBefunde meldet je Absatz und Muster höchstens einen Treffer; im Zitat gilt nur die Menge C
// (SPEC-129 Nr. 13 c, d).
func mengenBefunde(as []absatz) []befund {
	var bf []befund
	ms := mengen()
	res := make([]*regexp.Regexp, len(ms))
	for i, m := range ms {
		res[i] = regexp.MustCompile(m.ausdruck)
	}
	for _, a := range as {
		text := normalisiere(a.zeilen)
		for i, m := range ms {
			nr, name := mengeVon(m.id)
			if a.zitat && nr != nrKennung {
				continue
			}
			if t, ok := trifft(res[i], text); ok {
				bf = append(bf, befund{nr: nr, zeile: a.zeile, text: fmt.Sprintf("Zeile %d: %s %s: %s", a.zeile, name, m.id, t)})
			}
		}
	}
	return bf
}

// zielliste liefert die Namen der Zeilen `^[a-z0-9-]+:.*##` in Makefile, harness/mk/*.mk, d-check.mk
// und repo.mk unter wurzel, die Regel von `make help` (SPEC-129 Nr. 9 a).
func zielliste(wurzel string) map[string]bool {
	re := regexp.MustCompile(`^([a-z0-9-]+):.*##`)
	ziele := map[string]bool{}
	for _, d := range makeDateien(wurzel, true) {
		for _, z := range dateiZeilen(wurzel, d) {
			if m := re.FindStringSubmatch(z); m != nil {
				ziele[m[1]] = true
			}
		}
	}
	return ziele
}

func dateiZeilen(wurzel, datei string) []string {
	return zeilenVon(lies(wurzel, datei))
}

// codeSpan ist ein Inline-Code-Span nach CommonMark: eine Folge von n Backticks öffnet, die nächste
// Folge von genau n Backticks schließt; zeile ist die Zeile, in der er öffnet.
type codeSpan struct {
	text  string
	zeile int
}

func codeSpans(a absatz) []codeSpan {
	s := strings.Join(a.zeilen, "\n")
	var out []codeSpan
	for i := 0; i < len(s); {
		if s[i] != '`' {
			i++
			continue
		}
		n := laufLaenge(s, i)
		j := i + n
		k := schliessendes(s, j, n)
		if k < 0 {
			i = j
			continue
		}
		out = append(out, codeSpan{text: spanText(s[j:k]), zeile: a.zeile + strings.Count(s[:i], "\n")})
		i = k + n
	}
	return out
}

// laufLaenge ist die Länge der Folge von Backticks, die in s an Stelle i beginnt.
func laufLaenge(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	return n
}

// schliessendes liefert den Anfang der nächsten Folge von genau n Backticks ab ab, sonst -1.
func schliessendes(s string, ab, n int) int {
	for k := ab; k < len(s); {
		if s[k] != '`' {
			k++
			continue
		}
		m := laufLaenge(s, k)
		if m == n {
			return k
		}
		k += m
	}
	return -1
}

// spanText macht aus Zeilenumbrüchen Leerzeichen und entfernt ein Leerzeichen an beiden Enden,
// wenn der Text nicht nur aus Leerzeichen besteht.
func spanText(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) >= 2 && s[0] == ' ' && s[len(s)-1] == ' ' && strings.Trim(s, " ") != "" {
		return s[1 : len(s)-1]
	}
	return s
}

// makeBefunde prüft jeden Span der Prosa, dessen Text mit `make ` beginnt (SPEC-129 Nr. 9 b).
func makeBefunde(as []absatz, ziele map[string]bool) []befund {
	var bf []befund
	for _, a := range as {
		if a.zitat {
			continue
		}
		for _, sp := range codeSpans(a) {
			if !strings.HasPrefix(sp.text, "make ") {
				continue
			}
			w := strings.Fields(sp.text)
			if len(w) < 2 {
				continue
			}
			switch {
			case strings.HasPrefix(w[1], "-") || strings.Contains(w[1], "="):
				bf = append(bf, befund{nr: nrMake, zeile: sp.zeile, text: fmt.Sprintf("Zeile %d: make-Aufruf ohne Ziel an erster Stelle: %s", sp.zeile, sp.text)})
			case !ziele[w[1]]:
				bf = append(bf, befund{nr: nrMake, zeile: sp.zeile, text: fmt.Sprintf("Zeile %d: make-Ziel %s unbekannt", sp.zeile, w[1])})
			}
		}
	}
	return bf
}
