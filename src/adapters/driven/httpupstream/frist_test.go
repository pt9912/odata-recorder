package httpupstream_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driven/httpupstream"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// freigabe liefert einen Kanal, den frei schließt, auch mehrfach gerufen. Wer einen Server auf ihn
// warten lässt, registriert frei nach dem Start des Servers mit t.Cleanup: dann schließt der Kanal,
// bevor die Aufräumfunktion des Servers auf dessen Goroutinen wartet.
func freigabe() (<-chan struct{}, func()) {
	ch := make(chan struct{})
	var einmal sync.Once
	return ch, func() { einmal.Do(func() { close(ch) }) }
}

// haltServer liest einen Request, meldet ihn auf da und sendet keinen Kopf.
func haltServer(t *testing.T, da chan<- struct{}) *roherServer {
	t.Helper()
	return starteServer(t, func(_ net.Conn, r *bufio.Reader) {
		if _, err := leseAnfrage(r); err != nil {
			return
		}
		if da != nil {
			da <- struct{}{}
		}
		_, _ = io.Copy(io.Discard, r)
	})
}

// pipeServer liest einen Request in der Blase, wartet warten und schreibt antwort.
func pipeServer(warten time.Duration, antwort string) func(net.Conn) {
	return func(c net.Conn) {
		r := bufio.NewReader(c)
		if _, err := http.ReadRequest(r); err != nil {
			return
		}
		time.Sleep(warten)
		_, _ = io.WriteString(c, antwort)
		_, _ = io.Copy(io.Discard, r)
	}
}

// --- A8: Frist -----------------------------------------------------------------------------------

func fristFehlerFaelle() []fehlerFall {
	bg := context.Background()
	zeit := outbound.ErrUpstreamZeit
	return []fehlerFall{
		{name: "a Frist 20 ms ohne Kopf", art: zeit, lauf: func(t *testing.T) error {
			s := haltServer(t, nil)
			_, err := sende(bg, t, neuerClient(t, 20*time.Millisecond), mitGeheim(anfrage(s.ziel(), "GET", "")))
			return err
		}},
		{name: "c Auslösen vor dem Kopf", art: zeit, lauf: func(t *testing.T) error {
			da := make(chan struct{}, 1)
			s := haltServer(t, da)
			c := neuerClient(t, time.Hour)
			uhr := neueTestUhr(c)
			ch := sendeNebenbei(bg, c, mitGeheim(anfrage(s.ziel(), "GET", "")))
			warte(t, da)
			uhr.ausloesen()
			return warte(t, ch).err
		}},
		{name: "d Kopf, Beenden meldet abgelaufen", art: zeit, lauf: func(t *testing.T) error {
			c := neuerClient(t, time.Hour)
			uhr := neueTestUhr(c)
			body := &merkBody{Reader: strings.NewReader(geheim + "-a")}
			httpupstream.SetzeTransport(c, fakeTransport(func(*http.Request) (*http.Response, error) {
				uhr.ausloesen()
				return fakeKopf(body), nil
			}))
			_, err := c.Senden(bg, mitGeheim(anfrage(zielBlase, "GET", "")))
			if !body.zu.Load() {
				t.Error("Body des Fakes offen")
			}
			return err
		}},
		{name: "e Kopf 1 ns nach dem Ablauf", art: zeit, lauf: func(t *testing.T) error {
			return kopfNach(t, time.Second+time.Nanosecond)
		}},
		// Läuft die Frist im selben Augenblick ab, in dem der Server den Kopf sendet, hat die Uhr der
		// Blase den Timer gefeuert, bevor eine Goroutine weiterläuft: das Beenden der Frist misslingt.
		{name: "e Kopf beim Ablauf", art: zeit, lauf: func(t *testing.T) error {
			return kopfNach(t, time.Second)
		}},
	}
}

