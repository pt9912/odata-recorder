package hybrid_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// schranke hält Aufrufe an, bis der Test sie freigibt; jeder Wartende und jede Erwartung tragen eine harte Frist, so dass
// eine Verletzung als Fehler des Tests endet und nicht als Zeitlimit des Programms.
type schranke struct {
	ankunft chan struct{}
	frei    chan struct{}
	einmal  sync.Once
}

func neueSchranke(n int) *schranke {
	return &schranke{ankunft: make(chan struct{}, n), frei: make(chan struct{})}
}

// warte meldet die Ankunft und hält an, bis die Schranke fällt (nil), ctx endet (der Fehler von ctx) oder die Frist abläuft.
func (s *schranke) warte(ctx context.Context, t *testing.T) error {
	s.ankunft <- struct{}{}
	select {
	case <-s.frei:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(frist):
		t.Errorf("Schranke: Frist abgelaufen")
		return errors.New("Frist der Schranke")
	}
}

// erwarte wartet auf n Ankünfte und meldet, ob sie eintrafen.
func (s *schranke) erwarte(t *testing.T, n int) bool {
	t.Helper()
	for range n {
		select {
		case <-s.ankunft:
		case <-time.After(frist):
			t.Errorf("Schranke: Frist abgelaufen vor der Ankunft von %d Aufrufen", n)
			return false
		}
	}
	return true
}

func (s *schranke) gib() { s.einmal.Do(func() { close(s.frei) }) }

// warteAufEnde wartet auf die Goroutinen der Gruppe, höchstens die Frist.
func warteAufEnde(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	fertig := make(chan struct{})
	go func() {
		wg.Wait()
		close(fertig)
	}()
	select {
	case <-fertig:
	case <-time.After(frist):
		t.Fatalf("die Aufrufe endeten nicht innerhalb der Frist")
	}
}

// anfrageCtx liefert das ctx der Anfrage mit der Kennung id, das der Test abbrechen kann.
func anfrageCtx(id int) (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithValue(context.Background(), anfrageSchluessel{}, id))
}

// spansJeAnfrage liefert je Kennung der Anfrage die Spans als `art=ergebnis` in der Folge ihres Beginns.
func spansJeAnfrage(j *journal) map[int][]string {
	out := map[int][]string{}
	for _, s := range j.spanListe() {
		_, _, e := s.buch()
		out[s.anfrage] = append(out[s.anfrage], s.art()+"="+e)
	}
	return out
}

// Die sechs Szenarien von D16: Kennung der Anfrage modulo 6.
const (
	szReplayAntwort = iota
	szReplayFehler
	szKette
	szUpstreamAntwort
	szUpstreamFehler
	szAbbruch
	szenarien   = 6
	proSzenario = 16
)

func szenarioVon(ctx context.Context) int {
	id, _ := ctx.Value(anfrageSchluessel{}).(int)
	return id % szenarien
}

