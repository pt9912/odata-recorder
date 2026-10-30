package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Fälle 38 bis 120 (SPEC-102): `make groessenmarke` (SPEC-128). Die Fälle 38 bis 102 rufen das Programm
// groessenmarke mit einem Plan und einer Messung im Wegwerfverzeichnis, die Fälle 103 bis 114 den
// Treiber groessenmarke.sh mit Ersatz-Skripten für git und für das Programm, die Fälle 115 bis 120
// den Treiber mit dem echten git auf einem Repo, das der Fall anlegt. Zeichen jenseits von ASCII
// stehen nur als \x-Folgen (AGENTS.md §3.18).

const (
	gmBasis  = "0123456789abcdef0123456789abcdef01234567"
	gmSpitze = "fedcba9876543210fedcba9876543210fedcba98"
	gmKopf5  = "| Nr. | Zusage | F\xc3\xa4nger | Mutation | Klasse |\n|---|---|---|---|---|\n"
	gmNeben  = "nebenl\xc3\xa4ufig"
)

// gmTabelle ist die Tabelle mit a nebenläufigen, dann b anderen Zeilen.
func gmTabelle(a, b int) string {
	var sb strings.Builder
	sb.WriteString(gmKopf5)
	n := 0
	for i := 0; i < a+b; i++ {
		n++
		klasse := "andere"
		if i < a {
			klasse = gmNeben
		}
		fmt.Fprintf(&sb, "| Z%d | z | f | m | %s |\n", n, klasse)
	}
	return sb.String()
}

// gmBlock ist der Block der Fixture B (SPEC-102): 10 andere Zeilen, S = 800, Marke 1100. Ein
// Argument ersetzt die Zeile desselben Schlüssels oder hängt eine an (pfad, ueberschreitung);
// `-schluessel` entfernt die Zeilen des Schlüssels.
func gmBlock(aend ...string) []string {
	zeilen := []string{"pfad: src/", "produktcode: 0", "fakes: 0", "basis: " + gmBasis, "schaetzung: 800", "marke: 1100"}
	for _, a := range aend {
		if k, weg := strings.CutPrefix(a, "-"); weg {
			var rest []string
			for _, z := range zeilen {
				if !strings.HasPrefix(z, k+":") {
					rest = append(rest, z)
				}
			}
			zeilen = rest
			continue
		}
		k, _, _ := strings.Cut(a, ":")
		ersetzt := false
		for i := range zeilen {
			if strings.HasPrefix(zeilen[i], k+":") && k != "pfad" && k != "ueberschreitung" {
				zeilen[i], ersetzt = a, true
				break
			}
		}
		if !ersetzt {
			zeilen = append(zeilen, a)
		}
	}
	return zeilen
}

func gmPlan(tab string, block []string) string {
	return "# Slice\n\n" + tab + "\n```groessenmarke\n" + strings.Join(block, "\n") + "\n```\n"
}

// gmPlanB ist die Fixture B mit Änderungen am Block.
func gmPlanB(aend ...string) string {
	return gmPlan(gmTabelle(0, 10), gmBlock(aend...))
}

func gmSatz(ins, del, pfad string) string {
	return ins + "\t" + del + "\t" + pfad + "\x00"
}

func gmKopf(basis string, pfade ...string) string {
	s := "basis " + basis + "\nspitze " + gmSpitze + "\n"
	for _, p := range pfade {
		s += "pfad " + p + "\n"
	}
	return s + "numstat\n"
}

// gmMessungB ist die Messung der Fixture B mit der Summe n (ein Datensatz, bei 0 keiner).
func gmMessungB(n int) string {
	s := gmKopf(gmBasis, "src/")
	if n > 0 {
		s += gmSatz(fmt.Sprint(n), "0", "f.go")
	}
	return s
}

