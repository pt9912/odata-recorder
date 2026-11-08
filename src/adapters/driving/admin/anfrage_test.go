package admin_test

import (
	"context"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// ctxSchluessel ist der Schlüssel eines Wertes im ctx der Requests.
type ctxSchluessel struct{}

// z7Fall ist ein Request mit der Erwartung an ein Feld der Anfrage; die übrigen Felder prüft ein anderer Fall.
type z7Fall struct {
	name    string
	request func() *http.Request
	feld    string
	soll    any
}

// z7Faelle sind die acht Felder der Anfrage mit je eigener Eingabe: Pfad mit %2F, Query mit %-Folgen, Länge 0, 5 und unbekannt.
func z7Faelle() []z7Fall {
	post := func(ziel string, body io.Reader) func() *http.Request {
		return func() *http.Request { return anfrageMit("POST", ziel, nil, body) }
	}
	return []z7Fall{
		{"Methode", post("/a/x", nil), "Methode", "POST"},
		{"Pfad mit %2F", post("/a/x%2Fy", nil), "Pfad", "/a/x/y"},
		{"Pfadwerte mit %2F", post("/a/x%2Fy", nil), "Pfadwerte", map[string]string{"id": "x/y"}},
		{"RawQuery unverändert", post("/a/x?q=%zz&r=%41&s", nil), "RawQuery", "q=%zz&r=%41&s"},
		{"Header", func() *http.Request {
			return anfrageMit("POST", "/a/x", map[string][]string{"X-A": {"1", "2"}}, nil)
		}, "Header", map[string][]string{"X-A": {"1", "2"}}},
		{"Body", post("/a/x", strings.NewReader("hello")), "Body", "hello"},
		{"Länge 0", post("/a/x", nil), "Laenge", int64(0)},
		{"Länge 5", post("/a/x", strings.NewReader("hello")), "Laenge", int64(5)},
		{"Länge unbekannt", post("/a/x", iotest.OneByteReader(strings.NewReader("hello"))), "Laenge", int64(-1)},
	}
}

// TestAnfrageFelder hält Z7: die Anfrage trägt Methode, dekodierten Pfad, RawQuery unverändert, Pfadwerte, Header, Body
// und Laenge (-1 bei unbekannter Länge); die KorrelationsID hält Z7b.
func TestAnfrageFelder(t *testing.T) {
	felder := map[string]bool{}
	for _, f := range z7Faelle() {
		felder[f.feld] = true
		route := neueRoute()
		u := neuMit(t, admin.Route{Methode: "POST", Muster: "/a/{id}", Handler: route.handle})
		rec := u.rufe(f.request())
		a, ok := route.einzige()
		if !ok || rec.Code != 200 {
			t.Errorf("%s: Aufrufe %d, Status %d", f.name, route.aufrufe(), rec.Code)
			continue
		}
		var ist any
		switch f.feld {
		case "Methode":
			ist = a.Methode
		case "Pfad":
			ist = a.Pfad
		case "Pfadwerte":
			ist = a.Pfadwerte
		case "RawQuery":
			ist = a.RawQuery
		case "Header":
			ist = a.Header
		case "Body":
			b, _ := io.ReadAll(a.Body)
			ist = string(b)
		case "Laenge":
			ist = a.Laenge
		}
		if !gleichWert(ist, f.soll) {
			t.Errorf("%s: %s = %#v, erwartet %#v", f.name, f.feld, ist, f.soll)
		}
	}
	// acht Felder ohne die Korrelations-ID
	if len(felder) != 7 {
		t.Errorf("%d Felder geprüft, erwartet 7 (Methode, Pfad, Pfadwerte, RawQuery, Header, Body, Laenge)", len(felder))
	}
}

// gleichWert vergleicht die Werte der Fälle von z7Faelle.
func gleichWert(a, b any) bool {
	switch x := a.(type) {
	case map[string]string:
		y, ok := b.(map[string]string)
		return ok && maps.Equal(x, y)
	case map[string][]string:
		y, ok := b.(map[string][]string)
		return ok && maps.EqualFunc(x, y, slices.Equal[[]string])
	}
	return a == b
}

// TestPfadwerteNieNil hält Z7: Pfadwerte ist bei einer Route ohne Parameter leer, nie nil.
func TestPfadwerteNieNil(t *testing.T) {
	route := neueRoute()
	u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a/b", Handler: route.handle})
	u.rufe(anfrageMit("GET", "/a/b", nil, nil))
	a, ok := route.einzige()
	if !ok || a.Pfadwerte == nil || len(a.Pfadwerte) != 0 {
		t.Errorf("Pfadwerte = %#v (Aufruf %v), erwartet leer und nicht nil", a.Pfadwerte, ok)
	}
	if a.Header == nil {
		t.Error("Header ist nil")
	}
}

// TestKorrelationsID hält Z7: IDs.KorrelationsID läuft je Request genau einmal, vor allem anderen, und keine andere
// Operation von IDs; die Anfrage der Route trägt die vergebene ID. Requestarten: Route, 404, 405, GET /healthz.
func TestKorrelationsID(t *testing.T) {
	arten := []struct{ name, methode, ziel string }{
		{"Route", "GET", "/a"}, {"404", "GET", "/zzz"}, {"405", "POST", "/a"}, {"healthz", "GET", "/healthz"},
	}
	for _, art := range arten {
		route := neueRoute()
		u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		var vorRoute, vorAudit, vorTele int
		u.ids.vor = func() { vorRoute, vorAudit, vorTele = route.aufrufe(), u.audit.Aufrufe(), len(u.tele.Aufrufe()) }
		u.rufe(anfrageMit(art.methode, art.ziel, nil, nil))
		k, uebrige := u.ids.zahlen()
		if k != 1 || uebrige != 0 {
			t.Errorf("%s: KorrelationsID %d-mal, andere Operationen %d-mal, erwartet 1 und 0", art.name, k, uebrige)
		}
		if vorRoute != 0 || vorAudit != 0 || vorTele != 0 {
			t.Errorf("%s: beim Aufruf von KorrelationsID liefen Route %d, Audit %d, Telemetrie %d", art.name, vorRoute, vorAudit, vorTele)
		}
		if a, ok := route.einzige(); art.name == "Route" && (!ok || a.KorrelationsID != u.ids.letzte() || a.KorrelationsID == "") {
			t.Errorf("Route: KorrelationsID %q, vergeben %q", a.KorrelationsID, u.ids.letzte())
		}
	}
}

// TestCtxDerRoute hält Z7: die Route erhält den ctx des Requests.
func TestCtxDerRoute(t *testing.T) {
	route := neueRoute()
	u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	r := anfrageMit("GET", "/a", nil, nil)
	u.rufe(r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, "wert")))
	route.mu.Lock()
	defer route.mu.Unlock()
	if len(route.ctxs) != 1 || route.ctxs[0].Value(ctxSchluessel{}) != "wert" {
		t.Errorf("die Route erhielt nicht den ctx des Requests (%d Aufrufe)", len(route.ctxs))
	}
}

