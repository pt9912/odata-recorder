package model

import (
	"slices"
	"strings"
)

// VorbedingungErfuellt wertet `If-Match` und `If-None-Match` gegen die Version des Eintrags aus (SPEC-283): ifMatch und
// ifNoneMatch sind die Zeilen der Kopfzeile, ohne Zeile (nil oder leer) ist der Header nicht gesetzt und erfüllt. Ein gesetzter
// Header ist erfüllt, wenn sein Wert lesbar ist (Nr. 3) und die Bedingung gilt (Nr. 5): `If-Match` mit `*` oder einem Tag, das dem
// Tag von EtagFuer(version) gleicht, `If-None-Match` ohne `*` und ohne ein solches Tag. Der Vergleich ist der schwache auf den
// Bytes des Tags (Nr. 4). Beide gesetzte Header müssen erfüllt sein. Die Funktion setzt einen vorhandenen Eintrag voraus.
func VorbedingungErfuellt(ifMatch, ifNoneMatch []string, version int64) bool {
	etag := EtagFuer(version)
	tag := etag[strings.IndexByte(etag, '"'):]
	if len(ifMatch) > 0 {
		liste, ok := liesEtagListe(ifMatch)
		if !ok || !liste.stern && !liste.enthaelt(tag) {
			return false
		}
	}
	if len(ifNoneMatch) > 0 {
		liste, ok := liesEtagListe(ifNoneMatch)
		if !ok || liste.stern || liste.enthaelt(tag) {
			return false
		}
	}
	return true
}

// etagListe ist ein lesbarer Wert: entweder das eine Element `*` oder die Tags (von `"` bis `"`, ohne `W/`).
type etagListe struct {
	stern bool
	tags  []string
}

func (l etagListe) enthaelt(tag string) bool { return slices.Contains(l.tags, tag) }

// liesEtagListe liest die Zeilen als eine Liste nach SPEC-283 Nr. 3 in einem Durchgang; ok ist falsch bei einem unlesbaren Wert.
func liesEtagListe(zeilen []string) (etagListe, bool) {
	s := strings.Join(zeilen, ",")
	var l etagListe
	elemente := 0
	for i := ueberLeerraum(s, 0); i < len(s); i = ueberLeerraum(s, i) {
		if s[i] == ',' {
			i++
			continue
		}
		elemente++
		if s[i] == '*' {
			l.stern = true
			i++
		} else {
			tag, naechste, ok := liesTag(s, i)
			if !ok {
				return etagListe{}, false
			}
			l.tags, i = append(l.tags, tag), naechste
		}
		if i = ueberLeerraum(s, i); i < len(s) && s[i] != ',' {
			return etagListe{}, false
		}
	}
	if elemente == 0 || l.stern && elemente > 1 {
		return etagListe{}, false
	}
	return l, true
}

// liesTag liest ab i ein ETag (optional `W/`, dann `"`, Zeichen des Tags, `"`) und liefert den Teil von `"` bis `"` und die Stelle dahinter;
// ok ist falsch, wenn dort kein ETag steht.
func liesTag(s string, i int) (tag string, naechste int, ok bool) {
	if s[i] == 'W' {
		if i+1 >= len(s) || s[i+1] != '/' {
			return "", i, false
		}
		i += 2
	}
	if i >= len(s) || s[i] != '"' {
		return "", i, false
	}
	start := i
	for i++; i < len(s) && s[i] != '"'; i++ {
		if !etagZeichen(s[i]) {
			return "", i, false
		}
	}
	if i >= len(s) {
		return "", i, false
	}
	return s[start : i+1], i + 1, true
}

// ueberLeerraum liefert die Stelle nach Leerzeichen und Tabulator ab i (nichts sonst ist Leerraum).
func ueberLeerraum(s string, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return i
}

// etagZeichen sagt, ob b in einem Tag stehen darf: 0x21, 0x23 bis 0x7E und 0x80 bis 0xFF.
func etagZeichen(b byte) bool {
	return b == 0x21 || b >= 0x23 && b <= 0x7E || b >= 0x80
}
