package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// treibUhr ist die fakeUhr, die sich beim Warten auf einen Zeitpunkt unter einer Sekunde Abstand selbst vorstellt (Takt und Abstand der
// Verbindungsversuche); längere Fristen warten auf Vorstellen. vorher läuft vor jedem eigenen Vorstellen mit dem Abstand und dem Ziel,
// extra verschiebt das Ziel um einen Überschuss.
type treibUhr struct {
	*fakeUhr
	vorher func(d time.Duration, ziel time.Time)
	extra  func(ziel time.Time) time.Duration
}

func neueTreibUhr() *treibUhr { return &treibUhr{fakeUhr: neueFakeUhr()} }

func (u *treibUhr) WarteBis(ctx context.Context, bis time.Time) error {
	d := bis.Sub(u.Jetzt())
	if d <= 0 || d >= time.Second {
		return u.fakeUhr.WarteBis(ctx, bis)
	}
	if u.vorher != nil {
		u.vorher(d, bis)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	ueber := time.Duration(0)
	if u.extra != nil {
		ueber = u.extra(bis)
	}
	if rest := bis.Sub(u.Jetzt()); rest+ueber > 0 {
		u.Vorstellen(rest + ueber)
	}
	return nil
}

// ereignisse hält die Aufrufe eines Laufs in ihrer Folge fest.
type ereignisse struct {
	mu  sync.Mutex
	vor []string
}

func (e *ereignisse) merke(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vor = append(e.vor, s)
}

func (e *ereignisse) liste() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.vor)
}

// ersatzLast ersetzt /proc/loadavg: der Aufruf mit der laufenden Nummer i (ab 0) liefert inhalt(i); halte läuft vor der Antwort.
type ersatzLast struct {
	n      atomic.Int64
	e      *ereignisse
	inhalt func(i int) string
	halte  func(i int)
}

func (l *ersatzLast) Lies() ([]byte, error) {
	i := int(l.n.Add(1) - 1)
	if l.e != nil {
		l.e.merke("last")
	}
	if l.halte != nil {
		l.halte(i)
	}
	if l.inhalt != nil {
		return []byte(l.inhalt(i)), nil
	}
	return []byte("1.00 1.00 1.00 1/10 5\n"), nil
}

// ersatzBinaer ist ein Lauf des Ersatz-Binary: seine Umgebung, seine Ströme und was an ihm geschah.
type ersatzBinaer struct {
	nr       int
	env      []string
	modus    string
	start    time.Time
	aus      *io.PipeWriter
	fehler   *io.PipeWriter
	ende     chan struct{}
	einmal   sync.Once
	exit     atomic.Int32
	nummer   atomic.Int32
	angekomm atomic.Int32
	fertig   atomic.Int32

	mu      sync.Mutex
	signale []os.Signal
	zeiten  map[int]time.Duration
}

// zeit ist die Zeit der Uhr seit dem Start, zu der die Anfrage k ankam.
func (b *ersatzBinaer) zeit(k int) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.zeiten[k]
}

func (b *ersatzBinaer) beende(code int) {
	b.einmal.Do(func() {
		b.exit.Store(int32(code))
		_, _ = b.aus.Close(), b.fehler.Close()
		close(b.ende)
	})
}

func (b *ersatzBinaer) gesendet() []os.Signal {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.signale)
}

// alleBeendet meldet, ob der laufende Prozess n Anfragen beendet hat.
func (e *ersatz) alleBeendet(n int32) func() bool {
	return func() bool {
		b := e.aktuell()
		return b != nil && b.fertig.Load() == n
	}
}

// signaleBei meldet, ob der laufende Prozess n Signale bekam.
func (e *ersatz) signaleBei(n int) func() bool {
	return func() bool {
		b := e.aktuell()
		return b != nil && len(b.gesendet()) >= n
	}
}

// zeile schreibt eine Zeile in den Strom von stderr; nach dem Ende wirkt sie nicht.
func (b *ersatzBinaer) zeile(s string) { _, _ = io.WriteString(b.fehler, s+"\n") }

// ms ist eine Dauer in Mikrosekunden in der Form der Span-Zeile.
func ms(mikros int) string { return fmt.Sprintf("%d.%03d", mikros/1000, mikros%1000) }

func kindZeile(modus, id string, mikros int) string {
	name := spanReplay
	if modus == string(betriebMock) {
		name = spanMock
	}
	return `{"ereignis":"span","msg":"` + name + `","span_id":"k-` + id + `","parent_id":"` + id + `","duration_ms":` + ms(mikros) + `}`
}

