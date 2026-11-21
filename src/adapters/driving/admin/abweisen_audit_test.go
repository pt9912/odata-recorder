package admin_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Dieser Teil hält die Zusagen von SPEC-238 Nr. 8 (Audit des Abweisens): bei jeder Antwort 401 genau ein Eintrag der Art
// admin_abgewiesen mit der Korrelations-ID der Antwort, bei keiner anderen Antwort einer (D1, D2), ohne Wert des Requests
// (D3), vor dem Schreiben der Antwort (D4), ohne Einfluss des Ergebnisses von Append (D5), auf der Goroutine des Requests
// (D6) und ohne Drosselung (D7). Das Orakel trägt Art, Code, Status und Header als Literale des Vertrags, nie als
// Konstanten des Codes.

// textAdminAbgewiesen ist der Text der fünften Art (SPEC-234 Nr. 2, SPEC-238 Nr. 8).
const textAdminAbgewiesen = "admin_abgewiesen"

// eintragSoll ist der eine Eintrag zu einer Antwort 401 mit der Korrelations-ID id: Art und Code aus dem Wortlaut des
// Vertrags, alle übrigen Felder leer.
func eintragSoll(t *testing.T, id string) support.Eintrag {
	t.Helper()
	code, ok := model.ParseCode("ODR-E2101")
	if !ok {
		t.Fatal("ODR-E2101 ist kein Code des Modells")
	}
	return support.Eintrag{Art: support.AuditArt(textAdminAbgewiesen), ErrorCode: code, KorrelationsID: id}
}

// quelle401 ist eine Form, in der ein Request mit 401 endet: der Server und der Request, den die Form auslöst.
type quelle401 struct {
	name string
	bau  func(t *testing.T) (*umgebung, *http.Request)
}

// u7Request bildet den Request von M-Methode × Pfad (U7); der Pfad `*` steht unverändert in der Anfragezeile.
func u7Request(methode, pfad string, kopf map[string][]string) *http.Request {
	if pfad == "*" {
		r := ohneAuth(methode, "/", kopf)
		r.URL.Path, r.URL.RawPath = pfad, ""
		return r
	}
	return ohneAuth(methode, pfad, kopf)
}

// fehlerE2101 ist der gültige *inbound.Fehler der Route mit ODR-E2101 und 401 (SPEC-238 Nr. 6 a 2).
func fehlerE2101(ursache error) error {
	return &inbound.Fehler{Code: model.CodeE2101, Status: http.StatusUnauthorized, Ursache: ursache}
}

// quellen401 ist die Menge der Formen, in denen eine Antwort 401 entsteht: die 20 Formen von U1, die 40 Zellen von U7
// ohne Token und mit falschem Token (Vergleich gleich false), und der Fehler der Route mit ODR-E2101 und 401 in 5
// Verpackungen, ohne und mit Ursache (10).
func quellen401() []quelle401 {
	var out []quelle401
	for _, f := range u1Formen(string(tokenGueltig())) {
		out = append(out, quelle401{"U1 " + f.name, func(t *testing.T) (*umgebung, *http.Request) {
			return neuMit(t), ohneAuth("GET", "/a", f.kopf)
		}})
	}
	for _, m := range u7Methoden() {
		for _, p := range u7Pfade() {
			out = append(out,
				quelle401{"U7 ohne Token " + m + " " + p, func(t *testing.T) (*umgebung, *http.Request) {
					return neuMit(t), u7Request(m, p, nil)
				}},
				quelle401{"U7 falsches Token " + m + " " + p, func(t *testing.T) (*umgebung, *http.Request) {
					return neuMitVergleich(t, nil, neuerVergleich(false)), u7Request(m, p, bearer("falsch"))
				}})
		}
	}
	for _, v := range verpackungen() {
		for _, ursache := range []error{nil, errors.New(canary())} {
			out = append(out, quelle401{fmt.Sprintf("Fehler der Route %s, Ursache %v", v.name, ursache != nil), func(t *testing.T) (*umgebung, *http.Request) {
				u, route := einRoute(t)
				route.err = v.packe(fehlerE2101(ursache))
				return u, anfrageMit("GET", "/r", nil, nil)
			}})
		}
	}
	return out
}

