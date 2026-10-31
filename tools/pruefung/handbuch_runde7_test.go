package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Fänger der siebten Runde zu slice-handbuch-geruest (Z21 erweitert, Z90 bis Z93), je als Schleife über
// eine Tabelle; die Orakel stehen von Hand im Test. Zeichen jenseits von ASCII entstehen aus Bytewerten.

func rune1(r rune) string { return string(r) }

// Z21 (Nr. 6 c): Menge der Zeilen des Ausgangswertes, 50 Elemente: 7 gültige, 43 ungültige. Das Element
// `leere Zeile` und das Element `Abschnitt ohne Zeile` sind verschiedene Elemente.
func TestHandbuchAusgangswertMenge(t *testing.T) {
	type el struct {
		ex      string // Rest der Datei hinter der Trennzeile `--- Ausgangswert`
		gueltig bool
	}
	var els []el
	for _, z := range []string{"0", "1", "9", "10", "99", "100", "255"} {
		els = append(els, el{z + "\n", true})
	}
	ungueltig := []string{
		// (i) Zahl der Zeilen im Abschnitt
		"", "\n", "1\n\n", "\n1\n", "1\n1\n", "\n\n",
		// (ii) keine Dezimalzahl
		"a\n", "-\n", "+1\n", "-1\n", "1.5\n", "0x1\n", "1e2\n", rune1(0xFF11) + "\n", rune1(0x0663) + "\n",
		// (iii) außerhalb von 0 bis 255
		"256\n", "257\n", "999\n", "1000\n", "18446744073709551616\n", "18446744073709551621\n",
		"99999999999999999999\n", "4294967296\n", "100000\n",
		// (iv) führende Null
		"00\n", "01\n", "000\n", "007\n", "0255\n",
		// (v) Rest vor, nach oder statt der Zahl
		"1 \n", "1\t\n", "1" + nbsp + "\n", "1 2\n", "1a\n", "1\r\r\n", " 1\n", "\t1\n", nbsp + "1\n",
		" \n", "\t\n", nbsp + "\n", rune1(0x3000) + "\n",
		// (v) Schluss-LF fehlt
		"1",
	}
	for _, ex := range ungueltig {
		els = append(els, el{ex, false})
	}
	if len(els) != 50 {
		t.Fatalf("%d Elemente, erwartet 50", len(els))
	}
	for _, e := range els {
		code0 := 0
		if e.gueltig {
			code0, _ = strconv.Atoi(strings.TrimSuffix(e.ex, "\n"))
		}
		a := map[string]antwort{"version": {out: "odata-recorder dev\n"}, "": {code: code0}}
		m := bspUmfeld(t, "k.txt", bspDatei(praef, nil, nil, e.ex), a)
		code, _, errText := m.lauf()
		starts := len(m.e.starts(t))
		if e.gueltig {
			if code != exitOK || starts != 2 {
				t.Errorf("%q: Exit %d, %d Starts, Ausgabe %q", e.ex, code, starts, errText)
			}
			continue
		}
		if c := strings.Count(errText, "k.txt: Beispiel-Datei ungültig: "); code != exitBefund || c != 1 || starts != 1 {
			t.Errorf("%q: Exit %d, %d Befunde, %d Starts, Ausgabe %q", e.ex, code, c, starts, errText)
		}
	}
}

