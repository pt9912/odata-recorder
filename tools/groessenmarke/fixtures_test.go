package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Zeichen jenseits von ASCII entstehen aus Bytewerten (AGENTS.md §3.18): ä ist c3 a4.
const (
	basisB  = "0123456789abcdef0123456789abcdef01234567"
	spitzeB = "fedcba9876543210fedcba9876543210fedcba98"
	kopf5   = "| Nr. | Zusage | F\xc3\xa4nger | Mutation | Klasse |\n|---|---|---|---|---|\n"
)

func nebenlaeufig() string {
	return "nebenl\xc3\xa4ufig"
}

// zeile ist eine Tabellenzeile mit den gegebenen Zellen.
func zeile(c ...string) string {
	return "| " + strings.Join(c, " | ") + " |\n"
}

// tabelleText ist die Tabelle mit a nebenläufigen Zeilen, dann b anderen.
func tabelleText(a, b int) string {
	var sb strings.Builder
	sb.WriteString(kopf5)
	n := 0
	for i := 0; i < a; i++ {
		n++
		sb.WriteString(zeile(fmt.Sprintf("Z%d", n), "z", "f", "m", nebenlaeufig()))
	}
	for i := 0; i < b; i++ {
		n++
		sb.WriteString(zeile(fmt.Sprintf("Z%d", n), "z", "f", "m", "andere"))
	}
	return sb.String()
}

// blockB ist der Block der Fixture B (SPEC-102): 10 andere Zeilen, kein Produktcode, keine Fakes,
// S = 800, Marke 1100. Jedes Argument ersetzt die Zeile desselben Schlüssels oder hängt eine an;
// `-schluessel` entfernt die Zeilen des Schlüssels.
func blockB(aend ...string) []string {
	zeilen := []string{"pfad: src/", "produktcode: 0", "fakes: 0", "basis: " + basisB, "schaetzung: 800", "marke: 1100"}
	for _, a := range aend {
		if k, weg := strings.CutPrefix(a, "-"); weg {
			var rest []string
			for _, z := range zeilen {
				if !strings.HasPrefix(z, k+":") {
					rest = append(rest, z)
				}
			}
			zeilen = rest
			continue
		}
		k, _, _ := strings.Cut(a, ":")
		ersetzt := false
		for i := range zeilen {
			if strings.HasPrefix(zeilen[i], k+":") && !wiederholbar(k) {
				zeilen[i], ersetzt = a, true
				break
			}
		}
		if !ersetzt {
			zeilen = append(zeilen, a)
		}
	}
	return zeilen
}

// planMit ist ein Plan mit der Tabelle tab und dem Block aus den Zeilen.
func planMit(tab string, block []string) string {
	return "# Slice\n\n" + tab + "\n```groessenmarke\n" + strings.Join(block, "\n") + "\n```\n"
}

// planB ist die Fixture B mit Änderungen am Block.
func planB(aend ...string) string {
	return planMit(tabelleText(0, 10), blockB(aend...))
}

// satz ist ein Datensatz von `git diff --numstat -z`.
func satz(ins, del, pfad string) string {
	return ins + "\t" + del + "\t" + pfad + "\x00"
}

// kopfText ist der Kopf der Messung.
func kopfText(basis, spitze string, pfade ...string) string {
	s := "basis " + basis + "\nspitze " + spitze + "\n"
	for _, p := range pfade {
		s += "pfad " + p + "\n"
	}
	return s + "numstat\n"
}

// messungB ist die Messung der Fixture B mit der Summe n (ein Datensatz, bei 0 keiner).
func messungB(n int) string {
	s := kopfText(basisB, spitzeB, "src/")
	if n > 0 {
		s += satz(fmt.Sprint(n), "0", "f.go")
	}
	return s
}

// ergebnis ist Exit-Code und Ausgabe eines Aufrufs.
type ergebnis struct {
	code     int
	out, err string
}

func (e ergebnis) alles() string {
	return e.out + e.err
}

// soll prüft Exit-Code und Texte der Ausgabe; ein Text mit `!` davor darf nicht vorkommen.
func (e ergebnis) soll(t *testing.T, was string, code int, texte ...string) {
	t.Helper()
	if e.code != code {
		t.Errorf("%s: Exit %d statt %d\nstdout: %s\nstderr: %s", was, e.code, code, e.out, e.err)
		return
	}
	for _, x := range texte {
		if nicht, ok := strings.CutPrefix(x, "!"); ok {
			if strings.Contains(e.alles(), nicht) {
				t.Errorf("%s: Ausgabe nennt %q\nstdout: %s\nstderr: %s", was, nicht, e.out, e.err)
			}
		} else if !strings.Contains(e.alles(), x) {
			t.Errorf("%s: Ausgabe nennt %q nicht\nstdout: %s\nstderr: %s", was, x, e.out, e.err)
		}
	}
}

// pr ruft `pruefen` mit dem Plan und der Messung als Text.
func pr(plan, messungText string) ergebnis {
	var o, e bytes.Buffer
	c := pruefen("p.md", plan, []byte(messungText), &o, &e)
	return ergebnis{c, o.String(), e.String()}
}

// ab ruft `abfrage` mit dem Plan als Text.
func ab(plan string) ergebnis {
	var o, e bytes.Buffer
	c := abfrage("p.md", plan, &o, &e)
	return ergebnis{c, o.String(), e.String()}
}

// hat ist wahr, wenn eine Fehlermeldung genau diesen Text trägt.
func hat(fehler []string, text string) bool {
	for _, f := range fehler {
		if f == text {
			return true
		}
	}
	return false
}

// feld liest ein Feld `schluessel=wert` der Messpunkt-Zeile aus der Ausgabe.
func feld(out, schluessel string) string {
	for _, z := range strings.Split(out, "\n") {
		if !strings.HasPrefix(z, "Messpunkt ") {
			continue
		}
		for _, f := range strings.Fields(z) {
			if w, ok := strings.CutPrefix(f, schluessel+"="); ok {
				return w
			}
		}
	}
	return ""
}
