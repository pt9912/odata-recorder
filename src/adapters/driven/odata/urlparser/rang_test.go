package urlparser_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// verstoss ist ein Verstoß mit seinem Ort.
type verstoss struct {
	u   model.URLUrsache
	ort string
}

// reihe ist die Folge der Stufe 0 (SPEC-160 Nr. 5): Länge des Pfads, Länge der Query, Segmente des Pfads, Optionen.
func reihe(v verstoss) int {
	switch {
	case v.u == model.URLZuLang && v.ort == "pfad":
		return 0
	case v.u == model.URLZuLang && v.ort == "query":
		return 1
	case v.u == model.URLZuViele && v.ort == "pfad":
		return 2
	}
	return 3
}

// rang() ist das Orakel der Rangfolge, unabhängig vom Parser: ein Verstoß der Stufe 0 geht vor, den Verstößen mit der
// Folge der Stufe 0; sonst gewinnt der erste der Klasse `ungueltig` in Lesefolge, sonst der erste der anderen Klasse.
func rang(stufe0, folge []verstoss) verstoss {
	if len(stufe0) > 0 {
		best := stufe0[0]
		for _, v := range stufe0 {
			if reihe(v) < reihe(best) {
				best = v
			}
		}
		return best
	}
	for _, v := range folge {
		if !klasse501(v.u) {
			return v
		}
	}
	return folge[0]
}

func pruefeVerstoss(t *testing.T, name, pfad, query string, soll verstoss) {
	t.Helper()
	_, err := parse(pfad, query)
	var uf *model.URLFehler
	if !errors.As(err, &uf) || uf.Ursache != soll.u || uf.Ort != soll.ort {
		t.Errorf("%s: %.60q %.60q: %v, erwartet %s %q", name, pfad, query, err, soll.u, soll.ort)
	}
}

func teileMit(n int) string {
	s := make([]string, n)
	for i := range s {
		s[i] = fmt.Sprintf("a%d=1", i)
	}
	return strings.Join(s, ",")
}

// pfadErst ist ein erstes Segment mit genau dem Verstoß der Ursache (ohne führenden `/`); pfad_form und service_dokument
// haben keines.
func pfadErst(u model.URLUrsache) (string, bool) {
	switch u {
	case model.URLKodierung:
		return "%G1", true
	case model.URLSegment:
		return "$foo", true
	case model.URLKey:
		return "S(1,2)", true
	case model.URLLiteral:
		return "S(1x)", true
	case model.URLZuViele:
		return "S(" + teileMit(17) + ")", true
	case model.URLZuLang:
		return "S('" + wiederhole("a", 513) + "')", true
	case model.URLPfadNichtUnterstuetzt:
		return "$batch", true
	case model.URLLiteralTyp:
		return "S(X'AA')", true
	}
	return "", false
}

// pfadAllein ist der ganze Pfad mit genau dem Verstoß der zehn Ursachen des Pfads und seinem Ort.
func pfadAllein(u model.URLUrsache) (string, verstoss) {
	switch u {
	case model.URLPfadForm:
		return "/$metadata/x", verstoss{u, "pfad/2"}
	case model.URLServiceDokument:
		return "/", verstoss{u, "pfad"}
	}
	s, _ := pfadErst(u)
	return "/" + s, verstoss{u, "pfad/1"}
}

func ursachenPfad() []model.URLUrsache {
	return []model.URLUrsache{model.URLKodierung, model.URLSegment, model.URLPfadForm, model.URLZuLang, model.URLZuViele, model.URLKey,
		model.URLLiteral, model.URLServiceDokument, model.URLPfadNichtUnterstuetzt, model.URLLiteralTyp}
}

func ursachenQuery() []model.URLUrsache {
	return []model.URLUrsache{model.URLKodierung, model.URLOptionName, model.URLOptionDoppelt, model.URLWert, model.URLZuViele,
		model.URLOptionNichtUnterstuetzt, model.URLFilter, model.URLWertNichtUnterstuetzt}
}

