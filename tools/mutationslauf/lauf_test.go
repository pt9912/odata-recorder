package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func istNull(b Bau) bool { return strings.Contains(filepath.Base(b.Dir), "null") }

func startNull(s Start) bool { return s.Nullprobe }

// TestErgebniszeile (SPEC-210 Nr. 1 b): Zeile und Zusammenfassung gleichen dem Format von SPEC-210 Nr. 1 b; `prozesse` gleicht
// den Starts des Fakes ohne die Nullprobe, `last` dem Wert zu Beginn der Mutation.
func TestErgebniszeile(t *testing.T) {
	baum := fixtureBaum(t)
	rot := func(s Start) bool { return !s.Nullprobe && hatArg(s, "-test.run") }
	stufe2 := func(s Start) bool { return !s.Nullprobe && !hatArg(s, "-test.run") }
	faelle := []struct {
		klasse, erg     string
		bau             func(Bau) Ausgang
		start           func(Start) Ausgang
		mit, ohne       string
		prozesse        int
		keinStart       bool
		erwarteteStarts int
	}{
		{klasseAndere, ergUebersetz, func(b Bau) Ausgang { return Ausgang{Code: map[bool]int{false: 1}[istNull(b)]} }, nil, "-", "-", 0, true, 0},
		{klasseAndere, ergAbbruch, nil, func(s Start) Ausgang { return Ausgang{Abgebrochen: rot(s)} }, "-", "-", 1, false, 1},
		{klasseAndere, ergRot, nil, func(s Start) Ausgang { return Ausgang{Code: map[bool]int{true: 1}[rot(s)]} }, "-", "-", 1, false, 1},
		{klasseAndere, ergRotPaket, nil, func(s Start) Ausgang { return Ausgang{Code: map[bool]int{true: 1}[stufe2(s)]} }, "-", "-", 2, false, 2},
		{klasseAndere, ergGruen, nil, nil, "-", "-", 2, false, 2},
		{klasseNebenlaeufig, ergUebersetz, func(b Bau) Ausgang { return Ausgang{Code: map[bool]int{false: 1}[istNull(b)]} }, nil, "0/0", "0/0", 0, true, 0},
		{klasseNebenlaeufig, ergAbbruch, nil, func(s Start) Ausgang { return Ausgang{Abgebrochen: rot(s)} }, "0/3", "0/2", 5, false, 5},
		{klasseNebenlaeufig, ergRot, nil, func(s Start) Ausgang { return Ausgang{Code: map[bool]int{true: 1}[rot(s)]} }, "3/3", "2/2", 5, false, 5},
		{klasseNebenlaeufig, ergFlattern, nil, func(s Start) Ausgang { return Ausgang{Code: map[bool]int{true: 1}[rot(s) && s.Detektor]} }, "3/3", "0/2", 5, false, 5},
		{klasseNebenlaeufig, ergRotPaket, nil, func(s Start) Ausgang { return Ausgang{Code: map[bool]int{true: 1}[stufe2(s)]} }, "0/3", "0/2", 6, false, 6},
		{klasseNebenlaeufig, ergGruen, nil, nil, "0/3", "0/2", 6, false, 6},
	}
	format := regexp.MustCompile(`^mutation \S+ ergebnis=(rot|rot-paket|gruen|flatternd|uebersetzungsfehler|abbruch) test=\S+ paket=\S+ prozesse=\d+ mit-detektor=(-|\d+/\d+) ohne-detektor=(-|\d+/\d+) dauer=\d+\.\d last=\S+/\d+$`)
	for _, f := range faelle {
		t.Run(f.klasse+"/"+f.erg, func(t *testing.T) {
			fp := &fakeP{bauAus: f.bau, startAus: f.start}
			if fp.startAus == nil {
				fp.startAus = func(s Start) Ausgang { return (&fakeP{}).Starten(context.Background(), s) }
			}
			if f.start != nil {
				inner := f.start
				fp.startAus = func(s Start) Ausgang {
					if startNull(s) && len(s.Args) == 2 && s.Args[0] == "-test.list" {
						return Ausgang{Text: strings.TrimSuffix(strings.TrimPrefix(s.Args[1], "^"), "$") + "\n"}
					}
					return inner(s)
				}
			}
			w, aus, _ := fakeWerk(t, baum, fp)
			ms := muts(t, baum, mitRate(mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", f.klasse), 3, 2))
			code := w.Lauf(context.Background(), ms)
			zeile := zeileVon(aus.String(), "M")
			nullStarts := len(fp.starts) - len(fp.mutStarts())
			erwartet := fmt.Sprintf("mutation M ergebnis=%s test=TestAdd paket=src/p prozesse=%d mit-detektor=%s ohne-detektor=%s dauer=1.0 last=S%d/4",
				f.erg, f.prozesse, f.mit, f.ohne, nullStarts)
			if zeile != erwartet {
				t.Errorf("Zeile\n got %q\nwant %q (Exit %d)", zeile, erwartet, code)
			}
			if !format.MatchString(zeile) {
				t.Errorf("Zeile %q gleicht nicht dem Format", zeile)
			}
			if n := len(fp.mutStarts()); n != f.prozesse || n != f.erwarteteStarts {
				t.Errorf("Starts des Fakes %d, prozesse %d", n, f.prozesse)
			}
			zus := fmt.Sprintf("zusammenfassung mutationen=1 rot=%d rot-paket=%d gruen=%d flatternd=%d uebersetzungsfehler=%d abbruch=%d dauer=3.0",
				b2i(f.erg == ergRot), b2i(f.erg == ergRotPaket), b2i(f.erg == ergGruen), b2i(f.erg == ergFlattern), b2i(f.erg == ergUebersetz), b2i(f.erg == ergAbbruch))
			if !strings.Contains(aus.String(), zus+"\n") {
				t.Errorf("Zusammenfassung %q fehlt in %q", zus, aus.String())
			}
		})
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// TestDiagnose (SPEC-210 Nr. 1 c): zu rot-paket stehen die ersten 40 Zeilen der Ausgabe mit dem Namen auf stderr,
// zu rot und gruen nichts.
func TestDiagnose(t *testing.T) {
	baum := fixtureBaum(t)
	var text strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&text, "zeile %d\n", i)
	}
	for _, c := range []struct {
		name  string
		fehl  func(Start) bool
		zeile int
	}{
		{"rot-paket", func(s Start) bool { return !hatArg(s, "-test.run") }, 40},
		{"rot", func(s Start) bool { return hatArg(s, "-test.run") }, 0},
		{"gruen", func(Start) bool { return false }, 0},
	} {
		fp := &fakeP{startAus: func(s Start) Ausgang {
			if s.Nullprobe {
				return (&fakeP{}).Starten(context.Background(), s)
			}
			return Ausgang{Code: map[bool]int{true: 1}[c.fehl(s)], Text: text.String()}
		}}
		w, _, fehl := fakeWerk(t, baum, fp)
		w.Lauf(context.Background(), muts(t, baum, mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere)))
		zeilen := strings.Split(strings.TrimSpace(fehl.String()), "\n")
		if c.zeile == 0 && fehl.Len() != 0 || c.zeile > 0 && (len(zeilen) != c.zeile || zeilen[0] != "M: zeile 1" || zeilen[39] != "M: zeile 40") {
			t.Errorf("%s: %d Zeilen auf stderr: %.80q", c.name, len(zeilen), fehl.String())
		}
	}
}

// TestRateNurFuerNebenlaeufig (SPEC-210 Nr. 7): die Klasse andere startet zwei Prozesse, nebenläufig mit 7 und 3 zehn plus
// einen der Stufe 2, nie mehr als vier zugleich; ein Arm mit 0 entfällt; gemischte Exit-Codes sind flatternd.
func TestRateNurFuerNebenlaeufig(t *testing.T) {
	baum := fixtureBaum(t)
	lauf := func(t *testing.T, fp *fakeP, m ml) (string, *fakeP) {
		t.Helper()
		w, aus, _ := fakeWerk(t, baum, fp)
		w.Lauf(context.Background(), muts(t, baum, m))
		return aus.String(), fp
	}
	m := func(klasse string, d, o int) ml {
		return mitRate(mut("M", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasse), d, o)
	}
	t.Run("andere", func(t *testing.T) {
		_, fp := lauf(t, &fakeP{}, m(klasseAndere, 0, 0))
		s := fp.mutStarts()
		if len(s) != 2 || !hatArg(s[0], "-test.run") || hatArg(s[1], "-test.run") || s[0].Detektor || s[1].Detektor {
			t.Errorf("Starts der Klasse andere: %+v", s)
		}
	})
	t.Run("7 und 3", func(t *testing.T) {
		erg, fp := lauf(t, &fakeP{sperre: 4}, m(klasseNebenlaeufig, 7, 3))
		var race, ohne, paket int
		for _, s := range fp.mutStarts() {
			switch {
			case s.Detektor:
				race++
			case hatArg(s, "-test.run"):
				ohne++
			default:
				paket++
			}
		}
		if race != 7 || ohne != 3 || paket != 1 || fp.maxStart != 4 || fp.verletzung != "" {
			t.Errorf("race %d, ohne %d, paket %d, max zugleich %d (%s); %s", race, ohne, paket, fp.maxStart, fp.verletzung, erg)
		}
	})
	t.Run("Arm mit 0", func(t *testing.T) {
		for _, c := range []struct{ d, o, race, ohne int }{{0, 3, 0, 3}, {3, 0, 3, 0}} {
			_, fp := lauf(t, &fakeP{}, m(klasseNebenlaeufig, c.d, c.o))
			var race, ohne int
			for _, s := range fp.mutStarts() {
				switch {
				case s.Detektor:
					race++
				case hatArg(s, "-test.run"):
					ohne++
				}
			}
			if race != c.race || ohne != c.ohne {
				t.Errorf("Arme %d/%d: race %d, ohne %d", c.d, c.o, race, ohne)
			}
		}
	})
	t.Run("flatternd", func(t *testing.T) {
		for _, c := range []struct {
			name  string
			rot   func(i int, s Start) bool
			erwar string
		}{
			{"ein Prozess im Detektor-Arm", func(i int, s Start) bool { return s.Detektor && i == 3 }, ergFlattern},
			{"nur der Arm ohne Detektor", func(_ int, s Start) bool { return !s.Detektor && hatArg(s, "-test.run") }, ergFlattern},
			{"alle", func(_ int, s Start) bool { return hatArg(s, "-test.run") }, ergRot},
			{"keiner", func(int, Start) bool { return false }, ergGruen},
		} {
			var mu sync.Mutex
			n := 0
			fp := &fakeP{}
			fp.startAus = func(s Start) Ausgang {
				if s.Nullprobe {
					return (&fakeP{}).Starten(context.Background(), s)
				}
				mu.Lock()
				n++
				i := n
				mu.Unlock()
				return Ausgang{Code: map[bool]int{true: 1}[c.rot(i, s)]}
			}
			erg, _ := lauf(t, fp, m(klasseNebenlaeufig, 5, 5))
			if got := ergebnisse(erg)["M"]; got != c.erwar {
				t.Errorf("%s: %s statt %s", c.name, got, c.erwar)
			}
		}
	})
}

// TestSpeicherUndGleichzeitigkeit (SPEC-210 Nr. 9 b, 8 d, 10 c): nie zwei Baue zugleich und nie ein Bau neben einem Start, über
// alle Mutationen und Arme; nach dem Lauf bleibt nichts in der Ablage; der Cache wird nur über der Grenze geleert.
func TestSpeicherUndGleichzeitigkeit(t *testing.T) {
	baum := fixtureBaum(t)
	ms := func() []Mutation {
		return muts(t, baum,
			mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere),
			mitRate(mut("b", "src/p/a.go", "return a + b", "return a * b", "TestAdd", klasseNebenlaeufig), 6, 6),
			mitRate(mut("c", "src/p/a.go", "return a - b", "return a * b", "TestSub", klasseNebenlaeufig), 2, 0))
	}
	t.Run("Baue", func(t *testing.T) {
		fp := &fakeP{bauHalt: 20 * 1e6}
		w, _, _ := fakeWerk(t, baum, fp)
		w.Lauf(context.Background(), ms())
		if fp.maxBau != 1 || fp.verletzung != "" || len(fp.baue) < 6 {
			t.Errorf("Baue zugleich %d, Verletzung %q, Baue %d", fp.maxBau, fp.verletzung, len(fp.baue))
		}
		if e, _ := os.ReadDir(w.Ablage); len(e) != 0 {
			t.Errorf("Ablage nicht leer: %v", e)
		}
	})
	t.Run("Cache", func(t *testing.T) {
		// 1,5 GiB sind 1610612736 Byte; das Programm trägt die Grenze als Konstante.
		if cacheGrenze != 1610612736 {
			t.Errorf("Größengrenze des Caches %d", cacheGrenze)
		}
		for _, c := range []struct {
			groesse int64
			geleert int
		}{{1610612737, 3}, {1610612736, 0}, {1610612735, 0}} {
			fp := &fakeP{cache: c.groesse}
			w, _, fehl := fakeWerk(t, baum, fp)
			w.Lauf(context.Background(), ms())
			if fp.geleert != c.geleert || strings.Count(fehl.String(), "cache geleert") != c.geleert {
				t.Errorf("Cache %d: %d-mal geleert, %d Meldungen, erwartet %d", c.groesse, fp.geleert, strings.Count(fehl.String(), "cache geleert"), c.geleert)
			}
		}
	})
}

