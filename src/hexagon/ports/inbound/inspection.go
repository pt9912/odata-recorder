package inbound

import (
	"context"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Seite wählt die Exchanges einer Session (SPEC-236 Nr. 3): Nach ist die sequence_no, hinter der
// die Seite beginnt (0: am Anfang), Limit die gewünschte Zahl der Einträge (0: Standard).
type Seite struct {
	Limit, Nach int
}

// ExchangeSeite ist die Antwort auf Exchanges (SPEC-236 Nr. 4): Limit ist das wirksame Limit,
// Gesamt die Zahl aller Exchanges der Session, Weiter der Zeiger der nächsten Seite (0: keine).
type ExchangeSeite struct {
	Eintraege             []ExchangeZeile
	Limit, Gesamt, Weiter int
}

// ExchangeZeile führt genau die Felder der Allowlist des Exchange (SPEC-236 Nr. 5); kein Header,
// keine Query und keine Blob-Referenz.
type ExchangeZeile struct {
	ID             string
	SequenceNo     int
	Method, Path   string
	StartedAt      time.Time
	FinishedAt     time.Time
	KeyRedacted    bool
	RequestBody    model.BodyState
	ResponseBody   model.BodyState
	ResponseStatus int
	Capture        model.CaptureStatus
	RequestRedact  model.RedactionStatus
	ResponseRedact model.RedactionStatus
	Integrity      model.BodyIntegrity
	Equivalence    model.ReplayEquivalence
}

// CoverageEintrag ist ein Eintrag der Coverage einer Baseline (SPEC-236 Nr. 6); Evidenz ist der
// EvidenceJSON des Eintrags, byte-gleich.
type CoverageEintrag struct {
	EntitySet string
	Status    model.CoverageStatus
	Evidenz   string
}

// FehlerZaehler ist der Zähler eines Fehlercodes (SPEC-236 Nr. 7).
type FehlerZaehler struct {
	Code   model.Code
	Anzahl uint64
}

// Diagnose ist die Antwort auf Diagnose (SPEC-236 Nr. 7): der Modus, die Einstellungen aus der
// Allowlist und ein Zähler je Code von model.AlleCodes() in dessen Reihenfolge.
type Diagnose struct {
	Modus         string
	Einstellungen map[string]string
	Fehler        []FehlerZaehler
}

// InspectionUseCase liest für die Admin-API, ohne etwas zu verändern (SPEC-236). Jeder Fehler mit
// Code ist ein *Fehler; model.ErrUngueltig und die Kontextfehler kehren ohne Code zurück.
type InspectionUseCase interface {
	// Exchanges liefert die Exchanges einer Session seitenweise, in der Folge ihrer sequence_no.
	Exchanges(ctx context.Context, sessionID string, seite Seite) (ExchangeSeite, error)
	// Coverage liefert die Einträge der Coverage einer Baseline in der Reihenfolge des Ports.
	Coverage(ctx context.Context, baselineID string) ([]CoverageEintrag, error)
	// Diagnose liefert Modus, Einstellungen aus der Allowlist und die Zähler je Fehlercode.
	Diagnose(ctx context.Context) (Diagnose, error)
}
