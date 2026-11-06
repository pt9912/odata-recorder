package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const stubSkript = `#!/bin/sh
d=$STUB_DIR
echo "$*" >> "$d/log"
case $1 in
	image) [ "${STUB_IMAGE:-ja}" = ja ] ;;
	ps) echo ${STUB_PS:-} ;;
	create) [ "${STUB_CREATE:-ja}" = ja ] || exit 1; echo cid1 ;;
	cp)
		[ "${STUB_CP:-ja}" = ja ] || exit 1
		if [ "$2" = - ]; then tar -tf - >> "$d/tar"; fi ;;
	start)
		case ${STUB_START:-ok} in
			ok) exit 0 ;;
			exit2) exit 2 ;;
			block)
				i=0
				while [ ! -e "$d/killed" ] && [ $i -lt 1500 ]; do sleep 0.1; i=$((i + 1)); done
				exit 3 ;;
		esac ;;
	kill) echo "$*" > "$d/killed" ;;
esac
`

// treiberUmgebung ist ein Wegwerfbaum mit dem Treiber, ein Ersatz-Skript für docker und der Zustand des Ersatzes.
type treiberUmgebung struct {
	baum, stub, zustand, liste string
}

func neueTreiberUmgebung(t *testing.T) *treiberUmgebung {
	t.Helper()
	baum := t.TempDir()
	skript, err := os.ReadFile(filepath.Join("..", "harness", "mutationslauf.sh"))
	if err != nil {
		t.Fatal(err)
	}
	schreibe(t, baum, "tools/harness/mutationslauf.sh", string(skript))
	schreibe(t, baum, "tools/mutationslauf/main.go", "package main\n")
	schreibe(t, baum, ".git/HEAD", "ref\n")
	schreibe(t, baum, "src/a.go", "package a\n")
	u := &treiberUmgebung{baum: baum, stub: filepath.Join(t.TempDir(), "docker"), zustand: t.TempDir()}
	if err := os.WriteFile(u.stub, []byte(stubSkript), 0o755); err != nil {
		t.Fatal(err)
	}
	u.liste = filepath.Join(t.TempDir(), "liste.json")
	if err := os.WriteFile(u.liste, listeJSON(t, mut("a", "src/a.go", "x", "y", "TestA", klasseAndere)), 0o644); err != nil {
		t.Fatal(err)
	}
	return u
}

func (u *treiberUmgebung) cmd(env ...string) *exec.Cmd {
	c := exec.Command("sh", filepath.Join(u.baum, "tools", "harness", "mutationslauf.sh"), u.liste)
	c.Env = append(os.Environ(), "DOCKER="+u.stub, "STUB_DIR="+u.zustand)
	c.Env = append(c.Env, env...)
	return c
}

// lauf führt den Treiber aus und liefert Exit-Code, stdout und stderr.
func (u *treiberUmgebung) lauf(t *testing.T, env ...string) (int, string, string) {
	t.Helper()
	var aus, fehl bytes.Buffer
	c := u.cmd(env...)
	c.Stdout, c.Stderr = &aus, &fehl
	err := c.Run()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return c.ProcessState.ExitCode(), aus.String(), fehl.String()
}

func (u *treiberUmgebung) log() []string {
	roh, _ := os.ReadFile(filepath.Join(u.zustand, "log"))
	return strings.Split(strings.TrimSpace(string(roh)), "\n")
}

func (u *treiberUmgebung) aufrufe(verb string) []string {
	var r []string
	for _, z := range u.log() {
		if z == verb || strings.HasPrefix(z, verb+" ") {
			r = append(r, z)
		}
	}
	return r
}

