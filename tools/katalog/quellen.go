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

	abschnittSchluessel = "Einstellungen"
	abschnittKommandos  = "Kommandos"
	pfadKonfig          = "src/bootstrap"
	pfadCli             = "src/adapters/driving/cli/cli.go"
	ueberschriftSchl    = "### SPEC-110 — "
	zeileSchl           = "2. **Schlüssel dieser Version.**"
	hinterSchl          = "Geprüft werden"
)

// lies liest eine Datei unter der Wurzel; fehlt sie, oder ist eine Komponente des Pfads eine Datei, ist das
// `<Datei> fehlt: <Pfad>`, sonst `<Datei> nicht lesbar: <Grund>` (SPEC-201 Nr. 8).
func lies(wurzel, pfad, datei string, b *bericht) (string, bool) {
	roh, err := os.ReadFile(filepath.Join(wurzel, filepath.FromSlash(pfad)))
	if err == nil {
		return string(roh), true
	}
	meldeLesefehler(err, datei, pfad, b)
	return "", false
}

// meldeLesefehler macht aus dem Fehler beim Lesen den Eingabefehler `fehlt` oder `nicht lesbar`.
func meldeLesefehler(err error, datei, pfad string, b *bericht) {
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
		b.fehler(datei + " fehlt: " + pfad)
		return
	}
	b.fehler(datei + " nicht lesbar: " + grundVon(err))
}

// grundVon ist der Text des Systemfehlers ohne Pfad.
func grundVon(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
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

// handbuch trägt, was das Handbuch zu den Katalogen sagt: je Katalog die Namen der Tabelle und für die Fehlercodes jede
// Nennung.
type handbuch struct {
	tabelle menge // Fehlercodes: Namen der Tabelle
	nennung menge // Fehlercodes: jede Nennung im Text
	schl    menge // Schlüssel: Namen der Tabelle Einstellungen
	komm    menge // Kommandos: Namen der Tabelle Kommandos
}

// liesHandbuch liest die Tabellen der Abschnitte `Fehlercodes`, `Einstellungen` und `Kommandos` und jede Nennung eines
// Codes; die Abschnitte werden je für sich geprüft (SPEC-201 Nr. 3 d, 4 c, 5 b, 6, 8).
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
	h := handbuch{nennung: menge{}}
	gemeldet := map[string]bool{}
	meldeForm := func(n int, x string) {
		if k := strconv.Itoa(n) + " " + x; !gemeldet[k] {
			gemeldet[k] = true
			b.befund(artForm, x, n, formText("Handbuch", n, x))
		}
	}
	h.tabelle = m.katalogTabelle(katFehlercodes, abschnittCodes, "Code", istCode, meldeForm, b)
	h.schl = m.katalogTabelle(katSchluessel, abschnittSchluessel, "Schlüssel", istSchluessel, direktForm(katSchluessel, b), b)
	h.komm = m.katalogTabelle(katKommandos, abschnittKommandos, "Kommando", istKommando, direktForm(katKommandos, b), b)
	if h.tabelle != nil {
		for i, l := range m.zeilen {
			nenne(l, i+1, h.nennung, meldeForm)
		}
	}
	return h
}

// direktForm meldet einen Namen falscher Form sofort als Art 1 des Katalogs.
func direktForm(katalog int, b *bericht) func(int, string) {
	return func(n int, x string) { b.befundIn(katalog, artForm, x, n, formText("Handbuch", n, x)) }
}

// katalogTabelle liest die Tabelle des Abschnitts eines Katalogs: der Name einer Zeile ist der Inline-Code-Span der
// ersten Zelle (SPEC-201 Nr. 6 b bis d). Fehlt der Abschnitt oder ist er mehrdeutig, ist das ein Eingabefehler und das
// Ergebnis nil.
func (m md) katalogTabelle(katalog int, abschnitt, kopf string, istName func(string) bool, form func(int, string), b *bericht) menge {
	abschnitte := m.abschnitte(abschnitt)
	if len(abschnitte) != 1 {
		art := "fehlt"
		if len(abschnitte) > 1 {
			art = "mehrdeutig"
		}
		b.fehler("Handbuch: Abschnitt " + abschnitt + " " + art)
		return nil
	}
	f := tabellenForm{quelle: "Handbuch", ort: " in " + abschnitt, kopf: []string{kopf, "Bedeutung"}}
	tab := neueNamenIn(katalog, "Handbuch")
	for _, z := range m.tabelle(abschnitte[0], f, b) {
		x, ok := inlineCode(z.zellen[0])
		switch {
		case !ok:
			b.befundIn(katalog, artZelle, "", z.n, "Zelle: Handbuch Zeile "+strconv.Itoa(z.n)+": erste Zelle ist nicht genau ein Name in Inline-Code")
		case !istName(x):
			form(z.n, x)
		default:
			tab.add(b, x, z.n)
		}
	}
	return tab.erste
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
