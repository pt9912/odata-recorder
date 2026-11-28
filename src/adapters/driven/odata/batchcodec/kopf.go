package batchcodec

import (
	"bytes"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// maxZeilen ist die Zahl der Kopfzeilen eines Blocks; die leere Zeile zählt nicht (SPEC-183 Nr. 7).
const maxZeilen = 100

// block liest die Kopfzeilen eines Blocks bis zur ersten leeren Zeile oder zum Ende (SPEC-183 Nr. 3) und ruft je für jede. Je Zeile
// gehen die Zahl, die Form und das falsche Ergebnis von je in dieser Folge ein (SPEC-182 Nr. 5 g). rest sind die Bytes hinter der
// leeren Zeile, ohne sie ist er leer; ursache ist ohne Verstoß leer.
func block(inhalt []byte, je func(name, wert []byte) bool) (rest []byte, ursache model.BatchUrsache) {
	for n := 1; len(inhalt) > 0; n++ {
		zeile, danach, _ := bytes.Cut(inhalt, []byte(zeilenende))
		if len(zeile) == 0 {
			return danach, ""
		}
		if n > maxZeilen {
			return nil, model.BatchZuViele
		}
		name, wert, ok := kopfzeile(zeile)
		if !ok || !je(name, wert) {
			return nil, model.BatchKopfzeile
		}
		inhalt = danach
	}
	return nil, ""
}

// kopfzeile zerlegt `Name ":" Wert`: der Name ein Token, der Wert ohne SP und HTAB an den Rändern aus HTAB, 0x20 bis 0x7E und 0x80
// bis 0xFF.
func kopfzeile(zeile []byte) (name, wert []byte, ok bool) {
	name, wert, ok = bytes.Cut(zeile, []byte(":"))
	if !ok || !istToken(name) {
		return nil, nil, false
	}
	wert = bytes.Trim(wert, " \t")
	for _, c := range wert {
		if c != '\t' && (c < 0x20 || c == 0x7F) {
			return nil, nil, false
		}
	}
	return name, wert, true
}

// istToken ist wahr für ein nichtleeres Token aus tchar (RFC 9110 §5.6.2).
func istToken(b []byte) bool {
	for _, c := range b {
		if !istBuchstabeOderZiffer(c) && !strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)) {
			return false
		}
	}
	return len(b) > 0
}
