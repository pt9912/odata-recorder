package main

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"
)

// Namen der Spans, die der Lasttest auswertet (SPEC-216 Nr. 10 c); jeder andere Name wird nicht ausgewertet.
const (
	spanGesamt = "odata.gateway.request"
	spanReplay = "odata.replay.match"
	spanMock   = "odata.mock.query"
)

// Die Tokens des Schlüssels grund (SPEC-216 Nr. 13); grundAufbau und grundFrist stehen in aufbau.go.
const (
	grundOK          = "ok"
	grundFehler      = "fehler"
	grundZaehlung    = "zaehlung"
	grundModus       = "modus"
	grundLast        = "last"
	grundBinary      = "binary"
	grundStreuung    = "streuung"
	grundProfil      = "profil"
	grundZiel        = "ziel"
	grundBetriebsart = "betriebsart"
	grundErgebnis    = "ergebnis"
)

// nachlaufFrist ist die Zeit, die das Programm nach der letzten Anfrage auf die Zeilen der Gesamtspans wartet (SPEC-216 Nr. 10 e).
const nachlaufFrist = 10 * time.Second

// errKeineWerte ist der Fehler von p95 ohne Wert (SPEC-216 Nr. 10 f).
var errKeineWerte = errors.New("kein Wert")

// kindName ist der Name des Kindes des Gesamtspans einer Betriebsart; für eine andere Betriebsart gibt es keines.
func kindName(betriebsart string) string {
	switch betriebsart {
	case "replay":
		return spanReplay
	case "mock":
		return spanMock
	}
	return ""
}

// p95 ist der Wert am Rang ⌈95·n÷100⌉ der aufsteigend sortierten Werte (SPEC-216 Nr. 10 f), ganzzahlig und ohne Interpolation.
func p95(werte []int64) (int64, error) {
	if len(werte) == 0 {
		return 0, errKeineWerte
	}
	s := slices.Clone(werte)
	slices.Sort(s)
	return s[(95*len(s)+99)/100-1], nil
}

// anteil ist die Summe der Kinder geteilt durch die Summe der Gesamtspans in Zehntelprozent, abgeschnitten (SPEC-216 Nr. 10 g); ohne
// Gesamtspans 0.
func anteil(kinder, gesamt int64) int {
	if gesamt == 0 {
		return 0
	}
	return int(kinder * 1000 / gesamt)
}

// verstoesse sind die fünf Verstöße der Gültigkeit eines Laufs (SPEC-216 Nr. 10 h).
type verstoesse struct {
	Fehler, Zaehlung, Modus, Last, Binary bool
}

// gueltigkeit liefert, ob der Lauf gültig ist, und den ersten Verstoß in der Folge fehler, zaehlung, modus, last, binary, sonst ok.
func gueltigkeit(v verstoesse) (bool, string) {
	switch {
	case v.Fehler:
		return false, grundFehler
	case v.Zaehlung:
		return false, grundZaehlung
	case v.Modus:
		return false, grundModus
	case v.Last:
		return false, grundLast
	case v.Binary:
		return false, grundBinary
	}
	return true, grundOK
}

// Messung ist die Zeile `Messung` eines Laufs (SPEC-216 Nr. 13); Dauern stehen in ganzen Mikrosekunden, der Anteil in Zehntelprozent.
// Anfragen, Fehler, FremdeZeilen, die Werte der Last und Gueltig und Grund setzt der Aufrufer.
type Messung struct {
	Betriebsart     string
	Lauf            int
	Anfragen        int
	Spans           int
	Messphase       int
	Fehler          int
	FremdeZeilen    int
	P95             int64
	Max             int64
	Kind            int
	KindP95         int64
	KindAnteil      int
	OhneKind        int
	KindOhneEltern  int
	GleichzeitigMax int
	VerspaetungMax  int64
	Bytes           int64
	TreiberCPU      time.Duration
	Gueltig         bool
	Grund           string
}

// gesamtSpan ist ein gelesener Gesamtspan mit der Dauer seines Kindes, falls es eines hat.
type gesamtSpan struct {
	Mikros     int64
	KindMikros int64
	HatKind    bool
}

