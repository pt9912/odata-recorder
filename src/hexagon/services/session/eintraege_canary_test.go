package session_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
)

// kanarie ist ein wohlgeformter Wert (Buchstaben und Ziffern): ihn lässt die Form eines Attributs durch, so dass nur der Dienst ihn fernhält
// (SPEC-232 Nr. 2, *Ungeprüft*).
const kanarie = "KanarieQ7f"

// TestEintraegeOhneFehlertext hält E3 für den Session-Dienst (SPEC-231 Nr. 8): Menge M-Quellen-Einträge, 3 Stellen (bereinigung_fehlgeschlagen,
// bereinigung_zurueckgestellt, beenden_fehlgeschlagen), je mit dem Text des Fehlers des Ports als Canary in beiden Formen (ein Fehler mit
// Code und ein Fehler ohne); jede Stelle löst ihren Eintrag aus, und weder Text noch Attribut trägt den Canary. bereinigung_zurueckgestellt
// liest keinen Fehler (nur den Bericht): die Quelle bleibt leer, der Eintrag steht aber in der Prüfung (Negativ im Plan).
func TestEintraegeOhneFehlertext(t *testing.T) {
	bg := context.Background()
	ursachen := []error{errors.New(kanarie), &inbound.Fehler{Code: model.CodeE3001, Status: 500, Ursache: errors.New(kanarie)}}
	stellen := []struct {
		name string
		lauf func(u *welt, ursache error)
	}{
		{"bereinigung_fehlgeschlagen", func(u *welt, e error) {
			u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
			u.wartung.fehler = e
			if _, err := u.del(bg, "s1"); err != nil {
				t.Fatal(err)
			}
		}},
		{"bereinigung_zurueckgestellt", func(u *welt, _ error) {
			u.sessions.fuege("s1", dienstA, t0(), model.SessionComplete)
			u.wartung.bericht = inbound.Loeschbericht{Zurueckgestellt: true}
			if _, err := u.del(bg, "s1"); err != nil {
				t.Fatal(err)
			}
		}},
		{"beenden_fehlgeschlagen", func(u *welt, e error) {
			u.dienst(dienstA)
			u.offene("o1", dienstA, 1)
			u.exchange.haenge("o1", model.CaptureComplete)
			u.st.scheitereFuer("Sessions.Get", "o1", e)
			_, _ = u.closeAll(bg)
		}},
	}
	if len(stellen) != 3 {
		t.Fatalf("%d Stellen, erwartet 3", len(stellen))
	}
	for _, s := range stellen {
		for _, e := range ursachen {
			u := neu(t)
			s.lauf(u, e)
			gefunden := 0
			for _, x := range u.log.alle() {
				if x.attrs["ereignis"] == s.name {
					gefunden++
				}
				alle := x.text
				for k, v := range x.attrs {
					alle += " " + k + "=" + v
				}
				if strings.Contains(alle, kanarie) {
					t.Errorf("%s: der Eintrag trägt den Canary: %s", s.name, alle)
				}
			}
			if gefunden != 1 {
				t.Errorf("%s: %d Einträge, erwartet 1", s.name, gefunden)
			}
		}
	}
}
