package urlparser

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die sieben behandelten Systemoptionen (SPEC-162 Nr. 3 a), als Stelle im Feld der gesehenen.
const (
	optSelect = iota
	optOrderBy
	optTop
	optSkip
	optCount
	optFilter
	optFormat
	optKeine = -1
)

func behandelt(name string) int {
	switch name {
	case "$select":
		return optSelect
	case "$orderby":
		return optOrderBy
	case "$top":
		return optTop
	case "$skip":
		return optSkip
	case "$count":
		return optCount
	case "$filter":
		return optFilter
	case "$format":
		return optFormat
	}
	return optKeine
}

// nichtUnterstuetzteOption nennt die Namen, die OData 4.01 kennt und dieser Mock nicht liest (Nr. 3 b).
func nichtUnterstuetzteOption(name string) bool {
	switch name {
	case "$expand", "$search", "$apply", "$compute", "$levels", "$schemaversion", "$deltatoken", "$skiptoken", "$index", "$id":
		return true
	}
	return false
}

// sonstigerName erkennt einen Parameter-Alias und eine herstellerspezifische Option (Nr. 3 d, e).
func sonstigerName(name string) bool {
	if alias, ok := strings.CutPrefix(name, "@"); ok {
		return einfacherName(alias)
	}
	if name == "" || len(name) > maxName {
		return false
	}
	for i := range len(name) {
		c := name[i]
		buchstabe := c == '_' || (c|0x20) >= 'a' && (c|0x20) <= 'z'
		if !buchstabe && (i == 0 || (c < '0' || c > '9') && c != '-' && c != '.') {
			return false
		}
	}
	return true
}

// leseQuery liest die Optionen von links nach rechts (SPEC-162 Nr. 2 und 3). Die Folge je Option: Kodierung des Namens,
// Klasse, Wiederholung, Kodierung des Werts, Grammatik des Werts.
func (l *lauf) leseQuery(segs []model.QuerySegment, o *model.Optionen) error {
	var gesehen [optFormat + 1]bool
	for i, seg := range segs {
		ort := ortNr("query", i+1)
		name, ok := dekodiere(seg.Name)
		if !ok {
			return fehler(model.URLKodierung, ort)
		}
		idx := behandelt(name)
		switch {
		case idx != optKeine && gesehen[idx]:
			return fehler(model.URLOptionDoppelt, ort)
		case nichtUnterstuetzteOption(name):
			l.nichtUnterstuetzt(model.URLOptionNichtUnterstuetzt, ort)
			continue
		case idx == optKeine && !sonstigerName(name):
			return fehler(model.URLOptionName, ort)
		}
		wert, ok := dekodiere(seg.Wert)
		if !ok {
			return fehler(model.URLKodierung, ort)
		}
		if idx == optKeine {
			o.Sonstige = append(o.Sonstige, model.SonstigeOption{Name: name, Wert: wert})
			continue
		}
		gesehen[idx] = true
		if err := l.leseOption(idx, wert, ort, o); err != nil {
			return err
		}
	}
	return nil
}

// leseOption liest den Wert einer behandelten Option (Nr. 4 bis 8).
func (l *lauf) leseOption(idx int, wert, ort string, o *model.Optionen) error {
	switch idx {
	case optSelect:
		return l.leseSelect(wert, ort, o)
	case optOrderBy:
		return l.leseOrderBy(wert, ort, o)
	case optTop:
		return leseZahl(wert, ort, &o.Top)
	case optSkip:
		return leseZahl(wert, ort, &o.Skip)
	case optCount:
		k := asciiKlein(wert)
		if k != "true" && k != "false" {
			return fehler(model.URLWert, ort)
		}
		o.Count = model.WahrOption{Gesetzt: true, Wert: k == "true"}
	case optFilter:
		return l.leseFilter(wert, ort, o)
	case optFormat:
		return l.leseFormat(wert, ort)
	}
	return nil
}

// leseZahl liest `$top` und `$skip`: ASCII-Ziffern, führende Nullen erlaubt, von 0 bis 2147483647 (Nr. 4).
func leseZahl(wert, ort string, z *model.ZahlOption) error {
	t := strings.TrimLeft(wert, "0")
	if !nurZiffern(wert) || len(t) > 10 {
		return fehler(model.URLWert, ort)
	}
	v := 0
	for i := range len(t) {
		v = v*10 + int(t[i]-'0')
	}
	if v > 1<<31-1 {
		return fehler(model.URLWert, ort)
	}
	*z = model.ZahlOption{Gesetzt: true, Wert: v}
	return nil
}

