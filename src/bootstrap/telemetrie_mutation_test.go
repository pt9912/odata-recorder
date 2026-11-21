package bootstrap_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// Die Tests dieser Datei halten den Span `odata.workspace.mutation` durch den Composition Root mit dem echten Adapter der Telemetrie
// (SPEC-290 Nr. 3 e, 4 c und e, 8; Zeilen M5, M7, M8, M9 des Plans). Die Zeilen der Senke stehen als Literal im Test.

// mutAnfrage ist eine Mutation an den Dienst svc (Root /) im Workspace ws-1; body leer heißt ohne Body; kopf ersetzt und ergänzt Header.
func mutAnfrage(methode, pfad, body string, kopf map[string][]string) inbound.Anfrage {
	h := map[string][]string{"X-Recorder-Workspace": {"ws-1"}}
	if body != "" {
		h["Content-Type"] = []string{"application/json"}
	}
	for k, v := range kopf {
		h[k] = v
	}
	a := inbound.Anfrage{Methode: methode, Pfad: pfad, Header: h, Laenge: -1}
	if body != "" {
		a.Body, a.Laenge = io.NopCloser(strings.NewReader(body)), int64(len(body))
	}
	return a
}

// mutiere ruft den Mock mit ctx und liest den Body einer Antwort bis zum Ende.
func (w *wurzel) mutiere(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	antwort, err := w.mock.Mock(ctx, a)
	if err == nil {
		_, _ = io.Copy(io.Discard, antwort.Body)
		_ = antwort.Body.Close()
	}
	return antwort, err
}

// mutSchritt ist eine Mutation: Methode, Pfad, Body, Art und Status der Antwort.
type mutSchritt struct {
	methode, pfad, body, art string
	status                   int
}

// mutFolge sind die vier Mutationen an das Produkt 10 (POST legt es an).
func mutFolge() []mutSchritt {
	return []mutSchritt{
		{"POST", "/Products", `{"Id":10,"Name":"neu"}`, "post", 201},
		{"PATCH", "/Products(10)", `{"Name":"x"}`, "patch", 204},
		{"PUT", "/Products(10)", `{"Id":10,"Name":"y"}`, "put", 204},
		{"DELETE", "/Products(10)", "", "delete", 204},
	}
}

// mutZeile ist die Zeile des Spans einer Mutation als Literal (SPEC-231 Nr. 6): n ist die Nummer des Spans im Adapter des Prozesses.
func mutZeile(n int, ws, art string) string {
	return `{"timestamp":"2026-10-20T10:00:00.000000000Z","level":"INFO","msg":"odata.workspace.mutation","ereignis":"span",` +
		`"span_id":"spn-` + sechzehnstellig(n) + `","duration_ms":0.000,"recorder.workspace_id":"` + ws + `","recorder.mutation_type":"` + art + `"}` + "\n"
}

// sechzehnstellig ist n mit führenden Nullen auf 16 Stellen (die Form der Kennung des Spans).
func sechzehnstellig(n int) string {
	s := strconv.Itoa(n)
	return strings.Repeat("0", 16-len(s)) + s
}

// TestWurzelMutationSpanZeile hält M5 und M1 durch den Root (SPEC-139 Nr. 13 a, d; SPEC-290 Nr. 2 b): in den Modi mock und hybrid schreibt jede der vier
// Mutationen eine Zeile mit `ereignis` `span`, der Kennung des Workspace und der Art, in dieser Folge und Form (Literal); kein Attribut wird verworfen.
// Modus 2 × Methode 4 = 8 Zeilen.
func TestWurzelMutationSpanZeile(t *testing.T) {
	n := 0
	for _, modus := range []string{modusMock, modusHybrid} {
		w := baueWurzel(t, modus, nil)
		var soll string
		for i, m := range mutFolge() {
			n++
			_, err := w.mutiere(context.Background(), mutAnfrage(m.methode, m.pfad, m.body, nil))
			if err != nil {
				t.Fatalf("%s %s: %v", modus, m.methode, err)
			}
			soll += mutZeile(i+1, "ws-1", m.art)
		}
		if got := w.senke.String(); got != soll {
			t.Errorf("%s: Senke\n%s\nerwartet\n%s", modus, got, soll)
		}
		for i, a := range w.adapter {
			keinVerwurf(t, modus, a)
			if want := len(mutFolge()) * (1 - min(i, 1)); len(a.Snapshot().Spans) != want {
				t.Errorf("%s: Adapter %d trägt %d Spans, erwartet %d (allein der des Prozesses)", modus, i, len(a.Snapshot().Spans), want)
			}
		}
	}
	if n != 8 {
		t.Errorf("%d Zeilen, erwartet 8", n)
	}
}

