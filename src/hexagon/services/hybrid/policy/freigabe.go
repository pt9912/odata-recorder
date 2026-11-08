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

// routeFrei sagt, ob ein Eintrag von Schreibfreigaben Methode und Pfad deckt (SPEC-261 Nr. 6 c, d).
func (p *Policy) routeFrei(methode, pfad string) bool {
	if !strings.HasPrefix(pfad, "/") {
		return false
	}
	teile := strings.Split(pfad[1:], "/")
	for _, f := range p.freigaben {
		if f.methode == methode && passt(f, teile) {
			return true
		}
	}
	return false
}

// passt vergleicht die Segmente des Pfades mit denen des Musters: einmal von vorn nach hinten, ohne
// Zurücksetzen (SPEC-261 Nr. 6 f). Die Zahl der Segmente ist gleich, außer bei `**` am Ende.
func passt(f freigabe, teile []string) bool {
	n := len(f.segmente)
	rest := f.segmente[n-1].art == segRest
	if rest && len(teile) < n || !rest && len(teile) != n {
		return false
	}
	for i, s := range f.segmente {
		switch s.art {
		case segLiteral:
			if teile[i] != s.anfang {
				return false
			}
		case segStern:
			if !sternPasst(s, teile[i]) {
				return false
			}
		case segRest:
			for _, t := range teile[i:] {
				if !gedeckt(t) {
					return false
				}
			}
		}
	}
	return true
}

// sternPasst sagt, ob ein Segment mit einem `*` passt: Anfang und Ende byte-gleich, dazwischen
// mindestens ein Byte der Menge W; deckt das `*` das ganze Segment, ist es weder `.` noch `..`.
func sternPasst(s segment, t string) bool {
	if len(t) < len(s.anfang)+len(s.ende)+1 || !strings.HasPrefix(t, s.anfang) || !strings.HasSuffix(t, s.ende) {
		return false
	}
	mitte := t[len(s.anfang) : len(t)-len(s.ende)]
	if s.anfang == "" && s.ende == "" {
		return gedeckt(mitte)
	}
	return wBytes(mitte)
}

// gedeckt sagt, ob ein ganzes Segment von `*` oder `**` gedeckt werden darf: nicht leer, nur Bytes
// der Menge W, weder `.` noch `..`.
func gedeckt(t string) bool {
	return t != "" && t != "." && t != ".." && wBytes(t)
}

// wBytes sagt, ob jedes Byte zur Menge W gehört: 0x21 bis 0x7E außer `/`, `%`, `\`, `?` und `#`.
func wBytes(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x21 || c > 0x7E || c == '/' || c == '%' || c == '\\' || c == '?' || c == '#' {
			return false
		}
	}
	return true
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
		if punktEinzel(t[:k], t[k+1:]) {
			return segment{}, false
		}
		return segment{art: segStern, anfang: t[:k], ende: t[k+1:]}, true
	case punktSegment(t):
		return segment{}, false
	}
	return segment{art: segLiteral, anfang: t}, true
}

// punktSegment sagt, ob ein Segment nach dem Ersetzen jedes `%2e` und `%2E` durch `.` gleich `.`
// oder `..` ist.
func punktSegment(s string) bool {
	s = ersetzePunkte(s)
	return s == "." || s == ".."
}

// punktEinzel sagt, ob ein Segment mit einem `*` (Text davor und dahinter) nach dem Ersetzen jedes
// `%2e` und `%2E` durch `.` und dem Streichen des `*` gleich `.` ist: das `*` könnte den Rest
// eines Punkt-Segments decken (SPEC-261 Nr. 6 b, e).
func punktEinzel(vor, nach string) bool {
	return ersetzePunkte(vor+nach) == "."
}

func ersetzePunkte(s string) string {
	s = strings.ReplaceAll(s, "%2e", ".")
	return strings.ReplaceAll(s, "%2E", ".")
}

func hexZiffer(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}
