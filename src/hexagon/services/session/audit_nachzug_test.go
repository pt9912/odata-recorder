package session_test

import (
	"context"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
)

// TestAuditLeereKennung (V-5, Raster »leere Kennung«): Delete mit leerer id ist ODR-E7003 und schreibt nichts,
// auch wenn eine andere Session besteht.
func TestAuditLeereKennung(t *testing.T) {
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
	if _, err := u.del(context.Background(), ""); err == nil {
		t.Fatal("kein Fehler")
	}
	if n := u.audit.Aufrufe(); n != 0 {
		t.Errorf("%d Aufrufe des Audits", n)
	}
}
