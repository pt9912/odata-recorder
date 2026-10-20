package model

import (
	"strings"
	"unicode/utf8"
)

// maxRedaktorTiefe ist die höchste Verschachtelung, die der Redaktor liest (SPEC-118 Nr. 4).
const maxRedaktorTiefe = 512

type jsonZustand uint8

// Die Zustände des Automaten: zuerst die Struktur, dann der String, dann die Zahl.
const (
	jzWert         jsonZustand = iota // erwartet einen Wert
	jzArrayAnfang                     // nach `[`: Wert oder `]`
	jzObjektAnfang                    // nach `{`: Name oder `}`
	jzName                            // nach `,` im Objekt: Name
	jzNachName                        // nach dem Namen: `:`
	jzNachWert                        // nach einem Wert in einem Container
	jzEnde                            // der Wurzelwert ist vollständig
	jzString                          // im Inhalt eines Strings
	jzEscape                          // nach `\`
	jzHex                             // in den vier Hex-Ziffern von `\u`
	jzMinus                           // nach `-`
	jzNull                            // nach einer führenden `0`
	jzInt                             // in den Ziffern vor `.` und `e`
	jzPunkt                           // nach `.`
	jzBruch                           // in den Nachkommastellen
	jzE                               // nach `e`
	jzEVz                             // nach dem Vorzeichen des Exponenten
	jzExp                             // in den Ziffern des Exponenten
	jzLiteral                         // in `true`, `false` oder `null`
)

type stringArt uint8

const (
	saWert stringArt = iota
	saName
	saErsetzt
)

const (
	escapeEin = "\"\\/bfnrt"
	escapeAus = "\"\\/\b\f\n\r\t"
	// ersatz ist die Ausgabe für einen ersetzten String.
	ersatz = `"` + Marker + `"`
)

// jsonRedaktor liest den Body Byte für Byte (SPEC-118 Nr. 4 und 5). Sein Speicher hängt an der
// Tiefe, am längsten Muster und an den Puffern der Ausgabe, nicht an der Länge des Bodys.
type jsonRedaktor struct {
	pfade   []JSONPfad
	maxName int
	aus     *bodyAusgabe

	z      jsonZustand
	stapel []byte
	stumm  bool // das aktuelle Byte gehört zu einem ersetzten String und wird nicht geschrieben

	art        stringArt
	name       []byte // der dekodierte Name, höchstens maxName Bytes
	ueberlang  bool   // der Name ist länger als das längste Muster und kann nicht treffen
	nameTrifft bool
	trifft     bool // der nächste Wert gehört zu einem treffenden Namen

	utf8Rest uint8 // fehlende Folgebytes einer Mehrbyte-Folge
	lo, hi   byte  // zulässiger Bereich des nächsten Folgebytes
	hexN     uint8
	cp       rune

	lit    string
	litPos int

	mPos      int // so viele Rohbytes des Inhalts gleichen dem Anfang von Marker
	mAbw      bool
	geaendert bool
}

func neuerJSONRedaktor(pfade []JSONPfad, aus *bodyAusgabe) *jsonRedaktor {
	maxName := 0
	for _, p := range pfade {
		maxName = max(maxName, len(p.name))
	}
	return &jsonRedaktor{pfade: pfade, maxName: maxName, aus: aus, name: make([]byte, 0, maxName)}
}

func (s *jsonRedaktor) schreibe(p []byte) *UnsicherFehler {
	for _, c := range p {
		if g := s.step(c); g != "" {
			return unsicher(g, nil)
		}
		if f := s.aus.schreibeVoll(); f != nil {
			return f
		}
	}
	return nil
}

func (s *jsonRedaktor) status() RedactionStatus {
	if s.geaendert {
		return RedactionApplied
	}
	return RedactionNone
}