// TestTreiberOptionen (SPEC-210 Nr. 3 b): der Container ist netzlos; kein Aufruf außer `rm -v` (Volumes des Containers
// entfernen) trägt -v, --volume, --mount oder --tmpfs.
func TestTreiberOptionen(t *testing.T) {
	for _, start := range []string{"ok", "exit2"} {
		t.Run(start, func(t *testing.T) {
			u := neueTreiberUmgebung(t)
			u.lauf(t, "STUB_START="+start, "STUB_PS=alt1")
			if c := u.aufrufe("create"); len(c) != 1 || !strings.Contains(c[0], "--network none") {
				t.Errorf("create: %q", c)
			}
			for _, z := range u.log() {
				if strings.HasPrefix(z, "rm ") {
					continue
				}
				for _, f := range strings.Fields(z) {
					for _, v := range []string{"-v", "--volume", "--mount", "--tmpfs"} {
						if f == v || strings.HasPrefix(f, v+"=") {
							t.Errorf("Aufruf mit %s: %q", v, z)
						}
					}
				}
			}
		})
	}
}

// TestEinContainerJeAufruf (SPEC-210 Nr. 3 b): ein create und ein start je Aufruf, unabhängig von der Zahl der Mutationen.
func TestEinContainerJeAufruf(t *testing.T) {
	u := neueTreiberUmgebung(t)
	var ms []ml
	for i := range 5 {
		ms = append(ms, mut(fmt.Sprintf("m%d", i), "src/a.go", "x", "y", "TestA", klasseAndere))
	}
	if err := os.WriteFile(u.liste, listeJSON(t, ms...), 0o644); err != nil {
		t.Fatal(err)
	}
	u.lauf(t)
	if c, s := u.aufrufe("create"), u.aufrufe("start"); len(c) != 1 || len(s) != 1 {
		t.Errorf("create %d-mal, start %d-mal", len(c), len(s))
	}
	if r := u.aufrufe("rm"); len(r) != 1 || r[0] != "rm -f -v cid1" {
		t.Errorf("rm: %q", r)
	}
	for _, v := range []string{"build", "commit", "tag", "rmi", "image rm", "volume"} {
		if len(u.aufrufe(v)) != 0 {
			t.Errorf("der Treiber ruft docker %s", v)
		}
	}
}

// TestTreiberRaeumtAuf (SPEC-210 Nr. 10 a, c): der Container wird nach Erfolg, bei Exit 2 und nach einem Fehler beim Kopieren entfernt;
// der Start entfernt nur Container mit dem Label, die nicht laufen.
func TestTreiberRaeumtAuf(t *testing.T) {
	faelle := []struct {
		name string
		env  []string
		rc   int
	}{{"Erfolg", []string{"STUB_START=ok"}, 0}, {"Exit 2", []string{"STUB_START=exit2"}, 2}, {"cp scheitert", []string{"STUB_CP=nein"}, 2}}
	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			u := neueTreiberUmgebung(t)
			rc, _, _ := u.lauf(t, append(f.env, "STUB_PS=alt1 alt2")...)
			r := u.aufrufe("rm")
			if rc != f.rc || len(r) != 2 || r[0] != "rm -f -v alt1 alt2" || r[1] != "rm -f -v cid1" {
				t.Errorf("Exit %d, rm: %q", rc, r)
			}
			p := u.aufrufe("ps")
			if len(p) != 1 || strings.Contains(p[0], "running") {
				t.Fatalf("ps: %q", p)
			}
			for _, w := range []string{"-aq", "label=odata-recorder.mutationslauf=1", "status=exited", "status=created", "status=dead"} {
				if !strings.Contains(p[0], w) {
					t.Errorf("ps ohne %s: %q", w, p[0])
				}
			}
		})
	}
	for _, sig := range []syscall.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			u := neueTreiberUmgebung(t)
			c := u.cmd("STUB_START=block")
			if err := c.Start(); err != nil {
				t.Fatal(err)
			}
			stop := time.Now().Add(20 * time.Second)
			for len(u.aufrufe("start")) == 0 && time.Now().Before(stop) {
				time.Sleep(20 * time.Millisecond)
			}
			time.Sleep(200 * time.Millisecond)
			_ = c.Process.Signal(sig)
			_ = c.Wait()
			if k := u.aufrufe("kill"); len(k) != 1 || k[0] != "kill --signal TERM cid1" {
				t.Errorf("kill: %q", k)
			}
			if r := u.aufrufe("rm"); len(r) != 1 || r[0] != "rm -f -v cid1" || c.ProcessState.ExitCode() != 3 {
				t.Errorf("rm: %q, Exit %d", r, c.ProcessState.ExitCode())
			}
		})
	}
}

