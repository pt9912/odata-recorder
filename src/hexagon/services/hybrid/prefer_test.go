package hybrid_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
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

// preferBodyForm ist eine Form der Anfrage nach `vorbereiten` (SPEC-284 Nr. 5): was der Body ist, welche Policy gilt und ob ein Fallback erlaubt bleibt
// (dann steht der Replay-Fallback hinter dem Fehler des Mock, sonst kommt der Fehler des Mock zum Client).
type preferBodyForm struct {
	name     string
	policy   func(t *testing.T) *policy.Policy
	body     func() (io.Reader, int64)
	fallback bool
}

// preferBodyFormen sind die acht Formen von `vorbereiten`, bei der die Stufen dieselbe Anfrage erhalten sollen: ohne Body, Laenge 0 mit Leser, gepuffert bei
// deklarierter Laenge und bei -1, nicht fallback-fähig, Laenge über der Grenze, Body über der Grenze bei Laenge -1, Lesefehler.
func preferBodyFormen() []preferBodyForm {
	voll := func(t *testing.T) *policy.Policy { return offenste(t) }
	bytes4 := func(laenge int64) func() (io.Reader, int64) {
		return func() (io.Reader, int64) { return neuerStrom(muster(4, 3)), laenge }
	}
	return []preferBodyForm{
		{"ohne Body", voll, func() (io.Reader, int64) { return nil, 0 }, true},
		{"Laenge 0 mit Leser", voll, bytes4(0), true},
		{"gepuffert, Laenge deklariert", voll, bytes4(4), true},
		{"gepuffert, Laenge -1", voll, bytes4(-1), true},
		{"nicht fallback-fähig", func(t *testing.T) *policy.Policy { return mitKategorien(t) }, bytes4(4), false},
		{"Laenge über der Grenze", voll, func() (io.Reader, int64) { return neuerStrom(muster(standardMax+1, 3)), standardMax + 1 }, false},
		{"Body über der Grenze, Laenge -1", voll, func() (io.Reader, int64) { return neuerStrom(muster(standardMax+1, 3)), -1 }, false},
		{"Lesefehler", voll, func() (io.Reader, int64) { return &fehlerNach{rest: muster(3, 3), err: errors.New("Lesefehler")}, -1 }, false},
	}
}

// TestPreferGehtUnveraendertInDenFallback (Z24, SPEC-287 Nr. 6, SPEC-284 Nr. 5, 8 c): im Modus `hybrid` geht `Prefer` mit jeder Zeile und unter jedem Namen unverändert in
// den Mock und, wo die Anfrage einen Fallback behält, in den Fallback, und der Dienst setzt oder streicht `Preference-Applied` nie: die Antwort des Fallbacks trägt
// den Header genau, wenn die Stufe ihn trug, die des Mock erreicht den Client als dieselbe Map. Form des Bodys 8 × Wert 172 × Name 2 × Methode 5 × Header der Stufe 2;
// ohne Fallback kommt der Fehler des Mock zum Client und der Replay läuft nicht.
func TestPreferGehtUnveraendertInDenFallback(t *testing.T) {
	e6002 := code(codeNamed(t, "ODR-E6002"))
	n := 0
	for _, form := range preferBodyFormen() {
		pol := form.policy(t)
		for _, zeilen := range preferZeilen() {
			for _, name := range []string{"Prefer", "prefer"} {
				for _, methode := range []string{"POST", "PATCH", "PUT", "DELETE", "GET"} {
					for _, applied := range [][]string{nil, {"return=minimal"}} {
						n++
						g := neuesRig(t, pol, standardMax)
						g.mockFehler(e6002)
						h := map[string][]string{"Content-Type": {"application/json"}}
						if applied != nil {
							h["Preference-Applied"] = applied
						}
						want := inbound.Antwort{Status: 200, Header: h, Body: neuerStrom([]byte("aufgezeichnet"))}
						g.replayAntwort(want)
						a := anfrage(methode)
						a.Body, a.Laenge = form.body()
						if zeilen != nil {
							a.Header[name] = zeilen
						}
						got, err := g.v.Hybrid(context.Background(), a)
						kontext := fmt.Sprintf("%s, %s %s %q", form.name, methode, name, zeilen)
						stufen := map[string]*stufe{"Mock": g.m}
						if form.fallback {
							stufen["Fallback"] = g.r
							if err != nil || got.Status != 200 || got.Body != want.Body {
								t.Errorf("%s: Antwort %+v, Fehler %v", kontext, got, err)
								continue
							}
							if _, da := got.Header["Preference-Applied"]; da != (applied != nil) || !slices.Equal(got.Header["Preference-Applied"], applied) {
								t.Errorf("%s: Preference-Applied %v, die Stufe trug %v", kontext, got.Header["Preference-Applied"], applied)
							}
							if k := kennzeichen(got); !reflect.DeepEqual(k, []string{"replay"}) {
								t.Errorf("%s: Kennzeichnung %v", kontext, k)
							}
						} else {
							var f *inbound.Fehler
							if !errors.As(err, &f) || f != e6002 || g.r.zahl() != 0 {
								t.Errorf("%s: Fehler %v, Replay %d-mal gerufen, erwartet der Fehler des Mock ohne Fallback", kontext, err, g.r.zahl())
							}
						}
						for stelle, s := range stufen {
							if s.zahl() == 0 {
								t.Errorf("%s: %s nie gerufen", kontext, stelle)
								continue
							}
							if p := s.erster(t).a.Header[name]; !reflect.DeepEqual(p, zeilen) {
								t.Errorf("%s: %s erhielt %q", kontext, stelle, p)
							}
						}
					}
				}
			}
		}
	}
	if n != 8*172*2*5*2 {
		t.Errorf("Menge %d, erwartet %d", n, 8*172*2*5*2)
	}
}

