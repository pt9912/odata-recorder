package httpodata

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// schreibeFehler liefert den Fehler des Dienstes aus (SPEC-136 Nr. 7): Kontextfehler vor
// *inbound.Fehler vor jedem anderen; ein Fehler trägt keine Antwort, deren Body wird geschlossen.
// Die Suche im Fehlerbaum ist finde: ein nil-Zeiger *inbound.Fehler gilt nicht und wird nicht entpackt.
func schreibeFehler(w http.ResponseWriter, r *http.Request, id string, sp *anfrageSpan, antwort inbound.Antwort, err error) {
	if antwort.Body != nil {
		_ = antwort.Body.Close()
	}
	if finde(err, istKontextfehler) {
		brichAb()
	}
	code, status := model.CodeE9001, http.StatusInternalServerError
	var f *inbound.Fehler
	if finde(err, func(e error) bool {
		f = fehlerKnoten(e)
		return f != nil
	}) && f.Status >= 400 && f.Status <= 599 && f.Code.StatusErlaubt(f.Status) {
		code, status = f.Code, f.Status
	}
	schreibeFehlerobjekt(w, r, id, sp, code, status)
}

// istKontextfehler meldet, ob e context.Canceled oder context.DeadlineExceeded ist.
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

// finde sucht im Baum der Fehler von err, die Wurzel zuerst, durch Unwrap() error und Unwrap() []error (auch
// errors.Join) von vorn nach hinten, und meldet, ob f für einen Fehler wahr ist; der erste Treffer gilt. Ein nil-Zeiger
// *inbound.Fehler ist kein Fehler des Baums und wird nicht entpackt (SPEC-136 Nr. 7 a): errors.Is und errors.As rufen
// sein Unwrap und greifen auf nil zu, deshalb sucht der Adapter selbst. Ein Zyklus im Fehlerbaum ist ein Programmfehler
// des Dienstes, den die Suche wie errors.Is nicht erkennt.
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

// brichAb bricht die Verbindung ab, ohne die Antwort abzuschließen (SPEC-136 Nr. 9).
func brichAb() {
	panic(http.ErrAbortHandler)
}

// schreibeFehlerobjekt ist die eine Stelle, die das Fehlerobjekt des Codes schreibt und zählt
// (SPEC-294 Nr. 5): Header aus SPEC-136 Nr. 8, dann in der Folge CountError, Status und Ende des
// Spans, WriteHeader; bei HEAD entfällt der Body.
func schreibeFehlerobjekt(w http.ResponseWriter, r *http.Request, id string, sp *anfrageSpan, code model.Code, status int) {
	obj, err := model.NeuesFehlerobjekt(code)
	if err != nil {
		brichAb()
	}
	body, err := json.Marshal(obj)
	if err != nil {
		brichAb()
	}
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Set("X-Correlation-Id", id)
	sp.zaehle(code)
	sp.vorKopf(status)
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}
