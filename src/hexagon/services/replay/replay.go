// Package replay ist der Replay-Dienst: er setzt ReplayUseCase um (SPEC-135). Er importiert Domain und
// Ports, keinen Adapter, keinen anderen Dienst und keinen Port zu einem Upstream.
package replay

import (
	"context"
	"errors"
	"io"
	"slices"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

var _ inbound.ReplayUseCase = (*Replayer)(nil)

// Dienst ist ein Service der Konfiguration: Kennung und Root (SPEC-135 Nr. 1).
type Dienst struct {
	ID   string
	Root string
}

// Konfig ist die Konfiguration des Dienstes (SPEC-135 Nr. 1).
type Konfig struct {
	Dienste          []Dienst
	Sessions         outbound.SessionRepository
	Exchanges        outbound.RecordingRepository
	Blobs            outbound.BlobRepository
	Transaktion      outbound.WorkspaceTransactor
	Regeln           model.SchluesselRegeln
	Pfade            []model.JSONPfad
	MaxBytes         int64
	Strategie        model.ReplayStrategy
	RedigierteBodies bool
	Telemetrie       support.Telemetry
}

// Replayer setzt ReplayUseCase um. Er hält nach New nur seine Konfig, die er nicht ändert, keine Sperre
// und keine Goroutine (SPEC-135 Nr. 17).
type Replayer struct {
	k Konfig
}

// New prüft die Konfiguration (SPEC-135 Nr. 1) und ruft keinen Port; Form und Überlappung der Roots
// prüft sie nicht.
func New(k Konfig) (*Replayer, error) {
	if err := pruefeKonfig(k); err != nil {
		return nil, err
	}
	k.Dienste = slices.Clone(k.Dienste)
	k.Pfade = slices.Clone(k.Pfade)
	return &Replayer{k: k}, nil
}

func pruefeKonfig(k Konfig) error {
	switch {
	case len(k.Dienste) == 0:
		return ungueltig("kein Dienst")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.ID == "" || d.Root == "" }):
		return ungueltig("Dienst ohne Kennung oder Root")
	case k.Sessions == nil || k.Exchanges == nil || k.Blobs == nil || k.Transaktion == nil || k.Telemetrie == nil:
		return ungueltig("Port fehlt")
	}
	return pruefeWerte(k)
}

func pruefeWerte(k Konfig) error {
	switch {
	case k.MaxBytes < 1:
		return ungueltig("MaxBytes ist kleiner als 1")
	case !slices.Contains([]model.ReplayStrategy{model.StrategySequential, model.StrategyStrict, model.StrategyFirst, model.StrategyLast}, k.Strategie):
		return ungueltig("REPLAY_STRATEGY ist unbekannt")
	case k.Regeln.InvalidEncoding != model.KodierungLiteral && k.Regeln.InvalidEncoding != model.KodierungReject:
		return ungueltig("MATCH_INVALID_ENCODING ist unbekannt")
	case k.Regeln.BodyMode != model.BodyBytes && k.Regeln.BodyMode != model.BodyJSONCanonical:
		return ungueltig("MATCH_BODY_MODE ist unbekannt")
	case slices.Contains(k.Pfade, model.JSONPfad{}):
		return ungueltig("Pfadmuster ist der Nullwert")
	}
	return nil
}

// Replay prüft den Request in der Reihenfolge von SPEC-135 Nr. 3, wählt den Exchange, schreibt den
// Cursor in derselben Transaktion fort und liefert Status, Header und Body der Aufzeichnung.
func (r *Replayer) Replay(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	if ctx == nil || a.Methode == "" || a.Laenge < -1 || a.Body == nil && a.Laenge > 0 {
		return inbound.Antwort{}, ungueltig("Anfrage, die kein HTTP-Adapter liefert")
	}
	d, pfad, ok := dienstFuer(r.k.Dienste, a.Pfad)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE5005, nil)
	}
	ws, rec, rep, ok := zuordnung(a.Header)
	if !ok {
		return inbound.Antwort{}, mitCode(model.CodeE5003, nil)
	}
	if err := r.pruefeSession(ctx, d, rec); err != nil {
		return inbound.Antwort{}, err
	}
	key, err := r.bildeKey(a, d, pfad)
	if err != nil {
		return inbound.Antwort{}, err
	}
	liste, err := r.k.Exchanges.ListBySession(ctx, rec)
	if err != nil {
		return inbound.Antwort{}, abbilden(err, outbound.GegenstandSession)
	}
	e, strom, err := r.fortschreiben(ctx, d, ws, rep, rec, key, liste)
	if err != nil {
		return inbound.Antwort{}, err
	}
	protokolliere(ctx, r.k.Telemetrie, e, d.ID, rec, ws, a.KorrelationsID)
	return inbound.Antwort{Status: e.ResponseStatus, Header: antwortHeader(e, r.k.Regeln.VolatileHeaders), Body: strom}, nil
}

