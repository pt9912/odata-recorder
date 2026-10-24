package model

import (
	"errors"
	"strconv"
)

// DatenUrsache ist die Ursache eines DatenFehler (SPEC-126 Nr. 14).
type DatenUrsache string

// Die fünfzehn Ursachen: fünf des Lesens, neun des Deutens und entity_set.
const (
	DatenSyntax             DatenUrsache = "syntax"
	DatenKodierung          DatenUrsache = "kodierung"
	DatenTiefe              DatenUrsache = "tiefe"
	DatenZuGross            DatenUrsache = "zu_gross"
	DatenDoppelterName      DatenUrsache = "doppelter_name"
	DatenStruktur           DatenUrsache = "struktur"
	DatenKontext            DatenUrsache = "kontext"
	DatenODataTyp           DatenUrsache = "odata_typ"
	DatenUnbekannteProperty DatenUrsache = "unbekannte_property"
	DatenTyp                DatenUrsache = "typ"
	DatenWert               DatenUrsache = "wert"
	DatenNull               DatenUrsache = "null"
	DatenZahlZuLang         DatenUrsache = "zahl_zu_lang"
	DatenSchluessel         DatenUrsache = "schluessel"
	DatenEntitySet          DatenUrsache = "entity_set"
)

// ErrDatenUngueltig erkennt jeden DatenFehler über errors.Is; die Abbildung macht daraus ODR-E6101.
var ErrDatenUngueltig = errors.New("model: Daten ungültig")

// DatenFehler ist der Fehler des Lesers der JSON-Antworten (SPEC-126 Nr. 14): ODR-E6101 mit Ursache und Ort.
type DatenFehler struct {
	Ursache DatenUrsache
	Ort     string
}

// Error nennt den Code und die Ursache; der Ort steht in der Schreibweise von strconv.Quote dahinter, denn er
// enthält Text des Dokuments und darf kein rohes Steuerzeichen tragen.
func (e *DatenFehler) Error() string {
	s := "ODR-E6101: Baseline-Import ungültig oder inkonsistent: " + string(e.Ursache)
	if e.Ort != "" {
		s += " " + strconv.Quote(e.Ort)
	}
	return s
}

// Code liefert ODR-E6101.
func (*DatenFehler) Code() Code { return CodeE6101 }

// Unwrap macht ErrDatenUngueltig erreichbar.
func (*DatenFehler) Unwrap() error { return ErrDatenUngueltig }
