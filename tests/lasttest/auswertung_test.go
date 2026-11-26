package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
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
// Kind ohne Eltern kind_ohne_eltern, ein Kind hinter dem Eltern ist beides; ein Span eines anderen Namens ist kein Kind (je Betriebsart sechs andere
// Namen, darunter der leere). 21 Fälle: 9 Formen (fünf der Tabelle, die sechste mit falschem Namen führen die Namen; „Kind ohne Eltern“ in zwei
// weiteren; ein Gesamtspan ohne Kind im Warm-up; zwei hinter w + n) und 12 Namen, Erwartung je Fall als Literal, dazu eine Zeile des Kindes mit leerer parent_id über den Leser; die Messphase
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
		{"Gesamtspan ohne Kind im Warm-up", []spanZeile{gz("a", 1000), kz("b", 300), gz("b", 2000)}, 1, 1, erwartet{1, 0, 1, 300, 150}},
		{"Gesamtspans ohne Kind hinter w + n", []spanZeile{kz("a", 300), gz("a", 1000), gz("b", 2000), gz("c", 2000)}, 0, 1, erwartet{2, 0, 1, 300, 300}},
	}
	for _, f := range formen {
		m, _ := sammle("replay", f.zeilen...).auswerte(f.w, f.n)
		if got := (erwartet{m.OhneKind, m.KindOhneEltern, m.Kind, int(m.KindP95), m.KindAnteil}); got != f.want {
			t.Errorf("%s: %+v, erwartet %+v", f.name, got, f.want)
		}
	}
	kindZeile := `{"ereignis":"span","msg":"odata.replay.match","span_id":"k","parent_id":"","duration_ms":0.300}`
	gesamtZeile := `{"ereignis":"span","msg":"odata.gateway.request","span_id":"","duration_ms":1.000,"recorder.mode":"replay"}`
	s := neuerSammler("replay")
	if fremd, err := lies(strings.NewReader(kindZeile+"\n"+gesamtZeile+"\n"), s.nimm); err != nil || fremd != 0 {
		t.Fatalf("Zeilen mit leerer parent_id und leerer span_id: fremd %d, Fehler %v", fremd, err)
	}
	if m, _ := s.auswerte(0, 1); m.KindOhneEltern != 1 || m.OhneKind != 1 || m.Kind != 0 {
		t.Errorf("Kind mit leerer parent_id über den Leser: kind_ohne_eltern %d, ohne_kind %d, kind %d; erwartet 1, 1 und 0", m.KindOhneEltern, m.OhneKind, m.Kind)
	}
	namen := []string{"odata.record.capture", spanReplay, spanMock, "odata.workspace.mutation", "odata.hybrid.fallback", "x", ""}
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
	if n != 9+12 {
		t.Errorf("%d Fälle, erwartet 9 Formen und 12 Namen", n)
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

// TestPhasen: die ersten w Spans sind Warm-up, die nächsten n die Messphase (Ränder 20 und 21 mit 5.000.000 und 8000 µs, 100 und 101 mit
// 1000 und 9.000.000 µs, die p95 und Maximum verschöben); 99 und 101 Spans sind keine passende Zählung, n gleich 0 auch; ein falscher
// recorder.mode an der ersten, einer mittleren und der letzten Zeile sowie ein fehlender sind modus; Gutfall; eine Messphase aus lauter
// Dauern 0.000 gibt p95 und Maximum 0 und bleibt gültig; die Frist des Nachlaufs ist 10 s (den Aufruf führt TestNachlaufWarten).
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
	s := neuerSammler("replay")
	nullen := `{"ereignis":"span","msg":"odata.gateway.request","span_id":"g","duration_ms":0.000,"recorder.mode":"replay"}` + "\n"
	if fremd, err := lies(strings.NewReader(strings.Repeat(nullen, 7)), s.nimm); err != nil || fremd != 0 {
		t.Fatalf("Zeilen mit der Dauer 0.000: fremd %d, Fehler %v", fremd, err)
	}
	if m, v := s.auswerte(2, 5); m.Messphase != 5 || m.P95 != 0 || m.Max != 0 || v != (verstoesse{}) {
		t.Errorf("Messphase aus Dauern 0.000: %+v %+v; erwartet 5 Spans, p95 und Maximum 0 und einen gültigen Lauf", m, v)
	}
	if nachlaufFrist != 10*time.Second {
		t.Errorf("Nachlauf %v, erwartet 10 s", nachlaufFrist)
	}
}

// wechsel ist die Zahl der Wechsel zwischen Goroutinen, die ein Test abwartet; er zählt keine Zeit.
const wechsel = 1000000

