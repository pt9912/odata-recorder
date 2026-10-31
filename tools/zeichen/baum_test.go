package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"testing/iotest"
	"time"
)

func xs(n int) []byte { return bytes.Repeat(b("x"), n) }

func nulBei(i int) []byte { return cat(xs(i), []byte{0}) }

func TestBinaerGrenze(t *testing.T) {
	type el struct {
		name   string
		daten  []byte
		binaer bool
		spalte int
	}
	elemente := []el{
		{"nul-0", nulBei(0), true, 0}, {"nul-1", nulBei(1), true, 0},
		{"nul-7998", nulBei(7998), true, 0}, {"nul-7999", nulBei(7999), true, 0},
		{"nul-8000", nulBei(8000), false, 8001}, {"nul-8001", nulBei(8001), false, 8002},
		{"laenge-0", nil, false, 0}, {"laenge-1-nul", []byte{0}, true, 0},
		{"laenge-7999-nul-am-ende", nulBei(7998), true, 0}, {"laenge-8000-nul-am-ende", nulBei(7999), true, 0},
	}
	if len(elemente) != 10 {
		t.Fatalf("%d Elemente statt 10", len(elemente))
	}
	for _, e := range elemente {
		r := fahre(baum(t, map[string][]byte{"d.md": e.daten, "z.md": b("a\n")}))
		switch {
		case e.binaer:
			gleich(t, e.name, r.out, okZeile(1, 1, 0))
			gleich(t, e.name+" stderr", r.err, "")
		case e.spalte > 0:
			gleich(t, e.name, r.err, fmt.Sprintf("zeichen: d.md:1:%d: U+0000 Cc NUL\n", e.spalte)+summe(1, 1))
		default:
			gleich(t, e.name, r.out, okZeile(2, 0, 0))
		}
	}
}

// TestStellungenInDatei führt jedes Element von Z außer U+0000 (am Anfang einer Datei macht es sie
// binär, TestBinaerGrenze) in jeder Stellung als Datei durch die ganze Prüfung.
func TestStellungenInDatei(t *testing.T) {
	type fall struct {
		name string
		r    rune
		s    stellung
	}
	var faelle []fall
	dateien := map[string][]byte{}
	for _, r := range zElemente() {
		if r == 0 {
			continue
		}
		for i, s := range stellungen() {
			n := fmt.Sprintf("%06X-%d.md", r, i)
			faelle = append(faelle, fall{n, r, s})
			dateien[n] = s.bau(zb(r))
		}
	}
	sort.Slice(faelle, func(i, j int) bool { return faelle[i].name < faelle[j].name })
	soll, k := "", 0
	for _, f := range faelle {
		for _, o := range f.s.orte {
			soll += orakelZeile(f.name, o[0], o[1], f.r)
			k++
		}
	}
	if len(faelle) != 252*7 {
		t.Fatalf("%d Dateien statt %d", len(faelle), 252*7)
	}
	r := fahre(baum(t, dateien))
	gleich(t, "Stellungen in Dateien", r.err, soll+summe(k, len(faelle)))
	if r.code != 1 {
		t.Errorf("Exit %d", r.code)
	}
}

func TestDateiarten(t *testing.T) {
	namen := []string{"a.go", "a.md", "a.mk", "a.sh", "a.yml", "a.yaml", "a.json", "a.xml", "a.sql", "a.txt",
		"go.mod", "go.sum", "Makefile", "Dockerfile", "LICENSE", ".dockerignore", ".gitattributes",
		".claude/settings.json", ".githooks/commit-msg", "a.zip", "a.stempel"}
	if len(namen) != 21 {
		t.Fatalf("%d Namen statt 21", len(namen))
	}
	dateien := map[string][]byte{}
	for _, n := range namen {
		dateien[n] = cat(b("a"), zb(0x00A0))
	}
	sort.Strings(namen)
	soll := ""
	for _, n := range namen {
		soll += orakelZeile(n, 1, 2, 0x00A0)
	}
	r := fahre(baum(t, dateien))
	gleich(t, "Dateiarten", r.err, soll+summe(21, 21))
}

