package record_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"maps"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
)

func gzipt(t *testing.T, s string) string {
	t.Helper()
	var b bytes.Buffer
	z := gzip.NewWriter(&b)
	if _, err := z.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func leer() *plan {
	return antwort(200, map[string][]string{"Content-Type": {"application/json"}}, koerper())
}

// gruende liefert die Gründe der Einträge capture je Richtung.
func (u *umgebung) gruende() map[string]string {
	out := map[string]string{}
	for _, e := range u.log.alle() {
		if e.attrs["ereignis"] == "capture" {
			out[e.attrs["richtung"]] = e.attrs["grund"]
		}
	}
	return out
}

// TestRequestAufnahme: ohne Body oder bei Laenge 0 keine Aufnahme und Laenge 0 am Upstream; sonst
// Tee mit allen Werten von Content-Type und Content-Encoding, der Upstream liest aus der Aufnahme
// und erhält Laenge (B11).
func TestRequestAufnahme(t *testing.T) {
	u := neu(t)
	u.upstream.plane(leer())
	u.rundlauf(anfrage("GET", "/a/x"))
	if e, g := u.letzter(), u.upstream.gesehen()[0]; e.RequestBody != model.BodyAbsent || g.a.Laenge != 0 || g.a.Body != nil ||
		u.blobs.anzahlBegin() != 0 || e.Capture != model.CaptureComplete {
		t.Errorf("GET: %+v, Laenge %d, Begin %d", e, g.a.Laenge, u.blobs.anzahlBegin())
	}

	n := neu(t)
	n.upstream.plane(leer())
	a := mitBody(anfrage("POST", "/a/x"), "application/json", "x")
	quelle := strings.NewReader("x")
	a.Body, a.Laenge = quelle, 0
	n.rundlauf(a)
	if e, g := n.letzter(), n.upstream.gesehen()[0]; e.RequestBody != model.BodyAbsent || e.Capture != model.CaptureComplete ||
		quelle.Len() != 1 || n.blobs.anzahlBegin() != 0 || g.a.Laenge != 0 || g.a.Body != nil {
		t.Errorf("Laenge 0: %s %s, ungelesen %d, am Upstream Laenge %d", e.RequestBody, e.Capture, quelle.Len(), g.a.Laenge)
	}

	j := neu(t)
	j.upstream.plane(leer())
	a = mitBody(anfrage("POST", "/a/x"), "application/json", `{"password":"S3CR3T","a":1}`)
	a.Header = map[string][]string{"content-type": {"application/json"}}
	a.Laenge = -1
	j.rundlauf(a)
	e := j.letzter()
	if blob, _ := j.blobs.inhalt(e.RequestBlob); blob != `{"password":"[REDACTED]","a":1}` || e.RequestRedact != model.RedactionApplied {
		t.Errorf("Blob %q, %s", blob, e.RequestRedact)
	}
	if g := j.upstream.gesehen()[0]; g.body != `{"password":"S3CR3T","a":1}` || g.a.Laenge != -1 {
		t.Errorf("Upstream erhielt %q mit Laenge %d", g.body, g.a.Laenge)
	}

	gz := gzipt(t, `{"password":"S3CR3T"}`)
	// Alle Werte von Content-Encoding in jeder Schreibweise des Namens: die Kette entpackt gzip;
	// der Upstream erhält die Bytes und ihre Länge.
	for _, ce := range []map[string][]string{
		{"Content-Encoding": {"gzip"}},
		{"content-encoding": {"gzip"}},
		{"Content-Encoding": {"identity", "gzip"}},
		{"Content-Encoding": {"gzip", "identity"}},
	} {
		z := neu(t)
		z.upstream.plane(leer())
		a = mitBody(anfrage("POST", "/a/x"), "application/json", gz)
		maps.Copy(a.Header, ce)
		z.rundlauf(a)
		e, g := z.letzter(), z.upstream.gesehen()[0]
		if blob, _ := z.blobs.inhalt(e.RequestBlob); blob != `{"password":"[REDACTED]"}` || e.Capture != model.CaptureComplete ||
			g.body != gz || g.a.Laenge != int64(len(gz)) {
			t.Errorf("%v: Blob %q, %s, am Upstream Laenge %d", ce, blob, e.Capture, g.a.Laenge)
		}
	}

	// Alle Werte von Content-Type: zwei Werte sind typ (SPEC-118 Nr. 3), in jeder Reihenfolge.
	for _, ct := range [][]string{{"application/json", "text/plain"}, {"text/plain", "application/json"}} {
		w := neu(t)
		w.upstream.plane(leer())
		a = mitBody(anfrage("POST", "/a/x"), "application/json", `{"a":1}`)
		a.Header["Content-Type"] = ct
		w.rundlauf(a)
		if e := w.letzter(); e.Capture != model.CaptureFailed || w.gruende()["request"] != "typ" {
			t.Errorf("Content-Type %v: %s, Gründe %v", ct, e.Capture, w.gruende())
		}
	}

	for _, typ := range []string{"text/plain", "multipart/mixed; boundary=b"} {
		w := neu(t)
		w.upstream.plane(leer())
		w.rundlauf(mitBody(anfrage("POST", "/a/x"), typ, "--b S3CR3T"))
		if e := w.letzter(); e.Capture != model.CaptureFailed || e.MatchKey != "" || w.gruende()["request"] != "typ" ||
			w.upstream.gesehen()[0].body != "--b S3CR3T" {
			t.Errorf("%s: %s, Key %q, Gründe %v", typ, e.Capture, e.MatchKey, w.gruende())
		}
	}
}

// TestResponseAufnahme: Tee mit allen Werten von Content-Type und Content-Encoding; Metadata genau
// bei /$metadata; keine Aufnahme bei HEAD, 1xx, 204, 304 (B12).
func TestResponseAufnahme(t *testing.T) {
	xml := map[string][]string{"Content-Type": {"application/xml"}}
	for _, f := range []struct {
		pfad, query string
		soll        model.CaptureStatus
	}{
		{"/a/$metadata", "", model.CaptureComplete},
		{"/a/$metadata", "x=1", model.CaptureComplete},
		{"/a/%24metadata", "", model.CaptureFailed},
		{"/a/Products", "", model.CaptureFailed},
		{"/a/$metadataX", "", model.CaptureFailed},
		{"/a/$metadata/x", "", model.CaptureFailed},
		{"/a/x/$metadata", "", model.CaptureFailed},
	} {
		u := neu(t)
		u.upstream.plane(antwort(200, map[string][]string{"content-type": xml["Content-Type"]}, koerper("<edmx/>")))
		a := anfrage("GET", f.pfad)
		a.RawQuery = f.query
		u.rundlauf(a)
		e := u.letzter()
		if e.Capture != f.soll {
			t.Errorf("%s?%s: %s, Gründe %v", f.pfad, f.query, e.Capture, u.gruende())
		}
		if f.soll == model.CaptureComplete && (e.ResponseBody != model.BodyPresent || e.ResponseRedact != model.RedactionNone) {
			t.Errorf("%s: %s %s", f.pfad, e.ResponseBody, e.ResponseRedact)
		}
		if f.soll == model.CaptureFailed && u.gruende()["response"] != "typ" {
			t.Errorf("%s: Gründe %v", f.pfad, u.gruende())
		}
	}
	h := map[string][]string{"Content-Encoding": {"br"}, "Content-Type": {"text/plain"}}
	for _, f := range []struct {
		methode string
		status  int
	}{{"HEAD", 200}, {"GET", 101}, {"GET", 204}, {"GET", 304}} {
		u := neu(t)
		u.upstream.plane(antwort(f.status, h, koerper()))
		u.rundlauf(anfrage(f.methode, "/a/x"))
		if e := u.letzter(); e.ResponseBody != model.BodyAbsent || e.Capture != model.CaptureComplete || u.blobs.anzahlBegin() != 0 {
			t.Errorf("%s %d: %s %s, Gründe %v", f.methode, f.status, e.ResponseBody, e.Capture, u.gruende())
		}
	}
	u := neu(t)
	u.upstream.plane(antwort(200, h, koerper()))
	u.rundlauf(anfrage("head", "/a/x"))
	if e := u.letzter(); e.Capture != model.CaptureFailed || u.gruende()["response"] != "encoding" {
		t.Errorf("head: %s, Gründe %v", e.Capture, u.gruende())
	}

	// Alle Werte von Content-Encoding und Content-Type der Antwort, je in beiden Reihenfolgen.
	gz := gzipt(t, `{"password":"S3CR3T"}`)
	for _, f := range []struct {
		header map[string][]string
		body   string
		soll   model.CaptureStatus
	}{
		{map[string][]string{"Content-Type": {"application/json"}, "Content-Encoding": {"identity", "gzip"}}, gz, model.CaptureComplete},
		{map[string][]string{"Content-Type": {"application/json"}, "Content-Encoding": {"gzip", "identity"}}, gz, model.CaptureComplete},
		{map[string][]string{"Content-Type": {"application/json", "text/plain"}}, `{"a":1}`, model.CaptureFailed},
		{map[string][]string{"Content-Type": {"text/plain", "application/json"}}, `{"a":1}`, model.CaptureFailed},
	} {
		v := neu(t)
		v.upstream.plane(antwort(200, f.header, koerper(f.body)))
		_, gelesen := v.rundlauf(anfrage("GET", "/a/x"))
		e := v.letzter()
		blob, _ := v.blobs.inhalt(e.ResponseBlob)
		if gelesen != f.body || e.Capture != f.soll || f.soll == model.CaptureComplete && blob != `{"password":"[REDACTED]"}` {
			t.Errorf("%v: %s, Blob %q, Gründe %v", f.header, e.Capture, blob, v.gruende())
		}
		if f.soll == model.CaptureFailed && v.gruende()["response"] != "typ" {
			t.Errorf("%v: Gründe %v", f.header, v.gruende())
		}
	}
}

// TestExchangeFelder: Kennung, Methode wie empfangen, Zeiten (FinishedAt nach beiden Ergebnissen),
// Pfad ohne Root, Header und Query wie empfangen und redigiert, Antwort-Header redigiert (B14).
func TestExchangeFelder(t *testing.T) {
	u := neu(t)
	p := antwort(201, map[string][]string{"Location": {"/x?token=S3CR3T"}, "Content-Type": {"application/json"}}, koerper())
	p.vorAntwort = func() { u.uhr.Advance(5 * time.Second) }
	u.upstream.plane(p)
	a := anfrage("GET", "/a/Products")
	a.RawQuery = "token=S3CR3T&a=1"
	a.Header = map[string][]string{sessionHeader: {"s1"}, "Connection": {"keep-alive"}, "Authorization": {"S3CR3T"}}
	u.rundlauf(a)
	e := u.letzter()
	soll := model.Exchange{
		ID: kennung("exc", 1), SessionID: "s1", SequenceNo: 1, Method: "GET", Path: "/Products",
		RawQuery: "token=[REDACTED]&a=1", MatchKey: e.MatchKey, StartedAt: t0(), FinishedAt: t0().Add(5 * time.Second),
		KeyRedacted:    true,
		RequestHeaders: map[string][]string{sessionHeader: {"s1"}, "Connection": {"keep-alive"}, "Authorization": {model.Marker}},
		RequestBody:    model.BodyAbsent, ResponseStatus: 201,
		ResponseHeaders: map[string][]string{"Location": {"/x?token=[REDACTED]"}, "Content-Type": {"application/json"}},
		ResponseBody:    model.BodyAbsent, Capture: model.CaptureComplete, RequestRedact: model.RedactionNone,
		ResponseRedact: model.RedactionNone, Integrity: model.IntegrityAbsent, Equivalence: model.EquivalenceByte,
	}
	if !reflect.DeepEqual(e, soll) {
		t.Errorf("Exchange\n%+v\nerwartet\n%+v", e, soll)
	}
	if e.MatchKey == "" {
		t.Error("Key fehlt")
	}

	m := neu(t)
	m.rundlauf(anfrage("patch", "/a/x"))
	if e, g := m.letzter(), m.upstream.gesehen()[0].a; e.Method != "patch" || g.Methode != "patch" {
		t.Errorf("Methode im Exchange %q, am Upstream %q, erwartet patch", e.Method, g.Methode)
	}

	// FinishedAt nach beiden Ergebnissen: der Commit der Response wartet, bis Close auf ihr Ergebnis
	// wartet (synctest.Wait); die Uhr rückt nach dem Kopf und in dieser Zeit vor.
	synctest.Test(t, func(t *testing.T) {
		f := neu(t)
		f.blobs.haltCommit = make(chan struct{})
		ant, err := f.record(context.Background(), anfrage("GET", "/a/x"))
		if err != nil {
			t.Fatal(err)
		}
		f.uhr.Advance(2 * time.Second)
		if _, err := io.ReadAll(ant.Body); err != nil {
			t.Fatal(err)
		}
		zu := nebenbei(ant.Body.Close)
		synctest.Wait()
		f.uhr.Advance(3 * time.Second)
		close(f.blobs.haltCommit)
		if err := <-zu; err != nil {
			t.Fatal(err)
		}
		if e := f.letzter(); !e.StartedAt.Equal(t0()) || !e.FinishedAt.Equal(t0().Add(5*time.Second)) {
			t.Errorf("StartedAt %v, FinishedAt %v, erwartet t0 und t0 + 5 s", e.StartedAt, e.FinishedAt)
		}
	})
}

// TestMatchKey: Key aus dem redigierten Body und RequestRedact; ohne KeyMoeglich leer mit
// KeyRedacted aus dem Vorab-Aufruf; Kanonform-Rückfall ohne Fehler (B15).
func TestMatchKey(t *testing.T) {
	u := neu(t)
	for _, pw := range []string{"S3CR3T-A", "S3CR3T-B"} {
		u.rundlauf(mitBody(anfrage("POST", "/a/x"), "application/json", `{"password":"`+pw+`"}`))
	}
	soll, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "POST", ServiceID: "svc-a", Path: "/x",
		Header: map[string][]string{"Content-Type": {"application/json"}}, Body: []byte(`{"password":"[REDACTED]"}`),
		RequestRedaction: model.RedactionApplied}, regeln())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range u.exchanges.angehaengt() {
		if e.MatchKey != soll.MatchKey || !e.KeyRedacted {
			t.Errorf("Key %q (%v), erwartet %q", e.MatchKey, e.KeyRedacted, soll.MatchKey)
		}
	}
	for query, kr := range map[string]bool{"token=x": true, "": false} {
		w := neu(t)
		a := mitBody(anfrage("POST", "/a/x"), "text/plain", "x")
		a.RawQuery = query
		w.rundlauf(a)
		if e := w.letzter(); e.MatchKey != "" || e.Capture != model.CaptureFailed || e.KeyRedacted != kr {
			t.Errorf("text/plain mit %q: Key %q, %s, KeyRedacted %v", query, e.MatchKey, e.Capture, e.KeyRedacted)
		}
	}
	kanon := func(k *record.Konfig) {
		k.Regeln.BodyMode, k.Regeln.InvalidEncoding = model.BodyJSONCanonical, model.KodierungReject
	}
	c := neu(t, kanon)
	surrogat := `{"a":"` + `\` + `uD800"}`
	c.rundlauf(mitBody(anfrage("POST", "/a/x"), "application/json", surrogat))
	r := regeln()
	r.BodyMode, r.InvalidEncoding = model.BodyJSONCanonical, model.KodierungReject
	k, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "POST", ServiceID: "svc-a", Path: "/x",
		Header: map[string][]string{"Content-Type": {"application/json"}}, Body: []byte(surrogat),
		RequestRedaction: model.RedactionNone}, r)
	if e := c.letzter(); err != nil || e.Capture != model.CaptureComplete || e.MatchKey != k.MatchKey {
		t.Errorf("Surrogat: %s, Key %q, erwartet %q (%v)", e.Capture, e.MatchKey, k.MatchKey, err)
	}
}
