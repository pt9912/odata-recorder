package admin_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// postGut ist ein gültiger Body von `POST /__admin/sessions`.
const postGut = `{"service_id":"svc-body","id":"id-body"}`

// pruefeErfolgsHeader hält die Header einer Erfolgsantwort gegen SPEC-292 Nr. 6 und SPEC-238 Nr. 5: Content-Type,
// Content-Length gleich der Länge des Bodys und X-Correlation-Id, sonst keiner (kein Location).
func pruefeErfolgsHeader(t *testing.T, name string, rec *httptest.ResponseRecorder) {
	t.Helper()
	var namen []string
	for k := range rec.Header() {
		namen = append(namen, k)
	}
	slices.Sort(namen)
	gleichAuf(t, name+": Header", namen, []string{"Content-Length", "Content-Type", "X-Correlation-Id"})
	if rec.Header().Get("Content-Type") != "application/json" || rec.Header().Get("Content-Length") != strconv.Itoa(rec.Body.Len()) {
		t.Errorf("%s: Content-Type %q, Content-Length %q bei %d Bytes", name, rec.Header().Get("Content-Type"),
			rec.Header().Get("Content-Length"), rec.Body.Len())
	}
}

// TestS8FelderDerSession hält S8 (SPEC-292 Nr. 6, 7 b): `POST` antwortet 201 mit den vier Feldern von model.Session in der
// Folge id, service_id, created_at, status, aus dem Ergebnis des Dienstes (nicht aus dem Body), jedes mit eigenem Wert, auch
// ein unbekannter Status, mit den Headern der Erfolgsantwort und ohne Location. Zählprüfung gegen reflect: 4 Felder.
func TestS8FelderDerSession(t *testing.T) {
	gleichAuf(t, "Felder von model.Session", reflect.TypeOf(model.Session{}).NumField(), 4)
	for _, status := range []model.SessionStatus{"open", "complete", "incomplete", "x"} {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = model.Session{ID: "ses-dienst", ServiceID: "svc-dienst", CreatedAt: zeitA(), Status: status}
		rec := u.rufe(sessionsReq("POST", "", nil, strings.NewReader(postGut)))
		soll := `{"id":"ses-dienst","service_id":"svc-dienst","created_at":"2026-11-06T10:00:00.123456789Z","status":"` + string(status) + `"}`
		if rec.Code != 201 || rec.Body.String() != soll {
			t.Errorf("Status %q: Antwort %d %q, erwartet 201 und %q", status, rec.Code, rec.Body.String(), soll)
		}
		pruefeErfolgsHeader(t, "Status "+string(status), rec)
		gleichAuf(t, "Argumente von Create", fake.argumente(), [][2]string{{"svc-body", "id-body"}})
	}
}

// TestS8ZeichenDerWerte hält SPEC-292 Nr. 7 a: der dekodierte Wert gilt, ob ein Zeichen als Escape oder roh steht, ist offen;
// jedes ASCII-Zeichen und die Zeichen `<`, `>`, `&`, `"`, Backslash, U+2028, U+2029 und die Mehrbytefolgen erscheinen
// dekodiert unverändert, jedes ungültige Byte als U+FFFD. Orakel: `encoding/json` liest die Antwort.
func TestS8ZeichenDerWerte(t *testing.T) {
	ungueltig, ersatz := string([]byte{0xFF}), string([]byte{0xEF, 0xBF, 0xBD})
	zeilentrenner := string([]byte{0xE2, 0x80, 0xA8, 0xE2, 0x80, 0xA9})
	sonder := "s<>&\"" + rueck() + "z"
	werte := map[string]string{
		sonder: sonder, zeilentrenner: zeilentrenner, "a" + ungueltig + "b" + ungueltig + ungueltig: "a" + ersatz + "b" + ersatz + ersatz,
	}
	for c := range byte(0x20) {
		werte["c"+string([]byte{c})] = "c" + string([]byte{c})
	}
	for roh, soll := range werte {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = model.Session{ID: roh, ServiceID: roh, CreatedAt: zeitA(), Status: model.SessionStatus(roh)}
		rec := u.rufe(postReq([]byte(postGut)))
		var got map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("Wert %q: die Antwort ist kein JSON: %v (%q)", roh, err, rec.Body.String())
		}
		for _, feld := range []string{"id", "service_id", "status"} {
			if got[feld] != soll {
				t.Errorf("Wert %q Feld %s: dekodiert %q, erwartet %q", roh, feld, got[feld], soll)
			}
		}
	}
	gleichAuf(t, "Zahl der Werte", len(werte), 3+32)
}

