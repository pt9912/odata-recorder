package urlparserv2

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// maxFilterTiefe ist die Zahl der offenen `(` und `not`, deren Operand noch nicht gelesen ist (SPEC-164 Nr. 7).
const maxFilterTiefe = 32

// FilterGrammatik liest `$filter` des V2-Umfangs (SPEC-173 Nr. 5) und setzt FilterLeser um. Sie hat keinen Zustand: der Nullwert
// ist benutzbar, und gleichzeitige Aufrufe beeinflussen sich nicht.
type FilterGrammatik struct{}

// NewFilterLeser liefert den Leser von `$filter` (SPEC-173 Nr. 5).
func NewFilterLeser() *FilterGrammatik { return &FilterGrammatik{} }

// LeseFilter überführt den dekodierten Wert von `$filter` in den Filter-AST (SPEC-165). Der erste Verstoß in der Lesefolge
// entscheidet: ein *model.URLFehler der Klasse `ungueltig` oder ein *model.KompatFehler mit der Ursache `filter` oder
// `literal_typ`, beide ohne Ort; es gibt kein Teilergebnis.
func (*FilterGrammatik) LeseFilter(text string) (interface{ IstFilterAusdruck() }, error) {
	if !gueltigerText(text) {
		return nil, fehler(model.URLKodierung, "")
	}
	if text == "" || text[0] == ' ' {
		return nil, fehler(model.URLWert, "")
	}
	p := &filterLauf{z: zerleger{s: text}}
	ausdruck, err := p.stufe(stufeOder)
	if err != nil {
		return nil, err
	}
	return ausdruck, nil
}

// kompatFehler ist der Kompatibilitätsfehler des Lesers ohne Ort.
func kompatFehler(u model.KompatUrsache) error { return &model.KompatFehler{Ursache: u} }

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
func (p *filterLauf) stufe(s int) (interface{ IstFilterAusdruck() }, error) {
	if s == stufeUnaer {
		return p.unaer()
	}
	links, err := p.stufe(s + 1)
	for err == nil {
		op, err2 := p.operator()
		if err2 != nil {
			return nil, err2
		}
		if op == "" || operatorStufe(op) != s {
			return links, nil
		}
		p.z.weiter()
		var rechts interface{ IstFilterAusdruck() }
		if rechts, err = p.stufe(s + 1); err == nil {
			links = knoten(op, links, rechts)
		}
	}
	return nil, err
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

// operator beurteilt das nächste Token am Platz eines Operators (SPEC-164 Nr. 4): ein Operatorwort des V2-Umfangs liefert seine
// Kleinschreibung, das Ende des Operanden (Textende, passendes `)`) den leeren Text; `has`, `in`, `add`, `sub`, `mul`, `div`,
// `divby` und `mod` sind die Kompatibilitätsursache `filter` (SPEC-173 Nr. 5), alles andere ist ungültig.
func (p *filterLauf) operator() (string, error) {
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
			return "", kompatFehler(model.KompatFilter)
		}
	}
	return "", fehler(model.URLWert, "")
}

// betrete zählt eine offene `(` oder ein `not`; die 33. ist `zu_viele`.
func (p *filterLauf) betrete() error {
	p.tiefe++
	if p.tiefe > maxFilterTiefe {
		return fehler(model.URLZuViele, "")
	}
	return nil
}

func (p *filterLauf) unaer() (interface{ IstFilterAusdruck() }, error) {
	t := p.z.peek()
	if t.art != tokLauf || t.klein != "not" {
		return p.primaer()
	}
	p.z.weiter()
	if err := p.betrete(); err != nil {
		return nil, err
	}
	operand, err := p.unaer()
	p.tiefe--
	if err != nil {
		return nil, err
	}
	return model.FilterNicht{Operand: operand}, nil
}

