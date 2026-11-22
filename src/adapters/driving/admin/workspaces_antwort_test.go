package admin_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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

// TestWS10Status hält WS10 (SPEC-298 Nr. 6; SPEC-238 Nr. 5): 201 bei `POST /__admin/workspaces` und `POST …/snapshots`, sonst
// 200; die Erfolgsantwort trägt `Content-Type`, `Content-Length` und `X-Correlation-Id` und sonst keinen Header, der Body ist
// nicht leer, und der Status hängt allein von der Route ab: M-Status sind 8 Zellen (je Route) und `DELETE` mit
// `cleanup_deferred` wahr, je Zelle mit zwei Ergebnissen, die sich in `revision` (0 und 7) und `Entfernt` (0 und 3) unterscheiden.
func TestWS10Status(t *testing.T) {
	wsFrist(t)
	zellen := 0
	for _, r := range wsRouten() {
		for _, deferred := range []bool{false, true} {
			if deferred && r.op != wsDelete {
				continue
			}
			for _, wert := range []int{0, 3, 7} {
				e := wsNeu(t)
				e.wf.ws.Revision, e.wf.bericht = int64(wert), inbound.Loeschbericht{Entfernt: wert, Zurueckgestellt: deferred}
				rec := e.rufe(r.gut())
				name := fmt.Sprintf("%s (Wert %d, zurückgestellt %v)", r.name, wert, deferred)
				if rec.Code != r.status || rec.Body.Len() == 0 {
					t.Errorf("%s: Status %d, Body %q, erwartet %d und nicht leer", name, rec.Code, rec.Body.String(), r.status)
				}
				pruefeErfolgsHeader(t, name, rec)
				if wert == 0 {
					zellen++
				}
			}
		}
	}
	gleichAuf(t, "Zellen", zellen, 9)
}

// wsFolge liefert die Schlüssel des Objekts auf oberster Ebene in der Folge ihres Auftretens.
func wsFolge(t *testing.T, body string) []string {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(body))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		t.Fatalf("Body %q ist kein Objekt: %v", body, err)
	}
	var out []string
	for dec.More() {
		k, _ := dec.Token()
		out = append(out, k.(string))
		var wert json.RawMessage
		if err := dec.Decode(&wert); err != nil {
			t.Fatalf("Body %q: %v", body, err)
		}
	}
	return out
}

// wsTexte sind die 8 Formen eines Textfeldes (SPEC-298 Nr. 7 c): der Wert des Dienstes und der Wert, den ein Leser der Antwort
// bekommt (die Bytes aus Bytewerten).
func wsTexte() [][2]string {
	return [][2]string{
		{"x", "x"}, {"", ""}, {`a"b`, `a"b`}, {"a" + rueck() + "b", "a" + rueck() + "b"}, {"a<b>&c", "a<b>&c"},
		{"a" + string([]byte{0xE2, 0x80, 0xA8}) + "b", "a" + string([]byte{0xE2, 0x80, 0xA8}) + "b"},
		{string([]byte{0xC3, 0xA4}), string([]byte{0xC3, 0xA4})}, {"a" + string([]byte{0xFF}) + "b", "a" + string([]byte{0xEF, 0xBF, 0xBD}) + "b"},
	}
}

// wsZeiten sind die 7 Formen von `created_at`: die Zeit des Dienstes und das Literal in der Antwort.
func wsZeiten() []struct {
	z   time.Time
	lit string
} {
	ort := time.FixedZone("x", 2*3600)
	ort2 := time.FixedZone("y", -(5*3600 + 30*60))
	return []struct {
		z   time.Time
		lit string
	}{
		{time.Date(2026, 11, 6, 10, 0, 0, 123456789, time.UTC), `"2026-11-06T10:00:00.123456789Z"`},
		{time.Date(2026, 11, 6, 12, 0, 0, 123456789, ort), `"2026-11-06T10:00:00.123456789Z"`},
		{time.Date(2026, 11, 6, 4, 30, 0, 123456789, ort2), `"2026-11-06T10:00:00.123456789Z"`},
		{time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC), `"2026-11-06T10:00:00.000000000Z"`},
		{time.Date(2026, 11, 6, 10, 0, 0, 1, time.UTC), `"2026-11-06T10:00:00.000000001Z"`},
		{time.Time{}, `null`},
		{time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC), `"9999-12-31T23:59:59.999999999Z"`},
	}
}