// zeitFaelle ist M-Zeit (SPEC-292 Nr. 7 c): 7 Werte mit der Zeichenkette des Vertrags als Literal, `null` für die Nullzeit.
func zeitFaelle() []struct {
	name string
	zeit time.Time
	soll string
} {
	return []struct {
		name string
		zeit time.Time
		soll string
	}{
		{"Nanosekunden", time.Date(2026, 11, 6, 10, 0, 0, 123456789, time.UTC), `"2026-11-06T10:00:00.123456789Z"`},
		{"neun Nullen", time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC), `"2026-11-06T10:00:00.000000000Z"`},
		{"hundert Millionen", time.Date(2026, 11, 6, 10, 0, 0, 100000000, time.UTC), `"2026-11-06T10:00:00.100000000Z"`},
		{"anderer Ort", time.Date(2026, 11, 6, 10, 0, 0, 0, time.FixedZone("plus2", 2*3600)), `"2026-11-06T08:00:00.000000000Z"`},
		{"Nullzeit", time.Time{}, `null`},
		{"Jahr 1 mit 1 ns", time.Date(1, 1, 1, 0, 0, 0, 1, time.UTC), `"0001-01-01T00:00:00.000000001Z"`},
		{"Jahr 9999", time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC), `"9999-12-31T23:59:59.999999999Z"`},
	}
}

// TestS9Zeit hält S9 (SPEC-292 Nr. 7 c): jeder der 7 Werte steht als `created_at` der Antwort von `POST` und eines Eintrags
// von `GET` (14 Fälle) in der Form des Vertrags. Zählprüfung: ein Zeitfeld in model.Session.
func TestS9Zeit(t *testing.T) {
	zeitfelder := 0
	for i := range reflect.TypeOf(model.Session{}).NumField() {
		if reflect.TypeOf(model.Session{}).Field(i).Type == reflect.TypeOf(time.Time{}) {
			zeitfelder++
		}
	}
	gleichAuf(t, "Zeitfelder und Werte", []int{zeitfelder, len(zeitFaelle())}, []int{1, 7})
	faelle := 0
	for _, z := range zeitFaelle() {
		s := model.Session{ID: "ses-1", ServiceID: "svc", CreatedAt: z.zeit, Status: model.SessionOpen}
		eintrag := `{"id":"ses-1","service_id":"svc","created_at":` + z.soll + `,"status":"open"}`
		u, fake, _ := sessionsUmgebung(t)
		fake.session, fake.sessionen = s, []model.Session{s}
		if rec := u.rufe(postReq([]byte(postGut))); rec.Code != 201 || rec.Body.String() != eintrag {
			t.Errorf("POST %s: %d %q, erwartet %q", z.name, rec.Code, rec.Body.String(), eintrag)
		}
		if rec := u.rufe(sessionsReq("GET", "", nil, nil)); rec.Code != 200 || rec.Body.String() != `{"sessions":[`+eintrag+`]}` {
			t.Errorf("GET %s: %d %q, erwartet %q", z.name, rec.Code, rec.Body.String(), eintrag)
		}
		faelle += 2
	}
	gleichAuf(t, "Fälle", faelle, 14)
}

// TestS10Liste hält S10 (SPEC-292 Nr. 6, 7 d): `GET` antwortet 200 mit `{"sessions":[…]}` in der Reihenfolge des Dienstes
// (nicht sortiert), ohne Session `[]`, auch bei nil; Prüfgrenze 3 Sessions (akzeptiertes Negativ im Plan).
func TestS10Liste(t *testing.T) {
	mit := func(id string) model.Session {
		return model.Session{ID: id, ServiceID: "svc", CreatedAt: zeitA(), Status: model.SessionComplete}
	}
	eintrag := func(id string) string {
		return `{"id":"` + id + `","service_id":"svc","created_at":"2026-11-06T10:00:00.123456789Z","status":"complete"}`
	}
	faelle := []struct {
		name  string
		liste []model.Session
		soll  string
	}{
		{"nil", nil, `{"sessions":[]}`},
		{"leer", []model.Session{}, `{"sessions":[]}`},
		{"eine", []model.Session{mit("a")}, `{"sessions":[` + eintrag("a") + `]}`},
		{"drei", []model.Session{mit("c"), mit("a"), mit("b")}, `{"sessions":[` + eintrag("c") + "," + eintrag("a") + "," + eintrag("b") + `]}`},
	}
	for _, f := range faelle {
		u, fake, _ := sessionsUmgebung(t)
		fake.sessionen = f.liste
		rec := u.rufe(sessionsReq("GET", "", nil, nil))
		if rec.Code != 200 || rec.Body.String() != f.soll {
			t.Errorf("%s: %d %q, erwartet 200 und %q", f.name, rec.Code, rec.Body.String(), f.soll)
		}
		pruefeErfolgsHeader(t, f.name, rec)
	}
	gleichAuf(t, "Fälle", len(faelle), 4)
}