// ende urteilt über das Ende des Stroms: allein ein vollständiger Wurzelwert ist gültig, auch
// eine Zahl, die am Ende steht.
func (s *jsonRedaktor) ende() *UnsicherFehler {
	if len(s.stapel) == 0 {
		switch s.z {
		case jzEnde, jzNull, jzInt, jzBruch, jzExp:
			return nil
		}
	}
	return unsicher(GrundUnvollstaendig, nil)
}

func (s *jsonRedaktor) step(c byte) UnsicherGrund {
	s.stumm = false
	g := s.verarbeite(c)
	if g == "" && !s.stumm {
		s.aus.byte(c)
	}
	return g
}

func (s *jsonRedaktor) verarbeite(c byte) UnsicherGrund {
	switch s.z {
	case jzString, jzEscape, jzHex:
		return s.zeichen(c)
	case jzMinus, jzNull, jzInt, jzPunkt, jzBruch, jzE, jzEVz, jzExp:
		return s.zahl(c)
	case jzLiteral:
		return s.literal(c)
	}
	return s.struktur(c)
}

func (s *jsonRedaktor) struktur(c byte) UnsicherGrund {
	if istLeerraum(c) {
		return ""
	}
	switch s.z {
	case jzWert:
		return s.wertStart(c)
	case jzArrayAnfang:
		if c == ']' {
			return s.schliesse()
		}
		return s.wertStart(c)
	case jzObjektAnfang:
		if c == '}' {
			return s.schliesse()
		}
		return s.nameStart(c)
	case jzName:
		return s.nameStart(c)
	case jzNachName:
		if c != ':' {
			return GrundSyntax
		}
		s.trifft = s.nameTrifft
		s.z = jzWert
		return ""
	case jzNachWert:
		return s.nachWert(c)
	}
	return GrundSyntax // nach dem Wurzelwert
}

func (s *jsonRedaktor) nachWert(c byte) UnsicherGrund {
	objekt := s.stapel[len(s.stapel)-1] == '{'
	switch {
	case c == ',' && objekt:
		s.z = jzName
	case c == ',':
		s.z = jzWert
	case c == '}' && objekt, c == ']' && !objekt:
		return s.schliesse()
	default:
		return GrundSyntax
	}
	return ""
}

func (s *jsonRedaktor) schliesse() UnsicherGrund {
	s.stapel = s.stapel[:len(s.stapel)-1]
	s.wertFertig()
	return ""
}

func (s *jsonRedaktor) wertFertig() {
	if len(s.stapel) == 0 {
		s.z = jzEnde
		return
	}
	s.z = jzNachWert
}

func (s *jsonRedaktor) oeffne(art byte, danach jsonZustand) UnsicherGrund {
	if len(s.stapel) >= maxRedaktorTiefe {
		return GrundTiefe
	}
	s.stapel = append(s.stapel, art)
	s.z = danach
	return ""
}

// wertStart beginnt einen Wert. Unter einem treffenden Namen ist allein ein String, `null`, `true`
// oder `false` zulässig; Zahl, Objekt und Array sind nichtstring, entschieden am ersten Byte.
func (s *jsonRedaktor) wertStart(c byte) UnsicherGrund {
	trifft := s.trifft
	s.trifft = false
	if trifft && strings.IndexByte("{[-0123456789", c) >= 0 {
		return GrundNichtstring
	}
	switch c {
	case '{':
		return s.oeffne('{', jzObjektAnfang)
	case '[':
		return s.oeffne('[', jzArrayAnfang)
	case '"':
		s.beginneString(saWert)
		if trifft {
			s.art, s.mPos, s.mAbw, s.stumm = saErsetzt, 0, false, true
		}
		return ""
	}
	return s.skalarStart(c)
}

func (s *jsonRedaktor) skalarStart(c byte) UnsicherGrund {
	switch {
	case c == '-':
		s.z = jzMinus
	case c == '0':
		s.z = jzNull
	case c >= '1' && c <= '9':
		s.z = jzInt
	case c == 't':
		s.beginneLiteral("true")
	case c == 'f':
		s.beginneLiteral("false")
	case c == 'n':
		s.beginneLiteral("null")
	default:
		return GrundSyntax
	}
	return ""
}

