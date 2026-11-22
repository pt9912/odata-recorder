package bootstrap_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// gwKanarie bildet den Wert des Requests aus Bytewerten (AGENTS.md §3.18); die Form Token trägt die kleine, die Form Methode die große
// Schreibweise.
func gwKanarie(gross bool) string {
	b := []byte{0x63, 0x61, 0x6e, 0x61, 0x72, 0x79}
	if gross {
		return strings.ToUpper(string(b))
	}
	return string(b)
}

// gwServer ist der Adapter `httpodata` über dem echten Adapter der Telemetrie (feste Uhr, Senke ein Puffer) mit dem Dienst tu im Modus;
// er hört auf 127.0.0.1:0 und schließt mit dem Test.
func gwServer(t *testing.T, modus string, tu httpodata.Verarbeiter) (*httpodata.Server, string, *telemetry.Adapter, *syncPuffer) {
	t.Helper()
	senke := &syncPuffer{}
	ad, err := telemetry.New(telemetry.Config{Writer: senke, Clock: supporttest.NewFestUhr(wsT0())})
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: tu, IDs: supporttest.NewFestIDs(), Listener: l, Telemetrie: ad, Modus: modus})
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve() }()
	t.Cleanup(func() { _ = srv.Schliessen(time.Second) })
	return srv, l.Addr().String(), ad, senke
}

func gwClient() *http.Client {
	return &http.Client{Transport: &http.Transport{DisableKeepAlives: true}, Timeout: 10 * time.Second}
}

// gwRuf sendet die Methode an pfad und liefert den Status; ein Abbruch der Verbindung liefert 0.
func gwRuf(t *testing.T, adr, methode, pfad string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), methode, "http://"+adr+pfad, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := gwClient().Do(req)
	if err != nil {
		return 0
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp.StatusCode
}

