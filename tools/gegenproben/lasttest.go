package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Fälle 217 bis 225 (SPEC-102, SPEC-216 Nr. 2, 3 und 15): tools/harness/lasttest.sh im Verzeichnis des Falls mit einem Ersatz für docker am
// Anfang von PATH. Der Ersatz hält jeden Aufruf mit allen Argumenten fest und antwortet nach den Variablen LL_*. Weder ein Daemon noch ein
// Image noch Netz; weder go noch make.
const llErsatzDocker = `#!/bin/sh
l=${LL_LOG:?}
echo AUFRUF >> "$l/docker.log"
for a in "$@"; do printf 'ARG %s\n' "$a" >> "$l/docker.log"; done
case ${1:-} in
info)
	[ -z "${LL_SIG_INFO:-}" ] || kill -"$LL_SIG_INFO" "$PPID"
	printf '%s\n' "${LL_NCPU-16}"
	exit "${LL_INFO_RC:-0}" ;;
build)
	[ -z "${LL_SIG_BUILD:-}" ] || kill -"$LL_SIG_BUILD" "$PPID"
	exit "${LL_BUILD_RC:-0}" ;;
run)
	if [ -n "${LL_SIG_RUN:-}" ]; then
		trap 'echo TERM >> "$l/empfangen"; exit 143' TERM
		exec > /dev/null 2>&1
		kill -"$LL_SIG_RUN" "$PPID"
		i=0
		while [ "$i" -lt 30 ]; do sleep 0.1; i=$((i + 1)); done
		exit 0
	fi
	echo "lasttest: Ausgabe des Containers"
	echo "stderr des Containers" >&2
	exit "${LL_RUN_RC:-0}" ;;
*) exit 99 ;;
esac
`

const (
	llImage      = "odata-recorder:lasttest"
	llProgramm   = "/out/lasttest"
	llAusgabe    = "lasttest: Ausgabe des Containers\n"
	llContainerE = "stderr des Containers\n"
)

// llVars sind die vier Variablen in der Folge von SPEC-216 Nr. 2 mit ihrer Vorgabe.
func llVars() []struct{ name, vorgabe string } {
	return []struct{ name, vorgabe string }{
		{"LASTTEST_BETRIEBSART", "beide"}, {"LASTTEST_PROFIL", "abnahme"}, {"LASTTEST_ZIEL_MS", "100"}, {"LASTTEST_STARTWERT", "1"},
	}
}

func llInfo() []string { return []string{"info", "--format", "{{.NCPU}}"} }

func llBau() []string {
	return []string{"build", "--progress=plain", "--target", "lasttest", "-t", llImage, "."}
}

// llEnvWerte sind die fünf Einträge, die docker run mit --env tragen muss: der wirksame Wert jeder Variablen und die Kerne aus docker info.
func llEnvWerte(gesetzt map[string]string, kerne string) []string {
	var out []string
	for _, v := range llVars() {
		w, ok := gesetzt[v.name]
		if !ok {
			w = v.vorgabe
		}
		out = append(out, v.name+"="+w)
	}
	return append(out, "LASTTEST_KERNE="+kerne)
}

// llPflicht sind die fünf Pflichtoptionen von docker run mit ihrem Wert (SPEC-216 Nr. 3 b); "" ist eine Option ohne Wert.
func llPflicht() map[string]string {
	return map[string]string{"--rm": "", "--network": "none", "--cpus": "2", "--memory": "4g", "--memory-swap": "4g"}
}

// llVerboten sind die 14 Optionen, die docker run nie trägt (SPEC-216 Nr. 3 c).
func llVerboten() []string {
	return []string{"-v", "--volume", "--mount", "--privileged", "-p", "--publish", "-P", "--device", "--pid", "--ipc", "--uts", "--userns",
		"--read-only", "--tmpfs"}
}

// llOptionen sind die Optionen von docker run als Liste: die Werte je Pflichtoption (--env eingeschlossen), die Befunde zu jeder anderen
// Option und die Argumente dahinter.
type llOptionen struct {
	werte   map[string][]string
	befunde []string
	rest    []string
}

