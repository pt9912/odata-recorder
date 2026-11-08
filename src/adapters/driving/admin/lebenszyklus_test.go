package admin_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/admin"
	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestAnhaltenSchliesstListener hält S3 (SPEC-136 Nr. 13 a): Anhalten schließt den Listener und kehrt ohne Warten
// zurück; der laufende Request bleibt offen und endet mit seiner Antwort.
func TestAnhaltenSchliesstListener(t *testing.T) {
	route := neueHaengRoute(t, wartetFrei)
	u := neuLive(t, route)
	r := u.async("GET", "/p")
	route.warteBegonnen(t, 1)
	if !innerhalb(2*time.Second, u.srv.Anhalten) {
		t.Fatal("Anhalten kehrte in 2 s nicht zurück")
	}
	if !dialScheitert(u.addr) {
		t.Error("nach Anhalten gelingt ein Dial")
	}
	select {
	case x := <-r:
		t.Fatalf("der laufende Request endete mit Anhalten: %+v", x)
	default:
	}
	route.gibFrei()
	if x := erwarte(t, r, "Antwort des laufenden Requests"); x.err != nil || x.status != 200 || string(x.body) != "ok" {
		t.Errorf("laufender Request: %+v", x)
	}
}

// TestAnhaltenWiederholt hält S3 (SPEC-136 Nr. 13 d): ein weiteres und ein gleichzeitiges Anhalten sind wirkungslos;
// nach der Rückkehr jedes Aufrufs ist der Listener zu (20 Läufe mit 16 Aufrufern, Close wartet 30 ms: das Fenster W2
// ist breit); ohne Serve ergeben 16 gleichzeitige und ein weiterer Aufruf genau einen Close.
func TestAnhaltenWiederholt(t *testing.T) {
	const aufrufer = 16
	for lauf := range 20 {
		u := neuLiveMit(t, neueRoute(), liveOpt{umhuelle: func(l net.Listener) net.Listener {
			return &langsamListener{Listener: l, verz: 30 * time.Millisecond}
		}})
		s := neueSchranke(aufrufer)
		var wg sync.WaitGroup
		offen := make([]bool, aufrufer)
		for i := range aufrufer {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.warte()
				u.srv.Anhalten()
				offen[i] = !dialScheitert(u.addr)
			}()
		}
		wg.Wait()
		for i, o := range offen {
			if o {
				t.Errorf("Lauf %d, Aufrufer %d: nach der Rückkehr von Anhalten gelingt ein Dial", lauf, i)
			}
		}
	}

	k, env := konfigMit()
	env.lst.closeVerz = 10 * time.Millisecond
	srv := neuMitKonfig(t, k, env).srv
	var wg sync.WaitGroup
	for range aufrufer {
		wg.Add(1)
		go func() { defer wg.Done(); srv.Anhalten() }()
	}
	wg.Wait()
	srv.Anhalten()
	if accepts, closes, addrs := env.lst.zahlen(); accepts != 0 || closes != 1 || addrs != 0 {
		t.Errorf("Listener nach %d Aufrufen von Anhalten ohne Serve: %d Accept, %d Close, %d Addr; erwartet 0, 1 und 0", aufrufer+1, accepts, closes, addrs)
	}
}

// form ist eine Request-Form von S4, je Element eine eigene Verbindung.
type form struct {
	name, methode string
	kopf          string
	body          []byte
	weiter        bool // der Client sendet den Body in einer Goroutine weiter
	client        bool // der Client schließt selbst, der Body ist nicht fertig gesendet
}

