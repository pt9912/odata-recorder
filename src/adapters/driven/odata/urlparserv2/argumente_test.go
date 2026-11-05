package urlparserv2_test

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestArgumente (Zeile 1): ein Verstoß gegen die Form der Argumente ist model.ErrUngueltig ohne URLFehler und ohne KompatFehler,
// das Ergebnis der Nullwert. Menge 75: pfad 37 (leer, ohne `/`, `?`, `#`, 33 Steuerzeichen) + rawQuery 33 + 5 Gegenfälle.
func TestArgumente(t *testing.T) {
	p := urlparserv2.New()
	n := 0
	for _, pfad := range []string{"", "x", "/a?b", "/a#b"} {
		pruefe(t, p, pfad, "", aufr())
		n++
	}
	steuer := make([]string, 0, 33)
	for c := 0; c <= 0x1F; c++ {
		steuer = append(steuer, string([]byte{byte(c)}))
	}
	steuer = append(steuer, string([]byte{0x7F}))
	for _, c := range steuer {
		pruefe(t, p, "/a"+c+"b", "", aufr())
		pruefe(t, p, "/S", "a"+c+"b=1", aufr())
		n += 2
	}
	// Gegenfälle: ein rohes `#` oder `?` in rawQuery ist Text, `%3F` und `%23` im Pfad ebenso, und ein rohes Zeichen jenseits von
	// ASCII im Pfad ist keine Formverletzung.
	pruefe(t, p, "/S", "a#b", ung(model.URLOptionName, "query/1"))
	pruefe(t, p, "/S", "a?b", ung(model.URLOptionName, "query/1"))
	pruefe(t, p, "/S%3F", "", ung(model.URLSegment, "pfad/1"))
	pruefe(t, p, "/S%23", "", ung(model.URLSegment, "pfad/1"))
	pruefe(t, p, "/"+zeichen(0xE4), "", okR(mengeR(zeichen(0xE4))))
	n += 5
	if len(steuer) != 33 || n != 75 {
		t.Errorf("%d Steuerzeichen und %d Fälle, erwartet 33 und 75", len(steuer), n)
	}
}

// TestWiederholung (Zeile 2): gleiche Eingabe, gleiches Ergebnis und gleicher Fehler; das Ergebnis gehört dem Aufrufer. Menge 68:
// 30 Eingaben × 2 Aufrufe und 4 Felder × 2 Elemente.
func TestWiederholung(t *testing.T) {
	guidT := "guid'AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE'"
	eingaben := [][2]string{
		{"/$metadata", ""}, {"/S", ""}, {"/S(1)", ""}, {"/S(A=1,B='x')", ""}, {"/S/$count", ""}, {"/S", "$top=1"}, {"/S", "$skip=2"},
		{"/S", "$select=A,B"}, {"/S", "$orderby=A,B%20desc"}, {"/S", "$inlinecount=allpages"}, {"/S", "$format=json"},
		{"/S", "x=1&y=2"}, {"/S", "@a=1"}, {"/S(" + guidT + ")", ""}, {"/S(datetime'2020-01-02T03:04')", ""}, {"/S(42L)", ""},
		{"/S(true)", "$top=0"}, {"/S", "$select=*"}, {"/GetX(1)", "a=b"},
		{"/S(A=1,B=2)", "$select=A,B&$orderby=A,B%20desc&x=1&y=2"},
		{"/S%", ""}, {"/S", "$top=x"}, {"/S(1,,2)", ""}, {"/S", "$foo=1"}, {"/S", "%FF=1"},
		{"/", ""}, {"/S/N", ""}, {"/S", "$expand=1"}, {"/S", "$format=xml"}, {"/S(1d)", ""},
	}
	p := urlparserv2.New()
	n := 0
	for _, e := range eingaben {
		a, errA := p.ParseURL(e[0], e[1])
		b, errB := p.ParseURL(e[0], e[1])
		if !reflect.DeepEqual(a, b) || !reflect.DeepEqual(errA, errB) || (errA == nil) != (a.Ressource.Art != 0) {
			t.Errorf("ParseURL(%q, %q) zweimal: %+v %v und %+v %v", e[0], e[1], a, errA, b, errB)
		}
		n += 2
	}
	want := model.ODataURL{
		Ressource: entityR("S", teil("A", model.LiteralZahl, "1"), teil("B", model.LiteralZahl, "2")),
		Optionen: model.Optionen{Select: &model.SelectOption{Namen: []string{"A", "B"}},
			OrderBy:  []model.Sortierung{{Name: "A"}, {Name: "B", Absteigend: true}},
			Sonstige: []model.SonstigeOption{{Name: "x", Wert: "1"}, {Name: "y", Wert: "2"}}},
	}
	eigentum := []struct {
		feld    string
		aendere func(u *model.ODataURL, i int)
		n       func(u model.ODataURL) int
	}{
		{"Schluessel", func(u *model.ODataURL, i int) { u.Ressource.Schluessel[i].Literal.Text = "X" }, func(u model.ODataURL) int { return len(u.Ressource.Schluessel) }},
		{"Select.Namen", func(u *model.ODataURL, i int) { u.Optionen.Select.Namen[i] = "X" }, func(u model.ODataURL) int { return len(u.Optionen.Select.Namen) }},
		{"OrderBy", func(u *model.ODataURL, i int) { u.Optionen.OrderBy[i].Name = "X" }, func(u model.ODataURL) int { return len(u.Optionen.OrderBy) }},
		{"Sonstige", func(u *model.ODataURL, i int) { u.Optionen.Sonstige[i].Wert = "X" }, func(u model.ODataURL) int { return len(u.Optionen.Sonstige) }},
	}
	for _, f := range eigentum {
		for _, erstes := range []bool{true, false} {
			a, _ := p.ParseURL("/S(A=1,B=2)", "$select=A,B&$orderby=A,B%20desc&x=1&y=2")
			b, _ := p.ParseURL("/S(A=1,B=2)", "$select=A,B&$orderby=A,B%20desc&x=1&y=2")
			i := 0
			if !erstes {
				i = f.n(a) - 1
			}
			f.aendere(&a, i)
			c, _ := p.ParseURL("/S(A=1,B=2)", "$select=A,B&$orderby=A,B%20desc&x=1&y=2")
			if !reflect.DeepEqual(b, want) || !reflect.DeepEqual(c, want) || reflect.DeepEqual(a, want) {
				t.Errorf("%s[%d] geändert: früheres Ergebnis %+v, späteres %+v", f.feld, i, b, c)
			}
			n++
		}
	}
	if len(eingaben) != 30 || n != 60+8 {
		t.Errorf("%d Eingaben, %d Fälle, erwartet 30 und 68", len(eingaben), n)
	}
}

