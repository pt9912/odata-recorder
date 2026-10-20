package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fake legt ein Shell-Skript als Programm an und liefert seinen Pfad.
func fake(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// werkzeuge legt gofmt und go als Skripte an; beide schreiben ihre Argumente je Aufruf als
// eine Zeile in die Datei log. Der Rumpf entscheidet über Ausgabe und Exit-Code.
type werkzeuge struct {
	u   Umgebung
	log string
}

func neueWerkzeuge(t *testing.T, gofmtRumpf, goRumpf string) werkzeuge {
	t.Helper()
	d := t.TempDir()
	log := filepath.Join(d, "log")
	return werkzeuge{
		u: Umgebung{
			Gofmt: fake(t, d, "gofmt", `echo "gofmt $*" >> `+log+"\n"+gofmtRumpf),
			Go:    fake(t, d, "go", `echo "go $*" >> `+log+"\n"+goRumpf),
		},
		log: log,
	}
}

// aufrufe liefert die Zeilen des Logs; ohne Datei wurde nichts gestartet.
func (w werkzeuge) aufrufe(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(w.log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(string(b), "\n"), "\n")
}

func fuehre(u Umgebung, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(u, args, &out, &errb)
	return code, out.String(), errb.String()
}

func erwarte(t *testing.T, code int, errText string, wantCode int, wantText string) {
	t.Helper()
	if code != wantCode || !strings.Contains(errText, wantText) {
		t.Fatalf("Exit %d (erwartet %d), Ausgabe %q (erwartet Text %q)", code, wantCode, errText, wantText)
	}
}

func TestAufrufFehlerStartenNichts(t *testing.T) {
	faelle := [][]string{
		nil,
		{"unbekannt"},
		{"test"},
		{"test", ""},
		{"test", ".", ""},
		{"build"},
		{"build", "b"},
		{"build", "b", "v", "x"},
		{"build", "", "v"},
		{"build", "b", ""},
		{""},
	}
	for _, args := range faelle {
		t.Run(strings.Join(args, "|"), func(t *testing.T) {
			w := neueWerkzeuge(t, "", "")
			code, _, errText := fuehre(w.u, args...)
			erwarte(t, code, errText, exitAufruf, "Aufruf: pruefung test")
			if !strings.Contains(errText, "pruefung build <Binary> <Version>") {
				t.Fatalf("Aufruf nennt die Form build nicht: %q", errText)
			}
			if got := w.aufrufe(t); got != nil {
				t.Fatalf("es wurde gestartet: %v", got)
			}
		})
	}
}

func TestWurzelFehlt(t *testing.T) {
	datei := filepath.Join(t.TempDir(), "datei")
	if err := os.WriteFile(datei, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, wurzel := range []string{filepath.Join(t.TempDir(), "gibt-es-nicht"), datei} {
		w := neueWerkzeuge(t, "", "")
		code, _, errText := fuehre(w.u, "test", wurzel)
		erwarte(t, code, errText, exitAufruf, "Wurzel fehlt: "+wurzel)
		if got := w.aufrufe(t); got != nil {
			t.Fatalf("es wurde gestartet: %v", got)
		}
	}
}

func TestProgrammNichtAufrufbar(t *testing.T) {
	d := t.TempDir()
	fehlt := filepath.Join(d, "gibt-es-nicht")
	bin := fake(t, d, "bin", `echo "odata-recorder v"`)
	tests := []struct {
		name  string
		gofmt bool // gofmt ist aufrufbar
		args  []string
		prog  string
	}{
		{"test ohne gofmt", false, []string{"test", d}, "gofmt"},
		{"test ohne go", true, []string{"test", d}, "go"},
		{"build ohne go", true, []string{"build", bin, "v"}, "go"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := neueWerkzeuge(t, "", "")
			if tc.prog == "gofmt" {
				w.u.Gofmt = fehlt
			} else {
				w.u.Go = fehlt
			}
			code, _, errText := fuehre(w.u, tc.args...)
			erwarte(t, code, errText, exitAufruf, fehlt+" nicht aufrufbar")
			if got := w.aufrufe(t); got != nil {
				t.Fatalf("es wurde gestartet: %v", got)
			}
		})
	}
}

func TestTestOhneBefund(t *testing.T) {
	w := neueWerkzeuge(t, "", "")
	wurzel := t.TempDir()
	code, out, errText := fuehre(w.u, "test", wurzel)
	if code != exitOK || !strings.Contains(out, "pruefung test: ok") || errText != "" {
		t.Fatalf("Exit %d, stdout %q, stderr %q", code, out, errText)
	}
	want := []string{"gofmt -l .", "go vet ./...", "go test -trimpath -buildvcs=false ./..."}
	if got := w.aufrufe(t); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("Aufrufe %v, erwartet %v", got, want)
	}
}

