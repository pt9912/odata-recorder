package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/blobs"
	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die vier Fehlerklassen von LH-QA-05 am HTTP-Eingang (SPEC-042 Nr. 4 bis 6), auf m3Aufbau: Aufnahme, Import und Workspace wie TestM3.

// vertragMarker sind die Marker der Geheimnisse (SPEC-042 Nr. 6).
func vertragMarker() []string { return []string{"GEHEIM-A", "GEHEIM-B", "GEHEIM-Q"} }

// vertragKlasse ist eine Zeile der Tabelle von SPEC-042 Nr. 5: Aufbau, Anfragen (Pfad mit Query) und die erwartete Antwort.
type vertragKlasse struct {
	name   string
	status int
	code   model.Code
	ziel   string
	pfade  []string
	q      bool // die Anfragen der Klasse tragen GEHEIM-Q in der Query (SPEC-042 Nr. 6)
	start  func(t *testing.T) (*client, map[string]string)
}

// vertragKlassen ist die Tabelle: vier Klassen, je mit dem Code, den der Eingang liefert.
func vertragKlassen() []vertragKlasse {
	return []vertragKlasse{
		{"fehlender Replay-Treffer", 404, model.CodeE5001, "replay", []string{"/odata/Products(99)?x=GEHEIM-Q"}, true, vertragReplay},
		// Die Zeile führt allein die Collection (SPEC-042 Nr. 9 i).
		{"ungültige Metadaten", 422, model.CodeE6004, "import", []string{"/odata/Products"}, false, vertragKaputt},
		{"nicht unterstützte Query-Option", 501, model.CodeE6001, "mock",
			[]string{"/odata/Products?$search=GEHEIM-Q", "/odata/Products?$apply=GEHEIM-Q", "/odata/Products?$expand=GEHEIM-Q"}, true, vertragMock},
		{"unvollständige Coverage", 422, model.CodeE6003, "mock", []string{"/odata/Products"}, false, vertragTeil},
	}
}

// vertragHoch baut und startet den Handle wie hochgefahren; ein Fehler beim Herunterfahren am Ende des Tests ist rot (SPEC-042 Nr. 8 d).
func vertragHoch(t *testing.T, k bootstrap.Konfiguration) *bootstrap.Handle {
	t.Helper()
	ctx := context.Background()
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		t.Fatalf("Build %s: %v", k.Modus, err)
	}
	t.Cleanup(func() {
		if err := h.Herunterfahren(context.WithoutCancel(ctx)); err != nil {
			t.Errorf("Herunterfahren %s: %v", k.Modus, err)
		}
	})
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatalf("Hochfahren %s: %v", k.Modus, err)
	}
	return h
}

// vertragMockStart fährt den Modus mock auf DATA_DIR von w hoch; extra sind weitere Zeilen der Umgebung.
func vertragMockStart(t *testing.T, w *m3Welt, extra ...string) *client {
	t.Helper()
	return neuerClient(t, vertragHoch(t, konfig(t, "mock", w.dir, "", extra...)).Adresse())
}

// vertragGesehen ist, was der Client abschickte: die Werte von Authorization und Cookie und die Query.
type vertragGesehen struct {
	Authorization, Cookie []string
	Query                 string
}

// vertragSpiegel hält am Transport des Clients fest, was jede Anfrage trug (SPEC-042 Nr. 6, Sendeseite).
type vertragSpiegel struct {
	next  http.RoundTripper
	liste []vertragGesehen
}

func (s *vertragSpiegel) RoundTrip(r *http.Request) (*http.Response, error) {
	s.liste = append(s.liste, vertragGesehen{append([]string(nil), r.Header.Values("Authorization")...), append([]string(nil), r.Header.Values("Cookie")...), r.URL.RawQuery})
	return s.next.RoundTrip(r)
}

