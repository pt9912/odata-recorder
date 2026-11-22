package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// TestS12Eintrag hält S12 (SPEC-292 Nr. 11 a, b, d): nach einem Erfolg von Create ruft `POST` Append genau einmal, mit der
// Art `session_angelegt`, `ServiceID` und `SessionID` aus dem Ergebnis des Dienstes (der Fake vergibt andere Werte, als der
// Body nennt), der Korrelations-ID der Antwort und sonst leeren Feldern (das Literal trägt alle sieben). Im Haltepunkt von
// Append ist Create gelaufen, und der Adapter hat nichts geschrieben (kein WriteHeader, kein Write, kein Header außer der
// Korrelations-ID); der Haltepunkt lief vor der Rückkehr von ServeHTTP genau einmal. Ein zweites `POST` ohne `id` schreibt
// einen zweiten Eintrag.
func TestS12Eintrag(t *testing.T) {
	u, fake, _ := sessionsUmgebung(t)
	fake.session = model.Session{ID: "ses-vergeben", ServiceID: "svc-dienst", CreatedAt: zeitA(), Status: model.SessionOpen}
	var e ereignisse
	rec := httptest.NewRecorder()
	var vorAufrufe, vorCreate, vorBody int
	var vorEreignisse, vorKopf []string
	u.audit.Vor = func(support.Eintrag) {
		vorAufrufe++
		vorCreate = fake.zahlen()[opCreate]
		vorEreignisse, vorBody = e.lies(), rec.Body.Len()
		for k := range rec.Header() {
			vorKopf = append(vorKopf, k)
		}
	}
	u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, sessionsReq("POST", "", nil, strings.NewReader(postGut)))
	if vorAufrufe != 1 || vorCreate != 1 || len(vorEreignisse) != 0 || vorBody != 0 || !slices.Equal(vorKopf, []string{"X-Correlation-Id"}) {
		t.Errorf("Haltepunkt: %d Aufrufe, Create %d-mal gelaufen, Ereignisse %q, Body %d Bytes, Header %q", vorAufrufe, vorCreate, vorEreignisse, vorBody, vorKopf)
	}
	gleichAuf(t, "Ereignisse nach Append", e.lies(), []string{"kopf", "schreiben"})
	korr := rec.Header().Get("X-Correlation-Id")
	if n := u.audit.Aufrufe(); n != 1 || korr == "" || korr != u.ids.letzte() {
		t.Fatalf("%d Aufrufe von Audit, Korrelations-ID %q (letzte vergebene %q)", n, korr, u.ids.letzte())
	}
	gleichAuf(t, "Eintrag", u.audit.Eintraege(), []support.Eintrag{{
		Art: "session_angelegt", ServiceID: "svc-dienst", SessionID: "ses-vergeben", WorkspaceID: "", ImportID: "", ErrorCode: 0, KorrelationsID: korr,
	}})

	u.audit.Vor = nil
	rec2 := u.rufe(postReq([]byte(`{"service_id":"svc-b"}`)))
	eintraege := u.audit.Eintraege()
	if len(eintraege) != 2 || eintraege[1].KorrelationsID != rec2.Header().Get("X-Correlation-Id") || eintraege[1].KorrelationsID == korr {
		t.Errorf("zweites POST: Einträge %+v", eintraege)
	}
}

// TestS14AppendFehler hält S14 (SPEC-292 Nr. 11 e): jeder der 7 Fehler von Append ändert Status, Body und Header der Antwort
// nicht; Append lief je Fall einmal.
func TestS14AppendFehler(t *testing.T) {
	sende := func(fehler error) (*umgebung, *httptest.ResponseRecorder) {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		u.audit.Fehler = fehler
		return u, u.rufe(sessionsReq("POST", "", nil, strings.NewReader(postGut)))
	}
	_, ohne := sende(nil)
	gleichAuf(t, "Zahl der Fehler", len(auditFehler()), 7)
	for _, fehler := range auditFehler() {
		u, rec := sende(fehler)
		if rec.Code != 201 || rec.Body.String() != sessionAJSON || rec.Body.String() != ohne.Body.String() ||
			!reflect.DeepEqual(rec.Result().Header, ohne.Result().Header) {
			t.Errorf("Fehler %v: Antwort %d %q %v, ohne Fehler %d %q %v", fehler, rec.Code, rec.Body.String(), rec.Result().Header,
				ohne.Code, ohne.Body.String(), ohne.Result().Header)
		}
		if n := u.audit.Aufrufe(); n != 1 {
			t.Errorf("Fehler %v: %d Aufrufe von Audit, erwartet 1", fehler, n)
		}
		if codes, andere := zaehlungenVon(u.tele); len(codes) != 0 || andere != 0 {
			t.Errorf("Fehler %v: Zählungen %q, andere Aufrufe %d", fehler, codes, andere)
		}
	}
}

