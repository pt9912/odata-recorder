package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// bspMit legt ein Beispiel mit args an; das Ersatz-Binary antwortet auf args mit a.
func bspMit(t *testing.T, args []string, out, errA []string, code string, a antwort) umfeld {
	t.Helper()
	kopf := praef
	if len(args) > 0 {
		kopf += " " + strings.Join(args, " ")
	}
	ant := okAntworten()
	ant[strings.Join(args, " ")] = a
	return bspUmfeld(t, "k.txt", bspDatei(kopf, out, errA, code), ant)
}

func gleich(a, b []string) bool {
	return strings.Join(a, "\x00") == strings.Join(b, "\x00") && len(a) == len(b)
}

// Z22: Argumente unverändert und in der Reihenfolge, ohne Shell, an beiden Aufrufstellen
// (version, Beispiel).
func TestHandbuchArgumente(t *testing.T) {
	listen := [][]string{{"a=b"}, {"-x"}, {"--"}, {"A.b_c"}, {"zweites", "erstes"}}
	for _, args := range listen {
		m := bspMit(t, args, nil, nil, "0\n", antwort{})
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("%v: Exit %d %q", args, code, errText)
		}
		s := m.e.starts(t)
		if len(s) != 2 || !gleich(s[0].args, []string{"version"}) || !gleich(s[1].args, args) {
			t.Fatalf("%v: Starts %+v", args, s)
		}
	}
}

// Z23: leere Umgebung, leeres eigenes Arbeitsverzeichnis außerhalb der Wurzel, leere Standardeingabe,
// an beiden Aufrufstellen.
func TestHandbuchUmgebung(t *testing.T) {
	t.Setenv("PRUEFUNG_PROBE", "durchgereicht")
	stdin, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stdin.WriteString("xxxxx"); err != nil {
		t.Fatal(err)
	}
	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	alt := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() { os.Stdin = alt })
	m := neuesUmfeld(t, standardAntworten())
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("Exit %d %q", code, errText)
	}
	hier, _ := os.Getwd()
	s := m.e.starts(t)
	if len(s) != 2 {
		t.Fatalf("%d Starts", len(s))
	}
	for i, st := range s {
		cwd, _ := filepath.EvalSymlinks(st.cwd)
		wurzel, _ := filepath.EvalSymlinks(m.wurzel)
		if st.env != 0 || st.in != 0 || st.ent != 0 {
			t.Fatalf("Start %d: Umgebung %d, Eingabe %d, Einträge %d", i, st.env, st.in, st.ent)
		}
		if cwd == wurzel || strings.HasPrefix(cwd, wurzel+"/") || cwd == hier || !strings.HasPrefix(st.cwd, os.TempDir()) {
			t.Fatalf("Start %d: Arbeitsverzeichnis %q (Wurzel %q, hier %q)", i, st.cwd, wurzel, hier)
		}
	}
	if s[0].cwd == s[1].cwd {
		t.Fatalf("beide Läufe teilen ein Arbeitsverzeichnis: %q", s[0].cwd)
	}
}

// Z24: Standardausgabe und Fehlerausgabe werden getrennt verglichen, an beiden Aufrufstellen.
func TestHandbuchStroeme(t *testing.T) {
	faelle := []struct {
		name         string
		out, err     string
		abwStd, abwF bool
	}{
		{"vertauscht", "b\n", "a\n", true, true},
		{"beide", "c\n", "d\n", true, true},
		{"nur Standardausgabe", "c\n", "b\n", true, false},
		{"nur Fehlerausgabe", "a\n", "d\n", false, true},
		{"gleich", "a\n", "b\n", false, false},
	}
	for _, f := range faelle {
		m := bspMit(t, nil, []string{"a"}, []string{"b"}, "0\n", antwort{out: f.out, err: f.err})
		_, _, errText := m.lauf()
		if got := strings.Contains(errText, "k.txt: Standardausgabe weicht ab (erwartet \"a\\n\", erhalten "); got != f.abwStd {
			t.Fatalf("%s: Standardausgabe weicht ab=%v: %q", f.name, got, errText)
		}
		if got := strings.Contains(errText, "k.txt: Fehlerausgabe weicht ab (erwartet \"b\\n\", erhalten "); got != f.abwF {
			t.Fatalf("%s: Fehlerausgabe weicht ab=%v: %q", f.name, got, errText)
		}
	}
	version := []struct {
		name string
		a    antwort
		text string
	}{
		{"gleich", antwort{out: "odata-recorder dev\n"}, ""},
		{"Fehlerausgabe", antwort{out: "odata-recorder dev\n", err: "x\n"}, "Ausgabe nicht lesbar: Fehlerausgabe nicht leer"},
		{"vertauscht", antwort{out: "", err: "odata-recorder dev\n"}, "Ausgabe nicht lesbar: Fehlerausgabe nicht leer"},
		{"nur Standardausgabe falsch", antwort{out: "x\n"}, "Ausgabe nicht lesbar: Standardausgabe \"x\\n\""},
	}
	for _, f := range version {
		m := neuesUmfeld(t, map[string]antwort{"version": f.a, "": {out: "odata-recorder dev\n"}})
		_, _, errText := m.lauf()
		if f.text == "" {
			nichtEnthalten(t, errText, "version:")
		} else {
			enthaelt2(t, errText, "version: "+f.text)
		}
	}
}

