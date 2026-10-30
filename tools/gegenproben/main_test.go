package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBewerte(t *testing.T) {
	verstoss := Fall{Name: "v", WantExit: 1, WantText: []string{"regel"}}
	erlaubt := Fall{Name: "e"}
	tests := []struct {
		name       string
		f          Fall
		code       int
		out        string
		abgelaufen bool
		rot        string // Teiltext des Grundes; leer: grün
	}{
		{"Verstoß gemeldet", verstoss, 1, "regel: x", false, ""},
		{"Verstoß nicht gemeldet", verstoss, 0, "", false, "Verstoß nicht gemeldet"},
		{"anderer Code", verstoss, 2, "regel", false, "erwartet Exit 1, war 2"},
		{"falsches Wort", verstoss, 1, "anderes", false, `nennt "regel" nicht`},
		{"zweiter Text fehlt", Fall{Name: "w", WantExit: 1, WantText: []string{"a", "b"}}, 1, "a", false, `nennt "b" nicht`},
		{"verbotener Text", Fall{Name: "v", WantExit: 1, Verboten: []string{"ok"}}, 1, "alles ok", false, `nennt "ok"`},
		{"Erlaubt grün", erlaubt, 0, "", false, ""},
		{"Erlaubt abgelehnt", erlaubt, 1, "boese", false, "erlaubter Fall abgelehnt"},
		{"Frist", verstoss, -1, "", true, "Zeit überschritten"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bewerte(tc.f, tc.code, tc.out, tc.abgelaufen)
			if tc.rot == "" && got != "" || !strings.Contains(got, tc.rot) {
				t.Fatalf("bewerte = %q, erwartet %q", got, tc.rot)
			}
		})
	}
}

func TestFaelleHabenEindeutigeNamen(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range faelle() {
		if seen[f.Name] {
			t.Fatalf("Name %s doppelt", f.Name)
		}
		seen[f.Name] = true
	}
}

// wurzelMitEingaben legt ein Verzeichnis an, das alle Eingaben als kleine Dateien trägt.
func wurzelMitEingaben(t *testing.T) string {
	t.Helper()
	w := t.TempDir()
	for _, d := range []string{"src/a.go", "go.mod", ".a-check.yml", ".golangci.yml", "tools/harness/lint.sh", "tools/harness/migration-check.sh"} {
		p := filepath.Join(w, d)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

// umgebungMit liefert eine Umgebung, deren Gate-Aufruf das Skript ist.
func umgebungMit(skript func(dir string, g Gate) string, prog ...string) Umgebung {
	return Umgebung{
		Eingaben:   standard().Eingaben,
		Programme:  prog,
		Aufraeumen: os.RemoveAll,
		Befehl: func(dir string, f Fall) (string, []string, string) {
			return "sh", []string{"-c", skript(dir, f.Gate)}, ""
		},
		Bauen: func(dir string, b Bau) (string, []string) {
			return "sh", []string{"-c", "echo '" + b.Version + "' > " + filepath.Join(dir, binaerName)}
		},
	}
}

func TestRunEingabeFehlt(t *testing.T) {
	u := umgebungMit(func(string, Gate) string { return "exit 0" }, "sh")
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, t.TempDir(), []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if out.Len() != 0 {
		t.Fatalf("kein Fall darf laufen, Ausgabe: %q", out.String())
	}
}

func TestRunProgrammFehlt(t *testing.T) {
	u := umgebungMit(func(string, Gate) string { return "exit 0" }, "gibt-es-nicht-gegenproben")
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if !strings.Contains(errb.String(), "nicht aufrufbar") {
		t.Fatalf("Meldung: %q", errb.String())
	}
}

func TestRunAlleFaelleLaufenZuEnde(t *testing.T) {
	u := umgebungMit(func(dir string, g Gate) string { return "exit 0" }, "sh")
	fs := []Fall{
		{Name: "verstoss-nicht-gemeldet", WantExit: 1},
		{Name: "erlaubt-a"},
		{Name: "erlaubt-b"},
	}
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), fs); got != exitRot {
		t.Fatalf("Exit %d, erwartet %d", got, exitRot)
	}
	for _, n := range []string{"ROT  verstoss-nicht-gemeldet", "ok   erlaubt-a", "ok   erlaubt-b", "3 Fälle, 1 rot"} {
		if !strings.Contains(out.String(), n) {
			t.Fatalf("Ausgabe nennt %q nicht:\n%s", n, out.String())
		}
	}
}