// fehlerFall ist eine Zelle von M-Fehler (SPEC-292 Nr. 8): der Fehler des Dienstes, ob zugleich ein Wert kommt, und die
// Erwartung als Literale: abbruch heißt, die Verbindung wird abgebrochen (kein Zählen, keine Antwort), sonst Status, Code
// und Body.
type fehlerFall struct {
	name         string
	err          error
	mitWert      bool
	abbruch      bool
	status       int
	code, objekt string
}

// fehlerE7003 ist ein Fehler mit Code, den ein Dienst der Sessions liefert.
func fehlerE7003() *inbound.Fehler { return &inbound.Fehler{Code: model.CodeE7003, Status: 404} }

// allgemeineAusgaenge liefert die 8 allgemeinen Ausgänge je Route.
func allgemeineAusgaenge() []fehlerFall {
	e7003 := fehlerE7003()
	return []fehlerFall{
		{name: "ErrUngueltig", err: model.ErrUngueltig, status: 400, code: "ODR-E2102", objekt: objekt400},
		{name: "Fehler ohne Code", err: errors.New("x"), status: 500, code: "ODR-E9001", objekt: objekt500},
		{name: "ErrNichtGefunden", err: errors.New("nicht gefunden"), status: 500, code: "ODR-E9001", objekt: objekt500},
		{name: "Canceled", err: context.Canceled, abbruch: true},
		{name: "DeadlineExceeded", err: context.DeadlineExceeded, abbruch: true},
		{name: "Canceled neben E7003", err: errors.Join(context.Canceled, e7003), abbruch: true},
		{name: "verpackter E7003", err: fmt.Errorf("x: %w", e7003), status: 404, code: "ODR-E7003", objekt: objektE7003},
		{name: "Wert und E7003", err: e7003, mitWert: true, status: 404, code: "ODR-E7003", objekt: objektE7003},
	}
}

// codeFall ist ein Fall der Tabelle von SPEC-292 Nr. 8: ein Fehler mit Code und seinem Status.
func codeFall(c model.Code, status int, code, objekt string) fehlerFall {
	return fehlerFall{name: code, err: &inbound.Fehler{Code: c, Status: status}, status: status, code: code, objekt: objekt}
}

// fehlerZellen liefert die Zellen von M-Fehler je Route (die Tabelle von Nr. 8 und die acht allgemeinen Ausgänge).
func fehlerZellen() map[string][]fehlerFall {
	gemeinsam := []fehlerFall{codeFall(model.CodeE3001, 500, "ODR-E3001", objektE3001), codeFall(model.CodeE9001, 500, "ODR-E9001", objekt500)}
	post := append([]fehlerFall{
		codeFall(model.CodeE7010, 409, "ODR-E7010", objektE7010), codeFall(model.CodeE5005, 404, "ODR-E5005", objektE5005),
	}, gemeinsam...)
	return map[string][]fehlerFall{
		"GET":  append(allgemeineAusgaenge(), gemeinsam...),
		"POST": append(allgemeineAusgaenge(), post...),
	}
}

