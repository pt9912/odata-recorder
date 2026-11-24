package admin_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// wsNum liefert die Zahl hinter dem ersten Zeichen einer Kennung `g<Zahl>`.
func wsNum(id string) int {
	n, _ := strconv.Atoi(id[1:])
	return n
}

// wsWn ist der Workspace, den der Dienst zur Kennung id liefert (jedes Feld trägt id), wsSnWn der Snapshot zum Workspace id.
func wsWn(id string) model.Workspace {
	return model.Workspace{ID: id, ServiceID: "s-" + id, BaselineSnapshotID: "b-" + id, Revision: int64(wsNum(id)), CreatedAt: zeitA()}
}

func wsSnWn(id string) model.WorkspaceSnapshot {
	return model.WorkspaceSnapshot{ID: "snp-" + id, ServiceID: "s-" + id, WorkspaceID: id, Name: "n-" + id, CreatedAt: zeitA()}
}

// wsWnAntwort ist der Body, den die Route k zur Kennung id liefert, als Literal (SPEC-298 Nr. 6, 7).
func wsWnAntwort(k int, id string) string {
	switch k {
	case rList:
		return `{"workspaces":[` + wsWorkspaceLit(wsWn(id)) + `]}`
	case rDelete:
		return fmt.Sprintf(`{"blobs_removed":%d,"cleanup_deferred":%v}`, wsNum(id), wsNum(id)%2 == 0)
	case rSnList:
		return `{"snapshots":[` + wsSnapshotLit(wsSnWn(id)) + `]}`
	case rSnCreate:
		return wsSnapshotLit(wsSnWn(id))
	}
	return wsWorkspaceLit(wsWn(id))
}

// wsWnEnv stellt die Dienste auf Ergebnisse, in denen jedes Feld die Kennung des Requests trägt: bei den Routen mit Pfadwert
// den Pfadwert, bei `POST` anlegen die `id` des Bodys, bei `GET` Liste den Wert im ctx des Requests.
func wsWnEnv(t *testing.T) *wsEnv {
	e := wsNeu(t)
	e.wf.ergebnis = func(ctx context.Context, op wsOp, a, b string) wsErgebnis {
		id := a
		switch op {
		case wsCreate:
			id = b
		case wsList:
			id, _ = ctx.Value(ctxSchluessel{}).(string)
		}
		w := wsWn(id)
		return wsErgebnis{ws: w, liste: []model.Workspace{w}, bericht: inbound.Loeschbericht{Entfernt: wsNum(id), Zurueckgestellt: wsNum(id)%2 == 0}}
	}
	e.sf.ergebnis = func(_ context.Context, _ wsOp, a, _ string) snErgebnis {
		s := wsSnWn(a)
		return snErgebnis{snap: s, liste: []model.WorkspaceSnapshot{s}}
	}
	return e
}

// wsWnRequest ist der Request der Route k zur Kennung id; ende läuft am Ende des Lesens des Bodys (bei den drei Routen mit Body).
func wsWnRequest(k int, id string, ende func()) *http.Request {
	body := map[int]string{
		rCreate: `{"baseline_snapshot_id":"b-` + id + `","id":"` + id + `"}`, rSnCreate: `{"name":"n-` + id + `"}`, rRestore: `{"snapshot_id":"snp-` + id + `"}`,
	}[k]
	var leser io.Reader
	if body != "" {
		l := wsNeuerLeser([]byte(body), nil)
		l.ende = ende
		leser = l
	}
	req := wsRouten()[k].request(id, "", nil, leser)
	if k == rList {
		req = req.WithContext(context.WithValue(req.Context(), ctxSchluessel{}, id))
	}
	return req
}

