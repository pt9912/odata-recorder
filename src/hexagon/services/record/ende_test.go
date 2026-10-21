package record_test

import (
	"context"
	"io"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func jsonAntwort(b *fakeBody) *plan {
	return antwort(200, map[string][]string{"Content-Type": {"application/json"}}, b)
}

// TestEnde: Close beendet den Exchange; io.EOF mit n > 0 zählt; der Request ohne io.EOF ist
// Abbruch; Upstream-Body einmal geschlossen; zweites Close wirkungslos; Read nach Close (B16).
func TestEnde(t *testing.T) {
	a := neu(t)
	ba := (&fakeBody{}).dann(nil, `{"a":1}`, io.EOF)
	a.upstream.plane(jsonAntwort(ba))
	ant, gelesen := a.rundlauf(anfrage("GET", "/a/x"))
	if e := a.letzter(); gelesen != `{"a":1}` || e.Capture != model.CaptureComplete {
		t.Errorf("(a) io.EOF mit n > 0: %q, %s", gelesen, e.Capture)
	}
	if n := ba.closes.Load(); n != 1 {
		t.Errorf("(f) Upstream-Body %d-mal geschlossen", n)
	}
	schliesse(t, ant.Body)
	if n := len(a.exchanges.angehaengt()); n != 1 || ba.closes.Load() != 1 {
		t.Errorf("(d) zweites Close: %d Exchanges, %d-mal geschlossen", n, ba.closes.Load())
	}
	if n, err := ant.Body.Read(make([]byte, 4)); n != 0 || err != io.ErrClosedPipe {
		t.Errorf("(e) Read nach Close: %d %v", n, err)
	}

	b := neu(t)
	b.upstream.plane(jsonAntwort(koerper(`{"a":`, `1}`)))
	ant, err := b.record(context.Background(), anfrage("GET", "/a/x"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ant.Body.Read(make([]byte, 5)); err != nil {
		t.Fatal(err)
	}
	schliesse(t, ant.Body)
	if e := b.letzter(); e.Capture != model.CaptureIncomplete || e.ResponseBody != model.BodyDiscarded || b.gruende()["response"] != "abbruch" {
		t.Errorf("(b) Close nach der Hälfte: %s %s %v", e.Capture, e.ResponseBody, b.gruende())
	}

	c := neu(t)
	p := jsonAntwort(koerper(`{"b":2}`))
	p.lies = 3
	c.upstream.plane(p)
	c.rundlauf(mitBody(anfrage("POST", "/a/x"), "application/json", `{"a":1234}`))
	if e := c.letzter(); e.Capture != model.CaptureFailed || e.MatchKey != "" || e.RequestBody != model.BodyDiscarded ||
		c.gruende()["request"] != "abbruch" || c.upstream.gesehen()[0].body != `{"a` {
		t.Errorf("(c) Request nicht zu Ende gelesen: %s, Key %q, %v", e.Capture, e.MatchKey, c.gruende())
	}
}

// TestAbbruchNachKopf: Client-Abbruch und Lesefehler nach dem Kopf sind incomplete; ein Ergebnis
// vor dem Abbruch bleibt complete (B30).
func TestAbbruchNachKopf(t *testing.T) {
	u := neu(t)
	halt := make(chan struct{})
	u.upstream.plane(jsonAntwort(koerper(`{"a":`).dann(halt, `1}`, nil)))
	ctx, abbrechen := context.WithCancel(context.Background())
	ant, err := u.record(ctx, anfrage("GET", "/a/x"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ant.Body.Read(make([]byte, 5)); err != nil {
		t.Fatal(err)
	}
	abbrechen()
	if _, err := ant.Body.Read(make([]byte, 5)); err == nil || err == io.EOF {
		t.Errorf("Read nach dem Abbruch: %v", err)
	}
	schliesse(t, ant.Body)
	if e := u.letzter(); e.Capture != model.CaptureIncomplete || u.gruende()["response"] != "abbruch" {
		t.Errorf("Client-Abbruch: %s, %v", e.Capture, u.gruende())
	}

	v := neu(t)
	v.upstream.plane(jsonAntwort(koerper(`{"a":`).dann(nil, "", errVerbindung)))
	v.rundlauf(anfrage("GET", "/a/x"))
	if e := v.letzter(); e.Capture != model.CaptureIncomplete || v.gruende()["response"] != "abbruch" {
		t.Errorf("Lesefehler: %s, %v", e.Capture, v.gruende())
	}

	w := neu(t)
	w.upstream.plane(jsonAntwort(koerper(`{"b":2}`)))
	ctx, abbrechen = context.WithCancel(context.Background())
	defer abbrechen()
	ant, err = w.record(ctx, mitBody(anfrage("POST", "/a/x"), "application/json", `{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := io.ReadAll(ant.Body); string(b) != `{"b":2}` {
		t.Fatalf("gelesen %q", b)
	}
	empfange(t, "Commit 1", w.blobs.commits)
	empfange(t, "Commit 2", w.blobs.commits)
	abbrechen()
	schliesse(t, ant.Body)
	if e := w.letzter(); e.Capture != model.CaptureComplete || len(w.gruende()) != 0 {
		t.Errorf("Ergebnis vor dem Abbruch: %s, %v", e.Capture, w.gruende())
	}
}
