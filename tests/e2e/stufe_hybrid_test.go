package e2e_test

import (
	"fmt"
	"strings"
	"testing"
)

// hybridWelt ist der Zustand vor dem Modus hybrid: aufgenommen (`$metadata`, die erste Seite von Products und die Entity 1), importiert mit der
// Voreinstellung von E2E_UPSTREAM_URL (die Coverage der Baseline ist `partial`, SPEC-037 Nr. 6), Workspace angelegt; der Referenzdienst läuft
// weiter und schließt nach jeder Antwort die Verbindung, damit Verbindungen und Requests gleich sind.
type hybridWelt struct {
	l   *lauf
	ref *referenz
}

func neueHybridWelt(t *testing.T) *hybridWelt {
	t.Helper()
	ref := neueReferenz(t)
	ref.jedeVerbindungNeu.Store(true)
	l := neuerLauf(t, ref)
	l.nimmAuf()
	l.frage(anfrage{pfad: collPfad})
	l.frage(anfrage{pfad: collPfad + "(1)"})
	l.beendeAufnahme()
	out, errOut, code := uebergangMit(t, nil, "import", l.dir, l.session)
	if code != 0 {
		t.Fatalf("Übergang import: Exit %d\nstderr: %s", code, errOut)
	}
	for _, z := range strings.Split(out, "\n") {
		if w, ok := strings.CutPrefix(z, "workspace="); ok && w != "" {
			l.workspace = w
		}
	}
	if l.workspace == "" {
		t.Fatalf("Übergang import: stdout ohne workspace=: %q", out)
	}
	return &hybridWelt{l: l, ref: ref}
}

// host ist der Eintrag für FALLBACK_ALLOW_HOSTS: Host und Port des Referenzdienstes.
func (w *hybridWelt) host() string { return strings.TrimPrefix(w.ref.url(), "http://") }

// start startet das Binary im Modus hybrid auf dem DATA_DIR des Laufs mit der UPSTREAM_URL des laufenden Referenzdienstes und wartet auf die
// Verbindung ohne Request (SPEC-037 Nr. 5).
func (w *hybridWelt) start(t *testing.T, extra ...string) *prozess {
	t.Helper()
	p := starte(t, w.l.konfig("hybrid", extra...))
	p.bereit()
	return p
}

// hole stellt a mit dem Workspace des Laufs an p; mitSession fügt die beiden Session-Header des Replay-Fallbacks hinzu.
func (w *hybridWelt) hole(t *testing.T, p *prozess, a anfrage, mitSession string) antwort {
	t.Helper()
	h := map[string][]string{wsKopf: {w.l.workspace}}
	if mitSession != "" {
		h = zusammen(h, map[string][]string{"X-Recorder-Recording-Session": {w.l.session}, "X-Recorder-Replay-Session": {mitSession}})
	}
	a.header = zusammen(a.header, h)
	return holen(t, p.adresse, a)
}

// hybridMethoden sind die Methoden, die der Test je Request unterscheidet.
func hybridMethoden() []string {
	return []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "MERGE", "OPTIONS", "TRACE", "CONNECT"}
}

// messung ist der Zählerstand des Referenzdienstes.
type messung struct {
	z       [2]int64
	methode map[string]int64
}

func (w *hybridWelt) messe() messung {
	m := messung{z: w.ref.zaehler(), methode: map[string]int64{}}
	for _, n := range hybridMethoden() {
		m.methode[n] = w.ref.nachMethode(n)
	}
	return m
}

// erwarte prüft, dass seit vor genau verb Verbindungen und je Methode von soll die Zahl der Requests dazukamen, sonst keine.
func (w *hybridWelt) erwarte(t *testing.T, name string, vor messung, verb int64, soll map[string]int64) {
	t.Helper()
	nach := w.messe()
	summe := int64(0)
	for _, n := range soll {
		summe += n
	}
	if got := nach.z[0] - vor.z[0]; got != verb {
		t.Errorf("%s: %d neue Verbindungen am Referenzdienst, erwartet %d", name, got, verb)
	}
	if got := nach.z[1] - vor.z[1]; got != summe {
		t.Errorf("%s: %d neue Requests am Referenzdienst, erwartet %d", name, got, summe)
	}
	for _, m := range hybridMethoden() {
		if got := nach.methode[m] - vor.methode[m]; got != soll[m] {
			t.Errorf("%s: %d neue Requests mit der Methode %s, erwartet %d", name, got, m, soll[m])
		}
	}
}

