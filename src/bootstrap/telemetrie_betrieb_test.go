package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
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

// betriebWeg ist ein Weg eines Modus, auf dem der Dienst ein Fehlerobjekt liefert oder einen Body verwirft. pfad (mit Query) und rumpf sind
// Muster mit den Orten {q} (Literal der Query), {k} (Literal des Schlüssels im Pfad), {n} (der Schlüssel als Zahl, im Pfad und im Body),
// {m} (eine zweite Zahl im Body) und {b} (Text im Body). kopf ersetzt Header des Modus, ein Wert nil entfernt einen. status und code sind
// die erwartete Antwort, code leer heißt ohne Fehlerobjekt; eintrag ist ein Stück, das die Senke dann tragen muss. satz wählt die Werte
// (0: die Nummer des Wegs). gespeichert legt vorher eine Entity {n} mit dem Namen {g} an, geloescht löscht sie danach.
type betriebWeg struct {
	name, methode, pfad, rumpf string
	kopf                       http.Header
	status                     int
	code, eintrag              string
	satz                       int
	gespeichert, geloescht     bool
}

// Die Muster eines Wegs, der alle drei Orte trägt, und der Wert, mit dem der Upstream in record und beim Aufzeichnen antwortet.
const (
	wegPfad    = "/Products('{k}')?$filter=Name%20eq%20'{q}'"
	wegKurz    = "/Products('{k}')?x={q}"
	wegZahl    = "/Products({n})?x={q}"
	wegRumpf   = `{"Name":"{b}"}`
	wegAntwort = "cf_rec_up_r"
	wegRecord  = "record: Aufzeichnung eines Bodys verworfen"
)

// betriebOrte sind die Orte der Muster je Platzhalter.
func betriebOrte() []string { return []string{"q", "k", "n", "m", "b", "g"} }

// betriebWert ist der Wert am Ort des Satzes im Modus; {n} und {m} sind Zahlen im Bereich von Edm.Int32, die übrigen erfüllen die Form Token.
func betriebWert(modus string, satz int, ort string) string {
	nummer := map[string]int{modusRecord: 1, modusReplay: 2, modusMock: 3}[modus]
	switch ort {
	case "n":
		return fmt.Sprintf("7%d%02d00001", nummer, satz)
	case "m":
		return fmt.Sprintf("7%d%02d00002", nummer, satz)
	}
	return fmt.Sprintf("cf_%s_%02d_%s", modus[:3], satz, ort)
}

// betriebSetze ersetzt die Orte im Muster und hängt die ersetzten Werte an werte an.
func betriebSetze(modus string, satz int, muster string, werte *[]string) string {
	for _, ort := range betriebOrte() {
		if platz := "{" + ort + "}"; strings.Contains(muster, platz) {
			wert := betriebWert(modus, satz, ort)
			muster = strings.ReplaceAll(muster, platz, wert)
			*werte = append(*werte, wert)
		}
	}
	return muster
}

// betriebEingang hält Pfad, Query und Body der letzten Anfrage, die der Dienst des Modus erhielt.
type betriebEingang struct {
	mu     sync.Mutex
	letzte string
}

// haken legt sich zwischen den Listener und den Dienst.
func (e *betriebEingang) haken(b *bootstrap.Bauteile) {
	neu := b.NeuerServer
	b.NeuerServer = func(k httpodata.Konfig) (bootstrap.Server, error) {
		innen := k.Verarbeiten
		k.Verarbeiten = func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
			var body []byte
			if a.Body != nil {
				body, _ = io.ReadAll(a.Body)
				a.Body = bytes.NewReader(body)
			}
			e.mu.Lock()
			e.letzte = a.Pfad + "?" + a.RawQuery + " " + string(body)
			e.mu.Unlock()
			return innen(ctx, a)
		}
		return neu(k)
	}
}

func (e *betriebEingang) lies() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.letzte
}

// betriebWegUpstream ist der Upstream der Wege: der Header Betrieb-Upstream wählt den Fehler oder die nicht entpackbare Antwort; außer bei
// `ungelesen` liest er den Body des Requests.
type betriebWegUpstream struct{}

