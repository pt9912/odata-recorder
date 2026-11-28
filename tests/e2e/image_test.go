package e2e_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/tests/e2e/fixture"
)

// Die zwei Phasen des Neustarttests (SPEC-323 Nr. 4). tools/harness/test-image.sh fährt jede in einem Client-Container im
// Netz-Namensraum des Servers: der Server hört auf den Vorgaben von LISTEN und ADMIN_LISTEN, SERVICE_ID und SERVICE_ROOT_PATH sind
// die Vorgaben. Das Token kommt über die Standardeingabe; ohne Token endet eine Phase mit SKIP, und der Treiber wertet SKIP als rot.
const (
	imageOData  = "127.0.0.1:8080"
	imageAdmin  = "127.0.0.1:8081"
	imageRef    = "127.0.0.1:9000"
	imageDienst = "default"
	// imageBereit ist die Frist der Bereitschaft (SPEC-323 Nr. 5).
	imageBereit = 30 * time.Second
	// imageMarke steht vor den Werten der Phase aufnehmen, die der Treiber der Phase pruefen als imageVariable gibt.
	imageMarke    = "NEUSTART-WERTE="
	imageVariable = "NEUSTART_WERTE"
)

// imageStand sind die vier Antworten, die beide Phasen byte-gleich vergleichen.
type imageStand struct {
	Exchanges, Report, Workspaces, Snapshots []byte
}

// imageWerte sind die Kennungen und der Stand aus der Phase aufnehmen.
type imageWerte struct {
	Session, Import, Workspace, Snapshot string
	Stand                                imageStand
}

// imageToken liest das Token von der Standardeingabe. Es steht in keiner Meldung und keiner Ausgabe der Phasen.
func imageToken(t *testing.T) string {
	t.Helper()
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("Standardeingabe: %v", err)
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		t.Skip("kein Token auf der Standardeingabe: die Phase läuft nur im Client-Container von make test-image")
	}
	return s
}

// imageClient ruft den Admin-Listener des Servers.
type imageClient struct {
	t     *testing.T
	token string
}

// rufe stellt eine Anfrage mit dem Token und verlangt den Status.
func (c imageClient) rufe(status int, methode, pfad, body string) []byte {
	c.t.Helper()
	w, err := sende(imageAdmin, anfrage{methode: methode, pfad: pfad, body: body, header: map[string][]string{"Authorization": {"Bearer " + c.token}}})
	if err != nil {
		c.t.Fatalf("%s %s: %v", methode, pfad, err)
	}
	if w.status != status {
		c.t.Fatalf("%s %s: %d %s, erwartet %d", methode, pfad, w.status, w.body, status)
	}
	return w.body
}

// feld liest das nichtleere Textfeld name aus einem JSON-Objekt.
func (c imageClient) feld(b []byte, name string) string {
	c.t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		c.t.Fatalf("kein JSON: %s", b)
	}
	s, ok := m[name].(string)
	if !ok || s == "" {
		c.t.Fatalf("Antwort ohne %s: %s", name, b)
	}
	return s
}

// bereit wartet höchstens imageBereit auf GET /healthz mit 200.
func (c imageClient) bereit() {
	c.t.Helper()
	for ende := time.Now().Add(imageBereit); ; time.Sleep(100 * time.Millisecond) {
		w, err := sende(imageAdmin, anfrage{pfad: "/healthz"})
		if err == nil && w.status == 200 {
			return
		}
		if time.Now().After(ende) {
			c.t.Fatalf("keine Bereitschaft in %s: %d, %v", imageBereit, w.status, err)
		}
	}
}

