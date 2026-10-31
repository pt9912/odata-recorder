package urlparser

import (
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// rohTeil ist ein Teil des Körpers einer Klammer vor dem Lesen des Literals.
type rohTeil struct {
	name, literal string
	benannt       bool
}

// leseKey liest den dekodierten Körper einer Klammer (SPEC-161 Nr. 5 und 6) in der Folge: Zahl der Teile, Struktur,
// dann die Literale von links nach rechts.
func (l *lauf) leseKey(body, ort string) ([]model.SchluesselTeil, error) {
	roh, ok := zerlegeTeile(body)
	if !ok {
		return nil, fehler(model.URLZuViele, ort)
	}
	teile, ok := strukturiere(roh)
	if !ok {
		return nil, fehler(model.URLKey, ort)
	}
	out := make([]model.SchluesselTeil, len(teile))
	for i, t := range teile {
		lit, u := leseLiteral(t.literal)
		switch {
		case u == "":
			out[i] = model.SchluesselTeil{Name: t.name, Literal: lit}
		case u.NichtUnterstuetzt():
			l.nichtUnterstuetzt(u, ort)
		default:
			return nil, fehler(u, ort)
		}
	}
	return out, nil
}

// zerlegeTeile teilt den Körper an `,` außerhalb eines Anführungszeichen-Paares; mehr als 16 Teile sind ein Verstoß, der
// Rest wird dann nicht mehr zerlegt.
func zerlegeTeile(body string) ([]string, bool) {
	var out []string
	start, imText := 0, false
	for i := range len(body) {
		switch body[i] {
		case '\'':
			imText = !imText
		case ',':
			if imText {
				continue
			}
			if len(out)+1 == maxTeile {
				return nil, false
			}
			out = append(out, body[start:i])
			start = i + 1
		}
	}
	return append(out, body[start:]), true
}

// trenneTeil trennt Name und Literal am ersten `=` außerhalb eines Anführungszeichen-Paares.
func trenneTeil(p string) rohTeil {
	imText := false
	for i := range len(p) {
		switch p[i] {
		case '\'':
			imText = !imText
		case '=':
			if !imText {
				return rohTeil{name: p[:i], literal: p[i+1:], benannt: true}
			}
		}
	}
	return rohTeil{literal: p}
}

// nacktesWort meldet einen unbenannten Teil, der nur ein Name ist („Name ohne Literal“, SPEC-161 Nr. 5 c); die Wörter,
// die auch ein Literal sein können (`true`, `false`, `null`, `INF`, `NaN`), gehören dem Literal.
func nacktesWort(p string) bool {
	switch asciiKlein(p) {
	case "true", "false", "null":
		return false
	}
	return einfacherName(p) && p != "INF" && p != "NaN"
}

// strukturiere prüft die Struktur des Körpers (SPEC-161 Nr. 5 c): ein unbenannter Teil allein oder lauter benannte Teile
// mit Namen und Literal, ohne leeren Teil und ohne doppelten Namen.
func strukturiere(roh []string) ([]rohTeil, bool) {
	out := make([]rohTeil, 0, len(roh))
	for _, p := range roh {
		t := trenneTeil(p)
		if t.benannt && (!einfacherName(t.name) || t.literal == "") || !t.benannt && (p == "" || nacktesWort(p)) {
			return nil, false
		}
		if t.benannt && slices.ContainsFunc(out, func(o rohTeil) bool { return o.name == t.name }) {
			return nil, false
		}
		out = append(out, t)
	}
	if len(out) > 1 && slices.ContainsFunc(out, func(o rohTeil) bool { return !o.benannt }) {
		return nil, false
	}
	return out, true
}