// sende baut den Server der Quelle, stellt vor und fehler am FakeAudit ein und sendet den Request.
func (q quelle401) sende(t *testing.T, vor func(support.Eintrag), fehler error) (*umgebung, *httptest.ResponseRecorder) {
	t.Helper()
	u, r := q.bau(t)
	u.audit.Vor, u.audit.Fehler = vor, fehler
	return u, u.rufeSicher(t, q.name, r)
}

// TestQuellen401Zahl hält die Menge der Formen der Antwort 401 gegen die Quelle des Tests: 20 + 40 + 40 + 10, jede mit
// eigenem Namen.
func TestQuellen401Zahl(t *testing.T) {
	namen := map[string]bool{}
	qs := quellen401()
	for _, q := range qs {
		namen[q.name] = true
	}
	gleichAuf(t, "Zahl der Formen", len(qs), 20+len(u7Methoden())*len(u7Pfade())*2+10)
	gleichAuf(t, "Zahl der Namen", len(namen), len(qs))
	gleichAuf(t, "Zahl der Zellen von U7", len(u7Methoden())*len(u7Pfade()), 40)
}

// TestAuditBeiJederAntwort401 hält D1 (Nr. 8): bei jeder Antwort 401 genau ein Append mit der Art admin_abgewiesen, dem
// Code ODR-E2101 und der Korrelations-ID der Antwort, alle übrigen Felder leer; Schleife über die 110 Formen.
func TestAuditBeiJederAntwort401(t *testing.T) {
	qs := quellen401()
	for _, q := range qs {
		u, rec := q.sende(t, nil, nil)
		pruefe401(t, q.name, rec)
		if n := u.audit.Aufrufe(); n != 1 {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet 1", q.name, n)
			continue
		}
		id := rec.Header().Get("X-Correlation-Id")
		if id == "" || id != u.ids.letzte() {
			t.Errorf("%s: X-Correlation-Id %q, vergeben %q", q.name, id, u.ids.letzte())
		}
		if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{eintragSoll(t, id)}) {
			t.Errorf("%s: Einträge %+v, erwartet genau %+v", q.name, got, eintragSoll(t, id))
		}
	}
	gleichAuf(t, "Zahl der Formen", len(qs), 110)
}

// TestAuditUeberEchtenListener hält D1 und D5 über eine echte Verbindung (Server von net/http auf 127.0.0.1): ein Client ohne
// Token, mit falschem Token und mit falschem Token bei einem Fehler von Append erhält 401 mit dem Literal, der Eintrag
// trägt die ID aus dem Header der Antwort, und der Fehler von Append ändert die Antwort nicht.
func TestAuditUeberEchtenListener(t *testing.T) {
	faelle := []struct {
		name, auth string
		fehler     error
	}{
		{"ohne Token", "", nil}, {"falsches Token", "Bearer falsch", nil}, {"falsches Token, Fehler von Append", "Bearer falsch", model.ErrUngueltig},
	}
	for _, f := range faelle {
		u := neuLive(t, neueRoute())
		u.audit.Fehler = f.fehler
		req, err := http.NewRequest("GET", "http://"+u.addr+"/p", nil)
		if err != nil {
			t.Fatal(err)
		}
		if f.auth != "" {
			req.Header.Set("Authorization", f.auth)
		}
		resp, err := u.client().Do(req)
		if err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		id := resp.Header.Get("X-Correlation-Id")
		if resp.StatusCode != 401 || string(body) != objekt401 || resp.Header.Get("WWW-Authenticate") != wwwBearer || id == "" {
			t.Errorf("%s: Antwort %d %q, WWW-Authenticate %q, X-Correlation-Id %q", f.name, resp.StatusCode, body, resp.Header.Get("WWW-Authenticate"), id)
		}
		if n := u.audit.Aufrufe(); n != 1 {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet 1", f.name, n)
			continue
		}
		if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{eintragSoll(t, id)}) {
			t.Errorf("%s: Einträge %+v, erwartet genau %+v", f.name, got, eintragSoll(t, id))
		}
	}
}

