package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sync"
	"testing"
	"time"
)

func gz(id string, mikros int64) spanZeile {
	return spanZeile{Name: spanGesamt, SpanID: id, Modus: "replay", Mikros: mikros}
}

func kz(parent string, mikros int64) spanZeile {
	return spanZeile{Name: spanReplay, SpanID: "k-" + parent, ParentID: parent, Mikros: mikros}
}

func sammle(betriebsart string, zeilen ...spanZeile) *sammler {
	s := neuerSammler(betriebsart)
	for _, z := range zeilen {
		s.nimm(z)
	}
	return s
}

// TestZuordnung: das Kind trägt parent_id gleich der span_id des Gesamtspans und steht vor ihm; ein Gesamtspan ohne Kind zählt ohne_kind, ein
// Kind ohne Eltern kind_ohne_eltern, ein Kind hinter dem Eltern ist beides; ein Span eines anderen Namens ist kein Kind (alle fünf anderen Namen
// in beiden Betriebsarten). 17 Fälle (die 5 Formen der Tabelle, „Kind ohne Eltern“ in drei Fällen, dazu 10 Namen), Erwartung je Fall als Literal; die Messphase
// ist so gewählt, dass eine Zuordnung nach der Reihenfolge eine andere Dauer des Kindes liefert.
func TestZuordnung(t *testing.T) {
	wache(t, 30*time.Second)
	type erwartet struct{ ohneKind, kindOhneEltern, kind, kindP95, anteil int }
	formen := []struct {
		name   string
		zeilen []spanZeile
		w, n   int
		want   erwartet
	}{
		{"regulär", []spanZeile{kz("a", 300), gz("a", 1000)}, 0, 1, erwartet{0, 0, 1, 300, 300}},
		{"Eltern ohne Kind", []spanZeile{gz("a", 1000)}, 0, 1, erwartet{1, 0, 0, 0, 0}},
		{"Kind ohne Eltern", []spanZeile{kz("a", 300)}, 0, 0, erwartet{0, 1, 0, 0, 0}},
		{"Kind ohne parent_id", []spanZeile{{Name: spanReplay, SpanID: "k", Mikros: 300}}, 0, 0, erwartet{0, 1, 0, 0, 0}},
		{"Kind ohne parent_id, Eltern mit leerer span_id", []spanZeile{{Name: spanReplay, SpanID: "k", Mikros: 300}, gz("", 1000)}, 0, 1, erwartet{1, 1, 0, 0, 0}},
		{"verschränkt", []spanZeile{kz("a", 300), kz("b", 700), gz("b", 2000), gz("a", 1000)}, 1, 1, erwartet{0, 0, 1, 300, 300}},
		{"Kind hinter dem Eltern", []spanZeile{gz("a", 1000), kz("a", 300)}, 0, 1, erwartet{1, 1, 0, 0, 0}},
	}
	for _, f := range formen {
		m, _ := sammle("replay", f.zeilen...).auswerte(f.w, f.n)
		if got := (erwartet{m.OhneKind, m.KindOhneEltern, m.Kind, int(m.KindP95), m.KindAnteil}); got != f.want {
			t.Errorf("%s: %+v, erwartet %+v", f.name, got, f.want)
		}
	}
	namen := []string{"odata.record.capture", spanReplay, spanMock, "odata.workspace.mutation", "odata.hybrid.fallback", "x"}
	n := len(formen)
	for _, betriebsart := range []string{"replay", "mock"} {
		for _, name := range namen {
			if name == kindName(betriebsart) {
				continue
			}
			kind := spanZeile{Name: name, SpanID: "k", ParentID: "a", Mikros: 300}
			m, _ := sammle(betriebsart, kind, spanZeile{Name: spanGesamt, SpanID: "a", Modus: betriebsart, Mikros: 1000}).auswerte(0, 1)
			if m.Kind != 0 || m.OhneKind != 1 || m.KindOhneEltern != 0 {
				t.Errorf("%s als Kind in %s: Kind %d, ohne_kind %d, kind_ohne_eltern %d; erwartet 0, 1 und 0", name, betriebsart, m.Kind, m.OhneKind, m.KindOhneEltern)
			}
			n++
		}
	}
	if n != 7+10 {
		t.Errorf("%d Fälle, erwartet 7 Formen und 10 Namen", n)
	}
	if kindName("beide") != "" || kindName("") != "" {
		t.Error("eine Betriebsart außer replay und mock hat kein Kind")
	}
}

