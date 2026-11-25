package hybrid_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// lesendOrakel und mutierendOrakel sind die Klassen der Methoden nach SPEC-261 Nr. 2, unabhängig vom Code.
func lesendOrakel(m string) bool { return m == "GET" || m == "HEAD" }
func mutierendOrakel(m string) bool {
	return m == "POST" || m == "PUT" || m == "PATCH" || m == "DELETE" || m == "MERGE"
}

func alleMethoden() []string {
	return []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "get"}
}

// TestZ01DienstwahlBleibt hält Z01 (D): New hält keine Referenz auf Dienste; ändert der Aufrufer jedes Element der Liste
// und kürzt sie danach, wählt der Dienst nach der Liste, die er bekam (Ziel und OData-Pfad am Fake U).
func TestZ01DienstwahlBleibt(t *testing.T) {
	liste := []hybrid.Dienst{{Root: "/a", UpstreamURL: "http://a"}, {Root: "/a/b", UpstreamURL: "http://ab"}}
	g := &rig{m: &stufe{}, r: &stufe{}, u: &stufeU{}}
	v, err := hybrid.New(hybrid.Konfig{
		Dienste: liste, Mock: fakeMock{g.m}, Replay: fakeReplay{g.r}, Upstream: g.u,
		Policy: uPolicy(t, []string{"a", "ab", "zzz", "changed"}, nil), MaxBytes: standardMax, Telemetrie: supporttest.NewFakeTelemetry(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range liste {
		liste[i] = hybrid.Dienst{Root: "/zzz", UpstreamURL: "http://changed"}
	}
	_ = liste[:0]
	g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
		return inbound.Antwort{}, code(codeNamed(t, mockCode))
	}
	for _, c := range []struct{ pfad, ziel, odata string }{{"/a/x", "http://a", "/x"}, {"/a/b/y", "http://ab", "/y"}} {
		a := anfrage("GET")
		a.Pfad = c.pfad
		if _, err := v.Hybrid(context.Background(), a); err != nil {
			t.Fatalf("%s: %v", c.pfad, err)
		}
	}
	for i, c := range []struct{ ziel, odata string }{{"http://a", "/x"}, {"http://ab", "/y"}} {
		s := g.u.aufrufe[i].a
		if s.Ziel != c.ziel || s.Pfad != c.odata {
			t.Errorf("Aufruf %d: Ziel %q Pfad %q, erwartet %q %q", i, s.Ziel, s.Pfad, c.ziel, c.odata)
		}
	}
}

// TestUpstreamNachMock hält Z07: bei einem Fehler von M mit einem Code aus model.AlleCodes() ruft der Dienst U genau
// einmal und R nie, wenn der Code ODR-E6001, ODR-E6002 oder ODR-E6003 ist, der Host frei ist und die Methode lesend ist
// oder mutierend mit Route (ODR-E5001 als Fehler von M führt nur bei lesender Methode zu U); sonst kehrt der Fehler von
// M als derselbe Wert zurück.
func TestUpstreamNachMock(t *testing.T) {
	zaehler, upstreams := 0, 0
	for _, m := range alleMethoden() {
		for _, route := range []bool{false, true} {
			for _, c := range model.AlleCodes() {
				var routen []string
				if route && mutierendOrakel(m) {
					routen = []string{m + " " + standardP}
				}
				g := neuesRigZiel(t, uPolicy(t, []string{"h"}, routen), standardMax, zielH)
				fehler := code(c)
				g.mockFehler(fehler)
				_, err := g.v.Hybrid(context.Background(), anfrage(m))
				ausl := c.String() == "ODR-E6001" || c.String() == "ODR-E6002" || c.String() == "ODR-E6003"
				soll := ausl && (lesendOrakel(m) || mutierendOrakel(m) && route)
				// Entscheide entscheidet auch für ODR-E5001 als Fehler von M: lesend ja, mutierend nie (SPEC-261 Nr. 4).
				soll = soll || c.String() == missCode && lesendOrakel(m)
				zaehler++
				if soll {
					upstreams++
				}
				if got := g.u.zahl() == 1; got != soll || g.u.zahl() > 1 || g.r.zahl() != 0 {
					t.Errorf("%s route=%v %s: U %d (erwartet %v), R %d", m, route, c, g.u.zahl(), soll, g.r.zahl())
				}
				if !soll && err != error(fehler) {
					t.Errorf("%s route=%v %s: Fehler %v, erwartet derselbe Wert", m, route, c, err)
				}
			}
		}
	}
	if zaehler != 2*len(alleMethoden())*len(model.AlleCodes()) || upstreams != 3*(2+7)+2*2 {
		t.Errorf("Zählprüfung: %d Fälle, %d mit U", zaehler, upstreams)
	}
}

// TestUpstreamNachReplayMiss hält Z08 (U): liefert R ODR-E5001, gilt Entscheide mit diesem Code: GET und HEAD rufen U,
// jede mutierende Methode (auch mit Route) lässt ODR-E5001 unverändert zurückkehren, U 0; nie freigegebene Methoden
// erreichen R nicht.
func TestUpstreamNachReplayMiss(t *testing.T) {
	for _, m := range alleMethoden() {
		routen := []string(nil)
		if mutierendOrakel(m) {
			routen = []string{m + " " + standardP}
		}
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, routen, katCoverage), standardMax, zielH)
		g.mockE6003(t)
		miss := code(codeNamed(t, missCode))
		g.replayFehler(miss)
		_, err := g.v.Hybrid(context.Background(), anfrage(m))
		switch {
		case lesendOrakel(m):
			if err != nil || g.u.zahl() != 1 || g.r.zahl() != 1 {
				t.Errorf("%s: Fehler %v, U %d, R %d; erwartet U 1", m, err, g.u.zahl(), g.r.zahl())
			}
		case mutierendOrakel(m):
			if err != error(miss) || g.u.zahl() != 0 || g.r.zahl() != 1 {
				t.Errorf("%s: Fehler %v, U %d, R %d; erwartet E5001 und U 0", m, err, g.u.zahl(), g.r.zahl())
			}
		default:
			if err == nil || g.u.zahl() != 0 || g.r.zahl() != 0 {
				t.Errorf("%s: Fehler %v, U %d, R %d; erwartet kein Fallback", m, err, g.u.zahl(), g.r.zahl())
			}
		}
	}
}

