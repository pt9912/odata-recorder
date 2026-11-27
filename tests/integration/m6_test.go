package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// m6Lauf ist der Rundlauf von SPEC-300 Nr. 14: zwei Starts auf demselben DATA_DIR und zwei weitere zum Lesen.
type m6Lauf struct {
	t   *testing.T
	ref *referenzdienst
	dir string
	// a und b sind die Sessions, baseline die des Imports, ws der Workspace.
	a, b, baseline, ws string
	// korrelation sind die Korrelations-IDs der Antworten in der Folge der Zeilen, die der Admin-Adapter selbst schreibt;
	// leer steht für eine Zeile eines Dienstes.
	korrelation []string
	zellen      int
}

func (w *m6Lauf) merke(a antwort) {
	w.korrelation = append(w.korrelation, a.header.Get("X-Correlation-Id"))
}
func (w *m6Lauf) merkeDienstZeile() { w.korrelation = append(w.korrelation, "") }

// TestM6Rundlauf (D01) fährt den Rundlauf ausschließlich über Admin- und OData-Listener (LH-FA-13, LH-FA-18, LH-FA-09,
// LH-FA-10, LH-FA-02): Orakel sind die geparste audit.jsonl, die Lesewege und die Bytes der Token-Datei.
func TestM6Rundlauf(t *testing.T) {
	w := &m6Lauf{t: t, ref: neuerReferenzdienst(t), dir: t.TempDir()}
	rec := m6Start(t, "record", w.dir, w.ref.URL()+"/odata")
	w.abweisungen(rec)
	w.aufnahme(rec)
	w.importiere(rec)
	rec.beende()

	mock := m6Start(t, "mock", w.dir, "")
	if !bytes.Equal(mock.tokenDatei, rec.tokenDatei) {
		t.Error("der zweite Start hat die Token-Datei verändert")
	}
	w.aendere(mock)
	w.lesewege(mock, "mock")
	mock.rufe(200, "DELETE", "/__admin/workspaces/"+w.ws, "")
	w.merkeDienstZeile()
	mock.beende()

	for _, modus := range []string{"replay", "hybrid"} {
		h := m6Start(t, modus, w.dir, "")
		if !bytes.Equal(h.tokenDatei, rec.tokenDatei) {
			t.Errorf("%s: der Start hat die Token-Datei verändert", modus)
		}
		w.lesewege(h, modus)
		h.beende()
	}
	if w.zellen != 6 {
		t.Errorf("%d Zellen der Lesewege, erwartet 6", w.zellen)
	}
	w.pruefeAudit(rec.token)
}

// abweisungen ist Schritt (1): healthz ohne Token, die Sessions ohne und mit falschem Token.
func (w *m6Lauf) abweisungen(rec *m6Handle) {
	t := w.t
	if a := rec.m6Sende("GET", "/healthz", "", ""); a.status != 200 || string(a.body) != `{"status":"ok"}` {
		t.Errorf("healthz: %d %s", a.status, a.body)
	}
	for _, auth := range []string{"", "Bearer " + strings.Repeat("0", 64)} {
		a := rec.m6Sende("GET", "/__admin/sessions", auth, "")
		if a.status != 401 || odrCode(a.body) != "ODR-E2101" {
			t.Errorf("ohne gültiges Token (%q): %d %s", auth, a.status, a.body)
		}
		w.merke(a)
	}
}

