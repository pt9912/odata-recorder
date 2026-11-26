package main

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// Die neun Fälle 217 bis 225 tragen die Namen aus SPEC-102, jeder genau einmal und in dieser Folge, mit getrennten Strömen.
func TestLasttestFaelleHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{
		"variablen", "variablen-ziffern", "variablen-erste-meldung", "docker-run", "argument", "docker-info-scheitert", "bau-scheitert",
		"exit-des-containers", "abbruch",
	}
	var got []string
	for _, f := range lasttestFaelle() {
		got = append(got, strings.TrimPrefix(f.Name, "lasttest/"))
		if f.Gate != gateLasttest || !f.Getrennt || !f.Wegwerfmodul {
			t.Errorf("%s: Gate %d, getrennt %v, Wegwerfverzeichnis %v", f.Name, f.Gate, f.Getrennt, f.Wegwerfmodul)
		}
	}
	if !slices.Equal(got, want) || len(got) != 9 {
		t.Errorf("Fälle %v, erwartet %v", got, want)
	}
}

// Die Menge der Formen hat die Größen des Vertrags: 31 Formen der Variablen (7, 5, 11, 8) und 139 Elemente Ziffer mal Stelle; 14 verbotene
// Optionen, fünf Pflichtoptionen und neun Paare aus Signal und Ort.
func TestLasttestMengen(t *testing.T) {
	je := map[string]int{}
	for _, f := range llFormen() {
		je[f.name]++
	}
	if len(llFormen()) != 31 || je["LASTTEST_BETRIEBSART"] != 7 || je["LASTTEST_PROFIL"] != 5 || je["LASTTEST_ZIEL_MS"] != 11 ||
		je["LASTTEST_STARTWERT"] != 8 {
		t.Errorf("Formen %v", je)
	}
	ungueltig := map[string]int{}
	for _, f := range llZifferFormen() {
		if f.meldung != "" {
			ungueltig[f.name]++
		}
	}
	// Zielwert: an der ersten Stelle jeder der vier Längen sechs ungültige Zeichen (0, a, -, +, Leerraum, .), an den sechs weiteren Stellen je
	// fünf, dazu die Längen 5 und 6. Startwert: 0 allein an Stelle 1, a an jeder der 18 Stellen und die Länge 19.
	if len(llZifferFormen()) != 225 || ungueltig["LASTTEST_ZIEL_MS"] != 4*6+6*5+2 || ungueltig["LASTTEST_STARTWERT"] != 1+18+1 {
		t.Errorf("%d Elemente, ungültig %v", len(llZifferFormen()), ungueltig)
	}
	if len(llVerboten()) != 14 || len(llPflicht()) != 5 || len(llSignale()) != 9 || len(llExits()) != 12 || len(llEnvWerte(nil, "1")) != 5 {
		t.Error("Mengen des Vertrags")
	}
}

// llGueltig ist der Aufruf von docker run, den der Vertrag verlangt.
func llGueltig() []string {
	args := []string{"run", "--rm", "--network", "none", "--cpus", "2", "--memory", "4g", "--memory-swap", "4g"}
	for _, e := range llEnvWerte(nil, "16") {
		args = append(args, "--env", e)
	}
	return append(args, llImage, llProgramm)
}