func gwOk(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
	return inbound.Antwort{Status: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
}

// gwVerwuerfe liefert die Zähler der Verwürfe, die nicht 0 sind.
func gwVerwuerfe(ad *telemetry.Adapter) map[support.Verwurf]uint64 {
	aus := map[support.Verwurf]uint64{}
	for v, n := range ad.Snapshot().Dropped {
		if n != 0 {
			aus[v] = n
		}
	}
	return aus
}

// gwMethoden ist M-Methoden (G5): zehn Methoden mit Form, ein Token aus 16 und eines aus 17 `A` und zwei Verstöße gegen die Form.
func gwMethoden() []string {
	return []string{
		"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "TRACE", "PROPFIND",
		strings.Repeat("A", 16), strings.Repeat("A", 17), "get", "M-SEARCH",
	}
}

// gwSpanZeile liefert die eine Zeile des Gesamtspans in der Senke.
func gwSpanZeile(t *testing.T, senke *syncPuffer) map[string]any {
	t.Helper()
	var gesamt []map[string]any
	for _, z := range zeilen(t, senke.String()) {
		if z["msg"] == "odata.gateway.request" {
			gesamt = append(gesamt, z)
		}
	}
	if len(gesamt) != 1 {
		t.Fatalf("%d Zeilen des Gesamtspans in der Senke, erwartet 1: %s", len(gesamt), senke.String())
	}
	return gesamt[0]
}

// TestGatewayMethodeMitDemEchtenAdapter hält G5 (SPEC-294 Nr. 4 a, d): die Methode steht in der Zeile, wenn sie die Form erfüllt; sonst fehlt
// das Attribut, und `wert` und `pflicht` zählen je 1. 14 Methoden, 3 Verstöße (17 Bytes, `get`, `M-SEARCH`).
func TestGatewayMethodeMitDemEchtenAdapter(t *testing.T) {
	ms := gwMethoden()
	if len(ms) != 14 {
		t.Fatalf("%d Methoden, erwartet 14", len(ms))
	}
	verstoesse := 0
	for _, m := range ms {
		_, adr, ad, senke := gwServer(t, "record", gwOk)
		if s := gwRuf(t, adr, m, "/p"); s != 200 {
			t.Errorf("%q: Status %d, erwartet 200", m, s)
		}
		z := gwSpanZeile(t, senke)
		soll := map[support.Verwurf]uint64{}
		got, da := z["http.request.method"]
		if m == strings.Repeat("A", 17) || m == "get" || m == "M-SEARCH" {
			verstoesse++
			soll = map[support.Verwurf]uint64{support.VerwurfWert: 1, support.VerwurfPflicht: 1}
			if da {
				t.Errorf("%q: die Zeile trägt die Methode %v", m, got)
			}
		} else if got != m {
			t.Errorf("%q: die Zeile trägt die Methode %v", m, got)
		}
		gwGleich(t, m+": Verwürfe", gwVerwuerfe(ad), soll)
	}
	if verstoesse != 3 {
		t.Errorf("%d Verstöße, erwartet 3", verstoesse)
	}
}

func gwGleich(t *testing.T, was string, got, soll any) {
	t.Helper()
	if !reflect.DeepEqual(got, soll) {
		t.Errorf("%s = %v, erwartet %v", was, got, soll)
	}
}

// TestGatewayModusMitDemEchtenAdapter hält G6 (SPEC-294 Nr. 4 b, d): der Modus steht unverändert in der Zeile; ein Wert außerhalb der Form
// fehlt, und `wert` und `pflicht` zählen je 1. 4 Werte und einer außerhalb der Form, 5 Fälle; je Request der Header `X-Recorder-Mode`.
func TestGatewayModusMitDemEchtenAdapter(t *testing.T) {
	werte := []string{"record", "replay", "mock", "hybrid", "Mock"}
	for _, w := range werte {
		_, adr, ad, senke := gwServer(t, w, gwOk)
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://"+adr+"/p", nil)
		req.Header.Set("X-Recorder-Mode", gwKanarie(false))
		resp, err := gwClient().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		z := gwSpanZeile(t, senke)
		got, da := z["recorder.mode"]
		soll := map[support.Verwurf]uint64{}
		if w == "Mock" {
			soll = map[support.Verwurf]uint64{support.VerwurfWert: 1, support.VerwurfPflicht: 1}
			if da {
				t.Errorf("%q: die Zeile trägt den Modus %v", w, got)
			}
		} else if got != w {
			t.Errorf("%q: die Zeile trägt den Modus %v", w, got)
		}
		gwGleich(t, w+": Verwürfe", gwVerwuerfe(ad), soll)
	}
}

// TestGatewayAbbruchMitDemEchtenAdapter hält G8 (SPEC-294 Nr. 3 b, 4 c): ein Kontextfehler ohne Code bricht die Verbindung ab; die Zeile
// trägt keinen Status, `pflicht` zählt 1, `wert` 0. 3 Formen.
func TestGatewayAbbruchMitDemEchtenAdapter(t *testing.T) {
	formen := []error{context.Canceled, context.DeadlineExceeded, errors.Join(context.Canceled, &inbound.Fehler{Code: model.CodeE6008, Status: 404})}
	for _, f := range formen {
		_, adr, ad, senke := gwServer(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) { return inbound.Antwort{}, f })
		if s := gwRuf(t, adr, "GET", "/p"); s != 0 {
			t.Errorf("%v: Status %d, erwartet den Abbruch der Verbindung", f, s)
		}
		z := gwSpanZeile(t, senke)
		if _, da := z["http.response.status_code"]; da {
			t.Errorf("%v: die Zeile trägt einen Status: %v", f, z)
		}
		gwGleich(t, "Verwürfe", gwVerwuerfe(ad), map[support.Verwurf]uint64{support.VerwurfPflicht: 1})
	}
}

// TestGatewayKindHatDenGesamtspanAlsEltern hält G12 (SPEC-294 Nr. 3 d, e): ein Span, den der Dienst mit dem ctx des Aufrufs öffnet, hat den
// Gesamtspan als Eltern; der Gesamtspan hat keine.
func TestGatewayKindHatDenGesamtspanAlsEltern(t *testing.T) {
	var ad *telemetry.Adapter
	tu := func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		_, kind := ad.StartSpan(ctx, support.SpanMockQuery, support.Str("recorder.entity_set", "Products"), support.Str("recorder.coverage_status", "complete"))
		kind.End()
		return gwOk(ctx, a)
	}
	_, adr, adapter, _ := gwServer(t, "mock", tu)
	ad = adapter
	if s := gwRuf(t, adr, "GET", "/p"); s != 200 {
		t.Fatalf("Status %d", s)
	}
	spans := ad.Snapshot().Spans
	if len(spans) != 2 || spans[0].Name != support.SpanMockQuery || spans[1].Name != support.SpanGatewayRequest {
		t.Fatalf("Spans %+v; erwartet das Kind vor dem Gesamtspan", spans)
	}
	if spans[0].ParentID == "" || spans[0].ParentID != spans[1].ID || spans[1].ParentID != "" {
		t.Errorf("ParentID des Kindes %q, ID des Gesamtspans %q, ParentID des Gesamtspans %q", spans[0].ParentID, spans[1].ID, spans[1].ParentID)
	}
}

