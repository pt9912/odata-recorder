package model

// AbfrageOption ist die Option, deren Name den AbfrageFehler auslöst (SPEC-190 Nr. 4).
type AbfrageOption string

// Die drei Optionen, deren Namen der Interpreter prüft.
const (
	AbfrageFilter  AbfrageOption = "$filter"
	AbfrageOrderBy AbfrageOption = "$orderby"
	AbfrageSelect  AbfrageOption = "$select"
)

// AbfrageUrsache ist die Ursache eines AbfrageFehler (SPEC-190 Nr. 4).
type AbfrageUrsache string

// Die vier Ursachen.
const (
	AbfrageEigenschaftUnbekannt     AbfrageUrsache = "eigenschaft_unbekannt"      // Klasse `ungueltig` (ODR-E6005, 400)
	AbfrageTypNichtUnterstuetzt     AbfrageUrsache = "typ_nicht_unterstuetzt"     // Klasse `nicht_unterstuetzt` (ODR-E6001, 501)
	AbfrageOrdnungNichtUnterstuetzt AbfrageUrsache = "ordnung_nicht_unterstuetzt" // Klasse `nicht_unterstuetzt` (ODR-E6001, 501)
	AbfrageNichtBeobachtet          AbfrageUrsache = "nicht_beobachtet"           // Klasse `nicht_beobachtet` (ODR-E6003, 422)
)

// AlleAbfrageUrsachen liefert die vier Ursachen.
func AlleAbfrageUrsachen() []AbfrageUrsache {
	return []AbfrageUrsache{
		AbfrageEigenschaftUnbekannt, AbfrageTypNichtUnterstuetzt, AbfrageOrdnungNichtUnterstuetzt, AbfrageNichtBeobachtet,
	}
}

// AbfrageFehler ist der Fehler des Interpreters für einen Namen in `$orderby`, `$select` oder für eine Property, die eine
// Entity nicht nennt (SPEC-190 Nr. 4): genau eine Ursache an einer Option, nie ein Teil der Eingabe.
type AbfrageFehler struct {
	Option  AbfrageOption
	Ursache AbfrageUrsache
}

// klasse liefert die Sentinel der Klasse der Ursache; jede Ursache außerhalb der vier gehört zur Klasse `ungueltig`.
func (e *AbfrageFehler) klasse() error {
	switch e.Ursache {
	case AbfrageTypNichtUnterstuetzt, AbfrageOrdnungNichtUnterstuetzt:
		return ErrURLNichtUnterstuetzt
	case AbfrageNichtBeobachtet:
		return ErrFilterNichtBeobachtet
	}
	return ErrURLUngueltig
}

// Code liefert ODR-E6001 für `nicht_unterstuetzt`, ODR-E6003 für `nicht_beobachtet`, sonst ODR-E6005.
func (e *AbfrageFehler) Code() Code {
	switch e.klasse() {
	case ErrURLNichtUnterstuetzt:
		return CodeE6001
	case ErrFilterNichtBeobachtet:
		return CodeE6003
	}
	return CodeE6005
}

// Error nennt Code, Klasse, Ursache und Option.
func (e *AbfrageFehler) Error() string {
	s := "ODR-E6005: Abfrage ungültig: "
	switch e.klasse() {
	case ErrURLNichtUnterstuetzt:
		s = "ODR-E6001: Abfrage nicht unterstützt: "
	case ErrFilterNichtBeobachtet:
		s = "ODR-E6003: Abfrage nicht beobachtet: "
	}
	return s + string(e.Ursache) + " (" + string(e.Option) + ")"
}

// Is trifft genau die Sentinel der Klasse der Ursache, nie die der anderen.
func (e *AbfrageFehler) Is(target error) bool { return target == e.klasse() }

// abfrageVerstoss bildet den Fehler einer Anfrage.
func abfrageVerstoss(o AbfrageOption, u AbfrageUrsache) error {
	return &AbfrageFehler{Option: o, Ursache: u}
}
