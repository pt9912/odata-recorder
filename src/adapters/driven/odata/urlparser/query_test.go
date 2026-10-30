package urlparser_test

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

const (
	wertU = model.URLWert
	nameU = model.URLOptionName
	doppU = model.URLOptionDoppelt
	optNU = model.URLOptionNichtUnterstuetzt
	wertN = model.URLWertNichtUnterstuetzt
)

// q liest die Query hinter dem Pfad `/S` und verlangt, dass sie gelesen wird.
func q(t *testing.T, query string) model.Optionen {
	t.Helper()
	return gelesen(t, "/S", query).Optionen
}

// qf verlangt für die Query einen Verstoß an der Option 1.
func qf(t *testing.T, query string, u model.URLUrsache) {
	t.Helper()
	pruefeFall(t, parserNeu(), fall{pfad: "/S", query: query, u: u, ort: "query/1"})
}

func namenBehandelt() []string {
	return []string{"$select", "$orderby", "$top", "$skip", "$count", "$filter", "$format"}
}

func namenNichtUnterstuetzt() []string {
	return []string{"$expand", "$search", "$apply", "$compute", "$levels", "$schemaversion", "$deltatoken", "$skiptoken", "$index", "$id"}
}

// TestOptionsklassen (Zeile 46): 37 Namen in ihren Klassen.
func TestOptionsklassen(t *testing.T) {
	werte := map[string]string{"$select": "Id", "$orderby": "Id", "$top": "1", "$skip": "1", "$count": "true", "$filter": "x", "$format": "json"}
	n := 0
	for _, name := range namenBehandelt() {
		if name == "$filter" {
			qf(t, name+"="+werte[name], model.URLFilter)
		} else {
			q(t, name+"="+werte[name])
		}
		n++
	}
	for _, name := range namenNichtUnterstuetzt() {
		qf(t, name+"=v", optNU)
		n++
	}
	for _, name := range []string{"$Top", "$TOP", "$foo", "$", "$top%20"} {
		qf(t, name+"=1", nameU)
		n++
	}
	for _, name := range []string{"@p", "@a1", "sap-client", "a.b", "_x", wiederhole("a", 128)} {
		if got := q(t, name+"=v").Sonstige; len(got) != 1 || got[0].Name != name {
			t.Errorf("%q: %+v", name, got)
		}
		n++
	}
	for _, name := range []string{"@", "@1", "@a-b", "", "1a", "a%20b", "%C3%A4", wiederhole("a", 129), "?$top"} {
		qf(t, name+"=v", nameU)
		n++
	}
	if n != 37 {
		t.Fatalf("%d Namen, erwartet 37", n)
	}
}

// TestOptionDoppelt (Zeile 47): jede der 7 behandelten Optionen zweimal ist `option_doppelt`, roh und mit Escape, bei gleichem
// und anderem Wert (28 Fälle); nicht unterstützte Namen und Sonstige zweimal sind erlaubt (2 Fälle).
func TestOptionDoppelt(t *testing.T) {
	n := 0
	erster := map[string]string{"$select": "Id", "$orderby": "Id", "$top": "1", "$skip": "1", "$count": "true", "$filter": "x", "$format": "json"}
	for _, name := range namenBehandelt() {
		for _, zweiter := range []string{name, "%24" + name[1:]} {
			for _, wert2 := range []string{erster[name], "%G1"} {
				pruefeFall(t, parserNeu(), fall{pfad: "/S", query: name + "=" + erster[name] + "&" + zweiter + "=" + wert2, u: doppU, ort: "query/2"})
				n++
			}
		}
	}
	qf(t, "$expand=a&$expand=b", optNU)
	if got := q(t, "x=1&x=2").Sonstige; len(got) != 2 {
		t.Errorf("Sonstige zweimal: %+v", got)
	}
	if n != 28 {
		t.Fatalf("%d Fälle, erwartet 28", n)
	}
}

