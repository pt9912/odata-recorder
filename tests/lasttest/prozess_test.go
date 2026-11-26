package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// zaehlTransport zählt die beendeten Anfragen und hält die Replay-Sessions fest.
type zaehlTransport struct {
	inner    *http.Transport
	fertig   atomic.Int32
	mu       sync.Mutex
	sessions []string
}

func (z *zaehlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	z.mu.Lock()
	z.sessions = append(z.sessions, req.Header.Get(headerReplay))
	z.mu.Unlock()
	defer z.fertig.Add(1)
	return z.inner.RoundTrip(req)
}

// hilfsstand ist ein Probestand, dessen Binary ein echter Prozess ist: befehl liefert sein Kommando zur Umgebung; Port, Verbindung und
// Transport sind die des Programms. Die Uhr geht im Takt erst weiter, wenn die fälligen Anfragen beendet sind; zwischen zwei
// Verbindungsversuchen vergeht die Zeit pauseBereit der wirklichen Zeit.
type hilfsstand struct {
	*probestand
	z *zaehlTransport

	mu       sync.Mutex
	prozesse []*prozess
	signale  []os.Signal
}

func neuerHilfsstand(t *testing.T, p Profil, pauseBereit time.Duration, befehl func(env []string) (*prozess, error)) *hilfsstand {
	t.Helper()
	h := &hilfsstand{probestand: neuerProbestand(t, p, "replay"), z: &zaehlTransport{inner: &http.Transport{MaxIdleConnsPerHost: 64}}}
	t.Cleanup(h.z.inner.CloseIdleConnections)
	var takte atomic.Int32
	h.uhr.vorher = func(d time.Duration, _ time.Time) {
		if d <= bereitAbstand {
			pause(pauseBereit)
			return
		}
		n := takte.Add(1)
		bisWahr(func() bool { return h.z.fertig.Load() >= n })
	}
	h.ein.Port, h.ein.Verbinde, h.ein.Transport = freierPort, verbindeTCP, h.z
	h.ein.Starte = func(_ context.Context, env []string) (*prozess, error) {
		// Je Prozess zählen die Schritte des Takts ab den schon beendeten Anfragen.
		takte.Store(h.z.fertig.Load())
		pr, err := befehl(env)
		if err != nil {
			return nil, err
		}
		signal := pr.Signal
		pr.Signal = func(s os.Signal) error {
			h.mu.Lock()
			h.signale = append(h.signale, s)
			h.mu.Unlock()
			return signal(s)
		}
		h.mu.Lock()
		h.prozesse = append(h.prozesse, pr)
		h.mu.Unlock()
		return pr, nil
	}
	t.Cleanup(func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		for _, pr := range h.prozesse {
			if !beendet(pr.Ende) {
				_ = pr.Signal(syscall.SIGKILL)
			}
		}
	})
	return h
}

func (h *hilfsstand) gesendet() []os.Signal {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.signale)
}

// warteExit fährt den Ablauf und stellt die Uhr vor, wo eine Frist es verlangt: nach dem Ende aller Anfragen in 10 s, wenn nachlauf gesetzt
// ist, und nach einem SIGTERM ohne SIGKILL in 15 s, wenn stur gesetzt ist.
func (h *hilfsstand) warteExit(t *testing.T, nachlauf, stur bool) int {
	t.Helper()
	exit := h.fahre(context.Background())
	for i := 0; i < 4000; i++ {
		select {
		case code := <-exit:
			return code
		default:
		}
		sig := h.gesendet()
		switch {
		case stur && len(sig) == 1:
			pause(20 * time.Millisecond)
			h.uhr.Vorstellen(15 * time.Second)
		case nachlauf && len(sig) == 0 && h.z.fertig.Load() >= int32(h.ein.Profil.Anfragen()):
			pause(20 * time.Millisecond)
			h.uhr.Vorstellen(10 * time.Second)
		}
		pause(5 * time.Millisecond)
	}
	t.Fatalf("der Ablauf endet nicht: Meldungen %q", h.fehler.String())
	return 0
}

func hilfe(modus, vorweg string) func(env []string) (*prozess, error) {
	return func(env []string) (*prozess, error) { return starteProzess(hilfeBefehl(env, modus, vorweg)) }
}

