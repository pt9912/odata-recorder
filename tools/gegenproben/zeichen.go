package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Fälle 131 bis 150 (SPEC-102, SPEC-233): `zeichen pruefen <Wurzel>` im Wegwerfverzeichnis des Falls.
// Der Inhalt jeder Datei entsteht aus Bytewerten (\x-Folgen und Codepunkten), nie aus Rohzeichen oder
// einer Folge aus Backslash, u und vier Hex-Ziffern (AGENTS.md §3.18).

const (
	zcNBSP     = "\xc2\xa0"
	zcBOM      = "\xef\xbb\xbf"
	zcAufruf   = "Aufruf: zeichen pruefen <Wurzel>"
	zcKeineTxt = "zeichen: Eingabe: keine Textdatei gepr\xc3\xbcft"
)

// zcOk ist die Zeile ok von SPEC-233 Nr. 10 c.
func zcOk(n, bin, andere int) string {
	return fmt.Sprintf("zeichen: ok, %d Datei(en) gepr\xc3\xbcft, %d Bin\xc3\xa4rdatei(en) und %d andere Eintr\xc3\xa4ge \xc3\xbcbergangen\n", n, bin, andere)
}

// zcSumme ist die letzte Zeile mit Befunden (Nr. 10 d).
func zcSumme(k, d int) string { return fmt.Sprintf("zeichen: %d Befund(e) in %d Datei(en)\n", k, d) }

// zcUngueltig ist die Fundstelle eines ungültigen Bytes (Nr. 10 a).
func zcUngueltig(pfad string, zeile, spalte int, b byte) string {
	return fmt.Sprintf("zeichen: %s:%d:%d: ung\xc3\xbcltiges UTF-8 (Byte 0x%02X)\n", pfad, zeile, spalte, b)
}

// zcFall baut einen Fall der Gruppe: das Programm mit dem Wegwerfverzeichnis als Wurzel; ein Text
// mit `!` davor darf nicht in der Ausgabe stehen, einer mit `1>` steht auf stdout, sonst auf stderr.
func zcFall(name string, neu map[string]string, exit int, want ...string) Fall {
	f := gmFall(name, gateZeichen, neu, []string{"pruefen", "{dir}"}, exit, want...)
	f.Name = "zeichen/" + name
	return f
}

// zcBereich ist ein Bereich von Z mit seiner Kategorie; die Tabelle steht von Hand, aus dem Wortlaut
// von SPEC-233 Nr. 7, ohne die Tabellen von unicode.
type zcBereich struct {
	von, bis rune
	kat      string
}