// beendeSauber sendet SIGTERM (Exit 0) und prüft, dass die Schleife keine Verbindung mehr annimmt (SPEC-139 Nr. 7, SPEC-138 Nr. 8).
func beendeSauber(t *testing.T, p *prozess) {
	t.Helper()
	if code := p.beende(); code != 0 {
		t.Errorf("SIGTERM: Exit %d, erwartet 0\nstderr: %s", code, p.fehl)
	}
	p.verweigert()
}

// TestE2EHybrid ist die Stufe von LH-FA-12 und LH-FA-04 (SPEC-037 Nr. 7): der Modus hybrid am gebauten Binary mit dem laufenden Referenzdienst als
// Upstream; das Orakel ist der Zähler des Referenzdienstes nach Methode.
func TestE2EHybrid(t *testing.T) {
	binaryPfad(t)
	w := neueHybridWelt(t)
	t.Run("Happy Path", w.happyPath)
	t.Run("Boundary", w.boundary)
	t.Run("Negative", w.negative)
}

// happyPath: ein Treffer des Mock (GET einer aufgezeichneten Entity) ergibt keine Verbindung und keine Kennzeichnung; eine lesende Anfrage an die
// Collection mit ODR-E6003 ergibt genau eine Verbindung mit der Methode GET und die Kennzeichnung upstream.
func (w *hybridWelt) happyPath(t *testing.T) {
	p := w.start(t, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+w.host())
	vor := w.messe()
	r := w.hole(t, p, anfrage{pfad: collPfad + "(1)"}, "")
	if r.status != 200 || r.header.Get("X-Recorder-Fallback") != "" || !contains(string(r.body), `"Anvil"`) {
		t.Errorf("Treffer des Mock: %d, X-Recorder-Fallback %q, Body %q", r.status, r.header.Get("X-Recorder-Fallback"), r.body)
	}
	w.erwarte(t, "Treffer des Mock", vor, 0, nil)
	r = w.hole(t, p, anfrage{pfad: collPfad}, "")
	if r.status != 200 || r.header.Get("X-Recorder-Fallback") != "upstream" || !contains(string(r.body), `"Anvil"`) {
		t.Errorf("Collection mit ODR-E6003: %d, X-Recorder-Fallback %q, Body %q", r.status, r.header.Get("X-Recorder-Fallback"), r.body)
	}
	w.erwarte(t, "Collection mit ODR-E6003", vor, 1, map[string]int64{"GET": 1})
	beendeSauber(t, p)
}

