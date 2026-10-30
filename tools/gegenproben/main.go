// Command gegenproben weist an a-check, golangci-lint (tools/harness/lint.sh),
// tools/harness/migration-check.sh und pruefung (test, build, race) nach, dass sie einen Verstoß
// rot und einen erlaubten Fall nicht rot melden (SPEC-102). Die Verdrahtung von pruefung im
// Dockerfile prüft es nicht.
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

// Umgebung bündelt, was ein Lauf von außen braucht: die Pfade, die jede Kopie enthält
// (go.sum nur, wenn das Modul Abhängigkeiten hat), die Programme, die aufrufbar sein
// müssen, und den Aufruf des Gates eines Falls.
type Umgebung struct {
	Eingaben  []string
	Programme []string
	// Aufraeumen entfernt das Wegwerfverzeichnis eines Laufs.
	Aufraeumen func(string) error
	// Befehl liefert Programm, Argumente und Arbeitsverzeichnis, mit denen ein Fall sein
	// Gate in dir ruft.
	Befehl func(dir string, f Fall) (string, []string, string)
	// Bauen liefert Programm und Argumente, mit denen ein Fall mit Bau sein Binary in dir
	// erzeugt (Arbeitsverzeichnis dir).
	Bauen func(dir string, b Bau) (string, []string)
	// Treiber ist der Pfad von tools/harness/groessenmarke.sh, den die Fälle 103 bis 120 mit
	// `sh` rufen.
	Treiber string
}

func standard() Umgebung {
	treiber, err := filepath.Abs("tools/harness/groessenmarke.sh")
	if err != nil {
		treiber = "tools/harness/groessenmarke.sh"
	}
	return Umgebung{
		Eingaben:   []string{"src", "go.mod", "go.sum", ".a-check.yml", ".golangci.yml", "tools/harness/lint.sh", "tools/harness/migration-check.sh"},
		Programme:  []string{"a-check", "golangci-lint", "sh", "pruefung", "go", "gofmt", "groessenmarke", "git"},
		Aufraeumen: os.RemoveAll,
		Befehl:     befehlMit(treiber),
		Bauen:      bauen,
		Treiber:    treiber,
	}
}

// befehlMit liefert den Aufruf der Gates; die Fälle zu groessenmarke rufen das Programm bzw. den
// Treiber unter dem Pfad treiber, jeweils im Verzeichnis des Falls.
func befehlMit(treiber string) func(dir string, f Fall) (string, []string, string) {
	return func(dir string, f Fall) (string, []string, string) {
		switch f.Gate {
		case gateGroessenmarke:
			return "groessenmarke", f.ersetzeOrte(dir), dir
		case gateTreiber:
			return "sh", append([]string{treiber}, f.ersetzeOrte(dir)...), dir
		}
		return befehl(dir, f)
	}
}

// befehl ruft das Gate des Falls f im Verzeichnis dir des Falls.
func befehl(dir string, f Fall) (string, []string, string) {
	switch f.Gate {
	case gateACheck:
		return "a-check", []string{dir}, ""
	case gateMigration:
		args := []string{filepath.Join(dir, erzeugtesDDL), filepath.Join(dir, migrationen)}
		if f.Argumente != nil {
			args = f.Argumente
		}
		return "sh", append([]string{"tools/harness/migration-check.sh"}, args...), dir
	case gatePruefung:
		return "pruefung", f.ersetzeOrte(dir), ""
	}
	return "sh", []string{"tools/harness/lint.sh"}, dir
}

// bauen erzeugt das Binary eines Falls mit go build in der Kopie: Flags des Dockerfile, Stufe
// build, ohne Zielplattform.
func bauen(dir string, b Bau) (string, []string) {
	args := []string{"build", "-buildvcs=false"}
	if !b.OhneTrimpath {
		args = append(args, "-trimpath")
	}
	args = append(args, "-ldflags=-s -w -X main.version="+b.Version, "-o", filepath.Join(dir, binaerName), "./src/bootstrap/main")
	return "go", args
}

// Ergebnis ist der Befund eines Falls; Rot leer heißt grün.
type Ergebnis struct {
	Name string
	Rot  string
}

