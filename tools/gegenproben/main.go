// Command gegenproben weist an a-check und an tools/harness/lint.sh nach, dass sie
// einen Verstoß rot und einen erlaubten Fall nicht rot melden (SPEC-102).
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	exitOK    = 0
	exitRot   = 1
	exitSetup = 2

	fallFrist = 120 * time.Second
)

// eingaben sind die Pfade, die jede Kopie enthält; go.sum ist nur da, wenn das Modul
// Abhängigkeiten hat.
var eingaben = []string{"src", "go.mod", "go.sum", ".a-check.yml", ".golangci.yml", "tools/harness/lint.sh"}

// programme müssen aufrufbar sein, bevor ein Fall läuft.
var programme = []string{"a-check", "golangci-lint", "sh"}

// befehl liefert Programm, Argumente und Arbeitsverzeichnis, mit denen ein Fall sein Gate
// in dir ruft.
var befehl = func(dir string, g Gate) (string, []string, string) {
	if g == gateACheck {
		return "a-check", []string{dir}, ""
	}
	return "sh", []string{"tools/harness/lint.sh"}, dir
}

// Ergebnis ist der Befund eines Falls; Rot leer heißt grün.
type Ergebnis struct {
	Name string
	Rot  string
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, ".", faelle()))
}

func run(stdout, stderr io.Writer, wurzel string, fs []Fall) int {
	for _, p := range programme {
		if _, err := exec.LookPath(p); err != nil {
			_, _ = fmt.Fprintf(stderr, "gegenproben: %s nicht aufrufbar\n", p)
			return exitSetup
		}
	}
	vorher, err := hashEingaben(wurzel)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "gegenproben: %v\n", err)
		return exitSetup
	}
	ergebnisse, err := fahre(fs, wurzel)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "gegenproben: %v\n", err)
		return exitSetup
	}
	nachher, err := hashEingaben(wurzel)
	if err != nil || nachher != vorher {
		_, _ = fmt.Fprintln(stderr, "gegenproben: die Eingaben haben sich im Lauf verändert")
		return exitSetup
	}
	rot := 0
	for _, e := range ergebnisse {
		if e.Rot == "" {
			_, _ = fmt.Fprintf(stdout, "ok   %s\n", e.Name)
			continue
		}
		rot++
		_, _ = fmt.Fprintf(stdout, "ROT  %s: %s\n", e.Name, e.Rot)
	}
	_, _ = fmt.Fprintf(stdout, "gegenproben: %d Fälle, %d rot\n", len(ergebnisse), rot)
	if rot > 0 {
		return exitRot
	}
	return exitOK
}

// fahre legt je Fall eine Kopie von wurzel an und fährt alle Fälle nebenläufig; ein
// roter Fall hält die übrigen nicht an. Die Reihenfolge der Ergebnisse ist die der Fälle.
func fahre(fs []Fall, wurzel string) ([]Ergebnis, error) {
	namen := map[string]bool{}
	for _, f := range fs {
		if namen[f.Name] {
			return nil, fmt.Errorf("Fall %s kommt mehrfach vor", f.Name)
		}
		namen[f.Name] = true
	}
	// Die Kopien entstehen vor dem ersten Gate-Aufruf, damit ein Fehler dort alle Fälle hält.
	basis, err := os.MkdirTemp("", "gegenproben-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(basis)
	dirs := make([]string, len(fs))
	for i, f := range fs {
		dirs[i] = filepath.Join(basis, fmt.Sprintf("%02d", i))
		if err := kopiere(wurzel, dirs[i]); err != nil {
			return nil, fmt.Errorf("Kopie für %s: %w", f.Name, err)
		}
		if err := greife(dirs[i], f); err != nil {
			return nil, fmt.Errorf("Eingriff für %s: %w", f.Name, err)
		}
	}
	erg := make([]Ergebnis, len(fs))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i := range fs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			name, args, wd := befehl(dirs[i], fs[i].Gate)
			code, out, abgelaufen := starte(fallFrist, wd, name, args...)
			erg[i] = Ergebnis{Name: fs[i].Name, Rot: bewerte(fs[i], code, out, abgelaufen)}
		}()
	}
	wg.Wait()
	return erg, nil
}

