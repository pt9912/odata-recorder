package hybrid_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// TestHeaderUnveraendert hält Z17: M und R erhalten dieselbe Header-Map mit allen Namen, auch X-Recorder-*, und die
// Map der Anfrage ist nach dem Aufruf unverändert (Tiefenvergleich gegen die Kopie davor), in jedem Pfad.
func TestHeaderUnveraendert(t *testing.T) {
	e6003 := code(codeNamed(t, "ODR-E6003"))
	pfade := map[string]func(g *rig){
		"Erfolg":       func(g *rig) {},
		"Replay":       func(g *rig) { g.mockFehler(e6003); g.replayAntwort(antwortVon(200, nil)) },
		"Fehler von R": func(g *rig) { g.mockFehler(e6003); g.replayFehler(code(codeNamed(t, "ODR-E5001"))) },
		"ohne Fallback": func(g *rig) {
			g.mockFehler(code(codeNamed(t, "ODR-E6005")))
		},
	}
	for name, einrichten := range pfade {
		for _, methode := range []string{"GET", "POST"} {
			g := neuesRig(t, offenste(t), standardMax)
			einrichten(g)
			a := anfrage(methode)
			a.Header["X-Recorder-Fallback"] = []string{"vorbelegt"}
			vorher := tiefeKopie(a.Header)
			_, _ = g.v.Hybrid(context.Background(), a)
			if !reflect.DeepEqual(a.Header, vorher) {
				t.Errorf("%s %s: Header der Anfrage verändert: %v", name, methode, a.Header)
			}
			for stelle, s := range map[string]*stufe{"M": g.m, "R": g.r} {
				if s.zahl() == 0 {
					continue
				}
				if got := s.erster(t).a.Header; zeiger(got) != zeiger(a.Header) || !reflect.DeepEqual(got, vorher) {
					t.Errorf("%s %s %s: nicht dieselbe Map mit allen Namen: %v", name, methode, stelle, got)
				}
			}
		}
	}
}

// TestSessionHeaderErstImReplay hält Z18: ein Erfolg des Mock ohne die Zuordnungs-Header wird ausgeliefert (der Dienst
// prüft die Header nicht vor M), und ODR-E5003 des Replay-Dienstes kehrt unverändert zurück.
func TestSessionHeaderErstImReplay(t *testing.T) {
	g := neuesRig(t, offenste(t), standardMax)
	want := antwortVon(200, []byte("ok"))
	g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return want, nil }
	a := anfrage("GET")
	a.Header = nil
	got, err := g.v.Hybrid(context.Background(), a)
	if err != nil || got.Body != want.Body {
		t.Errorf("Erfolg ohne Header: %+v, %v", got, err)
	}
	g.zaehlen(t, "Erfolg ohne Header", 1, 0)

	g = neuesRig(t, offenste(t), standardMax)
	g.mockFehler(code(codeNamed(t, "ODR-E6003")))
	fehler := code(codeNamed(t, "ODR-E5003"))
	g.replayFehler(fehler)
	a = anfrage("GET")
	a.Header = map[string][]string{"X-Recorder-Workspace": {"ws"}}
	if _, err := g.v.Hybrid(context.Background(), a); err != error(fehler) {
		t.Errorf("ODR-E5003 des Replay: %v", err)
	}
	g.zaehlen(t, "E5003", 1, 1)
}

// TestKennzeichnung hält Z26: eine Antwort aus R trägt X-Recorder-Fallback mit genau einem Wert replay unter genau
// einem Namen, ein Header gleichen Namens der Stufe in jeder ASCII-Schreibweise (und mehreren Werten) wird ersetzt,
// ein Name mit dem Zeichen KELVIN SIGN statt k bleibt; die Map der Stufe bleibt unverändert, Status und Body gehen
// durch; aus M kommt die Antwort ohne Kennzeichnung und unverändert.
func TestKennzeichnung(t *testing.T) {
	kelvin := "X-Recorder-Fallbac" + string([]byte{0xE2, 0x84, 0xAA})
	namen := []string{"X-Recorder-Fallback", "x-recorder-fallback", "X-RECORDER-FALLBACK", "X-recorder-Fallback", "x-Recorder-fallBack"}
	vorbelegungen := []map[string][]string{nil, {}, {"Content-Type": {"a"}}}
	for _, n := range namen {
		vorbelegungen = append(vorbelegungen, map[string][]string{n: {"upstream"}, "Content-Type": {"a", "b"}})
		vorbelegungen = append(vorbelegungen, map[string][]string{n: {"x", "y"}, "Set-Cookie": {"1", "2"}})
	}
	vorbelegungen = append(vorbelegungen,
		map[string][]string{"x-recorder-fallback": {"1"}, "X-Recorder-Fallback": {"2"}},
		map[string][]string{kelvin: {"bleibt"}},
		map[string][]string{kelvin: {"bleibt"}, "X-Recorder-Fallback": {"alt"}},
	)
	gezaehlt := 0
	for i, vor := range vorbelegungen {
		for _, stufeName := range []string{"M", "R"} {
			gezaehlt++
			g := neuesRig(t, offenste(t), standardMax)
			ant := inbound.Antwort{Status: 202, Header: tiefeKopie(vor), Body: neuerStrom(nil)}
			vorher := tiefeKopie(ant.Header)
			var got inbound.Antwort
			var err error
			if stufeName == "M" {
				g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) { return ant, nil }
			} else {
				g.mockFehler(code(codeNamed(t, "ODR-E6003")))
				g.replayAntwort(ant)
			}
			got, err = g.v.Hybrid(context.Background(), anfrage("GET"))
			if err != nil || got.Status != 202 || got.Body != ant.Body {
				t.Errorf("Fall %d %s: %+v, %v", i, stufeName, got, err)
				continue
			}
			if !reflect.DeepEqual(ant.Header, vorher) {
				t.Errorf("Fall %d %s: Map der Stufe verändert: %v", i, stufeName, ant.Header)
			}
			if stufeName == "M" {
				if zeiger(got.Header) != zeiger(ant.Header) || !reflect.DeepEqual(got.Header, vorher) {
					t.Errorf("Fall %d M: Header verändert: %v", i, got.Header)
				}
				continue
			}
			pruefeKennzeichnung(t, i, got.Header, ant.Header, vorher)
		}
	}
	if gezaehlt != 2*len(vorbelegungen) || len(vorbelegungen) != 3+2*len(namen)+3 {
		t.Errorf("Zählprüfung: %d Fälle", gezaehlt)
	}
}

