package main

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

// orakelMarke ist die kleinste volle Hundert, die 1,35·S nicht unterschreitet, mit big.Rat.
func orakelMarke(s int64) int64 {
	r := big.NewRat(135*s, 10000)
	q := new(big.Int).Div(r.Num(), r.Denom())
	if !r.IsInt() {
		q.Add(q, big.NewInt(1))
	}
	return q.Int64() * 100
}

func TestSchaetzungGitter(t *testing.T) {
	gitterA, gitterB, gitterF, gitterP := []int64{0, 1, 2, 7}, []int64{0, 1, 5, 49}, []int64{0, 1, 8}, []int64{0, 1, 1500}
	n, mitZeilen, ohneZeilen := 0, 0, 0
	for _, a := range gitterA {
		for _, b := range gitterB {
			for _, f := range gitterF {
				for _, p := range gitterP {
					n++
					want := 85*a + 80*b + 200*f + p
					if got := schaetzung(a, b, f, p); got != want {
						t.Errorf("a=%d b=%d f=%d p=%d: %d statt %d", a, b, f, p, got, want)
					}
					r := pr(planMit(tabelleText(int(a), int(b)), blockB(fmt.Sprintf("produktcode: %d", p), fmt.Sprintf("fakes: %d", f))), messungB(0))
					if a+b == 0 {
						ohneZeilen++
						r.soll(t, "Tabelle ohne Zeilen", 2, "Tabelle ohne Zeilen")
						continue
					}
					mitZeilen++
					if got := feld(r.out, "schaetzung"); got != fmt.Sprint(want) {
						t.Errorf("a=%d b=%d f=%d p=%d: Messpunkt schaetzung=%s statt %d", a, b, f, p, got, want)
					}
				}
			}
		}
	}
	if n != 144 || mitZeilen != 135 || ohneZeilen != 9 {
		t.Errorf("Gitter: %d Elemente, %d mit Zeilen, %d ohne; erwartet 144, 135, 9", n, mitZeilen, ohneZeilen)
	}
}

func TestMarkeRundung(t *testing.T) {
	n := 0
	for s := int64(0); s <= 30000; s++ {
		n++
		if got, want := markeZu(s), orakelMarke(s); got != want {
			t.Fatalf("S=%d: Marke %d statt %d", s, got, want)
		}
	}
	// Zweites Orakel: die kleinste volle Hundert m mit m ≥ 1,35·S, gesucht in Schritten von 100.
	for s := int64(0); s <= 3000; s++ {
		m := int64(0)
		for m*10000 < 135*s {
			m++
		}
		if got := markeZu(s); got != m*100 {
			t.Fatalf("S=%d: Marke %d statt %d", s, got, m*100)
		}
	}
	for _, c := range []struct{ s, m int64 }{{2000, 2700}, {2001, 2800}, {999999999, orakelMarke(999999999)}} {
		if got := markeZu(c.s); got != c.m {
			t.Errorf("S=%d: Marke %d statt %d", c.s, got, c.m)
		}
	}
	if n != 30001 {
		t.Errorf("%d Elemente statt 30001", n)
	}
}

func TestSchnittfrageGrenze(t *testing.T) {
	for _, c := range []struct {
		s     int64
		frage bool
		marke int64
	}{{2221, false, 3000}, {2222, false, 3000}, {2223, true, 3100}, {2224, true, 3100}} {
		if got := schnittfrage(c.s); got != c.frage {
			t.Errorf("S=%d: Schnittfrage %v statt %v", c.s, got, c.frage)
		}
		if got := markeZu(c.s); got != c.marke {
			t.Errorf("S=%d: Marke %d statt %d", c.s, got, c.marke)
		}
		r := pr(planMit(tabelleText(0, 1), blockB(fmt.Sprintf("schaetzung: %d", c.s), fmt.Sprintf("marke: %d", c.marke))), messungB(0))
		if c.frage {
			r.soll(t, fmt.Sprint(c.s), 1, "Schnittfrage ohne Grund", "!weicht von der Formel ab")
		} else {
			r.soll(t, fmt.Sprint(c.s), 1, "!Schnittfrage", "!weicht von der Formel ab")
		}
	}
}

