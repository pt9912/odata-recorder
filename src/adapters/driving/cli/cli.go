// Package cli ist der Driving Adapter der Kommandozeile (SPEC-138): Er wählt das Kommando aus
// den Argumenten, übersetzt SIGTERM und SIGINT in den Abbruch des Kontexts von Start und bildet
// das Ergebnis auf den Exit-Code des Prozesses ab. Er importiert nur die Standardbibliothek.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// Exit-Codes (SPEC-138 Nr. 5).
const (
	exitOK       = 0
	exitStart    = 1
	exitAufruf   = 2
	exitSIGINT   = 130
	exitSonstige = 143
)

// Namen der Kommandos und der Version (SPEC-138 Nr. 2, 3).
const (
	kommandoName    = "odata-recorder"
	kommandoStart   = "start"
	kommandoVersion = "version"
)

// signalKanal ist die Größe des Kanals der Laufzeit-Anmeldung: Platz für das erste und das zweite Signal.
const signalKanal = 2

// Konfig trägt, was der Composition Root dem Adapter übergibt (SPEC-138 Nr. 1).
type Konfig struct {
	// Version ist die Software-Version.
	Version string
	// Start ist der Start des Dienstes; nil heißt: kein Kommando start.
	Start func(ctx context.Context, stdout, stderr io.Writer) int
	// Stdout und Stderr sind die Ausgabeströme; nil verwirft.
	Stdout, Stderr io.Writer
	// Signale meldet SIGTERM und SIGINT an und liefert Kanal und Abmeldung; nil ist die Laufzeit.
	Signale func() (<-chan os.Signal, func())
}

// Run führt das Kommando aus args (ohne Programmnamen) aus und liefert den Exit-Code (SPEC-138).
func Run(args []string, k Konfig) int {
	if k.Stdout == nil {
		k.Stdout = io.Discard
	}
	if k.Stderr == nil {
		k.Stderr = io.Discard
	}
	switch {
	case len(args) == 0 || (len(args) == 1 && args[0] == kommandoVersion):
		_, _ = fmt.Fprintf(k.Stdout, "%s %s\n", kommandoName, k.Version)
		return exitOK
	case len(args) == 1 && args[0] == kommandoStart && k.Start != nil:
		return starten(k)
	}
	_, _ = fmt.Fprintf(k.Stderr, "Unbekanntes Kommando: %s\n", args[0])
	return exitAufruf
}

// starten meldet die Signale an, ruft Start und wartet auf dessen Rückkehr; das erste Signal bricht
// den Kontext ab, ein weiteres erzwingt das Ende (SPEC-138 Nr. 6, 7).
func starten(k Konfig) int {
	signale := k.Signale
	if signale == nil {
		signale = laufzeitSignale
	}
	kanal, abmelden := signale()
	defer abmelden()
	ctx, abbrechen := context.WithCancel(context.Background())
	defer abbrechen()

	erstes := false
	select {
	case <-kanal:
		abbrechen()
		erstes = true
	default:
	}
	ergebnis := make(chan int, 1)
	go func() { ergebnis <- k.Start(ctx, k.Stdout, k.Stderr) }()
	for {
		select {
		case code := <-ergebnis:
			return exitCode(code)
		case sig := <-kanal:
			if erstes {
				return erzwungen(sig)
			}
			erstes = true
			abbrechen()
		}
	}
}

// laufzeitSignale meldet SIGTERM und SIGINT bei der Laufzeit an.
func laufzeitSignale() (<-chan os.Signal, func()) {
	kanal := make(chan os.Signal, signalKanal)
	signal.Notify(kanal, syscall.SIGTERM, os.Interrupt)
	return kanal, func() { signal.Stop(kanal) }
}

// exitCode bildet die Rückgabe von Start ab: 0 bleibt 0, jede andere wird 1.
func exitCode(rueckgabe int) int {
	if rueckgabe == exitOK {
		return exitOK
	}
	return exitStart
}

// erzwungen liefert den Code des erzwungenen Endes: 130 bei SIGINT, sonst 143.
func erzwungen(sig os.Signal) int {
	if sig == os.Interrupt {
		return exitSIGINT
	}
	return exitSonstige
}