// aufnahme ist Schritt (2): Sessions A und B, in A $metadata, die zwei Seiten der Collection und ein POST mit 201, beide
// beenden, B löschen.
func (w *m6Lauf) aufnahme(rec *m6Handle) {
	t := w.t
	for _, ziel := range []*string{&w.a, &w.b} {
		a := rec.rufe(201, "POST", "/__admin/sessions", `{"service_id":"`+serviceID+`"}`)
		w.merke(a)
		*ziel = m6Feld(t, a, "id")
	}
	for _, f := range []struct {
		methode, pfad, accept, body string
		status                      int
	}{
		{"GET", "/odata/$metadata", acceptXML, "", 200},
		{"GET", "/odata/Products", acceptJSON, "", 200},
		{"GET", "/odata/Products?$skiptoken=1", acceptJSON, "", 200},
		{"POST", "/odata/Products", acceptJSON, `{"Id":3,"Name":"Cutter"}`, 201},
	} {
		kopfzeilen := map[string]string{"Accept": f.accept, "X-Recorder-Recording-Session": w.a}
		if a := rec.odata.sende(f.methode, f.pfad, f.body, "", kopfzeilen); a.status != f.status {
			t.Fatalf("Aufnahme %s %s: %d %s", f.methode, f.pfad, a.status, a.body)
		}
	}
	for _, id := range []string{w.a, w.b} {
		a := rec.rufe(200, "POST", "/__admin/sessions/"+id+"/close", "")
		w.merke(a)
		if got := m6Feld(t, a, "status"); got != "complete" {
			t.Errorf("close %s: Status %s, erwartet complete", id, got)
		}
	}
	rec.rufe(200, "DELETE", "/__admin/sessions/"+w.b, "")
	w.merkeDienstZeile()
}

// importiere ist Schritt (3): Import von A, Workspace aus der Baseline.
func (w *m6Lauf) importiere(rec *m6Handle) {
	a := rec.rufe(201, "POST", "/__admin/imports",
		`{"import_id":"m6-import","service_id":"`+serviceID+`","source_session_ids":["`+w.a+`"],"metadata_hash":"`+csdlHash()+`"}`)
	w.merke(a)
	w.baseline = m6Feld(w.t, a, "baseline_snapshot_id")
	a = rec.rufe(201, "POST", "/__admin/workspaces", `{"baseline_snapshot_id":"`+w.baseline+`"}`)
	w.merke(a)
	w.ws = m6Feld(w.t, a, "id")
}

// aendere ist Schritt (4): ändern, Snapshot, weiter ändern, zurücksetzen (Baseline), wiederherstellen (Snapshot).
func (w *m6Lauf) aendere(mock *m6Handle) {
	t := w.t
	z := map[string]string{"X-Recorder-Workspace": w.ws}
	liste := func(schritt string, soll ...string) {
		t.Helper()
		a := mock.odata.sende("GET", m3Pfad+"?$orderby=Id", "", "", z)
		if got := namen(t, a.body); a.status != 200 || !slices.Equal(got, soll) {
			t.Errorf("%s: %d %v, erwartet %v", schritt, a.status, got, soll)
		}
	}
	liste("Baseline", "Anvil", "Bolt")
	if a := mock.odata.sende("POST", m3Pfad, `{"Id":3,"Name":"Cutter"}`, "", z); a.status != 201 {
		t.Fatalf("Änderung: %d %s", a.status, a.body)
	}
	liste("nach der Änderung", "Anvil", "Bolt", "Cutter")
	a := mock.rufe(201, "POST", "/__admin/workspaces/"+w.ws+"/snapshots", `{"name":"m6"}`)
	w.merke(a)
	snapshot := m6Feld(t, a, "id")
	if a := mock.odata.sende("DELETE", m3Pfad+"(1)", "", "", z); a.status != 204 {
		t.Fatalf("weitere Änderung: %d %s", a.status, a.body)
	}
	liste("nach der weiteren Änderung", "Bolt", "Cutter")
	w.merke(mock.rufe(200, "POST", "/__admin/workspaces/"+w.ws+"/reset", ""))
	liste("nach reset", "Anvil", "Bolt")
	w.merke(mock.rufe(200, "POST", "/__admin/workspaces/"+w.ws+"/restore", `{"snapshot_id":"`+snapshot+`"}`))
	liste("nach restore", "Anvil", "Bolt", "Cutter")
}

