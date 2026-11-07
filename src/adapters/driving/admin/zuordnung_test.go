package admin_test

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// z4Muster sind die drei Muster von M-Zuordnung (Z4).
func z4Muster() []string { return []string{"/a", "/a/{id}", "/{p}/c/{q}"} }

// z4Werte sind die 19 Werte eines Parameters, roh im Pfad und dekodiert als Literale des Vertrags (SPEC-238 Nr. 2): sieben
// Formen aus Punkten, Prozentfolgen und Bytes sowie zwölf Formen, in denen die Dekodierung eines Pfadsegments von der
// eines Query-Wertes abweicht (`+` roh und als `%2B`/`%2b`, `%20`, `;`, `%3B`, `%25`, `&=`, `%3F`, `%23`, `~`); die
// Werte mit Bytes jenseits von ASCII entstehen aus Bytewerten.
func z4Werte() (roh, dekodiert []string) {
	roh = []string{
		"x", ".", "..", "a%2Fb", "%C3%A4", "%FF", "%00",
		"+", "a+b", "%2B", "%2b", "%20", ";", "%3B", "%25", "&=", "%3F", "%23", "~",
	}
	dekodiert = []string{
		"x", ".", "..", "a/b", string([]byte{0xC3, 0xA4}), string([]byte{0xFF}), string([]byte{0x00}),
		"+", "a+b", "+", "+", " ", ";", ";", "%", "&=", "?", "#", "~",
	}
	return roh, dekodiert
}

// z4Fall ist ein Pfad zu einem Muster mit der Erwartung des Orakels: Treffer samt Pfadwerten oder 404.
type z4Fall struct {
	name    string
	muster  string
	pfad    string
	treffer bool
	werte   map[string]string
}

// segmenteVon zerlegt ein Muster in Segmente (das Orakel des Tests, unabhängig vom Adapter).
func segmenteVon(muster string) []string { return strings.Split(muster[1:], "/") }

// pfadAus setzt einen Pfad aus Segmenten zusammen.
func pfadAus(segs []string) string { return "/" + strings.Join(segs, "/") }

// z4Generator erzeugt die Pfade zu allen drei Mustern: Literal gleich, Schreibweise abweichend, als Prozentfolge in
// Klein- und Großbuchstaben, ein Segment mehr und eines weniger, dazu je Stelle eines Parameters sieben Werte.
func z4Generator() []z4Fall {
	roh, dek := z4Werte()
	var faelle []z4Fall
	for _, muster := range z4Muster() {
		segs := segmenteVon(muster)
		variante := func(name string, treffer bool, aendere func(i int, s string) string) {
			neu := make([]string, len(segs))
			werte := map[string]string{}
			for i, s := range segs {
				neu[i] = aendere(i, s)
				if strings.HasPrefix(s, "{") {
					werte[s[1:len(s)-1]] = "x"
				}
			}
			if !treffer {
				werte = nil
			}
			faelle = append(faelle, z4Fall{name, muster, pfadAus(neu), treffer, werte})
		}
		param := func(s string) bool { return strings.HasPrefix(s, "{") }
		variante("gleich", true, func(_ int, s string) string {
			if param(s) {
				return "x"
			}
			return s
		})
		variante("Schreibweise", false, func(_ int, s string) string {
			if param(s) {
				return "x"
			}
			return strings.ToUpper(s)
		})
		variante("Prozentfolge klein", true, func(_ int, s string) string {
			if param(s) {
				return "x"
			}
			return "%" + strings.ToLower(fmt.Sprintf("%X", s[0])) + s[1:]
		})
		variante("Prozentfolge groß", false, func(_ int, s string) string {
			if param(s) {
				return "x"
			}
			return "%" + fmt.Sprintf("%X", strings.ToUpper(s)[0]) + s[1:]
		})
		faelle = append(faelle,
			z4Fall{"ein Segment mehr", muster, pfadAus(append(slices.Clone(segs), "x")), false, nil},
			z4Fall{"ein Segment weniger", muster, pfadAus(segs[:len(segs)-1]), false, nil})
		// die Pfade mit Segmenten, die aus den Mustern selbst stammen, tragen den Parameternamen als Literal nie
		for i, s := range segs {
			if !param(s) {
				continue
			}
			for k := range roh {
				neu := make([]string, len(segs))
				werte := map[string]string{}
				for j, t := range segs {
					switch {
					case !param(t):
						neu[j] = t
					case j == i:
						neu[j], werte[t[1:len(t)-1]] = roh[k], dek[k]
					default:
						neu[j], werte[t[1:len(t)-1]] = "x", "x"
					}
				}
				faelle = append(faelle, z4Fall{fmt.Sprintf("Parameter %d = %q", i, roh[k]), muster, pfadAus(neu), true, werte})
			}
			// ein Parameter nimmt nur ein nichtleeres Segment
			leer := make([]string, len(segs))
			for j, t := range segs {
				leer[j] = t
				if param(t) {
					leer[j] = "x"
				}
			}
			leer[i] = ""
			faelle = append(faelle, z4Fall{fmt.Sprintf("Parameter %d leer", i), muster, pfadAus(leer), false, nil})
		}
	}
	return faelle
}

