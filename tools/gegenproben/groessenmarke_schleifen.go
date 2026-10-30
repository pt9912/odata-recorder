package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Die Mengen M11 bis M14 des Plans zu den Fällen 110, 112 und 114 (SPEC-102, SPEC-128 Nr. 7): jede
// Menge ist eine Funktion, die ihre Elemente liefert; die Fälle laufen über sie, und
// TestMengenDerTreiberFaelle zählt sie.

const (
	gmKeinSlice            = "kein Slice in Arbeit"
	gmPlanverzeichnisFehlt = "Planverzeichnis nicht-da fehlt"
	gmOhneErgebnis         = "abfrage ohne Ergebnis"
)

// gmStromElement ist ein Element von M11: eine Meldung des Treibers und der Strom, in dem sie steht
// (stdout, stderr oder keiner); statisch heißt, dass der Fall die Erwartung in WantOut, WantErr oder
// Verboten führt, sonst prüft sie sein Nachher.
type gmStromElement struct {
	fall, text, strom string
	statisch          bool
}

// gmM11 sind die zwölf Meldungen des Treibers und Ströme nach SPEC-128 Nr. 7 c und 7 d: sieben auf
// stderr, eine auf stdout, dazu das Durchreichen der Ausgabe des Programms (vier).
func gmM11() []gmStromElement {
	return []gmStromElement{
		{"treiber-keine-plaene", gmPlanverzeichnisFehlt, "stderr", false},
		{"treiber-exit-aggregation", gmOhneErgebnis, "stderr", false},
		{"treiber-git-nicht-aufrufbar", "git nicht aufrufbar", "stderr", true},
		{"treiber-head-nicht-lesbar", "HEAD nicht lesbar", "stderr", true},
		{"treiber-basis-kein-vorfahr", "ist kein Vorfahr von HEAD", "stderr", true},
		{"treiber-basis-unbekannt", "unbekannt oder git scheitert", "stderr", true},
		{"treiber-git-diff-scheitert", "git diff scheitert", "stderr", true},
		{"treiber-keine-plaene", gmKeinSlice, "stdout", true},
		{"treiber-aufrufe-von-git", gmPruefenOut, "stdout", true},
		{"treiber-aufrufe-von-git", gmPruefenErr, "stderr", true},
		{"treiber-aufrufe-von-git", gmAbfrageErr, "stderr", true},
		{"treiber-aufrufe-von-git", gmAbfrageOut, "keiner", true},
	}
}

// gmOhnePlan sind die Planverzeichnisse ohne Datei slice-*.md (M14): leer, und eines mit Dateien, die
// keine Pläne sind. Das erste fährt der Fall selbst, das zweite sein Nachher.
func gmOhnePlan() []string {
	return []string{"leer", "nur-andere"}
}

// gmNurAndere sind die Dateien des Verzeichnisses ohne Plan: roadmap.md, README.md, slice-c.txt und
// ein Unterverzeichnis mit slice-d.md.
func gmNurAndere() map[string]string {
	return map[string]string{
		"nur-andere/roadmap.md": "x\n", "nur-andere/README.md": "x\n", "nur-andere/slice-c.txt": "x\n", "nur-andere/unter/slice-d.md": "x\n",
	}
}

// gmKeinePlaene ist das Nachher von Fall 110: das Verzeichnis mit Nicht-Plänen meldet `kein Slice in
// Arbeit` auf stdout, ein fehlendes Verzeichnis Exit 2 mit der Meldung auf stderr; der Treiber ruft
// weder git noch das Programm.
func gmKeinePlaene(l Lauf) string {
	for _, d := range gmOhnePlan()[1:] {
		code, out, errOut := l.treiberStroeme(gmProgramm(l), d)
		if code != 0 || !strings.Contains(out, gmKeinSlice) || strings.Contains(errOut, gmKeinSlice) {
			return fmt.Sprintf("Verzeichnis %s ohne Plan: Exit %d\nstdout: %s\nstderr: %s", d, code, out, errOut)
		}
	}
	code, out, errOut := l.treiberStroeme(gmProgramm(l), "nicht-da")
	if code != 2 || !strings.Contains(errOut, gmPlanverzeichnisFehlt) || strings.Contains(out, gmPlanverzeichnisFehlt) || strings.Contains(out+errOut, gmKeinSlice) {
		return fmt.Sprintf("fehlendes Planverzeichnis: Exit %d\nstdout: %s\nstderr: %s", code, out, errOut)
	}
	return gmErste(gmErwartet(l, "git.log", nil), gmErwartet(l, "programm.log", nil))
}

