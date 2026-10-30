package urlparser_test

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestArgumenteFormverstoss (Zeile 1): Formverstöße der Argumente sind `model.ErrUngueltig` ohne `URLFehler`. Menge 204:
// 4 Formen des Pfads, 33 Steuerzeichen an drei Stellen von `pfad` (99) und von `rawQuery` (99), 2 Gegenfälle.
func TestArgumenteFormverstoss(t *testing.T) {
	var pfade, queries []string
	pfade = append(pfade, "", "Set", "/Set?a", "/Set#a")
	n := len(pfade)
	for c := 0; c <= 0x20; c++ {
		if c == 0x20 {
			c = 0x7f
		}
		s := string([]byte{byte(c)})
		pfade = append(pfade, "/"+s+"Set", "/Se"+s+"t", "/Set"+s)
		queries = append(queries, s+"x=1", "x"+s+"=1", "x=1"+s)
	}
	faelle := make([][2]string, 0, 204)
	for _, p := range pfade {
		faelle = append(faelle, [2]string{p, ""})
	}
	for _, q := range queries {
		faelle = append(faelle, [2]string{"/Set", q})
	}
	if len(pfade) != n+99 || len(queries) != 99 {
		t.Fatalf("%d Pfade, %d Queries", len(pfade), len(queries))
	}
	for _, f := range faelle {
		u, err := parse(f[0], f[1])
		var uf *model.URLFehler
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &uf) || !gleich(u, model.ODataURL{}) {
			t.Errorf("%q %q: %v", f[0], f[1], err)
		}
	}
	// Gegenfälle: ein rohes `?` und `#` in `rawQuery` ist Text.
	for _, q := range []string{"x=?a", "x=a#b"} {
		if got := gelesen(t, "/Set", q).Optionen.Sonstige; len(got) != 1 || got[0].Wert != q[2:] {
			t.Errorf("%q: %+v", q, got)
		}
	}
	if len(faelle)+2 != 204 {
		t.Fatalf("%d Fälle, erwartet 204", len(faelle)+2)
	}
}

// pfadLaenge baut einen gültigen Pfad `/S(P1='…',P2='…')` von genau n Bytes, n von 40 bis 2400.
func pfadLaenge(n int) string {
	rest := n - len("/S()")
	k := rest/480 + 1
	frei := rest - (k - 1)
	teile := make([]string, k)
	for i := range teile {
		kopf := "P" + strconv.Itoa(i+1) + "='"
		anteil := frei / (k - i)
		frei -= anteil
		teile[i] = kopf + wiederhole("a", anteil-len(kopf)-1) + "'"
	}
	return "/S(" + strings.Join(teile, ",") + ")"
}

// TestGrenzenLaenge (Zeile 2): die Länge von `pfad` (2048) und `rawQuery` (8192) auf beiden Seiten der Grenze.
func TestGrenzenLaenge(t *testing.T) {
	for _, n := range []int{2047, 2048} {
		if p := pfadLaenge(n); len(p) != n {
			t.Fatalf("pfadLaenge(%d) = %d", n, len(p))
		}
		gelesen(t, pfadLaenge(n), "")
	}
	pruefeFaelle(t, []fall{
		{pfad: pfadLaenge(2049), u: model.URLZuLang, ort: "pfad"},
		{pfad: "/S", query: "x=" + wiederhole("a", 8190)},
		{pfad: "/S", query: "x=" + wiederhole("a", 8189)},
		{pfad: "/S", query: "x=" + wiederhole("a", 8191), u: model.URLZuLang, ort: "query"},
	})
}

// TestGrenzenAnzahl (Zeile 3): sieben Grenzen, je auf der Grenze (gelesen) und um eins darüber; 14 Fälle.
func TestGrenzenAnzahl(t *testing.T) {
	namen := func(praefix string, n int) string {
		s := make([]string, n)
		for i := range s {
			s[i] = praefix + strconv.Itoa(i)
		}
		return strings.Join(s, ",")
	}
	zahlen := func(n int) string {
		s := make([]string, n)
		for i := range s {
			s[i] = "x" + strconv.Itoa(i) + "=1"
		}
		return strings.Join(s, "&")
	}
	grenzen := []struct {
		name  string
		bau   func(n int) (pfad, query string)
		n     int
		lesen model.URLUrsache // Ursache auf der Grenze; leer: gelesen
		ort   string           // Ort des Verstoßes darüber
		ueber model.URLUrsache
	}{
		{"Segmente des Pfads", func(n int) (string, string) { return "/S" + wiederhole("/a", n-1), "" }, 16, pnu, "pfad", model.URLZuViele},
		{"Segmente der Query", func(n int) (string, string) { return "/S", zahlen(n) }, 64, "", "query", model.URLZuViele},
		{"Teile eines Schlüssels", func(n int) (string, string) {
			s := make([]string, n)
			for i := range s {
				s[i] = "P" + strconv.Itoa(i) + "=1"
			}
			return "/S(" + strings.Join(s, ",") + ")", ""
		}, 16, "", "pfad/1", model.URLZuViele},
		{"Länge eines Literals", func(n int) (string, string) { return "/S('" + wiederhole("a", n-2) + "')", "" }, 512, "", "pfad/1", model.URLZuLang},
		{"Elemente von $select", func(n int) (string, string) { return "/S", "$select=" + namen("e", n) }, 256, "", "query/1", model.URLZuViele},
		{"Elemente von $orderby", func(n int) (string, string) { return "/S", "$orderby=" + namen("e", n) }, 32, "", "query/1", model.URLZuViele},
		{"Länge eines Namens", func(n int) (string, string) { return "/" + wiederhole("a", n), "" }, 128, "", "pfad/1", segm},
	}
	if len(grenzen) != 7 {
		t.Fatal("sieben Grenzen")
	}
	for _, g := range grenzen {
		p, q := g.bau(g.n)
		pruefeFall(t, parserNeu(), fall{pfad: p, query: q, u: g.lesen, ort: "pfad/2"})
		p, q = g.bau(g.n + 1)
		pruefeFall(t, parserNeu(), fall{pfad: p, query: q, u: g.ueber, ort: g.ort})
	}
	// Name eines Schlüsselteils, einer Option und im $select auf der Grenze von 128 Zeichen und darüber.
	gelesen(t, "/S("+wiederhole("k", 128)+"=1)", "")
	gelesen(t, "/S", wiederhole("o", 128)+"=1&$select="+wiederhole("s", 128))
	pruefeFaelle(t, []fall{
		{pfad: "/S(" + wiederhole("k", 129) + "=1)", u: keyU, ort: "pfad/1"},
		{pfad: "/S", query: wiederhole("o", 129) + "=1", u: nameU, ort: "query/1"},
		{pfad: "/S", query: "$select=" + wiederhole("s", 129), u: wertU, ort: "query/1"},
	})
}