// Z25: der Ausgangswert wird verglichen, ein Signal ist Signal, an beiden Aufrufstellen.
func TestHandbuchAusgangswert(t *testing.T) {
	faelle := []struct {
		datei, ist int
		signal     bool
		text       string
	}{
		{0, 0, false, ""}, {2, 2, false, ""}, {255, 255, false, ""},
		{0, 1, false, "Ausgangswert weicht ab (erwartet 0, erhalten 1)"},
		{2, 0, false, "Ausgangswert weicht ab (erwartet 2, erhalten 0)"},
		{255, 254, false, "Ausgangswert weicht ab (erwartet 255, erhalten 254)"},
		{0, 0, true, "Ausgangswert weicht ab (erwartet 0, erhalten Signal)"},
	}
	for _, f := range faelle {
		m := bspMit(t, nil, nil, nil, fmt.Sprint(f.datei, "\n"), antwort{code: f.ist, signal: f.signal})
		code, _, errText := m.lauf()
		if f.text == "" {
			if code != exitOK {
				t.Fatalf("%+v: Exit %d %q", f, code, errText)
			}
			continue
		}
		erwarte(t, code, errText, exitBefund, "k.txt: "+f.text)
	}
	for _, f := range []struct {
		a    antwort
		text string
	}{{antwort{out: "odata-recorder dev\n", code: 1}, "Exit 1"}, {antwort{out: "odata-recorder dev\n", code: 255}, "Exit 255"},
		{antwort{out: "odata-recorder dev\n", signal: true}, "Signal"}} {
		m := neuesUmfeld(t, map[string]antwort{"version": f.a, "": {out: "odata-recorder dev\n"}})
		_, _, errText := m.lauf()
		enthaelt2(t, errText, "version: Ausgabe nicht lesbar: "+f.text)
	}
}

// Z26: Zeitgrenze, Wirkungs-Fänger: der Prozess wird beendet und der Befund kommt, an beiden
// Aufrufstellen.
func TestHandbuchZeitgrenze(t *testing.T) {
	for _, stelle := range []string{"version", "Beispiel"} {
		a := okAntworten()
		text := "k.txt: Zeitgrenze überschritten"
		if stelle == "version" {
			a["version"] = antwort{schlaf: "5"}
			text = "version: Zeitgrenze überschritten"
		} else {
			a[""] = antwort{schlaf: "5"}
		}
		m := bspUmfeld(t, "k.txt", bspDatei(praef, nil, nil, "0\n"), a)
		m.u.Zeitgrenze = 200 * time.Millisecond
		t0 := time.Now()
		code, _, errText := m.lauf()
		dauer := time.Since(t0)
		erwarte(t, code, errText, exitBefund, text)
		if dauer > 2*time.Second {
			t.Fatalf("%s: Dauer %v", stelle, dauer)
		}
		b, err := os.ReadFile(filepath.Join(m.e.dir, "pid"))
		if err != nil {
			t.Fatal(err)
		}
		if err := syscall.Kill(zahl(string(b)), 0); err != syscall.ESRCH {
			t.Fatalf("%s: Prozess %s läuft noch (%v)", stelle, strings.TrimSpace(string(b)), err)
		}
	}
	m := neuesUmfeld(t, standardAntworten())
	m.u.Zeitgrenze = 200 * time.Millisecond
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("Gutfall mit gekürzter Zeitgrenze: %d %q", code, errText)
	}
}

// Z26 (Wert): die Zeitgrenze des Programms ist 10 s.
func TestHandbuchZeitgrenzeWert(t *testing.T) {
	if g := standard().Zeitgrenze; g != 10*time.Second {
		t.Fatalf("Zeitgrenze %v", g)
	}
}

