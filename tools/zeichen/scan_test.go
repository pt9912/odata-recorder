package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

type leserForm struct {
	name string
	neu  func([]byte) io.Reader
}

// leserFormen sind die vier Leser der Tabelle: ganz, ein Byte je Aufruf, halbe Puffer, Daten mit EOF.
func leserFormen() []leserForm {
	return []leserForm{
		{"ganz", func(d []byte) io.Reader { return bytes.NewReader(d) }},
		{"ein-byte", func(d []byte) io.Reader { return iotest.OneByteReader(bytes.NewReader(d)) }},
		{"halb", func(d []byte) io.Reader { return iotest.HalfReader(bytes.NewReader(d)) }},
		{"daten-mit-eof", func(d []byte) io.Reader { return iotest.DataErrReader(bytes.NewReader(d)) }},
	}
}

// scanne fährt die Prüfung eines Stroms und liefert die Zahl der Fundstellen und stderr.
func scanne(r io.Reader) (int, string) {
	var o, e bytes.Buffer
	n := neuerPruefer(&o, &e).scan("d", r)
	return n, e.String()
}

type stellung struct {
	name string
	bau  func(z []byte) []byte
	orte [][2]int
}

// stellungen sind die sieben Stellungen eines Zeichens in einer Datei mit ihren Orten (Zeile, Spalte).
func stellungen() []stellung {
	uml := cat(b("\xC3\xA4"), b("\xC3\xB6"))
	return []stellung{
		{"dateianfang", func(z []byte) []byte { return cat(z, b("abc\n")) }, [][2]int{{1, 1}}},
		{"zeilenanfang", func(z []byte) []byte { return cat(b("ab\n"), z, b("cd\n")) }, [][2]int{{2, 1}}},
		{"hinter-umlauten", func(z []byte) []byte { return cat(uml, z, b("x\n")) }, [][2]int{{1, 3}}},
		{"vor-zeilenvorschub", func(z []byte) []byte { return cat(b("ab"), z, b("\n")) }, [][2]int{{1, 3}}},
		{"dateiende-ohne-lf", func(z []byte) []byte { return cat(b("ab\ncd"), z) }, [][2]int{{2, 3}}},
		{"nach-dem-letzten-lf", func(z []byte) []byte { return cat(b("ab\n"), z) }, [][2]int{{2, 1}}},
		{"zweimal-benachbart", func(z []byte) []byte { return cat(b("a"), z, z, b("b\n")) }, [][2]int{{1, 2}, {1, 3}}},
	}
}

func TestStellungen(t *testing.T) {
	elemente, faelle := zElemente(), 0
	if len(elemente) != 253 || len(stellungen()) != 7 || len(leserFormen()) != 4 {
		t.Fatalf("Mengen: %d, %d, %d statt 253, 7, 4", len(elemente), len(stellungen()), len(leserFormen()))
	}
	for _, r := range elemente {
		for _, s := range stellungen() {
			daten := s.bau(zb(r))
			soll := ""
			for _, o := range s.orte {
				soll += orakelZeile("d", o[0], o[1], r)
			}
			for _, l := range leserFormen() {
				faelle++
				n, got := scanne(l.neu(daten))
				if got != soll || n != len(s.orte) {
					t.Fatalf("U+%04X, %s, %s: %d Befunde, %q statt %q", r, s.name, l.name, n, got, soll)
				}
			}
		}
	}
	if faelle != 253*7*4 {
		t.Errorf("%d Fälle", faelle)
	}
}

// nichtZ sind die 208 Zeichen, die passieren (SPEC-233 Nr. 7 e, 16 b), als Codepunkte.
func nichtZ() []rune {
	var l []rune
	for c := rune(0x20); c <= 0x7E; c++ {
		l = append(l, c)
	}
	l = append(l, '\t', '\n')
	for c := rune(0xA1); c <= 0xFF; c++ {
		if c != 0xAD {
			l = append(l, c)
		}
	}
	return append(l, 0x2013, 0x2014, 0x2018, 0x2019, 0x201C, 0x201D, 0x2026, 0x20AC, 0x2192, 0x2713,
		0xFFFD, 0xFE0F, 0x034F, 0x3164, 0x2800, 0x4E2D, 0x1F600)
}