func TestRunRaeumtWegwerfverzeichnisseAuf(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	u := umgebungMit(func(string, Gate) string { return "exit 1" }, "sh")
	var out, errb bytes.Buffer
	run(u, &out, &errb, wurzelMitEingaben(t), []Fall{{Name: "rot"}})
	rest, err := os.ReadDir(tmp)
	if err != nil || len(rest) != 0 {
		t.Fatalf("Wegwerfverzeichnisse bleiben: %v %v", rest, err)
	}
}

func TestRunErkenntVeraenderteEingabe(t *testing.T) {
	w := wurzelMitEingaben(t)
	u := umgebungMit(func(string, Gate) string { return "echo x >> " + filepath.Join(w, "go.mod") }, "sh")
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, w, []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if !strings.Contains(errb.String(), "verändert") {
		t.Fatalf("Meldung: %q", errb.String())
	}
}

func TestStarteBeendetDieProzessgruppeBeiFristablauf(t *testing.T) {
	start := time.Now()
	_, _, abgelaufen := starte(200*time.Millisecond, "", nil, "sh", "-c", "sleep 30 & wait")
	if !abgelaufen {
		t.Fatal("Frist nicht erkannt")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("Rückkehr nach %v, der Kindprozess hält die Ausgabe offen", d)
	}
}

func TestJederFallHatEigenesTmpdir(t *testing.T) {
	sammel := filepath.Join(t.TempDir(), "tmpdirs")
	u := umgebungMit(func(string, Gate) string { return `echo "$TMPDIR" >> ` + sammel }, "sh")
	fs := []Fall{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), fs); got != exitOK {
		t.Fatalf("Exit %d: %s%s", got, out.String(), errb.String())
	}
	b, err := os.ReadFile(sammel)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, z := range strings.Fields(string(b)) {
		if z == os.Getenv("TMPDIR") {
			t.Fatalf("Fall teilt das TMPDIR des Prozesses: %s", z)
		}
		seen[z] = true
	}
	if len(seen) != len(fs) {
		t.Fatalf("%d verschiedene TMPDIR bei %d Fällen: %v", len(seen), len(fs), seen)
	}
}