// llLies zerlegt die Argumente hinter run; eine Pflichtoption mit Wert nimmt ihn aus `=` oder aus dem folgenden Argument.
func llLies(args []string) llOptionen {
	o := llOptionen{werte: map[string][]string{}}
	i := 0
	for ; i < len(args) && strings.HasPrefix(args[i], "-"); i++ {
		name, wert, gleich := strings.Cut(args[i], "=")
		soll, pflicht := llPflicht()[name]
		switch {
		case name == "--rm" && !gleich:
			o.werte[name] = append(o.werte[name], "")
		case (pflicht && soll != "") || name == "--env":
			if !gleich && i+1 < len(args) {
				i++
				wert = args[i]
			}
			o.werte[name] = append(o.werte[name], wert)
		case slices.Contains(llVerboten(), name):
			o.befunde = append(o.befunde, "verboten: "+name)
		default:
			o.befunde = append(o.befunde, "unbekannte Option: "+args[i])
		}
	}
	o.rest = args[i:]
	return o
}

// llPruefeRun liest die Argumente von docker run als Liste und nennt jeden Befund: eine fehlende, doppelte oder abweichende Pflichtoption,
// ein --env außer env, jede andere Option und alles außer dem Image und dem Programm dahinter.
func llPruefeRun(args, env []string) []string {
	if len(args) == 0 || args[0] != "run" {
		return []string{"kein docker run"}
	}
	o := llLies(args[1:])
	befunde := o.befunde
	for name, soll := range llPflicht() {
		if !slices.Equal(o.werte[name], []string{soll}) {
			befunde = append(befunde, fmt.Sprintf("Pflicht %s: %q statt %q", name, o.werte[name], soll))
		}
	}
	envs, soll := slices.Sorted(slices.Values(o.werte["--env"])), slices.Sorted(slices.Values(env))
	if !slices.Equal(envs, soll) {
		befunde = append(befunde, fmt.Sprintf("--env %q statt %q", envs, soll))
	}
	if !slices.Equal(o.rest, []string{llImage, llProgramm}) {
		befunde = append(befunde, fmt.Sprintf("hinter den Optionen %q", o.rest))
	}
	return befunde
}

// llLauf ist ein Aufruf des Treibers mit eigenem Protokoll und eigenem TMPDIR.
type llLauf struct {
	code        int
	out, errOut string
	aufrufe     [][]string
	log         string
}

// llFahre ruft den Treiber im Verzeichnis name unter dem TMPDIR des Falls mit der Umgebung des Falls und env.
func llFahre(l Lauf, name string, env []string, args ...string) (llLauf, string) {
	var r llLauf
	basis := filepath.Join(l.Tmp, name)
	r.log = filepath.Join(basis, "log")
	tmp := filepath.Join(basis, "tmp")
	for _, d := range []string{r.log, tmp} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return r, err.Error()
		}
	}
	if err := os.WriteFile(filepath.Join(r.log, "docker.log"), nil, 0o644); err != nil {
		return r, err.Error()
	}
	umgebung := append(slices.Clone(l.Env), "LL_LOG="+r.log, "TMPDIR="+tmp)
	umgebung = append(umgebung, env...)
	var abgelaufen bool
	r.code, r.out, r.errOut, abgelaufen = starteStroeme(fallFrist, l.Dir, umgebung, true, "sh", append([]string{l.U.LasttestTreiber}, args...)...)
	if abgelaufen {
		return r, name + ": Zeit überschritten"
	}
	aufrufe, err := gmProtokoll(basis, "docker.log")
	if err != nil {
		return r, err.Error()
	}
	r.aufrufe = aufrufe
	if grund := gmLeer(tmp); grund != "" {
		return r, name + ": " + grund
	}
	return r, ""
}