// pruefeSession liest die Recording-Session und prüft Service und Zustand (SPEC-135 Nr. 6 a).
func (r *Replayer) pruefeSession(ctx context.Context, d Dienst, id string) error {
	s, err := r.k.Sessions.Get(ctx, id)
	switch {
	case err != nil:
		return abbilden(err, outbound.GegenstandSession)
	case s.ServiceID != d.ID:
		return mitCode(model.CodeE5004, nil)
	case s.Status == model.SessionOpen:
		return mitCode(model.CodeE7009, nil)
	}
	return nil
}

// fortschreiben führt Lesen, Auswahl, Öffnen des Blobs und Schreiben des Cursors in einer Transaktion
// aus (SPEC-135 Nr. 11). Bei einem Fehler nach dem Öffnen schließt es den Strom genau einmal (Nr. 12).
func (r *Replayer) fortschreiben(ctx context.Context, d Dienst, ws, rep, rec, key string, liste []model.Exchange) (model.Exchange, io.ReadCloser, error) {
	var (
		gewaehlt model.Exchange
		strom    io.ReadCloser
		fehlerF  error
	)
	err := r.k.Transaktion.WithWorkspaceTransaction(ctx, func(st outbound.WorkspaceStores) error {
		var err error
		gewaehlt, strom, err = r.imSchreiber(ctx, st, d, ws, rep, rec, key, liste)
		fehlerF = err
		return err
	})
	if err != nil {
		if strom != nil {
			_ = strom.Close()
		}
		if fehlerF != nil && errors.Is(err, fehlerF) {
			return model.Exchange{}, nil, err
		}
		return model.Exchange{}, nil, abbilden(err, outbound.GegenstandSonstiges)
	}
	return gewaehlt, strom, nil
}

// imSchreiber ist f der Transaktion (SPEC-135 Nr. 11): Workspace lesen, Cursor lesen, wählen, Blob öffnen,
// Cursor schreiben, in dieser Reihenfolge und je höchstens einmal.
func (r *Replayer) imSchreiber(ctx context.Context, st outbound.WorkspaceStores, d Dienst, ws, rep, rec, key string, liste []model.Exchange) (model.Exchange, io.ReadCloser, error) {
	w, err := st.Workspaces.Get(ctx, ws)
	switch {
	case err != nil:
		return model.Exchange{}, nil, abbilden(err, outbound.GegenstandWorkspace)
	case w.ServiceID != d.ID:
		return model.Exchange{}, nil, mitCode(model.CodeE5004, nil)
	}
	c, vorhanden, err := liesCursor(ctx, st, ws, rep, rec)
	if err != nil {
		return model.Exchange{}, nil, err
	}
	strategie := r.k.Strategie
	if vorhanden {
		strategie = c.Strategy
	}
	e, err := waehle(strategie, key, liste, r.k.RedigierteBodies, c.Consumed, c.Position)
	if err != nil {
		return model.Exchange{}, nil, err
	}
	if e.ResponseBody != model.BodyPresent && e.ResponseBody != model.BodyAbsent {
		// Ein replayfähiger Exchange ist complete; jeder andere Zustand des Bodys verletzt eine Invariante des Speichers.
		return model.Exchange{}, nil, mitCode(model.CodeE9001, nil)
	}
	strom := leererStrom()
	if e.ResponseBody == model.BodyPresent {
		if strom, err = r.k.Blobs.Open(ctx, e.ResponseBlob); err != nil {
			return model.Exchange{}, nil, abbilden(err, outbound.GegenstandSonstiges)
		}
	}
	if strategie != model.StrategySequential && strategie != model.StrategyStrict {
		return e, strom, nil
	}
	consumed := append(slices.Clone(c.Consumed), int64(e.SequenceNo))
	neu := model.ReplayCursor{WorkspaceID: ws, ReplaySessionID: rep, RecordingSessionID: rec, ServiceID: d.ID,
		Strategy: strategie, Position: int64(len(consumed)), Consumed: consumed}
	if err := st.Cursors.Put(ctx, neu); err != nil {
		return model.Exchange{}, strom, abbilden(err, outbound.GegenstandSession)
	}
	return e, strom, nil
}

// liesCursor liest den Cursor; ein fehlender ist die neue Replay-Session mit Position 0 (SPEC-135 Nr. 10).
// Fehlend ist, was die Tabelle von SPEC-109 Nr. 5 wie ErrNichtGefunden abbildet.
func liesCursor(ctx context.Context, st outbound.WorkspaceStores, ws, rep, rec string) (model.ReplayCursor, bool, error) {
	c, err := st.Cursors.Get(ctx, ws, rep, rec)
	if err == nil {
		return c, true, nil
	}
	if fehlt(err) {
		return model.ReplayCursor{}, false, nil
	}
	return model.ReplayCursor{}, false, abbilden(err, outbound.GegenstandSonstiges)
}

// fehlt sagt, ob err der fehlende Cursor ist: ein Fehler, der ErrNichtGefunden enthält und den die Tabelle
// wie ErrNichtGefunden allein abbildet (der erste zutreffende Eintrag entscheidet).
func fehlt(err error) bool {
	if !errors.Is(err, outbound.ErrNichtGefunden) {
		return false
	}
	c, ok := outbound.Abbilden(err, outbound.GegenstandSession)
	w, _ := outbound.Abbilden(outbound.ErrNichtGefunden, outbound.GegenstandSession)
	return ok && c == w
}