// Z27: jede Datei einmal gefahren, version einmal, ungültige und ungebundene Dateien nie.
func TestHandbuchStartsJeDatei(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	b := binde("beispiele/aufruf.txt")
	m.handbuch(rumpf(b + zaunVon(aufrufDatei) + "\n" + b + zaunVon(aufrufDatei)))
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("Exit %d %q", code, errText)
	}
	if s := m.e.starts(t); len(s) != 2 || !gleich(s[0].args, []string{"version"}) || len(s[1].args) != 0 {
		t.Fatalf("zwei Blöcke: Starts %+v", s)
	}
	m = neuesUmfeld(t, standardAntworten())
	m.schreibe("docs/user/beispiele/verwaist.txt", praef+" foo\n--- Standardausgabe\n--- Fehlerausgabe\n--- Ausgangswert\n0\n")
	m.lauf()
	if s := m.e.starts(t); len(s) != 2 {
		t.Fatalf("verwaiste Datei: Starts %+v", s)
	}
}

// Z28: das Aufrufpräfix ist durch Image und Pfad im Bau gedeckt.
func TestHandbuchAufrufpraefix(t *testing.T) {
	faelle := []struct {
		name  string
		setup func(m umfeld)
		text  string
	}{
		{"Image fehlt", func(m umfeld) { m.schreibe("Makefile", "build: ## x\n") }, "Aufrufpräfix nicht gedeckt: Image"},
		{"Makefile fehlt", func(m umfeld) { m.entferne("Makefile") }, "Aufrufpräfix nicht gedeckt: Image"},
		{"Image in mk-Datei", func(m umfeld) {
			m.schreibe("Makefile", "build: ## x\n")
			m.schreibe("harness/mk/x.mk", "\tdocker build -t odata-recorder:build .\n")
		}, ""},
		{"Pfad fehlt", func(m umfeld) { m.schreibe("Dockerfile", "RUN go build -o /out/anders\n") }, "Aufrufpräfix nicht gedeckt: Pfad"},
		{"Dockerfile fehlt", func(m umfeld) { m.entferne("Dockerfile") }, "Aufrufpräfix nicht gedeckt: Pfad"},
		{"gedeckt", func(umfeld) {}, ""},
		{"ohne Beispiele", func(m umfeld) {
			m.entferne("docs/user/beispiele")
			m.handbuch(rumpf("Text.\n"))
			m.entferne("Makefile")
			m.entferne("Dockerfile")
		}, ""},
		{"leeres Verzeichnis", func(m umfeld) {
			m.entferne("docs/user/beispiele/aufruf.txt")
			m.handbuch(rumpf("Text.\n"))
			m.entferne("Makefile")
		}, ""},
	}
	for _, f := range faelle {
		m := neuesUmfeld(t, standardAntworten())
		f.setup(m)
		code, _, errText := m.lauf()
		if f.text == "" {
			if code != exitOK {
				t.Fatalf("%s: Exit %d %q", f.name, code, errText)
			}
			continue
		}
		erwarte(t, code, errText, exitBefund, f.text)
		other := "Image"
		if strings.HasSuffix(f.text, "Image") {
			other = "Pfad"
		}
		nichtEnthalten(t, errText, "gedeckt: "+other)
	}
}

func zielDateien(m umfeld) {
	m.schreibe("Makefile", "gates: ## a\nhelp: ## b\n")
	m.schreibe("harness/mk/a.mk", "build: ## x\ntest: ## y\n")
	m.schreibe("harness/mk/b.mk", "lint: ## x\ndocs-check: ## y\n")
	m.schreibe("d-check.mk", "dcheck-x: ## z\n")
	m.schreibe("repo.mk", "repo-y: ## z\n")
}

// Z29: Make-Ziele; Generator über jedes Ziel der Zielliste und sechs Formen, Orakel: die Liste der Fixture.
func TestHandbuchMakeZiele(t *testing.T) {
	m := ohneBeispiele(t)
	zielDateien(m)
	n := 0
	for _, z := range []string{"gates", "help", "build", "test", "lint", "docs-check", "dcheck-x", "repo-y"} {
		n++
		m.handbuch(rumpf("Siehe `make " + z + "`.\n"))
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("Ziel %s: Exit %d %q", z, code, errText)
		}
	}
	faelle := []struct{ prosa, text string }{
		{"Siehe `make unbekannt`.\n", "Zeile 9: make-Ziel unbekannt unbekannt"},
		{"Siehe `make -j build`.\n", "Zeile 9: make-Aufruf ohne Ziel an erster Stelle: make -j build"},
		{"Siehe `make X=1 build`.\n", "Zeile 9: make-Aufruf ohne Ziel an erster Stelle: make X=1 build"},
		{"Siehe `make` allein.\n", ""},
		{"Siehe `make build` und `make nix`.\n", "Zeile 9: make-Ziel nix unbekannt"},
		{"Siehe `make\nunbekannt` hier.\n", "Zeile 9: make-Ziel unbekannt unbekannt"},
	}
	for _, f := range faelle {
		n++
		m.handbuch(rumpf(f.prosa))
		code, _, errText := m.lauf()
		if f.text == "" {
			if code != exitOK {
				t.Fatalf("%q: Exit %d %q", f.prosa, code, errText)
			}
			continue
		}
		erwarte(t, code, errText, exitBefund, f.text)
		if strings.Count(errText, "\n") != 1 {
			t.Fatalf("%q: mehr als ein Befund: %q", f.prosa, errText)
		}
	}
	if n != 14 {
		t.Fatalf("%d Elemente, erwartet 14", n)
	}
	m.handbuch(rumpf("Erst Text.\nDann `make\nbuild` und `make nix` hier.\n"))
	_, _, errText := m.lauf()
	enthaelt2(t, errText, "Zeile 11: make-Ziel nix unbekannt")
}

