package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

// Die Matrix der Abnahme über echtes HTTP auf der Schleife (SPEC-042 Nr. 3, 4, 6): die Fixtures liefert abFixtures() der Abnahme, der Eingang ist
// ein httpodata.Server über dem Prüfling.

// vertragMarker sind die Marker der Geheimnisse (SPEC-042 Nr. 6); keine Fixture trägt einen von ihnen als Wert.
func vertragMarker() []string { return []string{"GEHEIM-A", "GEHEIM-B", "GEHEIM-Q"} }

// drahtQuery setzt die Query einer Fixture in die Form um, die auf dem Draht steht (SPEC-042 Nr. 3 c): jedes Byte aus Leerzeichen, ", <, >, \, ^,
// Gravis, {, |, }, unter 0x20, 0x7F und ab 0x80 wird %XX mit Großbuchstaben, jedes andere bleibt.
func drahtQuery(q string) string {
	var b strings.Builder
	for i := 0; i < len(q); i++ {
		c := q[i]
		if c < 0x20 || c >= 0x7F || strings.IndexByte(" \"<>\\^`{|}", c) >= 0 {
			fmt.Fprintf(&b, "%%%02X", c)
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// drahtURL ist die Adresse eines Schritts: /odata, Pfad der Fixture, bei leerer Query ohne ?.
func drahtURL(basis string, a abAnfrage) string {
	u := basis + "/odata" + a.Pfad
	if a.Query != "" {
		u += "?" + drahtQuery(a.Query)
	}
	return u
}

// drahtAntwort ist, was der Client sah: Status, Header (kanonische Namen) und der Body bis zum Ende.
type drahtAntwort struct {
	Status int
	Header http.Header
	Body   []byte
}

// drahtSende schickt den Schritt a an basis: Workspace ws-1, die Marker GEHEIM-A und GEHEIM-B, die Header der Tabelle, bei einem Body dessen
// Content-Type; der Body der Antwort wird bis zum Ende gelesen und geschlossen.
func drahtSende(c *http.Client, basis string, a abAnfrage) (drahtAntwort, error) {
	var body io.Reader
	if a.Body != "" {
		body = strings.NewReader(a.Body)
	}
	req, err := http.NewRequest(a.Methode, drahtURL(basis, a), body)
	if err != nil {
		return drahtAntwort{}, err
	}
	req.Header.Set("X-Recorder-Workspace", "ws-1")
	req.Header.Set("Authorization", "Bearer GEHEIM-A")
	req.Header.Set("Cookie", "GEHEIM-B")
	if a.Body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range a.Header {
		req.Header[http.CanonicalHeaderKey(k)] = v
	}
	resp, err := c.Do(req)
	if err != nil {
		return drahtAntwort{}, err
	}
	b, err := io.ReadAll(resp.Body)
	if cerr := resp.Body.Close(); err == nil {
		err = cerr
	}
	return drahtAntwort{Status: resp.StatusCode, Header: resp.Header, Body: b}, err
}

// vertragClient ist der Client einer ID: ohne Kompression, ohne Weiterleiten.
func vertragClient(t *testing.T) *http.Client {
	t.Helper()
	tr := &http.Transport{DisableCompression: true}
	t.Cleanup(tr.CloseIdleConnections)
	return &http.Client{Transport: tr, Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// vertragServer fährt einen httpodata.Server auf 127.0.0.1:0 über m und liefert seine Basis; das Herunterfahren am Ende des Tests ist Teil des
// Urteils (SPEC-042 Nr. 8 d).
func vertragServer(t *testing.T, id string, m *mock.Mocker) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("%s: Listener: %v", id, err)
	}
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: m.Mock, IDs: supporttest.NewFestIDs(), Listener: l})
	if err != nil {
		t.Fatalf("%s: Server: %v", id, err)
	}
	ende := make(chan struct{})
	go func() {
		_ = srv.Serve()
		close(ende)
	}()
	t.Cleanup(func() {
		if err := srv.Schliessen(5 * time.Second); err != nil {
			t.Errorf("%s: Herunterfahren: %v", id, err)
		}
		<-ende
	})
	return "http://" + l.Addr().String()
}

// vertragFelder liest den Body als Folge von Tokens: die Schlüssel der oberen Ebene, die des Objekts error in der Folge des Bodys und deren Werte;
// fehler ist leer, wenn alle Werte von error Zeichenketten sind und error ein Objekt ist.
func vertragFelder(body []byte) (oben, innen []string, werte map[string]string, fehler string) {
	werte = map[string]string{}
	dec := json.NewDecoder(bytes.NewReader(body))
	if d, ok := vertragToken(dec).(json.Delim); !ok || d != '{' {
		return nil, nil, werte, "der Body ist kein Objekt"
	}
	for dec.More() {
		name, _ := vertragToken(dec).(string)
		oben = append(oben, name)
		if name != "error" {
			var roh json.RawMessage
			_ = dec.Decode(&roh)
			continue
		}
		if d, ok := vertragToken(dec).(json.Delim); !ok || d != '{' {
			return oben, nil, werte, "error ist kein Objekt"
		}
		for dec.More() {
			n, _ := vertragToken(dec).(string)
			innen = append(innen, n)
			var v any
			_ = dec.Decode(&v)
			if s, ok := v.(string); ok {
				werte[n] = s
			} else {
				fehler = "das Feld " + n + " ist keine Zeichenkette"
			}
		}
		vertragToken(dec)
	}
	return oben, innen, werte, fehler
}

func vertragToken(dec *json.Decoder) json.Token {
	tok, err := dec.Token()
	if err != nil {
		return nil
	}
	return tok
}

// vertragFehlerobjekt nennt die Verstöße einer Fehlerantwort gegen SPEC-042 Nr. 4: Content-Type, Content-Length, X-Correlation-Id, genau die drei
// Felder in der Folge code, message, target, code aus SPEC-109 Nr. 2, target und message gleich dem Fehlerobjekt des Codes, keine weiteren Bytes.
func vertragFehlerobjekt(a drahtAntwort) []string {
	var d []string
	if ct := a.Header.Values("Content-Type"); len(ct) != 1 || ct[0] != "application/json" {
		d = append(d, fmt.Sprintf("Content-Type %q, erwartet application/json", ct))
	}
	if cl := a.Header.Values("Content-Length"); len(cl) != 1 || cl[0] != strconv.Itoa(len(a.Body)) {
		d = append(d, fmt.Sprintf("Content-Length %q, Länge des Bodys %d", cl, len(a.Body)))
	}
	if id := a.Header.Values("X-Correlation-Id"); len(id) != 1 || id[0] == "" {
		d = append(d, fmt.Sprintf("X-Correlation-Id %q, erwartet ein nichtleerer Wert", id))
	}
	oben, innen, werte, fehler := vertragFelder(a.Body)
	if fehler != "" {
		d = append(d, "Body: "+fehler)
	}
	if len(oben) != 1 || oben[0] != "error" {
		d = append(d, fmt.Sprintf("Felder der oberen Ebene %v, erwartet [error]", oben))
	}
	switch {
	case len(innen) == 3 && innen[0] == "code" && innen[1] == "message" && innen[2] == "target":
	case len(innen) == 3 && vertragSortiert(innen) == "code,message,target":
		d = append(d, fmt.Sprintf("Reihenfolge der Felder %v, erwartet [code message target]", innen))
	default:
		d = append(d, fmt.Sprintf("Felder von error %v, erwartet genau [code message target]", innen))
	}
	code, ok := model.ParseCode(werte["code"])
	if !ok {
		return append(d, fmt.Sprintf("Code %q steht nicht in SPEC-109 Nr. 2", werte["code"]))
	}
	soll, _ := model.NeuesFehlerobjekt(code)
	if werte["target"] != soll.Error.Target {
		d = append(d, fmt.Sprintf("target %q, erwartet %q", werte["target"], soll.Error.Target))
	}
	if werte["message"] != soll.Error.Message {
		d = append(d, fmt.Sprintf("message %q, erwartet %q", werte["message"], soll.Error.Message))
	}
	kanon, _ := json.Marshal(model.Fehlerobjekt{Error: model.FehlerInhalt{Code: werte["code"], Message: werte["message"], Target: werte["target"]}})
	if len(d) == 0 && !bytes.Equal(kanon, a.Body) {
		d = append(d, fmt.Sprintf("Bytes des Bodys %q, erwartet ohne Leerraum und Zeilenende %q", a.Body, kanon))
	}
	return d
}

func vertragSortiert(l []string) string {
	c := append([]string(nil), l...)
	sort.Strings(c)
	return strings.Join(c, ",")
}

// vertragGeheimnisse nennt jeden Marker, der in einem Header (Name oder Wert) oder im Body der Antwort steht (SPEC-042 Nr. 6).
func vertragGeheimnisse(a drahtAntwort) []string {
	var d []string
	for _, m := range vertragMarker() {
		if bytes.Contains(a.Body, []byte(m)) {
			d = append(d, m+" im Body")
		}
		for k, v := range a.Header {
			if strings.Contains(strings.ToUpper(k), m) ||strings.Contains(strings.Join(v, "\n"), m) {
				d = append(d, m+" im Header "+k)
			}
		}
	}
	return d
}

// vertragCode ist der Wert des Feldes code im Fehlerobjekt (leer, wenn es fehlt).
func vertragCode(body []byte) string {
	_, _, werte, _ := vertragFelder(body)
	return werte["code"]
}

// vertragVergleiche nennt die Unterschiede von ist gegen soll (SPEC-042 Nr. 3 d): Status; bei einem Fehler Code gleich dem der Tabelle und die
// Prüfung von Nr. 4; sonst die Menge der Header ohne Date und Content-Length genau und der Body byte-gleich.
func vertragVergleiche(ist drahtAntwort, soll abSoll) []string {
	var d []string
	if ist.Status != soll.Status {
		d = append(d, fmt.Sprintf("Status %d, erwartet %d", ist.Status, soll.Status))
	}
	if soll.Code != 0 {
		if code := vertragCode(ist.Body); code != soll.Code.String() {
			d = append(d, fmt.Sprintf("Code %q, erwartet %s", code, soll.Code))
		}
		return append(d, vertragFehlerobjekt(ist)...)
	}
	kopf := map[string][]string{}
	for k, v := range ist.Header {
		if k != "Date" && k != "Content-Length" {
			kopf[k] = v
		}
	}
	if !abKopfGleich(kopf, soll.Kopf) {
		d = append(d, fmt.Sprintf("Header %v, erwartet %v", kopf, soll.Kopf))
	}
	if string(ist.Body) != soll.Body {
		d = append(d, fmt.Sprintf("Body %q, erwartet %q", ist.Body, soll.Body))
	}
	return d
}

// vertragLaufe führt die Schritte einer Fixture in einer frischen Umgebung über zwei Server aus (lax; strikt für die Schritte mit Strikt) und prüft
// Revision und Ausgangszustand wie die Abnahme (SPEC-042 Nr. 3 b, e); es liefert die Zahl der gelaufenen Schritte.
func vertragLaufe(t *testing.T, f abFixture) int {
	t.Helper()
	u := abFrisch(t)
	abPruefeAusgang(t, u)
	lax, strikt := vertragServer(t, f.ID, u.mock), vertragServer(t, f.ID, u.strikt)
	c := vertragClient(t)
	n := 0
	for _, s := range f.Schritte {
		basis := lax
		if s.Strikt {
			basis = strikt
		}
		ist, err := drahtSende(c, basis, s.Anf)
		if err != nil {
			t.Fatalf("%s %s %s %s: %v", f.ID, s.Name, s.Anf.Methode, s.Anf.Pfad, err)
		}
		n++
		for _, d := range vertragVergleiche(ist, s.Soll) {
			t.Errorf("%s %s %s %s: %s", f.ID, s.Name, s.Anf.Methode, s.Anf.Pfad, d)
		}
		for _, d := range vertragGeheimnisse(ist) {
			t.Errorf("%s %s %s %s: Geheimnis %s", f.ID, s.Name, s.Anf.Methode, s.Anf.Pfad, d)
		}
	}
	if w, err := u.db.Workspaces().Get(context.Background(), "ws-1"); err != nil || w.Revision != f.Revision {
		t.Errorf("%s: Revision %d (%v), erwartet %d", f.ID, w.Revision, err, f.Revision)
	}
	if f.Revision == 0 {
		for _, d := range abStandGleich(u.stand("ws-1"), abAusgangsstand("ws-1", 0)) {
			t.Errorf("%s: Zustand nach der ID, erwartet der Ausgangszustand: %s", f.ID, d)
		}
	}
	return n
}

// TestVertragMatrix (T01, T05 bis T07, T15): die 36 IDs der Abnahme über HTTP, jede einmal, jeder Schritt gegen die Tabelle.
func TestVertragMatrix(t *testing.T) {
	gelaufen := map[string]int{}
	fixtures := abFixtures()
	for _, f := range fixtures {
		t.Run(f.ID, func(t *testing.T) { gelaufen[f.ID] = vertragLaufe(t, f) })
	}
	for _, d := range abPruefeZaehlung(fixtures, gelaufen) {
		t.Error(d)
	}
}

// TestVertragMatrixZaehlung (T01): die Tabelle führt 36 IDs mit 92 Schritten; je ein Lauf ohne eine ID, ohne einen Schritt, mit einer ID doppelt und mit
// einem nicht gelaufenen Schritt ist ein Verstoß, und zwar für jede ID und jeden Schritt der Tabelle.
func TestVertragMatrixZaehlung(t *testing.T) {
	fixtures := abFixtures()
	gelaufen := map[string]int{}
	for _, f := range fixtures {
		gelaufen[f.ID] = len(f.Schritte)
	}
	if d := abPruefeZaehlung(fixtures, gelaufen); len(d) != 0 {
		t.Fatalf("Tabelle: %v", d)
	}
	faelle := 0
	for i, f := range fixtures {
		ohneID := append(append([]abFixture(nil), fixtures[:i]...), fixtures[i+1:]...)
		doppelt := append(append([]abFixture(nil), fixtures...), f)
		for name, ids := range map[string][]abFixture{"ID fehlt": ohneID, "ID doppelt": doppelt} {
			faelle++
			if len(abPruefeZaehlung(ids, gelaufen)) == 0 {
				t.Errorf("%s %s: kein Verstoß gemeldet", name, f.ID)
			}
		}
		for k := range f.Schritte {
			weniger := map[string]int{}
			for id, n := range gelaufen {
				weniger[id] = n
			}
			weniger[f.ID] = k
			faelle++
			if len(abPruefeZaehlung(fixtures, weniger)) == 0 {
				t.Errorf("%s: nur %d von %d Schritten gelaufen, kein Verstoß gemeldet", f.ID, k, len(f.Schritte))
			}
		}
	}
	if faelle != 2*36+92 {
		t.Errorf("%d Fälle gelaufen, erwartet %d", faelle, 2*36+92)
	}
}

// vertragErfolg ist eine Erfolgsantwort, die dem Soll genügt, mit Date und Content-Length des Servers.
func vertragErfolg(soll abSoll) drahtAntwort {
	h := http.Header{}
	for k, v := range soll.Kopf {
		h[k] = append([]string(nil), v...)
	}
	h.Set("Date", "Mon, 02 Jan 2006 15:04:05 GMT")
	h.Set("Content-Length", strconv.Itoa(len(soll.Body)))
	return drahtAntwort{Status: soll.Status, Header: h, Body: []byte(soll.Body)}
}

// vertragFehlerAntwort ist die Fehlerantwort des Adapters zu code und status (SPEC-136 Nr. 8).
func vertragFehlerAntwort(status int, code model.Code) drahtAntwort {
	obj, _ := model.NeuesFehlerobjekt(code)
	b, _ := json.Marshal(obj)
	return drahtAntwort{Status: status, Body: b, Header: http.Header{"Content-Type": {"application/json"}, "Content-Length": {strconv.Itoa(len(b))}, "X-Correlation-Id": {"cor-1"}}}
}

// TestVertragVergleich (T02, T03, T08): der Vergleich nennt die Menge der Header genau (allein Date und Content-Length einer Erfolgsantwort zählen
// nicht), den Body byte-gleich, bei einem Fehler Status und Code; die Prüfung des Fehlerobjekts nennt jede ihrer Eigenschaften.
func TestVertragVergleich(t *testing.T) {
	soll := abKK(`{"a":1}`)
	type fall struct {
		name   string
		ist    drahtAntwort
		rot    bool
		soll   abSoll
		stichw string
	}
	var faelle []fall
	add := func(name string, rot bool, mut func(*drahtAntwort)) {
		a := vertragErfolg(soll)
		mut(&a)
		faelle = append(faelle, fall{name: name, ist: a, rot: rot, soll: soll})
	}
	add("gleich", false, func(*drahtAntwort) {})
	add("nur Date anders", false, func(a *drahtAntwort) { a.Header.Set("Date", "Tue, 03 Jan 2006 15:04:05 GMT") })
	add("nur Content-Length anders", false, func(a *drahtAntwort) { a.Header.Set("Content-Length", "999") })
	add("Date fehlt", false, func(a *drahtAntwort) { a.Header.Del("Date") })
	add("Content-Length fehlt", false, func(a *drahtAntwort) { a.Header.Del("Content-Length") })
	add("weiterer Header", true, func(a *drahtAntwort) { a.Header.Set("X-Weiter", "1") })
	add("Status anders", true, func(a *drahtAntwort) { a.Status = 201 })
	for name := range soll.Kopf {
		add("Header "+name+" fehlt", true, func(a *drahtAntwort) { a.Header.Del(name) })
		add("Header "+name+" anders", true, func(a *drahtAntwort) { a.Header.Set(name, "anders") })
		add("Header "+name+" doppelt", true, func(a *drahtAntwort) { a.Header.Add(name, a.Header.Get(name)) })
	}
	for name, b := range map[string]string{"Zeilenende": soll.Body + "\n", "Leerzeichen": soll.Body + " ", "gekürzt": soll.Body[:len(soll.Body)-1], "anders": `{"a":2}`} {
		add("Body "+name, true, func(a *drahtAntwort) { a.Body = []byte(b) })
	}
	fehler := abF(404, model.CodeE6008)
	for _, f := range []struct {
		name string
		ist  drahtAntwort
		rot  bool
	}{
		{"Fehler gleich", vertragFehlerAntwort(404, model.CodeE6008), false},
		{"Fehler anderer Status", vertragFehlerAntwort(400, model.CodeE6008), true},
		{"Fehler anderer Code", vertragFehlerAntwort(404, model.CodeE6003), true},
		{"Erfolg statt Fehler", vertragErfolg(soll), true},
	} {
		faelle = append(faelle, fall{name: f.name, ist: f.ist, rot: f.rot, soll: fehler})
	}
	faelle = append(faelle, fall{name: "Fehler statt Erfolg", ist: vertragFehlerAntwort(200, model.CodeE6008), rot: true, soll: soll})
	for _, f := range faelle {
		if got := vertragVergleiche(f.ist, f.soll); (len(got) != 0) != f.rot {
			t.Errorf("%s: Verstöße %v, erwartet rot=%v", f.name, got, f.rot)
		}
	}
	if len(faelle) != 7+3*len(soll.Kopf)+4+5 {
		t.Errorf("%d Fälle, erwartet %d", len(faelle), 7+3*len(soll.Kopf)+4+5)
	}
	t.Run("Fehlerobjekt", func(t *testing.T) { vertragFehlerobjektFaelle(t) })
	t.Run("Geheimnisse", func(t *testing.T) {
		if d := vertragGeheimnisse(vertragErfolg(soll)); len(d) != 0 {
			t.Errorf("Antwort ohne Marker gemeldet: %v", d)
		}
		for _, m := range vertragMarker() {
			imBody, imWert, imName := vertragErfolg(soll), vertragErfolg(soll), vertragErfolg(soll)
			imBody.Body = append(imBody.Body, m...)
			imWert.Header.Set("X-Echo", "Bearer "+m)
			imName.Header.Set("X-"+m, "1")
			for name, a := range map[string]drahtAntwort{"Body": imBody, "Header-Wert": imWert, "Header-Name": imName} {
				if len(vertragGeheimnisse(a)) == 0 {
					t.Errorf("Marker %s im %s: nicht gemeldet", m, name)
				}
			}
		}
	})
}

// vertragCodes sind die Codes, die der Eingang in den Fixtures und den Klassen liefert.
func vertragCodes() []model.Code {
	return []model.Code{model.CodeE5001, model.CodeE6001, model.CodeE6002, model.CodeE6003, model.CodeE6004, model.CodeE6005, model.CodeE6007,
		model.CodeE6008, model.CodeE6009}
}

// vertragFehlerobjektFaelle (T08): je Code der Menge und je Eigenschaft ein verletztes Beispiel; das Beispiel ist rot aus dem benannten Grund
// (das Stichwort steht in einem Verstoß), das unverletzte ist grün.
func vertragFehlerobjektFaelle(t *testing.T) {
	body := func(code, message, target string) string {
		return fmt.Sprintf(`{"error":{"code":%q,"message":%q,"target":%q}}`, code, message, target)
	}
	zahl := 0
	for _, c := range vertragCodes() {
		obj, _ := model.NeuesFehlerobjekt(c)
		code, msg, ziel := obj.Error.Code, obj.Error.Message, obj.Error.Target
		gut := body(code, msg, ziel)
		if d := vertragFehlerobjekt(vertragFehlerAntwort(422, c)); len(d) != 0 {
			t.Errorf("%s: gültige Antwort gemeldet: %v", code, d)
		}
		mitBody := func(b string) drahtAntwort {
			a := vertragFehlerAntwort(422, c)
			a.Body = []byte(b)
			a.Header.Set("Content-Length", strconv.Itoa(len(b)))
			return a
		}
		mitKopf := func(mut func(http.Header)) drahtAntwort {
			a := vertragFehlerAntwort(422, c)
			mut(a.Header)
			return a
		}
		for _, v := range []struct {
			name, stichwort string
			ist             drahtAntwort
		}{
			{"Content-Type anders", "Content-Type", mitKopf(func(h http.Header) { h.Set("Content-Type", "text/plain") })},
			{"Content-Type mit charset", "Content-Type", mitKopf(func(h http.Header) { h.Set("Content-Type", "application/json; charset=utf-8") })},
			{"Content-Type fehlt", "Content-Type", mitKopf(func(h http.Header) { h.Del("Content-Type") })},
			{"Content-Type doppelt", "Content-Type", mitKopf(func(h http.Header) { h.Add("Content-Type", "application/json") })},
			{"Content-Length zu groß", "Content-Length", mitKopf(func(h http.Header) { h.Set("Content-Length", strconv.Itoa(len(gut)+1)) })},
			{"Content-Length fehlt", "Content-Length", mitKopf(func(h http.Header) { h.Del("Content-Length") })},
			{"X-Correlation-Id leer", "X-Correlation-Id", mitKopf(func(h http.Header) { h.Set("X-Correlation-Id", "") })},
			{"X-Correlation-Id fehlt", "X-Correlation-Id", mitKopf(func(h http.Header) { h.Del("X-Correlation-Id") })},
			{"viertes Feld", "Felder von error", mitBody(`{"error":{"code":"` + code + `","message":"` + msg + `","target":"` + ziel + `","x":"y"}}`)},
			{"Feld target fehlt", "Felder von error", mitBody(`{"error":{"code":"` + code + `","message":"` + msg + `"}}`)},
			{"Feld weiter oben", "oberen Ebene", mitBody(`{"error":{"code":"` + code + `","message":"` + msg + `","target":"` + ziel + `"},"x":"y"}`)},
			{"Reihenfolge", "Reihenfolge", mitBody(`{"error":{"message":"` + msg + `","code":"` + code + `","target":"` + ziel + `"}}`)},
			{"target anders", "target", mitBody(body(code, msg, "config"))},
			{"message anders", "message", mitBody(body(code, msg+"x", ziel))},
			{"Code außerhalb von SPEC-109", "SPEC-109", mitBody(body("ODR-E0000", msg, ziel))},
			{"Leerraum im Body", "Bytes des Bodys", mitBody(strings.Replace(gut, `","message"`, `", "message"`, 1))},
			{"Zeilenende am Ende", "Bytes des Bodys", mitBody(gut + "\n")},
			{"Feld keine Zeichenkette", "Zeichenkette", mitBody(`{"error":{"code":"` + code + `","message":"` + msg + `","target":1}}`)},
			{"error kein Objekt", "kein Objekt", mitBody(`{"error":"x"}`)},
			{"leerer Body", "kein Objekt", mitBody("")},
		} {
			zahl++
			d := vertragFehlerobjekt(v.ist)
			if !strings.Contains(strings.Join(d, "\n"), v.stichwort) {
				t.Errorf("%s %s: Verstöße %v, erwartet einer mit %q", code, v.name, d, v.stichwort)
			}
		}
	}
	if zahl != len(vertragCodes())*20 {
		t.Errorf("%d Fälle, erwartet %d", zahl, len(vertragCodes())*20)
	}
}

// TestVertragDrahtQuery (T04): über alle 256 Byte-Werte ist drahtQuery %XX mit Großbuchstaben genau für die Menge von SPEC-042 Nr. 3 c und lässt jedes
// andere Byte unverändert; %41 bleibt, die leere Query trägt kein ?.
func TestVertragDrahtQuery(t *testing.T) {
	kodiert := 0
	for i := 0; i < 256; i++ {
		b := byte(i)
		roh := "a" + string([]byte{b}) + "z"
		soll := roh
		if b < 0x20 || b >= 0x7F || strings.ContainsRune(" \"<>\\^`{|}", rune(b)) {
			soll = "a" + fmt.Sprintf("%%%02X", b) + "z"
			kodiert++
		}
		if got := drahtQuery(roh); got != soll {
			t.Errorf("Byte 0x%02X: %q, erwartet %q", b, got, soll)
		}
	}
	if kodiert != 10+32+1+128 {
		t.Errorf("%d Bytes kodiert, erwartet %d", kodiert, 10+32+1+128)
	}
	if got := drahtQuery("$filter=Id eq 1&a=%41"); got != "$filter=Id%20eq%201&a=%41" {
		t.Errorf("%%41 oder Leerzeichen: %q", got)
	}
	if got := drahtURL("http://x", abAnfrage{Methode: "GET", Pfad: "/Products"}); got != "http://x/odata/Products" {
		t.Errorf("leere Query: %q", got)
	}
	if got := drahtURL("http://x", abAnfrage{Methode: "GET", Pfad: "/Products", Query: "a b"}); got != "http://x/odata/Products?a%20b" {
		t.Errorf("Query: %q", got)
	}
}
