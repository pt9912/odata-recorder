package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/pt9912/odata-recorder/src/bootstrap"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// m4Methoden sind die zehn Methoden von SPEC-261 Nr. 2: die lesenden, die mutierenden und drei, die nie freigegeben sind.
func m4Methoden() []string {
	return []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "CONNECT"}
}

// m4Mutierend sind die fünf mutierenden Methoden (SPEC-261 Nr. 2 b).
func m4Mutierend() []string { return []string{"POST", "PUT", "PATCH", "DELETE", "MERGE"} }

// m4Pfad ist der Pfad, auf dem jede mutierende Methode den Mock mit ODR-E6002 scheitern lässt (keine Collection und keine Entity).
const m4Pfad = m3Pfad + "/$count"

// odrCode liefert den ersten Code `ODR-Exxxx` des Bodys, sonst "".
func odrCode(b []byte) string {
	i := bytes.Index(b, []byte("ODR-E"))
	if i < 0 || i+9 > len(b) {
		return ""
	}
	return string(b[i : i+9])
}

// m4Welt ist m3Welt plus eine dritte Aufnahme-Session: sie trägt die Austausche, die der Replay-Fallback abspielt, und ist nie importiert.
type m4Welt struct {
	*m3Welt
	s3 string
	// aufgenommen ist die Antwort je Austausch von m4Replay, wie der Client sie bei der Aufnahme sah.
	aufgenommen map[string]antwort
}

// m4Replay ist ein Austausch der dritten Session: ein Request, den der Mock mit dem Auslöser der Kategorie beantwortet.
type m4Replay struct {
	name, kategorie, methode, pfad, body string
	teil                                 bool
	status                               int
}

// m4ReplayFaelle sind die Austausche der Replay-Fallbacks: Methode (6) × Kategorie (2) ohne die Zelle POST × coverage_insufficient, die der
// Mock nie mit ODR-E6003 beantwortet (eine POST auf eine Collection legt an, eine auf eine Entity ist ODR-E6002).
func m4ReplayFaelle() []m4Replay {
	var out []m4Replay
	for _, m := range []string{"GET", "HEAD"} {
		out = append(out, m4Replay{m + " unsupported", "unsupported", m, m3Pfad + "?$search=x", "", false, 501},
			m4Replay{m + " coverage_insufficient", "coverage_insufficient", m, m3Pfad, "", true, 422})
	}
	for _, m := range []string{"POST", "PUT", "PATCH", "DELETE"} {
		body := ""
		if m != "DELETE" {
			body = `{"Id":2,"Name":"x"}`
		}
		out = append(out, m4Replay{m + " unsupported", "unsupported", m, m4Pfad, body, false, 501})
		if m != "POST" {
			out = append(out, m4Replay{m + " coverage_insufficient", "coverage_insufficient", m, m3Pfad + "(2)", body, true, 422})
		}
	}
	return out
}

// m4Aufbau baut die Welt von M3 und nimmt danach die Austausche von m4ReplayFaelle in einer dritten Session auf (Modus record gegen einen
// zweiten Referenzdienst, ohne Import).
func m4Aufbau(t *testing.T) *m4Welt {
	t.Helper()
	ctx := context.Background()
	w := &m4Welt{m3Welt: m3Aufbau(t), aufgenommen: map[string]antwort{}}
	ref := neuerReferenzdienst(t)
	rec := hochgefahren(ctx, t, konfig(t, "record", w.dir, ref.URL()+"/odata"))
	sessions, err := rec.Sessions().List(ctx)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	for _, s := range sessions {
		if s.Status == model.SessionOpen {
			w.s3 = s.ID
		}
	}
	if w.s3 == "" || w.s3 == w.sessionID {
		t.Fatalf("keine offene dritte Session: %+v", sessions)
	}
	rc := neuerClient(t, rec.Adresse())
	for _, f := range m4ReplayFaelle() {
		w.aufgenommen[f.name] = rc.sende(f.methode, f.pfad, f.body, "", nil)
	}
	if err := rec.Herunterfahren(ctx); err != nil {
		t.Fatalf("Herunterfahren(record): %v", err)
	}
	return w
}

