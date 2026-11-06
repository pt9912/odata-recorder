package main

import (
	"errors"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"
)

// Die Dateien unter der Wurzel (SPEC-201 Nr. 10 a).
const (
	pfadFehler     = "src/hexagon/model/fehler.go"
	pfadSpez       = "spec/spezifikation.md"
	pfadAusnahmen  = "harness/katalog-ausnahmen.txt"
	pfadHandbuch   = "docs/user/benutzerhandbuch.md"
	ueberschriftK  = "4. Fehler-Codes und Logging-Felder"
	abschnittCodes = "Fehlercodes"
)

// lies liest eine Datei unter der Wurzel; fehlt sie, oder ist eine Komponente des Pfads eine Datei, ist das
// `<Datei> fehlt: <Pfad>`, sonst `<Datei> nicht lesbar: <Grund>` (SPEC-201 Nr. 8).
func lies(wurzel, pfad, datei string, b *bericht) (string, bool) {
	roh, err := os.ReadFile(filepath.Join(wurzel, filepath.FromSlash(pfad)))
	switch {
	case err == nil:
		return string(roh), true
	case errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR):
		b.fehler(datei + " fehlt: " + pfad)
	default:
		grund := err.Error()
		var pe *fs.PathError
		if errors.As(err, &pe) {
			grund = pe.Err.Error()
		}
		b.fehler(datei + " nicht lesbar: " + grund)
	}
	return "", false
}

// liesC liest die Tabelle im Quelltext: jedes Zeichenketten-Literal, dessen Wert mit `ODR-` beginnt, ist ein
// Code, wenn es die Form trifft, sonst ein Befund Form (SPEC-201 Nr. 3 a).
func liesC(src string, b *bericht) menge {
	datei := token.NewFileSet().AddFile("fehler.go", -1, len(src))
	var s scanner.Scanner
	fehlerZeile := 0
	s.Init(datei, []byte(src), func(p token.Position, _ string) {
		if fehlerZeile == 0 {
			fehlerZeile = p.Line
		}
	}, 0)
	codes := neueNamen("fehler.go")
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.STRING {
			continue
		}
		v, err := strconv.Unquote(lit)
		zeile := datei.PositionFor(pos, false).Line
		if err != nil {
			if fehlerZeile == 0 {
				fehlerZeile = zeile
			}
			break
		}
		if !strings.HasPrefix(v, "ODR-") {
			continue
		}
		if !istCode(v) {
			b.befund(artForm, v, zeile, formText("fehler.go", zeile, v))
			continue
		}
		codes.add(b, v, zeile)
	}
	switch {
	case fehlerZeile > 0:
		b.fehler("fehler.go nicht lesbar: kein Go (Zeile " + strconv.Itoa(fehlerZeile) + ")")
	case len(codes.erste) == 0:
		b.fehler("fehler.go: kein Code gefunden")
	}
	return codes.erste
}

// liesK liest den Katalog der Spezifikation: die Codes der zweiten Zelle der Tabelle in §4 (SPEC-201 Nr. 3 b).
func liesK(text string, b *bericht) menge {
	m := neuesMd(text)
	var von []int
	for i := range m.zeilen {
		if e, t, ok := m.ueberschrift(i); ok && e == 2 && t == ueberschriftK {
			von = append(von, i)
		}
	}
	if len(von) != 1 {
		art := "fehlt"
		if len(von) > 1 {
			art = "mehrdeutig"
		}
		b.fehler("§4: Überschrift " + art)
		return nil
	}
	bis := len(m.zeilen)
	for j := von[0] + 1; j < len(m.zeilen); j++ {
		if !m.zaun[j] && strings.HasPrefix(m.zeilen[j], "## ") {
			bis = j
			break
		}
	}
	zeilen := m.tabelle(spanne{von[0] + 1, bis}, tabellenForm{quelle: "§4", kopf: []string{"ID", "Code", "Bedingung", "Aktion"}, daten: true}, b)
	k := neueNamen("§4")
	for _, z := range zeilen {
		x, ok := inlineCode(z.zellen[1])
		if !ok {
			x = z.zellen[1]
		}
		if !ok || !istCode(x) {
			b.befund(artForm, x, z.n, formText("§4", z.n, x))
			continue
		}
		k.add(b, x, z.n)
	}
	return k.erste
}