func main() {
	os.Exit(run(standard(), os.Stdout, os.Stderr, ".", faelle()))
}

func run(u Umgebung, stdout, stderr io.Writer, wurzel string, fs []Fall) int {
	for _, p := range u.Programme {
		if _, err := exec.LookPath(p); err != nil {
			_, _ = fmt.Fprintf(stderr, "gegenproben: %s nicht aufrufbar\n", p)
			return exitSetup
		}
	}
	vorher, err := hashEingaben(u.Eingaben, wurzel)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "gegenproben: %v\n", err)
		return exitSetup
	}
	ergebnisse, fehler := fahre(u, fs, wurzel)
	if ergebnisse == nil && fehler != nil {
		_, _ = fmt.Fprintf(stderr, "gegenproben: %v\n", fehler)
		return exitSetup
	}
	nachher, err := hashEingaben(u.Eingaben, wurzel)
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
	if fehler != nil {
		_, _ = fmt.Fprintf(stderr, "gegenproben: %v\n", fehler)
		return exitSetup
	}
	if rot > 0 {
		return exitRot
	}
	return exitOK
}

// fahre legt je Fall sein Verzeichnis an (eine Kopie von wurzel, bei einem Wegwerfmodul ein leeres
// Verzeichnis) und fährt alle Fälle nebenläufig; ein roter Fall hält die übrigen nicht an. Die
// Reihenfolge der Ergebnisse ist die der Fälle.
func fahre(u Umgebung, fs []Fall, wurzel string) (erg []Ergebnis, err error) {
	namen := map[string]bool{}
	for _, f := range fs {
		if namen[f.Name] {
			return nil, fmt.Errorf("Fall %s kommt mehrfach vor", f.Name)
		}
		namen[f.Name] = true
	}
	// Die Verzeichnisse entstehen vor dem ersten Gate-Aufruf, damit ein Fehler dort alle Fälle hält.
	basis, err := os.MkdirTemp("", "gegenproben-")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, u.Aufraeumen(basis)) }()
	dirs := make([]string, len(fs))
	for i, f := range fs {
		dirs[i] = filepath.Join(basis, fmt.Sprintf("%02d", i))
		if err := legeVerzeichnisAn(u, f, wurzel, dirs[i]); err != nil {
			return nil, fmt.Errorf("Kopie für %s: %w", f.Name, err)
		}
		if err := greife(dirs[i], f); err != nil {
			return nil, fmt.Errorf("Eingriff für %s: %w", f.Name, err)
		}
	}
	tmps := make([]string, len(fs))
	for i := range fs {
		tmps[i] = dirs[i] + ".tmp"
		if err := os.MkdirAll(tmps[i], 0o755); err != nil {
			return nil, err
		}
	}
	erg = make([]Ergebnis, len(fs))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i := range fs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			erg[i] = fuehreFall(u, fs[i], dirs[i], tmps[i])
		}()
	}
	wg.Wait()
	return erg, nil
}

// legeVerzeichnisAn legt dir an: die Kopie der Eingaben, bei einem Wegwerfmodul ein leeres
// Verzeichnis.
func legeVerzeichnisAn(u Umgebung, f Fall, wurzel, dir string) error {
	if f.Wegwerfmodul {
		return os.MkdirAll(dir, 0o755)
	}
	return kopiere(u.Eingaben, wurzel, dir)
}

// fuehreFall erzeugt bei einem Fall mit Bau das Binary, ruft das Gate im Verzeichnis dir des Falls
// und bewertet das Ergebnis.
func fuehreFall(u Umgebung, f Fall, dir, tmp string) Ergebnis {
	// golangci-lint sperrt eine Datei in TMPDIR gegen parallele Läufe; jeder Fall hat sein eigenes.
	env := f.umgebung(dir, tmp)
	if f.Bau != nil {
		name, args := u.Bauen(dir, *f.Bau)
		code, out, abgelaufen := starte(fallFrist, dir, env, name, args...)
		if abgelaufen || code != 0 {
			return Ergebnis{Name: f.Name, Rot: "Bau des Binary scheitert\n" + letzte(out)}
		}
	}
	name, args, wd := u.Befehl(dir, f)
	code, out, abgelaufen := starte(fallFrist, wd, env, name, args...)
	rot := bewerte(f, code, out, abgelaufen)
	if rot == "" && f.Nachher != nil {
		rot = f.Nachher(Lauf{U: u, Dir: dir, Tmp: tmp, Env: env, Out: out})
	}
	return Ergebnis{Name: f.Name, Rot: rot}
}