func (betriebWegUpstream) Senden(_ context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	art := http.Header(a.Header).Get("Betrieb-Upstream")
	if art != "ungelesen" && a.Body != nil {
		_, _ = io.Copy(io.Discard, a.Body)
	}
	switch art {
	case "nicht_erreichbar", "ungelesen":
		return outbound.UpstreamAntwort{}, outbound.ErrUpstreamNichtErreichbar
	case "zeit":
		return outbound.UpstreamAntwort{}, outbound.ErrUpstreamZeit
	case "intern":
		return outbound.UpstreamAntwort{}, errors.New("betrieb: Upstream")
	}
	kopf := map[string][]string{"Content-Type": {"application/json"}}
	if art == "kodierung" {
		kopf["Content-Encoding"] = []string{"gzip"}
	}
	return outbound.UpstreamAntwort{Status: 200, Header: kopf, Body: io.NopCloser(strings.NewReader(`{"Name":"` + wegAntwort + `"}`))}, nil
}

// betriebProzess baut den Prozess des Modus mit zwei Haken (Eingang des Dienstes, Upstream der Wege) und fährt ihn hoch.
func betriebProzess(t *testing.T, modus, dir string, aendere func(*bootstrap.Konfiguration)) (*wurzel, *betriebEingang) {
	t.Helper()
	ein := &betriebEingang{}
	w := baueWurzelHaken(t, modus, nil, dir, aendere, func(b *bootstrap.Bauteile) {
		ein.haken(b)
		b.NeuerUpstream = func(httpupstream.Konfig) (outbound.UpstreamClient, error) { return betriebWegUpstream{}, nil }
	})
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("%s: Hochfahren: %v", modus, err)
	}
	return w, ein
}

// betriebDienste setzt die Dienste mit den Roots; der Upstream wird nie angewählt (Haken).
func betriebDienste(roots ...string) func(*bootstrap.Konfiguration) {
	return func(k *bootstrap.Konfiguration) {
		namen := map[string]string{"/": "svc", "/odata": "svc", "/andere": "andere", "/doppelt": "doppelt"}
		k.Dienste = nil
		for _, root := range roots {
			d := bootstrap.Dienst{ID: namen[root], Root: root}
			if k.Modus == modusRecord {
				d.UpstreamURL = "http://upstream.invalid"
			}
			k.Dienste = append(k.Dienste, d)
		}
		k.MatchInvalidEncoding = "reject"
		k.EtagEnabled = true
	}
}

// betriebSende sendet die Anfrage mit genau den Headern und liefert Status und Body.
func betriebSende(t *testing.T, adr, methode, pfad, rumpf string, kopf http.Header) (int, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), methode, "http://"+adr+pfad, strings.NewReader(rumpf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header = kopf
	resp, err := gwClient().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", methode, pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: %v", methode, pfad, err)
	}
	return resp.StatusCode, string(body)
}

// betriebKopf sind die Header des Modus, darüber die des Wegs.
func betriebKopf(basis, weg http.Header, rumpf string) http.Header {
	kopf := basis.Clone()
	if kopf == nil {
		kopf = http.Header{}
	}
	if rumpf != "" {
		kopf.Set("Content-Type", "application/json")
	}
	for k, v := range weg {
		if v == nil {
			kopf.Del(k)
			continue
		}
		kopf[k] = v
	}
	return kopf
}

// betriebFahre sendet einen Weg und prüft Status, Code und dass jeder Wert seiner Muster beim Dienst ankam; es liefert diese Werte, bei
// gespeichert auch den der Entity.
func betriebFahre(t *testing.T, modus string, w *wurzel, ein *betriebEingang, basis http.Header, weg betriebWeg) []string {
	t.Helper()
	var werte []string
	pfad := betriebSetze(modus, weg.satz, weg.pfad, &werte)
	rumpf := betriebSetze(modus, weg.satz, weg.rumpf, &werte)
	status, body := betriebSende(t, w.h.Adresse(), weg.methode, pfad, rumpf, betriebKopf(basis, weg.kopf, rumpf))
	var objekt struct {
		Error struct{ Code string } `json:"error"`
	}
	if status >= 400 {
		if err := json.Unmarshal([]byte(body), &objekt); err != nil {
			t.Fatalf("%s %s: kein Fehlerobjekt: %q", modus, weg.name, body)
		}
	}
	if status != weg.status || objekt.Error.Code != weg.code {
		t.Errorf("%s %s: Status %d, Code %q, erwartet %d %q: %q", modus, weg.name, status, objekt.Error.Code, weg.status, weg.code, body)
	}
	for _, wert := range werte {
		if gesehen := ein.lies(); !strings.Contains(gesehen, wert) {
			t.Errorf("%s %s: der Wert %q kam nicht beim Dienst an: %.200q", modus, weg.name, wert, gesehen)
		}
	}
	if weg.gespeichert {
		werte = append(werte, betriebWert(modus, weg.satz, "g"))
	}
	return werte
}

