package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// taktUhr ist eine treibende Uhr, die beim k-ten eigenen Vorstellen erst weiterstellt, wenn k Aufrufe ankamen.
func taktUhr(angekommen *atomic.Int32) (*treibUhr, time.Time) {
	u := neueTreibUhr()
	t0 := u.Jetzt()
	var schritte atomic.Int32
	u.vorher = func(time.Duration, time.Time) {
		n := schritte.Add(1)
		bisWahr(func() bool { return angekommen.Load() >= n })
	}
	return u, t0
}

// starts fährt anzahl Aufrufe bei 20 je Sekunde und liefert die Startzeit jedes Aufrufs seit dem Beginn; ueberschuss verschiebt jeden
// Weckzeitpunkt der Uhr.
func starts(anzahl int, ueberschuss time.Duration) ([]time.Duration, taktStand) {
	var angekommen atomic.Int32
	u, t0 := taktUhr(&angekommen)
	u.extra = func(time.Time) time.Duration { return ueberschuss }
	out := make([]time.Duration, anzahl)
	st := fahreTakt(context.Background(), u, 20, anzahl, func(_ context.Context, k int) {
		out[k] = u.Jetzt().Sub(t0)
		angekommen.Add(1)
	})
	return out, st
}

// TestTakt (SPEC-216 Nr. 8 a): Anfrage k startet zur Zeit k × 50 ms aus k, auch wenn die Anfrage 3 hängt und auch wenn jedes Wecken 1 ms
// später kommt; die Zahl ist bei den Profilen abnahme und rauch 13200 und 260.
func TestTakt(t *testing.T) {
	wache(t, 120*time.Second)
	literal := []time.Duration{0, 50 * time.Millisecond, 100 * time.Millisecond, 150 * time.Millisecond, 200 * time.Millisecond}
	for _, f := range []struct {
		p       Profil
		anzahl  int
		letzter time.Duration
	}{{abnahme(), 13200, 659950 * time.Millisecond}, {rauch(), 260, 12950 * time.Millisecond}} {
		s, st := starts(f.p.Anfragen(), 0)
		if len(s) != f.anzahl || st.Gestartet != f.anzahl || f.p.Rate != 20 {
			t.Fatalf("%s: %d Anfragen, %d gestartet bei Rate %d; erwartet %d bei 20", f.p.Name, len(s), st.Gestartet, f.p.Rate, f.anzahl)
		}
		for k, want := range literal {
			if s[k] != want {
				t.Errorf("%s: Anfrage %d startet bei %v, erwartet %v", f.p.Name, k, s[k], want)
			}
		}
		if s[f.anzahl-1] != f.letzter {
			t.Errorf("%s: Anfrage %d startet bei %v, erwartet %v", f.p.Name, f.anzahl-1, s[f.anzahl-1], f.letzter)
		}
	}
	// Drift: jedes Wecken kommt 1 ms zu spät, die Fälligkeit bleibt k × 50 ms.
	s, _ := starts(13200, time.Millisecond)
	if s[13199] != 659951*time.Millisecond || s[1] != 51*time.Millisecond {
		t.Errorf("mit 1 ms Überschuss: Starts 1 und 13199 bei %v und %v, erwartet 51ms und 10m59.951s", s[1], s[13199])
	}
}

// TestTaktWartetNicht: Anfrage 3 kehrt erst zurück, nachdem Anfrage 7 gestartet ist; die Starts 4 bis 7 folgen im Takt.
func TestTaktWartetNicht(t *testing.T) {
	wache(t, 60*time.Second)
	var angekommen atomic.Int32
	u, t0 := taktUhr(&angekommen)
	frei, sieben := make(chan struct{}), make(chan time.Duration, 1)
	out := make([]time.Duration, 10)
	fertig := make(chan taktStand, 1)
	go func() {
		fertig <- fahreTakt(context.Background(), u, 20, 10, func(_ context.Context, k int) {
			out[k] = u.Jetzt().Sub(t0)
			angekommen.Add(1)
			switch k {
			case 3:
				<-frei
			case 7:
				sieben <- out[7]
			}
		})
	}()
	if got := warteAuf(t, sieben, 5*time.Second); got != 350*time.Millisecond {
		t.Errorf("Anfrage 7 startet bei %v, erwartet 350ms", got)
	}
	close(frei)
	warteAuf(t, fertig, 5*time.Second)
	for k := 4; k < 8; k++ {
		if out[k] != time.Duration(k)*50*time.Millisecond {
			t.Errorf("Anfrage %d startet bei %v", k, out[k])
		}
	}
}

