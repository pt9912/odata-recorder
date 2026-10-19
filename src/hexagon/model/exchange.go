package model

import (
	"errors"
	"fmt"
	"time"
	"unicode/utf8"
)

// BodyState beschreibt je Richtung, was mit dem Body eines Exchange geschah (SPEC-001).
type BodyState string

// Zustände eines Body.
const (
	BodyAbsent    BodyState = "absent"
	BodyPresent   BodyState = "present"
	BodyDiscarded BodyState = "discarded"
)

// CaptureStatus sagt, wie vollständig ein Exchange erfasst wurde.
type CaptureStatus string

// Zustände der Erfassung.
const (
	CaptureComplete   CaptureStatus = "complete"
	CaptureIncomplete CaptureStatus = "incomplete"
	CaptureFailed     CaptureStatus = "failed"
)

// RedactionStatus sagt, ob eine Redaction nötig war und gelang.
type RedactionStatus string

// Zustände der Redaction.
const (
	RedactionNone    RedactionStatus = "none"
	RedactionApplied RedactionStatus = "applied"
	RedactionUnsafe  RedactionStatus = "unsafe"
)

// BodyIntegrity sagt, ob der Blob zum Body passt.
type BodyIntegrity string

// Zustände der Integrität.
const (
	IntegrityVerified BodyIntegrity = "verified"
	IntegrityAbsent   BodyIntegrity = "absent"
	IntegrityInvalid  BodyIntegrity = "invalid"
)

// ReplayEquivalence sagt, wie getreu ein Replay den Exchange wiedergibt.
type ReplayEquivalence string

// Stufen der Replay-Gleichwertigkeit.
const (
	EquivalenceByte     ReplayEquivalence = "byte"
	EquivalenceSemantic ReplayEquivalence = "semantic_or_limited"
	EquivalenceNone     ReplayEquivalence = "none"
)

// Exchange ist ein aufgezeichneter HTTP-Austausch. Leere Zeichenketten bei MatchKey,
// RequestBlob und ResponseBlob heißen: keiner; ein Response-Status 0 heißt: keiner.
type Exchange struct {
	ID        string
	SessionID string
	// SequenceNo zählt lückenlos ab 1 je Session.
	SequenceNo int
	Method     string
	Path       string
	RawQuery   string
	MatchKey   string
	StartedAt  time.Time
	// FinishedAt ist die Nullzeit, solange der Austausch nicht abgeschlossen ist.
	FinishedAt      time.Time
	KeyRedacted     bool
	RequestHeaders  map[string][]string
	RequestBody     BodyState
	RequestBlob     string
	ResponseStatus  int
	ResponseHeaders map[string][]string
	ResponseBody    BodyState
	ResponseBlob    string
	Capture         CaptureStatus
	RequestRedact   RedactionStatus
	ResponseRedact  RedactionStatus
	Integrity       BodyIntegrity
	Equivalence     ReplayEquivalence
}

func ungueltig(format string, a ...any) error {
	return errors.Join(ErrUngueltig, fmt.Errorf(format, a...))
}

// Validate prüft die Invarianten aus SPEC-001 und SPEC-004.
func (e Exchange) Validate() error {
	for _, pruefung := range []func() error{e.pruefeKopf, e.pruefeWerte, e.pruefeAufzaehlungen, e.pruefeBodies, e.pruefeZustand} {
		if err := pruefung(); err != nil {
			return err
		}
	}
	return nil
}

func headerGueltig(h map[string][]string) bool {
	for name, werte := range h {
		if !utf8.ValidString(name) {
			return false
		}
		for _, w := range werte {
			if !utf8.ValidString(w) {
				return false
			}
		}
	}
	return true
}

func (e Exchange) pruefeKopf() error {
	switch {
	case e.ID == "" || e.SessionID == "":
		return ungueltig("Exchange ohne Kennung oder Session")
	case e.SequenceNo < 1:
		return ungueltig("sequence_no %d unter 1", e.SequenceNo)
	case e.Method == "" || e.Path == "":
		return ungueltig("Exchange ohne Methode oder Pfad")
	case !zeitGueltig(e.StartedAt):
		return ungueltig("Exchange ohne gültige Startzeit")
	}
	return nil
}

func (e Exchange) pruefeWerte() error {
	switch {
	case !e.FinishedAt.IsZero() && (!zeitGueltig(e.FinishedAt) || e.FinishedAt.Before(e.StartedAt)):
		return ungueltig("finished_at ungültig oder vor started_at")
	case e.ResponseStatus != 0 && (e.ResponseStatus < 100 || e.ResponseStatus > 599):
		return ungueltig("Response-Status %d außerhalb von 100 bis 599", e.ResponseStatus)
	case !headerGueltig(e.RequestHeaders) || !headerGueltig(e.ResponseHeaders):
		return ungueltig("Header sind kein gültiges UTF-8")
	}
	return nil
}

func eines[T comparable](v T, erlaubt ...T) bool {
	for _, x := range erlaubt {
		if v == x {
			return true
		}
	}
	return false
}

func (e Exchange) pruefeAufzaehlungen() error {
	bodies := []BodyState{BodyAbsent, BodyPresent, BodyDiscarded}
	redactions := []RedactionStatus{RedactionNone, RedactionApplied, RedactionUnsafe}
	switch {
	case !eines(e.RequestBody, bodies...) || !eines(e.ResponseBody, bodies...):
		return ungueltig("unbekannter Body-Zustand")
	case !eines(e.RequestRedact, redactions...) || !eines(e.ResponseRedact, redactions...):
		return ungueltig("unbekannter Redaction-Zustand")
	case !eines(e.Capture, CaptureComplete, CaptureIncomplete, CaptureFailed):
		return ungueltig("unbekannter Capture-Zustand %q", e.Capture)
	case !eines(e.Integrity, IntegrityVerified, IntegrityAbsent, IntegrityInvalid):
		return ungueltig("unbekannte Integrität %q", e.Integrity)
	case !eines(e.Equivalence, EquivalenceByte, EquivalenceSemantic, EquivalenceNone):
		return ungueltig("unbekannte Replay-Gleichwertigkeit %q", e.Equivalence)
	}
	return nil
}

func (e Exchange) pruefeBodies() error {
	if (e.RequestBody == BodyPresent) != (e.RequestBlob != "") {
		return ungueltig("Request-Body present und Blob passen nicht zusammen")
	}
	if (e.ResponseBody == BodyPresent) != (e.ResponseBlob != "") {
		return ungueltig("Response-Body present und Blob passen nicht zusammen")
	}
	return nil
}

func (e Exchange) pruefeZustand() error {
	switch {
	case e.MatchKey == "" && e.Capture != CaptureFailed:
		return ungueltig("match_key fehlt, aber capture_status ist %q", e.Capture)
	case (e.RequestBody == BodyDiscarded || e.ResponseBody == BodyDiscarded) && e.Capture == CaptureComplete:
		return ungueltig("verworfener Body bei capture_status complete")
	case e.RequestRedact == RedactionUnsafe && (e.Capture != CaptureFailed || e.RequestBody != BodyDiscarded):
		return ungueltig("redaction.request unsafe verlangt failed und verworfenen Request-Body")
	case e.ResponseRedact == RedactionUnsafe && (e.ResponseBody != BodyDiscarded || e.Equivalence != EquivalenceNone):
		return ungueltig("redaction.response unsafe verlangt verworfenen Body und replay_equivalence none")
	}
	return nil
}
