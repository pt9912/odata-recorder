package inbound

import "github.com/pt9912/odata-recorder/src/hexagon/model"

// Fehler ist ein Fehler eines Use Case mit Code (SPEC-121 Nr. 14): Status ist der HTTP-Status
// der Antwort, Ursache bleibt über errors.Is erreichbar.
type Fehler struct {
	Code    model.Code
	Status  int
	Ursache error
}

// Error ist der Code, ein Doppelpunkt und die message des Codes (SPEC-109 Nr. 3); die Ursache
// steht nicht darin.
func (e *Fehler) Error() string { return e.Code.String() + ": " + e.Code.Message() }

// Unwrap macht die Ursache erreichbar.
func (e *Fehler) Unwrap() error { return e.Ursache }
