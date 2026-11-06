package baselineimport_test

import (
	"slices"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/support"
	"github.com/pt9912/odata-recorder/src/hexagon/services/baselineimport"
)

// TestAuditMehrereDienste (E05, M-1 des Reviews): bei drei konfigurierten Diensten trägt der Eintrag des
// fehlgeschlagenen Imports den Service der Anfrage, gleich an welcher Stelle der Konfiguration er steht: drei
// Dienste × drei Folgen der Konfiguration (Rotation) = 9 Läufe.
func TestAuditMehrereDienste(t *testing.T) {
	namen := []string{"cat-a", "Cat.B", "cat-c"}
	n := 0
	for rot := range namen {
		var konf []baselineimport.Dienst
		for _, d := range append(append([]string{}, namen[rot:]...), namen[:rot]...) {
			konf = append(konf, baselineimport.Dienst{ID: d, UpstreamURL: upstream})
		}
		for _, svc := range namen {
			w := neu(t)
			w.sessionR()
			w.konfig.Dienste = konf
			w.sessions.m[sitzungID] = model.Session{ID: sitzungID, ServiceID: svc, CreatedAt: t0(), Status: model.SessionOpen}
			a := anfrage()
			a.ServiceID = svc
			if _, err := w.importiere(a); err == nil {
				t.Fatalf("Rotation %d, Service %q: kein Fehler", rot, svc)
			}
			if got := w.audit.Eintraege(); !slices.Equal(got, []support.Eintrag{fehlgeschlagen(svc, "imp-1", model.CodeE7009)}) {
				t.Errorf("Rotation %d, Service %q: Einträge %+v", rot, svc, got)
			}
			n++
		}
	}
	if n != 9 {
		t.Errorf("Läufe %d, erwartet 9", n)
	}
}
