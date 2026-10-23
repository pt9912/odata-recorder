package session

import (
	"context"
	"errors"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// nichtOffen ist der Fehler für das Beenden einer Session, die nicht offen ist; der Dienst bildet
// ihn aus dem gelesenen Status und ohne Ursache (SPEC-124 Nr. 11 c).
func nichtOffen() error {
	c := model.CodeE7011
	return &inbound.Fehler{Code: c, Status: c.Stati()[0]}
}

// statusNach wendet die Statusregel an (SPEC-124 Nr. 12): complete genau dann, wenn kein Exchange
// unangehängt blieb und jeder gelieferte Exchange complete trägt; Abgebrochen geht nicht ein.
func statusNach(b inbound.DrainBericht, exchanges []model.Exchange) model.SessionStatus {
	if b.NichtAngehaengt != 0 {
		return model.SessionIncomplete
	}
	for _, e := range exchanges {
		if e.Capture != model.CaptureComplete {
			return model.SessionIncomplete
		}
	}
	return model.SessionComplete
}

// Beenden liest die Session, weist eine nicht offene ab, ruft Drain bei noch offener Session und
// setzt erst danach den Status nach der Statusregel (SPEC-124 Nr. 11). Das Lesen der Exchanges und
// der Statuswechsel laufen unter einem Kontext ohne den Abbruch des Aufrufers (Nr. 11 e).
func (v *Verwaltung) Beenden(ctx context.Context, id string) (model.Session, error) {
	if ctx == nil {
		return model.Session{}, ctxNil()
	}
	s, err := v.k.Sessions.Get(ctx, id)
	if err != nil {
		return model.Session{}, mitCode(err, outbound.GegenstandSession)
	}
	if s.Status != model.SessionOpen {
		return model.Session{}, nichtOffen()
	}
	bericht, err := v.k.Record.Drain(ctx, id, v.k.Frist)
	if err != nil {
		return model.Session{}, err
	}
	nach := context.WithoutCancel(ctx)
	exchanges, err := v.k.Exchanges.ListBySession(nach, id)
	if err != nil {
		return model.Session{}, mitCode(err, outbound.GegenstandSession)
	}
	status := statusNach(bericht, exchanges)
	if err := v.k.Sessions.SetStatus(nach, id, status); err != nil {
		return model.Session{}, mitCode(err, outbound.GegenstandSession)
	}
	s.Status = status
	return s, nil
}

// bereitsBeendet sagt, ob err ODR-E7011 ist: eine Session, die nicht mehr offen war (SPEC-124 Nr. 14 d).
func bereitsBeendet(err error) bool {
	var f *inbound.Fehler
	return errors.As(err, &f) && f.Code == model.CodeE7011
}

// CloseAll liest die offenen Sessions einmal und beendet jede in einer eigenen Goroutine mit dem
// ctx des Aufrufs; der Fehler einer Session hält die übrigen nicht auf (SPEC-124 Nr. 14).
func (v *Verwaltung) CloseAll(ctx context.Context) (inbound.CloseBericht, error) {
	if ctx == nil {
		return inbound.CloseBericht{}, ctxNil()
	}
	offen, err := v.k.Sessions.FindOpen(ctx)
	if err != nil {
		return inbound.CloseBericht{}, mitCode(err, outbound.GegenstandSonstiges)
	}
	fehler := make([]error, len(offen))
	var wg sync.WaitGroup
	for i, s := range offen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, fehler[i] = v.Beenden(ctx, s.ID)
		}()
	}
	wg.Wait()
	var bericht inbound.CloseBericht
	var erster error
	for i, err := range fehler {
		switch {
		case err == nil:
			bericht.Beendet++
		case bereitsBeendet(err):
		default:
			bericht.Fehlgeschlagen++
			if erster == nil {
				erster = err
			}
			v.protokolliereBeenden(ctx, offen[i], err)
		}
	}
	return bericht, erster
}
