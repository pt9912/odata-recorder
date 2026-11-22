package httpodata_test

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// gleichzeitigerWeg ist einer der 7 Wege von G14: die Methode wählt, was der Dienst liefert (der Weg 503 hat keinen Dienst), Status und
// Zähler als Literal, mode der Modus des Servers, den der Weg trifft.
type gleichzeitigerWeg struct {
	methode string
	status  int
	code    model.Code
	mode    string
}

func gleichzeitigeWege() []gleichzeitigerWeg {
	return []gleichzeitigerWeg{
		{"GET", 200, 0, "record"}, {"HEAD", 200, 0, "record"},
		{"POST", 409, model.CodeE6009, "record"}, {"PUT", 412, model.CodeE6010, "record"},
		{"DELETE", 503, model.CodeE5010, "hybrid"},
		{"PATCH", 500, model.CodeE9001, "record"},
		{"OPTIONS", 0, 0, "record"},
	}
}

// dienstDerWege liefert, was die Methode des Weges verlangt (SPEC-294 Nr. 7 a).
func dienstDerWege(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	switch a.Methode {
	case "POST":
		return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE6009, Status: 409}
	case "PUT":
		return inbound.Antwort{}, &inbound.Fehler{Code: model.CodeE6010, Status: 412}
	case "PATCH":
		return inbound.Antwort{Status: 600, Body: leer()}, nil
	case "OPTIONS":
		return inbound.Antwort{}, context.Canceled
	}
	return inbound.Antwort{Status: 200, Body: ausBytes([]byte("abc"))}, nil
}

// TestGesamtspanGleichzeitig hält G14 (SPEC-294 Nr. 9 a): n gleichzeitige Anfragen ergeben n StartSpan und n End; jeder Span trägt Methode,
// Modus und Status seiner Anfrage und endet genau einmal, jedes Fehlerobjekt zählt je Code. 16 Goroutinen × 7 Wege × 4 Runden gleich 448
// Spans; im Dienst hält eine Schranke je Aufruf alle 16 zusammen, so dass 16 Spans zugleich offen sind (Wirkungs-Fänger statt Datenrennen).
func TestGesamtspanGleichzeitig(t *testing.T) {
	const goroutinen, runden = 16, 4
	wege := gleichzeitigeWege()
	schranken := make([]*schranke, runden*len(wege))
	for i := range schranken {
		schranken[i] = neueSchranke(goroutinen)
	}
	j := &journal{}
	tel := neuJTel(j)
	tu := func(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		i, _ := strconv.Atoi(erster(a.Header["X-Lauf"]))
		schranken[i].warte()
		return dienstDerWege(ctx, a)
	}
	offen := neuProbeMit(t, "record", j, tel, tu)
	zu := neuProbeMit(t, "hybrid", j, tel, tu)
	zu.srv.Anhalten()
	start := neueSchranke(goroutinen)
	var wg sync.WaitGroup
	for range goroutinen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.warte()
			for r := range runden {
				for i, w := range wege {
					p := offen
					if w.mode == "hybrid" {
						p = zu
					}
					req := jRequest(context.Background(), w.methode, "/p")
					req.Header.Set("X-Lauf", strconv.Itoa(r*len(wege)+i))
					jw := neuJWriter(&journal{})
					p.ruf(jw, req)
					if jw.status != w.status {
						t.Errorf("%s: Status %d, erwartet %d", w.methode, jw.status, w.status)
					}
				}
			}
		}()
	}
	wg.Wait()
	b := buecher(tel.FakeTelemetry)
	if len(b) != goroutinen*len(wege)*runden {
		t.Fatalf("%d Spans, erwartet %d", len(b), goroutinen*len(wege)*runden)
	}
	sollWeg := map[string]gleichzeitigerWeg{}
	for _, w := range wege {
		sollWeg[w.methode] = w
	}
	for _, s := range b {
		m, _ := attrText(s.start, "http.request.method")
		modus, _ := attrText(s.start, "recorder.mode")
		w, ok := sollWeg[m]
		if !ok || modus != w.mode || s.ends != 1 || s.name != "odata.gateway.request" {
			t.Fatalf("Span %+v: Methode %q, Modus %q, End %d; erwartet Modus %q und 1 End", s, m, modus, s.ends, w.mode)
		}
		if w.status == 0 {
			if len(s.sets) != 0 {
				t.Fatalf("%s: Set %v ohne Antwort", m, s.sets)
			}
		} else if len(s.sets) != 1 || len(s.sets[0]) != 1 || s.sets[0][0].Ganz != int64(w.status) {
			t.Fatalf("%s: Set %v, erwartet den Status %d", m, s.sets, w.status)
		}
	}
	je := map[model.Code]int{}
	for _, c := range zaehlerFolge(tel.FakeTelemetry) {
		je[c]++
	}
	n := goroutinen * runden
	gleich(t, "Zähler je Code", je, map[model.Code]int{model.CodeE6009: n, model.CodeE6010: n, model.CodeE5010: n, model.CodeE9001: n})
}

