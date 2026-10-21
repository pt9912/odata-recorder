package record_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

// TestOhneAntwortZustand: der Zustand ohne Antwort je Ergebnis des Requests, von Hand; jeder
// besteht Validate mit Key genau bei KeyMoeglich (B22).
func TestOhneAntwortZustand(t *testing.T) {
	blob := refVon([]byte(`{}`))
	verworfen := func(a capture.Ausgang) capture.Ergebnis {
		return capture.Ergebnis{Richtung: capture.Request, Ausgang: a, Grund: capture.GrundTyp, Body: model.BodyDiscarded,
			Redaction: model.RedactionUnsafe, Integrity: model.IntegrityAbsent, Fehler: &capture.Fehler{Grund: capture.GrundTyp}}
	}
	vollstaendig := func(body model.BodyState, ref string, r model.RedactionStatus, i model.BodyIntegrity) capture.Ergebnis {
		return capture.Ergebnis{Richtung: capture.Request, Ausgang: capture.Vollstaendig, Body: body, Blob: ref, Redaction: r, Integrity: i}
	}
	ohne := func(c model.CaptureStatus, key bool, i model.BodyIntegrity, body model.BodyState, ref string, r model.RedactionStatus) capture.Zustand {
		return capture.Zustand{Capture: c, KeyMoeglich: key, Integrity: i, Equivalence: model.EquivalenceNone,
			RequestBody: body, RequestBlob: ref, RequestRedact: r, ResponseBody: model.BodyAbsent, ResponseRedact: model.RedactionNone}
	}
	faelle := []struct {
		name string
		req  capture.Ergebnis
		soll capture.Zustand
	}{
		{"absent", vollstaendig(model.BodyAbsent, "", model.RedactionNone, model.IntegrityAbsent),
			ohne(model.CaptureIncomplete, true, model.IntegrityAbsent, model.BodyAbsent, "", model.RedactionNone)},
		{"present none", vollstaendig(model.BodyPresent, blob, model.RedactionNone, model.IntegrityVerified),
			ohne(model.CaptureIncomplete, true, model.IntegrityVerified, model.BodyPresent, blob, model.RedactionNone)},
		{"present applied", vollstaendig(model.BodyPresent, blob, model.RedactionApplied, model.IntegrityVerified),
			ohne(model.CaptureIncomplete, true, model.IntegrityVerified, model.BodyPresent, blob, model.RedactionApplied)},
		{"Abbruch", verworfen(capture.Abbruch),
			ohne(model.CaptureFailed, false, model.IntegrityAbsent, model.BodyDiscarded, "", model.RedactionUnsafe)},
		{"Verstoss", verworfen(capture.Verstoss),
			ohne(model.CaptureFailed, false, model.IntegrityAbsent, model.BodyDiscarded, "", model.RedactionUnsafe)},
	}
	for _, f := range faelle {
		z := record.OhneAntwort(f.req)
		if z != f.soll {
			t.Errorf("%s:\n%+v\nerwartet\n%+v", f.name, z, f.soll)
		}
		e := model.Exchange{ID: "e", SessionID: "s", SequenceNo: 1, Method: "GET", Path: "/", StartedAt: t0(),
			Capture: z.Capture, Integrity: z.Integrity, Equivalence: z.Equivalence, RequestBody: z.RequestBody,
			RequestBlob: z.RequestBlob, RequestRedact: z.RequestRedact, ResponseBody: z.ResponseBody, ResponseRedact: z.ResponseRedact}
		if z.KeyMoeglich {
			e.MatchKey = blob
		}
		if err := e.Validate(); err != nil {
			t.Errorf("%s: Validate %v", f.name, err)
		}
	}
}

