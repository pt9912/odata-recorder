package admin_test

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// wsVariante ist eine der zwei Fassungen eines Schrittes der Folge: der Body (leer: der gültige Body der Route), was der Dienst
// liefert (stelle setzt jedes Ergebnis und jeden Fehler beider Fakes; was die Fassung nicht nennt, ist leer) und die Antwort als
// Literal des Vertrags.
type wsVariante struct {
	body   string
	stelle func(e *wsEnv)
	status int
	lit    string
}

// wsSchritt ist ein Request der Folge in zwei Fassungen, die sich in jedem Wert der Antwort unterscheiden: der erste Request
// eines Paares läuft mit der Fassung 0, der zweite mit der Fassung 1. So unterscheidet sich auch bei gleicher Route die Antwort
// des zweiten von der des ersten, und ein Zwischenspeicher, der die letzte Antwort hält, bleibt nicht unsichtbar.
type wsSchritt struct {
	name  string
	route int
	v     [2]wsVariante
}

// wsB ist ein zweiter Workspace, der sich in jedem Feld von wsA unterscheidet, snapB ein zweiter Snapshot.
func wsB() model.Workspace {
	return model.Workspace{ID: "wsp-2", ServiceID: "svc-2", BaselineSnapshotID: "base-2", Revision: 9, CreatedAt: zeitA().Add(time.Hour)}
}

func snapB() model.WorkspaceSnapshot {
	return model.WorkspaceSnapshot{ID: "snp-2", ServiceID: "svc-2", WorkspaceID: "wsp-2", Name: "nm-2", CreatedAt: zeitA().Add(time.Hour)}
}

// wsBJSON und snapBJSON sind die Antworten von wsB und snapB als Literale.
const (
	wsBJSON   = `{"id":"wsp-2","service_id":"svc-2","baseline_snapshot_id":"base-2","revision":9,"created_at":"2026-11-06T11:00:00.123456789Z"}`
	snapBJSON = `{"id":"snp-2","workspace_id":"wsp-2","name":"nm-2","created_at":"2026-11-06T11:00:00.123456789Z"}`
)

// wsSchritte sind die 12 Antworten von M-Folge, je in zwei Fassungen; jede der 8 Routen kommt mit mindestens einem Erfolg oder
// Fehler vor.
func wsSchritte() []wsSchritt {
	fehler := func(c model.Code, status int, lit string) wsVariante {
		return wsVariante{stelle: func(e *wsEnv) {
			e.wf.err, e.sf.err = &inbound.Fehler{Code: c, Status: status}, &inbound.Fehler{Code: c, Status: status}
		}, status: status, lit: lit}
	}
	ws := func(w model.Workspace, status int, lit string) wsVariante {
		return wsVariante{stelle: func(e *wsEnv) { e.wf.ws = w }, status: status, lit: lit}
	}
	snap := func(s model.WorkspaceSnapshot, status int, lit string) wsVariante {
		return wsVariante{stelle: func(e *wsEnv) { e.sf.snap = s }, status: status, lit: lit}
	}
	liste := func(w []model.Workspace, lit string) wsVariante {
		return wsVariante{stelle: func(e *wsEnv) { e.wf.liste = w }, status: 200, lit: `{"workspaces":[` + lit + `]}`}
	}
	snapListe := func(s []model.WorkspaceSnapshot, lit string) wsVariante {
		return wsVariante{stelle: func(e *wsEnv) { e.sf.liste = s }, status: 200, lit: `{"snapshots":[` + lit + `]}`}
	}
	bericht := func(b inbound.Loeschbericht, lit string) wsVariante {
		return wsVariante{stelle: func(e *wsEnv) { e.wf.bericht = b }, status: 200, lit: lit}
	}
	ungueltig := func(body string) wsVariante {
		return wsVariante{body: body, stelle: func(*wsEnv) {}, status: 400, lit: objekt400}
	}
	a4 := wsA()
	a4.Revision = 4
	a4JSON := strings.Replace(wsAJSON, `"revision":3`, `"revision":4`, 1)
	return []wsSchritt{
		{"GET Liste 200", rList, [2]wsVariante{liste([]model.Workspace{wsA(), wsB()}, wsAJSON+","+wsBJSON), liste([]model.Workspace{wsB(), wsA()}, wsBJSON+","+wsAJSON)}},
		{"POST Workspace 201", rCreate, [2]wsVariante{ws(wsA(), 201, wsAJSON), ws(wsB(), 201, wsBJSON)}},
		{"POST Workspace 400", rCreate, [2]wsVariante{ungueltig("{"), ungueltig("x")}},
		{"POST Workspace 409", rCreate, [2]wsVariante{fehler(model.CodeE7007, 409, wsObjE7007), fehler(model.CodeE7002, 404, wsObjE7002)}},
		{"GET {id} 200", rGet, [2]wsVariante{ws(wsA(), 200, wsAJSON), ws(wsB(), 200, wsBJSON)}},
		{"GET {id} 404", rGet, [2]wsVariante{fehler(model.CodeE7001, 404, wsObjE7001), fehler(model.CodeE3001, 500, objektE3001)}},
		{"DELETE 200", rDelete, [2]wsVariante{
			bericht(inbound.Loeschbericht{Entfernt: 7, Zurueckgestellt: true}, `{"blobs_removed":7,"cleanup_deferred":true}`),
			bericht(inbound.Loeschbericht{Entfernt: 2}, `{"blobs_removed":2,"cleanup_deferred":false}`)}},
		{"reset 200", rReset, [2]wsVariante{ws(wsB(), 200, wsBJSON), ws(a4, 200, a4JSON)}},
		{"GET snapshots 200", rSnList, [2]wsVariante{
			snapListe([]model.WorkspaceSnapshot{snapA(), snapB()}, snapAJSON+","+snapBJSON), snapListe([]model.WorkspaceSnapshot{snapB()}, snapBJSON)}},
		{"POST snapshots 201", rSnCreate, [2]wsVariante{snap(snapB(), 201, snapBJSON), snap(snapA(), 201, snapAJSON)}},
		{"POST snapshots 409", rSnCreate, [2]wsVariante{fehler(model.CodeE7006, 409, wsObjE7006), fehler(model.CodeE7001, 404, wsObjE7001)}},
		{"restore 422", rRestore, [2]wsVariante{fehler(model.CodeE7008, 422, wsObjE7008), fehler(model.CodeE5004, 409, wsObjE5004)}},
	}
}

