package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

const (
	m3Basis = "http://127.0.0.1:9999"
	m3Pfad  = "/odata/Products"
)

// sende schickt eine Anfrage beliebiger Methode mit Body; host überschreibt den Host-Header, wenn er nicht leer ist.
func (c *client) sende(methode, pfad, body, host string, header map[string]string) antwort {
	c.t.Helper()
	req, err := http.NewRequest(methode, c.basis+pfad, strings.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	if host != "" {
		req.Host = host
	}
	req.Header.Set("Accept", acceptJSON)
	if body != "" {
		req.Header.Set("Content-Type", acceptJSON)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := c.c.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", methode, pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("%s %s: Body: %v", methode, pfad, err)
	}
	return antwort{status: resp.StatusCode, header: resp.Header, body: b}
}

// m3Welt ist der Zustand nach der Aufnahme: zwei Baselines in DATA_DIR (Products vollständig, Products nur als Entity gesehen, also
// nicht vollständig), je ein Workspace, und der Referenzdienst ist beendet.
type m3Welt struct {
	t           *testing.T
	dir         string
	ref         *referenzdienst
	ws, wsTeil  model.Workspace
	nachDemEnde [2]int64
}

// m3Aufbau nimmt auf, importiert zweimal, legt je einen Workspace an, fährt den Recorder herunter und beendet den Referenzdienst.
func m3Aufbau(t *testing.T) *m3Welt {
	t.Helper()
	ctx := context.Background()
	d := &durchstich{t: t, ref: neuerReferenzdienst(t), dir: t.TempDir()}
	d.kRec = konfig(t, "record", d.dir, d.ref.URL()+"/odata")
	rec := hochgefahren(ctx, t, d.kRec)
	d.nimmAuf(rec)
	d.beendeUndImportiere(rec)
	// zweite Aufnahme: nur $metadata und eine Entity, kein ungefilterter Lauf über die Collection
	s2, err := rec.Sessions().Create(ctx, serviceID, "")
	if err != nil {
		t.Fatalf("zweite Session: %v", err)
	}
	rc := neuerClient(t, rec.Adresse())
	for _, a := range []anfrage{{"/odata/$metadata", acceptXML}, {"/odata/Products(1)", acceptJSON}} {
		if w := rc.hole(a, nil); w.status != 200 {
			t.Fatalf("zweite Aufnahme %s: %d", a.pfad, w.status)
		}
	}
	if _, err := rec.Sessions().Beenden(ctx, s2.ID); err != nil {
		t.Fatalf("Beenden der zweiten Session: %v", err)
	}
	erg, err := rec.Import().Import(ctx, inbound.ImportAnfrage{
		ImportID: "m3-import-2", ServiceID: serviceID, SourceSessionIDs: []string{s2.ID}, MetadataHash: csdlHash(),
	})
	if err != nil || !erg.Neu || !strings.Contains(erg.Import.ReportJSON, `"partial"`) {
		t.Fatalf("zweiter Import: %+v, %v", erg, err)
	}
	teil, err := rec.Workspaces().Create(ctx, erg.Import.BaselineSnapshotID, "")
	if err != nil {
		t.Fatalf("Workspace aus der zweiten Baseline: %v", err)
	}
	if err := rec.Herunterfahren(ctx); err != nil {
		t.Fatalf("Herunterfahren(record): %v", err)
	}
	d.ref.srv.Close()
	return &m3Welt{t: t, dir: d.dir, ref: d.ref, ws: d.ws, wsTeil: teil,
		nachDemEnde: [2]int64{d.ref.verbindungen.Load(), d.ref.requests.Load()}}
}

// startMock fährt den Modus mock auf DATA_DIR hoch; extra sind weitere Zeilen der Umgebung.
func (w *m3Welt) startMock(extra ...string) (*bootstrap.Handle, *client) {
	w.t.Helper()
	h := hochgefahren(context.Background(), w.t, konfig(w.t, "mock", w.dir, "", extra...))
	return h, neuerClient(w.t, h.Adresse())
}

func kopf(ws model.Workspace) map[string]string {
	return map[string]string{"X-Recorder-Workspace": ws.ID}
}

func jsonVon(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("kein JSON: %s", b)
	}
	return m
}

// namen liefert die Namen der Entities einer Collection in ihrer Folge.
func namen(t *testing.T, b []byte) []string {
	t.Helper()
	var out []string
	werte, _ := jsonVon(t, b)["value"].([]any)
	for _, w := range werte {
		out = append(out, w.(map[string]any)["Name"].(string))
	}
	return out
}

// TestM3 ist der M3-Test (LH-FA-12): aufgenommen, importiert, Workspace angelegt, Referenzdienst beendet, dann Modus mock.
func TestM3(t *testing.T) {
	w := m3Aufbau(t)
	t.Run("Mock", w.mockLauf)
	t.Run("Coverage", w.coverage)
	t.Run("Adressen", w.adressen)
	t.Run("Neustart", w.neustart)
	if got := [2]int64{w.ref.verbindungen.Load(), w.ref.requests.Load()}; got != w.nachDemEnde {
		t.Errorf("Zähler des Referenzdienstes (Verbindungen, Requests) %v nach dem Mock, %v nach dem Beenden", got, w.nachDemEnde)
	}
}