// TestAuditBeiBeendetemKontext hält die Zelle *Abbruch* des Rasters (Nr. 8): `Append` hat keinen `ctx` und läuft zu Ende, der
// Eintrag steht, auch wenn der Kontext des Requests beendet ist (der Client trennte): Zustand des Kontexts (abgebrochen,
// abgelaufen) × die 110 Formen; Antwort und Eintrag sind die des aktiven Kontexts.
func TestAuditBeiBeendetemKontext(t *testing.T) {
	zustaende := map[string]func() (context.Context, context.CancelFunc){
		"abgebrochen": func() (context.Context, context.CancelFunc) {
			ctx, abbruch := context.WithCancel(context.Background())
			abbruch()
			return ctx, abbruch
		},
		"abgelaufen": func() (context.Context, context.CancelFunc) {
			return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		},
	}
	faelle := 0
	for name, neuerKontext := range zustaende {
		for _, q := range quellen401() {
			faelle++
			u, r := q.bau(t)
			ctx, abbruch := neuerKontext()
			rec := u.rufeSicher(t, q.name, r.WithContext(ctx))
			abbruch()
			pruefe401(t, name+", "+q.name, rec)
			if n := u.audit.Aufrufe(); n != 1 {
				t.Errorf("%s, %s: %d Aufrufe von Audit, erwartet 1", name, q.name, n)
				continue
			}
			id := rec.Header().Get("X-Correlation-Id")
			if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{eintragSoll(t, id)}) {
				t.Errorf("%s, %s: Einträge %+v, erwartet genau %+v", name, q.name, got, eintragSoll(t, id))
			}
		}
	}
	gleichAuf(t, "Zahl der Fälle", faelle, 2*110)
}

// TestAppendEineAufrufstelle hält die Aufrufstelle von D1 (Nr. 8): im Paket ohne _test.go ruft und nennt genau eine Stelle
// `Append`, die eine Schreibstelle des Fehlerobjekts (wie C3 für `CountError`); der Zähler ist der von C3, die Probe
// dort hält ihn gegen Namensgleichheit und Importformen.
func TestAppendEineAufrufstelle(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	dateien, stellen := 0, 0
	for _, e := range eintraege {
		if n := e.Name(); strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") {
			src, err := os.ReadFile(n)
			if err != nil {
				t.Fatal(err)
			}
			dateien++
			stellen += zaehleSelektoren(t, n, src, "", "Append")
		}
	}
	if dateien < 4 || stellen != 1 {
		t.Errorf("%d Dateien ohne _test.go: %d Stellen von Append, erwartet mindestens 4 Dateien und 1 Stelle", dateien, stellen)
	}
}

// TestEintragHatSiebenFelder hält die Voraussetzung von D1 und D3: der Eintrag hat sieben Felder, und der Vergleich mit dem
// Literal prüft jedes; wächst der Eintrag, ist das Orakel zu erweitern.
func TestEintragHatSiebenFelder(t *testing.T) {
	gleichAuf(t, "Zahl der Felder von Eintrag", reflect.TypeOf(support.Eintrag{}).NumField(), 7)
}

