package jsoncodec

import (
	"strings"
	"time"
)

// maxBruchteil ist die Zahl der Ziffern, die ein Sekundenbruchteil haben darf.
const maxBruchteil = 12

// ziffernWert liest einen nichtleeren Text aus ASCII-Ziffern; die Aufrufer übergeben Teile fester Länge.
func ziffernWert(s string) (int, bool) {
	v := 0
	for i := range len(s) {
		if !ziffer(s[i]) {
			return 0, false
		}
		v = v*10 + int(s[i]-'0')
	}
	return v, true
}

// tageImMonat ist die Länge des Monats im gregorianischen Kalender.
func tageImMonat(jahr, monat int) int {
	switch monat {
	case 4, 6, 9, 11:
		return 30
	case 2:
		if jahr%4 == 0 && (jahr%100 != 0 || jahr%400 == 0) {
			return 29
		}
		return 28
	}
	return 31
}

// datumTeile zerlegt `JJJJ-MM-TT` mit einem Jahr von 0001 bis 9999 und einem Tag, den der Monat hat.
func datumTeile(text string) (jahr, monat, tag int, ok bool) {
	if len(text) != 10 || text[4] != '-' || text[7] != '-' {
		return 0, 0, 0, false
	}
	var o1, o2, o3 bool
	jahr, o1 = ziffernWert(text[:4])
	monat, o2 = ziffernWert(text[5:7])
	tag, o3 = ziffernWert(text[8:])
	if !o1 || !o2 || !o3 || jahr < 1 || monat < 1 || monat > 12 || tag < 1 || tag > tageImMonat(jahr, monat) {
		return 0, 0, 0, false
	}
	return jahr, monat, tag, true
}

// datumText ist die kanonische Schreibweise einer Edm.Date: der Text selbst, wenn er ein Datum ist (SPEC-126 Nr. 10 d).
func datumText(text string) (string, bool) {
	if _, _, _, ok := datumTeile(text); !ok {
		return "", false
	}
	return text, true
}

// uhrzeit zerlegt `hh:mm`, optional gefolgt von `:ss` und `.f`; rest ist der Versatz.
func uhrzeit(text string) (h, mi, s int, bruch, rest string, ok bool) {
	if len(text) < 5 || text[2] != ':' {
		return 0, 0, 0, "", "", false
	}
	var o1, o2 bool
	h, o1 = ziffernWert(text[:2])
	mi, o2 = ziffernWert(text[3:5])
	if !o1 || !o2 || h > 23 || mi > 59 {
		return 0, 0, 0, "", "", false
	}
	if !strings.HasPrefix(text[5:], ":") {
		return h, mi, 0, "", text[5:], true
	}
	s, bruch, rest, ok = sekunden(text[5:])
	return h, mi, s, bruch, rest, ok
}

// sekunden zerlegt `:ss` und den Bruchteil `.f` (1 bis 12 Ziffern), der nur hinter den Sekunden steht.
func sekunden(text string) (s int, bruch, rest string, ok bool) {
	if len(text) < 3 {
		return 0, "", "", false
	}
	s, ok = ziffernWert(text[1:3])
	rest = text[3:]
	if !ok || s > 59 {
		return 0, "", "", false
	}
	if !strings.HasPrefix(rest, ".") {
		return s, "", rest, true
	}
	i := 1
	for i < len(rest) && ziffer(rest[i]) {
		i++
	}
	if i == 1 || i-1 > maxBruchteil {
		return 0, "", "", false
	}
	return s, rest[1:i], rest[i:], true
}

// versatzMinuten liest `Z` oder `+hh:mm` oder `-hh:mm`.
func versatzMinuten(text string) (int, bool) {
	if text == "Z" {
		return 0, true
	}
	if len(text) != 6 || (text[0] != '+' && text[0] != '-') || text[3] != ':' {
		return 0, false
	}
	oh, o1 := ziffernWert(text[1:3])
	om, o2 := ziffernWert(text[4:])
	if !o1 || !o2 || oh > 23 || om > 59 {
		return 0, false
	}
	if text[0] == '-' {
		return -(oh*60 + om), true
	}
	return oh*60 + om, true
}

// zeitpunktText ist die kanonische Schreibweise einer Edm.DateTimeOffset: der Zeitpunkt in UTC, der Bruchteil ohne Nullen am
// Ende (SPEC-126 Nr. 10 d); ein Zeitpunkt, dessen UTC-Jahr außerhalb von 0001 bis 9999 liegt, ist keiner.
func zeitpunktText(text string) (string, bool) {
	if len(text) < 17 || text[10] != 'T' {
		return "", false
	}
	jahr, monat, tag, ok := datumTeile(text[:10])
	if !ok {
		return "", false
	}
	h, mi, s, bruch, rest, ok := uhrzeit(text[11:])
	if !ok {
		return "", false
	}
	versatz, ok := versatzMinuten(rest)
	if !ok {
		return "", false
	}
	utc := time.Date(jahr, time.Month(monat), tag, h, mi, s, 0, time.UTC).Add(-time.Duration(versatz) * time.Minute)
	if utc.Year() < 1 || utc.Year() > 9999 {
		return "", false
	}
	out := utc.Format("2006-01-02T15:04:05")
	if bruch = strings.TrimRight(bruch, "0"); bruch != "" {
		out += "." + bruch
	}
	return out + "Z", true
}
