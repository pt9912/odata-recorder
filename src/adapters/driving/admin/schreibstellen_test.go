package admin_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// orte sind die Orte, an denen ein Request oder eine Route einen Wert trägt, der in keiner Ausgabe stehen darf (E5): je
// Ort ein eigener Wert aus Bytewerten.
type orte struct {
	pfad, query, header, body, ursache, panik, antwort, token string
}

// neueOrte liefert die Werte der Orte: sieben Canary und das Token des Servers. Keine zwei Werte teilen ein Teilstück von
// 8 Bytes, so findet ein Fund genau den Ort, dessen Wert durchsickerte.
func neueOrte() orte {
	return orte{pfad: "qa81Kz4Pwe7Mnb3X", query: "Tc59Yd2Lhs6Vru0G", header: "Jf47Bg8Nxo1Eip5W", body: "Rm36Zk0Cvt9Aly2D",
		ursache: "Hw72Qe5Ujn4Sbx8F", panik: "Xo10Ia6Gdr3Kpc7M", antwort: "Lz93Vb4Ytf8Ohe1U", token: string(tokenVielfalt())}
}

// TestOrteGetrennt hält die Orte von E5 auseinander: kein Teilstück von 8 Bytes steht in zwei Werten.
func TestOrteGetrennt(t *testing.T) {
	werte := neueOrte().alle()
	for i, a := range werte {
		for j, b := range werte {
			if i != j && enthaeltTeil(b, a) {
				t.Errorf("Wert Nr. %d (%q) und Wert Nr. %d (%q) teilen ein Teilstück von 8 Bytes", i, a, j, b)
			}
		}
	}
}

// alle liefert die acht Werte.
func (o orte) alle() []string {
	return []string{o.pfad, o.query, o.header, o.body, o.ursache, o.panik, o.antwort, o.token}
}

// request bildet den Request: Pfad, Query, Header und Body tragen je ihren Wert; kandidat ist der Kandidat des Headers
// Authorization (leer: kein Header).
func (o orte) request(methode, pfad, kandidat string) *http.Request {
	r := httptest.NewRequest(methode, pfad+"/"+o.pfad+"?q="+o.query, strings.NewReader(o.body))
	r.Header.Set("X-Test", o.header)
	if kandidat != "" {
		r.Header.Set("Authorization", "Bearer "+kandidat)
	}
	return r
}

// schreibstelle ist eine der neun Stellen W, an denen der Adapter ein Fehlerobjekt mit Code schreibt (SPEC-238 Nr. 7):
// die Erwartung als Literal des Vertrags und die Mittel, sie auszulösen.
type schreibstelle struct {
	name   string
	status int
	code   string
	body   string
	allow  string
	// aufbau stellt die Route ein und liefert den Request.
	aufbau func(u *umgebung, route *routeFake, o orte) *http.Request
}

