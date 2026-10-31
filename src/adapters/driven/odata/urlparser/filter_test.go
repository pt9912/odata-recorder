package urlparser_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

type testAusdruck struct{ text string }

func (testAusdruck) IstFilterAusdruck() {}

// zaehlLeser ist der Zähl-Fake des FilterLeser: er hält jedes Argument fest und zählt jeden Aufruf.
type zaehlLeser struct {
	mu       sync.Mutex
	texte    []string
	ausdruck model.FilterAusdruck
	fehler   error
	aufrufe  int
	beides   bool // liefert den Ausdruck auch dann, wenn er einen Fehler liefert
}

func (z *zaehlLeser) LeseFilter(text string) (interface{ IstFilterAusdruck() }, error) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.aufrufe++
	z.texte = append(z.texte, text)
	if z.fehler != nil {
		if z.beides {
			return z.ausdruck, z.fehler
		}
		return nil, z.fehler
	}
	return z.ausdruck, nil
}

// nilAusdruck hat einen Zeiger als Empfänger; ein typisierter nil-Zeiger davon ist ein Ausdruck ungleich nil.
type nilAusdruck struct{}

func (*nilAusdruck) IstFilterAusdruck() {}

// TestFilterLeserRueckgaben (Zeile 19, Randformen des Lesers, SPEC-160 Nr. 8 b bis e): Form des Fehlers × Verpackung × mit und
// ohne Ausdruck zugleich. Menge: 3 ungültige Formen (fremder Fehler, `*URLFehler` mit fremder Ursache, `(*URLFehler)(nil)`)
// und 2 gültige (400, 501) × 2 Verpackungen (unverpackt, `%w`) × 2 (ohne und mit Ausdruck) = 20, dazu `nil` und `nil` und
// ein Ausdruck mit typisiertem nil-Zeiger, 22 Fälle. Jeder ungültige Fall ist `model.ErrUngueltig` ohne `URLFehler` und der
// Nullwert, jeder gültige der `URLFehler` mit Ursache und Ort der Option; der Leser wird genau einmal gerufen, keine Panik.
func TestFilterLeserRueckgaben(t *testing.T) {
	var nilFehler *model.URLFehler
	formen := []struct {
		name    string
		fehler  error
		gueltig model.URLUrsache // leer: Fehler des Aufrufers
	}{
		{"fremder Fehler", errors.New("kaputt"), ""},
		{"fremde Ursache", &model.URLFehler{Ursache: "gibt_es_nicht"}, ""},
		{"typisierter nil-Zeiger", nilFehler, ""},
		{"400", &model.URLFehler{Ursache: model.URLWert, Ort: "irgendwo"}, model.URLWert},
		{"501", &model.URLFehler{Ursache: model.URLFilter, Ort: "irgendwo"}, model.URLFilter},
	}
	n := 0
	for _, f := range formen {
		for _, verpackt := range []bool{false, true} {
			for _, beides := range []bool{false, true} {
				fehler := f.fehler
				if verpackt {
					fehler = fmt.Errorf("lesen: %w", f.fehler)
				}
				l := &zaehlLeser{ausdruck: testAusdruck{"ok"}, fehler: fehler, beides: beides}
				pruefeLeserFall(t, fmt.Sprintf("%s verpackt=%v beides=%v", f.name, verpackt, beides), l, f.gueltig)
				n++
			}
		}
	}
	pruefeLeserFall(t, "nil und nil", &zaehlLeser{}, "")
	u, err := urlparser.New(urlparser.WithFilter(&zaehlLeser{ausdruck: (*nilAusdruck)(nil)})).ParseURL("/S", "$filter=a")
	if err != nil || u.Optionen.Filter == nil || u.Optionen.Filter.Ausdruck != model.FilterAusdruck((*nilAusdruck)(nil)) {
		t.Errorf("typisierter nil-Ausdruck: %+v %v", u.Optionen.Filter, err)
	}
	if n != 20 {
		t.Fatalf("%d Fälle", n)
	}
}