func TestNichtZ(t *testing.T) {
	if len(nichtZ()) != 208 {
		t.Fatalf("%d Zeichen statt 208", len(nichtZ()))
	}
	for _, r := range nichtZ() {
		for _, s := range stellungen() {
			for _, l := range leserFormen() {
				if n, got := scanne(l.neu(s.bau(zb(r)))); n != 0 || got != "" {
					t.Fatalf("U+%04X, %s, %s: %q", r, s.name, l.name, got)
				}
			}
		}
	}
	// Die Schreibweise aus Backslash, u und vier Hex-Ziffern ist Text (SPEC-233 Nr. 7 e).
	if n, got := scanne(bytes.NewReader([]byte{0x5C, 'u', '0', '0', 'A', '0', '\n'})); n != 0 || got != "" {
		t.Errorf("Schreibweise: %q", got)
	}
}

type ungueltig struct {
	name  string
	folge []byte
	nach  []byte
}

// ungueltige sind die 90 Elemente: jedes Byte der Folge wird zurückgewiesen.
func ungueltige() []ungueltig {
	var l []ungueltig
	add := func(name string, folge []byte, nach string) {
		l = append(l, ungueltig{name, folge, []byte(nach)})
	}
	for c := 0x80; c <= 0xBF; c++ {
		add(fmt.Sprintf("fortsetzung-%02X", c), []byte{byte(c)}, "b\n")
	}
	add("C0", []byte{0xC0}, "b\n")
	add("C1", []byte{0xC1}, "b\n")
	for c := 0xF5; c <= 0xFF; c++ {
		add(fmt.Sprintf("anfang-%02X", c), []byte{byte(c)}, "b\n")
	}
	for _, f := range [][]byte{{0xC2}, {0xE2, 0x82}, {0xF0, 0x9F, 0x98}} {
		add(fmt.Sprintf("abgebrochen-%X-ende", f), f, "")
		add(fmt.Sprintf("abgebrochen-%X-lf", f), f, "\n")
	}
	add("ueberlang-E0", []byte{0xE0, 0x80, 0x80}, "b\n")
	add("ueberlang-F0", []byte{0xF0, 0x80, 0x80, 0x80}, "b\n")
	add("ueberlang-C0AF", []byte{0xC0, 0xAF}, "b\n")
	add("surrogat-A0", []byte{0xED, 0xA0, 0x80}, "b\n")
	add("surrogat-BF", []byte{0xED, 0xBF, 0xBF}, "b\n")
	add("ueber-10FFFF-F4", []byte{0xF4, 0x90, 0x80, 0x80}, "b\n")
	add("ueber-10FFFF-F7", []byte{0xF7, 0xBF, 0xBF, 0xBF}, "b\n")
	return l
}

func TestUngueltigesUTF8(t *testing.T) {
	if len(ungueltige()) != 90 {
		t.Fatalf("%d Elemente statt 90", len(ungueltige()))
	}
	for _, u := range ungueltige() {
		daten := cat(b("a"), u.folge, u.nach)
		soll := ""
		for i, c := range u.folge {
			soll += fmt.Sprintf("zeichen: d:1:%d: ungültiges UTF-8 (Byte 0x%02X)\n", 2+i, c)
		}
		for _, l := range leserFormen() {
			if n, got := scanne(l.neu(daten)); got != soll || n != len(u.folge) {
				t.Fatalf("%s, %s: %d Befunde, %q statt %q", u.name, l.name, n, got, soll)
			}
		}
	}
}

