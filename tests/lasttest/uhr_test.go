package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// fakeUhr ist die Uhr des Tests: sie geht nur durch Vorstellen vor und weckt dabei jeden Wartenden und jeden Ticker, den sie erreicht.
type fakeUhr struct {
	mu      sync.Mutex
	jetzt   time.Time
	warter  []fakeWarter
	tickers []*fakeTicker
}

type fakeWarter struct {
	bis  time.Time
	frei chan struct{}
}

type fakeTicker struct {
	d         time.Duration
	naechster time.Time
	ch        chan time.Time
	gestoppt  bool
}

func neueFakeUhr() *fakeUhr { return &fakeUhr{jetzt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)} }

func (u *fakeUhr) Jetzt() time.Time {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.jetzt
}

func (u *fakeUhr) WarteBis(ctx context.Context, bis time.Time) error {
	u.mu.Lock()
	if !u.jetzt.Before(bis) {
		u.mu.Unlock()
		return nil
	}
	w := fakeWarter{bis: bis, frei: make(chan struct{})}
	u.warter = append(u.warter, w)
	u.mu.Unlock()
	select {
	case <-w.frei:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (u *fakeUhr) Ticker(d time.Duration) (<-chan time.Time, func()) {
	u.mu.Lock()
	defer u.mu.Unlock()
	t := &fakeTicker{d: d, naechster: u.jetzt.Add(d), ch: make(chan time.Time, 1)}
	u.tickers = append(u.tickers, t)
	return t.ch, func() {
		u.mu.Lock()
		defer u.mu.Unlock()
		t.gestoppt = true
	}
}

// wartende ist die Zahl der Wartenden in WarteBis.
func (u *fakeUhr) wartende() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.warter)
}

// Vorstellen setzt die Uhr um d vor; jeder Wartende, dessen Zeit erreicht ist, kehrt zurück, bevor Vorstellen endet.
func (u *fakeUhr) Vorstellen(d time.Duration) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.jetzt = u.jetzt.Add(d)
	offen := u.warter[:0]
	for _, w := range u.warter {
		if u.jetzt.Before(w.bis) {
			offen = append(offen, w)
		} else {
			close(w.frei)
		}
	}
	u.warter = offen
	for _, t := range u.tickers {
		for !t.gestoppt && !u.jetzt.Before(t.naechster) {
			select {
			case t.ch <- t.naechster:
			default:
			}
			t.naechster = t.naechster.Add(t.d)
		}
	}
}

// wache beendet den Prozess mit panic, wenn der Test nach d der wirklichen Zeit nicht fertig ist; jeder Test ruft sie zuerst. Sie und warteAuf
// lesen die Zeit des Prozesses ohne die Schnittstelle Uhr, damit ein Fehler in systemuhr sie nicht mitnimmt; diese Datei darf allein
// time.NewTimer nennen (TestPaketOhneZeit); kein Test führt das panic von wache und die Grenze von warteAuf (Plan, Negativ (4)).
func wache(t *testing.T, d time.Duration) {
	t.Helper()
	fertig := make(chan struct{})
	t.Cleanup(func() { close(fertig) })
	name := t.Name()
	go func() {
		z := time.NewTimer(d)
		defer z.Stop()
		select {
		case <-z.C:
			panic("Frist des Tests " + name + " überschritten")
		case <-fertig:
		}
	}()
}

// warteAuf liefert den Wert aus ch und beendet den Test, wenn nach d der wirklichen Zeit keiner da ist.
func warteAuf[T any](t *testing.T, ch <-chan T, d time.Duration) T {
	t.Helper()
	z := time.NewTimer(d)
	defer z.Stop()
	select {
	case v := <-ch:
		return v
	case <-z.C:
		t.Fatalf("nach %v kein Wert", d)
	}
	var null T
	return null
}

// pause wartet d der wirklichen Zeit.
func pause(d time.Duration) {
	z := time.NewTimer(d)
	defer z.Stop()
	<-z.C
}

// bisWahr gibt anderen Goroutinen den Vortritt, bis f wahr ist, höchstens 5 s der wirklichen Zeit, und liefert, ob f wahr wurde.
func bisWahr(f func() bool) bool {
	z := time.NewTimer(5 * time.Second)
	defer z.Stop()
	for !f() {
		select {
		case <-z.C:
			return false
		default:
			runtime.Gosched()
		}
	}
	return true
}

