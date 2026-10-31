package main

import (
	"os"
	"sort"
	"strings"
	"testing"
)

// Die Mengen dieser Datei bilden die Allaussage von SPEC-233 Nr. 5 a, e, f und Nr. 6 c: außer dem Teilbaum
// `.harness/baseline` wird jede Datei in jedem Verzeichnis gelesen, gleich unter welchem Namen und welcher
// Endung, und ohne Grenze der Größe. Die Erwartung folgt allein dem Pfad: ausgenommen ist, was mit
// `.harness/baseline/` beginnt.
// Prüfgrenze (akzeptiertes Negativ): ein Verzeichnispfad aus mehr als zwei Namen, der nicht in
// repoVerzeichnisse steht, ein Dateiname und eine Endung außerhalb von repoNamen und endungen sowie eine
// Größe über den Mengen von TestGrosseGrenzen sind nicht geprüft (unbeschränkte Mengen); eine Ausnahme
// für sie passiert bis zum Review.

// repoVerzeichnisse sind die Verzeichnisse des Repos außer `.harness/baseline` (ein Beobachtungs-Slug
// steht für alle); ihre Komponenten bilden die Menge der Namen, ihre Präfixe die der Pfade.
const repoVerzeichnisse = `.claude .claude/agents .claude/commands .claude/hooks .claude/rules docs/plan/adr
	docs/plan/carveouts docs/plan/planning docs/plan/planning/done docs/plan/planning/done/welle-fundament
	docs/plan/planning/done/welle-persistenz docs/plan/planning/in-progress docs/plan/planning/next
	docs/plan/planning/observations docs/plan/planning/observations/BEO-REPO/akzeptiertes-negativ-ohne-folge/evidence
	docs/plan/planning/open docs/reviews docs/user docs/user/beispiele .githooks .harness harness harness/conventions
	harness/mk harness/sensors .harness/skills harness/vertraege spec src/adapters/driven/blobs
	src/adapters/driven/clock src/adapters/driven/httpupstream src/adapters/driven/odata/csdl
	src/adapters/driven/odata/jsoncodec src/adapters/driven/sqlite src/adapters/driven/sqlite/migrations
	src/adapters/driven/telemetry src/bootstrap src/bootstrap/main src/hexagon/model src/hexagon/ports/inbound
	src/hexagon/ports/outbound src/hexagon/ports/outbound/outboundtest
	src/hexagon/ports/outbound/outboundtest/testdata src/hexagon/ports/support
	src/hexagon/ports/support/supporttest src/hexagon/services/baselineimport src/hexagon/services/maintenance
	src/hexagon/services/record src/hexagon/services/record/capture src/hexagon/services/session
	src/hexagon/services/workspace tools/gegenproben tools/groessenmarke tools/harness tools/harness/blocked
	tools/pruefung tools/schema tools/zeichen`

// repoNamen sind Dateinamen des Repos und gängiger Bäume.
const repoNamen = `Makefile Dockerfile LICENSE .gitignore .gitattributes .dockerignore .golangci.yml .a-check.yml
	.d-check.yml go.mod go.sum settings.json commit-msg .gitkeep AGENTS.md README.md CLAUDE.md roadmap.md
	SECURITY.md CHANGELOG.md schema.yaml gegenproben.stempel archiv.zip pretooluse-command-guard.sh package.json
	yarn.lock Cargo.lock .env .editorconfig`

// endungen sind die Endungen des Repos, die der Verifier nannte, und gängige weitere.
const endungen = `go md mk sh txt yml yaml xml zip stempel sql py python json awk sum mod golden csv lock snap pem html svg css
	js ts toml ini cfg conf log tsv proto bin dat tar gz png jpg pdf env tmpl diff patch rst adoc gitkeep gitignore
	tpl map wasm so a o class jar exe`

// verzeichnisMengen liefert die Namen (Komponenten von repoVerzeichnisse und `baseline`) und alle
// Verzeichnispfade der Kette samt ihren Präfixen, je sortiert.
func verzeichnisMengen() (namen, kette []string) {
	nm, km := map[string]bool{"baseline": true}, map[string]bool{}
	for _, d := range strings.Fields(repoVerzeichnisse) {
		teile := strings.Split(d, "/")
		for i := range teile {
			nm[teile[i]] = true
			km[strings.Join(teile[:i+1], "/")] = true
		}
	}
	for n := range nm {
		namen = append(namen, n)
	}
	for k := range km {
		kette = append(kette, k)
	}
	sort.Strings(namen)
	sort.Strings(kette)
	return namen, kette
}