func pruefeLeserFall(t *testing.T, name string, l *zaehlLeser, gueltig model.URLUrsache) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: Panik %v", name, r)
		}
	}()
	u, err := urlparser.New(urlparser.WithFilter(l)).ParseURL("/S", "x=1&$filter=a")
	var uf *model.URLFehler
	switch {
	case gueltig == "":
		if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &uf) {
			t.Errorf("%s: %v", name, err)
		}
	case !errors.As(err, &uf) || uf == nil || uf.Ursache != gueltig || uf.Ort != "query/2":
		t.Errorf("%s: %v", name, err)
	}
	if !gleich(u, model.ODataURL{}) || l.aufrufe != 1 {
		t.Errorf("%s: Ergebnis %+v, %d Aufrufe", name, u, l.aufrufe)
	}
}

func leserOK() *zaehlLeser { return &zaehlLeser{ausdruck: testAusdruck{"ok"}} }

func leserMit(err error) *zaehlLeser { return &zaehlLeser{fehler: err} }

// TestFilterOhneLeser (Zeile 18): ohne Leser ist `$filter` mit lesbarem Wert `filter`, nie ein Ergebnis ohne Filter; leer ist
// `wert`, ungültig kodiert `kodierung`; die letzte Option gilt, `nil` ist kein Leser. Menge 7.
func TestFilterOhneLeser(t *testing.T) {
	l1, l2, l3 := leserOK(), leserOK(), leserOK()
	n := 0
	for _, p := range []*urlparser.Parser{urlparser.New(), urlparser.New(urlparser.WithFilter(nil)),
		urlparser.New(urlparser.WithFilter(l1), urlparser.WithFilter(nil))} {
		pruefeFall(t, p, fall{pfad: "/S", query: "$filter=a", u: model.URLFilter, ort: "query/1"})
		n++
	}
	pruefeFall(t, urlparser.New(), fall{pfad: "/S", query: "$filter=", u: wertU, ort: "query/1"})
	pruefeFall(t, urlparser.New(), fall{pfad: "/S", query: "$filter=%G1", u: kodi, ort: "query/1"})
	n += 2
	// die letzte Option gilt: nil, dann ein Leser
	u, err := urlparser.New(urlparser.WithFilter(nil), urlparser.WithFilter(l2)).ParseURL("/S", "$filter=a")
	if err != nil || u.Optionen.Filter == nil || u.Optionen.Filter.Ausdruck != (testAusdruck{"ok"}) || l2.aufrufe != 1 {
		t.Errorf("nil, dann Leser: %+v %v, Aufrufe %d", u.Optionen.Filter, err, l2.aufrufe)
	}
	n++
	// ein Leser, den WithFilter(nil) ersetzt, wird nie gerufen
	if l1.aufrufe != 0 || l3.aufrufe != 0 {
		t.Errorf("Aufrufe: %d %d", l1.aufrufe, l3.aufrufe)
	}
	n++
	if n != 7 {
		t.Fatalf("%d Fälle, erwartet 7", n)
	}
}

