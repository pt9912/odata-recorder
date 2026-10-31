package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// Das Orakel der Menge Z steht hier von Hand, aus dem Wortlaut von SPEC-233 Nr. 7, ohne die
// Tabellen von unicode: Bereiche mit ihrer Kategorie. Zeichen entstehen aus Zahlen (AGENTS.md
// §3.18), nie als Rohzeichen im Quelltext.
type zBereich struct {
	von, bis rune
	kat      string
}

func zOrakel() []zBereich {
	return []zBereich{
		{0x0000, 0x0008, "Cc"}, {0x000B, 0x001F, "Cc"}, {0x007F, 0x009F, "Cc"},
		{0x00A0, 0x00A0, "Zs"}, {0x00AD, 0x00AD, "Cf"}, {0x0600, 0x0605, "Cf"}, {0x061C, 0x061C, "Cf"},
		{0x06DD, 0x06DD, "Cf"}, {0x070F, 0x070F, "Cf"}, {0x0890, 0x0891, "Cf"}, {0x08E2, 0x08E2, "Cf"},
		{0x1680, 0x1680, "Zs"}, {0x180E, 0x180E, "Cf"}, {0x2000, 0x200A, "Zs"}, {0x200B, 0x200F, "Cf"},
		{0x2028, 0x2028, "Zl"}, {0x2029, 0x2029, "Zp"}, {0x202A, 0x202E, "Cf"}, {0x202F, 0x202F, "Zs"},
		{0x205F, 0x205F, "Zs"}, {0x2060, 0x2064, "Cf"}, {0x2066, 0x206F, "Cf"}, {0x3000, 0x3000, "Zs"},
		{0xFEFF, 0xFEFF, "Cf"}, {0xFFF9, 0xFFFB, "Cf"}, {0xFFFE, 0xFFFF, "Cn"}, {0x110BD, 0x110BD, "Cf"},
		{0x110CD, 0x110CD, "Cf"}, {0x13430, 0x1343F, "Cf"}, {0x1BCA0, 0x1BCA3, "Cf"},
		{0x1D173, 0x1D17A, "Cf"}, {0xE0001, 0xE0001, "Cf"}, {0xE0020, 0xE007F, "Cf"},
	}
}

// orakelKat ist die Kategorie von r nach dem Orakel oder "".
func orakelKat(r rune) string {
	for _, b := range zOrakel() {
		if r >= b.von && r <= b.bis {
			return b.kat
		}
	}
	return ""
}

// zElemente sind die 253 Elemente von Z in der Folge der Codepunkte.
func zElemente() []rune {
	var r []rune
	for _, b := range zOrakel() {
		for c := b.von; c <= b.bis; c++ {
			r = append(r, c)
		}
	}
	return r
}

// kurznamenOrakel sind die 29 Zeichen mit Kurzname aus SPEC-233 Nr. 7 f.
func kurznamenOrakel() map[rune]string {
	return map[rune]string{
		0x0000: "NUL", 0x000D: "CR", 0x007F: "DEL", 0x0085: "NEL", 0x00A0: "NBSP", 0x00AD: "SHY",
		0x061C: "ALM", 0x180E: "MVS", 0x200B: "ZWSP", 0x200C: "ZWNJ", 0x200D: "ZWJ", 0x200E: "LRM",
		0x200F: "RLM", 0x2028: "LS", 0x2029: "PS", 0x202A: "LRE", 0x202B: "RLE", 0x202C: "PDF",
		0x202D: "LRO", 0x202E: "RLO", 0x202F: "NNBSP", 0x2060: "WJ", 0x2066: "LRI", 0x2067: "RLI",
		0x2068: "FSI", 0x2069: "PDI", 0xFEFF: "BOM", 0xFFFE: "NONCHAR", 0xFFFF: "NONCHAR",
	}
}

// orakelZeile ist die Fundstelle eines Zeichens von Z nach SPEC-233 Nr. 10 a.
func orakelZeile(pfad string, zeile, spalte int, r rune) string {
	t := fmt.Sprintf("zeichen: %s:%d:%d: U+%04X %s", pfad, zeile, spalte, r, orakelKat(r))
	if k := kurznamenOrakel()[r]; k != "" {
		t += " " + k
	}
	return t + "\n"
}

func zb(r rune) []byte { return utf8.AppendRune(nil, r) }

func cat(teile ...[]byte) []byte { return bytes.Join(teile, nil) }

func b(s string) []byte { return []byte(s) }

type ergebnis struct {
	code     int
	out, err string
}

func fahre(wurzel string) ergebnis {
	var o, e bytes.Buffer
	c := run([]string{"pruefen", wurzel}, &o, &e)
	return ergebnis{c, o.String(), e.String()}
}