// formen ist die Klasse „jede Form des Requests“ (SPEC-136 Nr. 13 b), 50 Elemente: 10 Methoden (darunter `get` und
// ein Token ohne Bedeutung) mal 4 Body-Formen (keiner, Länge 0, 1 KiB, chunked) in der Origin-Form mit gültigem Token,
// dazu 10 Elemente für die Formen, die der Server anders liest: Body größer als der Rest, den net/http verwirft (1 MiB,
// weitergesendet), angekündigter und nicht gesendeter Body, Expect, `OPTIONS *`, `CONNECT`, absolute-form (GET, POST),
// `GET /healthz`, ein unbekannter Pfad und ein Request ohne Token.
func formen() []form {
	m := muster(1<<20, 3)
	a := kopfAuth()
	var out []form
	methoden := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "TRACE", "PROPFIND", "get"}
	bodies := []struct {
		name, kopf string
		body       []byte
	}{
		{"ohne Body", "", nil},
		{"Content-Length 0", "Content-Length: 0\r\n", nil},
		{"Content-Length 1 KiB", "Content-Length: 1024\r\n", m[:1024]},
		{"chunked", "Transfer-Encoding: chunked\r\n", []byte("5\r\nhello\r\n0\r\n\r\n")},
	}
	for _, me := range methoden {
		for _, b := range bodies {
			out = append(out, form{name: me + " " + b.name, methode: me, kopf: me + " /p HTTP/1.1\r\nHost: h\r\n" + a + b.kopf + "\r\n", body: b.body})
		}
	}
	return append(out,
		form{name: "POST 1 MiB weitergesendet", methode: "POST", kopf: "POST /p HTTP/1.1\r\nHost: h\r\n" + a + "Content-Length: 1048576\r\n\r\n", body: m, weiter: true},
		form{name: "POST angekündigt 1 MiB, 10 gesendet", methode: "POST", kopf: "POST /p HTTP/1.1\r\nHost: h\r\n" + a + "Content-Length: 1048576\r\n\r\n", body: m[:10], client: true},
		form{name: "POST Expect", methode: "POST", kopf: "POST /p HTTP/1.1\r\nHost: h\r\n" + a + "Expect: 100-continue\r\nContent-Length: 1024\r\n\r\n", client: true},
		form{name: "OPTIONS *", methode: "OPTIONS", kopf: "OPTIONS * HTTP/1.1\r\nHost: h\r\n" + a + "\r\n"},
		form{name: "CONNECT", methode: "CONNECT", kopf: "CONNECT h:443 HTTP/1.1\r\nHost: h:443\r\n" + a + "\r\n"},
		form{name: "GET absolute-form", methode: "GET", kopf: "GET http://h/p?q=1 HTTP/1.1\r\nHost: h\r\n" + a + "\r\n"},
		form{name: "POST absolute-form", methode: "POST", kopf: "POST http://h/p HTTP/1.1\r\nHost: h\r\n" + a + "Content-Length: 0\r\n\r\n"},
		form{name: "GET /healthz", methode: "GET", kopf: "GET /healthz HTTP/1.1\r\nHost: h\r\n\r\n"},
		form{name: "GET unbekannter Pfad", methode: "GET", kopf: "GET /unbekannt HTTP/1.1\r\nHost: h\r\n" + a + "\r\n"},
		form{name: "GET ohne Token", methode: "GET", kopf: "GET /p HTTP/1.1\r\nHost: h\r\n\r\n"},
	)
}