// queryStueck ist die Query mit genau dem Verstoß der Ursache in der Stelle slot (1 oder 2), die Zahl ihrer Segmente und
// die Nummer des Segments mit dem Verstoß. Die Stellen nehmen verschiedene Optionen, damit keine zweite doppelt wird.
func queryStueck(u model.URLUrsache, slot int) (query string, n, idx int) {
	listen := func(laenge int) string {
		s := make([]string, laenge)
		for i := range s {
			s[i] = fmt.Sprintf("e%d", i)
		}
		return strings.Join(s, ",")
	}
	je := func(eins, zwei string) string {
		if slot == 1 {
			return eins
		}
		return zwei
	}
	switch u {
	case model.URLKodierung:
		return je("a1=%G1", "a2=%G1"), 1, 1
	case model.URLOptionName:
		return je("$foo1=1", "$foo2=1"), 1, 1
	case model.URLOptionDoppelt:
		return je("$count=true&$count=true", "$skip=1&$skip=1"), 2, 2
	case model.URLWert:
		return je("$top=-1", "$format="), 1, 1
	case model.URLZuViele:
		return je("$orderby="+listen(33), "$select="+listen(257)), 1, 1
	case model.URLOptionNichtUnterstuetzt:
		return je("$expand=1", "$search=1"), 1, 1
	case model.URLFilter:
		return "$filter=a", 1, 1
	}
	return je("$orderby=a/b", "$select=a/b"), 1, 1
}

// TestRangEinzeln: jeder Verstoß für sich ergibt seine Ursache und seinen Ort; die Paare bauen darauf auf.
func TestRangEinzeln(t *testing.T) {
	for _, u := range ursachenPfad() {
		p, v := pfadAllein(u)
		pruefeVerstoss(t, "Pfad "+string(u), p, "", v)
	}
	for _, u := range ursachenQuery() {
		for slot := 1; slot <= 2; slot++ {
			q, _, idx := queryStueck(u, slot)
			pruefeVerstoss(t, "Query "+string(u), "/S", q, verstoss{u, fmt.Sprintf("query/%d", idx)})
		}
	}
}

// padding der Gesamtform (Stufe 0).
func s0Pfad(basis string, u model.URLUrsache) string {
	if u == model.URLZuLang {
		return basis + "/" + wiederhole("b", 2100)
	}
	return basis + wiederhole("/b", 17)
}

func s0Query(basis string, u model.URLUrsache) string {
	sep := ""
	if basis != "" {
		sep = "&"
	}
	if u == model.URLZuLang {
		return basis + sep + "x=" + wiederhole("a", 8200)
	}
	return basis + sep + strings.TrimSuffix(wiederhole("y&", 65), "&")
}

// s0Bau baut Pfad und Query mit den gewählten Verstößen der Gesamtform: a Länge des Pfads, b Länge der Query, c Segmente des
// Pfads, d Optionen.
func s0Bau(a, b, c, d bool) (pfad, query string) {
	switch {
	case a && c:
		pfad = "/" + wiederhole("b/", 1100)
	case a:
		pfad = "/" + wiederhole("b", 2100)
	case c:
		pfad = wiederhole("/b", 17)
	default:
		pfad = "/S"
	}
	switch {
	case b && d:
		query = wiederhole("a&", 4200)
	case b:
		query = "x=" + wiederhole("a", 8200)
	case d:
		query = strings.TrimSuffix(wiederhole("y&", 65), "&")
	}
	return pfad, query
}

