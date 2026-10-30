package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestTabelleErkennen(t *testing.T) {
	zeileOk := zeile("Z1", "z", "f", "m", "andere")
	sep := "|---|---|---|---|---|\n"
	var koepfe []string
	koepfe = append(koepfe,
		"| Zusage | Nr. | F\xc3\xa4nger | Mutation | Klasse |\n",
		"| Nr. | Zusage | F\xc3\xa4nger | Mutation |\n",
		"| Nr. | Zusage | F\xc3\xa4nger | Mutation | Klasse | Extra |\n",
		"| Nr. | Zusage | Faenger | Mutation | Klasse |\n",
		"| Zusage | Mutation | roter Test |\n",
	)
	basis := []string{"Nr.", "Zusage", "F\xc3\xa4nger", "Mutation", "Klasse"}
	for i := range basis {
		c := append([]string(nil), basis...)
		c[i] = "X"
		koepfe = append(koepfe, zeile(c...))
	}
	var faelle []string
	for _, k := range koepfe {
		faelle = append(faelle, k+sep+zeileOk)
	}
	faelle = append(faelle, "```\n"+kopf5+zeileOk+"```\n", "~~~\n"+kopf5+zeileOk+"~~~\n")
	if len(faelle) != 12 {
		t.Fatalf("%d Köpfe statt 12", len(faelle))
	}
	for i, tab := range faelle {
		p := lesePlan(planMit(tab, blockB()))
		if !hat(p.fehler, "Tabelle fehlt") {
			t.Errorf("Kopf %d: %v statt Tabelle fehlt", i, p.fehler)
		}
	}
	// Die Zellen des Kopfs werden getrimmt verglichen.
	getrimmt := "|Nr.|  Zusage |F\xc3\xa4nger  |\tMutation\t|Klasse|\n" + sep + zeileOk
	if p := lesePlan(planMit(getrimmt, blockB())); len(p.fehler) != 0 {
		t.Errorf("Kopf mit Leerraum: %v", p.fehler)
	}
	zwei := lesePlan(planMit(kopf5+zeileOk+"\nText\n\n"+kopf5+zeileOk, blockB()))
	if !hat(zwei.fehler, "Tabelle mehrdeutig") {
		t.Errorf("zwei Tabellen: %v", zwei.fehler)
	}
	if p := lesePlan(planMit(kopf5, blockB())); !hat(p.fehler, "Tabelle ohne Zeilen") {
		t.Errorf("keine Datenzeile: %v", p.fehler)
	}
	// Ein kürzerer Zaun schließt einen längeren nicht: die Tabelle hinter der Zeile mit drei
	// Backticks liegt noch im Zaun mit vier.
	if p := lesePlan(planMit("````\n```\n"+kopf5+zeileOk+"````\n", blockB())); !hat(p.fehler, "Tabelle fehlt") {
		t.Errorf("kürzerer Zaun schließt: %v", p.fehler)
	}
	// Eine Tabelle im Zaun neben der echten zählt nicht mit.
	neben := lesePlan(planMit("~~~\n"+kopf5+zeileOk+"~~~\n\n"+kopf5+zeileOk, blockB()))
	if len(neben.fehler) != 0 {
		t.Errorf("Tabelle im Zaun neben der echten: %v", neben.fehler)
	}
}

