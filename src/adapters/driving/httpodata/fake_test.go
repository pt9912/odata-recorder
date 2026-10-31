package httpodata_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/httpodata"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support/supporttest"
)

var (
	_ inbound.RecordUseCase = (*dienst)(nil)
	_ inbound.ReplayUseCase = (*dienst)(nil)
)

// aufruf hält, was der Dienst bei einem Aufruf sah.
type aufruf struct {
	a      inbound.Anfrage
	abbild map[string][]string
}

// dienst ist der Fake von RecordUseCase und ReplayUseCase in einem Typ: er hält die Anfrage und das
// Abbild ihrer Header fest und zählt jede Operation, auch Drain. Den ctx wertet nur A12 aus, in der
// Funktion des Tests; containedctx verbietet ihn als Feld.
// fest und zählt jede Operation, auch Drain.
type dienst struct {
	mu      sync.Mutex
	aufrufe []aufruf
	record  int
	replay  int
	drain   int
	tu      func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error)
}

func (d *dienst) los(ctx context.Context, a inbound.Anfrage, zaehler *int) (inbound.Antwort, error) {
	d.mu.Lock()
	*zaehler++
	d.aufrufe = append(d.aufrufe, aufruf{a: a, abbild: kopiereMap(a.Header)})
	tu := d.tu
	d.mu.Unlock()
	if tu == nil {
		return inbound.Antwort{Status: 200, Body: leer()}, nil
	}
	return tu(ctx, a)
}

func (d *dienst) Record(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	return d.los(ctx, a, &d.record)
}

func (d *dienst) Replay(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	return d.los(ctx, a, &d.replay)
}

func (d *dienst) Drain(context.Context, string, time.Duration) (inbound.DrainBericht, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.drain++
	return inbound.DrainBericht{}, nil
}

func (d *dienst) zaehler() (record, replay, drain, gesamt int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.record, d.replay, d.drain, len(d.aufrufe)
}

func (d *dienst) alle() []aufruf {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]aufruf(nil), d.aufrufe...)
}

func kopiereMap(m map[string][]string) map[string][]string {
	if m == nil {
		return nil
	}
	out := make(map[string][]string, len(m))
	for k, v := range m {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// bindung wählt die Methode des Fakes, die der Composition Root als Verarbeiten bindet.
type bindung string

func bindungen() []bindung { return []bindung{"record", "replay"} }

func (d *dienst) binde(b bindung) httpodata.Verarbeiter {
	if b == "record" {
		return d.Record
	}
	return d.Replay
}

// zaehler der Bindung: der eigene Dienst wird gerufen, der andere nie.
func (d *dienst) eigen(b bindung) (eigen, fremd int) {
	r, p, _, _ := d.zaehler()
	if b == "record" {
		return r, p
	}
	return p, r
}

// koerper ist der Body-Fake: er zählt Close, Read nach Close und Read nach dem Ende.
type koerper struct {
	mu        sync.Mutex
	lies      func(p []byte) (int, error)
	closes    int
	nachClose int
	nachEnde  int
	reads     int
	ende      bool
	beiClose  func()
	closeErr  error
}

func (k *koerper) Read(p []byte) (int, error) {
	k.mu.Lock()
	k.reads++
	if k.closes > 0 {
		k.nachClose++
	}
	if k.ende {
		k.nachEnde++
	}
	k.mu.Unlock()
	n, err := k.lies(p)
	if err != nil {
		k.mu.Lock()
		k.ende = true
		k.mu.Unlock()
	}
	return n, err
}

func (k *koerper) Close() error {
	k.mu.Lock()
	k.closes++
	f := k.beiClose
	k.mu.Unlock()
	if f != nil {
		f()
	}
	return k.closeErr
}

func (k *koerper) stand() (closes, nachClose, nachEnde, reads int) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.closes, k.nachClose, k.nachEnde, k.reads
}

func ausBytes(b []byte) *koerper {
	r := bytes.NewReader(b)
	return &koerper{lies: r.Read}
}

func leer() *koerper { return ausBytes(nil) }

// muster liefert n Zufallsbytes, deterministisch je Startwert.
func muster(n int, start uint64) []byte {
	r := rand.New(rand.NewPCG(start, 7))
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.IntN(256))
	}
	return b
}

// zaehlIDs zählt jede Operation von FestIDs.
type zaehlIDs struct {
	*supporttest.FestIDs
	n [5]atomic.Int64
}

