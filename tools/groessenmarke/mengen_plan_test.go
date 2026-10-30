package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// Die Mengen M1, M2, M3, M7, M8 und M9 des Plans (§3, Runde nach dem Review): jede Menge folgt aus
// dem Wortlaut von SPEC-128 Nr. 3 c, 3 h, 3 p, 3 q und 7 a; die Orakel stehen im Test und nutzen
// keinen Code des Programms. Jeder Test zählt seine Menge selbst. Zeichen jenseits von ASCII
// stehen als Bytewerte (AGENTS.md §3.18).

const prosaZeile = "**Gr\xc3\xb6\xc3\x9fenmarke** (`AGENTS.md`): \xc3\xbcber 2400 Zeilen"

func wiederhole(z byte, n int) string {
	return strings.Repeat(string(z), n)
}

func blockText() string {
	return "```groessenmarke\n" + strings.Join(blockB(), "\n") + "\n```\n"
}

// koeder ist eine zweite Tabelle mit dem Kopf der fünf Zellen: sie zählt nur dort, wo ein Leser sie
// im dargestellten Markdown sieht, also außerhalb jedes Zauns.
func koeder() string {
	return kopf5 + zeile("K1", "z", "f", "m", "andere")
}

// m1Soll prüft das Ergebnis eines Plans, dessen echte Tabelle 10 Zeilen und dessen echter Block
// den Zustand vollständig hat.
func m1Soll(t *testing.T, was, plan string, koederZaehlt bool) {
	t.Helper()
	p := lesePlan(plan)
	if koederZaehlt {
		if !hat(p.fehler, "Tabelle mehrdeutig") {
			t.Errorf("%s: %v statt Tabelle mehrdeutig (der Köder zählt)", was, p.fehler)
		}
		return
	}
	if len(p.fehler) != 0 || p.zustand != vollstaendig || p.tab.andere != 10 {
		t.Errorf("%s: %v, Zustand %d, %d Zeilen (der Köder zählt nicht)", was, p.fehler, p.zustand, p.tab.andere)
	}
}