// gmFall baut einen Fall der Gruppe; ein Text mit `!` davor darf nicht in der Ausgabe stehen.
func gmFall(name string, gate Gate, neu map[string]string, args []string, exit int, want ...string) Fall {
	f := Fall{Name: "groessenmarke/" + name, Gate: gate, Wegwerfmodul: true, Neu: neu, Argumente: args, WantExit: exit, Getrennt: true}
	for _, w := range want {
		if v, nicht := strings.CutPrefix(w, "!"); nicht {
			f.Verboten = append(f.Verboten, v)
		} else if v, out := strings.CutPrefix(w, "1>"); out {
			f.WantOut = append(f.WantOut, v)
		} else if w != "" {
			f.WantErr = append(f.WantErr, w)
		}
	}
	return f
}

func gmPruefen(name, plan, messung string, exit int, want ...string) Fall {
	return gmFall(name, gateGroessenmarke, map[string]string{"plan.md": plan, "messung": messung},
		[]string{"pruefen", "{dir}/plan.md", "{dir}/messung"}, exit, want...)
}

func gmAbfrage(name, plan string, exit int, want ...string) Fall {
	return gmFall(name, gateGroessenmarke, map[string]string{"plan.md": plan},
		[]string{"abfrage", "{dir}/plan.md"}, exit, want...)
}

// gmPunkt ist ein Messpunkt aus SPEC-128 Nr. 2.
type gmPunkt struct {
	name                          string
	neben, zeilen, fakes, produkt int
	schaetzung, marke, ende       int
	platte                        int
}

func gmPunkte() []gmPunkt {
	return []gmPunkt{
		{"csdl", 2, 44, 0, 1506, 5036, 6800, 4333, 5600},
		{"json", 2, 49, 0, 1500, 5430, 7400, 7236, 6000},
		{"import", 3, 70, 8, 2000, 9215, 12500, 10027, 8200},
	}
}

func (p gmPunkt) plan(marke int) string {
	return gmPlan(gmTabelle(p.neben, p.zeilen-p.neben), gmBlock(
		fmt.Sprintf("produktcode: %d", p.produkt), fmt.Sprintf("fakes: %d", p.fakes),
		fmt.Sprintf("schaetzung: %d", p.schaetzung), fmt.Sprintf("marke: %d", marke), "schnitt-grund: Grund"))
}

// messpunktZeile ist die Messpunkt-Zeile (SPEC-128 Nr. 6 b) ohne Planname.
func (p gmPunkt) messpunktZeile(marke, messung int) string {
	return fmt.Sprintf("zeilen=%d nebenlaeufig=%d fakes=%d produktcode=%d schaetzung=%d marke=%d messung=%d",
		p.zeilen, p.neben, p.fakes, p.produkt, p.schaetzung, marke, messung)
}

// gmMesspunktFaelle sind die Fälle 38 bis 43: die drei Messpunkte mit der Marke des Vertrags und
// mit der Marke der Formel vor SPEC-128. Die Schätzung des Blocks gleicht der Tabelle, also steht
// weder ein Hinweis noch ein Befund zur Tabelle in der Ausgabe.
func gmMesspunktFaelle() []Fall {
	var fs []Fall
	for _, p := range gmPunkte() {
		fs = append(fs, gmPruefen("messpunkt-"+p.name, p.plan(p.marke), gmMessungB(p.ende), 0,
			fmt.Sprintf("1>ok, Messung %d von Marke %d", p.ende, p.marke), "1>"+p.messpunktZeile(p.marke, p.ende), "!Hinweis"))
	}
	for _, p := range gmPunkte() {
		fs = append(fs, gmPruefen("marke-der-platte-"+p.name, p.plan(p.platte), gmMessungB(0), 1,
			"weicht von der Formel ab", fmt.Sprintf("erwartet %d", p.marke), "1>"+p.messpunktZeile(p.platte, 0), "!weicht von der Tabelle ab"))
	}
	return fs
}