// TestOhneAntwort: Rangfolge Kontext, Drain, Zeit, nicht erreichbar, andere; der Exchange ist
// angehängt, die Antwort ist der Nullwert, FinishedAt liegt nach dem Ergebnis des Requests (B22).
func TestOhneAntwort(t *testing.T) {
	for _, f := range []struct {
		err    error
		code   model.Code
		status int
	}{
		{outbound.ErrUpstreamNichtErreichbar, model.CodeE4001, 502},
		{outbound.ErrUpstreamZeit, model.CodeE4001, 504},
		{errAnders, model.CodeE9001, 500},
		{model.ErrUngueltig, model.CodeE9001, 500},
	} {
		for _, post := range []bool{false, true} {
			u := neu(t)
			p := &plan{lies: 0, fehler: f.err, vorAntwort: func() { u.uhr.Advance(3 * time.Second) }}
			a := anfrage("GET", "/a/x")
			soll := model.CaptureIncomplete
			if post {
				a, soll = mitBody(anfrage("POST", "/a/x"), "application/json", `{"a":1}`), model.CaptureFailed
			}
			u.upstream.plane(p)
			r := empfange(t, "Record", u.recordNebenbei(context.Background(), a))
			ant, err := r.ant, r.err
			pruefeCode(t, err, f.code, f.status)
			if !errors.Is(err, f.err) || ant.Body != nil || ant.Status != 0 || ant.Header != nil {
				t.Errorf("%v: Ursache fehlt oder Antwort %+v", f.err, ant)
			}
			pruefeOhneAntwort(t, u, soll)
			if e := u.letzter(); !e.StartedAt.Equal(t0()) || !e.FinishedAt.Equal(t0().Add(3*time.Second)) {
				t.Errorf("%v: StartedAt %v, FinishedAt %v, erwartet t0 und t0 + 3 s", f.err, e.StartedAt, e.FinishedAt)
			}
		}
	}
	z := neu(t)
	z.upstream.plane(&plan{lies: -1, fehler: errors.Join(outbound.ErrUpstreamNichtErreichbar, outbound.ErrUpstreamZeit)})
	_, err := z.record(context.Background(), anfrage("GET", "/a/x"))
	pruefeCode(t, err, model.CodeE4001, 504)
	pruefeEintraege(t, "Zeit und nicht erreichbar", z.log.alle(),
		[]map[string]string{{"ereignis": "upstream", "error_code": "ODR-E4001", "grund": "zeit"}})

	u := neu(t)
	u.upstream.plane(&plan{lies: -1, warteAufCtx: true})
	ctx, abbrechen := context.WithCancel(context.Background())
	res := u.recordNebenbei(ctx, anfrage("GET", "/a/x"))
	empfange(t, "Senden", u.upstream.imSenden)
	abbrechen()
	if r := empfange(t, "Record", res); r.err != context.Canceled || r.ant.Body != nil {
		t.Errorf("Client-Abbruch vor dem Kopf: %v", r.err)
	}
	pruefeOhneAntwort(t, u, model.CaptureIncomplete)

	d := neu(t)
	d.upstream.plane(&plan{lies: -1, warteAufCtx: true})
	res = d.recordNebenbei(context.Background(), anfrage("GET", "/a/x"))
	empfange(t, "Senden", d.upstream.imSenden)
	dr := d.drainNebenbei(context.Background(), "s1", 0)
	pruefeCode(t, empfange(t, "Record", res).err, model.CodeE4001, 504)
	empfange(t, "Drain", dr)
	pruefeOhneAntwort(t, d, model.CaptureIncomplete)

	b := neu(t)
	nach := make(chan struct{})
	b.upstream.plane(&plan{lies: -1, warteAufCtx: true, nachCtx: nach})
	ctx, abbrechen = context.WithCancel(context.Background())
	defer abbrechen()
	res = b.recordNebenbei(ctx, anfrage("GET", "/a/x"))
	empfange(t, "Senden", b.upstream.imSenden)
	dr = b.drainNebenbei(context.Background(), "s1", 0)
	empfange(t, "Ende des Kontexts", b.upstream.ctxEnde)
	abbrechen()
	close(nach)
	if r := empfange(t, "Record", res); r.err != context.Canceled {
		t.Errorf("Drain und Client zugleich: %v", r.err)
	}
	empfange(t, "Drain", dr)
}