// gwStelle ist eine der 8 Stellen von SPEC-294 Nr. 8: setze legt den Wert k an der Stelle in den Request, ursache an den Fehler des Dienstes.
type gwStelle struct {
	name    string
	setze   func(r *http.Request, k string)
	ursache bool
}

func gwStellen() []gwStelle {
	return []gwStelle{
		{"Pfad", func(r *http.Request, k string) { r.URL.Path = "/" + k }, false},
		{"Name der Query", func(r *http.Request, k string) { r.URL.RawQuery = k + "=1" }, false},
		{"Wert der Query", func(r *http.Request, k string) { r.URL.RawQuery = "x=" + k }, false},
		{"Name eines Headers", func(r *http.Request, k string) { r.Header.Set("X-"+k, "1") }, false},
		{"Wert eines Headers", func(r *http.Request, k string) { r.Header.Set("X-Wert", k) }, false},
		{"Body", func(r *http.Request, k string) {
			r.Body, r.ContentLength = io.NopCloser(strings.NewReader(k)), int64(len(k))
		}, false},
		{"Host", func(r *http.Request, k string) { r.Host = k }, false},
		{"Text der Ursache", func(*http.Request, string) {}, true},
	}
}

// TestGatewayKeinRequestWertInDerSenke hält G13 (SPEC-294 Nr. 8): ein Wert an einer der 8 Stellen des Requests erreicht in keiner der 3 Arten
// der Antwort (Antwort des Dienstes, Fehlerobjekt, 503) die Bytes der Senke des echten Adapters, und kein Attribut wird verworfen (ein Wert,
// der die Form verfehlt, würde zählen); der Wert steht in zwei Formen (Token `canary`, Methode `CANARY`).
func TestGatewayKeinRequestWertInDerSenke(t *testing.T) {
	stellen, arten := gwStellen(), []string{"Antwort", "Fehlerobjekt", "503"}
	if len(stellen) != 8 || len(arten) != 3 {
		t.Fatalf("%d Stellen und %d Arten, erwartet 8 und 3", len(stellen), len(arten))
	}
	n := 0
	for _, s := range stellen {
		for _, art := range arten {
			n++
			for _, gross := range []bool{false, true} {
				k := gwKanarie(gross)
				tu := gwOk
				if art == "Fehlerobjekt" {
					ursache := errors.New("roh")
					if s.ursache {
						ursache = errors.New(k)
					}
					tu = func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
						return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE6008, Status: 404, Ursache: ursache}
					}
				}
				srv, adr, ad, senke := gwServer(t, "record", tu)
				c := &http.Client{Timeout: 10 * time.Second}
				if art == "503" {
					warm, _ := http.NewRequestWithContext(context.Background(), "GET", "http://"+adr+"/warm", nil)
					resp, err := c.Do(warm)
					if err != nil {
						t.Fatal(err)
					}
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					srv.Anhalten()
				}
				req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://"+adr+"/p", nil)
				s.setze(req, k)
				resp, err := c.Do(req)
				if err != nil {
					t.Fatalf("%s %s: %v", s.name, art, err)
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if strings.Contains(strings.ToLower(senke.String()), gwKanarie(false)) {
					t.Errorf("%s %s %v: die Senke trägt den Wert des Requests: %s", s.name, art, gross, senke.String())
				}
				if v := gwVerwuerfe(ad); len(v) != 0 {
					t.Errorf("%s %s %v: Verwürfe %v", s.name, art, gross, v)
				}
				soll := map[string]int{"Antwort": 200, "Fehlerobjekt": 404, "503": 503}[art]
				if resp.StatusCode != soll {
					t.Errorf("%s %s: Status %d, erwartet %d", s.name, art, resp.StatusCode, soll)
				}
			}
		}
	}
	if n != 24 {
		t.Errorf("%d Fälle, erwartet 24", n)
	}
}

