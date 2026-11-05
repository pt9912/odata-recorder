package urlparserv2_test

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func optOk(o model.Optionen) erw { return ok(mengeR("S"), o) }

func sonstige(paare ...string) model.Optionen {
	var o model.Optionen
	for i := 0; i < len(paare); i += 2 {
		o.Sonstige = append(o.Sonstige, model.SonstigeOption{Name: paare[i], Wert: paare[i+1]})
	}
	return o
}

func optFall(q string, e erw) pfadFall { return pfadFall{"/S", q, e} }

// TestQueryZerlegung (Zeile 36): die Zerlegung der Query ist model.QuerySegmente (SPEC-162 Nr. 1): `&&`, `;` trennt nichts,
// `a=b=c`, `=v`, ohne `=`, leere Query. Menge 8; das Orakel ist die Referenz-Zerlegung.
func TestQueryZerlegung(t *testing.T) {
	eingaben := []string{"a=1&&b=2", "a=1;b=2", "a=b=c", "=v", "x", "", "&&&", "a=1&"}
	p := urlparserv2.New()
	for _, q := range eingaben {
		segs := model.QuerySegmente(q)
		var soll erw
		var paare []string
		for _, s := range segs {
			paare = append(paare, s.Name, s.Wert)
		}
		if len(segs) == 1 && segs[0].Name == "" {
			soll = ung(model.URLOptionName, "query/1")
		} else {
			soll = optOk(sonstige(paare...))
		}
		pruefe(t, p, "/S", q, soll)
	}
	// feste Erwartung neben dem Orakel: das erste `=` trennt, `;` nicht
	pruefe(t, p, "/S", "a=b=c", optOk(sonstige("a", "b=c")))
	pruefe(t, p, "/S", "a=1;b=2", optOk(sonstige("a", "1;b=2")))
	if len(eingaben) != 8 {
		t.Errorf("%d Eingaben, erwartet 8", len(eingaben))
	}
}

// TestQueryDekodierung (Zeile 37): Name und Wert je genau einmal, `+` bleibt; der Wert einer Kompatibilitätsoption wird nicht
// gelesen; Folge der Prüfungen: Kodierung des Namens, Klasse, Wiederholung, Kodierung des Werts, Grammatik. Menge 21.
func TestQueryDekodierung(t *testing.T) {
	kod := ung(model.URLKodierung, "query/1")
	faelle := []pfadFall{
		optFall("%24top=5", optOk(model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 5}})),
		optFall("%2524top=5", ung(model.URLOptionName, "query/1")), // der Name wird genau einmal dekodiert: `%24top`
		optFall("%=1", kod), optFall("%1=1", kod), optFall("%G1=1", kod), optFall("%FF=1", kod), optFall("%00=1", kod),
		optFall("x=%41", optOk(sonstige("x", "A"))),
		optFall("x=%", kod), optFall("x=%FF", kod), optFall("x=%00", kod), optFall("x=%1", kod), optFall("x=%G1", kod),
		optFall("$top=%FFx", kod), optFall("$foo=%FF", ung(model.URLOptionName, "query/1")),
		optFall("$top=1&$top=%FF", ung(model.URLOptionDoppelt, "query/2")), optFall("%FF=1&$foo=1", kod),
		optFall("$expand=%FF", kmp(model.KompatOption, "query/1")), optFall("$skiptoken=%", kmp(model.KompatOption, "query/1")),
		optFall("x=%2524", optOk(sonstige("x", "%24"))),
		optFall("$orderby=Name+desc", ung(model.URLWert, "query/1")),
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 21 {
		t.Errorf("%d Fälle, erwartet 21", len(faelle))
	}
}