// schreibstellen sind die neun Stellen W: 401, 404, 405, 503, ErrUngueltig, *inbound.Fehler der Route, Panik, Bruch des
// Vertrags, anderer Fehler.
func schreibstellen() []schreibstelle {
	mitAntwort := func(route *routeFake, o orte) { route.antwort = admin.Antwort{Status: 200, Body: []byte(o.antwort)} }
	get := func(o orte) *http.Request { return o.request("GET", "/r", string(tokenVielfalt())) }
	return []schreibstelle{
		{"401", 401, "ODR-E2101", objekt401, "", func(_ *umgebung, _ *routeFake, o orte) *http.Request {
			return o.request("GET", "/r", "falsch-"+o.header)
		}},
		{"404", 404, "ODR-E2102", objekt404, "", func(_ *umgebung, _ *routeFake, o orte) *http.Request {
			return o.request("GET", "/zzz", string(tokenVielfalt()))
		}},
		{"405", 405, "ODR-E2102", objekt404, "GET", func(_ *umgebung, _ *routeFake, o orte) *http.Request {
			return o.request("POST", "/r", string(tokenVielfalt()))
		}},
		{"503", 503, "ODR-E5010", objekt503, "", func(u *umgebung, _ *routeFake, o orte) *http.Request {
			u.srv.Anhalten()
			return get(o)
		}},
		{"ErrUngueltig", 400, "ODR-E2102", objekt400, "", func(_ *umgebung, route *routeFake, o orte) *http.Request {
			mitAntwort(route, o)
			route.err = fmt.Errorf("%w: %s", model.ErrUngueltig, o.ursache)
			return get(o)
		}},
		{"*inbound.Fehler", 404, "ODR-E7003", objektE7003, "", func(_ *umgebung, route *routeFake, o orte) *http.Request {
			mitAntwort(route, o)
			route.err = &inbound.Fehler{Code: model.CodeE7003, Status: 404, Ursache: errors.New(o.ursache)}
			return get(o)
		}},
		{"Panik", 500, "ODR-E9001", objekt500, "", func(_ *umgebung, route *routeFake, o orte) *http.Request {
			route.aendere = func(admin.Anfrage) { panic(o.panik) }
			return get(o)
		}},
		{"Bruch des Vertrags", 500, "ODR-E9001", objekt500, "", func(_ *umgebung, route *routeFake, o orte) *http.Request {
			mitAntwort(route, o)
			route.err = &inbound.Fehler{Code: 0, Status: 500, Ursache: errors.New(o.ursache)}
			return get(o)
		}},
		{"anderer Fehler", 500, "ODR-E9001", objekt500, "", func(_ *umgebung, route *routeFake, o orte) *http.Request {
			mitAntwort(route, o)
			route.err = errors.New(o.ursache)
			return get(o)
		}},
	}
}

// ereignisse ist das Protokoll, in das Telemetrie und ResponseWriter schreiben (C5).
type ereignisse struct {
	mu    sync.Mutex
	liste []string
}

func (e *ereignisse) add(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.liste = append(e.liste, s)
}

func (e *ereignisse) lies() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.liste)
}

// zaehlTele ist die Telemetrie des Fakes mit einem Eintrag `zaehlen` je CountError im Protokoll.
type zaehlTele struct {
	*supporttest.FakeTelemetry
	e *ereignisse
}

func (z zaehlTele) CountError(c model.Code) {
	z.e.add("zaehlen")
	z.FakeTelemetry.CountError(c)
}

// schreibWriter ist ein ResponseWriter mit einem Eintrag `kopf` je WriteHeader und `schreiben` je Write im Protokoll.
type schreibWriter struct {
	http.ResponseWriter
	e *ereignisse
}

func (w schreibWriter) WriteHeader(status int) {
	w.e.add("kopf")
	w.ResponseWriter.WriteHeader(status)
}

func (w schreibWriter) Write(b []byte) (int, error) {
	w.e.add("schreiben")
	return w.ResponseWriter.Write(b)
}

// lauf baut den Server mit der Route GET /r/{x} und dem Token tokenVielfalt, lässt die Stelle ihren Request senden und
// liefert Umgebung, Antwort und Orte; mit e gleich nil bricht jede Panik als Fehler des Tests ab, sonst schreiben
// Telemetrie und ResponseWriter in e.
func (s schreibstelle) lauf(t *testing.T, e *ereignisse) (*umgebung, *httptest.ResponseRecorder, orte) {
	t.Helper()
	route := neueRoute()
	k, u := konfigMit(admin.Route{Methode: "GET", Muster: "/r/{x}", Handler: route.handle})
	k.Token = tokenVielfalt()
	if e != nil {
		k.Telemetrie = zaehlTele{FakeTelemetry: u.tele, e: e}
	}
	neuMitKonfig(t, k, u)
	o := neueOrte()
	r := s.aufbau(u, route, o)
	rec := httptest.NewRecorder()
	var w http.ResponseWriter = rec
	if e != nil {
		w = schreibWriter{ResponseWriter: rec, e: e}
	}
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Errorf("%s: eine Panik verließ ServeHTTP: %v", s.name, p)
			}
		}()
		u.srv.ServeHTTP(w, r)
	}()
	return u, rec, o
}

