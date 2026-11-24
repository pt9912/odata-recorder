package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testBasis = "http://referenz.test"

// bediene ruft den Handler des Referenzdienstes und liefert Status, Header und Body.
func bediene(h http.Handler, methode, ziel string) (int, http.Header, []byte) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(methode, testBasis+ziel, nil))
	return rec.Code, rec.Header(), rec.Body.Bytes()
}

// Die Namen der Defaults von REDACT_HEADERS, REDACT_QUERY_PARAMS und REDACT_JSON_PATHS (SPEC-110), als Literale.
func standardHeader() []string {
	return []string{"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie", "X-API-Key", "X-Auth-Token", "X-CSRF-Token"}
}

func standardQuery() []string {
	return []string{"access_token", "token", "api_key", "apikey", "key", "sig", "signature", "code", "client_secret"}
}

func standardJSON() []string {
	return []string{"password", "passwd", "secret", "client_secret", "access_token", "refresh_token", "api_key"}
}

// schluessel sammelt jeden Schlüssel eines JSON-Werts bis in die Tiefe.
func schluessel(v any, out *[]string) {
	switch x := v.(type) {
	case map[string]any:
		for k, w := range x {
			*out = append(*out, k)
			schluessel(w, out)
		}
	case []any:
		for _, w := range x {
			schluessel(w, out)
		}
	}
}

// keinStandardname prüft, dass keiner der Namen in den Headern und, mit felder, in den Feldern der Antwort vorkommt und dass kein ETag besteht.
func keinStandardname(t *testing.T, ziel string, header http.Header, body []byte, felder bool) {
	t.Helper()
	for _, n := range append([]string{"ETag"}, standardHeader()...) {
		if len(header.Values(n)) > 0 {
			t.Errorf("%s: Header %s", ziel, n)
		}
	}
	if !felder {
		return
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("%s: %v", ziel, err)
	}
	var namen []string
	schluessel(v, &namen)
	for _, f := range namen {
		for _, n := range standardJSON() {
			if strings.EqualFold(f, n) {
				t.Errorf("%s: Feld %s", ziel, f)
			}
		}
	}
}

// TestReferenzdienst: 5 Sets zu 2000 Entities, Gruppe gleich Id modulo 20, Nutzlast zu 7000 Entities 200 B, zu 2000 8 KiB und zu 1000 60 KiB
// (Zeichen a-z), jede Entity höchstens 64 KiB, Seiten zu 50 mit absolutem `@odata.nextLink` bis zur letzten, die keinen trägt, und kein Name aus
// den Defaults von REDACT_* und kein ETag (SPEC-216 Nr. 6 b).
func TestReferenzdienst(t *testing.T) {
	wache(t, 180*time.Second)
	if len(standardHeader()) != 7 || len(standardQuery()) != 9 || len(standardJSON()) != 7 {
		t.Fatalf("Defaults: %d, %d, %d Namen, erwartet 7, 9 und 7", len(standardHeader()), len(standardQuery()), len(standardJSON()))
	}
	r := referenz{p: abnahme(), startwert: 1, basis: testBasis}
	laengen := [10]int{200, 200, 200, 200, 200, 200, 200, 8192, 8192, 61440}
	klassen := map[int]int{}
	var gruppeJeSet [5][gruppen]int
	entities := 0
	for set := 1; set <= 5; set++ {
		for id := 1; id <= 2000; id++ {
			ziel := fmt.Sprintf("/odata/Set%d(%d)", set, id)
			status, header, body := bediene(r, http.MethodGet, ziel)
			var e struct {
				Context  string `json:"@odata.context"`
				ID       int    `json:"Id"`
				Name     string
				Gruppe   int
				Nutzlast string
			}
			dec := json.NewDecoder(bytes.NewReader(body))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&e); err != nil || status != 200 {
				t.Fatalf("%s: Status %d, %v", ziel, status, err)
			}
			if e.Context != fmt.Sprintf("%s/odata/$metadata#Set%d/$entity", testBasis, set) || e.ID != id || e.Name != fmt.Sprintf("Set%d-%d", set, id) {
				t.Fatalf("%s: Kontext %q, Id %d, Name %q", ziel, e.Context, e.ID, e.Name)
			}
			if len(body) > 65536 || len(e.Nutzlast) != laengen[id%10] || e.Gruppe != id%20 || strings.Trim(e.Nutzlast, "abcdefghijklmnopqrstuvwxyz") != "" {
				t.Fatalf("%s: Body %d Bytes, Nutzlast %d Zeichen, Gruppe %d, erwartet höchstens 65536, %d, %d und nur a-z",
					ziel, len(body), len(e.Nutzlast), e.Gruppe, laengen[id%10], id%20)
			}
			keinStandardname(t, ziel, header, body, id == 1)
			entities++
			klassen[len(e.Nutzlast)]++
			gruppeJeSet[set-1][e.Gruppe]++
		}
	}
	if entities != 10000 || fmt.Sprint(klassen) != "map[200:7000 8192:2000 61440:1000]" {
		t.Errorf("Entities %d, Klassen %v, erwartet 10000 und 7000, 2000, 1000", entities, klassen)
	}
	zellen := 0
	for set, je := range gruppeJeSet {
		for g, n := range je {
			zellen++
			if n != 100 {
				t.Errorf("Set%d, Gruppe %d: %d Entities, erwartet 100", set+1, g, n)
			}
		}
	}
	if zellen != 100 {
		t.Errorf("%d Zellen, erwartet 5 Sets mal 20 Gruppen", zellen)
	}
}