// z8Fall ist ein Satz von Headern mit den Namen, die die Route nicht sehen darf.
type z8Fall struct {
	name string
	kopf map[string][]string
}

// z8Nachbarn sind die Namen, die Authorization ähneln und bleiben: jedes Zeichen gestrichen (13), nach jedem Zeichen und
// davor ein `x` eingefügt (14), jedes Zeichen durch `x` ersetzt (13) und acht feste Namen (Präfix, Suffix und verwandte
// Felder); 48 verschiedene, keiner gleicht Authorization in einer Schreibweise.
func z8Nachbarn() []string {
	const a = "authorization"
	var namen []string
	for i := 0; i <= len(a); i++ {
		namen = append(namen, a[:i]+"x"+a[i:])
	}
	for i := 0; i < len(a); i++ {
		namen = append(namen, a[:i]+a[i+1:], a[:i]+"x"+a[i+1:])
	}
	return append(namen, "Proxy-Authorization", "X-Authorization", "Authorization-Info", "Auth", "AuthX", "Auth-Info",
		"Authentication", "WWW-Authenticate")
}

// z8Schreibweisen sind alle 2^13 Groß-/Kleinschreibungen von Authorization.
func z8Schreibweisen() []string {
	const a = "authorization"
	var namen []string
	for m := 0; m < 1<<len(a); m++ {
		b := []byte(a)
		for i := range b {
			if m>>i&1 == 1 {
				b[i] -= 'a' - 'A'
			}
		}
		namen = append(namen, string(b))
	}
	return namen
}

