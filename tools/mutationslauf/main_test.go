package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestUmgebung (SPEC-210 Nr. 1 a): MUTATIONSLAUF_GRENZE 1 bis 14400, MUTATIONSLAUF_KALT leer oder 1; jeder andere
// Wert, auch die leere Grenze, ist Exit 2 mit dem Namen der Variable.
func TestUmgebung(t *testing.T) {
	for _, g := range []string{"0", "14401", "abc", "", "-1", "01", "3600s"} {
		d, _, ungueltig := umgebung([]string{"MUTATIONSLAUF_GRENZE=" + g})
		if ungueltig != "MUTATIONSLAUF_GRENZE" || d != 0 {
			t.Errorf("GRENZE=%q: %v %q", g, d, ungueltig)
		}
	}
	for _, g := range []string{"1", "3600", "14400"} {
		if d, _, ungueltig := umgebung([]string{"MUTATIONSLAUF_GRENZE=" + g}); ungueltig != "" || d == 0 {
			t.Errorf("GRENZE=%q: %v %q", g, d, ungueltig)
		}
	}
	for _, k := range []string{"0", "2", "ja", "true"} {
		if _, _, ungueltig := umgebung([]string{"MUTATIONSLAUF_KALT=" + k}); ungueltig != "MUTATIONSLAUF_KALT" {
			t.Errorf("KALT=%q: %q", k, ungueltig)
		}
	}
	if d, kalt, ungueltig := umgebung(nil); d != time.Hour || kalt || ungueltig != "" {
		t.Errorf("Vorgaben: %v %v %q", d, kalt, ungueltig)
	}
	if _, kalt, _ := umgebung([]string{"MUTATIONSLAUF_KALT=1"}); !kalt {
		t.Error("KALT=1 schaltet nicht um")
	}
	if _, kalt, ungueltig := umgebung([]string{"MUTATIONSLAUF_KALT="}); kalt || ungueltig != "" {
		t.Error("leeres KALT ist warm")
	}
}

// TestRunEingabe (SPEC-210 Nr. 1 a, 3 a): Aufruf, fehlende Liste, ungültige Umgebung und veraltetes Image sind Exit 2.
func TestRunEingabe(t *testing.T) {
	baum, ablage, image := fixtureBaum(t), t.TempDir(), t.TempDir()
	liste := filepath.Join(t.TempDir(), "liste.json")
	if err := os.WriteFile(liste, listeJSON(t, mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere)), 0o644); err != nil {
		t.Fatal(err)
	}
	faelle := []struct {
		name string
		args []string
		env  []string
		want string
	}{
		{"ohne Argumente", nil, nil, "Aufruf: mutationslauf"},
		{"zu viele", []string{liste, baum, ablage, "", "x"}, nil, "Aufruf: mutationslauf"},
		{"Liste fehlt", []string{filepath.Join(ablage, "fehlt"), baum}, nil, "Liste fehlt"},
		{"Umgebung", []string{liste, baum, ablage, ""}, []string{"MUTATIONSLAUF_GRENZE=0"}, "Umgebung ungültig: MUTATIONSLAUF_GRENZE"},
		{"Image veraltet", []string{liste, baum, ablage, image}, nil, "Image veraltet: go.mod oder go.sum weichen ab"},
	}
	if err := os.WriteFile(filepath.Join(image, "go.mod"), []byte("module anders\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, f := range faelle {
		var aus, fehl bytes.Buffer
		code := run(t.Context(), f.args, f.env, &aus, &fehl)
		if code != 2 || aus.Len() != 0 || !strings.Contains(fehl.String(), f.want) {
			t.Errorf("%s: Exit %d, stdout %q, stderr %q", f.name, code, aus.String(), fehl.String())
		}
	}
}

