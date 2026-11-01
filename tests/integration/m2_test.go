package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

const (
	serviceID  = "ref"
	acceptJSON = "application/json"
	acceptXML  = "application/xml"
)

// anfrage ist ein Request des Clients: Pfad mit Query und der Accept-Header (steht in MATCH_VARY_HEADERS).
type anfrage struct {
	pfad   string
	accept string
}

// antwort hält, was der Client sah.
type antwort struct {
	status int
	header http.Header
	body   []byte
}

type client struct {
	t     *testing.T
	c     *http.Client
	basis string
}

func neuerClient(t *testing.T, adresse string) *client {
	t.Helper()
	tr := &http.Transport{DisableKeepAlives: true}
	t.Cleanup(tr.CloseIdleConnections)
	return &client{t: t, c: &http.Client{Timeout: 20 * time.Second, Transport: tr}, basis: "http://" + adresse}
}

// hole schickt den GET mit Accept, Authorization (GEHEIM-A) und den Zusatz-Headern.
func (c *client) hole(a anfrage, zusatz map[string]string) antwort {
	c.t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.basis+a.pfad, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Accept", a.accept)
	req.Header.Set("Authorization", "Bearer "+geheimA)
	for k, v := range zusatz {
		req.Header.Set(k, v)
	}
	resp, err := c.c.Do(req)
	if err != nil {
		c.t.Fatalf("%s: %v", a.pfad, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("%s: Body: %v", a.pfad, err)
	}
	return antwort{status: resp.StatusCode, header: resp.Header, body: body}
}

// konfig lädt die Konfiguration wie im Betrieb; nur Listen ist auf Port 0 überschrieben. Auch der
// Replay-Lauf nennt den Upstream: er könnte den Referenzdienst erreichen, und der Zähler zeigt, dass er es nicht tut.
func konfig(t *testing.T, modus, dir, upstream string) bootstrap.Konfiguration {
	t.Helper()
	env := []string{"MODE=" + modus, "LISTEN=127.0.0.1:8080", "DATA_DIR=" + dir, "SERVICE_ID=" + serviceID, "SERVICE_ROOT_PATH=/odata"}
	if upstream != "" {
		env = append(env, "UPSTREAM_URL="+upstream)
	}
	k, err := bootstrap.LadeKonfiguration(env)
	if err != nil {
		t.Fatalf("Konfiguration: %v", err)
	}
	k.Listen = "127.0.0.1:0"
	return k
}

func hochgefahren(ctx context.Context, t *testing.T, k bootstrap.Konfiguration) *bootstrap.Handle {
	t.Helper()
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		t.Fatalf("Build %s: %v", k.Modus, err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(context.WithoutCancel(ctx)) })
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatalf("Hochfahren %s: %v", k.Modus, err)
	}
	return h
}

// ausnahme ist die Menge der Header, die der Vergleich von Aufnahme und Wiedergabe auslässt
// (SPEC-135 Nr. 13 a bis d): Länge und Coding der Leitung, vom Go-Server gesetztes Date,
// Hop-by-Hop, jeder Name aus VolatileHeaders und RedactHeaders.
func ausnahme(k bootstrap.Konfiguration) map[string]bool {
	m := map[string]bool{}
	for _, n := range []string{"Content-Length", "Content-Encoding", "Date", "Connection", "Keep-Alive", "Proxy-Connection",
		"Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		m[http.CanonicalHeaderKey(n)] = true
	}
	for _, n := range slices.Concat(k.VolatileHeaders, k.RedactHeaders) {
		m[http.CanonicalHeaderKey(n)] = true
	}
	return m
}

func verglichen(h http.Header, aus map[string]bool) map[string][]string {
	out := map[string][]string{}
	for n, v := range h {
		if !aus[http.CanonicalHeaderKey(n)] {
			out[http.CanonicalHeaderKey(n)] = v
		}
	}
	return out
}