// TestWN1GleichzeitigeRequests hält WN1 (SPEC-298 Nr. 12): in 150 Runden je 16 Requests je Route zugleich (8 Routen, 128
// Requests), jeder mit eigener Kennung, die der Fake in allen Feldern zurückgibt; jede Antwort (Literal) und jeder Eintrag
// (Kennungen des Dienstes, Korrelations-ID der Antwort) gehört zum eigenen Request, die vier Routen ohne Eintrag schreiben keinen.
// Alle 16 Requests einer Route sind zugleich in jedem Fenster zwischen einem Aufruf eines Ports und dem nächsten Schritt der Route
// (Treffpunkt je Fenster, M-Fenster 23): am Ende des Lesens des Bodys (3), im Aufruf und nach der Rückkehr der Operation (16) und
// nach der Rückkehr von Append (4). Ein Zustand unter einer Sperre ohne Datenrennen, der ein Fenster umschließt, trifft die
// Antwort eines anderen Requests; ein Fenster ohne Fuge hält kein Treffpunkt (Plan §3, akzeptiertes Negativ (4)). Hält eine Sperre
// die Requests auseinander, endet der Test nach der ersten Frist mit Fatalf.
func TestWN1GleichzeitigeRequests(t *testing.T) {
	wsFrist(t)
	const n, runden = 16, 150
	e := wsWnEnv(t)
	routen := wsRouten()
	soll := map[string]support.Eintrag{}
	for r := range runden {
		fenster := map[string]*treffpunkt{}
		nimm := func(name string) *treffpunkt {
			fenster[name] = neuerTreffpunkt(n)
			return fenster[name]
		}
		lesen, nachAppend := map[int]*treffpunkt{}, map[string]*treffpunkt{}
		for k, rt := range routen {
			if rt.body != "" {
				lesen[k] = nimm(rt.name + " nach dem Lesen des Bodys")
			}
			vor, nach := nimm(rt.name+" im Aufruf"), nimm(rt.name+" nach der Rückkehr der Operation")
			e.halt(rt.op, false, func(context.Context) { vor.halt() })
			e.halt(rt.op, true, func(context.Context) { nach.halt() })
			if art := wsEintragArt(k); art != "" {
				nachAppend[art] = nimm(rt.name + " nach der Rückkehr von Append")
			}
		}
		e.nachAppend.nach = func(es support.Eintrag) { nachAppend[string(es.Art)].halt() }
		antworten := gleichzeitig(e.umgebung, len(routen)*n, func(i int) *http.Request {
			k := i / n
			var ende func()
			if p := lesen[k]; p != nil {
				ende = p.halt
			}
			return wsWnRequest(k, "g"+strconv.Itoa(r*n+i%n), ende)
		})
		gleichAuf(t, "Fenster, Fenster des Lesens, Fenster nach Append", []int{len(fenster), len(lesen), len(nachAppend)}, []int{23, 3, 4})
		for name, p := range fenster {
			if p.verpasst.Load() || p.n.Load() != n {
				t.Fatalf("Runde %d: nicht alle %d Requests waren zugleich im Fenster %s (%d eingetreten): eine Sperre hält sie auseinander", r, n, name, p.n.Load())
			}
		}
		for i, rec := range antworten {
			k, id := i/n, "g"+strconv.Itoa(r*n+i%n)
			if rec.Code != routen[k].status || rec.Body.String() != wsWnAntwort(k, id) {
				t.Errorf("Runde %d, %s mit Kennung %s: %d %q, erwartet %d %q", r, routen[k].name, id, rec.Code, rec.Body.String(), routen[k].status, wsWnAntwort(k, id))
			}
			if art := wsEintragArt(k); art != "" {
				korr := rec.Header().Get("X-Correlation-Id")
				soll[korr] = support.Eintrag{Art: support.AuditArt(art), ServiceID: "s-" + id, WorkspaceID: id, KorrelationsID: korr}
			}
		}
	}
	ist := map[string]support.Eintrag{}
	for _, es := range e.audit.Eintraege() {
		ist[es.KorrelationsID] = es
	}
	gleichAuf(t, "Einträge je Korrelations-ID", ist, soll)
	gleichAuf(t, "Zahl der Einträge gleich 4 Routen × 16 Requests × 150 Runden", []int{len(e.audit.Eintraege()), len(soll)}, []int{4 * n * runden, 4 * n * runden})
}

