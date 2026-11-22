package bootstrap_test

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
)

// replayDaten ergänzt die Daten des Mock (mockDaten: Dienst svc mit Root /, Workspace ws-1) um die abgeschlossene Recording-Session ses-r mit
// einem Exchange für GET /Products (Key aus den Regeln der Konfiguration von basisKonfig); der Pfad /Orders ist nirgends aufgezeichnet.
func replayDaten(t *testing.T) string {
	t.Helper()
	dir := mockDaten(t)
	u := abOeffne(t, filepath.Join(dir, dateiDB), filepath.Join(dir, verzeichnisB))
	ctx := context.Background()
	abMust(t, u.db.Sessions().Create(ctx, model.Session{ID: "ses-r", ServiceID: "svc", CreatedAt: wsT0(), Status: model.SessionOpen}))
	s, err := model.BildeSchluessel(model.SchluesselEingabe{Method: "GET", ServiceID: "svc", Path: "/Products", RequestRedaction: model.RedactionNone},
		model.SchluesselRegeln{InvalidEncoding: model.KodierungLiteral, BodyMode: model.BodyBytes})
	abMust(t, err)
	abMust(t, u.db.Recording().Append(ctx, model.Exchange{
		ID: "exc-r", SessionID: "ses-r", SequenceNo: 1, Method: "GET", Path: "/Products", MatchKey: s.MatchKey, StartedAt: wsT0(), FinishedAt: wsT0(),
		RequestHeaders: map[string][]string{}, RequestBody: model.BodyAbsent, ResponseStatus: 200,
		ResponseHeaders: map[string][]string{"Content-Type": {"text/plain"}}, ResponseBody: model.BodyAbsent, Capture: model.CaptureComplete,
		RequestRedact: model.RedactionNone, ResponseRedact: model.RedactionNone, Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte,
	}))
	abMust(t, u.db.Sessions().SetStatus(ctx, "ses-r", model.SessionComplete))
	u.schliesse()
	return dir
}

// replayRuf sendet GET an pfad mit den drei Steuerheadern des Replay (Workspace ws-1, Recording-Session ses-r, Replay-Session rep-1); aendere
// setzt davon abweichende Werte. Es liefert den Status.
func replayRuf(t *testing.T, w *wurzel, pfad string, aendere func(*http.Request)) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://"+w.h.Adresse()+pfad, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Recorder-Workspace", "ws-1")
	req.Header.Set("X-Recorder-Recording-Session", "ses-r")
	req.Header.Set("X-Recorder-Replay-Session", "rep-1")
	if aendere != nil {
		aendere(req)
	}
	resp, err := gwClient().Do(req)
	if err != nil {
		t.Fatalf("%s: %v", pfad, err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp.StatusCode
}

// replayZeile ist die Zeile des Match-Spans als Literal aus SPEC-231 Nr. 6: Felder in der Folge timestamp, level, msg, ereignis, span_id,
// parent_id, duration_ms und die Attribute in der Folge der Tabelle von §5 (Strategie, Ergebnis).
func replayZeile(spanID, elternID, strategie, ergebnis string) string {
	return `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"INFO","msg":"odata.replay.match","ereignis":"span","span_id":"` + spanID +
		`","parent_id":"` + elternID + `","duration_ms":0.000,"recorder.match_strategy":"` + strategie + `","recorder.match_result":"` + ergebnis + `"}` + "\n"
}

// TestWurzelReplayMatchLiegtImGesamtspan hält W10 (SPEC-139 Nr. 13 c, SPEC-296 Nr. 3 e, 7 e): durch den Composition Root steht die Zeile
// `odata.replay.match` als Literal in der Senke des einen Adapters des Prozesses, vor der Zeile des Gesamtspans, deren `span_id` ihr `parent_id`
// ist, und `Dropped` ist überall 0; im Modus replay trägt sie den Treffer, im Modus hybrid den Versuch mit `kein_treffer` (der Mock scheitert an der
// Coverage, der Replay findet den Key nicht, der Upstream antwortet). In beiden Modi besteht allein dieser eine Adapter, und der Replay-Dienst
// hält ihn.
func TestWurzelReplayMatchLiegtImGesamtspan(t *testing.T) {
	up := neuerZaehlUpstream(t)
	faelle := []struct {
		modus, pfad, ergebnis string
		zeilen                int
	}{
		{modusReplay, "/Products", "treffer", 2},
		{modusHybrid, "/Orders", "kein_treffer", 3},
	}
	for _, f := range faelle {
		w := baueWurzelMit(t, f.modus, nil, replayDaten(t), func(k *bootstrap.Konfiguration) {
			if f.modus == modusHybrid {
				k.Fallback = policy.Werte{Upstream: true, Hosts: []string{up.host()}, ReplayKategorien: []string{"coverage_insufficient"}}
				k.Dienste[0].UpstreamURL = up.ziel()
			}
		})
		if err := w.h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: Hochfahren: %v", f.modus, err)
		}
		if got := replayRuf(t, w, f.pfad, nil); got != 200 {
			t.Errorf("%s: Status %d, erwartet 200", f.modus, got)
		}
		z := zeilen(t, w.senke.String())
		if len(z) != f.zeilen || z[len(z)-1]["msg"] != "odata.gateway.request" {
			t.Fatalf("%s: Zeilen %v; erwartet %d, die des Gesamtspans zuletzt", f.modus, z, f.zeilen)
		}
		gesamt, match := z[len(z)-1], z[len(z)-2]
		soll := replayZeile(match["span_id"].(string), gesamt["span_id"].(string), "sequential", f.ergebnis)
		if !strings.Contains(w.senke.String(), soll) || match["msg"] != "odata.replay.match" {
			t.Errorf("%s: Senke\n%s\nenthält nicht\n%s", f.modus, w.senke.String(), soll)
		}
		if len(w.adapter) != 1 {
			t.Fatalf("%s: %d Adapter der Telemetrie, erwartet 1", f.modus, len(w.adapter))
		}
		if w.replay == nil || w.replay.Telemetrie != support.Telemetry(w.adapter[0]) {
			t.Errorf("%s: der Replay-Dienst hält nicht den Adapter des Prozesses", f.modus)
		}
		keinVerwurf(t, f.modus, w.adapter[0])
		if f.modus == modusHybrid {
			if n, _ := up.stand(); n != 1 {
				t.Errorf("hybrid: %d Verbindungen zum Upstream, erwartet 1 (der Fallback)", n)
			}
		}
	}
}