// wsRoutenMitWorkspace sind die vier Routen, die einen Workspace liefern: anlegen, lesen, `reset` und `restore`.
func wsRoutenMitWorkspace() []wsRoute {
	rs := wsRouten()
	return []wsRoute{rs[rCreate], rs[rGet], rs[rReset], rs[rRestore]}
}

// TestWS11Workspace hält WS11 (SPEC-298 Nr. 7 a, c) für den Workspace in M-Workspace: je Route 36 Werte (3 Textfelder × 8
// Formen, `revision` mit 5 Werten, `created_at` mit 7 Formen) und der Nullwert; die Felder stehen in der Folge des Vertrags,
// keines fehlt und keines kommt hinzu, die dekodierten Werte sind die des Dienstes, `revision` steht als Ziffernfolge.
func TestWS11Workspace(t *testing.T) {
	wsFrist(t)
	gleichAuf(t, "Felder von model.Workspace", reflect.TypeFor[model.Workspace]().NumField(), 5)
	folge := []string{"id", "service_id", "baseline_snapshot_id", "revision", "created_at"}
	faelle := 0
	pruefe := func(name string, r wsRoute, w model.Workspace, soll map[string]any, revision, zeit string) {
		t.Helper()
		e := wsNeu(t)
		e.wf.ws = w
		rec := e.rufe(r.gut())
		faelle++
		body := rec.Body.String()
		if rec.Code != r.status || !slices.Equal(wsFolge(t, body), folge) {
			t.Errorf("%s %s: Status %d, Felder %v, erwartet %d und %v (%q)", r.name, name, rec.Code, wsFolge(t, body), r.status, folge, body)
		}
		var got map[string]any
		dec := json.NewDecoder(strings.NewReader(body))
		dec.UseNumber()
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("%s %s: %v", r.name, name, err)
		}
		for k, v := range soll {
			if got[k] != v {
				t.Errorf("%s %s: Feld %s = %#v, erwartet %#v", r.name, name, k, got[k], v)
			}
		}
		if !strings.Contains(body, `"revision":`+revision+`,"created_at":`+zeit+`}`) {
			t.Errorf("%s %s: Body %q trägt nicht `revision` %s und `created_at` %s", r.name, name, body, revision, zeit)
		}
	}
	for _, r := range wsRoutenMitWorkspace() {
		for i, feld := range folge[:3] {
			for _, f := range wsTexte() {
				w := model.Workspace{ID: "i", ServiceID: "s", BaselineSnapshotID: "b", Revision: 4, CreatedAt: zeitA()}
				soll := map[string]any{"id": "i", "service_id": "s", "baseline_snapshot_id": "b"}
				switch i {
				case 0:
					w.ID = f[0]
				case 1:
					w.ServiceID = f[0]
				default:
					w.BaselineSnapshotID = f[0]
				}
				soll[feld] = f[1]
				pruefe(feld+" "+strconv.Quote(f[0]), r, w, soll, "4", `"2026-11-06T10:00:00.123456789Z"`)
			}
		}
		for _, rev := range []int64{-1, 0, 1, 2147483648, math.MaxInt64} {
			w := model.Workspace{ID: "i", ServiceID: "s", BaselineSnapshotID: "b", Revision: rev, CreatedAt: zeitA()}
			pruefe("revision "+strconv.FormatInt(rev, 10), r, w, map[string]any{"revision": json.Number(strconv.FormatInt(rev, 10))}, strconv.FormatInt(rev, 10), `"2026-11-06T10:00:00.123456789Z"`)
		}
		for _, z := range wsZeiten() {
			pruefe("created_at "+z.lit, r, model.Workspace{ID: "i", ServiceID: "s", BaselineSnapshotID: "b", Revision: 4, CreatedAt: z.z}, nil, "4", z.lit)
		}
		pruefe("Nullwert", r, model.Workspace{}, map[string]any{"id": "", "service_id": "", "baseline_snapshot_id": "", "revision": json.Number("0"), "created_at": nil}, "0", "null")
	}
	gleichAuf(t, "Fälle", faelle, 4*(24+5+7+1))
}

