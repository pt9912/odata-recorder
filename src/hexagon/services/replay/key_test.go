package replay_test

import (
	"context"
	"testing"

	"github.com/pt9912/odata-recorder/src/hexagon/model"
	"github.com/pt9912/odata-recorder/src/hexagon/ports/inbound"
	"github.com/pt9912/odata-recorder/src/hexagon/services/replay"
)

// TestKeyHeaderWieEmpfangen hält Z20: Menge 4; als Vary konfiguriert gehen X-Recorder-*, Hop-by-Hop und Accept mit
// ihrem Wert in den Key, ein redigierter Header (Authorization) nicht.
func TestKeyHeaderWieEmpfangen(t *testing.T) {
	faelle := []struct {
		header    string
		redigiert bool
		wert      [2]string
		haengtAb  bool
	}{
		{"X-Recorder-Replay-Session", false, [2]string{"rep-a", "rep-b"}, true},
		{"Connection", false, [2]string{"keep-alive", "close"}, true},
		{"Accept", false, [2]string{"application/json", "application/xml"}, true},
		{"Authorization", true, [2]string{"Bearer eins", "Bearer zwei"}, false},
	}
	for _, f := range faelle {
		u := neu(t)
		u.k.Regeln.VaryHeaders = []string{f.header}
		if f.redigiert {
			u.k.Regeln.RedactHeaders = []string{f.header}
		}
		mit := func(w string) inbound.Anfrage {
			a := anf()
			a.Header[f.header] = []string{w}
			return a
		}
		u.lege(u.austausch(1, u.keyFuer(mit(f.wert[0]), svcID, "/Products"), true))
		u.k.Strategie = model.StrategyFirst
		if _, err := u.replay(context.Background(), mit(f.wert[0])); err != nil {
			t.Errorf("%s: gleicher Wert: %v", f.header, err)
		}
		_, err := u.replay(context.Background(), mit(f.wert[1]))
		if f.haengtAb {
			erwarteCode(t, f.header+" anderer Wert", err, model.CodeE5001)
		} else if err != nil {
			t.Errorf("%s: anderer Wert: %v", f.header, err)
		}
	}
}

// TestKodierungFehler hält Z21: Menge 4 (Ort × Modus); reject ist ODR-E5006 ohne Cursor, literal kein Fehler.
func TestKodierungFehler(t *testing.T) {
	for _, ort := range []string{"Pfad", "Query"} {
		for _, modus := range []model.KodierungsModus{model.KodierungLiteral, model.KodierungReject} {
			u := neu(t)
			u.k.Regeln.InvalidEncoding = modus
			a := anf()
			odata := "/Products"
			if ort == "Pfad" {
				a.Pfad = root + "/%zz"
				odata = "/%zz"
			} else {
				a.RawQuery = "a=%zz"
			}
			if modus == model.KodierungLiteral {
				u.lege(u.austausch(1, u.keyFuer(a, svcID, odata), true))
			}
			resp, err := u.replay(context.Background(), a)
			if modus == model.KodierungReject {
				erwarteCode(t, ort+" reject", err, model.CodeE5006)
				if _, da := u.cursorStand(); da || u.p.anzahl("Transaktion.WithWorkspaceTransaction") != 0 {
					t.Errorf("%s reject: Fortschritt", ort)
				}
				continue
			}
			if err != nil {
				t.Errorf("%s literal: %v", ort, err)
				continue
			}
			_ = resp.Body.Close()
		}
	}
}

// TestSchluesselUngueltig hält Z22: Menge 5; ein Pfad mit ?, #, 0x00, 0x1F oder 0x7F ist ODR-E9001, nicht ODR-E5006.
func TestSchluesselUngueltig(t *testing.T) {
	for _, b := range []byte{'?', '#', 0x00, 0x1F, 0x7F} {
		u := neu(t)
		u.k.Dienste = []replay.Dienst{{ID: svcID, Root: "/"}}
		u.k.Regeln.InvalidEncoding = model.KodierungReject
		a := anf()
		a.Pfad = "/a" + string([]byte{b})
		_, err := u.replay(context.Background(), a)
		erwarteCode(t, "Byte "+string(rune('0'+b%10)), err, model.CodeE9001)
	}
}
