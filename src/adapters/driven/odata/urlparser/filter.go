package urlparser

import (
	"strings"
	"unicode"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// maxFilterTiefe ist die Zahl der offenen `(` und `not`, deren Operand noch nicht gelesen ist (SPEC-164 Nr. 7).
const maxFilterTiefe = 32

// FilterGrammatik liest `$filter` nach SPEC-164 und setzt FilterLeser um. Sie hat keinen Zustand: der Nullwert ist benutzbar,
// und gleichzeitige Aufrufe beeinflussen sich nicht.
type FilterGrammatik struct{}

// NewFilterLeser liefert den Leser der Grammatik von `$filter` (SPEC-164).
func NewFilterLeser() *FilterGrammatik { return &FilterGrammatik{} }

// LeseFilter überführt den dekodierten Wert von `$filter` in den Filter-AST (SPEC-165). Der erste Verstoß in der Lesefolge
// entscheidet und ist ein *model.URLFehler ohne Ort (SPEC-164 Nr. 6); es gibt kein Teilergebnis.
func (*FilterGrammatik) LeseFilter(text string) (interface{ IstFilterAusdruck() }, error) {
	if !gueltigerText(text) {
		return nil, fehler(model.URLKodierung, "")
	}
	if text == "" || text[0] == ' ' {
		return nil, fehler(model.URLWert, "")
	}
	p := &filterLauf{z: zerleger{s: text}}
	ausdruck, f := p.stufe(stufeOder)
	if f != nil {
		return nil, f
	}
	return ausdruck, nil
}

// filterLauf ist der Zustand eines Aufrufs; er lebt auf dem Stapel des Aufrufs.
type filterLauf struct {
	z        zerleger
	tiefe    int // offene `(` und `not` ohne gelesenen Operand
	klammern int // offene `(`
}

// Die Stufen der Rangfolge von schwach nach stark (SPEC-164 Nr. 3).
const (
	stufeOder = iota
	stufeUnd
	stufeGleich
	stufeRelation
	stufeUnaer
)

// operatorStufe liefert die Stufe eines Operatorworts in Kleinschreibung.
func operatorStufe(op string) int {
	switch op {
	case "or":
		return stufeOder
	case "and":
		return stufeUnd
	case "eq", "ne":
		return stufeGleich
	}
	return stufeRelation
}

// stufe liest eine Kette der Operatoren dieser Stufe, linksassoziativ.
func (p *filterLauf) stufe(s int) (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	if s == stufeUnaer {
		return p.unaer()
	}
	links, f := p.stufe(s + 1)
	for f == nil {
		op, f2 := p.operator()
		if f2 != nil {
			return nil, f2
		}
		if op == "" || operatorStufe(op) != s {
			return links, nil
		}
		p.z.weiter()
		var rechts interface{ IstFilterAusdruck() }
		if rechts, f = p.stufe(s + 1); f == nil {
			links = knoten(op, links, rechts)
		}
	}
	return nil, f
}

func knoten(op string, links, rechts interface{ IstFilterAusdruck() }) interface{ IstFilterAusdruck() } {
	switch op {
	case "and":
		return model.FilterVerknuepfung{Op: model.VerknuepfungUnd, Links: links, Rechts: rechts}
	case "or":
		return model.FilterVerknuepfung{Op: model.VerknuepfungOder, Links: links, Rechts: rechts}
	}
	return model.FilterVergleich{Op: model.VergleichOp(op), Links: links, Rechts: rechts}
}

// operator beurteilt das nächste Token am Platz eines Operators (SPEC-164 Nr. 4): ein Operatorwort liefert seine
// Kleinschreibung, das Ende des Operanden (Textende, passendes `)`) den leeren Text; alles andere ist ein Verstoß.
func (p *filterLauf) operator() (string, *model.URLFehler) {
	t := p.z.peek()
	switch t.art {
	case tokEnde:
		return "", nil
	case tokKlammerZu:
		if p.klammern > 0 {
			return "", nil
		}
	case tokLauf:
		switch t.klein {
		case "eq", "ne", "gt", "ge", "lt", "le", "and", "or":
			return t.klein, nil
		case "has", "in", "add", "sub", "mul", "div", "divby", "mod":
			return "", fehler(model.URLFilter, "")
		}
	}
	return "", fehler(model.URLWert, "")
}

// betrete zählt eine offene `(` oder ein `not`; die 33. ist `zu_viele`.
func (p *filterLauf) betrete() *model.URLFehler {
	p.tiefe++
	if p.tiefe > maxFilterTiefe {
		return fehler(model.URLZuViele, "")
	}
	return nil
}

func (p *filterLauf) unaer() (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	t := p.z.peek()
	if t.art != tokLauf || t.klein != "not" {
		return p.primaer()
	}
	p.z.weiter()
	if f := p.betrete(); f != nil {
		return nil, f
	}
	operand, f := p.unaer()
	p.tiefe--
	if f != nil {
		return nil, f
	}
	return model.FilterNicht{Operand: operand}, nil
}

func (p *filterLauf) primaer() (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	t := p.z.peek()
	switch t.art {
	case tokKlammerAuf:
		return p.gruppe()
	case tokZeichenkette:
		p.z.weiter()
		return zeichenketteOperand(t)
	case tokLauf:
		p.z.weiter()
		return laufOperand(t)
	}
	return nil, fehler(model.URLWert, "")
}

// gruppe liest `(` oder `)`; die Klammer ist kein Knoten (SPEC-164 Nr. 3).
func (p *filterLauf) gruppe() (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	p.z.weiter()
	if f := p.betrete(); f != nil {
		return nil, f
	}
	p.klammern++
	inneres, f := p.stufe(stufeOder)
	if f != nil {
		return nil, f
	}
	if p.z.peek().art != tokKlammerZu {
		return nil, fehler(model.URLWert, "")
	}
	p.z.weiter()
	p.klammern--
	p.tiefe--
	return inneres, nil
}

// zeichenketteOperand liest `'…'` mit oder ohne Präfix. Die Länge der Zeichenkette (bis zum schließenden `'`, ohne ein solches
// bis zum Textende) geht der Form vor, der Rest dahinter zählt nicht dazu (SPEC-164 Nr. 4 a und Nr. 6; Fänger `TestLaengeVorForm`).
func zeichenketteOperand(t token) (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	switch {
	case len(t.text) > maxLiteral:
		return nil, fehler(model.URLZuLang, "")
	case !t.geschlossen || !t.endeOK:
		return nil, fehler(model.URLLiteral, "")
	}
	return literalKnoten(t.text)
}

func literalKnoten(s string) (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	lit, u := leseLiteral(s)
	if u != "" {
		return nil, fehler(u, "")
	}
	return model.FilterLiteral{Literal: lit}, nil
}

// laufOperand beurteilt einen Lauf am Platz eines Operanden (SPEC-164 Nr. 4 b bis e).
func laufOperand(t token) (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	s := t.text
	switch c := s[0]; {
	case literalLauf(s):
		return literalKnoten(s)
	case c == '-':
		return nil, fehler(model.URLFilter, "")
	case c == '_' || c == '$' || c == '@' || unicode.IsLetter(firstRune(s)):
		return wortLauf(t)
	}
	return nil, fehler(model.URLWert, "")
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

// literalLauf meldet einen Lauf, der mit einer ASCII-Ziffer, `+` oder `.` beginnt, mit `-` und dann Ziffer oder `.`, gleich
// `-INF` ist oder genau die Form einer Guid hat, auch mit einem Hex-Buchstaben am Anfang (SPEC-164 Nr. 4 b).
func literalLauf(s string) bool {
	if _, ok := guid(s); ok || s == "-INF" {
		return true
	}
	c := s[0]
	if c == '-' {
		return len(s) > 1 && ziffernAnfang(s[1])
	}
	return ziffernAnfang(c) || c == '+'
}

func ziffernAnfang(c byte) bool { return c >= '0' && c <= '9' || c == '.' }

// wortLauf beurteilt einen Lauf, der mit einem Buchstaben, `_`, `$` oder `@` beginnt, in der Folge von SPEC-164 Nr. 4 c. Der
// Schritt (4), ein Zeichen außer Buchstabe, Ziffer, `_`, `.` und `/` ist `wert`, liegt in `einfacherName` je Glied und je Name.
func wortLauf(t token) (interface{ IstFilterAusdruck() }, *model.URLFehler) {
	s := t.text
	switch {
	case t.vorKlammer:
		return nil, fehler(model.URLFilter, "")
	case s[0] == '$' || s[0] == '@':
		return nil, sonderName(s)
	case strings.IndexByte(s, '/') >= 0:
		return nil, pfad(s)
	case strings.IndexByte(s, '.') >= 0:
		return nil, qualifiziert(s)
	}
	switch asciiKlein(s) {
	case "true", "false":
		return literalKnoten(s)
	case "null":
		return model.FilterNull{}, nil
	}
	if s == "INF" || s == "NaN" {
		return nil, fehler(model.URLLiteralTyp, "")
	}
	if !einfacherName(s) {
		return nil, fehler(model.URLWert, "")
	}
	return model.FilterEigenschaft{Name: s}, nil
}

// sonderName beurteilt `$…` und `@…`: `$it`, `$root`, `$this` (allein oder mit `/` und Rest) und `@` mit einem einfachen
// Namen sind nicht unterstützt, alles andere ungültig.
func sonderName(s string) *model.URLFehler {
	if s[0] == '$' {
		kopf, _, _ := strings.Cut(s, "/")
		if kopf == "$it" || kopf == "$root" || kopf == "$this" {
			return fehler(model.URLFilter, "")
		}
		return fehler(model.URLWert, "")
	}
	if einfacherName(s[1:]) {
		return fehler(model.URLFilter, "")
	}
	return fehler(model.URLWert, "")
}

// alleGlieder meldet, ob jedes durch sep getrennte Glied von s ein einfacher Name ist.
func alleGlieder(s string, sep byte) bool {
	for {
		i := strings.IndexByte(s, sep)
		if i < 0 {
			return einfacherName(s)
		}
		if !einfacherName(s[:i]) {
			return false
		}
		s = s[i+1:]
	}
}

// pfad beurteilt einen Lauf mit `/`: ein Pfad ist nicht unterstützt, ein fehlerhafter ungültig (SPEC-164 Nr. 4 c 5).
func pfad(s string) *model.URLFehler {
	if !alleGlieder(s, '/') {
		return fehler(model.URLWert, "")
	}
	return fehler(model.URLWertNichtUnterstuetzt, "")
}

// qualifiziert beurteilt einen Lauf mit `.`: ein qualifizierter Name ist nicht unterstützt (SPEC-164 Nr. 4 c 6).
func qualifiziert(s string) *model.URLFehler {
	if !alleGlieder(s, '.') {
		return fehler(model.URLWert, "")
	}
	return fehler(model.URLFilter, "")
}