// TestWN1ZweiGleichzeitigeMitZustand hält WN1 (SPEC-298 Nr. 12) für zwei gleichzeitige `reset` desselben Workspace und zwei
// gleichzeitige `POST` snapshots mit demselben Namen, in 40 Runden: ein Fake mit Zustand (eigene Sperre) hebt die Revision je
// `Reset` und vergibt einen Namen einmal (sonst `ODR-E7006`); beide Requests sind zugleich im Dienst (Treffpunkt). Orakel: beide
// `reset` sind 200 mit den Revisionen 6 und 7 (Start 5) und haben je einen Eintrag mit der Korrelations-ID ihrer Antwort; ein
// `POST` snapshots ist 201 mit einem Eintrag, das andere 409 `ODR-E7006` ohne Eintrag.
func TestWN1ZweiGleichzeitigeMitZustand(t *testing.T) {
	wsFrist(t)
	const runden = 40
	var mu sync.Mutex
	revision, namen := map[string]int64{}, map[string]bool{}
	e := wsNeu(t)
	e.wf.ergebnis = func(_ context.Context, _ wsOp, a, _ string) wsErgebnis {
		mu.Lock()
		defer mu.Unlock()
		revision[a]++
		return wsErgebnis{ws: model.Workspace{ID: a, ServiceID: "svc-1", BaselineSnapshotID: "base-1", Revision: 5 + revision[a], CreatedAt: zeitA()}}
	}
	e.sf.ergebnis = func(_ context.Context, _ wsOp, a, b string) snErgebnis {
		mu.Lock()
		defer mu.Unlock()
		if namen[a+"/"+b] {
			return snErgebnis{err: &inbound.Fehler{Code: model.CodeE7006, Status: 409}}
		}
		namen[a+"/"+b] = true
		return snErgebnis{snap: model.WorkspaceSnapshot{ID: "snp-" + a, ServiceID: "svc-1", WorkspaceID: a, Name: b, CreatedAt: zeitA()}}
	}
	for r := range runden {
		id := "z" + strconv.Itoa(r)
		vorher := len(e.audit.Eintraege())
		p := neuerTreffpunkt(2)
		e.halt(wsReset, false, func(context.Context) { p.halt() })
		antworten := gleichzeitig(e.umgebung, 2, func(int) *http.Request { return wsRouten()[rReset].request(id, "", nil, nil) })
		if p.verpasst.Load() || p.n.Load() != 2 {
			t.Fatalf("Runde %d: beide reset waren nicht zugleich im Dienst (%d Aufrufe): eine Sperre hält sie auseinander", r, p.n.Load())
		}
		lit := func(rev int64) string {
			return "200 " + wsWorkspaceLit(model.Workspace{ID: id, ServiceID: "svc-1", BaselineSnapshotID: "base-1", Revision: rev, CreatedAt: zeitA()})
		}
		got, korr := []string{}, map[string]string{}
		for _, rec := range antworten {
			got = append(got, strconv.Itoa(rec.Code)+" "+rec.Body.String())
			korr[rec.Header().Get("X-Correlation-Id")] = id
		}
		slices.Sort(got)
		gleichAuf(t, "Antworten von reset", got, []string{lit(6), lit(7)})
		eintraege := e.audit.Eintraege()[vorher:]
		if len(eintraege) != 2 || len(korr) != 2 {
			t.Fatalf("Runde %d: %d Einträge nach zwei reset, erwartet 2", r, len(eintraege))
		}
		for _, es := range eintraege {
			if es.Art != "workspace_zurueckgesetzt" || es.ServiceID != "svc-1" || es.WorkspaceID != id || korr[es.KorrelationsID] != id {
				t.Errorf("Runde %d: Eintrag %+v gehört zu keiner der zwei Antworten von reset", r, es)
			}
		}
		if eintraege[0].KorrelationsID == eintraege[1].KorrelationsID {
			t.Errorf("Runde %d: beide Einträge tragen die Korrelations-ID %q", r, eintraege[0].KorrelationsID)
		}

		q := neuerTreffpunkt(2)
		e.halt(snCreate, false, func(context.Context) { q.halt() })
		antworten = gleichzeitig(e.umgebung, 2, func(int) *http.Request {
			return wsRouten()[rSnCreate].request(id, "", nil, wsNeuerLeser([]byte(`{"name":"gleich"}`), nil))
		})
		if q.verpasst.Load() || q.n.Load() != 2 {
			t.Fatalf("Runde %d: beide POST snapshots waren nicht zugleich im Dienst (%d Aufrufe): eine Sperre hält sie auseinander", r, q.n.Load())
		}
		gewinner := 0
		if antworten[0].Code != 201 {
			gewinner = 1
		}
		sieger := model.WorkspaceSnapshot{ID: "snp-" + id, ServiceID: "svc-1", WorkspaceID: id, Name: "gleich", CreatedAt: zeitA()}
		if antworten[gewinner].Code != 201 || antworten[gewinner].Body.String() != wsSnapshotLit(sieger) ||
			antworten[1-gewinner].Code != 409 || antworten[1-gewinner].Body.String() != wsObjE7006 {
			t.Errorf("Runde %d: Antworten %d %q und %d %q, erwartet 201 mit dem Snapshot und 409 ODR-E7006", r, antworten[0].Code, antworten[0].Body.String(),
				antworten[1].Code, antworten[1].Body.String())
		}
		gleichAuf(t, "Eintrag zum POST snapshots mit 201", e.audit.Eintraege()[vorher+2:], []support.Eintrag{{
			Art: "snapshot_erzeugt", ServiceID: "svc-1", WorkspaceID: id, KorrelationsID: antworten[gewinner].Header().Get("X-Correlation-Id"),
		}})
	}
}