// TestReferenzSeiten: je Set 40 Seiten zu 50 Entities mit den Ids der Folge, jede außer der letzten mit dem `@odata.nextLink` der nächsten (als
// Literal gebildet), die letzte ohne; 200 Seiten, die Links tragen keinen Namen aus den Defaults der Parameter.
func TestReferenzSeiten(t *testing.T) {
	wache(t, 180*time.Second)
	r := referenz{p: abnahme(), startwert: 1, basis: testBasis}
	seiten := 0
	for set := 1; set <= 5; set++ {
		ziel := fmt.Sprintf("/odata/Set%d", set)
		for j := 0; ; j++ {
			status, header, body := bediene(r, http.MethodGet, ziel)
			var s struct {
				Context string `json:"@odata.context"`
				Weiter  string `json:"@odata.nextLink"`
				Value   []struct {
					ID int `json:"Id"`
				}
			}
			if err := json.Unmarshal(body, &s); err != nil || status != 200 || j > 39 {
				t.Fatalf("%s: Status %d, Seite %d, %v", ziel, status, j, err)
			}
			keinStandardname(t, ziel, header, body, j == 0)
			weiter := ""
			if j < 39 {
				weiter = fmt.Sprintf("%s/odata/Set%d?$skiptoken=%d", testBasis, set, (j+1)*50)
			}
			if s.Context != fmt.Sprintf("%s/odata/$metadata#Set%d", testBasis, set) || s.Weiter != weiter || len(s.Value) != 50 || s.Value[0].ID != j*50+1 || s.Value[49].ID != j*50+50 {
				t.Fatalf("%s: Kontext %q, nextLink %q (erwartet %q), %d Entities, Ids %v", ziel, s.Context, s.Weiter, weiter, len(s.Value), s.Value[:1])
			}
			seiten++
			if s.Weiter == "" {
				if j != 39 {
					t.Fatalf("Set%d: %d Seiten, erwartet 40", set, j+1)
				}
				break
			}
			u, err := url.Parse(s.Weiter)
			if err != nil {
				t.Fatal(err)
			}
			for _, n := range standardQuery() {
				if _, ok := u.Query()[n]; ok {
					t.Errorf("%s: Parameter %s", s.Weiter, n)
				}
			}
			ziel = u.RequestURI()
		}
	}
	if seiten != 200 {
		t.Errorf("%d Seiten, erwartet 200", seiten)
	}
}

// TestReferenzMetadata: `$metadata` ist ein CSDL-Dokument mit genau einem EntityType Eintrag (Id Edm.Int32 und Schlüssel, Name Edm.String, Gruppe
// Edm.Int32, Nutzlast Edm.String) und genau den EntitySets Set1 bis Set5 (Orakel: die Elemente als Literale; das Lesen des Dokuments prüft der
// Import in TestAufbau); csdlHash ist der SHA-256 der gelieferten Bytes.
func TestReferenzMetadata(t *testing.T) {
	wache(t, 30*time.Second)
	r := referenz{p: abnahme(), startwert: 1, basis: testBasis}
	status, header, body := bediene(r, http.MethodGet, "/odata/$metadata")
	if status != 200 || header.Get("Content-Type") != "application/xml" {
		t.Fatalf("Status %d, Typ %q", status, header.Get("Content-Type"))
	}
	for _, e := range []struct {
		text string
		n    int
	}{
		{`<EntityType Name="Eintrag">`, 1}, {`<Key><PropertyRef Name="Id"/></Key>`, 1},
		{`<Property Name="Id" Type="Edm.Int32" Nullable="false"/>`, 1}, {`<Property Name="Name" Type="Edm.String"/>`, 1},
		{`<Property Name="Gruppe" Type="Edm.Int32"/>`, 1}, {`<Property Name="Nutzlast" Type="Edm.String"/>`, 1},
		{`<EntitySet Name="Set1" EntityType="Lasttest.Eintrag"/>`, 1}, {`<EntitySet Name="Set2" EntityType="Lasttest.Eintrag"/>`, 1},
		{`<EntitySet Name="Set3" EntityType="Lasttest.Eintrag"/>`, 1}, {`<EntitySet Name="Set4" EntityType="Lasttest.Eintrag"/>`, 1},
		{`<EntitySet Name="Set5" EntityType="Lasttest.Eintrag"/>`, 1}, {"<EntityType ", 1}, {"<Property ", 4}, {"<EntitySet ", 5},
	} {
		if got := strings.Count(string(body), e.text); got != e.n {
			t.Errorf("%q kommt %d-mal vor, erwartet %d", e.text, got, e.n)
		}
	}
	s := sha256.Sum256(body)
	if r.csdlHash() != "sha256:"+hex.EncodeToString(s[:]) {
		t.Errorf("csdlHash %s", r.csdlHash())
	}
}

