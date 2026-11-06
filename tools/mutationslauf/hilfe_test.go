package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

const aGo = `package p

import "sync"

// Add addiert.
func Add(a, b int) int { return a + b }

// Sub subtrahiert.
func Sub(a, b int) int { return a - b }

// Id gibt a zurück.
func Id(a int) int { return a }

// Schleife zählt bis n.
func Schleife(n int) int {
	i := 0
	for i < n {
		i++
	}
	return i
}

// Zaehle zählt aus zwei Goroutinen unter einem Mutex.
func Zaehle() int {
	var mu sync.Mutex
	var wg sync.WaitGroup
	z := 0
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			z++
			mu.Unlock()
		}()
	}
	wg.Wait()
	return z
}
`

const aTestGo = `package p

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("Add")
	}
}

func TestAddx(t *testing.T) {
	if Add(2, 2) != 4 {
		t.Fatal("Addx")
	}
}

func TestSub(t *testing.T) {
	if Sub(5, 3) != 2 {
		t.Fatal("Sub")
	}
}

func TestId(t *testing.T) {
	if Id(7) != 7 {
		t.Fatal("Id")
	}
}

func TestSchleife(t *testing.T) {
	if Schleife(10) != 10 {
		t.Fatal("Schleife")
	}
}

func TestZaehle(t *testing.T) {
	if Zaehle() != 2 {
		t.Fatal("Zaehle")
	}
}
`

// hGo und hTestGo: ein Test-Binary, das vor allen Tests hängt, wenn Haengt() wahr ist; -test.timeout greift dort nicht.
const hGo = "package h\n\n// Haengt sagt, ob der Test-Prozess hängt.\nfunc Haengt() bool { return false }\n"
const hTestGo = "package h\n\nimport (\n\t\"os\"\n\t\"testing\"\n\t\"time\"\n)\n\nfunc TestMain(m *testing.M) {\n\tif Haengt() {\n\t\ttime.Sleep(time.Hour)\n\t}\n\tos.Exit(m.Run())\n}\n\nfunc TestH(t *testing.T) {}\n"

// wGo und wTestGo: ein Test, der hängt und einen Kindprozess startet, wenn Haengt() wahr ist.
const wGo = "package w\n\n// Haengt sagt, ob der Test hängt.\nfunc Haengt() bool { return false }\n"
const wTestGo = `package w

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestW(t *testing.T) {
	if !Haengt() {
		return
	}
	d := os.Getenv("FX_MARKER")
	k := exec.Command("sleep", "300")
	if err := k.Start(); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(d, "kind"), []byte(strconv.Itoa(k.Process.Pid)), 0o644)
	_ = os.WriteFile(filepath.Join(d, "test"), []byte(strconv.Itoa(os.Getpid())), 0o644)
	time.Sleep(time.Hour)
}
`

