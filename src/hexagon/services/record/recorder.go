// Package record ist der Record-Dienst: er setzt RecordUseCase um (SPEC-121). Er importiert
// Domain, Ports und das Paket capture, keine HTTP-Bibliothek und keinen Adapter.
package record

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/record/capture"
)

var _ inbound.RecordUseCase = (*Recorder)(nil)

// Dienst ist ein Service der Konfiguration: Kennung, Root und Upstream (SPEC-121 Nr. 1).
type Dienst struct {
	ID          string
	Root        string
	UpstreamURL string
}

// Konfig ist die Konfiguration des Dienstes (SPEC-121 Nr. 1).
type Konfig struct {
	Dienste   []Dienst
	Sessions  outbound.SessionRepository
	Exchanges outbound.RecordingRepository
	Upstream  outbound.UpstreamClient
	Pipeline  *capture.Pipeline
	Regeln    model.SchluesselRegeln
	Uhr       support.Clock
	IDs       support.IDGenerator
	Log       *slog.Logger
}

// Recorder setzt RecordUseCase um. Der Zustand je Session (Annahme, Drain, Zahl, Bericht) bleibt
// für die Lebenszeit des Dienstes im Speicher (SPEC-121 Ungeprüft).
type Recorder struct {
	k Konfig
	// mu schützt sessions und die Felder von sessionZustand außer Zahl und Sperre der Anhänge.
	mu       sync.Mutex
	sessions map[string]*sessionZustand
}

func ungueltig(text string) error {
	return errors.Join(model.ErrUngueltig, errors.New(text))
}

// New prüft die Konfiguration (SPEC-121 Nr. 1); Form und Überlappung der Roots prüft sie nicht.
func New(k Konfig) (*Recorder, error) {
	if err := pruefeKonfig(k); err != nil {
		return nil, err
	}
	k.Dienste = slices.Clone(k.Dienste)
	return &Recorder{k: k, sessions: map[string]*sessionZustand{}}, nil
}

func pruefeKonfig(k Konfig) error {
	switch {
	case len(k.Dienste) == 0:
		return ungueltig("kein Dienst")
	case slices.ContainsFunc(k.Dienste, func(d Dienst) bool { return d.ID == "" || d.Root == "" || d.UpstreamURL == "" }):
		return ungueltig("Dienst ohne Kennung, Root oder Upstream")
	case k.Sessions == nil || k.Exchanges == nil || k.Upstream == nil || k.Pipeline == nil:
		return ungueltig("Port oder Pipeline fehlt")
	case k.Uhr == nil || k.IDs == nil || k.Log == nil:
		return ungueltig("Uhr, Kennungen oder Log fehlt")
	}
	return pruefeRegeln(k.Regeln)
}

func pruefeRegeln(r model.SchluesselRegeln) error {
	switch {
	case r.InvalidEncoding != model.KodierungLiteral && r.InvalidEncoding != model.KodierungReject:
		return ungueltig("MATCH_INVALID_ENCODING ist unbekannt")
	case r.BodyMode != model.BodyBytes && r.BodyMode != model.BodyJSONCanonical:
		return ungueltig("MATCH_BODY_MODE ist unbekannt")
	}
	return nil
}

// Record prüft den Request in der Reihenfolge von SPEC-121 Nr. 2, nimmt ihn an, leitet ihn
// weiter und kehrt mit dem Kopf der Antwort zurück.
func (r *Recorder) Record(ctx context.Context, a inbound.Anfrage) (inbound.Antwort, error) {
	if ctx == nil || a.Methode == "" || a.Laenge < -1 || a.Body == nil && a.Laenge > 0 {
		return inbound.Antwort{}, ungueltig("Anfrage, die kein HTTP-Adapter liefert")
	}
	w, err := r.pruefe(ctx, a)
	if err != nil {
		return inbound.Antwort{}, err
	}
	lctx, abbrechen := context.WithCancelCause(ctx)
	l := &lauf{abbrechen: abbrechen, ursache: func() error { return context.Cause(lctx) }}
	sessionID, sz, err := r.nimmAn(w.kandidaten, l)
	if err != nil {
		abbrechen(errEnde)
		return inbound.Antwort{}, err
	}
	x := &austausch{
		r: r, sz: sz, l: l, dienst: w.dienst, anfrage: a, vorab: w.vorab, client: ctx.Err,
		e: model.Exchange{
			ID:             r.k.IDs.ExchangeID(),
			SessionID:      sessionID,
			Method:         a.Methode,
			Path:           w.pfad,
			RawQuery:       w.vorab.RawQuery,
			RequestHeaders: w.vorab.Header,
			StartedAt:      r.k.Uhr.Now(),
		},
	}
	return x.leite(lctx, context.WithoutCancel(ctx))
}

// mitCode bildet den Fehlerwert eines Codes mit seinem einen Status (SPEC-121 Nr. 14).
func mitCode(c model.Code, ursache error) *inbound.Fehler {
	return &inbound.Fehler{Code: c, Status: c.Stati()[0], Ursache: ursache}
}