func TestZeilenZellen(t *testing.T) {
	bad := func(k int) string {
		return zeile(strings.Split(strings.Repeat("x,", k), ",")[:k]...)
	}
	for _, k := range []int{3, 4, 6, 7} {
		p := lesePlan(planMit(kopf5+bad(k), blockB()))
		if !hat(p.fehler, fmt.Sprintf("Zeile 1: %d Zellen statt 5", k)) {
			t.Errorf("%d Zellen: %v", k, p.fehler)
		}
	}
	zwei := lesePlan(planMit(kopf5+zeile("Z1", "z", "f", "m", "andere")+bad(4), blockB()))
	if !hat(zwei.fehler, "Zeile 2: 4 Zellen statt 5") || len(zwei.fehler) != 1 {
		t.Errorf("zweite Zeile: %v", zwei.fehler)
	}
	if p := lesePlan(planMit(kopf5+"| Z1 | z | f | m | andere\n", blockB())); len(p.fehler) != 0 || p.tab.andere != 1 {
		t.Errorf("Zeile ohne Schlussstrich: %v", p.fehler)
	}
	if p := lesePlan(planMit(kopf5+"| Z1 | `a|b` | f | m | andere |\n", blockB())); !hat(p.fehler, "Zeile 1: 6 Zellen statt 5") {
		t.Errorf("Strich im Code-Span: %v", p.fehler)
	}
	geschuetzt := 0
	for i := 0; i < 5; i++ {
		c := []string{"Z1", "z", "f", "m", "andere"}
		c[i] = c[i][:1] + `\|` + c[i][1:]
		p := lesePlan(planMit(kopf5+zeile(c...), blockB()))
		geschuetzt++
		for _, f := range p.fehler {
			if strings.Contains(f, "Zellen statt") {
				t.Errorf("geschützter Strich in Zelle %d: %v", i, p.fehler)
			}
		}
		if i < 4 && len(p.fehler) != 0 {
			t.Errorf("geschützter Strich in Zelle %d: %v", i, p.fehler)
		}
		if i == 4 && !hat(p.fehler, "Zeile Z1: Klasse 'a|ndere' unbekannt") {
			t.Errorf("geschützter Strich in der Klasse: %v", p.fehler)
		}
	}
	if geschuetzt != 5 {
		t.Errorf("%d Zellen statt 5", geschuetzt)
	}
	row := zeile("Z1", "z", "f", "m", "andere")
	for _, s := range []string{"|:---|:---|:---|:---|:---|", "|---:|---:|---:|---:|---:|", "|:---:|:---:|:---:|:---:|:---:|", "|-|-|-|-|-|"} {
		if p := lesePlan(planMit(strings.Split(kopf5, "\n")[0]+"\n"+s+"\n"+row, blockB())); len(p.fehler) != 0 {
			t.Errorf("Trennzeile %s: %v", s, p.fehler)
		}
	}
	kopfZeile := strings.Split(kopf5, "\n")[0] + "\n"
	for _, s := range []string{"|     |     |     |     |     |", "|---|---|---|---|--x|", "| Z1 | z | f | m | andere |", "|:|:|:|:|:|"} {
		if p := lesePlan(planMit(kopfZeile+s+"\n"+row, blockB())); !hat(p.fehler, "Trennzeile fehlt") {
			t.Errorf("Trennzeile %s: %v", s, p.fehler)
		}
	}
	if p := lesePlan(planMit(kopfZeile, blockB())); !hat(p.fehler, "Trennzeile fehlt") {
		t.Errorf("Kopf allein: %v", p.fehler)
	}
}

func TestNrEindeutig(t *testing.T) {
	zw := func(a, b string) string {
		return kopf5 + zeile(a, "z", "f", "m", "andere") + zeile(b, "z", "f", "m", "andere")
	}
	for _, c := range []struct {
		name, tab, fehler string
	}{
		{"leer", kopf5 + zeile("", "z", "f", "m", "andere"), "Nr. leer"},
		{"doppelt", zw("Z1", "Z1"), "Nr. Z1 doppelt"},
		{"doppelt bis auf Leerraum", zw("Z1", "  Z1 "), "Nr. Z1 doppelt"},
		{"verschieden", zw("Z1", "Z2"), ""},
	} {
		p := lesePlan(planMit(c.tab, blockB()))
		if c.fehler == "" && len(p.fehler) != 0 || c.fehler != "" && (!hat(p.fehler, c.fehler) || len(p.fehler) != 1) {
			t.Errorf("%s: %v statt %q", c.name, p.fehler, c.fehler)
		}
	}
}