// TestNamenKlassen (Zeile 38): die sieben behandelten Namen doppelt sind `option_doppelt`; ein unbekannter Name mit `$` ist
// `option_name`; Alias und Herstellername sind `Sonstige`; die Schreibweise zählt. Menge 48.
func TestNamenKlassen(t *testing.T) {
	behandelt := []struct{ name, wert string }{{"$select", "A"}, {"$orderby", "A"}, {"$top", "1"}, {"$skip", "1"}, {"$inlinecount", "none"},
		{"$filter", "x"}, {"$format", "json"}}
	var faelle []pfadFall
	for _, b := range behandelt {
		var alles strings.Builder
		for i := range len(b.name) {
			fmt.Fprintf(&alles, "%%%02x", b.name[i])
		}
		for _, zweiter := range []string{b.name, "%24" + b.name[1:], alles.String()} {
			faelle = append(faelle, optFall(b.name+"="+b.wert+"&"+zweiter+"="+b.wert, ung(model.URLOptionDoppelt, "query/2")))
		}
	}
	for _, n := range []string{"$Top", "$foo", "$", "$links", "$value", "$ref", "$batch", "$FORMAT", "$Format", "$expandx"} {
		faelle = append(faelle, optFall(n+"=1", ung(model.URLOptionName, "query/1")))
	}
	for _, f := range [][3]string{{"@a", "", ""}, {"@a1", "", ""}, {"sap-client=100", "sap-client", "100"}, {"x", "x", ""}, {"x=", "x", ""},
		{"api-version=2", "api-version", "2"}, {"a.b=1", "a.b", "1"}, {"_x=1", "_x", "1"}} {
		name := strings.SplitN(f[0], "=", 2)[0]
		wert := ""
		if _, w, hat := strings.Cut(f[0], "="); hat {
			wert = w
		}
		faelle = append(faelle, optFall(f[0], optOk(sonstige(name, wert))))
	}
	a := strings.Repeat("a", 128)
	for _, n := range []string{"@", "@1", "?x", "=v", zeichen(0xE4) + "=1", "1x=1", "-x=1", a + "a=1"} {
		faelle = append(faelle, optFall(n, ung(model.URLOptionName, "query/1")))
	}
	faelle = append(faelle, optFall(a+"=1", optOk(sonstige(a, "1"))))
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 48 {
		t.Errorf("%d Fälle, erwartet 48", len(faelle))
	}
}

// TestOptionenAusserhalb (Zeile 39): die zwölf Optionen außerhalb sind Kompatibilität `option`, mit jedem Wert und beliebig oft.
// Menge 79: 12 Namen × 3 Werte (leer, nichtleer, ungültig kodiert) × 2 Anzahlen und 7 Gegenfälle im Umfang.
func TestOptionenAusserhalb(t *testing.T) {
	var faelle []pfadFall
	for _, n := range ausserhalbOptionen() {
		for _, w := range []string{"", "x", "%FF"} {
			faelle = append(faelle, optFall(n+"="+w, kmp(model.KompatOption, "query/1")),
				optFall(n+"="+w+"&"+n+"="+w, kmp(model.KompatOption, "query/1")))
		}
	}
	top := model.ZahlOption{Gesetzt: true, Wert: 1}
	imUmfang := []pfadFall{
		optFall("$top=1", optOk(model.Optionen{Top: top})), optFall("$skip=1", optOk(model.Optionen{Skip: top})),
		optFall("$select=Id", optOk(model.Optionen{Select: &model.SelectOption{Namen: []string{"Id"}}})),
		optFall("$orderby=Id", optOk(model.Optionen{OrderBy: []model.Sortierung{{Name: "Id"}}})),
		optFall("$inlinecount=allpages", optOk(model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}})),
		optFall("$format=json", optOk(model.Optionen{})),
	}
	laufe(t, urlparserv2.New(), append(faelle, imUmfang...))
	f := &fakeFilterLeser{}
	pruefe(t, urlparserv2.New(urlparserv2.WithFilter(f)), "/S", "$filter=x",
		optOk(model.Optionen{Filter: &model.FilterOption{Text: "x", Ausdruck: model.FilterEigenschaft{Name: "x"}}}))
	if len(faelle)+len(imUmfang)+1 != 79 || len(ausserhalbOptionen()) != 12 {
		t.Errorf("%d Fälle, erwartet 79", len(faelle)+len(imUmfang)+1)
	}
}

// randLeerraum sind die neun Zeichen am Rand eines Werts (U+0020, U+0009, U+000A, U+000D, U+0085, U+00A0, U+2003, U+3000, U+FEFF),
// als Escape der UTF-8-Bytes geschrieben.
func randLeerraum() []string {
	var out []string
	for _, cp := range []int{0x20, 0x09, 0x0A, 0x0D, 0x85, 0xA0, 0x2003, 0x3000, 0xFEFF} {
		out = append(out, pct(zeichen(cp)))
	}
	return out
}