// kopie legt eine Kopie des DATA_DIR an: jeder Start im Modus hybrid arbeitet auf einer eigenen, weil ein Fall den Mock verändern darf.
func (w *m4Welt) kopie(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(w.dir)); err != nil {
		t.Fatalf("Kopie des DATA_DIR: %v", err)
	}
	return dir
}

// host ist der Eintrag für FALLBACK_ALLOW_HOSTS zum Referenzdienst.
func (r *referenzdienst) host() string { return strings.TrimPrefix(r.URL(), "http://") }

// startMock fährt den Modus mock auf einer Kopie des DATA_DIR hoch: die Antworten des Mock sind das Orakel der Fälle ohne Fallback.
func (w *m4Welt) startMock(t *testing.T) *client {
	t.Helper()
	return neuerClient(t, hochgefahren(context.Background(), t, konfig(t, "mock", w.kopie(t), "", "ETAG_ENABLED=true")).Adresse())
}

// startHybrid fährt den Modus hybrid auf einer Kopie des DATA_DIR hoch; extra sind die Zeilen der Umgebung außer den Grundwerten.
func (w *m4Welt) startHybrid(t *testing.T, up *referenzdienst, extra ...string) *client {
	t.Helper()
	k := konfig(t, "hybrid", w.kopie(t), up.URL()+"/odata", extra...)
	h := hochgefahren(context.Background(), t, k)
	if h.Sessions() != nil || h.Record() != nil || h.Import() != nil {
		t.Fatal("Sessions(), Record() oder Import() im Modus hybrid ist nicht nil")
	}
	return neuerClient(t, h.Adresse())
}

// offen sind die Schlüssel der offensten Policy: Hauptschalter an, der Host des Referenzdienstes, jede Methode freigegeben auf jedem Pfad
// des Service und beide Kategorien des Replay-Fallbacks.
func offen(up *referenzdienst, ohneKategorien bool) []string {
	z := []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=" + up.host(), "ETAG_ENABLED=true",
		"FALLBACK_WRITE_ALLOW=" + routen(m4Mutierend(), "/odata/**")}
	if !ohneKategorien {
		z = append(z, "HYBRID_REPLAY_FALLBACK=unsupported,coverage_insufficient")
	}
	return z
}

// routen ist die Liste der Einträge von FALLBACK_WRITE_ALLOW: je Methode ein Eintrag mit dem Muster.
func routen(methoden []string, muster string) string {
	e := make([]string, len(methoden))
	for i, m := range methoden {
		e[i] = m + " " + muster
	}
	return strings.Join(e, ",")
}

// stand ist der Zählerstand eines Referenzdienstes.
type stand struct {
	verbindungen, requests int64
	methoden               map[string]int64
	gesehen                int
}

func (r *referenzdienst) stand() stand {
	s := stand{verbindungen: r.verbindungen.Load(), requests: r.requests.Load(), methoden: map[string]int64{}}
	for _, m := range m4Methoden() {
		s.methoden[m] = r.nachMethode(m)
	}
	r.mu.Lock()
	s.gesehen = len(r.gesehen)
	r.mu.Unlock()
	return s
}

// erwarte prüft die Zähler gegen das Orakel: seit vor sind es genau `verb` Verbindungen und je Methode in soll die Zahl der Requests, sonst keine.
func (r *referenzdienst) erwarte(t *testing.T, name string, vor stand, verb int64, soll map[string]int64) {
	t.Helper()
	nach := r.stand()
	summe := int64(0)
	for _, n := range soll {
		summe += n
	}
	if got := nach.verbindungen - vor.verbindungen; got != verb {
		t.Errorf("%s: %d neue Verbindungen am Referenzdienst, erwartet %d", name, got, verb)
	}
	if got := nach.requests - vor.requests; got != summe {
		t.Errorf("%s: %d neue Requests am Referenzdienst, erwartet %d", name, got, summe)
	}
	for _, m := range m4Methoden() {
		if got := nach.methoden[m] - vor.methoden[m]; got != soll[m] {
			t.Errorf("%s: %d neue Requests mit der Methode %s, erwartet %d", name, got, m, soll[m])
		}
	}
}