// bewerte gibt den Grund zurück, aus dem ein Fall rot ist, oder "".
func bewerte(f Fall, code int, out string, abgelaufen bool) string {
	switch {
	case abgelaufen:
		return "Zeit überschritten"
	case f.WantExit == 0 && code != 0:
		return fmt.Sprintf("erlaubter Fall abgelehnt (Exit %d)\n%s", code, letzte(out, 8))
	case f.WantExit != 0 && code == 0:
		return "Verstoß nicht gemeldet"
	case code != f.WantExit:
		return fmt.Sprintf("erwartet Exit %d, war %d\n%s", f.WantExit, code, letzte(out, 8))
	case f.WantText != "" && !strings.Contains(out, f.WantText):
		return fmt.Sprintf("Ausgabe nennt %q nicht\n%s", f.WantText, letzte(out, 8))
	}
	return ""
}

func letzte(s string, n int) string {
	z := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(z) > n {
		z = z[len(z)-n:]
	}
	return strings.Join(z, "\n")
}

// starte führt das Programm aus und liefert Exit-Code, Ausgabe und ob die Frist ablief.
// Bei Ablauf endet die ganze Prozessgruppe, damit auch ein Kindprozess die Ausgabe nicht
// offen hält.
func starte(frist time.Duration, dir, name string, args ...string) (int, string, bool) {
	ctx, stop := context.WithTimeout(context.Background(), frist)
	defer stop()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 5 * time.Second
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	err := cmd.Run()
	if ctx.Err() != nil {
		return -1, buf.String(), true
	}
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0, buf.String(), false
	case errors.As(err, &ee):
		return ee.ExitCode(), buf.String(), false
	default:
		return -1, err.Error(), false
	}
}

func kopiere(von, nach string) error {
	for _, p := range eingaben {
		err := filepath.WalkDir(filepath.Join(von, p), func(pfad string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(von, pfad)
			if err != nil {
				return err
			}
			ziel := filepath.Join(nach, rel)
			if d.IsDir() {
				return os.MkdirAll(ziel, 0o755)
			}
			b, err := os.ReadFile(pfad)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(ziel), 0o755); err != nil {
				return err
			}
			return os.WriteFile(ziel, b, 0o644)
		})
		if err != nil && !(p == "go.sum" && errors.Is(err, fs.ErrNotExist)) {
			return err
		}
	}
	return nil
}

func greife(dir string, f Fall) error {
	for d, inhalt := range f.Neu {
		p := filepath.Join(dir, d)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(inhalt), 0o644); err != nil {
			return err
		}
	}
	for d, text := range f.Anhang {
		h, err := os.OpenFile(filepath.Join(dir, d), os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		_, werr := h.WriteString(text)
		if cerr := h.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return werr
		}
	}
	for _, d := range f.Entferne {
		if err := os.Remove(filepath.Join(dir, d)); err != nil {
			return err
		}
	}
	for _, e := range f.Ersetze {
		p := filepath.Join(dir, e.Datei)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.Count(string(b), e.Alt) != 1 {
			return fmt.Errorf("%s: die Stelle %q kommt nicht genau einmal vor", e.Datei, e.Alt)
		}
		if err := os.WriteFile(p, []byte(strings.Replace(string(b), e.Alt, e.Neu, 1)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func hashEingaben(wurzel string) (string, error) {
	var pfade []string
	for _, p := range eingaben {
		err := filepath.WalkDir(filepath.Join(wurzel, p), func(pfad string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				pfade = append(pfade, pfad)
			}
			return err
		})
		if err != nil && !(p == "go.sum" && errors.Is(err, fs.ErrNotExist)) {
			return "", err
		}
	}
	sort.Strings(pfade)
	h := sha256.New()
	for _, p := range pfade {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%x\n", p, sha256.Sum256(b))
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
