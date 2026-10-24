package csdl

import (
	"bytes"
	"io"
	"math"
	"unicode/utf8"
)

type stromArt int

const (
	stromKodierung stromArt = iota + 1
	stromSyntax
	stromZuGross
	stromKontext
	stromQuelle
)

// stromFehler trägt das Ende des Stroms vor dem Ende des Dokuments durch den Decoder hindurch.
type stromFehler struct {
	art stromArt
	err error
}

// Error erfüllt error; der Fehler verlässt den Leser nie, tokenFehler macht aus ihm den Fehler des Aufrufs.
func (*stromFehler) Error() string { return "csdl: Strom" }

const bomUTF8 = "\xEF\xBB\xBF"

// strom liest src für den Decoder: höchstens max+1 Bytes, ein BOM übergangen, jedes Byte als
// XML-1.0-Zeichen in UTF-8 geprüft, vor jedem Read der Kontext. Ein Fehler wird erst geliefert,
// nachdem die davor liegenden Bytes ausgeliefert sind; so entscheidet die Reihenfolge der Bytes.
type strom struct {
	kontext    func() error
	src        io.Reader
	max        int64
	gelesen    int64
	kopfFertig bool
	halde      []byte
	alles      []byte
	ende       error
}

func xmlZeichenGueltig(b []byte) (v int, ungueltig bool) {
	for v < len(b) {
		c := b[v]
		if c < utf8.RuneSelf {
			if c < 0x20 && c != '\t' && c != '\n' && c != '\r' {
				return v, true
			}
			v++
			continue
		}
		if !utf8.FullRune(b[v:]) {
			return v, false
		}
		r, size := utf8.DecodeRune(b[v:])
		if (r == utf8.RuneError && size == 1) || r == 0xFFFE || r == 0xFFFF {
			return v, true
		}
		v += size
	}
	return v, false
}

// kopf entscheidet über BOM und UTF-16-Marke; entschieden ist false, solange mehr Bytes fehlen.
func kopf(b []byte, aus bool) (rest []byte, entschieden bool, err error) {
	if len(b) == 0 {
		return b, aus, nil
	}
	switch b[0] {
	case 0xFF, 0xFE:
		if len(b) < 2 && !aus {
			return b, false, nil
		}
		if len(b) >= 2 && b[0]^b[1] == 0x01 && b[1] >= 0xFE {
			return b, true, &stromFehler{art: stromKodierung}
		}
		return b, true, nil
	case 0xEF:
		if len(b) < 3 && !aus {
			return b, false, nil
		}
		if bytes.HasPrefix(b, []byte(bomUTF8)) {
			return b[3:], true, nil
		}
	}
	return b, true, nil
}

func (s *strom) Read(p []byte) (int, error) {
	if s.ende != nil {
		return 0, s.ende
	}
	if err := s.kontext(); err != nil {
		s.ende = &stromFehler{art: stromKontext, err: err}
		return 0, s.ende
	}
	k := len(s.halde)
	copy(p, s.halde)
	s.halde = s.halde[:0]
	want := len(p) - k
	if rest := s.max - s.gelesen; rest < math.MaxInt64 && int64(want) > rest+1 {
		want = int(rest + 1)
	}
	n, err := s.src.Read(p[k : k+want])
	s.gelesen += int64(n)
	b := p[:k+n]
	ueber := s.gelesen > s.max
	if ueber {
		b = b[:len(b)-1]
	}
	return s.liefere(p, b, ueber, err)
}

// liefere prüft die Bytes b (ein Präfix von p) und gibt den gültigen Teil zurück; den Rest und das
// Ende des Stroms merkt es sich.
func (s *strom) liefere(p, b []byte, ueber bool, err error) (int, error) {
	aus := ueber || err != nil
	if !s.kopfFertig {
		rest, fertig, kerr := kopf(b, aus)
		if kerr != nil {
			s.ende = kerr
			return 0, s.ende
		}
		if !fertig {
			s.halde = append(s.halde[:0], b...)
			return 0, nil
		}
		s.kopfFertig = true
		b = p[:copy(p, rest)]
	}
	v, ungueltig := xmlZeichenGueltig(b)
	tail := b[v:]
	switch {
	case ungueltig:
		s.ende = &stromFehler{art: stromSyntax}
	case ueber:
		s.ende = &stromFehler{art: stromZuGross}
	case err == io.EOF && len(tail) > 0:
		s.ende = &stromFehler{art: stromSyntax}
	case err == io.EOF:
		s.ende = io.EOF
	case err != nil:
		s.ende = &stromFehler{art: stromQuelle, err: err}
	default:
		s.halde = append(s.halde[:0], tail...)
	}
	s.alles = append(s.alles, p[:v]...)
	if v == 0 && s.ende != nil {
		return 0, s.ende
	}
	return v, nil
}