func TestKlasse(t *testing.T) {
	gueltig := []struct{ zelle, neb string }{{nebenlaeufig(), "1"}, {"andere", "0"}, {"  andere \t", "0"}}
	for _, g := range gueltig {
		p := lesePlan(planMit(kopf5+zeile("Z1", "z", "f", "m", g.zelle), blockB()))
		if len(p.fehler) != 0 || fmt.Sprint(p.tab.nebenlaeufig) != g.neb {
			t.Errorf("Klasse %q: %v, nebenläufig %d", g.zelle, p.fehler, p.tab.nebenlaeufig)
		}
	}
	ungueltig := []string{
		"", "nebenlaeufig", "Nebenl\xc3\xa4ufig", "NEBENL\xc3\x84UFIG", "z\xc3\xa4hlt nicht", "**andere**", "andere.",
		"nebenl\xc3\xa4ufig/andere", "nebenla\xcc\x88ufig",
	}
	if len(ungueltig) != 9 {
		t.Fatalf("%d ungültige Formen statt 9", len(ungueltig))
	}
	for _, z := range ungueltig {
		p := lesePlan(planMit(kopf5+zeile("Z1", "z", "f", "m", z), blockB()))
		if !hat(p.fehler, fmt.Sprintf("Zeile Z1: Klasse '%s' unbekannt", strings.TrimSpace(z))) || p.tab.nebenlaeufig+p.tab.andere != 0 {
			t.Errorf("Klasse %q: %v", z, p.fehler)
		}
	}
}

func TestZaehlung(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	plaene := 0
	for i := 0; i < 200; i++ {
		plaene++
		n := 1 + rng.IntN(60)
		var sb strings.Builder
		sb.WriteString(kopf5)
		var neb, and int64
		for j := 0; j < n; j++ {
			klasse := "andere"
			if rng.IntN(2) == 0 {
				klasse = nebenlaeufig()
				neb++
			} else {
				and++
			}
			text := "z"
			if j == 0 {
				text = "Allaussage: jede Referenzart, jede Position, jeder Rest"
			}
			sb.WriteString(zeile(fmt.Sprintf("Z%d", j+1), text, "f", "m", klasse))
		}
		p := lesePlan(planMit(sb.String(), blockB()))
		if len(p.fehler) != 0 || p.tab.nebenlaeufig != neb || p.tab.andere != and {
			t.Fatalf("Plan %d mit %d Zeilen: a=%d b=%d (erwartet %d und %d) %v", i, n, p.tab.nebenlaeufig, p.tab.andere, neb, and, p.fehler)
		}
	}
	if plaene != 200 {
		t.Errorf("%d Pläne statt 200", plaene)
	}
}

func TestBlockErkennen(t *testing.T) {
	tab := tabelleText(0, 1)
	zeilen := strings.Join(blockB(), "\n")
	zaunZu := func(z, info string) string { return z + info + "\n" + zeilen + "\n" + z + "\n" }
	faelle := []struct {
		name, plan string
		fehler     string
		zustand    zustand
	}{
		{"zwei Blöcke", "# P\n\n" + tab + zaunZu("```", "groessenmarke") + zaunZu("```", "groessenmarke"), "Block mehrdeutig", markeAusstehend},
		{"nicht geschlossen", "# P\n\n" + tab + "```groessenmarke\n" + zeilen + "\n", "Block nicht geschlossen", markeAusstehend},
		{"groessenmarke2", "# P\n\n" + tab + zaunZu("```", "groessenmarke2"), "", keinBlock},
		{"Groessenmarke", "# P\n\n" + tab + zaunZu("```", "Groessenmarke"), "", keinBlock},
		{"Info mit Zusatz", "# P\n\n" + tab + zaunZu("```", "groessenmarke extra"), "", keinBlock},
		{"Info mit Leerraum", "# P\n\n" + tab + zaunZu("```", "groessenmarke  "), "", vollstaendig},
		{"Tilden", "# P\n\n" + tab + zaunZu("~~~", "groessenmarke"), "", vollstaendig},
		{"im Zaun eines anderen", "# P\n\n" + tab + "~~~\n" + zaunZu("```", "groessenmarke") + "~~~\n", "", keinBlock},
		{"im Zaun mit Backticks", "# P\n\n" + tab + "````\n" + zaunZu("```", "groessenmarke") + "````\n", "", keinBlock},
		{"Zeile ohne Doppelpunkt", planMit(tab, append(blockB(), "xyz")), "Zeile im Block ohne Schlüssel: xyz", vollstaendig},
	}
	faelle = append(faelle, struct {
		name, plan string
		fehler     string
		zustand    zustand
	}{"zwei Backticks und Leerzeilen", "# P\n\n" + tab + "``code`` im Text\n\n```groessenmarke\n\n" + zeilen + "\n\n```\n", "", vollstaendig})
	for _, f := range faelle {
		p := lesePlan(f.plan)
		if f.fehler == "" && len(p.fehler) != 0 || f.fehler != "" && !hat(p.fehler, f.fehler) {
			t.Errorf("%s: %v statt %q", f.name, p.fehler, f.fehler)
		}
		if f.fehler == "" && p.zustand != f.zustand {
			t.Errorf("%s: Zustand %d statt %d", f.name, p.zustand, f.zustand)
		}
	}
}

