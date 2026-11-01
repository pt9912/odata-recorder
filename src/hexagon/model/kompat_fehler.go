package model

import (
	"errors"
	"strconv"
)

// KompatUrsache ist die Ursache eines KompatFehler (SPEC-174 Nr. 2).
type KompatUrsache string

// Die sechs Ursachen: eine V2-Eigenschaft außerhalb des Umfangs (SPEC-170 Nr. 2 und 3).
const (
	KompatFormat     KompatUrsache = "format"
	KompatMethode    KompatUrsache = "methode"
	KompatPfad       KompatUrsache = "pfad"
	KompatOption     KompatUrsache = "option"
	KompatFilter     KompatUrsache = "filter"
	KompatLiteralTyp KompatUrsache = "literal_typ"
)

// AlleKompatUrsachen liefert die sechs Ursachen in der Reihenfolge der Spezifikation.
func AlleKompatUrsachen() []KompatUrsache {
	return []KompatUrsache{KompatFormat, KompatMethode, KompatPfad, KompatOption, KompatFilter, KompatLiteralTyp}
}

// Gueltig meldet, ob u eine der sechs Ursachen ist.
func (u KompatUrsache) Gueltig() bool {
	for _, v := range AlleKompatUrsachen() {
		if u == v {
			return true
		}
	}
	return false
}

// ErrKompatibilitaet erkennt jeden KompatFehler; die Abbildung macht daraus ODR-E6006.
var ErrKompatibilitaet = errors.New("model: OData-V2-Kompatibilität")

// KompatFehler ist der Fehler einer V2-Eigenschaft außerhalb des Umfangs (SPEC-174 Nr. 1): Ursache und Ort, nie ein Teil der
// Eingabe. Er ist kein URLFehler.
type KompatFehler struct {
	Ursache KompatUrsache
	Ort     string
}

// Error nennt den Code und die Ursache; der Ort steht in der Schreibweise von strconv.Quote dahinter.
func (e *KompatFehler) Error() string {
	s := "ODR-E6006: OData-V2-Kompatibilität: " + string(e.Ursache)
	if e.Ort != "" {
		s += " " + strconv.Quote(e.Ort)
	}
	return s
}

// Code liefert ODR-E6006.
func (e *KompatFehler) Code() Code { return CodeE6006 }

// Unwrap macht ErrKompatibilitaet erreichbar, keinen Sentinel der URL- oder Eingabefehler.
func (e *KompatFehler) Unwrap() error { return ErrKompatibilitaet }