// TestWS11Snapshot hält WS11 (SPEC-298 Nr. 7 b, c) für den Snapshot in M-Snapshot: 3 Textfelder × 8 Formen, `created_at` mit 7
// Formen und der Nullwert (32 Fälle); `service_id`, `state_blob`, `state_hash` und `metadata_hash` stehen in keiner Antwort
// (4 Canary, je einer in einem der vier Felder).
func TestWS11Snapshot(t *testing.T) {
	wsFrist(t)
	gleichAuf(t, "Felder von model.WorkspaceSnapshot", reflect.TypeFor[model.WorkspaceSnapshot]().NumField(), 8)
	r := wsRouten()[rSnCreate]
	folge := []string{"id", "workspace_id", "name", "created_at"}
	faelle := 0
	pruefe := func(name string, s model.WorkspaceSnapshot, soll map[string]any, zeit string) {
		t.Helper()
		e := wsNeu(t)
		e.sf.snap = s
		rec := e.rufe(r.gut())
		faelle++
		body := rec.Body.String()
		if rec.Code != 201 || !slices.Equal(wsFolge(t, body), folge) || !strings.HasSuffix(body, `"created_at":`+zeit+`}`) {
			t.Errorf("%s: Status %d, Body %q, erwartet 201, die Felder %v und created_at %s", name, rec.Code, body, folge, zeit)
		}
		var got map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for k, v := range soll {
			if got[k] != v {
				t.Errorf("%s: Feld %s = %#v, erwartet %#v", name, k, got[k], v)
			}
		}
	}
	for i, feld := range folge[:3] {
		for _, f := range wsTexte() {
			s := model.WorkspaceSnapshot{ID: "i", WorkspaceID: "w", Name: "n", CreatedAt: zeitA()}
			soll := map[string]any{"id": "i", "workspace_id": "w", "name": "n"}
			switch i {
			case 0:
				s.ID = f[0]
			case 1:
				s.WorkspaceID = f[0]
			default:
				s.Name = f[0]
			}
			soll[feld] = f[1]
			pruefe(feld+" "+strconv.Quote(f[0]), s, soll, `"2026-11-06T10:00:00.123456789Z"`)
		}
	}
	for _, z := range wsZeiten() {
		pruefe("created_at "+z.lit, model.WorkspaceSnapshot{ID: "i", WorkspaceID: "w", Name: "n", CreatedAt: z.z}, nil, z.lit)
	}
	pruefe("Nullwert", model.WorkspaceSnapshot{}, map[string]any{"id": "", "workspace_id": "", "name": "", "created_at": nil}, "null")
	gleichAuf(t, "Fälle", faelle, 32)
	for i := range 4 {
		s := model.WorkspaceSnapshot{ID: "i", WorkspaceID: "w", Name: "n", CreatedAt: zeitA()}
		versteckt := []*string{&s.ServiceID, &s.StateBlob, &s.StateHash, &s.MetadataHash}
		*versteckt[i] = canary() + strconv.Itoa(i)
		e := wsNeu(t)
		e.sf.snap = s
		if rec := e.rufe(r.gut()); strings.Contains(antwortText(rec), canary()) {
			t.Errorf("Feld %d des Snapshots, das keine Antwort trägt, steht in der Antwort: %q", i, rec.Body.String())
		}
	}
}