// TestErgebnisForm (Zeile 3): Art je Ressource, Nullwerte nicht genannter Optionen, `$top=0` gesetzt auf Null. Menge 20 Zeilen.
func TestErgebnisForm(t *testing.T) {
	top := func(v int) model.Optionen { return model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: v}} }
	zeilen := []struct {
		pfad, q string
		want    erw
	}{
		{"/$metadata", "", okR(model.Ressource{Art: model.RessourceMetadaten})},
		{"/S", "", okR(mengeR("S"))},
		{"/S(1)", "", okR(entityR("S", zahlT("1")))},
		{"/S/$count", "", okR(anzahlR("S"))},
		{"/S", "$top=0", ok(mengeR("S"), top(0))},
		{"/S", "$top=7", ok(mengeR("S"), top(7))},
		{"/S", "$top=007&$skip=00", ok(mengeR("S"), model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 7}, Skip: model.ZahlOption{Gesetzt: true}})},
		{"/S", "", ok(mengeR("S"), model.Optionen{})},
		{"/S", "&&", ok(mengeR("S"), model.Optionen{})},
		{"/$metadata", "$top=1", ok(model.Ressource{Art: model.RessourceMetadaten}, top(1))},
		{"/S(1)", "$top=1", ok(entityR("S", zahlT("1")), top(1))},
		{"/S('a')", "", okR(entityR("S", teil("", model.LiteralZeichenkette, "'a'")))},
		{"/S(K='a')", "", okR(entityR("S", teil("K", model.LiteralZeichenkette, "'a'")))},
		{"/S(A=1,B=2)", "", okR(entityR("S", teil("A", model.LiteralZahl, "1"), teil("B", model.LiteralZahl, "2")))},
		{"/S(B=2,A=1)", "", okR(entityR("S", teil("B", model.LiteralZahl, "2"), teil("A", model.LiteralZahl, "1")))},
		{"/S", "$top=1&$skip=2&$select=A", ok(mengeR("S"), model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1},
			Skip: model.ZahlOption{Gesetzt: true, Wert: 2}, Select: &model.SelectOption{Namen: []string{"A"}}})},
		{"/S", "$select=A&$skip=2&$top=1", ok(mengeR("S"), model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1},
			Skip: model.ZahlOption{Gesetzt: true, Wert: 2}, Select: &model.SelectOption{Namen: []string{"A"}}})},
		{"/S", "$skip=2&$select=A&$top=1", ok(mengeR("S"), model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1},
			Skip: model.ZahlOption{Gesetzt: true, Wert: 2}, Select: &model.SelectOption{Namen: []string{"A"}}})},
		{"/S", "$inlinecount=none", ok(mengeR("S"), model.Optionen{Count: model.WahrOption{Gesetzt: true}})},
		{"/S", "$orderby=A", ok(mengeR("S"), model.Optionen{OrderBy: []model.Sortierung{{Name: "A"}}})},
	}
	p := urlparserv2.New()
	for _, z := range zeilen {
		pruefe(t, p, z.pfad, z.q, z.want)
	}
	if len(zeilen) != 20 {
		t.Errorf("%d Zeilen, erwartet 20", len(zeilen))
	}
}