// reihe liefert die Spans 1 bis anzahl: das Warm-up (1 bis 20) mit 5.000.000 µs, die Messphase (21 bis 100) mit den Dauern 1000 bis 80.000 µs
// in einer Folge ohne Ordnung (das Maximum steht am Span 77) und jeden weiteren Span mit 9.000.000 µs; Modus ist je Span der Wert modus(i).
func reihe(anzahl int, modus func(int) string) *sammler {
	s := neuerSammler("replay")
	for i := 1; i <= anzahl; i++ {
		d := int64(5000000)
		switch {
		case i > 100:
			d = 9000000
		case i > 20:
			d = int64((i-20)*7%80+1) * 1000
		}
		z := gz("g", d)
		z.Modus = modus(i)
		s.nimm(z)
	}
	return s
}

func richtig(int) string { return "replay" }

// TestPhasen: die ersten w Spans sind Warm-up, die nächsten n die Messphase (Ränder 20 und 21 mit 5.000.000 und 1000 µs, 100 und 101 mit
// 80.000 und 9.000.000 µs, die p95 und Maximum verschöben); 99 und 101 Spans sind keine passende Zählung, n gleich 0 auch; ein falscher
// recorder.mode an der ersten, einer mittleren und der letzten Zeile sowie ein fehlender sind modus; Gutfall; der Nachlauf endet mit der
// Zahl, spätestens nach 10 s der Uhr.
func TestPhasen(t *testing.T) {
	wache(t, 30*time.Second)
	for _, anzahl := range []int{99, 100, 101} {
		m, v := reihe(anzahl, richtig).auswerte(20, 80)
		if anzahl == 99 {
			if v != (verstoesse{Zaehlung: true}) || m.Messphase != 79 {
				t.Errorf("99 Spans: Verstöße %+v, Messphase %d; erwartet Zählung und 79", v, m.Messphase)
			}
			continue
		}
		if m.Messphase != 80 || m.P95 != 76000 || m.Max != 80000 || m.Spans != anzahl || v.Zaehlung != (anzahl == 101) || v.Modus {
			t.Errorf("%d Spans: %+v %+v; erwartet Messphase 80, p95 76000, Maximum 80000, Zählung nur bei 101", anzahl, m, v)
		}
	}
	if _, v := reihe(20, richtig).auswerte(20, 0); !v.Zaehlung {
		t.Error("n gleich 0 ist ungültig")
	}
	for _, falsch := range []int{1, 50, 100} {
		if _, v := reihe(100, func(i int) string {
			if i == falsch {
				return "mock"
			}
			return "replay"
		}).auswerte(20, 80); !v.Modus || v.Zaehlung {
			t.Errorf("recorder.mode falsch an Span %d: %+v", falsch, v)
		}
	}
	if _, v := reihe(100, func(i int) string {
		if i == 77 {
			return ""
		}
		return "replay"
	}).auswerte(20, 80); !v.Modus {
		t.Error("recorder.mode fehlt an einer Zeile: erwartet modus")
	}
	if nachlaufFrist != 10*time.Second {
		t.Errorf("Nachlauf %v, erwartet 10 s", nachlaufFrist)
	}
	nachlauf(t, 99, true)
	nachlauf(t, 100, false)
}

// nachlauf lässt warteAufSpans auf 100 Gesamtspans warten, wenn schon anfangs gelesen sind: mit 99 endet es erst mit dem Vorstellen der
// Uhr um 10 s; mit 100 sofort, ohne dass die Uhr vorgestellt wird.
func nachlauf(t *testing.T, anfangs int, ueberUhr bool) {
	t.Helper()
	uhr := neueFakeUhr()
	s := reihe(anfangs, richtig)
	fertig := make(chan struct{})
	go func() {
		s.warteAufSpans(context.Background(), uhr, 100)
		close(fertig)
	}()
	if !ueberUhr {
		warteAuf(t, fertig, time.Second)
		return
	}
	for uhr.wartende() < 1 {
		runtime.Gosched()
	}
	uhr.Vorstellen(nachlaufFrist)
	warteAuf(t, fertig, time.Second)
	if s.anzahl() != 99 {
		t.Errorf("%d Spans gelesen, erwartet 99", s.anzahl())
	}
}