// Z90 (Nr. 2 f, 3 a, 2 c, 5 d, 5 f, 8, 9 a): Pfade, die es nicht gibt, 12 Fälle E1 bis E12. Bestand heißt:
// der Fall liegt schon in einem früheren Test (E4: Z44, E5 und E7: Z2 und Z3, E8 in Teilen: Z65).
func TestHandbuchNichtVorhandenePfade(t *testing.T) {
	kopfZeilen := strings.Count(kopfSauber+"\n## A\n\n", "\n")
	fall := func(t *testing.T, m umfeld, bin, wurzel string, code int, soll []string, starts int, bestand bool, name string) {
		t.Helper()
		c, out, errText := fuehre(m.u, "handbuch", bin, wurzel)
		ist := befundZeilen(t, errText)
		ok := c == code && len(ist) == len(soll) && len(m.e.starts(t)) == starts
		for i := 0; ok && i < len(soll); i++ {
			ok = strings.Contains(ist[i], soll[i])
		}
		if !ok {
			t.Errorf("%s (Bestand %v): Exit %d (erwartet %d), stdout %q, Zeilen %q (erwartet %q), %d Starts (erwartet %d)",
				name, bestand, c, code, out, ist, soll, len(m.e.starts(t)), starts)
		}
	}
	datei := func(t *testing.T, p string) string {
		t.Helper()
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	bind := binde("beispiele/a.txt")
	n := 0
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		fall(t, m, m.e.bin, datei(t, filepath.Join(t.TempDir(), "w")), exitAufruf, []string{"Eingabe: Handbuch fehlt: "}, 0, false, "E1 Wurzel ist eine Datei")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("docs")
		m.schreibe("docs", "x\n")
		fall(t, m, m.e.bin, m.wurzel, exitAufruf, []string{"Eingabe: Handbuch fehlt: "}, 0, false, "E2 docs ist eine Datei")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("docs/user")
		m.schreibe("docs/user", "x\n")
		fall(t, m, m.e.bin, m.wurzel, exitAufruf, []string{"Eingabe: Handbuch fehlt: "}, 0, false, "E3 docs/user ist eine Datei")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne(handbuchPfad)
		if err := os.MkdirAll(filepath.Join(m.wurzel, handbuchPfad), 0o755); err != nil {
			t.Fatal(err)
		}
		fall(t, m, m.e.bin, m.wurzel, exitAufruf, []string{"Eingabe: Handbuch nicht lesbar: "}, 0, true, "E4 Handbuch ist ein Verzeichnis")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		fall(t, m, m.e.bin, filepath.Join(t.TempDir(), "gibtsnicht"), exitAufruf, []string{"Eingabe: Handbuch fehlt: "}, 0, true, "E5 Wurzel nicht vorhanden")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		bin := datei(t, filepath.Join(t.TempDir(), "d")) + "/odata-recorder"
		fall(t, m, bin, m.wurzel, exitAufruf, []string{"Eingabe: Binary fehlt: " + bin}, 0, false, "E6 Komponente vor dem Binary ist eine Datei")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		fall(t, m, t.TempDir(), m.wurzel, exitAufruf, []string{"Eingabe: Binary fehlt: "}, 0, true, "E7 Binary ist ein Verzeichnis")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		bin := datei(t, filepath.Join(t.TempDir(), "d")) + "/odata-recorder"
		fall(t, m, bin, datei(t, filepath.Join(t.TempDir(), "w")), exitAufruf, []string{"Eingabe: Binary fehlt: ", "Eingabe: Handbuch fehlt: "}, 0, true, "E8 Wurzel ist eine Datei und das Binary fehlt")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("docs/user/beispiele")
		m.schreibe("docs/user/beispiele", "x\n")
		m.handbuch(rumpf(bind + zaunVon(aufrufDatei)))
		z := kopfZeilen + 1 + 1
		fall(t, m, m.e.bin, m.wurzel, exitBefund, []string{fmt.Sprintf("Zeile %d: beispiele/a.txt fehlt", z)}, 1, false, "E9 beispiele ist eine Datei")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("docs/user/beispiele")
		if err := os.MkdirAll(filepath.Join(m.wurzel, "docs/user/beispiele/a.txt"), 0o755); err != nil {
			t.Fatal(err)
		}
		m.handbuch(rumpf(bind + zaunVon(aufrufDatei)))
		z := kopfZeilen + 1 + 1
		fall(t, m, m.e.bin, m.wurzel, exitBefund, []string{"beispiele: a.txt keine Beispiel-Datei", fmt.Sprintf("Zeile %d: beispiele/a.txt fehlt", z)}, 1, false, "E10 a.txt ist ein Verzeichnis")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("docs/user/beispiele")
		m.schreibe("harness/mk", "x\n")
		m.schreibe("Makefile", "x: ## x\n")
		m.handbuch(rumpf("Nenne `make x` und `make y`.\n"))
		fall(t, m, m.e.bin, m.wurzel, exitBefund, []string{fmt.Sprintf("Zeile %d: make-Ziel y unbekannt", kopfZeilen+1)}, 1, false, "E11 harness/mk ist eine Datei")
	}
	{
		n++
		m := neuesUmfeld(t, standardAntworten())
		m.entferne("Makefile")
		if err := os.MkdirAll(filepath.Join(m.wurzel, "Makefile"), 0o755); err != nil {
			t.Fatal(err)
		}
		h := sauber() + "\nNenne `make x`.\n"
		m.handbuch(h)
		fall(t, m, m.e.bin, m.wurzel, exitBefund, []string{"Aufrufpräfix nicht gedeckt: Image", fmt.Sprintf("Zeile %d: make-Ziel x unbekannt", strings.Count(sauber(), "\n")+2)}, 2, false, "E12 Makefile ist ein Verzeichnis")
	}
	if n != 12 {
		t.Fatalf("%d Fälle, erwartet 12", n)
	}
}

