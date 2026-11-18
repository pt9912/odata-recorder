package hybrid_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
)

// kippe liefert s mit vertauschter Schreibweise des ASCII-Buchstabens an Stelle i.
func kippe(s string, i int) string {
	b := []byte(s)
	switch c := b[i]; {
	case 'a' <= c && c <= 'z':
		b[i] = c - ('a' - 'A')
	case 'A' <= c && c <= 'Z':
		b[i] = c + ('a' - 'A')
	}
	return string(b)
}

// TestDienstwahlSchreibweise hält Z19 für die Klasse „byteweise“ (SPEC-284 Nr. 4): zu jedem Root jede Form, die von ihm
// nur in der Schreibweise abweicht (je ein ASCII-Buchstabe vertauscht, alle groß), als Pfad ohne und mit Rest. Mit einem
// Dienst ergibt sie keinen Dienst (Ziel leer, U 0); mit beiden Diensten in der Liste wählt der Dienst mit dem
// byte-gleichen Root, in beiden Reihenfolgen. Beide Seiten der Grenze: der Pfad mit dem Root selbst wählt ihn.
func TestDienstwahlSchreibweise(t *testing.T) {
	n := 0
	for _, root := range []string{"/a", "/Ab", "/odata/Svc", "/a%2Fb"} {
		varianten := map[string]bool{strings.ToUpper(root): true}
		for i := range root {
			varianten[kippe(root, i)] = true
		}
		delete(varianten, root)
		for v := range varianten {
			n++
			for _, pfad := range []string{v, v + "/x"} {
				g := neuesRigMit(t, uPolicy(t, []string{"u", "v"}, nil), standardMax, []hybrid.Dienst{{Root: root, UpstreamURL: "http://u"}})
				g.mockE6003(t)
				a := anfrage("GET")
				a.Pfad = pfad
				if _, err := g.v.Hybrid(context.Background(), a); err == nil || g.u.zahl() != 0 {
					t.Errorf("Root %q, Pfad %q: Fehler %v, U %d; erwartet kein Dienst", root, pfad, err, g.u.zahl())
				}
			}
			paare := [][]hybrid.Dienst{
				{{Root: root, UpstreamURL: "http://u"}, {Root: v, UpstreamURL: "http://v"}},
				{{Root: v, UpstreamURL: "http://v"}, {Root: root, UpstreamURL: "http://u"}},
			}
			for _, liste := range paare {
				for _, c := range []struct{ pfad, ziel, odata string }{
					{v + "/x", "http://v", "/x"}, {v, "http://v", "/"}, {root + "/x", "http://u", "/x"}, {root, "http://u", "/"},
				} {
					g := neuesRigMit(t, uPolicy(t, []string{"u", "v"}, nil), standardMax, liste)
					g.mockE6003(t)
					a := anfrage("GET")
					a.Pfad = c.pfad
					if _, err := g.v.Hybrid(context.Background(), a); err != nil || g.u.zahl() != 1 {
						t.Errorf("Root %q, Variante %q, Pfad %q: Fehler %v, U %d", root, v, c.pfad, err, g.u.zahl())
						continue
					}
					if s := g.u.erster(t).a; s.Ziel != c.ziel || s.Pfad != c.odata {
						t.Errorf("Root %q, Variante %q, Pfad %q: Ziel %q Pfad %q, erwartet %q %q", root, v, c.pfad, s.Ziel, s.Pfad, c.ziel, c.odata)
					}
				}
			}
		}
	}
	// Roots: /a 1 Form, /Ab 2, /odata/Svc 9, /a%2Fb 4.
	if n != 1+2+9+4 {
		t.Errorf("Zählprüfung: %d Formen", n)
	}
}

