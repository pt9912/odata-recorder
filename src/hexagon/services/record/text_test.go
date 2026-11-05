package record_test

import (
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Z12 (SPEC-121 Nr. 5): die Antwort trägt die Ziele aus TextZiel des OData-Pfads, der Request nie;
// Antworten ohne Body gehen nicht in die Pipeline. Die Pfade stehen nach dem Service-Root `/a`.

func textHeader(typ string) map[string][]string { return map[string][]string{"Content-Type": {typ}} }

func TestTextAntwortZiele(t *testing.T) {
	faelle := []struct {
		pfad, query string
		ct, body    string
		soll        model.CaptureStatus
		blob        string
		red         model.RedactionStatus
	}{
		{"/a/Products/$count", "", "text/plain", "42", model.CaptureComplete, "42", model.RedactionNone},
		{"/a/Products/$count", "x=1", "text/plain", "42", model.CaptureComplete, "42", model.RedactionNone},
		{"/a/Products(1)/Orders/$count", "", "text/plain; charset=utf-8", "7", model.CaptureComplete, "7", model.RedactionNone},
		{"/a/Products(1)/Name/$value", "", "text/plain", "Anvil", model.CaptureComplete, "Anvil", model.RedactionNone},
		{"/a/Products(1)/password/$value", "", "text/plain", "S3CR3T", model.CaptureComplete, model.Marker, model.RedactionApplied},
		{"/a/Products(1)/PASSWORD/$value", "", "text/plain", "S3CR3T", model.CaptureComplete, model.Marker, model.RedactionApplied},
		{"/a/Products(1)/pass%77ord/$value", "", "text/plain", "S3CR3T", model.CaptureComplete, model.Marker, model.RedactionApplied},
		{"/a/Products(1)/Name/$value", "token=S3CR3T", "text/csv", "Anvil", model.CaptureComplete, "Anvil", model.RedactionNone},
		// Keine Ziele: Text ist unbekannt.
		{"/a/$count", "", "text/plain", "42", model.CaptureFailed, "", ""},
		{"/a/%24count", "", "text/plain", "42", model.CaptureFailed, "", ""},
		{"/a/Products/%24count", "", "text/plain", "42", model.CaptureFailed, "", ""},
		{"/a/Products", "x=/$count", "text/plain", "42", model.CaptureFailed, "", ""},
		{"/a/Products", "x=/Name/$value", "text/plain", "Anvil", model.CaptureFailed, "", ""},
		{"/a/Products(1)/$value", "", "text/plain", "Anvil", model.CaptureFailed, "", ""},
		{"/a/Products(1)/Name", "", "text/plain", "Anvil", model.CaptureFailed, "", ""},
		{"/a/Products('a/password')/$value", "", "text/plain", "S3CR3T", model.CaptureFailed, "", ""},
		{"/a/Products(1)/Name/%24value", "", "text/plain", "Anvil", model.CaptureFailed, "", ""},
		// Ein anderer Zeichensatz und eine Nicht-Ziffer scheitern an der Antwort.
		{"/a/Products/$count", "", "text/plain; charset=latin1", "42", model.CaptureFailed, "", ""},
		{"/a/Products/$count", "", "text/plain", "42\n", model.CaptureFailed, "", ""},
	}
	for _, f := range faelle {
		u := neu(t)
		u.upstream.plane(antwort(200, textHeader(f.ct), koerper(f.body)))
		a := anfrage("GET", f.pfad)
		a.RawQuery = f.query
		_, gelesen := u.rundlauf(a)
		e := u.letzter()
		if gelesen != f.body {
			t.Errorf("%s?%s: der Client las %q", f.pfad, f.query, gelesen)
		}
		if e.Capture != f.soll {
			t.Errorf("%s?%s: %s, Gründe %v", f.pfad, f.query, e.Capture, u.gruende())
			continue
		}
		if f.soll == model.CaptureFailed {
			if u.gruende()["response"] == "" || e.ResponseBody != model.BodyDiscarded {
				t.Errorf("%s?%s: Body %s, Gründe %v", f.pfad, f.query, e.ResponseBody, u.gruende())
			}
			continue
		}
		if blob, _ := u.blobs.inhalt(e.ResponseBlob); blob != f.blob || e.ResponseRedact != f.red || e.ResponseBody != model.BodyPresent {
			t.Errorf("%s?%s: Blob %q, %s, %s; erwartet %q, %s", f.pfad, f.query, blob, e.ResponseRedact, e.ResponseBody, f.blob, f.red)
		}
	}
	if len(faelle) != 19 {
		t.Errorf("%d Fälle", len(faelle))
	}
}

// TestTextAntwortKeinGeheimnisImBlob: das Geheimnis unter `/<Property>/$value` erreicht keinen Blob,
// auch nicht über die angebotenen Bytes des Writers (LH-FA-14).
func TestTextAntwortKeinGeheimnisImBlob(t *testing.T) {
	for _, pfad := range []string{"/a/Products(1)/password/$value", "/a/Products(1)/pass%77ord/$value"} {
		u := neu(t)
		u.upstream.plane(antwort(200, textHeader("text/plain"), koerper(geheim)))
		_, gelesen := u.rundlauf(anfrage("GET", pfad))
		if gelesen != geheim {
			t.Errorf("%s: der Client las %q", pfad, gelesen)
		}
		if angeboten := u.blobs.alleAngeboten(); angeboten != model.Marker {
			t.Errorf("%s: der Writer bekam %q", pfad, angeboten)
		}
	}
}

// TestTextRequestOhneZiel: ein Request als Text bekommt nie ein Ziel, auch auf den Pfaden der Ziele.
func TestTextRequestOhneZiel(t *testing.T) {
	zahl := 0
	for _, pfad := range []string{"/a/Products/$count", "/a/Products(1)/Orders/$count", "/a/Products(1)/Name/$value", "/a/Products(1)/password/$value"} {
		for _, methode := range []string{"POST", "PUT", "PATCH"} {
			zahl++
			u := neu(t)
			u.upstream.plane(leer())
			u.rundlauf(mitBody(anfrage(methode, pfad), "text/plain", "12"))
			if e := u.letzter(); e.Capture != model.CaptureFailed || e.MatchKey != "" || u.gruende()["request"] != "typ" ||
				u.upstream.gesehen()[0].body != "12" || u.blobs.anzahlBegin() != 0 {
				t.Errorf("%s %s: %s, Key %q, Gründe %v", methode, pfad, e.Capture, e.MatchKey, u.gruende())
			}
		}
	}
	if zahl != 12 {
		t.Errorf("%d Fälle", zahl)
	}
}

// TestTextAntwortOhneBody: HEAD, 1xx, 204 und 304 gehen nicht in die Pipeline, auch auf den Pfaden der Ziele.
func TestTextAntwortOhneBody(t *testing.T) {
	zahl := 0
	for _, pfad := range []string{"/a/Products/$count", "/a/Products(1)/Name/$value"} {
		for _, f := range []struct {
			methode string
			status  int
		}{{"HEAD", 200}, {"GET", 101}, {"GET", 204}, {"GET", 304}} {
			zahl++
			u := neu(t)
			h := textHeader("text/plain")
			h["Content-Encoding"] = []string{"br"}
			u.upstream.plane(antwort(f.status, h, koerper()))
			u.rundlauf(anfrage(f.methode, pfad))
			if e := u.letzter(); e.ResponseBody != model.BodyAbsent || e.Capture != model.CaptureComplete || u.blobs.anzahlBegin() != 0 {
				t.Errorf("%s %d %s: %s %s, Gründe %v", f.methode, f.status, pfad, e.ResponseBody, e.Capture, u.gruende())
			}
		}
	}
	if zahl != 8 {
		t.Errorf("%d Fälle", zahl)
	}
}
