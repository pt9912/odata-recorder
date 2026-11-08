package model

import "strconv"

// EtagFuer ist der ETag zu einer Version der Entity (SPEC-280 Nr. 1): `W/"` + die Dezimalzahl ohne führende Null + `"`, schwach,
// allein aus `entity_version`. Die Funktion prüft nicht, dass version mindestens 0 ist (SPEC-108).
func EtagFuer(version int64) string {
	return `W/"` + strconv.FormatInt(version, 10) + `"`
}