// betriebWege fährt die Wege nacheinander, fährt den Prozess herunter und prüft: keiner ihrer Werte (und der Antwort des Upstream) in der
// Senke, jeder Eintrag der Wege in der Senke, kein Verwurf.
func betriebWege(t *testing.T, modus string, w *wurzel, ein *betriebEingang, basis http.Header, wege []betriebWeg) {
	t.Helper()
	werte := []string{wegAntwort}
	for i, weg := range wege {
		if weg.satz == 0 {
			weg.satz = i + 1
		}
		werte = append(werte, betriebFahre(t, modus, w, ein, basis, weg)...)
	}
	if err := w.h.Herunterfahren(context.Background()); err != nil {
		t.Fatalf("%s: Herunterfahren: %v", modus, err)
	}
	senke := strings.ToLower(w.senke.String())
	for _, wert := range werte {
		if strings.Contains(senke, wert) {
			t.Errorf("%s: die Senke trägt %q:\n%s", modus, wert, w.senke.String())
		}
	}
	for _, weg := range wege {
		if !strings.Contains(w.senke.String(), weg.eintrag) {
			t.Errorf("%s %s: die Senke trägt den Eintrag %q nicht:\n%s", modus, weg.name, weg.eintrag, w.senke.String())
		}
	}
	keinVerwurf(t, modus, w.adapter[0])
}

// betriebOhneDienst fährt den Weg ohne Service für den Pfad über einem leeren DATA_DIR.
func betriebOhneDienst(t *testing.T, modus string) {
	t.Helper()
	w, ein := betriebProzess(t, modus, t.TempDir(), betriebDienste("/odata"))
	betriebWege(t, modus, w, ein, nil, []betriebWeg{
		{name: "kein_dienst", methode: "POST", pfad: "/draussen" + wegKurz, rumpf: wegRumpf, satz: 99, status: 404, code: "ODR-E5005"},
	})
}

// betriebSessions sind die Kennungen der Sessions je Service, die offen sind, wenn nurOffen gilt.
func betriebSessions(t *testing.T, w *wurzel, nurOffen bool) map[string]string {
	t.Helper()
	liste, err := w.h.Sessions().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, s := range liste {
		if !nurOffen || s.Status == model.SessionOpen {
			ids[s.ServiceID] = s.ID
		}
	}
	return ids
}

// betriebRecordWege sind die Wege von record: die Prüfungen vor der Annahme, die Annahme, der Upstream ohne Antwort und die verworfenen
// Bodies.
func betriebRecordWege() []betriebWeg {
	weg := func(name, pfad string, status int, code string, kopf http.Header) betriebWeg {
		return betriebWeg{name: name, methode: "POST", pfad: pfad, rumpf: wegRumpf, status: status, code: code, kopf: kopf}
	}
	session := func(ids ...string) http.Header { return http.Header{"X-Recorder-Recording-Session": ids} }
	upstream := func(art string) http.Header { return http.Header{"Betrieb-Upstream": {art}} }
	wege := []betriebWeg{
		weg("kodierung", wegPfad+"&x=%zz", 400, "ODR-E5006", nil),
		weg("zwei_session_header", wegPfad, 400, "ODR-E5003", session("ses-1", "ses-1")),
		weg("unbekannte_session", wegPfad, 404, "ODR-E7003", session("ses-unbekannt")),
		weg("fremde_session", "/andere"+wegPfad, 409, "ODR-E5004", session("ses-1")),
		weg("geschlossene_session", wegPfad, 409, "ODR-E5009", session("ses-1")),
		weg("keine_offene_session", "/andere"+wegPfad, 409, "ODR-E5009", nil),
		weg("mehrere_offene_sessions", "/doppelt"+wegPfad, 400, "ODR-E5003", nil),
		weg("nicht_erreichbar", wegPfad, 502, "ODR-E4001", upstream("nicht_erreichbar")),
		weg("zeit", wegPfad, 504, "ODR-E4001", upstream("zeit")),
		weg("intern", wegPfad, 500, "ODR-E9001", upstream("intern")),
		weg("body_ungelesen", wegPfad, 502, "ODR-E4001", upstream("ungelesen")),
		weg("antwort_verworfen", wegPfad, 200, "", http.Header{"Betrieb-Upstream": {"kodierung"}, "Accept-Encoding": {"identity"}}),
		weg("request_verworfen", wegPfad, 200, "", http.Header{"Content-Encoding": {"br"}}),
	}
	for i := 7; i < 11; i++ {
		wege[i].eintrag = "record: Upstream ohne Antwort"
	}
	wege[11].eintrag, wege[12].eintrag = wegRecord, wegRecord
	return wege
}

