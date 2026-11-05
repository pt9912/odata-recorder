package urlparserv2_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.URLCodec = (*urlparserv2.Parser)(nil)

type urlFake struct{}

func (urlFake) ParseURL(string, string) (model.ODataURL, error) { return model.ODataURL{}, nil }

// TestSatzTraegtURL (Zeilen 15 und 22): der CodecSatz trägt das Feld URL als Schnittstelle; *Parser und ein Fake erfüllen sie; ein
// Satz ohne URL bleibt gültig; die Signatur von ParseURL trägt keine Kopfzeile. Menge 3 und die Signatur.
func TestSatzTraegtURL(t *testing.T) {
	var leer outbound.CodecSatz
	if leer.URL != nil {
		t.Error("der Nullwert des Feldes URL ist nicht nil")
	}
	satz := outbound.CodecSatz{Version: "2.0", URL: urlparserv2.New(), NurLesend: true}
	if got, err := satz.URL.ParseURL("/S", ""); err != nil || got.Ressource.EntitySet != "S" {
		t.Errorf("Satz.URL.ParseURL = %+v, %v", got, err)
	}
	satz.URL = urlFake{}
	if got, err := satz.URL.ParseURL("/S", ""); err != nil || got.Ressource.EntitySet != "" {
		t.Errorf("Fake im Satz: %+v, %v", got, err)
	}
	feld, ok := reflect.TypeOf(leer).FieldByName("URL")
	if !ok || feld.Type != reflect.TypeFor[outbound.URLCodec]() {
		t.Errorf("das Feld URL ist %v, erwartet outbound.URLCodec", feld.Type)
	}
	m := reflect.TypeOf(urlparserv2.New().ParseURL)
	if m.NumIn() != 2 || m.In(0).Kind() != reflect.String || m.In(1).Kind() != reflect.String || m.NumOut() != 2 {
		t.Errorf("Signatur von ParseURL: %v", m)
	}
}

func fakeErgebnis(a model.FilterAusdruck, err error) *fakeFilterLeser {
	return fakeMit(func(string) (model.FilterAusdruck, error) { return a, err })
}

func filterOk(text string, o model.Optionen) erw {
	o.Filter = &model.FilterOption{Text: text, Ausdruck: model.FilterEigenschaft{Name: text}}
	return optOk(o)
}

// TestEinhaengungAufruf (Zeile 19): der Leser wird höchstens einmal je ParseURL gerufen, mit dem dekodierten Wert, an der Stelle von
// `$filter`, nicht mehr, sobald davor ein Verstoß der Klasse `ungueltig` feststeht. Menge 16.
func TestEinhaengungAufruf(t *testing.T) {
	top := model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1}}
	topSkip := model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1}, Skip: model.ZahlOption{Gesetzt: true, Wert: 1}}
	faelle := []struct {
		pfad, q string
		aufrufe []string
		want    erw
	}{
		{"/S", "$filter=a&$top=1", []string{"a"}, filterOk("a", top)},
		{"/S", "$top=1&$filter=a&$skip=1", []string{"a"}, filterOk("a", topSkip)},
		{"/S", "$top=1&$filter=a", []string{"a"}, filterOk("a", top)},
		{"/S", "$top=1&&$filter=a", []string{"a"}, filterOk("a", top)},
		{"/S", "%24filter=a", []string{"a"}, filterOk("a", model.Optionen{})},
		{"/S", "$filter=a%20eq%20%27x%27", []string{"a eq 'x'"}, filterOk("a eq 'x'", model.Optionen{})},
		{"/S", "$filter=", nil, ung(model.URLWert, "query/1")},
		{"/S", "$filter=a&$filter=b", []string{"a"}, ung(model.URLOptionDoppelt, "query/2")},
		{"/S.", "$filter=a", nil, ung(model.URLSegment, "pfad/1")},
		{"/S(1,,2)", "$filter=a", nil, ung(model.URLKey, "pfad/1")},
		{"/S", "$foo=1&$filter=a", nil, ung(model.URLOptionName, "query/1")},
		{"/S", "$top=x&$filter=a", nil, ung(model.URLWert, "query/1")},
		{"/S", "%FF=1&$filter=a", nil, ung(model.URLKodierung, "query/1")},
		{"/S", "$filter=%FF", nil, ung(model.URLKodierung, "query/1")},
		{"/S/N", "$filter=a", []string{"a"}, kmp(model.KompatPfad, "pfad/2")},
		{"/S", "$expand=1&$filter=a", []string{"a"}, kmp(model.KompatOption, "query/1")},
	}
	for _, f := range faelle {
		l := &fakeFilterLeser{}
		pruefe(t, urlparserv2.New(urlparserv2.WithFilter(l)), f.pfad, f.q, f.want)
		if !slices.Equal(l.aufrufe(), f.aufrufe) || l.n.Load() != int64(len(f.aufrufe)) {
			t.Errorf("ParseURL(%q, %q): Aufrufe des Lesers %q (%d), erwartet %q", f.pfad, f.q, l.aufrufe(), l.n.Load(), f.aufrufe)
		}
	}
	if len(faelle) != 16 {
		t.Errorf("%d Fälle, erwartet 16", len(faelle))
	}
}