// gmGrenzFaelle sind die Fälle 44 bis 59: Marke, Schätzung, Messung, Einträge, Paare, Schnittfrage.
func gmGrenzFaelle() []Fall {
	formel := "weicht von der Formel ab"
	tabelle := "weicht von der Tabelle ab"
	const (
		ueberschreitung = "\xc3\x9cberschreitung"
		groesstes       = "gr\xc3\xb6\xc3\x9ftes N 1100"
		hinweisTabelle  = "Tabelle ergibt 800, Sch\xc3\xa4tzung 799"
	)
	schnitt := func(produkt, s, marke int, aend ...string) string {
		return gmPlan(gmTabelle(0, 27), gmBlock(append([]string{
			fmt.Sprintf("produktcode: %d", produkt), fmt.Sprintf("schaetzung: %d", s), fmt.Sprintf("marke: %d", marke),
		}, aend...)...))
	}
	return []Fall{
		gmPruefen("marke-eine-stufe-zu-niedrig", gmPlanB("marke: 1000"), gmMessungB(0), 1, formel, "erwartet 1100"),
		gmPruefen("marke-eine-stufe-zu-hoch", gmPlanB("marke: 1200"), gmMessungB(0), 1, formel, "erwartet 1100"),
		gmPruefen("schaetzung-um-eins-zu-niedrig", gmPlanB("schaetzung: 799"), gmMessungB(0), 1, tabelle, "Tabelle ergibt 800", "1>schaetzung=800 marke=1100", "!"+formel),
		gmPruefen("schaetzung-um-eins-zu-hoch", gmPlanB("schaetzung: 801"), gmMessungB(0), 1, tabelle, "Tabelle ergibt 800", "!"+formel),
		gmPruefen("tabelle-gewachsen-im-code", gmPlanB("schaetzung: 799"), gmMessungB(1), 0, "1>Hinweis", "1>"+hinweisTabelle, "1>schaetzung=800 marke=1100"),
		gmPruefen("messung-gleich-marke", gmPlanB(), gmMessungB(1100), 0, "1>ok, Messung 1100 von Marke 1100"),
		gmPruefen("messung-ueber-marke-ohne-eintrag", gmPlanB(), gmMessungB(1101), 1, "ohne Eintrag", "ueberschreitung: 1200 |"),
		gmPruefen("eintrag-deckt", gmPlanB("ueberschreitung: 1101 | Grund"), gmMessungB(1101), 0),
		gmPruefen("eintrag-reicht-nicht", gmPlanB("ueberschreitung: 1100 | Grund"), gmMessungB(1101), 1, "Eintrag reicht nicht", groesstes),
		gmPruefen("eintrag-ohne-grund", gmPlanB("ueberschreitung: 1200 |"), gmMessungB(1101), 1, "ohne Grund", "!ohne Eintrag"),
		gmPruefen("eintrag-ohne-ueberschreitung", gmPlanB("ueberschreitung: 1200 | Grund"), gmMessungB(1100), 0, "1>Eintrag ohne "+ueberschreitung),
		gmPruefen("paar-marke-unter-formel-und-messung-ueber-marke", gmPlanB("marke: 1000"), gmMessungB(1001), 1, formel, "ohne Eintrag"),
		gmPruefen("paar-eintrag-und-marke-unter-formel", gmPlanB("marke: 1000", "ueberschreitung: 1100 | Grund"), gmMessungB(1001), 1, formel, "!ohne Eintrag"),
		gmPruefen("schnitt-ohne-grund", schnitt(63, 2223, 3100), gmMessungB(0), 1, "Schnittfrage ohne Grund"),
		gmPruefen("schnitt-mit-grund", schnitt(63, 2223, 3100, "schnitt-grund: Grund"), gmMessungB(0), 0),
		gmPruefen("schnitt-genau-darunter", schnitt(62, 2222, 3000), gmMessungB(0), 0),
	}
}