// betriebReplayWege sind die Wege von replay vor und in der Auswahl, ohne den der Strategie strict; die Aufzeichnung trägt den Treffer
// (Satz 9) und den nicht replayfähigen Exchange (Satz 10).
func betriebReplayWege(fremd string) []betriebWeg {
	weg := func(name, pfad string, status int, code string, kopf http.Header) betriebWeg {
		return betriebWeg{name: name, methode: "POST", pfad: pfad, rumpf: wegRumpf, status: status, code: code, kopf: kopf}
	}
	session := func(id string) http.Header { return http.Header{"X-Recorder-Recording-Session": {id}} }
	treffer := weg("treffer", wegPfad, 200, "", nil)
	erschoepft := weg("erschoepft", wegPfad, 410, "ODR-E5002", nil)
	unfaehig := weg("nicht_replayfaehig", wegPfad, 422, "ODR-E5008", nil)
	treffer.satz, erschoepft.satz, unfaehig.satz = 9, 9, 10
	zuGross := weg("body_zu_gross", wegPfad, 404, "ODR-E5001", nil)
	zuGross.rumpf = `{"Name":"{b}","x":"` + strings.Repeat("a", 1<<20) + `"}`
	return []betriebWeg{
		weg("ohne_zuordnung", wegPfad, 400, "ODR-E5003", http.Header{"X-Recorder-Replay-Session": nil}),
		weg("unbekannte_session", wegPfad, 404, "ODR-E7003", session("ses-unbekannt")),
		weg("fremde_session", "/andere"+wegPfad, 409, "ODR-E5004", nil),
		weg("offene_session", wegPfad, 409, "ODR-E7009", session("ses-offen")),
		weg("kodierung", wegPfad+"&x=%zz", 400, "ODR-E5006", nil),
		weg("content_encoding", wegPfad, 404, "ODR-E5001", http.Header{"Content-Encoding": {"gzip"}}),
		zuGross,
		weg("kein_treffer", wegPfad, 404, "ODR-E5001", nil),
		treffer, erschoepft, unfaehig,
		weg("unbekannter_workspace", wegPfad, 404, "ODR-E7001", http.Header{"X-Recorder-Workspace": {"ws-fremd"}}),
		weg("fremder_workspace", "/andere"+wegPfad, 409, "ODR-E5004", session(fremd)),
	}
}

