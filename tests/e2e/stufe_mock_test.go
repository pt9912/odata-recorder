package e2e_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	wsKopf   = "X-Recorder-Workspace"
	jsonTyp  = "application/json"
	collPfad = "/odata/Products"
)

// mockWelt ist der Zustand vor dem Modus mock: aufgenommen (`$metadata` und beide Seiten von Products),
// importiert, Workspace angelegt, Referenzdienst gestoppt (SPEC-037 Nr. 7).
type mockWelt struct {
	l   *lauf
	ref *referenz
	// vor sind die Zähler des Referenzdienstes nach dem Stoppen; kein Schritt ändert sie.
	vor [2]int64
}

func neueMockWelt(t *testing.T) *mockWelt {
	t.Helper()
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	l.frage(anfrage{pfad: collPfad})
	l.frage(anfrage{pfad: collPfad + "?$skiptoken=1"})
	l.beendeAufnahme()
	ref.stoppe()
	l.importiere()
	return &mockWelt{l: l, ref: ref, vor: ref.zaehler()}
}

// konfig ist die Umgebung des Modus mock auf dir: ohne UPSTREAM_URL und ohne PUBLIC_BASE_URL.
func (w *mockWelt) konfig(dir string) konfig {
	return konfig{modus: "mock", dir: dir, dienst: w.l.dienst}
}

// startAuf startet das Binary im Modus mock auf dir und adresse und wartet auf die Verbindung ohne Request.
func (w *mockWelt) startAuf(t *testing.T, dir, adresse string) *prozess {
	t.Helper()
	env := w.konfig(dir).env(adresse)
	for _, e := range env {
		if strings.HasPrefix(e, "UPSTREAM_URL=") || strings.HasPrefix(e, "PUBLIC_BASE_URL=") {
			t.Fatalf("die Umgebung des Modus mock trägt %s", e)
		}
	}
	p := startRoh(t, adresse, env)
	p.modus = "mock"
	p.bereit()
	return p
}

// start startet den Modus mock auf dem DATA_DIR des Laufs an einem freien Port.
func (w *mockWelt) start(t *testing.T) *prozess {
	t.Helper()
	return w.startAuf(t, w.l.dir, freierPort(t))
}

// neustart beendet p mit SIGTERM (Exit 0) und startet das Binary auf demselben DATA_DIR und derselben
// Adresse: der Wert von PUBLIC_BASE_URL folgt aus LISTEN und bleibt gleich.
func (w *mockWelt) neustart(t *testing.T, p *prozess) *prozess {
	t.Helper()
	if code := p.beende(); code != 0 {
		t.Fatalf("SIGTERM: Exit %d, erwartet 0\nstderr: %s", code, p.fehl)
	}
	p.verweigert()
	return w.startAuf(t, w.l.dir, p.adresse)
}

// hole stellt a mit dem Workspace des Laufs an p.
func (w *mockWelt) hole(t *testing.T, p *prozess, a anfrage) antwort {
	t.Helper()
	a.header = zusammen(a.header, map[string][]string{wsKopf: {w.l.workspace}})
	return holen(t, p.adresse, a)
}

// schreibe stellt eine Mutation mit JSON-Body.
func (w *mockWelt) schreibe(t *testing.T, p *prozess, methode, pfad, body string) antwort {
	t.Helper()
	return w.hole(t, p, anfrage{methode: methode, pfad: pfad, body: body, header: map[string][]string{"Content-Type": {jsonTyp}}})
}

// keinKontakt: der Referenzdienst zählt nach seinem Stoppen weder Verbindung noch Request (SPEC-037 Nr. 7).
func (w *mockWelt) keinKontakt(t *testing.T, name string) {
	t.Helper()
	if got := w.ref.zaehler(); got != w.vor {
		t.Errorf("%s: Zähler (Verbindungen, Requests) %v, nach dem Stoppen %v", name, got, w.vor)
	}
}

// namen liefert die Namen der Entities einer Collection nach `$orderby=Id`.
func (w *mockWelt) namen(t *testing.T, p *prozess) []string {
	t.Helper()
	r := w.hole(t, p, anfrage{pfad: collPfad + "?$orderby=Id"})
	var doc struct{ Value []struct{ Name string } }
	if err := json.Unmarshal(r.body, &doc); err != nil || r.status != 200 {
		t.Fatalf("Collection: %d %q (%v)", r.status, r.body, err)
	}
	var out []string
	for _, e := range doc.Value {
		out = append(out, e.Name)
	}
	return out
}