func gesamtZeile(modus, id string, mikros int) string {
	return `{"ereignis":"span","msg":"odata.gateway.request","span_id":"` + id + `","duration_ms":` + ms(mikros) + `,"recorder.mode":"` +
		modus + `"}`
}

// ersatz ist der Ersatz-Dienst: ein Binary und ein Transport im Prozess des Tests, die je Anfrage eine Span-Zeile mit vorgegebener Dauer
// schreiben und jede Anfrage zählen. Jede Funktion nil hat die Vorgabe: Status 200, Body `{}`, Gesamtspan 1000 µs, Kind 100 µs, SIGTERM
// beendet mit Exit 0. nr zählt die Prozesse, k die Anfragen eines Prozesses in der Folge der Ankunft (ab 0).
type ersatz struct {
	uhr      Uhr
	e        *ereignisse
	status   func(nr, k int) int
	body     func(nr, k int) string
	dauer    func(nr, k int) (gesamt, kind int)
	halte    func(nr, k int, req *http.Request)
	danach   func(b *ersatzBinaer, k int)
	selbst   func(nr, k int) (code int, ende bool)
	beiStart func(b *ersatzBinaer)
	signal   func(b *ersatzBinaer, s os.Signal)
	stumm    func(nr, k int) bool
	termExit int
	stur     bool
	stderr   func(io.ReadCloser) io.ReadCloser
	stdout   func(io.ReadCloser) io.ReadCloser

	mu       sync.Mutex
	laeufe   []*ersatzBinaer
	anfragen []string
}

func (e *ersatz) lauf(nr int) *ersatzBinaer {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.laeufe[nr]
}

func (e *ersatz) zahl() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.laeufe)
}

func (e *ersatz) aktuell() *ersatzBinaer {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.laeufe) == 0 {
		return nil
	}
	return e.laeufe[len(e.laeufe)-1]
}

func (e *ersatz) liste() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.anfragen)
}

func (e *ersatz) starte(_ context.Context, env []string) (*prozess, error) {
	ausL, ausS := io.Pipe()
	fehlerL, fehlerS := io.Pipe()
	b := &ersatzBinaer{env: env, aus: ausS, fehler: fehlerS, ende: make(chan struct{}), start: e.uhr.Jetzt(), zeiten: map[int]time.Duration{}}
	for _, s := range env {
		if m, ok := strings.CutPrefix(s, "MODE="); ok {
			b.modus = m
		}
	}
	e.mu.Lock()
	b.nr = len(e.laeufe)
	e.laeufe = append(e.laeufe, b)
	e.mu.Unlock()
	if e.e != nil {
		e.e.merke("start MODE=" + b.modus)
	}
	if e.beiStart != nil {
		go e.beiStart(b)
	}
	var aus, fehler io.ReadCloser = ausL, fehlerL
	if e.stdout != nil {
		aus = e.stdout(aus)
	}
	if e.stderr != nil {
		fehler = e.stderr(fehler)
	}
	return &prozess{
		PID: 100 + b.nr, Stdout: aus, Stderr: fehler, Ende: b.ende,
		Exit: func() int { <-b.ende; return int(b.exit.Load()) },
		Signal: func(s os.Signal) error {
			b.mu.Lock()
			b.signale = append(b.signale, s)
			b.mu.Unlock()
			if e.e != nil {
				e.e.merke("signal " + s.String())
			}
			if e.signal != nil {
				e.signal(b, s)
			}
			switch {
			case s == syscall.SIGKILL:
				b.beende(137)
			case s == syscall.SIGTERM && !e.stur:
				b.beende(e.termExit)
			}
			return nil
		},
	}, nil
}

// anfrageText ist eine Anfrage als Zeile: Methode, Ziel und die Header in sortierter Folge.
func anfrageText(methode, ziel string, header http.Header) string {
	var h []string
	for k, v := range header {
		h = append(h, k+": "+strings.Join(v, ","))
	}
	slices.Sort(h)
	return methode + " " + ziel + " " + strings.Join(h, "; ")
}

