package admin

import (
	"net/url"
	"slices"
	"strings"
)

// segment ist ein Teil eines Musters (SPEC-238 Nr. 2 b): ein Literal oder, wenn param nicht leer ist, ein Parameter.
type segment struct {
	literal string
	param   string
}

// route ist eine geprüfte Route mit zerlegtem Muster.
type route struct {
	methode string
	segs    []segment
	handler Handler
}

// gueltigeMethode meldet, ob m eine der fünf Methoden einer Route ist (SPEC-238 Nr. 2 a); die Schreibweise zählt.
func gueltigeMethode(m string) bool {
	return slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, m)
}

// literalZeichen meldet, ob c in einem Literal vorkommen darf: A-Z, a-z, 0-9, `.`, `_`, `~`, `-` (SPEC-238 Nr. 2 b).
func literalZeichen(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '~' || c == '-'
}

// gueltigerName meldet, ob name zu `[a-z][a-z_]*` passt (SPEC-238 Nr. 2 b).
func gueltigerName(name string) bool {
	if name == "" || name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for i := 1; i < len(name); i++ {
		if (name[i] < 'a' || name[i] > 'z') && name[i] != '_' {
			return false
		}
	}
	return true
}

// parseSegment liest ein Segment des Musters; ok ist falsch bei einem Segment außerhalb der Grammatik.
func parseSegment(s string) (segment, bool) {
	if strings.HasPrefix(s, "{") {
		if !strings.HasSuffix(s, "}") || !gueltigerName(s[1:len(s)-1]) {
			return segment{}, false
		}
		return segment{param: s[1 : len(s)-1]}, true
	}
	if s == "" {
		return segment{}, false
	}
	for i := 0; i < len(s); i++ {
		if !literalZeichen(s[i]) {
			return segment{}, false
		}
	}
	return segment{literal: s}, true
}

// parseMuster zerlegt ein Muster nach SPEC-238 Nr. 2 b; ok ist falsch bei einem Muster außerhalb der Grammatik
// oder bei zwei Parametern gleichen Namens.
func parseMuster(muster string) ([]segment, bool) {
	if !strings.HasPrefix(muster, "/") {
		return nil, false
	}
	teile := strings.Split(muster[1:], "/")
	segs := make([]segment, 0, len(teile))
	namen := map[string]bool{}
	for _, t := range teile {
		sg, ok := parseSegment(t)
		if !ok || (sg.param != "" && namen[sg.param]) {
			return nil, false
		}
		namen[sg.param] = true
		segs = append(segs, sg)
	}
	return segs, true
}

// ueberlappt meldet, ob zwei Routen auf denselben Pfad passen können: gleiche Methode, gleiche Zahl der Segmente
// und an jeder Stelle gleiches Literal oder mindestens ein Parameter (SPEC-238 Nr. 2 d).
func ueberlappt(a, b route) bool {
	if a.methode != b.methode || len(a.segs) != len(b.segs) {
		return false
	}
	for i := range a.segs {
		x, y := a.segs[i], b.segs[i]
		if x.param == "" && y.param == "" && x.literal != y.literal {
			return false
		}
	}
	return true
}

// passt meldet, ob die dekodierten Segmente des Pfades auf das Muster der Route passen, und liefert die Werte
// der Parameter (SPEC-238 Nr. 2).
func (r route) passt(pfad []string) (map[string]string, bool) {
	if len(pfad) != len(r.segs) {
		return nil, false
	}
	werte := map[string]string{}
	for i, sg := range r.segs {
		switch {
		case sg.param != "" && pfad[i] != "":
			werte[sg.param] = pfad[i]
		case sg.param == "" && pfad[i] == sg.literal:
		default:
			return nil, false
		}
	}
	return werte, true
}

// zerlegePfad zerlegt den kodierten Pfad an `/` und dekodiert jedes Segment einzeln; ok ist falsch, wenn der
// Pfad nicht mit `/` beginnt oder sich ein Segment nicht dekodieren lässt (SPEC-238 Nr. 2).
func zerlegePfad(kodiert string) ([]string, bool) {
	if !strings.HasPrefix(kodiert, "/") {
		return nil, false
	}
	teile := strings.Split(kodiert[1:], "/")
	for i, t := range teile {
		d, err := url.PathUnescape(t)
		if err != nil {
			return nil, false
		}
		teile[i] = d
	}
	return teile, true
}

// treffer ist das Ergebnis der Zuordnung: die Route mit ihren Parameterwerten, oder die Methoden aller Routen,
// die auf den Pfad passen (leer: 404).
type treffer struct {
	route  *route
	werte  map[string]string
	allow  []string
	gefund bool
}

// ordne ordnet Methode und kodierten Pfad den Routen zu (SPEC-238 Nr. 2): kein Pfad-Treffer heißt 404
// (gefund falsch, allow leer), Pfad-Treffer ohne die Methode 405 (allow sortiert nach Bytefolge).
func ordne(routen []route, methode, kodiert string) treffer {
	pfad, ok := zerlegePfad(kodiert)
	if !ok {
		return treffer{}
	}
	var t treffer
	for i := range routen {
		werte, passt := routen[i].passt(pfad)
		if !passt {
			continue
		}
		t.allow = append(t.allow, routen[i].methode)
		if routen[i].methode == methode {
			t.route, t.werte, t.gefund = &routen[i], werte, true
		}
	}
	slices.Sort(t.allow)
	return t
}