// TestTopSkip (Zeile 40): ASCII-Ziffern, 0 bis 2147483647. Menge 70: 2 Optionen × 35 Werte. Das Orakel ist ein Vergleich der
// Ziffernfolge mit strconv.ParseUint, nicht der Leser.
func TestTopSkip(t *testing.T) {
	werte := []string{"0", "00", "007", "1", "2147483647", "2147483648", strings.Repeat("1", 20), "-1", "-0", "+1", "1.0", "1e3", "0x10", ""}
	for _, c := range randLeerraum() {
		werte = append(werte, c+"5", "5"+c)
	}
	// Ziffern jenseits von ASCII: arabisch-indisch, volle Breite, Devanagari
	for _, cp := range []int{0x663, 0xFF11, 0x967} {
		werte = append(werte, pct(zeichen(cp)))
	}
	if len(werte) != 35 {
		t.Fatalf("%d Werte, erwartet 35", len(werte))
	}
	p := urlparserv2.New()
	for _, name := range []string{"$top", "$skip"} {
		for _, w := range werte {
			e := ung(model.URLWert, "query/1")
			if v, err := strconv.ParseUint(w, 10, 64); err == nil && v <= math.MaxInt32 && w != "" && !strings.ContainsAny(w, "+-") {
				z := model.ZahlOption{Gesetzt: true, Wert: int(v)}
				o := model.Optionen{Top: z}
				if name == "$skip" {
					o = model.Optionen{Skip: z}
				}
				e = optOk(o)
			}
			pruefe(t, p, "/S", name+"="+w, e)
		}
	}
}

