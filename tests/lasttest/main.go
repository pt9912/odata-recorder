package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Orte im Container der Stufe lasttest (SPEC-216 Nr. 3 b, 4 a und 11).
const (
	binaerPfad = "/out/odata-recorder"
	datenPfad  = "/daten"
	lastPfad   = "/proc/loadavg"
)

// main ist das Programm /out/lasttest: es meldet INT und TERM als Ende des Kontexts an den Ablauf (SPEC-216 Nr. 4 d) und endet mit dessen
// Exit (Nr. 12 e). Kein Test führt diese Datei (Plan, Negativ (8)).
func main() { os.Exit(haupt()) }

func haupt() int {
	signale, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	w, err := liesUmgebung(os.LookupEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lasttest: %v\n", err)
		return exitAbbruch
	}
	return ablaufEingabe{
		werte: w, Uhr: systemuhr{}, Last: dateiLast{lastPfad}, DataDir: datenPfad,
		Aufbau: func(ctx context.Context, e aufbauEingabe) (aufbauErgebnis, error) { return e.aufbauen(ctx) },
		Port:   freierPort,
		Starte: func(ctx context.Context, env []string) (*prozess, error) {
			return starteProzess(binaerBefehl(ctx, binaerPfad, env))
		},
		Verbinde: verbindeTCP, CPU: prozessCPU, Aus: os.Stdout, Err: os.Stderr,
	}.fahre(signale)
}