// TestZuordnung hält Z4 (M-Zuordnung): ein Literal passt bytegleich auf das dekodierte Segment, ein Parameter auf jedes
// nichtleere Segment, die Segmentzahl muss gleich sein; bei einem Treffer läuft die Route einmal mit den Pfadwerten,
// sonst läuft sie nicht und die Antwort ist 404 mit ODR-E2102.
func TestZuordnung(t *testing.T) {
	faelle := z4Generator()
	roh, _ := z4Werte()
	// 6 Formen je Muster, dazu Slots × (Werte und das leere Segment): /a hat 0, /a/{id} 1, /{p}/c/{q} 2 Parameter
	if soll := 6*len(z4Muster()) + (0+1+2)*(len(roh)+1); len(faelle) != soll || soll != 78 || len(roh) != 19 {
		t.Fatalf("%d Fälle, erwartet 6 × 3 + 3 × (19 + 1) = %d", len(faelle), soll)
	}
	for _, f := range faelle {
		route := neueRoute()
		u := neuMit(t, admin.Route{Methode: "GET", Muster: f.muster, Handler: route.handle})
		rec := u.rufe(anfrageMit("GET", f.pfad, nil, nil))
		name := f.muster + " " + f.name + " " + f.pfad
		if !f.treffer {
			pruefeFehlerAntwort(t, name, rec, http.StatusNotFound, objekt404, "")
			if route.aufrufe() != 0 {
				t.Errorf("%s: die Route lief", name)
			}
			continue
		}
		a, ok := route.einzige()
		if !ok || rec.Code != 200 || !maps.Equal(a.Pfadwerte, f.werte) {
			t.Errorf("%s: Aufrufe %d, Status %d, Pfadwerte %q, erwartet %q", name, route.aufrufe(), rec.Code, a.Pfadwerte, f.werte)
		}
	}
}

// TestZuordnungGrenze hält M-Grenze (Z4): 64 Routen mit vier Parametern; /r<i>/1/2/3/4 trifft genau die Route i.
func TestZuordnungGrenze(t *testing.T) {
	var routen []admin.Route
	var fakes []*routeFake
	for i := 0; i < 64; i++ {
		f := neueRoute()
		fakes = append(fakes, f)
		routen = append(routen, admin.Route{Methode: "GET", Muster: "/r" + strconv.Itoa(i) + "/{a}/{b}/{c}/{d}", Handler: f.handle})
	}
	if len(routen) != 64 || len(fakes) != len(routen) {
		t.Fatalf("%d Routen, erwartet 64", len(routen))
	}
	u := neuMit(t, routen...)
	for i := range routen {
		if rec := u.rufe(anfrageMit("GET", "/r"+strconv.Itoa(i)+"/1/2/3/4", nil, nil)); rec.Code != 200 {
			t.Fatalf("Route %d: Status %d", i, rec.Code)
		}
		for j, f := range fakes {
			if want := map[bool]int{true: 1, false: 0}[j <= i]; f.aufrufe() != want {
				t.Errorf("nach Route %d: Route %d lief %d-mal, erwartet %d", i, j, f.aufrufe(), want)
			}
		}
		a, _ := fakes[i].einzige()
		if !maps.Equal(a.Pfadwerte, map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"}) {
			t.Errorf("Route %d: Pfadwerte %v", i, a.Pfadwerte)
		}
	}
}

// TestZuordnungKopiertRouten hält Z4: New kopiert die Liste der Routen; ein Überschreiben danach wirkt nicht.
func TestZuordnungKopiertRouten(t *testing.T) {
	erste, andere := neueRoute(), neueRoute()
	routen := []admin.Route{{Methode: "GET", Muster: "/a", Handler: erste.handle}}
	u := neuMit(t, routen...)
	routen[0] = admin.Route{Methode: "POST", Muster: "/zzz", Handler: andere.handle}
	if rec := u.rufe(anfrageMit("GET", "/a", nil, nil)); rec.Code != 200 || erste.aufrufe() != 1 || andere.aufrufe() != 0 {
		t.Errorf("GET /a: Status %d, erste Route %d, andere %d Aufrufe", rec.Code, erste.aufrufe(), andere.aufrufe())
	}
	if rec := u.rufe(anfrageMit("POST", "/zzz", nil, nil)); rec.Code != 404 || andere.aufrufe() != 0 {
		t.Errorf("POST /zzz: Status %d, andere Route %d Aufrufe, erwartet 404 und 0", rec.Code, andere.aufrufe())
	}
	routen[0].Handler = andere.handle
	if u.rufe(anfrageMit("GET", "/a", nil, nil)); erste.aufrufe() != 2 || andere.aufrufe() != 0 {
		t.Errorf("nach dem Überschreiben des Handlers: erste %d, andere %d", erste.aufrufe(), andere.aufrufe())
	}
}