// blockMitSchluessel ist ein Block mit allen neun Schlüsseln, der Schlüssel k n-mal.
func blockMitSchluessel(k string, n int) []string {
	alle := []string{"pfad: src/", "produktcode: 0", "fakes: 0", "basis: " + basisB, "schaetzung: 800", "marke: 1100",
		"schnitt-grund: Grund", "ersetzt: 1000 | alt", "ueberschreitung: 1200 | Grund"}
	var out []string
	for _, z := range alle {
		if strings.HasPrefix(z, k+":") {
			continue
		}
		out = append(out, z)
	}
	for i := 0; i < n; i++ {
		z := alle[0]
		for _, a := range alle {
			if strings.HasPrefix(a, k+":") {
				z = a
			}
		}
		if i > 0 && k == "pfad" {
			z = "pfad: tools/"
		}
		if i > 0 && k == "ueberschreitung" {
			z = "ueberschreitung: 1300 | Grund"
		}
		out = append(out, z)
	}
	return out
}

func TestSchluessel(t *testing.T) {
	n := 0
	for _, k := range []string{"pfad", "produktcode", "fakes", "basis", "schaetzung", "marke", "schnitt-grund", "ersetzt", "ueberschreitung"} {
		for _, mal := range []int{1, 2} {
			n++
			p := lesePlan(planMit(tabelleText(0, 10), blockMitSchluessel(k, mal)))
			doppelt := mal == 2 && k != "pfad" && k != "ueberschreitung"
			if doppelt != hat(p.fehler, "Schlüssel "+k+" doppelt") || (!doppelt && len(p.fehler) != 0) {
				t.Errorf("Schlüssel %s %d-mal: %v", k, mal, p.fehler)
			}
		}
	}
	if n != 18 {
		t.Errorf("%d Elemente statt 18", n)
	}
	for _, k := range []string{"marken", "Marke", "mar ke"} {
		p := lesePlan(planMit(tabelleText(0, 10), blockB(k+": 1100")))
		if !hat(p.fehler, "Schlüssel "+k+" unbekannt") {
			t.Errorf("Schlüssel %q: %v", k, p.fehler)
		}
	}
}

func TestZahlen(t *testing.T) {
	gueltig := []string{"0", "7", "123456789"}
	ungueltig := []string{"", "1 500", "1.500", "1,5", "1'500", "-1", "+1", "01", "1234567890", "0x10", "\xef\xbc\x91\xef\xbc\x92", "7 7"}
	n := 0
	for _, k := range []string{"produktcode", "fakes", "schaetzung", "marke"} {
		for _, v := range append(append([]string(nil), gueltig...), ungueltig...) {
			n++
			p := lesePlan(planMit(tabelleText(0, 10), blockB(k+": "+v)))
			erwartet := !contains(gueltig, v)
			if erwartet != hat(p.fehler, "Zahl ungültig: "+k) || (!erwartet && len(p.fehler) != 0) || (erwartet && len(p.fehler) != 1) {
				t.Errorf("%s: %q ergibt %v", k, v, p.fehler)
			}
		}
	}
	if n != 60 {
		t.Errorf("%d Elemente statt 60", n)
	}
}

func contains(liste []string, s string) bool {
	for _, x := range liste {
		if x == s {
			return true
		}
	}
	return false
}

