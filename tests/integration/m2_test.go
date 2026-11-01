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

// durchstich hält den Zustand des Ablaufs von Schritt 1 bis 10; jede Phase ist eine Methode.
type durchstich struct {
	t           *testing.T
	ref        *referenzdienst
	dir         string
	kRec, kRep  bootstrap.Konfiguration
	sessionID   string
	anfragen    []anfrage
	aufgenommen []antwort
	ws          model.Workspace
}

// TestDurchstichM2 ist der Ablauf Schritt 1 bis 10: Aufzeichnen gegen den Referenzdienst, Session
// beenden, importieren, Workspace anlegen, abspielen bei lebendem Referenzdienst, Referenzdienst
// beenden, noch einmal abspielen (LH-FA-04, LH-FA-05).
func TestDurchstichM2(t *testing.T) {
	ctx := context.Background()
	d := &durchstich{t: t, ref: neuerReferenzdienst(t), dir: t.TempDir()}
	d.kRec = konfig(t, "record", d.dir, d.ref.URL()+"/odata")
	d.kRep = konfig(t, "replay", d.dir, d.ref.URL()+"/odata")
	rec := hochgefahren(ctx, t, d.kRec) // Schritt 1
	d.nimmAuf(rec)                        // Schritt 2
	d.pruefeAufnahme()
	d.beendeUndImportiere(rec) // Schritte 3 und 4
	vorher := d.sucheGeheimnisse(0)
	if err := rec.Herunterfahren(ctx); err != nil { // Schritt 5
		t.Fatalf("Herunterfahren(record): %v", err)
	}
	d.pruefeMetadataHash()
	d.sucheGeheimnisseNachHerunterfahren(vorher) // Schritt 6
	d.spieleAb()                                 // Schritte 7 bis 10
}