// TestAntwortDerRoute hält Z4: Status und Body der Route gehen durch, auch ein leerer; ein Fehler der Route ist
// vorläufig ODR-E9001 mit 500.
func TestAntwortDerRoute(t *testing.T) {
	faelle := []struct {
		name    string
		antwort admin.Antwort
		err     error
		status  int
		body    string
	}{
		{"204 ohne Body", admin.Antwort{Status: 204}, nil, 204, ""},
		{"200 mit Body", admin.Antwort{Status: 200, Body: []byte(`{"a":1}`)}, nil, 200, `{"a":1}`},
		{"201 mit Body", admin.Antwort{Status: 201, Body: []byte(`{"b":2}`)}, nil, 201, `{"b":2}`},
		{"Fehler", admin.Antwort{Status: 200, Body: []byte(`{"c":3}`)}, errors.New("Fehler der Route"), 500, objekt500},
	}
	for _, f := range faelle {
		route := neueRoute()
		route.antwort, route.err = f.antwort, f.err
		u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		rec := u.rufe(anfrageMit("GET", "/a", nil, nil))
		if rec.Code != f.status || rec.Body.String() != f.body || route.aufrufe() != 1 {
			t.Errorf("%s: Antwort %d %q, Aufrufe %d, erwartet %d %q", f.name, rec.Code, rec.Body.String(), route.aufrufe(), f.status, f.body)
		}
		if f.err != nil {
			pruefeFehlerAntwort(t, f.name, rec, 500, objekt500, "")
		}
	}
}

// TestKeinBereinigen hält Z5: der Adapter bereinigt den Pfad nicht, ein Pfad ohne führenden / passt auf keine Route.
func TestKeinBereinigen(t *testing.T) {
	pfade := []string{"//", "/a/", "//a", "/a//b", "/a/./b", "/a/../a/b", "*", ""}
	methoden := []string{"GET", "POST"}
	faelle := 0
	for _, pfad := range pfade {
		for _, m := range methoden {
			faelle++
			route := neueRoute()
			u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a/{id}", Handler: route.handle}, admin.Route{Methode: "POST", Muster: "/{x}", Handler: route.handle})
			r := anfrageMit(m, "/", nil, nil)
			if pfad == "" || pfad == "*" {
				r.URL.Path, r.URL.RawPath = pfad, ""
			} else {
				r = anfrageMit(m, pfad, nil, nil)
			}
			rec := u.rufe(r)
			pruefeFehlerAntwort(t, m+" "+strconv.Quote(pfad), rec, 404, objekt404, "")
			if route.aufrufe() != 0 {
				t.Errorf("%s %q: die Route lief", m, pfad)
			}
		}
	}
	if faelle != len(pfade)*len(methoden) || faelle != 16 {
		t.Errorf("%d Fälle, erwartet 16", faelle)
	}
}

// z6Erwartung ist das Orakel zu einem Paar aus Methode und Pfad: die Route (leer: keine), sonst Status und Allow.
type z6Erwartung struct {
	route string
	id    string
	allow string
}

// z6Orakel liefert die Erwartung als Literale des Vertrags für die Routen GET /a/b, POST /a/{id}, DELETE /a/{id}, PUT /b.
func z6Orakel(methode, pfad string) z6Erwartung {
	switch pfad {
	case "/a/b":
		switch methode {
		case "GET":
			return z6Erwartung{route: "GET /a/b"}
		case "POST":
			return z6Erwartung{route: "POST /a/{id}", id: "b"}
		case "DELETE":
			return z6Erwartung{route: "DELETE /a/{id}", id: "b"}
		}
		return z6Erwartung{allow: "DELETE, GET, POST"}
	case "/b":
		if methode == "PUT" {
			return z6Erwartung{route: "PUT /b"}
		}
		return z6Erwartung{allow: "PUT"}
	case "/healthz":
		if methode == "GET" {
			return z6Erwartung{route: "healthz"}
		}
		return z6Erwartung{allow: "GET"}
	}
	return z6Erwartung{}
}

