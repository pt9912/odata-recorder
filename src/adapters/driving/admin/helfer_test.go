package admin_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Die Fehlerobjekte als Literale des Vertrags (SPEC-109 Nr. 1 bis 3, SPEC-240), nicht aus dem Code. objekt404 und
// objekt400 sind dasselbe Fehlerobjekt von ODR-E2102 unter dem Namen des Status, mit dem ein Test es erwartet.
const (
	objekt404   = `{"error":{"code":"ODR-E2102","message":"Admin request invalid","target":"admin"}}`
	objekt400   = objekt404
	objekt500   = `{"error":{"code":"ODR-E9001","message":"Internal error","target":"internal"}}`
	objekt503   = `{"error":{"code":"ODR-E5010","message":"Service is shutting down","target":"service"}}`
	objektE7003 = `{"error":{"code":"ODR-E7003","message":"Recording session unknown","target":"session"}}`
	healthzObj  = `{"status":"ok"}`
)

// umgebung ist ein Server mit allen Fakes seiner Ports.
type umgebung struct {
	srv   *admin.Server
	ids   *zaehlendeIDs
	audit *supporttest.FakeAudit
	tele  *supporttest.FakeTelemetry
	lst   *listenerFake
}

// canary liefert einen Wert, der in keiner Ausgabe stehen darf; er entsteht aus Bytewerten und besteht nur aus
// Zeichen, die ein Literal eines Musters tragen darf.
func canary() string {
	return string([]byte{0x5a, 0x71, 0x37, 0x4b, 0x2d, 0x63, 0x61, 0x6e, 0x61, 0x72, 0x79, 0x2d, 0x78, 0x39, 0x33})
}

// tokenGueltig liefert ein Token von 32 Bytes.
func tokenGueltig() []byte { return bytes.Repeat([]byte{'t'}, 32) }

// konfigMit liefert eine gültige Konfiguration mit frischen Fakes.
func konfigMit(routen ...admin.Route) (admin.Konfig, *umgebung) {
	u := &umgebung{ids: neueIDs(), audit: supporttest.NewFakeAudit(), tele: supporttest.NewFakeTelemetry(), lst: neuerListener()}
	return admin.Konfig{
		Listener: u.lst, Token: tokenGueltig(), IDs: u.ids, Audit: u.audit, Telemetrie: u.tele, Routen: routen,
	}, u
}

// neuMit baut einen Server aus gültiger Konfiguration; ein Fehler von New bricht den Test ab.
func neuMit(t *testing.T, routen ...admin.Route) *umgebung {
	t.Helper()
	k, u := konfigMit(routen...)
	srv, err := admin.New(k)
	if err != nil || srv == nil {
		t.Fatalf("New: %v (Server %v)", err, srv)
	}
	u.srv = srv
	return u
}

// anfrageMit bildet einen Request; ziel steht unverändert in der Anfragezeile (kein Bereinigen), kopf setzt Header
// unter genau dem angegebenen Schlüssel. Jeder Request trägt einen Bearer-Header, den der Adapter nicht liest.
func anfrageMit(methode, ziel string, kopf map[string][]string, body io.Reader) *http.Request {
	r := httptest.NewRequest(methode, ziel, body)
	r.Header.Set("Authorization", "Bearer "+strings.Repeat("t", 32))
	for k, v := range kopf {
		r.Header[k] = v
	}
	return r
}

// rufe sendet den Request und liefert die Antwort.
func (u *umgebung) rufe(r *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	u.srv.ServeHTTP(rec, r)
	return rec
}

// antwortText liefert Status, Header und Body der Antwort als Text, um darin nach einem Wert zu suchen.
func antwortText(rec *httptest.ResponseRecorder) string {
	return fmt.Sprintf("%d %v %s", rec.Code, rec.Header(), rec.Body.Bytes())
}