// Z91 (Nr. 5 c, 5 f, 7 a): eine Bindungszeile nennt die Datei nur vor einem geschlossenen Zaun. 6 Fälle
// L1 bis L6 mit Befundliste (Teilzeichenketten, in dieser Reihenfolge) und Zahl der Starts von Hand.
func TestHandbuchBindungszeileAllein(t *testing.T) {
	bz := binde("beispiele/a.txt")
	datei := bspDatei(praef, nil, nil, "0\n")
	abweichend := bspDatei(praef, []string{"x"}, nil, "0\n")
	faelle := []struct {
		name   string
		h      string
		code   int
		befund []string
		starts int
	}{
		{"L1 am Dateiende", bz, exitBefund, []string{"beispiele: a.txt ohne Bindung", "Bindungszeile ohne Zaun"}, 1},
		{"L2 vor Prosa", bz + "Prosa.\n", exitBefund, []string{"beispiele: a.txt ohne Bindung", "Bindungszeile ohne Zaun"}, 1},
		{"L3 durch Leerzeile vom Zaun getrennt", bz + "\n" + zaunVon(datei), exitBefund, []string{"beispiele: a.txt ohne Bindung", "Bindungszeile ohne Zaun", "Beispiel ohne Bindungszeile"}, 1},
		{"L4 vor nicht geschlossenem Zaun", bz + "```\ntext\n", exitBefund, []string{"beispiele: a.txt ohne Bindung", "Zaun nicht geschlossen"}, 1},
		{"L5 Block weicht ab", bz + zaunVon(abweichend), exitBefund, []string{"Beispiel weicht von beispiele/a.txt ab"}, 2},
		{"L6 allein und gültiger Zaun", bz + "\n" + bz + zaunVon(datei), exitBefund, []string{"Bindungszeile ohne Zaun"}, 2},
	}
	if len(faelle) != 6 {
		t.Fatalf("%d Fälle, erwartet 6", len(faelle))
	}
	for _, f := range faelle {
		m := neuesUmfeld(t, okAntworten())
		m.entferne("docs/user/beispiele/aufruf.txt")
		m.schreibe("docs/user/beispiele/a.txt", datei)
		m.handbuch(rumpf(f.h))
		code, _, errText := m.lauf()
		ist := befundZeilen(t, errText)
		ok := code == f.code && len(ist) == len(f.befund) && len(m.e.starts(t)) == f.starts
		for i := 0; ok && i < len(f.befund); i++ {
			ok = strings.Contains(ist[i], f.befund[i])
		}
		if !ok {
			t.Errorf("%s: Exit %d, Zeilen %q (erwartet %q), %d Starts (erwartet %d)", f.name, code, ist, f.befund, len(m.e.starts(t)), f.starts)
		}
	}
}