func schreibe(t *testing.T, dir, rel, inhalt string) {
	t.Helper()
	p := unter(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(inhalt), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixtureBaum legt das Wegwerfmodul mit bekannten Mutationen an.
func fixtureBaum(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	schreibe(t, d, "go.mod", "module fx\n\ngo 1.27\n")
	schreibe(t, d, "src/p/a.go", aGo)
	schreibe(t, d, "src/p/a_test.go", aTestGo)
	schreibe(t, d, "src/q/q.go", "package q\n\n// Q liefert eins.\nfunc Q() int { return 1 }\n")
	schreibe(t, d, "src/r/r.go", "package r\n")
	schreibe(t, d, "src/r/r_test.go", "package r\n\nimport \"testing\"\n\nfunc hilfe(t *testing.T) {}\n")
	schreibe(t, d, "src/s/s.go", "package s\n\n// S liefert eins.\nfunc S() int { return 1 }\n")
	schreibe(t, d, "src/s/s_test.go", "package s\n\nimport \"testing\"\n\nfunc TestRot(t *testing.T) {\n\tt.Fatal(\"rot\")\n}\n")
	schreibe(t, d, "src/u/u.go", "package u\n\n// U übersetzt im Ausgangsstand nicht.\nfunc U() int { return \"x\" }\n")
	schreibe(t, d, "src/u/u_test.go", "package u\n\nimport \"testing\"\n\nfunc TestU(t *testing.T) {\n\t_ = U()\n}\n")
	schreibe(t, d, "src/v/v.go", "package v\n\nimport \"sync\"\n\n// Z zählt ohne Schutz aus zwei Goroutinen; nur der Detektor meldet es.\nfunc Z() int {\n\tz := 0\n\tvar wg sync.WaitGroup\n\tfor range 2 {\n\t\twg.Add(1)\n\t\tgo func() {\n\t\t\tdefer wg.Done()\n\t\t\tz++\n\t\t}()\n\t}\n\twg.Wait()\n\treturn z\n}\n")
	schreibe(t, d, "src/v/v_test.go", "package v\n\nimport \"testing\"\n\nfunc TestV(t *testing.T) {\n\tif Z() < 1 {\n\t\tt.Fatal(\"Z\")\n\t}\n}\n")
	schreibe(t, d, "src/h/h.go", hGo)
	schreibe(t, d, "src/h/h_test.go", hTestGo)
	schreibe(t, d, "src/w/w.go", wGo)
	schreibe(t, d, "src/w/w_test.go", wTestGo)
	return d
}

// ml ist ein Listeneintrag als JSON-Objekt.
type ml map[string]any

func mut(name, datei, alt, neu, test, klasse string) ml {
	return ml{"name": name, "datei": datei, "alt": alt, "neu": neu, "test": test, "klasse": klasse}
}

func mitRate(m ml, d, o int) ml {
	if m["klasse"] == klasseNebenlaeufig {
		m["n_detektor"], m["n_ohne"] = d, o
	}
	return m
}

func listeJSON(t *testing.T, ms ...ml) []byte {
	t.Helper()
	roh, err := json.Marshal(map[string]any{"mutationen": ms})
	if err != nil {
		t.Fatal(err)
	}
	return roh
}

func muts(t *testing.T, baum string, ms ...ml) []Mutation {
	t.Helper()
	r, befunde := leseListe(listeJSON(t, ms...), baum, true)
	if len(befunde) > 0 {
		t.Fatalf("Liste hat Befunde: %v", befunde)
	}
	return r
}

// Die acht Arten der Mutation (SPEC-210 Nr. 6 c), je Klasse.
func arten(klasse string) []ml {
	a := "src/p/a.go"
	return []ml{
		mut("zustaendig", a, "return a + b", "return a - b", "TestAdd", klasse),
		mut("anderer", a, "return a - b", "return a + b", "TestAdd", klasse),
		mut("aequivalent", a, "return a }", "return a + 0 }", "TestId", klasse),
		mut("syntax", a, "return a + b", "return a +", "TestAdd", klasse),
		mut("typ", a, "return a + b", "return \"x\"", "TestAdd", klasse),
		mut("panik", a, "return a }", "panic(\"boom\") }", "TestId", klasse),
		mut("endlos", a, "i++", "", "TestSchleife", klasse),
		mut("rennen", a, "mu.Lock()\n\t\t\tz++\n\t\t\tmu.Unlock()", "_ = &mu\n\t\t\tz++", "TestZaehle", klasse),
	}
}

// ergebnisse liest name -> ergebnis aus den Ergebniszeilen.
func ergebnisse(out string) map[string]string {
	r := map[string]string{}
	for _, z := range strings.Split(out, "\n") {
		if f := strings.Fields(z); len(f) > 2 && f[0] == "mutation" {
			r[f[1]] = strings.TrimPrefix(f[2], "ergebnis=")
		}
	}
	return r
}

// feldWert liest key=Wert aus einer Ergebniszeile.
func feldWert(zeile, key string) string {
	for _, f := range strings.Fields(zeile) {
		if v, ok := strings.CutPrefix(f, key+"="); ok {
			return v
		}
	}
	return ""
}

func zeileVon(out, name string) string {
	for _, z := range strings.Split(out, "\n") {
		if strings.HasPrefix(z, "mutation "+name+" ") {
			return z
		}
	}
	return ""
}

// baumHash bildet eine Prüfsumme über Pfade, Modi und Inhalte eines Baums.
func baumHash(t *testing.T, wurzel string) string {
	t.Helper()
	h := sha256.New()
	var pfade []string
	_ = filepath.WalkDir(wurzel, func(p string, d fs.DirEntry, err error) error {
		if err == nil {
			pfade = append(pfade, p)
		}
		return nil
	})
	sort.Strings(pfade)
	for _, p := range pfade {
		i, err := os.Lstat(p)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fmt.Fprintf(h, "%s %v %d\n", p, i.Mode(), i.Size())
		if i.Mode().IsRegular() {
			roh, _ := os.ReadFile(p)
			_, _ = h.Write(roh)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// echtWerk baut ein Werk mit den echten Ports und kurzen Fristen; Ausgaben liegen in den Puffern.
func echtWerk(t *testing.T, baum string) (*Werk, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var aus, fehl bytes.Buffer
	env := append(os.Environ(), "GOFLAGS=-mod=readonly", "GOPROXY=off", "GOTOOLCHAIN=local")
	cache := ""
	if c, err := os.UserCacheDir(); err == nil {
		cache = filepath.Join(c, "go-build")
	}
	w := &Werk{
		Konfig: Konfig{Baum: baum, Ablage: t.TempDir(), TestZeit: 2 * time.Second, Gnade: 3 * time.Second,
			CacheGrenze: cacheGrenze, CPUs: 4},
		P: &echt{env: env, cache: cache, bauZeit: 5 * time.Minute}, D: echteDateien{},
		Last: func() string { return "0.00" }, Jetzt: time.Now, Aus: &aus, Fehl: &fehl,
	}
	return w, &aus, &fehl
}

// orakel fährt den Einzellauf `go test -count=1 [-race]` auf einer frischen Kopie des Baums.
func orakel(t *testing.T, baum string, m ml, race bool) string {
	t.Helper()
	kopie := t.TempDir()
	if out, err := exec.Command("cp", "-a", baum+"/.", kopie).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v %s", err, out)
	}
	p := filepath.Join(kopie, filepath.FromSlash(m["datei"].(string)))
	roh, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(roh), m["alt"].(string), m["neu"].(string), 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	gehe := func(extra ...string) (string, bool) {
		args := append([]string{"test", "-count=1", "-timeout", "2s"}, extra...)
		if race {
			args = append(args, "-race")
		}
		c := exec.Command("go", append(args, "./src/p")...)
		c.Dir = kopie
		c.Env = append(os.Environ(), "GOFLAGS=-mod=readonly", "GOPROXY=off", "GOTOOLCHAIN=local", fmt.Sprintf("CGO_ENABLED=%d", map[bool]int{false: 0, true: 1}[race]))
		out, err := c.CombinedOutput()
		return string(out), err == nil
	}
	out, ok := gehe()
	switch {
	case ok:
		return ergGruen
	case strings.Contains(out, "build failed") || strings.Contains(out, "setup failed"):
		return ergUebersetz
	}
	if _, ok := gehe("-run", "^"+m["test"].(string)+"$"); !ok {
		return ergRot
	}
	return ergRotPaket
}

// lebt sagt, ob ein Prozess läuft; ein Zombie zählt als beendet.
func lebt(pid int) bool {
	roh, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	s := string(roh)
	return s[strings.LastIndex(s, ")")+2] != 'Z'
}

// fakeP ist der Ersatz für den Prozess-Port: zählt Baue und Starts, misst Gleichzeitigkeit, hält Argumente fest.
type fakeP struct {
	mu                   sync.Mutex
	baue                 []Bau
	starts               []Start
	aktivBau, aktivStart int
	maxBau, maxStart     int
	verletzung           string
	bauAus               func(Bau) Ausgang
	startAus             func(Start) Ausgang
	sperre               int // die ersten n Starts der Mutationen warten, bis n zugleich laufen
	mutN                 int
	bauHalt              time.Duration
	cache                int64
	geleert              int
}

func (f *fakeP) Bauen(_ context.Context, b Bau) Ausgang {
	f.mu.Lock()
	if f.aktivBau > 0 || f.aktivStart > 0 {
		f.verletzung = "Bau neben Bau oder Start"
	}
	f.aktivBau++
	f.maxBau = max(f.maxBau, f.aktivBau)
	f.baue = append(f.baue, b)
	f.mu.Unlock()
	time.Sleep(f.bauHalt)
	defer func() { f.mu.Lock(); f.aktivBau--; f.mu.Unlock() }()
	if f.bauAus != nil {
		return f.bauAus(b)
	}
	return Ausgang{}
}

func (f *fakeP) Starten(_ context.Context, s Start) Ausgang {
	f.mu.Lock()
	if f.aktivBau > 0 {
		f.verletzung = "Start neben Bau"
	}
	f.aktivStart++
	f.maxStart = max(f.maxStart, f.aktivStart)
	f.starts = append(f.starts, s)
	nr := 0
	if !s.Nullprobe {
		f.mutN++
		nr = f.mutN
	}
	f.mu.Unlock()
	if nr > 0 && nr <= f.sperre {
		t0 := time.Now()
		for time.Since(t0) < 5*time.Second {
			f.mu.Lock()
			a := f.aktivStart
			f.mu.Unlock()
			if a >= f.sperre {
				break
			}
			time.Sleep(time.Millisecond)
		}
		time.Sleep(150 * time.Millisecond)
	}
	defer func() { f.mu.Lock(); f.aktivStart--; f.mu.Unlock() }()
	if f.startAus != nil {
		return f.startAus(s)
	}
	if s.Nullprobe && len(s.Args) == 2 && s.Args[0] == "-test.list" {
		return Ausgang{Text: strings.TrimSuffix(strings.TrimPrefix(s.Args[1], "^"), "$") + "\n"}
	}
	return Ausgang{}
}

func (f *fakeP) CacheGroesse() int64 { return f.cache }

func (f *fakeP) CacheLeeren(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.geleert++
	return nil
}

// mutStarts sind die Starts der Mutationen ohne die der Nullprobe.
func (f *fakeP) mutStarts() []Start {
	f.mu.Lock()
	defer f.mu.Unlock()
	var r []Start
	for _, s := range f.starts {
		if !s.Nullprobe {
			r = append(r, s)
		}
	}
	return r
}

func hatArg(s Start, a string) bool {
	for _, x := range s.Args {
		if strings.HasPrefix(x, a) {
			return true
		}
	}
	return false
}

// fakeWerk baut ein Werk mit dem Fake-Port, echten Dateien und einer Uhr, die je Aufruf eine Sekunde vorrückt;
// die Last ist die Zahl der bisherigen Starts, so zeigt `last` den Zeitpunkt der Messung.
func fakeWerk(t *testing.T, baum string, f *fakeP) (*Werk, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var aus, fehl bytes.Buffer
	var mu sync.Mutex
	uhr := time.Unix(1000, 0)
	w := &Werk{
		Konfig: Konfig{Baum: baum, Ablage: t.TempDir(), TestZeit: 2 * time.Minute, Gnade: 30 * time.Second,
			CacheGrenze: 1610612736, CPUs: 4},
		P: f, D: echteDateien{},
		Last:  func() string { f.mu.Lock(); defer f.mu.Unlock(); return fmt.Sprintf("S%d", len(f.starts)) },
		Jetzt: func() time.Time { mu.Lock(); defer mu.Unlock(); uhr = uhr.Add(time.Second); return uhr },
		Aus:   &aus, Fehl: &fehl,
	}
	return w, &aus, &fehl
}
