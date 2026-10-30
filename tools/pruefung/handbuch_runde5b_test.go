package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests zu W-5 der zweiten Verifikation und zu den Mengen R1 bis R12 des Architect (SPEC-200 Nr. 2 c, 3, 4, 5, 6, 7, 9,
// 10 bis 13): jede Zusage läuft als Schleife über die Menge, die ihr Wortlaut meint; das Orakel steht im Test, von Hand
// gesetzt, und wo eine Funktion direkt aufrufbar ist, läuft sie ohne Prozess. Werte mit Rohbytes entstehen aus Bytewerten.

// Z65 (R1, Nr. 2 c): ein Eingabefehler schließt jeden Befund aus: nur die Zeile `Eingabe: …`, stdout leer. Menge: ein
// Handbuch mit Befunden der Nummern 3, 4 und 10 × {Binary fehlt, Binary nicht startbar, Arbeitsverzeichnis nicht anlegbar}
// (3), dazu die Paare Binary {kein Eintrag, Verzeichnis} × Handbuch {fehlt, Verzeichnis, kein UTF-8, NUL, BOM} (10): genau
// zwei Zeilen, `Binary fehlt` zuerst und kein Start des Ersatz-Binary.
func TestHandbuchEingabefehlerSchliesstBefundeAus(t *testing.T) {
	mitBefunden := "# H\n## A\nDas ist geplant.\nWort" + string(rune(0x200B)) + "Wort\n"
	m := neuesUmfeld(t, standardAntworten())
	m.handbuch(mitBefunden)
	n := 0
	fehlt := filepath.Join(m.wurzel, "gibt-es-nicht")
	code, out, errText := fuehre(m.u, "handbuch", fehlt, m.wurzel)
	if ist := befundZeilen(t, errText); code != exitAufruf || out != "" || !gleichListe(ist, []string{"Eingabe: Binary fehlt: " + fehlt}) {
		t.Fatalf("Binary fehlt: Exit %d, stdout %q, %q", code, out, ist)
	}
	n++
	if err := os.Chmod(filepath.Join(m.wurzel, "Makefile"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, out, errText = fuehre(m.u, "handbuch", filepath.Join(m.wurzel, "Makefile"), m.wurzel)
	if ist := befundZeilen(t, errText); code != exitAufruf || out != "" || len(ist) != 1 || !strings.HasPrefix(ist[0], "Eingabe: Binary nicht startbar: ") {
		t.Fatalf("Binary nicht startbar: Exit %d, stdout %q, %q", code, out, ist)
	}
	n++

	handbuecher := []struct {
		name  string
		bytes []byte
		dir   bool
		soll  string
	}{
		{"fehlt", nil, false, "Eingabe: Handbuch fehlt: "},
		{"Verzeichnis", nil, true, "Eingabe: Handbuch nicht lesbar: "},
		{"kein UTF-8", []byte{0xFF, '\n'}, false, "Eingabe: Handbuch nicht lesbar: kein UTF-8"},
		{"NUL", []byte{'a', 0x00, 'b', '\n'}, false, "Eingabe: Handbuch nicht lesbar: NUL"},
		{"BOM", []byte{0xEF, 0xBB, 0xBF, '#', ' ', 'H', '\n'}, false, "Eingabe: Handbuch nicht lesbar: BOM"},
	}
	for _, binaryIstVerz := range []bool{false, true} {
		for _, h := range handbuecher {
			w := neuesUmfeld(t, standardAntworten())
			w.entferne(handbuchPfad)
			switch {
			case h.dir:
				if err := os.MkdirAll(filepath.Join(w.wurzel, handbuchPfad), 0o755); err != nil {
					t.Fatal(err)
				}
			case h.bytes != nil:
				if err := os.WriteFile(filepath.Join(w.wurzel, handbuchPfad), h.bytes, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			bin := filepath.Join(w.wurzel, "gibt-es-nicht")
			if binaryIstVerz {
				bin = filepath.Join(w.wurzel, "docs")
			}
			code, out, errText := fuehre(w.u, "handbuch", bin, w.wurzel)
			ist := befundZeilen(t, errText)
			if code != exitAufruf || out != "" || len(ist) != 2 || ist[0] != "Eingabe: Binary fehlt: "+bin || !strings.HasPrefix(ist[1], h.soll) {
				t.Fatalf("Binary Verzeichnis %v, Handbuch %s: Exit %d, stdout %q, %q", binaryIstVerz, h.name, code, out, ist)
			}
			if s := len(w.e.starts(t)); s != 0 {
				t.Fatalf("Binary Verzeichnis %v, Handbuch %s: %d Starts, erwartet 0", binaryIstVerz, h.name, s)
			}
			n++
		}
	}

	a := neuesUmfeld(t, standardAntworten())
	a.handbuch(mitBefunden)
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "gibt-es-nicht"))
	code, out, errText = a.lauf()
	if ist := befundZeilen(t, errText); code != exitAufruf || out != "" || len(ist) != 1 || !strings.HasPrefix(ist[0], "Eingabe: Arbeitsverzeichnis nicht anlegbar: ") {
		t.Fatalf("Arbeitsverzeichnis: Exit %d, stdout %q, %q", code, out, ist)
	}
	n++
	if n != 13 {
		t.Fatalf("%d Elemente, erwartet 13", n)
	}
}

// Z66 (R2, Nr. 3 a): trifft mehr als ein Grund zu, gilt der erste in der Reihenfolge `kein UTF-8`, `NUL`, `BOM`.
// Menge: die vier Kombinationen (`kein UTF-8` mit NUL, mit BOM, NUL mit BOM, alle drei), die fünf ungültigen Folgen von Z21
// × Stelle (Anfang, Mitte, Ende), NUL × Stelle (3) und U+FEFF mitten im Text (kein Eingabefehler, Befund U+FEFF) = 23.
// Die Funktion `liesHandbuch` läuft direkt; die Bytes entstehen aus Bytewerten.
func TestHandbuchLesegruende(t *testing.T) {
	dir := t.TempDir()
	lies := func(b []byte) ([]string, string) {
		p := filepath.Join(dir, "h.md")
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return liesHandbuch(p)
	}
	bom := []byte{0xEF, 0xBB, 0xBF}
	mit := func(teile ...[]byte) []byte {
		var b []byte
		for _, t := range teile {
			b = append(b, t...)
		}
		return b
	}
	n := 0
	kombis := []struct {
		name string
		b    []byte
		soll string
	}{
		{"kein UTF-8 und NUL", []byte{0xFF, 0x00, '\n'}, "kein UTF-8"},
		{"kein UTF-8 und BOM", mit(bom, []byte{0xFF, '\n'}), "kein UTF-8"},
		{"NUL und BOM", mit(bom, []byte{0x00, '\n'}), "NUL"},
		{"alle drei", mit(bom, []byte{0xFF, 0x00, '\n'}), "kein UTF-8"},
	}
	for _, k := range kombis {
		if _, f := lies(k.b); f != "Handbuch nicht lesbar: "+k.soll {
			t.Fatalf("%s: %q, erwartet Grund %q", k.name, f, k.soll)
		}
		n++
	}
	stellen := func(f []byte) map[string][]byte {
		return map[string][]byte{
			"Anfang": mit(f, []byte("a\nb\n")),
			"Mitte":  mit([]byte("a\n"), f, []byte("b\n")),
			"Ende":   mit([]byte("a\nb\n"), f),
		}
	}
	folgen := [][]byte{{0xFF}, {0xC0, 0x80}, {0xED, 0xA0, 0x80}, {0xF4, 0x90, 0x80, 0x80}, {0xC3}}
	for _, f := range folgen {
		for stelle, b := range stellen(f) {
			if _, g := lies(b); g != "Handbuch nicht lesbar: kein UTF-8" {
				t.Fatalf("% X an der Stelle %s: %q", f, stelle, g)
			}
			n++
		}
	}
	for stelle, b := range stellen([]byte{0x00}) {
		if _, g := lies(b); g != "Handbuch nicht lesbar: NUL" {
			t.Fatalf("NUL an der Stelle %s: %q", stelle, g)
		}
		n++
	}
	zeilen, f := lies(mit([]byte("a"), bom, []byte("b\n")))
	var bf []string
	for _, b := range zeichenBefunde(zeilen) {
		bf = append(bf, b.text)
	}
	if f != "" || !gleichListe(bf, []string{"Zeile 1: Zeichen U+FEFF nicht zulässig"}) {
		t.Fatalf("U+FEFF mitten im Text: Fehler %q, Befunde %q", f, bf)
	}
	n++
	if n != 23 {
		t.Fatalf("%d Elemente, erwartet 23", n)
	}
}

// Z68 (R4, Nr. 3 c): jedes Zeichen der Kategorie Cc außer Tabulator und `\n` ist ein Befund: U+0001 bis U+001F ohne U+0009
// und U+000A (29) und U+007F bis U+009F (33) = 62, je ein Befund in `Wort<c>Wort`. Gutfälle (20, kein Befund): Tabulator, die
// 17 Zeichen der Kategorie Zs (U+0020, U+00A0, U+1680, U+2000 bis U+200A, U+202F, U+205F, U+3000), U+2028 und U+2029. Die
// Funktion `zeichenBefunde` läuft direkt.
func TestHandbuchSteuerzeichenAlle(t *testing.T) {
	var cc []rune
	for r := rune(0x01); r <= 0x1F; r++ {
		if r != '\t' && r != '\n' {
			cc = append(cc, r)
		}
	}
	for r := rune(0x7F); r <= 0x9F; r++ {
		cc = append(cc, r)
	}
	if len(cc) != 62 {
		t.Fatalf("%d Zeichen der Kategorie Cc, erwartet 62", len(cc))
	}
	for _, r := range cc {
		var ist []string
		for _, b := range zeichenBefunde([]string{"Wort" + string(r) + "Wort"}) {
			ist = append(ist, b.text)
		}
		if soll := fmt.Sprintf("Zeile 1: Zeichen U+%04X nicht zulässig", r); !gleichListe(ist, []string{soll}) {
			t.Fatalf("U+%04X: %q, erwartet %q", r, ist, soll)
		}
	}
	gut := []rune{'\t', ' ', 0xA0, 0x1680, 0x202F, 0x205F, 0x3000, 0x2028, 0x2029}
	for r := rune(0x2000); r <= 0x200A; r++ {
		gut = append(gut, r)
	}
	if len(gut) != 20 {
		t.Fatalf("%d Gutfälle, erwartet 20", len(gut))
	}
	for _, r := range gut {
		if bf := zeichenBefunde([]string{"Wort" + string(r) + "Wort"}); len(bf) != 0 {
			t.Fatalf("U+%04X: %d Befunde, erwartet keinen", r, len(bf))
		}
	}
}

// Z69 (R5, Nr. 4 a: „getrennt am ersten `:`“): vier Elemente. `Gültigkeitsbereich: Binary: x` (Wert `Binary: x`, kein
// Befund); `Software-Version: 1:0` mit `version` = `1:0` (gleich, kein Befund); `Stand: 2026-10-30:` (Befund `Stand ungültig:
// 2026-10-30:`); `a:b: c` in Zeile 6 (Befund `Schlüssel a unbekannt`).
func TestHandbuchErsterDoppelpunkt(t *testing.T) {
	m := ohneBeispiele(t)
	m.handbuch("# H\n\nSoftware-Version: dev\nGültigkeitsbereich: Binary: x\nStand: 2026-10-30\n\n## A\n")
	if code, _, errText := m.lauf(); code != exitOK {
		t.Fatalf("Wert mit Doppelpunkt: Exit %d %q", code, errText)
	}
	v := neuesUmfeld(t, map[string]antwort{"version": {out: "odata-recorder 1:0\n"}})
	v.entferne("docs/user/beispiele")
	v.handbuch("# H\n\nSoftware-Version: 1:0\nGültigkeitsbereich: Binary\nStand: 2026-10-30\n\n## A\n")
	if code, _, errText := v.lauf(); code != exitOK {
		t.Fatalf("Version mit Doppelpunkt: Exit %d %q", code, errText)
	}
	m.handbuch("# H\n\nSoftware-Version: dev\nGültigkeitsbereich: Binary\nStand: 2026-10-30:\n\n## A\n")
	code, _, errText := m.lauf()
	if ist := befundZeilen(t, errText); code != exitBefund || !gleichListe(ist, []string{"Kopf: Stand ungültig: 2026-10-30:"}) {
		t.Fatalf("Stand mit Doppelpunkt: Exit %d, %q", code, ist)
	}
	m.handbuch("# H\n\nSoftware-Version: dev\nGültigkeitsbereich: Binary\nStand: 2026-10-30\na:b: c\n\n## A\n")
	code, _, errText = m.lauf()
	if ist := befundZeilen(t, errText); code != exitBefund || !gleichListe(ist, []string{"Zeile 6: Kopf: Schlüssel a unbekannt"}) {
		t.Fatalf("Schlüssel a: Exit %d, %q", code, ist)
	}
}

// Z70 (R6, Nr. 5 a): Zaunformen, 15 Elemente. Einrückung {4 Leerzeichen, Tabulator, NBSP} vor drei Backticks (3) und ein
// Backtick in der Info {```a`, ``` `, ```x`y} (3) sind kein Zaun (der Inhalt ist Prosa: Befund A1, nie ein Befund des
// Zaunes); acht Schlusszeilen schließen nicht ({1 Leerzeichen, 2 Leerzeichen, Tabulator} vor drei Backticks, vier
// Backticks, ``` x```, ```x```, drei Tilden, zwei Backticks): `Zaun nicht geschlossen` in Zeile 9; ein Zaun in Zeile 1:
// `Zeile 1: Beispiel ohne Bindungszeile`.
func TestHandbuchZaunformenRand(t *testing.T) {
	m := ohneBeispiele(t)
	keinZaun := func(h string) {
		t.Helper()
		m.handbuch(rumpf(h))
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, trefferZeile("A1")+"geplant")
		nichtEnthalten(t, errText, "Zaun", "Bindungszeile", "Info-Zeichenkette")
	}
	n := 0
	for _, e := range []string{"    ", "\t", string(rune(0xA0))} {
		keinZaun(e + "```\ngeplant\n" + e + "```\n")
		n++
	}
	for _, o := range []string{"```a`", "``` `", "```x`y"} {
		keinZaun(o + "\ngeplant\ntext\n")
		n++
	}
	for _, s := range []string{" ```", "  ```", "\t```", "````", "``` x```", "```x```", "~~~", "``"} {
		m.handbuch(rumpf("```\nbisher\n" + s + "\nweiter\n"))
		code, _, errText := m.lauf()
		erwarte(t, code, errText, exitBefund, "Zeile 9: Zaun nicht geschlossen")
		n++
	}
	m.handbuch("```\nx\n```\n")
	code, _, errText := m.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 1: Beispiel ohne Bindungszeile")
	n++
	if n != 15 {
		t.Fatalf("%d Elemente, erwartet 15", n)
	}
}

// Z71 (R7, Nr. 5 f): ein Symbolischer Verweis in docs/user/beispiele ist keine Beispiel-Datei, auch wenn er auf eine
// Datei zeigt: `beispiele: a.txt keine Beispiel-Datei`, nie `ohne Bindung`.
func TestHandbuchBeispielVerweis(t *testing.T) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele/aufruf.txt")
	m.schreibe("docs/user/ziel.dat", aufrufDatei)
	if err := os.Symlink(filepath.Join(m.wurzel, "docs/user/ziel.dat"), filepath.Join(m.wurzel, "docs/user/beispiele/a.txt")); err != nil {
		t.Skipf("kein Symbolischer Verweis erzeugbar: %v", err)
	}
	m.handbuch(rumpf("Text.\n"))
	code, _, errText := m.lauf()
	if ist := befundZeilen(t, errText); code != exitBefund || !gleichListe(ist, []string{"beispiele: a.txt keine Beispiel-Datei"}) {
		t.Fatalf("Exit %d, %q", code, ist)
	}
}

// Z72 (R8, Nr. 7 d: „bytegleich“, je Strom): die Datei erwartet `a\n`; erhalten {`a`, `a\n\n`, `a\r\n`, `a \n`, `A\n`, leer}
// (6), dazu die Datei leer und erhalten `\n` (1), je Strom (Standardausgabe, Fehlerausgabe) = 14. Orakel von Hand: genau ein
// Befund, *q* in Go-Schreibweise gequotet.
func TestHandbuchStromBytegleich(t *testing.T) {
	faelle := []struct {
		erwartet []string
		qErw     string
		erhalten string
		qErh     string
	}{
		{[]string{"a"}, `"a\n"`, "a", `"a"`},
		{[]string{"a"}, `"a\n"`, "a\n\n", `"a\n\n"`},
		{[]string{"a"}, `"a\n"`, "a\r\n", `"a\r\n"`},
		{[]string{"a"}, `"a\n"`, "a \n", `"a \n"`},
		{[]string{"a"}, `"a\n"`, "A\n", `"A\n"`},
		{[]string{"a"}, `"a\n"`, "", `""`},
		{nil, `""`, "\n", `"\n"`},
	}
	n := 0
	for _, strom := range []string{"Standardausgabe", "Fehlerausgabe"} {
		for _, f := range faelle {
			var m umfeld
			if strom == "Standardausgabe" {
				m = bspMit(t, nil, f.erwartet, nil, "0\n", antwort{out: f.erhalten})
			} else {
				m = bspMit(t, nil, nil, f.erwartet, "0\n", antwort{err: f.erhalten})
			}
			code, _, errText := m.lauf()
			soll := fmt.Sprintf("k.txt: %s weicht ab (erwartet %s, erhalten %s)", strom, f.qErw, f.qErh)
			if ist := befundZeilen(t, errText); code != exitBefund || !gleichListe(ist, []string{soll}) {
				t.Fatalf("%s, erhalten %q: Exit %d, %q, erwartet %q", strom, f.erhalten, code, ist, soll)
			}
			n++
		}
	}
	if n != 14 {
		t.Fatalf("%d Elemente, erwartet 14", n)
	}
}

// Z73 (R10, Nr. 10 bis 12, 13 a: „in der Reihenfolge der Tabelle“): die 43 benachbarten Paare (i, i+1) der 44 Muster. Je
// Paar steht die erste Form von i + 1 im Text vor der ersten Form von i; die Meldung für i steht dennoch vor der für i + 1.
// Die Funktion `mengenBefunde` läuft direkt, die Reihenfolge schafft `meldeBefunde`.
func TestHandbuchMusterReihenfolge(t *testing.T) {
	ms := orakel()
	if len(ms) != 44 {
		t.Fatalf("%d Muster, erwartet 44", len(ms))
	}
	for i := 0; i+1 < len(ms); i++ {
		prosa := ms[i+1].formen[0] + " und " + ms[i].formen[0]
		ist := geordnet(mengenBefunde([]absatz{{zeile: 9, zeilen: []string{prosa}}}))
		pos := func(id string) int {
			for k, z := range ist {
				if strings.HasPrefix(z, "Zeile 9"+trefferZeile(id)) {
					return k
				}
			}
			return -1
		}
		a, b := pos(ms[i].id), pos(ms[i+1].id)
		if a < 0 || b < 0 || a >= b {
			t.Fatalf("%s vor %s: Stellen %d und %d in %q", ms[i].id, ms[i+1].id, a, b, ist)
		}
	}
}

// Z74 (R11, Nr. 13 a): `*`, `_` und Backtick werden entfernt, auch mitten im Wort. Menge (6): `ge*plant`, `ge` Backtick
// `plant`, `ge_plant`, `*geplant`, `geplant*`, `geplant_` treffen alle A1. Die Funktion `mengenBefunde` läuft direkt.
func TestHandbuchMarkupMittenImWort(t *testing.T) {
	formen := []string{"ge*plant", "ge" + "`" + "plant", "ge_plant", "*geplant", "geplant*", "geplant_"}
	for _, f := range formen {
		ist := geordnet(mengenBefunde([]absatz{{zeile: 9, zeilen: []string{f}}}))
		if !gleichListe(ist, []string{"Zeile 9: Planung A1: geplant"}) {
			t.Fatalf("%q: %q", f, ist)
		}
	}
}

// Z75 (R12, Nr. 13 d: *n* ist die erste Zeile des Absatzes, im Zitat die Zeile des Zaunes): sechs Elemente. Ein Absatz
// unmittelbar hinter einem gebundenen Zaun (Zeile 18), mit Leerzeile dazwischen (19), unmittelbar vor der Bindungszeile
// (Absatz beginnt in Zeile 9), hinter einem ungebundenen Zaun (Absatz beginnt mit dem Zaun in Zeile 9), im Kopf
// (Zeile 3), im Zitat (Zeile 10 des Zaunes).
func TestHandbuchAbsatzZeile(t *testing.T) {
	mit := neuesUmfeld(t, standardAntworten())
	zaun := binde("beispiele/aufruf.txt") + zaunVon(aufrufDatei)
	mit.handbuch(rumpf(zaun + "geplant\n"))
	code, _, errText := mit.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 18: Planung A1: geplant")
	mit.handbuch(rumpf(zaun + "\ngeplant\n"))
	code, _, errText = mit.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 19: Planung A1: geplant")
	mit.handbuch(rumpf("geplant\n" + zaun))
	code, _, errText = mit.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 9: Planung A1: geplant")
	ohne := ohneBeispiele(t)
	ohne.handbuch(rumpf("```\nx\n```\ngeplant\n"))
	code, _, errText = ohne.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 9: Planung A1: geplant")
	ohne.handbuch("# H\n\nSoftware-Version: dev\nGültigkeitsbereich: geplant\nStand: 2026-10-30\n\n## A\n")
	code, _, errText = ohne.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 3: Planung A1: geplant")
	zm, _ := bspZitat(t, "roadmap")
	code, _, errText = zm.lauf()
	erwarte(t, code, errText, exitBefund, "Zeile 10: Kennung C1: roadmap")
}

// Z76 (n03, Nr. 4 d: Gregorianischer Kalender): die Jahrhundertregel. Menge: die Jahre 1600 bis 2800 in Hunderterschritten
// (13); Orakel von Hand: `-02-29` gilt genau in 1600, 2000, 2400 und 2800, `-02-28` gilt in allen. Die Funktion läuft direkt.
// Die Jahre außerhalb der Hunderter stehen als akzeptiertes Negativ N7 im Plan.
func TestHandbuchStandJahrhundert(t *testing.T) {
	n := 0
	for j := 1600; j <= 2800; j += 100 {
		schalt := j == 1600 || j == 2000 || j == 2400 || j == 2800
		if standGueltig(fmt.Sprintf("%d-02-29", j)) != schalt {
			t.Fatalf("%d-02-29: gültig %v, erwartet %v", j, !schalt, schalt)
		}
		if !standGueltig(fmt.Sprintf("%d-02-28", j)) {
			t.Fatalf("%d-02-28 ist gültig", j)
		}
		n++
	}
	if n != 13 {
		t.Fatalf("%d Jahre, erwartet 13", n)
	}
}

// zieleVon liefert die Zielliste einer Datei rel mit den Zeilen zeilen.
func zieleVon(t *testing.T, rel string, zeilen []string) map[string]bool {
	t.Helper()
	m := umfeld{t: t, wurzel: t.TempDir()}
	m.schreibe(rel, strings.Join(zeilen, "\n")+"\n")
	return zielliste(m.wurzel)
}

func gleichMenge(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// Z77 (n21, R9, Nr. 9 a: `^[a-z0-9-]+:.*##`): die Zielliste. Menge: eine Zeile `a<c>b: ## x` für jedes Zeichen c von
// U+0020 bis U+007E außer `:` (94); das Ziel `a<c>b` gilt genau, wenn c in `a-z`, `0-9` oder `-` liegt (37 Ziele). Dazu
// 11 Zeilenformen: `p1: ## a`, `p2:## a`, `p3: a ## b`, `p4:: ## a`, `p5: ##` gelten; ` q1: ## a`, Tabulator vor `q2`,
// `q3 : ## a`, `q4 ## a:`, `q5: a`, `q6: #` gelten nicht. Die Funktion `zielliste` läuft direkt.
func TestHandbuchZiellisteZeilenformen(t *testing.T) {
	var zeilen []string
	soll := map[string]bool{}
	n := 0
	for c := 0x20; c <= 0x7E; c++ {
		if c == ':' {
			continue
		}
		zeilen = append(zeilen, "a"+string(rune(c))+"b: ## x")
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' {
			soll["a"+string(rune(c))+"b"] = true
		}
		n++
	}
	if n != 94 || len(soll) != 37 {
		t.Fatalf("%d Zeichen und %d Ziele, erwartet 94 und 37", n, len(soll))
	}
	if ist := zieleVon(t, "Makefile", zeilen); !gleichMenge(ist, soll) {
		t.Fatalf("Zeichen im Namen: %v, erwartet %v", ist, soll)
	}
	formen := []string{"p1: ## a", "p2:## a", "p3: a ## b", "p4:: ## a", "p5: ##",
		" q1: ## a", "\tq2: ## a", "q3 : ## a", "q4 ## a:", "q5: a", "q6: #"}
	gelten := map[string]bool{"p1": true, "p2": true, "p3": true, "p4": true, "p5": true}
	if ist := zieleVon(t, "Makefile", formen); !gleichMenge(ist, gelten) {
		t.Fatalf("Zeilenformen: %v, erwartet %v", ist, gelten)
	}
}

// Z78 (n22, Nr. 9 a: „Dateien mit der Endung `.mk`, nicht rekursiv“): in harness/mk liest die Zielliste `x.mk` (t1) und
// keine der acht anderen Formen: `x.mk.bak` (t2), `x.mk~` (t3), `x.mkx` (t4), `x.MK` (t5), `xmk` (t6), `x.txt` (t7), ein
// Verzeichnis `y.mk/` mit `z.mk` darin (t8). Die Funktion `zielliste` läuft direkt.
func TestHandbuchZiellisteMkDateien(t *testing.T) {
	m := umfeld{t: t, wurzel: t.TempDir()}
	dateien := []string{"x.mk", "x.mk.bak", "x.mk~", "x.mkx", "x.MK", "xmk", "x.txt", "y.mk/z.mk"}
	for i, d := range dateien {
		m.schreibe("harness/mk/"+d, fmt.Sprintf("t%d: ## x\n", i+1))
	}
	if ist := zielliste(m.wurzel); !gleichMenge(ist, map[string]bool{"t1": true}) {
		t.Fatalf("Ziele %v, erwartet nur t1", ist)
	}
}

// Z79 (n27, Nr. 6 a: „genau ein Leerzeichen“): jedes der 19 Leerraum-Zeichen (`leerraum19`) hinter dem Präfix (19) und
// zwischen zwei Argumenten (19) ist eine ungültige Kopfzeile; dazu zwei Gutfälle (`<Präfix> a b`, das Präfix allein).
// Die Funktion `kopfzeile` läuft direkt.
func TestHandbuchKopfzeileTrenner(t *testing.T) {
	n := 0
	for _, sep := range leerraum19() {
		for _, z := range []string{aufrufPraefix + sep + "arg", aufrufPraefix + " a" + sep + "b"} {
			if args, grund := kopfzeile(z); grund == "" {
				t.Fatalf("%q ist gültig (Argumente %q)", z, args)
			}
			n++
		}
	}
	if n != 38 {
		t.Fatalf("%d Elemente, erwartet 38", n)
	}
	if args, grund := kopfzeile(aufrufPraefix + " a b"); grund != "" || strings.Join(args, "|") != "a|b" {
		t.Fatalf("Gutfall: Argumente %q, Grund %q", args, grund)
	}
	if args, grund := kopfzeile(aufrufPraefix); grund != "" || len(args) != 0 {
		t.Fatalf("Präfix allein: Argumente %q, Grund %q", args, grund)
	}
}

// Z80 (n31, Nr. 4 a: „beginnt mit `## `“): jedes der 19 Leerraum-Zeichen (`leerraum19`) vor `## T` beendet den Kopf nicht;
// der Schlüssel dahinter wird gelesen. Die Funktionen `kopfEnde` und `leseKopf` laufen direkt.
func TestHandbuchKopfEndeLeerraumVorher(t *testing.T) {
	n := 0
	for _, sep := range leerraum19() {
		zeilen := []string{"# H", "", sep + "## T", "Software-Version: dev", "## A"}
		d := zerlege(zeilen)
		if e := kopfEnde(d); e != 4 {
			t.Fatalf("%q vor ## T: der Kopf endet bei Index %d, erwartet 4", sep, e)
		}
		if k := leseKopf(d); !k.vergleich || k.version != "dev" {
			t.Fatalf("%q vor ## T: Schlüssel dahinter nicht gelesen (%+v)", sep, k)
		}
		n++
	}
	if n != 19 {
		t.Fatalf("%d Elemente, erwartet 19", n)
	}
}

// Z81 (n32, Nr. 4 a: „ein anderer Schlüssel … ist `unbekannt`“): Schlüssel mit 14 Anfangszeichen-Klassen (`#`, `-`, `*`, `>`,
// `|`, `[`, `1`, `"`, `(`, `A`, `a`, `Ä`, `ä`, `x y`) und drei Schreibweisen von `Software-Version` (`software-version`,
// `SOFTWARE-VERSION`, `Software Version`) = 17. Orakel: genau `Zeile 3: Kopf: Schlüssel <k> unbekannt`. Die Funktion
// `leseKopfZeilen` läuft direkt; die Prüfgrenze ist N8.
func TestHandbuchKopfSchluesselAnfang(t *testing.T) {
	schluessel := []string{"#", "-", "*", ">", "|", "[", "1", "\"", "(", "A", "a", string(rune(0xC4)), string(rune(0xE4)), "x y",
		"software-version", "SOFTWARE-VERSION", "Software Version"} // 14 Klassen und 3 Schreibweisen
	if len(schluessel) != 17 {
		t.Fatalf("%d Schlüssel, erwartet 17", len(schluessel))
	}
	for _, k := range schluessel {
		werte, zeilen := map[string][]string{}, map[string]int{}
		bf := leseKopfZeilen([]string{"# H", "", k + ": x"}, werte, zeilen)
		if ist := geordnet(bf); !gleichListe(ist, []string{"Zeile 3: Kopf: Schlüssel " + k + " unbekannt"}) {
			t.Fatalf("Schlüssel %q: %q", k, ist)
		}
	}
}

// Z82 (n35, Nr. 9 b: „ein Leerzeichen (U+0020)“): ein Span `<c>make nix<c>` mit c in den 19 Leerraum-Zeichen
// (`leerraum19`) behält den Rand, beginnt nicht mit `make ` und ist kein Aufruf: kein Befund. Die Kontrolle ist der Span mit
// Leerzeichen (`make nix`, unbekanntes Ziel). Die Funktion `makeBefunde` läuft direkt.
func TestHandbuchSpanRandLeerraum(t *testing.T) {
	ziele := map[string]bool{}
	pruefe := func(c string) []string {
		return geordnet(makeBefunde([]absatz{{zeile: 9, zeilen: []string{"`" + c + "make nix" + c + "`"}}}, ziele))
	}
	if ist := pruefe(" "); !gleichListe(ist, []string{"Zeile 9: make-Ziel nix unbekannt"}) {
		t.Fatalf("Kontrolle: %q", ist)
	}
	for _, c := range leerraum19() {
		if ist := pruefe(c); len(ist) != 0 {
			t.Fatalf("Rand %q: %q, erwartet kein Befund", c, ist)
		}
	}
}