// Z92 (Nr. 3 b, 13 a): eine Zeile aus Leerraum jeder Art trennt Absätze. 30 Fälle; Zeile 1 `Ein Text.`, Zeile 2
// das Element, Zeile 3 `Das ist geplant.`; der Befund `Planung A1` steht in Zeile 3, wenn das Element trennt, sonst in Zeile 1.
func TestHandbuchLeereZeileTrenntAbsaetze(t *testing.T) {
	type el struct {
		z      string
		trennt bool
		cc     bool // Befund der Nr. 3 c in Zeile 2 (Kategorie Cc oder Cf)
	}
	var els []el
	// Leerraum, eine Zeile aus genau einem Zeichen: 24 Zeichen, die unicode.IsSpace trifft, ohne den Zeilenumbruch.
	hand := []rune{0x09, 0x0B, 0x0C, 0x0D, 0x20, 0x85, 0xA0, 0x1680}
	for r := rune(0x2000); r <= 0x200A; r++ {
		hand = append(hand, r)
	}
	hand = append(hand, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000)
	if len(hand) != 24 {
		t.Fatalf("%d Leerraum-Zeichen, erwartet 24", len(hand))
	}
	for _, r := range hand {
		els = append(els, el{string(r), true, r == 0x0B || r == 0x0C || r == 0x0D || r == 0x85})
	}
	els = append(els, el{" \t" + nbsp, true, false})
	els = append(els, el{nbsp + "x", false, false})
	for _, r := range []rune{0x180E, 0x200B, 0x2060, 0xFEFF} {
		els = append(els, el{string(r), false, true})
	}
	if len(els) != 30 {
		t.Fatalf("%d Elemente, erwartet 30", len(els))
	}
	l1 := strings.Count(rumpf(""), "\n") + 1
	for _, e := range els {
		m := ohneBeispiele(t)
		zeile := e.z + "\n"
		if e.z == "\r" {
			zeile = "\r\r\n" // ein CR vor dem LF gehört zum Zeilenumbruch
		}
		m.handbuch(rumpf("Ein Text.\n" + zeile + "Das ist geplant.\n"))
		code, _, errText := m.lauf()
		n := l1
		if e.trennt {
			n = l1 + 2
		}
		ist := befundZeilen(t, errText)
		want := 1
		if e.cc {
			want = 2
		}
		var a1 int
		for _, z := range ist {
			if z == fmt.Sprintf("Zeile %d: Planung A1: geplant", n) {
				a1++
			}
		}
		if code != exitBefund || len(ist) != want || a1 != 1 {
			t.Errorf("%q: Exit %d, Zeilen %q (erwartet %d, Planung A1 in Zeile %d)", e.z, code, ist, want, n)
		}
	}
}

// ordnungKleiner ist die Ordnung der Befunde nach SPEC-200 Nr. 2 d ohne den Gleichstand: Nummer, Zeile, Name.
func ordnungKleiner(a, b befund) bool {
	if a.nr != b.nr {
		return a.nr < b.nr
	}
	if a.zeile != b.zeile {
		return a.zeile < b.zeile
	}
	return a.name < b.name
}

