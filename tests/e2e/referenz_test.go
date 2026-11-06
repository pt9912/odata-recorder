package e2e_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pt9912/odata-recorder/tests/e2e/fixture"
)

// Die drei Geheimnisse des Referenzdienstes (SPEC-037 Nr. 4): A reist im Request-Header Authorization,
// B im Request-Header Cookie, C im Antwort-Header Set-Cookie.
const (
	geheimA = "GEHEIM-A-4f7c1e"
	geheimB = "GEHEIM-B-9d2a60"
	geheimC = "GEHEIM-C-5b8e33"
)

// sensibel ist ein Feld der Standardmenge von REDACT_JSON_PATHS mit je einem Wert im Request- und im
// Antwort-Body.
type sensibel struct{ name, anfrage, antwort string }

// sensibleFelder sind die 7 Pfade `$..<name>` der Standardmenge von REDACT_JSON_PATHS.
func sensibleFelder() []sensibel {
	var out []sensibel
	for i, n := range []string{"password", "passwd", "secret", "client_secret", "access_token", "refresh_token", "api_key"} {
		out = append(out, sensibel{n, fmt.Sprintf("GEHEIM-Q%d-anfrage", i+1), fmt.Sprintf("GEHEIM-Q%d-antwort", i+1)})
	}
	return out
}

// feldBody baut einen JSON-Body mit den sensiblen Feldern unter `Konto` und einem Feld ohne Geheimnis.
func feldBody(antwort bool) string {
	var teile []string
	for _, f := range sensibleFelder() {
		wert := f.anfrage
		if antwort {
			wert = f.antwort
		}
		teile = append(teile, fmt.Sprintf("%q:%q", f.name, wert))
	}
	return `{"Titel":"` + kontrollBody + `","Konto":{` + strings.Join(teile, ",") + `}}`
}

// kontrollBody steht in beiden Bodies ohne Redaktion und muss in einem Blob gefunden werden.
const kontrollBody = "KONTROLLE-BODY-7d2e"

// zaehlerFolge sind die Antworten auf `/odata/Products/$count` in der Reihenfolge der Requests.
func zaehlerFolge() []string { return []string{"7", "11", "13"} }

// referenz ist der Referenzdienst: ein Server im Prozess des Tests mit zwei Zählern (SPEC-037 Nr. 4).
type referenz struct {
	srv          *httptest.Server
	verbindungen atomic.Int64
	requests     atomic.Int64
	// mitGeheimnis zählt die Requests, deren Authorization GEHEIM-A trug.
	mitGeheimnis atomic.Int64
	// zaehlerRequests wählt mit der Zahl der Requests auf $count den Wert der Antwort.
	zaehlerRequests atomic.Int64
	// felderErreicht zählt die POST, deren Body jeden sensiblen Wert der Anfrage trug.
	felderErreicht atomic.Int64
}

func neueReferenz(t *testing.T) *referenz {
	t.Helper()
	r := &referenz{}
	r.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.requests.Add(1)
		if strings.Contains(req.Header.Get("Authorization"), geheimA) {
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

// url ist die Adresse des Dienstes ohne Pfad; der Dienst liegt unter url()+"/odata".
func (r *referenz) url() string { return r.srv.URL }

// stoppe schließt den Server; die Schleife nimmt danach keine Verbindung an, ein zweites Close ist wirkungslos.
func (r *referenz) stoppe() { r.srv.Close() }

// zaehler nennt Verbindungen und Requests.
func (r *referenz) zaehler() [2]int64 { return [2]int64{r.verbindungen.Load(), r.requests.Load()} }

// seite ist der Body einer Seite mit einem Eintrag.
func (r *referenz) seite(id, name string, weiter bool) string {
	link := ""
	if weiter {
		link = fmt.Sprintf(`"@odata.nextLink":%q,`, r.url()+"/odata/Products?$skiptoken=1")
	}
	return fmt.Sprintf(`{"@odata.context":%q,%s"value":[{"Id":%s,"Name":%q}]}`, r.url()+"/odata/$metadata#Products", link, id, name)
}

func (r *referenz) bediene(w http.ResponseWriter, req *http.Request) {
	h := w.Header()
	h.Set("OData-Version", "4.0")
	h.Set("Set-Cookie", "sid="+geheimC)
	schreibe := func(status int, typ, etag, body string) {
		h.Set("Content-Type", typ)
		if etag != "" {
			h.Set("ETag", etag)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
	const jsonTyp = "application/json;odata.metadata=minimal"
	q := req.URL.Query()
	switch pfad := req.URL.Path; {
	case pfad == "/odata/$metadata":
		schreibe(200, "application/xml", "", fixture.CSDL)
	case pfad == "/odata/Products" && req.Method == http.MethodPost:
		b, _ := io.ReadAll(req.Body)
		if strings.Contains(string(b), feldBody(false)) {
			r.felderErreicht.Add(1)
		}
		schreibe(http.StatusCreated, jsonTyp, "", fmt.Sprintf(`{"Id":%d,"Name":"neu"}`, len(b)))
	case pfad == "/odata/Products" && strings.HasPrefix(q.Get("$filter"), "Id eq "):
		n := strings.TrimPrefix(q.Get("$filter"), "Id eq ")
		schreibe(200, jsonTyp, `W/"f`+n+`"`, r.seite(n, map[string]string{"1": "Anvil", "2": "Bolt"}[n], false))
	case pfad == "/odata/Products" && q.Get("$skiptoken") == "1":
		schreibe(200, jsonTyp, `W/"p2"`, r.seite("2", "Bolt", false))
	case pfad == "/odata/Products":
		schreibe(200, jsonTyp, `W/"p1"`, r.seite("1", "Anvil", true))
	case pfad == "/odata/Products(1)":
		schreibe(200, jsonTyp, `W/"e1"`, fmt.Sprintf(`{"@odata.context":%q,"Id":1,"Name":"Anvil"}`, r.url()+"/odata/$metadata#Products/$entity"))
	case pfad == "/odata/Products/$count":
		n := int(r.zaehlerRequests.Add(1)) - 1
		schreibe(200, "text/plain", "", zaehlerFolge()[n%len(zaehlerFolge())])
	case pfad == "/odata/Products(1)/Name/$value":
		schreibe(200, "text/plain", "", "Anvil")
	case pfad == "/odata/Konto":
		schreibe(200, jsonTyp, "", feldBody(true))
	case pfad == "/odata/Leer":
		w.WriteHeader(http.StatusOK)
	case pfad == "/odata/Nichts":
		w.WriteHeader(http.StatusNoContent)
	default:
		schreibe(http.StatusNotFound, "application/json", "", `{"error":{"code":"404","message":"nicht gefunden"}}`)
	}
}
