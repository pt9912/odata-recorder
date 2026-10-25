package baselineimport

import (
	"slices"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Gründe, aus denen der Import einen Exchange nicht liest, in der Reihenfolge der Prüfung
// (SPEC-127 Nr. 8); MetadatenDoppelt trifft allein einen zweiten Kandidaten des $metadata.
const (
	grundMethode          = "methode"
	grundPfad             = "pfad"
	grundCapture          = "capture"
	grundStatus           = "status"
	grundBody             = "body"
	grundQuery            = "query"
	grundKeinJSON         = "kein_json"
	grundKeinXML          = "kein_xml"
	grundOhneKontext      = "ohne_kontext"
	grundOhneSchluessel   = "projektion_ohne_schluessel"
	grundMetadatenDoppelt = "metadata_doppelt"
	pfadMetadaten         = "/$metadata"
	optionFormat          = "$format"
	optionSelect          = "$select"
)

// spur ist die Art der Antwort nach ihrem Pfad (SPEC-127 Nr. 8 b).
type spur int

const (
	spurKeine spur = iota
	spurMetadaten
	spurCollection
	spurEntity
)

// urteil ist der Ausgang der Auswahl für einen Exchange: grund leer heißt lesbar.
type urteil struct {
	spur        spur
	set         string
	grund       string
	ungefiltert bool
}

func asciiKlein(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

func hexWert(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// dekodiere löst %XX auf; ist eine Kodierung ungültig, bleibt der Text unverändert (SPEC-127 Nr. 8 h, i).
func dekodiere(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			out = append(out, s[i])
			continue
		}
		if i+2 >= len(s) || hexWert(s[i+1]) < 0 || hexWert(s[i+2]) < 0 {
			return s
		}
		out = append(out, byte(hexWert(s[i+1])<<4|hexWert(s[i+2])))
		i += 2
	}
	return string(out)
}

// option ist ein Segment der Query: der dekodierte Name und der Rest hinter dem ersten `=`.
type option struct {
	name string
	wert string
}

// optionen teilt die Query an `&`; leere Segmente zählen nicht (SPEC-127 Nr. 8 i).
func optionen(raw string) []option {
	var out []option
	for _, seg := range strings.Split(raw, "&") {
		if seg == "" {
			continue
		}
		n, w, _ := strings.Cut(seg, "=")
		out = append(out, option{name: dekodiere(n), wert: w})
	}
	return out
}

// optionLesbar sagt, ob die Option die Form der Elemente erhält: die zehn lesbaren Systemoptionen
// und jeder Name ohne `$` (SPEC-127 Nr. 8 i).
func optionLesbar(o option) bool {
	switch o.name {
	case "$format", "$count", "$select", "$orderby", "$filter", "$top", "$skip", "$skiptoken", "$search", "$expand":
		return true
	}
	return !strings.HasPrefix(o.name, "$")
}

func ungefilterteOption(o option) bool {
	switch o.name {
	case "$format", "$count", "$select", "$orderby":
		return true
	}
	return false
}

// ungefilterteQuery sagt, ob jede Option eine der vier ist, die die Menge nicht einschränken
// (SPEC-127 Nr. 8 i); die leere Query ist es.
func ungefilterteQuery(os []option) bool {
	return !slices.ContainsFunc(os, func(o option) bool { return !ungefilterteOption(o) })
}

// ohneSchluessel sagt, ob `$select` den Schlüssel nicht auswählt (SPEC-127 Nr. 8 h).
func ohneSchluessel(os []option, keys []string) bool {
	var auswahl []string
	gesehen := false
	for _, o := range os {
		if o.name == optionSelect {
			gesehen = true
			auswahl = append(auswahl, strings.Split(dekodiere(o.wert), ",")...)
		}
	}
	if !gesehen || slices.Contains(auswahl, "*") {
		return false
	}
	return slices.ContainsFunc(keys, func(k string) bool { return !slices.Contains(auswahl, k) })
}

// pfadSpur ordnet den Pfad nach SPEC-127 Nr. 8 j einer Spur und einem Set zu.
func pfadSpur(pfad string, m model.ServiceModel) (spur, string) {
	rest, ok := strings.CutPrefix(pfad, "/")
	if !ok {
		return spurKeine, ""
	}
	i := strings.IndexByte(rest, '(')
	if i < 0 {
		if _, ok := m.EntitySet(rest); ok {
			return spurCollection, rest
		}
		return spurKeine, ""
	}
	name := rest[:i]
	if _, ok := m.EntitySet(name); !ok {
		return spurKeine, ""
	}
	inText := false
	j := i + 1
	for ; j < len(rest); j++ {
		if c := rest[j]; c == '\'' {
			inText = !inText
		} else if c == ')' && !inText {
			break
		}
	}
	if j == i+1 || j != len(rest)-1 {
		return spurKeine, ""
	}
	return spurEntity, name
}

// inhaltstypOK prüft den Header Content-Type (SPEC-127 Nr. 8 g): genau ein Wert, ein Medientyp
// aus medien und, wenn ein charset genannt ist, utf-8; ohneKontext meldet odata.metadata=none.
func inhaltstypOK(h map[string][]string, medien ...string) (ok, ohneKontext bool) {
	var werte []string
	for n, w := range h {
		if asciiKlein(n) == "content-type" {
			werte = append(werte, w...)
		}
	}
	if len(werte) != 1 {
		return false, false
	}
	typ, params, _ := strings.Cut(werte[0], ";")
	if !slices.Contains(medien, asciiKlein(strings.Trim(typ, " \t"))) {
		return false, false
	}
	for _, p := range strings.Split(params, ";") {
		n, w, _ := strings.Cut(p, "=")
		w = strings.Trim(w, " \t")
		if len(w) >= 2 && w[0] == '"' && w[len(w)-1] == '"' {
			w = w[1 : len(w)-1]
		}
		switch w = asciiKlein(w); asciiKlein(strings.Trim(n, " \t")) {
		case "charset":
			if w != "utf-8" {
				return false, false
			}
		case "odata.metadata":
			ohneKontext = ohneKontext || w == "none"
		}
	}
	return true, ohneKontext
}

// klassifiziere wendet die Prüfungen von SPEC-127 Nr. 8 in ihrer Reihenfolge an; die erste
// zutreffende entscheidet. Ohne Modell erkennt es allein die Spur Metadaten: jeder andere Pfad
// bleibt ohne Spur und ohne Grund.
func klassifiziere(e model.Exchange, m *model.ServiceModel) urteil {
	if e.Method != "GET" {
		return urteil{grund: grundMethode}
	}
	u := urteil{}
	switch {
	case e.Path == pfadMetadaten:
		u.spur = spurMetadaten
	case m != nil:
		u.spur, u.set = pfadSpur(e.Path, *m)
		if u.spur == spurKeine {
			return urteil{grund: grundPfad}
		}
	default:
		return u
	}
	meta := u.spur == spurMetadaten
	os := optionen(e.RawQuery)
	if u.grund = grundDerAntwort(e, meta, os); u.grund == "" {
		u.grund = grundDesInhalts(e, meta, os, m, u.set)
	}
	u.ungefiltert = u.grund == "" && u.spur == spurCollection && ungefilterteQuery(os)
	return u
}

// grundDerAntwort prüft Capture, Status, Body und Query (Nr. 8 c bis f).
func grundDerAntwort(e model.Exchange, meta bool, os []option) string {
	switch {
	case e.Capture != model.CaptureComplete:
		return grundCapture
	case e.ResponseStatus != 200:
		return grundStatus
	case e.ResponseBody != model.BodyPresent || meta && e.ResponseRedact != model.RedactionNone:
		return grundBody
	case meta && slices.ContainsFunc(os, func(o option) bool { return o.name != optionFormat }),
		!meta && slices.ContainsFunc(os, func(o option) bool { return !optionLesbar(o) }):
		return grundQuery
	}
	return ""
}

// grundDesInhalts prüft den Content-Type und bei Collection und Entity die Auswahl des Schlüssels
// (Nr. 8 g, h); m ist bei diesen Spuren gesetzt.
func grundDesInhalts(e model.Exchange, meta bool, os []option, m *model.ServiceModel, set string) string {
	medien := []string{"application/json"}
	if meta {
		medien = []string{"application/xml", "text/xml"}
	}
	ok, ohneKontext := inhaltstypOK(e.ResponseHeaders, medien...)
	switch {
	case !ok && meta:
		return grundKeinXML
	case !ok:
		return grundKeinJSON
	case ohneKontext && !meta:
		return grundOhneKontext
	case !meta && ohneSchluessel(os, schluesselDes(*m, set)):
		return grundOhneSchluessel
	}
	return ""
}

// schluesselDes liefert die Schlüssel-Properties des Typs des Sets.
func schluesselDes(m model.ServiceModel, set string) []string {
	s, ok := m.EntitySet(set)
	if !ok {
		return nil
	}
	t, ok := m.EntityType(s.EntityType)
	if !ok {
		return nil
	}
	return t.Keys
}
