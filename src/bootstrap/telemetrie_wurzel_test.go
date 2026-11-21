package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// telemetrieFuerTest ist die Telemetrie der Tests, die den Mock ohne den Adapter des Prozesses bauen.
func telemetrieFuerTest() support.Telemetry { return supporttest.NewFakeTelemetry() }

// kanarieWurzel ist ein wohlgeformter Wert (Buchstaben und Ziffern): ihn lässt die Form des Adapters durch, so dass nur der Dienst
// ihn fernhält (SPEC-232 Nr. 2, *Ungeprüft*).
const kanarieWurzel = "KanarieQ7f"

// mockDaten legt in einem frischen DATA_DIR an, was der Mock liest: den Dienst svc (Root /) mit den Metadaten, die Baseline bsl-1 mit der
// Coverage und den Workspace ws-1 (SPEC-049 Nr. 2); die Datei ist danach geschlossen.
func mockDaten(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, verzeichnisB), 0o700); err != nil {
		t.Fatal(err)
	}
	u := abOeffne(t, filepath.Join(dir, dateiDB), filepath.Join(dir, verzeichnisB))
	u.ausgangsdaten()
	abMust(t, func() error { _, err := u.ws.Create(context.Background(), "bsl-1", "ws-1"); return err }())
	u.schliesse()
	return dir
}

// wurzelKonfig ist die Konfiguration des Dienstes svc (Root /) im Modus über dem DATA_DIR.
func wurzelKonfig(modus, dir string) bootstrap.Konfiguration {
	d := bootstrap.Dienst{ID: "svc", Root: "/"}
	switch modus {
	case modusMock:
		k := basisKonfig(dir, modus, d)
		k.PublicBaseURL = "http://recorder.test"
		return k
	case modusHybrid:
		return hybridKonfig(dir, policy.Werte{}, d)
	}
	d.UpstreamURL = "http://127.0.0.1:9/svc"
	return basisKonfig(dir, modus, d)
}

// wurzel ist ein gebauter Prozess über den echten Adaptern: die Adapter der Telemetrie in der Folge ihrer Erzeugung (der erste ist der
// des Prozesses, mit fester Uhr), der Mock und die Senke des Prozesses.
type wurzel struct {
	h       *bootstrap.Handle
	mock    inbound.MockUseCase
	adapter []*telemetry.Adapter
	senke   *syncPuffer
}

// bereinigungFake ist die Wartung des Tests: CleanupBlobs liefert das eingestellte Ergebnis und zählt.
type bereinigungFake struct {
	inbound.MaintenanceUseCase
	bericht inbound.Loeschbericht
	err     error
	mu      sync.Mutex
	n       int
}

func (f *bereinigungFake) CleanupBlobs(context.Context) (inbound.Loeschbericht, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	return f.bericht, f.err
}

func (f *bereinigungFake) aufrufe() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.n
}

