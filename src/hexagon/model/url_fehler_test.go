package model_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestURLFehler (Zeile 4): jede der 16 Ursachen trägt den Code ihrer Klasse, errors.Is trifft genau die Sentinel der Klasse, und
// Error() nennt Code, Ursache und den Ort in der Schreibweise von strconv.Quote.
func TestURLFehler(t *testing.T) {
	alle := model.AlleURLUrsachen()
	if len(alle) != 16 {
		t.Fatalf("%d Ursachen", len(alle))
	}
	nicht := map[model.URLUrsache]bool{model.URLServiceDokument: true, model.URLPfadNichtUnterstuetzt: true, model.URLLiteralTyp: true,
		model.URLOptionNichtUnterstuetzt: true, model.URLFilter: true, model.URLWertNichtUnterstuetzt: true}
	for _, u := range alle {
		e := &model.URLFehler{Ursache: u, Ort: "pfad/3"}
		code, praefix, anderes := model.CodeE6005, "ODR-E6005: OData-URL ungültig: ", model.ErrURLNichtUnterstuetzt
		if nicht[u] {
			code, praefix, anderes = model.CodeE6001, "ODR-E6001: OData-URL nicht unterstützt: ", model.ErrURLUngueltig
		}
		if e.Code() != code || !u.Gueltig() || u.NichtUnterstuetzt() != nicht[u] {
			t.Errorf("%s: Code %v", u, e.Code())
		}
		if errors.Is(e, anderes) || !errors.Is(e, e.Unwrap()) || e.Error() != praefix+string(u)+` "pfad/3"` {
			t.Errorf("%s: %q, Unwrap %v", u, e.Error(), e.Unwrap())
		}
		if got := (&model.URLFehler{Ursache: u}).Error(); got != praefix+string(u) {
			t.Errorf("ohne Ort: %q", got)
		}
	}
	if model.URLUrsache("gibt_es_nicht").Gueltig() || model.URLUrsache("").Gueltig() {
		t.Error("eine fremde Ursache ist gültig")
	}
	if strings.Contains((&model.URLFehler{Ursache: model.URLKey, Ort: "a\nb"}).Error(), "\n") {
		t.Error("der Ort ist nicht quotiert")
	}
}