// gleichzeitigeRunde ruft den Dienst mit 16 Goroutinen je Szenario, die alle in ihrer Stufe offen bleiben, bis der Test sie
// freigibt: so ist jeder Span beim Beginn aller anderen noch offen.
func gleichzeitigeRunde(t *testing.T) {
	t.Helper()
	f := fehlerCodes(t, "ODR-E6001", "ODR-E6003", "ODR-E5003", "ODR-E5001")
	e6001, e6003, e5003, miss := f[0], f[1], f[2], f[3]
	j := neuesJournal()
	g := neuesRigJ(t, polK(t), zielH, j)
	sR, sU := neueSchranke(3*proSzenario), neueSchranke(4*proSzenario)
	g.m.fn = func(_ int, ctx context.Context, _ inbound.Anfrage) (inbound.Antwort, error) {
		if szenarioVon(ctx) < szUpstreamAntwort {
			return inbound.Antwort{}, e6003
		}
		return inbound.Antwort{}, e6001
	}
	g.r.fn = func(_ int, ctx context.Context, _ inbound.Anfrage) (inbound.Antwort, error) {
		if err := sR.warte(ctx, t); err != nil {
			return inbound.Antwort{}, err
		}
		switch szenarioVon(ctx) {
		case szReplayAntwort:
			return inbound.Antwort{Status: 200}, nil
		case szReplayFehler:
			return inbound.Antwort{}, e5003
		}
		return inbound.Antwort{}, miss
	}
	g.u.fn = func(_ int, ctx context.Context, _ outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
		if err := sU.warte(ctx, t); err != nil {
			return outbound.UpstreamAntwort{}, err
		}
		if szenarioVon(ctx) == szUpstreamFehler {
			return outbound.UpstreamAntwort{}, outbound.ErrUpstreamNichtErreichbar
		}
		return outbound.UpstreamAntwort{Status: 200}, nil
	}
	var wg sync.WaitGroup
	cancels := make([]context.CancelFunc, szenarien*proSzenario)
	fehler := make([]error, szenarien*proSzenario)
	for id := range cancels {
		ctx, cancel := anfrageCtx(id)
		cancels[id] = cancel
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, fehler[id] = g.v.Hybrid(ctx, anfrage("GET"))
		}()
	}
	defer sR.gib()
	defer sU.gib()
	if sR.erwarte(t, 3*proSzenario) && sU.erwarte(t, 3*proSzenario) {
		sR.gib()
		if sU.erwarte(t, proSzenario) {
			for id, cancel := range cancels {
				if id%szenarien == szAbbruch {
					cancel()
				}
			}
			sU.gib()
		}
	}
	warteAufEnde(t, &wg)
	for _, cancel := range cancels {
		cancel()
	}
	pruefeGleichzeitig(t, j, fehler)
}

// pruefeGleichzeitig hält das Buch nach einer Runde: je Anfrage ihre Spans mit Art und Ergebnis ihrer Szenarien, je Span ein
// Set und ein End, die Summe der Zähler je Code und der Ausgang je Szenario.
func pruefeGleichzeitig(t *testing.T, j *journal, fehler []error) {
	t.Helper()
	soll := [szenarien][]string{
		szReplayAntwort: {"replay=antwort"}, szReplayFehler: {"replay=fehler"}, szKette: {"replay=fehler", "upstream=antwort"},
		szUpstreamAntwort: {"upstream=antwort"}, szUpstreamFehler: {"upstream=fehler"}, szAbbruch: {"upstream=abgebrochen"},
	}
	je := spansJeAnfrage(j)
	for id := range szenarien * proSzenario {
		if !slices.Equal(je[id], soll[id%szenarien]) {
			t.Errorf("Anfrage %d (Szenario %d): Spans %v, erwartet %v", id, id%szenarien, je[id], soll[id%szenarien])
		}
		var f *inbound.Fehler
		ok := fehler[id] == nil
		switch id % szenarien {
		case szReplayFehler:
			ok = errors.As(fehler[id], &f) && f.Code == model.CodeE5003
		case szUpstreamFehler:
			ok = errors.As(fehler[id], &f) && f.Code == model.CodeE4001
		case szAbbruch:
			ok = errors.Is(fehler[id], context.Canceled)
		}
		if !ok {
			t.Errorf("Anfrage %d (Szenario %d): Fehler %v gehört nicht zum Szenario", id, id%szenarien, fehler[id])
		}
	}
	pruefeBuch(t, "gleichzeitig", j)
	summe := map[string]int{}
	for _, c := range j.gezaehlt() {
		summe[c]++
	}
	if summe["ODR-E6003"] != 3*proSzenario || summe["ODR-E6001"] != 3*proSzenario || summe["ODR-E5001"] != proSzenario || len(summe) != 3 {
		t.Errorf("Zähler je Code %v, erwartet ODR-E6003 48, ODR-E6001 48, ODR-E5001 16", summe)
	}
	if n := len(j.spanListe()); n != 7*proSzenario {
		t.Errorf("%d Spans, erwartet %d", n, 7*proSzenario)
	}
}

