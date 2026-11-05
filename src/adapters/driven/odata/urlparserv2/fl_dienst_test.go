package urlparserv2_test

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flQ ist die Option `$filter` mit text, jedes Byte als `%XX` geschrieben.
func flQ(text string) string { return "$filter=" + pct(text) }

// flZaehlLeser ist der Zähl-Fake aus `slice-v2-query` um den echten Leser: er hält jedes Argument fest und zählt jeden Aufruf.
func flZaehlLeser() *fakeFilterLeser {
	return fakeMit(func(text string) (model.FilterAusdruck, error) { return urlparserv2.NewFilterLeser().LeseFilter(text) })
}

// flFilterOk ist die Erwartung einer gelesenen URL der Menge `/S` mit dem Filter text und dem Baum.
func flFilterOk(text string, baum flBaum, o model.Optionen) erw {
	o.Filter = &model.FilterOption{Text: text, Ausdruck: baum}
	return optOk(o)
}

// flUrsachen sind die sieben Ursachen des Lesers mit je einem Text, der sie unmittelbar auslöst, und einem Text, der ein Stück
// des Markers `ZZMARK` im Verstoß selbst trägt.
func flUrsachen() []struct {
	name, text, mitMarker string
	v                     flV
} {
	return []struct {
		name, text, mitMarker string
		v                     flV
	}{
		{"kodierung", "\xff", "ZZMARK\xff", flU(model.URLKodierung)},
		{"wert", ")", "ZZMARK#", flU(model.URLWert)},
		{"literal", "1x", "1ZZMARK", flU(model.URLLiteral)},
		{"zu_lang", "'" + flWiederholt("a", 512) + "'", "'ZZMARK" + flWiederholt("a", 520) + "'", flU(model.URLZuLang)},
		{"zu_viele", flWiederholt("(", 33), flWiederholt("(", 33) + "ZZMARK", flU(model.URLZuViele)},
		{"filter", "foo(1)", "ZZMARK(1)", flK(model.KompatFilter)},
		{"literal_typ", "X'a'", "ZZMARKx'a'", flK(model.KompatLiteralTyp)},
	}
}

// flOrt ist die Erwartung für den Verstoß v an der Stelle `query/n`.
func flOrt(v flV, n int) erw {
	if v.u != "" {
		return ung(v.u, fmt.Sprintf("query/%d", n))
	}
	return kmp(v.k, fmt.Sprintf("query/%d", n))
}

// flQueryMit stellt `$filter` an die n-te Stelle der Query.
func flQueryMit(text string, n int) string {
	return strings.Join(append([]string{"$top=1", "$skip=1"}[:n-1], flQ(text)), "&")
}

// TestFlFehlerwert (Zeile 27): `*URLFehler` mit Code und Sentinel der Klasse `ungueltig` für `kodierung`, `wert`, `literal`,
// `zu_lang`, `zu_viele`; `*KompatFehler` mit `ODR-E6006` und `ErrKompatibilitaet` für `filter` und `literal_typ`, nie beides;
// Ergebnis `nil`, `Ort` leer, über `ParseURL` `query/<n>`; kein Teil der Eingabe steht in `Error()` oder `Ort`; die Kodierung
// geht der Grammatik vor. Menge 52 = 7 + 3 + 21 + 21.
func TestFlFehlerwert(t *testing.T) {
	ursachen := flUrsachen()
	n := 0
	for _, u := range ursachen {
		flFehler(t, u.text, u.v) // die Form des Fehlers prüft flUrteil
		n++
	}
	for _, bad := range []string{"\x00", "\xff", "\xc3\x28"} {
		flFehler(t, "foo("+bad+")", flU(model.URLKodierung))
		n++
	}
	p := flMitLeser()
	for _, u := range ursachen {
		for pos := 1; pos <= 3; pos++ {
			pruefe(t, p, "/S", flQueryMit(u.text, pos), flOrt(u.v, pos))
			n++
		}
	}
	for _, u := range ursachen {
		for _, text := range []string{"ZZMARK eq 1 and " + u.text, "Id eq 'ZZMARK' and " + u.text, u.mitMarker} {
			_, err := flLies(text)
			_, errURL := p.ParseURL("/S", flQ(text))
			if err == nil || errURL == nil || strings.Contains(err.Error(), "ZZMARK") || strings.Contains(errURL.Error(), "ZZMARK") {
				t.Errorf("%s: Eingabe %.40q: Fehler %v, über ParseURL %v", u.name, text, err, errURL)
			}
			flFehler(t, text, u.v)
			n++
		}
	}
	if n != 52 {
		t.Fatalf("%d Fälle, erwartet 52", n)
	}
}

