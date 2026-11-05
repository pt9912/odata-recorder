package urlparserv2

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die sieben behandelten Systemoptionen (SPEC-173 Nr. 4), als Stelle im Feld der gesehenen.
const (
	optSelect = iota
	optOrderBy
	optTop
	optSkip
	optInlineCount
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
	case "$inlinecount":
		return optInlineCount
	case "$filter":
		return optFilter
	case "$format":
		return optFormat
	}
	return optKeine
}

// ausserhalbOption nennt die Optionen, die V2 nicht liest oder die V4 ohne V2-Entsprechung kennt (SPEC-173 Nr. 4): jede ist die
// Kompatibilitätsursache `option`.
func ausserhalbOption(name string) bool {
	switch name {
	case "$expand", "$skiptoken", "$callback", "$id", "$count", "$search", "$apply", "$compute", "$deltatoken", "$levels",
		"$index", "$schemaversion":
		return true
	}
	return false
}

// sonstigerName erkennt einen Parameter-Alias und eine herstellerspezifische Option (SPEC-162 Nr. 3 d, e).
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

// leseQuery liest die Optionen von links nach rechts (SPEC-162 Nr. 2 und 3 mit SPEC-173 Nr. 4). Die Folge je Option: Kodierung
// des Namens, Klasse, Wiederholung, Kodierung des Werts, Grammatik des Werts.
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
		case ausserhalbOption(name):
			l.kompat(model.KompatOption, ort)
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

// leseOption liest den Wert einer behandelten Option (SPEC-162 Nr. 4 bis 8 mit SPEC-173 Nr. 4).
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
	case optInlineCount:
		switch wert {
		case "allpages":
			o.Count = model.WahrOption{Gesetzt: true, Wert: true}
		case "none":
			o.Count = model.WahrOption{Gesetzt: true}
		default:
			return fehler(model.URLWert, ort)
		}
	case optFilter:
		return l.leseFilter(wert, ort, o)
	case optFormat:
		if wert == "" {
			return fehler(model.URLWert, ort)
		}
		l.leseFormat(wert, ort)
	}
	return nil
}

// leseZahl liest `$top` und `$skip`: ASCII-Ziffern, führende Nullen erlaubt, von 0 bis 2147483647 (SPEC-162 Nr. 4).
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

// leseSelect liest `$select` (SPEC-162 Nr. 6): `*` und einfache Namen; ein `/` in einem Element ist die Kompatibilitätsursache
// `option`, und der Wert wird dann nicht weiter gelesen (SPEC-173 Nr. 4).
func (l *lauf) leseSelect(wert, ort string, o *model.Optionen) error {
	elemente := strings.Split(wert, ",")
	if len(elemente) > maxSelect {
		return fehler(model.URLZuViele, ort)
	}
	if strings.Contains(wert, "/") {
		l.kompat(model.KompatOption, ort)
		return nil
	}
	sel := &model.SelectOption{}
	for _, e := range elemente {
		switch {
		case e == "*" && !sel.Alle:
			sel.Alle = true
		case einfacherName(e) && !slices.Contains(sel.Namen, e):
			sel.Namen = append(sel.Namen, e)
		default:
			return fehler(model.URLWert, ort)
		}
	}
	o.Select = sel
	return nil
}

// leseOrderBy liest `$orderby` (SPEC-162 Nr. 7): ein `(` oder `/` im Wert macht ihn zum Ausdruck oder Pfad (Kompatibilität
// `option`, nicht weiter gelesen), sonst Elemente `Name` oder `Name asc|desc`.
func (l *lauf) leseOrderBy(wert, ort string, o *model.Optionen) error {
	if strings.ContainsAny(wert, "(/") {
		l.kompat(model.KompatOption, ort)
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
		if !einfacherName(name) || enthaeltSortierung(sort, name) {
			return fehler(model.URLWert, ort)
		}
		sort = append(sort, model.Sortierung{Name: name, Absteigend: k == "desc"})
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

// leseFormat liest `$format` (SPEC-173 Nr. 4): `json`, `application/json` und `application/json;` mit beliebigem Rest sind
// lesbar und ohne Eintrag, jeder andere Wert ist die Kompatibilitätsursache `format`; der Aufrufer weist den leeren Wert ab.
func (l *lauf) leseFormat(wert, ort string) {
	k := asciiKlein(wert)
	if k != "json" && k != "application/json" && !strings.HasPrefix(k, "application/json;") {
		l.kompat(model.KompatFormat, ort)
	}
}

// leseFilter liest `$filter` (SPEC-173 Nr. 1 und SPEC-160 Nr. 8): ohne Leser ist es die Kompatibilitätsursache `filter`, mit
// Leser entscheidet er.
func (l *lauf) leseFilter(wert, ort string, o *model.Optionen) error {
	if wert == "" {
		return fehler(model.URLWert, ort)
	}
	if l.p.filter == nil {
		l.kompat(model.KompatFilter, ort)
		return nil
	}
	ausdruck, err := l.p.filter.LeseFilter(wert)
	if err != nil {
		return l.fehlerDesLesers(err, ort)
	}
	if ausdruck == nil {
		return fmt.Errorf("%w: urlparserv2: FilterLeser lieferte weder Ausdruck noch Fehler", model.ErrUngueltig)
	}
	o.Filter = &model.FilterOption{Text: wert, Ausdruck: ausdruck}
	return nil
}

// fehlerDesLesers ordnet den Fehler des FilterLeser zu (SPEC-173 Nr. 1 c und d): genau ein Zeiger ungleich nil auf einen
// URLFehler der Klasse `ungueltig` oder auf einen KompatFehler mit der Ursache `filter` oder `literal_typ` gilt mit seiner
// Ursache am Ort der Option; jeder andere Fehler ist ein Fehler des Aufrufers.
func (l *lauf) fehlerDesLesers(err error, ort string) error {
	uf, hatU := ersterTreffer[*model.URLFehler](err)
	kf, hatK := ersterTreffer[*model.KompatFehler](err)
	switch {
	case hatU && !hatK && uf != nil && uf.Ursache.Gueltig() && !uf.Ursache.NichtUnterstuetzt():
		return fehler(uf.Ursache, ort)
	case hatK && !hatU && kf != nil && (kf.Ursache == model.KompatFilter || kf.Ursache == model.KompatLiteralTyp):
		l.kompat(kf.Ursache, ort)
		return nil
	}
	return fmt.Errorf("%w: urlparserv2: FilterLeser lieferte einen fremden Fehler", model.ErrUngueltig)
}

// ersterTreffer sucht in der Kette von err (Tiefe zuerst, wie errors.As) den ersten Fehler vom Typ T und meldet auch einen
// typisierten nil-Zeiger als Treffer. Ein nil-Zeiger der beiden Fehlertypen hat keine Kette: sein Unwrap wird nie gerufen.
func ersterTreffer[T error](err error) (T, bool) {
	var keiner T
	if err == nil {
		return keiner, false
	}
	if t, ok := err.(T); ok {
		return t, true
	}
	switch e := err.(type) {
	case *model.URLFehler:
		if e == nil {
			return keiner, false
		}
	case *model.KompatFehler:
		if e == nil {
			return keiner, false
		}
	}
	switch e := err.(type) {
	case interface{ Unwrap() error }:
		return ersterTreffer[T](e.Unwrap())
	case interface{ Unwrap() []error }:
		for _, kind := range e.Unwrap() {
			if t, ok := ersterTreffer[T](kind); ok {
				return t, true
			}
		}
	}
	return keiner, false
}