// TestReferenzUnbekannt: jede Form außerhalb der Zusage ist 404 mit Body (andere Methode, fremde Wurzel, unbekanntes Set, Id außerhalb, Schreibweise
// mit führender Null, fehlende Klammer, unbekannte oder unpassende Query, Skiptoken außerhalb der Seitenfolge); die Grenzen der gültigen Werte
// antworten mit 200.
func TestReferenzUnbekannt(t *testing.T) {
	wache(t, 30*time.Second)
	r := referenz{p: rauch(), startwert: 1, basis: testBasis}
	unbekannt := []struct{ methode, ziel string }{
		{"POST", "/odata/Set1"}, {"GET", "/other/Set1"}, {"GET", "/odata/Foo"}, {"GET", "/odata/Set0"}, {"GET", "/odata/Set6"},
		{"GET", "/odata/Set01"}, {"GET", "/odata/Seta"}, {"GET", "/odata/Set1(0)"}, {"GET", "/odata/Set1(61)"}, {"GET", "/odata/Set1(01)"},
		{"GET", "/odata/Set1(5"}, {"GET", "/odata/Set1(5)?x=1"}, {"GET", "/odata/Set1?x=1"}, {"GET", "/odata/Set1?$skiptoken=0"},
		{"GET", "/odata/Set1?$skiptoken=10"}, {"GET", "/odata/Set1?$skiptoken=60"}, {"GET", "/odata/Set1?$skiptoken=100"},
		{"GET", "/odata/$metadata?x=1"}, {"GET", "/odata/$metadata/x"},
	}
	for _, u := range unbekannt {
		status, header, body := bediene(r, u.methode, u.ziel)
		if status != 404 || header.Get("Content-Type") != "application/json" || !strings.Contains(string(body), `"error"`) {
			t.Errorf("%s %s: Status %d, Typ %q, Body %q", u.methode, u.ziel, status, header.Get("Content-Type"), body)
		}
	}
	for _, ziel := range []string{"/odata/Set1", "/odata/Set5", "/odata/Set1(1)", "/odata/Set5(60)", "/odata/Set1?$skiptoken=50", "/odata/$metadata"} {
		if status, _, _ := bediene(r, http.MethodGet, ziel); status != 200 {
			t.Errorf("GET %s: Status %d, erwartet 200", ziel, status)
		}
	}
}

// TestStarteReferenz: der Server antwortet über HTTP auf den Adressen seiner Links und lässt sich beenden; umhuellen ersetzt den Handler.
func TestStarteReferenz(t *testing.T) {
	wache(t, 30*time.Second)
	var gesehen atomic.Value
	if _, _, err := starteReferenz(t.Context(), "127.0.0.1:99999", kleinesProfil(), 1, nil); err == nil {
		t.Error("eine Adresse außerhalb der Ports bindet")
	}
	r, schliesse, err := starteReferenz(t.Context(), "127.0.0.1:0", kleinesProfil(), 1, aendere(func(w http.ResponseWriter, req *http.Request, inner http.Handler) {
		gesehen.Store(req.URL.Path)
		inner.ServeHTTP(w, req)
	}))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(r.basis + "/odata/Set1?$skiptoken=10")
	if err != nil || resp.StatusCode != 200 || gesehen.Load() != "/odata/Set1" {
		t.Fatalf("Antwort %v, %v, Pfad %v", resp, err, gesehen.Load())
	}
	_ = resp.Body.Close()
	schliesse()
	if resp, err := http.Get(r.basis + "/odata/Set1"); err == nil {
		_ = resp.Body.Close()
		t.Error("der Server antwortet nach dem Beenden")
	}
}