func TestHashEingabenGoSumIstOptional(t *testing.T) {
	w := wurzelMitEingaben(t)
	ohne, err := hashEingaben(standard().Eingaben, w)
	if err != nil {
		t.Fatalf("ohne go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(w, "go.sum"), []byte("s\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mit, err := hashEingaben(standard().Eingaben, w)
	if err != nil || mit == ohne {
		t.Fatalf("go.sum geht nicht in die Prüfsumme ein: %v", err)
	}
	if err := os.Remove(filepath.Join(w, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if _, err := hashEingaben(standard().Eingaben, w); err == nil {
		t.Fatal("fehlendes go.mod bleibt unbemerkt")
	}
}

func TestRunMeldetEntfernenFehlerNachDenErgebnissen(t *testing.T) {
	u := umgebungMit(func(string, Gate) string { return "exit 0" }, "sh")
	u.Aufraeumen = func(string) error { return errors.New("nicht entfernbar") }
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{{Name: "a"}}); got != exitSetup {
		t.Fatalf("Exit %d, erwartet %d", got, exitSetup)
	}
	if !strings.Contains(out.String(), "ok   a") || !strings.Contains(errb.String(), "nicht entfernbar") {
		t.Fatalf("Ergebnisse und Meldung erwartet:\n%s%s", out.String(), errb.String())
	}
}

func TestErsetzeOrte(t *testing.T) {
	f := Fall{Argumente: []string{"build", "{binary}", "v", "{dir}/x"}}
	got := f.ersetzeOrte("/k")
	want := []string{"build", "/k/" + binaerName, "v", "/k/x"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("ersetzeOrte = %v, erwartet %v", got, want)
	}
	if len(Fall{}.ersetzeOrte("/k")) != 0 {
		t.Fatal("ohne Argumente bleibt die Liste leer")
	}
}

func TestBauenTrimpathUndVersion(t *testing.T) {
	_, mit := bauen("/k", Bau{Version: "1.2"})
	_, ohne := bauen("/k", Bau{Version: "1.2", OhneTrimpath: true})
	if !strings.Contains(strings.Join(mit, " "), "-trimpath") {
		t.Fatalf("Bau ohne -trimpath: %v", mit)
	}
	if strings.Contains(strings.Join(ohne, " "), "-trimpath") {
		t.Fatalf("Bau mit -trimpath trotz OhneTrimpath: %v", ohne)
	}
	if !strings.Contains(strings.Join(mit, " "), "main.version=1.2") {
		t.Fatalf("Version fehlt im Bau: %v", mit)
	}
}

func TestBefehlPruefungGibtArgumenteMitOrtenAus(t *testing.T) {
	f := Fall{Gate: gatePruefung, Argumente: []string{"test", "{dir}"}}
	name, args, wd := befehl("/k", f)
	if name != "pruefung" || strings.Join(args, "|") != "test|/k" || wd != "" {
		t.Fatalf("befehl = %q %v %q", name, args, wd)
	}
}

func TestFallMitBauBautVorDemGate(t *testing.T) {
	u := umgebungMit(func(dir string, g Gate) string { return "test -f " + filepath.Join(dir, binaerName) }, "sh")
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{{Name: "b", Bau: &Bau{Version: "v"}}}); got != exitOK {
		t.Fatalf("Exit %d: %s%s", got, out.String(), errb.String())
	}
}

func TestFallMitBauMeldetScheiternDesBaus(t *testing.T) {
	u := umgebungMit(func(dir string, g Gate) string { return "echo gate-lief; exit 1" }, "sh")
	u.Bauen = func(string, Bau) (string, []string) { return "sh", []string{"-c", "echo bau-kaputt; exit 3"} }
	var out, errb bytes.Buffer
	got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{{Name: "b", WantExit: 1, Bau: &Bau{Version: "v"}}})
	if got != exitRot || !strings.Contains(out.String(), "Bau des Binary scheitert") || !strings.Contains(out.String(), "bau-kaputt") {
		t.Fatalf("Exit %d, Ausgabe %q", got, out.String())
	}
	if strings.Contains(out.String(), "gate-lief") {
		t.Fatalf("das Gate lief trotz gescheitertem Bau: %q", out.String())
	}
}

func TestFaelleDerPruefungHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{
		"test/basis", "test/nicht-formatiert", "test/vet-befund", "test/roter-test",
		"build/basis", "build/ohne-trimpath", "build/falsche-version", "build/binary-fehlt", "test/aufruf",
		"race/ohne-rennen", "race/rennen-in-jedem-teilbaum", "race/rennen-und-zeitgrenze",
	}
	have := map[string]bool{}
	for _, f := range faelle() {
		have[f.Name] = true
	}
	for _, n := range want {
		if !have[n] {
			t.Errorf("Fall %s fehlt", n)
		}
	}
}

// Ein Wegwerfmodul trägt nur, was der Fall schreibt; die Kopie der Eingaben trägt src/.
func TestWegwerfmodulKopiertDieEingabenNicht(t *testing.T) {
	skript := func(dir string, _ Gate) string {
		return "test -f " + filepath.Join(dir, "go.mod") + " && test ! -e " + filepath.Join(dir, "src") +
			" && test ! -e " + filepath.Join(dir, ".golangci.yml")
	}
	u := umgebungMit(skript, "sh")
	wegwerf := Fall{Name: "w", Wegwerfmodul: true, Neu: map[string]string{"go.mod": "module w\n"}}
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{wegwerf}); got != exitOK {
		t.Fatalf("Wegwerfmodul: Exit %d: %s%s", got, out.String(), errb.String())
	}
	kopie := Fall{Name: "k", Neu: map[string]string{"go.mod": "module w\n"}}
	out.Reset()
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{kopie}); got != exitRot {
		t.Fatalf("Kopie der Eingaben: Exit %d (erwartet %d, src/ gehört dazu): %s", got, exitRot, out.String())
	}
}

