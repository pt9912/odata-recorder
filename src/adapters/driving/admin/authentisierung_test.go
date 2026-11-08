package admin_test

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// u1Form ist eine ungültige Form des Headers Authorization.
type u1Form struct {
	name string
	kopf map[string][]string
}

// u1Formen sind die 20 ungültigen Formen von M-Header (U1); k ist das Token, wo eine Form eines trägt. Die Formen 4, 7
// und 20 sind direkte Aufrufe (net/http kürzt Leerraum am Rand und kanonisiert den Schlüssel auf der Leitung).
func u1Formen(k string) []u1Form {
	const a = "Authorization"
	gueltig := "Bearer " + k
	return []u1Form{
		{"1 fehlt", nil},
		{"2 leere Liste", map[string][]string{a: {}}},
		{"3 leerer Wert", map[string][]string{a: {""}}},
		{"4 nur Leerzeichen", map[string][]string{a: {"   "}}},
		{"5 Bea", map[string][]string{a: {"Bea"}}},
		{"6 Bearer allein", map[string][]string{a: {"Bearer"}}},
		{"7 Bearer und Leerzeichen", map[string][]string{a: {"Bearer   "}}},
		{"8 Basic", map[string][]string{a: {"Basic " + k}}},
		{"9 Token", map[string][]string{a: {"Token " + k}}},
		{"10 Bearer ohne Leerzeichen", map[string][]string{a: {"Bearer" + k}}},
		{"11 Tab", map[string][]string{a: {"Bearer\t" + k}}},
		{"12 ohne Schema", map[string][]string{a: {k}}},
		{"13 zwei gleichlautende", map[string][]string{a: {gueltig, gueltig}}},
		{"14 zwei Schreibweisen", map[string][]string{a: {"Bearer " + k, "bearer " + k}}},
		{"15 gültig, dann Basic", map[string][]string{a: {gueltig, "Basic x"}}},
		{"16 Basic, dann gültig", map[string][]string{a: {"Basic x", gueltig}}},
		{"17 leer, dann gültig", map[string][]string{a: {"", gueltig}}},
		{"18 drei gültige", map[string][]string{a: {gueltig, gueltig, gueltig}}},
		{"19 X-Authorization", map[string][]string{"X-Authorization": {gueltig}}},
		{"20 nichtkanonischer Schlüssel", map[string][]string{"authorization": {gueltig}}},
	}
}

// requestArt ist eine Requestart aus R (27) mit der Erwartung des Orakels; angehalten heißt: der Server ist vor dem
// Request angehalten.
type requestArt struct {
	name, methode, ziel string
	kopf                map[string][]string
	vergleich           bool
	status              int
	body                string
	angehalten          bool
}

// requestArten sind die 27 Requestarten R: healthz, die 20 Formen von U1, falsches Token, 404, 405, Route mit 2xx,
// Route mit Fehler und der Request eines angehaltenen Servers (503). Status und Body stehen als Literale des Vertrags.
func requestArten() []requestArt {
	tok := string(tokenGueltig())
	gueltig := bearer(tok)
	arten := []requestArt{{"healthz", "GET", "/healthz", nil, true, 200, healthzObj, false}}
	for _, f := range u1Formen(tok) {
		arten = append(arten, requestArt{"U1 " + f.name, "GET", "/a", f.kopf, true, 401, objekt401, false})
	}
	return append(arten,
		requestArt{"falsches Token", "GET", "/a", bearer("falsch"), false, 401, objekt401, false},
		requestArt{"404", "GET", "/zzz", gueltig, true, 404, objekt404, false},
		requestArt{"405", "POST", "/a", gueltig, true, 405, objekt404, false},
		requestArt{"Route 2xx", "GET", "/a", gueltig, true, 200, routeBody, false},
		requestArt{"Route Fehler", "GET", "/fehler", gueltig, true, 500, objekt500, false},
		requestArt{"503", "GET", "/a", gueltig, true, 503, objekt503, true},
	)
}

// lauf ist ein Server mit den Fakes einer Requestart; die Routen rufen beide den Use-Case-Fake.
type lauf struct {
	u          *umgebung
	vf         *vergleichFake
	ok, fehler *routeFake
	uc         *useCaseFake
}

// baue baut den Server der Requestart a (Routen GET /a mit Body und GET /fehler mit Fehler) ohne einen Request; bei
// angehalten hält es ihn an.
func baue(t *testing.T, a requestArt) *lauf {
	t.Helper()
	uc := &useCaseFake{}
	ok, fehler := routeMitUseCase(uc), routeMitUseCase(uc)
	ok.antwort = admin.Antwort{Status: 200, Body: []byte(routeBody)}
	fehler.err = errors.New("Fehler der Route")
	vf := neuerVergleich(a.vergleich)
	u := neuMitVergleich(t, nil, vf,
		admin.Route{Methode: "GET", Muster: "/a", Handler: ok.handle},
		admin.Route{Methode: "GET", Muster: "/fehler", Handler: fehler.handle})
	if a.angehalten {
		u.srv.Anhalten()
	}
	return &lauf{u: u, vf: vf, ok: ok, fehler: fehler, uc: uc}
}