// TestKeinAuditBeiAndererAntwort hält D2 (Nr. 8): bei keiner anderen Antwort ein Append, und nie eine der vier Arten des
// Use Case. Mengen: die 30 Requestarten R (genau die mit Status 401 schreiben einen Eintrag), M-Code × Status (jeder Code
// der Tabelle und drei ungültige mit jedem Status, den die Tabelle trägt, und sechs Status außerhalb: nur ODR-E2101 mit
// 401 schreibt), ein Kontextfehler neben dem Fehler ODR-E2101 und 401 oder als dessen Ursache (2 Fehler × 7 Formen × 2
// Stellen), die 20 Formen von U1 bei angehaltenem Server (503).
func TestKeinAuditBeiAndererAntwort(t *testing.T) {
	for _, a := range requestArten() {
		l := baue(t, a)
		rec := l.u.rufe(ohneAuth(a.methode, a.ziel, a.kopf))
		want := 0
		if a.status == 401 {
			want = 1
		}
		if rec.Code != a.status || l.u.audit.Aufrufe() != want {
			t.Errorf("%s: Status %d, %d Aufrufe von Audit, erwartet %d und %d", a.name, rec.Code, l.u.audit.Aufrufe(), a.status, want)
		}
	}
	stati := slices.Concat([]int{0, 100, 200, 399, 600, 999}, alleStati())
	codes := make([]model.Code, 0, len(specCodes())+3)
	namen := make([]string, 0, cap(codes))
	for i, z := range specCodes() {
		codes, namen = append(codes, modellCode(i)), append(namen, z.code)
	}
	codes, namen = append(codes, 0, -1, model.Code(len(specCodes())+1)), append(namen, "Code 0", "Code -1", "Code nach dem letzten")
	paare, mitEintrag := 0, 0
	for i, c := range codes {
		for _, s := range stati {
			paare++
			u, route := einRoute(t)
			route.err = &inbound.Fehler{Code: c, Status: s}
			u.rufeSicher(t, namen[i], anfrageMit("GET", "/r", nil, nil))
			want := 0
			if namen[i] == codeAbgewiesen && s == 401 {
				want, mitEintrag = 1, mitEintrag+1
			}
			if n := u.audit.Aufrufe(); n != want {
				t.Errorf("%s mit Status %d: %d Aufrufe von Audit, erwartet %d", namen[i], s, n, want)
			}
		}
	}
	gleichAuf(t, "Zahl der Paare", paare, len(codes)*len(stati))
	gleichAuf(t, "Zahl der Paare mit Eintrag", mitEintrag, 1)
	faelle := 0
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, f := range kontextFormen() {
			for name, err := range map[string]error{
				"neben dem Fehler": errors.Join(fehlerE2101(nil), f.packe(e)),
				"als Ursache":      fehlerE2101(f.packe(e)),
			} {
				faelle++
				u, route := einRoute(t)
				route.err = err
				var wert any
				func() {
					defer func() { wert = recover() }()
					u.rufe(anfrageMit("GET", "/r", nil, nil))
				}()
				if wert != http.ErrAbortHandler || u.audit.Aufrufe() != 0 {
					t.Errorf("%v, %s, %s: Panik %v, %d Aufrufe von Audit, erwartet den Abbruch und 0", e, f.name, name, wert, u.audit.Aufrufe())
				}
			}
		}
	}
	gleichAuf(t, "Zahl der Kontextfehler", faelle, 2*len(kontextFormen())*2)
	angehalten := 0
	for _, f := range u1Formen(string(tokenGueltig())) {
		angehalten++
		u := neuMit(t)
		u.srv.Anhalten()
		if rec := u.rufe(ohneAuth("GET", "/a", f.kopf)); rec.Code != 503 || u.audit.Aufrufe() != 0 {
			t.Errorf("%s bei angehaltenem Server: Status %d, %d Aufrufe von Audit, erwartet 503 und 0", f.name, rec.Code, u.audit.Aufrufe())
		}
	}
	gleichAuf(t, "Zahl der Formen bei angehaltenem Server", angehalten, 20)
}

// TestEintragOhneWerteDesRequests hält D3 (Nr. 8, 13 c): der Eintrag trägt keinen Tokenwert, kein Teilstück, keinen Pfad,
// keine Methode und keine Adresse: ein Canary je Ort (Kandidat des falschen Tokens, Pfad, Query, Header, Body, Methode,
// Adresse des Clients, Ursache des Fehlers der Route) und alle zugleich; der gehaltene Eintrag ist als Struktur gleich dem
// Literal, und kein Wert steht in seinem Text.
func TestEintragOhneWerteDesRequests(t *testing.T) {
	o := neueOrte()
	kandidat, methode, adresse := "Kd83Pq1Wv5Lm0Xe7", canary(), "198.51.100.77:4711"
	werte := append(o.alle(), kandidat, methode, adresse)
	// bilde setzt in einen sonst neutralen Request `GET /a` den Canary an jeden genannten Ort.
	bilde := func(orte ...string) *http.Request {
		meth, pfad, query, body := "GET", "/a", "", ""
		for _, ort := range orte {
			switch ort {
			case "Pfad":
				pfad = "/" + o.pfad
			case "Query":
				query = "?q=" + o.query
			case "Body":
				body = o.body
			case "Methode":
				meth = methode
			}
		}
		r := httptest.NewRequest(meth, pfad+query, strings.NewReader(body))
		for _, ort := range orte {
			switch ort {
			case "Header":
				r.Header.Set("X-Test", o.header)
			case "Token":
				r.Header.Set("Authorization", "Bearer "+kandidat)
			case "Adresse":
				r.RemoteAddr = adresse
			}
		}
		return r
	}
	faelle := []struct {
		name  string
		orte  []string
		route error
	}{
		{"Pfad", []string{"Pfad"}, nil}, {"Query", []string{"Query"}, nil}, {"Header", []string{"Header"}, nil},
		{"Body", []string{"Body"}, nil}, {"Kandidat des falschen Tokens", []string{"Token"}, nil},
		{"Methode", []string{"Methode"}, nil}, {"Adresse des Clients", []string{"Adresse"}, nil},
		{"alle Orte zugleich", []string{"Pfad", "Query", "Header", "Body", "Token", "Methode", "Adresse"}, nil},
		{"Ursache des Fehlers der Route", nil, fehlerE2101(errors.New(o.ursache))},
	}
	for _, f := range faelle {
		var u *umgebung
		var rec *httptest.ResponseRecorder
		if f.route != nil {
			var route *routeFake
			u, route = einRoute(t)
			route.err = f.route
			rec = u.rufe(anfrageMit("GET", "/r", nil, nil))
		} else {
			u = neuMit(t)
			rec = u.rufe(bilde(f.orte...))
		}
		pruefe401(t, f.name, rec)
		if n := u.audit.Aufrufe(); n != 1 {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet 1", f.name, n)
			continue
		}
		got := u.audit.Eintraege()
		id := rec.Header().Get("X-Correlation-Id")
		if !slices.Equal(got, []support.Eintrag{eintragSoll(t, id)}) {
			t.Errorf("%s: Einträge %+v, erwartet genau %+v", f.name, got, eintragSoll(t, id))
		}
		text := fmt.Sprintf("%+v", got)
		for i, w := range werte {
			if enthaeltTeil(text, w) {
				t.Errorf("%s: der Wert Nr. %d (%q) steht im Eintrag %s", f.name, i, w, text)
			}
		}
	}
	gleichAuf(t, "Zahl der Orte und Fälle", []int{len(werte), len(faelle)}, []int{8 + 3, 9})
}