// Jeder Fall von pruefung race legt ein Wegwerfmodul an (SPEC-102, Aufbau): das Feld steht in den
// Fällen 34 bis 36, nicht nur in einem Fall, den ein Test eigens baut.
func TestRaceFaelleLegenEinWegwerfmodulAn(t *testing.T) {
	for _, f := range raceFaelle() {
		if !f.Wegwerfmodul {
			t.Errorf("%s: Wegwerfmodul nicht gesetzt, der Fall kopiert die Eingaben", f.Name)
		}
	}
}

// Jeder Fall 25 bis 28 von pruefung test legt ein Wegwerfmodul an (SPEC-102, Aufbau) und fährt damit kein Paket von
// src/; der Fall test/aufruf ruft ohne Wurzel und gehört nicht dazu.
func TestTestFaelleLegenEinWegwerfmodulAn(t *testing.T) {
	var geprueft int
	for _, f := range testFaelle() {
		if f.Name == "test/aufruf" {
			continue
		}
		geprueft++
		if !f.Wegwerfmodul {
			t.Errorf("%s: Wegwerfmodul nicht gesetzt, der Fall kopiert die Eingaben", f.Name)
		}
	}
	if geprueft != 4 {
		t.Errorf("%d Fälle geprüft, erwartet 4 (Fälle 25 bis 28)", geprueft)
	}
}

// Die Quellen der Fälle von pruefung race führen kein Zeichen über ASCII.
func TestRaceFaelleFuehrenNurASCII(t *testing.T) {
	for _, f := range raceFaelle() {
		for datei, inhalt := range f.Neu {
			for i := range len(inhalt) {
				if inhalt[i] > 0x7F {
					t.Errorf("%s: %s trägt ein Zeichen über ASCII bei Byte %d", f.Name, datei, i)
					break
				}
			}
		}
	}
}

// Die Programmliste von standard() nennt die acht Programme aus SPEC-102 Randform 2.
func TestStandardNenntDieAchtProgramme(t *testing.T) {
	have := map[string]bool{}
	for _, p := range standard().Programme {
		have[p] = true
	}
	for _, n := range []string{"a-check", "golangci-lint", "sh", "pruefung", "go", "gofmt", "groessenmarke", "git"} {
		if !have[n] {
			t.Errorf("Programm %s fehlt in standard().Programme", n)
		}
	}
	if !filepath.IsAbs(standard().Treiber) || !strings.HasSuffix(standard().Treiber, "tools/harness/groessenmarke.sh") {
		t.Errorf("Treiber: %q", standard().Treiber)
	}
}

