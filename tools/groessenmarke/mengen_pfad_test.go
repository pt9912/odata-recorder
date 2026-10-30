package main

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// Die Mengen M4 (Pfade, SPEC-128 Nr. 3 m) und M5 (Kopf der Messung, Nr. 4 d). Das Orakel zu M4
// liest den Pfad nach dem Wortlaut von Nr. 3 m, Segment für Segment.

// pfadOrakel: relativ, nicht leer, kein leeres Segment (auch kein führender und höchstens ein
// schließender Schrägstrich), kein `.` oder `..`, kein Steuerzeichen, kein `:` am Anfang, nicht in
// `spec/`, `docs/`, `harness/vertraege/` und nicht gleich einem von ihnen.
func pfadOrakel(p string) bool {
	if p == "" || p[0] == ':' {
		return false
	}
	for _, c := range []byte(p) {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	segmente := strings.Split(p, "/")
	if segmente[len(segmente)-1] == "" && len(segmente) > 1 {
		segmente = segmente[:len(segmente)-1]
	}
	for _, s := range segmente {
		if s == "" || s == "." || s == ".." {
			return false
		}
	}
	if segmente[0] == "spec" || segmente[0] == "docs" {
		return false
	}
	return len(segmente) < 2 || segmente[0] != "harness" || segmente[1] != "vertraege"
}

func TestM4Pfade(t *testing.T) {
	var pfade []string
	teile := []string{"a", "", ".", "..", "spec", "docs", "harness", "vertraege"}
	folgen := [][]string{nil}
	var alle [][]string
	for laenge := 1; laenge <= 3; laenge++ {
		var naechste [][]string
		for _, f := range folgen {
			for _, s := range teile {
				naechste = append(naechste, append(append([]string(nil), f...), s))
			}
		}
		folgen = naechste
		alle = append(alle, naechste...)
	}
	for _, f := range alle {
		for _, schluss := range []string{"", "/", "//"} {
			pfade = append(pfade, strings.Join(f, "/")+schluss)
		}
	}
	if len(alle) != 584 || len(pfade) != 1752 {
		t.Fatalf("(a) %d Folgen und %d Pfade statt 584 und 1752", len(alle), len(pfade))
	}
	// (c) Steuerzeichen 0x00 bis 0x1F und 0x7F an Anfang, Mitte und Ende.
	var steuer []byte
	for c := 0; c < 0x20; c++ {
		steuer = append(steuer, byte(c))
	}
	steuer = append(steuer, 0x7f)
	for _, c := range steuer {
		pfade = append(pfade, string(c)+"a", "a"+string(c)+"b", "a"+string(c))
	}
	// (d) Bereiche und ihre Nachbarn.
	for _, b := range []string{"spec", "docs", "harness/vertraege"} {
		pfade = append(pfade, b, b+"/", b+"/x", b+"x", b+"x/y")
	}
	pfade = append(pfade, "harness", "harness/")
	// Die Doppelung (b) prüft der Block, nicht ein einzelner Pfad.
	if len(pfade) != 1752+99+17 {
		t.Fatalf("%d Pfade statt %d", len(pfade), 1752+99+17)
	}

	n, imBlock := 0, 0
	for _, p := range pfade {
		n++
		want := pfadOrakel(p)
		if got := pfadGueltig(p); got != want {
			t.Errorf("pfadGueltig(%q) = %v statt %v", p, got, want)
		}
		// Im Block gilt der getrimmte Wert einer Zeile: ein Pfad mit Leerraum am Rand oder mit
		// Zeilenumbruch bleibt auf der Ebene von pfadGueltig.
		if strings.TrimSpace(p) != p || strings.Contains(p, "\n") {
			continue
		}
		imBlock++
		r := lesePlan(planMit(tabelleText(0, 10), blockB("-pfad", "pfad: "+p)))
		if want && len(r.fehler) != 0 || !want && (len(r.fehler) != 1 || r.fehler[0] != "Pfad ungültig: "+p) {
			t.Errorf("Pfad %q im Block: %v (gültig: %v)", p, r.fehler, want)
		}
	}

	// (b) Doppelung: nur dieselbe Zeichenkette ist doppelt.
	for _, d := range [][2]string{{"x", "x"}, {"x/", "x/"}, {"x", "x/"}, {"x/", "x"}, {"x", "x/y"}} {
		n++
		r := lesePlan(planMit(tabelleText(0, 10), blockB("-pfad", "pfad: "+d[0], "pfad: "+d[1])))
		if d[0] == d[1] && (len(r.fehler) != 1 || r.fehler[0] != "Pfad doppelt: "+d[1]) || d[0] != d[1] && len(r.fehler) != 0 {
			t.Errorf("Paar %q und %q: %v", d[0], d[1], r.fehler)
		}
	}
	if n != 1873 {
		t.Errorf("M4: %d Elemente statt 1873", n)
	}
	// Elf Pfade tragen Leerraum oder Zeilenumbruch am Rand oder in der Mitte und bleiben auf der
	// Ebene von pfadGueltig (fünf Zeichen am Anfang, fünf am Ende, ein Zeilenumbruch in der Mitte).
	if imBlock != 1752+99+17-11 {
		t.Errorf("M4: %d Pfade im Block statt %d", imBlock, 1752+99+17-11)
	}
}

func TestM5KopfDerMessung(t *testing.T) {
	pfade := []string{"src/", "tools/", "Dockerfile"}
	b := lesePlan(planMit(tabelleText(0, 10), blockB("-pfad", "pfad: src/", "pfad: tools/", "pfad: Dockerfile"))).blk
	if !slices.Equal(b.pfade, pfade) {
		t.Fatalf("Pfade des Blocks: %v", b.pfade)
	}
	d := satz("1", "2", "x")
	kopf := func(basis, spitze string, ps []string) string { return kopfText(basis, spitze, ps...) + d }
	mitPfaden := func(ps []string) string { return kopf(basisB, spitzeB, ps) }

	type element struct {
		name, text, fehler string // fehler: "" gültig, sonst Anfang der Meldung
	}
	var es []element
	es = append(es, element{"gültiger Kopf", mitPfaden(pfade), ""})
	for i := range pfade {
		ohne := append(append([]string(nil), pfade[:i]...), pfade[i+1:]...)
		es = append(es, element{fmt.Sprintf("Pfad %d fehlt", i), mitPfaden(ohne), fremdePlan})
	}
	for i := 0; i <= len(pfade); i++ {
		mit := append(append(append([]string(nil), pfade[:i]...), "neu/"), pfade[i:]...)
		es = append(es, element{fmt.Sprintf("Pfad zusätzlich an Stelle %d", i), mitPfaden(mit), fremdePlan})
	}
	for i, p := range pfade {
		erstes := append([]string(nil), pfade...)
		erstes[i] = "#" + p[1:]
		letztes := append([]string(nil), pfade...)
		letztes[i] = p[:len(p)-1] + "#"
		es = append(es,
			element{fmt.Sprintf("Pfad %d: erstes Byte anders", i), mitPfaden(erstes), fremdePlan},
			element{fmt.Sprintf("Pfad %d: letztes Byte anders", i), mitPfaden(letztes), fremdePlan})
	}
	for _, paar := range [][2]int{{0, 1}, {0, 2}, {1, 2}} {
		v := append([]string(nil), pfade...)
		v[paar[0]], v[paar[1]] = v[paar[1]], v[paar[0]]
		es = append(es, element{fmt.Sprintf("Pfade %d und %d vertauscht", paar[0], paar[1]), mitPfaden(v), fremdePlan})
	}
	for i := range pfade {
		v := append([]string(nil), pfade...)
		v[i] += "/"
		es = append(es, element{fmt.Sprintf("Schrägstrich an Pfad %d", i), mitPfaden(v), fremdePlan})
	}
	if len(es) != 20 {
		t.Fatalf("Pfadliste: %d Elemente statt 20", len(es))
	}
	// Basis: erstes Zeichen, letztes Zeichen, um eins kürzer, um eins länger.
	for name, basis := range map[string]string{
		"Basis erstes Zeichen": "1" + basisB[1:], "Basis letztes Zeichen": basisB[:len(basisB)-1] + "8",
		"Basis kürzer": basisB[:len(basisB)-1], "Basis länger": basisB + "0",
	} {
		es = append(es, element{name, kopf(basis, spitzeB, pfade), fremdePlan})
	}
	// spitze: 39, 40, 64 und 65 Zeichen; ein Großbuchstabe; Nicht-Hex an erster und letzter Stelle.
	sp := func(n int) string { return strings.Repeat("a", n) }
	for name, f := range map[string]struct{ s, fehler string }{
		"spitze 39": {sp(39), nichtLesbar}, "spitze 40": {sp(40), ""}, "spitze 64": {sp(64), ""}, "spitze 65": {sp(65), nichtLesbar},
		"spitze Großbuchstabe": {"A" + sp(39), nichtLesbar}, "spitze Nicht-Hex vorn": {"g" + sp(39), nichtLesbar},
		"spitze Nicht-Hex hinten": {sp(39) + "g", nichtLesbar},
	} {
		es = append(es, element{name, kopf(basisB, f.s, pfade), f.fehler})
	}
	if len(es) != 31 {
		t.Fatalf("M5: %d Elemente statt 31", len(es))
	}
	for _, e := range es {
		m := leseMessung([]byte(e.text), b)
		if e.fehler == "" && (m.fehler != "" || m.summe != 3) || e.fehler != "" && !strings.HasPrefix(m.fehler, e.fehler) {
			t.Errorf("%s: %q (Summe %d) statt %q", e.name, m.fehler, m.summe, e.fehler)
		}
	}
}