func TestTestMusterWerdenWeitergegeben(t *testing.T) {
	w := neueWerkzeuge(t, "", "")
	code, _, errText := fuehre(w.u, "test", t.TempDir(), "./a/...", "./b")
	if code != exitOK {
		t.Fatalf("Exit %d: %s", code, errText)
	}
	want := []string{"gofmt -l .", "go vet ./a/... ./b", "go test -trimpath -buildvcs=false ./a/... ./b"}
	if got := w.aufrufe(t); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("Aufrufe %v, erwartet %v", got, want)
	}
}

func TestTestBefundHaeltAnDerStelle(t *testing.T) {
	tests := []struct {
		name           string
		gofmt, goRumpf string
		text           string
		aufrufe        int // Zahl der gestarteten Programme bis zum Befund
	}{
		{"nicht formatiert", `echo a.go; echo b.go`, "", "nicht gofmt-formatiert:\na.go\nb.go", 1},
		{"gofmt scheitert", `echo "x.go:1:1: expected" >&2; exit 2`, "", "gofmt:\nx.go:1:1: expected", 1},
		{"vet", "", `[ "$1" = vet ] && { echo "Printf-Befund"; exit 1; }; exit 0`, "go vet:\nPrintf-Befund", 2},
		{"test", "", `[ "$1" = test ] && { echo "FAIL x"; exit 1; }; exit 0`, "go test:\nFAIL x", 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := neueWerkzeuge(t, tc.gofmt, tc.goRumpf)
			code, out, errText := fuehre(w.u, "test", t.TempDir())
			erwarte(t, code, errText, exitBefund, tc.text)
			if strings.Contains(out, "ok") {
				t.Fatalf("stdout meldet ok: %q", out)
			}
			if got := len(w.aufrufe(t)); got != tc.aufrufe {
				t.Fatalf("%d Aufrufe, erwartet %d: %v", got, tc.aufrufe, w.aufrufe(t))
			}
		})
	}
}

// Ein Verzeichnis ohne Go-Dateien ist für pruefung kein Befund: es entscheiden die Exit-Codes
// von go vet und go test.
func TestTestOhneGoDateienEntscheidenDieProgramme(t *testing.T) {
	wurzel := t.TempDir()
	w := neueWerkzeuge(t, "", "")
	code, out, errText := fuehre(w.u, "test", wurzel)
	if code != exitOK || !strings.Contains(out, "pruefung test: ok") {
		t.Fatalf("Exit %d, stdout %q, stderr %q", code, out, errText)
	}
	w = neueWerkzeuge(t, "", `echo "no packages to vet"; exit 1`)
	code, _, errText = fuehre(w.u, "test", wurzel)
	erwarte(t, code, errText, exitBefund, "go vet:\nno packages to vet")
}

// binaerWerkzeuge liefert ein Binary, das `version` mit dem Rumpf beantwortet, und ein go,
// das auf `version -m` mit dem Rumpf antwortet.
func binaerWerkzeuge(t *testing.T, binRumpf, goRumpf string) (werkzeuge, string) {
	t.Helper()
	w := neueWerkzeuge(t, "", goRumpf)
	return w, fake(t, t.TempDir(), "odata-recorder", binRumpf)
}

const trimpathOK = `echo "	build	-trimpath=true"`

func TestBuildOhneBefund(t *testing.T) {
	w, bin := binaerWerkzeuge(t, `echo "odata-recorder 1.2.3"; echo; echo`, trimpathOK)
	code, out, errText := fuehre(w.u, "build", bin, "1.2.3")
	if code != exitOK || !strings.Contains(out, "pruefung build: ok") {
		t.Fatalf("Exit %d, stdout %q, stderr %q", code, out, errText)
	}
	if got := w.aufrufe(t); len(got) != 1 || !strings.HasPrefix(got[0], "go version -m ") {
		t.Fatalf("Aufrufe %v", got)
	}
}

func TestBuildBinaryFehlt(t *testing.T) {
	w := neueWerkzeuge(t, "", "")
	fehlt := filepath.Join(t.TempDir(), "odata-recorder")
	code, _, errText := fuehre(w.u, "build", fehlt, "v")
	erwarte(t, code, errText, exitAufruf, "Binary fehlt: "+fehlt)
	if got := w.aufrufe(t); got != nil {
		t.Fatalf("es wurde gestartet: %v", got)
	}
}

func TestBuildVersionBefund(t *testing.T) {
	tests := []struct {
		name string
		bin  string
		text string
	}{
		{"andere Version", `echo "odata-recorder 9"`, "Version des Binary ist 'odata-recorder 9', erwartet 'odata-recorder 1.2.3'"},
		{"Zusatztext", `echo "odata-recorder 1.2.3 extra"`, "Version des Binary ist 'odata-recorder 1.2.3 extra'"},
		{"Zusatzzeile", `echo "odata-recorder 1.2.3"; echo "mehr"`, "erwartet 'odata-recorder 1.2.3'"},
		{"nur stderr", `echo "odata-recorder 1.2.3" >&2`, "Version des Binary ist '', erwartet"},
		{"ohne Namen", `echo "1.2.3"`, "Version des Binary ist '1.2.3'"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, bin := binaerWerkzeuge(t, tc.bin, trimpathOK)
			code, _, errText := fuehre(w.u, "build", bin, "1.2.3")
			erwarte(t, code, errText, exitBefund, tc.text)
			if got := w.aufrufe(t); got != nil {
				t.Fatalf("go lief trotz Versionsbefund: %v", got)
			}
		})
	}
}