// flBaeume sind die sechs Bäume der Einhängung mit ihrem Text.
func flBaeume() []flFall {
	return []flFall{
		{"Id eq 1", flBin("eq", flEig("Id"), flZahl("1"))}, {"not (a and b)", flNicht(flBin("and", flEig("a"), flEig("b")))},
		{"a or b and c", flBin("or", flEig("a"), flBin("and", flEig("b"), flEig("c")))}, {"(a)", flEig("a")},
		{"x eq 'O''Hara'", flBin("eq", flEig("x"), flZk("'O''Hara'"))}, {"null eq null", flBin("eq", flNull(), flNull())},
	}
}

// TestFlEinhaengung (Zeile 28): mit echtem Leser ist `Filter.Text` der dekodierte Wert und `Filter.Ausdruck` die Wurzel, die
// Reihenfolge der Optionen ändert nichts; `New()` und `WithFilter(nil)` bleiben `filter`; die Klasse `ungueltig` des Lesers und
// die einer anderen Option ordnet `SPEC-173` Nr. 6; `$filter` doppelt ist `option_doppelt`, der Leser wird nur für das erste
// gerufen. Menge 15 = 6 Bäume + 3 Stellen + 2 ohne Leser + 4 Rangfolgen.
func TestFlEinhaengung(t *testing.T) {
	top, skip := model.ZahlOption{Gesetzt: true, Wert: 1}, model.ZahlOption{Gesetzt: true, Wert: 2}
	n := 0
	lauf := func(l *fakeFilterLeser, q string, aufrufe []string, want erw) {
		t.Helper()
		pruefe(t, urlparserv2.New(urlparserv2.WithFilter(l)), "/S", q, want)
		if !slices.Equal(l.aufrufe(), aufrufe) || l.n.Load() != int64(len(aufrufe)) {
			t.Errorf("Query %q: Aufrufe des Lesers %q (%d), erwartet %q", q, l.aufrufe(), l.n.Load(), aufrufe)
		}
		n++
	}
	for _, f := range flBaeume() {
		lauf(flZaehlLeser(), flQ(f.text), []string{f.text}, flFilterOk(f.text, f.want, model.Optionen{}))
	}
	b := flBaeume()[0]
	mitTopSkip := model.Optionen{Top: top, Skip: skip}
	for _, q := range []string{flQ(b.text) + "&$top=1&$skip=2", "$top=1&" + flQ(b.text) + "&$skip=2", "$top=1&$skip=2&" + flQ(b.text)} {
		lauf(flZaehlLeser(), q, []string{b.text}, flFilterOk(b.text, b.want, mitTopSkip))
	}
	for _, p := range []*urlparserv2.Parser{urlparserv2.New(), urlparserv2.New(urlparserv2.WithFilter(nil))} {
		pruefe(t, p, "/S", flQ(b.text), kmp(model.KompatFilter, "query/1"))
		n++
	}
	lauf(flZaehlLeser(), flQ("foo(1)")+"&$top=x", []string{"foo(1)"}, ung(model.URLWert, "query/2"))
	lauf(flZaehlLeser(), flQ(")")+"&$expand=1", []string{")"}, ung(model.URLWert, "query/1"))
	lauf(flZaehlLeser(), "$expand=1&"+flQ("foo(1)"), []string{"foo(1)"}, kmp(model.KompatOption, "query/1"))
	lauf(flZaehlLeser(), flQ("Id eq 1")+"&"+flQ("Id eq 2"), []string{"Id eq 1"}, ung(model.URLOptionDoppelt, "query/2"))
	if n != 15 {
		t.Fatalf("%d Fälle, erwartet 15", n)
	}
}