func zcZ() []zcBereich {
	return []zcBereich{
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

// zcKurznamen sind die 29 Kurznamen von SPEC-233 Nr. 7 f.
func zcKurznamen() map[rune]string {
	return map[rune]string{
		0x0000: "NUL", 0x000D: "CR", 0x007F: "DEL", 0x0085: "NEL", 0x00A0: "NBSP", 0x00AD: "SHY",
		0x061C: "ALM", 0x180E: "MVS", 0x200B: "ZWSP", 0x200C: "ZWNJ", 0x200D: "ZWJ", 0x200E: "LRM",
		0x200F: "RLM", 0x2028: "LS", 0x2029: "PS", 0x202A: "LRE", 0x202B: "RLE", 0x202C: "PDF",
		0x202D: "LRO", 0x202E: "RLO", 0x202F: "NNBSP", 0x2060: "WJ", 0x2066: "LRI", 0x2067: "RLI",
		0x2068: "FSI", 0x2069: "PDI", 0xFEFF: "BOM", 0xFFFE: "NONCHAR", 0xFFFF: "NONCHAR",
	}
}

// zcZeile ist die Fundstelle eines Zeichens von Z (Nr. 10 a); kat ist seine Kategorie.
func zcZeile(pfad string, zeile, spalte int, r rune, kat string) string {
	t := fmt.Sprintf("zeichen: %s:%d:%d: U+%04X %s", pfad, zeile, spalte, r, kat)
	if k := zcKurznamen()[r]; k != "" {
		t += " " + k
	}
	return t + "\n"
}

// zcNbsp ist die Fundstelle eines NBSP.
func zcNbsp(pfad string, zeile, spalte int) string { return zcZeile(pfad, zeile, spalte, 0x00A0, "Zs") }

// zcErr prüft nach dem Aufruf, dass stderr genau text ist und stdout leer.
func zcErr(text string) func(Lauf) string {
	return func(l Lauf) string {
		if l.Err != text {
			return fmt.Sprintf("stderr %q statt %q", l.Err, text)
		}
		if l.Out != "" {
			return fmt.Sprintf("stdout %q statt leer", l.Out)
		}
		return ""
	}
}

// zcSauber ist Fall 131: Text, Tabulator und Zeichen jenseits von ASCII, die nicht zu Z gehören.
func zcSauber() Fall {
	text := "Text mit Leerzeichen\tTab\n\xc3\xa4\xc2\xa7\xe2\x80\x93\xe2\x82\xac\xef\xbf\xbd\xf0\x9f\x98\x80\n"
	f := zcFall("sauber", map[string]string{"a.go": text, "docs/x.md": text, "Makefile": text, "b.txt": text}, 0,
		"1>"+zcOk(4, 0, 0))
	f.Nachher = func(l Lauf) string {
		if l.Err != "" {
			return fmt.Sprintf("stderr %q statt leer", l.Err)
		}
		return ""
	}
	return f
}

// zcMengeZ ist Fall 132: eine Datei je Element von Z (253), stderr genau die erwartete Liste.
func zcMengeZ() Fall {
	neu := map[string]string{}
	soll := ""
	n := 0
	for _, b := range zcZ() {
		for r := b.von; r <= b.bis; r++ {
			pfad := fmt.Sprintf("z/%06X.txt", r)
			if r == 0 {
				neu[pfad] = strings.Repeat("x", 8000) + "\na\x00b\n"
				soll += zcZeile(pfad, 2, 2, r, b.kat)
			} else {
				neu[pfad] = "a" + string(r) + "b\n"
				soll += zcZeile(pfad, 1, 2, r, b.kat)
			}
			n++
		}
	}
	soll += zcSumme(n, n)
	f := zcFall("menge-z", neu, 1)
	f.Nachher = func(l Lauf) string {
		if n != 253 {
			return fmt.Sprintf("Menge Z hat %d Elemente statt 253", n)
		}
		if l.Err != soll {
			return "stderr weicht von der erwarteten Liste ab: " + abweichung(l.Err, soll)
		}
		for i := 0; i < len(l.Err); i++ {
			if c := l.Err[i]; c != '\n' && (c < 0x20 || c > 0x7E) {
				return fmt.Sprintf("Byte 0x%02X in der Ausgabe", c)
			}
		}
		return ""
	}
	return f
}

// abweichung nennt die erste Zeile, in der zwei Texte sich unterscheiden.
func abweichung(got, want string) string {
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		a, b := "", ""
		if i < len(g) {
			a = g[i]
		}
		if i < len(w) {
			b = w[i]
		}
		if a != b {
			return fmt.Sprintf("Zeile %d: %q statt %q (%d gegen %d Zeilen)", i+1, a, b, len(g), len(w))
		}
	}
	return "gleich"
}

// zcExakt macht aus der Erwartung an stderr eine Prüfung des ganzen Textes.
func zcExakt(f Fall, text string) Fall {
	f.Nachher = zcErr(text)
	return f
}

// zcEinzelFaelle sind die Fälle 133 bis 138 mit kleinen Dateien.
func zcEinzelFaelle() []Fall {
	return []Fall{
		zcExakt(zcFall("bom-anfang-und-mitte", map[string]string{"d.md": zcBOM + "a\nb" + zcBOM + "c\n"}, 1),
			zcZeile("d.md", 1, 1, 0xFEFF, "Cf")+zcZeile("d.md", 2, 2, 0xFEFF, "Cf")+zcSumme(2, 1)),
		zcExakt(zcFall("ausgabeform", map[string]string{"docs/a b.md": "x\ny\n\xc3\xa4\xc3\xa4" + zcNBSP}, 1, "!:3:5:"),
			zcNbsp("docs/a b.md", 3, 3)+zcSumme(1, 1)),
		zcExakt(zcFall("cr-und-zeilenende", map[string]string{"d.md": "a\r\nb" + zcNBSP + "\r\nc\n"}, 1),
			zcZeile("d.md", 1, 2, 0x0D, "Cc")+zcNbsp("d.md", 2, 2)+zcZeile("d.md", 2, 3, 0x0D, "Cc")+zcSumme(3, 1)),
		zcExakt(zcFall("zeilen-trenner-brechen-nicht", map[string]string{"d.md": "a\xe2\x80\xa8b\xe2\x80\xa9c\xc2\x85d\ne" + zcNBSP}, 1),
			zcZeile("d.md", 1, 2, 0x2028, "Zl")+zcZeile("d.md", 1, 4, 0x2029, "Zp")+zcZeile("d.md", 1, 6, 0x85, "Cc")+
				zcNbsp("d.md", 2, 2)+zcSumme(4, 1)),
		zcExakt(zcFall("ungueltiges-utf8", map[string]string{"latin1.md": "ok\na\xe4b\n", "abgebrochen.md": "x\xe2\x80", "ersatz.md": "r\xef\xbf\xbd\n"}, 1, "!ersatz.md"),
			zcUngueltig("abgebrochen.md", 1, 2, 0xE2)+zcUngueltig("abgebrochen.md", 1, 3, 0x80)+
				zcUngueltig("latin1.md", 2, 2, 0xE4)+zcSumme(3, 2)),
		zcExakt(zcFall("utf16-bom", map[string]string{
			"le.md": "\xff\xfea\x00b\x00", "be.md": "\xfe\xff\x00a\x00b", "utf32.md": "\xff\xfe\x00\x00a\x00\x00\x00",
			"nbsp.md": "\xff\xfea\x00" + zcNBSP + "\x00", "ff.md": "\xffa\n"}, 1, "!U+00A0"),
			"zeichen: be.md:1:1: UTF-16-BOM (Bytes 0xFE 0xFF)\n"+zcUngueltig("ff.md", 1, 1, 0xFF)+
				"zeichen: le.md:1:1: UTF-16-BOM (Bytes 0xFF 0xFE)\nzeichen: nbsp.md:1:1: UTF-16-BOM (Bytes 0xFF 0xFE)\n"+
				"zeichen: utf32.md:1:1: UTF-16-BOM (Bytes 0xFF 0xFE)\n"+zcSumme(5, 5)),
	}
}

// zcBinaerGrenze ist Fall 139: NUL an den Grenzen 7999 und 8000, der Name macht nicht binär; danach
// entfernt der Fall die beiden Textdateien mit Befund und ruft erneut.
func zcBinaerGrenze() Fall {
	f := zcFall("binaer-grenze", map[string]string{
		"bin.zip": "\x00" + zcNBSP + "\r" + strings.Repeat("x", 96), "clean.md": "a\n",
		"g7999.md": strings.Repeat("x", 7999) + "\x00", "g8000.md": strings.Repeat("x", 8000) + "\x00",
		"text.zip": "a" + zcNBSP,
	}, 1, "!bin.zip", "!g7999.md")
	f.Nachher = func(l Lauf) string {
		if r := zcErr(zcZeile("g8000.md", 1, 8001, 0, "Cc") + zcNbsp("text.zip", 1, 2) + zcSumme(2, 2))(l); r != "" {
			return r
		}
		for _, d := range []string{"g8000.md", "text.zip"} {
			if err := os.Remove(filepath.Join(l.Dir, d)); err != nil {
				return err.Error()
			}
		}
		code, out, errOut := l.programm("pruefen", l.Dir)
		if code != 0 || out != zcOk(1, 2, 0) || errOut != "" {
			return fmt.Sprintf("zweiter Lauf: Exit %d, stdout %q, stderr %q", code, out, errOut)
		}
		return ""
	}
	return f
}

// zcBaseline ist Fall 140: der Teilbaum .harness/baseline ist ausgenommen, Beinahe-Treffer nicht.
func zcBaseline() Fall {
	inhalt := "a" + zcNBSP
	f := zcFall("baseline-ausgenommen", map[string]string{
		".harness/baseline/v/x.md": inhalt, ".harness/baselinex/y.md": inhalt, "x/.harness/baseline/z.md": inhalt, ".harness/skills/s.md": inhalt,
	}, 1, "!.harness/baseline/v")
	return zcExakt(f, zcNbsp(".harness/baselinex/y.md", 1, 2)+zcNbsp(".harness/skills/s.md", 1, 2)+
		zcNbsp("x/.harness/baseline/z.md", 1, 2)+zcSumme(3, 3))
}

// zcSymlinks ist Fall 141: Symlinks werden nicht gefolgt und als andere Einträge gezählt.
func zcSymlinks() Fall {
	f := zcFall("symlinks-nicht-gefolgt", map[string]string{"aussen/x.md": "a" + zcNBSP, "wurzel/t.md": "a\n"}, 0, "1>"+zcOk(1, 0, 4))
	f.Argumente = []string{"pruefen", "{dir}/wurzel"}
	f.Vorbereitung = func(dir string) error {
		for name, ziel := range map[string]string{"l.md": "t.md", "ld": ".", "h.md": "gibt-es-nicht", "a.md": "../aussen/x.md"} {
			if err := os.Symlink(ziel, filepath.Join(dir, "wurzel", name)); err != nil {
				return err
			}
		}
		return nil
	}
	f.Nachher = func(l Lauf) string {
		if l.Err != "" {
			return fmt.Sprintf("stderr %q statt leer", l.Err)
		}
		if err := os.WriteFile(filepath.Join(l.Dir, "wurzel", "t.md"), []byte("a"+zcNBSP), 0o644); err != nil {
			return err.Error()
		}
		code, out, errOut := l.programm("pruefen", filepath.Join(l.Dir, "wurzel"))
		if code != 1 || out != "" || errOut != zcNbsp("t.md", 1, 2)+zcSumme(1, 1) {
			return fmt.Sprintf("zweiter Lauf: Exit %d, stdout %q, stderr %q", code, out, errOut)
		}
		return ""
	}
	return f
}

// zcWurzeln sind die Fälle 142 bis 145 mit Eingabefehlern und dem Aufruf.
func zcWurzeln() []Fall {
	leer := zcFall("leere-datei", map[string]string{"leer.md": ""}, 0, "1>zeichen: ok, 1 Datei(en) gepr\xc3\xbcft")
	menge := zcFall("leere-menge", nil, 2, zcKeineTxt)
	menge.Nachher = func(l Lauf) string {
		for _, name := range []string{"nur-bin", "nur-base", "nur-symlink", "nur-ordner", "mischung"} {
			w := filepath.Join(l.Tmp, name)
			if err := zcLegeWurzel(w, name); err != nil {
				return err.Error()
			}
			code, out, errOut := l.programm("pruefen", w)
			if code != 2 || out != "" || errOut != zcKeineTxt+"\n" {
				return fmt.Sprintf("%s: Exit %d, stdout %q, stderr %q", name, code, out, errOut)
			}
		}
		if l.Out != "" || strings.Contains(l.Err, "Befund") {
			return fmt.Sprintf("leeres Verzeichnis: stdout %q, stderr %q", l.Out, l.Err)
		}
		return ""
	}
	fehlt := zcFall("wurzel-fehlt", map[string]string{"datei.md": "a\n"}, 2, "zeichen: Eingabe: Wurzel fehlt oder ist kein Verzeichnis: ", "!keine Textdatei")
	fehlt.Argumente = []string{"pruefen", "{dir}/nix"}
	fehlt.Nachher = func(l Lauf) string {
		datei := filepath.Join(l.Dir, "datei.md")
		code, out, errOut := l.programm("pruefen", datei)
		if code != 2 || out != "" || errOut != "zeichen: Eingabe: Wurzel fehlt oder ist kein Verzeichnis: "+datei+"\n" {
			return fmt.Sprintf("Wurzel ist eine Datei: Exit %d, stdout %q, stderr %q", code, out, errOut)
		}
		if !strings.Contains(l.Err, filepath.Join(l.Dir, "nix")) {
			return fmt.Sprintf("stderr nennt den Pfad nicht: %q", l.Err)
		}
		return ""
	}
	return []Fall{leer, menge, fehlt, zcAufrufFall()}
}

// zcLegeWurzel legt in w eine Wurzel ohne Textdatei an (Fall 143): nur eine Binärdatei, nur den Teilbaum
// `.harness/baseline`, nur einen hängenden Symlink, nur ein leeres Unterverzeichnis oder die Mischung aus
// Binärdatei, Symlink und Teilbaum.
func zcLegeWurzel(w, name string) error {
	dateien := map[string]string{}
	if name == "nur-bin" || name == "mischung" {
		dateien["b.bin"] = "\x00\x01"
	}
	if name == "nur-base" || name == "mischung" {
		dateien[".harness/baseline/x.md"] = "a" + zcNBSP
	}
	if err := os.MkdirAll(w, 0o755); err != nil {
		return err
	}
	if err := gmSchreibe(w, dateien); err != nil {
		return err
	}
	if name == "nur-symlink" || name == "mischung" {
		if err := os.Symlink("gibt-es-nicht", filepath.Join(w, "l")); err != nil {
			return err
		}
	}
	if name == "nur-ordner" {
		return os.Mkdir(filepath.Join(w, "d"), 0o755)
	}
	return nil
}

// zcAufrufFall ist Fall 145: acht Aufrufe, der erste ohne Argumente.
func zcAufrufFall() Fall {
	f := zcFall("aufruf", map[string]string{"w/z.md": "a\n"}, 2, zcAufruf)
	f.Argumente = nil
	f.Nachher = func(l Lauf) string {
		w := filepath.Join(l.Dir, "w")
		formen := [][]string{{"pruefen"}, {"pruefen", ""}, {"pruefen", w, "extra"}, {w}, {"check", w}, {"--help"}, {"pruefen", "--ausnahme", w}}
		if len(formen) != 7 {
			return "Aufrufformen: nicht sieben"
		}
		for _, a := range formen {
			code, out, errOut := l.programm(a...)
			if code != 2 || out != "" || !strings.Contains(errOut, zcAufruf) {
				return fmt.Sprintf("%q: Exit %d, stdout %q, stderr %q", a, code, out, errOut)
			}
		}
		if l.Out != "" {
			return fmt.Sprintf("ohne Argumente: stdout %q", l.Out)
		}
		return ""
	}
	return f
}

// zcDateiarten ist Fall 146: 21 Dateien, Name und Verzeichnis haben keine Wirkung.
func zcDateiarten() Fall {
	namen := []string{"a.go", "a.md", "a.mk", "a.sh", "a.yml", "a.yaml", "a.json", "a.xml", "a.sql", "a.txt", "go.mod", "go.sum",
		"Makefile", "Dockerfile", "LICENSE", ".dockerignore", ".gitattributes", ".claude/settings.json", ".githooks/commit-msg",
		"a.zip", "a.stempel"}
	neu := map[string]string{}
	want := []string{"21 Befund(e) in 21 Datei(en)"}
	for _, n := range namen {
		neu[n] = "a" + zcNBSP
		want = append(want, n+":1:2: U+00A0 Zs NBSP")
	}
	f := zcFall("dateiarten", neu, 1, want...)
	f.Nachher = func(Lauf) string {
		if len(namen) != 21 {
			return fmt.Sprintf("%d Namen statt 21", len(namen))
		}
		return ""
	}
	return f
}

// zcPruefsumme ist die Prüfsumme aller Pfade und Inhalte unter dir.
func zcPruefsumme(dir string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00", p)
		if d.Type().IsRegular() {
			b, err := os.ReadFile(p)
			fmt.Fprintf(h, "%x", sha256.Sum256(b))
			return err
		}
		return nil
	})
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