// TestTreiberFrist (SPEC-210 Nr. 10 b): der Treiber tötet den Container 60 Sekunden nach der Gesamtgrenze.
func TestTreiberFrist(t *testing.T) {
	if testing.Short() {
		t.Skip("die Frist braucht 61 Sekunden")
	}
	u := neueTreiberUmgebung(t)
	t0 := time.Now()
	rc, _, _ := u.lauf(t, "STUB_START=block", "MUTATIONSLAUF_GRENZE=1")
	d, k := time.Since(t0), u.aufrufe("kill")
	if rc != 3 || len(k) != 1 || k[0] != "kill cid1" || d < 60*time.Second || d > 75*time.Second {
		t.Errorf("Exit %d, kill %q nach %v", rc, k, d)
	}
}

// TestTreiberSchreibtNichts (SPEC-210 Nr. 3 e): der Baum bleibt unverändert, auch bei Schreibschutz; der Treiber legt keine Datei an.
func TestTreiberSchreibtNichts(t *testing.T) {
	u := neueTreiberUmgebung(t)
	vor := baumHash(t, u.baum)
	modi := map[string]os.FileMode{}
	_ = filepath.WalkDir(u.baum, func(p string, d os.DirEntry, _ error) error {
		i, _ := d.Info()
		modi[p] = i.Mode().Perm()
		return nil
	})
	setze := func(schreibschutz bool) {
		for p, m := range modi {
			if schreibschutz {
				m &^= 0o222
			}
			_ = os.Chmod(p, m)
		}
	}
	setze(true)
	t.Cleanup(func() { setze(false) })
	rc, _, fehl := u.lauf(t)
	setze(false)
	if rc != 0 || fehl != "" {
		t.Errorf("Exit %d, stderr %q", rc, fehl)
	}
	if baumHash(t, u.baum) != vor {
		t.Error("der Treiber hat den Baum verändert")
	}
}

// TestTreiberUmgebung (SPEC-210 Nr. 1 a, 9 a): --memory und --memory-swap gleich dem Wert für 1g bis 4g; jeder andere Wert,
// auch der gesetzte leere, ist Exit 2 ohne Container; ebenso GRENZE und KALT.
func TestTreiberUmgebung(t *testing.T) {
	for wert, gml := range map[string]string{"1g": "768MiB", "2g": "1500MiB", "3g": "2304MiB", "4g": "3072MiB"} {
		u := neueTreiberUmgebung(t)
		rc, _, _ := u.lauf(t, "MUTATIONSLAUF_SPEICHER="+wert)
		c := strings.Join(u.aufrufe("create"), "")
		if rc != 0 || !strings.Contains(c, "--memory="+wert+" --memory-swap="+wert+" ") || !strings.Contains(c, "-e GOMEMLIMIT="+gml+" ") {
			t.Errorf("Speicher %s: Exit %d, create %q", wert, rc, c)
		}
	}
	ungueltig := map[string][]string{
		"MUTATIONSLAUF_SPEICHER": {"0g", "5g", "abc", "", "2G", "2", "10g"},
		"MUTATIONSLAUF_GRENZE":   {"0", "14401", "abc", "", "-1", "01", "3600s"},
		"MUTATIONSLAUF_KALT":     {"0", "2", "ja"},
	}
	for name, werte := range ungueltig {
		for _, w := range werte {
			u := neueTreiberUmgebung(t)
			rc, _, fehl := u.lauf(t, name+"="+w)
			if rc != 2 || !strings.Contains(fehl, "Umgebung ungültig: "+name) || len(u.aufrufe("create")) != 0 {
				t.Errorf("%s=%q: Exit %d, stderr %q", name, w, rc, fehl)
			}
		}
	}
	for _, g := range []string{"1", "3600", "14400"} {
		u := neueTreiberUmgebung(t)
		if rc, _, fehl := u.lauf(t, "MUTATIONSLAUF_GRENZE="+g); rc != 0 {
			t.Errorf("GRENZE=%s: Exit %d, %q", g, rc, fehl)
		}
	}
}

