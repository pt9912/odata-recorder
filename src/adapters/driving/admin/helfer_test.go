package admin_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Die Fehlerobjekte als Literale des Vertrags (SPEC-109 Nr. 1 bis 3, SPEC-240), nicht aus dem Code.
const (
	objekt404  = `{"error":{"code":"ODR-E2102","message":"Admin request invalid","target":"admin"}}`
	objekt500  = `{"error":{"code":"ODR-E9001","message":"Internal error","target":"internal"}}`
	healthzObj = `{"status":"ok"}`
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

// pruefeFehlerAntwort prüft Status, Body, Content-Type, Content-Length und Allow einer Fehlerantwort; allow leer heißt:
// kein Header Allow. Der Canary steht in keinem Byte der Antwort.
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
	if strings.Contains(antwortText(rec), canary()) {
		t.Errorf("%s: der Canary steht in der Antwort", name)
	}
}

// keinePortAufrufe meldet, wenn Audit, Telemetrie oder Listener einen Aufruf sahen.
func (u *umgebung) keinePortAufrufe(t *testing.T, name string) {
	t.Helper()
	a, c, ad := u.lst.zahlen()
	if u.audit.Aufrufe() != 0 || len(u.tele.Aufrufe()) != 0 || a+c+ad != 0 {
		t.Errorf("%s: Audit %d, Telemetrie %d, Listener %d/%d/%d Aufrufe, erwartet 0", name, u.audit.Aufrufe(), len(u.tele.Aufrufe()), a, c, ad)
	}
}
