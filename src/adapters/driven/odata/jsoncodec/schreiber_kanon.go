package jsoncodec

import (
	"strconv"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// artPruefer liefert zur Art den EDM-Typ, zu dem sie gehört, und die Prüfung auf die kanonische Schreibweise des Textes
// (SPEC-169 Nr. 4); bekannt ist falsch für `null`, `redigiert` und jede unbekannte Art.
func artPruefer(art model.WertArt) (edm string, kanon func(string) bool, bekannt bool) {
	switch art {
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
	case model.ArtDate:
		return "Edm.Date", func(s string) bool { _, _, _, ok := datumTeile(s); return ok }, true
	case model.ArtDateTimeOffset:
		return "Edm.DateTimeOffset", kanonZeitpunkt, true
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

// kanonZeitpunkt: `JJJJ-MM-TTThh:mm:ssZ`, mit Bruchteil `.f` (1 bis 12 Ziffern, die letzte nicht `0`) vor dem `Z`.
func kanonZeitpunkt(s string) bool {
	if len(s) < 20 || s[10] != 'T' || s[13] != ':' || s[16] != ':' || s[len(s)-1] != 'Z' {
		return false
	}
	if _, _, _, ok := datumTeile(s[:10]); !ok {
		return false
	}
	return kanonUhrzeit(s[11:19]) && kanonBruchteil(s[19:len(s)-1])
}

// kanonUhrzeit: `hh:mm:ss` mit der Stunde bis 23, der Minute und der Sekunde bis 59.
func kanonUhrzeit(t string) bool {
	h, o1 := ziffernWert(t[:2])
	mi, o2 := ziffernWert(t[3:5])
	sek, o3 := ziffernWert(t[6:8])
	return o1 && o2 && o3 && h <= 23 && mi <= 59 && sek <= 59
}

// kanonBruchteil: leer oder `.f` mit 1 bis maxBruchteil Ziffern, die letzte nicht `0`.
func kanonBruchteil(rest string) bool {
	if rest == "" {
		return true
	}
	bruch := rest[1:]
	return rest[0] == '.' && len(bruch) >= 1 && len(bruch) <= maxBruchteil && ziffernVon(bruch, 0) == len(bruch) && bruch[len(bruch)-1] != '0'
}