func TestM1Zaeune(t *testing.T) {
	arten := []byte{'`', '~'}
	tab := tabelleText(0, 10)
	zeilen := strings.Join(blockB(), "\n")
	n := 0

	// (a) Öffnung: Einrückung × Art × Gegenstand. Null bis drei Leerzeichen öffnen einen Zaun,
	// vier Leerzeichen und ein Tab nicht (Nr. 3 q).
	for _, e := range []string{"", " ", "  ", "   ", "    ", "\t"} {
		istZaun := len(e) <= 3 && strings.Trim(e, " ") == ""
		for _, art := range arten {
			z := wiederhole(art, 3)
			n++
			m1Soll(t, fmt.Sprintf("Öffnung %q %c Köder", e, art),
				"# P\n\n"+tab+"\n"+e+z+"\n"+koeder()+e+z+"\n\n"+blockText(), !istZaun)
			n++
			p := lesePlan("# P\n\n" + tab + "\n" + e + z + "groessenmarke\n" + zeilen + "\n" + e + z + "\n")
			if istZaun && (len(p.fehler) != 0 || p.zustand != vollstaendig) ||
				!istZaun && (len(p.fehler) != 0 || p.zustand != keinBlock) {
				t.Errorf("Öffnung %q %c Block: %v, Zustand %d (Zaun: %v)", e, art, p.fehler, p.zustand, istZaun)
			}
		}
	}

	// (b) Info-Zeichenkette: ein Backtick im Rest macht die Zeile eines Backtick-Zauns zu Text, bei
	// Tilden ist jeder Rest erlaubt.
	for _, f := range []struct {
		art  byte
		rest string
	}{
		{'`', ""}, {'`', "x"}, {'`', "x`y"}, {'`', " `"}, {'`', "groessenmarke`"},
		{'~', "x`y"}, {'~', "groessenmarke`"},
	} {
		istZaun := f.art == '~' || !strings.Contains(f.rest, "`")
		z := wiederhole(f.art, 3)
		n++
		m1Soll(t, fmt.Sprintf("Info %c %q", f.art, f.rest),
			"# P\n\n"+tab+"\n"+blockText()+"\n"+z+f.rest+"\n"+koeder()+z+"\n", !istZaun)
	}

	// (c) Schluss: gleiche Art, mindestens so viele Zeichen, kein weiterer Text. Hinter dem
	// Schluss-Kandidaten steht die Tabelle samt Block; sie zählt nur nach einem Schluss.
	schliesst := func(art byte, nOffen int, kandidat string) bool {
		return len(kandidat) >= nOffen && strings.Trim(kandidat, string(art)) == ""
	}
	for _, nOffen := range []int{3, 4, 5} {
		for _, art := range arten {
			andere := byte('`') + byte('~') - art
			kandidaten := []string{
				wiederhole(art, nOffen-1), wiederhole(art, nOffen), wiederhole(art, nOffen+1), wiederhole(art, nOffen+4),
				wiederhole(andere, nOffen-1), wiederhole(andere, nOffen), wiederhole(andere, nOffen+1), wiederhole(andere, nOffen+4),
				wiederhole(art, nOffen) + " x", wiederhole(art, nOffen+1) + " x",
			}
			if len(kandidaten) != 10 {
				t.Fatalf("%d Schlusszeilen statt 10", len(kandidaten))
			}
			for _, k := range kandidaten {
				n++
				p := lesePlan("# P\n\n" + wiederhole(art, nOffen) + "\n" + k + "\n" + tab + "\n" + blockText())
				if schliesst(art, nOffen, k) && (len(p.fehler) != 0 || p.zustand != vollstaendig || p.tab.andere != 10) ||
					!schliesst(art, nOffen, k) && (p.zustand != keinBlock || len(p.fehler) != 0) {
					t.Errorf("Schluss n=%d %c %q: %v, Zustand %d, %d Zeilen (schließt: %v)", nOffen, art, k, p.fehler, p.zustand, p.tab.andere, schliesst(art, nOffen, k))
				}
			}
		}
	}

	// (d) Leerraum an der Schlusszeile: alle 16 Formen schließen.
	for _, f := range [][2]string{{"", ""}, {" ", ""}, {"   ", ""}, {"    ", ""}, {"\t", ""}, {"", " "}, {"", "\t"}, {"", "\r"}} {
		for _, art := range arten {
			n++
			z := wiederhole(art, 3)
			p := lesePlan("# P\n\n" + z + "\n" + f[0] + z + f[1] + "\n" + tab + "\n" + blockText())
			if len(p.fehler) != 0 || p.zustand != vollstaendig || p.tab.andere != 10 {
				t.Errorf("Leerraum %q %c %q: %v, Zustand %d, %d Zeilen", f[0], art, f[1], p.fehler, p.zustand, p.tab.andere)
			}
		}
	}

	// (e) Zaun im Zaun: eine Zeile, die einen Block öffnen würde, öffnet in einem offenen Zaun
	// keinen und schließt ihn nicht.
	for _, aussen := range arten {
		for _, innen := range arten {
			n++
			p := lesePlan("# P\n\n" + tab + "\n" + wiederhole(aussen, 4) + "\n" + wiederhole(innen, 3) + "groessenmarke\n" +
				zeilen + "\n" + wiederhole(innen, 3) + "\n" + wiederhole(aussen, 4) + "\n" + blockText())
			if len(p.fehler) != 0 || p.zustand != vollstaendig || p.tab.andere != 10 {
				t.Errorf("Zaun im Zaun %c in %c: %v, Zustand %d", innen, aussen, p.fehler, p.zustand)
			}
		}
	}

	if n != 111 {
		t.Errorf("M1: %d Elemente statt 111", n)
	}
}

