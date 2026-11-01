package urlparserv2_test

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// vd ist ein Verstoß in der Lesefolge einer URL: klasseU ist die Klasse `ungueltig`, sonst ein Kompatibilitätsfehler.
type vd struct {
	klasseU bool
	u       model.URLUrsache
	k       model.KompatUrsache
	ort     string
}

func vU(u model.URLUrsache, ort string) vd    { return vd{klasseU: true, u: u, ort: ort} }
func vK(k model.KompatUrsache, ort string) vd { return vd{k: k, ort: ort} }

// ersterFund ist das Orakel dieser Zeilen, eine Funktion des Tests und nicht des Lesers (SPEC-160 Nr. 5, SPEC-173 Nr. 6): der
// erste Verstoß der Klasse `ungueltig` in der Lesefolge, sonst der erste Kompatibilitätsfehler.
func ersterFund(vs []vd) erw {
	for _, v := range vs {
		if v.klasseU {
			return ung(v.u, v.ort)
		}
	}
	for _, v := range vs {
		return kmp(v.k, v.ort)
	}
	panic("kein Verstoß")
}

// posten ist ein Teil einer URL: im Pfad der ganze Pfad mit seinen Verstößen in Lesefolge, in der Query eine Option (höchstens
// ein Verstoß, sein Ort ist `query/<n>` mit der Stelle der Option).
type posten struct {
	text string
	vds  []vd
}

func pfadKompat() []posten {
	return []posten{
		{"/", []vd{vK(model.KompatPfad, "pfad")}},
		{"/S/N", []vd{vK(model.KompatPfad, "pfad/2")}},
		{"/S(1d)", []vd{vK(model.KompatLiteralTyp, "pfad/1")}},
	}
}

func pfadUngueltig() []posten {
	return []posten{
		{"/S.", []vd{vU(model.URLSegment, "pfad/1")}},
		{"/S(1,,2)", []vd{vU(model.URLKey, "pfad/1")}},
	}
}

func queryKompat() []posten {
	return []posten{
		{"$expand=x", []vd{vK(model.KompatOption, "")}},
		{"$format=xml", []vd{vK(model.KompatFormat, "")}},
		{"$filter=Id%20eq%201", []vd{vK(model.KompatFilter, "")}},
	}
}

func queryUngueltig() []posten {
	return []posten{
		{"$foo=1", []vd{vU(model.URLOptionName, "")}},
		{"$top=x", []vd{vU(model.URLWert, "")}},
		{"%FF=1", []vd{vU(model.URLKodierung, "")}},
	}
}

// zusammen setzt eine URL aus einem Pfad und Optionen zusammen: die Verstöße des Pfads, dann die der Optionen mit ihrem Ort.
func zusammen(p posten, qs ...posten) (pfad, q string, vds []vd) {
	vds = append(vds, p.vds...)
	var texte []string
	for i, qp := range qs {
		texte = append(texte, qp.text)
		for _, v := range qp.vds {
			v.ort = "query/" + strconv.Itoa(i+1)
			vds = append(vds, v)
		}
	}
	return p.text, strings.Join(texte, "&"), vds
}

func pfadOhne() posten { return posten{text: "/S"} }

