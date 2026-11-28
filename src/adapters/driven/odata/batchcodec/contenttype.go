package batchcodec

import (
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Die Medientypen, die der Leser unterscheidet, in Kleinbuchstaben.
const (
	typMultipart = "multipart/mixed"
	typHTTP      = "application/http"
	typJSON      = "application/json"
)

// grenzzeichenLaenge ist die größte Länge einer Grenzzeichenfolge (SPEC-183 Nr. 1).
const grenzzeichenLaenge = 70

// huellenGrenze prüft die Kopfzeilen Content-Type der Hülle (SPEC-183 Nr. 1) und liefert ihre Grenzzeichenfolge.
func huellenGrenze(zeilen []string) (string, error) {
	if len(zeilen) != 1 {
		return "", fehler(model.BatchContentType, "")
	}
	typ, parameter := medientyp(zeilen[0])
	switch {
	case gleichASCII(typ, typJSON):
		return "", fehler(model.BatchFormat, "")
	case !gleichASCII(typ, typMultipart):
		return "", fehler(model.BatchContentType, "")
	}
	g, ok := grenzzeichen(parameter)
	if !ok {
		return "", fehler(model.BatchBoundary, "")
	}
	return g, nil
}

// medientyp teilt einen Content-Type am ersten `;`: der Medientyp ohne SP und HTAB am Rand und die Parameter dahinter.
func medientyp(wert string) (typ, parameter string) {
	typ, parameter, _ = strings.Cut(wert, ";")
	return getrimmt(typ), parameter
}

// grenzzeichen liest aus den Parametern den einen Parameter boundary, ohne oder in einem Paar Anführungszeichen (SPEC-183 Nr. 1).
func grenzzeichen(parameter string) (string, bool) {
	var wert string
	anzahl, mitGleich := 0, false
	for rest, mehr := parameter, true; mehr; {
		var teil string
		teil, rest, mehr = strings.Cut(rest, ";")
		name, w, hat := strings.Cut(teil, "=")
		if gleichASCII(getrimmt(name), "boundary") {
			anzahl++
			wert, mitGleich = getrimmt(w), hat
		}
	}
	if anzahl != 1 || !mitGleich {
		return "", false
	}
	if len(wert) >= 2 && wert[0] == '"' && wert[len(wert)-1] == '"' {
		wert = wert[1 : len(wert)-1]
	}
	return wert, grenzzeichenGueltig(wert)
}

// grenzzeichenGueltig ist wahr für 1 bis 70 Zeichen aus `A-Z a-z 0-9 ' ( ) + _ , - . / : = ?` und SP, das letzte kein SP.
func grenzzeichenGueltig(wert string) bool {
	if wert == "" || len(wert) > grenzzeichenLaenge || wert[len(wert)-1] == ' ' {
		return false
	}
	for i := range len(wert) {
		c := wert[i]
		if !istBuchstabeOderZiffer(c) && !strings.ContainsRune("'()+_,-./:=? ", rune(c)) {
			return false
		}
	}
	return true
}

func istBuchstabeOderZiffer(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

// getrimmt entfernt SP und HTAB am Rand.
func getrimmt(s string) string { return strings.Trim(s, " \t") }

// gleichASCII vergleicht ohne Beachtung der Schreibweise allein der Buchstaben A bis Z; soll steht in Kleinbuchstaben.
func gleichASCII(s, soll string) bool {
	if len(s) != len(soll) {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != soll[i] {
			return false
		}
	}
	return true
}
