package admin_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// schritt ist ein Request der Folge und das, was der Dienst ihm liefert, mit der Antwort als Literal des Vertrags: die Route
// ruft den Dienst neu, bei jedem Aufruf (SPEC-292 Nr. 12), und die Antwort hängt allein vom Aufruf ab.
type schritt struct {
	name, route string
	session     model.Session
	liste       []model.Session
	err         error
	status      int
	body        string
	eintrag     bool
}

// sessionB ist eine zweite Session, die sich in jedem Feld von sessionA unterscheidet; sessionBJSON ist ihre Antwort.
func sessionB() model.Session {
	return model.Session{ID: "ses-2", ServiceID: "svc-2", CreatedAt: zeitA().Add(time.Hour), Status: model.SessionComplete}
}

const sessionBJSON = `{"id":"ses-2","service_id":"svc-2","created_at":"2026-11-06T11:00:00.123456789Z","status":"complete"}`

// schritte sind die 14 Antworten von M-Folge: 8 für `GET` (keine und leere Liste, eine Session, die andere, zwei in beiden
// Folgen, zwei Fehler) und 6 für `POST` (zwei Sessions, drei Fehler mit und ohne Code, ErrUngueltig).
func schritte() []schritt {
	a, b := sessionA(), sessionB()
	e3001 := &inbound.Fehler{Code: model.CodeE3001, Status: 500}
	return []schritt{
		{name: "GET nil", route: "GET", status: 200, body: `{"sessions":[]}`},
		{name: "GET leer", route: "GET", liste: []model.Session{}, status: 200, body: `{"sessions":[]}`},
		{name: "GET [A]", route: "GET", liste: []model.Session{a}, status: 200, body: `{"sessions":[` + sessionAJSON + `]}`},
		{name: "GET [B]", route: "GET", liste: []model.Session{b}, status: 200, body: `{"sessions":[` + sessionBJSON + `]}`},
		{name: "GET [A,B]", route: "GET", liste: []model.Session{a, b}, status: 200, body: `{"sessions":[` + sessionAJSON + "," + sessionBJSON + `]}`},
		{name: "GET [B,A]", route: "GET", liste: []model.Session{b, a}, status: 200, body: `{"sessions":[` + sessionBJSON + "," + sessionAJSON + `]}`},
		{name: "GET E3001", route: "GET", err: e3001, status: 500, body: objektE3001},
		{name: "GET ohne Code", route: "GET", err: errors.New("x"), status: 500, body: objekt500},
		{name: "POST A", route: "POST", session: a, status: 201, body: sessionAJSON, eintrag: true},
		{name: "POST B", route: "POST", session: b, status: 201, body: sessionBJSON, eintrag: true},
		{name: "POST E7010", route: "POST", err: &inbound.Fehler{Code: model.CodeE7010, Status: 409}, status: 409, body: objektE7010},
		{name: "POST E5005", route: "POST", err: &inbound.Fehler{Code: model.CodeE5005, Status: 404}, status: 404, body: objektE5005},
		{name: "POST ohne Code", route: "POST", err: errors.New("x"), status: 500, body: objekt500},
		{name: "POST ErrUngueltig", route: "POST", err: model.ErrUngueltig, status: 400, body: objekt400},
	}
}

// TestFolgeOhneZustand hält SPEC-292 Nr. 12 („`GET` liest bei jedem Aufruf neu“, die Routen halten keinen Zustand außer den
// Ports) als Wirkung ohne Datenrennen: M-Folge ist jedes geordnete Paar aus den 14 Schritten (196 Paare, `GET` und `POST` in
// jeder Folge, auch dieselbe Antwort zweimal), je auf demselben Server mit dem Fake neu eingestellt. Jede Antwort ist die zu
// ihrem Schritt, nie die eines früheren, und die Einträge sind genau die der gelungenen `POST` in ihrer Folge mit der
// Korrelations-ID ihrer Antwort. Fänger gegen einen Zwischenspeicher, auch einen unter einer Sperre und einen, der nur
// bestimmte Antworten hält.
func TestFolgeOhneZustand(t *testing.T) {
	alle := schritte()
	gleichAuf(t, "Zahl der Schritte", len(alle), 14)
	paare := 0
	for _, x := range alle {
		for _, y := range alle {
			paare++
			u, fake, _ := sessionsUmgebung(t)
			var soll []support.Eintrag
			for _, s := range []schritt{x, y} {
				fake.session, fake.sessionen, fake.err = s.session, s.liste, s.err
				rec := u.rufe(sessionsReq(s.route, "", nil, strings.NewReader(postGut)))
				if rec.Code != s.status || rec.Body.String() != s.body {
					t.Errorf("%s dann %s, Schritt %s: %d %q, erwartet %d %q", x.name, y.name, s.name, rec.Code, rec.Body.String(), s.status, s.body)
				}
				if s.eintrag {
					soll = append(soll, support.Eintrag{
						Art: "session_angelegt", ServiceID: s.session.ServiceID, SessionID: s.session.ID, KorrelationsID: rec.Header().Get("X-Correlation-Id"),
					})
				}
			}
			gleichAuf(t, x.name+" dann "+y.name+": Einträge", u.audit.Eintraege(), soll)
			if n := fake.gesamt(); n != 2 {
				t.Errorf("%s dann %s: %d Aufrufe des Dienstes, erwartet 2", x.name, y.name, n)
			}
		}
	}
	gleichAuf(t, "Zahl der Paare", paare, 14*14)
}
