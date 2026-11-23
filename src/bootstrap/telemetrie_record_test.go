package bootstrap_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/adapters/driven/telemetry"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
	"github.com/pt9912/odata-recorder/src/hexagon/services/maintenance"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record"
	"github.com/pt9912/odata-recorder/src/hexagon/services/session"
)

// fakeUp ist der Upstream-Client des Tests: Senden liest den Body des Requests und liefert die eingestellte Antwort oder den Fehler.
type fakeUp struct {
	header map[string][]string
	body   string
	fehler error
}

func (f *fakeUp) Senden(_ context.Context, a outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
	if a.Body != nil {
		_, _ = io.Copy(io.Discard, a.Body)
	}
	if f.fehler != nil {
		return outbound.UpstreamAntwort{}, f.fehler
	}
	return outbound.UpstreamAntwort{Status: 200, Header: f.header, Body: io.NopCloser(strings.NewReader(f.body))}, nil
}

// baueRecord baut den Prozess im Modus record über frischen Daten mit dem Fake als Upstream-Client und hält die Konfig des Record-Dienstes.
func baueRecord(t *testing.T, up *fakeUp) (*wurzel, *record.Konfig) {
	t.Helper()
	var rk record.Konfig
	w := baueWurzelHaken(t, modusRecord, nil, mockDaten(t), nil, func(b *bootstrap.Bauteile) {
		echt := b.NeuerRecorder
		b.NeuerUpstream = func(httpupstream.Konfig) (outbound.UpstreamClient, error) { return up, nil }
		b.NeuerRecorder = func(k record.Konfig) (inbound.RecordUseCase, error) { rk = k; return echt(k) }
	})
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("Hochfahren: %v", err)
	}
	return w, &rk
}

// captureZeile ist die Zeile des Capture-Spans als Literal aus SPEC-231 Nr. 6: Felder in der Folge timestamp, level, msg, ereignis, span_id,
// parent_id (nur mit Eltern), duration_ms und die Attribute in der Folge der Tabelle von §5 (Kennung, Status).
func captureZeile(spanID, elternID, sessionID, status string) string {
	eltern := ""
	if elternID != "" {
		eltern = `"parent_id":"` + elternID + `",`
	}
	return `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"INFO","msg":"odata.record.capture","ereignis":"span","span_id":"` + spanID +
		`",` + eltern + `"duration_ms":0.000,"recorder.session_id":"` + sessionID + `","recorder.capture_status":"` + status + `"}` + "\n"
}

