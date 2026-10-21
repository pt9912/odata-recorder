package record_test

import (
	"bytes"
	"compress/gzip"
	"reflect"
	"strings"
	"testing"
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

// TestRequestAufnahme: ohne Body oder bei Laenge 0 keine Aufnahme; sonst Tee mit Content-Type und
// Content-Encoding, und der Upstream liest aus der Aufnahme (B11).
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
	if e := n.letzter(); e.RequestBody != model.BodyAbsent || e.Capture != model.CaptureComplete || quelle.Len() != 1 || n.blobs.anzahlBegin() != 0 {
		t.Errorf("Laenge 0: %s %s, ungelesen %d", e.RequestBody, e.Capture, quelle.Len())
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

	z := neu(t)
	z.upstream.plane(leer())
	gz := gzipt(t, `{"password":"S3CR3T"}`)
	a = mitBody(anfrage("POST", "/a/x"), "application/json", gz)
	a.Header["Content-Encoding"] = []string{"gzip"}
	z.rundlauf(a)
	if blob, _ := z.blobs.inhalt(z.letzter().RequestBlob); blob != `{"password":"[REDACTED]"}` || z.upstream.gesehen()[0].body != gz {
		t.Errorf("gzip: Blob %q", blob)
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

// TestResponseAufnahme: Metadata genau bei /$metadata; keine Aufnahme bei HEAD, 1xx, 204, 304 (B12).
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
}

// TestExchangeFelder: Kennung, Zeiten, Pfad ohne Root, Header und Query wie empfangen und
// redigiert, Antwort-Header redigiert (B14).
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
