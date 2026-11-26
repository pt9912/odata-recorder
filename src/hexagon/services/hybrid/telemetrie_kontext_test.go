package hybrid_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// form ist eine Verpackung eines Fehlers der Stufe.
type form struct {
	name string
	err  error
}

// formen liefert die Verpackungen von e (ein Kontextfehler oder ein Fehler ohne Kontext) allein oder neben x: nackt und mit %w;
// neben x als errors.Join in beiden Folgen, mit %w um ein Join und, wenn x ein *inbound.Fehler ist, als dessen Ursache
// (SPEC-284 Nr. 2: verpackt oder in einem errors.Join, auch neben einem Fehler mit Code).
func formen(e, x error) []form {
	if x == nil {
		return []form{{"nackt", e}, {"%w", fmt.Errorf("stufe: %w", e)}}
	}
	out := []form{
		{"Join(e, x)", errors.Join(e, x)},
		{"Join(x, e)", errors.Join(x, e)},
		{"%w um Join", fmt.Errorf("stufe: %w", errors.Join(x, e))},
	}
	if f, ok := x.(*inbound.Fehler); ok {
		out = append(out, form{"Ursache", &inbound.Fehler{Code: f.Code, Status: f.Status, Ursache: e}})
	}
	return out
}

// TestKontextfehlerKlasse hält D5 und D11 gegen die Klasse (SPEC-321 Nr. 4 b, SPEC-284 Nr. 2): die Menge ist Ort der Stufe (M, R, Senden)
// × Kontextfehler (Canceled, DeadlineExceeded) × Fehler daneben (keiner, ohne Code, mit Code) × Verpackung (höchstens zwei Schritte,
// ein Join trägt einen Kontextfehler) endet den Span von R und von Senden mit abgebrochen, kehrt als derselbe Wert zurück, löst weder den Miss noch den Upstream
// aus, und bei M entsteht kein Span und keine Zählung. Zählprüfung: M 26, R 26, Senden 22 und als Gegenprobe 15 Verpackungen ohne
// Kontextfehler (R 6, Senden 9), die fehler ergeben und bei Senden den Code der Abbildung tragen.
func TestKontextfehlerKlasse(t *testing.T) {
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6003", "ODR-E5001", "ODR-E5003")
	e6001, e6003, miss, e5003 := f[0], f[1], f[2], f[3]
	plain, zweiter := errors.New("anderer Fehler"), errors.New("zweiter Fehler")
	zaehlung := map[string]int{}
	fuehre := func(ort string, w weg) {
		w.lauf(t)
		zaehlung[ort]++
	}
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, x := range []error{nil, plain, e6001, e6003} {
			for _, fm := range formen(e, x) {
				fuehre("M", weg{name: fmt.Sprintf("M %v neben %v, %s", e, x, fm.name), pol: offenste, m: verhalten{err: fm.err},
					ident: fm.err, folge: "Mock"})
			}
		}
		for _, x := range []error{nil, plain, miss, e5003} {
			for _, fm := range formen(e, x) {
				fuehre("R", weg{name: fmt.Sprintf("R %v neben %v, %s", e, x, fm.name), pol: polK, m: verhalten{err: e6003},
					r: verhalten{err: fm.err, body: true}, ident: fm.err, folge: mR + " Replay Close Set(replay=abgebrochen) End(replay)"})
			}
		}
		for _, x := range []error{nil, plain, outbound.ErrUpstreamZeit, outbound.ErrUpstreamNichtErreichbar} {
			for _, fm := range formen(e, x) {
				fuehre("S", weg{name: fmt.Sprintf("Senden %v neben %v, %s", e, x, fm.name), pol: polU, m: verhalten{err: e6001},
					u: verhalten{err: fm.err, body: true}, ident: fm.err, folge: mU + " Senden Close Set(upstream=abgebrochen) End(upstream)"})
			}
		}
	}
	gegen := func(x error) []form {
		return []form{{"%w", fmt.Errorf("stufe: %w", x)}, {"Join(x, y)", errors.Join(x, zweiter)}, {"Join(y, x)", errors.Join(zweiter, x)}}
	}
	for _, x := range []error{plain, e5003} {
		for _, fm := range gegen(x) {
			fuehre("G", weg{name: fmt.Sprintf("R ohne Kontext %v, %s", x, fm.name), pol: polK, m: verhalten{err: e6003},
				r: verhalten{err: fm.err, body: true}, ident: fm.err, folge: mR + " Replay Close Set(replay=fehler) End(replay)"})
		}
	}
	senden := []struct {
		x      error
		code   model.Code
		status int
	}{{plain, model.CodeE9001, 500}, {outbound.ErrUpstreamZeit, model.CodeE4001, 504}, {outbound.ErrUpstreamNichtErreichbar, model.CodeE4001, 502}}
	for _, s := range senden {
		for _, fm := range gegen(s.x) {
			fuehre("G", weg{name: fmt.Sprintf("Senden ohne Kontext %v, %s", s.x, fm.name), pol: polU, m: verhalten{err: e6001},
				u: verhalten{err: fm.err, body: true}, code: s.code, status: s.status,
				folge: mU + " Senden Close Set(upstream=fehler) End(upstream)"})
		}
	}
	if zaehlung["M"] != 26 || zaehlung["R"] != 26 || zaehlung["S"] != 22 || zaehlung["G"] != 15 {
		t.Errorf("Zählprüfung: %v, erwartet M 26, R 26, S 22 und G 15", zaehlung)
	}
}