// TestTopSkip (Zeile 48): `$top` und `$skip`, 21 Werte je Option, 5 gültige und 16 ungültige.
func TestTopSkip(t *testing.T) {
	gueltig := map[string]int{"0": 0, "1": 1, "007": 7, "00": 0, "2147483647": 2147483647}
	ungueltig := []string{"", "-1", "-0", "+1", "1.0", "1e3", "0x10", "%201", "1%20", "\xef\xbc\x91", "\xd9\xa1", "2147483648", "99999999999999999999",
		"18446744073709551616", "18446744073709551621"}
	n := 0
	for _, name := range []string{"$top", "$skip"} {
		for wert, soll := range gueltig {
			o := q(t, name+"="+wert)
			ist, anderes := o.Top, o.Skip
			if name == "$skip" {
				ist, anderes = o.Skip, o.Top
			}
			if ist != (model.ZahlOption{Gesetzt: true, Wert: soll}) || anderes.Gesetzt {
				t.Errorf("%s=%s: %+v", name, wert, o)
			}
			n++
		}
		for _, wert := range ungueltig {
			qf(t, name+"="+wert, wertU)
			n++
		}
		qf(t, name, wertU)
		n++
	}
	if n != 42 {
		t.Fatalf("%d Fälle, erwartet 42", n)
	}
}

// TestTopSkipBereich (Zeile 48, Klasse „Ziffernfolge über dem Bereich“, SPEC-162 Nr. 4: der Wert liegt von 0 bis 2147483647, alles
// darüber ist `wert`, gleich wie lang die Ziffernfolge ist). Menge: die Werte 10^(n-1) und 10^n − 1 für n von 1 bis 40, die
// Grenzen 2^31 − 1, 2^31, 2^63 − 1, 2^63, 2^63 + 5, 2^64 − 1, 2^64, 2^64 + 5 und 9999999999999999999, jeder mit 0 bis 3 führenden
// Nullen, für `$top` und `$skip`. Das Orakel ist math/big; ein Wert über der Grenze ergibt `wert` und nie ein gesetztes Feld,
// ein Wert darunter genau seine Zahl.
func TestTopSkipBereich(t *testing.T) {
	werte := map[string]bool{}
	zehn := big.NewInt(10)
	for n := 1; n <= 40; n++ {
		p := new(big.Int).Exp(zehn, big.NewInt(int64(n-1)), nil)
		werte[p.String()] = true
		werte[new(big.Int).Sub(new(big.Int).Mul(p, zehn), big.NewInt(1)).String()] = true
	}
	zwei := func(e uint) *big.Int { return new(big.Int).Lsh(big.NewInt(1), e) }
	for _, g := range []*big.Int{zwei(31), zwei(63), zwei(64)} {
		for _, d := range []int64{-1, 0, 5} {
			werte[new(big.Int).Add(g, big.NewInt(d)).String()] = true
		}
	}
	werte["9999999999999999999"] = true
	grenze, n := big.NewInt(2147483647), 0
	for wert := range werte {
		v, _ := new(big.Int).SetString(wert, 10)
		for nullen := range 4 {
			text := strings.Repeat("0", nullen) + wert
			for _, name := range []string{"$top", "$skip"} {
				if v.Cmp(grenze) > 0 {
					qf(t, name+"="+text, wertU)
				} else {
					o := q(t, name+"="+text)
					ist := o.Top
					if name == "$skip" {
						ist = o.Skip
					}
					if ist != (model.ZahlOption{Gesetzt: true, Wert: int(v.Int64())}) {
						t.Errorf("%s=%s: %+v", name, text, ist)
					}
				}
				n++
			}
		}
	}
	if n != len(werte)*8 || len(werte) < 80 {
		t.Fatalf("%d Fälle, %d Werte", n, len(werte))
	}
}

