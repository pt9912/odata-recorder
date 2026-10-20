// Command pruefung führt die Prüfungen von make test und make build aus (SPEC-036):
// `pruefung test <Wurzel> [Muster...]` und `pruefung build <Binary> <Version>`.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	exitOK     = 0
	exitBefund = 1
	exitAufruf = 2

	aufrufText = "Aufruf: pruefung test <Wurzel> [Muster...]\n       pruefung build <Binary> <Version>\n"
	kommando   = "odata-recorder"
)

// Umgebung nennt die Programme, die eine Prüfung aufruft, als Name oder Pfad.
type Umgebung struct {
	Gofmt string
	Go    string
}

func standard() Umgebung {
	return Umgebung{Gofmt: "gofmt", Go: "go"}
}

func main() {
	os.Exit(run(standard(), os.Args[1:], os.Stdout, os.Stderr))
}

// lauf ist das Ergebnis eines Programmaufrufs, der gestartet werden konnte.
type lauf struct {
	stdout string
	stderr string
	code   int
}

// alles ist stdout gefolgt von stderr.
func (l lauf) alles() string {
	return l.stdout + l.stderr
}

// starte führt prog in dir aus (dir leer: im Arbeitsverzeichnis des Prozesses). Der Fehler
// ist gesetzt, wenn das Programm nicht gestartet werden konnte; ein Exit-Code ungleich 0
// ist keiner.
func starte(dir, prog string, args ...string) (lauf, error) {
	cmd := exec.CommandContext(context.Background(), prog, args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		return lauf{}, err
	}
	l := lauf{stdout: out.String(), stderr: errb.String()}
	if ee != nil {
		l.code = ee.ExitCode()
	}
	return l, nil
}

func run(u Umgebung, args []string, stdout, stderr io.Writer) int {
	if !aufrufGueltig(args) {
		_, _ = fmt.Fprint(stderr, aufrufText)
		return exitAufruf
	}
	if args[0] == "test" {
		return pruefeTest(u, args[1], args[2:], stdout, stderr)
	}
	return pruefeBuild(u, args[1], args[2], stdout, stderr)
}

// aufrufGueltig ist wahr für `test <Wurzel> [Muster...]` und `build <Binary> <Version>`
// ohne leeres Argument.
func aufrufGueltig(args []string) bool {
	if len(args) == 0 {
		return false
	}
	for _, a := range args {
		if a == "" {
			return false
		}
	}
	switch args[0] {
	case "test":
		return len(args) >= 2
	case "build":
		return len(args) == 3
	}
	return false
}

// nichtAufrufbar meldet das erste Programm, das der Prozess nicht findet.
func nichtAufrufbar(stderr io.Writer, progs ...string) bool {
	for _, p := range progs {
		if _, err := exec.LookPath(p); err != nil {
			_, _ = fmt.Fprintf(stderr, "%s nicht aufrufbar\n", p)
			return true
		}
	}
	return false
}

// schritt ist ein Programmaufruf der Prüfung `test` und der Text seines Befunds.
type schritt struct {
	prog   string
	args   []string
	befund string
}

func pruefeTest(u Umgebung, wurzel string, muster []string, stdout, stderr io.Writer) int {
	if fi, err := os.Stat(wurzel); err != nil || !fi.IsDir() {
		_, _ = fmt.Fprintf(stderr, "Wurzel fehlt: %s\n", wurzel)
		return exitAufruf
	}
	if nichtAufrufbar(stderr, u.Gofmt, u.Go) {
		return exitAufruf
	}
	if len(muster) == 0 {
		muster = []string{"./..."}
	}
	if code := gofmtPruefen(u, wurzel, stderr); code != exitOK {
		return code
	}
	folge := []schritt{
		{u.Go, append([]string{"vet"}, muster...), "go vet:"},
		{u.Go, append([]string{"test", "-trimpath", "-buildvcs=false"}, muster...), "go test:"},
	}
	for _, s := range folge {
		l, err := starte(wurzel, s.prog, s.args...)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "%s nicht aufrufbar\n", s.prog)
			return exitAufruf
		}
		if l.code != 0 {
			_, _ = fmt.Fprintf(stderr, "%s\n%s", s.befund, l.alles())
			return exitBefund
		}
	}
	_, _ = fmt.Fprintln(stdout, "pruefung test: ok")
	return exitOK
}

// gofmtPruefen meldet nicht formatierte Dateien und einen Fehler von gofmt selbst.
func gofmtPruefen(u Umgebung, wurzel string, stderr io.Writer) int {
	l, err := starte(wurzel, u.Gofmt, "-l", ".")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%s nicht aufrufbar\n", u.Gofmt)
		return exitAufruf
	}
	if l.code != 0 {
		_, _ = fmt.Fprintf(stderr, "gofmt:\n%s", l.alles())
		return exitBefund
	}
	if strings.TrimSpace(l.stdout) != "" {
		_, _ = fmt.Fprintf(stderr, "nicht gofmt-formatiert:\n%s", l.stdout)
		return exitBefund
	}
	return exitOK
}

func pruefeBuild(u Umgebung, binary, version string, stdout, stderr io.Writer) int {
	if _, err := os.Stat(binary); err != nil {
		_, _ = fmt.Fprintf(stderr, "Binary fehlt: %s\n", binary)
		return exitAufruf
	}
	if nichtAufrufbar(stderr, u.Go) {
		return exitAufruf
	}
	abs, err := filepath.Abs(binary)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Binary fehlt: %s\n", binary)
		return exitAufruf
	}
	if code := versionPruefen(abs, version, stderr); code != exitOK {
		return code
	}
	l, err := starte("", u.Go, "version", "-m", abs)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%s nicht aufrufbar\n", u.Go)
		return exitAufruf
	}
	if l.code != 0 {
		_, _ = fmt.Fprintf(stderr, "Build-Information des Binary nicht lesbar: %s", l.alles())
		return exitBefund
	}
	if !strings.Contains(l.stdout, "-trimpath=true") {
		_, _ = fmt.Fprintln(stderr, "das Binary ist nicht mit -trimpath gebaut")
		return exitBefund
	}
	_, _ = fmt.Fprintln(stdout, "pruefung build: ok")
	return exitOK
}

// versionPruefen ruft `<Binary> version`; nur stdout zählt, abschließende Zeilenumbrüche
// nicht.
func versionPruefen(binary, version string, stderr io.Writer) int {
	l, err := starte("", binary, "version")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Binary startet nicht: %v\n", err)
		return exitBefund
	}
	if l.code != 0 {
		_, _ = fmt.Fprintf(stderr, "Binary startet nicht: Exit %d\n%s", l.code, l.alles())
		return exitBefund
	}
	soll := kommando + " " + version
	if ist := strings.TrimRight(l.stdout, "\n"); ist != soll {
		_, _ = fmt.Fprintf(stderr, "Version des Binary ist '%s', erwartet '%s'\n", ist, soll)
		return exitBefund
	}
	return exitOK
}