// fuellePfad liefert `/S/` und einen qualifizierten Namen aus Gliedern zu höchstens 128 Zeichen, zusammen genau laenge Bytes: ein
// gültiger Pfad (Kompatibilität `pfad`), dessen Länge die Zeile 4 und 5 einstellt.
func fuellePfad(laenge int) string {
	var b strings.Builder
	b.WriteString("/S/")
	n := laenge - 3
	for n > 0 {
		c := min(128, n)
		if n-c == 1 {
			c--
		}
		b.WriteString(strings.Repeat("a", c))
		n -= c
		if n > 0 {
			b.WriteByte('.')
			n--
		}
	}
	return b.String()
}

func namenListe(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("n%d", i)
	}
	return out
}

// TestGrenzen (Zeile 4): jede Grenze von SPEC-160 Nr. 6 ist erlaubt, die nächste Zahl der Verstoß. 13 Grenzen × beide Seiten = 26
// Fälle: pfad 2048, rawQuery 8192, Segmente 16, Segmente der Query 64, Teile 16, Literal 512, Name 128 an fünf Stellen (Set-Name,
// Schlüsselname, Optionsname, Name in `$select`, Name in `$orderby`), Elemente von `$select` 256 und von `$orderby` 32.
func TestGrenzen(t *testing.T) {
	a := func(n int) string { return strings.Repeat("a", n) }
	sonstige := func(n int) []model.SonstigeOption {
		out := make([]model.SonstigeOption, n)
		for i := range out {
			out[i] = model.SonstigeOption{Name: "a", Wert: "1"}
		}
		return out
	}
	teile := func(n int) ([]string, []model.SchluesselTeil) {
		var text []string
		var out []model.SchluesselTeil
		for i := range n {
			text = append(text, fmt.Sprintf("K%d=1", i))
			out = append(out, teil(fmt.Sprintf("K%d", i), model.LiteralZahl, "1"))
		}
		return text, out
	}
	segmente := func(n int) string { return "/S" + strings.Repeat("/N", n-1) }
	var qsegs65 []string
	for range 65 {
		qsegs65 = append(qsegs65, "a=1")
	}
	t16, w16 := teile(16)
	t17, _ := teile(17)
	n256, n257 := namenListe(256), namenListe(257)
	sortierung := func(n int) []model.Sortierung {
		var out []model.Sortierung
		for _, nm := range namenListe(n) {
			out = append(out, model.Sortierung{Name: nm})
		}
		return out
	}
	faelle := []struct {
		name, pfad, q string
		want          erw
	}{
		{"pfad 2048", fuellePfad(2048), "", kmp(model.KompatPfad, "pfad/2")},
		{"pfad 2049", fuellePfad(2049), "", ung(model.URLZuLang, "pfad")},
		{"query 8192", "/S", "x=" + a(8190), ok(mengeR("S"), model.Optionen{Sonstige: []model.SonstigeOption{{Name: "x", Wert: a(8190)}}})},
		{"query 8193", "/S", "x=" + a(8191), ung(model.URLZuLang, "query")},
		{"segmente 16", segmente(16), "", kmp(model.KompatPfad, "pfad/2")},
		{"segmente 17", segmente(17), "", ung(model.URLZuViele, "pfad")},
		{"query-segmente 64", "/S", strings.Join(qsegs65[:64], "&"), ok(mengeR("S"), model.Optionen{Sonstige: sonstige(64)})},
		{"query-segmente 65", "/S", strings.Join(qsegs65, "&"), ung(model.URLZuViele, "query")},
		{"teile 16", "/S(" + strings.Join(t16, ",") + ")", "", okR(entityR("S", w16...))},
		{"teile 17", "/S(" + strings.Join(t17, ",") + ")", "", ung(model.URLZuViele, "pfad/1")},
		{"literal 512", "/S(K='" + a(510) + "')", "", okR(entityR("S", teil("K", model.LiteralZeichenkette, "'"+a(510)+"'")))},
		{"literal 513", "/S(K='" + a(511) + "')", "", ung(model.URLZuLang, "pfad/1")},
		{"set-name 128", "/" + a(128), "", okR(mengeR(a(128)))},
		{"set-name 129", "/" + a(129), "", ung(model.URLSegment, "pfad/1")},
		{"schluessel-name 128", "/S(" + a(128) + "=1)", "", okR(entityR("S", teil(a(128), model.LiteralZahl, "1")))},
		{"schluessel-name 129", "/S(" + a(129) + "=1)", "", ung(model.URLKey, "pfad/1")},
		{"options-name 128", "/S", a(128) + "=1", ok(mengeR("S"), model.Optionen{Sonstige: []model.SonstigeOption{{Name: a(128), Wert: "1"}}})},
		{"options-name 129", "/S", a(129) + "=1", ung(model.URLOptionName, "query/1")},
		{"select-name 128", "/S", "$select=" + a(128), ok(mengeR("S"), model.Optionen{Select: &model.SelectOption{Namen: []string{a(128)}}})},
		{"select-name 129", "/S", "$select=" + a(129), ung(model.URLWert, "query/1")},
		{"orderby-name 128", "/S", "$orderby=" + a(128), ok(mengeR("S"), model.Optionen{OrderBy: []model.Sortierung{{Name: a(128)}}})},
		{"orderby-name 129", "/S", "$orderby=" + a(129), ung(model.URLWert, "query/1")},
		{"select 256", "/S", "$select=" + strings.Join(n256, ","), ok(mengeR("S"), model.Optionen{Select: &model.SelectOption{Namen: n256}})},
		{"select 257", "/S", "$select=" + strings.Join(n257, ","), ung(model.URLZuViele, "query/1")},
		{"orderby 32", "/S", "$orderby=" + strings.Join(namenListe(32), ","), ok(mengeR("S"), model.Optionen{OrderBy: sortierung(32)})},
		{"orderby 33", "/S", "$orderby=" + strings.Join(namenListe(33), ","), ung(model.URLZuViele, "query/1")},
	}
	p := urlparserv2.New()
	for _, f := range faelle {
		pruefe(t, p, f.pfad, f.q, f.want)
	}
	if len(faelle) != 26 {
		t.Errorf("%d Fälle, erwartet 26", len(faelle))
	}
}