// keinKontakt ist erwarte ohne Verbindung und ohne Request.
func (r *referenzdienst) keinKontakt(t *testing.T, name string, vor stand) {
	t.Helper()
	r.erwarte(t, name, vor, 0, nil)
}

// fallback ist der Wert des Headers X-Recorder-Fallback, "" ohne ihn.
func fallback(a antwort) string {
	if v := a.header.Values("X-Recorder-Fallback"); len(v) == 1 {
		return v[0]
	}
	return ""
}

// replayKopf sind die drei Header der Zuordnung.
func (w *m4Welt) replayKopf(ws model.Workspace, aufnahme, replaySession string) map[string]string {
	return map[string]string{"X-Recorder-Workspace": ws.ID, "X-Recorder-Recording-Session": aufnahme, "X-Recorder-Replay-Session": replaySession}
}

// TestM4 ist der M4-Test (LH-FA-12): zwei Baselines (eine vollständig, eine nicht) und eine dritte Aufnahme-Session in einem DATA_DIR, je Fall ein
// Start im Modus hybrid auf einer Kopie, ein Referenzdienst, der Verbindungen und Methoden zählt, und der echte Adapter httpupstream.
func TestM4(t *testing.T) {
	w := m4Aufbau(t)
	t.Run("Ohne Auslöser kein Upstream", w.ohneAusloeser)
	t.Run("Lesender Fallback", w.lesenderFallback)
	t.Run("Ohne Freigabe", w.ohneFreigabe)
	t.Run("Mutierender Fallback", w.mutierend)
	t.Run("Replay-Miss", w.replayMiss)
	t.Run("Erschöpfter Key", w.erschoepft)
	t.Run("Replay-Fallback", w.replayFallback)
	t.Run("Gleichzeitig", w.gleichzeitig)
}

// ohneAusloeser fängt M12: kein Auslöser, kein Upstream; Fall × offenste Policy, Zählprüfung gegen die Ergebnisse des Mock in der Liste von SPEC-261
// Nr. 3 c (Treffer lesend und schreibend, 404, 409, 412, ODR-E6005, ODR-E6007). Von den Fehlern des Replays in dieser Liste führen `erschoepft`
// (ODR-E5002) und `replayFallback` (ODR-E5003) je einen; die übrigen sind hier nicht geprüft.
func (w *m4Welt) ohneAusloeser(t *testing.T) {
	up := neuerReferenzdienst(t)
	c := w.startHybrid(t, up, offen(up, false)...)
	faelle := []struct {
		name, methode, pfad, body string
		hdr                       map[string]string
		status                    int
		code                      string
	}{
		{"Treffer lesend", "GET", m3Pfad + "?$orderby=Id", "", nil, 200, ""},
		{"Treffer schreibend", "POST", m3Pfad, `{"Id":601,"Name":"N"}`, nil, 201, ""},
		{"fachliche 404", "GET", m3Pfad + "(999)", "", nil, 404, "ODR-E6008"},
		{"fachliche 409", "POST", m3Pfad, `{"Id":1,"Name":"D"}`, nil, 409, "ODR-E6009"},
		{"fachliche 412", "PATCH", m3Pfad + "(1)", `{"Name":"i"}`, map[string]string{"If-Match": `"nope"`}, 412, "ODR-E6010"},
		{"ungültige Anfrage E6005", "GET", m3Pfad + "?$orderby=Unbekannt", "", nil, 400, "ODR-E6005"},
		{"ungültige Anfrage E6007", "POST", m3Pfad, `{"Name":"x"}`, nil, 400, "ODR-E6007"},
	}
	if len(faelle) != 7 {
		t.Fatalf("%d Fälle, die Liste von SPEC-261 Nr. 3 c hat 7 (Treffer lesend und schreibend, 404, 409, 412, E6005, E6007)", len(faelle))
	}
	for _, f := range faelle {
		vor := up.stand()
		hdr := kopf(w.ws)
		for k, v := range f.hdr {
			hdr[k] = v
		}
		r := c.sende(f.methode, f.pfad, f.body, "", hdr)
		if r.status != f.status || odrCode(r.body) != f.code || fallback(r) != "" {
			t.Errorf("%s: %d %s %q, erwartet %d %s ohne Kennzeichnung", f.name, r.status, odrCode(r.body), fallback(r), f.status, f.code)
		}
		up.keinKontakt(t, f.name, vor)
	}
}