// TestWurzelMutationSpanEltern hält M7 durch den Root (SPEC-290 Nr. 3 e): der Span der Mutation ist Kind des Gesamtspans, wenn das ctx ihn trägt,
// sonst ohne Eltern; ParentID im Stand ist die ID des Elternspans, und die Zeile trägt `parent_id` gleich dessen `span_id`. Methode 4, je mit und ohne Eltern.
func TestWurzelMutationSpanEltern(t *testing.T) {
	w := baueWurzel(t, modusMock, nil)
	ad := w.adapter[0]
	for _, m := range mutFolge() {
		ctx, eltern := ad.StartSpan(context.Background(), support.SpanGatewayRequest,
			support.Str("http.request.method", m.methode), support.Int("http.response.status_code", int64(m.status)), support.Str("recorder.mode", "mock"))
		if _, err := w.mutiere(ctx, mutAnfrage(m.methode, m.pfad, m.body, nil)); err != nil {
			t.Fatalf("%s mit Eltern: %v", m.methode, err)
		}
		eltern.End()
	}
	var kinder, gesamt []support.SpanRecord
	for _, r := range ad.Snapshot().Spans {
		if r.Name == support.SpanWorkspaceMutation {
			kinder = append(kinder, r)
		} else {
			gesamt = append(gesamt, r)
		}
	}
	if len(kinder) != 4 || len(gesamt) != 4 {
		t.Fatalf("Spans %+v %+v; erwartet 4 Kinder und 4 Elternspans", kinder, gesamt)
	}
	z := zeilen(t, w.senke.String())
	if len(z) != 8 {
		t.Fatalf("%d Zeilen, erwartet 8", len(z))
	}
	for i, k := range kinder {
		if k.ParentID == "" || k.ParentID != gesamt[i].ID {
			t.Errorf("Mutation %d: ParentID %q, erwartet die ID %q des Elternspans", i, k.ParentID, gesamt[i].ID)
		}
	}
	for i := range 4 {
		kind, eltern := z[2*i], z[2*i+1]
		if kind["msg"] != "odata.workspace.mutation" || kind["parent_id"] != eltern["span_id"] || kind["parent_id"] == "" {
			t.Errorf("Zeilen %v %v: erwartet parent_id des Kindes gleich span_id des Elternspans", kind, eltern)
		}
	}
	// ohne Eltern: die Zeile trägt kein parent_id
	w = baueWurzel(t, modusMock, nil)
	if _, err := w.mutiere(context.Background(), mutAnfrage("POST", "/Products", `{"Id":10,"Name":"neu"}`, nil)); err != nil {
		t.Fatal(err)
	}
	z = zeilen(t, w.senke.String())
	if len(z) != 1 {
		t.Fatalf("%d Zeilen, erwartet 1", len(z))
	}
	if _, da := z[0]["parent_id"]; da {
		t.Errorf("die Zeile ohne Eltern trägt parent_id: %v", z[0])
	}
	if rs := w.adapter[0].Snapshot().Spans; len(rs) != 1 || rs[0].ParentID != "" {
		t.Errorf("Spans %+v; erwartet einen ohne Eltern", rs)
	}
}

// mutDatenMitWorkspace legt die Daten von mockDaten an und dazu einen Workspace mit der Kennung id, die die Form des Adapters verletzen darf
// (ein Fremdbestand, SPEC-290 Nr. 4 e); die Datei ist danach geschlossen.
func mutDatenMitWorkspace(t *testing.T, id string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, verzeichnisB), 0o700); err != nil {
		t.Fatal(err)
	}
	u := abOeffne(t, filepath.Join(dir, dateiDB), filepath.Join(dir, verzeichnisB))
	u.ausgangsdaten()
	abMust(t, u.db.Workspaces().Create(context.Background(), model.Workspace{ID: id, ServiceID: "svc", BaselineSnapshotID: "bsl-1", CreatedAt: wsT0()}))
	u.schliesse()
	return dir
}

