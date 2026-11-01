package jsoncodecv2

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Die Millisekunden seit 1970-01-01T00:00:00Z, zwischen denen ein Zeitpunkt liegen darf: 0001-01-01T00:00:00.000Z bis
// 9999-12-31T23:59:59.999Z (SPEC-172 Nr. 4).
const (
	zeitMin = -62135596800000
	zeitMax = 253402300799999
)

// zeitpunktText ist die kanonische Schreibweise des Zeitpunkts in UTC zu einem Text der Form `/Date(<ms>)/` (SPEC-172 Nr. 4):
// `JJJJ-MM-TTThh:mm:ss`, bei Millisekunden ungleich null mit dem Bruchteil ohne Nullen am Ende, und `Z`. Die Zahl ist dezimal,
// `-` steht bei negativem Wert, eine führende Null außer `0` allein, `-0`, `+` und jeder Zusatz sind keine Zahl.
func zeitpunktText(text string) (string, bool) {
	rest, ok := strings.CutPrefix(text, "/Date(")
	if !ok {
		return "", false
	}
	ganz, ok := strings.CutSuffix(rest, ")/")
	if !ok {
		return "", false
	}
	negativ := strings.HasPrefix(ganz, "-")
	ziffern := strings.TrimPrefix(ganz, "-")
	if ziffern == "" || (ziffern[0] == '0' && (len(ziffern) > 1 || negativ)) {
		return "", false
	}
	for i := range len(ziffern) {
		if !ziffer(ziffern[i]) {
			return "", false
		}
	}
	ms, err := strconv.ParseInt(ganz, 10, 64)
	if err != nil || ms < zeitMin || ms > zeitMax {
		return "", false
	}
	out := time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05")
	if bruch := ((ms % 1000) + 1000) % 1000; bruch != 0 {
		out += "." + strings.TrimRight(fmt.Sprintf("%03d", bruch), "0")
	}
	return out + "Z", true
}