// TestWurzelReplayKeinRequestWertInDerSenke hält W12 (SPEC-296 Nr. 8) an der Senke des echten Adapters: ein Wert in der Form Token an einer der
// 7 Stellen des Requests, die ein HTTP-Request tragen kann (Pfad, Name und Wert der Query, Name und Wert eines Headers, Wert des Headers der
// Replay-Session, Body), erreicht die Senke des Prozesses nicht, weder in der Zeile des Match-Spans noch in der des Gesamtspans. Der Text der
// Ursache eines Fehlers eines Ports ist hier nicht erreichbar; ihn fängt der Test des Dienstes mit dem Fake.
func TestWurzelReplayKeinRequestWertInDerSenke(t *testing.T) {
	k := gwKanarie(false)
	stellen := map[string]func(*http.Request){
		"Query-Name":    func(r *http.Request) { r.URL.RawQuery = k + "=1" },
		"Query-Wert":    func(r *http.Request) { r.URL.RawQuery = "q=" + k },
		"Header-Name":   func(r *http.Request) { r.Header.Set("X-"+k, "v") },
		"Header-Wert":   func(r *http.Request) { r.Header.Set("X-Wert", k) },
		"Replay-Header": func(r *http.Request) { r.Header.Set("X-Recorder-Replay-Session", k) },
		"Body":          func(r *http.Request) { r.Body, r.ContentLength = io.NopCloser(strings.NewReader(k)), int64(len(k)) },
	}
	n := 0
	for name, aendere := range stellen {
		n++
		w := baueWurzelMit(t, modusReplay, nil, replayDaten(t), nil)
		if err := w.h.Hochfahren(context.Background()); err != nil {
			t.Fatalf("%s: Hochfahren: %v", name, err)
		}
		replayRuf(t, w, "/Products", aendere)
		if s := w.senke.String(); !strings.Contains(s, "odata.replay.match") || strings.Contains(strings.ToLower(s), k) {
			t.Errorf("%s: Senke %q: ohne Match-Span oder mit dem Canary", name, s)
		}
		keinVerwurf(t, name, w.adapter[0])
	}
	w := baueWurzelMit(t, modusReplay, nil, replayDaten(t), nil)
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatal(err)
	}
	replayRuf(t, w, "/Products/"+k, nil)
	n++
	if s := w.senke.String(); !strings.Contains(s, "odata.replay.match") || strings.Contains(strings.ToLower(s), k) {
		t.Errorf("Pfad: Senke %q: ohne Match-Span oder mit dem Canary", s)
	}
	if n != 7 {
		t.Errorf("Menge: %d Stellen, erwartet 7", n)
	}
}