// z6Routen sind die Routen von Z6 mit je einem Fake.
func z6Routen() ([]admin.Route, map[string]*routeFake) {
	fakes := map[string]*routeFake{}
	var routen []admin.Route
	for _, r := range [][2]string{{"GET", "/a/b"}, {"POST", "/a/{id}"}, {"DELETE", "/a/{id}"}, {"PUT", "/b"}} {
		f := neueRoute()
		fakes[r[0]+" "+r[1]] = f
		routen = append(routen, admin.Route{Methode: r[0], Muster: r[1], Handler: f.handle})
	}
	return routen, fakes
}

// z6Pruefe sendet einen Request und hält Status, Allow, Aufrufe der Routen und Pfadwerte gegen die Erwartung.
func z6Pruefe(t *testing.T, methode, pfad string, soll z6Erwartung) {
	t.Helper()
	routen, fakes := z6Routen()
	u := neuMit(t, routen...)
	name := methode + " " + pfad
	rec := u.rufe(anfrageMit(methode, pfad, nil, nil))
	switch {
	case soll.route == "healthz":
		if rec.Code != 200 || rec.Body.String() != healthzObj {
			t.Errorf("%s: Antwort %d %q, erwartet 200 %q", name, rec.Code, rec.Body.String(), healthzObj)
		}
	case soll.route != "":
		a, ok := fakes[soll.route].einzige()
		wert, hat := a.Pfadwerte["id"]
		if !ok || rec.Code != 200 || hat != (soll.id != "") || wert != soll.id {
			t.Errorf("%s: Antwort %d, Aufrufe %d, id %q", name, rec.Code, fakes[soll.route].aufrufe(), wert)
		}
	case soll.allow != "":
		pruefeFehlerAntwort(t, name, rec, http.StatusMethodNotAllowed, objekt404, soll.allow)
	default:
		pruefeFehlerAntwort(t, name, rec, http.StatusNotFound, objekt404, "")
	}
	gesamt := 0
	for k, f := range fakes {
		gesamt += f.aufrufe()
		if f.aufrufe() != 0 && k != soll.route {
			t.Errorf("%s: die Route %s lief", name, k)
		}
	}
	if soll.route == "" || soll.route == "healthz" {
		if gesamt != 0 {
			t.Errorf("%s: %d Aufrufe der Routen, erwartet 0", name, gesamt)
		}
	}
	u.keinePortAufrufe(t, name)
}

// TestMethodeUndPfad hält Z6 (M-Methode × Pfad): 404 ohne passenden Pfad, 405 mit Allow aller Methoden der Routen des
// Pfades (die eingebaute eingeschlossen) nach Bytefolge, sonst die Route; Fehlerobjekt, Content-Type und Content-Length.
func TestMethodeUndPfad(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}
	pfade := []string{"/a/b", "/b", "/zzz-" + canary(), "/healthz"}
	faelle := 0
	for _, m := range methoden {
		for _, p := range pfade {
			faelle++
			z6Pruefe(t, m, p, z6Orakel(m, p))
		}
	}
	if faelle != len(methoden)*len(pfade) || faelle != 32 {
		t.Errorf("%d Fälle, erwartet 32", faelle)
	}
}

// TestMethodeSchreibweise hält Z6: die Methode des Requests gilt bytegleich (get ist nicht GET), auf jedem Pfad mit einer
// Route GET: 405 mit dem Allow aller Methoden des Pfades. Dazu die leere Liste der Routen: nur GET /healthz.
func TestMethodeSchreibweise(t *testing.T) {
	zusatz := []struct {
		methode, pfad string
		soll          z6Erwartung
	}{
		{"get", "/b", z6Erwartung{allow: "PUT"}},
		{"get", "/a/b", z6Erwartung{allow: "DELETE, GET, POST"}},
		{"get", "/healthz", z6Erwartung{allow: "GET"}},
		{"Get", "/healthz", z6Erwartung{allow: "GET"}},
	}
	for _, z := range zusatz {
		z6Pruefe(t, z.methode, z.pfad, z.soll)
	}
	u := neuMit(t)
	if rec := u.rufe(anfrageMit("GET", "/healthz", nil, nil)); rec.Code != 200 || rec.Body.String() != healthzObj {
		t.Errorf("leere Liste, GET /healthz: %d %q", rec.Code, rec.Body.String())
	}
	pruefeFehlerAntwort(t, "leere Liste, GET /x", u.rufe(anfrageMit("GET", "/x", nil, nil)), 404, objekt404, "")
	pruefeFehlerAntwort(t, "leere Liste, POST /healthz", u.rufe(anfrageMit("POST", "/healthz", nil, nil)), 405, objekt404, "GET")
}
