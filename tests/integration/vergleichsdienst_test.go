package integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// vergleichsdienst ist der Referenzdienst des Vergleichs der Protokolltreue (LH-QA-07.a): ein OData-V4-Dienst
// als httptest-Server, dessen Antwort allein vom Request abhängt, nie von einem Zähler oder der Reihenfolge
// (Nr. 6). Er trägt je Antwortart eine Route; jede Antwort nennt den empfangenen Accept-Encoding im Header
// X-Echo-Accept-Encoding, und eine Antwort mit Body folgt dem Accept-Encoding mit gzip, deflate oder beiden.
type vergleichsdienst struct {
	srv *httptest.Server
}

// dienstantwort ist, was die Route bestimmt: Status, Header (vor dem Echo) und der Body vor dem Coding.
type dienstantwort struct {
	status int
	header http.Header
	body   []byte
}

func neuerVergleichsdienst(t *testing.T) *vergleichsdienst {
	t.Helper()
	d := &vergleichsdienst{}
	d.srv = httptest.NewServer(http.HandlerFunc(d.bediene))
	t.Cleanup(d.srv.Close)
	return d
}

// brotliBytes sind feste Bytes unter Content-Encoding br; sie sind kein gültiger Strom, kein Test entpackt sie.
func brotliBytes() []byte { return []byte{0x0b, 0x02, 0x80, 'A', 'n', 'v', 'i', 'l', 0x03, 0xff} }

func (d *vergleichsdienst) bediene(w http.ResponseWriter, req *http.Request) {
	a := d.route(req)
	h := w.Header()
	for n, v := range a.header {
		h[n] = v
	}
	if v, da := req.Header["Accept-Encoding"]; da {
		h.Set("X-Echo-Accept-Encoding", strings.Join(v, "|"))
	} else {
		h.Set("X-Echo-Accept-Encoding", "fehlt")
	}
	body := a.body
	if g := gewaehlteKette(req); len(g) > 0 && len(body) > 0 && h.Get("Content-Encoding") == "" {
		h.Set("Content-Encoding", strings.Join(g, ", "))
		body = kodiere(g, body)
	}
	w.WriteHeader(a.status)
	_, _ = w.Write(body)
}

// gewaehlteKette ist das Coding, das der Dienst nach dem Accept-Encoding wählt: beide genannt ergeben
// `deflate, gzip`, sonst das eine genannte, sonst keines; Parameter wertet er nicht aus.
func gewaehlteKette(req *http.Request) []string {
	genannt := map[string]bool{}
	for _, v := range req.Header["Accept-Encoding"] {
		for _, t := range teileKomma(v) {
			genannt[asciiKlein(strings.SplitN(t, ";", 2)[0])] = true
		}
	}
	switch {
	case genannt["gzip"] && genannt["deflate"]:
		return []string{"deflate", "gzip"}
	case genannt["gzip"]:
		return []string{"gzip"}
	case genannt["deflate"]:
		return []string{"deflate"}
	}
	return nil
}

// route bestimmt die Antwort auf einen Request: eine Route je Antwortart des Vergleichs.
func (d *vergleichsdienst) route(req *http.Request) dienstantwort {
	h := http.Header{}
	h.Set("OData-Version", "4.0")
	h.Set("X-Dienst", "ref")
	odataJSON := "application/json;odata.metadata=minimal"
	basis := d.srv.URL + "/odata"
	switch req.URL.Path {
	case "/odata/$metadata":
		h.Set("Content-Type", "application/xml")
		return dienstantwort{200, h, []byte(csdlText)}
	case "/odata/Products":
		h.Set("Content-Type", odataJSON)
		return dienstantwort{200, h, []byte(fmt.Sprintf(`{"@odata.context":%q,"value":[{"Id":1,"Name":"Anvil"},{"Id":2,"Name":"Bolt"}]}`, basis+"/$metadata#Products"))}
	case "/odata/Products(1)":
		h.Set("Content-Type", odataJSON)
		h.Set("ETag", `W/"e1"`)
		return dienstantwort{200, h, []byte(fmt.Sprintf(`{"@odata.context":%q,"Id":1,"Name":"Anvil"}`, basis+"/$metadata#Products/$entity"))}
	case "/odata/Products/$count":
		h.Set("Content-Type", "text/plain")
		return dienstantwort{200, h, []byte("7")}
	case "/odata/Products(1)/Name/$value":
		h.Set("Content-Type", "text/plain")
		return dienstantwort{200, h, []byte("Anvil")}
	case "/odata/Leer":
		return dienstantwort{204, h, nil}
	case "/odata/Unveraendert":
		h.Set("ETag", `W/"e1"`)
		return dienstantwort{304, h, nil}
	case "/odata/Weiter":
		h.Set("Location", basis+"/Products")
		return dienstantwort{302, h, nil}
	case "/odata/Header":
		h.Set("Content-Type", odataJSON)
		h["X-Mehrfach"] = []string{"a", "b"}
		h["X-Leer"] = []string{""}
		for n, v := range map[string]string{"Server": "ref", "Via": "1.1 ref", "Age": "0", "X-Request-Id": "r-1",
			"X-Correlation-Id": "c-1", "Traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"} {
			h.Set(n, v)
		}
		return dienstantwort{200, h, []byte(`{"value":[]}`)}
	case "/odata/Brotli":
		h.Set("Content-Type", odataJSON)
		h.Set("Content-Encoding", "br")
		return dienstantwort{200, h, brotliBytes()}
	}
	h.Set("Content-Type", "application/json")
	return dienstantwort{404, h, []byte(`{"error":{"code":"404","message":"nicht gefunden"}}`)}
}
