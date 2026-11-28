package bootstrap_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Werte, die in keiner Zeile der Senke stehen dürfen (LH-QA-04.a): je Modus und Ort (Literal der Query, Body des Requests, Body der
// Antwort) ein eigener, dazu der Body und das Literal der Query zweier Requests, die der Replay nicht findet. Jeder erfüllt die Form Token,
// so dass ihn nicht die Formprüfung des Adapters fernhält.
const (
	recQuery, recRequest, recResponse    = "canary_rec_query", "canary_rec_request", "canary_rec_response"
	repQuery, repRequest, repResponse    = "canary_rep_query", "canary_rep_request", "canary_rep_response"
	repFehlrumpf, repFehlquery           = "canary_rep_miss_body", "canary_rep_miss_query"
	mockQuery, mockRequest, mockResponse = "canary_mock_query", "canary_mock_request", "canary_mock_response"
)

// betriebSpan ist ein Span eines Dienstes: Name und seine Pflicht-Attribute mit Wert.
type betriebSpan struct {
	name  string
	attrs map[string]any
}

// betriebAnfrage ist eine Anfrage an den OData-Listener mit der erwarteten Antwort: Status, Code des Fehlerobjekts (leer bei Erfolg), ein
// Stück, das der Body der Antwort trägt, und eines, das er nicht trägt (je leer: nicht geprüft), und der Span des Dienstes unter ihrem
// Gesamtspan (nil: keiner).
type betriebAnfrage struct {
	methode, pfad, rumpf string
	kopf                 map[string]string
	vorher               func()
	status               int
	code, antwort, ohne  string
	dienst               *betriebSpan
}

// betriebLauf sendet die Anfragen nacheinander an den Listener des Prozesses, fährt ihn herunter (erst danach ist jeder Span beendet) und
// prüft Antworten, Spans, Zähler, Werte in der Senke und Verwürfe. zaehler sind die Codes der Fehlerobjekte, die der Client erhalten muss;
// namen sammelt die Namen der gesehenen Spans.
func betriebLauf(t *testing.T, modus string, w *wurzel, anfragen []betriebAnfrage, zaehler map[string]uint64, namen map[string]bool) {
	t.Helper()
	erhalten := map[string]uint64{}
	for i, a := range anfragen {
		if a.vorher != nil {
			a.vorher()
		}
		status, body := betriebRuf(t, w.h.Adresse(), a)
		if status != a.status || !strings.Contains(body, a.antwort) || a.ohne != "" && strings.Contains(body, a.ohne) {
			t.Errorf("%s Anfrage %d (%s %s): Status %d, Body %q; erwartet %d mit %q ohne %q", modus, i, a.methode, a.pfad, status, body,
				a.status, a.antwort, a.ohne)
		}
		var objekt struct {
			Error struct{ Code string } `json:"error"`
		}
		if status >= 400 {
			if err := json.Unmarshal([]byte(body), &objekt); err != nil {
				t.Fatalf("%s Anfrage %d: kein Fehlerobjekt: %q", modus, i, body)
			}
			erhalten[objekt.Error.Code]++
		}
		if objekt.Error.Code != a.code {
			t.Errorf("%s Anfrage %d: Code %q, erwartet %q", modus, i, objekt.Error.Code, a.code)
		}
	}
	if err := w.h.Herunterfahren(context.Background()); err != nil {
		t.Fatalf("%s: Herunterfahren: %v", modus, err)
	}
	betriebSpans(t, modus, w.senke.String(), anfragen, namen)
	gezaehlt := map[string]uint64{}
	for c, n := range w.adapter[0].Snapshot().Errors {
		if n != 0 {
			gezaehlt[c.String()] = n
		}
	}
	if !reflect.DeepEqual(erhalten, zaehler) || !reflect.DeepEqual(gezaehlt, erhalten) {
		t.Errorf("%s: Zähler %v, der Client erhielt %v, erwartet %v", modus, gezaehlt, erhalten, zaehler)
	}
	senke := strings.ToLower(w.senke.String())
	for _, wert := range betriebWerte() {
		if strings.Contains(senke, wert) {
			t.Errorf("%s: die Senke trägt %q:\n%s", modus, wert, w.senke.String())
		}
	}
	keinVerwurf(t, modus, w.adapter[0])
}

func betriebWerte() []string {
	return []string{recQuery, recRequest, recResponse, repQuery, repRequest, repResponse, repFehlrumpf, repFehlquery,
		mockQuery, mockRequest, mockResponse}
}

// betriebRuf sendet die Anfrage an adr und liefert Status und Body.
func betriebRuf(t *testing.T, adr string, a betriebAnfrage) (int, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), a.methode, "http://"+adr+a.pfad, strings.NewReader(a.rumpf))
	if err != nil {
		t.Fatal(err)
	}
	if a.rumpf != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range a.kopf {
		req.Header.Set(k, v)
	}
	resp, err := gwClient().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", a.methode, a.pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: %v", a.methode, a.pfad, err)
	}
	return resp.StatusCode, string(body)
}