// zcReihenfolge ist Fall 147: die Folge der Pfade, bytegleich, und ein zweiter Lauf auf dem
// unveränderten Baum.
func zcReihenfolge() Fall {
	neu := map[string]string{}
	for _, n := range []string{"b.md", "a/b.md", "a-b.md", "a.md", "A.md", "a0.md", "a~.md", "\xc3\xa4.md"} {
		neu[n] = "x" + zcNBSP
	}
	soll := ""
	for _, n := range []string{"A.md", "a-b.md", "a.md", "a/b.md", "a0.md", "a~.md", "b.md", `\xC3\xA4.md`} {
		soll += zcNbsp(n, 1, 2)
	}
	soll += zcSumme(8, 8)
	var vorher string
	f := zcFall("reihenfolge-und-wiederholung", neu, 1)
	f.Vorbereitung = func(dir string) (err error) {
		vorher, err = zcPruefsumme(dir)
		return err
	}
	f.Nachher = func(l Lauf) string {
		if l.Err != soll {
			return "stderr weicht ab: " + abweichung(l.Err, soll)
		}
		code, out, errOut := l.programm("pruefen", l.Dir)
		if code != 1 || out != l.Out || errOut != l.Err {
			return fmt.Sprintf("zweiter Lauf weicht ab: Exit %d, stdout %q, stderr %q", code, out, errOut)
		}
		if nachher, err := zcPruefsumme(l.Dir); err != nil || nachher != vorher {
			return fmt.Sprintf("die Wurzel hat sich verändert (%v)", err)
		}
		return ""
	}
	return f
}