// lesenderFallback fängt M13 und Z35: lesend und ODR-E6003 oder ODR-E6001 mit freigegebenem Host: genau eine Verbindung mit der Methode des Requests
// zum Referenzdienst der UPSTREAM_URL und zu keinem anderen; die Antwort trägt X-Recorder-Fallback: upstream. Der Client nennt im Host-Header den
// zweiten Referenzdienst: der Adapter verbindet sich mit dem Host, den die Policy verglich.
func (w *m4Welt) lesenderFallback(t *testing.T) {
	up, fremd := neuerReferenzdienst(t), neuerReferenzdienst(t)
	up.jedeVerbindungNeu.Store(true)
	c := w.startHybrid(t, up, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+up.host())
	n := 0
	for _, m := range []string{"GET", "HEAD"} {
		for _, code := range []string{"ODR-E6003", "ODR-E6001"} {
			n++
			ws, pfad, query := w.wsTeil, m3Pfad, ""
			if code == "ODR-E6001" {
				ws, pfad, query = w.ws, m3Pfad+"?$search=x", "$search=x"
			}
			name := m + " " + code
			vor, vorFremd := up.stand(), fremd.stand()
			r := c.sende(m, pfad, "", fremd.host(), kopf(ws))
			if r.status != 200 || fallback(r) != "upstream" {
				t.Errorf("%s: %d, Kennzeichnung %q, erwartet 200 und upstream", name, r.status, fallback(r))
			}
			up.erwarte(t, name, vor, 1, map[string]int64{m: 1})
			fremd.keinKontakt(t, name+" (zweiter Referenzdienst)", vorFremd)
			if g := up.gesehenSeit(vor.gesehen); len(g) == 1 && (g[0].Pfad != "/odata/Products" || g[0].Query != query) {
				t.Errorf("%s: Request am Referenzdienst %+v", name, g[0])
			}
		}
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet {GET, HEAD} × {ODR-E6003, ODR-E6001}", n)
	}
}

