package admin

import (
	"context"
	"net/http"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// ergebnisRoute ist, was ein Aufruf des Handlers lieferte; panik ist wahr, wenn er in Panik geriet (dann sind antwort
// und err leer).
type ergebnisRoute struct {
	antwort Antwort
	err     error
	panik   bool
}

// rufeHandler ruft h genau einmal und fängt jede Panik, auch eine mit dem Wert nil und mit http.ErrAbortHandler
// (SPEC-238 Nr. 6 c). Der Wert der Panik wird weder gelesen noch gespeichert noch ausgegeben.
func rufeHandler(ctx context.Context, h Handler, a Anfrage) (e ergebnisRoute) {
	zurueck := false
	defer func() {
		if !zurueck {
			_ = recover()
			e = ergebnisRoute{panik: true}
		}
	}()
	e.antwort, e.err = h(ctx, a)
	zurueck = true
	return e
}

// schreibeFehlerDerRoute bildet den Fehler einer Route ab (SPEC-238 Nr. 6 a): ein Kontextfehler bricht die Verbindung ab
// und schreibt weder Fehlerobjekt noch Antwort (Nr. 9), sonst bestimmt abbilde Code und Status.
func (s *Server) schreibeFehlerDerRoute(w http.ResponseWriter, err error) {
	if finde(err, istKontextfehler) {
		panic(http.ErrAbortHandler)
	}
	code, status := abbilde(err)
	s.schreibeFehler(w, code, status, "", "")
}

// abbilde liefert Code und Status zu einem Fehler ohne Kontextfehler (SPEC-238 Nr. 6 a 2 bis 4): ein gültiger
// *inbound.Fehler gilt; ein ungültiger (Code außerhalb der Tabelle, Status außerhalb von 400 bis 599, ein Status, den der
// Code nicht trägt) ist ein Bruch des Vertrags; ohne *inbound.Fehler ist model.ErrUngueltig ODR-E2102 mit 400, jeder
// andere Fehler ODR-E9001 mit 500.
func abbilde(err error) (model.Code, int) {
	var f *inbound.Fehler
	if finde(err, func(e error) bool {
		z, ok := e.(*inbound.Fehler)
		if ok {
			f = z
		}
		return ok
	}) {
		if f.Status >= http.StatusBadRequest && f.Status <= 599 && f.Code.StatusErlaubt(f.Status) {
			return f.Code, f.Status
		}
		return model.CodeE9001, http.StatusInternalServerError
	}
	if finde(err, func(e error) bool { return gleichOderIs(e, model.ErrUngueltig) }) {
		return model.CodeE2102, http.StatusBadRequest
	}
	return model.CodeE9001, http.StatusInternalServerError
}

// istKontextfehler meldet, ob e der Kontextfehler context.Canceled oder context.DeadlineExceeded ist (SPEC-238 Nr. 6 a 1).
func istKontextfehler(e error) bool {
	return gleichOderIs(e, context.Canceled) || gleichOderIs(e, context.DeadlineExceeded)
}

// gleichOderIs meldet, ob e der Fehler ziel ist: gleich oder nach der Methode Is, ohne e zu entpacken.
func gleichOderIs(e, ziel error) bool {
	if e == ziel {
		return true
	}
	m, ok := e.(interface{ Is(target error) bool })
	return ok && m.Is(ziel)
}

// finde sucht im Baum der Fehler von err, die Wurzel zuerst, durch Unwrap() error und Unwrap() []error (auch
// errors.Join), und meldet, ob f für einen Fehler wahr ist. Ein nil-Zeiger *inbound.Fehler ist kein Fehler des Baums und
// wird nicht entpackt (SPEC-238 Nr. 6 a 2); errors.Is und errors.As entpacken ihn und greifen dabei auf nil zu, deshalb
// sucht der Adapter selbst.
func finde(err error, f func(error) bool) bool {
	if err == nil {
		return false
	}
	if z, ok := err.(*inbound.Fehler); ok && z == nil {
		return false
	}
	if f(err) {
		return true
	}
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		return finde(u.Unwrap(), f)
	case interface{ Unwrap() []error }:
		for _, e := range u.Unwrap() {
			if finde(e, f) {
				return true
			}
		}
	}
	return false
}