// mockLauf ist der Happy Path: $metadata, kombinierte Abfrage, POST, PATCH, DELETE mit Lesen nach jedem Schritt, Snapshot, weitere
// Änderung, Restore, Reset (LH-FA-06 bis LH-FA-10).
func (w *m3Welt) mockLauf(t *testing.T) {
	ctx := context.Background()
	h, c := w.startMock()
	z := kopf(w.ws)
	liste := func() []string {
		r := c.sende("GET", m3Pfad+"?$orderby=Id", "", "", z)
		if r.status != 200 {
			t.Fatalf("Collection: %d %s", r.status, r.body)
		}
		return namen(t, r.body)
	}
	if r := c.sende("GET", "/odata/$metadata", "", "", z); r.status != 200 || string(r.body) != csdlText || !strings.HasPrefix(r.header.Get("Content-Type"), acceptXML) {
		t.Errorf("$metadata: %d %q", r.status, r.header.Get("Content-Type"))
	}
	// $filter, $orderby, $top, $skip und $count in einer Abfrage: Id >= 1 gibt Anvil und Bolt, absteigend nach Name Bolt, Anvil, übersprungen eins.
	r := c.sende("GET", m3Pfad+"?$filter=Id%20ge%201&$orderby=Name%20desc&$top=2&$skip=1&$count=true", "", "", z)
	if r.status != 200 || !slices.Equal(namen(t, r.body), []string{"Anvil"}) || jsonVon(t, r.body)["@odata.count"] != float64(2) {
		t.Errorf("kombinierte Abfrage: %d %s", r.status, r.body)
	}
	if got := liste(); !slices.Equal(got, []string{"Anvil", "Bolt"}) {
		t.Fatalf("Baseline: %v", got)
	}
	if r := c.sende("POST", m3Pfad, `{"Id":3,"Name":"Cutter"}`, "", z); r.status != 201 || !strings.HasSuffix(r.header.Get("Location"), "/odata/Products(3)") {
		t.Fatalf("POST: %d %s %v", r.status, r.body, r.header)
	}
	if got := liste(); !slices.Equal(got, []string{"Anvil", "Bolt", "Cutter"}) {
		t.Errorf("nach POST: %v", got)
	}
	if r := c.sende("PATCH", m3Pfad+"(3)", `{"Name":"Cutter2"}`, "", z); r.status != 204 {
		t.Fatalf("PATCH: %d %s", r.status, r.body)
	}
	if got := liste(); !slices.Equal(got, []string{"Anvil", "Bolt", "Cutter2"}) {
		t.Errorf("nach PATCH: %v", got)
	}
	if r := c.sende("DELETE", m3Pfad+"(1)", "", "", z); r.status != 204 {
		t.Fatalf("DELETE: %d %s", r.status, r.body)
	}
	if got := liste(); !slices.Equal(got, []string{"Bolt", "Cutter2"}) {
		t.Errorf("nach DELETE: %v", got)
	}
	if r := c.sende("GET", m3Pfad+"(1)", "", "", z); r.status != 404 {
		t.Errorf("gelöschte Entity: %d", r.status)
	}
	snap, err := h.Snapshots().Create(ctx, w.ws.ID, "m3")
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if r := c.sende("POST", m3Pfad, `{"Id":4,"Name":"Drill"}`, "", z); r.status != 201 {
		t.Fatalf("POST nach dem Snapshot: %d %s", r.status, r.body)
	}
	if got := liste(); !slices.Equal(got, []string{"Bolt", "Cutter2", "Drill"}) {
		t.Errorf("vor Restore: %v", got)
	}
	if _, err := h.Workspaces().Restore(ctx, w.ws.ID, snap.ID); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got := liste(); !slices.Equal(got, []string{"Bolt", "Cutter2"}) {
		t.Errorf("nach Restore: %v", got)
	}
	if _, err := h.Workspaces().Reset(ctx, w.ws.ID); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if got := liste(); !slices.Equal(got, []string{"Anvil", "Bolt"}) {
		t.Errorf("nach Reset: %v", got)
	}
}

