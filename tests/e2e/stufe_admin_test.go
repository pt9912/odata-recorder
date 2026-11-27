package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestE2EAdmin ist die Stufe 11 (SPEC-037 Nr. 7): die Szenarien über den Admin-Listener am gebauten Binary, die Negative von
// LH-FA-14 über capture_status und in jedem Prozess die Suche nach dem Token in Strömen, Dateien und Antworten.
func TestE2EAdmin(t *testing.T) {
	binaryPfad(t)
	t.Run("Rundlauf", adminRundlauf)
	t.Run("leere Listen", adminLeer)
	t.Run("außerhalb von Loopback", adminOhneTLS)
	t.Run("nicht redigierbar", adminUnredigierbar)
	t.Run("Speicher voll", adminSpeicherVoll)
}

// adminMarke ist der Kontrollwert der Suche unter DATA_DIR (SPEC-037 Nr. 9).
const adminMarke = "KONTROLLE-ADMIN-DATEI-7d"

// adminProz ist ein Prozess des Binary mit Admin-Listener; antworten sammelt Header und Body jeder Antwort dieses Listeners.
type adminProz struct {
	t          *testing.T
	p          *prozess
	dir        string
	adresse    string
	token      string
	antworten  bytes.Buffer
	abgewiesen int
}

// adminStart startet k mit ADMIN_LISTEN auf einem freien Port, liest das Token aus DATA_DIR/admin.token und legt die Marke
// der Suche unter DATA_DIR ab.
func adminStart(t *testing.T, k konfig) *adminProz {
	t.Helper()
	a := &adminProz{t: t, dir: k.dir, adresse: freierPort(t)}
	k.extra = append(slices.Clone(k.extra), "ADMIN_LISTEN="+a.adresse)
	a.p = starte(t, k)
	a.p.bereit()
	b, err := os.ReadFile(filepath.Join(k.dir, "admin.token"))
	if err != nil {
		t.Fatalf("Token-Datei: %v\nstderr: %s", err, a.p.fehl)
	}
	a.token = strings.TrimSpace(string(b))
	if len(a.token) != 64 {
		t.Fatalf("Token mit %d Zeichen, erwartet 64", len(a.token))
	}
	if err := os.WriteFile(filepath.Join(k.dir, "kontrolle.txt"), []byte(adminMarke), 0o600); err != nil {
		t.Fatal(err)
	}
	return a
}