// TestPreferGehtUnveraendertInDenUpstream (Z24, SPEC-287 Nr. 6): wo die Anfrage zum Upstream weitergeht, erhält er jede Zeile von `Prefer` unter jedem Namen unverändert, auch aus dem
// gepufferten Body. Weg (nach M: GET, HEAD, POST; nach R: GET, HEAD) × Body (ohne, 4 Bytes mit Laenge 4, mit Laenge -1) × Wert 172 × Name 2: 5 · 3 · 172 · 2 = 5160 Fälle.
func TestPreferGehtUnveraendertInDenUpstream(t *testing.T) {
	bodies := []struct {
		name    string
		laenge  int64
		mitBody bool
	}{{"ohne Body", 0, false}, {"Laenge 4", 4, true}, {"Laenge -1", -1, true}}
	wege := []struct {
		name     string
		methoden []string
		neu      func() *rig
	}{
		{"nach M", []string{"GET", "HEAD", "POST"}, func() *rig {
			g := neuesRigZiel(t, uPolicy(t, []string{"h"}, []string{"POST " + standardP}), standardMax, zielH)
			g.mockE6003(t)
			return g
		}},
		{"nach R", []string{"GET", "HEAD"}, func() *rig {
			g := neuesRigZiel(t, uPolicy(t, []string{"h"}, []string{"POST " + standardP}, katCoverage), standardMax, zielH)
			g.mockE6003(t)
			g.replayFehler(code(codeNamed(t, missCode)))
			return g
		}},
	}
	n, soll := 0, 0
	for _, w := range wege {
		soll += len(w.methoden) * len(bodies) * 172 * 2
		for _, m := range w.methoden {
			for _, b := range bodies {
				for _, zeilen := range preferZeilen() {
					for _, name := range []string{"Prefer", "prefer"} {
						n++
						g := w.neu()
						a := anfrage(m)
						if b.mitBody {
							a = mitBody(m, neuerStrom(muster(4, 3)), b.laenge)
						}
						if zeilen != nil {
							a.Header[name] = zeilen
						}
						kontext := fmt.Sprintf("%s %s, %s, %s %q", w.name, m, b.name, name, zeilen)
						if _, err := g.v.Hybrid(context.Background(), a); err != nil {
							t.Errorf("%s: %v", kontext, err)
							continue
						}
						if g.u.zahl() != 1 {
							t.Errorf("%s: Upstream %d-mal gerufen", kontext, g.u.zahl())
							continue
						}
						if p := g.u.erster(t).a.Header[name]; !reflect.DeepEqual(p, zeilen) {
							t.Errorf("%s: Upstream erhielt %q", kontext, p)
						}
					}
				}
			}
		}
	}
	if n != soll || n != 5*3*172*2 {
		t.Errorf("Menge %d, erwartet %d", n, 5*3*172*2)
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
