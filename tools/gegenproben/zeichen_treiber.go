package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Fälle 151 bis 155 (SPEC-102, SPEC-233 Nr. 2 bis 4): tools/harness/zeichen.sh in einem Git-Repo des
// Falls, mit einem Ersatz-Skript für docker am Anfang von PATH, dem echten git und dem echten tar.
// Ersatz, Werkzeuge und Protokolle liegen im TMPDIR des Falls, außerhalb des Repos.

// zcErsatzDocker hält jeden Aufruf mit allen Argumenten in $ZC_LOG/docker.log fest und antwortet nach
// ZC_CREATE_RC, ZC_CP_RC, ZC_START_RC und ZC_SIGNAL; `cp -` legt den Strom nach $ZC_LOG/strom.tar.
const zcErsatzDocker = `#!/bin/sh
l=${ZC_LOG:?}
echo "$*" >> "$l/docker.log"
case $1 in
create)
	[ "${ZC_CREATE_RC:-0}" = 0 ] || exit "$ZC_CREATE_RC"
	echo zc-kennung ;;
cp)
	cat > "$l/strom.tar"
	exit "${ZC_CP_RC:-0}" ;;
start)
	echo start-stdout
	echo start-stderr >&2
	[ -z "${ZC_SIGNAL:-}" ] || kill -"$ZC_SIGNAL" "$PPID"
	exit "${ZC_START_RC:-0}" ;;
rm) ;;
*) exit 99 ;;
esac
`

// zcWerkzeuge legt im TMPDIR des Falls (dir + ".tmp") den Ersatz für docker und drei Verzeichnisse mit
// Verweisen auf die echten Programme an: werkzeug (git, tar, cat), ohne-git und ohne-tar.
func zcWerkzeuge(dir string) error {
	tmp := dir + ".tmp"
	for _, d := range []string{"bin", "werkzeug", "ohne-git", "ohne-tar", "log"} {
		if err := os.MkdirAll(filepath.Join(tmp, d), 0o755); err != nil {
			return err
		}
	}
	for _, n := range []string{"docker", "docker-ersatz"} {
		if err := os.WriteFile(filepath.Join(tmp, "bin", n), []byte(zcErsatzDocker), 0o755); err != nil {
			return err
		}
	}
	for ordner, namen := range map[string][]string{"werkzeug": {"git", "tar", "cat"}, "ohne-git": {"tar", "cat"}, "ohne-tar": {"git", "cat"}} {
		for _, n := range namen {
			p, err := exec.LookPath(n)
			if err != nil {
				return fmt.Errorf("%s nicht aufrufbar", n)
			}
			if err := os.Symlink(p, filepath.Join(tmp, ordner, n)); err != nil {
				return err
			}
		}
	}
	return nil
}

// zcTreiberFall baut einen Fall mit dem Treiber: das Repo legt vor an, Exit ist der des Falls.
func zcTreiberFall(name string, vor func(dir string) error, exit int, want ...string) Fall {
	f := gmFall(name, gateZeichenTreiber, nil, nil, exit, want...)
	f.Name = "zeichen/" + name
	f.Env = []string{"PATH={dir}.tmp/bin:{dir}.tmp/werkzeug", "ZC_LOG={dir}.tmp/log", "GIT_CONFIG_GLOBAL=/dev/null"}
	f.Vorbereitung = func(dir string) error {
		if err := zcWerkzeuge(dir); err != nil {
			return err
		}
		if vor == nil {
			return nil
		}
		return vor(dir)
	}
	return f
}

// zcRepoMit legt in dir ein Repo mit den Dateien an (ein Commit).
func zcRepoMit(dir string, dateien map[string]string) error {
	r := &gmRepo{dir: dir}
	r.git("init", "-q")
	r.commit(dateien)
	return r.err
}

// zcLauf ist ein Aufruf des Treibers mit eigenem Protokollverzeichnis.
type zcLauf struct {
	dir  string
	env  []string
	args []string
}

