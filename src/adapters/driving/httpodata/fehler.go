package httpodata

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// schreibeFehler liefert den Fehler des Dienstes aus (SPEC-136 Nr. 7): Kontextfehler vor
// *inbound.Fehler vor jedem anderen; ein Fehler trägt keine Antwort, deren Body wird geschlossen.
func schreibeFehler(w http.ResponseWriter, r *http.Request, id string, sp *anfrageSpan, antwort inbound.Antwort, err error) {
	if antwort.Body != nil {
		_ = antwort.Body.Close()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		brichAb()
	}
	code, status := model.CodeE9001, http.StatusInternalServerError
	var f *inbound.Fehler
	if errors.As(err, &f) && f != nil && f.Status >= 400 && f.Status <= 599 && f.Code.StatusErlaubt(f.Status) {
		code, status = f.Code, f.Status
	}
	schreibeFehlerobjekt(w, r, id, sp, code, status)
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
