package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Gemeinsame Fixtures und das Ersatz-Binary der Tests von `pruefung handbuch` (SPEC-200).

const aufrufDatei = "$ docker run --rm odata-recorder:build /out/odata-recorder\n" +
	"--- Standardausgabe\nodata-recorder dev\n--- Fehlerausgabe\n--- Ausgangswert\n0\n"

const kopfSauber = "# Handbuch\n\nSoftware-Version: dev\nGültigkeitsbereich: Binary\nStand: 2026-10-30\n"

func binde(p string) string {
	return "Datei: [" + p + "](" + p + ")\n"
}

// zaunVon liefert den umzäunten Block mit den Zeilen der Datei (ohne Schluss-LF je Zeile doppelt).
func zaunVon(inhalt string) string {
	return "```\n" + inhalt + "```\n"
}

func sauber() string {
	return kopfSauber + "\n## Aufruf\n\n" + binde("beispiele/aufruf.txt") + zaunVon(aufrufDatei)
}

// antwort ist die Antwort des Ersatz-Binary auf eine Argumentliste.
type antwort struct {
	out, err string
	code     int
	signal   bool
	schlaf   string // Sekunden, die das Ersatz-Binary schläft (dann ohne Ausgabe)
	kind     string // Sekunden, die ein Kindprozess die Ströme offen hält, nachdem das Ersatz-Binary geendet hat
}

func standardAntworten() map[string]antwort {
	ok := antwort{out: "odata-recorder dev\n"}
	return map[string]antwort{
		"version": ok,
		"":        ok,
		"foo":     {err: "Unbekanntes Kommando: foo\n", code: 2},
	}
}

// ersatz ist ein Shell-Skript als Binary. Es hält je Start Argumente, Zahl der Umgebungseinträge,
// Arbeitsverzeichnis, Zahl seiner Einträge und gelesene Bytes der Standardeingabe fest, trennt
// die Ströme, endet mit Code oder Signal, schläft auf Wunsch und zählt jeden Start, auch einen
// unerwarteten (Exit 99).
type ersatz struct {
	bin, dir string
}

