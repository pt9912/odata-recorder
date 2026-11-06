package maintenance_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
)

// TestAuditMehrereDienste (E05, M-1 des Reviews): bei drei Diensten mit je zwei Sessions trägt der Eintrag jeder
// Löschung den Service seiner Session; die Folge der Dienste in der Ablage rotiert (3 Läufe), sodass kein Dienst
// immer an der ersten oder letzten Stelle steht und keine Kennung dem Wert einer Mutation gleicht. 18 Einträge.
func TestAuditMehrereDienste(t *testing.T) {
	namen := []string{"svc-a", "Svc.B", "svc-c"}
	n := 0
	for rot := range namen {
		u := neu(t, mitMaxAge(30*tag))
		u.dienst(append(append([]string{}, namen[rot:]...), namen[:rot]...)...)
		dienstVon := map[string]string{}
		for di, d := range namen {
			for k := range 2 {
				id := fmt.Sprintf("s%d%d", di, k)
				dienstVon[id] = d
				u.sessions.fuege(id, d, t0().Add(-time.Duration(40-(di*2+k))*tag), model.SessionComplete)
			}
		}
		if _, err := u.startup(context.Background()); err != nil {
			t.Fatal(err)
		}
		got := u.audit.Eintraege()
		if len(got) != 6 {
			t.Fatalf("Rotation %d: %d Einträge, erwartet 6", rot, len(got))
		}
		for _, e := range got {
			if want, da := dienstVon[e.SessionID]; !da || e != (support.Eintrag{Art: support.ArtSessionAufbewahrung, ServiceID: want, SessionID: e.SessionID}) {
				t.Errorf("Rotation %d: Eintrag %+v, Service der Session %q", rot, e, want)
			}
			n++
		}
	}
	if n != 18 {
		t.Errorf("Einträge %d, erwartet 18", n)
	}
}