// bewerte gibt den Grund zurück, aus dem ein Fall rot ist, oder "".
func bewerte(f Fall, code int, out string, abgelaufen bool) string {
	switch {
	case abgelaufen:
		return "Zeit überschritten"
	case f.WantExit == 0 && code != 0:
		return fmt.Sprintf("erlaubter Fall abgelehnt (Exit %d)\n%s", code, letzte(out))
	case f.WantExit != 0 && code == 0:
		return "Verstoß nicht gemeldet"
	case code != f.WantExit:
		return fmt.Sprintf("erwartet Exit %d, war %d\n%s", f.WantExit, code, letzte(out))
	}
	for _, w := range f.WantText {
		if !strings.Contains(out, w) {
			return fmt.Sprintf("Ausgabe nennt %q nicht\n%s", w, letzte(out))
		}
	}
	for _, v := range f.Verboten {
		if strings.Contains(out, v) {
			return fmt.Sprintf("Ausgabe nennt %q\n%s", v, letzte(out))
		}
	}
	return ""
}

func letzte(s string) string {
	const n = 8
	z := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(z) > n {
		z = z[len(z)-n:]
	}
	return strings.Join(z, "\n")
}

// starte führt das Programm aus und liefert Exit-Code, Ausgabe und ob die Frist ablief.
// Bei Ablauf endet die ganze Prozessgruppe, damit auch ein Kindprozess die Ausgabe nicht
// offen hält.
func starte(frist time.Duration, dir string, env []string, name string, args ...string) (int, string, bool) {
	ctx, stop := context.WithTimeout(context.Background(), frist)
	defer stop()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
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

// ohneFehlendesGoSum lässt das Fehlen von go.sum durch; jeden anderen Fehler gibt es zurück.
func ohneFehlendesGoSum(pfad string, err error) error {
	if pfad == "go.sum" && errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func kopiere(eingaben []string, von, nach string) error {
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
		if err = ohneFehlendesGoSum(p, err); err != nil {
			return err
		}
	}
	return nil
}

func greife(dir string, f Fall) error {
	for _, schritt := range []func(string, Fall) error{kopiereDateien, legeAn, haengeAn, entferne, ersetze, macheAusfuehrbar, bereite} {
		if err := schritt(dir, f); err != nil {
			return err
		}
	}
	return nil
}

func kopiereDateien(dir string, f Fall) error {
	for ziel, quelle := range f.Kopiere {
		b, err := os.ReadFile(filepath.Join(dir, quelle))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, ziel), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func legeAn(dir string, f Fall) error {
	for d, inhalt := range f.Neu {
		p := filepath.Join(dir, d)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(inhalt), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func haengeAn(dir string, f Fall) error {
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
	return nil
}

func entferne(dir string, f Fall) error {
	if f.OhneGoDateienInSrc {
		err := filepath.WalkDir(filepath.Join(dir, "src"), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
				return err
			}
			return os.Remove(p)
		})
		if err != nil {
			return err
		}
	}
	for _, d := range f.Entferne {
		if err := os.Remove(filepath.Join(dir, d)); err != nil {
			return err
		}
	}
	return nil
}

func ersetze(dir string, f Fall) error {
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

func macheAusfuehrbar(dir string, f Fall) error {
	for _, d := range f.Ausfuehrbar {
		if err := os.Chmod(filepath.Join(dir, d), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func bereite(dir string, f Fall) error {
	if f.Vorbereitung == nil {
		return nil
	}
	return f.Vorbereitung(dir)
}

func hashEingaben(eingaben []string, wurzel string) (string, error) {
	var pfade []string
	for _, p := range eingaben {
		err := filepath.WalkDir(filepath.Join(wurzel, p), func(pfad string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				pfade = append(pfade, pfad)
			}
			return err
		})
		if err = ohneFehlendesGoSum(p, err); err != nil {
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
