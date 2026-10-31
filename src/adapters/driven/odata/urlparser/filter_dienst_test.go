package urlparser_test

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func flParser() *urlparser.Parser {
	return urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser()))
}

// flURLFehler liest die URL und liefert den URLFehler (nil: gelesen); bei einem Fehler ist das Ergebnis der Nullwert.
func flURLFehler(t *testing.T, p *urlparser.Parser, query string) (model.ODataURL, *model.URLFehler) {
	t.Helper()
	u, err := p.ParseURL("/S", query)
	if err == nil {
		return u, nil
	}
	var uf *model.URLFehler
	if !errors.As(err, &uf) || uf == nil {
		t.Fatalf("%q: %v", query, err)
	}
	if !gleich(u, model.ODataURL{}) {
		t.Errorf("%q: Ergebnis bei Fehler ist nicht der Nullwert", query)
	}
	return u, uf
}

// TestFilterEinhaengung (Zeile 25): `Filter.Text` ist der dekodierte Wert, `Filter.Ausdruck` die Wurzel, die Reihenfolge der
// Optionen ändert nichts; ohne Leser bleibt `$filter` `filter`; die Rangfolge der Klassen gilt zwischen den Optionen. Menge 13.
func TestFilterEinhaengung(t *testing.T) {
	n := 0
	k := flKorpus()
	for _, i := range []int{0, 5, 10, 14, 22, 36} { // sechs Bäume
		u, uf := flURLFehler(t, flParser(), "$filter="+flKodiere(k[i].text))
		if uf != nil || u.Optionen.Filter == nil || u.Optionen.Filter.Text != k[i].text || !gleich(u.Optionen.Filter.Ausdruck, k[i].want) {
			t.Errorf("%q: %+v %v", k[i].text, u.Optionen.Filter, uf)
		}
		n++
	}
	want := bin("eq", eig("Id"), zahl("1"))
	for _, q := range []string{"$top=1&$filter=Id eq 1", "$filter=Id eq 1&$top=1", "$top=1&$filter=Id eq 1&$skip=2"} {
		u, uf := flURLFehler(t, flParser(), q)
		if uf != nil || u.Optionen.Filter == nil || !gleich(u.Optionen.Filter.Ausdruck, want) || u.Optionen.Filter.Text != "Id eq 1" || !u.Optionen.Top.Gesetzt && !u.Optionen.Skip.Gesetzt {
			t.Errorf("%q: %+v %v", q, u.Optionen, uf)
		}
		n++
	}
	for _, p := range []*urlparser.Parser{urlparser.New(), urlparser.New(urlparser.WithFilter(nil))} {
		if _, uf := flURLFehler(t, p, "$filter=Id eq 1"); uf == nil || uf.Ursache != model.URLFilter || uf.Ort != "query/1" {
			t.Errorf("ohne Leser: %v", uf)
		}
		n++
	}
	// Rangfolge der Klassen zwischen den Optionen: ein `filter` des Lesers vor einem `wert` der späteren Option, ein `wert` des
	// Lesers vor einer nicht unterstützten Option
	if _, uf := flURLFehler(t, flParser(), "$filter=foo(1)&$top=x"); uf == nil || uf.Ursache != model.URLWert || uf.Ort != "query/2" {
		t.Errorf("filter vor wert: %v", uf)
	}
	if _, uf := flURLFehler(t, flParser(), "$expand=x&$filter=Id eq"); uf == nil || uf.Ursache != model.URLWert || uf.Ort != "query/2" {
		t.Errorf("wert nach 501: %v", uf)
	}
	n += 2
	if n != 13 {
		t.Fatalf("%d Fälle", n)
	}
}

// flFehlerFaelle sind die acht Ursachen mit je einer Verletzung, in die der Marker m eingesetzt wird.
func flFehlerFaelle() []struct {
	u      model.URLUrsache
	bauen  func(m string) string
	direkt bool // nur bei unmittelbarem Aufruf (`kodierung`)
} {
	return []struct {
		u      model.URLUrsache
		bauen  func(m string) string
		direkt bool
	}{
		{model.URLKodierung, func(m string) string { return m + "\xff" }, true},
		{model.URLWert, func(m string) string { return "Id " + m + " 1" }, false},
		{model.URLLiteral, func(m string) string { return "Id eq '" + m }, false},
		{model.URLZuLang, func(m string) string { return "Id eq '" + m + strings.Repeat("x", 520) + "'" }, false},
		{model.URLZuViele, func(m string) string { return strings.Repeat("(", 33) + m }, false},
		{model.URLFilter, func(m string) string { return m + "(Id)" }, false},
		{model.URLWertNichtUnterstuetzt, func(m string) string { return m + "/" + m }, false},
		{model.URLLiteralTyp, func(m string) string { return m + "'x'" }, false},
	}
}

