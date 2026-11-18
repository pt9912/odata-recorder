package hybrid_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
)

// TestUpstreamLeserJederWeg hält Z12 (U) und Z24 für die Klasse „eigener, vollständiger Leser für U“ (SPEC-284 Nr. 5 c,
// 8 b) an jeder Aufrufstelle von U: Weg nach M (GET, HEAD, POST mit Route) und Weg nach R (GET, HEAD; mutierend führt
// nach einem Miss nie zu U) × Deklaration der Laenge -1, gleich und kleiner als die Bytes. M und R lesen ihren Leser
// vollständig; U erhält alle Bytes, die Laenge des Puffers und weder den Leser von M noch den von R.
func TestUpstreamLeserJederWeg(t *testing.T) {
	inhalt := muster(5, 9)
	wegeBody := []struct {
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
	n := 0
	for _, w := range wegeBody {
		for _, m := range w.methoden {
			for _, decl := range []int64{-1, 5, 3} {
				n++
				g := w.neu()
				g.m.lies, g.r.lies = liesAlles, liesAlles
				if _, err := g.v.Hybrid(context.Background(), mitBody(m, neuerStrom(inhalt), decl)); err != nil {
					t.Errorf("%s %s %d: %v", w.name, m, decl, err)
					continue
				}
				u := g.u.erster(t)
				if g.u.zahl() != 1 || !reflect.DeepEqual(u.gelesen, inhalt) || u.a.Laenge != 5 || u.a.Body == nil {
					t.Errorf("%s %s %d: U %d-mal, gelesen %v, Laenge %d", w.name, m, decl, g.u.zahl(), u.gelesen, u.a.Laenge)
					continue
				}
				if u.a.Body == g.m.erster(t).a.Body || (g.r.zahl() > 0 && u.a.Body == g.r.erster(t).a.Body) {
					t.Errorf("%s %s %d: U erhält den Leser einer anderen Stufe", w.name, m, decl)
				}
			}
		}
	}
	if n != 3*3+2*3 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestDienstwahlKeinDienst hält Z19 (kein Dienst) für beide Seiten der Grenze des Roots / : ein Pfad ohne führenden / (leer,
// ohne Slash, *, Teil eines Pfads, Escape, Leerzeichen davor) wählt keinen Dienst, auch nicht den Root /, in jeder Liste;
// Ziel ist leer, U läuft nie, und der Fehler von M kehrt zurück.
func TestDienstwahlKeinDienst(t *testing.T) {
	listen := [][]hybrid.Dienst{
		{{Root: "/", UpstreamURL: "http://a"}},
		{{Root: "/", UpstreamURL: "http://a"}, {Root: "/a", UpstreamURL: "http://b"}},
		{{Root: "/a", UpstreamURL: "http://b"}, {Root: "/", UpstreamURL: "http://a"}},
		{{Root: "/odata/svc", UpstreamURL: "http://a"}},
	}
	pfade := []string{"", "odata", "*", "a", "a/b", "odata/svc", "%2F", " /a"}
	n := 0
	for i, liste := range listen {
		for _, pfad := range pfade {
			n++
			g := neuesRigMit(t, uPolicy(t, []string{"a", "b"}, nil), standardMax, liste)
			g.mockE6003(t)
			a := anfrage("GET")
			a.Pfad = pfad
			if _, err := g.v.Hybrid(context.Background(), a); err == nil || g.u.zahl() != 0 {
				t.Errorf("Liste %d, Pfad %q: Fehler %v, U %d; erwartet kein Dienst", i, pfad, err, g.u.zahl())
			}
		}
	}
	if n != len(listen)*len(pfade) {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestDienstwahlOhneURL hält Z19 (Ziel des gewählten Dienstes) für Listen mit und ohne UpstreamURL gemischt: gewählt
// wird nach dem Root, nicht nach der URL; ein gewählter Dienst ohne UpstreamURL ergibt das leere Ziel und kein U, auch
// wenn ein kürzerer oder anderer Dienst eine URL trägt; in beiden Reihenfolgen der Liste.
func TestDienstwahlOhneURL(t *testing.T) {
	dienste := []hybrid.Dienst{
		{Root: "/a", UpstreamURL: "http://a"}, {Root: "/a/b"}, {Root: "/"}, {Root: "/ab", UpstreamURL: "http://abx"},
		{Root: "/c"}, {Root: "/c/d", UpstreamURL: "http://cd"},
	}
	faelle := []struct{ pfad, ziel string }{
		{"/a/b", ""}, {"/a/b/c", ""}, {"/a/x", "http://a"}, {"/q", ""}, {"/ab", "http://abx"}, {"/c", ""},
		{"/c/d/e", "http://cd"}, {"/c/x", ""},
	}
	n := 0
	for _, umgekehrt := range []bool{false, true} {
		liste := append([]hybrid.Dienst(nil), dienste...)
		if umgekehrt {
			for i, j := 0, len(liste)-1; i < j; i, j = i+1, j-1 {
				liste[i], liste[j] = liste[j], liste[i]
			}
		}
		for _, f := range faelle {
			n++
			g := neuesRigMit(t, uPolicy(t, []string{"a", "abx", "cd"}, nil), standardMax, liste)
			g.mockE6003(t)
			a := anfrage("GET")
			a.Pfad = f.pfad
			_, err := g.v.Hybrid(context.Background(), a)
			if f.ziel == "" {
				if err == nil || g.u.zahl() != 0 {
					t.Errorf("%s (umgekehrt %v): Fehler %v, U %d; erwartet kein U", f.pfad, umgekehrt, err, g.u.zahl())
				}
				continue
			}
			if err != nil || g.u.zahl() != 1 || g.u.erster(t).a.Ziel != f.ziel {
				t.Errorf("%s (umgekehrt %v): Fehler %v, U %d", f.pfad, umgekehrt, err, g.u.zahl())
			}
		}
	}
	if n != 16 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestUpstreamAbbruchKlassen hält Z10 (U) für die Klasse „Kontext beendet vor U“ (SPEC-284 Nr. 3 e): Weg nach M und Weg
// nach R × Ende durch Abbruch und durch Ablauf der Frist; das Ergebnis ist ctx.Err() derselben Art ohne Code, U läuft
// nie. Die Stufe, die den Kontext beendet, wartet beim Ablauf auf ctx.Done().
func TestUpstreamAbbruchKlassen(t *testing.T) {
	enden := []struct {
		name string
		neu  func() (context.Context, func())
		ist  error
	}{
		{"Abbruch", func() (context.Context, func()) {
			ctx, cancel := context.WithCancel(context.Background())
			return ctx, cancel
		}, context.Canceled},
		{"Frist", func() (context.Context, func()) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			return ctx, func() { <-ctx.Done(); cancel() }
		}, context.DeadlineExceeded},
	}
	n := 0
	for _, e := range enden {
		for _, weg := range []string{"nach M", "nach R"} {
			n++
			ctx, beende := e.neu()
			defer beende()
			var g *rig
			if weg == "nach M" {
				g = neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
				g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
					beende()
					return inbound.Antwort{}, code(codeNamed(t, mockCode))
				}
			} else {
				g = neuesRigZiel(t, uPolicy(t, []string{"h"}, nil, katCoverage), standardMax, zielH)
				g.mockE6003(t)
				g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
					beende()
					return inbound.Antwort{}, code(codeNamed(t, missCode))
				}
			}
			_, err := g.v.Hybrid(ctx, anfrage("GET"))
			var f *inbound.Fehler
			if err != ctx.Err() || !errors.Is(err, e.ist) || errors.As(err, &f) || g.u.zahl() != 0 {
				t.Errorf("%s %s: Fehler %v, U %d; erwartet %v ohne Code", e.name, weg, err, g.u.zahl(), e.ist)
			}
		}
	}
	if n != 4 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// bytesVon bildet einen Text aus Bytewerten (AGENTS.md §3.18): kein Escape im Quelltext.
func bytesVon(b ...byte) string { return string(b) }

// TestHeaderRand hält Z23 für den Rand jenseits der Menge (SPEC-284 Nr. 8 c): an Kommas geteilt, nur Leerzeichen und
// Tabulator am Rand eines Teils entfernt, Namen ohne Beachtung der ASCII-Schreibweise verglichen. Ein Teil von
// Connection mit jedem anderen Randzeichen (NBSP, EM SPACE, NEL, VT, FF, CR, LF) vorn oder hinten nennt den Header
// nicht, ein Name mit Zeichen jenseits von ASCII, das eine Unicode-Faltung auf einen Namen der Liste abbildet (Kelvin-
// Zeichen), ist weder Hop-by-Hop-Name noch Name aus Connection. Alle diese Header bleiben.
func TestHeaderRand(t *testing.T) {
	rand := map[string]string{
		"NBSP": bytesVon(0xC2, 0xA0), "EM SPACE": bytesVon(0xE2, 0x80, 0x83), "NEL": bytesVon(0xC2, 0x85),
		"VT": "\x0b", "FF": "\x0c", "CR": "\r", "LF": "\n",
	}
	kelvin := bytesVon(0xE2, 0x84, 0xAA)
	n := 0
	pruefe := func(was string, h map[string][]string, erwartet map[string][]string) {
		t.Helper()
		n++
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
		g.mockE6003(t)
		a := anfrage("GET")
		a.Header = h
		if _, err := g.v.Hybrid(context.Background(), a); err != nil {
			t.Errorf("%s: %v", was, err)
			return
		}
		if got := g.u.erster(t).a.Header; !reflect.DeepEqual(got, erwartet) {
			t.Errorf("%s: Header an Senden %q, erwartet %q", was, got, erwartet)
		}
	}
	for name, z := range rand {
		for _, seite := range []string{"vorn", "hinten"} {
			teil := "X-Edge" + z
			if seite == "vorn" {
				teil = z + "X-Edge"
			}
			pruefe(name+" "+seite, map[string][]string{"Connection": {teil}, "X-Edge": {"e"}}, map[string][]string{"X-Edge": {"e"}})
		}
	}
	// Dieselben Randzeichen entfernen den Namen nicht, wenn Leerzeichen und Tabulator dazukommen: der Rest ist kein Name.
	pruefe("Leerzeichen und Tabulator entfernen", map[string][]string{"Connection": {" \tX-Edge \t"}, "X-Edge": {"e"}}, map[string][]string{})
	// Kelvin-Zeichen: bytesVon(0xE2, 0x84, 0xAA) faltet strings.ToLower zu k.
	pruefe("Hop-by-Hop-Name mit Kelvin", map[string][]string{kelvin + "eep-Alive": {"1"}},
		map[string][]string{kelvin + "eep-Alive": {"1"}})
	pruefe("Name aus Connection mit Kelvin", map[string][]string{"Connection": {"X-key"}, "X-" + kelvin + "ey": {"1"}},
		map[string][]string{"X-" + kelvin + "ey": {"1"}})
	pruefe("Token von Connection mit Kelvin", map[string][]string{"Connection": {"X-" + kelvin + "ey"}, "X-key": {"1"}},
		map[string][]string{"X-key": {"1"}})
	pruefe("Präfix mit Kelvin", map[string][]string{"X-Recorder-" + kelvin: {"1"}}, map[string][]string{})
	if n != len(rand)*2+5 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}
