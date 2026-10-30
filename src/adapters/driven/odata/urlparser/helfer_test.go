package urlparser_test

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparser"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

var _ outbound.URLCodec = (*urlparser.Parser)(nil)

// fall ist eine URL mit der erwarteten Ursache (leer: gelesen) und dem Ort; want ist, wenn gesetzt, das erwartete Ergebnis.
type fall struct {
	pfad, query string
	u           model.URLUrsache
	ort         string
	want        *model.ODataURL
}

// klasse501 ist das Orakel der Klasse: unabhängig von model.URLUrsache.NichtUnterstuetzt.
func klasse501(u model.URLUrsache) bool {
	switch string(u) {
	case "service_dokument", "pfad_nicht_unterstuetzt", "literal_typ", "option_nicht_unterstuetzt", "filter", "wert_nicht_unterstuetzt":
		return true
	}
	return false
}

func parse(pfad, query string) (model.ODataURL, error) {
	return urlparser.New().ParseURL(pfad, query)
}

// pruefeFall hält ein Ergebnis gegen den Fall: Ursache, Ort, Klasse über errors.Is, Nullwert bei Fehler.
func pruefeFall(t *testing.T, p *urlparser.Parser, f fall) {
	t.Helper()
	u, err := p.ParseURL(f.pfad, f.query)
	if f.u == "" {
		if err != nil {
			t.Errorf("%q %q: Fehler %v", f.pfad, f.query, err)
		} else if f.want != nil && !reflect.DeepEqual(u, *f.want) {
			t.Errorf("%q %q:\n  ist   %+v\n  soll  %+v", f.pfad, f.query, u, *f.want)
		}
		return
	}
	var uf *model.URLFehler
	if !errors.As(err, &uf) {
		t.Errorf("%q %q: erwartet %s, bekam %v", f.pfad, f.query, f.u, err)
		return
	}
	if uf.Ursache != f.u || uf.Ort != f.ort {
		t.Errorf("%q %q: Ursache %s Ort %q, erwartet %s %q", f.pfad, f.query, uf.Ursache, uf.Ort, f.u, f.ort)
	}
	if !reflect.DeepEqual(u, model.ODataURL{}) {
		t.Errorf("%q %q: Ergebnis bei Fehler ist nicht der Nullwert", f.pfad, f.query)
	}
	if errors.Is(err, model.ErrURLNichtUnterstuetzt) != klasse501(f.u) || errors.Is(err, model.ErrURLUngueltig) == klasse501(f.u) {
		t.Errorf("%q %q: Klasse von %s", f.pfad, f.query, f.u)
	}
}

func pruefeFaelle(t *testing.T, faelle []fall) {
	t.Helper()
	p := urlparser.New()
	for _, f := range faelle {
		pruefeFall(t, p, f)
	}
}

// gelesen liest eine URL, die gelesen werden muss.
func gelesen(t *testing.T, pfad, query string) model.ODataURL {
	t.Helper()
	u, err := parse(pfad, query)
	if err != nil {
		t.Fatalf("%q %q: %v", pfad, query, err)
	}
	return u
}

// rng bildet den Zufall eines Laufs aus der Saat; die Saat steht bei einem Rot in der Ausgabe.
func rng(saat uint64) *rand.Rand { return rand.New(rand.NewPCG(saat, 0x6f64)) }

func parserNeu() *urlparser.Parser { return urlparser.New() }

func gleich(a, b any) bool { return reflect.DeepEqual(a, b) }

func zahlLit(text string) model.Literal { return model.Literal{Art: model.LiteralZahl, Text: text} }
func textLit(text string) model.Literal {
	return model.Literal{Art: model.LiteralZeichenkette, Text: text}
}

func teil(name string, l model.Literal) model.SchluesselTeil {
	return model.SchluesselTeil{Name: name, Literal: l}
}

// entity ist das erwartete Ergebnis von `/<set>(<teile>)`.
func entity(set string, teile ...model.SchluesselTeil) *model.ODataURL {
	return &model.ODataURL{Ressource: model.Ressource{Art: model.RessourceEntity, EntitySet: set, Schluessel: teile}}
}

func wiederhole(s string, n int) string { return strings.Repeat(s, n) }
