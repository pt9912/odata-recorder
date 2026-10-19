package model_test

import (
	"errors"
	"strings"
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
		"Jahr 10000":   func(s *model.Session) { s.CreatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
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
		"ohne Session":                   func(e *model.Exchange) { e.SessionID = "" },
		"sequence_no 0":                  func(e *model.Exchange) { e.SequenceNo = 0 },
		"ohne Pfad":                      func(e *model.Exchange) { e.Path = "" },
		"unbekannter Request-Body":       func(e *model.Exchange) { e.RequestBody = "x" },
		"unbekannte Redaction":           func(e *model.Exchange) { e.ResponseRedact = "x" },
		"unbekannte Integrität":          func(e *model.Exchange) { e.Integrity = "x" },
		"unbekannte Gleichwertigkeit":    func(e *model.Exchange) { e.Equivalence = "x" },
		"present ohne Blob":              func(e *model.Exchange) { e.ResponseBlob = "" },
		"Request-Blob ohne present":      func(e *model.Exchange) { e.RequestBlob = "sha256:x" },
		"match_key leer bei incomplete":  func(e *model.Exchange) { e.MatchKey, e.Capture = "", model.CaptureIncomplete },
		"verworfen bei complete":         func(e *model.Exchange) { e.RequestBody = model.BodyDiscarded },
		"Startzeit nach Jahr 9999":       func(e *model.Exchange) { e.StartedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		"finished_at vor started_at":     func(e *model.Exchange) { e.FinishedAt = e.StartedAt.Add(-time.Second) },
		"Status 99":                      func(e *model.Exchange) { e.ResponseStatus = 99 },
		"Status 600":                     func(e *model.Exchange) { e.ResponseStatus = 600 },
		"Request-Header kein UTF-8":      func(e *model.Exchange) { e.RequestHeaders = map[string][]string{"X": {"\xff"}} },
		"Response-Headername kein UTF-8": func(e *model.Exchange) { e.ResponseHeaders = map[string][]string{"\xff": {"a"}} },
		"unbekannter Response-Body":      func(e *model.Exchange) { e.ResponseBody, e.ResponseBlob = "x", "" },
		"unbekannte Request-Redaction":   func(e *model.Exchange) { e.RequestRedact = "x" },
		"Request-Body present ohne Blob": func(e *model.Exchange) { e.RequestBody = model.BodyPresent },
		"Response-Blob ohne present":     func(e *model.Exchange) { e.ResponseBody = model.BodyAbsent },
		"finished_at nach Jahr 9999":     func(e *model.Exchange) { e.FinishedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		"UTC-Jahr 10000 in Zone": func(e *model.Exchange) {
			e.StartedAt = time.Date(9999, 12, 31, 23, 0, 0, 0, time.FixedZone("m2", -2*3600))
		},
		"unsafe Request ohne failed, Body verworfen": func(e *model.Exchange) {
			e.RequestBody, e.RequestRedact, e.Capture = model.BodyDiscarded, model.RedactionUnsafe, model.CaptureIncomplete
		},
		"unsafe Request mit Body": func(e *model.Exchange) {
			e.MatchKey, e.Capture, e.RequestRedact = "", model.CaptureFailed, model.RedactionUnsafe
		},
		"unsafe Request ohne failed": func(e *model.Exchange) { e.RequestRedact = model.RedactionUnsafe },
		"unsafe Response mit Body":   func(e *model.Exchange) { e.ResponseRedact = model.RedactionUnsafe },
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
	for _, status := range []int{0, 100, 599} {
		e := gueltigerExchange()
		e.ResponseStatus = status
		if err := e.Validate(); err != nil {
			t.Errorf("Status %d: %v", status, err)
		}
	}
	zone := gueltigerExchange()
	zone.StartedAt = time.Date(10000, 1, 1, 1, 0, 0, 0, time.FixedZone("p2", 2*3600))
	if err := zone.Validate(); err != nil {
		t.Errorf("UTC-Jahr 9999 in Zone: %v", err)
	}
	erlaubt := gueltigerExchange()
	erlaubt.MatchKey, erlaubt.Capture = "", model.CaptureFailed
	erlaubt.RequestRedact, erlaubt.RequestBody = model.RedactionUnsafe, model.BodyDiscarded
	if err := erlaubt.Validate(); err != nil {
		t.Errorf("failed ohne match_key mit unsafe Request: %v", err)
	}
}

func TestBlobReferenzGueltig(t *testing.T) {
	hex := strings.Repeat("0123456789abcdef", 4)
	for ref, soll := range map[string]bool{
		"sha256:" + hex:                  true,
		"sha256:" + hex[:63]:             false,
		"sha256:" + hex + "0":            false,
		"sha256:" + strings.ToUpper(hex): false,
		"sha256:" + hex[:63] + "g":       false,
		"abcdefg" + hex:                  false,
		"sha1:" + hex:                    false,
		hex:                              false,
		"":                               false,
		"sha256:../" + hex[:61]:          false,
	} {
		if got := model.BlobReferenzGueltig(ref); got != soll {
			t.Errorf("%q: %v, erwartet %v", ref, got, soll)
		}
	}
}

func gueltigePublikation() model.Publication {
	blob := "sha256:" + strings.Repeat("a", 64)
	return model.Publication{
		Import:           model.Import{ID: "i", RequestHash: "r", BaselineSnapshotID: "b", ServiceID: "d", CreatedAt: jetzt(), ReportJSON: "{}"},
		Snapshot:         model.BaselineSnapshot{ID: "b", ServiceID: "d", CreatedAt: jetzt(), StateBlob: blob, StateHash: "h", MetadataHash: "m"},
		SourceSessionIDs: []string{"s"},
		Coverage:         []model.Coverage{{EntitySet: "P", Status: model.CoveragePartial, EvidenceJSON: "null"}},
	}
}

func TestPublicationValidate(t *testing.T) {
	if err := gueltigePublikation().Validate(); err != nil {
		t.Fatal(err)
	}
	ohne := gueltigePublikation()
	ohne.Coverage = nil
	if err := ohne.Validate(); err != nil {
		t.Fatalf("leere Coverage: %v", err)
	}
	for name, mut := range map[string]func(*model.Publication){
		"keine Quelle":         func(p *model.Publication) { p.SourceSessionIDs = nil },
		"doppelte Quelle":      func(p *model.Publication) { p.SourceSessionIDs = []string{"s", "s"} },
		"Import ohne Zeit":     func(p *model.Publication) { p.Import.CreatedAt = time.Time{} },
		"andere Baseline":      func(p *model.Publication) { p.Import.BaselineSnapshotID = "x" },
		"anderer Service":      func(p *model.Publication) { p.Import.ServiceID = "x" },
		"Evidenz kein JSON":    func(p *model.Publication) { p.Coverage[0].EvidenceJSON = "{" },
		"unbekannter Status":   func(p *model.Publication) { p.Coverage[0].Status = "voll" },
		"doppeltes entity_set": func(p *model.Publication) { p.Coverage = append(p.Coverage, p.Coverage[0]) },
	} {
		p := gueltigePublikation()
		mut(&p)
		if err := p.Validate(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestServiceValidate(t *testing.T) {
	blob := "sha256:" + strings.Repeat("a", 64)
	gut := model.Service{ID: "d", ServiceRoot: "/", ODataVersion: "4.0"}
	if err := gut.Validate(); err != nil {
		t.Fatal(err)
	}
	gut.MetadataBlob, gut.MetadataHash = blob, "h"
	if err := gut.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, ungueltig := range []model.Service{{}, {ID: "d"}, {ID: "d", ServiceRoot: "/"}, {ID: "d", ServiceRoot: "/", ODataVersion: "4", MetadataBlob: blob}} {
		if err := ungueltig.Validate(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%+v: %v", ungueltig, err)
		}
	}
}

func TestBaselineSnapshotValidate(t *testing.T) {
	gut := gueltigePublikation().Snapshot
	if err := gut.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*model.BaselineSnapshot){
		"ohne Kennung": func(b *model.BaselineSnapshot) { b.ID = "" },
		"ohne Service": func(b *model.BaselineSnapshot) { b.ServiceID = "" },
	} {
		b := gut
		mut(&b)
		if err := b.Validate(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestImportValidate(t *testing.T) {
	gut := gueltigePublikation().Import
	if err := gut.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*model.Import){
		"ohne Kennung":      func(i *model.Import) { i.ID = "" },
		"ohne Service":      func(i *model.Import) { i.ServiceID = "" },
		"ohne Baseline":     func(i *model.Import) { i.BaselineSnapshotID = "" },
		"ohne request_hash": func(i *model.Import) { i.RequestHash = "" },
		"JSON kein UTF-8":   func(i *model.Import) { i.ReportJSON = "\"\xff\"" },
		"ohne Zeit":         func(i *model.Import) { i.CreatedAt = time.Time{} },
	} {
		i := gut
		mut(&i)
		if err := i.Validate(); !errors.Is(err, model.ErrUngueltig) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