// TestFilterMitLeser (Zeile 19): der Leser wird höchstens einmal je Aufruf gerufen, mit dem dekodierten Text, an der Stelle von
// `$filter`, nicht mehr nach einem Verstoß der Klasse `ungueltig` davor; sein Ergebnis steht in Filter.Ausdruck; sein Fehler
// bekommt den Ort der Option. Menge 11.
func TestFilterMitLeser(t *testing.T) {
	fremd := &model.URLFehler{Ursache: "gibt_es_nicht", Ort: "query/1"}
	faelle := []struct {
		name           string
		leser          *zaehlLeser
		pfad, query    string
		u              model.URLUrsache // leer: gelesen
		ort            string
		aufrufe        int
		text           string
		programmfehler bool
	}{
		{"ok", leserOK(), "/S", "$filter=a", "", "", 1, "a", false},
		{"400", leserMit(&model.URLFehler{Ursache: model.URLWert, Ort: "irgendwo"}), "/S", "x=1&y=2&$filter=a", wertU, "query/3", 1, "a", false},
		{"501", leserMit(&model.URLFehler{Ursache: model.URLFilter, Ort: "irgendwo"}), "/S", "x=1&$filter=a", model.URLFilter, "query/2", 1, "a", false},
		{"fremder Fehler", leserMit(errors.New("kaputt")), "/S", "$filter=a", "", "", 1, "a", true},
		{"fremde Ursache", leserMit(fremd), "/S", "$filter=a", "", "", 1, "a", true},
		{"weder Ausdruck noch Fehler", &zaehlLeser{}, "/S", "$filter=a", "", "", 1, "a", true},
		{"Text dekodiert", leserOK(), "/S", "$filter=a%26b%C3%A4%2B", "", "", 1, "a&b\xc3\xa4+", false},
		{"400 im Pfad davor", leserOK(), "/S(1x)", "$filter=a", litU, "pfad/1", 0, "", false},
		{"400 in der Query davor", leserOK(), "/S", "$top=-1&$filter=a", wertU, "query/1", 0, "", false},
		{"501 davor", leserOK(), "/$batch", "$expand=1&$filter=a", pnu, "pfad/1", 1, "a", false},
		{"$filter doppelt", leserOK(), "/S", "$filter=a&$filter=b", doppU, "query/2", 1, "a", false},
	}
	if len(faelle) != 11 {
		t.Fatalf("%d Fälle", len(faelle))
	}
	for _, f := range faelle {
		p := urlparser.New(urlparser.WithFilter(f.leser))
		var uf *model.URLFehler
		switch {
		case f.programmfehler:
			u, err := p.ParseURL(f.pfad, f.query)
			if !errors.Is(err, model.ErrUngueltig) || errors.As(err, &uf) || !gleich(u, model.ODataURL{}) {
				t.Errorf("%s: %v", f.name, err)
			}
		case f.u == "":
			u, err := p.ParseURL(f.pfad, f.query)
			if err != nil || u.Optionen.Filter == nil || u.Optionen.Filter.Text != f.text || u.Optionen.Filter.Ausdruck != (testAusdruck{"ok"}) {
				t.Errorf("%s: %+v %v", f.name, u.Optionen.Filter, err)
			}
		default:
			pruefeFall(t, p, fall{pfad: f.pfad, query: f.query, u: f.u, ort: f.ort})
		}
		if f.leser.aufrufe != f.aufrufe || (f.aufrufe > 0 && (len(f.leser.texte) != 1 || f.leser.texte[0] != f.text)) {
			t.Errorf("%s: Aufrufe %d mit %q, erwartet %d mit %q", f.name, f.leser.aufrufe, f.leser.texte, f.aufrufe, f.text)
		}
	}
}

// TestFilterModi (Zeile 19, Zusatztabelle Z7): vier Modi des Lesers × ein `$top=-1` davor oder dahinter; der Leser wird nach einem
// 400 davor nicht gerufen, und ein 400 dahinter geht seinem 501 vor. Menge 8.
func TestFilterModi(t *testing.T) {
	modi := []struct {
		name  string
		bau   func() *zaehlLeser
		davor model.URLUrsache
		oVor  string
		dahin model.URLUrsache
		oDah  string
	}{
		{"kein Leser", func() *zaehlLeser { return nil }, wertU, "query/1", wertU, "query/2"},
		{"Leser ok", leserOK, wertU, "query/1", wertU, "query/2"},
		{"Leser 400", func() *zaehlLeser { return leserMit(&model.URLFehler{Ursache: model.URLKey}) }, wertU, "query/1", model.URLKey, "query/1"},
		{"Leser 501", func() *zaehlLeser { return leserMit(&model.URLFehler{Ursache: model.URLFilter}) }, wertU, "query/1", wertU, "query/2"},
	}
	n := 0
	for _, m := range modi {
		for i, q := range []string{"$top=-1&$filter=a", "$filter=a&$top=-1"} {
			l := m.bau()
			p := urlparser.New()
			if l != nil {
				p = urlparser.New(urlparser.WithFilter(l))
			}
			soll, aufrufe := fall{pfad: "/S", query: q, u: m.davor, ort: m.oVor}, 0
			if i == 1 {
				soll.u, soll.ort, aufrufe = m.dahin, m.oDah, 1
			}
			pruefeFall(t, p, soll)
			if l != nil && l.aufrufe != aufrufe {
				t.Errorf("%s %q: %d Aufrufe, erwartet %d", m.name, q, l.aufrufe, aufrufe)
			}
			n++
		}
	}
	if n != 8 {
		t.Fatalf("%d Fälle", n)
	}
}