// Die 83 Fälle 38 bis 120 tragen die Namen aus SPEC-102, jeder genau einmal.
func TestFaelleDerGroessenmarkeHabenDieNamenAusDemVertrag(t *testing.T) {
	var want []string
	for _, p := range []string{"csdl", "json", "import"} {
		want = append(want, "messpunkt-"+p)
	}
	for _, p := range []string{"csdl", "json", "import"} {
		want = append(want, "marke-der-platte-"+p)
	}
	want = append(want, "marke-eine-stufe-zu-niedrig", "marke-eine-stufe-zu-hoch", "schaetzung-um-eins-zu-niedrig",
		"schaetzung-um-eins-zu-hoch", "tabelle-gewachsen-im-code", "messung-gleich-marke", "messung-ueber-marke-ohne-eintrag",
		"eintrag-deckt", "eintrag-reicht-nicht", "eintrag-ohne-grund", "eintrag-ohne-ueberschreitung",
		"paar-marke-unter-formel-und-messung-ueber-marke", "paar-eintrag-und-marke-unter-formel", "schnitt-ohne-grund",
		"schnitt-mit-grund", "schnitt-genau-darunter", "kein-block", "prosa-form", "marke-steht-aus",
		"paar-marke-und-basis-fehlen", "block-unbekannter-schluessel")
	for _, z := range []string{"tausender-leerzeichen", "tausender-punkt", "minus", "fuehrende-null", "leer", "zehn-stellen"} {
		want = append(want, "zahl-"+z)
	}
	for _, z := range []string{"leer", "absolut", "punktpunkt", "magie", "doppelt", "in-spec", "in-docs", "in-vertraegen"} {
		want = append(want, "pfad-"+z)
	}
	want = append(want, "tabelle-fehlt", "tabelle-ohne-zeilen", "tabelle-mehrdeutig", "tabelle-zellenzahl", "tabelle-nr-doppelt")
	for _, z := range []string{"leer", "ascii-ae", "grossbuchstabe", "zaehlt-nicht", "fett"} {
		want = append(want, "klasse-"+z)
	}
	want = append(want, "messung-binaer", "messung-umbenennung", "messung-leer", "messung-fremde-basis", "messung-pfad-fehlt",
		"messung-pfad-zusaetzlich", "messung-pfade-vertauscht", "messung-kopf-fehlt", "messung-ohne-numstat-zeile",
		"messung-datensatz-ohne-nul", "messung-datensatz-ohne-tab", "aufruf", "plan-fehlt", "messung-fehlt",
		"treiber-aufrufe-von-git", "treiber-pfad-argumente", "treiber-basis-kein-vorfahr", "treiber-basis-unbekannt",
		"treiber-head-nicht-lesbar", "treiber-git-diff-scheitert", "treiber-git-nicht-aufrufbar", "treiber-keine-plaene",
		"treiber-plan-suche", "treiber-exit-aggregation", "treiber-ohne-block-kein-git", "treiber-aufraeumen",
		"treiber-git-ausschluesse", "treiber-git-sonderzeichen", "treiber-git-umbenennung", "treiber-git-binaer",
		"treiber-git-basis-kein-vorfahr", "treiber-git-basis-gleich-spitze")
	have := map[string]int{}
	fs := groessenmarkeFaelle()
	for _, f := range fs {
		have[f.Name]++
	}
	if len(want) != 83 || len(fs) != 83 {
		t.Errorf("%d erwartete Namen und %d Fälle, erwartet 83", len(want), len(fs))
	}
	for _, n := range want {
		if have["groessenmarke/"+n] != 1 {
			t.Errorf("Fall groessenmarke/%s kommt %d-mal vor", n, have["groessenmarke/"+n])
		}
	}
}

// ersatzLauf legt die Ersatz-Skripte eines Treiber-Falls an und ruft eines mit den Argumenten.
func ersatzLauf(t *testing.T, dir, skript string, args ...string) (int, string) {
	t.Helper()
	code, out, _ := starte(10*time.Second, dir, nil, "sh", append([]string{filepath.Join(dir, "bin", skript)}, args...)...)
	return code, out
}

