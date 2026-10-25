package model

import (
	"errors"
	"strconv"
)

// KonfliktUrsache ist die Ursache eines KonfliktFehler (SPEC-127 Nr. 10).
type KonfliktUrsache string

// Die zwei Ursachen: zwei bekannte Werte derselben Property oder zwei Typen derselben Entity.
const (
	KonfliktWert KonfliktUrsache = "wert"
	KonfliktTyp  KonfliktUrsache = "typ"
)

// ErrKonflikt erkennt jeden KonfliktFehler über errors.Is.
var ErrKonflikt = errors.New("model: Konflikt zwischen zwei Beobachtungen")

// KonfliktFehler ist der Widerspruch, der den Import beendet (SPEC-127 Nr. 10 h): ODR-E6101 mit den
// Kennungen der zwei Exchanges, dem Set und der Property (bei KonfliktTyp leer). Er trägt weder
// einen Wert noch den Schlüssel der Entity.
type KonfliktFehler struct {
	Ursache   KonfliktUrsache
	EntitySet string
	Property  string
	Erste     string
	Zweite    string
}

// Error nennt den Code, die Ursache und die Namen; sie stehen in der Schreibweise von strconv.Quote
// dahinter, denn Namen und Kennungen sind Text der Aufzeichnung.
func (e *KonfliktFehler) Error() string {
	s := "ODR-E6101: Baseline-Import ungültig oder inkonsistent: konflikt " + string(e.Ursache) + " " + strconv.Quote(e.EntitySet)
	if e.Property != "" {
		s += " " + strconv.Quote(e.Property)
	}
	return s + " " + strconv.Quote(e.Erste) + " " + strconv.Quote(e.Zweite)
}

// Code liefert ODR-E6101.
func (*KonfliktFehler) Code() Code { return CodeE6101 }

// Unwrap macht ErrKonflikt und ErrDatenUngueltig erreichbar.
func (*KonfliktFehler) Unwrap() []error { return []error{ErrKonflikt, ErrDatenUngueltig} }