// vertragSendeseite nennt die Verstöße gegen die Sendeseite von SPEC-042 Nr. 6: genau n Anfragen gingen aus, jede mit Authorization
// `Bearer GEHEIM-A` und Cookie `GEHEIM-B`, und die Query trägt GEHEIM-Q genau dann, wenn mitQ gilt.
func vertragSendeseite(g []vertragGesehen, n int, mitQ bool) []string {
	var d []string
	if len(g) != n {
		d = append(d, fmt.Sprintf("%d Anfragen gesendet, erwartet %d", len(g), n))
	}
	for i, a := range g {
		if len(a.Authorization) != 1 || a.Authorization[0] != "Bearer GEHEIM-A" {
			d = append(d, fmt.Sprintf("Anfrage %d: Authorization %q, erwartet Bearer GEHEIM-A", i, a.Authorization))
		}
		if len(a.Cookie) != 1 || a.Cookie[0] != "GEHEIM-B" {
			d = append(d, fmt.Sprintf("Anfrage %d: Cookie %q, erwartet GEHEIM-B", i, a.Cookie))
		}
		if strings.Contains(a.Query, "GEHEIM-Q") != mitQ {
			d = append(d, fmt.Sprintf("Anfrage %d: Query %q, GEHEIM-Q erwartet=%v", i, a.Query, mitQ))
		}
	}
	return d
}

// vertragReplay fährt den Modus replay auf der Aufnahme hoch; die Zuordnung nennt Workspace, aufgenommene Session und eine Replay-Session.
func vertragReplay(t *testing.T) (*client, map[string]string) {
	w := m3Aufbau(t)
	db := vertragOeffne(t, w.dir)
	sessions, err := db.Sessions().List(context.Background(), serviceID)
	if cerr := db.Close(); err != nil || cerr != nil || len(sessions) == 0 {
		t.Fatalf("Sessions der Aufnahme: %v, %v, %v", sessions, err, cerr)
	}
	h := vertragHoch(t, konfig(t, "replay", w.dir, ""))
	z := kopf(w.ws)
	z["X-Recorder-Recording-Session"] = sessions[0].ID // die erste Session ist die der Baseline von w.ws
	z["X-Recorder-Replay-Session"] = "vertrag-1"
	return neuerClient(t, h.Adresse()), z
}

// vertragOeffne öffnet die Datei der Ablage in dir bei gestopptem Handle.
func vertragOeffne(t *testing.T, dir string) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(context.Background(), filepath.Join(dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		t.Fatalf("sqlite.Open: %v", err)
	}
	return db
}

