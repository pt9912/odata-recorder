package e2e_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/adapters/driven/sqlite"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// paar ist ein aufgezeichneter Austausch aus Sicht des Clients.
type paar struct {
	a anfrage
	w antwort
}

// lauf ist ein Rundlauf Record → Upstream stoppen → Replay auf einem DATA_DIR (SPEC-037 Nr. 5, 6).
type lauf struct {
	t       *testing.T
	ref     *referenz
	dir     string
	dienst  string
	session string
	// aufnahme sind die Austausche der Aufnahme in der Reihenfolge der Requests.
	aufnahme []paar
	rec      *prozess
	// workspace ist die Kennung aus dem Übergang.
	workspace string
}

func neuerLauf(t *testing.T, ref *referenz) *lauf {
	t.Helper()
	binaryPfad(t)
	return &lauf{t: t, ref: ref, dir: t.TempDir()}
}

func (l *lauf) konfig(modus string, extra ...string) konfig {
	return konfig{modus: modus, dir: l.dir, upstream: l.ref.url(), dienst: l.dienst, extra: extra}
}

// nimmAuf startet das Binary im Modus record und wartet auf die Bereitschaft; der Request der
// Bereitschaft (`$metadata`) ist der erste Austausch der Aufnahme.
func (l *lauf) nimmAuf(extra ...string) *prozess {
	l.t.Helper()
	l.rec = starte(l.t, l.konfig("record", extra...))
	l.aufnahme = append(l.aufnahme, paar{anfrage{pfad: "/odata/$metadata"}, l.rec.bereit()})
	return l.rec
}

// frage stellt a an den Recorder und merkt den Austausch.
func (l *lauf) frage(a anfrage) antwort {
	l.t.Helper()
	w := holen(l.t, l.rec.adresse, a)
	l.aufnahme = append(l.aufnahme, paar{a, w})
	return w
}

// beendeAufnahme sendet SIGTERM (Exit 0) und liest die Kennung der ersten Session der Ablage.
func (l *lauf) beendeAufnahme() {
	l.t.Helper()
	if code := l.rec.beende(); code != 0 {
		l.t.Fatalf("SIGTERM: Exit %d\nstderr: %s", code, l.rec.fehl)
	}
	l.session = l.sessionDerAblage()
}

// sessionDerAblage liest die älteste Session der Ablage mit dem Adapter, nachdem der Prozess beendet ist.
func (l *lauf) sessionDerAblage() string { return l.sessionen()[0] }

// sessionen liest die Kennungen der Sessions der Ablage nach ihrer Erzeugung.
func (l *lauf) sessionen() []string {
	l.t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, filepath.Join(l.dir, "odata-recorder.db"), sqlite.Options{})
	if err != nil {
		l.t.Fatalf("Ablage öffnen: %v", err)
	}
	defer func() { _ = db.Close() }()
	dienst := l.dienst
	if dienst == "" {
		dienst = "ref"
	}
	ss, err := db.Sessions().List(ctx, dienst)
	if err != nil || len(ss) == 0 {
		l.t.Fatalf("Sessions der Ablage: %v, %v", ss, err)
	}
	slices.SortFunc(ss, func(a, b model.Session) int { return a.CreatedAt.Compare(b.CreatedAt) })
	ids := make([]string, len(ss))
	for i, s := range ss {
		ids[i] = s.ID
	}
	return ids
}

// uebergang führt das Hilfsprogramm aus und liefert stdout, stderr und den Exit-Code.
func uebergang(t *testing.T, args ...string) (stdout, stderr string, exit int) {
	t.Helper()
	p := os.Getenv("E2E_UEBERGANG")
	if p == "" {
		t.Fatal("E2E_UEBERGANG nicht gesetzt")
	}
	var o, e bytes.Buffer
	cmd := exec.Command(p, args...)
	cmd.Env, cmd.Stdout, cmd.Stderr = []string{}, &o, &e
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		exit = ee.ExitCode()
	default:
		t.Fatalf("Übergang: %v", err)
	}
	return o.String(), e.String(), exit
}