// TestAngehaltenDreiNullDrei hält S4 (SPEC-136 Nr. 13 b, SPEC-238 Nr. 3 b, 11): nach Anhalten ergibt jeder Request auf
// einer bestehenden Verbindung 503 ODR-E5010 mit Connection: close, ohne die Route und ohne den Body zu lesen; die
// Verbindung wird nicht wiederverwendet.
func TestAngehaltenDreiNullDrei(t *testing.T) {
	route := neueRoute()
	u := neuLive(t, route)
	fs := formen()
	gleichAuf(t, "Zahl der Formen", len(fs), 50)
	vs := make([]*verbindung, len(fs))
	for i := range fs {
		vs[i] = u.verbinde(t)
		vs[i].erledige(t)
	}
	u.srv.Anhalten()
	for i, f := range fs {
		v := vs[i]
		v.schreibe(t, []byte(f.kopf))
		if f.weiter {
			go func() { _, _ = v.c.Write(f.body) }()
		} else {
			v.schreibe(t, f.body)
		}
		resp, body := v.antwort(t, f.methode)
		if resp.StatusCode != 503 || !resp.Close || resp.Header.Get("X-Correlation-Id") == "" {
			t.Errorf("%s: Status %d, Close %v, ID %q; erwartet 503, Close und eine ID", f.name, resp.StatusCode, resp.Close, resp.Header.Get("X-Correlation-Id"))
		}
		if resp.Header.Get("Content-Type") != "application/json" || resp.Header.Get("Content-Length") != fmt.Sprint(len(objekt503)) {
			t.Errorf("%s: Content-Type %q, Content-Length %q", f.name, resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"))
		}
		if f.methode == "HEAD" {
			if len(body) != 0 {
				t.Errorf("%s: Body %q, erwartet leer", f.name, body)
			}
		} else if string(body) != objekt503 {
			t.Errorf("%s: Body %q, erwartet %q", f.name, body, objekt503)
		}
		if n := route.aufrufe(); n != len(fs) {
			t.Errorf("%s: die Route wurde gerufen (%d Aufrufe, erwartet %d)", f.name, n, len(fs))
		}
		if k, _ := u.ids.zahlen(); k != len(fs)+i+1 {
			t.Errorf("%s: %d Aufrufe von KorrelationsID, erwartet %d", f.name, k, len(fs)+i+1)
		}
		if u.audit.Aufrufe() != 0 {
			t.Errorf("%s: %d Einträge im Audit, erwartet 0", f.name, u.audit.Aufrufe())
		}
	}
	for i, f := range fs {
		if f.client {
			continue
		}
		if n, err := vs[i].ende(500 * time.Millisecond); n != 0 || err == nil || istZeitgrenze(err) {
			t.Errorf("%s: nach der Antwort %d Bytes und Fehler %v; erwartet kein Byte und eine geschlossene Verbindung", f.name, n, err)
		}
	}
}

// TestLaufenderRequestEndet hält S5 (SPEC-136 Nr. 13 c): ein vor Anhalten zugelassener Request läuft zu Ende,
// vollständig; sein ctx endet nicht.
func TestLaufenderRequestEndet(t *testing.T) {
	body := muster(1<<20, 5)
	route := neueHaengRoute(t, wartetFrei)
	route.antwort.Body = body
	u := neuLive(t, route)
	r := u.async("GET", "/p")
	route.warteBegonnen(t, 1)
	u.srv.Anhalten()
	route.gibFrei()
	x := erwarte(t, r, "Antwort des laufenden Requests")
	if x.err != nil || x.status != 200 || !bytes.Equal(x.body, body) {
		t.Errorf("laufender Request: Status %d, %d Bytes, Fehler %v", x.status, len(x.body), x.err)
	}
	gleichAuf(t, "Aufrufe der Route", route.aufrufe(), 1)
	gleichAuf(t, "Fehler von ctx: Anhalten beendet ihn nicht", route.ctxFehlerListe(), []error{nil})
}

// TestZulassungAtomar hält S6 (SPEC-136 Nr. 13 c): Prüfen auf angehalten und Zählen liegen unter einer Sperre ohne
// Freigabe dazwischen. Fall 0 hält den Request vor der Zulassung (KorrelationsID), Fall 1 ist der Wirkungs-Fänger.
func TestZulassungAtomar(t *testing.T) {
	t.Run("Fall 0: Haltepunkt an KorrelationsID", func(t *testing.T) {
		erreicht, weiter := make(chan struct{}), make(chan struct{})
		var einmal sync.Once
		route := neueRoute()
		u := neuLiveMit(t, route, liveOpt{vorKorr: func() {
			einmal.Do(func() { close(erreicht); <-weiter })
		}})
		r := u.async("GET", "/p")
		select {
		case <-erreicht:
		case <-time.After(5 * time.Second):
			t.Fatal("der Request erreichte KorrelationsID nicht")
		}
		u.srv.Anhalten()
		if _, err := schliesse(t, u.srv, time.Second); err != nil {
			t.Errorf("Schliessen(1 s) = %v, erwartet nil", err)
		}
		close(weiter)
		if x := erwarte(t, r, "Antwort"); x.status != 503 {
			t.Errorf("Status %d, erwartet 503", x.status)
		}
		gleichAuf(t, "Aufrufe der Route nach Schliessen", route.aufrufe(), 0)
	})
	t.Run("Fall 1: Wirkung unter 64 gleichzeitigen Requests in 200 Läufen", func(t *testing.T) {
		const n = 64
		for lauf := range 200 {
			route := neueRoute()
			route.begonnen, route.frei = make(chan struct{}, 4096), make(chan struct{})
			route.endeVerz = 100 * time.Microsecond
			// Alle n Requests warten vor der Zulassung am Haltepunkt, das Tor öffnet sich unmittelbar vor Anhalten:
			// n Zulassungen laufen gleichzeitig mit dem Setzen des Flags.
			tor := make(chan struct{})
			var da atomic.Int64
			u := neuLiveMit(t, route, liveOpt{vorKorr: func() { da.Add(1); <-tor }})
			res := make(chan ergebnis, n)
			for range n {
				go func() { res <- u.schleife() }()
			}
			warteBis(t, "alle Requests am Haltepunkt", func() bool { return da.Load() == n })
			close(tor)
			u.srv.Anhalten()
			_, err := schliesse(t, u.srv, 2*time.Second)
			aufrufe, laufend := route.aufrufe(), route.laufend.Load()
			if err != nil || laufend != 0 {
				t.Fatalf("Lauf %d: Schliessen = %v; laufende Handler bei der Rückkehr %d; erwartet nil und 0", lauf, err, laufend)
			}
			for range n {
				if x := <-res; x.err == nil && x.status != 200 && x.status != 503 {
					t.Fatalf("Lauf %d: Status %d, erwartet 200, 503 oder einen Verbindungsfehler", lauf, x.status)
				}
			}
			if nach := route.aufrufe(); nach != aufrufe {
				t.Fatalf("Lauf %d: %d Aufrufe der Route bei der Rückkehr von Schliessen, %d danach; erwartet gleich", lauf, aufrufe, nach)
			}
		}
	})
}

// TestSchliessenNegativ hält S7 (SPEC-238 Nr. 11 b, SPEC-136 Nr. 14): eine negative Frist ergibt in jedem Zustand
// (frisch, angehalten, geschlossen) ErrUngueltig und ändert nichts; 0 ist gültig. 3 Zustände × 2 Fristen = 6 Elemente.
func TestSchliessenNegativ(t *testing.T) {
	zustaende := []struct {
		name string
		// pruefe läuft nach den negativen Fristen mit der Verbindung v, die vor dem Zustandswechsel im Leerlauf lag.
		richte func(t *testing.T, u *live, v *verbindung)
		pruefe func(t *testing.T, u *live, v *verbindung)
	}{
		{"frisch", func(*testing.T, *live, *verbindung) {}, func(t *testing.T, u *live, v *verbindung) {
			v.erledige(t)
			select {
			case err := <-u.serve:
				t.Errorf("Serve endete nach einer negativen Frist: %v", err)
			default:
			}
		}},
		{"angehalten", func(_ *testing.T, u *live, _ *verbindung) { u.srv.Anhalten() }, func(t *testing.T, u *live, v *verbindung) {
			v.schreibe(t, getP())
			if resp, _ := v.antwort(t, "GET"); resp.StatusCode != 503 {
				t.Errorf("Status %d, erwartet 503", resp.StatusCode)
			}
		}},
		{"geschlossen", func(t *testing.T, u *live, _ *verbindung) {
			if err := u.srv.Schliessen(0); err != nil {
				t.Fatalf("Schliessen(0) = %v", err)
			}
		}, func(t *testing.T, u *live, _ *verbindung) {
			if !dialScheitert(u.addr) {
				t.Error("nach Schliessen gelingt ein Dial")
			}
		}},
	}
	fristen := []time.Duration{-1, time.Duration(math.MinInt64)}
	n := 0
	for _, z := range zustaende {
		for _, f := range fristen {
			n++
			u := neuLive(t, neueRoute())
			v := u.verbinde(t)
			v.erledige(t)
			z.richte(t, u, v)
			if err := u.srv.Schliessen(f); !errors.Is(err, model.ErrUngueltig) {
				t.Errorf("%s: Schliessen(%d) = %v, erwartet ErrUngueltig", z.name, int64(f), err)
			}
			z.pruefe(t, u, v)
			if err := u.srv.Schliessen(0); err != nil {
				t.Errorf("%s: Schliessen(0) = %v, erwartet nil", z.name, err)
			}
		}
	}
	gleichAuf(t, "Zahl der Elemente", n, 6)
}

// TestSchliessenFristen hält S8 (SPEC-136 Nr. 14): Schliessen wartet höchstens frist (0: nicht) und kehrt früher
// zurück, wenn alle Requests fertig sind; ohne laufenden Request ist die Rückgabe nil.
func TestSchliessenFristen(t *testing.T) {
	route := neueHaengRoute(t, wartetCtx)
	u := neuLive(t, route)
	_ = u.async("GET", "/p")
	route.warteBegonnen(t, 1)
	if _, err := schliesse(t, u.srv, 0); !errors.Is(err, admin.ErrFristAbgelaufen) {
		t.Errorf("Schliessen(0) mit laufendem Request = %v, erwartet ErrFristAbgelaufen", err)
	}
	gleichAuf(t, "laufende Handler nach Schliessen(0)", route.laufend.Load(), int64(0))

	frei := neueHaengRoute(t, wartetFrei)
	f := neuLive(t, frei)
	r := f.async("GET", "/p")
	frei.warteBegonnen(t, 1)
	time.AfterFunc(100*time.Millisecond, frei.gibFrei)
	if dauer, err := schliesse(t, f.srv, 30*time.Second); err != nil || dauer < 90*time.Millisecond || dauer >= 5*time.Second {
		t.Errorf("Schliessen(30 s) = %v nach %v, erwartet nil in [90 ms, 5 s)", err, dauer)
	}
	gleichAuf(t, "laufende Handler bei der Rückkehr: Schliessen wartete auf den Handler", frei.laufend.Load(), int64(0))
	gleichAuf(t, "Fehler von ctx", frei.ctxFehlerListe(), []error{nil})
	if x := erwarte(t, r, "Antwort"); x.err != nil || x.status != 200 {
		t.Errorf("Antwort: %+v", x)
	}

	fristen := []time.Duration{0, 1, 30 * time.Second, math.MaxInt64}
	gleichAuf(t, "Zahl der Fristen", len(fristen), 4)
	for _, fr := range fristen {
		v := neuLive(t, neueRoute())
		if dauer, err := schliesse(t, v.srv, fr); err != nil || dauer >= 5*time.Second {
			t.Errorf("Schliessen(%d) ohne Request = %v nach %v, erwartet nil in unter 5 s", int64(fr), err, dauer)
		}
	}

	// Frist 0 und 1 ns ohne Request, 64 frische Server ohne Serve: sind der leere Zähler und der abgelaufene Timer
	// zugleich bereit, darf die Wahl nicht dem Zufall gehören.
	for i := range 64 {
		raw, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = raw.Close() })
		if err := neuMitListener(t, raw).Schliessen(time.Duration(i % 2)); err != nil {
			t.Fatalf("Schliessen(%d ns) ohne Request im Lauf %d = %v, erwartet nil", i%2, i, err)
		}
	}

	// Die größte Frist mit laufendem Request: sie läuft nicht ab, der Request endet von selbst.
	g := neueHaengRoute(t, wartetFrei)
	w := neuLive(t, g)
	_ = w.async("GET", "/p")
	g.warteBegonnen(t, 1)
	time.AfterFunc(100*time.Millisecond, g.gibFrei)
	if _, err := schliesse(t, w.srv, math.MaxInt64); err != nil {
		t.Errorf("Schliessen(MaxInt64) mit laufendem Request = %v, erwartet nil", err)
	}
	gleichAuf(t, "Fehler von ctx", g.ctxFehlerListe(), []error{nil})
}