// kopfNach sendet in der Blase mit Frist 1 s; der Server sendet den Kopf nach warten.
func kopfNach(t *testing.T, warten time.Duration) error {
	t.Helper()
	var err error
	blase(t, time.Second, pipeServer(warten, antwortOK), func(_ *testing.T, c *httpupstream.Client) {
		ant, e := c.Senden(context.Background(), mitGeheim(anfrage(zielBlase, "GET", "")))
		err = fehlerVon(ergebnis{ant, e})
	})
	return err
}

// antwortVorAblauf sendet in der Blase mit Frist frist; der Server sendet den Kopf nach warten, vor
// dem Ablauf: Senden liefert die Antwort mit ihrem Body.
func antwortVorAblauf(t *testing.T, frist, warten time.Duration) {
	t.Helper()
	blase(t, frist, pipeServer(warten, antwortOK), func(t *testing.T, c *httpupstream.Client) {
		ant, err := c.Senden(context.Background(), anfrage(zielBlase, "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		if b, err := io.ReadAll(ant.Body); err != nil || string(b) != "ok" {
			t.Errorf("Body %q, %v", b, err)
		}
		_ = ant.Body.Close()
	})
}

// fehlerVon liefert den Fehler eines Ergebnisses; kam stattdessen eine Antwort, schließt es ihren
// Body, damit die Blase ohne wartende Goroutine endet.
func fehlerVon(e ergebnis) error {
	if e.err == nil {
		_ = e.ant.Body.Close()
	}
	return e.err
}

// TestFrist: die Frist begrenzt die Zeit bis zum Kopf, nicht den Body; treffen Kopf und Ablauf
// zusammen, entscheidet das Beenden der Frist (A8, SPEC-120 Nr. 6).
func TestFrist(t *testing.T) {
	fahre(t, fristFehlerFaelle())
	t.Run("b Auslösen nach dem Kopf", func(t *testing.T) {
		weiter, frei := freigabe()
		s := starteServer(t, func(c net.Conn, r *bufio.Reader) {
			if _, err := leseAnfrage(r); err != nil {
				return
			}
			_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\n")
			<-weiter
			_, _ = io.WriteString(c, "body")
		})
		t.Cleanup(frei)
		c := neuerClient(t, time.Hour)
		uhr := neueTestUhr(c)
		ant, err := sende(context.Background(), t, c, anfrage(s.ziel(), "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		uhr.ausloesen()
		frei()
		if b, err := liesAlles(t, ant.Body); err != nil || string(b) != "body" {
			t.Errorf("Body %q, %v", b, err)
		}
	})
	t.Run("e Kopf 1 ns vor dem Ablauf", func(t *testing.T) {
		antwortVorAblauf(t, time.Second, time.Second-time.Nanosecond)
	})
	// Die Frist hat keine Obergrenze (SPEC-120 Nr. 1): jeder Wert bis math.MaxInt64 ns ist gültig
	// (`UPSTREAM_TIMEOUT`, SPEC-110 Nr. 4). Die Frist math.MaxInt64 ns läuft nicht vor dem spätesten
	// Kopf ab, den die Blase trägt: ihre Uhr beginnt am 2000-01-01 und endet bei math.MaxInt64 ns ab
	// 1970 (rund 262,4 Jahre später); ein Kopf zu diesem Zeitpunkt fällt mit dem Ablauf zusammen,
	// gesendet wird der Kopf eine Nanosekunde davor. Eine Klemme der Frist auf höchstens diese Dauer
	// macht den Fall rot; eine größere ist in der Blase nicht von der unbegrenzten Frist zu unterscheiden.
	t.Run("f Frist ohne Obergrenze", func(t *testing.T) {
		beginn := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		antwortVorAblauf(t, time.Duration(math.MaxInt64), time.Duration(math.MaxInt64-beginn.UnixNano()-1))
	})
	t.Run("e Body 2 s nach dem Kopf", func(t *testing.T) {
		blase(t, time.Second, func(c net.Conn) {
			r := bufio.NewReader(c)
			if _, err := http.ReadRequest(r); err != nil {
				return
			}
			_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\n")
			time.Sleep(2 * time.Second)
			_, _ = io.WriteString(c, "abc")
			_, _ = io.Copy(io.Discard, r)
		}, func(t *testing.T, c *httpupstream.Client) {
			ant, err := c.Senden(context.Background(), anfrage(zielBlase, "GET", "/"))
			if err != nil {
				t.Fatal(err)
			}
			if b, err := io.ReadAll(ant.Body); err != nil || string(b) != "abc" {
				t.Errorf("Body %q, %v", b, err)
			}
			_ = ant.Body.Close()
		})
	})
}

// --- A12: Kontext --------------------------------------------------------------------------------

func kontextFehlerFaelle() []fehlerFall {
	return []fehlerFall{
		{name: "schon beendet, gültige Argumente", art: context.Canceled, lauf: func(t *testing.T) error {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			s := haltServer(t, nil)
			a := mitGeheim(anfrage(s.ziel(), "POST", ""))
			a.Body, a.Laenge = strings.NewReader(geheim+"-b"), int64(len(geheim+"-b"))
			_, err := sende(ctx, t, neuerClient(t, time.Hour), a)
			return err
		}},
		{name: "Abbruch vor dem Kopf", art: context.Canceled, lauf: func(t *testing.T) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			da := make(chan struct{}, 1)
			s := haltServer(t, da)
			ch := sendeNebenbei(ctx, neuerClient(t, time.Hour), mitGeheim(anfrage(s.ziel(), "GET", "")))
			warte(t, da)
			cancel()
			return warte(t, ch).err
		}},
		{name: "Deadline des Aufrufers", art: context.DeadlineExceeded, lauf: func(t *testing.T) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			s := haltServer(t, nil)
			_, err := sende(ctx, t, neuerClient(t, time.Hour), mitGeheim(anfrage(s.ziel(), "GET", "")))
			return err
		}},
	}
}

// TestKontext: ist ctx beendet, sein Fehler; im Body endet ein laufendes und ein späteres Read mit
// einem Fehler, der nicht io.EOF ist (A12, SPEC-120 Nr. 5, 8).
func TestKontext(t *testing.T) {
	fahre(t, kontextFehlerFaelle())
	t.Run("späteres Read", func(t *testing.T) {
		s := starteServer(t, func(c net.Conn, r *bufio.Reader) {
			if _, err := leseAnfrage(r); err != nil {
				return
			}
			_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nab")
			_, _ = io.Copy(io.Discard, r)
		})
		ctx, cancel := context.WithCancel(context.Background())
		ant, err := sende(ctx, t, neuerClient(t, time.Hour), anfrage(s.ziel(), "GET", "/"))
		if err != nil {
			t.Fatal(err)
		}
		cancel()
		if _, err := ant.Body.Read(make([]byte, 8)); err == nil || errors.Is(err, io.EOF) {
			t.Errorf("späteres Read: %v", err)
		}
		_ = ant.Body.Close()
	})
	t.Run("laufendes Read", func(t *testing.T) {
		blase(t, time.Hour, pipeServer(0, "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\n"), func(t *testing.T, c *httpupstream.Client) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ant, err := c.Senden(ctx, anfrage(zielBlase, "GET", "/"))
			if err != nil {
				t.Fatal(err)
			}
			gelesen := make(chan error, 1)
			go func() {
				_, err := ant.Body.Read(make([]byte, 8))
				gelesen <- err
			}()
			synctest.Wait()
			cancel()
			if err := <-gelesen; err == nil || errors.Is(err, io.EOF) {
				t.Errorf("laufendes Read: %v", err)
			}
			_ = ant.Body.Close()
		})
	})
}