// TestUngueltigVorKompat (Zeile 7): ODR-E6005 geht jeder Kompatibilität vor, gleich wo sie steht. Menge 38 ausführbare Stellungen
// von sechs Kompatibilitätsformen mit fünf Verstößen: Service-Dokument × Query-Verstoß 3, Pfad × Pfad 5, Pfad-Form × Query-Verstoß 6,
// Query-Form × Pfad-Verstoß 6, Query-Form × Query-Verstoß 18 (früher und später).
func TestUngueltigVorKompat(t *testing.T) {
	pfadKompat, pfadUngueltig, queryKompat, queryUngueltig, pfadOhne := pfadKompat(), pfadUngueltig(), queryKompat(), queryUngueltig(), pfadOhne()
	p := urlparserv2.New()
	n := 0
	pruefeURL := func(pf posten, qs ...posten) {
		pfad, q, vds := zusammen(pf, qs...)
		e := ersterFund(vds)
		if !e.erwartetU() {
			t.Fatalf("%q %q: das Orakel nennt keinen Verstoß der Klasse ungueltig", pfad, q)
		}
		pruefe(t, p, pfad, q, e)
		n++
	}
	for _, u := range queryUngueltig {
		pruefeURL(pfadKompat[0], u)
	}
	for _, pp := range []posten{
		{"/S./N", []vd{vU(model.URLSegment, "pfad/1"), vK(model.KompatPfad, "pfad/2")}},
		{"/S/N/.", []vd{vK(model.KompatPfad, "pfad/2"), vU(model.URLSegment, "pfad/3")}},
		{"/S(1,,2)/N", []vd{vU(model.URLKey, "pfad/1"), vK(model.KompatPfad, "pfad/2")}},
		{"/S(1d)/.", []vd{vK(model.KompatLiteralTyp, "pfad/1"), vU(model.URLSegment, "pfad/2")}},
		{"/S(1d,,2)", []vd{vU(model.URLKey, "pfad/1"), vK(model.KompatLiteralTyp, "pfad/1")}},
	} {
		pruefeURL(pp)
	}
	for _, k := range pfadKompat[1:] {
		for _, u := range queryUngueltig {
			pruefeURL(k, u)
		}
	}
	for _, u := range pfadUngueltig {
		for _, k := range queryKompat {
			pruefeURL(u, k)
		}
	}
	for _, k := range queryKompat {
		for _, u := range queryUngueltig {
			pruefeURL(pfadOhne, k, u)
			pruefeURL(pfadOhne, u, k)
		}
	}
	if n != 3+5+6+6+18 {
		t.Errorf("%d Stellungen, erwartet 38", n)
	}
}

func (e erw) erwartetU() bool { return e.u != "" }

// TestErsteKompat (Zeile 8): unter mehreren Kompatibilitätsfehlern gilt der erste in der Lesefolge. Menge 21: Service-Dokument vor
// einer Query-Form 3, Literal vor Rest-Segment 1, Pfad-Form (Rest-Segment, Literal) vor einer Query-Form 6, Query-Form vor
// Query-Form 9 (gleiche Form zweimal eingeschlossen), zwei Rest-Segmente 1, zwei Literale im Schlüssel 1.
func TestErsteKompat(t *testing.T) {
	pfadKompat, queryKompat, pfadOhne := pfadKompat(), queryKompat(), pfadOhne()
	p := urlparserv2.New()
	n := 0
	pruefeURL := func(pf posten, qs ...posten) {
		pfad, q, vds := zusammen(pf, qs...)
		pruefe(t, p, pfad, q, ersterFund(vds))
		n++
	}
	for _, k := range queryKompat {
		pruefeURL(pfadKompat[0], k)
	}
	pruefeURL(posten{"/S(1d)/N", []vd{vK(model.KompatLiteralTyp, "pfad/1"), vK(model.KompatPfad, "pfad/2")}})
	for _, pk := range pfadKompat[1:] {
		for _, k := range queryKompat {
			pruefeURL(pk, k)
		}
	}
	for _, a := range queryKompat {
		for _, b := range queryKompat {
			pfad, q, vds := zusammen(pfadOhne, a, b)
			if a.text == b.text && a.text != "$expand=x" {
				// `$format` und `$filter` sind behandelte Namen: das zweite Vorkommen ist `option_doppelt` (SPEC-162 Nr. 3 a).
				vds = append(vds, vU(model.URLOptionDoppelt, "query/2"))
			}
			pruefe(t, p, pfad, q, ersterFund(vds))
			n++
		}
	}
	pruefeURL(posten{"/S/N/M", []vd{vK(model.KompatPfad, "pfad/2"), vK(model.KompatPfad, "pfad/3")}})
	pruefeURL(posten{"/S(A=1d,B=2d)", []vd{vK(model.KompatLiteralTyp, "pfad/1"), vK(model.KompatLiteralTyp, "pfad/1")}})
	if n != 21 {
		t.Errorf("%d Paare, erwartet 21", n)
	}
}

