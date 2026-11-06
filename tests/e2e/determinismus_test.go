package e2e_test

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// laeufe ist die Zahl der Wiederholungen (SPEC-038 Nr. 3); determinismusSession ist die Replay-Session aller Läufe.
const (
	laeufe               = 100
	determinismusSession = "determinismus"
	// platzhalterListen besteht die Prüfung von LISTEN (Port 1 bis 65535); der Test setzt danach Konfiguration.Listen
	// auf Port 0 und liest die Adresse aus Handle.Adresse.
	platzhalterListen = "127.0.0.1:8080"
)

// determinismusFolge sind die zehn Requests der Folge (SPEC-038 Nr. 4).
func determinismusFolge() []string {
	return []string{"/odata/$metadata", "/odata/Products", "/odata/Products?$skiptoken=1", "/odata/Products(1)",
		"/odata/Products(1)/Name/$value", "/odata/Products/$count", "/odata/Products/$count", "/odata/Products/$count",
		"/odata/Products/$count", "/odata/Products(99)"}
}

// kopfMenge sind die Header der Antwort ohne die Namen aus volatile, Namen kanonisiert, je Name die Werte in ihrer
// Reihenfolge (SPEC-038 Nr. 5).
func kopfMenge(h http.Header, volatile []string) map[string][]string {
	aus := map[string]bool{}
	for _, n := range volatile {
		aus[http.CanonicalHeaderKey(n)] = true
	}
	out := map[string][]string{}
	for n, v := range h {
		if k := http.CanonicalHeaderKey(n); !aus[k] {
			out[k] = append(out[k], v...)
		}
	}
	return out
}

// vergleicheAntworten nennt die erste Abweichung von ist gegen soll in der Reihenfolge Status, Body, Header,
// sonst "" (SPEC-038 Nr. 5). Ausgenommen sind allein die Namen aus volatile.
func vergleicheAntworten(soll, ist antwort, volatile []string) string {
	if soll.status != ist.status {
		return fmt.Sprintf("Status %d gegen %d", soll.status, ist.status)
	}
	if string(soll.body) != string(ist.body) {
		return fmt.Sprintf("Body %q gegen %q", soll.body, ist.body)
	}
	s, i := kopfMenge(soll.header, volatile), kopfMenge(ist.header, volatile)
	for _, n := range slices.Sorted(maps.Keys(s)) {
		w, da := i[n]
		switch {
		case !da:
			return "Header " + n + " fehlt"
		case !slices.Equal(s[n], w):
			return fmt.Sprintf("Header %s: Werte %q gegen %q", n, s[n], w)
		}
	}
	for _, n := range slices.Sorted(maps.Keys(i)) {
		if _, da := s[n]; !da {
			return "Header " + n + " zusätzlich"
		}
	}
	return ""
}

// vergleicheStaende nennt den ersten Unterschied zweier Stände in jedem Feld, sonst "" (SPEC-038 Nr. 6).
func vergleicheStaende(a, b []model.ReplayCursor) string {
	if len(a) != len(b) {
		return fmt.Sprintf("Zahl der Cursor %d gegen %d", len(a), len(b))
	}
	for i := range a {
		x, y := a[i], b[i]
		for _, f := range []struct {
			name   string
			gleich bool
		}{
			{"WorkspaceID", x.WorkspaceID == y.WorkspaceID}, {"ReplaySessionID", x.ReplaySessionID == y.ReplaySessionID},
			{"RecordingSessionID", x.RecordingSessionID == y.RecordingSessionID}, {"ServiceID", x.ServiceID == y.ServiceID},
			{"Strategy", x.Strategy == y.Strategy}, {"Position", x.Position == y.Position},
			{"Consumed", slices.Equal(x.Consumed, y.Consumed)},
		} {
			if !f.gleich {
				return fmt.Sprintf("Cursor %d, Feld %s: %+v gegen %+v", i, f.name, x, y)
			}
		}
	}
	return ""
}

