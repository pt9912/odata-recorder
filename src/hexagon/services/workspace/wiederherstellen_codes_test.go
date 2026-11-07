package workspace_test

import (
	"context"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestCodesSnapshot (T20): die vier Codes von Restore und Snapshot als Literale aus dem Vertrag (SPEC-025, SPEC-065,
// SPEC-094, SPEC-096): Code, Status, Ziel und Text, 4 × 4 = 16 Vergleiche gegen das Modell; Restore liefert die
// Status von ODR-E7002, ODR-E5004 und ODR-E7008.
func TestCodesSnapshot(t *testing.T) {
	soll := []struct {
		code, ziel, text string
		status           int
	}{
		{"ODR-E7002", "workspace", "Snapshot unknown", 404},
		{"ODR-E5004", "replay", "Assignment belongs to another service", 409},
		{"ODR-E7006", "workspace", "Snapshot name already taken", 409},
		{"ODR-E7008", "workspace", "Snapshot kind does not fit the operation", 422},
	}
	n := 0
	for _, s := range soll {
		c, ok := model.ParseCode(s.code)
		if !ok {
			t.Fatalf("%s fehlt im Modell", s.code)
		}
		if c.String() != s.code {
			t.Errorf("%s: Code %s", s.code, c.String())
		}
		if st := c.Stati(); len(st) != 1 || st[0] != s.status {
			t.Errorf("%s: Status %v, erwartet %d", s.code, st, s.status)
		}
		if string(c.Zielbereich()) != s.ziel {
			t.Errorf("%s: Ziel %s, erwartet %s", s.code, c.Zielbereich(), s.ziel)
		}
		if c.Message() != s.text {
			t.Errorf("%s: Text %q, erwartet %q", s.code, c.Message(), s.text)
		}
		n += 4
	}
	if n != 16 {
		t.Errorf("%d Vergleiche", n)
	}
	w := restoreWelt(t)
	w.baseline("bsl-x", "svc2", standardEnts())
	for id, s := range map[string]int{"snp-gibtsnicht": 0, "bsl-x": 1, "bsl-1": 3} {
		_, err := w.v.Restore(context.Background(), "ws-1", id)
		f, ok := codeVon(err)
		if !ok || f.Code.String() != soll[s].code || f.Status != soll[s].status {
			t.Errorf("%s: %v, erwartet %s %d", id, err, soll[s].code, soll[s].status)
		}
	}
}
