package admin

import (
	"context"
	"net/http"
	"reflect"
	"slices"

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
// und schreibt weder Fehlerobjekt noch Antwort (Nr. 9), sonst bestimmt abbilde Code und Status. allow ist der Wert des
// Headers Allow, den ein Fehler mit 405 trägt: die Methoden aller Routen des Pfades (Nr. 5).
func (s *Server) schreibeFehlerDerRoute(w http.ResponseWriter, err error, allow string) {
	if finde(err, istKontextfehler) {
		panic(http.ErrAbortHandler)
	}
	code, status := abbilde(err)
	s.schreibeFehler(w, code, status, allow)
}

// abbilde liefert Code und Status zu einem Fehler ohne Kontextfehler (SPEC-238 Nr. 6 a 2 bis 4): der erste
// *inbound.Fehler der Suche gilt; ist er ungültig (Code außerhalb der Tabelle, Status außerhalb von 400 bis 599, ein
// Status, den der Code nicht trägt), ist das ein Bruch des Vertrags; ohne *inbound.Fehler ist model.ErrUngueltig
// ODR-E2102 mit 400, jeder andere Fehler ODR-E9001 mit 500.
func abbilde(err error) (model.Code, int) {
	var f *inbound.Fehler
	if finde(err, func(e error) bool {
		f = fehlerKnoten(e)
		return f != nil
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

// fehlerKnoten liefert den *inbound.Fehler, den e selbst darstellt: e ist ein solcher Zeiger, oder die Methode As von e
// setzt einen (wie errors.As). Ein nil-Zeiger ist kein Fehler: das Ergebnis ist nil.
func fehlerKnoten(e error) *inbound.Fehler {
	if z, ok := e.(*inbound.Fehler); ok {
		return z
	}
	if m, ok := e.(interface{ As(target any) bool }); ok {
		var z *inbound.Fehler
		if m.As(&z) {
			return z
		}
	}
	return nil
}

// zeigerAdresse liefert die Adresse eines Fehlers, der ein Zeiger auf einen Wert mit Größe ist; sonst ist ok falsch.
func zeigerAdresse(err error) (adresse uintptr, ok bool) {
	v := reflect.ValueOf(err)
	if v.Kind() != reflect.Pointer || v.Type().Elem().Size() == 0 {
		return 0, false
	}
	return v.Pointer(), true
}

// finde sucht im Baum der Fehler von err, die Wurzel zuerst, durch Unwrap() error und Unwrap() []error (auch
// errors.Join) von vorn nach hinten, und meldet, ob f für einen Fehler wahr ist; der erste Treffer gilt. Ein nil-Zeiger
// *inbound.Fehler ist kein Fehler des Baums und wird nicht entpackt (SPEC-238 Nr. 6 a 2); errors.Is und errors.As
// entpacken ihn und greifen dabei auf nil zu, deshalb sucht der Adapter selbst. Ein Zeiger, der auf dem Weg von der
// Wurzel schon vorkam (sein Unwrap zeigt auf sich oder einen Vorfahren), wird nicht noch einmal betreten. Prüfgrenze:
// einen Zyklus über Werte, die keine Zeiger sind, erkennt die Suche nicht.
func finde(err error, f func(error) bool) bool {
	return findeAuf(err, f, nil)
}

// findeAuf ist finde mit den Adressen der Fehler auf dem Weg von der Wurzel zu err.
func findeAuf(err error, f func(error) bool, weg []uintptr) bool {
	if err == nil {
		return false
	}
	if z, ok := err.(*inbound.Fehler); ok && z == nil {
		return false
	}
	if a, ok := zeigerAdresse(err); ok {
		if slices.Contains(weg, a) {
			return false
		}
		weg = append(weg, a)
	}
	if f(err) {
		return true
	}
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		return findeAuf(u.Unwrap(), f, weg)
	case interface{ Unwrap() []error }:
		for _, e := range u.Unwrap() {
			if findeAuf(e, f, weg) {
				return true
			}
		}
	}
	return false
}
