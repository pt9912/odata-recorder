package baselineimport

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strconv"
)

// roh ist ein Wert, der schon in der Kanonform steht und unverändert eingesetzt wird.
type roh []byte

type objekt = map[string]any

const hexKlein = "0123456789abcdef"

// kanon schreibt v in der Kanonform von SPEC-127 Nr. 14: ohne Leerraum, Mitglieder eines Objekts
// aufsteigend nach den Bytes ihrer Namen, Strings mit `\"`, `\\` und `\u00XX` (kleine Hexziffern)
// für jedes Byte unter 0x20 und sonst unverändert, Zahlen als Ganzzahlen, eine leere Liste als `[]`.
// Erlaubt sind nil, bool, int, int64, string, roh, []any und objekt.
func kanon(v any) []byte {
	var b bytes.Buffer
	schreibe(&b, v)
	return b.Bytes()
}

func schreibe(b *bytes.Buffer, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case int:
		b.WriteString(strconv.Itoa(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case string:
		schreibeText(b, x)
	case roh:
		b.Write(x)
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			schreibe(b, e)
		}
		b.WriteByte(']')
	case objekt:
		namen := make([]string, 0, len(x))
		for n := range x {
			namen = append(namen, n)
		}
		slices.Sort(namen)
		b.WriteByte('{')
		for i, n := range namen {
			if i > 0 {
				b.WriteByte(',')
			}
			schreibeText(b, n)
			b.WriteByte(':')
			schreibe(b, x[n])
		}
		b.WriteByte('}')
	default:
		panic("baselineimport: kanon: Typ ohne Kanonform")
	}
}

func schreibeText(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c < 0x20:
			b.WriteString(`\u00`)
			b.WriteByte(hexKlein[c>>4])
			b.WriteByte(hexKlein[c&15])
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
}

// liste bildet die Liste aus xs; die Liste ist nie nil.
func liste[T any](xs []T, f func(T) any) []any {
	out := make([]any, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

// sha256Hex ist "sha256:" und der SHA-256 von b in 64 kleinen Hexziffern.
func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(s[:])
}