// Die Ersatz-Skripte halten jedes Argument fest und zählen jeden Aufruf, auch die Operationen, die der
// Treiber nie rufen darf (SPEC-102 Randform 12).
func TestErsatzSkripteProtokollierenJedenAufruf(t *testing.T) {
	dir := t.TempDir()
	neu := gmErsatz(map[string]string{
		"ersatz/git.merge-base.rc": "1\n", "ersatz/git.diff.out": "3\t4\tx\x00", "ersatz/git.diff.err": "fatal: boese\n",
		"ersatz/git.diff.rc": "128\n",
	}, gmEintrag{pfad: "plaene/slice-a.md", abfrage: "messen\n", hinweis: "Hinweis\n", pruefenRc: "2"})
	if err := gmSchreibe(dir, neu); err != nil {
		t.Fatal(err)
	}
	messung := filepath.Join(dir, "m")
	if err := os.WriteFile(messung, []byte("kopf\x00rest"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAufrufe := [][]string{
		{"rev-parse", "--verify", "HEAD"}, {"merge-base", "--is-ancestor", "a b", "HEAD"}, {"diff", "-z", "x*", "q\"r", "\xc3\xa4", ""},
		{"status"}, {"log", "--all"}, {},
	}
	wantRc := []int{0, 1, 128, 0, 0, 0}
	for i, a := range gitAufrufe {
		if code, _ := ersatzLauf(t, dir, "git", a...); code != wantRc[i] {
			t.Errorf("git %q: Exit %d statt %d", a, code, wantRc[i])
		}
	}
	if code, out := ersatzLauf(t, dir, "git", "diff"); code != 128 || out != "3\t4\tx\x00fatal: boese\n" && out != "fatal: boese\n3\t4\tx\x00" {
		t.Errorf("git diff: Exit %d, Ausgabe %q", code, out)
	}
	l := Lauf{Dir: dir}
	wantGit := append(append([][]string(nil), gitAufrufe...), []string{"diff"})
	if r := gmErwartet(l, "git.log", wantGit); r != "" {
		t.Errorf("Protokoll von git: %s", r)
	}
	pa := "plaene/slice-a.md"
	if code, out := ersatzLauf(t, dir, "programm", "abfrage", pa); code != 0 || out != "Hinweis\nmessen\n" {
		t.Errorf("abfrage: Exit %d, Ausgabe %q", code, out)
	}
	if code, _ := ersatzLauf(t, dir, "programm", "pruefen", pa, messung); code != 2 {
		t.Errorf("pruefen: Exit %d statt 2", code)
	}
	if code, _ := ersatzLauf(t, dir, "programm", "unbekannt", "x"); code != 99 {
		t.Errorf("unbekannte Operation: Exit %d statt 99", code)
	}
	wantProgramm := [][]string{{"abfrage", pa}, {"pruefen", pa, messung}, {"unbekannt", "x"}}
	if r := gmErwartet(l, "programm.log", wantProgramm); r != "" {
		t.Errorf("Protokoll des Programms: %s", r)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "log", "messung-slice-a.md")); err != nil || string(b) != "kopf\x00rest" {
		t.Errorf("Kopie der Messung: %q %v", b, err)
	}
}

// Ein Fall, dessen Protokoll fehlt, ist rot, auch wenn er keine Aufrufe erwartet (SPEC-102
// Randform 12); ein Protokoll mit einem Aufruf mehr oder weniger ist es ebenso.
func TestProtokollFehltOderWeichtAbIstRot(t *testing.T) {
	for _, f := range groessenmarkeFaelle() {
		if f.Nachher == nil {
			continue
		}
		if r := f.Nachher(Lauf{Dir: t.TempDir(), Tmp: t.TempDir()}); r == "" {
			t.Errorf("%s: Nachher ohne Protokoll ist grün", f.Name)
		}
	}
	dir := t.TempDir()
	if err := gmSchreibe(dir, map[string]string{"log/git.log": "AUFRUF\nARG a\n"}); err != nil {
		t.Fatal(err)
	}
	l := Lauf{Dir: dir}
	for _, want := range [][][]string{nil, {{"a"}, {"a"}}, {{"b"}}, {{"a", "b"}}} {
		if r := gmErwartet(l, "git.log", want); r == "" {
			t.Errorf("Erwartung %q gilt für das Protokoll mit einem Aufruf a", want)
		}
	}
	if r := gmErwartet(l, "git.log", [][]string{{"a"}}); r != "" {
		t.Errorf("Erwartung a: %s", r)
	}
	if err := gmSchreibe(dir, map[string]string{"log/git.log": "AUFRUF\nARG a\nARG b\n"}); err != nil {
		t.Fatal(err)
	}
	if r := gmErwartet(l, "git.log", [][]string{{"a"}}); r == "" {
		t.Error("ein Aufruf mit mehr Argumenten als erwartet gilt")
	}
	if err := gmSchreibe(dir, map[string]string{"log/git.log": "murks\n"}); err != nil {
		t.Fatal(err)
	}
	if r := gmErwartet(l, "git.log", nil); r == "" {
		t.Error("unlesbares Protokoll gilt")
	}
}

// Env, Ausfuehrbar, Vorbereitung und Nachher eines Falls wirken vor und nach dem Aufruf.
func TestFallMitEnvVorbereitungUndNachher(t *testing.T) {
	skript := func(dir string, _ Gate) string {
		return `test "$EIGENES" = "` + dir + `/x" && test -x ` + filepath.Join(dir, "lauf.sh") + ` && test -f ` + filepath.Join(dir, "von-vorbereitung") + ` && echo "$PATH" | grep -q "^` + dir + `:"`
	}
	u := umgebungMit(skript, "sh")
	f := Fall{
		Name: "w", Wegwerfmodul: true, Neu: map[string]string{"lauf.sh": "#!/bin/sh\n"}, Ausfuehrbar: []string{"lauf.sh"},
		Env: []string{"EIGENES={dir}/x", "PATH={dir}:{PATH}"},
		Vorbereitung: func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "von-vorbereitung"), nil, 0o644)
		},
	}
	var gesehen Lauf
	f.Nachher = func(l Lauf) string {
		gesehen = l
		return ""
	}
	var out, errb bytes.Buffer
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{f}); got != exitOK {
		t.Fatalf("Exit %d: %s%s", got, out.String(), errb.String())
	}
	if gesehen.Dir == "" || gesehen.Tmp == "" || len(gesehen.Env) != 3 || !strings.HasPrefix(gesehen.Env[0], "TMPDIR=") {
		t.Errorf("Nachher sah %+v", gesehen)
	}
	f.Nachher = func(Lauf) string { return "Protokoll zeigt es" }
	out.Reset()
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{f}); got != exitRot || !strings.Contains(out.String(), "ROT  w: Protokoll zeigt es") {
		t.Errorf("rotes Nachher: Exit %d, Ausgabe %q", got, out.String())
	}
	// Nachher wird nur gefragt, wenn der Aufruf die Erwartung erfüllt.
	f.WantExit = 1
	called := false
	f.Nachher = func(Lauf) string {
		called = true
		return ""
	}
	out.Reset()
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{f}); got != exitRot || called {
		t.Errorf("Verstoß nicht gemeldet: Exit %d, Nachher gerufen: %v", got, called)
	}
	f.WantExit = 0
	f.Vorbereitung = func(string) error { return errors.New("kaputt") }
	out.Reset()
	if got := run(u, &out, &errb, wurzelMitEingaben(t), []Fall{f}); got != exitSetup || !strings.Contains(errb.String(), "kaputt") {
		t.Errorf("Vorbereitung scheitert: Exit %d, %q", got, errb.String())
	}
}

