package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeLast ersetzt /proc/loadavg: Lies liefert den Inhalt des Aufrufs mit der laufenden Nummer (ab 0) aus inhalte, danach den letzten; es zählt
// jeden Aufruf und meldet ihn auf gelesen. Ist blockiereAb gesetzt, hält der Aufruf mit dieser Nummer an, bis freigabe geschlossen wird.
type fakeLast struct {
	mu          sync.Mutex
	inhalte     []string
	aufrufe     int
	gelesen     chan int
	blockiereAb int
	blockiert   chan struct{}
	freigabe    chan struct{}
}

func (f *fakeLast) Lies() ([]byte, error) {
	f.mu.Lock()
	i := f.aufrufe
	f.aufrufe++
	f.mu.Unlock()
	if f.freigabe != nil && i == f.blockiereAb {
		f.blockiert <- struct{}{}
		<-f.freigabe
	}
	f.gelesen <- i
	return []byte(f.inhalte[min(i, len(f.inhalte)-1)]), nil
}

func neueFakeLast(inhalte ...string) *fakeLast {
	return &fakeLast{inhalte: inhalte, gelesen: make(chan int, 64), blockiert: make(chan struct{}, 1)}
}

// fehlerLast liefert nie einen Inhalt.
type fehlerLast struct{}

func (fehlerLast) Lies() ([]byte, error) { return []byte("1.00 1.00 1.00"), errors.New("nicht lesbar") }

// TestLastProben: zulässig bei 2 × Last ≤ Kerne (gleich zulässig): 9 Grenzen (bei 20 Kernen 10,00 und 10,01, bei 1 Kern 0,50, 0,51 und 0,00,
// dazu 2 Kerne 1,00 und 1,01 und 19 Kerne 9,50 und 9,51); 14 nicht lesbare Formen (die 5 des Plans: leer, Text, drei Nachkommastellen, -1, 1e1,
// dazu 8 weitere und der Fehler der Quelle) tragen den kleinsten Wert über der Schwelle und sind nicht zulässig; 17 lesbare Formen (6 mit weiteren
// Feldern, 2 mit allein der Zahl, 4 mit einem Trennzeichen aus Tab, LF, CR und U+00A0 hinter ihr, 5 mit einem aus Leerzeichen, Tab, LF, CR und
// U+00A0 davor; die Zeichen jenseits von ASCII aus Bytewerten) mit ihren
// Hundertsteln; eine Spitze in der Mitte der Messphase (24 Proben, die 12. darüber) macht den Lauf ungültig mit genau dieser Probe als Maximum;
// die Zeitpunkte der 24 Proben als Literal; vor dem Aufbau darüber läuft der Aufbau 0 mal.
func TestLastProben(t *testing.T) {
	wache(t, 30*time.Second)
	grenzen := []struct {
		kerne       int
		inhalt      string
		hundertstel int
		zulaessig   bool
	}{
		{20, "10.00 9.00 8.00 1/200 12\n", 1000, true}, {20, "10.01 9.00 8.00 1/200 12\n", 1001, false},
		{1, "0.50 0.40 0.30 1/10 5\n", 50, true}, {1, "0.51 0.40 0.30 1/10 5\n", 51, false}, {1, "0.00 0.00 0.00 1/10 5\n", 0, true},
		{2, "1.00 0.00 0.00 1/10 5\n", 100, true}, {2, "1.01 0.00 0.00 1/10 5\n", 101, false},
		{19, "9.50 0.00 0.00 1/10 5\n", 950, true}, {19, "9.51 0.00 0.00 1/10 5\n", 951, false},
	}
	for _, g := range grenzen {
		if p := nimmProbe(neueFakeLast(g.inhalt), g.kerne); p.Hundertstel != g.hundertstel || p.Zulaessig != g.zulaessig {
			t.Errorf("%d Kerne, %q: %+v; erwartet %d Hundertstel, zulässig %v", g.kerne, g.inhalt, p, g.hundertstel, g.zulaessig)
		}
	}
	unlesbar := []string{
		"", "abc 1 2", "10.001 1 1", "-1 1 1", "1e1 1 1", ".5 1 1", "5. 1 1", " \n", "1,5 1 1", "+1 1 1", "0x10 1 1", "1.5.2 1 1",
		"99999999999999999999 1 1",
	}
	for _, inhalt := range unlesbar {
		if p := nimmProbe(neueFakeLast(inhalt), 20); p.Hundertstel != 1001 || p.Zulaessig {
			t.Errorf("%q: %+v; erwartet 1001 Hundertstel und nicht zulässig", inhalt, p)
		}
	}
	if p := nimmProbe(fehlerLast{}, 20); p.Hundertstel != 1001 || p.Zulaessig {
		t.Errorf("Fehler der Quelle: %+v; erwartet 1001 Hundertstel und nicht zulässig", p)
	}
	if n := len(unlesbar) + 1; n != 14 {
		t.Errorf("%d nicht lesbare Formen, erwartet 14", n)
	}
	nbsp := string([]byte{0xc2, 0xa0})
	lesbar := map[string]int{"3 1 1": 300, "3.5 1 1": 350, "03.50 1 1": 350, "0.05 1 1": 5, "0 1 1": 0, "0.1 1 1": 10}
	for _, inhalt := range []string{
		"2.50", "2.50\n",
		"2.50\t9", "2.50\n9", "2.50\r9", "2.50" + nbsp + "9",
		" 2.50", "\t2.50", "\n2.50", "\r2.50", nbsp + "2.50",
	} {
		lesbar[inhalt] = 250
	}
	if len(lesbar) != 17 {
		t.Fatalf("%d lesbare Formen, erwartet 17", len(lesbar))
	}
	for inhalt, h := range lesbar {
		if p := nimmProbe(neueFakeLast(inhalt), 20); p.Hundertstel != h || !p.Zulaessig {
			t.Errorf("%q: %+v; erwartet %d Hundertstel und zulässig", inhalt, p, h)
		}
	}
	lastKlassen(t)
	spitze(t)
	vorDemAufbau(t)
}