// Z29 (Rand): ein Span, der nicht mit `make ` beginnt, eine Kette und ein Span ohne zweites Wort
// sind kein Aufruf; ein Span schließt nur an einer Folge gleicher Länge.
func TestHandbuchMakeSpanRand(t *testing.T) {
	m := ohneBeispiele(t)
	zielDateien(m)
	for _, prosa := range []string{"Siehe `cd x && make nix` hier.\n", "Siehe `make  ` hier.\n", "Siehe `x make nix` hier.\n"} {
		m.handbuch(rumpf(prosa))
		if code, _, errText := m.lauf(); code != exitOK {
			t.Fatalf("%q: Exit %d %q", prosa, code, errText)
		}
	}
	m.handbuch(rumpf("Siehe `make ``build`` nix` hier.\n"))
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 9: make-Ziel ``build`` unbekannt")
	m.handbuch(rumpf("Siehe ``make nix`` hier.\n"))
	code, _, errText = m.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 9: make-Ziel nix unbekannt")
}

// Z30: die Zielliste kennt nur Zeilen der Form ^[a-z0-9-]+:.*## in Makefile, harness/mk/*.mk, d-check.mk, repo.mk.
func TestHandbuchZielliste(t *testing.T) {
	faelle := []struct {
		name, ziel string
		setup      func(m umfeld)
	}{
		{"ohne ##", "x", func(m umfeld) { m.schreibe("Makefile", "x: weil kein Kommentar\n") }},
		{"ohne Endung mk", "x", func(m umfeld) { m.schreibe("harness/mk/foo.txt", "x: ## a\n") }},
		{"Unterverzeichnis", "x", func(m umfeld) { m.schreibe("harness/mk/sub/x.mk", "x: ## a\n") }},
		{"Großbuchstabe", "X", func(m umfeld) { m.schreibe("Makefile", "X: ## a\n") }},
		{"Dateien fehlen", "x", func(m umfeld) { m.entferne("Makefile") }},
	}
	for _, f := range faelle {
		m := ohneBeispiele(t)
		m.schreibe("Makefile", "y: ## a\n")
		f.setup(m)
		m.handbuch(rumpf("Siehe `make " + f.ziel + "`.\n"))
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "make-Ziel "+f.ziel+" unbekannt")
	}
	m := ohneBeispiele(t)
	m.schreibe("Makefile", "x: ## a\n")
	m.handbuch(rumpf("Siehe `make x`.\n"))
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("Gutfall: %d %q", code, errText)
	}
}

// tropfen ist eine Prüfsumme über Pfade und Inhalte unter wurzel.
func pruefsumme(t *testing.T, wurzel string) [32]byte {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(wurzel, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Fprintln(h, p, d.IsDir())
		if !d.IsDir() {
			b, err := os.ReadFile(p)
			h.Write(b)
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var s [32]byte
	copy(s[:], h.Sum(nil))
	return s
}

// Z39: die Wurzel bleibt unverändert, das Arbeitsverzeichnis wird entfernt, am normalen Ende
// (ohne Befund, mit Befund, bei nicht startbarem Binary).
func TestHandbuchWirkung(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	faelle := []struct {
		name string
		h    string
		bin  string
	}{
		{"ohne Befund", sauber(), m.e.bin},
		{"mit Befund", sauber() + "\nDas ist geplant.\n", m.e.bin},
		{"Binary nicht startbar", sauber(), filepath.Join(m.wurzel, "Makefile")},
	}
	for _, f := range faelle {
		m.handbuch(f.h)
		if err := os.Chmod(filepath.Join(m.wurzel, "Makefile"), 0o755); err != nil {
			t.Fatal(err)
		}
		vorher := pruefsumme(t, m.wurzel)
		fuehre(m.u, "handbuch", f.bin, m.wurzel)
		if nachher := pruefsumme(t, m.wurzel); nachher != vorher {
			t.Fatalf("%s: die Wurzel hat sich geändert", f.name)
		}
		if e, _ := os.ReadDir(tmp); len(e) != 0 {
			t.Fatalf("%s: %d Einträge in TMPDIR", f.name, len(e))
		}
	}
}