// TestGesetzt (Zeile 49): „auf Null gesetzt“ und „nicht gesetzt“ sind unterscheidbar; 6 Fälle.
func TestGesetzt(t *testing.T) {
	gesetzt := func(o model.Optionen) [3]bool { return [3]bool{o.Top.Gesetzt, o.Skip.Gesetzt, o.Count.Gesetzt} }
	for i, f := range [][2]string{{"$top", "0"}, {"$skip", "0"}, {"$count", "false"}} {
		var soll [3]bool
		if got := gesetzt(q(t, "x=1")); got != soll {
			t.Errorf("%s fehlt, gesetzt: %v", f[0], got)
		}
		soll[i] = true
		o := q(t, f[0]+"="+f[1])
		if got := gesetzt(o); got != soll || o.Top.Wert != 0 || o.Skip.Wert != 0 || o.Count.Wert {
			t.Errorf("%s=%s: %+v", f[0], f[1], o)
		}
	}
}

// TestCount (Zeile 50): `true` in 16, `false` in 32 Schreibweisen und 8 ungültige Werte; 56 Fälle.
func TestCount(t *testing.T) {
	n := 0
	for _, wort := range []string{"true", "false"} {
		for maske := range 1 << len(wort) {
			b := []byte(wort)
			for i := range b {
				if maske>>i&1 == 1 {
					b[i] -= 'a' - 'A'
				}
			}
			if got := q(t, "$count="+string(b)).Count; got != (model.WahrOption{Gesetzt: true, Wert: wort == "true"}) {
				t.Errorf("%s: %+v", b, got)
			}
			n++
		}
	}
	for _, w := range []string{"1", "0", "yes", "", "%20true", "true%20", "truee"} {
		qf(t, "$count="+w, wertU)
		n++
	}
	qf(t, "$count", wertU)
	n++
	if n != 56 {
		t.Fatalf("%d Fälle, erwartet 56", n)
	}
}

// TestSelect (Zeile 51): `$select` mit 7 gültigen, 12 ungültigen und 3 nicht unterstützten Werten.
func TestSelect(t *testing.T) {
	namen128, viele := wiederhole("n", 128), make([]string, 256)
	for i := range viele {
		viele[i] = "e" + strconv.Itoa(i)
	}
	gueltig := map[string]model.SelectOption{
		"*": {Alle: true}, "Id": {Namen: []string{"Id"}}, "Id,Name": {Namen: []string{"Id", "Name"}},
		"*,Name": {Alle: true, Namen: []string{"Name"}}, "Name,*": {Alle: true, Namen: []string{"Name"}},
		namen128: {Namen: []string{namen128}}, strings.Join(viele, ","): {Namen: viele},
	}
	for wert, soll := range gueltig {
		got := q(t, "$select="+wert).Select
		if got == nil || got.Alle != soll.Alle || !gleich(got.Namen, soll.Namen) {
			t.Errorf("%.20s: %+v", wert, got)
		}
	}
	ungueltig := []string{"", "Id,", ",Id", "Id,,Name", "Id,%20Name", "Id%20", "Id,Id", "*,*", "(", "$x", "a-b", wiederhole("n", 129)}
	for _, w := range ungueltig {
		qf(t, "$select="+w, wertU)
	}
	for _, w := range []string{"a/b", "a/b/c", "Id,a/b"} {
		qf(t, "$select="+w, wertN)
	}
	if len(gueltig)+len(ungueltig)+3 != 22 {
		t.Fatal("Zahl der Fälle")
	}
}