// lastKlassen führt die Menge der Zahlen `d.dd` als Ziffer × Stelle (30 Elemente: die Ziffern 0 bis 9 an den drei Stellen von `1.11`, Erwartung
// aus den Stellenwerten) und die Schwelle über 1 bis 64 Kerne an beiden Seiten (128 Elemente, Orakel 2 × Last ≤ Kerne in Hundertsteln).
func lastKlassen(t *testing.T) {
	t.Helper()
	gewicht := []int{100, 10, 1}
	n := 0
	for stelle, w := range gewicht {
		for ziffer := range 10 {
			ziffern := []byte("111")
			ziffern[stelle] = byte('0' + ziffer)
			inhalt := string(ziffern[:1]) + "." + string(ziffern[1:]) + " 1.00 1.00 1/10 5\n"
			if p := nimmProbe(neueFakeLast(inhalt), 64); p.Hundertstel != 111+(ziffer-1)*w || !p.Zulaessig {
				t.Errorf("%q: %+v; erwartet %d Hundertstel", inhalt, p, 111+(ziffer-1)*w)
			}
			n++
		}
	}
	for kerne := 1; kerne <= 64; kerne++ {
		for _, h := range []int{50 * kerne, 50*kerne + 1} {
			inhalt := fmt.Sprintf("%d.%02d 0.00 0.00 1/10 5\n", h/100, h%100)
			if p := nimmProbe(neueFakeLast(inhalt), kerne); p.Hundertstel != h || p.Zulaessig != (2*h <= 100*kerne) {
				t.Errorf("%d Kerne, %q: %+v; erwartet %d Hundertstel, zulässig %v", kerne, inhalt, p, h, 2*h <= 100*kerne)
			}
			n++
		}
	}
	if n != 30+128 {
		t.Errorf("%d Elemente, erwartet 158", n)
	}
}

// TestProbenEndenMitKontext: die Proben enden mit dem Kontext des Aufrufs, ob er ohne oder mit Ursache beendet wird (Canceled und die Ursache
// errFrist); der Code liest allein Done, nie Err, so dass DeadlineExceeded denselben Weg nimmt (Negativ im Plan).
func TestProbenEndenMitKontext(t *testing.T) {
	wache(t, 30*time.Second)
	for name, ende := range map[string]func(context.CancelCauseFunc){
		"ohne Ursache": func(c context.CancelCauseFunc) { c(nil) },
		"mit Ursache":  func(c context.CancelCauseFunc) { c(errFrist) },
	} {
		ctx, cancel := context.WithCancelCause(context.Background())
		p := starteProben(ctx, neueFakeUhr(), neueFakeLast("1.00 1.00 1.00 1/10 5\n"), 20, 660*time.Second)
		ende(cancel)
		warteAuf(t, p.fertig, time.Second)
		if e := p.beende(); e.Proben != 2 {
			t.Errorf("%s: %d Proben, erwartet die vor dem Warm-up und die nach dem Nachlauf", name, e.Proben)
		}
	}
}