// TestHeaderTrenner hält Z23 für die Klasse „an Kommas geteilt“ (SPEC-284 Nr. 8 c): jedes andere Byte von 0x00 bis 0x7F
// und vier Zeichen jenseits von ASCII (NBSP, EM SPACE, NEL, ideographischer Leerraum) innerhalb eines Tokens von
// Connection (vor und hinter dem Namen) trennt nicht: das Token nennt den Header X-Edge nicht, er bleibt. Das Komma
// trennt: dann entfällt er.
func TestHeaderTrenner(t *testing.T) {
	seps := make([]string, 0, 140)
	for b := 0; b <= 0x7F; b++ {
		if b != ',' {
			seps = append(seps, bytesVon(byte(b)))
		}
	}
	seps = append(seps, bytesVon(0xC2, 0xA0), bytesVon(0xE2, 0x80, 0x83), bytesVon(0xC2, 0x85), bytesVon(0xE3, 0x80, 0x80))
	n := 0
	pruefe := func(was, token string, soll map[string][]string) {
		t.Helper()
		n++
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
		g.mockE6003(t)
		a := anfrage("GET")
		a.Header = map[string][]string{"Connection": {token}, "X-Edge": {"e"}}
		if _, err := g.v.Hybrid(context.Background(), a); err != nil {
			t.Errorf("%s: %v", was, err)
			return
		}
		if got := g.u.erster(t).a.Header; !reflect.DeepEqual(got, soll) {
			t.Errorf("%s: Header an Senden %q, erwartet %q", was, got, soll)
		}
	}
	for _, sep := range seps {
		pruefe("Trenner vor dem Namen "+strings.ToValidUTF8(sep, "?"), "a"+sep+"X-Edge", map[string][]string{"X-Edge": {"e"}})
		pruefe("Trenner hinter dem Namen "+strings.ToValidUTF8(sep, "?"), "X-Edge"+sep+"a", map[string][]string{"X-Edge": {"e"}})
	}
	pruefe("Komma vor dem Namen", "a,X-Edge", map[string][]string{})
	pruefe("Komma hinter dem Namen", "X-Edge,a", map[string][]string{})
	if n != (127+4)*2+2 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestUpstreamMethodeUnveraendert hält Z24 für die Klasse „Methode der Anfrage“ (SPEC-284 Nr. 8 a): von allen zehn
// Methoden der Menge erreicht jede, die U erreicht, Senden byte-gleich. Grenze: U erreichen nur die sieben Methoden in
// Großbuchstaben (SPEC-261 Nr. 2); eine Mutation, die die Methode in Großbuchstaben wandelt, ändert für sie nichts
// (äquivalent), und für alle anderen läuft U nie.
func TestUpstreamMethodeUnveraendert(t *testing.T) {
	routen := []string{"POST /**", "PUT /**", "PATCH /**", "DELETE /**", "MERGE /**"}
	n, mitU := 0, 0
	for _, m := range alleMethoden() {
		n++
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, routen), standardMax, zielH)
		g.mockE6003(t)
		_, _ = g.v.Hybrid(context.Background(), anfrage(m))
		if g.u.zahl() == 0 {
			continue
		}
		mitU++
		if got := g.u.erster(t).a.Methode; got != m {
			t.Errorf("Methode %q kam als %q an", m, got)
		}
	}
	if n != 10 || mitU != 7 {
		t.Errorf("Zählprüfung: %d Methoden, %d mit U", n, mitU)
	}
}

// TestKontextMitUrsache hält Z10 und Teil 1 für die Klasse „ctx beendet“ (SPEC-284 Nr. 3 e): ein Kontext mit Ursache
// (WithCancelCause, WithTimeoutCause), dessen Ursache von ctx.Err() abweicht, an jeder Aufrufstelle, die ctx.Err()
// zurückgibt (vor R, vor U nach M, vor U nach R): das Ergebnis ist ctx.Err() selbst, nicht die Ursache, ohne Code; R und U
// laufen nicht.
func TestKontextMitUrsache(t *testing.T) {
	ursache := errors.New("ursache des Abbruchs")
	formen := []struct {
		name string
		neu  func() (context.Context, func())
		ist  error
	}{
		{"CancelCause", func() (context.Context, func()) {
			ctx, cancel := context.WithCancelCause(context.Background())
			return ctx, func() { cancel(ursache) }
		}, context.Canceled},
		{"TimeoutCause", func() (context.Context, func()) {
			ctx, cancel := context.WithTimeoutCause(context.Background(), 20*time.Millisecond, ursache)
			return ctx, func() { <-ctx.Done(); cancel() }
		}, context.DeadlineExceeded},
	}
	n := 0
	for _, f := range formen {
		for _, stelle := range []string{"vor R", "vor U nach M", "vor U nach R"} {
			n++
			ctx, beende := f.neu()
			defer beende()
			var g *rig
			switch stelle {
			case "vor R":
				g = neuesRigZiel(t, uPolicy(t, []string{"h"}, nil, katCoverage), standardMax, zielH)
				g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
					beende()
					return inbound.Antwort{}, code(codeNamed(t, mockCode))
				}
			case "vor U nach M":
				g = neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
				g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
					beende()
					return inbound.Antwort{}, code(codeNamed(t, mockCode))
				}
			default:
				g = neuesRigZiel(t, uPolicy(t, []string{"h"}, nil, katCoverage), standardMax, zielH)
				g.mockE6003(t)
				g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
					beende()
					return inbound.Antwort{}, code(codeNamed(t, missCode))
				}
			}
			_, err := g.v.Hybrid(ctx, anfrage("GET"))
			var fe *inbound.Fehler
			if err != ctx.Err() || !errors.Is(err, f.ist) || errors.Is(err, ursache) || errors.As(err, &fe) || g.u.zahl() != 0 {
				t.Errorf("%s %s: Fehler %v, U %d; erwartet ctx.Err() (%v) ohne Ursache und ohne Code", f.name, stelle, err, g.u.zahl(), f.ist)
			}
			if stelle == "vor R" && g.r.zahl() != 0 {
				t.Errorf("%s %s: R %d-mal gerufen", f.name, stelle, g.r.zahl())
			}
		}
	}
	if n != 6 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}