// gmM12Element ist ein Element von M12: der Exit von `pruefen` (art pruefen), der Exit von `abfrage`
// ungleich 0 (abfrage) oder `abfrage` mit Exit 0 ohne Ausgabe (abfrage-leer).
type gmM12Element struct {
	art string
	rc  int
}

// gmM12 sind die 263 Elemente: Exit 0 bis 255 von `pruefen`, sechs Exits von `abfrage` und
// `abfrage` ohne Ergebnis.
func gmM12() []gmM12Element {
	var es []gmM12Element
	for rc := 0; rc <= 255; rc++ {
		es = append(es, gmM12Element{"pruefen", rc})
	}
	for _, rc := range []int{1, 2, 3, 125, 127, 255} {
		es = append(es, gmM12Element{"abfrage", rc})
	}
	return append(es, gmM12Element{"abfrage-leer", 0})
}

// gmM12Dirs nennt das Planverzeichnis je Art der Elemente von M12.
func gmM12Dirs() map[string]string {
	return map[string]string{"pruefen": "schleife", "abfrage": "abfrage-exit", "abfrage-leer": "abfrage-leer"}
}

// gmExitSchleife fährt M12 (SPEC-128 Nr. 7 c, 7 e): 0 und 1 von `pruefen` bleiben, 2 bis 255 ergeben
// 2; jeder Exit von `abfrage` ungleich 0 ergibt 2 ohne Aufruf von git und pruefen; `abfrage` ohne
// Ergebnis ergibt 2 und die Meldung auf stderr. Das Protokoll belegt, dass jedes Element lief.
func gmExitSchleife(l Lauf) string {
	dateien := map[string]string{}
	for _, d := range gmM12Dirs() {
		dateien["ersatz/"+d+"/slice-a.md.abfrage"] = gmMessen
		dateien[d+"/slice-a.md"] = "x\n"
	}
	dateien["ersatz/abfrage-leer/slice-a.md.abfrage"] = ""
	if err := gmSchreibe(l.Dir, dateien); err != nil {
		return err.Error()
	}
	vorher, err := gmProtokoll(l.Dir, "git.log")
	if err != nil {
		return err.Error()
	}
	for _, e := range gmM12() {
		if r := gmExitElement(l, e); r != "" {
			return r
		}
	}
	return gmExitProtokoll(l, len(vorher))
}

// gmExitElement fährt ein Element von M12: der Exit des Treibers folgt SPEC-128 Nr. 7 c und 7 e.
func gmExitElement(l Lauf, e gmM12Element) string {
	d := gmM12Dirs()[e.art]
	want, rcDatei := 2, ""
	switch e.art {
	case "pruefen":
		want, rcDatei = min(e.rc, 2), "ersatz/"+d+"/slice-a.md.pruefen.rc"
	case "abfrage":
		rcDatei = "ersatz/" + d + "/slice-a.md.abfrage.rc"
	}
	if rcDatei != "" {
		if err := gmSchreibe(l.Dir, map[string]string{rcDatei: fmt.Sprint(e.rc) + "\n"}); err != nil {
			return err.Error()
		}
	}
	code, out, errOut := l.treiberStroeme(gmProgramm(l), d)
	if code != want {
		return fmt.Sprintf("%s mit Exit %d: Exit %d statt %d\n%s%s", e.art, e.rc, code, want, out, errOut)
	}
	if e.art == "abfrage-leer" && (!strings.Contains(errOut, gmOhneErgebnis) || strings.Contains(out, gmOhneErgebnis)) {
		return fmt.Sprintf("abfrage ohne Ergebnis: stdout %q, stderr %q", out, errOut)
	}
	return ""
}

// gmExitProtokoll belegt mit den Protokollen, dass jedes Element von M12 lief: drei Aufrufe von git
// je Lauf von `pruefen`, und je Art die Zahl der Aufrufe von `abfrage` und `pruefen`.
func gmExitProtokoll(l Lauf, gitVorher int) string {
	nachher, err := gmProtokoll(l.Dir, "git.log")
	if err != nil {
		return err.Error()
	}
	if nach, soll := len(nachher)-gitVorher, 3*256; nach != soll {
		return fmt.Sprintf("git %d-mal gerufen statt %d-mal (drei Aufrufe je Lauf von pruefen)", nach, soll)
	}
	aufrufe, err := gmProtokoll(l.Dir, "programm.log")
	if err != nil {
		return err.Error()
	}
	for _, z := range []struct {
		art           string
		abfrage, gepr int
	}{{"pruefen", 256, 256}, {"abfrage", 6, 0}, {"abfrage-leer", 1, 0}} {
		plan := gmM12Dirs()[z.art] + "/slice-a.md"
		if a, g := gmZaehle(aufrufe, "abfrage", plan), gmZaehle(aufrufe, "pruefen", plan); a != z.abfrage || g != z.gepr {
			return fmt.Sprintf("%s: %d-mal abgefragt und %d-mal geprüft statt %d und %d", plan, a, g, z.abfrage, z.gepr)
		}
	}
	return ""
}

