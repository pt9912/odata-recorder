package model

// Batch ist ein gelesener `$batch`-Request: die Teile der Hülle in der Reihenfolge des Bodys (SPEC-182 Nr. 3). Teile, Requests und
// Header eines vom Leser gelieferten Batch sind nie nil.
type Batch struct {
	Teile []BatchTeil
}

// BatchTeil ist ein Teil der Hülle: ein Einzelrequest (Changeset falsch, genau ein Request) oder ein Changeset (Changeset wahr,
// seine Requests, auch keiner).
type BatchTeil struct {
	Changeset bool
	Requests  []BatchRequest
}

// BatchRequest ist ein Einzelrequest, wie er im Body steht (SPEC-183 Nr. 4 und 6): Methode und Ziel ungeprüft, ContentID leer,
// wenn der Teil keine trägt, Body nil, wenn der Request kein Body-Byte trägt.
type BatchRequest struct {
	ContentID, Methode, Ziel string
	Header                   []BatchKopf
	Body                     []byte
}

// BatchKopf ist eine Kopfzeile eines Einzelrequests: der Name wie geschrieben, der Wert ohne SP und HTAB an den Rändern.
type BatchKopf struct {
	Name, Wert string
}

// contentIDGrenze ist die größte Länge einer Content-ID in Zeichen (SPEC-183 Nr. 4 c).
const contentIDGrenze = 128

// ContentIDGueltig meldet, ob wert die Form einer Content-ID hat: 1 bis 128 Zeichen aus `A-Z a-z 0-9 - . _ ~` (SPEC-183 Nr. 4 c).
func ContentIDGueltig(wert string) bool {
	if wert == "" || len(wert) > contentIDGrenze {
		return false
	}
	for i := range len(wert) {
		if !contentIDZeichen(wert[i]) {
			return false
		}
	}
	return true
}

func contentIDZeichen(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	}
	return c == '-' || c == '.' || c == '_' || c == '~'
}
