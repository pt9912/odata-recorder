package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
)

// Fänger der sechsten Runde zu slice-handbuch-geruest (Z83 bis Z89), je als Schleife über die
// Menge der Zusage; die Orakel stehen von Hand im Test, nie im Code.

// zeitErsatz ist ein Ersatz-Binary, dessen Lauf je Aufrufstelle (version, Beispiel) in Dateien
// neben dem Skript steht: wie lange es bis zur Ausgabe wartet, ob es über jede Grenze schläft
// (ersetzt sich durch sleep) und wie lange ein Kindprozess die Ströme hält.
func zeitErsatz(t *testing.T, stelle, b, k string, bleibt bool) string {
	t.Helper()
	d := t.TempDir()
	s := "e"
	if stelle == "version" {
		s = "v"
	}
	dateien := map[string]string{}
	if b != "" {
		dateien[s+".b"] = b
	}
	if k != "" {
		dateien[s+".k"] = k
	}
	if bleibt {
		dateien[s+".bleibt"] = ""
	}
	for n, inhalt := range dateien {
		if err := os.WriteFile(filepath.Join(d, n), []byte(inhalt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	skript := `#!/bin/sh
d=${0%/*}
if [ "$*" = version ]; then s=v; else s=e; fi
[ -f "$d/$s.k" ] && { sleep "$(cat "$d/$s.k")" & }
[ -f "$d/$s.bleibt" ] && exec sleep 30
[ -f "$d/$s.b" ] && sleep "$(cat "$d/$s.b")"
if [ $s = v ]; then echo "odata-recorder dev"; exit 0; fi
echo hallo
exit 3
`
	bin := filepath.Join(d, "odata-recorder")
	if err := os.WriteFile(bin, []byte(skript), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

// Z83: Die Zeitgrenze gilt dem Binary, nicht dem Warten (SPEC-200 Nr. 7 c, 7 e, 4 c). Wirkungs-Fänger:
// Grenze 2 s, Ende des Binary B und Ende des Kindprozesses K, 8 Zellen mal beide Aufrufstellen (16
// Elemente); Befund ja oder nein, Ausgabe und Ausgangswert, Dauer. Abstand zu jeder Grenze mindestens 0,3 s.
func TestHandbuchZeitgrenzeUndWarten(t *testing.T) {
	type zelle struct {
		name   string
		b, k   string
		bleibt bool
		befund bool
		dauer  float64 // Sekunden, ungefähr
	}
	zellen := []zelle{
		{"B1/kein Kind", "0.2", "", false, false, 0.2},
		{"B1/K1", "0.2", "1.0", false, false, 1.0},
		{"B1/K3", "0.2", "6", false, false, 1.2},
		{"B2/kein Kind", "1.7", "", false, false, 1.7},
		{"B2/K2", "1.7", "2.35", false, false, 2.35},
		{"B2/K3", "1.7", "6", false, false, 2.7},
		{"B3/kein Kind", "", "", true, true, 2.0},
		{"B3/Kind", "", "30", true, true, 3.0},
	}
	n := 0
	for _, z := range zellen {
		for _, stelle := range []string{"version", "Beispiel"} {
			n++
			t.Run(z.name+"/"+stelle, func(t *testing.T) {
				t.Parallel()
				m := bspUmfeld(t, "k.txt", bspDatei(praef, []string{"hallo"}, nil, "3\n"), okAntworten())
				m.e.bin = zeitErsatz(t, stelle, z.b, z.k, z.bleibt)
				m.u.Zeitgrenze = 2 * time.Second
				t0 := time.Now()
				code, _, errText := m.lauf()
				dauer := time.Since(t0).Seconds()
				text := "k.txt: Zeitgrenze überschritten"
				if stelle == "version" {
					text = "version: Zeitgrenze überschritten"
				}
				if z.befund {
					erwarte(t, code, errText, exitBefund, "pruefung handbuch: "+text)
					nichtEnthalten(t, errText, "weicht ab", "Ausgabe nicht lesbar")
				} else {
					if code != exitOK {
						t.Fatalf("Exit %d, Ausgabe %q (kein Befund erwartet; Ausgabe und Ausgangswert gelten)", code, errText)
					}
				}
				if dauer < z.dauer-0.05 || dauer > z.dauer+0.6 {
					t.Fatalf("Dauer %.2f s, erwartet %.2f s bis %.2f s", dauer, z.dauer-0.05, z.dauer+0.6)
				}
			})
		}
	}
	if n != 16 {
		t.Fatalf("%d Elemente, erwartet 16", n)
	}
}

// leerraum ist die Menge der Zeichen, die in einer Zeile Leerraum nach SPEC-200 Nr. 3 b sind: jedes
// Zeichen, das unicode.IsSpace trifft, ohne den Zeilenumbruch. Die Werte stehen von Hand im Test;
// die Schleife über alle Runen und die Zählprüfung halten sie gegen die Standardbibliothek.
func leerraum(t *testing.T) []string {
	t.Helper()
	hand := []rune{0x09, 0x0B, 0x0C, 0x0D, 0x20, 0x85, 0xA0, 0x1680}
	for r := rune(0x2000); r <= 0x200A; r++ {
		hand = append(hand, r)
	}
	hand = append(hand, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000)
	n := 0
	for r := rune(0); r <= 0x10FFFF; r++ {
		if unicode.IsSpace(r) && r != '\n' {
			n++
		}
	}
	if n != 24 || len(hand) != 24 {
		t.Fatalf("Leerraum-Menge: %d aus unicode.IsSpace, %d von Hand, erwartet 24", n, len(hand))
	}
	out := make([]string, 0, 24)
	for _, r := range hand {
		if !unicode.IsSpace(r) {
			t.Fatalf("U+%04X ist kein Leerraum", r)
		}
		out = append(out, string(r))
	}
	return out
}

// keinLeerraum sind Zeichen, die kein Leerraum sind, aber wie Leerraum aussehen oder wirken.
func keinLeerraum() []string {
	return []string{string(rune(0x180E)), string(rune(0x200B)), string(rune(0x2060)), string(rune(0xFEFF))}
}

func makeAbsatz(text string) []absatz {
	return []absatz{{zeile: 1, zeilen: []string{"`" + text + "`"}}}
}

// Z84: Wörter des Spans trennt jeder Leerraum nach Nr. 3 b, vor dem Ziel wie hinter ihm; nur das
// Zeichen unmittelbar hinter make muss U+0020 sein (SPEC-200 Nr. 9 b, 14 d). 236 Elemente.
func TestHandbuchMakeWort(t *testing.T) {
	lr := leerraum(t)
	gegen := keinLeerraum()
	for _, g := range gegen {
		if unicode.IsSpace([]rune(g)[0]) {
			t.Fatalf("%q ist Leerraum", g)
		}
	}
	ziele := map[string]bool{"build": true}
	arten := []struct{ name, zweites, drittes string }{
		{"bekannt", "build", "nix"},
		{"unbekannt", "nix", "build"},
		{"Option", "-n", "build"},
		{"Zuweisung", "V=1", "build"},
	}
	// sollFuer ist das Orakel von Hand: der Befund je Art des zweiten Wortes.
	sollFuer := func(art, text string) []string {
		switch art {
		case "bekannt":
			return nil
		case "unbekannt":
			return []string{"Zeile 1: make-Ziel nix unbekannt"}
		}
		return []string{"Zeile 1: make-Aufruf ohne Ziel an erster Stelle: " + text}
	}
	n := 0
	pruefe := func(name, text string, soll []string) {
		t.Helper()
		n++
		var ist []string
		for _, b := range makeBefunde(makeAbsatz(text), ziele) {
			if b.nr != nrMake {
				t.Fatalf("%s: Befund anderer Nummer %+v", name, b)
			}
			ist = append(ist, b.text)
		}
		if strings.Join(ist, "|") != strings.Join(soll, "|") {
			t.Errorf("%s: Text %q: Befunde %q, erwartet %q", name, text, ist, soll)
		}
	}
	// Abstände vor dem Ziel hinter dem einen Leerzeichen nach make: 23 einzelne Zeichen außer U+0020,
	// Folgen, die mit diesem Leerzeichen zusammen 2, 3 und 5 Leerzeichen lang sind, eine Mischfolge.
	var abstaende []string
	for _, x := range lr {
		if x != " " {
			abstaende = append(abstaende, x)
		}
	}
	abstaende = append(abstaende, " ", "  ", "    ", string(rune(0xA0))+"\t ")
	if len(abstaende) != 27 {
		t.Fatalf("%d Abstände, erwartet 27", len(abstaende))
	}
	for _, a := range arten {
		for _, ab := range abstaende {
			text := "make " + ab + a.zweites + " " + a.drittes
			pruefe(fmt.Sprintf("vor/%s/%q", a.name, ab), text, sollFuer(a.name, text))
		}
		for _, x := range lr {
			text := "make " + a.zweites + x + a.drittes
			pruefe(fmt.Sprintf("hinter/%s/%q", a.name, x), text, sollFuer(a.name, text))
		}
	}
	// Gegenmenge: ein Zeichen ohne Leerraum-Eigenschaft trennt nicht und gehört zum Wort.
	for _, g := range gegen {
		pruefe(fmt.Sprintf("Gegenmenge vor/%q", g), "make "+g+"build nix", []string{"Zeile 1: make-Ziel " + g + "build unbekannt"})
		pruefe(fmt.Sprintf("Gegenmenge hinter/%q", g), "make build"+g+"foo", []string{"Zeile 1: make-Ziel build" + g + "foo unbekannt"})
	}
	// Nur Leerraum hinter make und dem Leerzeichen: kein Aufruf, kein Befund.
	for _, x := range lr {
		pruefe(fmt.Sprintf("nur Leerraum/%q", x), "make "+x, nil)
	}
	if n != 236 {
		t.Fatalf("%d Elemente, erwartet 236", n)
	}
}

// Z85: Bei beispiele: <Name> ... ist der Name der Name des Eintrags, verglichen als Bytefolge
// (SPEC-200 Nr. 2 d). 8 Namen, alle 28 Paare je als eigenes Verzeichnis, dazu ein Lauf mit allen 8.
func TestHandbuchReihenfolgeNamenBytefolge(t *testing.T) {
	zeichen := []string{"0", "A", "Z", "_", "a", "z", string(rune(0x00C4)), string(rune(0x00E4))}
	if len(zeichen[6]) != 2 || zeichen[6][0] != 0xC3 || zeichen[6][1] != 0x84 || zeichen[7][1] != 0xA4 {
		t.Fatalf("Bytes von U+00C4/U+00E4: % x % x", zeichen[6], zeichen[7])
	}
	gueltig := map[string]bool{"0": true, "a": true, "z": true}
	menge := func(namen []string) {
		t.Helper()
		m := ohneBeispiele(t)
		for _, s := range namen {
			m.schreibe("docs/user/beispiele/"+s+".txt", "x\n")
		}
		m.schreibe("Makefile", "build: ## bauen\n\tdocker build --target build -t odata-recorder:build .\n")
		soll := append([]string(nil), namen...)
		sort.Strings(soll)
		var sollZeilen []string
		for _, s := range soll {
			text := "keine Beispiel-Datei"
			if gueltig[s] {
				text = "ohne Bindung"
			}
			sollZeilen = append(sollZeilen, "pruefung handbuch: beispiele: "+s+".txt "+text)
		}
		code, _, errText := m.lauf()
		var ist []string
		for _, z := range strings.Split(strings.TrimRight(errText, "\n"), "\n") {
			if strings.HasPrefix(z, "pruefung handbuch: beispiele: ") {
				ist = append(ist, z)
			}
		}
		if code != exitBefund || strings.Join(ist, "\n") != strings.Join(sollZeilen, "\n") {
			t.Errorf("Namen %q: Exit %d, Zeilen %q, erwartet %q", namen, code, ist, sollZeilen)
		}
	}
	n := 0
	for i := range zeichen {
		for j := i + 1; j < len(zeichen); j++ {
			menge([]string{zeichen[i], zeichen[j]})
			n++
		}
	}
	menge(zeichen)
	n++
	if n != 29 {
		t.Fatalf("%d Elemente, erwartet 29", n)
	}
}

// Z86: Der Text Kopf: Schlüssel <k> unbekannt nennt den getrimmten Schlüssel in der Schreibung der
// Zeile (SPEC-200 Nr. 4 a). 24 Leerraum-Zeichen mal drei Stellungen, dazu Leerraum im Inneren und
// drei Schreibungen: 76 Elemente.
func TestHandbuchKopfSchluesselText(t *testing.T) {
	lr := leerraum(t)
	type el struct{ zeile, k string }
	var els []el
	for _, x := range lr {
		els = append(els, el{x + "Autor: x", "Autor"}, el{"Autor" + x + ": x", "Autor"}, el{x + "Autor" + x + ": x", "Autor"})
	}
	els = append(els, el{"Auto r: x", "Auto r"})
	for _, k := range []string{"Autor", "autor", string(rune(0x00C4)) + "nderungsdatum"} {
		els = append(els, el{k + ": x", k})
	}
	if len(els) != 76 {
		t.Fatalf("%d Elemente, erwartet 76", len(els))
	}
	for _, e := range els {
		werte := map[string][]string{}
		zeilen := map[string]int{}
		bf := leseKopfZeilen([]string{"# Titel", e.zeile}, werte, zeilen)
		soll := "Zeile 2: Kopf: Schlüssel " + e.k + " unbekannt"
		if len(bf) != 1 || bf[0].text != soll {
			t.Errorf("Zeile %q: Befunde %+v, erwartet genau %q", e.zeile, bf, soll)
		}
	}
}

// Z87: Eine nicht leere Info-Zeichenkette ändert die Bindung nicht (SPEC-200 Nr. 5 b, 5 e): Info
// mal Zustand der Bindung, 15 Elemente; Orakel von Hand.
func TestHandbuchInfoGebunden(t *testing.T) {
	inhalt := []string{"geplant", "spec-110", "`make nix`"}
	datei := bspDatei(praef, inhalt, nil, "0\n")
	abweichend := bspDatei(praef, append(append([]string(nil), inhalt...), "x"), nil, "0\n")
	n := 0
	for _, info := range []string{"text", "a b", "sh"} {
		for zustand := 1; zustand <= 5; zustand++ {
			n++
			info, zustand := info, zustand
			t.Run(fmt.Sprintf("Info %q Zustand %d", info, zustand), func(t *testing.T) {
				a := map[string]antwort{"version": {out: "odata-recorder dev\n"}, "": {out: strings.Join(inhalt, "\n") + "\n"}}
				m := neuesUmfeld(t, a)
				m.entferne("docs/user/beispiele/aufruf.txt")
				block := datei
				vor := binde("beispiele/k.txt")
				switch zustand {
				case 1:
					m.schreibe("docs/user/beispiele/k.txt", datei)
				case 2:
					m.schreibe("docs/user/beispiele/k.txt", datei)
					block = abweichend
				case 4:
					vor = "Datei: [beispiele/K.txt](beispiele/K.txt)\n"
				case 5:
					vor = "Text.\n\n"
				}
				kopf := rumpf(vor)
				zeile := strings.Count(kopf, "\n") + 1
				m.handbuch(kopf + "```" + info + "\n" + block + "```\n")
				code, _, errText := m.lauf()
				name := fmt.Sprintf("Info %q Zustand %d", info, zustand)
				infoBefund := fmt.Sprintf("pruefung handbuch: Zeile %d: Info-Zeichenkette %s nicht zulässig", zeile, info)
				if code != exitBefund || strings.Count(errText, "Info-Zeichenkette") != 1 || !strings.Contains(errText, infoBefund) {
					t.Fatalf("%s: Exit %d, Ausgabe %q", name, code, errText)
				}
				kennung := "pruefung handbuch: Zeile " + strconv.Itoa(zeile) + ": Kennung C6: spec-110"
				prosa := []string{"Planung A1: geplant", "Kennung C6: spec-110"}
				starts := len(m.e.starts(t))
				zeilenAnz := len(strings.Split(strings.TrimRight(errText, "\n"), "\n"))
				switch zustand {
				case 1:
					nichtEnthalten(t, errText, "weicht", "k.txt ohne Bindung", "Planung A1", "make-Ziel", "fehlt", "Bindung ungültig")
					enthaelt2(t, errText, kennung)
					if starts != 2 || zeilenAnz != 2 {
						t.Fatalf("%s: %d Starts, %d Zeilen: %q", name, starts, zeilenAnz, errText)
					}
				case 2:
					enthaelt2(t, errText, "Beispiel weicht von beispiele/k.txt ab")
					nichtEnthalten(t, errText, "k.txt ohne Bindung")
					if starts != 2 || strings.Index(errText, "Info-Zeichenkette") > strings.Index(errText, "weicht") {
						t.Fatalf("%s: %d Starts, Ausgabe %q", name, starts, errText)
					}
					enthaelt2(t, errText, prosa...)
				default:
					text := map[int]string{3: "beispiele/k.txt fehlt", 4: "Bindung ungültig: Datei: [beispiele/K.txt]", 5: "Beispiel ohne Bindungszeile"}[zustand]
					enthaelt2(t, errText, text)
					nichtEnthalten(t, errText, "weicht", "k.txt ohne Bindung")
					if starts != 1 || strings.Index(errText, "Info-Zeichenkette") > strings.Index(errText, text) {
						t.Fatalf("%s: %d Starts, Ausgabe %q", name, starts, errText)
					}
					enthaelt2(t, errText, prosa...)
				}
			})
		}
	}
	if n != 15 {
		t.Fatalf("%d Elemente, erwartet 15", n)
	}
}

// Z88: Eine Zeile Datei: [ im Inneren eines geschlossenen Zaunes ist keine Bindungszeile ohne Zaun,
// gleich ob der Block gebunden ist; im nicht geschlossenen Zaun ist sie es (SPEC-200 Nr. 5 c). 18 Elemente.
func TestHandbuchBindungszeileLage(t *testing.T) {
	const bz = "Datei: [beispiele/a.txt](beispiele/a.txt)"
	const ohneZaun = "Bindungszeile ohne Zaun"
	// zeileVon ist die Zeilennummer (ab 1) des k-ten Vorkommens (ab 0) der Zeile bz in h.
	zeileVon := func(h string, k int) int {
		for i, z := range strings.Split(h, "\n") {
			if z == bz {
				if k == 0 {
					return i + 1
				}
				k--
			}
		}
		t.Fatalf("Zeile fehlt in %q", h)
		return 0
	}
	n := 0
	// 12 im Inneren eines geschlossenen Zaunes: Stelle mal Zustand des Blocks. Beim gebundenen Block
	// steht die Zeile in der Standardausgabe (erste, mittlere, letzte Zeile der Ausgabe), sonst ist
	// sie die erste, mittlere, letzte Innenzeile des Zaunes.
	for _, zustand := range []string{"gebunden", "Datei fehlt", "Bindungszeile fehlt", "Bindung ungültig"} {
		for stelle := 0; stelle < 3; stelle++ {
			n++
			name := fmt.Sprintf("%s/Stelle %d", zustand, stelle)
			var block, vor string
			m := neuesUmfeld(t, standardAntworten())
			m.entferne("docs/user/beispiele/aufruf.txt")
			if zustand == "gebunden" {
				xy := []string{"x", "y"}
				zeilen := append(append(append([]string(nil), xy[:stelle]...), bz), xy[stelle:]...)
				block = bspDatei(praef, zeilen, nil, "0\n")
				m.schreibe("docs/user/beispiele/a.txt", block)
				m.e = neuesErsatz(t, map[string]antwort{"version": {out: "odata-recorder dev\n"}, "": {out: strings.Join(zeilen, "\n") + "\n"}})
				vor = binde("beispiele/a.txt")
			} else {
				innen := []string{"p1", "p2", "p3", "p4", "p5"}
				innen[stelle*2] = bz
				block = strings.Join(innen, "\n") + "\n"
				switch zustand {
				case "Datei fehlt":
					vor = binde("beispiele/a.txt")
				case "Bindungszeile fehlt":
					vor = "Text.\n\n"
				default:
					vor = "Datei: [beispiele/A.txt](beispiele/A.txt)\n"
				}
			}
			m.handbuch(rumpf(vor + "```\n" + block + "```\n"))
			code, _, errText := m.lauf()
			if strings.Contains(errText, ohneZaun) {
				t.Errorf("%s: Befund %q: %q", name, ohneZaun, errText)
			}
			if (zustand == "gebunden") != (code == exitOK) {
				t.Errorf("%s: Exit %d, Ausgabe %q", name, code, errText)
			}
		}
	}
	// 2 im nicht geschlossenen Zaun: erste Zeile nach der Öffnung, spätere Zeile.
	for _, inneres := range []string{bz + "\ntext\n", "text\n" + bz + "\n"} {
		n++
		t.Run("nicht geschlossen/"+strconv.Quote(inneres), func(t *testing.T) {
			h := rumpf("```\n" + inneres)
			m := ohneBeispiele(t)
			m.handbuch(h)
			_, _, errText := m.lauf()
			enthaelt2(t, errText, fmt.Sprintf("Zeile %d: %s", zeileVon(h, 0), ohneZaun), "nicht geschlossen")
		})
	}
	// 1 im nicht geschlossenen Rest, unmittelbar vor einer zweiten Öffnungszeile mit Info.
	n++
	t.Run("vor zweiter Öffnungszeile", func(t *testing.T) {
		h := rumpf("```\ntext\n" + bz + "\n```text\n")
		m := ohneBeispiele(t)
		m.handbuch(h)
		_, _, errText := m.lauf()
		enthaelt2(t, errText, fmt.Sprintf("Zeile %d: %s", zeileVon(h, 0), ohneZaun), "nicht geschlossen")
	})
	// 1 unmittelbar vor der Öffnungszeile eines nicht geschlossenen Zaunes: deren Bindungszeile.
	n++
	t.Run("vor nicht geschlossenem Zaun", func(t *testing.T) {
		m := ohneBeispiele(t)
		m.handbuch(rumpf(bz + "\n```\ntext\n"))
		_, _, errText := m.lauf()
		enthaelt2(t, errText, "nicht geschlossen")
		nichtEnthalten(t, errText, ohneZaun)
	})
	// 1 unmittelbar vor der Öffnungszeile eines geschlossenen Zaunes.
	n++
	t.Run("vor geschlossenem Zaun", func(t *testing.T) {
		m := ohneBeispiele(t)
		m.handbuch(rumpf(bz + "\n```\ntext\n```\n"))
		_, _, errText := m.lauf()
		enthaelt2(t, errText, "beispiele/a.txt fehlt")
		nichtEnthalten(t, errText, ohneZaun)
	})
	// 1 unmittelbar nach der Schlusszeile eines geschlossenen Zaunes, vor Prosa.
	n++
	t.Run("nach der Schlusszeile", func(t *testing.T) {
		h := rumpf(bz + "\n```\ntext\n```\n" + bz + "\n\nProsa.\n")
		m := ohneBeispiele(t)
		m.handbuch(h)
		_, _, errText := m.lauf()
		enthaelt2(t, errText, fmt.Sprintf("Zeile %d: %s", zeileVon(h, 1), ohneZaun))
		if c := strings.Count(errText, ohneZaun); c != 1 {
			t.Fatalf("%d Befunde %q: %q", c, ohneZaun, errText)
		}
	})
	if n != 18 {
		t.Fatalf("%d Elemente, erwartet 18", n)
	}
}

// Z89: Ein Treffer ist der des Ausdrucks an der ersten Stelle, an der er trifft; er gilt mit beiden
// Rändern, sonst kein Treffer an dieser Stelle und keine kürzere Suche; die Suche geht ein Zeichen
// weiter (SPEC-200 Nr. 13 b). 3 Formen mal 4 Anfänge mal 5 Enden, dazu 7 Zweitversuche: 67 Elemente.
func TestHandbuchTrefferRand(t *testing.T) {
	re := regexp.MustCompile(`lh-[a-z]{2}(-[a-z]+)?-\d+`)
	ae := string(rune(0x00E4))
	if ae != "\xc3\xa4" {
		t.Fatalf("U+00E4: % x", ae)
	}
	formen := []string{"lh-fa-01", "lh-qa-sup-002", "lh-rb-03"}
	anfaenge := []string{"", " ", "x", ae}
	enden := []string{"", " ", ".", "x", ae}
	n, treffer := 0, 0
	for _, f := range formen {
		for _, a := range anfaenge {
			for _, e := range enden {
				n++
				text := a + f + e
				want := ""
				if (a == "" || a == " ") && (e == "" || e == " " || e == ".") {
					want = f
					treffer++
				}
				got, ok := trifft(re, text)
				if got != want || ok != (want != "") {
					t.Errorf("Text %q: Treffer %q (%v), erwartet %q", text, got, ok, want)
				}
			}
		}
	}
	if treffer != 18 || n != 60 {
		t.Fatalf("%d Elemente, %d Treffer, erwartet 60 und 18", n, treffer)
	}
	for _, f := range formen {
		for _, text := range []string{f + "x " + f, "x" + f + " " + f} {
			n++
			if got, ok := trifft(re, text); !ok || got != f {
				t.Errorf("Zweitversuch %q: Treffer %q (%v), erwartet %q", text, got, ok, f)
			}
		}
	}
	n++
	if got, ok := trifft(re, "xlh-lh-fa-01"); !ok || got != "lh-fa-01" {
		t.Errorf("überlappend: Treffer %q (%v), erwartet lh-fa-01", got, ok)
	}
	if n != 67 {
		t.Fatalf("%d Elemente, erwartet 67", n)
	}
}