// nimmAuf ist Schritt 2: der Client fragt über den Recorder; die erste Session ist offen (Schritt 1).
func (d *durchstich) nimmAuf(rec *bootstrap.Handle) {
	t := d.t
	sessions, err := rec.Sessions().List(context.Background())
	if err != nil || len(sessions) != 1 || sessions[0].ServiceID != serviceID || sessions[0].Status != model.SessionOpen {
		t.Fatalf("erste Session: %+v, %v", sessions, err)
	}
	d.sessionID = sessions[0].ID
	rc := neuerClient(t, rec.Adresse())
	nimm := func(a anfrage) antwort {
		w := rc.hole(a, nil)
		d.anfragen = append(d.anfragen, a)
		d.aufgenommen = append(d.aufgenommen, w)
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
	if err != nil || folge.Host != strings.TrimPrefix(d.ref.URL(), "http://") {
		t.Fatalf("nextLink %q zeigt nicht auf den Referenzdienst: %v", kopf.NextLink, err)
	}
	nimm(anfrage{folge.RequestURI(), acceptJSON})
	nimm(anfrage{"/odata/Products(1)", acceptJSON})
	for range zaehlerFolge() {
		nimm(anfrage{"/odata/Products/$count", acceptJSON})
	}
}

// pruefeAufnahme ist die Gegenprobe der Zähler und der Geheimnisse (M2, M3): jeder Request erreichte den
// Referenzdienst genau einmal, GEHEIM-A erreichte ihn, GEHEIM-B erreichte den Client.
func (d *durchstich) pruefeAufnahme() {
	t, ref := d.t, d.ref
	if got := ref.requests.Load(); got != int64(len(d.anfragen)) {
		t.Errorf("Requests am Referenzdienst %d, Requests des Clients %d", got, len(d.anfragen))
	}
	if ref.verbindungen.Load() < 1 {
		t.Error("der Zähler der Verbindungen zählt nichts")
	}
	if got := ref.mitGeheimnis.Load(); got != int64(len(d.anfragen)) {
		t.Errorf("Requests mit GEHEIM-A am Referenzdienst %d, erwartet %d", got, len(d.anfragen))
	}
	for i, w := range d.aufgenommen {
		if w.status != 200 {
			t.Errorf("Aufnahme %d (%s): Status %d", i, d.anfragen[i].pfad, w.status)
		}
		if !strings.Contains(strings.Join(w.header.Values("Set-Cookie"), ";"), geheimB) {
			t.Errorf("Aufnahme %d (%s): Set-Cookie trägt GEHEIM-B nicht", i, d.anfragen[i].pfad)
		}
	}
}

// beendeUndImportiere sind die Schritte 3 und 4: Beenden, Workspace aus unbekannter Baseline scheitert (M9),
// Import, Report, Workspace aus der Baseline.
func (d *durchstich) beendeUndImportiere(rec *bootstrap.Handle) {
	t, ctx := d.t, context.Background()
	beendet, err := rec.Sessions().Beenden(ctx, d.sessionID)
	if err != nil || beendet.Status != model.SessionComplete {
		t.Fatalf("Beenden: %+v, %v", beendet, err)
	}
	if rec.Workspaces() == nil || rec.Import() == nil {
		t.Fatalf("Zugänge im Modus record: Workspaces %v, Import %v", rec.Workspaces(), rec.Import())
	}
	_, err = rec.Workspaces().Create(ctx, "unbekannt", "")
	var fe *inbound.Fehler
	if !errors.As(err, &fe) || fe.Code != model.CodeE7002 {
		t.Fatalf("Create aus unbekannter Baseline: %v", err)
	}
	erg, err := rec.Import().Import(ctx, inbound.ImportAnfrage{
		ImportID: "m2-import", ServiceID: serviceID, SourceSessionIDs: []string{d.sessionID}, MetadataHash: csdlHash(),
	})
	if err != nil || !erg.Neu || erg.Import.BaselineSnapshotID == "" {
		t.Fatalf("Import: %+v, %v", erg, err)
	}
	d.pruefeCoverage(erg.Import.ReportJSON)
	d.ws, err = rec.Workspaces().Create(ctx, erg.Import.BaselineSnapshotID, "")
	if err != nil || d.ws.ID == "" || d.ws.BaselineSnapshotID != erg.Import.BaselineSnapshotID {
		t.Fatalf("Create aus der Baseline: %+v, %v", d.ws, err)
	}
}

// pruefeCoverage liest den Report (M8): genau ein Eintrag, Products, complete.
func (d *durchstich) pruefeCoverage(reportJSON string) {
	var report struct {
		Coverage []struct {
			EntitySet string `json:"entity_set"`
			Status    string `json:"status"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		d.t.Fatalf("Report: %v", err)
	}
	if len(report.Coverage) != 1 {
		d.t.Fatalf("Coverage hat %d Einträge: %s", len(report.Coverage), reportJSON)
	}
	for _, e := range report.Coverage {
		if e.EntitySet != "Products" || e.Status != "complete" {
			d.t.Errorf("Coverage %s: %s, erwartet Products complete", e.EntitySet, e.Status)
		}
	}
}

// sucheGeheimnisse liest DATA_DIR und meldet jedes Geheimnis in einer Datei; es liefert die gelesenen Dateien.
func (d *durchstich) sucheGeheimnisse(nr int) []string {
	dateien, funde := durchsuche(d.t, d.dir, []string{geheimA, geheimB})
	for _, f := range funde {
		d.t.Errorf("Geheimnis %s in %s (Suche %d)", f.geheimnis, f.datei, nr)
	}
	return dateien
}

// pruefeMetadataHash ist die zweite Hälfte von Schritt 5: die Ablage selbst öffnen, MetadataHash gleich dem Orakel (M7).
func (d *durchstich) pruefeMetadataHash() {
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(d.dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		d.t.Fatalf("Ablage öffnen: %v", err)
	}
	svc, err := db.Metadata().Get(ctx, serviceID)
	_ = db.Close()
	if err != nil || svc.MetadataHash != csdlHash() {
		d.t.Errorf("MetadataHash %q (%v), erwartet %q", svc.MetadataHash, err, csdlHash())
	}
}

// sucheGeheimnisseNachHerunterfahren ist Schritt 6: kein Geheimnis in einer Datei, und der Sucher findet
// einen Marker, den es geben muss (M3).
func (d *durchstich) sucheGeheimnisseNachHerunterfahren(vorher []string) {
	nachher, funde := durchsuche(d.t, d.dir, []string{geheimA, geheimB, "Anvil"})
	blobFunde := 0
	for _, f := range funde {
		switch {
		case f.geheimnis != "Anvil":
			d.t.Errorf("Geheimnis %s in %s", f.geheimnis, f.datei)
		case strings.HasPrefix(f.datei, "blobs"+string(filepath.Separator)):
			blobFunde++
		}
	}
	if blobFunde == 0 {
		d.t.Errorf("der Sucher findet den Produktnamen in keinem Blob (Dateien: %v)", nachher)
	}
	if !slices.Contains(vorher, "odata-recorder.db") || !slices.Contains(nachher, "odata-recorder.db") {
		d.t.Errorf("der Sucher sieht die Datenbank nicht: %v, %v", vorher, nachher)
	}
}

// spieleAb sind die Schritte 7 bis 10: Build(replay) auf derselben Ablage, Durchgang A bei lebendem
// Referenzdienst, Close, Durchgang B.
func (d *durchstich) spieleAb() {
	t, ctx := d.t, context.Background()
	rep := hochgefahren(ctx, t, d.kRep)
	if rep.Import() != nil {
		t.Error("Import() im Modus replay ist nicht nil")
	}
	if rep.Workspaces() == nil {
		t.Fatal("Workspaces() im Modus replay ist nil")
	}
	if got, err := rep.Workspaces().Get(ctx, d.ws.ID); err != nil || got.ID != d.ws.ID || got.BaselineSnapshotID != d.ws.BaselineSnapshotID {
		t.Fatalf("Get im Replay-Handle: %+v, %v", got, err)
	}
	pc := neuerClient(t, rep.Adresse())
	vor := [2]int64{d.ref.verbindungen.Load(), d.ref.requests.Load()}
	d.durchgang(pc, "A", "rep-a")
	d.keinKontakt("A", vor)
	d.ref.srv.Close()
	d.durchgang(pc, "B", "rep-b")
	d.keinKontakt("B", vor)
}

func (d *durchstich) keinKontakt(name string, vor [2]int64) {
	if got := [2]int64{d.ref.verbindungen.Load(), d.ref.requests.Load()}; got != vor {
		d.t.Errorf("Durchgang %s: Zähler (Verbindungen, Requests) %v, vor dem Replay %v", name, got, vor)
	}
}

// durchgang spielt jeden aufgenommenen Request mit den drei Headern der Zuordnung ab und vergleicht
// ihn mit der Aufnahme (M1); die Folge von $count nennt die Fixture (M4).
func (d *durchstich) durchgang(pc *client, name, replaySession string) {
	t := d.t
	t.Helper()
	z := map[string]string{
		"X-Recorder-Workspace": d.ws.ID, "X-Recorder-Recording-Session": d.sessionID, "X-Recorder-Replay-Session": replaySession,
	}
	folge := zaehlerFolge()
	naechster := 0
	for i, a := range d.anfragen {
		w := pc.hole(a, z)
		if diff := gleich(d.kRep, d.aufgenommen[i], w); diff != "" {
			t.Errorf("Durchgang %s, Request %d (%s): Wiedergabe weicht von der Aufnahme ab: %s", name, i, a.pfad, diff)
		}
		if strings.HasSuffix(a.pfad, "/$count") && naechster < len(folge) {
			if soll := fmt.Sprintf(`{"count":%d}`, folge[naechster]); string(w.body) != soll {
				t.Errorf("Durchgang %s: %d. $count %s, erwartet %s", name, naechster+1, w.body, soll)
			}
			naechster++
		}
	}
	if naechster != len(folge) {
		t.Errorf("Durchgang %s: %d Requests auf $count, erwartet %d", name, naechster, len(folge))
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