// TestSchliessenKontextNachFrist hält S9 und X2 Fall 2 und 3 (SPEC-136 Nr. 14, SPEC-238 Nr. 9): läuft die Frist ab,
// endet der ctx jeder laufenden Route, auch einer mit ungelesenem Body; Schliessen wartet danach auf den Handler und
// liefert ErrFristAbgelaufen, nicht ErrUngueltig. Der Handler braucht nach dem Ende von ctx noch 150 ms.
func TestSchliessenKontextNachFrist(t *testing.T) {
	faelle := []struct {
		name string
		los  func(t *testing.T, u *live)
	}{
		{"GET", func(t *testing.T, u *live) { _ = u.async("GET", "/p") }},
		{"POST, angekündigt 100, 3 gesendet, ungelesen", func(t *testing.T, u *live) {
			v := u.verbinde(t)
			v.schreibe(t, []byte("POST /p HTTP/1.1\r\nHost: h\r\n"+kopfAuth()+"Content-Length: 100\r\n\r\nabc"))
		}},
	}
	fristen := []time.Duration{100 * time.Millisecond, 400 * time.Millisecond}
	gleichAuf(t, "Zahl der Fristen", len(fristen), 2)
	for _, f := range faelle {
		for _, frist := range fristen {
			t.Run(fmt.Sprintf("%s/%v", f.name, frist), func(t *testing.T) {
				route := neueHaengRoute(t, wartetCtx)
				route.endeVerz = 150 * time.Millisecond
				u := neuLive(t, route)
				f.los(t, u)
				route.warteBegonnen(t, 1)
				start := time.Now()
				_, err := schliesse(t, u.srv, frist)
				if !errors.Is(err, admin.ErrFristAbgelaufen) || errors.Is(err, model.ErrUngueltig) {
					t.Errorf("Schliessen(%v) = %v, erwartet ErrFristAbgelaufen und nicht ErrUngueltig", frist, err)
				}
				gleichAuf(t, "laufende Handler bei der Rückkehr", route.laufend.Load(), int64(0))
				gleichAuf(t, "Fehler von ctx", route.ctxFehlerListe(), []error{context.Canceled})
				// Untere Schranke: von Aufrufbeginn bis zum Ende von ctx vergeht mindestens frist (keine Obergrenze, N-4).
				enden := route.ctxEndenListe()
				if len(enden) != 1 {
					t.Fatalf("%d Enden von ctx, erwartet 1", len(enden))
				}
				if dauer := enden[0].Sub(start); dauer < frist {
					t.Errorf("ctx endete nach %v, erwartet mindestens %v", dauer, frist)
				}
			})
		}
	}
}