// z8Faelle sind die Header-Zeilen von Z8 neben der gültigen kanonischen Zeile, die anfrageMit setzt: ein Schlüssel in
// jeder der 8192 Schreibweisen von Authorization trägt einen Canary (die kanonische Schreibweise ist die gültige Zeile
// selbst und trägt keinen), jeder Nachbarname steht neben ihr: nur Authorization fällt weg.
func z8Faelle() []z8Fall {
	bearer := []string{"Bearer " + canary()}
	faelle := []z8Fall{
		{"Authorization", map[string][]string{}},
		{"X-A", map[string][]string{"X-A": {"1"}}},
		{"zwei Werte", map[string][]string{"X-Multi": {"1", "2"}}},
	}
	for _, n := range z8Schreibweisen() {
		kopf := map[string][]string{n: bearer}
		if n == "Authorization" {
			kopf = map[string][]string{}
		}
		faelle = append(faelle, z8Fall{"Schreibweise " + n, kopf})
	}
	for _, n := range z8Nachbarn() {
		faelle = append(faelle, z8Fall{"Nachbar " + n, map[string][]string{n: {"1"}}})
	}
	return faelle
}

// TestHeaderOhneAuthorization hält Z8: Authorization erreicht keine Route, jede andere Zeile bleibt.
func TestHeaderOhneAuthorization(t *testing.T) {
	faelle := z8Faelle()
	verschieden := map[string]bool{}
	for _, n := range z8Nachbarn() {
		verschieden[n] = true
		if strings.EqualFold(n, "Authorization") {
			t.Errorf("Nachbarname %q gleicht Authorization", n)
		}
	}
	if len(verschieden) != 48 || len(z8Nachbarn()) != 48 || len(z8Schreibweisen()) != 8192 || len(faelle) != 3+8192+48 {
		t.Fatalf("%d Nachbarn (%d verschieden), %d Schreibweisen, %d Fälle, erwartet 48, 8192 und 3 + 8192 + 48",
			len(z8Nachbarn()), len(verschieden), len(z8Schreibweisen()), len(faelle))
	}
	for _, f := range faelle {
		route := neueRoute()
		u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
		u.rufe(anfrageMit("GET", "/a", f.kopf, nil))
		a, ok := route.einzige()
		if !ok {
			t.Fatalf("%s: die Route lief nicht", f.name)
		}
		soll := map[string][]string{}
		for k, v := range f.kopf {
			if !strings.EqualFold(k, "Authorization") {
				soll[k] = v
			}
		}
		if !gleichWert(a.Header, soll) {
			t.Errorf("%s: Header %v, erwartet %v", f.name, a.Header, soll)
		}
		for k, v := range a.Header {
			if strings.EqualFold(k, "Authorization") || strings.Contains(strings.Join(v, ","), canary()) {
				t.Errorf("%s: der Header %s trägt das Token", f.name, k)
			}
		}
	}
}

// TestHeaderKopie hält Z8: Header ist eine Kopie; überschreibt die Route einen Wert und löscht sie einen Namen, bleiben
// die Header des Requests gleich.
func TestHeaderKopie(t *testing.T) {
	route := neueRoute()
	route.aendere = func(a admin.Anfrage) {
		a.Header["X-A"][0] = "geändert"
		a.Header["X-Multi"][1] = "geändert"
		delete(a.Header, "X-Multi")
	}
	u := neuMit(t, admin.Route{Methode: "GET", Muster: "/a", Handler: route.handle})
	r := anfrageMit("GET", "/a", map[string][]string{"X-A": {"1"}, "X-Multi": {"1", "2"}}, nil)
	vorher := map[string][]string{}
	for k, v := range r.Header {
		vorher[k] = slices.Clone(v)
	}
	u.rufe(r)
	if route.aufrufe() != 1 || !gleichWert(map[string][]string(r.Header), vorher) {
		t.Errorf("Header des Requests %v, erwartet %v (Aufrufe %d)", r.Header, vorher, route.aufrufe())
	}
}
