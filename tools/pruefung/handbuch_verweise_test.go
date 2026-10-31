package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Z95 (Nr. 2 f, 2 c, 3 a, 5 d, 5 f, 8, 9 a): ein Symbolischer Verweis gilt überall als sein Ziel. Menge: 12 Stellen
// (Lese-Aufrufe des Programms) × 4 Arten (Verweis auf Datei, auf Verzeichnis, ins Leere, Kette aus zwei Verweisen
// zur Datei) = 48, dazu die Schleife von Verweisen als Handbuch (1) und die Verzeichnisse selbst als Verweis (3): 52.

const (
	verwMakefile = "build: ## bauen\n\tdocker build --target build -t odata-recorder:build .\n"
	verwDocker   = "RUN go build -o /out/odata-recorder ./x\n"
	verwBindung  = "beispiele/a.txt"
)

// verwSoll ist die Erwartung von Hand: Exit (-1: egal), Teilzeichenketten der Ausgabe, Zahl der Starts (-1: egal).
type verwSoll struct {
	code   int
	da     []string
	nicht  []string
	starts int
}

// verwBeob ist, was ein Lauf zeigt.
type verwBeob struct {
	code   int
	err    string
	starts int
}

// verwLegen legt an lp einen Verweis der Art an; sein Ziel (bei Datei und Kette) ist ziel oder eine neue Datei mit inhalt.
func verwLegen(t *testing.T, m umfeld, art, lp, inhalt, ziel string) {
	t.Helper()
	z := filepath.Join(m.wurzel, "_z")
	for _, d := range []string{z, filepath.Dir(lp)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	datei := ziel
	if datei == "" {
		datei = filepath.Join(z, "f"+filepath.Base(lp))
		if err := os.WriteFile(datei, []byte(inhalt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var err error
	switch art {
	case "datei":
		err = os.Symlink(datei, lp)
	case "kette":
		l1 := filepath.Join(z, "l"+filepath.Base(lp))
		if err = os.Symlink(datei, l1); err == nil {
			err = os.Symlink(l1, lp)
		}
	case "verz":
		err = os.Symlink(z, lp)
	default:
		err = os.Symlink(filepath.Join(z, "gibts-nicht"), lp)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func verwPruefe(t *testing.T, name string, b verwBeob, s verwSoll) {
	t.Helper()
	ok := (s.code < 0 || b.code == s.code) && (s.starts < 0 || b.starts == s.starts)
	for _, d := range s.da {
		ok = ok && strings.Contains(b.err, d)
	}
	for _, d := range s.nicht {
		ok = ok && !strings.Contains(b.err, d)
	}
	if !ok {
		t.Errorf("%s: Exit %d (erwartet %d), %d Starts (erwartet %d), Ausgabe %q (muss %q, darf nicht %q)", name, b.code, s.code, b.starts, s.starts, b.err, s.da, s.nicht)
	}
}

func verwOK(art string) bool { return art == "datei" || art == "kette" }

func verwBinary(t *testing.T, art string) (verwBeob, verwSoll) {
	m := ohneBeispiele(t)
	lp := filepath.Join(m.e.dir, "lnk")
	verwLegen(t, m, art, lp, "", m.e.bin)
	c, _, e := fuehre(m.u, "handbuch", lp, m.wurzel)
	b := verwBeob{c, e, len(m.e.starts(t))}
	if verwOK(art) {
		return b, verwSoll{code: exitOK, starts: 1}
	}
	return b, verwSoll{code: exitAufruf, da: []string{"Eingabe: Binary fehlt: "}, starts: 0}
}

func verwHandbuch(t *testing.T, art string) (verwBeob, verwSoll) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne(handbuchPfad)
	verwLegen(t, m, art, filepath.Join(m.wurzel, handbuchPfad), sauber(), "")
	c, _, e := m.lauf()
	b := verwBeob{c, e, len(m.e.starts(t))}
	switch {
	case verwOK(art):
		return b, verwSoll{code: exitOK, starts: 2}
	case art == "verz":
		return b, verwSoll{code: exitAufruf, da: []string{"Eingabe: Handbuch nicht lesbar: "}, nicht: []string{"Handbuch fehlt"}, starts: 0}
	}
	return b, verwSoll{code: exitAufruf, da: []string{"Eingabe: Handbuch fehlt: "}, starts: 0}
}

// verwBeispiel prüft die Beispiel-Datei a.txt als Verweis an der Stelle 3 (Block), 4 (Eintrag) oder 5 (Lauf).
func verwBeispiel(t *testing.T, art string, stelle int) (verwBeob, verwSoll) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne("docs/user/beispiele/aufruf.txt")
	verwLegen(t, m, art, filepath.Join(m.wurzel, "docs/user/beispiele/a.txt"), aufrufDatei, "")
	h := rumpf(binde(verwBindung) + zaunVon(aufrufDatei))
	if stelle == 4 {
		h = rumpf("Text.\n")
	}
	m.handbuch(h)
	c, _, e := m.lauf()
	b := verwBeob{c, e, len(m.e.starts(t))}
	ok := verwOK(art)
	switch {
	case stelle == 3 && ok:
		return b, verwSoll{code: -1, nicht: []string{"weicht", "a.txt fehlt"}, starts: -1}
	case stelle == 3:
		return b, verwSoll{code: -1, da: []string{"beispiele/a.txt fehlt"}, starts: -1}
	case stelle == 4 && ok:
		return b, verwSoll{code: -1, da: []string{"beispiele: a.txt ohne Bindung"}, nicht: []string{"keine Beispiel-Datei"}, starts: -1}
	case stelle == 4:
		return b, verwSoll{code: -1, da: []string{"beispiele: a.txt keine Beispiel-Datei"}, starts: -1}
	case ok:
		return b, verwSoll{code: -1, starts: 2}
	}
	return b, verwSoll{code: -1, starts: 1}
}

// verwZiel prüft die Zielliste: das Ziel x steht in der Datei rel, die ein Verweis ist; bekannt heißt Exit 0.
func verwZiel(t *testing.T, art, rel string, makefileWeg bool) (verwBeob, verwSoll) {
	m := ohneBeispiele(t)
	if makefileWeg {
		m.entferne("Makefile")
	}
	verwLegen(t, m, art, filepath.Join(m.wurzel, rel), "x: ## x\n", "")
	m.handbuch(rumpf("Nenne `make x`.\n"))
	c, _, e := m.lauf()
	b := verwBeob{c, e, -1}
	if verwOK(art) {
		return b, verwSoll{code: exitOK, starts: -1}
	}
	return b, verwSoll{code: exitBefund, da: []string{"make-Ziel x unbekannt"}, starts: -1}
}

// verwDeckung prüft die Deckung des Aufrufpräfix: die Datei rel mit dem Inhalt ist ein Verweis; gedeckt heißt Exit 0.
func verwDeckung(t *testing.T, art, rel, inhalt, dann string, vorbereite func(m umfeld)) (verwBeob, verwSoll) {
	m := neuesUmfeld(t, standardAntworten())
	vorbereite(m)
	verwLegen(t, m, art, filepath.Join(m.wurzel, rel), inhalt, "")
	c, _, e := m.lauf()
	b := verwBeob{c, e, -1}
	if verwOK(art) {
		return b, verwSoll{code: exitOK, starts: -1}
	}
	return b, verwSoll{code: exitBefund, da: []string{dann}, starts: -1}
}

type verwStelle struct {
	name string
	f    func(t *testing.T, art string) (verwBeob, verwSoll)
}

func verwStellen() []verwStelle {
	bsp := func(s int) func(t *testing.T, art string) (verwBeob, verwSoll) {
		return func(t *testing.T, art string) (verwBeob, verwSoll) { return verwBeispiel(t, art, s) }
	}
	ziel := func(rel string, makefileWeg bool) func(t *testing.T, art string) (verwBeob, verwSoll) {
		return func(t *testing.T, art string) (verwBeob, verwSoll) { return verwZiel(t, art, rel, makefileWeg) }
	}
	deck := func(rel, inhalt, dann string, vorbereite func(m umfeld)) func(t *testing.T, art string) (verwBeob, verwSoll) {
		return func(t *testing.T, art string) (verwBeob, verwSoll) {
			return verwDeckung(t, art, rel, inhalt, dann, vorbereite)
		}
	}
	return []verwStelle{
		{"1 Binary", verwBinary},
		{"2 Handbuch", verwHandbuch},
		{"3 Beispiel-Datei Block", bsp(3)},
		{"4 Beispiel-Datei Eintrag", bsp(4)},
		{"5 Beispiel-Datei Lauf", bsp(5)},
		{"6 Zielliste Makefile", ziel("Makefile", true)},
		{"7 Zielliste d-check.mk", ziel("d-check.mk", false)},
		{"8 Zielliste repo.mk", ziel("repo.mk", false)},
		{"9 Zielliste harness/mk", ziel("harness/mk/x.mk", false)},
		{"10 Image Makefile", deck("Makefile", verwMakefile, "Aufrufpräfix nicht gedeckt: Image", func(m umfeld) { m.entferne("Makefile") })},
		{"11 Image harness/mk", deck("harness/mk/y.mk", verwMakefile, "Aufrufpräfix nicht gedeckt: Image", func(m umfeld) { m.schreibe("Makefile", "build: ## x\n") })},
		{"12 Pfad Dockerfile", deck("Dockerfile", verwDocker, "Aufrufpräfix nicht gedeckt: Pfad", func(m umfeld) { m.entferne("Dockerfile") })},
	}
}

// verwSchleife: eine Schleife von Verweisen als Handbuch ist ein vorhandener, nicht lesbarer Eintrag.
func verwSchleife(t *testing.T) (verwBeob, verwSoll) {
	m := neuesUmfeld(t, standardAntworten())
	m.entferne(handbuchPfad)
	p := filepath.Join(m.wurzel, handbuchPfad)
	q := filepath.Join(m.wurzel, "docs/user/schleife")
	if err := os.Symlink(q, p); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(p, q); err != nil {
		t.Fatal(err)
	}
	c, _, e := m.lauf()
	return verwBeob{c, e, len(m.e.starts(t))}, verwSoll{code: exitAufruf, da: []string{"Eingabe: Handbuch nicht lesbar: "}, nicht: []string{"Handbuch fehlt"}, starts: 0}
}

// verwVerzeichnis ersetzt das Verzeichnis rel durch einen Verweis auf ein Verzeichnis unter _z.
func verwVerzeichnis(t *testing.T, rel string) umfeld {
	t.Helper()
	m := neuesUmfeld(t, standardAntworten())
	ziel := filepath.Join(m.wurzel, "_z", filepath.Base(rel))
	if err := os.MkdirAll(filepath.Dir(ziel), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(m.wurzel, rel), ziel); err != nil {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if err := os.MkdirAll(ziel, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(m.wurzel, rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(ziel, filepath.Join(m.wurzel, rel)); err != nil {
		t.Fatal(err)
	}
	return m
}

// verwBeispieleVerz: docs/user/beispiele ist selbst ein Verweis; a.txt liegt darin, ungebunden (Eintrag) oder gebunden (Lauf).
func verwBeispieleVerz(t *testing.T, gebunden bool) (verwBeob, verwSoll) {
	m := verwVerzeichnis(t, "docs/user/beispiele")
	m.entferne("_z/beispiele/aufruf.txt")
	m.schreibe("_z/beispiele/a.txt", aufrufDatei)
	if gebunden {
		m.handbuch(rumpf(binde(verwBindung) + zaunVon(aufrufDatei)))
		c, _, e := m.lauf()
		return verwBeob{c, e, len(m.e.starts(t))}, verwSoll{code: exitOK, starts: 2}
	}
	m.handbuch(rumpf("Text.\n"))
	c, _, e := m.lauf()
	return verwBeob{c, e, -1}, verwSoll{code: exitBefund, da: []string{"beispiele: a.txt ohne Bindung"}, starts: -1}
}

// verwMkVerz: harness/mk ist selbst ein Verweis; x.mk trägt das Ziel x und die Image-Deckung.
func verwMkVerz(t *testing.T) (verwBeob, verwSoll) {
	m := verwVerzeichnis(t, "harness/mk")
	m.schreibe("Makefile", "build: ## x\n")
	m.schreibe("_z/mk/x.mk", "x: ## x\n"+verwMakefile)
	m.handbuch(sauber() + "\nNenne `make x`.\n")
	c, _, e := m.lauf()
	return verwBeob{c, e, -1}, verwSoll{code: exitOK, starts: -1}
}

func TestHandbuchVerweise(t *testing.T) {
	if err := os.Symlink("x", filepath.Join(t.TempDir(), "probe")); err != nil {
		t.Skipf("kein Symbolischer Verweis erzeugbar: %v", err)
	}
	n := 0
	for _, art := range []string{"datei", "kette", "verz", "leer"} {
		for _, st := range verwStellen() {
			n++
			b, s := st.f(t, art)
			verwPruefe(t, st.name+"/"+art, b, s)
		}
	}
	for name, f := range map[string]func(t *testing.T) (verwBeob, verwSoll){
		"Schleife als Handbuch":          verwSchleife,
		"beispiele als Verweis, Eintrag": func(t *testing.T) (verwBeob, verwSoll) { return verwBeispieleVerz(t, false) },
		"beispiele als Verweis, Lauf":    func(t *testing.T) (verwBeob, verwSoll) { return verwBeispieleVerz(t, true) },
		"harness/mk als Verweis":         verwMkVerz,
	} {
		n++
		b, s := f(t)
		verwPruefe(t, name, b, s)
	}
	if n != 52 {
		t.Fatalf("%d Elemente, erwartet 52", n)
	}
}
