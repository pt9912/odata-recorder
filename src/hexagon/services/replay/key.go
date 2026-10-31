package replay

import (
	"context"
	"errors"
	"strings"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay/begrenzt"
)

// bildeKey bildet den Match-Key aus dem Request (SPEC-135 Nr. 7): zuerst der Body (a bis c), dann der Key (d).
func (r *Replayer) bildeKey(a inbound.Anfrage, d Dienst, pfad string) (string, error) {
	var (
		body []byte
		rr   = model.RedactionNone
	)
	if a.Body != nil && a.Laenge != 0 {
		var err error
		if body, rr, err = r.redigiereBody(a); err != nil {
			return "", err
		}
	}
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

// redigiereBody prüft Content-Encoding und leitet den Body durch den Redaktor (SPEC-135 Nr. 7 b, c): jeder
// Fehler ist ODR-E5001, ein Kontextfehler kehrt ohne Code zurück. Der Key entsteht aus der Ausgabe.
func (r *Replayer) redigiereBody(a inbound.Anfrage) ([]byte, model.RedactionStatus, error) {
	if !nurIdentity(werteVon(a.Header, "content-encoding")) {
		return nil, "", mitCode(model.CodeE5001, nil)
	}
	aus := begrenzt.Neu(r.k.MaxBytes)
	st, err := model.RedigiereBody(aus, a.Body,
		model.BodyEingang{ContentType: werteVon(a.Header, "content-type"), Metadata: false},
		model.BodyRegeln{Pfade: r.k.Pfade, MaxBytes: r.k.MaxBytes})
	switch {
	case errors.Is(err, context.Canceled):
		return nil, "", context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return nil, "", context.DeadlineExceeded
	case err != nil:
		return nil, "", mitCode(model.CodeE5001, err)
	}
	return aus.Bytes(), st, nil
}

// nurIdentity sagt, ob jeder Teil der Werte von Content-Encoding identity ist: Teilung an Kommas, Rand ohne
// Leerzeichen und Tabulatoren, ohne Beachtung der ASCII-Schreibweise; ein leerer Teil ist keiner (SPEC-135 Nr. 7 b).
func nurIdentity(werte []string) bool {
	for _, w := range werte {
		for _, t := range strings.Split(w, ",") {
			if asciiKlein(strings.Trim(t, " \t")) != "identity" {
				return false
			}
		}
	}
	return true
}
