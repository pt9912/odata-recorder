package main

import (
	"fmt"
	"strings"
	"unicode"
)

// kopf ist das Ergebnis der Prüfung des Kopfes (SPEC-129 Nr. 4 a, b, d); version ist der Wert von
// Software-Version, vergleich wahr, wenn der Schlüssel genau einmal und nicht leer steht.
type kopf struct {
	befunde   []befund
	version   string
	vergleich bool
	zeile     int
}

func kopfSchluessel() []string {
	return []string{"Software-Version", "Gültigkeitsbereich", "Stand"}
}

// kopfEnde ist der Index der ersten Zeile `## ` außerhalb von Zäunen, sonst die Zahl der Zeilen.
func kopfEnde(d dokument) int {
	for i, z := range d.zeilen {
		if d.im[i] == 0 && strings.HasPrefix(z, "## ") {
			return i
		}
	}
	return len(d.zeilen)
}

// leseKopf liest die Zeilen vor der ersten Zeile `## ` außerhalb von Zäunen.
func leseKopf(d dokument) kopf {
	var k kopf
	werte := map[string][]string{}
	zeilen := map[string]int{}
	k.befunde = leseKopfZeilen(d.zeilen[:kopfEnde(d)], werte, zeilen)
	for _, s := range kopfSchluessel() {
		k.befunde = append(k.befunde, schluesselBefunde(s, werte[s], zeilen[s])...)
	}
	if w := werte["Software-Version"]; len(w) == 1 && w[0] != "" {
		k.version, k.vergleich, k.zeile = w[0], true, zeilen["Software-Version"]
	}
	return k
}

// leseKopfZeilen sammelt die Werte und die ersten Zeilen der bekannten Schlüssel und meldet die
// Überschrift und die Zeilen, die kein Schlüssel sind (SPEC-200 Nr. 4 a).
func leseKopfZeilen(kopfZeilen []string, werte map[string][]string, zeilen map[string]int) []befund {
	var bf []befund
	ueberschrift := false
	for i, z := range kopfZeilen {
		if strings.TrimSpace(z) == "" {
			continue
		}
		if !ueberschrift {
			ueberschrift = true
			if !strings.HasPrefix(z, "# ") {
				bf = append(bf, kopfBefund(i+1, fmt.Sprintf("Zeile %d: Kopf: Überschrift fehlt", i+1)))
			}
			continue
		}
		schluessel, wert, _ := strings.Cut(z, ":")
		schluessel = strings.TrimSpace(schluessel)
		switch {
		case !strings.Contains(z, ":") || schluessel == "":
			bf = append(bf, kopfBefund(i+1, fmt.Sprintf("Zeile %d: Kopf: Zeile ohne Schlüssel", i+1)))
		case !enthaelt(kopfSchluessel(), schluessel):
			bf = append(bf, kopfBefund(i+1, fmt.Sprintf("Zeile %d: Kopf: Schlüssel %s unbekannt", i+1, schluessel)))
		default:
			if len(werte[schluessel]) == 0 {
				zeilen[schluessel] = i + 1
			}
			werte[schluessel] = append(werte[schluessel], strings.TrimSpace(wert))
		}
	}
	if !ueberschrift {
		bf = append(bf, kopfBefund(1, "Zeile 1: Kopf: Überschrift fehlt"))
	}
	return bf
}

func kopfBefund(zeile int, text string) befund {
	return befund{nr: nrKopf, zeile: zeile, text: text}
}

func enthaelt(liste []string, s string) bool {
	for _, e := range liste {
		if e == s {
			return true
		}
	}
	return false
}

// schluesselBefunde meldet je Schlüssel höchstens einen Befund in der Rangfolge fehlt, doppelt,
// leer (SPEC-129 Nr. 4 b); der Wert von Stand wird auf die Form geprüft (Nr. 4 d).
func schluesselBefunde(s string, werte []string, zeile int) []befund {
	switch {
	case len(werte) == 0:
		return []befund{kopfBefund(0, fmt.Sprintf("Kopf: %s fehlt", s))}
	case len(werte) > 1:
		return []befund{kopfBefund(zeile, fmt.Sprintf("Kopf: %s doppelt", s))}
	case werte[0] == "":
		return []befund{kopfBefund(zeile, fmt.Sprintf("Kopf: %s leer", s))}
	case s == "Stand" && !standGueltig(werte[0]):
		return []befund{kopfBefund(zeile, fmt.Sprintf("Kopf: Stand ungültig: %s", werte[0]))}
	}
	return nil
}

// standGueltig ist wahr für JJJJ-MM-TT mit einem Tag des Gregorianischen Kalenders.
func standGueltig(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	j, m, t := ziffern(s[0:4]), ziffern(s[5:7]), ziffern(s[8:10])
	return j >= 0 && m >= 1 && m <= 12 && t >= 1 && t <= tageImMonat(j, m)
}

// ziffern liefert den Wert einer Folge ASCII-Ziffern, sonst -1.
func ziffern(s string) int {
	w := 0
	for _, c := range []byte(s) {
		if c < '0' || c > '9' {
			return -1
		}
		w = w*10 + int(c-'0')
	}
	return w
}

func tageImMonat(jahr, monat int) int {
	switch monat {
	case 2:
		if (jahr%4 == 0 && jahr%100 != 0) || jahr%400 == 0 {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	}
	return 31
}

// versionBefunde vergleicht die Ausgabe von `version` mit dem Kopf (SPEC-129 Nr. 4 c); eine
// überschrittene Zeitgrenze ist ein Befund der Nummer 7 c.
func versionBefunde(l ergebnis, k kopf) []befund {
	if l.zeit {
		return []befund{{nr: nrLauf, name: "version", text: "version: Zeitgrenze überschritten"}}
	}
	v, grund := versionAusgabe(l)
	if grund != "" {
		return []befund{kopfBefund(k.zeile, "version: Ausgabe nicht lesbar: "+grund)}
	}
	if k.vergleich && k.version != v {
		return []befund{kopfBefund(k.zeile, fmt.Sprintf("Software-Version %s weicht von der Ausgabe von version ab (%s)", k.version, v))}
	}
	return nil
}

// versionAusgabe liest `odata-recorder <V>` mit einem Zeilenumbruch, Exit 0 und leerer Fehlerausgabe.
func versionAusgabe(l ergebnis) (string, string) {
	switch {
	case l.signal:
		return "", "Signal"
	case l.code != 0:
		return "", fmt.Sprintf("Exit %d", l.code)
	case l.stderr != "":
		return "", "Fehlerausgabe nicht leer"
	}
	rest, ok := strings.CutPrefix(l.stdout, kommando+" ")
	v, ok2 := strings.CutSuffix(rest, "\n")
	if !ok || !ok2 || v == "" || strings.ContainsFunc(v, unicode.IsSpace) {
		return "", fmt.Sprintf("Standardausgabe %q", l.stdout)
	}
	return v, ""
}