func TestArchitektTeil(t *testing.T) {
	namen := []string{"pfad", "produktcode", "fakes"}
	n := 0
	for mask := 1; mask < 8; mask++ {
		var weg, fehlt []string
		for i, k := range namen {
			if mask&(1<<i) != 0 {
				weg = append(weg, "-"+k)
				fehlt = append(fehlt, k)
			}
		}
		n++
		p := lesePlan(planMit(tabelleText(0, 10), blockB(weg...)))
		if len(p.fehler) != 1 || p.fehler[0] != "Block unvollständig: es fehlen "+strings.Join(fehlt, ", ") {
			t.Errorf("fehlt %v: %v", fehlt, p.fehler)
		}
	}
	if n != 7 {
		t.Errorf("%d Teilmengen statt 7", n)
	}
}

func TestPlannerTeil(t *testing.T) {
	namen := []string{"basis", "schaetzung", "marke"}
	n := 0
	for mask := 0; mask < 8; mask++ {
		var weg, fehlt []string
		for i, k := range namen {
			if mask&(1<<i) == 0 {
				weg = append(weg, "-"+k)
				fehlt = append(fehlt, k)
			}
		}
		n++
		plan := planMit(tabelleText(0, 10), blockB(weg...))
		p := lesePlan(plan)
		switch len(fehlt) {
		case 3:
			a := ab(plan)
			a.soll(t, "ausstehend", 0, "Marke steht aus: Schätzung 800, Marke 1100")
			if len(p.fehler) != 0 || p.zustand != markeAusstehend || a.out != "keine-messung\n" {
				t.Errorf("ausstehend: %v %d %q", p.fehler, p.zustand, a.out)
			}
		case 0:
			if len(p.fehler) != 0 || p.zustand != vollstaendig {
				t.Errorf("voll: %v %d", p.fehler, p.zustand)
			}
		default:
			if len(p.fehler) != 1 || p.fehler[0] != "Block unvollständig: es fehlen "+strings.Join(fehlt, ", ") {
				t.Errorf("fehlt %v: %v", fehlt, p.fehler)
			}
			ab(plan).soll(t, "unvollständig", 2, "Block unvollständig")
		}
	}
	if n != 8 {
		t.Errorf("%d Teilmengen statt 8", n)
	}
}

func TestBasis(t *testing.T) {
	gueltig := []string{"abc1234", "abc12345", strings.Repeat("a", 40)}
	ungueltig := []string{"", "abc123", strings.Repeat("a", 41), "ABCDEF0123", "-xabcdef", "HEAD1234", "abcd..bcd", "abc:defg", "@abcdefg", "abcdef~1", "abc def0"}
	for _, b := range append(append([]string(nil), gueltig...), ungueltig...) {
		p := lesePlan(planMit(tabelleText(0, 10), blockB("basis: "+b)))
		bad := !contains(gueltig, b)
		if bad != hat(p.fehler, "Basis ungültig: "+b) || (!bad && len(p.fehler) != 0) {
			t.Errorf("Basis %q: %v", b, p.fehler)
		}
	}
}

func TestPfade(t *testing.T) {
	gueltig := []string{"src/a.go", "tools/x", "tools/x/", "harness/", "a b/c d.txt", "a*", "\xc3\xa4.txt",
		"specs/x", "docsx", "harness/vertraegex", "harness/vertraege-x"}
	ungueltig := []string{"", "/x", ":x", ".", "..", "a/./b", "a/../b", "a//b", "a\x00b", "a\x1fb", "a\x7fb", "spec", "spec/lastenheft.md",
		"docs", "docs/plan", "harness/vertraege", "harness/vertraege/README.md", "spec/", "docs/"}
	for _, p := range gueltig {
		if r := lesePlan(planMit(tabelleText(0, 10), blockB("-pfad", "pfad: "+p))); len(r.fehler) != 0 {
			t.Errorf("Pfad %q: %v", p, r.fehler)
		}
	}
	for _, p := range ungueltig {
		if r := lesePlan(planMit(tabelleText(0, 10), blockB("-pfad", "pfad: "+p))); len(r.fehler) != 1 || r.fehler[0] != "Pfad ungültig: "+p {
			t.Errorf("Pfad %q: %v", p, r.fehler)
		}
	}
	r := lesePlan(planMit(tabelleText(0, 10), blockB("pfad: src/")))
	if len(r.fehler) != 1 || r.fehler[0] != "Pfad doppelt: src/" {
		t.Errorf("Pfad doppelt: %v", r.fehler)
	}
	if len(gueltig) != 11 || len(ungueltig) != 19 {
		t.Errorf("%d gültige und %d ungültige Pfade", len(gueltig), len(ungueltig))
	}
}