// sessions sind die Sessions des Dienstes mit ihrem Status.
func (c imageClient) sessions() map[string]string {
	c.t.Helper()
	b := c.rufe(200, "GET", "/__admin/sessions", "")
	var aus struct {
		Sessions []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(b, &aus); err != nil {
		c.t.Fatalf("Sessions: %v: %s", err, b)
	}
	m := map[string]string{}
	for _, s := range aus.Sessions {
		m[s.ID] = s.Status
	}
	return m
}

// stand liest die vier Antworten zu den Kennungen in w.
func (c imageClient) stand(w imageWerte) imageStand {
	c.t.Helper()
	return imageStand{
		Exchanges:  c.rufe(200, "GET", "/__admin/sessions/"+w.Session+"/exchanges", ""),
		Report:     c.rufe(200, "GET", "/__admin/imports/"+w.Import, ""),
		Workspaces: c.rufe(200, "GET", "/__admin/workspaces", ""),
		Snapshots:  c.rufe(200, "GET", "/__admin/workspaces/"+w.Workspace+"/snapshots", ""),
	}
}

// imageRevision ist die Revision des Workspace in der Liste b, die allein ihn führt.
func imageRevision(t *testing.T, b []byte, workspace string) int {
	t.Helper()
	var aus struct {
		Workspaces []struct {
			ID       string `json:"id"`
			Revision int    `json:"revision"`
		} `json:"workspaces"`
	}
	if err := json.Unmarshal(b, &aus); err != nil || len(aus.Workspaces) != 1 || aus.Workspaces[0].ID != workspace {
		t.Fatalf("Workspaces, erwartet allein %s: %v: %s", workspace, err, b)
	}
	return aus.Workspaces[0].Revision
}

// imageReferenz startet den Referenzdienst auf imageRef: $metadata und die Collection Products.
func imageReferenz(t *testing.T) {
	t.Helper()
	l, err := net.Listen("tcp", imageRef)
	if err != nil {
		t.Fatalf("Referenzdienst: %v", err)
	}
	srv := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("OData-Version", "4.0")
		switch r.URL.Path {
		case "/$metadata":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, fixture.CSDL)
		case "/Products":
			w.Header().Set("Content-Type", "application/json;odata.metadata=minimal")
			_, _ = io.WriteString(w, `{"@odata.context":"http://`+imageRef+`/$metadata#Products","value":[{"Id":1,"Name":"Anvil"},{"Id":2,"Name":"Bolt"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })
}

// TestImageAufnehmen ist die Phase aufnehmen (SPEC-323 Nr. 4 b): frisches Datenverzeichnis, Aufnahme gegen den Referenzdienst,
// Import, Workspace, Reset und Snapshot; die Zeile mit imageMarke im Log (mit `-test.v` auf stdout) trägt, was die Phase pruefen
// vergleicht.
func TestImageAufnehmen(t *testing.T) {
	c := imageClient{t: t, token: imageToken(t)}
	c.bereit()
	imageReferenz(t)

	vorher := c.sessions()
	var w imageWerte
	for id, status := range vorher {
		if status == "open" {
			w.Session = id
		}
	}
	if len(vorher) != 1 || w.Session == "" {
		t.Fatalf("kein frisches Datenverzeichnis: Sessions %v, erwartet genau eine open", vorher)
	}
	if b := c.rufe(200, "GET", "/__admin/workspaces", ""); string(b) != `{"workspaces":[]}` {
		t.Fatalf("kein frisches Datenverzeichnis: Workspaces %s", b)
	}

	for _, pfad := range []string{"/$metadata", "/Products"} {
		if a := holen(t, imageOData, anfrage{pfad: pfad}); a.status != 200 {
			t.Fatalf("Aufnahme %s: %d %s", pfad, a.status, a.body)
		}
	}
	if b := c.rufe(200, "POST", "/__admin/sessions/"+w.Session+"/close", ""); c.feld(b, "status") != "complete" {
		t.Fatalf("close: %s", b)
	}
	w.Import = "neustart"
	baseline := c.feld(c.rufe(201, "POST", "/__admin/imports", imageImport(w.Import, w.Session)), "baseline_snapshot_id")
	w.Workspace = c.feld(c.rufe(201, "POST", "/__admin/workspaces", `{"baseline_snapshot_id":"`+baseline+`"}`), "id")
	c.rufe(200, "POST", "/__admin/workspaces/"+w.Workspace+"/reset", "")
	w.Snapshot = c.feld(c.rufe(201, "POST", "/__admin/workspaces/"+w.Workspace+"/snapshots", `{"name":"neustart"}`), "id")

	w.Stand = c.stand(w)
	if r := imageRevision(t, w.Stand.Workspaces, w.Workspace); r != 1 {
		t.Fatalf("Revision %d nach reset, erwartet 1", r)
	}
	pruefeImageStand(t, w)
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(imageMarke + base64.StdEncoding.EncodeToString(b))
}

// imageImport ist der Body eines Imports der Session session unter der Kennung id.
func imageImport(id, session string) string {
	return `{"import_id":"` + id + `","service_id":"` + imageDienst + `","source_session_ids":["` + session + `"]}`
}

// pruefeImageStand verlangt in jeder Antwort ihren Gegenstand: eine leere Liste ist rot (SPEC-323 Nr. 7).
func pruefeImageStand(t *testing.T, w imageWerte) {
	t.Helper()
	for _, p := range []struct {
		name  string
		b     []byte
		teile []string
	}{
		{"Exchanges", w.Stand.Exchanges, []string{`"path":"/$metadata"`, `"path":"/Products"`, `"capture_status":"complete"`}},
		{"Import-Report", w.Stand.Report, []string{`"` + w.Import + `"`, `"` + w.Session + `"`}},
		{"Workspaces", w.Stand.Workspaces, []string{`"id":"` + w.Workspace + `"`}},
		{"Snapshots", w.Stand.Snapshots, []string{`"id":"` + w.Snapshot + `"`}},
	} {
		for _, teil := range p.teile {
			if !strings.Contains(string(p.b), teil) {
				t.Fatalf("%s ohne %s: %s", p.name, teil, p.b)
			}
		}
	}
}

// TestImagePruefen ist die Phase pruefen (SPEC-323 Nr. 4 d): nach Stopp und Start desselben Containers sind die vier Antworten
// byte-gleich, die Session ist complete neben genau einer neuen open, ein neuer Import liest die Bodies, und restore antwortet mit
// der nächsten Revision.
func TestImagePruefen(t *testing.T) {
	c := imageClient{t: t, token: imageToken(t)}
	roh, err := base64.StdEncoding.DecodeString(os.Getenv(imageVariable))
	if err != nil {
		t.Fatalf("%s: %v", imageVariable, err)
	}
	var w imageWerte
	if err := json.Unmarshal(roh, &w); err != nil || w.Session == "" || w.Import == "" || w.Workspace == "" || w.Snapshot == "" {
		t.Fatalf("%s ohne die Werte der Phase aufnehmen: %v", imageVariable, err)
	}
	pruefeImageStand(t, w)
	c.bereit()

	sessions := c.sessions()
	var offen []string
	for id, status := range sessions {
		if status == "open" {
			offen = append(offen, id)
		}
	}
	if len(sessions) != 2 || sessions[w.Session] != "complete" || len(offen) != 1 || slices.Contains(offen, w.Session) {
		t.Fatalf("Sessions nach dem Neustart: %v, erwartet %s complete und genau eine neue open", sessions, w.Session)
	}

	ist := c.stand(w)
	for _, p := range []struct {
		name      string
		ist, soll []byte
	}{
		{"Exchanges", ist.Exchanges, w.Stand.Exchanges},
		{"Import-Report", ist.Report, w.Stand.Report},
		{"Workspaces", ist.Workspaces, w.Stand.Workspaces},
		{"Snapshots", ist.Snapshots, w.Stand.Snapshots},
	} {
		if string(p.ist) != string(p.soll) {
			t.Errorf("%s nach dem Neustart:\n%s\nerwartet\n%s", p.name, p.ist, p.soll)
		}
	}
	if t.Failed() {
		return
	}

	c.rufe(201, "POST", "/__admin/imports", imageImport("neustart-2", w.Session))
	soll := imageRevision(t, w.Stand.Workspaces, w.Workspace) + 1
	b := c.rufe(200, "POST", "/__admin/workspaces/"+w.Workspace+"/restore", `{"snapshot_id":"`+w.Snapshot+`"}`)
	var nach struct {
		Revision int `json:"revision"`
	}
	if err := json.Unmarshal(b, &nach); err != nil || nach.Revision != soll {
		t.Errorf("restore: %s, erwartet Revision %d (%v)", b, soll, err)
	}
}