// pruefeStandNachReset prüft den Stand nach dem Reset: in Lauf 1 leer, danach die Identitäten des Stands lauf1, jeder
// Cursor mit Position 0 und leerem Consumed (SPEC-038 Nr. 6). Das Ergebnis nennt den ersten Verstoß, sonst "".
func pruefeStandNachReset(stand, lauf1 []model.ReplayCursor, erster bool) string {
	if erster {
		if len(stand) != 0 {
			return fmt.Sprintf("in Lauf 1 nicht leer: %d Cursor", len(stand))
		}
		return ""
	}
	if len(stand) != len(lauf1) {
		return fmt.Sprintf("Zahl der Cursor %d, in Lauf 1 %d", len(stand), len(lauf1))
	}
	for i, c := range stand {
		o := lauf1[i]
		switch {
		case c.WorkspaceID != o.WorkspaceID || c.ReplaySessionID != o.ReplaySessionID || c.RecordingSessionID != o.RecordingSessionID:
			return fmt.Sprintf("Identität des Cursors %d: %s/%s/%s", i, c.WorkspaceID, c.ReplaySessionID, c.RecordingSessionID)
		case c.Position != 0:
			return fmt.Sprintf("Cursor %d: Position %d, erwartet 0", i, c.Position)
		case len(c.Consumed) != 0:
			return fmt.Sprintf("Cursor %d: Consumed %v, erwartet leer", i, c.Consumed)
		}
	}
	return ""
}

// pruefeKontrollwert prüft den Stand nach Lauf 1: genau ein Cursor mit der Replay-Session `determinismus`, der
// Recording-Session der Aufnahme, Position größer 0 und nichtleerem Consumed (SPEC-038 Nr. 6).
func pruefeKontrollwert(stand []model.ReplayCursor, recording string) string {
	if len(stand) != 1 {
		return fmt.Sprintf("%d Cursor, erwartet 1", len(stand))
	}
	c := stand[0]
	switch {
	case c.ReplaySessionID != determinismusSession || c.RecordingSessionID != recording:
		return fmt.Sprintf("Cursor %s/%s, erwartet %s/%s", c.ReplaySessionID, c.RecordingSessionID, determinismusSession, recording)
	case c.Position <= 0 || len(c.Consumed) == 0:
		return fmt.Sprintf("Position %d, Consumed %v", c.Position, c.Consumed)
	}
	return ""
}

// leseStand liest den Stand aus der Ablage über eine zweite Verbindung, nur lesend und nach dem Lesen geschlossen;
// jeder Fehler beendet den Test mit der Nummer des Laufs (SPEC-038 Nr. 6, Nr. 8 a).
func leseStand(t *testing.T, lauf int, dir, workspace string) []model.ReplayCursor {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		t.Fatalf("Lauf %d: Ablage öffnen: %v", lauf, err)
	}
	stand, err := db.Cursors().List(ctx, workspace)
	if cerr := db.Close(); err == nil && cerr != nil {
		err = cerr
	}
	if err != nil {
		t.Fatalf("Lauf %d: Stand lesen: %v", lauf, err)
	}
	slices.SortFunc(stand, func(a, b model.ReplayCursor) int {
		return strings.Compare(a.ReplaySessionID+"\x00"+a.RecordingSessionID, b.ReplaySessionID+"\x00"+b.RecordingSessionID)
	})
	return stand
}

// sendeFolge sendet die zehn Requests nacheinander über c mit den Headern der Zuordnung h (SPEC-038 Nr. 3, 4).
func sendeFolge(c *http.Client, adresse string, h map[string][]string) ([]antwort, error) {
	var out []antwort
	for n, pfad := range determinismusFolge() {
		req, err := http.NewRequest(http.MethodGet, "http://"+adresse+pfad, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+geheimA)
		req.Header.Set("Cookie", "sid="+geheimB)
		req.Header.Set("Accept", "*/*")
		for k, vs := range h {
			req.Header.Del(k)
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
		resp, err := c.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Request %d (%s): %w", n+1, pfad, err)
		}
		b, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("Request %d (%s): Body: %w", n+1, pfad, err)
		}
		out = append(out, antwort{status: resp.StatusCode, header: resp.Header, body: b})
	}
	return out, nil
}

// abweichungOrakel nennt die Abweichungen von Lauf 1 gegen das Orakel (SPEC-038 Nr. 4): Status und Body der Aufnahme
// für die Requests 1 bis 8, 410 mit ODR-E5002 für 9, 404 mit ODR-E5001 für 10.
func abweichungOrakel(aufnahme []paar, a []antwort) []string {
	if len(a) != 10 {
		return []string{fmt.Sprintf("%d Antworten, erwartet 10", len(a))}
	}
	var out []string
	for i, p := range aufnahme {
		if a[i].status != p.w.status || string(a[i].body) != string(p.w.body) {
			out = append(out, fmt.Sprintf("Orakel, Request %d: %d %q, Aufnahme %d %q", i+1, a[i].status, a[i].body, p.w.status, p.w.body))
		}
	}
	for i, soll := range map[int]struct {
		status int
		code   string
	}{8: {410, "ODR-E5002"}, 9: {404, "ODR-E5001"}} {
		if a[i].status != soll.status || !strings.Contains(string(a[i].body), soll.code) {
			out = append(out, fmt.Sprintf("Orakel, Request %d: %d %q, erwartet %d mit %s", i+1, a[i].status, a[i].body, soll.status, soll.code))
		}
	}
	return out
}