// TestRangStufe0 (Zeile 6): die vier Verstöße der Gesamtform gehen vor und stehen untereinander in der Folge Länge `pfad`, Länge
// `rawQuery`, Segmente, Optionen. Menge 78 Paare (6 + 72); zwei sind nicht bildbar: ein Pfad `/` nimmt kein Segment mehr auf.
func TestRangStufe0(t *testing.T) {
	vierer := []struct {
		v    verstoss
		pfad bool
	}{{verstoss{model.URLZuLang, "pfad"}, true}, {verstoss{model.URLZuLang, "query"}, false},
		{verstoss{model.URLZuViele, "pfad"}, true}, {verstoss{model.URLZuViele, "query"}, false}}
	n, nichtBildbar := 0, 0
	// 6 Paare innerhalb der Stufe 0
	for i := range vierer {
		for j := i + 1; j < 4; j++ {
			da := func(k int) bool { return i == k || j == k }
			pfad, query := s0Bau(da(0), da(1), da(2), da(3))
			pruefeVerstoss(t, "Stufe 0 innerhalb", pfad, query, rang([]verstoss{vierer[i].v, vierer[j].v}, nil))
			n++
		}
	}
	// 72 Paare mit einem Verstoß der Lesefolge
	for _, x := range vierer {
		for _, y := range append(ursachenPfad(), ursachenQuery()...) {
			pfadY := slices.Contains(ursachenPfad(), y)
			var pfad, query string
			var vy verstoss
			if pfadY {
				pfad, vy = pfadAllein(y)
			} else {
				q, _, idx := queryStueck(y, 1)
				pfad, query, vy = "/S", q, verstoss{y, fmt.Sprintf("query/%d", idx)}
			}
			switch {
			case x.pfad && pfad == "/":
				nichtBildbar++
				continue
			case x.pfad:
				pfad = s0Pfad(pfad, x.v.u)
			default:
				query = s0Query(query, x.v.u)
			}
			pruefeVerstoss(t, "Stufe 0 gegen Lesefolge", pfad, query, rang([]verstoss{x.v}, []verstoss{vy}))
			n++
		}
	}
	if n != 76 || nichtBildbar != 2 || n+nichtBildbar != 78 {
		t.Fatalf("%d Paare gebaut, %d nicht bildbar, erwartet 76 und 2", n, nichtBildbar)
	}
}

// TestRangPfadQuery (Zeile 7): ein Verstoß der Klasse `ungueltig` geht jedem der Klasse `nicht_unterstuetzt` vor, gleich ob
// Pfad oder Query; sonst der erste (der Pfad). Menge 80 Paare (10 × 8).
func TestRangPfadQuery(t *testing.T) {
	n := 0
	for _, a := range ursachenPfad() {
		p, va := pfadAllein(a)
		for _, b := range ursachenQuery() {
			q, _, idx := queryStueck(b, 1)
			vb := verstoss{b, fmt.Sprintf("query/%d", idx)}
			pruefeVerstoss(t, "Pfad gegen Query", p, q, rang(nil, []verstoss{va, vb}))
			n++
		}
	}
	if n != 80 {
		t.Fatalf("%d Paare, erwartet 80", n)
	}
}

// pfadPaar baut die URL für Verstoß a links von b im Pfad, oder meldet, dass das Paar nicht bildbar ist: nur das erste Segment
// trägt Namen, Schlüssel und Literale, und `pfad_form` verlangt `$metadata` oder `$count` davor.
func pfadPaar(a, b model.URLUrsache) (pfad string, va, vb verstoss, ok bool) {
	spaeter := map[model.URLUrsache]string{model.URLKodierung: "%G1", model.URLSegment: "$foo", model.URLPfadNichtUnterstuetzt: "Rest"}
	erst, aErst := pfadErst(a)
	switch {
	case aErst && spaeter[b] != "":
		return "/" + erst + "/" + spaeter[b], verstoss{a, "pfad/1"}, verstoss{b, "pfad/2"}, true
	case (a == model.URLKodierung || a == model.URLSegment) && b == model.URLPfadForm:
		return "/$metadata/" + spaeter[a] + "/x", verstoss{a, "pfad/2"}, verstoss{b, "pfad/3"}, true
	case a == model.URLPfadForm && (b == model.URLKodierung || b == model.URLSegment):
		return "/$metadata/x/" + spaeter[b], verstoss{a, "pfad/2"}, verstoss{b, "pfad/3"}, true
	case a == model.URLPfadForm && b == model.URLPfadForm:
		return "/$metadata/x/y", verstoss{a, "pfad/2"}, verstoss{b, "pfad/3"}, true
	}
	return "", verstoss{}, verstoss{}, false
}