// stufe0Pfad und stufe0Query bauen den Pfad bzw. die Query so, dass sie die Verstöße der Stufe 0 tragen oder nicht (SPEC-160 Nr. 5).
func stufe0Pfad(lang, viele, dahinter bool) string {
	segs, laenge := 2, 100
	if viele {
		segs = 20
	}
	if lang {
		laenge = 2100
	}
	erstes := strings.Repeat("a", laenge-1-2*(segs-1))
	if dahinter {
		erstes = "%FF" + erstes[3:]
	}
	return "/" + erstes + strings.Repeat("/a", segs-1)
}

func stufe0Query(lang, viele, dahinter bool) string {
	segs, laenge := 1, 20
	if viele {
		segs, laenge = 65, 4*64+2
	}
	if lang {
		laenge = 9000
	}
	teile := []string{"a=" + strings.Repeat("v", laenge-2-4*(segs-1))}
	for range segs - 1 {
		teile = append(teile, "a=1")
	}
	if dahinter {
		teile[0] = "$expand=" + teile[0][2:]
	}
	return strings.Join(teile, "&")
}

// TestStufeNull (Zeile 5): Gesamtform vor allem anderen, in der Reihenfolge `zu_lang` (pfad), `zu_lang` (query), `zu_viele` (pfad),
// `zu_viele` (query); hinter einem Verstoß der Stufe 0 wird nichts gelesen. Menge 14: 4 einzeln, 6 Paare, 4 mit einem Verstoß der
// Stufe 1 dahinter.
func TestStufeNull(t *testing.T) {
	folge := []struct {
		u   model.URLUrsache
		ort string
	}{{model.URLZuLang, "pfad"}, {model.URLZuLang, "query"}, {model.URLZuViele, "pfad"}, {model.URLZuViele, "query"}}
	p := urlparserv2.New()
	n := 0
	lauf := func(menge [4]bool, dahinter bool) {
		pfad, q := "/S", ""
		if menge[0] || menge[2] {
			pfad = stufe0Pfad(menge[0], menge[2], dahinter)
		}
		if menge[1] || menge[3] {
			q = stufe0Query(menge[1], menge[3], dahinter)
		}
		for i, drin := range menge {
			if drin {
				pruefe(t, p, pfad, q, ung(folge[i].u, folge[i].ort))
				break
			}
		}
		n++
	}
	for i := range 4 {
		var m [4]bool
		m[i] = true
		lauf(m, false)
		lauf(m, true)
		for j := i + 1; j < 4; j++ {
			m[j] = true
			lauf(m, false)
			m[j] = false
		}
	}
	if n != 4+4+6 {
		t.Errorf("%d Fälle, erwartet 14", n)
	}
}

