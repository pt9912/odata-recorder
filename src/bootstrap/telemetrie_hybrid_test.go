package bootstrap_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid"
	"github.com/pt9912/odata-recorder/src/hexagon/services/hybrid/policy"
	"github.com/pt9912/odata-recorder/src/hexagon/services/mock"
)

// hybridZeile ist die Zeile des Spans `odata.hybrid.fallback` als Literal aus SPEC-231 Nr. 6: Felder in der Folge timestamp, level, msg,
// ereignis, span_id, parent_id, duration_ms und die Attribute in der Folge der Tabelle von §5 (Art, Ergebnis).
func hybridZeile(spanID, elternID, art, ergebnis string) string {
	return `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"INFO","msg":"odata.hybrid.fallback","ereignis":"span","span_id":"` + spanID +
		`","parent_id":"` + elternID + `","duration_ms":0.000,"recorder.fallback_type":"` + art + `","recorder.fallback_result":"` + ergebnis + `"}` + "\n"
}

// feldText liefert den Text des Feldes einer Zeile (leer, wenn es fehlt oder kein Text ist).
func feldText(z map[string]any, feld string) string {
	s, _ := z[feld].(string)
	return s
}

// zaehlerStand liefert die Zähler der Reihe, die nicht 0 sind, je Code.
func zaehlerStand(m map[model.Code]uint64) map[model.Code]uint64 {
	aus := map[model.Code]uint64{}
	for c, n := range m {
		if n != 0 {
			aus[c] = n
		}
	}
	return aus
}

// pruefeHybridKette hält W-D12 an den Zeilen der Kette Mock-Query, Replay-Match, Fallback `replay` mit `fehler`, Fallback `upstream` mit
// `antwort` und Gesamtspan: die vier Spans haben den Gesamtspan zum Elternspan, die zwei Zeilen des Fallbacks stehen als Literal in der
// Senke, `Dropped` ist überall 0, `Verworfen` zählt ODR-E6003 und ODR-E5001 je einmal und `Errors` nichts (SPEC-321 Nr. 5, SPEC-139 Nr. 13 g).
func pruefeHybridKette(t *testing.T, w *wurzel, z []map[string]any) {
	t.Helper()
	namen := []string{"odata.mock.query", "odata.replay.match", "odata.hybrid.fallback", "odata.hybrid.fallback", "odata.gateway.request"}
	if len(z) != len(namen) {
		t.Fatalf("hybrid: %d Zeilen, erwartet %d", len(z), len(namen))
	}
	eltern := feldText(z[len(z)-1], "span_id")
	for i, n := range namen {
		if feldText(z[i], "msg") != n {
			t.Errorf("hybrid: Zeile %d ist %q, erwartet %q", i, feldText(z[i], "msg"), n)
		}
		if i < len(namen)-1 && feldText(z[i], "parent_id") != eltern {
			t.Errorf("hybrid: Zeile %d hat parent_id %q, erwartet die span_id des Gesamtspans %q", i, feldText(z[i], "parent_id"), eltern)
		}
	}
	for _, soll := range []string{hybridZeile(feldText(z[2], "span_id"), eltern, "replay", "fehler"), hybridZeile(feldText(z[3], "span_id"), eltern, "upstream", "antwort")} {
		if !strings.Contains(w.senke.String(), soll) {
			t.Errorf("hybrid: Senke\n%s\nenthält nicht\n%s", w.senke.String(), soll)
		}
	}
	st := w.adapter[0].Snapshot()
	want := map[model.Code]uint64{model.CodeE6003: 1, model.CodeE5001: 1}
	if got := zaehlerStand(st.Verworfen); len(got) != 2 || got[model.CodeE6003] != 1 || got[model.CodeE5001] != 1 {
		t.Errorf("hybrid: Verworfen %v, erwartet %v", got, want)
	}
	if got := zaehlerStand(st.Errors); len(got) != 0 {
		t.Errorf("hybrid: Errors %v, erwartet keiner (der Fallback ersetzt den Fehler, der Client erhielt eine Antwort)", got)
	}
}