// gmFormFaelle sind die Fälle 60 bis 78: Zustand des Plans, Schlüssel, Zahlen und Pfade des Blocks.
func gmFormFaelle() []Fall {
	prosa := "# Slice\n\n**Gr\xc3\xb6\xc3\x9fenmarke** (`AGENTS.md` \xc2\xa73.14): \xc3\xbcber 2400 Zeilen (Summe)\n"
	fs := []Fall{
		gmAbfrage("kein-block", "# Slice\n\nText ohne Block.\n", 0, "1>keine-messung", "kein Block"),
		gmAbfrage("prosa-form", prosa, 0, "1>keine-messung", "Prosa-Form"),
		gmAbfrage("marke-steht-aus", gmPlanB("-basis", "-schaetzung", "-marke"), 0, "1>keine-messung", "Marke steht aus: Sch\xc3\xa4tzung 800, Marke 1100"),
		gmAbfrage("paar-marke-und-basis-fehlen", gmPlanB("-basis", "-marke"), 2, "Block unvollst\xc3\xa4ndig: es fehlen basis, marke"),
		gmPruefen("block-unbekannter-schluessel", gmPlanB("marken: 1100"), gmMessungB(0), 2, "Schl\xc3\xbcssel marken unbekannt"),
	}
	zahlen := [][2]string{{"tausender-leerzeichen", "1 500"}, {"tausender-punkt", "1.500"}, {"minus", "-1"},
		{"fuehrende-null", "01"}, {"leer", ""}, {"zehn-stellen", "1234567890"}}
	for _, z := range zahlen {
		fs = append(fs, gmPruefen("zahl-"+z[0], gmPlanB("produktcode: "+z[1]), gmMessungB(0), 2, "Zahl ung\xc3\xbcltig: produktcode"))
	}
	pfade := [][2]string{{"leer", ""}, {"absolut", "/src"}, {"punktpunkt", "a/../b"}, {"magie", ":(top)src"}, {"doppelt", "src/"},
		{"in-spec", "spec/lastenheft.md"}, {"in-docs", "docs/plan"}, {"in-vertraegen", "harness/vertraege/README.md"}}
	for _, p := range pfade {
		text := "Pfad ung\xc3\xbcltig: " + p[1]
		if p[0] == "doppelt" {
			text = "Pfad doppelt: " + p[1]
		}
		fs = append(fs, gmPruefen("pfad-"+p[0], gmPlanB("pfad: "+p[1]), gmMessungB(0), 2, text))
	}
	return fs
}

// gmTabellenFaelle sind die Fälle 79 bis 88: Tabelle, Zeilen, Nr. und Klasse.
func gmTabellenFaelle() []Fall {
	block := gmBlock()
	zeile := func(nr, klasse string) string {
		return "| " + nr + " | z | f | m | " + klasse + " |\n"
	}
	fs := []Fall{
		gmPruefen("tabelle-fehlt", gmPlan("| Zeile | Zusage | Mutation | roter Test |\n|---|---|---|---|\n| 1 | z | m | t |\n", block), gmMessungB(0), 2, "Tabelle fehlt"),
		gmPruefen("tabelle-ohne-zeilen", gmPlan(gmKopf5, block), gmMessungB(0), 2, "Tabelle ohne Zeilen"),
		gmPruefen("tabelle-mehrdeutig", gmPlan(gmTabelle(0, 2)+"\nText\n\n"+gmTabelle(0, 2), block), gmMessungB(0), 2, "Tabelle mehrdeutig"),
		gmPruefen("tabelle-zellenzahl", gmPlan(gmKopf5+"| Z1 | `a|b` | f | m | andere |\n", block), gmMessungB(0), 2, "Zeile 1: 6 Zellen statt 5"),
		gmPruefen("tabelle-nr-doppelt", gmPlan(gmKopf5+zeile("Z1", "andere")+zeile("Z1", "andere"), block), gmMessungB(0), 2, "doppelt"),
	}
	klassen := [][2]string{{"leer", ""}, {"ascii-ae", "nebenlaeufig"}, {"grossbuchstabe", "Nebenl\xc3\xa4ufig"},
		{"zaehlt-nicht", "z\xc3\xa4hlt nicht"}, {"fett", "**andere**"}}
	for _, k := range klassen {
		fs = append(fs, gmPruefen("klasse-"+k[0], gmPlan(gmKopf5+zeile("Z1", k[1]), block), gmMessungB(0), 2, "Klasse", "unbekannt", k[1]))
	}
	return fs
}

