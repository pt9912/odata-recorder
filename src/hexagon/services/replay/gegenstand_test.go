package replay_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/outbound"
)

// TestAbbildungGegenstand hält Z6 für die drei Zeilen von SPEC-109 Nr. 5, deren Code vom Gegenstand abhängt
// (ErrNichtGefunden, ErrBereitsVorhanden, ErrUngueltig; die dritte nennt allein Import): Menge 70 = Fehler {5:
// die drei Zeilen und die beiden Verbünde aus ErrBereitsVorhanden und ErrUngueltig} × Stelle {7} × Verpackung {2}.
// Der Gegenstand je Stelle steht in SPEC-135 Nr. 18 (S1, S3 Session; S5 Workspace; S4, S6, S7 Sonstiges; S8 Session
// bei ErrNichtGefunden, sonst Sonstiges); die Handtabelle der Codes folgt SPEC-109 Nr. 5, nicht dem Code. Ein Verbund
// gilt nach dem ersten zutreffenden Eintrag der Tabelle (ErrBereitsVorhanden steht vor ErrUngueltig).
func TestAbbildungGegenstand(t *testing.T) {
	ng, bv, ug := outbound.ErrNichtGefunden, outbound.ErrBereitsVorhanden, model.ErrUngueltig
	// Codes je Stelle in der Reihenfolge S1, S3, S4, S5, S6, S7, S8; 0 ist der fehlende Cursor (Erfolg).
	tafel := []struct {
		name  string
		err   error
		codes [7]model.Code
	}{
		{"ErrNichtGefunden", ng, [7]model.Code{model.CodeE7003, model.CodeE7003, model.CodeE9001, model.CodeE7001, 0, model.CodeE9001, model.CodeE7003}},
		{"ErrBereitsVorhanden", bv, [7]model.Code{model.CodeE7010, model.CodeE7010, model.CodeE9001, model.CodeE7007, model.CodeE9001, model.CodeE9001, model.CodeE9001}},
		{"ErrUngueltig", ug, [7]model.Code{model.CodeE9001, model.CodeE9001, model.CodeE9001, model.CodeE9001, model.CodeE9001, model.CodeE9001, model.CodeE9001}},
		{"Join(ErrBereitsVorhanden, ErrUngueltig)", errors.Join(bv, ug), [7]model.Code{model.CodeE7010, model.CodeE7010, model.CodeE9001, model.CodeE7007, model.CodeE9001, model.CodeE9001, model.CodeE9001}},
		{"Join(ErrUngueltig, ErrBereitsVorhanden)", errors.Join(ug, bv), [7]model.Code{model.CodeE7010, model.CodeE7010, model.CodeE9001, model.CodeE7007, model.CodeE9001, model.CodeE9001, model.CodeE9001}},
	}
	n := 0
	for i, s := range stellen() {
		for _, f := range tafel {
			for _, verpackt := range []bool{false, true} {
				n++
				err := f.err
				if verpackt {
					err = fmt.Errorf("verpackt: %w", err)
				}
				was := fmt.Sprintf("%s %s verpackt: %v", s.name, f.name, verpackt)
				u := neu(t)
				a := erfolgsLauf(u)
				u.p.setzeFehler(s.op, err)
				resp, got := u.replay(context.Background(), a)
				if want := f.codes[i]; want == 0 {
					if got != nil {
						t.Errorf("%s: Fehler %v, erwartet Erfolg", was, got)
					} else {
						_ = resp.Body.Close()
					}
				} else {
					erwarteCode(t, was, got, want)
				}
			}
		}
	}
	if n != 70 {
		t.Errorf("Menge: %d Elemente, erwartet 70", n)
	}
}
