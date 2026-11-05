package urlparserv2_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/odata/urlparserv2"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// flSchluessel liest lit als Schlüsselliteral `/S(lit)` und liefert das Literal oder das Urteil des Fehlers.
func flSchluessel(t *testing.T, lit string) (model.Literal, flV) {
	t.Helper()
	u, err := urlparserv2.New().ParseURL("/S("+lit+")", "")
	var uf *model.URLFehler
	var kf *model.KompatFehler
	switch {
	case err == nil && len(u.Ressource.Schluessel) == 1:
		return u.Ressource.Schluessel[0].Literal, flV{}
	case errors.As(err, &uf):
		return model.Literal{}, flU(uf.Ursache)
	case errors.As(err, &kf):
		return model.Literal{}, flK(kf.Ursache)
	}
	t.Fatalf("Schlüssel %q: %v", lit, err)
	return model.Literal{}, flV{}
}

// flLiteraleDesSchluessels sind die 122 Formen der Zeilen 28 bis 31 von `slice-v2-query` (35 + 20 + 42 + 25) in der Schreibweise
// der URL.
func flLiteraleDesSchluessels() []string {
	var formen []string
	for _, f := range schluesselLiterale() {
		formen = append(formen, f.eingabe)
	}
	for _, g := range guidAbweichend() {
		formen = append(formen, "guid'"+g+"'")
	}
	formen = append(formen, guidRoh()...)
	for _, f := range datetimeGut() {
		formen = append(formen, "datetime'"+f.eingabe+"'")
	}
	for _, s := range datetimeSchlecht() {
		formen = append(formen, "datetime'"+s+"'")
	}
	return append(formen, ausserhalbLiterale()...)
}

// TestFlLiteraleWieSchluessel (Zeile 18): als `Id eq <Literal>` ergibt jede Form des Schlüssels dieselbe Art und denselben Text
// bzw. dieselbe Ursache wie `/Set(<Literal>)` (`literal_typ` als `*KompatFehler`); sechs Ausnahmen stehen mit Ergebnis und Grund
// in der Liste. Menge 122 = 35 + 20 + 42 + 25.
func TestFlLiteraleWieSchluessel(t *testing.T) {
	formen := flLiteraleDesSchluessels()
	if len(formen) != 122 {
		t.Fatalf("%d Formen, erwartet 122", len(formen))
	}
	// Die Schreibweise `guid` ohne Anführungszeichen ist im Schlüssel `literal`; im Filter beginnt sie wie ein Name und ist `wert`.
	abweichend := map[string]flV{"guid" + guidBasis: flU(model.URLWert)}
	ausnahmen := 0
	for _, f := range formen {
		text, err := url.PathUnescape(f)
		if err != nil {
			t.Fatalf("%q: %v", f, err)
		}
		lit, v := flSchluessel(t, f)
		if w, ok := abweichend[f]; ok {
			v, ausnahmen = w, ausnahmen+1
		}
		if v == (flV{}) {
			flGut(t, "Id eq "+text, flBin("eq", flEig("Id"), model.FilterLiteral{Literal: lit}))
		} else {
			flFehler(t, "Id eq "+text, v)
		}
	}
	if ausnahmen != 1 {
		t.Errorf("%d Ausnahmen in den Formen, erwartet 1", ausnahmen)
	}
	// Die sechs Ausnahmen, die der Schlüssel anders liest: im Filter ein Knoten, ein Name, die Form von V4 oder ein Literal
	// fremden Typs (`grund`).
	for _, a := range []struct {
		text, grund string
		want        flBaum
		v           flV
	}{
		{"null", "Knoten", flNull(), flV{}}, {"yes", "Name", flEig("yes"), flV{}},
		{guidBasis, "Guid ohne Präfix ist die Form von V4", nil, flU(model.URLWert)},
		{"2020-01-02", "Datum ohne Präfix ist die Form von V4", nil, flU(model.URLWert)},
		{"INF", "Literal fremden Typs", nil, flK(model.KompatLiteralTyp)}, {"NaN", "Literal fremden Typs", nil, flK(model.KompatLiteralTyp)},
	} {
		t.Run(a.grund, func(t *testing.T) {
			if a.want != nil {
				flGut(t, "Id eq "+a.text, flBin("eq", flEig("Id"), a.want))
			} else {
				flFehler(t, "Id eq "+a.text, a.v)
			}
		})
	}
}

// TestFlV4FormenOhnePraefix (Zeile 19): ein Literal-Lauf ohne Präfix in der Folge Zahl · `literal_typ` · `wert` (Form von V4) ·
// `literal`. Menge 24 = 15 `wert` + 3 `literal_typ` + 6 `literal`.
func TestFlV4FormenOhnePraefix(t *testing.T) {
	gross := []byte(guidBasis)
	for i, c := range gross {
		if c >= 'a' && c <= 'f' {
			gross[i] = c - 'a' + 'A'
		}
	}
	wert := []string{
		guidBasis, string(gross), "00" + guidBasis[2:], "a0EEbc99-9C0b-4ef8-bb6d-6bb9bd380a11", // Guid-Form, Buchstabe und Ziffer am Anfang
		"2020-01-01", "2020-13-45", "2020-01-01x",
		"2020-01-01T00:00:00Z", "2020-01-01T00:00:00+02:00", "2020-01-01T00:00:00", "2020-01-01T00:00:00.123Z",
		"12:30", "12:30:00", "12:30:00.5", "12:30x",
	}
	typ := []string{"INF", "NaN", "-INF"}
	literal := []string{"2020-1-1", "+1", ".5", "5.", "1e", "0x10"}
	for _, f := range []struct {
		formen []string
		v      flV
	}{{wert, flU(model.URLWert)}, {typ, flK(model.KompatLiteralTyp)}, {literal, flU(model.URLLiteral)}} {
		for _, s := range f.formen {
			flFehler(t, "Id eq "+s, f.v)
			flFehler(t, s+" eq 1", f.v)
		}
	}
	if len(wert) != 15 || len(typ) != 3 || len(literal) != 6 {
		t.Fatalf("%d, %d, %d Formen, erwartet 15, 3, 6", len(wert), len(typ), len(literal))
	}
}