// betriebMockWege sind die Wege von mock beim Lesen und beim Schreiben.
func betriebMockWege() []betriebWeg {
	weg := func(name, methode, pfad, rumpf string, status int, code string) betriebWeg {
		return betriebWeg{name: name, methode: methode, pfad: pfad, rumpf: rumpf, status: status, code: code}
	}
	ohne := weg("ohne_workspace", "GET", wegKurz, wegRumpf, 400, "ODR-E5003")
	ohne.kopf = http.Header{"X-Recorder-Workspace": nil}
	unbekannt := weg("unbekannter_workspace", "GET", wegKurz, wegRumpf, 404, "ODR-E7001")
	unbekannt.kopf = http.Header{"X-Recorder-Workspace": {"ws-fremd"}}
	existiert := weg("existiert", "POST", "/Products?x={q}", `{"Id":{n},"Name":"{b}"}`, 409, "ODR-E6009")
	existiert.gespeichert = true
	geloescht := weg("geloescht", "GET", wegZahl, wegRumpf, 404, "ODR-E6008")
	geloescht.gespeichert, geloescht.geloescht = true, true
	vorbedingung := weg("vorbedingung", "PATCH", wegZahl, wegRumpf, 412, "ODR-E6010")
	vorbedingung.gespeichert, vorbedingung.kopf = true, http.Header{"If-Match": {`"999"`}}
	zuGross := weg("body_zu_gross", "PATCH", wegZahl, "", 400, "ODR-E6007")
	zuGross.rumpf = `{"Name":"{b}` + strings.Repeat("a", 1<<20) + `"}`
	patchGeloescht := weg("patch_geloescht", "PATCH", wegZahl, wegRumpf, 404, "ODR-E6008")
	patchGeloescht.gespeichert, patchGeloescht.geloescht = true, true
	return []betriebWeg{
		ohne,
		weg("url_ungueltig", "GET", "/Products('{k}')?$top={q}", wegRumpf, 400, "ODR-E6005"),
		weg("url_nicht_unterstuetzt", "GET", "/Products('{k}')?$apply={q}", wegRumpf, 501, "ODR-E6001"),
		weg("methode", "OPTIONS", wegKurz, wegRumpf, 501, "ODR-E6002"),
		unbekannt,
		weg("fremder_workspace", "GET", "/andere"+wegKurz, wegRumpf, 409, "ODR-E5004"),
		weg("set_unbekannt", "GET", "/Nix('{k}')?x={q}", wegRumpf, 404, "ODR-E6008"),
		weg("schluessel_ungueltig", "GET", wegKurz, wegRumpf, 400, "ODR-E6005"),
		weg("filter_auf_entity", "GET", "/Products({n})?$filter=Name%20eq%20'{q}'", wegRumpf, 501, "ODR-E6001"),
		weg("entity_fehlt", "GET", wegZahl, wegRumpf, 404, "ODR-E6008"),
		weg("entity_ungedeckt", "GET", "/Orders({n})?x={q}", wegRumpf, 422, "ODR-E6003"),
		weg("filter_unbeobachtet", "GET", "/Orders?$filter=Note%20eq%20'{q}'", wegRumpf, 422, "ODR-E6003"),
		weg("filter_typ", "GET", "/Products?$filter=Id%20eq%20'{q}'", wegRumpf, 400, "ODR-E6005"),
		weg("filter_property", "GET", "/Products?$filter={q}%20eq%201", wegRumpf, 400, "ODR-E6005"),
		weg("post_auf_entity", "POST", wegKurz, wegRumpf, 501, "ODR-E6002"),
		weg("post_set_unbekannt", "POST", "/Nix?x={q}", wegRumpf, 404, "ODR-E6008"),
		weg("patch_schluessel_ungueltig", "PATCH", wegKurz, wegRumpf, 400, "ODR-E6005"),
		weg("body_unbekannte_property", "PATCH", wegZahl, `{"Unbekannt":"{b}"}`, 400, "ODR-E6007"),
		weg("body_reserviert", "PUT", wegZahl, `{"Id":{n},"Name":"[REDACTED]"}`, 400, "ODR-E6007"),
		weg("body_widerspruch", "PATCH", wegZahl, `{"Id":{m},"Name":"{b}"}`, 400, "ODR-E6007"),
		existiert,
		weg("patch_fehlt", "PATCH", wegZahl, wegRumpf, 404, "ODR-E6008"),
		weg("patch_ungedeckt", "PATCH", "/Orders({n})?x={q}", `{"Note":"{b}"}`, 422, "ODR-E6003"),
		geloescht,
		weg("entity_unvollstaendig", "GET", "/Orders(1)?x={q}", wegRumpf, 422, "ODR-E6003"),
		vorbedingung,
		zuGross,
		weg("filter_auf_mutation", "PATCH", "/Products({n})?$filter=Name%20eq%20'{q}'", wegRumpf, 501, "ODR-E6002"),
		patchGeloescht,
		weg("patch_unvollstaendig", "PATCH", "/Orders(1)?x={q}", `{"Note":"{b}"}`, 422, "ODR-E6003"),
	}
}

