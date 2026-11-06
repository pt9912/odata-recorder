package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const ausgabeMax = 64 << 10

// begrenzt hält den Anfang einer Ausgabe und verwirft den Rest.
type begrenzt struct{ b []byte }

func (g *begrenzt) Write(p []byte) (int, error) {
	if frei := ausgabeMax - len(g.b); frei > 0 {
		g.b = append(g.b, p[:min(len(p), frei)]...)
	}
	return len(p), nil
}

// echt führt Go-Werkzeug und Test-Binaries als Prozesse in eigenen Prozessgruppen aus (SPEC-210 Nr. 9, 10).
type echt struct {
	env     []string
	cache   string
	bauZeit time.Duration
}

// prozess führt einen Befehl bis zum Ende, zur Frist oder zum Abbruch von ctx; bei Frist und Abbruch
// endet die ganze Prozessgruppe.
func prozess(ctx context.Context, dir string, env []string, frist time.Duration, name string, args ...string) Ausgang {
	c, stop := context.WithTimeout(ctx, frist)
	defer stop()
	cmd := exec.CommandContext(c, name, args...)
	cmd.Dir, cmd.Env = dir, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 5 * time.Second
	var puffer begrenzt
	cmd.Stdout, cmd.Stderr = &puffer, &puffer
	err := cmd.Run()
	var beendet *exec.ExitError
	switch {
	case c.Err() != nil:
		return Ausgang{Code: -1, Text: string(puffer.b), Abgebrochen: true}
	case err != nil && !errors.As(err, &beendet):
		return Ausgang{Code: -1, Text: err.Error(), Abgebrochen: true}
	}
	return Ausgang{Code: cmd.ProcessState.ExitCode(), Text: string(puffer.b)}
}

// Bauen übersetzt das Paket zu einem Test-Binary (SPEC-210 Nr. 5 b).
func (e *echt) Bauen(ctx context.Context, b Bau) Ausgang {
	args := []string{"test", "-c", "-trimpath", "-buildvcs=false", "-p", "2", "-o", b.Ziel}
	cgo := "CGO_ENABLED=0"
	if b.Detektor {
		args = append(args, "-race")
		cgo = "CGO_ENABLED=1"
	}
	a := prozess(ctx, b.Dir, append(slicesClone(e.env), cgo), e.bauZeit, "go", append(args, b.Paket)...)
	if a.Code == 0 && !a.Abgebrochen && !vorhanden(b.Ziel) {
		a.KeinBinary = true
	}
	return a
}

func slicesClone(s []string) []string { return append([]string(nil), s...) }

// Starten startet ein Test-Binary.
func (e *echt) Starten(ctx context.Context, s Start) Ausgang {
	return prozess(ctx, s.Dir, e.env, s.Frist, s.Binary, s.Args...)
}

// CacheGroesse summiert die Dateien des Build-Caches.
func (e *echt) CacheGroesse() int64 {
	var summe int64
	_ = filepath.WalkDir(e.cache, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if i, err := d.Info(); err == nil && i.Mode().IsRegular() {
			summe += i.Size()
		}
		return nil
	})
	return summe
}

// CacheLeeren ruft `go clean -cache`.
func (e *echt) CacheLeeren(ctx context.Context) error {
	if a := prozess(ctx, os.TempDir(), e.env, e.bauZeit, "go", "clean", "-cache"); a.Code != 0 {
		return os.ErrInvalid
	}
	return nil
}