func neuesErsatz(t *testing.T, a map[string]antwort) ersatz {
	t.Helper()
	d := t.TempDir()
	var faelle strings.Builder
	n := 0
	for schluessel, v := range a {
		n++
		o := filepath.Join(d, "r"+strconv.Itoa(n))
		if err := os.MkdirAll(o, 0o755); err != nil {
			t.Fatal(err)
		}
		dateien := map[string]string{"out": v.out, "err": v.err, "code": strconv.Itoa(v.code)}
		if v.signal {
			dateien["sig"] = ""
		}
		if v.schlaf != "" {
			dateien["sleep"] = v.schlaf
		}
		if v.kind != "" {
			dateien["kind"] = v.kind
		}
		for name, inhalt := range dateien {
			if err := os.WriteFile(filepath.Join(o, name), []byte(inhalt), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		faelle.WriteString("'" + schluessel + "') r=" + strconv.Itoa(n) + ";;\n")
	}
	skript := `#!/bin/sh
d=${0%/*}
{
  echo START
  for a in "$@"; do echo "ARG:$a"; done
  echo "ENV:$(env | grep -cvE '^(PWD|OLDPWD|SHLVL|_)=')"
  echo "CWD:$(pwd)"
  echo "ENT:$(ls -A | wc -l)"
  echo "IN:$(cat | wc -c)"
} >> "$d/log"
case "$*" in
` + faelle.String() + `*) echo unerwartet >&2; exit 99;;
esac
o=$d/r$r
[ -f "$o/sleep" ] && { echo $$ > "$d/pid"; exec sleep "$(cat "$o/sleep")"; }
[ -f "$o/kind" ] && { sleep "$(cat "$o/kind")" & }
cat "$o/out"; cat "$o/err" >&2
[ -f "$o/sig" ] && kill -KILL $$
exit "$(cat "$o/code")"
`
	bin := filepath.Join(d, "odata-recorder")
	if err := os.WriteFile(bin, []byte(skript), 0o755); err != nil {
		t.Fatal(err)
	}
	return ersatz{bin: bin, dir: d}
}

// erwarteN prüft den Exit-Code und dass die Ausgabe jeden Text nennt.
func erwarteN(t *testing.T, code int, errText string, wantCode int, texte ...string) {
	t.Helper()
	if code != wantCode {
		t.Fatalf("Exit %d (erwartet %d), Ausgabe %q", code, wantCode, errText)
	}
	enthaelt2(t, errText, texte...)
}

// start ist ein Start des Ersatz-Binary.
type start struct {
	args         []string
	env, ent, in int
	cwd          string
}

func (e ersatz) starts(t *testing.T) []start {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(e.dir, "log"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []start
	for _, z := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		k, v, _ := strings.Cut(z, ":")
		switch k {
		case "START":
			out = append(out, start{})
		case "ARG":
			out[len(out)-1].args = append(out[len(out)-1].args, v)
		case "ENV":
			out[len(out)-1].env = zahl(v)
		case "ENT":
			out[len(out)-1].ent = zahl(v)
		case "IN":
			out[len(out)-1].in = zahl(v)
		case "CWD":
			out[len(out)-1].cwd = v
		}
	}
	return out
}

func zahl(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// umfeld ist eine Wurzel mit sauberem Handbuch, Beispiel, Makefile und Dockerfile samt Ersatz-Binary.
type umfeld struct {
	t      *testing.T
	wurzel string
	e      ersatz
	u      Umgebung
}

func neuesUmfeld(t *testing.T, a map[string]antwort) umfeld {
	t.Helper()
	m := umfeld{t: t, wurzel: t.TempDir(), e: neuesErsatz(t, a), u: Umgebung{Zeitgrenze: 10 * time.Second}}
	m.schreibe(handbuchPfad, sauber())
	m.schreibe("docs/user/beispiele/aufruf.txt", aufrufDatei)
	m.schreibe("Makefile", "build: ## bauen\n\tdocker build --target build -t odata-recorder:build .\n")
	m.schreibe("Dockerfile", "RUN go build -o /out/odata-recorder ./x\n")
	return m
}

func (m umfeld) schreibe(rel, inhalt string) {
	m.t.Helper()
	p := filepath.Join(m.wurzel, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		m.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(inhalt), 0o644); err != nil {
		m.t.Fatal(err)
	}
}

func (m umfeld) entferne(rel string) {
	m.t.Helper()
	if err := os.RemoveAll(filepath.Join(m.wurzel, rel)); err != nil {
		m.t.Fatal(err)
	}
}

func (m umfeld) handbuch(h string) {
	m.t.Helper()
	m.schreibe(handbuchPfad, h)
}

func (m umfeld) lauf() (int, string, string) {
	return fuehre(m.u, "handbuch", m.e.bin, m.wurzel)
}

// naechste ist die Zeile, in der der erste Inhalt steht, den ein Test mit einer Leerzeile
// Abstand hinter h anhängt.
func naechste(h string) int {
	return strings.Count(h, "\n") + 2
}

func enthaelt2(t *testing.T, errText string, teile ...string) {
	t.Helper()
	for _, s := range teile {
		if !strings.Contains(errText, s) {
			t.Fatalf("Ausgabe nennt %q nicht: %q", s, errText)
		}
	}
}

func nichtEnthalten(t *testing.T, errText string, teile ...string) {
	t.Helper()
	for _, s := range teile {
		if strings.Contains(errText, s) {
			t.Fatalf("Ausgabe nennt %q: %q", s, errText)
		}
	}
}

// mf ist ein Muster der Mengen mit seinen Formen: das Orakel der Tests, getrennt von den Mustern des Programms.
type mf struct {
	id, ausdruck string
	formen       []string
}

func orakel() []mf {
	z := func(id, ausdruck, formen string) mf { return mf{id, ausdruck, strings.Split(formen, "|")} }
	return []mf{
		z("A1", `geplant\p{L}*`, "geplant|geplante|geplanten|Geplantes"),
		z("A2", `planung\p{L}*`, "Planung|in Planung|Planungen"),
		z("A3", `vorgesehen\p{L}*`, "vorgesehen|vorgesehene"),
		z("A4", `(künftig|kuenftig)\p{L}*`, "künftig|künftige|kuenftig"),
		z("A5", `(zukünftig|zukuenftig)\p{L}*`, "zukünftig|zukünftige|zukuenftig"),
		z("A6", `zukunft\p{L}*`, "Zukunft|in Zukunft"),
		z("A7", `(spät|spaet)\p{L}*`, "später|spätere|spaeter|spät"),
		z("A8", `(demnächst|demnaechst)`, "demnächst|demnaechst"),
		z("A9", `bald\p{L}*`, "bald"),
		z("A10", `folgt`, "folgt|wie folgt"),
		z("A11", `noch nicht`, "noch nicht"),
		z("A12", `noch kein\p{L}*`, "noch kein|noch keine|noch keinen"),
		z("A13", `nicht (verfügbar|verfuegbar)`, "nicht verfügbar|nicht verfuegbar"),
		z("A14", `wird kommen`, "wird kommen"),
		z("A15", `ausblick\p{L}*`, "Ausblick|Ausblicke"),
		z("A16", `irgendwann`, "irgendwann"),
		z("B1", `bisher\p{L}*`, "bisher|bisherige|bisherigen"),
		z("B2", `(früher|frueher)\p{L}*`, "früher|frühere|frueher"),
		z("B3", `neu (in|ab|seit)`, "neu in|neu ab|neu seit"),
		z("B4", `seit (version|release|v\d)`, "seit Version|seit Release|seit v2"),
		z("B5", `ab (version|release|v\d)`, "ab Version|ab Release|ab v2"),
		z("B6", `nicht mehr`, "nicht mehr"),
		z("B7", `(änderungs|aenderungs)\p{L}*`, "Änderungshistorie|Änderungsverlauf|Aenderungsdatum"),
		z("B8", `changelog`, "Changelog"),
		z("B9", `release[ -]?notes?`, "Release Notes|Release-Note|Releasenotes"),
		z("B10", `\p{L}*verlauf\p{L}*`, "Verlauf|Versionsverlauf"),
		z("B11", `\p{L}*historie\p{L}*`, "Historie|Versionshistorie"),
		z("B12", `handbuch-version`, "Handbuch-Version"),
		z("B13", `wurden? (geändert|eingeführt|entfernt|ergänzt|hinzugefügt|ersetzt|umbenannt|behoben|korrigiert|verbessert)`,
			"wurde geändert|wurde eingeführt|wurde entfernt|wurde ergänzt|wurde hinzugefügt|wurde ersetzt|wurde umbenannt|wurde behoben|wurde korrigiert|wurde verbessert|wurden geändert"),
		z("B14", `(damals|ehemals|vormals|inzwischen|mittlerweile)`, "damals|ehemals|vormals|inzwischen|mittlerweile"),
		z("C1", `roadmap\p{L}*`, "Roadmap|Roadmaps"),
		z("C2", `welle\p{L}*`, "Welle|Wellen|welle-benutzerhandbuch"),
		z("C3", `slice\p{L}*`, "Slice|Slices|slice-handbuch-geruest"),
		z("C4", `adrs?`, "ADR|ADRs|adr-0001|adr-idx-0004"),
		z("C5", `lh-[a-z]{2}(-[a-z]+)?-\d+`, "lh-fa-01|lh-qa-sup-002|lh-rb-03"),
		z("C6", `spec-\d{3}`, "spec-110"),
		z("C7", `arc-\d{3}`, "arc-004"),
		z("C8", `mr-\d{3}`, "mr-001"),
		z("C9", `co-\d{3}`, "co-001"),
		z("C10", `rc-\d{3}`, "rc-001"),
		z("C11", `beo-[a-z]\p{L}*`, "beo-repo/abc"),
		z("C12", `(docs/plan/)\p{L}*`, "docs/plan/adr/"),
		z("C13", `(spec/)\p{L}*`, "spec/lastenheft.md"),
		z("C14", `(harness/vertraege)\p{L}*`, "harness/vertraege/"),
	}
}

func mengeName(id string) string {
	switch id[0] {
	case 'A':
		return "Planung"
	case 'B':
		return "Chronik"
	}
	return "Kennung"
}

// trefferZeile ist der Teil der Meldung, der Menge und Muster nennt.
func trefferZeile(id string) string {
	return ": " + mengeName(id) + " " + id + ": "
}

// kern ist der Teil einer Form, den das Muster trifft.
func kern(form string) string {
	return strings.TrimPrefix(strings.TrimPrefix(form, "in "), "wie ")
}

// stellungen setzt eine Form an die sieben Stellen eines sauberen Handbuchs.
func stellungen(form string) map[string]string {
	h := sauber()
	return map[string]string{
		"Satzanfang":    h + "\n" + form + " ist so.\n",
		"Satzende":      h + "\nDas ist " + form + ".\n",
		"Überschrift":   h + "\n## " + form + "\n",
		"Listenpunkt":   h + "\n- " + form + "\n",
		"Tabellenzelle": h + "\n| " + form + " | x |\n",
		"Linktext":      h + "\n[" + form + "](x.md)\n",
		"Kopf":          strings.Replace(h, "Gültigkeitsbereich: Binary", "Gültigkeitsbereich: "+form, 1),
	}
}
