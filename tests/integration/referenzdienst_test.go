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
	"slices"
	"strings"
	"sync"
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
	// methoden zählt die Requests je Methode, wie der Request sie trägt (byte-gleich); der Wert je Methode ist ein *atomic.Int64.
	methoden sync.Map
	// gesehen hält Methode, Pfad und Body jedes Requests in der Reihenfolge seines Eintreffens.
	mu      sync.Mutex
	gesehen []gesehenerRequest
	// jedeVerbindungNeu lässt den Dienst nach jeder Antwort die Verbindung schließen: Verbindungen und Requests sind dann gleich.
	jedeVerbindungNeu atomic.Bool
}

// gesehenerRequest ist ein Request, den der Referenzdienst bediente.
type gesehenerRequest struct {
	Methode, Pfad, Query string
	Body                 []byte
}

// nachMethode ist die Zahl der Requests mit genau dieser Methode.
func (r *referenzdienst) nachMethode(methode string) int64 {
	if z, ok := r.methoden.Load(methode); ok {
		return z.(*atomic.Int64).Load()
	}
	return 0
}

// zaehleMethode zählt einen Request der Methode.
func (r *referenzdienst) zaehleMethode(methode string) {
	z, _ := r.methoden.LoadOrStore(methode, new(atomic.Int64))
	z.(*atomic.Int64).Add(1)
}

// zaehleRequest zählt einen Request, zaehleVerbindung eine neue Verbindung; beide sind die einzigen Stellen, die die Zähler erhöhen.
func (r *referenzdienst) zaehleRequest()    { r.requests.Add(1) }
func (r *referenzdienst) zaehleVerbindung() { r.verbindungen.Add(1) }

// merke hält einen Request fest; es ist die einzige Stelle, die gesehen schreibt.
func (r *referenzdienst) merke(g gesehenerRequest) {
	r.mu.Lock()
	r.gesehen = append(r.gesehen, g)
	r.mu.Unlock()
}

// gesehenSeit liefert die Requests ab dem n-ten.
func (r *referenzdienst) gesehenSeit(n int) []gesehenerRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.gesehen[n:])
}

