// Package begrenzt enthält den Puffer, in den der Replay-Dienst den redigierten Request-Body schreibt: er nimmt
// höchstens eine feste Zahl von Bytes an (SPEC-135 Nr. 7 c).
package begrenzt

import (
	"bytes"
	"errors"
)

// ErrGrenze meldet, dass die Ausgabe die Grenze überschreitet.
var ErrGrenze = errors.New("begrenzt: Ausgabe über der Grenze")

// Puffer sammelt Bytes bis zu einer Grenze; er ist nicht für gleichzeitigen Gebrauch.
type Puffer struct {
	buf bytes.Buffer
	max int64
}

// Neu liefert einen Puffer, der höchstens grenze Bytes annimmt.
func Neu(grenze int64) *Puffer { return &Puffer{max: grenze} }

// Write nimmt von p so viele Bytes an, wie die Grenze zulässt; bleibt ein Rest, liefert es ErrGrenze.
func (b *Puffer) Write(p []byte) (int, error) {
	frei := b.max - int64(b.buf.Len())
	if int64(len(p)) <= frei {
		return b.buf.Write(p)
	}
	n, _ := b.buf.Write(p[:frei])
	return n, ErrGrenze
}

// Bytes liefert die angenommenen Bytes.
func (b *Puffer) Bytes() []byte {
	return b.buf.Bytes()
}
