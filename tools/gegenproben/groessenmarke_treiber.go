package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Fälle 110 bis 120 (SPEC-102): Plan-Suche, Rangfolge der Exit-Codes, Aufräumen und die Fälle mit
// dem echten git.

// gmSchreibe legt die Dateien unter dir an.
func gmSchreibe(dir string, dateien map[string]string) error {
	for pfad, inhalt := range dateien {
		p := filepath.Join(dir, pfad)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(inhalt), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// gmLeer ist der Grund, aus dem dir nicht leer ist, oder "".
func gmLeer(dir string) string {
	es, err := os.ReadDir(dir)
	if err != nil {
		return err.Error()
	}
	if len(es) > 0 {
		return fmt.Sprintf("%s ist nicht leer: %d Einträge, etwa %s", dir, len(es), es[0].Name())
	}
	return ""
}

func gmProgramm(l Lauf) string {
	return filepath.Join(l.Dir, "bin", "programm")
}

// gmZaehle zählt die Aufrufe des Programms mit dem Kommando und dem Plan.
func gmZaehle(aufrufe [][]string, kommando, plan string) int {
	n := 0
	for _, a := range aufrufe {
		if len(a) >= 2 && a[0] == kommando && a[1] == plan {
			n++
		}
	}
	return n
}

// gmSzenario legt Pläne name/slice-a.md, … an, deren Ersatz für `pruefen` die Exit-Codes rcs liefert,
// ruft den Treiber und prüft Exit-Code und dass jeder Plan genau einmal abgefragt und geprüft wurde.
func gmSzenario(l Lauf, name string, rcs []string, want int) string {
	dateien := map[string]string{}
	pfade := make([]string, len(rcs))
	for i, rc := range rcs {
		pfade[i] = fmt.Sprintf("%s/slice-%c.md", name, 'a'+i)
		dateien[pfade[i]] = "x\n"
		dateien["ersatz/"+pfade[i]+".abfrage"] = gmMessen
		dateien["ersatz/"+pfade[i]+".pruefen.rc"] = rc + "\n"
	}
	if err := gmSchreibe(l.Dir, dateien); err != nil {
		return err.Error()
	}
	code, out := l.treiber(gmProgramm(l), name)
	if code != want {
		return fmt.Sprintf("Szenario %s: Exit %d statt %d\n%s", name, code, want, out)
	}
	aufrufe, err := gmProtokoll(l.Dir, "programm.log")
	if err != nil {
		return err.Error()
	}
	for _, p := range pfade {
		if a, g := gmZaehle(aufrufe, "abfrage", p), gmZaehle(aufrufe, "pruefen", p); a != 1 || g != 1 {
			return fmt.Sprintf("Szenario %s: %s %d-mal abgefragt und %d-mal geprüft", name, p, a, g)
		}
	}
	return ""
}

// gmTreiberPlanFaelle sind die Fälle 110 bis 114.
func gmTreiberPlanFaelle() []Fall {
	keine := "keine-messung\n"
	ohneGit := func(l Lauf, programm ...[]string) string {
		return gmErste(gmErwartet(l, "git.log", nil), gmErwartet(l, "programm.log", programm))
	}
	keinePlaene := gmTreiber("treiber-keine-plaene", gmErsatz(gmNurAndere()), 0, gmKeinePlaene, "1>"+gmKeinSlice)
	keinePlaene.Argumente = []string{"{dir}/bin/programm", gmOhnePlan()[0]}
	keinePlaene.Vorbereitung = func(dir string) error { return os.MkdirAll(filepath.Join(dir, gmOhnePlan()[0]), 0o755) }
	suche := gmTreiber("treiber-plan-suche", gmErsatz(map[string]string{
		"plaene/roadmap.md": "x\n", "plaene/README.md": "x\n", "plaene/slice-c.txt": "x\n", "plaene/unter/slice-d.md": "x\n",
	}, gmEintrag{pfad: "plaene/slice-b.md", abfrage: keine}, gmEintrag{pfad: "plaene/slice-a.md", abfrage: keine}), 0, func(l Lauf) string {
		return ohneGit(l, []string{"abfrage", gmPlanA}, []string{"abfrage", "plaene/slice-b.md"})
	})
	aggregation := gmTreiber("treiber-exit-aggregation", gmErsatz(nil,
		gmEintrag{pfad: gmPlanA, abfrage: gmMessen, pruefenRc: "2"},
		gmEintrag{pfad: "plaene/slice-b.md", abfrage: gmMessen, pruefenRc: "1"},
		gmEintrag{pfad: "plaene/slice-c.md", abfrage: gmMessen, pruefenRc: "0"}), 2, func(l Lauf) string {
		reihenfolge := gmErwartet(l, "programm.log", [][]string{
			{"abfrage", gmPlanA}, {"pruefen", gmPlanA, "*messung"}, {"abfrage", "plaene/slice-b.md"},
			{"pruefen", "plaene/slice-b.md", "*messung"}, {"abfrage", "plaene/slice-c.md"}, {"pruefen", "plaene/slice-c.md", "*messung"},
		})
		return gmErste(reihenfolge, gmSzenario(l, "p1", []string{"0", "1"}, 1), gmSzenario(l, "p2", []string{"1", "0"}, 1),
			gmSzenario(l, "p3", []string{"0", "0"}, 0), gmExitSchleife(l))
	})
	hinweis := func(plan, text string) string {
		return "groessenmarke " + plan + ": Hinweis: " + text + "\n"
	}
	ohneBlock := gmTreiber("treiber-ohne-block-kein-git", gmErsatz(nil,
		gmEintrag{pfad: gmPlanA, abfrage: keine, hinweis: hinweis(gmPlanA, "kein Block")},
		gmEintrag{pfad: "plaene/slice-b.md", abfrage: keine, hinweis: hinweis("plaene/slice-b.md", "Prosa-Form")},
		gmEintrag{pfad: "plaene/slice-c.md", abfrage: keine, hinweis: hinweis("plaene/slice-c.md", "Marke steht aus: Sch\xc3\xa4tzung 800, Marke 1100")}),
		0, func(l Lauf) string {
			return ohneGit(l, []string{"abfrage", gmPlanA}, []string{"abfrage", "plaene/slice-b.md"}, []string{"abfrage", "plaene/slice-c.md"})
		}, "Hinweis: kein Block", "Hinweis: Prosa-Form", "Hinweis: Marke steht aus")
	aufraeumen := gmTreiber("treiber-aufraeumen", gmErsatz(nil, gmEintrag{pfad: gmPlanA, abfrage: gmMessen}), 0, gmAufgeraeumt)
	return []Fall{keinePlaene, suche, aggregation, ohneBlock, aufraeumen}
}

// gmRepo ist ein Git-Repo, das ein Fall der Gegenproben anlegt; Konfiguration allein über -c und
// GIT_CONFIG_GLOBAL=/dev/null. Der erste Fehler bleibt in err, danach tut jede Operation nichts.
type gmRepo struct {
	dir string
	err error
}

func (r *gmRepo) git(args ...string) string {
	if r.err != nil {
		return ""
	}
	a := append([]string{"-c", "user.name=gegenprobe", "-c", "user.email=gegenprobe@example.invalid",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)
	code, out, abgelaufen := starte(fallFrist, r.dir, []string{"GIT_CONFIG_GLOBAL=/dev/null"}, "git", a...)
	if abgelaufen || code != 0 {
		r.err = fmt.Errorf("git %s scheitert: %s", strings.Join(args, " "), letzte(out))
		return ""
	}
	return strings.TrimSpace(out)
}

// commit schreibt die Dateien, legt sie ab und liefert den Hash des neuen Commits.
func (r *gmRepo) commit(dateien map[string]string) string {
	if r.err == nil {
		r.err = gmSchreibe(r.dir, dateien)
	}
	return r.abgelegt()
}

// abgelegt legt den Arbeitsbaum ab und liefert den Hash des neuen Commits.
func (r *gmRepo) abgelegt() string {
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "gegenprobe")
	return r.git("rev-parse", "HEAD")
}

// plan schreibt den Plan mit der Basis und den Pfaden unter plaene/.
func (r *gmRepo) plan(basis string, pfade ...string) {
	if r.err != nil {
		return
	}
	aend := []string{"-pfad"}
	for _, p := range pfade {
		aend = append(aend, "pfad: "+p)
	}
	aend = append(aend, "basis: "+basis)
	r.err = gmSchreibe(r.dir, map[string]string{"plaene/slice-x.md": gmPlan(gmTabelle(0, 10), gmBlock(aend...))})
}

// gmGitFall ist ein Fall mit dem echten git und dem Programm groessenmarke: vor legt die Commits an.
func gmGitFall(name string, vor func(r *gmRepo), exit int, want ...string) Fall {
	f := gmFall(name, gateTreiber, nil, []string{"groessenmarke", "plaene"}, exit, want...)
	f.Verboten = append(f.Verboten, gmKeinSlice)
	f.Env = []string{"GIT_CONFIG_GLOBAL=/dev/null"}
	f.Vorbereitung = func(dir string) error {
		r := &gmRepo{dir: dir}
		r.git("init", "-q")
		vor(r)
		return r.err
	}
	return f
}

func gmZeilen(n int) string {
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, "%d\n", i)
	}
	return sb.String()
}