// TestFakeUhr: die Uhr des Tests weckt einen Wartenden erst bei Erreichen seiner Zeit, kehrt für eine erreichte Zeit sofort zurück, endet mit
// dem Kontext und tickt je Intervall; mitFrist beendet den Kontext erst über der Frist, mit der Ursache errFrist.
func TestFakeUhr(t *testing.T) {
	wache(t, 30*time.Second)
	u := neueFakeUhr()
	start := u.Jetzt()
	fertig := make(chan error, 1)
	go func() { fertig <- u.WarteBis(context.Background(), start.Add(10*time.Second)) }()
	for u.wartende() < 1 {
		runtime.Gosched()
	}
	u.Vorstellen(9 * time.Second)
	if len(fertig) != 0 {
		t.Fatal("WarteBis kehrte vor der Zeit zurück")
	}
	u.Vorstellen(time.Second)
	if err := <-fertig; err != nil {
		t.Errorf("WarteBis zur erreichten Zeit: %v", err)
	}
	if err := u.WarteBis(context.Background(), start); err != nil {
		t.Errorf("WarteBis für eine vergangene Zeit: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := u.WarteBis(ctx, start.Add(time.Hour)); !errors.Is(err, context.Canceled) {
		t.Errorf("WarteBis mit beendetem Kontext: %v", err)
	}
	ch, stop := u.Ticker(5 * time.Second)
	u.Vorstellen(4 * time.Second)
	if len(ch) != 0 {
		t.Error("Ticker vor dem Intervall")
	}
	u.Vorstellen(time.Second)
	if len(ch) != 1 {
		t.Error("Ticker nach dem Intervall ohne Wert")
	}
	stop()
	<-ch
	u.Vorstellen(10 * time.Second)
	if len(ch) != 0 {
		t.Error("gestoppter Ticker tickt")
	}
	fctx, fcancel := mitFrist(context.Background(), u, time.Minute)
	defer fcancel()
	for u.wartende() < 1 {
		runtime.Gosched()
	}
	u.Vorstellen(time.Minute - time.Second)
	if u.wartende() != 1 {
		t.Fatal("mitFrist ließ den Wartenden vor der Frist frei")
	}
	u.Vorstellen(time.Second)
	warteAuf(t, fctx.Done(), 5*time.Second)
	if !errors.Is(context.Cause(fctx), errFrist) {
		t.Errorf("Ursache %v, erwartet errFrist", context.Cause(fctx))
	}
	ctx2, cancel2 := mitFrist(context.Background(), u, time.Hour)
	cancel2()
	warteAuf(t, ctx2.Done(), 5*time.Second)
	if errors.Is(context.Cause(ctx2), errFrist) {
		t.Error("mitFrist: die Funktion zum Beenden gibt errFrist als Ursache an")
	}
}

// TestSystemuhr: die Uhr des Prozesses zeigt die Zeit des Dateisystems (innerhalb von 5 s), kehrt für eine vergangene Zeit sofort zurück,
// wartet bis zur Zeit, endet mit dem Kontext und tickt bis zum Aufruf der Funktion, die ihren Ticker beendet; jeder Aufruf von WarteBis läuft
// in einer Goroutine, deren Rückkehr der Test innerhalb von 1 s verlangt (kein Warten dieses Tests dauert länger als eine Sekunde).
func TestSystemuhr(t *testing.T) {
	wache(t, 30*time.Second)
	u := systemuhr{}
	anker := filepath.Join(t.TempDir(), "anker")
	if err := os.WriteFile(anker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(anker)
	if err != nil {
		t.Fatal(err)
	}
	if abstand := u.Jetzt().Sub(fi.ModTime()); abstand < -5*time.Second || abstand > 5*time.Second {
		t.Errorf("Jetzt weicht um %v von der Zeit des Dateisystems ab", abstand)
	}
	warte := func(ctx context.Context, bis time.Time) error {
		c := make(chan error, 1)
		go func() { c <- u.WarteBis(ctx, bis) }()
		return warteAuf(t, c, time.Second)
	}
	a := u.Jetzt()
	if err := warte(context.Background(), a.Add(-time.Hour)); err != nil {
		t.Errorf("WarteBis für eine vergangene Zeit: %v", err)
	}
	if err := warte(context.Background(), a.Add(20*time.Millisecond)); err != nil || u.Jetzt().Sub(a) < 20*time.Millisecond {
		t.Errorf("WarteBis: %v, vergangen %v, erwartet mindestens 20ms", err, u.Jetzt().Sub(a))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := warte(ctx, u.Jetzt().Add(time.Hour)); !errors.Is(err, context.Canceled) {
		t.Errorf("WarteBis mit beendetem Kontext: %v", err)
	}
	ch, stop := u.Ticker(5 * time.Millisecond)
	defer stop()
	warteAuf(t, ch, time.Second)
	warteAuf(t, ch, time.Second)
	stop()
	_ = warte(context.Background(), u.Jetzt().Add(40*time.Millisecond))
	select {
	case <-ch:
		t.Error("der Ticker tickt nach dem Aufruf der Funktion, die ihn beendet")
	default:
	}
}