// TestWurzelHybridUpstreamNichtErreichbar hält W-D12 (zweiter Fall): scheitert der Upstream nach dem Replay-Miss (nicht erreichbar, 502), zählt der
// Fehler genau in einer der beiden Reihen: `Errors` trägt ODR-E4001 einmal, `ODR-E6003` und `ODR-E5001` nirgends in `Errors`, und `Verworfen`
// trägt beide je einmal (SPEC-321 Nr. 5, SPEC-294 Nr. 5).
func TestWurzelHybridUpstreamNichtErreichbar(t *testing.T) {
	tot := httptest.NewServer(http.NotFoundHandler())
	host, ziel := tot.Listener.Addr().String(), tot.URL+"/odata"
	tot.Close()
	w := baueWurzelMit(t, modusHybrid, nil, replayDaten(t), func(k *bootstrap.Konfiguration) {
		k.Fallback = policy.Werte{Upstream: true, Hosts: []string{host}, ReplayKategorien: []string{"coverage_insufficient"}}
		k.Dienste[0].UpstreamURL = ziel
	})
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("Hochfahren: %v", err)
	}
	if got := replayRuf(t, w, "/Orders", nil); got != 502 {
		t.Errorf("Status %d, erwartet 502", got)
	}
	st := w.adapter[0].Snapshot()
	if got := zaehlerStand(st.Errors); len(got) != 1 || got[model.CodeE4001] != 1 {
		t.Errorf("Errors %v, erwartet allein ODR-E4001 einmal", got)
	}
	if got := zaehlerStand(st.Verworfen); len(got) != 2 || got[model.CodeE6003] != 1 || got[model.CodeE5001] != 1 {
		t.Errorf("Verworfen %v, erwartet ODR-E6003 und ODR-E5001 je einmal", got)
	}
	if !strings.Contains(w.senke.String(), `"recorder.fallback_type":"upstream","recorder.fallback_result":"fehler"`) {
		t.Errorf("Senke ohne die Zeile upstream mit fehler: %s", w.senke.String())
	}
	keinVerwurf(t, "hybrid 502", w.adapter[0])
}

// TestWurzelHybridHaeltDenAdapterDesProzesses hält W-D12 (SPEC-139 Nr. 13 g): der Hybrid-Dienst erhält den einen Adapter des Prozesses als
// Telemetrie (Identität); in den Modi record, replay und mock wird kein Hybrid-Dienst gebaut und kein Zähler der Reihe gesetzt (3 Modi).
func TestWurzelHybridHaeltDenAdapterDesProzesses(t *testing.T) {
	var gesehen []hybrid.Konfig
	haken := func(b *bootstrap.Bauteile) {
		vorher := b.NeuerHybrid
		b.NeuerHybrid = func(k hybrid.Konfig) (inbound.HybridUseCase, error) {
			gesehen = append(gesehen, k)
			return vorher(k)
		}
	}
	for _, modus := range []string{modusRecord, modusReplay, modusMock} {
		w := baueWurzelHaken(t, modus, nil, replayDaten(t), nil, haken)
		if len(gesehen) != 0 || len(zaehlerStand(w.adapter[0].Snapshot().Verworfen)) != 0 {
			t.Errorf("%s: %d Hybrid-Dienste gebaut, Verworfen %v; erwartet keiner", modus, len(gesehen), w.adapter[0].Snapshot().Verworfen)
		}
	}
	w := baueWurzelHaken(t, modusHybrid, nil, replayDaten(t), nil, haken)
	if len(gesehen) != 1 || len(w.adapter) != 1 {
		t.Fatalf("hybrid: %d Hybrid-Dienste, %d Adapter der Telemetrie; erwartet je 1", len(gesehen), len(w.adapter))
	}
	if gesehen[0].Telemetrie != support.FallbackTelemetry(w.adapter[0]) {
		t.Errorf("hybrid: der Hybrid-Dienst hält nicht den Adapter des Prozesses")
	}
}

// hybridMitFehlerMock baut den Prozess im Modus hybrid über replayDaten: der Mock antwortet jede Anfrage mit ODR-E6003, die Policy gibt nur den
// Upstream frei (Host des Referenzdienstes), und der eine Dienst unter / hat dessen Adresse als Ziel. aendere setzt weitere Werte.
func hybridMitFehlerMock(t *testing.T, up *httptest.Server, aendere func(*bootstrap.Konfiguration)) *wurzel {
	t.Helper()
	w := baueWurzelHaken(t, modusHybrid, nil, replayDaten(t), func(k *bootstrap.Konfiguration) {
		k.Fallback = policy.Werte{Upstream: true, Hosts: []string{up.Listener.Addr().String()}}
		k.Dienste[0].UpstreamURL = up.URL + "/odata"
		if aendere != nil {
			aendere(k)
		}
	}, func(b *bootstrap.Bauteile) {
		b.NeuerMock = func(mock.Konfig) (inbound.MockUseCase, error) {
			return &fehlerMock{err: mockFehler(model.CodeE6003, 501)}, nil
		}
	})
	if err := w.h.Hochfahren(context.Background()); err != nil {
		t.Fatalf("Hochfahren: %v", err)
	}
	return w
}

