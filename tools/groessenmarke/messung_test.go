package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

// blkB ist der Block der Fixture B.
func blkB() block {
	return lesePlan(planB()).blk
}

func TestMessungSumme(t *testing.T) {
	b := blkB()
	rng := rand.New(rand.NewPCG(3, 4))
	stroeme, leer, voll := 0, 0, 0
	for i := 0; i < 300; i++ {
		n := rng.IntN(41)
		var sb strings.Builder
		sb.WriteString(kopfText(basisB, spitzeB, "src/"))
		var want int64
		for j := 0; j < n; j++ {
			ins, del := rng.Int64N(1_000_000_000), rng.Int64N(1_000_000_000)
			if j%7 == 3 {
				ins = 0
			}
			want += ins + del
			sb.WriteString(satz(fmt.Sprint(ins), fmt.Sprint(del), fmt.Sprintf("f%d.go", j)))
		}
		stroeme++
		if n == 0 {
			leer++
		}
		if n == 40 {
			voll++
		}
		if m := leseMessung([]byte(sb.String()), b); m.fehler != "" || m.summe != want || m.binaer != 0 {
			t.Fatalf("Strom %d mit %d Datensätzen: Summe %d (%q), erwartet %d", i, n, m.summe, m.fehler, want)
		}
	}
	if stroeme != 300 || leer == 0 || voll == 0 {
		t.Errorf("%d Ströme, davon %d leer und %d mit 40 Datensätzen", stroeme, leer, voll)
	}
	if m := leseMessung([]byte(kopfText(basisB, spitzeB, "src/")), b); m.fehler != "" || m.summe != 0 {
		t.Errorf("leerer Strom: %+v", m)
	}
	// Einfügungen und Löschungen zählen je für sich.
	if m := leseMessung([]byte(kopfText(basisB, spitzeB, "src/")+satz("5", "0", "a")+satz("0", "7", "b")), b); m.summe != 12 {
		t.Errorf("Einfügungen und Löschungen: %d statt 12", m.summe)
	}
}

func TestDatensatzPfad(t *testing.T) {
	b := blkB()
	rng := rand.New(rand.NewPCG(5, 6))
	var pfade []string
	for i := 0; i < 500; i++ {
		n := 1 + rng.IntN(4096)
		bs := make([]byte, n)
		for j := range bs {
			bs[j] = byte(1 + rng.IntN(255))
		}
		pfade = append(pfade, string(bs))
	}
	generiert := len(pfade)
	pfade = append(pfade, "a b", "a\tb", "a\nb", `a"b`, `a\b`, "\xc3\xa4.txt", "\xff\xfe", strings.Repeat("x", 4096))
	if generiert != 500 || len(pfade) != 508 {
		t.Fatalf("%d generierte und %d Pfade", generiert, len(pfade))
	}
	for i, p := range pfade {
		m := leseMessung([]byte(kopfText(basisB, spitzeB, "src/")+satz("1", "2", p)+satz("3", "4", "x")), b)
		if m.fehler != "" || m.summe != 10 {
			t.Fatalf("Pfad %d (%q): %q, Summe %d", i, p[:min(len(p), 20)], m.fehler, m.summe)
		}
	}
}

func TestBinaer(t *testing.T) {
	b := blkB()
	for _, k := range []int{1, 2, 5} {
		s := kopfText(basisB, spitzeB, "src/") + satz("3", "2", "y.go")
		for i := 0; i < k; i++ {
			s += satz("-", "-", fmt.Sprintf("x%d.bin", i))
		}
		m := leseMessung([]byte(s), b)
		hinweis := fmt.Sprintf("%d Binärdatei(en) nicht gezählt", k)
		if m.fehler != "" || m.summe != 5 || m.binaer != k || m.hinweis != hinweis {
			t.Errorf("%d Binärdateien: %+v", k, m)
		}
		pr(planB(), s).soll(t, hinweis, 0, hinweis, "Messung 5 von Marke 1100")
	}
	for _, d := range []string{satz("-", "3", "x"), satz("3", "-", "x")} {
		m := leseMessung([]byte(kopfText(basisB, spitzeB, "src/")+d), b)
		if !strings.HasPrefix(m.fehler, "Messung nicht lesbar") {
			t.Errorf("einseitiges Minus %q: %+v", d, m)
		}
	}
}

