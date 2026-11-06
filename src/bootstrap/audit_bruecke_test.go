package bootstrap_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// TestAuditEineInstanz (SPEC-139 Nr. 3, SPEC-234 Nr. 8): alle Dienste, die der Modus baut, erhalten dieselbe
// Instanz des Audits. Im Modus record vier (Wartung, Session, Import, Workspace), im Modus replay zwei
// (Wartung, Workspace).
func TestAuditEineInstanz(t *testing.T) {
	for _, modus := range []string{modusRecord, modusReplay} {
		h := neuHaken()
		baue(context.Background(), t, konfigMitWerten(t.TempDir(), modus), h)
		quelle := h.wartung.Audit
		if quelle == nil {
			t.Fatalf("%s: Wartung ohne Audit", modus)
		}
		trager := map[string]support.AuditLog{"Workspace": nil}
		if h.workspace != nil {
			trager["Workspace"] = h.workspace.Audit
		}
		if modus == modusRecord {
			if h.session == nil || h.importKfg == nil {
				t.Fatalf("%s: Session oder Import nicht gebaut", modus)
			}
			trager["Session"], trager["Import"] = h.session.Audit, h.importKfg.Audit
		}
		want := 1
		if modus == modusRecord {
			want = 3
		}
		if len(trager) != want {
			t.Errorf("%s: %d Träger neben der Wartung, erwartet %d", modus, len(trager), want)
		}
		for name, a := range trager {
			if a == nil || a != quelle {
				t.Errorf("%s: %s trägt nicht die Instanz der Wartung", modus, name)
			}
		}
	}
}

// TestAuditPfadUndZeile (SPEC-139 Nr. 2, SPEC-235 Nr. 2 und 4): die Instanz schreibt in <DATA_DIR>/audit.jsonl und
// nirgends sonst, `Build` legt die Datei nicht an, und die Zeile trägt Zeit und Korrelations-ID, die der Adapter
// vergibt. Beide Modi.
func TestAuditPfadUndZeile(t *testing.T) {
	for _, modus := range []string{modusRecord, modusReplay} {
		dir := t.TempDir()
		h := neuHaken()
		baue(context.Background(), t, konfigMitWerten(dir, modus), h)
		pfad := filepath.Join(dir, "audit.jsonl")
		if _, err := os.Stat(pfad); err == nil {
			t.Fatalf("%s: Build legte audit.jsonl an", modus)
		}
		vorher := time.Now().UTC().Add(-time.Second)
		e := support.Eintrag{Art: support.ArtSessionAufbewahrung, ServiceID: "x1", SessionID: "s-1"}
		if err := h.wartung.Audit.Append(e); err != nil {
			t.Fatalf("%s: Append: %v", modus, err)
		}
		nachher := time.Now().UTC().Add(time.Second)
		b, err := os.ReadFile(pfad)
		if err != nil {
			t.Fatalf("%s: %v", modus, err)
		}
		z := string(b)
		for _, teil := range []string{`"art":"session_aufbewahrung"`, `"service_id":"x1"`, `"session_id":"s-1"`, `"correlation_id":"`} {
			if !strings.Contains(z, teil) {
				t.Errorf("%s: Zeile %q ohne %s", modus, z, teil)
			}
		}
		if strings.Contains(z, `"correlation_id":""`) {
			t.Errorf("%s: Zeile ohne vergebene Korrelations-ID: %q", modus, z)
		}
		i := strings.Index(z, `"timestamp":"`)
		if i < 0 || len(z) < i+13+20 {
			t.Fatalf("%s: Zeile ohne Zeit: %q", modus, z)
		}
		ts, err := time.Parse(time.RFC3339Nano, z[i+13:i+13+strings.Index(z[i+13:], `"`)])
		if err != nil || ts.Before(vorher) || ts.After(nachher) {
			t.Errorf("%s: Zeit %v (%v) außerhalb von %v bis %v", modus, ts, err, vorher, nachher)
		}
		ents, _ := os.ReadDir(dir)
		for _, en := range ents {
			if strings.HasSuffix(en.Name(), ".jsonl") && en.Name() != "audit.jsonl" {
				t.Errorf("%s: weitere Datei %s", modus, en.Name())
			}
		}
	}
}
