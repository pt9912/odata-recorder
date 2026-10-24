package jsoncodec

import (
	"strconv"
	"strings"
)

// maxDezimalLaenge ist die Zahl der Zeichen, die die kanonische Schreibweise einer Dezimalzahl haben darf
// (SPEC-126 Nr. 5).
const maxDezimalLaenge = 1024

// expoGrenze ist der Betrag, bei dem ein Exponent nicht weiter wächst: weit über jeder zulässigen Länge, weit unter
// dem Überlauf von int64 bei der Rechnung mit der Zahl der Ziffern.
const expoGrenze = 1 << 40

// zahl ist der Wert einer JSON-Zahl als Vorzeichen, Ziffern und Zehnerpotenz: der Wert ist ziffern × 10^exp,
// die Ziffern haben keine führende und keine letzte Null; der Wert Null hat keine Ziffern.
type zahl struct {
	negativ bool
	ziffern string
	exp     int64
}

func (z zahl) null() bool { return z.ziffern == "" }

// zerlege liest text als JSON-Zahl nach RFC 8259 §6; ok ist falsch, wenn der ganze Text keine ist. Die Ziffern
// werden nie ausgeschrieben: der Exponent bleibt eine Zahl.
func zerlege(text string) (z zahl, ok bool) {
	i := 0
	if strings.HasPrefix(text, "-") {
		z.negativ, i = true, 1
	}
	ganzEnde, ok := ganzTeil(text, i)
	if !ok {
		return zahl{}, false
	}
	ganz, bruch := text[i:ganzEnde], ""
	i = ganzEnde
	if i < len(text) && text[i] == '.' {
		ende := ziffernEnde(text, i+1)
		if ende == i+1 {
			return zahl{}, false
		}
		bruch, i = text[i+1:ende], ende
	}
	exp, i, ok := exponent(text, i)
	if !ok || i != len(text) {
		return zahl{}, false
	}
	m := strings.TrimLeft(ganz+bruch, "0")
	rest := strings.TrimRight(m, "0")
	z.ziffern = rest
	z.exp = exp - int64(len(bruch)) + int64(len(m)-len(rest))
	return z, true
}

func ziffernEnde(text string, i int) int {
	for i < len(text) && ziffer(text[i]) {
		i++
	}
	return i
}

// ganzTeil liefert das Ende des ganzen Teils ab i: eine Null oder eine Ziffernfolge ohne führende Null.
func ganzTeil(text string, i int) (ende int, ok bool) {
	switch {
	case i >= len(text):
		return 0, false
	case text[i] == '0':
		return i + 1, true
	case text[i] >= '1' && text[i] <= '9':
		return ziffernEnde(text, i), true
	}
	return 0, false
}

// exponent liest den optionalen Exponenten ab i; sein Betrag wächst nur bis expoGrenze.
func exponent(text string, i int) (exp int64, ende int, ok bool) {
	if i >= len(text) || (text[i] != 'e' && text[i] != 'E') {
		return 0, i, true
	}
	i++
	vorzeichen := int64(1)
	if i < len(text) && (text[i] == '+' || text[i] == '-') {
		if text[i] == '-' {
			vorzeichen = -1
		}
		i++
	}
	ende = ziffernEnde(text, i)
	if ende == i {
		return 0, i, false
	}
	for ; i < ende && exp < expoGrenze; i++ {
		exp = exp*10 + int64(text[i]-'0')
	}
	return exp * vorzeichen, ende, true
}

// dezimalText ist die kanonische Schreibweise der Zahl als Dezimalzahl (SPEC-126 Nr. 10); ok ist falsch, wenn sie
// länger als maxDezimalLaenge Zeichen wäre. Die Länge wird vor dem Schreiben berechnet.
func (z zahl) dezimalText() (text string, ok bool) {
	if z.null() {
		return "0", true
	}
	n := int64(len(z.ziffern))
	var laenge int64
	switch {
	case z.exp >= 0:
		laenge = n + z.exp
	case -z.exp < n:
		laenge = n + 1
	default:
		laenge = 2 - z.exp
	}
	if z.negativ {
		laenge++
	}
	if laenge > maxDezimalLaenge {
		return "", false
	}
	var b strings.Builder
	if z.negativ {
		b.WriteByte('-')
	}
	switch {
	case z.exp >= 0:
		b.WriteString(z.ziffern)
		b.WriteString(strings.Repeat("0", int(z.exp)))
	case -z.exp < n:
		b.WriteString(z.ziffern[:n+z.exp])
		b.WriteByte('.')
		b.WriteString(z.ziffern[n+z.exp:])
	default:
		b.WriteString("0.")
		b.WriteString(strings.Repeat("0", int(-z.exp-n)))
		b.WriteString(z.ziffern)
	}
	return b.String(), true
}

// ganzText ist der Wert als ganze Zahl im Bereich von bits Bit, in der kanonischen Schreibweise; ok ist falsch bei
// einem Bruch und bei einem Wert außerhalb des Bereichs.
func (z zahl) ganzText(bits int) (text string, ok bool) {
	if z.null() {
		return "0", true
	}
	if z.exp < 0 || int64(len(z.ziffern))+z.exp > 19 {
		return "", false
	}
	s := z.ziffern + strings.Repeat("0", int(z.exp))
	if z.negativ {
		s = "-" + s
	}
	v, err := strconv.ParseInt(s, 10, bits)
	if err != nil {
		return "", false
	}
	return strconv.FormatInt(v, 10), true
}