// TestS16RahmenSchreibtDasAbweisen hält S16 am Verhalten (SPEC-238 Nr. 15 b): liefert der Dienst einen `*inbound.Fehler` mit
// `ODR-E2101` und 401, schreibt der Rahmen genau einen Eintrag `admin_abgewiesen` und zählt genau einmal; die Route fügt
// weder einen zweiten Eintrag noch eine zweite Zählung noch einen Eintrag `session_angelegt` oder `session_beendet` hinzu
// (alle vier Routen, Pflicht b von SPEC-238 Nr. 15).
func TestS16RahmenSchreibtDasAbweisen(t *testing.T) {
	for _, vr := range vierRouten() {
		route := vr.name
		u, fake, _ := sessionsUmgebung(t)
		fake.err = &inbound.Fehler{Code: model.CodeE2101, Status: 401}
		rec := u.rufe(reqFuer(route, postGut))
		pruefeFehlerAntwort(t, route, rec, 401, objekt401, "")
		korr := rec.Header().Get("X-Correlation-Id")
		gleichAuf(t, route+": Einträge", u.audit.Eintraege(),
			[]support.Eintrag{{Art: "admin_abgewiesen", ErrorCode: model.CodeE2101, KorrelationsID: korr}})
		if codes, andere := zaehlungenVon(u.tele); !slices.Equal(codes, []string{"ODR-E2101"}) || andere != 0 {
			t.Errorf("%s: CountError %q, andere Aufrufe %d", route, codes, andere)
		}
	}
}

// TestS16ErfolgsstatusUndBody hält SPEC-238 Nr. 15 d am Handler: der Status einer Erfolgsantwort liegt von 200 bis 299, das
// JSON steht als `Antwort.Body`, bei allen vier Routen.
func TestS16ErfolgsstatusUndBody(t *testing.T) {
	for i, soll := range []int{200, 201, 200, 200} {
		_, fake, routen := sessionsUmgebung(t)
		fake.session, fake.sessionen = sessionA(), []model.Session{sessionA()}
		ant, err := routen[i].Handler(context.Background(), admin.Anfrage{
			Body: strings.NewReader(postGut), KorrelationsID: "cor-1", Pfadwerte: map[string]string{"id": "ses-pfad"},
		})
		if err != nil || ant.Status != soll || len(ant.Body) == 0 {
			t.Errorf("Route %d: Status %d, Body %q, Fehler %v, erwartet %d und einen Body", i, ant.Status, ant.Body, err, soll)
		}
	}
}

// treffpunkt hält jeden Aufruf von halt, bis ziel Aufrufe eingetreten sind (Wirkungs-Fänger gegen eine Sperre über dem
// Aufruf: unter einer Sperre tritt nie ein zweiter Aufruf ein); die Frist ist nur der Abbruch im Fehlschlag.
type treffpunkt struct {
	n        atomic.Int64
	ziel     int64
	zu       chan struct{}
	einmal   sync.Once
	verpasst atomic.Bool
}

func neuerTreffpunkt(ziel int) *treffpunkt {
	return &treffpunkt{ziel: int64(ziel), zu: make(chan struct{})}
}

func (p *treffpunkt) halt() {
	if p.n.Add(1) == p.ziel {
		p.einmal.Do(func() { close(p.zu) })
	}
	select {
	case <-p.zu:
	case <-time.After(10 * time.Second):
		p.verpasst.Store(true)
	}
}

// gleichzeitig sendet n Requests zugleich (nach einer Schranke) und liefert die Antworten in der Folge der Requests.
func gleichzeitig(u *umgebung, n int, bau func(i int) *http.Request) []*httptest.ResponseRecorder {
	out := make([]*httptest.ResponseRecorder, n)
	schranke := neueSchranke(n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := bau(i)
			out[i] = httptest.NewRecorder()
			schranke.warte()
			u.srv.ServeHTTP(out[i], r)
		}()
	}
	wg.Wait()
	return out
}

// TestN2KeineSperreUeberDemAufruf hält N2 (SPEC-292 Nr. 12): zwei Requests derselben Route sind zugleich im Aufruf von
// sessions und zugleich in Append; der Aufruf kehrt erst zurück, wenn der zweite eingetreten ist (Wirkungs-Fänger).
func TestN2KeineSperreUeberDemAufruf(t *testing.T) {
	stellen := []struct {
		name, route string
		status      int
	}{{"List", "GET", 200}, {"Create", "POST", 201}, {"Append", "POST", 201}}
	for _, s := range stellen {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		p := neuerTreffpunkt(2)
		halt := func(context.Context) { p.halt() }
		switch s.name {
		case "List":
			fake.vor[opList] = halt
		case "Create":
			fake.vor[opCreate] = halt
		case "Append":
			u.audit.Vor = func(support.Eintrag) { p.halt() }
		}
		for i, rec := range gleichzeitig(u, 2, func(int) *http.Request { return sessionsReq(s.route, "", nil, strings.NewReader(postGut)) }) {
			if rec.Code != s.status {
				t.Errorf("%s: Request %d endete mit %d, erwartet %d", s.name, i, rec.Code, s.status)
			}
		}
		if p.verpasst.Load() || p.n.Load() != 2 {
			t.Errorf("%s: der zweite Aufruf trat nicht ein, solange der erste im Aufruf war (%d Aufrufe)", s.name, p.n.Load())
		}
	}
	gleichAuf(t, "Zahl der Stellen", len(stellen), 3)
}