// TestWurzelHybridHerunterfahrenBeendetDenSpan hält D18 (SPEC-321 Nr. 7 h, SPEC-139 Nr. 12 c): hält der Upstream die Antwort und läuft die Frist
// von Schliessen ab, steht nach der Rückkehr von Herunterfahren die Zeile `odata.hybrid.fallback` mit upstream und abgebrochen in der Senke, vor
// der Zeile des Gesamtspans und mit deren span_id als parent_id.
func TestWurzelHybridHerunterfahrenBeendetDenSpan(t *testing.T) {
	angekommen, frei := make(chan struct{}, 1), make(chan struct{})
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		angekommen <- struct{}{}
		select {
		case <-frei:
		case <-r.Context().Done():
		case <-time.After(20 * time.Second):
		}
		w.WriteHeader(200)
	}))
	defer up.Close()
	defer close(frei)
	w := hybridMitFehlerMock(t, up, func(k *bootstrap.Konfiguration) { k.ShutdownTimeout = 300 * time.Millisecond })
	fertig := make(chan int, 1)
	go func() { fertig <- gwRuf(t, w.h.Adresse(), "GET", "/Orders") }()
	select {
	case <-angekommen:
	case <-time.After(20 * time.Second):
		t.Fatal("der Upstream erhielt keine Anfrage")
	}
	if err := w.h.Herunterfahren(context.Background()); !errors.Is(err, httpodata.ErrFristAbgelaufen) {
		t.Errorf("Herunterfahren: %v, erwartet ErrFristAbgelaufen (die Anfrage hängt am Upstream)", err)
	}
	z := zeilen(t, w.senke.String())
	if len(z) != 2 || feldText(z[1], "msg") != "odata.gateway.request" {
		t.Fatalf("Zeilen %v; erwartet die des Fallbacks, dann die des Gesamtspans", z)
	}
	soll := hybridZeile(feldText(z[0], "span_id"), feldText(z[1], "span_id"), "upstream", "abgebrochen")
	if feldText(z[0], "msg") != "odata.hybrid.fallback" || !strings.Contains(w.senke.String(), soll) {
		t.Errorf("Senke\n%s\nenthält nicht\n%s", w.senke.String(), soll)
	}
	select {
	case <-fertig:
	case <-time.After(20 * time.Second):
		t.Error("der Client erhielt keine Rückkehr")
	}
}

// TestWurzelHybridKeinWertInDerSenke hält D13 an der Senke des echten Adapters: ein Wert in der Form Token an einer der 9 Stellen, die ein HTTP-Request
// oder die Antwort des Upstream tragen kann (Pfad, Name und Wert der Query, Name und Wert eines Headers der Anfrage, Body der Anfrage, Name und
// Wert eines Headers der Antwort, Body der Antwort), erreicht die Senke nicht; die Zeile des Fallbacks steht dort. Die übrigen 5 Stellen
// (Ziel, Texte der Ursachen) führt der Test des Dienstes.
func TestWurzelHybridKeinWertInDerSenke(t *testing.T) {
	k := gwKanarie(false)
	antwort := func(name, wert, body string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(name, wert)
			_, _ = io.WriteString(w, body)
		}
	}
	stellen := []struct {
		name  string
		frage func(*http.Request)
		up    http.HandlerFunc
	}{
		{"Pfad", func(r *http.Request) { r.URL.Path += "/" + k }, nil},
		{"Query-Name", func(r *http.Request) { r.URL.RawQuery = k + "=1" }, nil},
		{"Query-Wert", func(r *http.Request) { r.URL.RawQuery = "q=" + k }, nil},
		{"Header-Name", func(r *http.Request) { r.Header.Set("X-"+k, "v") }, nil},
		{"Header-Wert", func(r *http.Request) { r.Header.Set("X-Wert", k) }, nil},
		{"Body", func(r *http.Request) { r.Body, r.ContentLength = io.NopCloser(strings.NewReader(k)), int64(len(k)) }, nil},
		{"Antwort-Header-Name", nil, antwort("X-"+k, "v", "u")},
		{"Antwort-Header-Wert", nil, antwort("X-Wert", k, "u")},
		{"Antwort-Body", nil, antwort("X-U", "v", k)},
	}
	for _, s := range stellen {
		up := httptest.NewServer(handlerOder(s.up, antwort("X-U", "v", "u")))
		w := hybridMitFehlerMock(t, up, nil)
		replayRuf(t, w, "/Orders", s.frage)
		if got := w.senke.String(); !strings.Contains(got, "odata.hybrid.fallback") || strings.Contains(strings.ToLower(got), k) {
			t.Errorf("%s: Senke %q: ohne Zeile des Fallbacks oder mit dem Canary", s.name, got)
		}
		keinVerwurf(t, s.name, w.adapter[0])
		up.Close()
	}
	if len(stellen) != 9 {
		t.Errorf("Menge: %d Stellen, erwartet 9", len(stellen))
	}
}

// handlerOder liefert a, wenn es gesetzt ist, sonst b.
func handlerOder(a, b http.HandlerFunc) http.HandlerFunc {
	if a != nil {
		return a
	}
	return b
}