func (e *ersatz) RoundTrip(req *http.Request) (*http.Response, error) {
	b := e.aktuell()
	k := int(b.nummer.Add(1) - 1)
	b.mu.Lock()
	b.zeiten[k] = e.uhr.Jetzt().Sub(b.start)
	b.mu.Unlock()
	e.mu.Lock()
	e.anfragen = append(e.anfragen, anfrageText(req.Method, req.URL.Scheme+"://"+req.URL.Host+req.URL.RequestURI(), req.Header))
	e.mu.Unlock()
	b.angekomm.Add(1)
	if e.halte != nil {
		e.halte(b.nr, k, req)
	}
	defer b.fertig.Add(1)
	if beendet(b.ende) {
		return nil, errors.New("Verbindung abgelehnt")
	}
	status, body := http.StatusOK, "{}"
	if e.status != nil {
		status = e.status(b.nr, k)
	}
	if e.body != nil {
		body = e.body(b.nr, k)
	}
	if e.stumm == nil || !e.stumm(b.nr, k) {
		gesamt, kind := 1000, 100
		if e.dauer != nil {
			gesamt, kind = e.dauer(b.nr, k)
		}
		id := "g-" + strconv.Itoa(b.nr) + "-" + strconv.Itoa(k)
		if kind >= 0 {
			b.zeile(kindZeile(b.modus, id, kind))
		}
		b.zeile(gesamtZeile(b.modus, id, gesamt))
	}
	if e.danach != nil {
		e.danach(b, k)
	}
	if e.selbst != nil {
		if code, ende := e.selbst(b.nr, k); ende {
			b.beende(code)
		}
	}
	return &http.Response{
		StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: int64(len(body)), Request: req,
	}, nil
}

func (e *ersatz) verbinde(context.Context, string) error {
	if e.e != nil {
		e.e.merke("verbinde")
	}
	return nil
}

// kurzesProfil ist ein Profil des Tests: je Lauf 20 Anfragen Warm-up und 20 Messphase (Rate 20), 12 Lookups, 2 Sets zu 20 Entities.
func kurzesProfil(name string, laeufe int) Profil {
	p := kleinesProfil()
	p.Name, p.Laeufe, p.Warmup, p.Messphase = name, laeufe, time.Second, time.Second
	return p
}

// probestand ist ein Aufbau eines Ersatz-Dienstes samt Ablauf: Ausgabe, Meldungen, Aufrufe und die Eingaben an den Aufbau.
type probestand struct {
	uhr     *treibUhr
	e       *ereignisse
	dienst  *ersatz
	last    *ersatzLast
	ein     ablaufEingabe
	aus     bytes.Buffer
	fehler  bytes.Buffer
	aufbaue []aufbauEingabe
	cpu     []time.Duration
}

func neuerProbestand(t *testing.T, p Profil, betriebsart string) *probestand {
	t.Helper()
	s := &probestand{uhr: neueTreibUhr(), e: &ereignisse{}}
	s.dienst = &ersatz{uhr: s.uhr, e: s.e}
	s.last = &ersatzLast{e: s.e}
	// Der eigene Takt stellt die Uhr erst weiter, wenn jede fällige Anfrage am Ersatz angekommen ist.
	s.uhr.vorher = func(d time.Duration, ziel time.Time) {
		if b := s.dienst.aktuell(); b != nil && !beendet(b.ende) {
			n := int32(ziel.Sub(b.start) / (time.Second / time.Duration(p.Rate)))
			bisWahr(func() bool { return b.angekomm.Load() >= n })
		}
	}
	n := 0
	s.ein = ablaufEingabe{
		werte: werte{Profil: p, Betriebsart: betriebsart, ZielMs: 100, Startwert: 7, Kerne: 20},
		Uhr:   s.uhr, Last: s.last, DataDir: t.TempDir(),
		Aufbau: func(_ context.Context, a aufbauEingabe) (aufbauErgebnis, error) {
			s.e.merke("aufbau")
			s.aufbaue = append(s.aufbaue, a)
			return aufbauErgebnis{
				Session: "s-1", Baseline: "b-1", Workspace: "w-1", Exchanges: p.Exchanges, Entities: p.Sets * p.Entities,
			}, nil
		},
		Port: func(context.Context) (string, error) {
			s.e.merke("port")
			return "127.0.0.1:4711", nil
		},
		Starte: s.dienst.starte, Verbinde: s.dienst.verbinde, Transport: s.dienst,
		CPU: func() time.Duration {
			n++
			if n <= len(s.cpu) {
				return s.cpu[n-1]
			}
			return 0
		},
		Aus: &s.aus, Err: &s.fehler,
	}
	return s
}

// fahre fährt den Ablauf in einer eigenen Goroutine und liefert den Kanal seines Exit.
func (s *probestand) fahre(ctx context.Context) <-chan int {
	exit := make(chan int, 1)
	go func() { exit <- s.ein.fahre(ctx) }()
	return exit
}