// gmMessungFaelle sind die Fälle 89 bis 102: Datensätze und Kopf der Messung, Aufruf.
func gmMessungFaelle() []Fall {
	kopf := gmKopf(gmBasis, "src/")
	zwei := gmPlanB("pfad: tools/")
	fremd := "geh\xc3\xb6rt nicht zu diesem Plan"
	lesbar := "Messung nicht lesbar"
	return []Fall{
		gmPruefen("messung-binaer", gmPlanB(), kopf+gmSatz("-", "-", "x.bin")+gmSatz("3", "2", "y.go"), 0,
			"1>Messung 5 von Marke 1100", "1>1 Bin\xc3\xa4rdatei(en) nicht gez\xc3\xa4hlt"),
		gmPruefen("messung-umbenennung", gmPlanB(), kopf+"1\t2\t\x00", 2, "Umbenennung in der Messung"),
		gmPruefen("messung-leer", gmPlanB(), kopf, 0, "1>Messung 0 von Marke 1100"),
		gmPruefen("messung-fremde-basis", gmPlanB(), gmKopf("abcdef0123", "src/"), 2, fremd),
		gmPruefen("messung-pfad-fehlt", zwei, gmKopf(gmBasis, "src/"), 2, fremd),
		gmPruefen("messung-pfad-zusaetzlich", zwei, gmKopf(gmBasis, "src/", "tools/", "x/"), 2, fremd),
		gmPruefen("messung-pfade-vertauscht", zwei, gmKopf(gmBasis, "tools/", "src/"), 2, fremd),
		gmPruefen("messung-kopf-fehlt", gmPlanB(), gmSatz("1", "2", "x"), 2, lesbar),
		gmPruefen("messung-ohne-numstat-zeile", gmPlanB(), "basis "+gmBasis+"\nspitze "+gmSpitze+"\npfad src/\n"+gmSatz("1", "2", "x"), 2, lesbar),
		gmPruefen("messung-datensatz-ohne-nul", gmPlanB(), kopf+"1\t2\tx", 2, lesbar),
		gmPruefen("messung-datensatz-ohne-tab", gmPlanB(), kopf+"12x\x00", 2, lesbar),
		gmFall("aufruf", gateGroessenmarke, nil, nil, 2, "Aufruf:"),
		gmFall("plan-fehlt", gateGroessenmarke, map[string]string{"messung": gmMessungB(0)},
			[]string{"pruefen", "{dir}/gibt-es-nicht", "{dir}/messung"}, 2, "Plan fehlt"),
		gmFall("messung-fehlt", gateGroessenmarke, map[string]string{"plan.md": gmPlanB()},
			[]string{"pruefen", "{dir}/plan.md", "{dir}/gibt-es-nicht"}, 2, "Messung fehlt"),
	}
}