// TestEinhaengungFehler (Zeile 20): ein URLFehler der Klasse `ungueltig` oder ein KompatFehler mit der Ursache `filter` oder
// `literal_typ` in der Kette gilt mit seiner Ursache und dem Ort `query/<n>`; alles andere ist ein Fehler des Aufrufers; ein Fehler
// gewinnt über einen gleichzeitig gelieferten Ausdruck. Menge 36: 12 gültige Fehler, 19 Fehler des Aufrufers (18 und die Kette mit
// beiden Typen), 3 Ausdruck mit Fehler, 2 ohne Leser.
func TestEinhaengungFehler(t *testing.T) {
	var nilU *model.URLFehler
	var nilK *model.KompatFehler
	n := 0
	lauf := func(l *fakeFilterLeser, e erw) {
		pruefe(t, urlparserv2.New(urlparserv2.WithFilter(l)), "/S", "$filter=x", e)
		if l.n.Load() != 1 {
			t.Errorf("der Leser wurde %d-mal gerufen", l.n.Load())
		}
		n++
	}
	for _, u := range model.AlleURLUrsachen() {
		if !u.NichtUnterstuetzt() {
			lauf(fakeErgebnis(nil, &model.URLFehler{Ursache: u, Ort: "query/9"}), ung(u, "query/1"))
		}
	}
	for _, k := range []model.KompatUrsache{model.KompatFilter, model.KompatLiteralTyp} {
		lauf(fakeErgebnis(nil, &model.KompatFehler{Ursache: k, Ort: "pfad/3"}), kmp(k, "query/1"))
	}
	fremd := []error{
		errors.New("ohne Kette"), nilU, fmt.Errorf("verpackt: %w", nilU), nilK, fmt.Errorf("verpackt: %w", nilK),
		&model.URLFehler{Ursache: "fremd"}, &model.KompatFehler{Ursache: "fremd"},
		errors.Join(&model.URLFehler{Ursache: model.URLWert}, &model.KompatFehler{Ursache: model.KompatFilter}),
	}
	for _, u := range model.AlleURLUrsachen() {
		if u.NichtUnterstuetzt() {
			fremd = append(fremd, &model.URLFehler{Ursache: u})
		}
	}
	for _, k := range []model.KompatUrsache{model.KompatFormat, model.KompatPfad, model.KompatOption, model.KompatMethode} {
		fremd = append(fremd, &model.KompatFehler{Ursache: k})
	}
	for _, err := range fremd {
		lauf(fakeErgebnis(nil, err), aufr())
	}
	lauf(fakeErgebnis(nil, nil), aufr())
	ausdruck := model.FilterEigenschaft{Name: "x"}
	lauf(fakeErgebnis(ausdruck, &model.URLFehler{Ursache: model.URLWert}), ung(model.URLWert, "query/1"))
	lauf(fakeErgebnis(ausdruck, &model.KompatFehler{Ursache: model.KompatFilter}), kmp(model.KompatFilter, "query/1"))
	lauf(fakeErgebnis(ausdruck, errors.New("x")), aufr())
	for _, p := range []*urlparserv2.Parser{urlparserv2.New(), urlparserv2.New(urlparserv2.WithFilter(nil))} {
		pruefe(t, p, "/S", "$filter=x", kmp(model.KompatFilter, "query/1"))
		n++
	}
	if n != 12+19+3+2 || len(fremd) != 18 {
		t.Errorf("%d Fälle und %d Aufruferfehler, erwartet 36 und 18", n, len(fremd))
	}
}

// TestEinhaengungRang (Zeile 21): ein Fehler des Lesers reiht sich an der Stelle von `$filter` in die Lesefolge ein; die
// Reihenfolge der Optionen ändert den Baum nicht. Menge 13.
func TestEinhaengungRang(t *testing.T) {
	baeume := map[string]model.FilterAusdruck{"a": model.FilterEigenschaft{Name: "a"}, "b": model.FilterNicht{Operand: model.FilterEigenschaft{Name: "b"}}}
	liefert := func(text string) (model.FilterAusdruck, error) {
		switch text {
		case "k":
			return nil, &model.KompatFehler{Ursache: model.KompatFilter}
		case "w":
			return nil, &model.URLFehler{Ursache: model.URLWert}
		}
		return baeume[text], nil
	}
	p := urlparserv2.New(urlparserv2.WithFilter(fakeMit(liefert)))
	topSkip := model.Optionen{Top: model.ZahlOption{Gesetzt: true, Wert: 1}, Skip: model.ZahlOption{Gesetzt: true, Wert: 2}}
	n := 0
	for _, text := range []string{"a", "b"} {
		for _, q := range []string{"$filter=%s&$top=1&$skip=2", "$top=1&$filter=%s&$skip=2", "$top=1&$skip=2&$filter=%s"} {
			o := topSkip
			o.Filter = &model.FilterOption{Text: text, Ausdruck: baeume[text]}
			pruefe(t, p, "/S", fmt.Sprintf(q, text), optOk(o))
			n++
		}
	}
	for i, q := range []string{"$filter=k&$top=1&$skip=2", "$top=1&$filter=k&$skip=2", "$top=1&$skip=2&$filter=k"} {
		pruefe(t, p, "/S", q, kmp(model.KompatFilter, fmt.Sprintf("query/%d", i+1)))
		n++
	}
	pruefe(t, p, "/S", "$filter=k&$top=x", ung(model.URLWert, "query/2"))
	pruefe(t, p, "/S", "$filter=w&$expand=1", ung(model.URLWert, "query/1"))
	pruefe(t, urlparserv2.New(), "/S", "$filter=a&$top=x", ung(model.URLWert, "query/2"))
	pruefe(t, urlparserv2.New(), "/S", "$filter=a&$expand=x", kmp(model.KompatFilter, "query/1"))
	n += 4
	if n != 13 {
		t.Errorf("%d Fälle, erwartet 13", n)
	}
}
