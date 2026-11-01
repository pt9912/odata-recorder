package urlparserv2_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// erw ist die Erwartung an einen Aufruf: ein Ergebnis, ein URLFehler der Klasse `ungueltig`, ein KompatFehler oder ein Fehler des
// Aufrufers (model.ErrUngueltig ohne Fehlerwert).
type erw struct {
	url      *model.ODataURL
	u        model.URLUrsache
	k        model.KompatUrsache
	ort      string
	aufrufer bool
}

func ok(r model.Ressource, o model.Optionen) erw {
	return erw{url: &model.ODataURL{Ressource: r, Optionen: o}}
}
func okR(r model.Ressource) erw                 { return ok(r, model.Optionen{}) }
func ung(u model.URLUrsache, ort string) erw    { return erw{u: u, ort: ort} }
func kmp(k model.KompatUrsache, ort string) erw { return erw{k: k, ort: ort} }
func aufr() erw                                 { return erw{aufrufer: true} }

func mengeR(set string) model.Ressource {
	return model.Ressource{Art: model.RessourceMenge, EntitySet: set}
}
func anzahlR(set string) model.Ressource {
	return model.Ressource{Art: model.RessourceAnzahl, EntitySet: set}
}
func entityR(set string, teile ...model.SchluesselTeil) model.Ressource {
	return model.Ressource{Art: model.RessourceEntity, EntitySet: set, Schluessel: teile}
}
func teil(name string, art model.LiteralArt, text string) model.SchluesselTeil {
	return model.SchluesselTeil{Name: name, Literal: model.Literal{Art: art, Text: text}}
}
func zahlT(text string) model.SchluesselTeil { return teil("", model.LiteralZahl, text) }

// zeichen bildet ein Zeichen aus seiner Codepunktzahl; Werte jenseits von ASCII entstehen nie aus einem Escape im Quelltext.
func zeichen(cp int) string { return string(rune(cp)) }

// pct schreibt jedes Byte von s als `%XX` (Großbuchstaben).
func pct(s string) string {
	var b strings.Builder
	for i := range len(s) {
		fmt.Fprintf(&b, "%%%02X", s[i])
	}
	return b.String()
}

// pruefe ruft ParseURL und hält Ergebnis und Fehler gegen e: bei einem Fehler ist das Ergebnis der Nullwert, der Fehler hat genau
// den Typ, die Ursache, den Ort und die Sentinel seiner Klasse.
func pruefe(t *testing.T, p *urlparserv2.Parser, pfad, q string, e erw) {
	t.Helper()
	got, err := p.ParseURL(pfad, q)
	bez := fmt.Sprintf("ParseURL(%q, %q)", pfad, q)
	if e.url != nil {
		if err != nil || !reflect.DeepEqual(got, *e.url) {
			t.Errorf("%s = %+v, %v; erwartet %+v", bez, got, err, *e.url)
		}
		return
	}
	if !reflect.DeepEqual(got, model.ODataURL{}) {
		t.Errorf("%s: Ergebnis %+v ist nicht der Nullwert", bez, got)
	}
	var uf *model.URLFehler
	var kf *model.KompatFehler
	istU, istK := errors.As(err, &uf), errors.As(err, &kf)
	switch {
	case e.aufrufer:
		if err == nil || !errors.Is(err, model.ErrUngueltig) || istU || istK {
			t.Errorf("%s: %v (URLFehler %v, KompatFehler %v), erwartet model.ErrUngueltig ohne Fehlerwert", bez, err, istU, istK)
		}
	case e.u != "":
		if !istU || istK || uf.Ursache != e.u || uf.Ort != e.ort || !errors.Is(err, model.ErrURLUngueltig) ||
			errors.Is(err, model.ErrURLNichtUnterstuetzt) || errors.Is(err, model.ErrKompatibilitaet) {
			t.Errorf("%s: %v, erwartet URLFehler %s %q", bez, err, e.u, e.ort)
		}
	default:
		if !istK || istU || kf.Ursache != e.k || kf.Ort != e.ort || !errors.Is(err, model.ErrKompatibilitaet) ||
			errors.Is(err, model.ErrURLNichtUnterstuetzt) || errors.Is(err, model.ErrURLUngueltig) {
			t.Errorf("%s: %v, erwartet KompatFehler %s %q", bez, err, e.k, e.ort)
		}
	}
}

// fakeFilterLeser ist der Zähl-Fake des FilterLeser: er hält jedes Argument fest, zählt jeden Aufruf atomar und liefert je
// Fall Ausdruck und Fehler.
type fakeFilterLeser struct {
	n       atomic.Int64
	mu      sync.Mutex
	args    []string
	liefert func(text string) (model.FilterAusdruck, error)
}