// TestTaktEndetMitKontext: nach dem Ende des Kontexts startet keine weitere Anfrage, auch keine, deren Fälligkeit die Uhr schon erreicht hat.
func TestTaktEndetMitKontext(t *testing.T) {
	wache(t, 30*time.Second)
	zu, beende := context.WithCancel(context.Background())
	beende()
	if st := fahreTakt(zu, neueFakeUhr(), 20, 3, func(context.Context, int) {}); st.Gestartet != 0 {
		t.Errorf("beendeter Kontext: %d gestartet", st.Gestartet)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var angekommen atomic.Int32
	u, _ := taktUhr(&angekommen)
	st := fahreTakt(ctx, u, 20, 100, func(_ context.Context, k int) {
		if k == 4 {
			cancel()
		}
		angekommen.Add(1)
	})
	if st.Gestartet != 5 || angekommen.Load() != 5 {
		t.Errorf("%d gestartet, %d angekommen, erwartet 5", st.Gestartet, angekommen.Load())
	}
}

// TestTaktNebenProbe (SPEC-216 Nr. 11): eine Probe der Last, die ab dem Tick bei 30 s hängt und erst bei der Anfrage 1900 freigegeben
// wird, hält den Takt nicht an: jede der 2000 Anfragen startet zur Zeit k × 50 ms.
func TestTaktNebenProbe(t *testing.T) {
	wache(t, 120*time.Second)
	p := kurzesProfil("klein", 1)
	p.Warmup, p.Messphase = 40*time.Second, 60*time.Second
	s := neuerProbestand(t, p, "replay")
	frei := make(chan struct{})
	var einmal sync.Once
	s.last.halte = func(i int) {
		if i == 2 {
			<-frei
		}
	}
	s.dienst.halte = func(_, k int, _ *http.Request) {
		if k == 1900 {
			einmal.Do(func() { close(frei) })
		}
	}
	exit := s.fahre(context.Background())
	if code := warteAuf(t, exit, 30*time.Second); code != 0 {
		t.Fatalf("Exit %d, Meldungen %q", code, s.fehler.String())
	}
	b := s.dienst.lauf(0)
	for k := range 2000 {
		if d := b.zeit(k); d != time.Duration(k)*50*time.Millisecond {
			t.Fatalf("Anfrage %d startet bei %v", k, d)
		}
	}
}

// TestGleichzeitig (SPEC-216 Nr. 8 b, Wirkungs-Fänger): fünf Anfragen, die der Dienst erst beantwortet, wenn alle fünf bei ihm sind, zählen
// gleichzeitig_max 5; keine Obergrenze der Verbindungen hält eine zurück.
func TestGleichzeitig(t *testing.T) {
	wache(t, 60*time.Second)
	for lauf := range 5 {
		var n atomic.Int32
		alle, aufgeben := make(chan struct{}), make(chan struct{})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if n.Add(1) == 5 {
				close(alle)
			}
			select {
			case <-alle:
			case <-aufgeben:
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		c := lastClient(nil)
		ruhig := neueFakeUhr()
		var fehler atomic.Int32
		fertig := make(chan taktStand, 1)
		go func() {
			fertig <- fahreTakt(context.Background(), neueTreibUhr(), 20, 5, func(ctx context.Context, _ int) {
				if _, err := holeLast(ctx, ruhig, c, srv.URL, nil); err != nil {
					fehler.Add(1)
				}
			})
		}()
		select {
		case <-alle:
		case <-wartenKanal(2 * time.Second):
			close(aufgeben)
		}
		st := warteAuf(t, fertig, 10*time.Second)
		if st.GleichzeitigMax != 5 || fehler.Load() != 0 {
			t.Errorf("Lauf %d: gleichzeitig_max %d, Fehler %d; erwartet 5 und 0", lauf, st.GleichzeitigMax, fehler.Load())
		}
		c.CloseIdleConnections()
		srv.Close()
	}
}

// wartenKanal schließt nach d der wirklichen Zeit.
func wartenKanal(d time.Duration) <-chan struct{} {
	c := make(chan struct{})
	go func() {
		pause(d)
		close(c)
	}()
	return c
}

// TestVerspaetung: weckt die Uhr die Anfrage 3 sieben Millisekunden nach ihrer Fälligkeit, ist die größte Verspätung 7 ms; die übrigen
// Anfragen starten pünktlich, und gleichzeitig_max ist 1.
func TestVerspaetung(t *testing.T) {
	wache(t, 60*time.Second)
	u := neueFakeUhr()
	t0 := u.Jetzt()
	angekommen := make(chan int, 1)
	fertig := make(chan taktStand, 1)
	var zurueck atomic.Int32
	go func() {
		fertig <- fahreTakt(context.Background(), u, 20, 6, func(_ context.Context, k int) {
			zurueck.Add(1)
			angekommen <- k
		})
	}()
	warteAuf(t, angekommen, 5*time.Second)
	for k := 1; k < 6; k++ {
		bis := t0.Add(time.Duration(k) * 50 * time.Millisecond)
		if k == 3 {
			bis = bis.Add(7 * time.Millisecond)
		}
		bisWahr(func() bool { return u.wartende() >= 1 && zurueck.Load() == int32(k) })
		pause(time.Millisecond)
		u.Vorstellen(bis.Sub(u.Jetzt()))
		warteAuf(t, angekommen, 5*time.Second)
	}
	st := warteAuf(t, fertig, 5*time.Second)
	if st.VerspaetungMax != 7*time.Millisecond || st.GleichzeitigMax != 1 || st.Gestartet != 6 {
		t.Errorf("Stand %+v; erwartet 6 gestartet, Verspätung 7ms, gleichzeitig 1", st)
	}
}

// zaehlServer ist ein Dienst, der je Pfad die Anfragen und die neuen Verbindungen zählt.
type zaehlServer struct {
	*httptest.Server
	verbindungen atomic.Int32
	mu           sync.Mutex
	je           map[string]int
	haelt        chan struct{}
}

func (s *zaehlServer) anzahl(pfad string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.je[pfad]
}

const antwortGross = 65536

func neuerZaehlServer(t *testing.T) *zaehlServer {
	t.Helper()
	s := &zaehlServer{je: map[string]int{}, haelt: make(chan struct{})}
	s.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.je[r.URL.Path]++
		s.mu.Unlock()
		switch r.URL.Path {
		case "/gut":
			if r.ContentLength != 0 || len(r.TransferEncoding) != 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = io.WriteString(w, strings.Repeat("a", antwortGross))
		case "/404", "/500", "/503":
			status := map[string]int{"/404": 404, "/500": 500, "/503": 503}[r.URL.Path]
			w.WriteHeader(status)
			_, _ = io.WriteString(w, "fehler")
		case "/204":
			w.WriteHeader(http.StatusNoContent)
		case "/302":
			w.Header().Set("Location", "/gut")
			w.WriteHeader(http.StatusFound)
		case "/transport":
			panic(http.ErrAbortHandler)
		case "/kurz":
			w.Header().Set("Content-Length", "100")
			_, _ = io.WriteString(w, "0123456789")
			w.(http.Flusher).Flush()
			panic(http.ErrAbortHandler)
		case "/haengt":
			select {
			case <-s.haelt:
			case <-r.Context().Done():
			}
		}
	}))
	s.Config.ConnState = func(_ net.Conn, z http.ConnState) {
		if z == http.StateNew {
			s.verbindungen.Add(1)
		}
	}
	s.Start()
	t.Cleanup(func() {
		close(s.haelt)
		s.Close()
	})
	return s
}