// TestErsterUngueltig (Zeile 9): unter mehreren Verstößen der Klasse `ungueltig` gilt der erste in der Lesefolge. Menge 18 Paare der
// fünf Repräsentanten in der Reihenfolge der URL: Pfad × Pfad 3, Pfad × Query 6, Query × Query 9.
func TestErsterUngueltig(t *testing.T) {
	pfadUngueltig, queryUngueltig, pfadOhne := pfadUngueltig(), queryUngueltig(), pfadOhne()
	p := urlparserv2.New()
	n := 0
	pruefeURL := func(pf posten, qs ...posten) {
		pfad, q, vds := zusammen(pf, qs...)
		pruefe(t, p, pfad, q, ersterFund(vds))
		n++
	}
	for _, pp := range []posten{
		{"/S./.", []vd{vU(model.URLSegment, "pfad/1"), vU(model.URLSegment, "pfad/2")}},
		{"/S(1,,2)/.", []vd{vU(model.URLKey, "pfad/1"), vU(model.URLSegment, "pfad/2")}},
		{"/S(1,,2,,)", []vd{vU(model.URLKey, "pfad/1"), vU(model.URLKey, "pfad/1")}},
	} {
		pruefeURL(pp)
	}
	for _, pu := range pfadUngueltig {
		for _, qu := range queryUngueltig {
			pruefeURL(pu, qu)
		}
	}
	for _, a := range queryUngueltig {
		for _, b := range queryUngueltig {
			pruefeURL(pfadOhne, a, b)
		}
	}
	if n != 18 {
		t.Errorf("%d Paare, erwartet 18", n)
	}
}

const marker = "Qx9Marker"

var ortForm = regexp.MustCompile(`^(pfad|query)(/[1-9][0-9]*)?$`)

// ersetze setzt die Rollen des Markers ein: @N steht an einer Stelle eines Namens, @W an der eines Werts oder Literals; die
// Rolle ohne Marker bekommt `x`.
func ersetze(vorlage string, rolleName bool) string {
	n, w := "x", "x"
	if rolleName {
		n = marker
	} else {
		w = marker
	}
	return strings.NewReplacer("@N", n, "@W", w).Replace(vorlage)
}

// TestUrlFehlerWert (Zeile 10): jede der zehn Ursachen der Klasse `ungueltig` hat Code ODR-E6005, genau den Sentinel ihrer Klasse,
// den Ort wie vorgeschrieben und den Nullwert; kein Teil der Eingabe steht in Error() oder Ort. Menge 30: zehn Ursachen × drei
// Eingaben, in denen ein Marker als Name, Literal oder Wert steht.
func TestUrlFehlerWert(t *testing.T) {
	lang := strings.Repeat("a", 2100)
	var q65, k17 []string
	for i := range 17 {
		k17 = append(k17, "K"+strconv.Itoa(i)+"=1")
	}
	for range 64 {
		q65 = append(q65, "a=1")
	}
	pfad17 := strings.Repeat("/a", 17)
	faelle := []struct {
		u        model.URLUrsache
		eingaben [3][3]string // pfad, query, ort
	}{
		{model.URLZuLang, [3][3]string{{"/" + marker + lang, "", "pfad"}, {"/S(K='" + marker + strings.Repeat("a", 520) + "')", "", "pfad/1"}, {"/S", "x=" + marker + strings.Repeat("v", 8200), "query"}}},
		{model.URLZuViele, [3][3]string{{"/" + marker + pfad17, "", "pfad"}, {"/S(K0='" + marker + "'," + strings.Join(k17[1:], ",") + ")", "", "pfad/1"}, {"/S", strings.Join(q65, "&") + "&x=" + marker, "query"}}},
		{model.URLKodierung, [3][3]string{{"/" + marker + "%FF", "", "pfad/1"}, {"/S('" + marker + "%FF')", "", "pfad/1"}, {"/S", "x=" + marker + "%FF", "query/1"}}},
		{model.URLSegment, [3][3]string{{"/" + marker + "-", "", "pfad/1"}, {"/S('" + marker + "')/.", "", "pfad/2"}, {"/S(1)/" + marker + "$", "", "pfad/2"}}},
		{model.URLPfadForm, [3][3]string{{"/$metadata/" + marker, "", "pfad/2"}, {"/S('" + marker + "')/$count", "", "pfad/2"}, {"/S/$count/" + marker, "", "pfad/3"}}},
		{model.URLKey, [3][3]string{{"/S(" + marker + ")", "", "pfad/1"}, {"/S(" + marker + "=)", "", "pfad/1"}, {"/S(A='" + marker + "',A=1)", "", "pfad/1"}}},
		{model.URLLiteral, [3][3]string{{"/S(K=" + marker + ")", "", "pfad/1"}, {"/S(K='" + marker + "'x)", "", "pfad/1"}, {"/S(" + marker + "=null)", "", "pfad/1"}}},
		{model.URLOptionName, [3][3]string{{"/S", "$" + marker + "=1", "query/1"}, {"/S", "$foo=" + marker, "query/1"}, {"/S", "?" + marker + "=1", "query/1"}}},
		{model.URLOptionDoppelt, [3][3]string{{"/S", "$top=1&$top=" + marker, "query/2"}, {"/S", "$select=" + marker + "&$select=" + marker, "query/2"}, {"/S", "$orderby=" + marker + "&%24orderby=1", "query/2"}}},
		{model.URLWert, [3][3]string{{"/S", "$top=" + marker, "query/1"}, {"/S", "$select=" + marker + ",", "query/1"}, {"/S", "$inlinecount=" + marker, "query/1"}}},
	}
	p := urlparserv2.New()
	n := 0
	for _, f := range faelle {
		for _, e := range f.eingaben {
			pruefe(t, p, e[0], e[1], ung(f.u, e[2]))
			_, err := p.ParseURL(e[0], e[1])
			var uf *model.URLFehler
			if !errors.As(err, &uf) {
				continue
			}
			if strings.Contains(err.Error(), marker) || strings.Contains(uf.Ort, marker) || !ortForm.MatchString(uf.Ort) || uf.Code() != model.CodeE6005 ||
				uf.Ursache.NichtUnterstuetzt() || !strings.HasPrefix(err.Error(), "ODR-E6005: OData-URL ungültig: "+string(f.u)) {
				t.Errorf("%s: Fehler %q mit Ort %q", f.u, err.Error(), uf.Ort)
			}
			n++
		}
	}
	if n != 30 {
		t.Errorf("%d Eingaben, erwartet 30", n)
	}
}

