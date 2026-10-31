package replay

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// bildeKey bildet den Match-Key aus dem Request (SPEC-135 Nr. 7). Ein Request mit Body ist ODR-E5001.
func (r *Replayer) bildeKey(a inbound.Anfrage, d Dienst, pfad string) (string, error) {
	if a.Body != nil && a.Laenge != 0 {
		return "", mitCode(model.CodeE5001, nil)
	}
	return r.schluessel(a, d, pfad, nil, model.RedactionNone)
}

// schluessel ruft BildeSchluessel mit den Header des Requests wie empfangen (SPEC-135 Nr. 7 d).
func (r *Replayer) schluessel(a inbound.Anfrage, d Dienst, pfad string, body []byte, rr model.RedactionStatus) (string, error) {
	s, err := model.BildeSchluessel(model.SchluesselEingabe{
		Method: a.Methode, ServiceID: d.ID, Path: pfad, RawQuery: a.RawQuery, Header: a.Header,
		Body: body, RequestRedaction: rr,
	}, r.k.Regeln)
	switch {
	case errors.Is(err, model.ErrKodierung):
		return "", mitCode(model.CodeE5006, err)
	case err != nil:
		return "", mitCode(model.CodeE9001, err)
	}
	return s.MatchKey, nil
}