// TestProzess (SPEC-216 Nr. 9 b bis d, 13 c): acht Formen des Binary als eigener Prozess; je Form grund, Meldung, Hinweis, Zahl der
// Anfragen und die Signale; ein Ende durch ein Signal zählt 128 plus dessen Nummer.
func TestProzess(t *testing.T) {
	wache(t, 120*time.Second)
	term, kill := []os.Signal{syscall.SIGTERM}, []os.Signal{syscall.SIGTERM, syscall.SIGKILL}
	formen := []struct {
		name           string
		befehl         func(env []string) (*prozess, error)
		pause          time.Duration
		nachlauf, stur bool
		grund, meldung string
		hinweis        string
		anfragen       string
		signale        []os.Signal
	}{
		{"Start scheitert", func(env []string) (*prozess, error) {
			return starteProzess(binaerBefehl(context.Background(), "/gibt/es/nicht", env))
		}, 0, false, false, "binary", "lasttest: Binary startet nicht\n", "", "0", nil},
		{"endet vor der Bereitschaft", hilfe("exit1", ""), 5 * time.Millisecond, false, false, "binary",
			"lasttest: Binary endete (Exit 1)\n", "", "0", nil},
		{"endet durch SIGKILL", func(env []string) (*prozess, error) {
			cmd := hilfeBefehl(env, "", "")
			cmd.Path, cmd.Args = "/bin/sh", []string{"sh", "-c", "kill -KILL $$"}
			return starteProzess(cmd)
		}, 5 * time.Millisecond, false, false, "binary", "lasttest: Binary endete (Exit 137)\n", "", "0", nil},
		{"lauscht nie", func(env []string) (*prozess, error) {
			cmd := hilfeBefehl(env, "", "")
			cmd.Path, cmd.Args = "/bin/sleep", []string{"sleep", "60"}
			return starteProzess(cmd)
		}, 0, false, false, "binary", "lasttest: Binary startet nicht\n",
			"lasttest: Hinweis betriebsart=replay binary_exit=143", "0", term},
		{"stirbt im Lauf", hilfe("stirbt", ""), 5 * time.Millisecond, true, false, "binary",
			"lasttest: Binary endete (Exit 3)\n", "", "40", nil},
		{"SIGTERM mit 0", hilfe("term0", ""), 5 * time.Millisecond, false, false, "ok", "", "", "40", term},
		{"SIGTERM mit 1", hilfe("term1", ""), 5 * time.Millisecond, false, false, "ok", "",
			"lasttest: Hinweis betriebsart=replay binary_exit=1", "40", term},
		{"ignoriert SIGTERM", hilfe("stur", ""), 5 * time.Millisecond, false, true, "ok", "",
			"lasttest: Hinweis betriebsart=replay binary_kill=ja", "40", kill},
	}
	n := 0
	for _, f := range formen {
		h := neuerHilfsstand(t, kurzesProfil("klein", 1), f.pause, f.befehl)
		code := h.warteExit(t, f.nachlauf, f.stur)
		m := h.zeilenVon("Messung")
		if len(m) != 1 {
			t.Fatalf("%s: Zeilen %q, Meldungen %q", f.name, h.zeilen(), h.fehler.String())
		}
		want := 0
		if f.grund != "ok" {
			want = 3
		}
		if code != want || feld(m[0], "grund") != f.grund || feld(m[0], "anfragen") != f.anfragen || h.fehler.String() != f.meldung {
			t.Errorf("%s: Exit %d, %q, Meldungen %q; erwartet Exit %d, grund %s, anfragen %s, %q", f.name, code, m[0], h.fehler.String(), want, f.grund,
				f.anfragen, f.meldung)
		}
		if hinweise := strings.Join(h.zeilenVon("Hinweis"), "|"); hinweise != f.hinweis {
			t.Errorf("%s: Hinweis %q, erwartet %q", f.name, hinweise, f.hinweis)
		}
		if !slices.Equal(h.gesendet(), f.signale) {
			t.Errorf("%s: Signale %v, erwartet %v", f.name, h.gesendet(), f.signale)
		}
		if f.name == "stirbt im Lauf" && (feld(m[0], "fehler") == "0" || feld(m[0], "spans") == "40") {
			t.Errorf("%s: %q; erwartet Fehler und fehlende Spans", f.name, m[0])
		}
		if feld(m[0], "gueltig") != map[bool]string{true: "ja", false: "nein"}[f.grund == "ok"] || len(h.zeilenVon("Last")) != 1 {
			t.Errorf("%s: Zeilen %q", f.name, h.zeilen())
		}
		n++
	}
	if n != 8 {
		t.Errorf("%d Formen, erwartet 8", n)
	}
}