// coverage fängt A14: strikt und nicht strikt × Baseline vollständig und nicht vollständig × Collection und Anzahl (SPEC-197 Nr. 2).
func (w *m3Welt) coverage(t *testing.T) {
	n := 0
	for _, strikt := range []string{"true", "false"} {
		h, c := w.startMock("STRICT_COVERAGE=" + strikt)
		for _, ws := range []struct {
			ws   model.Workspace
			cov  string
			name string
		}{{w.ws, "complete", "vollständig"}, {w.wsTeil, "partial", "nicht vollständig"}} {
			for _, pfad := range []string{m3Pfad, m3Pfad + "/$count"} {
				n++
				r := c.sende("GET", pfad, "", "", kopf(ws.ws))
				ablehnen := strikt == "true" && ws.cov != "complete"
				switch {
				case ablehnen && (r.status != 422 || !strings.Contains(string(r.body), "ODR-E6003") || r.header.Get("X-Recorder-Coverage") != ""):
					t.Errorf("strikt=%s %s %s: %d %s", strikt, ws.name, pfad, r.status, r.body)
				case !ablehnen && (r.status != 200 || r.header.Get("X-Recorder-Coverage") != ws.cov):
					t.Errorf("strikt=%s %s %s: %d, Header %q", strikt, ws.name, pfad, r.status, r.header.Get("X-Recorder-Coverage"))
				}
			}
		}
		if err := h.Herunterfahren(context.Background()); err != nil { // der nächste Lauf braucht die Ablage frei
			t.Fatalf("Herunterfahren: %v", err)
		}
	}
	if n != 8 {
		t.Errorf("Menge hat %d Elemente, erwartet 8", n)
	}
}

// adressen fängt A15: Location und @odata.context tragen PUBLIC_BASE_URL plus Root, mit Basis ohne und mit Pfad, nie den Host des Clients;
// die Fehlerantwort trägt Content-Type und Korrelations-Kennung und kein Kennzeichen der Version (SPEC-136 Nr. 8, SPEC-198 Nr. 6).
func (w *m3Welt) adressen(t *testing.T) {
	for i, basis := range []string{m3Basis, m3Basis + "/api"} {
		h, c := w.startMock("PUBLIC_BASE_URL=" + basis)
		z := kopf(w.ws)
		post := c.sende("POST", m3Pfad, `{"Id":`+string(rune('5'+i))+`,"Name":"Neu"}`, "evil.example", z)
		soll := basis + "/odata/Products(" + string(rune('5'+i)) + ")"
		if post.status != 201 || post.header.Get("Location") != soll {
			t.Errorf("Basis %s: POST %d, Location %q, erwartet %q", basis, post.status, post.header.Get("Location"), soll)
		}
		lies := c.sende("GET", m3Pfad, "", "evil.example", z)
		if ctx, _ := jsonVon(t, lies.body)["@odata.context"].(string); lies.status != 200 || !strings.HasPrefix(ctx, basis+"/odata/$metadata") || strings.Contains(string(lies.body), "evil.example") {
			t.Errorf("Basis %s: Collection %d %s", basis, lies.status, lies.body)
		}
		fehler := c.sende("GET", m3Pfad+"(99)", "", "", z)
		if fehler.status != 404 || !strings.Contains(string(fehler.body), "ODR-E6008") || fehler.header.Get("Content-Type") != "application/json" ||
			fehler.header.Get("X-Correlation-Id") == "" || len(fehler.header.Values("OData-Version")) != 0 || len(fehler.header.Values("DataServiceVersion")) != 0 {
			t.Errorf("Basis %s: Fehlerantwort %d %s %v", basis, fehler.status, fehler.body, fehler.header)
		}
		if err := h.Herunterfahren(context.Background()); err != nil {
			t.Fatalf("Herunterfahren: %v", err)
		}
	}
}

// neustart fängt A12 im Prozess: Build, Hochfahren, Mutation, Herunterfahren, Build auf derselben Ablage; die Mutation ist lesbar,
// auch nach einem zweiten Neustart ohne Änderung (SPEC-139 Nr. 11 f, SPEC-049 Nr. 7).
func (w *m3Welt) neustart(t *testing.T) {
	ctx := context.Background()
	h, c := w.startMock()
	ws, err := h.Workspaces().Create(ctx, w.ws.BaselineSnapshotID, "")
	if err != nil {
		t.Fatalf("Workspace: %v", err)
	}
	if r := c.sende("POST", m3Pfad, `{"Id":7,"Name":"Neustart"}`, "", kopf(ws)); r.status != 201 {
		t.Fatalf("POST: %d %s", r.status, r.body)
	}
	vorher, err := h.Workspaces().Get(ctx, ws.ID)
	if err != nil || vorher.Revision < 1 {
		t.Fatalf("Workspace nach der Mutation: %+v, %v", vorher, err)
	}
	for lauf := 1; lauf <= 2; lauf++ {
		if err := h.Herunterfahren(ctx); err != nil {
			t.Fatalf("Herunterfahren vor Neustart %d: %v", lauf, err)
		}
		h, c = w.startMock()
		if r := c.sende("GET", m3Pfad+"(7)", "", "", kopf(ws)); r.status != 200 || !strings.Contains(string(r.body), "Neustart") {
			t.Errorf("Neustart %d: Entity 7: %d %s", lauf, r.status, r.body)
		}
		if got, err := h.Workspaces().Get(ctx, ws.ID); err != nil || got.Revision != vorher.Revision {
			t.Errorf("Neustart %d: Workspace %+v, %v", lauf, got, err)
		}
	}
}