// arten sind die Zeilenarten der Ausgabe in ihrer Folge.
func (s *probestand) arten() []string {
	var out []string
	for _, z := range s.zeilen() {
		out = append(out, strings.Fields(strings.TrimPrefix(z, "lasttest: "))[0])
	}
	return out
}

func (s *probestand) zeilen() []string {
	return strings.Split(strings.TrimSuffix(s.aus.String(), "\n"), "\n")
}

// zeilenVon sind die Zeilen einer Art.
func (s *probestand) zeilenVon(art string) []string {
	var out []string
	for _, z := range s.zeilen() {
		if strings.HasPrefix(z, "lasttest: "+art+" ") {
			out = append(out, z)
		}
	}
	return out
}

// feld ist der Wert des Schlüssels in einer Zeile.
func feld(zeile, schluessel string) string {
	for _, f := range strings.Fields(zeile) {
		if v, ok := strings.CutPrefix(f, schluessel+"="); ok {
			return v
		}
	}
	return ""
}

// selbstNachAllen beendet den Ersatz-Prozess mit code bei der Anfrage k, nachdem jede frühere beendet ist.
func (s *probestand) selbstNachAllen(code, k int) func(nr, i int) (int, bool) {
	return func(nr, i int) (int, bool) {
		if i != k {
			return 0, false
		}
		b := s.dienst.lauf(nr)
		bisWahr(func() bool { return b.fertig.Load() == int32(k) })
		return code, true
	}
}

// treibeBis wartet, bis bereit wahr ist, und stellt die Uhr dann in Schritten von d vor, bis fertig wahr ist (höchstens 5 s der wirklichen Zeit).
func treibeBis(t *testing.T, u *treibUhr, d time.Duration, bereit, fertig func() bool) {
	t.Helper()
	if !bisWahr(bereit) {
		t.Fatal("die Bedingung vor dem Vorstellen wird nicht wahr")
	}
	for i := 0; !fertig(); i++ {
		if i > 2500 {
			t.Fatal("die Bedingung wird nicht wahr")
		}
		pause(2 * time.Millisecond)
		if fertig() {
			return
		}
		u.Vorstellen(d)
	}
}

// hilfeVar wählt das Verhalten des Test-Binary als Hilfsprozess.
const hilfeVar = "LASTTEST_HILFE"

// hilfeBefehl ist der Aufruf des Test-Binary als Hilfsprozess mit der Umgebung env und dem Verhalten modus; vorweg ist ein Kommando von
// `sh`, das vorher in demselben Prozess läuft (leer: keines).
func hilfeBefehl(env []string, modus, vorweg string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestHilfsprozess$")
	if vorweg != "" {
		cmd = exec.Command("sh", "-c", vorweg+`; exec "$0" -test.run='^TestHilfsprozess$'`, os.Args[0])
	}
	cmd.Env = append([]string{hilfeVar + "=" + modus}, env...)
	return cmd
}

// TestHilfsprozess ist das Ersatz-Binary als eigener Prozess: es lauscht auf LISTEN, schreibt je Anfrage die Zeilen des Spans nach stderr
// und antwortet mit 200. Verhalten: `term0` und `term1` enden auf SIGTERM mit diesem Exit, `stur` ignoriert SIGTERM, `exit1` endet sofort
// mit 1, `stirbt` endet bei der fünften Anfrage mit 3. Ohne die Variable tut der Test nichts.
func TestHilfsprozess(t *testing.T) {
	modus := os.Getenv(hilfeVar)
	if modus == "" {
		return
	}
	if modus == "exit1" {
		os.Exit(1)
	}
	signale := make(chan os.Signal, 2)
	signal.Notify(signale, syscall.SIGTERM)
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "tcp", os.Getenv("LISTEN"))
	if err != nil {
		os.Exit(9)
	}
	aus := log.New(log.Writer(), "", 0)
	var n atomic.Int32
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := int(n.Add(1))
		if modus == "stirbt" && k == 5 {
			os.Exit(3)
		}
		id := "h-" + strconv.Itoa(k)
		aus.Print(kindZeile(os.Getenv("MODE"), id, 100))
		aus.Print(gesamtZeile(os.Getenv("MODE"), id, 1000))
		_, _ = io.WriteString(w, "{}")
	})}
	go func() { _ = srv.Serve(l) }()
	for range signale {
		switch modus {
		case "term0":
			os.Exit(0)
		case "term1":
			os.Exit(1)
		}
	}
}
