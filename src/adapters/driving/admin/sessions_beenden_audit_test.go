package admin_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// TestB11Eintrag hält B11 (SPEC-292 Nr. 11 a, b, d): nach einem Erfolg von Beenden ruft `close` Append genau einmal, mit der
// Art `session_beendet`, `ServiceID` und `SessionID` aus dem Ergebnis des Dienstes (der Fake liefert andere Werte als der
// Pfad), der Korrelations-ID der Antwort und sonst leeren Feldern (das Literal trägt alle sieben), bei `complete` wie bei
// `incomplete`. Im Haltepunkt von Append ist Beenden gelaufen, und die Route hat nichts geschrieben (kein WriteHeader, kein
// Write, kein Header außer der Korrelations-ID); der Haltepunkt lief genau einmal. Ein zweites `close` schreibt einen zweiten
// Eintrag mit eigener Korrelations-ID.
func TestB11Eintrag(t *testing.T) {
	for _, status := range []model.SessionStatus{model.SessionComplete, model.SessionIncomplete} {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = model.Session{ID: "ses-vergeben", ServiceID: "svc-dienst", CreatedAt: zeitA(), Status: status}
		var e ereignisse
		rec := httptest.NewRecorder()
		var vorAufrufe, vorBeenden, vorBody int
		var vorEreignisse, vorKopf []string
		u.audit.Vor = func(support.Eintrag) {
			vorAufrufe++
			vorBeenden = fake.zahlen()[opBeenden]
			vorEreignisse, vorBody = e.lies(), rec.Body.Len()
			for k := range rec.Header() {
				vorKopf = append(vorKopf, k)
			}
		}
		u.srv.ServeHTTP(schreibWriter{ResponseWriter: rec, e: &e}, closeReq("ses-pfad"))
		if vorAufrufe != 1 || vorBeenden != 1 || len(vorEreignisse) != 0 || vorBody != 0 || !slices.Equal(vorKopf, []string{"X-Correlation-Id"}) {
			t.Errorf("%s: Haltepunkt %d Aufrufe, Beenden %d-mal gelaufen, Ereignisse %q, Body %d Bytes, Header %q", status, vorAufrufe, vorBeenden, vorEreignisse, vorBody, vorKopf)
		}
		gleichAuf(t, string(status)+": Ereignisse nach Append", e.lies(), []string{"kopf", "schreiben"})
		korr := rec.Header().Get("X-Correlation-Id")
		if n := u.audit.Aufrufe(); n != 1 || korr == "" || korr != u.ids.letzte() {
			t.Fatalf("%s: %d Aufrufe von Audit, Korrelations-ID %q (letzte vergebene %q)", status, n, korr, u.ids.letzte())
		}
		gleichAuf(t, string(status)+": Eintrag", u.audit.Eintraege(), []support.Eintrag{{
			Art: "session_beendet", ServiceID: "svc-dienst", SessionID: "ses-vergeben", WorkspaceID: "", ImportID: "", ErrorCode: 0, KorrelationsID: korr,
		}})

		u.audit.Vor = nil
		fake.session.ID = "ses-zwei"
		rec2 := u.rufe(closeReq("ses-pfad"))
		eintraege := u.audit.Eintraege()
		if len(eintraege) != 2 || eintraege[1].SessionID != "ses-zwei" || eintraege[1].KorrelationsID != rec2.Header().Get("X-Correlation-Id") || eintraege[1].KorrelationsID == korr {
			t.Errorf("%s: zweites close: Einträge %+v", status, eintraege)
		}
	}
}