func TestMesspunkte(t *testing.T) {
	// Zeilen der Tabelle (davon nebenläufig), Fakes, Produktcode, S, Marke, Ende und die Marke der
	// Formel vor SPEC-128 aus der Tabelle in SPEC-128 Nr. 2.
	punkte := []struct {
		name                     string
		zeilen, neben            int
		fakes, produkt           int
		s, marke, ende, plattMrk int
	}{
		{"csdl", 44, 2, 0, 1506, 5036, 6800, 4333, 5600},
		{"json", 49, 2, 0, 1500, 5430, 7400, 7236, 6000},
		{"import", 70, 3, 8, 2000, 9215, 12500, 10027, 8200},
	}
	for _, p := range punkte {
		blk := func(marke int, extra ...string) []string {
			return blockB(append([]string{
				fmt.Sprintf("produktcode: %d", p.produkt), fmt.Sprintf("fakes: %d", p.fakes),
				fmt.Sprintf("schaetzung: %d", p.s), fmt.Sprintf("marke: %d", marke), "schnitt-grund: Grund",
			}, extra...)...)
		}
		tab := tabelleText(p.neben, p.zeilen-p.neben)
		pr(planMit(tab, blk(p.marke)), messungB(p.ende)).soll(t, p.name, 0,
			fmt.Sprintf("ok, Messung %d von Marke %d", p.ende, p.marke), "!Hinweis")
		pr(planMit(tab, blk(p.plattMrk)), messungB(0)).soll(t, p.name+" Platte", 1,
			"weicht von der Formel ab", fmt.Sprintf("erwartet %d", p.marke))
	}
}

// pruefPlan ist die Fixture B mit der Schätzung s (Tabelle: 1 Zeile, Produktcode s − 80).
func pruefPlan(s int64, aend ...string) string {
	return planMit(tabelleText(0, 1), blockB(append([]string{
		fmt.Sprintf("produktcode: %d", s-80), fmt.Sprintf("schaetzung: %d", s), fmt.Sprintf("marke: %d", markeZu(s)),
	}, aend...)...))
}

func TestSchnittGrund(t *testing.T) {
	n := 0
	for _, s := range []int64{2222, 2223} {
		for _, g := range []struct {
			name string
			zeil []string
			ok   bool
		}{{"fehlt", nil, false}, {"leer", []string{"schnitt-grund:"}, false}, {"Leerraum", []string{"schnitt-grund:   "}, false}, {"Text", []string{"schnitt-grund: Text"}, true}} {
			n++
			r := pr(pruefPlan(s, g.zeil...), messungB(1))
			if s == 2223 && !g.ok {
				r.soll(t, fmt.Sprintf("S=%d %s", s, g.name), 1, "Schnittfrage ohne Grund")
			} else {
				r.soll(t, fmt.Sprintf("S=%d %s", s, g.name), 0, "!Schnittfrage")
			}
		}
	}
	if n != 8 {
		t.Errorf("%d Elemente statt 8", n)
	}
}

func TestMarkeGleichFormel(t *testing.T) {
	n := 0
	for _, marke := range []int{1000, 1099, 1100, 1101, 1200, 0} {
		for _, m := range []int{0, 1} {
			n++
			r := pr(planB(fmt.Sprintf("marke: %d", marke)), messungB(m))
			was := fmt.Sprintf("Marke %d Messung %d", marke, m)
			if marke == 1100 {
				r.soll(t, was, 0, "!weicht von der Formel ab")
			} else {
				r.soll(t, was, 1, fmt.Sprintf("Marke %d weicht von der Formel ab (Schätzung 800, erwartet 1100)", marke))
			}
		}
	}
	if n != 12 {
		t.Errorf("%d Elemente statt 12", n)
	}
}

func TestSchaetzungGleichTabelle(t *testing.T) {
	n := 0
	for _, s := range []int{799, 800, 801} {
		for _, m := range []int{0, 1} {
			n++
			r := pr(planB(fmt.Sprintf("schaetzung: %d", s)), messungB(m))
			was := fmt.Sprintf("Schätzung %d Messung %d", s, m)
			switch {
			case s == 800:
				r.soll(t, was, 0, "!Tabelle ergibt", "!weicht")
			case m == 0:
				r.soll(t, was, 1, fmt.Sprintf("Schätzung %d weicht von der Tabelle ab (Tabelle ergibt 800)", s), "!weicht von der Formel ab")
			default:
				r.soll(t, was, 0, fmt.Sprintf("Hinweis: Tabelle ergibt 800, Schätzung %d", s), "!weicht")
			}
		}
	}
	if n != 6 {
		t.Errorf("%d Elemente statt 6", n)
	}
}

func TestZustandGrenze(t *testing.T) {
	pr(planB("schaetzung: 799"), messungB(0)).soll(t, "Messung 0", 1, "weicht von der Tabelle ab")
	pr(planB("schaetzung: 799"), messungB(1)).soll(t, "Messung 1", 0, "Hinweis: Tabelle ergibt 800, Schätzung 799")
}

// eintragZeile ist die Zeile ueberschreitung mit N und Grund.
func eintragZeile(n int, grund string) string {
	return fmt.Sprintf("ueberschreitung: %d | %s", n, grund)
}