// llErwarte prüft Exit, die Meldung des Treibers (leer: stderr ohne Zeile des Treibers) und die Unterbefehle von docker in ihrer Folge.
func llErwarte(r llLauf, name string, code int, meldung string, kommandos ...string) string {
	var got []string
	for _, a := range r.aufrufe {
		if len(a) > 0 {
			got = append(got, a[0])
		}
	}
	var meldungen []string
	for _, z := range strings.Split(r.errOut, "\n") {
		if strings.HasPrefix(z, "lasttest: ") {
			meldungen = append(meldungen, z)
		}
	}
	var want []string
	if meldung != "" {
		want = []string{"lasttest: " + meldung}
	}
	switch {
	case r.code != code:
		return fmt.Sprintf("%s: Exit %d statt %d\n%s", name, r.code, code, r.errOut)
	case !slices.Equal(meldungen, want):
		return fmt.Sprintf("%s: Meldungen %q statt %q", name, meldungen, want)
	case !slices.Equal(got, kommandos):
		return fmt.Sprintf("%s: Aufrufe von docker %v statt %v", name, got, kommandos)
	}
	return ""
}

// llGut prüft einen Lauf, der den Container startet: Exit 0, Aufrufe info, build und run mit den Argumenten der Variablen in gesetzt.
func llGut(r llLauf, name string, gesetzt map[string]string, kerne string) string {
	if g := llErwarte(r, name, 0, "", "info", "build", "run"); g != "" {
		return g
	}
	if !slices.Equal(r.aufrufe[0], llInfo()) || !slices.Equal(r.aufrufe[1], llBau()) {
		return fmt.Sprintf("%s: info %q, build %q", name, r.aufrufe[0], r.aufrufe[1])
	}
	if befunde := llPruefeRun(r.aufrufe[2], llEnvWerte(gesetzt, kerne)); len(befunde) > 0 {
		return fmt.Sprintf("%s: %s", name, strings.Join(befunde, "; "))
	}
	return ""
}

// llForm ist eine Form einer Variablen: nil ist ungesetzt; meldung leer heißt gültig.
type llForm struct {
	name    string
	wert    *string
	meldung string
}

// llFormen sind die 31 Formen von SPEC-216 Nr. 2: je Variable ungesetzt, gültig an den Grenzen, leer, null und Form.
func llFormen() []llForm {
	var out []llForm
	hinzu := func(name string, gueltig, ungueltig []string, meldungen map[string]string) {
		out = append(out, llForm{name: name})
		for _, w := range gueltig {
			out = append(out, llForm{name: name, wert: &w})
		}
		for _, w := range ungueltig {
			m := meldungen[w]
			if m == "" {
				m = name + " Form: '" + w + "'"
			}
			if w == "" {
				m = name + " leer"
			}
			out = append(out, llForm{name: name, wert: &w, meldung: m})
		}
	}
	hinzu("LASTTEST_BETRIEBSART", []string{"replay", "mock", "beide"}, []string{"", "alle", "Replay"}, nil)
	hinzu("LASTTEST_PROFIL", []string{"abnahme", "rauch"}, []string{"", "voll"}, nil)
	null := "LASTTEST_ZIEL_MS null"
	hinzu("LASTTEST_ZIEL_MS", []string{"1", "9999"}, []string{"", "0", "00", "10000", "01", "1.5", "-1", "1 "},
		map[string]string{"0": null, "00": null})
	hinzu("LASTTEST_STARTWERT", []string{"0", "1", strings.Repeat("9", 18)}, []string{strings.Repeat("9", 19), "", "01", "-1"}, nil)
	return out
}

