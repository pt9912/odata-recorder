package policy

import "strings"

// segmentArt unterscheidet die Segmente eines Musters (SPEC-261 Nr. 6 b, c).
type segmentArt int

const (
	segLiteral segmentArt = iota // Text ohne `*`, byte-gleich
	segStern                     // genau ein `*`: anfang davor, ende dahinter, dazwischen mindestens ein Byte
	segRest                      // `**` als letztes Segment: ein oder mehr Segmente
)

type segment struct {
	art          segmentArt
	anfang, ende string // bei segLiteral steht der Text in anfang
}

// freigabe ist ein geparster Eintrag von Schreibfreigaben: Methode und Segmente des Musters.
type freigabe struct {
	methode  string
	segmente []segment
}

// mutierend sind die Methoden, für die eine Routenfreigabe stehen darf (SPEC-261 Nr. 2 b).
func mutierend(m string) bool {
	switch m {
	case "POST", "PUT", "PATCH", "DELETE", "MERGE":
		return true
	}
	return false
}

// parseFreigabe prüft `<Methode> <Muster>` mit genau einem Leerzeichen und liefert die Ursache in
// der Rangfolge leer, trenner, methode, muster (SPEC-261 Nr. 6 a).
func parseFreigabe(e string) (freigabe, string) {
	if e == "" {
		return freigabe{}, UrsacheLeer
	}
	i := strings.IndexByte(e, ' ')
	if i < 0 {
		return freigabe{}, UrsacheTrenner
	}
	if !mutierend(e[:i]) {
		return freigabe{}, UrsacheMethode
	}
	segs, ok := parseMuster(e[i+1:])
	if !ok {
		return freigabe{}, UrsacheMuster
	}
	return freigabe{methode: e[:i], segmente: segs}, ""
}

// parseMuster prüft ein Muster nach SPEC-261 Nr. 6 b; nichts wird dekodiert oder normalisiert.
func parseMuster(m string) ([]segment, bool) {
	if len(m) < 2 || m[0] != '/' || !musterBytes(m) {
		return nil, false
	}
	teile := strings.Split(m[1:], "/")
	segs := make([]segment, 0, len(teile))
	for i, t := range teile {
		s, ok := parseSegment(t, i == len(teile)-1)
		if !ok {
			return nil, false
		}
		segs = append(segs, s)
	}
	return segs, true
}

// musterBytes sagt, ob jedes Byte zur Menge der Muster gehört: 0x21 bis 0x7E außer `?`, `#`, `\`
// und `,`, und `%` nur vor zwei Hexadezimalziffern.
func musterBytes(m string) bool {
	for i := 0; i < len(m); i++ {
		c := m[i]
		if c < 0x21 || c > 0x7E || c == '?' || c == '#' || c == '\\' || c == ',' {
			return false
		}
		if c == '%' && (i+2 >= len(m) || !hexZiffer(m[i+1]) || !hexZiffer(m[i+2])) {
			return false
		}
	}
	return true
}

// parseSegment prüft ein Segment: nicht leer, höchstens ein `*`, `**` nur als ganzes letztes
// Segment, ein Segment ohne `*` nicht `.` oder `..`.
func parseSegment(t string, letztes bool) (segment, bool) {
	n := strings.Count(t, "*")
	switch {
	case t == "":
		return segment{}, false
	case t == "**":
		return segment{art: segRest}, letztes
	case n > 1:
		return segment{}, false
	case n == 1:
		k := strings.IndexByte(t, '*')
		return segment{art: segStern, anfang: t[:k], ende: t[k+1:]}, true
	case punktSegment(t):
		return segment{}, false
	}
	return segment{art: segLiteral, anfang: t}, true
}

// punktSegment sagt, ob ein Segment nach dem Ersetzen jedes `%2e` und `%2E` durch `.` gleich `.`
// oder `..` ist.
func punktSegment(s string) bool {
	s = strings.ReplaceAll(s, "%2e", ".")
	s = strings.ReplaceAll(s, "%2E", ".")
	return s == "." || s == ".."
}

func hexZiffer(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}