// TestFilterFehler (Zeile 24): jede Ursache ist ein *URLFehler mit Code und Sentinel ihrer Klasse, das Ergebnis nil, der Ort
// leer; über ParseURL ist der Ort `query/<n>` und das Ergebnis der Nullwert; kein Teil der Eingabe steht im Text. Menge 53 = 8 +
// 21 + 24.
func TestFilterFehler(t *testing.T) {
	const marker = "QQMARKERQQ"
	n := 0
	pruefe := func(uf *model.URLFehler, err error, u model.URLUrsache) {
		t.Helper()
		klasse501 := klasse501(u)
		code := model.CodeE6005
		if klasse501 {
			code = model.CodeE6001
		}
		if uf.Ursache != u || uf.Code() != code || errors.Is(err, model.ErrURLNichtUnterstuetzt) != klasse501 ||
			errors.Is(err, model.ErrURLUngueltig) == klasse501 || strings.Contains(err.Error(), marker) || strings.Contains(uf.Ort, marker) {
			t.Errorf("%s: %v (Ort %q)", u, err, uf.Ort)
		}
	}
	for _, f := range flFehlerFaelle() {
		a, err := flLies(f.bauen("v"))
		var uf *model.URLFehler
		if !errors.As(err, &uf) || a != nil || uf.Ort != "" {
			t.Fatalf("%s: %v %v", f.u, a, err)
		}
		pruefe(uf, err, f.u)
		n++
		if !f.direkt {
			for k := 1; k <= 3; k++ {
				q := strings.Repeat("x=1&", k-1) + "$filter=" + flKodiere(f.bauen("v"))
				_, uf := flURLFehler(t, flParser(), q)
				if uf == nil || uf.Ursache != f.u || uf.Ort != "query/"+strconv.Itoa(k) {
					t.Errorf("%s n=%d: %v", f.u, k, uf)
				}
				n++
			}
		}
		// der Marker als Name, als Literal und in der Verletzung selbst: er steht weder im Fehlertext noch im Ort
		for _, text := range []string{marker + " eq 1 and " + f.bauen("v"), "Id eq '" + marker + "' and " + f.bauen("v"), f.bauen(marker)} {
			_, err := flLies(text)
			var uf *model.URLFehler
			if !errors.As(err, &uf) {
				t.Fatalf("%q: %v", text, err)
			}
			pruefe(uf, err, f.u)
			n++
		}
	}
	if n != 53 {
		t.Fatalf("%d Fälle", n)
	}
}

// flWiederholungskorpus sind die 40 Gutfälle und die 24 Fehlerfälle (64).
func flWiederholungskorpus() []string {
	var out []string
	for _, f := range flKorpus() {
		out = append(out, f.text)
	}
	return append(out, flFehlerKorpus()...)
}

// TestFilterWiederholung (Zeile 28): gleicher Text, gleicher Baum (tiefe Gleichheit) und gleicher Fehler, auf `NewFilterLeser()`
// und `FilterGrammatik{}`. Menge 64 × 2 Aufrufe × 2 Instanzen.
func TestFilterWiederholung(t *testing.T) {
	var nullwert urlparser.FilterGrammatik
	n := 0
	for _, text := range flWiederholungskorpus() {
		a0, e0 := urlparser.NewFilterLeser().LeseFilter(text)
		for _, g := range []*urlparser.FilterGrammatik{urlparser.NewFilterLeser(), &nullwert} {
			for range 2 {
				a, e := g.LeseFilter(text)
				if !gleich(a, a0) || !gleich(e, e0) {
					t.Errorf("%q: %v %v gegen %v %v", text, a, e, a0, e0)
				}
				if a != nil {
					flZaehleKnoten(t, a)
				}
				n++
			}
		}
	}
	if n != 256 {
		t.Fatalf("%d Aufrufe", n)
	}
}

// TestFilterGleichzeitig (Zeilen 29 und 30): 16 Goroutinen × 64 Texte des Korpus × 10 Runden auf einer Instanz, dazu zwei
// verschränkte Ausdrücke (Tiefe 32 gültig, Tiefe 33 `zu_viele`): jeder Aufruf liefert das Ergebnis des einzelnen Aufrufs. Das
// Rennen fängt `make test-race`; dieser Test zählt die Wirkung, ohne auf den Detektor zu warten.
func TestFilterGleichzeitig(t *testing.T) {
	// Die Wirkung braucht mindestens zwei Prozessoren (mit `-test.cpu=1` bliebe ein geteilter Zustand unsichtbar).
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(max(4, runtime.GOMAXPROCS(0))))
	tiefe := func(n int) string { return strings.Repeat("(", n) + "a" + strings.Repeat(")", n) }
	texte := append(flWiederholungskorpus(), tiefe(32), tiefe(33))
	type ergebnis struct {
		a interface{ IstFilterAusdruck() }
		e error
	}
	soll := make([]ergebnis, len(texte))
	for i, s := range texte {
		soll[i].a, soll[i].e = flLies(s)
	}
	g := urlparser.NewFilterLeser()
	var wg sync.WaitGroup
	start := make(chan struct{})
	for w := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for runde := range 10 {
				for k := range texte {
					i := (k + w*7 + runde) % len(texte)
					a, e := g.LeseFilter(texte[i])
					if !gleich(a, soll[i].a) || !gleich(e, soll[i].e) {
						t.Errorf("Goroutine %d: %.40q: %v %v, erwartet %v %v", w, texte[i], a, e, soll[i].a, soll[i].e)
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
}
