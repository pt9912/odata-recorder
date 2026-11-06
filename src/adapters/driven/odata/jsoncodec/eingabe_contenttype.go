package jsoncodec

import "strings"

// contentTypeGueltig prüft die Kopfzeilen Content-Type nach SPEC-181 Nr. 1: genau eine Zeile, deren Medientyp `application/json`
// ist, und jedes Vorkommen von `charset` mit dem Wert `utf-8`; die Schreibweise zählt nur in ASCII nicht.
func contentTypeGueltig(zeilen []string) bool {
	if len(zeilen) != 1 {
		return false
	}
	teile := strings.Split(zeilen[0], ";")
	if !gleichASCII(randGetrimmt(teile[0]), "application/json") {
		return false
	}
	for _, teil := range teile[1:] {
		name, wert, hatWert := strings.Cut(teil, "=")
		if !gleichASCII(randGetrimmt(name), "charset") {
			continue
		}
		if !hatWert || !gleichASCII(randGetrimmt(wert), "utf-8") {
			return false
		}
	}
	return true
}

// randGetrimmt entfernt Leerzeichen und Tabulatoren am Rand.
func randGetrimmt(s string) string { return strings.Trim(s, " \t") }

// gleichASCII vergleicht ohne Beachtung der Schreibweise der ASCII-Buchstaben; ein anderes Zeichen wird nicht gefaltet
// (kein U+017F für `s`). soll steht in Kleinbuchstaben.
func gleichASCII(s, soll string) bool {
	if len(s) != len(soll) {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != soll[i] {
			return false
		}
	}
	return true
}