// elementePfad meldet einen Pfad aus mindestens zwei einfachen Namen, durch `/` getrennt (Nr. 6).
func elementePfad(e string) bool {
	glieder := strings.Split(e, "/")
	if len(glieder) < 2 {
		return false
	}
	for _, g := range glieder {
		if !einfacherName(g) {
			return false
		}
	}
	return true
}

// leseSelect liest `$select` (Nr. 6): `*`, einfache Namen und, nicht unterstützt, Pfade.
func (l *lauf) leseSelect(wert, ort string, o *model.Optionen) error {
	elemente := strings.Split(wert, ",")
	if len(elemente) > maxSelect {
		return fehler(model.URLZuViele, ort)
	}
	sel := &model.SelectOption{}
	for _, e := range elemente {
		switch {
		case e == "*" && !sel.Alle:
			sel.Alle = true
		case einfacherName(e) && !slices.Contains(sel.Namen, e):
			sel.Namen = append(sel.Namen, e)
		case elementePfad(e):
			l.nichtUnterstuetzt(model.URLWertNichtUnterstuetzt, ort)
		default:
			return fehler(model.URLWert, ort)
		}
	}
	o.Select = sel
	return nil
}

// leseOrderBy liest `$orderby` (Nr. 7): ein `(` macht den Wert zum Ausdruck, sonst Elemente `Name` oder `Name asc|desc`.
func (l *lauf) leseOrderBy(wert, ort string, o *model.Optionen) error {
	if strings.Contains(wert, "(") {
		l.nichtUnterstuetzt(model.URLWertNichtUnterstuetzt, ort)
		return nil
	}
	elemente := strings.Split(wert, ",")
	if len(elemente) > maxOrderBy {
		return fehler(model.URLZuViele, ort)
	}
	var sort []model.Sortierung
	for _, e := range elemente {
		name, richtung, hatRichtung := strings.Cut(e, " ")
		k := asciiKlein(richtung)
		if hatRichtung && k != "asc" && k != "desc" {
			return fehler(model.URLWert, ort)
		}
		switch {
		case elementePfad(name):
			l.nichtUnterstuetzt(model.URLWertNichtUnterstuetzt, ort)
		case einfacherName(name) && !enthaeltSortierung(sort, name):
			sort = append(sort, model.Sortierung{Name: name, Absteigend: k == "desc"})
		default:
			return fehler(model.URLWert, ort)
		}
	}
	o.OrderBy = sort
	return nil
}

func enthaeltSortierung(sort []model.Sortierung, name string) bool {
	for _, s := range sort {
		if s.Name == name {
			return true
		}
	}
	return false
}

// leseFormat liest `$format` (Nr. 8): JSON ist lesbar und ohne Eintrag, jeder andere nichtleere Wert nicht unterstützt.
func (l *lauf) leseFormat(wert, ort string) error {
	k := asciiKlein(wert)
	switch {
	case wert == "":
		return fehler(model.URLWert, ort)
	case k != "json" && k != "application/json" && !strings.HasPrefix(k, "application/json;"):
		l.nichtUnterstuetzt(model.URLOptionNichtUnterstuetzt, ort)
	}
	return nil
}

// leseFilter liest `$filter` (Nr. 8 und SPEC-160 Nr. 8): ohne Leser ist es nicht unterstützt, mit Leser entscheidet er.
func (l *lauf) leseFilter(wert, ort string, o *model.Optionen) error {
	if wert == "" {
		return fehler(model.URLWert, ort)
	}
	if l.p.filter == nil {
		l.nichtUnterstuetzt(model.URLFilter, ort)
		return nil
	}
	ausdruck, err := l.p.filter.LeseFilter(wert)
	var uf *model.URLFehler
	switch {
	case errors.As(err, &uf) && uf != nil && uf.Ursache.Gueltig():
		if uf.Ursache.NichtUnterstuetzt() {
			l.nichtUnterstuetzt(uf.Ursache, ort)
			return nil
		}
		return fehler(uf.Ursache, ort)
	case err != nil:
		return fmt.Errorf("%w: urlparser: FilterLeser lieferte einen fremden Fehler", model.ErrUngueltig)
	case ausdruck == nil:
		return fmt.Errorf("%w: urlparser: FilterLeser lieferte weder Ausdruck noch Fehler", model.ErrUngueltig)
	}
	o.Filter = &model.FilterOption{Text: wert, Ausdruck: ausdruck}
	return nil
}