// --- A13: Fehler des Request-Bodys ---------------------------------------------------------------

func bodyMitFehler() *stueckBody {
	return &stueckBody{stuecke: []stueckWert{{daten: geheim + "-ab"}}, ende: errBody}
}

func bodyFehlerFaelle() []fehlerFall {
	return []fehlerFall{
		{name: "a Body liefert einen Fehler", art: errBody, lauf: func(t *testing.T) error {
			s := haltServer(t, nil)
			a := mitGeheim(anfrage(s.ziel(), "POST", ""))
			a.Body, a.Laenge = bodyMitFehler(), -1
			_, err := sende(context.Background(), t, neuerClient(t, time.Hour), a)
			return err
		}},
		{name: "b Body-Fehler vor der Frist", art: errBody, lauf: func(t *testing.T) error {
			c := neuerClient(t, time.Hour)
			uhr := neueTestUhr(c)
			httpupstream.SetzeTransport(c, fakeTransport(func(req *http.Request) (*http.Response, error) {
				_, _ = io.ReadAll(req.Body)
				uhr.ausloesen()
				return nil, errVerbindung
			}))
			a := mitGeheim(anfrage(zielBlase, "POST", ""))
			a.Body, a.Laenge = bodyMitFehler(), -1
			_, err := c.Senden(context.Background(), a)
			return err
		}},
	}
}

