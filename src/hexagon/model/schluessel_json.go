package model

import (
	"bytes"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// maxJSONTiefe ist die höchste Verschachtelung, die json-canonical kanonisiert (SPEC-116 Nr. 6).
const maxJSONTiefe = 512

// kanonischesJSON liefert die Kanonform, wenn b genau ein gültiger JSON-Wert ist (SPEC-116
// Nr. 6); sonst false.
func kanonischesJSON(b []byte) ([]byte, bool) {
	p := &jsonLeser{in: b}
	p.leerraum()
	var out bytes.Buffer
	if !p.wert(&out, 0) {
		return nil, false
	}
	p.leerraum()
	if p.pos != len(b) {
		return nil, false
	}
	return out.Bytes(), true
}

type jsonLeser struct {
	in  []byte
	pos int
}

func (p *jsonLeser) leerraum() {
	for p.pos < len(p.in) && strings.IndexByte(" \t\r\n", p.in[p.pos]) >= 0 {
		p.pos++
	}
}

func (p *jsonLeser) wort(w string) bool {
	if !bytes.HasPrefix(p.in[p.pos:], []byte(w)) {
		return false
	}
	p.pos += len(w)
	return true
}

func (p *jsonLeser) wert(out *bytes.Buffer, tiefe int) bool {
	if p.pos >= len(p.in) {
		return false
	}
	switch c := p.in[p.pos]; {
	case c == '{':
		return p.objekt(out, tiefe+1)
	case c == '[':
		return p.array(out, tiefe+1)
	case c == '"':
		s, ok := p.text()
		if ok {
			schreibeText(out, s)
		}
		return ok
	case c == '-' || c >= '0' && c <= '9':
		return p.zahl(out)
	}
	for _, w := range []string{"true", "false", "null"} {
		if p.wort(w) {
			out.WriteString(w)
			return true
		}
	}
	return false
}

func (p *jsonLeser) array(out *bytes.Buffer, tiefe int) bool {
	if tiefe > maxJSONTiefe {
		return false
	}
	p.pos++
	out.WriteByte('[')
	p.leerraum()
	if p.pos < len(p.in) && p.in[p.pos] == ']' {
		p.pos++
		out.WriteByte(']')
		return true
	}
	for first := true; ; first = false {
		if !first {
			out.WriteByte(',')
		}
		p.leerraum()
		if !p.wert(out, tiefe) {
			return false
		}
		p.leerraum()
		if p.pos >= len(p.in) {
			return false
		}
		c := p.in[p.pos]
		p.pos++
		if c == ']' {
			out.WriteByte(']')
			return true
		}
		if c != ',' {
			return false
		}
	}
}

type jsonMitglied struct {
	name string
	wert []byte
}

func (p *jsonLeser) objekt(out *bytes.Buffer, tiefe int) bool {
	if tiefe > maxJSONTiefe {
		return false
	}
	p.pos++
	p.leerraum()
	if p.nimm('}') {
		out.WriteString("{}")
		return true
	}
	var ms []jsonMitglied
	for {
		m, ok := p.mitglied(tiefe)
		if !ok {
			return false
		}
		ms = append(ms, m)
		p.leerraum()
		if p.nimm('}') {
			break
		}
		if !p.nimm(',') {
			return false
		}
	}
	slices.SortStableFunc(ms, func(a, b jsonMitglied) int { return strings.Compare(a.name, b.name) })
	out.WriteByte('{')
	for i, m := range ms {
		if i > 0 {
			out.WriteByte(',')
		}
		schreibeText(out, m.name)
		out.WriteByte(':')
		out.Write(m.wert)
	}
	out.WriteByte('}')
	return true
}

// nimm verbraucht das Byte c, wenn es das nächste ist.
func (p *jsonLeser) nimm(c byte) bool {
	if p.pos < len(p.in) && p.in[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *jsonLeser) mitglied(tiefe int) (jsonMitglied, bool) {
	p.leerraum()
	if p.pos >= len(p.in) || p.in[p.pos] != '"' {
		return jsonMitglied{}, false
	}
	name, ok := p.text()
	if !ok {
		return jsonMitglied{}, false
	}
	p.leerraum()
	if !p.nimm(':') {
		return jsonMitglied{}, false
	}
	p.leerraum()
	var wert bytes.Buffer
	if !p.wert(&wert, tiefe) {
		return jsonMitglied{}, false
	}
	return jsonMitglied{name, wert.Bytes()}, true
}

func (p *jsonLeser) ziffern() int {
	n := 0
	for p.pos < len(p.in) && p.in[p.pos] >= '0' && p.in[p.pos] <= '9' {
		p.pos++
		n++
	}
	return n
}

// zahl übernimmt das Literal nach der Grammatik von RFC 8259 unverändert.
func (p *jsonLeser) zahl(out *bytes.Buffer) bool {
	start := p.pos
	p.nimm('-')
	if !p.nimm('0') && p.ziffern() == 0 {
		return false
	}
	if p.nimm('.') && p.ziffern() == 0 {
		return false
	}
	if p.nimm('e') || p.nimm('E') {
		_ = p.nimm('+') || p.nimm('-')
		if p.ziffern() == 0 {
			return false
		}
	}
	out.Write(p.in[start:p.pos])
	return true
}

// text liest einen String ab dem öffnenden Anführungszeichen und liefert den dekodierten Wert;
// ungültiges UTF-8 und einzelne Surrogat-Escapes sind ungültig.
func (p *jsonLeser) text() (string, bool) {
	p.pos++
	var sb strings.Builder
	for p.pos < len(p.in) {
		c := p.in[p.pos]
		switch {
		case c == '"':
			p.pos++
			return sb.String(), true
		case c < 0x20:
			return "", false
		case c == '\\':
			if !p.escape(&sb) {
				return "", false
			}
		default:
			r, n := utf8.DecodeRune(p.in[p.pos:])
			if r == utf8.RuneError && n <= 1 {
				return "", false
			}
			sb.WriteRune(r)
			p.pos += n
		}
	}
	return "", false
}

func (p *jsonLeser) hex4(at int) (rune, bool) {
	if at+4 > len(p.in) {
		return 0, false
	}
	var r rune
	for _, c := range p.in[at : at+4] {
		h := hexWert(c)
		if h < 0 {
			return 0, false
		}
		r = r<<4 | rune(h)
	}
	return r, true
}

func (p *jsonLeser) escape(sb *strings.Builder) bool {
	if p.pos+1 >= len(p.in) {
		return false
	}
	c := p.in[p.pos+1]
	p.pos += 2
	if r, ok := einfacherEscape(c); ok {
		sb.WriteByte(r)
		return true
	}
	if c != 'u' {
		return false
	}
	r, ok := p.hex4(p.pos)
	if !ok {
		return false
	}
	p.pos += 4
	switch {
	case utf16.IsSurrogate(r) && r >= 0xDC00:
		return false
	case utf16.IsSurrogate(r):
		if p.pos+2 > len(p.in) || p.in[p.pos] != '\\' || p.in[p.pos+1] != 'u' {
			return false
		}
		lo, ok := p.hex4(p.pos + 2)
		if !ok || lo < 0xDC00 || lo > 0xDFFF {
			return false
		}
		p.pos += 6
		r = utf16.DecodeRune(r, lo)
	}
	sb.WriteRune(r)
	return true
}

const hexKlein = "0123456789abcdef"

func schreibeText(out *bytes.Buffer, s string) {
	out.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"' || c == '\\':
			out.WriteByte('\\')
			out.WriteByte(c)
		case c < 0x20:
			out.WriteString(`\u00`)
			out.WriteByte(hexKlein[c>>4])
			out.WriteByte(hexKlein[c&15])
		default:
			out.WriteByte(c)
		}
	}
	out.WriteByte('"')
}

func einfacherEscape(c byte) (byte, bool) {
	switch c {
	case '"', '\\', '/':
		return c, true
	case 'b':
		return '\b', true
	case 'f':
		return '\f', true
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	}
	return 0, false
}