// TestGleichzeitigeSpans hält D16 (SPEC-321 Nr. 9 a): 16 Goroutinen mal 6 Szenarien (Replay antwortet, Replay scheitert, Kette
// mit Antwort des Upstream, Upstream antwortet, Upstream scheitert, Abbruch), alle in ihrer Stufe offen, bis der Test sie
// freigibt; je Anfrage trägt jeder Span Art und Ergebnis seiner Stufe, kein Span endet zweimal oder nie, und die Summen der
// Zähler stimmen (48, 48 und 16). Fünf Runden je Prozess. Einen geteilten Zustand ohne Ordnung fängt der Detektor von make test-race, einen geordnet geteilten
// dieser Test nur in einem Teil der Läufe (gemessen 0 von 20 Prozessen bei einem Prozessor); die Zusage „kein Zustand, keine Goroutine“ tragen
// TestQuelltext und TestVerteilerFelder.
func TestGleichzeitigeSpans(t *testing.T) {
	for range 5 {
		gleichzeitigeRunde(t)
	}
}

// TestAbbruchBeendetAllein hält D17 (SPEC-321 Nr. 9 b, 3 f): acht Anfragen warten vor Senden; wird eine abgebrochen, endet allein
// ihr Span mit abgebrochen, die sieben anderen sind noch offen (kein Set, kein End) und enden nach der Freigabe mit antwort.
func TestAbbruchBeendetAllein(t *testing.T) {
	const n, abbruch = 8, 3
	j := neuesJournal()
	g := neuesRigJ(t, polU(t), zielH, j)
	s, e6001 := neueSchranke(n), fehlerCodes(t, "ODR-E6001")[0]
	g.m.fn = func(int, context.Context, inbound.Anfrage) (inbound.Antwort, error) {
		return inbound.Antwort{}, e6001
	}
	g.u.fn = func(_ int, ctx context.Context, _ outbound.UpstreamAnfrage) (outbound.UpstreamAntwort, error) {
		if err := s.warte(ctx, t); err != nil {
			return outbound.UpstreamAntwort{}, err
		}
		return outbound.UpstreamAntwort{Status: 200}, nil
	}
	var wg sync.WaitGroup
	cancels, fertig, fehler := make([]context.CancelFunc, n), make([]chan struct{}, n), make([]error, n)
	for id := range n {
		ctx, cancel := anfrageCtx(id)
		cancels[id], fertig[id] = cancel, make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(fertig[id])
			_, fehler[id] = g.v.Hybrid(ctx, anfrage("GET"))
		}()
	}
	defer s.gib()
	if s.erwarte(t, n) {
		cancels[abbruch]()
		select {
		case <-fertig[abbruch]:
		case <-time.After(frist):
			t.Errorf("die abgebrochene Anfrage endete nicht innerhalb der Frist")
		}
		for _, sp := range j.spanListe() {
			sets, ends, e := sp.buch()
			switch {
			case sp.anfrage == abbruch && (sets != 1 || ends != 1 || e != "abgebrochen"):
				t.Errorf("abgebrochene Anfrage: %d Set, %d End, Ergebnis %q; erwartet 1, 1 und abgebrochen", sets, ends, e)
			case sp.anfrage != abbruch && (sets != 0 || ends != 0):
				t.Errorf("Anfrage %d: %d Set, %d End nach dem Abbruch einer anderen; erwartet 0 und 0 (der Span ist noch offen)", sp.anfrage, sets, ends)
			}
		}
		s.gib()
	}
	warteAufEnde(t, &wg)
	je := spansJeAnfrage(j)
	for id := range n {
		want := "upstream=antwort"
		if id == abbruch {
			want = "upstream=abgebrochen"
		}
		if !slices.Equal(je[id], []string{want}) {
			t.Errorf("Anfrage %d: Spans %v, erwartet [%s]", id, je[id], want)
		}
		if (id == abbruch) != errors.Is(fehler[id], context.Canceled) {
			t.Errorf("Anfrage %d: Fehler %v", id, fehler[id])
		}
	}
	pruefeBuch(t, "Abbruch", j)
	if len(j.spanListe()) != n {
		t.Errorf("%d Spans, erwartet %d", len(j.spanListe()), n)
	}
}