// TestStroeme (SPEC-216 Nr. 9 e, f, Wirkungs-Fänger): das Binary schreibt vor dem Lauschen je 8 MiB auf stdout und stderr; beide Läufe enden
// gültig, jeder mit eigenem Prozess und eigener Replay-Session.
func TestStroeme(t *testing.T) {
	wache(t, 60*time.Second)
	flut := "head -c 8388608 /dev/zero; head -c 8388608 /dev/zero >&2; echo >&2"
	h := neuerHilfsstand(t, kurzesProfil("klein", 2), 5*time.Millisecond, hilfe("term0", flut))
	code := h.warteExit(t, false, false)
	m := h.zeilenVon("Messung")
	if code != 0 || len(m) != 2 || h.fehler.Len() != 0 {
		t.Fatalf("Exit %d, Zeilen %q, Meldungen %q", code, h.zeilen(), h.fehler.String())
	}
	for _, z := range m {
		if feld(z, "grund") != "ok" || feld(z, "fremde_zeilen") != "1" || feld(z, "spans") != "40" {
			t.Errorf("Zeile %q; erwartet grund=ok, fremde_zeilen=1, spans=40", z)
		}
	}
	if len(h.prozesse) != 2 || h.prozesse[0].PID == h.prozesse[1].PID {
		t.Errorf("%d Prozesse, erwartet 2 verschiedene", len(h.prozesse))
	}
	sessions := slices.Compact(slices.Clone(h.z.sessions))
	if !slices.Equal(sessions, []string{"lasttest-replay-1-0", "lasttest-replay-2-0"}) {
		t.Errorf("Replay-Sessions %v", sessions)
	}
}

// bruchStrom liefert nach nach Aufrufen von Read genau einmal einen Fehler und zählt danach die Bytes, die gelesen werden.
type bruchStrom struct {
	io.ReadCloser
	nach int
	n    int
	rest atomic.Int64
	brch atomic.Bool
}

var errBruch = errors.New("Lesen gescheitert")

func (b *bruchStrom) Read(p []byte) (int, error) {
	if !b.brch.Load() && b.n == b.nach {
		b.brch.Store(true)
		return 0, errBruch
	}
	n, err := b.ReadCloser.Read(p)
	b.n++
	if b.brch.Load() {
		b.rest.Add(int64(n))
	}
	return n, err
}

// TestLesefehler (SPEC-216 Nr. 10 a): ein Fehler des Stroms von stderr nach allen Gesamtspans oder davor wird gemeldet, der Rest gelesen, und
// der Lauf trägt grund=zaehlung; ein Fehler von stdout hat keine Wirkung.
func TestLesefehler(t *testing.T) {
	wache(t, 60*time.Second)
	for _, f := range []struct {
		name  string
		nach  int
		davor bool
	}{{"nach allen Gesamtspans", 80, false}, {"davor", 10, true}} {
		s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
		var strom *bruchStrom
		s.dienst.stderr = func(r io.ReadCloser) io.ReadCloser {
			strom = &bruchStrom{ReadCloser: r, nach: f.nach}
			return strom
		}
		s.dienst.signal = func(b *ersatzBinaer, _ os.Signal) { b.zeile("nach dem Lauf") }
		exit := s.fahre(context.Background())
		if f.davor {
			treibeBis(t, s.uhr, 10*time.Second, s.dienst.alleBeendet(40), s.dienst.signaleBei(1))
		}
		code := warteAuf(t, exit, 30*time.Second)
		m := s.zeilenVon("Messung")
		// Davor fehlen die Spans hinter dem Fehler; danach sind alle 40 gelesen, und allein der Fehler macht den Lauf ungültig.
		if code != 3 || len(m) != 1 || feld(m[0], "grund") != "zaehlung" || (feld(m[0], "spans") == "40") == f.davor ||
			s.fehler.String() != "lasttest: Lesefehler auf stderr des Binary\n" {
			t.Errorf("%s: Exit %d, %q, Meldungen %q", f.name, code, m, s.fehler.String())
		}
		if rest := strom.rest.Load(); rest < int64(len("nach dem Lauf\n")) {
			t.Errorf("%s: %d Bytes nach dem Fehler gelesen", f.name, rest)
		}
	}
	s := neuerProbestand(t, kurzesProfil("klein", 1), "replay")
	s.dienst.stdout = func(r io.ReadCloser) io.ReadCloser { return &bruchStrom{ReadCloser: r} }
	code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
	if m := s.zeilenVon("Messung"); code != 0 || len(m) != 1 || feld(m[0], "grund") != "ok" || s.fehler.Len() != 0 {
		t.Errorf("Fehler auf stdout: Exit %d, %q, Meldungen %q", code, s.zeilen(), s.fehler.String())
	}
}