// boundary: ohne den Hauptschalter bleibt ODR-E6003; ein POST auf eine nicht unterstützte Operation ist ohne Eintrag ODR-E6002 und mit passendem Eintrag
// eine Verbindung mit der Methode POST; mit Replay-Fallback und den Session-Headern endet eine Anfrage, die die Aufzeichnung nicht trägt, als
// Replay-Miss ODR-E5001, und nur GET und HEAD erreichen danach den Referenzdienst.
func (w *hybridWelt) boundary(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		p := w.start(t, "HYBRID_FALLBACK=disabled", "FALLBACK_ALLOW_HOSTS="+w.host(), "FALLBACK_WRITE_ALLOW=POST /odata/**")
		vor := w.messe()
		pruefeFehler(t, "Collection bei disabled", w.hole(t, p, anfrage{pfad: collPfad}, ""), 422, "ODR-E6003")
		w.erwarte(t, "disabled", vor, 0, nil)
		beendeSauber(t, p)
	})
	t.Run("POST ohne Eintrag", func(t *testing.T) {
		p := w.start(t, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+w.host())
		vor := w.messe()
		a := anfrage{methode: "POST", pfad: collPfad + "(1)", body: `{"Id":9}`, header: map[string][]string{"Content-Type": {jsonTyp}}}
		pruefeFehler(t, "POST ohne Eintrag", w.hole(t, p, a, ""), 501, "ODR-E6002")
		w.erwarte(t, "POST ohne Eintrag", vor, 0, nil)
		beendeSauber(t, p)
	})
	t.Run("POST mit Eintrag", func(t *testing.T) {
		p := w.start(t, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+w.host(), "FALLBACK_WRITE_ALLOW=POST /odata/Products(1)")
		vor := w.messe()
		a := anfrage{methode: "POST", pfad: collPfad + "(1)", body: `{"Id":9}`, header: map[string][]string{"Content-Type": {jsonTyp}}}
		if r := w.hole(t, p, a, ""); r.header.Get("X-Recorder-Fallback") != "upstream" {
			t.Errorf("POST mit Eintrag: %d, X-Recorder-Fallback %q", r.status, r.header.Get("X-Recorder-Fallback"))
		}
		w.erwarte(t, "POST mit Eintrag", vor, 1, map[string]int64{"POST": 1})
		beendeSauber(t, p)
	})
	t.Run("Replay-Miss", func(t *testing.T) {
		p := w.start(t, "HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS="+w.host(), "HYBRID_REPLAY_FALLBACK=unsupported,coverage_insufficient",
			"FALLBACK_WRITE_ALLOW=POST /odata/**,PATCH /odata/**")
		n := 0
		// GET und HEAD an die Collection (ODR-E6003), ein anderer Request als die Aufzeichnung: nach dem Replay-Miss genau eine Verbindung.
		for _, m := range []string{"GET", "HEAD"} {
			n++
			vor := w.messe()
			r := w.hole(t, p, anfrage{methode: m, pfad: collPfad + "?$top=1"}, fmt.Sprintf("hybrid-%d", n))
			if r.header.Get("X-Recorder-Fallback") != "upstream" {
				t.Errorf("%s nach dem Replay-Miss: %d, X-Recorder-Fallback %q", m, r.status, r.header.Get("X-Recorder-Fallback"))
			}
			w.erwarte(t, m+" nach dem Replay-Miss", vor, 1, map[string]int64{m: 1})
		}
		// PATCH auf eine unbekannte Entity (ODR-E6003) und POST auf eine Entity (ODR-E6002): der Replay-Miss bleibt, auch mit Eintrag der Route.
		for _, f := range []struct{ methode, pfad string }{{"PATCH", collPfad + "(2)"}, {"POST", collPfad + "(1)"}} {
			n++
			vor := w.messe()
			a := anfrage{methode: f.methode, pfad: f.pfad, body: `{"Name":"x"}`, header: map[string][]string{"Content-Type": {jsonTyp}}}
			pruefeFehler(t, f.methode+" nach dem Replay-Miss", w.hole(t, p, a, fmt.Sprintf("hybrid-%d", n)), 404, "ODR-E5001")
			w.erwarte(t, f.methode+" nach dem Replay-Miss", vor, 0, nil)
		}
		beendeSauber(t, p)
	})
}

// negative: enabled ohne FALLBACK_ALLOW_HOSTS endet mit Exit 1, ODR-E2001 am Schlüssel HYBRID_FALLBACK und ohne Listener; enabled mit Host und ohne
// UPSTREAM_URL endet mit Exit 1 und ODR-E2001 am Schlüssel UPSTREAM_URL (SPEC-110 Nr. 16).
func (w *hybridWelt) negative(t *testing.T) {
	faelle := []struct {
		name, schluessel string
		upstream         string
		extra            []string
	}{
		{"enabled ohne Hosts", "HYBRID_FALLBACK", w.ref.url(), []string{"HYBRID_FALLBACK=enabled"}},
		{"enabled ohne UPSTREAM_URL", "UPSTREAM_URL", "", []string{"HYBRID_FALLBACK=enabled", "FALLBACK_ALLOW_HOSTS=" + w.host()}},
	}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			k := w.l.konfig("hybrid", f.extra...)
			k.upstream = f.upstream
			p := starte(t, k)
			p.warteEnde()
			if p.exit != 1 || !contains(p.fehl.String(), "ODR-E2001 "+f.schluessel) {
				t.Errorf("Exit %d, stderr %q; erwartet Exit 1 und ODR-E2001 %s", p.exit, p.fehl, f.schluessel)
			}
			p.verweigert()
		})
	}
}
