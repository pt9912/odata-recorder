package jsoncodec

import (
	"strconv"
	"strings"
)

// pruefeEtag prüft `Entity.Etag` (SPEC-168 Nr. 3 e): leer heißt kein ETag, sonst `W/"`, die Ziffern einer Ganzzahl von 0 bis
// 9223372036854775807 ohne führende Null (`0` allein) und `"`; jede andere Zeichenfolge ist ErrUngueltig.
func pruefeEtag(etag string) error {
	if etag == "" {
		return nil
	}
	ziffern, ok := strings.CutPrefix(etag, `W/"`)
	if ok {
		ziffern, ok = strings.CutSuffix(ziffern, `"`)
	}
	if !ok || ziffern == "" || ziffern[0] == '0' && len(ziffern) > 1 {
		return ungueltig("Etag hat nicht die Form W/\"<Ganzzahl>\"")
	}
	for i := range len(ziffern) {
		if ziffern[i] < '0' || ziffern[i] > '9' {
			return ungueltig("Etag hat nicht die Form W/\"<Ganzzahl>\"")
		}
	}
	if _, err := strconv.ParseInt(ziffern, 10, 64); err != nil {
		return ungueltig("Etag liegt außerhalb von 0 bis 9223372036854775807")
	}
	return nil
}