// pruefeFehlerAntwort prüft Status, Body, Content-Type, Content-Length und die Header des Status einer Fehlerantwort
// (SPEC-238 Nr. 5): WWW-Authenticate: Bearer nur bei 401, Connection: close nur bei 503, Allow nur bei 405; allow leer
// heißt: kein Header Allow. Der Canary steht in keinem Byte der Antwort.
func pruefeFehlerAntwort(t *testing.T, name string, rec *httptest.ResponseRecorder, status int, body, allow string) {
	t.Helper()
	if rec.Code != status || rec.Body.String() != body {
		t.Errorf("%s: Antwort %d %q, erwartet %d %q", name, rec.Code, rec.Body.String(), status, body)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("%s: Content-Type %q", name, got)
	}
	if got := rec.Header().Get("Content-Length"); got != strconv.Itoa(len(body)) {
		t.Errorf("%s: Content-Length %q, erwartet %d", name, got, len(body))
	}
	if got := rec.Header().Values("Allow"); allow == "" && len(got) != 0 || allow != "" && (len(got) != 1 || got[0] != allow) {
		t.Errorf("%s: Allow %q, erwartet %q", name, got, allow)
	}
	soll := map[string][]string{"WWW-Authenticate": nil, "Connection": nil}
	if status == 401 {
		soll["WWW-Authenticate"] = []string{"Bearer"}
	}
	if status == 503 {
		soll["Connection"] = []string{"close"}
	}
	for k, v := range soll {
		if got := rec.Header().Values(k); !slices.Equal(got, v) {
			t.Errorf("%s: %s %q, erwartet %q", name, k, got, v)
		}
	}
	if strings.Contains(antwortText(rec), canary()) {
		t.Errorf("%s: der Canary steht in der Antwort", name)
	}
}

// keinePortAufrufe meldet, wenn Audit, Telemetrie oder Listener einen Aufruf sahen.
func (u *umgebung) keinePortAufrufe(t *testing.T, name string) {
	t.Helper()
	u.nurZaehlung(t, name)
}

// zaehlungenVon liefert die Codes der Aufrufe von CountError in der Reihenfolge und die Zahl der Aufrufe jeder anderen
// Operation (Log, StartSpan, Set, End, Snapshot).
func zaehlungenVon(f *supporttest.FakeTelemetry) (codes []string, andere int) {
	for _, a := range f.Aufrufe() {
		if a.Op == supporttest.OpCountError {
			codes = append(codes, a.Code.String())
		} else {
			andere++
		}
	}
	return codes, andere
}

// codeAbgewiesen ist der Code, den allein die Antwort 401 trägt (SPEC-109 Nr. 2): jede Zählung dieses Codes ist ein
// Abweisen mit genau einem Eintrag im Audit (SPEC-238 Nr. 8).
const codeAbgewiesen = "ODR-E2101"

// nurZaehlung meldet, wenn die Telemetrie etwas anderes sah als je einen Aufruf von CountError mit den Codes soll (in
// dieser Reihenfolge), wenn Audit nicht genau so viele Aufrufe sah, wie soll den Code ODR-E2101 der Antwort 401 nennt
// (bei jedem anderen Code 0), oder wenn der Listener einen Aufruf sah.
func (u *umgebung) nurZaehlung(t *testing.T, name string, soll ...string) {
	t.Helper()
	codes, andere := zaehlungenVon(u.tele)
	a, c, ad := u.lst.zahlen()
	abweisungen := 0
	for _, code := range soll {
		if code == codeAbgewiesen {
			abweisungen++
		}
	}
	if !slices.Equal(codes, soll) || andere != 0 || u.audit.Aufrufe() != abweisungen || a+c+ad != 0 {
		t.Errorf("%s: CountError %q, andere Aufrufe der Telemetrie %d, Audit %d, Listener %d/%d/%d Aufrufe, erwartet CountError %q, Audit %d und sonst 0",
			name, codes, andere, u.audit.Aufrufe(), a, c, ad, soll, abweisungen)
	}
}

// rufeSicher sendet den Request wie rufe und meldet eine Panik, die ServeHTTP verlässt, als Fehler des Tests.
func (u *umgebung) rufeSicher(t *testing.T, name string, r *http.Request) (rec *httptest.ResponseRecorder) {
	t.Helper()
	rec = httptest.NewRecorder()
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("%s: eine Panik verließ ServeHTTP: %v", name, p)
		}
	}()
	u.srv.ServeHTTP(rec, r)
	return rec
}