// warteBis wartet mit Wechseln der Goroutinen, höchstens wechsel Mal, bis f wahr ist; sonst endet der Test mit der Aussage was.
func warteBis(t *testing.T, was string, f func() bool) {
	t.Helper()
	for i := 0; !f(); i++ {
		if i == wechsel {
			t.Fatalf("%s: nicht eingetreten", was)
		}
		runtime.Gosched()
	}
}

// bleibtOffen verlangt, dass ch auch nach 2000 Wechseln der Goroutinen keinen Wert trägt; eine Rückkehr in dieser Zeit färbt den Test.
func bleibtOffen[T any](t *testing.T, ch <-chan T, was string) {
	t.Helper()
	for range 2000 {
		select {
		case <-ch:
			t.Fatalf("%s: der Aufruf ist zurückgekehrt", was)
		default:
			runtime.Gosched()
		}
	}
}

// fristen liefert die Zeitpunkte, auf die die Wartenden der Uhr des Tests warten.
func fristen(u *fakeUhr) []time.Time {
	u.mu.Lock()
	defer u.mu.Unlock()
	var bis []time.Time
	for _, w := range u.warter {
		bis = append(bis, w.bis)
	}
	return bis
}

// wartender startet warteAufSpans auf 100 Gesamtspans mit bisher 100 − fehlend gelesenen und meldet, sobald die Uhr genau einen Wartenden
// trägt, dessen Zeitpunkt 10 s nach dem Aufruf liegt.
func wartender(ctx context.Context, t *testing.T, uhr *fakeUhr, fehlend int) (*sammler, <-chan bool) {
	t.Helper()
	s := reihe(100-fehlend, richtig)
	want := uhr.Jetzt().Add(10 * time.Second)
	fertig := make(chan bool, 1)
	go func() { fertig <- s.warteAufSpans(ctx, uhr, 100) }()
	warteBis(t, "Wartender der Frist", func() bool { return uhr.wartende() == 1 })
	if f := fristen(uhr); len(f) != 1 || !f[0].Equal(want) {
		t.Fatalf("Zeitpunkte der Wartenden %v, erwartet einer 10 s nach dem Aufruf", f)
	}
	bleibtOffen(t, fertig, "vor dem ersten Span")
	return s, fertig
}

// nachlaufRaender führt die Fälle neben der Zahl der fehlenden Spans: kein fehlender Span und mehr als erwartet (je mit aktivem und beendetem
// Kontext; sofort wahr, ohne dass die Uhr vorgestellt wird), mehr Spans als erwartet während des Wartens, ein Teil der Spans und dann die Frist
// (falsch), die Frist ab dem Aufruf, wenn Spans verteilt bei 3, 6 und 9 s eintreffen, und das Ende mit dem letzten Span
// (nachlaufEndeMitLetztemSpan). Es liefert die Zahl der Fälle.
func nachlaufRaender(t *testing.T) int {
	t.Helper()
	n := 0
	beendet, cancel := context.WithCancel(context.Background())
	cancel()
	for _, anfangs := range []int{100, 101, 105} {
		for name, ctx := range map[string]context.Context{"aktiv": context.Background(), "beendet": beendet} {
			fertig := make(chan bool, 1)
			s, uhr := reihe(anfangs, richtig), neueFakeUhr()
			go func() { fertig <- s.warteAufSpans(ctx, uhr, 100) }()
			if !warteAuf(t, fertig, time.Second) {
				t.Errorf("%d gelesen, Kontext %s: Ergebnis falsch, erwartet wahr", anfangs, name)
			}
			n++
		}
	}
	for _, fehlend := range []int{1, 5} {
		s, fertig := wartender(context.Background(), t, neueFakeUhr(), fehlend)
		for range fehlend + 2 {
			s.nimm(gz("g", 1))
		}
		if !warteAuf(t, fertig, time.Second) {
			t.Errorf("%d fehlend, 2 mehr als erwartet: Ergebnis falsch", fehlend)
		}
		n++
	}
	for _, f := range []struct{ fehlend, kommen int }{{2, 1}, {5, 1}, {5, 4}, {100, 1}, {100, 99}} {
		uhr := neueFakeUhr()
		s, fertig := wartender(context.Background(), t, uhr, f.fehlend)
		for range f.kommen {
			s.nimm(gz("g", 1))
		}
		warteBis(t, "Signal gelesen", func() bool { return len(s.neu) == 0 })
		bleibtOffen(t, fertig, "nach einem Teil der Spans")
		uhr.Vorstellen(10 * time.Second)
		if got := warteAuf(t, fertig, time.Second); got || s.anzahl() != 100-f.fehlend+f.kommen {
			t.Errorf("%d fehlend, %d kommen: Ergebnis %v, %d Spans; erwartet falsch und %d", f.fehlend, f.kommen, got, s.anzahl(), 100-f.fehlend+f.kommen)
		}
		n++
	}
	for _, fehlend := range []int{2, 5, 100} {
		uhr := neueFakeUhr()
		want := uhr.Jetzt().Add(10 * time.Second)
		s, fertig := wartender(context.Background(), t, uhr, fehlend)
		vergangen := time.Duration(0)
		for range min(3, fehlend-1) {
			uhr.Vorstellen(3 * time.Second)
			vergangen += 3 * time.Second
			s.nimm(gz("g", 1))
			warteBis(t, "Signal gelesen", func() bool { return len(s.neu) == 0 })
		}
		bleibtOffen(t, fertig, "vor Ablauf der Frist ab dem Aufruf")
		if f := fristen(uhr); len(f) != 1 || !f[0].Equal(want) {
			t.Errorf("%d fehlend: Zeitpunkte der Wartenden %v, erwartet einer 10 s nach dem Aufruf", fehlend, f)
		}
		uhr.Vorstellen(10*time.Second - vergangen)
		if warteAuf(t, fertig, time.Second) {
			t.Errorf("%d fehlend, Spans verteilt: Ergebnis wahr, erwartet falsch", fehlend)
		}
		n++
	}
	return n + nachlaufEndeMitLetztemSpan(t)
}

