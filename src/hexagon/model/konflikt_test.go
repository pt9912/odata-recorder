package model_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestKonfliktFehlerCode: Code liefert ODR-E6101 für beide Ursachen, mit und ohne Property (4 Fälle).
func TestKonfliktFehlerCode(t *testing.T) {
	n := 0
	for _, u := range []model.KonfliktUrsache{model.KonfliktWert, model.KonfliktTyp} {
		for _, prop := range []string{"", "Name"} {
			n++
			e := &model.KonfliktFehler{Ursache: u, EntitySet: "Products", Property: prop, Erste: "exc-1", Zweite: "exc-2"}
			if e.Code() != model.CodeE6101 {
				t.Errorf("%s mit Property %q: Code %v, erwartet ODR-E6101", u, prop, e.Code())
			}
		}
	}
	if n != 4 {
		t.Errorf("Menge %d, erwartet 4", n)
	}
}

// TestKonfliktFehlerText: der Text nennt Code, Ursache und die vier Namen in der Schreibweise von strconv.Quote
// (Menge 24: vier Felder mal sechs Zeichenklassen), mit Sollbytes von Hand; ein leerer Name der Property entfällt.
func TestKonfliktFehlerText(t *testing.T) {
	const kopf = "ODR-E6101: Baseline-Import ungültig oder inkonsistent: konflikt "
	klassen := []struct{ name, roh, quote string }{
		{"Anführungszeichen", "\"", `"\""`}, {"Backslash", "\\", `"\\"`}, {"Zeilenvorschub", "\n", `"\n"`},
		{"NUL", "\x00", `"\x00"`}, {"UTF-8-Zeichen", "\xc3\xa9", "\"\xc3\xa9\""}, {"Leerzeichen", " ", `" "`},
	}
	n := 0
	for feld := 0; feld < 4; feld++ {
		for _, k := range klassen {
			n++
			werte := [4]string{"S", "P", "e1", "e2"}
			quote := [4]string{`"S"`, `"P"`, `"e1"`, `"e2"`}
			werte[feld], quote[feld] = k.roh, k.quote
			e := &model.KonfliktFehler{Ursache: model.KonfliktWert, EntitySet: werte[0], Property: werte[1], Erste: werte[2], Zweite: werte[3]}
			soll := kopf + "wert " + quote[0] + " " + quote[1] + " " + quote[2] + " " + quote[3]
			if got := e.Error(); got != soll {
				t.Errorf("Feld %d, %s: %q, erwartet %q", feld, k.name, got, soll)
			}
		}
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
	typ := &model.KonfliktFehler{Ursache: model.KonfliktTyp, EntitySet: "Products", Erste: "exc-1", Zweite: "exc-2"}
	if got, soll := typ.Error(), kopf+`typ "Products" "exc-1" "exc-2"`; got != soll {
		t.Errorf("ohne Property: %q, erwartet %q", got, soll)
	}
}

// TestKonfliktFehlerUnwrap: ErrKonflikt und ErrDatenUngueltig sind erreichbar, auch durch Verpacken, und errors.As
// liefert denselben Zeiger.
func TestKonfliktFehlerUnwrap(t *testing.T) {
	e := &model.KonfliktFehler{Ursache: model.KonfliktTyp, EntitySet: "Products", Erste: "a", Zweite: "b"}
	err := fmt.Errorf("import: %w", e)
	if !errors.Is(err, model.ErrKonflikt) || !errors.Is(err, model.ErrDatenUngueltig) {
		t.Error("ErrKonflikt oder ErrDatenUngueltig nicht erreichbar")
	}
	var k *model.KonfliktFehler
	if !errors.As(err, &k) || k != e {
		t.Error("errors.As liefert nicht denselben Zeiger")
	}
	if errors.Is(err, model.ErrMetadatenUngueltig) {
		t.Error("ErrMetadatenUngueltig erreichbar")
	}
}