func TestAusschluesse(t *testing.T) {
	faelle := []struct {
		pfad        string
		ausgenommen bool
	}{
		{".harness/baseline/v/x.md", true}, {".harness/baseline/x.md", true},
		{".harness/baselinex/y.md", false}, {"x/.harness/baseline/z.md", false},
		{".harness/skills/s.md", false}, {".harness/baseline", false}, {"baseline/x.md", false},
	}
	for _, f := range faelle {
		r := fahre(baum(t, map[string][]byte{f.pfad: cat(b("a"), zb(0x00A0)), "z.md": b("a\n")}))
		if f.ausgenommen {
			gleich(t, f.pfad, r.out, okZeile(1, 0, 0))
		} else {
			gleich(t, f.pfad, r.err, orakelZeile(f.pfad, 1, 2, 0x00A0)+summe(1, 1))
		}
	}
	// Keine zweite Ausnahme (SPEC-233 Nr. 5 a): 22 Namen mal 3 Stellungen (an der Wurzel, unter x/, unter
	// .harness/) ohne das Paar .harness/baseline = 65 Dateien, alle ein Befund.
	namen := strings.Fields(`.git .github .idea .vscode .cache .claude .githooks .harness baseline skills harness docs tools src
		testdata vendor node_modules tmp build dist bin out`)
	if len(namen) != 22 {
		t.Fatalf("%d Namen statt 22", len(namen))
	}
	dateien := map[string][]byte{}
	var pfade []string
	for _, n := range namen {
		for _, p := range []string{n + "/f.md", "x/" + n + "/f.md", ".harness/" + n + "/f.md"} {
			if p == ".harness/baseline/f.md" {
				continue
			}
			dateien[p] = cat(b("a"), zb(0x00A0))
			pfade = append(pfade, p)
		}
	}
	if len(pfade) != 65 {
		t.Fatalf("%d Dateien statt 65", len(pfade))
	}
	sort.Strings(pfade)
	soll := ""
	for _, p := range pfade {
		soll += orakelZeile(p, 1, 2, 0x00A0)
	}
	ersteAbweichung(t, "keine zweite Ausnahme", fahre(baum(t, dateien)).err, soll+summe(65, 65))
}