// TestKompatFehlerWert (Zeile 11): die fünf Ursachen des URL-Lesers mit dem Ort von SPEC-160 Nr. 4; nie ein URLFehler; kein Teil
// der Eingabe im Fehler. Menge 30: fünf Ursachen × drei Stellen × zwei Marker (Name, Wert).
func TestKompatFehlerWert(t *testing.T) {
	faelle := []struct {
		k      model.KompatUrsache
		stelle [3][3]string // pfad, query, ort
	}{
		{model.KompatFormat, [3][3]string{{"/S", "$format=@W", "query/1"}, {"/S", "@N=1&$format=@W", "query/2"}, {"/S", "@N=1&a=2&$format=@W", "query/3"}}},
		{model.KompatOption, [3][3]string{{"/S", "$expand=@W", "query/1"}, {"/S", "@N=1&$expand=@W", "query/2"}, {"/S", "@N=1&a=2&$expand=@W", "query/3"}}},
		{model.KompatFilter, [3][3]string{{"/S", "$filter=@W", "query/1"}, {"/S", "@N=1&$filter=@W", "query/2"}, {"/S", "@N=1&a=2&$filter=@W", "query/3"}}},
		{model.KompatPfad, [3][3]string{{"/$batch", "@N=@W", "pfad/1"}, {"/S/@W", "@N=1", "pfad/2"}, {"/S(1)/@N/@W", "a=@W", "pfad/2"}}},
		{model.KompatLiteralTyp, [3][3]string{{"/S(binary'@W')", "@N=1", "pfad/1"}, {"/S(@N=1,B=binary'@W')", "", "pfad/1"}, {"/S(A=1,B=2,C=X'@W')", "@N=1", "pfad/1"}}},
	}
	p := urlparserv2.New()
	n := 0
	for _, f := range faelle {
		for _, s := range f.stelle {
			for _, rolleName := range []bool{true, false} {
				pfad, q := ersetze(s[0], rolleName), ersetze(s[1], rolleName)
				pruefe(t, p, pfad, q, kmp(f.k, s[2]))
				_, err := p.ParseURL(pfad, q)
				var kf *model.KompatFehler
				var uf *model.URLFehler
				if !errors.As(err, &kf) || errors.As(err, &uf) {
					continue
				}
				if strings.Contains(err.Error(), marker) || strings.Contains(kf.Ort, marker) || !ortForm.MatchString(kf.Ort) ||
					kf.Code() != model.CodeE6006 || !strings.HasPrefix(err.Error(), "ODR-E6006: OData-V2-Kompatibilität: "+string(f.k)) {
					t.Errorf("%s: Fehler %q mit Ort %q", f.k, err.Error(), kf.Ort)
				}
				n++
			}
		}
	}
	if n != 30 {
		t.Errorf("%d Eingaben, erwartet 30", n)
	}
}