// TestGatewayKeinFehlertextInDerSenke hält G13 (SPEC-294 Nr. 8) für die zweite Form des Fehlers der Stelle der Ursache: ein Fehler ohne Code,
// dessen Meldung der Wert ist, erreicht die Bytes der Senke nicht, und kein Attribut wird verworfen; beide Formen des Werts.
func TestGatewayKeinFehlertextInDerSenke(t *testing.T) {
	for _, gross := range []bool{false, true} {
		k := gwKanarie(gross)
		_, adr, ad, senke := gwServer(t, "record", func(context.Context, inbound.Anfrage) (inbound.Antwort, error) {
			return inbound.Antwort{}, errors.New(k)
		})
		if s := gwRuf(t, adr, "GET", "/p"); s != 500 {
			t.Fatalf("Status %d, erwartet 500", s)
		}
		if strings.Contains(strings.ToLower(senke.String()), gwKanarie(false)) || len(gwVerwuerfe(ad)) != 0 {
			t.Errorf("%v: Senke %q, Verwürfe %v; erwartet keinen Text des Fehlers und keinen Verwurf", gross, senke.String(), gwVerwuerfe(ad))
		}
	}
}

// gwZeile ist die Zeile des Gesamtspans als Literal aus SPEC-231 Nr. 6: Felder in der Folge timestamp, level, msg, ereignis, span_id, duration_ms
// und die Attribute in der Folge der Tabelle von §5 (Methode, Status, Modus).
func gwZeile(spanID, methode string, status int, modus string) string {
	return `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"INFO","msg":"odata.gateway.request","ereignis":"span","span_id":"` + spanID +
		`","duration_ms":0.000,"http.request.method":"` + methode + `","http.response.status_code":` + strconv.Itoa(status) + `,"recorder.mode":"` + modus + `"}` + "\n"
}

// gwAnfrage ist eine Anfrage an den Prozess: Methode, Pfad und ob der Header des Workspace fehlt.
type gwAnfrage struct {
	pfad      string
	workspace bool
	status    int
}