// TestWurzelCaptureLiegtImGesamtspan hält C7 (SPEC-139 Nr. 13 d, SPEC-295 Nr. 3 e): durch den Composition Root steht die Zeile `odata.record.capture`
// als Literal in der Senke des einen Adapters des Prozesses, nach der Zeile des Gesamtspans, deren `span_id` ihr `parent_id` ist; ein Aufruf über
// den Handle ohne Gesamtspan trägt keinen `parent_id`; `Dropped` ist überall 0, und der Record-Dienst hält den Adapter des Prozesses.
func TestWurzelCaptureLiegtImGesamtspan(t *testing.T) {
	up := &fakeUp{header: map[string][]string{"Content-Type": {"application/json"}}, body: `{"a":1}`}
	w, rk := baueRecord(t, up)
	ctx := context.Background()
	alle, err := w.h.Sessions().List(ctx)
	var ss []model.Session
	for _, s := range alle {
		if s.Status == model.SessionOpen {
			ss = append(ss, s)
		}
	}
	if err != nil || len(ss) != 1 {
		t.Fatalf("Sessions %v, %v; erwartet die eine offene", alle, err)
	}
	ant, err := w.h.Record().Record(ctx, inbound.Anfrage{Methode: "GET", Pfad: "/Products", Header: map[string][]string{}, Laenge: -1, KorrelationsID: "cor-1"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	_, _ = io.Copy(io.Discard, ant.Body)
	_ = ant.Body.Close()
	if got := holeGET(t, w.h.Adresse(), "/Products"); got != 200 {
		t.Fatalf("GET: Status %d, erwartet 200", got)
	}
	if err := w.h.Herunterfahren(ctx); err != nil {
		t.Fatal(err)
	}
	z := zeilen(t, w.senke.String())
	if len(z) != 3 || z[1]["msg"] != "odata.gateway.request" {
		t.Fatalf("Zeilen %v; erwartet Capture ohne Eltern, Gesamtspan, Capture mit Eltern", z)
	}
	ohne := captureZeile(z[0]["span_id"].(string), "", ss[0].ID, "complete")
	mit := captureZeile(z[2]["span_id"].(string), z[1]["span_id"].(string), ss[0].ID, "complete")
	if !strings.HasPrefix(w.senke.String(), ohne) || !strings.HasSuffix(w.senke.String(), mit) {
		t.Errorf("Senke\n%s\nerwartet Anfang\n%s\nund Ende\n%s", w.senke.String(), ohne, mit)
	}
	if len(w.adapter) != 1 || rk.Telemetrie != support.Telemetry(w.adapter[0]) {
		t.Errorf("%d Adapter der Telemetrie; der Record-Dienst hält nicht den Adapter des Prozesses", len(w.adapter))
	}
	keinVerwurf(t, "record", w.adapter[0])
}

// TestWurzelKeinRequestWertInDerSenke hält C9 (SPEC-295 Nr. 8) an der Senke des echten Adapters: ein Canary, der die Formen Kennung und Token erfüllt
// (nur Kleinbuchstaben, bei `canary`), liegt an einer der neun Stellen (Pfad, Name und Wert der Query, Name und Wert eines Headers, Body des
// Requests, Header und Body der Antwort, Text der Ursache von Senden), in beiden Arten (angehängt, ohne Antwort), gleich 18 Paare; die Bytes der
// Senke tragen ihn nicht, und die Zeile des Spans steht dort. Was die Form verletzt (Pfad, Body), verwirft der Adapter selbst; die Stellen mit einem
// bloßen Wort (Query, Header, Ursache) hält allein der Dienst fern.
func TestWurzelKeinRequestWertInDerSenke(t *testing.T) {
	const k = "canary"
	stellen := []struct {
		name  string
		setze func(a *inbound.Anfrage, up *fakeUp)
	}{
		{"Pfad", func(a *inbound.Anfrage, _ *fakeUp) { a.Pfad = "/" + k }},
		{"Name der Query", func(a *inbound.Anfrage, _ *fakeUp) { a.RawQuery = k + "=1" }},
		{"Wert der Query", func(a *inbound.Anfrage, _ *fakeUp) { a.RawQuery = "q=" + k }},
		{"Name des Headers", func(a *inbound.Anfrage, _ *fakeUp) { a.Header["X-"+k] = []string{"1"} }},
		{"Wert des Headers", func(a *inbound.Anfrage, _ *fakeUp) { a.Header["X-Info"] = []string{k} }},
		{"Body des Requests", func(a *inbound.Anfrage, _ *fakeUp) {
			a.Body, a.Laenge = strings.NewReader(`{"x":"`+k+`"}`), int64(len(`{"x":"`+k+`"}`))
		}},
		{"Header der Antwort", func(_ *inbound.Anfrage, up *fakeUp) { up.header["X-Out"] = []string{k} }},
		{"Body der Antwort", func(_ *inbound.Anfrage, up *fakeUp) { up.body = `{"y":"` + k + `"}` }},
		{"Text der Ursache von Senden", func(_ *inbound.Anfrage, up *fakeUp) { up.fehler = errors.New(k) }},
	}
	n := 0
	for _, s := range stellen {
		for art := 0; art < 2; art++ {
			n++
			up := &fakeUp{header: map[string][]string{"Content-Type": {"application/json"}}, body: `{"a":1}`}
			a := inbound.Anfrage{Methode: "POST", Pfad: "/Products", Header: map[string][]string{"Content-Type": {"application/json"}}, Laenge: -1, KorrelationsID: "cor-1"}
			s.setze(&a, up)
			if art == 0 {
				up.fehler = nil
			} else if up.fehler == nil {
				up.fehler = outbound.ErrUpstreamNichtErreichbar
			}
			w, _ := baueRecord(t, up)
			ant, err := w.h.Record().Record(context.Background(), a)
			if (art == 1) != (err != nil) {
				t.Fatalf("%s, Art %d: Fehler %v", s.name, art, err)
			}
			if err == nil {
				_, _ = io.Copy(io.Discard, ant.Body)
				_ = ant.Body.Close()
			}
			if err := w.h.Herunterfahren(context.Background()); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(w.senke.String(), `"msg":"odata.record.capture"`) || strings.Contains(w.senke.String(), k) {
				t.Errorf("%s, Art %d: Senke %q; erwartet die Zeile des Capture-Spans ohne den Canary", s.name, art, w.senke.String())
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Paare, erwartet 18", n)
	}
}

// konfigs hält, was der Root den Konstruktoren übergab, und die Adapter der Telemetrie.
type konfigs struct {
	wartung  maintenance.Konfig
	recorder *record.Konfig
	session  *session.Konfig
	adapter  []*telemetry.Adapter
}

// baueMitSenke baut den Prozess im Modus über den Daten in dir mit eigener Senke und eigenem Logger für Build; die Uhr des Adapters ist fest.
func baueMitSenke(t *testing.T, modus, dir string) (*konfigs, *syncPuffer, *syncPuffer) {
	t.Helper()
	k, senke, log := &konfigs{}, &syncPuffer{}, &syncPuffer{}
	b := bootstrap.StandardBauteile()
	tel, wartung, rec, ses := b.NeueTelemetrie, b.NeueWartung, b.NeuerRecorder, b.NeueSession
	b.NeueTelemetrie = func(c telemetry.Config) (*telemetry.Adapter, error) {
		c.Clock = supporttest.NewFestUhr(wsT0())
		a, err := tel(c)
		k.adapter = append(k.adapter, a)
		return a, err
	}
	b.NeueWartung = func(c maintenance.Konfig) (inbound.MaintenanceUseCase, error) { k.wartung = c; return wartung(c) }
	b.NeuerRecorder = func(c record.Konfig) (inbound.RecordUseCase, error) { k.recorder = &c; return rec(c) }
	b.NeueSession = func(c session.Konfig) (inbound.SessionUseCase, error) { k.session = &c; return ses(c) }
	h, err := bootstrap.BuildMitSenke(context.Background(), wurzelKonfig(modus, dir), slog.New(slog.NewTextHandler(log, nil)), senke, b)
	if err != nil {
		t.Fatalf("Build im Modus %s: %v", modus, err)
	}
	t.Cleanup(func() { _ = h.Herunterfahren(context.Background()) })
	return k, senke, log
}

// TestWurzelLogDerDiensteIstDerHandler hält E2 (SPEC-139 Nr. 13 d, SPEC-231 Nr. 7): Menge M-Verdrahtung, 4 Stellen (record, session, maintenance,
// audit), maintenance und audit in 4 Modi, record und session im Modus record, gleich 10 Fälle: ein Eintrag über das Log der Stelle steht als
// JSON-Zeile in der Senke, der Logger von Build bleibt leer, und `Dropped` ist 0. Der Audit-Adapter hält sein Log nicht in einer Konfig: sein
// Eintrag ist der eines abgewiesenen Append.
func TestWurzelLogDerDiensteIstDerHandler(t *testing.T) {
	zeile := `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"ERROR","msg":"x","ereignis":"e"}` + "\n"
	audit := `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"ERROR","msg":"audit: Eintrag nicht geschrieben",` +
		`"error_code":"ODR-E9001","ereignis":"audit_fehlgeschlagen"}` + "\n"
	type stelle struct {
		name, soll string
		lauf       func(k *konfigs)
	}
	eintrag := func(l func(k *konfigs) *slog.Logger) func(k *konfigs) {
		return func(k *konfigs) {
			l(k).LogAttrs(context.Background(), slog.LevelError, "x", slog.String("ereignis", "e"))
		}
	}
	n := 0
	for _, modus := range []string{modusRecord, modusReplay, modusMock, modusHybrid} {
		stellen := []stelle{
			{"maintenance", zeile, eintrag(func(k *konfigs) *slog.Logger { return k.wartung.Log })},
			{"audit", audit, func(k *konfigs) { _ = k.wartung.Audit.Append(support.Eintrag{Art: "keine-art"}) }},
		}
		if modus == modusRecord {
			stellen = append(stellen,
				stelle{"record", zeile, eintrag(func(k *konfigs) *slog.Logger { return k.recorder.Log })},
				stelle{"session", zeile, eintrag(func(k *konfigs) *slog.Logger { return k.session.Log })})
		}
		for _, s := range stellen {
			n++
			k, senke, log := baueMitSenke(t, modus, mockDaten(t))
			s.lauf(k)
			if senke.String() != s.soll || log.String() != "" {
				t.Errorf("%s %s: Senke %q, Logger von Build %q; erwartet %q und leer", modus, s.name, senke.String(), log.String(), s.soll)
			}
			keinVerwurf(t, modus+" "+s.name, k.adapter[0])
		}
	}
	if n != 10 {
		t.Errorf("%d Fälle, erwartet 10", n)
	}
}

// TestWurzelAuditFehlgeschlagenInDerSenke hält E4 (SPEC-231 Nr. 4, SPEC-235 Nr. 5): ist audit.jsonl ein Verzeichnis, steht der Eintrag
// `audit_fehlgeschlagen` eines gültigen, gescheiterten Append als Zeile in der Senke, mit Text, Stufe und Feldern der Stelle (Literal), ohne
// Verwurf, ohne Fehlertext und ohne Pfad, in jedem Modus; der Logger von Build bleibt leer.
func TestWurzelAuditFehlgeschlagenInDerSenke(t *testing.T) {
	soll := `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"ERROR","msg":"audit: Eintrag nicht geschrieben","service_id":"svc",` +
		`"session_id":"ses-1","error_code":"ODR-E9001","ereignis":"audit_fehlgeschlagen","grund":"session_geloescht"}` + "\n"
	for _, modus := range []string{modusRecord, modusReplay, modusMock, modusHybrid} {
		dir := mockDaten(t)
		if err := os.Mkdir(filepath.Join(dir, "audit.jsonl"), 0o700); err != nil {
			t.Fatal(err)
		}
		k, senke, log := baueMitSenke(t, modus, dir)
		if err := k.wartung.Audit.Append(support.Eintrag{Art: support.ArtSessionGeloescht, ServiceID: "svc", SessionID: "ses-1"}); err == nil {
			t.Fatalf("%s: Append in ein Verzeichnis ohne Fehler", modus)
		}
		if senke.String() != soll || log.String() != "" || strings.Contains(senke.String(), dir) {
			t.Errorf("%s: Senke %q, Logger von Build %q; erwartet %q und leer", modus, senke.String(), log.String(), soll)
		}
		keinVerwurf(t, modus, k.adapter[0])
	}
}