// nachlaufEndeMitLetztemSpan: liegt der letzte Span vor, wenn der Kontext oder die Frist endet, liefert der Aufruf wahr. Der Test hält die Sperre
// des Sammlers, beendet den Aufruf und legt den Span ohne Signal ab; der Aufruf liest die Zahl danach. Es liefert die Zahl der Fälle (2).
func nachlaufEndeMitLetztemSpan(t *testing.T) int {
	t.Helper()
	n := 0
	for _, name := range []string{"Ende des Kontexts", "Ende der Frist"} {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		uhr := neueFakeUhr()
		s, fertig := wartender(ctx, t, uhr, 1)
		s.mu.Lock()
		if name == "Ende des Kontexts" {
			cancel()
		} else {
			uhr.Vorstellen(10 * time.Second)
		}
		s.gesamt = append(s.gesamt, gesamtSpan{})
		s.mu.Unlock()
		if !warteAuf(t, fertig, time.Second) {
			t.Errorf("%s mit dem letzten Span: Ergebnis falsch, erwartet wahr", name)
		}
		n++
	}
	return n
}

// TestNachlaufWarten: der Nachlauf endet mit der Zahl oder mit der Frist (SPEC-216 Nr. 10 e, Nr. 4 c: bei Beginn + 10 s abgelaufen, eine
// Nanosekunde davor nicht). Menge: 4 Zahlen fehlender Spans (1, 2, 5 und alle 100) × 3 Folgen des Eintreffens (einzeln, alle zugleich,
// der letzte erst bei 10 s − 1 ns) gleich 12 Fälle, in jedem hält der Aufruf bis zum letzten Span und meldet wahr, ohne dass die Uhr vor der
// Frist vorgestellt wird; 4 Fälle der Frist (derselben Zahlen, kein Span kommt): offen bei 5 s, 9 s, 9,999 s und 10 s − 1 ns, zu bei 10 s, Zahl
// unverändert, falsch; 2 Fälle des Kontexts (Canceled ohne und mit Ursache); dazu 18 Fälle der Ränder (nachlaufRaender) gleich 36.
func TestNachlaufWarten(t *testing.T) {
	wache(t, 60*time.Second)
	n := nachlaufRaender(t)
	if n != 18 {
		t.Errorf("%d Fälle der Ränder, erwartet 18", n)
	}
	for _, fehlend := range []int{1, 2, 5, 100} {
		for folge := range 3 {
			uhr := neueFakeUhr()
			start := uhr.Jetzt()
			s, fertig := wartender(context.Background(), t, uhr, fehlend)
			for i := range fehlend {
				if folge == 2 && i == fehlend-1 {
					uhr.Vorstellen(nachlaufFrist - time.Nanosecond)
					bleibtOffen(t, fertig, "bei 10 s − 1 ns")
				}
				s.nimm(gz("g", 1))
				if folge == 0 && i < fehlend-1 {
					bleibtOffen(t, fertig, "vor dem letzten Span")
				}
			}
			if got := warteAuf(t, fertig, time.Second); !got || s.anzahl() != 100 || (folge != 2 && !uhr.Jetzt().Equal(start)) {
				t.Errorf("%d fehlend, Folge %d: Ergebnis %v, %d Spans; erwartet wahr und 100", fehlend, folge, got, s.anzahl())
			}
			n++
		}
		uhr := neueFakeUhr()
		s, fertig := wartender(context.Background(), t, uhr, fehlend)
		vorher := uhr.Jetzt()
		for _, bis := range []time.Duration{5 * time.Second, 9 * time.Second, 9999 * time.Millisecond, nachlaufFrist - time.Nanosecond} {
			uhr.Vorstellen(bis - uhr.Jetzt().Sub(vorher))
			bleibtOffen(t, fertig, "vor Ablauf der Frist")
		}
		uhr.Vorstellen(time.Nanosecond)
		if got := warteAuf(t, fertig, time.Second); got || s.anzahl() != 100-fehlend || uhr.Jetzt().Sub(vorher) != 10*time.Second {
			t.Errorf("%d fehlend, Frist: Ergebnis %v, %d Spans, Uhr %v; erwartet falsch, %d und 10 s", fehlend, got, s.anzahl(), uhr.Jetzt().Sub(vorher), 100-fehlend)
		}
		n++
	}
	for name, ende := range map[string]func(context.CancelCauseFunc){
		"Canceled ohne Ursache": func(c context.CancelCauseFunc) { c(nil) },
		"Canceled mit Ursache":  func(c context.CancelCauseFunc) { c(errFrist) },
	} {
		ctx, cancel := context.WithCancelCause(context.Background())
		_, fertig := wartender(ctx, t, neueFakeUhr(), 1)
		ende(cancel)
		if got := warteAuf(t, fertig, time.Second); got {
			t.Errorf("%s: Ergebnis wahr, erwartet falsch", name)
		}
		n++
	}
	if n != 36 {
		t.Errorf("%d Fälle, erwartet 36", n)
	}
}

