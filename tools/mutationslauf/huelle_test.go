package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// stubGo legt ein Ersatz-Skript für `go` an, das Argumente und CGO_ENABLED festhält und die Ausgabedatei von -o anlegt.
func stubGo(t *testing.T) (log string, env []string) {
	t.Helper()
	d := t.TempDir()
	log = filepath.Join(d, "log")
	skript := "#!/bin/sh\necho \"ARGS $*\" >> \"$STUB_LOG\"\necho \"CGO $CGO_ENABLED\" >> \"$STUB_LOG\"\np=\nfor a; do [ \"$p\" = -o ] && : > \"$a\"; p=$a; done\n"
	if err := os.WriteFile(filepath.Join(d, "go"), []byte(skript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", d+":"+os.Getenv("PATH"))
	return log, append(os.Environ(), "STUB_LOG="+log)
}

func zeilen(t *testing.T, p string) []string {
	t.Helper()
	roh, _ := os.ReadFile(p)
	return strings.Split(strings.TrimSpace(string(roh)), "\n")
}

// TestBauOptionen (SPEC-210 Nr. 5 b, 9 a): der Bau trägt -trimpath, -buildvcs=false, -p 2 und -o; mit Detektor -race und
// cgo, ohne CGO_ENABLED=0; das Paket steht am Ende.
func TestBauOptionen(t *testing.T) {
	log, env := stubGo(t)
	e := &echt{env: env, bauZeit: time.Minute}
	for _, detektor := range []bool{false, true} {
		_ = os.Remove(log)
		ziel := filepath.Join(t.TempDir(), "p.test")
		a := e.Bauen(context.Background(), Bau{Dir: t.TempDir(), Paket: "./src/p", Ziel: ziel, Detektor: detektor})
		want := "ARGS test -c -trimpath -buildvcs=false -p 2 -o " + ziel + " ./src/p"
		cgo := "CGO 0"
		if detektor {
			want = "ARGS test -c -trimpath -buildvcs=false -p 2 -o " + ziel + " -race ./src/p"
			cgo = "CGO 1"
		}
		if got := zeilen(t, log); a.Code != 0 || a.KeinBinary || !reflect.DeepEqual(got, []string{want, cgo}) {
			t.Errorf("Detektor %v: %q, %+v", detektor, got, a)
		}
	}
}

// TestKindUmgebung (SPEC-210 Nr. 3 c, 8 a): netzlos und ohne Toolchain-Wechsel; der Cache steht zuletzt.
func TestKindUmgebung(t *testing.T) {
	env := kindUmgebung([]string{"A=1", "GOPROXY=https://x"}, "/c")
	for k, want := range map[string]string{"GOCACHE": "/c", "GOFLAGS": "-mod=readonly", "GOPROXY": "off", "GOTOOLCHAIN": "local", "A": "1"} {
		if got, _ := lookup(env, k); got != want {
			t.Errorf("%s=%q statt %q", k, got, want)
		}
	}
}

// TestCacheAdapter (SPEC-210 Nr. 8 d): die Größe zählt die Bytes regulärer Dateien, keine Verzeichnisse und keine
// Verweise; geleert wird mit `go clean -cache`.
func TestCacheAdapter(t *testing.T) {
	log, env := stubGo(t)
	cache := t.TempDir()
	schreibe(t, cache, "a", strings.Repeat("a", 10))
	schreibe(t, cache, "sub/b", strings.Repeat("b", 20))
	schreibe(t, cache, "sub/tief/c", "c")
	gross := filepath.Join(t.TempDir(), "gross")
	if err := os.WriteFile(gross, make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(gross, filepath.Join(cache, "verweis")); err != nil {
		t.Fatal(err)
	}
	e := &echt{env: env, cache: cache, bauZeit: time.Minute}
	if g := e.CacheGroesse(); g != 31 {
		t.Errorf("Größe %d statt 31", g)
	}
	if g := (&echt{cache: filepath.Join(cache, "fehlt")}).CacheGroesse(); g != 0 {
		t.Errorf("Größe eines fehlenden Caches %d", g)
	}
	if err := e.CacheLeeren(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := zeilen(t, log); got[0] != "ARGS clean -cache" {
		t.Errorf("Leeren ruft %q", got)
	}
}

// TestKopieUmfang (SPEC-210 Nr. 4 a): das Wurzelmodul kopiert go.mod, go.sum, src und tests; ein anderes Modul sein
// ganzes Verzeichnis; verschachtelte Module nie; Modus und Verweise bleiben.
func TestKopieUmfang(t *testing.T) {
	baum := t.TempDir()
	datei := func(rel, inhalt string, modus os.FileMode) {
		schreibe(t, baum, rel, inhalt)
		if err := os.Chmod(unter(baum, rel), modus); err != nil {
			t.Fatal(err)
		}
	}
	datei("go.mod", "module r\n", 0o644)
	datei("go.sum", "s\n", 0o644)
	datei("src/a.go", "package a\n", 0o755)
	datei("src/b.go", "package a\n", 0o644)
	datei("src/c.go", "package a\n", 0o600)
	datei("src/inner/go.mod", "module i\n", 0o644)
	datei("src/inner/x.go", "package i\n", 0o644)
	datei("tests/t.go", "package t\n", 0o644)
	datei("docs/readme.md", "d\n", 0o644)
	datei("tools/m/go.mod", "module m\n", 0o644)
	datei("tools/m/m.go", "package m\n", 0o644)
	datei("tools/m/run.sh", "x\n", 0o755)
	datei("tools/m/testdata/d.txt", "d\n", 0o644)
	datei("tools/m/sub/go.mod", "module s\n", 0o644)
	datei("tools/m/sub/s.go", "package s\n", 0o644)
	for _, l := range [][2]string{{"a.go", "src/link"}, {"m.go", "tools/m/link"}} {
		if err := os.Symlink(l[0], unter(baum, l[1])); err != nil {
			t.Fatal(err)
		}
	}
	liste := func(m Modul) map[string]string {
		ziel := filepath.Join(t.TempDir(), "k")
		if err := (echteDateien{}).Kopieren(baum, m, ziel); err != nil {
			t.Fatal(err)
		}
		r := map[string]string{}
		_ = filepath.WalkDir(ziel, func(p string, d fs.DirEntry, _ error) error {
			rel, _ := filepath.Rel(ziel, p)
			i, _ := os.Lstat(p)
			switch {
			case i.Mode()&fs.ModeSymlink != 0:
				l, _ := os.Readlink(p)
				r[rel] = "-> " + l
			case i.Mode().IsRegular():
				r[rel] = fmt.Sprintf("%v", i.Mode().Perm())
			}
			return nil
		})
		return r
	}
	wurzel := map[string]string{"go.mod": "-rw-r--r--", "go.sum": "-rw-r--r--", "src/a.go": "-rwxr-xr-x", "src/b.go": "-rw-r--r--",
		"src/c.go": "-rw-------", "src/link": "-> a.go", "tests/t.go": "-rw-r--r--"}
	if got := liste(Modul{}); !reflect.DeepEqual(got, wurzel) {
		t.Errorf("Wurzelmodul:\n got %v\nwant %v", got, wurzel)
	}
	modul := map[string]string{"go.mod": "-rw-r--r--", "m.go": "-rw-r--r--", "run.sh": "-rwxr-xr-x", "link": "-> m.go", "testdata/d.txt": "-rw-r--r--"}
	if got := liste(Modul{Dir: "tools/m"}); !reflect.DeepEqual(got, modul) {
		t.Errorf("Modul tools/m:\n got %v\nwant %v", got, modul)
	}
}

// TestSignalInNullprobe (SPEC-210 Nr. 10 a): ein Signal während der Nullprobe ist Exit 3, jede Mutation abbruch, kein
// Ausgangsstand rot.
func TestSignalInNullprobe(t *testing.T) {
	baum := fixtureBaum(t)
	ctx, cancel := context.WithCancel(t.Context())
	fp := &fakeP{bauAus: func(b Bau) Ausgang {
		if istNull(b) {
			cancel()
			return Ausgang{Abgebrochen: true}
		}
		return Ausgang{}
	}}
	w, aus, fehl := fakeWerk(t, baum, fp)
	code := w.Lauf(ctx, muts(t, baum, mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere),
		mut("b", "src/p/a.go", "return a - b", "return a + b", "TestSub", klasseAndere)))
	if r := ergebnisse(aus.String()); code != 3 || r["a"] != ergAbbruch || r["b"] != ergAbbruch || strings.Contains(fehl.String(), "Ausgangsstand") || len(fp.mutStarts()) != 0 {
		t.Errorf("Exit %d, %v, stderr %q", code, r, fehl.String())
	}
}

// TestDiagnoseAlle (SPEC-210 Nr. 1 c): zu uebersetzungsfehler, rot-paket, flatternd und abbruch stehen die ersten
// 40 Zeilen der Ausgabe, die das Ergebnis bestimmte; bei flatternd beide Arme.
func TestDiagnoseAlle(t *testing.T) {
	baum := fixtureBaum(t)
	var text strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&text, "zeile %d\n", i)
	}
	null := func(s Start) bool { return s.Nullprobe }
	faelle := []struct {
		erg, klasse string
		fp          *fakeP
	}{
		{ergUebersetz, klasseAndere, &fakeP{bauAus: func(b Bau) Ausgang { return Ausgang{Code: map[bool]int{false: 1}[istNull(b)], Text: text.String()} }}},
		{ergRotPaket, klasseAndere, &fakeP{startAus: func(s Start) Ausgang {
			return Ausgang{Code: map[bool]int{true: 1}[!null(s) && !hatArg(s, "-test.run")], Text: text.String()}
		}}},
		{ergFlattern, klasseNebenlaeufig, &fakeP{startAus: func(s Start) Ausgang {
			return Ausgang{Code: map[bool]int{true: 1}[!null(s) && s.Detektor], Text: text.String()}
		}}},
		{ergAbbruch, klasseAndere, &fakeP{startAus: func(s Start) Ausgang { return Ausgang{Abgebrochen: !null(s), Text: text.String()} }}},
		{ergAbbruch, klasseNebenlaeufig, &fakeP{startAus: func(s Start) Ausgang { return Ausgang{Abgebrochen: !null(s), Text: text.String()} }}},
	}
	// Die Nullprobe antwortet wie der Fake ohne Eingriff.
	nullAntwort := func(fp *fakeP) {
		if inner := fp.startAus; inner != nil {
			fp.startAus = func(s Start) Ausgang {
				if s.Nullprobe {
					return (&fakeP{}).Starten(context.Background(), s)
				}
				return inner(s)
			}
		}
	}
	for _, f := range faelle {
		nullAntwort(f.fp)
		w, aus, fehl := fakeWerk(t, baum, f.fp)
		w.Lauf(context.Background(), muts(t, baum, mitRate(mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", f.klasse), 2, 2)))
		z := strings.Split(strings.TrimSpace(fehl.String()), "\n")
		if ergebnisse(aus.String())["M"] != f.erg || len(z) != 40 || z[0] != "M: zeile 1" || z[39] != "M: zeile 40" {
			t.Errorf("%s: Ergebnis %s, %d Zeilen auf stderr", f.erg, ergebnisse(aus.String())["M"], len(z))
		}
	}
	// flatternd: ein Prozess des Arms ohne Detektor ist rot, der Arm mit Detektor ganz; beide Texte stehen auf stderr.
	var erster atomic.Bool
	fp := &fakeP{startAus: func(s Start) Ausgang {
		switch {
		case s.Detektor:
			return Ausgang{Code: 1, Text: "mit-zeile\n"}
		case hatArg(s, "-test.run") && erster.CompareAndSwap(false, true):
			return Ausgang{Code: 1, Text: "ohne-zeile\n"}
		}
		return Ausgang{}
	}}
	nullAntwort(fp)
	w, _, fehl := fakeWerk(t, baum, fp)
	w.Lauf(context.Background(), muts(t, baum, mitRate(mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseNebenlaeufig), 1, 2)))
	if fehl.String() != "M: mit-zeile\nM: ohne-zeile\n" {
		t.Errorf("flatternd: stderr %q", fehl.String())
	}
}

// TestListeZeichen (SPEC-210 Nr. 2 b): name erlaubt genau A-Za-z0-9._-, test nach dem Präfix Test genau A-Za-z0-9_.
func TestListeZeichen(t *testing.T) {
	baum := listeBaum(t)
	gut := func(name, test string) bool {
		_, b := leseListe(listeJSON(t, mut(name, "src/p/a.go", "return a + b", "return a - b", test, klasseAndere)), baum, true)
		return len(b) == 0
	}
	var erlaubt []rune
	for _, r := range []struct{ von, bis rune }{{'a', 'z'}, {'A', 'Z'}, {'0', '9'}, {'.', '.'}, {'_', '_'}, {'-', '-'}} {
		for c := r.von; c <= r.bis; c++ {
			erlaubt = append(erlaubt, c)
		}
	}
	if len(erlaubt) != 65 {
		t.Fatalf("%d Zeichen in der Menge", len(erlaubt))
	}
	for _, c := range erlaubt {
		if !gut("x"+string(c), "TestAdd") {
			t.Errorf("Zeichen %q im Namen abgelehnt", c)
		}
		if c != '.' && c != '-' && !gut("x", "Test"+string(c)) {
			t.Errorf("Zeichen %q im Test abgelehnt", c)
		}
	}
	for _, c := range "@[`{/:+ !ä\\\"" {
		if gut("x"+string(c), "TestAdd") {
			t.Errorf("Zeichen %q im Namen angenommen", c)
		}
	}
	for _, c := range ".-@[`{/:+ ä" {
		if gut("x", "Test"+string(c)) {
			t.Errorf("Zeichen %q im Test angenommen", c)
		}
	}
	for _, p := range []string{"T", "Tes", "test", "TEST", "Tes_t", "_Test", "XTest"} {
		if gut("x", p+"Add") {
			t.Errorf("Präfix von %q angenommen", p+"Add")
		}
	}
	if !gut("x", "Test") || !gut("x", "Test1") {
		t.Error("Test und Test1 abgelehnt")
	}
}

// TestBinariesWerdenEntfernt (SPEC-210 Nr. 4 b, 7 b): jede Mutation entfernt Kopie, Binary und Race-Binary, die
// Nullprobe ihr Race-Binary; ein Arm mit 0 baut kein Binary mit Detektor.
func TestBinariesWerdenEntfernt(t *testing.T) {
	baum := fixtureBaum(t)
	fp := &fakeP{}
	w, _, _ := fakeWerk(t, baum, fp)
	d := &fakeD{}
	w.D = d
	w.Lauf(context.Background(), muts(t, baum,
		mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere),
		mitRate(mut("b", "src/p/a.go", "return a - b", "return a + b", "TestSub", klasseNebenlaeufig), 1, 1),
		mitRate(mut("c", "src/p/a.go", "return a }", "return a + 0 }", "TestId", klasseNebenlaeufig), 0, 1)))
	l := filepath.Join(w.Ablage, "lauf")
	for _, p := range []string{"1", "1.test", "2", "2.test", "2.test.race", "3", "3.test", "null0", "null0.test", "null0.test.race"} {
		if d.index("entf:"+filepath.Join(l, p)) < 0 {
			t.Errorf("%s wird nicht entfernt", p)
		}
	}
	race := map[string]bool{}
	for _, b := range fp.baue {
		if b.Detektor && !istNull(b) {
			race[filepath.Base(b.Ziel)] = true
		}
	}
	if !reflect.DeepEqual(race, map[string]bool{"2.test.race": true}) {
		t.Errorf("Baue mit Detektor: %v", race)
	}
}
