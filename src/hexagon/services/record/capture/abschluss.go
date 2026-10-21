package capture

import (
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// Zustand ist der Zustand eines Austauschs aus den Ergebnissen seiner beiden Bodys
// (SPEC-119 Nr. 12); die Felder tragen die Namen der Felder von model.Exchange.
type Zustand struct {
	Capture        model.CaptureStatus
	KeyMoeglich    bool
	Integrity      model.BodyIntegrity
	Equivalence    model.ReplayEquivalence
	RequestBody    model.BodyState
	RequestBlob    string
	RequestRedact  model.RedactionStatus
	ResponseBody   model.BodyState
	ResponseBlob   string
	ResponseRedact model.RedactionStatus
}

// Abschluss bildet den Zustand des Austauschs; die Funktion ist rein. Sie prüft von den
// Ergebnissen nur die Richtung, ob der Exchange gültig ist, entscheidet model.Exchange.Validate.
func Abschluss(request, response Ergebnis) (Zustand, error) {
	if request.Richtung != Request || response.Richtung != Response {
		return Zustand{}, errors.Join(model.ErrUngueltig, errors.New("Abschluss verlangt ein Ergebnis je Richtung, Request zuerst"))
	}
	return Zustand{
		Capture:        captureStatus(request, response),
		KeyMoeglich:    request.Ausgang == Vollstaendig,
		Integrity:      integritaet(request, response),
		Equivalence:    aequivalenz(request, response),
		RequestBody:    request.Body,
		RequestBlob:    request.Blob,
		RequestRedact:  request.Redaction,
		ResponseBody:   response.Body,
		ResponseBlob:   response.Blob,
		ResponseRedact: response.Redaction,
	}, nil
}

func captureStatus(request, response Ergebnis) model.CaptureStatus {
	switch {
	case request.Ausgang != Vollstaendig || response.Ausgang == Verstoss:
		return model.CaptureFailed
	case response.Ausgang == Abbruch:
		return model.CaptureIncomplete
	}
	return model.CaptureComplete
}

func integritaet(request, response Ergebnis) model.BodyIntegrity {
	if request.Body == model.BodyPresent || response.Body == model.BodyPresent {
		return model.IntegrityVerified
	}
	return model.IntegrityAbsent
}

// aequivalenz ändert die Redaction des Request-Bodys nicht (SPEC-045).
func aequivalenz(request, response Ergebnis) model.ReplayEquivalence {
	switch {
	case request.Body == model.BodyDiscarded || response.Body == model.BodyDiscarded:
		return model.EquivalenceNone
	case response.Redaction == model.RedactionApplied:
		return model.EquivalenceSemantic
	}
	return model.EquivalenceByte
}
