package urlparser

import "strings"

// tokenArt ist die Art eines Tokens von `$filter` (SPEC-164 Nr. 2).
type tokenArt int

const (
	tokEnde         tokenArt = iota // Textende
	tokRand                         // Leerzeichen am Ende des Textes
	tokKlammerAuf                   // `(`
	tokKlammerZu                    // `)`
	tokKomma                        // `,`
	tokZeichenkette                 // `'…'`, mit oder ohne Präfix-Lauf davor
	tokLauf                         // längste Folge von Zeichen außer U+0020, `(`, `)`, `,` und `'`
)

// token ist ein Token mit seinem Text. Ein Lauf trägt, ob unmittelbar `(` folgt, und, bis fünf Bytes lang, seine
// Kleinschreibung (die längsten Schlüsselwörter sind `divby` und `false`); eine Zeichenkette, ob sie geschlossen ist und ob
// hinter ihr ein Ende folgt (SPEC-164 Nr. 4 a).
type token struct {
	art         tokenArt
	text        string
	klein       string
	vorKlammer  bool
	geschlossen bool
	endeOK      bool
}

// zerleger schneidet den Text von links nach rechts in Token; er liest jedes Byte in einem Durchlauf und geht nie zurück.
type zerleger struct {
	s    string
	pos  int
	tok  token
	ende int
	hat  bool
}

const maxSchluesselwort = 5

func laufEnde(c byte) bool {
	return c == ' ' || c == '(' || c == ')' || c == ',' || c == '\''
}

// peek liefert das nächste Token, ohne es zu verbrauchen.
func (z *zerleger) peek() token {
	if !z.hat {
		z.tok, z.ende = z.schneide()
		z.hat = true
	}
	return z.tok
}

// weiter verbraucht das Token, das peek liefert.
func (z *zerleger) weiter() {
	z.peek()
	z.pos, z.hat = z.ende, false
}

func (z *zerleger) schneide() (token, int) {
	i := z.pos
	for i < len(z.s) && z.s[i] == ' ' {
		i++
	}
	if i == len(z.s) {
		if i > z.pos {
			return token{art: tokRand}, i
		}
		return token{art: tokEnde}, i
	}
	switch z.s[i] {
	case '(':
		return token{art: tokKlammerAuf}, i + 1
	case ')':
		return token{art: tokKlammerZu}, i + 1
	case ',':
		return token{art: tokKomma}, i + 1
	}
	j := i
	for j < len(z.s) && !laufEnde(z.s[j]) {
		j++
	}
	if j < len(z.s) && z.s[j] == '\'' {
		return z.zeichenkette(i, j)
	}
	t := token{art: tokLauf, text: z.s[i:j], vorKlammer: j < len(z.s) && z.s[j] == '('}
	if len(t.text) <= maxSchluesselwort {
		t.klein = asciiKlein(t.text)
	}
	return t, j
}

// zeichenkette schneidet `'…'` ab dem Präfix bei i; das erste `'` steht bei j. Sie endet am ersten `'`, auf das kein zweites
// folgt; ohne ein solches reicht sie bis zum Textende.
func (z *zerleger) zeichenkette(i, j int) (token, int) {
	k := j + 1
	for k < len(z.s) {
		n := strings.IndexByte(z.s[k:], '\'')
		if n < 0 {
			break
		}
		k += n + 1
		if k < len(z.s) && z.s[k] == '\'' {
			k++
			continue
		}
		t := token{art: tokZeichenkette, text: z.s[i:k], geschlossen: true}
		t.endeOK = k == len(z.s) || z.s[k] == ' ' || z.s[k] == ')' || z.s[k] == ','
		return t, k
	}
	return token{art: tokZeichenkette, text: z.s[i:]}, len(z.s)
}