func pruefeKennzeichnung(t *testing.T, i int, got, stufe, vorher map[string][]string) {
	t.Helper()
	erwartet := map[string][]string{}
	for n, w := range vorher {
		if !asciiGleich(n, "X-Recorder-Fallback") {
			erwartet[n] = w
		}
	}
	erwartet["X-Recorder-Fallback"] = []string{"replay"}
	if !reflect.DeepEqual(got, erwartet) {
		t.Errorf("Fall %d R: Header %v, erwartet %v", i, got, erwartet)
	}
	if zeiger(got) == zeiger(stufe) && stufe != nil {
		t.Errorf("Fall %d R: dieselbe Map wie die Stufe", i)
	}
	for n, w := range stufe {
		if g := got[n]; len(w) > 0 && len(g) > 0 && &g[0] == &w[0] {
			t.Errorf("Fall %d R: Wertliste %q mit der Stufe geteilt", i, n)
		}
	}
}

// TestStromSchliessen hält Z28: liefert M oder R Fehler und Antwort mit Body zugleich, schließt der Dienst den Body genau
// einmal und liefert den Fehler (auch Kontextfehler und Fehler ohne Code) unverändert; ausgelieferte Antworten schließt
// er nie.
func TestStromSchliessen(t *testing.T) {
	e6003 := code(codeNamed(t, "ODR-E6003"))
	fehlerArten := map[string]error{
		"ohne Code":    errors.New("x"),
		"ErrUngueltig": model.ErrUngueltig,
		"Canceled":     context.Canceled,
		"Deadline":     fmt.Errorf("v: %w", context.DeadlineExceeded),
		"E6003":        e6003,
		"E6005":        code(codeNamed(t, "ODR-E6005")),
		"E5001":        code(codeNamed(t, "ODR-E5001")),
		"Join":         errors.Join(e6003, context.Canceled),
	}
	for name, fehler := range fehlerArten {
		g := neuesRig(t, offenste(t), standardMax)
		sm, sr := neuerStrom([]byte("m")), neuerStrom([]byte("r"))
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{Status: 500, Body: sm}, fehler
		}
		g.replayAntwort(inbound.Antwort{Status: 200, Body: sr})
		got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
		replayLaeuft := fehler == error(e6003)
		if replayLaeuft != (err == nil) || !replayLaeuft && !sameErr(err, fehler) {
			t.Errorf("M %s: Fehler %v", name, err)
		}
		if sm.closes.Load() != 1 || sr.closes.Load() != 0 || replayLaeuft && got.Body != sr {
			t.Errorf("M %s: Close von M %d (erwartet 1), von R %d (erwartet 0)", name, sm.closes.Load(), sr.closes.Load())
		}

		g = neuesRig(t, offenste(t), standardMax)
		sm, sr = neuerStrom([]byte("m")), neuerStrom([]byte("r"))
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{Status: 501, Body: sm}, e6003
		}
		g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{Status: 500, Body: sr}, fehler
		}
		if _, err := g.v.Hybrid(context.Background(), anfrage("GET")); !sameErr(err, fehler) {
			t.Errorf("R %s: Fehler %v", name, err)
		}
		if sm.closes.Load() != 1 || sr.closes.Load() != 1 {
			t.Errorf("R %s: Close von M %d, von R %d; erwartet je 1", name, sm.closes.Load(), sr.closes.Load())
		}
	}
	if len(fehlerArten) != 8 {
		t.Errorf("Zählprüfung: %d Fehlerarten", len(fehlerArten))
	}
	g := neuesRig(t, offenste(t), standardMax)
	ausgeliefert := neuerStrom([]byte("m"))
	g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
		return inbound.Antwort{Status: 200, Body: ausgeliefert}, nil
	}
	if got, err := g.v.Hybrid(context.Background(), anfrage("GET")); err != nil || got.Body != io.ReadCloser(ausgeliefert) {
		t.Fatalf("Erfolg des Mock: %+v, %v", got, err)
	}
	if ausgeliefert.closes.Load() != 0 || ausgeliefert.reads.Load() != 0 {
		t.Errorf("ausgelieferte Antwort des Mock: %d-mal geschlossen, %d-mal gelesen", ausgeliefert.closes.Load(), ausgeliefert.reads.Load())
	}
}