// Die Ersatz-Skripte schreiben jeden Aufruf mit allen Argumenten in ihr Protokoll (log/git.log,
// log/programm.log: je Aufruf eine Zeile AUFRUF und je Argument eine Zeile ARG), auch für Operationen,
// die der Treiber nie rufen darf. Ausgabe und Exit-Code bestimmen die Dateien unter ersatz/; der
// Ersatz des Programms schreibt je Kommando einen Text auf stdout und einen auf stderr. Eine Datei
// `<Operation>.signal` lässt den Ersatz dem Treiber (seinem Elternprozess) das genannte Signal senden.
const (
	gmAbfrageOut = "ABFRAGE-STDOUT"
	gmAbfrageErr = "ABFRAGE-STDERR"
	gmPruefenOut = "PRUEFEN-STDOUT"
	gmPruefenErr = "PRUEFEN-STDERR"

	ersatzGit = `#!/bin/sh
{ echo AUFRUF; for a in "$@"; do printf 'ARG %s\n' "$a"; done; } >> log/git.log
sub=$1
rc=0
[ -f "ersatz/git.$sub.rc" ] && read -r rc < "ersatz/git.$sub.rc"
[ -f "ersatz/git.$sub.err" ] && cat "ersatz/git.$sub.err" >&2
[ -f "ersatz/git.$sub.out" ] && cat "ersatz/git.$sub.out"
[ -f "ersatz/git.$sub.signal" ] && { read -r sig < "ersatz/git.$sub.signal"; kill -"$sig" "$PPID"; }
exit "$rc"
`
	ersatzProgramm = `#!/bin/sh
{ echo AUFRUF; for a in "$@"; do printf 'ARG %s\n' "$a"; done; } >> log/programm.log
name=${2##*/}
rc=0
case $1 in
abfrage)
	echo ` + gmAbfrageOut + `
	echo ` + gmAbfrageErr + ` >&2
	[ -f "ersatz/$2.abfrage.err" ] && cat "ersatz/$2.abfrage.err" >&2
	[ -f "ersatz/$2.abfrage" ] && cat "ersatz/$2.abfrage"
	[ -f "ersatz/$2.abfrage.signal" ] && { read -r sig < "ersatz/$2.abfrage.signal"; kill -"$sig" "$PPID"; }
	[ -f "ersatz/$2.abfrage.rc" ] && read -r rc < "ersatz/$2.abfrage.rc"
	exit "$rc" ;;
pruefen)
	echo ` + gmPruefenOut + `
	echo ` + gmPruefenErr + ` >&2
	cat "$3" > "log/messung-$name"
	[ -f "ersatz/$2.pruefen.signal" ] && { read -r sig < "ersatz/$2.pruefen.signal"; kill -"$sig" "$PPID"; }
	[ -f "ersatz/$2.pruefen.rc" ] && read -r rc < "ersatz/$2.pruefen.rc"
	exit "$rc" ;;
esac
exit 99
`
)

// gmEintrag ist ein Plan des Planverzeichnisses eines Treiber-Falls und die Antworten des Ersatzes
// für das Programm: Ausgabe und Hinweis von `abfrage`, Exit-Code von `pruefen`.
type gmEintrag struct {
	pfad, abfrage, hinweis, pruefenRc string
}

const gmMessen = "messen\nbasis " + gmBasis + "\npfad src/\n"