// treiber ruft zeichen.sh wie der Fall und liefert Exit und beide Ströme; das Protokoll liegt in
// <Tmp>/log-<name>.
func (l Lauf) zeichenTreiber(name string, z zcLauf) (code int, out, errOut, log string) {
	log = filepath.Join(l.Tmp, "log-"+name)
	if err := os.MkdirAll(log, 0o755); err != nil {
		return -1, "", err.Error(), log
	}
	dir := z.dir
	if dir == "" {
		dir = l.Dir
	}
	env := append(slices.Clone(l.Env), "ZC_LOG="+log)
	env = append(env, z.env...)
	code, out, errOut, _ = starteStroeme(fallFrist, dir, env, true, "bash", append([]string{l.U.ZeichenTreiber}, z.args...)...)
	return code, out, errOut, log
}

// zcAufrufe liest das Protokoll des Ersatzes in log: eine Zeile je Aufruf; ein fehlendes Protokoll
// ist kein Aufruf.
func zcAufrufe(log string) []string {
	b, err := os.ReadFile(filepath.Join(log, "docker.log"))
	if err != nil || len(b) == 0 {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

// zcZaehle zählt die Aufrufe, die mit dem Kommando beginnen.
func zcZaehle(aufrufe []string, kommando string) int {
	n := 0
	for _, a := range aufrufe {
		if strings.HasPrefix(a, kommando+" ") || a == kommando {
			n++
		}
	}
	return n
}

// zcStrom liest den Tar-Strom des Ersatzes: Name und Typ jedes Eintrags.
func zcStrom(log string) (map[string]byte, error) {
	b, err := os.ReadFile(filepath.Join(log, "strom.tar"))
	if err != nil {
		return nil, fmt.Errorf("Tar-Strom fehlt: %w", err)
	}
	r := tar.NewReader(bytes.NewReader(b))
	e := map[string]byte{}
	for {
		h, err := r.Next()
		if err == io.EOF {
			return e, nil
		}
		if err != nil {
			return nil, err
		}
		e[h.Name] = h.Typeflag
	}
}

// zcMenge ist Fall 151: zehn Arten Pfad (getrackt, ungetrackt, ignoriert, im Arbeitsbaum gelöscht, Name
// mit Leerzeichen, Name mit Zeilenumbruch, Symlink samt hängendem, mitgelieferte Baseline, Gitlink mit
// Verzeichnis, ungetracktes verschachteltes Repo); die beiden letzten stehen als Verzeichnis ohne Inhalt im Strom.
func zcMenge() Fall {
	f := zcTreiberFall("treiber-menge", zcMengeRepo, 0)
	f.Nachher = func(l Lauf) string {
		strom, err := zcStrom(filepath.Join(l.Tmp, "log"))
		if err != nil {
			return err.Error()
		}
		soll := []string{".gitignore", ".harness/baseline/v/x.md", "Makefile", "a b.md", "docs/a.md", "docs/neu.md", "gl/", "h", "l", "n\nm.md", "src/a.go", "sub/"}
		var got []string
		for n := range strom {
			got = append(got, n)
		}
		slices.Sort(got)
		if !slices.Equal(got, soll) {
			return fmt.Sprintf("Tar-Strom nennt %q statt %q", got, soll)
		}
		for _, n := range []string{"h", "l"} {
			if strom[n] != tar.TypeSymlink {
				return fmt.Sprintf("%s steht nicht als Symlink im Strom (Typ %q)", n, strom[n])
			}
		}
		for _, n := range []string{"gl/", "sub/"} {
			if strom[n] != tar.TypeDir {
				return fmt.Sprintf("%s steht nicht als Verzeichnis im Strom (Typ %q)", n, strom[n])
			}
		}
		return ""
	}
	return f
}

// zcMengeRepo legt das Repo von Fall 151 an: getrackte Dateien und Symlinks, danach eine ungetrackte
// Datei und die Löschung einer getrackten; `ign/` ist ignoriert. Dazu ein Gitlink `gl` (Index-Eintrag 160000
// mit Verzeichnis und `x.md`) und ein verschachteltes Repo `sub/`, das der Fall nicht mit `git add` aufnimmt
// (`add -A` nähme es als Gitlink auf).
func zcMengeRepo(dir string) error {
	r := &gmRepo{dir: dir}
	r.git("init", "-q")
	err := gmSchreibe(dir, map[string]string{
		"src/a.go": "x\n", "docs/a.md": "x\n", "Makefile": "x\n", "a b.md": "x\n", "n\nm.md": "x\n",
		".gitignore": "ign/\n", ".harness/baseline/v/x.md": "x\n", "ign/x.md": "x\n", "docs/weg.md": "x\n",
	})
	if err != nil {
		return err
	}
	for name, ziel := range map[string]string{"l": "src/a.go", "h": "gibt-es-nicht"} {
		if err := os.Symlink(ziel, filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	r.abgelegt()
	if r.err != nil {
		return r.err
	}
	if err := gmSchreibe(dir, map[string]string{"docs/neu.md": "x\n", "gl/x.md": "x\n", "sub/x.md": "x\n"}); err != nil {
		return err
	}
	r.git("update-index", "--add", "--cacheinfo", "160000,"+strings.Repeat("1", 40)+",gl")
	(&gmRepo{dir: filepath.Join(dir, "sub")}).git("init", "-q")
	if r.err != nil {
		return r.err
	}
	return os.Remove(filepath.Join(dir, "docs", "weg.md"))
}

// zcPruefeAufrufe prüft das Protokoll eines Laufs: genau die vier Aufrufe in dieser Folge mit dem Image
// image, und nirgends run, -v, --volume oder --mount.
func zcPruefeAufrufe(got []string, image string) string {
	soll := []string{"create --network none " + image + " pruefen /work", "cp - zc-kennung:/work", "start -a zc-kennung", "rm -f zc-kennung"}
	if !slices.Equal(got, soll) {
		return fmt.Sprintf("Protokoll %q statt %q", got, soll)
	}
	for _, a := range got {
		for _, w := range strings.Fields(a) {
			if w == "-v" || w == "--volume" || w == "--mount" || w == "run" {
				return fmt.Sprintf("Aufruf %q nennt %s", a, w)
			}
		}
	}
	return ""
}

// zcAufrufeFall ist Fall 152: genau vier Aufrufe von docker in dieser Reihenfolge, nie run, nie
// ein Mount; ZEICHEN_IMAGE bestimmt das Image.
func zcAufrufeFall() Fall {
	f := zcTreiberFall("treiber-aufrufe-von-docker", func(dir string) error {
		return zcRepoMit(dir, map[string]string{"a.md": "a\n"})
	}, 0)
	f.Nachher = func(l Lauf) string {
		if r := zcPruefeAufrufe(zcAufrufe(filepath.Join(l.Tmp, "log")), "odata-recorder:zeichen"); r != "" {
			return r
		}
		// ZEICHEN_IMAGE nennt das Image; DOCKER nennt das Programm, PATH führt keines mit dem Namen docker.
		ersatz := filepath.Join(l.Tmp, "bin", "docker-ersatz")
		for _, e := range []struct {
			name, image string
			z           zcLauf
		}{
			{"image", "x:y", zcLauf{env: []string{"ZEICHEN_IMAGE=x:y"}}},
			{"programm", "odata-recorder:zeichen", zcLauf{env: []string{"DOCKER=" + ersatz, "PATH=" + l.Tmp + "/werkzeug"}}},
		} {
			code, _, errOut, log := l.zeichenTreiber(e.name, e.z)
			if code != 0 {
				return fmt.Sprintf("Lauf %s: Exit %d\n%s", e.name, code, errOut)
			}
			if r := zcPruefeAufrufe(zcAufrufe(log), e.image); r != "" {
				return "Lauf " + e.name + ": " + r
			}
		}
		return ""
	}
	return f
}

// zcExit ist Fall 153: der Exit von start geht für 0 und 1 durch, jeder andere wird 2; beide Texte kommen
// unverändert durch. Der Lauf des Falls selbst ist der Wert 0.
func zcExit() Fall {
	f := zcTreiberFall("treiber-exit", func(dir string) error {
		return zcRepoMit(dir, map[string]string{"a.md": "a\n"})
	}, 0, "1>start-stdout\n", "start-stderr\n")
	f.Nachher = func(l Lauf) string {
		werte := []int{1, 2, 3, 125, 126, 127, 130, 137, 255}
		if len(werte) != 9 {
			return "Exit-Werte: nicht neun neben dem Lauf des Falls"
		}
		for _, rc := range werte {
			soll := 2
			if rc == 1 {
				soll = 1
			}
			code, out, errOut, _ := l.zeichenTreiber(fmt.Sprintf("rc%d", rc), zcLauf{env: []string{fmt.Sprintf("ZC_START_RC=%d", rc)}})
			if code != soll || out != "start-stdout\n" || errOut != "start-stderr\n" {
				return fmt.Sprintf("start endet mit %d: Exit %d statt %d, stdout %q, stderr %q", rc, code, soll, out, errOut)
			}
		}
		return ""
	}
	return f
}

// zcAufraeumen ist Fall 154: acht Läufe, der erste ist der des Falls.
func zcAufraeumen() Fall {
	f := zcTreiberFall("treiber-aufraeumen", func(dir string) error {
		return zcRepoMit(dir, map[string]string{"a.md": "a\n"})
	}, 0)
	f.Nachher = func(l Lauf) string {
		if r := zcPruefeLauf(zcAufrufe(filepath.Join(l.Tmp, "log")), 1, 1); r != "" {
			return "Lauf 1: " + r
		}
		laeufe := []struct {
			name         string
			env          []string
			exit, rm, st int
			text         string
		}{
			{"start1", []string{"ZC_START_RC=1"}, 1, 1, 1, ""},
			{"create", []string{"ZC_CREATE_RC=1"}, 2, 0, 0, "zeichen.sh: docker create scheitert"},
			{"cp", []string{"ZC_CP_RC=1"}, 2, 1, 0, "zeichen.sh: \xc3\x9cbergabe scheitert"},
			{"ohne-tar", []string{"PATH={tmp}/bin:{tmp}/ohne-tar"}, 2, 1, 0, "zeichen.sh: \xc3\x9cbergabe scheitert"},
			{"hup", []string{"ZC_SIGNAL=HUP"}, 2, 1, 1, ""},
			{"int", []string{"ZC_SIGNAL=INT"}, 2, 1, 1, ""},
			{"term", []string{"ZC_SIGNAL=TERM"}, 2, 1, 1, ""},
		}
		if len(laeufe) != 7 {
			return "Läufe: nicht sieben neben dem Lauf des Falls"
		}
		for _, e := range laeufe {
			env := slices.Clone(e.env)
			for i := range env {
				env[i] = strings.ReplaceAll(env[i], "{tmp}", l.Tmp)
			}
			code, _, errOut, log := l.zeichenTreiber(e.name, zcLauf{env: env})
			if code != e.exit {
				return fmt.Sprintf("Lauf %s: Exit %d statt %d\n%s", e.name, code, e.exit, errOut)
			}
			if e.text != "" && !strings.Contains(errOut, e.text) {
				return fmt.Sprintf("Lauf %s: stderr nennt %q nicht: %q", e.name, e.text, errOut)
			}
			if r := zcPruefeLauf(zcAufrufe(log), e.rm, e.st); r != "" {
				return "Lauf " + e.name + ": " + r
			}
		}
		return ""
	}
	return f
}

// zcPruefeLauf prüft die Zahl der Aufrufe rm -f und start in einem Protokoll.
func zcPruefeLauf(aufrufe []string, rm, start int) string {
	if got := zcZaehle(aufrufe, "rm"); got != rm {
		return fmt.Sprintf("%d Aufrufe von rm statt %d (%q)", got, rm, aufrufe)
	}
	if got := zcZaehle(aufrufe, "start"); got != start {
		return fmt.Sprintf("%d Aufrufe von start statt %d (%q)", got, start, aufrufe)
	}
	return ""
}

// zcZweiRepos legt unter tmp ein Repo ohne Datei (leer) und eines mit einer Datei (voll) an.
func zcZweiRepos(tmp string) (leer, voll string, err error) {
	leer, voll = filepath.Join(tmp, "leer"), filepath.Join(tmp, "voll")
	if err := os.MkdirAll(leer, 0o755); err != nil {
		return "", "", err
	}
	r := &gmRepo{dir: leer}
	r.git("init", "-q")
	if r.err != nil {
		return "", "", r.err
	}
	if err := os.MkdirAll(voll, 0o755); err != nil {
		return "", "", err
	}
	return leer, voll, zcRepoMit(voll, map[string]string{"a.md": "a\n"})
}

// zcKaputtesRepo legt unter tmp ein Repo mit einem Commit an, dessen Index das eine Byte x ist: `git rev-parse
// --show-toplevel` gelingt, `git ls-files` endet mit 128.
func zcKaputtesRepo(tmp string) (string, error) {
	kaputt := filepath.Join(tmp, "kaputt")
	if err := os.MkdirAll(kaputt, 0o755); err != nil {
		return "", err
	}
	if err := zcRepoMit(kaputt, map[string]string{"a.md": "a\n"}); err != nil {
		return "", err
	}
	return kaputt, os.WriteFile(filepath.Join(kaputt, ".git", "index"), []byte("x"), 0o644)
}

// zcOhneRepo ist Fall 155: kein Repo (der Lauf des Falls), ein Repo ohne Datei, PATH ohne git, ein Argument,
// ein leeres Argument und ein Repo, dessen Index das eine Byte x ist (git ls-files scheitert); in keinem Lauf
// ruft der Treiber docker.
func zcOhneRepo() Fall {
	f := zcTreiberFall("treiber-ohne-repo-und-ohne-dateien", nil, 2, "zeichen.sh: kein Git-Arbeitsbaum")
	f.Nachher = func(l Lauf) string {
		leer, voll, err := zcZweiRepos(l.Tmp)
		if err != nil {
			return err.Error()
		}
		kaputt, err := zcKaputtesRepo(l.Tmp)
		if err != nil {
			return err.Error()
		}
		laeufe := []struct {
			name string
			z    zcLauf
			text string
		}{
			{"leer", zcLauf{dir: leer}, "zeichen.sh: Pr\xc3\xbcfmenge leer"},
			{"ohne-git", zcLauf{dir: voll, env: []string{"PATH=" + l.Tmp + "/bin:" + l.Tmp + "/ohne-git"}}, "zeichen.sh: git nicht aufrufbar"},
			{"argument", zcLauf{dir: voll, args: []string{"x"}}, "Aufruf: zeichen.sh"},
			{"argument-leer", zcLauf{dir: voll, args: []string{""}}, "Aufruf: zeichen.sh"},
			{"index-kaputt", zcLauf{dir: kaputt}, "zeichen.sh: git ls-files scheitert"},
		}
		if len(laeufe) != 5 {
			return "Läufe: nicht fünf neben dem Lauf des Falls"
		}
		for _, e := range laeufe {
			code, _, errOut, log := l.zeichenTreiber(e.name, e.z)
			if code != 2 || !strings.Contains(errOut, e.text) {
				return fmt.Sprintf("Lauf %s: Exit %d, stderr %q statt %q", e.name, code, errOut, e.text)
			}
			if a := zcAufrufe(log); len(a) != 0 {
				return fmt.Sprintf("Lauf %s: Aufrufe von docker %q", e.name, a)
			}
		}
		if a := zcAufrufe(filepath.Join(l.Tmp, "log")); len(a) != 0 {
			return fmt.Sprintf("Lauf ohne Repo: Aufrufe von docker %q", a)
		}
		return ""
	}
	return f
}

// zeichenTreiberFaelle sind die Fälle 151 bis 155.
func zeichenTreiberFaelle() []Fall {
	return []Fall{zcMenge(), zcAufrufeFall(), zcExit(), zcAufraeumen(), zcOhneRepo()}
}

// zeichenFaelle sind die Fälle 131 bis 155.
func zeichenFaelle() []Fall {
	return append(zeichenProgrammFaelle(), zeichenTreiberFaelle()...)
}
