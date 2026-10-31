package urlparser

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// lesePfad liest den Ressourcenpfad (SPEC-161): `/` allein ist das Service-Dokument, sonst wird vor der Dekodierung an `/`
// zerlegt und jedes Segment genau einmal dekodiert. Ein Verstoß der Klasse `ungueltig` ist der Rückgabewert, einer der
// Klasse `nicht_unterstuetzt` merkt sich der Lauf.
func (l *lauf) lesePfad(pfad string, r *model.Ressource) error {
	if pfad == "/" {
		l.nichtUnterstuetzt(model.URLServiceDokument, "pfad")
		return nil
	}
	for i, roh := range strings.Split(pfad[1:], "/") {
		ort := ortNr("pfad", i+1)
		t, ok := dekodiere(roh)
		if !ok {
			return fehler(model.URLKodierung, ort)
		}
		var err error
		if i == 0 {
			err = l.ersteSegment(t, ort, r)
		} else {
			err = l.weiteresSegment(t, ort, i, r)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ersteSegment liest das erste Segment nach der Tabelle von SPEC-161 Nr. 3.
func (l *lauf) ersteSegment(t, ort string, r *model.Ressource) error {
	switch {
	case t == "$metadata":
		r.Art = model.RessourceMetadaten
	case t == "$batch" || t == "$entity" || t == "$all" || t == "$crossjoin":
		l.nichtUnterstuetzt(model.URLPfadNichtUnterstuetzt, ort)
	case strings.HasPrefix(t, "$"):
		return fehler(model.URLSegment, ort)
	default:
		return l.mengeOderEntity(t, ort, r)
	}
	return nil
}

// mengeOderEntity liest Name und Klammer (SPEC-161 Nr. 4).
func (l *lauf) mengeOderEntity(t, ort string, r *model.Ressource) error {
	name, rest, n := namenAnfang(t)
	if n < 1 || n > maxName {
		return fehler(model.URLSegment, ort)
	}
	switch {
	case rest == "":
		r.EntitySet, r.Art = name, model.RessourceMenge
		return nil
	case rest[0] != '(':
		return fehler(model.URLSegment, ort)
	case !strings.HasSuffix(rest, ")"):
		return fehler(model.URLKey, ort)
	case rest == "()":
		l.nichtUnterstuetzt(model.URLPfadNichtUnterstuetzt, ort)
		return nil
	}
	teile, err := l.leseKey(rest[1:len(rest)-1], ort)
	if err != nil {
		return err
	}
	r.EntitySet, r.Art, r.Schluessel = name, model.RessourceEntity, teile
	return nil
}

// weiteresSegment liest ein Segment ab dem zweiten (SPEC-161 Nr. 7): erst Kodierung und Form des Segments, dann die Frage,
// ob an dieser Stelle ein weiteres zulässig ist.
func (l *lauf) weiteresSegment(t, ort string, i int, r *model.Ressource) error {
	if !restForm(t) {
		return fehler(model.URLSegment, ort)
	}
	switch {
	case r.Art == model.RessourceMetadaten || r.Art == model.RessourceAnzahl:
		return fehler(model.URLPfadForm, ort)
	case i == 1 && t == "$count" && r.Art == model.RessourceEntity:
		return fehler(model.URLPfadForm, ort)
	case i == 1 && t == "$count" && r.Art == model.RessourceMenge:
		r.Art = model.RessourceAnzahl
	default:
		l.nichtUnterstuetzt(model.URLPfadNichtUnterstuetzt, ort)
	}
	return nil
}

// restForm prüft die Form eines Rest-Segments: ein Name, ein qualifizierter Name oder `$value`, `$ref`, `$count`, jeweils
// ohne oder mit einer Klammer, die mit `)` endet und deren Inhalt nicht gelesen wird.
func restForm(t string) bool {
	kopf, klammer, hat := strings.Cut(t, "(")
	if hat && !strings.HasSuffix(klammer, ")") {
		return false
	}
	if kopf == "$value" || kopf == "$ref" || kopf == "$count" {
		return true
	}
	for _, glied := range strings.Split(kopf, ".") {
		if !einfacherName(glied) {
			return false
		}
	}
	return true
}
