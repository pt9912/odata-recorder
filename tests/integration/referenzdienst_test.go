package integration_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
)

// Die drei Geheimnisse des Referenzdienstes: GEHEIM-A erreicht ihn im Request-Header Authorization,
// GEHEIM-B verlässt ihn im Antwort-Header Set-Cookie, GEHEIM-C steht als Text-Antwort auf
// /Products(1)/Password/$value (LH-FA-14).
const (
	geheimA = "GEHEIM-A-4f7c1e"
	geheimB = "GEHEIM-B-9d2a60"
	geheimC = "GEHEIM-C-5b8e33"
)

// propertyWert ist die Route `/Products(<Schlüssel>)/<Property>/$value`, propertyWerte nennt den Wert
// je `<Schlüssel>/<Property>`. `Password` trägt ein Geheimnis unter einem Namen, den ein Standardmuster
// trifft, `Name` einen Wert ohne Treffer.
var propertyWert = regexp.MustCompile(`^/odata/Products\(([0-9]+)\)/([A-Za-z]+)/\$value$`)

func propertyWerte() map[string]string {
	return map[string]string{"1/Name": "Anvil", "2/Name": "Bolt", "1/Password": geheimC}
}

// Die Werte der Fixture: die Antworten auf /Products/$count in der Reihenfolge der Requests, ungleich
// Index und Anzahl.
func zaehlerFolge() []int { return []int{7, 11, 13} }

// csdlText ist das Dokument von $metadata; der Hash, den der Import erhält, entsteht aus diesen Bytes.
const csdlText = `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="Ref" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Product">
        <Key><PropertyRef Name="Id"/></Key>
        <Property Name="Id" Type="Edm.Int32" Nullable="false"/>
        <Property Name="Name" Type="Edm.String"/>
      </EntityType>
      <EntityContainer Name="Container">
        <EntitySet Name="Products" EntityType="Ref.Product"/>
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>
`

// csdlHash ist das Orakel für MetadataHash: SHA-256 der Bytes der Fixture, nicht aus dem Code.
func csdlHash() string {
	s := sha256.Sum256([]byte(csdlText))
	return "sha256:" + hex.EncodeToString(s[:])
}

// referenzdienst ist der Upstream des Durchstichs: ein OData-V4-Dienst als httptest-Server mit zwei
// Zählern. Verbindungen zählt ConnState (Zustand StateNew), Requests der Wrapper um den Handler.
type referenzdienst struct {
	srv          *httptest.Server
	verbindungen atomic.Int64
	requests     atomic.Int64
	// mitGeheimnis zählt die Requests, deren Authorization GEHEIM-A trug.
	mitGeheimnis atomic.Int64
	// zaehlerRequests zählt die Requests auf /Products/$count und wählt damit den Wert der Antwort.
	zaehlerRequests atomic.Int64
}

func neuerReferenzdienst(t *testing.T) *referenzdienst {
	t.Helper()
	r := &referenzdienst{}
	r.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.requests.Add(1)
		if req.Header.Get("Authorization") == "Bearer "+geheimA {
			r.mitGeheimnis.Add(1)
		}
		r.bediene(w, req)
	}))
	r.srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			r.verbindungen.Add(1)
		}
	}
	r.srv.Start()
	t.Cleanup(r.srv.Close)
	return r
}

// URL ist die Adresse des Servers, URL() + "/odata" der Dienst.
func (r *referenzdienst) URL() string { return r.srv.URL }