// TestWurzelMutationKennungAusserhalbDerForm hält M5 (SPEC-290 Nr. 4 e): ist die Kennung des gelesenen Workspace außerhalb der Form `Kennung`
// (ein Leerzeichen), zählt der Adapter je Mutation `wert` einmal und `pflicht` einmal und lässt das Attribut weg; die Antwort bleibt unberührt
// (Status der Mutation), die Zeile trägt die Kennung nicht. Methode 4, die Zähler wachsen je Mutation um 1.
func TestWurzelMutationKennungAusserhalbDerForm(t *testing.T) {
	const id = "ws mit Leerzeichen"
	w := baueWurzelMit(t, modusMock, nil, mutDatenMitWorkspace(t, id), nil)
	ad := w.adapter[0]
	for i, m := range mutFolge() {
		a := mutAnfrage(m.methode, m.pfad, m.body, map[string][]string{"X-Recorder-Workspace": {id}})
		antwort, err := w.mutiere(context.Background(), a)
		if err != nil || antwort.Status != m.status {
			t.Fatalf("%s: Status %d, Fehler %v; erwartet %d", m.methode, antwort.Status, err, m.status)
		}
		st := ad.Snapshot()
		if st.Dropped[support.VerwurfWert] != uint64(i+1) || st.Dropped[support.VerwurfPflicht] != uint64(i+1) {
			t.Errorf("%s: Verwürfe wert %d pflicht %d, erwartet je %d", m.methode, st.Dropped[support.VerwurfWert], st.Dropped[support.VerwurfPflicht], i+1)
		}
		if n := len(st.Spans); n != i+1 {
			t.Errorf("%s: %d Spans im Stand, erwartet %d (der Span wird aufgezeichnet)", m.methode, n, i+1)
		}
	}
	if strings.Contains(w.senke.String(), "Leerzeichen") {
		t.Errorf("die Senke trägt die Kennung: %s", w.senke.String())
	}
}

// wurzelStelle ist eine Stelle mit Request-Wert der Mutation durch den Root (M8): der Aufbau trägt den wohlgeformten Wert kanarieWurzel an der
// Stelle in den Request; zeilen sagt, wie viele Zeilen des Spans ein Request liefert (0 beim unbekannten Workspace).
type wurzelStelle struct {
	name     string
	methoden []string
	zeilen   int
	etag     bool
	aufbau   func(methode string) inbound.Anfrage
}

func wurzelStellen() []wurzelStelle {
	k := kanarieWurzel
	alle := []string{"POST", "PATCH", "PUT", "DELETE"}
	mitBody := []string{"POST", "PATCH", "PUT"}
	ohnePost := []string{"PATCH", "PUT", "DELETE"}
	pfad := func(methode string) string {
		if methode == "POST" {
			return "/Products"
		}
		return "/Products(1)"
	}
	// body ist der Body der Methode mit dem Mitglied m (DELETE ohne Body); der Schlüssel des Bodys gleicht dem der URL.
	body := func(methode, m string) string {
		switch methode {
		case "POST":
			return `{"Id":50,` + m + `}`
		case "PATCH":
			return `{` + m + `}`
		case "PUT":
			return `{"Id":1,` + m + `}`
		}
		return ""
	}
	standard := func(methode string, kopf map[string][]string) inbound.Anfrage {
		return mutAnfrage(methode, pfad(methode), body(methode, `"Name":"x"`), kopf)
	}
	kopf := func(name, wert string, methoden []string, etag bool) wurzelStelle {
		return wurzelStelle{name, methoden, 1, etag, func(methode string) inbound.Anfrage {
			return standard(methode, map[string][]string{name: {wert}})
		}}
	}
	return []wurzelStelle{
		{"Name einer Property im Body", mitBody, 1, false, func(methode string) inbound.Anfrage {
			return mutAnfrage(methode, pfad(methode), body(methode, `"`+k+`Prop":1`), nil)
		}},
		{"Wert einer Property im Body", mitBody, 1, false, func(methode string) inbound.Anfrage {
			return mutAnfrage(methode, pfad(methode), body(methode, `"Name":"`+k+`"`), nil)
		}},
		kopf("Content-Type", k, mitBody, false),
		kopf("Prefer", k, mitBody, false),
		kopf("If-Match", k, ohnePost, true),
		kopf("If-None-Match", k, ohnePost, true),
		{"Set", alle, 1, false, func(methode string) inbound.Anfrage {
			p := "/" + k + "Set(1)"
			if methode == "POST" {
				p = "/" + k + "Set"
			}
			return mutAnfrage(methode, p, body(methode, `"Name":"x"`), nil)
		}},
		{"Schlüssel", ohnePost, 1, false, func(methode string) inbound.Anfrage {
			return mutAnfrage(methode, "/Products('"+k+"')", body(methode, `"Name":"x"`), nil)
		}},
		{"Header X-Recorder-Workspace", alle, 0, false, func(methode string) inbound.Anfrage {
			return standard(methode, map[string][]string{"X-Recorder-Workspace": {k}})
		}},
		kopf("Accept", k, alle, false),
	}
}