// TestBodyFehler: ein Fehler von Body ist weder ErrUpstreamNichtErreichbar noch ErrUpstreamZeit,
// errors.Is erreicht ihn, er geht der Frist vor (A13, SPEC-120 Nr. 4, 8).
func TestBodyFehler(t *testing.T) {
	fahre(t, bodyFehlerFaelle())
}

// --- A16: Rangfolge ------------------------------------------------------------------------------

// rangFall fährt eine Folge des Fakes: folge bekommt den Request, cancel von ctx, den Timer der
// Frist und den Body, den ein Kopf trägt. Ist kopf gesetzt, muss der Body danach geschlossen sein.
func rangFall(name string, art error, kopf bool,
	folge func(req *http.Request, cancel func(), uhr *testUhr, body *merkBody) (*http.Response, error),
) fehlerFall {
	return fehlerFall{name: name, art: art, lauf: func(t *testing.T) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c := neuerClient(t, time.Hour)
		uhr := neueTestUhr(c)
		body := &merkBody{Reader: strings.NewReader(geheim + "-a")}
		httpupstream.SetzeTransport(c, fakeTransport(func(req *http.Request) (*http.Response, error) {
			return folge(req, cancel, uhr, body)
		}))
		a := mitGeheim(anfrage(zielBlase, "POST", ""))
		a.Body, a.Laenge = bodyMitFehler(), -1
		_, err := c.Senden(ctx, a)
		if kopf && !body.zu.Load() {
			t.Error("Body des Fakes offen")
		}
		return err
	}}
}

func rangfolgeFaelle() []fehlerFall {
	return []fehlerFall{
		rangFall("a ctx, Verbindung", context.Canceled, false, func(_ *http.Request, cancel func(), _ *testUhr, _ *merkBody) (*http.Response, error) {
			cancel()
			return nil, errVerbindung
		}),
		rangFall("b Body, ctx", context.Canceled, false, func(req *http.Request, cancel func(), _ *testUhr, _ *merkBody) (*http.Response, error) {
			_, err := io.ReadAll(req.Body)
			cancel()
			return nil, err
		}),
		rangFall("c ctx, Kopf", context.Canceled, true, func(_ *http.Request, cancel func(), _ *testUhr, b *merkBody) (*http.Response, error) {
			cancel()
			return fakeKopf(b), nil
		}),
		rangFall("d Frist, ctx, Kopf", context.Canceled, true, func(_ *http.Request, cancel func(), uhr *testUhr, b *merkBody) (*http.Response, error) {
			uhr.ausloesen()
			cancel()
			return fakeKopf(b), nil
		}),
		rangFall("e Frist, ctx, Verbindung", context.Canceled, false, func(_ *http.Request, cancel func(), uhr *testUhr, _ *merkBody) (*http.Response, error) {
			uhr.ausloesen()
			cancel()
			return nil, errVerbindung
		}),
		rangFall("f Body, Verbindung", errBody, false, func(req *http.Request, _ func(), _ *testUhr, _ *merkBody) (*http.Response, error) {
			_, _ = io.ReadAll(req.Body)
			return nil, errVerbindung
		}),
		rangFall("g Frist, Verbindung", outbound.ErrUpstreamZeit, false, func(_ *http.Request, _ func(), uhr *testUhr, _ *merkBody) (*http.Response, error) {
			uhr.ausloesen()
			return nil, errVerbindung
		}),
		rangFall("h ctx, Kopf 101", context.Canceled, true, func(_ *http.Request, cancel func(), _ *testUhr, b *merkBody) (*http.Response, error) {
			cancel()
			return fakeKopfMit(b, http.StatusSwitchingProtocols), nil
		}),
		rangFall("i Frist, Kopf 101", outbound.ErrUpstreamZeit, true, func(_ *http.Request, _ func(), uhr *testUhr, b *merkBody) (*http.Response, error) {
			uhr.ausloesen()
			return fakeKopfMit(b, http.StatusSwitchingProtocols), nil
		}),
	}
}