// TestS11S13Fehler hält S11 und S13 (SPEC-292 Nr. 8, 11 c): M-Fehler (8 allgemeine Ausgänge und die Codes der Tabelle je
// Route, 22 Zellen) mit Status, Code und Body als Literale, genau einem Zählen je Fehlerantwort und keinem bei einem
// Kontextfehler (Verbindung abgebrochen, nichts geschrieben); M-Audit (24 Zellen: die 22 und je ein Erfolg): genau eine
// Zelle, `POST` mit Erfolg, schreibt einen Eintrag.
func TestS11S13Fehler(t *testing.T) {
	zellen, mitEintrag := 0, 0
	for _, route := range []string{"GET", "POST"} {
		erfolg := fehlerFall{name: "Erfolg", status: 200, objekt: `{"sessions":[` + sessionAJSON + `]}`}
		if route == "POST" {
			erfolg.status, erfolg.objekt = 201, sessionAJSON
		}
		for _, f := range slices.Concat(fehlerZellen()[route], []fehlerFall{erfolg}) {
			zellen++
			name := route + " " + f.name
			u, fake, _ := sessionsUmgebung(t)
			fake.err = f.err
			if f.mitWert || f.err == nil {
				fake.session, fake.sessionen = sessionA(), []model.Session{sessionA()}
			}
			rec := httptest.NewRecorder()
			r := sessionsReq(route, "", nil, strings.NewReader(postGut))
			abbruch := abbruchVon(func() { u.srv.ServeHTTP(rec, r) })
			codes, andere := zaehlungenVon(u.tele)
			var soll []string
			switch {
			case f.abbruch:
				if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || rec.Result().StatusCode != 200 {
					t.Errorf("%s: Panik %v, Body %q: erwartet den Abbruch der Verbindung ohne Antwort", name, abbruch, rec.Body.String())
				}
			case f.err == nil:
				if abbruch != nil || rec.Code != f.status || rec.Body.String() != f.objekt {
					t.Errorf("%s: Antwort %d %q, erwartet %d %q", name, rec.Code, rec.Body.String(), f.status, f.objekt)
				}
			default:
				soll = []string{f.code}
				if abbruch != nil {
					t.Errorf("%s: Panik %v", name, abbruch)
				}
				pruefeFehlerAntwort(t, name, rec, f.status, f.objekt, "")
			}
			if !slices.Equal(codes, soll) || andere != 0 {
				t.Errorf("%s: CountError %q, andere Aufrufe %d, erwartet %q", name, codes, andere, soll)
			}
			wollen := 0
			if f.err == nil && route == "POST" {
				wollen, mitEintrag = 1, mitEintrag+1
			}
			if n := u.audit.Aufrufe(); n != wollen {
				t.Errorf("%s: %d Aufrufe von Audit, erwartet %d", name, n, wollen)
			}
		}
	}
	gleichAuf(t, "Zellen von M-Audit und Zellen mit Eintrag", []int{zellen, mitEintrag}, []int{24, 1})
}

// TestS11Zahlen hält M-Fehler gegen die Quelle: 22 Zellen gleich 2 × 8 + 6, jeder Code der Tabelle von SPEC-292 Nr. 8 steht
// in model.AlleCodes() mit dem Status des Literals, und die Fehlerobjekte der Literale gleichen denen der Tabelle der Codes.
func TestS11Zahlen(t *testing.T) {
	z := fehlerZellen()
	gleichAuf(t, "Zellen", len(z["GET"])+len(z["POST"]), 22)
	gleichAuf(t, "Zellen gleich 2 × 8 + 6", len(z["GET"])+len(z["POST"]), 2*8+(2+4))
	alle := model.AlleCodes()
	for route, fs := range z {
		for _, f := range fs {
			fe, ok := f.err.(*inbound.Fehler)
			if !ok || f.abbruch || f.name == "Wert und E7003" || f.name == "verpackter E7003" {
				continue
			}
			c, bekannt := model.ParseCode(f.code)
			if !bekannt || !slices.Contains(alle, c) || c != fe.Code || !c.StatusErlaubt(f.status) || fe.Status != f.status {
				t.Errorf("%s %s: Code %q, Status %d passt nicht zur Tabelle der Codes", route, f.name, f.code, f.status)
			}
			if !strings.Contains(f.objekt, `"message":"`+c.Message()+`"`) {
				t.Errorf("%s %s: das Literal %q nennt die message %q nicht", route, f.name, f.objekt, c.Message())
			}
		}
	}
}

// TestS11Unveraendert hält SPEC-292 Nr. 8 an der Route allein: der Handler gibt jeden Fehler des Dienstes unverändert zurück
// (dasselbe Objekt, nicht verpackt) und keine Antwort daneben, auch wenn der Dienst einen Wert und den Fehler liefert.
func TestS11Unveraendert(t *testing.T) {
	zellen := 0
	for i, route := range []string{"GET", "POST"} {
		for _, f := range fehlerZellen()[route] {
			zellen++
			_, fake, routen := sessionsUmgebung(t)
			fake.err = f.err
			fake.session, fake.sessionen = sessionA(), []model.Session{sessionA()}
			ant, err := routen[i].Handler(context.Background(), admin.Anfrage{Methode: route, Body: strings.NewReader(postGut)})
			if err != f.err || ant.Status != 0 || len(ant.Body) != 0 {
				t.Errorf("%s %s: Handler liefert %v und Antwort %d %q, erwartet denselben Fehler und keine Antwort", route, f.name, err, ant.Status, ant.Body)
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 22)
}
