package main

import (
	"context"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// zaehlP zählt die Starts der Mutationen (ohne Nullprobe) vor dem echten Port.
type zaehlP struct {
	Prozesse
	mu sync.Mutex
	n  int
}

func (z *zaehlP) Starten(ctx context.Context, s Start) Ausgang {
	if !s.Nullprobe {
		z.mu.Lock()
		z.n++
		z.mu.Unlock()
	}
	return z.Prozesse.Starten(ctx, s)
}

// TestKopieJeMutation (SPEC-210 Nr. 4 b): das Ergebnis jeder Mutation ist in jeder Liste gleich dem der allein gefahrenen.
func TestKopieJeMutation(t *testing.T) {
	baum := fixtureBaum(t)
	a := func(n string) ml {
		return mut(n, "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere)
	}
	b := func() ml { return mut("B", "src/p/a.go", "return a - b", "return a - b + 0", "TestAdd", klasseAndere) }
	u := func() ml { return mut("U", "src/p/a.go", "return a + b", "return a +", "TestAdd", klasseAndere) }
	solo := map[string]string{"A": ergRot, "A2": ergRot, "B": ergGruen, "U": ergUebersetz}
	listen := [][]ml{{a("A"), b()}, {b(), a("A")}, {u(), b()}, {a("A"), a("A2")}, {u(), a("A"), b(), a("A2")}}
	for i, l := range listen {
		w, aus, _ := echtWerk(t, baum)
		w.Lauf(context.Background(), muts(t, baum, l...))
		for n, e := range ergebnisse(aus.String()) {
			if e != solo[n] {
				t.Errorf("Liste %d: %s ist %s, allein %s", i, n, e, solo[n])
			}
		}
		if len(ergebnisse(aus.String())) != len(l) {
			t.Errorf("Liste %d: %d Zeilen für %d Mutationen", i, len(ergebnisse(aus.String())), len(l))
		}
	}
}

// TestGleichwertigWieEinzellauf (SPEC-210 Nr. 6 c, 4 c): je Art der Mutation und Klasse gleicht das Ergebnis dem Einzellauf
// `go test -count=1 [-race]` auf frischer Kopie; der Baum bleibt unverändert.
func TestGleichwertigWieEinzellauf(t *testing.T) {
	erwartet := map[string]map[string]string{
		klasseAndere: {"zustaendig": ergRot, "anderer": ergRotPaket, "aequivalent": ergGruen, "syntax": ergUebersetz,
			"typ": ergUebersetz, "panik": ergRot, "endlos": ergRot, "rennen": ergGruen},
		klasseNebenlaeufig: {"zustaendig": ergRot, "anderer": ergRotPaket, "aequivalent": ergGruen, "syntax": ergUebersetz,
			"typ": ergUebersetz, "panik": ergRot, "endlos": ergRot, "rennen": ergFlattern},
	}
	for _, klasse := range []string{klasseAndere, klasseNebenlaeufig} {
		t.Run(klasse, func(t *testing.T) {
			baum := fixtureBaum(t)
			vor := baumHash(t, baum)
			w, aus, fehl := echtWerk(t, baum)
			ms := arten(klasse)
			for _, m := range ms {
				mitRate(m, 3, 3)
			}
			code := w.Lauf(context.Background(), muts(t, baum, ms...))
			if code != 1 || baumHash(t, baum) != vor {
				t.Errorf("Exit %d (erwartet 1), Baum verändert: %v\n%s", code, baumHash(t, baum) != vor, fehl)
			}
			r := ergebnisse(aus.String())
			if len(r) != 8 {
				t.Fatalf("%d Zeilen statt 8:\n%s", len(r), aus)
			}
			for _, m := range ms {
				n := m["name"].(string)
				o := orakel(t, baum, m, klasse == klasseNebenlaeufig)
				grob := func(e string) string {
					if e == ergRotPaket || e == ergFlattern {
						return ergRot
					}
					return e
				}
				if r[n] != erwartet[klasse][n] || grob(r[n]) != grob(o) {
					t.Errorf("%s: Werkzeug %s, erwartet %s, Einzellauf %s", n, r[n], erwartet[klasse][n], o)
				}
			}
		})
	}
}

// TestBaumUnveraendert (SPEC-210 Nr. 4 c): jede Art der Mutation lässt den Baum des Aufrufers unverändert.
func TestBaumUnveraendert(t *testing.T) {
	for _, m := range arten(klasseAndere) {
		t.Run(m["name"].(string), func(t *testing.T) {
			baum := fixtureBaum(t)
			vor := baumHash(t, baum)
			w, _, _ := echtWerk(t, baum)
			w.Lauf(context.Background(), muts(t, baum, m))
			if baumHash(t, baum) != vor {
				t.Error("der Baum des Aufrufers hat sich geändert")
			}
		})
	}
}

// TestUebersetzungsfehlerIstKeinRot (SPEC-210 Nr. 5 d): Syntaxfehler, Typfehler und fehlender Import sind uebersetzungsfehler,
// starten keinen Prozess und zählen nicht als rot.
func TestUebersetzungsfehlerIstKeinRot(t *testing.T) {
	formen := map[string][2]string{
		"syntax": {"return a + b", "return a +"},
		"typ":    {"return a + b", "return \"x\""},
		"import": {"import \"sync\"\n\n", ""},
	}
	for _, klasse := range []string{klasseAndere, klasseNebenlaeufig} {
		for name, f := range formen {
			t.Run(klasse+"/"+name, func(t *testing.T) {
				baum := fixtureBaum(t)
				w, aus, _ := echtWerk(t, baum)
				z := &zaehlP{Prozesse: w.P}
				w.P = z
				code := w.Lauf(context.Background(), muts(t, baum, mitRate(mut("M", "src/p/a.go", f[0], f[1], "TestAdd", klasse), 2, 2)))
				zeile := zeileVon(aus.String(), "M")
				if code != 1 || feldWert(zeile, "ergebnis") != ergUebersetz || feldWert(zeile, "prozesse") != "0" || z.n != 0 ||
					!strings.Contains(aus.String(), "rot=0 rot-paket=0 gruen=0 flatternd=0 uebersetzungsfehler=1") {
					t.Errorf("Exit %d, Starts %d, Zeile %q", code, z.n, zeile)
				}
			})
		}
	}
}

// TestNullprobe (SPEC-210 Nr. 5 a): ein Name ohne Treffer, ein Paket ohne Testdatei oder ohne Tests und ein roter Ausgangsstand
// sind Eingabefehler (Exit 2) ohne Ergebniszeile; ein Name als Präfix eines zweiten Tests trifft nur sich selbst.
func TestNullprobe(t *testing.T) {
	faelle := []struct{ name, datei, alt, neu, test, meldung string }{
		{"Name ohne Treffer", "src/p/a.go", "return a + b", "return a - b", "TestGibtEsNicht", "Test TestGibtEsNicht trifft 0 Tests in src/p"},
		{"Paket ohne Testdatei", "src/q/q.go", "return 1", "return 2", "TestQ", "Test TestQ trifft 0 Tests in src/q"},
		{"Paket ohne Tests", "src/r/r.go", "package r", "package r\n\nvar X = 1", "TestR", "Test TestR trifft 0 Tests in src/r"},
		{"roter Ausgangsstand", "src/s/s.go", "return 1", "return 2", "TestRot", "Ausgangsstand rot: src/s"},
		{"Ausgangsstand übersetzt nicht", "src/u/u.go", "return \"x\"", "return 1", "TestU", "Ausgangsstand rot: src/u"},
	}
	for _, klasse := range []string{klasseAndere, klasseNebenlaeufig} {
		for _, f := range faelle {
			t.Run(klasse+"/"+f.name, func(t *testing.T) {
				baum := fixtureBaum(t)
				w, aus, fehl := echtWerk(t, baum)
				code := w.Lauf(context.Background(), muts(t, baum, mitRate(mut("M", f.datei, f.alt, f.neu, f.test, klasse), 2, 2)))
				if code != 2 || aus.Len() != 0 || !strings.Contains(fehl.String(), f.meldung) {
					t.Errorf("Exit %d, stdout %q, stderr %q, erwartet %q", code, aus, fehl, f.meldung)
				}
			})
		}
	}
	t.Run("Detektor-Nullprobe", func(t *testing.T) {
		baum := fixtureBaum(t)
		w, aus, fehl := echtWerk(t, baum)
		code := w.Lauf(context.Background(), muts(t, baum, mitRate(mut("M", "src/v/v.go", "z++", "z += 1", "TestV", klasseNebenlaeufig), 2, 2)))
		if code != 2 || aus.Len() != 0 || !strings.Contains(fehl.String(), "Ausgangsstand rot: src/v") {
			t.Errorf("Exit %d, stdout %q, stderr %q", code, aus, fehl)
		}
		w, aus, fehl = echtWerk(t, baum)
		code = w.Lauf(context.Background(), muts(t, baum, mut("M", "src/v/v.go", "z++", "z += 1", "TestV", klasseAndere)))
		if code != 0 || feldWert(zeileVon(aus.String(), "M"), "ergebnis") != ergGruen {
			t.Errorf("andere: Exit %d, stdout %q, stderr %q", code, aus, fehl)
		}
	})
	t.Run("Präfix", func(t *testing.T) {
		baum := fixtureBaum(t)
		w, aus, fehl := echtWerk(t, baum)
		code := w.Lauf(context.Background(), muts(t, baum, mut("M", "src/p/a.go", "return a + b", "if a == b {\n\t\treturn 0\n\t}\n\treturn a + b", "TestAdd", klasseAndere)))
		if code != 0 || feldWert(zeileVon(aus.String(), "M"), "ergebnis") != ergRotPaket {
			t.Errorf("Exit %d, %q, %q", code, aus, fehl)
		}
	})
}

// TestCacheAendertAussageNicht (SPEC-210 Nr. 8 b): zwei Mutationen derselben Datei von gleicher Länge hintereinander geben mit
// leerem und mit warmem Cache dieselben Zeilen außer dauer; die zweite erbt das Ergebnis der ersten nicht.
func TestCacheAendertAussageNicht(t *testing.T) {
	baum := fixtureBaum(t)
	ms := []ml{
		mut("erste", "src/p/a.go", "return a + b", "return a - b", "TestAdd", klasseAndere),
		mut("zweite", "src/p/a.go", "return a + b", "return b + a", "TestAdd", klasseAndere),
	}
	cache := t.TempDir()
	dauer := regexp.MustCompile(` dauer=\S+`)
	var laeufe []string
	for _, art := range []string{"kalt", "warm"} {
		w, aus, _ := echtWerk(t, baum)
		e := w.P.(*echt)
		e.env, e.cache = append(e.env, "GOCACHE="+cache), cache
		w.Lauf(context.Background(), muts(t, baum, ms...))
		laeufe = append(laeufe, dauer.ReplaceAllString(aus.String(), ""))
		if r := ergebnisse(aus.String()); r["erste"] != ergRot || r["zweite"] != ergGruen {
			t.Errorf("%s: %v", art, r)
		}
	}
	if laeufe[0] != laeufe[1] {
		t.Errorf("kalt und warm unterscheiden sich:\n%s\n%s", laeufe[0], laeufe[1])
	}
	if e, _ := os.ReadDir(cache); len(e) == 0 {
		t.Error("der Cache blieb leer")
	}
}