func neuerReferenzdienst(t *testing.T) *referenzdienst {
	t.Helper()
	r := &referenzdienst{}
	r.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.zaehleRequest()
		r.zaehleMethode(req.Method)
		b, _ := io.ReadAll(req.Body)
		r.merke(gesehenerRequest{Methode: req.Method, Pfad: req.URL.Path, Query: req.URL.RawQuery, Body: b})
		if r.jedeVerbindungNeu.Load() {
			w.Header().Set("Connection", "close")
		}
		if req.Header.Get("Authorization") == "Bearer "+geheimA {
			r.mitGeheimnis.Add(1)
		}
		r.bediene(w, req)
	}))
	r.srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			r.zaehleVerbindung()
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
	// Ein POST auf die Collection ist 201 mit der angelegten Entity (SPEC-300 Nr. 14: die Session A trägt einen).
	case req.URL.Path == "/odata/Products" && req.Method == http.MethodPost:
		h.Set("Content-Type", "application/json;odata.metadata=minimal")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"@odata.context":%q,"Id":3,"Name":"Cutter"}`, r.srv.URL+"/odata/$metadata#Products/$entity")
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

// TestReferenzdienstZaehlerGleichzeitig fängt die Zusage „Zähler atomar“ (M11, M19) an der Wirkung, in jedem Lauf: 64 Goroutinen rufen nach einem
// gemeinsamen Start je 2000-mal die drei Zählstellen (Request, Verbindung, Methode) und die Stelle `merke` der Liste der Requests; ein Lost Update
// oder ein Datenrennen auf einer der vier Stellen verfehlt die Summe, ohne dass es auf ein Fenster zwischen zwei Requests ankommt. Die Methode der Goroutine g ist die (g mod 10)-te.
func TestReferenzdienstZaehlerGleichzeitig(t *testing.T) {
	const goroutinen, runden = 64, 2000
	methoden := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "CONNECT"}
	r := &referenzdienst{}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range runden {
				r.zaehleRequest()
				r.zaehleVerbindung()
				r.zaehleMethode(methoden[g%len(methoden)])
				r.merke(gesehenerRequest{Methode: methoden[g%len(methoden)], Pfad: "/odata/Nichts"})
			}
		}()
	}
	close(start)
	wg.Wait()
	const alle = int64(goroutinen * runden)
	if r.requests.Load() != alle || r.verbindungen.Load() != alle {
		t.Errorf("Requests %d, Verbindungen %d; erwartet je %d", r.requests.Load(), r.verbindungen.Load(), alle)
	}
	if got := r.gesehenSeit(0); int64(len(got)) != alle {
		t.Errorf("%d festgehaltene Requests, erwartet %d", len(got), alle)
	}
	summe := int64(0)
	for i, m := range methoden {
		soll := int64(goroutinen / len(methoden) * runden)
		if i < goroutinen%len(methoden) {
			soll += runden
		}
		if got := r.nachMethode(m); got != soll {
			t.Errorf("Zähler von %s: %d, erwartet %d", m, got, soll)
		}
		summe += r.nachMethode(m)
	}
	if summe != alle {
		t.Errorf("Summe je Methode %d, erwartet %d", summe, alle)
	}
}

// TestReferenzdienstNachMethode fängt M11 (SPEC-037 Nr. 4): zehn Methoden, die Methode i wird (i+1)-mal gesendet; der Zähler je Methode, die Summe
// der Requests und die Verbindungen stimmen mit dem Orakel (der Zahl der gesendeten Requests) überein, einmal mit neuer Verbindung je Request
// (Verbindungen gleich Requests), einmal mit Keep-Alive (weniger Verbindungen als Requests, mindestens eine); die Schreibweise der Methode zählt,
// und eine unbekannte Methode hat ihren eigenen Zähler.
func TestReferenzdienstNachMethode(t *testing.T) {
	methoden := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "CONNECT"}
	for _, keepAlive := range []bool{false, true} {
		t.Run(fmt.Sprintf("keepAlive=%v", keepAlive), func(t *testing.T) {
			r := neuerReferenzdienst(t)
			tr := &http.Transport{DisableKeepAlives: !keepAlive}
			t.Cleanup(tr.CloseIdleConnections)
			c := &http.Client{Transport: tr}
			summe := int64(0)
			for i, m := range append(slices.Clone(methoden), "get") {
				for range i + 1 {
					req, err := http.NewRequest(m, r.URL()+"/odata/Nichts", nil)
					if err != nil {
						t.Fatal(err)
					}
					resp, err := c.Do(req)
					if err != nil {
						t.Fatalf("%s: %v", m, err)
					}
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					summe++
				}
			}
			for i, m := range methoden {
				if got := r.nachMethode(m); got != int64(i+1) {
					t.Errorf("Zähler von %s: %d, erwartet %d", m, got, i+1)
				}
			}
			if got := r.nachMethode("get"); got != int64(len(methoden)+1) || r.nachMethode("Get") != 0 {
				t.Errorf("Zähler von get: %d, von Get: %d; die Schreibweise zählt", got, r.nachMethode("Get"))
			}
			if got := r.requests.Load(); got != summe {
				t.Errorf("Requests %d, gesendet %d", got, summe)
			}
			v := r.verbindungen.Load()
			if keepAlive && (v < 1 || v >= summe) || !keepAlive && v != summe {
				t.Errorf("Verbindungen %d bei %d Requests mit keepAlive=%v", v, summe, keepAlive)
			}
			if got := r.gesehenSeit(0); int64(len(got)) != summe {
				t.Errorf("%d festgehaltene Requests, gesendet %d", len(got), summe)
			}
		})
	}
	// Wirkungs-Fänger zu M19: 64 gleichzeitige Requests, der Start über einen Kanal; die Zähler summieren sich auf die Zahl der Requests.
	t.Run("gleichzeitig", func(t *testing.T) {
		const anzahl = 64
		r := neuerReferenzdienst(t)
		r.jedeVerbindungNeu.Store(true)
		tr := &http.Transport{DisableKeepAlives: true}
		t.Cleanup(tr.CloseIdleConnections)
		c := &http.Client{Transport: tr}
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := range anzahl {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				req, err := http.NewRequest(methoden[i%len(methoden)], r.URL()+"/odata/Nichts", nil)
				if err != nil {
					t.Error(err)
					return
				}
				resp, err := c.Do(req)
				if err != nil {
					t.Error(err)
					return
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}()
		}
		close(start)
		wg.Wait()
		summe := int64(0)
		for i, m := range methoden {
			soll := int64(anzahl / len(methoden))
			if i < anzahl%len(methoden) {
				soll++
			}
			if got := r.nachMethode(m); got != soll {
				t.Errorf("Zähler von %s: %d, erwartet %d", m, got, soll)
			}
			summe += r.nachMethode(m)
		}
		if r.requests.Load() != anzahl || r.verbindungen.Load() != anzahl || summe != anzahl || len(r.gesehenSeit(0)) != anzahl {
			t.Errorf("Requests %d, Verbindungen %d, Summe je Methode %d, festgehalten %d; erwartet je %d",
				r.requests.Load(), r.verbindungen.Load(), summe, len(r.gesehenSeit(0)), anzahl)
		}
	})
	t.Run("jedeVerbindungNeu", func(t *testing.T) {
		r := neuerReferenzdienst(t)
		r.jedeVerbindungNeu.Store(true)
		tr := &http.Transport{}
		t.Cleanup(tr.CloseIdleConnections)
		c := &http.Client{Transport: tr}
		for range 5 {
			resp, err := c.Get(r.URL() + "/odata/Nichts")
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		if r.verbindungen.Load() != 5 || r.requests.Load() != 5 {
			t.Errorf("Verbindungen %d, Requests %d; erwartet je 5", r.verbindungen.Load(), r.requests.Load())
		}
	})
}