// TestWS11Bericht hält WS11 (SPEC-298 Nr. 7, 10) für den Bericht von `DELETE`: `Entfernt` mit 6 Werten (darunter der größte
// `int` und -1, ohne Kappen) × `Zurueckgestellt` mit 2 = 12 Fälle, die Felder in der Folge des Vertrags.
func TestWS11Bericht(t *testing.T) {
	wsFrist(t)
	gleichAuf(t, "Felder von inbound.Loeschbericht", reflect.TypeFor[inbound.Loeschbericht]().NumField(), 2)
	faelle := 0
	for _, n := range []int{0, 1, 3, 7, math.MaxInt, -1} {
		for _, z := range []bool{false, true} {
			e := wsNeu(t)
			e.wf.bericht = inbound.Loeschbericht{Entfernt: n, Zurueckgestellt: z}
			rec := e.rufe(wsRouten()[rDelete].gut())
			soll := `{"blobs_removed":` + strconv.Itoa(n) + `,"cleanup_deferred":` + strconv.FormatBool(z) + `}`
			if rec.Code != 200 || rec.Body.String() != soll {
				t.Errorf("Entfernt %d, zurückgestellt %v: %d %q, erwartet 200 und %q", n, z, rec.Code, rec.Body.String(), soll)
			}
			faelle++
		}
	}
	gleichAuf(t, "Fälle", faelle, 12)
}

