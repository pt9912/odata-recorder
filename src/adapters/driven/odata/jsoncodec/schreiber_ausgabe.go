package jsoncodec

import "unicode/utf8"

// ausgabe baut die Bytes einer Antwort auf; ein Text, der kein gültiges UTF-8 ist, setzt err (SPEC-169 Nr. 1).
type ausgabe struct {
	b   []byte
	err error
}

const hexKlein = "0123456789abcdef"

// roh hängt Text an, der aus dem Code oder aus geprüften Teilen stammt.
func (a *ausgabe) roh(s string) { a.b = append(a.b, s...) }

// text hängt s als JSON-String an: `"` und `\` mit Backslash, jedes Zeichen unter U+0020 als Escape aus Backslash, `u`
// und vier Hex-Ziffern in Kleinbuchstaben, jedes andere Zeichen unverändert (SPEC-169 Nr. 1).
func (a *ausgabe) text(s string) {
	if !utf8.ValidString(s) {
		a.err = ungueltig("Text ist kein gültiges UTF-8")
		return
	}
	a.b = append(a.b, '"')
	for i := range len(s) {
		c := s[i]
		switch {
		case c == '"' || c == '\\':
			a.b = append(a.b, '\\', c)
		case c < 0x20:
			a.b = append(a.b, '\\', 'u', '0', '0', hexKlein[c>>4], hexKlein[c&15])
		default:
			a.b = append(a.b, c)
		}
	}
	a.b = append(a.b, '"')
}

// mitglied hängt `,` (außer vor dem ersten Mitglied) und `"<name>":` an.
func (a *ausgabe) mitglied(erstes *bool, name string) {
	if !*erstes {
		a.b = append(a.b, ',')
	}
	*erstes = false
	a.text(name)
	a.b = append(a.b, ':')
}
