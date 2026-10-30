package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// Tests der Hülle (SPEC-102 Randform 13), der Ersatz-Skripte (Randform 12) und der Mengen M11 bis
// M14 zu den Fällen 110, 112 und 114.

// Die Hülle liefert stdout und stderr getrennt oder gemeinsam; Exit-Code und Frist sind in beiden
// Formen dieselben.
func TestHuelleLiefertStroemeGetrennt(t *testing.T) {
	const skript = "echo O; echo E >&2; exit 3"
	code, out, errOut, abgelaufen := starteStroeme(5*time.Second, "", nil, true, "sh", "-c", skript)
	if code != 3 || out != "O\n" || errOut != "E\n" || abgelaufen {
		t.Errorf("getrennt: Exit %d, stdout %q, stderr %q, abgelaufen %v", code, out, errOut, abgelaufen)
	}
	code, out, errOut, abgelaufen = starteStroeme(5*time.Second, "", nil, false, "sh", "-c", skript)
	if code != 3 || out != "O\nE\n" || errOut != "" || abgelaufen {
		t.Errorf("gemeinsam: Exit %d, stdout %q, stderr %q, abgelaufen %v", code, out, errOut, abgelaufen)
	}
	code, out, abgelaufen = starte(5*time.Second, "", nil, "sh", "-c", skript)
	if code != 3 || out != "O\nE\n" || abgelaufen {
		t.Errorf("starte: Exit %d, Ausgabe %q, abgelaufen %v", code, out, abgelaufen)
	}
	for _, getrennt := range []bool{false, true} {
		start := time.Now()
		_, _, _, abgelaufen := starteStroeme(200*time.Millisecond, "", nil, getrennt, "sh", "-c", "echo E >&2; sleep 30 & wait")
		if !abgelaufen || time.Since(start) > 3*time.Second {
			t.Errorf("getrennt=%v: Frist nicht erkannt oder zu spät (%v)", getrennt, time.Since(start))
		}
	}
}

// Ein Text einer Erwartung steht im genannten Strom und nicht im anderen; ein Fall mit getrennten
// Strömen nennt keinen Text ohne Strom, ein Fall mit gemeinsamer Ausgabe keinen Strom.
func TestBewerteStroeme(t *testing.T) {
	getrennt := func(f Fall) Fall { f.Getrennt = true; return f }
	for _, c := range []struct {
		name     string
		f        Fall
		out, err string
		rot      string // Teiltext des Grundes; leer: grün
	}{
		{"stdout im richtigen Strom", getrennt(Fall{WantOut: []string{"a"}}), "a\n", "b\n", ""},
		{"stderr im richtigen Strom", getrennt(Fall{WantErr: []string{"b"}}), "a\n", "b\n", ""},
		{"stdout nur auf stderr", getrennt(Fall{WantOut: []string{"b"}}), "a\n", "b\n", `stdout nennt "b" nicht`},
		{"stderr nur auf stdout", getrennt(Fall{WantErr: []string{"a"}}), "a\n", "b\n", `stderr nennt "a" nicht`},
		{"stdout auch auf stderr", getrennt(Fall{WantOut: []string{"a"}}), "a\n", "a\n", "steht auch im anderen Strom"},
		{"stderr auch auf stdout", getrennt(Fall{WantErr: []string{"b"}}), "b\n", "b\n", "steht auch im anderen Strom"},
		{"verboten auf stdout", getrennt(Fall{Verboten: []string{"x"}}), "x\n", "", `Ausgabe nennt "x"`},
		{"verboten auf stderr", getrennt(Fall{Verboten: []string{"x"}}), "", "x\n", `Ausgabe nennt "x"`},
		{"Text ohne Strom bei getrennter Hülle", getrennt(Fall{WantText: []string{"a"}}), "a\n", "", "ohne Strom"},
		{"Strom bei gemeinsamer Hülle", Fall{WantOut: []string{"a"}}, "a\n", "", "gemeinsam"},
		{"Text bei gemeinsamer Hülle", Fall{WantText: []string{"a"}}, "a\n", "", ""},
	} {
		got := bewerte(c.f, 0, c.out, c.err, false)
		if c.rot == "" && got != "" || !strings.Contains(got, c.rot) {
			t.Errorf("%s: %q statt %q", c.name, got, c.rot)
		}
	}
}

