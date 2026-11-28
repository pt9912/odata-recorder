package main

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// Die elf Fälle 226 bis 236 tragen die Namen aus SPEC-102, jeder genau einmal und in dieser Folge, mit getrennten Strömen.
func TestTestImageFaelleHabenDieNamenAusDemVertrag(t *testing.T) {
	want := []string{
		"gut", "zwei-laeufe", "aufruf", "image", "phasen", "stopp", "token", "exit-codes", "aufraeumen", "abbruch", "docker-variable",
	}
	var got []string
	for _, f := range testImageFaelle() {
		got = append(got, strings.TrimPrefix(f.Name, "testimage/"))
		if f.Gate != gateTestImage || !f.Getrennt || !f.Wegwerfmodul {
			t.Errorf("%s: Gate %d, getrennt %v, Wegwerfverzeichnis %v", f.Name, f.Gate, f.Getrennt, f.Wegwerfmodul)
		}
	}
	if !slices.Equal(got, want) || len(got) != 11 {
		t.Errorf("Fälle %v, erwartet %v", got, want)
	}
}

// Die Mengen haben die Größen des Vertrags: 15 verbotene Optionen, sechs Rollen, drei Tags, 29 Abweichungen des Image, sieben Formen je
// Phase, 14 Aufrufe von docker, neun Orte eines Signals, 14 Läufe mit Rest und fünf der Token-Abholung; ein grüner Lauf hat 32 Aufrufe.
func TestTestImageMengen(t *testing.T) {
	mengen := []struct {
		name      string
		ist, soll int
	}{
		{"verbotene", len(tiVerbotene()), 15}, {"rollen", len(tiRollen()), 6}, {"arten", len(tiArten()), 3},
		{"bilder", len(tiImageBilder()), 29}, {"formen", len(tiPhasenFormen()), 7}, {"orte", len(tiOrte()), 14},
		{"signalorte", len(tiSignalOrte()), 9}, {"aufrufe", len(tiSoll("1-2")), 32}, {"reste", len(tiReste()), 14},
		{"token", len(tiTokenLaeufe()), 5},
	}
	for _, m := range mengen {
		if m.ist != m.soll {
			t.Errorf("%s: %d Elemente, erwartet %d", m.name, m.ist, m.soll)
		}
	}
}

// Der Leser der Optionen nennt keinen Befund für die Aufrufe des Vertrags und für jede der 18 Formen von SPEC-323 Nr. 2 e einen: 15
// verbotene Optionen, auch in der Form mit `=`, und --network mit einem anderen Wert (bridge, host, container eines anderen Laufs).
func TestTiPruefeOptionenFaengtJedeVerboteneForm(t *testing.T) {
	server := tiContainer("server", "1-2")
	var gueltig [][]string
	for _, a := range tiSoll("1-2") {
		if a[0] == "create" || a[0] == "run" {
			gueltig = append(gueltig, a)
			if b := tiPruefeOptionen(a, server); len(b) != 0 {
				t.Errorf("Aufruf des Vertrags %q: %v", a, b)
			}
		}
	}
	if len(gueltig) != 6 {
		t.Fatalf("%d create und run statt 6", len(gueltig))
	}
	var formen [][]string
	for _, o := range tiVerbotene() {
		formen = append(formen, []string{o, "x"})
	}
	for _, netz := range []string{"bridge", "host", "container:" + tiContainer("server", "3-4")} {
		formen = append(formen, []string{"--network", netz})
	}
	if len(formen) != 18 {
		t.Fatalf("%d Formen statt 18", len(formen))
	}
	for _, a := range gueltig {
		for _, f := range formen {
			for _, form := range [][]string{f, {f[0] + "=" + f[1]}} {
				args := slices.Concat(a[:1], form, a[1:])
				if len(tiPruefeOptionen(args, server)) == 0 {
					t.Errorf("kein Befund für %q", args)
				}
			}
		}
	}
}

// Die elf Fälle gegen den Treiber, soweit er neben dem Modul liegt (im Arbeitsbaum; die Stufe gegenproben fährt sie über das Programm).
func TestTestImageFaelleLaufenGegenDenTreiber(t *testing.T) {
	treiber := absolut("../harness/test-image.sh")
	if _, err := os.Stat(treiber); err != nil {
		t.Skip("tools/harness/test-image.sh liegt nicht neben dem Modul")
	}
	u := umgebungMit(func(string, Gate) string { return "exit 99" }, "sh")
	u.Befehl = befehlTestImage(u.Befehl, treiber)
	u.TestImageTreiber = treiber
	ergebnisse, err := fahre(u, testImageFaelle(), t.TempDir())
	if err != nil || len(ergebnisse) != 11 {
		t.Fatalf("%d Ergebnisse, %v", len(ergebnisse), err)
	}
	for _, e := range ergebnisse {
		if e.Rot != "" {
			t.Errorf("%s: %s", e.Name, e.Rot)
		}
	}
}

// standard nennt den Treiber des Neustarttests mit absolutem Pfad und ruft ihn im Verzeichnis des Falls; die übrigen Gates bleiben.
func TestStandardNenntDenTestImageTreiber(t *testing.T) {
	if got := standard().TestImageTreiber; !strings.HasSuffix(got, "/tools/harness/test-image.sh") || got[0] != '/' {
		t.Errorf("TestImageTreiber %q", got)
	}
	name, args, wd := standard().Befehl("/k", Fall{Gate: gateTestImage, Argumente: []string{""}})
	if name != "sh" || len(args) != 2 || args[0] != standard().TestImageTreiber || args[1] != "" || wd != "/k" {
		t.Errorf("%q %q %q", name, args, wd)
	}
	if name, args, _ := standard().Befehl("/k", Fall{Gate: gateLasttest}); name != "sh" || args[0] != standard().LasttestTreiber {
		t.Errorf("Lasttest: %q %q", name, args)
	}
}
