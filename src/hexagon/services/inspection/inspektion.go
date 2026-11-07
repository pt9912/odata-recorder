// Package inspection ist der Inspektionsdienst: er liest für die Admin-API und setzt
// InspectionUseCase um (SPEC-236). Er importiert Domain und Ports, keinen anderen Use Case und
// keinen Adapter, und er verändert nichts.
package inspection

import (
	"context"
	"errors"
	"maps"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// StandardLimit ist das Limit einer Seite ohne Angabe, MaxLimit die Obergrenze (SPEC-236 Nr. 3).
const (
	StandardLimit = 100
	MaxLimit      = 1000
)

// Konfig ist die Konfiguration des Dienstes (SPEC-236 Nr. 1). Exchanges dient allein dem Lesen
// der Exchanges einer Session, Coverage dem der Coverage einer Baseline, Zaehler dem der Zähler je
// Fehlercode. Modus und Einstellungen liefert der Composition Root als fertigen Wert.
type Konfig struct {
	Exchanges     outbound.RecordingRepository
	Coverage      outbound.CoverageRepository
	Zaehler       support.TelemetryReader
	Modus         string
	Einstellungen map[string]string
}

// Inspektion setzt die Operationen von InspectionUseCase um. Sie hält keinen Zustand außer ihrer
// Konfiguration (SPEC-236 Nr. 10).
type Inspektion struct {
	k Konfig
}

// Erlaubt liefert die Schlüssel der Einstellungen, die die Diagnose zeigt, in der Folge von
// SPEC-236 Nr. 7; jeder andere Schlüssel wird verworfen. Jeder Aufruf liefert einen frischen Slice.
func Erlaubt() []string {
	return []string{
		"REPLAY_STRATEGY", "REPLAY_ASSIGNMENT", "REPLAY_REDACTED_BODIES", "MATCH_VARY_HEADERS",
		"MATCH_INVALID_ENCODING", "MATCH_BODY_MODE", "REDACT_HEADERS", "REDACT_QUERY_PARAMS",
		"REDACT_JSON_PATHS", "MAX_CAPTURE_BODY", "UPSTREAM_TIMEOUT", "UPSTREAM_TLS_VERIFY",
		"UPSTREAM_TLS_INSECURE_SKIP_VERIFY", "SHUTDOWN_TIMEOUT", "RETENTION_MAX_AGE",
		"SQLITE_BUSY_TIMEOUT", "VOLATILE_HEADERS", "LINK_REWRITE",
	}
}

// New prüft die Konfiguration und kopiert die Einstellungen (SPEC-236 Nr. 1).
func New(k Konfig) (*Inspektion, error) {
	if k.Exchanges == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Exchanges fehlt"))
	}
	if k.Coverage == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Coverage fehlt"))
	}
	if k.Zaehler == nil {
		return nil, errors.Join(model.ErrUngueltig, errors.New("Zaehler fehlt"))
	}
	switch k.Modus {
	case "record", "replay", "mock", "hybrid":
	default:
		return nil, errors.Join(model.ErrUngueltig, errors.New("Modus ist keiner der vier"))
	}
	k.Einstellungen = maps.Clone(k.Einstellungen)
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

// Coverage liest die Einträge der Baseline mit genau einem Aufruf von List und gibt sie in der
// Reihenfolge des Ports weiter (SPEC-236 Nr. 6); Fehler des Ports gehen mit dem Gegenstand Snapshot
// nach Nr. 8, im Fehlerfall ist das Ergebnis nil.
func (s *Inspektion) Coverage(ctx context.Context, baselineID string) ([]inbound.CoverageEintrag, error) {
	if ctx == nil {
		return nil, ctxNil()
	}
	if baselineID == "" {
		return nil, errors.Join(model.ErrUngueltig, errors.New("baselineID ist leer"))
	}
	liste, err := s.k.Coverage.List(ctx, baselineID)
	if err != nil {
		return nil, mitCode(err, outbound.GegenstandSnapshot)
	}
	eintraege := make([]inbound.CoverageEintrag, 0, len(liste))
	for _, c := range liste {
		eintraege = append(eintraege, inbound.CoverageEintrag{EntitySet: c.EntitySet, Status: c.Status, Evidenz: c.EvidenceJSON})
	}
	return eintraege, nil
}

// Diagnose prüft den Kontext einmal und liest die Zähler mit genau einem Aufruf von Snapshot
// (SPEC-236 Nr. 7): Einstellungen sind die Einträge der Konfiguration mit Schlüsseln aus Erlaubt,
// Fehler ein Eintrag je Code von model.AlleCodes().
func (s *Inspektion) Diagnose(ctx context.Context) (inbound.Diagnose, error) {
	if ctx == nil {
		return inbound.Diagnose{}, ctxNil()
	}
	if err := ctx.Err(); err != nil {
		return inbound.Diagnose{}, err
	}
	stand := s.k.Zaehler.Snapshot()
	einstellungen := map[string]string{}
	for _, schluessel := range Erlaubt() {
		if wert, da := s.k.Einstellungen[schluessel]; da {
			einstellungen[schluessel] = wert
		}
	}
	codes := model.AlleCodes()
	fehler := make([]inbound.FehlerZaehler, 0, len(codes))
	for _, c := range codes {
		fehler = append(fehler, inbound.FehlerZaehler{Code: c, Anzahl: stand.Errors[c]})
	}
	return inbound.Diagnose{Modus: s.k.Modus, Einstellungen: einstellungen, Fehler: fehler}, nil
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
