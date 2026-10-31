package cli_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/adapters/driving/cli"
)

// frist ist die Zeitgrenze jedes Wartens auf ein Ereignis, das eintreten muss.
const frist = 5 * time.Second

// fenster ist die Zeit, in der ein Ereignis, das nicht eintreten darf, nicht eintritt.
const fenster = 30 * time.Millisecond

// StartZustand ist ein Schnappschuss dessen, was FakeStart gesehen hat.
type StartZustand struct {
	Aufrufe         int
	Rueckkehren     int
	Ctx             []context.Context
	Stdout          []io.Writer
	Stderr          []io.Writer
	ErrBeiAufruf    []error
	AnmBeiAufruf    []int
	AbmBeiAufruf    []int
	AbmBeiRueckkehr []int
}

// FakeStart ist der Start, den der Composition Root liefert. Er hält jedes Argument, zählt Aufrufe und
// Rückkehren und liest bei Aufruf und Rückkehr die Zähler von FakeSignale.
type FakeStart struct {
	Rueckgabe    int
	WarteKontext bool
	Halten       bool
	Signale      *FakeSignale
	Bereit       chan struct{}
	Abgebrochen  chan struct{}

	freigabe chan struct{}
	frei     sync.Once
	mu       sync.Mutex
	z        StartZustand
}

// NewFakeStart legt einen Start an, der rueckgabe liefert; der Cleanup gibt ihn frei.
func NewFakeStart(t *testing.T, rueckgabe int) *FakeStart {
	t.Helper()
	f := &FakeStart{
		Rueckgabe:   rueckgabe,
		Bereit:      make(chan struct{}, 1),
		Abgebrochen: make(chan struct{}, 1),
		freigabe:    make(chan struct{}),
	}
	t.Cleanup(f.Frei)
	return f
}

// Frei gibt jedes Warten des Starts frei (mehrfach aufrufbar).
func (f *FakeStart) Frei() { f.frei.Do(func() { close(f.freigabe) }) }

// Zustand liefert den Schnappschuss.
func (f *FakeStart) Zustand() StartZustand {
	f.mu.Lock()
	defer f.mu.Unlock()
	z := f.z
	z.Ctx = append([]context.Context(nil), f.z.Ctx...)
	z.Stdout = append([]io.Writer(nil), f.z.Stdout...)
	z.Stderr = append([]io.Writer(nil), f.z.Stderr...)
	z.ErrBeiAufruf = append([]error(nil), f.z.ErrBeiAufruf...)
	z.AnmBeiAufruf = append([]int(nil), f.z.AnmBeiAufruf...)
	z.AbmBeiAufruf = append([]int(nil), f.z.AbmBeiAufruf...)
	z.AbmBeiRueckkehr = append([]int(nil), f.z.AbmBeiRueckkehr...)
	return z
}

func meldeKurz(c chan struct{}) {
	select {
	case c <- struct{}{}:
	default:
	}
}

// Func ist der Wert für Konfig.Start.
func (f *FakeStart) Func() func(context.Context, io.Writer, io.Writer) int {
	return func(ctx context.Context, stdout, stderr io.Writer) int {
		anm, abm := f.signaleZaehler()
		f.mu.Lock()
		f.z.Aufrufe++
		f.z.Ctx = append(f.z.Ctx, ctx)
		f.z.Stdout = append(f.z.Stdout, stdout)
		f.z.Stderr = append(f.z.Stderr, stderr)
		f.z.ErrBeiAufruf = append(f.z.ErrBeiAufruf, ctx.Err())
		f.z.AnmBeiAufruf = append(f.z.AnmBeiAufruf, anm)
		f.z.AbmBeiAufruf = append(f.z.AbmBeiAufruf, abm)
		f.mu.Unlock()
		meldeKurz(f.Bereit)
		if f.WarteKontext {
			select {
			case <-ctx.Done():
			case <-f.freigabe:
			}
			meldeKurz(f.Abgebrochen)
		}
		if f.Halten {
			<-f.freigabe
		}
		_, abm = f.signaleZaehler()
		f.mu.Lock()
		f.z.Rueckkehren++
		f.z.AbmBeiRueckkehr = append(f.z.AbmBeiRueckkehr, abm)
		f.mu.Unlock()
		return f.Rueckgabe
	}
}

