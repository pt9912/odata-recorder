package model_test

import (
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestKompatFehlerModell (Zeile 12): sechs Ursachen in der Reihenfolge der Spezifikation; je Ursache neun Prüfungen (Error ohne
// und mit Ort in der Schreibweise von strconv.Quote, Code, der eigene Sentinel, vier Gegen-Sentinel, die Zeichenkette der Ursache);
// drei ungültige Ursachen. Menge 6 × 9 + 3 = 57.
func TestKompatFehlerModell(t *testing.T) {
	soll := []string{"format", "methode", "pfad", "option", "filter", "literal_typ"}
	alle := model.AlleKompatUrsachen()
	if len(alle) != 6 {
		t.Fatalf("%d Ursachen", len(alle))
	}
	n := 0
	for i, u := range alle {
		e := &model.KompatFehler{Ursache: u}
		geprueft := []bool{
			string(u) == soll[i], u.Gueltig(),
			e.Error() == "ODR-E6006: OData-V2-Kompatibilität: "+soll[i],
			(&model.KompatFehler{Ursache: u, Ort: "query/\"3\"\n"}).Error() == "ODR-E6006: OData-V2-Kompatibilität: "+soll[i]+" "+strconv.Quote("query/\"3\"\n"),
			e.Code() == model.CodeE6006, errors.Is(e, model.ErrKompatibilitaet),
			!errors.Is(e, model.ErrURLUngueltig) && !errors.Is(e, model.ErrURLNichtUnterstuetzt),
			!errors.Is(e, model.ErrUngueltig) && !errors.Is(e, model.ErrDatenUngueltig),
			errors.Is(e, e.Unwrap()),
		}
		for j, ok := range geprueft {
			if !ok {
				t.Errorf("Ursache %s, Prüfung %d", u, j)
			}
			n++
		}
	}
	for _, u := range []model.KompatUrsache{"", "Format", "gibt_es_nicht"} {
		if u.Gueltig() {
			t.Errorf("%q gilt als Ursache", u)
		}
		n++
	}
	if n != 57 {
		t.Errorf("%d Prüfungen, erwartet 57", n)
	}
	if !slices.Equal(alle, []model.KompatUrsache{"format", "methode", "pfad", "option", "filter", "literal_typ"}) {
		t.Errorf("Reihenfolge %v", alle)
	}
	// Ein URLFehler ist nie ein KompatFehler und umgekehrt (SPEC-174 Nr. 1).
	var kf *model.KompatFehler
	if errors.As(&model.URLFehler{Ursache: model.URLFilter}, &kf) || errors.Is(&model.URLFehler{Ursache: model.URLFilter}, model.ErrKompatibilitaet) {
		t.Error("ein URLFehler gilt als KompatFehler")
	}
	var uf *model.URLFehler
	if errors.As(&model.KompatFehler{Ursache: model.KompatFilter}, &uf) {
		t.Error("ein KompatFehler gilt als URLFehler")
	}
}