// TestSchliessenZweimal hält S10 (SPEC-136 Nr. 14): nach der Rückkehr läuft kein Handler und keiner beginnt; ein
// weiteres Schliessen liefert nil und wartet nicht, auch während das erste wartet; gleichzeitige Aufrufe geraten nicht
// in Panik.
func TestSchliessenZweimal(t *testing.T) {
	route := neueHaengRoute(t, wartetFrei)
	route.nurPfad = "/haengt"
	u := neuLive(t, route)
	v1 := u.verbinde(t)
	v1.erledige(t)
	route.warteBegonnen(t, 1)
	_ = u.async("GET", "/haengt")
	route.warteBegonnen(t, 1)
	erstes := make(chan error, 1)
	go func() { erstes <- u.srv.Schliessen(30 * time.Second) }()
	// V1 im Leerlauf schließt Shutdown als Erstes: dann hat das erste Schliessen die Zulassung passiert.
	if n, err := v1.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("V1: %d Bytes, Fehler %v; erwartet EOF", n, err)
	}
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= time.Second {
		t.Errorf("zweites Schliessen = %v nach %v, erwartet nil in unter 1 s", err, dauer)
	}
	route.gibFrei()
	select {
	case err := <-erstes:
		if err != nil {
			t.Errorf("erstes Schliessen = %v, erwartet nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("erstes Schliessen kehrte nach der Freigabe nicht zurück")
	}
	gleichAuf(t, "laufende Handler nach der Rückkehr", route.laufend.Load(), int64(0))
	if !dialScheitert(u.addr) {
		t.Error("nach Schliessen gelingt ein Dial")
	}
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= time.Second {
		t.Errorf("drittes Schliessen = %v nach %v, erwartet nil in unter 1 s", err, dauer)
	}

	w := neuLive(t, neueRoute())
	var wg sync.WaitGroup
	fehler := make([]error, 8)
	for i := range fehler {
		wg.Add(1)
		go func() { defer wg.Done(); fehler[i] = w.srv.Schliessen(0) }()
	}
	wg.Wait()
	for i, err := range fehler {
		if err != nil {
			t.Errorf("gleichzeitiges Schliessen(0) %d = %v, erwartet nil", i, err)
		}
	}
}

// TestVerbindungszustaende hält S11 (SPEC-136 Nr. 14): Schliessen schließt jede Verbindung ohne laufenden Request
// sofort (im Leerlauf, neu ohne ein Byte, mit halbem Kopf) und lässt die mit einem laufenden Request offen (halber Body
// mit hängendem Handler, laufender Request), bis dieser antwortet. 5 Elemente.
func TestVerbindungszustaende(t *testing.T) {
	route := neueHaengRoute(t, wartetFrei)
	route.nurPfad = "/haengt"
	u := neuLive(t, route)
	leerlauf := u.verbinde(t)
	leerlauf.erledige(t)
	route.warteBegonnen(t, 1)
	neu := u.verbinde(t)
	kopf := u.verbinde(t)
	kopf.schreibe(t, []byte("GET /p HTTP/1.1\r\nHo"))
	halbBody := u.verbinde(t)
	halbBody.schreibe(t, []byte("POST /haengt HTTP/1.1\r\nHost: h\r\n"+kopfAuth()+"Content-Length: 100\r\n\r\nabc"))
	route.warteBegonnen(t, 1)
	laeuft := u.verbinde(t)
	laeuft.schreibe(t, []byte("GET /haengt HTTP/1.1\r\nHost: h\r\n"+kopfAuth()+"\r\n"))
	route.warteBegonnen(t, 1)
	// Der Server hat die Verbindungen angenommen und vermerkt, ehe Schliessen sie ansieht.
	time.Sleep(100 * time.Millisecond)

	schluss := make(chan error, 1)
	go func() { schluss <- u.srv.Schliessen(30 * time.Second) }()
	ohne := map[string]*verbindung{"Leerlauf": leerlauf, "neu ohne Byte": neu, "halber Kopf": kopf}
	mit := map[string]*verbindung{"halber Body": halbBody, "laufender Request": laeuft}
	gleichAuf(t, "Zahl der Verbindungszustände", len(ohne)+len(mit), 5)
	for name, v := range ohne {
		if n, err := v.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
			t.Errorf("%s: %d Bytes, Fehler %v; erwartet EOF", name, n, err)
		}
	}
	for name, v := range mit {
		if n, err := v.ende(300 * time.Millisecond); n != 0 || !istZeitgrenze(err) {
			t.Errorf("%s: %d Bytes, Fehler %v; erwartet offen (Zeitgrenze)", name, n, err)
		}
	}
	route.gibFrei()
	if resp, _ := laeuft.antwort(t, "GET"); resp.StatusCode != 200 {
		t.Errorf("laufender Request: Status %d, erwartet 200", resp.StatusCode)
	}
	if resp, _ := halbBody.antwort(t, "POST"); resp.StatusCode != 200 {
		t.Errorf("halber Body: Status %d, erwartet 200", resp.StatusCode)
	}
	// Nach der Antwort schließt der Server jede Verbindung mit Handler (Shutdown), sie bleibt nicht im Leerlauf. Der
	// Server verwirft den ungelesenen Rest des Bodys erst, wenn er da ist: der Client sendet die fehlenden 97 Bytes.
	halbBody.schreibe(t, bytes.Repeat([]byte("x"), 97))
	for name, v := range mit {
		if n, err := v.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
			t.Errorf("%s nach der Antwort: %d Bytes, Fehler %v; erwartet EOF", name, n, err)
		}
	}
	select {
	case err := <-schluss:
		if err != nil {
			t.Errorf("Schliessen(30 s) = %v, erwartet nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Schliessen kehrte nach der Antwort nicht zurück")
	}
}