// TestRequestArtenZahl hält die Mengen R (27) und M-Header (20) gegen die Quelle des Tests.
func TestRequestArtenZahl(t *testing.T) {
	if n, m := len(requestArten()), len(u1Formen("x")); n != 27 || m != 20 {
		t.Errorf("%d Requestarten und %d Formen von Authorization, erwartet 27 und 20", n, m)
	}
}

// TestKorrelationsIDJeRequestart hält A1: IDs.KorrelationsID läuft je Request genau einmal, die anderen Operationen nie,
// und beim Aufruf haben Vergleich, Routen, Audit und Telemetrie noch nichts gesehen; Schleife über R.
func TestKorrelationsIDJeRequestart(t *testing.T) {
	for _, a := range requestArten() {
		l := baue(t, a)
		vorVergleich, vorOk, vorFehler, vorAudit, vorTele := -1, -1, -1, -1, -1
		l.u.ids.vor = func() {
			vorVergleich, vorOk, vorFehler = l.vf.zahl(), l.ok.aufrufe(), l.fehler.aufrufe()
			vorAudit, vorTele = l.u.audit.Aufrufe(), len(l.u.tele.Aufrufe())
		}
		l.u.rufe(ohneAuth(a.methode, a.ziel, a.kopf))
		if k, uebrige := l.u.ids.zahlen(); k != 1 || uebrige != 0 {
			t.Errorf("%s: KorrelationsID %d-mal, andere Operationen %d-mal, erwartet 1 und 0", a.name, k, uebrige)
		}
		if vorVergleich != 0 || vorOk != 0 || vorFehler != 0 || vorAudit != 0 || vorTele != 0 {
			t.Errorf("%s: beim Aufruf von KorrelationsID liefen Vergleich %d, Routen %d und %d, Audit %d, Telemetrie %d",
				a.name, vorVergleich, vorOk, vorFehler, vorAudit, vorTele)
		}
	}
}

// TestKorrelationsHeader hält E4: jede Antwort trägt X-Correlation-Id gleich der ID, die der Generator für diesen Request
// vergab; zwei Requests desselben Servers tragen verschiedene. Schleife über R.
func TestKorrelationsHeader(t *testing.T) {
	for _, a := range requestArten() {
		l := baue(t, a)
		var ids []string
		for range 2 {
			rec := l.u.rufe(ohneAuth(a.methode, a.ziel, a.kopf))
			got := kopfWerte(rec, "X-Correlation-Id")
			if len(got) != 1 || got[0] == "" || got[0] != l.u.ids.letzte() {
				t.Errorf("%s: X-Correlation-Id %q, vergeben %q", a.name, got, l.u.ids.letzte())
			}
			ids = append(ids, got...)
		}
		if len(ids) != 2 || ids[0] == ids[1] {
			t.Errorf("%s: die IDs zweier Requests %q sind nicht verschieden", a.name, ids)
		}
	}
}

// e2Zeile ist eine Antwortart mit den Werten der drei Header, die nur bei einer Art stehen.
type e2Zeile struct {
	art              string
	www, allow, conn []string
}

// TestHeaderJeAntwortart hält E2: WWW-Authenticate nur bei 401, Allow nur bei 405, Connection: close nur bei 503: sieben
// Antwortarten × drei Header = 21 Zellen mit Literalen des Vertrags.
func TestHeaderJeAntwortart(t *testing.T) {
	zeilen := []e2Zeile{
		{"U1 1 fehlt", []string{"Bearer"}, nil, nil},
		{"404", nil, nil, nil},
		{"405", nil, []string{"GET"}, nil},
		{"Route 2xx", nil, nil, nil},
		{"healthz", nil, nil, nil},
		{"Route Fehler", nil, nil, nil},
		{"503", nil, nil, []string{"close"}},
	}
	zellen := 0
	for _, z := range zeilen {
		for _, a := range requestArten() {
			if a.name != z.art {
				continue
			}
			rec := baue(t, a).u.rufe(ohneAuth(a.methode, a.ziel, a.kopf))
			if rec.Code != a.status || rec.Body.String() != a.body {
				t.Errorf("%s: Antwort %d %q, erwartet %d %q", a.name, rec.Code, rec.Body.String(), a.status, a.body)
			}
			for name, soll := range map[string][]string{"WWW-Authenticate": z.www, "Allow": z.allow, "Connection": z.conn} {
				zellen++
				if got := kopfWerte(rec, name); !slices.Equal(got, soll) {
					t.Errorf("%s: %s %q, erwartet %q", a.name, name, got, soll)
				}
			}
		}
	}
	if zellen != 21 {
		t.Errorf("%d Zellen geprüft, erwartet 21", zellen)
	}
}