// vertragKaputt legt vor dem Start ein Dokument <kaputt als MetadataBlob des Dienstes ab und fährt den Modus mock hoch (SPEC-042 Nr. 5).
func vertragKaputt(t *testing.T) (*client, map[string]string) {
	w := m3Aufbau(t)
	ctx := context.Background()
	db := vertragOeffne(t, w.dir)
	store, err := blobs.New(filepath.Join(w.dir, "blobs"))
	if err != nil {
		t.Fatalf("blobs.New: %v", err)
	}
	dok := []byte("<kaputt")
	wr, err := store.Begin(ctx, int64(len(dok))+1)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if _, err := wr.Write(dok); err != nil {
		t.Fatalf("Write: %v", err)
	}
	ref, _, err := wr.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := db.Metadata().SetMetadata(ctx, serviceID, ref, csdlHash()); err != nil {
		t.Fatalf("SetMetadata: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return vertragMockStart(t, w), kopf(w.ws)
}

// vertragMock fährt den Modus mock lax hoch, Workspace aus der vollständigen Baseline.
func vertragMock(t *testing.T) (*client, map[string]string) {
	w := m3Aufbau(t)
	return vertragMockStart(t, w), kopf(w.ws)
}

// vertragTeil fährt den Modus mock mit STRICT_COVERAGE=true hoch, Workspace aus der Baseline mit nicht vollständiger Collection.
func vertragTeil(t *testing.T) (*client, map[string]string) {
	w := m3Aufbau(t)
	return vertragMockStart(t, w, "STRICT_COVERAGE=true"), kopf(w.wsTeil)
}

// vertragPruefeKlassen nennt die Verstöße der Tabelle: genau vier Klassen, die Codes paarweise verschieden (SPEC-042 Nr. 5).
func vertragPruefeKlassen(k []vertragKlasse) []string {
	var d []string
	if len(k) != 4 {
		d = append(d, fmt.Sprintf("%d Klassen, erwartet genau 4", len(k)))
	}
	for i := range k {
		for j := i + 1; j < len(k); j++ {
			if k[i].code == k[j].code {
				d = append(d, fmt.Sprintf("%q und %q tragen denselben Code %s", k[i].name, k[j].name, k[i].code))
			}
		}
	}
	return d
}

// vertragGeheimnisse nennt jeden Marker, der in einem Header (Name oder Wert) oder im Body der Antwort steht.
func vertragGeheimnisse(a antwort) []string {
	var d []string
	for _, m := range vertragMarker() {
		if bytes.Contains(a.body, []byte(m)) {
			d = append(d, m+" im Body")
		}
		for k, v := range a.header {
			if strings.Contains(strings.ToUpper(k), m) || strings.Contains(strings.Join(v, "\n"), m) {
				d = append(d, m+" im Header "+k)
			}
		}
	}
	return d
}

func vertragToken(dec *json.Decoder) json.Token {
	tok, err := dec.Token()
	if err != nil {
		return nil
	}
	return tok
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

// vertragFehlerobjekt nennt die Verstöße einer Fehlerantwort gegen SPEC-042 Nr. 4: Content-Type, Content-Length, X-Correlation-Id, genau die drei
// Felder in der Folge code, message, target, code aus SPEC-109 Nr. 2, target und message gleich dem Fehlerobjekt des Codes, keine weiteren Bytes.
func vertragFehlerobjekt(a antwort) []string {
	var d []string
	if ct := a.header.Values("Content-Type"); len(ct) != 1 || ct[0] != "application/json" {
		d = append(d, fmt.Sprintf("Content-Type %q, erwartet application/json", ct))
	}
	if cl := a.header.Values("Content-Length"); len(cl) != 1 || cl[0] != strconv.Itoa(len(a.body)) {
		d = append(d, fmt.Sprintf("Content-Length %q, Länge des Bodys %d", cl, len(a.body)))
	}
	if id := a.header.Values("X-Correlation-Id"); len(id) != 1 || id[0] == "" {
		d = append(d, fmt.Sprintf("X-Correlation-Id %q, erwartet ein nichtleerer Wert", id))
	}
	oben, innen, werte, fehler := vertragFelder(a.body)
	if fehler != "" {
		d = append(d, "Body: "+fehler)
	}
	if len(oben) != 1 || oben[0] != "error" {
		d = append(d, fmt.Sprintf("Felder der oberen Ebene %v, erwartet [error]", oben))
	}
	switch {
	case len(innen) == 3 && innen[0] == "code" && innen[1] == "message" && innen[2] == "target":
	case vertragGleichesFeldSet(innen):
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
	if len(d) == 0 && !bytes.Equal(kanon, a.body) {
		d = append(d, fmt.Sprintf("Bytes des Bodys %q, erwartet ohne Leerraum und Zeilenende %q", a.body, kanon))
	}
	return d
}

// vertragGleichesFeldSet meldet, ob l genau die Felder code, message und target in beliebiger Folge nennt.
func vertragGleichesFeldSet(l []string) bool {
	n := map[string]bool{}
	for _, s := range l {
		n[s] = true
	}
	return len(l) == 3 && n["code"] && n["message"] && n["target"]
}

// TestVertragFehlerklassen (T09 bis T15): je Klasse von LH-QA-05 ein Negativfall am HTTP-Eingang, mit Status, Code, Zielbereich, Fehlerobjekt und ohne
// Geheimnis; die Codes der vier Klassen sind paarweise verschieden.
func TestVertragFehlerklassen(t *testing.T) {
	klassen := vertragKlassen()
	for _, d := range vertragPruefeKlassen(klassen) {
		t.Error(d)
	}
	t.Run("Tabelle", func(t *testing.T) {
		ohne := klassen[1:]
		fuenf := append(append([]vertragKlasse(nil), klassen...), vertragKlasse{name: "fünfte", code: model.CodeE6008})
		gleich := append([]vertragKlasse(nil), klassen...)
		gleich[1].code = gleich[3].code
		for name, k := range map[string][]vertragKlasse{"eine Klasse fehlt": ohne, "eine fünfte Klasse": fuenf, "zwei Klassen mit demselben Code": gleich} {
			if len(vertragPruefeKlassen(k)) == 0 {
				t.Errorf("%s: kein Verstoß gemeldet", name)
			}
		}
	})
	var alle []antwort
	for _, k := range klassen {
		t.Run(k.name, func(t *testing.T) {
			c, z := k.start(t)
			sp := &vertragSpiegel{next: c.c.Transport}
			c.c.Transport = sp
			t.Cleanup(func() {
				for _, d := range vertragSendeseite(sp.liste, len(k.pfade), k.q) {
					t.Errorf("%s: %s", k.name, d)
				}
			})
			zuordnung := map[string]string{"Authorization": "Bearer GEHEIM-A", "Cookie": "GEHEIM-B"}
			for n, v := range z {
				zuordnung[n] = v
			}
			for _, pfad := range k.pfade {
				a := c.sende(http.MethodGet, pfad, "", "", zuordnung)
				alle = append(alle, a)
				if a.status != k.status {
					t.Errorf("%s: Status %d, erwartet %d (%s)", pfad, a.status, k.status, a.body)
				}
				if got := vertragCodeAus(a.body); got != k.code.String() {
					t.Errorf("%s: Code %q, erwartet %s", pfad, got, k.code)
				}
				if _, _, werte, _ := vertragFelder(a.body); werte["target"] != k.ziel {
					t.Errorf("%s: target %q, erwartet %q", pfad, werte["target"], k.ziel)
				}
				if v := a.header.Values("X-Recorder-Coverage"); len(v) != 0 {
					t.Errorf("%s: Header X-Recorder-Coverage %q, erwartet keiner", pfad, v)
				}
				for _, d := range vertragFehlerobjekt(a) {
					t.Errorf("%s: %s", pfad, d)
				}
			}
		})
	}
	if len(alle) != 1+1+3+1 {
		t.Errorf("%d Antworten, erwartet %d", len(alle), 1+1+3+1)
	}
	for i, a := range alle {
		for _, d := range vertragGeheimnisse(a) {
			t.Errorf("Antwort %d: Geheimnis %s", i, d)
		}
	}
}

func vertragCodeAus(body []byte) string {
	_, _, werte, _ := vertragFelder(body)
	return werte["code"]
}

// vertragCodes sind die Codes, die der Eingang in den Klassen und in den Fixtures der Matrix liefert.
func vertragCodes() []model.Code {
	return []model.Code{model.CodeE5001, model.CodeE6001, model.CodeE6002, model.CodeE6003, model.CodeE6004, model.CodeE6005, model.CodeE6007,
		model.CodeE6008, model.CodeE6009}
}

func vertragFehlerAntwort(code model.Code) antwort {
	obj, _ := model.NeuesFehlerobjekt(code)
	b, _ := json.Marshal(obj)
	return antwort{status: 422, body: b, header: http.Header{"Content-Type": {"application/json"}, "Content-Length": {strconv.Itoa(len(b))}, "X-Correlation-Id": {"cor-1"}}}
}

// TestVertragFehlerobjekt (T08, T15): je Code der Menge und je Eigenschaft von SPEC-042 Nr. 4 ein verletztes Beispiel, das aus dem benannten Grund rot
// ist (das Stichwort steht in einem Verstoß); das unverletzte Beispiel ist grün. Der Marker steht in keiner Antwort, in der er fehlt, und wird in
// jeder gefunden, in der er steht.
func TestVertragFehlerobjekt(t *testing.T) {
	body := func(code, message, target string) string {
		return fmt.Sprintf(`{"error":{"code":%q,"message":%q,"target":%q}}`, code, message, target)
	}
	zahl := 0
	for _, c := range vertragCodes() {
		obj, _ := model.NeuesFehlerobjekt(c)
		code, msg, ziel := obj.Error.Code, obj.Error.Message, obj.Error.Target
		gut := body(code, msg, ziel)
		if d := vertragFehlerobjekt(vertragFehlerAntwort(c)); len(d) != 0 {
			t.Errorf("%s: gültige Antwort gemeldet: %v", code, d)
		}
		mitBody := func(b string) antwort {
			a := vertragFehlerAntwort(c)
			a.body = []byte(b)
			a.header.Set("Content-Length", strconv.Itoa(len(b)))
			return a
		}
		mitKopf := func(mut func(http.Header)) antwort {
			a := vertragFehlerAntwort(c)
			mut(a.header)
			return a
		}
		for _, v := range []struct {
			name, stichwort string
			ist             antwort
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
	// Sendeseite: je Marker (A in Authorization, B in Cookie, Q in der Query) und je Zahl der Anfragen ein verletztes Beispiel.
	gut := vertragGesehen{[]string{"Bearer GEHEIM-A"}, []string{"GEHEIM-B"}, "x=GEHEIM-Q"}
	if d := vertragSendeseite([]vertragGesehen{gut}, 1, true); len(d) != 0 {
		t.Errorf("gültige Anfrage gemeldet: %v", d)
	}
	for name, f := range map[string]struct {
		g     []vertragGesehen
		n     int
		mitQ  bool
		stich string
	}{
		"Authorization fehlt":    {[]vertragGesehen{{nil, gut.Cookie, gut.Query}}, 1, true, "Authorization"},
		"Authorization anders":   {[]vertragGesehen{{[]string{"x"}, gut.Cookie, gut.Query}}, 1, true, "Authorization"},
		"Cookie fehlt":           {[]vertragGesehen{{gut.Authorization, nil, gut.Query}}, 1, true, "Cookie"},
		"Cookie anders":          {[]vertragGesehen{{gut.Authorization, []string{"x"}, gut.Query}}, 1, true, "Cookie"},
		"GEHEIM-Q fehlt":         {[]vertragGesehen{{gut.Authorization, gut.Cookie, "x=1"}}, 1, true, "Query"},
		"GEHEIM-Q unerwartet":    {[]vertragGesehen{gut}, 1, false, "Query"},
		"eine Anfrage zu wenig":  {[]vertragGesehen{gut}, 2, true, "Anfragen gesendet"},
		"keine Anfrage gesendet": {nil, 1, true, "Anfragen gesendet"},
	} {
		if d := strings.Join(vertragSendeseite(f.g, f.n, f.mitQ), "\n"); !strings.Contains(d, f.stich) {
			t.Errorf("Sendeseite %s: Verstöße %q, erwartet einer mit %q", name, d, f.stich)
		}
	}
	for _, m := range vertragMarker() {
		sauber := vertragFehlerAntwort(model.CodeE6001)
		if d := vertragGeheimnisse(sauber); len(d) != 0 {
			t.Errorf("Marker %s in einer Antwort ohne Marker gemeldet: %v", m, d)
		}
		imBody, imWert, imName := vertragFehlerAntwort(model.CodeE6001), vertragFehlerAntwort(model.CodeE6001), vertragFehlerAntwort(model.CodeE6001)
		imBody.body = append(imBody.body, m...)
		imWert.header.Set("X-Echo", "Bearer "+m)
		imName.header.Set("X-"+m, "1")
		for name, a := range map[string]antwort{"Body": imBody, "Header-Wert": imWert, "Header-Name": imName} {
			if len(vertragGeheimnisse(a)) == 0 {
				t.Errorf("Marker %s im %s: nicht gemeldet", m, name)
			}
		}
	}
}