// TestAuditVorDemSchreiben hält D4 (Nr. 8): bis Append zurückkehrt, hat der Adapter weder Header noch Body geschrieben: in
// Vor ist kein WriteHeader und kein Write geschehen, der Body ist leer, und im Header steht keiner der Header der Antwort
// (Content-Type, Content-Length, WWW-Authenticate); danach folgen kopf und schreiben. Schleife über die 110 Formen von
// 401, je einmal Vor.
func TestAuditVorDemSchreiben(t *testing.T) {
	for _, q := range quellen401() {
		u, r := q.bau(t)
		var e ereignisse
		rec := httptest.NewRecorder()
		var vorEreignisse, vorKopf []string
		vorBody, vorAufrufe := -1, 0
		u.audit.Vor = func(support.Eintrag) {
			vorAufrufe++
			vorEreignisse = e.lies()
			for k := range rec.Header() {
				vorKopf = append(vorKopf, k)
			}
			vorBody = rec.Body.Len()
		}
		u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, r)
		if vorAufrufe != 1 || len(vorEreignisse) != 0 || vorBody != 0 {
			t.Errorf("%s: Vor %d-mal, davor Ereignisse %q und Body von %d Bytes, erwartet 1, keine und 0", q.name, vorAufrufe, vorEreignisse, vorBody)
		}
		for _, k := range vorKopf {
			if k != "X-Correlation-Id" {
				t.Errorf("%s: vor der Rückkehr von Append steht der Header %s", q.name, k)
			}
		}
		gleichAuf(t, q.name+": Ereignisse nach Append", e.lies(), []string{"kopf", "schreiben"})
		pruefe401(t, q.name, rec)
	}
}

// auditFehler sind die Fehler, die Append liefern kann und die nichts ändern dürfen (D5): die beiden Arten des Ports
// (ErrUngueltig, ein Fehler der Ablage) und Fehler, die der Adapter nicht als eigene deuten darf (Kontextfehler, ein
// *inbound.Fehler mit Code und Status, ein Fehler mit ErrUngueltig in der Ursache).
func auditFehler() []error {
	return []error{
		model.ErrUngueltig, errors.New("Ablage"), io.ErrShortWrite, context.Canceled, context.DeadlineExceeded,
		&inbound.Fehler{Code: model.CodeE7003, Status: 404}, fmt.Errorf("verpackt: %w", model.ErrUngueltig),
	}
}

