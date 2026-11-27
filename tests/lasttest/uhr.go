// Das Paket ist das Programm des Lasttests (SPEC-216): Profile, Referenzdienst, Aufbau, Listen der Anfragen, Lauf und Auswertung.
// Jede Grenze der Zeit liest die Uhr über die Schnittstelle Uhr; allein diese Datei ruft die Funktionen der Zeit.
package main

import (
	"context"
	"errors"
	"syscall"
	"time"
)

// Uhr ist die Schnittstelle der Zeit (SPEC-216 Nr. 14 a), die der Test ersetzt.
type Uhr interface {
	// Jetzt liefert die Zeit der Uhr.
	Jetzt() time.Time
	// WarteBis kehrt mit nil zurück, sobald die Uhr bis erreicht hat (sofort, wenn sie ihn schon erreicht hat), und mit dem Fehler des
	// Kontexts, wenn dieser vorher endet.
	WarteBis(ctx context.Context, bis time.Time) error
	// Ticker liefert einen Kanal, der alle d einen Wert trägt, und die Funktion, die ihn beendet.
	Ticker(d time.Duration) (<-chan time.Time, func())
}

// systemuhr ist die Uhr des Prozesses.
type systemuhr struct{}

func (systemuhr) Jetzt() time.Time { return time.Now() }

func (systemuhr) WarteBis(ctx context.Context, bis time.Time) error {
	t := time.NewTimer(time.Until(bis))
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (systemuhr) Ticker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

// prozessCPU ist die CPU-Zeit, die der Prozess bisher verbraucht hat (Nutzer und System); scheitert die Abfrage, 0.
func prozessCPU() time.Duration {
	var r syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &r); err != nil {
		return 0
	}
	return time.Duration(r.Utime.Nano() + r.Stime.Nano())
}

// errFrist ist die Ursache (context.Cause) eines Kontexts, den mitFrist beendet hat.
var errFrist = errors.New("Frist überschritten")

// mitFrist liefert einen Kontext, den die Uhr nach d beendet; die Ursache ist dann errFrist. Die Funktion beendet ihn früher.
func mitFrist(ctx context.Context, uhr Uhr, d time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(ctx)
	bis := uhr.Jetzt().Add(d)
	go func() {
		if uhr.WarteBis(ctx, bis) == nil {
			cancel(errFrist)
		}
	}()
	return ctx, func() { cancel(context.Canceled) }
}