// gmErsatz liefert die Dateien eines Falls mit Ersatz-Skripten, die Protokolle (leer) und die
// Antworten für die Pläne; extra ergänzt oder überschreibt Dateien.
func gmErsatz(extra map[string]string, plaene ...gmEintrag) map[string]string {
	m := map[string]string{
		"bin/git": ersatzGit, "bin/programm": ersatzProgramm, "log/git.log": "", "log/programm.log": "",
		"ersatz/git.rev-parse.out": gmSpitze + "\n", "plaene/.keep": "",
	}
	for _, p := range plaene {
		m[p.pfad] = "x\n"
		m["ersatz/"+p.pfad+".abfrage"] = p.abfrage
		if p.hinweis != "" {
			m["ersatz/"+p.pfad+".abfrage.err"] = p.hinweis
		}
		if p.pruefenRc != "" {
			m["ersatz/"+p.pfad+".pruefen.rc"] = p.pruefenRc + "\n"
		}
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

// gmTreiber baut einen Fall mit dem Treiber, den Ersatz-Skripten am Anfang von PATH und dem
// Planverzeichnis plaene/.
func gmTreiber(name string, neu map[string]string, exit int, nachher func(Lauf) string, want ...string) Fall {
	f := gmFall(name, gateTreiber, neu, []string{"{dir}/bin/programm", "plaene"}, exit, want...)
	if !slices.Contains(f.WantOut, gmKeinSlice) {
		f.Verboten = append(f.Verboten, gmKeinSlice)
	}
	f.Ausfuehrbar = []string{"bin/git", "bin/programm"}
	f.Env = []string{"PATH={dir}/bin:{PATH}", "GIT_CONFIG_GLOBAL=/dev/null"}
	f.Nachher = nachher
	return f
}

// gmProtokoll liest ein Protokoll der Ersatz-Skripte: je Aufruf die Liste seiner Argumente. Eine
// fehlende Datei ist ein Fehler.
func gmProtokoll(dir, datei string) ([][]string, error) {
	b, err := os.ReadFile(filepath.Join(dir, "log", datei))
	if err != nil {
		return nil, fmt.Errorf("Protokoll %s fehlt", datei)
	}
	var aufrufe [][]string
	for _, z := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		switch {
		case z == "AUFRUF":
			aufrufe = append(aufrufe, []string{})
		case strings.HasPrefix(z, "ARG ") && len(aufrufe) > 0:
			aufrufe[len(aufrufe)-1] = append(aufrufe[len(aufrufe)-1], strings.TrimPrefix(z, "ARG "))
		case z != "":
			return nil, fmt.Errorf("Protokoll %s: unlesbare Zeile %q", datei, z)
		}
	}
	return aufrufe, nil
}

// gmErwartet vergleicht das Protokoll datei mit den erwarteten Aufrufen, Argument für Argument; ein
// fehlendes Protokoll ist ein Grund. Das Argument `*messung` steht für die Datei der Messung im
// Verzeichnis, das mktemp unter dem TMPDIR des Falls anlegt.
func gmErwartet(l Lauf, datei string, want [][]string) string {
	got, err := gmProtokoll(l.Dir, datei)
	if err != nil {
		return err.Error()
	}
	if len(got) != len(want) {
		return fmt.Sprintf("Protokoll %s: %d Aufrufe statt %d: %q", datei, len(got), len(want), got)
	}
	for i := range got {
		if !gmGleichAufruf(l, got[i], want[i]) {
			return fmt.Sprintf("Protokoll %s, Aufruf %d: %q statt %q", datei, i+1, got[i], want[i])
		}
	}
	return ""
}

func gmGleichAufruf(l Lauf, got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if want[i] == "*messung" {
			if !strings.HasPrefix(got[i], l.Tmp+"/") || !strings.HasSuffix(got[i], "/messung") {
				return false
			}
		} else if got[i] != want[i] {
			return false
		}
	}
	return true
}

// gmErste liefert den ersten Grund, der nicht leer ist.
func gmErste(gruende ...string) string {
	for _, g := range gruende {
		if g != "" {
			return g
		}
	}
	return ""
}

func gmDiff(pfade ...string) []string {
	a := []string{"diff", "--numstat", "-z", "--no-renames", gmBasis + "..HEAD", "--"}
	for _, p := range pfade {
		a = append(a, ":(literal)"+p)
	}
	return append(a, ":(exclude,literal)harness/vertraege")
}

func gmRevParse() []string {
	return []string{"rev-parse", "--verify", "HEAD"}
}

func gmMergeBase() []string {
	return []string{"merge-base", "--is-ancestor", gmBasis, "HEAD"}
}

const gmPlanA = "plaene/slice-a.md"