// gmM13Element ist ein Element von M13: das Signal und der Ort, an dem der Ersatz es dem Treiber
// sendet.
type gmM13Element struct {
	signal, ort string
}

// gmM13 sind die neun Läufe: HUP, INT und TERM an `abfrage`, `git diff` und `pruefen`.
func gmM13() []gmM13Element {
	var es []gmM13Element
	for _, s := range []string{"HUP", "INT", "TERM"} {
		for _, o := range []string{"abfrage", "git diff", "pruefen"} {
			es = append(es, gmM13Element{s, o})
		}
	}
	return es
}

// signalDatei ist die Datei, die den Ersatz am Ort das Signal senden lässt.
func (e gmM13Element) signalDatei() string {
	if e.ort == "git diff" {
		return "ersatz/git.diff.signal"
	}
	return "ersatz/" + gmPlanA + "." + e.ort + ".signal"
}

// aufrufe zählt die Aufrufe der Operation des Ortes in den Protokollen.
func (e gmM13Element) aufrufe(dir string) (int, error) {
	datei, kommando := "programm.log", e.ort
	if e.ort == "git diff" {
		datei, kommando = "git.log", "diff"
	}
	as, err := gmProtokoll(dir, datei)
	n := 0
	for _, a := range as {
		if len(a) > 0 && a[0] == kommando {
			n++
		}
	}
	return n, err
}

// gmAufgeraeumt prüft Fall 114 (SPEC-128 Nr. 7 f): nach dem Lauf mit Exit 0, nach einem Lauf mit
// Fehler von git diff und nach jedem der neun Läufe aus M13 ist das TMPDIR des Falls leer.
func gmAufgeraeumt(l Lauf) string {
	if r := gmLeer(l.Tmp); r != "" {
		return "nach Exit 0: " + r
	}
	if r := gmFehlerVonGitDiff(l); r != "" {
		return r
	}
	for _, e := range gmM13() {
		if r := gmSignalLauf(l, e); r != "" {
			return e.signal + " an " + e.ort + ": " + r
		}
	}
	return ""
}

// gmFehlerVonGitDiff fährt den Treiber mit einem Fehler von git diff und nimmt den Fehler danach
// zurück, damit die Läufe aus M13 ohne das Signal Exit 0 ergäben.
func gmFehlerVonGitDiff(l Lauf) string {
	fehler := map[string]string{"ersatz/git.diff.rc": "128\n", "ersatz/git.diff.err": "fatal: boese\n"}
	if err := gmSchreibe(l.Dir, fehler); err != nil {
		return err.Error()
	}
	code, out := l.treiber(gmProgramm(l), "plaene")
	if code != 2 || !strings.Contains(out, "git diff scheitert") {
		return fmt.Sprintf("Fehler von git diff: Exit %d\n%s", code, out)
	}
	if r := gmLeer(l.Tmp); r != "" {
		return "nach dem Fehler von git diff: " + r
	}
	for d := range fehler {
		if err := os.Remove(filepath.Join(l.Dir, d)); err != nil {
			return err.Error()
		}
	}
	return ""
}

// gmSignalLauf fährt ein Element von M13: der Ersatz am Ort sendet dem Treiber das Signal, der Treiber
// endet mit 2 und lässt kein Verzeichnis unter TMPDIR zurück. Der Lauf muss den Ort erreicht haben,
// sonst hätte der Treiber vorher aus anderem Grund geendet.
func gmSignalLauf(l Lauf, e gmM13Element) string {
	vorher, err := e.aufrufe(l.Dir)
	if err != nil {
		return err.Error()
	}
	if err := gmSchreibe(l.Dir, map[string]string{e.signalDatei(): e.signal + "\n"}); err != nil {
		return err.Error()
	}
	code, out := l.treiber(gmProgramm(l), "plaene")
	nachher, err := e.aufrufe(l.Dir)
	if err != nil {
		return err.Error()
	}
	if err := os.Remove(filepath.Join(l.Dir, e.signalDatei())); err != nil {
		return err.Error()
	}
	switch {
	case nachher != vorher+1:
		return fmt.Sprintf("%d Aufrufe des Ortes statt 1", nachher-vorher)
	case code != 2:
		return fmt.Sprintf("Exit %d statt 2\n%s", code, out)
	}
	return gmLeer(l.Tmp)
}
