package hybrid_test

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// preferZeilen liefert jede Folge der Länge 0 bis 2 aus neun Elementen von `Prefer` in jeder Verteilung auf Kopfzeilen (je Lücke dieselbe Zeile mit `,` oder
// die nächste Zeile): 1 + 9 + 81 · 2 = 172 Werte des Headers.
func preferZeilen() [][]string {
	elemente := []string{"return=minimal", "return=representation", "return=foo", "return", "return=", "respond-async", "odata.x=1", "return@x", ""}
	var r [][]string
	r = append(r, nil)
	for _, a := range elemente {
		r = append(r, []string{a})
		for _, b := range elemente {
			r = append(r, []string{a + "," + b}, []string{a, b})
		}
	}
	return r
}

// TestPreferGehtUnveraendertInDenFallback (Z24, SPEC-287 Nr. 6, SPEC-284 Nr. 8 c): im Modus `hybrid` geht `Prefer` mit jeder Zeile und unter jedem Namen unverändert in
// den Mock und in den Fallback, und der Dienst setzt oder streicht `Preference-Applied` nie: die Antwort des Fallbacks trägt den Header genau, wenn die Stufe ihn
// trug, die des Mock erreicht den Client als dieselbe Map. Wert 172 × Name 2 × Methode 5 × Header der Stufe 2.
func TestPreferGehtUnveraendertInDenFallback(t *testing.T) {
	e6002 := code(codeNamed(t, "ODR-E6002"))
	n := 0
	for _, zeilen := range preferZeilen() {
		for _, name := range []string{"Prefer", "prefer"} {
			for _, methode := range []string{"POST", "PATCH", "PUT", "DELETE", "GET"} {
				for _, applied := range [][]string{nil, {"return=minimal"}} {
					n++
					g := neuesRig(t, offenste(t), standardMax)
					g.mockFehler(e6002)
					h := map[string][]string{"Content-Type": {"application/json"}}
					if applied != nil {
						h["Preference-Applied"] = applied
					}
					want := inbound.Antwort{Status: 200, Header: h, Body: neuerStrom([]byte("aufgezeichnet"))}
					g.replayAntwort(want)
					a := anfrage(methode)
					if zeilen != nil {
						a.Header[name] = zeilen
					}
					got, err := g.v.Hybrid(context.Background(), a)
					kontext := methode + " " + name
					if err != nil || got.Status != 200 || got.Body != want.Body {
						t.Errorf("%s %q: Antwort %+v, Fehler %v", kontext, zeilen, got, err)
						continue
					}
					for stelle, s := range map[string]*stufe{"Mock": g.m, "Fallback": g.r} {
						if p := s.erster(t).a.Header[name]; !reflect.DeepEqual(p, zeilen) {
							t.Errorf("%s %q: %s erhielt %q", kontext, zeilen, stelle, p)
						}
					}
					if _, da := got.Header["Preference-Applied"]; da != (applied != nil) || !slices.Equal(got.Header["Preference-Applied"], applied) {
						t.Errorf("%s %q: Preference-Applied %v, die Stufe trug %v", kontext, zeilen, got.Header["Preference-Applied"], applied)
					}
					if k := kennzeichen(got); !reflect.DeepEqual(k, []string{"replay"}) {
						t.Errorf("%s %q: Kennzeichnung %v", kontext, zeilen, k)
					}
				}
			}
		}
	}
	if n != 172*2*5*2 {
		t.Errorf("Menge %d, erwartet %d", n, 172*2*5*2)
	}
}

// TestPreferErfolgDesMockUnveraendert (Z24): eine Antwort des Mock mit `Preference-Applied` erreicht den Client als dieselbe Map mit demselben Wert, der Replay läuft nicht.
func TestPreferErfolgDesMockUnveraendert(t *testing.T) {
	for _, applied := range []string{"return=minimal", "return=representation"} {
		g := neuesRig(t, offenste(t), standardMax)
		want := inbound.Antwort{Status: 200, Header: map[string][]string{"Preference-Applied": {applied}}, Body: neuerStrom(nil)}
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return want, nil }
		a := anfrage("PATCH")
		a.Header["Prefer"] = []string{applied}
		got, err := g.v.Hybrid(context.Background(), a)
		if err != nil || zeiger(got.Header) != zeiger(want.Header) || !slices.Equal(got.Header["Preference-Applied"], []string{applied}) {
			t.Errorf("%s: Antwort %+v, Fehler %v", applied, got, err)
		}
		g.zaehlen(t, applied, 1, 0)
	}
}
