package model_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// eingabeUrsachenSoll sind die neunzehn Ursachen aus SPEC-180 Nr. 3 im Wortlaut, die ersten fünfzehn der Klasse `ungueltig`, die letzten vier
// der Klasse `nicht_unterstuetzt`; das Orakel ist die Liste, nie der Quelltext des Modells.
func eingabeUrsachenSoll() []string {
	return []string{"syntax", "kodierung", "tiefe", "zu_gross", "doppelter_name", "content_type", "struktur", "odata_typ", "unbekannte_property",
		"typ", "wert", "null", "zahl_zu_lang", "schluessel", "fehlend", "property_typ", "navigation", "bind", "offen"}
}

// TestEingabeFehlerForm hält Zeile 35: neunzehn Ursachen; `errors.Is` trifft genau den Sentinel der Klasse, `errors.As` den Zeiger, `Code()` der Klasse
// `nicht_unterstuetzt` ist ODR-E6002, der der Klasse `ungueltig` ODR-E6007; `Error()` nennt Code, Ursache und den Ort in der Schreibweise von
// `strconv.Quote`. Menge 19 × 2 Sentinel = 38.
func TestEingabeFehlerForm(t *testing.T) {
	soll := eingabeUrsachenSoll()
	alle := model.AlleEingabeUrsachen()
	if len(alle) != 19 || len(soll) != 19 {
		t.Fatalf("%d Ursachen im Modell, %d im Orakel, erwartet 19", len(alle), len(soll))
	}
	n := 0
	for i, u := range alle {
		if string(u) != soll[i] {
			t.Errorf("Ursache %d heißt %q, erwartet %q", i, u, soll[i])
		}
		nicht := i >= 15
		for _, ort := range []string{"", "/a b", "/a\x01b"} {
			e := &model.EingabeFehler{Ursache: u, Ort: ort}
			for _, form := range []error{e, fmt.Errorf("lesen: %w", e)} {
				var ziel *model.EingabeFehler
				if !errors.As(form, &ziel) || ziel != e {
					t.Errorf("%s: errors.As trifft nicht", u)
				}
				for _, s := range []struct {
					err  error
					wahr bool
				}{{model.ErrEingabeNichtUnterstuetzt, nicht}, {model.ErrEingabeUngueltig, !nicht}} {
					if errors.Is(form, s.err) != s.wahr {
						t.Errorf("%s %q: errors.Is(%v) ist nicht %v", u, ort, s.err, s.wahr)
					}
					n++
				}
			}
			code, kopf := "ODR-E6007", "ODR-E6007: Request-Body ungültig: "
			if nicht {
				code, kopf = "ODR-E6002", "ODR-E6002: Request-Body nicht unterstützt: "
			}
			want := kopf + string(u)
			if ort != "" {
				want += " " + strconv.Quote(ort)
			}
			if e.Code().String() != code || e.Error() != want {
				t.Errorf("%s: Code %v Error() %q, erwartet %s %q", u, e.Code(), e.Error(), code, want)
			}
		}
		if nicht != u.NichtUnterstuetzt() {
			t.Errorf("%s: NichtUnterstuetzt() ist %v", u, !nicht)
		}
	}
	if n != 19*3*2*2 {
		t.Errorf("Prüfungen %d", n)
	}
	if !slices.Contains(soll, "content_type") {
		t.Error("content_type fehlt im Orakel")
	}
}

// TestEingabeArt hält die drei Konstanten von SPEC-180: nur `POST`, `PATCH` und `PUT` in dieser Schreibweise sind gültig.
func TestEingabeArt(t *testing.T) {
	for _, a := range []model.EingabeArt{"POST", "PATCH", "PUT"} {
		if !a.Gueltig() {
			t.Errorf("%q ungültig", a)
		}
	}
	// Klasse: je Konstante sieben Schreibweisen (klein, Anfangsbuchstabe groß, Leerzeichen dahinter und davor, Tabulator, ein Zeichen zu viel, eines zu wenig) = 21,
	// dazu `DELETE`, `GET` und leer.
	n := 0
	for _, k := range []string{"POST", "PATCH", "PUT"} {
		kl := strings.ToLower(k)
		for _, v := range []string{kl, strings.ToUpper(kl[:1]) + kl[1:], k + " ", " " + k, k + "\t", k + "X", k[:len(k)-1]} {
			if model.EingabeArt(v).Gueltig() {
				t.Errorf("%q gültig", v)
			}
			n++
		}
	}
	for _, a := range []model.EingabeArt{"DELETE", "GET", ""} {
		if a.Gueltig() {
			t.Errorf("%q gültig", a)
		}
		n++
	}
	if n != 24 {
		t.Errorf("Menge %d, erwartet 24", n)
	}
	if model.EingabePost != "POST" || model.EingabePatch != "PATCH" || model.EingabePut != "PUT" {
		t.Error("Konstanten tragen nicht den Text der Methode")
	}
}
