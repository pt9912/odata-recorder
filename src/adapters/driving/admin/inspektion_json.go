package admin

import (
	"encoding/json"
	"errors"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// exchangeZeileJSON ist die Zeile der Exchanges: die 15 Felder der Allowlist von SPEC-236 Nr. 5 in der Folge von SPEC-293
// Nr. 6; response_status gleich 0 ist `null`, die Zustände stehen als Text unverändert.
type exchangeZeileJSON struct {
	ID                      string   `json:"id"`
	SequenceNo              int      `json:"sequence_no"`
	Method                  string   `json:"method"`
	Path                    string   `json:"path"`
	StartedAt               zeitJSON `json:"started_at"`
	FinishedAt              zeitJSON `json:"finished_at"`
	KeyRedacted             bool     `json:"key_redacted"`
	RequestBodyState        string   `json:"request_body_state"`
	ResponseBodyState       string   `json:"response_body_state"`
	ResponseStatus          *int     `json:"response_status"`
	CaptureStatus           string   `json:"capture_status"`
	BodyIntegrity           string   `json:"body_integrity"`
	ReplayEquivalence       string   `json:"replay_equivalence"`
	RequestRedactionStatus  string   `json:"request_redaction_status"`
	ResponseRedactionStatus string   `json:"response_redaction_status"`
}

// exchangeSeiteJSON ist der Body von `GET /__admin/sessions/{id}/exchanges`; next ist `null`, wenn der Dienst Weiter gleich 0
// liefert.
type exchangeSeiteJSON struct {
	Exchanges []exchangeZeileJSON `json:"exchanges"`
	Limit     int                 `json:"limit"`
	Total     int                 `json:"total"`
	Next      *int                `json:"next"`
}

// zahlOderNull liefert nil für 0, sonst die Zahl (SPEC-293 Nr. 6).
func zahlOderNull(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

// exchangeSeiteAus bildet die Seite des Dienstes ab, ohne ein Feld zu prüfen; die Liste ist nie `null`.
func exchangeSeiteAus(s inbound.ExchangeSeite) exchangeSeiteJSON {
	zeilen := make([]exchangeZeileJSON, len(s.Eintraege))
	for i, z := range s.Eintraege {
		zeilen[i] = exchangeZeileJSON{
			ID: z.ID, SequenceNo: z.SequenceNo, Method: z.Method, Path: z.Path,
			StartedAt: zeitJSON(z.StartedAt), FinishedAt: zeitJSON(z.FinishedAt), KeyRedacted: z.KeyRedacted,
			RequestBodyState: string(z.RequestBody), ResponseBodyState: string(z.ResponseBody),
			ResponseStatus: zahlOderNull(z.ResponseStatus), CaptureStatus: string(z.Capture),
			BodyIntegrity: string(z.Integrity), ReplayEquivalence: string(z.Equivalence),
			RequestRedactionStatus: string(z.RequestRedact), ResponseRedactionStatus: string(z.ResponseRedact),
		}
	}
	return exchangeSeiteJSON{Exchanges: zeilen, Limit: s.Limit, Total: s.Gesamt, Next: zahlOderNull(s.Weiter)}
}

// fehlerZaehlerJSON ist ein Zähler der Diagnose; count behält den ganzen Bereich von uint64.
type fehlerZaehlerJSON struct {
	Code  string `json:"code"`
	Count uint64 `json:"count"`
}

// diagnoseJSON ist der Body von `GET /__admin/diagnostics`; settings ist nie `null`, errors nie `null`.
type diagnoseJSON struct {
	Mode     string              `json:"mode"`
	Settings map[string]string   `json:"settings"`
	Errors   []fehlerZaehlerJSON `json:"errors"`
}

// diagnoseAus bildet die Diagnose des Dienstes ab, ohne einen Schlüssel zu filtern (SPEC-293 Nr. 8): encoding/json
// schreibt die Schlüssel von settings aufsteigend nach den Bytes.
func diagnoseAus(d inbound.Diagnose) diagnoseJSON {
	einstellungen := make(map[string]string, len(d.Einstellungen))
	for k, v := range d.Einstellungen {
		einstellungen[k] = v
	}
	fehler := make([]fehlerZaehlerJSON, len(d.Fehler))
	for i, f := range d.Fehler {
		fehler[i] = fehlerZaehlerJSON{Code: f.Code.String(), Count: f.Anzahl}
	}
	return diagnoseJSON{Mode: d.Modus, Settings: einstellungen, Errors: fehler}
}

// errEvidenzUngueltig ist der Fehler einer Evidenz, die kein gültiges JSON ist; der Text nennt keinen Wert. Er trägt keinen
// Code: die Antwort ist ODR-E9001 (SPEC-293 Nr. 7).
func errEvidenzUngueltig() error {
	return errors.New("admin: Evidenz der Coverage ist kein gültiges JSON")
}

// schreibeCoverage schreibt den Body von `GET /__admin/coverage`: `{"entries":[…]}` mit `entity_set`, `status` und `evidence`
// je Eintrag. Die Evidenz steht mit denselben Bytes im Body, ohne Umformatieren und ohne Maskierung (SPEC-293 Nr. 7);
// encoding/json maskiert `<`, `>`, `&` und U+2028 auch in einem eingesetzten Wert, deshalb setzt die Funktion die Bytes
// selbst ein. Ist eine Evidenz kein gültiges JSON, ist das Ergebnis der Fehler und kein Teil des Bodys.
func schreibeCoverage(es []inbound.CoverageEintrag) ([]byte, error) {
	body := []byte(`{"entries":[`)
	for i, e := range es {
		if !json.Valid([]byte(e.Evidenz)) {
			return nil, errEvidenzUngueltig()
		}
		if i > 0 {
			body = append(body, ',')
		}
		body = append(body, `{"entity_set":`...)
		body = append(body, schreibeJSON(e.EntitySet)...)
		body = append(body, `,"status":`...)
		body = append(body, schreibeJSON(string(e.Status))...)
		body = append(body, `,"evidence":`...)
		body = append(body, e.Evidenz...)
		body = append(body, '}')
	}
	return append(body, `]}`...), nil
}