// baumMitNbsp legt unter jedem Pfad eine Datei mit einem NBSP an (dazu `z.md` ohne Befund) und liefert die
// Wurzel und die erwartete Ausgabe: ein Befund je Pfad außerhalb von `.harness/baseline/`, in der Folge der
// Pfade.
func baumMitNbsp(t *testing.T, pfade []string) (wurzel, soll string) {
	t.Helper()
	dateien := map[string][]byte{"z.md": b("a\n")}
	var sichtbar []string
	for _, p := range pfade {
		dateien[p] = cat(b("x"), zb(0x00A0))
		if !strings.HasPrefix(p, ".harness/baseline/") {
			sichtbar = append(sichtbar, p)
		}
	}
	sort.Strings(sichtbar)
	for _, p := range sichtbar {
		soll += orakelZeile(p, 1, 2, 0x00A0)
	}
	return baum(t, dateien), soll + summe(len(sichtbar), len(sichtbar))
}

func zaehle(t *testing.T, was string, ist, soll int) {
	t.Helper()
	if ist != soll {
		t.Fatalf("%s: %d statt %d", was, ist, soll)
	}
}

// stellen sind die Verzeichnisse, in denen die Tests Dateien anlegen: die Wurzel (""), jeder Name allein
// und jeder Pfad der Kette; mit paare auch jedes Paar zweier Namen.
func stellen(namen, kette []string, paare bool) []string {
	m := map[string]bool{"": true}
	for _, a := range namen {
		m[a] = true
		for _, c := range namen {
			if paare {
				m[a+"/"+c] = true
			}
		}
	}
	for _, k := range kette {
		m[k] = true
	}
	var l []string
	for s := range m {
		l = append(l, s)
	}
	sort.Strings(l)
	return l
}

func unter(stelle, name string) string {
	if stelle == "" {
		return name
	}
	return stelle + "/" + name
}

// TestKeineAusnahmeNachPfad: kein Verzeichnispfad ist ausgenommen außer `.harness/baseline`. Menge:
// Verzeichnis (die Wurzel, die Namen allein, alle Paare zweier Namen und die Pfade der Kette samt
// Präfixen) mit einer Datei darin.
func TestKeineAusnahmeNachPfad(t *testing.T) {
	namen, kette := verzeichnisMengen()
	zaehle(t, "Namen", len(namen), 70)
	zaehle(t, "Pfade der Kette", len(kette), 70)
	var pfade []string
	for _, s := range stellen(namen, kette, true) {
		pfade = append(pfade, unter(s, "f.md"))
	}
	zaehle(t, "Verzeichnisse", len(pfade), 5012)
	wurzel, soll := baumMitNbsp(t, pfade)
	ersteAbweichung(t, "kein Verzeichnispfad ausgenommen", fahre(wurzel).err, soll)
}

// TestKeineAusnahmeNachDateiname: kein Dateiname und keine Endung ist ausgenommen, in keinem Verzeichnis.
// Menge: Stelle (die Wurzel, die Namen allein und die Pfade der Kette) mal Name (Dateinamen aus repoNamen
// und `f.<Endung>` je Endung aus endungen).
func TestKeineAusnahmeNachDateiname(t *testing.T) {
	namen, kette := verzeichnisMengen()
	dateinamen, exts := strings.Fields(repoNamen), strings.Fields(endungen)
	zaehle(t, "Dateinamen", len(dateinamen), 29)
	zaehle(t, "Endungen", len(exts), 58)
	st := stellen(namen, kette, false)
	zaehle(t, "Stellen", len(st), 133)
	var pfade []string
	for _, s := range st {
		for _, n := range dateinamen {
			pfade = append(pfade, unter(s, n))
		}
		for _, e := range exts {
			pfade = append(pfade, unter(s, "f."+e))
		}
	}
	zaehle(t, "Dateien", len(pfade), 133*(29+58))
	wurzel, soll := baumMitNbsp(t, pfade)
	ersteAbweichung(t, "kein Name und keine Endung ausgenommen", fahre(wurzel).err, soll)
}

// baselineAbwandlungen sind der Name c und seine Abwandlungen: Suffix, Präfix, Schreibweise, Abschnitte.
func baselineAbwandlungen(c string) []string {
	m := map[string]bool{c: true}
	for _, s := range []string{"x", "2", ".d", ".", "~", "-", "_", ".git", "s", "1", "_old", ".bak"} {
		m[c+s] = true
	}
	for _, p := range []string{"x", ".", "_", "-", "x.", "1"} {
		m[p+c] = true
	}
	m[strings.ToUpper(c)] = true
	m[strings.ToUpper(c[:1])+c[1:]] = true
	m[strings.ToUpper(c[:2])+c[2:]] = true
	m[c[:len(c)-1]] = true
	m[c[1:]] = true
	m[c[:len(c)/2]] = true
	var l []string
	for v := range m {
		l = append(l, v)
	}
	sort.Strings(l)
	return l
}