// wsFuehre stellt die Fakes für die Fassung ein (alles leer außer dem, was sie nennt), sendet den Request des Schrittes und
// liefert Status und Body.
func wsFuehre(e *wsEnv, s wsSchritt, fassung int) (int, string) {
	e.wf.ws, e.wf.liste, e.wf.bericht, e.wf.err = model.Workspace{}, nil, inbound.Loeschbericht{}, nil
	e.sf.snap, e.sf.liste, e.sf.err = model.WorkspaceSnapshot{}, nil, nil
	v := s.v[fassung]
	v.stelle(e)
	r := wsRouten()[s.route]
	body := r.body
	if v.body != "" {
		body = v.body
	}
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	rec := e.rufe(r.request(wsPfadID, "", nil, rd))
	return rec.Code, rec.Body.String()
}

// TestWS16FolgeOhneZustand hält WS16 (SPEC-298 Nr. 12: die Routen halten keinen Zustand zwischen Requests) als Wirkung ohne
// Datenrennen: M-Folge ist jedes geordnete Paar aus den 12 Schritten (144 Paare, auch dieselbe Route zweimal) auf demselben
// Server, der erste Request mit der Fassung 0 seines Schrittes, der zweite mit der Fassung 1; die Antwort des zweiten Requests
// ist das Literal des Vertrags und die auf einem frischen Server. Fänger gegen jeden Zwischenspeicher, auch einen unter einer
// Sperre, der auch ohne Überlappung wirkt, in jeder der 8 Routen.
func TestWS16FolgeOhneZustand(t *testing.T) {
	wsFrist(t)
	alle := wsSchritte()
	gleichAuf(t, "Zahl der Schritte", len(alle), 12)
	routen := map[int]bool{}
	for _, s := range alle {
		routen[s.route] = true
		for f := range 2 {
			if code, body := wsFuehre(wsNeu(t), s, f); code != s.v[f].status || body != s.v[f].lit {
				t.Errorf("%s Fassung %d auf frischem Server: %d %q, erwartet %d %q", s.name, f, code, body, s.v[f].status, s.v[f].lit)
			}
		}
		if s.v[0].lit == s.v[1].lit && s.v[0].status != 400 {
			t.Errorf("%s: beide Fassungen haben dieselbe Antwort", s.name)
		}
	}
	gleichAuf(t, "Routen der Schritte gleich der Länge der Liste des Konstruktors", len(routen), len(wsNeu(t).routen))
	paare := 0
	for _, x := range alle {
		for _, y := range alle {
			paare++
			e := wsNeu(t)
			wsFuehre(e, x, 0)
			code, body := wsFuehre(e, y, 1)
			if code != y.v[1].status || body != y.v[1].lit {
				t.Errorf("%s dann %s: %d %q, erwartet %d %q", x.name, y.name, code, body, y.v[1].status, y.v[1].lit)
			}
			soll := 2
			for _, v := range []wsVariante{x.v[0], y.v[1]} {
				if v.status == 400 {
					soll--
				}
			}
			if n := e.p.gesamt(); n != soll {
				t.Errorf("%s dann %s: %d Aufrufe der Dienste, erwartet %d", x.name, y.name, n, soll)
			}
			if n := e.audit.Aufrufe(); n != 0 {
				t.Errorf("%s dann %s: %d Aufrufe von Audit, erwartet 0", x.name, y.name, n)
			}
		}
	}
	gleichAuf(t, "Zahl der Paare", paare, 12*12)
}
