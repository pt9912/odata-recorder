package urlparserv2

import (
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// maxBruchteil ist die Zahl der Ziffern des Sekundenbruchteils von `datetime'…'` (SPEC-173 Nr. 3).
const maxBruchteil = 7

// ziffernWert liest einen nichtleeren Text aus ASCII-Ziffern.
func ziffernWert(s string) (int, bool) {
	if !nurZiffern(s) {
		return 0, false
	}
	v := 0
	for i := range len(s) {
		v = v*10 + int(s[i]-'0')
	}
	return v, true
}

// datumTeile zerlegt `JJJJ-MM-TT` mit einem Jahr von 0001 bis 9999 und einem Tag, den der Monat hat.
func datumTeile(s string) (jahr, monat, tag int, ok bool) {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return 0, 0, 0, false
	}
	var o1, o2, o3 bool
	jahr, o1 = ziffernWert(s[:4])
	monat, o2 = ziffernWert(s[5:7])
	tag, o3 = ziffernWert(s[8:])
	if !o1 || !o2 || !o3 || jahr < 1 || monat < 1 || monat > 12 || tag < 1 {
		return 0, 0, 0, false
	}
	t := time.Date(jahr, time.Month(monat), tag, 0, 0, 0, 0, time.UTC)
	return jahr, monat, tag, t.Day() == tag
}

// uhrzeitTeile zerlegt `hh:mm`, danach `:ss` und dann `.f` (1 bis 7 Ziffern); rest ist, was dahinter steht.
func uhrzeitTeile(s string) (h, mi, sek int, bruch, rest string, ok bool) {
	if len(s) < 5 || s[2] != ':' {
		return 0, 0, 0, "", "", false
	}
	var o1, o2 bool
	h, o1 = ziffernWert(s[:2])
	mi, o2 = ziffernWert(s[3:5])
	if !o1 || !o2 || h > 23 || mi > 59 {
		return 0, 0, 0, "", "", false
	}
	if !strings.HasPrefix(s[5:], ":") {
		return h, mi, 0, "", s[5:], true
	}
	sek, bruch, rest, ok = sekunden(s[5:])
	return h, mi, sek, bruch, rest, ok
}

// sekunden zerlegt `:ss`, danach `.f` mit 1 bis 7 Ziffern.
func sekunden(s string) (sek int, bruch, rest string, ok bool) {
	if len(s) < 3 {
		return 0, "", "", false
	}
	sek, ok = ziffernWert(s[1:3])
	if !ok || sek > 59 {
		return 0, "", "", false
	}
	rest = s[3:]
	if !strings.HasPrefix(rest, ".") {
		return sek, "", rest, true
	}
	n := 1
	for n < len(rest) && rest[n] >= '0' && rest[n] <= '9' {
		n++
	}
	if n == 1 || n-1 > maxBruchteil {
		return 0, "", "", false
	}
	return sek, rest[1:n], rest[n:], true
}

// zeitpunkt liest `datetime'JJJJ-MM-TTThh:mm'` mit `:ss` und `.f`, ohne Zonenangabe, und schreibt den Zeitpunkt als UTC
// (SPEC-173 Nr. 3); der Bruchteil bleibt ohne Nullen am Ende erhalten und wird nicht gekürzt.
func zeitpunkt(s string) (model.Literal, bool) {
	p, inhalt, ok := praefixLiteral(s)
	if !ok || p != "datetime" || len(inhalt) < 16 || inhalt[10] != 'T' {
		return model.Literal{}, false
	}
	jahr, monat, tag, ok := datumTeile(inhalt[:10])
	if !ok {
		return model.Literal{}, false
	}
	h, mi, sek, bruch, rest, ok := uhrzeitTeile(inhalt[11:])
	if !ok || rest != "" {
		return model.Literal{}, false
	}
	text := time.Date(jahr, time.Month(monat), tag, h, mi, sek, 0, time.UTC).Format("2006-01-02T15:04:05")
	if bruch = strings.TrimRight(bruch, "0"); bruch != "" {
		text += "." + bruch
	}
	return model.Literal{Art: model.LiteralZeitpunkt, Text: text + "Z"}, true
}