// sammler ordnet die Span-Zeilen eines Laufs: das Kind endet vor dem Eltern und wartet, bis dessen Zeile kommt (SPEC-216 Nr. 10 c). Er ist
// für gleichzeitigen Aufruf sicher.
type sammler struct {
	betriebsart string
	kind        string
	neu         chan struct{}

	mu             sync.Mutex
	gesamt         []gesamtSpan
	wartend        map[string]int64
	kindOhneEltern int
	ohneKind       int
	modusFalsch    bool
}

func neuerSammler(betriebsart string) *sammler {
	return &sammler{betriebsart: betriebsart, kind: kindName(betriebsart), neu: make(chan struct{}, 1), wartend: map[string]int64{}}
}

// nimm ordnet eine Zeile ein; jeder Gesamtspan wird auf recorder.mode gegen die Betriebsart geprüft (SPEC-216 Nr. 10 d).
func (s *sammler) nimm(z spanZeile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch z.Name {
	case spanGesamt:
		g := gesamtSpan{Mikros: z.Mikros}
		if k, ok := s.wartend[z.SpanID]; ok {
			g.HatKind, g.KindMikros = true, k
			delete(s.wartend, z.SpanID)
		} else {
			s.ohneKind++
		}
		if z.Modus != s.betriebsart {
			s.modusFalsch = true
		}
		s.gesamt = append(s.gesamt, g)
		select {
		case s.neu <- struct{}{}:
		default:
		}
	case s.kind:
		if z.ParentID == "" {
			s.kindOhneEltern++
		} else {
			s.wartend[z.ParentID] = z.Mikros
		}
	}
}

// anzahl ist die Zahl der gelesenen Gesamtspans.
func (s *sammler) anzahl() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.gesamt)
}

// warteAufSpans wartet, bis erwartet Gesamtspans gelesen sind, höchstens nachlaufFrist der Uhr (SPEC-216 Nr. 10 e) oder bis ctx endet, und
// liefert, ob die Zahl erreicht ist.
func (s *sammler) warteAufSpans(ctx context.Context, uhr Uhr, erwartet int) bool {
	ctx, cancel := mitFrist(ctx, uhr, nachlaufFrist)
	defer cancel()
	for s.anzahl() < erwartet {
		select {
		case <-s.neu:
		case <-ctx.Done():
			return s.anzahl() >= erwartet
		}
	}
	return true
}

// auswerte teilt die Gesamtspans nach der Reihenfolge in die ersten w (Warm-up) und die nächsten n (Messphase) und rechnet p95, Maximum und
// den Anteil des Kindes über die Messphase (SPEC-216 Nr. 10 e bis g). Die Zählung stimmt, wenn genau w+n Gesamtspans gelesen sind und die
// Messphase mindestens einen enthält; Modus zeigt eine falsche Betriebsart an irgendeiner Zeile eines Gesamtspans an.
func (s *sammler) auswerte(w, n int) (Messung, verstoesse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	von, bis := min(w, len(s.gesamt)), min(w+n, len(s.gesamt))
	var dauern, kinder []int64
	var summeGesamt, summeKind int64
	m := Messung{Spans: len(s.gesamt), OhneKind: s.ohneKind, KindOhneEltern: s.kindOhneEltern + len(s.wartend)}
	for _, g := range s.gesamt[von:bis] {
		dauern = append(dauern, g.Mikros)
		m.Max = max(m.Max, g.Mikros)
		if g.HatKind {
			kinder = append(kinder, g.KindMikros)
			summeGesamt += g.Mikros
			summeKind += g.KindMikros
		}
	}
	m.Messphase, m.Kind, m.KindAnteil = len(dauern), len(kinder), anteil(summeKind, summeGesamt)
	m.KindP95, _ = p95(kinder)
	var err error
	m.P95, err = p95(dauern)
	return m, verstoesse{Zaehlung: len(s.gesamt) != w+n || err != nil, Modus: s.modusFalsch}
}