// Der Leser der Argumente von docker run nennt keinen Befund für den verlangten Aufruf und für jedes der 29 Elemente von SPEC-216 Nr. 3 b
// und c einen: fünf Pflichtoptionen und fünf --env je entfernt, 14 verbotene Optionen und vier Optionen mit falschem Wert hinzugefügt oder
// geändert, und LASTTEST_KERNE aus der Umgebung statt aus docker info.
func TestLlPruefeRunFaengtJedesElement(t *testing.T) {
	soll := llEnvWerte(nil, "16")
	if b := llPruefeRun(llGueltig(), soll); len(b) != 0 {
		t.Fatalf("der verlangte Aufruf: %v", b)
	}
	if b := llPruefeRun([]string{"run", "--rm", "--network=none", "--cpus=2", "--memory=4g", "--memory-swap=4g", "--env=" + soll[0],
		"--env=" + soll[1], "--env=" + soll[2], "--env=" + soll[3], "--env=" + soll[4], llImage, llProgramm}, soll); len(b) != 0 {
		t.Fatalf("der Aufruf mit =: %v", b)
	}
	n := 0
	pruefe := func(was string, args []string) {
		n++
		if len(llPruefeRun(args, soll)) == 0 {
			t.Errorf("%s: kein Befund für %v", was, args)
		}
	}
	ohne := func(i, laenge int) []string {
		return slices.Concat(llGueltig()[:i], llGueltig()[i+laenge:])
	}
	pruefe("ohne --rm", ohne(1, 1))
	for _, opt := range []string{"--network", "--cpus", "--memory", "--memory-swap"} {
		i := slices.Index(llGueltig(), opt)
		pruefe("ohne "+opt, ohne(i, 2))
	}
	for _, e := range soll {
		i := slices.Index(llGueltig(), e)
		pruefe("ohne --env "+e, ohne(i-1, 2))
	}
	davor := func(zusatz ...string) []string {
		return slices.Concat(llGueltig()[:1], zusatz, llGueltig()[1:])
	}
	for _, v := range llVerboten() {
		pruefe("mit "+v, davor(v))
	}
	for opt, falsch := range map[string]string{"--network": "host", "--cpus": "4", "--memory": "8g", "--memory-swap": "8g"} {
		args := llGueltig()
		args[slices.Index(args, opt)+1] = falsch
		pruefe(opt+" "+falsch, args)
	}
	args := llGueltig()
	args[slices.Index(args, "LASTTEST_KERNE=16")] = "LASTTEST_KERNE=99"
	pruefe("LASTTEST_KERNE aus der Umgebung", args)
	if n != 29 {
		t.Errorf("%d Elemente, erwartet 29", n)
	}
	for was, args := range map[string][]string{
		"doppelt --cpus": davor("--cpus", "2"), "zusätzliches --env": davor("--env", "X=1"), "anderes Image": append(llGueltig()[:len(llGueltig())-2], "x", llProgramm),
		"anderes Programm": append(llGueltig()[:len(llGueltig())-1], "/bin/sh"), "--rm=false": slices.Concat([]string{"run", "--rm=false"}, llGueltig()[2:]),
		"kein run": append([]string{"create"}, llGueltig()[1:]...),
	} {
		if len(llPruefeRun(args, soll)) == 0 {
			t.Errorf("%s: kein Befund", was)
		}
	}
}

// Die neun Fälle gegen den Treiber, soweit er neben dem Modul liegt (im Arbeitsbaum; die Stufe gegenproben fährt sie über das Programm).
func TestLasttestFaelleLaufenGegenDenTreiber(t *testing.T) {
	treiber := absolut("../harness/lasttest.sh")
	if _, err := os.Stat(treiber); err != nil {
		t.Skip("tools/harness/lasttest.sh liegt nicht neben dem Modul")
	}
	u := umgebungMit(func(string, Gate) string { return "exit 99" }, "sh")
	u.Befehl = befehlMit("", "", treiber)
	u.LasttestTreiber = treiber
	ergebnisse, err := fahre(u, lasttestFaelle(), t.TempDir())
	if err != nil || len(ergebnisse) != 9 {
		t.Fatalf("%d Ergebnisse, %v", len(ergebnisse), err)
	}
	for _, e := range ergebnisse {
		if e.Rot != "" {
			t.Errorf("%s: %s", e.Name, e.Rot)
		}
	}
}

// standard nennt den Treiber des Lasttests mit absolutem Pfad.
func TestStandardNenntDenLasttestTreiber(t *testing.T) {
	if got := standard().LasttestTreiber; !strings.HasSuffix(got, "/tools/harness/lasttest.sh") || got[0] != '/' {
		t.Errorf("LasttestTreiber %q", got)
	}
	name, args, wd := standard().Befehl("/k", Fall{Gate: gateLasttest, Argumente: []string{""}})
	if name != "sh" || len(args) != 2 || args[0] != standard().LasttestTreiber || args[1] != "" || wd != "/k" {
		t.Errorf("%q %q %q", name, args, wd)
	}
}