// TestUpstreamNurNachE5001 hält Z09 (U): jeder Code von model.AlleCodes() als Fehler von R, unter der offensten Policy
// und unter einer, die nur coverage_insufficient führt (dort würde Entscheide für ODR-E6001 und ODR-E6002 den Upstream
// freigeben), mit GET kehrt als derselbe Wert zurück und lässt U ungerufen, außer ODR-E5001.
func TestUpstreamNurNachE5001(t *testing.T) {
	n, mitU := 0, 0
	policies := map[string]func() *policy.Policy{
		"offenste":     func() *policy.Policy { return offenste(t) },
		"nur coverage": func() *policy.Policy { return uPolicy(t, []string{"h"}, nil, katCoverage) },
	}
	for pname, pol := range policies {
		for _, c := range model.AlleCodes() {
			n++
			g := neuesRigZiel(t, pol(), standardMax, zielH)
			g.mockE6003(t)
			fehler := code(c)
			g.replayFehler(fehler)
			_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			if c.String() == missCode {
				mitU++
				if err != nil || g.u.zahl() != 1 {
					t.Errorf("%s %s: Fehler %v, U %d", pname, c, err, g.u.zahl())
				}
				continue
			}
			if err != error(fehler) || g.u.zahl() != 0 {
				t.Errorf("%s %s: Fehler %v, U %d; erwartet derselbe Wert ohne U", pname, c, err, g.u.zahl())
			}
		}
	}
	if n != 2*len(model.AlleCodes()) || len(model.AlleCodes()) < 40 || mitU != 2 {
		t.Errorf("Zählprüfung: %d Codes, %d mit U", n, mitU)
	}
}

// TestUpstreamAbbruch hält Z10 (U): ist ctx nach R (ODR-E5001) oder nach M beendet, läuft U nicht und das Ergebnis ist
// ctx.Err() ohne Code.
func TestUpstreamAbbruch(t *testing.T) {
	t.Run("nach R", func(t *testing.T) {
		ctx, cancel := abbruchKontext()
		g := neuesRigZiel(t, offenste(t), standardMax, zielH)
		g.mockE6003(t)
		g.r.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			cancel()
			return inbound.Antwort{}, code(codeNamed(t, missCode))
		}
		_, err := g.v.Hybrid(ctx, anfrage("GET"))
		var f *inbound.Fehler
		if err != ctx.Err() || errors.As(err, &f) || g.u.zahl() != 0 {
			t.Errorf("Fehler %v, U %d; erwartet ctx.Err() ohne Code", err, g.u.zahl())
		}
	})
	t.Run("nach M", func(t *testing.T) {
		ctx, cancel := abbruchKontext()
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
		g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			cancel()
			return inbound.Antwort{}, code(codeNamed(t, mockCode))
		}
		_, err := g.v.Hybrid(ctx, anfrage("GET"))
		var f *inbound.Fehler
		if err != ctx.Err() || errors.As(err, &f) || g.u.zahl() != 0 {
			t.Errorf("Fehler %v, U %d; erwartet ctx.Err() ohne Code", err, g.u.zahl())
		}
	})
}