func (z *zaehlIDs) SessionID() string      { z.n[0].Add(1); return z.FestIDs.SessionID() }
func (z *zaehlIDs) ExchangeID() string     { z.n[1].Add(1); return z.FestIDs.ExchangeID() }
func (z *zaehlIDs) KorrelationsID() string { z.n[2].Add(1); return z.FestIDs.KorrelationsID() }
func (z *zaehlIDs) WorkspaceID() string    { z.n[3].Add(1); return z.FestIDs.WorkspaceID() }
func (z *zaehlIDs) SnapshotID() string     { z.n[4].Add(1); return z.FestIDs.SnapshotID() }

func (z *zaehlIDs) zahlen() [5]int64 {
	var out [5]int64
	for i := range z.n {
		out[i] = z.n[i].Load()
	}
	return out
}

// umgebung ist ein laufender Server auf 127.0.0.1:0 mit Fake-Dienst und gezählten IDs.
type umgebung struct {
	addr string
	d    *dienst
	ids  *zaehlIDs
	lis  net.Listener
}

func neuUmgebung(t *testing.T, b bindung, tu func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error)) *umgebung {
	t.Helper()
	d := &dienst{tu: tu}
	ids := &zaehlIDs{FestIDs: supporttest.NewFestIDs()}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := httpodata.New(httpodata.Konfig{Verarbeiten: d.binde(b), IDs: ids, Listener: lis})
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve() }()
	t.Cleanup(func() { _ = lis.Close() })
	return &umgebung{addr: lis.Addr().String(), d: d, ids: ids, lis: lis}
}

func (u *umgebung) client() *http.Client {
	return &http.Client{
		Transport:     &http.Transport{DisableKeepAlives: true},
		Timeout:       10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// roh sendet die Bytes unverändert und liest bis zum Ende der Verbindung.
func (u *umgebung) roh(t *testing.T, bytesGesendet []byte) []byte {
	t.Helper()
	c, err := net.Dial("tcp", u.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.Write(bytesGesendet); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(c)
	if err != nil && len(out) == 0 {
		t.Logf("Lesen der Antwort: %v", err)
	}
	return out
}

// zeile baut eine Request-Zeile mit den Zeilen des Kopfes; Connection: close beendet die Antwort.
func zeile(methode, target string, kopf ...string) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s %s HTTP/1.1\r\nHost: h\r\nConnection: close\r\n", methode, target)
	for _, k := range kopf {
		b.WriteString(k + "\r\n")
	}
	b.WriteString("\r\n")
	return b.Bytes()
}

// teile trennt die rohe Antwort in Kopf (Zeilen) und Body.
func teile(roh []byte) (zeilen []string, body []byte) {
	kopf, rest, _ := bytes.Cut(roh, []byte("\r\n\r\n"))
	return strings.Split(string(kopf), "\r\n"), rest
}

func warte(t *testing.T, was string, ok func() bool) {
	t.Helper()
	ende := time.Now().Add(5 * time.Second)
	for !ok() {
		if time.Now().After(ende) {
			t.Fatalf("Frist von 5 s abgelaufen: %s", was)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func gleich(t *testing.T, was string, got, soll any) {
	t.Helper()
	if !reflect.DeepEqual(got, soll) {
		t.Errorf("%s = %#v, erwartet %#v", was, got, soll)
	}
}

func erster(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// rohOffen sendet eine Request-Zeile ohne Connection: close und liest, bis die Antwort mit dem
// Abschluss des chunked Bodys endet.
func (u *umgebung) rohOffen(t *testing.T, methode, target string) []byte {
	t.Helper()
	c, err := net.Dial("tcp", u.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := fmt.Fprintf(c, "%s %s HTTP/1.1\r\nHost: h\r\n\r\n", methode, target); err != nil {
		t.Fatal(err)
	}
	var out []byte
	puffer := make([]byte, 4096)
	for !bytes.HasSuffix(out, []byte("0\r\n\r\n")) {
		n, err := c.Read(puffer)
		out = append(out, puffer[:n]...)
		if err != nil {
			t.Fatalf("Antwort endete nicht mit dem chunked Abschluss: %v, gelesen %q", err, out)
		}
	}
	return out
}

// schranke gibt alle frei, sobald n Teilnehmer warten (oder nach 5 s).
type schranke struct {
	mu   sync.Mutex
	n    int
	da   int
	frei chan struct{}
}

func neueSchranke(n int) *schranke { return &schranke{n: n, frei: make(chan struct{})} }

func (s *schranke) warte() {
	s.mu.Lock()
	s.da++
	if s.da == s.n {
		close(s.frei)
	}
	s.mu.Unlock()
	select {
	case <-s.frei:
	case <-time.After(5 * time.Second):
	}
}