// TestB13AppendFehler hält B13 (SPEC-292 Nr. 11 e): jeder der 7 Fehler von Append ändert Status, Body und Header der Antwort
// von `close` nicht (Byte für Byte gleich der Antwort ohne Fehler); Append lief je Fall einmal, und es gab keine Zählung.
func TestB13AppendFehler(t *testing.T) {
	sende := func(fehler error) (*umgebung, *httptest.ResponseRecorder) {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		u.audit.Fehler = fehler
		return u, u.rufe(closeReq("ses-pfad"))
	}
	_, ohne := sende(nil)
	gleichAuf(t, "Zahl der Fehler", len(auditFehler()), 7)
	for _, fehler := range auditFehler() {
		u, rec := sende(fehler)
		if rec.Code != 200 || rec.Body.String() != sessionAJSON || rec.Body.String() != ohne.Body.String() ||
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

// TestN3GleichzeitigeRequests hält N3 (SPEC-292 Nr. 12): in 40 Runden je 16 `close` und 16 `DELETE` zugleich, jeder Request mit
// eigener Kennung, die der Fake in der Antwort zurückgibt (`DELETE`: im Bericht); jede Antwort und jeder Eintrag gehört zum
// eigenen Request (der Eintrag trägt die Korrelations-ID der Antwort derselben Session), und kein `DELETE` schreibt einen
// Eintrag. Alle 32 Requests einer Runde sind zugleich im Aufruf des Dienstes (Treffpunkt): der Detektor von `make test-race`
// sieht die Zugriffe nach dem Treffpunkt ungeordnet, und ein Zustand unter einer Sperre ohne Datenrennen trifft in 40 Runden
// bei jedem Lauf (ein Wirkungs-Fänger, nicht die Hoffnung auf ein Datenrennen).
func TestN3GleichzeitigeRequests(t *testing.T) {
	const n, runden = 16, 40
	nummer := func(id string) int { v, _ := strconv.Atoi(id[1:]); return v }
	u, fake, _ := sessionsUmgebung(t)
	fake.beende = func(_ context.Context, id string) (model.Session, error) {
		return model.Session{ID: id, ServiceID: "s-" + id, CreatedAt: zeitA(), Status: model.SessionComplete}, nil
	}
	fake.loesche = func(_ context.Context, id string) (inbound.Loeschbericht, error) {
		return inbound.Loeschbericht{Entfernt: nummer(id), Zurueckgestellt: nummer(id)%2 == 0}, nil
	}
	korrZuID := map[string]string{}
	for r := range runden {
		p := neuerTreffpunkt(2 * n)
		fake.vor[opBeenden] = func(context.Context) { p.halt() }
		fake.vor[opDelete] = func(context.Context) { p.halt() }
		antworten := gleichzeitig(u, 2*n, func(i int) *http.Request {
			if i < n {
				return closeReq("k" + strconv.Itoa(r*n+i))
			}
			return deleteReq("d" + strconv.Itoa(r*n+i-n))
		})
		if p.verpasst.Load() || p.n.Load() != 2*n {
			t.Errorf("Runde %d: nicht alle %d Requests waren zugleich im Aufruf des Dienstes (%d Aufrufe)", r, 2*n, p.n.Load())
		}
		for i := range n {
			id := "k" + strconv.Itoa(r*n+i)
			soll := `{"id":"` + id + `","service_id":"s-` + id + `","created_at":"2026-11-06T10:00:00.123456789Z","status":"complete"}`
			if rec := antworten[i]; rec.Code != 200 || rec.Body.String() != soll {
				t.Errorf("Runde %d close %d: %d %q", r, i, rec.Code, rec.Body.String())
			}
			korrZuID[antworten[i].Header().Get("X-Correlation-Id")] = id

			num := r*n + i
			sollD := `{"blobs_removed":` + strconv.Itoa(num) + `,"cleanup_deferred":` + strconv.FormatBool(num%2 == 0) + `}`
			if rec := antworten[n+i]; rec.Code != 200 || rec.Body.String() != sollD {
				t.Errorf("Runde %d DELETE %d: %d %q, erwartet %q", r, i, rec.Code, rec.Body.String(), sollD)
			}
		}
	}
	gesehen := map[string]bool{}
	eintraege := u.audit.Eintraege()
	for _, e := range eintraege {
		if korrZuID[e.KorrelationsID] != e.SessionID || e.ServiceID != "s-"+e.SessionID || e.Art != "session_beendet" {
			t.Errorf("Eintrag %+v gehört nicht zur Antwort mit der Korrelations-ID", e)
		}
		gesehen[e.SessionID] = true
	}
	gleichAuf(t, "Einträge, verschiedene Kennungen, Korrelations-IDs", []int{len(eintraege), len(gesehen), len(korrZuID)},
		[]int{runden * n, runden * n, runden * n})
}

// TestN3ZweiGleichzeitigeAufDieselbeSession hält N3 (SPEC-292 Nr. 12, `SPEC-124` Nr. 13) für zwei Requests zugleich auf
// dieselbe Session, in 40 Runden mit je eigener Kennung: ein Fake mit Zustand (eigene Sperre) liefert dem zweiten `close`
// `ODR-E7011` und dem zweiten `DELETE` `ODR-E7003`; beide sind zugleich im Dienst (Treffpunkt). Orakel: je Runde genau ein
// Erfolg und ein Fehler als Literal; `close` schreibt genau einen Eintrag mit der Korrelations-ID des Erfolgs, `DELETE` keinen.
func TestN3ZweiGleichzeitigeAufDieselbeSession(t *testing.T) {
	const runden = 40
	var mu sync.Mutex
	beendet, geloescht := map[string]bool{}, map[string]bool{}
	u, fake, _ := sessionsUmgebung(t)
	fake.beende = func(_ context.Context, id string) (model.Session, error) {
		mu.Lock()
		defer mu.Unlock()
		if beendet[id] {
			return model.Session{}, &inbound.Fehler{Code: model.CodeE7011, Status: 409}
		}
		beendet[id] = true
		return model.Session{ID: id, ServiceID: "svc", CreatedAt: zeitA(), Status: model.SessionComplete}, nil
	}
	fake.loesche = func(_ context.Context, id string) (inbound.Loeschbericht, error) {
		mu.Lock()
		defer mu.Unlock()
		if geloescht[id] {
			return inbound.Loeschbericht{}, &inbound.Fehler{Code: model.CodeE7003, Status: 404}
		}
		geloescht[id] = true
		return inbound.Loeschbericht{Entfernt: 1}, nil
	}
	faelle := []struct {
		name, praefix string
		op            sessionsOp
		req           func(id string) *http.Request
		erfolg        string
		fehlerStatus  int
		fehler        string
	}{
		{"close", "b", opBeenden, closeReq, "", 409, objektE7011},
		{"DELETE", "l", opDelete, deleteReq, `{"blobs_removed":1,"cleanup_deferred":false}`, 404, objektE7003},
	}
	var sieger []string
	for _, f := range faelle {
		for r := range runden {
			id := f.praefix + strconv.Itoa(r)
			p := neuerTreffpunkt(2)
			fake.vor[f.op] = func(context.Context) { p.halt() }
			antworten := gleichzeitig(u, 2, func(int) *http.Request { return f.req(id) })
			if p.verpasst.Load() || p.n.Load() != 2 {
				t.Errorf("%s Runde %d: beide Requests waren nicht zugleich im Dienst (%d Aufrufe)", f.name, r, p.n.Load())
			}
			erfolge, fehler := 0, 0
			for _, rec := range antworten {
				switch {
				case rec.Code == 200:
					erfolge++
					if f.name == "close" {
						sieger = append(sieger, rec.Header().Get("X-Correlation-Id"))
						continue
					}
					if rec.Body.String() != f.erfolg {
						t.Errorf("%s Runde %d: Erfolg %q, erwartet %q", f.name, r, rec.Body.String(), f.erfolg)
					}
				case rec.Code == f.fehlerStatus && rec.Body.String() == f.fehler:
					fehler++
				default:
					t.Errorf("%s Runde %d: Antwort %d %q", f.name, r, rec.Code, rec.Body.String())
				}
			}
			if erfolge != 1 || fehler != 1 {
				t.Errorf("%s Runde %d: %d Erfolge und %d Fehler, erwartet 1 und 1", f.name, r, erfolge, fehler)
			}
		}
	}
	eintraege := u.audit.Eintraege()
	if len(eintraege) != runden || len(sieger) != runden {
		t.Fatalf("%d Einträge und %d Erfolge von close, erwartet je %d (DELETE schreibt keinen)", len(eintraege), len(sieger), runden)
	}
	for i, e := range eintraege {
		if e.Art != "session_beendet" || e.SessionID != "b"+strconv.Itoa(i) || !slices.Contains(sieger, e.KorrelationsID) {
			t.Errorf("Eintrag %d: %+v, erwartet session_beendet mit b%d und der Korrelations-ID eines Erfolgs", i, e, i)
		}
	}
}

// folgeSchritt ist ein Request von M-Folge für `close` und `DELETE` mit dem, was der Dienst ihm liefert, und der Antwort als
// Literal des Vertrags.
type folgeSchritt struct {
	name, route string
	session     model.Session
	bericht     inbound.Loeschbericht
	err         error
	status      int
	body        string
	eintrag     bool
}

// folgeSchritte sind die 9 Antworten von M-Folge: 5 für `close` (zwei verschiedene Sessions mit `complete`, eine mit
// `incomplete`, `ODR-E7011`, ein Fehler ohne Code) und 4 für `DELETE` (zwei Berichte, `ODR-E7003`, `ODR-E7009`).
func folgeSchritte() []folgeSchritt {
	mitStatus := func(s model.Session, status model.SessionStatus) model.Session { s.Status = status; return s }
	return []folgeSchritt{
		{name: "close A complete", route: "close", session: mitStatus(sessionA(), model.SessionComplete), status: 200,
			body: strings.Replace(sessionAJSON, `"status":"open"`, `"status":"complete"`, 1), eintrag: true},
		{name: "close B complete", route: "close", session: sessionB(), status: 200, body: sessionBJSON, eintrag: true},
		{name: "close A incomplete", route: "close", session: mitStatus(sessionA(), model.SessionIncomplete), status: 200,
			body: strings.Replace(sessionAJSON, `"status":"open"`, `"status":"incomplete"`, 1), eintrag: true},
		{name: "close E7011", route: "close", err: &inbound.Fehler{Code: model.CodeE7011, Status: 409}, status: 409, body: objektE7011},
		{name: "close ohne Code", route: "close", err: errors.New("x"), status: 500, body: objekt500},
		{name: "DELETE 3 zurückgestellt", route: "DELETE", bericht: inbound.Loeschbericht{Entfernt: 3, Zurueckgestellt: true}, status: 200, body: `{"blobs_removed":3,"cleanup_deferred":true}`},
		{name: "DELETE 0", route: "DELETE", status: 200, body: berichtNull},
		{name: "DELETE E7003", route: "DELETE", err: &inbound.Fehler{Code: model.CodeE7003, Status: 404}, status: 404, body: objektE7003},
		{name: "DELETE E7009", route: "DELETE", err: &inbound.Fehler{Code: model.CodeE7009, Status: 409}, status: 409, body: objektE7009},
	}
}

// TestFolgeOhneZustandBeendenLoeschen hält SPEC-292 Nr. 12 (die Routen halten keinen Zustand außer den Ports) als Wirkung ohne
// Datenrennen: M-Folge ist jedes geordnete Paar aus den 9 Schritten (81 Paare, auch dieselbe Antwort zweimal), je auf demselben
// Server mit dem Fake neu eingestellt. Jede Antwort ist die zu ihrem Schritt, nie die eines früheren, und die Einträge sind genau
// die der gelungenen `close` in ihrer Folge mit der Korrelations-ID ihrer Antwort. Fänger gegen einen Zwischenspeicher, auch
// einen unter einer Sperre.
func TestFolgeOhneZustandBeendenLoeschen(t *testing.T) {
	alle := folgeSchritte()
	paare := 0
	for _, x := range alle {
		for _, y := range alle {
			paare++
			u, fake, _ := sessionsUmgebung(t)
			var soll []support.Eintrag
			for _, s := range []folgeSchritt{x, y} {
				fake.session, fake.bericht, fake.err = s.session, s.bericht, s.err
				rec := u.rufe(reqFuer(s.route, ""))
				if rec.Code != s.status || rec.Body.String() != s.body {
					t.Errorf("%s dann %s, Schritt %s: %d %q, erwartet %d %q", x.name, y.name, s.name, rec.Code, rec.Body.String(), s.status, s.body)
				}
				if s.eintrag {
					soll = append(soll, support.Eintrag{
						Art: "session_beendet", ServiceID: s.session.ServiceID, SessionID: s.session.ID, KorrelationsID: rec.Header().Get("X-Correlation-Id"),
					})
				}
			}
			gleichAuf(t, x.name+" dann "+y.name+": Einträge", u.audit.Eintraege(), soll)
			if n := fake.gesamt(); n != 2 {
				t.Errorf("%s dann %s: %d Aufrufe des Dienstes, erwartet 2", x.name, y.name, n)
			}
		}
	}
	gleichAuf(t, "Zahl der Schritte und Paare", []int{len(alle), paare}, []int{9, 9 * 9})
}

// TestN4KeineSperreUeberDemAufruf hält N4 (SPEC-292 Nr. 12): zwei Requests derselben Route mit verschiedenen Kennungen sind
// zugleich im Aufruf von Beenden (`close`), von Delete (`DELETE`) und in Append (`close`); der Aufruf kehrt erst zurück, wenn
// der zweite eingetreten ist (Wirkungs-Fänger mit Kanal; die Frist von 10 s ist nur der Abbruch im Fehlschlag).
func TestN4KeineSperreUeberDemAufruf(t *testing.T) {
	stellen := []struct {
		name string
		req  func(id string) *http.Request
	}{{"Beenden", closeReq}, {"Delete", deleteReq}, {"Append", closeReq}}
	for _, s := range stellen {
		u, fake, _ := sessionsUmgebung(t)
		fake.session = sessionA()
		p := neuerTreffpunkt(2)
		switch s.name {
		case "Beenden":
			fake.vor[opBeenden] = func(context.Context) { p.halt() }
		case "Delete":
			fake.vor[opDelete] = func(context.Context) { p.halt() }
		case "Append":
			u.audit.Vor = func(support.Eintrag) { p.halt() }
		}
		for i, rec := range gleichzeitig(u, 2, func(i int) *http.Request { return s.req("ses-" + strconv.Itoa(i)) }) {
			if rec.Code != 200 {
				t.Errorf("%s: Request %d endete mit %d, erwartet 200", s.name, i, rec.Code)
			}
		}
		if p.verpasst.Load() || p.n.Load() != 2 {
			t.Errorf("%s: der zweite Aufruf trat nicht ein, solange der erste im Aufruf war (%d Aufrufe)", s.name, p.n.Load())
		}
	}
	gleichAuf(t, "Zahl der Stellen", len(stellen), 3)
}