// TestErfolgsantwort hält E3: die Erfolgsantwort trägt den Status der Route; ein nichtleerer Body mit Content-Type und
// Content-Length, ein leerer ohne Body.
func TestErfolgsantwort(t *testing.T) {
	faelle := []struct {
		name    string
		antwort admin.Antwort
		ziel    string
		status  int
		body    string
		laenge  string
	}{
		{"200 mit Body", admin.Antwort{Status: 200, Body: []byte(`{"a":1}`)}, "/a", 200, `{"a":1}`, "7"},
		{"201 mit Body", admin.Antwort{Status: 201, Body: []byte(`{"bb":22}`)}, "/a", 201, `{"bb":22}`, "9"},
		{"204 ohne Body", admin.Antwort{Status: 204}, "/a", 204, "", ""},
		{"200 mit leerem Body", admin.Antwort{Status: 200, Body: []byte{}}, "/a", 200, "", ""},
		{"200 mit nil als Body", admin.Antwort{Status: 200, Body: nil}, "/a", 200, "", ""},
		{"GET /healthz", admin.Antwort{}, "/healthz", 200, healthzObj, "15"},
	}
	if len(faelle) != 6 {
		t.Fatalf("%d Fälle, erwartet 6", len(faelle))
	}
	for _, f := range faelle {
		route := neueRoute()
		route.antwort = f.antwort
		u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		rec := u.rufe(anfrageMit("GET", f.ziel, nil, nil))
		typ, laenge := []string{"application/json"}, []string{f.laenge}
		if f.body == "" {
			typ, laenge = nil, nil
		}
		if rec.Code != f.status || rec.Body.String() != f.body ||
			!slices.Equal(kopfWerte(rec, "Content-Type"), typ) || !slices.Equal(kopfWerte(rec, "Content-Length"), laenge) {
			t.Errorf("%s: Antwort %d %q, Content-Type %q, Content-Length %q, erwartet %d %q, %q, %q", f.name, rec.Code,
				rec.Body.String(), kopfWerte(rec, "Content-Type"), kopfWerte(rec, "Content-Length"), f.status, f.body, typ, laenge)
		}
	}
}

// TestGueltigeAuthentisierung hält Nr. 4 b und c im Gutfall: ein gültiger Request erreicht die Route genau einmal, und
// Vergleich lief einmal mit dem Token und dem Kandidaten.
func TestGueltigeAuthentisierung(t *testing.T) {
	route := neueRoute()
	vf := neuerVergleich(true)
	u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	rec := u.rufe(ohneAuth("GET", "/a", bearer("kandidat-lang-genug-aber-falsch")))
	args, ok := vf.einziger()
	if rec.Code != 200 || route.aufrufe() != 1 || !ok || args != [2]string{string(tokenGueltig()), "kandidat-lang-genug-aber-falsch"} {
		t.Errorf("Antwort %d, Route %d, Vergleich %v (ein Aufruf %v)", rec.Code, route.aufrufe(), args, ok)
	}
}