// TestLeerlaufUndAbgewiesene hält S11 (SPEC-136 Nr. 13 b, 14): Anhalten lässt eine Verbindung im Leerlauf offen,
// Schliessen schließt sie; ein abgewiesener Request (503) mit unfertigem Body zählt nicht als laufend und hält die
// Rückgabe von Schliessen nicht auf.
func TestLeerlaufUndAbgewiesene(t *testing.T) {
	route := neueHaengRoute(t, wartetFrei)
	route.nurPfad = "/haengt"
	u := neuLive(t, route)
	v1, v2 := u.verbinde(t), u.verbinde(t)
	v1.erledige(t)
	v2.erledige(t)
	route.warteBegonnen(t, 2)
	laeuft := u.async("GET", "/haengt")
	route.warteBegonnen(t, 1)
	u.srv.Anhalten()
	if n, err := v1.ende(200 * time.Millisecond); n != 0 || !istZeitgrenze(err) {
		t.Errorf("V1 nach Anhalten: %d Bytes, Fehler %v; erwartet offen (Zeitgrenze)", n, err)
	}
	v2.schreibe(t, []byte("POST /p HTTP/1.1\r\nHost: h\r\n"+kopfAuth()+"Content-Length: 100\r\n\r\nabc"))
	if resp, _ := v2.antwort(t, "POST"); resp.StatusCode != 503 {
		t.Errorf("V2 nach Anhalten: Status %d, erwartet 503", resp.StatusCode)
	}
	// Der laufende Request endet nach dem abgewiesenen: dessen 503 darf Schliessen nicht aufhalten.
	route.gibFrei()
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= 5*time.Second {
		t.Errorf("Schliessen(30 s) = %v nach %v, erwartet nil in unter 5 s", err, dauer)
	}
	if x := erwarte(t, laeuft, "Antwort des laufenden Requests"); x.err != nil || x.status != 200 {
		t.Errorf("laufender Request: %+v", x)
	}
	if n, err := v1.ende(3 * time.Second); n != 0 || !errors.Is(err, io.EOF) {
		t.Errorf("V1 nach Schliessen: %d Bytes, Fehler %v; erwartet EOF", n, err)
	}
}