// lesewege führt die zwei Zellen {Exchanges, Coverage} eines Modus am Listener: Orakel sind Form und Daten der eigenen Route.
func (w *m6Lauf) lesewege(m *m6Handle, modus string) {
	t := w.t
	t.Helper()
	w.zellen++
	e := jsonVon(t, m.rufe(200, "GET", "/__admin/sessions/"+w.a+"/exchanges", "").body)
	var ist []string
	zeilen, _ := e["exchanges"].([]any)
	for _, z := range zeilen {
		x := z.(map[string]any)
		ist = append(ist, fmt.Sprintf("%v %v %v %v %v", x["sequence_no"], x["method"], x["path"], x["response_status"], x["capture_status"]))
	}
	soll := []string{"1 GET /$metadata 200 complete", "2 GET /Products 200 complete", "3 GET /Products 200 complete",
		"4 POST /Products 201 complete"}
	if !slices.Equal(m6Schluessel(e), []string{"exchanges", "limit", "next", "total"}) || !slices.Equal(ist, soll) || e["total"] != float64(4) {
		t.Errorf("%s: Exchanges %v (%v), erwartet %v", modus, ist, m6Schluessel(e), soll)
	}
	w.zellen++
	c := jsonVon(t, m.rufe(200, "GET", "/__admin/coverage?baseline_id="+w.baseline, "").body)
	eintraege, _ := c["entries"].([]any)
	if !slices.Equal(m6Schluessel(c), []string{"entries"}) || len(eintraege) != 1 {
		t.Fatalf("%s: Coverage %v", modus, c)
	}
	if x := eintraege[0].(map[string]any); x["entity_set"] != "Products" || x["status"] != "complete" || x["evidence"] == nil {
		t.Errorf("%s: Coverage-Eintrag %v", modus, x)
	}
}

// m6Schluessel sind die Mitglieder eines Objekts, sortiert.
func m6Schluessel(m map[string]any) []string {
	var k []string
	for n := range m {
		k = append(k, n)
	}
	slices.Sort(k)
	return k
}

// pruefeAudit liest die Datei am Ende: genau die zwei Zeilen der Abweisungen und die elf Zeilen der neun Arten in der Folge
// der Schritte, die Zeilen des Adapters mit der Korrelations-ID ihrer Antwort, und kein Byte des Tokens (SPEC-300 Nr. 12).
func (w *m6Lauf) pruefeAudit(token string) {
	t := w.t
	roh, zeilen := m6Audit(t, w.dir)
	for _, teil := range []string{token, token[:len(token)/2], token[len(token)/2:]} {
		if len(teil) < 32 || bytes.Contains(roh, []byte(teil)) {
			t.Errorf("audit.jsonl trägt das Token oder eine Hälfte (Länge %d)", len(teil))
		}
	}
	s := " service_id=" + serviceID
	soll := []string{
		"art=admin_abgewiesen error_code=ODR-E2101",
		"art=admin_abgewiesen error_code=ODR-E2101",
		"art=session_angelegt" + s + " session_id=" + w.a,
		"art=session_angelegt" + s + " session_id=" + w.b,
		"art=session_beendet" + s + " session_id=" + w.a,
		"art=session_beendet" + s + " session_id=" + w.b,
		"art=session_geloescht" + s + " session_id=" + w.b,
		"art=import_veroeffentlicht import_id=m6-import" + s,
		"art=workspace_angelegt" + s + " workspace_id=" + w.ws,
		"art=snapshot_erzeugt" + s + " workspace_id=" + w.ws,
		"art=workspace_zurueckgesetzt" + s + " workspace_id=" + w.ws,
		"art=workspace_wiederhergestellt" + s + " workspace_id=" + w.ws,
		"art=workspace_geloescht workspace_id=" + w.ws,
	}
	var ist []string
	arten := map[string]int{}
	for _, z := range zeilen {
		ist = append(ist, m6Felder(z))
		arten[z["art"]]++
	}
	if !slices.Equal(ist, soll) {
		t.Fatalf("audit.jsonl:\n%s\nerwartet:\n%s", strings.Join(ist, "\n"), strings.Join(soll, "\n"))
	}
	if len(arten) != 10 || arten["admin_abgewiesen"] != 2 || len(zeilen)-2 != 11 {
		t.Errorf("%d Zeilen der Arten %v, erwartet 11 Zeilen der 9 Arten und 2 admin_abgewiesen", len(zeilen), arten)
	}
	if len(w.korrelation) != len(zeilen) {
		t.Fatalf("%d gemerkte Aufrufe, %d Zeilen", len(w.korrelation), len(zeilen))
	}
	for i, z := range zeilen {
		if k := w.korrelation[i]; k != "" && z["correlation_id"] != k {
			t.Errorf("Zeile %d: correlation_id %q, Antwort %q", i+1, z["correlation_id"], k)
		}
	}
}

