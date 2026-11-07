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

// ortZeichen sind die fünf Zeichen der Menge M23, aus Bytewerten gebildet (AGENTS.md §3.18): ASCII, zwei, drei und vier Bytes
// im UTF-8 und ein Steuerzeichen, dessen Schreibweise in strconv.Quote vier Zeichen lang ist.
func ortZeichen() [][]byte {
	return [][]byte{{0x61}, {0xC3, 0xA9}, {0xE2, 0x82, 0xAC}, {0xF0, 0x9F, 0x98, 0x80}, {0x01}}
}

// ortVonLaenge bildet einen Ort von genau n Bytes: vorn (n mod Länge des Zeichens) ASCII-Bytes `x`, dahinter das Zeichen so oft, wie
// es passt; ein Zeichen aus mehreren Bytes liegt so an wechselnden Stellen der Grenze von 256 Bytes.
func ortVonLaenge(n int, z []byte) string {
	b := []byte(strings.Repeat("x", n%len(z)))
	b = append(b, []byte(strings.Repeat(string(z), n/len(z)))...)
	return string(b)
}

// sollKurz ist das Orakel zu M23, unabhängig vom Code: die Zeichen von vorn lesen, solange sie ganz in 256 Bytes passen.
func sollKurz(ort string) (string, bool) {
	if len(ort) <= 256 {
		return ort, false
	}
	n := 0
	for _, r := range ort {
		w := len(string(r))
		if n+w > 256 {
			break
		}
		n += w
	}
	return ort[:n], true
}

// TestOrtGekuerzt (W64, F-15, SPEC-180 Nr. 3): Länge 7 × Zeichen 5 = 35; Error() trägt höchstens 256 Bytes des Ortes am Rand eines
// Zeichens, `...` hinter dem schließenden Anführungszeichen nur bei einer Kürzung, ist nie länger als 1280 Bytes; das Feld Ort,
// Code(), Unwrap und errors.Is sind unverändert.
func TestOrtGekuerzt(t *testing.T) {
	laengen := []int{0, 255, 256, 257, 300, 1024, 1 << 20}
	zeichen := ortZeichen()
	if len(laengen) != 7 || len(zeichen) != 5 {
		t.Fatal("Menge M23 ist nicht 7 × 5")
	}
	n := 0
	for _, u := range []model.EingabeUrsache{model.EingabeUnbekannteProperty, model.EingabeOffen} {
		for _, l := range laengen {
			for zi, z := range zeichen {
				ort := ortVonLaenge(l, z)
				if len(ort) != l {
					t.Fatalf("Ort von %d Bytes hat %d", l, len(ort))
				}
				e := &model.EingabeFehler{Ursache: u, Ort: ort}
				kurz, gekuerzt := sollKurz(ort)
				kopf := "ODR-E6007: Request-Body ungültig: "
				if u.NichtUnterstuetzt() {
					kopf = "ODR-E6002: Request-Body nicht unterstützt: "
				}
				want := kopf + string(u)
				if ort != "" {
					want += " " + strconv.Quote(kurz)
				}
				if gekuerzt {
					want += "..."
				}
				if got := e.Error(); got != want || len(got) > 1280 {
					t.Errorf("Länge %d Zeichen %d: Error() hat %d Bytes, erwartet %d (%v)", l, zi, len(got), len(want), got == want)
				}
				if gekuerzt != (l > 256) || (l > 256) != strings.HasSuffix(e.Error(), "...") {
					t.Errorf("Länge %d Zeichen %d: Kürzung %v, `...` %v", l, zi, gekuerzt, strings.HasSuffix(e.Error(), "..."))
				}
				if e.Ort != ort || e.Code().String() != kopf[:9] || !errors.Is(e, e.Unwrap()) {
					t.Errorf("Länge %d Zeichen %d: Ort, Code oder Unwrap verändert", l, zi)
				}
				if u.NichtUnterstuetzt() != errors.Is(e, model.ErrEingabeNichtUnterstuetzt) || errors.Is(e, model.ErrEingabeUngueltig) == u.NichtUnterstuetzt() {
					t.Errorf("Länge %d: errors.Is verändert", l)
				}
				n++
			}
		}
	}
	if n != 2*35 {
		t.Errorf("Menge %d, erwartet 70", n)
	}
	// Goldens von Hand: 85 Zeichen zu drei Bytes (255 Bytes) bei 257 Bytes, ein Zeichen zu vier Bytes an der Grenze von 256 Bytes.
	drei := string([]byte{0xE2, 0x82, 0xAC})
	e := &model.EingabeFehler{Ursache: model.EingabeUnbekannteProperty, Ort: strings.Repeat(drei, 86)}
	if got, want := e.Error(), "ODR-E6007: Request-Body ungültig: unbekannte_property \""+strings.Repeat(drei, 85)+"\"..."; got != want {
		t.Errorf("drei Bytes: %q", got)
	}
	vier := string([]byte{0xF0, 0x9F, 0x98, 0x80})
	e = &model.EingabeFehler{Ursache: model.EingabeUnbekannteProperty, Ort: strings.Repeat("x", 254) + vier}
	if got, want := e.Error(), "ODR-E6007: Request-Body ungültig: unbekannte_property \""+strings.Repeat("x", 254)+"\"..."; got != want {
		t.Errorf("vier Bytes an der Grenze: %q", got)
	}
}
