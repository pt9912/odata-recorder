package session_test

import (
	"context"
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// TestAuditStatusAchse (E05, M-2 des Reviews): jede nicht offene Session wird gelöscht und schreibt genau einen
// Eintrag, gleich in welchem Status sie steht: complete, incomplete (die Folge jedes Abbruchs) und ein Status
// außerhalb der Menge (SPEC-124 Nr. 7 c prüft allein auf open); eine offene Session schreibt keinen.
func TestAuditStatusAchse(t *testing.T) {
	n := 0
	for _, st := range []model.SessionStatus{model.SessionComplete, model.SessionIncomplete, "zz"} {
		u := neu(t)
		u.sessions.fuege("s1", dienstA, t0(), st)
		if _, err := u.del(context.Background(), "s1"); err != nil {
			t.Fatalf("Status %q: %v", st, err)
		}
		if got := u.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{geloescht(dienstA, "s1")}) {
			t.Errorf("Status %q: Einträge %+v", st, got)
		}
		n++
	}
	if n != 3 {
		t.Errorf("Status %d, erwartet 3", n)
	}
	u := neu(t)
	u.sessions.fuege("s1", dienstA, t0(), model.SessionOpen)
	if _, err := u.del(context.Background(), "s1"); err == nil || u.audit.Aufrufe() != 0 {
		t.Errorf("offen: %v, Aufrufe %d", err, u.audit.Aufrufe())
	}
}