// TestTreiberAufrufUndImage (SPEC-210 Nr. 1 a, 3 a, 3 d): Aufruf, fehlendes Image, Baum über 256 MiB, Inhalt des Tar-Stroms.
func TestTreiberAufrufUndImage(t *testing.T) {
	u := neueTreiberUmgebung(t)
	for _, args := range [][]string{{}, {u.liste, u.liste}, {filepath.Join(u.zustand, "fehlt.json")}, {u.zustand}} {
		c := exec.Command("sh", append([]string{filepath.Join(u.baum, "tools", "harness", "mutationslauf.sh")}, args...)...)
		c.Env = append(os.Environ(), "DOCKER="+u.stub, "STUB_DIR="+u.zustand)
		var fehl bytes.Buffer
		c.Stderr = &fehl
		_ = c.Run()
		if c.ProcessState.ExitCode() != 2 || fehl.String() != "Aufruf: mutationslauf.sh <Liste>\n" {
			t.Errorf("Argumente %q: Exit %d, %q", args, c.ProcessState.ExitCode(), fehl.String())
		}
	}
	rc, _, fehl := u.lauf(t, "STUB_IMAGE=nein")
	if rc != 2 || fehl != "Image fehlt: odata-recorder:race (make test-race bzw. make test)\n" || len(u.aufrufe("create")) != 0 {
		t.Errorf("Image fehlt: Exit %d, %q", rc, fehl)
	}
	rc, _, _ = u.lauf(t, "MUTATIONSLAUF_IMAGE=odata-recorder:test")
	if i := u.aufrufe("create"); rc != 0 || !strings.Contains(i[0], " odata-recorder:test sh -c ") {
		t.Errorf("Image aus der Umgebung: Exit %d, %q", rc, i)
	}
	tar, _ := os.ReadFile(filepath.Join(u.zustand, "tar"))
	if !strings.Contains(string(tar), "baum/src/a.go\n") || !strings.Contains(string(tar), "mutationslauf/main.go\n") || strings.Contains(string(tar), ".git") {
		t.Errorf("Tar-Strom: %q", tar)
	}
}

// TestTreiberBaumGroesse (SPEC-210 Nr. 3 d): 256 MiB sind erlaubt, ein Byte mehr nicht.
func TestTreiberBaumGroesse(t *testing.T) {
	u := neueTreiberUmgebung(t)
	pad := filepath.Join(u.baum, "pad.bin")
	if err := os.WriteFile(pad, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("du", "-sb", "--exclude=.git", u.baum).Output()
	if err != nil {
		t.Fatal(err)
	}
	var bytes0 int64
	_, _ = fmt.Sscan(string(out), &bytes0)
	for _, c := range []struct {
		extra int64
		rc    int
	}{{0, 0}, {1, 2}} {
		if err := os.Truncate(pad, 256<<20-bytes0+c.extra); err != nil {
			t.Fatal(err)
		}
		rc, _, fehl := u.lauf(t)
		if rc != c.rc || (c.rc == 2 && !strings.Contains(fehl, "Baum zu groß")) {
			t.Errorf("+%d Byte: Exit %d, %q", c.extra, rc, fehl)
		}
	}
}