// llZifferFormen ist die Menge Ziffer mal Stelle der zwei Dezimalzahlen (SPEC-216 Nr. 2): für den Zielwert jedes der 15 Zeichen an jeder
// Stelle der Werte aus Einsen der Längen 1 bis 4 und die Längen 5 und 6, für den Startwert die Zeichen 0, 9 und a an jeder der 18 Stellen
// eines Werts aus 18 Einsen und die Längen 1 bis 19. Das Orakel sind die Ausdrücke des Vertrags, nicht der Code des Treibers.
func llZifferFormen() []llForm {
	ziel := regexp.MustCompile(`^[1-9][0-9]{0,3}$`)
	start := regexp.MustCompile(`^(0|[1-9][0-9]{0,17})$`)
	var out []llForm
	hinzu := func(name, w string, ok *regexp.Regexp) {
		f := llForm{name: name, wert: &w}
		switch {
		case w == "":
			f.meldung = name + " leer"
		case name == "LASTTEST_ZIEL_MS" && strings.Trim(w, "0") == "":
			f.meldung = name + " null"
		case !ok.MatchString(w):
			f.meldung = name + " Form: '" + w + "'"
		}
		out = append(out, f)
	}
	for laenge := 1; laenge <= 4; laenge++ {
		einsen := strings.Repeat("1", laenge)
		for stelle := range laenge {
			for _, z := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "-", "+", " ", "."} {
				hinzu("LASTTEST_ZIEL_MS", einsen[:stelle]+z+einsen[stelle+1:], ziel)
			}
		}
	}
	for _, n := range []int{5, 6} {
		hinzu("LASTTEST_ZIEL_MS", strings.Repeat("9", n), ziel)
	}
	basis := strings.Repeat("1", 18)
	for stelle := range 18 {
		for _, z := range []string{"0", "9", "a"} {
			hinzu("LASTTEST_STARTWERT", basis[:stelle]+z+basis[stelle+1:], start)
		}
	}
	for n := 1; n <= 19; n++ {
		hinzu("LASTTEST_STARTWERT", strings.Repeat("7", n), start)
	}
	return out
}

// llFormLaeufe fährt jede Form: eine gültige startet den Container mit dem wirksamen Wert, eine ungültige endet mit 2, ihrer Meldung und
// ohne einen Aufruf von docker.
func llFormLaeufe(l Lauf, formen []llForm) string {
	for i, f := range formen {
		var env []string
		gesetzt := map[string]string{}
		if f.wert != nil {
			env = []string{f.name + "=" + *f.wert}
			gesetzt[f.name] = *f.wert
		}
		name := fmt.Sprintf("form-%d", i)
		r, grund := llFahre(l, name, env)
		if grund != "" {
			return grund
		}
		was := name + " " + f.name
		if f.wert != nil {
			was += "=" + strconv.Quote(*f.wert)
		}
		if f.meldung == "" {
			grund = llGut(r, was, gesetzt, "16")
		} else {
			grund = llErwarte(r, was, 2, f.meldung)
		}
		if grund != "" {
			return grund
		}
	}
	return ""
}

// llFall baut einen Fall mit dem Treiber; der erste Aufruf läuft mit env und endet mit exit.
func llFall(name string, env []string, exit int, nachher func(Lauf) string, want ...string) Fall {
	f := gmFall(name, gateLasttest, map[string]string{"bin/docker": llErsatzDocker, "log/docker.log": ""}, nil, exit, want...)
	f.Name = "lasttest/" + name
	f.Ausfuehrbar = []string{"bin/docker"}
	f.Env = append([]string{"PATH={dir}/bin:{PATH}", "LL_LOG={dir}/log"}, env...)
	f.Nachher = nachher
	return f
}

