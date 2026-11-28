package batchcodec

import "bytes"

// zeilenende ist CRLF, das einzige Zeilenende des Rahmens und der Kopfzeilen (SPEC-183).
const zeilenende = "\r\n"

// rahmen zerlegt eine Ebene an ihren Grenzzeilen (SPEC-183 Nr. 2): anzahl zählt alle Inhalte, inhalte trägt die ersten
// hoechstens davon; geschlossen ist falsch, wenn die Ebene keine schließende Grenzzeile hat. Präambel und Epilog gehören zu keinem
// Inhalt, das CRLF vor einer Grenzzeile gehört zu ihr.
func rahmen(ebene []byte, grenze string, hoechstens int) (inhalte [][]byte, anzahl int, geschlossen bool) {
	strich := []byte("--" + grenze)
	anfang := -1
	for pos := 0; ; {
		zeile, _, weiter := bytes.Cut(ebene[pos:], []byte(zeilenende))
		naechste := pos + len(zeile) + len(zeilenende)
		if trifft, schliesst := grenzzeile(zeile, strich); trifft {
			if anfang >= 0 {
				anzahl++
				if len(inhalte) < hoechstens {
					ende := max(anfang, pos-len(zeilenende))
					inhalte = append(inhalte, ebene[anfang:ende:ende])
				}
			}
			if schliesst {
				return inhalte, anzahl, true
			}
			anfang = naechste
		}
		if !weiter {
			return inhalte, anzahl, false
		}
		pos = naechste
	}
}

// grenzzeile erkennt `--`, die Grenzzeichenfolge bytegleich, wahlweise `--` (dann schließend) und beliebig viele SP und HTAB.
func grenzzeile(zeile, strich []byte) (trifft, schliesst bool) {
	rest, ok := bytes.CutPrefix(zeile, strich)
	if !ok {
		return false, false
	}
	rest, schliesst = bytes.CutPrefix(rest, []byte("--"))
	if len(bytes.TrimLeft(rest, " \t")) != 0 {
		return false, false
	}
	return true, schliesst
}