// betriebSpans hält die Zeilen der Spans: je Anfrage in der Folge der Anfragen ein Gesamtspan mit Methode, Status und Modus, darunter
// (parent_id) genau der Span des Dienstes der Anfrage mit genau seinen Attributen, und sonst keine Zeile eines Spans.
func betriebSpans(t *testing.T, modus, senke string, anfragen []betriebAnfrage, namen map[string]bool) {
	t.Helper()
	standard := map[string]bool{"timestamp": true, "level": true, "msg": true, "ereignis": true, "span_id": true, "parent_id": true, "duration_ms": true}
	lies := func(z map[string]any) betriebSpan {
		s := betriebSpan{attrs: map[string]any{}}
		s.name, _ = z["msg"].(string)
		for k, v := range z {
			if !standard[k] {
				s.attrs[k] = v
			}
		}
		namen[s.name] = true
		return s
	}
	var gesamt, andere []map[string]any
	for _, z := range zeilen(t, senke) {
		switch {
		case z["ereignis"] != "span":
		case z["msg"] == "odata.gateway.request":
			gesamt = append(gesamt, z)
		default:
			andere = append(andere, z)
		}
	}
	if len(gesamt) != len(anfragen) {
		t.Fatalf("%s: %d Gesamtspans, erwartet %d:\n%s", modus, len(gesamt), len(anfragen), senke)
	}
	dienste := 0
	for i, a := range anfragen {
		soll := betriebSpan{"odata.gateway.request", map[string]any{
			"http.request.method": a.methode, "http.response.status_code": float64(a.status), "recorder.mode": modus,
		}}
		if got := lies(gesamt[i]); !reflect.DeepEqual(got, soll) {
			t.Errorf("%s Anfrage %d: Gesamtspan %v, erwartet %v", modus, i, got, soll)
		}
		var kinder []betriebSpan
		for _, z := range andere {
			if z["parent_id"] == gesamt[i]["span_id"] {
				kinder = append(kinder, lies(z))
			}
		}
		var erwartet []betriebSpan
		if a.dienst != nil {
			dienste++
			erwartet = []betriebSpan{*a.dienst}
		}
		if !reflect.DeepEqual(kinder, erwartet) {
			t.Errorf("%s Anfrage %d (%s %s): Spans unter dem Gesamtspan %v, erwartet %v", modus, i, a.methode, a.pfad, kinder, erwartet)
		}
	}
	if len(andere) != dienste {
		t.Errorf("%s: %d Zeilen der Spans der Dienste, erwartet %d:\n%s", modus, len(andere), dienste, senke)
	}
}

// betriebUpstream ist ein Upstream, der je Request Query und Body festhält und mit einem eigenen Body antwortet.
type betriebUpstream struct {
	mu      sync.Mutex
	gesehen []string
}