// llVariablen sind die Fälle 217 bis 219.
func llVariablen() []Fall {
	formen := llFall("variablen", nil, 0, func(l Lauf) string {
		alle := llFormen()
		if len(alle) != 31 {
			return fmt.Sprintf("%d Formen statt 31", len(alle))
		}
		return llFormLaeufe(l, alle)
	}, "1>"+strings.TrimSuffix(llAusgabe, "\n"))
	ziffern := llFall("variablen-ziffern", nil, 0, func(l Lauf) string {
		alle := llZifferFormen()
		if len(alle) != (1+2+3+4)*15+2+18*3+19 {
			return fmt.Sprintf("%d Formen statt 225", len(alle))
		}
		return llFormLaeufe(l, alle)
	})
	reihenfolge := llFall("variablen-erste-meldung", nil, 0, func(l Lauf) string {
		n := 0
		vars := llVars()
		for i := range vars {
			for j := i + 1; j < len(vars); j++ {
				name := fmt.Sprintf("paar-%d-%d", i, j)
				r, grund := llFahre(l, name, []string{vars[j].name + "=x", vars[i].name + "=y"})
				if grund == "" {
					grund = llErwarte(r, name, 2, vars[i].name+" Form: 'y'")
				}
				if grund != "" {
					return grund
				}
				n++
			}
		}
		if n != 6 {
			return fmt.Sprintf("%d Paare statt 6", n)
		}
		return ""
	})
	return []Fall{formen, ziffern, reihenfolge}
}

// llDockerRun ist Fall 220: die Argumente von docker run mit den Vorgaben und mit anderen Werten, LASTTEST_KERNE aus docker info statt aus der
// Umgebung, auch ohne Zahl unverändert, die Ströme des Containers unverändert und nur info, build und run.
func llDockerRun() Fall {
	f := llFall("docker-run", []string{"LASTTEST_KERNE=99"}, 0, func(l Lauf) string {
		basis, err := gmProtokoll(l.Dir, "docker.log")
		if err != nil {
			return err.Error()
		}
		if len(basis) != 3 {
			return fmt.Sprintf("%d Aufrufe statt 3: %q", len(basis), basis)
		}
		if b := llPruefeRun(basis[2], llEnvWerte(nil, "16")); len(b) > 0 {
			return strings.Join(b, "; ")
		}
		andere := map[string]string{
			"LASTTEST_BETRIEBSART": "mock", "LASTTEST_PROFIL": "rauch", "LASTTEST_ZIEL_MS": "80", "LASTTEST_STARTWERT": "42",
		}
		env := []string{"LASTTEST_KERNE=3", "LL_NCPU=20"}
		for k, v := range andere {
			env = append(env, k+"="+v)
		}
		r, grund := llFahre(l, "andere", env)
		if grund == "" {
			grund = llGut(r, "andere", andere, "20")
		}
		if grund == "" && (r.out != llAusgabe || r.errOut != llContainerE) {
			grund = fmt.Sprintf("Ströme: stdout %q, stderr %q", r.out, r.errOut)
		}
		return gmErste(grund, llOhneZahl(l))
	}, "1>"+strings.TrimSuffix(llAusgabe, "\n"), strings.TrimSuffix(llContainerE, "\n"))
	return f
}

// llOhneZahl fährt eine Ausgabe von docker info ohne Zahl, auch eine leere: sie geht unverändert in --env (SPEC-216 Nr. 3 a).
func llOhneZahl(l Lauf) string {
	for i, ncpu := range []string{"", "viele"} {
		name := fmt.Sprintf("ncpu-%d", i)
		r, grund := llFahre(l, name, []string{"LL_NCPU=" + ncpu})
		if grund == "" {
			grund = llGut(r, name, nil, ncpu)
		}
		if grund != "" {
			return grund
		}
	}
	return ""
}

// llReihenfolge sind die Fälle 221 bis 223: Argument, docker info und Bau.
func llReihenfolge() []Fall {
	aufruf := "Aufruf: tools/harness/lasttest.sh (ohne Argument)"
	argument := llFall("argument", nil, 0, func(l Lauf) string {
		for i, a := range []string{"", "x"} {
			name := fmt.Sprintf("argument-%d", i)
			r, grund := llFahre(l, name, nil, a)
			if grund == "" {
				grund = llErwarte(r, name, 2, aufruf)
			}
			if grund != "" {
				return grund
			}
		}
		return ""
	})
	einzeln := func(name, env, meldung string, kommandos ...string) Fall {
		f := llFall(name, []string{env}, 2, func(l Lauf) string {
			r, grund := llFahre(l, name, []string{env})
			if grund == "" {
				grund = llErwarte(r, name, 2, meldung, kommandos...)
			}
			return grund
		}, "lasttest: "+meldung)
		return f
	}
	info := einzeln("docker-info-scheitert", "LL_INFO_RC=1", "docker nicht erreichbar", "info")
	bau := einzeln("bau-scheitert", "LL_BUILD_RC=1", "Bau gescheitert", "info", "build")
	return []Fall{argument, info, bau}
}