// TestN1GleichzeitigeRequests hält N1 (SPEC-292 Nr. 12): in 40 Runden je 16 gleichzeitige Requests je Route (das Fenster
// zwischen zwei Sperren eines Zustands der Route trifft in einer Runde nur selten, in 40 Runden bei jedem Lauf), jeder mit
// eigener Kennung, die der Fake in der Session zurückgibt; jede Antwort und jeder Eintrag gehört zum eigenen Request (der Eintrag trägt die
// Korrelations-ID der Antwort mit derselben Session). Alle 16 Requests einer Route sind zugleich im Aufruf des Dienstes
// (Treffpunkt); der Detektor von `make test-race` fängt den Zustand zwischen den Requests.
func TestN1GleichzeitigeRequests(t *testing.T) {
	const n, runden = 16, 40
	wert := func(ctx context.Context) string { return "g" + strconv.Itoa(ctx.Value(ctxSchluessel{}).(int)) }
	u, fake, _ := sessionsUmgebung(t)
	fake.liste = func(ctx context.Context) ([]model.Session, error) {
		return []model.Session{{ID: wert(ctx), ServiceID: "s-" + wert(ctx), CreatedAt: zeitA(), Status: model.SessionOpen}}, nil
	}
	fake.erzeuge = func(_ context.Context, serviceID, id string) (model.Session, error) {
		return model.Session{ID: id, ServiceID: "s-" + id, CreatedAt: zeitA(), Status: model.SessionOpen}, nil
	}
	mitKennung := func(r *http.Request, i int) *http.Request {
		return r.WithContext(context.WithValue(r.Context(), ctxSchluessel{}, i))
	}
	// Treffpunkt im Aufruf des Dienstes: alle n Requests sind zugleich im Handler, bevor einer fortfährt; der Detektor sieht
	// die Zugriffe nach dem Treffpunkt so ungeordnet, auch wenn der Host die Requests sonst nacheinander abarbeitet.
	treffen := func(op sessionsOp) *treffpunkt {
		p := neuerTreffpunkt(n)
		fake.vor[op] = func(context.Context) { p.halt() }
		return p
	}
	korrZuID := map[string]string{}
	for r := range runden {
		pGet := treffen(opList)
		gets := gleichzeitig(u, n, func(i int) *http.Request { return mitKennung(sessionsReq("GET", "", nil, nil), r*n+i) })
		pPost := treffen(opCreate)
		posts := gleichzeitig(u, n, func(i int) *http.Request {
			return sessionsReq("POST", "", nil, strings.NewReader(`{"service_id":"x","id":"g`+strconv.Itoa(r*n+i)+`"}`))
		})
		for name, p := range map[string]*treffpunkt{"GET": pGet, "POST": pPost} {
			if p.verpasst.Load() || p.n.Load() != n {
				t.Errorf("Runde %d %s: nicht alle %d Requests waren zugleich im Aufruf des Dienstes (%d Aufrufe)", r, name, n, p.n.Load())
			}
		}
		for i := range n {
			id := "g" + strconv.Itoa(r*n+i)
			soll := `{"id":"` + id + `","service_id":"s-` + id + `","created_at":"2026-11-06T10:00:00.123456789Z","status":"open"}`
			if gets[i].Code != 200 || gets[i].Body.String() != `{"sessions":[`+soll+`]}` {
				t.Errorf("Runde %d GET %d: %d %q", r, i, gets[i].Code, gets[i].Body.String())
			}
			if posts[i].Code != 201 || posts[i].Body.String() != soll {
				t.Errorf("Runde %d POST %d: %d %q", r, i, posts[i].Code, posts[i].Body.String())
			}
			korrZuID[posts[i].Header().Get("X-Correlation-Id")] = id
		}
	}
	eintraege := u.audit.Eintraege()
	gesehen := map[string]bool{}
	for _, e := range eintraege {
		if korrZuID[e.KorrelationsID] != e.SessionID || e.ServiceID != "s-"+e.SessionID || e.Art != "session_angelegt" {
			t.Errorf("Eintrag %+v gehört nicht zur Antwort mit der Korrelations-ID", e)
		}
		gesehen[e.SessionID] = true
	}
	gleichAuf(t, "Einträge, verschiedene Kennungen, Korrelations-IDs", []int{len(eintraege), len(gesehen), len(korrZuID)},
		[]int{runden * n, runden * n, runden * n})
}
