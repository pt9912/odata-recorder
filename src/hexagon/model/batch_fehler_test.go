package model_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// batchUrsachenSoll sind die sechzehn Ursachen aus SPEC-182 Nr. 4 in der Folge der Tabelle mit ihrer Klasse; das Orakel ist die
// Liste, nie der Quelltext des Modells.
func batchUrsachenSoll() []struct {
	u     string
	nicht bool
} {
	return []struct {
		u     string
		nicht bool
	}{
		{"content_type", false}, {"boundary", false}, {"zu_gross", false}, {"zu_viele", false}, {"rahmen", false}, {"kopfzeile", false},
		{"content_id", false}, {"anfragezeile", false}, {"format", true}, {"teil_typ", true}, {"verschachtelt", true},
		{"transfer_kodierung", true}, {"methode", false}, {"content_id_doppelt", false}, {"referenz", false},
		{"methode_nicht_unterstuetzt", true},
	}
}

// TestBatchFehlerform hält Zeile 28: jede der sechzehn Ursachen trägt ihre Klasse, ihren Code und genau ihren Sentinel; Error()
// nennt Code, Ursache und bei nichtleerem Ort ein Leerzeichen und ihn. Menge 16 Ursachen × 3 Orte × 3 Formen.
func TestBatchFehlerform(t *testing.T) {
	soll := batchUrsachenSoll()
	alle := model.AlleBatchUrsachen()
	if len(alle) != 16 || len(soll) != 16 {
		t.Fatalf("%d Ursachen im Modell, %d im Orakel", len(alle), len(soll))
	}
	fremde := []error{model.ErrUngueltig, model.ErrEingabeUngueltig, model.ErrEingabeNichtUnterstuetzt, model.ErrURLUngueltig,
		model.ErrURLNichtUnterstuetzt, model.ErrDatenUngueltig, model.ErrMetadatenUngueltig}
	n := 0
	for i, u := range alle {
		if string(u) != soll[i].u {
			t.Errorf("Ursache %d heißt %q, erwartet %q", i, u, soll[i].u)
		}
		nicht := soll[i].nicht
		code, kopf := model.CodeE6007, "ODR-E6007: Batch ungültig: "
		sentinel, anderer := model.ErrBatchUngueltig, model.ErrBatchNichtUnterstuetzt
		if nicht {
			code, kopf = model.CodeE6001, "ODR-E6001: Batch nicht unterstützt: "
			sentinel, anderer = anderer, sentinel
		}
		for _, ort := range []string{"", "1", "3/2"} {
			e := &model.BatchFehler{Ursache: u, Ort: ort}
			want := kopf + soll[i].u
			if ort != "" {
				want += " " + ort
			}
			if e.Error() != want {
				t.Errorf("Error() %q, erwartet %q", e.Error(), want)
			}
			if e.Code() != code {
				t.Errorf("%s: Code %v, erwartet %v", u, e.Code(), code)
			}
			for _, form := range []error{e, fmt.Errorf("lesen: %w", e), errors.Join(errors.New("x"), e)} {
				var ziel *model.BatchFehler
				if !errors.As(form, &ziel) || ziel != e || !errors.Is(form, sentinel) || errors.Is(form, anderer) {
					t.Errorf("%s %q: Zeiger oder Sentinel falsch", u, ort)
				}
				for _, f := range fremde {
					if errors.Is(form, f) {
						t.Errorf("%s: trifft fremden Sentinel %v", u, f)
					}
				}
				n++
			}
		}
	}
	if n != 16*3*3 {
		t.Errorf("%d Fälle", n)
	}
}
