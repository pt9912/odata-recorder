package admin_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
)

// Das Fehlerobjekt zu ODR-E2101 als Literal des Vertrags (SPEC-109 Nr. 1, 2, SPEC-238 Nr. 5) und seine Länge in Bytes.
const (
	objekt401  = `{"error":{"code":"ODR-E2101","message":"Admin authentication missing or invalid","target":"admin"}}`
	laenge401  = 99
	laengeHZ   = 15
	routeBody  = `{"a":1}`
	wwwBearer  = "Bearer"
	allowNurGE = "GET"
)

// vergleichFake ist Konfig.Vergleich des Tests: er zählt, hält beide Argumente jedes Aufrufs als Kopie und liefert das
// eingestellte Ergebnis, bei echt das der Funktion des Adapters.
type vergleichFake struct {
	mu       sync.Mutex
	aufrufe  [][2]string
	ergebnis bool
	echt     bool
}

func neuerVergleich(ergebnis bool) *vergleichFake { return &vergleichFake{ergebnis: ergebnis} }

func (v *vergleichFake) vergleich(erwartet, gegeben []byte) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.aufrufe = append(v.aufrufe, [2]string{string(erwartet), string(gegeben)})
	if v.echt {
		return admin.Vergleiche(erwartet, gegeben)
	}
	return v.ergebnis
}

func (v *vergleichFake) zahl() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return len(v.aufrufe)
}

// einziger liefert die Argumente des einen Aufrufs; ok ist falsch, wenn Vergleich nicht genau einmal lief.
func (v *vergleichFake) einziger() (args [2]string, ok bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if len(v.aufrufe) != 1 {
		return args, false
	}
	return v.aufrufe[0], true
}

// useCaseFake steht hinter der Testroute und zählt jede Operation, Lesen wie Schreiben.
type useCaseFake struct {
	mu                sync.Mutex
	lesen, schreibend int
}

func (f *useCaseFake) lies()     { f.mu.Lock(); f.lesen++; f.mu.Unlock() }
func (f *useCaseFake) schreibe() { f.mu.Lock(); f.schreibend++; f.mu.Unlock() }
func (f *useCaseFake) summe() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lesen + f.schreibend
}

// routeMitUseCase liefert einen Route-Fake, dessen Aufruf beide Operationen des Use-Case-Fakes ruft.
func routeMitUseCase(uc *useCaseFake) *routeFake {
	r := neueRoute()
	r.aendere = func(admin.Anfrage) { uc.lies(); uc.schreibe() }
	return r
}

// tokenVielfalt liefert ein Token von 36 Bytes aus Bytewerten mit Großbuchstaben, Ziffern, innerem Leerzeichen und
// Satzzeichen.
func tokenVielfalt() []byte {
	return []byte{
		'A', 'b', 'C', '1', '2', '3', ' ', 'X', 'y', 'Z', '-', '_', '.', '~', 'Q', 'w', 'E', 'r', 'T',
		'5', '6', '7', '8', '9', '0', 'p', 'O', 'i', 'U', 'y', 'T', 'R', 'e', 'W', 'q', 'L',
	}
}

// bom liefert die drei Bytes eines BOM aus Bytewerten.
func bom() []byte { return []byte{0xEF, 0xBB, 0xBF} }

// neuMitKonfig baut den Server aus k; ein Fehler von New bricht den Test ab.
func neuMitKonfig(t *testing.T, k admin.Konfig, u *umgebung) *umgebung {
	t.Helper()
	srv, err := admin.New(k)
	if err != nil || srv == nil {
		t.Fatalf("New: %v (Server %v)", err, srv)
	}
	u.srv = srv
	return u
}

// neuMitVergleich baut einen Server mit gültiger Konfiguration, dem Vergleich vf (nil: Vorgabe) und dem Token.
func neuMitVergleich(t *testing.T, token []byte, vf *vergleichFake, routen ...admin.Route) *umgebung {
	t.Helper()
	k, u := konfigMit(routen...)
	if token != nil {
		k.Token = token
	}
	if vf != nil {
		k.Vergleich = vf.vergleich
	}
	return neuMitKonfig(t, k, u)
}

// ohneAuth bildet einen Request ohne Authorization; kopf setzt Header unter genau dem angegebenen Schlüssel.
func ohneAuth(methode, ziel string, kopf map[string][]string) *http.Request {
	r := httptest.NewRequest(methode, ziel, nil)
	for k, v := range kopf {
		r.Header[k] = v
	}
	return r
}

// bearer liefert den Header einer gültigen Zeile mit dem Kandidaten.
func bearer(kandidat string) map[string][]string {
	return map[string][]string{"Authorization": {"Bearer " + kandidat}}
}

// kopfWerte liefert die Werte des Headers name der Antwort, wie der Server sie gesendet hätte (Stand bei WriteHeader).
func kopfWerte(rec *httptest.ResponseRecorder, name string) []string {
	return rec.Result().Header.Values(name)
}

// textVon liefert Status, Header und Body der Antwort als Text, um darin nach einem Wert zu suchen.
func textVon(rec *httptest.ResponseRecorder) string {
	res := rec.Result()
	return fmt.Sprintf("%d %v %s", res.StatusCode, res.Header, rec.Body.Bytes())
}

// enthaeltTeil meldet, ob text den Wert oder ein Teilstück ab 8 Bytes enthält (jedes längere Teilstück enthält ein
// Fenster von 8 Bytes).
func enthaeltTeil(text, wert string) bool {
	if len(wert) < 8 {
		return strings.Contains(text, wert)
	}
	for i := 0; i+8 <= len(wert); i++ {
		if strings.Contains(text, wert[i:i+8]) {
			return true
		}
	}
	return false
}

// pruefe401 hält die Antwort gegen das 401 des Vertrags: Status, Body, Content-Type, Content-Length und
// WWW-Authenticate als Literale, kein Allow und kein Connection.
func pruefe401(t *testing.T, name string, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusUnauthorized || rec.Body.String() != objekt401 {
		t.Errorf("%s: Antwort %d %q, erwartet 401 %q", name, rec.Code, rec.Body.String(), objekt401)
	}
	soll := map[string][]string{
		"Content-Type": {"application/json"}, "Content-Length": {strconv.Itoa(laenge401)},
		"Www-Authenticate": {wwwBearer}, "Allow": nil, "Connection": nil,
	}
	for k, v := range soll {
		if got := kopfWerte(rec, k); !slices.Equal(got, v) {
			t.Errorf("%s: %s %q, erwartet %q", name, k, got, v)
		}
	}
}

// artSchema liefert alle 2 hoch 6 Schreibweisen des Schemas Bearer aus den Buchstaben b, e, a, r, e, r.
func artSchema() []string {
	const s = "bearer"
	var out []string
	for m := 0; m < 1<<len(s); m++ {
		b := []byte(s)
		for i := range b {
			if m>>i&1 == 1 {
				b[i] -= 'a' - 'A'
			}
		}
		out = append(out, string(b))
	}
	return out
}