// TestNewKopiertToken hält K4: ein Überschreiben des Quellslice nach New wirkt nicht; das ursprüngliche Token
// authentisiert, das überschriebene nicht.
func TestNewKopiertToken(t *testing.T) {
	route := neueRoute()
	k, u := konfigMit(admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	k.Token = tokenVielfalt()
	original := string(k.Token)
	neuMitKonfig(t, k, u)
	clear(k.Token)
	if rec := u.rufe(ohneAuth("GET", "/a", bearer(original))); rec.Code != 200 {
		t.Errorf("das ursprüngliche Token: Status %d, erwartet 200", rec.Code)
	}
	if rec := u.rufe(ohneAuth("GET", "/a", bearer(string(k.Token)))); rec.Code != 401 || route.aufrufe() != 1 {
		t.Errorf("das überschriebene Token: Status %d, Route %d Aufrufe, erwartet 401 und 1", rec.Code, route.aufrufe())
	}
}

// TestVergleichNil hält K5: ohne Vergleich entscheidet vergleiche; das richtige Token authentisiert, jede Abweichung
// nicht (Token plus ein Byte, ohne letztes Byte, letztes Byte geändert, Präfix, ein Byte).
func TestVergleichNil(t *testing.T) {
	tok := string(tokenVielfalt())
	letzteGeaendert := tok[:len(tok)-1] + "M"
	faelle := []struct {
		name, kandidat string
		status         int
	}{
		{"richtiges Token", tok, 200},
		{"Token plus ein Byte", tok + "x", 401},
		{"ohne letztes Byte", tok[:len(tok)-1], 401},
		{"letztes Byte geändert", letzteGeaendert, 401},
		{"Präfix", tok[:16], 401},
		{"ein Byte", "x", 401},
	}
	if tok[len(tok)-1] == 'M' || len(faelle) != 6 {
		t.Fatalf("Fixture: letztes Byte %q, %d Fälle, erwartet ein anderes als M und 6", tok[len(tok)-1], len(faelle))
	}
	for _, f := range faelle {
		route := neueRoute()
		u := neuMitVergleich(t, tokenVielfalt(), nil, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		rec := u.rufe(ohneAuth("GET", "/a", bearer(f.kandidat)))
		if rec.Code != f.status || route.aufrufe() != map[int]int{200: 1, 401: 0}[f.status] {
			t.Errorf("%s: Status %d, Route %d Aufrufe, erwartet %d", f.name, rec.Code, route.aufrufe(), f.status)
		}
	}
}

// TestHealthzMitFormen hält A3: GET /healthz antwortet 200 mit dem Literal für jede der 20 Formen von Authorization und
// für das richtige Token (21), ohne Vergleich, Audit, Telemetrie oder Route.
func TestHealthzMitFormen(t *testing.T) {
	tok := string(tokenGueltig())
	kopfe := []u1Form{{"richtiges Token", bearer(tok)}}
	kopfe = append(kopfe, u1Formen(tok)...)
	if len(kopfe) != 21 {
		t.Fatalf("%d Fälle, erwartet 21", len(kopfe))
	}
	for _, f := range kopfe {
		route := neueRoute()
		vf := neuerVergleich(true)
		u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		rec := u.rufe(ohneAuth("GET", "/healthz", f.kopf))
		if rec.Code != 200 || rec.Body.String() != healthzObj ||
			!slices.Equal(kopfWerte(rec, "Content-Length"), []string{strconv.Itoa(laengeHZ)}) {
			t.Errorf("%s: Antwort %d %q", f.name, rec.Code, rec.Body.String())
		}
		if vf.zahl() != 0 || route.aufrufe() != 0 {
			t.Errorf("%s: Vergleich %d, Route %d Aufrufe, erwartet 0", f.name, vf.zahl(), route.aufrufe())
		}
		u.keinePortAufrufe(t, f.name)
	}
}

// healthzOrakel liefert die Erwartung zu Methode und Pfad von M-healthz: ohne Token 200 für GET /healthz, sonst 401; mit
// Token 200, 405 mit Allow GET für jede andere Methode auf /healthz, sonst 404.
func healthzOrakel(methode, pfad string, mitToken bool) (status int, allow string) {
	switch {
	case methode == "GET" && pfad == "/healthz":
		return 200, ""
	case !mitToken:
		return 401, ""
	case pfad == "/healthz":
		return 405, "GET"
	}
	return 404, ""
}

// TestHealthzMethodeUndPfad hält A4 (M-healthz): nur GET auf genau /healthz ist frei; jede andere Methode und jeder andere
// Pfad geht durch die Authentisierung (401 ohne Token), mit Token 405 mit Allow GET auf /healthz, sonst 404; die Query
// ist gleichgültig.
func TestHealthzMethodeUndPfad(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE", "get"}
	pfade := []string{"/healthz", "/healthz/", "/HEALTHZ", "/healthz/x", "/healthz%2F", "/healthzx", "/xhealthz", "/health", "//healthz"}
	faelle := 0
	for _, m := range methoden {
		for _, p := range pfade {
			for _, mit := range []bool{false, true} {
				faelle++
				u := neuMit(t)
				kopf := map[string][]string(nil)
				if mit {
					kopf = bearer(string(tokenGueltig()))
				}
				rec := u.rufe(ohneAuth(m, p, kopf))
				status, allow := healthzOrakel(m, p, mit)
				switch status {
				case 200:
					if rec.Body.String() != healthzObj {
						t.Errorf("%s %s (Token %v): Body %q", m, p, mit, rec.Body.String())
					}
				case 401:
					pruefe401(t, m+" "+p, rec)
				case 405:
					pruefeFehlerAntwort(t, m+" "+p, rec, 405, objekt404, allow)
				default:
					pruefeFehlerAntwort(t, m+" "+p, rec, 404, objekt404, "")
				}
			}
		}
	}
	if faelle != 81*2 {
		t.Errorf("%d Fälle, erwartet 81 × 2 (ohne und mit Token)", faelle)
	}
	for _, q := range []string{"", "?a=1", "?"} {
		u := neuMit(t)
		if rec := u.rufe(ohneAuth("GET", "/healthz"+q, nil)); rec.Code != 200 || rec.Body.String() != healthzObj {
			t.Errorf("GET /healthz%s ohne Token: %d %q", q, rec.Code, rec.Body.String())
		}
	}
}

// TestAuthentisierungVorZuordnung hält A5: die Menge von Z6 (32 Paare aus Methode und Pfad) ohne Token und mit falschem
// Token ergibt 401 (außer dem freien GET /healthz: 200), mit Token die Antwort der Zuordnung.
func TestAuthentisierungVorZuordnung(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}
	pfade := []string{"/a/b", "/b", "/zzz-" + canary(), "/healthz"}
	paare := 0
	for _, m := range methoden {
		for _, p := range pfade {
			paare++
			frei := m == "GET" && p == "/healthz"
			for _, modus := range []string{"ohne Token", "falsches Token", "mit Token"} {
				routen, fakes := z6Routen()
				vf := neuerVergleich(modus == "mit Token")
				u := neuMitVergleich(t, nil, vf, routen...)
				var kopf map[string][]string
				if modus != "ohne Token" {
					kopf = bearer(string(tokenGueltig()))
				}
				rec := u.rufe(ohneAuth(m, p, kopf))
				name := m + " " + p + " " + modus
				switch {
				case frei:
					if rec.Code != 200 || rec.Body.String() != healthzObj {
						t.Errorf("%s: %d %q", name, rec.Code, rec.Body.String())
					}
				case modus != "mit Token":
					pruefe401(t, name, rec)
					for k, f := range fakes {
						if f.aufrufe() != 0 {
							t.Errorf("%s: die Route %s lief", name, k)
						}
					}
				default:
					soll, want := z6Orakel(m, p), 404
					if soll.route != "" {
						want = 200
					} else if soll.allow != "" {
						want = 405
					}
					if rec.Code != want {
						t.Errorf("%s: Status %d, erwartet %d", name, rec.Code, want)
					}
				}
			}
		}
	}
	if paare != 32 {
		t.Errorf("%d Paare, erwartet 32", paare)
	}
}

// TestHandlerGenauEinmal hält A6: der Handler läuft genau einmal je Request mit gültiger Zuordnung und Token, für jede
// der fünf Methoden und auch dann, wenn er einen Fehler liefert.
func TestHandlerGenauEinmal(t *testing.T) {
	methoden := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, mitFehler := range []bool{false, true} {
		fakes := map[string]*routeFake{}
		var routen []admin.Route
		for _, m := range methoden {
			f := neueRoute()
			if mitFehler {
				f.err = errors.New("Fehler der Route")
			}
			fakes[m] = f
			routen = append(routen, admin.Route{Methode: m, Muster: "/r/{id}", Handler: f.handle})
		}
		u := neuMit(t, routen...)
		for _, m := range methoden {
			rec := u.rufe(anfrageMit(m, "/r/1", nil, nil))
			for k, f := range fakes {
				want := map[bool]int{true: 1, false: 0}[k == m]
				if f.aufrufe() != want {
					t.Errorf("%s (Fehler %v): Route %s lief %d-mal, erwartet %d", m, mitFehler, k, f.aufrufe(), want)
				}
			}
			if want := map[bool]int{true: 500, false: 200}[mitFehler]; rec.Code != want {
				t.Errorf("%s (Fehler %v): Status %d, erwartet %d", m, mitFehler, rec.Code, want)
			}
			fakes[m].mu.Lock()
			fakes[m].anfragen, fakes[m].ctxs = nil, nil
			fakes[m].mu.Unlock()
		}
	}
}

// TestFormenUngueltigerAuthorization hält U1 und U4 (für die Formen), U8: jede der 20 Formen ergibt 401 mit dem Literal,
// ohne Aufruf von Vergleich, Route oder Use Case.
func TestFormenUngueltigerAuthorization(t *testing.T) {
	formen := u1Formen(string(tokenGueltig()))
	if len(formen) != 20 {
		t.Fatalf("%d Formen, erwartet 20", len(formen))
	}
	for _, f := range formen {
		uc := &useCaseFake{}
		route := routeMitUseCase(uc)
		vf := neuerVergleich(true)
		u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		pruefe401(t, f.name, u.rufe(ohneAuth("GET", "/a", f.kopf)))
		if vf.zahl() != 0 || route.aufrufe() != 0 || uc.summe() != 0 {
			t.Errorf("%s: Vergleich %d, Route %d, Use Case %d Aufrufe, erwartet 0", f.name, vf.zahl(), route.aufrufe(), uc.summe())
		}
	}
}

// TestSchemaJedesByte hält U1 (Nr. 4 b) gegen die Klasse "anderes Schema": an jeder der sechs Stellen des Schemas jedes
// Byte außer den beiden Schreibweisen des Buchstabens (6 × 254 = 1524) und jedes Byte hinter dem Schema außer dem
// Leerzeichen (255) ergibt 401 ohne Aufruf von Vergleich; allein das Leerzeichen hinter dem Schema authentisiert (256
// Bytes an dieser Stelle).
func TestSchemaJedesByte(t *testing.T) {
	const klein = "bearer"
	tok := string(tokenGueltig())
	route := neueRoute()
	vf := neuerVergleich(true)
	u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	anStelle, hinterSchema := 0, 0
	for i := 0; i < len(klein); i++ {
		for b := 0; b < 256; b++ {
			c := byte(b)
			if c == klein[i] || c == klein[i]-('a'-'A') {
				continue
			}
			anStelle++
			wert := []byte("Bearer " + tok)
			wert[i] = c
			if rec := u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": {string(wert)}})); rec.Code != 401 {
				t.Errorf("Stelle %d, Byte %#x: Status %d, erwartet 401", i, b, rec.Code)
			}
		}
	}
	for b := 0; b < 256; b++ {
		hinterSchema++
		wert := "Bearer" + string([]byte{byte(b)}) + tok
		rec := u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": {wert}}))
		if want := map[bool]int{true: 200, false: 401}[b == ' ']; rec.Code != want {
			t.Errorf("Byte %#x hinter dem Schema: Status %d, erwartet %d", b, rec.Code, want)
		}
	}
	if anStelle != 1524 || hinterSchema != 256 || vf.zahl() != 1 || route.aufrufe() != 1 {
		t.Errorf("%d und %d Fälle (erwartet 1524 und 256), Vergleich %d, Route %d Aufrufe (erwartet 1 und 1)",
			anStelle, hinterSchema, vf.zahl(), route.aufrufe())
	}
}

