package workspace_test

import (
	"context"
	"testing"
)

// TestAuditLeereKennung (V-5, Raster »leere Kennung«): Delete mit leerer id ist ODR-E7001 und schreibt nichts,
// auch wenn ein anderer Workspace besteht.
func TestAuditLeereKennung(t *testing.T) {
	w := neu(t)
	legeWorkspaceAn(w, "ws-1")
	if _, err := w.v.Delete(context.Background(), ""); err == nil {
		t.Fatal("kein Fehler")
	}
	if n := w.audit.Aufrufe(); n != 0 {
		t.Errorf("%d Aufrufe des Audits", n)
	}
}
