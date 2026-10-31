package urlparser

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func hexWert(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// dekodiere löst `%XX` genau einmal auf (SPEC-161 Nr. 2): `+` bleibt `+`, ein rohes Byte ab 0x80 gilt wie sein Escape.
// Ungültig sind ein `%` ohne zwei Hex-Ziffern, ein Ergebnis ohne gültiges UTF-8 und das Zeichen U+0000.
func dekodiere(s string) (string, bool) {
	if strings.IndexByte(s, '%') < 0 {
		return s, gueltigerText(s)
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			out = append(out, s[i])
			continue
		}
		if i+2 >= len(s) || hexWert(s[i+1]) < 0 || hexWert(s[i+2]) < 0 {
			return "", false
		}
		out = append(out, byte(hexWert(s[i+1])<<4|hexWert(s[i+2])))
		i += 2
	}
	t := string(out)
	return t, gueltigerText(t)
}

func gueltigerText(s string) bool {
	return utf8.ValidString(s) && strings.IndexByte(s, 0) < 0
}

func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

// namenAnfang liefert den längsten Anfang von s, der ein Name sein kann (SPEC-125 Nr. 8 b: Buchstabe oder `_`, dann
// Buchstaben, Ziffern, `_`), und die Zahl seiner Zeichen.
func namenAnfang(s string) (name, rest string, zeichen int) {
	for i, r := range s {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			zeichen++
			continue
		}
		return s[:i], s[i:], zeichen
	}
	return s, "", zeichen
}

// einfacherName meldet, ob s als Ganzes ein Name von 1 bis 128 Zeichen ist.
func einfacherName(s string) bool {
	_, rest, n := namenAnfang(s)
	return rest == "" && n >= 1 && n <= maxName
}
