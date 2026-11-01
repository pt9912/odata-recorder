package urlparserv2

import (
	"strings"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// litFolge ist der Ausgang beim Lesen eines Schlüsselliterals.
type litFolge int

const (
	litGelesen litFolge = iota
	litZuLang
	litFalsch // `literal`
	litTyp    // `literal_typ`, ein Kompatibilitätsfehler
)

// leseLiteral liest ein Schlüsselliteral nach SPEC-173 Nr. 3: zuerst die Länge, dann die Formen im Umfang, dann die Literale
// außerhalb (Kompatibilität `literal_typ`); jede andere Form ist `literal`.
func leseLiteral(s string) (model.Literal, litFolge) {
	if len(s) > maxLiteral {
		return model.Literal{}, litZuLang
	}
	for _, f := range []func(string) (model.Literal, bool){zeichenkette, wahrheitswert, zahl, guid, zeitpunkt} {
		if lit, ok := f(s); ok {
			return lit, litGelesen
		}
	}
	if ausserhalb(s) {
		return model.Literal{}, litTyp
	}
	return model.Literal{}, litFalsch
}

// textInAnfuehrung meldet, ob s mit `'` beginnt, mit dem schließenden `'` endet und dazwischen jedes `'` verdoppelt ist.
func textInAnfuehrung(s string) bool {
	if len(s) < 2 || s[0] != '\'' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] != '\'' {
			continue
		}
		if i+1 < len(s) && s[i+1] == '\'' {
			i++
			continue
		}
		return i == len(s)-1
	}
	return false
}

// zeichenkette liest `'…'`; der Text ist die Literalform selbst.
func zeichenkette(s string) (model.Literal, bool) {
	if !textInAnfuehrung(s) {
		return model.Literal{}, false
	}
	return model.Literal{Art: model.LiteralZeichenkette, Text: s}, true
}

// wahrheitswert liest `true` und `false` ohne Beachtung der ASCII-Schreibweise.
func wahrheitswert(s string) (model.Literal, bool) {
	k := asciiKlein(s)
	if k != "true" && k != "false" {
		return model.Literal{}, false
	}
	return model.Literal{Art: model.LiteralWahrheitswert, Text: k}, true
}

func nurZiffern(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// dezimal liest `-`? Ziffern (`.` Ziffern)? und bildet die Dezimalschreibweise von SPEC-161 Nr. 6 c.
func dezimal(s string) (string, bool) {
	t, negativ := strings.CutPrefix(s, "-")
	ganz, bruch, hatBruch := strings.Cut(t, ".")
	if !nurZiffern(ganz) || (hatBruch && !nurZiffern(bruch)) {
		return "", false
	}
	text := strings.TrimLeft(ganz, "0")
	if text == "" {
		text = "0"
	}
	if bruch = strings.TrimRight(bruch, "0"); bruch != "" {
		text += "." + bruch
	}
	if negativ && text != "0" {
		text = "-" + text
	}
	return text, true
}

// zahl liest eine Zahl ohne Suffix, mit `L` oder `l` (nur ganz) oder mit `M` oder `m` (SPEC-173 Nr. 3).
func zahl(s string) (model.Literal, bool) {
	rumpf, ganz := s, false
	if n := len(s); n > 0 {
		switch s[n-1] {
		case 'L', 'l':
			rumpf, ganz = s[:n-1], true
		case 'M', 'm':
			rumpf = s[:n-1]
		}
	}
	if ganz && strings.Contains(rumpf, ".") {
		return model.Literal{}, false
	}
	text, ok := dezimal(rumpf)
	if !ok {
		return model.Literal{}, false
	}
	return model.Literal{Art: model.LiteralZahl, Text: text}, true
}

// praefixLiteral zerlegt `Präfix'…'` in das Präfix in ASCII-Kleinschreibung und den Inhalt zwischen den Anführungszeichen.
func praefixLiteral(s string) (praefix, inhalt string, ok bool) {
	i := strings.IndexByte(s, '\'')
	if i < 1 || !textInAnfuehrung(s[i:]) {
		return "", "", false
	}
	return asciiKlein(s[:i]), s[i+1 : len(s)-1], true
}

// guid liest `guid'…'` mit 8-4-4-4-12 Hex-Ziffern, getrennt durch `-`, und schreibt sie klein.
func guid(s string) (model.Literal, bool) {
	p, inhalt, ok := praefixLiteral(s)
	if !ok || p != "guid" || len(inhalt) != 36 {
		return model.Literal{}, false
	}
	for i := range len(inhalt) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if inhalt[i] != '-' {
				return model.Literal{}, false
			}
		} else if hexWert(inhalt[i]) < 0 {
			return model.Literal{}, false
		}
	}
	return model.Literal{Art: model.LiteralGUID, Text: asciiKlein(inhalt)}, true
}

// ausserhalb erkennt die gültigen Literale außerhalb des Umfangs (SPEC-173 Nr. 3): ein Präfix vor `'…'` außer `guid` und
// `datetime`, eine Zahl mit Exponent und eine Zahl mit Suffix `d`, `D`, `f` oder `F`.
func ausserhalb(s string) bool {
	if i := strings.IndexByte(s, '\''); i > 0 {
		p := asciiKlein(s[:i])
		return p != "guid" && p != "datetime" && typPraefix(s[:i]) && textInAnfuehrung(s[i:])
	}
	return zahlMitExponent(s) || zahlMitGleitSuffix(s)
}

// typPraefix meldet einen Namen, `.`, Ziffern, Buchstaben und `_` vor dem `'` (`binary`, `X`, `Ns.Typ`).
func typPraefix(s string) bool {
	for _, r := range s {
		if r != '.' && r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// zahlMitExponent erkennt `-`? Ziffern (`.` Ziffern)? (`e`|`E`) (`+`|`-`)? Ziffern.
func zahlMitExponent(s string) bool {
	mantisse, exp, hat := strings.Cut(asciiKlein(s), "e")
	if !hat {
		return false
	}
	if exp != "" && (exp[0] == '+' || exp[0] == '-') {
		exp = exp[1:]
	}
	_, ok := dezimal(mantisse)
	return ok && nurZiffern(exp)
}

// zahlMitGleitSuffix erkennt `-`? Ziffern (`.` Ziffern)? mit dem Suffix `d`, `D`, `f` oder `F`.
func zahlMitGleitSuffix(s string) bool {
	n := len(s)
	if n < 2 {
		return false
	}
	switch s[n-1] {
	case 'd', 'D', 'f', 'F':
		_, ok := dezimal(s[:n-1])
		return ok
	}
	return false
}