func (s *jsonRedaktor) beginneLiteral(w string) {
	s.z, s.lit, s.litPos = jzLiteral, w, 1
}

func (s *jsonRedaktor) literal(c byte) UnsicherGrund {
	if c != s.lit[s.litPos] {
		return GrundSyntax
	}
	s.litPos++
	if s.litPos == len(s.lit) {
		s.wertFertig()
	}
	return ""
}

func (s *jsonRedaktor) nameStart(c byte) UnsicherGrund {
	if c != '"' {
		return GrundSyntax
	}
	s.beginneString(saName)
	s.name, s.ueberlang = s.name[:0], false
	return ""
}

func (s *jsonRedaktor) beginneString(art stringArt) {
	s.z, s.art, s.utf8Rest = jzString, art, 0
}

// zeichen liest ein Byte im Inhalt eines Strings. Ein ersetzter String schreibt nichts; sein
// Inhalt wird geprüft und nur auf Gleichheit mit dem Marker verglichen.
func (s *jsonRedaktor) zeichen(c byte) UnsicherGrund {
	if s.art == saErsetzt {
		s.stumm = true
		if s.z != jzString || c != '"' {
			s.markerByte(c)
		}
	}
	switch s.z {
	case jzEscape:
		return s.escape(c)
	case jzHex:
		return s.hex(c)
	}
	return s.stringByte(c)
}

func (s *jsonRedaktor) markerByte(c byte) {
	if !s.mAbw && s.mPos < len(Marker) && c == Marker[s.mPos] {
		s.mPos++
		return
	}
	s.mAbw = true
}

func (s *jsonRedaktor) stringByte(c byte) UnsicherGrund {
	switch {
	case s.utf8Rest > 0:
		return s.folgebyte(c)
	case c == '"':
		s.stringEnde()
	case c == '\\':
		s.z = jzEscape
	case c < 0x20:
		return GrundSyntax
	case c >= 0x80:
		return s.mehrbyteStart(c)
	default:
		s.nameByte(c)
	}
	return ""
}

// mehrbyteStart prüft das erste Byte einer UTF-8-Folge nach den Regeln von Go: kein überlanger
// Code, kein Surrogat, höchstens U+10FFFF.
func (s *jsonRedaktor) mehrbyteStart(c byte) UnsicherGrund {
	s.lo, s.hi = 0x80, 0xBF
	switch {
	case c >= 0xC2 && c <= 0xDF:
		s.utf8Rest = 1
	case c >= 0xE0 && c <= 0xEF:
		s.utf8Rest = 2
		switch c {
		case 0xE0:
			s.lo = 0xA0
		case 0xED:
			s.hi = 0x9F
		}
	case c >= 0xF0 && c <= 0xF4:
		s.utf8Rest = 3
		switch c {
		case 0xF0:
			s.lo = 0x90
		case 0xF4:
			s.hi = 0x8F
		}
	default:
		return GrundUTF8
	}
	s.nameByte(c)
	return ""
}

func (s *jsonRedaktor) folgebyte(c byte) UnsicherGrund {
	if c < s.lo || c > s.hi {
		return GrundUTF8
	}
	s.utf8Rest--
	s.lo, s.hi = 0x80, 0xBF
	s.nameByte(c)
	return ""
}

func (s *jsonRedaktor) escape(c byte) UnsicherGrund {
	if c == 'u' {
		s.z, s.hexN, s.cp = jzHex, 0, 0
		return ""
	}
	i := strings.IndexByte(escapeEin, c)
	if i < 0 {
		return GrundSyntax
	}
	s.z = jzString
	s.nameByte(escapeAus[i])
	return ""
}

