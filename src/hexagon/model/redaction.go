package model

import (
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

// Marker ist der einzige Platzhalter eines redigierten Header-, Query- und JSON-String-Wertes
// (SPEC-115 Nr. 1).
const Marker = "[REDACTED]"

// ErrKodierung meldet eine ungültige %-Kodierung im Request bei MATCH_INVALID_ENCODING=reject
// (SPEC-116 Nr. 9). Sie ist nicht ErrUngueltig und trägt keinen Wert des Requests.
var ErrKodierung = errors.New("model: ungültige Kodierung im Request")

// asciiKlein setzt allein ASCII-Großbuchstaben herab; andere Bytes bleiben (SPEC-115 Nr. 2).
func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func istTokenByte(c byte) bool {
	return buchstabeOderZiffer(c) ||
		c < 0x80 && strings.IndexByte("!#$%&'*+-.^_`|~", c) >= 0
}

// kanonischerName liefert die kanonische Schreibweise eines HTTP-Token-Namens (`x-api-key` wird
// `X-Api-Key`); ein Name, der kein Token ist, bleibt unverändert (SPEC-115 Nr. 2).
func kanonischerName(name string) string {
	if name == "" {
		return name
	}
	b := []byte(name)
	gross := true
	for i, c := range b {
		if !istTokenByte(c) {
			return name
		}
		switch {
		case gross && c >= 'a' && c <= 'z':
			b[i] = c - ('a' - 'A')
		case !gross && c >= 'A' && c <= 'Z':
			b[i] = c + ('a' - 'A')
		}
		gross = c == '-'
	}
	return string(b)
}

func mengeKlein(namen []string) map[string]bool {
	m := make(map[string]bool, len(namen))
	for _, n := range namen {
		m[asciiKlein(n)] = true
	}
	return m
}

// fasseNamenZusammen führt Einträge, die sich nur in der Schreibweise des Namens unterscheiden,
// zu einem Eintrag unter dem kanonischen Namen zusammen: die Werte folgen einander in der
// Bytefolge der Eingabenamen, danach in ihrer Reihenfolge. Die Eingabe bleibt unverändert.
func fasseNamenZusammen(h map[string][]string) map[string][]string {
	namen := make([]string, 0, len(h))
	for n := range h {
		namen = append(namen, n)
	}
	slices.Sort(namen)
	out := make(map[string][]string, len(h))
	for _, n := range namen {
		k := kanonischerName(n)
		out[k] = append(out[k], h[n]...)
		if len(h[n]) == 0 && out[k] == nil {
			out[k] = []string{}
		}
	}
	return out
}

// RedigiereHeader liefert eine neue Map, in der jeder Header, dessen Name ohne Beachtung der
// ASCII-Schreibweise in namen steht, genau den Wert Marker trägt (SPEC-115 Nr. 2). In
// Location und Content-Location, die nicht ganz redigiert sind, ersetzt sie die Werte der
// Query-Parameter aus parameter (Nr. 3). Das Ergebnis ist nie nil, die Eingabe bleibt unverändert.
func RedigiereHeader(h map[string][]string, namen, parameter []string) map[string][]string {
	redigiert := mengeKlein(namen)
	out := fasseNamenZusammen(h)
	for name, werte := range out {
		switch {
		case redigiert[asciiKlein(name)]:
			out[name] = []string{Marker}
		case urlHeader(name):
			neu := make([]string, len(werte))
			for i, w := range werte {
				neu[i] = redigiereURLWert(w, parameter)
			}
			out[name] = neu
		default:
			out[name] = slices.Clone(werte)
		}
	}
	return out
}

func urlHeader(name string) bool {
	k := asciiKlein(name)
	return k == "location" || k == "content-location"
}

// redigiereURLWert redigiert den Query-Teil (nach dem ersten ? bis zum ersten # oder Ende).
func redigiereURLWert(w string, parameter []string) string {
	vor, rest, ok := strings.Cut(w, "?")
	if !ok {
		return w
	}
	query, fragment := rest, ""
	if i := strings.IndexByte(rest, '#'); i >= 0 {
		query, fragment = rest[:i], rest[i:]
	}
	neu, _ := RedigiereQuery(query, parameter)
	return vor + "?" + neu + fragment
}

// RedigiereQuery ersetzt den Wert jedes Parameters aus parameter durch Marker (SPEC-115 Nr. 4)
// und meldet, ob sich ein Wert änderte. Sie scheitert nie.
func RedigiereQuery(rawQuery string, parameter []string) (string, bool) {
	if rawQuery == "" {
		return "", false
	}
	treffer := mengeKlein(parameter)
	segmente := strings.Split(rawQuery, "&")
	geaendert := false
	for i, s := range segmente {
		schluessel, wert, hatWert := strings.Cut(s, "=")
		if !hatWert {
			continue
		}
		name := schluessel
		if d, ok := dekodiereProzent(schluessel); ok {
			name = d
		}
		if !treffer[asciiKlein(name)] {
			continue
		}
		if wert != Marker {
			geaendert = true
		}
		segmente[i] = schluessel + "=" + Marker
	}
	return strings.Join(segmente, "&"), geaendert
}

// OhneRedigierte liefert die Map ohne jeden Header, der mindestens einen Wert hat und dessen
// Werte alle Marker sind (SPEC-115 Nr. 5). Das Ergebnis ist nie nil.
func OhneRedigierte(h map[string][]string) map[string][]string {
	out := make(map[string][]string, len(h))
	for name, werte := range h {
		if len(werte) > 0 && !slices.ContainsFunc(werte, func(w string) bool { return w != Marker }) {
			continue
		}
		out[name] = slices.Clone(werte)
	}
	return out
}

// JSONPfad ist ein Pfadmuster der Form `$..name` (SPEC-115 Nr. 6).
type JSONPfad struct{ name string }

// ParseJSONPfad nimmt allein `$..name` an, mit name aus einem oder mehr ASCII-Buchstaben,
// Ziffern, `_` und `-`; jede andere Form ist ErrUngueltig, ohne den Text im Fehler.
func ParseJSONPfad(text string) (JSONPfad, error) {
	name, ok := strings.CutPrefix(text, "$..")
	if !ok || name == "" {
		return JSONPfad{}, ungueltig("Pfadmuster hat nicht die Form $..name")
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !buchstabeOderZiffer(c) && c != '_' && c != '-' {
			return JSONPfad{}, ungueltig("Pfadmuster hat nicht die Form $..name")
		}
	}
	return JSONPfad{name: name}, nil
}

// Passt meldet, ob ein Objektmitglied dieses Namens vom Muster gemeint ist: der Vergleich
// ignoriert die ASCII-Schreibweise (SPEC-115 Nr. 6).
func (p JSONPfad) Passt(mitgliedName string) bool {
	return p.name != "" && asciiKlein(p.name) == asciiKlein(mitgliedName)
}

// dekodiereProzent löst %XX auf (`+` bleibt `+`). Ungültig ist ein % ohne zwei Hex-Ziffern
// und ein Ergebnis, das kein gültiges UTF-8 ist (SPEC-116 Nr. 4).
func dekodiereProzent(s string) (string, bool) {
	if !strings.ContainsRune(s, '%') && utf8.ValidString(s) {
		return s, true
	}
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			b = append(b, s[i])
			continue
		}
		if i+2 >= len(s) {
			return "", false
		}
		h, l := hexWert(s[i+1]), hexWert(s[i+2])
		if h < 0 || l < 0 {
			return "", false
		}
		b = append(b, byte(h<<4|l))
		i += 2
	}
	out := string(b)
	return out, utf8.ValidString(out)
}

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

func buchstabeOderZiffer(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}