func TestDatensatzUnlesbar(t *testing.T) {
	b := blkB()
	kopf := kopfText(basisB, spitzeB, "src/")
	formen := []struct{ name, rest, fehler string }{
		{"leeres Pfadfeld", "1\t2\t\x00", "Umbenennung in der Messung"},
		{"kein Tab", "12x\x00", "Messung nicht lesbar"},
		{"ein Tab", "1\t2x\x00", "Messung nicht lesbar"},
		{"Buchstaben", satz("a", "b", "x"), "Messung nicht lesbar"},
		{"Minus", satz("-1", "2", "x"), "Messung nicht lesbar"},
		{"zehn Stellen", satz("1234567890", "1", "x"), "Messung nicht lesbar"},
		{"Plus", satz("+1", "2", "x"), "Messung nicht lesbar"},
		{"kein abschließendes NUL", "1\t2\tx", "Messung nicht lesbar"},
		{"Rest nach dem letzten NUL", satz("1", "2", "x") + "3\t4\ty", "Messung nicht lesbar"},
		{"führende Null", satz("01", "2", "x"), "Messung nicht lesbar"},
	}
	for _, f := range formen {
		m := leseMessung([]byte(kopf+f.rest), b)
		if !strings.HasPrefix(m.fehler, f.fehler) {
			t.Errorf("%s: %q statt %q", f.name, m.fehler, f.fehler)
		}
	}
}

func TestMessungKopf(t *testing.T) {
	b := lesePlan(planMit(tabelleText(0, 10), blockB("pfad: tools/"))).blk
	if len(b.pfade) != 2 {
		t.Fatalf("Pfade: %v", b.pfade)
	}
	d := satz("1", "2", "x")
	sp := func(n int) string { return strings.Repeat("a", n) }
	formen := []struct{ name, text, fehler string }{
		{"Kontrolle", kopfText(basisB, spitzeB, "src/", "tools/") + d, ""},
		{"spitze 40", kopfText(basisB, sp(40), "src/", "tools/") + d, ""},
		{"spitze 64", kopfText(basisB, sp(64), "src/", "tools/") + d, ""},
		{"Basis abweichend", kopfText("abcdef0123", spitzeB, "src/", "tools/") + d, fremdePlan},
		{"Pfad fehlt", kopfText(basisB, spitzeB, "src/") + d, fremdePlan},
		{"Pfad zusätzlich", kopfText(basisB, spitzeB, "src/", "tools/", "x/") + d, fremdePlan},
		{"Pfade vertauscht", kopfText(basisB, spitzeB, "tools/", "src/") + d, fremdePlan},
		{"Kopf fehlt", d, nichtLesbar},
		{"numstat fehlt", "basis " + basisB + "\nspitze " + spitzeB + "\npfad src/\npfad tools/\n" + d, nichtLesbar},
		{"spitze fehlt", "basis " + basisB + "\npfad src/\npfad tools/\nnumstat\n" + d, nichtLesbar},
		{"unbekannte Kopfzeile", "basis " + basisB + "\nspitze " + spitzeB + "\nfoo\nnumstat\n" + d, nichtLesbar},
		{"Pfad ohne Wert", "basis " + basisB + "\nspitze " + spitzeB + "\npfad\nnumstat\n" + d, nichtLesbar},
		{"spitze kurz", kopfText(basisB, sp(39), "src/", "tools/") + d, nichtLesbar},
		{"spitze lang", kopfText(basisB, sp(65), "src/", "tools/") + d, nichtLesbar},
		{"spitze mit Großbuchstaben", kopfText(basisB, "A"+sp(39), "src/", "tools/") + d, nichtLesbar},
		{"spitze mit ungültigem Zeichen", kopfText(basisB, "g"+sp(39), "src/", "tools/") + d, nichtLesbar},
	}
	for _, f := range formen {
		m := leseMessung([]byte(f.text), b)
		if f.fehler == "" && m.fehler != "" || f.fehler != "" && !strings.HasPrefix(m.fehler, f.fehler) {
			t.Errorf("%s: %q statt %q", f.name, m.fehler, f.fehler)
		}
	}
}