// TestFlDekodierung (Zeile 29): der Leser sieht den einmal dekodierten Text. Menge 22 = 6 Zeichen × 3 Schreibweisen (roh, `%XX`,
// `%xx`) mit gleichem Ergebnis + `%2527` + 3 für `+`.
func TestFlDekodierung(t *testing.T) {
	p := flMitLeser()
	erwartung := func(text string) erw {
		a, v := flUrteil(t, text)
		if v == (flV{}) {
			return flFilterOk(text, a, model.Optionen{})
		}
		return flOrt(v, 1)
	}
	n := 0
	for _, f := range []struct{ zeichen, basis string }{
		{"'", "Id eq 'a'"}, {"(", "not(Id)"}, {")", "(Id)"}, {" ", "Id eq 1"}, {",", "Id eq 1,"}, {"/", "a/b"},
	} {
		hex := fmt.Sprintf("%%%02X", f.zeichen[0])
		for _, schreibweise := range []string{f.zeichen, hex, strings.ToLower(hex)} {
			pruefe(t, p, "/S", "$filter="+strings.ReplaceAll(f.basis, f.zeichen, schreibweise), erwartung(f.basis))
			n++
		}
	}
	pruefe(t, p, "/S", "$filter=Name%20eq%20'%2527'", flFilterOk("Name eq '%27'", flBin("eq", flEig("Name"), flZk("'%27'")), model.Optionen{}))
	pruefe(t, p, "/S", "$filter=Id+eq+1", ung(model.URLWert, "query/1"))
	pruefe(t, p, "/S", "$filter='a+b'", flFilterOk("'a+b'", flZk("'a+b'"), model.Optionen{}))
	pruefe(t, p, "/S", "$filter='a%2Bb'", flFilterOk("'a+b'", flZk("'a+b'"), model.Optionen{}))
	if n != 18 {
		t.Fatalf("%d Schreibweisen, erwartet 18", n)
	}
}

// flWerteOnly prüft, dass der Baum nur Werte trägt: kein Zeiger, keine Map, kein Slice, kein Kanal (SPEC-165 Nr. 2); so teilt er
// nichts mit einem anderen.
func flWerteOnly(t *testing.T, v reflect.Value, pfad string) {
	t.Helper()
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		t.Errorf("%s: %s teilt Speicher", pfad, v.Kind())
	case reflect.Interface:
		flWerteOnly(t, v.Elem(), pfad+"."+v.Elem().Type().Name())
	case reflect.Struct:
		for i := range v.NumField() {
			flWerteOnly(t, v.Field(i), pfad+"."+v.Type().Field(i).Name)
		}
	}
}

// flGleichesErgebnis bildet Baum oder Fehler als vergleichbaren Wert: Baum und Typ samt Text des Fehlers.
func flGleichesErgebnis(l interface {
	LeseFilter(text string) (model.FilterAusdruck, error)
}, text string) string {
	a, err := l.LeseFilter(text)
	return fmt.Sprintf("%#v | %T %v", a, err, err)
}

// TestFlWiederholung (Zeile 31): gleicher Text, gleicher Baum (tiefe Gleichheit) und gleicher Fehler, auch auf `NewFilterLeser()`
// und `FilterGrammatik{}`; der Baum besteht nur aus Werten und teilt nichts mit einem anderen. Menge 64 × 2 Aufrufe × 2 Instanzen:
// 40 Gutfälle und die 24 Fehlerfälle der Zeile 30.
func TestFlWiederholung(t *testing.T) {
	var texte []string
	for _, f := range flKorpus() {
		texte = append(texte, f.text)
	}
	texte = append(texte, flFehlerKorpus()...)
	if len(texte) != 64 {
		t.Fatalf("%d Texte", len(texte))
	}
	instanzen := []interface {
		LeseFilter(text string) (model.FilterAusdruck, error)
	}{urlparserv2.NewFilterLeser(), &urlparserv2.FilterGrammatik{}}
	// Der Bezug ist der Aufruf auf einer frischen Instanz; die geteilten Instanzen lesen die Texte in der Folge der Liste und in
	// einer Folge, die Gutfälle und Fehlerfälle abwechselt (Schritt 37 ist teilerfremd zu 64).
	soll := make([]string, len(texte))
	for i, text := range texte {
		soll[i] = flGleichesErgebnis(urlparserv2.NewFilterLeser(), text)
	}
	for k, l := range instanzen {
		for _, schritt := range []int{1, 37} {
			for j := range texte {
				i := j * schritt % len(texte)
				if got := flGleichesErgebnis(l, texte[i]); got != soll[i] {
					t.Errorf("%q: Instanz %d, Schritt %d: %s, erwartet %s", texte[i], k, schritt, got, soll[i])
				}
			}
		}
	}
	for _, text := range texte {
		if a, err := flLies(text); err == nil {
			flWerteOnly(t, reflect.ValueOf(&a).Elem(), "Baum")
		}
	}
}

// flBezug ist das Ergebnis des einzelnen Aufrufs auf einer frischen Instanz: Baum und Fehlerwert.
type flBezug struct {
	baum flBaum
	err  error
}