// TestAntwort (SPEC-216 Nr. 8 c): der Body wird vollständig gelesen; jede der neun Formen endet genau einmal als Erfolg oder Fehler mit
// einer Anfrage am Dienst: 200 mit 64 KiB, 404, 500, 503, 204, Weiterleitung, Transportfehler, Zeitüberschreitung bei 10 s der Uhr (bei
// 10 s − 1 ns noch offen) und Lesefehler des Bodys.
func TestAntwort(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerZaehlServer(t)
	c := lastClient(nil)
	defer c.CloseIdleConnections()
	uhr := neueFakeUhr()
	formen := []struct {
		pfad    string
		gelesen int64
		fehler  bool
	}{
		{"/gut", antwortGross, false}, {"/404", 6, true}, {"/500", 6, true}, {"/503", 6, true}, {"/204", 0, true}, {"/302", 0, true},
		{"/transport", 0, true}, {"/kurz", 10, true},
	}
	for _, f := range formen {
		n, err := holeLast(context.Background(), uhr, c, s.URL+f.pfad, [][2]string{{headerAccept, acceptJSON}})
		if n != f.gelesen || (err != nil) != f.fehler || s.anzahl(f.pfad) != 1 {
			t.Errorf("%s: %d Bytes, Fehler %v, %d Anfragen; erwartet %d, Fehler %v und eine Anfrage", f.pfad, n, err, s.anzahl(f.pfad), f.gelesen, f.fehler)
		}
	}
	erg := make(chan error, 1)
	go func() {
		_, err := holeLast(context.Background(), uhr, c, s.URL+"/haengt", nil)
		erg <- err
	}()
	bisWahr(func() bool { return s.anzahl("/haengt") == 1 })
	uhr.Vorstellen(10*time.Second - time.Nanosecond)
	bleibtOffen(t, erg, "die Anfrage vor der Frist von 10 s")
	uhr.Vorstellen(time.Nanosecond)
	if err := warteAuf(t, erg, 5*time.Second); err == nil {
		t.Error("an der Frist von 10 s: kein Fehler")
	}
	if n := len(formen) + 1; n != 9 || s.anzahl("/gut") != 1 {
		t.Errorf("%d Formen, %d Anfragen an /gut; erwartet 9 und 1 (keine Weiterleitung folgt)", n, s.anzahl("/gut"))
	}
}