// m6Vorlage ist ein DATA_DIR, in dem jede schreibende Route mit gültigem Token etwas ändert: eine aufgenommene und eine leere
// beendete Session, eine Baseline, ein Workspace mit Snapshot und einer Änderung danach.
type m6Vorlage struct {
	dir, aufnahme, leer, baseline, ws, snapshot string
}

func m6Aufbau(t *testing.T) m6Vorlage {
	t.Helper()
	v := m6Vorlage{dir: t.TempDir()}
	ref := neuerReferenzdienst(t)
	rec := m6Start(t, "record", v.dir, ref.URL()+"/odata")
	v.aufnahme = m6Feld(t, rec.rufe(201, "POST", "/__admin/sessions", `{"service_id":"`+serviceID+`"}`), "id")
	for _, pfad := range []string{"/odata/$metadata", "/odata/Products", "/odata/Products?$skiptoken=1"} {
		accept := acceptJSON
		if pfad == "/odata/$metadata" {
			accept = acceptXML
		}
		if a := rec.odata.sende("GET", pfad, "", "", map[string]string{"Accept": accept, "X-Recorder-Recording-Session": v.aufnahme}); a.status != 200 {
			t.Fatalf("Aufnahme %s: %d %s", pfad, a.status, a.body)
		}
	}
	rec.rufe(200, "POST", "/__admin/sessions/"+v.aufnahme+"/close", "")
	v.leer = m6Feld(t, rec.rufe(201, "POST", "/__admin/sessions", `{"service_id":"`+serviceID+`"}`), "id")
	rec.rufe(200, "POST", "/__admin/sessions/"+v.leer+"/close", "")
	v.baseline = m6Feld(t, rec.rufe(201, "POST", "/__admin/imports",
		`{"import_id":"m6-vorlage","service_id":"`+serviceID+`","source_session_ids":["`+v.aufnahme+`"]}`), "baseline_snapshot_id")
	rec.beende()
	mock := m6Start(t, "mock", v.dir, "")
	v.ws = m6Feld(t, mock.rufe(201, "POST", "/__admin/workspaces", `{"baseline_snapshot_id":"`+v.baseline+`"}`), "id")
	z := map[string]string{"X-Recorder-Workspace": v.ws}
	if a := mock.odata.sende("POST", m3Pfad, `{"Id":3,"Name":"Cutter"}`, "", z); a.status != 201 {
		t.Fatalf("Änderung: %d %s", a.status, a.body)
	}
	v.snapshot = m6Feld(t, mock.rufe(201, "POST", "/__admin/workspaces/"+v.ws+"/snapshots", `{"name":"vorlage"}`), "id")
	if a := mock.odata.sende("DELETE", m3Pfad+"(1)", "", "", z); a.status != 204 {
		t.Fatalf("Änderung nach dem Snapshot: %d %s", a.status, a.body)
	}
	mock.beende()
	return v
}