// gmGitAusschluesse ist Fall 115: je eine Datei mit 3 Einfügungen in sechs Bereichen, der Block nennt
// drei Pfade (tools/x/ und tools/x zählen die Datei einmal), harness/vertraege/ zählt trotz harness/
// nicht.
func gmGitAusschluesse() Fall {
	return gmGitFall("treiber-git-ausschluesse", func(r *gmRepo) {
		alt, neu := map[string]string{}, map[string]string{}
		for _, d := range []string{"tools/x/f.txt", "harness/mk/f.txt", "harness/vertraege/f.txt", "spec/f.txt", "docs/f.txt", "src/f.txt"} {
			alt[d], neu[d] = "0\n", "0\na\nb\nc\n"
		}
		basis := r.commit(alt)
		r.commit(neu)
		r.plan(basis, "tools/x/", "tools/x", "harness/")
	}, 0, "1>Messung 6 von Marke 1100")
}

// gmGitSonderzeichen ist Fall 116: Namen mit Leerzeichen, Stern, Klammern, Umlaut und Anführungszeichen;
// `ab` gehört nicht zum Pfad `a*`.
func gmGitSonderzeichen() Fall {
	return gmGitFall("treiber-git-sonderzeichen", func(r *gmRepo) {
		basis := r.commit(map[string]string{"README": "x\n"})
		neu := map[string]string{}
		for _, n := range []string{"a b.txt", "a*", "ab", "[x].txt", "\xc3\xa4.txt", "q\"r.txt"} {
			neu[n] = "1\n2\n"
		}
		r.commit(neu)
		r.plan(basis, "a*", "a b.txt", "[x].txt", "\xc3\xa4.txt", "q\"r.txt")
	}, 0, "1>Messung 10 von Marke 1100")
}