// bereitschaftsFall fährt warteBereit mit der Uhr des Tests; verbinde gelingt, sobald die Uhr ab (seit dem Aufruf) erreicht hat.
func bereitschaftsFall(ctx context.Context, ab time.Duration, ende <-chan struct{}) (*fakeUhr, <-chan bereitschaft) {
	u := neueFakeUhr()
	start := u.Jetzt()
	erg := make(chan bereitschaft, 1)
	go func() {
		erg <- warteBereit(ctx, u, func(context.Context, string) error {
			if u.Jetzt().Sub(start) >= ab {
				return nil
			}
			return errBruch
		}, "127.0.0.1:1", ende)
	}()
	return u, erg
}

// TestBereitschaft (SPEC-216 Nr. 9 b, 4 c): ohne Verbindung läuft die Frist erst bei 10 s der Uhr ab (bei 10 s − 1 ns nicht); eine
// Verbindung bei 10 s − 1 ns ist Bereitschaft, eine bei 10 s nicht; ein beendetes Binary und ein beendeter Kontext enden sofort.
func TestBereitschaft(t *testing.T) {
	wache(t, 60*time.Second)
	nie := time.Hour
	warten := func(u *fakeUhr, n int) {
		bisWahr(func() bool { return u.wartende() >= n })
	}
	u, erg := bereitschaftsFall(context.Background(), nie, nil)
	warten(u, 2)
	u.Vorstellen(10*time.Second - time.Nanosecond)
	warten(u, 2)
	bleibtOffen(t, erg, "warteBereit vor 10 s")
	u.Vorstellen(time.Nanosecond)
	if got := warteAuf(t, erg, 5*time.Second); got != nichtBereit {
		t.Errorf("ohne Verbindung bei 10 s: %d", got)
	}
	u, erg = bereitschaftsFall(context.Background(), 10*time.Second-time.Nanosecond, nil)
	warten(u, 2)
	u.Vorstellen(10*time.Second - time.Nanosecond)
	if got := warteAuf(t, erg, 5*time.Second); got != bereit {
		t.Errorf("Verbindung bei 10 s − 1 ns: %d", got)
	}
	u, erg = bereitschaftsFall(context.Background(), 10*time.Second, nil)
	warten(u, 2)
	u.Vorstellen(10 * time.Second)
	if got := warteAuf(t, erg, 5*time.Second); got != nichtBereit {
		t.Errorf("Verbindung bei 10 s: %d", got)
	}
	zu := make(chan struct{})
	close(zu)
	_, erg = bereitschaftsFall(context.Background(), 0, zu)
	if got := warteAuf(t, erg, 5*time.Second); got != vorherBeendet {
		t.Errorf("Binary beendet, Verbindung gelingt: %d", got)
	}
	spaeter := make(chan struct{})
	u, erg = bereitschaftsFall(context.Background(), nie, spaeter)
	warten(u, 2)
	close(spaeter)
	if got := warteAuf(t, erg, 5*time.Second); got != vorherBeendet {
		t.Errorf("Binary endet im Warten: %d", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, erg = bereitschaftsFall(ctx, nie, nil)
	if got := warteAuf(t, erg, 5*time.Second); got != nichtBereit {
		t.Errorf("Kontext beendet: %d", got)
	}
}

// signalProzess ist ein Prozess ohne Kommando: er endet auf SIGKILL, auf SIGTERM allein, wenn aufTerm gesetzt ist.
func signalProzess(aufTerm bool) (*prozess, func() []os.Signal) {
	ende := make(chan struct{})
	var mu sync.Mutex
	var signale []os.Signal
	var einmal sync.Once
	return &prozess{
		Ende: ende, Exit: func() int { return 0 },
		Signal: func(s os.Signal) error {
			mu.Lock()
			defer mu.Unlock()
			signale = append(signale, s)
			if s == syscall.SIGKILL || aufTerm {
				einmal.Do(func() { close(ende) })
			}
			return nil
		},
	}, func() []os.Signal {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(signale)
	}
}

// TestBeendeBinaer (SPEC-216 Nr. 9 d): SIGTERM, SIGKILL erst bei 15 s der Uhr (bei 15 s − 1 ns nicht), auch wenn der Kontext schon beendet
// ist; ein Prozess, der auf SIGTERM endet, bekommt kein SIGKILL.
func TestBeendeBinaer(t *testing.T) {
	wache(t, 60*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, c := range []context.Context{context.Background(), ctx} {
		u := neueFakeUhr()
		p, signale := signalProzess(false)
		erg := make(chan bool, 1)
		go func() { erg <- beendeBinaer(c, u, p) }()
		bisWahr(func() bool { return u.wartende() >= 1 })
		u.Vorstellen(15*time.Second - time.Nanosecond)
		bleibtOffen(t, erg, "beendeBinaer vor 15 s")
		if s := signale(); !slices.Equal(s, []os.Signal{syscall.SIGTERM}) {
			t.Errorf("vor 15 s: Signale %v", s)
		}
		u.Vorstellen(time.Nanosecond)
		if kill := warteAuf(t, erg, 5*time.Second); !kill || !slices.Equal(signale(), []os.Signal{syscall.SIGTERM, syscall.SIGKILL}) {
			t.Errorf("bei 15 s: %v, Signale %v", kill, signale())
		}
	}
	p, signale := signalProzess(true)
	if beendeBinaer(context.Background(), neueFakeUhr(), p) || !slices.Equal(signale(), []os.Signal{syscall.SIGTERM}) {
		t.Errorf("endet auf SIGTERM: Signale %v", signale())
	}
}

// TestBinaerBefehl (SPEC-216 Nr. 9 a, 4 d): das Kommando ist `<pfad> start` mit genau der Umgebung; das Ende des Kontexts beendet den
// Prozess nicht, SIGTERM tut es.
func TestBinaerBefehl(t *testing.T) {
	wache(t, 60*time.Second)
	env := umgebung(betriebMock, "127.0.0.1:4711")
	cmd := binaerBefehl(context.Background(), "/x/y", env)
	if !slices.Equal(cmd.Args, []string{"/x/y", "start"}) || !slices.Equal(cmd.Env, env) || len(cmd.Env) != 5 {
		t.Errorf("Args %v, Env %v", cmd.Args, cmd.Env)
	}
	skript := filepath.Join(t.TempDir(), "binaer")
	inhalt := "#!/bin/sh\n[ \"$1\" = start ] || exit 7\ntrap 'exit 4' TERM\necho da >&2\nwhile :; do sleep 0.05; done\n"
	if err := os.WriteFile(skript, []byte(inhalt), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p, err := starteProzess(binaerBefehl(ctx, skript, []string{"PATH=/bin:/usr/bin"}))
	if err != nil {
		t.Fatal(err)
	}
	zeile := make([]byte, 3)
	if _, err := io.ReadFull(p.Stderr, zeile); err != nil || string(zeile) != "da\n" {
		t.Fatalf("stderr %q, %v", zeile, err)
	}
	cancel()
	select {
	case <-p.Ende:
		t.Fatalf("der Kontext beendet den Prozess, Exit %d", p.Exit())
	case <-wartenKanal(200 * time.Millisecond):
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	warteAuf(t, p.Ende, 5*time.Second)
	if p.Exit() != 4 {
		t.Errorf("Exit %d, erwartet 4", p.Exit())
	}
	_, _ = p.Stdout.Close(), p.Stderr.Close()
	if _, err := starteProzess(binaerBefehl(context.Background(), skript+".fehlt", nil)); err == nil {
		t.Error("Start eines fehlenden Programms ohne Fehler")
	}
}
