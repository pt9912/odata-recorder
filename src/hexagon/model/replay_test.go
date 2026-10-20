package model_test

import (
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func replayExchange(c model.CaptureStatus, r model.RedactionStatus, i model.BodyIntegrity) model.Exchange {
	e := gueltigerExchange()
	e.Capture, e.ResponseRedact, e.Integrity = c, r, i
	return e
}

// Die Erwartung ist je (Capture, Redaction, Integrität) für RedigierteBodies=false bzw. true
// ausgeschrieben (SPEC-117 Nr. 1); die Indizes zählen in der Reihenfolge der Slices cs, rs, is.
func TestReplayfaehigMatrix(t *testing.T) {
	cs := []model.CaptureStatus{model.CaptureComplete, model.CaptureIncomplete, model.CaptureFailed}
	rs := []model.RedactionStatus{model.RedactionNone, model.RedactionApplied, model.RedactionUnsafe}
	is := []model.BodyIntegrity{model.IntegrityVerified, model.IntegrityAbsent, model.IntegrityInvalid}
	erwartung := map[[3]int][2]bool{
		{0, 0, 0}: {true, true}, {0, 0, 1}: {true, true}, {0, 0, 2}: {false, false},
		{0, 1, 0}: {false, true}, {0, 1, 1}: {false, true}, {0, 1, 2}: {false, false},
		{0, 2, 0}: {false, false}, {0, 2, 1}: {false, false}, {0, 2, 2}: {false, false},
		{1, 0, 0}: {false, false}, {1, 0, 1}: {false, false}, {1, 0, 2}: {false, false},
		{1, 1, 0}: {false, false}, {1, 1, 1}: {false, false}, {1, 1, 2}: {false, false},
		{1, 2, 0}: {false, false}, {1, 2, 1}: {false, false}, {1, 2, 2}: {false, false},
		{2, 0, 0}: {false, false}, {2, 0, 1}: {false, false}, {2, 0, 2}: {false, false},
		{2, 1, 0}: {false, false}, {2, 1, 1}: {false, false}, {2, 1, 2}: {false, false},
		{2, 2, 0}: {false, false}, {2, 2, 1}: {false, false}, {2, 2, 2}: {false, false},
	}
	if len(erwartung) != 27 {
		t.Fatalf("%d Zeilen der Erwartung", len(erwartung))
	}
	n := 0
	for ci, c := range cs {
		for ri, r := range rs {
			for ii, i := range is {
				for bi, rb := range []bool{false, true} {
					want := erwartung[[3]int{ci, ri, ii}][bi]
					n++
					if got := model.Replayfaehig(replayExchange(c, r, i), rb); got != want {
						t.Errorf("%s %s %s redigierteBodies=%v: %v, erwartet %v", c, r, i, rb, got, want)
					}
				}
			}
		}
	}
	if n != 54 {
		t.Errorf("%d Kombinationen", n)
	}
}

func TestReplayfaehigUnbekannt(t *testing.T) {
	for _, rb := range []bool{false, true} {
		for name, e := range map[string]model.Exchange{
			"Capture leer":     replayExchange("", model.RedactionNone, model.IntegrityVerified),
			"Capture fremd":    replayExchange("kaputt", model.RedactionNone, model.IntegrityVerified),
			"Redaction leer":   replayExchange(model.CaptureComplete, "", model.IntegrityVerified),
			"Redaction fremd":  replayExchange(model.CaptureComplete, "kaputt", model.IntegrityVerified),
			"Integrität leer":  replayExchange(model.CaptureComplete, model.RedactionNone, ""),
			"Integrität fremd": replayExchange(model.CaptureComplete, model.RedactionNone, "kaputt"),
		} {
			if model.Replayfaehig(e, rb) {
				t.Errorf("%s (%v) ist replayfähig", name, rb)
			}
		}
	}
}

func TestReplayfaehigUnabhaengig(t *testing.T) {
	basis := replayExchange(model.CaptureComplete, model.RedactionApplied, model.IntegrityVerified)
	for _, rb := range []bool{false, true} {
		want := model.Replayfaehig(basis, rb)
		for name, f := range map[string]func(*model.Exchange){
			"Request-Redaction":  func(e *model.Exchange) { e.RequestRedact = model.RedactionUnsafe },
			"key_redacted":       func(e *model.Exchange) { e.KeyRedacted = !e.KeyRedacted },
			"Request-Header":     func(e *model.Exchange) { e.RequestHeaders = map[string][]string{"A": {"b"}} },
			"Response-Header":    func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{"A": {"[REDACTED]"}} },
			"Request-Body":       func(e *model.Exchange) { e.RequestBody = model.BodyDiscarded },
			"Response-Body":      func(e *model.Exchange) { e.ResponseBody = model.BodyDiscarded },
			"replay_equivalence": func(e *model.Exchange) { e.Equivalence = model.EquivalenceNone },
			"match_key":          func(e *model.Exchange) { e.MatchKey = "" },
			"Zeiten":             func(e *model.Exchange) { e.StartedAt = e.StartedAt.Add(-time.Hour); e.FinishedAt = time.Time{} },
			"Kennungen":          func(e *model.Exchange) { e.ID, e.SessionID, e.SequenceNo = "", "", 0 },
			"Status":             func(e *model.Exchange) { e.ResponseStatus = 500 },
		} {
			e := basis
			f(&e)
			if got := model.Replayfaehig(e, rb); got != want {
				t.Errorf("%s ändert das Ergebnis (%v): %v", name, rb, got)
			}
		}
	}
}

func TestReplayfaehigNullwert(t *testing.T) {
	if model.Replayfaehig(model.Exchange{}, true) || model.Replayfaehig(model.Exchange{}, false) {
		t.Error("Nullwert ist replayfähig")
	}
	e := replayExchange(model.CaptureComplete, model.RedactionNone, model.IntegrityVerified)
	erst := model.Replayfaehig(e, false)
	if zweit := model.Replayfaehig(e, false); !erst || erst != zweit {
		t.Error("nicht deterministisch")
	}
}