// fakeD zeichnet Kopien und Entfernungen auf und kann sie scheitern lassen.
type fakeD struct {
	echteDateien
	mu             sync.Mutex
	ereignisse     []string
	scheitertKopie func(ziel string) bool
	scheitertEntf  func(pfad string) bool
}

func (d *fakeD) Kopieren(baum string, m Modul, ziel string) error {
	d.mu.Lock()
	d.ereignisse = append(d.ereignisse, "kopie:"+ziel)
	d.mu.Unlock()
	if d.scheitertKopie != nil && d.scheitertKopie(ziel) {
		if err := os.MkdirAll(ziel, 0o755); err != nil {
			return err
		}
		_ = os.WriteFile(filepath.Join(ziel, "halb.go"), []byte("pack"), 0o644)
		return errors.New("Platte voll")
	}
	return d.echteDateien.Kopieren(baum, m, ziel)
}

func (d *fakeD) Entfernen(pfad string) error {
	d.mu.Lock()
	d.ereignisse = append(d.ereignisse, "entf:"+pfad)
	d.mu.Unlock()
	if d.scheitertEntf != nil && d.scheitertEntf(pfad) {
		return errors.New("nicht löschbar")
	}
	return d.echteDateien.Entfernen(pfad)
}

func (d *fakeD) index(e string) int {
	for i, x := range d.ereignisse {
		if x == e {
			return i
		}
	}
	return -1
}