// TestE2EDeterminismus ist die Messung von LH-QA-02 für den Replay (SPEC-038): hundert Läufe derselben Folge nach
// Reset gegen dasselbe Handle, Antworten und Cursor-Stände gleich Lauf 1.
func TestE2EDeterminismus(t *testing.T) {
	binaryPfad(t)
	ref := neueReferenz(t)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	for _, a := range aufnahmeReplay() {
		l.frage(a)
	}
	l.beendeAufnahme()
	l.importiere()
	ref.stoppe()

	ctx := context.Background()
	k, err := bootstrap.LadeKonfiguration(l.konfig("replay").env(platzhalterListen))
	if err != nil {
		t.Fatalf("Vorbereitung: Konfiguration: %v", err)
	}
	k.Listen = "127.0.0.1:0"
	h, err := bootstrap.Build(ctx, k, nil)
	if err != nil {
		t.Fatalf("Vorbereitung: Build: %v", err)
	}
	t.Cleanup(func() {
		if err := h.Herunterfahren(ctx); err != nil {
			t.Errorf("Herunterfahren: %v", err)
		}
	})
	if err := h.Hochfahren(ctx); err != nil {
		t.Fatalf("Vorbereitung: Hochfahren: %v", err)
	}
	tr := &http.Transport{DisableCompression: true}
	t.Cleanup(tr.CloseIdleConnections)
	client := &http.Client{Transport: tr, Timeout: 20 * time.Second}

	var antwort1 []antwort
	var stand1 []model.ReplayCursor
	abweichend := map[int]int{}
	weicht := func(i int, format string, args ...any) {
		t.Helper()
		abweichend[i]++
		if abweichend[i] <= 3 {
			t.Errorf("Lauf %d: %s", i, fmt.Sprintf(format, args...))
		}
	}
	gelaufen := 0
	for i := 1; i <= laeufe; i++ {
		w, err := h.Workspaces().Reset(ctx, l.workspace)
		if err != nil {
			t.Fatalf("Lauf %d: Reset: %v", i, err)
		}
		if w.Revision != int64(i) {
			t.Fatalf("Lauf %d: Revision %d, erwartet %d", i, w.Revision, i)
		}
		if m := pruefeStandNachReset(leseStand(t, i, l.dir, l.workspace), stand1, i == 1); m != "" {
			weicht(i, "Stand nach Reset: %s", m)
		}
		antworten, err := sendeFolge(client, h.Adresse(), l.replayHeader(determinismusSession))
		if err != nil {
			t.Fatalf("Lauf %d: %v", i, err)
		}
		stand := leseStand(t, i, l.dir, l.workspace)
		if i == 1 {
			for _, m := range abweichungOrakel(l.aufnahme, antworten) {
				weicht(i, "%s", m)
			}
			if m := pruefeKontrollwert(stand, l.session); m != "" {
				t.Fatalf("Lauf 1: Leser ungültig: %s", m)
			}
			antwort1, stand1 = antworten, stand
		} else {
			for j := range antworten {
				if m := vergleicheAntworten(antwort1[j], antworten[j], k.VolatileHeaders); m != "" {
					weicht(i, "Request %d: %s", j+1, m)
				}
			}
			if m := vergleicheStaende(stand1, stand); m != "" {
				weicht(i, "Stand nach dem Lauf: %s", m)
			}
		}
		gelaufen++
	}
	if gelaufen != laeufe {
		t.Errorf("Läufe: %d, erwartet %d", gelaufen, laeufe)
	}
	if ws, err := h.Workspaces().Get(ctx, l.workspace); err != nil || ws.Revision != laeufe {
		t.Errorf("Revision nach Lauf %d: %d, %v, erwartet %d", laeufe, ws.Revision, err, laeufe)
	}
	if err := h.Herunterfahren(ctx); err != nil {
		t.Errorf("Herunterfahren: %v", err)
	}
	t.Logf("Läufe: %d, abweichend: %d", gelaufen, len(abweichend))
}