// gmTreiberGitFaelle sind die Fälle 103 bis 109: die Aufrufe von git und ihre Fehler.
func gmTreiberGitFaelle() []Fall {
	messen := gmEintrag{pfad: gmPlanA, abfrage: gmMessen}
	nur := func(l Lauf, git ...[]string) string {
		return gmErste(gmErwartet(l, "git.log", git), gmErwartet(l, "programm.log", [][]string{{"abfrage", gmPlanA}}))
	}
	fehler := func(name, sub, rc, meldung, text string, git ...[]string) Fall {
		extra := map[string]string{"ersatz/git." + sub + ".rc": rc + "\n"}
		if meldung != "" {
			extra["ersatz/git."+sub+".err"] = meldung + "\n"
		}
		return gmTreiber(name, gmErsatz(extra, messen), 2, func(l Lauf) string { return nur(l, git...) }, text, meldung)
	}
	return []Fall{
		gmTreiber("treiber-aufrufe-von-git", gmErsatz(map[string]string{"ersatz/git.diff.out": gmSatz("3", "4", "src/a/x.go")},
			gmEintrag{pfad: gmPlanA, abfrage: "messen\nbasis " + gmBasis + "\npfad src/a\npfad src/b\n"}), 0, gmPruefeAufrufe,
			"1>"+gmPruefenOut, gmPruefenErr, gmAbfrageErr, "!"+gmAbfrageOut),
		gmTreiber("treiber-pfad-argumente", gmErsatz(nil, gmEintrag{pfad: gmPlanA, abfrage: "messen\nbasis " + gmBasis +
			"\npfad a b\npfad a*\npfad a\"b\npfad [x]\npfad \xc3\xa4\n"}), 0, func(l Lauf) string {
			return gmErwartet(l, "git.log", [][]string{gmRevParse(), gmMergeBase(), gmDiff("a b", "a*", "a\"b", "[x]", "\xc3\xa4")})
		}),
		fehler("treiber-basis-kein-vorfahr", "merge-base", "1", "", "ist kein Vorfahr von HEAD", gmRevParse(), gmMergeBase()),
		fehler("treiber-basis-unbekannt", "merge-base", "128", "fatal: Not a valid commit name", "unbekannt oder git scheitert", gmRevParse(), gmMergeBase()),
		fehler("treiber-head-nicht-lesbar", "rev-parse", "128", "", "HEAD nicht lesbar", gmRevParse()),
		fehler("treiber-git-diff-scheitert", "diff", "128", "fatal: boese", "git diff scheitert", gmRevParse(), gmMergeBase(), gmDiff("src/")),
		gmOhneGit(),
	}
}

// gmPruefeAufrufe prüft Fall 103: die drei Aufrufe von git mit den Pfaden src/a und src/b, die zwei
// Aufrufe des Programms und die Datei der Messung.
func gmPruefeAufrufe(l Lauf) string {
	git := gmErwartet(l, "git.log", [][]string{gmRevParse(), gmMergeBase(), gmDiff("src/a", "src/b")})
	programm := gmErwartet(l, "programm.log", [][]string{{"abfrage", gmPlanA}, {"pruefen", gmPlanA, "*messung"}})
	messung, err := os.ReadFile(filepath.Join(l.Dir, "log", "messung-slice-a.md"))
	want := "basis " + gmBasis + "\nspitze " + gmSpitze + "\npfad src/a\npfad src/b\nnumstat\n" + gmSatz("3", "4", "src/a/x.go")
	if err != nil || string(messung) != want {
		return fmt.Sprintf("Messung %q statt %q (%v)", messung, want, err)
	}
	return gmErste(git, programm)
}

// gmOhneGit ist Fall 109: PATH trägt nur mktemp, rm und cat, kein git.
func gmOhneGit() Fall {
	f := gmTreiber("treiber-git-nicht-aufrufbar", gmErsatz(nil, gmEintrag{pfad: gmPlanA, abfrage: gmMessen}), 2, func(l Lauf) string {
		return gmErste(gmErwartet(l, "git.log", nil), gmErwartet(l, "programm.log", [][]string{{"abfrage", gmPlanA}}))
	}, "git nicht aufrufbar")
	f.Env = []string{"PATH={dir}/ohne-git", "GIT_CONFIG_GLOBAL=/dev/null"}
	f.Vorbereitung = func(dir string) error {
		if err := os.MkdirAll(filepath.Join(dir, "ohne-git"), 0o755); err != nil {
			return err
		}
		for _, p := range []string{"mktemp", "rm", "cat"} {
			pfad, err := exec.LookPath(p)
			if err != nil {
				return err
			}
			if err := os.Symlink(pfad, filepath.Join(dir, "ohne-git", p)); err != nil {
				return err
			}
		}
		return nil
	}
	return f
}