// sende stellt eine Anfrage an den Admin-Listener mit genau dem Header Authorization auth (leer: keiner).
func (a *adminProz) sende(methode, pfad, auth, body string) antwort {
	a.t.Helper()
	req, err := http.NewRequest(methode, "http://"+a.adresse+pfad, strings.NewReader(body))
	if err != nil {
		a.t.Fatal(err)
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	tr := &http.Transport{DisableKeepAlives: true}
	defer tr.CloseIdleConnections()
	resp, err := (&http.Client{Transport: tr, Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		a.t.Fatalf("%s %s: %v\nstderr: %s", methode, pfad, err, a.p.fehl)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		a.t.Fatalf("%s %s: Body: %v", methode, pfad, err)
	}
	_ = resp.Header.Write(&a.antworten)
	a.antworten.Write(b)
	return antwort{status: resp.StatusCode, header: resp.Header, body: b}
}

// rufe ruft mit dem Token und verlangt den Status.
func (a *adminProz) rufe(status int, methode, pfad, body string) antwort {
	a.t.Helper()
	w := a.sende(methode, pfad, "Bearer "+a.token, body)
	if w.status != status {
		a.t.Fatalf("%s %s: %d %s, erwartet %d\nstderr: %s", methode, pfad, w.status, w.body, status, a.p.fehl)
	}
	return w
}

func (a *adminProz) feld(w antwort, name string) string {
	a.t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.body, &m); err != nil {
		a.t.Fatalf("kein JSON: %s", w.body)
	}
	s, ok := m[name].(string)
	if !ok || s == "" {
		a.t.Fatalf("Antwort ohne %s: %s", name, w.body)
	}
	return s
}

// weise schickt jede Route ohne und mit falschem Token: 401 ODR-E2101 mit X-Correlation-Id, und der Zustand, den die
// Lesewege lesen, bleibt gleich.
func (a *adminProz) weise(lesen []string, routen ...[3]string) {
	a.t.Helper()
	stand := func() string {
		var s strings.Builder
		for _, pfad := range lesen {
			w := a.rufe(200, "GET", pfad, "")
			s.WriteString(pfad + " " + string(w.body) + "\n")
		}
		return s.String()
	}
	vorher := stand()
	for _, r := range routen {
		for _, auth := range []string{"", "Bearer " + strings.Repeat("0", 64)} {
			a.abgewiesen++
			w := a.sende(r[0], r[1], auth, r[2])
			if w.status != 401 || !bytes.Contains(w.body, []byte("ODR-E2101")) || w.header.Get("X-Correlation-Id") == "" {
				a.t.Errorf("%s %s mit %q: %d %s, Korrelations-ID %q", r[0], r[1], auth, w.status, w.body, w.header.Get("X-Correlation-Id"))
			}
		}
	}
	if nachher := stand(); nachher != vorher {
		a.t.Errorf("die Abweisungen änderten den Zustand:\nvorher %s\nnachher %s", vorher, nachher)
	}
}

// suche sucht das Token ganz und je Hälfte in stdout, stderr, jeder Datei unter DATA_DIR außer der Token-Datei und in den
// Antworten (SPEC-300 Nr. 12 a). Gültig ist sie nur mit den Kontrollwerten: der Marke unter DATA_DIR und der Zeile
// `listen=<LISTEN>` auf stderr.
func (a *adminProz) suche(wann string) {
	a.t.Helper()
	teile := []string{a.token, a.token[:32], a.token[32:]}
	_, funde := durchsuche(a.t, a.dir, append(slices.Clone(teile), adminMarke))
	marke := false
	for _, f := range funde {
		switch {
		case f.wert == adminMarke:
			marke = marke || f.ort == "kontrolle.txt"
		case f.ort != "admin.token":
			a.t.Errorf("%s: das Token steht in %s", wann, f.ort)
		}
	}
	if !marke {
		a.t.Errorf("%s: die Suche findet den Kontrollwert unter DATA_DIR nicht", wann)
	}
	if !hat(funde, "admin.token", a.token) {
		a.t.Errorf("%s: die Suche findet das Token in der Token-Datei nicht", wann)
	}
	for _, f := range sucheStroeme(a.p, teile) {
		a.t.Errorf("%s: das Token steht in %s", wann, f.ort)
	}
	if len(inBytes("stderr", a.p.fehl.bytes(), []string{"listen=" + a.p.adresse})) == 0 {
		a.t.Errorf("%s: die Suche findet den Kontrollwert auf stderr nicht: %s", wann, a.p.fehl)
	}
	if len(inBytes("Antworten", a.antworten.Bytes(), teile)) != 0 {
		a.t.Errorf("%s: das Token steht in einer Antwort", wann)
	}
	if a.antworten.Len() > 0 && !bytes.Contains(a.antworten.Bytes(), []byte("X-Correlation-Id")) {
		a.t.Errorf("%s: die Suche findet den Kontrollwert in den Antworten nicht", wann)
	}
}

// beende sucht vor und nach dem Herunterfahren; SIGTERM endet mit Exit 0.
func (a *adminProz) beende() {
	a.t.Helper()
	a.suche("vor dem Herunterfahren")
	if code := a.p.beende(); code != 0 {
		a.t.Fatalf("SIGTERM: Exit %d\nstderr: %s", code, a.p.fehl)
	}
	a.suche("nach dem Herunterfahren")
}

// exchange ist eine Zeile des Lesewegs Exchanges.
type exchange struct {
	ID      string `json:"id"`
	Methode string `json:"method"`
	Pfad    string `json:"path"`
	Capture string `json:"capture_status"`
}

func (a *adminProz) exchanges(session string) []exchange {
	a.t.Helper()
	var aus struct {
		Exchanges []exchange `json:"exchanges"`
	}
	w := a.rufe(200, "GET", "/__admin/sessions/"+session+"/exchanges", "")
	if err := json.Unmarshal(w.body, &aus); err != nil {
		a.t.Fatalf("Exchanges: %v: %s", err, w.body)
	}
	return aus.Exchanges
}

// session ist die eine Session, die der Start im Modus record angelegt hat.
func (a *adminProz) session() string {
	a.t.Helper()
	var aus struct {
		Sessions []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"sessions"`
	}
	w := a.rufe(200, "GET", "/__admin/sessions", "")
	if err := json.Unmarshal(w.body, &aus); err != nil || len(aus.Sessions) != 1 || aus.Sessions[0].Status != "open" {
		a.t.Fatalf("Sessions des Starts: %v: %s", err, w.body)
	}
	return aus.Sessions[0].ID
}

// adminNamen liest die Namen der Collection im Modus mock über den OData-Listener.
func adminNamen(t *testing.T, p *prozess, ws string) []string {
	t.Helper()
	w := holen(t, p.adresse, anfrage{pfad: collPfad + "?$orderby=Id", header: map[string][]string{wsKopf: {ws}}})
	var aus struct {
		Value []struct {
			Name string `json:"Name"`
		} `json:"value"`
	}
	if err := json.Unmarshal(w.body, &aus); err != nil || w.status != 200 {
		t.Fatalf("Collection: %d %s (%v)", w.status, w.body, err)
	}
	var n []string
	for _, v := range aus.Value {
		n = append(n, v.Name)
	}
	return n
}

// adminRundlauf ist der Happy Path über die Prozessgrenze (Aufnahme, close, Import, Workspace, Neustart im Modus mock,
// Snapshot, Reset, Restore), die Abweisungen jeder geschützten Route der Läufe, der Snapshot nach einem weiteren Neustart
// (Nr. 10 j) und audit.jsonl mit einer Zeile je Aufruf eines Schreibendpunkts.
func adminRundlauf(t *testing.T) {
	ref := neueReferenz(t)
	dir := t.TempDir()
	rec := adminStart(t, konfig{modus: "record", dir: dir, upstream: ref.url()})
	for _, pfad := range []string{collPfad, collPfad + "?$skiptoken=1"} {
		if w := holen(t, rec.p.adresse, anfrage{pfad: pfad}); w.status != 200 {
			t.Fatalf("Aufnahme %s: %d %s", pfad, w.status, w.body)
		}
	}
	if w := rec.sende("GET", "/healthz", "", ""); w.status != 200 || string(w.body) != `{"status":"ok"}` {
		t.Errorf("healthz ohne Token: %d %s", w.status, w.body)
	}
	s := rec.session()
	lesen := []string{"/__admin/sessions", "/__admin/sessions/" + s + "/exchanges", "/__admin/workspaces"}
	rec.weise(lesen, [3]string{"GET", "/__admin/sessions"}, [3]string{"GET", "/__admin/sessions/" + s + "/exchanges"},
		[3]string{"POST", "/__admin/sessions/" + s + "/close"})
	if w := rec.rufe(200, "POST", "/__admin/sessions/"+s+"/close", ""); rec.feld(w, "status") != "complete" {
		t.Fatalf("close: %s", w.body)
	}
	imp := `{"import_id":"e2e-admin","service_id":"ref","source_session_ids":["` + s + `"]}`
	rec.weise(lesen, [3]string{"POST", "/__admin/imports", imp})
	baseline := rec.feld(rec.rufe(201, "POST", "/__admin/imports", imp), "baseline_snapshot_id")
	neu := `{"baseline_snapshot_id":"` + baseline + `"}`
	rec.weise(lesen, [3]string{"POST", "/__admin/workspaces", neu})
	ws := rec.feld(rec.rufe(201, "POST", "/__admin/workspaces", neu), "id")
	// LH-FA-13 Happy Path: Sessions und Workspaces mit Kennung und Status.
	for pfad, teile := range map[string][]string{
		"/__admin/sessions":   {`"id":"` + s + `"`, `"status":"complete"`},
		"/__admin/workspaces": {`"id":"` + ws + `"`, `"baseline_snapshot_id":"` + baseline + `"`, `"revision":0`},
	} {
		w := rec.rufe(200, "GET", pfad, "")
		for _, teil := range teile {
			if !bytes.Contains(w.body, []byte(teil)) {
				t.Errorf("%s ohne %s: %s", pfad, teil, w.body)
			}
		}
	}
	rec.beende()
	ref.stoppe()

	mock := adminStart(t, konfig{modus: "mock", dir: dir})
	if mock.token != rec.token {
		t.Error("der zweite Start hat ein anderes Token")
	}
	pfadWS := "/__admin/workspaces/" + ws
	lesen = []string{"/__admin/sessions", "/__admin/workspaces", pfadWS + "/snapshots"}
	namen := func(schritt string, soll ...string) {
		t.Helper()
		if got := adminNamen(t, mock.p, ws); !slices.Equal(got, soll) {
			t.Errorf("%s: %v, erwartet %v", schritt, got, soll)
		}
	}
	aendere := func(methode, pfad, body string, status int) {
		t.Helper()
		a := anfrage{methode: methode, pfad: pfad, body: body, header: map[string][]string{wsKopf: {ws}, jsonKopf: {jsonTyp}}}
		if w := holen(t, mock.p.adresse, a); w.status != status {
			t.Fatalf("%s %s: %d %s", methode, pfad, w.status, w.body)
		}
	}
	namen("Baseline", "Anvil", "Bolt")
	aendere("POST", collPfad, `{"Id":3,"Name":"Cutter"}`, 201)
	snap := `{"name":"e2e"}`
	mock.weise(lesen, [3]string{"GET", "/__admin/sessions"}, [3]string{"GET", "/__admin/workspaces"},
		[3]string{"POST", pfadWS + "/snapshots", snap}, [3]string{"GET", pfadWS + "/snapshots"})
	if w := mock.sende("POST", "/__admin/imports", "", imp); w.status != 401 {
		t.Errorf("imports in mock ohne Token: %d %s", w.status, w.body)
	}
	mock.abgewiesen++
	if w := mock.rufe(404, "POST", "/__admin/imports", imp); !bytes.Contains(w.body, []byte("ODR-E2102")) {
		t.Errorf("imports in mock mit Token: %s, erwartet ODR-E2102", w.body)
	}
	snapshot := mock.feld(mock.rufe(201, "POST", pfadWS+"/snapshots", snap), "id")
	aendere("DELETE", collPfad+"(1)", "", 204)
	namen("nach den Änderungen", "Bolt", "Cutter")
	wieder := `{"snapshot_id":"` + snapshot + `"}`
	mock.weise(lesen, [3]string{"POST", pfadWS + "/reset"}, [3]string{"POST", pfadWS + "/restore", wieder})
	namen("nach den Abweisungen", "Bolt", "Cutter")
	mock.rufe(200, "POST", pfadWS+"/reset", "")
	namen("nach reset", "Anvil", "Bolt")
	mock.rufe(200, "POST", pfadWS+"/restore", wieder)
	namen("nach restore", "Anvil", "Bolt", "Cutter")
	mock.beende()

	neustart := adminStart(t, konfig{modus: "mock", dir: dir})
	if w := neustart.rufe(200, "GET", pfadWS+"/snapshots", ""); !bytes.Contains(w.body, []byte(`"id":"`+snapshot+`"`)) || !bytes.Contains(w.body, []byte(`"name":"e2e"`)) {
		t.Errorf("Snapshots nach dem Neustart: %s", w.body)
	}
	neustart.beende()

	pruefeAdminAudit(t, dir, rec.abgewiesen+mock.abgewiesen)
	pruefeStdout(t, rec.token)
}

// pruefeAdminAudit liest audit.jsonl: neben je einer Zeile je Abweisung genau eine Zeile je Aufruf eines Schreibendpunkts
// in der Folge der Schritte.
func pruefeAdminAudit(t *testing.T, dir string, abgewiesen int) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var arten []string
	n := 0
	for _, z := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		var m map[string]string
		if err := json.Unmarshal([]byte(z), &m); err != nil {
			t.Fatalf("Zeile %q: %v", z, err)
		}
		if m["art"] == "admin_abgewiesen" {
			n++
			continue
		}
		arten = append(arten, m["art"])
	}
	soll := []string{"session_beendet", "import_veroeffentlicht", "workspace_angelegt", "snapshot_erzeugt",
		"workspace_zurueckgesetzt", "workspace_wiederhergestellt"}
	if !slices.Equal(arten, soll) || n != abgewiesen || abgewiesen == 0 {
		t.Errorf("audit.jsonl: %v und %d Zeilen admin_abgewiesen, erwartet %v und %d", arten, n, soll, abgewiesen)
	}
}

// pruefeStdout belegt den Weg, auf dem die Suche stdout liest, mit dem Kommando version, das dorthin schreibt; stdout von
// `start` trägt keinen Wert, den der Test einspielen könnte.
func pruefeStdout(t *testing.T, token string) {
	t.Helper()
	aus := &puffer{}
	cmd := exec.Command(binaryPfad(t), "version")
	cmd.Env, cmd.Stdout = []string{}, aus
	if err := cmd.Run(); err != nil {
		t.Fatalf("version: %v", err)
	}
	if v := strings.TrimSpace(aus.String()); v == "" || len(inBytes("stdout", aus.bytes(), []string{v})) == 0 || len(inBytes("stdout", aus.bytes(), []string{token})) != 0 {
		t.Errorf("stdout von version: %q", aus)
	}
}

// adminLeer: auf einem frischen DATA_DIR im Modus mock sind beide Listen leer (LH-FA-13 Boundary).
func adminLeer(t *testing.T) {
	a := adminStart(t, konfig{modus: "mock", dir: t.TempDir()})
	for pfad, soll := range map[string]string{"/__admin/sessions": `{"sessions":[]}`, "/__admin/workspaces": `{"workspaces":[]}`} {
		if w := a.rufe(200, "GET", pfad, ""); string(w.body) != soll {
			t.Errorf("%s: %s, erwartet %s", pfad, w.body, soll)
		}
	}
	a.beende()
}

// adminOhneTLS: ADMIN_LISTEN außerhalb von Loopback ohne Paar und ohne Flag ist Exit 1 ohne Listener.
func adminOhneTLS(t *testing.T) {
	adresse, adminAdresse := freierPort(t), freierPort(t)
	port := adminAdresse[strings.LastIndex(adminAdresse, ":"):]
	env := append(konfig{modus: "mock", dir: t.TempDir()}.env(adresse), "ADMIN_LISTEN=0.0.0.0"+port)
	p := startRoh(t, adresse, env)
	verbunden := false
	for stop := time.Now().Add(bereitFrist); !p.beendet() && time.Now().Before(stop); {
		for _, ziel := range []string{adresse, adminAdresse} {
			if c, err := net.DialTimeout("tcp", ziel, 20*time.Millisecond); err == nil {
				verbunden = true
				_ = c.Close()
			}
		}
	}
	p.warteEnde()
	if p.exit != 1 || verbunden || !strings.Contains(p.fehl.String(), "ODR-E2001 ADMIN_LISTEN:") {
		t.Errorf("Exit %d, Listener geöffnet: %v\nstderr: %s", p.exit, verbunden, p.fehl)
	}
}

// adminUnredigierbar: der Austausch mit nicht redigierbarem Body steht in der Liste als failed und nie als complete
// (LH-FA-14 Negative, SPEC-037 Nr. 10 h); der Austausch von $metadata ist der Kontrollwert.
func adminUnredigierbar(t *testing.T) {
	ref := neueReferenz(t)
	a := adminStart(t, konfig{modus: "record", dir: t.TempDir(), upstream: ref.url()})
	s := a.session()
	if w := holen(t, a.p.adresse, postJSON(`{"Name": "MARKER-ADMIN-UNREDIGIERBAR`)); w.status != 201 {
		t.Errorf("Status des Clients %d, erwartet 201", w.status)
	}
	a.p.warteAufZeile(`"error_code":"ODR-E2401"`, `"richtung":"request"`)
	a.rufe(200, "POST", "/__admin/sessions/"+s+"/close", "")
	var ist []string
	for _, e := range a.exchanges(s) {
		ist = append(ist, e.Methode+" "+e.Pfad+" "+e.Capture)
	}
	if soll := []string{"GET /$metadata complete", "POST /Products failed"}; !slices.Equal(ist, soll) {
		t.Errorf("Exchanges %v, erwartet %v", ist, soll)
	}
	a.beende()
}

// adminSpeicherVoll: nach dem Entfernen der Ballast-Datei sind die Exchanges mit capture_status complete genau die vor dem
// Füllen entstandenen, und close meldet incomplete (SPEC-037 Nr. 7; ob der Austausch bei vollem Speicher fehlt oder markiert
// in der Liste steht, fordert die Stufe nicht, Nr. 10 o).
func adminSpeicherVoll(t *testing.T) {
	ref := neueReferenz(t)
	dir, err := os.MkdirTemp("/voll", "data")
	if err != nil {
		t.Fatalf("DATA_DIR unter /voll: %v", err)
	}
	a := adminStart(t, konfig{modus: "record", dir: dir, upstream: ref.url()})
	s := a.session()
	if w := holen(t, a.p.adresse, anfrage{pfad: "/odata/Products(1)"}); w.status != 200 {
		t.Fatalf("vor dem Füllen: %d %s", w.status, w.body)
	}
	vollstaendig := func() []string {
		var ids []string
		for _, e := range a.exchanges(s) {
			if e.Capture == "complete" {
				ids = append(ids, e.ID)
			}
		}
		return ids
	}
	// Kontrollwert: der Anhang folgt der Antwort, und vor dem Füllen stehen $metadata und die Entität mit complete in der
	// Liste.
	vorher := vollstaendig()
	for ende := time.Now().Add(bereitFrist); len(vorher) != 2; vorher = vollstaendig() {
		if time.Now().After(ende) {
			t.Fatalf("vor dem Füllen stehen %d Exchanges mit complete in der Liste, erwartet 2: %v", len(vorher), a.exchanges(s))
		}
		time.Sleep(25 * time.Millisecond)
	}
	ballast := fuelle(t)
	arten := []string{"/odata/Products", "/odata/Products/$count"}
	for _, pfad := range arten {
		if w := holen(t, a.p.adresse, anfrage{pfad: pfad}); w.status != 200 {
			t.Errorf("%s bei vollem Speicher: %d %s", pfad, w.status, w.body)
		}
	}
	a.p.warteAufZeilen(len(arten), `"ereignis":"ablage"`)
	if err := os.Remove(ballast); err != nil {
		t.Fatal(err)
	}
	if w := a.rufe(200, "POST", "/__admin/sessions/"+s+"/close", ""); a.feld(w, "status") != "incomplete" {
		t.Errorf("close nach vollem Speicher: %s, erwartet incomplete", w.body)
	}
	if nachher := vollstaendig(); !slices.Equal(nachher, vorher) {
		t.Errorf("Exchanges mit complete %v, erwartet die vor dem Füllen entstandenen %v", nachher, vorher)
	}
	a.beende()
}