func TestUeberschreitungGrenze(t *testing.T) {
	n := 0
	for _, m := range []int{1099, 1100, 1101} {
		for _, mit := range []bool{false, true} {
			n++
			var aend []string
			if mit {
				aend = append(aend, eintragZeile(m+100, "Grund"))
			}
			r := pr(planB(aend...), messungB(m))
			was := fmt.Sprintf("Messung %d mit Eintrag %v", m, mit)
			switch {
			case m <= 1100:
				r.soll(t, was, 0, fmt.Sprintf("ok, Messung %d von Marke 1100", m), "!ohne Eintrag")
			case mit:
				r.soll(t, was, 0, "ok, Messung 1101 von Marke 1100")
			default:
				r.soll(t, was, 1, "Messung 1101 über Marke 1100 ohne Eintrag", "ueberschreitung: 1200 | <Grund>")
			}
		}
	}
	if n != 6 {
		t.Errorf("%d Elemente statt 6", n)
	}
}

func TestEintragDeckt(t *testing.T) {
	const m = 1101
	layouts := []struct {
		name   string
		stelle int // Zahl der Einträge und Stelle des größten
		pos    int
	}{{"1", 1, 0}, {"2 erster", 2, 0}, {"2 letzter", 2, 1}, {"3 erster", 3, 0}, {"3 mittlerer", 3, 1}, {"3 letzter", 3, 2}}
	n := 0
	for _, groesstes := range []int{m - 1, m, m + 1} {
		for _, l := range layouts {
			n++
			var aend []string
			for i := 0; i < l.stelle; i++ {
				v := groesstes - 50*(i+1)
				if i == l.pos {
					v = groesstes
				}
				aend = append(aend, eintragZeile(v, "Grund"))
			}
			r := pr(planB(aend...), messungB(m))
			was := fmt.Sprintf("größtes N %d, %s", groesstes, l.name)
			if groesstes >= m {
				r.soll(t, was, 0, "ok, Messung 1101")
			} else {
				r.soll(t, was, 1, fmt.Sprintf("Eintrag reicht nicht (größtes N %d)", groesstes), "!ohne Eintrag")
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Elemente statt 18", n)
	}
}

func TestGrund(t *testing.T) {
	// Menge: Form des leeren Grundes (3) × Stelle (erster bis dritter von einem bis drei Einträgen
	// `ueberschreitung`, jeder einzeln leer; ein Eintrag `ersetzt`) × Messung {0, 1, Marke, Marke + 1}.
	// Die Einträge tragen verschiedene N (1200, 1300, 1400), jedes deckt die Messung 1101.
	type stelle struct {
		schluessel string
		k, leer    int
	}
	var stellen []stelle
	for k := 1; k <= 3; k++ {
		for p := 0; p < k; p++ {
			stellen = append(stellen, stelle{"ueberschreitung", k, p})
		}
	}
	stellen = append(stellen, stelle{"ersetzt", 1, 0})
	n := 0
	for _, g := range []string{"", "   ", "\t"} {
		for _, st := range stellen {
			for _, m := range []int{0, 1, 1100, 1101} {
				n++
				var aend []string
				var erwartet []string
				for i := 0; i < st.k; i++ {
					v := 1200 + 100*i
					grund := "Grund"
					if st.schluessel == "ueberschreitung" && i == st.leer {
						grund = g
						erwartet = append(erwartet, fmt.Sprintf("ueberschreitung bis %d ohne Grund", v))
					} else if st.schluessel == "ueberschreitung" {
						erwartet = append(erwartet, fmt.Sprintf("!ueberschreitung bis %d ohne Grund", v))
					}
					if st.schluessel == "ueberschreitung" {
						aend = append(aend, eintragZeile(v, grund))
					}
				}
				if st.schluessel == "ersetzt" {
					aend = append(aend, fmt.Sprintf("ersetzt: 1200 |%s", g))
					erwartet = append(erwartet, "ersetzt 1200 ohne Grund")
				} else {
					erwartet = append(erwartet, "!ersetzt")
				}
				r := pr(planB(aend...), messungB(m))
				was := fmt.Sprintf("Grund %q %s %d Einträge, leer %d, Messung %d", g, st.schluessel, st.k, st.leer, m)
				r.soll(t, was, 1, append(erwartet, "!reicht nicht")...)
				if st.schluessel == "ersetzt" && m == 1101 {
					r.soll(t, was, 1, "ohne Eintrag")
				} else {
					r.soll(t, was, 1, "!ohne Eintrag")
				}
			}
		}
	}
	if want := 3 * 7 * 4; n != want {
		t.Errorf("%d Elemente statt %d", n, want)
	}
	pr(planB(eintragZeile(1200, "Grund")), messungB(1100)).soll(t, "Eintrag ohne Überschreitung", 0, "Hinweis: Eintrag ohne Überschreitung")
	pr(planB("ersetzt: 1200 | Grund"), messungB(1100)).soll(t, "ersetzt ohne Überschreitung", 0, "!Eintrag ohne Überschreitung")
	pr(planB(eintragZeile(1200, "Grund")), messungB(1101)).soll(t, "Eintrag mit Überschreitung", 0, "!Eintrag ohne Überschreitung")
}

// befundSchluessel sind die Texte der fünf Befunde in der Reihenfolge Schnitt, Marke, Tabelle,
// Überschreitung, Grund (SPEC-128 Nr. 5 i).
func befundSchluessel() []string {
	return []string{"Schnittfrage ohne Grund", "weicht von der Formel ab", "weicht von der Tabelle ab", "über Marke", "ueberschreitung bis"}
}

// befundPlan baut einen Plan, der genau die Befunde der Schalter meldet (Reihenfolge g, b, c, d, e).
func befundPlan(sch [5]bool) (string, string) {
	s := int64(800)
	if sch[0] {
		s = 2223
	}
	schaetzungBlock := s
	if sch[2] {
		schaetzungBlock = s + 1
	}
	marke := markeZu(schaetzungBlock)
	if sch[1] {
		marke += 100
	}
	m := 1
	if sch[2] {
		m = 0
	}
	aend := []string{fmt.Sprintf("produktcode: %d", s-80), fmt.Sprintf("schaetzung: %d", schaetzungBlock), fmt.Sprintf("marke: %d", marke)}
	if !sch[0] {
		aend = append(aend, "schnitt-grund: Text")
	}
	if sch[3] {
		m = int(marke) + 1
	}
	if sch[4] {
		n := 5000
		if sch[3] {
			n = 0
		}
		aend = append(aend, eintragZeile(n, ""))
	}
	messungText := kopfText(basisB, spitzeB, "src/")
	if m > 0 {
		messungText += satz(fmt.Sprint(m), "0", "f.go")
	}
	return planMit(tabelleText(0, 1), blockB(aend...)), messungText
}

func TestRangfolge(t *testing.T) {
	schluessel := befundSchluessel()
	n, moeglich := 0, 0
	for mask := 0; mask < 32; mask++ {
		var sch [5]bool
		var erwartet []string
		for i := range sch {
			if mask&(1<<i) != 0 {
				sch[i] = true
				erwartet = append(erwartet, schluessel[i])
			}
		}
		n++
		if sch[2] && sch[3] {
			continue // Schätzung ≠ Tabelle gilt nur bei Messung 0, eine Überschreitung braucht mehr als 0
		}
		moeglich++
		plan, m := befundPlan(sch)
		r := pr(plan, m)
		was := fmt.Sprintf("Schalter %05b", mask)
		code := 1
		if len(erwartet) == 0 {
			code = 0
		}
		r.soll(t, was, code)
		zeilen := strings.Split(strings.TrimRight(r.err, "\n"), "\n")
		if len(erwartet) == 0 {
			zeilen = nil
		}
		if len(zeilen) != len(erwartet) {
			t.Errorf("%s: %d Meldungen statt %d\n%s", was, len(zeilen), len(erwartet), r.err)
			continue
		}
		for i, z := range zeilen {
			if !strings.Contains(z, erwartet[i]) {
				t.Errorf("%s: Meldung %d ist %q, erwartet Text %q", was, i, z, erwartet[i])
			}
		}
	}
	if n != 32 || moeglich != 24 {
		t.Errorf("%d Schalter, %d möglich; erwartet 32 und 24", n, moeglich)
	}
	// Ein Eingabefehler schließt jeden Befund aus.
	r := pr(planB("produktcode: 1.5", "marke: 1000"), messungB(1200))
	r.soll(t, "Eingabefehler mit Befunden", 2, "Zahl ungültig: produktcode", "!weicht", "!ohne Eintrag")
	r = pr(planB("marke: 1000"), kopfText("abcdef0", spitzeB, "src/"))
	r.soll(t, "Messung fremd mit Befund", 2, "gehört nicht zu diesem Plan", "!weicht")
	// Ein Eintrag deckt nur die Überschreitung, nie die Abweichung der Marke.
	pr(planB("marke: 1000", eintragZeile(1100, "Grund")), messungB(1001)).soll(t, "Eintrag und Marke unter Formel", 1,
		"weicht von der Formel ab", "!ohne Eintrag", "!reicht nicht")
	pr(planB("marke: 1000"), messungB(1001)).soll(t, "Marke unter Formel und Messung über Marke", 1,
		"weicht von der Formel ab", "ohne Eintrag")
}
