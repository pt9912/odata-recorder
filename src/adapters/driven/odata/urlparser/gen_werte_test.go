package urlparser_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Generatoren erzeugen den Wert (oder den AST) zuerst und schreiben daraus die Eingabe in zufälligen, aber gültigen
// Schreibweisen. Das Orakel ist nie der Parser: kanonischer Text aus math/big, time und strings, Dekodierung aus net/url.

func pick[T any](r *rand.Rand, xs ...T) T { return xs[r.IntN(len(xs))] }

func hexZiffer(r *rand.Rand, b byte) string {
	if r.IntN(2) == 0 {
		return fmt.Sprintf("%%%02X", b)
	}
	return fmt.Sprintf("%%%02x", b)
}

// kodiere schreibt s: jedes Byte, für das zwingend wahr ist, als `%XX`, jedes andere roh oder, zufällig, ebenfalls als `%XX`.
func kodiere(r *rand.Rand, s string, zwingend func(b byte) bool) string {
	var b strings.Builder
	for i := range len(s) {
		if zwingend(s[i]) || r.IntN(3) == 0 {
			b.WriteString(hexZiffer(r, s[i]))
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func pfadZwingend(b byte) bool {
	return b < 0x20 || b == 0x7f || b == '/' || b == '%' || b == '?' || b == '#'
}

func nameZwingend(b byte) bool { return b < 0x20 || b == 0x7f || b == '&' || b == '%' || b == '=' }

func wertZwingend(b byte) bool { return b < 0x20 || b == 0x7f || b == '&' || b == '%' }

func buchstaben() []rune {
	return []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_äöüßéЖΩ中")
}

// genName erzeugt einen einfachen Namen von 1 bis laenge Zeichen: Buchstabe oder `_`, dann Buchstaben, Ziffern, `_`.
func genName(r *rand.Rand, laenge int) string {
	n := 1 + r.IntN(laenge)
	rs := make([]rune, n)
	b := buchstaben()
	for i := range rs {
		rs[i] = b[r.IntN(len(b))]
		if i > 0 && r.IntN(4) == 0 {
			rs[i] = pick(r, '0', '7', '9', '٣')
		}
	}
	return string(rs)
}

// zeichenText erzeugt Text ohne U+0000 mit den Zeichen, die Grammatik und Kodierung belasten.
func zeichenText(r *rand.Rand, laenge int) string {
	alphabet := []rune("ab Z09'\",()=/%?#&+;:\\äé中\U0001F600\n\t.-_*@$")
	rs := make([]rune, r.IntN(laenge+1))
	for i := range rs {
		rs[i] = alphabet[r.IntN(len(alphabet))]
	}
	return string(rs)
}

func zufallsZeichen(r *rand.Rand) rune {
	for {
		c := rune(r.IntN(0x10000))
		if r.IntN(8) == 0 {
			c = rune(r.IntN(0x110000))
		}
		surrogat := c >= 0xd800 && c < 0xe000
		if c != 0 && c < 0x10ffff && !surrogat && c != 0xfffd {
			return c
		}
	}
}

// literalWert ist ein Wert mit seinem kanonischen Text; roh liefert bei jedem Aufruf eine neue gültige Schreibweise.
type literalWert struct {
	art  model.LiteralArt
	text string
	roh  func(r *rand.Rand) string
}

func zufallsGross(r *rand.Rand, s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' && r.IntN(2) == 0 {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}

func genLiteral(r *rand.Rand, art model.LiteralArt) literalWert {
	switch art {
	case model.LiteralWahrheitswert:
		w := pick(r, "true", "false")
		return literalWert{art, w, func(r *rand.Rand) string { return zufallsGross(r, w) }}
	case model.LiteralZahl:
		return genZahl(r)
	case model.LiteralGUID:
		b := make([]byte, 16)
		for i := range b {
			b[i] = byte(r.IntN(256))
		}
		h := hex.EncodeToString(b)
		text := h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
		return literalWert{art, text, func(r *rand.Rand) string { return zufallsGross(r, text) }}
	case model.LiteralDatum:
		t := time.Date(1+r.IntN(9999), time.Month(1+r.IntN(12)), 1+r.IntN(28), 0, 0, 0, 0, time.UTC)
		text := t.Format("2006-01-02")
		return literalWert{art, text, func(*rand.Rand) string { return text }}
	case model.LiteralZeitpunkt:
		return genZeitpunkt(r)
	}
	inhalt := zeichenText(r, 40)
	text := "'" + strings.ReplaceAll(inhalt, "'", "''") + "'"
	return literalWert{art, text, func(*rand.Rand) string { return text }}
}

func genZahl(r *rand.Rand) literalWert {
	ganz := new(big.Int)
	if r.IntN(4) > 0 {
		ganz.SetString(fmt.Sprintf("%d%0*d", r.IntN(1000000), r.IntN(30), 0), 10)
	}
	bruch := ""
	for range r.IntN(8) {
		bruch += string(rune('0' + r.IntN(10)))
	}
	bruch = strings.TrimRight(bruch, "0")
	text := ganz.String()
	if bruch != "" {
		text += "." + bruch
	}
	negativ := r.IntN(2) == 0
	if negativ && text != "0" {
		text = "-" + text
	}
	roh := func(r *rand.Rand) string {
		s := strings.Repeat("0", r.IntN(4)) + ganz.String()
		if bruch != "" || r.IntN(3) == 0 {
			s += "." + bruch + strings.Repeat("0", r.IntN(4))
			if bruch == "" && !strings.HasSuffix(s, "0") {
				s += "0"
			}
		}
		if negativ {
			s = "-" + s
		}
		return s
	}
	return literalWert{model.LiteralZahl, text, roh}
}

func genZeitpunkt(r *rand.Rand) literalWert {
	t := time.Date(2+r.IntN(9996), time.Month(1+r.IntN(12)), 1+r.IntN(28), r.IntN(24), r.IntN(60), r.IntN(60), 0, time.UTC)
	bruch := ""
	if r.IntN(2) == 0 {
		for range 1 + r.IntN(9) {
			bruch += string(rune('0' + r.IntN(10)))
		}
		bruch = strings.TrimRight(bruch, "0")
	}
	text := t.Format("2006-01-02T15:04:05")
	if bruch != "" {
		text += "." + bruch
	}
	text += "Z"
	roh := func(r *rand.Rand) string {
		versatz := 0
		if r.IntN(2) == 0 {
			versatz = r.IntN(2879) - 1439
		}
		lokal := t.Add(time.Duration(versatz) * time.Minute)
		s := lokal.Format("2006-01-02T15:04:05")
		if bruch != "" {
			s += "." + bruch + strings.Repeat("0", r.IntN(3))
		}
		switch {
		case versatz == 0 && r.IntN(2) == 0:
			return s + "Z"
		case versatz < 0:
			return fmt.Sprintf("%s-%02d:%02d", s, -versatz/60, -versatz%60)
		}
		return fmt.Sprintf("%s+%02d:%02d", s, versatz/60, versatz%60)
	}
	return literalWert{model.LiteralZeitpunkt, text, roh}
}

func alleArten() []model.LiteralArt {
	return []model.LiteralArt{model.LiteralZeichenkette, model.LiteralWahrheitswert, model.LiteralZahl, model.LiteralGUID, model.LiteralDatum, model.LiteralZeitpunkt}
}