func TestZeilenUndSpalten(t *testing.T) {
	// Nur der Zeilenvorschub beginnt eine Zeile (SPEC-233 Nr. 9); die Spalte zählt Zeichen.
	daten := cat(b("a"), zb(0x2028), b("b"), zb(0x2029), b("c"), zb(0x0085), b("d\ne"), zb(0x00A0), b("\n"),
		b("f"), zb(0x000D), b("\n"), b("\xC3\xA4\xC3\xA4"), zb(0x00A0))
	soll := orakelZeile("d", 1, 2, 0x2028) + orakelZeile("d", 1, 4, 0x2029) + orakelZeile("d", 1, 6, 0x0085) +
		orakelZeile("d", 2, 2, 0x00A0) + orakelZeile("d", 3, 2, 0x000D) + orakelZeile("d", 4, 3, 0x00A0)
	for _, l := range leserFormen() {
		if _, got := scanne(l.neu(daten)); got != soll {
			t.Errorf("%s: %q statt %q", l.name, got, soll)
		}
	}
}

func TestBomAnfangUndMitte(t *testing.T) {
	daten := cat(zb(0xFEFF), b("a\nb"), zb(0xFEFF), b("c\n"))
	_, got := scanne(bytes.NewReader(daten))
	gleich(t, "BOM", got, orakelZeile("d", 1, 1, 0xFEFF)+orakelZeile("d", 2, 2, 0xFEFF))
}

func TestUTF16Bom(t *testing.T) {
	type el struct {
		name  string
		daten []byte
		soll  string
		bin   bool
	}
	bom := func(a, c byte) string {
		return fmt.Sprintf("zeichen: @:1:1: UTF-16-BOM (Bytes 0x%02X 0x%02X)\n", a, c)
	}
	ung := func(folge ...byte) string {
		s := ""
		for i, c := range folge {
			s += fmt.Sprintf("zeichen: @:1:%d: ungültiges UTF-8 (Byte 0x%02X)\n", 1+i, c)
		}
		return s
	}
	elemente := []el{
		{"ff-fe-nul", []byte{0xFF, 0xFE, 0x61, 0x00}, bom(0xFF, 0xFE), false},
		{"fe-ff-nul", []byte{0xFE, 0xFF, 0x00, 0x61}, bom(0xFE, 0xFF), false},
		{"utf32", []byte{0xFF, 0xFE, 0x00, 0x00}, bom(0xFF, 0xFE), false},
		{"nur-bom", []byte{0xFF, 0xFE}, bom(0xFF, 0xFE), false},
		{"ff", []byte{0xFF}, ung(0xFF), false},
		{"fe", []byte{0xFE}, ung(0xFE), false},
		{"ff-ff", []byte{0xFF, 0xFF}, ung(0xFF, 0xFF), false},
		{"fe-fe", []byte{0xFE, 0xFE}, ung(0xFE, 0xFE), false},
	}
	if len(elemente) != 8 {
		t.Fatalf("%d Elemente statt 8", len(elemente))
	}
	for _, e := range elemente {
		r := fahre(baum(t, map[string][]byte{"d.md": e.daten, "z.md": b("a\n")}))
		gleich(t, e.name, r.err, strings.ReplaceAll(e.soll, "@", "d.md")+summe(strings.Count(e.soll, "\n"), 1))
		if r.code != 1 || r.out != "" {
			t.Errorf("%s: Exit %d, stdout %q", e.name, r.code, r.out)
		}
	}
	// Ein NBSP hinter dem BOM wird nicht mehr gelesen; die Datei zählt als gelesen (n = 1: kein
	// Eingabefehler `keine Textdatei geprüft`).
	r := fahre(baum(t, map[string][]byte{"nbsp.md": cat([]byte{0xFF, 0xFE, 0x61, 0x00}, zb(0x00A0), []byte{0x00})}))
	gleich(t, "BOM und NBSP", r.err, strings.ReplaceAll(bom(0xFF, 0xFE), "@", "nbsp.md")+summe(1, 1))
	if r.code != 1 {
		t.Errorf("BOM und NBSP: Exit %d", r.code)
	}
}
