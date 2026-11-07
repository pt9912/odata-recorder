package jsoncodecv2

import (
	"strings"
	"unicode/utf8"
)

// maxWurzel ist die Länge in Bytes, die Wurzel haben darf (SPEC-172 Nr. 6 h).
const maxWurzel = 2048

func buchstabe(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func hexZiffer(c byte) bool { return ziffer(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' }

// hostZeichen sind die Zeichen eines Hosts ohne eckige Klammern.
func hostZeichen(c byte) bool { return buchstabe(c) || ziffer(c) || strings.IndexByte("-._~", c) >= 0 }

// wurzelGueltig prüft die enge Form von Wurzel (SPEC-172 Nr. 6 h): Schema, Host, Port, Segmente.
func wurzelGueltig(w string) bool {
	if len(w) > maxWurzel {
		return false
	}
	rest, ok := strings.CutPrefix(w, "http://")
	if !ok {
		if rest, ok = strings.CutPrefix(w, "https://"); !ok {
			return false
		}
	}
	rest, ok = hostEnde(rest)
	if !ok {
		return false
	}
	if p, ok := strings.CutPrefix(rest, ":"); ok {
		i := 0
		for i < len(p) && ziffer(p[i]) {
			i++
		}
		if i == 0 || i > 5 {
			return false
		}
		rest = p[i:]
	}
	if rest == "" {
		return true
	}
	segmente, ok := strings.CutPrefix(rest, "/")
	if !ok {
		return false
	}
	for _, seg := range strings.Split(segmente, "/") {
		if !segmentGueltig(seg) {
			return false
		}
	}
	return true
}

// hostEnde liest den Host am Anfang von s und liefert den Rest (leer, `:port` oder `/…`).
func hostEnde(s string) (string, bool) {
	if strings.HasPrefix(s, "[") {
		i := strings.IndexByte(s, ']')
		if i < 2 {
			return "", false
		}
		for j := 1; j < i; j++ {
			if !hexZiffer(s[j]) && s[j] != ':' && s[j] != '.' {
				return "", false
			}
		}
		return s[i+1:], true
	}
	i := 0
	for i < len(s) && hostZeichen(s[i]) {
		i++
	}
	return s[i:], i > 0
}

// segmentGueltig prüft ein Segment: nicht leer, erlaubte Zeichen, `%` mit zwei Hex-Ziffern, dekodiert nicht `.` und `..`.
func segmentGueltig(seg string) bool {
	if seg == "" {
		return false
	}
	var dek strings.Builder
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		switch {
		case hostZeichen(c):
			dek.WriteByte(c)
		case c == '%' && i+2 < len(seg) && hexZiffer(seg[i+1]) && hexZiffer(seg[i+2]):
			dek.WriteByte(byte(hexWert(seg[i+1])<<4 | hexWert(seg[i+2])))
			i += 2
		default:
			return false
		}
	}
	d := dek.String()
	return d != "." && d != ".."
}

// kodiere schreibt s bytegenau: ASCII-Buchstabe, ASCII-Ziffer und `-._~',=` unverändert, jedes andere Byte als `%` und
// zwei Hex-Ziffern in Großbuchstaben (SPEC-172 Nr. 6 h).
func kodiere(s string) string {
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		if buchstabe(c) || ziffer(c) || strings.IndexByte("-._~',=", c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte("0123456789ABCDEF"[c>>4])
		b.WriteByte("0123456789ABCDEF"[c&15])
	}
	return b.String()
}

// schluesselGueltig: nicht leer, gültiges UTF-8, ohne U+0000, keine Längengrenze (SPEC-168 Nr. 3 b).
func schluesselGueltig(s string) bool {
	return s != "" && utf8.ValidString(s) && !strings.Contains(s, "\x00")
}

// adresse bildet `<Wurzel>/<Set>(<Schlüssel>)` nach Prüfung von Wurzel und Schlüssel; das Set prüft der Aufrufer.
func adresse(wurzel, set, schluessel string) (string, error) {
	if !wurzelGueltig(wurzel) {
		return "", ungueltig("Wurzel hat nicht die Form von SPEC-172 Nr. 6 h")
	}
	if !schluesselGueltig(schluessel) {
		return "", ungueltig("Schlüssel ist leer, ungültiges UTF-8 oder enthält U+0000")
	}
	return wurzel + "/" + kodiere(set) + "(" + kodiere(schluessel) + ")", nil
}