// TestWN2KeineSperreUeberDemAufruf hält WN2 (SPEC-298 Nr. 12): zwei Requests derselben Route mit verschiedenen Kennungen sind
// zugleich im Aufruf jeder der 8 Operationen und, bei den 4 Routen mit Eintrag, in Append (12 Stellen); der Aufruf kehrt erst
// zurück, wenn der zweite eingetreten ist (Wirkungs-Fänger mit Kanal, die Frist von 10 s ist nur der Abbruch im Fehlschlag).
// Zählprüfung: 9 gleich 6 + 2 + 1 gegen `NumMethod` der drei Ports, 12 gleich 8 + 4.
func TestWN2KeineSperreUeberDemAufruf(t *testing.T) {
	wsFrist(t)
	stellen := 0
	pruefe := func(name string, k int, p *treffpunkt, e *wsEnv) {
		stellen++
		for i, rec := range gleichzeitig(e.umgebung, 2, func(i int) *http.Request { return wsWnRequest(k, "g"+strconv.Itoa(i+1), nil) }) {
			if rec.Code != wsRouten()[k].status {
				t.Errorf("%s: Request %d endete mit %d, erwartet %d", name, i, rec.Code, wsRouten()[k].status)
			}
		}
		if p.verpasst.Load() || p.n.Load() != 2 {
			t.Errorf("%s: der zweite Aufruf trat nicht ein, solange der erste im Aufruf war (%d Aufrufe)", name, p.n.Load())
		}
	}
	for k, r := range wsRouten() {
		e, p := wsWnEnv(t), neuerTreffpunkt(2)
		e.halt(r.op, false, func(context.Context) { p.halt() })
		pruefe(r.name+" im Aufruf der Operation", k, p, e)
		if wsEintragArt(k) == "" {
			continue
		}
		e, p = wsWnEnv(t), neuerTreffpunkt(2)
		e.audit.Vor = func(support.Eintrag) { p.halt() }
		pruefe(r.name+" in Append", k, p, e)
	}
	gleichAuf(t, "Stellen gleich 8 + 4", stellen, 12)
	gleichAuf(t, "Aufrufstellen gleich NumMethod der drei Ports", reflect.TypeFor[inbound.WorkspaceUseCase]().NumMethod()+
		reflect.TypeFor[inbound.SnapshotUseCase]().NumMethod()+reflect.TypeFor[support.AuditLog]().NumMethod(), 9)
}