// TestNachlaufZahl: der Nachlauf endet mit dem hundertsten Span, auch wenn die Uhr nicht vorgestellt wird, und mit dem Kontext.
func TestNachlaufZahl(t *testing.T) {
	wache(t, 30*time.Second)
	uhr := neueFakeUhr()
	s := reihe(99, richtig)
	fertig := make(chan struct{})
	go func() {
		s.warteAufSpans(context.Background(), uhr, 100)
		close(fertig)
	}()
	for uhr.wartende() < 1 {
		runtime.Gosched()
	}
	s.nimm(gz("g", 1))
	warteAuf(t, fertig, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ende := make(chan struct{})
	go func() {
		reihe(1, richtig).warteAufSpans(ctx, uhr, 100)
		close(ende)
	}()
	warteAuf(t, ende, time.Second)
}

// TestSammlerGleichzeitig (nebenläufig, 20 Läufe): vier Goroutinen ordnen je 500 Gesamtspans mit ihrem Kind ein, während eine fünfte die Zahl der
// Gesamtspans liest und die Messphase auswertet; danach sind alle 2000 mit ihrem Kind gezählt, ohne ohne_kind und kind_ohne_eltern, und die Gültigkeit hat keinen Verstoß.
func TestSammlerGleichzeitig(t *testing.T) {
	wache(t, 60*time.Second)
	for lauf := range 20 {
		s := neuerSammler("replay")
		var schreiber, leser sync.WaitGroup
		stop := make(chan struct{})
		leser.Add(1)
		go func() {
			defer leser.Done()
			for {
				select {
				case <-stop:
					return
				default:
					s.anzahl()
					s.auswerte(0, 50)
				}
			}
		}()
		for g := range 4 {
			schreiber.Add(1)
			go func() {
				defer schreiber.Done()
				for i := range 500 {
					id := fmt.Sprintf("g%d-%d", g, i)
					s.nimm(kz(id, 100))
					s.nimm(gz(id, 1000))
				}
			}()
		}
		schreiber.Wait()
		close(stop)
		leser.Wait()
		m, v := s.auswerte(0, 2000)
		if m.Spans != 2000 || m.Kind != 2000 || m.OhneKind != 0 || m.KindOhneEltern != 0 || v != (verstoesse{}) {
			t.Errorf("Lauf %d: %+v %+v; erwartet 2000 Gesamtspans mit je einem Kind und keinen Verstoß", lauf, m, v)
		}
	}
}

// TestP95: Rang ⌈95·n÷100⌉ ganzzahlig, ohne Interpolation: Schleife über n = 1 bis 1000 und 2000 bis 12000 in Tausenderschritten (1011 Werte,
// darunter die Messphasen 200 und 12000) und 4 Folgen (aufsteigend, absteigend, alle gleich, ein Ausreißer), 4044 Elemente, gegen ein Orakel,
// das den kleinsten Rang r mit 100·r ≥ 95·n sucht; die Eingabe bleibt unverändert; Literale für n = 20, 21, 100, 101 und 12000; n gleich 0
// ist ein Fehler.
func TestP95(t *testing.T) {
	wache(t, 30*time.Second)
	folgen := map[string]func(n, i int) int64{
		"aufsteigend": func(_, i int) int64 { return int64(i+1) * 1000 },
		"absteigend":  func(n, i int) int64 { return int64(n-i) * 1000 },
		"gleich":      func(int, int) int64 { return 7 },
		"Ausreißer": func(n, i int) int64 {
			if i == n/2 {
				return 1000000
			}
			return 10
		},
	}
	var ns []int
	for n := 1; n <= 1000; n++ {
		ns = append(ns, n)
	}
	for n := 2000; n <= 12000; n += 1000 {
		ns = append(ns, n)
	}
	elemente := 0
	for name, f := range folgen {
		for _, n := range ns {
			werte := make([]int64, n)
			for i := range werte {
				werte[i] = f(n, i)
			}
			vorher := slices.Clone(werte)
			sortiert := slices.Clone(werte)
			slices.Sort(sortiert)
			rang := 1
			for 100*rang < 95*n {
				rang++
			}
			if got, err := p95(werte); err != nil || got != sortiert[rang-1] || !slices.Equal(werte, vorher) {
				t.Errorf("%s n=%d: %d, %v, Eingabe %v; erwartet %d und die Eingabe unverändert", name, n, got, err, slices.Equal(werte, vorher), sortiert[rang-1])
			}
			elemente++
		}
	}
	if elemente != 4*1011 {
		t.Errorf("%d Elemente, erwartet 4044", elemente)
	}
	for n, rang := range map[int]int64{20: 19, 21: 20, 100: 95, 101: 96, 12000: 11400} {
		werte := make([]int64, n)
		for i := range werte {
			werte[i] = int64(i+1) * 1000
		}
		if got, err := p95(werte); err != nil || got != rang*1000 {
			t.Errorf("n=%d: %d, %v; erwartet der %d. Wert", n, got, err, rang)
		}
	}
	if _, err := p95(nil); !errors.Is(err, errKeineWerte) {
		t.Errorf("n gleich 0: Fehler %v", err)
	}
}

// TestAnteil: Summe der Kinder durch Summe der Gesamtspans mit Kind in Zehntelprozent abgeschnitten: 8000 von 10000 gleich 800, 1 von 3 gleich
// 333, 2 von 3 gleich 666; ein Gesamtspan ohne Kind zählt nicht in der Summe; kind_p95 aus 5 Kindern gleich dem 5. Wert.
func TestAnteil(t *testing.T) {
	wache(t, 30*time.Second)
	for _, f := range []struct {
		kind, gesamt int64
		want         int
	}{{8000, 10000, 800}, {1, 3, 333}, {2, 3, 666}, {0, 0, 0}, {0, 5, 0}, {5, 5, 1000}} {
		if got := anteil(f.kind, f.gesamt); got != f.want {
			t.Errorf("anteil(%d, %d) = %d, erwartet %d", f.kind, f.gesamt, got, f.want)
		}
	}
	m, _ := sammle("replay", kz("a", 500), gz("a", 1000), gz("b", 9000)).auswerte(0, 2)
	if m.KindAnteil != 500 || m.Kind != 1 || m.OhneKind != 1 {
		t.Errorf("ein Gesamtspan mit Kind (1000 µs, 500 µs) und einer ohne (9000 µs): Anteil %d, Kind %d, ohne_kind %d; erwartet 500, 1 und 1", m.KindAnteil, m.Kind, m.OhneKind)
	}
	var zeilen []spanZeile
	for i, k := range []int64{500, 100, 300, 200, 400} {
		id := string(rune('a' + i))
		zeilen = append(zeilen, kz(id, k), gz(id, 1000))
	}
	if m, _ := sammle("replay", zeilen...).auswerte(0, 5); m.KindP95 != 500 {
		t.Errorf("kind_p95 aus 5 Kindern: %d, erwartet 500", m.KindP95)
	}
}

// TestGueltigkeit: alle 32 Teilmengen der fünf Verstöße; gültig genau ohne Verstoß, grund der erste in der Folge fehler, zaehlung, modus, last,
// binary, gegen die Tabelle als Literal (Bit 0 gleich fehler bis Bit 4 gleich binary).
func TestGueltigkeit(t *testing.T) {
	wache(t, 30*time.Second)
	gruende := []string{
		"ok", "fehler", "zaehlung", "fehler", "modus", "fehler", "zaehlung", "fehler", "last", "fehler", "zaehlung", "fehler", "modus", "fehler", "zaehlung", "fehler",
		"binary", "fehler", "zaehlung", "fehler", "modus", "fehler", "zaehlung", "fehler", "last", "fehler", "zaehlung", "fehler", "modus", "fehler", "zaehlung", "fehler",
	}
	if len(gruende) != 32 {
		t.Fatalf("%d Gründe, erwartet 32", len(gruende))
	}
	for maske := range 32 {
		v := verstoesse{Fehler: maske&1 != 0, Zaehlung: maske&2 != 0, Modus: maske&4 != 0, Last: maske&8 != 0, Binary: maske&16 != 0}
		if gueltig, grund := gueltigkeit(v); gueltig != (maske == 0) || grund != gruende[maske] {
			t.Errorf("%+v: gültig %v, grund %s; erwartet %v und %s", v, gueltig, grund, maske == 0, gruende[maske])
		}
	}
}