// TestKopieAbbruch (SPEC-210 Nr. 10 e): bricht das Anlegen einer Kopie ab, ist die Mutation abbruch und ihre Kopie
// verschwindet vor der nächsten; der Lauf geht weiter, Exit 1.
func TestKopieAbbruch(t *testing.T) {
	baum := fixtureBaum(t)
	w, aus, _ := fakeWerk(t, baum, &fakeP{})
	d := &fakeD{scheitertKopie: func(z string) bool { return filepath.Base(z) == "1" }}
	w.D = d
	ms := muts(t, baum,
		mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere),
		mut("b", "src/p/a.go", "return a - b", "return a + b", "TestSub", klasseAndere))
	code := w.Lauf(context.Background(), ms)
	r := ergebnisse(aus.String())
	k1 := filepath.Join(w.Ablage, "lauf", "1")
	if code != 1 || r["a"] != ergAbbruch || r["b"] != ergGruen {
		t.Errorf("Exit %d, Ergebnisse %v", code, r)
	}
	if e, k2 := d.index("entf:"+k1), d.index("kopie:"+filepath.Join(w.Ablage, "lauf", "2")); e < 0 || k2 < 0 || e > k2 {
		t.Errorf("Kopie 1 nicht vor Kopie 2 entfernt: %v", d.ereignisse)
	}
	if vorhanden(filepath.Join(w.Ablage, "lauf")) {
		t.Error("lauf bleibt zurück")
	}
}