// TestRangfolge: die Paare gleichzeitiger Ursachen in der Rangfolge aus SPEC-120 Nr. 8; der Fake
// bestimmt die Folge. Ein Fehler von Body zählt nur vor dem Kopf: mit dem Kopf kommt die Antwort
// (A16, SPEC-120 Nr. 4, 5).
func TestRangfolge(t *testing.T) {
	fahre(t, rangfolgeFaelle())
	t.Run("j Body, Kopf 401", func(t *testing.T) {
		c := neuerClient(t, time.Hour)
		body := &merkBody{Reader: strings.NewReader("abgelehnt")}
		httpupstream.SetzeTransport(c, fakeTransport(func(req *http.Request) (*http.Response, error) {
			if _, err := io.ReadAll(req.Body); !errors.Is(err, errBody) {
				t.Errorf("Lesen des Request-Bodys: %v, erwartet errBody", err)
			}
			return fakeKopfMit(body, http.StatusUnauthorized), nil
		}))
		a := anfrage(zielBlase, "POST", "/")
		a.Body, a.Laenge = bodyMitFehler(), -1
		ant, err := c.Senden(context.Background(), a)
		if err != nil {
			t.Fatalf("Fehler %v, erwartet die Antwort 401", err)
		}
		b, err := io.ReadAll(ant.Body)
		_ = ant.Body.Close()
		if ant.Status != http.StatusUnauthorized || err != nil || string(b) != "abgelehnt" {
			t.Errorf("Status %d, Body %q, %v", ant.Status, b, err)
		}
	})
}

// --- A18: blockiertes Read des Request-Bodys -----------------------------------------------------

// blockBody liefert `ab`, dann wartet sein zweites Read auf frei und liefert `cd` mit zweitEnde;
// jedes weitere Read liefert io.EOF.
func blockBody(frei <-chan struct{}, zweitEnde error) *stueckBody {
	return &stueckBody{stuecke: []stueckWert{{daten: geheim + "ab"}, {daten: "cd", err: zweitEnde, freigabe: frei}}, ende: io.EOF}
}

// blockServer liest in der Blase Kopf und das erste Stück; dann schließt er die Verbindung oder
// schreibt antwort und liest weiter.
func blockServer(schliessen bool, antwort string) func(net.Conn) {
	return func(c net.Conn) {
		r := bufio.NewReader(c)
		req, err := http.ReadRequest(r)
		if err != nil {
			return
		}
		if _, err := io.ReadFull(req.Body, make([]byte, len(geheim+"ab"))); err != nil || schliessen {
			return
		}
		_, _ = io.WriteString(c, antwort)
		_, _ = io.Copy(io.Discard, r)
	}
}

