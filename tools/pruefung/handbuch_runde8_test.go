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

// Fänger der achten Runde zu slice-handbuch-geruest (Z94, V-1 bis V-3 der fünften Verifikation; Z95 steht in
// handbuch_verweise_test.go), je als Schleife über eine Menge aus dem Wortlaut der Zusage; die Orakel stehen von Hand im Test.

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
