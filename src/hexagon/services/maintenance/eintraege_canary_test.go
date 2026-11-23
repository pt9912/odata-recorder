package maintenance_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// kanarie ist ein wohlgeformter Wert (Buchstaben und Ziffern): ihn lässt die Form eines Attributs durch, so dass nur der Dienst ihn fernhält
// (SPEC-232 Nr. 2, *Ungeprüft*).
const kanarie = "KanarieQ7f"

// TestEintraegeOhneFehlertext hält E3 für die Wartung (SPEC-231 Nr. 8): Menge M-Quellen-Einträge, 4 Stellen (integritaet, aufbewahrung_uebersprungen,
// blob_fehlt, blob_beschaedigt), je mit dem Canary an der einen Quelle der Stelle (der Text eines Befunds bei integritaet, der Text des Fehlers
// von Delete, des ersten und des zweiten Verify) und dem Marker an den übrigen; jede Stelle löst ihren Eintrag aus, und weder Text noch Attribut
// trägt den Canary.
func TestEintraegeOhneFehlertext(t *testing.T) {
	namen := []string{"integritaet", "aufbewahrung_uebersprungen", "blob_fehlt", "blob_beschaedigt"}
	n := 0
	for i, name := range namen {
		n++
		wo := func(stelle int) string {
			if stelle == i {
				return kanarie
			}
			return marker
		}
		var u *welt
		if i == 0 {
			u = neu(t)
			u.speicher.befunde = []string{"Fremdschl" + string([]byte{0xc3, 0xbc}) + "ssel " + kanarie, "Index " + kanarie}
		} else {
			u = neu(t, mitMaxAge(30*tag))
			u.dienst(dienstA)
			u.sessions.fuege("q1", dienstA, t0().Add(-40*tag), model.SessionComplete)
			u.speicher.fest = []string{r("d"), r("f")}
			u.blobs.lege(r("f"))
			u.st.scheitere("Delete", 1, fmt.Errorf("%s: %w", wo(1), outbound.ErrReferenziert))
			u.st.scheitere("Verify", 1, fmt.Errorf("%s: %w", wo(2), outbound.ErrBlobFehlt))
			u.st.scheitere("Verify", 2, fmt.Errorf("%s: %w", wo(3), outbound.ErrBlobBeschaedigt))
		}
		_, _ = u.startup(context.Background())
		gefunden := 0
		for _, e := range u.log.alle() {
			if e.attrs["ereignis"] == name {
				gefunden++
			}
			alle := e.text
			for k, v := range e.attrs {
				alle += " " + k + "=" + v
			}
			if strings.Contains(alle, kanarie) {
				t.Errorf("%s: der Eintrag trägt den Canary: %s", name, alle)
			}
		}
		if gefunden != 1 {
			t.Errorf("%s: %d Einträge, erwartet 1", name, gefunden)
		}
	}
	if n != 4 {
		t.Errorf("%d Stellen, erwartet 4", n)
	}
}