// Z93 (Nr. 2 d): die Ordnung der Befunde ist stabil, auch bei umgeordneter Ausgabe mit mehr als 12 Befunden.
// Kopplung: sort.Slice ordnet bis 12 Elemente durch Einfügen stabil und erst ab 13 anders; die Schwelle und die
// Eingaben, die sort.Slice umordnet, gehören zur Go-Version des Images. Ändert Go das Verfahren, meldet die
// Zählprüfung `anders` statt dass die Mutation sort.Slice still grün bleibt. 27 Fälle U (Direktaufruf von
// meldeBefunde: Schlüssel der Gruppen {nr, zeile, name} mal Erzeugungsfolge {rr-ab, rr-auf, block-ab} mal G in
// {2, 3, 4} Gruppen mit R = 8 Befunden), 3 Fälle E (Ende zu Ende mit k in {2, 3, 6} Absätzen mal 44 Mustern), 1
// Fall K (Kopf-Gleichstand der Zeile 0 mit 16 Befunden der Nr. 5 dazwischen).
func TestHandbuchOrdnungStabil(t *testing.T) {
	n, anders := 0, 0
	for _, sl := range []string{"nr", "zeile", "name"} {
		for _, o := range []string{"rr-ab", "rr-auf", "block-ab"} {
			for g := 2; g <= 4; g++ {
				n++
				var liste []befund
				add := func(gr, r int) {
					b := befund{nr: 4, text: fmt.Sprintf("G%dR%d", gr, r)}
					switch sl {
					case "nr":
						b.nr = 4 + gr
					case "zeile":
						b.zeile = gr
					default:
						b.name = fmt.Sprintf("n%d", gr)
					}
					liste = append(liste, b)
				}
				switch o {
				case "rr-ab":
					for r := 1; r <= 8; r++ {
						for gr := g; gr >= 1; gr-- {
							add(gr, r)
						}
					}
				case "rr-auf":
					for r := 1; r <= 8; r++ {
						for gr := 1; gr <= g; gr++ {
							add(gr, r)
						}
					}
				default:
					for gr := g; gr >= 1; gr-- {
						for r := 1; r <= 8; r++ {
							add(gr, r)
						}
					}
				}
				var soll []string
				for gr := 1; gr <= g; gr++ {
					for r := 1; r <= 8; r++ {
						soll = append(soll, fmt.Sprintf("G%dR%d", gr, r))
					}
				}
				var buf bytes.Buffer
				meldeBefunde(append([]befund(nil), liste...), &bytes.Buffer{}, &buf)
				ist := befundZeilen(t, buf.String())
				if !gleichListe(ist, soll) {
					t.Errorf("%s/%s/G=%d: Ordnung %q, erwartet %q", sl, o, g, ist, soll)
				}
				cp := append([]befund(nil), liste...)
				sort.Slice(cp, func(i, j int) bool { return ordnungKleiner(cp[i], cp[j]) })
				var un []string
				for _, b := range cp {
					un = append(un, b.text)
				}
				if !gleichListe(un, soll) {
					anders++
				}
			}
		}
	}
	if n != 27 || anders != 27 {
		t.Errorf("U: %d Fälle (erwartet 27), davon %d, die sort.Slice umordnet (erwartet 27); Kopplung an die Go-Version", n, anders)
	}
	// E: ein Absatz je 44 Formen, k Absätze; erwartet Menge, Absatz, Muster in der Reihenfolge der Tabelle.
	or := orakel()
	if len(or) != 44 {
		t.Fatalf("%d Muster, erwartet 44", len(or))
	}
	for _, k := range []int{2, 3, 6} {
		formen := make([]string, len(or))
		for i, mu := range or {
			formen[i] = mu.formen[0]
		}
		if k == 3 {
			for i, j := 0, len(formen)-1; i < j; i, j = i+1, j-1 {
				formen[i], formen[j] = formen[j], formen[i]
			}
		}
		text := strings.Join(formen, " ; ") + "\n"
		kopf := rumpf("")
		h := kopf + strings.Repeat(text+"\n", k)
		m := ohneBeispiele(t)
		m.handbuch(h)
		code, _, errText := m.lauf()
		ist := befundZeilen(t, errText)
		var soll []string
		first := strings.Count(kopf, "\n") + 1
		for _, mg := range []string{"Planung", "Chronik", "Kennung"} {
			for a := 0; a < k; a++ {
				for _, mu := range or {
					if mengeName(mu.id) == mg {
						soll = append(soll, fmt.Sprintf("Zeile %d: %s %s: ", first+2*a, mg, mu.id))
					}
				}
			}
		}
		ok := code == exitBefund && len(ist) == len(soll) && len(soll) == 44*k
		for i := 0; ok && i < len(soll); i++ {
			ok = strings.HasPrefix(ist[i], soll[i])
		}
		if !ok {
			t.Errorf("E k=%d: Exit %d, %d Zeilen (erwartet %d)", k, code, len(ist), len(soll))
		}
	}
	// K: drei Befunde `Kopf: … fehlt`, 16 Befunde der Nr. 5, dann `version: Ausgabe nicht lesbar`; alle der Nummer 4 mit Zeile 0.
	m := neuesUmfeld(t, map[string]antwort{"version": {out: "x\n"}, "": {}})
	m.entferne("docs/user/beispiele")
	for i := 0; i < 16; i++ {
		m.schreibe(fmt.Sprintf("docs/user/beispiele/E%02d.txt", i), "x\n")
	}
	m.handbuch("# Handbuch\n\n## A\n\nText.\n")
	code, _, errText := m.lauf()
	ist := befundZeilen(t, errText)
	soll := []string{"Kopf: Software-Version fehlt", "Kopf: Gültigkeitsbereich fehlt", "Kopf: Stand fehlt", "version: Ausgabe nicht lesbar: "}
	ok := code == exitBefund && len(ist) == 4+16
	for i := 0; ok && i < 4; i++ {
		ok = strings.HasPrefix(ist[i], soll[i])
	}
	for i := 4; ok && i < len(ist); i++ {
		ok = strings.HasPrefix(ist[i], fmt.Sprintf("beispiele: E%02d.txt keine Beispiel-Datei", i-4))
	}
	if !ok {
		t.Errorf("K: Exit %d, Zeilen %q", code, ist)
	}
}