// liesAusnahmen liest die Ausnahmen: je Zeile `<Code> | <Grund>`, getrennt am ersten `|` (SPEC-201 Nr. 3 c).
func liesAusnahmen(text string, b *bericht) menge {
	a := neueNamen("Ausnahmen")
	for i, roh := range zerlege(text) {
		z := strings.TrimSpace(roh)
		n := i + 1
		if z == "" {
			continue
		}
		code, grund, da := strings.Cut(z, "|")
		if !da {
			b.fehler("Ausnahmen Zeile " + strconv.Itoa(n) + ": Zeile ohne Trennstrich")
			continue
		}
		code, grund = strings.TrimSpace(code), strings.TrimSpace(grund)
		if istCode(code) {
			a.add(b, code, n)
		} else {
			b.befund(artForm, code, n, formText("Ausnahmen", n, code))
		}
		if grund == "" {
			b.befund(artAusnahmeOhneGrund, code, n, "Ausnahme "+code+" (Zeile "+strconv.Itoa(n)+") ohne Grund")
		}
	}
	return a.erste
}

// handbuch trägt, was das Handbuch zu den Fehlercodes sagt: die Namen der Tabelle (T) und jede Nennung (N).
type handbuch struct {
	tabelle menge
	nennung menge
}

// liesHandbuch liest die Tabelle des Abschnitts `Fehlercodes` und jede Nennung eines Codes (SPEC-201 Nr. 3 d, 6).
func liesHandbuch(text string, b *bericht) handbuch {
	switch {
	case !utf8.ValidString(text):
		b.fehler("Handbuch nicht lesbar: kein UTF-8")
		return handbuch{}
	case strings.ContainsRune(text, 0):
		b.fehler("Handbuch nicht lesbar: NUL")
		return handbuch{}
	}
	m := neuesMd(text)
	if m.offen > 0 {
		b.fehler("Handbuch: Zaun nicht geschlossen (Zeile " + strconv.Itoa(m.offen) + ")")
		return handbuch{}
	}
	abschnitte := m.abschnitte(abschnittCodes)
	if len(abschnitte) != 1 {
		art := "fehlt"
		if len(abschnitte) > 1 {
			art = "mehrdeutig"
		}
		b.fehler("Handbuch: Abschnitt " + abschnittCodes + " " + art)
		return handbuch{}
	}
	form := tabellenForm{quelle: "Handbuch", ort: " in " + abschnittCodes, kopf: []string{"Code", "Bedeutung"}}
	zeilen := m.tabelle(abschnitte[0], form, b)
	h := handbuch{tabelle: menge{}, nennung: menge{}}
	gemeldet := map[string]bool{}
	meldeForm := func(n int, x string) {
		if k := strconv.Itoa(n) + " " + x; !gemeldet[k] {
			gemeldet[k] = true
			b.befund(artForm, x, n, formText("Handbuch", n, x))
		}
	}
	tab := neueNamen("Handbuch")
	for _, z := range zeilen {
		x, ok := inlineCode(z.zellen[0])
		switch {
		case !ok:
			b.befund(artZelle, "", z.n, "Zelle: Handbuch Zeile "+strconv.Itoa(z.n)+": erste Zelle ist nicht genau ein Name in Inline-Code")
		case !istCode(x):
			meldeForm(z.n, x)
		default:
			tab.add(b, x, z.n)
		}
	}
	h.tabelle = tab.erste
	for i, l := range m.zeilen {
		nenne(l, i+1, h.nennung, meldeForm)
	}
	return h
}

// nenne liest die Nennungen eines Codes in einer Zeile: `ODR-E` mit der ganzen Folge ASCII-Ziffern dahinter;
// vier Ziffern sind ein Code, jede andere Länge eine Form (SPEC-201 Nr. 3 d).
func nenne(zeile string, n int, genannt menge, meldeForm func(int, string)) {
	rest := zeile
	for {
		j := strings.Index(rest, "ODR-E")
		if j < 0 {
			return
		}
		rest = rest[j+5:]
		z := 0
		for z < len(rest) && rest[z] >= '0' && rest[z] <= '9' {
			z++
		}
		if z == 0 {
			continue
		}
		name := "ODR-E" + rest[:z]
		rest = rest[z:]
		switch {
		case !istCode(name):
			meldeForm(n, name)
		default:
			if _, da := genannt[name]; !da {
				genannt[name] = n
			}
		}
	}
}