// gleich nennt den ersten Unterschied zwischen Aufnahme und Wiedergabe (Status, Body, Header außer der
// Ausnahme, redigierte Header fehlen in der Wiedergabe), sonst "".
func gleich(k bootstrap.Konfiguration, aufnahme, wieder antwort) string {
	aus := ausnahme(k)
	switch {
	case aufnahme.status != wieder.status:
		return "Status"
	case !bytes.Equal(aufnahme.body, wieder.body):
		return "Body"
	}
	a, w := verglichen(aufnahme.header, aus), verglichen(wieder.header, aus)
	if len(a) != len(w) {
		return "Menge der Header"
	}
	for n, v := range a {
		if !slices.Equal(v, w[n]) {
			return "Header " + n
		}
	}
	for _, n := range k.RedactHeaders {
		if _, da := wieder.header[http.CanonicalHeaderKey(n)]; da {
			return "redigierter Header " + n + " in der Wiedergabe"
		}
	}
	return ""
}

type fundstelle struct{ datei, geheimnis string }

// durchsuche liest jede Datei unter dir (auch -wal und -shm) auf Rohbytes und liefert die Dateien und
// die Stellen, an denen ein Marker steht.
func durchsuche(t *testing.T, dir string, marker []string) (dateien []string, funde []fundstelle) {
	t.Helper()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		dateien = append(dateien, rel)
		for _, m := range marker {
			if bytes.Contains(b, []byte(m)) {
				funde = append(funde, fundstelle{rel, m})
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Suche unter %s: %v", dir, err)
	}
	return dateien, funde
}

// TestDurchstichM2 ist der Ablauf Schritt 1 bis 10: Aufzeichnen gegen den Referenzdienst, Session
// beenden, importieren, Workspace anlegen, abspielen bei lebendem Referenzdienst, Referenzdienst
// beenden, noch einmal abspielen (LH-FA-04, LH-FA-05).
func TestDurchstichM2(t *testing.T) {
	ctx := context.Background()
	ref := neuerReferenzdienst(t)
	dir := t.TempDir()

	// 1. Build(record), Hochfahren, erste Session.
	kRec := konfig(t, "record", dir, ref.URL()+"/odata")
	rec := hochgefahren(ctx, t, kRec)
	sessions, err := rec.Sessions().List(ctx)
	if err != nil || len(sessions) != 1 || sessions[0].ServiceID != serviceID || sessions[0].Status != model.SessionOpen {
		t.Fatalf("erste Session: %+v, %v", sessions, err)
	}
	sessionID := sessions[0].ID

	// 2. Aufnahme: der Client fragt über den Recorder.
	rc := neuerClient(t, rec.Adresse())
	var anfragen []anfrage
	var aufgenommen []antwort
	nimm := func(a anfrage) antwort {
		w := rc.hole(a, nil)
		anfragen = append(anfragen, a)
		aufgenommen = append(aufgenommen, w)
		return w
	}
	nimm(anfrage{"/odata/$metadata", acceptXML})
	seite1 := nimm(anfrage{"/odata/Products", acceptJSON})
	var kopf struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if err := json.Unmarshal(seite1.body, &kopf); err != nil || kopf.NextLink == "" {
		t.Fatalf("Seite 1 ohne nextLink: %s, %v", seite1.body, err)
	}
	folge, err := url.Parse(kopf.NextLink)
	if err != nil || folge.Host != strings.TrimPrefix(ref.URL(), "http://") {
		t.Fatalf("nextLink %q zeigt nicht auf den Referenzdienst: %v", kopf.NextLink, err)
	}
	nimm(anfrage{folge.RequestURI(), acceptJSON})
	nimm(anfrage{"/odata/Products(1)", acceptJSON})
	for range zaehlerFolge {
		nimm(anfrage{"/odata/Products/$count", acceptJSON})
	}

	// Gegenprobe der Zähler und der Geheimnisse (M2, M3): jeder Request erreichte den Referenzdienst genau
	// einmal, GEHEIM-A erreichte ihn, GEHEIM-B erreichte den Client.
	if got := ref.requests.Load(); got != int64(len(anfragen)) {
		t.Errorf("Requests am Referenzdienst %d, Requests des Clients %d", got, len(anfragen))
	}
	if ref.verbindungen.Load() < 1 {
		t.Error("der Zähler der Verbindungen zählt nichts")
	}
	if got := ref.mitGeheimnis.Load(); got != int64(len(anfragen)) {
		t.Errorf("Requests mit GEHEIM-A am Referenzdienst %d, erwartet %d", got, len(anfragen))
	}
	for i, w := range aufgenommen {
		if w.status != 200 {
			t.Errorf("Aufnahme %d (%s): Status %d", i, anfragen[i].pfad, w.status)
		}
		if !strings.Contains(strings.Join(w.header.Values("Set-Cookie"), ";"), geheimB) {
			t.Errorf("Aufnahme %d (%s): Set-Cookie trägt GEHEIM-B nicht", i, anfragen[i].pfad)
		}
	}

	// 3. Beenden.
	beendet, err := rec.Sessions().Beenden(ctx, sessionID)
	if err != nil || beendet.Status != model.SessionComplete {
		t.Fatalf("Beenden: %+v, %v", beendet, err)
	}

	// 4. Workspace aus unbekannter Baseline scheitert (M9), Import, Workspace aus der Baseline.
	if rec.Workspaces() == nil || rec.Import() == nil {
		t.Fatalf("Zugänge im Modus record: Workspaces %v, Import %v", rec.Workspaces(), rec.Import())
	}
	_, err = rec.Workspaces().Create(ctx, "unbekannt", "")
	var fe *inbound.Fehler
	if !errors.As(err, &fe) || fe.Code != model.CodeE7002 {
		t.Fatalf("Create aus unbekannter Baseline: %v", err)
	}
	erg, err := rec.Import().Import(ctx, inbound.ImportAnfrage{
		ImportID: "m2-import", ServiceID: serviceID, SourceSessionIDs: []string{sessionID}, MetadataHash: csdlHash(),
	})
	if err != nil || !erg.Neu || erg.Import.BaselineSnapshotID == "" {
		t.Fatalf("Import: %+v, %v", erg, err)
	}
	var report struct {
		Coverage []struct {
			EntitySet string `json:"entity_set"`
			Status    string `json:"status"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal([]byte(erg.Import.ReportJSON), &report); err != nil {
		t.Fatalf("Report: %v", err)
	}
	if len(report.Coverage) != 1 {
		t.Fatalf("Coverage hat %d Einträge: %s", len(report.Coverage), erg.Import.ReportJSON)
	}
	for _, e := range report.Coverage {
		if e.EntitySet != "Products" || e.Status != "complete" {
			t.Errorf("Coverage %s: %s, erwartet Products complete", e.EntitySet, e.Status)
		}
	}
	ws, err := rec.Workspaces().Create(ctx, erg.Import.BaselineSnapshotID, "")
	if err != nil || ws.ID == "" || ws.BaselineSnapshotID != erg.Import.BaselineSnapshotID {
		t.Fatalf("Create aus der Baseline: %+v, %v", ws, err)
	}

	// Suche vor dem Herunterfahren (die Ablage ist offen, das WAL liegt neben der Datenbank) und danach.
	marker := []string{geheimA, geheimB}
	vorher, funde := durchsuche(t, dir, marker)
	if len(funde) > 0 {
		t.Errorf("Geheimnis vor dem Herunterfahren in %v", funde)
	}

	// 5. Herunterfahren(record); MetadataHash gleich dem Orakel (M7).
	if err := rec.Herunterfahren(ctx); err != nil {
		t.Fatalf("Herunterfahren(record): %v", err)
	}
	db, err := sqlite.Open(ctx, filepath.Join(dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		t.Fatalf("Ablage öffnen: %v", err)
	}
	svc, err := db.Metadata().Get(ctx, serviceID)
	_ = db.Close()
	if err != nil || svc.MetadataHash != csdlHash() {
		t.Errorf("MetadataHash %q (%v), erwartet %q", svc.MetadataHash, err, csdlHash())
	}

	// 6. Kein Geheimnis in einer Datei, und der Sucher findet einen Marker, den es geben muss (M3).
	nachher, funde := durchsuche(t, dir, append(slices.Clone(marker), "Anvil"))
	var blobFunde int
	for _, f := range funde {
		if f.geheimnis == "Anvil" && strings.HasPrefix(f.datei, "blobs"+string(filepath.Separator)) {
			blobFunde++
			continue
		}
		if f.geheimnis != "Anvil" {
			t.Errorf("Geheimnis %s in %s", f.geheimnis, f.datei)
		}
	}
	if blobFunde == 0 {
		t.Errorf("der Sucher findet den Produktnamen in keinem Blob (Dateien: %v)", nachher)
	}
	if !slices.Contains(vorher, "odata-recorder.db") || !slices.Contains(nachher, "odata-recorder.db") {
		t.Errorf("der Sucher sieht die Datenbank nicht: %v, %v", vorher, nachher)
	}

	// 7. Build(replay) auf derselben Ablage; der Workspace ist da. Zähler lesen.
	kRep := konfig(t, "replay", dir, ref.URL()+"/odata")
	rep := hochgefahren(ctx, t, kRep)
	if rep.Import() != nil {
		t.Error("Import() im Modus replay ist nicht nil")
	}
	if rep.Workspaces() == nil {
		t.Fatal("Workspaces() im Modus replay ist nil")
	}
	if got, err := rep.Workspaces().Get(ctx, ws.ID); err != nil || got.ID != ws.ID || got.BaselineSnapshotID != ws.BaselineSnapshotID {
		t.Fatalf("Get im Replay-Handle: %+v, %v", got, err)
	}
	pc := neuerClient(t, rep.Adresse())
	zuordnung := func(replaySession string) map[string]string {
		return map[string]string{
			"X-Recorder-Workspace": ws.ID, "X-Recorder-Recording-Session": sessionID, "X-Recorder-Replay-Session": replaySession,
		}
	}
	zaehlerVor := [2]int64{ref.verbindungen.Load(), ref.requests.Load()}
	durchgang := func(name, replaySession string) {
		t.Helper()
		z := zuordnung(replaySession)
		naechster := 0
		for i, a := range anfragen {
			w := pc.hole(a, z)
			if d := gleich(kRep, aufgenommen[i], w); d != "" {
				t.Errorf("Durchgang %s, Request %d (%s): Wiedergabe weicht von der Aufnahme ab: %s", name, i, a.pfad, d)
			}
			// M4: die Folge von $count nach den Werten der Fixture, unabhängig von der Aufnahme.
			if strings.HasSuffix(a.pfad, "/$count") {
				if soll := fmt.Sprintf(`{"count":%d}`, zaehlerFolge[naechster]); string(w.body) != soll {
					t.Errorf("Durchgang %s: %d. $count %s, erwartet %s", name, naechster+1, w.body, soll)
				}
				naechster++
			}
		}
		if naechster != len(zaehlerFolge) {
			t.Errorf("Durchgang %s: %d Requests auf $count, erwartet %d", name, naechster, len(zaehlerFolge))
		}
		// Ein nie aufgezeichneter Request ist 404 ODR-E5001, der vierte $count nach drei verbrauchten 410 ODR-E5002.
		unbekannt := pc.hole(anfrage{"/odata/Products(99)", acceptJSON}, z)
		if unbekannt.status != 404 || !bytes.Contains(unbekannt.body, []byte("ODR-E5001")) {
			t.Errorf("Durchgang %s: Products(99): %d %s", name, unbekannt.status, unbekannt.body)
		}
		erschoepft := pc.hole(anfrage{"/odata/Products/$count", acceptJSON}, z)
		if erschoepft.status != 410 || !bytes.Contains(erschoepft.body, []byte("ODR-E5002")) {
			t.Errorf("Durchgang %s: vierter $count: %d %s", name, erschoepft.status, erschoepft.body)
		}
	}
	keinKontakt := func(name string) {
		t.Helper()
		if got := [2]int64{ref.verbindungen.Load(), ref.requests.Load()}; got != zaehlerVor {
			t.Errorf("Durchgang %s: Zähler (Verbindungen, Requests) %v, vor dem Replay %v", name, got, zaehlerVor)
		}
	}

	// 8. Durchgang A: der Referenzdienst lebt, der Replay darf ihn nicht erreichen.
	durchgang("A", "rep-a")
	keinKontakt("A")

	// 9. Der Referenzdienst endet; Durchgang B liefert dieselben Antworten.
	ref.srv.Close()
	durchgang("B", "rep-b")
	keinKontakt("B")
}