// TestRangPfadPfad (Zeile 8): zwei Verstöße im Pfad: `ungueltig` vor `nicht_unterstuetzt`, sonst links vor rechts. Menge 81 Paare
// (9 × 9); 29 sind bildbar, 52 nicht (beide gehören zum ersten Segment, oder `pfad_form` steht ohne Kontext), sie sind gezählt.
func TestRangPfadPfad(t *testing.T) {
	neun := []model.URLUrsache{model.URLKodierung, model.URLSegment, model.URLPfadForm, model.URLZuLang, model.URLZuViele, model.URLKey,
		model.URLLiteral, model.URLPfadNichtUnterstuetzt, model.URLLiteralTyp}
	n, nicht := 0, 0
	for _, a := range neun {
		for _, b := range neun {
			p, va, vb, ok := pfadPaar(a, b)
			if !ok {
				nicht++
				continue
			}
			pruefeVerstoss(t, "Pfad gegen Pfad", p, "", rang(nil, []verstoss{va, vb}))
			n++
		}
	}
	if n != 29 || nicht != 52 {
		t.Fatalf("%d bildbar, %d nicht bildbar, erwartet 29 und 52", n, nicht)
	}
}

// TestRangQueryQuery (Zeile 9): dieselbe Regel über die Optionen. Menge 64 Paare (8 × 8); `$filter` zweimal ist `option_doppelt`,
// das Paar (filter, filter) ist nicht bildbar und gezählt.
func TestRangQueryQuery(t *testing.T) {
	n, nicht := 0, 0
	for _, a := range ursachenQuery() {
		for _, b := range ursachenQuery() {
			if a == model.URLFilter && b == model.URLFilter {
				nicht++
				continue
			}
			qa, na, ia := queryStueck(a, 1)
			qb, _, ib := queryStueck(b, 2)
			va, vb := verstoss{a, fmt.Sprintf("query/%d", ia)}, verstoss{b, fmt.Sprintf("query/%d", na+ib)}
			pruefeVerstoss(t, "Query gegen Query", "/S", qa+"&"+qb, rang(nil, []verstoss{va, vb}))
			n++
		}
	}
	if n != 63 || nicht != 1 {
		t.Fatalf("%d Paare, %d nicht bildbar, erwartet 63 und 1", n, nicht)
	}
}

// TestRangInDerOption (Zeile 11): die Folge innerhalb einer Option: Kodierung des Namens, Klasse, Wiederholung, Kodierung des
// Werts, Grammatik des Werts. Das Paar (option_name, option_doppelt) ist nicht bildbar: ein doppelter Name hat eine gültige Klasse.
func TestRangInDerOption(t *testing.T) {
	pruefeFaelle(t, []fall{
		{pfad: "/S", query: "%G1foo=1", u: kodi, ort: "query/1"},                  // Kodierung des Namens vor der Klasse
		{pfad: "/S", query: "$foo=%G1", u: nameU, ort: "query/1"},                 // Klasse vor der Kodierung des Werts
		{pfad: "/S", query: "$top=1&$top=-1", u: doppU, ort: "query/2"},           // Wiederholung vor der Grammatik des Werts
		{pfad: "/S", query: "$top=1&$top=%G1", u: doppU, ort: "query/2"},          // Wiederholung vor der Kodierung des Werts
		{pfad: "/S", query: "$top=%G1", u: kodi, ort: "query/1"},                  // Kodierung des Werts vor der Grammatik
		{pfad: "/S", query: "$expand=%G1", u: optNU, ort: "query/1"},              // der Wert einer nicht unterstützten Option bleibt ungelesen
		{pfad: "/S", query: "$top=-1&$expand=1", u: wertU, ort: "query/1"},        // ungueltig vor nicht_unterstuetzt
		{pfad: "/S", query: "$expand=1&$top=-1", u: wertU, ort: "query/2"},        // ebenso, wenn die 501 links steht
		{pfad: "/S", query: "$expand=1&$foo=1&$top=-1", u: nameU, ort: "query/2"}, // der erste der Klasse `ungueltig`
	})
}