// TestImageAbgleich (SPEC-210 Nr. 3 a): go.mod und go.sum müssen gleich sein, ein fehlendes go.sum auf beiden Seiten
// ist gleich.
func TestImageAbgleich(t *testing.T) {
	faelle := []struct {
		name        string
		baum, image map[string]string
		gleich      bool
	}{
		{"gleich", map[string]string{"go.mod": "m", "go.sum": "s"}, map[string]string{"go.mod": "m", "go.sum": "s"}, true},
		{"ohne go.sum", map[string]string{"go.mod": "m"}, map[string]string{"go.mod": "m"}, true},
		{"go.mod anders", map[string]string{"go.mod": "m"}, map[string]string{"go.mod": "n"}, false},
		{"go.sum anders", map[string]string{"go.mod": "m", "go.sum": "s"}, map[string]string{"go.mod": "m", "go.sum": "t"}, false},
		{"go.sum nur im Baum", map[string]string{"go.mod": "m", "go.sum": ""}, map[string]string{"go.mod": "m"}, false},
		{"go.sum nur im Image", map[string]string{"go.mod": "m"}, map[string]string{"go.mod": "m", "go.sum": ""}, false},
	}
	for _, f := range faelle {
		b, i := t.TempDir(), t.TempDir()
		for d, m := range map[string]map[string]string{b: f.baum, i: f.image} {
			for n, inhalt := range m {
				schreibe(t, d, n, inhalt)
			}
		}
		if got := imageAbgleich(b, i); got != f.gleich {
			t.Errorf("%s: %v statt %v", f.name, got, f.gleich)
		}
	}
}

// TestKaltSetztCache (SPEC-210 Nr. 8 c): MUTATIONSLAUF_KALT=1 legt den Build-Cache unter der Ablage an und lässt den
// vorgegebenen unberührt; ohne den Wert nutzt der Lauf den vorgegebenen.
func TestKaltSetztCache(t *testing.T) {
	baum := fixtureBaum(t)
	liste := filepath.Join(t.TempDir(), "liste.json")
	if err := os.WriteFile(liste, listeJSON(t, mut("a", "src/p/a.go", "return a + b", "return b + a", "TestAdd", klasseAndere)), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, kalt := range []bool{true, false} {
		ablage, vorgabe := t.TempDir(), t.TempDir()
		env := append(os.Environ(), "GOCACHE="+vorgabe)
		if kalt {
			env = append(env, "MUTATIONSLAUF_KALT=1")
		}
		var aus, fehl bytes.Buffer
		code := run(t.Context(), []string{liste, baum, ablage, ""}, env, &aus, &fehl)
		eigener, _ := os.ReadDir(filepath.Join(ablage, "gocache"))
		benutzt, _ := os.ReadDir(vorgabe)
		if code != 0 || (len(eigener) > 0) != kalt || (len(benutzt) > 0) == kalt {
			t.Errorf("kalt=%v: Exit %d, Cache der Ablage %d Einträge, vorgegebener %d; %s", kalt, code, len(eigener), len(benutzt), fehl.String())
		}
	}
}

// TestExitCodes (SPEC-210 Nr. 1 d): rot, rot-paket, gruen und flatternd sind Exit 0, uebersetzungsfehler und abbruch Exit 1.
func TestExitCodes(t *testing.T) {
	baum := fixtureBaum(t)
	m := func(klasse string) ml {
		return mitRate(mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasse), 2, 2)
	}
	rotBei := func(f func(Start) bool) func(Start) Ausgang {
		return func(s Start) Ausgang {
			if s.Nullprobe {
				return (&fakeP{}).Starten(context.Background(), s)
			}
			return Ausgang{Code: map[bool]int{true: 1}[f(s)]}
		}
	}
	faelle := []struct {
		erg    string
		klasse string
		fp     *fakeP
		code   int
	}{
		{ergRot, klasseAndere, &fakeP{startAus: rotBei(func(Start) bool { return true })}, 0},
		{ergRotPaket, klasseAndere, &fakeP{startAus: rotBei(func(s Start) bool { return !hatArg(s, "-test.run") })}, 0},
		{ergGruen, klasseAndere, &fakeP{}, 0},
		{ergFlattern, klasseNebenlaeufig, &fakeP{startAus: rotBei(func(s Start) bool { return s.Detektor })}, 0},
		{ergUebersetz, klasseAndere, &fakeP{bauAus: func(b Bau) Ausgang { return Ausgang{Code: map[bool]int{false: 1}[istNull(b)]} }}, 1},
		{ergAbbruch, klasseAndere, &fakeP{startAus: func(s Start) Ausgang {
			if s.Nullprobe {
				return (&fakeP{}).Starten(context.Background(), s)
			}
			return Ausgang{Abgebrochen: true}
		}}, 1},
	}
	for _, f := range faelle {
		w, aus, _ := fakeWerk(t, baum, f.fp)
		code := w.Lauf(context.Background(), muts(t, baum, m(f.klasse)))
		if got := ergebnisse(aus.String())["M"]; got != f.erg || code != f.code {
			t.Errorf("%s: Ergebnis %s, Exit %d statt %d", f.erg, got, code, f.code)
		}
	}
}