// TestSchliessenOhneAnhalten hält S12 (SPEC-136 Nr. 14): Schliessen ohne vorheriges Anhalten hält an.
func TestSchliessenOhneAnhalten(t *testing.T) {
	u := neuLive(t, neueRoute())
	if _, err := schliesse(t, u.srv, time.Second); err != nil {
		t.Errorf("Schliessen(1 s) = %v, erwartet nil", err)
	}
	if !dialScheitert(u.addr) {
		t.Error("nach Schliessen gelingt ein Dial")
	}
	if err := serveEnde(t, u.serve, "nach Schliessen"); err != nil {
		t.Errorf("Serve lieferte %v, erwartet nil", err)
	}

	// Ohne Serve schließt allein Schliessen den Listener: kein Shutdown, der ihn verfolgte.
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	zl := &langsamListener{Listener: raw, verz: time.Millisecond}
	if err := neuMitListener(t, zl).Schliessen(time.Second); err != nil {
		t.Errorf("Schliessen(1 s) ohne Serve = %v, erwartet nil", err)
	}
	if !dialScheitert(raw.Addr().String()) {
		t.Error("ohne Serve gelingt nach Schliessen ein Dial")
	}
	gleichAuf(t, "Aufrufe von Close ohne Serve", zl.closes.Load(), int64(1))
}