// TestVerbindungen (SPEC-216 Nr. 8 b): zehn Anfragen nacheinander nehmen eine Verbindung; zwei Wellen zu fünf gleichzeitigen Anfragen öffnen
// höchstens so viele Verbindungen wie die größte Zahl gleichzeitig offener.
func TestVerbindungen(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerZaehlServer(t)
	c := lastClient(nil)
	defer c.CloseIdleConnections()
	summe := int64(0)
	for range 10 {
		n, err := holeLast(context.Background(), neueFakeUhr(), c, s.URL+"/gut", nil)
		if err != nil {
			t.Fatal(err)
		}
		summe += n
	}
	if summe != 10*antwortGross || s.verbindungen.Load() != 1 {
		t.Errorf("%d Bytes, %d Verbindungen; erwartet %d und 1", summe, s.verbindungen.Load(), 10*antwortGross)
	}
	var da atomic.Int32
	var tor atomic.Pointer[chan struct{}]
	h := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		ch := *tor.Load()
		if da.Add(1)%5 == 0 {
			close(ch)
		}
		<-ch
	}))
	var neue atomic.Int32
	h.Config.ConnState = func(_ net.Conn, z http.ConnState) {
		if z == http.StateNew {
			neue.Add(1)
		}
	}
	h.Start()
	defer h.Close()
	for range 2 {
		ch := make(chan struct{})
		tor.Store(&ch)
		var wg sync.WaitGroup
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := holeLast(context.Background(), neueFakeUhr(), c, h.URL, nil); err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()
	}
	if neue.Load() > 5 {
		t.Errorf("%d neue Verbindungen, erwartet höchstens 5", neue.Load())
	}
}

