package urlparser

import (
	"strings"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// leseLiteral liest ein Schlüsselliteral nach SPEC-161 Nr. 6. Die Ursache ist leer, wenn es gelesen ist, sonst `zu_lang`,
// `literal` oder `literal_typ`.
func leseLiteral(s string) (model.Literal, model.URLUrsache) {
	if len(s) > maxLiteral {
		return model.Literal{}, model.URLZuLang
	}
	for _, f := range []func(string) (model.Literal, bool){zeichenkette, wahrheitswert, zahl, guid, datum, zeitpunkt} {
		if lit, ok := f(s); ok {
			return lit, ""
		}
	}
	if nichtModelliert(s) {
		return model.Literal{}, model.URLLiteralTyp
	}
	return model.Literal{}, model.URLLiteral
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

// zeichenkette liest `'…'` (SPEC-161 Nr. 6 a); der Text ist die Literalform selbst.
func zeichenkette(s string) (model.Literal, bool) {
	if !textInAnfuehrung(s) {
		return model.Literal{}, false
	}
	return model.Literal{Art: model.LiteralZeichenkette, Text: s}, true
}

// wahrheitswert liest `true` und `false` ohne Beachtung der ASCII-Schreibweise (Nr. 6 b).
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

// zahl liest `-`? Ziffern (`.` Ziffern)? und bildet die Dezimalschreibweise von SPEC-126 Nr. 10 (Nr. 6 c).
func zahl(s string) (model.Literal, bool) {
	t, negativ := strings.CutPrefix(s, "-")
	ganz, bruch, hatBruch := strings.Cut(t, ".")
	if !nurZiffern(ganz) || (hatBruch && !nurZiffern(bruch)) {
		return model.Literal{}, false
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
	return model.Literal{Art: model.LiteralZahl, Text: text}, true
}

// guid liest 8-4-4-4-12 Hex-Ziffern, getrennt durch `-`, und schreibt sie klein (Nr. 6 d).
func guid(s string) (model.Literal, bool) {
	if len(s) != 36 {
		return model.Literal{}, false
	}
	for i := range len(s) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if s[i] != '-' {
				return model.Literal{}, false
			}
		} else if hexWert(s[i]) < 0 {
			return model.Literal{}, false
		}
	}
	return model.Literal{Art: model.LiteralGUID, Text: asciiKlein(s)}, true
}

// nichtModelliert erkennt die gültigen Literale eines Typs, den das Modell nicht trägt (Nr. 6 g); der Inhalt zwischen den
// Anführungszeichen wird nicht gelesen.
func nichtModelliert(s string) bool {
	switch s {
	case "INF", "-INF", "NaN":
		return true
	}
	if i := strings.IndexByte(s, '\''); i > 0 && typPraefix(s[:i]) && textInAnfuehrung(s[i:]) {
		return true
	}
	return zahlMitExponent(s) || uhrzeit(s)
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
	_, ok := zahl(mantisse)
	return ok && nurZiffern(exp)
}