// baum legt die Dateien (Pfad mit "/", Inhalt) in einem Wegwerfverzeichnis an.
func baum(t *testing.T, dateien map[string][]byte) string {
	t.Helper()
	dir := t.TempDir()
	for p, inhalt := range dateien {
		ziel := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(ziel), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ziel, inhalt, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func gleich(t *testing.T, was, ist, soll string) {
	t.Helper()
	if ist != soll {
		t.Errorf("%s: %q statt %q", was, ist, soll)
	}
}

func okZeile(n, bin, s int) string {
	return fmt.Sprintf("zeichen: ok, %d Datei(en) geprüft, %d Binärdatei(en) und %d andere Einträge übergangen\n", n, bin, s)
}

func summe(k, d int) string { return fmt.Sprintf("zeichen: %d Befund(e) in %d Datei(en)\n", k, d) }

func TestMengeZ(t *testing.T) {
	if unicode.Version != "17.0.0" {
		t.Fatalf("unicode.Version %s statt 17.0.0 (SPEC-233 Nr. 7 b)", unicode.Version)
	}
	zahl := map[string]int{}
	skalare := 0
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		skalare++
		soll, ist := orakelKat(r), kategorie(r)
		if soll != ist {
			t.Fatalf("U+%04X: Kategorie %q statt %q", r, ist, soll)
		}
		zahl[ist]++
	}
	if skalare != 1112064 {
		t.Errorf("%d Skalarwerte statt 1112064", skalare)
	}
	for kat, n := range map[string]int{"Cc": 63, "Cf": 170, "Zl": 1, "Zp": 1, "Zs": 16, "Cn": 2, "": 1112064 - 253} {
		if zahl[kat] != n {
			t.Errorf("Kategorie %q: %d Zeichen statt %d", kat, zahl[kat], n)
		}
	}
	if len(zElemente()) != 253 {
		t.Errorf("Orakel: %d Elemente statt 253", len(zElemente()))
	}
	anker := []rune{0xFEFF, 0x00A0, 0x2028, 0x2029, 0x200B, 0x0000, 0x0001, 0x001F, 0x007F, 0x0080, 0x009F, 0x000D, 0x000B, 0x000C}
	for _, r := range anker {
		if kategorie(r) == "" {
			t.Errorf("U+%04X fehlt in Z", r)
		}
	}
	for _, r := range []rune{'\t', '\n', ' '} {
		if kategorie(r) != "" {
			t.Errorf("U+%04X liegt in Z", r)
		}
	}
}

func TestKurznamen(t *testing.T) {
	soll := kurznamenOrakel()
	if len(soll) != 29 {
		t.Fatalf("Orakel: %d Kurznamen statt 29", len(soll))
	}
	namen := map[string]bool{}
	for r, n := range soll {
		namen[n] = true
		if kurzname(r) != n {
			t.Errorf("U+%04X: Kurzname %q statt %q", r, kurzname(r), n)
		}
	}
	if len(namen) != 28 {
		t.Errorf("%d verschiedene Namen statt 28", len(namen))
	}
	tabellen := map[string]*unicode.RangeTable{"Cc": unicode.Cc, "Cf": unicode.Cf, "Zl": unicode.Zl, "Zp": unicode.Zp, "Zs": unicode.Zs}
	zahl := 0
	for _, r := range zElemente() {
		zahl++
		kat := kategorie(r)
		if tab, ok := tabellen[kat]; ok && !unicode.Is(tab, r) {
			t.Errorf("U+%04X: Kategorie %s nicht in der Tabelle von unicode", r, kat)
		}
		text := befundText(r, kat)
		gleich(t, fmt.Sprintf("U+%04X", r), text, strings.TrimSuffix(orakelZeile("", 0, 0, r)[len("zeichen: :0:0: "):], "\n"))
		for i := 0; i < len(text); i++ {
			if text[i] < 0x20 || text[i] > 0x7E {
				t.Errorf("U+%04X: Byte 0x%02X jenseits von ASCII im Befund-Text", r, text[i])
			}
		}
		if _, hat := soll[r]; !hat && kurzname(r) != "" {
			t.Errorf("U+%04X: Kurzname %q ohne Eintrag", r, kurzname(r))
		}
	}
	if zahl != 253 {
		t.Errorf("%d Elemente statt 253", zahl)
	}
}

func TestAusgabeAscii(t *testing.T) {
	const hex = "0123456789ABCDEF"
	for v := 0; v < 256; v++ {
		c := byte(v)
		soll := string([]byte{c})
		switch {
		case c == '\\':
			soll = `\\`
		case c < 0x20 || c > 0x7E:
			soll = `\x` + string([]byte{hex[c>>4], hex[c&15]})
		}
		gleich(t, fmt.Sprintf("Byte 0x%02X", c), maskiere("a"+string([]byte{c})+"b"), "a"+soll+"b")
	}
	// Eine Datei je Namensbyte (außer NUL und dem Schrägstrich, die kein Name tragen kann): jede
	// Befund-Zeile besteht aus druckbarem ASCII und dem Zeilenende.
	dateien := map[string][]byte{}
	for v := 1; v < 256; v++ {
		if v != '/' {
			dateien["p"+string([]byte{byte(v)})+"q.md"] = cat(b("x"), zb(0x00A0))
		}
	}
	r := fahre(baum(t, dateien))
	if r.code != 1 || r.out != "" {
		t.Fatalf("Exit %d, stdout %q", r.code, r.out)
	}
	zeilen := strings.Split(strings.TrimSuffix(r.err, "\n"), "\n")
	if len(zeilen) != 255 {
		t.Fatalf("%d Zeilen statt 254 Befunden und der Summe", len(zeilen))
	}
	for _, z := range zeilen {
		for i := 0; i < len(z); i++ {
			if z[i] < 0x20 || z[i] > 0x7E {
				t.Fatalf("Byte 0x%02X in der Zeile %q", z[i], z)
			}
		}
	}
	gleich(t, "letzte Zeile", zeilen[254]+"\n", summe(254, 254))
}