// ohneFreigabe fängt M14: lesend ohne Freigabe ist ODR-E6003, byte-gleich der Antwort des Modus mock, und 0 Verbindungen; sechs Zustände der
// Policy × {GET, HEAD}.
func (w *m4Welt) ohneFreigabe(t *testing.T) {
	mc := w.startMock(t)
	up, anderer := neuerReferenzdienst(t), neuerReferenzdienst(t)
	teilName := strings.TrimPrefix(up.URL(), "http://127.0.0.1:")
	zustaende := []struct {
		name  string
		extra []string
		// dienste heißt: ein Literal mit zwei Services, der Service der Anfrage ohne UpstreamURL.
		dienste bool
	}{
		{"Hauptschalter aus bei gesetzten Listen", []string{"HYBRID_FALLBACK=disabled", "FALLBACK_ALLOW_HOSTS=" + up.host(),
			"FALLBACK_WRITE_ALLOW=" + routen(m4Mutierend(), "/odata/**"), "HYBRID_REPLAY_FALLBACK=unsupported"}, false},
		{"fremder Host", []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=fremd.test:" + teilName}, false},
		{"Host mit anderem Port", []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=" + anderer.host()}, false},
		{"Endung eines Namens", []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=0.0.1:" + teilName}, false},
		{"IPv6 statt Name", []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=[::1]:" + teilName}, false},
		{"Service ohne UPSTREAM_URL neben einem mit", []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=" + up.host()}, true},
	}
	if len(zustaende) != 6 {
		t.Fatalf("%d Zustände, erwartet 6", len(zustaende))
	}
	for _, z := range zustaende {
		var c *client
		if z.dienste {
			// Die Konfiguration verlangt bei enabled die UPSTREAM_URL jedes Service (SPEC-110 Nr. 9): der Fall entsteht nur mit einem Literal.
			k := konfig(t, "hybrid", w.kopie(t), up.URL()+"/odata", z.extra...)
			k.Dienste[0].UpstreamURL = ""
			k.Dienste = append(k.Dienste, bootstrap.Dienst{ID: "zwei", Root: "/zwei", UpstreamURL: up.URL() + "/zwei"})
			c = neuerClient(t, hochgefahren(context.Background(), t, k).Adresse())
		} else {
			c = w.startHybrid(t, up, z.extra...)
		}
		for _, m := range []string{"GET", "HEAD"} {
			name := z.name + " " + m
			vor, vorAnderer := up.stand(), anderer.stand()
			soll := mc.sende(m, m3Pfad, "", "", kopf(w.wsTeil))
			r := c.sende(m, m3Pfad, "", "", kopf(w.wsTeil))
			if r.status != 422 || m == "GET" && odrCode(r.body) != "ODR-E6003" || fallback(r) != "" {
				t.Errorf("%s: %d %s %q, erwartet 422 ODR-E6003 ohne Kennzeichnung", name, r.status, odrCode(r.body), fallback(r))
			}
			if r.status != soll.status || !bytes.Equal(r.body, soll.body) || r.header.Get("Content-Type") != soll.header.Get("Content-Type") {
				t.Errorf("%s: Antwort weicht von der des Modus mock ab: %d %q gegen %d %q", name, r.status, r.body, soll.status, soll.body)
			}
			up.keinKontakt(t, name, vor)
			anderer.keinKontakt(t, name+" (anderer Referenzdienst)", vorAnderer)
		}
	}
}

// m15Art ist eine Art des Eintrags von FALLBACK_WRITE_ALLOW für den Pfad m4Pfad; sendet ist das Orakel nach SPEC-261 Nr. 6 c (Methode byte-gleich,
// Segmente gleich, `*` und `**` wie dort), nicht der Code der Policy.
type m15Art struct {
	name     string
	eintrag  func(m string, i int) (methode, muster string)
	hostFrei bool
	sendet   bool
}

// m15Arten sind die Arten des Eintrags: fehlt, passt, Platzhalter, andere Methode, Host fremd und die Formen eines Musters, das einen anderen Pfad meint
// (anderes Segment, Anfang des Segments, Verlängerung des Segments, ein Segment weniger, ein Segment mehr, anderer Root, Teil eines Namens, andere
// Schreibweise).
func m15Arten(meth []string) []m15Art {
	gleich := func(muster string) func(string, int) (string, string) {
		return func(m string, _ int) (string, string) { return m, muster }
	}
	arten := []m15Art{
		{"fehlt", nil, true, false},
		{"passt", gleich(m4Pfad), true, true},
		{"Platzhalter *", gleich(m3Pfad + "/*"), true, true},
		{"Platzhalter **", gleich("/odata/**"), true, true},
		{"andere Methode", func(_ string, i int) (string, string) { return meth[(i+1)%len(meth)], m4Pfad }, true, false},
		{"Host fremd", gleich(m4Pfad), false, false},
	}
	for _, p := range []string{m3Pfad + "/Anderer", m3Pfad + "/$cou", m3Pfad + "/$counts", m3Pfad, m4Pfad + "/x", "/odata2/Products/$count",
		"/odata/Product/$count", "/odata/products/$count"} {
		arten = append(arten, m15Art{"anderer Pfad " + p, gleich(p), true, false})
	}
	return arten
}

// mutierend fängt M15: mutierend auf eine nicht unterstützte Operation: {POST, PUT, PATCH, DELETE, MERGE} × Art des Eintrags (14, siehe m15Arten);
// mit passendem Eintrag genau eine Verbindung mit Methode, Pfad und Body byte-gleich, sonst ODR-E6002.
func (w *m4Welt) mutierend(t *testing.T) {
	up := neuerReferenzdienst(t)
	up.jedeVerbindungNeu.Store(true)
	meth := m4Mutierend()
	bodyVon := func(m string) string {
		if m == "DELETE" {
			return ""
		}
		return `{"Id":9,"Name":"` + m + `"}`
	}
	arten := m15Arten(meth)
	n := 0
	for _, a := range arten {
		// Je Methode ein Start mit genau einem Eintrag: ein Eintrag für eine andere Methode darf die geprüfte nicht decken.
		for i, m := range meth {
			host := up.host()
			if !a.hostFrei {
				host = "fremd.test:1"
			}
			extra := []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=" + host}
			if a.eintrag != nil {
				methode, muster := a.eintrag(m, i)
				extra = append(extra, "FALLBACK_WRITE_ALLOW="+methode+" "+muster)
			}
			c := w.startHybrid(t, up, extra...)
			n++
			name := m + " " + a.name
			vor := up.stand()
			r := c.sende(m, m4Pfad, bodyVon(m), "", kopf(w.ws))
			if a.sendet {
				if fallback(r) != "upstream" {
					t.Errorf("%s: %d, Kennzeichnung %q, erwartet upstream", name, r.status, fallback(r))
				}
				up.erwarte(t, name, vor, 1, map[string]int64{m: 1})
				if g := up.gesehenSeit(vor.gesehen); len(g) == 1 && (g[0].Methode != m || g[0].Pfad != "/odata/Products/$count" || string(g[0].Body) != bodyVon(m)) {
					t.Errorf("%s: Request am Referenzdienst %+v", name, g[0])
				}
				continue
			}
			if r.status != 501 || odrCode(r.body) != "ODR-E6002" || fallback(r) != "" {
				t.Errorf("%s: %d %s %q, erwartet 501 ODR-E6002 ohne Kennzeichnung", name, r.status, odrCode(r.body), fallback(r))
			}
			up.keinKontakt(t, name, vor)
		}
	}
	if n != 5*14 {
		t.Errorf("%d Fälle, erwartet 5 Methoden × 14 Arten des Eintrags", n)
	}
}

// replayMiss fängt M16: Replay-Miss ODR-E5001 mit Opt-in: zehn Methoden × {Route, keine} × {Host frei, fremd}. GET und HEAD zum freigegebenen Host
// eine Verbindung; jede mutierende Methode, auch mit Route, keine und ODR-E5001 unverändert; OPTIONS, TRACE und CONNECT sind nie freigegeben und
// gehen nie zum Replay (der Mock antwortet ODR-E6002).
func (w *m4Welt) replayMiss(t *testing.T) {
	up := neuerReferenzdienst(t)
	up.jedeVerbindungNeu.Store(true)
	n := 0
	for _, route := range []bool{false, true} {
		for _, hostFrei := range []bool{true, false} {
			host := up.host()
			if !hostFrei {
				host = "fremd.test:1"
			}
			extra := []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=" + host, "HYBRID_REPLAY_FALLBACK=unsupported"}
			if route {
				extra = append(extra, "FALLBACK_WRITE_ALLOW="+routen(m4Mutierend(), "/odata/**"))
			}
			c := w.startHybrid(t, up, extra...)
			for _, m := range m4Methoden() {
				n++
				name := fmt.Sprintf("%s route=%v hostFrei=%v", m, route, hostFrei)
				pfad, body := m4Pfad+"?miss=1", ""
				if m == "GET" || m == "HEAD" {
					pfad = m3Pfad + "?$search=miss"
				} else if slices.Contains([]string{"POST", "PUT", "PATCH", "MERGE"}, m) {
					body = `{"Id":9}`
				}
				vor := up.stand()
				r := c.sende(m, pfad, body, "", w.replayKopf(w.ws, w.s3, "m16-"+fmt.Sprint(n)))
				lesend := m == "GET" || m == "HEAD"
				switch {
				case lesend && hostFrei:
					if fallback(r) != "upstream" {
						t.Errorf("%s: %d, Kennzeichnung %q, erwartet upstream nach dem Replay-Miss", name, r.status, fallback(r))
					}
					up.erwarte(t, name, vor, 1, map[string]int64{m: 1})
				case slices.Contains([]string{"OPTIONS", "TRACE", "CONNECT"}, m):
					if r.status != 501 || odrCode(r.body) != "ODR-E6002" || fallback(r) != "" {
						t.Errorf("%s: %d %s %q, erwartet 501 ODR-E6002", name, r.status, odrCode(r.body), fallback(r))
					}
					up.keinKontakt(t, name, vor)
				default:
					if r.status != 404 || m != "HEAD" && odrCode(r.body) != "ODR-E5001" || fallback(r) != "" {
						t.Errorf("%s: %d %s %q, erwartet 404 ODR-E5001 ohne Kennzeichnung", name, r.status, odrCode(r.body), fallback(r))
					}
					up.keinKontakt(t, name, vor)
				}
			}
		}
	}
	if n != 40 {
		t.Errorf("%d Fälle, erwartet 10 × 2 × 2", n)
	}
}

// erschoepft fängt M17: ein erschöpfter Key (ODR-E5002) erreicht den Upstream nie und kehrt unverändert zurück: der vierte `$count` nach drei
// aufgezeichneten (7, 11, 13) ist 410, auch bei der offensten Policy.
func (w *m4Welt) erschoepft(t *testing.T) {
	up := neuerReferenzdienst(t)
	c := w.startHybrid(t, up, offen(up, false)...)
	hdr := w.replayKopf(w.wsTeil, w.sessionID, "m17")
	hdr["Accept"] = acceptText
	for i, soll := range zaehlerFolge() {
		vor := up.stand()
		r := c.sende("GET", m3Pfad+"/$count", "", "", hdr)
		if r.status != 200 || string(r.body) != fmt.Sprint(soll) || fallback(r) != "replay" {
			t.Errorf("$count Nr. %d: %d %q, Kennzeichnung %q, erwartet 200 %d replay", i+1, r.status, r.body, fallback(r), soll)
		}
		up.keinKontakt(t, fmt.Sprintf("$count Nr. %d", i+1), vor)
	}
	vor := up.stand()
	r := c.sende("GET", m3Pfad+"/$count", "", "", hdr)
	if r.status != 410 || odrCode(r.body) != "ODR-E5002" || fallback(r) != "" {
		t.Errorf("vierter $count: %d %s %q, erwartet 410 ODR-E5002 ohne Kennzeichnung", r.status, odrCode(r.body), fallback(r))
	}
	up.keinKontakt(t, "vierter $count", vor)
}

// replayFallback fängt M18: Replay-Fallback: Methode (6) × Kategorie (2) × Header {vollständig, fehlt}; mit Header ist die Antwort die aufgezeichnete mit
// der Kennzeichnung replay und ohne Verbindung, auch bei geöffnetem Upstream; ohne die beiden Session-Header ODR-E5003 aus dem Replay, und eine
// Antwort des Mock braucht sie nie. Vorbedingung je Fall: der Modus mock beantwortet den Request mit dem Status des Auslösers.
func (w *m4Welt) replayFallback(t *testing.T) {
	mc := w.startMock(t)
	faelle := m4ReplayFaelle()
	if len(faelle) != 11 {
		t.Fatalf("%d Austausche, erwartet 6 × 2 ohne POST × coverage_insufficient", len(faelle))
	}
	for _, f := range faelle {
		ws := w.ws
		if f.teil {
			ws = w.wsTeil
		}
		if r := mc.sende(f.methode, f.pfad, f.body, "", kopf(ws)); r.status != f.status {
			t.Fatalf("%s: der Mock antwortet %d, der Fall braucht %d", f.name, r.status, f.status)
		}
	}
	n := 0
	for _, kat := range []string{"unsupported", "coverage_insufficient"} {
		up := neuerReferenzdienst(t)
		c := w.startHybrid(t, up, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+up.host(),
			"FALLBACK_WRITE_ALLOW="+routen(m4Mutierend(), "/odata/**"), "HYBRID_REPLAY_FALLBACK="+kat, "ETAG_ENABLED=true")
		for _, f := range faelle {
			if f.kategorie != kat {
				continue
			}
			ws := w.ws
			if f.teil {
				ws = w.wsTeil
			}
			aufn := w.aufgenommen[f.name]
			for _, mitKopf := range []bool{true, false} {
				n++
				name := fmt.Sprintf("%s kopf=%v", f.name, mitKopf)
				hdr := kopf(ws)
				if mitKopf {
					hdr = w.replayKopf(ws, w.s3, "m18-"+fmt.Sprint(n))
				}
				vor := up.stand()
				r := c.sende(f.methode, f.pfad, f.body, "", hdr)
				if mitKopf {
					if r.status != aufn.status || !bytes.Equal(r.body, aufn.body) || fallback(r) != "replay" {
						t.Errorf("%s: %d %q, Kennzeichnung %q; aufgezeichnet %d %q, erwartet replay", name, r.status, r.body, fallback(r), aufn.status, aufn.body)
					}
				} else if r.status != 400 || f.methode != "HEAD" && odrCode(r.body) != "ODR-E5003" || fallback(r) != "" {
					t.Errorf("%s: %d %s %q, erwartet 400 ODR-E5003 aus dem Replay", name, r.status, odrCode(r.body), fallback(r))
				}
				up.keinKontakt(t, name, vor)
			}
		}
		// Ein Treffer des Mock braucht die Session-Header nie: die Prüfung gehört dem Replay-Fallback und läuft nicht vor dem Mock.
		vor := up.stand()
		if r := c.sende("GET", m3Pfad+"?$orderby=Id", "", "", kopf(w.ws)); r.status != 200 || fallback(r) != "" {
			t.Errorf("%s: Treffer des Mock ohne Session-Header: %d, Kennzeichnung %q", kat, r.status, fallback(r))
		}
		up.keinKontakt(t, kat+": Treffer des Mock", vor)
	}
	if n != 2*len(faelle) {
		t.Errorf("%d Fälle, erwartet %d", n, 2*len(faelle))
	}
}

// gleichzeitig fängt M19: 64 gleichzeitige Anfragen mit Upstream-Fallback, ein Start der Anfragen über einen Kanal statt über Schlaf; die Summe der
// Zähler des Referenzdienstes ist die Zahl der Anfragen, jede Antwort trägt die Kennzeichnung (Wirkungs-Fänger neben dem Detektor von make test-race).
func (w *m4Welt) gleichzeitig(t *testing.T) {
	const anzahl = 64
	up := neuerReferenzdienst(t)
	up.jedeVerbindungNeu.Store(true)
	c := w.startHybrid(t, up, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+up.host())
	vor := up.stand()
	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	var antworten []antwort
	for range anzahl {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			r := c.sende("GET", m3Pfad, "", "", kopf(w.wsTeil))
			mu.Lock()
			antworten = append(antworten, r)
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()
	for i, r := range antworten {
		if r.status != http.StatusOK || fallback(r) != "upstream" {
			t.Errorf("Antwort %d: %d, Kennzeichnung %q", i, r.status, fallback(r))
		}
	}
	up.erwarte(t, "64 gleichzeitige Anfragen", vor, anzahl, map[string]int64{"GET": anzahl})
}