// TestModusJeBetriebsart: recorder.mode muss gleich der Betriebsart des Laufs sein (SPEC-216 Nr. 10 d); jeder andere Wert, ein fehlendes Attribut, null
// und ein Wert anderer Art sind modus, und die Zeile bleibt eine Span-Zeile. Menge: Betriebsart (replay, mock) × Wert (die vier Modi des
// Produkts, `Replay`, der Wert `beide`, leer, fehlend, null, Zahl, Wahrheitswert; 11) × Gesamtspan ohne und mit Kind gleich 44, Zählprüfung gegen
// die vier Modi (`SPEC-075`).
func TestModusJeBetriebsart(t *testing.T) {
	wache(t, 30*time.Second)
	werte := []string{`"record"`, `"replay"`, `"mock"`, `"hybrid"`, `"Replay"`, `"beide"`, `""`, ``, `null`, `1`, `true`}
	n := 0
	for _, betriebsart := range []string{"replay", "mock"} {
		kindZeile := `{"ereignis":"span","msg":"` + kindName(betriebsart) + `","span_id":"k","parent_id":"g","duration_ms":0.500}` + "\n"
		for _, roh := range werte {
			for kinder, vorweg := range []string{"", kindZeile} {
				zeile := `{"ereignis":"span","msg":"odata.gateway.request","span_id":"g","duration_ms":1.000`
				if roh != "" {
					zeile += `,"recorder.mode":` + roh
				}
				s := neuerSammler(betriebsart)
				fremd, err := lies(strings.NewReader(vorweg+zeile+"}\n"), s.nimm)
				m, v := s.auswerte(0, 1)
				if err != nil || fremd != 0 || s.anzahl() != 1 || m.Kind != kinder || v.Modus != (roh != `"`+betriebsart+`"`) || v.Zaehlung {
					t.Errorf("%s mit recorder.mode %q und %d Kind: fremd %d, Spans %d, Kind %d, Verstöße %+v; erwartet modus %v", betriebsart, roh, kinder, fremd, s.anzahl(), m.Kind, v, roh != `"`+betriebsart+`"`)
				}
				n++
			}
		}
	}
	if n != 2*11*2 {
		t.Errorf("%d Fälle, erwartet 44", n)
	}
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
	for name, zeilen := range map[string][]spanZeile{
		"Kind und Gesamtspan 0": {kz("a", 0), gz("a", 0)}, "Summe 0, ein Gesamtspan ohne Kind davor": {gz("a", 0), kz("b", 0), gz("b", 0)},
		"ohne Kind, Dauern über 0": {gz("a", 1000), gz("b", 2000)}, "ohne Kind, Dauern 0": {gz("a", 0), gz("b", 0)},
	} {
		if m, _ := sammle("replay", zeilen...).auswerte(0, 2); m.KindAnteil != 0 || m.KindP95 != 0 {
			t.Errorf("%s: kind_anteil %d, kind_p95 %d; erwartet 0 und 0", name, m.KindAnteil, m.KindP95)
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
