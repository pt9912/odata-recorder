package main

import (
	"fmt"
	"unicode"
)

// kategorie nennt die Kategorie, wenn r zur Menge Z gehört (SPEC-233 Nr. 7 a), sonst "".
// Z ist Cc ohne Tabulator und Zeilenvorschub, Cf, Zl, Zp, Zs ohne das Leerzeichen sowie
// U+FFFE und U+FFFF (Kategorie Cn); maßgeblich ist unicode.Version.
func kategorie(r rune) string {
	switch {
	case r == '\t' || r == '\n':
		return ""
	case unicode.Is(unicode.Cc, r):
		return "Cc"
	case unicode.Is(unicode.Cf, r):
		return "Cf"
	case unicode.Is(unicode.Zl, r):
		return "Zl"
	case unicode.Is(unicode.Zp, r):
		return "Zp"
	case unicode.Is(unicode.Zs, r) && r != ' ':
		return "Zs"
	case r == 0xFFFE || r == 0xFFFF:
		return "Cn"
	}
	return ""
}

// kurznamen sind die Kurznamen der 29 Zeichen aus SPEC-233 Nr. 7 f.
func kurznamen() map[rune]string {
	return map[rune]string{
		0x0000: "NUL", 0x000D: "CR", 0x007F: "DEL", 0x0085: "NEL", 0x00A0: "NBSP", 0x00AD: "SHY",
		0x061C: "ALM", 0x180E: "MVS", 0x200B: "ZWSP", 0x200C: "ZWNJ", 0x200D: "ZWJ", 0x200E: "LRM",
		0x200F: "RLM", 0x2028: "LS", 0x2029: "PS", 0x202A: "LRE", 0x202B: "RLE", 0x202C: "PDF",
		0x202D: "LRO", 0x202E: "RLO", 0x202F: "NNBSP", 0x2060: "WJ", 0x2066: "LRI", 0x2067: "RLI",
		0x2068: "FSI", 0x2069: "PDI", 0xFEFF: "BOM", 0xFFFE: "NONCHAR", 0xFFFF: "NONCHAR",
	}
}

// kurzname ist der Kurzname von r oder "".
func kurzname(r rune) string { return kurznamen()[r] }

// befundText ist der Text einer Fundstelle ohne Ort: U+HEX, Kategorie und, wo es einen gibt, der Kurzname.
func befundText(r rune, kat string) string {
	t := fmt.Sprintf("U+%04X %s", r, kat)
	if k := kurzname(r); k != "" {
		t += " " + k
	}
	return t
}

// maskiere macht einen Text ASCII-sicher (SPEC-233 Nr. 10 b): druckbares ASCII bleibt, der
// Rückwärtsschrägstrich wird zu zwei, jedes andere Byte zu \xHH in Großbuchstaben.
func maskiere(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\':
			b = append(b, '\\', '\\')
		case c >= 0x20 && c <= 0x7E:
			b = append(b, c)
		default:
			b = fmt.Appendf(b, "\\x%02X", c)
		}
	}
	return string(b)
}