// spitze fährt einen Lauf des Profils abnahme (660 s) mit der Uhr des Tests: 24 Proben mit der 12. bei 12,00 (20 Kerne), Zeitpunkte als Literal.
func spitze(t *testing.T) {
	t.Helper()
	inhalte := make([]string, 24)
	for i := range inhalte {
		inhalte[i] = "1.00 1.00 1.00 1/10 5\n"
	}
	inhalte[11] = "12.00 1.00 1.00 1/10 5\n"
	uhr, q := neueFakeUhr(), neueFakeLast(inhalte...)
	p := starteProben(context.Background(), uhr, q, 20, 660*time.Second)
	warteAuf(t, q.gelesen, time.Second)
	for range 22 {
		uhr.Vorstellen(30 * time.Second)
		warteAuf(t, q.gelesen, time.Second)
	}
	warteAuf(t, p.fertig, time.Second)
	uhr.Vorstellen(5 * time.Second)
	e := p.beende()
	var want []time.Duration
	for s := 0; s <= 660; s += 30 {
		want = append(want, time.Duration(s)*time.Second)
	}
	want = append(want, 665*time.Second)
	if len(p.zeiten) != 24 || len(want) != 24 {
		t.Fatalf("%d Zeitpunkte %v, erwartet 24", len(p.zeiten), p.zeiten)
	}
	for i := range want {
		if p.zeiten[i] != want[i] {
			t.Errorf("Probe %d zur Zeit %v, erwartet %v", i+1, p.zeiten[i], want[i])
		}
	}
	if e != (lastErgebnis{Proben: 24, Min: 100, Max: 1200, Schwelle: 1000, Zulaessig: false}) {
		t.Errorf("Ergebnis %+v; erwartet 24 Proben, 1,00 bis 12,00, Schwelle 10,00, nicht zulässig", e)
	}
	for i, pr := range p.proben {
		if pr.Zulaessig == (i == 11) {
			t.Errorf("Probe %d: %+v; erwartet nur die 12. über der Schwelle", i+1, pr)
		}
	}
}

// vorDemAufbau: über der Schwelle läuft der Aufbau 0 mal und der Fehler ist errLast; auf der Schwelle läuft er einmal mit dem ctx des Aufrufs
// und sein Fehler kommt zurück.
func vorDemAufbau(t *testing.T) {
	t.Helper()
	type schluessel struct{}
	ctx := context.WithValue(context.Background(), schluessel{}, "x")
	boom := errors.New("Aufbau gescheitert")
	for _, f := range []struct {
		inhalt string
		laeufe int
		err    error
	}{{"10.01 1 1", 0, errLast}, {"10.00 1 1", 1, boom}} {
		n := 0
		_, err := vorAufbau(ctx, neueFakeLast(f.inhalt), 20, func(c context.Context) error {
			n++
			if c.Value(schluessel{}) != "x" {
				t.Error("der Aufbau läuft nicht mit dem ctx des Aufrufs")
			}
			return boom
		})
		if n != f.laeufe || !errors.Is(err, f.err) {
			t.Errorf("%q: Aufbau %d mal, Fehler %v; erwartet %d mal und %v", f.inhalt, n, err, f.laeufe, f.err)
		}
	}
}

// TestLastNebenDerLast (Wirkungs-Fänger, 20 Läufe): starteProben kehrt zurück, ohne auf den ersten Tick zu warten (rot wird allein diese
// Rückkehr, `warteAuf` nach einer Sekunde; ob eine hängende Probe die Last anhält, prüft erst der Slice, der die Last fährt); beende wartet auf
// die Probe, die in der Quelle hängt: das Ergebnis zählt sie mit (3 Proben, Maximum 12,00), und es liegt erst vor, wenn die Quelle sie freigibt.
func TestLastNebenDerLast(t *testing.T) {
	wache(t, 60*time.Second)
	for lauf := range 20 {
		uhr := neueFakeUhr()
		q := neueFakeLast("1.00 1.00 1.00 1/10 5\n", "12.00 1.00 1.00 1/10 5\n", "1.00 1.00 1.00 1/10 5\n")
		q.blockiereAb, q.freigabe = 1, make(chan struct{})
		gestartet := make(chan *lastProben, 1)
		go func() { gestartet <- starteProben(context.Background(), uhr, q, 20, 660*time.Second) }()
		p := warteAuf(t, gestartet, time.Second)
		uhr.Vorstellen(30 * time.Second)
		warteAuf(t, q.blockiert, time.Second)
		ergebnis := make(chan lastErgebnis, 1)
		go func() { ergebnis <- p.beende() }()
		bleibtOffen(t, ergebnis, "beende vor der Freigabe der Probe")
		close(q.freigabe)
		if e := warteAuf(t, ergebnis, time.Second); e != (lastErgebnis{Proben: 3, Min: 100, Max: 1200, Schwelle: 1000, Zulaessig: false}) {
			t.Errorf("Lauf %d: Ergebnis %+v; erwartet 3 Proben von 1,00 bis 12,00, Schwelle 10,00, nicht zulässig", lauf, e)
		}
	}
}

// TestDateiLast: dateiLast liest die Datei, und eine fehlende Datei ist eine nicht lesbare Probe.
func TestDateiLast(t *testing.T) {
	wache(t, 30*time.Second)
	pfad := filepath.Join(t.TempDir(), "loadavg")
	if err := os.WriteFile(pfad, []byte("2.50 1.00 1.00 1/10 5\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if p := nimmProbe(dateiLast{pfad}, 20); p.Hundertstel != 250 || !p.Zulaessig {
		t.Errorf("Datei: %+v; erwartet 250 Hundertstel, zulässig", p)
	}
	if p := nimmProbe(dateiLast{pfad + ".fehlt"}, 20); p.Zulaessig || p.Hundertstel != 1001 {
		t.Errorf("fehlende Datei: %+v; erwartet nicht zulässig mit 1001", p)
	}
}