// TestBaselineNachbarn: nur das Verzeichnis `.harness/baseline` unmittelbar unter der Wurzel ist ausgenommen.
// Menge: Abwandlung des ersten Namens (`.harness`) mal Abwandlung des zweiten (`baseline`), je als Verzeichnis
// mit einer Datei darin und als Datei, jeweils an vier Stellen (Wurzel, `x/`, `a/b/`, `.harness/`).
func TestBaselineNachbarn(t *testing.T) {
	hs, bs := baselineAbwandlungen(".harness"), baselineAbwandlungen("baseline")
	zaehle(t, "Abwandlungen von .harness", len(hs), 24)
	zaehle(t, "Abwandlungen von baseline", len(bs), 25)
	var verz, dat []string
	for _, e := range []string{"", "x/", "a/b/", ".harness/"} {
		for _, h := range hs {
			for _, bl := range bs {
				verz = append(verz, e+h+"/"+bl+"/f.md")
				dat = append(dat, e+h+"/"+bl)
			}
		}
	}
	zaehle(t, "Pfade", len(verz), 4*24*25)
	// Als Verzeichnis ist allein das Paar selbst ausgenommen, als Datei keines; die Pfade beider Formen
	// liegen in getrennten Bäumen, weil derselbe Pfad nicht Datei und Verzeichnis sein kann.
	for _, f := range []struct {
		was   string
		pfade []string
	}{{"als Verzeichnis", verz}, {"als Datei", dat}} {
		wurzel, soll := baumMitNbsp(t, f.pfade)
		ersteAbweichung(t, "Abwandlungen von .harness/baseline "+f.was, fahre(wurzel).err, soll)
	}
	gleich(t, "das Paar als Datei", fahre(baum(t, map[string][]byte{".harness/baseline": cat(b("x"), zb(0x00A0))})).err,
		orakelZeile(".harness/baseline", 1, 2, 0x00A0)+summe(1, 1))
}

// TestGrosseGrenzen: SPEC-233 Nr. 6 c sagt „ohne Grenze für Dateigröße und Zeilenlänge“. Menge: Dateigrößen
// 4 MiB, 16 MiB und 64 MiB (je ein Byte darüber, Zeilen zu 16 Bytes, NBSP am Ende) und Zeilenlängen 4 MiB
// und 16 MiB (je ein Byte darüber, eine Zeile, NBSP am Ende). Prüfgrenze (akzeptiertes Negativ): eine Grenze
// über 64 MiB Dateigröße oder über 16 MiB Zeilenlänge ist nicht geprüft (Laufzeit und Platz der Stufe); Folge:
// sie passiert bis zum Review.
func TestGrosseGrenzen(t *testing.T) {
	const mib = 1 << 20
	wurzel := t.TempDir()
	schreibe := func(name string, n int, block []byte) {
		f, err := os.Create(wurzel + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		for w := 0; w < n; w += len(block) {
			if _, err := f.Write(block); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := f.Write(zb(0x00A0)); err != nil {
			t.Fatal(err)
		}
	}
	zeilen, eine := b("xxxxxxxxxxxxxxx\n"), b(strings.Repeat("x", 1024))
	schreibe("datei-16m.md", 16*mib+1, zeilen)
	schreibe("datei-4m.md", 4*mib+1, zeilen)
	schreibe("datei-64m.md", 64*mib+1, zeilen)
	schreibe("zeile-16m.md", 16*mib+1, eine)
	schreibe("zeile-4m.md", 4*mib+1, eine)
	// n+1 Bytes in Blöcken zu 16 Zeilenlängen ergeben n/16+1 Zeilen, das NBSP steht in Spalte 1 der Zeile
	// danach; Blöcke zu 1024 Bytes ergeben eine Zeile der Länge n+1024, das NBSP steht dahinter.
	soll := orakelZeile("datei-16m.md", 16*mib/16+2, 1, 0x00A0) + orakelZeile("datei-4m.md", 4*mib/16+2, 1, 0x00A0) +
		orakelZeile("datei-64m.md", 64*mib/16+2, 1, 0x00A0) + orakelZeile("zeile-16m.md", 1, 16*mib+1024+1, 0x00A0) +
		orakelZeile("zeile-4m.md", 1, 4*mib+1024+1, 0x00A0)
	r := fahre(wurzel)
	gleich(t, "große Dateien und Zeilen", r.err, soll+summe(5, 5))
	if r.code != 1 {
		t.Errorf("Exit %d", r.code)
	}
}