// TestKeinRequestWertInDerSenkeMutation hält M8 durch den Root (SPEC-290 Nr. 8): ein wohlgeformter Wert an jeder der 10 Stellen des Requests,
// die eine Mutation liest, bei jeder Methode, die sie liest, erreicht die Bytes der Senke des echten Adapters nicht, in keinem Attribut und
// in keinem Feld; der Wert besteht aus Buchstaben und Ziffern und geht durch die Form des Adapters (SPEC-232 Nr. 2, *Ungeprüft*), allein der
// Dienst hält ihn fern; ob ein Header, der als Wert nichts bedeutet (Content-Type, If-Match), den Dienst erreicht oder früher scheitert, ist
// gleich. Jeder Request liefert die Zeile seines Spans (außer beim unbekannten Workspace) und keinen Verwurf.
func TestKeinRequestWertInDerSenkeMutation(t *testing.T) {
	stellen := wurzelStellen()
	if len(stellen) != 10 {
		t.Fatalf("%d Stellen, erwartet 10", len(stellen))
	}
	paare := 0
	for _, s := range stellen {
		w := baueWurzelMit(t, modusMock, nil, mockDaten(t), func(k *bootstrap.Konfiguration) { k.EtagEnabled = s.etag })
		for _, methode := range s.methoden {
			paare++
			vorher := w.senke.String()
			_, _ = w.mutiere(context.Background(), s.aufbau(methode))
			neu := strings.TrimPrefix(w.senke.String(), vorher)
			if strings.Contains(neu, kanarieWurzel) {
				t.Errorf("%s %s: die Senke trägt den Wert des Requests: %s", s.name, methode, neu)
			}
			if got := strings.Count(neu, "\n"); got != s.zeilen {
				t.Errorf("%s %s: %d Zeilen, erwartet %d: %s", s.name, methode, got, s.zeilen, neu)
			}
		}
		keinVerwurf(t, s.name, w.adapter[0])
	}
	if paare != 3+3+3+3+3+3+4+3+4+4 {
		t.Errorf("%d Paare, erwartet 33", paare)
	}
}

// TestKanarienNameDerLaenge hält M9 (SPEC-290 Nr. 8, SPEC-180 Nr. 3; Sendung R2 von slice-mock-schreiben): der Name einer unbekannten Property von
// `grenze` Bytes durch POST, PATCH und PUT erreicht weder eine Zeile der Senke noch ein Attribut eines Spans: weder die ersten 8 Bytes des Namens
// noch ein Teil davon; die Antwort ist ODR-E6007. Die Namen entstehen aus Bytewerten (AGENTS.md §3.18). Modus 2 × Methode 3 × Länge 4 = 24.
func TestKanarienNameDerLaenge(t *testing.T) {
	laengen := []int{255, 256, 257, 4096}
	pfad := map[string]string{"POST": "/Products", "PATCH": "/Products(1)", "PUT": "/Products(1)"}
	n := 0
	for _, modus := range []string{modusMock, modusHybrid} {
		w := baueWurzel(t, modus, nil)
		for _, methode := range []string{"POST", "PATCH", "PUT"} {
			for _, l := range laengen {
				n++
				name := string(append([]byte("Kq7f"), []byte(strings.Repeat("x", l-4))...))
				m := `"` + name + `":1`
				body := map[string]string{"POST": `{"Id":60,` + m + `}`, "PATCH": `{` + m + `}`, "PUT": `{"Id":1,` + m + `}`}[methode]
				kontext := modus + " " + methode + " " + strconv.Itoa(l)
				_, err := w.mutiere(context.Background(), mutAnfrage(methode, pfad[methode], body, nil))
				if f, ok := err.(*inbound.Fehler); !ok || f.Code != model.CodeE6007 {
					t.Errorf("%s: Fehler %v, erwartet ODR-E6007", kontext, err)
				}
				if strings.Contains(w.senke.String(), name[:8]) {
					t.Errorf("%s: die Senke trägt den Anfang des Namens", kontext)
				}
				for _, r := range w.adapter[0].Snapshot().Spans {
					for _, a := range r.Attrs {
						if strings.Contains(a.Text, name[:8]) {
							t.Errorf("%s: Attribut %s trägt den Anfang des Namens", kontext, a.Key)
						}
					}
				}
			}
		}
		if z := zeilen(t, w.senke.String()); len(z) != 12 {
			t.Errorf("%s: %d Zeilen, erwartet 12 (je Anfrage die Zeile des Spans)", modus, len(z))
		}
		keinVerwurf(t, modus, w.adapter[0])
	}
	if n != 24 {
		t.Errorf("%d Läufe, erwartet 24", n)
	}
}
