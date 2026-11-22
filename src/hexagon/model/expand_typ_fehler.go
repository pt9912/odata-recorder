package model

// ExpandUrsache ist die Ursache eines ExpandFehler (SPEC-263 Nr. 3).
type ExpandUrsache string

// Die zwei Ursachen der Klasse `ungueltig` (ODR-E6005, 400).
const (
	ExpandNavigationUnbekannt ExpandUrsache = "navigation_unbekannt"
	ExpandKeineNavigation     ExpandUrsache = "keine_navigation"
)

// ExpandNavigationNichtUnterstuetzt ist die Ursache der Klasse `nicht_unterstuetzt` (ODR-E6001, 501).
const ExpandNavigationNichtUnterstuetzt ExpandUrsache = "navigation_nicht_unterstuetzt"

// ExpandFehler ist der Fehler der Prüfung von `$expand` gegen den Typ (SPEC-263 Nr. 3): genau eine Ursache, nie ein Teil
// der Eingabe.
type ExpandFehler struct {
	Ursache ExpandUrsache
}

// klasse liefert die Sentinel der Klasse der Ursache; jede Ursache außerhalb der drei gehört zur Klasse `ungueltig`.
func (e *ExpandFehler) klasse() error {
	if e.Ursache == ExpandNavigationNichtUnterstuetzt {
		return ErrURLNichtUnterstuetzt
	}
	return ErrURLUngueltig
}

// Code liefert ODR-E6001 für `nicht_unterstuetzt`, sonst ODR-E6005.
func (e *ExpandFehler) Code() Code {
	if e.klasse() == ErrURLNichtUnterstuetzt {
		return CodeE6001
	}
	return CodeE6005
}

// Error nennt Code, Klasse und Ursache.
func (e *ExpandFehler) Error() string {
	if e.klasse() == ErrURLNichtUnterstuetzt {
		return "ODR-E6001: Expand nicht unterstützt: " + string(e.Ursache)
	}
	return "ODR-E6005: Expand ungültig: " + string(e.Ursache)
}

// Is trifft genau die Sentinel der Klasse der Ursache, nie die der anderen.
func (e *ExpandFehler) Is(target error) bool { return target == e.klasse() }
