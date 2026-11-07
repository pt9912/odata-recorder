// Package inspection ist der Inspektionsdienst: er liest für die Admin-API und setzt
// InspectionUseCase um (SPEC-236). Er importiert Domain und Ports, keinen anderen Use Case und
// keinen Adapter, und er verändert nichts.
package inspection

import (
	"context"
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// StandardLimit ist das Limit einer Seite ohne Angabe, MaxLimit die Obergrenze (SPEC-236 Nr. 3).
const (
	StandardLimit = 100
	MaxLimit      = 1000
)

// Konfig ist die Konfiguration des Dienstes (SPEC-236 Nr. 1). Exchanges dient allein dem Lesen
// der Exchanges einer Session.
type Konfig struct {
	Exchanges outbound.RecordingRepository
}

// Inspektion setzt die Operationen von InspectionUseCase um. Sie hält keinen Zustand außer ihrer
// Konfiguration (SPEC-236 Nr. 10).
type Inspektion struct {
	k Konfig
}

// New prüft die Konfiguration (SPEC-236 Nr. 1).
func New(k Konfig) (*Inspektion, error) {
	if k.Exchanges == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Exchanges fehlt"))
	}
	return &Inspektion{k: k}, nil
}

func ctxNil() error {
	return errors.Join(model.ErrUngueltig, errors.New("Kontext fehlt"))
}

// Exchanges prüft die Argumente in der Folge aus SPEC-236 Nr. 3, liest die Exchanges der Session
// mit genau einem Aufruf von ListBySession und schneidet die Seite aus der Liste (Nr. 4).
func (s *Inspektion) Exchanges(ctx context.Context, sessionID string, seite inbound.Seite) (inbound.ExchangeSeite, error) {
	if ctx == nil {
		return inbound.ExchangeSeite{}, ctxNil()
	}
	if seite.Nach < 0 {
		return inbound.ExchangeSeite{}, errors.Join(model.ErrUngueltig, errors.New("Nach ist negativ"))
	}
	if seite.Limit < 0 {
		return inbound.ExchangeSeite{}, errors.Join(model.ErrUngueltig, errors.New("Limit ist negativ"))
	}
	alle, err := s.k.Exchanges.ListBySession(ctx, sessionID)
	if err != nil {
		return inbound.ExchangeSeite{}, mitCode(err, outbound.GegenstandSession)
	}
	return schneide(alle, seite.Nach, wirksamesLimit(seite.Limit)), nil
}

// wirksamesLimit ist StandardLimit für 0, höchstens MaxLimit und sonst der Wert (SPEC-236 Nr. 3).
func wirksamesLimit(limit int) int {
	switch {
	case limit == 0:
		return StandardLimit
	case limit > MaxLimit:
		return MaxLimit
	}
	return limit
}

// schneide bildet die Seite: die Exchanges mit sequence_no über nach, höchstens limit, und Weiter
// als sequence_no des letzten Eintrags, wenn danach noch ein Exchange folgt (SPEC-236 Nr. 4).
func schneide(alle []model.Exchange, nach, limit int) inbound.ExchangeSeite {
	seite := inbound.ExchangeSeite{Eintraege: []inbound.ExchangeZeile{}, Limit: limit, Gesamt: len(alle)}
	rest := false
	for _, e := range alle {
		if e.SequenceNo <= nach {
			continue
		}
		if len(seite.Eintraege) == limit {
			rest = true
			break
		}
		seite.Eintraege = append(seite.Eintraege, zeile(e))
	}
	if rest {
		seite.Weiter = seite.Eintraege[len(seite.Eintraege)-1].SequenceNo
	}
	return seite
}

// zeile kopiert die Felder der Allowlist (SPEC-236 Nr. 5).
func zeile(e model.Exchange) inbound.ExchangeZeile {
	return inbound.ExchangeZeile{
		ID: e.ID, SequenceNo: e.SequenceNo, Method: e.Method, Path: e.Path,
		StartedAt: e.StartedAt, FinishedAt: e.FinishedAt, KeyRedacted: e.KeyRedacted,
		RequestBody: e.RequestBody, ResponseBody: e.ResponseBody, ResponseStatus: e.ResponseStatus,
		Capture: e.Capture, RequestRedact: e.RequestRedact, ResponseRedact: e.ResponseRedact,
		Integrity: e.Integrity, Equivalence: e.Equivalence,
	}
}