func textsumme(u model.ODataURL) int {
	n := len(u.Ressource.EntitySet)
	for _, s := range u.Ressource.Schluessel {
		n += len(s.Name) + len(s.Literal.Text)
	}
	if u.Optionen.Select != nil {
		for _, s := range u.Optionen.Select.Namen {
			n += len(s)
		}
	}
	for _, s := range u.Optionen.OrderBy {
		n += len(s.Name)
	}
	for _, s := range u.Optionen.Sonstige {
		n += len(s.Name) + len(s.Wert)
	}
	if u.Optionen.Filter != nil {
		n += len(u.Optionen.Filter.Text)
	}
	return n
}

// TestWaechter (Zeile 6): die Summe der Längen der Texte im Ergebnis ist höchstens Länge(pfad) + Länge(rawQuery) + 48, und die
// Zuteilung eines Aufrufs höchstens 64 · (Länge(pfad) + Länge(rawQuery)) + 65536 Bytes. Menge 29: sechs Zeichenfamilien × vier
// Formen (pfad an der Grenze, rawQuery an der Grenze, beide, ein Pfad aus lauter Familie), dazu 16 benannte Teile von einem
// Zeichen, ein Zeitpunkt ohne Sekunden und drei große Erfolge.
func TestWaechter(t *testing.T) {
	p := urlparserv2.New()
	type eingabe struct{ pfad, q string }
	var alle []eingabe
	for _, fam := range []string{"'", "%", ",", "(", "&", "/"} {
		pf := "/S(" + strings.Repeat(fam, 2044) + ")"
		q := "x=" + strings.Repeat(fam, 8190)
		alle = append(alle, eingabe{pf, ""}, eingabe{"/S", q}, eingabe{pf, q}, eingabe{"/" + strings.Repeat(fam, 2047), ""})
	}
	var teile16 []string
	for _, c := range "abcdefghijklmnop" {
		teile16 = append(teile16, string(c)+"=1")
	}
	gross := strings.Repeat("a", 500)
	alle = append(alle,
		eingabe{"/S(" + strings.Join(teile16, ",") + ")", ""},
		eingabe{"/S(datetime'2020-01-02T03:04')", ""},
		eingabe{"/S(A='" + gross + "',B='" + gross + "',C='" + gross + "')", "x=" + strings.Repeat("v", 8190)},
		eingabe{"/S('" + strings.Repeat("''", 250) + "')", "$select=" + strings.Repeat("a,", 200) + "b"},
		eingabe{"/S", "$orderby=" + strings.Repeat("a,", 20) + "b%20desc&x=" + strings.Repeat("%41", 2000)},
	)
	if len(alle) != 6*4+5 {
		t.Fatalf("%d Eingaben", len(alle))
	}
	var vorher, nachher runtime.MemStats
	for _, e := range alle {
		runtime.ReadMemStats(&vorher)
		u, err := p.ParseURL(e.pfad, e.q)
		runtime.ReadMemStats(&nachher)
		laenge := len(e.pfad) + len(e.q)
		if zugeteilt := nachher.TotalAlloc - vorher.TotalAlloc; zugeteilt > uint64(64*laenge+65536) {
			t.Errorf("ParseURL(%.20q, %.20q): %d Bytes zugeteilt, erlaubt %d", e.pfad, e.q, zugeteilt, 64*laenge+65536)
		}
		if err == nil && textsumme(u) > laenge+48 {
			t.Errorf("ParseURL(%.20q, %.20q): Texte %d Bytes, erlaubt %d", e.pfad, e.q, textsumme(u), laenge+48)
		}
	}
}