// wsWorkspaceLit ist die Antwort eines Workspace als Literal, aus den Feldern gesetzt (UTC, neun Stellen).
func wsWorkspaceLit(w model.Workspace) string {
	return fmt.Sprintf(`{"id":%q,"service_id":%q,"baseline_snapshot_id":%q,"revision":%d,"created_at":%q}`,
		w.ID, w.ServiceID, w.BaselineSnapshotID, w.Revision, w.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000000")+"Z")
}

// wsSnapshotLit ist die Antwort eines Snapshots als Literal.
func wsSnapshotLit(s model.WorkspaceSnapshot) string {
	return fmt.Sprintf(`{"id":%q,"workspace_id":%q,"name":%q,"created_at":%q}`,
		s.ID, s.WorkspaceID, s.Name, s.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000000")+"Z")
}

// TestWS12Listen hält WS12 (SPEC-298 Nr. 7 d): die Listen tragen die Elemente in der Reihenfolge des Dienstes, ohne
// Neuordnung, Entdoppelung und Kappung, und sind nie `null`: M-Liste, 7 Formen (`nil`, leer, 1, 2, 3 und 100 Elemente in
// absteigender Folge von Kennung und Zeit, zwei gleiche) × 2 Routen = 14.
func TestWS12Listen(t *testing.T) {
	wsFrist(t)
	absteigend := func(n int) []int {
		out := make([]int, n)
		for i := range out {
			out[i] = n - i
		}
		return out
	}
	formen := []struct {
		name string
		ids  []int
	}{{"nil", nil}, {"leer", []int{}}, {"1", absteigend(1)}, {"2", absteigend(2)}, {"3", absteigend(3)}, {"100", absteigend(100)}, {"gleiche", []int{5, 5}}}
	faelle := 0
	for _, f := range formen {
		var ws []model.Workspace
		var ss []model.WorkspaceSnapshot
		var wl, sl []string
		if f.ids != nil {
			ws, ss, wl, sl = []model.Workspace{}, []model.WorkspaceSnapshot{}, []string{}, []string{}
		}
		for _, n := range f.ids {
			z := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Hour)
			w := model.Workspace{ID: "w-" + strconv.Itoa(n), ServiceID: "svc", BaselineSnapshotID: "b", Revision: int64(n), CreatedAt: z}
			s := model.WorkspaceSnapshot{ID: "s-" + strconv.Itoa(n), ServiceID: "svc", WorkspaceID: "w", Name: "n-" + strconv.Itoa(n), CreatedAt: z}
			ws, ss, wl, sl = append(ws, w), append(ss, s), append(wl, wsWorkspaceLit(w)), append(sl, wsSnapshotLit(s))
		}
		e := wsNeu(t)
		e.wf.liste, e.sf.liste = ws, ss
		for _, c := range []struct {
			r    wsRoute
			soll string
		}{
			{wsRouten()[rList], `{"workspaces":[` + strings.Join(wl, ",") + `]}`},
			{wsRouten()[rSnList], `{"snapshots":[` + strings.Join(sl, ",") + `]}`},
		} {
			rec := e.rufe(c.r.gut())
			faelle++
			if rec.Code != 200 || rec.Body.String() != c.soll {
				t.Errorf("%s %s: %d %q, erwartet %q", c.r.name, f.name, rec.Code, rec.Body.String(), c.soll)
			}
		}
	}
	gleichAuf(t, "Fälle", []int{faelle, len(formen)}, []int{14, 7})
}

// wsRahmenAntwort ruft den Server mit dem Request und liefert die Antwort und den Wert einer Panik (http.ErrAbortHandler: der
// Server bricht die Verbindung ab).
func wsRahmenAntwort(e *wsEnv, r *http.Request) (*httptest.ResponseRecorder, any) {
	rec := httptest.NewRecorder()
	p := abbruchVon(func() { e.srv.ServeHTTP(rec, r) })
	return rec, p
}

// TestWS13Fehler hält WS13 (SPEC-298 Nr. 8; SPEC-238 Nr. 6, 15 a): M-Fehler, 125 Zellen (8 × 11 allgemeine Ausgänge, 28 Codes
// der Tabelle, 9 der Achse Lesen) mit Status, Code und Body als Literale, genau einem Zählen je Fehlerantwort, keinem bei einem
// Kontextfehler (Verbindung abgebrochen, nichts geschrieben), `FakeAudit.Aufrufe()` gleich 0 und genau einem Aufruf der
// Operation der Route (bei der Achse Lesen keinem).
func TestWS13Fehler(t *testing.T) {
	wsFrist(t)
	zellen := wsFehlerZellen()
	gruppen := map[string]int{}
	for _, z := range zellen {
		r := wsRouten()[z.route]
		name := r.name + " " + z.name
		e := wsNeu(t)
		e.wf.liste, e.sf.liste = []model.Workspace{wsA()}, []model.WorkspaceSnapshot{snapA()}
		req := r.gut()
		if z.lesen {
			req = r.request(wsPfadID, "", nil, errLeser{z.err})
			gruppen["Lesen"]++
		} else {
			e.wf.err, e.sf.err = z.err, z.err
			gruppen["Dienst"]++
		}
		rec, abbruch := wsRahmenAntwort(e, req)
		codes, andere := zaehlungenVon(e.tele)
		var soll []string
		if z.abbruch {
			if abbruch != http.ErrAbortHandler || rec.Body.Len() != 0 || rec.Result().StatusCode != 200 {
				t.Errorf("%s: Panik %v, Body %q: erwartet den Abbruch der Verbindung ohne Antwort", name, abbruch, rec.Body.String())
			}
		} else {
			soll = []string{z.code}
			if abbruch != nil {
				t.Errorf("%s: Panik %v", name, abbruch)
			}
			pruefeFehlerAntwort(t, name, rec, z.status, z.objekt, "")
		}
		if !slices.Equal(codes, soll) || andere != 0 {
			t.Errorf("%s: CountError %q, andere Aufrufe %d, erwartet %q", name, codes, andere, soll)
		}
		if n := e.audit.Aufrufe(); n != 0 {
			t.Errorf("%s: %d Aufrufe von Audit, erwartet 0 (jede Zelle ist ein Fehlerausgang)", name, n)
		}
		if z.lesen {
			if n := e.p.gesamt(); n != 0 {
				t.Errorf("%s: %d Aufrufe der Dienste, erwartet 0", name, n)
			}
			continue
		}
		e.p.nurDiagonale(t, name, r.op)
	}
	gleichAuf(t, "Zellen", len(zellen), 125)
	gleichAuf(t, "Zellen gleich 8 × 11 + 28 + 9", len(zellen), 8*11+(2+4+3+3+3+3+4+6)+9)
	gleichAuf(t, "Gruppen", gruppen, map[string]int{"Dienst": 116, "Lesen": 9})
}

// TestWS13Zahlen hält M-Fehler gegen die Quelle: jede Zeile der Tabelle von SPEC-298 Nr. 8 hat ihre Zellen (2, 4, 3, 3, 3, 3,
// 4 und 6), jeder Code steht in model.AlleCodes() mit dem Status des Literals, und das Literal trägt die message des Codes.
func TestWS13Zahlen(t *testing.T) {
	wsFrist(t)
	je := []int{2, 4, 3, 3, 3, 3, 4, 6}
	alle := model.AlleCodes()
	for r := range 8 {
		gleichAuf(t, "Codes der Route "+strconv.Itoa(r), len(wsCodes(r)), je[r])
		gleichAuf(t, "allgemeine Ausgänge der Route "+strconv.Itoa(r), len(wsAllgemein(r)), 11)
		for _, f := range wsCodes(r) {
			fe := f.err.(*inbound.Fehler)
			c, bekannt := model.ParseCode(f.code)
			if !bekannt || !slices.Contains(alle, c) || c != fe.Code || !c.StatusErlaubt(f.status) || fe.Status != f.status {
				t.Errorf("Route %d %s: Code %q, Status %d passt nicht zur Tabelle der Codes", r, f.name, f.code, f.status)
			}
			if !strings.Contains(f.objekt, `"message":"`+c.Message()+`"`) {
				t.Errorf("Route %d %s: das Literal %q nennt die message %q nicht", r, f.name, f.objekt, c.Message())
			}
		}
	}
}

// TestWS13Unveraendert hält SPEC-298 Nr. 8 an der Route allein: der Handler gibt jeden Fehler des Dienstes unverändert zurück
// (dasselbe Objekt, nicht verpackt) und keine Antwort daneben, auch wenn der Dienst einen Wert und den Fehler liefert; ein
// Kontextfehler beim Lesen des Bodys kommt ebenso unverändert zurück, ohne Aufruf eines Dienstes.
func TestWS13Unveraendert(t *testing.T) {
	wsFrist(t)
	zellen := 0
	for _, z := range wsFehlerZellen() {
		r := wsRouten()[z.route]
		e := wsNeu(t)
		e.wf.liste, e.sf.liste = []model.Workspace{wsA()}, []model.WorkspaceSnapshot{snapA()}
		a := admin.Anfrage{Methode: r.methode, Body: strings.NewReader(r.body), Pfadwerte: map[string]string{"id": wsPfadID}}
		if z.lesen {
			a.Body = errLeser{z.err}
		} else {
			e.wf.err, e.sf.err = z.err, z.err
		}
		ant, err := e.routen[z.route].Handler(context.Background(), a)
		zellen++
		if err != z.err || ant.Status != 0 || len(ant.Body) != 0 {
			t.Errorf("%s %s: Handler liefert %v und Antwort %d %q, erwartet denselben Fehler und keine Antwort", r.name, z.name, err, ant.Status, ant.Body)
		}
		if z.lesen && e.p.gesamt() != 0 {
			t.Errorf("%s %s: %d Aufrufe der Dienste, erwartet 0", r.name, z.name, e.p.gesamt())
		}
	}
	gleichAuf(t, "Zellen", zellen, 125)
}