// probe ist eine Antwort mit jedem Header, den die Fälle ändern; Vary trägt zwei Werte.
func probe() antwort {
	return antwort{status: 200, body: []byte(`{"value":[{"Id":1}]}`), header: http.Header{
		"Content-Type": {"application/json"}, "Odata-Version": {"4.0"}, "Content-Length": {"20"}, "Etag": {`W/"p1"`},
		"Vary": {"Accept", "Prefer"}, "X-Recorder-Match-Redacted": {"false"}, "Date": {"Mon, 02 Jan 2006 15:04:05 GMT"},
	}}
}

func klon(a antwort) antwort {
	h := http.Header{}
	for n, v := range a.header {
		h[n] = slices.Clone(v)
	}
	return antwort{status: a.status, body: slices.Clone(a.body), header: h}
}

func standardVolatile(t *testing.T, extra ...string) []string {
	t.Helper()
	k, err := bootstrap.LadeKonfiguration(konfig{modus: "replay", dir: t.TempDir(), upstream: "http://127.0.0.1:9", extra: extra}.env(platzhalterListen))
	if err != nil {
		t.Fatalf("Konfiguration: %v", err)
	}
	return k.VolatileHeaders
}

// TestDeterminismusVergleich fängt SPEC-038 Nr. 5 ohne Binary: je Art der Abweichung ein Fall in beiden Richtungen,
// die Reihenfolge der Meldung, jeder flüchtige Name in jeder Schreibweise und die Liste des Handle.
func TestDeterminismusVergleich(t *testing.T) {
	vol := standardVolatile(t)
	mit := func(f func(a *antwort)) antwort { a := klon(probe()); f(&a); return a }
	faelle := map[string]antwort{
		"Status":           mit(func(a *antwort) { a.status = 201 }),
		"Body erstes Byte": mit(func(a *antwort) { a.body[0] ^= 1 }), "Body mittleres Byte": mit(func(a *antwort) { a.body[len(a.body)/2] ^= 1 }),
		"Body letztes Byte": mit(func(a *antwort) { a.body[len(a.body)-1] ^= 1 }), "Body gekürzt": mit(func(a *antwort) { a.body = a.body[1:] }),
		"Body leer":               mit(func(a *antwort) { a.body = nil }),
		"zusätzlicher Name":       mit(func(a *antwort) { a.header["X-Neu"] = []string{"1"} }),
		"Reihenfolge der Werte":   mit(func(a *antwort) { a.header["Vary"] = []string{"Prefer", "Accept"} }),
		"zusätzlicher Wert":       mit(func(a *antwort) { a.header["Odata-Version"] = []string{"4.0", "4.0"} }),
		"fehlender Wert von Vary": mit(func(a *antwort) { a.header["Vary"] = []string{"Accept"} }),
	}
	for n := range probe().header {
		if !slices.Contains(vol, n) {
			faelle["Wert von "+n] = mit(func(a *antwort) { a.header[n] = []string{a.header[n][0] + "x"} })
			faelle["fehlender Name "+n] = mit(func(a *antwort) { delete(a.header, n) })
			faelle["leerer Wert von "+n] = mit(func(a *antwort) { a.header[n] = []string{""} })
		}
	}
	if len(faelle) < 25 {
		t.Fatalf("%d Fälle, erwartet mindestens 25", len(faelle))
	}
	for name, ab := range faelle {
		for richtung, p := range map[string][2]antwort{"soll gegen ist": {probe(), ab}, "ist gegen soll": {ab, probe()}} {
			if vergleicheAntworten(p[0], p[1], vol) == "" {
				t.Errorf("%s (%s): kein Unterschied gemeldet", name, richtung)
			}
		}
	}
	if m := vergleicheAntworten(probe(), klon(probe()), vol); m != "" {
		t.Errorf("gleiche Antworten: %s", m)
	}
	t.Run("Reihenfolge", func(t *testing.T) {
		st, bo, kopf := mit(func(a *antwort) { a.status = 500 }), mit(func(a *antwort) { a.body[0] ^= 1 }), mit(func(a *antwort) { a.header["Etag"] = []string{"x"} })
		for _, f := range []struct {
			name string
			ab   antwort
			art  string
		}{{"Status und Body", mit(func(a *antwort) { a.status, a.body = st.status, bo.body }), "Status"},
			{"Status und Header", mit(func(a *antwort) { a.status, a.header = st.status, kopf.header }), "Status"},
			{"Body und Header", mit(func(a *antwort) { a.body, a.header = bo.body, kopf.header }), "Body"}} {
			if m := vergleicheAntworten(probe(), f.ab, vol); !strings.HasPrefix(m, f.art) {
				t.Errorf("%s: Meldung %q, erwartet Art %s", f.name, m, f.art)
			}
		}
	})
	t.Run("flüchtige Namen", func(t *testing.T) {
		if len(vol) != 7 {
			t.Fatalf("Default von VOLATILE_HEADERS: %v, erwartet 7 Namen", vol)
		}
		for _, n := range vol {
			for _, schreibweise := range []string{strings.ToLower(n), strings.ToUpper(n), n[:1] + strings.ToUpper(n[1:2]) + n[2:]} {
				a, b := klon(probe()), klon(probe())
				a.header = http.Header{schreibweise: {"eins"}}
				b.header = http.Header{n: {"zwei"}}
				if m := vergleicheAntworten(a, b, vol); m != "" {
					t.Errorf("flüchtiger Name %s als %s: %s", n, schreibweise, m)
				}
			}
		}
		for _, n := range []string{"Content-Length", "Content-Encoding", "Transfer-Encoding", "Connection", "ETag", "X-Recorder-Match-Redacted"} {
			a, b := klon(probe()), klon(probe())
			a.header = http.Header{n: {"eins"}}
			b.header = http.Header{n: {"zwei"}}
			if vergleicheAntworten(a, b, vol) == "" {
				t.Errorf("Name %s ist nicht ausgenommen, der Wert unterscheidet", n)
			}
		}
	})
	t.Run("Liste des Handle", func(t *testing.T) {
		x1, x2, datum := klon(probe()), klon(probe()), klon(probe())
		x1.header["X-Probe"], x2.header["X-Probe"] = []string{"eins"}, []string{"zwei"}
		datum.header["Date"] = []string{"anders"}
		for _, env := range []string{"X-Probe", "x-probe"} {
			v := standardVolatile(t, "VOLATILE_HEADERS="+env)
			if m := vergleicheAntworten(x1, x2, v); m != "" {
				t.Errorf("VOLATILE_HEADERS=%s: X-Probe unterscheidet: %s", env, m)
			}
			if m := vergleicheAntworten(probe(), datum, v); !strings.HasPrefix(m, "Header Date") {
				t.Errorf("VOLATILE_HEADERS=%s: Date unterscheidet nicht: %q", env, m)
			}
		}
		keine := standardVolatile(t, "VOLATILE_HEADERS=none")
		if len(keine) != 0 || vergleicheAntworten(x1, x2, keine) == "" || vergleicheAntworten(probe(), datum, keine) == "" {
			t.Errorf("VOLATILE_HEADERS=none: Liste %v, jeder Name unterscheidet", keine)
		}
		if m := vergleicheAntworten(x1, x2, []string{"x-probe"}); m != "" {
			t.Errorf("Liste in kleiner Schreibweise: %s", m)
		}
	})
}