// zcPaare ist Fall 149: zwei Zeichen derselben Zeile, dasselbe Zeichen mehrfach und in zwei Dateien,
// Zeichen und ungültiges Byte derselben Datei.
func zcPaare() Fall {
	f := zcFall("paare", map[string]string{
		"a.md": "p\xe2\x80\x8b" + zcNBSP + "q\xe2\x80\x8b\nr\xe4s" + zcNBSP, "b.md": "u" + zcNBSP, "c.md": "v" + zcNBSP + zcNBSP + zcNBSP,
	}, 1)
	return zcExakt(f, zcZeile("a.md", 1, 2, 0x200B, "Cf")+zcNbsp("a.md", 1, 3)+zcZeile("a.md", 1, 5, 0x200B, "Cf")+
		zcUngueltig("a.md", 2, 2, 0xE4)+zcNbsp("a.md", 2, 4)+zcNbsp("b.md", 1, 2)+
		zcNbsp("c.md", 1, 2)+zcNbsp("c.md", 1, 3)+zcNbsp("c.md", 1, 4)+zcSumme(9, 3))
}

// zcPfade ist Fall 150: sieben Namen mit Leerzeichen, Zeilenumbruch, Rückwärtsschrägstrich, Tabulator,
// Umlaut, NBSP und einem ungültigen Byte; die Ausgabe führt sie maskiert.
func zcPfade() Fall {
	namen := map[string]string{
		"a b.md": "a b.md", "a\nb.md": `a\x0Ab.md`, "a\\b.md": `a\\b.md`, "a\tb.md": `a\x09b.md`,
		"\xc3\xa4.md": `\xC3\xA4.md`, "n" + zcNBSP + "b.md": `n\xC2\xA0b.md`, "p\xffq.md": `p\xFFq.md`,
	}
	neu := map[string]string{}
	var roh []string
	for n := range namen {
		neu[n] = "x" + zcNBSP
		roh = append(roh, n)
	}
	slices.Sort(roh)
	soll := ""
	for _, n := range roh {
		soll += zcNbsp(namen[n], 1, 2)
	}
	soll += zcSumme(7, 7)
	f := zcFall("pfad-sonderzeichen", neu, 1)
	f.Nachher = func(l Lauf) string {
		if len(namen) != 7 {
			return fmt.Sprintf("%d Namen statt 7", len(namen))
		}
		if l.Err != soll {
			return "stderr weicht ab: " + abweichung(l.Err, soll)
		}
		for i := 0; i < len(l.Err); i++ {
			if c := l.Err[i]; c != '\n' && (c < 0x20 || c > 0x7E) {
				return fmt.Sprintf("Byte 0x%02X in der Ausgabe", c)
			}
		}
		return ""
	}
	return f
}

// zcGrosseZeile ist Fall 148: eine Zeile von 200 000 Zeichen.
func zcGrosseZeile() Fall {
	return zcFall("grosse-zeile", map[string]string{"g.md": strings.Repeat("x", 200000) + zcNBSP + "\n" + zcNBSP}, 1,
		":1:200001: U+00A0", ":2:1: U+00A0", "2 Befund(e) in 1 Datei(en)")
}

// zeichenProgrammFaelle sind die Fälle 131 bis 150.
func zeichenProgrammFaelle() []Fall {
	alle := []Fall{zcSauber(), zcMengeZ()}
	alle = append(alle, zcEinzelFaelle()...)
	alle = append(alle, zcBinaerGrenze(), zcBaseline(), zcSymlinks())
	alle = append(alle, zcWurzeln()...)
	return append(alle, zcDateiarten(), zcReihenfolge(), zcGrosseZeile(), zcPaare(), zcPfade())
}