func (w *mockWelt) pruefeNamen(t *testing.T, p *prozess, schritt string, soll ...string) {
	t.Helper()
	if got := w.namen(t, p); !slices.Equal(got, soll) {
		t.Errorf("nach %s: Collection %v, erwartet %v", schritt, got, soll)
	}
	w.keinKontakt(t, schritt)
}

// TestE2EMock ist die Stufe von LH-FA-07, LH-FA-11 und LH-FA-12 (SPEC-037 Nr. 7): Mock CRUD am gebauten
// Binary, Reset über den Übergang, Zuordnung durch den Header, kein Kontakt zum Referenzdienst.
func TestE2EMock(t *testing.T) {
	binaryPfad(t)
	w := neueMockWelt(t)
	p := w.start(t)
	w.keinKontakt(t, "nach dem Start")
	vorher := w.hole(t, p, anfrage{pfad: collPfad})
	if vorher.status != 200 || vorher.header.Get("X-Recorder-Coverage") != "complete" {
		t.Fatalf("Collection: %d, X-Recorder-Coverage %q, erwartet 200 und complete", vorher.status, vorher.header.Get("X-Recorder-Coverage"))
	}
	w.pruefeNamen(t, p, "Baseline", "Anvil", "Bolt")
	t.Run("Zuordnung", func(t *testing.T) { w.zuordnung(t, p, vorher) })
	t.Run("CRUD", func(t *testing.T) { w.crud(t, p) })
	// Reset: SIGTERM, reset des Übergangs, Start; die Collection gleicht der Antwort vor den Mutationen.
	if code := p.beende(); code != 0 {
		t.Fatalf("SIGTERM: Exit %d\nstderr: %s", code, p.fehl)
	}
	out, errOut, code := uebergang(t, "reset", w.l.dir, w.l.workspace)
	if code != 0 || !strings.HasPrefix(out, "revision=") {
		t.Fatalf("Übergang reset: Exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	p = w.startAuf(t, w.l.dir, p.adresse)
	pruefeGleich(t, "Collection nach dem Reset", vorher, w.hole(t, p, anfrage{pfad: collPfad}))
	w.pruefeNamen(t, p, "Reset", "Anvil", "Bolt")
	if code := p.beende(); code != 0 {
		t.Errorf("SIGTERM nach dem Reset: Exit %d\nstderr: %s", code, p.fehl)
	}
}

// zuordnungsFormen sind die fünf Formen des Headers `X-Recorder-Workspace` (SPEC-195 Nr. 5, 8): vier sind
// `ODR-E5003`, der unbekannte Wert `ODR-E7001`. Die Zeilen tragen den Namen unter der Schreibweise der Form.
func zuordnungsFormen(ws string) []struct {
	name, code string
	status     int
	zeilen     []string
} {
	return []struct {
		name, code string
		status     int
		zeilen     []string
	}{
		{"fehlend", "ODR-E5003", 400, nil},
		{"leer", "ODR-E5003", 400, []string{"X-Recorder-Workspace:"}},
		{"zwei gleiche Werte", "ODR-E5003", 400, []string{"X-Recorder-Workspace: " + ws, "X-Recorder-Workspace: " + ws}},
		{"zwei Schreibweisen", "ODR-E5003", 400, []string{"X-Recorder-Workspace: " + ws, "x-recorder-workspace: " + ws}},
		{"unbekannter Wert", "ODR-E7001", 404, []string{"X-Recorder-Workspace: ws-gibt-es-nicht"}},
	}
}

// zuordnung: jede der fünf Formen an jedem von fünf Requests (lesend und schreibend); der Zustand bleibt
// die Antwort vor den Mutationen.
func (w *mockWelt) zuordnung(t *testing.T, p *prozess, vorher antwort) {
	requests := []struct{ methode, pfad, body string }{
		{"GET", collPfad, ""}, {"GET", "/odata/$metadata", ""}, {"GET", collPfad + "(1)", ""},
		{"POST", collPfad, `{"Id":9,"Name":"Nie"}`}, {"DELETE", collPfad + "(1)", ""},
	}
	n := 0
	for _, f := range zuordnungsFormen(w.l.workspace) {
		for _, r := range requests {
			n++
			a, err := sendeRoh(p.adresse, r.methode, r.pfad, r.body, f.zeilen)
			if err != nil {
				t.Fatalf("%s %s %s: %v", f.name, r.methode, r.pfad, err)
			}
			pruefeFehler(t, f.name+" "+r.methode+" "+r.pfad, a, f.status, f.code)
			w.keinKontakt(t, f.name)
		}
	}
	if n != 25 {
		t.Errorf("Menge hat %d Elemente, erwartet 25 (5 Formen, 5 Requests)", n)
	}
	pruefeGleich(t, "Collection nach den Fehlern der Zuordnung", vorher, w.hole(t, p, anfrage{pfad: collPfad}))
}

// crud: POST (201, Location aus LISTEN), GET, PATCH (204), GET, DELETE (204), GET (404 ODR-E6008), je Schritt
// der Zustand der Collection; ein nicht unterstützter Ausdruck von `$filter` ist 501.
func (w *mockWelt) crud(t *testing.T, p *prozess) {
	post := w.schreibe(t, p, "POST", collPfad, `{"Id":3,"Name":"Cutter"}`)
	if soll := "http://" + p.adresse + "/odata/Products(3)"; post.status != 201 || post.header.Get("Location") != soll {
		t.Fatalf("POST: %d, Location %q, erwartet 201 und %q", post.status, post.header.Get("Location"), soll)
	}
	w.pruefeNamen(t, p, "POST", "Anvil", "Bolt", "Cutter")
	if r := w.hole(t, p, anfrage{pfad: collPfad + "(3)"}); r.status != 200 || !strings.Contains(string(r.body), `"Cutter"`) {
		t.Errorf("GET nach POST: %d %s", r.status, r.body)
	}
	if r := w.schreibe(t, p, "PATCH", collPfad+"(3)", `{"Name":"Cutter2"}`); r.status != 204 {
		t.Fatalf("PATCH: %d %s", r.status, r.body)
	}
	w.pruefeNamen(t, p, "PATCH", "Anvil", "Bolt", "Cutter2")
	if r := w.hole(t, p, anfrage{pfad: collPfad + "(3)"}); r.status != 200 || !strings.Contains(string(r.body), `"Cutter2"`) {
		t.Errorf("GET nach PATCH: %d %s", r.status, r.body)
	}
	if r := w.hole(t, p, anfrage{methode: "DELETE", pfad: collPfad + "(3)"}); r.status != 204 {
		t.Fatalf("DELETE: %d %s", r.status, r.body)
	}
	w.pruefeNamen(t, p, "DELETE", "Anvil", "Bolt")
	pruefeFehler(t, "GET nach DELETE", w.hole(t, p, anfrage{pfad: collPfad + "(3)"}), 404, "ODR-E6008")
	// Ein Aufruf und ein Operatorwort sind gültig und nicht unterstützt (SPEC-164 Nr. 5).
	for _, ausdruck := range []string{"contains(Name,'A')", "Id%20add%201%20eq%202"} {
		pruefeFehler(t, "$filter="+ausdruck, w.hole(t, p, anfrage{pfad: collPfad + "?$filter=" + ausdruck}), 501, "ODR-E6001")
	}
	w.keinKontakt(t, "CRUD")
	// Mutationen für den Reset: Anlegen, Ändern und Löschen an Entities der Baseline.
	w.schreibe(t, p, "POST", collPfad, `{"Id":4,"Name":"Drill"}`)
	w.schreibe(t, p, "PATCH", collPfad+"(1)", `{"Name":"Anvil2"}`)
	w.hole(t, p, anfrage{methode: "DELETE", pfad: collPfad + "(2)"})
	w.pruefeNamen(t, p, "den Mutationen vor dem Reset", "Anvil2", "Drill")
}

// TestE2EMockNeustart ist die Stufe von LH-FA-03 und LH-QA-03 (SPEC-037 Nr. 7): Mutationen an drei Entities,
// SIGTERM, Start auf demselben DATA_DIR, die Zustände sind lesbar; ein zweiter Neustart ohne Änderung ergibt
// dasselbe; ein DATA_DIR, das nicht besteht, wird angelegt.
func TestE2EMockNeustart(t *testing.T) {
	binaryPfad(t)
	t.Run("Mutationen über den Neustart", neustartMutationen)
	t.Run("DATA_DIR besteht nicht", neustartOhneVerzeichnis)
}

// lesezustand sind die Lesezugriffe, deren Antworten vor dem Signal das Orakel nach dem Neustart sind.
func lesezustand() []anfrage {
	return []anfrage{{pfad: collPfad + "?$orderby=Id"}, {pfad: collPfad + "(1)"}, {pfad: collPfad + "(2)"}, {pfad: collPfad + "(3)"}}
}

func neustartMutationen(t *testing.T) {
	w := neueMockWelt(t)
	p := w.start(t)
	if r := w.schreibe(t, p, "POST", collPfad, `{"Id":3,"Name":"Neu"}`); r.status != 201 {
		t.Fatalf("POST: %d %s", r.status, r.body)
	}
	if r := w.schreibe(t, p, "PATCH", collPfad+"(1)", `{"Name":"Anvil2"}`); r.status != 204 {
		t.Fatalf("PATCH: %d %s", r.status, r.body)
	}
	if r := w.hole(t, p, anfrage{methode: "DELETE", pfad: collPfad + "(2)"}); r.status != 204 {
		t.Fatalf("DELETE: %d %s", r.status, r.body)
	}
	// Die Antworten des Lesens sind das Orakel; SIGTERM folgt erst danach.
	var orakel []antwort
	for _, a := range lesezustand() {
		orakel = append(orakel, w.hole(t, p, a))
	}
	w.pruefeNamen(t, p, "den Mutationen", "Anvil2", "Neu")
	for nr := 1; nr <= 2; nr++ {
		p = w.neustart(t, p)
		for i, a := range lesezustand() {
			pruefeGleich(t, fmt.Sprintf("Neustart %d, %s", nr, a.pfad), orakel[i], w.hole(t, p, a))
		}
		w.keinKontakt(t, fmt.Sprintf("Neustart %d", nr))
	}
	if code := p.beende(); code != 0 {
		t.Errorf("SIGTERM des letzten Prozesses: Exit %d\nstderr: %s", code, p.fehl)
	}
}

// neustartOhneVerzeichnis: der Start legt das fehlende DATA_DIR an und ist bereit; jeder Request mit einem
// Workspace ist 404 ODR-E7001 (SPEC-139 Nr. 11 f).
func neustartOhneVerzeichnis(t *testing.T) {
	w := &mockWelt{l: &lauf{}}
	dir := t.TempDir() + "/neu/ablage"
	p := w.startAuf(t, dir, freierPort(t))
	n := 0
	for _, pfad := range []string{collPfad, collPfad + "(1)", "/odata/$metadata", collPfad + "/$count"} {
		n++
		a, err := sendeRoh(p.adresse, "GET", pfad, "", []string{"X-Recorder-Workspace: ws-ohne-ablage"})
		if err != nil {
			t.Fatal(err)
		}
		pruefeFehler(t, "GET "+pfad, a, 404, "ODR-E7001")
	}
	if n != 4 {
		t.Errorf("Menge hat %d Elemente, erwartet 4", n)
	}
	if code := p.beende(); code != 0 {
		t.Errorf("SIGTERM: Exit %d\nstderr: %s", code, p.fehl)
	}
}

// sendeRoh schreibt den Request mit genau den Zeilen kopf auf eine Verbindung: der Client von Go kennt
// weder zwei Schreibweisen desselben Namens noch einen leeren Wert in der Zeile.
func sendeRoh(adresse, methode, pfad, body string, kopf []string) (antwort, error) {
	c, err := net.DialTimeout("tcp", adresse, 5*time.Second)
	if err != nil {
		return antwort{}, err
	}
	defer func() { _ = c.Close() }()
	if err := c.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return antwort{}, err
	}
	zeilen := []string{methode + " " + pfad + " HTTP/1.1", "Host: " + adresse, "Connection: close"}
	if body != "" {
		zeilen = append(zeilen, "Content-Type: "+jsonTyp, fmt.Sprintf("Content-Length: %d", len(body)))
	}
	zeilen = append(zeilen, kopf...)
	if _, err := io.WriteString(c, strings.Join(zeilen, "\r\n")+"\r\n\r\n"+body); err != nil {
		return antwort{}, err
	}
	resp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil {
		return antwort{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	return antwort{status: resp.StatusCode, header: resp.Header, body: b}, err
}