// TestKandidatAlleBytes hält U3 gegen die Klasse "Byte des Kandidaten": ein Token aus allen 256 Bytewerten authentisiert
// mit dem gleichen Kandidaten (200), und jede der 256 Stellen mit einem um eins erhöhten Byte ergibt 401.
func TestKandidatAlleBytes(t *testing.T) {
	tok := make([]byte, 256)
	for i := range tok {
		tok[i] = byte(i)
	}
	route := neueRoute()
	u := neuMitVergleich(t, tok, nil, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	if rec := u.rufe(ohneAuth("GET", "/a", bearer(string(tok)))); rec.Code != 200 {
		t.Errorf("gleiches Token: Status %d, erwartet 200", rec.Code)
	}
	stellen := 0
	for i := range tok {
		stellen++
		kandidat := bytes.Clone(tok)
		kandidat[i]++
		if rec := u.rufe(ohneAuth("GET", "/a", bearer(string(kandidat)))); rec.Code != 401 {
			t.Errorf("Stelle %d geändert: Status %d, erwartet 401", i, rec.Code)
		}
	}
	if stellen != 256 || route.aufrufe() != 1 {
		t.Errorf("%d Stellen (erwartet 256), Route %d Aufrufe (erwartet 1)", stellen, route.aufrufe())
	}
}

// TestKandidatRaender hält U3 (SPEC-238 Nr. 4 b: der Rest hinter den führenden Leerzeichen ist der Kandidat, Byte für
// Byte unverändert) gegen die Klasse "Rand des Kandidaten": jedes der 256 Bytes am Anfang (hinter einem und hinter drei
// Leerzeichen; allein das Leerzeichen gehört zu den führenden und fällt weg, 255 Bytes bleiben), am Ende (256) und an
// beiden Rändern zugleich (255) erreicht Vergleich unverändert, genau einmal; direkter Aufruf, ohne Leitung.
func TestKandidatRaender(t *testing.T) {
	kern := string(tokenVielfalt())
	type fall struct{ name, wert, kandidat string }
	var faelle []fall
	for b := 0; b < 256; b++ {
		c := string([]byte{byte(b)})
		if b != ' ' {
			for _, n := range []int{1, 3} {
				faelle = append(faelle, fall{fmt.Sprintf("Anfang %#x nach %d Leerzeichen", b, n),
					"Bearer " + strings.Repeat(" ", n-1) + c + kern, c + kern})
			}
			faelle = append(faelle, fall{fmt.Sprintf("Anfang und Ende %#x", b), "Bearer " + c + kern + c, c + kern + c})
		}
		faelle = append(faelle, fall{fmt.Sprintf("Ende %#x", b), "Bearer " + kern + c, kern + c})
	}
	if len(faelle) != 255*2+256+255 {
		t.Fatalf("%d Fälle, erwartet 255 × 2 + 256 + 255", len(faelle))
	}
	for _, f := range faelle {
		vf := neuerVergleich(true)
		u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: neueRoute().handle})
		rec := u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": {f.wert}}))
		args, ok := vf.einziger()
		if rec.Code != 200 || !ok || args[1] != f.kandidat {
			t.Errorf("%s: Status %d, Kandidat %q (ein Aufruf %v), erwartet %q", f.name, rec.Code, args[1], ok, f.kandidat)
		}
	}
}