// TestInlinecount (Zeile 41): `allpages` und `none`, die Schreibweise zählt, Leerraum am Rand ist `wert`, doppelt ist
// `option_doppelt`, zusammen mit `$count` gilt dessen Kompatibilität `option`. Menge 33.
func TestInlinecount(t *testing.T) {
	faelle := []pfadFall{
		optFall("$inlinecount=allpages", optOk(model.Optionen{Count: model.WahrOption{Gesetzt: true, Wert: true}})),
		optFall("$inlinecount=none", optOk(model.Optionen{Count: model.WahrOption{Gesetzt: true}})),
		optFall("$inlinecount=", ung(model.URLWert, "query/1")),
		optFall("$inlinecount=allpages&$inlinecount=none", ung(model.URLOptionDoppelt, "query/2")),
		optFall("$inlinecount=allpages&$count=true", kmp(model.KompatOption, "query/2")),
		optFall("$count=true&$inlinecount=allpages", kmp(model.KompatOption, "query/1")),
		optFall("a=1", optOk(sonstige("a", "1"))), // ohne `$inlinecount` ist Count nicht gesetzt
	}
	for _, w := range []string{"ALLPAGES", "AllPages", "None", "all", "true", "false", "1", "allpages,none", "allpagesx"} {
		faelle = append(faelle, optFall("$inlinecount="+w, ung(model.URLWert, "query/1")))
	}
	for _, c := range randLeerraum() {
		faelle = append(faelle, optFall("$inlinecount="+c+"allpages", ung(model.URLWert, "query/1")),
			optFall("$inlinecount=allpages"+c, ung(model.URLWert, "query/1")))
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 34 {
		t.Errorf("%d Fälle, erwartet 34", len(faelle))
	}
}

// TestSelect (Zeile 42): `*` und Namen in der Reihenfolge der URL; ein Pfad mit `/` ist Kompatibilität `option`; sonst `wert`;
// 256 Elemente gültig, 257 `zu_viele`. Menge 21.
func TestSelect(t *testing.T) {
	sel := func(namen ...string) erw { return optOk(model.Optionen{Select: &model.SelectOption{Namen: namen}}) }
	alle := func(namen ...string) erw {
		return optOk(model.Optionen{Select: &model.SelectOption{Alle: true, Namen: namen}})
	}
	wert := ung(model.URLWert, "query/1")
	faelle := []pfadFall{
		optFall("$select=*", alle()), optFall("$select=Id", sel("Id")), optFall("$select=Id,Name", sel("Id", "Name")),
		optFall("$select=*,Name", alle("Name")), optFall("$select=Name,Id", sel("Name", "Id")),
		optFall("$select=", wert), optFall("$select=Id,,Name", wert), optFall("$select=,Id", wert), optFall("$select=Id,", wert),
		optFall("$select=Id,%20Name", wert), optFall("$select=Id,(x)", wert), optFall("$select=Id,Id", wert), optFall("$select=*,*", wert),
		optFall("$select="+strings.Repeat("a", 129), wert),
		optFall("$select=A/B", kmp(model.KompatOption, "query/1")), optFall("$select=A/B,C", kmp(model.KompatOption, "query/1")),
		optFall("$select=C,A/B", kmp(model.KompatOption, "query/1")), optFall("$select=A/B/C", kmp(model.KompatOption, "query/1")),
		optFall("$select="+strings.Join(namenListe(256), ","), sel(namenListe(256)...)),
		optFall("$select="+strings.Join(namenListe(257), ","), ung(model.URLZuViele, "query/1")),
		optFall("$select=Id,id", sel("Id", "id")),
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 21 {
		t.Errorf("%d Fälle, erwartet 21", len(faelle))
	}
}

// TestOrderBy (Zeile 43): `asc` und `desc` in jeder ASCII-Schreibweise, genau ein Leerzeichen; ein `(` oder `/` im Wert ist
// Kompatibilität `option`; sonst `wert`; 32 Elemente gültig, 33 `zu_viele`. Menge 24.
func TestOrderBy(t *testing.T) {
	ord := func(s ...model.Sortierung) erw { return optOk(model.Optionen{OrderBy: s}) }
	wert := ung(model.URLWert, "query/1")
	opt := kmp(model.KompatOption, "query/1")
	var n32 []model.Sortierung
	for _, nm := range namenListe(32) {
		n32 = append(n32, model.Sortierung{Name: nm})
	}
	faelle := []pfadFall{
		optFall("$orderby=A", ord(model.Sortierung{Name: "A"})), optFall("$orderby=A%20asc", ord(model.Sortierung{Name: "A"})),
		optFall("$orderby=A%20desc", ord(model.Sortierung{Name: "A", Absteigend: true})),
		optFall("$orderby=A%20DESC", ord(model.Sortierung{Name: "A", Absteigend: true})),
		optFall("$orderby=A,B%20desc", ord(model.Sortierung{Name: "A"}, model.Sortierung{Name: "B", Absteigend: true})),
		optFall("$orderby=A%20Asc", ord(model.Sortierung{Name: "A"})),
		optFall("$orderby="+strings.Join(namenListe(32), ","), ord(n32...)),
		optFall("$orderby=Name%20DeSc,Id", ord(model.Sortierung{Name: "Name", Absteigend: true}, model.Sortierung{Name: "Id"})),
		optFall("$orderby=", wert), optFall("$orderby=,A", wert), optFall("$orderby=A,", wert), optFall("$orderby=A%20%20desc", wert),
		optFall("$orderby=A%20up", wert), optFall("$orderby=A%20asc%20desc", wert), optFall("$orderby=A%20mul%202", wert),
		optFall("$orderby=A,A", wert), optFall("$orderby=A,A%20desc", wert),
		optFall("$orderby="+strings.Join(namenListe(33), ","), ung(model.URLZuViele, "query/1")),
		optFall("$orderby=tolower(Name)", opt), optFall("$orderby=tolower(Name)%20desc", opt), optFall("$orderby=A,tolower(B)", opt),
		optFall("$orderby=C/Name", opt), optFall("$orderby=C/Name%20desc", opt), optFall("$orderby=A,C/Name", opt),
	}
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 24 {
		t.Errorf("%d Fälle, erwartet 24", len(faelle))
	}
}

// TestFormat (Zeile 44): ohne Beachtung der ASCII-Schreibweise `json`, `application/json` oder `application/json;` mit beliebigem
// Rest ist lesbar und ohne Eintrag im Ergebnis; jeder andere nichtleere Wert ist Kompatibilität `format`; leer ist `wert`. Menge 41.
func TestFormat(t *testing.T) {
	gueltig := []string{"json", "JSON", "Json", "application/json", "Application/JSON", "application/json;", "application/json;odata=verbose",
		"application/json;odata=minimalmetadata;streaming=true"}
	var faelle []pfadFall
	for _, namen := range []string{"$format", "%24format"} {
		for _, v := range gueltig {
			faelle = append(faelle, optFall(namen+"="+v, optOk(model.Optionen{})))
		}
		for _, v := range ausserhalbFormate() {
			faelle = append(faelle, optFall(namen+"="+v, kmp(model.KompatFormat, "query/1")))
		}
	}
	faelle = append(faelle,
		optFall("$format=json%20", kmp(model.KompatFormat, "query/1")), optFall("$format=%20json", kmp(model.KompatFormat, "query/1")),
		optFall("$format=", ung(model.URLWert, "query/1")), optFall("$Format=json", ung(model.URLOptionName, "query/1")),
		optFall("$format=json&$format=json", ung(model.URLOptionDoppelt, "query/2")))
	laufe(t, urlparserv2.New(), faelle)
	if len(faelle) != 41 {
		t.Errorf("%d Fälle, erwartet 41", len(faelle))
	}
}

// TestFilterOhneLeser (Zeile 45): ein nichtleerer lesbarer Wert von `$filter` ist ohne Leser Kompatibilität `filter` und wird nicht
// weiter gelesen; leer ist `wert`; ungültig kodiert ist `kodierung`; die letzte Option `WithFilter` gilt. Menge 10 und 2.
func TestFilterOhneLeser(t *testing.T) {
	filter := kmp(model.KompatFilter, "query/1")
	laufe(t, urlparserv2.New(), []pfadFall{
		optFall("$filter=Id%20eq%201", filter), optFall("$filter=(", filter), optFall("$filter=x", filter),
		optFall("$filter=", ung(model.URLWert, "query/1")), optFall("$filter=%FF", ung(model.URLKodierung, "query/1")),
		optFall("$filter=a&$filter=b", ung(model.URLOptionDoppelt, "query/2")), optFall("$top=x&$filter=a", ung(model.URLWert, "query/1")),
	})
	l := &fakeFilterLeser{}
	for name, p := range map[string]*urlparserv2.Parser{
		"New(WithFilter(nil))":                urlparserv2.New(urlparserv2.WithFilter(nil)),
		"New(WithFilter(l), WithFilter(nil))": urlparserv2.New(urlparserv2.WithFilter(l), urlparserv2.WithFilter(nil)),
	} {
		t.Run(name, func(t *testing.T) { pruefe(t, p, "/S", "$filter=a", filter) })
	}
	mit := urlparserv2.New(urlparserv2.WithFilter(nil), urlparserv2.WithFilter(l))
	pruefe(t, mit, "/S", "$filter=a", optOk(model.Optionen{Filter: &model.FilterOption{Text: "a", Ausdruck: model.FilterEigenschaft{Name: "a"}}}))
	if l.n.Load() != 1 {
		t.Errorf("der Leser wurde %d-mal gerufen, erwartet 1 (die letzte Option gilt); der ohne Leser ruft nichts", l.n.Load())
	}
}

// TestGleichwertig (Zeile 46): die Reihenfolge der Systemoptionen, die Schreibweise der Hex-Ziffern, Escape gegen rohes Zeichen,
// `true`/`false`, `asc`/`desc`, leere Segmente und führende Nullen ändern das Ergebnis nicht; die Schreibweise der Namen, die
// Reihenfolge der Elemente, `+` gegen `%20` und die Reihenfolge von `Sonstige` tun es. Menge 136: 120 Permutationen von fünf
// Optionen und 16 Paare.
func TestGleichwertig(t *testing.T) {
	optionen := []string{"$top=5", "$skip=2", "$select=Id,Name", "$orderby=Name%20desc", "$inlinecount=allpages"}
	soll := optOk(model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 5}, Skip: model.ZahlOption{Gesetzt: true, Wert: 2},
		Select:  &model.SelectOption{Namen: []string{"Id", "Name"}},
		OrderBy: []model.Sortierung{{Name: "Name", Absteigend: true}}, Count: model.WahrOption{Gesetzt: true, Wert: true}})
	p := urlparserv2.New()
	n := 0
	var permutiere func(k int, a []string)
	permutiere = func(k int, a []string) {
		if k == len(a) {
			pruefe(t, p, "/S", strings.Join(a, "&"), soll)
			n++
			return
		}
		for i := k; i < len(a); i++ {
			a[k], a[i] = a[i], a[k]
			permutiere(k+1, a)
			a[k], a[i] = a[i], a[k]
		}
	}
	permutiere(0, append([]string(nil), optionen...))
	g := "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE"
	paare := []struct {
		a, qa, b, qb string
		gleich       bool
	}{
		{"/S", "x=%C3%A4", "/S", "x=%c3%a4", true}, {"/S", "%24top=1", "/S", "$top=1", true}, {"/S", "x=%41", "/S", "x=A", true},
		{"/S", "$orderby=A%20asc", "/S", "$orderby=A%20ASC", true}, {"/S", "$orderby=A%20asc", "/S", "$orderby=A", true},
		{"/S", "$top=007", "/S", "$top=7", true}, {"/S", "a=1&&b=2", "/S", "a=1&b=2", true}, {"/S(TRUE)", "", "/S(true)", "", true},
		{"/S(guid'" + g + "')", "", "/S(guid'" + strings.ToLower(g) + "')", "", true}, {"/S(%27a%27)", "", "/S('a')", "", true},
		{"/S(007)", "", "/S(7)", "", true}, {"/S(1.50)", "", "/S(1.5)", "", true},
		{"/S", "$orderby=A,B", "/S", "$orderby=B,A", false}, {"/S", "$orderby=Name+desc", "/S", "$orderby=Name%20desc", false},
		{"/S", "a=1&b=2", "/S", "b=2&a=1", false}, {"/s", "", "/S", "", false},
	}
	for _, f := range paare {
		a, errA := p.ParseURL(f.a, f.qa)
		b, errB := p.ParseURL(f.b, f.qb)
		if f.gleich && (errA != nil || errB != nil || !reflect.DeepEqual(a, b)) {
			t.Errorf("%q %q und %q %q: %+v %v und %+v %v, erwartet gleich", f.a, f.qa, f.b, f.qb, a, errA, b, errB)
		}
		if !f.gleich && errA == nil && errB == nil && reflect.DeepEqual(a, b) {
			t.Errorf("%q %q und %q %q: gleich, erwartet verschieden", f.a, f.qa, f.b, f.qb)
		}
		n++
	}
	if n != 136 {
		t.Errorf("%d Fälle, erwartet 136", n)
	}
}