// cursor ist ein Cursor mit den Werten eines Stands nach Lauf 1.
func cursor(replay, recording string) model.ReplayCursor {
	return model.ReplayCursor{WorkspaceID: "w", ReplaySessionID: replay, RecordingSessionID: recording, ServiceID: "ref",
		Strategy: model.StrategySequential, Position: 3, Consumed: []int64{1, 2, 3}}
}

// TestDeterminismusStand fängt SPEC-038 Nr. 6 ohne Binary: der Vergleich in jedem Feld jedes Cursors, der Stand nach
// dem Reset in jedem Verstoß und der Kontrollwert des Lesers.
func TestDeterminismusStand(t *testing.T) {
	basis := func() []model.ReplayCursor { return []model.ReplayCursor{cursor("a", "r1"), cursor("b", "r1")} }
	aendern := map[string]func(c *model.ReplayCursor){
		"WorkspaceID": func(c *model.ReplayCursor) { c.WorkspaceID = "x" }, "ReplaySessionID": func(c *model.ReplayCursor) { c.ReplaySessionID = "x" },
		"RecordingSessionID": func(c *model.ReplayCursor) { c.RecordingSessionID = "x" }, "ServiceID": func(c *model.ReplayCursor) { c.ServiceID = "x" },
		"Strategy": func(c *model.ReplayCursor) { c.Strategy = model.StrategyStrict }, "Position": func(c *model.ReplayCursor) { c.Position++ },
		"Consumed Länge":       func(c *model.ReplayCursor) { c.Consumed = c.Consumed[:2] },
		"Consumed Element":     func(c *model.ReplayCursor) { c.Consumed = []int64{1, 2, 4} },
		"Consumed Reihenfolge": func(c *model.ReplayCursor) { c.Consumed = []int64{1, 3, 2} },
		"Consumed leer":        func(c *model.ReplayCursor) { c.Consumed = nil },
	}
	for name, f := range aendern {
		for i := range basis() {
			a := basis()
			f(&a[i])
			if vergleicheStaende(basis(), a) == "" || vergleicheStaende(a, basis()) == "" {
				t.Errorf("Feld %s von Cursor %d: kein Unterschied gemeldet", name, i)
			}
		}
	}
	for name, a := range map[string][]model.ReplayCursor{"Cursor fehlt": basis()[:1], "zusätzlicher Cursor": append(basis(), cursor("c", "r1"))} {
		if vergleicheStaende(basis(), a) == "" || vergleicheStaende(a, basis()) == "" {
			t.Errorf("%s: kein Unterschied gemeldet", name)
		}
	}
	if m := vergleicheStaende(basis(), basis()); m != "" {
		t.Errorf("gleiche Stände: %s", m)
	}
	a, b := basis(), basis()
	a[0].Consumed, a[0].Position, b[0].Consumed, b[0].Position = nil, 0, []int64{}, 0
	if m := vergleicheStaende(a, b); m != "" {
		t.Errorf("leeres Consumed nil gegen leer: %s", m)
	}

	t.Run("nach Reset", func(t *testing.T) {
		leer := func() []model.ReplayCursor {
			s := basis()
			for i := range s {
				s[i].Position, s[i].Consumed = 0, nil
			}
			return s
		}
		if m := pruefeStandNachReset(leer(), basis(), false); m != "" {
			t.Errorf("zurückgesetzter Stand: %s", m)
		}
		if m := pruefeStandNachReset(nil, basis(), true); m != "" {
			t.Errorf("leerer Stand in Lauf 1: %s", m)
		}
		verstoesse := map[string]func(s []model.ReplayCursor) []model.ReplayCursor{
			"Position":                 func(s []model.ReplayCursor) []model.ReplayCursor { s[1].Position = 1; return s },
			"Consumed":                 func(s []model.ReplayCursor) []model.ReplayCursor { s[1].Consumed = []int64{1}; return s },
			"fehlt":                    func(s []model.ReplayCursor) []model.ReplayCursor { return s[:1] },
			"zusätzlich":               func(s []model.ReplayCursor) []model.ReplayCursor { return append(s, cursor("c", "r1")) },
			"andere Replay-Session":    func(s []model.ReplayCursor) []model.ReplayCursor { s[0].ReplaySessionID = "x"; return s },
			"andere Recording-Session": func(s []model.ReplayCursor) []model.ReplayCursor { s[0].RecordingSessionID = "x"; return s },
			"andere Workspace-ID":      func(s []model.ReplayCursor) []model.ReplayCursor { s[0].WorkspaceID = "x"; return s },
			"keiner":                   func(s []model.ReplayCursor) []model.ReplayCursor { return nil },
		}
		for name, f := range verstoesse {
			if pruefeStandNachReset(f(leer()), basis(), false) == "" {
				t.Errorf("Verstoß %s: nicht gemeldet", name)
			}
		}
		if pruefeStandNachReset(basis(), nil, true) == "" {
			t.Errorf("Stand in Lauf 1 nicht leer: nicht gemeldet")
		}
	})
	t.Run("Kontrollwert", func(t *testing.T) {
		gut := func() []model.ReplayCursor {
			c := cursor(determinismusSession, "r1")
			return []model.ReplayCursor{c}
		}
		if m := pruefeKontrollwert(gut(), "r1"); m != "" {
			t.Errorf("gültiger Stand: %s", m)
		}
		for name, s := range map[string][]model.ReplayCursor{
			"leer": nil, "zwei": append(gut(), cursor("zweite", "r1")),
			"andere Replay-Session": {cursor("x", "r1")}, "andere Recording-Session": {cursor(determinismusSession, "x")},
			"Position 0": {func() model.ReplayCursor { c := gut()[0]; c.Position, c.Consumed = 0, nil; return c }()},
		} {
			if pruefeKontrollwert(s, "r1") == "" {
				t.Errorf("%s: Leser gilt als gültig", name)
			}
		}
	})
}