type ctxSchluessel struct{}

// wege liefert die beiden Wege zu U: nach M und nach R (Replay-Miss); die Policy gibt den Upstream frei.
func wege(t *testing.T) map[string]func() *rig {
	return map[string]func() *rig{
		"nach M": func() *rig {
			g := neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
			g.mockE6003(t)
			return g
		},
		"nach R": func() *rig {
			g := neuesRigZiel(t, offenste(t), standardMax, zielH)
			g.mockE6003(t)
			g.replayFehler(code(codeNamed(t, missCode)))
			return g
		},
	}
}

// TestUpstreamKontext hält Z11 (U): H6 erhält ctx unverändert (dieselbe Identität), und ein Kontextfehler von Senden
// (roh, verpackt, in einem Join) kehrt als derselbe Wert ohne Code zurück, nicht als ODR-E4001 oder ODR-E9001.
func TestUpstreamKontext(t *testing.T) {
	fehler := map[string]error{
		"Canceled":         context.Canceled,
		"DeadlineExceeded": context.DeadlineExceeded,
		"verpackt":         fmt.Errorf("x: %w", context.Canceled),
		"Join":             errors.Join(outbound.ErrUpstreamZeit, context.DeadlineExceeded),
	}
	n := 0
	for weg, neu := range wege(t) {
		ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ctxSchluessel{}, "v"))
		g := neu()
		if _, err := g.v.Hybrid(ctx, anfrage("GET")); err != nil || g.u.zahl() != 1 {
			t.Fatalf("%s: Fehler %v, U %d", weg, err, g.u.zahl())
		}
		if any(g.u.erster(t).ctx) != any(ctx) {
			t.Errorf("%s: ctx an Senden ist nicht der des Aufrufs", weg)
		}
		cancel()
		for name, e := range fehler {
			n++
			g := neu()
			g.u.liefert(outbound.UpstreamAntwort{}, e)
			_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			var f *inbound.Fehler
			if err != e || errors.As(err, &f) || g.u.zahl() != 1 {
				t.Errorf("%s %s: Fehler %v, erwartet derselbe Wert ohne Code", weg, name, err)
			}
		}
	}
	if n != 8 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestUpstreamProbe hält Z12 (U): mit nichtleerem Ziel macht eine Policy mit nur einer Host-Freigabe und Route den POST
// mit Body fallback-fähig (gepuffert, M und U je ein eigener Leser über den vollständigen Puffer); mit leerem Ziel und
// ohne Route ist die Anfrage nicht fähig (Body ungelesen und unverändert an M).
func TestUpstreamProbe(t *testing.T) {
	inhalt := muster(5, 9)
	t.Run("fähig", func(t *testing.T) {
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, []string{"POST " + standardP}), standardMax, zielH)
		g.m.lies = liesAlles
		g.mockE6003(t)
		s := neuerStrom(inhalt)
		_, err := g.v.Hybrid(context.Background(), mitBody("POST", s, -1))
		m := g.m.erster(t)
		u := g.u.erster(t)
		if err != nil || m.a.Laenge != 5 || !reflect.DeepEqual(m.gelesen, inhalt) || m.a.Body == io.Reader(s) {
			t.Errorf("M: Fehler %v, Laenge %d, gelesen %v", err, m.a.Laenge, m.gelesen)
		}
		if u.a.Laenge != 5 || !reflect.DeepEqual(u.gelesen, inhalt) || u.a.Body == m.a.Body {
			t.Errorf("U: Laenge %d, gelesen %v, eigener Leser %v", u.a.Laenge, u.gelesen, u.a.Body != m.a.Body)
		}
	})
	t.Run("leeres Ziel", func(t *testing.T) {
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, []string{"POST " + standardP}), standardMax, "")
		g.mockE6003(t)
		s := neuerStrom(inhalt)
		_, _ = g.v.Hybrid(context.Background(), mitBody("POST", s, -1))
		m := g.m.erster(t)
		if m.a.Body != io.Reader(s) || m.a.Laenge != -1 || s.reads.Load() != 0 || g.u.zahl() != 0 {
			t.Errorf("Body gelesen oder ersetzt: Reads %d, Laenge %d, U %d", s.reads.Load(), m.a.Laenge, g.u.zahl())
		}
	})
	t.Run("nur Host, keine Route", func(t *testing.T) {
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
		g.mockE6003(t)
		s := neuerStrom(inhalt)
		_, _ = g.v.Hybrid(context.Background(), mitBody("POST", s, -1))
		if m := g.m.erster(t); m.a.Body != io.Reader(s) || s.reads.Load() != 0 || g.u.zahl() != 0 {
			t.Errorf("POST ohne Route ist nicht fähig: Reads %d, U %d", s.reads.Load(), g.u.zahl())
		}
	})
}