func (p *filterLauf) primaer() (interface{ IstFilterAusdruck() }, error) {
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
func (p *filterLauf) gruppe() (interface{ IstFilterAusdruck() }, error) {
	p.z.weiter()
	if err := p.betrete(); err != nil {
		return nil, err
	}
	p.klammern++
	inneres, err := p.stufe(stufeOder)
	if err != nil {
		return nil, err
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
// bis zum Textende) geht der Form vor, der Rest dahinter zählt nicht dazu (SPEC-164 Nr. 4 a und Nr. 6).
func zeichenketteOperand(t token) (interface{ IstFilterAusdruck() }, error) {
	switch {
	case len(t.text) > maxLiteral:
		return nil, fehler(model.URLZuLang, "")
	case !t.geschlossen || !t.endeOK:
		return nil, fehler(model.URLLiteral, "")
	}
	return literalKnoten(leseLiteral(t.text))
}

// literalKnoten bildet den Knoten eines Literals oder den Fehler seiner Folge (SPEC-173 Nr. 3): `literal_typ` ist eine
// Kompatibilitätsursache, die übrigen Verstöße sind ungültig.
func literalKnoten(lit model.Literal, folge litFolge) (interface{ IstFilterAusdruck() }, error) {
	switch folge {
	case litGelesen:
		return model.FilterLiteral{Literal: lit}, nil
	case litZuLang:
		return nil, fehler(model.URLZuLang, "")
	case litTyp:
		return nil, kompatFehler(model.KompatLiteralTyp)
	}
	return nil, fehler(model.URLLiteral, "")
}

// laufOperand beurteilt einen Lauf am Platz eines Operanden (SPEC-164 Nr. 4 b bis e mit SPEC-173 Nr. 5).
func laufOperand(t token) (interface{ IstFilterAusdruck() }, error) {
	s := t.text
	switch c := s[0]; {
	case literalLauf(s):
		return laufLiteral(s)
	case c == '-':
		return nil, kompatFehler(model.KompatFilter)
	case c == '_' || c == '$' || c == '@' || unicode.IsLetter(erstesZeichen(s)):
		return wortLauf(t)
	}
	return nil, fehler(model.URLWert, "")
}

func erstesZeichen(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

// literalLauf meldet einen Lauf, der mit einer ASCII-Ziffer, `+` oder `.` beginnt, mit `-` und dann Ziffer oder `.`, gleich
// `-INF` ist oder genau die Form einer Guid hat, auch mit einem Hex-Buchstaben am Anfang (SPEC-164 Nr. 4 b).
func literalLauf(s string) bool {
	if guidForm(s) || s == "-INF" {
		return true
	}
	c := s[0]
	if c == '-' {
		return len(s) > 1 && ziffernAnfang(s[1])
	}
	return ziffernAnfang(c) || c == '+'
}

func ziffernAnfang(c byte) bool { return c >= '0' && c <= '9' || c == '.' }

// guidForm meldet genau 36 Zeichen: 8, 4, 4, 4, 12 ASCII-Hex-Ziffern in jeder Schreibweise, durch `-` getrennt.
func guidForm(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := range len(s) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if s[i] != '-' {
				return false
			}
		} else if hexWert(s[i]) < 0 {
			return false
		}
	}
	return true
}

// laufLiteral liest einen Literal-Lauf ohne Präfix in der Folge von SPEC-173 Nr. 5: zuerst die Länge, dann die Zahl, dann
// `literal_typ` (Suffix `d`, `D`, `f`, `F`, Exponent, `-INF`), dann `wert` für die Formen von V4 (Guid, Datum und Zeitpunkt,
// Uhrzeit), sonst `literal`.
func laufLiteral(s string) (interface{ IstFilterAusdruck() }, error) {
	lit, folge := leseLiteral(s)
	if folge == litFalsch {
		switch {
		case s == "-INF":
			folge = litTyp
		case formVonV4(s):
			return nil, fehler(model.URLWert, "")
		}
	}
	return literalKnoten(lit, folge)
}

// formVonV4 meldet einen Lauf, der nach der Form von V4 beginnt: die Guid-Form, vier Ziffern, `-`, zwei Ziffern, `-`, zwei Ziffern
// oder zwei Ziffern, `:`, zwei Ziffern; der Rest und die Bereiche bleiben ungeprüft.
func formVonV4(s string) bool {
	ziffern := func(von, bis int) bool { return len(s) >= bis && nurZiffern(s[von:bis]) }
	switch {
	case guidForm(s):
		return true
	case ziffern(0, 4) && len(s) >= 10 && s[4] == '-' && s[7] == '-':
		return ziffern(5, 7) && ziffern(8, 10)
	}
	return ziffern(0, 2) && len(s) >= 5 && s[2] == ':' && ziffern(3, 5)
}

// wortLauf beurteilt einen Lauf, der mit einem Buchstaben, `_`, `$` oder `@` beginnt, in der Folge von SPEC-164 Nr. 4 c mit
// SPEC-173 Nr. 5: Aufruf, Pfad, qualifizierter Name, `$it` und Alias sind die Kompatibilitätsursache `filter`.
func wortLauf(t token) (interface{ IstFilterAusdruck() }, error) {
	s := t.text
	switch {
	case t.vorKlammer:
		return nil, kompatFehler(model.KompatFilter)
	case s[0] == '$' || s[0] == '@':
		return nil, sonderName(s)
	case strings.IndexByte(s, '/') >= 0:
		return nil, glieder(s, '/')
	case strings.IndexByte(s, '.') >= 0:
		return nil, glieder(s, '.')
	}
	switch asciiKlein(s) {
	case "true", "false":
		return literalKnoten(leseLiteral(s))
	case "null":
		return model.FilterNull{}, nil
	}
	if s == "INF" || s == "NaN" {
		return nil, kompatFehler(model.KompatLiteralTyp)
	}
	if !einfacherName(s) {
		return nil, fehler(model.URLWert, "")
	}
	return model.FilterEigenschaft{Name: s}, nil
}

// sonderName beurteilt `$…` und `@…`: `$it`, `$root`, `$this` (allein oder mit `/` und Rest) und `@` mit einem einfachen
// Namen sind `filter`, alles andere ist ungültig.
func sonderName(s string) error {
	if s[0] == '$' {
		kopf, _, _ := strings.Cut(s, "/")
		if kopf == "$it" || kopf == "$root" || kopf == "$this" {
			return kompatFehler(model.KompatFilter)
		}
		return fehler(model.URLWert, "")
	}
	if einfacherName(s[1:]) {
		return kompatFehler(model.KompatFilter)
	}
	return fehler(model.URLWert, "")
}

// glieder beurteilt einen Lauf mit `/` (Pfad) oder `.` (qualifizierter Name): jedes durch sep getrennte Glied ist ein einfacher
// Name und der Lauf die Kompatibilitätsursache `filter`, sonst ist er ungültig (SPEC-173 Nr. 5).
func glieder(s string, sep byte) error {
	for {
		i := strings.IndexByte(s, sep)
		if i < 0 {
			break
		}
		if !einfacherName(s[:i]) {
			return fehler(model.URLWert, "")
		}
		s = s[i+1:]
	}
	if !einfacherName(s) {
		return fehler(model.URLWert, "")
	}
	return kompatFehler(model.KompatFilter)
}