// TestKeineWiederholung (SPEC-216 Nr. 8 d): bricht der Dienst eine Anfrage auf einer wiederverwendeten Verbindung ab, geht sie nicht ein
// zweites Mal an ihn und ist ein Fehler.
func TestKeineWiederholung(t *testing.T) {
	wache(t, 60*time.Second)
	s := neuerZaehlServer(t)
	c := lastClient(nil)
	defer c.CloseIdleConnections()
	for i := range 3 {
		if _, err := holeLast(context.Background(), neueFakeUhr(), c, s.URL+"/gut", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := holeLast(context.Background(), neueFakeUhr(), c, s.URL+"/transport", nil); err == nil || s.anzahl("/transport") != i+1 {
			t.Errorf("Abbruch %d: Fehler %v, %d Anfragen an /transport, erwartet einen Fehler und %d", i, err, s.anzahl("/transport"), i+1)
		}
	}
}

// TestFehlerJePhase (SPEC-216 Nr. 8 c, 10 h): ein Fehler allein im Warm-up (erste Anfrage), allein in der ersten und allein in der letzten
// Anfrage der Messphase zählt fehler=1 und macht den Lauf mit grund=fehler ungültig.
func TestFehlerJePhase(t *testing.T) {
	wache(t, 60*time.Second)
	n := 0
	for _, bei := range []int{0, 20, 39} {
		s := neuerProbestand(t, kurzesProfil("klein", 1), "mock")
		s.dienst.status = func(_, k int) int {
			if k == bei {
				return http.StatusServiceUnavailable
			}
			return http.StatusOK
		}
		code := warteAuf(t, s.fahre(context.Background()), 30*time.Second)
		m := s.zeilenVon("Messung")
		if code != 3 || len(m) != 1 || feld(m[0], "fehler") != "1" || feld(m[0], "grund") != "fehler" || feld(m[0], "anfragen") != "40" ||
			len(s.dienst.liste()) != 40 {
			t.Errorf("Fehler bei %d: Exit %d, Zeilen %q, %d Anfragen am Dienst (eine je Fälligkeit, keine Wiederholung)", bei, code, m,
				len(s.dienst.liste()))
		}
		n++
	}
	if n != 3 {
		t.Errorf("%d Fälle, erwartet 3", n)
	}
}

// erwarteteZeilen sind die Anfragen der Liste als Zeilen, wie der Ersatz sie führt.
func erwarteteZeilen(liste []Anfrage) []string {
	var out []string
	for _, a := range liste {
		h := http.Header{}
		for _, k := range a.Header {
			h.Set(k[0], k[1])
		}
		out = append(out, anfrageText(http.MethodGet, "http://127.0.0.1:4711"+a.Ziel(), h))
	}
	return out
}

// TestNurDieListe (SPEC-216 Nr. 8 d): je Betriebsart erreicht den Dienst genau die Liste des Laufs in ihrer Folge und sonst nichts, in den
// Profilen rauch und abnahme (je ein Lauf); vor der Liste steht allein die Verbindung der Bereitschaft.
func TestNurDieListe(t *testing.T) {
	wache(t, 120*time.Second)
	voll := abnahme()
	voll.Laeufe = 1
	n := 0
	for _, p := range []Profil{rauch(), voll} {
		for _, art := range []string{"replay", "mock"} {
			s := neuerProbestand(t, p, art)
			if code := warteAuf(t, s.fahre(context.Background()), 60*time.Second); code > 1 {
				t.Fatalf("%s %s: Exit %d, %q", p.Name, art, code, s.fehler.String())
			}
			liste := listeMock(p, 7, "w-1")
			if art == "replay" {
				liste = listeReplay(p, 7, 1, "w-1", "s-1")
			}
			want, got := erwarteteZeilen(liste), s.dienst.liste()
			if len(got) != len(want) || len(want) != p.Anfragen() {
				t.Fatalf("%s %s: %d Anfragen am Dienst, erwartet %d", p.Name, art, len(got), len(want))
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("%s %s: Anfrage %d ist %q, erwartet %q", p.Name, art, i, got[i], want[i])
				}
			}
			if art == "mock" && strings.Contains(strings.Join(got, "\n"), "$metadata") {
				t.Errorf("%s mock: $metadata in der Last", p.Name)
			}
			erst := strings.Join(s.e.liste()[:5], " ")
			if erst != "last aufbau port start MODE="+art+" verbinde" || strings.Count(strings.Join(s.e.liste(), " "), "verbinde") != 1 {
				t.Errorf("%s %s: Aufrufe %v", p.Name, art, s.e.liste())
			}
			n++
		}
	}
	if n != 4 {
		t.Errorf("%d Fälle, erwartet 4", n)
	}
}

// TestVerbindeTCP (SPEC-216 Nr. 9 b): die Verbindung der Bereitschaft ist eine TCP-Verbindung, auf der nichts gesendet wird; ohne
// lauschenden Dienst ist sie ein Fehler.
func TestVerbindeTCP(t *testing.T) {
	wache(t, 30*time.Second)
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	gelesen := make(chan []byte, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			gelesen <- nil
			return
		}
		b, _ := io.ReadAll(c)
		gelesen <- b
	}()
	if err := verbindeTCP(context.Background(), l.Addr().String()); err != nil {
		t.Fatal(err)
	}
	if b := warteAuf(t, gelesen, 5*time.Second); b == nil || len(b) != 0 {
		t.Errorf("auf der Verbindung der Bereitschaft stehen %q", b)
	}
	_ = l.Close()
	if err := verbindeTCP(context.Background(), l.Addr().String()); err == nil {
		t.Error("Verbindung zu einem geschlossenen Port ohne Fehler")
	}
	adresse, err := freierPort(context.Background())
	if err != nil || !strings.HasPrefix(adresse, "127.0.0.1:") || strings.HasSuffix(adresse, ":0") {
		t.Errorf("freier Port %q, %v", adresse, err)
	}
	if l, err := lc.Listen(context.Background(), "tcp", adresse); err != nil {
		t.Errorf("der gewählte Port ist nicht frei: %v", err)
	} else {
		_ = l.Close()
	}
}