// TestKeineSperreUeberAufruf hält S13 (SPEC-238 Nr. 12): die Sperre liegt nie über einem Aufruf von Handler oder
// Vergleich. Während die Route (Stelle 1) oder der Vergleich (Stelle 2) blockiert, kehren Anhalten zurück und ein
// zweiter Request wird abgewiesen; Schliessen(0) endet, sobald der Aufruf frei ist.
func TestKeineSperreUeberAufruf(t *testing.T) {
	stellen := []string{"Route", "Vergleich"}
	for _, stelle := range stellen {
		route := neueHaengRoute(t, wartetFrei)
		route.nurPfad = "/haengt"
		frei := make(chan struct{})
		var einmal sync.Once
		gibFrei := func() { einmal.Do(func() { close(frei) }) }
		t.Cleanup(gibFrei)
		begonnen := make(chan struct{}, 4)
		var aufrufe atomic.Int64
		opt := liveOpt{}
		if stelle == "Vergleich" {
			// Der erste Aufruf (der Leerlauf-Request) kehrt zurück, der zweite blockiert.
			opt.vergleich = func(erwartet, gegeben []byte) bool {
				if aufrufe.Add(1) >= 2 {
					begonnen <- struct{}{}
					<-frei
				}
				return admin.Vergleiche(erwartet, gegeben)
			}
		}
		u := neuLiveMit(t, route, opt)
		v := u.verbinde(t)
		v.erledige(t)
		route.warteBegonnen(t, 1)
		r := u.async("GET", "/haengt")
		if stelle == "Route" {
			route.warteBegonnen(t, 1)
		} else {
			<-begonnen
		}
		if !innerhalb(2*time.Second, u.srv.Anhalten) {
			t.Fatalf("%s: Anhalten kehrte in 2 s nicht zurück", stelle)
		}
		v.schreibe(t, getP())
		if resp, _ := v.antwort(t, "GET"); resp.StatusCode != 503 {
			t.Errorf("%s: zweiter Request mit Status %d, erwartet 503", stelle, resp.StatusCode)
		}
		schluss := make(chan error, 1)
		go func() { schluss <- u.srv.Schliessen(0) }()
		select {
		case err := <-schluss:
			t.Fatalf("%s: Schliessen(0) kehrte vor der Freigabe zurück: %v", stelle, err)
		case <-time.After(100 * time.Millisecond):
		}
		route.gibFrei()
		gibFrei()
		select {
		case err := <-schluss:
			if !errors.Is(err, admin.ErrFristAbgelaufen) {
				t.Errorf("%s: Schliessen(0) = %v, erwartet ErrFristAbgelaufen", stelle, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: Schliessen(0) kehrte nach der Freigabe nicht zurück", stelle)
		}
		erwarte(t, r, stelle+": Antwort")
	}
	gleichAuf(t, "Zahl der Stellen", len(stellen), 2)
}

// TestFuenfzigHandlerEndenGleichzeitig hält S13: 50 zugelassene Handler enden gleichzeitig, der Zähler ist danach 0;
// Anhalten läuft erst nach den 50 Antworten.
func TestFuenfzigHandlerEndenGleichzeitig(t *testing.T) {
	const n = 50
	route := neueHaengRoute(t, wartetFrei)
	u := neuLive(t, route)
	res := make([]<-chan ergebnis, n)
	for i := range res {
		res[i] = u.async("GET", "/p")
	}
	route.warteBegonnen(t, n)
	route.gibFrei()
	for i, c := range res {
		if x := erwarte(t, c, "Antwort"); x.err != nil || x.status != 200 {
			t.Errorf("Request %d: %+v", i, x)
		}
	}
	u.srv.Anhalten()
	if _, err := schliesse(t, u.srv, time.Second); err != nil {
		t.Errorf("Schliessen(1 s) = %v, erwartet nil", err)
	}
	gleichAuf(t, "laufende Handler nach Schliessen", route.laufend.Load(), int64(0))
	gleichAuf(t, "Aufrufe der Route", route.aufrufe(), n)
}

// TestZaehlerUnterLast hält S13 als Wirkungs-Fänger neben dem Detektor: in 10 Runden enden 64 zugelassene Handler
// gleichzeitig, während 64 neue Requests eintreffen; danach ist kein Handler mehr gezählt, und Schliessen kehrt ohne
// Ablauf der Frist zurück (ein verlorener Zähler ließe es auf die Frist warten).
func TestZaehlerUnterLast(t *testing.T) {
	const n = 64
	for runde := range 10 {
		route := neueHaengRoute(t, wartetFrei)
		route.nurPfad = "/haengt"
		u := neuLive(t, route)
		alt := make([]<-chan ergebnis, n)
		for i := range alt {
			alt[i] = u.async("GET", "/haengt")
		}
		route.warteBegonnen(t, n)
		neu := make([]<-chan ergebnis, n)
		route.gibFrei()
		for i := range neu {
			neu[i] = u.async("GET", "/p")
		}
		for i, c := range append(alt, neu...) {
			if x := erwarte(t, c, "Antwort"); x.err != nil || x.status != 200 {
				t.Errorf("Runde %d, Request %d: %+v", runde, i, x)
			}
		}
		gleichAuf(t, "laufende Handler nach den Antworten", route.laufend.Load(), int64(0))
		if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= 3*time.Second {
			t.Errorf("Runde %d: Schliessen(30 s) = %v nach %v, erwartet nil in unter 3 s", runde, err, dauer)
		}
	}
}

// TestHandlerPanikZaehltNichtWeiter hält S13: ein Handler in Panik verlässt die Zulassung; Schliessen kehrt zurück.
func TestHandlerPanikZaehltNichtWeiter(t *testing.T) {
	route := neueRoute()
	route.aendere = func(admin.Anfrage) { panic(http.ErrAbortHandler) }
	u := neuLive(t, route)
	if x := erwarte(t, u.async("GET", "/p"), "Antwort"); x.err == nil {
		t.Errorf("der Client sah keinen Abbruch, Status %d", x.status)
	}
	if dauer, err := schliesse(t, u.srv, 30*time.Second); err != nil || dauer >= 5*time.Second {
		t.Errorf("Schliessen(30 s) nach einer Panik = %v nach %v, erwartet nil in unter 5 s", err, dauer)
	}
}