// fuehreFall startet das Gate eines Falls mit der Hülle, die der Fall nennt.
func TestFuehreFallStartetMitDerHuelleDesFalls(t *testing.T) {
	u := umgebungMit(func(string, Gate) string { return "echo O; echo E >&2" }, "sh")
	lauf := func(f Fall) (int, string) {
		f.Name, f.Wegwerfmodul = "f", true
		var out, errb bytes.Buffer
		return run(u, &out, &errb, wurzelMitEingaben(t), []Fall{f}), out.String()
	}
	for _, c := range []struct {
		name string
		f    Fall
		code int
	}{
		{"gemeinsam, beide Texte", Fall{WantText: []string{"O", "E"}}, exitOK},
		{"getrennt, stdout und stderr richtig", Fall{Getrennt: true, WantOut: []string{"O"}, WantErr: []string{"E"}}, exitOK},
		{"getrennt, stdout falsch", Fall{Getrennt: true, WantOut: []string{"E"}}, exitRot},
		{"getrennt, stderr falsch", Fall{Getrennt: true, WantErr: []string{"O"}}, exitRot},
	} {
		if code, out := lauf(c.f); code != c.code {
			t.Errorf("%s: Exit %d statt %d\n%s", c.name, code, c.code, out)
		}
	}
	// Nachher sieht beide Ströme eines Falls mit getrennter Hülle.
	var gesehen Lauf
	f := Fall{Getrennt: true, Nachher: func(l Lauf) string { gesehen = l; return "" }}
	if code, _ := lauf(f); code != exitOK || gesehen.Out != "O\n" || gesehen.Err != "E\n" {
		t.Errorf("Nachher sah stdout %q und stderr %q", gesehen.Out, gesehen.Err)
	}
}

// Die Fälle 38 bis 120 lesen die Ströme getrennt, die Fälle 1 bis 37 gemeinsam (Randform 13), und
// kein getrennter Fall nennt einen Text ohne Strom.
func TestStroemeJeNachFall(t *testing.T) {
	alt, neu := 0, 0
	for _, f := range faelle() {
		gm := strings.HasPrefix(f.Name, "groessenmarke/")
		if f.Getrennt != gm {
			t.Errorf("%s: Getrennt %v", f.Name, f.Getrennt)
		}
		if gm {
			neu++
			if len(f.WantText) > 0 {
				t.Errorf("%s: Text ohne Strom: %v", f.Name, f.WantText)
			}
		} else {
			alt++
			if len(f.WantOut)+len(f.WantErr) > 0 {
				t.Errorf("%s: nennt einen Strom", f.Name)
			}
		}
	}
	if alt != 37 || neu != 83 {
		t.Errorf("%d Fälle mit gemeinsamer und %d mit getrennter Hülle statt 37 und 83", alt, neu)
	}
}

func fallNamens(t *testing.T, name string) Fall {
	t.Helper()
	for _, f := range groessenmarkeFaelle() {
		if f.Name == "groessenmarke/"+name {
			return f
		}
	}
	t.Fatalf("Fall %s fehlt", name)
	return Fall{}
}

