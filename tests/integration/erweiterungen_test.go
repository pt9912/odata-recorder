package integration_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// Der Nachweis für unbekannte OData-Erweiterungen auf dem Record-Pfad (LH-QA-07, LH-QA-07.a): je Fall nimmt der Recorder im Modus record
// über den Listener gegen einen eigenen Referenzdienst auf. Der Referenzdienst erhält Pfad, Query und die eigenen Header unverändert, die
// Antwort über den Proxy gleicht der direkten, und die Session trägt den Exchange mit capture_status=complete.

// erwFall ist ein Fall: das Ziel, wie der Client es sendet (Pfad mit Query), seine Header und die Antwort des Referenzdienstes.
type erwFall struct {
	name, ziel, accept string
	// odataPfad ist der Pfad des Exchange unter dem Root des Dienstes.
	odataPfad string
	// sendet sind die eigenen Header der Anfrage, Namen in kanonischer Schreibweise.
	sendet http.Header
	typ    string
	body   string
	// antwortet sind die eigenen Header der Antwort, Namen in kanonischer Schreibweise.
	antwortet http.Header
}

func erwFaelle() []erwFall {
	json := "application/json;odata.metadata=full"
	return []erwFall{
		{name: "Query", ziel: "/odata/Products?$apply=groupby((Name))&$search=%22blau%22&$unbekannt=1&sap-client=100", accept: acceptJSON,
			odataPfad: "/Products", typ: json, body: `{"@odata.context":"$metadata#Products(Name)","value":[{"Name":"Bolt"}]}`},
		{name: "Pfad", ziel: "/odata/Products/Ns.Beste()", accept: acceptJSON, odataPfad: "/Products/Ns.Beste()", typ: json,
			body: `{"@odata.context":"$metadata#Products/$entity","Id":2,"Name":"Bolt"}`},
		{name: "Header", ziel: "/odata/Products", accept: acceptJSON, odataPfad: "/Products",
			sendet: http.Header{"Sap-Contextid-Accept": {"header"}, "X-Vendor-Mehrfach": {"b", "a"}, "X-Vendor-Leer": {""}},
			typ:    json, body: `{"@odata.context":"$metadata#Products","value":[]}`,
			antwortet: http.Header{"Sap-Message": {`{"code":"V/001"}`}, "X-Vendor-Folge": {"z", "a"}, "X-Vendor-Ohne-Wert": {""}}},
		{name: "JSON", ziel: "/odata/Orders?$expand=Positionen", accept: acceptJSON, odataPfad: "/Orders", typ: json,
			body: `{"@odata.context":"$metadata#Orders","@com.vendor.Stand":"alt","value":[{"@com.vendor.Zeile":{"farbe":"blau"},` +
				`"Id":1,"Note@com.vendor.Format":"kurz","Note":"n","Positionen":[{"@com.vendor.Tief":true,"Nr":10}]}]}`},
		{name: "Metadaten", ziel: "/odata/$metadata", accept: acceptXML, odataPfad: "/$metadata", typ: "application/xml",
			body: `<?xml version="1.0" encoding="utf-8"?><edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" ` +
				`xmlns:v="http://vendor.example/ns"><edmx:DataServices><Schema Namespace="N" xmlns="http://docs.oasis-open.org/odata/ns/edm">` +
				`<EntityType Name="Product" v:label="Artikel"><Key><PropertyRef Name="Id"/></Key><Property Name="Id" Type="Edm.Int32" Nullable="false"/>` +
				`<v:Hinweis stufe="2">frei</v:Hinweis></EntityType><EntityContainer Name="C"><EntitySet Name="Products" EntityType="N.Product"/>` +
				`</EntityContainer><Annotations Target="N.Product"><Annotation Term="com.vendor.Farbe" String="blau"/></Annotations></Schema>` +
				`</edmx:DataServices></edmx:Edmx>`},
	}
}

// erwGesehen ist eine Anfrage, wie der Referenzdienst sie erhielt: das Ziel der Request-Zeile und die Header.
type erwGesehen struct {
	ziel   string
	header http.Header
}

// erwDienst ist der Referenzdienst eines Falls: er antwortet auf jede Anfrage gleich und hält jede fest.
type erwDienst struct {
	srv     *httptest.Server
	mu      sync.Mutex
	gesehen []erwGesehen
}

func neuerErwDienst(t *testing.T, f erwFall) *erwDienst {
	t.Helper()
	d := &erwDienst{}
	d.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		d.mu.Lock()
		d.gesehen = append(d.gesehen, erwGesehen{ziel: req.RequestURI, header: req.Header.Clone()})
		d.mu.Unlock()
		for n, v := range f.antwortet {
			w.Header()[n] = v
		}
		w.Header().Set("Content-Type", f.typ)
		w.Header().Set("OData-Version", "4.0")
		_, _ = io.WriteString(w, f.body)
	}))
	t.Cleanup(d.srv.Close)
	return d
}

func (d *erwDienst) anfragen() []erwGesehen {
	d.mu.Lock()
	defer d.mu.Unlock()
	return slices.Clone(d.gesehen)
}