func gesamtZuteilung(f func()) uint64 {
	var a, b runtime.MemStats
	runtime.ReadMemStats(&a)
	f()
	runtime.ReadMemStats(&b)
	return b.TotalAlloc - a.TotalAlloc
}

// textsumme ist die Summe der Längen aller Texte des Ergebnisses.
func textsumme(u model.ODataURL) int {
	n := len(u.Ressource.EntitySet)
	for _, k := range u.Ressource.Schluessel {
		n += len(k.Name) + len(k.Literal.Text)
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

// TestWaechterZuteilung (Zeile 16): sechs Zeichenfamilien in `pfad` und in `rawQuery`, jeweils auf der Grenze: die Gesamtzuteilung
// bleibt unter 64 · n + 65536 Bytes, die Textsumme des Ergebnisses unter n. Menge 12.
func TestWaechterZuteilung(t *testing.T) {
	// vier Teile mit je einem Literal von rund 440 Bytes: der Pfad bleibt unter den Grenzen von Literal und Pfad
	teile := func(inhalt string) string {
		return "/S(P1='" + inhalt + "',P2='" + inhalt + "',P3='" + inhalt + "',P4='" + inhalt + "')"
	}
	familien := []struct {
		name        string
		pfad, query string
		pfadAuchOK  bool
		queryAuchOK bool
	}{
		{"'", teile(wiederhole("''", 220)), "x='" + wiederhole("'", 8189), true, true},
		{"%", teile(wiederhole("%41", 140)), "x=" + wiederhole("%41", 2730), true, true},
		{",", "/S(" + wiederhole(",", 2040) + ")", "$select=" + wiederhole(",", 8184), false, false},
		{"(", "/S(" + wiederhole("(", 2040) + ")", "$orderby=" + wiederhole("(", 8183), false, false},
		{"&", "/S", wiederhole("&", 8192), true, true},
		{"/", "/" + wiederhole("/", 2047), "$select=" + wiederhole("/", 8184), false, false},
	}
	n := 0
	for _, f := range familien {
		for i, args := range [][2]string{{f.pfad, ""}, {"/S", f.query}} {
			laenge := len(args[0]) + len(args[1])
			var u model.ODataURL
			var err error
			z := gesamtZuteilung(func() { u, err = parse(args[0], args[1]) })
			if grenze := uint64(64*laenge + 65536); z > grenze {
				t.Errorf("Familie %s Argument %d: Zuteilung %d über %d", f.name, i, z, grenze)
			}
			if err == nil && textsumme(u) > laenge {
				t.Errorf("Familie %s Argument %d: Textsumme %d über %d", f.name, i, textsumme(u), laenge)
			}
			if i == 0 && f.pfadAuchOK && err != nil || i == 1 && f.queryAuchOK && err != nil {
				t.Errorf("Familie %s Argument %d: %v", f.name, i, err)
			}
			n++
		}
	}
	if n != 12 {
		t.Fatalf("%d Fälle, erwartet 12", n)
	}
}

// TestUeberlangeEingabe (Zeile 17): eine Eingabe über der Grenze endet mit `zu_lang`, ohne gelesen zu werden: höchstens
// 4096 Bytes Zuteilung für 1 MiB. Menge 4: `pfad` und `rawQuery`, je einmal aus Escapes (kostet die Dekodierung) und einmal aus
// Segmenten (kostet die Zerlegung).
func TestUeberlangeEingabe(t *testing.T) {
	gross := 1 << 20
	for _, f := range []fall{
		{pfad: "/S('" + wiederhole("%41", gross/3) + "')", u: model.URLZuLang, ort: "pfad"},
		{pfad: "/S" + wiederhole("/a", gross/2), u: model.URLZuLang, ort: "pfad"},
		{pfad: "/S", query: "x=" + wiederhole("%41", gross/3), u: model.URLZuLang, ort: "query"},
		{pfad: "/S", query: "x=1" + wiederhole("&a", gross/2), u: model.URLZuLang, ort: "query"},
	} {
		p := parserNeu()
		z := gesamtZuteilung(func() { _, _ = p.ParseURL(f.pfad, f.query) })
		if z > 4096 {
			t.Errorf("%s: Zuteilung %d", f.ort, z)
		}
		pruefeFall(t, p, f)
	}
}