// flBezuege liefert die 64 Texte der Zeile 31 und die zwei Ausdrücke der Tiefe (32 gültig, 33 `zu_viele`) mit ihrem Bezug.
func flBezuege() ([]string, []flBezug) {
	var texte []string
	for _, f := range flKorpus() {
		texte = append(texte, f.text)
	}
	texte = append(texte, flFehlerKorpus()...)
	texte = append(texte, flWiederholt("(", 32)+"a"+flWiederholt(")", 32), flWiederholt("(", 33)+"a"+flWiederholt(")", 33))
	bezug := make([]flBezug, len(texte))
	for i, text := range texte {
		bezug[i].baum, bezug[i].err = urlparserv2.NewFilterLeser().LeseFilter(text)
	}
	return texte, bezug
}

// flParallel ruft l aus 16 Goroutinen je 10 Runden über alle Texte (verschränkt versetzt) und zählt die Ergebnisse, die vom
// Bezug abweichen. Die Goroutinen teilen nur den Leser und die Bezüge (lesend): kein Kanal, keine Sperre, kein Zähler und kein
// Aufruf von fmt zwischen den Aufrufen, damit der Detektor jeden ungeordneten Zugriff des Lesers sieht.
func flParallel(l interface {
	LeseFilter(text string) (model.FilterAusdruck, error)
}, texte []string, bezug []flBezug) int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	abweichend := make([]int, 16)
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 10 {
				for k := range texte {
					i := (k + g*7) % len(texte)
					a, err := l.LeseFilter(texte[i])
					if !reflect.DeepEqual(a, bezug[i].baum) || !reflect.DeepEqual(err, bezug[i].err) {
						abweichend[g]++
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	summe := 0
	for _, n := range abweichend {
		summe += n
	}
	return summe
}

// flParallelFrisch ruft l aus 16 Goroutinen je 10 Runden mit 64 Texten, die vorher kein Aufruf gelesen hat (Name und Zahl sind je
// Aufruf verschieden); der erwartete Baum entsteht im Test, nicht im Leser. So schreibt ein Cache oder eine andere Zuteilung des
// Lesers schon beim ersten Lesen, gleichzeitig mit den anderen Goroutinen, und der Detektor sieht diesen Zugriff. Die Zahl der
// abweichenden Ergebnisse ist die Rückgabe.
func flParallelFrisch(l interface {
	LeseFilter(text string) (model.FilterAusdruck, error)
}) int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	abweichend := make([]int, 16)
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for r := range 10 {
				for k := range 64 {
					n := strconv.Itoa((g*10+r)*64 + k)
					a, err := l.LeseFilter("N" + n + " eq " + n)
					if err != nil || !reflect.DeepEqual(a, flBin("eq", flEig("N"+n), flZahl(n))) {
						abweichend[g]++
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	summe := 0
	for _, n := range abweichend {
		summe += n
	}
	return summe
}

// TestFlGleichzeitig (Zeilen 32 und 33): gleichzeitige Aufrufe auf einer Instanz liefern je das Ergebnis des einzelnen Aufrufs.
// Menge: 16 Goroutinen × 64 Texte × 10 Runden, dazu zwei verschränkte Ausdrücke (Tiefe 32 gültig, Tiefe 33 `zu_viele`): ein
// Tiefenzähler im Feld der Instanz lieferte dem einen das Ergebnis des anderen; dazu 16 × 10 × 64 Texte, die vorher kein Aufruf
// gelesen hat. Der Fänger zählt die Wirkung und ist vom Detektor unabhängig; `make test-race` meldet das Fehlen der Ordnung
// (Zeile 33), auch für einen Zugriff, der erst beim ersten Lesen eines Textes geschieht.
func TestFlGleichzeitig(t *testing.T) {
	texte, bezug := flBezuege()
	n := len(texte)
	if _, ok := bezug[n-2].baum.(model.FilterEigenschaft); !ok || bezug[n-2].err != nil || bezug[n-1].baum != nil ||
		!strings.Contains(bezug[n-1].err.Error(), "zu_viele") {
		t.Fatalf("Bezug der Tiefe: %v %v | %v %v", bezug[n-2].baum, bezug[n-2].err, bezug[n-1].baum, bezug[n-1].err)
	}
	if abweichend := flParallel(urlparserv2.NewFilterLeser(), texte, bezug); abweichend != 0 {
		t.Errorf("%d von %d Ergebnissen weichen vom einzelnen Aufruf ab", abweichend, 16*n*10)
	}
	if abweichend := flParallelFrisch(urlparserv2.NewFilterLeser()); abweichend != 0 {
		t.Errorf("%d von %d Ergebnissen für Texte ohne frühere Lesung weichen vom Baum des Tests ab", abweichend, 16*10*64)
	}
}