// erwHole sendet die Anfrage des Falls und liest die Antwort roh; jeder Fall trägt dazu den eigenen Header X-Vendor-Fall.
func erwHole(t *testing.T, c *vclient, f erwFall) rohantwort {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.basis+f.ziel, nil)
	if err != nil {
		t.Fatal(err)
	}
	for n, v := range erwEigene(f) {
		req.Header[n] = v
	}
	req.Header.Set("Accept", f.accept)
	resp, err := c.c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Body: %v", err)
	}
	return rohantwort{status: resp.StatusCode, header: resp.Header, body: body}
}

func erwEigene(f erwFall) http.Header {
	h := f.sendet.Clone()
	if h == nil {
		h = http.Header{}
	}
	h["X-Vendor-Fall"] = []string{f.name}
	return h
}

// erwPruefeGesehen nennt die Abweichungen der erhaltenen Anfrage von der gesendeten: Pfad mit Query byte-gleich und jeder eigene Header mit
// denselben Feldzeilen.
func erwPruefeGesehen(t *testing.T, seite string, f erwFall, g erwGesehen) {
	t.Helper()
	if g.ziel != f.ziel {
		t.Errorf("%s: Referenzdienst erhielt %q, gesendet %q", seite, g.ziel, f.ziel)
	}
	for n, v := range erwEigene(f) {
		if got, da := g.header[n]; !da || !slices.Equal(got, v) {
			t.Errorf("%s: Header %s am Referenzdienst %q (vorhanden %t), gesendet %q", seite, n, got, da, v)
		}
	}
}

func erwPruefe(t *testing.T, f erwFall) {
	ctx := context.Background()
	d := neuerErwDienst(t, f)
	k := konfig(t, "record", t.TempDir(), d.srv.URL+"/odata")
	rec := hochgefahren(ctx, t, k)
	sessions, err := rec.Sessions().List(ctx)
	if err != nil || len(sessions) != 1 {
		t.Fatalf("Sessions: %+v, %v", sessions, err)
	}
	direkt := erwHole(t, neuerVClient(t, d.srv.URL), f)
	proxy := erwHole(t, neuerVClient(t, "http://"+rec.Adresse()), f)
	regeln := vergleichsregeln{volatile: k.VolatileHeaders}

	// Die direkte Antwort trägt, was der Fall prüft; erst dann sagt der Vergleich etwas.
	if direkt.status != http.StatusOK || string(direkt.body) != f.body {
		t.Fatalf("direkt: Status %d, Body %q", direkt.status, direkt.body)
	}
	for n, v := range f.antwortet {
		if !slices.Equal(direkt.header[n], v) || !slices.Contains(regeln.relevanteNamen(direkt), strings.ToLower(n)) {
			t.Fatalf("direkt: Header %s %q ist nicht relevant oder fehlt", n, direkt.header[n])
		}
	}
	abweichungen, err := vergleiche(regeln, direkt, proxy)
	if err != nil || len(abweichungen) != 0 {
		t.Errorf("Antwort über den Proxy: Abweichungen %v, %v; Status %d, Body %q", abweichungen, err, proxy.status, proxy.body)
	}
	anfragen := d.anfragen()
	if len(anfragen) != 2 {
		t.Errorf("Anfragen am Referenzdienst %d, erwartet 2 (direkt und über den Proxy)", len(anfragen))
	}
	for i, g := range anfragen {
		erwPruefeGesehen(t, []string{"direkt", "über den Proxy"}[min(i, 1)], f, g)
	}
	erwPruefeSession(t, rec, sessions[0].ID, f)
}

// erwPruefeSession beendet die Session, das wartet auf laufende Exchanges, und liest genau den einen Exchange des Falls.
func erwPruefeSession(t *testing.T, rec *bootstrap.Handle, sessionID string, f erwFall) {
	t.Helper()
	ctx := context.Background()
	if _, err := rec.Sessions().Beenden(ctx, sessionID); err != nil {
		t.Fatalf("Beenden: %v", err)
	}
	seite, err := rec.Inspektion().Exchanges(ctx, sessionID, inbound.Seite{})
	if err != nil || seite.Gesamt != 1 || len(seite.Eintraege) != 1 {
		t.Fatalf("Exchanges der Session: %+v, %v; erwartet genau einer", seite, err)
	}
	e := seite.Eintraege[0]
	if e.Capture != model.CaptureComplete || e.Method != http.MethodGet || e.Path != f.odataPfad || e.ResponseStatus != http.StatusOK {
		t.Errorf("Exchange: capture_status %q, %s %q, Status %d; erwartet complete, GET %q, 200", e.Capture, e.Method, e.Path, e.ResponseStatus, f.odataPfad)
	}
}

// TestRecordErweiterungen (E1 bis E7, E9): die fünf Fälle von LH-QA-07.a; die gelaufenen Fälle stehen mit Zahl und Namen als Literal.
func TestRecordErweiterungen(t *testing.T) {
	var gelaufen []string
	for _, f := range erwFaelle() {
		t.Run(f.name, func(t *testing.T) {
			gelaufen = append(gelaufen, f.name)
			erwPruefe(t, f)
		})
	}
	if soll := []string{"Query", "Pfad", "Header", "JSON", "Metadaten"}; len(gelaufen) != 5 || !slices.Equal(gelaufen, soll) {
		t.Errorf("%d Fälle gelaufen %q, erwartet 5: %q", len(gelaufen), gelaufen, soll)
	}
}
