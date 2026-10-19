package model_test

import (
	"errors"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

func jetzt() time.Time { return time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC) }

func gueltigerExchange() model.Exchange {
	return model.Exchange{
		ID: "e", SessionID: "s", SequenceNo: 1, Method: "GET", Path: "/P", MatchKey: "sha256:a",
		StartedAt: jetzt(), RequestBody: model.BodyAbsent, ResponseBody: model.BodyPresent, ResponseBlob: "sha256:b",
		Capture: model.CaptureComplete, RequestRedact: model.RedactionNone, ResponseRedact: model.RedactionNone,
		Integrity: model.IntegrityVerified, Equivalence: model.EquivalenceByte,
	}
}

func TestSessionStatusWechsel(t *testing.T) {
	alle := []model.SessionStatus{model.SessionOpen, model.SessionComplete, model.SessionIncomplete}
	for _, von := range alle {
		for _, nach := range alle {
			erlaubt := von == model.SessionOpen && nach != model.SessionOpen
			if got := von.KannWechselnNach(nach); got != erlaubt {
				t.Errorf("%s → %s: %v, erwartet %v", von, nach, got, erlaubt)
			}
		}
	}
	if model.SessionStatus("kaputt").Gueltig() || !model.SessionOpen.Gueltig() {
		t.Error("Gueltig stimmt nicht")
	}
}

func TestSessionValidate(t *testing.T) {
	gut := model.Session{ID: "s", ServiceID: "d", CreatedAt: jetzt(), Status: model.SessionOpen}
	if err := gut.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*model.Session){
		"ohne Kennung": func(s *model.Session) { s.ID = "" },
		"ohne Service": func(s *model.Session) { s.ServiceID = "" },
		"ohne Zeit":    func(s *model.Session) { s.CreatedAt = time.Time{} },
		"nicht open":   func(s *model.Session) { s.Status = model.SessionComplete },
	} {
		s := gut
		mut(&s)
		if err := s.Validate(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestExchangeValidate(t *testing.T) {
	if err := gueltigerExchange().Validate(); err != nil {
		t.Fatal(err)
	}
	fehlerhaft := map[string]func(*model.Exchange){
		"ohne Session":                  func(e *model.Exchange) { e.SessionID = "" },
		"sequence_no 0":                 func(e *model.Exchange) { e.SequenceNo = 0 },
		"ohne Pfad":                     func(e *model.Exchange) { e.Path = "" },
		"unbekannter Request-Body":      func(e *model.Exchange) { e.RequestBody = "x" },
		"unbekannte Redaction":          func(e *model.Exchange) { e.ResponseRedact = "x" },
		"unbekannte Integrität":         func(e *model.Exchange) { e.Integrity = "x" },
		"unbekannte Gleichwertigkeit":   func(e *model.Exchange) { e.Equivalence = "x" },
		"present ohne Blob":             func(e *model.Exchange) { e.ResponseBlob = "" },
		"Request-Blob ohne present":     func(e *model.Exchange) { e.RequestBlob = "sha256:x" },
		"match_key leer bei incomplete": func(e *model.Exchange) { e.MatchKey, e.Capture = "", model.CaptureIncomplete },
		"verworfen bei complete":        func(e *model.Exchange) { e.RequestBody = model.BodyDiscarded },
		"unsafe Request ohne failed":    func(e *model.Exchange) { e.RequestRedact = model.RedactionUnsafe },
		"unsafe Response mit Body":      func(e *model.Exchange) { e.ResponseRedact = model.RedactionUnsafe },
		"unsafe Response ohne none": func(e *model.Exchange) {
			e.ResponseRedact, e.ResponseBody, e.ResponseBlob, e.Capture = model.RedactionUnsafe, model.BodyDiscarded, "", model.CaptureIncomplete
		},
	}
	for name, mut := range fehlerhaft {
		e := gueltigerExchange()
		mut(&e)
		if err := e.Validate(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
	erlaubt := gueltigerExchange()
	erlaubt.MatchKey, erlaubt.Capture = "", model.CaptureFailed
	erlaubt.RequestRedact, erlaubt.RequestBody = model.RedactionUnsafe, model.BodyDiscarded
	if err := erlaubt.Validate(); err != nil {
		t.Errorf("failed ohne match_key mit unsafe Request: %v", err)
	}
}
