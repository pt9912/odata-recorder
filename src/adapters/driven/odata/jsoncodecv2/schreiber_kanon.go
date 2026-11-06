package jsoncodecv2

import (
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// artPruefer liefert zur Art den EDM-Typ, zu dem sie gehört (SPEC-170 Nr. 4), und die Prüfung auf die kanonische Schreibweise
// des Textes (SPEC-172 Nr. 6 k); bekannt ist falsch für `null`, `redigiert`, `date` und jede unbekannte Art.
func artPruefer(a model.WertArt) (edm string, kanon func(string) bool, bekannt bool) {
	switch a {
	case model.ArtString:
		return "Edm.String", utf8.ValidString, true
	case model.ArtBoolean:
		return "Edm.Boolean", func(s string) bool { return s == "true" || s == "false" }, true
	case model.ArtInt32:
		return "Edm.Int32", func(s string) bool { return kanonGanz(s, 32) }, true
	case model.ArtInt64:
		return "Edm.Int64", func(s string) bool { return kanonGanz(s, 64) }, true
	case model.ArtDecimal:
		return "Edm.Decimal", kanonDezimal, true
	case model.ArtGUID:
		return "Edm.Guid", kanonGUID, true
	case model.ArtDateTimeOffset:
		return "Edm.DateTimeOffset", func(s string) bool { _, ok := zeitpunktMillis(s); return ok }, true
	}
	return "", nil, false
}

// pruefeWert prüft Art und Text gegen die Property (SPEC-168 Nr. 4): `null` nur bei nullbarer Property und mit leerem Text.
func pruefeWert(p model.Property, w model.Wert) error {
	if w.Art == model.ArtNull {
		if !p.Nullable || w.Text != "" {
			return ungueltig("null nicht zulässig")
		}
		return nil
	}
	edm, kanon, bekannt := artPruefer(w.Art)
	if !bekannt || edm != p.EdmType {
		return ungueltig("Art passt nicht zum Typ der Property")
	}
	if !kanon(w.Text) {
		return ungueltig("Text ist nicht die kanonische Schreibweise der Art")
	}
	return nil
}

// ziffernVon liefert das Ende der Folge ASCII-Ziffern ab i.
func ziffernVon(s string, i int) int {
	for i < len(s) && ziffer(s[i]) {
		i++
	}
	return i
}

// ganzTeilKanon prüft `0` oder eine Ziffernfolge ohne führende Null ab i und liefert ihr Ende (i bei keiner).
func ganzTeilKanon(s string, i int) int {
	switch {
	case i >= len(s):
		return i
	case s[i] == '0':
		return i + 1
	case s[i] >= '1' && s[i] <= '9':
		return ziffernVon(s, i)
	}
	return i
}

// kanonGanz: `0`, sonst Ziffern ohne führende Null, `-` bei negativem Wert, im Bereich von bits Bit.
func kanonGanz(s string, bits int) bool {
	i := 0
	if len(s) > 0 && s[0] == '-' {
		i = 1
	}
	e := ganzTeilKanon(s, i)
	if e == i || e != len(s) || s == "-0" {
		return false
	}
	_, err := strconv.ParseInt(s, 10, bits)
	return err == nil
}

// kanonDezimal: ganzer Teil, bei Rest `.` und ein gebrochener Teil ohne Null am Ende, `-` nur bei einem Wert ungleich null,
// höchstens maxDezimalLaenge Zeichen.
func kanonDezimal(s string) bool {
	if len(s) > maxDezimalLaenge {
		return false
	}
	i := 0
	if len(s) > 0 && s[0] == '-' {
		i = 1
	}
	e := ganzTeilKanon(s, i)
	if e == i {
		return false
	}
	if e == len(s) {
		return s != "-0"
	}
	if s[e] != '.' {
		return false
	}
	f := ziffernVon(s, e+1)
	return f == len(s) && f > e+1 && s[f-1] != '0'
}

// kanonGUID: `8-4-4-4-12` Hex-Ziffern in Kleinbuchstaben.
func kanonGUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !ziffer(c) && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// zahlAus liest n ASCII-Ziffern ab s[i] als Zahl.
func zahlAus(s string, i, n int) (int, bool) {
	if ziffernVon(s[:i+n], i) != i+n {
		return 0, false
	}
	z, err := strconv.Atoi(s[i : i+n])
	return z, err == nil
}

// zeitpunktMillis liefert zur kanonischen Schreibweise `JJJJ-MM-TTThh:mm:ssZ` oder `JJJJ-MM-TTThh:mm:ss.fZ` (`f` 1 bis 3 Ziffern,
// die letzte nicht `0`; Jahr 0001 bis 9999) die Millisekunden seit 1970-01-01T00:00:00Z (SPEC-172 Nr. 6 k).
func zeitpunktMillis(s string) (int64, bool) {
	if len(s) < 20 || len(s) > 24 || len(s) == 21 || s[4] != '-' || s[7] != '-' || s[10] != 'T' || s[13] != ':' || s[16] != ':' || s[len(s)-1] != 'Z' {
		return 0, false
	}
	teile := [6]int{}
	for k, st := range [6][2]int{{0, 4}, {5, 2}, {8, 2}, {11, 2}, {14, 2}, {17, 2}} {
		z, ok := zahlAus(s, st[0], st[1])
		if !ok {
			return 0, false
		}
		teile[k] = z
	}
	jahr, monat, tag, std, mi, sek := teile[0], teile[1], teile[2], teile[3], teile[4], teile[5]
	t := time.Date(jahr, time.Month(monat), tag, 0, 0, 0, 0, time.UTC)
	if jahr < 1 || t.Year() != jahr || int(t.Month()) != monat || t.Day() != tag || std > 23 || mi > 59 || sek > 59 {
		return 0, false
	}
	ms := 0
	if bruch := s[19 : len(s)-1]; bruch != "" {
		ziffern := bruch[1:]
		if bruch[0] != '.' || ziffernVon(ziffern, 0) != len(ziffern) || ziffern[len(ziffern)-1] == '0' {
			return 0, false
		}
		ms, _ = strconv.Atoi(ziffern)
		for range 3 - len(ziffern) {
			ms *= 10
		}
	}
	return time.Date(jahr, time.Month(monat), tag, std, mi, sek, ms*int(time.Millisecond), time.UTC).UnixMilli(), true
}