func (f *fakeFilterLeser) LeseFilter(text string) (model.FilterAusdruck, error) {
	f.n.Add(1)
	f.mu.Lock()
	f.args = append(f.args, text)
	f.mu.Unlock()
	if f.liefert == nil {
		return model.FilterEigenschaft{Name: text}, nil
	}
	return f.liefert(text)
}

func (f *fakeFilterLeser) aufrufe() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.args...)
}

func fakeMit(f func(string) (model.FilterAusdruck, error)) *fakeFilterLeser {
	return &fakeFilterLeser{liefert: f}
}

// TestFakeFilterLeser hält den Zähl-Fake gegen sich selbst: Zähler und festgehaltene Argumente stimmen unter Gleichzeitigkeit.
func TestFakeFilterLeser(t *testing.T) {
	f := &fakeFilterLeser{}
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 50 {
				ausdruck, err := f.LeseFilter(fmt.Sprintf("%d-%d", g, i))
				if err != nil || ausdruck != (model.FilterEigenschaft{Name: fmt.Sprintf("%d-%d", g, i)}) {
					t.Errorf("Fake lieferte %v, %v", ausdruck, err)
				}
			}
		}()
	}
	wg.Wait()
	if f.n.Load() != 400 || len(f.aufrufe()) != 400 {
		t.Errorf("Zähler %d, Argumente %d, erwartet 400", f.n.Load(), len(f.aufrufe()))
	}
	g := fakeMit(func(string) (model.FilterAusdruck, error) { return nil, errors.New("x") })
	if _, err := g.LeseFilter("a"); err == nil || g.n.Load() != 1 || g.aufrufe()[0] != "a" {
		t.Error("der Fake liefert den vorgegebenen Fehler nicht oder zählt nicht")
	}
}

// Die Mengen außerhalb des Umfangs stehen in SPEC-170 Nr. 3 und SPEC-173 Nr. 4; der Test trägt sie als eigene Listen und zählt
// sie, nie aus dem Code.
var (
	// ausserhalbOptionen sind die zwölf Optionen mit der Kompatibilitätsursache `option` (SPEC-173 Nr. 4).
	ausserhalbOptionen = []string{"$expand", "$skiptoken", "$callback", "$id", "$count", "$search", "$apply", "$compute",
		"$deltatoken", "$levels", "$index", "$schemaversion"}
	// ausserhalbPfade sind die 15 Pfadformen außerhalb mit der Kompatibilitätsursache `pfad` und ihrem Ort.
	ausserhalbPfade = []struct{ pfad, ort string }{
		{"/", "pfad"}, {"/$batch", "pfad/1"}, {"/$entity", "pfad/1"}, {"/$all", "pfad/1"}, {"/$crossjoin", "pfad/1"},
		{"/S()", "pfad/1"}, {"/S(1)/N", "pfad/2"}, {"/S/N", "pfad/2"}, {"/S(1)/$links/N", "pfad/2"}, {"/S(1)/$value", "pfad/2"},
		{"/S(1)/N/$value", "pfad/2"}, {"/S(1)/$ref", "pfad/2"}, {"/S(1)/N/$count", "pfad/2"}, {"/S(1)/Ns.T", "pfad/2"},
		{"/S(1)/N(2)", "pfad/2"},
	}
	// ausserhalbFormate sind die zehn Werte von `$format` außerhalb (SPEC-170 Nr. 3).
	ausserhalbFormate = []string{"atom", "xml", "application/atom+xml", "application/xml", "text/plain", "jsonx", "application/jsonx",
		"application/json-patch+json", "verbose", "ATOM"}
	// ausserhalbLiterale sind die 25 Formen eines Schlüsselliterals mit der Kompatibilitätsursache `literal_typ` (SPEC-173 Nr. 3).
	ausserhalbLiterale = []string{"1d", "1D", "1f", "1F", "1.5d", "1.5D", "1.5f", "1.5F", "1e3", "1E3", "1.5e-2", "-1e+3",
		"binary'AA'", "X'AB'", "x'AB'", "time'PT1S'", "datetimeoffset'2020-01-01T00:00:00Z'", "geography'POINT(1 2)'", "geometry'x'",
		"duration'P1D'", "Ns.Typ'v'", "Edm.Foo'v'", zeichen(0xE4) + "x'v'", "guidx'a'", "datetime2'a'"}
)

func TestMengenGezaehlt(t *testing.T) {
	if len(ausserhalbOptionen) != 12 || len(ausserhalbPfade) != 15 || len(ausserhalbFormate) != 10 || len(ausserhalbLiterale) != 25 {
		t.Errorf("Mengen %d, %d, %d, %d; erwartet 12, 15, 10, 25",
			len(ausserhalbOptionen), len(ausserhalbPfade), len(ausserhalbFormate), len(ausserhalbLiterale))
	}
}