// TestKopieNichtEntfernt (SPEC-210 Nr. 10 e): eine nicht löschbare Kopie beendet den Lauf mit Exit 1.
func TestKopieNichtEntfernt(t *testing.T) {
	baum := fixtureBaum(t)
	w, aus, fehl := fakeWerk(t, baum, &fakeP{})
	w.D = &fakeD{scheitertEntf: func(p string) bool { return filepath.Base(p) == "1" }}
	ms := muts(t, baum,
		mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere),
		mut("b", "src/p/a.go", "return a - b", "return a + b", "TestSub", klasseAndere))
	code := w.Lauf(context.Background(), ms)
	r := ergebnisse(aus.String())
	if code != 1 || r["a"] != ergGruen || r["b"] != ergAbbruch || !strings.Contains(fehl.String(), "Kopie nicht entfernt: "+filepath.Join(w.Ablage, "lauf", "1")) {
		t.Errorf("Exit %d, %v, %q", code, r, fehl.String())
	}
}

// TestMehrereTreffer (SPEC-210 Nr. 5 a): liefert -test.list mehr als eine Zeile, ist das ein Eingabefehler.
func TestMehrereTreffer(t *testing.T) {
	baum := fixtureBaum(t)
	faelle := []struct {
		name string
		list Ausgang
		want string
	}{
		{"zwei Zeilen", Ausgang{Text: "TestAdd\nTestAddx\n"}, "Test TestAdd trifft 2 Tests in src/p"},
		{"fremde Zeile", Ausgang{Text: "TestAndere\n"}, "Test TestAdd trifft 0 Tests in src/p"},
		{"Aufruf scheitert", Ausgang{Code: 1, Text: "TestAdd\n"}, "Test TestAdd trifft 0 Tests in src/p"},
	}
	for _, f := range faelle {
		fp := &fakeP{startAus: func(s Start) Ausgang {
			if len(s.Args) == 2 && s.Args[0] == "-test.list" {
				return f.list
			}
			return Ausgang{}
		}}
		w, aus, fehl := fakeWerk(t, baum, fp)
		code := w.Lauf(context.Background(), muts(t, baum, mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere)))
		if code != 2 || aus.Len() != 0 || !strings.Contains(fehl.String(), f.want) || len(fp.mutStarts()) != 0 {
			t.Errorf("%s: Exit %d, stdout %q, stderr %q", f.name, code, aus.String(), fehl.String())
		}
	}
}

// TestAusgangsstandUebersetztNicht (SPEC-210 Nr. 5 a): übersetzt der unveränderte Stand nicht, ist das ein
// Eingabefehler; es startet kein Prozess.
func TestAusgangsstandUebersetztNicht(t *testing.T) {
	baum := fixtureBaum(t)
	fp := &fakeP{bauAus: func(Bau) Ausgang { return Ausgang{Code: 1} }}
	w, aus, fehl := fakeWerk(t, baum, fp)
	code := w.Lauf(context.Background(), muts(t, baum, mut("a", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere)))
	if code != 2 || aus.Len() != 0 || !strings.Contains(fehl.String(), "Ausgangsstand rot: src/p") || len(fp.starts) != 0 {
		t.Errorf("Exit %d, stdout %q, stderr %q, Starts %d", code, aus.String(), fehl.String(), len(fp.starts))
	}
}