func TestM2Tabelle(t *testing.T) {
	n := 0
	nach := zeile("N1", "z", "f", "m", nebenlaeufig())
	folge := []string{"| F1 | z | f | m | nebenl\xc3\xa4ufig |", "F1 | z | f | m | nebenl\xc3\xa4ufig |"}

	// (a) Ende: die Zeile unter der letzten Datenzeile beendet die Tabelle, ob leer, Text, Text mit
	// Strich in der Mitte, Leerzeichen oder `>` vor dem Strich oder ein Zaun; die Folgezeile und
	// die Zeile nach dem Zaun zählen nie.
	for _, u := range []struct {
		zeile string
		zaun  bool
	}{
		{"", false}, {"Text ohne senkrechten Strich", false}, {"Text | mit Strich", false}, {" | U1 | z | f | m | nebenl\xc3\xa4ufig |", false},
		{"```", true}, {"~~~", true}, {"> | U1 | z | f | m | nebenl\xc3\xa4ufig |", false},
	} {
		for _, f := range folge {
			n++
			text := tabelleText(0, 3) + u.zeile + "\n" + f + "\n"
			if u.zaun {
				text += u.zeile + "\n"
			}
			text += nach
			p := lesePlan(planMit(text, blockB()))
			if len(p.fehler) != 0 || p.tab.andere != 3 || p.tab.nebenlaeufig != 0 {
				t.Errorf("Ende nach %q mit Folgezeile %q: %v, %d andere, %d nebenläufig", u.zeile, f, p.fehler, p.tab.andere, p.tab.nebenlaeufig)
			}
		}
	}

	// (b) Schlussstrich: mit Leerraum, Tab oder `\r` dahinter und ohne Schlussstrich bleiben es
	// fünf Zellen mit erkannter Klasse.
	for _, schluss := range []string{" |", " | ", " |  ", " |\t", " |\r", " | \r", ""} {
		for _, klasse := range []string{nebenlaeufig(), "andere"} {
			n++
			p := lesePlan(planMit(kopf5+"| Z1 | z | f | m | "+klasse+schluss+"\n", blockB()))
			wantNeb := int64(0)
			if klasse == nebenlaeufig() {
				wantNeb = 1
			}
			if len(p.fehler) != 0 || p.tab.nebenlaeufig != wantNeb || p.tab.andere != 1-wantNeb {
				t.Errorf("Schluss %q Klasse %q: %v, %d nebenläufig, %d andere", schluss, klasse, p.fehler, p.tab.nebenlaeufig, p.tab.andere)
			}
		}
	}

	// (c) Trennzeile: genau fünf Zellen aus Strichen, mit Doppelpunkt an einem oder beiden Enden.
	gueltig := regexp.MustCompile(`^\|(:?-+:?\|){5}$`)
	kopfZeile := strings.Split(kopf5, "\n")[0] + "\n"
	for _, sep := range []string{
		"|---|", "|---|---|---|---|", "|---|---|---|---|---|", "|---|---|---|---|---|---|",
		"|:---|:---|:---|:---|:---|", "|---:|---:|---:|---:|---:|", "|:---:|:---:|:---:|:---:|:---:|", "|:|:|:|:|:|",
	} {
		n++
		p := lesePlan(planMit(kopfZeile+sep+"\n"+zeile("Z1", "z", "f", "m", "andere"), blockB()))
		if gueltig.MatchString(sep) && len(p.fehler) != 0 || !gueltig.MatchString(sep) && !hat(p.fehler, "Trennzeile fehlt") {
			t.Errorf("Trennzeile %s: %v", sep, p.fehler)
		}
	}

	if n != 36 {
		t.Errorf("M2: %d Elemente statt 36", n)
	}
}

func TestM3ProsaZeile(t *testing.T) {
	prosa := prosaZeile
	n := 0
	for _, f := range []struct {
		name, text string
		prosa      bool
	}{
		{"Marker allein", "**Gr\xc3\xb6\xc3\x9fenmarke**\n", true},
		{"Marker mit Text", prosa + "\n", true},
		{"Marker mit Text und CR", prosa + "\r\n", true},
		{"ein Leerzeichen davor", " " + prosa + "\n", false},
		{"Listenzeichen davor", "- " + prosa + "\n", false},
		{"Zitat davor", "> " + prosa + "\n", false},
		{"mitten in der Zeile", "Text " + prosa + "\n", false},
		{"im Backtick-Zaun", "```\n" + prosa + "\n```\n", false},
		{"im Tilden-Zaun", "~~~\n" + prosa + "\n~~~\n", false},
		{"im Block groessenmarke", "```groessenmarke\n" + prosa + "\n```\n", false},
	} {
		n++
		p := lesePlan("# P\n\n" + f.text)
		switch {
		case f.prosa && p.zustand != prosaForm:
			t.Errorf("%s: Zustand %d statt Prosa-Form", f.name, p.zustand)
		case !f.prosa && p.zustand == prosaForm:
			t.Errorf("%s: Zustand Prosa-Form", f.name)
		case !f.prosa && f.name != "im Block groessenmarke" && p.zustand != keinBlock:
			t.Errorf("%s: Zustand %d statt kein Block", f.name, p.zustand)
		}
		a := ab("# P\n\n" + f.text)
		if f.name != "im Block groessenmarke" {
			wantErr := "kein Block"
			if f.prosa {
				wantErr = "Prosa-Form"
			}
			a.soll(t, f.name, 0, wantErr)
		}
	}
	if n != 10 {
		t.Errorf("M3: %d Elemente statt 10", n)
	}
}

