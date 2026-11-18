package urlparser

import (
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Klasse eines Elements von `$expand` (SPEC-262 Nr. 4).
const (
	elUngueltig = iota
	elName
	elStern
	elPfad
)

// leseExpand liest `$expand` (SPEC-262 Nr. 3): ein `(` im Wert ist nicht unterstützt und beendet das Lesen des Werts, dann
// die Zahl der Elemente, dann die Elemente von links nach rechts; ein Verstoß der Klasse `ungueltig` beendet das Lesen, einer
// der Klasse `nicht_unterstuetzt` nicht.
func (l *lauf) leseExpand(wert, ort string, o *model.Optionen) error {
	if wert == "" {
		return fehler(model.URLWert, ort)
	}
	if strings.Contains(wert, "(") {
		l.nichtUnterstuetzt(model.URLWertNichtUnterstuetzt, ort)
		return nil
	}
	if strings.Count(wert, ",")+1 > maxExpand {
		return fehler(model.URLZuViele, ort)
	}
	teile := strings.Split(wert, ",")
	namen := make([]model.ExpandElement, 0, len(teile))
	for i, e := range teile {
		if slices.Contains(teile[:i], e) {
			return fehler(model.URLWert, ort)
		}
		switch elementKlasse(e) {
		case elName:
			namen = append(namen, model.ExpandElement{Name: e})
		case elStern, elPfad:
			l.nichtUnterstuetzt(model.URLWertNichtUnterstuetzt, ort)
		default:
			return fehler(model.URLWert, ort)
		}
	}
	o.Expand = &model.ExpandOption{Elemente: namen}
	return nil
}

// elementKlasse ordnet ein Element zu: Name, genau `*`, Pfad (SPEC-262 Nr. 4 a bis c) oder jede andere Form (Nr. 4 d).
func elementKlasse(e string) int {
	switch {
	case einfacherName(e):
		return elName
	case e == "*":
		return elStern
	case expandPfad(e):
		return elPfad
	}
	return elUngueltig
}

// expandPfad meldet einen Pfad: mindestens zwei Segmente, jedes ein Name, ein qualifizierter Name oder `$ref`, `$count`,
// `$value` (SPEC-262 Nr. 4 c).
func expandPfad(e string) bool {
	segmente := strings.Split(e, "/")
	if len(segmente) < 2 {
		return false
	}
	for _, s := range segmente {
		if !expandSegment(s) {
			return false
		}
	}
	return true
}

func expandSegment(s string) bool {
	switch s {
	case "$ref", "$count", "$value":
		return true
	}
	for _, n := range strings.Split(s, ".") {
		if !einfacherName(n) {
			return false
		}
	}
	return true
}