// TestSchreibstellenZahl hält M-Schreibstellen W gegen die Aufzählung in SPEC-238 Nr. 7: neun Stellen, je eine mit eigenem
// Namen; die Spezifikation liegt im Container der Tests nicht vor, die Zahl ist die der Aufzählung dort.
func TestSchreibstellenZahl(t *testing.T) {
	namen := map[string]bool{}
	for _, s := range schreibstellen() {
		namen[s.name] = true
	}
	gleichAuf(t, "Zahl der Schreibstellen", len(namen), 9)
}

// TestZaehlenJedeSchreibstelle hält C1 (Nr. 7): an jeder der neun Schreibstellen ruft der Adapter `CountError` genau einmal
// mit dem Code der Antwort und sonst keine Operation der Telemetrie und kein Audit.
func TestZaehlenJedeSchreibstelle(t *testing.T) {
	for _, s := range schreibstellen() {
		u, rec, _ := s.lauf(t, nil)
		pruefeFehlerAntwort(t, s.name, rec, s.status, s.body, s.allow)
		if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{s.code}) || andere != 0 || u.audit.Aufrufe() != 0 {
			t.Errorf("%s: CountError %q, andere Aufrufe der Telemetrie %d, Audit %d, erwartet einmal %s und sonst 0",
				s.name, codes, andere, u.audit.Aufrufe(), s.code)
		}
	}
}

// TestZaehlenVorDemSchreiben hält C5 (Nr. 7): an jeder Schreibstelle steht der Eintrag `zaehlen` vor jedem `kopf` und
// `schreiben` im gemeinsamen Protokoll von Telemetrie und ResponseWriter, je einmal.
func TestZaehlenVorDemSchreiben(t *testing.T) {
	for _, s := range schreibstellen() {
		var e ereignisse
		s.lauf(t, &e)
		gleichAuf(t, s.name+": Ereignisse", e.lies(), []string{"zaehlen", "kopf", "schreiben"})
	}
}

// TestKeinWertImFehlerobjekt hält E5 (Nr. 5, 13): an jeder der neun Schreibstellen steht keiner der acht Werte (Pfad,
// Query, Header-Wert, Body des Requests, Ursache des Fehlers der Route, Panikwert, Body der Antwort der Route, Token), auch
// nicht als Teilstück ab 8 Bytes, in Status, Header oder Body der Antwort und nicht in einem Aufruf der Telemetrie.
func TestKeinWertImFehlerobjekt(t *testing.T) {
	geprueft := 0
	for _, s := range schreibstellen() {
		u, rec, o := s.lauf(t, nil)
		text := textVon(rec) + " " + fmt.Sprintf("%+v", u.tele.Aufrufe())
		for i, wert := range o.alle() {
			geprueft++
			if enthaeltTeil(text, wert) {
				t.Errorf("%s: der Wert Nr. %d (%q) steht in der Antwort oder der Telemetrie: %s", s.name, i, wert, text)
			}
		}
	}
	gleichAuf(t, "Zahl der Prüfungen", geprueft, 9*8)
}

// TestErfolgZaehltNicht hält C2 (Nr. 7): eine Erfolgsantwort (healthz, 200, 201, 204) zählt nicht und ruft keinen Port; der
// Kontextfehler ohne Code zählt ebenso nicht (X1).
func TestErfolgZaehltNicht(t *testing.T) {
	faelle := []struct {
		name, ziel string
		antwort    admin.Antwort
	}{
		{"healthz", "/healthz", admin.Antwort{}},
		{"200", "/r", admin.Antwort{Status: 200, Body: []byte(`{"a":1}`)}},
		{"201", "/r", admin.Antwort{Status: 201, Body: []byte(`{"a":1}`)}},
		{"204", "/r", admin.Antwort{Status: 204}},
	}
	for _, f := range faelle {
		u, route := einRoute(t)
		route.antwort = f.antwort
		if rec := u.rufeSicher(t, f.name, anfrageMit("GET", f.ziel, nil, nil)); rec.Code != max(f.antwort.Status, 200) {
			t.Errorf("%s: Status %d", f.name, rec.Code)
		}
		u.keinePortAufrufe(t, f.name)
	}
	gleichAuf(t, "Zahl der Fälle", len(faelle), 4)
}
