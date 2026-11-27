package main

import (
	"context"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Fristen des Prozesses des Binary (SPEC-216 Nr. 9 b, d); bereitAbstand ist der Abstand der Verbindungsversuche der Bereitschaft.
const (
	bereitFrist   = 10 * time.Second
	bereitAbstand = 10 * time.Millisecond
	endeFrist     = 15 * time.Second
)

// prozess ist ein gestartetes Binary: seine Kennung, seine Ströme, der Kanal, der mit seinem Ende schließt, sein Exit (gültig nach dem Ende;
// nach einem Signal 128 plus dessen Nummer) und die Zustellung eines Signals.
type prozess struct {
	PID    int
	Stdout io.ReadCloser
	Stderr io.ReadCloser
	Ende   <-chan struct{}
	Exit   func() int
	Signal func(os.Signal) error
}

// binaerBefehl ist der Aufruf `<pfad> start` mit genau den Einträgen von env als Umgebung (SPEC-216 Nr. 9 a). Das Ende von ctx beendet den
// Prozess nicht; das tut allein beendeBinaer.
func binaerBefehl(ctx context.Context, pfad string, env []string) *exec.Cmd {
	cmd := exec.CommandContext(context.WithoutCancel(ctx), pfad, "start")
	cmd.Env = env
	return cmd
}

// starteProzess startet cmd mit stdout und stderr als Pipes, die der Aufrufer bis zum Ende liest und danach schließt.
func starteProzess(cmd *exec.Cmd) (*prozess, error) {
	aus, ausSchreib, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	fehler, fehlerSchreib, err := os.Pipe()
	if err != nil {
		_, _ = aus.Close(), ausSchreib.Close()
		return nil, err
	}
	cmd.Stdout, cmd.Stderr = ausSchreib, fehlerSchreib
	err = cmd.Start()
	_, _ = ausSchreib.Close(), fehlerSchreib.Close()
	if err != nil {
		_, _ = aus.Close(), fehler.Close()
		return nil, err
	}
	ende := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(ende)
	}()
	return &prozess{
		PID: cmd.Process.Pid, Stdout: aus, Stderr: fehler, Ende: ende,
		Exit: func() int {
			<-ende
			if w, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && w.Signaled() {
				return 128 + int(w.Signal())
			}
			return cmd.ProcessState.ExitCode()
		},
		Signal: cmd.Process.Signal,
	}, nil
}

// beendet meldet, ob ende geschlossen ist, ohne zu warten.
func beendet(ende <-chan struct{}) bool {
	select {
	case <-ende:
		return true
	default:
		return false
	}
}

// freierPort wählt einen freien Port von 127.0.0.1 und liefert die Adresse `127.0.0.1:<Port>`.
func freierPort(ctx context.Context) (string, error) {
	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	adresse := l.Addr().String()
	return adresse, l.Close()
}

// verbindeTCP baut eine TCP-Verbindung zu adresse auf und schließt sie; es sendet nichts.
func verbindeTCP(ctx context.Context, adresse string) error {
	var d net.Dialer
	c, err := d.DialContext(ctx, "tcp", adresse)
	if err != nil {
		return err
	}
	return c.Close()
}

// bereitschaft ist das Ergebnis des Wartens auf das Binary.
type bereitschaft int

// Die Ergebnisse: eine Verbindung gelang, die Frist oder ctx endete, oder das Binary endete vorher.
const (
	bereit bereitschaft = iota
	nichtBereit
	vorherBeendet
)

// warteBereit versucht eine Verbindung zu adresse, bis sie gelingt, das Binary endet, ctx endet oder die Uhr bereitFrist seit dem Aufruf
// erreicht (Gleichheit eingeschlossen, SPEC-216 Nr. 9 b und 4 c); eine Verbindung, die erst an der Frist gelingt, ist keine Bereitschaft.
func warteBereit(ctx context.Context, uhr Uhr, verbinde func(context.Context, string) error, adresse string, ende <-chan struct{}) bereitschaft {
	bis := uhr.Jetzt().Add(bereitFrist)
	fctx, cancel := mitFrist(ctx, uhr, bereitFrist)
	defer cancel()
	wctx, stop := context.WithCancel(fctx)
	defer stop()
	go func() {
		select {
		case <-ende:
			stop()
		case <-wctx.Done():
		}
	}()
	for {
		err := verbinde(fctx, adresse)
		switch {
		case beendet(ende):
			return vorherBeendet
		case ctx.Err() != nil || !uhr.Jetzt().Before(bis):
			return nichtBereit
		case err == nil:
			return bereit
		}
		_ = uhr.WarteBis(wctx, uhr.Jetzt().Add(bereitAbstand))
	}
}

// beendeBinaer sendet SIGTERM, wartet höchstens endeFrist der Uhr auf das Ende und sendet danach SIGKILL (SPEC-216 Nr. 9 d); true heißt,
// SIGKILL war nötig. Das Ende von ctx kürzt die Frist nicht.
func beendeBinaer(ctx context.Context, uhr Uhr, p *prozess) bool {
	_ = p.Signal(syscall.SIGTERM)
	fctx, cancel := mitFrist(context.WithoutCancel(ctx), uhr, endeFrist)
	defer cancel()
	select {
	case <-p.Ende:
		return false
	case <-fctx.Done():
	}
	if beendet(p.Ende) {
		return false
	}
	_ = p.Signal(syscall.SIGKILL)
	<-p.Ende
	return true
}

// entleere liest stdout und stderr von p ohne Unterbrechung in eigenen Goroutinen (SPEC-216 Nr. 9 e); jede Span-Zeile von stderr geht an
// verarbeite, allein in der Goroutine von stderr. Meldet stderr einen anderen Fehler als sein Ende, wird der Rest gelesen und verworfen
// (SPEC-216 Nr. 10 a). Die Funktion wartet auf das Ende beider Ströme, schließt sie und liefert die fremden Zeilen und den Lesefehler.
func entleere(p *prozess, verarbeite func(spanZeile)) func() (fremd int, lesefehler bool) {
	var (
		wg         sync.WaitGroup
		fremd      int
		lesefehler bool
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(io.Discard, p.Stdout)
	}()
	go func() {
		defer wg.Done()
		var err error
		if fremd, err = lies(p.Stderr, verarbeite); err != nil {
			lesefehler = true
			_, _ = io.Copy(io.Discard, p.Stderr)
		}
	}()
	return func() (int, bool) {
		wg.Wait()
		_, _ = p.Stdout.Close(), p.Stderr.Close()
		return fremd, lesefehler
	}
}
