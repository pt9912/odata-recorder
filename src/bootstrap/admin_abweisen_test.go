package bootstrap_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestAdminAbweisenBegrenzt hält W17 (SPEC-301 Nr. 1, 8, SPEC-300 Nr. 5 d, 12 b): am Admin-Listener eines hochgefahrenen
// Handle sind 21 Anfragen ohne Token in kurzer Folge je 401; audit.jsonl trägt 20 Zeilen admin_abgewiesen mit den vier
// Schlüsseln in ihrer Folge und den IDs der ersten 20 Antworten, und die Diagnose zählt ODR-E2101 21-mal; je Modus.
func TestAdminAbweisenBegrenzt(t *testing.T) {
	zellen := 0
	for _, modus := range vierModi() {
		zellen++
		k := listenerKonfig(t.TempDir(), modus)
		url := "http://" + adminHoch(t, k).AdminAdresse()
		var ids []string
		for i := range 21 {
			status, body, kopf := sende(neuerAdminClient(), http.MethodGet, url+"/__admin/sessions", "", "")
			if status != 401 || !strings.Contains(body, `"code":"ODR-E2101"`) {
				t.Fatalf("%s Anfrage %d: %d %s", modus, i+1, status, body)
			}
			ids = append(ids, kopf.Get("X-Correlation-Id"))
		}
		inhalt, err := os.ReadFile(filepath.Join(k.DataDir, "audit.jsonl"))
		if err != nil {
			t.Fatalf("%s: %v", modus, err)
		}
		var geschrieben []string
		for _, zeile := range strings.Split(strings.TrimSuffix(string(inhalt), "\n"), "\n") {
			schluessel, werte := auditZeile(t, zeile)
			if !slices.Equal(schluessel, []string{"timestamp", "correlation_id", "art", "error_code"}) ||
				werte["art"] != "admin_abgewiesen" || werte["error_code"] != "ODR-E2101" {
				t.Errorf("%s: Zeile %s", modus, zeile)
			}
			geschrieben = append(geschrieben, werte["correlation_id"])
		}
		if !slices.Equal(geschrieben, ids[:20]) {
			t.Errorf("%s: %d Zeilen, erwartet 20 mit den IDs der ersten 20 Antworten", modus, len(geschrieben))
		}
		status, body, _ := sende(neuerAdminClient(), http.MethodGet, url+"/__admin/diagnostics", bearer(t, k.AdminTokenFile), "")
		var d diagnose
		if err := json.Unmarshal([]byte(body), &d); status != 200 || err != nil {
			t.Fatalf("%s: Diagnose %d %s (%v)", modus, status, body, err)
		}
		anzahl := 0
		for _, e := range d.Errors {
			if e.Code == "ODR-E2101" {
				anzahl = e.Count
			}
		}
		if anzahl != 21 {
			t.Errorf("%s: Zähler ODR-E2101 %d, erwartet 21", modus, anzahl)
		}
	}
	if zellen != 4 {
		t.Errorf("%d Zellen, erwartet 4", zellen)
	}
}