func (u *betriebUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	u.mu.Lock()
	u.gesehen = append(u.gesehen, r.URL.RawQuery+" "+string(body))
	u.mu.Unlock()
	antwort := recResponse
	if strings.Contains(r.URL.RawQuery, repQuery) {
		antwort = repResponse
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"Name":"` + antwort + `"}`))
}

// TestBetriebSpansUndZaehler fährt über den Composition Root die Modi record, replay und mock mit Anfragen am OData-Listener und hält je Modus:
// den Gesamtspan jeder Anfrage und die Spans der Dienste (odata.record.capture, odata.replay.match, odata.mock.query,
// odata.workspace.mutation) mit genau ihren Pflicht-Attributen (SPEC-026 bis SPEC-030); den Stand der Zähler je Fehlercode gleich den Codes
// der Fehlerobjekte, die der Client erhielt, mit Anzahl (SPEC-230 Nr. 8, SPEC-294 Nr. 5); keinen der Werte aus Query, Request-Body und
// Antwort-Body in einer Zeile der Senke und keinen Verwurf (LH-QA-04.a). Gesucht wird der ganze Wert in jeder Groß- und Kleinschreibung,
// kein Teilstück und keine Kodierung. Dass die Werte den Prozess erreichen, belegen: in record der Upstream und die Antwort an den Client;
// in replay die Antwort zum Treffer und der fehlende Treffer, wenn allein der Body oder allein das Literal der Query abweicht; in mock die
// Entity, die der Filter allein findet, und die Antwort an den Client.
func TestBetriebSpansUndZaehler(t *testing.T) {
	token := regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	verschieden := map[string]bool{}
	for _, wert := range betriebWerte() {
		verschieden[wert] = token.MatchString(wert)
		if !verschieden[wert] {
			t.Fatalf("%q erfüllt die Form Token nicht", wert)
		}
	}
	if len(verschieden) != 11 {
		t.Fatalf("%d verschiedene Werte, erwartet 11", len(verschieden))
	}
	namen := map[string]bool{}
	query := func(wert string) string { return "/Products?$filter=Name%20eq%20'" + wert + "'" }
	rumpf := func(wert string) string { return `{"Name":"` + wert + `"}` }

	// record: zwei Exchanges (der zweite ist die Aufzeichnung des Replay), eine unbekannte Session und ein beendeter Upstream.
	up := &betriebUpstream{}
	upServer := httptest.NewServer(up)
	defer upServer.Close()
	dir := mockDaten(t)
	w := baueWurzelMit(t, modusRecord, nil, dir, func(k *bootstrap.Konfiguration) { k.Dienste[0].UpstreamURL = upServer.URL })
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("record: Hochfahren: %v", err)
	}
	sessions, err := w.h.Sessions().List(context.Background())
	var offen []string
	for _, s := range sessions {
		if s.Status == model.SessionOpen {
			offen = append(offen, s.ID)
		}
	}
	if err != nil || len(offen) != 1 {
		t.Fatalf("record: offene Sessions %v, %v; erwartet eine", offen, err)
	}
	session := offen[0]
	capture := func(status string) *betriebSpan {
		return &betriebSpan{"odata.record.capture", map[string]any{"recorder.session_id": session, "recorder.capture_status": status}}
	}
	betriebLauf(t, modusRecord, w, []betriebAnfrage{
		{methode: "POST", pfad: query(recQuery), rumpf: rumpf(recRequest), status: 200, antwort: recResponse, dienst: capture("complete")},
		{methode: "POST", pfad: query(repQuery), rumpf: rumpf(repRequest), status: 200, antwort: repResponse, dienst: capture("complete")},
		{methode: "GET", pfad: "/Products", kopf: map[string]string{"X-Recorder-Recording-Session": "ses-unbekannt"}, status: 404, code: "ODR-E7003"},
		{methode: "GET", pfad: "/Products", vorher: upServer.Close, status: 502, code: "ODR-E4001", dienst: capture("incomplete")},
	}, map[string]uint64{"ODR-E7003": 1, "ODR-E4001": 1}, namen)
	if len(up.gesehen) != 2 || !strings.Contains(up.gesehen[0], recQuery) || !strings.Contains(up.gesehen[0], recRequest) ||
		!strings.Contains(up.gesehen[1], repQuery) || !strings.Contains(up.gesehen[1], repRequest) {
		t.Errorf("record: der Upstream sah %q; erwartet je Exchange Query und Body", up.gesehen)
	}

	// replay: über der Aufzeichnung; die Strategie weicht vom Default ab. Ein anderer Body unter derselben Query und eine andere Query mit
	// demselben Body finden nichts.
	w = baueWurzelMit(t, modusReplay, nil, dir, func(k *bootstrap.Konfiguration) { k.ReplayStrategy = "last" })
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("replay: Hochfahren: %v", err)
	}
	zuordnung := map[string]string{"X-Recorder-Workspace": "ws-1", "X-Recorder-Recording-Session": session, "X-Recorder-Replay-Session": "rep-1"}
	match := func(ergebnis string) *betriebSpan {
		return &betriebSpan{"odata.replay.match", map[string]any{"recorder.match_strategy": "last", "recorder.match_result": ergebnis}}
	}
	fehltreffer := func(q, body string) betriebAnfrage {
		return betriebAnfrage{methode: "POST", pfad: query(q), rumpf: rumpf(body), kopf: zuordnung,
			status: 404, code: "ODR-E5001", dienst: match("kein_treffer")}
	}
	betriebLauf(t, modusReplay, w, []betriebAnfrage{
		{methode: "POST", pfad: query(repQuery), rumpf: rumpf(repRequest), kopf: zuordnung, status: 200, antwort: repResponse, dienst: match("treffer")},
		fehltreffer(repQuery, repFehlrumpf),
		fehltreffer(repFehlquery, repRequest),
		{methode: "GET", pfad: "/Products", status: 400, code: "ODR-E5003"},
	}, map[string]uint64{"ODR-E5001": 2, "ODR-E5003": 1}, namen)

	// mock: die Entities mit den Werten für Filter und Antwort legt ein eigener Prozess an, so dass im geprüften Prozess kein Body eines
	// Requests den Wert der Query oder der Antwort trägt.
	dir = mockDaten(t)
	ws := map[string]string{"X-Recorder-Workspace": "ws-1"}
	w = baueWurzelMit(t, modusMock, nil, dir, nil)
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("mock: Hochfahren: %v", err)
	}
	for _, entity := range []string{`{"Id":10,"Name":"` + mockQuery + `"}`, `{"Id":11,"Name":"` + mockResponse + `"}`} {
		if status, body := betriebRuf(t, w.h.Adresse(), betriebAnfrage{methode: "POST", pfad: "/Products", rumpf: entity, kopf: ws}); status != 201 {
			t.Fatalf("mock: Anlegen: Status %d, Body %q", status, body)
		}
	}
	if err := w.h.Herunterfahren(context.Background()); err != nil {
		t.Fatalf("mock: Herunterfahren nach dem Anlegen: %v", err)
	}
	w = baueWurzelMit(t, modusMock, nil, dir, nil)
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("mock: Hochfahren: %v", err)
	}
	abfrage := func(set, coverage string) *betriebSpan {
		return &betriebSpan{"odata.mock.query", map[string]any{"recorder.entity_set": set, "recorder.coverage_status": coverage}}
	}
	mutation := func(art string) *betriebSpan {
		return &betriebSpan{"odata.workspace.mutation", map[string]any{"recorder.workspace_id": "ws-1", "recorder.mutation_type": art}}
	}
	unbekannt := betriebAnfrage{methode: "GET", pfad: "/Products(99)", kopf: ws, status: 404, code: "ODR-E6008",
		dienst: abfrage("Products", "complete")}
	betriebLauf(t, modusMock, w, []betriebAnfrage{
		{methode: "GET", pfad: query(mockQuery), kopf: ws, status: 200, antwort: `"Id":10`, ohne: mockResponse, dienst: abfrage("Products", "complete")},
		{methode: "GET", pfad: "/Products?$filter=Id%20eq%2011", kopf: ws, status: 200, antwort: mockResponse, dienst: abfrage("Products", "complete")},
		{methode: "GET", pfad: "/Orders", kopf: ws, status: 200, dienst: abfrage("Orders", "partial")},
		{methode: "POST", pfad: "/Products", rumpf: `{"Id":12,"Name":"` + mockRequest + `"}`, kopf: ws, status: 201, antwort: mockRequest,
			dienst: mutation("post")},
		{methode: "PATCH", pfad: "/Products(12)", rumpf: rumpf(mockRequest), kopf: ws, status: 204, dienst: mutation("patch")},
		{methode: "PUT", pfad: "/Products(12)", rumpf: `{"Id":12,"Name":"` + mockRequest + `"}`, kopf: ws, status: 204, dienst: mutation("put")},
		{methode: "DELETE", pfad: "/Products(12)", kopf: ws, status: 204, dienst: mutation("delete")},
		unbekannt,
		unbekannt,
		{methode: "GET", pfad: "/Products", status: 400, code: "ODR-E5003"},
	}, map[string]uint64{"ODR-E6008": 2, "ODR-E5003": 1}, namen)

	if soll := map[string]bool{"odata.gateway.request": true, "odata.record.capture": true, "odata.replay.match": true,
		"odata.mock.query": true, "odata.workspace.mutation": true}; !reflect.DeepEqual(namen, soll) {
		t.Errorf("gesehene Spans %v, erwartet die fünf %v", namen, soll)
	}
}