// TestKeineSperreUeberTelemetrie hält G15 (SPEC-294 Nr. 9 b): der Handler hält keine Sperre des Servers über einen Aufruf von Telemetrie; der
// Fake ruft in StartSpan, Set, End und CountError Anhalten (nimmt mu), und der Aufruf kehrt in 5 s zurück. 4 Fälle, je ein Weg, der die Stelle
// erreicht.
func TestKeineSperreUeberTelemetrie(t *testing.T) {
	faelle := []struct {
		op  string
		weg int
	}{{"StartSpan", 0}, {"Set", 0}, {"End", 0}, {"CountError", 2}}
	for _, f := range faelle {
		p := neuProbe(t, "record", wege()[f.weg].tu)
		treffer := 0
		p.tel.setzeBei(func(op string) {
			if strings.HasPrefix(op, f.op) {
				treffer++
				if !innerhalb(5*time.Second, p.srv.Anhalten) {
					t.Errorf("%s: Anhalten blockiert innerhalb des Aufrufs von Telemetrie", f.op)
				}
			}
		})
		jw := neuJWriter(p.j)
		p.ruf(jw, jRequest(context.Background(), "GET", "/p"))
		if treffer != 1 || jw.status == 0 {
			t.Errorf("%s: Haltepunkt %d-mal erreicht, Status %d; erwartet 1 und eine Antwort", f.op, treffer, jw.status)
		}
	}
}

// zaehlAccept zählt die Rückkehr von Accept ohne Fehler.
type zaehlAccept struct {
	net.Listener
	n atomic.Int64
}

func (l *zaehlAccept) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err == nil {
		l.n.Add(1)
	}
	return c, err
}

// TestHerunterfahrenBeendetJedenSpan hält G16 (SPEC-294 Nr. 7 f): 8 Handler (4 mit Antwort, 4 mit Abbruch ohne Antwort) warten an einer
// Schranke, Anhalten, 3 Anfragen auf bestehender Verbindung, Freigabe der 4 mit Antwort (der Test liest ihre Antworten), Freigabe der 4 mit
// Abbruch, Schliessen: bei der Rückkehr haben die 8 zugelassenen Spans je ein End (4 mit dem Status 200, 4 ohne Status), die 3 Anfragen den
// Status 503 mit einem End, und ODR-E5010 zählt 3-mal. Der Haltepunkt von End wartet 50 ms vor der Aufzeichnung: ein End, das nach der Rückkehr
// des Handlers läuft (Goroutine, Abzug aus dem Zähler vor dem End des Abbruchs), fehlt bei der Rückkehr von Schliessen in jedem Lauf.
func TestHerunterfahrenBeendetJedenSpan(t *testing.T) {
	const antworten, abbrueche, abgewiesen = 4, 4, 3
	const zugelassen = antworten + abbrueche
	frei, freiAbbruch, begonnen := make(chan struct{}), make(chan struct{}), make(chan struct{}, zugelassen)
	var al *zaehlAccept
	u := neuUmgebungMit(t, "record", func(_ context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
		begonnen <- struct{}{}
		if a.Pfad == "/abbruch" {
			select {
			case <-freiAbbruch:
			case <-time.After(5 * time.Second):
			}
			return inbound.Antwort{}, context.Canceled
		}
		select {
		case <-frei:
		case <-time.After(5 * time.Second):
		}
		return inbound.Antwort{Status: 200, Body: leer()}, nil
	}, nil, func(l net.Listener) net.Listener { al = &zaehlAccept{Listener: l}; return al })
	u.jtel.setzeBei(func(op string) {
		if op == "End" {
			time.Sleep(50 * time.Millisecond)
		}
	})
	var erg []<-chan ergebnis
	for range antworten {
		erg = append(erg, u.async("GET", "/halt"))
	}
	for range abbrueche {
		erg = append(erg, u.async("GET", "/abbruch"))
	}
	for range zugelassen {
		select {
		case <-begonnen:
		case <-time.After(5 * time.Second):
			t.Fatal("Frist von 5 s abgelaufen: ein Request erreichte den Dienst nicht")
		}
	}
	var vs []*verbindung
	for range abgewiesen {
		v := u.verbinde(t)
		v.schreibe(t, []byte("GET /p HTTP/1.1\r\nHost: h\r\n"))
		vs = append(vs, v)
	}
	warte(t, "alle Verbindungen angenommen", func() bool { return al.n.Load() >= zugelassen+abgewiesen })
	u.srv.Anhalten()
	for _, v := range vs {
		v.schreibe(t, []byte("\r\n"))
		if resp, _ := v.antwort(t, "GET"); resp.StatusCode != 503 {
			t.Errorf("Anfrage nach Anhalten: Status %d, erwartet 503", resp.StatusCode)
		}
	}
	close(frei)
	for i := range antworten {
		if r := erwarte(t, erg[i], "Antwort der zugelassenen Anfrage"); r.status != 200 || r.err != nil {
			t.Errorf("zugelassene Anfrage %d: Status %d, Fehler %v", i, r.status, r.err)
		}
	}
	close(freiAbbruch)
	if err := u.srv.Schliessen(5 * time.Second); err != nil {
		t.Fatalf("Schliessen: %v", err)
	}
	zaehler := map[int]int{}
	for _, s := range buecher(u.tel) {
		if s.ends != 1 || len(s.sets) > 1 {
			t.Errorf("Span %+v: End %d, Set %d; erwartet 1 und höchstens 1", s, s.ends, len(s.sets))
			continue
		}
		status := 0
		if len(s.sets) == 1 {
			status = int(s.sets[0][0].Ganz)
		}
		zaehler[status]++
	}
	gleich(t, "Spans je Status bei der Rückkehr von Schliessen", zaehler, map[int]int{200: antworten, 0: abbrueche, 503: abgewiesen})
	gleich(t, "Zähler", zaehlerFolge(u.tel), []model.Code{model.CodeE5010, model.CodeE5010, model.CodeE5010})
	for i := antworten; i < zugelassen; i++ {
		if r := erwarte(t, erg[i], "Abbruch der zugelassenen Anfrage"); r.err == nil {
			t.Errorf("zugelassene Anfrage %d mit Abbruch: Status %d, erwartet einen Fehler des Clients", i, r.status)
		}
	}
}