// baueWurzel baut den Prozess im Modus über frischen Daten; wartung ersetzt die Wartung, wenn nicht nil.
func baueWurzel(t *testing.T, modus string, wartung inbound.MaintenanceUseCase) *wurzel {
	t.Helper()
	w := &wurzel{senke: &syncPuffer{}}
	b := bootstrap.StandardBauteile()
	tel, neuerMock := b.NeueTelemetrie, b.NeuerMock
	b.NeueTelemetrie = func(k telemetry.Config) (*telemetry.Adapter, error) {
		k.Clock = supporttest.NewFestUhr(wsT0())
		a, err := tel(k)
		w.adapter = append(w.adapter, a)
		return a, err
	}
	b.NeuerMock = func(k mock.Konfig) (inbound.MockUseCase, error) {
		m, err := neuerMock(k)
		w.mock = m
		return m, err
	}
	if wartung != nil {
		b.NeueWartung = func(maintenance.Konfig) (inbound.MaintenanceUseCase, error) { return wartung, nil }
	}
	h, err := bootstrap.BuildMitSenke(context.Background(), wurzelKonfig(modus, mockDaten(t)), nil, w.senke, b)
	if err != nil {
		t.Fatalf("Build im Modus %s: %v", modus, err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
	w.h = h
	return w
}

// queryAnfrage ist GET auf pfad an den Dienst svc im Workspace ws-1.
func queryAnfrage(pfad, rohQuery string) inbound.Anfrage {
	return inbound.Anfrage{Methode: "GET", Pfad: pfad, RawQuery: rohQuery, Header: map[string][]string{"X-Recorder-Workspace": {"ws-1"}}, Laenge: -1}
}

// frage ruft den Mock und liest den Body der Antwort.
func (w *wurzel) frage(ctx context.Context, t *testing.T, a inbound.Anfrage) error {
	t.Helper()
	antwort, err := w.mock.Mock(ctx, a)
	if err == nil {
		_, _ = io.Copy(io.Discard, antwort.Body)
		_ = antwort.Body.Close()
	}
	return err
}

// zeilen zerlegt die Senke in ihre JSON-Zeilen.
func zeilen(t *testing.T, s string) []map[string]any {
	t.Helper()
	var r []map[string]any
	for _, z := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		if z == "" {
			continue
		}
		m := map[string]any{}
		if err := json.Unmarshal([]byte(z), &m); err != nil {
			t.Fatalf("Zeile %q: %v", z, err)
		}
		r = append(r, m)
	}
	return r
}

// keinVerwurf prüft, dass der Adapter keinen Verwurf gezählt hat (Q15, Q17).
func keinVerwurf(t *testing.T, kontext string, a *telemetry.Adapter) {
	t.Helper()
	st := a.Snapshot()
	for _, v := range support.Verwuerfe() {
		if st.Dropped[v] != 0 {
			t.Errorf("%s: Verwurf %s %d-mal", kontext, v, st.Dropped[v])
		}
	}
}

// TestWurzelSpanZeile hält Q15 (SPEC-139 Nr. 13 a, d): in den Modi mock und hybrid erhält der Mock den Adapter des Prozesses; eine Anfrage
// schreibt die Zeile mit `ereignis` `span` und beiden Attributen in die Senke, Felder und Reihenfolge nach SPEC-231 Nr. 6 (Literal), und
// kein Attribut wird verworfen. Im Modus hybrid besteht daneben der Adapter des Replay-Dienstes; die Zeile steht nur in der Senke des Prozesses.
func TestWurzelSpanZeile(t *testing.T) {
	soll := `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"INFO","msg":"odata.mock.query","ereignis":"span",` +
		`"span_id":"spn-0000000000000001","duration_ms":0.000,"recorder.entity_set":"Products","recorder.coverage_status":"complete"}` + "\n"
	for modus, adapter := range map[string]int{modusMock: 1, modusHybrid: 2} {
		w := baueWurzel(t, modus, nil)
		if err := w.frage(context.Background(), t, queryAnfrage("/Products", "")); err != nil {
			t.Fatalf("%s: %v", modus, err)
		}
		if got := w.senke.String(); got != soll {
			t.Errorf("%s: Senke\n%s\nerwartet\n%s", modus, got, soll)
		}
		if len(w.adapter) != adapter {
			t.Errorf("%s: %d Adapter der Telemetrie, erwartet %d", modus, len(w.adapter), adapter)
		}
		for i, a := range w.adapter {
			keinVerwurf(t, modus, a)
			if n := len(a.Snapshot().Spans); n != 1-min(i, 1) {
				t.Errorf("%s: Adapter %d trägt %d Spans, erwartet %d (allein der des Prozesses)", modus, i, n, 1-min(i, 1))
			}
		}
	}
}

// TestBuildTelemetrieFehler hält Q15 (Randform): scheitert der Bau des Adapters des Prozesses, scheitert Build mit diesem Fehler und lässt die
// Ablage geschlossen (SPEC-139 Nr. 2), in jedem Modus.
func TestBuildTelemetrieFehler(t *testing.T) {
	for _, modus := range []string{modusRecord, modusReplay, modusMock, modusHybrid} {
		h := neuHaken()
		b := h.bauteile()
		b.NeueTelemetrie = func(telemetry.Config) (*telemetry.Adapter, error) { return nil, errPhase }
		_, err := bootstrap.BuildMit(context.Background(), wurzelKonfig(modus, t.TempDir()), nil, b)
		if !errors.Is(err, errPhase) || h.p.zaehle(pAblageZu) != 1 {
			t.Errorf("%s: Fehler %v, Ablage %d-mal geschlossen; erwartet den Fehler der Telemetrie und 1", modus, err, h.p.zaehle(pAblageZu))
		}
	}
}

// TestWurzelSpanEltern hält Q8 (SPEC-290 Nr. 3 e): der Span der Query ist Kind des Gesamtspans, wenn das ctx ihn trägt, sonst ohne
// Eltern; ParentID im Stand ist die ID des Elternspans, und die Zeile trägt `parent_id` gleich dessen `span_id`.
func TestWurzelSpanEltern(t *testing.T) {
	w := baueWurzel(t, modusMock, nil)
	ad := w.adapter[0]
	ctx, eltern := ad.StartSpan(context.Background(), support.SpanGatewayRequest,
		support.Str("http.request.method", "GET"), support.Int("http.response.status_code", 200), support.Str("recorder.mode", "mock"))
	if err := w.frage(ctx, t, queryAnfrage("/Products", "")); err != nil {
		t.Fatal(err)
	}
	eltern.End()
	if err := w.frage(context.Background(), t, queryAnfrage("/Products", "")); err != nil {
		t.Fatal(err)
	}
	var kinder, gesamt []support.SpanRecord
	for _, r := range ad.Snapshot().Spans {
		if r.Name == support.SpanMockQuery {
			kinder = append(kinder, r)
		} else {
			gesamt = append(gesamt, r)
		}
	}
	if len(kinder) != 2 || len(gesamt) != 1 || kinder[0].ParentID == "" || kinder[0].ParentID != gesamt[0].ID || kinder[1].ParentID != "" {
		t.Fatalf("Spans %+v %+v; erwartet ein Kind mit ParentID des Gesamtspans und eines ohne", kinder, gesamt)
	}
	z := zeilen(t, w.senke.String())
	if len(z) != 3 || z[0]["parent_id"] != z[1]["span_id"] || z[0]["msg"] != "odata.mock.query" || z[2]["msg"] != "odata.mock.query" {
		t.Fatalf("Zeilen %v", z)
	}
	if _, da := z[2]["parent_id"]; da {
		t.Errorf("die Zeile ohne Eltern trägt parent_id: %v", z[2])
	}
}

// csdlMitSet ist ein CSDL mit dem einen Set name des Typs Product (Schlüssel Id).
func csdlMitSet(name string) string {
	return `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"><edmx:DataServices>` +
		`<Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm"><EntityType Name="Product"><Key><PropertyRef Name="Id"/></Key>` +
		`<Property Name="Id" Type="Edm.Int32" Nullable="false"/></EntityType>` +
		`<EntityContainer Name="C"><EntitySet Name="` + name + `" EntityType="N.Product"/></EntityContainer></Schema></edmx:DataServices></edmx:Edmx>`
}

// TestSpanSetAusserhalbDerForm hält Q9 (SPEC-290 Nr. 4 a, e): trägt das Modell ein Set, dessen Name die Form des Adapters verletzt (ein Byte
// ab 0x80), bleibt der Platzhalter `_.` stehen, der Adapter zählt `wert` einmal und nie `pflicht`; die Antwort bleibt unberührt. Die Bytes
// des Namens entstehen aus Bytewerten (AGENTS.md §3.18).
func TestSpanSetAusserhalbDerForm(t *testing.T) {
	name := "Pr" + string([]byte{0xc3, 0xbc}) + "fer"
	senke := &syncPuffer{}
	ad, err := telemetry.New(telemetry.Config{Writer: senke, Clock: supporttest.NewFestUhr(wsT0())})
	if err != nil {
		t.Fatal(err)
	}
	m := mlDienstMit(t, csdlMitSet(name), model.CoverageComplete, ad)
	status, _, err := mlRuf(t, m, "/svc/Pr%C3%BCfer", "", nil)
	if err != nil || status != 200 {
		t.Fatalf("Antwort %d, %v", status, err)
	}
	st := ad.Snapshot()
	if len(st.Spans) != 1 || st.Spans[0].Attrs[0].Key != "recorder.entity_set" || st.Spans[0].Attrs[0].Text != "_." {
		t.Fatalf("Spans %+v; erwartet einen mit recorder.entity_set gleich _.", st.Spans)
	}
	if st.Dropped[support.VerwurfWert] != 1 || st.Dropped[support.VerwurfPflicht] != 0 || strings.Contains(senke.String(), name) {
		t.Errorf("Verwürfe wert %d pflicht %d, Zeile %q", st.Dropped[support.VerwurfWert], st.Dropped[support.VerwurfPflicht], senke.String())
	}
}

// TestKeinRequestWertInDerSenke hält Q14 (SPEC-290 Nr. 8): ein wohlgeformter Wert an jeder der 8 Stellen des Requests erreicht die Bytes der
// Senke des echten Adapters nicht; das Set mit dem Namen des Requests ist unbekannt, und der Header des Workspace nennt einen, den es nicht gibt.
func TestKeinRequestWertInDerSenke(t *testing.T) {
	k := kanarieWurzel
	faelle := []struct {
		name, pfad, query string
		header            map[string][]string
	}{
		{"Set", "/svc/" + k + "Set", "", nil},
		{"Schlüssel", "/svc/Products('" + k + "')", "", nil},
		{"Literal in $filter", "/svc/Products", "$filter=Name%20eq%20'" + k + "'", nil},
		{"Name in $filter", "/svc/Products", "$filter=" + k + "%20eq%201", nil},
		{"Name in $orderby", "/svc/Products", "$orderby=" + k, nil},
		{"Name in $select", "/svc/Products", "$select=" + k, nil},
		{"Header X-Recorder-Workspace", "/svc/Products", "", map[string][]string{"X-Recorder-Workspace": {k}}},
		{"Accept", "/svc/Products", "", map[string][]string{"Accept": {k}}},
	}
	if len(faelle) != 8 {
		t.Fatalf("%d Stellen, erwartet 8", len(faelle))
	}
	for _, f := range faelle {
		senke := &syncPuffer{}
		ad, err := telemetry.New(telemetry.Config{Writer: senke, Clock: supporttest.NewFestUhr(wsT0())})
		if err != nil {
			t.Fatal(err)
		}
		m := mlDienstMit(t, msProductsXML(), model.CoverageComplete, ad)
		_, _, _ = mlRuf(t, m, f.pfad, f.query, f.header)
		if strings.Contains(senke.String(), k) {
			t.Errorf("%s: die Senke trägt den Wert des Requests: %s", f.name, senke.String())
		}
		if f.name != "Header X-Recorder-Workspace" && len(ad.Snapshot().Spans) != 1 {
			t.Errorf("%s: %d Spans, erwartet 1 (der Test erreicht den Span)", f.name, len(ad.Snapshot().Spans))
		}
	}
}

// holeGET sendet GET an adr und liefert den Status.
func holeGET(t *testing.T, adr, pfad string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+adr+pfad, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Recorder-Workspace", "ws-1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp.StatusCode
}

// TestStartUebergibtStderrAlsSenke hält Q16 (SPEC-139 Nr. 13): Start übergibt seinen Fehlerstrom als Senke des Adapters des Prozesses;
// nach einer Anfrage im Modus mock steht die Zeile des Spans im Puffer, als JSON auf einer Zeile mit beiden Attributen. `start` erhält
// keinen Strom für stdout und schreibt nie dorthin.
func TestStartUebergibtStderrAlsSenke(t *testing.T) {
	dir := mockDaten(t)
	env := []string{"MODE=mock", "LISTEN=127.0.0.1:8080", "DATA_DIR=" + dir, "SERVICE_ID=svc", "SHUTDOWN_TIMEOUT=2s"}
	b := bootstrap.StandardBauteile()
	echt := b.Listen
	adresse := make(chan string, 1)
	b.Listen = func(ctx context.Context, network, _ string) (net.Listener, error) {
		l, err := echt(ctx, network, "127.0.0.1:0")
		if err == nil {
			adresse <- l.Addr().String()
		}
		return l, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stderr := &syncPuffer{}
	code := make(chan int, 1)
	go func() { code <- bootstrap.StartMit(ctx, env, stderr, b) }()
	select {
	case adr := <-adresse:
		if status := holeGET(t, adr, "/Products"); status != 200 {
			t.Errorf("GET /Products: %d", status)
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("kein Listener; stderr:\n%s", stderr.String())
	}
	cancel()
	select {
	case c := <-code:
		if c != 0 {
			t.Errorf("Rückgabe %d, stderr:\n%s", c, stderr.String())
		}
	case <-time.After(20 * time.Second):
		t.Fatal("start kehrt nicht zurück")
	}
	var spanZeilen []map[string]any
	for _, z := range strings.Split(stderr.String(), "\n") {
		if strings.HasPrefix(z, "{") {
			spanZeilen = append(spanZeilen, zeilen(t, z)...)
		}
	}
	if len(spanZeilen) != 1 || spanZeilen[0]["ereignis"] != "span" || spanZeilen[0]["msg"] != "odata.mock.query" ||
		spanZeilen[0]["recorder.entity_set"] != "Products" || spanZeilen[0]["recorder.coverage_status"] != "complete" {
		t.Errorf("Zeilen des Spans in stderr: %v\n%s", spanZeilen, stderr.String())
	}
}

// TestBuildOhneStartSchreibtNichts hält Q16: Build ohne Start übergibt io.Discard; eine Anfrage am Listener des Prozesses schreibt nach
// os.Stderr und os.Stdout kein Byte (beide sind für die Dauer des Tests auf Rohre umgelenkt).
func TestBuildOhneStartSchreibtNichts(t *testing.T) {
	dir := mockDaten(t)
	altErr, altOut := os.Stderr, os.Stdout
	re, we, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	ro, wo, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr, os.Stdout = we, wo
	gelesen := make(chan [2][]byte, 1)
	go func() {
		e, _ := io.ReadAll(re)
		o, _ := io.ReadAll(ro)
		gelesen <- [2][]byte{e, o}
	}()
	k := wurzelKonfig(modusMock, dir)
	h, err := bootstrap.Build(context.Background(), k, nil)
	if err == nil {
		err = h.Hochfahren(context.Background())
	}
	status := 0
	if err == nil {
		status = holeGET(t, h.Adresse(), "/Products")
		err = h.Herunterfahren(context.Background())
	}
	os.Stderr, os.Stdout = altErr, altOut
	_ = we.Close()
	_ = wo.Close()
	strom := <-gelesen
	if err != nil || status != 200 {
		t.Fatalf("Build, Hochfahren, GET, Herunterfahren: %v, Status %d", err, status)
	}
	if len(strom[0]) != 0 || len(strom[1]) != 0 {
		t.Errorf("Build ohne Start schrieb %d Bytes nach stderr und %d nach stdout: %q %q", len(strom[0]), len(strom[1]), strom[0], strom[1])
	}
}

// eintragsZeile ist die Zeile des Eintrags `ereignis` des Dienstes workspace als Literal aus SPEC-129 Nr. 13 und SPEC-231 Nr. 1; die Texte
// tragen ö und ü, die aus Bytewerten entstehen (AGENTS.md §3.18).
func eintragsZeile(ereignis, code string) string {
	oe, ue := string([]byte{0xc3, 0xb6}), string([]byte{0xc3, 0xbc})
	kopf := `{"timestamp":"2026-10-20T10:00:00.000000000Z",`
	if ereignis == "bereinigung_fehlgeschlagen" {
		return kopf + `"level":"ERROR","msg":"workspace: Bereinigung nach dem L` + oe + `schen fehlgeschlagen","workspace_id":"ws-1",` +
			`"error_code":"` + code + `","ereignis":"bereinigung_fehlgeschlagen"}` + "\n"
	}
	return kopf + `"level":"INFO","msg":"workspace: Bereinigung nach dem L` + oe + `schen zur` + ue + `ckgestellt","workspace_id":"ws-1",` +
		`"ereignis":"bereinigung_zurueckgestellt"}` + "\n"
}

// TestWurzelEintraegeWorkspace hält Q17 (SPEC-231 Nr. 4 und 7, SPEC-139 Nr. 13 b): die Einträge des Dienstes workspace laufen in jedem Modus
// über den Handler des Adapters des Prozesses: Zeile mit Text, Stufe und Feldern der Stelle (Literal), kein Verwurf; ein Kontextfehler der
// Bereinigung schreibt nichts. Modus 4 × Fall 3 = 12, dazu je Modus der Kontextfehler.
func TestWurzelEintraegeWorkspace(t *testing.T) {
	faelle := []struct {
		name    string
		bericht inbound.Loeschbericht
		err     error
		soll    string
	}{
		{"ODR-E3001", inbound.Loeschbericht{}, &inbound.Fehler{Code: model.CodeE3001, Status: 500}, eintragsZeile("bereinigung_fehlgeschlagen", "ODR-E3001")},
		{"ODR-E9001", inbound.Loeschbericht{}, errors.New("roh"), eintragsZeile("bereinigung_fehlgeschlagen", "ODR-E9001")},
		{"zurückgestellt", inbound.Loeschbericht{Zurueckgestellt: true}, nil, eintragsZeile("bereinigung_zurueckgestellt", "")},
		{"Kontextfehler", inbound.Loeschbericht{}, context.Canceled, ""},
	}
	n := 0
	for _, modus := range []string{modusRecord, modusReplay, modusMock, modusHybrid} {
		for _, f := range faelle {
			n++
			fake := &bereinigungFake{bericht: f.bericht, err: f.err}
			w := baueWurzel(t, modus, fake)
			if _, err := w.h.Workspaces().Delete(context.Background(), "ws-1"); err != nil {
				t.Fatalf("%s %s: Delete %v", modus, f.name, err)
			}
			if got := w.senke.String(); got != f.soll || fake.aufrufe() != 1 {
				t.Errorf("%s %s: Senke %q, erwartet %q; CleanupBlobs %d-mal", modus, f.name, got, f.soll, fake.aufrufe())
			}
			keinVerwurf(t, modus+" "+f.name, w.adapter[0])
		}
	}
	if n != 16 {
		t.Errorf("%d Läufe, erwartet 16 (12 mit Zeile und 4 des Kontextfehlers)", n)
	}
}

// TestWurzelEintragOhneAbgewiesenenWert hält Q18 (SPEC-231 Nr. 4): ein abgewiesenes Delete schreibt nichts und ruft CleanupBlobs nicht;
// der Eintrag eines gelungenen Delete trägt die Kennung allein in `workspace_id`. Kennung außerhalb der Form: 70 Bytes, `/`, Leerzeichen,
// ein Byte ab 0x80 (aus Bytewerten).
func TestWurzelEintragOhneAbgewiesenenWert(t *testing.T) {
	ids := []string{strings.Repeat("x", 70), kanarieWurzel + "/x", kanarieWurzel + " x", kanarieWurzel + string([]byte{0xc3, 0xbc})}
	for _, modus := range []string{modusMock, modusRecord} {
		fake := &bereinigungFake{bericht: inbound.Loeschbericht{Zurueckgestellt: true}}
		w := baueWurzel(t, modus, fake)
		for _, id := range ids {
			_, err := w.h.Workspaces().Delete(context.Background(), id)
			var f *inbound.Fehler
			if !errors.As(err, &f) || f.Code != model.CodeE7001 {
				t.Errorf("%s: Delete(%q): %v, erwartet ODR-E7001", modus, id, err)
			}
		}
		if w.senke.String() != "" || fake.aufrufe() != 0 {
			t.Errorf("%s: abgewiesene Aufrufe schrieben %q und riefen CleanupBlobs %d-mal", modus, w.senke.String(), fake.aufrufe())
		}
		if _, err := w.h.Workspaces().Delete(context.Background(), "ws-1"); err != nil {
			t.Fatalf("%s: Delete(ws-1): %v", modus, err)
		}
		z := zeilen(t, w.senke.String())
		if len(z) != 1 || z[0]["workspace_id"] != "ws-1" {
			t.Fatalf("%s: Zeilen %v", modus, z)
		}
		for key, v := range z[0] {
			if s, _ := v.(string); key != "workspace_id" && strings.Contains(s, "ws-1") {
				t.Errorf("%s: Feld %s trägt die Kennung: %q", modus, key, s)
			}
		}
	}
}

// ohneSpanPakete sind die Pakete der Dienste, die keinen Span tragen (SPEC-290 Nr. 2).
func ohneSpanPakete() []string {
	return []string{"../hexagon/services/workspace", "../hexagon/services/snapshot"}
}

// TestWorkspaceUndSnapshotOhneSpan hält Q19 (SPEC-290 Nr. 2): die Dienste workspace und snapshot rufen StartSpan und CountError nie und
// führen kein Feld vom Typ support.Telemetry (der Quelltext beider Pakete ohne _test.go); durch den Root lassen die acht Operationen der
// beiden Dienste die Verteilungen beider Span-Namen leer und schreiben keine Zeile eines Spans.
func TestWorkspaceUndSnapshotOhneSpan(t *testing.T) {
	pakete := ohneSpanPakete()
	for _, p := range pakete {
		dateien, err := filepath.Glob(filepath.Join(p, "*.go"))
		if err != nil || len(dateien) == 0 {
			t.Fatalf("Paket %s: %v, %d Dateien", p, err, len(dateien))
		}
		for _, d := range dateien {
			if strings.HasSuffix(d, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(token.NewFileSet(), d, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.SelectorExpr:
					if x.Sel.Name == "StartSpan" || x.Sel.Name == "CountError" {
						t.Errorf("%s: Selektor %s", d, x.Sel.Name)
					}
					if id, ok := x.X.(*ast.Ident); ok && id.Name == "support" && x.Sel.Name == "Telemetry" {
						t.Errorf("%s: Typ support.Telemetry", d)
					}
				}
				return true
			})
		}
	}
	if len(pakete) != 2 {
		t.Fatalf("%d Pakete, erwartet 2", len(pakete))
	}
	w := baueWurzel(t, modusMock, &bereinigungFake{})
	ctx := context.Background()
	ws, snaps := w.h.Workspaces(), w.h.Snapshots()
	neu, err := ws.Create(ctx, "bsl-1", "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	snap, err := snaps.Create(ctx, neu.ID, "stand")
	if err != nil {
		t.Fatalf("Snapshot Create: %v", err)
	}
	operationen := []struct {
		name string
		op   func() error
	}{
		{"Get", func() error { _, err := ws.Get(ctx, neu.ID); return err }},
		{"List", func() error { _, err := ws.List(ctx); return err }},
		{"Reset", func() error { _, err := ws.Reset(ctx, neu.ID); return err }},
		{"Restore", func() error { _, err := ws.Restore(ctx, neu.ID, snap.ID); return err }},
		{"Snapshot List", func() error { _, err := snaps.List(ctx, neu.ID); return err }},
		{"Delete", func() error { _, err := ws.Delete(ctx, neu.ID); return err }},
	}
	for _, o := range operationen {
		if err := o.op(); err != nil {
			t.Errorf("%s: %v", o.name, err)
		}
	}
	st := w.adapter[0].Snapshot()
	for _, n := range []support.SpanName{support.SpanMockQuery, support.SpanWorkspaceMutation} {
		if st.Durations[n].Count != 0 {
			t.Errorf("Span %s %d-mal beendet, erwartet 0", n, st.Durations[n].Count)
		}
	}
	if len(st.Spans) != 0 || strings.Contains(w.senke.String(), `"ereignis":"span"`) || len(operationen)+2 != 8 {
		t.Errorf("Spans %d, Senke %q, Operationen %d (erwartet 8 mit Create und Snapshot Create)", len(st.Spans), w.senke.String(), len(operationen)+2)
	}
}