// importiere legt Import und Workspace an; stdout trägt `workspace=<ID>` und `session=<ID>`.
func (l *lauf) importiere() {
	l.t.Helper()
	out, errOut, code := uebergang(l.t, "import", l.dir, l.session)
	if code != 0 {
		l.t.Fatalf("Übergang import: Exit %d\nstderr: %s", code, errOut)
	}
	if !strings.Contains(out, "session="+l.session+"\n") {
		l.t.Fatalf("Übergang import: stdout ohne session=%s: %q", l.session, out)
	}
	for _, z := range strings.Split(out, "\n") {
		if w, ok := strings.CutPrefix(z, "workspace="); ok && w != "" {
			l.workspace = w
		}
	}
	if l.workspace == "" {
		l.t.Fatalf("Übergang import: stdout ohne workspace=: %q", out)
	}
}

// spieleAb startet das Binary im Modus replay mit dem Referenzdienst als UPSTREAM_URL (ein Kontakt
// ist am Zähler sichtbar) und wartet auf die Verbindung.
func (l *lauf) spieleAb(extra ...string) *prozess {
	l.t.Helper()
	p := starte(l.t, l.konfig("replay", extra...))
	p.bereit()
	return p
}

// replayHeader sind die drei Header der Zuordnung (SPEC-135 Nr. 5).
func (l *lauf) replayHeader(replaySession string) map[string][]string {
	return map[string][]string{
		"X-Recorder-Workspace":         {l.workspace},
		"X-Recorder-Recording-Session": {l.session},
		"X-Recorder-Replay-Session":    {replaySession},
	}
}

// zusammen vereinigt Header; ein Name aus b ersetzt den aus a.
func zusammen(a, b map[string][]string) map[string][]string {
	out := map[string][]string{}
	for _, m := range []map[string][]string{a, b} {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// replay stellt a mit der Zuordnung an p.
func (l *lauf) replay(p *prozess, a anfrage, replaySession string) antwort {
	l.t.Helper()
	a.header = zusammen(a.header, l.replayHeader(replaySession))
	return holen(l.t, p.adresse, a)
}

// ausnahme ist die Menge der Header, die der Vergleich von Aufnahme und Wiedergabe auslässt (SPEC-135
// Nr. 13): Länge und Coding der Leitung, Date, Hop-by-Hop, die Standardmengen von VOLATILE_HEADERS und
// REDACT_HEADERS.
func ausnahme() map[string]bool {
	m := map[string]bool{}
	for _, n := range []string{"Content-Length", "Content-Encoding", "Date", "Connection", "Keep-Alive", "Proxy-Connection",
		"Te", "Trailer", "Transfer-Encoding", "Upgrade", "Server", "Via", "Age", "X-Request-Id", "X-Correlation-Id", "Traceparent",
		"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie", "X-Api-Key", "X-Auth-Token", "X-Csrf-Token"} {
		m[http.CanonicalHeaderKey(n)] = true
	}
	return m
}

func sichtbar(h http.Header) map[string][]string {
	out := map[string][]string{}
	for n, v := range h {
		if !ausnahme()[http.CanonicalHeaderKey(n)] {
			out[http.CanonicalHeaderKey(n)] = v
		}
	}
	return out
}

// unterschied nennt den ersten Unterschied zwischen Aufnahme und Wiedergabe, sonst "".
func unterschied(aufnahme, wieder antwort) string {
	if aufnahme.status != wieder.status {
		return "Status"
	}
	if !bytes.Equal(aufnahme.body, wieder.body) {
		return "Body"
	}
	a, w := sichtbar(aufnahme.header), sichtbar(wieder.header)
	if len(a) != len(w) {
		return "Menge der Header"
	}
	for n, v := range a {
		if !slices.Equal(v, w[n]) {
			return "Header " + n
		}
	}
	return ""
}

// pruefeGleich vergleicht die Wiedergabe von a mit der Aufnahme w und meldet den Unterschied.
func pruefeGleich(t *testing.T, name string, aufnahme, wieder antwort) {
	t.Helper()
	if d := unterschied(aufnahme, wieder); d != "" {
		t.Errorf("%s: Wiedergabe weicht von der Aufnahme ab: %s (Aufnahme %d %q, Wiedergabe %d %q)",
			name, d, aufnahme.status, aufnahme.body, wieder.status, wieder.body)
	}
}

// pruefeFehler prüft Status und Fehlercode einer Antwort des Recorders.
func pruefeFehler(t *testing.T, name string, a antwort, status int, code string) {
	t.Helper()
	if a.status != status || !bytes.Contains(a.body, []byte(code)) {
		t.Errorf("%s: %d %q, erwartet %d mit %s", name, a.status, a.body, status, code)
	}
}

func contains(s, teil string) bool { return strings.Contains(s, teil) }