// blockFall fährt einen Fall in der Blase: Senden läuft, bis das zweite Read wartet (synctest.Wait),
// dann handelt aktion (Abbruch, Auslösen der Frist, Freigabe des Reads oder nichts). Nach der
// Rückkehr gibt der Fall das Read frei; ein drittes Read darf es nicht geben. Ohne mitUhr läuft die
// Frist von 1 s mit dem echten Timer auf der Uhr der Blase.
func blockFall(name string, art error, schliessen, mitUhr bool, zweitEnde error,
	aktion func(cancel func(), uhr *testUhr, frei func()), dauer func(d time.Duration) bool,
) fehlerFall {
	return fehlerFall{name: name, art: art, lauf: func(t *testing.T) error {
		var err error
		blase(t, time.Second, blockServer(schliessen, ""), func(t *testing.T, c *httpupstream.Client) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var uhr *testUhr
			if mitUhr {
				uhr = neueTestUhr(c)
			}
			frei, gibFrei := freigabe()
			defer gibFrei()
			body := blockBody(frei, zweitEnde)
			a := mitGeheim(anfrage(zielBlase, "POST", ""))
			a.Body, a.Laenge = body, -1
			start := time.Now()
			ch := sendeNebenbei(ctx, c, a)
			synctest.Wait()
			if n := body.aufrufe.Load(); n != 2 {
				t.Fatalf("%d Aufrufe von Read vor dem Haltepunkt", n)
			}
			aktion(cancel, uhr, gibFrei)
			err = fehlerVon(warte(t, ch))
			if d := time.Since(start); !dauer(d) {
				t.Errorf("Senden kehrte nach %v zurück", d)
			}
			gibFrei()
			synctest.Wait()
			if n := body.aufrufe.Load(); n != 2 {
				t.Errorf("%d Aufrufe von Read", n)
			}
		})
		return err
	}}
}

func sofort(d time.Duration) bool { return d == 0 }

// nachFehlerFall: der Fake liest das erste Stück des Request-Bodys und endet mit folge; danach
// liest er noch einmal aus dem Request. Nach einem Fehler von Senden ruft der Client kein neues
// Read von Body (A18, SPEC-120 Nr. 4).
func nachFehlerFall(name string, folge func() (*http.Response, error)) fehlerFall {
	return fehlerFall{name: name, art: outbound.ErrUpstreamNichtErreichbar, lauf: func(t *testing.T) error {
		c := neuerClient(t, time.Hour)
		var gemerkt *http.Request
		httpupstream.SetzeTransport(c, fakeTransport(func(req *http.Request) (*http.Response, error) {
			gemerkt = req
			if _, err := req.Body.Read(make([]byte, 64)); err != nil {
				t.Errorf("erstes Stück: %v", err)
			}
			return folge()
		}))
		body := &stueckBody{stuecke: []stueckWert{{daten: geheim + "ab"}, {daten: "cd"}}, ende: io.EOF}
		a := mitGeheim(anfrage(zielBlase, "POST", ""))
		a.Body, a.Laenge = body, -1
		_, err := c.Senden(context.Background(), a)
		if n, rerr := gemerkt.Body.Read(make([]byte, 64)); n != 0 || rerr == nil {
			t.Errorf("Read nach dem Fehler: %d, %v", n, rerr)
		}
		if n := body.aufrufe.Load(); n != 1 {
			t.Errorf("%d Aufrufe von Read", n)
		}
		return err
	}}
}

// nachFrist verlangt die Rückkehr genau mit dem Ablauf der Frist von 1 s.
func nachFrist(d time.Duration) bool { return d == time.Second }

func blockiertFehlerFaelle() []fehlerFall {
	return []fehlerFall{
		blockFall("a Abbruch von ctx", context.Canceled, false, false, nil, func(cancel func(), _ *testUhr, _ func()) { cancel() }, sofort),
		blockFall("b Ablauf der Frist", outbound.ErrUpstreamZeit, false, true, nil, func(_ func(), uhr *testUhr, _ func()) { uhr.ausloesen() }, sofort),
		blockFall("d Verbindung endet, Frist", outbound.ErrUpstreamZeit, true, false, nil, func(func(), *testUhr, func()) {}, nachFrist),
		blockFall("e Verbindung endet, Read kehrt zurück", outbound.ErrUpstreamNichtErreichbar, true, false, io.EOF,
			func(_ func(), _ *testUhr, frei func()) { frei() }, sofort),
		nachFehlerFall("f Fehler der Verbindung", func() (*http.Response, error) { return nil, errVerbindung }),
		nachFehlerFall("g Antwort 101", func() (*http.Response, error) {
			return fakeKopfMit(&merkBody{Reader: strings.NewReader("")}, http.StatusSwitchingProtocols), nil
		}),
		verworfenFall(),
	}
}

