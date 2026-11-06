// Command mutationslauf arbeitet eine Liste von Mutationen in einem Lauf ab (SPEC-210): je Mutation
// eine Wegwerfkopie, ein Test-Binary, der zuständige Test und das Paket, bei der Klasse nebenläufig
// die Rate mit und ohne Detektor. Aufruf: `mutationslauf <Liste> <Baum> [<Ablage> [<Image>]]`
// (Vorgaben /work und /src; ein leeres <Image> schaltet den Abgleich von go.mod und go.sum aus); den
// Container startet der Treiber tools/harness/mutationslauf.sh. Das Programm ändert den Baum nicht.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	exitEingabe   = 2
	cacheGrenze   = 3 << 29 // 1,5 GiB
	grenzeVorgabe = 3600
	grenzeMax     = 14400
	aufrufText    = "Aufruf: mutationslauf <Liste> <Baum> [<Ablage> [<Image>]]\n"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	code := run(ctx, os.Args[1:], os.Environ(), os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// lookup liest eine Umgebungsvariable aus einer Liste `NAME=Wert`.
func lookup(env []string, name string) (string, bool) {
	wert, da := "", false
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, name+"="); ok {
			wert, da = v, true
		}
	}
	return wert, da
}

// umgebung liest MUTATIONSLAUF_GRENZE und MUTATIONSLAUF_KALT (SPEC-210 Nr. 1 a, 8 c, 9 c); der
// zweite Wert ist der Name einer ungültigen Variable.
func umgebung(env []string) (time.Duration, bool, string) {
	grenze := grenzeVorgabe
	if v, da := lookup(env, "MUTATIONSLAUF_GRENZE"); da {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > grenzeMax || strconv.Itoa(n) != v {
			return 0, false, "MUTATIONSLAUF_GRENZE"
		}
		grenze = n
	}
	v, _ := lookup(env, "MUTATIONSLAUF_KALT")
	if v != "" && v != "1" {
		return 0, false, "MUTATIONSLAUF_KALT"
	}
	return time.Duration(grenze) * time.Second, v == "1", ""
}

// imageAbgleich vergleicht go.mod und go.sum des Baums mit denen des Images (SPEC-210 Nr. 3 a).
func imageAbgleich(baum, image string) bool {
	for _, f := range []string{"go.mod", "go.sum"} {
		a, errA := os.ReadFile(filepath.Join(baum, f))
		b, errB := os.ReadFile(filepath.Join(image, f))
		if (errA == nil) != (errB == nil) || !bytes.Equal(a, b) {
			return false
		}
	}
	return true
}

// ladeLast liest den ersten Wert von /proc/loadavg.
func ladeLast() string {
	roh, err := os.ReadFile("/proc/loadavg")
	if f := strings.Fields(string(roh)); err == nil && len(f) > 0 {
		return f[0]
	}
	return "-"
}

func run(ctx context.Context, args, env []string, stdout, stderr io.Writer) int {
	if len(args) < 2 || len(args) > 4 {
		_, _ = fmt.Fprint(stderr, aufrufText)
		return exitEingabe
	}
	grenze, kalt, ungueltig := umgebung(env)
	if ungueltig != "" {
		_, _ = fmt.Fprintf(stderr, "mutationslauf: Umgebung ungültig: %s\n", ungueltig)
		return exitEingabe
	}
	roh, err := os.ReadFile(args[0])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "mutationslauf: Liste fehlt: %s\n", args[0])
		return exitEingabe
	}
	baum, ablage, image := args[1], "/work", "/src"
	if len(args) >= 3 {
		ablage = args[2]
	}
	if len(args) == 4 {
		image = args[3]
	}
	_, kein := exec.LookPath("gcc")
	muts, befunde := leseListe(roh, baum, kein == nil)
	for _, b := range befunde {
		_, _ = fmt.Fprintf(stderr, "mutationslauf: Eingabe: %s\n", b)
	}
	if len(befunde) > 0 {
		return exitEingabe
	}
	if image != "" && !imageAbgleich(baum, image) {
		_, _ = fmt.Fprintln(stderr, "mutationslauf: Image veraltet: go.mod oder go.sum weichen ab")
		return exitEingabe
	}
	cache, _ := lookup(env, "GOCACHE")
	if d, err := os.UserCacheDir(); cache == "" && err == nil {
		cache = filepath.Join(d, "go-build")
	}
	if kalt {
		cache = filepath.Join(ablage, "gocache")
	}
	penv := append(append([]string(nil), env...), "GOCACHE="+cache, "GOFLAGS=-mod=readonly", "GOPROXY=off", "GOTOOLCHAIN=local")
	w := &Werk{
		Konfig: Konfig{Baum: baum, Ablage: ablage, TestZeit: 2 * time.Minute, Gnade: 30 * time.Second,
			CacheGrenze: cacheGrenze, CPUs: runtime.GOMAXPROCS(0)},
		P:    &echt{env: penv, cache: cache, bauZeit: 10 * time.Minute},
		D:    echteDateien{},
		Last: ladeLast, Jetzt: time.Now, Aus: stdout, Fehl: stderr,
	}
	ctx, cancel := context.WithTimeout(ctx, grenze)
	defer cancel()
	return w.Lauf(ctx, muts)
}