// TestOrderby (Zeile 52): `$orderby` mit 31 gültigen, 13 ungültigen und 5 nicht unterstützten Werten.
func TestOrderby(t *testing.T) {
	n := 0
	for _, f := range []struct {
		wert string
		soll []model.Sortierung
	}{
		{"Name", []model.Sortierung{{Name: "Name"}}}, {"Name%20asc", []model.Sortierung{{Name: "Name"}}},
		{"Name%20desc", []model.Sortierung{{Name: "Name", Absteigend: true}}}, {"Name%20DESC", []model.Sortierung{{Name: "Name", Absteigend: true}}},
		{"a,b%20desc,c", []model.Sortierung{{Name: "a"}, {Name: "b", Absteigend: true}, {Name: "c"}}},
		{"c%20desc,b,a%20asc", []model.Sortierung{{Name: "c", Absteigend: true}, {Name: "b"}, {Name: "a"}}},
		{"x,y", []model.Sortierung{{Name: "x"}, {Name: "y"}}},
	} {
		if got := q(t, "$orderby="+f.wert).OrderBy; !gleich(got, f.soll) {
			t.Errorf("%s: %+v", f.wert, got)
		}
		n++
	}
	for _, wort := range []string{"asc", "desc"} {
		for maske := range 1 << len(wort) {
			b := []byte(wort)
			for i := range b {
				if maske>>i&1 == 1 {
					b[i] -= 'a' - 'A'
				}
			}
			if got := q(t, "$orderby=N%20"+string(b)).OrderBy; len(got) != 1 || got[0].Absteigend != (wort == "desc") {
				t.Errorf("%s: %+v", b, got)
			}
			n++
		}
	}
	for _, w := range []string{"", "a,", ",a", "a,,b", "a%20%20desc", "a%20up", "a%20asc%20desc", "a,a", "a%20asc,a%20desc", "a%20desc,a%20desc",
		"Price%20mul%202", "a%20b", "a+desc"} {
		qf(t, "$orderby="+w, wertU)
		n++
	}
	for _, w := range []string{"a/b", "a/b%20desc", "length(a)", "concat(a,b)", "("} {
		qf(t, "$orderby="+w, wertN)
		n++
	}
	if n != 49 {
		t.Fatalf("%d Fälle, erwartet 49", n)
	}
}

// TestFormat (Zeile 53): 6 lesbare, 5 nicht unterstützte Werte und der leere.
func TestFormat(t *testing.T) {
	for _, w := range []string{"json", "JSON", "application/json", "Application/Json", "application/json;odata.metadata=none", "application/json;"} {
		if o := q(t, "$format="+w); !gleich(o, model.Optionen{}) {
			t.Errorf("%s: %+v", w, o)
		}
	}
	for _, w := range []string{"xml", "atom", "application/jsonx", "text/json", "json%20"} {
		qf(t, "$format="+w, optNU)
	}
	qf(t, "$format=", wertU)
}

// TestSonstige (Zeile 54): Optionen ohne `$` in der Reihenfolge der Query, dekodiert, `x` und `x=` gleich; 6 Fälle.
func TestSonstige(t *testing.T) {
	so := func(n, w string) model.SonstigeOption { return model.SonstigeOption{Name: n, Wert: w} }
	for query, soll := range map[string][]model.SonstigeOption{
		"x=1&x=2": {so("x", "1"), so("x", "2")}, "x": {so("x", "")}, "x=": {so("x", "")}, "@p=1": {so("@p", "1")},
		"sap-client=100": {so("sap-client", "100")}, "x=%41": {so("x", "A")}, "sap%2Dclient=100": {so("sap-client", "100")}, "b=1&a=2&c": {so("b", "1"), so("a", "2"), so("c", "")},
	} {
		if got := q(t, query).Sonstige; !gleich(got, soll) {
			t.Errorf("%q: %+v", query, got)
		}
	}
}

// TestWertUngelesen (Zeile 55): der Wert einer nicht unterstützten Option wird nicht gelesen, auch nicht dekodiert; 10 Fälle.
func TestWertUngelesen(t *testing.T) {
	for _, name := range namenNichtUnterstuetzt() {
		qf(t, name+"=%G1", optNU)
	}
}

// TestAnwendbarkeitNichtBeurteilt (Zeile 58): der Parser beurteilt nicht, welche Option zu welcher Art der Ressource gehört: jede
// der fünf Optionen `$select`, `$orderby`, `$top`, `$skip`, `$count` hinter jeder der vier Arten ist lesbar. Menge 20.
func TestAnwendbarkeitNichtBeurteilt(t *testing.T) {
	n := 0
	for _, pfad := range []string{"/$metadata", "/Set", "/Set(1)", "/Set/$count"} {
		for _, query := range []string{"$select=Id", "$orderby=Id", "$top=1", "$skip=1", "$count=true"} {
			gelesen(t, pfad, query)
			n++
		}
	}
	if n != 20 {
		t.Fatalf("%d Fälle, erwartet 20", n)
	}
}