// gmGitUmbenennung ist Fall 117: mit diff.renames=true im Repo zählt die Umbenennung als Löschung
// und Einfügung.
func gmGitUmbenennung() Fall {
	return gmGitFall("treiber-git-umbenennung", func(r *gmRepo) {
		r.git("config", "diff.renames", "true")
		basis := r.commit(map[string]string{"tools/x/alt.txt": gmZeilen(10)})
		r.git("mv", "tools/x/alt.txt", "tools/x/neu.txt")
		r.abgelegt()
		r.plan(basis, "tools/x/")
	}, 0, "1>Messung 20 von Marke 1100")
}

// gmGitBinaer ist Fall 118: eine Binärdatei zählt nicht und steht im Hinweis.
func gmGitBinaer() Fall {
	return gmGitFall("treiber-git-binaer", func(r *gmRepo) {
		basis := r.commit(map[string]string{"README": "x\n"})
		r.commit(map[string]string{"tools/x/b.bin": "a\x00b\x00c\n", "tools/x/t.txt": gmZeilen(4)})
		r.plan(basis, "tools/x/")
	}, 0, "1>Messung 4 von Marke 1100", "1>1 Bin\xc3\xa4rdatei(en) nicht gez\xc3\xa4hlt")
}

// gmGitKeinVorfahr ist Fall 119: die Basis liegt auf einem anderen Zweig.
func gmGitKeinVorfahr() Fall {
	return gmGitFall("treiber-git-basis-kein-vorfahr", func(r *gmRepo) {
		r.commit(map[string]string{"README": "x\n"})
		r.git("checkout", "-q", "-b", "andere")
		basis := r.commit(map[string]string{"tools/x/f.txt": "1\n"})
		r.git("checkout", "-q", "main")
		r.commit(map[string]string{"tools/x/g.txt": "1\n"})
		r.plan(basis, "tools/x/")
	}, 2, "ist kein Vorfahr von HEAD")
}

// gmGitGleichSpitze ist Fall 120: die Basis ist HEAD als Kurzform des Hashs.
func gmGitGleichSpitze() Fall {
	return gmGitFall("treiber-git-basis-gleich-spitze", func(r *gmRepo) {
		r.commit(map[string]string{"README": "x\n"})
		r.plan(r.git("rev-parse", "--short=12", "HEAD"), "tools/x/")
	}, 0, "1>Messung 0 von Marke 1100")
}

// gmGitFaelle sind die Fälle 115 bis 120.
func gmGitFaelle() []Fall {
	return []Fall{gmGitAusschluesse(), gmGitSonderzeichen(), gmGitUmbenennung(), gmGitBinaer(), gmGitKeinVorfahr(), gmGitGleichSpitze()}
}

// groessenmarkeFaelle sind die Fälle 38 bis 120.
func groessenmarkeFaelle() []Fall {
	var alle []Fall
	for _, gruppe := range [][]Fall{gmMesspunktFaelle(), gmGrenzFaelle(), gmFormFaelle(), gmTabellenFaelle(), gmMessungFaelle(),
		gmTreiberGitFaelle(), gmTreiberPlanFaelle(), gmGitFaelle()} {
		alle = append(alle, gruppe...)
	}
	return alle
}
