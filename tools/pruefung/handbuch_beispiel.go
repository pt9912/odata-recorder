package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"
)

// ergebnis ist der Lauf des Binary: beide Ströme getrennt, der Ausgangswert, ob ein Signal den
// Prozess beendete, und ob die Zeitgrenze ihn beendete (SPEC-200 Nr. 7).
type ergebnis struct {
	stdout, stderr string
	code           int
	signal, zeit   bool
}

// arbeitFehler ist der Fehler, das leere Verzeichnis eines Laufs anzulegen (SPEC-200 Nr. 2 c).
type arbeitFehler struct {
	err error
}

func (a *arbeitFehler) Error() string {
	return a.err.Error()
}

// eingabeText ist der Eingabefehler zu einem Fehler von fahre: das Verzeichnis des Laufs ist
// `Arbeitsverzeichnis nicht anlegbar`, jeder andere Fehler `Binary nicht startbar` (SPEC-200 Nr. 2 c, 7 a).
func eingabeText(err error) string {
	var a *arbeitFehler
	if errors.As(err, &a) {
		return "Arbeitsverzeichnis nicht anlegbar: " + a.Error()
	}
	return "Binary nicht startbar: " + err.Error()
}

// fahre startet bin mit args ohne Shell, in leerer Umgebung, in einem leeren Verzeichnis unter
// arbeit und mit leerer Standardeingabe. Der Fehler ist gesetzt, wenn der Prozess nicht gestartet
// werden konnte oder das Verzeichnis des Laufs nicht anzulegen war (arbeitFehler; SPEC-200 Nr. 7 a, b).
func fahre(bin string, args []string, arbeit string, grenze time.Duration) (ergebnis, error) {
	dir, err := os.MkdirTemp(arbeit, "lauf-")
	if err != nil {
		return ergebnis{}, &arbeitFehler{err}
	}
	ctx, cancel := context.WithTimeout(context.Background(), grenze)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = []string{}
	cmd.WaitDelay = time.Second
	// Die Zeitgrenze gilt dem Binary, nicht dem Warten auf die Ströme (SPEC-200 Nr. 7 c, 7 e): sie
	// ist überschritten, wenn das Binary beim Ablauf noch lebte und beendet wurde; Cancel läuft
	// nur dann, und nur dann ist gekillt gesetzt.
	var gekillt atomic.Bool
	cmd.Cancel = func() error {
		err := cmd.Process.Kill()
		if err == nil {
			gekillt.Store(true)
		}
		return err
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err = cmd.Run()
	r := ergebnis{stdout: out.String(), stderr: errb.String()}
	if gekillt.Load() {
		r.zeit = true
		return r, nil
	}
	var ee *exec.ExitError
	switch {
	case err == nil, errors.Is(err, exec.ErrWaitDelay):
	case errors.As(err, &ee):
		r.code = ee.ExitCode()
		ws, ok := ee.Sys().(syscall.WaitStatus)
		r.signal = ok && ws.Signaled()
	default:
		return ergebnis{}, err
	}
	return r, nil
}

func regexpBindung() *regexp.Regexp {
	return regexp.MustCompile(`^Datei: \[(beispiele/` + beispielName + `)\]\((beispiele/` + beispielName + `)\)$`)
}

func regexpName() *regexp.Regexp {
	return regexp.MustCompile(`^` + beispielName + `$`)
}

// beispiel ist der Inhalt einer Beispiel-Datei (SPEC-200 Nr. 6).
type beispiel struct {
	args           []string
	stdout, stderr string
	code           int
}

// aufrufPraefix ist die Kopfzeile jeder Beispiel-Datei bis zu den Argumenten.
const aufrufPraefix = "$ docker run --rm odata-recorder:build /out/odata-recorder"

func trennzeilen() []string {
	return []string{"--- Standardausgabe", "--- Fehlerausgabe", "--- Ausgangswert"}
}

// leseBeispiel liest eine Beispiel-Datei; der zweite Wert ist der Grund eines Verstoßes.
func leseBeispiel(b []byte) (beispiel, string) {
	s := string(b)
	switch {
	case !utf8.Valid(b):
		return beispiel{}, "kein UTF-8"
	case strings.Contains(s, "\r"):
		return beispiel{}, "enthält \\r"
	case !strings.HasSuffix(s, "\n"):
		return beispiel{}, "endet nicht mit einem Zeilenumbruch"
	}
	zeilen := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	var bsp beispiel
	var grund string
	if bsp.args, grund = kopfzeile(zeilen[0]); grund != "" {
		return beispiel{}, grund
	}
	abschnitte, grund := schneide(zeilen[1:])
	if grund != "" {
		return beispiel{}, grund
	}
	bsp.stdout, bsp.stderr = text(abschnitte[0]), text(abschnitte[1])
	if bsp.code, grund = ausgangswert(abschnitte[2]); grund != "" {
		return beispiel{}, grund
	}
	return bsp, ""
}

// kopfzeile liest das wörtliche Präfix und die Argumente (SPEC-200 Nr. 6 a).
func kopfzeile(z string) ([]string, string) {
	rest, ok := strings.CutPrefix(z, aufrufPraefix)
	if !ok || (rest != "" && rest[0] != ' ') {
		return nil, "Kopfzeile: Präfix"
	}
	if rest == "" {
		return nil, ""
	}
	args := strings.Split(rest[1:], " ")
	for _, a := range args {
		if a == "" {
			return nil, "Kopfzeile: leeres Argument"
		}
		for _, c := range []byte(a) {
			if !argumentZeichen(c) {
				return nil, fmt.Sprintf("Kopfzeile: Zeichen %q im Argument", c)
			}
		}
	}
	return args, ""
}

// argumentZeichen: A-Z a-z 0-9 . _ = - (SPEC-200 Nr. 6 a).
func argumentZeichen(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.IndexByte("._=-", c) >= 0
}

// schneide trennt die Zeilen nach der Kopfzeile in die Zeilen der drei Abschnitte; die drei
// Trennzeilen stehen je einmal in der Reihenfolge Standardausgabe, Fehlerausgabe, Ausgangswert,
// die erste unmittelbar nach der Kopfzeile (SPEC-200 Nr. 6 b).
func schneide(zeilen []string) ([3][]string, string) {
	var pos [3]int
	for k, t := range trennzeilen() {
		pos[k] = -1
		for i, z := range zeilen {
			if z != t {
				continue
			}
			if pos[k] >= 0 {
				return [3][]string{}, "Trennzeile doppelt: " + t
			}
			pos[k] = i
		}
		if pos[k] < 0 {
			return [3][]string{}, "Trennzeile fehlt: " + t
		}
	}
	if pos[0] >= pos[1] || pos[1] >= pos[2] {
		return [3][]string{}, "Trennzeilen nicht in der Reihenfolge Standardausgabe, Fehlerausgabe, Ausgangswert"
	}
	if pos[0] != 0 {
		return [3][]string{}, "Zeilen zwischen Kopfzeile und erster Trennzeile"
	}
	return [3][]string{zeilen[pos[0]+1 : pos[1]], zeilen[pos[1]+1 : pos[2]], zeilen[pos[2]+1:]}, ""
}

// text hängt an jede Zeile einen Zeilenumbruch.
func text(zeilen []string) string {
	var b strings.Builder
	for _, z := range zeilen {
		b.WriteString(z + "\n")
	}
	return b.String()
}

// ausgangswert liest genau eine Zeile mit einer Dezimalzahl von 0 bis 255 ohne führende Null.
func ausgangswert(zeilen []string) (int, string) {
	if len(zeilen) != 1 {
		return 0, "Ausgangswert: genau eine Zeile erwartet"
	}
	z := zeilen[0]
	n := ziffern(z)
	if z == "" || n < 0 || len(z) > 3 || (len(z) > 1 && z[0] == '0') || n > 255 {
		return 0, "Ausgangswert: " + strconv.Quote(z)
	}
	return n, ""
}

// pruefeBeispiele fährt jede Datei, auf die eine gültige Bindung zeigt und die vorhanden und
// formgültig ist, einmal und vergleicht beide Ströme und den Ausgangswert (SPEC-200 Nr. 6, 7).
func pruefeBeispiele(u Umgebung, bin, wurzel, arbeit string, bindungen []string) ([]befund, []string) {
	namen := map[string]bool{}
	for _, p := range bindungen {
		namen[path.Base(p)] = true
	}
	sortiert := make([]string, 0, len(namen))
	for n := range namen {
		sortiert = append(sortiert, n)
	}
	sort.Strings(sortiert)
	var bf []befund
	for _, name := range sortiert {
		inhalt, err := os.ReadFile(filepath.Join(wurzel, beispielVerz, name))
		if err != nil {
			continue
		}
		bsp, grund := leseBeispiel(inhalt)
		if grund != "" {
			bf = append(bf, befund{nr: nrBeispiel, name: name, text: fmt.Sprintf("%s: Beispiel-Datei ungültig: %s", name, grund)})
			continue
		}
		l, err := fahre(bin, bsp.args, arbeit, u.Zeitgrenze)
		if err != nil {
			return nil, []string{eingabeText(err)}
		}
		bf = append(bf, vergleiche(name, bsp, l)...)
	}
	return bf, nil
}

// vergleiche hält beide Ströme und den Ausgangswert je für sich gegen die Datei (SPEC-200 Nr. 7 c, d).
func vergleiche(name string, bsp beispiel, l ergebnis) []befund {
	if l.zeit {
		return []befund{{nr: nrLauf, name: name, text: name + ": Zeitgrenze überschritten"}}
	}
	var bf []befund
	abw := func(was, soll, ist string) {
		bf = append(bf, befund{nr: nrLauf, name: name, text: fmt.Sprintf("%s: %s weicht ab (erwartet %s, erhalten %s)", name, was, kuerze(soll), kuerze(ist))})
	}
	if l.stdout != bsp.stdout {
		abw("Standardausgabe", strconv.Quote(bsp.stdout), strconv.Quote(l.stdout))
	}
	if l.stderr != bsp.stderr {
		abw("Fehlerausgabe", strconv.Quote(bsp.stderr), strconv.Quote(l.stderr))
	}
	if l.signal {
		abw("Ausgangswert", strconv.Itoa(bsp.code), "Signal")
	} else if l.code != bsp.code {
		abw("Ausgangswert", strconv.Itoa(bsp.code), strconv.Itoa(l.code))
	}
	return bf
}

// kuerze schneidet die Zeichenkette in Go-Schreibweise auf 200 Zeichen.
func kuerze(q string) string {
	if r := []rune(q); len(r) > 200 {
		return string(r[:200])
	}
	return q
}