// TestAuditFehlerAendertNichts hält D5 (Nr. 8): das Ergebnis von Append ändert weder die Antwort noch das Zählen: mit
// jedem der 7 Fehler ist die Antwort (Status, Header, Body) Byte für Byte die ohne Fehler, CountError lief einmal mit
// ODR-E2101, und der Eintrag steht trotzdem im Fake; Schleife über die 110 Formen × 7 Fehler.
func TestAuditFehlerAendertNichts(t *testing.T) {
	faelle := 0
	for _, q := range quellen401() {
		_, ohne := q.sende(t, nil, nil)
		for _, fehler := range auditFehler() {
			faelle++
			u, rec := q.sende(t, nil, fehler)
			if rec.Code != ohne.Code || rec.Body.String() != ohne.Body.String() || !reflect.DeepEqual(rec.Result().Header, ohne.Result().Header) {
				t.Errorf("%s mit Fehler %v: Antwort %d %q %v, ohne Fehler %d %q %v", q.name, fehler, rec.Code, rec.Body.String(),
					rec.Result().Header, ohne.Code, ohne.Body.String(), ohne.Result().Header)
			}
			if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{codeAbgewiesen}) || andere != 0 {
				t.Errorf("%s mit Fehler %v: CountError %q, andere Aufrufe %d, erwartet einmal %s", q.name, fehler, codes, andere, codeAbgewiesen)
			}
			if n := u.audit.Aufrufe(); n != 1 || len(u.audit.Eintraege()) != 1 {
				t.Errorf("%s mit Fehler %v: %d Aufrufe von Audit", q.name, fehler, n)
			}
		}
	}
	gleichAuf(t, "Zahl der Fälle", faelle, 110*len(auditFehler()))
	gleichAuf(t, "Zahl der Fehler", len(auditFehler()), 7)
}

// aufServeHTTP meldet, ob der Stapel der Goroutine, die ruft, die Methode ServeHTTP des Servers von admin trägt: dann läuft
// der Aufrufer auf der Goroutine des Requests (frames, auch eingebettete Aufrufe).
func aufServeHTTP() bool {
	pcs := make([]uintptr, 128)
	frames := runtime.CallersFrames(pcs[:runtime.Callers(2, pcs)])
	for {
		f, mehr := frames.Next()
		if strings.HasSuffix(f.Function, "/admin.(*Server).ServeHTTP") {
			return true
		}
		if !mehr {
			return false
		}
	}
}

// TestAuditAufDerGoroutineDesRequests hält D6 (Nr. 8): Append läuft auf der Goroutine des Requests, ohne Frist: in Vor
// trägt der Stapel ServeHTTP des Adapters (eine andere Goroutine oder eine Hülle mit Frist trüge ihn nicht), und ServeHTTP
// kehrt erst nach Append zurück: solange Vor auf einen Kanal wartet, ist der Aufruf nicht zu Ende, und nach dem
// Freigeben ist er es. Schleife über die 110 Formen; der Test wartet auf Kanäle, die Frist von 5 s begrenzt allein ein
// Hängen im Fehlerfall.
func TestAuditAufDerGoroutineDesRequests(t *testing.T) {
	for _, q := range quellen401() {
		u, r := q.bau(t)
		begonnen, frei, fertig := make(chan struct{}), make(chan struct{}), make(chan struct{})
		var einmal sync.Once
		aufStapel := false
		u.audit.Vor = func(support.Eintrag) {
			einmal.Do(func() {
				aufStapel = aufServeHTTP()
				close(begonnen)
			})
			<-frei
		}
		rec := httptest.NewRecorder()
		go func() {
			u.srv.ServeHTTP(rec, r)
			close(fertig)
		}()
		select {
		case <-begonnen:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: Append begann nicht innerhalb von 5 s", q.name)
		}
		select {
		case <-fertig:
			t.Errorf("%s: ServeHTTP kehrte zurück, bevor Append zurückkehrte", q.name)
		default:
		}
		close(frei)
		select {
		case <-fertig:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: ServeHTTP kehrte nach der Rückkehr von Append nicht zurück", q.name)
		}
		if !aufStapel {
			t.Errorf("%s: Append lief nicht auf der Goroutine von ServeHTTP", q.name)
		}
		pruefe401(t, q.name, rec)
	}
}

