package main

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const (
	nichtLesbar  = "Messung nicht lesbar"
	fremdePlan   = "Messung gehört nicht zu diesem Plan"
	umbenennung  = "Umbenennung in der Messung"
	spitzeMin    = 40
	spitzeMax    = 64
	zeileNumstat = "numstat"
)

// messung ist das Ergebnis des Lesens der Messung (SPEC-128 Nr. 4 c bis f): die Summe der Zeilen
// und die Zahl der Binärdateien, die nicht zählen.
type messung struct {
	summe   int64
	binaer  int
	fehler  string
	hinweis string
}

// naechsteZeile liefert die Zeile bis zum Zeilenumbruch und den Rest; ok ist falsch, wenn kein
// Zeilenumbruch mehr folgt.
func naechsteZeile(s []byte) (string, []byte, bool) {
	i := bytes.IndexByte(s, '\n')
	if i < 0 {
		return "", s, false
	}
	return string(s[:i]), s[i+1:], true
}

// leseMessung prüft den Kopf gegen den Block und summiert die Datensätze.
func leseMessung(data []byte, b block) messung {
	basis, pfade, rest, grund := leseKopf(data)
	if grund != "" {
		return messung{fehler: nichtLesbar + ": " + grund}
	}
	if basis != b.basis || !slices.Equal(pfade, b.pfade) {
		return messung{fehler: fremdePlan}
	}
	return summiere(rest)
}

// leseKopf liest die Zeilen `basis`, `spitze`, `pfad`… und `numstat`; grund ist gesetzt, wenn
// der Kopf fehlt oder unvollständig ist.
func leseKopf(data []byte) (string, []string, []byte, string) {
	basis, rest, ok := kopfFeld(data, "basis")
	if !ok {
		return "", nil, nil, "Zeile basis fehlt"
	}
	spitze, rest, ok := kopfFeld(rest, "spitze")
	if !ok {
		return "", nil, nil, "Zeile spitze fehlt"
	}
	if len(spitze) < spitzeMin || len(spitze) > spitzeMax || !nurHex(spitze) {
		return "", nil, nil, "spitze ungültig"
	}
	pfade, rest, grund := lesePfade(rest)
	return basis, pfade, rest, grund
}

// kopfFeld liest die Zeile `<name> <wert>` und liefert den Wert.
func kopfFeld(s []byte, name string) (string, []byte, bool) {
	z, rest, ok := naechsteZeile(s)
	wert, hat := strings.CutPrefix(z, name+" ")
	return wert, rest, ok && hat
}

// lesePfade liest die Zeilen `pfad <Pfad>` bis zur Zeile `numstat`.
func lesePfade(rest []byte) ([]string, []byte, string) {
	var pfade []string
	for {
		z, r, ok := naechsteZeile(rest)
		rest = r
		switch {
		case !ok:
			return nil, nil, "Zeile " + zeileNumstat + " fehlt"
		case z == zeileNumstat:
			return pfade, rest, ""
		case strings.HasPrefix(z, "pfad "):
			pfade = append(pfade, z[5:])
		default:
			return nil, nil, "unbekannte Kopfzeile"
		}
	}
}

// summiere liest die Datensätze `ins TAB del TAB Pfad NUL` bis zum Ende.
func summiere(rest []byte) messung {
	var m messung
	for len(rest) > 0 {
		i := bytes.IndexByte(rest, 0)
		if i < 0 {
			return messung{fehler: nichtLesbar + ": Datensatz ohne abschließendes NUL"}
		}
		n, binaer, fehler := leseDatensatz(rest[:i])
		if fehler != "" {
			return messung{fehler: fehler}
		}
		rest = rest[i+1:]
		if binaer {
			m.binaer++
		}
		m.summe += n
	}
	if m.binaer > 0 {
		m.hinweis = fmt.Sprintf("%d Binärdatei(en) nicht gezählt", m.binaer)
	}
	return m
}

// leseDatensatz liest einen Datensatz ohne das abschließende NUL. Der Pfad ist der Rest hinter
// dem zweiten Tab und wird nicht weiter gelesen.
func leseDatensatz(rec []byte) (int64, bool, string) {
	fehler := nichtLesbar + ": Datensatz"
	t1 := bytes.IndexByte(rec, '\t')
	if t1 < 0 {
		return 0, false, fehler
	}
	t2 := bytes.IndexByte(rec[t1+1:], '\t')
	if t2 < 0 {
		return 0, false, fehler
	}
	ins, del := string(rec[:t1]), string(rec[t1+1:t1+1+t2])
	pfad := rec[t1+t2+2:]
	n, binaer, ok := zeilen(ins, del)
	switch {
	case !ok:
		return 0, false, fehler
	case len(pfad) == 0:
		return 0, false, umbenennung
	}
	return n, binaer, ""
}

// zeilen liest die beiden Zahlen eines Datensatzes: beide Zahlen nach Nr. 3 j, oder beide `-`
// für eine Binärdatei.
func zeilen(ins, del string) (int64, bool, bool) {
	if ins == "-" && del == "-" {
		return 0, true, true
	}
	a, ok1 := zahl(ins)
	b, ok2 := zahl(del)
	return a + b, false, ok1 && ok2
}
