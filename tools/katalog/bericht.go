package main

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strconv"
)

// Die Befundarten (SPEC-201 Nr. 7); die Zahl ist die Stelle in der Reihenfolge der Ausgabe.
const (
	artForm = iota + 1
	artZelle
	artDoppelt
	artOhneSpezifikation
	artAusnahmeOhneGrund
	artAusnahmeOhneCode
	artFehlt
	artUnbekannt
	artAusnahmeGenannt
	artAbgelehnt
)

// Die Kataloge in der Reihenfolge der Ausgabe (SPEC-201 Nr. 2 d).
const (
	katFehlercodes = iota
	katSchluessel
	katKommandos
)

func katalogName(katalog int) string {
	return [...]string{"Fehlercodes", "Schlüssel", "Kommandos"}[katalog]
}

// befund ist eine Zeile der Ausgabe mit den Schlüsseln ihrer Reihenfolge (SPEC-201 Nr. 2 d).
type befund struct {
	katalog int
	art     int
	name    string
	zeile   int
	text    string
}

// bericht sammelt die Eingabefehler und die Befunde eines Laufs.
type bericht struct {
	eingabe []string
	befunde []befund
}

func (b *bericht) fehler(text string) { b.eingabe = append(b.eingabe, text) }

func (b *bericht) befund(art int, name string, zeile int, text string) {
	b.befundIn(katFehlercodes, art, name, zeile, text)
}

func (b *bericht) befundIn(katalog, art int, name string, zeile int, text string) {
	b.befunde = append(b.befunde, befund{katalog, art, name, zeile, text})
}

// ausgeben schreibt den Bericht und liefert den Ausgangswert: 2 bei einem Eingabefehler (er schließt
// jeden Befund aus), 1 bei einem Befund, sonst 0 (SPEC-201 Nr. 2 b).
func (b *bericht) ausgeben(stdout, stderr io.Writer) int {
	if len(b.eingabe) > 0 {
		for _, t := range b.eingabe {
			_, _ = fmt.Fprintf(stderr, "katalog: Eingabe: %s\n", t)
		}
		return 2
	}
	if len(b.befunde) == 0 {
		_, _ = fmt.Fprintln(stdout, "katalog: ok")
		return 0
	}
	slices.SortFunc(b.befunde, func(x, y befund) int {
		return cmp.Or(cmp.Compare(x.katalog, y.katalog), cmp.Compare(x.art, y.art), cmp.Compare(x.name, y.name), cmp.Compare(x.zeile, y.zeile), cmp.Compare(x.text, y.text))
	})
	for _, f := range b.befunde {
		_, _ = fmt.Fprintf(stderr, "katalog: %s: %s\n", katalogName(f.katalog), f.text)
	}
	return 1
}

// formText ist der Text der Art 1: der Wert in Go-Schreibweise, auf 200 Zeichen gekürzt.
func formText(quelle string, zeile int, x string) string {
	q := []rune(strconv.Quote(x))
	if len(q) > 200 {
		q = q[:200]
	}
	return fmt.Sprintf("Form: %s Zeile %d: %s hat nicht die Form", quelle, zeile, string(q))
}

// menge ordnet einem Namen die Zeile seines ersten Vorkommens zu.
type menge map[string]int

// namen sammelt die Namen einer Quelle; ein Name mit mehr als einem Vorkommen ist ein Befund, einmal je
// Name (Art 3).
type namen struct {
	katalog  int
	quelle   string
	erste    menge
	gemeldet map[string]bool
}

func neueNamen(quelle string) *namen { return neueNamenIn(katFehlercodes, quelle) }

func neueNamenIn(katalog int, quelle string) *namen {
	return &namen{katalog: katalog, quelle: quelle, erste: menge{}, gemeldet: map[string]bool{}}
}

func (n *namen) add(b *bericht, name string, zeile int) {
	erste, da := n.erste[name]
	switch {
	case !da:
		n.erste[name] = zeile
	case !n.gemeldet[name]:
		n.gemeldet[name] = true
		b.befundIn(n.katalog, artDoppelt, name, erste, fmt.Sprintf("%s doppelt in %s, erstes Vorkommen Zeile %d", name, n.quelle, erste))
	}
}

// istCode sagt, ob v ganz `^ODR-E[0-9]{4}$` trifft (ASCII-Ziffern, SPEC-201 Nr. 3 a).
func istCode(v string) bool {
	if len(v) != 9 || v[:5] != "ODR-E" {
		return false
	}
	for i := 5; i < 9; i++ {
		if v[i] < '0' || v[i] > '9' {
			return false
		}
	}
	return true
}