func (s *jsonRedaktor) hex(c byte) UnsicherGrund {
	v := hexWert(c)
	if v < 0 {
		return GrundSyntax
	}
	s.cp = s.cp<<4 | rune(v)
	s.hexN++
	if s.hexN == 4 {
		s.z = jzString
		var b [utf8.UTFMax]byte
		for _, x := range b[:utf8.EncodeRune(b[:], s.cp)] {
			s.nameByte(x)
		}
	}
	return ""
}

// nameByte hängt ein dekodiertes Byte an den Namen, solange er nicht länger ist als das längste
// Muster; ein längerer Name trifft nie und wird nicht weiter gehalten.
func (s *jsonRedaktor) nameByte(b byte) {
	if s.art != saName || s.ueberlang {
		return
	}
	if len(s.name) >= s.maxName {
		s.ueberlang = true
		return
	}
	s.name = append(s.name, b)
}

func (s *jsonRedaktor) stringEnde() {
	switch s.art {
	case saName:
		s.nameTrifft = s.passt()
		s.z = jzNachName
	case saErsetzt:
		s.aus.text(ersatz)
		if s.mAbw || s.mPos != len(Marker) {
			s.geaendert = true
		}
		s.wertFertig()
	default:
		s.wertFertig()
	}
}

func (s *jsonRedaktor) passt() bool {
	if s.ueberlang {
		return false
	}
	n := string(s.name)
	for _, p := range s.pfade {
		if p.Passt(n) {
			return true
		}
	}
	return false
}

func (s *jsonRedaktor) zahl(c byte) UnsicherGrund {
	switch s.z {
	case jzMinus, jzPunkt, jzE, jzEVz:
		return s.zahlPflicht(c)
	}
	return s.zahlZiffern(c)
}

// zahlPflicht liest das Byte, das nach `-`, `.`, `e` oder dem Vorzeichen des Exponenten
// folgen muss; endet der Strom hier, ist die Zahl unvollständig.
func (s *jsonRedaktor) zahlPflicht(c byte) UnsicherGrund {
	ziffer := c >= '0' && c <= '9'
	switch {
	case s.z == jzMinus && c == '0':
		s.z = jzNull
	case s.z == jzMinus && ziffer:
		s.z = jzInt
	case s.z == jzPunkt && ziffer:
		s.z = jzBruch
	case s.z == jzE && (c == '+' || c == '-'):
		s.z = jzEVz
	case (s.z == jzE || s.z == jzEVz) && ziffer:
		s.z = jzExp
	default:
		return GrundSyntax
	}
	return ""
}

// zahlZiffern liest ein Byte in einer Ziffernfolge: eine weitere Ziffer, der Übergang zu
// Bruch oder Exponent oder das Ende der Zahl. Auf eine führende `0` folgt keine Ziffer.
func (s *jsonRedaktor) zahlZiffern(c byte) UnsicherGrund {
	if c >= '0' && c <= '9' {
		if s.z == jzNull {
			return GrundSyntax
		}
		return ""
	}
	switch s.z {
	case jzBruch:
		return s.ziffernEnde(c, false)
	case jzExp:
		return s.zahlEnde(c)
	}
	return s.ziffernEnde(c, true)
}

// ziffernEnde setzt eine Zahl nach ihren Ziffern fort: `.` nur vor den Nachkommastellen, `e`
// vor dem Exponenten, sonst endet die Zahl.
func (s *jsonRedaktor) ziffernEnde(c byte, punkt bool) UnsicherGrund {
	switch {
	case c == '.' && punkt:
		s.z = jzPunkt
	case c == 'e' || c == 'E':
		s.z = jzE
	default:
		return s.zahlEnde(c)
	}
	return ""
}

// zahlEnde schließt die Zahl und liest das Byte, das sie beendet hat, als Byte nach einem Wert.
func (s *jsonRedaktor) zahlEnde(c byte) UnsicherGrund {
	s.wertFertig()
	return s.verarbeite(c)
}