// betriebLege legt in einem eigenen Prozess die Entities der Wege mit gespeichert an und löscht die mit geloescht.
func betriebLege(t *testing.T, dir string, wege []betriebWeg) {
	t.Helper()
	w, _ := betriebProzess(t, modusMock, dir, betriebDienste("/", "/andere"))
	kopf := http.Header{"X-Recorder-Workspace": {"ws-1"}, "Content-Type": {"application/json"}}
	for i, weg := range wege {
		if !weg.gespeichert {
			continue
		}
		n := betriebWert(modusMock, i+1, "n")
		rumpf := `{"Id":` + n + `,"Name":"` + betriebWert(modusMock, i+1, "g") + `"}`
		if status, body := betriebSende(t, w.h.Adresse(), "POST", "/Products", rumpf, kopf); status != 201 {
			t.Fatalf("mock %s: Anlegen: Status %d, Body %q", weg.name, status, body)
		}
		if !weg.geloescht {
			continue
		}
		if status, body := betriebSende(t, w.h.Adresse(), "DELETE", "/Products("+n+")", "", kopf); status != 204 {
			t.Fatalf("mock %s: Löschen: Status %d, Body %q", weg.name, status, body)
		}
	}
	if err := w.h.Herunterfahren(context.Background()); err != nil {
		t.Fatalf("mock: Herunterfahren nach dem Anlegen: %v", err)
	}
}

// betriebOrteJe zählt je Ort die Wege, deren Muster ihn tragen.
func betriebOrteJe(wege []betriebWeg) map[string]int {
	je := map[string]int{}
	for _, weg := range wege {
		muster := weg.pfad + weg.rumpf
		for ort, plaetze := range map[string][]string{"query": {"{q}"}, "schluessel": {"{k}", "{n}"}, "body": {"{b}", "{m}"}} {
			if slices.ContainsFunc(plaetze, func(p string) bool { return strings.Contains(muster, p) }) {
				je[ort]++
			}
		}
	}
	return je
}

