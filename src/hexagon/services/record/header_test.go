package record_test

import (
	"context"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

// TestWeitergeleitet: ohne X-Recorder-*, Hop-by-Hop, Namen aus Connection, Expect, Host,
// Content-Length; alle übrigen mit allen Werten (B9).
func TestWeitergeleitet(t *testing.T) {
	fest := []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
		"TE", "Trailer", "Transfer-Encoding", "Upgrade", "Expect", "Host", "Content-Length",
		"X-Recorder-Workspace", "x-recorder-foo"}
	for _, n := range fest {
		for _, name := range []string{n, strings.ToLower(n)} {
			out := record.Weiterzuleiten(map[string][]string{name: {"v"}, "X-Bleibt": {"1"}})
			if _, da := out[name]; da || len(out) != 1 {
				t.Errorf("%s bleibt: %v", name, out)
			}
		}
	}
	bleiben := map[string][]string{
		"Authorization": {"S3CR3T"}, "Cookie": {"a=1"}, "Accept": {"a", "b"}, "X-Leer": {""},
		"Content-Type": {"application/json"}, "Content-Encoding": {"gzip"}, "X-Recorderx": {"1"}, "": {"leer"},
	}
	ein := maps.Clone(bleiben)
	ein["Connection"] = []string{"X-A,\tx-b ,,"}
	ein["X-A"], ein["x-B"] = []string{"1"}, []string{"2"}
	if out := record.Weiterzuleiten(ein); !reflect.DeepEqual(out, bleiben) {
		t.Errorf("weitergeleitet %v, erwartet %v", out, bleiben)
	}
	// Connection in kleiner Schreibweise nennt Header; nennt sie Accept-Encoding, entfällt er.
	for _, h := range []map[string][]string{
		{"connection": {"X-C"}, "X-C": {"1"}, "X-Bleibt": {"1"}},
		{"Connection": {"accept-encoding"}, "Accept-Encoding": {"gzip"}, "X-Bleibt": {"1"}},
	} {
		if out := record.Weiterzuleiten(h); !reflect.DeepEqual(out, map[string][]string{"X-Bleibt": {"1"}}) {
			t.Errorf("%v: weitergeleitet %v", h, out)
		}
	}

	u := neu(t)
	a := anfrage("DELETE", "/a/Products(1)")
	a.RawQuery = "token=S3CR3T&a=1"
	a.Header = maps.Clone(ein)
	u.rundlauf(a)
	g := u.upstream.gesehen()[0].a
	if g.Ziel != "http://a.example" || g.Methode != "DELETE" || g.Pfad != "/Products(1)" || g.RawQuery != "token=S3CR3T&a=1" {
		t.Errorf("am Upstream: %+v", g)
	}
	if !reflect.DeepEqual(g.Header, bleiben) {
		t.Errorf("Header am Upstream %v", g.Header)
	}
}

// TestAcceptEncoding: ein Fall je Glied; entfällt ein Glied mit nichtleerem Coding, entfallen die
// Glieder identity; keines ergibt identity, fehlend bleibt fehlend; Exchange und Key tragen den
// Wert des Clients (B10).
func TestAcceptEncoding(t *testing.T) {
	ae := func(werte ...string) map[string][]string { return map[string][]string{"Accept-Encoding": werte} }
	faelle := []struct {
		werte map[string][]string
		soll  string
		da    bool
	}{
		{ae("gzip"), "gzip", true},
		{ae("deflate"), "deflate", true},
		{ae("identity"), "identity", true},
		{ae("GZIP"), "GZIP", true},
		{ae("gzip;q=0.5"), "gzip;q=0.5", true},
		{ae(" \tgzip ; q=1\t "), "gzip ; q=1", true},
		{ae("gzip\t;q=1"), "gzip\t;q=1", true},
		{ae("br, gzip"), "gzip", true},
		{ae("gzip, br, deflate;q=0.1"), "gzip, deflate;q=0.1", true},
		{ae("br", "gzip"), "gzip", true},
		{map[string][]string{"accept-encoding": {"br, deflate"}, "Accept-Encoding": {"gzip"}}, "gzip, deflate", true},
		{map[string][]string{"Accept-Encoding": {}}, "", false},
		{map[string][]string{}, "", false},
		// Glieder identity: sie entfallen, wenn ein Glied mit nichtleerem Coding entfällt.
		{ae("gzip, identity;q=0"), "gzip, identity;q=0", true},
		{ae("gzip, IDENTITY"), "gzip, IDENTITY", true},
		{ae("br, identity;q=0"), "identity", true},
		{ae("identity;q=0, br"), "identity", true},
		{ae("br, gzip, Identity;q=0.5"), "gzip", true},
		{ae("gzip, ;q=1, identity;q=0"), "gzip, identity;q=0", true},
		{ae("gzip;q=0, br"), "gzip;q=0", true},
		// Jedes entfallende Glied mit nichtleerem Coding löst das aus, auch *, ein unbekanntes Coding
		// und eines mit q=0, über alle Werte aller Schreibweisen des Namens.
		{ae("zstd, identity;q=0"), "identity", true},
		{ae("compress, identity;q=0"), "identity", true},
		{ae("x-gzip, gzip, identity;q=0"), "gzip", true},
		{ae("*, identity;q=0"), "identity", true},
		{ae("x-foo, identity;q=0"), "identity", true},
		{ae("sdch, IDENTITY"), "identity", true},
		{ae("br;q=0, identity;q=0"), "identity", true},
		{ae("br", "identity;q=0"), "identity", true},
		{map[string][]string{"Accept-Encoding": {"identity;q=0"}, "accept-encoding": {"x-foo"}}, "identity", true},
	}
	for _, allein := range []string{"br", "zstd", "compress", "x-gzip", "*", "", ",,"} {
		faelle = append(faelle, struct {
			werte map[string][]string
			soll  string
			da    bool
		}{ae(allein), "identity", true})
	}
	for _, f := range faelle {
		got, da := record.AcceptEncoding(f.werte)
		if got != f.soll || da != f.da {
			t.Errorf("%v: %q %v, erwartet %q %v", f.werte, got, da, f.soll, f.da)
		}
	}

	u := neu(t)
	a := anfrage("GET", "/a/x")
	a.Header["Accept-Encoding"] = []string{"br, gzip"}
	u.rundlauf(a)
	if g := u.upstream.gesehen()[0].a.Header["Accept-Encoding"]; !slices.Equal(g, []string{"gzip"}) {
		t.Errorf("Upstream sieht %v", g)
	}
	e := u.letzter()
	if g := e.RequestHeaders["Accept-Encoding"]; !slices.Equal(g, []string{"br, gzip"}) {
		t.Errorf("Exchange trägt %v", g)
	}
	k, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "svc-a", Path: "/x",
		Header: map[string][]string{"Accept-Encoding": {"br, gzip"}}, RequestRedaction: model.RedactionNone}, regeln())
	if err != nil || e.MatchKey != k.MatchKey {
		t.Errorf("Key %q, erwartet %q (%v)", e.MatchKey, k.MatchKey, err)
	}
}

