package model

// EingabeArt ist die Methode eines Requests, dessen Body ODataEingabe liest (SPEC-180).
type EingabeArt string

// Die drei Methoden; jede andere Zeichenfolge, auch in anderer Schreibweise, ist keine EingabeArt.
const (
	EingabePost  EingabeArt = "POST"
	EingabePatch EingabeArt = "PATCH"
	EingabePut   EingabeArt = "PUT"
)

// Gueltig ist wahr genau für die drei Konstanten.
func (a EingabeArt) Gueltig() bool {
	return a == EingabePost || a == EingabePatch || a == EingabePut
}

// EntityEingabe ist der gelesene Body eines POST, PATCH oder PUT (SPEC-180 Nr. 2). Properties hat einen Eintrag je Property, die
// der Body nennt, in der Reihenfolge des Dokuments; eine Property ohne Eintrag war nicht genannt. Schluessel ist die Literalform
// des Schlüssels, wenn der Body alle Schlüssel-Properties nennt, sonst leer.
type EntityEingabe struct {
	Typ        string
	Schluessel string
	Properties []PropertyWert
}