// Die Fälle zu groessenmarke rufen das Programm bzw. den Treiber im Verzeichnis des Falls.
func TestBefehlGroessenmarke(t *testing.T) {
	u := standard()
	name, args, wd := u.Befehl("/k", Fall{Gate: gateGroessenmarke, Argumente: []string{"pruefen", "{dir}/plan.md"}})
	if name != "groessenmarke" || strings.Join(args, "|") != "pruefen|/k/plan.md" || wd != "/k" {
		t.Errorf("Programm: %q %v %q", name, args, wd)
	}
	name, args, wd = u.Befehl("/k", Fall{Gate: gateTreiber, Argumente: []string{"{dir}/bin/programm", "plaene"}})
	if name != "sh" || args[0] != u.Treiber || strings.Join(args[1:], "|") != "/k/bin/programm|plaene" || wd != "/k" {
		t.Errorf("Treiber: %q %v %q", name, args, wd)
	}
	name, _, _ = u.Befehl("/k", Fall{Gate: gateLint})
	if name != "sh" {
		t.Errorf("andere Gates bleiben: %q", name)
	}
}

// Die Fälle zu groessenmarke führen Zeichen jenseits von ASCII nur als Bytes, die der Vertrag verlangt
// (SPEC-102): Plan, Messung und Ersatz-Skripte tragen keine Backslash-u-Folge als Text.
func TestGroessenmarkeFaelleTragenKeineEscapeFolgen(t *testing.T) {
	for _, f := range groessenmarkeFaelle() {
		for datei, inhalt := range f.Neu {
			if strings.Contains(inhalt, "\\u") {
				t.Errorf("%s: %s trägt eine Escape-Folge als Text", f.Name, datei)
			}
		}
	}
}