// TestOhneAntwortRangfolge: endet der Kontext des Clients oder bricht Drain den Exchange ab,
// während der Dienst ihn ohne Antwort anhängt, gilt die Rangfolge beim Liefern: Kontext vor Drain
// vor jedem Fehler des Ports (B22). Der Recording-Fake hält Append an.
func TestOhneAntwortRangfolge(t *testing.T) {
	for _, portFehler := range []error{outbound.ErrUpstreamZeit, outbound.ErrUpstreamNichtErreichbar, errAnders} {
		for _, ursache := range []string{"Client", "Drain"} {
			u := neu(t)
			u.exchanges.imAppend, u.exchanges.haltAppend = make(chan struct{}, 1), make(chan struct{})
			u.upstream.plane(&plan{lies: -1, fehler: portFehler})
			ctx, abbrechen := context.WithCancel(context.Background())
			res := u.recordNebenbei(ctx, anfrage("GET", "/a/x"))
			empfange(t, "Append", u.exchanges.imAppend)
			var dr <-chan drainErgebnis
			if ursache == "Client" {
				abbrechen()
			} else {
				dr = u.drainNebenbei(context.Background(), "s1", 0)
				empfange(t, "Ende des Kontexts", u.upstream.gesehen()[0].done)
			}
			close(u.exchanges.haltAppend)
			r := empfange(t, "Record", res)
			abbrechen()
			pruefeOhneAntwort(t, u, model.CaptureIncomplete)
			if ursache == "Client" {
				if r.err != context.Canceled || len(u.log.alle()) != 0 {
					t.Errorf("%v, Client: %v, Einträge %v", portFehler, r.err, u.log.alle())
				}
				continue
			}
			pruefeCode(t, r.err, model.CodeE4001, 504)
			pruefeEintraege(t, fmt.Sprintf("%v, Drain", portFehler), u.log.alle(),
				[]map[string]string{{"ereignis": "upstream", "error_code": "ODR-E4001", "grund": "drain"}})
			if b := empfange(t, "Drain", dr); b.b.Abgebrochen != 1 {
				t.Errorf("%v, Drain: Bericht %+v", portFehler, b.b)
			}
		}
	}
}

func pruefeOhneAntwort(t *testing.T, u *umgebung, soll model.CaptureStatus) {
	t.Helper()
	e := u.letzter()
	if e.Capture != soll || e.ResponseStatus != 0 || e.ResponseHeaders != nil || e.ResponseBody != model.BodyAbsent ||
		e.Equivalence != model.EquivalenceNone || (e.MatchKey != "") != (soll == model.CaptureIncomplete) {
		t.Errorf("ohne Antwort: %+v, erwartet %s", e, soll)
	}
}

// TestFehlerWert: jeder *Fehler, den Record in diesem Paket liefert, prüft die Hilfe record
// (erlaubter Status, Text); hier die Ursache und der Text ohne Wert der Ursache (B25).
func TestFehlerWert(t *testing.T) {
	u := neu(t)
	u.upstream.plane(&plan{lies: -1, fehler: errors.Join(outbound.ErrUpstreamZeit, errDisk)})
	_, err := u.record(context.Background(), anfrage("GET", "/a/x"))
	var f *inbound.Fehler
	if !errors.As(err, &f) || !errors.Is(err, outbound.ErrUpstreamZeit) || f.Error() != "ODR-E4001: Upstream unavailable" {
		t.Errorf("Zeit: %v", err)
	}
	for _, pfad := range []string{"/c/x", "/a/x"} {
		v := neu(t)
		v.sessions.setze("s1", "svc-a", model.SessionComplete)
		_, err := v.record(context.Background(), anfrage("GET", pfad))
		pruefeFehlerWert(t, err)
		if f := alsFehler(t, err); f.Status != f.Code.Stati()[0] {
			t.Errorf("%s: Status %d", pfad, f.Status)
		}
	}
}
