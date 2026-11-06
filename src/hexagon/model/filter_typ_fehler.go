package model

import "errors"

// FilterTypUrsache ist die Ursache eines FilterTypFehler (SPEC-167 Nr. 6).
type FilterTypUrsache string

// Die vier Ursachen der Klasse `ungueltig` (ODR-E6005, 400).
const (
	FilterEigenschaftUnbekannt FilterTypUrsache = "eigenschaft_unbekannt"
	FilterNichtBoolesch        FilterTypUrsache = "nicht_boolesch"
	FilterTypUnvertraeglich    FilterTypUrsache = "typ_unvertraeglich"
	FilterLiteralBereich       FilterTypUrsache = "literal_bereich"
)

// Die zwei Ursachen der Klasse `nicht_unterstuetzt` (ODR-E6001, 501).
const (
	FilterTypNichtUnterstuetzt      FilterTypUrsache = "typ_nicht_unterstuetzt"
	FilterOperatorNichtUnterstuetzt FilterTypUrsache = "operator_nicht_unterstuetzt"
)

// FilterNichtBeobachtet ist die Ursache, die ErrFilterNichtBeobachtet trägt (ODR-E6003, 422).
const FilterNichtBeobachtet FilterTypUrsache = "nicht_beobachtet"

// AlleFilterTypUrsachen liefert die sieben Ursachen.
func AlleFilterTypUrsachen() []FilterTypUrsache {
	return []FilterTypUrsache{
		FilterEigenschaftUnbekannt, FilterNichtBoolesch, FilterTypUnvertraeglich, FilterLiteralBereich,
		FilterTypNichtUnterstuetzt, FilterOperatorNichtUnterstuetzt, FilterNichtBeobachtet,
	}
}

// ErrFilterNichtBeobachtet erkennt den FilterTypFehler der Ursache `nicht_beobachtet`; die Abbildung macht daraus ODR-E6003.
var ErrFilterNichtBeobachtet = errors.New("model: Filter nennt eine nicht beobachtete Property")

// FilterTypFehler ist der Fehler der Typprüfung des Filter-AST (SPEC-167 Nr. 6): genau eine Ursache, nie ein Teil
// der Eingabe.
type FilterTypFehler struct {
	Ursache FilterTypUrsache
}

// klasse liefert die Sentinel der Klasse der Ursache; jede Ursache außerhalb der sieben gehört zur Klasse `ungueltig`.
func (e *FilterTypFehler) klasse() error {
	switch e.Ursache {
	case FilterTypNichtUnterstuetzt, FilterOperatorNichtUnterstuetzt:
		return ErrURLNichtUnterstuetzt
	case FilterNichtBeobachtet:
		return ErrFilterNichtBeobachtet
	}
	return ErrURLUngueltig
}

// Code liefert ODR-E6001 für `nicht_unterstuetzt`, ODR-E6003 für `nicht_beobachtet`, sonst ODR-E6005.
func (e *FilterTypFehler) Code() Code {
	switch e.klasse() {
	case ErrURLNichtUnterstuetzt:
		return CodeE6001
	case ErrFilterNichtBeobachtet:
		return CodeE6003
	}
	return CodeE6005
}

// Error nennt Code, Klasse und Ursache.
func (e *FilterTypFehler) Error() string {
	switch e.klasse() {
	case ErrURLNichtUnterstuetzt:
		return "ODR-E6001: Filter nicht unterstützt: " + string(e.Ursache)
	case ErrFilterNichtBeobachtet:
		return "ODR-E6003: Filter nicht beobachtet: " + string(e.Ursache)
	}
	return "ODR-E6005: Filter ungültig: " + string(e.Ursache)
}

// Is trifft genau die Sentinel der Klasse der Ursache, nie die der anderen.
func (e *FilterTypFehler) Is(target error) bool { return target == e.klasse() }