func (r *referenzdienst) bediene(w http.ResponseWriter, req *http.Request) {
	h := w.Header()
	h.Set("OData-Version", "4.0")
	h.Set("Set-Cookie", "sid="+geheimB)
	json := func(etag, body string) {
		h.Set("Content-Type", "application/json;odata.metadata=minimal")
		if etag != "" {
			h.Set("ETag", etag)
		}
		_, _ = w.Write([]byte(body))
	}
	ctx := r.srv.URL + "/odata/$metadata#Products"
	switch {
	case req.URL.Path == "/odata/$metadata":
		h.Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(csdlText))
	case req.URL.Path == "/odata/Products" && req.URL.Query().Get("$skiptoken") == "":
		json(`W/"p1"`, fmt.Sprintf(`{"@odata.context":%q,"@odata.nextLink":%q,"value":[{"Id":1,"Name":"Anvil"}]}`,
			ctx, r.srv.URL+"/odata/Products?$skiptoken=1"))
	case req.URL.Path == "/odata/Products" && req.URL.Query().Get("$skiptoken") == "1":
		json(`W/"p2"`, fmt.Sprintf(`{"@odata.context":%q,"value":[{"Id":2,"Name":"Bolt"}]}`, ctx))
	case req.URL.Path == "/odata/Products(1)":
		json(`W/"e1"`, fmt.Sprintf(`{"@odata.context":%q,"Id":1,"Name":"Anvil"}`, r.srv.URL+"/odata/$metadata#Products/$entity"))
	case req.URL.Path == "/odata/Products/$count":
		n, folge := int(r.zaehlerRequests.Add(1))-1, zaehlerFolge()
		h.Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "%d", folge[n%len(folge)])
	case propertyWert.MatchString(req.URL.Path) && property(req.URL.Path) != "":
		h.Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(property(req.URL.Path)))
	default:
		h.Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"404","message":"nicht gefunden"}}`))
	}
}

// property liefert den Wert der Route `$value` oder "" für ein unbekanntes Paar aus Schlüssel und Property.
func property(pfad string) string {
	m := propertyWert.FindStringSubmatch(pfad)
	return propertyWerte()[m[1]+"/"+m[2]]
}

// TestReferenzdienstSelbst fängt den Referenzdienst gegen die Annahmen des M2-Tests: die Zähler zählen,
// das Geheimnis A erreicht ihn, die Folge von $count ist 7, 11, 13, eine Seite verweist absolut auf die nächste.
func TestReferenzdienstSelbst(t *testing.T) {
	r := neuerReferenzdienst(t)
	c := &http.Client{}
	hole := func(pfad string) (*http.Response, string) {
		req, err := http.NewRequest(http.MethodGet, r.URL()+pfad, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+geheimA)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return resp, string(b)
	}
	resp, body := hole("/odata/Products")
	if !strings.Contains(body, `"@odata.nextLink":"`+r.URL()+`/odata/Products?$skiptoken=1"`) {
		t.Errorf("Seite 1 ohne absoluten nextLink: %s", body)
	}
	if resp.Header.Get("Set-Cookie") != "sid="+geheimB {
		t.Errorf("Set-Cookie: %q", resp.Header.Get("Set-Cookie"))
	}
	for i, soll := range zaehlerFolge() {
		resp, b := hole("/odata/Products/$count")
		if b != fmt.Sprintf("%d", soll) || resp.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("$count Nr. %d: %q als %q", i+1, b, resp.Header.Get("Content-Type"))
		}
	}
	for _, f := range []struct{ pfad, wert string }{
		{"/odata/Products(1)/Name/$value", "Anvil"}, {"/odata/Products(2)/Name/$value", "Bolt"},
		{"/odata/Products(1)/Password/$value", geheimC},
	} {
		resp, b := hole(f.pfad)
		if resp.StatusCode != 200 || b != f.wert || resp.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("%s: %d %q als %q", f.pfad, resp.StatusCode, b, resp.Header.Get("Content-Type"))
		}
	}
	for _, pfad := range []string{"/odata/Products(3)/Name/$value", "/odata/Products(2)/Password/$value", "/odata/Products(1)/Id/$value"} {
		if resp, _ := hole(pfad); resp.StatusCode != 404 || resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("%s: %d als %q", pfad, resp.StatusCode, resp.Header.Get("Content-Type"))
		}
	}
	if got := r.requests.Load(); got != 10 {
		t.Errorf("Requests %d, erwartet 10", got)
	}
	if got := r.mitGeheimnis.Load(); got != 10 {
		t.Errorf("Requests mit Geheimnis %d, erwartet 10", got)
	}
	if r.verbindungen.Load() < 1 {
		t.Error("Verbindungen zählen nichts")
	}
}