// TestRangElementzahl (Zeilen 42 und 43): die Zahl der Elemente geht der Pfad- oder Funktionsform vor, wo der Vertrag sie vor dem
// Lesen des Werts prüft. `$select` mit 257 Elementen ist `zu_viele`, auch wenn ein Element ein `/` trägt (Menge 257: das `/` an jeder
// Stelle); `$orderby` mit 33 Elementen und einem `(` oder `/` ist Kompatibilität `option` (Menge 66: 33 Stellen × 2 Formen).
func TestRangElementzahl(t *testing.T) {
	var faelle []pfadFall
	n := 0
	mit := func(namen []string, i int, ersatz string) string {
		k := slices.Clone(namen)
		k[i] = ersatz
		return strings.Join(k, ",")
	}
	s := namenListe(257)
	for i := range s {
		faelle = append(faelle, optFall("$select="+mit(s, i, "A/B"), ung(model.URLZuViele, "query/1")))
		n++
	}
	o := namenListe(33)
	for i := range o {
		for _, form := range []string{"tolower(A)", "A/B"} {
			faelle = append(faelle, optFall("$orderby="+mit(o, i, form), kmp(model.KompatOption, "query/1")))
			n++
		}
	}
	laufe(t, urlparserv2.New(), faelle)
	if n != 257+66 {
		t.Errorf("%d Fälle, erwartet 323", n)
	}
}