func TestSymlinks(t *testing.T) {
	art := []struct {
		name  string
		legen func(t *testing.T, dir string)
	}{
		{"auf-datei", func(t *testing.T, dir string) { link(t, "t.md", dir+"/wurzel/l.md") }},
		{"auf-verzeichnis", func(t *testing.T, dir string) { link(t, ".", dir+"/wurzel/ld") }},
		{"haengend", func(t *testing.T, dir string) { link(t, "gibt-es-nicht", dir+"/wurzel/h.md") }},
		{"schleife", func(t *testing.T, dir string) { link(t, "s", dir+"/wurzel/s") }},
		{"ausserhalb", func(t *testing.T, dir string) {
			if err := os.MkdirAll(dir+"/aussen", 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dir+"/aussen/x.md", cat(b("a"), zb(0x00A0)), 0o644); err != nil {
				t.Fatal(err)
			}
			link(t, "../aussen/x.md", dir+"/wurzel/a.md")
		}},
		{"pipe", func(t *testing.T, dir string) {
			if err := syscall.Mkfifo(dir+"/wurzel/p", 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}
	if len(art) != 6 {
		t.Fatalf("%d Elemente statt 6", len(art))
	}
	for _, a := range art {
		dir := t.TempDir()
		if err := os.MkdirAll(dir+"/wurzel", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dir+"/wurzel/t.md", b("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		a.legen(t, dir)
		r := mitFrist(t, a.name, dir+"/wurzel")
		gleich(t, a.name, r.out, okZeile(1, 0, 1))
		gleich(t, a.name+" stderr", r.err, "")
		if err := os.WriteFile(dir+"/wurzel/t.md", cat(b("a"), zb(0x00A0)), 0o644); err != nil {
			t.Fatal(err)
		}
		r = mitFrist(t, a.name, dir+"/wurzel")
		gleich(t, a.name+" mit Zeichen", r.err, orakelZeile("t.md", 1, 2, 0x00A0)+summe(1, 1))
	}
}

func link(t *testing.T, ziel, name string) {
	t.Helper()
	if err := os.Symlink(ziel, name); err != nil {
		t.Fatal(err)
	}
}

// mitFrist fährt die Prüfung; ein Lauf, der länger als zehn Sekunden braucht, hängt (eine geöffnete Pipe).
func mitFrist(t *testing.T, was, wurzel string) ergebnis {
	t.Helper()
	c := make(chan ergebnis, 1)
	go func() { c <- fahre(wurzel) }()
	select {
	case r := <-c:
		return r
	case <-time.After(10 * time.Second):
		t.Fatalf("%s: die Prüfung hängt", was)
	}
	return ergebnis{}
}

type schliesser struct{ io.Reader }

func (schliesser) Close() error { return nil }

func TestLesefehler(t *testing.T) {
	kaputt := errors.New("kaputt")
	laengen := []int{0, 1, 7999, 8000, 20000}
	for _, l := range laengen {
		daten := xs(l)
		if l >= 2 {
			daten = cat(xs(l-2), zb(0x00A0))
		}
		dir := baum(t, map[string][]byte{"f.md": nil, "g.md": cat(b("a"), zb(0x00A0))})
		var o, e bytes.Buffer
		p := neuerPruefer(&o, &e)
		echt := p.open
		p.open = func(pfad string) (io.ReadCloser, error) {
			if strings.HasSuffix(pfad, "/f.md") {
				return schliesser{io.MultiReader(bytes.NewReader(daten), iotest.ErrReader(kaputt))}, nil
			}
			return echt(pfad)
		}
		code := p.pruefen(dir)
		soll, k, d := "", 1, 1
		if l >= 2 {
			soll += orakelZeile("f.md", 1, l-1, 0x00A0)
			k, d = 2, 2
		}
		soll += "zeichen: Eingabe: f.md: nicht lesbar: kaputt\n" + orakelZeile("g.md", 1, 2, 0x00A0) + summe(k, d)
		gleich(t, fmt.Sprintf("Fehler nach %d Bytes", l), e.String(), soll)
		if code != 2 || o.Len() != 0 {
			t.Errorf("Fehler nach %d Bytes: Exit %d, stdout %q", l, code, o.String())
		}
	}
	if len(laengen) != 5 {
		t.Fatal("Mengen")
	}
}

func TestLesefehlerBeiOeffnenUndVerzeichnis(t *testing.T) {
	dir := baum(t, map[string][]byte{"d/x.md": b("a\n"), "f.md": b("a\n"), "g.md": cat(b("a"), zb(0x00A0)),
		"n\xC2\xA0.md": b("a\n")})
	var o, e bytes.Buffer
	p := neuerPruefer(&o, &e)
	echt, echtDir := p.open, p.readDir
	p.open = func(pfad string) (io.ReadCloser, error) {
		switch {
		case strings.HasSuffix(pfad, "/f.md"):
			return nil, &fs.PathError{Op: "open", Path: pfad, Err: syscall.EACCES}
		case strings.HasSuffix(pfad, "\xA0.md"):
			return nil, errors.New(string([]byte{'a', 0xFF, '\\', 'b'}))
		}
		return echt(pfad)
	}
	p.readDir = func(pfad string) ([]os.DirEntry, error) {
		if strings.HasSuffix(pfad, "/d") {
			return nil, errors.New("boom")
		}
		return echtDir(pfad)
	}
	code := p.pruefen(dir)
	soll := "zeichen: Eingabe: d: nicht lesbar: boom\nzeichen: Eingabe: f.md: nicht lesbar: permission denied\n" +
		orakelZeile("g.md", 1, 2, 0x00A0) + `zeichen: Eingabe: n\xC2\xA0.md: nicht lesbar: a\xFF\\b` + "\n" + summe(1, 1)
	gleich(t, "Öffnen und Verzeichnis", e.String(), soll)
	if code != 2 {
		t.Errorf("Exit %d", code)
	}
}

func TestAufruf(t *testing.T) {
	wurzel := baum(t, map[string][]byte{"z.md": b("a\n")})
	formen := [][]string{{}, {"pruefen"}, {"pruefen", ""}, {"pruefen", wurzel, "extra"}, {wurzel},
		{"check", wurzel}, {"--help"}, {"pruefen", "--ausnahme", wurzel}}
	if len(formen) != 8 {
		t.Fatalf("%d Formen statt 8", len(formen))
	}
	for _, f := range formen {
		var o, e bytes.Buffer
		code := run(f, &o, &e)
		if code != 2 || o.Len() != 0 || e.String() != "Aufruf: zeichen pruefen <Wurzel>\n" {
			t.Errorf("%q: Exit %d, stdout %q, stderr %q", f, code, o.String(), e.String())
		}
	}
}

func TestExit(t *testing.T) {
	// Befund (nein, ja) mal Eingabefehler (keiner, Lesefehler); eine leere Menge und eine fehlende
	// Wurzel tragen nie einen Befund (n = 0).
	for _, befund := range []bool{false, true} {
		for _, lese := range []bool{false, true} {
			d := map[string][]byte{"z.md": b("a\n")}
			if befund {
				d["b.md"] = cat(b("a"), zb(0x00A0))
			}
			dir := baum(t, d)
			var o, e bytes.Buffer
			p := neuerPruefer(&o, &e)
			if lese {
				echt := p.open
				p.open = func(pfad string) (io.ReadCloser, error) {
					if strings.HasSuffix(pfad, "/z.md") {
						return schliesser{iotest.ErrReader(errors.New("kaputt"))}, nil
					}
					return echt(pfad)
				}
			}
			soll := 0
			switch {
			case lese:
				soll = 2
			case befund:
				soll = 1
			}
			if code := p.pruefen(dir); code != soll {
				t.Errorf("Befund %v, Lesefehler %v: Exit %d statt %d", befund, lese, code, soll)
			}
		}
	}
	leer := t.TempDir()
	for was, wurzel := range map[string]string{"leer": leer, "fehlt": leer + "/nix"} {
		if r := fahre(wurzel); r.code != 2 {
			t.Errorf("%s: Exit %d", was, r.code)
		}
	}
}

func TestLeereMengeUndWurzel(t *testing.T) {
	gleich(t, "leere Datei", fahre(baum(t, map[string][]byte{"leer.md": nil})).out, okZeile(1, 0, 0))
	// Sieben Wurzeln ohne Textdatei (SPEC-233 Nr. 11 b); ein Verzeichnis zählt weder als Datei noch als Eintrag.
	symlink, pipe, ordner := t.TempDir(), t.TempDir(), t.TempDir()
	mischung := baum(t, map[string][]byte{"b.bin": {0x00, 0x01}, ".harness/baseline/x.md": cat(b("a"), zb(0x00A0))})
	link(t, "gibt-es-nicht", symlink+"/l")
	link(t, "gibt-es-nicht", mischung+"/l")
	if err := syscall.Mkfifo(pipe+"/p", 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ordner+"/d", 0o755); err != nil {
		t.Fatal(err)
	}
	wurzeln := map[string]string{
		"leer":        t.TempDir(),
		"nur-bin":     baum(t, map[string][]byte{"b.bin": {0x00, 0x01}}),
		"nur-base":    baum(t, map[string][]byte{".harness/baseline/x.md": cat(b("a"), zb(0x00A0))}),
		"nur-symlink": symlink, "nur-pipe": pipe, "nur-ordner": ordner, "mischung": mischung,
	}
	if len(wurzeln) != 7 {
		t.Fatalf("%d Wurzeln statt 7", len(wurzeln))
	}
	for was, w := range wurzeln {
		r := fahre(w)
		if r.code != 2 || r.out != "" || r.err != "zeichen: Eingabe: keine Textdatei geprüft\n" {
			t.Errorf("%s: Exit %d, stdout %q, stderr %q", was, r.code, r.out, r.err)
		}
	}
	datei := baum(t, map[string][]byte{"d.md": b("a\n")}) + "/d.md"
	for w, maskiert := range map[string]string{t.TempDir() + "/nix": "", datei: "", t.TempDir() + "/n\xFF\\x": `/n\xFF\\x`} {
		r := fahre(w)
		erwartet := w
		if maskiert != "" {
			erwartet = strings.TrimSuffix(w, "/n\xFF\\x") + maskiert
		}
		if r.code != 2 || r.out != "" || r.err != "zeichen: Eingabe: Wurzel fehlt oder ist kein Verzeichnis: "+erwartet+"\n" {
			t.Errorf("%s: Exit %d, stdout %q, stderr %q", w, r.code, r.out, r.err)
		}
	}
}

func TestSauber(t *testing.T) {
	text := cat(b("Text\tTab\n"), b("\xC3\xA4"), zb(0x00A7), zb(0x2013), zb(0x20AC), zb(0xFFFD), zb(0x1F600), b("\n"))
	dir := baum(t, map[string][]byte{"a.go": text, "docs/x.md": text, "Makefile": text, "b.txt": text})
	r := fahre(dir)
	if r.code != 0 || r.err != "" {
		t.Errorf("Exit %d, stderr %q", r.code, r.err)
	}
	gleich(t, "stdout", r.out, okZeile(4, 0, 0))
	// Zählung: zwei Binärdateien, ein Symlink.
	if err := os.WriteFile(dir+"/b1.bin", []byte{1, 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/b2.zip", []byte{0, 1}, 0o644); err != nil {
		t.Fatal(err)
	}
	link(t, "a.go", dir+"/l.go")
	gleich(t, "Zählung", fahre(dir).out, okZeile(4, 2, 1))
}

func TestReihenfolge(t *testing.T) {
	namen := []string{"b.md", "a/b.md", "a-b.md", "a.md", "A.md", "a0.md", "a~.md", "a\xC3\xA4.md", "\xC3\xA4.md"}
	dateien := map[string][]byte{}
	for _, n := range namen {
		dateien[n] = cat(b("x"), zb(0x00A0))
	}
	dir := baum(t, dateien)
	soll := ""
	for _, n := range []string{"A.md", "a-b.md", "a.md", "a/b.md", "a0.md", "a~.md"} {
		soll += orakelZeile(n, 1, 2, 0x00A0)
	}
	soll += orakelZeile(`a\xC3\xA4.md`, 1, 2, 0x00A0) + orakelZeile("b.md", 1, 2, 0x00A0) +
		orakelZeile(`\xC3\xA4.md`, 1, 2, 0x00A0) + summe(9, 9)
	vorher := inhalt(t, dir)
	r1, r2 := fahre(dir), fahre(dir)
	gleich(t, "Reihenfolge", r1.err, soll)
	if r1 != r2 {
		t.Errorf("zweiter Lauf weicht ab: %v gegen %v", r1, r2)
	}
	if nachher := inhalt(t, dir); nachher != vorher {
		t.Errorf("die Wurzel hat sich verändert")
	}
}

// inhalt ist ein Abbild aller Pfade und Inhalte unter dir.
func inhalt(t *testing.T, dir string) string {
	t.Helper()
	var s strings.Builder
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Fprintf(&s, "%s|", p)
		if d.Type().IsRegular() {
			c, err := os.ReadFile(p)
			s.Write(c)
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return s.String()
}

func TestGrosseDatei(t *testing.T) {
	gross := cat(bytes.Repeat(b("xxxxxxxxxxxxxxx\n"), 3*1024*1024/16), zb(0x00A0))
	zeile := cat(xs(200000), zb(0x00A0), b("\n"), zb(0x00A0))
	r := fahre(baum(t, map[string][]byte{"gross.md": gross, "zeile.md": zeile}))
	gleich(t, "große Datei", r.err, orakelZeile("gross.md", 196609, 1, 0x00A0)+
		orakelZeile("zeile.md", 1, 200001, 0x00A0)+orakelZeile("zeile.md", 2, 1, 0x00A0)+summe(3, 2))
}

func TestPaare(t *testing.T) {
	a := cat(b("p"), zb(0x200B), zb(0x00A0), b("q"), zb(0x200B), b("\nr\xE4s"), zb(0x00A0))
	dateien := map[string][]byte{"a.md": a, "b.md": cat(b("u"), zb(0x00A0)), "c.md": cat(b("v"), zb(0x00A0), zb(0x00A0), zb(0x00A0))}
	soll := orakelZeile("a.md", 1, 2, 0x200B) + orakelZeile("a.md", 1, 3, 0x00A0) + orakelZeile("a.md", 1, 5, 0x200B) +
		"zeichen: a.md:2:2: ungültiges UTF-8 (Byte 0xE4)\n" + orakelZeile("a.md", 2, 4, 0x00A0) +
		orakelZeile("b.md", 1, 2, 0x00A0) + orakelZeile("c.md", 1, 2, 0x00A0) + orakelZeile("c.md", 1, 3, 0x00A0) +
		orakelZeile("c.md", 1, 4, 0x00A0) + summe(9, 3)
	gleich(t, "Paare", fahre(baum(t, dateien)).err, soll)
}

func TestSelbstanwendung(t *testing.T) {
	var n int
	err := filepath.WalkDir(".", func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			n++
		}
		return err
	})
	if err != nil || n < 1 {
		t.Fatalf("%d Dateien, %v", n, err)
	}
	r := fahre(".")
	if r.code != 0 || r.err != "" {
		t.Fatalf("Exit %d, stderr %q", r.code, r.err)
	}
	gleich(t, "Selbstanwendung", r.out, okZeile(n, 0, 0))
}