func TestEintraege(t *testing.T) {
	formen := []struct {
		wert    string
		gueltig bool
		n       int64
		grund   string
	}{
		{"1200", false, 0, ""}, {"x | G", false, 0, ""}, {"1200 |", true, 1200, ""}, {"1200 |   ", true, 1200, ""},
		{"1200 | a | b", true, 1200, "a | b"}, {"0 | G", true, 0, "G"},
	}
	n := 0
	for _, k := range []string{"ueberschreitung", "ersetzt"} {
		for _, f := range formen {
			n++
			p := lesePlan(planMit(tabelleText(0, 10), blockB(k+": "+f.wert)))
			if !f.gueltig {
				if len(p.fehler) != 1 || p.fehler[0] != "Eintrag ungültig: "+k {
					t.Errorf("%s %q: %v", k, f.wert, p.fehler)
				}
				continue
			}
			liste := p.blk.ueberschreitung
			if k == "ersetzt" {
				liste = p.blk.ersetzt
			}
			if len(p.fehler) != 0 || len(liste) != 1 || liste[0].n != f.n || liste[0].grund != f.grund {
				t.Errorf("%s %q: %v %v", k, f.wert, p.fehler, liste)
			}
		}
	}
	if n != 12 {
		t.Errorf("%d Elemente statt 12", n)
	}
}

func TestZustand(t *testing.T) {
	prosa := "**Gr\xc3\xb6\xc3\x9fenmarke** (`AGENTS.md` \xc2\xa73.14): \xc3\xbcber **%d** Zeilen (Summe aus Einf\xc3\xbcgungen und L\xc3\xb6schungen)\n"
	type zFall struct {
		name, plan     string
		zustand        zustand
		stdout, stderr string
	}
	var plaene []zFall
	for _, n := range []int{2400, 4500, 8200, 1400} {
		plaene = append(plaene, zFall{fmt.Sprint("Prosa ", n), "# P\n\n" + fmt.Sprintf(prosa, n), prosaForm, "keine-messung\n", "Prosa-Form"})
	}
	plaene = append(plaene,
		zFall{"kein Block", "# P\n\nText ohne Tabelle\n", keinBlock, "keine-messung\n", "kein Block"},
		zFall{"ausstehend", planB("-basis", "-schaetzung", "-marke"), markeAusstehend, "keine-messung\n", "Marke steht aus"},
		zFall{"vollständig", planB(), vollstaendig, "messen\nbasis " + basisB + "\npfad src/\n", ""},
		zFall{"Prosa im Zaun", "# P\n\n```\n" + fmt.Sprintf(prosa, 1) + "```\n", keinBlock, "keine-messung\n", "kein Block"},
	)
	if len(plaene) != 8 {
		t.Fatalf("%d Pläne statt 8", len(plaene))
	}
	for _, c := range plaene {
		p := lesePlan(c.plan)
		a := ab(c.plan)
		if len(p.fehler) != 0 || p.zustand != c.zustand || a.out != c.stdout || a.code != 0 || !strings.Contains(a.err, c.stderr) {
			t.Errorf("%s: %v Zustand %d, stdout %q, stderr %q, Exit %d", c.name, p.fehler, p.zustand, a.out, a.err, a.code)
		}
	}
	// Mit Block gilt der Block, auch neben einer Prosa-Zeile.
	if p := lesePlan(planB() + fmt.Sprintf(prosa, 1)); p.zustand != vollstaendig {
		t.Errorf("Block neben Prosa: Zustand %d", p.zustand)
	}
}