// verworfenFall: ein Read von Body, das beim Ende des Austauschs läuft, endet danach mit errBody;
// sein Ergebnis ist verworfen, kein Fehler von Body, Senden endet mit ErrUpstreamZeit. Der Fake in
// der Blase liest das erste Stück, startet das zweite Read, löst die Frist aus, gibt das Read frei
// und wartet mit synctest.Wait, bis der Leser es verarbeitet hat (A18, SPEC-120 Nr. 4, 8).
func verworfenFall() fehlerFall {
	return fehlerFall{name: "h Fehler von Body nach dem Ende", art: outbound.ErrUpstreamZeit, lauf: func(t *testing.T) error {
		var err error
		synctest.Test(t, func(t *testing.T) {
			c, nerr := httpupstream.New(httpupstream.Konfig{Frist: time.Hour, Pruefen: true})
			if nerr != nil {
				t.Fatal(nerr)
			}
			uhr := neueTestUhr(c)
			frei := make(chan struct{})
			body := &stueckBody{stuecke: []stueckWert{{daten: geheim + "ab"}, {err: errBody, freigabe: frei}}, ende: io.EOF}
			httpupstream.SetzeTransport(c, fakeTransport(func(req *http.Request) (*http.Response, error) {
				if _, rerr := req.Body.Read(make([]byte, 64)); rerr != nil {
					t.Errorf("erstes Stück: %v", rerr)
				}
				zweites := make(chan error, 1)
				go func() {
					_, rerr := req.Body.Read(make([]byte, 64))
					zweites <- rerr
				}()
				synctest.Wait()
				uhr.ausloesen()
				if rerr := <-zweites; rerr == nil {
					t.Error("zweites Read nach dem Ende ohne Fehler")
				}
				close(frei)
				synctest.Wait()
				return nil, errVerbindung
			}))
			a := mitGeheim(anfrage(zielBlase, "POST", ""))
			a.Body, a.Laenge = body, -1
			_, err = c.Senden(context.Background(), a)
		})
		return err
	}}
}

// TestBodyBlockiert: das Lesen aus Body endet mit dem Austausch; Senden wartet nicht auf ein
// laufendes Read; dieses Ende ist kein Fehler von Body (A18, SPEC-120 Nr. 4, 6, 8).
func TestBodyBlockiert(t *testing.T) {
	fahre(t, blockiertFehlerFaelle())
	t.Run("c Schließen des Antwort-Bodys", func(t *testing.T) {
		serverEnde := make(chan struct{})
		server := blockServer(false, antwortOK)
		blase(t, time.Hour, func(c net.Conn) {
			defer close(serverEnde)
			server(c)
		}, func(t *testing.T, c *httpupstream.Client) {
			frei, gibFrei := freigabe()
			defer gibFrei()
			body := blockBody(frei, nil)
			a := anfrage(zielBlase, "POST", "/")
			a.Body, a.Laenge = body, -1
			ant, err := c.Senden(context.Background(), a)
			if err != nil {
				t.Fatal(err)
			}
			synctest.Wait()
			if n := body.aufrufe.Load(); n != 2 {
				t.Fatalf("%d Aufrufe von Read vor dem Haltepunkt", n)
			}
			_ = ant.Body.Close()
			synctest.Wait()
			select {
			case <-serverEnde:
			default:
				t.Error("nach dem Schließen wartet die Verbindung auf das laufende Read")
			}
			gibFrei()
			synctest.Wait()
			if n := body.aufrufe.Load(); n != 2 {
				t.Errorf("%d Aufrufe von Read", n)
			}
		})
	})
}