// TestDienstwahl hält Z19 (Wahl): der längste Root an einer Segmentgrenze, in jeder Reihenfolge der Liste; Ziel ist die
// UpstreamURL des gewählten Dienstes und der OData-Pfad der Rest nach dem Root; ohne Dienst läuft U nie.
func TestDienstwahl(t *testing.T) {
	dienste := []hybrid.Dienst{
		{Root: "/a", UpstreamURL: "http://a"}, {Root: "/a/b", UpstreamURL: "http://ab"},
		{Root: "/", UpstreamURL: "http://root"}, {Root: "/ab", UpstreamURL: "http://abx"},
	}
	faelle := []struct{ pfad, ziel, odata string }{
		{"/a", "http://a", "/"}, {"/a/x", "http://a", "/x"}, {"/a/b", "http://ab", "/"}, {"/a/b/c", "http://ab", "/c"},
		{"/ab", "http://abx", "/"}, {"/ab/c", "http://abx", "/c"}, {"/abc", "http://root", "/abc"},
		{"/", "http://root", "/"}, {"/a/", "http://a", "/"}, {"/a/bc", "http://a", "/bc"},
	}
	hosts := []string{"a", "ab", "abx", "root"}
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
			g := neuesRigMit(t, uPolicy(t, hosts, nil), standardMax, liste)
			g.mockE6003(t)
			a := anfrage("GET")
			a.Pfad = f.pfad
			if _, err := g.v.Hybrid(context.Background(), a); err != nil || g.u.zahl() != 1 {
				t.Errorf("%s: Fehler %v, U %d", f.pfad, err, g.u.zahl())
				continue
			}
			if s := g.u.erster(t).a; s.Ziel != f.ziel || s.Pfad != f.odata {
				t.Errorf("%s (umgekehrt %v): Ziel %q Pfad %q, erwartet %q %q", f.pfad, umgekehrt, s.Ziel, s.Pfad, f.ziel, f.odata)
			}
		}
	}
	if n != 20 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
	g := neuesRigZiel(t, uPolicy(t, hosts, nil), standardMax, "http://a")
	g.mockE6003(t)
	for _, pfad := range []string{"/other", "/odata", "/odata/sv", "odata/svc"} {
		a := anfrage("GET")
		a.Pfad = pfad
		if _, err := g.v.Hybrid(context.Background(), a); err == nil || g.u.zahl() != 0 {
			t.Errorf("Pfad %q ohne Dienst: Fehler %v, U %d", pfad, err, g.u.zahl())
		}
	}
}

// ziel-Formen: die Zeichenkette der Konfig, die Einträge der Policy, die den Host freigeben, und ob die Policy sie liest.
type zielForm struct {
	ziel   string
	hosts  []string
	lesbar bool
}

func zielFormen() []zielForm {
	return []zielForm{
		{"http://h", []string{"h"}, true},
		{"HTTP://H:8080", []string{"h:8080"}, true},
		{"https://h:0443", []string{"h"}, true},
		{"http://h/", []string{"h"}, true},
		{"http://h/p", []string{"h"}, true},
		{"http://[::1]:80", []string{"[::1]"}, true},
		{"http://h ", []string{"h"}, false},
	}
}