// TestAntwortAnAdapter: Status je Wert und Header ohne Hop-by-Hop, unredigiert; Bytes und
// Lesefehler unverändert, auch bei Capture-Fehler und bei leerem p; gespeichert redigiert (B13).
func TestAntwortAnAdapter(t *testing.T) {
	u := neu(t)
	h := map[string][]string{"Set-Cookie": {"sid=S3CR3T"}, "Connection": {"X-Z"}, "X-Z": {"1"},
		"Keep-Alive": {"t"}, "Content-Type": {"text/plain"}}
	u.upstream.plane(antwort(200, h, koerper("abc")))
	ant, gelesen := u.rundlauf(anfrage("GET", "/a/x"))
	soll := map[string][]string{"Set-Cookie": {"sid=S3CR3T"}, "Content-Type": {"text/plain"}}
	if ant.Status != 200 || !reflect.DeepEqual(ant.Header, soll) || gelesen != "abc" {
		t.Errorf("Antwort %d %v %q", ant.Status, ant.Header, gelesen)
	}
	e := u.letzter()
	if gespeichert := (map[string][]string{"Set-Cookie": {model.Marker}, "Content-Type": {"text/plain"}}); !reflect.DeepEqual(e.ResponseHeaders, gespeichert) {
		t.Errorf("gespeichert %v", e.ResponseHeaders)
	}
	if e.Capture != model.CaptureFailed {
		t.Errorf("Capture %s, erwartet failed (typ)", e.Capture)
	}

	v := neu(t)
	v.upstream.plane(antwort(200, map[string][]string{"Content-Type": {"application/json"}},
		koerper(`{"a":`).dann(nil, "", errVerbindung)))
	ant, err := v.record(context.Background(), anfrage("GET", "/a/x"))
	if err != nil {
		t.Fatal(err)
	}
	p := make([]byte, 8)
	if n, err := ant.Body.Read(p); n != 5 || err != nil {
		t.Fatalf("erstes Read %d %v", n, err)
	}
	if n, err := ant.Body.Read(p); n != 0 || err != errVerbindung {
		t.Errorf("Lesefehler %d %v, erwartet errVerbindung unverändert", n, err)
	}
	schliesse(t, ant.Body)
	if e := v.letzter(); e.Capture != model.CaptureIncomplete {
		t.Errorf("Capture %s nach Lesefehler", e.Capture)
	}

	for _, status := range []int{201, 304, 404, 502} {
		s := neu(t)
		s.upstream.plane(antwort(status, map[string][]string{"Content-Type": {"application/json"}}, koerper()))
		ant, _ := s.rundlauf(anfrage("GET", "/a/x"))
		if e := s.letzter(); ant.Status != status || e.ResponseStatus != status {
			t.Errorf("Status %d: an den Adapter %d, im Exchange %d", status, ant.Status, e.ResponseStatus)
		}
	}

	// Ein Read mit leerem p geht an den Body des Upstreams: dessen Fehler kommt zurück (SPEC-119 Nr. 3).
	for _, p := range [][]byte{nil, {}} {
		w := neu(t)
		w.upstream.plane(jsonAntwort((&fakeBody{}).dann(nil, "", errVerbindung)))
		ant, err := w.record(context.Background(), anfrage("GET", "/a/x"))
		if err != nil {
			t.Fatal(err)
		}
		if n, err := ant.Body.Read(p); n != 0 || err != errVerbindung {
			t.Errorf("Read mit p %v: %d %v, erwartet errVerbindung vom Upstream", p == nil, n, err)
		}
		schliesse(t, ant.Body)
	}
}