// kopie ist eine Kopie der Vorlage; die Token-Datei behält die Rechte 0600.
func (v m6Vorlage) kopie(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(v.dir)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(dir, "admin.token"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

// m6Route ist eine schreibende Route mit einem Aufruf, der mit gültigem Token die Ablage ändert.
type m6Route struct {
	name, methode, pfad, body string
	status                    int
}

// schreibende sind die Routen von SPEC-005, die schreiben, im Modus modus: 9 in record, 8 sonst; die Folge ist die, in der
// jede mit gültigem Token gelingt.
func (v m6Vorlage) schreibende(modus, offen string) []m6Route {
	ws := "/__admin/workspaces/" + v.ws
	r := []m6Route{
		{"POST sessions", "POST", "/__admin/sessions", `{"service_id":"` + serviceID + `"}`, 201},
		{"POST close", "POST", "/__admin/sessions/" + offen + "/close", "", 200},
		{"DELETE session", "DELETE", "/__admin/sessions/" + v.leer, "", 200},
	}
	if modus == "record" {
		r = append(r, m6Route{"POST imports", "POST", "/__admin/imports",
			`{"import_id":"m6-neu","service_id":"` + serviceID + `","source_session_ids":["` + v.aufnahme + `"]}`, 201})
	}
	return append(r,
		m6Route{"POST workspaces", "POST", "/__admin/workspaces", `{"baseline_snapshot_id":"` + v.baseline + `"}`, 201},
		m6Route{"POST snapshots", "POST", ws + "/snapshots", `{"name":"m6-neu"}`, 201},
		m6Route{"POST reset", "POST", ws + "/reset", "", 200},
		m6Route{"POST restore", "POST", ws + "/restore", `{"snapshot_id":"` + v.snapshot + `"}`, 200},
		m6Route{"DELETE workspace", "DELETE", ws, "", 200},
	)
}

// TestM6AbweisungOhneAenderung (D02, LH-FA-13 Negative): ohne und mit falschem Token ist jede schreibende Route 401
// ODR-E2101, die Ablage bleibt Zeile für Zeile und Datei für Datei gleich, und audit.jsonl wächst um genau die Zeile der
// Abweisung; 18 Zellen in record, je 16 in replay, mock und hybrid. Danach ändert jede der Routen mit dem Token die Ablage:
// das ist der Kontrollwert des Vergleichs.
func TestM6AbweisungOhneAenderung(t *testing.T) {
	v := m6Aufbau(t)
	zellen := 0
	for _, modus := range []string{"record", "replay", "mock", "hybrid"} {
		dir := v.kopie(t)
		upstream := ""
		if modus == "record" {
			upstream = neuerReferenzdienst(t).URL() + "/odata"
		}
		m := m6Start(t, modus, dir, upstream)
		offen := m6Feld(t, m.rufe(201, "POST", "/__admin/sessions", `{"service_id":"`+serviceID+`"}`), "id")
		routen := v.schreibende(modus, offen)
		if n := map[string]int{"record": 9, "replay": 8, "mock": 8, "hybrid": 8}[modus]; len(routen) != n {
			t.Fatalf("%s: %d Routen, erwartet %d", modus, len(routen), n)
		}
		for _, r := range routen {
			for _, anlass := range []struct{ name, auth string }{{"ohne Token", ""}, {"falsches Token", "Bearer " + strings.Repeat("0", 64)}} {
				zellen++
				name := modus + " " + r.name + " " + anlass.name
				vorher := m6Zustand(t, dir)
				rohVorher, zeilenVorher := m6Audit(t, dir)
				a := m.m6Sende(r.methode, r.pfad, anlass.auth, r.body)
				if a.status != 401 || odrCode(a.body) != "ODR-E2101" || a.header.Get("X-Correlation-Id") == "" {
					t.Errorf("%s: %d %s", name, a.status, a.body)
				}
				if d := vorher.unterschiede(m6Zustand(t, dir)); len(d) != 0 {
					t.Errorf("%s: die Ablage änderte sich: %v", name, d)
				}
				roh, zeilen := m6Audit(t, dir)
				if !bytes.HasPrefix(roh, rohVorher) || len(zeilen) != len(zeilenVorher)+1 ||
					m6Felder(zeilen[len(zeilen)-1]) != "art=admin_abgewiesen error_code=ODR-E2101" ||
					zeilen[len(zeilen)-1]["correlation_id"] != a.header.Get("X-Correlation-Id") {
					t.Errorf("%s: audit.jsonl von %d auf %d Zeilen: %v", name, len(zeilenVorher), len(zeilen), zeilen[len(zeilenVorher):])
				}
			}
		}
		for _, r := range routen {
			vorher := m6Zustand(t, dir)
			a := m.m6Sende(r.methode, r.pfad, "Bearer "+m.token, r.body)
			if a.status != r.status || len(vorher.unterschiede(m6Zustand(t, dir))) == 0 {
				t.Errorf("%s: %s mit gültigem Token: %d %s; erwartet %d und eine Änderung, sonst belegt der Vergleich für die Route nichts",
					modus, r.name, a.status, a.body, r.status)
			}
		}
		m.beende()
	}
	if zellen != 66 {
		t.Errorf("%d Zellen, erwartet 66", zellen)
	}
}

// TestM6Cursor (D07, LH-FA-10 Happy Path, LH-FA-09 Boundary): am Admin-Listener lässt restore den Replay-Cursor stehen und
// reset setzt ihn zurück. Orakel sind Status und Body der Replay-Antworten.
func TestM6Cursor(t *testing.T) {
	ref := neuerReferenzdienst(t)
	dir := t.TempDir()
	rec := m6Start(t, "record", dir, ref.URL()+"/odata")
	session := m6Feld(t, rec.rufe(201, "POST", "/__admin/sessions", `{"service_id":"`+serviceID+`"}`), "id")
	inS := map[string]string{"X-Recorder-Recording-Session": session}
	if a := rec.odata.hole(anfrage{"/odata/$metadata", acceptXML}, inS); a.status != 200 {
		t.Fatalf("$metadata: %d", a.status)
	}
	zaehler := anfrage{"/odata/Products/$count", acceptText}
	for _, soll := range []string{"7", "11", "13"} {
		if a := rec.odata.hole(zaehler, inS); a.status != 200 || string(a.body) != soll {
			t.Fatalf("Aufnahme von $count: %d %q, erwartet %s", a.status, a.body, soll)
		}
	}
	rec.rufe(200, "POST", "/__admin/sessions/"+session+"/close", "")
	baseline := m6Feld(t, rec.rufe(201, "POST", "/__admin/imports",
		`{"import_id":"m6-cursor","service_id":"`+serviceID+`","source_session_ids":["`+session+`"]}`), "baseline_snapshot_id")
	rec.beende()

	rep := m6Start(t, "replay", dir, "")
	ws := m6Feld(t, rep.rufe(201, "POST", "/__admin/workspaces", `{"baseline_snapshot_id":"`+baseline+`"}`), "id")
	zuordnung := map[string]string{"X-Recorder-Workspace": ws, "X-Recorder-Recording-Session": session, "X-Recorder-Replay-Session": "m6-cursor"}
	antworte := func(schritt, soll string) {
		t.Helper()
		if a := rep.odata.hole(zaehler, zuordnung); a.status != 200 || string(a.body) != soll {
			t.Fatalf("%s: %d %q, erwartet 200 %q", schritt, a.status, a.body, soll)
		}
	}
	antworte("erste Antwort", "7")
	snapshot := m6Feld(t, rep.rufe(201, "POST", "/__admin/workspaces/"+ws+"/snapshots", `{"name":"cursor"}`), "id")
	antworte("zweite Antwort", "11")
	rep.rufe(200, "POST", "/__admin/workspaces/"+ws+"/restore", `{"snapshot_id":"`+snapshot+`"}`)
	antworte("nach restore die dritte Antwort", "13")
	rep.rufe(200, "POST", "/__admin/workspaces/"+ws+"/reset", "")
	antworte("nach reset wieder die erste Antwort", "7")
	rep.beende()
}