// TestZielAnStellen hält Z19 (Ziel): dieselbe Zeichenkette, byte-gleich, geht an Entscheide und an Senden; je Form an
// den Stellen H1 (Probe, vier Aufrufe zugleich), H3 und H5. Entscheide wird über die echte Policy beobachtet: ein
// verlorenes Ziel an der Stelle ändert die Entscheidung.
func TestZielAnStellen(t *testing.T) {
	n := 0
	for _, f := range zielFormen() {
		erwartetU := 0
		if f.lesbar {
			erwartetU = 1
		}
		pruefe := func(stelle string, g *rig, err error) {
			t.Helper()
			n++
			if g.u.zahl() != erwartetU || (f.lesbar && err != nil) {
				t.Errorf("%s, Ziel %q: U %d, Fehler %v; erwartet U %d", stelle, f.ziel, g.u.zahl(), err, erwartetU)
				return
			}
			if f.lesbar {
				if s := g.u.erster(t).a; s.Ziel != f.ziel || s.Pfad != "/Products" {
					t.Errorf("%s: Ziel %q Pfad %q an Senden", stelle, s.Ziel, s.Pfad)
				}
			}
		}
		// H1 und H3: POST mit Route und Body; fähig nur mit gelesenem Ziel, und der volle Pfad erreicht Route und U.
		g := neuesRigZiel(t, uPolicy(t, f.hosts, []string{"POST " + standardP}), standardMax, f.ziel)
		g.mockE6003(t)
		_, err := g.v.Hybrid(context.Background(), mitBody("POST", neuerStrom(muster(4, 1)), -1))
		wantLaenge := int64(-1)
		if f.lesbar {
			wantLaenge = 4
		}
		if m := g.m.erster(t); m.a.Laenge != wantLaenge {
			t.Errorf("H1, Ziel %q: Laenge an M %d, erwartet %d", f.ziel, m.a.Laenge, wantLaenge)
		}
		pruefe("H1/H3", g, err)
		// H3 allein: GET ohne Body.
		g = neuesRigZiel(t, uPolicy(t, f.hosts, nil), standardMax, f.ziel)
		g.mockE6003(t)
		_, err = g.v.Hybrid(context.Background(), anfrage("GET"))
		pruefe("H3", g, err)
		// H5: Replay-Miss, dann Upstream.
		g = neuesRigZiel(t, uPolicy(t, f.hosts, nil, katCoverage), standardMax, f.ziel)
		g.mockE6003(t)
		g.replayFehler(code(codeNamed(t, missCode)))
		_, err = g.v.Hybrid(context.Background(), anfrage("GET"))
		pruefe("H5", g, err)
	}
	if n != 3*len(zielFormen()) {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// snapshot kopiert eine Header-Map tief.
func snapshot(h map[string][]string) map[string][]string { return tiefeKopie(h) }

func schreibweisen(n string) []string {
	wechsel := []byte(n)
	for i := range wechsel {
		if i%2 == 1 && wechsel[i] >= 'a' && wechsel[i] <= 'z' {
			wechsel[i] -= 'a' - 'A'
		} else if i%2 == 0 && wechsel[i] >= 'A' && wechsel[i] <= 'Z' {
			wechsel[i] += 'a' - 'A'
		}
	}
	var aus []string
	for _, s := range []string{n, strings.ToLower(n), strings.ToUpper(n), string(wechsel)} {
		dup := false
		for _, a := range aus {
			dup = dup || a == s
		}
		if !dup {
			aus = append(aus, s)
		}
	}
	return aus
}

// TestHeaderFilter hält Z23: jeder Name der Menge (Präfix X-Recorder-, die neun Hop-by-Hop-Namen, Expect, Host,
// Content-Length, die drei Header zum Überschreiben der Methode) entfällt in jeder Schreibweise, ebenso jeder Name, den
// ein Wert von Connection nennt; Namen auf der anderen Seite der Grenze bleiben mit allen Werten; die Kopie ist neu, die
// Map der Anfrage bleibt unverändert.
func TestHeaderFilter(t *testing.T) {
	entfaellt := []string{
		"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "TE", "Trailer",
		"Transfer-Encoding", "Upgrade", "Expect", "Host", "Content-Length", "X-HTTP-Method", "X-HTTP-Method-Override",
		"X-Method-Override", "X-Recorder-Workspace", "X-Recorder-", "X-Recorder-Recording-Session",
		"X-Recorder-Replay-Session", "X-Recorder-Fallback", "X-Recorder-Zusatz",
	}
	bleibt := map[string][]string{
		"X-Recorderx": {"1"}, "X-Recorder": {"2"}, "X-Recorde-": {"3"}, "Authorization": {"a", "b"}, "Accept": {"x"},
		"If-Match": {"\"e\""}, "Content-Type": {"t"}, "Accept-Encoding": {"br"}, "X-Empty": {""}, "X-Multi": {"b", "a", "c"},
		"Connectionx": {"4"}, "Xconnection": {"5"}, "Content-Lengthx": {"6"}, "Qux": {"q"}, "Hostx": {"7"},
		"X-HTTP-Methodx": {"8"}, "X-Foox": {"9"},
	}
	h := map[string][]string{}
	for k, v := range bleibt {
		h[k] = append([]string{}, v...)
	}
	n := 0
	for _, name := range entfaellt {
		for _, s := range schreibweisen(name) {
			h[s] = []string{"w1", "w2"}
			n++
		}
	}
	// Namen aus Connection: Leerraum, leere Teile, mehrere Werte, ungewöhnliche Schreibweise des Namens Connection.
	h["cOnNeCtIoN"] = []string{" X-Foo , ,x-bar\t", "BAZ", ""}
	h["x-foo"], h["X-BAR"], h["Baz"] = []string{"f"}, []string{"b"}, []string{"z"}
	vorher := snapshot(h)
	g := neuesRigZiel(t, uPolicy(t, []string{"h"}, nil), standardMax, zielH)
	g.mockE6003(t)
	a := anfrage("GET")
	a.Header = h
	if _, err := g.v.Hybrid(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	got := g.u.erster(t).a.Header
	if !reflect.DeepEqual(got, bleibt) {
		t.Errorf("Header an Senden %v, erwartet %v", got, bleibt)
	}
	if !reflect.DeepEqual(h, vorher) {
		t.Errorf("die Map der Anfrage änderte sich")
	}
	if n < 60 {
		t.Errorf("Zählprüfung: %d Namen in Schreibweisen", n)
	}
	// Die Kopie teilt keine Wertliste mit der Anfrage.
	got["Authorization"][0] = "geändert"
	got["neu"] = []string{"x"}
	if !reflect.DeepEqual(h, vorher) {
		t.Errorf("die Kopie teilt Speicher mit der Anfrage")
	}
}

// TestUpstreamArgumente hält Z24: Methode, OData-Pfad (Rest nach dem Root), RawQuery unverändert, Body und Laenge aus
// dem Puffer; ohne Body nil und 0.
func TestUpstreamArgumente(t *testing.T) {
	roots := []struct {
		root  string
		pfade map[string]string
	}{
		{"/odata/svc", map[string]string{
			"/odata/svc": "/", "/odata/svc/": "/", "/odata/svc/Products(1)": "/Products(1)", "/odata/svc/a/b": "/a/b",
		}},
		{"/", map[string]string{"/": "/", "/x/y": "/x/y", "/odata/svc": "/odata/svc"}},
	}
	queries := []string{"", "%2F", "%2f", "+", "$top=1&a=b%20c", "$filter=a%2Feq+b"}
	n := 0
	for _, r := range roots {
		for pfad, want := range r.pfade {
			for _, q := range queries {
				n++
				g := neuesRigMit(t, uPolicy(t, []string{"h"}, nil), standardMax, []hybrid.Dienst{{Root: r.root, UpstreamURL: zielH}})
				g.mockE6003(t)
				a := anfrage("GET")
				a.Pfad, a.RawQuery = pfad, q
				if _, err := g.v.Hybrid(context.Background(), a); err != nil {
					t.Errorf("%s: %v", pfad, err)
					continue
				}
				s := g.u.erster(t).a
				if s.Methode != "GET" || s.Pfad != want || s.RawQuery != q || s.Body != nil || s.Laenge != 0 {
					t.Errorf("Root %s Pfad %s Query %q: %+v", r.root, pfad, q, s)
				}
			}
		}
	}
	if n != 4*6+3*6 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
	inhalt := muster(4, 5)
	faelle := []struct {
		name   string
		a      inbound.Anfrage
		body   bool
		laenge int64
	}{
		{"Body nil, Laenge 0", func() inbound.Anfrage { a := anfrage("POST"); return a }(), false, 0},
		{"Body nil, Laenge -1", func() inbound.Anfrage { a := anfrage("POST"); a.Laenge = -1; return a }(), false, 0},
		{"Leser, Laenge 0", mitBody("POST", neuerStrom(inhalt), 0), false, 0},
		{"Leser, Laenge -1", mitBody("POST", neuerStrom(inhalt), -1), true, 4},
		{"Leser, Laenge 4", mitBody("POST", neuerStrom(inhalt), 4), true, 4},
		{"Leser, Laenge 2 bei 4 Bytes", mitBody("POST", neuerStrom(inhalt), 2), true, 4},
	}
	for _, f := range faelle {
		g := neuesRigZiel(t, uPolicy(t, []string{"h"}, []string{"POST /**"}), standardMax, zielH)
		g.mockE6003(t)
		if _, err := g.v.Hybrid(context.Background(), f.a); err != nil {
			t.Errorf("%s: %v", f.name, err)
			continue
		}
		s := g.u.erster(t)
		if (s.a.Body != nil) != f.body || s.a.Laenge != f.laenge || (f.body && !reflect.DeepEqual(s.gelesen, inhalt)) {
			t.Errorf("%s: Body %v, Laenge %d, gelesen %v", f.name, s.a.Body != nil, s.a.Laenge, s.gelesen)
		}
	}
}

// fehlerTabelle ist das Orakel der Abbildung von Senden (SPEC-284 Nr. 11).
type fehlerFall struct {
	name   string
	err    error
	code   model.Code
	status int
}

func fehlerTabelle() []fehlerFall {
	return []fehlerFall{
		{"Zeit", outbound.ErrUpstreamZeit, model.CodeE4001, 504},
		{"Zeit verpackt", fmt.Errorf("x: %w", outbound.ErrUpstreamZeit), model.CodeE4001, 504},
		{"NichtErreichbar", outbound.ErrUpstreamNichtErreichbar, model.CodeE4001, 502},
		{"Join Zeit, NichtErreichbar", errors.Join(outbound.ErrUpstreamZeit, outbound.ErrUpstreamNichtErreichbar), model.CodeE4001, 504},
		{"Join NichtErreichbar, Zeit", errors.Join(outbound.ErrUpstreamNichtErreichbar, outbound.ErrUpstreamZeit), model.CodeE4001, 504},
		{"ErrUngueltig", model.ErrUngueltig, model.CodeE9001, 500},
		{"sonst", errors.New("boom"), model.CodeE9001, 500},
	}
}

// TestUpstreamFehler hält Z25: jeder Fehler von Senden ergibt den Code und den Status der Tabelle, die Ursache bleibt
// über errors.Is erreichbar, der Text nennt nur Code und message; Senden einmal, kein weiterer Aufruf von M und R, und
// der Fehler ersetzt den früheren.
func TestUpstreamFehler(t *testing.T) {
	n := 0
	for weg, neu := range wege(t) {
		for _, f := range fehlerTabelle() {
			n++
			g := neu()
			g.u.liefert(outbound.UpstreamAntwort{}, f.err)
			_, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			var fe *inbound.Fehler
			if !errors.As(err, &fe) || fe.Code != f.code || fe.Status != f.status || !errors.Is(err, f.err) {
				t.Errorf("%s %s: Fehler %v", weg, f.name, err)
				continue
			}
			if want := f.code.String() + ": " + f.code.Message(); err.Error() != want || strings.Contains(err.Error(), "boom") {
				t.Errorf("%s %s: Text %q", weg, f.name, err.Error())
			}
			if strings.HasPrefix(f.name, "Join") && (!errors.Is(err, outbound.ErrUpstreamZeit) || !errors.Is(err, outbound.ErrUpstreamNichtErreichbar)) {
				t.Errorf("%s %s: die Ursache trägt nicht beide Fehler des Ports", weg, f.name)
			}
			wantR := 0
			if weg == "nach R" {
				wantR = 1
			}
			if g.u.zahl() != 1 || g.m.zahl() != 1 || g.r.zahl() != wantR {
				t.Errorf("%s %s: U %d, M %d, R %d", weg, f.name, g.u.zahl(), g.m.zahl(), g.r.zahl())
			}
		}
	}
	if n != 14 {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestUpstreamKennzeichnung hält Z26 (U): die Antwort aus U trägt X-Recorder-Fallback mit genau einem Wert upstream;
// ein Header gleichen Namens der Stufe (jede Schreibweise, auch mehrere zugleich) entfällt; die Map der Stufe bleibt.
func TestUpstreamKennzeichnung(t *testing.T) {
	vorbelegungen := [][]string{nil}
	for _, s := range schreibweisen("X-Recorder-Fallback") {
		vorbelegungen = append(vorbelegungen, []string{s})
	}
	vorbelegungen = append(vorbelegungen, schreibweisen("X-Recorder-Fallback"))
	n := 0
	for weg, neu := range wege(t) {
		for _, namen := range vorbelegungen {
			n++
			g := neu()
			h := map[string][]string{"X-Keep": {"k"}}
			for _, nm := range namen {
				h[nm] = []string{"a", "b"}
			}
			vorher := snapshot(h)
			g.u.liefert(outbound.UpstreamAntwort{Status: 200, Header: h, Body: neuerStrom(nil)}, nil)
			got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			if err != nil || !reflect.DeepEqual(kennzeichen(got), []string{"upstream"}) {
				t.Errorf("%s %v: Fehler %v, Kennzeichnung %v", weg, namen, err, kennzeichen(got))
			}
			if !reflect.DeepEqual(h, vorher) || zeiger(got.Header) == zeiger(h) {
				t.Errorf("%s %v: die Map der Stufe änderte sich oder wurde weitergereicht", weg, namen)
			}
		}
	}
	if n != 2*len(vorbelegungen) {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}

// TestUpstreamAntwort hält Z27: Status und Strom gehen unverändert aus (nie gelesen, nie geschlossen); die Header ohne
// die Hop-by-Hop-Namen und die Namen aus Connection der Antwort; alle übrigen mit allen Werten.
func TestUpstreamAntwort(t *testing.T) {
	hop := []string{
		"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "TE", "Trailer",
		"Transfer-Encoding", "Upgrade",
	}
	bleibt := map[string][]string{
		"Set-Cookie": {"a=1", "b=2"}, "Location": {"http://upstream/x"}, "Content-Length": {"7"}, "Content-Encoding": {"gzip"},
		"Content-Type": {"application/json"}, "X-Eigen": {"1", ""}, "Content-Location": {"http://upstream/y"},
	}
	n := 0
	for _, status := range []int{200, 204, 301, 404, 500} {
		h := map[string][]string{}
		for k, v := range bleibt {
			h[k] = append([]string{}, v...)
		}
		for _, name := range hop {
			for _, s := range schreibweisen(name) {
				h[s] = []string{"x"}
				n++
			}
		}
		h["cOnNeCtIoN"] = []string{"X-Gelistet , ,close"}
		h["x-gelistet"] = []string{"g"}
		s := neuerStrom([]byte("http://upstream/@odata.id"))
		g := wege(t)["nach M"]()
		g.u.liefert(outbound.UpstreamAntwort{Status: status, Header: h, Body: s}, nil)
		got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
		if err != nil || got.Status != status || got.Body != io.ReadCloser(s) || s.reads.Load() != 0 || s.closes.Load() != 0 {
			t.Errorf("Status %d: Fehler %v, Status %d, Reads %d, Closes %d", status, err, got.Status, s.reads.Load(), s.closes.Load())
		}
		want := map[string][]string{headerFallbackName: {"upstream"}}
		for k, v := range bleibt {
			want[k] = v
		}
		if !reflect.DeepEqual(got.Header, want) {
			t.Errorf("Status %d: Header %v, erwartet %v", status, got.Header, want)
		}
	}
	if n < 5*9 {
		t.Errorf("Zählprüfung: %d Hop-by-Hop-Einträge", n)
	}
}

const headerFallbackName = "X-Recorder-Fallback"

// TestUpstreamBodySchliessen hält Z28 (U): liefert Senden Fehler und Antwort mit Body zugleich, schließt der Dienst den
// Body genau einmal und liefert den abgebildeten Fehler (ein Kontextfehler ohne Code); die ausgelieferte Antwort schließt
// er nie.
func TestUpstreamBodySchliessen(t *testing.T) {
	fehler := []error{context.Canceled, fmt.Errorf("x: %w", context.DeadlineExceeded)}
	for _, f := range fehlerTabelle() {
		fehler = append(fehler, f.err)
	}
	n := 0
	for weg, neu := range wege(t) {
		for _, e := range fehler {
			n++
			s := neuerStrom([]byte("x"))
			g := neu()
			g.u.liefert(outbound.UpstreamAntwort{Status: 200, Body: s}, e)
			got, err := g.v.Hybrid(context.Background(), anfrage("GET"))
			if s.closes.Load() != 1 || s.reads.Load() != 0 || err == nil || got.Body != nil {
				t.Errorf("%s %v: Closes %d, Reads %d, Fehler %v", weg, e, s.closes.Load(), s.reads.Load(), err)
			}
			if kontext := errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded); kontext && err != e {
				t.Errorf("%s %v: Kontextfehler verändert: %v", weg, e, err)
			}
		}
		s := neuerStrom(nil)
		g := neu()
		g.u.liefert(outbound.UpstreamAntwort{Status: 200, Body: s}, nil)
		if _, err := g.v.Hybrid(context.Background(), anfrage("GET")); err != nil || s.closes.Load() != 0 {
			t.Errorf("%s: Erfolg: Fehler %v, Closes %d", weg, err, s.closes.Load())
		}
	}
	if n != 2*(2+len(fehlerTabelle())) {
		t.Errorf("Zählprüfung: %d Fälle", n)
	}
}