func TestBuildBinaryStartetNicht(t *testing.T) {
	d := t.TempDir()
	keinExec := filepath.Join(d, "kein-exec")
	if err := os.WriteFile(keinExec, []byte("#!/bin/sh\necho x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	keinFormat := filepath.Join(d, "kein-format")
	if err := os.WriteFile(keinFormat, []byte("kein Programm\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	endetMitFehler := fake(t, d, "fehler", `echo "odata-recorder 1.2.3"; echo "kaputt" >&2; exit 3`)
	for name, bin := range map[string]string{"nicht ausführbar": keinExec, "kein Programm": keinFormat, "Exit ungleich 0": endetMitFehler} {
		t.Run(name, func(t *testing.T) {
			w := neueWerkzeuge(t, "", trimpathOK)
			code, _, errText := fuehre(w.u, "build", bin, "1.2.3")
			erwarte(t, code, errText, exitBefund, "Binary startet nicht:")
			if got := w.aufrufe(t); got != nil {
				t.Fatalf("go lief trotz startunfähigem Binary: %v", got)
			}
		})
	}
}

func TestBuildOhneTrimpath(t *testing.T) {
	w, bin := binaerWerkzeuge(t, `echo "odata-recorder 1.2.3"`, `echo "	build	-buildvcs=false"; echo "	build	-trimpath=false"`)
	code, _, errText := fuehre(w.u, "build", bin, "1.2.3")
	erwarte(t, code, errText, exitBefund, "das Binary ist nicht mit -trimpath gebaut")
}

func TestBuildInformationNichtLesbar(t *testing.T) {
	w, bin := binaerWerkzeuge(t, `echo "odata-recorder 1.2.3"`, `echo "not a Go binary" >&2; exit 1`)
	code, _, errText := fuehre(w.u, "build", bin, "1.2.3")
	erwarte(t, code, errText, exitBefund, "Build-Information des Binary nicht lesbar: not a Go binary")
	if strings.Contains(errText, "-trimpath") {
		t.Fatalf("Befund nennt -trimpath: %q", errText)
	}
}

// Ein Programm, das keine Go-Binary ist, besteht die Versionsprüfung und scheitert an
// `go version -m` des echten go.
func TestBuildKeinGoBinaryMitEchtemGo(t *testing.T) {
	bin := fake(t, t.TempDir(), "odata-recorder", `echo "odata-recorder 1.2.3"`)
	code, _, errText := fuehre(standard(), "build", bin, "1.2.3")
	erwarte(t, code, errText, exitBefund, "Build-Information des Binary nicht lesbar:")
}

func TestBuildRelativerPfad(t *testing.T) {
	d := t.TempDir()
	fake(t, d, "odata-recorder", `echo "odata-recorder 1.2.3"`)
	w := neueWerkzeuge(t, "", trimpathOK)
	t.Chdir(d)
	code, out, errText := fuehre(w.u, "build", "odata-recorder", "1.2.3")
	if code != exitOK {
		t.Fatalf("Exit %d, stdout %q, stderr %q", code, out, errText)
	}
}

// verzeichnisstand liefert Pfad, Größe und Änderungszeit aller Einträge unter dir.
func verzeichnisstand(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s %d %s\n", p, fi.Size(), fi.ModTime())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// pruefung hält keinen Zustand und schreibt keine Datei, auch nicht bei einem Befund.
func TestSchreibtNichts(t *testing.T) {
	wurzel := t.TempDir()
	if err := os.WriteFile(filepath.Join(wurzel, "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := fake(t, t.TempDir(), "odata-recorder", `echo "odata-recorder 1"`)
	vorher := verzeichnisstand(t, wurzel) + verzeichnisstand(t, filepath.Dir(bin))
	gut := neueWerkzeuge(t, "", trimpathOK)
	schlecht := neueWerkzeuge(t, "echo a.go", "exit 1")
	for _, u := range []Umgebung{gut.u, schlecht.u} {
		fuehre(u, "test", wurzel)
		fuehre(u, "build", bin, "1")
		fuehre(u, "build", bin, "2")
	}
	// Das Log der Fake-Programme liegt außerhalb; Wurzel und Binary-Verzeichnis bleiben unverändert.
	if nachher := verzeichnisstand(t, wurzel) + verzeichnisstand(t, filepath.Dir(bin)); nachher != vorher {
		t.Fatalf("pruefung hat Dateien verändert:\nvorher\n%s\nnachher\n%s", vorher, nachher)
	}
}