// TestBetriebFehlerwegeOhneWerte fährt in record, replay und mock die Wege der Listen unten mit Fehlerobjekt (und in record die verworfenen
// Bodies) am OData-Listener, je Weg mit Werten als Literal der Query, als Literal des Schlüssels und im Body, soweit das Muster den Ort trägt
// (die Wege über eine Aufzeichnung mit deren Werten), und hält: Status und Code des Wegs, jeder Wert kam beim Dienst an, keiner steht in der
// Senke, kein Verwurf (LH-QA-04.a). Je Modus ist die Zahl der Orte der drei Hauptlisten fest. Gesucht wird der ganze Wert in jeder Groß- und
// Kleinschreibung, kein Teilstück und keine Kodierung. Die Wege sind eine Auswahl der Rückgabestellen der Dienste, nicht alle.
func TestBetriebFehlerwegeOhneWerte(t *testing.T) {
	record, replay, mockWege := betriebRecordWege(), betriebReplayWege("ses-andere"), betriebMockWege()
	for _, f := range []struct {
		modus string
		wege  []betriebWeg
		soll  map[string]int
	}{
		{modusRecord, record, map[string]int{"query": 13, "schluessel": 13, "body": 13}},
		{modusReplay, replay, map[string]int{"query": 13, "schluessel": 13, "body": 13}},
		{modusMock, mockWege, map[string]int{"query": 30, "schluessel": 24, "body": 29}},
	} {
		if got := betriebOrteJe(f.wege); !reflect.DeepEqual(got, f.soll) {
			t.Errorf("%s: %d Wege, Orte %v, erwartet %v", f.modus, len(f.wege), got, f.soll)
		}
	}

	for _, modus := range []string{modusRecord, modusReplay, modusMock} {
		betriebOhneDienst(t, modus)
	}

	// record: drei Services; die Session von andere ist beendet, doppelt hat zwei offene.
	w, ein := betriebProzess(t, modusRecord, mockDaten(t), betriebDienste("/", "/andere", "/doppelt"))
	offen := betriebSessions(t, w, true)
	if _, err := w.h.Sessions().Beenden(context.Background(), offen["andere"]); err != nil {
		t.Fatalf("record: Beenden: %v", err)
	}
	if _, err := w.h.Sessions().Create(context.Background(), "doppelt", ""); err != nil {
		t.Fatalf("record: zweite Session: %v", err)
	}
	betriebWege(t, modusRecord, w, ein, nil, record)

	// replay: die Aufzeichnung trägt den Treffer (Satz 9) und einen unvollständigen Exchange (Satz 10).
	dir := mockDaten(t)
	w, ein = betriebProzess(t, modusRecord, dir, betriebDienste("/", "/andere"))
	sessions := betriebSessions(t, w, true)
	for _, a := range []betriebWeg{
		{name: "aufzeichnen_treffer", methode: "POST", pfad: wegPfad, rumpf: wegRumpf, satz: 9, status: 200},
		{name: "aufzeichnen_ohne_antwort", methode: "POST", pfad: wegPfad, rumpf: wegRumpf, satz: 10, status: 502, code: "ODR-E4001",
			kopf: http.Header{"Betrieb-Upstream": {"nicht_erreichbar"}}},
	} {
		betriebFahre(t, modusReplay, w, ein, nil, a)
	}
	if err := w.h.Herunterfahren(context.Background()); err != nil {
		t.Fatalf("replay: Herunterfahren nach dem Aufzeichnen: %v", err)
	}
	basis := http.Header{"X-Recorder-Workspace": {"ws-1"}, "X-Recorder-Recording-Session": {sessions["svc"]}, "X-Recorder-Replay-Session": {"rep-1"}}
	w, ein = betriebProzess(t, modusReplay, dir, betriebDienste("/", "/andere"))
	if _, err := w.h.Sessions().Create(context.Background(), "svc", "ses-offen"); err != nil {
		t.Fatalf("replay: offene Session: %v", err)
	}
	betriebWege(t, modusReplay, w, ein, basis, betriebReplayWege(sessions["andere"]))
	betriebStrategien(t, dir, basis)

	// mock: die Entities der Wege legt ein eigener Prozess an; zuletzt ein Prozess mit STRICT_COVERAGE.
	dir = mockDaten(t)
	betriebLege(t, dir, mockWege)
	w, ein = betriebProzess(t, modusMock, dir, betriebDienste("/", "/andere"))
	ws := http.Header{"X-Recorder-Workspace": {"ws-1"}}
	betriebWege(t, modusMock, w, ein, ws, mockWege)
	w, ein = betriebProzess(t, modusMock, dir, func(k *bootstrap.Konfiguration) {
		betriebDienste("/", "/andere")(k)
		k.StrictCoverage = true
	})
	betriebWege(t, modusMock, w, ein, ws, []betriebWeg{
		{name: "strikt_ungedeckt", methode: "GET", pfad: "/Orders?x={q}", rumpf: wegRumpf, satz: 98, status: 422, code: "ODR-E6003"},
	})
}

// betriebStrategien fährt über der Aufzeichnung die Wege der Strategien strict (anderer Key, erwarteter Exchange nicht replayfähig) und
// last (nicht replayfähig), je Strategie in einem eigenen Prozess mit eigener Replay-Session.
func betriebStrategien(t *testing.T, dir string, basis http.Header) {
	t.Helper()
	weg := func(name string, satz, status int, code string) betriebWeg {
		return betriebWeg{name: name, methode: "POST", pfad: wegPfad, rumpf: wegRumpf, satz: satz, status: status, code: code}
	}
	for strategie, wege := range map[string][]betriebWeg{
		"strict": {weg("anderer_key", 10, 409, "ODR-E5007"), weg("treffer", 9, 200, ""),
			weg("strict_nicht_replayfaehig", 10, 422, "ODR-E5008")},
		"last": {weg("last_nicht_replayfaehig", 10, 422, "ODR-E5008")},
	} {
		w, ein := betriebProzess(t, modusReplay, dir, func(k *bootstrap.Konfiguration) {
			betriebDienste("/", "/andere")(k)
			k.ReplayStrategy = strategie
		})
		kopf := basis.Clone()
		kopf.Set("X-Recorder-Replay-Session", "rep-"+strategie)
		betriebWege(t, modusReplay, w, ein, kopf, wege)
	}
}
