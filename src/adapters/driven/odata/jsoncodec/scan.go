package jsoncodec

import (
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// maxTiefe ist die Zahl der Objekte und Arrays, die ineinander liegen dürfen (SPEC-126 Nr. 5).
const maxTiefe = 512

// namenLinear ist die Zahl der Namen eines Objekts, bis zu der doppelte Namen linear gesucht werden.
const namenLinear = 16

type zustand uint8

// Die Zustände stehen in Gruppen, nach denen byte verteilt: Kopf und Struktur, Text, Zahl, Literal.
const (
	zStart zustand = iota
	zBom
	zWert
	zArrayAnfang
	zObjektAnfang
	zObjektName
	zDoppelpunkt
	zNachWert
	zString
	zEscape
	zHex
	zUtf8
	zMinus
	zNull
	zGanz
	zPunkt
	zBruch
	zExpE
	zExpVz
	zExp
	zLiteral
)

// rahmen ist ein offenes Objekt oder Array. Bei einem Objekt ist len(namen) gleich len(kinder), solange kein Wert
// aussteht, und um eins größer, solange der Wert des letzten Namens aussteht.
type rahmen struct {
	objekt bool
	kinder []knoten
	namen  []string
	menge  map[string]struct{}
}

// hat meldet, ob der Name im Objekt schon steht.
func (r *rahmen) hat(name string) bool {
	if r.menge != nil {
		_, ok := r.menge[name]
		return ok
	}
	for _, n := range r.namen {
		if n == name {
			return true
		}
	}
	return false
}

func (r *rahmen) fuegeNameEin(name string) {
	r.namen = append(r.namen, name)
	if r.menge != nil {
		r.menge[name] = struct{}{}
		return
	}
	if len(r.namen) > namenLinear {
		r.menge = make(map[string]struct{}, 2*len(r.namen))
		for _, n := range r.namen {
			r.menge[n] = struct{}{}
		}
	}
}

// scanner liest den Strom Byte für Byte als Zustandsmaschine und baut den Baum auf (SPEC-126 Nr. 2 bis 5).
// Er kennt das Modell nicht; der erste Verstoß in der Reihenfolge der Bytes entscheidet, gleich wie der Strom
// in Reads zerlegt ist.
type scanner struct {
	z       zustand
	stapel  []rahmen
	wurzel  knoten
	fertig  bool
	puffer  []byte
	istName bool
	escape  bool
	hoch    rune
	hexN    int
	code    rune
	rest    int
	lo, hi  byte
	lit     string
	art     art
	litPos  int
	bom     byte
	bomN    int
}

func hexWert(c byte) rune {
	switch {
	case c >= '0' && c <= '9':
		return rune(c - '0')
	case c >= 'a' && c <= 'f':
		return rune(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return rune(c-'A') + 10
	}
	return -1
}

func leerraum(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

func ziffer(c byte) bool { return c >= '0' && c <= '9' }

// fuettere verarbeitet die Bytes eines Reads; eine leere Ursache heißt: kein Verstoß.
func (s *scanner) fuettere(bytes []byte) model.DatenUrsache {
	for _, c := range bytes {
		if u := s.byte(c); u != "" {
			return u
		}
	}
	return ""
}

// ende prüft das Ende des Stroms: das Dokument muss ein vollständiger Wert sein.
func (s *scanner) ende() model.DatenUrsache {
	switch s.z {
	case zNull, zGanz, zBruch, zExp:
		s.zahlBeenden()
	}
	if s.z == zNachWert && len(s.stapel) == 0 && s.fertig {
		return ""
	}
	return model.DatenSyntax
}

// wert hängt einen fertigen Wert an das offene Objekt, Array oder die Wurzel.
func (s *scanner) wert(k knoten) {
	s.z = zNachWert
	if len(s.stapel) == 0 {
		s.wurzel, s.fertig = k, true
		return
	}
	f := &s.stapel[len(s.stapel)-1]
	f.kinder = append(f.kinder, k)
}

func (s *scanner) oeffne(objekt bool) model.DatenUrsache {
	if len(s.stapel) >= maxTiefe {
		return model.DatenTiefe
	}
	s.stapel = append(s.stapel, rahmen{objekt: objekt})
	if objekt {
		s.z = zObjektAnfang
	} else {
		s.z = zArrayAnfang
	}
	return ""
}

func (s *scanner) schliesse() {
	f := s.stapel[len(s.stapel)-1]
	s.stapel = s.stapel[:len(s.stapel)-1]
	k := knoten{art: artArray, kinder: f.kinder}
	if f.objekt {
		k = knoten{art: artObjekt, kinder: f.kinder, namen: f.namen}
	}
	s.wert(k)
}

func (s *scanner) beginneString(name bool) {
	s.puffer, s.istName, s.escape, s.hoch, s.z = s.puffer[:0], name, false, 0, zString
}

func (s *scanner) beginneWert(c byte) model.DatenUrsache {
	switch {
	case c == '{':
		return s.oeffne(true)
	case c == '[':
		return s.oeffne(false)
	case c == '"':
		s.beginneString(false)
	case c == '-':
		s.puffer, s.z = append(s.puffer[:0], c), zMinus
	case c == '0':
		s.puffer, s.z = append(s.puffer[:0], c), zNull
	case c >= '1' && c <= '9':
		s.puffer, s.z = append(s.puffer[:0], c), zGanz
	default:
		return s.beginneLiteral(c)
	}
	return ""
}

func (s *scanner) beginneLiteral(c byte) model.DatenUrsache {
	switch c {
	case 't':
		s.lit, s.art = "true", artWahr
	case 'f':
		s.lit, s.art = "false", artFalsch
	case 'n':
		s.lit, s.art = "null", artNull
	default:
		return model.DatenSyntax
	}
	s.litPos, s.z = 1, zLiteral
	return ""
}

func (s *scanner) zahlBeenden() {
	s.wert(knoten{art: artZahl, text: string(s.puffer)})
}

func (s *scanner) endeString() model.DatenUrsache {
	text := string(s.puffer)
	if !s.istName {
		s.wert(knoten{art: artString, text: text, marker: !s.escape && text == "[REDACTED]"})
		return ""
	}
	f := &s.stapel[len(s.stapel)-1]
	if f.hat(text) {
		return model.DatenDoppelterName
	}
	f.fuegeNameEin(text)
	s.z = zDoppelpunkt
	return ""
}

// einheit verarbeitet eine fertige `\u`-Einheit: ein Surrogatpaar muss aus einem hohen und unmittelbar einem
// tiefen bestehen.
func (s *scanner) einheit(code rune) model.DatenUrsache {
	switch {
	case s.hoch != 0:
		if code < 0xDC00 || code > 0xDFFF {
			return model.DatenKodierung
		}
		s.puffer = utf8.AppendRune(s.puffer, 0x10000+(s.hoch-0xD800)<<10+(code-0xDC00))
		s.hoch = 0
	case code >= 0xD800 && code <= 0xDBFF:
		s.hoch = code
	case code >= 0xDC00 && code <= 0xDFFF:
		return model.DatenKodierung
	default:
		s.puffer = utf8.AppendRune(s.puffer, code)
	}
	s.z = zString
	return ""
}

// folge setzt die Erwartung für die Bytes hinter dem ersten Byte eines mehrbyteigen Zeichens (Regeln von Go:
// kein überlanger Code, kein Surrogat, höchstens U+10FFFF).
func (s *scanner) folge(c byte) model.DatenUrsache {
	switch {
	case c >= 0xC2 && c <= 0xDF:
		s.rest, s.lo, s.hi = 1, 0x80, 0xBF
	case c == 0xE0:
		s.rest, s.lo, s.hi = 2, 0xA0, 0xBF
	case c == 0xED:
		s.rest, s.lo, s.hi = 2, 0x80, 0x9F
	case c >= 0xE1 && c <= 0xEF:
		s.rest, s.lo, s.hi = 2, 0x80, 0xBF
	case c == 0xF0:
		s.rest, s.lo, s.hi = 3, 0x90, 0xBF
	case c >= 0xF1 && c <= 0xF3:
		s.rest, s.lo, s.hi = 3, 0x80, 0xBF
	case c == 0xF4:
		s.rest, s.lo, s.hi = 3, 0x80, 0x8F
	default:
		return model.DatenKodierung
	}
	s.puffer = append(s.puffer, c)
	s.z = zUtf8
	return ""
}

// byte verarbeitet ein Byte; ein Byte, das eine Zahl beendet oder den Kopf abschließt, wird danach im folgenden
// Zustand gelesen.
func (s *scanner) byte(c byte) model.DatenUrsache {
	for {
		var u model.DatenUrsache
		var weiter bool
		switch {
		case s.z <= zNachWert:
			u, weiter = s.struktur(c)
		case s.z <= zUtf8:
			u = s.text(c)
		case s.z <= zExp:
			u, weiter = s.zahlByte(c)
		default:
			u = s.literalByte(c)
		}
		if !weiter {
			return u
		}
	}
}

func (s *scanner) struktur(c byte) (model.DatenUrsache, bool) {
	switch s.z {
	case zStart, zBom:
		return s.kopf(c)
	case zWert, zArrayAnfang:
		return s.wertByte(c), false
	case zNachWert:
		return s.nachWert(c), false
	}
	return s.objektByte(c), false
}

// kopf erkennt ein BOM als erste Bytes des Stroms (SPEC-126 Nr. 3); jedes andere Byte beginnt den Wert.
func (s *scanner) kopf(c byte) (model.DatenUrsache, bool) {
	if s.z == zStart {
		if c != 0xEF && c != 0xFF && c != 0xFE {
			s.z = zWert
			return "", true
		}
		s.z, s.bom, s.bomN = zBom, c, 1
		return "", false
	}
	switch {
	case s.bom == 0xEF && s.bomN == 1 && c == 0xBB:
		s.bomN = 2
		return "", false
	case s.bom == 0xEF && s.bomN == 2 && c == 0xBF, s.bom == 0xFF && c == 0xFE, s.bom == 0xFE && c == 0xFF:
		return model.DatenKodierung, false
	}
	return model.DatenSyntax, false
}

func (s *scanner) wertByte(c byte) model.DatenUrsache {
	switch {
	case leerraum(c):
		return ""
	case c == ']' && s.z == zArrayAnfang:
		s.schliesse()
		return ""
	}
	return s.beginneWert(c)
}

func (s *scanner) objektByte(c byte) model.DatenUrsache {
	switch {
	case leerraum(c):
	case s.z == zDoppelpunkt && c == ':':
		s.z = zWert
	case s.z != zDoppelpunkt && c == '"':
		s.beginneString(true)
	case s.z == zObjektAnfang && c == '}':
		s.schliesse()
	default:
		return model.DatenSyntax
	}
	return ""
}

func (s *scanner) nachWert(c byte) model.DatenUrsache {
	switch {
	case leerraum(c):
		return ""
	case len(s.stapel) == 0:
		return model.DatenSyntax
	case c == ',':
		if s.stapel[len(s.stapel)-1].objekt {
			s.z = zObjektName
		} else {
			s.z = zWert
		}
		return ""
	case c == '}' || c == ']':
		if s.stapel[len(s.stapel)-1].objekt != (c == '}') {
			return model.DatenSyntax
		}
		s.schliesse()
		return ""
	}
	return model.DatenSyntax
}

func (s *scanner) text(c byte) model.DatenUrsache {
	switch s.z {
	case zString:
		return s.imString(c)
	case zEscape:
		return s.escapeByte(c)
	case zHex:
		return s.hexByte(c)
	}
	if c < s.lo || c > s.hi {
		return model.DatenKodierung
	}
	s.puffer = append(s.puffer, c)
	s.rest, s.lo, s.hi = s.rest-1, 0x80, 0xBF
	if s.rest == 0 {
		s.z = zString
	}
	return ""
}

func (s *scanner) imString(c byte) model.DatenUrsache {
	if s.hoch != 0 && c != '\\' {
		return model.DatenKodierung
	}
	switch {
	case c == '"':
		return s.endeString()
	case c == '\\':
		s.escape, s.z = true, zEscape
	case c < 0x20:
		return model.DatenSyntax
	case c < utf8.RuneSelf:
		s.puffer = append(s.puffer, c)
	default:
		return s.folge(c)
	}
	return ""
}

func (s *scanner) escapeByte(c byte) model.DatenUrsache {
	if s.hoch != 0 && c != 'u' {
		return model.DatenKodierung
	}
	switch c {
	case '"', '\\', '/':
		s.puffer = append(s.puffer, c)
	case 'b':
		s.puffer = append(s.puffer, '\b')
	case 'f':
		s.puffer = append(s.puffer, '\f')
	case 'n':
		s.puffer = append(s.puffer, '\n')
	case 'r':
		s.puffer = append(s.puffer, '\r')
	case 't':
		s.puffer = append(s.puffer, '\t')
	case 'u':
		s.z, s.hexN, s.code = zHex, 0, 0
		return ""
	default:
		return model.DatenSyntax
	}
	s.z = zString
	return ""
}

func (s *scanner) hexByte(c byte) model.DatenUrsache {
	v := hexWert(c)
	if v < 0 {
		return model.DatenSyntax
	}
	s.code, s.hexN = s.code<<4|v, s.hexN+1
	if s.hexN < 4 {
		return ""
	}
	return s.einheit(s.code)
}

func (s *scanner) literalByte(c byte) model.DatenUrsache {
	if c != s.lit[s.litPos] {
		return model.DatenSyntax
	}
	s.litPos++
	if s.litPos == len(s.lit) {
		s.wert(knoten{art: s.art})
	}
	return ""
}

// zahlByte liest ein Byte einer Zahl nach RFC 8259 §6. weiter ist wahr, wenn das Byte die Zahl beendet hat und
// im Zustand hinter dem Wert noch zu lesen ist.
func (s *scanner) zahlByte(c byte) (model.DatenUrsache, bool) {
	switch s.z {
	case zNull, zGanz, zBruch, zExp:
		return s.zahlFortsetzung(c)
	}
	return s.zahlPflicht(c), false
}

// zahlPflicht liest das Byte, das in den Zuständen hinter `-`, `.`, `e` und dem Vorzeichen des Exponenten folgen muss.
func (s *scanner) zahlPflicht(c byte) model.DatenUrsache {
	var next zustand
	switch {
	case s.z == zMinus && c == '0':
		next = zNull
	case s.z == zMinus && c >= '1' && c <= '9':
		next = zGanz
	case s.z == zPunkt && ziffer(c):
		next = zBruch
	case (s.z == zExpE || s.z == zExpVz) && ziffer(c):
		next = zExp
	case s.z == zExpE && (c == '+' || c == '-'):
		next = zExpVz
	default:
		return model.DatenSyntax
	}
	s.puffer, s.z = append(s.puffer, c), next
	return ""
}

// zahlFortsetzung liest ein Byte in den Zuständen, in denen eine Zahl enden darf; ein anderes Byte beendet sie.
func (s *scanner) zahlFortsetzung(c byte) (model.DatenUrsache, bool) {
	next := s.z
	switch {
	case ziffer(c) && s.z != zNull:
	case c == '.' && (s.z == zNull || s.z == zGanz):
		next = zPunkt
	case (c == 'e' || c == 'E') && s.z != zExp:
		next = zExpE
	default:
		s.zahlBeenden()
		return "", true
	}
	s.puffer, s.z = append(s.puffer, c), next
	return "", false
}