// TestWurzelGesamtspanJedeAnfrage hält G17 (SPEC-139 Nr. 13 e und f, SPEC-294 Nr. 2): durch den Composition Root schreibt in jedem der 4 Modi
// sowohl eine Anfrage mit Antwort des Dienstes als auch eine abgewiesene die Zeile des Gesamtspans in die Senke des Prozesses: Literal der
// Zeile mit `recorder.mode` gleich dem Modus und dem Status, `Dropped` überall 0, und der Server erhält den einen Adapter des Prozesses.
func TestWurzelGesamtspanJedeAnfrage(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[]}`))
	}))
	defer up.Close()
	faelle := map[string][2]gwAnfrage{
		modusMock:   {{"/Products", true, 200}, {"/Products", false, 400}},
		modusHybrid: {{"/Products", true, 200}, {"/Products", false, 400}},
		modusRecord: {{"/svc/Products", true, 200}, {"/zzz", true, 404}},
		modusReplay: {{"/svc/Products", false, 400}, {"/zzz", true, 404}},
	}
	n := 0
	for modus, anfragen := range faelle {
		for i, a := range anfragen {
			n++
			dir := t.TempDir()
			if modus == modusMock || modus == modusHybrid {
				dir = mockDaten(t)
			}
			w := baueWurzelMit(t, modus, nil, dir, func(k *bootstrap.Konfiguration) {
				if modus == modusRecord || modus == modusReplay {
					k.Dienste[0].Root = "/svc"
				}
				if modus == modusRecord {
					k.Dienste[0].UpstreamURL = up.URL
				}
			})
			if err := w.h.Hochfahren(context.Background()); err != nil {
				t.Fatalf("%s: Hochfahren: %v", modus, err)
			}
			req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://"+w.h.Adresse()+a.pfad, nil)
			if a.workspace {
				req.Header.Set("X-Recorder-Workspace", "ws-1")
			}
			resp, err := gwClient().Do(req)
			if err != nil {
				t.Fatalf("%s %d: %v", modus, i, err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != a.status {
				t.Errorf("%s %s: Status %d, erwartet %d", modus, a.pfad, resp.StatusCode, a.status)
			}
			z := zeilen(t, w.senke.String())
			if len(z) == 0 || z[len(z)-1]["msg"] != "odata.gateway.request" {
				t.Fatalf("%s %s: Senke %q, erwartet die Zeile des Gesamtspans zuletzt", modus, a.pfad, w.senke.String())
			}
			soll := gwZeile("spn-0000000000000001", "GET", a.status, modus)
			if letzte := w.senke.String()[strings.LastIndex(strings.TrimSuffix(w.senke.String(), "\n"), "\n")+1:]; letzte != soll {
				t.Errorf("%s %s: Zeile\n%s\nerwartet\n%s", modus, a.pfad, letzte, soll)
			}
			keinVerwurf(t, modus, w.adapter[0])
			if soll := map[string]int{modusRecord: 1, modusReplay: 2, modusMock: 1, modusHybrid: 2}[modus]; len(w.adapter) != soll {
				t.Errorf("%s: %d Adapter der Telemetrie, erwartet %d (der Server erhält den des Prozesses, keinen eigenen)", modus, len(w.adapter), soll)
			}
		}
	}
	if n != 8 {
		t.Errorf("%d Fälle, erwartet 8", n)
	}
}

// TestWurzelMockSpanLiegtImGesamtspan hält G17 (SPEC-294 Nr. 3 e): in den Modi mock und hybrid steht die Zeile `odata.mock.query` vor der
// Zeile des Gesamtspans, und ihr `parent_id` ist dessen `span_id`.
func TestWurzelMockSpanLiegtImGesamtspan(t *testing.T) {
	for _, modus := range []string{modusMock, modusHybrid} {
		w := baueWurzel(t, modus, nil)
		if err := w.h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: Hochfahren: %v", modus, err)
		}
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://"+w.h.Adresse()+"/Products", nil)
		req.Header.Set("X-Recorder-Workspace", "ws-1")
		resp, err := gwClient().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		z := zeilen(t, w.senke.String())
		if len(z) != 2 || z[0]["msg"] != "odata.mock.query" || z[1]["msg"] != "odata.gateway.request" || z[0]["parent_id"] != z[1]["span_id"] || z[1]["span_id"] == nil {
			t.Errorf("%s: Zeilen %v; erwartet die Zeile des Mock-Spans mit parent_id gleich span_id der folgenden Zeile des Gesamtspans", modus, z)
		}
	}
}

// TestHochfahrenUebergibtAdapterUndModus hält G17 (SPEC-139 Nr. 13 e): Hochfahren übergibt dem Server den Adapter des Prozesses (den ersten,
// der entsteht) und den Modus der Konfiguration, in jedem der 4 Modi.
func TestHochfahrenUebergibtAdapterUndModus(t *testing.T) {
	n := 0
	for _, modus := range []string{modusRecord, modusReplay, modusMock, modusHybrid} {
		n++
		hk := neuHaken()
		b := hk.bauteile()
		var adapter []*telemetry.Adapter
		neu := b.NeueTelemetrie
		b.NeueTelemetrie = func(k telemetry.Config) (*telemetry.Adapter, error) {
			a, err := neu(k)
			adapter = append(adapter, a)
			return a, err
		}
		h, err := bootstrap.BuildMit(context.Background(), wurzelKonfig(modus, t.TempDir()), nil, b)
		if err != nil {
			t.Fatalf("%s: Build: %v", modus, err)
		}
		if err := h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: Hochfahren: %v", modus, err)
		}
		_ = h.Herunterfahren(context.Background())
		hk.mu.Lock()
		k := hk.server
		hk.mu.Unlock()
		if k.Modus != modus || len(adapter) == 0 || k.Telemetrie != support.Telemetry(adapter[0]) {
			t.Errorf("%s: Modus %q, Telemetrie %v; erwartet den Modus und den ersten Adapter %v", modus, k.Modus, k.Telemetrie, adapter)
		}
	}
	if n != 4 {
		t.Errorf("%d Modi, erwartet 4", n)
	}
}