// M11 (12 Elemente), M12 (263), M13 (9) und M14 (2 ohne Plan, 17 mit Plan) zählen ihre Elemente.
func TestMengenDerTreiberFaelle(t *testing.T) {
	// M11: sieben Meldungen auf stderr, eine auf stdout, das Durchreichen der Ausgabe des Programms.
	m11 := gmM11()
	if len(m11) != 12 {
		t.Errorf("M11: %d Elemente statt 12", len(m11))
	}
	statisch := 0
	for _, e := range m11 {
		if !e.statisch {
			continue
		}
		statisch++
		f := fallNamens(t, e.fall)
		var liste []string
		switch e.strom {
		case "stdout":
			liste = f.WantOut
		case "stderr":
			liste = f.WantErr
		default:
			liste = f.Verboten
		}
		if !slices.Contains(liste, e.text) {
			t.Errorf("M11: %s nennt %q nicht im Strom %s", e.fall, e.text, e.strom)
		}
	}
	if statisch != 10 {
		t.Errorf("M11: %d statische Elemente statt 10", statisch)
	}

	// M12: 256 Exits von pruefen, sechs von abfrage, abfrage ohne Ergebnis.
	zaehle := map[string]int{}
	exits := map[int]bool{}
	for _, e := range gmM12() {
		zaehle[e.art]++
		if e.art == "pruefen" {
			exits[e.rc] = true
		}
	}
	if len(gmM12()) != 263 || zaehle["pruefen"] != 256 || len(exits) != 256 || zaehle["abfrage"] != 6 || zaehle["abfrage-leer"] != 1 {
		t.Errorf("M12: %d Elemente (%v, %d verschiedene Exits) statt 263", len(gmM12()), zaehle, len(exits))
	}

	// M13: drei Signale an drei Orte, jedes Paar einmal.
	paare := map[gmM13Element]bool{}
	for _, e := range gmM13() {
		paare[e] = true
	}
	if len(gmM13()) != 9 || len(paare) != 9 {
		t.Errorf("M13: %d Elemente, %d verschiedene, statt 9", len(gmM13()), len(paare))
	}

	// M14: zwei Verzeichnisse ohne Plan, 17 Fälle mit mindestens einem Plan, die die Meldung verbieten
	// (103 bis 109, 111 bis 120); Fall 110 nennt sie auf stdout.
	mit := 0
	for _, f := range groessenmarkeFaelle() {
		if f.Gate != gateTreiber {
			continue
		}
		verbietet, nennt := slices.Contains(f.Verboten, gmKeinSlice), slices.Contains(f.WantOut, gmKeinSlice)
		if verbietet == nennt {
			t.Errorf("%s: verbietet %v, nennt %v", f.Name, verbietet, nennt)
		}
		if verbietet {
			mit++
		}
	}
	if len(gmOhnePlan()) != 2 || mit != 17 {
		t.Errorf("M14: %d Verzeichnisse ohne Plan und %d Fälle mit Plan statt 2 und 17", len(gmOhnePlan()), mit)
	}
}

// Die Ersatz-Skripte senden dem Treiber, ihrem Elternprozess, auf Wunsch das Signal; ohne die
// Datei senden sie nichts (SPEC-128 Nr. 7 f, M13).
func TestErsatzSkripteSendenDemTreiberDasSignal(t *testing.T) {
	dir := t.TempDir()
	neu := gmErsatz(nil, gmEintrag{pfad: gmPlanA, abfrage: gmMessen})
	if err := gmSchreibe(dir, neu); err != nil {
		t.Fatal(err)
	}
	messung := filepath.Join(dir, "m")
	if err := os.WriteFile(messung, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	aufrufe := map[string][]string{
		"abfrage": {"programm", "abfrage", gmPlanA}, "git diff": {"git", "diff"}, "pruefen": {"programm", "pruefen", gmPlanA, messung},
	}
	n := 0
	for _, e := range gmM13() {
		n++
		got := filepath.Join(dir, "got")
		_ = os.Remove(got)
		// Der äußere Prozess fängt das Signal und merkt es sich; der Ersatz läuft als sein Kind.
		fang := `trap 'echo GOT >> got' ` + e.signal + `; sh bin/` + strings.Join(aufrufe[e.ort], " ") + ` > /dev/null 2>&1; true`
		for _, mitSignal := range []bool{false, true} {
			_ = os.Remove(got)
			if mitSignal {
				if err := gmSchreibe(dir, map[string]string{e.signalDatei(): e.signal + "\n"}); err != nil {
					t.Fatal(err)
				}
			}
			if code, out, _ := starte(10*time.Second, dir, nil, "sh", "-c", fang); code != 0 {
				t.Fatalf("%s %s: Exit %d\n%s", e.signal, e.ort, code, out)
			}
			b, _ := os.ReadFile(got)
			if want := map[bool]string{false: "", true: "GOT\n"}[mitSignal]; string(b) != want {
				t.Errorf("%s an %s, Signal-Datei %v: gefangen %q statt %q", e.signal, e.ort, mitSignal, b, want)
			}
		}
		if err := os.Remove(filepath.Join(dir, e.signalDatei())); err != nil {
			t.Fatal(err)
		}
	}
	if n != 9 {
		t.Errorf("%d Läufe statt 9", n)
	}
}