// llExit ist die Erwartung zu einem Exit von docker run.
type llExit struct {
	run, exit int
	meldung   string
}

// llExits sind die neun Exits des Containers aus SPEC-216 Nr. 3 d und drei Grenzen der Bereiche (4, 124, 128).
func llExits() []llExit {
	lauf := "docker run gescheitert"
	endete := func(c int) llExit { return llExit{c, 3, "Container endete mit Exit " + strconv.Itoa(c)} }
	return []llExit{
		{0, 0, ""}, {1, 1, ""}, {2, 2, ""}, {3, 3, ""}, {125, 2, lauf}, {126, 2, lauf}, {127, 2, lauf}, endete(137), endete(255),
		endete(4), endete(124), endete(128),
	}
}

// llExitFall ist Fall 224.
func llExitFall() Fall {
	return llFall("exit-des-containers", nil, 0, func(l Lauf) string {
		n := 0
		for _, e := range llExits() {
			name := fmt.Sprintf("exit-%d", e.run)
			r, grund := llFahre(l, name, []string{"LL_RUN_RC=" + strconv.Itoa(e.run)})
			if grund == "" {
				grund = llErwarte(r, name, e.exit, e.meldung, "info", "build", "run")
			}
			if grund == "" && r.out != llAusgabe {
				grund = fmt.Sprintf("%s: stdout %q", name, r.out)
			}
			if grund != "" {
				return grund
			}
			n++
		}
		if n != 12 {
			return fmt.Sprintf("%d Exits statt 12", n)
		}
		return ""
	})
}

// llSignale sind die neun Paare aus Signal und Ort (SPEC-216 Nr. 3 e).
func llSignale() [][2]string {
	var out [][2]string
	for _, s := range []string{"INT", "TERM", "HUP"} {
		for _, ort := range []string{"INFO", "BUILD", "RUN"} {
			out = append(out, [2]string{s, ort})
		}
	}
	return out
}

// llAbbruch ist Fall 225: INT, TERM und HUP während docker info, build und run enden mit 2 und `abgebrochen`, ohne einen späteren Aufruf;
// während docker run reicht der Treiber das Signal als TERM an den laufenden Aufruf weiter und wartet auf dessen Rückkehr.
func llAbbruch() Fall {
	return llFall("abbruch", nil, 0, func(l Lauf) string {
		n := 0
		for _, p := range llSignale() {
			name := "signal-" + p[0] + "-" + p[1]
			r, grund := llFahre(l, name, []string{"LL_SIG_" + p[1] + "=" + p[0]})
			kommandos := map[string][]string{"INFO": {"info"}, "BUILD": {"info", "build"}, "RUN": {"info", "build", "run"}}[p[1]]
			if grund == "" {
				grund = llErwarte(r, name, 2, "abgebrochen", kommandos...)
			}
			if grund == "" && p[1] == "RUN" {
				b, err := os.ReadFile(filepath.Join(r.log, "empfangen"))
				if err != nil || string(b) != "TERM\n" {
					grund = fmt.Sprintf("%s: der Aufruf von docker empfing %q, %v", name, b, err)
				}
			}
			if grund != "" {
				return grund
			}
			n++
		}
		if n != 9 {
			return fmt.Sprintf("%d Paare statt 9", n)
		}
		return ""
	})
}

// lasttestFaelle sind die Fälle 217 bis 225.
func lasttestFaelle() []Fall {
	alle := append(llVariablen(), llDockerRun())
	alle = append(alle, llReihenfolge()...)
	return append(alle, llExitFall(), llAbbruch())
}