func TestM7PruefenOhneVollstaendigenBlock(t *testing.T) {
	n := 0
	for _, f := range []struct{ name, plan, fehler string }{
		{"kein Block", "# P\n\nText\n", "kein vollständiger Block"},
		{"Prosa-Form", "# P\n\n" + prosaZeile + "\n", "kein vollständiger Block"},
		{"Marke ausstehend", planB("-basis", "-schaetzung", "-marke"), "kein vollständiger Block"},
		{"ausstehend, Tabelle ohne Treffer", strings.Replace(planB("-basis", "-schaetzung", "-marke"), "| Nr. |", "| Nummer |", 1), "Tabelle fehlt"},
	} {
		n++
		r := pr(f.plan, messungB(0))
		r.soll(t, f.name, 2, "Eingabe: "+f.fehler)
		if f.fehler == "Tabelle fehlt" && strings.Contains(r.alles(), "kein vollständiger Block") {
			t.Errorf("%s: nennt auch kein vollständiger Block: %s", f.name, r.err)
		}
		if r.out != "" {
			t.Errorf("%s: stdout %q", f.name, r.out)
		}
	}
	if n != 4 {
		t.Errorf("M7: %d Elemente statt 4", n)
	}
}

func TestM8Zeilenenden(t *testing.T) {
	n := 0
	lf := func(s string) string { return s }
	crlf := func(s string) string { return strings.ReplaceAll(s, "\n", "\r\n") }
	for _, ende := range []struct {
		name string
		f    func(string) string
	}{{"LF", lf}, {"CRLF", crlf}} {
		zeilen := strings.Join(blockB(), "\n")
		// Gegenstand Tabelle
		n++
		p := lesePlan("# P\n\n" + ende.f(tabelleText(0, 10)) + "\n" + blockText())
		if len(p.fehler) != 0 || p.zustand != vollstaendig || p.tab.andere != 10 {
			t.Errorf("%s Tabelle: %v, Zustand %d, %d Zeilen", ende.name, p.fehler, p.zustand, p.tab.andere)
		}
		// Gegenstand Block (Öffnung und Zeilen)
		n++
		p = lesePlan("# P\n\n" + tabelleText(0, 10) + "\n" + ende.f("```groessenmarke\n"+zeilen+"\n") + "```\n")
		if len(p.fehler) != 0 || p.zustand != vollstaendig || p.blk.marke != 1100 || p.blk.basis != basisB || !reflect.DeepEqual(p.blk.pfade, []string{"src/"}) {
			t.Errorf("%s Block: %v, Zustand %d, %+v", ende.name, p.fehler, p.zustand, p.blk)
		}
		// Gegenstand Prosa-Zeile
		n++
		if p = lesePlan("# P\n\n" + prosaZeile + ende.f("\n")); p.zustand != prosaForm || len(p.fehler) != 0 {
			t.Errorf("%s Prosa-Zeile: %v, Zustand %d", ende.name, p.fehler, p.zustand)
		}
		// Gegenstand Schluss-Zaun
		n++
		p = lesePlan("# P\n\n" + tabelleText(0, 10) + "\n```groessenmarke\n" + zeilen + "\n```" + ende.f("\n") + "Text\n")
		if len(p.fehler) != 0 || p.zustand != vollstaendig {
			t.Errorf("%s Schluss-Zaun: %v, Zustand %d", ende.name, p.fehler, p.zustand)
		}
	}
	// Ein einzelnes `\r` mitten in `marke` ist ein Zeichen der Zeile.
	n++
	if p := lesePlan(planMit(tabelleText(0, 10), blockB("marke: 11\r00"))); !hat(p.fehler, "Zahl ungültig: marke") {
		t.Errorf("einzelnes CR in marke: %v", p.fehler)
	}
	if n != 9 {
		t.Errorf("M8: %d Elemente statt 9", n)
	}
}

func TestM9SchluesselMitLeerraum(t *testing.T) {
	n := 0
	for _, f := range []struct {
		zeile   string
		bekannt bool
	}{
		{"marke: 1100", true}, {" marke: 1100", true}, {"marke : 1100", true}, {"\tmarke: 1100", true}, {"marke\t: 1100", true},
		{"mar ke: 1100", false},
	} {
		n++
		p := lesePlan(planMit(tabelleText(0, 10), blockB("-marke", f.zeile)))
		if f.bekannt && (len(p.fehler) != 0 || p.blk.marke != 1100 || p.zustand != vollstaendig) ||
			!f.bekannt && !hat(p.fehler, "Schlüssel mar ke unbekannt") {
			t.Errorf("Schlüssel %q: %v, marke %d", f.zeile, p.fehler, p.blk.marke)
		}
	}
	if n != 6 {
		t.Errorf("M9: %d Elemente statt 6", n)
	}
}