func (f *FakeStart) signaleZaehler() (anm, abm int) {
	if f.Signale == nil {
		return 0, 0
	}
	return f.Signale.Anmeldungen(), f.Signale.Abmeldungen()
}

// FakeSignale ist die austauschbare Signalquelle. Sie zählt Anmeldungen und Abmeldungen.
type FakeSignale struct {
	// Verzoegerung hält die Anmeldung auf, bevor sie zählt: Ein Start, der neben der Anmeldung läuft,
	// sieht in dieser Zeit den Zähler noch auf 0.
	Verzoegerung time.Duration
	kanal        chan os.Signal
	mu           sync.Mutex
	anm          int
	abm          int
}

// NewFakeSignale legt eine Quelle an.
func NewFakeSignale() *FakeSignale { return &FakeSignale{kanal: make(chan os.Signal, 8)} }

// Func ist der Wert für Konfig.Signale.
func (f *FakeSignale) Func() func() (<-chan os.Signal, func()) {
	return func() (<-chan os.Signal, func()) {
		time.Sleep(f.Verzoegerung)
		f.mu.Lock()
		f.anm++
		f.mu.Unlock()
		return f.kanal, func() {
			f.mu.Lock()
			f.abm++
			f.mu.Unlock()
		}
	}
}

// Senden legt ein Signal in den Kanal.
func (f *FakeSignale) Senden(s os.Signal) { f.kanal <- s }

// Anmeldungen zählt die Aufrufe der Anmeldung.
func (f *FakeSignale) Anmeldungen() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.anm
}

// Abmeldungen zählt die Aufrufe der Abmeldung.
func (f *FakeSignale) Abmeldungen() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.abm
}

// fehlerWriter ist ein Writer, dessen Schreiben scheitert.
type fehlerWriter struct{ aufrufe int }

func (w *fehlerWriter) Write([]byte) (int, error) {
	w.aufrufe++
	return 0, errors.New("Schreibfehler")
}

// lauf ruft Run in einer Goroutine und liefert den Kanal des Exit-Codes.
func lauf(args []string, k cli.Konfig) <-chan int {
	c := make(chan int, 1)
	go func() { c <- cli.Run(args, k) }()
	return c
}

// erwarte wartet höchstens frist auf einen Wert.
func erwarte[T any](t *testing.T, c <-chan T, was string) T {
	t.Helper()
	select {
	case v := <-c:
		return v
	case <-time.After(frist):
		t.Fatalf("Zeitgrenze: %s", was)
		var zero T
		return zero
	}
}

// bleibtAus prüft, dass im fenster kein Wert eintrifft.
func bleibtAus[T any](t *testing.T, c <-chan T, was string) {
	t.Helper()
	select {
	case v := <-c:
		t.Fatalf("%s: unerwartet %v", was, v)
	case <-time.After(fenster):
	}
}

// echteSignale meldet einen eigenen Kanal an, damit ein Signal an den Prozess ihn nie beendet.
func echteSignale(t *testing.T) {
	t.Helper()
	eigen := make(chan os.Signal, 8)
	signal.Notify(eigen, syscall.SIGTERM, syscall.SIGINT)
	t.Cleanup(func() {
		// Ein Signal, das der Test gesendet hat und der Prozess noch nicht zugestellt hat, trifft sonst
		// nach der Abmeldung die Voreinstellung der Laufzeit und beendet den Testprozess.
		time.Sleep(fenster)
		signal.Stop(eigen)
	})
}

// sendeEcht schickt sig an den eigenen Prozess.
func sendeEcht(t *testing.T, sig syscall.Signal) {
	t.Helper()
	if err := syscall.Kill(syscall.Getpid(), sig); err != nil {
		t.Fatalf("Signal %v: %v", sig, err)
	}
}

// konfig baut die Konfiguration; ein nil-Start und ein nil-FakeSignale bleiben nil.
func konfig(version string, start *FakeStart, sig *FakeSignale, stdout, stderr io.Writer) cli.Konfig {
	k := cli.Konfig{Version: version, Stdout: stdout, Stderr: stderr}
	if start != nil {
		k.Start = start.Func()
	}
	if sig != nil {
		k.Signale = sig.Func()
	}
	return k
}