// TestSchemaUndLeerzeichen hält U2 (M-Schema): das Schema gilt in jeder ASCII-Schreibweise (64) mit einem, zwei und fünf
// Leerzeichen (3): 192 Fälle, 200 und der Kandidat ist das Token.
func TestSchemaUndLeerzeichen(t *testing.T) {
	tok := string(tokenVielfalt())
	schemata := artSchema()
	verschieden := map[string]bool{}
	faelle := 0
	for _, s := range schemata {
		verschieden[s] = true
		for _, n := range []int{1, 2, 5} {
			faelle++
			route := neueRoute()
			vf := neuerVergleich(false)
			vf.echt = true
			u := neuMitVergleich(t, tokenVielfalt(), vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
			wert := s + strings.Repeat(" ", n) + tok
			rec := u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": {wert}}))
			args, ok := vf.einziger()
			if rec.Code != 200 || route.aufrufe() != 1 || !ok || args[1] != tok {
				t.Errorf("%q: Status %d, Route %d, Kandidat %q (ein Aufruf %v)", wert, rec.Code, route.aufrufe(), args[1], ok)
			}
		}
	}
	if faelle != 192 || len(verschieden) != 64 || !verschieden["Bearer"] {
		t.Errorf("%d Fälle aus %d Schreibweisen, erwartet 192 aus 64", faelle, len(verschieden))
	}
}

// u3Fall ist ein Kandidat von M-Kandidat mit der Erwartung: Status, und der Kandidat, den Vergleich erhält.
type u3Fall struct {
	name     string
	token    []byte
	wert     string
	status   int
	kandidat string
}

// TestKandidatByteFuerByte hält U3 (M-Kandidat, 7 Formen): der Kandidat ist der Rest Byte für Byte; Vergleich läuft
// einmal mit dem Rest als Kandidat, die Entscheidung trifft der echte Vergleich.
func TestKandidatByteFuerByte(t *testing.T) {
	tok := string(tokenVielfalt())
	mitBOM := append(bom(), tokenVielfalt()...)
	faelle := []u3Fall{
		{"1 Token plus x", tokenVielfalt(), "Bearer " + tok + "x", 401, tok + "x"},
		{"2 ohne letztes Byte", tokenVielfalt(), "Bearer " + tok[:len(tok)-1], 401, tok[:len(tok)-1]},
		{"3 erstes Byte klein", tokenVielfalt(), "Bearer a" + tok[1:], 401, "a" + tok[1:]},
		{"4 führender Tab", tokenVielfalt(), "Bearer \t" + tok, 401, "\t" + tok},
		{"5 Schema doppelt", tokenVielfalt(), "Bearer Bearer " + tok, 401, "Bearer " + tok},
		{"6 innerer Leerraum", tokenVielfalt(), "Bearer " + tok, 200, tok},
		{"7 BOM, Kandidat mit BOM", mitBOM, "Bearer " + string(mitBOM), 200, string(mitBOM)},
		{"7 BOM, Kandidat ohne BOM", mitBOM, "Bearer " + tok, 401, tok},
	}
	if tok[0] != 'A' || !strings.Contains(tok, " ") || !bytes.HasPrefix(mitBOM, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatalf("Fixture: Token %q ohne Großbuchstaben am Anfang oder inneres Leerzeichen, oder BOM fehlt", tok)
	}
	for _, f := range faelle {
		route := neueRoute()
		vf := neuerVergleich(false)
		vf.echt = true
		u := neuMitVergleich(t, f.token, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		rec := u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": {f.wert}}))
		args, ok := vf.einziger()
		if rec.Code != f.status || !ok || args != [2]string{string(f.token), f.kandidat} {
			t.Errorf("%s: Status %d, Vergleich %q (ein Aufruf %v), erwartet %d und Kandidat %q", f.name, rec.Code, args, ok, f.status, f.kandidat)
		}
	}
}

// TestTokenLaengen hält die Prüfgrenze von U2 und U3: Tokens der Längen 32, 33 und 8192 authentisieren mit dem
// gleichen Kandidaten und scheitern mit einem Byte mehr, einem weniger und einem geänderten letzten.
func TestTokenLaengen(t *testing.T) {
	for _, n := range []int{32, 33, 8192} {
		tok := make([]byte, n)
		for i := range tok {
			tok[i] = byte('a' + i*7%26)
		}
		varianten := map[string]string{
			string(tok): "gleich", string(tok) + "x": "ein Byte mehr", string(tok[:n-1]): "ein Byte weniger",
			string(tok[:n-1]) + "A": "letztes Byte geändert",
		}
		for kandidat, name := range varianten {
			route := neueRoute()
			u := neuMitVergleich(t, tok, nil, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
			rec := u.rufe(ohneAuth("GET", "/a", bearer(kandidat)))
			if want := map[bool]int{true: 200, false: 401}[name == "gleich"]; rec.Code != want {
				t.Errorf("Länge %d, %s: Status %d, erwartet %d", n, name, rec.Code, want)
			}
		}
	}
}

// TestVergleichEntscheidetAllein hält U4: Vergleich läuft genau einmal für einen nichtleeren Kandidaten, allein sein
// Ergebnis entscheidet (true bei falschem Kandidaten 200, false bei richtigem 401), die Argumente sind Token und
// Kandidat Byte für Byte; ein Kandidat der Länge 1 ruft einmal, der leere nie.
func TestVergleichEntscheidetAllein(t *testing.T) {
	tok := string(tokenGueltig())
	faelle := []struct {
		name, wert string
		ergebnis   bool
		status     int
		aufrufe    int
		kandidat   string
	}{
		{"true bei falschem Kandidaten", "Bearer falsch", true, 200, 1, "falsch"},
		{"false bei richtigem Kandidaten", "Bearer " + tok, false, 401, 1, tok},
		{"true bei richtigem Kandidaten", "Bearer " + tok, true, 200, 1, tok},
		{"false bei falschem Kandidaten", "Bearer falsch", false, 401, 1, "falsch"},
		{"Kandidat der Länge 1", "Bearer x", true, 200, 1, "x"},
		{"leerer Kandidat", "Bearer   ", true, 401, 0, ""},
	}
	for _, f := range faelle {
		route := neueRoute()
		vf := neuerVergleich(f.ergebnis)
		u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		rec := u.rufe(ohneAuth("GET", "/a", map[string][]string{"Authorization": {f.wert}}))
		if rec.Code != f.status || vf.zahl() != f.aufrufe {
			t.Errorf("%s: Status %d, %d Aufrufe von Vergleich, erwartet %d und %d", f.name, rec.Code, vf.zahl(), f.status, f.aufrufe)
		}
		if args, ok := vf.einziger(); f.aufrufe == 1 && (!ok || args != [2]string{tok, f.kandidat}) {
			t.Errorf("%s: Argumente %q, erwartet (%q, %q)", f.name, args, tok, f.kandidat)
		}
	}
}

// u5Fall ist ein Paar von Werten für vergleiche.
type u5Fall struct {
	name string
	a, b []byte
}

// u5Faelle erzeugt die Menge von U5: sechs Formen × drei Längen und drei Randfälle.
func u5Faelle() []u5Fall {
	var faelle []u5Fall
	for _, n := range []int{32, 33, 64} {
		a := make([]byte, n)
		for i := range a {
			a[i] = byte(33 + i*5%90)
		}
		mit := func(name string, aendere func(b []byte) []byte) {
			faelle = append(faelle, u5Fall{name + " " + strconv.Itoa(n), a, aendere(slices.Clone(a))})
		}
		mit("gleich", func(b []byte) []byte { return b })
		mit("kürzer", func(b []byte) []byte { return b[:len(b)-1] })
		mit("länger", func(b []byte) []byte { return append(b, 'x') })
		for name, i := range map[string]int{"erste Stelle": 0, "mittlere Stelle": n / 2, "letzte Stelle": n - 1} {
			mit("abweichend an "+name, func(b []byte) []byte { b[i]++; return b })
		}
	}
	return append(faelle, u5Fall{"leer, leer", []byte{}, []byte{}}, u5Fall{"leer, nichtleer", []byte{}, []byte{'x'}},
		u5Fall{"nil, leer", nil, []byte{}})
}

// TestVergleiche hält U5 (M-vergleiche): vergleiche liefert true genau bei gleichen Werten; Orakel ist der Vergleich der
// Zeichenketten im Test, in beiden Reihenfolgen.
func TestVergleiche(t *testing.T) {
	faelle := u5Faelle()
	if len(faelle) != 21 {
		t.Fatalf("%d Fälle, erwartet 21", len(faelle))
	}
	for _, f := range faelle {
		soll := string(f.a) == string(f.b)
		if got := admin.Vergleiche(f.a, f.b); got != soll {
			t.Errorf("%s: vergleiche = %v, erwartet %v", f.name, got, soll)
		}
		if got := admin.Vergleiche(f.b, f.a); got != soll {
			t.Errorf("%s (vertauscht): vergleiche = %v, erwartet %v", f.name, got, soll)
		}
	}
}

// u7Methoden sind die acht Methoden von M-Methode × Pfad (U7).
func u7Methoden() []string {
	return []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}
}

// u7Pfade sind die fünf Pfade von M-Methode × Pfad (U7): registriert, unbekannt, Wurzel, `/__admin` und `*`.
func u7Pfade() []string { return []string{"/a", "/zzz", "/", "/__admin", "*"} }

// TestAuthentisierungJedeMethodeJederPfad hält U7 und U8: ohne Token ergibt jede Methode auf jedem Pfad 401 (8 × 5),
// auch HEAD, OPTIONS und `*`, und weder Route noch Use Case laufen.
func TestAuthentisierungJedeMethodeJederPfad(t *testing.T) {
	uc := &useCaseFake{}
	route := routeMitUseCase(uc)
	var routen []admin.Route
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		routen = append(routen, admin.Route{Methode: m, Muster: "/a", Handler: route.handle})
	}
	faelle := 0
	for _, m := range u7Methoden() {
		for _, p := range u7Pfade() {
			faelle++
			u := neuMit(t, routen...)
			r := ohneAuth(m, "/", nil)
			if p == "*" {
				r.URL.Path, r.URL.RawPath = p, ""
			} else {
				r = ohneAuth(m, p, nil)
			}
			pruefe401(t, m+" "+p, u.rufe(r))
		}
	}
	if faelle != 40 || route.aufrufe() != 0 || uc.summe() != 0 {
		t.Errorf("%d Fälle (erwartet 40), Route %d, Use Case %d Aufrufe (erwartet 0)", faelle, route.aufrufe(), uc.summe())
	}
}

// TestFalschesTokenOhneZustand hält U8 für Vergleich false: 401, Route und Use Case laufen nicht.
func TestFalschesTokenOhneZustand(t *testing.T) {
	uc := &useCaseFake{}
	route := routeMitUseCase(uc)
	vf := neuerVergleich(false)
	u := neuMitVergleich(t, nil, vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	pruefe401(t, "falsches Token", u.rufe(ohneAuth("GET", "/a", bearer("falsch"))))
	if vf.zahl() != 1 || route.aufrufe() != 0 || uc.summe() != 0 {
		t.Errorf("Vergleich %d, Route %d, Use Case %d Aufrufe, erwartet 1, 0 und 0", vf.zahl(), route.aufrufe(), uc.summe())
	}
}

// TestKeinGeheimnisInDerAntwort hält U11: weder das Token noch der Kandidat erreichen Body oder Header einer Antwort,
// auch nicht als Teilstück ab 8 Bytes; die 20 Formen mit dem Canary als Kandidat, Vergleich falsch, die Erfolgsantwort.
func TestKeinGeheimnisInDerAntwort(t *testing.T) {
	tok := string(tokenVielfalt())
	type fall struct {
		name, kandidat string
		kopf           map[string][]string
		ergebnis       bool
	}
	var faelle []fall
	for _, f := range u1Formen(canary()) {
		faelle = append(faelle, fall{f.name, canary(), f.kopf, true})
	}
	faelle = append(faelle, fall{"Vergleich falsch", canary(), bearer(canary()), false}, fall{"Erfolg", tok, bearer(tok), true})
	if len(faelle) != 22 {
		t.Fatalf("%d Fälle, erwartet 22", len(faelle))
	}
	for _, f := range faelle {
		vf := neuerVergleich(f.ergebnis)
		route := neueRoute()
		route.antwort = admin.Antwort{Status: 200, Body: []byte(routeBody)}
		u := neuMitVergleich(t, tokenVielfalt(), vf, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		text := textVon(u.rufe(ohneAuth("GET", "/a", f.kopf)))
		for _, geheim := range []string{tok, f.kandidat, canary()} {
			if enthaeltTeil(text, geheim) {
				t.Errorf("%s: die Antwort enthält %q oder ein Teilstück ab 8 Bytes: %s", f.name, geheim, text)
			}
		}
	}
}

// TestUnauthorized hält U9: das 401 trägt das Fehlerobjekt ODR-E2101 und WWW-Authenticate: Bearer, mit der Länge des
// Literals als fester Zahl.
func TestUnauthorized(t *testing.T) {
	if len(objekt401) != laenge401 {
		t.Fatalf("Literal von %d Bytes, erwartet %d", len(objekt401), laenge401)
	}
	u := neuMit(t)
	pruefe401(t, "ohne Token", u.rufe(ohneAuth("GET", "/a", nil)))
}