// abweisungsRequest liefert einen von vier Requests mit Antwort 401 für D7, gewählt durch art modulo 4: fehlender Header,
// falsches Token, `Bearer` allein und eine Route mit dem Fehler ODR-E2101.
func abweisungsRequest(art int) *http.Request {
	switch art % 4 {
	case 0:
		return ohneAuth("GET", "/ok", nil)
	case 1:
		return ohneAuth("GET", "/ok", bearer("falsch"))
	case 2:
		return ohneAuth("GET", "/ok", map[string][]string{"Authorization": {"Bearer"}})
	}
	return anfrageMit("GET", "/e401", nil, nil)
}

// abweisungsServer baut den Server von D7 mit den Routen GET /ok und GET /e401 (Fehler der Route ODR-E2101 und 401).
func abweisungsServer(t *testing.T) *umgebung {
	t.Helper()
	ok, e401 := neueRoute(), neueRoute()
	e401.err = fehlerE2101(nil)
	return neuMit(t, admin.Route{Methode: "GET", Muster: "/ok", Handler: ok.handle}, admin.Route{Methode: "GET", Muster: "/e401", Handler: e401.handle})
}

// TestJedeAbweisungEinEintrag hält D7 (Nr. 8, 12): jede Abweisung schreibt einen Eintrag, ohne Drosselung und ohne
// Zusammenfassung: 1100 Requests nacheinander und 8 Goroutinen × 150 Requests zugleich (unter dem Detektor) ergeben je so
// viele Einträge mit je einer anderen Korrelations-ID, die Menge der IDs der Einträge ist die der Antworten, jeder Eintrag
// gleicht dem Literal, und CountError lief je Abweisung mit ODR-E2101. Die Zahl der Abweisungen ist unbeschränkt; die
// Prüfgrenze liegt bei 1100 und 1200: eine Schwelle der Drosselung darunter fängt der Test, eine darüber nicht (Plan §3).
func TestJedeAbweisungEinEintrag(t *testing.T) {
	const goroutinen, jeGoroutine, nacheinander = 8, 150, 1100
	pruefe := func(name string, u *umgebung, ids []string, erwartet int) {
		t.Helper()
		gleichAuf(t, name+": Zahl der Antworten", len(ids), erwartet)
		if n := u.audit.Aufrufe(); n != erwartet {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", name, n, erwartet)
			return
		}
		eintraege := u.audit.Eintraege()
		var vonEintraegen []string
		abweichend := 0
		for _, e := range eintraege {
			vonEintraegen = append(vonEintraegen, e.KorrelationsID)
			if e != eintragSoll(t, e.KorrelationsID) {
				abweichend++
			}
		}
		if abweichend != 0 {
			t.Errorf("%s: %d Einträge sind nicht das Literal mit ihrer ID", name, abweichend)
		}
		slices.Sort(vonEintraegen)
		antworten := slices.Clone(ids)
		slices.Sort(antworten)
		if !slices.Equal(vonEintraegen, antworten) || len(slices.Compact(slices.Clone(vonEintraegen))) != erwartet {
			t.Errorf("%s: die %d IDs der Einträge sind nicht die der %d Antworten oder nicht paarweise verschieden", name, len(vonEintraegen), len(antworten))
		}
		codes, andere := zaehlungenVon(u.tele)
		if andere != 0 || len(codes) != erwartet || slices.ContainsFunc(codes, func(c string) bool { return c != codeAbgewiesen }) {
			t.Errorf("%s: %d Aufrufe von CountError, andere Aufrufe %d, erwartet %d-mal %s", name, len(codes), andere, erwartet, codeAbgewiesen)
		}
	}
	u := abweisungsServer(t)
	var ids []string
	for i := range nacheinander {
		rec := u.rufe(abweisungsRequest(i))
		if rec.Code != 401 {
			t.Fatalf("Request %d: Status %d, erwartet 401", i, rec.Code)
		}
		ids = append(ids, rec.Header().Get("X-Correlation-Id"))
	}
	pruefe("nacheinander", u, ids, nacheinander)
	u = abweisungsServer(t)
	var mu sync.Mutex
	var wg sync.WaitGroup
	ids = nil
	for g := range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jeGoroutine {
				rec := u.rufe(abweisungsRequest(g + j))
				mu.Lock()
				if rec.Code != 401 {
					t.Errorf("Goroutine %d, Request %d: Status %d, erwartet 401", g, j, rec.Code)
				}
				ids = append(ids, rec.Header().Get("X-Correlation-Id"))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	pruefe("gleichzeitig", u, ids, goroutinen*jeGoroutine)
}
