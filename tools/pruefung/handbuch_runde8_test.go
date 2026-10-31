package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Fänger der achten Runde zu slice-handbuch-geruest (Z94, Z95, V-1 bis V-3 der fünften Verifikation), je als
// Schleife über eine Menge aus dem Wortlaut der Zusage; die Orakel stehen von Hand im Test.

// Z94 (Nr. 4 d): der Stand ist ein Tag des proleptischen Gregorianischen Kalenders in jedem Jahr 0000 bis 9999.
// Menge: 10000 Jahre × 12 Monate × {letzter Tag (gültig), Folgetag (ungültig)} = 240000 Elemente; Orakel: die
// Standardbibliothek `time` (nicht standGueltig). Dazu 12 Fälle von Hand, deren Orakel der Wortlaut ist.
func TestHandbuchStandJahre(t *testing.T) {
	n, rot, erstes := 0, 0, ""
	for j := 0; j <= 9999; j++ {
		for m := 1; m <= 12; m++ {
			letzter := time.Date(j, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
			for _, f := range []struct {
				tag     int
				gueltig bool
			}{{letzter, true}, {letzter + 1, false}} {
				n++
				s := fmt.Sprintf("%04d-%02d-%02d", j, m, f.tag)
				if standGueltig(s) != f.gueltig {
					rot++
					if erstes == "" {
						erstes = s
					}
				}
			}
		}
	}
	if n != 240000 {
		t.Fatalf("%d Elemente, erwartet 240000", n)
	}
	if rot > 0 {
		t.Errorf("%d von %d Elementen falsch, das erste: %s", rot, n, erstes)
	}
	hand := []struct {
		s       string
		gueltig bool
	}{
		{"0000-01-01", true}, {"0000-02-29", true}, {"0000-12-31", true},
		{"0000-00-01", false}, {"0000-01-00", false}, {"0000-13-01", false},
		{"0001-02-29", false}, {"1900-02-29", false},
		{"0004-02-29", true}, {"2000-02-29", true}, {"9999-12-31", true}, {"9999-12-32", false},
	}
	if len(hand) != 12 {
		t.Fatalf("%d Hand-Fälle, erwartet 12", len(hand))
	}
	for _, h := range hand {
		if standGueltig(h.s) != h.gueltig {
			t.Errorf("%s: erwartet %v", h.s, h.gueltig)
		}
	}
}

// Z95 (Nr. 2 f, 2 c, 3 a, 5 d, 5 f, 8, 9 a): ein Symbolischer Verweis gilt überall als sein Ziel. Menge: 12 Stellen
// (Lese-Aufrufe des Programms) × 4 Arten (Verweis auf Datei, auf Verzeichnis, ins Leere, Kette aus zwei Verweisen
// zur Datei) = 48, dazu die Schleife von Verweisen als Handbuch (1) und die Verzeichnisse selbst als Verweis (3): 52.
func TestHandbuchVerweise(t *testing.T) {
	if err := os.Symlink("x", filepath.Join(t.TempDir(), "probe")); err != nil {
		t.Skipf("kein Symbolischer Verweis erzeugbar: %v", err)
	}
	const (
		makefile = "build: ## bauen\n\tdocker build --target build -t odata-recorder:build .\n"
		dockerf  = "RUN go build -o /out/odata-recorder ./x\n"
		bz       = "beispiele/a.txt"
	)
	// verweis legt an lp einen Verweis der Art an, dessen Ziel (bei Datei und Kette) den Inhalt trägt oder ziel ist.
	verweis := func(t *testing.T, m umfeld, art, lp, inhalt, ziel string) {
		t.Helper()
		z := filepath.Join(m.wurzel, "_z")
		if err := os.MkdirAll(z, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(lp), 0o755); err != nil {
			t.Fatal(err)
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
	type soll struct {
		code   int
		da     []string
		nicht  []string
		starts int
	}
	prufe := func(t *testing.T, name string, code int, errText string, starts int, s soll) {
		t.Helper()
		ok := (s.code < 0 || code == s.code) && (s.starts < 0 || starts == s.starts)
		for _, d := range s.da {
			ok = ok && strings.Contains(errText, d)
		}
		for _, d := range s.nicht {
			ok = ok && !strings.Contains(errText, d)
		}
		if !ok {
			t.Errorf("%s: Exit %d (erwartet %d), %d Starts (erwartet %d), Ausgabe %q (muss %q, darf nicht %q)", name, code, s.code, starts, s.starts, errText, s.da, s.nicht)
		}
	}
	bsp := func(t *testing.T, m umfeld, art, h string) {
		t.Helper()
		m.entferne("docs/user/beispiele/aufruf.txt")
		verweis(t, m, art, filepath.Join(m.wurzel, "docs/user/beispiele/a.txt"), aufrufDatei, "")
		m.handbuch(h)
	}
	n := 0
	for _, art := range []string{"datei", "kette", "verz", "leer"} {
		ok := art == "datei" || art == "kette"
		stellen := []struct {
			name string
			f    func(t *testing.T) (int, string, int, soll)
		}{
			{"1 Binary", func(t *testing.T) (int, string, int, soll) {
				m := ohneBeispiele(t)
				lp := filepath.Join(m.e.dir, "lnk")
				verweis(t, m, art, lp, "", m.e.bin)
				c, _, e := fuehre(m.u, "handbuch", lp, m.wurzel)
				if ok {
					return c, e, len(m.e.starts(t)), soll{code: exitOK, starts: 1}
				}
				return c, e, len(m.e.starts(t)), soll{code: exitAufruf, da: []string{"Eingabe: Binary fehlt: "}, starts: 0}
			}},
			{"2 Handbuch", func(t *testing.T) (int, string, int, soll) {
				m := neuesUmfeld(t, standardAntworten())
				m.entferne(handbuchPfad)
				verweis(t, m, art, filepath.Join(m.wurzel, handbuchPfad), sauber(), "")
				c, _, e := m.lauf()
				switch {
				case ok:
					return c, e, len(m.e.starts(t)), soll{code: exitOK, starts: 2}
				case art == "verz":
					return c, e, len(m.e.starts(t)), soll{code: exitAufruf, da: []string{"Eingabe: Handbuch nicht lesbar: "}, nicht: []string{"Handbuch fehlt"}, starts: 0}
				}
				return c, e, len(m.e.starts(t)), soll{code: exitAufruf, da: []string{"Eingabe: Handbuch fehlt: "}, starts: 0}
			}},
			{"3 Beispiel-Datei Block", func(t *testing.T) (int, string, int, soll) {
				m := neuesUmfeld(t, standardAntworten())
				bsp(t, m, art, rumpf(binde(bz)+zaunVon(aufrufDatei)))
				c, _, e := m.lauf()
				if ok {
					return c, e, -1, soll{code: -1, nicht: []string{"weicht", "a.txt fehlt"}, starts: -1}
				}
				return c, e, -1, soll{code: -1, da: []string{"beispiele/a.txt fehlt"}, starts: -1}
			}},
			{"4 Beispiel-Datei Eintrag", func(t *testing.T) (int, string, int, soll) {
				m := neuesUmfeld(t, standardAntworten())
				bsp(t, m, art, rumpf("Text.\n"))
				c, _, e := m.lauf()
				if ok {
					return c, e, -1, soll{code: -1, da: []string{"beispiele: a.txt ohne Bindung"}, nicht: []string{"keine Beispiel-Datei"}, starts: -1}
				}
				return c, e, -1, soll{code: -1, da: []string{"beispiele: a.txt keine Beispiel-Datei"}, starts: -1}
			}},
			{"5 Beispiel-Datei Lauf", func(t *testing.T) (int, string, int, soll) {
				m := neuesUmfeld(t, standardAntworten())
				bsp(t, m, art, rumpf(binde(bz)+zaunVon(aufrufDatei)))
				c, _, e := m.lauf()
				if ok {
					return c, e, len(m.e.starts(t)), soll{code: -1, starts: 2}
				}
				return c, e, len(m.e.starts(t)), soll{code: -1, starts: 1}
			}},
		}
		// Make-Dateien: Ziel x in der Zielliste; ok heißt bekannt (Exit 0), sonst `make-Ziel x unbekannt`.
		ziel := func(rel, inhalt, vorher string) func(t *testing.T) (int, string, int, soll) {
			return func(t *testing.T) (int, string, int, soll) {
				m := ohneBeispiele(t)
				if vorher != "" {
					m.entferne("Makefile")
				}
				verweis(t, m, art, filepath.Join(m.wurzel, rel), inhalt, "")
				m.handbuch(rumpf("Nenne `make x`.\n"))
				c, _, e := m.lauf()
				if ok {
					return c, e, -1, soll{code: exitOK, starts: -1}
				}
				return c, e, -1, soll{code: exitBefund, da: []string{"make-Ziel x unbekannt"}, starts: -1}
			}
		}
		// Deckung des Aufrufpräfix: ok heißt gedeckt (Exit 0), sonst der Befund `dann`.
		deckung := func(rel, inhalt, dann string, vorbereite func(m umfeld)) func(t *testing.T) (int, string, int, soll) {
			return func(t *testing.T) (int, string, int, soll) {
				m := neuesUmfeld(t, standardAntworten())
				vorbereite(m)
				if err := os.MkdirAll(filepath.Dir(filepath.Join(m.wurzel, rel)), 0o755); err != nil {
					t.Fatal(err)
				}
				verweis(t, m, art, filepath.Join(m.wurzel, rel), inhalt, "")
				c, _, e := m.lauf()
				if ok {
					return c, e, -1, soll{code: exitOK, starts: -1}
				}
				return c, e, -1, soll{code: exitBefund, da: []string{dann}, starts: -1}
			}
		}
		stellen = append(stellen, []struct {
			name string
			f    func(t *testing.T) (int, string, int, soll)
		}{
			{"6 Zielliste Makefile", ziel("Makefile", "x: ## x\n", "ja")},
			{"7 Zielliste d-check.mk", ziel("d-check.mk", "x: ## x\n", "")},
			{"8 Zielliste repo.mk", ziel("repo.mk", "x: ## x\n", "")},
			{"9 Zielliste harness/mk", ziel("harness/mk/x.mk", "x: ## x\n", "")},
			{"10 Image Makefile", deckung("Makefile", makefile, "Aufrufpräfix nicht gedeckt: Image", func(m umfeld) { m.entferne("Makefile") })},
			{"11 Image harness/mk", deckung("harness/mk/y.mk", makefile, "Aufrufpräfix nicht gedeckt: Image", func(m umfeld) { m.schreibe("Makefile", "build: ## x\n") })},
			{"12 Pfad Dockerfile", deckung("Dockerfile", dockerf, "Aufrufpräfix nicht gedeckt: Pfad", func(m umfeld) { m.entferne("Dockerfile") })},
		}...)
		for _, st := range stellen {
			n++
			code, errText, starts, s := st.f(t)
			prufe(t, st.name+"/"+art, code, errText, starts, s)
		}
	}
	// Eine Schleife von Verweisen als Handbuch ist ein vorhandener, nicht lesbarer Eintrag.
	{
		n++
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
		prufe(t, "Schleife als Handbuch", c, e, len(m.e.starts(t)), soll{code: exitAufruf, da: []string{"Eingabe: Handbuch nicht lesbar: "}, nicht: []string{"Handbuch fehlt"}, starts: 0})
	}
	// Das Verzeichnis selbst ist ein Verweis auf ein Verzeichnis mit Inhalt.
	verzTausch := func(t *testing.T, rel string) umfeld {
		t.Helper()
		m := neuesUmfeld(t, standardAntworten())
		real := filepath.Join(m.wurzel, "_z", filepath.Base(rel))
		if err := os.MkdirAll(filepath.Dir(real), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(filepath.Join(m.wurzel, rel), real); err != nil {
			if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if err := os.MkdirAll(real, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(m.wurzel, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(real, filepath.Join(m.wurzel, rel)); err != nil {
			t.Fatal(err)
		}
		return m
	}
	{
		n++
		m := verzTausch(t, "docs/user/beispiele")
		m.entferne("_z/beispiele/aufruf.txt")
		m.schreibe("_z/beispiele/a.txt", aufrufDatei)
		m.handbuch(rumpf("Text.\n"))
		c, _, e := m.lauf()
		prufe(t, "beispiele als Verweis, Eintrag", c, e, -1, soll{code: exitBefund, da: []string{"beispiele: a.txt ohne Bindung"}, starts: -1})
	}
	{
		n++
		m := verzTausch(t, "docs/user/beispiele")
		m.entferne("_z/beispiele/aufruf.txt")
		m.schreibe("_z/beispiele/a.txt", aufrufDatei)
		m.handbuch(rumpf(binde(bz) + zaunVon(aufrufDatei)))
		c, _, e := m.lauf()
		prufe(t, "beispiele als Verweis, Lauf", c, e, len(m.e.starts(t)), soll{code: exitOK, starts: 2})
	}
	{
		n++
		m := verzTausch(t, "harness/mk")
		m.schreibe("Makefile", "build: ## x\n")
		m.schreibe("_z/mk/x.mk", "x: ## x\n"+makefile)
		m.handbuch(sauber() + "\nNenne `make x`.\n")
		c, _, e := m.lauf()
		prufe(t, "harness/mk als Verweis", c, e, -1, soll{code: exitOK, starts: -1})
	}
	if n != 52 {
		t.Fatalf("%d Elemente, erwartet 52", n)
	}
}

// V-1 (Nr. 2 c): ein Binary, das keine reguläre Datei ist, ist `Binary fehlt`. Menge aus dem Wortlaut: jede Dateiart, die
// kein regulärer Eintrag ist und kein Verzeichnis: Gerät, FIFO, Socket, und je ein Verweis darauf: 6 Elemente.
func TestHandbuchBinaryKeineDatei(t *testing.T) {
	d := t.TempDir()
	fifo := filepath.Join(d, "fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("kein FIFO erzeugbar: %v", err)
	}
	sock := filepath.Join(d, "sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Skipf("kein Socket erzeugbar: %v", err)
	}
	defer func() { _ = l.Close() }()
	arten := []struct{ name, pfad string }{{"Gerät", "/dev/null"}, {"FIFO", fifo}, {"Socket", sock}}
	var faelle []struct{ name, pfad string }
	for _, a := range arten {
		faelle = append(faelle, a)
		v := filepath.Join(d, "v-"+a.name)
		if err := os.Symlink(a.pfad, v); err != nil {
			t.Skipf("kein Symbolischer Verweis erzeugbar: %v", err)
		}
		faelle = append(faelle, struct{ name, pfad string }{"Verweis auf " + a.name, v})
	}
	if len(faelle) != 6 {
		t.Fatalf("%d Elemente, erwartet 6", len(faelle))
	}
	m := neuesUmfeld(t, standardAntworten())
	for _, f := range faelle {
		code, _, errText := fuehre(m.u, "handbuch", f.pfad, m.wurzel)
		if code != exitAufruf || !strings.Contains(errText, "Eingabe: Binary fehlt: "+f.pfad) || len(m.e.starts(t)) != 0 {
			t.Errorf("%s: Exit %d, Ausgabe %q", f.name, code, errText)
		}
	}
}

// V-2 (Nr. 4 c, 3 b): die Version *V* in `odata-recorder <V>` enthält keinen Leerraum. Menge: die 24 Zeichen, die
// unicode.IsSpace trifft (ohne Zeilenumbruch), an drei Stellen von V (vorn, innen, hinten) = 72 Elemente.
func TestHandbuchVersionOhneLeerraum(t *testing.T) {
	lr := leerraum(t)
	n := 0
	for _, x := range lr {
		for _, v := range []string{x + "dev", "d" + x + "ev", "dev" + x} {
			n++
			m := neuesUmfeld(t, map[string]antwort{"version": {out: "odata-recorder " + v + "\n"}, "": {out: "odata-recorder dev\n"}})
			m.entferne("docs/user/beispiele")
			m.handbuch(rumpf("Text.\n"))
			code, _, errText := m.lauf()
			if code != exitBefund || !strings.Contains(errText, "version: Ausgabe nicht lesbar: Standardausgabe ") {
				t.Errorf("V %q: Exit %d, Ausgabe %q", v, code, errText)
			}
		}
	}
	if n != 72 {
		t.Fatalf("%d Elemente, erwartet 72", n)
	}
}

// V-3 (Nr. 9 b): ein Span hinter einem Span aus n >= 2 Backticks wird gefunden. Menge: erster Span mit 2 bis 5 Backticks
// (Inhalt `a` oder `a b`) × zweiter Span mit 1 bis 4 Backticks × Trenner {Leerzeichen, ` und `, Zeilenumbruch} × Inhalt des
// zweiten {`make nix` (ein Befund), `make build` (keiner)} = 4 × 2 × 4 × 3 × 2 = 192 Elemente; Orakel von Hand.
func TestHandbuchSpanNachMehrfachSpan(t *testing.T) {
	ziele := map[string]bool{"build": true}
	n := 0
	for n1 := 2; n1 <= 5; n1++ {
		for _, inhalt1 := range []string{"a", "a b"} {
			for n2 := 1; n2 <= 4; n2++ {
				for _, sep := range []string{" ", " und ", "\n"} {
					for _, inhalt2 := range []string{"make nix", "make build"} {
						n++
						b1, b2 := strings.Repeat("`", n1), strings.Repeat("`", n2)
						text := b1 + inhalt1 + b1 + sep + b2 + inhalt2 + b2
						zeile := 1
						if sep == "\n" {
							zeile = 2
						}
						var soll []string
						if inhalt2 == "make nix" {
							soll = []string{fmt.Sprintf("Zeile %d: make-Ziel nix unbekannt", zeile)}
						}
						var ist []string
						for _, b := range makeBefunde([]absatz{{zeile: 1, zeilen: strings.Split(text, "\n")}}, ziele) {
							ist = append(ist, b.text)
						}
						if !gleichListe(ist, soll) {
							t.Errorf("%q: Befunde %q, erwartet %q", text, ist, soll)
						}
					}
				}
			}
		}
	}
	if n != 192 {
		t.Fatalf("%d Elemente, erwartet 192", n)
	}
}
